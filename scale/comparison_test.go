package scale_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/frequency"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/scale"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// O-5...11/25, K-11...15/27/28/37/41.
func TestSearchComparisons(t *testing.T) {
	t.Parallel()
	g4 := at("G", 4, cents.Zero(), nil)
	g30 := at("G", 4, cents.MustFromInt(30), nil)
	g5 := at("G", 5, cents.Zero(), nil)
	s := scale.MustNew(g4, g30, g5, g4)
	for _, tt := range []struct {
		name string
		by   []scale.Comparison
		want []int
	}{
		{"default", nil, []int{0, 3}}, {"zero", []scale.Comparison{{}}, []int{0, 3}}, {"note", []scale.Comparison{scale.ByNote()}, []int{0, 3}}, {"record", []scale.Comparison{scale.ByRecord()}, []int{0, 1, 3}}, {"name", []scale.Comparison{scale.ByName()}, []int{0, 1, 2, 3}}, {"position", []scale.Comparison{scale.MustByPositionIn(system.Traditional())}, []int{0, 1, 3}}, {"sound", []scale.Comparison{scale.MustBySoundIn(tuning.Standard())}, []int{0, 3}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			indices, err := s.Find(g4, tt.by...)
			require.NoError(t, err)
			assert.Equal(t, tt.want, indices)
			contains, err := s.Contains(g4, tt.by...)
			require.NoError(t, err)
			assert.True(t, contains)
			count, err := s.Count(g4, tt.by...)
			require.NoError(t, err)
			assert.Equal(t, len(tt.want), count)
			first, ok, err := s.FindFirst(g4, tt.by...)
			require.NoError(t, err)
			assert.True(t, ok)
			assert.Equal(t, tt.want[0], first)
			last, ok, err := s.FindLast(g4, tt.by...)
			require.NoError(t, err)
			assert.True(t, ok)
			assert.Equal(t, tt.want[len(tt.want)-1], last)
		})
	}
	for _, s := range []*scale.Scale{s, scale.MustNew()} {
		absent := names("F")[0]
		got, err := s.Find(absent)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Empty(t, got)
		first, ok, err := s.FindFirst(absent)
		require.NoError(t, err)
		assert.Zero(t, first)
		assert.False(t, ok)
		last, ok, err := s.FindLast(absent)
		require.NoError(t, err)
		assert.Zero(t, last)
		assert.False(t, ok)
		contains, err := s.Contains(absent)
		require.NoError(t, err)
		assert.False(t, contains)
		count, err := s.Count(absent)
		require.NoError(t, err)
		assert.Zero(t, count)
		_, err = s.Find(note.Note{})
		assert.ErrorIs(t, err, scale.ErrUnset)
		assert.ErrorIs(t, err, note.ErrUnset)
		_, err = s.Find(g4, scale.ByName(), scale.ByRecord())
		assert.ErrorIs(t, err, scale.ErrInvalidArgument)
	}
	cs := scale.MustNew(at("C+1", 4, cents.Zero(), nil))
	db := at("D-1", 4, cents.Zero(), nil)
	matched, err := cs.Contains(db, scale.MustByPositionIn(system.Traditional()))
	require.NoError(t, err)
	assert.True(t, matched)
	matched, err = cs.Contains(db, scale.ByRecord())
	require.NoError(t, err)
	assert.False(t, matched)
}

func TestSetAlgebraAndRotationRepresentatives(t *testing.T) {
	t.Parallel()
	left, right := list("G A G B"), list("A C G")
	for _, tt := range []struct {
		name, want string
		op         func(*scale.Scale) (*scale.Scale, error)
	}{
		{"intersection", "G A", func(s *scale.Scale) (*scale.Scale, error) { return s.Intersection(right) }},
		{"difference", "B", func(s *scale.Scale) (*scale.Scale, error) { return s.Difference(right) }},
		{"union", "G A B C", func(s *scale.Scale) (*scale.Scale, error) { return s.Union(right) }},
		{"unique", "G A B", func(s *scale.Scale) (*scale.Scale, error) { return s.Unique() }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.op(left)
			require.NoError(t, err)
			checkScale(t, names(tt.want), got)
			require.NoError(t, got.Clear())
			checkScale(t, names("G A G B"), left)
		})
	}
	for _, tt := range []struct {
		a, b               string
		set, multi, subset bool
	}{{"G A G", "G A", true, false, true}, {"D E F G A B C", "C D E F G A B", true, true, true}, {"G", "G A", false, false, true}, {"", "G", false, false, true}, {"", "", true, true, true}, {"G G A", "A G G", true, true, true}} {
		a, b := list(tt.a), list(tt.b)
		same, err := a.SameSet(b)
		require.NoError(t, err)
		assert.Equal(t, tt.set, same)
		same, err = a.SameMultiset(b)
		require.NoError(t, err)
		assert.Equal(t, tt.multi, same)
		same, err = a.IsSubsetOf(b)
		require.NoError(t, err)
		assert.Equal(t, tt.subset, same)
	}
	first := at("G", 4, cents.MustFromInt(30), nil)
	second := at("G", 5, cents.Zero(), nil)
	got, err := scale.MustNew(first, second).Unique(scale.ByName())
	require.NoError(t, err)
	checkScale(t, []note.Note{first}, got)
	for _, tt := range []struct {
		a, b     string
		starts   []int
		rotation bool
	}{{"G A G A", "A G A G", []int{1, 3}, true}, {"D E F G A B C", "C D E F G A B", []int{6}, true}, {"G A", "G G", []int{}, false}, {"", "", []int{}, true}, {"G", "", []int{}, false}} {
		a, b := list(tt.a), list(tt.b)
		indices, err := a.RotationStarts(b)
		require.NoError(t, err)
		assert.Equal(t, tt.starts, indices)
		yes, err := a.IsRotationOf(b)
		require.NoError(t, err)
		assert.Equal(t, tt.rotation, yes)
	}
	assert.False(t, list("G A G").Equal(list("G A")))
	assert.False(t, list("G A").Equal(list("A G")))
	assert.True(t, list("").Equal(list("")))
	var nilScale *scale.Scale
	assert.False(t, nilScale.Equal(nilScale))
	assert.Equal(t, 0, nilScale.Compare(nilScale))
	assert.Negative(t, nilScale.Compare(list("")))
	assert.Positive(t, list("").Compare(nilScale))
	for _, tt := range []struct{ a, b string }{{"", "G"}, {"G", "G A"}} {
		assert.Negative(t, list(tt.a).Compare(list(tt.b)))
		assert.Positive(t, list(tt.b).Compare(list(tt.a)))
	}
	g, a := names("G A")[0], names("G A")[1]
	assert.Equal(t, g.Compare(a), list("G").Compare(list("A")))
}

func TestContextPreflightAndNoSourceReads(t *testing.T) {
	t.Parallel()
	src := &source{set: settings.Standard()}
	g := at("G", 4, cents.Zero(), src)
	src.forbidden = true
	s := scale.MustNew(g, g)
	for _, by := range []scale.Comparison{scale.ByNote(), scale.ByRecord(), scale.ByName(), scale.MustByPositionIn(system.Traditional()), scale.MustBySoundIn(tuning.Standard())} {
		_, err := s.Find(g, by)
		require.NoError(t, err)
		_, err = s.EqualBy(s.Clone(), by)
		require.NoError(t, err)
		_, err = s.Union(s, by)
		require.NoError(t, err)
	}
	s.MustMove(0, 1)
	s.MustTake(0, 3, scale.Forward)
	seq := s.MustCycle(0, scale.Forward)
	checkNotes(t, []note.Note{g, g, g}, collect(seq, 3))
	s.MustWithoutOctaves()
	assert.Equal(t, int64(1), src.calls.Load())
	src.forbidden = false
	src.calls.Store(0)
	_, err := s.Find(g, scale.ByPosition())
	require.NoError(t, err)
	assert.Equal(t, int64(1), src.calls.Load(), "one snapshot per distinct source")
	badSource := &source{set: settings.Standard()}
	bad := at("B", 4, cents.Zero(), badSource)
	badSource.set = nil
	invalid := scale.MustNew(g, bad)
	binary := []struct {
		name string
		run  func() error
	}{
		{"equal length mismatch", func() error { _, e := invalid.EqualBy(scale.MustNew(g), scale.ByPosition()); return e }},
		{"set early match", func() error { _, e := invalid.SameSet(scale.MustNew(g), scale.ByPosition()); return e }},
		{"multiset length", func() error { _, e := invalid.SameMultiset(scale.MustNew(g), scale.ByPosition()); return e }},
		{"subset empty other", func() error { _, e := invalid.IsSubsetOf(scale.MustNew(), scale.ByPosition()); return e }},
		{"intersection empty", func() error { _, e := scale.MustNew().Intersection(invalid, scale.ByPosition()); return e }},
		{"difference", func() error { _, e := invalid.Difference(scale.MustNew(), scale.ByPosition()); return e }},
		{"union", func() error { _, e := invalid.Union(scale.MustNew(), scale.ByPosition()); return e }},
		{"rotation lengths", func() error { _, e := invalid.RotationStarts(scale.MustNew(), scale.ByPosition()); return e }},
		{"unique", func() error { _, e := invalid.Unique(scale.ByPosition()); return e }},
		{"contains early", func() error { _, e := invalid.Contains(g, scale.ByPosition()); return e }},
		{"find first early", func() error { _, _, e := invalid.FindFirst(g, scale.ByPosition()); return e }},
		{"melody late invalid", func() error { _, e := s.ShiftMelody([]note.Note{g, bad}, 1, scale.ByPosition()); return e }},
	}
	for _, tt := range binary {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.run()
			assert.ErrorIs(t, err, note.ErrUnsetSettings)
			assert.ErrorIs(t, err, settings.ErrUnset)
		})
	}
	unbound := scale.MustNew(at("G", 4, cents.Zero(), nil))
	_, err = unbound.Find(g, scale.ByPosition())
	assert.ErrorIs(t, err, note.ErrNoSource)
	mixed := scale.MustNew(g, names("G")[0])
	_, err = mixed.Contains(g, scale.MustByPositionIn(system.Traditional()))
	assert.ErrorIs(t, err, note.ErrNoOctave)
	_, err = mixed.Contains(g, scale.MustBySoundIn(tuning.Standard()))
	assert.ErrorIs(t, err, note.ErrNoOctave)
	other := settings.Standard().MustWithSystem(system.Positions19())
	nineteen := at("G", 4, cents.Zero(), other)
	_, err = scale.MustNew(g, nineteen).Contains(g, scale.ByPosition())
	assert.ErrorIs(t, err, note.ErrDifferentSystems)
	_, err = scale.ByPositionIn(nil)
	assert.ErrorIs(t, err, scale.ErrUnset)
	assert.ErrorIs(t, err, system.ErrInvalid)
	_, err = scale.BySoundIn(nil)
	assert.ErrorIs(t, err, scale.ErrUnset)
	assert.ErrorIs(t, err, tuning.ErrUnset)
}

func TestOwnSoundSearchIsNotSetEquivalence(t *testing.T) {
	t.Parallel()
	base := tuning.Standard()
	other := base.MustWithReferenceFrequency(440)
	// Unequal reference pitches retain the same A4 frequency, forcing the pairwise frequency path.
	other = other.MustWithReferenceNote(at("C", 4, cents.Zero(), nil).MustPitch())
	own := settings.Standard().MustWithTuning(base)
	foreign := settings.Standard().MustWithTuning(other)
	a := at("A", 4, cents.Zero(), own)
	b := at("A", 4, cents.Zero(), foreign)
	c := at("A", 4, cents.MustFromFraction(1, 1_000_000_000_000_000), own)
	require.False(t, base.Equal(other))
	require.Equal(t, a.MustFrequency(), b.MustFrequency())
	require.Equal(t, b.MustFrequency(), c.MustFrequency())
	s := scale.MustNew(a, b, c)
	for _, tt := range []struct {
		name string
		n    note.Note
		want []int
	}{{"A", a, []int{0, 1}}, {"B", b, []int{0, 1, 2}}, {"C", c, []int{1, 2}}} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.FindSound(tt.n, frequency.Strict())
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
	got, err := s.FindSoundInTuning(base, a, frequency.Strict())
	require.NoError(t, err)
	assert.Equal(t, []int{0, 1}, got)
	unique, err := s.Unique(scale.MustBySoundIn(base))
	require.NoError(t, err)
	checkScale(t, []note.Note{a, c}, unique)
	got, err = s.FindSoundInTuning(base, a, frequency.MustWithinCents(cents.MustFromInt(1)))
	require.NoError(t, err)
	assert.Equal(t, []int{0, 1, 2}, got)
	got, err = s.FindSound(a, frequency.Criterion{})
	assert.Nil(t, got)
	assert.ErrorIs(t, err, scale.ErrUnset)
	got, err = scale.MustNew(a, names("G")[0]).FindSound(a, frequency.Strict())
	assert.Nil(t, got)
	assert.ErrorIs(t, err, note.ErrNoOctave)
}

func TestContextEdgesAndEmptyMelody(t *testing.T) {
	t.Parallel()
	sys := system.Traditional()
	high := at("C", math.MaxInt, cents.Zero(), nil)
	enharmonic := at("B+1", math.MaxInt-1, cents.Zero(), nil)
	got, err := scale.MustNew(high).Find(enharmonic, scale.MustByPositionIn(sys))
	require.NoError(t, err)
	assert.Equal(t, []int{0}, got)
	for _, unbound := range []note.Note{names("G")[0], high} {
		got, err := scale.MustNew().Find(unbound, scale.ByPosition())
		assert.Nil(t, got)
		assert.ErrorIs(t, err, note.ErrNoSource)
	}
	src := &source{set: settings.Standard()}
	n := at("G", 4, cents.Zero(), src)
	src.forbidden = true
	s := scale.MustNew(n)
	empty, err := s.ShiftMelody(nil, 2, scale.ByPosition())
	require.NoError(t, err)
	checkScale(t, names(""), empty)
	empty, err = s.MirrorSelected(nil, 0, scale.ByPosition())
	require.NoError(t, err)
	checkScale(t, names(""), empty)
	assert.Equal(t, int64(1), src.calls.Load())
	// No early sound match can conceal a later note without an octave.
	mixed := scale.MustNew(at("G", 4, cents.Zero(), nil), names("A")[0])
	for _, explicit := range []bool{false, true} {
		var result []int
		var err error
		if explicit {
			result, err = mixed.FindSoundInTuning(tuning.Standard(), high, frequency.Strict())
		} else {
			bound := at("G", 4, cents.Zero(), settings.Standard())
			result, err = scale.MustNew(bound, names("A")[0]).FindSound(bound, frequency.Strict())
		}
		assert.Nil(t, result)
		assert.ErrorIs(t, err, note.ErrNoOctave)
	}
}

func TestLateInadmissibleComparisonAndMelodyAtomicity(t *testing.T) {
	t.Parallel()
	good := names("G")[0]
	bad := names("C+1/2")[0]
	s := scale.MustNew(good, bad)
	by := scale.MustByPositionIn(system.Traditional())
	got, err := s.Find(good, by)
	assert.Nil(t, got)
	assert.ErrorIs(t, err, note.ErrInadmissible)
	assert.ErrorIs(t, err, system.ErrInadmissible)
	yes, err := s.SameSet(scale.MustNew(good), by)
	assert.False(t, yes)
	assert.ErrorIs(t, err, note.ErrInadmissible)
	result, err := list("G A B").ShiftSelected([]scale.Selection{{Note: good}, {Note: names("A")[0], Index: new(99)}}, 1)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, scale.ErrIndex)
	result, err = list("G A B").MirrorSelected([]scale.Selection{{Note: good}, {Note: note.Note{}}}, 0)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, scale.ErrUnset)
	checkScale(t, []note.Note{good, bad}, s)
}
