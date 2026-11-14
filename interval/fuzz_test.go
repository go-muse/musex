package interval_test

import (
	"errors"
	"testing"

	"github.com/go-muse/musex/interval"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T-7, fuzzing of parsing (plan, "Fuzz"; architecture 4.12, 4.13): no panic
// on any input; a parsed interval prints back to an equal one; a failure is
// one of the parse failures and gives the zero value.

func FuzzParse(f *testing.F) {
	for _, s := range []string{
		"P1", "-P1", "M3", "-m6", "n3", "AA4", "dddd5", "{+1/3}2", "-{+1/2}1", "{+16777216}4", "aug4", "°5",
		"perf5", "Maj3", "TT", "", "{+2/4}3", "M9223372036854775808", "P9223372036854775808", "{+1/65537}3",
		"M 3", "--m3", "ё", "\xff",
	} {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, s string) {
		var (
			i   interval.Interval
			err error
		)

		noPanic(t, func() { i, err = interval.Parse(s) })

		if err != nil {
			assert.Equal(t, interval.Interval{}, i)
			assert.True(t, errors.Is(err, interval.ErrParse) || errors.Is(err, interval.ErrInvalidArgument) ||
				errors.Is(err, interval.ErrOutOfRange), "%q: %v", s, err)

			return
		}

		require.True(t, i.IsValid(), "%q", s)

		back, err := interval.Parse(i.String())
		require.NoError(t, err, "%q → %q", s, i.String())
		assert.True(t, back.Equal(i), "%q → %q → %v", s, i.String(), back)

		var u interval.Interval
		require.NoError(t, u.UnmarshalText([]byte(s)))
		assert.True(t, u.Equal(i))
	})
}

func FuzzParseName(f *testing.F) {
	for _, s := range []string{
		"major third", "perfect unison", "double octave", "tritave", "medium second", "perfect 8ve", "ум. 5",
		"дв.ум.5", "нейтральная терция", "tritone", "major fifth", "", "major  third", "ч.16",
	} {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, s string) {
		var (
			i   interval.Interval
			err error
		)

		noPanic(t, func() { i, err = interval.ParseName(s) })

		if err != nil {
			assert.Equal(t, interval.Interval{}, i)
			assert.True(t, errors.Is(err, interval.ErrParse) || errors.Is(err, interval.ErrInvalidArgument), "%q: %v", s, err)

			return
		}

		require.True(t, i.IsValid(), "%q", s)
		assert.True(t, i.IsAscending(), "%q", s)
		assert.LessOrEqual(t, i.Size(), 15, "%q", s)
	})
}
