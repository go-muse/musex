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
// then lexicographically by item kind (written values before groups), ratio,
// unit, and contents; MaxReadings keeps the first ones in that order. Ratios
// compare by the actual count, then by the normal count: 2:2 before 3:2 before
// 6:4. This is not the order of tuplet.Tuplet.Compare, which puts 3:2 and 6:4
// (coefficient 2/3) before 2:2 (coefficient 1).
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

	// A sounding quarter alone in a group: in 2:2 a quarter over an eighth
	// unit (written 1/4 sounds 1/4), in 3:2 a dotted quarter over an eighth
	// unit (written 3/8 sounds 3/8 times 2/3 = 1/4). Both groups need FreeContents,
	// since they hold one item instead of the actual count.
	neutralQuarter := gcon(neutral, eighth, wcon(duration.Quarter, 0, neutral))
	tripletQuarter := gcon(triplet, eighth, wcon(duration.Quarter, 1, triplet))

	// Readings of two sounding quarters under 2:2 and 3:2 with free contents
	// and one dot, in the order of Readings. Without groups: two quarters.
	// With one group, the written value first comes first, then the group
	// by ratio, then by unit (eighth before quarter): a group over both
	// quarters has a quarter unit, in 2:2 with two quarters and in 3:2 with two
	// dotted quarters. With two groups, one per quarter, by the ratio of the
	// first group, then of the second.
	twoQuartersFree := [][]duration.Content{
		{plainQuarter, plainQuarter},
		{plainQuarter, neutralQuarter},
		{plainQuarter, tripletQuarter},
		{neutralQuarter, plainQuarter},
		{gcon(neutral, quarter, wcon(duration.Quarter, 0, neutral), wcon(duration.Quarter, 0, neutral))},
		{tripletQuarter, plainQuarter},
		{gcon(triplet, quarter, wcon(duration.Quarter, 1, triplet), wcon(duration.Quarter, 1, triplet))},
		{neutralQuarter, neutralQuarter},
		{neutralQuarter, tripletQuarter},
		{tripletQuarter, neutralQuarter},
		{tripletQuarter, tripletQuarter},
	}

	// The cases with 2:2 and 3:2 pin the ratio order only if tuplet.Tuplet.Compare
	// orders this pair the other way: 3:2 (coefficient 2/3) before 2:2
	// (coefficient 1), while Readings puts 2:2 (actual 2) before 3:2 (actual 3).
	require.Negative(t, triplet.Compare(neutral), "Tuplet.Compare puts 3:2 before 2:2")
	require.Positive(t, neutral.Compare(triplet), "Tuplet.Compare puts 2:2 after 3:2")

	// The actual count is the first key: 2:3 (actual 2) precedes 4:2
	// (actual 4), although the normal count alone, 3 against 2, orders the
	// pair the other way.
	twoThree := tp(2, 3)
	fourTwo := tp(4, 2)

	require.Less(t, noErr(twoThree.Actual()), noErr(fourTwo.Actual()), "2:3 has the smaller actual count")
	require.Greater(t, noErr(twoThree.Normal()), noErr(fourTwo.Normal()), "2:3 has the larger normal count")

	// A sounding 3/16 alone in a group: in 2:3 an eighth over a sixteenth
	// unit (written 1/8 sounds 1/8 times 3/2 = 3/16), in 4:2 a dotted quarter
	// over a dotted sixteenth unit (written 3/8 sounds 3/8 times 1/2 = 3/16).
	// Both need FreeContents, since they hold one item instead of the actual
	// count. Outside groups 3/16 is a dotted eighth.
	dottedEighth := wcon(duration.Eighth, 1)
	twoThreeAlone := gcon(twoThree, sixteenth, wcon(duration.Eighth, 0, twoThree))
	fourTwoAlone := gcon(fourTwo, wr(duration.Sixteenth, 1), wcon(duration.Quarter, 1, fourTwo))

	// Readings of two sounding 3/16 under 4:2 and 2:3 with free contents and
	// one dot, in the order of Readings: the same shape as for two quarters
	// above, with 2:3 in place of 2:2 and 4:2 in place of 3:2. A group over
	// both durations has the span 3/8: in 2:3 an eighth unit with two eighths,
	// in 4:2 a dotted eighth unit with two dotted quarters; it comes after the
	// group of the same ratio over one duration, whose unit is shorter.
	twoDottedEighthsFree := [][]duration.Content{
		{dottedEighth, dottedEighth},
		{dottedEighth, twoThreeAlone},
		{dottedEighth, fourTwoAlone},
		{twoThreeAlone, dottedEighth},
		{gcon(twoThree, eighth, wcon(duration.Eighth, 0, twoThree), wcon(duration.Eighth, 0, twoThree))},
		{fourTwoAlone, dottedEighth},
		{gcon(fourTwo, wr(duration.Eighth, 1), wcon(duration.Quarter, 1, fourTwo), wcon(duration.Quarter, 1, fourTwo))},
		{twoThreeAlone, twoThreeAlone},
		{twoThreeAlone, fourTwoAlone},
		{fourTwoAlone, twoThreeAlone},
		{fourTwoAlone, fourTwoAlone},
	}

	// The normal count is the second key: at the equal actual count 3, 3:2
	// precedes 3:4, although its unit for three sounding 1/6 is longer (a
	// quarter against an eighth), so the unit alone would order the pair the
	// other way.
	threeFour := tp(3, 4)

	require.Equal(t, noErr(triplet.Actual()), noErr(threeFour.Actual()), "3:2 and 3:4 have equal actual counts")
	require.Less(t, noErr(triplet.Normal()), noErr(threeFour.Normal()), "3:2 has the smaller normal count")
	require.Negative(t, eighth.Compare(quarter), "an eighth unit precedes a quarter unit")

	// Three sounding 1/6 have no reading outside groups (1/6 is no sign) and,
	// without free contents, exactly one group of three items per ratio: in
	// 3:2 three quarters over a quarter unit (written 1/4 sounds 1/4 times
	// 2/3 = 1/6), in 3:4 three eighths over an eighth unit (written 1/8 sounds
	// 1/8 times 4/3 = 1/6).
	threeSixths := [][]duration.Content{
		{gcon(triplet, quarter, wcon(duration.Quarter, 0, triplet), wcon(duration.Quarter, 0, triplet),
			wcon(duration.Quarter, 0, triplet))},
		{gcon(threeFour, eighth, wcon(duration.Eighth, 0, threeFour), wcon(duration.Eighth, 0, threeFour),
			wcon(duration.Eighth, 0, threeFour))},
	}

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
		{
			// Equal group counts whose first differing items are groups of
			// 2:2 and 3:2: the readings with 2:2 come first, at the second
			// item (one group), at the first item over both quarters (one
			// group, quarter units), and at either item (two groups). Allowed
			// lists 3:2 first; that does not change the order.
			name: "two quarters under 3:2 and 2:2 with free contents, all readings",
			ds:   repeatDur(2, 1, 4),
			opts: duration.ReadingsOptions{
				Allowed: []tuplet.Tuplet{triplet, neutral}, MaxDots: 1, FreeContents: true,
				MaxReadings: manyReadings,
			},
			want: twoQuartersFree,
		},
		{
			// MaxReadings keeps the reading with the 2:2 group and drops the
			// one with the 3:2 group, which follows it.
			name: "two quarters under 3:2 and 2:2 with free contents truncated to two",
			ds:   repeatDur(2, 1, 4),
			opts: duration.ReadingsOptions{
				Allowed: []tuplet.Tuplet{triplet, neutral}, MaxDots: 1, FreeContents: true, MaxReadings: 2,
			},
			want: twoQuartersFree[:2],
		},
		{
			// MaxReadings keeps the one-group reading over both quarters in
			// 2:2 and drops the one in 3:2, both in quarter units.
			name: "two quarters under 3:2 and 2:2 with free contents truncated to five",
			ds:   repeatDur(2, 1, 4),
			opts: duration.ReadingsOptions{
				Allowed: []tuplet.Tuplet{triplet, neutral}, MaxDots: 1, FreeContents: true, MaxReadings: 5,
			},
			want: twoQuartersFree[:5],
		},
		{
			// The actual count decides before the normal count: in every pair
			// with equal group counts whose first differing items are groups
			// of 2:3 and 4:2, the 2:3 reading comes first.
			name: "two 3/16 under 4:2 and 2:3 with free contents, all readings",
			ds:   repeatDur(2, 3, 16),
			opts: duration.ReadingsOptions{
				Allowed: []tuplet.Tuplet{fourTwo, twoThree}, MaxDots: 1, FreeContents: true,
				MaxReadings: manyReadings,
			},
			want: twoDottedEighthsFree,
		},
		{
			// MaxReadings keeps the reading with the 2:3 group and drops the
			// one with the 4:2 group, which follows it.
			name: "two 3/16 under 4:2 and 2:3 with free contents truncated to two",
			ds:   repeatDur(2, 3, 16),
			opts: duration.ReadingsOptions{
				Allowed: []tuplet.Tuplet{fourTwo, twoThree}, MaxDots: 1, FreeContents: true, MaxReadings: 2,
			},
			want: twoDottedEighthsFree[:2],
		},
		{
			// At the equal actual count the normal count decides, before the
			// unit: 3:2 in quarters precedes 3:4 in eighths.
			name: "three 1/6 under 3:4 and 3:2, all readings",
			ds:   repeatDur(3, 1, 6),
			opts: duration.ReadingsOptions{Allowed: []tuplet.Tuplet{threeFour, triplet}, MaxReadings: manyReadings},
			want: threeSixths,
		},
		{
			// MaxReadings keeps the 3:2 reading and drops the 3:4 one.
			name: "three 1/6 under 3:4 and 3:2 truncated to one",
			ds:   repeatDur(3, 1, 6),
			opts: duration.ReadingsOptions{Allowed: []tuplet.Tuplet{threeFour, triplet}, MaxReadings: 1},
			want: threeSixths[:1],
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
