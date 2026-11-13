package duration_test

import (
	"math/big"
	"reflect"
	"slices"
	"testing"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sampleWrittens are written durations with equal and different values and
// records, including the unset value.
func sampleWrittens() []duration.Written {
	triplet := tuplet.Triplet()

	return []duration.Written{
		{},
		wr(duration.Quarter, 0),
		wr(duration.Half, 0, tp(2, 1)),
		wr(duration.Quarter, 0, tp(1, 1)),
		wr(duration.Quarter, 1, triplet),
		wr(duration.Quarter, 0, triplet),
		wr(duration.Quarter, 0, tp(6, 4)),
		wr(duration.Quarter, 0, triplet, tuplet.Quintuplet()),
		wr(duration.Quarter, 0, tuplet.Quintuplet(), triplet),
		wr(duration.Half, 1),
		wr(duration.Quarter, 1),
		wr(duration.Eighth, 0, tuplet.Duplet()),
		wr(duration.Breve, 2),
		wr(duration.ThousandTwentyFourth, 0),
		wr(duration.Maxima, 3),
	}
}

// sampleTieds are tied durations with equal and different values and
// records, including the unset value.
func sampleTieds() []duration.Tied {
	h, q, e := wr(duration.Half, 0), wr(duration.Quarter, 0), wr(duration.Eighth, 0)

	return []duration.Tied{
		{},
		tie(h, q),
		tie(q, h),
		tie(q, q, q),
		tie(h, e, e),
		tie(q, e),
		tie(wr(duration.Quarter, 0, tuplet.Triplet()), wr(duration.Eighth, 0, tuplet.Triplet())),
		tie(wr(duration.Eighth, 0, tuplet.Triplet()), wr(duration.Quarter, 0, tuplet.Triplet())),
		tie(h, h),
	}
}

// checkOrder checks that compare is a total order consistent with ==:
// compare(a, b) == 0 exactly when a == b, antisymmetry, and transitivity;
// sorting gives the same result for any input order.
func checkOrder[T comparable](t *testing.T, values []T, compare func(a, b T) int) {
	t.Helper()

	for _, a := range values {
		for _, b := range values {
			ab, ba := compare(a, b), compare(b, a)
			assert.Contains(t, []int{-1, 0, 1}, ab)
			assert.Equal(t, a == b, ab == 0, "%v and %v", a, b)
			assert.Equal(t, -ab, ba, "antisymmetry of %v and %v", a, b)

			for _, c := range values {
				if ab <= 0 && compare(b, c) <= 0 {
					assert.LessOrEqual(t, compare(a, c), 0, "transitivity of %v, %v, %v", a, b, c)
				}
			}
		}
	}

	sorted := slices.Clone(values)
	slices.SortFunc(sorted, compare)

	reversed := slices.Clone(values)
	slices.Reverse(reversed)
	slices.SortFunc(reversed, compare)
	assert.Equal(t, sorted, reversed, "the order does not depend on the input order")
}

// TestComparable covers architecture 3.1 and KR-D5: the five types are
// comparable with == and usable as map keys.
func TestComparable(t *testing.T) {
	t.Parallel()

	assert.True(t, reflect.TypeFor[duration.Base]().Comparable())
	assert.True(t, reflect.TypeFor[duration.Duration]().Comparable())
	assert.True(t, reflect.TypeFor[duration.Written]().Comparable())
	assert.True(t, reflect.TypeFor[duration.Tied]().Comparable())
	assert.True(t, reflect.TypeFor[duration.Position]().Comparable())

	written := map[duration.Written]int{}
	for i, w := range sampleWrittens() {
		written[w] = i
	}

	assert.Len(t, written, len(sampleWrittens()), "different records are different keys")
	assert.Equal(t, 6, written[duration.MustParseWritten("4[6:4]")])
	assert.Equal(t, 5, written[wr(duration.Quarter, 0, tuplet.Triplet())], "an equal record built again is the same key")

	tied := map[duration.Tied]int{}
	for i, tt := range sampleTieds() {
		tied[tt] = i
	}

	assert.Len(t, tied, len(sampleTieds()))
	assert.Equal(t, 1, tied[tie(wr(duration.Half, 0), wr(duration.Quarter, 0))])

	durations := map[duration.Duration]bool{dur(6, 16): true}
	assert.True(t, durations[dur(3, 8)], "equal values are the same key")

	positions := map[duration.Position]bool{duration.Origin(): true}
	assert.True(t, positions[pos(0, 5)])
	assert.False(t, positions[duration.Position{}])

	// A structure of a consumer that holds the values stays comparable.
	type unit struct {
		W duration.Written
		T duration.Tied
		D duration.Duration
		P duration.Position
		B duration.Base
	}

	assert.True(t, reflect.TypeFor[unit]().Comparable())
}

// TestWrittenEqualCompare covers business-spec 2.1, KP-5, KP-6, KR-D6, and
// architecture 4.5: Equal by record matches ==; Compare is a total order by
// value, then the number of tuplets, the base (longer first), the dots, and
// the tuplets; EqualDuration and CompareDuration compare values.
func TestWrittenEqualCompare(t *testing.T) {
	t.Parallel()

	values := sampleWrittens()

	for _, a := range values {
		for _, b := range values {
			assert.Equal(t, a == b, a.Equal(b), "%s Equal %s", a, b)

			if !a.IsSet() || !b.IsSet() {
				continue
			}

			ra, rb := ratOfDuration(a.MustDuration()), ratOfDuration(b.MustDuration())
			assert.Equal(t, ra.Cmp(rb) == 0, a.MustEqualDuration(b))
			assert.Equal(t, ra.Cmp(rb), a.MustCompareDuration(b))

			if ra.Cmp(rb) != 0 {
				assert.Equal(t, ra.Cmp(rb), a.Compare(b), "Compare by value first: %s, %s", a, b)
			}
		}
	}

	checkOrder(t, values, duration.Written.Compare)

	triplet := tuplet.Triplet()

	// KP-6.
	q32, q64 := wr(duration.Quarter, 0, triplet), wr(duration.Quarter, 0, tp(6, 4))
	assert.False(t, q32.Equal(q64))
	assert.True(t, q32.MustEqualDuration(q64))

	nested1 := wr(duration.Quarter, 0, triplet, tuplet.Quintuplet())
	nested2 := wr(duration.Quarter, 0, tuplet.Quintuplet(), triplet)
	assert.False(t, nested1.Equal(nested2))
	assert.True(t, nested1.MustEqualDuration(nested2))

	// KR-D6 at equal values.
	for _, c := range []struct {
		name          string
		first, second duration.Written
	}{
		{"fewer tuplets first", wr(duration.Quarter, 0), wr(duration.Half, 0, tp(2, 1))},
		{"longer base first", wr(duration.Half, 0, tp(2, 1)), wr(duration.Quarter, 0, tp(1, 1))},
		{"fewer dots first", wr(duration.Quarter, 0, tp(1, 1)), wr(duration.Quarter, 1, triplet)},
		{"tuplets by tuplet.Compare", q32, q64},
		{"tuplets in order", nested1, nested2},
	} {
		require.True(t, c.first.MustEqualDuration(c.second), c.name)
		assert.Equal(t, -1, c.first.Compare(c.second), c.name)
		assert.Equal(t, 1, c.second.Compare(c.first), c.name)
	}

	assert.Equal(t, triplet.Compare(tuplet.Quintuplet()), nested1.Compare(nested2), "the first differing tuplet decides")
}

// TestTiedEqualCompare covers business-spec 2.1, KR-D6, and architecture
// 4.7: Equal by parts in order matches ==; Compare by value, then the number
// of parts, then the parts.
func TestTiedEqualCompare(t *testing.T) {
	t.Parallel()

	values := sampleTieds()

	for _, a := range values {
		for _, b := range values {
			assert.Equal(t, a == b, a.Equal(b), "%s Equal %s", a, b)

			if !a.IsSet() || !b.IsSet() {
				continue
			}

			ra, rb := ratOfDuration(a.MustDuration()), ratOfDuration(b.MustDuration())
			assert.Equal(t, ra.Cmp(rb) == 0, a.MustEqualDuration(b))
			assert.Equal(t, ra.Cmp(rb), a.MustCompareDuration(b))

			if ra.Cmp(rb) != 0 {
				assert.Equal(t, ra.Cmp(rb), a.Compare(b))
			}
		}
	}

	checkOrder(t, values, duration.Tied.Compare)

	h, q := wr(duration.Half, 0), wr(duration.Quarter, 0)

	assert.Equal(t, -1, tie(h, q).Compare(tie(q, q, q)), "fewer parts first at equal value")
	assert.Equal(t, 1, tie(h, q).Compare(tie(q, h)), "then parts by Written.Compare: the quarter precedes the half")
	assert.True(t, tie(h, q).MustEqualDuration(tie(q, q, q)))
}

// TestDurationPositionEqual covers architecture 4.4 and 4.9: Equal by value
// matches == for all values, including unset ones.
func TestDurationPositionEqual(t *testing.T) {
	t.Parallel()

	durations := []duration.Duration{{}, dur(1, 4), dur(2, 8), dur(3, 8), durRat(new(big.Rat).SetFrac64(1, 3))}
	for _, a := range durations {
		for _, b := range durations {
			assert.Equal(t, a == b, a.Equal(b))
		}
	}

	checkOrder(t, durations, duration.Duration.Compare)

	positions := []duration.Position{{}, duration.Origin(), pos(0, 3), pos(1, 4), pos(2, 8), pos(5, 1)}
	for _, a := range positions {
		for _, b := range positions {
			assert.Equal(t, a == b, a.Equal(b))
		}
	}

	checkOrder(t, positions, duration.Position.Compare)
}
