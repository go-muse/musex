package system_test

import (
	"fmt"
	"slices"
	"sync"
	"testing"

	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for concurrent use: P-15, NF-2, contract 5.3. Run with -race.

const workers = 16

// snapshot computes a deterministic digest of many read operations on shared values.
func snapshot(s *system.System, r *system.Rule, rule system.SpellingRule) string {
	out := make([]string, 0, 128)
	add := func(v any, err error) { out = append(out, fmt.Sprint(v, err)) }

	for _, n := range []spelling.Spelling{spelling.C(), spelling.BSharp(), spelling.CFlat(), dHalfSharp()} {
		add(s.Admits(n))
		add(s.Class(n))
		add(s.Carry(n))
		add(s.ShiftedClass(n, 3))
		add(s.FromWritten(n, o(3)))
	}
	c4, e5 := pt(spelling.C(), 4), pt(spelling.E(), 5)
	add(s.Compare(c4, e5))
	add(s.Distance(c4, e5))
	add(s.Equivalent(c4, pt(spelling.BSharp(), 4)))
	add(s.WrittenOctave(pt(spelling.BSharp(), 4)))
	add(s.RaiseWritten(pt(spelling.B(), 3)))
	add(s.Sorted([]pitch.Pitch{e5, c4, pt(spelling.BSharp(), 4)}))
	for c := range 12 {
		add(s.Spell(cls(c, 12), rule))
		add(s.Move(c4, system.Step(c), rule))
	}
	add(s.Regularity())
	add(s.HasTraditionalLayout())
	add(s.NamesSubsetOf(system.QuarterTones24()))
	add(r.Contains(q(1, 2)))
	add(r.SubsetOf(system.Integers()))
	add(system.Integers().SubsetOf(r))
	out = append(out, s.String(), r.String(), rule.String())

	if seq, err := r.Values(a(2)); err == nil {
		out = append(out, fmt.Sprint(slices.Collect(seq)))
	}
	if seq, err := s.Walk(c4, e5, rule); err == nil {
		out = append(out, fmt.Sprint(slices.Collect(seq)))
	}
	if seq, err := s.Spellings(cls(3, 12), a(2)); err == nil {
		out = append(out, fmt.Sprint(slices.Collect(seq)))
	}
	return fmt.Sprint(out)
}

func TestConcurrentUse(t *testing.T) {
	t.Parallel()

	s := system.Traditional()
	r := system.MustMultiples(2)
	rule := system.Sharps()
	want := snapshot(s, r, rule)

	got := make([]string, workers)
	var wg sync.WaitGroup
	for i := range workers {
		wg.Go(func() { got[i] = snapshot(s, r, rule) })
	}
	wg.Wait()

	for i := range workers {
		assert.Equal(t, want, got[i], "worker %d", i)
	}
}

func TestConcurrentBuiltins(t *testing.T) {
	t.Parallel()

	rows := builtinTable()
	got := make([][]*system.System, workers)
	var wg sync.WaitGroup
	for i := range workers {
		wg.Go(func() {
			got[i] = slices.Collect(system.Builtins())
			for _, row := range rows {
				lookup, _ := system.LookupBuiltin(row.name)
				got[i] = append(got[i], lookup, row.get())
			}
		})
	}
	wg.Wait()

	want := make([]*system.System, 0, 3*len(rows))
	for _, row := range rows {
		want = append(want, row.get())
	}
	for _, row := range rows {
		want = append(want, row.get(), row.get())
	}
	for i := range workers {
		require.Len(t, got[i], len(want))
		for j := range want {
			assert.Same(t, want[j], got[i][j], "worker %d, element %d", i, j)
		}
	}
}

func TestConcurrentTraversalOfOneSequence(t *testing.T) {
	t.Parallel()

	s := system.Traditional()
	walk, err := s.Walk(pt(spelling.C(), 4), pt(spelling.C(), 6), system.Flats())
	require.NoError(t, err)
	values, err := system.MustMultiples(4).Values(a(3))
	require.NoError(t, err)
	names, err := s.Spellings(cls(0, 12), a(3))
	require.NoError(t, err)

	wantWalk := slices.Collect(walk)
	wantValues := slices.Collect(values)
	wantNames := slices.Collect(names)
	require.Len(t, wantWalk, 25)

	type result struct {
		walk   []pitch.Pitch
		values string
		names  []spelling.Spelling
	}
	got := make([]result, workers)
	var wg sync.WaitGroup
	for i := range workers {
		wg.Go(func() {
			got[i] = result{
				walk:   slices.Collect(walk),
				values: fmt.Sprint(slices.Collect(values)),
				names:  slices.Collect(names),
			}
		})
	}
	wg.Wait()

	for i := range workers {
		assert.Equal(t, wantWalk, got[i].walk, "worker %d", i)
		assert.Equal(t, fmt.Sprint(wantValues), got[i].values, "worker %d", i)
		assert.Equal(t, wantNames, got[i].names, "worker %d", i)
	}
}

func TestConcurrentCustomSpelling(t *testing.T) {
	t.Parallel()

	// The function is safe for concurrent use, so the rule is too (contract 5.3).
	rule := system.MustCustomSpelling(sharpNames(tradSharpNames()...))
	s := system.Traditional()
	want := snapshot(s, system.Integers(), rule)

	got := make([]string, workers)
	var wg sync.WaitGroup
	for i := range workers {
		wg.Go(func() { got[i] = snapshot(s, system.Integers(), rule) })
	}
	wg.Wait()
	for i := range workers {
		assert.Equal(t, want, got[i], "worker %d", i)
	}
}
