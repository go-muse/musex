package settings_test

import (
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
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

// traditionalOffsets are the letter offsets C 0, D 2, E 4, F 5, G 7, A 9, B 11.
var traditionalOffsets = [letter.Count]int{0, 2, 4, 5, 7, 9, 11}

// tuningFor returns the equal temperament of sys rooted at C with A4 = 440 Hz.
func tuningFor(tb testing.TB, sys *system.System) *tuning.Tuning {
	tb.Helper()
	temp, err := temperament.EDO(sys.Classes())
	require.NoError(tb, err)
	tu, err := tuning.New(sys, temp, tuning.RootFromName(spelling.C()), tuning.A4At440())
	require.NoError(tb, err)
	return tu
}

// builtin returns the built-in notation id; ascii-updown takes its chromatic unit.
func builtin(id string, chromaticUnit ...int) notation.Notation {
	return notation.MustBuiltIn(id, chromaticUnit...)
}

// newSet creates a set and requires success.
func newSet(tb testing.TB, tu *tuning.Tuning, n notation.Notation, num number.Numbering, c *naming.Catalog) *settings.Settings {
	tb.Helper()
	s, err := settings.New(tu, n, num, c)
	require.NoError(tb, err)
	require.NotNil(tb, s)
	return s
}

// setIn returns a set with the tuning of sys (see tuningFor), notation n,
// ASPN numbering and the English catalog.
func setIn(tb testing.TB, sys *system.System, n notation.Notation) *settings.Settings {
	tb.Helper()
	return newSet(tb, tuningFor(tb, sys), n, number.ASPN(), naming.English())
}

// styled returns a custom notation with the name rules of rules and the octave
// style style. The name style uses the separator " ", other styles none.
func styled(tb testing.TB, rules notation.NameRules, style notation.OctaveStyle) notation.Notation {
	tb.Helper()
	sep := ""
	if style == notation.OctaveStyleName {
		sep = " "
	}
	n, err := notation.New(notation.Definition{Name: "custom " + style.String(), Rules: rules, OctaveStyle: style, OctaveSeparator: sep})
	require.NoError(tb, err)
	return n
}

// englishNamed is the custom notation with the name rules of english-unicode
// and the octave style "name" (no built-in notation has this style).
func englishNamed(tb testing.TB) notation.Notation {
	tb.Helper()
	return styled(tb, builtin(notation.EnglishUnicodeID).NameRules(), notation.OctaveStyleName)
}

// tableRules returns name rules made of exactly the given names, written by
// their machine form.
func tableRules(tb testing.TB, names ...spelling.Spelling) notation.NameRules {
	tb.Helper()
	forms := make([]notation.NameForm, 0, len(names))
	for _, n := range names {
		forms = append(forms, notation.NameForm{Name: n, Text: n.String()})
	}
	rules, err := notation.NewNameRules(notation.NameRulesDefinition{Names: forms})
	require.NoError(tb, err)
	return rules
}

// frac returns the alteration value num/den.
func frac(num, den int) alteration.Alteration {
	return alteration.MustFromFraction(num, den)
}

// altered returns the name with the letter of base and the alteration value num/den.
func altered(base spelling.Spelling, num, den int) spelling.Spelling {
	return base.WithAlteration(frac(num, den))
}

// at returns the pitch with the name name and the Octave with ASPN number aspn.
func at(name spelling.Spelling, aspn int) pitch.Pitch {
	return pitch.MustFromASPN(name, aspn)
}

// customSystem returns the system with the traditional offsets, the
// chromatic unit c and the rule rule.
func customSystem(tb testing.TB, c int, rule *system.Rule) *system.System {
	tb.Helper()
	sys, err := system.New(system.Definition{
		Name:        "custom",
		Classes:     12,
		Boundary:    letter.C,
		Offsets:     traditionalOffsets,
		Chromatic:   c,
		Alterations: rule,
	})
	require.NoError(tb, err)
	return sys
}

// k7System is the user system of criterion K-7: N = 12, b = C, offsets C 0,
// D 3, E 4, F 5, G 7, A 9, B 11, c = 1, A = Z.
func k7System(tb testing.TB) *system.System {
	tb.Helper()
	sys, err := system.New(system.Definition{
		Name:        "k7",
		Classes:     12,
		Boundary:    letter.C,
		Offsets:     [letter.Count]int{0, 3, 4, 5, 7, 9, 11},
		Chromatic:   1,
		Alterations: system.Integers(),
	})
	require.NoError(tb, err)
	return sys
}

// panicValue calls f and returns the value it panicked with, or nil.
func panicValue(f func()) (v any) {
	defer func() { v = recover() }()
	f()
	return nil
}

// assertPanicsWith asserts that f panics with an error that matches target.
func assertPanicsWith(t *testing.T, target error, f func()) {
	t.Helper()
	v := panicValue(f)
	if !assert.NotNil(t, v, "expected a panic") {
		return
	}
	err, ok := v.(error)
	if assert.True(t, ok, "the panic value %v is not an error", v) {
		assert.ErrorIs(t, err, target)
	}
}

// assertStructureViolation checks criterion K-18 for a structure
// counterexample: a name admissible in prev and not admissible in next.
func assertStructureViolation(t *testing.T, prev, next *system.System, v settings.Violation) {
	t.Helper()
	assert.Equal(t, settings.RuleStructure, v.Rule)
	if !assert.True(t, v.Name.IsValid(), "a structure counterexample is a name") {
		return
	}
	assert.False(t, v.Pitch.IsValid(), "a name counterexample has no pitch")
	assert.True(t, prev.MustAdmits(v.Name), "%v must be admissible in the previous system", v.Name)
	assert.False(t, next.MustAdmits(v.Name), "%v must not be admissible in the next system", v.Name)
}

// assertDesignationsViolation checks criterion K-18 for a designations
// counterexample: exactly one of a name or a pitch, expressible by prev and
// not by next.
func assertDesignationsViolation(t *testing.T, prev, next *settings.Settings, v settings.Violation) {
	t.Helper()
	assert.Equal(t, settings.RuleDesignations, v.Rule)
	if !assert.NotEqual(t, v.Name.IsValid(), v.Pitch.IsValid(), "exactly one of Name and Pitch is set") {
		return
	}
	if v.Name.IsValid() {
		assert.True(t, prev.MustExpresses(v.Name), "%v must be expressible by the previous set", v.Name)
		assert.False(t, next.MustExpresses(v.Name), "%v must not be expressible by the next set", v.Name)
		return
	}
	assert.True(t, prev.MustExpressesPitch(v.Pitch), "%v must be expressible by the previous set", v.Pitch)
	assert.False(t, next.MustExpressesPitch(v.Pitch), "%v must not be expressible by the next set", v.Pitch)
}

// nameBreaks reports whether some name of the domain of the notation of prev
// is expressible by prev and not by next. Then the designations
// counterexample must be a name.
func nameBreaks(prev, next *settings.Settings) bool {
	for _, n := range prev.Notation().NameRules().Names() {
		if prev.MustExpresses(n) && !next.MustExpresses(n) {
			return true
		}
	}
	return false
}

// assertReport checks the verdict and the counterexamples of a report of
// a check that covered exactly the rules rules.
func assertReport(t *testing.T, prev, next *settings.Settings, r settings.Report, rules ...settings.Rule) {
	t.Helper()
	compatible := true
	for _, rule := range []settings.Rule{settings.RuleStructure, settings.RuleDesignations} {
		checked := false
		for _, want := range rules {
			checked = checked || want == rule
		}
		assert.Equal(t, checked, r.Checked(rule), "Checked(%v)", rule)
		v, ok := r.Violation(rule)
		if !checked {
			assert.False(t, ok, "an unchecked rule has no violation")
			assert.Equal(t, settings.Violation{}, v)
			continue
		}
		switch rule {
		case settings.RuleStructure:
			assert.Equal(t, !prev.System().MustNamesSubsetOf(next.System()), ok, "structure violated")
			if ok {
				assertStructureViolation(t, prev.System(), next.System(), v)
			}
		case settings.RuleDesignations:
			if nameBreaks(prev, next) {
				assert.True(t, ok, "a name breaks the designations")
				assert.True(t, v.Name.IsValid(), "the counterexample must be a name when a name breaks")
			}
			if ok {
				assertDesignationsViolation(t, prev, next, v)
			}
		}
		compatible = compatible && !ok
	}
	assert.Equal(t, compatible, r.Compatible())
	if compatible {
		assert.Empty(t, r.Violations())
		assert.NoError(t, r.Err())
	} else {
		assert.NotEmpty(t, r.Violations())
		assert.ErrorIs(t, r.Err(), settings.ErrIncompatible)
		assert.NotErrorIs(t, r.Err(), settings.ErrUnset)
	}
}

// pointerSource is a source implemented by a pointer, as an instrument would.
// Its Settings method fails the test: SameSource must not read sources.
type pointerSource struct {
	t *testing.T
	s *settings.Settings
}

func (p *pointerSource) Settings() *settings.Settings {
	if p.t != nil {
		p.t.Error("SameSource must not call Settings")
	}
	return p.s
}

// valueSource is a source implemented by a comparable value.
type valueSource struct {
	id int
	s  *settings.Settings
}

func (v valueSource) Settings() *settings.Settings { return v.s }

// sliceSource is a source implemented by a value that is not comparable.
type sliceSource struct {
	sets []*settings.Settings
}

func (v sliceSource) Settings() *settings.Settings { return v.sets[0] }

// anySource is a comparable struct whose interface field can hold a value
// that is not comparable; == on two such values panics at run time.
type anySource struct {
	held any
}

func (v anySource) Settings() *settings.Settings { return nil }
