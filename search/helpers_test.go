package search_test

import (
	"iter"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/search"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sp(text string) spelling.Spelling {
	machineForm := strings.NewReplacer(
		"𝄪", "+2",
		"𝄫", "-2",
		"♯", "+1",
		"♭", "-1",
	).Replace(text)
	return spelling.MustParse(machineForm)
}

func spellingText(value spelling.Spelling) string {
	text := value.Letter().String()
	num, den := value.Alteration().Fraction()
	if num == 0 {
		return text
	}
	switch {
	case den == 1 && num == 1:
		return text + "♯"
	case den == 1 && num == -1:
		return text + "♭"
	case den == 1 && num == 2:
		return text + "𝄪"
	case den == 1 && num == -2:
		return text + "𝄫"
	}
	if num > 0 {
		text += "+"
	} else {
		text += "-"
		num = -num
	}
	text += strconv.Itoa(num)
	if den > 1 {
		text += "/" + strconv.Itoa(den)
	}
	return text
}

func alt(value int) alteration.Alteration {
	return alteration.MustFromInt(value)
}

func fractional(num, den int) alteration.Alteration {
	return alteration.MustFromFraction(num, den)
}

func notes(names ...string) []note.Note {
	result := make([]note.Note, len(names))
	for i, name := range names {
		result[i] = note.MustFromSpelling(sp(name), cents.Zero(), nil)
	}
	return result
}

func notesIn(source settings.Source, names ...string) []note.Note {
	result := make([]note.Note, len(names))
	for i, name := range names {
		result[i] = note.MustFromSpelling(sp(name), cents.Zero(), source)
	}
	return result
}

func entry(code string, tmpl template.Template, names ...string) template.Entry {
	if len(names) == 0 {
		names = []string{code}
	}
	return template.MustNewEntry(code, names, tmpl)
}

func searcherFor(t testing.TB, entries ...template.Entry) search.Searcher {
	t.Helper()
	catalog := template.MustNewCatalog(entries...)
	searcher, err := search.New(&catalog)
	require.NoError(t, err)
	return searcher
}

func builtInSearcher(t testing.TB, codes ...string) search.Searcher {
	t.Helper()
	catalog := template.BuiltIn()
	entries := make([]template.Entry, len(codes))
	for i, code := range codes {
		entry, ok := catalog.ByCode(code)
		require.True(t, ok, code)
		entries[i] = entry
	}
	return searcherFor(t, entries...)
}

func customSystem(name string, classes, chromatic int, offsets map[letter.Letter]int, rule *system.Rule) *system.System {
	definition := system.Definition{
		Name:        name,
		Classes:     classes,
		Boundary:    letter.C,
		Chromatic:   chromatic,
		Alterations: rule,
	}
	for _, l := range letter.All() {
		offset := offsets[l]
		offset %= classes
		if offset < 0 {
			offset += classes
		}
		definition.Offsets[l.Ordinal()] = offset
	}
	return system.MustNew(definition)
}

func traditionalWithRule(name string, rule *system.Rule) *system.System {
	return customSystem(name, 12, 1, naturalOffsets(), rule)
}

func naturalOffsets() map[letter.Letter]int {
	return map[letter.Letter]int{
		letter.C: 0,
		letter.D: 2,
		letter.E: 4,
		letter.F: 5,
		letter.G: 7,
		letter.A: 9,
		letter.B: 11,
	}
}

func nonRegularSystem() *system.System {
	return customSystem("NR", 12, 1, map[letter.Letter]int{
		letter.C: 0,
		letter.D: 3,
		letter.E: 4,
		letter.F: 5,
		letter.G: 7,
		letter.A: 9,
		letter.B: 11,
	}, system.Integers())
}

func c3System() *system.System {
	return customSystem("C3", 12, 3, map[letter.Letter]int{
		letter.C: 0,
		letter.D: 2,
		letter.E: 3,
		letter.F: 5,
		letter.G: 7,
		letter.A: 9,
		letter.B: 10,
	}, system.Integers())
}

func naturalOnlySystem() *system.System {
	return traditionalWithRule("T-natural-only", system.Single(alteration.Natural()))
}

func onePlaceEntry(code string) template.Entry {
	tmpl := template.MustClosedFromPlaceIntervals(
		[]interval.Interval{interval.PerfectUnison()},
		interval.MustParse("P8"),
	)
	return entry(code, tmpl)
}

func onePlaceSystem(name string, classes, chromatic int, rule *system.Rule) *system.System {
	return customSystem(name, classes, chromatic, naturalOffsets(), rule)
}

type countedSource struct {
	current atomic.Pointer[settings.Settings]
	reads   atomic.Int64
}

func newCountedSource(set *settings.Settings) *countedSource {
	source := &countedSource{}
	source.current.Store(set)
	return source
}

func (s *countedSource) Settings() *settings.Settings {
	s.reads.Add(1)
	return s.current.Load()
}

func (s *countedSource) resetReads() {
	s.reads.Store(0)
}

func (s *countedSource) readCount() int {
	return int(s.reads.Load())
}

func settingsWithSystem(sys *system.System) *settings.Settings {
	return settings.Standard().MustWithSystem(sys)
}

type spellingPair struct {
	Code  string
	Tonic string
}

func spellingPairs(matches []search.SpellingMatch) []spellingPair {
	result := make([]spellingPair, len(matches))
	for i, match := range matches {
		origin, ok := match.Mode().Origin()
		if !ok {
			panic("spelling match without origin")
		}
		result[i] = spellingPair{Code: origin.Code(), Tonic: spellingText(match.Tonic())}
	}
	return result
}

type positionPair struct {
	Code   string
	Class  int
	Letter string
	Tonic  string
}

func positionPairs(matches []search.PositionMatch) []positionPair {
	result := make([]positionPair, len(matches))
	for i, match := range matches {
		origin, ok := match.Mode().Origin()
		if !ok {
			panic("position match without origin")
		}
		l, hasLetter := match.Family().Letter()
		letterName := ""
		if hasLetter {
			letterName = l.String()
		}
		result[i] = positionPair{
			Code:   origin.Code(),
			Class:  match.Family().Class().Number(),
			Letter: letterName,
			Tonic:  spellingText(match.Tonic()),
		}
	}
	return result
}

func classNumbers(classes []pitchclass.Class) []int {
	result := make([]int, len(classes))
	for i, class := range classes {
		result[i] = class.Number()
	}
	return result
}

func classSet(matches []search.PositionMatch) []int {
	seen := make(map[int]struct{})
	for _, match := range matches {
		seen[match.Family().Class().Number()] = struct{}{}
	}
	result := make([]int, 0, len(seen))
	for number := range seen {
		result = append(result, number)
	}
	slices.Sort(result)
	return result
}

func assertSpellingMatch(t *testing.T, match search.SpellingMatch) {
	t.Helper()
	require.True(t, match.IsValid())
	built, err := match.Mode().NotesFrom(note.MustFromSpelling(match.Tonic(), cents.Zero(), nil))
	require.NoError(t, err)
	require.Len(t, built, match.Mode().Len())
	for _, input := range match.InputNames() {
		expected := make([]int, 0)
		for i, current := range built {
			if current.Spelling().Equal(input) {
				expected = append(expected, i+1)
			}
		}
		assert.NotEmpty(t, expected)
		assert.Equal(t, expected, match.DegreesFor(input))
	}
}

func assertPositionMatch(t *testing.T, match search.PositionMatch) {
	t.Helper()
	require.True(t, match.IsValid())
	sys := match.Family().System()
	require.NotNil(t, sys)
	assert.True(t, sys.Equal(sys))
	built, err := match.Mode().NotesFromIn(sys, note.MustFromSpelling(match.Tonic(), cents.Zero(), nil))
	require.NoError(t, err)
	require.Len(t, built, match.Mode().Len())
	for _, input := range match.InputClasses() {
		expected := make([]int, 0)
		for i, current := range built {
			class, classErr := sys.Class(current.Spelling())
			require.NoError(t, classErr)
			if class.Equal(input) {
				expected = append(expected, i+1)
			}
		}
		assert.NotEmpty(t, expected)
		assert.Equal(t, expected, match.DegreesForClass(input))
	}
}

func collectSpellings(seq iter.Seq[spelling.Spelling], maximum int) []spelling.Spelling {
	if seq == nil {
		return nil
	}
	result := make([]spelling.Spelling, 0)
	for value := range seq {
		result = append(result, value)
		if maximum >= 0 && len(result) == maximum {
			break
		}
	}
	return result
}

func spellingStrings(values []spelling.Spelling) []string {
	result := make([]string, len(values))
	for i, value := range values {
		result[i] = spellingText(value)
	}
	return result
}

func requireMustPanic(t *testing.T, expected error, call func()) {
	t.Helper()
	require.NotNil(t, expected)
	defer func() {
		recovered := recover()
		require.NotNil(t, recovered)
		err, ok := recovered.(error)
		require.True(t, ok)
		assert.EqualError(t, err, expected.Error())
	}()
	call()
}

func requireError(t *testing.T, err error, expected ...error) {
	t.Helper()
	require.Error(t, err)
	for _, target := range expected {
		assert.ErrorIs(t, err, target)
	}
}

func requireNonNilEmpty[T any](t *testing.T, values []T, err error) {
	t.Helper()
	require.NoError(t, err)
	require.NotNil(t, values)
	assert.Empty(t, values)
}

func independentSpellingMembership(t *testing.T, searcher search.Searcher, input []note.Note, condition search.Condition) {
	t.Helper()
	matches, err := searcher.BySpelling(input, condition)
	require.NoError(t, err)
	for _, match := range matches {
		assertSpellingMatch(t, match)
	}
	full, err := searcher.BySpelling(input, search.FullMatch)
	require.NoError(t, err)
	containment, err := searcher.BySpelling(input, search.Containment)
	require.NoError(t, err)
	assert.Subset(t, spellingPairs(containment), spellingPairs(full))
}

func independentPositionMembership(t *testing.T, searcher search.Searcher, sys *system.System, input []note.Note) {
	t.Helper()
	matches, err := searcher.ByPosition(sys, input, search.Containment)
	require.NoError(t, err)
	for _, match := range matches {
		assertPositionMatch(t, match)
	}
	full, err := searcher.ByPosition(sys, input, search.FullMatch)
	require.NoError(t, err)
	assert.Subset(t, positionPairs(matches), positionPairs(full))
	for _, match := range full {
		assert.Contains(t, classNumbers(match.InputClasses()), match.Family().Class().Number())
	}
}

func assertPositionOrder(t *testing.T, matches []search.PositionMatch) {
	t.Helper()
	for i := 1; i < len(matches); i++ {
		previous := matches[i-1].Family()
		current := matches[i].Family()
		previousClass := previous.Class().Number()
		currentClass := current.Class().Number()
		if !assert.LessOrEqual(t, previousClass, currentClass) {
			continue
		}
		if previousClass != currentClass {
			continue
		}
		previousLetter, previousOK := previous.Letter()
		currentLetter, currentOK := current.Letter()
		if previousOK && currentOK {
			assert.Less(t, previousLetter.Ordinal(), currentLetter.Ordinal())
		}
	}
}

func assertNotesUnchanged(t *testing.T, expected, actual []note.Note) {
	t.Helper()
	require.Len(t, actual, len(expected))
	for i := range expected {
		want := expected[i]
		got := actual[i]
		require.True(t, got.IsValid())
		assert.True(t, got.Spelling().Equal(want.Spelling()), "note %d spelling", i)
		assert.Equal(t, want.HasOctave(), got.HasOctave(), "note %d octave presence", i)
		if want.HasOctave() {
			wantASPN, err := want.ASPN()
			require.NoError(t, err)
			gotASPN, err := got.ASPN()
			require.NoError(t, err)
			assert.Equal(t, wantASPN, gotASPN, "note %d octave", i)
		}
		assert.True(t, got.Cents().Equal(want.Cents()), "note %d cents", i)
		assert.Equal(t, want.IsBound(), got.IsBound(), "note %d source presence", i)
		if want.IsBound() {
			assert.True(t, got.SameSource(want), "note %d source identity", i)
		}
	}
}
