package mode_test

import (
	"iter"
	"math"
	"slices"
	"sync"
	"testing"

	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The expected pairs below come from business-spec 21.2, KP-14, and from the
// rules P-34 and F-10. They are not derived from the code under test.

type offsetPair struct {
	Offset  system.Step
	Degrees []int
}

// pair builds an expected pair; an empty list is a non-nil slice.
func pair(offset system.Step, degrees ...int) offsetPair {
	if degrees == nil {
		degrees = []int{}
	}
	return offsetPair{Offset: offset, Degrees: degrees}
}

// walkPairs collects at most limit pairs. A walk that offers more pairs is
// stopped and reported as a failure, so a walk that never ends cannot hang
// the test. It also fails when yield is called again after returning false.
func walkPairs(t *testing.T, seq iter.Seq2[system.Step, []int], limit int) []offsetPair {
	t.Helper()
	require.NotNil(t, seq, "iterator methods return a non-nil iterator")
	pairs := []offsetPair{}
	stopped := false
	seq(func(offset system.Step, degrees []int) bool {
		if stopped {
			t.Errorf("yield called again after it returned false")
			return false
		}
		pairs = append(pairs, offsetPair{Offset: offset, Degrees: degrees})
		if len(pairs) > limit {
			stopped = true
			return false
		}
		return true
	})
	require.LessOrEqual(t, len(pairs), limit, "the walk did not end after %d pairs", limit)
	return pairs
}

// firstPairs stops the walk after count pairs and counts the calls of yield.
func firstPairs(seq iter.Seq2[system.Step, []int], count int) (pairs []offsetPair, calls int) {
	seq(func(offset system.Step, degrees []int) bool {
		calls++
		pairs = append(pairs, offsetPair{Offset: offset, Degrees: degrees})
		return len(pairs) < count
	})
	return pairs, calls
}

// collectPairs gathers a finite walk without any testing helper, so it is
// safe in worker goroutines.
func collectPairs(seq iter.Seq2[system.Step, []int]) []offsetPair {
	pairs := []offsetPair{}
	for offset, degrees := range seq {
		pairs = append(pairs, offsetPair{Offset: offset, Degrees: degrees})
	}
	return pairs
}

func reversedPairs(pairs []offsetPair) []offsetPair {
	result := slices.Clone(pairs)
	slices.Reverse(result)
	return result
}

func positionsOf(degrees ...system.Step) mode.Positions {
	return mode.Positions{Degrees: degrees}
}

func ionianTraditional(t *testing.T) mode.Positions {
	t.Helper()
	p := mode.MustFromCode("ionian").MustNumbers(system.Traditional())
	require.Equal(t, []system.Step{0, 2, 4, 5, 7, 9, 11}, p.Degrees)
	return p
}

func TestDegreesAtControlValues(t *testing.T) {
	t.Parallel()
	traditional := system.Traditional()
	q24 := mode.MustFromCode("ionian").MustNumbers(system.QuarterTones24())
	require.Equal(t, []system.Step{0, 4, 8, 10, 14, 18, 22}, q24.Degrees)
	irregular := mode.MustFromCode("ionian").MustNumbersFrom(irregularSystem(), spelling.MustParse("D"))
	require.Equal(t, []system.Step{0, 1, 3, 4, 6, 8, 10}, irregular.Degrees)
	major := ionianTraditional(t)
	unisonPair := mode.MustParse("d2 A7").MustNumbers(traditional)
	triple := mode.MustParse("A1 dd2 A7").MustNumbers(traditional)
	below := mode.MustParse("dd2 AA7").MustNumbers(traditional)
	openThird := mode.MustParse("m3 M2 |").MustNumbers(traditional)
	for _, tt := range []struct {
		name   string
		p      mode.Positions
		offset system.Step
		want   []int
	}{
		{"issue: ionian at 4", major, 4, []int{3}},
		{"ionian at a free offset between degrees", major, 1, []int{}},
		{"ionian at 0", major, 0, []int{1}},
		{"ionian at the last degree", major, 11, []int{7}},
		{"ionian at the cycle boundary", major, 12, []int{}},
		{"ionian at a negative offset", major, -1, []int{}},
		{"ionian far away", major, math.MaxInt, []int{}},
		{"ionian far below", major, math.MinInt, []int{}},
		{"issue: d2 A7 at 0", unisonPair, 0, []int{1, 2}},
		{"d2 A7 at the cycle boundary", unisonPair, 12, []int{}},
		{"A1 dd2 A7, non-adjacent degrees at 0", triple, 0, []int{1, 3}},
		{"A1 dd2 A7 at 1", triple, 1, []int{2}},
		{"dd2 AA7 at -1", below, -1, []int{2}},
		{"24 positions at 8", q24, 8, []int{3}},
		{"24 positions at 4", q24, 4, []int{2}},
		{"24 positions at 2", q24, 2, []int{}},
		{"irregular from D at 1", irregular, 1, []int{2}},
		{"irregular from D at 2", irregular, 2, []int{}},
		{"open mode at 5", openThird, 5, []int{3}},
		{"empty value", mode.Positions{}, 0, []int{}},
		{"empty non-nil degrees", positionsOf(), 0, []int{}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.p.DegreesAt(tt.offset)
			assert.NotNil(t, got, "an empty result is a non-nil slice")
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestOffsetWalksControlValues(t *testing.T) {
	t.Parallel()
	traditional := system.Traditional()
	major := ionianTraditional(t)
	unisonPair := mode.MustParse("d2 A7").MustNumbers(traditional)
	triple := mode.MustParse("A1 dd2 A7").MustNumbers(traditional)
	below := mode.MustParse("dd2 AA7").MustNumbers(traditional)
	openThird := mode.MustParse("m3 M2 |").MustNumbers(traditional)
	empty := mode.Positions{}
	const maxStep, minStep = system.Step(math.MaxInt), system.Step(math.MinInt)
	majorOccupied := []offsetPair{pair(0, 1), pair(2, 2), pair(4, 3), pair(5, 4), pair(7, 5), pair(9, 6), pair(11, 7)}

	for _, tt := range []struct {
		name string
		seq  iter.Seq2[system.Step, []int]
		want []offsetPair
	}{
		{"issue: ionian occupied offsets", major.OccupiedSteps(), majorOccupied},
		{"issue: ionian offsets 0 to 4", major.StepsIn(0, 4), []offsetPair{pair(0, 1), pair(1), pair(2, 2), pair(3), pair(4, 3)}},
		{"ionian single occupied offset", major.StepsIn(4, 4), []offsetPair{pair(4, 3)}},
		{"ionian single free offset", major.StepsIn(1, 1), []offsetPair{pair(1)}},
		{"ionian from above to", major.StepsIn(5, 4), []offsetPair{}},
		{"ionian across the cycle boundary", major.StepsIn(11, 13), []offsetPair{pair(11, 7), pair(12), pair(13)}},
		{"ionian below the first degree", major.StepsIn(-2, 0), []offsetPair{pair(-2), pair(-1), pair(0, 1)}},
		{"ionian occupied offsets downward", major.OccupiedStepsBackward(), reversedPairs(majorOccupied)},
		{"ionian offsets 0 to 4 downward", major.StepsInBackward(0, 4), []offsetPair{pair(4, 3), pair(3), pair(2, 2), pair(1), pair(0, 1)}},
		{"ionian single offset downward", major.StepsInBackward(4, 4), []offsetPair{pair(4, 3)}},
		{"ionian from above to downward", major.StepsInBackward(5, 4), []offsetPair{}},
		{"issue: d2 A7 occupied offsets", unisonPair.OccupiedSteps(), []offsetPair{pair(0, 1, 2)}},
		{"issue: d2 A7 offsets 0 to 1", unisonPair.StepsIn(0, 1), []offsetPair{pair(0, 1, 2), pair(1)}},
		{"d2 A7 occupied offsets downward", unisonPair.OccupiedStepsBackward(), []offsetPair{pair(0, 1, 2)}},
		{"A1 dd2 A7 occupied offsets", triple.OccupiedSteps(), []offsetPair{pair(0, 1, 3), pair(1, 2)}},
		{"dd2 AA7 occupied offsets follow offsets, not numbers", below.OccupiedSteps(), []offsetPair{pair(-1, 2), pair(0, 1)}},
		{"dd2 AA7 occupied offsets downward", below.OccupiedStepsBackward(), []offsetPair{pair(0, 1), pair(-1, 2)}},
		{"open mode occupied offsets", openThird.OccupiedSteps(), []offsetPair{pair(0, 1), pair(3, 2), pair(5, 3)}},
		{"empty value, occupied", empty.OccupiedSteps(), []offsetPair{}},
		{"empty value, occupied downward", empty.OccupiedStepsBackward(), []offsetPair{}},
		{"empty value, offsets 0 to 2", empty.StepsIn(0, 2), []offsetPair{pair(0), pair(1), pair(2)}},
		{"empty value, offsets 0 to 2 downward", empty.StepsInBackward(0, 2), []offsetPair{pair(2), pair(1), pair(0)}},
		{"empty non-nil degrees, occupied", positionsOf().OccupiedSteps(), []offsetPair{}},
		{"far occupied offsets", positionsOf(0, 1_000_000_000_000).OccupiedSteps(), []offsetPair{pair(0, 1), pair(1_000_000_000_000, 2)}},
		{"far occupied offsets downward", positionsOf(0, 1_000_000_000_000).OccupiedStepsBackward(), []offsetPair{pair(1_000_000_000_000, 2), pair(0, 1)}},
		{"ends after MaxInt", major.StepsIn(maxStep-1, maxStep), []offsetPair{pair(maxStep - 1), pair(maxStep)}},
		{"single MinInt offset", major.StepsIn(minStep, minStep), []offsetPair{pair(minStep)}},
		{"starts at MinInt", major.StepsIn(minStep, minStep+1), []offsetPair{pair(minStep), pair(minStep + 1)}},
		{"downward ends after MinInt", major.StepsInBackward(minStep, minStep+1), []offsetPair{pair(minStep + 1), pair(minStep)}},
		{"single MaxInt offset downward", major.StepsInBackward(maxStep, maxStep), []offsetPair{pair(maxStep)}},
		{"downward starts at MaxInt", major.StepsInBackward(maxStep-1, maxStep), []offsetPair{pair(maxStep), pair(maxStep - 1)}},
		{"MinInt above MaxInt gives no pairs", major.StepsIn(maxStep, minStep), []offsetPair{}},
		{"MinInt above MaxInt gives no pairs downward", major.StepsInBackward(maxStep, minStep), []offsetPair{}},
		{"occupied extreme offsets", positionsOf(maxStep, minStep, 0).OccupiedSteps(), []offsetPair{pair(minStep, 2), pair(0, 3), pair(maxStep, 1)}},
		{"occupied extreme offsets downward", positionsOf(maxStep, minStep, 0).OccupiedStepsBackward(), []offsetPair{pair(maxStep, 1), pair(0, 3), pair(minStep, 2)}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, walkPairs(t, tt.seq, len(tt.want)+2))
		})
	}
}

func TestOffsetWalksCoverAWholeCycle(t *testing.T) {
	t.Parallel()
	major := ionianTraditional(t)
	all := walkPairs(t, major.StepsIn(0, 11), 20)
	require.Len(t, all, 12)
	var occupied []offsetPair
	for i, p := range all {
		assert.Equal(t, system.Step(i), p.Offset)
		if len(p.Degrees) > 0 {
			occupied = append(occupied, p)
		}
	}
	assert.Equal(t, walkPairs(t, major.OccupiedSteps(), 20), occupied)
	assert.Equal(t, reversedPairs(all), walkPairs(t, major.StepsInBackward(0, 11), 20))
}

// TestOffsetWalksAgreeWithDegreesAt checks F-10 on accepted numbers and on
// values built by hand: unsorted, repeated, negative, and with unrelated
// Adjacent, Cycle, and HasCycle fields.
func TestOffsetWalksAgreeWithDegreesAt(t *testing.T) {
	t.Parallel()
	traditional := system.Traditional()
	for _, tt := range []struct {
		name string
		p    mode.Positions
	}{
		{"ionian", ionianTraditional(t)},
		{"d2 A7", mode.MustParse("d2 A7").MustNumbers(traditional)},
		{"A1 dd2 A7", mode.MustParse("A1 dd2 A7").MustNumbers(traditional)},
		{"dd2 AA7", mode.MustParse("dd2 AA7").MustNumbers(traditional)},
		{"open", mode.MustParse("m3 M2 |").MustNumbers(traditional)},
		{"unsorted, repeated, and negative", positionsOf(5, -3, 5, 0, -3, 7)},
		{"single degree", positionsOf(-4)},
		{"unrelated fields", mode.Positions{Degrees: []system.Step{9, 2, 2}, Adjacent: []system.Step{100, -7}, Cycle: -40, HasCycle: true}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			low, high := slices.Min(tt.p.Degrees), slices.Max(tt.p.Degrees)
			window := walkPairs(t, tt.p.StepsIn(low-2, high+2), int(high-low)+10)
			require.Len(t, window, int(high-low)+5)
			for i, got := range window {
				assert.Equal(t, low-2+system.Step(i), got.Offset)
				assert.Equal(t, tt.p.DegreesAt(got.Offset), got.Degrees, "offset %d", got.Offset)
			}
			var nonEmpty []offsetPair
			for _, got := range walkPairs(t, tt.p.StepsIn(low, high), int(high-low)+10) {
				if len(got.Degrees) > 0 {
					nonEmpty = append(nonEmpty, got)
				}
			}
			occupied := walkPairs(t, tt.p.OccupiedSteps(), len(tt.p.Degrees)+2)
			assert.Equal(t, nonEmpty, occupied)
			var numbers []int
			for _, got := range occupied {
				assert.True(t, slices.IsSorted(got.Degrees), "numbers ascend within a pair")
				numbers = append(numbers, got.Degrees...)
			}
			slices.Sort(numbers)
			want := make([]int, len(tt.p.Degrees))
			for i := range want {
				want[i] = i + 1
			}
			assert.Equal(t, want, numbers, "every degree is in exactly one occupied pair")
			assert.Equal(t, reversedPairs(occupied), walkPairs(t, tt.p.OccupiedStepsBackward(), len(tt.p.Degrees)+2))
			assert.Equal(t, reversedPairs(window), walkPairs(t, tt.p.StepsInBackward(low-2, high+2), int(high-low)+10))

			bare := mode.Positions{Degrees: slices.Clone(tt.p.Degrees)}
			assert.Equal(t, occupied, walkPairs(t, bare.OccupiedSteps(), len(tt.p.Degrees)+2), "only Degrees is read")
			assert.Equal(t, window, walkPairs(t, bare.StepsIn(low-2, high+2), int(high-low)+10), "only Degrees is read")
		})
	}
	t.Run("hand-built value", func(t *testing.T) {
		t.Parallel()
		p := positionsOf(5, -3, 5, 0, -3, 7)
		assert.Equal(t, []offsetPair{pair(-3, 2, 5), pair(0, 4), pair(5, 1, 3), pair(7, 6)}, walkPairs(t, p.OccupiedSteps(), 10))
		assert.Equal(t, []int{2, 5}, p.DegreesAt(-3))
		assert.Equal(t, []int{1, 3}, p.DegreesAt(5))
	})
}

func TestOffsetWalksStopEarlyAndLazily(t *testing.T) {
	t.Parallel()
	const far = system.Step(1_000_000_000_000)
	p := positionsOf(0, far)
	major := ionianTraditional(t)
	for _, tt := range []struct {
		name string
		seq  iter.Seq2[system.Step, []int]
		want offsetPair
	}{
		{"long forward walk", p.StepsIn(0, far), pair(0, 1)},
		{"long backward walk", p.StepsInBackward(0, far), pair(far, 2)},
		{"whole int range forward", major.StepsIn(math.MinInt, math.MaxInt), pair(math.MinInt)},
		{"whole int range backward", major.StepsInBackward(math.MinInt, math.MaxInt), pair(math.MaxInt)},
		{"occupied", major.OccupiedSteps(), pair(0, 1)},
		{"occupied backward", major.OccupiedStepsBackward(), pair(11, 7)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pairs, calls := firstPairs(tt.seq, 1)
			assert.Equal(t, 1, calls, "no yield after it returned false")
			assert.Equal(t, []offsetPair{tt.want}, pairs)
		})
	}
	t.Run("stop after the third pair", func(t *testing.T) {
		t.Parallel()
		pairs, calls := firstPairs(major.StepsIn(0, 11), 3)
		assert.Equal(t, 3, calls)
		assert.Equal(t, []offsetPair{pair(0, 1), pair(1), pair(2, 2)}, pairs)
	})
	t.Run("far occupied offsets give exactly two pairs", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, []offsetPair{pair(0, 1), pair(far, 2)}, walkPairs(t, p.OccupiedSteps(), 2))
	})
}

func TestOffsetResultsAreOwnedByTheCaller(t *testing.T) {
	t.Parallel()
	traditional := system.Traditional()
	t.Run("DegreesAt", func(t *testing.T) {
		t.Parallel()
		p := mode.MustParse("d2 A7").MustNumbers(traditional)
		first := p.DegreesAt(0)
		require.Equal(t, []int{1, 2}, first)
		first[0] = 99
		_ = append(first[:1], 77)
		assert.Equal(t, []int{1, 2}, p.DegreesAt(0))
		assert.Equal(t, []system.Step{0, 0}, p.Degrees)
		empty := p.DegreesAt(5)
		_ = append(empty, 3)
		assert.Equal(t, []int{}, p.DegreesAt(5))
	})
	p := mode.MustParse("A1 dd2 A7").MustNumbers(traditional)
	want := []offsetPair{pair(-1), pair(0, 1, 3), pair(1, 2), pair(2)}
	for _, tt := range []struct {
		name string
		seq  func() iter.Seq2[system.Step, []int]
		want []offsetPair
	}{
		{"StepsIn", func() iter.Seq2[system.Step, []int] { return p.StepsIn(-1, 2) }, want},
		{"StepsInBackward", func() iter.Seq2[system.Step, []int] { return p.StepsInBackward(-1, 2) }, reversedPairs(want)},
		{"OccupiedSteps", p.OccupiedSteps, []offsetPair{pair(0, 1, 3), pair(1, 2)}},
		{"OccupiedStepsBackward", p.OccupiedStepsBackward, []offsetPair{pair(1, 2), pair(0, 1, 3)}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			seq := tt.seq()
			for i := range tt.want {
				pairs := walkPairs(t, seq, len(tt.want))
				// Change one list in place and through append; the other lists
				// of the same walk, later walks, and the value stay intact.
				for j := range pairs[i].Degrees {
					pairs[i].Degrees[j] = -9
				}
				pairs[i].Degrees = append(pairs[i].Degrees[:0], -5, -6, -7, -8)
				for j, other := range pairs {
					if j != i {
						assert.Equal(t, tt.want[j], other, "pair %d after changing pair %d", j, i)
					}
				}
				assert.Equal(t, tt.want, walkPairs(t, seq, len(tt.want)), "a repeated walk")
				assert.Equal(t, tt.want, walkPairs(t, tt.seq(), len(tt.want)), "a new iterator")
			}
			assert.Equal(t, []system.Step{0, 1, 0}, p.Degrees)
			assert.Equal(t, []int{1, 3}, p.DegreesAt(0))
		})
	}
}

func TestOffsetIteratorsSnapshotTheValue(t *testing.T) {
	t.Parallel()
	type method struct {
		name string
		seq  func(mode.Positions) iter.Seq2[system.Step, []int]
		want []offsetPair
	}
	methods := []method{
		{"OccupiedSteps", mode.Positions.OccupiedSteps, []offsetPair{pair(0, 1), pair(2, 2), pair(4, 3)}},
		{"OccupiedStepsBackward", mode.Positions.OccupiedStepsBackward, []offsetPair{pair(4, 3), pair(2, 2), pair(0, 1)}},
		{"StepsIn", func(p mode.Positions) iter.Seq2[system.Step, []int] { return p.StepsIn(0, 2) }, []offsetPair{pair(0, 1), pair(1), pair(2, 2)}},
		{"StepsInBackward", func(p mode.Positions) iter.Seq2[system.Step, []int] { return p.StepsInBackward(0, 2) }, []offsetPair{pair(2, 2), pair(1), pair(0, 1)}},
	}
	for _, m := range methods {
		t.Run(m.name+" after the call", func(t *testing.T) {
			t.Parallel()
			p := positionsOf(0, 2, 4)
			seq := m.seq(p)
			p.Degrees[0], p.Degrees[1], p.Degrees[2] = 1, 1, 1
			assert.Equal(t, m.want, walkPairs(t, seq, len(m.want)+2))
			p.Degrees = nil
			assert.Equal(t, m.want, walkPairs(t, seq, len(m.want)+2))
		})
		t.Run(m.name+" inside the walk", func(t *testing.T) {
			t.Parallel()
			p := positionsOf(0, 2, 4)
			seq := m.seq(p)
			var got []offsetPair
			for offset, degrees := range seq {
				got = append(got, offsetPair{Offset: offset, Degrees: degrees})
				for i := range p.Degrees {
					p.Degrees[i]++
				}
				p.Degrees = append(p.Degrees, 1)
			}
			assert.Equal(t, m.want, got)
			assert.Equal(t, m.want, walkPairs(t, seq, len(m.want)+2), "a repeated walk")
		})
	}
	t.Run("value receiver on a temporary result", func(t *testing.T) {
		t.Parallel()
		seq := mode.MustParse("m3 M2 |").MustNumbers(system.Traditional()).OccupiedStepsBackward()
		assert.Equal(t, []offsetPair{pair(5, 3), pair(3, 2), pair(0, 1)}, walkPairs(t, seq, 5))
	})
}

func TestOffsetIteratorsConcurrentWalks(t *testing.T) {
	t.Parallel()
	p := mode.MustParse("A1 dd2 A7").MustNumbers(system.Traditional())
	seqs := []iter.Seq2[system.Step, []int]{p.OccupiedSteps(), p.OccupiedStepsBackward(), p.StepsIn(-3, 3), p.StepsInBackward(-3, 3)}
	wants := make([][]offsetPair, len(seqs))
	for i, seq := range seqs {
		wants[i] = collectPairs(seq)
	}
	var wg sync.WaitGroup
	for range 4 {
		for i, seq := range seqs {
			wg.Go(func() {
				for range 20 {
					assert.Equal(t, wants[i], collectPairs(seq))
					assert.Equal(t, []int{1, 3}, p.DegreesAt(0))
				}
			})
		}
	}
	wg.Wait()
}
