package naming_test

import (
	"iter"
	"maps"
	"math"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/go-muse/musex/octave/naming"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seqPair is one pair "ASPN number - name" yielded by Catalog.All.
type seqPair struct {
	aspn int
	name string
}

// collectPairs traverses seq to the end with a range loop and returns the
// pairs in the order they were yielded. The result is never nil.
func collectPairs(seq iter.Seq2[int, string]) []seqPair {
	pairs := []seqPair{}
	for aspn, name := range seq {
		pairs = append(pairs, seqPair{aspn: aspn, name: name})
	}

	return pairs
}

// sortedPairs returns the pairs of an oracle map with a non-empty name in
// ascending order of the number. It does not depend on the package under
// test. The result is never nil.
func sortedPairs(oracle map[int]string) []seqPair {
	pairs := []seqPair{}
	for _, aspn := range slices.Sorted(maps.Keys(oracle)) {
		if name := oracle[aspn]; name != "" {
			pairs = append(pairs, seqPair{aspn: aspn, name: name})
		}
	}

	return pairs
}

// extremePairs returns the pairs of the K-23 catalog with the extreme numbers
// of int (int is 64-bit).
func extremePairs() map[int]string {
	return map[int]string{math.MinInt: "lo", 0: "z", math.MaxInt: "hi"}
}

// russianPrefix returns the first three pairs of the Russian catalog (K-25).
func russianPrefix() []seqPair {
	return []seqPair{
		{aspn: -1, name: "субсубконтроктава"},
		{aspn: 0, name: "субконтроктава"},
		{aspn: 1, name: "контроктава"},
	}
}

// enumeratedCatalogs returns the catalogs of K-24 with their expected
// enumeration: the built-in catalogs and the custom catalogs of K-22, K-23 and
// K-7. Every call makes new catalogs.
func enumeratedCatalogs() []struct {
	namedCatalog
	want []seqPair
} {
	type enumerated = struct {
		namedCatalog
		want []seqPair
	}

	must := func(pairs func() map[int]string) func() *naming.Catalog {
		return func() *naming.Catalog { return naming.MustNew(pairs()) }
	}
	exact := func() map[int]string { return map[int]string{3: "малая октава", 5: " x "} }

	cases := make([]enumerated, 0, 8)
	for _, b := range builtins() {
		cases = append(cases, enumerated{namedCatalog{name: b.name, catalog: b.catalog}, sortedPairs(b.oracle())})
	}

	return append(cases,
		enumerated{namedCatalog{name: "custom K-7", catalog: must(customPairs)}, []seqPair{
			{aspn: -3, name: "deep"}, {aspn: 4, name: "middle"}, {aspn: 12, name: "high"},
		}},
		enumerated{namedCatalog{name: "one name and one empty name", catalog: must(func() map[int]string {
			return map[int]string{1: "a", 2: ""}
		})}, []seqPair{{aspn: 1, name: "a"}}},
		enumerated{namedCatalog{name: "extreme numbers", catalog: must(extremePairs)}, []seqPair{
			{aspn: math.MinInt, name: "lo"}, {aspn: 0, name: "z"}, {aspn: math.MaxInt, name: "hi"},
		}},
		enumerated{namedCatalog{name: "exact strings K-7", catalog: must(exact)}, []seqPair{
			{aspn: 3, name: "малая октава"}, {aspn: 5, name: " x "},
		}},
		enumerated{namedCatalog{name: "spread numbers", catalog: must(func() map[int]string {
			return map[int]string{100: "c", -100: "a", 7: "b2", 0: "b0", -1: "a2", math.MaxInt - 1: "d", math.MinInt + 1: "z"}
		})}, []seqPair{
			{aspn: math.MinInt + 1, name: "z"}, {aspn: -100, name: "a"}, {aspn: -1, name: "a2"},
			{aspn: 0, name: "b0"}, {aspn: 7, name: "b2"}, {aspn: 100, name: "c"}, {aspn: math.MaxInt - 1, name: "d"},
		}},
	)
}

// K-21: each built-in catalog yields exactly the pairs of K-1...K-3 in
// ascending order of the number: Russian 11 pairs -1...9, English and German
// 8 pairs 0...7.
func TestCatalogAllBuiltin(t *testing.T) {
	t.Parallel()

	cases := []struct {
		builtin

		count       int
		first, last int
	}{
		{builtin: builtins()[0], count: 11, first: -1, last: 9},
		{builtin: builtins()[1], count: 8, first: 0, last: 7},
		{builtin: builtins()[2], count: 8, first: 0, last: 7},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := collectPairs(tc.catalog().All())

			assert.Equal(t, sortedPairs(tc.oracle()), got)
			require.Len(t, got, tc.count)
			assert.Equal(t, tc.first, got[0].aspn)
			assert.Equal(t, tc.last, got[len(got)-1].aspn)
		})
	}
}

// K-22, K-23: custom catalogs yield their named octaves in ascending order of
// the number; pairs with the empty name given to New are left out; the
// extreme numbers of int are yielded in their order.
func TestCatalogAllCustom(t *testing.T) {
	t.Parallel()

	for _, tc := range enumeratedCatalogs()[len(builtins()):] {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, collectPairs(tc.catalog().All()))
		})
	}
}

// K-23, guard of NF-7: the traversal of the catalog {math.MinInt, 0,
// math.MaxInt} ends. An implementation that steps through the numbers one by
// one would not end; the deadline turns that into a test failure instead of a
// test binary timeout.
func TestCatalogAllExtremeNumbersEnd(t *testing.T) {
	t.Parallel()

	const deadline = 10 * time.Second

	c := naming.MustNew(extremePairs())
	done := make(chan []seqPair, 1)
	go func() {
		done <- collectPairs(c.All())
	}()

	select {
	case got := <-done:
		assert.Equal(t, []seqPair{
			{aspn: math.MinInt, name: "lo"}, {aspn: 0, name: "z"}, {aspn: math.MaxInt, name: "hi"},
		}, got)
	case <-time.After(deadline):
		require.FailNow(t, "the traversal of {math.MinInt, 0, math.MaxInt} did not end", "deadline %v", deadline)
	}
}

// K-23: a catalog without names - nil, the zero value, created from no pairs
// or from pairs with the empty name only - gives a non-nil sequence that
// yields no pairs: yield is never called, and nothing panics.
func TestCatalogAllWithoutNames(t *testing.T) {
	t.Parallel()

	newCatalog := func(pairs map[int]string) func() *naming.Catalog {
		return func() *naming.Catalog { return naming.MustNew(pairs) }
	}

	cases := append(zeroReceivers(),
		namedCatalog{name: "New(nil)", catalog: newCatalog(nil)},
		namedCatalog{name: "New(empty map)", catalog: newCatalog(map[int]string{})},
		namedCatalog{name: "New(empty names only)", catalog: newCatalog(map[int]string{1: "", 2: ""})},
	)

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := tc.catalog()

			var seq iter.Seq2[int, string]
			require.NotPanics(t, func() { seq = c.All() })
			require.NotNil(t, seq)

			calls := 0
			require.NotPanics(t, func() {
				seq(func(int, string) bool {
					calls++

					return true
				})
			})
			assert.Zero(t, calls, "calls of yield")

			var got []seqPair
			require.NotPanics(t, func() { got = collectPairs(seq) })
			assert.Empty(t, got)
		})
	}
}

// K-24: every yielded pair agrees with Name and ASPN, the numbers strictly
// ascend and no name is empty; every octave with a name among the numbers of
// K-1...K-4 and -20...20 is yielded.
func TestCatalogAllConsistency(t *testing.T) {
	t.Parallel()

	for _, tc := range enumeratedCatalogs() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := tc.catalog()
			got := collectPairs(c.All())
			require.Equal(t, tc.want, got)

			for i, p := range got {
				assert.NotEmpty(t, p.name, "pair %d", i)
				assert.Equal(t, p.name, c.Name(p.aspn), "Name(%d)", p.aspn)

				aspn, err := c.ASPN(p.name)
				if assert.NoError(t, err, "ASPN(%q)", p.name) {
					assert.Equal(t, p.aspn, aspn, "ASPN(%q)", p.name)
				}

				if i > 0 {
					assert.Less(t, got[i-1].aspn, p.aspn, "pairs %d and %d", i-1, i)
				}
			}

			probe := append(probeNumbers(), -3, 3, 5, 12, math.MinInt+1, math.MaxInt-1, -100, 100)
			for _, p := range tc.want {
				probe = append(probe, p.aspn)
			}
			for _, aspn := range probe {
				if name := c.Name(aspn); name != "" {
					assert.Contains(t, got, seqPair{aspn: aspn, name: name}, "octave %d", aspn)
				}
			}
		})
	}
}

// K-24: two catalogs are equal by Equal exactly when they yield the same
// sequence of pairs. Every pair of the set below, including a catalog with
// itself, is compared in both positions: catalogs of one class are equal and
// yield the same pairs; catalogs of different classes are not equal and
// yield different pairs.
func TestCatalogAllEqualIffSameSequence(t *testing.T) {
	t.Parallel()

	must := func(pairs map[int]string) func() *naming.Catalog {
		return func() *naming.Catalog { return naming.MustNew(pairs) }
	}

	withEmpty := customPairs()
	withEmpty[0] = ""
	withEmpty[5] = ""
	otherName := customPairs()
	otherName[12] = "higher"
	otherNumber := map[int]string{4: "middle", -3: "deep", 13: "high"}

	makers := []equalityClass{
		{namedCatalog{name: "custom", catalog: must(customPairs())}, 1},
		{namedCatalog{name: "custom, other fill order", catalog: must(pairsInOrder(
			[]int{12, 4, -3}, []string{"high", "middle", "deep"}))}, 1},
		{namedCatalog{name: "custom with empty names", catalog: must(withEmpty)}, 1},
		{namedCatalog{name: "custom, other name", catalog: must(otherName)}, 2},
		{namedCatalog{name: "custom, other number", catalog: must(otherNumber)}, 3},
		{namedCatalog{name: "Russian", catalog: naming.Russian}, 4},
		{namedCatalog{name: "Russian from pairs", catalog: must(russianOracle())}, 4},
		{namedCatalog{name: "English", catalog: naming.English}, 5},
		{namedCatalog{name: "German", catalog: naming.German}, 6},
		{namedCatalog{name: "German from pairs", catalog: must(germanOracle())}, 6},
		{namedCatalog{name: "German from pairs with empty names", catalog: must(func() map[int]string {
			pairs := germanOracle()
			pairs[-1] = ""
			pairs[8] = ""

			return pairs
		}())}, 6},
		{namedCatalog{name: "nil", catalog: func() *naming.Catalog { return nil }}, 7},
		{namedCatalog{name: "zero value", catalog: func() *naming.Catalog { return &naming.Catalog{} }}, 7},
		{namedCatalog{name: "empty names only", catalog: must(map[int]string{1: "", 2: ""})}, 7},
	}

	for _, left := range makers {
		for _, right := range makers {
			t.Run(left.name+" vs "+right.name, func(t *testing.T) {
				t.Parallel()

				l, r := left.catalog(), right.catalog()
				same := left.class == right.class

				assert.Equal(t, same, l.Equal(r), "Equal")
				assert.Equal(t, same, slices.Equal(collectPairs(l.All()), collectPairs(r.All())), "same pairs")
			})
		}
	}
}

// K-25: a traversal stopped after k pairs gives exactly the first k pairs of
// the full enumeration, and yield is not called again. The sequence is
// called directly with a counting yield and through a range loop with break.
func TestCatalogAllEarlyStop(t *testing.T) {
	t.Parallel()

	t.Run("break after three pairs", func(t *testing.T) {
		t.Parallel()

		var got []seqPair
		require.NotPanics(t, func() {
			for aspn, name := range naming.Russian().All() {
				got = append(got, seqPair{aspn: aspn, name: name})
				if len(got) == 3 {
					break
				}
			}
		})
		assert.Equal(t, russianPrefix(), got)
	})

	full := sortedPairs(russianOracle())
	for k := 1; k <= len(full); k++ {
		t.Run("yield false at pair "+strconv.Itoa(k), func(t *testing.T) {
			t.Parallel()

			var got []seqPair
			calls := 0
			require.NotPanics(t, func() {
				naming.Russian().All()(func(aspn int, name string) bool {
					calls++
					got = append(got, seqPair{aspn: aspn, name: name})

					return calls < k
				})
			})

			assert.Equal(t, k, calls, "calls of yield")
			assert.Equal(t, full[:k], got)
		})
	}

	t.Run("yield false at the third pair", func(t *testing.T) {
		t.Parallel()

		var got []seqPair
		calls := 0
		naming.Russian().All()(func(aspn int, name string) bool {
			calls++
			got = append(got, seqPair{aspn: aspn, name: name})

			return calls != 3
		})

		assert.Equal(t, 3, calls, "calls of yield")
		assert.Equal(t, russianPrefix(), got)
	})

	t.Run("yield true for every pair", func(t *testing.T) {
		t.Parallel()

		calls := 0
		naming.Russian().All()(func(int, string) bool {
			calls++

			return true
		})

		assert.Equal(t, len(full), calls, "calls of yield")
	})
}

// K-25: the same sequence traversed again, also after an early stop, and a
// new call of All give the same pairs.
func TestCatalogAllRepeatable(t *testing.T) {
	t.Parallel()

	for _, tc := range enumeratedCatalogs() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := tc.catalog()
			seq := c.All()

			assert.Equal(t, tc.want, collectPairs(seq), "first traversal")

			for range seq {
				break
			}

			assert.Equal(t, tc.want, collectPairs(seq), "traversal after an early stop")
			assert.Equal(t, tc.want, collectPairs(c.All()), "new call of All")
		})
	}
}

// K-25: the loop body may call any method of the same catalog, including a
// nested traversal of the same sequence and a new All; the results are the
// same as outside the loop. All and the traversal do not change the catalog.
func TestCatalogAllLoopBody(t *testing.T) {
	t.Parallel()

	for _, tc := range enumeratedCatalogs() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := tc.catalog()
			fresh := tc.catalog()
			seq := c.All()

			var got []seqPair
			for aspn, name := range seq {
				got = append(got, seqPair{aspn: aspn, name: name})

				assert.Equal(t, name, c.Name(aspn), "Name(%d) in the loop", aspn)

				number, err := c.ASPN(name)
				if assert.NoError(t, err, "ASPN(%q) in the loop", name) {
					assert.Equal(t, aspn, number, "ASPN(%q) in the loop", name)
				}

				assert.True(t, c.Equal(fresh), "Equal in the loop")
				assert.Equal(t, tc.want, collectPairs(c.All()), "nested new All")
				assert.Equal(t, tc.want, collectPairs(seq), "nested traversal of the same sequence")
			}

			assert.Equal(t, tc.want, got)
			assert.True(t, c.Equal(fresh), "Equal after the traversal")
			assert.True(t, fresh.Equal(c), "Equal after the traversal")
			for _, p := range tc.want {
				assert.Equal(t, p.name, c.Name(p.aspn), "Name(%d) after the traversal", p.aspn)
			}
		})
	}
}

// A panic of the loop body (the yield function) reaches the caller unchanged;
// the catalog and the sequence remain usable afterward.
func TestCatalogAllBodyPanic(t *testing.T) {
	t.Parallel()

	type bodyPanic struct{ at int }

	t.Run("direct call", func(t *testing.T) {
		t.Parallel()

		c := naming.Russian()
		seq := c.All()

		recovered, panicked := recoverPanic(func() {
			seq(func(aspn int, _ string) bool {
				if aspn == 1 {
					panic(bodyPanic{at: aspn})
				}

				return true
			})
		})
		require.True(t, panicked)
		assert.Equal(t, bodyPanic{at: 1}, recovered)

		assert.Equal(t, sortedPairs(russianOracle()), collectPairs(seq))
		assert.Equal(t, sortedPairs(russianOracle()), collectPairs(c.All()))
		assert.True(t, c.Equal(naming.Russian()))
	})

	t.Run("range loop", func(t *testing.T) {
		t.Parallel()

		c := naming.MustNew(customPairs())
		seq := c.All()

		recovered, panicked := recoverPanic(func() {
			for aspn := range seq {
				if aspn == 4 {
					panic(bodyPanic{at: aspn})
				}
			}
		})
		require.True(t, panicked)
		assert.Equal(t, bodyPanic{at: 4}, recovered)

		assert.Equal(t, sortedPairs(customPairs()), collectPairs(seq))
		assert.True(t, c.Equal(naming.MustNew(customPairs())))
	})
}

// A copy cp := *c yields the same pairs as c.
func TestCatalogAllCopy(t *testing.T) {
	t.Parallel()

	cases := enumeratedCatalogs()
	for _, z := range zeroReceivers()[:2] {
		cases = append(cases, struct {
			namedCatalog
			want []seqPair
		}{z, []seqPair{}})
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := tc.catalog()
			cp := *c

			assert.Equal(t, tc.want, collectPairs(cp.All()))
			assert.Equal(t, collectPairs(c.All()), collectPairs(cp.All()))
		})
	}
}

// KR-NM16: a sequence is bound to the pairs that c holds when All is called.
// Assigning to *c afterward does not change the sequence; a new c.All()
// yields the pairs of the new value; a copy made before the assignment
// yields the old pairs.
func TestCatalogAllBinding(t *testing.T) {
	t.Parallel()

	initial := []struct {
		name    string
		catalog func() *naming.Catalog
		want    []seqPair
	}{
		{name: "Russian", catalog: naming.Russian, want: sortedPairs(russianOracle())},
		{name: "custom", catalog: func() *naming.Catalog { return naming.MustNew(customPairs()) }, want: sortedPairs(customPairs())},
		{name: "zero value", catalog: func() *naming.Catalog { return &naming.Catalog{} }, want: []seqPair{}},
	}

	assignments := []struct {
		name   string
		assign func(c *naming.Catalog)
		want   []seqPair
	}{
		{name: "English", assign: func(c *naming.Catalog) { *c = *naming.English() }, want: sortedPairs(englishOracle())},
		{name: "zero value", assign: func(c *naming.Catalog) { *c = naming.Catalog{} }, want: []seqPair{}},
	}

	for _, in := range initial {
		for _, as := range assignments {
			t.Run(in.name+" then "+as.name, func(t *testing.T) {
				t.Parallel()

				c := in.catalog()
				seq := c.All()
				cp := *c

				as.assign(c)

				assert.Equal(t, in.want, collectPairs(seq), "sequence taken before the assignment")
				assert.Equal(t, as.want, collectPairs(c.All()), "new All after the assignment")
				assert.Equal(t, in.want, collectPairs(cp.All()), "copy made before the assignment")
				assert.Equal(t, in.want, collectPairs(seq), "sequence traversed again")
			})
		}
	}
}
