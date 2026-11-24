package duration_test

import (
	"testing"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// groupCount returns the number of groups among items, nested groups
// included.
func groupCount(items []duration.Content) int {
	n := 0

	for _, item := range items {
		if item.IsGroup() {
			n += 1 + groupCount(item.MustGroup().MustContents())
		}
	}

	return n
}

// assertReadingsInOrder checks that readings have exactly the items want, in
// that order.
func assertReadingsInOrder(t *testing.T, want [][]duration.Content, readings []duration.Reading) {
	t.Helper()

	got := make([]string, len(readings))
	for i, r := range readings {
		got[i] = describeContents(r.MustContents())
	}

	expected := make([]string, len(want))
	for i, w := range want {
		expected[i] = describeContents(w)
	}

	if !assert.Equal(t, expected, got) {
		return
	}

	for i, r := range readings {
		assert.True(t, contentsEqual(want[i], r.MustContents()), "reading %d: %s", i, got[i])
	}
}

// The readings are ordered by the number of groups, nested groups included,
// then lexicographically by item kind (written values before groups), ratio
// (tuplet.Tuplet.Compare: 3:2 before 6:4 before 2:2), unit, and contents;
// MaxReadings keeps the first ones in that order.
func TestReadingsExactOrder(t *testing.T) {
	t.Parallel()

	triplet := tuplet.Triplet()
	sextuplet := tuplet.Sextuplet()
	neutral := tp(2, 2)
	eighth := wr(duration.Eighth, 0)
	quarter := wr(duration.Quarter, 0)
	sixteenth := wr(duration.Sixteenth, 0)

	eighths := func(n int, chain ...tuplet.Tuplet) []duration.Content {
		contents := make([]duration.Content, n)
		for i := range contents {
			contents[i] = wcon(duration.Eighth, 0, chain...)
		}

		return contents
	}

	tripletOfEighths := gcon(triplet, eighth, eighths(3, triplet)...)
	sextupletOfEighths := gcon(sextuplet, eighth, eighths(6, sextuplet)...)
	plainQuarter := duration.WrittenContent(quarter)

	tests := []struct {
		name string
		ds   []duration.Duration
		opts duration.ReadingsOptions
		want [][]duration.Content
	}{
		{
			// One group, then two groups: a plain value at the start does not
			// change the count of the groups after it.
			name: "a quarter then six 1/12",
			ds:   append([]duration.Duration{dur(1, 4)}, repeatDur(6, 1, 12)...),
			opts: duration.ReadingsOptions{Allowed: []tuplet.Tuplet{triplet, sextuplet}},
			want: [][]duration.Content{
				{plainQuarter, sextupletOfEighths},
				{plainQuarter, tripletOfEighths, tripletOfEighths},
			},
		},
		{
			// Two, three, and four groups: nested groups count.
			name: "six 1/12 nested under 3:2 and 2:2",
			ds:   repeatDur(6, 1, 12),
			opts: duration.ReadingsOptions{Allowed: []tuplet.Tuplet{triplet, neutral}, MaxDepth: 1},
			want: [][]duration.Content{
				{tripletOfEighths, tripletOfEighths},
				{func() duration.Content {
					inner := gcon(triplet, wr(duration.Eighth, 0, neutral), eighths(3, neutral, triplet)...)

					return gcon(neutral, quarter, inner, inner)
				}()},
				{func() duration.Content {
					inner := gcon(neutral, wr(duration.Eighth, 0, triplet), eighths(2, triplet, neutral)...)

					return gcon(triplet, quarter, inner, inner, inner)
				}()},
			},
		},
		{
			// One group in 3:2 before one in 6:4, then the first of the
			// two-group readings; MaxReadings keeps the first three.
			name: "six 1/12 with free contents truncated to three",
			ds:   repeatDur(6, 1, 12),
			opts: duration.ReadingsOptions{
				Allowed: []tuplet.Tuplet{triplet, sextuplet}, FreeContents: true, MaxReadings: 3,
			},
			want: [][]duration.Content{
				{gcon(triplet, quarter, eighths(6, triplet)...)},
				{sextupletOfEighths},
				{tripletOfEighths, tripletOfEighths},
			},
		},
		{
			// All six free readings: equal counts compare by the ratio of the
			// first group, then by the second.
			name: "six 1/12 with free contents, all readings",
			ds:   repeatDur(6, 1, 12),
			opts: duration.ReadingsOptions{
				Allowed: []tuplet.Tuplet{triplet, sextuplet}, FreeContents: true, MaxReadings: manyReadings,
			},
			want: func() [][]duration.Content {
				sextupletHalf := gcon(sextuplet, sixteenth, eighths(3, sextuplet)...)

				return [][]duration.Content{
					{gcon(triplet, quarter, eighths(6, triplet)...)},
					{sextupletOfEighths},
					{tripletOfEighths, tripletOfEighths},
					{tripletOfEighths, sextupletHalf},
					{sextupletHalf, tripletOfEighths},
					{sextupletHalf, sextupletHalf},
				}
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			readings, err := duration.Readings(tt.ds, tt.opts)
			require.NoError(t, err)
			assertReadingsInOrder(t, tt.want, readings)
		})
	}
}

// For every k less than the number of readings, MaxReadings k returns the
// first k readings of the full list, and the full list never has fewer groups
// after more, nested groups included.
func TestReadingsTruncationIsAPrefix(t *testing.T) {
	t.Parallel()

	triplet := tuplet.Triplet()
	sextuplet := tuplet.Sextuplet()
	quintuplet := tuplet.Quintuplet()
	neutral := tp(2, 2)

	tests := []struct {
		name string
		ds   []duration.Duration
		opts duration.ReadingsOptions
	}{
		{
			name: "a quarter then six 1/12",
			ds:   append([]duration.Duration{dur(1, 4)}, repeatDur(6, 1, 12)...),
			opts: duration.ReadingsOptions{Allowed: []tuplet.Tuplet{triplet, sextuplet}},
		},
		{
			name: "six 1/12 nested under 3:2 and 2:2",
			ds:   repeatDur(6, 1, 12),
			opts: duration.ReadingsOptions{Allowed: []tuplet.Tuplet{triplet, neutral}, MaxDepth: 1},
		},
		{
			name: "six 1/12 with free contents",
			ds:   repeatDur(6, 1, 12),
			opts: duration.ReadingsOptions{Allowed: []tuplet.Tuplet{triplet, sextuplet}, FreeContents: true},
		},
		{
			name: "four eighths under 3:2 and 2:2",
			ds:   repeatDur(4, 1, 8),
			opts: duration.ReadingsOptions{Allowed: []tuplet.Tuplet{triplet, neutral}},
		},
		{
			name: "plain quarters around a free group",
			ds:   []duration.Duration{dur(1, 4), dur(1, 6), dur(1, 12), dur(1, 4)},
			opts: duration.ReadingsOptions{Allowed: []tuplet.Tuplet{triplet}, MaxDots: 1, FreeContents: true},
		},
		{
			name: "a quarter then three 1/12 nested and free",
			ds:   append([]duration.Duration{dur(1, 4)}, repeatDur(3, 1, 12)...),
			opts: duration.ReadingsOptions{
				Allowed: []tuplet.Tuplet{triplet, neutral}, MaxDepth: 1, FreeContents: true,
			},
		},
		{
			name: "four eighths nested under 3:2 and 2:2",
			ds:   repeatDur(4, 1, 8),
			opts: duration.ReadingsOptions{Allowed: []tuplet.Tuplet{triplet, neutral}, MaxDepth: 1},
		},
		{
			name: "two eighths and five 1/20 nested",
			ds:   append([]duration.Duration{dur(1, 8), dur(1, 8)}, repeatDur(5, 1, 20)...),
			opts: duration.ReadingsOptions{Allowed: []tuplet.Tuplet{quintuplet, neutral}, MaxDepth: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			opts := tt.opts
			opts.MaxReadings = manyReadings
			full, err := duration.Readings(tt.ds, opts)
			require.NoError(t, err)
			require.GreaterOrEqual(t, len(full), 2, "the input has several readings")
			require.Less(t, len(full), manyReadings, "the full list is not truncated")

			for i := 1; i < len(full); i++ {
				prev, next := full[i-1].MustContents(), full[i].MustContents()
				assert.LessOrEqual(t, groupCount(prev), groupCount(next),
					"reading %d (%s) has more groups than reading %d (%s)",
					i-1, describeContents(prev), i, describeContents(next))
			}

			for k := 1; k < len(full); k++ {
				opts.MaxReadings = k
				first, err := duration.Readings(tt.ds, opts)
				require.NoError(t, err)

				if !assert.Len(t, first, k, "MaxReadings %d", k) {
					continue
				}

				for i := range first {
					assert.True(t, first[i].Equal(full[i]),
						"MaxReadings %d: reading %d is %s, want %s", k, i,
						describeContents(first[i].MustContents()), describeContents(full[i].MustContents()))
				}
			}
		})
	}
}
