package scalechoice_test

import (
	"math"
	"slices"
	"sync"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/problemsolving/scalechoice"
	"github.com/go-muse/musex/search"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalysisInputErrorsBeforeMusicalAbsence(t *testing.T) {
	t.Parallel()
	invalidScope := template.MustNewCatalog(template.MustNewEntry("open", []string{"open"}, template.MustOpenFromIntervals([]interval.Interval{interval.MajorSecond()})))
	nonOctave := customCatalog("two-octaves", []interval.Interval{interval.PerfectUnison()}, interval.PerfectFifteenth())
	// The complete written cycle is P8, but its third place has deviation
	// 2*M-1 and cannot initialize a Mode. Catalog validation must catch it.
	magnitude := alteration.MaxMagnitude
	wide := template.MustClosedFromIntervals([]interval.Interval{
		interval.MustFromDeviation(2, alteration.MustFromInt(magnitude-1)),
		interval.MustFromDeviation(2, alteration.MustFromInt(magnitude)),
		interval.MustFromDeviation(2, alteration.MustFromInt(-magnitude)),
		interval.MustFromDeviation(5, alteration.MustFromInt(-magnitude)),
	})
	require.True(t, wide.IsOctave())
	unrepresentable := template.MustNewCatalog(template.MustNewEntry("wide", []string{"wide"}, wide))
	broken := source(system.Traditional())
	badNote := pitched("C", 4, broken)[0]
	broken.current.Store(nil)
	broken.reset()
	half := note.MustFromSpelling(halfFlatE(), cents.Zero(), nil)
	position := positionOptions()
	noSystem := position
	noSystem.System = nil
	invalidSystem := opts()
	invalidSystem.System = new(system.System)
	invalidPositionSystem := positionOptions()
	invalidPositionSystem.System = new(system.System)
	invalidComparison := opts()
	invalidComparison.Comparison = 99
	noComparison := opts()
	noComparison.Comparison = 0
	invalidCondition := opts()
	invalidCondition.Condition = search.Condition(99)
	noCondition := opts()
	noCondition.Condition = 0
	invalidCenter := position
	invalidCenter.Center = halfFlatE()
	for _, tc := range []struct {
		name        string
		input       []note.Note
		scope       template.Catalog
		o           scalechoice.Options
		want, cause error
	}{
		{"nil input", nil, template.Catalog{}, opts(), scalechoice.ErrEmptyInput, nil},
		{"empty input", []note.Note{}, scalechoice.Classical(), opts(), scalechoice.ErrEmptyInput, nil},
		{"unset note", []note.Note{{}}, template.Catalog{}, opts(), scalechoice.ErrUnset, nil},
		{"open scope", notes("C"), invalidScope, opts(), scalechoice.ErrInvalidArgument, nil},
		{"non octave", notes("C"), nonOctave, opts(), scalechoice.ErrInvalidArgument, nil},
		{"unrepresentable mode", notes("C"), unrepresentable, opts(), scalechoice.ErrInvalidArgument, nil},
		{"invalid explicit system", notes("C"), template.Catalog{}, invalidSystem, scalechoice.ErrUnset, system.ErrInvalid},
		{"invalid positional explicit system", notes("C"), template.Catalog{}, invalidPositionSystem, scalechoice.ErrUnset, system.ErrInvalid},
		{"comparison unset", notes("C"), template.Catalog{}, noComparison, scalechoice.ErrInvalidArgument, nil},
		{"comparison unknown", notes("C"), template.Catalog{}, invalidComparison, scalechoice.ErrInvalidArgument, nil},
		{"condition unset", notes("C"), template.Catalog{}, noCondition, scalechoice.ErrInvalidArgument, nil},
		{"condition unknown", notes("C"), template.Catalog{}, invalidCondition, scalechoice.ErrInvalidArgument, nil},
		{"context missing", notes("C"), template.Catalog{}, noSystem, scalechoice.ErrNoSystem, nil},
		{"different contexts", append(pitched("C", 4, source(system.Traditional())), pitched("D", 4, source(system.Positions19()))...), template.Catalog{}, noSystem, scalechoice.ErrDifferentSystems, nil},
		{"broken context", []note.Note{badNote, badNote}, template.Catalog{}, noSystem, scalechoice.ErrContext, note.ErrUnsetSettings},
		{"inadmissible input", []note.Note{half}, template.Catalog{}, position, scalechoice.ErrUninterpretableName, system.ErrInadmissible},
		{"inadmissible center", notes("C"), template.Catalog{}, invalidCenter, scalechoice.ErrUninterpretableName, system.ErrInadmissible},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			for _, analyze := range []func([]note.Note, template.Catalog, scalechoice.Options) (scalechoice.Analysis, error){scalechoice.AnalyzeSet, scalechoice.AnalyzePhrase} {
				a, err := analyze(tc.input, tc.scope, tc.o)
				assert.ErrorIs(t, err, tc.want)
				if tc.cause != nil {
					assert.ErrorIs(t, err, tc.cause)
				}
				assert.False(t, a.IsValid())
				assert.Empty(t, a.Notes())
				assert.Empty(t, a.Groups())
			}
		})
	}
}

func TestUnsetAndOperationErrorsAndMust(t *testing.T) {
	t.Parallel()
	var zero scalechoice.Analysis
	assert.False(t, zero.IsValid())
	assert.False(t, zero.IsPhrase())
	assert.Equal(t, scalechoice.Options{}, zero.Options())
	assert.Empty(t, zero.Scope().Entries())
	assert.Empty(t, zero.Notes())
	assert.Empty(t, zero.Groups())
	assert.Empty(t, zero.Common())
	assert.Empty(t, zero.Continuations())
	_, err := zero.AfterAdding(spelling.C())
	assert.ErrorIs(t, err, scalechoice.ErrUnset)
	ch, err := zero.ReplaceAll(spelling.C(), spelling.D())
	assert.ErrorIs(t, err, scalechoice.ErrUnset)
	assert.False(t, ch.After.IsValid())
	assert.Empty(t, ch.Changed)
	ch, err = zero.ReplaceAt(0, notes("C")[0])
	assert.ErrorIs(t, err, scalechoice.ErrUnset)
	assert.False(t, ch.After.IsValid())
	var r scalechoice.Reading
	assert.False(t, r.Entry().IsValid())
	assert.False(t, r.Tonic().IsValid())
	assert.Empty(t, r.Degrees())
	_, err = r.DegreesOf(spelling.C())
	assert.ErrorIs(t, err, scalechoice.ErrUnset)
	mustPanic(t, scalechoice.ErrUnset, func() { zero.MustAfterAdding(spelling.C()) })
	mustPanic(t, scalechoice.ErrUnset, func() { zero.MustReplaceAll(spelling.C(), spelling.D()) })
	mustPanic(t, scalechoice.ErrUnset, func() { zero.MustReplaceAt(0, notes("C")[0]) })
	mustPanic(t, scalechoice.ErrUnset, func() { r.MustDegreesOf(spelling.C()) })
	mustPanic(t, scalechoice.ErrEmptyInput, func() { scalechoice.MustAnalyzeSet(nil, scalechoice.Classical(), opts()) })
	mustPanic(t, scalechoice.ErrEmptyInput, func() { scalechoice.MustAnalyzePhrase(nil, scalechoice.Classical(), opts()) })
	a := scalechoice.MustAnalyzePhrase(pitched("C D", 4, nil), template.Catalog{}, opts())
	before := describe(a)
	for _, index := range []int{-1, 2, math.MaxInt} {
		ch, err = a.ReplaceAt(index, notes("E")[0])
		assert.ErrorIs(t, err, scalechoice.ErrInvalidArgument)
		assert.False(t, ch.After.IsValid())
	}
	ch, err = a.ReplaceAt(0, note.Note{})
	assert.ErrorIs(t, err, scalechoice.ErrUnset)
	assert.False(t, ch.After.IsValid())
	_, err = a.AfterAdding(spelling.Spelling{})
	assert.ErrorIs(t, err, scalechoice.ErrUnset)
	ch, err = a.ReplaceAll(spelling.Spelling{}, spelling.D())
	assert.ErrorIs(t, err, scalechoice.ErrUnset)
	assert.False(t, ch.After.IsValid())
	ch, err = a.ReplaceAll(spelling.C(), spelling.Spelling{})
	assert.ErrorIs(t, err, scalechoice.ErrUnset)
	assert.False(t, ch.After.IsValid())
	set := scalechoice.MustAnalyzeSet(notes("C"), template.Catalog{}, opts())
	_, err = set.ReplaceAt(0, notes("D")[0])
	assert.ErrorIs(t, err, scalechoice.ErrInvalidArgument)
	mustPanic(t, scalechoice.ErrInvalidArgument, func() { a.MustReplaceAt(2, notes("E")[0]) })
	assert.Equal(t, before, describe(a))
	o := positionOptions()
	empty := scalechoice.MustAnalyzePhrase(notes("C"), template.Catalog{}, o)
	_, err = empty.AfterAdding(halfFlatE())
	assert.ErrorIs(t, err, scalechoice.ErrUninterpretableName)
	assert.ErrorIs(t, err, system.ErrInadmissible)
	ch, err = empty.ReplaceAll(spelling.D(), halfFlatE())
	assert.ErrorIs(t, err, scalechoice.ErrUninterpretableName)
	assert.False(t, ch.After.IsValid())
	ch, err = empty.ReplaceAt(0, note.MustFromSpelling(halfFlatE(), cents.Zero(), nil))
	assert.ErrorIs(t, err, scalechoice.ErrUninterpretableName)
	assert.False(t, ch.After.IsValid())
	found := scalechoice.MustAnalyzeSet(notes("C D E F G A B"), scalechoice.Classical(), full())
	_, err = found.Groups()[0].Readings[0].DegreesOf(spelling.Spelling{})
	assert.ErrorIs(t, err, scalechoice.ErrUnset)
}

func TestIndependentResultsOwnEveryNestedSlice(t *testing.T) {
	t.Parallel()
	input := notes("C D E G A")
	a := scalechoice.MustAnalyzePhrase(input, scalechoice.Classical(), opts())
	baseline := describe(a)
	input[0] = notes("C+1")[0]
	ns := a.Notes()
	ns[0] = notes("D+1")[0]
	gs := a.Groups()
	gs[0].Tones[0].Names[0] = spelling.CSharp()
	gs[0].Missing[0].Names[0] = spelling.DSharp()
	gs[0].Readings[0] = scalechoice.Reading{}
	common := a.Common()
	common[0].Names[0] = spelling.CSharp()
	cs := a.Continuations()
	cs[0].Tone.Names[0] = spelling.DSharp()
	cs[0].Groups.Kept[0] = 999
	cs[0].Groups.Dropped[0] = 999
	p := a.MustAfterAdding(spelling.F())
	p.Kept[0] = 999
	p.Dropped[0] = 999
	r := a.Groups()[0].Readings[0]
	ds := r.Degrees()
	ds[0] = spelling.CSharp()
	numbers := r.MustDegreesOf(spelling.C())
	numbers[0] = 999
	entries := a.Scope().Entries()
	entries[0] = template.Entry{}
	assert.Equal(t, baseline, describe(a))
	assert.Equal(t, "C D E G A", notesText(a.Notes()))
	assert.Equal(t, "C D E F G A B", nameText(r.Degrees()))
	assert.Equal(t, []int{1}, r.MustDegreesOf(spelling.C()))
	checkPartition(t, a.MustAfterAdding(spelling.F()), []int{0, 1}, []int{2})
	freshGroups := a.Groups()
	require.Len(t, freshGroups, 3)
	assert.Equal(t, "F B", toneText(freshGroups[0].Missing))
	assert.Equal(t, "F B-1", toneText(freshGroups[1].Missing))
	assert.Equal(t, "F+1 B", toneText(freshGroups[2].Missing))
	freshContinuations := a.Continuations()
	require.Len(t, freshContinuations, 4)
	for index, expected := range []struct {
		name          string
		kept, dropped []int
	}{
		{"F", []int{0, 1}, []int{2}},
		{"F+1", []int{2}, []int{0, 1}},
		{"B-1", []int{1}, []int{0, 2}},
		{"B", []int{0, 2}, []int{1}},
	} {
		assert.Equal(t, expected.name, nameText(freshContinuations[index].Tone.Names))
		checkPartition(t, freshContinuations[index].Groups, expected.kept, expected.dropped)
	}
	assert.Equal(t, []string{"ionian", "dorian", "phrygian", "lydian", "mixolydian", "aeolian", "locrian"}, profileCodes(a.Scope()))
	// Change result slices also belong to that one returned value.
	o := full()
	o.System = system.Traditional()
	o.Center = spelling.C()
	line := scalechoice.MustAnalyzePhrase(pitched("C D E F G A B", 4, nil), scalechoice.Classical(), o)
	ch := line.MustReplaceAll(spelling.F(), spelling.FSharp())
	afterBaseline := describe(ch.After)
	ch.Matched[0] = 999
	ch.Changed[0] = 999
	ch.Tones.Kept[0].Names[0] = spelling.CSharp()
	ch.Groups.Removed[0] = 999
	ch.Groups.Added[0] = 999
	ch.Readings.Added[0] = scalechoice.Reading{}
	ch.Neighbors[0].From = 999
	assert.Equal(t, afterBaseline, describe(ch.After))
	again := line.MustReplaceAll(spelling.F(), spelling.FSharp())
	assert.Equal(t, []int{3}, again.Matched)
	assert.Equal(t, []int{3}, again.Changed)
	assert.Equal(t, "C D E G A B", toneText(again.Tones.Kept))
	assert.Equal(t, "F", toneText(again.Tones.Removed))
	assert.Equal(t, "F+1", toneText(again.Tones.Added))
	assert.Empty(t, again.Groups.Kept)
	assert.Equal(t, []int{0}, again.Groups.Removed)
	assert.Equal(t, []int{0}, again.Groups.Added)
	assert.Empty(t, again.Readings.Kept)
	assert.Equal(t, []string{"ionian/C"}, readingKeys(again.Readings.Removed))
	assert.Equal(t, []string{"lydian/C"}, readingKeys(again.Readings.Added))
	require.Len(t, again.Neighbors, 2)
	assert.Equal(t, 2, again.Neighbors[0].From)
	assert.Equal(t, 3, again.Neighbors[0].To)
	assert.Equal(t, 3, again.Neighbors[1].From)
	assert.Equal(t, 4, again.Neighbors[1].To)
	checkMotion(t, again.Neighbors[0].Before, "m2", 1, o.System)
	checkMotion(t, again.Neighbors[0].After, "M2", 2, o.System)
	checkMotion(t, again.Neighbors[1].Before, "M2", 2, o.System)
	checkMotion(t, again.Neighbors[1].After, "m2", 1, o.System)
}

func TestConcurrentReadsAndIndependentTrialBranches(t *testing.T) {
	t.Parallel()
	o := full()
	o.System = system.Traditional()
	o.Center = spelling.C()
	a := scalechoice.MustAnalyzePhrase(pitched("C D E F G A B", 4, nil), scalechoice.Classical(), o)
	before := describe(a)
	results := make([]scalechoice.Change, 12)
	partitions := make([]scalechoice.Partition, 12)
	errs := make([]error, 12)
	partitionErrors := make([]error, 12)
	var wg sync.WaitGroup
	for worker := range results {
		wg.Go(func() {
			results[worker], errs[worker] = a.ReplaceAll(spelling.F(), spelling.FSharp())
			partitions[worker], partitionErrors[worker] = a.AfterAdding(spelling.C())
			_ = a.Groups()
			_ = a.Common()
			_ = a.Continuations()
			_ = a.Notes()
			_ = a.Options()
			_ = a.Scope()
		})
	}
	wg.Wait()
	for index, ch := range results {
		require.NoError(t, errs[index])
		require.NoError(t, partitionErrors[index])
		checkPartition(t, partitions[index], []int{0}, nil)
		assert.Equal(t, []string{"lydian/C"}, allReadingKeys(ch.After))
		assert.Equal(t, []int{3}, ch.Changed)
		assert.Equal(t, []int{0}, ch.Groups.Removed)
	}
	assert.Equal(t, before, describe(a))
	// Copying the analysis shares only immutable state, not branch mutation.
	copy := a
	sharp := copy.MustReplaceAt(3, pitched("F+1", 4, nil)[0])
	flat := a.MustReplaceAt(3, pitched("F-1", 4, nil)[0])
	assert.Equal(t, "C D E F+1 G A B", notesText(sharp.After.Notes()))
	assert.Equal(t, "C D E F-1 G A B", notesText(flat.After.Notes()))
	assert.Equal(t, before, describe(a))
	assert.True(t, slices.Equal(profileCodes(a.Scope()), profileCodes(copy.Scope())))
}
