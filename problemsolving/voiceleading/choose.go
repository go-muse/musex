package voiceleading

import (
	"slices"

	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

// ConnectionConditions selects the conditions of ChooseConnections. Each
// condition is enabled separately; the zero value enables none.
type ConnectionConditions struct {
	// Complete requires every class of the target among the new voices.
	Complete bool
	// CommonTones requires every upper voice whose current class by spelling
	// belongs to the target to repeat its note.
	CommonTones bool
	// NoCrossings requires non-decreasing positions from the bass upward;
	// a unison is allowed.
	NoCrossings bool
	// NoOverlaps forbids overlaps of adjacent voices, as FindOverlaps.
	NoOverlaps bool
	// Distances is a distance norm without violations; remarks are allowed.
	// Rule == DistanceUnset disables the condition.
	Distances DistanceNorm
	// Parallels forbids every transition of a pair of voices, including the
	// bass, that FindParallels with these options would report.
	// Perfect == 0 disables the condition.
	Parallels ParallelOptions
	// BassDoubled requires the class of the bass in an upper voice.
	BassDoubled bool
	// MT21 requires CheckMT21Connection to conform.
	MT21 bool
}

// measured is an interval with its chromatic distance.
type measured struct {
	interval  interval.Interval
	chromatic system.Step
}

// candidate is one possible new note of a voice.
type candidate struct {
	pitch pitch.Pitch
	note  note.Note
	class int
	move  measured
}

// chooser holds every measurement of one ChooseConnections call.
type chooser struct {
	from    *Sonority
	sys     *system.System
	conds   ConnectionConditions
	classes []spelling.Spelling
	records []int
	common  []bool // common[k]: the current class of voice k is in the target
	limits  []interval.Interval
	cands   [][]candidate
	fromV   [][]measured
	vert    [][][][]measured
	choice  []int
	results []choiceResult
}

type choiceResult struct {
	choice   []int
	total    system.Step
	overflow bool
}

// ChooseConnections returns every distinct choice of upper voices that leads
// from to the target chord with the given bass and passes the enabled
// conditions, as connections from from to new sonorities. For each upper voice
// and each class of the target (by spelling, without repeats) the candidates
// are the nearest pitch of that class not below the current note and the
// nearest strictly below it. A new upper note is a copy of the first target
// note of its class, unbound from its source, with the candidate pitch; the
// bass note is a copy of bass as given. All measurements use the system of
// from; no source is read. The result is ordered by Total, then by the
// positions of the voices from the top down (lower first), then by the target
// class numbers of the voices from the top down. No match gives an empty
// slice. The search prunes only by conditions decided by the voices already
// chosen, so it returns the same set as a full enumeration of the
// (2m)^(n-1) combinations; their number grows exponentially with the voices.
//
// Errors, in this order: (1) ErrUnset for from; ErrUnset with chord.ErrUnset
// in a ChordError{0, -1} for an unset target; ErrUnset with note.ErrUnset for
// an unset bass; (2) ErrNoOrder, ErrSilentPart, ErrVoiceCount for fewer than
// two voices of from; (3) ErrOption in a ChordError{0, -1} for an empty
// target; ErrNoOctave for a bass without an octave; (4) ErrOption for invalid
// conditions; (5) system.ErrInadmissible in a ChordError{0, k} for a target
// name, then without context for the bass; (6) ErrOption when the class of the
// bass is not in the target; (7) a failure to measure the bass motion;
// (8) ErrOption for a selected condition that cannot apply: a distance norm
// unsuited to the number of voices; for MT21 not four voices, from outside the
// norm, a target without three classes, or a bass motion the norm does not
// cover; (9) failures to measure candidates, moves and verticals, in a
// ChordError{0, k} with the first record of the class concerned, except the
// verticals of from; (11) a total that does not fit in an int (with
// system.ErrOutOfRange). The enumeration itself never fails.
func ChooseConnections(from *Sonority, target chord.Chord, bass note.Note, conds ConnectionConditions) ([]Connection, error) {
	c, err := newChooser(from, target, bass, conds)
	if err != nil {
		return nil, err
	}
	c.choice = make([]int, len(from.notes))
	c.search(1)
	for _, r := range c.results {
		if r.overflow {
			return nil, errTotal
		}
	}
	slices.SortStableFunc(c.results, c.compare)
	out := make([]Connection, len(c.results))
	for i, r := range c.results {
		out[i] = c.connection(r)
	}
	return out, nil
}

// MustChooseConnections is like ChooseConnections but panics with its error.
func MustChooseConnections(from *Sonority, target chord.Chord, bass note.Note, conds ConnectionConditions) []Connection {
	return must(ChooseConnections(from, target, bass, conds))
}

func newChooser(from *Sonority, target chord.Chord, bass note.Note, conds ConnectionConditions) (*chooser, error) {
	// (1) Unset arguments.
	if !from.IsSet() {
		return nil, ErrUnset
	}
	if !target.IsValid() {
		return nil, chordErr(0, -1, unset("chord", chord.ErrUnset))
	}
	if !bass.IsValid() {
		return nil, unset("bass", note.ErrUnset)
	}
	// (2) The current sonority.
	switch {
	case from.order != LowToHigh:
		return nil, ErrNoOrder
	case len(from.silent) > 0:
		return nil, ErrSilentPart
	case len(from.notes) < 2:
		return nil, ErrVoiceCount
	}
	// (3) Empty target, bass without an octave.
	if target.IsEmpty() {
		return nil, chordErr(0, -1, ErrOption)
	}
	if !bass.HasOctave() {
		return nil, ErrNoOctave
	}
	// (4) The form of the conditions.
	text, err := validateConditions(conds)
	if err != nil {
		return nil, err
	}
	c := &chooser{from: from, sys: from.sys, conds: conds}
	// (5) Admissible names.
	names := target.Notes()
	for k, n := range names {
		if err := admit(c.sys, n.Spelling()); err != nil {
			return nil, chordErr(0, k, err)
		}
	}
	if err := admit(c.sys, bass.Spelling()); err != nil {
		return nil, cause("bass", err)
	}
	for k, n := range names {
		if !containsName(c.classes, n.Spelling()) {
			c.classes = append(c.classes, n.Spelling())
			c.records = append(c.records, k)
		}
	}
	// (6) The class of the bass.
	bassClass := slices.IndexFunc(c.classes, bass.Spelling().Equal)
	if bassClass < 0 {
		return nil, ErrOption
	}
	// (7) The bass motion.
	bassPitch := bass.MustPitch()
	bassMove, err := c.measure(from.pitches[0], bassPitch)
	if err != nil {
		return nil, cause("bass motion", err)
	}
	// (8) Conditions that cannot apply.
	if err := c.applicable(conds, text, bassMove.interval); err != nil {
		return nil, err
	}
	c.common = make([]bool, len(from.notes))
	for k, n := range from.notes {
		c.common[k] = containsName(c.classes, n.Spelling())
	}
	// (9) Measurements in a fixed order.
	c.cands = make([][]candidate, len(from.notes))
	c.cands[0] = []candidate{{pitch: bassPitch, note: bass, class: bassClass, move: bassMove}}
	if err := c.measureCandidates(names); err != nil {
		return nil, err
	}
	if err := c.measureVerticals(); err != nil {
		return nil, err
	}
	return c, nil
}

func validateConditions(conds ConnectionConditions) (distanceText, error) {
	var text distanceText
	if conds.Distances.Rule != DistanceUnset {
		var err error
		if text, err = validateNorm(conds.Distances); err != nil {
			return distanceText{}, err
		}
	} else if len(conds.Distances.Limits) > 0 {
		return distanceText{}, ErrOption
	}
	p := conds.Parallels
	if p.Perfect != 0 && !validPerfect(p.Perfect) || p.Perfect == 0 && (p.Compound || p.Unisons) {
		return distanceText{}, ErrOption
	}
	return text, nil
}

func (c *chooser) applicable(conds ConnectionConditions, text distanceText, bassMotion interval.Interval) error {
	n := len(c.from.notes)
	if conds.Distances.Rule != DistanceUnset {
		limits, ok := text.pairLimits(conds.Distances, n)
		if !text.fitsVoices(n) || !ok {
			return ErrOption
		}
		c.limits = limits
	}
	if conds.MT21 {
		switch {
		case n != 4,
			classCount(c.from.notes) != 3 || !bassDoubled(c.from.notes),
			len(c.classes) != 3,
			!bassMotionCovered(bassMotion):
			return ErrOption
		}
	}
	return nil
}

func (c *chooser) measure(x, y pitch.Pitch) (measured, error) {
	i, err := interval.BetweenPitches(c.sys, x, y)
	if err != nil {
		return measured{}, err
	}
	d, err := c.sys.Distance(x, y)
	if err != nil {
		return measured{}, err
	}
	return measured{interval: i, chromatic: d}, nil
}

// measureCandidates finds the candidates of every upper voice (9.1), then
// measures the moves to them (9.2).
func (c *chooser) measureCandidates(names []note.Note) error {
	for k := 1; k < len(c.from.notes); k++ {
		cur := c.from.pitches[k]
		for ci, name := range c.classes {
			upper, lower, err := c.nearest(cur, name)
			if err != nil {
				return chordErr(0, c.records[ci], err)
			}
			for _, p := range []pitch.Pitch{upper, lower} {
				n := names[c.records[ci]]
				if err := n.Unbind(); err != nil {
					return chordErr(0, c.records[ci], err)
				}
				if err := n.ReplaceWithPitch(p); err != nil {
					return chordErr(0, c.records[ci], err)
				}
				c.cands[k] = append(c.cands[k], candidate{pitch: p, note: n, class: ci})
			}
		}
	}
	for k := 1; k < len(c.from.notes); k++ {
		for i := range c.cands[k] {
			m, err := c.measure(c.from.pitches[k], c.cands[k][i].pitch)
			if err != nil {
				return chordErr(0, c.records[c.cands[k][i].class], err)
			}
			c.cands[k][i].move = m
		}
	}
	return nil
}

// nearest returns the lowest pitch named name not below cur and the highest
// pitch named name strictly below cur.
func (c *chooser) nearest(cur pitch.Pitch, name spelling.Spelling) (upper, lower pitch.Pitch, err error) {
	upper, err = cur.WithSpelling(name)
	if err != nil {
		return pitch.Pitch{}, pitch.Pitch{}, err
	}
	cmp, err := c.sys.Compare(upper, cur)
	if err != nil {
		return pitch.Pitch{}, pitch.Pitch{}, err
	}
	// Positions are octave·N + class with 0 <= class < N, so one octave of
	// adjustment at most separates the same octave from the nearest pitch.
	if cmp < 0 {
		if upper, err = upper.Up(); err != nil {
			return pitch.Pitch{}, pitch.Pitch{}, err
		}
	}
	if lower, err = upper.Down(); err != nil {
		return pitch.Pitch{}, pitch.Pitch{}, err
	}
	return upper, lower, nil
}

// measureVerticals measures every pair of voices of from (9.3), then every
// pair of candidates of the new sonority (9.4).
func (c *chooser) measureVerticals() error {
	n := len(c.from.notes)
	c.fromV = make([][]measured, n)
	for i := range n {
		c.fromV[i] = make([]measured, n)
		for j := i + 1; j < n; j++ {
			m, err := c.measure(c.from.pitches[i], c.from.pitches[j])
			if err != nil {
				return cause("vertical", err)
			}
			c.fromV[i][j] = m
		}
	}
	c.vert = make([][][][]measured, n)
	for i := range n {
		c.vert[i] = make([][][]measured, n)
		for j := i + 1; j < n; j++ {
			c.vert[i][j] = make([][]measured, len(c.cands[i]))
			for a, x := range c.cands[i] {
				c.vert[i][j][a] = make([]measured, len(c.cands[j]))
				for b, y := range c.cands[j] {
					m, err := c.measure(x.pitch, y.pitch)
					if err != nil {
						return chordErr(0, c.records[y.class], err)
					}
					c.vert[i][j][a][b] = m
				}
			}
		}
	}
	return nil
}

// search chooses voice k after voices 0..k-1 and prunes by the conditions
// decided by the chosen voices only.
func (c *chooser) search(k int) {
	if k == len(c.from.notes) {
		c.leaf()
		return
	}
	for a := range c.cands[k] {
		c.choice[k] = a
		if c.prefixPasses(k) {
			c.search(k + 1)
		}
	}
}

// prefixPasses checks the conditions that the choice of voice k completes:
// common tones of voice k, and crossings, overlaps, distances and parallels of
// the pairs ending at voice k.
func (c *chooser) prefixPasses(k int) bool {
	cand := c.cands[k][c.choice[k]]
	if c.conds.CommonTones && c.common[k] && !cand.move.interval.IsPerfectUnison() {
		return false
	}
	below := c.vert[k-1][k][c.choice[k-1]][c.choice[k]]
	if c.conds.NoCrossings && below.chromatic < 0 {
		return false
	}
	if c.conds.NoOverlaps && c.overlaps(k) {
		return false
	}
	if c.limits != nil && exceeds(below.interval, c.limits[k-1]) {
		return false
	}
	if c.conds.Parallels.Perfect != 0 {
		for i := range k {
			if c.parallel(i, k) {
				return false
			}
		}
	}
	return true
}

// overlaps applies the definition of FindOverlaps to voices k-1 and k.
func (c *chooser) overlaps(k int) bool {
	if c.fromV[k-1][k].chromatic < 0 {
		return false
	}
	lower, upper := c.cands[k-1][c.choice[k-1]].pitch, c.cands[k][c.choice[k]].pitch
	// Every pitch was admitted before the search, so Compare cannot fail.
	return c.sys.MustCompare(lower, c.from.pitches[k]) > 0 || c.sys.MustCompare(upper, c.from.pitches[k-1]) < 0
}

// parallel applies the definition of FindParallels to voices i and j.
func (c *chooser) parallel(i, j int) bool {
	x, y := c.cands[i][c.choice[i]], c.cands[j][c.choice[j]]
	to := c.vert[i][j][c.choice[i]][c.choice[j]]
	t := Transition{
		From:   Vertical{Sounding: true, Specific: c.fromV[i][j].interval, Chromatic: c.fromV[i][j].chromatic},
		To:     Vertical{Sounding: true, Specific: to.interval, Chromatic: to.chromatic},
		First:  x.move.interval,
		Second: y.move.interval,
		Motion: classify(x.move.interval, y.move.interval),
	}
	return strictParallel(t, c.conds.Parallels)
}

// leaf checks the conditions of the complete variant and records it.
func (c *chooser) leaf() {
	n := len(c.from.notes)
	present := make([]bool, len(c.classes))
	bassClass := c.cands[0][0].class
	doubled := false
	distinct := 0
	for k := range n {
		cl := c.cands[k][c.choice[k]].class
		if !present[cl] {
			distinct++
		}
		present[cl] = true
		doubled = doubled || k > 0 && cl == bassClass
	}
	if c.conds.Complete && distinct != len(c.classes) {
		return
	}
	if c.conds.BassDoubled && !doubled {
		return
	}
	if c.conds.MT21 && (distinct != 3 || !doubled || !c.mt21Conforms()) {
		return
	}
	r := choiceResult{choice: slices.Clone(c.choice)}
	var ok bool
	for k := range n {
		if r.total, ok = addAbs(r.total, c.cands[k][c.choice[k]].move.chromatic); !ok {
			r.overflow = true
			break
		}
	}
	c.results = append(c.results, r)
}

func (c *chooser) mt21Conforms() bool {
	moves := make([]VoiceMove, len(c.from.notes))
	for k := range moves {
		moves[k] = VoiceMove{Interval: c.cands[k][c.choice[k]].move.interval}
	}
	for _, v := range mt21Variants(moves) {
		if v.Satisfied {
			return true
		}
	}
	return false
}

// compare orders results by total, then by positions from the top voice
// down, then by class numbers from the top voice down.
func (c *chooser) compare(x, y choiceResult) int {
	if x.total != y.total {
		if x.total < y.total {
			return -1
		}
		return 1
	}
	for k := len(x.choice) - 1; k >= 0; k-- {
		if r := c.sys.MustCompare(c.cands[k][x.choice[k]].pitch, c.cands[k][y.choice[k]].pitch); r != 0 {
			return r
		}
	}
	for k := len(x.choice) - 1; k >= 0; k-- {
		if a, b := c.cands[k][x.choice[k]].class, c.cands[k][y.choice[k]].class; a != b {
			if a < b {
				return -1
			}
			return 1
		}
	}
	return 0
}

// connection builds the result of one choice from the measurements.
func (c *chooser) connection(r choiceResult) Connection {
	n := len(r.choice)
	to := &Sonority{
		set: true, sys: c.sys, order: LowToHigh, silent: []int{},
		notes: make([]note.Note, n), pitches: make([]pitch.Pitch, n), adjacent: make([]VoiceInterval, 0, n-1),
	}
	out := Connection{From: c.from, To: to, Moves: make([]VoiceMove, n), Total: r.total, Kind: MelodicConnection}
	for k := range n {
		cand := c.cands[k][r.choice[k]]
		to.notes[k], to.pitches[k] = cand.note, cand.pitch
		out.Moves[k] = VoiceMove{Voice: k, From: c.from.notes[k], To: cand.note, Interval: cand.move.interval, Chromatic: cand.move.chromatic}
		if out.Moves[k].IsRepeat() {
			out.Kind = HarmonicConnection
		}
	}
	pair := func(i, j int) VoiceInterval {
		m := c.vert[i][j][r.choice[i]][r.choice[j]]
		return VoiceInterval{Lower: i, Upper: j, Interval: m.interval, Chromatic: m.chromatic}
	}
	for k := 1; k < n; k++ {
		to.adjacent = append(to.adjacent, pair(k-1, k))
	}
	to.span = pair(0, n-1)
	out.Common = commonClasses(c.from.notes, to.notes)
	return out
}
