package ticks

import (
	"math/big"
	"slices"

	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/tempomap"
)

// checkMusexMap validates a musex tempo map argument.
func checkMusexMap(op string, m tempomap.Map) error {
	if !m.IsSet() {
		return invalidError(op, "tempo map", tempomap.ErrUnset)
	}
	return nil
}

// positionalMap returns m as a positional map: a positional map takes no grid,
// a measure map exactly one set grid and is converted with ToPositions.
func positionalMap(op string, m tempomap.Map, grid []measure.Grid) (tempomap.Map, error) {
	if m.MustAddressing() == tempomap.PositionAddressing {
		if len(grid) != 0 {
			return tempomap.Map{}, invalidError(op, "a positional map takes no grid")
		}
		return m, nil
	}
	if len(grid) != 1 {
		return tempomap.Map{}, invalidError(op, "a measure map requires exactly one grid")
	}
	if !grid[0].IsSet() {
		return tempomap.Map{}, invalidError(op, "grid", measure.ErrUnset)
	}
	positional, err := m.ToPositions(grid[0])
	if err != nil {
		return tempomap.Map{}, dependencyError(op, err)
	}
	return positional, nil
}

func at(position *big.Rat) tempomap.Address {
	return tempomap.MustAt(must(positionOf("ticks.at", position)))
}

// exportSegment is a segment of a positional musex map.
type exportSegment struct {
	a, b    *big.Rat // b is nil for the open final segment
	sa, sb  *big.Rat
	gradual bool // a gradual segment with different end speeds
}

// speed returns the exact speed at q of a gradual segment.
func (s exportSegment) speed(q *big.Rat) *big.Rat {
	v := new(big.Rat).Sub(s.sb, s.sa)
	v.Mul(v, new(big.Rat).Sub(q, s.a))
	v.Quo(v, new(big.Rat).Sub(s.b, s.a))
	return v.Add(v, s.sa)
}

type exportHold struct {
	p, e      *big.Rat
	extension tempomap.Extension
}

type exporter struct {
	op       string
	r        Resolution
	o        ExportOptions
	bare     tempomap.Map
	segments []exportSegment
	holds    []exportHold
	w        *big.Rat
	h        *big.Rat // step, nil without steps
	length   *big.Rat // stretch length
}

func newExporter(op string, pm tempomap.Map, r Resolution, o ExportOptions) (*exporter, error) {
	e := &exporter{op: op, r: r, o: o, w: r.quarter()}
	if o.method != 0 {
		e.h = ratOfDuration(o.step)
	}
	if o.holds == HoldsStretch {
		e.length = ratOfDuration(o.stretch)
	}
	segments, err := pm.Segments()
	if err != nil {
		return nil, dependencyError(op, err)
	}
	for _, s := range segments {
		start, _ := s.MustStart().MustPosition()
		seg := exportSegment{a: ratOfPosition(start), sa: ratOfFraction(s.MustStartSpeed())}
		seg.sb = seg.sa
		if end, ok := s.MustEnd(); ok {
			endPosition, _ := end.MustPosition()
			seg.b = ratOfPosition(endPosition)
			endSpeed, _ := s.MustEndSpeed()
			seg.sb = ratOfFraction(endSpeed)
		}
		seg.gradual = s.MustChange().MustKind() == tempomap.ChangeGradual && seg.sa.Cmp(seg.sb) != 0
		e.segments = append(e.segments, seg)
	}
	e.bare = pm
	for _, x := range pm.MustExtensions() {
		position, _ := x.MustAddress().MustPosition()
		e.bare = e.bare.MustWithoutExtension(x.MustAddress())
		seconds := ratOfSeconds(x.MustSeconds())
		if seconds.Sign() == 0 {
			continue
		}
		e.holds = append(e.holds, exportHold{p: ratOfPosition(position), e: seconds, extension: x})
	}
	return e, nil
}

// segmentAt returns the segment that contains q.
func (e *exporter) segmentAt(q *big.Rat) exportSegment {
	current := e.segments[0]
	for _, s := range e.segments[1:] {
		if s.a.Cmp(q) > 0 {
			break
		}
		current = s
	}
	return current
}

// stretchAt returns the hold whose stretched segment contains q.
func (e *exporter) stretchAt(q *big.Rat) (exportHold, bool) {
	if e.length == nil {
		return exportHold{}, false
	}
	for _, h := range e.holds {
		start := new(big.Rat).Sub(h.p, e.length)
		if start.Cmp(q) <= 0 && q.Cmp(h.p) < 0 {
			return h, true
		}
	}
	return exportHold{}, false
}

// irrational reports whether the seconds of [from, to) involve a gradual
// segment with different end speeds.
func (e *exporter) irrational(from, to *big.Rat) bool {
	for _, s := range e.segments {
		if !s.gradual {
			continue
		}
		left, right := s.a, s.b
		if from.Cmp(left) > 0 {
			left = from
		}
		if to.Cmp(right) < 0 {
			right = to
		}
		if left.Cmp(right) < 0 {
			return true
		}
	}
	return false
}

func (e *exporter) span(from, to *big.Rat) span {
	return span{m: e.bare, from: at(from), to: at(to)}
}

// points returns the sorted distinct positions of the changes to create and
// the positions where an error is reported.
func (e *exporter) points() []*big.Rat {
	var points []*big.Rat
	for _, s := range e.segments {
		points = append(points, s.a)
		if s.gradual && e.h != nil {
			for q := new(big.Rat).Add(s.a, e.h); q.Cmp(s.b) < 0; q = new(big.Rat).Add(q, e.h) {
				points = append(points, q)
			}
		}
	}
	for _, h := range e.holds {
		switch e.o.holds {
		case HoldsRefuse:
			points = append(points, h.p)
		case HoldsStretch:
			start := new(big.Rat).Sub(h.p, e.length)
			if h.p.Sign() == 0 || start.Sign() < 0 {
				points = append(points, new(big.Rat))
				continue
			}
			points = append(points, start, h.p)
		}
	}
	slices.SortFunc(points, func(a, b *big.Rat) int { return a.Cmp(b) })
	return slices.CompactFunc(points, func(a, b *big.Rat) bool { return a.Cmp(b) == 0 })
}

// holdError reports a hold that cannot be exported, at the point q.
func (e *exporter) holdError(q *big.Rat) error {
	for i, h := range e.holds {
		switch e.o.holds {
		case HoldsRefuse:
			if h.p.Cmp(q) == 0 {
				return newError(ErrNotConvertible, e.op, "fermata hold refused")
			}
		case HoldsStretch:
			start := new(big.Rat).Sub(h.p, e.length)
			if h.p.Sign() == 0 || start.Sign() < 0 {
				if q.Sign() == 0 {
					return newError(ErrNotConvertible, e.op, "stretched segment of a hold starts before the map")
				}
				continue
			}
			if i > 0 && start.Cmp(e.holds[i-1].p) < 0 && start.Cmp(q) == 0 {
				return newError(ErrNotConvertible, e.op, "stretched segments of two holds overlap")
			}
		}
	}
	return nil
}

// stepStart reports whether q starts a step of a gradual segment.
func (e *exporter) stepStart(q *big.Rat) bool {
	s := e.segmentAt(q)
	if !s.gradual || e.h == nil {
		return false
	}
	offset := new(big.Rat).Sub(q, s.a)
	return offset.Quo(offset, e.h).IsInt()
}

func (e *exporter) checkClock(q *big.Rat) error {
	if !e.o.clock || !e.stepStart(q) {
		return nil
	}
	if e.r.d%ClocksPerQuarter != 0 {
		return newError(ErrNotRepresentable, e.op, "MIDI clocks are not whole ticks")
	}
	clocks := new(big.Rat).Mul(q, big.NewRat(ClocksPerQuarter, 1))
	if !clocks.Quo(clocks, e.w).IsInt() {
		return newError(ErrNotRepresentable, e.op, "step not on a MIDI clock")
	}
	return nil
}

func (e *exporter) tick(q *big.Rat) (int, error) {
	units := e.r.unitsOf(q)
	if e.o.positionMode == 0 {
		return wholeTicks(e.op, units)
	}
	n, err := roundInt(e.op, units, e.o.positionMode)
	if err != nil {
		return 0, err
	}
	return intOf(e.op, n)
}

func (e *exporter) microseconds(speed *big.Rat) *big.Rat { return microsecondsOfSpeed(speed, e.w) }

// tempo returns the tempo of the change at q, rounded once from its exact
// value.
func (e *exporter) tempo(q *big.Rat) (MIDITempo, error) {
	mode := e.o.tempoMode
	s := e.segmentAt(q)
	var exact *big.Rat // the exact tempo, or nil for the equal-time tempo c*T
	var part span
	var c *big.Rat
	if !s.gradual {
		exact = e.microseconds(s.sa)
	} else {
		offset := new(big.Rat).Sub(q, s.a)
		steps := new(big.Rat).Quo(offset, e.h)
		k := new(big.Int).Quo(steps.Num(), steps.Denom())
		q1 := new(big.Rat).Add(s.a, new(big.Rat).Mul(new(big.Rat).SetInt(k), e.h))
		q2 := new(big.Rat).Add(q1, e.h)
		if q2.Cmp(s.b) > 0 {
			q2 = s.b
		}
		switch e.o.method {
		case StepStart:
			exact = e.microseconds(s.speed(q1))
		case StepEnd:
			exact = e.microseconds(s.speed(q2))
		default:
			c = new(big.Rat).Mul(big.NewRat(1000000, 1), e.w)
			c.Quo(c, new(big.Rat).Sub(q2, q1))
			part = e.span(q1, q2)
		}
	}
	hold, stretched := e.stretchAt(q)
	if !stretched {
		if exact != nil {
			return midiTempoOfRat(e.op, exact, mode)
		}
		return linearTempo(e.op, part, c, mode)
	}
	from := new(big.Rat).Sub(hold.p, e.length)
	stretch := e.span(from, hold.p)
	if !e.irrational(from, hold.p) {
		result, err := e.bare.Between(stretch.from, stretch.to, 0, rounding.Down)
		if err != nil {
			return MIDITempo{}, dependencyError(e.op, err)
		}
		seconds := ratOfSeconds(result.MustValue())
		factor := new(big.Rat).Add(seconds, hold.e)
		factor.Quo(factor, seconds)
		if exact != nil {
			return midiTempoOfRat(e.op, exact.Mul(exact, factor), mode)
		}
		return linearTempo(e.op, part, c.Mul(c, factor), mode)
	}
	if exact != nil {
		return stretchedTempo(e.op, stretch, exact, hold.e, mode)
	}
	return productTempo(e.op, part, stretch, c, hold.e, mode)
}

// MIDITempoMapOf converts the musex tempo map m to a MIDI tempo map of the
// resolution r with the options o, and returns the skipped holds (nil if
// none). A measure map takes exactly one grid and is converted to positions
// first.
//
// Constant segments give their exact tempo 60*10^6*w/s. A gradual segment with
// different end speeds needs steps (ExportOptions.WithSteps), otherwise it is
// ErrNotConvertible; steps start at a, a+h, ... and the last is cut by the end
// of the segment. Holds with zero seconds are ignored; other holds are refused,
// skipped or stretched by the options. Each change is rounded once from its
// exact tempo with the tempo mode; a tempo outside 1..16777215 is
// ErrFormatLimit. A position that is not a whole number of ticks is
// ErrNotRepresentable unless a position mode is set. The result has the
// initial tempo 500000 and its first change at tick 0. The only case of
// ErrOutOfRange for a precision that does not decide is a stretched
// equal-time step whose step and stretched segment are both irrational.
func MIDITempoMapOf(m tempomap.Map, r Resolution, o ExportOptions, grid ...measure.Grid) (MIDITempoMap, []tempomap.Extension, error) {
	const op = "MIDITempoMapOf"
	if err := checkMusexMap(op, m); err != nil {
		return MIDITempoMap{}, nil, err
	}
	if !r.IsSet() {
		return MIDITempoMap{}, nil, unsetError(op, "resolution")
	}
	if !o.IsSet() {
		return MIDITempoMap{}, nil, unsetError(op, "export options")
	}
	pm, err := positionalMap(op, m, grid)
	if err != nil {
		return MIDITempoMap{}, nil, err
	}
	if pm.MustUnfinished() {
		return MIDITempoMap{}, nil, newError(ErrNotConvertible, op, "unfinished gradual change", tempomap.ErrUnfinished)
	}
	e, err := newExporter(op, pm, r, o)
	if err != nil {
		return MIDITempoMap{}, nil, err
	}
	var changes []MIDITempoChange
	for _, q := range e.points() {
		s := e.segmentAt(q)
		if s.gradual && s.a.Cmp(q) == 0 && e.h == nil {
			return MIDITempoMap{}, nil, newError(ErrNotConvertible, op, "gradual change without steps")
		}
		if err := e.holdError(q); err != nil {
			return MIDITempoMap{}, nil, err
		}
		if err := e.checkClock(q); err != nil {
			return MIDITempoMap{}, nil, err
		}
		tick, err := e.tick(q)
		if err != nil {
			return MIDITempoMap{}, nil, err
		}
		t, err := e.tempo(q)
		if err != nil {
			return MIDITempoMap{}, nil, err
		}
		changes = append(changes, MIDITempoChange{Tick: tick, Tempo: t})
	}
	var skipped []tempomap.Extension
	if o.holds == HoldsSkip {
		for _, h := range e.holds {
			skipped = append(skipped, h.extension)
		}
	}
	return MIDITempoMap{r: r, initial: DefaultMIDITempo(), changes: changes}, skipped, nil
}

// MustMIDITempoMapOf is like MIDITempoMapOf but panics with its error.
func MustMIDITempoMapOf(m tempomap.Map, r Resolution, o ExportOptions, grid ...measure.Grid) (MIDITempoMap, []tempomap.Extension) {
	return must2(MIDITempoMapOf(m, r, o, grid...))
}
