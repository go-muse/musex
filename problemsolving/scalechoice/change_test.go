package scalechoice_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/problemsolving/scalechoice"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReplaceAllChangesCollectionReadingAndBothEdges(t *testing.T) {
	t.Parallel()
	input := pitched("C D A B E F G", 4, nil)
	o := full()
	o.Center = spelling.C()
	o.System = system.Traditional()
	a, err := scalechoice.AnalyzePhrase(input, scalechoice.Classical(), o)
	require.NoError(t, err)
	before := describe(a)
	change, err := a.ReplaceAll(spelling.F(), spelling.FSharp())
	require.NoError(t, err)
	assert.True(t, change.After.IsValid())
	assert.True(t, change.After.IsPhrase())
	assert.Equal(t, []int{5}, change.Matched)
	assert.Equal(t, []int{5}, change.Changed)
	assert.Equal(t, "C D E G A B", toneText(change.Tones.Kept))
	assert.Equal(t, "F", toneText(change.Tones.Removed))
	assert.Equal(t, "F+1", toneText(change.Tones.Added))
	assert.Empty(t, change.Groups.Kept)
	assert.Equal(t, []int{0}, change.Groups.Removed)
	assert.Equal(t, []int{0}, change.Groups.Added)
	assert.Empty(t, change.Readings.Kept)
	assert.Equal(t, []string{"ionian/C"}, readingKeys(change.Readings.Removed))
	assert.Equal(t, []string{"lydian/C"}, readingKeys(change.Readings.Added))
	assert.Equal(t, []string{"lydian/C"}, allReadingKeys(change.After))
	assert.True(t, change.After.Options().Center.Equal(spelling.C()))
	assert.True(t, change.After.Scope().Equal(a.Scope()))
	assert.True(t, change.After.Options().System.Equal(o.System))
	assert.Equal(t, "C D A B E F+1 G", notesText(change.After.Notes()))
	checkNotes(t, input, a.Notes())
	assert.Equal(t, before, describe(a))
	require.Len(t, change.Neighbors, 2)
	for index, edge := range change.Neighbors {
		assert.Equal(t, 4+index, edge.From)
		assert.Equal(t, 5+index, edge.To)
	}
	checkMotion(t, change.Neighbors[0].Before, "m2", 1, o.System)
	checkMotion(t, change.Neighbors[0].After, "M2", 2, o.System)
	checkMotion(t, change.Neighbors[1].Before, "M2", 2, o.System)
	checkMotion(t, change.Neighbors[1].After, "m2", 1, o.System)
	must := a.MustReplaceAll(spelling.F(), spelling.FSharp())
	assert.Equal(t, describe(change.After), describe(must.After))
	assert.Equal(t, change.Matched, must.Matched)
	// The removed reading can identify the exact new note outside its degrees.
	assert.Empty(t, change.Readings.Removed[0].MustDegreesOf(spelling.FSharp()))
	assert.Equal(t, []int{4}, change.Readings.Added[0].MustDegreesOf(spelling.FSharp()))
}

func TestReplaceAtKeepsOtherOccurrencesAndReportsNoMatches(t *testing.T) {
	t.Parallel()
	o := full()
	o.System = system.Traditional()
	o.Center = spelling.C()
	a, err := scalechoice.AnalyzePhrase(pitched("C D E F G A B F", 4, nil), scalechoice.Classical(), o)
	require.NoError(t, err)
	ch, err := a.ReplaceAt(3, pitched("F+1", 4, nil)[0])
	require.NoError(t, err)
	assert.Equal(t, []int{3}, ch.Matched)
	assert.Equal(t, []int{3}, ch.Changed)
	assert.Equal(t, "C D E F+1 G A B F", notesText(ch.After.Notes()))
	assert.Empty(t, ch.Tones.Removed)
	assert.Equal(t, "F+1", toneText(ch.Tones.Added))
	assert.Equal(t, "C D E F G A B", toneText(ch.Tones.Kept))
	assert.Empty(t, ch.After.Groups())
	assert.True(t, ch.After.IsValid())
	assert.Equal(t, []int{0}, ch.Groups.Removed)
	assert.Empty(t, ch.Groups.Added)
	assert.Empty(t, ch.Readings.Added)
	require.Len(t, ch.Neighbors, 2)
	assert.Equal(t, 2, ch.Neighbors[0].From)
	assert.Equal(t, 3, ch.Neighbors[1].From)
	assert.Equal(t, describe(ch.After), describe(a.MustReplaceAt(3, pitched("F+1", 4, nil)[0]).After))
	// A set has the same replacement over all occurrences, but no neighbors.
	set, err := scalechoice.AnalyzeSet(a.Notes(), scalechoice.Classical(), o)
	require.NoError(t, err)
	all, err := set.ReplaceAll(spelling.F(), spelling.FSharp())
	require.NoError(t, err)
	assert.Equal(t, []int{3, 7}, all.Matched)
	assert.Equal(t, []int{3, 7}, all.Changed)
	assert.Empty(t, all.Neighbors)
}

func TestNoOpAbsentAndCentsOnlyReplacement(t *testing.T) {
	t.Parallel()
	o := full()
	o.System = system.Traditional()
	input := pitched("C D E F G A B F", 4, nil)
	a := scalechoice.MustAnalyzePhrase(input, scalechoice.Classical(), o)
	for _, tc := range []struct {
		old, new spelling.Spelling
		matched  []int
	}{{spelling.F(), spelling.F(), []int{3, 7}}, {spelling.CSharp(), spelling.DFlat(), nil}} {
		ch, err := a.ReplaceAll(tc.old, tc.new)
		require.NoError(t, err)
		assert.Equal(t, tc.matched, ch.Matched)
		assert.Empty(t, ch.Changed)
		assert.Empty(t, ch.Neighbors)
		assert.Empty(t, ch.Tones.Removed)
		assert.Empty(t, ch.Tones.Added)
		assert.Equal(t, []int{0}, ch.Groups.Kept)
		assert.Empty(t, ch.Groups.Removed)
		assert.Empty(t, ch.Groups.Added)
		assert.Len(t, ch.Readings.Kept, 7)
		assert.Empty(t, ch.Readings.Removed)
		assert.Empty(t, ch.Readings.Added)
		checkNotes(t, input, ch.After.Notes())
	}
	same := a.MustReplaceAt(0, input[0])
	assert.Equal(t, []int{0}, same.Matched)
	assert.Empty(t, same.Changed)
	assert.Empty(t, same.Neighbors)
	replacement := note.MustFromPitch(pitch.MustFromASPN(spelling.F(), 4), cents.MustFromInt(25), nil)
	ch := a.MustReplaceAt(3, replacement)
	assert.Equal(t, []int{3}, ch.Changed)
	assert.Equal(t, "C D E F G A B", toneText(ch.Tones.Kept))
	assert.Empty(t, ch.Tones.Added)
	assert.Empty(t, ch.Tones.Removed)
	assert.True(t, ch.After.Notes()[3].Cents().Equal(cents.MustFromInt(25)))
	assert.Len(t, ch.Neighbors, 2)
	checkMotion(t, ch.Neighbors[0].Before, "m2", 1, o.System)
	checkMotion(t, ch.Neighbors[0].After, "m2", 1, o.System)
	checkMotion(t, ch.Neighbors[1].Before, "M2", 2, o.System)
	checkMotion(t, ch.Neighbors[1].After, "M2", 2, o.System)
	// Equal record with a different settings source is still unchanged.
	src := source(system.Positions19())
	replacement = pitched("F", 4, src)[0]
	src.reset()
	ch = a.MustReplaceAt(3, replacement)
	assert.Empty(t, ch.Changed)
	assert.Empty(t, ch.Neighbors)
	assert.Zero(t, src.reads.Load())
}

func TestNeighborsZeroOctaveAndEdges(t *testing.T) {
	t.Parallel()
	o := opts()
	o.System = system.Traditional()
	input := append(pitched("C C", 4, nil), pitched("C", 5, nil)...)
	a := scalechoice.MustAnalyzePhrase(input, oneToneCatalog(), o)
	ch := a.MustReplaceAt(0, pitched("C", 5, nil)[0])
	assert.Equal(t, []int{0}, ch.Changed)
	require.Len(t, ch.Neighbors, 1)
	assert.Equal(t, 0, ch.Neighbors[0].From)
	assert.Equal(t, 1, ch.Neighbors[0].To)
	checkMotion(t, ch.Neighbors[0].Before, "P1", 0, o.System)
	checkMotion(t, ch.Neighbors[0].After, "-P8", -12, o.System)
	ch = a.MustReplaceAt(2, pitched("C", 4, nil)[0])
	require.Len(t, ch.Neighbors, 1)
	assert.Equal(t, 1, ch.Neighbors[0].From)
	checkMotion(t, ch.Neighbors[0].Before, "P8", 12, o.System)
	checkMotion(t, ch.Neighbors[0].After, "P1", 0, o.System)
	a = scalechoice.MustAnalyzePhrase(pitched("C", 4, nil), oneToneCatalog(), o)
	ch = a.MustReplaceAt(0, pitched("C", 5, nil)[0])
	assert.Empty(t, ch.Neighbors)
	// Adjacent changed places share one edge; it must occur only once.
	a = scalechoice.MustAnalyzePhrase(pitched("E F F G", 4, nil), scalechoice.Classical(), o)
	ch = a.MustReplaceAll(spelling.F(), spelling.FSharp())
	require.Len(t, ch.Neighbors, 3)
	for index, edge := range ch.Neighbors {
		assert.Equal(t, index, edge.From)
		assert.Equal(t, index+1, edge.To)
	}
	checkMotion(t, ch.Neighbors[1].Before, "P1", 0, o.System)
	checkMotion(t, ch.Neighbors[1].After, "P1", 0, o.System)
}

func TestEnharmonicReplacementChangesWritingAndKeepsFamily(t *testing.T) {
	t.Parallel()
	input := pitched("C F+1 G", 4, nil)
	for _, comparison := range []scalechoice.Comparison{scalechoice.BySpelling, scalechoice.ByPosition} {
		o := opts()
		o.Comparison = comparison
		o.System = system.Traditional()
		a := scalechoice.MustAnalyzePhrase(input, scalechoice.Classical(), o)
		ch := a.MustReplaceAt(1, pitched("G-1", 4, nil)[0])
		require.Len(t, ch.Neighbors, 2)
		checkMotion(t, ch.Neighbors[0].Before, "A4", 6, o.System)
		checkMotion(t, ch.Neighbors[0].After, "d5", 6, o.System)
		checkMotion(t, ch.Neighbors[1].Before, "m2", 1, o.System)
		checkMotion(t, ch.Neighbors[1].After, "A1", 1, o.System)
		if comparison == scalechoice.BySpelling {
			assert.Equal(t, "F+1", toneText(ch.Tones.Removed))
			assert.Equal(t, "G-1", toneText(ch.Tones.Added))
		} else {
			assert.Equal(t, []int{0, 6, 7}, classes(ch.Tones.Kept))
			assert.Empty(t, ch.Tones.Removed)
			assert.Empty(t, ch.Tones.Added)
			assert.Empty(t, ch.Groups.Added)
			assert.Empty(t, ch.Groups.Removed)
			assert.Empty(t, ch.Readings.Added)
			assert.Empty(t, ch.Readings.Removed)
		}
	}
	// Changing the representative of a regular positional tonic family is not
	// changing its identity; kept readings belong to the After snapshot.
	o := positionOptions()
	o.Center = spelling.CSharp()
	a := scalechoice.MustAnalyzeSet(notes("D-1"), oneToneCatalog(), o)
	require.Len(t, a.Groups(), 1)
	assert.Equal(t, []string{"one/D-1"}, allReadingKeys(a))
	ch := a.MustReplaceAll(spelling.DFlat(), spelling.CSharp())
	assert.Equal(t, []int{0}, ch.Groups.Kept)
	assert.Empty(t, ch.Groups.Removed)
	assert.Empty(t, ch.Groups.Added)
	assert.Equal(t, []string{"one/C+1"}, readingKeys(ch.Readings.Kept))
	assert.Empty(t, ch.Readings.Removed)
	assert.Empty(t, ch.Readings.Added)
	assert.Equal(t, []int{1}, classes(ch.Tones.Kept))
	assert.Empty(t, ch.Tones.Removed)
	assert.Empty(t, ch.Tones.Added)
	// ReplaceAll selects exact names even when positional classes coincide.
	a = scalechoice.MustAnalyzeSet(notes("C+1 D-1"), oneToneCatalog(), o)
	ch = a.MustReplaceAll(spelling.CSharp(), spelling.D())
	assert.Equal(t, []int{0}, ch.Matched)
	assert.Equal(t, "D D-1", notesText(ch.After.Notes()))
}

func TestMotionOperationsFailIndependently(t *testing.T) {
	t.Parallel()
	// The written distance between these alterations exceeds Interval's
	// deviation limit, while its chromatic distance still fits in int.
	o := opts()
	o.System = system.Traditional()
	low := spelling.MustNew(spelling.C().Letter(), alteration.MustFromInt(-alteration.MaxMagnitude))
	high := spelling.MustNew(spelling.C().Letter(), alteration.MustFromInt(alteration.MaxMagnitude))
	input := []note.Note{
		note.MustFromPitch(o.System.MustFromWritten(low, octave.FromASPN(4)), cents.Zero(), nil),
		note.MustFromPitch(o.System.MustFromWritten(high, octave.FromASPN(4)), cents.Zero(), nil),
	}
	a := scalechoice.MustAnalyzePhrase(input, template.Catalog{}, o)
	replacement := note.MustFromPitch(o.System.MustFromWritten(high, octave.FromASPN(5)), cents.Zero(), nil)
	ch := a.MustReplaceAt(1, replacement)
	require.Len(t, ch.Neighbors, 1)
	for _, motion := range []scalechoice.Motion{ch.Neighbors[0].Before, ch.Neighbors[0].After} {
		assert.ErrorIs(t, motion.NamedError, interval.ErrOutOfRange)
		assert.False(t, motion.Named.IsValid())
		require.NoError(t, motion.ChromaticError)
		assert.True(t, motion.Chromatic.IsValid())
	}
	assert.Equal(t, 2*alteration.MaxMagnitude, ch.Neighbors[0].Before.Chromatic.Steps())
	assert.Equal(t, 2*alteration.MaxMagnitude+12, ch.Neighbors[0].After.Chromatic.Steps())
}

func TestNamedMotionCanExistWhenChromaticDistanceOverflows(t *testing.T) {
	t.Parallel()
	o := opts()
	o.System = system.Traditional()
	register := math.MaxInt / 10
	input := append(pitched("C", 0, nil), pitched("C", register, nil)...)
	a := scalechoice.MustAnalyzePhrase(input, template.Catalog{}, o)
	ch := a.MustReplaceAt(1, pitched("C", register+1, nil)[0])
	require.Len(t, ch.Neighbors, 1)
	for index, motion := range []scalechoice.Motion{ch.Neighbors[0].Before, ch.Neighbors[0].After} {
		require.NoError(t, motion.NamedError)
		assert.True(t, motion.Named.IsValid())
		assert.Equal(t, 7*(register+index)+1, motion.Named.Size())
		assert.True(t, motion.Named.IsAscending())
		assert.True(t, motion.Named.Deviation().IsNatural())
		assert.ErrorIs(t, motion.ChromaticError, interval.ErrOutOfRange)
		assert.ErrorIs(t, motion.ChromaticError, system.ErrOutOfRange)
		assert.False(t, motion.Chromatic.IsValid())
	}
}

func TestReplaceAllRejectsUnrepresentableWrittenRegister(t *testing.T) {
	t.Parallel()
	for _, comparison := range []scalechoice.Comparison{scalechoice.BySpelling, scalechoice.ByPosition} {
		for _, tc := range []struct {
			register    int
			replacement spelling.Spelling
		}{{math.MaxInt, spelling.BSharp()}, {math.MinInt, spelling.CFlat()}} {
			o := opts()
			o.Comparison = comparison
			o.System = system.Traditional()
			input := pitched("C", tc.register, nil)
			a := scalechoice.MustAnalyzePhrase(input, template.Catalog{}, o)
			ch, err := a.ReplaceAll(spelling.C(), tc.replacement)
			assert.ErrorIs(t, err, system.ErrOutOfRange)
			assert.False(t, ch.After.IsValid())
			assert.Empty(t, ch.Matched)
			assert.Empty(t, ch.Changed)
			checkNotes(t, input, a.Notes())
		}
	}
}

func TestDifferencesUseIndicesOfTheCorrectSide(t *testing.T) {
	t.Parallel()
	a := scalechoice.MustAnalyzeSet(notes("C D E G A"), scalechoice.Classical(), opts())
	// Replacing E by B-flat keeps the F-major collection, removes C and G
	// major, and introduces B-flat major. F moves from old group 1 to new 0.
	ch := a.MustReplaceAll(spelling.E(), spelling.BFlat())
	require.Len(t, ch.After.Groups(), 2)
	assert.Equal(t, []int{0}, ch.Groups.Kept)
	assert.Equal(t, []int{0, 2}, ch.Groups.Removed)
	assert.Equal(t, []int{1}, ch.Groups.Added)
	assert.Equal(t, "C D G A", toneText(ch.Tones.Kept))
	assert.Equal(t, "E", toneText(ch.Tones.Removed))
	assert.Equal(t, "B-1", toneText(ch.Tones.Added))
	assert.Equal(t, []string{"ionian/F", "dorian/G", "phrygian/A", "lydian/B-1", "mixolydian/C", "aeolian/D", "locrian/E"}, readingKeys(ch.Readings.Kept))
	assert.Equal(t, []string{"ionian/B-1", "dorian/C", "phrygian/D", "lydian/E-1", "mixolydian/F", "aeolian/G", "locrian/A"}, readingKeys(ch.Readings.Added))
	require.Len(t, ch.Readings.Removed, 14)
	assert.Equal(t, []string{"ionian/C", "dorian/D", "phrygian/E", "lydian/F", "mixolydian/G", "aeolian/A", "locrian/B", "ionian/G", "dorian/A", "phrygian/B", "lydian/C", "mixolydian/D", "aeolian/E", "locrian/F+1"}, readingKeys(ch.Readings.Removed))
}

func TestIrregularTonicFamiliesKeepLetterIdentity(t *testing.T) {
	t.Parallel()
	o := positionOptions()
	o.System = irregular()
	o.Center = spelling.C()
	a := scalechoice.MustAnalyzeSet(notes("C"), oneToneCatalog(), o)
	require.Len(t, a.Groups(), 1)
	rs := a.Groups()[0].Readings
	require.Len(t, rs, 7)
	for index, r := range rs {
		assert.Equal(t, index, r.Tonic().Letter().Ordinal())
	}
	ch := a.MustReplaceAll(spelling.C(), spelling.BSharp())
	assert.Equal(t, []int{0}, ch.Groups.Kept)
	assert.Empty(t, ch.Groups.Removed)
	assert.Empty(t, ch.Groups.Added)
	assert.Len(t, ch.Readings.Kept, 7)
	assert.Empty(t, ch.Readings.Removed)
	assert.Empty(t, ch.Readings.Added)
	assert.Equal(t, readingKeys(rs), readingKeys(ch.Readings.Kept))
}
