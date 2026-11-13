package settings_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/notation"
	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/octave/number"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var allRules = []settings.Rule{settings.RuleStructure, settings.RuleDesignations}

// TestCheckSystems covers K-12, K-19 and A 5.8 for two systems.
func TestCheckSystems(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name       string
		prev, next *system.System
		compatible bool
	}{
		{name: "K-12 traditional to 24-quarter-tone", prev: system.Traditional(), next: system.QuarterTones24(), compatible: true},
		{name: "K-12 24-quarter-tone to traditional", prev: system.QuarterTones24(), next: system.Traditional(), compatible: false},
		{name: "K-12 traditional to 19-position", prev: system.Traditional(), next: system.Positions19(), compatible: true},
		{name: "same system", prev: system.Traditional(), next: system.Traditional(), compatible: true},
		{name: "quarter tones to thirds", prev: system.QuarterTones24(), next: system.UpsAndDowns22(), compatible: false},
		{name: "sixths include thirds", prev: system.UpsAndDowns22(), next: system.ManeriSims72(), compatible: true},
		{name: "sixths to quarters", prev: system.ManeriSims72(), next: system.ArelEzgiUzdilek53(), compatible: false},
		{name: "K-7 user system", prev: system.Traditional(), next: k7System(t), compatible: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r, err := settings.CheckSystems(tc.prev, tc.next)
			require.NoError(t, err)
			assertSystemsReport(t, tc.prev, tc.next, r)
			assert.Equal(t, tc.compatible, r.Compatible())
			again, err := settings.CheckSystems(tc.prev, tc.next)
			require.NoError(t, err)
			assert.Equal(t, r, again, "the result is deterministic")
			assert.True(t, r == again, "equal reports are ==")
		})
	}
}

// assertSystemsReport checks a report of CheckSystems: only the structure is
// checked, the verdict is !NamesSubsetOf, and the counterexample is exact.
func assertSystemsReport(t *testing.T, prev, next *system.System, r settings.Report) {
	t.Helper()
	assert.True(t, r.Checked(settings.RuleStructure))
	assert.False(t, r.Checked(settings.RuleDesignations))
	subset := prev.MustNamesSubsetOf(next)
	assert.Equal(t, subset, r.Compatible())
	v, ok := r.Violation(settings.RuleStructure)
	assert.Equal(t, !subset, ok)
	if ok {
		assertStructureViolation(t, prev, next, v)
		assert.Equal(t, []settings.Violation{v}, r.Violations())
		assert.ErrorIs(t, r.Err(), settings.ErrIncompatible)
	} else {
		assert.Empty(t, r.Violations())
		assert.NoError(t, r.Err())
	}
	_, ok = r.Violation(settings.RuleDesignations)
	assert.False(t, ok)
}

// TestCheckSystemsErrors covers K-19: an unset system is a refusal, not a verdict.
func TestCheckSystemsErrors(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name       string
		prev, next *system.System
	}{
		{name: "nil previous", prev: nil, next: system.Traditional()},
		{name: "nil next", prev: system.Traditional(), next: nil},
		{name: "zero previous", prev: &system.System{}, next: system.Traditional()},
		{name: "zero next", prev: system.Traditional(), next: &system.System{}},
		{name: "both nil", prev: nil, next: nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r, err := settings.CheckSystems(tc.prev, tc.next)
			assert.ErrorIs(t, err, settings.ErrUnset)
			assert.Equal(t, settings.Report{}, r)
			assertPanicsWith(t, settings.ErrUnset, func() { settings.MustCheckSystems(tc.prev, tc.next) })
		})
	}
}

// compatCase is a pair of sets for the compatibility checks.
type compatCase struct {
	name       string
	prev, next *settings.Settings
	structure  bool // the structure rule holds
	design     bool // the designations rule holds
	// designName and designPitch require the kind of the designations
	// counterexample when it is not empty.
	designName, designPitch bool
}

// compatCases lists the pairs of criteria K-12 to K-17 and the octave-set
// branches of A 6.4.
func compatCases(t *testing.T) []compatCase {
	t.Helper()
	std := settings.Standard()
	eu := builtin(notation.EnglishUnicodeID)
	wide := std.MustWithTuning(tuningFor(t, system.QuarterTones24()))
	pos19 := std.MustWithTuning(tuningFor(t, system.Positions19()))
	russian := std.MustWithNotation(builtin(notation.RussianAMID))
	russian19 := russian.MustWithTuning(tuningFor(t, system.Positions19()))
	named := std.MustWithNotation(englishNamed(t)).MustWithCatalog(naming.Russian())
	namedEnglish := named.MustWithCatalog(naming.English())
	return []compatCase{
		{name: "G-4 same set", prev: std, next: settings.Standard(), structure: true, design: true},
		{name: "G-3 reference 440 to 415", prev: std, next: std.MustWithTuning(tuning.Standard().MustWithReferenceFrequency(415)), structure: true, design: true},
		{name: "K-12 traditional to 24-quarter-tone", prev: std, next: wide, structure: true, design: true},
		{name: "K-12 24-quarter-tone to traditional", prev: wide, next: std, structure: false, design: true},
		{name: "K-12 traditional to 19-position", prev: std, next: pos19, structure: true, design: true},
		{name: "K-13 ASPN to Franco-Belgian", prev: std, next: std.MustWithNumbering(number.FrancoBelgian()), structure: true, design: true},
		{name: "K-13 English Unicode to ABC", prev: std, next: std.MustWithNotation(builtin(notation.ABCID)), structure: true, design: true},
		{name: "K-13 English Unicode to German AM", prev: std, next: std.MustWithNotation(builtin(notation.GermanAMID)), structure: true, design: false, designName: true},
		{name: "K-13 English Unicode to Romance Italian", prev: std, next: std.MustWithNotation(builtin(notation.RomanceItalianID)), structure: true, design: false, designPitch: true},
		{name: "K-13 visible number, catalog to empty", prev: std, next: std.MustWithCatalog(nil), structure: true, design: true},
		{name: "K-13 name style, Russian catalog to empty", prev: named, next: named.MustWithCatalog(nil), structure: true, design: false, designPitch: true},
		{name: "K-14 English Unicode, traditional to 19-position", prev: std, next: pos19, structure: true, design: true},
		{name: "K-15 Russian AM, traditional to 19-position", prev: russian, next: russian19, structure: true, design: false, designPitch: true},
		{name: "K-15 Russian AM, 19-position to traditional", prev: russian19, next: russian, structure: true, design: false, designPitch: true},
		{
			name:      "K-16 structure and designations",
			prev:      setIn(t, system.QuarterTones24(), builtin(notation.LilyPondEnglishID)),
			next:      setIn(t, system.Traditional(), builtin(notation.GermanAMID)),
			structure: false, design: false, designName: true,
		},
		{name: "visible number to AM", prev: std, next: russian, structure: true, design: false, designPitch: true},
		{name: "visible number to a catalog", prev: std, next: named, structure: true, design: false, designPitch: true},
		{name: "AM to visible number", prev: russian, next: std, structure: true, design: true},
		{name: "AM to a smaller catalog", prev: russian, next: namedEnglish, structure: true, design: false, designPitch: true},
		{name: "AM to a covering catalog", prev: russian, next: named, structure: true, design: true},
		{name: "catalog to AM", prev: named, next: russian, structure: true, design: false, designPitch: true},
		{name: "catalog to a visible number", prev: named, next: std, structure: true, design: true},
		{name: "catalog to a subset catalog", prev: namedEnglish, next: named, structure: true, design: true},
		{name: "name-only notation to anything", prev: std.MustWithNotation(builtin(notation.RomanceItalianID)), next: std.MustWithNotation(builtin(notation.ABCID)), structure: true, design: true},
		{name: "LilyPond to ABC", prev: setIn(t, system.Traditional(), builtin(notation.LilyPondEnglishID)), next: setIn(t, system.Traditional(), builtin(notation.ABCID)), structure: true, design: true},
		{name: "wider domain to narrower", prev: setIn(t, system.QuarterTones24(), builtin(notation.LilyPondEnglishID)), next: setIn(t, system.QuarterTones24(), eu), structure: true, design: false, designName: true},
	}
}

// TestCheckReplacement covers K-12 to K-18 and A 5.8: the verdict of both
// rules, the reasons and exact counterexamples.
func TestCheckReplacement(t *testing.T) {
	t.Parallel()
	for _, tc := range compatCases(t) {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r, err := tc.prev.CheckReplacement(tc.next)
			require.NoError(t, err)
			assertReport(t, tc.prev, tc.next, r, allRules...)
			assert.Equal(t, tc.structure && tc.design, r.Compatible())
			_, broken := r.Violation(settings.RuleStructure)
			assert.Equal(t, !tc.structure, broken, "structure violated")
			v, broken := r.Violation(settings.RuleDesignations)
			assert.Equal(t, !tc.design, broken, "designations violated")
			if broken && tc.designName {
				assert.True(t, v.Name.IsValid(), "the designations counterexample is a name")
			}
			if broken && tc.designPitch {
				assert.True(t, v.Pitch.IsValid(), "the designations counterexample is a pitch")
			}

			again, err := tc.prev.CheckReplacement(tc.next)
			require.NoError(t, err)
			assert.True(t, r == again, "the result is deterministic")
			assert.Equal(t, r, tc.prev.MustCheckReplacement(tc.next))

			violations := r.Violations()
			var rules []settings.Rule
			for _, v := range violations {
				rules = append(rules, v.Rule)
			}
			switch {
			case !tc.structure && !tc.design:
				assert.Equal(t, allRules, rules, "violations in rule order")
			case !tc.structure:
				assert.Equal(t, []settings.Rule{settings.RuleStructure}, rules)
			case !tc.design:
				assert.Equal(t, []settings.Rule{settings.RuleDesignations}, rules)
			default:
				assert.Empty(t, rules)
			}
		})
	}
}

// TestPartialChecks covers A 5.8: CheckStructure and CheckDesignations check
// one rule each and say nothing about the other.
func TestPartialChecks(t *testing.T) {
	t.Parallel()
	for _, tc := range compatCases(t) {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			structure, err := tc.prev.CheckStructure(tc.next)
			require.NoError(t, err)
			assertReport(t, tc.prev, tc.next, structure, settings.RuleStructure)
			assert.Equal(t, tc.structure, structure.Compatible())
			assert.Equal(t, structure, tc.prev.MustCheckStructure(tc.next))

			bySystems, err := settings.CheckSystems(tc.prev.System(), tc.next.System())
			require.NoError(t, err)
			assert.Equal(t, bySystems.Compatible(), structure.Compatible())

			design, err := tc.prev.CheckDesignations(tc.next)
			require.NoError(t, err)
			assertReport(t, tc.prev, tc.next, design, settings.RuleDesignations)
			assert.Equal(t, tc.design, design.Compatible())
			assert.Equal(t, design, tc.prev.MustCheckDesignations(tc.next))
		})
	}
}

// TestCompatK14 covers K-14 and P-14: a changed written number at the same
// name and Octave is not a violation.
func TestCompatK14(t *testing.T) {
	t.Parallel()
	prev := settings.Standard()
	next := prev.MustWithTuning(tuningFor(t, system.Positions19()))
	p := at(spelling.BSharp(), 4)
	assert.Equal(t, "B♯3", prev.MustFormatPitch(p))
	assert.Equal(t, "B♯4", next.MustFormatPitch(p))
	assert.True(t, prev.MustCheckReplacement(next).Compatible())
}

// TestCompatK16 covers K-16: both rules are broken, both reasons are given.
func TestCompatK16(t *testing.T) {
	t.Parallel()
	prev := setIn(t, system.QuarterTones24(), builtin(notation.LilyPondEnglishID))
	next := setIn(t, system.Traditional(), builtin(notation.GermanAMID))
	r := prev.MustCheckReplacement(next)
	require.Len(t, r.Violations(), 2)
	assert.Equal(t, settings.RuleStructure, r.Violations()[0].Rule)
	assert.Equal(t, settings.RuleDesignations, r.Violations()[1].Rule)
	assertStructureViolation(t, prev.System(), next.System(), r.Violations()[0])
	assertDesignationsViolation(t, prev, next, r.Violations()[1])
	assert.ErrorIs(t, r.Err(), settings.ErrIncompatible)
}

// TestCheckErrors covers K-19: checks with an unset set fail with ErrUnset.
func TestCheckErrors(t *testing.T) {
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
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			for _, check := range []struct {
				name string
				call func() (settings.Report, error)
				must func()
			}{
				{name: "CheckReplacement", call: func() (settings.Report, error) { return tc.prev.CheckReplacement(tc.next) }, must: func() { tc.prev.MustCheckReplacement(tc.next) }},
				{name: "CheckStructure", call: func() (settings.Report, error) { return tc.prev.CheckStructure(tc.next) }, must: func() { tc.prev.MustCheckStructure(tc.next) }},
				{name: "CheckDesignations", call: func() (settings.Report, error) { return tc.prev.CheckDesignations(tc.next) }, must: func() { tc.prev.MustCheckDesignations(tc.next) }},
			} {
				r, err := check.call()
				assert.ErrorIs(t, err, settings.ErrUnset, check.name)
				assert.Equal(t, settings.Report{}, r, check.name)
				assertPanicsWith(t, settings.ErrUnset, check.must)
			}
		})
	}
}

// TestReportValues covers A 5.2 and A 5.8: the zero report, the methods of
// reports, invalid rules and the ownership of Violations.
func TestReportValues(t *testing.T) {
	t.Parallel()
	std := settings.Standard()
	wide := std.MustWithTuning(tuningFor(t, system.QuarterTones24()))
	compatible := std.MustCheckReplacement(wide)
	incompatible := setIn(t, system.QuarterTones24(), builtin(notation.LilyPondEnglishID)).
		MustCheckReplacement(setIn(t, system.Traditional(), builtin(notation.GermanAMID)))
	require.True(t, compatible.Compatible())
	require.False(t, incompatible.Compatible())

	t.Run("zero report", func(t *testing.T) {
		t.Parallel()
		var r settings.Report
		for _, rule := range allRules {
			assert.False(t, r.Checked(rule))
			v, ok := r.Violation(rule)
			assert.False(t, ok)
			assert.Equal(t, settings.Violation{}, v)
		}
		assert.False(t, r.Compatible())
		assert.Empty(t, r.Violations())
		assert.ErrorIs(t, r.Err(), settings.ErrUnset)
		assert.NotErrorIs(t, r.Err(), settings.ErrIncompatible)
		assert.NotPanics(t, func() { _ = r.String() })
	})

	for _, tc := range []struct {
		name string
		r    settings.Report
	}{
		{name: "zero", r: settings.Report{}},
		{name: "compatible", r: compatible},
		{name: "incompatible", r: incompatible},
	} {
		t.Run("invalid rules in a "+tc.name+" report", func(t *testing.T) {
			t.Parallel()
			for _, rule := range []settings.Rule{settings.RuleUnset, settings.Rule(3), settings.Rule(9), settings.Rule(255)} {
				assert.False(t, tc.r.Checked(rule))
				var (
					v  settings.Violation
					ok bool
				)
				assert.NotPanics(t, func() { v, ok = tc.r.Violation(rule) })
				assert.False(t, ok)
				assert.Equal(t, settings.Violation{}, v)
			}
			assert.NotPanics(t, func() { _ = tc.r.String() })
		})
	}

	t.Run("Violations returns a new slice", func(t *testing.T) {
		t.Parallel()
		first := incompatible.Violations()
		require.Len(t, first, 2)
		want := append([]settings.Violation(nil), first...)
		first[0] = settings.Violation{}
		first[1] = settings.Violation{Rule: settings.RuleStructure}
		assert.Equal(t, want, incompatible.Violations())
		v, ok := incompatible.Violation(settings.RuleStructure)
		assert.True(t, ok)
		assert.Equal(t, want[0], v)
		v, ok = incompatible.Violation(settings.RuleDesignations)
		assert.True(t, ok)
		assert.Equal(t, want[1], v)
	})

	t.Run("compatible report", func(t *testing.T) {
		t.Parallel()
		assert.True(t, compatible.Checked(settings.RuleStructure))
		assert.True(t, compatible.Checked(settings.RuleDesignations))
		assert.Empty(t, compatible.Violations())
		assert.NoError(t, compatible.Err())
		assert.NotPanics(t, func() { _ = compatible.String() })
	})

	t.Run("reports compare by ==", func(t *testing.T) {
		t.Parallel()
		assert.False(t, compatible == incompatible)
		assert.False(t, compatible == settings.Report{})
		assert.True(t, compatible == std.MustCheckReplacement(wide))
	})
}

// TestRuleString covers A 5.2: String of rules does not panic; its text is
// not part of the contract.
func TestRuleString(t *testing.T) {
	t.Parallel()
	seen := map[string]settings.Rule{}
	for _, rule := range []settings.Rule{settings.RuleUnset, settings.RuleStructure, settings.RuleDesignations, settings.Rule(9)} {
		var s string
		require.NotPanics(t, func() { s = rule.String() })
		assert.NotEmpty(t, s)
		if prior, dup := seen[s]; dup {
			assert.Failf(t, "duplicate string", "%v and %v share %q", uint8(prior), uint8(rule), s)
		}
		seen[s] = rule
	}
	assert.NotPanics(t, func() { _ = settings.Violation{}.String() })
	assert.NotPanics(t, func() {
		_ = settings.Violation{Rule: settings.RuleStructure, Name: spelling.CSharp()}.String()
	})
}

// structureCase is a pair of systems with user rules for the exact search
// of a structure counterexample (A 6.3).
type structureCase struct {
	name       string
	prev, next *system.System
	compatible bool
}

// structureCases builds systems with user rules: the built-in systems never
// exercise the exact search because their differences show up at |a| <= 2.
func structureCases(t *testing.T) []structureCase {
	t.Helper()
	const m = alteration.MaxMagnitude
	step := frac(1, 1<<16)
	fine := system.MustNewProgression(alteration.Natural(), step)
	split := system.MustNewRule(
		fine.MustWithUpper(frac(3<<16, 1<<16)),
		fine.MustWithLower(frac(3<<16+2, 1<<16)),
	)
	// Eight lattices (1/8192)Z + j/65536 together are (1/65536)Z; the lattice
	// j = 5 is cut above 3, so its points above 3 are missing.
	var lattices []system.Progression
	for j := range 8 {
		p := system.MustNewProgression(frac(j, 1<<16), frac(1, 1<<13))
		if j == 5 {
			p = p.MustWithUpper(alteration.MustFromInt(3))
		}
		lattices = append(lattices, p)
	}
	shifted := system.MustNewRule(lattices...)
	half := alteration.MustFromFraction(1, 2)
	withSeven := system.MustUnion(system.Integers(), system.Single(frac(7, 2)))
	withFive := system.MustUnion(system.Integers(), system.Single(frac(5, 2)))
	withBoth := system.MustUnion(system.Integers(), system.Single(frac(7, 2)), system.Single(frac(5, 2)))
	lowEdge := system.MustIntegerRange(-m+1, m)
	highEdge := system.MustIntegerRange(-m, m-1)
	ints := func(rule *system.Rule) *system.System { return customSystem(t, 1, rule) }
	halves := func(rule *system.Rule) *system.System { return customSystem(t, 2, rule) }
	fines := func(rule *system.Rule) *system.System { return customSystem(t, 1<<16, rule) }
	return []structureCase{
		{name: "difference only beyond |a| = 2", prev: ints(system.Integers()), next: ints(system.MustIntegerRange(-2, 2)), compatible: false},
		{name: "difference only at -2^24", prev: ints(system.Integers()), next: ints(lowEdge), compatible: false},
		{name: "difference only at +2^24", prev: ints(system.Integers()), next: ints(highEdge), compatible: false},
		{name: "single missing point of a fine lattice", prev: fines(system.MustMultiples(1 << 16)), next: fines(split), compatible: false},
		{name: "shifted lattices with one cut", prev: fines(system.MustMultiples(1 << 16)), next: fines(shifted), compatible: false},
		{name: "single values differ", prev: halves(withSeven), next: halves(withFive), compatible: false},
		{name: "single value in a progression", prev: halves(system.MustUnion(system.Integers(), system.Single(half))), next: halves(system.MustMultiples(2)), compatible: true},
		{name: "included range", prev: ints(system.MustIntegerRange(-2, 2)), next: ints(system.Integers()), compatible: true},
		{name: "included edge range", prev: ints(lowEdge), next: ints(system.Integers()), compatible: true},
		{name: "split lattice in the full lattice", prev: fines(split), next: fines(system.MustMultiples(1 << 16)), compatible: true},
		{name: "cut lattices in the full lattice", prev: fines(shifted), next: fines(system.MustMultiples(1 << 16)), compatible: true},
		{name: "single value included", prev: halves(withSeven), next: halves(withBoth), compatible: true},
		{name: "coarse lattice in the shifted lattices", prev: fines(system.MustMultiples(1 << 13)), next: fines(shifted), compatible: true},
	}
}

// TestStructureCounterexamples covers A 6.3 and K-18: the verdict equals
// !NamesSubsetOf, and every counterexample is admissible in the previous
// system and not in the next.
func TestStructureCounterexamples(t *testing.T) {
	t.Parallel()
	for _, tc := range structureCases(t) {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.compatible, tc.prev.MustNamesSubsetOf(tc.next), "fixture")
			r, err := settings.CheckSystems(tc.prev, tc.next)
			require.NoError(t, err)
			assertSystemsReport(t, tc.prev, tc.next, r)
			assert.Equal(t, tc.compatible, r.Compatible())

			prevSet := setIn(t, tc.prev, builtin(notation.EnglishUnicodeID))
			nextSet := setIn(t, tc.next, builtin(notation.EnglishUnicodeID))
			structure, err := prevSet.CheckStructure(nextSet)
			require.NoError(t, err)
			assert.Equal(t, tc.compatible, structure.Compatible())
			if v, ok := structure.Violation(settings.RuleStructure); ok {
				assertStructureViolation(t, tc.prev, tc.next, v)
			}
		})
	}
}

// TestDesignationsExtremeCatalogs covers A 6.4 and K-18: pitches with an
// Octave outside int do not exist and give no counterexample.
func TestDesignationsExtremeCatalogs(t *testing.T) {
	t.Parallel()
	std := settings.Standard()
	top := naming.MustNew(map[int]string{math.MaxInt: "top"})
	bottom := naming.MustNew(map[int]string{math.MinInt: "bottom"})
	design := func(rules notation.NameRules, style notation.OctaveStyle, cat *naming.Catalog) *settings.Settings {
		return std.MustWithDesignations(styled(t, rules, style), number.ASPN(), cat)
	}
	bSharp := tableRules(t, spelling.BSharp())
	bSharpAndC := tableRules(t, spelling.BSharp(), spelling.C())
	cFlat := tableRules(t, spelling.CFlat())
	cFlatAndC := tableRules(t, spelling.CFlat(), spelling.C())
	for _, tc := range []struct {
		name       string
		prev, next *settings.Settings
		compatible bool
	}{
		{name: "only the largest octave named, carry +1", prev: design(bSharp, notation.OctaveStyleName, top), next: design(bSharp, notation.OctaveStyleName, nil), compatible: true},
		{name: "only the largest octave named, carry +1, to AM", prev: design(bSharp, notation.OctaveStyleName, top), next: design(bSharp, notation.OctaveStyleAM, nil), compatible: true},
		{name: "only the largest octave named, carry 0 too", prev: design(bSharpAndC, notation.OctaveStyleName, top), next: design(bSharpAndC, notation.OctaveStyleName, nil), compatible: false},
		{name: "only the smallest octave named, carry -1", prev: design(cFlat, notation.OctaveStyleName, bottom), next: design(cFlat, notation.OctaveStyleAM, nil), compatible: true},
		{name: "only the smallest octave named, carry -1, to empty", prev: design(cFlat, notation.OctaveStyleName, bottom), next: design(cFlat, notation.OctaveStyleName, nil), compatible: true},
		{name: "only the smallest octave named, carry 0 too", prev: design(cFlatAndC, notation.OctaveStyleName, bottom), next: design(cFlatAndC, notation.OctaveStyleAM, nil), compatible: false},
		{name: "largest octave named in both", prev: design(bSharpAndC, notation.OctaveStyleName, top), next: design(bSharpAndC, notation.OctaveStyleName, top), compatible: true},
		{name: "largest octave to a visible number", prev: design(bSharpAndC, notation.OctaveStyleName, top), next: design(bSharpAndC, notation.OctaveStyleNumber, nil), compatible: true},
		{name: "visible number to the largest octave", prev: design(bSharpAndC, notation.OctaveStyleNumber, nil), next: design(bSharpAndC, notation.OctaveStyleName, top), compatible: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r, err := tc.prev.CheckDesignations(tc.next)
			require.NoError(t, err)
			assertReport(t, tc.prev, tc.next, r, settings.RuleDesignations)
			assert.Equal(t, tc.compatible, r.Compatible())
			if v, ok := r.Violation(settings.RuleDesignations); ok {
				assert.True(t, v.Pitch.IsValid(), "names agree; the counterexample is a pitch")
			}
			full, err := tc.prev.CheckReplacement(tc.next)
			require.NoError(t, err)
			assert.Equal(t, tc.compatible, full.Compatible())
		})
	}
}
