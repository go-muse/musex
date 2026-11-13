package settings

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-muse/musex/notation"
	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

// Rule is a rule of compatibility of a replacement of one set of settings by
// another.
type Rule uint8

// Rules. The numbers of the constants are stable.
const (
	// RuleUnset is the zero value: no rule.
	RuleUnset Rule = iota
	// RuleStructure: every name admissible in the previous system is
	// admissible in the next one.
	RuleStructure
	// RuleDesignations: every name and every pitch expressible by the
	// previous set is expressible by the next one.
	RuleDesignations
)

// String returns "structure" or "designations", "Rule(unset)" for RuleUnset
// and "Rule(<n>)" for other values. It is a debugging form.
func (r Rule) String() string {
	switch r {
	case RuleUnset:
		return "Rule(unset)"
	case RuleStructure:
		return "structure"
	case RuleDesignations:
		return "designations"
	default:
		return "Rule(" + strconv.Itoa(int(r)) + ")"
	}
}

// ruleBit returns the bit of a defined rule in Report.checked, or 0.
func ruleBit(r Rule) uint8 {
	switch r {
	case RuleStructure, RuleDesignations:
		return 1 << r
	default:
		return 0
	}
}

// Violation is a violated rule together with its counterexample: a name or
// a pitch expressible (for RuleStructure, admissible) by the previous set and
// not by the next one. Exactly one of Name and Pitch is set in a violation
// returned by the package. The zero Violation is no violation.
type Violation struct {
	// Rule is the violated rule.
	Rule Rule
	// Name is the counterexample name; unset when Pitch is set.
	Name spelling.Spelling
	// Pitch is the counterexample pitch; unset for a name counterexample.
	Pitch pitch.Pitch
}

// String returns a debugging form such as "structure: C𝄲" or
// "designations: B♯9", or "Violation(none)" for the zero value.
func (v Violation) String() string {
	if v == (Violation{}) {
		return "Violation(none)"
	}
	if v.Pitch.IsValid() {
		return v.Rule.String() + ": " + v.Pitch.String()
	}
	return v.Rule.String() + ": " + v.Name.String()
}

// Report is the result of a check of compatibility: which rules were
// checked and the violation of each violated rule.
//
// Report is a comparable value: two checks of the same pair give == reports.
// The zero Report checked nothing: Checked reports false for every rule,
// Compatible reports false and Err returns an error that matches ErrUnset.
type Report struct {
	checked      uint8
	structure    Violation
	designations Violation
}

// Checked reports whether the rule was checked. It is false for RuleUnset and
// for values outside the defined rules.
func (r Report) Checked(rule Rule) bool {
	bit := ruleBit(rule)
	return bit != 0 && r.checked&bit != 0
}

// Compatible reports whether at least one rule was checked and no checked
// rule is violated. A report of a partial check speaks only of its own rule.
func (r Report) Compatible() bool {
	return r.checked != 0 && r.structure.Rule == RuleUnset && r.designations.Rule == RuleUnset
}

// Violations returns a new slice with the violations in the order
// RuleStructure, RuleDesignations; nil when no rule is violated.
func (r Report) Violations() []Violation {
	var vs []Violation
	if r.structure.Rule != RuleUnset {
		vs = append(vs, r.structure)
	}
	if r.designations.Rule != RuleUnset {
		vs = append(vs, r.designations)
	}
	return vs
}

// Violation returns the violation of rule and whether the rule is violated.
// It returns (Violation{}, false) for an unviolated or unchecked rule, for
// RuleUnset and for values outside the defined rules.
func (r Report) Violation(rule Rule) (Violation, bool) {
	switch rule {
	case RuleStructure:
		return r.structure, r.structure.Rule != RuleUnset
	case RuleDesignations:
		return r.designations, r.designations.Rule != RuleUnset
	default:
		return Violation{}, false
	}
}

// Err returns nil for a compatible report and an error that matches
// ErrIncompatible and names the violations and their counterexamples for an
// incompatible one. For a report that checked nothing, such as the zero
// Report, it returns an error that matches ErrUnset.
func (r Report) Err() error {
	if r.checked == 0 {
		return unsetError("report", "nothing checked")
	}
	if r.Compatible() {
		return nil
	}
	return fmt.Errorf("%w: %s", ErrIncompatible, r.violationsText())
}

// String returns a debugging form: "Report(unchecked)", "Report(compatible:
// <rules>)" or "Report(incompatible: <violations>)".
func (r Report) String() string {
	if r.checked == 0 {
		return "Report(unchecked)"
	}
	if r.Compatible() {
		var rules []string
		for _, rule := range []Rule{RuleStructure, RuleDesignations} {
			if r.Checked(rule) {
				rules = append(rules, rule.String())
			}
		}
		return "Report(compatible: " + strings.Join(rules, ", ") + ")"
	}
	return "Report(incompatible: " + r.violationsText() + ")"
}

// violationsText joins the violations of r for messages.
func (r Report) violationsText() string {
	vs := r.Violations()
	parts := make([]string, len(vs))
	for i, v := range vs {
		parts[i] = v.String()
	}
	return strings.Join(parts, "; ")
}

// CheckSystems checks the rule RuleStructure for the replacement of the
// system prev by next: every name admissible in prev must be admissible in
// next. The counterexample of a violation is a name admissible in prev and
// not in next. The report has RuleStructure checked.
//
// It fails with ErrUnset when prev or next is unset. Incompatibility is a
// result, not an error.
func CheckSystems(prev, next *system.System) (Report, error) {
	if !prev.IsValid() || !next.IsValid() {
		return Report{}, unsetError("check systems", "system")
	}
	var r Report
	r.checked = ruleBit(RuleStructure)
	r.structure = structureViolation(prev, next)
	return r, nil
}

// MustCheckSystems is like CheckSystems but panics with its error on failure.
func MustCheckSystems(prev, next *system.System) Report {
	return must(CheckSystems(prev, next))
}

// CheckReplacement checks whether next can replace s: both RuleStructure and
// RuleDesignations. When both are violated, the report holds both
// violations. Notes are not taken into account, so a refusal can be false
// for the notes actually in use; that is not an error of the check.
//
// It fails with ErrUnset when s or next is unset.
func (s *Settings) CheckReplacement(next *Settings) (Report, error) {
	if err := checkPair("check replacement", s, next); err != nil {
		return Report{}, err
	}
	var r Report
	r.checked = ruleBit(RuleStructure) | ruleBit(RuleDesignations)
	r.structure = structureViolation(s.tuning.System(), next.tuning.System())
	r.designations = designationsViolation(s, next)
	return r, nil
}

// MustCheckReplacement is like CheckReplacement but panics with its error on
// failure.
func (s *Settings) MustCheckReplacement(next *Settings) Report {
	return must(s.CheckReplacement(next))
}

// CheckStructure checks only RuleStructure for the systems of s and next. It
// fails with ErrUnset when s or next is unset.
func (s *Settings) CheckStructure(next *Settings) (Report, error) {
	if err := checkPair("check structure", s, next); err != nil {
		return Report{}, err
	}
	var r Report
	r.checked = ruleBit(RuleStructure)
	r.structure = structureViolation(s.tuning.System(), next.tuning.System())
	return r, nil
}

// MustCheckStructure is like CheckStructure but panics with its error on
// failure.
func (s *Settings) MustCheckStructure(next *Settings) Report {
	return must(s.CheckStructure(next))
}

// CheckDesignations checks only RuleDesignations: every name and every pitch
// expressible by s must be expressible by next, each set using its own
// system. It fails with ErrUnset when s or next is unset.
func (s *Settings) CheckDesignations(next *Settings) (Report, error) {
	if err := checkPair("check designations", s, next); err != nil {
		return Report{}, err
	}
	var r Report
	r.checked = ruleBit(RuleDesignations)
	r.designations = designationsViolation(s, next)
	return r, nil
}

// MustCheckDesignations is like CheckDesignations but panics with its error
// on failure.
func (s *Settings) MustCheckDesignations(next *Settings) Report {
	return must(s.CheckDesignations(next))
}

// checkPair reports ErrUnset when s or next is unset.
func checkPair(op string, s, next *Settings) error {
	if !s.IsValid() {
		return unsetError(op, "settings")
	}
	if !next.IsValid() {
		return unsetError(op, "next settings")
	}
	return nil
}

// structureViolation returns the violation of RuleStructure for the valid
// systems prev and next, or the zero Violation.
func structureViolation(prev, next *system.System) Violation {
	subset, err := prev.NamesSubsetOf(next)
	if err != nil || subset {
		return Violation{}
	}
	v := Violation{Rule: RuleStructure}
	if a, ok := structureWitness(prev.Alterations(), next.Alterations()); ok {
		v.Name = spelling.C().WithAlteration(a)
	}
	return v
}

// octaveSet is a set W of writable written octaves (ASPN numbers) of a set of
// settings, with the rule that a style showing any octave shows all of them.
type octaveSet struct {
	kind    octaveSetKind
	catalog *naming.Catalog // for octaveSetCatalog
}

type octaveSetKind uint8

const (
	octaveSetEmpty   octaveSetKind = iota // no octave: a name-only notation
	octaveSetAll                          // every octave
	octaveSetAM                           // ASPN 0..8
	octaveSetCatalog                      // the named octaves of a catalog
)

// writableOctaves returns the set W of the valid set s.
func writableOctaves(s *Settings) octaveSet {
	switch s.notation.OctaveStyle() {
	case notation.OctaveStyleNone:
		return octaveSet{kind: octaveSetEmpty}
	case notation.OctaveStyleAM:
		return octaveSet{kind: octaveSetAM}
	case notation.OctaveStyleName:
		return octaveSet{kind: octaveSetCatalog, catalog: s.catalog}
	default:
		return octaveSet{kind: octaveSetAll}
	}
}

// contains reports whether the written octave w belongs to the set.
func (o octaveSet) contains(w int) bool {
	switch o.kind {
	case octaveSetAll:
		return true
	case octaveSetAM:
		return 0 <= w && w <= 8
	case octaveSetCatalog:
		return o.catalog.Name(w) != ""
	default:
		return false
	}
}

// missing returns the written octave nearest to 4 that does not belong to a
// set other than octaveSetAll, walking 4, 5, 3, 6, 2, … The walk is finite:
// the empty set and AM give 4 and 9, and a catalog names finitely many
// octaves.
func (o octaveSet) missing() int {
	for step := 0; ; step++ {
		w := 4 + (step+1)/2
		if step%2 == 0 {
			w = 4 - step/2
		}
		if !o.contains(w) {
			return w
		}
	}
}

// designationsViolation returns the violation of RuleDesignations for the
// valid sets prev and next, or the zero Violation.
func designationsViolation(prev, next *Settings) Violation {
	var expressible []spelling.Spelling
	for _, name := range prev.notation.NameRules().Names() {
		if !prev.expresses(name) {
			continue
		}
		if !next.expresses(name) {
			return Violation{Rule: RuleDesignations, Name: name}
		}
		expressible = append(expressible, name)
	}

	wp, wn := writableOctaves(prev), writableOctaves(next)
	if wp.kind == octaveSetEmpty || wn.kind == octaveSetAll {
		return Violation{}
	}
	prevSys, nextSys := prev.tuning.System(), next.tuning.System()
	type carries struct{ prev, next int }
	seen := make(map[carries]bool)
	for _, name := range expressible {
		kp, err := prevSys.Carry(name)
		if err != nil {
			continue
		}
		kn, err := nextSys.Carry(name)
		if err != nil {
			continue
		}
		key := carries{kp, kn}
		if seen[key] {
			continue
		}
		seen[key] = true
		if octave, ok := missingOctave(wp, wn, kp, kn); ok {
			if p, err := pitch.FromASPN(name, octave); err == nil {
				return Violation{Rule: RuleDesignations, Pitch: p}
			}
		}
	}
	return Violation{}
}

// missingOctave returns an Octave O of a pitch whose written octave O − kp
// belongs to wp and whose written octave O − kn does not belong to wn, where
// kp and kn are the carries of its name in the previous and the next system,
// and wn is not octaveSetAll. Octaves that do not fit in int do not exist and
// are skipped.
func missingOctave(wp, wn octaveSet, kp, kn int) (int, bool) {
	switch wp.kind {
	case octaveSetAll:
		w := wn.missing()
		return w + kn, true
	case octaveSetAM:
		for w := 0; w <= 8; w++ {
			if !wn.contains(w + kp - kn) {
				return w + kp, true
			}
		}
	case octaveSetCatalog:
		for a := range wp.catalog.All() {
			o, ok := checkedAdd(a, kp)
			if !ok {
				continue
			}
			w, ok := checkedSub(o, kn)
			if !ok || !wn.contains(w) {
				return o, true
			}
		}
	}
	return 0, false
}

// checkedAdd returns a + b and whether it fits in an int.
func checkedAdd(a, b int) (int, bool) {
	s := a + b
	if (b > 0 && s < a) || (b < 0 && s > a) {
		return 0, false
	}
	return s, true
}
