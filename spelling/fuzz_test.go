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

// FuzzParse checks the machine form on any string: reading never panics,
// a failed reading is a form error with the unset name, and an accepted
// string is the only form of its name (canonical form).
func FuzzParse(f *testing.F) {
	seeds := []string{"C", "B", "C+1", "D-1", "C+1/2", "B-7/2", "C+16777216", "C-1/65536", "unset"}
	for _, in := range invalidForms() {
		seeds = append(seeds, in.text)
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, text string) {
		var (
			s   spelling.Spelling
			err error
		)

		require.NotPanics(t, func() { s, err = spelling.Parse(text) })

		if err != nil {
			assert.ErrorIs(t, err, spelling.ErrInvalidForm)
			assert.NotErrorIs(t, err, letter.ErrInvalid)
			assert.Equal(t, spelling.Spelling{}, s)

			return
		}

		assert.Equal(t, text, s.String(), "an accepted string must be the form of its name")

		back, err := spelling.Parse(s.String())
		require.NoError(t, err)
		assert.Equal(t, s, back)
	})
}

// FuzzRaiseBy checks raising and lowering by n against the same operations
// of alteration, for any letter ordinal, value, and n.
func FuzzRaiseBy(f *testing.F) {
	f.Add(uint8(0), 1, 2, 3)
	f.Add(uint8(6), -7, 2, -5)
	f.Add(uint8(0), maxMagnitude, 1, 1)
	f.Add(uint8(0), -maxMagnitude, 1, math.MaxInt)
	f.Add(uint8(3), 1, maxDenominator, math.MinInt)
	f.Add(uint8(2), 0, 1, 2*maxMagnitude)

	f.Fuzz(func(t *testing.T, ordinal uint8, num, den, n int) {
		a, err := alteration.FromFraction(num, den)
		if err != nil {
			t.Skip("not an alteration value")
		}

		l := letter.MustFromOrdinal(int(ordinal % letter.Count))
		x := spelling.MustNew(l, a)

		checks := []struct {
			name   string
			got    func() (spelling.Spelling, error)
			oracle func() (alteration.Alteration, error)
		}{
			{"RaiseBy", func() (spelling.Spelling, error) { return x.RaiseBy(n) }, func() (alteration.Alteration, error) { return a.RaiseBy(n) }},
			{"LowerBy", func() (spelling.Spelling, error) { return x.LowerBy(n) }, func() (alteration.Alteration, error) { return a.LowerBy(n) }},
		}

		for _, c := range checks {
			want, wantErr := c.oracle()
			got, err := c.got()

			if wantErr != nil {
				assert.ErrorIs(t, err, alteration.ErrOutOfRange, c.name)
				assert.Equal(t, spelling.Spelling{}, got, c.name)

				continue
			}

			if assert.NoError(t, err, c.name) {
				assert.Equal(t, l, got.Letter(), c.name)
				assert.True(t, want.Equal(got.Alteration()), c.name)
			}
		}
	})
}
