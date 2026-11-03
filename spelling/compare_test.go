package spelling_test

import (
	"slices"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// K-23: the list of the 35 ready-made names, in the contractual order.
func TestPredefined(t *testing.T) {
	t.Parallel()

	t.Run("content and order", func(t *testing.T) {
		t.Parallel()

		got := spelling.Predefined()
		want := predefinedCases()

		require.Len(t, got, 35)
		assert.Equal(t, 35, cap(got), "capacity")

		for i, tc := range want {
			assertName(t, tc.l, tc.a, got[i])
			assert.Truef(t, got[i].Equal(tc.fn()), "element %d must equal %s()", i, tc.name)
		}

		for i := range got {
			for j := range got {
				if i != j {
					assert.Falsef(t, got[i].Equal(got[j]), "elements %d and %d must differ", i, j)
				}
			}

			if i > 0 {
				assert.Equalf(t, -1, got[i-1].Compare(got[i]), "elements %d and %d must be in increasing order", i-1, i)
			}
		}
	})

	t.Run("the slice belongs to the caller", func(t *testing.T) {
		t.Parallel()

		first := spelling.Predefined()
		for i := range first {
			first[i] = spelling.MustNew(letter.G, whole(7))
		}

		first = append(first, spelling.C())
		assert.Len(t, first, 36)

		second := spelling.Predefined()
		require.Len(t, second, 35)

		for i, tc := range predefinedCases() {
			assertName(t, tc.l, tc.a, second[i])
			assertName(t, tc.l, tc.a, tc.fn())
		}
	})
}

// K-24: the contractual order for sorting: by letter, then by value. It is
// not the pitch order.
func TestCompare(t *testing.T) {
	t.Parallel()

	chain := []nameCase{
		nc(letter.C, frac(-7, 2)),
		nc(letter.C, alteration.DoubleFlat()),
		nc(letter.C, alteration.Flat()),
		nc(letter.C, frac(-1, 2)),
		nc(letter.C, alteration.Natural()),
		nc(letter.C, frac(1, 3)),
		nc(letter.C, frac(1, 2)),
		nc(letter.C, alteration.Sharp()),
		nc(letter.C, alteration.DoubleSharp()),
		nc(letter.C, whole(3)),
		nc(letter.D, alteration.DoubleFlat()),
		nc(letter.D, alteration.Flat()),
		nc(letter.D, alteration.Natural()),
		nc(letter.E, alteration.Sharp()),
		nc(letter.F, alteration.Flat()),
		nc(letter.F, alteration.Natural()),
		nc(letter.B, alteration.Flat()),
		nc(letter.B, alteration.Natural()),
		nc(letter.B, alteration.Sharp()),
	}

	t.Run("strictly increasing chain", func(t *testing.T) {
		t.Parallel()

		for i, x := range chain {
			for j, y := range chain {
				want := 0

				switch {
				case i < j:
					want = -1
				case i > j:
					want = 1
				}

				assert.Equalf(t, want, x.build().Compare(y.build()), "%s compared with %s", x.name, y.name)
			}
		}
	})

	t.Run("pairs", func(t *testing.T) {
		t.Parallel()

		pairs := []struct {
			name string
			x, y spelling.Spelling
			want int
		}{
			{"B sharp after C", spelling.BSharp(), spelling.C(), 1},
			{"C flat before B", spelling.CFlat(), spelling.B(), -1},
			{"C double sharp before D flat", spelling.CDoubleSharp(), spelling.DFlat(), -1},
			{"C+100 before D-100", spelling.MustNew(letter.C, whole(100)), spelling.MustNew(letter.D, whole(-100)), -1},
			{"ready-made C sharp and (C, +1)", spelling.CSharp(), spelling.MustNew(letter.C, alteration.Sharp()), 0},
			{"ready-made C sharp and C raised", spelling.CSharp(), spelling.C().MustRaise(), 0},
			{"(C, +1) and C raised", spelling.MustNew(letter.C, alteration.Sharp()), spelling.C().MustRaise(), 0},
		}

		for _, tc := range pairs {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tc.want, tc.x.Compare(tc.y))
				assert.Equal(t, -tc.want, tc.y.Compare(tc.x))
			})
		}
	})

	// Two different values near the range limit whose difference,
	// 1/65535 - 1/65536 (about 2.3e-10), is less than half the float64 spacing
	// around 2^24: an inexact comparison through float64 would find them equal.
	t.Run("close fractions near the range limit", func(t *testing.T) {
		t.Parallel()

		// 2^24 - 1/65535 and 2^24 - 1/65536, both in lowest terms and in range.
		lowValue := frac(1099494850559, 65535)
		highValue := frac(1099511627775, 65536)

		pairs := []struct {
			name   string
			before alteration.Alteration
			after  alteration.Alteration
		}{
			{"positive", lowValue, highValue},
			{"negative", alteration.Natural().MustSub(highValue), alteration.Natural().MustSub(lowValue)},
		}

		for _, tc := range pairs {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				require.Equal(t, -1, tc.before.Compare(tc.after), "the alteration oracle")

				x := spelling.MustNew(letter.C, tc.before)
				y := spelling.MustNew(letter.C, tc.after)

				assert.Equal(t, -1, x.Compare(y))
				assert.Equal(t, 1, y.Compare(x))
				assert.False(t, x.Equal(y))
				assert.False(t, y.Equal(x))
				assert.True(t, x != y, "the names must differ by ==")
				assert.Equal(t, 0, x.Compare(x))
			})
		}

		t.Run("machine form", func(t *testing.T) {
			t.Parallel()

			x := spelling.MustParse("C+1099494850559/65535")
			y := spelling.MustParse("C+1099511627775/65536")

			assert.Equal(t, -1, x.Compare(y))
			assert.Equal(t, 1, y.Compare(x))
			assert.False(t, x.Equal(y))
			assert.False(t, y.Equal(x))
			assert.True(t, x != y, "the names must differ by ==")
		})
	})

	t.Run("properties on N", func(t *testing.T) {
		t.Parallel()

		names := setN()
		for _, x := range names {
			for _, y := range names {
				got := x.build().Compare(y.build())

				assert.Containsf(t, []int{-1, 0, 1}, got, "%s compared with %s", x.name, y.name)
				assert.Equalf(t, compareOracle(x, y), got, "%s compared with %s", x.name, y.name)
				assert.Equalf(t, -got, y.build().Compare(x.build()), "antisymmetry for %s and %s", x.name, y.name)
				assert.Equalf(t, got == 0, x.build().Equal(y.build()), "Compare == 0 and Equal for %s and %s", x.name, y.name)
			}
		}
	})

	t.Run("transitivity on N", func(t *testing.T) {
		t.Parallel()

		names := setN()
		for _, x := range names {
			for _, y := range names {
				for _, z := range names {
					xs, ys, zs := x.build(), y.build(), z.build()
					if xs.Compare(ys) <= 0 && ys.Compare(zs) <= 0 {
						assert.LessOrEqualf(t, xs.Compare(zs), 0, "%s <= %s <= %s", x.name, y.name, z.name)
					}

					if xs.Compare(ys) < 0 && ys.Compare(zs) < 0 {
						assert.Equalf(t, -1, xs.Compare(zs), "%s < %s < %s", x.name, y.name, z.name)
					}
				}
			}
		}
	})

	t.Run("sorting", func(t *testing.T) {
		t.Parallel()

		want := make([]spelling.Spelling, 0, len(chain))
		for _, c := range chain {
			want = append(want, c.build())
		}

		got := make([]spelling.Spelling, 0, len(chain))
		for i := len(chain) - 1; i >= 0; i -= 2 {
			got = append(got, chain[i].build())
		}

		for i := len(chain) - 2; i >= 0; i -= 2 {
			got = append(got, chain[i].build())
		}

		slices.SortFunc(got, spelling.Spelling.Compare)
		assert.Equal(t, want, got)
	})
}

// K-25: letter distances of two names depend only on their letters.
func TestLetterDistance(t *testing.T) {
	t.Parallel()

	t.Run("examples", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name     string
			from, to spelling.Spelling
			up, down int
		}{
			{"C sharp to E flat", spelling.CSharp(), spelling.EFlat(), 2, 5},
			{"B sharp to C flat", spelling.BSharp(), spelling.CFlat(), 1, 6},
			{"C to C sharp", spelling.C(), spelling.CSharp(), 0, 0},
			{"A+1/2 to A+3", spelling.MustNew(letter.A, frac(1, 2)), spelling.MustNew(letter.A, whole(3)), 0, 0},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tc.up, tc.from.LetterDistanceUp(tc.to), "up")
				assert.Equal(t, tc.down, tc.from.LetterDistanceDown(tc.to), "down")
			})
		}
	})

	t.Run("properties on N", func(t *testing.T) {
		t.Parallel()

		for _, x := range setN() {
			for _, y := range setN() {
				xs, ys := x.build(), y.build()
				up := xs.LetterDistanceUp(ys)
				down := xs.LetterDistanceDown(ys)

				assert.Equalf(t, x.l.DistanceUp(y.l), up, "up from %s to %s", x.name, y.name)
				assert.Equalf(t, x.l.DistanceDown(y.l), down, "down from %s to %s", x.name, y.name)
				assert.Equalf(t, up, ys.LetterDistanceDown(xs), "up(x, y) = down(y, x) for %s and %s", x.name, y.name)
				assert.GreaterOrEqual(t, up, 0)
				assert.LessOrEqual(t, up, 6)
				assert.GreaterOrEqual(t, down, 0)
				assert.LessOrEqual(t, down, 6)

				if x.l != y.l {
					assert.Equalf(t, 7, up+down, "up + down for %s and %s", x.name, y.name)
				}

				natural := xs.Natural()
				assert.Equalf(t, up, natural.LetterDistanceUp(ys.Natural()), "values do not matter for %s and %s", x.name, y.name)
			}
		}
	})
}

// K-26: the difference of the values of two names; the letters do not matter.
func TestAlterationDifference(t *testing.T) {
	t.Parallel()

	t.Run("examples", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name string
			x, y spelling.Spelling
			want alteration.Alteration
		}{
			{"C sharp minus C flat", spelling.CSharp(), spelling.CFlat(), alteration.DoubleSharp()},
			{"D flat minus C sharp", spelling.DFlat(), spelling.CSharp(), alteration.DoubleFlat()},
			{"A+1/2 minus A", spelling.MustNew(letter.A, frac(1, 2)), spelling.A(), frac(1, 2)},
			{"B sharp minus C", spelling.BSharp(), spelling.C(), alteration.Sharp()},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				got, err := tc.x.AlterationDifference(tc.y)
				require.NoError(t, err)
				assertAlteration(t, tc.want, got)
			})
		}
	})

	t.Run("properties on N", func(t *testing.T) {
		t.Parallel()

		for _, x := range setN() {
			for _, y := range setN() {
				xs, ys := x.build(), y.build()

				dxy, err := xs.AlterationDifference(ys)
				require.NoError(t, err)
				assertAlteration(t, x.a.MustSub(y.a), dxy)

				dyx, err := ys.AlterationDifference(xs)
				require.NoError(t, err)
				assertAlteration(t, alteration.Natural().MustSub(dyx), dxy)

				shifted, err := ys.RaiseByAlteration(dxy)
				require.NoError(t, err)
				assertName(t, y.l, x.a, shifted)
			}

			self, err := x.build().AlterationDifference(x.build())
			require.NoError(t, err)
			assertAlteration(t, alteration.Natural(), self)
		}
	})
}
