package mode

import (
	"cmp"
	"iter"
	"slices"

	"github.com/go-muse/musex/system"
)

// DegreesAt returns the numbers of all degrees whose number equals offset, in
// ascending order. Any offset is allowed. The time is O(Len(Degrees)).
func (p Positions) DegreesAt(offset system.Step) []int {
	result := make([]int, 0)
	for i, n := range p.Degrees {
		if n == offset {
			result = append(result, i+1)
		}
	}
	return result
}

// OccupiedSteps yields each distinct degree number once, in ascending order,
// with the ascending numbers of the degrees at that offset. Each degree is in
// exactly one pair. The walk does not visit the integers between distant
// offsets.
//
// Each iterator method copies the degree numbers when the method itself is
// called, not when the walk starts: later changes of p.Degrees, including
// changes made from inside a walk, do not affect the returned iterator. The
// iterator is never nil, can be walked again with the same pairs, and can be
// walked from several goroutines at once. It returns as soon as yield returns
// false; a panic in yield is not recovered. Calling the method while another
// goroutine changes p.Degrees is a data race. Preparing an iterator costs
// O(k log k) time and O(k) memory, where k is Len(Degrees).
func (p Positions) OccupiedSteps() iter.Seq2[system.Step, []int] {
	index := newOffsetIndex(p.Degrees)
	return func(yield func(system.Step, []int) bool) {
		for g, offset := range index.offsets {
			if !yield(offset, index.group(g)) {
				return
			}
		}
	}
}

// OccupiedStepsBackward yields the pairs of OccupiedSteps in reverse order,
// from the largest offset to the smallest. Numbers within a list are still
// ascending. The snapshot and walk rules are those of OccupiedSteps.
func (p Positions) OccupiedStepsBackward() iter.Seq2[system.Step, []int] {
	index := newOffsetIndex(p.Degrees)
	return func(yield func(system.Step, []int) bool) {
		for g := len(index.offsets) - 1; g >= 0; g-- {
			if !yield(index.offsets[g], index.group(g)) {
				return
			}
		}
	}
}

// StepsIn yields a pair for every integer offset of the inclusive range
// [from, to] in ascending order, including free offsets with an empty list.
// The list of each pair equals DegreesAt of its offset. When from > to there
// are no pairs; this is not an error. The range is not prepared in advance:
// the first pair is available at once for a range of any length, and the walk
// ends after to without overflow, also when to is math.MaxInt. The snapshot
// and walk rules are those of OccupiedSteps.
func (p Positions) StepsIn(from, to system.Step) iter.Seq2[system.Step, []int] {
	index := newOffsetIndex(p.Degrees)
	return func(yield func(system.Step, []int) bool) {
		if from > to {
			return
		}
		g, _ := slices.BinarySearch(index.offsets, from)
		for offset := from; ; offset++ {
			list := make([]int, 0)
			if g < len(index.offsets) && index.offsets[g] == offset {
				list = index.group(g)
				g++
			}
			if !yield(offset, list) || offset == to {
				return
			}
		}
	}
}

// StepsInBackward yields the pairs of StepsIn(from, to) in reverse order, from
// to down to from. The bounds are given as for StepsIn: from is the lower and
// to the upper one, and when from > to there are no pairs; the method, not the
// order of the bounds, selects the direction. The walk ends after from without
// overflow, also when from is math.MinInt. The snapshot and walk rules are
// those of OccupiedSteps.
func (p Positions) StepsInBackward(from, to system.Step) iter.Seq2[system.Step, []int] {
	index := newOffsetIndex(p.Degrees)
	return func(yield func(system.Step, []int) bool) {
		if from > to {
			return
		}
		g, found := slices.BinarySearch(index.offsets, to)
		if !found {
			g--
		}
		for offset := to; ; offset-- {
			list := make([]int, 0)
			if g >= 0 && index.offsets[g] == offset {
				list = index.group(g)
				g--
			}
			if !yield(offset, list) || offset == from {
				return
			}
		}
	}
}

// offsetIndex is an immutable snapshot of degree numbers grouped by offset.
// After construction it is only read, so walks can share it.
type offsetIndex struct {
	offsets []system.Step // distinct offsets in ascending order
	bounds  []int         // group g is numbers[bounds[g]:bounds[g+1]]
	numbers []int         // degree numbers ordered by offset, then number
}

func newOffsetIndex(degrees []system.Step) offsetIndex {
	type entry struct {
		offset system.Step
		number int
	}
	entries := make([]entry, len(degrees))
	for i, n := range degrees {
		entries[i] = entry{n, i + 1}
	}
	slices.SortFunc(entries, func(a, b entry) int {
		if c := cmp.Compare(a.offset, b.offset); c != 0 {
			return c
		}
		return cmp.Compare(a.number, b.number)
	})
	index := offsetIndex{
		offsets: make([]system.Step, 0),
		bounds:  make([]int, 0, len(entries)+1),
		numbers: make([]int, len(entries)),
	}
	for i, e := range entries {
		index.numbers[i] = e.number
		if i == 0 || e.offset != entries[i-1].offset {
			index.offsets = append(index.offsets, e.offset)
			index.bounds = append(index.bounds, i)
		}
	}
	index.bounds = append(index.bounds, len(entries))
	return index
}

// group returns a new caller-owned copy of the numbers of group g.
func (x offsetIndex) group(g int) []int {
	return slices.Clone(x.numbers[x.bounds[g]:x.bounds[g+1]])
}
