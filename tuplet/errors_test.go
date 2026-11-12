package tuplet_test

import (
	"errors"
	"math"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
)

// TestSentinels covers K-13 and architecture 4.1, 4.2: the seven sentinels
// are non-nil, pairwise distinct by errors.Is, none wraps another, and none
// matches a common/fraction sentinel.
func TestSentinels(t *testing.T) {
	t.Parallel()

	for _, a := range tupletSentinels {
		assert.Error(t, a.err, a.name)

		for _, b := range tupletSentinels {
			assert.Equal(t, a.name == b.name, errors.Is(a.err, b.err), "errors.Is(%s, %s)", a.name, b.name)
		}

		for _, f := range fractionSentinels {
			assert.False(t, errors.Is(a.err, f.err), "errors.Is(tuplet.%s, fraction.%s)", a.name, f.name)
			assert.False(t, errors.Is(f.err, a.err), "errors.Is(fraction.%s, tuplet.%s)", f.name, a.name)
		}
	}
}

// TestErrors_Classes covers K-13: each error of the package matches exactly
// one package sentinel, and causes from common/fraction are kept exactly
// where architecture 4.2 lists them.
func TestErrors_Classes(t *testing.T) {
	t.Parallel()

	maxInt := tuplet.MustNew(1, math.MaxInt)
	maxFrac := maxComponentFrac(t)

	tests := []struct {
		name   string
		call   func() error
		want   error
		causes []error
	}{
		{name: "E-1 New", call: func() error { _, err := tuplet.New(0, 1); return err }, want: tuplet.ErrInvalidCount},
		{name: "E-1 FromCoefficient", call: func() error {
			_, err := tuplet.FromCoefficient(frac(-1, 2))
			return err
		}, want: tuplet.ErrInvalidCount},
		{name: "E-2 Coefficient", call: func() error { _, err := tuplet.Tuplet{}.Coefficient(); return err }, want: tuplet.ErrUnset},
		{name: "E-3 Default", call: func() error { _, err := tuplet.Default(4, tuplet.Binary); return err }, want: tuplet.ErrNoDefault},
		{name: "E-4 Parse", call: func() error { _, err := tuplet.Parse("x"); return err }, want: tuplet.ErrInvalidText},
		{name: "E-5 Scale", call: func() error { _, err := tuplet.MustNew(math.MaxInt, 1).Scale(2); return err }, want: tuplet.ErrOutOfRange},
		{name: "E-5 Flatten", call: func() error {
			_, err := tuplet.Flatten(maxInt, tuplet.MustNew(1, 2))
			return err
		}, want: tuplet.ErrOutOfRange},
		{name: "E-5 Parse", call: func() error {
			_, err := tuplet.Parse("9223372036854775808:1")
			return err
		}, want: tuplet.ErrOutOfRange},
		{name: "E-5 FromCoefficient", call: func() error {
			_, err := tuplet.FromCoefficient(frac(math.MinInt, -1))
			return err
		}, want: tuplet.ErrOutOfRange},
		{name: "E-5 ChainCoefficient", call: func() error {
			_, err := tuplet.ChainCoefficient(maxInt, maxInt, maxInt, maxInt, maxInt)
			return err
		}, want: tuplet.ErrOutOfRange, causes: []error{fraction.ErrOutOfRange}},
		{name: "E-5 Apply", call: func() error {
			_, err := tuplet.Duplet().Apply(maxFrac)
			return err
		}, want: tuplet.ErrOutOfRange, causes: []error{fraction.ErrOutOfRange}},
		{name: "E-5 FromUnitRatio", call: func() error {
			_, err := tuplet.FromUnitRatio(1, math.MaxInt, frac(2, 1))
			return err
		}, want: tuplet.ErrOutOfRange, causes: []error{fraction.ErrNotRepresentable}},
		{name: "E-6 Scale", call: func() error { _, err := tuplet.Triplet().Scale(0); return err }, want: tuplet.ErrInvalidArgument},
		{name: "E-6 Apply", call: func() error {
			_, err := tuplet.Triplet().Apply(fraction.Fraction{})
			return err
		}, want: tuplet.ErrInvalidArgument},
		{name: "E-7 FromUnitRatio", call: func() error {
			_, err := tuplet.FromUnitRatio(3, 3, frac(1, 2))
			return err
		}, want: tuplet.ErrNotRepresentable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assertErr(t, tt.call(), tt.want, tt.causes...)
		})
	}
}

// TestErrors_NotRepresentableDistinction covers architecture 4.2: E-7 does not
// match fraction.ErrNotRepresentable, and E-5 of FromUnitRatio matches
// fraction.ErrNotRepresentable but not tuplet.ErrNotRepresentable.
func TestErrors_NotRepresentableDistinction(t *testing.T) {
	t.Parallel()

	_, err := tuplet.FromUnitRatio(3, 1, frac(1, 2))
	assert.ErrorIs(t, err, tuplet.ErrNotRepresentable)
	assert.NotErrorIs(t, err, fraction.ErrNotRepresentable)

	_, err = tuplet.FromUnitRatio(1, math.MaxInt, frac(2, 1))
	assert.ErrorIs(t, err, tuplet.ErrOutOfRange)
	assert.ErrorIs(t, err, fraction.ErrNotRepresentable)
	assert.NotErrorIs(t, err, tuplet.ErrNotRepresentable)
}
