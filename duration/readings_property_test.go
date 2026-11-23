package duration_test

import (
	"fmt"
	"math/big"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// naiveReadings enumerates the readings of the Readings contract directly,
// building every group with the public NewGroup. It reads the contract as:
//
//   - a reading covers the inputs in order by items: a written value outside
//     groups, or a group of an allowed ratio (Allowed taken as a set);
//   - a written value is one sign with at most maxDots dots and a base not
//     longer than longest, written under the enclosing chain, that sounds
//     exactly its input duration;
//   - a group of ratio t covers consecutive inputs; its span is their sounding
//     sum and its unit is one such sign under the enclosing chain sounding
//     span/t.Normal(); its contents are items under the chain extended by t;
//   - without free contents a group has exactly t.Actual() items, with free
//     contents any number of items, at least one;
//   - a group at the top is at depth 0, a group inside a group at depth d is
//     at depth d+1, and no group is deeper than maxDepth.
type naiveReadings struct {
	allowed []tuplet.Tuplet
	longest duration.Base
	maxDots int
	depth   int
	free    bool
}

// newNaive returns the enumerator of opts with Allowed reduced to a set.
func newNaive(opts duration.ReadingsOptions) naiveReadings {
	var allowed []tuplet.Tuplet
	for _, t := range opts.Allowed {
		if !slices.ContainsFunc(allowed, t.Equal) {
			allowed = append(allowed, t)
		}
	}

	longest := opts.Longest
	if longest == 0 {
		longest = duration.Breve
	}

	return naiveReadings{
		allowed: allowed,
		longest: longest,
		maxDots: opts.MaxDots,
		depth:   opts.MaxDepth,
		free:    opts.FreeContents,
	}
}

// signs returns the written signs under chain that sound exactly v.
func (n naiveReadings) signs(v *big.Rat, chain []tuplet.Tuplet) []duration.Written {
	var out []duration.Written

	for _, b := range duration.Bases() {
		if b < n.longest {
			continue
		}

		for dots := range n.maxDots + 1 {
			w, err := duration.NewWritten(b, dots, chain...)
			if err == nil && ratOfDuration(w.MustDuration()).Cmp(v) == 0 {
				out = append(out, w)
			}
		}
	}

	return out
}

// items returns every sequence of items under chain that covers ds exactly.
// Groups created at this layer are at depth level. A need of -1 allows any
// number of items; otherwise exactly need items are required.
func (n naiveReadings) items(ds []*big.Rat, chain []tuplet.Tuplet, level, need int) [][]duration.Content {
	if len(ds) == 0 {
		if need == -1 || need == 0 {
			return [][]duration.Content{nil}
		}

		return nil
	}

	if need == 0 {
		return nil
	}

	rest := need
	if need > 0 {
		rest = need - 1
	}

	// A first item covers ds[:consumed]; the rest of ds follows it.
	type first struct {
		item     duration.Content
		consumed int
	}

	var firsts []first

	for _, w := range n.signs(ds[0], chain) {
		firsts = append(firsts, first{duration.WrittenContent(w), 1})
	}

	if level <= n.depth {
		for j := 1; j <= len(ds); j++ {
			for _, g := range n.groups(ds[:j], chain, level) {
				firsts = append(firsts, first{duration.GroupContent(g), j})
			}
		}
	}

	var out [][]duration.Content

	for _, f := range firsts {
		for _, tail := range n.items(ds[f.consumed:], chain, level, rest) {
			out = append(out, append([]duration.Content{f.item}, tail...))
		}
	}

	return out
}

// groups returns every group at depth level under chain that covers all of ds.
func (n naiveReadings) groups(ds []*big.Rat, chain []tuplet.Tuplet, level int) []duration.Group {
	span := new(big.Rat)
	for _, d := range ds {
		span.Add(span, d)
	}

	var out []duration.Group

	for _, t := range n.allowed {
		unitValue := new(big.Rat).Quo(span, big.NewRat(int64(t.MustNormal()), 1))
		inner := append(slices.Clone(chain), t)

		need := t.MustActual()
		if n.free {
			need = -1
		} else if need > len(ds) {
			continue
		}

		for _, unit := range n.signs(unitValue, chain) {
			for _, contents := range n.items(ds, inner, level+1, need) {
				g, err := duration.NewGroup(t, unit, contents...)
				if err == nil {
					out = append(out, g)
				}
			}
		}
	}

	return out
}

// readings returns every reading of ds.
func (n naiveReadings) readings(ds []duration.Duration) [][]duration.Content {
	values := make([]*big.Rat, len(ds))
	for i, d := range ds {
		values[i] = ratOfDuration(d)
	}

	return n.items(values, nil, 0, -1)
}

// The set of readings equals the naive enumeration of the contract on small
// inputs, with and without FreeContents, at depth 0 and 1.
func TestReadingsMatchNaiveEnumeration(t *testing.T) {
	t.Parallel()

	// Blocks are short runs that usually have readings on their own: plain
	// values, the contents of a 3:2 group in eighths, and a 5:4 group in
	// sixteenths. Inputs concatenate blocks up to six values, so most samples
	// have readings and many have groups.
	blocks := [][]duration.Duration{
		{dur(1, 4)}, {dur(1, 8)}, {dur(3, 8)}, {dur(1, 12)},
		repeatDur(3, 1, 12),
		{dur(1, 6), dur(1, 12)},
		{dur(1, 12), dur(1, 6)},
		{dur(1, 12), dur(1, 12), dur(1, 24), dur(1, 24)},
		{dur(1, 10), dur(1, 10)},
		repeatDur(5, 1, 20),
	}
	ratioSets := [][]tuplet.Tuplet{
		{tuplet.Triplet()},
		{tuplet.Triplet(), tuplet.Sextuplet()},
		{tuplet.Triplet(), tp(2, 2)},
		{tuplet.Quintuplet(), tuplet.Triplet()},
		{tuplet.Triplet(), tuplet.Sextuplet(), tuplet.Quintuplet()},
	}

	rng := rand.New(rand.NewPCG(141, 150))

	type sample struct {
		ds   []duration.Duration
		opts duration.ReadingsOptions
	}

	var samples []sample

	for range 30 {
		var ds []duration.Duration
		for {
			b := blocks[rng.IntN(len(blocks))]
			if len(ds)+len(b) > 6 {
				break
			}

			ds = append(ds, b...)
		}

		allowed := ratioSets[rng.IntN(len(ratioSets))]
		for _, depth := range []int{0, 1} {
			for _, free := range []bool{false, true} {
				samples = append(samples, sample{ds, duration.ReadingsOptions{
					Allowed:      allowed,
					MaxDots:      rng.IntN(2),
					MaxDepth:     depth,
					MaxReadings:  manyReadings,
					FreeContents: free,
				}})
			}
		}
	}

	// Runs of equal values, where fixed and free groups coincide.
	for _, n := range []int{3, 6} {
		for _, free := range []bool{false, true} {
			samples = append(samples, sample{repeatDur(n, 1, 12), duration.ReadingsOptions{
				Allowed:      []tuplet.Tuplet{tuplet.Triplet(), tuplet.Sextuplet()},
				MaxReadings:  manyReadings,
				FreeContents: free,
			}})
		}
	}

	for i, s := range samples {
		name := fmt.Sprintf("%02d %v %v depth %d dots %d free %t",
			i, s.ds, s.opts.Allowed, s.opts.MaxDepth, s.opts.MaxDots, s.opts.FreeContents)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			want := newNaive(s.opts).readings(s.ds)
			require.Less(t, len(want), manyReadings, "the sample is small")

			got, err := duration.Readings(s.ds, s.opts)
			require.NoError(t, err)

			wantKeys := make(map[string]int, len(want))
			for _, w := range want {
				wantKeys[describeContents(w)]++
			}

			gotKeys := make(map[string]int, len(got))
			for _, r := range got {
				gotKeys[describeContents(r.MustContents())]++
			}

			for key, n := range wantKeys {
				assert.Equal(t, n, gotKeys[key], "reading %s", key)
			}

			for key, n := range gotKeys {
				assert.Equal(t, wantKeys[key], n, "reading %s", key)
			}
		})
	}
}
