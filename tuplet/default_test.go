package tuplet_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDefault covers KP-7, KP-8 and rule P-5 for large counts.
func TestDefault(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		actual   int
		division tuplet.Division
		normal   int
	}{
		{name: "KP-7 binary 3", actual: 3, division: tuplet.Binary, normal: 2},
		{name: "KP-7 binary 5", actual: 5, division: tuplet.Binary, normal: 4},
		{name: "KP-7 binary 6", actual: 6, division: tuplet.Binary, normal: 4},
		{name: "KP-7 binary 7", actual: 7, division: tuplet.Binary, normal: 4},
		{name: "KP-7 binary 9", actual: 9, division: tuplet.Binary, normal: 8},
		{name: "KP-7 binary 14", actual: 14, division: tuplet.Binary, normal: 8},
		{name: "KP-8 ternary 2", actual: 2, division: tuplet.Ternary, normal: 3},
		{name: "KP-8 ternary 4", actual: 4, division: tuplet.Ternary, normal: 3},
		{name: "KP-8 ternary 5", actual: 5, division: tuplet.Ternary, normal: 3},
		{name: "KP-8 ternary 7", actual: 7, division: tuplet.Ternary, normal: 6},
		{name: "binary 10", actual: 10, division: tuplet.Binary, normal: 8},
		{name: "ternary 8", actual: 8, division: tuplet.Ternary, normal: 6},
		{name: "ternary 13", actual: 13, division: tuplet.Ternary, normal: 12},
		{name: "binary 17", actual: 17, division: tuplet.Binary, normal: 16},
		{name: "binary max int", actual: math.MaxInt, division: tuplet.Binary, normal: 1 << 62},
		{name: "ternary max int", actual: math.MaxInt, division: tuplet.Ternary, normal: 3 << 61},
		{name: "ternary 1<<62", actual: 1 << 62, division: tuplet.Ternary, normal: 3 << 60},
		{name: "binary 3<<61", actual: 3 << 61, division: tuplet.Binary, normal: 1 << 62},
		{name: "binary 1<<62 + 1", actual: 1<<62 + 1, division: tuplet.Binary, normal: 1 << 62},
		{name: "ternary 3<<61 + 1", actual: 3<<61 + 1, division: tuplet.Ternary, normal: 3 << 61},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tuplet.Default(tt.actual, tt.division)
			require.NoError(t, err)
			assertTuplet(t, got, tt.actual, tt.normal)
			assert.Equal(t, got, tuplet.MustDefault(tt.actual, tt.division))
		})
	}
}

// TestDefault_Oracle covers K-3: Default agrees with rule P-5 computed by
// walking the series, for every small count and for counts near the series
// members up to math.MaxInt.
func TestDefault_Oracle(t *testing.T) {
	t.Parallel()

	actuals := make([]int, 0, 600)
	for n := 1; n <= 300; n++ {
		actuals = append(actuals, n)
	}

	for shift := uint(1); shift < 63; shift++ {
		for _, base := range []int{1, 3} {
			v := base << shift
			if v <= 0 || v/base != 1<<shift {
				continue
			}

			actuals = append(actuals, v-1, v, v+1)
		}
	}

	actuals = append(actuals, math.MaxInt-1, math.MaxInt)

	for _, division := range []tuplet.Division{tuplet.Binary, tuplet.Ternary} {
		t.Run(division.String(), func(t *testing.T) {
			t.Parallel()

			for _, actual := range actuals {
				normal, ok := oracleDefault(actual, division)
				got, err := tuplet.Default(actual, division)

				if !ok {
					assertErr(t, err, tuplet.ErrNoDefault)
					assert.Equal(t, tuplet.Tuplet{}, got, "actual %d", actual)

					continue
				}

				if assert.NoError(t, err, "actual %d", actual) {
					assertTuplet(t, got, actual, normal)
				}
			}
		})
	}
}

// TestDefault_Errors covers KP-9, E-1, E-3 and the order E-1 -> E-3.
func TestDefault_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		actual   int
		division tuplet.Division
		want     error
	}{
		{name: "KP-9 binary 1", actual: 1, division: tuplet.Binary, want: tuplet.ErrNoDefault},
		{name: "KP-9 binary 2", actual: 2, division: tuplet.Binary, want: tuplet.ErrNoDefault},
		{name: "KP-9 binary 4", actual: 4, division: tuplet.Binary, want: tuplet.ErrNoDefault},
		{name: "KP-9 binary 8", actual: 8, division: tuplet.Binary, want: tuplet.ErrNoDefault},
		{name: "KP-9 ternary 1", actual: 1, division: tuplet.Ternary, want: tuplet.ErrNoDefault},
		{name: "KP-9 ternary 3", actual: 3, division: tuplet.Ternary, want: tuplet.ErrNoDefault},
		{name: "KP-9 ternary 6", actual: 6, division: tuplet.Ternary, want: tuplet.ErrNoDefault},
		{name: "KP-9 ternary 12", actual: 12, division: tuplet.Ternary, want: tuplet.ErrNoDefault},
		{name: "binary 1<<62", actual: 1 << 62, division: tuplet.Binary, want: tuplet.ErrNoDefault},
		{name: "ternary 3<<61", actual: 3 << 61, division: tuplet.Ternary, want: tuplet.ErrNoDefault},
		{name: "zero actual", actual: 0, division: tuplet.Binary, want: tuplet.ErrInvalidCount},
		{name: "negative actual", actual: -1, division: tuplet.Ternary, want: tuplet.ErrInvalidCount},
		{name: "min int actual", actual: math.MinInt, division: tuplet.Binary, want: tuplet.ErrInvalidCount},
		{name: "division zero", actual: 5, division: tuplet.Division(0), want: tuplet.ErrNoDefault},
		{name: "division three", actual: 5, division: tuplet.Division(3), want: tuplet.ErrNoDefault},
		{name: "division 255", actual: 5, division: tuplet.Division(255), want: tuplet.ErrNoDefault},
		{name: "count before division", actual: 0, division: tuplet.Division(0), want: tuplet.ErrInvalidCount},
		{name: "unknown division with actual 1", actual: 1, division: tuplet.Division(9), want: tuplet.ErrNoDefault},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tuplet.Default(tt.actual, tt.division)
			assertErr(t, err, tt.want)
			assert.Equal(t, tuplet.Tuplet{}, got)
		})
	}
}

// TestDivisionOf covers KP-11 and architecture 4.3.
func TestDivisionOf(t *testing.T) {
	t.Parallel()

	tests := []struct {
		count    int
		division tuplet.Division
		ok       bool
	}{
		{count: 1},
		{count: 2, division: tuplet.Binary, ok: true},
		{count: 3, division: tuplet.Ternary, ok: true},
		{count: 4, division: tuplet.Binary, ok: true},
		{count: 6, division: tuplet.Ternary, ok: true},
		{count: 7},
		{count: 12, division: tuplet.Ternary, ok: true},
		{count: 16, division: tuplet.Binary, ok: true},
		{count: 0},
		{count: -1},
		{count: -2},
		{count: -3},
		{count: math.MinInt},
		{count: 5},
		{count: 9},
		{count: 24, division: tuplet.Ternary, ok: true},
		{count: 1 << 61, division: tuplet.Binary, ok: true},
		{count: 1 << 62, division: tuplet.Binary, ok: true},
		{count: 3 << 60, division: tuplet.Ternary, ok: true},
		{count: 3 << 61, division: tuplet.Ternary, ok: true},
		{count: math.MaxInt},
		{count: 1<<62 + 1},
	}

	for _, tt := range tests {
		t.Run(itoa(tt.count), func(t *testing.T) {
			t.Parallel()

			division, ok := tuplet.DivisionOf(tt.count)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.division, division)
		})
	}
}

// TestDivisionOf_Oracle checks DivisionOf against the series for small counts.
func TestDivisionOf_Oracle(t *testing.T) {
	t.Parallel()

	for count := -10; count <= 1000; count++ {
		want, wantOK := oracleDivision(count)
		got, ok := tuplet.DivisionOf(count)

		assert.Equal(t, wantOK, ok, "count %d", count)
		assert.Equal(t, want, got, "count %d", count)
	}
}

// TestNormalDivision covers O-7 and KP-11 through the normal count.
func TestNormalDivision(t *testing.T) {
	t.Parallel()

	for _, normal := range []int{1, 2, 3, 4, 5, 6, 7, 12, 16, 1 << 62, 3 << 61, math.MaxInt} {
		for _, actual := range []int{1, 3, 7, math.MaxInt} {
			t.Run(itoa(actual)+":"+itoa(normal), func(t *testing.T) {
				t.Parallel()

				wantDivision, wantOK := oracleDivision(normal)

				division, ok, err := tuplet.MustNew(actual, normal).NormalDivision()
				require.NoError(t, err)
				assert.Equal(t, wantOK, ok)
				assert.Equal(t, wantDivision, division)
			})
		}
	}
}

// TestHasDefaultNormal covers K-5 and architecture 4.5.
func TestHasDefaultNormal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		actual, normal  int
		binary, ternary bool
	}{
		{actual: 3, normal: 2, binary: true},
		{actual: 5, normal: 3, ternary: true},
		{actual: 5, normal: 2},
		{actual: 6, normal: 4, binary: true},
		{actual: 7, normal: 8},
		{actual: 2, normal: 3, ternary: true},
		{actual: 4, normal: 3, ternary: true},
		{actual: 7, normal: 4, binary: true},
		{actual: 7, normal: 6, ternary: true},
		{actual: 1, normal: 1},
		{actual: 2, normal: 2},
		{actual: 4, normal: 2},
		{actual: 3, normal: 3},
		{actual: math.MaxInt, normal: 1 << 62, binary: true},
		{actual: math.MaxInt, normal: 3 << 61, ternary: true},
		{actual: math.MaxInt, normal: math.MaxInt},
	}

	for _, tt := range tests {
		t.Run(itoa(tt.actual)+":"+itoa(tt.normal), func(t *testing.T) {
			t.Parallel()

			tp := tuplet.MustNew(tt.actual, tt.normal)

			binary, err := tp.HasDefaultNormalFor(tuplet.Binary)
			require.NoError(t, err)
			assert.Equal(t, tt.binary, binary, "binary")

			ternary, err := tp.HasDefaultNormalFor(tuplet.Ternary)
			require.NoError(t, err)
			assert.Equal(t, tt.ternary, ternary, "ternary")

			either, err := tp.HasDefaultNormal()
			require.NoError(t, err)
			assert.Equal(t, tt.binary || tt.ternary, either, "any division")
		})
	}
}

// TestHasDefaultNormal_Oracle checks O-8 against rule P-5 on a grid of pairs.
func TestHasDefaultNormal_Oracle(t *testing.T) {
	t.Parallel()

	for actual := 1; actual <= 40; actual++ {
		for normal := 1; normal <= 40; normal++ {
			tp := tuplet.MustNew(actual, normal)
			anyDivision := false

			for _, division := range []tuplet.Division{tuplet.Binary, tuplet.Ternary} {
				want, ok := oracleDefault(actual, division)
				expected := ok && want == normal
				anyDivision = anyDivision || expected

				got, err := tp.HasDefaultNormalFor(division)
				if assert.NoError(t, err) {
					assert.Equal(t, expected, got, "%d:%d %s", actual, normal, division)
				}
			}

			got, err := tp.HasDefaultNormal()
			if assert.NoError(t, err) {
				assert.Equal(t, anyDivision, got, "%d:%d", actual, normal)
			}
		}
	}
}

// TestHasDefaultNormalFor_Errors covers E-3 for an unknown division and the
// order E-2 -> E-3.
func TestHasDefaultNormalFor_Errors(t *testing.T) {
	t.Parallel()

	for _, division := range []tuplet.Division{0, 3, 255} {
		t.Run(itoa(int(division)), func(t *testing.T) {
			t.Parallel()

			got, err := tuplet.Triplet().HasDefaultNormalFor(division)
			assertErr(t, err, tuplet.ErrNoDefault)
			assert.False(t, got)

			got, err = tuplet.Tuplet{}.HasDefaultNormalFor(division)
			assertErr(t, err, tuplet.ErrUnset)
			assert.False(t, got)
		})
	}
}

// TestDivisionAndKindEnums covers architecture 3.1 and 4.3: values, IsValid,
// and String of the enumerations.
func TestDivisionAndKindEnums(t *testing.T) {
	t.Parallel()

	t.Run("values", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, uint8(1), uint8(tuplet.Binary))
		assert.Equal(t, uint8(2), uint8(tuplet.Ternary))
		assert.Equal(t, uint8(1), uint8(tuplet.Compressing))
		assert.Equal(t, uint8(2), uint8(tuplet.Expanding))
		assert.Equal(t, uint8(3), uint8(tuplet.Neutral))
	})

	divisions := []struct {
		d     tuplet.Division
		valid bool
		text  string
	}{
		{d: 0, text: "Division(invalid)"},
		{d: tuplet.Binary, valid: true, text: "binary"},
		{d: tuplet.Ternary, valid: true, text: "ternary"},
		{d: 3, text: "Division(invalid)"},
		{d: 255, text: "Division(invalid)"},
	}

	for _, tt := range divisions {
		t.Run("Division/"+itoa(int(tt.d)), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.valid, tt.d.IsValid())
			assert.Equal(t, tt.text, tt.d.String())
		})
	}

	kinds := []struct {
		k     tuplet.Kind
		valid bool
		text  string
	}{
		{k: 0, text: "Kind(invalid)"},
		{k: tuplet.Compressing, valid: true, text: "compressing"},
		{k: tuplet.Expanding, valid: true, text: "expanding"},
		{k: tuplet.Neutral, valid: true, text: "neutral"},
		{k: 4, text: "Kind(invalid)"},
		{k: 255, text: "Kind(invalid)"},
	}

	for _, tt := range kinds {
		t.Run("Kind/"+itoa(int(tt.k)), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.valid, tt.k.IsValid())
			assert.Equal(t, tt.text, tt.k.String())
		})
	}
}
