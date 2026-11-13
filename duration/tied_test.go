package duration_test

import (
	"math"
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewTied covers O-8, ID-2, P-5, KP-5, T-3, and architecture 4.7: two or
// more parts in order, each with its own chain; the value is the sum.
func TestNewTied(t *testing.T) {
	t.Parallel()

	h, q := wr(duration.Half, 0), wr(duration.Quarter, 0)

	hq, err := duration.NewTied(h, q)
	require.NoError(t, err)
	assert.True(t, hq.IsSet())
	assert.Equal(t, 2, hq.MustLen())
	assert.Equal(t, []duration.Written{h, q}, hq.MustParts())
	assert.Equal(t, dur(3, 4), hq.MustDuration(), "KP-5")

	// KP-5: equal to a dotted half by value only; the types differ, so the
	// comparison goes through Duration.
	assert.Equal(t, wr(duration.Half, 1).MustDuration(), hq.MustDuration())

	mixed := tie(wr(duration.Quarter, 0, tuplet.Triplet()), wr(duration.Eighth, 1), wr(duration.Sixteenth, 0, tuplet.Quintuplet(), tp(6, 4)))
	want := new(big.Rat)
	for _, p := range mixed.MustParts() {
		want.Add(want, ratOfDuration(p.MustDuration()))
	}

	assert.Equal(t, durRat(want), mixed.MustDuration())
	assert.Equal(t, 3, mixed.MustLen())

	// The order of parts is kept: 2~4 and 4~2 are different records.
	qh := tie(q, h)
	assert.NotEqual(t, hq, qh)
	assert.False(t, hq.Equal(qh))
	assert.True(t, hq.MustEqualDuration(qh))

	// Many parts.
	parts := make([]duration.Written, 1025)
	for i := range parts {
		parts[i] = wr(duration.Breve, 0)
	}

	many := tie(parts...)
	assert.Equal(t, 1025, many.MustLen())
	assert.Equal(t, dur(2050, 1), many.MustDuration())
}

// TestNewTiedErrors covers KP-35 and architecture 4.7.
func TestNewTiedErrors(t *testing.T) {
	t.Parallel()

	q := wr(duration.Quarter, 0)

	// Coprime chains whose sum has a denominator near 2^380.
	n1, n2 := math.MaxInt, math.MaxInt-2
	a := wr(duration.Quarter, 0, tp(n1, 1), tp(n1, 1), tp(n1, 1))
	b := wr(duration.Quarter, 0, tp(n2, 1), tp(n2, 1), tp(n2, 1))

	for _, c := range []struct {
		name   string
		parts  []duration.Written
		want   error
		causes []error
	}{
		{"KP-35 no parts", nil, duration.ErrInvalidArgument, nil},
		{"KP-35 one part", []duration.Written{q}, duration.ErrInvalidArgument, nil},
		{"one unset part", []duration.Written{{}}, duration.ErrInvalidArgument, nil},
		{"unset part", []duration.Written{q, {}}, duration.ErrUnset, nil},
		{"unset first part", []duration.Written{{}, q, q}, duration.ErrUnset, nil},
		{"sum out of range", []duration.Written{a, b}, duration.ErrOutOfRange, []error{fraction.ErrOutOfRange}},
	} {
		got, err := duration.NewTied(c.parts...)
		assertErr(t, err, c.want, c.causes...)
		assert.Equal(t, duration.Tied{}, got, c.name)
	}
}

// TestTiedExactSum covers KR-D11: the sum is checked exactly, so partial
// sums out of range do not matter and the order of parts does not matter.
func TestTiedExactSum(t *testing.T) {
	t.Parallel()

	// a = 2 − 2^−199 (whole with 199 dots), b = 1/MaxInt² (whole under two
	// MaxInt:1), c = 2^−199 (1024th under 2^62:1 three times and 8:1).
	a := wr(duration.Whole, 199)
	b := wr(duration.Whole, 0, bigChain(2)...)
	c := wr(duration.ThousandTwentyFourth, 0, tp(1<<62, 1), tp(1<<62, 1), tp(1<<62, 1), tp(8, 1))
	require.Equal(t, durRat(pow2(-199)), c.MustDuration())

	want := durRat(new(big.Rat).Add(big.NewRat(2, 1), ratOfDuration(b.MustDuration())))

	for _, order := range [][]duration.Written{{a, b, c}, {b, a, c}, {a, c, b}, {c, b, a}} {
		got, err := duration.NewTied(order...)
		require.NoError(t, err)
		assert.Equal(t, want, got.MustDuration())
	}

	_, err := duration.NewTied(a, b)
	assertErr(t, err, duration.ErrOutOfRange, fraction.ErrOutOfRange)
}

// TestTiedAppendJoin covers O-8 and KP-51.
func TestTiedAppendJoin(t *testing.T) {
	t.Parallel()

	h, q, e, s := wr(duration.Half, 0), wr(duration.Quarter, 0), wr(duration.Eighth, 0), wr(duration.Sixteenth, 0)

	appended, err := tie(h, q).Append(e)
	require.NoError(t, err)
	assert.Equal(t, tie(h, q, e), appended)
	assert.Equal(t, 3, appended.MustLen())
	assert.Equal(t, dur(7, 8), appended.MustDuration())

	joined, err := tie(q, s).Join(tie(h, e))
	require.NoError(t, err)
	assert.Equal(t, tie(q, s, h, e), joined)
	assert.Equal(t, []duration.Written{q, s, h, e}, joined.MustParts())
	assert.Equal(t, dur(15, 16), joined.MustDuration())

	// Errors: unset arguments and sums out of range.
	n1, n2 := math.MaxInt, math.MaxInt-2
	a := wr(duration.Quarter, 0, tp(n1, 1), tp(n1, 1), tp(n1, 1))
	b := wr(duration.Quarter, 0, tp(n2, 1), tp(n2, 1), tp(n2, 1))

	got, err := tie(q, q).Append(duration.Written{})
	assertErr(t, err, duration.ErrUnset)
	assert.Equal(t, duration.Tied{}, got)

	got, err = tie(q, q).Join(duration.Tied{})
	assertErr(t, err, duration.ErrUnset)
	assert.Equal(t, duration.Tied{}, got)

	got, err = tie(a, a).Append(b)
	assertErr(t, err, duration.ErrOutOfRange, fraction.ErrOutOfRange)
	assert.Equal(t, duration.Tied{}, got)

	got, err = tie(a, a).Join(tie(b, b))
	assertErr(t, err, duration.ErrOutOfRange, fraction.ErrOutOfRange)
	assert.Equal(t, duration.Tied{}, got)
}

// TestTiedPartsOwnership covers architecture 3.2: Parts returns a new slice.
func TestTiedPartsOwnership(t *testing.T) {
	t.Parallel()

	tied := tie(wr(duration.Half, 0), wr(duration.Quarter, 0))

	parts := tied.MustParts()
	parts[0] = wr(duration.Whole, 0)

	assert.Equal(t, []duration.Written{wr(duration.Half, 0), wr(duration.Quarter, 0)}, tied.MustParts())
}
