package scalechoice_test

import (
	"slices"
	"testing"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/problemsolving/scalechoice"
	"github.com/go-muse/musex/search"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfiles(t *testing.T) {
	t.Parallel()
	assert.Equal(t, []string{"ionian", "dorian", "phrygian", "lydian", "mixolydian", "aeolian", "locrian"}, profileCodes(scalechoice.Classical()))
	assert.Equal(t, []string{"pentatonic-major", "pentatonic-minor"}, profileCodes(scalechoice.Pentatonic()))
	for _, c := range []template.Catalog{scalechoice.Classical(), scalechoice.Pentatonic()} {
		for _, e := range c.Entries() {
			assert.True(t, e.Template().IsOctave())
		}
	}
}

func TestNaturalSetReadingsAndCenter(t *testing.T) {
	t.Parallel()
	wantKeys := []string{"ionian/C", "dorian/D", "phrygian/E", "lydian/F", "mixolydian/G", "aeolian/A", "locrian/B"}
	for _, phrase := range []bool{false, true} {
		input := notes("C D E F G A B C")
		var a scalechoice.Analysis
		var err error
		if phrase {
			a, err = scalechoice.AnalyzePhrase(input, scalechoice.Classical(), full())
		} else {
			a, err = scalechoice.AnalyzeSet(input, scalechoice.Classical(), full())
		}
		require.NoError(t, err)
		assert.True(t, a.IsValid())
		assert.Equal(t, phrase, a.IsPhrase())
		checkNotes(t, input, a.Notes())
		assert.Equal(t, profileCodes(scalechoice.Classical()), profileCodes(a.Scope()))
		assert.Equal(t, full().Condition, a.Options().Condition)
		assert.Equal(t, full().Comparison, a.Options().Comparison)
		groups := a.Groups()
		require.Len(t, groups, 1)
		assert.Equal(t, "C D E F G A B", toneText(groups[0].Tones))
		assert.Empty(t, groups[0].Missing)
		assert.Equal(t, wantKeys, readingKeys(groups[0].Readings))
		assert.Empty(t, a.Continuations())
		assert.Equal(t, "C D E F G A B", toneText(a.Common()))
		for _, tone := range groups[0].Tones {
			assert.Len(t, tone.Names, 1)
			assert.False(t, tone.Class.IsValid())
		}
		for index, r := range groups[0].Readings {
			assert.True(t, r.Entry().IsValid())
			assert.Equal(t, 7, len(r.Degrees()))
			expected := append(names("C D E F G A B")[index:], names("C D E F G A B")[:index]...)
			assert.Equal(t, nameText(expected), nameText(r.Degrees()))
			degree, err := r.DegreesOf(spelling.D())
			require.NoError(t, err)
			assert.Equal(t, []int{(1-index+7)%7 + 1}, degree)
			none, err := r.DegreesOf(spelling.CSharp())
			require.NoError(t, err)
			assert.Empty(t, none)
		}
	}
	o := full()
	o.Center = spelling.D()
	a, err := scalechoice.AnalyzeSet(notes("C D E F G A B"), scalechoice.Classical(), o)
	require.NoError(t, err)
	require.Len(t, a.Groups(), 1)
	assert.Equal(t, []string{"dorian/D"}, allReadingKeys(a))
	r := a.Groups()[0].Readings[0]
	assert.Equal(t, []int{1}, r.MustDegreesOf(spelling.D()))
	assert.Equal(t, []int{3}, r.MustDegreesOf(spelling.F()))
	a, err = scalechoice.AnalyzeSet(notes("C D E F G A B"), catalog("ionian"), o)
	require.NoError(t, err)
	assert.True(t, a.IsValid())
	assert.Empty(t, a.Groups())
}

func TestPentatonicContinuationChoices(t *testing.T) {
	t.Parallel()
	a, err := scalechoice.AnalyzeSet(notes("C D E G A"), scalechoice.Classical(), opts())
	require.NoError(t, err)
	groups := a.Groups()
	require.Len(t, groups, 3)
	for index, want := range []struct{ tones, missing string }{{"C D E F G A B", "F B"}, {"C D E F G A B-1", "F B-1"}, {"C D E F+1 G A B", "F+1 B"}} {
		assert.Equal(t, want.tones, toneText(groups[index].Tones))
		assert.Equal(t, want.missing, toneText(groups[index].Missing))
		assert.Len(t, groups[index].Readings, 7)
	}
	assert.Equal(t, "C D E G A", toneText(a.Common()))
	continuations := a.Continuations()
	require.Len(t, continuations, 4)
	for index, tc := range []struct {
		name          string
		kept, dropped []int
	}{{"F", []int{0, 1}, []int{2}}, {"F+1", []int{2}, []int{0, 1}}, {"B-1", []int{1}, []int{0, 2}}, {"B", []int{0, 2}, []int{1}}} {
		assert.Equal(t, tc.name, nameText(continuations[index].Tone.Names))
		checkPartition(t, continuations[index].Groups, tc.kept, tc.dropped)
	}
	for _, tc := range []struct {
		names         string
		kept, dropped []int
	}{{"", []int{0, 1, 2}, nil}, {"F", []int{0, 1}, []int{2}}, {"F B", []int{0}, []int{1, 2}}, {"F F", []int{0, 1}, []int{2}}, {"E", []int{0, 1, 2}, nil}, {"C+1", nil, []int{0, 1, 2}}} {
		p, err := a.AfterAdding(names(tc.names)...)
		require.NoError(t, err)
		checkPartition(t, p, tc.kept, tc.dropped)
		assert.Equal(t, p, a.MustAfterAdding(names(tc.names)...))
	}
	// Full match compares distinct tones, not note count or repetitions.
	for _, input := range []string{"C D E G A", "C D E G A C"} {
		no, err := scalechoice.AnalyzeSet(notes(input), scalechoice.Classical(), full())
		require.NoError(t, err)
		assert.Empty(t, no.Groups())
		c := scalechoice.Classical().MustWith(scalechoice.Pentatonic().Entries()...)
		pent, err := scalechoice.AnalyzeSet(notes(input), c, full())
		require.NoError(t, err)
		require.Len(t, pent.Groups(), 1)
		assert.Equal(t, []string{"pentatonic-major/C", "pentatonic-minor/A"}, allReadingKeys(pent))
		assert.Empty(t, pent.Continuations())
	}
	// A single group can still have an undiscriminating missing tone.
	o := opts()
	o.Center = spelling.C()
	one, err := scalechoice.AnalyzeSet(notes("C D E F G A"), catalog("ionian"), o)
	require.NoError(t, err)
	require.Len(t, one.Continuations(), 1)
	assert.Equal(t, "B", nameText(one.Continuations()[0].Tone.Names))
	checkPartition(t, one.Continuations()[0].Groups, []int{0}, nil)
}

func TestEmptyScopeAndCustomEntries(t *testing.T) {
	t.Parallel()
	a, err := scalechoice.AnalyzeSet(notes("C"), template.Catalog{}, opts())
	require.NoError(t, err)
	assert.True(t, a.IsValid())
	assert.Empty(t, a.Groups())
	assert.Empty(t, a.Common())
	assert.Empty(t, a.Continuations())
	p, err := a.AfterAdding(spelling.D())
	require.NoError(t, err)
	assert.Empty(t, p.Kept)
	assert.Empty(t, p.Dropped)
	e := template.MustNewEntry("own", []string{"first alias", "second alias"}, template.Ionian())
	c := template.MustNewCatalog(e, template.MustNewEntry("other", []string{"third alias"}, template.Ionian()))
	o := full()
	o.Center = spelling.C()
	a, err = scalechoice.AnalyzeSet(notes("C D E F G A B"), c, o)
	require.NoError(t, err)
	require.Len(t, a.Groups(), 1)
	rs := a.Groups()[0].Readings
	assert.Equal(t, []string{"own/C", "other/C"}, readingKeys(rs))
	assert.Equal(t, e.Names(), rs[0].Entry().Names())
}

func TestPositionNamesClassesAndRepeatedDegrees(t *testing.T) {
	t.Parallel()
	c := customCatalog("repeat", []interval.Interval{interval.PerfectUnison(), interval.AugmentedUnison(), interval.MinorSecond()}, interval.PerfectOctave())
	o := positionOptions()
	o.Center = spelling.C()
	a, err := scalechoice.AnalyzeSet(notes("C C+1 D-1"), c, o)
	require.NoError(t, err)
	require.Len(t, a.Groups(), 1)
	group := a.Groups()[0]
	assert.Equal(t, []int{0, 1}, classes(group.Tones))
	require.Len(t, group.Readings, 1)
	assert.Equal(t, "C C+1 D-1", nameText(group.Readings[0].Degrees()))
	assert.Equal(t, []int{2, 3}, group.Readings[0].MustDegreesOf(spelling.CSharp()))
	assert.Equal(t, []int{2, 3}, group.Readings[0].MustDegreesOf(spelling.DFlat()))
	assert.Equal(t, "C+1 D-1", nameText(group.Tones[1].Names))
	byName := full()
	byName.Center = spelling.C()
	b, err := scalechoice.AnalyzeSet(notes("C C+1 D-1"), c, byName)
	require.NoError(t, err)
	assert.Equal(t, []int{2}, b.Groups()[0].Readings[0].MustDegreesOf(spelling.CSharp()))
	assert.Equal(t, []int{3}, b.Groups()[0].Readings[0].MustDegreesOf(spelling.DFlat()))
	assert.Empty(t, a.Groups()[0].Readings[0].MustDegreesOf(spelling.D()))
	// Explicit comparison distinguishes spelling from position in the search.
	o = opts()
	o.Center = spelling.CSharp()
	b, err = scalechoice.AnalyzeSet(notes("C+1 F"), catalog("ionian"), o)
	require.NoError(t, err)
	assert.Empty(t, b.Groups())
	o.Comparison = scalechoice.ByPosition
	o.System = system.Traditional()
	b, err = scalechoice.AnalyzeSet(notes("C+1 F"), catalog("ionian"), o)
	require.NoError(t, err)
	require.Len(t, b.Groups(), 1)
	assert.Equal(t, "C+1", b.Groups()[0].Readings[0].Tonic().String())
	for _, tone := range b.Groups()[0].Tones {
		assert.True(t, tone.Class.IsValid())
		require.NotEmpty(t, tone.Names)
		assert.True(t, slices.IsSortedFunc(tone.Names, spelling.Spelling.Compare))
	}
	assert.True(t, slices.IsSorted(classes(b.Groups()[0].Tones)))
}

func TestSetAndPhraseHaveTheSameSearchWithoutMetadata(t *testing.T) {
	t.Parallel()
	input := append(pitched("B A G F E D C C", 5, nil), notes("D")...)
	a, err := scalechoice.AnalyzeSet(input, scalechoice.Classical(), full())
	require.NoError(t, err)
	b, err := scalechoice.AnalyzePhrase(input, scalechoice.Classical(), full())
	require.NoError(t, err)
	assert.Equal(t, allReadingKeys(a), allReadingKeys(b))
	assert.Equal(t, notesText(input), notesText(a.Notes()))
	assert.Equal(t, notesText(input), notesText(b.Notes()))
	checkNotes(t, input, a.Notes())
	checkNotes(t, input, b.Notes())
}

func TestPositionalContinuationsMergeNamesAndFilterByClass(t *testing.T) {
	t.Parallel()
	// All profiles have tonic class C. The first and third describe the same
	// collection {0,2} using C-D and C-C-double-sharp; the other is {0,3}.
	scope := template.MustNewCatalog(
		customCatalog("second", []interval.Interval{interval.PerfectUnison(), interval.MajorSecond()}, interval.PerfectOctave()).Entries()[0],
		customCatalog("third", []interval.Interval{interval.PerfectUnison(), interval.MinorThird()}, interval.PerfectOctave()).Entries()[0],
		customCatalog("augmented-unison", []interval.Interval{interval.PerfectUnison(), interval.MustParse("AA1")}, interval.PerfectOctave()).Entries()[0],
	)
	o := positionOptions()
	o.Condition = search.Containment
	o.Center = spelling.C()
	// B-sharp is a second observed spelling of class zero, not a second tone.
	a, err := scalechoice.AnalyzeSet(notes("C B+1"), scope, o)
	require.NoError(t, err)
	groups := a.Groups()
	require.Len(t, groups, 2)
	assert.Equal(t, []int{0, 2}, classes(groups[0].Tones))
	assert.Equal(t, []int{0, 3}, classes(groups[1].Tones))
	assert.Equal(t, []string{"second/C", "augmented-unison/C"}, readingKeys(groups[0].Readings))
	assert.Equal(t, []string{"third/C"}, readingKeys(groups[1].Readings))
	for _, group := range groups {
		assert.Equal(t, "C B+1", nameText(group.Tones[0].Names))
	}
	assert.Equal(t, []int{2}, classes(groups[0].Missing))
	assert.Equal(t, "C+2 D", nameText(groups[0].Missing[0].Names))
	assert.Equal(t, []int{3}, classes(groups[1].Missing))
	assert.Equal(t, "E-1", nameText(groups[1].Missing[0].Names))
	common := a.Common()
	require.Len(t, common, 1)
	assert.Equal(t, []int{0}, classes(common))
	assert.Equal(t, "C B+1", nameText(common[0].Names))
	continuations := a.Continuations()
	require.Len(t, continuations, 2, "enharmonic names share one continuation")
	assert.Equal(t, 2, continuations[0].Tone.Class.Number())
	assert.Equal(t, "C+2 D", nameText(continuations[0].Tone.Names))
	checkPartition(t, continuations[0].Groups, []int{0}, []int{1})
	assert.Equal(t, 3, continuations[1].Tone.Class.Number())
	assert.Equal(t, "E-1", nameText(continuations[1].Tone.Names))
	checkPartition(t, continuations[1].Groups, []int{1}, []int{0})
	// E-double-flat is absent from the displayed class-2 names, and D-sharp
	// is absent from the class-3 names. Their classes still select the groups.
	for _, tc := range []struct {
		added         string
		kept, dropped []int
	}{
		{"E-2", []int{0}, []int{1}},
		{"D C+2 E-2", []int{0}, []int{1}},
		{"D+1", []int{1}, []int{0}},
		{"D+1 E-1", []int{1}, []int{0}},
		{"E-2 D+1", nil, []int{0, 1}},
		{"B+1 C", []int{0, 1}, nil},
	} {
		partition, err := a.AfterAdding(names(tc.added)...)
		require.NoError(t, err)
		checkPartition(t, partition, tc.kept, tc.dropped)
	}
	assert.Len(t, a.Continuations(), 2, "filtering must not expand or mutate the choices")
}
