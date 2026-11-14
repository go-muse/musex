package meter_test

import (
	"math"
	"slices"
	"testing"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/meter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUnits covers О-5, ИР-2, Т-3, КП-2, КП-41, КП-50: units of every part
// follow each other from the start of the measure, with or without grouping.
func TestUnits(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		text  string
		units []meter.Span
	}{
		{"КП-2 6/8", equalSpans(6, "1/8")},
		{"КП-2 3/4", equalSpans(3, "1/4")},
		{"КП-2 2/4+3/8", sps("1/4@0", "1/4@1/4", "1/8@1/2", "1/8@5/8", "1/8@3/4")},
		{"КП-50 4/3", equalSpans(4, "1/3")},
		{"КП-41 1/1", sps("1@0")},
		{"КП-41 3/1", equalSpans(3, "1")},
		{"C", equalSpans(4, "1/4")},
		{"C|", equalSpans(2, "1/2")},
		{"(3+2+3)/8", equalSpans(8, "1/8")},
		{"without grouping 5/8", equalSpans(5, "1/8")},
		{"without grouping 5/8+2/4", sps("1/8@0", "1/8@1/8", "1/8@1/4", "1/8@3/8", "1/8@1/2", "1/4@5/8", "1/4@7/8")},
		{"1024/8", equalSpans(1024, "1/8")},
	} {
		t.Run(c.text, func(t *testing.T) {
			t.Parallel()

			m := mt(lastWord(c.text))

			units, err := m.Units()
			require.NoError(t, err)
			assert.Equal(t, c.units, units)
			assert.Equal(t, len(c.units), noErr(m.UnitCount()))
		})
	}
}

// TestUnit covers О-5, П-3, Т-3, КП-3, КП-31, КП-33, КП-45, КП-51 and the
// order of checks of Unit and UnitWritten.
func TestUnit(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		text    string
		unit    string
		written string // "" for E-6
	}{
		{"КП-3 3/4", "1/4", "4"},
		{"КП-3 6/8", "1/8", "8"},
		{"КП-3 2/2", "1/2", "2"},
		{"КП-3 3/1", "1", "1"},
		{"КП-3 6/1024", "1/1024", "1024"},
		{"КП-3 4/3", "1/3", ""},
		{"КП-33 3/2048", "1/2048", ""},
		{"КП-51 6/6", "1/6", ""},
		{"C", "1/4", "4"},
		{"C|", "1/2", "2"},
		{"КП-45 2/8+3/8", "1/8", "8"},
		{"(3+2+3)/8", "1/8", "8"},
		{"5/8[2+3]", "1/8", "8"},
		{"1/" + maxIntText(), "1/" + maxIntText(), ""},
	} {
		t.Run(c.text, func(t *testing.T) {
			t.Parallel()

			m := mt(lastWord(c.text))

			unit, err := m.Unit()
			require.NoError(t, err)
			assert.Equal(t, durT(c.unit), unit)

			w, err := m.UnitWritten()
			if c.written == "" {
				assertErr(t, err, meter.ErrNotNotatable, duration.ErrNotNotatable)
				assert.Equal(t, duration.Written{}, w)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, wrs(c.written)[0], w)
			assert.False(t, w.MustHasDots(), "the base without dots")
		})
	}

	for _, text := range []string{"2/4+3/8", "2/4+2/3", "(2+3)/8+2/4"} {
		t.Run("КП-31 "+text, func(t *testing.T) {
			t.Parallel()

			m := mt(text)

			unit, err := m.Unit()
			assertErr(t, err, meter.ErrNoCommonDenominator)
			assert.Equal(t, duration.Duration{}, unit)

			w, err := m.UnitWritten()
			assertErr(t, err, meter.ErrNoCommonDenominator)
			assert.Equal(t, duration.Written{}, w)
		})
	}
}

// TestStandardGrouping covers ИР-3, П-8, КП-4, КП-5: the standard grouping
// of the number of units, and its absence.
func TestStandardGrouping(t *testing.T) {
	t.Parallel()

	threes := func(n int) []int {
		out := make([]int, n)
		for i := range out {
			out[i] = 3
		}

		return out
	}

	for _, c := range []struct {
		text     string
		grouping []int // nil: no grouping
	}{
		{"2/4", []int{2}},
		{"3/4", []int{3}},
		{"4/4", []int{2, 2}},
		{"6/8", []int{3, 3}},
		{"9/8", threes(3)},
		{"12/8", threes(4)},
		{"6/4", []int{3, 3}},
		{"15/8", threes(5)},
		{"18/8", threes(6)},
		{"1023/8", threes(341)},
		{"1/4", []int{1}},
		{"1/1", []int{1}},
		{"C", []int{2, 2}},
		{"C|", []int{2}},
		{"3/2048", []int{3}},
		{"4/3", []int{2, 2}},
		{"5/8", nil},
		{"7/8", nil},
		{"8/8", nil},
		{"10/8", nil},
		{"11/8", nil},
		{"13/8", nil},
		{"14/8", nil},
		{"16/16", nil},
		{"1024/8", nil},
	} {
		t.Run(c.text, func(t *testing.T) {
			t.Parallel()

			m := mt(c.text)
			assert.Equal(t, c.grouping != nil, noErr(m.HasGrouping()))
			assert.False(t, noErr(m.HasExplicitGrouping()))

			grouping, err := m.Grouping()
			if c.grouping == nil {
				assertErr(t, err, meter.ErrGroupingRequired)
				assert.Nil(t, grouping)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, c.grouping, grouping)
		})
	}
}

// TestGroups covers О-6, ИР-4, Т-4, КП-6, КП-43, КП-44, КП-50.
func TestGroups(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name, text string
		grouping   []int
		groups     []meter.Span
		explicit   bool
	}{
		{"КП-6 4/4", "4/4", []int{2, 2}, sps("1/2@0", "1/2@1/2"), false},
		{"КП-6 6/8", "6/8", []int{3, 3}, sps("3/8@0", "3/8@3/8"), false},
		{"КП-6 (3+2+3)/8", "(3+2+3)/8", []int{3, 2, 3}, sps("3/8@0", "1/4@3/8", "3/8@5/8"), false},
		{"КП-6 5/8[2+3]", "5/8[2+3]", []int{2, 3}, sps("1/4@0", "3/8@1/4"), true},
		{"КП-6 5/8[3+2]", "5/8[3+2]", []int{3, 2}, sps("3/8@0", "1/4@3/8"), true},
		{"КП-6 2/4+3/8", "2/4+3/8", []int{2, 3}, sps("1/2@0", "3/8@1/2"), false},
		{"КП-6 4/4+3/8", "4/4+3/8", []int{2, 2, 3}, sps("1/2@0", "1/2@1/2", "3/8@1"), false},
		{"КП-6 (2+3)/8+2/4", "(2+3)/8+2/4", []int{2, 3, 2}, sps("1/4@0", "3/8@1/4", "1/2@5/8"), false},
		{"КП-43 5/8[5]", "5/8[5]", []int{5}, sps("5/8@0"), true},
		{"КП-44 C[3+1]", "C[3+1]", []int{3, 1}, sps("3/4@0", "1/4@3/4"), true},
		{"КП-50 4/3", "4/3", []int{2, 2}, sps("2/3@0", "2/3@2/3"), false},
		{"3/4", "3/4", []int{3}, sps("3/4@0"), false},
		{"C", "C", []int{2, 2}, sps("1/2@0", "1/2@1/2"), false},
		{"C|", "C|", []int{2}, sps("1@0"), false},
		{"6/8[2+2+2]", "6/8[2+2+2]", []int{2, 2, 2}, sps("1/4@0", "1/4@1/4", "1/4@1/2"), true},
		{"2/4+2/8", "2/4+2/8", []int{2, 2}, sps("1/2@0", "1/4@1/2"), false},
		{"sum of 6/8 and 3/4", "6/8+3/4", []int{3, 3, 3}, sps("3/8@0", "3/8@3/8", "3/4@3/4"), false},
		{"sum with 4/4 part", "2/8+4/4", []int{2, 2, 2}, sps("1/4@0", "1/2@1/4", "1/2@3/4"), false},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			m := mt(c.text)
			assert.True(t, noErr(m.HasGrouping()))
			assert.Equal(t, c.explicit, noErr(m.HasExplicitGrouping()))
			assert.Equal(t, c.grouping, noErr(m.Grouping()))
			assert.Equal(t, len(c.groups), noErr(m.GroupCount()))
			assert.Equal(t, c.groups, noErr(m.Groups()))
		})
	}
}

// TestGroupsWritten covers О-6, П-3, КП-8, КП-33.
func TestGroupsWritten(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		text    string
		written []string // nil: E-6
	}{
		{"4/4", []string{"2", "2"}},
		{"6/8", []string{"4.", "4."}},
		{"(3+2+3)/8", []string{"4.", "4", "4."}},
		{"7/8[7]", []string{"2.."}},
		{"2/4+3/8", []string{"2", "4."}},
		{"5/8[2+3]", []string{"4", "4."}},
		{"6/6", []string{"2", "2"}},
		{"6/3", []string{"1", "1"}},
		{"C|", []string{"1"}},
		{"КП-33 7/8[5+2]", nil},
		{"7/8[2+5]", nil},
		{"5/8[5]", nil},
		{"4/3", nil},
	} {
		t.Run(c.text, func(t *testing.T) {
			t.Parallel()

			got, err := mt(lastWord(c.text)).GroupsWritten()
			if c.written == nil {
				assertErr(t, err, meter.ErrNotNotatable, duration.ErrNotNotatable)
				assert.Nil(t, got, "no partial list")

				return
			}

			require.NoError(t, err)
			assert.Equal(t, wrs(c.written...), got)
		})
	}
}

// TestBeats covers О-8, ИР-5, П-1, П-12, Т-5, КП-7, КП-10, КП-41, КП-50, КП-51.
func TestBeats(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name, text string
		beats      []meter.Span
	}{
		{"КП-7 3/4", "3/4", equalSpans(3, "1/4")},
		{"КП-7 4/4", "4/4", equalSpans(4, "1/4")},
		{"КП-7 2/2", "2/2", equalSpans(2, "1/2")},
		{"КП-7 3/8", "3/8", equalSpans(3, "1/8")},
		{"КП-7 6/8", "6/8", equalSpans(2, "3/8")},
		{"КП-7 9/8", "9/8", equalSpans(3, "3/8")},
		{"КП-7 12/8", "12/8", equalSpans(4, "3/8")},
		{"КП-7 6/4", "6/4", equalSpans(2, "3/4")},
		{"КП-7 6/16", "6/16", equalSpans(2, "3/16")},
		{"КП-7 5/8[2+3]", "5/8[2+3]", sps("1/4@0", "3/8@1/4")},
		{"КП-7 (3+2+3)/8", "(3+2+3)/8", sps("3/8@0", "1/4@3/8", "3/8@5/8")},
		{"КП-7 6/8[2+2+2]", "6/8[2+2+2]", equalSpans(3, "1/4")},
		{"КП-10 5/4[3+2]", "5/4[3+2]", sps("3/4@0", "1/2@3/4")},
		{"КП-10 (2+2)/4", "(2+2)/4", equalSpans(4, "1/4")},
		{"КП-10 2/4+2/4", "2/4+2/4", equalSpans(4, "1/4")},
		{"КП-10 (3+3)/8", "(3+3)/8", equalSpans(2, "3/8")},
		{"КП-10 2/4+3/8", "2/4+3/8", sps("1/2@0", "3/8@1/2")},
		{"КП-10 2/4+2/8", "2/4+2/8", sps("1/2@0", "1/4@1/2")},
		{"КП-10 3/8+3/4", "3/8+3/4", sps("3/8@0", "3/4@3/8")},
		{"КП-10 (2+2+2+2)/8", "(2+2+2+2)/8", equalSpans(4, "1/4")},
		{"КП-41 1/1", "1/1", sps("1@0")},
		{"КП-41 3/1", "3/1", equalSpans(3, "1")},
		{"КП-43 5/8[5]", "5/8[5]", sps("5/8@0")},
		{"КП-50 4/3", "4/3", equalSpans(4, "1/3")},
		{"КП-51 6/6", "6/6", equalSpans(2, "1/2")},
		{"КП-51 3/6", "3/6", equalSpans(3, "1/6")},
		{"C", "C", equalSpans(4, "1/4")},
		{"C|", "C|", equalSpans(2, "1/2")},
		{"C[3+1]", "C[3+1]", sps("3/4@0", "1/4@3/4")},
		{"4/4[3+1]", "4/4[3+1]", sps("3/4@0", "1/4@3/4")},
		{"3/4[2+1]", "3/4[2+1]", sps("1/2@0", "1/4@1/2")},
		{"1/4: one group of one unit", "1/4", sps("1/4@0")},
		{"(4+4)/8", "(4+4)/8", equalSpans(2, "1/2")},
		{"15/8", "15/8", equalSpans(5, "3/8")},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			m := mt(c.text)

			beats, err := m.Beats()
			require.NoError(t, err)
			assert.Equal(t, c.beats, beats)
			assert.Equal(t, len(c.beats), noErr(m.BeatCount()))
		})
	}
}

// TestBeatsWritten covers О-8, П-3, КП-8, КП-33, КП-51.
func TestBeatsWritten(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		text    string
		written []string // nil: E-6
	}{
		{"3/4", []string{"4", "4", "4"}},
		{"6/8", []string{"4.", "4."}},
		{"6/4", []string{"2.", "2."}},
		{"6/16", []string{"8.", "8."}},
		{"6/1", []string{"breve.", "breve."}},
		{"5/8[2+3]", []string{"4", "4."}},
		{"(3+2+3)/8", []string{"4.", "4", "4."}},
		{"7/8[7]", []string{"2.."}},
		{"6/6", []string{"2", "2"}},
		{"6/3", []string{"1", "1"}},
		{"C", []string{"4", "4", "4", "4"}},
		{"5/8[5]", nil},
		{"4/3", nil},
		{"3/6", nil},
		{"7/8[5+2]", nil},
	} {
		t.Run(c.text, func(t *testing.T) {
			t.Parallel()

			got, err := mt(c.text).BeatsWritten()
			if c.written == nil {
				assertErr(t, err, meter.ErrNotNotatable, duration.ErrNotNotatable)
				assert.Nil(t, got, "no partial list")

				return
			}

			require.NoError(t, err)
			assert.Equal(t, wrs(c.written...), got)
		})
	}
}

// TestGroupingRequired covers СВ-30, КП-5, КП-30 for the operations of
// structure and their order of checks (architecture 4.2): a meter without
// grouping has units but no groups and no beats.
func TestGroupingRequired(t *testing.T) {
	t.Parallel()

	ops := []struct {
		name string
		call func(m meter.Meter) error
	}{
		{"Grouping", func(m meter.Meter) error { return e(m.Grouping()) }},
		{"GroupCount", func(m meter.Meter) error { return e(m.GroupCount()) }},
		{"Groups", func(m meter.Meter) error { return e(m.Groups()) }},
		{"GroupsWritten", func(m meter.Meter) error { return e(m.GroupsWritten()) }},
		{"GroupStart 1", func(m meter.Meter) error { return e(m.GroupStart(1)) }},
		{"GroupStart 0", func(m meter.Meter) error { return e(m.GroupStart(0)) }},
		{"GroupAt 0", func(m meter.Meter) error { return e2(m.GroupAt(duration.Origin())) }},
		{"GroupAt end", func(m meter.Meter) error { return e2(m.GroupAt(endOf(m))) }},
		{"BeatCount", func(m meter.Meter) error { return e(m.BeatCount()) }},
		{"Beats", func(m meter.Meter) error { return e(m.Beats()) }},
		{"BeatsWritten", func(m meter.Meter) error { return e(m.BeatsWritten()) }},
		{"BeatStart 1", func(m meter.Meter) error { return e(m.BeatStart(1)) }},
		{"BeatStart 99", func(m meter.Meter) error { return e(m.BeatStart(99)) }},
		{"BeatAt 0", func(m meter.Meter) error { return e2(m.BeatAt(duration.Origin())) }},
		{"BeatAt end", func(m meter.Meter) error { return e2(m.BeatAt(endOf(m))) }},
		{"StrongBeats", func(m meter.Meter) error { return e2(m.StrongBeats()) }},
		{"UnitWeights", func(m meter.Meter) error { return e(m.UnitWeights()) }},
		{"WeightAt 0", func(m meter.Meter) error { return e(m.WeightAt(duration.Origin())) }},
		{"WeightAt end", func(m meter.Meter) error { return e(m.WeightAt(endOf(m))) }},
		{"ToAdditive", func(m meter.Meter) error { return e(m.ToAdditive()) }},
	}

	for _, text := range []string{"5/8", "7/8", "8/8", "10/8", "5/8+2/8", "11/8", "1024/8"} {
		for _, op := range ops {
			t.Run(text+" "+op.name, func(t *testing.T) {
				t.Parallel()

				assertErr(t, op.call(mt(text)), meter.ErrGroupingRequired)
			})
		}

		t.Run(text+" units are available", func(t *testing.T) {
			t.Parallel()

			m := mt(text)
			assert.False(t, noErr(m.HasGrouping()))
			assert.Equal(t, noErr(m.Numerator()), len(noErr(m.Units())))
		})
	}

	for _, text := range []string{"5/8", "7/8", "8/8"} {
		t.Run(text+" ToSum", func(t *testing.T) {
			t.Parallel()

			assertErr(t, e(mt(text).ToSum()), meter.ErrGroupingRequired)
		})
	}

	t.Run("5/8+2/4 Groups", func(t *testing.T) {
		t.Parallel()

		assertErr(t, e(mt("5/8+2/4").Groups()), meter.ErrGroupingRequired)
	})
}

// endOf returns the position at the end of the measure m.
func endOf(m meter.Meter) duration.Position {
	return posRat(ratOfDuration(m.MustDuration()))
}

// TestStructureSequence covers ИР-2, ИР-4, ИР-5 and architecture 4.7 on the
// corpus: Units, Groups, and Beats follow each other from 0 without gaps,
// the end of the last is the size, Grouping sums to UnitCount, and the
// counts are the lengths of the lists.
func TestStructureSequence(t *testing.T) {
	t.Parallel()

	for _, m := range corpus() {
		t.Run(m.String(), func(t *testing.T) {
			t.Parallel()

			size := m.MustDuration()

			checkSequence(t, "Units", noErr(m.Units()), size)
			assert.Len(t, noErr(m.Units()), noErr(m.UnitCount()))

			if !noErr(m.HasGrouping()) {
				return
			}

			groups := noErr(m.Groups())
			checkSequence(t, "Groups", groups, size)
			assert.Len(t, groups, noErr(m.GroupCount()))

			grouping := noErr(m.Grouping())
			assert.Len(t, grouping, len(groups))

			sum := 0
			for _, g := range grouping {
				sum += g
			}

			assert.Equal(t, noErr(m.UnitCount()), sum, "the grouping covers all units")

			beats := noErr(m.Beats())
			checkSequence(t, "Beats", beats, size)
			assert.Len(t, beats, noErr(m.BeatCount()))

			units := noErr(m.Units())
			assert.True(t, slices.Equal(beats, units) || slices.Equal(beats, groups), "ИР-5: beats are units or groups")
		})
	}
}

// checkSequence checks that the spans follow each other from 0 and end at
// size, and that Span.End and Span.Equal agree with the spans.
func checkSequence(t *testing.T, name string, spans []meter.Span, size duration.Duration) {
	t.Helper()

	require.NotEmpty(t, spans, name)

	next := duration.Origin()

	for i, s := range spans {
		assert.Equal(t, next, s.Start, "%s[%d].Start", name, i)
		assert.True(t, s.Duration.IsSet(), "%s[%d].Duration", name, i)
		assert.True(t, s.Equal(s))

		end, err := s.End()
		require.NoError(t, err, "%s[%d].End", name, i)

		next = end
	}

	assert.Equal(t, posRat(ratOfDuration(size)), next, "%s: the end of the last span is the size", name)
}

// TestUnitsOfLargeDenominator checks the largest denominator: the units of
// 1024/MaxInt follow each other within range.
func TestUnitsOfLargeDenominator(t *testing.T) {
	t.Parallel()

	m := meter.MustNew(1024, math.MaxInt)
	units := noErr(m.Units())
	require.Len(t, units, 1024)
	checkSequence(t, "Units", units, m.MustDuration())
}
