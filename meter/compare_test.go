package meter_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/go-muse/musex/meter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEqualities covers О-18, ИР-7, Т-9, КП-18, КП-42, КП-44: equality by
// record, by structure, and by size.
func TestEqualities(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name                      string
		a, b                      string
		record, structure, ofSize bool
	}{
		{"КП-18 3/4 and 6/8", "3/4", "6/8", false, false, true},
		{"КП-18 C and 4/4", "C", "4/4", false, true, true},
		{"КП-18 (2+3)/8 and 5/8[2+3]", "(2+3)/8", "5/8[2+3]", false, true, true},
		{"КП-18 (2+3)/8 and 2/8+3/8", "(2+3)/8", "2/8+3/8", false, true, true},
		{"КП-18 5/8[2+3] and 2/8+3/8", "5/8[2+3]", "2/8+3/8", false, true, true},
		{"КП-18 5/8 and 5/8[2+3]", "5/8", "5/8[2+3]", false, false, true},
		{"КП-18 5/8[2+3] and 5/8[3+2]", "5/8[2+3]", "5/8[3+2]", false, false, true},
		{"КП-18 КП-42 6/8 and 6/8[3+3]", "6/8", "6/8[3+3]", true, true, true},
		{"КП-42 4/4 and 4/4[2+2]", "4/4", "4/4[2+2]", true, true, true},
		{"КП-18 4/4 and (2+2)/4", "4/4", "(2+2)/4", false, true, true},
		{"КП-18 4/4 and 2/4+2/4", "4/4", "2/4+2/4", false, true, true},
		{"КП-18 (2+2)/4 and 2/4+2/4", "(2+2)/4", "2/4+2/4", false, true, true},
		{"КП-44 C[3+1] and 4/4[3+1]", "C[3+1]", "4/4[3+1]", false, true, true},
		{"C[3+1] and C", "C[3+1]", "C", false, false, true},
		{"C| and 2/2", "C|", "2/2", false, true, true},
		{"C| and 4/4", "C|", "4/4", false, false, true},
		{"5/8 and 5/8", "5/8", "5/8", true, true, true},
		{"КП-10 5/8+2/8 and 7/8", "5/8+2/8", "7/8", false, true, true},
		{"2/4+3/8 and 3/8+2/4", "2/4+3/8", "3/8+2/4", false, false, true},
		{"2/4 and 3/4", "2/4", "3/4", false, false, false},
		{"2/4+2/8 and (2+2)/8", "2/4+2/8", "(2+2)/8", false, false, false},
		{"1/4+1/4 and 2/4", "1/4+1/4", "2/4", false, false, true},
		{"1/4+1/4 and 2/4[1+1]", "1/4+1/4", "2/4[1+1]", false, true, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			a, b := mt(c.a), mt(c.b)

			assert.Equal(t, c.record, a.Equal(b), "Equal")
			assert.Equal(t, c.record, b.Equal(a), "Equal is symmetric")
			assert.Equal(t, c.record, a == b, "== matches Equal")
			assert.Equal(t, c.structure, noErr(a.EqualStructure(b)), "EqualStructure")
			assert.Equal(t, c.structure, noErr(b.EqualStructure(a)), "EqualStructure is symmetric")
			assert.Equal(t, c.ofSize, noErr(a.EqualDuration(b)), "EqualDuration")
			assert.Equal(t, c.ofSize, noErr(b.EqualDuration(a)), "EqualDuration is symmetric")
		})
	}
}

// TestCompareDuration covers О-18 and architecture 4.13.
func TestCompareDuration(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		a, b string
		want int
	}{
		{"2/4", "3/4", -1},
		{"3/4", "6/8", 0},
		{"4/4", "3/4", 1},
		{"C", "4/4", 0},
		{"2/4+3/8", "7/8", 0},
		{"1/" + maxIntText(), "1/1024", -1},
		{"1024/8", "3/1", 1},
	} {
		t.Run(c.a+" vs "+c.b, func(t *testing.T) {
			t.Parallel()

			a, b := mt(c.a), mt(c.b)
			assert.Equal(t, c.want, noErr(a.CompareDuration(b)))
			assert.Equal(t, -c.want, noErr(b.CompareDuration(a)))
		})
	}
}

// TestCompare covers О-18, ИР-7, Т-9, КП-47, КП-49 and architecture 4.13: a
// total order by size and then by the canonical text, consistent with ==.
func TestCompare(t *testing.T) {
	t.Parallel()

	var unset meter.Meter

	for _, c := range []struct {
		name string
		a, b meter.Meter
		want int
	}{
		{"КП-49 2/4 before 3/4", mt("2/4"), mt("3/4"), -1},
		{"КП-49 3/4 before 6/8 by text", mt("3/4"), mt("6/8"), -1},
		{"КП-49 6/8 after 3/4", mt("6/8"), mt("3/4"), 1},
		{"4/4 before C by text", mt("4/4"), mt("C"), -1},
		{"(2+3)/8 before 5/8[2+3] by text", mt("(2+3)/8"), mt("5/8[2+3]"), -1},
		{"equal", mt("6/8"), mt("6/8[3+3]"), 0},
		{"КП-47 unset before set", unset, mt("1/" + maxIntText()), -1},
		{"КП-47 set after unset", mt("C"), unset, 1},
		{"КП-47 two unset", unset, unset, 0},
		{"size before text", mt("7/8"), mt("2/2"), -1},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, c.want, c.a.Compare(c.b))
		})
	}
}

// TestOrderProperties covers ИР-7, КП-49 and architecture 4.13 on the corpus
// with the unset meter: Compare returns −1, 0, or +1, is antisymmetric and
// transitive, gives 0 exactly for ==, orders by size and then by
// strings.Compare of the texts; Equal implies EqualStructure, which implies
// EqualDuration.
func TestOrderProperties(t *testing.T) {
	t.Parallel()

	ms := append(corpus(), meter.Meter{})

	for _, a := range ms {
		for _, b := range ms {
			c := a.Compare(b)
			require.Contains(t, []int{-1, 0, 1}, c, "%v vs %v", a, b)
			assert.Equal(t, -c, b.Compare(a), "antisymmetric: %v vs %v", a, b)
			assert.Equal(t, a == b, c == 0, "0 exactly for ==: %v vs %v", a, b)
			assert.Equal(t, a == b, a.Equal(b), "== matches Equal: %v vs %v", a, b)

			if !a.IsSet() || !b.IsSet() {
				assert.Equal(t, a.IsSet() == b.IsSet(), c == 0, "%v vs %v", a, b)

				if a.IsSet() != b.IsSet() {
					assert.Equal(t, map[bool]int{true: 1, false: -1}[a.IsSet()], c, "the unset meter is first")
				}

				continue
			}

			bySize := noErr(a.CompareDuration(b))
			if bySize != 0 {
				assert.Equal(t, bySize, c, "by size: %v vs %v", a, b)
			} else {
				assert.Equal(t, strings.Compare(a.String(), b.String()), c, "by text: %v vs %v", a, b)
			}

			if a.Equal(b) {
				assert.True(t, noErr(a.EqualStructure(b)), "ИР-7: %v vs %v", a, b)
			}

			if noErr(a.EqualStructure(b)) {
				assert.True(t, noErr(a.EqualDuration(b)), "ИР-7: %v vs %v", a, b)
			}

			assert.Equal(t, bySize == 0, noErr(a.EqualDuration(b)))
		}
	}

	sorted := slices.Clone(ms)
	slices.SortFunc(sorted, meter.Meter.Compare)

	for i := 1; i < len(sorted); i++ {
		assert.LessOrEqual(t, sorted[i-1].Compare(sorted[i]), 0, "sorted: %v, %v", sorted[i-1], sorted[i])
	}

	for i := 0; i+2 < len(sorted); i++ {
		assert.LessOrEqual(t, sorted[i].Compare(sorted[i+2]), 0, "transitive: %v, %v", sorted[i], sorted[i+2])
	}
}

// TestSort covers КП-49 and the use with slices.SortFunc of architecture 4.13.
func TestSort(t *testing.T) {
	t.Parallel()

	ms := mts("6/8", "C", "3/4", "2/4", "4/4", "2/4+3/8", "(2+3)/8")
	ms = append(ms, meter.Meter{})
	slices.SortFunc(ms, meter.Meter.Compare)

	want := append([]meter.Meter{{}}, mts("2/4", "(2+3)/8", "3/4", "6/8", "2/4+3/8", "4/4", "C")...)
	assert.Equal(t, want, ms)
}
