package meter_test

import (
	"fmt"
	"testing"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/meter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// structureSignature describes the results, or the kinds of failure, of the
// operations of structure listed by architecture 4.13 (ИР-6): Duration,
// UnitCount, Units, Unit, HasGrouping, Grouping, Groups, Beats, BeatCount,
// Kind, RussianKind, BeatDivision, StrongBeats, UnitWeights, place by offset
// and start by number, and grouping variants.
func structureSignature(m meter.Meter) []string {
	var out []string

	add := func(name string, v any, err error) { out = append(out, name+": "+describe(v, err)) }
	add2 := func(name string, v, w any, err error) { out = append(out, name+": "+describe([]any{v, w}, err)) }

	size := m.MustDuration()

	add("Duration", size, nil)
	add("UnitCount", noErr(m.UnitCount()), nil)
	add("Units", noErr(m.Units()), nil)

	unit, err := m.Unit()
	add("Unit", unit, err)

	hasGrouping, err := m.HasGrouping()
	add("HasGrouping", hasGrouping, err)

	grouping, err := m.Grouping()
	add("Grouping", grouping, err)

	groups, err := m.Groups()
	add("Groups", groups, err)

	beats, err := m.Beats()
	add("Beats", beats, err)

	beatCount, err := m.BeatCount()
	add("BeatCount", beatCount, err)

	kind, err := m.Kind()
	add("Kind", kind, err)

	russian, err := m.RussianKind()
	add("RussianKind", russian, err)

	div, sub, err := m.BeatDivision()
	add2("BeatDivision", div, sub, err)

	strong, rel, err := m.StrongBeats()
	add2("StrongBeats", strong, rel, err)

	weights, err := m.UnitWeights()
	add("UnitWeights", weights, err)

	variants, err := m.GroupingVariants()
	add("GroupingVariants", variants, err)

	half := posRat(ratOfDuration(size.MustDivInt(2)))
	third := posRat(ratOfDuration(size.MustDivInt(3)))

	for _, p := range []duration.Position{duration.Origin(), third, half, posRat(ratOfDuration(size))} {
		n, off, err := m.UnitAt(p)
		add2("UnitAt "+p.String(), n, off, err)

		n, off, err = m.GroupAt(p)
		add2("GroupAt "+p.String(), n, off, err)

		n, off, err = m.BeatAt(p)
		add2("BeatAt "+p.String(), n, off, err)

		w, err := m.WeightAt(p)
		add("WeightAt "+p.String(), w, err)
	}

	for _, n := range []int{0, 1, 2, 3} {
		s, err := m.UnitStart(n)
		add(fmt.Sprint("UnitStart ", n), s, err)

		s, err = m.GroupStart(n)
		add(fmt.Sprint("GroupStart ", n), s, err)

		s, err = m.BeatStart(n)
		add(fmt.Sprint("BeatStart ", n), s, err)
	}

	return out
}

// TestPropertiesOfStructure covers ИР-6 and ИР-7 on the corpus extended by the
// changes of record: signatures equal by structure have the same results or
// the same kinds of failure of the operations of structure.
func TestPropertiesOfStructure(t *testing.T) {
	t.Parallel()

	var ms []meter.Meter

	for _, m := range corpus() {
		ms = append(ms, m)

		for _, to := range toAll {
			if v, err := to.f(m); err == nil {
				ms = append(ms, v)
			}
		}
	}

	signatures := make([][]string, len(ms))
	for i, m := range ms {
		signatures[i] = structureSignature(m)
	}

	pairs := 0

	for i, a := range ms {
		for j := i + 1; j < len(ms); j++ {
			b := ms[j]

			if !noErr(a.EqualStructure(b)) {
				continue
			}

			pairs++

			require.Equal(t, signatures[i], signatures[j], "ИР-6: %v and %v", a, b)
		}
	}

	assert.Greater(t, pairs, 100, "the corpus has pairs equal by structure")
}

// TestEqualStructureKnownPairs covers ИР-6 with the examples of business-spec
// 2.3: C and 4/4; (2+3)/8, 5/8[2+3], and 2/8+3/8.
func TestEqualStructureKnownPairs(t *testing.T) {
	t.Parallel()

	for _, group := range [][]string{
		{"C", "4/4", "(2+2)/4", "2/4+2/4"},
		{"(2+3)/8", "5/8[2+3]", "2/8+3/8"},
		{"6/8", "(3+3)/8", "3/8+3/8"},
		{"C[3+1]", "4/4[3+1]", "(3+1)/4", "3/4+1/4"},
		{"5/8+2/8", "7/8"},
	} {
		first := mt(group[0])

		for _, text := range group[1:] {
			m := mt(text)
			assert.True(t, noErr(first.EqualStructure(m)), "%v and %v", first, m)
			assert.Equal(t, structureSignature(first), structureSignature(m), "%v and %v", first, m)
		}
	}
}
