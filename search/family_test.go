package search_test

import (
	"slices"
	"sync"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/search"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func onePlaceMatch(t *testing.T, sys *system.System, input ...string) search.PositionMatch {
	t.Helper()
	searcher := searcherFor(t, onePlaceEntry("one"))
	matches, err := searcher.ByPosition(sys, notes(input...), search.FullMatch)
	require.NoError(t, err)
	require.Len(t, matches, 1)
	assertPositionMatch(t, matches[0])
	return matches[0]
}

func TestRepresentativeUsesMinimumAbsoluteAlterationBeforeExactMatch(t *testing.T) {
	t.Parallel()
	rule := system.MustUnion(system.Integers(), system.Single(alt(2)))
	sys := onePlaceSystem("minimum-first", 1, 1, rule)
	match := onePlaceMatch(t, sys, "C𝄪")
	assert.Equal(t, "C", spellingText(match.Tonic()))
	assert.True(t, match.Tonic().Alteration().IsNatural())
}

func TestRepresentativeUsesExactInputMatchOnAlterationTie(t *testing.T) {
	t.Parallel()
	// The +2 and -2 deviations reject a natural tonic. Alterations +1 and -1
	// both fit, leaving C♯ and D♭ at the same minimum in the class-1 family.
	rule := system.MustUnion(
		system.Single(alt(-3)),
		system.Single(alt(-1)),
		system.Single(alteration.Natural()),
		system.Single(alt(1)),
		system.Single(alt(3)),
	)
	sys := traditionalWithRule("exact-match-tie", rule)
	tmpl := template.MustOpenFromPlaceIntervals([]interval.Interval{
		interval.PerfectUnison(),
		interval.MustParse("AA2"),
		interval.MustParse("dd5"),
	})
	searcher := searcherFor(t, entry("exact-match-tie", tmpl))
	matches, err := searcher.ByPosition(sys, notes("C♯"), search.Containment)
	require.NoError(t, err)
	index := slices.IndexFunc(matches, func(match search.PositionMatch) bool {
		return match.Family().Class().Number() == 1
	})
	if !assert.NotEqual(t, -1, index) {
		return
	}
	match := matches[index]
	assert.Equal(t, "C♯", spellingText(match.Tonic()))
	assertPositionMatch(t, match)
	assert.Equal(t, []string{"C♯", "D♭"}, spellingStrings(collectSpellings(match.Family().MustSpellings(alt(1)), -1)))
}

func TestRepresentativeUsesSpellingOrderAfterOtherTies(t *testing.T) {
	t.Parallel()
	sys := onePlaceSystem("spelling-tie", 1, 1, system.Integers())
	match := onePlaceMatch(t, sys, "D")
	assert.Equal(t, "D", spellingText(match.Tonic()))

	// With no compatible exact spelling, canonical letter order resolves the tie.
	match = onePlaceMatch(t, sys, "C♯")
	assert.Equal(t, "C", spellingText(match.Tonic()))
}

func TestRepresentativeWithNonZeroMinimum(t *testing.T) {
	t.Parallel()
	rule := system.MustIntegerRange(0, 1)
	sys := traditionalWithRule("nonzero-minimum", rule)
	tmpl := template.MustOpenFromIntervals([]interval.Interval{interval.MustParse("m2")})
	searcher := searcherFor(t, entry("nonzero-minimum", tmpl))
	matches, err := searcher.ByPosition(sys, notes("D"), search.Containment)
	require.NoError(t, err)
	index := slices.IndexFunc(matches, func(match search.PositionMatch) bool {
		return match.Family().Class().Number() == 1
	})
	if !assert.NotEqual(t, -1, index) {
		return
	}
	match := matches[index]
	assert.Equal(t, "C♯", spellingText(match.Tonic()))
	assert.True(t, match.Tonic().Alteration().Equal(alt(1)))
	assertPositionMatch(t, match)

	seq, err := match.Family().Spellings(alteration.Natural())
	require.NoError(t, err)
	require.NotNil(t, seq)
	assert.Empty(t, collectSpellings(seq, -1))
	values := collectSpellings(match.Family().MustSpellings(alt(1)), -1)
	assert.Equal(t, []string{"C♯"}, spellingStrings(values))
}

func TestSpellingsZeroAndFractionalLimits(t *testing.T) {
	t.Parallel()
	sys := onePlaceSystem("fractional", 1, 2, system.MustMultiples(2))
	match := onePlaceMatch(t, sys, "C+1/2")
	assert.Equal(t, "C", spellingText(match.Tonic()))

	zero := collectSpellings(match.Family().MustSpellings(alteration.Natural()), -1)
	assert.Equal(t, []string{"C", "D", "E", "F", "G", "A", "B"}, spellingStrings(zero))
	assert.Contains(t, spellingStrings(zero), spellingText(match.Tonic()))

	fractionalValues := collectSpellings(match.Family().MustSpellings(fractional(1, 2)), -1)
	require.Len(t, fractionalValues, 21)
	assert.Contains(t, spellingStrings(fractionalValues), "C+1/2")
	assert.Contains(t, spellingStrings(fractionalValues), "C-1/2")
	assert.True(t, slices.IsSortedFunc(fractionalValues, func(a, b spelling.Spelling) int { return a.Compare(b) }))
	assertUniqueSpellings(t, fractionalValues)
}

func TestSpellingsBoundedOverlappingAndSingleProgressions(t *testing.T) {
	t.Parallel()
	bounded := system.MustNewProgression(alt(-1), alt(1)).MustWithUpper(alt(1))
	single := system.MustNewProgression(alteration.Natural(), alteration.Natural())
	rule := system.MustNewRule(bounded, single)
	sys := onePlaceSystem("bounded-overlap", 1, 1, rule)
	match := onePlaceMatch(t, sys, "C")
	assert.Equal(t, "C", spellingText(match.Tonic()))

	values := collectSpellings(match.Family().MustSpellings(alt(1)), -1)
	require.Len(t, values, 21)
	assertUniqueSpellings(t, values)
	assert.Contains(t, spellingStrings(values), "C")
	assert.Contains(t, spellingStrings(values), "C♯")
	assert.Contains(t, spellingStrings(values), "C♭")
	assert.True(t, slices.IsSortedFunc(values, func(a, b spelling.Spelling) int { return a.Compare(b) }))
}

func TestLargeFractionalRuleRepresentative(t *testing.T) {
	t.Parallel()
	sys := onePlaceSystem("large-fractional", 1, alteration.MaxDenominator, system.MustMultiples(alteration.MaxDenominator))
	match := onePlaceMatch(t, sys, "C")
	assert.Equal(t, "C", spellingText(match.Tonic()))
	assert.True(t, match.Tonic().Alteration().IsNatural())
	assert.Equal(t, 0, match.Family().Class().Number())

	zero := collectSpellings(match.Family().MustSpellings(alteration.Natural()), -1)
	assert.Len(t, zero, 7)
	tiny := collectSpellings(match.Family().MustSpellings(fractional(1, alteration.MaxDenominator)), -1)
	assert.Len(t, tiny, 21)
	assertUniqueSpellings(t, tiny)
}

func TestAddressedLargeParameters(t *testing.T) {
	t.Parallel()
	t.Run("large classes chromatic unit and denominator", func(t *testing.T) {
		t.Parallel()
		sys := onePlaceSystem("large-parameters", alteration.MaxDenominator, alteration.MaxDenominator, system.MustMultiples(alteration.MaxDenominator))
		searcher := searcherFor(t, onePlaceEntry("one"))
		matches, err := searcher.ByPosition(sys, notes("C"), search.FullMatch)
		require.NoError(t, err)
		assert.Equal(t, []positionPair{
			{Code: "one", Class: 0, Letter: "C", Tonic: "C"},
			{Code: "one", Class: 0, Letter: "D", Tonic: "D-1/32768"},
			{Code: "one", Class: 0, Letter: "E", Tonic: "E-1/16384"},
			{Code: "one", Class: 0, Letter: "F", Tonic: "F-5/65536"},
			{Code: "one", Class: 0, Letter: "G", Tonic: "G-7/65536"},
			{Code: "one", Class: 0, Letter: "A", Tonic: "A-9/65536"},
			{Code: "one", Class: 0, Letter: "B", Tonic: "B-11/65536"},
		}, positionPairs(matches))
		for _, match := range matches {
			assert.Equal(t, alteration.MaxDenominator, match.Family().Class().Modulus())
			assertPositionMatch(t, match)
		}
	})
	t.Run("many degrees", func(t *testing.T) {
		t.Parallel()
		adjacent := make([]interval.Interval, 255)
		for i := range adjacent {
			adjacent[i] = interval.MustParse("A1")
		}
		tmpl := template.MustOpenFromIntervals(adjacent)
		searcher := searcherFor(t, entry("many-degrees", tmpl))
		sys := onePlaceSystem("many-degrees-system", 1, 1, system.Integers())
		matches, err := searcher.ByPosition(sys, notes("C"), search.Containment)
		require.NoError(t, err)
		require.Len(t, matches, 1)
		assert.Equal(t, 256, matches[0].Mode().Len())
		assertPositionMatch(t, matches[0])
	})
}

func TestRepresentableResultWithUnrepresentableIntermediate(t *testing.T) {
	t.Parallel()
	// From E, a size-2 interval has natural deviation -1. With d=A_max,
	// δ=d-d0=A_max+1 is unrepresentable, but tonic alteration -1 yields
	// the representable final alteration A_max on F.
	extremeInterval := interval.MustFromDeviation(2, alt(alteration.MaxMagnitude))
	tmpl := template.MustOpenFromIntervals([]interval.Interval{extremeInterval})
	rule := system.MustUnion(
		system.Single(alt(-1)),
		system.Single(alteration.Natural()),
		system.Single(alt(alteration.MaxMagnitude)),
	)
	sys := traditionalWithRule("extended-intermediate", rule)
	inputName := spelling.MustNew(sp("F").Letter(), alt(alteration.MaxMagnitude))
	input := []note.Note{note.MustFromSpelling(inputName, cents.Zero(), nil)}
	searcher := searcherFor(t, entry("extended-intermediate", tmpl))
	matches, err := searcher.ByPosition(sys, input, search.Containment)
	require.NoError(t, err)
	index := slices.IndexFunc(matches, func(match search.PositionMatch) bool {
		return match.Family().Class().Number() == 3
	})
	if !assert.NotEqual(t, -1, index, "missing E-flat family") {
		return
	}
	match := matches[index]
	assert.Equal(t, "E♭", spellingText(match.Tonic()))
	assert.Equal(t, []int{2}, match.DegreesForClass(sys.MustClass(inputName)))
	assertPositionMatch(t, match)
}

func TestSpellingsSequenceEarlyStopAndRepeatedTraversal(t *testing.T) {
	t.Parallel()
	match := onePlaceMatch(t, system.Traditional(), "C")
	seq, err := match.Family().Spellings(alt(3))
	require.NoError(t, err)
	require.NotNil(t, seq)
	expected := spellingStrings(collectSpellings(seq, -1))
	require.NotEmpty(t, expected)
	for range 20 {
		assert.Len(t, collectSpellings(seq, 1), 1)
	}
	assert.Equal(t, expected, spellingStrings(collectSpellings(seq, -1)))
	assert.Equal(t, expected, spellingStrings(collectSpellings(seq, -1)))
}

func TestSpellingsSequenceConcurrentTraversal(t *testing.T) {
	t.Parallel()
	match := onePlaceMatch(t, system.Traditional(), "C")
	seq := match.Family().MustSpellings(alt(2))
	expected := spellingStrings(collectSpellings(seq, -1))
	require.NotEmpty(t, expected)

	const workers = 8
	results := make(chan []string, workers)
	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			results <- spellingStrings(collectSpellings(seq, -1))
		})
	}
	wg.Wait()
	close(results)
	for result := range results {
		assert.Equal(t, expected, result)
	}
}

func TestConcurrentFamiliesRemainIndependent(t *testing.T) {
	t.Parallel()
	searcher := builtInSearcher(t, "ionian")
	matches := searcher.MustByPosition(system.Traditional(), notes("C♯", "F"), search.Containment)
	require.Len(t, matches, 3)
	expected := make([][]string, len(matches))
	for i, match := range matches {
		expected[i] = spellingStrings(collectSpellings(match.Family().MustSpellings(alt(1)), -1))
	}
	var wg sync.WaitGroup
	for i, match := range matches {
		wg.Go(func() {
			actual := spellingStrings(collectSpellings(match.Family().MustSpellings(alt(1)), -1))
			assert.Equal(t, expected[i], actual)
		})
	}
	wg.Wait()
}

func assertUniqueSpellings(t *testing.T, values []spelling.Spelling) {
	t.Helper()
	for i := range values {
		for j := i + 1; j < len(values); j++ {
			assert.False(t, values[i].Equal(values[j]), "duplicate %s", values[i])
		}
	}
}
