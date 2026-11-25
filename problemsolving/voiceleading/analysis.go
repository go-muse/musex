package voiceleading

import (
	"fmt"
	"slices"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/voice"
)

type pairData struct {
	pair        Pair
	verticals   []Vertical
	transitions []Transition
}

// Analysis is an immutable, measured snapshot of input parts. The zero value
// and nil pointer are unset. Assignment copies are safe and share only
// immutable data; values are not comparable with ==. Equal compares musical
// records and measured context without reading sources of settings.
// All reading and search methods are safe for concurrent use. Returned slices
// and assessments are independent; borrowed systems must remain immutable.
type Analysis struct {
	names []string
	order Order
	sys   *system.System
	data  []pairData
	index map[Pair]int
	// parts keeps every element of every part, including tails beyond all
	// pairwise intersections, for the melody and sonority operations.
	parts [][]ElementAt
}

type partInput struct {
	elements []ElementAt
	needed   []bool
}

type measuredNote struct {
	part, element int
	note          note.Note
	sys           *system.System
}

// Analyze measures all pairs of at least two set lines. Parts are identified
// by input slot. Each pair has its own grid, ending at the shorter line.
// Only notes occurring in a sounding vertical need a system. An explicit
// system bypasses sources; otherwise each participating element's settings
// are read once. The input must remain unchanged during the call.
// Failure returns nil and a contextual error; no partial analysis escapes.
func Analyze(lines []*voice.Line, opts Options) (*Analysis, error) {
	if len(lines) < 2 {
		return nil, ErrPartCount
	}
	for i, line := range lines {
		if !line.IsSet() {
			return nil, &PartError{Part: i, Element: -1, Err: unset("line", voice.ErrUnset)}
		}
	}
	if opts.Order != Unordered && opts.Order != LowToHigh {
		return nil, ErrOption
	}
	if opts.System != nil && !opts.System.IsValid() {
		return nil, unset("system", system.ErrInvalid)
	}
	a := &Analysis{names: make([]string, len(lines)), order: opts.Order, sys: opts.System, index: make(map[Pair]int), parts: make([][]ElementAt, len(lines))}
	parts := make([]partInput, len(lines))
	for i, line := range lines {
		a.names[i] = line.MustName()
		elems := line.MustElements()
		parts[i] = partInput{elements: make([]ElementAt, len(elems)), needed: make([]bool, len(elems))}
		for j, e := range elems {
			parts[i].elements[j] = ElementAt{Index: j, Element: e, Start: line.MustStartOf(j), End: line.MustEndOf(j)}
		}
		a.parts[i] = parts[i].elements
	}
	for i := range parts {
		for j := i + 1; j < len(parts); j++ {
			p := Pair{First: i, Second: j}
			vs := preparePair(p, parts[i], parts[j])
			a.index[p] = len(a.data)
			a.data = append(a.data, pairData{pair: p, verticals: vs})
		}
	}
	if err := a.chooseSystem(parts); err != nil {
		return nil, err
	}
	for i := range a.data {
		if err := a.measure(&a.data[i]); err != nil {
			return nil, err
		}
	}
	return a, nil
}

// MustAnalyze is like Analyze but panics with its error.
func MustAnalyze(lines []*voice.Line, opts Options) *Analysis { return must(Analyze(lines, opts)) }

// preparePair merges two already ordered element sequences in linear time.
// Original element boundaries are retained even at a clipped joint end.
func preparePair(pair Pair, first, second partInput) []Vertical {
	vs := []Vertical{}
	start := duration.Origin()
	i, j := 0, 0
	for i < len(first.elements) && j < len(second.elements) {
		x, y := first.elements[i], second.elements[j]
		end := x.End
		if y.End.Compare(end) < 0 {
			end = y.End
		}
		sounding := x.Element.IsSound() && y.Element.IsSound()
		vs = append(vs, Vertical{Pair: pair, Start: start, End: end, First: x, Second: y, Sounding: sounding})
		if sounding {
			first.needed[i], second.needed[j] = true, true
		}
		if x.End.Equal(end) {
			i++
		}
		if y.End.Equal(end) {
			j++
		}
		start = end
	}
	return vs
}

func (a *Analysis) chooseSystem(parts []partInput) error {
	notes := []measuredNote{}
	for i, part := range parts {
		for j, e := range part.elements {
			if part.needed[j] {
				n, _ := e.Element.MustNote()
				notes = append(notes, measuredNote{part: i, element: j, note: n})
			}
		}
	}
	if a.sys == nil {
		// Check all bindings before consulting even the first source.
		for _, n := range notes {
			if !n.note.IsBound() {
				return &PartError{Part: n.part, Element: n.element, Err: ErrNoSystem}
			}
		}
		for i := range notes {
			s, err := notes[i].note.Settings()
			if err != nil {
				return &PartError{Part: notes[i].part, Element: notes[i].element, Err: err}
			}
			notes[i].sys = s.System()
		}
		if len(notes) > 0 {
			a.sys = notes[0].sys
		}
		for _, n := range notes {
			if !a.sys.Equal(n.sys) {
				return &PartError{Part: n.part, Element: n.element, Err: ErrDifferentSystems}
			}
		}
	}
	for _, n := range notes {
		ok, err := a.sys.Admits(n.note.Spelling())
		if err != nil {
			return &PartError{Part: n.part, Element: n.element, Err: err}
		}
		if !ok {
			return &PartError{Part: n.part, Element: n.element, Err: system.ErrInadmissible}
		}
	}
	return nil
}

func elementPitch(e ElementAt) pitch.Pitch {
	n, _ := e.Element.MustNote()
	return n.MustPitch()
}

func (a *Analysis) measure(p *pairData) error {
	for i := range p.verticals {
		v := &p.verticals[i]
		if !v.Sounding {
			continue
		}
		x, y := elementPitch(v.First), elementPitch(v.Second)
		specific, err := interval.BetweenPitches(a.sys, x, y)
		if err != nil {
			return &MeasurementError{Pair: p.pair, At: v.Start, Err: err}
		}
		chromatic, err := a.sys.Distance(x, y)
		if err != nil {
			return &MeasurementError{Pair: p.pair, At: v.Start, Err: err}
		}
		v.Specific, v.Chromatic = specific, chromatic
	}
	p.transitions = make([]Transition, 0, max(0, len(p.verticals)-1))
	for i := 1; i < len(p.verticals); i++ {
		t, err := a.measureTransition(p.verticals[i-1], p.verticals[i])
		if err != nil {
			return &MeasurementError{Pair: p.pair, At: p.verticals[i].Start, Err: err}
		}
		p.transitions = append(p.transitions, t)
	}
	return nil
}

func (a *Analysis) measureTransition(from, to Vertical) (Transition, error) {
	t := Transition{From: from, To: to}
	if !from.Sounding || !to.Sounding {
		return t, nil
	}
	x0, x1 := elementPitch(from.First), elementPitch(to.First)
	y0, y1 := elementPitch(from.Second), elementPitch(to.Second)
	var err error
	if t.First, err = interval.BetweenPitches(a.sys, x0, x1); err != nil {
		return Transition{}, err
	}
	if t.Second, err = interval.BetweenPitches(a.sys, y0, y1); err != nil {
		return Transition{}, err
	}
	if t.FirstChromatic, err = a.sys.Distance(x0, x1); err != nil {
		return Transition{}, err
	}
	if t.SecondChromatic, err = a.sys.Distance(y0, y1); err != nil {
		return Transition{}, err
	}
	t.Motion = classify(t.First, t.Second)
	return t, nil
}

func classify(first, second interval.Interval) Motion {
	x, y := first.IsPerfectUnison(), second.IsPerfectUnison()
	switch {
	case x && y:
		return MotionStationary
	case x || y:
		return MotionOblique
	case first.Direction() != second.Direction():
		return MotionContrary
	case first.Size() == second.Size():
		return MotionParallel
	default:
		return MotionDirect
	}
}

// IsSet reports whether Analyze successfully initialized the snapshot.
func (a *Analysis) IsSet() bool { return a != nil && len(a.names) >= 2 }

func (a *Analysis) check() error {
	if !a.IsSet() {
		return ErrUnset
	}
	return nil
}

func (a *Analysis) pair(p Pair) (*pairData, error) {
	if err := a.check(); err != nil {
		return nil, err
	}
	i, ok := a.index[p]
	if !ok {
		return nil, fmt.Errorf("%w: (%d,%d)", ErrPart, p.First, p.Second)
	}
	return &a.data[i], nil
}

// PartCount returns the number of input slots, or ErrUnset.
func (a *Analysis) PartCount() (int, error) {
	if err := a.check(); err != nil {
		return 0, err
	}
	return len(a.names), nil
}

// MustPartCount is like PartCount but panics with its error.
func (a *Analysis) MustPartCount() int { return must(a.PartCount()) }

// Name returns the original name of a part, including an empty name.
// Errors are ErrUnset, then ErrPart for an out-of-range index.
func (a *Analysis) Name(part int) (string, error) {
	if err := a.check(); err != nil {
		return "", err
	}
	if part < 0 || part >= len(a.names) {
		return "", ErrPart
	}
	return a.names[part], nil
}

// MustName is like Name but panics with its error.
func (a *Analysis) MustName(part int) string { return must(a.Name(part)) }

// Order returns the declared nominal order, or ErrUnset.
func (a *Analysis) Order() (Order, error) {
	if err := a.check(); err != nil {
		return Unordered, err
	}
	return a.order, nil
}

// MustOrder is like Order but panics with its error.
func (a *Analysis) MustOrder() Order { return must(a.Order()) }

// System returns the borrowed immutable system and whether it is present.
// With no measurements or explicit system it returns nil, false, nil.
func (a *Analysis) System() (*system.System, bool, error) {
	if err := a.check(); err != nil {
		return nil, false, err
	}
	return a.sys, a.sys != nil, nil
}

// MustSystem is like System but panics with its error.
func (a *Analysis) MustSystem() (*system.System, bool) { return must2(a.System()) }

// Pairs returns all canonical pairs in lexicographic order in a new slice.
func (a *Analysis) Pairs() ([]Pair, error) {
	if err := a.check(); err != nil {
		return nil, err
	}
	out := make([]Pair, len(a.data))
	for i, p := range a.data {
		out[i] = p.pair
	}
	return out, nil
}

// MustPairs is like Pairs but panics with its error.
func (a *Analysis) MustPairs() []Pair { return must(a.Pairs()) }

// Verticals returns an independent slice in time order, including rests.
// Errors are ErrUnset, then ErrPart for a noncanonical or unknown pair.
func (a *Analysis) Verticals(pair Pair) ([]Vertical, error) {
	p, err := a.pair(pair)
	if err != nil {
		return nil, err
	}
	return slices.Clone(p.verticals), nil
}

// MustVerticals is like Verticals but panics with its error.
func (a *Analysis) MustVerticals(pair Pair) []Vertical { return must(a.Verticals(pair)) }

func verticalIndex(vs []Vertical, at duration.Position) (int, bool) {
	return slices.BinarySearchFunc(vs, at, func(v Vertical, p duration.Position) int { return v.Start.Compare(p) })
}

// VerticalAt returns the complete segment containing at. At or after the
// joint end it returns the zero Vertical, false, nil. A rest still has ok=true.
// Errors are ErrUnset, ErrPart, then ErrUnset with duration.ErrUnset for at.
func (a *Analysis) VerticalAt(pair Pair, at duration.Position) (Vertical, bool, error) {
	p, err := a.pair(pair)
	if err != nil {
		return Vertical{}, false, err
	}
	if !at.IsSet() {
		return Vertical{}, false, unset("position", duration.ErrUnset)
	}
	i, found := verticalIndex(p.verticals, at)
	if !found {
		i--
	}
	if i < 0 || at.Compare(p.verticals[i].End) >= 0 {
		return Vertical{}, false, nil
	}
	return p.verticals[i], true, nil
}

// MustVerticalAt is like VerticalAt but panics with its error.
func (a *Analysis) MustVerticalAt(pair Pair, at duration.Position) (Vertical, bool) {
	return must2(a.VerticalAt(pair, at))
}

// Transitions returns an independent slice of all adjacent transitions,
// including unclassified transitions interrupted by a rest.
func (a *Analysis) Transitions(pair Pair) ([]Transition, error) {
	p, err := a.pair(pair)
	if err != nil {
		return nil, err
	}
	return slices.Clone(p.transitions), nil
}

// MustTransitions is like Transitions but panics with its error.
func (a *Analysis) MustTransitions(pair Pair) []Transition { return must(a.Transitions(pair)) }

// TransitionAt requires exactly two adjacent increasing vertical starts.
// Validation order is receiver, pair, from, to, adjacency. Unset positions
// match ErrUnset and duration.ErrUnset; invalid adjacency matches ErrTransition.
func (a *Analysis) TransitionAt(pair Pair, from, to duration.Position) (Transition, error) {
	p, err := a.pair(pair)
	if err != nil {
		return Transition{}, err
	}
	if !from.IsSet() {
		return Transition{}, unset("from", duration.ErrUnset)
	}
	if !to.IsSet() {
		return Transition{}, unset("to", duration.ErrUnset)
	}
	i, found := verticalIndex(p.verticals, from)
	if !found || i+1 >= len(p.verticals) || !p.verticals[i+1].Start.Equal(to) {
		return Transition{}, ErrTransition
	}
	return p.transitions[i], nil
}

// MustTransitionAt is like TransitionAt but panics with its error.
func (a *Analysis) MustTransitionAt(pair Pair, from, to duration.Position) Transition {
	return must(a.TransitionAt(pair, from, to))
}
