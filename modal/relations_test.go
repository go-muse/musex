package modal_test

import (
	"iter"
	"math"
	"slices"
	"sync"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/modal"
	"github.com/go-muse/musex/mode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The expected relations below come from business-spec 21.2, MOD-6, MOD-7,
// MOD-9, example 3, and the table "additional checks of the new queries";
// the control mode for the failure order comes from architecture MOD-A10.

func relation(from, to, cycle int, value string, category modal.Category) modal.Relation {
	return modal.Relation{From: from, To: to, Cycle: cycle, Interval: interval.MustParse(value), Category: category}
}

func checkRelation(t *testing.T, want, got modal.Relation) {
	t.Helper()
	assert.Equal(t, want.From, got.From, "From")
	assert.Equal(t, want.To, got.To, "To")
	assert.Equal(t, want.Cycle, got.Cycle, "Cycle")
	assert.True(t, want.Interval.Equal(got.Interval), "interval %v, want %v", got.Interval, want.Interval)
	assert.Equal(t, want.Category, got.Category, "Category")
	assert.True(t, want == got, "== agrees with the field comparison")
}

func checkRelations(t *testing.T, want, got []modal.Relation) {
	t.Helper()
	require.Len(t, got, len(want))
	for i := range want {
		checkRelation(t, want[i], got[i])
	}
}

// modalCategories lists every modal sentinel, including the new one.
func modalCategories() []error {
	return []error{modal.ErrUnset, modal.ErrInvalidDegree, modal.ErrUnsupportedStructure, modal.ErrUnsupportedQuality, modal.ErrIncomparable}
}

// representable lists the only promised causes of an unrepresentable
// interval in a relation or comparison query.
func representable() []error {
	return []error{mode.ErrOutOfRange, interval.ErrOutOfRange}
}

// checkModalError checks the promised causes and the absence of every other
// modal category.
func checkModalError(t *testing.T, err error, causes ...error) {
	t.Helper()
	require.Error(t, err)
	for _, cause := range causes {
		assert.ErrorIs(t, err, cause)
	}
	for _, category := range modalCategories() {
		if !slices.Contains(causes, category) {
			assert.NotErrorIs(t, err, category, "only the first failing check is reported")
		}
	}
}

// checkMustPanic checks that a Must form panics with the error of the
// ordinary form.
func checkMustPanic(t *testing.T, want error, call func()) {
	t.Helper()
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		call()
	}()
	require.NotNil(t, recovered, "the Must form panics on failure")
	err, ok := recovered.(error)
	require.True(t, ok, "the Must form panics with an error, got %T", recovered)
	assert.Equal(t, want.Error(), err.Error(), "the Must form panics with the same error")
}

// tableRows collects at most limit rows and fails when the table offers more
// rows or calls yield again after it returned false.
func tableRows(t *testing.T, seq iter.Seq[[]modal.Relation], limit int) [][]modal.Relation {
	t.Helper()
	require.NotNil(t, seq, "a successful table is a non-nil iterator")
	var rows [][]modal.Relation
	stopped := false
	seq(func(row []modal.Relation) bool {
		if stopped {
			t.Errorf("yield called again after it returned false")
			return false
		}
		rows = append(rows, row)
		if len(rows) > limit {
			stopped = true
			return false
		}
		return true
	})
	require.LessOrEqual(t, len(rows), limit, "the table did not end after %d rows", limit)
	return rows
}

// controlMode is the closed P8 mode of architecture MOD-A10: P1, a second
// with deviation -2^23, a third with deviation (-2^24+1)/2, and a fourth with
// deviation (2^24+3)/2. II->III and every relation from I but I->II have a
// fractional quality; II->IV and IV->II' are not representable; II->I' is an
// augmented seventh with deviation 2^23-1.
func controlMode(t *testing.T) mode.Mode {
	t.Helper()
	second := interval.MustFromDeviation(2, alteration.MustFromInt(-(1 << 23)))
	third := interval.MustFromDeviation(3, alteration.MustFromFraction(-(1<<24)+1, 2))
	fourth := interval.MustFromDeviation(4, alteration.MustFromFraction((1<<24)+3, 2))
	m, err := mode.ClosedFromDegreeIntervals([]interval.Interval{interval.PerfectUnison(), second, third, fourth}, interval.PerfectOctave())
	require.NoError(t, err, "the control mode is valid in mode")
	return m
}

func ionianMode(t *testing.T) mode.Mode {
	t.Helper()
	return closedMode(t, "P1 M2 M3 P4 P5 M6 M7", "P8")
}

func pentatonicMode(t *testing.T) mode.Mode {
	t.Helper()
	return closedMode(t, "P1 M2 M3 P5 M6", "P8")
}

func TestRelationsFromControlRows(t *testing.T) {
	t.Parallel()
	const high, neutral, low = modal.High, modal.Neutral, modal.Low
	ionian := ionianMode(t)
	dorian := dorianMode(t)
	fourth := closedMode(t, "P1 M2 M3", "P4")
	wide := closedMode(t, "P1 M9", "P15")
	open := openMode(t, "P1 m3 P4")
	unisonPair := closedMode(t, "P1 d2", "P8")
	single := closedMode(t, "P1", "P8")
	for _, tt := range []struct {
		name string
		m    mode.Mode
		from int
		want []modal.Relation
	}{
		{"example 3: ionian from VII", ionian, 7, []modal.Relation{
			relation(7, 7, 0, "P1", neutral), relation(7, 1, 1, "m2", low), relation(7, 2, 1, "m3", low),
			relation(7, 3, 1, "P4", neutral), relation(7, 4, 1, "d5", low), relation(7, 5, 1, "m6", low), relation(7, 6, 1, "m7", low),
		}},
		{"example 3: ionian from IV", ionian, 4, []modal.Relation{
			relation(4, 4, 0, "P1", neutral), relation(4, 5, 0, "M2", high), relation(4, 6, 0, "M3", high),
			relation(4, 7, 0, "A4", high), relation(4, 1, 1, "P5", neutral), relation(4, 2, 1, "M6", high), relation(4, 3, 1, "M7", high),
		}},
		{"dorian from I", dorian, 1, []modal.Relation{
			relation(1, 1, 0, "P1", neutral), relation(1, 2, 0, "M2", high), relation(1, 3, 0, "m3", low),
			relation(1, 4, 0, "P4", neutral), relation(1, 5, 0, "P5", neutral), relation(1, 6, 0, "M6", high), relation(1, 7, 0, "m7", low),
		}},
		{"dorian from IV", dorian, 4, []modal.Relation{
			relation(4, 4, 0, "P1", neutral), relation(4, 5, 0, "M2", high), relation(4, 6, 0, "M3", high),
			relation(4, 7, 0, "P4", neutral), relation(4, 1, 1, "P5", neutral), relation(4, 2, 1, "M6", high), relation(4, 3, 1, "m7", low),
		}},
		{"pentatonic from III", pentatonicMode(t), 3, []modal.Relation{
			relation(3, 3, 0, "P1", neutral), relation(3, 4, 0, "m3", low), relation(3, 5, 0, "P4", neutral),
			relation(3, 1, 1, "m6", low), relation(3, 2, 1, "m7", low),
		}},
		{"P4 cycle from II", fourth, 2, []modal.Relation{
			relation(2, 2, 0, "P1", neutral), relation(2, 3, 0, "M2", high), relation(2, 1, 1, "m3", low),
		}},
		{"P4 cycle from III", fourth, 3, []modal.Relation{
			relation(3, 3, 0, "P1", neutral), relation(3, 1, 1, "m2", low), relation(3, 2, 1, "m3", low),
		}},
		{"P15 cycle from I keeps the compound interval", wide, 1, []modal.Relation{
			relation(1, 1, 0, "P1", neutral), relation(1, 2, 0, "M9", high),
		}},
		{"P15 cycle from II", wide, 2, []modal.Relation{
			relation(2, 2, 0, "P1", neutral), relation(2, 1, 1, "m7", low),
		}},
		{"open mode from I", open, 1, []modal.Relation{
			relation(1, 1, 0, "P1", neutral), relation(1, 2, 0, "m3", low), relation(1, 3, 0, "P4", neutral),
		}},
		{"open mode from II", open, 2, []modal.Relation{
			relation(2, 2, 0, "P1", neutral), relation(2, 3, 0, "M2", high),
		}},
		{"open mode from III", open, 3, []modal.Relation{relation(3, 3, 0, "P1", neutral)}},
		{"coinciding degrees from I", unisonPair, 1, []modal.Relation{
			relation(1, 1, 0, "P1", neutral), relation(1, 2, 0, "d2", low),
		}},
		{"coinciding degrees from II", unisonPair, 2, []modal.Relation{
			relation(2, 2, 0, "P1", neutral), relation(2, 1, 1, "A7", high),
		}},
		{"single degree", single, 1, []modal.Relation{relation(1, 1, 0, "P1", neutral)}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := modal.RelationsFrom(tt.m, tt.from)
			require.NoError(t, err)
			checkRelations(t, tt.want, got)
			checkRelations(t, tt.want, modal.MustRelationsFrom(tt.m, tt.from))
		})
	}
}

func TestRelationToControlValues(t *testing.T) {
	t.Parallel()
	ionian := ionianMode(t)
	for _, tt := range []struct {
		name     string
		m        mode.Mode
		from, to int
		want     modal.Relation
	}{
		{"VII to I of the next cycle", ionian, 7, 1, relation(7, 1, 1, "m2", modal.Low)},
		{"I to VII in the same cycle", ionian, 1, 7, relation(1, 7, 0, "M7", modal.High)},
		{"a degree to itself", ionian, 3, 3, relation(3, 3, 0, "P1", modal.Neutral)},
		{"open mode upward", openMode(t, "P1 m3 P4"), 2, 3, relation(2, 3, 0, "M2", modal.High)},
		{"another relation has an unsupported quality", openMode(t, "P1 M2 n3"), 1, 2, relation(1, 2, 0, "M2", modal.High)},
		{"control mode: augmented seventh to the next cycle", controlMode(t), 2, 1,
			modal.Relation{From: 2, To: 1, Cycle: 1, Interval: interval.MustFromDeviation(7, alteration.MustFromInt(1<<23-1)), Category: modal.High}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := modal.RelationTo(tt.m, tt.from, tt.to)
			require.NoError(t, err)
			checkRelation(t, tt.want, got)
			checkRelation(t, tt.want, modal.MustRelationTo(tt.m, tt.from, tt.to))
		})
	}
}

// relationFixtures are modes whose every relation has a supported quality.
func relationFixtures(t *testing.T) map[string]mode.Mode {
	t.Helper()
	return map[string]mode.Mode{
		"ionian":           ionianMode(t),
		"dorian":           dorianMode(t),
		"harmonic minor":   closedMode(t, "P1 M2 m3 P4 P5 m6 M7", "P8"),
		"pentatonic":       pentatonicMode(t),
		"P4 cycle":         closedMode(t, "P1 M2 M3", "P4"),
		"P15 cycle":        closedMode(t, "P1 M9", "P15"),
		"coinciding":       closedMode(t, "P1 d2", "P8"),
		"open":             openMode(t, "P1 m3 P4"),
		"single":           closedMode(t, "P1", "P8"),
		"double augmented": closedMode(t, "P1 M2 AA4 P5", "P8"),
	}
}

func TestRelationsAreConsistent(t *testing.T) {
	t.Parallel()
	for name, m := range relationFixtures(t) {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			k := m.Len()
			for from := 1; from <= k; from++ {
				row, err := modal.RelationsFrom(m, from)
				require.NoError(t, err)
				if m.IsOpen() {
					require.Len(t, row, k-from+1, "an open row ends at the last degree")
				} else {
					require.Len(t, row, k, "a closed row has every degree once")
				}
				targets := make([]int, len(row))
				for n, r := range row {
					assert.Equal(t, from, r.From)
					targets[n] = r.To
					want := 0
					if r.To < from {
						want = 1
					}
					assert.Equal(t, want, r.Cycle, "cycle of %d->%d", from, r.To)
					assert.True(t, r.Category.IsValid())
					one, err := modal.RelationTo(m, from, r.To)
					require.NoError(t, err)
					checkRelation(t, r, one)
				}
				var wantTargets []int
				for to := from; to <= k; to++ {
					wantTargets = append(wantTargets, to)
				}
				if !m.IsOpen() {
					for to := 1; to < from; to++ {
						wantTargets = append(wantTargets, to)
					}
				}
				assert.Equal(t, wantTargets, targets)
				checkRelation(t, modal.Relation{From: from, To: from, Cycle: 0, Interval: interval.PerfectUnison(), Category: modal.Neutral}, row[0])
			}
		})
	}
}

func qualityCounts(row []modal.Relation) (counts modal.QualityCounts, low, high int) {
	for _, r := range row {
		switch r.Interval.Quality() {
		case interval.Diminished:
			counts.Diminished++
		case interval.Minor:
			counts.Minor++
		case interval.Perfect:
			counts.Perfect++
		case interval.Major:
			counts.Major++
		case interval.Augmented:
			counts.Augmented++
		}
		switch r.Category {
		case modal.Low:
			low++
		case modal.High:
			high++
		}
	}
	return counts, low, high
}

// TestRelationsAgreeWithAcceptedResults checks M-14 and the links of MOD-6:
// the row from I is the relative analysis, a closed row is the relative
// analysis of the rotated mode, and with a P8 cycle the row gives the
// absolute counts and score.
func TestRelationsAgreeWithAcceptedResults(t *testing.T) {
	t.Parallel()
	for name, m := range relationFixtures(t) {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			first := modal.MustRelationsFrom(m, 1)
			relative := modal.MustRelativePositions(m)
			require.Len(t, first, len(relative))
			for n, p := range relative {
				checkRelation(t, modal.Relation{From: 1, To: p.Number, Cycle: 0, Interval: p.Interval, Category: p.Category}, first[n])
			}
			if m.IsOpen() {
				return
			}
			for from := 1; from <= m.Len(); from++ {
				row := modal.MustRelationsFrom(m, from)
				rotated := modal.MustRelativePositions(mode.MustFromDegree(m, from))
				require.Len(t, row, len(rotated))
				for n := range row {
					assert.True(t, rotated[n].Interval.Equal(row[n].Interval), "%s: relation %d of row %d", name, n+1, from)
					assert.Equal(t, rotated[n].Category, row[n].Category, "%s: relation %d of row %d", name, n+1, from)
				}
			}
			if !m.IsOctave() {
				return
			}
			absolute := modal.MustAbsolutePositions(m)
			for from := 1; from <= m.Len(); from++ {
				counts, low, high := qualityCounts(modal.MustRelationsFrom(m, from))
				assert.Equal(t, absolute[from-1].Counts, counts, "row %d", from)
				assert.Equal(t, absolute[from-1].Score, low-high, "row %d", from)
			}
		})
	}
}

func TestRelationTableControlValues(t *testing.T) {
	t.Parallel()
	const high, neutral, low = modal.High, modal.Neutral, modal.Low
	t.Run("P4 cycle", func(t *testing.T) {
		t.Parallel()
		m := closedMode(t, "P1 M2 M3", "P4")
		want := [][]modal.Relation{
			{relation(1, 1, 0, "P1", neutral), relation(1, 2, 0, "M2", high), relation(1, 3, 0, "M3", high)},
			{relation(2, 2, 0, "P1", neutral), relation(2, 3, 0, "M2", high), relation(2, 1, 1, "m3", low)},
			{relation(3, 3, 0, "P1", neutral), relation(3, 1, 1, "m2", low), relation(3, 2, 1, "m3", low)},
		}
		table, err := modal.RelationTable(m)
		require.NoError(t, err)
		rows := tableRows(t, table, 5)
		require.Len(t, rows, len(want))
		for i := range want {
			checkRelations(t, want[i], rows[i])
		}
	})
	t.Run("open mode", func(t *testing.T) {
		t.Parallel()
		m := openMode(t, "P1 m3 P4")
		want := [][]modal.Relation{
			{relation(1, 1, 0, "P1", neutral), relation(1, 2, 0, "m3", low), relation(1, 3, 0, "P4", neutral)},
			{relation(2, 2, 0, "P1", neutral), relation(2, 3, 0, "M2", high)},
			{relation(3, 3, 0, "P1", neutral)},
		}
		rows := tableRows(t, modal.MustRelationTable(m), 5)
		require.Len(t, rows, len(want))
		for i := range want {
			checkRelations(t, want[i], rows[i])
		}
	})
	t.Run("pentatonic scores", func(t *testing.T) {
		t.Parallel()
		rows := tableRows(t, modal.MustRelationTable(pentatonicMode(t)), 7)
		require.Len(t, rows, 5)
		for i, wantScore := range []int{-3, 0, 3, -2, 2} {
			require.Len(t, rows[i], 5)
			_, low, high := qualityCounts(rows[i])
			assert.Equal(t, wantScore, low-high, "row %d", i+1)
		}
	})
	t.Run("single degree", func(t *testing.T) {
		t.Parallel()
		rows := tableRows(t, modal.MustRelationTable(closedMode(t, "P1", "P8")), 3)
		require.Len(t, rows, 1)
		checkRelations(t, []modal.Relation{relation(1, 1, 0, "P1", neutral)}, rows[0])
	})
}

func TestRelationTableRowsAreRelationsFrom(t *testing.T) {
	t.Parallel()
	for name, m := range relationFixtures(t) {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			table, err := modal.RelationTable(m)
			require.NoError(t, err)
			rows := tableRows(t, table, m.Len()+2)
			require.Len(t, rows, m.Len())
			for i, row := range rows {
				checkRelations(t, modal.MustRelationsFrom(m, i+1), row)
			}
			assert.Equal(t, rows, tableRows(t, table, m.Len()+2), "a repeated walk")
			if m.IsOctave() {
				absolute := modal.MustAbsolutePositions(m)
				for i, row := range rows {
					counts, low, high := qualityCounts(row)
					assert.Equal(t, absolute[i].Counts, counts, "row %d", i+1)
					assert.Equal(t, absolute[i].Score, low-high, "row %d", i+1)
				}
			}
		})
	}
}

func TestRelationTableWalk(t *testing.T) {
	t.Parallel()
	t.Run("stops after the first row", func(t *testing.T) {
		t.Parallel()
		table := modal.MustRelationTable(ionianMode(t))
		calls := 0
		table(func(row []modal.Relation) bool {
			calls++
			assert.Len(t, row, 7)
			return false
		})
		assert.Equal(t, 1, calls, "no yield after it returned false")
	})
	t.Run("snapshot of the mode", func(t *testing.T) {
		t.Parallel()
		m := ionianMode(t)
		table := modal.MustRelationTable(m)
		row := modal.MustRelationsFrom(m, 7)
		want := tableRows(t, table, 9)
		require.NoError(t, m.Mirror())
		require.False(t, m.Equal(ionianMode(t)), "the mode variable changed")
		assert.Equal(t, want, tableRows(t, table, 9), "the iterator keeps the structure of its call")
		checkRelations(t, modal.MustRelationsFrom(ionianMode(t), 7), row)
	})
	t.Run("rows and relations belong to the caller", func(t *testing.T) {
		t.Parallel()
		m := closedMode(t, "P1 M2 M3", "P4")
		table := modal.MustRelationTable(m)
		want := tableRows(t, table, 5)
		for i := range want {
			rows := tableRows(t, table, 5)
			rows[i][0] = modal.Relation{}
			rows[i] = append(rows[i][:1], relation(9, 9, 9, "P1", modal.Low))
			for j := range rows {
				if j != i {
					assert.Equal(t, want[j], rows[j], "row %d after changing row %d", j+1, i+1)
				}
			}
			assert.Equal(t, want, tableRows(t, table, 5))
		}
		row := modal.MustRelationsFrom(m, 2)
		row[1] = modal.Relation{}
		checkRelations(t, want[1], modal.MustRelationsFrom(m, 2))
	})
}

func TestRelationFailures(t *testing.T) {
	t.Parallel()
	ionian := ionianMode(t)
	open := openMode(t, "P1 m3 P4")
	neutral := openMode(t, "P1 M2 n3")
	control := controlMode(t)
	const maxInt, minInt = math.MaxInt, math.MinInt
	type request struct {
		name   string
		run    func() (any, error)
		must   func()
		zero   any
		causes []error
	}
	to := func(name string, m mode.Mode, from, target int, causes ...error) request {
		return request{name: name, zero: modal.Relation{}, causes: causes,
			run:  func() (any, error) { return modal.RelationTo(m, from, target) },
			must: func() { modal.MustRelationTo(m, from, target) }}
	}
	row := func(name string, m mode.Mode, from int, causes ...error) request {
		return request{name: name, zero: []modal.Relation(nil), causes: causes,
			run:  func() (any, error) { return modal.RelationsFrom(m, from) },
			must: func() { modal.MustRelationsFrom(m, from) }}
	}
	table := func(name string, m mode.Mode, causes ...error) request {
		return request{name: name, zero: iter.Seq[[]modal.Relation](nil), causes: causes,
			run:  func() (any, error) { return modal.RelationTable(m) },
			must: func() { modal.MustRelationTable(m) }}
	}
	unset := []error{modal.ErrUnset, mode.ErrUnset}
	invalid := []error{modal.ErrInvalidDegree}
	structure := []error{modal.ErrUnsupportedStructure}
	quality := []error{modal.ErrUnsupportedQuality}
	for _, tt := range []request{
		to("unset mode", mode.Mode{}, 1, 1, unset...),
		to("unset mode before invalid numbers", mode.Mode{}, 0, maxInt, unset...),
		row("unset mode", mode.Mode{}, 1, unset...),
		row("unset mode before an invalid number", mode.Mode{}, minInt, unset...),
		table("unset mode", mode.Mode{}, unset...),

		to("source 0", ionian, 0, 1, invalid...),
		to("source k+1", ionian, 8, 1, invalid...),
		to("source -1", ionian, -1, 1, invalid...),
		to("target 0", ionian, 1, 0, invalid...),
		to("target k+1", ionian, 1, 8, invalid...),
		to("source 0 and target k+1", ionian, 0, 8, invalid...),
		to("source k+1 and target 0", ionian, 8, 0, invalid...),
		to("source MinInt and target MaxInt", ionian, minInt, maxInt, invalid...),
		to("source MaxInt", ionian, maxInt, 1, invalid...),
		to("target MinInt", ionian, 1, minInt, invalid...),
		row("source 0", ionian, 0, invalid...),
		row("source k+1", ionian, 8, invalid...),
		row("source MinInt", ionian, minInt, invalid...),
		row("source MaxInt", ionian, maxInt, invalid...),
		to("open: target 0 is a number error, not structure", open, 3, 0, invalid...),
		to("open: source 0 before structure", open, 0, 1, invalid...),

		to("open: target below the source", open, 3, 1, structure...),
		to("open: II to I", open, 2, 1, structure...),
		to("open: structure before quality", neutral, 3, 2, structure...),

		row("neutral third from I", neutral, 1, quality...),
		row("neutral second from II", neutral, 2, quality...),
		to("own neutral third", neutral, 1, 3, quality...),
		to("own neutral second", neutral, 2, 3, quality...),
		table("neutral relations", neutral, quality...),

		row("control: representability of the row from II before its quality", control, 2, representable()...),
		row("control: representability of the row from IV before its quality", control, 4, representable()...),
		table("control: representability of the whole table", control, representable()...),
		to("control: unrepresentable II to IV", control, 2, 4, representable()...),
		row("control: quality of the row from I", control, 1, quality...),
		to("control: quality of II to III", control, 2, 3, quality...),
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.run()
			checkModalError(t, err, tt.causes...)
			assert.Equal(t, tt.zero, got, "no partial result on failure")
			checkMustPanic(t, err, tt.must)
		})
	}
	t.Run("the control table fails without rows", func(t *testing.T) {
		t.Parallel()
		got, err := modal.RelationTable(control)
		assert.Nil(t, got)
		checkModalError(t, err, representable()...)
		assert.NotErrorIs(t, err, modal.ErrUnsupportedQuality)
	})
	t.Run("the open row from III has only its unison", func(t *testing.T) {
		t.Parallel()
		got, err := modal.RelationsFrom(neutral, 3)
		require.NoError(t, err)
		checkRelations(t, []modal.Relation{relation(3, 3, 0, "P1", modal.Neutral)}, got)
	})
}

func TestRelationQueriesDoNotChangeTheMode(t *testing.T) {
	t.Parallel()
	m := ionianMode(t)
	saved := m
	_ = modal.MustRelationsFrom(m, 4)
	_ = modal.MustRelationTo(m, 7, 1)
	for row := range modal.MustRelationTable(m) {
		row[0] = modal.Relation{}
	}
	assert.True(t, saved.Equal(m))
	checkRelation(t, relation(4, 7, 0, "A4", modal.High), modal.MustRelationsFrom(m, 4)[3])
}

func TestRelationQueriesConcurrently(t *testing.T) {
	t.Parallel()
	m := ionianMode(t)
	table := modal.MustRelationTable(m)
	want := tableRows(t, table, 9)
	var wg sync.WaitGroup
	for worker := range 8 {
		wg.Go(func() {
			from := worker%7 + 1
			for range 20 {
				got, err := modal.RelationsFrom(m, from)
				assert.NoError(t, err)
				assert.Equal(t, want[from-1], got)
				one, err := modal.RelationTo(m, from, 1)
				assert.NoError(t, err)
				assert.Equal(t, want[from-1][slices.IndexFunc(want[from-1], func(r modal.Relation) bool { return r.To == 1 })], one)
				assert.Equal(t, want, slices.Collect(table))
			}
		})
	}
	wg.Wait()
}
