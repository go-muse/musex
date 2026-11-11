package rounding_test

import (
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/rounding"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBigInputsAreNotChanged(t *testing.T) {
	t.Parallel()
	n, d := bigInt(7), bigInt(2)
	x, step := bigRat(7, 3), bigRat(1, 4)
	before := []string{n.String(), d.String(), x.RatString(), step.RatString()}
	q, r, err := rounding.HalfEven.QuoRemBig(n, d)
	require.NoError(t, err)
	q.SetInt64(100)
	r.SetInt64(200)
	integer, err := rounding.HalfUp.RatInt(x)
	require.NoError(t, err)
	integer.SetInt64(300)
	k, v, err := rounding.HalfEven.RatMultiple(x, step)
	require.NoError(t, err)
	k.SetInt64(400)
	v.SetInt64(500)
	_, err = rounding.HalfEven.RatDecimal(x, 2)
	require.NoError(t, err)
	assert.Equal(t, before, []string{n.String(), d.String(), x.RatString(), step.RatString()})
}

func TestBigResultsDoNotAliasInputsOrEachOther(t *testing.T) {
	t.Parallel()
	t.Run("quo_rem", func(t *testing.T) {
		t.Parallel()
		n, d := bigInt(7), bigInt(2)
		q, r, err := rounding.HalfEven.QuoRemBig(n, d)
		require.NoError(t, err)
		q.SetInt64(100)
		assert.Equal(t, "-1", r.String())
		assert.Equal(t, "7", n.String())
		assert.Equal(t, "2", d.String())
		r.SetInt64(200)
		assert.Equal(t, "100", q.String())
	})
	t.Run("rat_multiple_k_one", func(t *testing.T) {
		t.Parallel()
		x, step := bigRat(1, 2), bigRat(1, 2)
		k, v, err := rounding.HalfEven.RatMultiple(x, step)
		require.NoError(t, err)
		assert.NotSame(t, step, v)
		v.SetInt64(100)
		assert.Equal(t, "1/2", step.RatString())
		assert.Equal(t, "1/2", x.RatString())
		assert.Equal(t, "1", k.String())
		k.SetInt64(200)
		assert.Equal(t, "100", v.RatString())
	})
}

func TestBigInputsUnchangedOnFailure(t *testing.T) {
	t.Parallel()
	n, d := bigInt(7), new(big.Int)
	x, step := bigRat(7, 3), new(big.Rat)
	before := []string{n.String(), d.String(), x.RatString(), step.RatString()}
	_, _, err := rounding.HalfEven.QuoRemBig(n, d)
	checkError(t, err, rounding.ErrInvalidArgument)
	_, _, err = rounding.HalfEven.RatMultiple(x, step)
	checkError(t, err, rounding.ErrInvalidArgument)
	_, err = rounding.HalfEven.RatDecimal(x, rounding.MaxPlaces+1)
	checkError(t, err, rounding.ErrInvalidArgument)
	assert.Equal(t, before, []string{n.String(), d.String(), x.RatString(), step.RatString()})
}
