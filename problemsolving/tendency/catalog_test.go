package tendency_test

import (
	"reflect"
	"testing"

	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/problemsolving/tendency"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuiltinCatalog(t *testing.T) {
	t.Parallel()
	want := []tendency.ID{tendency.AMSupports, tendency.AMMajorRoles, tendency.AMMinorRoles, tendency.AMNeighborTargets, tendency.AMTextTargets, tendency.SeredaTonal, tendency.SeredaModal, tendency.AMConsonance, tendency.OMTHarmonic, tendency.OMTMelodic, tendency.OMTCounterpoint, tendency.MT21Consonance, tendency.SeredaConsonance, tendency.WPConsonance, tendency.AMModalResolution, tendency.AMIntervalResolution}
	strings := []string{"am-supports", "am-major-roles", "am-minor-roles", "am-neighbor-targets", "am-text-targets", "sereda-tonal", "sereda-modal", "am-consonance", "omt-harmonic", "omt-melodic", "omt-counterpoint", "mt21-consonance", "sereda-consonance", "wp-consonance", "am-modal-resolution", "am-interval-resolution"}
	infos, err := tendency.Catalog(tendency.CatalogQuery{})
	require.NoError(t, err)
	require.Len(t, infos, 16)
	for i, id := range want {
		assert.Equal(t, strings[i], string(id))
		assert.Equal(t, id, infos[i].ID)
		s, err := tendency.Builtin(id)
		require.NoError(t, err)
		require.NotNil(t, s)
		assert.True(t, s.IsValid())
		assert.False(t, reflect.TypeOf(*s).Comparable())
		assert.Equal(t, infos[i], s.Info())
		assert.True(t, infos[i].Builtin)
		assert.NotEmpty(t, infos[i].Name)
		assert.NotEmpty(t, infos[i].Description)
		assert.NotEmpty(t, infos[i].Scope)
		assert.NotEmpty(t, infos[i].References)
		assert.NotNil(t, infos[i].RequiredContext)
		require.NotEmpty(t, infos[i].Tasks)
		for j := 1; j < len(infos[i].Tasks); j++ {
			assert.Less(t, infos[i].Tasks[j-1], infos[i].Tasks[j])
		}
		def, ok := s.Definition()
		assert.False(t, ok)
		assert.Equal(t, tendency.Definition{}, def)
		copyValue := *s
		assert.True(t, s.Equal(&copyValue))
		assert.True(t, s.Equal(built(id)))
		infos[i].Tasks[0] = tendency.TaskUnset
		infos[i].References[0] = "changed"
		assert.NotEqual(t, infos[i], s.Info())
	}
	assert.False(t, built(want[0]).Equal(built(want[1])))
}

func TestCatalogFiltersAndApplicability(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		query tendency.CatalogQuery
		ids   []tendency.ID
	}{
		{"tendencies", tendency.CatalogQuery{Task: tendency.TendencyTask}, []tendency.ID{tendency.AMNeighborTargets, tendency.AMTextTargets, tendency.SeredaTonal, tendency.SeredaModal}},
		{"pentatonic", tendency.CatalogQuery{Mode: mode.MustFromCode("pentatonic-major")}, []tendency.ID{tendency.SeredaTonal, tendency.SeredaModal}},
		{"no supports", tendency.CatalogQuery{Mode: mode.MustFromCode("pentatonic-major"), Task: tendency.SupportTask}, []tendency.ID{}},
		{"interval only", tendency.CatalogQuery{Task: tendency.IntervalResolutionTask}, []tendency.ID{tendency.AMIntervalResolution}},
		{"interval and mode", tendency.CatalogQuery{Task: tendency.IntervalResolutionTask, Mode: ionian()}, []tendency.ID{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := tendency.Catalog(tc.query)
			require.NoError(t, err)
			require.NotNil(t, got)
			ids := make([]tendency.ID, 0, len(got))
			for _, info := range got {
				ids = append(ids, info.ID)
			}
			assert.Equal(t, tc.ids, ids)
		})
	}
	for _, tc := range []struct {
		id        tendency.ID
		m         mode.Mode
		task      tendency.Task
		knowledge tendency.Knowledge
	}{
		{tendency.AMSupports, ionian(), tendency.SupportTask, tendency.Known},
		{tendency.AMSupports, mode.MustFromCode("pentatonic-major"), tendency.SupportTask, tendency.OutsideScope},
		{tendency.AMSupports, ionian(), tendency.RoleTask, tendency.UnsupportedTask},
		{tendency.AMConsonance, ionian(), tendency.ClassificationTask, tendency.UnsupportedTask},
		{tendency.AMIntervalResolution, ionian(), tendency.IntervalResolutionTask, tendency.UnsupportedTask},
	} {
		s := built(tc.id)
		got, err := s.Applicability(tc.m, tc.task)
		require.NoError(t, err)
		assertAnswer(t, s, got, tc.knowledge)
	}
}

func TestSupportsAndRoles(t *testing.T) {
	t.Parallel()
	// TH-K1/K8/K16: structure, not catalog provenance, determines applicability.
	for _, text := range []string{"P1 M2 M3 P4 P5 M6 M7", "P1 M2 M3 P4 P5 m6 M7", "P1 M2 M3 P4 P5 m6 m7", "P1 M2 m3 P4 P5 m6 m7", "P1 M2 m3 P4 P5 m6 M7", "P1 M2 m3 P4 P5 M6 M7"} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			m := closed(text, "P8")
			s := built(tendency.AMSupports)
			got := s.MustSupports(m)
			assertAnswer(t, s, got.Answer, tendency.Known)
			require.Len(t, got.Rows, 7)
			assert.Equal(t, []tendency.SupportGroup{{Kind: tendency.MainSupport, Degrees: []int{1}}, {Kind: tendency.RelativeSupport, Degrees: []int{3, 5}}, {Kind: tendency.Unstable, Degrees: []int{2, 4, 6, 7}}}, got.Groups)
			assert.Equal(t, []tendency.RankGroup{{Rank: 1, Degrees: []int{1}}, {Rank: 2, Degrees: []int{5}}, {Rank: 3, Degrees: []int{3}}}, got.Order)
			for i, row := range got.Rows {
				assert.Equal(t, i+1, row.Degree)
				assert.Equal(t, row, s.MustSupport(m, i+1))
			}
			for _, comparison := range []struct{ a, b, order int }{{1, 5, -1}, {5, 3, -1}, {3, 1, 1}, {5, 5, 0}} {
				r := s.MustCompareSupport(m, comparison.a, comparison.b)
				assert.Equal(t, tendency.Known, r.Knowledge)
				assert.Equal(t, comparison.order, r.Order)
			}
			assert.Equal(t, tendency.MissingRule, s.MustCompareSupport(m, 2, 4).Knowledge)
		})
	}
	for _, tc := range []struct {
		id     tendency.ID
		m      mode.Mode
		labels []string
	}{
		{tendency.AMMajorRoles, ionian(), []string{"T", "", "M", "S", "D", "SM", ""}},
		{tendency.AMMinorRoles, aeolian(), []string{"t", "", "m", "s", "d", "sm", ""}},
	} {
		s := built(tc.id)
		r := s.MustRoles(tc.m)
		require.Len(t, r.Rows, 7)
		for i, label := range tc.labels {
			assert.Equal(t, label, r.Rows[i].Role)
			assert.Equal(t, tendency.Known, r.Rows[i].Knowledge)
			assert.Equal(t, r.Rows[i], s.MustRole(tc.m, i+1))
		}
		assert.Equal(t, "", r.Groups[0].Role)
		assert.Equal(t, []int{2, 7}, r.Groups[0].Degrees)
	}
	assert.Equal(t, tendency.OutsideScope, built(tendency.AMMajorRoles).MustRole(aeolian(), 4).Knowledge)
	m := ionian()
	require.NoError(t, m.SetDegreeInterval(6, iv("m6")))
	require.NoError(t, m.SetDegreeInterval(6, iv("M6")))
	assert.Equal(t, tendency.Known, built(tendency.AMSupports).MustSupport(m, 1).Knowledge)
	assert.Equal(t, tendency.OutsideScope, built(tendency.AMSupports).MustSupport(open("P1 M2 M3 P4 P5 M6 M7"), 1).Knowledge)
	assert.Equal(t, tendency.OutsideScope, built(tendency.AMSupports).MustSupport(closed("P1 M2 M3 P4 P5 M6 M7", "P15"), 1).Knowledge)
	assert.Equal(t, tendency.OutsideScope, built(tendency.AMSupports).MustSupport(closed("P1 M2 A2 P4 P5 m6 m7", "P8"), 1).Knowledge)
}
