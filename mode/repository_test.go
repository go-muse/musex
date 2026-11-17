package mode_test

import (
	"sync"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type rotationMatch struct {
	code   string
	degree int
}

func rotationMatches(rs []mode.Rotation) []rotationMatch {
	matches := make([]rotationMatch, len(rs))
	for i, r := range rs {
		matches[i] = rotationMatch{r.Entry.Code(), r.Degree}
	}
	return matches
}

func TestRepositorySnapshotAndEmptyCatalog(t *testing.T) {
	t.Parallel()
	var zero mode.Repository
	c, ok := zero.Catalog()
	assert.False(t, ok)
	assert.Zero(t, c.Len())
	assert.False(t, zero.IsValid())
	invalid, err := mode.NewRepository(nil)
	checkError(t, err, mode.ErrUnset)
	assert.False(t, invalid.IsValid())

	m := mode.MustParse("m3 M2 |")
	r, err := mode.NewRepository(&c)
	require.NoError(t, err)
	assert.True(t, r.IsValid())
	snap, ok := r.Catalog()
	assert.True(t, ok)
	assert.Zero(t, snap.Len())
	result, err := r.Identify(m)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result)
	rotations, err := r.IdentifyRotations(m)
	require.NoError(t, err)
	assert.NotNil(t, rotations)
	assert.Empty(t, rotations)
	relations, err := r.Related(m, 0, mode.FromFirst, true)
	require.NoError(t, err)
	assert.NotNil(t, relations)
	assert.Empty(t, relations)

	c = template.MustNewCatalog(entry("a", "m3 M2 |", "A"), entry("b", "m3 M2", "B"), entry("c", "m3 M2 |", "C"))
	assert.Empty(t, r.MustIdentify(m))
	newer := mode.MustNewRepository(&c)
	c = template.Catalog{}
	result, err = newer.Identify(m)
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "c"}, codes(result))
	result[0] = template.Entry{}
	snap, ok = newer.Catalog()
	require.True(t, ok)
	es := snap.Entries()
	es[0] = template.Entry{}
	names := newer.MustIdentify(m)[0].Names()
	names[0] = "changed"
	assert.Equal(t, []string{"a", "c"}, codes(newer.MustIdentify(m)))
	assert.Equal(t, "A", newer.MustIdentify(m)[0].Names()[0])
	assert.Equal(t, []rotationMatch{{"a", 1}, {"c", 1}}, rotationMatches(newer.MustIdentifyRotations(m)))
	assert.Equal(t, []string{"a", "c"}, codes(newer.MustIdentifyInSystem(m, system.Traditional())))
	assert.Equal(t, []rotationMatch{{"a", 1}, {"c", 1}}, rotationMatches(newer.MustIdentifyRotationsInSystem(m, system.Traditional())))
	assert.True(t, mode.BuiltInRepository().IsValid())
}

func TestIdentifyAndRelatedCatalogModes(t *testing.T) {
	t.Parallel()
	r := mode.BuiltInRepository()
	m := mode.MustFromNotesInOctave(notes(t, "D E F G A B C", cents.Cents{}, nil))
	exact, err := r.Identify(m)
	require.NoError(t, err)
	assert.Equal(t, []string{"dorian"}, codes(exact))
	want := []rotationMatch{{"ionian", 2}, {"dorian", 1}, {"phrygian", 7}, {"lydian", 6}, {"mixolydian", 5}, {"aeolian", 4}, {"locrian", 3}}
	got, err := r.IdentifyRotations(m)
	require.NoError(t, err)
	assert.Equal(t, want, rotationMatches(got))
	related, err := r.Related(m, 0, mode.FromFirst, true)
	require.NoError(t, err)
	require.Len(t, related, len(want))
	for i, row := range related {
		assert.Equal(t, want[i], rotationMatch{row.Entry.Code(), row.Degree})
		assert.NotNil(t, row.Degrees)
		assert.Zero(t, row.Count())
	}
	related, err = r.Related(m, 1, mode.FromFirst, false)
	require.NoError(t, err)
	require.Len(t, related, 7)
	for i, tt := range []struct {
		code    string
		degrees []int
	}{
		{"dorian", []int{}}, {"mixolydian", []int{3}}, {"aeolian", []int{6}}, {"melodic-minor", []int{7}},
		{"dorian-b2", []int{2}}, {"ukrainian-dorian", []int{4}}, {"dorian-b5", []int{5}},
	} {
		assert.Equal(t, tt.code, related[i].Entry.Code())
		assert.Equal(t, 1, related[i].Degree)
		assert.Equal(t, tt.degrees, related[i].Degrees)
		assert.Equal(t, len(tt.degrees), related[i].Count())
	}
	related[2].Degrees[0] = 0
	assert.Equal(t, []int{6}, r.MustRelated(m, 1, mode.FromFirst, false)[2].Degrees)
	assert.Equal(t, codes(exact), []string{r.MustRelated(m, 0, mode.Adjacent, false)[0].Entry.Code()})
	assert.Zero(t, (mode.Relation{}).Count())
	assert.Zero(t, (mode.Mismatches{}).Count())

	custom := template.MustNewCatalog(
		entry("further", "M2 m2 m2 M2", "Further"),
		entry("periodic", "M2 m2 M2 m2", "Periodic"),
		entry("open", "M2 m2 M2 |", "Open"),
		entry("duplicate", "M2 m2 M2 m2", "Duplicate"),
	)
	periodic := mode.MustParse("m2 M2 m2 M2")
	r = mode.MustNewRepository(&custom)
	assert.Equal(t, []rotationMatch{{"periodic", 2}, {"duplicate", 2}}, rotationMatches(r.MustIdentifyRotations(periodic)))
	assert.Equal(t, []rotationMatch{{"periodic", 2}, {"duplicate", 2}}, rotationMatches(r.MustIdentifyRotationsInSystem(periodic, system.Traditional())))
	related, err = r.Related(periodic, 0, mode.Adjacent, true)
	require.NoError(t, err)
	require.Len(t, related, 2)
	assert.Equal(t, 2, related[0].Degree)
	assert.Equal(t, 2, related[1].Degree)
	assert.Equal(t, "periodic", related[0].Entry.Code())
	assert.Equal(t, "duplicate", related[1].Entry.Code())
	assert.Zero(t, related[0].Count())
	assert.Zero(t, related[1].Count())

	bad, err := r.Related(m, -1, mode.FromFirst, false)
	checkError(t, err, mode.ErrInvalidArgument, template.ErrInvalidArgument)
	assert.Nil(t, bad)
	bad, err = r.Related(m, 0, mode.MeasureUnset, false)
	checkError(t, err, mode.ErrInvalidArgument, template.ErrInvalidArgument)
	assert.Nil(t, bad)
}

func TestRepositoryNumbersRotatePositionsNotNames(t *testing.T) {
	t.Parallel()
	m := mode.MustParse("M2 M2 M2 M2 d3 M2")
	r := mode.BuiltInRepository()
	assert.Contains(t, codes(r.MustIdentifyInSystem(m, system.Traditional())), "whole-tone")
	assert.NotContains(t, codes(r.MustIdentify(m)), "whole-tone")
	assert.False(t, m.Equal(mode.MustFromCode("whole-tone")))

	irregular := irregularSystem()
	start := spelling.MustParse("C")
	major := mode.MustParse(ionian)
	query, err := mode.FromNumbersFrom(irregular, []system.Step{1, 1, 2, 2, 2, 1, 3}, start)
	require.NoError(t, err)
	assert.False(t, query.MustSameInSystemFrom(irregular, start, mode.MustFromDegree(major, 2), start))
	c := template.MustNewCatalog(
		entry("fractional", "n2 M2 M2 M2 M2 m2 m2", "Unusable"),
		entry("original", ionian, "Original"),
		entry("matched", "d2 M2 M2 M2 M2 m2 A2", "Exact numeric match"),
		entry("open", "M2 M2 m2 M2 M2 M2 |", "Open"))
	r = mode.MustNewRepository(&c)
	matches, err := r.IdentifyInSystemFrom(query, irregular, start)
	require.NoError(t, err)
	assert.Equal(t, []string{"matched"}, codes(matches))
	rotations, err := r.IdentifyRotationsInSystemFrom(query, irregular, start)
	require.NoError(t, err)
	assert.Equal(t, []rotationMatch{{"original", 2}, {"matched", 1}}, rotationMatches(rotations))

	var empty template.Catalog
	r = mode.MustNewRepository(&empty)
	matches, err = r.IdentifyInSystem(major, irregular)
	assert.ErrorIs(t, err, interval.ErrStartRequired)
	assert.Nil(t, matches)
	rotations, err = r.IdentifyRotationsInSystem(major, irregular)
	assert.ErrorIs(t, err, interval.ErrStartRequired)
	assert.Nil(t, rotations)
	matches, err = r.IdentifyInSystemFrom(major, irregular, spelling.Spelling{})
	checkError(t, err, mode.ErrUnset)
	assert.Nil(t, matches)
	rotations, err = r.IdentifyRotationsInSystemFrom(major, nil, start)
	checkError(t, err, mode.ErrUnset)
	assert.Nil(t, rotations)
	matches, err = r.IdentifyInSystemFrom(major, irregular, start)
	require.NoError(t, err)
	assert.NotNil(t, matches)
	assert.Empty(t, matches)
	rotations, err = r.IdentifyRotationsInSystemFrom(major, irregular, start)
	require.NoError(t, err)
	assert.NotNil(t, rotations)
	assert.Empty(t, rotations)
}

func TestRepositoryConcurrentReads(t *testing.T) {
	t.Parallel()
	c := template.MustNewCatalog(entry("first", dorian, "First"), entry("second", dorian, "Second"))
	r := mode.MustNewRepository(&c)
	m := mode.MustParse(dorian)
	var wg sync.WaitGroup
	start := make(chan struct{})
	results := make([][]template.Entry, 8)
	errs := make([]error, len(results))
	for i := range results {
		wg.Go(func() {
			<-start
			results[i], errs[i] = r.Identify(m)
		})
	}
	close(start)
	wg.Wait()
	for i, result := range results {
		require.NoError(t, errs[i])
		assert.Equal(t, []string{"first", "second"}, codes(result))
		result[0] = template.Entry{}
	}
	assert.Equal(t, []string{"first", "second"}, codes(r.MustIdentify(m)))
}
