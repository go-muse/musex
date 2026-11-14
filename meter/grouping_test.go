package meter_test

import (
	"math"
	"slices"
	"testing"

	"github.com/go-muse/musex/meter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWithGrouping covers О-7, П-6, Т-4, КП-29, КП-42, КП-43, КП-44 and the
// order of checks of WithGrouping.
func TestWithGrouping(t *testing.T) {
	t.Parallel()
	defer guard(t)

	for _, c := range []struct {
		name, text string
		groups     []int
		want       string // the expected value by mt
		explicit   bool
	}{
		{"5/8 as 2+3", "5/8", []int{2, 3}, "5/8[2+3]", true},
		{"5/8 as 3+2", "5/8", []int{3, 2}, "5/8[3+2]", true},
		{"КП-43 one group 5/8 as 5", "5/8", []int{5}, "5/8[5]", true},
		{"nonstandard 6/8 as 2+2+2", "6/8", []int{2, 2, 2}, "6/8[2+2+2]", true},
		{"КП-42 standard 6/8 as 3+3", "6/8", []int{3, 3}, "6/8", false},
		{"КП-42 standard 4/4 as 2+2", "4/4", []int{2, 2}, "4/4", false},
		{"standard 3/4 as 3", "3/4", []int{3}, "3/4", false},
		{"КП-44 C as 3+1", "C", []int{3, 1}, "C[3+1]", true},
		{"standard C as 2+2", "C", []int{2, 2}, "C", false},
		{"C| as 1+1", "C|", []int{1, 1}, "C|[1+1]", true},
		{"replaces the explicit grouping", "5/8[2+3]", []int{3, 2}, "5/8[3+2]", true},
		{"back to standard", "6/8[2+2+2]", []int{3, 3}, "6/8", false},
		{"7/8 as 3+2+2", "7/8", []int{3, 2, 2}, "7/8[3+2+2]", true},
		{"1/4 as 1", "1/4", []int{1}, "1/4", false},
		{"1024/8 in groups of 512", "1024/8", []int{512, 512}, "1024/8[512+512]", true},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			m := mt(c.text)

			got, err := m.WithGrouping(c.groups...)
			require.NoError(t, err)
			assert.Equal(t, c.groups, noErr(got.Grouping()))
			assert.Equal(t, c.explicit, noErr(got.HasExplicitGrouping()))
			assert.Equal(t, noErr(m.Form()), noErr(got.Form()), "the record of the signature is kept")
			assert.Equal(t, noErr(m.Parts()), noErr(got.Parts()))

			if !c.explicit {
				assert.Equal(t, mt(baseOf(c.text)), got, "a standard grouping gives the value without explicit grouping")
			}

			want := mt(c.want)
			assert.Equal(t, want, got)
			assert.True(t, got.Equal(want))
		})
	}

	var unset meter.Meter

	for _, c := range []struct {
		name   string
		m      meter.Meter
		groups []int
		want   error
	}{
		{"КП-29 5/8 as 2+2", mt("5/8"), []int{2, 2}, meter.ErrInvalidArgument},
		{"КП-29 5/8 as 2+3+1", mt("5/8"), []int{2, 3, 1}, meter.ErrInvalidArgument},
		{"КП-29 group 0", mt("5/8"), []int{0, 5}, meter.ErrInvalidArgument},
		{"negative group", mt("5/8"), []int{6, -1}, meter.ErrInvalidArgument},
		{"КП-29 empty", mt("5/8"), nil, meter.ErrInvalidArgument},
		{"КП-29 compound numerator", mt("(2+3)/8"), []int{2, 3}, meter.ErrInvalidArgument},
		{"КП-29 sum", mt("2/4+3/8"), []int{2, 3}, meter.ErrInvalidArgument},
		{"the sum of groups wraps to the numerator", mt("5/8"), []int{math.MaxInt, math.MaxInt, 7}, meter.ErrInvalidArgument},
		{"the sum of groups overflows", mt("5/8"), []int{math.MaxInt, 1}, meter.ErrInvalidArgument},
		{"unset", unset, []int{2, 3}, meter.ErrUnset},
		{"order: E-1 before E-2", unset, nil, meter.ErrUnset},
		{"order: form before an empty list", mt("(2+3)/8"), nil, meter.ErrInvalidArgument},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			got, err := c.m.WithGrouping(c.groups...)
			assertErr(t, err, c.want)
			assert.Equal(t, meter.Meter{}, got)
		})
	}
}

// baseOf returns the signature text without its explicit grouping.
func baseOf(text string) string {
	if i := slices.Index([]byte(text), '['); i >= 0 {
		return text[:i]
	}

	return text
}

// TestResetGrouping covers О-7, КП-42.
func TestResetGrouping(t *testing.T) {
	t.Parallel()
	defer guard(t)

	for _, c := range []struct{ text, want string }{
		{"5/8[2+3]", "5/8"},
		{"6/8[2+2+2]", "6/8"},
		{"C[3+1]", "C"},
		{"C|[1+1]", "C|"},
		{"7/8[7]", "7/8"},
		{"6/8", "6/8"},
		{"5/8", "5/8"},
		{"(2+3)/8", "(2+3)/8"},
		{"2/4+3/8", "2/4+3/8"},
		{"C", "C"},
	} {
		t.Run(c.text, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			got, err := mt(c.text).ResetGrouping()
			require.NoError(t, err)
			assert.Equal(t, mt(c.want), got)
			assert.False(t, noErr(got.HasExplicitGrouping()))
		})
	}

	t.Run("unset", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		got, err := meter.Meter{}.ResetGrouping()
		assertErr(t, err, meter.ErrUnset)
		assert.Equal(t, meter.Meter{}, got)
	})
}

// TestGroupingVariants covers О-17, П-9, Т-4, КП-23, КП-35 and the order of
// architecture 4.9 (lexicographic, 2 before 3).
func TestGroupingVariants(t *testing.T) {
	t.Parallel()
	defer guard(t)

	for _, c := range []struct {
		units int
		want  [][]int
	}{
		{5, [][]int{{2, 3}, {3, 2}}},
		{7, [][]int{{2, 2, 3}, {2, 3, 2}, {3, 2, 2}}},
		{8, [][]int{{2, 2, 2, 2}, {2, 3, 3}, {3, 2, 3}, {3, 3, 2}}},
		{6, [][]int{{2, 2, 2}, {3, 3}}},
		{4, [][]int{{2, 2}}},
		{2, [][]int{{2}}},
		{3, [][]int{{3}}},
		{1, nil},
	} {
		t.Run("КП-23 "+itoa(c.units), func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			got, err := meter.GroupingVariants(c.units)
			require.NoError(t, err)
			assert.Equal(t, c.want, got)
		})
	}

	t.Run("against an independent enumeration", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		for n := 2; n <= 27; n++ {
			got, err := meter.GroupingVariants(n)
			require.NoError(t, err, n)
			assert.Equal(t, compositions(n), got, "units %d", n)
		}
	})

	t.Run("КП-23 27 gives 816 variants", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		got, err := meter.GroupingVariants(27)
		require.NoError(t, err)
		assert.Len(t, got, 816)
	})

	t.Run("max is MaxGroupingVariants", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		assert.Equal(t, 1024, meter.MaxGroupingVariants)
		assert.Equal(t, 1024, meter.MaxUnits)
	})

	for _, c := range []struct {
		name  string
		units int
		want  error
	}{
		{"КП-35 28 gives 1081 variants", 28, meter.ErrTooMany},
		{"29", 29, meter.ErrTooMany},
		{"1024", 1024, meter.ErrTooMany},
		{"MaxInt without enumeration", math.MaxInt, meter.ErrTooMany},
		{"КП-35 0", 0, meter.ErrInvalidArgument},
		{"КП-35 -1", -1, meter.ErrInvalidArgument},
		{"MinInt", math.MinInt, meter.ErrInvalidArgument},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			r := fast(t, func() result[[][]int] {
				v, err := meter.GroupingVariants(c.units)

				return result[[][]int]{v, err}
			})
			assertErr(t, r.err, c.want)
			assert.Nil(t, r.v)
		})
	}
}

// TestMeterGroupingVariants covers О-17 for a meter: the variants of its
// number of units, independent of its grouping.
func TestMeterGroupingVariants(t *testing.T) {
	t.Parallel()
	defer guard(t)

	for _, c := range []struct {
		text  string
		units int
	}{
		{"5/8", 5},
		{"5/8[2+3]", 5},
		{"C", 4},
		{"C|", 2},
		{"2/4+3/8", 5},
		{"(3+2+3)/8", 8},
		{"1/4", 1},
		{"27/8", 27},
	} {
		t.Run(c.text, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			got, err := mt(c.text).GroupingVariants()
			require.NoError(t, err)
			assert.Equal(t, noErr(meter.GroupingVariants(c.units)), got)
		})
	}

	for _, c := range []struct {
		name string
		m    meter.Meter
		want error
	}{
		{"28/8", mt("28/8"), meter.ErrTooMany},
		{"1024/8", mt("1024/8"), meter.ErrTooMany},
		{"unset", meter.Meter{}, meter.ErrUnset},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			got, err := c.m.GroupingVariants()
			assertErr(t, err, c.want)
			assert.Nil(t, got)
		})
	}
}
