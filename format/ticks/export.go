package ticks

import (
	"math/big"
	"slices"

	"github.com/go-muse/musex/internal/timecalc"
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
	for _, x := range pm.MustExtensions() {
		position, _ := x.MustAddress().MustPosition()
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

func (e *exporter) span(from, to *big.Rat) span {
	s := e.segmentAt(from)
	return span{form: timecalc.LinearSpan(s.a, s.b, s.sa, s.sb, from, to)}
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

// baseTempo retains an original source step, even after hold boundaries split
// it. A rational tempo is stored directly; otherwise the tempo is c*T.
type baseTempo struct {
	exact *big.Rat
	part  span
	c     *big.Rat
}

func (e *exporter) base(q *big.Rat) baseTempo {
	s := e.segmentAt(q)
	if !s.gradual {
		return baseTempo{exact: e.microseconds(s.sa)}
	}
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
		return baseTempo{exact: e.microseconds(s.speed(q1))}
	case StepEnd:
		return baseTempo{exact: e.microseconds(s.speed(q2))}
	default:
		c := new(big.Rat).Quo(new(big.Rat).Mul(big.NewRat(1000000, 1), e.w), new(big.Rat).Sub(q2, q1))
		return baseTempo{part: e.span(q1, q2), c: c}
	}
}

type exportPoint struct {
	position *big.Rat
	tick     int
	base     baseTempo
}
type tickStretch struct {
	from, to int
	extra    *big.Rat
	time     timeSum
}

// baseTime integrates the effective, unrounded staircase on the final tick
// grid. Zero-tick fragments contribute nothing; the last change wins.
func (e *exporter) baseTime(points []exportPoint, from, to int) timeSum {
	sum := timeSum{exact: new(big.Rat)}
	irrational := false
	for i, p := range points[:len(points)-1] {
		length := min(to, points[i+1].tick) - max(from, p.tick)
		if length <= 0 {
			continue
		}
		scale := new(big.Rat).Quo(ratInt(length), new(big.Rat).Mul(ratInt(e.r.d), big.NewRat(1000000, 1)))
		if p.base.exact != nil {
			sum.exact.Add(sum.exact, new(big.Rat).Mul(scale, p.base.exact))

		} else {
			irrational = true
			sum.parts = append(sum.parts, weightedSpan{source: p.base.part, weight: new(big.Rat).Mul(scale, p.base.c)})
		}
	}
	if irrational {
		sum.constant = sum.exact
		sum.exact = nil
	}
	return sum
}

func (e *exporter) roundedTempo(b baseTempo, stretch *tickStretch) (MIDITempo, error) {
	if stretch == nil {
		if b.exact != nil {
			return midiTempoOfRat(e.op, b.exact, e.o.tempoMode)
		}
		return linearTempo(e.op, b.part, b.c, e.o.tempoMode)
	}
	if stretch.time.exact != nil {
		factor := new(big.Rat).Quo(new(big.Rat).Add(stretch.time.exact, stretch.extra), stretch.time.exact)
		if b.exact != nil {
			return midiTempoOfRat(e.op, new(big.Rat).Mul(b.exact, factor), e.o.tempoMode)
		}
		return linearTempo(e.op, b.part, new(big.Rat).Mul(b.c, factor), e.o.tempoMode)
	}
	if b.exact != nil {
		return stretchedTempo(e.op, stretch.time, b.exact, stretch.extra, e.o.tempoMode)
	}
	return productTempo(e.op, b.part, stretch.time, b.c, stretch.extra, e.o.tempoMode)
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
	positions := e.points()
	// Source conversion policies precede tick representability checks.
	for _, s := range e.segments {
		if s.gradual && e.h == nil {
			return MIDITempoMap{}, nil, newError(ErrNotConvertible, op, "gradual change without steps")
		}
	}
	for _, q := range positions {
		if err := e.holdError(q); err != nil {
			return MIDITempoMap{}, nil, err
		}
	}
	points := make([]exportPoint, len(positions))
	tickAt := make(map[string]int, len(positions))
	for i, q := range positions {
		if err := e.checkClock(q); err != nil {
			return MIDITempoMap{}, nil, err
		}
		tick, err := e.tick(q)
		if err != nil {
			return MIDITempoMap{}, nil, err
		}
		points[i] = exportPoint{position: q, tick: tick}
		tickAt[q.RatString()] = tick
	}
	var stretches []tickStretch
	if o.holds == HoldsStretch {
		for _, h := range e.holds {
			from := tickAt[new(big.Rat).Sub(h.p, e.length).RatString()]
			to := tickAt[h.p.RatString()]
			if from == to {
				return MIDITempoMap{}, nil, newError(ErrNotConvertible, op, "positive hold collapsed to one tick")
			}
			stretches = append(stretches, tickStretch{from: from, to: to, extra: h.e})
		}
	}
	for i, q := range positions {
		points[i].base = e.base(q)
	}
	for i := range stretches {
		stretches[i].time = e.baseTime(points, stretches[i].from, stretches[i].to)
	}
	var changes []MIDITempoChange
	for _, p := range points {
		var stretch *tickStretch
		for i := range stretches {
			if stretches[i].from <= p.tick && p.tick < stretches[i].to {
				stretch = &stretches[i]
				break
			}
		}
		t, err := e.roundedTempo(p.base, stretch)
		if err != nil {
			return MIDITempoMap{}, nil, err
		}
		changes = append(changes, MIDITempoChange{Tick: p.tick, Tempo: t})
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
