package scalechoice_test

import (
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/problemsolving/scalechoice"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExplicitSystemAvoidsSourcesAndIsCopied(t *testing.T) {
	t.Parallel()
	sys := system.MustNew(system.Traditional().Definition())
	src := source(system.Positions19())
	input := pitched("C D E F G A B", 4, src)
	src.current.Store(nil)
	src.reset()
	o := positionOptions()
	o.System = sys
	a, err := scalechoice.AnalyzePhrase(input, scalechoice.Classical(), o)
	require.NoError(t, err)
	assert.Zero(t, src.reads.Load())
	assert.Len(t, a.Groups(), 1)
	before := describe(a)
	assert.NotSame(t, sys, a.Options().System)
	*sys = *system.QuarterTones24()
	assert.True(t, a.Options().System.Equal(system.Traditional()))
	returned := a.Options().System
	*returned = *system.Positions19()
	assert.True(t, a.Options().System.Equal(system.Traditional()))
	assert.Equal(t, before, describe(a))
	ch := a.MustReplaceAll(spelling.F(), spelling.FSharp())
	assert.Zero(t, src.reads.Load())
	assert.True(t, ch.After.Options().System.Equal(system.Traditional()))
	require.Len(t, ch.Neighbors, 2)
	checkMotion(t, ch.Neighbors[0].Before, "m2", 1, system.Traditional())
	checkNotes(t, input, a.Notes())
}

func TestContextReadsDistinctSourcesOnceAndKeepsSnapshot(t *testing.T) {
	t.Parallel()
	first, second := source(system.Traditional()), source(system.MustNew(system.Traditional().Definition()))
	input := append(pitched("C D E F", 4, first), pitched("G A B", 4, second)...)
	first.reset()
	second.reset()
	o := positionOptions()
	o.System = nil
	a, err := scalechoice.AnalyzePhrase(input, scalechoice.Classical(), o)
	require.NoError(t, err)
	assert.Equal(t, int64(1), first.reads.Load())
	assert.Equal(t, int64(1), second.reads.Load())
	assert.True(t, a.Options().System.Equal(system.Traditional()))
	first.current.Store(nil)
	second.current.Store(nil)
	before := describe(a)
	ch := a.MustReplaceAll(spelling.F(), spelling.FSharp())
	require.Len(t, ch.Neighbors, 2)
	checkMotion(t, ch.Neighbors[0].Before, "m2", 1, system.Traditional())
	checkMotion(t, ch.Neighbors[0].After, "M2", 2, system.Traditional())
	next := ch.After.MustReplaceAll(spelling.FSharp(), spelling.F())
	assert.Equal(t, before, describe(next.After))
	assert.Equal(t, int64(1), first.reads.Load())
	assert.Equal(t, int64(1), second.reads.Load())
}

func TestWrittenContextIsPerPlaceEvenAfterAnEarlierFailure(t *testing.T) {
	t.Parallel()
	for _, caseName := range []string{"unbound", "failed", "different"} {
		t.Run(caseName, func(t *testing.T) {
			t.Parallel()
			good := source(system.Traditional())
			bad := source(system.Positions19())
			outsider := pitched("C", 4, bad)[0]
			if caseName == "unbound" {
				outsider = pitched("C", 4, nil)[0]
			}
			if caseName == "failed" {
				bad.current.Store(nil)
			}
			input := append([]note.Note{outsider}, pitched("E F G", 4, good)...)
			good.reset()
			bad.reset()
			a, err := scalechoice.AnalyzePhrase(input, template.Catalog{}, opts())
			require.NoError(t, err)
			assert.Nil(t, a.Options().System)
			assert.Equal(t, int64(1), good.reads.Load())
			if caseName == "unbound" {
				assert.Zero(t, bad.reads.Load())
			} else {
				assert.Equal(t, int64(1), bad.reads.Load())
			}
			replacementSource := source(system.QuarterTones24())
			replacement := pitched("F+1", 4, replacementSource)[0]
			replacementSource.current.Store(nil)
			replacementSource.reset()
			good.current.Store(nil)
			ch := a.MustReplaceAt(2, replacement)
			require.Len(t, ch.Neighbors, 2)
			checkMotion(t, ch.Neighbors[0].Before, "m2", 1, system.Traditional())
			checkMotion(t, ch.Neighbors[0].After, "M2", 2, system.Traditional())
			checkMotion(t, ch.Neighbors[1].Before, "M2", 2, system.Traditional())
			checkMotion(t, ch.Neighbors[1].After, "m2", 1, system.Traditional())
			assert.True(t, ch.After.Notes()[2].HasOctave())
			assert.Zero(t, replacementSource.reads.Load())
			assert.Equal(t, int64(1), good.reads.Load())
			// The inherited place context survives another branch from After.
			next := ch.After.MustReplaceAll(spelling.FSharp(), spelling.F())
			require.Len(t, next.Neighbors, 2)
			checkMotion(t, next.Neighbors[0].After, "m2", 1, system.Traditional())
			assert.Equal(t, int64(1), good.reads.Load())
			assert.True(t, next.After.Notes()[2].HasOctave())
		})
	}
	// A repeated failing source is read once; successful later sources must
	// still be read, and their adjacent pairs remain measurable.
	bad, good := source(system.Traditional()), source(system.Traditional())
	input := append(pitched("C D", 4, bad), pitched("E F G", 4, good)...)
	bad.current.Store(nil)
	bad.reset()
	good.reset()
	a := scalechoice.MustAnalyzePhrase(input, template.Catalog{}, opts())
	assert.Equal(t, int64(1), bad.reads.Load())
	assert.Equal(t, int64(1), good.reads.Load())
	ch := a.MustReplaceAt(1, pitched("D+1", 4, nil)[0])
	require.Len(t, ch.Neighbors, 2)
	checkMotionError(t, ch.Neighbors[0].Before, scalechoice.ErrContext)
	checkMotionError(t, ch.Neighbors[0].After, scalechoice.ErrContext)
	checkMotionError(t, ch.Neighbors[1].Before, scalechoice.ErrContext)
	checkMotionError(t, ch.Neighbors[1].After, scalechoice.ErrContext)
	// Unknown context retains ReplaceAt's exact octave but cannot interpret it.
	assert.True(t, ch.After.Notes()[1].HasOctave())
	assert.Equal(t, 4, ch.After.Notes()[1].MustASPN())
	next := ch.After.MustReplaceAt(1, pitched("D", 5, nil)[0])
	assert.Equal(t, 5, next.After.Notes()[1].MustASPN())
	checkMotionError(t, next.Neighbors[1].After, scalechoice.ErrContext)
	assert.Equal(t, int64(1), bad.reads.Load())
	assert.Equal(t, int64(1), good.reads.Load())
}

func TestMotionContextFailuresAndMissingOctaves(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name  string
		input []note.Note
		want  error
	}{
		{"no systems", pitched("C D", 4, nil), scalechoice.ErrNoSystem},
		{"different systems", append(pitched("C", 4, source(system.Traditional())), pitched("D", 4, source(system.Positions19()))...), scalechoice.ErrDifferentSystems},
		{"missing first octave", append(notes("C"), pitched("D", 4, nil)...), interval.ErrNoOctave},
		{"missing second octave", append(pitched("C", 4, nil), notes("D")...), interval.ErrNoOctave},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := scalechoice.MustAnalyzePhrase(tc.input, template.Catalog{}, opts())
			ch := a.MustReplaceAt(1, notes("E")[0])
			require.Len(t, ch.Neighbors, 1)
			checkMotionError(t, ch.Neighbors[0].Before, tc.want)
			checkMotionError(t, ch.Neighbors[0].After, interval.ErrNoOctave)
		})
	}
}

func TestReplaceAllPreservesWrittenRegisterOrDropsUnavailableOctave(t *testing.T) {
	t.Parallel()
	for _, sys := range []*system.System{system.Traditional(), boundaryA()} {
		o := opts()
		o.System = sys
		input := []note.Note{note.MustFromPitch(sys.MustFromWritten(spelling.G(), octave.FromASPN(4)), cents.MustFromInt(17), nil), note.MustFromSpelling(spelling.G(), cents.MustFromInt(-9), nil)}
		a := scalechoice.MustAnalyzePhrase(input, template.Catalog{}, o)
		ch := a.MustReplaceAll(spelling.G(), spelling.A())
		assert.True(t, ch.After.Notes()[0].MustPitch().Equal(sys.MustFromWritten(spelling.A(), octave.FromASPN(4))))
		assert.True(t, ch.After.Notes()[0].Cents().Equal(cents.MustFromInt(17)))
		assert.False(t, ch.After.Notes()[1].HasOctave())
		assert.True(t, ch.After.Notes()[1].Cents().Equal(cents.MustFromInt(-9)))
	}
	a := scalechoice.MustAnalyzePhrase(pitched("F G", 4, nil), template.Catalog{}, opts())
	ch := a.MustReplaceAll(spelling.F(), spelling.FSharp())
	assert.False(t, ch.After.Notes()[0].HasOctave())
	assert.True(t, a.Notes()[0].HasOctave())
	checkMotionError(t, ch.Neighbors[0].After, interval.ErrNoOctave)
	// A known place system is still usable when a different source elsewhere
	// prevents a global system from being established.
	known := source(system.Traditional())
	input := append(pitched("C", 4, nil), pitched("F G", 4, known)...)
	known.reset()
	a = scalechoice.MustAnalyzePhrase(input, template.Catalog{}, opts())
	ch = a.MustReplaceAll(spelling.F(), spelling.FSharp())
	assert.True(t, ch.After.Notes()[1].HasOctave())
	assert.Equal(t, int64(1), known.reads.Load())
	// New spelling is inadmissible for this place; written search can retain it
	// without inventing a pitch or using another place's system.
	a = scalechoice.MustAnalyzePhrase(pitched("E G", 4, source(system.Traditional())), template.Catalog{}, opts())
	ch = a.MustReplaceAll(spelling.E(), halfFlatE())
	assert.True(t, ch.After.Notes()[0].Spelling().Equal(halfFlatE()))
	assert.False(t, ch.After.Notes()[0].HasOctave())
}
