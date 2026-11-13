package settings_test

import (
	"testing"

	"github.com/go-muse/musex/notation"
	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/octave/number"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"
	"github.com/go-muse/musex/tuning/temperament"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestChanges covers K-17, P-13 and A 5.9: the kinds of a change, as a full list.
func TestChanges(t *testing.T) {
	t.Parallel()
	std := settings.Standard()
	t24 := tuningFor(t, system.QuarterTones24())
	t19 := tuningFor(t, system.Positions19())
	wide := std.MustWithTuning(t24)
	for _, tc := range []struct {
		name       string
		prev, next *settings.Settings
		want       settings.Changes
		compatible bool
	}{
		{name: "K-17 reference 440 to 415", prev: std, next: std.MustWithTuning(tuning.Standard().MustWithReferenceFrequency(415)), want: settings.Changes{Structure: settings.StructureUnchanged, Sound: true}, compatible: true},
		{name: "K-17 same set", prev: std, next: settings.Standard(), want: settings.Changes{Structure: settings.StructureUnchanged}, compatible: true},
		{name: "K-17 12-EDO traditional to 24-EDO quarter tones", prev: std, next: wide, want: settings.Changes{Structure: settings.StructureExtended, Sound: true}, compatible: true},
		{name: "K-17 24-EDO quarter tones to 12-EDO traditional", prev: wide, next: std, want: settings.Changes{Structure: settings.StructureNarrowed, Sound: true}, compatible: false},
		{name: "K-17 12-EDO traditional to 19-EDO 19-position", prev: std, next: std.MustWithTuning(t19), want: settings.Changes{Structure: settings.StructureSameNames, Sound: true}, compatible: true},
		{name: "K-17 WithSystem to the user system of K-7", prev: std, next: std.MustWithSystem(k7System(t)), want: settings.Changes{Structure: settings.StructureSameNames}, compatible: true},
		{name: "K-17 notation ABC", prev: std, next: std.MustWithNotation(builtin(notation.ABCID)), want: settings.Changes{Structure: settings.StructureUnchanged, Designations: true}, compatible: true},
		{name: "numbering", prev: std, next: std.MustWithNumbering(number.FrancoBelgian()), want: settings.Changes{Structure: settings.StructureUnchanged, Designations: true}, compatible: true},
		{name: "equal numbering", prev: std.MustWithNumbering(number.Mexican()), next: std.MustWithNumbering(number.MiddleC5()), want: settings.Changes{Structure: settings.StructureUnchanged}, compatible: true},
		{name: "catalog", prev: std, next: std.MustWithCatalog(naming.Russian()), want: settings.Changes{Structure: settings.StructureUnchanged, Designations: true}, compatible: true},
		{name: "root number", prev: std, next: std.MustWithTuning(tuning.Standard().MustWithRoot(tuning.RootFromName(spelling.D()))), want: settings.Changes{Structure: settings.StructureUnchanged, Sound: true}, compatible: true},
		{name: "temperament of the same N", prev: std, next: std.MustWithTuning(tuning.Standard().MustWithTemperament(temperament.Just())), want: settings.Changes{Structure: settings.StructureUnchanged, Sound: true}, compatible: true},
		{name: "reference note C5", prev: std, next: std.MustWithTuning(tuning.Standard().MustWithReference(tuning.C5At523())), want: settings.Changes{Structure: settings.StructureUnchanged, Sound: true}, compatible: true},
		{
			name: "incomparable names", prev: std.MustWithTuning(t24), next: std.MustWithTuning(tuningFor(t, system.UpsAndDowns22())),
			want: settings.Changes{Structure: settings.StructureIncomparable, Sound: true}, compatible: false,
		},
		{
			name: "structure, sound and designations", prev: std, next: wide.MustWithNotation(builtin(notation.LilyPondEnglishID)),
			want: settings.Changes{Structure: settings.StructureExtended, Sound: true, Designations: true}, compatible: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := tc.prev.Changes(tc.next)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
			assert.True(t, got.IsValid())
			assert.Equal(t, tc.prev.Equal(tc.next), got.None())
			assert.Equal(t, got, tc.prev.MustChanges(tc.next))
			assert.NotPanics(t, func() { _ = got.String() })
			assert.Equal(t, tc.compatible, tc.prev.MustCheckReplacement(tc.next).Compatible())
		})
	}
}

// TestChangesNoneIsEqual covers A 5.9: for set sets None ⇔ Equal, on all pairs.
func TestChangesNoneIsEqual(t *testing.T) {
	t.Parallel()
	std := settings.Standard()
	sets := []*settings.Settings{
		std,
		settings.Standard(),
		std.MustWithTuning(tuning.Standard().MustWithReferenceFrequency(415)),
		std.MustWithTuning(tuningFor(t, system.QuarterTones24())),
		std.MustWithSystem(k7System(t)),
		std.MustWithNotation(builtin(notation.GermanAMID)),
		std.MustWithNumbering(number.Mexican()),
		std.MustWithNumbering(number.MiddleC5()),
		std.MustWithCatalog(nil),
		std.MustWithCatalog(&naming.Catalog{}),
	}
	for i, a := range sets {
		for j, b := range sets {
			c, err := a.Changes(b)
			require.NoError(t, err)
			assert.True(t, c.IsValid())
			assert.Equal(t, a.Equal(b), c.None(), "sets %d and %d", i, j)
			assert.Equal(t, c.Structure == settings.StructureUnchanged, a.System().Equal(b.System()), "sets %d and %d", i, j)
		}
	}
}

// TestChangesValues covers A 5.2 and A 5.9: the zero value and invalid values
// of Changes and StructureChange.
func TestChangesValues(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name  string
		c     settings.Changes
		valid bool
		none  bool
	}{
		{name: "zero", c: settings.Changes{}, valid: false, none: false},
		{name: "unset structure with sound", c: settings.Changes{Sound: true}, valid: false, none: false},
		{name: "out of range structure", c: settings.Changes{Structure: settings.StructureChange(9)}, valid: false, none: false},
		{name: "unchanged", c: settings.Changes{Structure: settings.StructureUnchanged}, valid: true, none: true},
		{name: "unchanged with sound", c: settings.Changes{Structure: settings.StructureUnchanged, Sound: true}, valid: true, none: false},
		{name: "unchanged with designations", c: settings.Changes{Structure: settings.StructureUnchanged, Designations: true}, valid: true, none: false},
		{name: "extended", c: settings.Changes{Structure: settings.StructureExtended}, valid: true, none: false},
		{name: "narrowed", c: settings.Changes{Structure: settings.StructureNarrowed}, valid: true, none: false},
		{name: "same names", c: settings.Changes{Structure: settings.StructureSameNames}, valid: true, none: false},
		{name: "incomparable", c: settings.Changes{Structure: settings.StructureIncomparable}, valid: true, none: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.valid, tc.c.IsValid())
			assert.Equal(t, tc.none, tc.c.None())
			assert.NotPanics(t, func() { _ = tc.c.String() })
		})
	}

	t.Run("StructureChange strings", func(t *testing.T) {
		t.Parallel()
		seen := map[string]bool{}
		for _, c := range []settings.StructureChange{
			settings.StructureUnset, settings.StructureUnchanged, settings.StructureExtended,
			settings.StructureNarrowed, settings.StructureSameNames, settings.StructureIncomparable,
			settings.StructureChange(9),
		} {
			var s string
			require.NotPanics(t, func() { s = c.String() })
			assert.NotEmpty(t, s)
			assert.False(t, seen[s], "duplicate string %q", s)
			seen[s] = true
		}
	})
}

// TestChangesErrors covers A 5.9: an unset set is a refusal.
func TestChangesErrors(t *testing.T) {
	t.Parallel()
	std := settings.Standard()
	var unsetNil *settings.Settings
	for _, tc := range []struct {
		name       string
		prev, next *settings.Settings
	}{
		{name: "nil previous", prev: unsetNil, next: std},
		{name: "nil next", prev: std, next: unsetNil},
		{name: "zero previous", prev: &settings.Settings{}, next: std},
		{name: "zero next", prev: std, next: &settings.Settings{}},
		{name: "both unset", prev: unsetNil, next: &settings.Settings{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c, err := tc.prev.Changes(tc.next)
			assert.ErrorIs(t, err, settings.ErrUnset)
			assert.Equal(t, settings.Changes{}, c)
			assertPanicsWith(t, settings.ErrUnset, func() { tc.prev.MustChanges(tc.next) })
		})
	}
}
