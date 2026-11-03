package spelling_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// unsetNames are the two ways to write the zero value of Spelling, which is
// the unset name.
func unsetNames() []namedSpelling {
	var declared spelling.Spelling

	return []namedSpelling{
		{"var", declared},
		{"literal", spelling.Spelling{}},
	}
}

// K-11, the architecture table "Unset name": every operation that takes a
// name has a defined result for the unset name, and none of them treats it
// as C or as a natural name.
func TestUnsetName(t *testing.T) {
	t.Parallel()

	others := concatNames(setN())
	for _, p := range predefinedCases() {
		others = append(others, nc(p.l, p.a))
	}

	for _, zc := range unsetNames() {
		t.Run(zc.name, func(t *testing.T) {
			t.Parallel()

			z := zc.s

			t.Run("checks", func(t *testing.T) {
				t.Parallel()

				assert.False(t, z.IsValid(), "IsValid")
				assert.False(t, z.IsNatural(), "IsNatural")
				assert.False(t, z.IsInteger(), "IsInteger")
				assert.Equal(t, spelling.DirectionUnset, z.Direction(), "Direction")
				assert.False(t, z.SameLetter(z), "SameLetter with itself")
				assert.False(t, z.SameAlteration(z), "SameAlteration with itself")

				for _, x := range others {
					xs := x.build()
					assert.Falsef(t, z.SameLetter(xs), "SameLetter with %s", x.name)
					assert.Falsef(t, xs.SameLetter(z), "SameLetter of %s", x.name)
					assert.Falsef(t, z.SameAlteration(xs), "SameAlteration with %s", x.name)
					assert.Falsef(t, xs.SameAlteration(z), "SameAlteration of %s", x.name)
				}
			})

			t.Run("equality and order", func(t *testing.T) {
				t.Parallel()

				assert.False(t, z.Equal(z), "Equal with itself")
				assert.Equal(t, spelling.Spelling{}, z, "== to the zero value")
				assert.NotEqual(t, spelling.C(), z, "!= C")
				assert.Equal(t, 0, z.Compare(z), "Compare with itself")
				assert.Equal(t, 0, z.Compare(spelling.Spelling{}), "Compare with the other unset name")

				for _, x := range others {
					xs := x.build()
					assert.Falsef(t, z.Equal(xs), "Equal with %s", x.name)
					assert.Falsef(t, xs.Equal(z), "Equal of %s", x.name)
					assert.Equalf(t, -1, z.Compare(xs), "the unset name is before %s", x.name)
					assert.Equalf(t, 1, xs.Compare(z), "%s is after the unset name", x.name)
				}
			})

			t.Run("parts", func(t *testing.T) {
				t.Parallel()

				assert.False(t, z.Letter().IsValid(), "the letter is invalid")
				assert.NotEqual(t, letter.C, z.Letter(), "the letter is not C")
				assertAlteration(t, alteration.Natural(), z.Alteration())

				l, a := z.Parts()
				assert.Equal(t, z.Letter(), l)
				assert.False(t, l.IsValid())
				assertAlteration(t, alteration.Natural(), a)
			})

			t.Run("letter distances", func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, -1, z.LetterDistanceUp(z))
				assert.Equal(t, -1, z.LetterDistanceDown(z))

				for _, x := range others {
					xs := x.build()
					assert.Equalf(t, -1, z.LetterDistanceUp(xs), "up to %s", x.name)
					assert.Equalf(t, -1, z.LetterDistanceDown(xs), "down to %s", x.name)
					assert.Equalf(t, -1, xs.LetterDistanceUp(z), "up from %s", x.name)
					assert.Equalf(t, -1, xs.LetterDistanceDown(z), "down from %s", x.name)
				}
			})

			t.Run("new names", func(t *testing.T) {
				t.Parallel()

				assertUnset(t, z.Natural())

				for _, a := range setS() {
					assertUnset(t, z.WithAlteration(a.a))
				}
			})

			t.Run("operations with an error", func(t *testing.T) {
				t.Parallel()

				half := frac(1, 2)
				ops := []struct {
					name string
					op   func() (spelling.Spelling, error)
					must func() spelling.Spelling
				}{
					{"Raise", z.Raise, z.MustRaise},
					{"Lower", z.Lower, z.MustLower},
					{"RaiseBy(0)", func() (spelling.Spelling, error) { return z.RaiseBy(0) }, func() spelling.Spelling { return z.MustRaiseBy(0) }},
					{"RaiseBy(1)", func() (spelling.Spelling, error) { return z.RaiseBy(1) }, func() spelling.Spelling { return z.MustRaiseBy(1) }},
					{"RaiseBy(MaxInt)", func() (spelling.Spelling, error) { return z.RaiseBy(math.MaxInt) }, func() spelling.Spelling { return z.MustRaiseBy(math.MaxInt) }},
					{"LowerBy(1)", func() (spelling.Spelling, error) { return z.LowerBy(1) }, func() spelling.Spelling { return z.MustLowerBy(1) }},
					{"LowerBy(MinInt)", func() (spelling.Spelling, error) { return z.LowerBy(math.MinInt) }, func() spelling.Spelling { return z.MustLowerBy(math.MinInt) }},
					{"RaiseByAlteration(1/2)", func() (spelling.Spelling, error) { return z.RaiseByAlteration(half) }, func() spelling.Spelling { return z.MustRaiseByAlteration(half) }},
					{"LowerByAlteration(1/2)", func() (spelling.Spelling, error) { return z.LowerByAlteration(half) }, func() spelling.Spelling { return z.MustLowerByAlteration(half) }},
					{"RaiseByAlteration(M)", func() (spelling.Spelling, error) { return z.RaiseByAlteration(whole(maxMagnitude)) }, func() spelling.Spelling { return z.MustRaiseByAlteration(whole(maxMagnitude)) }},
				}

				for _, o := range ops {
					got, err := o.op()
					if assertInvalidLetter(t, err) {
						assertUnset(t, got)
					}

					assert.Falsef(t, got.Equal(spelling.C()), "%s must not give C", o.name)

					assertPanicsWithErrorIs(t, letter.ErrInvalid, func() { _ = o.must() })
				}
			})

			t.Run("difference", func(t *testing.T) {
				t.Parallel()

				pairs := []struct {
					name string
					x, y spelling.Spelling
				}{
					{"unset minus unset", z, z},
					{"unset minus C", z, spelling.C()},
					{"C minus unset", spelling.C(), z},
					{"unset minus (C, M)", z, spelling.MustNew(letter.C, whole(maxMagnitude))},
					{"(C, -M) minus unset", spelling.MustNew(letter.C, whole(-maxMagnitude)), z},
				}

				for _, p := range pairs {
					got, err := p.x.AlterationDifference(p.y)
					if assertInvalidLetter(t, err) {
						assertAlteration(t, alteration.Natural(), got)
					}

					assertPanicsWithErrorIs(t, letter.ErrInvalid, func() { _ = p.x.MustAlterationDifference(p.y) })
				}
			})

			t.Run("machine form", func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, "unset", z.String())

				text, err := z.MarshalText()
				require.NoError(t, err)
				assert.Equal(t, "unset", string(text))

				appended, err := z.AppendText([]byte("x="))
				require.NoError(t, err)
				assert.Equal(t, "x=unset", string(appended))

				data, err := z.MarshalJSON()
				require.NoError(t, err)
				assert.Equal(t, `"unset"`, string(data))
			})
		})
	}
}

// K-11: a failed operation returns the whole zero value, so its result is
// the unset name with the same behavior.
func TestFailedResultIsUnset(t *testing.T) {
	t.Parallel()

	failed, err := spelling.MustNew(letter.C, whole(maxMagnitude)).Raise()
	require.Error(t, err)

	assert.Equal(t, spelling.Spelling{}, failed)
	assert.False(t, failed.IsValid())
	assert.Equal(t, spelling.DirectionUnset, failed.Direction())
	assert.Equal(t, "unset", failed.String())
	assert.Equal(t, -1, failed.Compare(spelling.CDoubleFlat()))
}
