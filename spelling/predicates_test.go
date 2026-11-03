package spelling_test

import (
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// namedSpelling is a spelling with a case name.
type namedSpelling struct {
	name string
	s    spelling.Spelling
}

// K-31: whether a name is natural.
func TestIsNatural(t *testing.T) {
	t.Parallel()

	t.Run("examples", func(t *testing.T) {
		t.Parallel()

		yes := []namedSpelling{
			{"C", spelling.C()},
			{"B", spelling.B()},
			{"created (C, natural)", spelling.MustNew(letter.C, alteration.Natural())},
		}
		for _, p := range predefinedCases() {
			if p.a.IsNatural() {
				yes = append(yes, namedSpelling{p.name, p.fn()})
			}
		}

		require.Len(t, yes, 3+7)

		for _, tc := range yes {
			assert.Truef(t, tc.s.IsNatural(), "%s must be natural", tc.name)
		}

		no := []namedSpelling{
			{"C sharp", spelling.CSharp()},
			{"D flat", spelling.DFlat()},
			{"A+1/2", spelling.MustNew(letter.A, frac(1, 2))},
			{"E+3", spelling.MustNew(letter.E, whole(3))},
		}
		for _, tc := range no {
			assert.Falsef(t, tc.s.IsNatural(), "%s must not be natural", tc.name)
		}
	})

	for _, tc := range setN() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			x := tc.build()
			assert.Equal(t, tc.a.IsNatural(), x.IsNatural())
			assert.True(t, x.Natural().IsNatural(), "the natural name is natural")
		})
	}
}

// K-32: whether two names have the same letter.
func TestSameLetter(t *testing.T) {
	t.Parallel()

	t.Run("examples", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name string
			x, y spelling.Spelling
			want bool
		}{
			{"C sharp and C flat", spelling.CSharp(), spelling.CFlat(), true},
			{"C and C+1/2", spelling.C(), spelling.MustNew(letter.C, frac(1, 2)), true},
			{"B sharp and B", spelling.BSharp(), spelling.B(), true},
			{"C sharp and D flat", spelling.CSharp(), spelling.DFlat(), false},
			{"B sharp and C", spelling.BSharp(), spelling.C(), false},
			{"E sharp and F", spelling.ESharp(), spelling.F(), false},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tc.want, tc.x.SameLetter(tc.y))
				assert.Equal(t, tc.want, tc.y.SameLetter(tc.x))
			})
		}
	})

	t.Run("properties on N", func(t *testing.T) {
		t.Parallel()

		for _, x := range setN() {
			for _, y := range setN() {
				xs, ys := x.build(), y.build()
				got := xs.SameLetter(ys)

				assert.Equalf(t, x.l == y.l, got, "%s and %s", x.name, y.name)
				assert.Equalf(t, xs.LetterDistanceUp(ys) == 0, got, "distance up for %s and %s", x.name, y.name)
				assert.Equalf(t, got, ys.SameLetter(xs), "symmetry for %s and %s", x.name, y.name)
			}

			assert.Truef(t, x.build().SameLetter(x.build()), "%s with itself", x.name)
		}
	})
}

// K-33: whether the value of a name is whole.
func TestIsInteger(t *testing.T) {
	t.Parallel()

	t.Run("examples", func(t *testing.T) {
		t.Parallel()

		yes := []namedSpelling{
			{"C", spelling.C()},
			{"C sharp", spelling.CSharp()},
			{"C double sharp", spelling.CDoubleSharp()},
			{"E+3", spelling.MustNew(letter.E, whole(3))},
			{"B sharp", spelling.BSharp()},
		}
		for _, p := range predefinedCases() {
			yes = append(yes, namedSpelling{p.name, p.fn()})
		}

		for _, tc := range yes {
			assert.Truef(t, tc.s.IsInteger(), "%s must be whole", tc.name)
		}

		no := []namedSpelling{
			{"A+1/2", spelling.MustNew(letter.A, frac(1, 2))},
			{"D-3/2", spelling.MustNew(letter.D, frac(-3, 2))},
			{"B-7/2", spelling.MustNew(letter.B, frac(-7, 2))},
			{"C+1/3", spelling.MustNew(letter.C, frac(1, 3))},
		}
		for _, tc := range no {
			assert.Falsef(t, tc.s.IsInteger(), "%s must not be whole", tc.name)
		}
	})

	for _, tc := range setN() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.a.IsInteger(), tc.build().IsInteger())
		})
	}
}

// K-34: the direction of the value of a name.
func TestDirection(t *testing.T) {
	t.Parallel()

	t.Run("examples", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			want  spelling.Direction
			names []namedSpelling
		}{
			{spelling.DirectionRaised, []namedSpelling{
				{"C sharp", spelling.CSharp()},
				{"C double sharp", spelling.CDoubleSharp()},
				{"A+1/2", spelling.MustNew(letter.A, frac(1, 2))},
				{"C+1/3", spelling.MustNew(letter.C, frac(1, 3))},
				{"E+3", spelling.MustNew(letter.E, whole(3))},
				{"B sharp", spelling.BSharp()},
			}},
			{spelling.DirectionLowered, []namedSpelling{
				{"D flat", spelling.DFlat()},
				{"C flat", spelling.CFlat()},
				{"F double flat", spelling.FDoubleFlat()},
				{"D-3/2", spelling.MustNew(letter.D, frac(-3, 2))},
				{"B-7/2", spelling.MustNew(letter.B, frac(-7, 2))},
			}},
			{spelling.DirectionNatural, []namedSpelling{
				{"C", spelling.C()},
				{"F", spelling.F()},
				{"D", spelling.D()},
			}},
		}

		for _, group := range cases {
			for _, tc := range group.names {
				assert.Equalf(t, group.want, tc.s.Direction(), "direction of %s", tc.name)
			}
		}
	})

	for _, tc := range setN() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			x := tc.build()
			got := x.Direction()

			assert.Equal(t, directionOf(tc.a), got)
			assert.Equal(t, x.IsNatural(), got == spelling.DirectionNatural, "natural direction and IsNatural")
			assert.NotEqual(t, spelling.DirectionUnset, got, "a set name has a direction")
		})
	}
}

// K-35: whether two names have the same value.
func TestSameAlteration(t *testing.T) {
	t.Parallel()

	t.Run("examples", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name string
			x, y spelling.Spelling
			want bool
		}{
			{"C sharp and D sharp", spelling.CSharp(), spelling.DSharp(), true},
			{"C and F", spelling.C(), spelling.F(), true},
			{"A+1/2 and C+1/2", spelling.MustNew(letter.A, frac(1, 2)), spelling.MustNew(letter.C, frac(1, 2)), true},
			{"C+2/4 and D+1/2", spelling.MustNew(letter.C, frac(2, 4)), spelling.MustNew(letter.D, frac(1, 2)), true},
			{"C sharp and C flat", spelling.CSharp(), spelling.CFlat(), false},
			{"C sharp and C double sharp", spelling.CSharp(), spelling.CDoubleSharp(), false},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tc.want, tc.x.SameAlteration(tc.y))
				assert.Equal(t, tc.want, tc.y.SameAlteration(tc.x))
			})
		}
	})

	t.Run("properties on N", func(t *testing.T) {
		t.Parallel()

		for _, x := range setN() {
			for _, y := range setN() {
				xs, ys := x.build(), y.build()
				got := xs.SameAlteration(ys)

				assert.Equalf(t, x.a.Equal(y.a), got, "%s and %s", x.name, y.name)

				d, err := xs.AlterationDifference(ys)
				require.NoError(t, err)
				assert.Equalf(t, d.IsNatural(), got, "difference for %s and %s", x.name, y.name)
				assert.Equalf(t, got, ys.SameAlteration(xs), "symmetry for %s and %s", x.name, y.name)
			}
		}
	})
}

// K-36 for set names: every name obtained through the API is set. Names read
// from the machine form are checked in the form tests.
func TestIsValid(t *testing.T) {
	t.Parallel()

	half := frac(1, 2)

	for _, tc := range setN() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			x, err := spelling.New(tc.l, tc.a)
			require.NoError(t, err)
			assert.True(t, x.IsValid(), "New")

			results := []namedSpelling{
				{"MustNew", tc.build()},
				{"MustRaise", x.MustRaise()},
				{"MustLower", x.MustLower()},
				{"MustRaiseBy", x.MustRaiseBy(3)},
				{"MustLowerBy", x.MustLowerBy(3)},
				{"Natural", x.Natural()},
				{"WithAlteration", x.WithAlteration(half)},
				{"MustRaiseByAlteration", x.MustRaiseByAlteration(half)},
				{"MustLowerByAlteration", x.MustLowerByAlteration(half)},
			}

			for _, r := range results {
				assert.Truef(t, r.s.IsValid(), "result of %s", r.name)
			}
		})
	}

	t.Run("ready-made names", func(t *testing.T) {
		t.Parallel()

		for _, p := range predefinedCases() {
			assert.Truef(t, p.fn().IsValid(), "%s()", p.name)
		}

		for i, s := range spelling.Predefined() {
			assert.Truef(t, s.IsValid(), "element %d of Predefined", i)
		}
	})
}
