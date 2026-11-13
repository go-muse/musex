package settings_test

import (
	"testing"

	"github.com/go-muse/musex/notation"
	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/octave/number"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"
	"github.com/go-muse/musex/tuning/temperament"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The set is its own source (A 5.6).
var _ settings.Source = (*settings.Settings)(nil)

// TestStandard covers K-1 and A 5.3: the parts of the standard settings.
func TestStandard(t *testing.T) {
	t.Parallel()
	s := settings.Standard()
	require.NotNil(t, s)
	assert.True(t, s.IsValid())
	assert.True(t, system.Traditional().Equal(s.System()))
	assert.Same(t, system.Traditional(), s.System(), "a built-in system is shared, not copied")
	assert.True(t, tuning.Standard().Equal(s.Tuning()))
	assert.True(t, builtin(notation.EnglishUnicodeID).Equal(s.Notation()))
	assert.True(t, number.ASPN().Equal(s.Numbering()))
	assert.True(t, naming.English().Equal(s.Catalog()))

	other := settings.Standard()
	assert.NotSame(t, s, other, "every call returns a new pointer")
	assert.True(t, s.Equal(other))
	assert.True(t, other.Equal(s))
}

// TestStandardIndependence covers A 5.3 and A 5.10: overwriting one result of
// Standard does not change later results.
func TestStandardIndependence(t *testing.T) {
	t.Parallel()
	first := settings.Standard()
	*first = settings.Settings{}
	assert.False(t, first.IsValid())
	later := settings.Standard()
	assert.True(t, later.IsValid())
	assert.True(t, later.Equal(settings.Standard()))
	assert.True(t, tuning.Standard().Equal(later.Tuning()))
}

// TestSourceReadsWholeSet covers K-2 and A 5.6: a set as a source gives itself.
func TestSourceReadsWholeSet(t *testing.T) {
	t.Parallel()
	german := settings.Standard().MustWithNotation(builtin(notation.GermanAMID))
	var unsetNil *settings.Settings
	unsetZero := &settings.Settings{}
	for _, tc := range []struct {
		name string
		set  *settings.Settings
	}{
		{name: "standard", set: settings.Standard()},
		{name: "German AM", set: german},
		{name: "nil set", set: unsetNil},
		{name: "zero set", set: unsetZero},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var src settings.Source = tc.set
			got := src.Settings()
			assert.Same(t, tc.set, got)
			assert.True(t, tc.set.Equal(got))
			assert.Same(t, tc.set, tc.set.Settings())
		})
	}
}

// TestNew covers K-3, C-2 and A 5.3: creation from parts and reading them back.
func TestNew(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name      string
		tune      *tuning.Tuning
		notation  notation.Notation
		numbering number.Numbering
		catalog   *naming.Catalog
		wantCat   *naming.Catalog
	}{
		{
			name: "K-3 German AM, Franco-Belgian, empty catalog", tune: tuning.Standard(),
			notation: builtin(notation.GermanAMID), numbering: number.FrancoBelgian(),
			catalog: naming.MustNew(nil), wantCat: naming.MustNew(nil),
		},
		{
			name: "nil catalog is empty", tune: tuning.Standard(),
			notation: builtin(notation.EnglishUnicodeID), numbering: number.ASPN(),
			catalog: nil, wantCat: naming.MustNew(nil),
		},
		{
			name: "zero catalog is empty", tune: tuning.Standard(),
			notation: builtin(notation.EnglishUnicodeID), numbering: number.Mexican(),
			catalog: &naming.Catalog{}, wantCat: naming.MustNew(nil),
		},
		{
			name: "own numbering and Russian catalog", tune: tuning.Standard(),
			notation: builtin(notation.RussianAMID), numbering: number.MustNew(4, 10, 7),
			catalog: naming.Russian(), wantCat: naming.Russian(),
		},
		{
			name: "ups and downs with a matching system", tune: tuningFor(t, system.UpsAndDowns15()),
			notation: builtin(notation.ASCIIUpDownID, 3), numbering: number.ASPN(),
			catalog: naming.German(), wantCat: naming.German(),
		},
		{
			name: "name-only notation", tune: tuningFor(t, system.QuarterTones24()),
			notation: builtin(notation.UnicodeSteinID), numbering: number.MiddleC3(),
			catalog: naming.English(), wantCat: naming.English(),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s, err := settings.New(tc.tune, tc.notation, tc.numbering, tc.catalog)
			require.NoError(t, err)
			require.NotNil(t, s)
			assert.True(t, s.IsValid())
			assert.True(t, tc.tune.System().Equal(s.System()), "the system is the system of the tuning")
			assert.True(t, tc.tune.Equal(s.Tuning()))
			assert.True(t, tc.notation.Equal(s.Notation()))
			assert.True(t, tc.numbering.Equal(s.Numbering()))
			require.NotNil(t, s.Catalog(), "Catalog never returns nil for a set")
			assert.True(t, tc.wantCat.Equal(s.Catalog()))
			ctx := s.OctaveContext()
			assert.True(t, tc.numbering.Equal(ctx.Numbering))
			require.NotNil(t, ctx.Catalog)
			assert.True(t, tc.wantCat.Equal(ctx.Catalog))
		})
	}
}

// TestNewErrors covers K-4, K-5, E-1, E-2 and the order of A 5.3.
func TestNewErrors(t *testing.T) {
	t.Parallel()
	updown3 := builtin(notation.ASCIIUpDownID, 3)
	eu := builtin(notation.EnglishUnicodeID)
	for _, tc := range []struct {
		name      string
		tune      *tuning.Tuning
		notation  notation.Notation
		numbering number.Numbering
		want      []error
		notWant   []error
	}{
		{name: "nil tuning", tune: nil, notation: eu, numbering: number.ASPN(), want: []error{settings.ErrUnset}, notWant: []error{settings.ErrInconsistent}},
		{name: "zero tuning", tune: &tuning.Tuning{}, notation: eu, numbering: number.ASPN(), want: []error{settings.ErrUnset}},
		{name: "zero notation", tune: tuning.Standard(), notation: notation.Notation{}, numbering: number.ASPN(), want: []error{settings.ErrUnset}},
		{name: "zero numbering", tune: tuning.Standard(), notation: eu, numbering: number.Numbering{}, want: []error{settings.ErrUnset}},
		{name: "unset before inconsistent: nil tuning", tune: nil, notation: updown3, numbering: number.ASPN(), want: []error{settings.ErrUnset}, notWant: []error{settings.ErrInconsistent}},
		{name: "unset before inconsistent: zero numbering", tune: tuning.Standard(), notation: updown3, numbering: number.Numbering{}, want: []error{settings.ErrUnset}, notWant: []error{settings.ErrInconsistent}},
		{
			name: "K-5 ups and downs c = 3 in the traditional system", tune: tuning.Standard(), notation: updown3, numbering: number.ASPN(),
			want: []error{settings.ErrInconsistent, notation.ErrIncompatibleSystem}, notWant: []error{settings.ErrUnset},
		},
		{
			name: "ups and downs c = 5 in a system with c = 3", tune: tuningFor(t, system.UpsAndDowns22()), notation: builtin(notation.ASCIIUpDownID, 5), numbering: number.ASPN(),
			want: []error{settings.ErrInconsistent, notation.ErrIncompatibleSystem},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s, err := settings.New(tc.tune, tc.notation, tc.numbering, naming.English())
			assert.Nil(t, s)
			require.Error(t, err)
			for _, want := range tc.want {
				assert.ErrorIs(t, err, want)
			}
			for _, notWant := range tc.notWant {
				assert.NotErrorIs(t, err, notWant)
			}
			assertPanicsWith(t, tc.want[0], func() { settings.MustNew(tc.tune, tc.notation, tc.numbering, naming.English()) })
		})
	}

	t.Run("K-5 ups and downs c = 3 with 15-ups-and-downs", func(t *testing.T) {
		t.Parallel()
		s, err := settings.New(tuningFor(t, system.UpsAndDowns15()), updown3, number.ASPN(), nil)
		require.NoError(t, err)
		assert.True(t, s.IsValid())
		assert.True(t, settings.MustNew(tuningFor(t, system.UpsAndDowns15()), updown3, number.ASPN(), nil).Equal(s))
	})
}

// TestWithParts covers K-6, C-3 and A 5.5: replacing one part or all
// designations gives a new set and leaves the receiver unchanged.
func TestWithParts(t *testing.T) {
	t.Parallel()
	t415 := tuning.Standard().MustWithReferenceFrequency(415)
	for _, tc := range []struct {
		name  string
		apply func(s *settings.Settings) (*settings.Settings, error)
		must  func(s *settings.Settings) *settings.Settings
		tune  *tuning.Tuning
		nota  notation.Notation
		num   number.Numbering
		cat   *naming.Catalog
	}{
		{
			name:  "tuning",
			apply: func(s *settings.Settings) (*settings.Settings, error) { return s.WithTuning(t415) },
			must:  func(s *settings.Settings) *settings.Settings { return s.MustWithTuning(t415) },
			tune:  t415, nota: builtin(notation.EnglishUnicodeID), num: number.ASPN(), cat: naming.English(),
		},
		{
			name: "tuning in another system",
			apply: func(s *settings.Settings) (*settings.Settings, error) {
				return s.WithTuning(tuningFor(t, system.Positions19()))
			},
			must: func(s *settings.Settings) *settings.Settings {
				return s.MustWithTuning(tuningFor(t, system.Positions19()))
			},
			tune: tuningFor(t, system.Positions19()), nota: builtin(notation.EnglishUnicodeID), num: number.ASPN(), cat: naming.English(),
		},
		{
			name:  "notation",
			apply: func(s *settings.Settings) (*settings.Settings, error) { return s.WithNotation(builtin(notation.ABCID)) },
			must:  func(s *settings.Settings) *settings.Settings { return s.MustWithNotation(builtin(notation.ABCID)) },
			tune:  tuning.Standard(), nota: builtin(notation.ABCID), num: number.ASPN(), cat: naming.English(),
		},
		{
			name:  "numbering",
			apply: func(s *settings.Settings) (*settings.Settings, error) { return s.WithNumbering(number.Mexican()) },
			must:  func(s *settings.Settings) *settings.Settings { return s.MustWithNumbering(number.Mexican()) },
			tune:  tuning.Standard(), nota: builtin(notation.EnglishUnicodeID), num: number.Mexican(), cat: naming.English(),
		},
		{
			name:  "catalog",
			apply: func(s *settings.Settings) (*settings.Settings, error) { return s.WithCatalog(naming.Russian()) },
			must:  func(s *settings.Settings) *settings.Settings { return s.MustWithCatalog(naming.Russian()) },
			tune:  tuning.Standard(), nota: builtin(notation.EnglishUnicodeID), num: number.ASPN(), cat: naming.Russian(),
		},
		{
			name:  "nil catalog",
			apply: func(s *settings.Settings) (*settings.Settings, error) { return s.WithCatalog(nil) },
			must:  func(s *settings.Settings) *settings.Settings { return s.MustWithCatalog(nil) },
			tune:  tuning.Standard(), nota: builtin(notation.EnglishUnicodeID), num: number.ASPN(), cat: naming.MustNew(nil),
		},
		{
			name: "designations",
			apply: func(s *settings.Settings) (*settings.Settings, error) {
				return s.WithDesignations(builtin(notation.GermanAMID), number.FrancoBelgian(), naming.Russian())
			},
			must: func(s *settings.Settings) *settings.Settings {
				return s.MustWithDesignations(builtin(notation.GermanAMID), number.FrancoBelgian(), naming.Russian())
			},
			tune: tuning.Standard(), nota: builtin(notation.GermanAMID), num: number.FrancoBelgian(), cat: naming.Russian(),
		},
		{
			name: "designations with a nil catalog",
			apply: func(s *settings.Settings) (*settings.Settings, error) {
				return s.WithDesignations(builtin(notation.RomanceItalianID), number.MiddleC5(), nil)
			},
			must: func(s *settings.Settings) *settings.Settings {
				return s.MustWithDesignations(builtin(notation.RomanceItalianID), number.MiddleC5(), nil)
			},
			tune: tuning.Standard(), nota: builtin(notation.RomanceItalianID), num: number.MiddleC5(), cat: naming.MustNew(nil),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s := settings.Standard()
			got, err := tc.apply(s)
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.NotSame(t, s, got, "a replacement returns a new set")
			assert.True(t, got.IsValid())
			assert.True(t, tc.tune.Equal(got.Tuning()))
			assert.True(t, tc.tune.System().Equal(got.System()))
			assert.True(t, tc.nota.Equal(got.Notation()))
			assert.True(t, tc.num.Equal(got.Numbering()))
			assert.True(t, tc.cat.Equal(got.Catalog()))
			assert.True(t, settings.Standard().Equal(s), "the receiver does not change")
			assert.True(t, got.Equal(tc.must(settings.Standard())), "the Must variant gives an equal set")
		})
	}
}

// TestWithErrors covers K-5, E-1, E-2 and the orders of A 5.5.
func TestWithErrors(t *testing.T) {
	t.Parallel()
	updown3 := builtin(notation.ASCIIUpDownID, 3)
	s15 := newSet(t, tuningFor(t, system.UpsAndDowns15()), updown3, number.ASPN(), nil)
	var unsetNil *settings.Settings
	unsetZero := &settings.Settings{}
	eu := builtin(notation.EnglishUnicodeID)
	for _, tc := range []struct {
		name    string
		recv    *settings.Settings
		apply   func(s *settings.Settings) (*settings.Settings, error)
		must    func(s *settings.Settings)
		want    []error
		notWant []error
	}{
		{name: "WithTuning on nil set", recv: unsetNil, apply: func(s *settings.Settings) (*settings.Settings, error) { return s.WithTuning(tuning.Standard()) }, must: func(s *settings.Settings) { s.MustWithTuning(tuning.Standard()) }, want: []error{settings.ErrUnset}},
		{name: "WithTuning on zero set", recv: unsetZero, apply: func(s *settings.Settings) (*settings.Settings, error) { return s.WithTuning(tuning.Standard()) }, must: func(s *settings.Settings) { s.MustWithTuning(tuning.Standard()) }, want: []error{settings.ErrUnset}},
		{name: "WithTuning nil", recv: settings.Standard(), apply: func(s *settings.Settings) (*settings.Settings, error) { return s.WithTuning(nil) }, must: func(s *settings.Settings) { s.MustWithTuning(nil) }, want: []error{settings.ErrUnset}},
		{name: "WithTuning zero", recv: settings.Standard(), apply: func(s *settings.Settings) (*settings.Settings, error) { return s.WithTuning(&tuning.Tuning{}) }, must: func(s *settings.Settings) { s.MustWithTuning(&tuning.Tuning{}) }, want: []error{settings.ErrUnset}},
		{
			name: "WithTuning inconsistent with ups and downs", recv: s15,
			apply: func(s *settings.Settings) (*settings.Settings, error) { return s.WithTuning(tuning.Standard()) },
			must:  func(s *settings.Settings) { s.MustWithTuning(tuning.Standard()) },
			want:  []error{settings.ErrInconsistent, notation.ErrIncompatibleSystem}, notWant: []error{settings.ErrUnset},
		},
		{name: "WithNotation on nil set", recv: unsetNil, apply: func(s *settings.Settings) (*settings.Settings, error) { return s.WithNotation(eu) }, must: func(s *settings.Settings) { s.MustWithNotation(eu) }, want: []error{settings.ErrUnset}},
		{name: "WithNotation zero", recv: settings.Standard(), apply: func(s *settings.Settings) (*settings.Settings, error) { return s.WithNotation(notation.Notation{}) }, must: func(s *settings.Settings) { s.MustWithNotation(notation.Notation{}) }, want: []error{settings.ErrUnset}},
		{
			name: "K-5 WithNotation inconsistent", recv: settings.Standard(),
			apply: func(s *settings.Settings) (*settings.Settings, error) { return s.WithNotation(updown3) },
			must:  func(s *settings.Settings) { s.MustWithNotation(updown3) },
			want:  []error{settings.ErrInconsistent, notation.ErrIncompatibleSystem}, notWant: []error{settings.ErrUnset},
		},
		{name: "unset set before inconsistent notation", recv: unsetZero, apply: func(s *settings.Settings) (*settings.Settings, error) { return s.WithNotation(updown3) }, must: func(s *settings.Settings) { s.MustWithNotation(updown3) }, want: []error{settings.ErrUnset}, notWant: []error{settings.ErrInconsistent}},
		{name: "WithNumbering on nil set", recv: unsetNil, apply: func(s *settings.Settings) (*settings.Settings, error) { return s.WithNumbering(number.ASPN()) }, must: func(s *settings.Settings) { s.MustWithNumbering(number.ASPN()) }, want: []error{settings.ErrUnset}},
		{name: "WithNumbering zero", recv: settings.Standard(), apply: func(s *settings.Settings) (*settings.Settings, error) { return s.WithNumbering(number.Numbering{}) }, must: func(s *settings.Settings) { s.MustWithNumbering(number.Numbering{}) }, want: []error{settings.ErrUnset}},
		{name: "WithCatalog on nil set", recv: unsetNil, apply: func(s *settings.Settings) (*settings.Settings, error) { return s.WithCatalog(naming.English()) }, must: func(s *settings.Settings) { s.MustWithCatalog(naming.English()) }, want: []error{settings.ErrUnset}},
		{name: "WithCatalog on zero set", recv: unsetZero, apply: func(s *settings.Settings) (*settings.Settings, error) { return s.WithCatalog(nil) }, must: func(s *settings.Settings) { s.MustWithCatalog(nil) }, want: []error{settings.ErrUnset}},
		{name: "WithDesignations on nil set", recv: unsetNil, apply: func(s *settings.Settings) (*settings.Settings, error) {
			return s.WithDesignations(eu, number.ASPN(), nil)
		}, must: func(s *settings.Settings) { s.MustWithDesignations(eu, number.ASPN(), nil) }, want: []error{settings.ErrUnset}},
		{name: "WithDesignations zero notation", recv: settings.Standard(), apply: func(s *settings.Settings) (*settings.Settings, error) {
			return s.WithDesignations(notation.Notation{}, number.ASPN(), nil)
		}, must: func(s *settings.Settings) { s.MustWithDesignations(notation.Notation{}, number.ASPN(), nil) }, want: []error{settings.ErrUnset}},
		{name: "WithDesignations zero numbering", recv: settings.Standard(), apply: func(s *settings.Settings) (*settings.Settings, error) {
			return s.WithDesignations(eu, number.Numbering{}, nil)
		}, must: func(s *settings.Settings) { s.MustWithDesignations(eu, number.Numbering{}, nil) }, want: []error{settings.ErrUnset}},
		{
			name: "WithDesignations unset numbering before inconsistent notation", recv: settings.Standard(),
			apply: func(s *settings.Settings) (*settings.Settings, error) {
				return s.WithDesignations(updown3, number.Numbering{}, nil)
			},
			must: func(s *settings.Settings) { s.MustWithDesignations(updown3, number.Numbering{}, nil) },
			want: []error{settings.ErrUnset}, notWant: []error{settings.ErrInconsistent},
		},
		{
			name: "WithDesignations inconsistent", recv: settings.Standard(),
			apply: func(s *settings.Settings) (*settings.Settings, error) {
				return s.WithDesignations(updown3, number.ASPN(), nil)
			},
			must: func(s *settings.Settings) { s.MustWithDesignations(updown3, number.ASPN(), nil) },
			want: []error{settings.ErrInconsistent, notation.ErrIncompatibleSystem},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var before *settings.Settings
			if tc.recv.IsValid() {
				before = settings.MustNew(tc.recv.Tuning(), tc.recv.Notation(), tc.recv.Numbering(), tc.recv.Catalog())
			}
			got, err := tc.apply(tc.recv)
			assert.Nil(t, got)
			require.Error(t, err)
			for _, want := range tc.want {
				assert.ErrorIs(t, err, want)
			}
			for _, notWant := range tc.notWant {
				assert.NotErrorIs(t, err, notWant)
			}
			assertPanicsWith(t, tc.want[0], func() { tc.must(tc.recv) })
			if before != nil {
				assert.True(t, before.Equal(tc.recv), "a failed replacement leaves the receiver unchanged")
			}
		})
	}
}

// TestWithSystem covers K-7, E-5 and A 5.5: changing the system keeps the
// temperament, the root number, the reference and the designations.
func TestWithSystem(t *testing.T) {
	t.Parallel()

	t.Run("K-7 user system with N = 12", func(t *testing.T) {
		t.Parallel()
		s := settings.Standard().MustWithCatalog(naming.Russian()).MustWithNumbering(number.FrancoBelgian())
		k7 := k7System(t)
		got, err := s.WithSystem(k7)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.True(t, k7.Equal(got.System()))
		assert.True(t, k7.Equal(got.Tuning().System()))
		assert.True(t, s.Tuning().Temperament().Equal(got.Tuning().Temperament()))
		assert.Equal(t, s.Tuning().RootClass().Number(), got.Tuning().RootClass().Number())
		assert.True(t, s.Tuning().Reference().Equal(got.Tuning().Reference()))
		assert.True(t, s.Notation().Equal(got.Notation()))
		assert.True(t, s.Numbering().Equal(got.Numbering()))
		assert.True(t, s.Catalog().Equal(got.Catalog()))
		assert.True(t, got.Equal(s.MustWithSystem(k7)))
		assert.True(t, system.Traditional().Equal(s.System()), "the receiver does not change")
	})

	t.Run("the same N with ups and downs", func(t *testing.T) {
		t.Parallel()
		s := newSet(t, tuningFor(t, system.UpsAndDowns22()), builtin(notation.ASCIIUpDownID, 3), number.ASPN(), nil)
		other, err := system.New(system.Definition{
			Name: "22 with c = 3", Classes: 22, Boundary: system.UpsAndDowns22().Boundary(),
			Offsets: system.UpsAndDowns22().Definition().Offsets, Chromatic: 3,
			Alterations: system.Integers(),
		})
		require.NoError(t, err)
		got, err := s.WithSystem(other)
		require.NoError(t, err)
		assert.True(t, other.Equal(got.System()))
	})

	reference := pitch.MustFromASPN(altered(spelling.C(), 1, 2), 4)
	halves := customSystem(t, 2, system.MustMultiples(2))
	halvesSet := newSet(t, tuning.MustNew(halves, temperament.Equal12(), tuning.RootFromName(spelling.C()), tuning.MustNewReference(reference, 450)),
		builtin(notation.EnglishUnicodeID), number.ASPN(), nil)
	s15 := newSet(t, tuningFor(t, system.UpsAndDowns15()), builtin(notation.ASCIIUpDownID, 3), number.ASPN(), nil)
	var unsetNil *settings.Settings
	for _, tc := range []struct {
		name    string
		recv    *settings.Settings
		sys     *system.System
		want    []error
		notWant []error
	}{
		{name: "nil set", recv: unsetNil, sys: system.Traditional(), want: []error{settings.ErrUnset}},
		{name: "zero set", recv: &settings.Settings{}, sys: system.Traditional(), want: []error{settings.ErrUnset}},
		{name: "nil system", recv: settings.Standard(), sys: nil, want: []error{settings.ErrUnset}},
		{name: "zero system", recv: settings.Standard(), sys: &system.System{}, want: []error{settings.ErrUnset}},
		{
			name: "K-7 another N", recv: settings.Standard(), sys: system.QuarterTones24(),
			want: []error{tuning.ErrClassCount}, notWant: []error{settings.ErrUnset, settings.ErrInconsistent},
		},
		{
			name: "reference note not admissible", recv: halvesSet, sys: system.Traditional(),
			want: []error{tuning.ErrInadmissible}, notWant: []error{settings.ErrUnset, settings.ErrInconsistent},
		},
		{
			name: "chromatic unit of ups and downs", recv: s15, sys: system.Blackwood15(),
			want: []error{settings.ErrInconsistent, notation.ErrIncompatibleSystem}, notWant: []error{settings.ErrUnset},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := tc.recv.WithSystem(tc.sys)
			assert.Nil(t, got)
			require.Error(t, err)
			for _, want := range tc.want {
				assert.ErrorIs(t, err, want)
			}
			for _, notWant := range tc.notWant {
				assert.NotErrorIs(t, err, notWant)
			}
			assertPanicsWith(t, tc.want[0], func() { tc.recv.MustWithSystem(tc.sys) })
		})
	}
}

// TestEqual covers K-8, P-16 and A 5.2.
func TestEqual(t *testing.T) {
	t.Parallel()
	std := settings.Standard()
	var unsetNil *settings.Settings
	unsetZero := &settings.Settings{}
	same := newSet(t, tuning.Standard(), builtin(notation.EnglishUnicodeID), number.ASPN(), naming.English())
	renamed := newSet(t, tuning.Standard(),
		styled(t, builtin(notation.EnglishUnicodeID).NameRules(), notation.OctaveStyleNumber),
		number.MustNew(4, 4), naming.MustNew(map[int]string{
			0: "subcontra", 1: "contra", 2: "great", 3: "small",
			4: "one-line", 5: "two-line", 6: "three-line", 7: "four-line",
		}))
	for _, tc := range []struct {
		name string
		a, b *settings.Settings
		want bool
	}{
		{name: "standard and standard", a: std, b: settings.Standard(), want: true},
		{name: "same parts", a: std, b: same, want: true},
		{name: "equal parts made differently", a: std, b: renamed, want: true},
		{name: "same pointer", a: std, b: std, want: true},
		{name: "other system", a: std, b: std.MustWithSystem(k7System(t)), want: false},
		{name: "other tuning", a: std, b: std.MustWithTuning(tuning.Standard().MustWithReferenceFrequency(415)), want: false},
		{name: "other notation", a: std, b: std.MustWithNotation(builtin(notation.ABCID)), want: false},
		{name: "other numbering", a: std, b: std.MustWithNumbering(number.Mexican()), want: false},
		{name: "equal numberings", a: std.MustWithNumbering(number.Mexican()), b: std.MustWithNumbering(number.MiddleC5()), want: true},
		{name: "other catalog", a: std, b: std.MustWithCatalog(naming.Russian()), want: false},
		{name: "empty catalog", a: std, b: std.MustWithCatalog(nil), want: false},
		{name: "empty catalogs", a: std.MustWithCatalog(nil), b: std.MustWithCatalog(&naming.Catalog{}), want: true},
		{name: "two nil sets", a: unsetNil, b: unsetNil, want: true},
		{name: "nil and zero sets", a: unsetNil, b: unsetZero, want: true},
		{name: "two zero sets", a: unsetZero, b: &settings.Settings{}, want: true},
		{name: "nil and set", a: unsetNil, b: std, want: false},
		{name: "zero and set", a: unsetZero, b: std, want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, tc.a.Equal(tc.b))
			assert.Equal(t, tc.want, tc.b.Equal(tc.a), "Equal is symmetric")
		})
	}
}

// TestString covers K-26: the debugging string names all five parts and
// does not take part in equality. Its text is not part of the contract.
func TestString(t *testing.T) {
	t.Parallel()
	std := settings.Standard()
	var unsetNil *settings.Settings
	unset := unsetNil.String()
	assert.NotEmpty(t, unset)
	assert.Equal(t, unset, (&settings.Settings{}).String(), "nil and zero sets are both unset")
	assert.NotEqual(t, unset, std.String())
	assert.Equal(t, std.String(), settings.Standard().String(), "equal sets made the same way have the same string")

	for _, tc := range []struct {
		name  string
		other *settings.Settings
	}{
		{name: "system", other: std.MustWithSystem(k7System(t))},
		{name: "tuning", other: std.MustWithTuning(tuning.Standard().MustWithReferenceFrequency(415))},
		{name: "notation", other: std.MustWithNotation(builtin(notation.ABCID))},
		{name: "numbering", other: std.MustWithNumbering(number.Mexican())},
		{name: "catalog", other: std.MustWithCatalog(nil)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.NotEqual(t, std.String(), tc.other.String(), "the string names the %s", tc.name)
			assert.NotEqual(t, unset, tc.other.String())
		})
	}
}

// TestUnsetSet covers A 5.2: the nil set and the pointer to the zero set are
// unset; readers give unset parts, other methods fail with ErrUnset.
func TestUnsetSet(t *testing.T) {
	t.Parallel()
	var unsetNil *settings.Settings
	for _, tc := range []struct {
		name string
		set  *settings.Settings
	}{
		{name: "nil", set: unsetNil},
		{name: "zero", set: &settings.Settings{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s := tc.set
			assert.False(t, s.IsValid())
			assert.Nil(t, s.System())
			assert.Nil(t, s.Tuning())
			assert.False(t, s.Notation().IsDefined())
			assert.False(t, s.Numbering().IsDefined())
			assert.Nil(t, s.Catalog())
			ctx := s.OctaveContext()
			assert.False(t, ctx.Numbering.IsDefined())
			assert.Nil(t, ctx.Catalog)
			assert.Same(t, s, s.Settings())
			assert.NotEmpty(t, s.String())

			std := settings.Standard()
			name := spelling.CSharp()
			p := at(name, 4)
			for _, op := range []struct {
				name string
				call func() error
			}{
				{name: "WithTuning", call: func() error { r, err := s.WithTuning(tuning.Standard()); assert.Nil(t, r); return err }},
				{name: "WithSystem", call: func() error { r, err := s.WithSystem(system.Traditional()); assert.Nil(t, r); return err }},
				{name: "WithNotation", call: func() error { r, err := s.WithNotation(builtin(notation.ABCID)); assert.Nil(t, r); return err }},
				{name: "WithNumbering", call: func() error { r, err := s.WithNumbering(number.ASPN()); assert.Nil(t, r); return err }},
				{name: "WithCatalog", call: func() error { r, err := s.WithCatalog(naming.English()); assert.Nil(t, r); return err }},
				{name: "WithDesignations", call: func() error {
					r, err := s.WithDesignations(builtin(notation.ABCID), number.ASPN(), nil)
					assert.Nil(t, r)
					return err
				}},
				{name: "Expresses", call: func() error { r, err := s.Expresses(name); assert.False(t, r); return err }},
				{name: "ExpressesPitch", call: func() error { r, err := s.ExpressesPitch(p); assert.False(t, r); return err }},
				{name: "Format", call: func() error { r, err := s.Format(name); assert.Empty(t, r); return err }},
				{name: "Parse", call: func() error { r, err := s.Parse("C♯"); assert.Equal(t, spelling.Spelling{}, r); return err }},
				{name: "FormatPitch", call: func() error { r, err := s.FormatPitch(p); assert.Empty(t, r); return err }},
				{name: "ParsePitch", call: func() error { r, err := s.ParsePitch("C♯4"); assert.Equal(t, pitch.Pitch{}, r); return err }},
				{name: "CheckReplacement receiver", call: func() error { r, err := s.CheckReplacement(std); assert.Equal(t, settings.Report{}, r); return err }},
				{name: "CheckReplacement next", call: func() error { r, err := std.CheckReplacement(s); assert.Equal(t, settings.Report{}, r); return err }},
				{name: "CheckStructure receiver", call: func() error { r, err := s.CheckStructure(std); assert.Equal(t, settings.Report{}, r); return err }},
				{name: "CheckStructure next", call: func() error { r, err := std.CheckStructure(s); assert.Equal(t, settings.Report{}, r); return err }},
				{name: "CheckDesignations receiver", call: func() error { r, err := s.CheckDesignations(std); assert.Equal(t, settings.Report{}, r); return err }},
				{name: "CheckDesignations next", call: func() error { r, err := std.CheckDesignations(s); assert.Equal(t, settings.Report{}, r); return err }},
				{name: "CheckReplacement both", call: func() error { r, err := s.CheckReplacement(s); assert.Equal(t, settings.Report{}, r); return err }},
				{name: "Changes receiver", call: func() error { r, err := s.Changes(std); assert.Equal(t, settings.Changes{}, r); return err }},
				{name: "Changes next", call: func() error { r, err := std.Changes(s); assert.Equal(t, settings.Changes{}, r); return err }},
				{name: "Changes both", call: func() error { r, err := s.Changes(s); assert.Equal(t, settings.Changes{}, r); return err }},
			} {
				assert.ErrorIs(t, op.call(), settings.ErrUnset, op.name)
			}
		})
	}
}

// TestMustVariants covers K-20 and A 5.12: every Must variant returns the
// same result on success and panics with the same error on failure.
func TestMustVariants(t *testing.T) {
	t.Parallel()
	std := settings.Standard()
	quarter := altered(spelling.C(), 1, 2)
	var unsetNil *settings.Settings

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		assert.True(t, std.MustExpresses(spelling.CSharp()))
		assert.True(t, std.MustExpressesPitch(at(spelling.CSharp(), 4)))
		assert.Equal(t, "C♯", std.MustFormat(spelling.CSharp()))
		assert.Equal(t, spelling.DFlat(), std.MustParse("D♭"))
		assert.Equal(t, "C♯4", std.MustFormatPitch(at(spelling.CSharp(), 4)))
		assert.True(t, at(spelling.BSharp(), 4).Equal(std.MustParsePitch("B♯3")))
		assert.Equal(t, std.MustCheckReplacement(std), std.MustCheckReplacement(std))
		r, err := std.CheckReplacement(std)
		require.NoError(t, err)
		assert.Equal(t, r, std.MustCheckReplacement(std))
		r, err = std.CheckStructure(std)
		require.NoError(t, err)
		assert.Equal(t, r, std.MustCheckStructure(std))
		r, err = std.CheckDesignations(std)
		require.NoError(t, err)
		assert.Equal(t, r, std.MustCheckDesignations(std))
		r, err = settings.CheckSystems(system.Traditional(), system.QuarterTones24())
		require.NoError(t, err)
		assert.Equal(t, r, settings.MustCheckSystems(system.Traditional(), system.QuarterTones24()))
		c, err := std.Changes(std)
		require.NoError(t, err)
		assert.Equal(t, c, std.MustChanges(std))
		assert.True(t, std.Equal(settings.MustNew(tuning.Standard(), builtin(notation.EnglishUnicodeID), number.ASPN(), naming.English())))
	})

	for _, tc := range []struct {
		name string
		call func() error
		must func()
		want error
	}{
		{name: "Expresses", call: func() error { _, err := std.Expresses(spelling.Spelling{}); return err }, must: func() { std.MustExpresses(spelling.Spelling{}) }, want: settings.ErrUnsetName},
		{name: "ExpressesPitch", call: func() error { _, err := std.ExpressesPitch(pitch.Pitch{}); return err }, must: func() { std.MustExpressesPitch(pitch.Pitch{}) }, want: settings.ErrUnsetName},
		{name: "Format", call: func() error { _, err := std.Format(quarter); return err }, must: func() { std.MustFormat(quarter) }, want: settings.ErrInadmissible},
		{name: "Parse", call: func() error { _, err := std.Parse("H"); return err }, must: func() { std.MustParse("H") }, want: notation.ErrParse},
		{name: "FormatPitch", call: func() error { _, err := std.FormatPitch(at(quarter, 4)); return err }, must: func() { std.MustFormatPitch(at(quarter, 4)) }, want: settings.ErrInadmissible},
		{name: "ParsePitch", call: func() error { _, err := std.ParsePitch("C"); return err }, must: func() { std.MustParsePitch("C") }, want: notation.ErrParse},
		{name: "CheckSystems", call: func() error { _, err := settings.CheckSystems(nil, system.Traditional()); return err }, must: func() { settings.MustCheckSystems(nil, system.Traditional()) }, want: settings.ErrUnset},
		{name: "CheckReplacement", call: func() error { _, err := std.CheckReplacement(nil); return err }, must: func() { std.MustCheckReplacement(nil) }, want: settings.ErrUnset},
		{name: "CheckStructure", call: func() error { _, err := unsetNil.CheckStructure(std); return err }, must: func() { unsetNil.MustCheckStructure(std) }, want: settings.ErrUnset},
		{name: "CheckDesignations", call: func() error { _, err := std.CheckDesignations(&settings.Settings{}); return err }, must: func() { std.MustCheckDesignations(&settings.Settings{}) }, want: settings.ErrUnset},
		{name: "Changes", call: func() error { _, err := std.Changes(nil); return err }, must: func() { std.MustChanges(nil) }, want: settings.ErrUnset},
		{name: "New", call: func() error {
			_, err := settings.New(nil, builtin(notation.EnglishUnicodeID), number.ASPN(), nil)
			return err
		}, must: func() { settings.MustNew(nil, builtin(notation.EnglishUnicodeID), number.ASPN(), nil) }, want: settings.ErrUnset},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.ErrorIs(t, tc.call(), tc.want)
			assertPanicsWith(t, tc.want, tc.must)
		})
	}
}

// TestCopies covers K-27, P-5 and A 5.10: the set keeps its own copies of the
// tuning and the catalog, and gives new copies when read.
func TestCopies(t *testing.T) {
	t.Parallel()

	t.Run("inputs of New", func(t *testing.T) {
		t.Parallel()
		tu := tuning.Standard()
		cat := naming.Russian()
		s := newSet(t, tu, builtin(notation.EnglishUnicodeID), number.ASPN(), cat)
		*tu = *tuningFor(t, system.Positions19())
		*cat = naming.Catalog{}
		assert.True(t, tuning.Standard().Equal(s.Tuning()))
		assert.True(t, system.Traditional().Equal(s.System()))
		assert.True(t, naming.Russian().Equal(s.Catalog()))
	})

	t.Run("inputs of With methods", func(t *testing.T) {
		t.Parallel()
		tu := tuning.Standard().MustWithReferenceFrequency(415)
		cat := naming.Russian()
		cat2 := naming.German()
		s := settings.Standard().MustWithTuning(tu).MustWithCatalog(cat)
		d := settings.Standard().MustWithDesignations(builtin(notation.GermanAMID), number.ASPN(), cat2)
		*tu = tuning.Tuning{}
		*cat = naming.Catalog{}
		*cat2 = naming.Catalog{}
		assert.True(t, tuning.Standard().MustWithReferenceFrequency(415).Equal(s.Tuning()))
		assert.True(t, naming.Russian().Equal(s.Catalog()))
		assert.True(t, naming.German().Equal(d.Catalog()))
	})

	t.Run("results of readers", func(t *testing.T) {
		t.Parallel()
		s := settings.Standard()
		tu := s.Tuning()
		assert.NotSame(t, tu, s.Tuning(), "Tuning returns a new tuning")
		*tu = tuning.Tuning{}
		cat := s.Catalog()
		assert.NotSame(t, cat, s.Catalog(), "Catalog returns a new catalog")
		*cat = naming.Catalog{}
		ctx := s.OctaveContext()
		require.NotNil(t, ctx.Catalog)
		assert.NotSame(t, ctx.Catalog, s.OctaveContext().Catalog, "OctaveContext returns a new catalog")
		*ctx.Catalog = naming.Catalog{}
		assert.True(t, s.Equal(settings.Standard()))
		assert.True(t, tuning.Standard().Equal(s.Tuning()))
		assert.True(t, naming.English().Equal(s.Catalog()))
		assert.True(t, naming.English().Equal(s.OctaveContext().Catalog))
		assert.True(t, settings.Standard().Equal(s))
	})

	t.Run("empty catalog reads as a new catalog", func(t *testing.T) {
		t.Parallel()
		s := settings.Standard().MustWithCatalog(nil)
		cat := s.Catalog()
		require.NotNil(t, cat)
		*cat = *naming.Russian()
		assert.True(t, naming.MustNew(nil).Equal(s.Catalog()))
	})
}

// TestSameSource covers K-28, P-9a and the table of A 5.6.
func TestSameSource(t *testing.T) {
	t.Parallel()
	std := settings.Standard()
	std2 := settings.Standard()
	german := std.MustWithNotation(builtin(notation.GermanAMID))
	var unsetNil *settings.Settings
	unsetZero := &settings.Settings{}
	ptr := &pointerSource{t: t, s: std}
	ptr2 := &pointerSource{t: t, s: std}
	var nilPtr *pointerSource
	slice := sliceSource{sets: []*settings.Settings{std}}
	slice2 := sliceSource{sets: []*settings.Settings{std}}
	withSlice := anySource{held: []int{1}}
	withSlice2 := anySource{held: []int{1}}
	withInt := anySource{held: 1}

	for _, tc := range []struct {
		name string
		a, b settings.Source
		want bool
	}{
		{name: "two separately obtained standard sets", a: std, b: std2, want: true},
		{name: "the same set", a: std, b: std, want: true},
		{name: "standard and an unequal set", a: std, b: german, want: false},
		{name: "equal sets made differently", a: german, b: settings.Standard().MustWithNotation(builtin(notation.GermanAMID)), want: true},
		{name: "set and absent source", a: std, b: nil, want: false},
		{name: "two absent sources", a: nil, b: nil, want: false},
		{name: "source of another kind and the set it gives", a: ptr, b: std, want: false},
		{name: "same instance of a pointer source", a: ptr, b: ptr, want: true},
		{name: "two instances of a pointer source", a: ptr, b: ptr2, want: false},
		{name: "nil pointer source with itself", a: nilPtr, b: nilPtr, want: true},
		{name: "pointer source and absent source", a: ptr, b: nil, want: false},
		{name: "two nil sets", a: unsetNil, b: unsetNil, want: true},
		{name: "nil and zero sets", a: unsetNil, b: unsetZero, want: true},
		{name: "unset and standard sets", a: unsetNil, b: std, want: false},
		{name: "zero and standard sets", a: unsetZero, b: std, want: false},
		{name: "unset set and absent source", a: unsetNil, b: nil, want: false},
		{name: "zero set and absent source", a: unsetZero, b: nil, want: false},
		{name: "equal comparable values", a: valueSource{id: 1, s: std}, b: valueSource{id: 1, s: std}, want: true},
		{name: "unequal comparable values", a: valueSource{id: 1, s: std}, b: valueSource{id: 2, s: std}, want: false},
		{name: "comparable value and the set it gives", a: valueSource{id: 1, s: std}, b: std, want: false},
		{name: "incomparable value with an equal value", a: slice, b: slice2, want: false},
		{name: "incomparable value with itself", a: slice, b: slice, want: false},
		{name: "incomparable value and the set it gives", a: slice, b: std, want: false},
		{name: "interface field holding a slice", a: withSlice, b: withSlice2, want: false},
		{name: "interface field holding a slice with itself", a: withSlice, b: withSlice, want: false},
		{name: "interface field holding a slice and an int", a: withSlice, b: withInt, want: false},
		{name: "interface fields holding equal ints", a: withInt, b: anySource{held: 1}, want: true},
		{name: "different kinds of tested sources", a: ptr, b: valueSource{id: 1, s: std}, want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var got, back bool
			assert.NotPanics(t, func() { got = settings.SameSource(tc.a, tc.b) })
			assert.NotPanics(t, func() { back = settings.SameSource(tc.b, tc.a) })
			assert.Equal(t, tc.want, got)
			assert.Equal(t, got, back, "SameSource is symmetric")
		})
	}

	t.Run("transitive for sets", func(t *testing.T) {
		t.Parallel()
		a, b, c := settings.Standard(), settings.Standard(), settings.Standard()
		require.True(t, settings.SameSource(a, b))
		require.True(t, settings.SameSource(b, c))
		assert.True(t, settings.SameSource(a, c))
	})
}
