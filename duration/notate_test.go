package duration_test

import (
	"errors"
	"math"
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestToWrittenExamples covers O-9, ID-4, ID-5, SV-2, KP-7, KP-8, KP-25, and
// KP-26.
func TestToWrittenExamples(t *testing.T) {
	t.Parallel()

	triplet := tuplet.Triplet()

	for _, c := range []struct {
		name  string
		v     duration.Duration
		chain []tuplet.Tuplet
		want  duration.Written
	}{
		{"KP-7 3/8", dur(3, 8), nil, wr(duration.Quarter, 1)},
		{"KP-7 7/16", dur(7, 16), nil, wr(duration.Quarter, 2)},
		{"KP-7 3", dur(3, 1), nil, wr(duration.Breve, 1)},
		{"KP-7 6", dur(6, 1), nil, wr(duration.Longa, 1)},
		{"KP-7 12", dur(12, 1), nil, wr(duration.Maxima, 1)},
		{"KP-7 15", dur(15, 1), nil, wr(duration.Maxima, 3)},
		{"KP-7 3/2048", dur(3, 2048), nil, wr(duration.ThousandTwentyFourth, 1)},
		{"basic quarter", dur(1, 4), nil, wr(duration.Quarter, 0)},
		{"maxima", dur(8, 1), nil, wr(duration.Maxima, 0)},
		{"1024th", dur(1, 1024), nil, wr(duration.ThousandTwentyFourth, 0)},
		{"KP-8 1/3 in 3:2", dur(1, 3), []tuplet.Tuplet{triplet}, wr(duration.Half, 0, triplet)},
		{"KP-8 1/12 in 3:2", dur(1, 12), []tuplet.Tuplet{triplet}, wr(duration.Eighth, 0, triplet)},
		{"chain kept as given", dur(1, 3), []tuplet.Tuplet{tp(6, 4)}, wr(duration.Half, 0, tp(6, 4))},
		{"neutral tuplet", dur(1, 4), []tuplet.Tuplet{tp(1, 1)}, wr(duration.Quarter, 0, tp(1, 1))},
		{"nested", dur(2, 15), []tuplet.Tuplet{tuplet.Quintuplet(), triplet}, wr(duration.Quarter, 0, tuplet.Quintuplet(), triplet)},
		{"dotted in 3:2", dur(1, 4), []tuplet.Tuplet{triplet}, wr(duration.Quarter, 1, triplet)},
		{"maxima with 254 dots", wr(duration.Maxima, 254).MustDuration(), nil, wr(duration.Maxima, 254)},
	} {
		got, err := c.v.ToWritten(c.chain...)
		require.NoError(t, err, c.name)
		assert.Equal(t, c.want, got, c.name)
	}

	for _, c := range []struct {
		name  string
		v     duration.Duration
		chain []tuplet.Tuplet
	}{
		{"KP-25 5/16", dur(5, 16), nil},
		{"KP-25 1/3", dur(1, 3), nil},
		{"KP-25 16", dur(16, 1), nil},
		{"KP-25 1/2048", dur(1, 2048), nil},
		{"KP-26 1/5 in 3:2", dur(1, 5), []tuplet.Tuplet{triplet}},
		{"too long", dur(32, 1), nil},
		{"huge", durRat(pow2(200)), nil},
		{"tiny", durRat(new(big.Rat).SetFrac(big.NewInt(1), maxComponent())), nil},
	} {
		got, err := c.v.ToWritten(c.chain...)
		assertErr(t, err, duration.ErrNotNotatable)
		assert.Equal(t, duration.Written{}, got, c.name)
	}
}

// TestToWrittenModel covers ID-4 and ID-5: ToWritten succeeds exactly when
// the enumeration of 14 bases and k <= 520 dots finds a sign for v' = v/C,
// and that sign is unique.
func TestToWrittenModel(t *testing.T) {
	t.Parallel()

	for _, chain := range testChains() {
		c := chainRat(chain)

		for _, v := range testValues() {
			d := durRat(v)
			signs := singleSigns(new(big.Rat).Quo(v, c))
			require.LessOrEqual(t, len(signs), 1, "ID-4: at most one sign")

			got, err := d.ToWritten(chain...)
			if len(signs) == 0 {
				assertErr(t, err, duration.ErrNotNotatable)
				assert.Equal(t, duration.Written{}, got)

				continue
			}

			require.NoError(t, err, "%s in %v", d, chain)
			assert.Equal(t, wr(signs[0].base, signs[0].dots, chain...), got)
			assert.Equal(t, d, got.MustDuration())
		}
	}
}

// TestTiedToWritten covers O-11, KP-13, KP-28, and architecture 4.8: the
// same rule as Duration.ToWritten on the value of the tie.
func TestTiedToWritten(t *testing.T) {
	t.Parallel()

	triplet := tuplet.Triplet()
	hq := tie(wr(duration.Half, 0), wr(duration.Quarter, 0))
	trip := tie(wr(duration.Quarter, 0, triplet), wr(duration.Eighth, 0, triplet))

	for _, c := range []struct {
		name  string
		t     duration.Tied
		chain []tuplet.Tuplet
		want  duration.Written
	}{
		{"KP-13 half and quarter", hq, nil, wr(duration.Half, 1)},
		{"KP-13 triplets without a tuplet", trip, nil, wr(duration.Quarter, 0)},
		{"KP-13 triplets in 3:2", trip, []tuplet.Tuplet{triplet}, wr(duration.Quarter, 1, triplet)},
	} {
		got, err := c.t.ToWritten(c.chain...)
		require.NoError(t, err, c.name)
		assert.Equal(t, c.want, got, c.name)
	}

	got, err := tie(wr(duration.Quarter, 0), wr(duration.Sixteenth, 0)).ToWritten()
	assertErr(t, err, duration.ErrNotNotatable)
	assert.Equal(t, duration.Written{}, got, "KP-28")

	// The same results as Duration.ToWritten for the value of the tie.
	ties := []duration.Tied{
		hq, trip, tie(wr(duration.Quarter, 0), wr(duration.Sixteenth, 0)),
		tie(wr(duration.Whole, 0), wr(duration.Half, 0), wr(duration.Quarter, 0)),
		tie(wr(duration.Breve, 0), wr(duration.Breve, 0), wr(duration.Breve, 0)),
		tie(wr(duration.Eighth, 0, tp(6, 4)), wr(duration.Eighth, 0, tp(6, 4))),
	}

	for _, tt := range ties {
		for _, chain := range testChains() {
			fromTie, errTie := tt.ToWritten(chain...)
			fromValue, errValue := tt.MustDuration().ToWritten(chain...)
			assert.Equal(t, fromValue, fromTie, "%s in %v", tt, chain)
			assert.Equal(t, errValue == nil, errTie == nil)

			if errTie != nil {
				assertErr(t, errTie, duration.ErrNotNotatable)
			}
		}
	}
}

// writtens builds expected tie parts without tuplets.
func writtens(pairs ...any) []duration.Written {
	var out []duration.Written

	for i := 0; i < len(pairs); i += 2 {
		out = append(out, wr(pairs[i].(duration.Base), pairs[i+1].(int)))
	}

	return out
}

// TestNotateExamples covers O-10, P-3, KP-9 to KP-12, KP-27, and KP-47.
func TestNotateExamples(t *testing.T) {
	t.Parallel()

	const (
		br = duration.Breve
		wh = duration.Whole
		hf = duration.Half
		qu = duration.Quarter
		ei = duration.Eighth
		si = duration.Sixteenth
	)

	triplet := tuplet.Triplet()

	ok := []struct {
		name    string
		v       duration.Duration
		longest duration.Base
		maxDots int
		chain   []tuplet.Tuplet
		want    []duration.Written
	}{
		{"KP-9 5/16", dur(5, 16), br, 0, nil, writtens(qu, 0, si, 0)},
		{"KP-9 5/4", dur(5, 4), br, 0, nil, writtens(wh, 0, qu, 0)},
		{"KP-9 11/16", dur(11, 16), br, 0, nil, writtens(hf, 0, ei, 0, si, 0)},
		{"KP-9 3/8", dur(3, 8), br, 0, nil, writtens(qu, 0, ei, 0)},
		{"KP-9 5", dur(5, 1), br, 0, nil, writtens(br, 0, br, 0, wh, 0)},
		{"one sign", dur(1, 4), br, 0, nil, writtens(qu, 0)},
		{"KP-10 11/16 one dot", dur(11, 16), br, 1, nil, writtens(hf, 0, ei, 1)},
		{"KP-10 7/16 one dot", dur(7, 16), br, 1, nil, writtens(qu, 1, si, 0)},
		{"KP-10 7/16 two dots", dur(7, 16), br, 2, nil, writtens(qu, 2)},
		{"KP-10 6 one dot", dur(6, 1), br, 1, nil, writtens(br, 1, br, 1)},
		{"KP-11 3 to whole", dur(3, 1), wh, 0, nil, writtens(wh, 0, wh, 0, wh, 0)},
		{"KP-11 12 to maxima", dur(12, 1), duration.Maxima, 0, nil, writtens(duration.Maxima, 0, duration.Longa, 0)},
		{"KP-11 12 to maxima one dot", dur(12, 1), duration.Maxima, 1, nil, writtens(duration.Maxima, 1)},
		{"KP-12 5/12 in 3:2", dur(5, 12), br, 0, []tuplet.Tuplet{triplet},
			[]duration.Written{wr(hf, 0, triplet), wr(ei, 0, triplet)}},
		{"KP-27 4 to breve 11 dots", dur(4, 1), br, 11, nil, writtens(br, 11, duration.ThousandTwentyFourth, 0)},
		{"3/1024", dur(3, 1024), br, 0, nil, writtens(duration.FiveHundredTwelfth, 0, duration.ThousandTwentyFourth, 0)},
		{"large maxDots short value", dur(1, 4), br, math.MaxInt, nil, writtens(qu, 0)},
		{"large maxDots below 2L", dur(3, 1), br, math.MaxInt, nil, writtens(br, 1)},
		{"1024th", dur(1, 1024), duration.ThousandTwentyFourth, 0, nil, writtens(duration.ThousandTwentyFourth, 0)},
	}

	for _, c := range ok {
		got, err := c.v.NotateWith(c.longest, c.maxDots, c.chain...)
		require.NoError(t, err, c.name)
		assert.Equal(t, c.want, got, c.name)

		if c.longest == br && c.maxDots == 0 {
			def, err := c.v.Notate(c.chain...)
			require.NoError(t, err, c.name)
			assert.Equal(t, c.want, def, "%s: Notate is NotateWith(Breve, 0)", c.name)
		}
	}

	// KP-47: 2048 is 1024 breves.
	parts, err := dur(2048, 1).Notate()
	require.NoError(t, err)
	require.Len(t, parts, duration.MaxParts)

	for _, p := range parts {
		assert.Equal(t, wr(br, 0), p)
	}

	failing := []struct {
		name    string
		v       duration.Duration
		longest duration.Base
		maxDots int
		chain   []tuplet.Tuplet
		want    error
	}{
		{"KP-27 1/3", dur(1, 3), br, 0, nil, duration.ErrNotNotatable},
		{"KP-27 1/2048", dur(1, 2048), br, 0, nil, duration.ErrNotNotatable},
		{"KP-27 5/2048 one dot", dur(5, 2048), br, 1, nil, duration.ErrNotNotatable},
		{"KP-27 4 to breve 12 dots", dur(4, 1), br, 12, nil, duration.ErrNotNotatable},
		{"KP-27 1/512 to 1024th one dot", dur(1, 512), duration.ThousandTwentyFourth, 1, nil, duration.ErrNotNotatable},
		{"4 to breve 1021 dots", dur(4, 1), br, 1021, nil, duration.ErrNotNotatable},
		{"4 to breve max dots", dur(4, 1), br, math.MaxInt, nil, duration.ErrNotNotatable},
		{"1/5 in 3:2", dur(1, 5), br, 0, []tuplet.Tuplet{triplet}, duration.ErrNotNotatable},
		{"KP-47 2049", dur(2049, 1), br, 0, nil, duration.ErrTooManyParts},
		{"1025 quarters", dur(1025, 4), qu, 0, nil, duration.ErrTooManyParts},
	}

	for _, c := range failing {
		got, err := c.v.NotateWith(c.longest, c.maxDots, c.chain...)
		assertErr(t, err, c.want)
		assert.Nil(t, got, c.name)
	}

	got, err := dur(1, 3).Notate()
	assertErr(t, err, duration.ErrNotNotatable)
	assert.Nil(t, got)

	got, err = dur(2049, 1).Notate()
	assertErr(t, err, duration.ErrTooManyParts)
	assert.Nil(t, got)
}

// TestNotateFast covers architecture 4.8 and 6: E-9 and E-5 are found
// without building parts, so the time does not depend on the value; E-5
// precedes E-9.
func TestNotateFast(t *testing.T) {
	t.Parallel()

	huge := durRat(pow2(200))
	maxValue := durRat(ratInt(maxComponent()))

	for _, c := range []struct {
		name    string
		v       duration.Duration
		longest duration.Base
		maxDots int
		want    error
	}{
		{"2^200 breves", huge, duration.Breve, 0, duration.ErrTooManyParts},
		{"max component", maxValue, duration.Breve, 0, duration.ErrTooManyParts},
		{"max component to 1024th", maxValue, duration.ThousandTwentyFourth, 3, duration.ErrTooManyParts},
		{"2^200 one dot", huge, duration.Breve, 1, duration.ErrTooManyParts},
		{"E-5 before E-9: 2^200 to maxima with 300 dots", huge, duration.Maxima, 300, duration.ErrNotNotatable},
		{"2^200 large maxDots", huge, duration.Breve, math.MaxInt, duration.ErrNotNotatable},
		{"2^200 1021 dots", huge, duration.Breve, 1021, duration.ErrNotNotatable},
		{"E-5 before E-9: non-dyadic", durRat(new(big.Rat).Add(pow2(200), rat(1, 3))), duration.Breve, 0, duration.ErrNotNotatable},
		{"E-5 before E-9: below 1024th", durRat(new(big.Rat).Add(pow2(200), rat(1, 2048))), duration.Breve, 0, duration.ErrNotNotatable},
		{"E-5 before E-9: 2049 and a third", durRat(rat(6148, 3)), duration.Breve, 0, duration.ErrNotNotatable},
	} {
		r := fast(t, func() result[[]duration.Written] {
			parts, err := c.v.NotateWith(c.longest, c.maxDots)

			return result[[]duration.Written]{parts, err}
		})
		assertErr(t, r.err, c.want)
		assert.Nil(t, r.v, c.name)
	}

	r := fast(t, func() result[[]duration.Written] {
		parts, err := huge.Notate()

		return result[[]duration.Written]{parts, err}
	})
	assertErr(t, r.err, duration.ErrTooManyParts)
}

// TestNotateErrors covers architecture 4.2 and 4.8: argument errors and E-11
// for a part out of range.
func TestNotateErrors(t *testing.T) {
	t.Parallel()

	v := dur(5, 16)

	for _, c := range []struct {
		name    string
		longest duration.Base
		maxDots int
		chain   []tuplet.Tuplet
		want    error
		causes  []error
	}{
		{"longest zero", 0, 0, nil, duration.ErrInvalidArgument, nil},
		{"longest 15", 15, 0, nil, duration.ErrInvalidArgument, nil},
		{"KP-45 negative maxDots", duration.Breve, -1, nil, duration.ErrInvalidArgument, nil},
		{"min int maxDots", duration.Breve, math.MinInt, nil, duration.ErrInvalidArgument, nil},
		{"unset tuplet", duration.Breve, 0, []tuplet.Tuplet{{}}, duration.ErrInvalidArgument, []error{tuplet.ErrUnset}},
		{"coefficient", duration.Breve, 0, bigChain(5), duration.ErrOutOfRange, []error{tuplet.ErrOutOfRange, fraction.ErrOutOfRange}},
	} {
		got, err := v.NotateWith(c.longest, c.maxDots, c.chain...)
		assertErr(t, err, c.want, c.causes...)
		assert.Nil(t, got, c.name)
	}

	got, err := v.Notate(tuplet.Tuplet{})
	assertErr(t, err, duration.ErrInvalidArgument, tuplet.ErrUnset)
	assert.Nil(t, got)

	got, err = v.Notate(bigChain(5)...)
	assertErr(t, err, duration.ErrOutOfRange, tuplet.ErrOutOfRange, fraction.ErrOutOfRange)
	assert.Nil(t, got)

	w, err := v.ToWritten(tuplet.Tuplet{})
	assertErr(t, err, duration.ErrInvalidArgument, tuplet.ErrUnset)
	assert.Equal(t, duration.Written{}, w)

	w, err = v.ToWritten(bigChain(5)...)
	assertErr(t, err, duration.ErrOutOfRange, tuplet.ErrOutOfRange, fraction.ErrOutOfRange)
	assert.Equal(t, duration.Written{}, w)

	// A quarter under four MaxInt:1 is in range; written up to the eighth
	// with one dot it is a dotted eighth and a sixteenth, and those parts
	// under the same chain are out of range.
	chain := bigChain(4)
	edgeValue := wr(duration.Quarter, 0, chain...).MustDuration()

	got, err = edgeValue.NotateWith(duration.Eighth, 1, chain...)
	assertErr(t, err, duration.ErrOutOfRange, fraction.ErrOutOfRange)
	assert.Nil(t, got)

	got, err = edgeValue.Notate(chain...)
	require.NoError(t, err)
	assert.Equal(t, []duration.Written{wr(duration.Quarter, 0, chain...)}, got)
}

// TestNotateModel covers P-3 and architecture 4.8 on sets of values, chains,
// L, and D <= 3: agreement with the model of the rule, and the properties of
// a result: the sum is v, parts do not increase, every part is an allowed
// sign with the same chain, the length is 1 to 1024, a single allowed sign
// is returned alone.
func TestNotateModel(t *testing.T) {
	t.Parallel()

	var values []*big.Rat

	for _, d := range []int64{1, 2, 3, 8, 12, 64, 1024, 2048} {
		for _, n := range []int64{1, 2, 3, 5, 7, 11, 13, 16, 23, 31, 33, 47} {
			values = append(values, rat(n, d))
		}
	}

	chains := [][]tuplet.Tuplet{nil, {tuplet.Triplet()}, {tuplet.Quintuplet(), tp(6, 4)}}

	for _, row := range baseTable {
		for maxDots := range 4 {
			for _, chain := range chains {
				c := chainRat(chain)

				for _, v := range values {
					d := durRat(v)
					vPrime := new(big.Rat).Quo(v, c)

					got, err := d.NotateWith(row.base, maxDots, chain...)
					checkNotateProperties(t, d, row.base, maxDots, chain, got, err)

					m := greedyTie(vPrime, row.base, maxDots)

					switch {
					case m.outcome == tieNotNotable:
						assertErr(t, err, duration.ErrNotNotatable)
					case m.count.Cmp(big.NewInt(duration.MaxParts)) > 0:
						assertErr(t, err, duration.ErrTooManyParts)
					default:
						require.NoError(t, err, "%s to %s with %d dots in %v", d, row.text, maxDots, chain)
						assert.Equal(t, writtenOfSigns(m.parts, chain), got)
					}
				}
			}
		}
	}
}

// checkNotateProperties checks the properties of architecture 4.8 that hold
// for any successful result, and the kind of a failure.
func checkNotateProperties(t *testing.T, d duration.Duration, longest duration.Base, maxDots int,
	chain []tuplet.Tuplet, got []duration.Written, err error,
) {
	t.Helper()

	if err != nil {
		assert.Nil(t, got)

		if !errorsIsAny(err, duration.ErrNotNotatable, duration.ErrTooManyParts, duration.ErrOutOfRange) {
			t.Errorf("unexpected error %v", err)
		}

		return
	}

	require.NotEmpty(t, got)
	require.LessOrEqual(t, len(got), duration.MaxParts)

	sum := new(big.Rat)
	vPrime := new(big.Rat).Quo(ratOfDuration(d), chainRat(chain))
	single := false

	for _, s := range allowedSigns(longest, maxDots) {
		if s.v.Cmp(vPrime) == 0 {
			single = true
		}
	}

	if single {
		assert.Len(t, got, 1, "a single allowed sign is returned alone")
	}

	for i, p := range got {
		sum.Add(sum, ratOfDuration(p.MustDuration()))

		assert.Equal(t, chain, p.MustChain(), "every part has the chain")
		assert.LessOrEqual(t, int(longest), int(p.MustBase()), "no base longer than longest")
		assert.LessOrEqual(t, p.MustDots(), maxDots)

		if i > 0 {
			assert.LessOrEqual(t, p.MustCompareDuration(got[i-1]), 0, "parts do not increase")
		}
	}

	assert.Equal(t, 0, sum.Cmp(ratOfDuration(d)), "the parts sum to the value")
}

// errorsIsAny reports whether err matches any of the targets.
func errorsIsAny(err error, targets ...error) bool {
	for _, target := range targets {
		if errors.Is(err, target) {
			return true
		}
	}

	return false
}

// TestNotateResultOwnership covers architecture 3.2: each call returns a new
// slice that belongs to the caller.
func TestNotateResultOwnership(t *testing.T) {
	t.Parallel()

	first := dur(5, 16).MustNotate()
	first[0] = wr(duration.Whole, 0)

	second := dur(5, 16).MustNotate()
	assert.Equal(t, writtens(duration.Quarter, 0, duration.Sixteenth, 0), second)

	third := dur(5, 16).MustNotateWith(duration.Breve, 0)
	third[1] = wr(duration.Whole, 0)
	assert.Equal(t, writtens(duration.Quarter, 0, duration.Sixteenth, 0), dur(5, 16).MustNotateWith(duration.Breve, 0))
}
