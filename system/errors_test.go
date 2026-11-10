package system_test

import (
	"testing"

	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
)

// Tests for the error contract: P-16, NF-3, contract 5.2, criterion K-18.

func TestSentinelErrorsAreDistinct(t *testing.T) {
	t.Parallel()

	all := append(categories(), causes()...)
	for i, x := range all {
		assert.Error(t, x.err, x.name)
		for j, y := range all {
			if i == j {
				assert.ErrorIs(t, x.err, y.err)
				continue
			}
			assert.NotErrorIs(t, x.err, y.err, "%s must not match %s", x.name, y.name)
		}
	}
}

// TestUnsetSystem_AllOperations checks that every operation with an error, called on an
// unset system with otherwise valid inputs, reports E-1 without a panic, and that its
// Must variant panics with that error (K-18).
func TestUnsetSystem_AllOperations(t *testing.T) {
	t.Parallel()

	c4 := pt(spelling.C(), 4)
	e4 := pt(spelling.E(), 4)
	c0 := cls(0, 12)

	for _, tc := range []struct {
		name string
		s    *system.System
	}{
		{"nil", nil},
		{"zero value", &system.System{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := tc.s
			ops := []struct {
				name string
				call func() error
				must func()
			}{
				{"Offset", func() error { _, err := s.Offset(letter.C); return err }, func() { s.MustOffset(letter.C) }},
				{"StepAfter", func() error { _, err := s.StepAfter(letter.C); return err }, func() { s.MustStepAfter(letter.C) }},
				{"NaturalClass", func() error { _, err := s.NaturalClass(letter.C); return err }, func() { s.MustNaturalClass(letter.C) }},
				{"Regularity", func() error { _, err := s.Regularity(); return err }, func() { s.MustRegularity() }},
				{"HasTraditionalLayout", func() error { _, err := s.HasTraditionalLayout(); return err }, func() { s.MustHasTraditionalLayout() }},
				{"NamesSubsetOf", func() error { _, err := s.NamesSubsetOf(system.Traditional()); return err }, func() { s.MustNamesSubsetOf(system.Traditional()) }},
				{"Admits", func() error { _, err := s.Admits(spelling.C()); return err }, func() { s.MustAdmits(spelling.C()) }},
				{"Class", func() error { _, err := s.Class(spelling.C()); return err }, func() { s.MustClass(spelling.C()) }},
				{"Carry", func() error { _, err := s.Carry(spelling.C()); return err }, func() { s.MustCarry(spelling.C()) }},
				{"ShiftedClass", func() error { _, err := s.ShiftedClass(spelling.C(), 1); return err }, func() { s.MustShiftedClass(spelling.C(), 1) }},
				{"EquivalentNames", func() error { _, err := s.EquivalentNames(spelling.C(), spelling.D()); return err }, func() { s.MustEquivalentNames(spelling.C(), spelling.D()) }},
				{"Spellings", func() error { _, err := s.Spellings(c0, a(1)); return err }, func() { s.MustSpellings(c0, a(1)) }},
				{"EnharmonicNames", func() error { _, err := s.EnharmonicNames(spelling.C(), a(1)); return err }, func() { s.MustEnharmonicNames(spelling.C(), a(1)) }},
				{"AdmitsPitch", func() error { _, err := s.AdmitsPitch(c4); return err }, func() { s.MustAdmitsPitch(c4) }},
				{"PitchClass", func() error { _, err := s.PitchClass(c4); return err }, func() { s.MustPitchClass(c4) }},
				{"Compare", func() error { _, err := s.Compare(c4, e4); return err }, func() { s.MustCompare(c4, e4) }},
				{"Between", func() error { _, err := s.Between(c4, c4, e4); return err }, func() { s.MustBetween(c4, c4, e4) }},
				{"Sorted", func() error { _, err := s.Sorted([]pitch.Pitch{c4}); return err }, func() { s.MustSorted([]pitch.Pitch{c4}) }},
				{"Distance", func() error { _, err := s.Distance(c4, e4); return err }, func() { s.MustDistance(c4, e4) }},
				{"Equivalent", func() error { _, err := s.Equivalent(c4, e4); return err }, func() { s.MustEquivalent(c4, e4) }},
				{"EnharmonicPitches", func() error { _, err := s.EnharmonicPitches(c4, a(1)); return err }, func() { s.MustEnharmonicPitches(c4, a(1)) }},
				{"WrittenOctave", func() error { _, err := s.WrittenOctave(c4); return err }, func() { s.MustWrittenOctave(c4) }},
				{"FromWritten", func() error { _, err := s.FromWritten(spelling.C(), o(4)); return err }, func() { s.MustFromWritten(spelling.C(), o(4)) }},
				{"RaiseWritten", func() error { _, err := s.RaiseWritten(c4); return err }, func() { s.MustRaiseWritten(c4) }},
				{"LowerWritten", func() error { _, err := s.LowerWritten(c4); return err }, func() { s.MustLowerWritten(c4) }},
				{"RaiseWrittenBy", func() error { _, err := s.RaiseWrittenBy(c4, 1); return err }, func() { s.MustRaiseWrittenBy(c4, 1) }},
				{"LowerWrittenBy", func() error { _, err := s.LowerWrittenBy(c4, 1); return err }, func() { s.MustLowerWrittenBy(c4, 1) }},
				{"RespellWritten", func() error { _, err := s.RespellWritten(c4, spelling.BSharp()); return err }, func() { s.MustRespellWritten(c4, spelling.BSharp()) }},
				{"NaturalWritten", func() error { _, err := s.NaturalWritten(c4); return err }, func() { s.MustNaturalWritten(c4) }},
				{"Spell", func() error { _, err := s.Spell(c0, system.Sharps()); return err }, func() { s.MustSpell(c0, system.Sharps()) }},
				{"PitchAt", func() error { _, err := s.PitchAt(c0, o(4), system.Sharps()); return err }, func() { s.MustPitchAt(c0, o(4), system.Sharps()) }},
				{"Move", func() error { _, err := s.Move(c4, 1, system.Sharps()); return err }, func() { s.MustMove(c4, 1, system.Sharps()) }},
				{"Walk", func() error { _, err := s.Walk(c4, e4, system.Sharps()); return err }, func() { s.MustWalk(c4, e4, system.Sharps()) }},
			}

			for _, op := range ops {
				var err error
				assert.NotPanics(t, func() { err = op.call() }, op.name)
				requireErr(t, err, system.ErrInvalid, nil)
				requirePanicErr(t, system.ErrInvalid, nil, op.must)
			}
		})
	}
}

// TestMust_ReturnsSameAsPlain checks that Must variants return the value of the
// wrapped function when it succeeds.
func TestMust_ReturnsSameAsPlain(t *testing.T) {
	t.Parallel()

	s := system.Traditional()
	c4 := pt(spelling.C(), 4)

	r, err := system.NewRule(system.MustNewProgression(a(0), a(1)))
	assert.NoError(t, err)
	assert.True(t, r.Equal(system.MustNewRule(system.MustNewProgression(a(0), a(1)))))

	sub, err := system.Integers().SubsetOf(system.MustMultiples(2))
	assert.NoError(t, err)
	assert.Equal(t, sub, system.Integers().MustSubsetOf(system.MustMultiples(2)))

	reg, err := s.Regularity()
	assert.NoError(t, err)
	assert.Equal(t, reg, s.MustRegularity())

	layout, err := s.HasTraditionalLayout()
	assert.NoError(t, err)
	assert.Equal(t, layout, s.MustHasTraditionalLayout())

	w, err := s.WrittenOctave(c4)
	assert.NoError(t, err)
	assert.Equal(t, w, s.MustWrittenOctave(c4))

	sorted, err := s.Sorted([]pitch.Pitch{pt(spelling.E(), 4), c4})
	assert.NoError(t, err)
	assert.Equal(t, sorted, s.MustSorted([]pitch.Pitch{pt(spelling.E(), 4), c4}))

	eq, err := s.EquivalentNames(spelling.ESharp(), spelling.F())
	assert.NoError(t, err)
	assert.Equal(t, eq, s.MustEquivalentNames(spelling.ESharp(), spelling.F()))
}
