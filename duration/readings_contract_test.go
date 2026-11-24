package duration_test

import (
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fiveRatios are the five tuplet ratios of the search-bound cases.
func fiveRatios() []tuplet.Tuplet {
	return []tuplet.Tuplet{tuplet.Triplet(), tuplet.Quintuplet(), tuplet.Sextuplet(), tp(7, 4), tp(9, 8)}
}

// readingKeys counts the readings by their rendered items.
func readingKeys(readings []duration.Reading) map[string]int {
	keys := make(map[string]int, len(readings))
	for _, r := range readings {
		keys[describeContents(r.MustContents())]++
	}

	return keys
}

// Every call of Readings finishes quickly: with readings, with an honestly
// empty result, or with ErrSearchLimit, never exhausting time or memory. A
// reading has exactly len(ds) written values, so an input longer than
// MaxParts has no readings, and the call returns at once.
func TestReadingsSearchBound(t *testing.T) {
	t.Parallel()

	dense := duration.ReadingsOptions{
		Allowed:      fiveRatios(),
		MaxDepth:     1,
		MaxReadings:  4,
		FreeContents: true,
	}

	tests := []struct {
		name      string
		ds        []duration.Duration
		opts      duration.ReadingsOptions
		limit     time.Duration
		wantEmpty bool
	}{
		{
			name:  "dense 1/64 free nested",
			ds:    repeatDur(1000, 1, 64),
			opts:  dense,
			limit: 10 * time.Second,
		},
		{
			name: "quarters and a fifth deep nesting",
			ds:   []duration.Duration{dur(1, 4), dur(1, 4), dur(1, 4), dur(1, 4), dur(1, 5)},
			opts: duration.ReadingsOptions{
				Allowed:  fiveRatios(),
				MaxDepth: 7,
			},
			limit: 10 * time.Second,
		},
		{
			name:      "longer than the default MaxParts",
			ds:        repeatDur(2000, 1, 64),
			opts:      dense,
			limit:     time.Second,
			wantEmpty: true,
		},
		{
			name: "longer than the given MaxParts",
			ds:   repeatDur(4, 1, 4),
			opts: duration.ReadingsOptions{
				Allowed:      fiveRatios(),
				MaxParts:     3,
				FreeContents: true,
			},
			limit:     time.Second,
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := within(t, tt.limit, func() result[[]duration.Reading] {
				readings, err := duration.Readings(tt.ds, tt.opts)

				return result[[]duration.Reading]{readings, err}
			})

			if tt.wantEmpty {
				require.NoError(t, r.err)
				assert.Empty(t, r.v)

				return
			}

			if r.err != nil {
				assertErr(t, r.err, duration.ErrSearchLimit)

				return
			}

			maxReadings := tt.opts.MaxReadings
			if maxReadings == 0 {
				maxReadings = 256
			}

			assert.LessOrEqual(t, len(r.v), maxReadings)

			for _, reading := range r.v {
				assertExactReading(t, tt.ds, reading)
			}
		})
	}
}

// An input exactly as long as MaxParts still has its readings.
func TestReadingsMaxPartsBoundary(t *testing.T) {
	t.Parallel()

	ds := repeatDur(4, 1, 4)
	readings, err := duration.Readings(ds, duration.ReadingsOptions{MaxParts: 4})
	require.NoError(t, err)
	require.Len(t, readings, 1)

	quarter := duration.WrittenContent(wr(duration.Quarter, 0))
	assert.True(t, contentsEqual([]duration.Content{quarter, quarter, quarter, quarter}, readings[0].MustContents()),
		"reading %s", describeContents(readings[0].MustContents()))
}

// Allowed is a set by Tuplet.Equal: a repeated ratio gives the same readings
// as the ratio once, without duplicate readings.
func TestReadingsAllowedIsASet(t *testing.T) {
	t.Parallel()

	triplet := tuplet.Triplet()
	sextuplet := tuplet.Sextuplet()

	tests := []struct {
		name      string
		ds        []duration.Duration
		repeated  []tuplet.Tuplet
		reference []tuplet.Tuplet
		free      bool
		wantLen   int // 0 means not fixed by the test
	}{
		{"three 1/12, 3:2 twice", repeatDur(3, 1, 12), []tuplet.Tuplet{triplet, triplet}, []tuplet.Tuplet{triplet}, false, 1},
		{"three 1/12, 3:2 twice, free", repeatDur(3, 1, 12), []tuplet.Tuplet{triplet, triplet}, []tuplet.Tuplet{triplet}, true, 0},
		{
			"six 1/12, 3:2 6:4 3:2", repeatDur(6, 1, 12),
			[]tuplet.Tuplet{triplet, sextuplet, triplet}, []tuplet.Tuplet{triplet, sextuplet}, false, 2,
		},
		{
			"six 1/12, 3:2 6:4 3:2, free", repeatDur(6, 1, 12),
			[]tuplet.Tuplet{triplet, sextuplet, triplet}, []tuplet.Tuplet{triplet, sextuplet}, true, 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := duration.Readings(tt.ds, duration.ReadingsOptions{
				Allowed: tt.repeated, MaxReadings: manyReadings, FreeContents: tt.free,
			})
			require.NoError(t, err)

			want, err := duration.Readings(tt.ds, duration.ReadingsOptions{
				Allowed: tt.reference, MaxReadings: manyReadings, FreeContents: tt.free,
			})
			require.NoError(t, err)
			require.NotEmpty(t, want)

			for i := range got {
				for j := i + 1; j < len(got); j++ {
					assert.False(t, got[i].Equal(got[j]),
						"readings %d and %d are equal: %s", i, j, describeContents(got[i].MustContents()))
				}
			}

			assert.Equal(t, readingKeys(want), readingKeys(got), "readings:\n%s", describeReadings(got))

			if tt.wantLen > 0 {
				assert.Len(t, got, tt.wantLen)
			}
		})
	}
}

// findReading returns the first reading satisfying ok.
func findReading(t *testing.T, readings []duration.Reading, ok func(items []duration.Content) bool) duration.Reading {
	t.Helper()

	for _, r := range readings {
		if ok(r.MustContents()) {
			return r
		}
	}

	require.Fail(t, "no matching reading", "readings:\n%s", describeReadings(readings))

	return duration.Reading{}
}

// Reading.Equal compares the items in order: kind, written value, group, and
// length.
func TestReadingEqual(t *testing.T) {
	t.Parallel()

	triplet := tuplet.Triplet()
	none := duration.ReadingsOptions{}
	isGroup := func(items []duration.Content) bool { return len(items) == 1 && items[0].IsGroup() }

	quarter := duration.MustReadings([]duration.Duration{dur(1, 4)}, none)[0]
	quarterAgain := duration.MustReadings([]duration.Duration{dur(1, 4)}, none)[0]
	eighth := duration.MustReadings([]duration.Duration{dur(1, 8)}, none)[0]
	twoQuarters := duration.MustReadings([]duration.Duration{dur(1, 4), dur(1, 4)}, none)[0]
	groupedQuarter := findReading(t, duration.MustReadings([]duration.Duration{dur(1, 4)}, duration.ReadingsOptions{
		Allowed: []tuplet.Tuplet{triplet}, MaxDots: 1, FreeContents: true,
	}), isGroup)
	threeEighths := duration.MustReadings(repeatDur(3, 1, 12), duration.ReadingsOptions{Allowed: []tuplet.Tuplet{triplet}})[0]
	quarterEighth := duration.MustReadings([]duration.Duration{dur(1, 6), dur(1, 12)}, duration.ReadingsOptions{
		Allowed: []tuplet.Tuplet{triplet}, FreeContents: true,
	})[0]

	tests := []struct {
		name string
		a, b duration.Reading
		want bool
	}{
		{"equal readings of separate calls", quarter, quarterAgain, true},
		{"a reading equals itself", threeEighths, threeEighths, true},
		{"written value and group at one position", quarter, groupedQuarter, false},
		{"unequal written values", quarter, eighth, false},
		{"unequal groups", threeEighths, quarterEighth, false},
		{"different lengths", quarter, twoQuarters, false},
		{"zero and a reading", duration.Reading{}, quarter, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.a.Equal(tt.b))
			assert.Equal(t, tt.want, tt.b.Equal(tt.a), "symmetric")
		})
	}
}

// Group.Equal compares the tuplet as given, the unit, and the contents in
// order; the unset group equals only the unset group.
func TestGroupEqual(t *testing.T) {
	t.Parallel()

	triplet := tuplet.Triplet()
	sextuplet := tuplet.Sextuplet()
	eighth := wr(duration.Eighth, 0)
	quarterIn3 := wcon(duration.Quarter, 0, triplet)
	eighthIn3 := wcon(duration.Eighth, 0, triplet)

	group := func(t tuplet.Tuplet, unit duration.Written, contents ...duration.Content) duration.Group {
		return duration.MustNewGroup(t, unit, contents...)
	}

	tests := []struct {
		name string
		a, b duration.Group
		want bool
	}{
		{
			"equal groups built separately",
			group(triplet, eighth, quarterIn3, eighthIn3), group(triplet, eighth, quarterIn3, eighthIn3), true,
		},
		{"different tuplets of equal coefficient", group(triplet, eighth), group(sextuplet, eighth), false},
		{"different units", group(triplet, eighth, eighthIn3), group(triplet, wr(duration.Quarter, 0), eighthIn3), false},
		{
			"contents in a different order",
			group(triplet, eighth, quarterIn3, eighthIn3), group(triplet, eighth, eighthIn3, quarterIn3), false,
		},
		{"different numbers of contents", group(triplet, eighth, eighthIn3), group(triplet, eighth, eighthIn3, eighthIn3), false},
		{"unset and set", duration.Group{}, group(triplet, eighth), false},
		{"unset and unset", duration.Group{}, duration.Group{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.a.Equal(tt.b))
			assert.Equal(t, tt.want, tt.b.Equal(tt.a), "symmetric")
		})
	}
}

// NewGroup fails with ErrOutOfRange when the span or the sounding sum of the
// contents leaves the range of common/fraction; a group at the edge of the
// range keeps exact spans.
func TestGroupFractionRange(t *testing.T) {
	t.Parallel()

	// widening is 1:MaxInt with coefficient MaxInt. A maxima under four of
	// them is 8·MaxInt^4, just below 2^255 − 1; twice that is not.
	widening := tp(1, math.MaxInt)
	under := func(n int) []tuplet.Tuplet {
		chain := make([]tuplet.Tuplet, n)
		for i := range chain {
			chain[i] = widening
		}

		return chain
	}

	m := big.NewInt(math.MaxInt)
	edge := new(big.Int).Mul(big.NewInt(8), new(big.Int).Exp(m, big.NewInt(4), nil))
	require.Negative(t, edge.Cmp(maxComponent()), "the edge value is in range")

	unit := wr(duration.Maxima, 0, under(3)...)
	content := duration.WrittenContent(wr(duration.Maxima, 0, under(4)...))

	t.Run("span out of range", func(t *testing.T) {
		t.Parallel()

		_, err := duration.NewGroup(widening, wr(duration.Maxima, 0, under(4)...))
		assertErr(t, err, duration.ErrOutOfRange, fraction.ErrOutOfRange)
	})

	t.Run("sounding sum out of range", func(t *testing.T) {
		t.Parallel()

		_, err := duration.NewGroup(widening, unit, content, content)
		assertErr(t, err, duration.ErrOutOfRange, fraction.ErrOutOfRange)
	})

	t.Run("edge of the range", func(t *testing.T) {
		t.Parallel()

		g, err := duration.NewGroup(widening, unit, content)
		require.NoError(t, err)

		edgeRat := new(big.Rat).SetInt(edge)
		assert.Zero(t, ratOf(g.MustSpan()).Cmp(edgeRat))
		assert.Zero(t, ratOf(g.MustShortfall()).Sign())

		spans, err := g.Spans()
		require.NoError(t, err)
		require.Len(t, spans, 1)
		assert.Zero(t, ratOf(spans[0][0]).Sign())
		assert.Zero(t, ratOf(spans[0][1]).Cmp(edgeRat))
	})
}

// Readings rejects a MaxParts above duration.MaxParts and a set Longest that
// is not a base with ErrInvalidArgument; the limits themselves are accepted.
func TestReadingsOptionErrors(t *testing.T) {
	t.Parallel()

	ds := []duration.Duration{dur(1, 4)}

	tests := []struct {
		name    string
		opts    duration.ReadingsOptions
		wantErr error
	}{
		{"MaxParts below zero", duration.ReadingsOptions{MaxParts: -1}, duration.ErrInvalidArgument},
		{"MaxParts above the limit", duration.ReadingsOptions{MaxParts: duration.MaxParts + 1}, duration.ErrInvalidArgument},
		{"MaxParts at the limit", duration.ReadingsOptions{MaxParts: duration.MaxParts}, nil},
		{"Longest past the 1024th", duration.ReadingsOptions{Longest: duration.ThousandTwentyFourth + 1}, duration.ErrInvalidArgument},
		{"Longest far outside the series", duration.ReadingsOptions{Longest: duration.Base(math.MaxUint8)}, duration.ErrInvalidArgument},
		{"Longest maxima", duration.ReadingsOptions{Longest: duration.Maxima}, nil},
		// Longest is checked before Allowed: an invalid Longest together with an
		// unset tuplet of Allowed fails with ErrInvalidArgument alone, without
		// tuplet.ErrUnset.
		{"Longest before an unset tuplet of Allowed", duration.ReadingsOptions{
			Longest: duration.ThousandTwentyFourth + 1, Allowed: []tuplet.Tuplet{{}},
		}, duration.ErrInvalidArgument},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			readings, err := duration.Readings(ds, tt.opts)
			if tt.wantErr != nil {
				assertErr(t, err, tt.wantErr)

				return
			}

			require.NoError(t, err)
			assert.Len(t, readings, 1)
		})
	}
}
