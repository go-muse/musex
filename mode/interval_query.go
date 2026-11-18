package mode

import (
	"math/big"
	"slices"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/template"
)

// IntervalFamily selects one of three structural seven-degree profiles.
type IntervalFamily uint8

const (
	// FamilyI has two and three major seconds between its minor seconds.
	FamilyI IntervalFamily = iota + 1
	// FamilyII has one and four major seconds between its minor seconds.
	FamilyII
	// FamilyIII has zero and five major seconds between its minor seconds.
	FamilyIII
)

// DegreePlace is an actual degree number (1..Len) and signed cycle count.
type DegreePlace struct {
	Number int
	Cycle  int
}

// CycleWindow selects an inclusive finite range of cycles. Zero selects [0,0].
type CycleWindow struct{ First, Last int }

// IntervalScope selects families and explicit templates. Cycles is required
// for a closed non-octave template and otherwise optionally filters results.
// Input slices and the window are read only during the call.
type IntervalScope struct {
	Families  []IntervalFamily
	Templates []template.Template
	Cycles    *CycleWindow
}

// IntervalWitness explains an exact directed interval occurrence. Family zero
// denotes an explicitly requested template; Rotation is zero-based. From is
// always in cycle zero. The returned slice belongs to the caller.
type IntervalWitness struct {
	Family   IntervalFamily
	Rotation int
	Template template.Template
	From, To DegreePlace
	Interval interval.Interval
}

func (w CycleWindow) valid() bool         { return w.First <= w.Last }
func (w CycleWindow) contains(c int) bool { return c >= w.First && c <= w.Last }
func sourceTemplates(scope IntervalScope) ([]IntervalWitness, error) {
	if len(scope.Families) == 0 && len(scope.Templates) == 0 {
		return nil, ErrEmptyInput
	}
	if scope.Cycles != nil && !scope.Cycles.valid() {
		return nil, ErrInvalidArgument
	}
	var selected [4]bool
	for _, f := range scope.Families {
		if f < FamilyI || f > FamilyIII {
			return nil, ErrInvalidArgument
		}
		selected[f] = true
	}
	for _, t := range scope.Templates {
		if !t.IsValid() {
			return nil, ErrUnset
		}
		cycle, closed := t.Cycle()
		if closed && !cycle.Equal(interval.PerfectOctave()) && scope.Cycles == nil {
			return nil, ErrInvalidArgument
		}
	}
	result := make([]IntervalWitness, 0)
	bases := [3]template.Template{template.Ionian(), template.MelodicMinor(), template.NeapolitanMajor()}
	for f := FamilyI; f <= FamilyIII; f++ {
		if selected[f] {
			for r := range 7 {
				result = append(result, IntervalWitness{Family: f, Rotation: r, Template: bases[f-1].MustRotate(r + 1)})
			}
		}
	}
	var seen []template.Template
	for _, t := range scope.Templates {
		if slices.ContainsFunc(seen, t.Equal) {
			continue
		}
		seen = append(seen, t)
		result = append(result, IntervalWitness{Template: t})
	}
	return result, nil
}

// solveCycle solves both exact coordinates without constructing a virtual
// place number, which could overflow even for a representable cycle count.
func solveCycle(want, base, cycle coordinates) (int, bool) {
	difference := want.sub(base)
	var q *big.Rat
	if cycle.letters.Sign() != 0 {
		q = new(big.Rat).SetFrac(difference.letters, cycle.letters)
	} else {
		if cycle.chromatic.Sign() == 0 {
			return 0, false
		}
		q = new(big.Rat).Quo(difference.chromatic, cycle.chromatic)
	}
	if !q.IsInt() || !q.Num().IsInt64() {
		return 0, false
	}
	n := int(q.Num().Int64())
	return n, base.add(cycle.mul(n)).equal(want)
}

// FindInterval returns structural witnesses in family, rotation, starting
// degree, and ending degree order, then explicit templates in request order.
// Duplicate selections are ignored. It includes zero and full-cycle spans
// and solves cycle counts exactly rather than scanning virtual places.
func FindInterval(want interval.Interval, scope IntervalScope) ([]IntervalWitness, error) {
	if !want.IsValid() {
		return nil, ErrUnset
	}
	sources, err := sourceTemplates(scope)
	if err != nil {
		return nil, err
	}
	target := coordinatesOf(want)
	result := make([]IntervalWitness, 0)
	for _, source := range sources {
		prefixes, err := source.Template.PlaceIntervals()
		if err != nil {
			return nil, templateError(err)
		}
		cycle, closed := source.Template.Cycle()
		var cc coordinates
		if closed {
			cc = coordinatesOf(cycle)
		}
		for i, a := range prefixes {
			for j, b := range prefixes {
				base := coordinatesOf(b).sub(coordinatesOf(a))
				n := 0
				ok := base.equal(target)
				if closed {
					n, ok = solveCycle(target, base, cc)
				}
				if !ok || (scope.Cycles != nil && !scope.Cycles.contains(n)) {
					continue
				}
				w := source
				w.From = DegreePlace{i + 1, 0}
				w.To = DegreePlace{j + 1, n}
				w.Interval = want
				result = append(result, w)
			}
		}
	}
	return result, nil
}

// MustFindInterval panics with the error returned by FindInterval.
func MustFindInterval(want interval.Interval, scope IntervalScope) []IntervalWitness {
	return must(FindInterval(want, scope))
}
