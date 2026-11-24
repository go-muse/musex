package mode_test

import (
	"math"
	"slices"
	"testing"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests of the catalog search by numbers of adjacent intervals (issue #146,
// item 2; business-spec, KP-16 and F-13).

// customRepository is the catalog K of KP-16: a and c are open, b and d are
// closed; a and b have the same intervals m3 M2.
func customRepository() mode.Repository {
	c := template.MustNewCatalog(
		entry("a", "m3 M2 |", "Open minor third"),
		entry("b", "m3 M2", "Closed minor third"),
		entry("c", "A2 d3 |", "Open augmented second"),
		entry("d", "d2 A7", "Closed diminished second"),
	)
	return mode.MustNewRepository(&c)
}

func emptyRepository() mode.Repository {
	return mode.MustNewRepository(&template.Catalog{})
}

// identifyNumbers calls the form without a start name for start "" and the
// form with start otherwise; "unset" passes the unset name.
func identifyNumbers(r mode.Repository, kind mode.Kind, sys *system.System, adjacent []system.Step, start string) ([]template.Entry, error) {
	if start == "" {
		return r.IdentifyNumbers(kind, sys, adjacent)
	}
	return r.IdentifyNumbersFrom(kind, sys, adjacent, spelling.MustParse(start))
}

// numbersOf reads the adjacent numbers of m in the same form as the search.
func numbersOf(m mode.Mode, sys *system.System, start string) (mode.Positions, error) {
	if start == "" {
		return m.Numbers(sys)
	}
	return m.NumbersFrom(sys, spelling.MustParse(start))
}

func TestIdentifyNumbers(t *testing.T) {
	t.Parallel()
	builtIn := mode.BuiltInRepository()
	custom := customRepository()
	empty := emptyRepository()
	traditional := system.Traditional()
	positions19 := system.Positions19()
	irregular := irregularSystem()
	t1 := singleAccidentalSystem()
	for _, tt := range []struct {
		name     string
		repo     mode.Repository
		kind     mode.Kind
		sys      *system.System
		start    string
		adjacent []system.Step
		want     []string
	}{
		{"issue example", builtIn, mode.Closed, traditional, "", stepList(2, 2, 3, 2, 3), []string{"pentatonic-major"}},
		{"minor pentatonic", builtIn, mode.Closed, traditional, "", stepList(3, 2, 2, 3, 2), []string{"pentatonic-minor"}},
		{"major", builtIn, mode.Closed, traditional, "", stepList(2, 2, 1, 2, 2, 2, 1), []string{"ionian"}},
		{"harmonic minor", builtIn, mode.Closed, traditional, "", stepList(2, 1, 2, 2, 1, 3, 1), []string{"harmonic-minor"}},
		{"rotations are not searched", builtIn, mode.Closed, traditional, "", stepList(2, 3, 2, 3, 2), []string{"pentatonic-suspended"}},
		{"different templates with equal numbers", builtIn, mode.Closed, traditional, "", stepList(3, 2, 1, 1, 3, 2),
			[]string{"blues", "blues-b5"}},
		{"equal templates", builtIn, mode.Closed, traditional, "", stepList(1, 4, 1, 4, 2), []string{"hirajoshi-sachs", "iwato"}},
		{"eight degrees", builtIn, mode.Closed, traditional, "", stepList(2, 2, 1, 1, 2, 2, 1, 1),
			[]string{"messiaen-6", "seventh-flat5-diminished"}},
		{"no match", builtIn, mode.Closed, traditional, "", stepList(2, 2, 3, 2), []string{}},
		{"single number", builtIn, mode.Closed, traditional, "", stepList(12), []string{}},
		{"zero number without a match", builtIn, mode.Closed, traditional, "", stepList(0, 12), []string{}},
		{"numbers at the edge of int", builtIn, mode.Closed, traditional, "", stepList(math.MaxInt, math.MinInt), []string{}},
		{"longer than the degree limit", builtIn, mode.Closed, traditional, "", ones(mode.MaxDegrees + 1), []string{}},
		{"from C", builtIn, mode.Closed, traditional, "C", stepList(2, 2, 3, 2, 3), []string{"pentatonic-major"}},
		{"from F sharp", builtIn, mode.Closed, traditional, "F+1", stepList(2, 2, 3, 2, 3), []string{"pentatonic-major"}},
		{"open kind, four numbers", builtIn, mode.Open, traditional, "", stepList(2, 2, 3, 2), []string{}},
		{"open kind, five numbers", builtIn, mode.Open, traditional, "", stepList(2, 2, 3, 2, 3), []string{}},
		{"19-position major", builtIn, mode.Closed, positions19, "", stepList(3, 3, 2, 3, 3, 3, 2), []string{"ionian"}},
		{"19-position major pentatonic", builtIn, mode.Closed, positions19, "", stepList(3, 3, 5, 3, 5), []string{"pentatonic-major"}},
		{"19-position minor pentatonic", builtIn, mode.Closed, positions19, "", stepList(5, 3, 3, 5, 3), []string{"pentatonic-minor"}},
		{"19-position blues", builtIn, mode.Closed, positions19, "", stepList(5, 3, 1, 2, 5, 3), []string{"blues"}},
		{"19-position blues b5", builtIn, mode.Closed, positions19, "", stepList(5, 3, 2, 1, 5, 3), []string{"blues-b5"}},
		{"19-position traditional numbers", builtIn, mode.Closed, positions19, "", stepList(2, 2, 3, 2, 3), []string{}},
		{"24-position major pentatonic", builtIn, mode.Closed, system.QuarterTones24(), "", stepList(4, 4, 6, 4, 6),
			[]string{"pentatonic-major"}},
		{"24-position no match", builtIn, mode.Closed, system.QuarterTones24(), "", stepList(4, 3, 3, 4, 4, 3, 3), []string{}},
		{"zero numbers in the query", builtIn, mode.Closed, system.UpsAndDowns15(), "", stepList(3, 3, 0, 3, 3, 3, 0),
			[]string{"ionian"}},
		{"irregular from C major", builtIn, mode.Closed, irregular, "C", stepList(3, 1, 1, 2, 2, 2, 1), []string{"ionian"}},
		{"irregular from C major pentatonic", builtIn, mode.Closed, irregular, "C", stepList(3, 1, 3, 2, 3), []string{"pentatonic-major"}},
		{"irregular from C equal numbers", builtIn, mode.Closed, irregular, "C", stepList(3, 2, 2, 3, 2),
			[]string{"pentatonic-minor", "pentatonic-suspended"}},
		{"irregular from C no match", builtIn, mode.Closed, irregular, "C", stepList(2, 2, 3, 2, 3), []string{}},
		{"irregular from D major", builtIn, mode.Closed, irregular, "D", stepList(1, 2, 1, 2, 2, 2, 2), []string{"ionian"}},
		{"irregular from D sharp major", builtIn, mode.Closed, irregular, "D+1", stepList(1, 2, 1, 2, 2, 2, 2), []string{"ionian"}},
		{"irregular from D blues minor", builtIn, mode.Closed, irregular, "D", stepList(2, 2, 3, 2, 3), []string{"pentatonic-blues-minor"}},
		{"irregular from D sharp blues minor", builtIn, mode.Closed, irregular, "D+1", stepList(2, 2, 3, 2, 3),
			[]string{"pentatonic-blues-minor"}},
		{"single accidentals without a name", builtIn, mode.Closed, t1, "", stepList(2, 2, 1, 2, 2, 2, 1), []string{"ionian"}},
		{"single accidentals from C", builtIn, mode.Closed, t1, "C", stepList(2, 2, 1, 2, 2, 2, 1), []string{"ionian"}},
		{"single accidentals from C sharp", builtIn, mode.Closed, t1, "C+1", stepList(2, 2, 1, 2, 2, 2, 1), []string{"ionian"}},
		{"entry skipped for an inadmissible degree name", builtIn, mode.Closed, t1, "G+1", stepList(2, 2, 1, 2, 2, 2, 1), []string{}},
		{"custom open", custom, mode.Open, traditional, "", stepList(3, 2), []string{"a", "c"}},
		{"custom open, one number", custom, mode.Open, traditional, "", stepList(3), []string{}},
		{"custom closed", custom, mode.Closed, traditional, "", stepList(3, 2), []string{"b"}},
		{"custom closed with a zero number", custom, mode.Closed, traditional, "", stepList(0, 12), []string{"d"}},
		{"custom open, 19-position", custom, mode.Open, positions19, "", stepList(5, 3), []string{"a"}},
		{"custom open, 19-position augmented second", custom, mode.Open, positions19, "", stepList(4, 4), []string{"c"}},
		{"empty catalog", empty, mode.Closed, traditional, "", stepList(2, 2, 3, 2, 3), []string{}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := slices.Clone(tt.adjacent)
			got, err := identifyNumbers(tt.repo, tt.kind, tt.sys, input, tt.start)
			require.NoError(t, err)
			require.NotNil(t, got, "no match is a non-nil empty slice")
			assert.Equal(t, tt.want, codes(got))
			assert.Equal(t, tt.adjacent, input, "the input numbers must not change")

			// F-13: a mode created from a found entry has the requested kind and
			// numbers, and its origin is that entry.
			for _, e := range got {
				m, err := mode.FromEntry(e)
				if err != nil {
					continue
				}
				assert.Equal(t, tt.kind, m.Kind(), e.Code())
				numbers, err := numbersOf(m, tt.sys, tt.start)
				require.NoError(t, err, e.Code())
				assert.Equal(t, tt.adjacent, numbers.Adjacent, e.Code())
				origin, ok := m.Origin()
				assert.True(t, ok)
				assert.True(t, origin.Equal(e), e.Code())
			}
		})
	}

	// Issue example: the found entry creates the major pentatonic with origin.
	found := builtIn.MustIdentifyNumbers(mode.Closed, traditional, stepList(2, 2, 3, 2, 3))
	require.Len(t, found, 1)
	created := mode.MustFromEntry(found[0])
	assert.True(t, created.Equal(mode.MustFromCode("pentatonic-major")))
	assert.True(t, created.HasOrigin())
}

func TestIdentifyNumbersFailures(t *testing.T) {
	t.Parallel()
	builtIn := mode.BuiltInRepository()
	empty := emptyRepository()
	traditional := system.Traditional()
	irregular := irregularSystem()
	t1 := singleAccidentalSystem()
	pentatonic := stepList(2, 2, 3, 2, 3)
	for _, tt := range []struct {
		name     string
		repo     mode.Repository
		kind     mode.Kind
		sys      *system.System
		start    string
		adjacent []system.Step
		// category is the mode sentinel; nil means only the listed cause of
		// another package, without any mode sentinel.
		category error
		cause    error
	}{
		{"unset repository", mode.Repository{}, mode.Closed, traditional, "", pentatonic, mode.ErrUnset, nil},
		{"unset repository with a name", mode.Repository{}, mode.Closed, traditional, "C", pentatonic, mode.ErrUnset, nil},
		{"nil system", builtIn, mode.Closed, nil, "", pentatonic, mode.ErrUnset, nil},
		{"nil system with a name", builtIn, mode.Closed, nil, "C", pentatonic, mode.ErrUnset, nil},
		{"unset system", builtIn, mode.Closed, new(system.System), "", pentatonic, mode.ErrUnset, nil},
		{"unset start", builtIn, mode.Closed, traditional, "unset", pentatonic, mode.ErrUnset, nil},
		{"nil numbers", builtIn, mode.Closed, traditional, "", nil, mode.ErrEmptyInput, nil},
		{"empty numbers", builtIn, mode.Closed, traditional, "", []system.Step{}, mode.ErrEmptyInput, nil},
		{"nil numbers with a name", builtIn, mode.Closed, traditional, "C", nil, mode.ErrEmptyInput, nil},
		{"unset kind", builtIn, mode.KindUnset, traditional, "", pentatonic, mode.ErrInvalidArgument, nil},
		{"undefined kind", builtIn, mode.Kind(7), traditional, "", pentatonic, mode.ErrInvalidArgument, nil},
		{"unset kind with a name", builtIn, mode.KindUnset, traditional, "C", pentatonic, mode.ErrInvalidArgument, nil},
		{"irregular system without a name", builtIn, mode.Closed, irregular, "", stepList(2, 2, 1, 2, 2, 2, 1), nil,
			interval.ErrStartRequired},
		{"irregular system without a name, empty catalog", empty, mode.Closed, irregular, "", pentatonic, nil,
			interval.ErrStartRequired},
		{"inadmissible start", builtIn, mode.Closed, t1, "C+2", stepList(2, 2, 1, 2, 2, 2, 1), nil, system.ErrInadmissible},
		{"inadmissible start, empty catalog", empty, mode.Closed, t1, "C+2", pentatonic, nil, system.ErrInadmissible},
		{"unset repository before missing numbers", mode.Repository{}, mode.Closed, traditional, "", nil, mode.ErrUnset, nil},
		{"missing numbers before an unset kind", builtIn, mode.KindUnset, traditional, "", nil, mode.ErrEmptyInput, nil},
		{"unset kind before the start requirement", builtIn, mode.KindUnset, irregular, "", pentatonic, mode.ErrInvalidArgument, nil},
		{"unset kind before an inadmissible start", builtIn, mode.KindUnset, t1, "C+2", pentatonic, mode.ErrInvalidArgument, nil},
		{"missing numbers before the start requirement", builtIn, mode.Closed, irregular, "", nil, mode.ErrEmptyInput, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := identifyNumbers(tt.repo, tt.kind, tt.sys, tt.adjacent, tt.start)
			if tt.category == nil {
				checkForeignCause(t, err, tt.cause)
			} else {
				checkCategory(t, err, tt.category)
			}
			assert.Nil(t, got)
		})
	}
}

func TestIdentifyNumbersAgreesWithIdentifyInSystem(t *testing.T) {
	t.Parallel()
	modes := []mode.Mode{
		mode.MustFromCode("ionian"),
		mode.MustFromCode("pentatonic-major"),
		mode.MustFromCode("blues"),
		mode.MustFromCode("blues-b5"),
		mode.MustParse("m3 M2 |"),
		mode.MustParse("M2 M2 m2"),
		mode.MustParse("d2 A7"),
	}
	queries := []struct {
		sys   *system.System
		start string
	}{
		{system.Traditional(), ""},
		{system.Traditional(), "C"},
		{system.Positions19(), ""},
		{irregularSystem(), "C"},
		{irregularSystem(), "D"},
	}
	for _, repo := range []mode.Repository{mode.BuiltInRepository(), customRepository()} {
		compared := 0
		for _, m := range modes {
			for _, q := range queries {
				numbers, err := numbersOf(m, q.sys, q.start)
				if err != nil {
					continue
				}
				var want []template.Entry
				if q.start == "" {
					want = repo.MustIdentifyInSystem(m, q.sys)
				} else {
					want = repo.MustIdentifyInSystemFrom(m, q.sys, spelling.MustParse(q.start))
				}
				got, err := identifyNumbers(repo, m.Kind(), q.sys, numbers.Adjacent, q.start)
				require.NoError(t, err)
				require.Len(t, got, len(want), "%s in %s from %q", m, q.sys, q.start)
				for i := range want {
					assert.True(t, want[i].Equal(got[i]), "%s in %s from %q: %s", m, q.sys, q.start, got[i].Code())
				}
				compared++
			}
		}
		assert.Greater(t, compared, 20)
	}
}

func TestIdentifyNumbersOwnership(t *testing.T) {
	t.Parallel()
	repo := customRepository()
	before, ok := repo.Catalog()
	require.True(t, ok)
	query := stepList(3, 2)

	first := repo.MustIdentifyNumbers(mode.Open, system.Traditional(), query)
	require.Equal(t, []string{"a", "c"}, codes(first))
	first[0] = template.Entry{}
	_ = append(first[:1], template.Entry{})
	second := repo.MustIdentifyNumbersFrom(mode.Open, system.Traditional(), query, spelling.MustParse("C"))
	assert.Equal(t, []string{"a", "c"}, codes(second))
	assert.Equal(t, []string{"a", "c"}, codes(repo.MustIdentifyNumbers(mode.Open, system.Traditional(), query)))
	assert.Equal(t, stepList(3, 2), query)

	after, ok := repo.Catalog()
	require.True(t, ok)
	assert.True(t, before.Equal(after))
}
