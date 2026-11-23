package duration_test

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// freeWorkLimit bounds a Readings call whose work the contract bounds
// independently of the size of the input.
const freeWorkLimit = 10 * time.Second

// manyReadings is a MaxReadings that none of the compared inputs reaches, so
// the readings returned are all readings of the input.
const manyReadings = 4096

// repeatDur returns n copies of the sounding duration num/den.
func repeatDur(n, num, den int) []duration.Duration {
	ds := make([]duration.Duration, n)
	for i := range ds {
		ds[i] = dur(num, den)
	}

	return ds
}

// wcon returns the content of the written duration with base, dots and chain.
func wcon(b duration.Base, dots int, chain ...tuplet.Tuplet) duration.Content {
	return duration.WrittenContent(wr(b, dots, chain...))
}

// gcon returns the content of the group of t in unit with contents.
func gcon(t tuplet.Tuplet, unit duration.Written, contents ...duration.Content) duration.Content {
	return duration.GroupContent(duration.MustNewGroup(t, unit, contents...))
}

// contentsEqual reports whether a and b are equal items in order: written
// values by Written.Equal and groups by Group.Equal.
func contentsEqual(a, b []duration.Content) bool {
	return slices.EqualFunc(a, b, func(x, y duration.Content) bool {
		if x.IsGroup() != y.IsGroup() {
			return false
		}

		if x.IsGroup() {
			return x.MustGroup().Equal(y.MustGroup())
		}

		return x.MustWritten().Equal(y.MustWritten())
	})
}

// hasReading reports whether one of readings has exactly the items want.
func hasReading(readings []duration.Reading, want []duration.Content) bool {
	return slices.ContainsFunc(readings, func(r duration.Reading) bool {
		return contentsEqual(r.MustContents(), want)
	})
}

// describeContents renders items for failure messages: written values by
// their canonical text and groups as tuplet(unit){contents}.
func describeContents(items []duration.Content) string {
	parts := make([]string, len(items))
	for i, item := range items {
		if !item.IsGroup() {
			parts[i] = item.MustWritten().String()

			continue
		}

		g := item.MustGroup()
		parts[i] = g.MustTuplet().String() + "(" + g.MustUnit().String() + "){" +
			describeContents(g.MustContents()) + "}"
	}

	return strings.Join(parts, " ")
}

// describeReadings renders readings for failure messages, one per line.
func describeReadings(readings []duration.Reading) string {
	lines := make([]string, len(readings))
	for i, r := range readings {
		lines[i] = describeContents(r.MustContents())
	}

	return strings.Join(lines, "\n")
}

// assertExactReading checks the contract of a reading of ds: its written
// values, flattened in order, sound exactly ds; every group is filled exactly
// (no shortfall); the unit of a group carries the enclosing chain; and every
// written content and nested unit of a group carries the chain of the group,
// the unit chain extended by the group tuplet (the rule of NewGroup).
func assertExactReading(t *testing.T, ds []duration.Duration, r duration.Reading) bool {
	t.Helper()

	ok := true
	text := describeContents(r.MustContents())

	var sounding []duration.Duration

	var walk func(items []duration.Content, chain []tuplet.Tuplet)
	walk = func(items []duration.Content, chain []tuplet.Tuplet) {
		for _, item := range items {
			if !item.IsGroup() {
				w := item.MustWritten()
				ok = assert.True(t, slices.Equal(chain, w.MustChain()),
					"written %s carries chain %v, want %v; reading %s", w, w.MustChain(), chain, text) && ok
				sounding = append(sounding, w.MustDuration())

				continue
			}

			g := item.MustGroup()
			unit := g.MustUnit()
			ok = assert.True(t, slices.Equal(chain, unit.MustChain()),
				"unit %s carries chain %v, want %v; reading %s", unit, unit.MustChain(), chain, text) && ok
			ok = assert.Zero(t, ratOf(g.MustShortfall()).Sign(),
				"group %s is not filled exactly; reading %s", describeContents([]duration.Content{item}), text) && ok

			walk(g.MustContents(), append(slices.Clone(unit.MustChain()), g.MustTuplet()))
		}
	}
	walk(r.MustContents(), nil)

	ok = assert.True(t, slices.EqualFunc(ds, sounding, duration.Duration.Equal),
		"reading %s sounds %v, want %v", text, sounding, ds) && ok

	return ok
}

// With FreeContents a group may end after any number of consecutive input
// durations whose written values fill it exactly; the input then continues
// with written values outside groups or with further groups.
func TestReadingsFreeContentsCovers(t *testing.T) {
	t.Parallel()

	triplet := tuplet.Triplet()
	sextuplet := tuplet.Sextuplet()
	eighth := wr(duration.Eighth, 0)
	quarter := wr(duration.Quarter, 0)

	eighthIn := func(t tuplet.Tuplet) duration.Content { return wcon(duration.Eighth, 0, t) }
	quarterIn3 := wcon(duration.Quarter, 0, triplet)
	plainQuarter := duration.WrittenContent(quarter)

	tripletOfEighths := gcon(triplet, eighth, eighthIn(triplet), eighthIn(triplet), eighthIn(triplet))
	quarterEighth := gcon(triplet, eighth, quarterIn3, eighthIn(triplet))
	eighthQuarter := gcon(triplet, eighth, eighthIn(triplet), quarterIn3)

	tests := []struct {
		name string
		ds   []duration.Duration
		opts duration.ReadingsOptions
		want [][]duration.Content
	}{
		{
			name: "six 1/12 keep both fixed readings",
			ds:   repeatDur(6, 1, 12),
			opts: duration.ReadingsOptions{Allowed: []tuplet.Tuplet{triplet, sextuplet}},
			want: [][]duration.Content{
				{tripletOfEighths, tripletOfEighths},
				{gcon(sextuplet, eighth,
					eighthIn(sextuplet), eighthIn(sextuplet), eighthIn(sextuplet),
					eighthIn(sextuplet), eighthIn(sextuplet), eighthIn(sextuplet))},
			},
		},
		{
			name: "free group then a plain quarter",
			ds:   []duration.Duration{dur(1, 6), dur(1, 12), dur(1, 4)},
			opts: duration.ReadingsOptions{Allowed: []tuplet.Tuplet{triplet}, MaxDots: 1},
			want: [][]duration.Content{{quarterEighth, plainQuarter}},
		},
		{
			name: "free group then a plain quarter without dots",
			ds:   []duration.Duration{dur(1, 6), dur(1, 12), dur(1, 4)},
			opts: duration.ReadingsOptions{Allowed: []tuplet.Tuplet{triplet}},
			want: [][]duration.Content{{quarterEighth, plainQuarter}},
		},
		{
			name: "two free groups",
			ds:   []duration.Duration{dur(1, 6), dur(1, 12), dur(1, 12), dur(1, 6)},
			opts: duration.ReadingsOptions{Allowed: []tuplet.Tuplet{triplet}},
			want: [][]duration.Content{{quarterEighth, eighthQuarter}},
		},
		{
			name: "plain quarter then a free group",
			ds:   []duration.Duration{dur(1, 4), dur(1, 6), dur(1, 12)},
			opts: duration.ReadingsOptions{Allowed: []tuplet.Tuplet{triplet}},
			want: [][]duration.Content{{plainQuarter, quarterEighth}},
		},
		{
			name: "free group between plain quarters",
			ds:   []duration.Duration{dur(1, 4), dur(1, 6), dur(1, 12), dur(1, 4)},
			opts: duration.ReadingsOptions{Allowed: []tuplet.Tuplet{triplet}},
			want: [][]duration.Content{{plainQuarter, quarterEighth, plainQuarter}},
		},
		{
			name: "free group then a fixed group",
			ds:   []duration.Duration{dur(1, 6), dur(1, 12), dur(1, 12), dur(1, 12), dur(1, 12)},
			opts: duration.ReadingsOptions{Allowed: []tuplet.Tuplet{triplet}},
			want: [][]duration.Content{{quarterEighth, tripletOfEighths}},
		},
		{
			name: "fixed group then a free group",
			ds:   []duration.Duration{dur(1, 12), dur(1, 12), dur(1, 12), dur(1, 6), dur(1, 12)},
			opts: duration.ReadingsOptions{Allowed: []tuplet.Tuplet{triplet}},
			want: [][]duration.Content{{tripletOfEighths, quarterEighth}},
		},
		{
			name: "quarter and eighth",
			ds:   []duration.Duration{dur(1, 6), dur(1, 12)},
			opts: duration.ReadingsOptions{Allowed: []tuplet.Tuplet{triplet}},
			want: [][]duration.Content{{quarterEighth}},
		},
		{
			name: "two eighths and two sixteenths",
			ds:   []duration.Duration{dur(1, 12), dur(1, 12), dur(1, 24), dur(1, 24)},
			opts: duration.ReadingsOptions{Allowed: []tuplet.Tuplet{triplet}},
			want: [][]duration.Content{{gcon(triplet, eighth,
				eighthIn(triplet), eighthIn(triplet),
				wcon(duration.Sixteenth, 0, triplet), wcon(duration.Sixteenth, 0, triplet))}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			opts := tt.opts
			opts.FreeContents = true
			readings, err := duration.Readings(tt.ds, opts)
			require.NoError(t, err)

			for _, want := range tt.want {
				assert.True(t, hasReading(readings, want),
					"missing reading %s among:\n%s", describeContents(want), describeReadings(readings))
			}

			for _, r := range readings {
				assertExactReading(t, tt.ds, r)
			}
		})
	}
}

// Turning FreeContents on only adds readings: every reading returned without
// it is returned with it under the same other options, as long as
// MaxReadings truncates neither result.
func TestReadingsFreeContentsMonotone(t *testing.T) {
	t.Parallel()

	triplet := tuplet.Triplet()
	sextuplet := tuplet.Sextuplet()
	quintuplet := tuplet.Quintuplet()

	tests := []struct {
		name string
		ds   []duration.Duration
		opts duration.ReadingsOptions
	}{
		{
			name: "six 1/12 under 3:2 and 6:4",
			ds:   repeatDur(6, 1, 12),
			opts: duration.ReadingsOptions{Allowed: []tuplet.Tuplet{triplet, sextuplet}},
		},
		{
			name: "triplet then a quarter",
			ds:   append(repeatDur(3, 1, 12), dur(1, 4)),
			opts: duration.ReadingsOptions{Allowed: []tuplet.Tuplet{triplet}},
		},
		{
			name: "quarter then a triplet",
			ds:   append([]duration.Duration{dur(1, 4)}, repeatDur(3, 1, 12)...),
			opts: duration.ReadingsOptions{Allowed: []tuplet.Tuplet{triplet}},
		},
		{
			name: "triplet then a dotted quarter",
			ds:   append(repeatDur(3, 1, 12), dur(3, 8)),
			opts: duration.ReadingsOptions{Allowed: []tuplet.Tuplet{triplet}, MaxDots: 1},
		},
		{
			name: "triplets around a quarter",
			ds:   slices.Concat(repeatDur(3, 1, 12), []duration.Duration{dur(1, 4)}, repeatDur(3, 1, 12)),
			opts: duration.ReadingsOptions{Allowed: []tuplet.Tuplet{triplet, sextuplet}},
		},
		{
			name: "quintuplet then two eighths",
			ds:   append(repeatDur(5, 1, 20), dur(1, 8), dur(1, 8)),
			opts: duration.ReadingsOptions{Allowed: []tuplet.Tuplet{quintuplet}},
		},
		{
			name: "four eighths under 3:2 and 2:2",
			ds:   repeatDur(4, 1, 8),
			opts: duration.ReadingsOptions{Allowed: []tuplet.Tuplet{triplet, tp(2, 2)}},
		},
		{
			name: "six 1/12 nested in 2:2",
			ds:   repeatDur(6, 1, 12),
			opts: duration.ReadingsOptions{Allowed: []tuplet.Tuplet{triplet, tp(2, 2)}, MaxDepth: 1},
		},
		{
			name: "triplet nested in a quintuplet",
			ds:   slices.Concat(repeatDur(2, 1, 10), repeatDur(3, 1, 30), repeatDur(2, 1, 10)),
			opts: duration.ReadingsOptions{Allowed: []tuplet.Tuplet{quintuplet, triplet}, MaxDepth: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			opts := tt.opts
			opts.MaxReadings = manyReadings
			off, err := duration.Readings(tt.ds, opts)
			require.NoError(t, err)
			require.NotEmpty(t, off, "the input has readings without FreeContents")
			require.Less(t, len(off), manyReadings, "MaxReadings truncates the readings without FreeContents")

			opts.FreeContents = true
			on, err := duration.Readings(tt.ds, opts)
			require.NoError(t, err)
			assert.Less(t, len(on), manyReadings, "MaxReadings truncates the readings with FreeContents")

			for _, r := range off {
				want := r.MustContents()
				assert.True(t, hasReading(on, want),
					"reading %s without FreeContents is missing with it; with it:\n%s",
					describeContents(want), describeReadings(on))
			}

			for _, r := range on {
				assertExactReading(t, tt.ds, r)
			}
		})
	}
}

// The work of a call with FreeContents and nesting is bounded: on a long input
// that has readings in triplets, the call returns quickly with readings or
// with ErrTooManyParts, never with an empty result and no error.
func TestReadingsFreeContentsWorkBound(t *testing.T) {
	t.Parallel()

	ds := repeatDur(240, 1, 12)
	opts := duration.ReadingsOptions{
		Allowed:      []tuplet.Tuplet{tuplet.Triplet(), tuplet.Sextuplet()},
		MaxDepth:     1,
		MaxReadings:  4,
		FreeContents: true,
	}

	r := within(t, freeWorkLimit, func() result[[]duration.Reading] {
		readings, err := duration.Readings(ds, opts)

		return result[[]duration.Reading]{readings, err}
	})

	if r.err != nil {
		assertErr(t, r.err, duration.ErrTooManyParts)

		return
	}

	require.NotEmpty(t, r.v, "readings in triplets exist: an empty result without an error is wrong")
	assert.LessOrEqual(t, len(r.v), opts.MaxReadings)

	for _, reading := range r.v {
		assertExactReading(t, ds, reading)
	}
}
