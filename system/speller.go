package system

import (
	"strconv"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/spelling"
)

// ScaleMode is the mode of the scale of a ChromaticScale rule.
type ScaleMode uint8

// Scale modes. The numbers of the constants are stable.
const (
	// ScaleModeUnset is the zero, unset mode.
	ScaleModeUnset ScaleMode = iota

	// Major is the major mode.
	Major

	// Minor is the minor mode, built from the natural minor scale.
	Minor
)

// String returns "unset", "major", "minor", or "ScaleMode(<n>)" for other
// values.
func (m ScaleMode) String() string {
	switch m {
	case ScaleModeUnset:
		return "unset"
	case Major:
		return "major"
	case Minor:
		return "minor"
	default:
		return "ScaleMode(" + strconv.Itoa(int(m)) + ")"
	}
}

// Motion is the direction of a ChromaticScale rule.
type Motion uint8

// Motions. The numbers of the constants are stable.
const (
	// MotionUnset is the zero, unset direction.
	MotionUnset Motion = iota

	// Ascending is the upward direction.
	Ascending

	// Descending is the downward direction.
	Descending
)

// String returns "unset", "ascending", "descending", or "Motion(<n>)" for
// other values.
func (m Motion) String() string {
	switch m {
	case MotionUnset:
		return "unset"
	case Ascending:
		return "ascending"
	case Descending:
		return "descending"
	default:
		return "Motion(" + strconv.Itoa(int(m)) + ")"
	}
}

// ruleKind tells the kinds of spelling rules apart.
type ruleKind uint8

const (
	kindUnset ruleKind = iota
	kindSharps
	kindFlats
	kindChromatic
	kindCustom
)

// SpellingRule is an explicit rule that chooses one admissible name of a
// class. The package never chooses a name without a rule. Rules are created by
// Sharps, Flats, ChromaticScale and CustomSpelling.
//
// SpellingRule is an immutable value; it is not comparable because a custom
// rule holds a function. The zero value is the unset rule: IsValid reports
// false, String returns "SpellingRule(unset)", and applying it gives an error
// that matches ErrInvalid. A rule is safe for concurrent use when its custom
// function, if any, is.
type SpellingRule struct {
	kind   ruleKind
	tonic  spelling.Spelling
	mode   ScaleMode
	motion Motion
	choose func(s *System, c pitchclass.Class) spelling.Spelling
}

// Sharps returns the rule "sharps": among the admissible names of the class,
// the one with the smallest |a|; on a tie, the one with a > 0; on a further
// tie, the first letter met going up the circle of letters from the boundary
// letter, starting with it. A class without admissible names gives an error
// that matches ErrNoSpelling.
func Sharps() SpellingRule {
	return SpellingRule{kind: kindSharps}
}

// Flats returns the rule "flats": like Sharps, but on a tie of |a| the name
// with a < 0 is chosen.
func Flats() SpellingRule {
	return SpellingRule{kind: kindFlats}
}

// ChromaticScale returns the rule "chromatic scale": the name of a class in the
// chromatic scale on the tonic tonic in the mode mode and the direction motion,
// by the rules of the traditional chromatic scale. The scale keeps the
// spelling of the degrees of the major or natural minor scale of the tonic;
// going up, it raises the degrees a whole step below the next one, except that
// the major lowers VII instead of raising VI and the minor lowers II instead of
// raising I; going down, the major lowers the degrees a whole step above the
// next one down, except that it raises IV instead of lowering V; the
// descending minor scale is written as the descending major scale on the same
// tonic. For example, C major ascending gives C C♯ D D♯ E F F♯ G G♯ A B♭ B.
//
// The rule applies only to a system with the traditional layout (see
// System.HasTraditionalLayout); with another system it gives an error that
// matches ErrNotApplicable. A tonic that is not admissible gives
// ErrInadmissible, and a needed name of the scale that is not admissible, or
// whose alteration value cannot be represented, gives ErrNoSpelling.
//
// It returns an error that matches ErrInvalid when tonic is unset, mode is not
// Major or Minor, or motion is not Ascending or Descending.
func ChromaticScale(tonic spelling.Spelling, mode ScaleMode, motion Motion) (SpellingRule, error) {
	if !tonic.IsValid() {
		return SpellingRule{}, errorf(ErrInvalid, "unset tonic")
	}

	if mode != Major && mode != Minor {
		return SpellingRule{}, errorf(ErrInvalid, "mode %v", mode)
	}

	if motion != Ascending && motion != Descending {
		return SpellingRule{}, errorf(ErrInvalid, "motion %v", motion)
	}

	return SpellingRule{kind: kindChromatic, tonic: tonic, mode: mode, motion: motion}, nil
}

// MustChromaticScale is like ChromaticScale but panics on error.
func MustChromaticScale(tonic spelling.Spelling, mode ScaleMode, motion Motion) SpellingRule {
	return must(ChromaticScale(tonic, mode, motion))
}

// CustomSpelling returns a rule that calls choose(s, c) to choose the name of
// the class c in the system s. choose is expected to return the same name for
// the same class; the package does not check this. The package checks that the
// returned name is set, admissible and of the class c; otherwise applying the
// rule gives an error that matches ErrBadSpelling.
//
// choose is called synchronously, possibly many times, and possibly from many
// goroutines when the rule is used concurrently; it must then be safe for
// concurrent use. A panic in choose is not recovered. Walk calls choose once
// per class before it returns.
//
// It returns an error that matches ErrInvalid when choose is nil.
func CustomSpelling(choose func(s *System, c pitchclass.Class) spelling.Spelling) (SpellingRule, error) {
	if choose == nil {
		return SpellingRule{}, errorf(ErrInvalid, "nil spelling function")
	}

	return SpellingRule{kind: kindCustom, choose: choose}, nil
}

// MustCustomSpelling is like CustomSpelling but panics on error.
func MustCustomSpelling(choose func(s *System, c pitchclass.Class) spelling.Spelling) SpellingRule {
	return must(CustomSpelling(choose))
}

// IsValid reports whether r is set.
func (r SpellingRule) IsValid() bool {
	return r.kind != kindUnset
}

// String returns "sharps", "flats", "chromatic(<tonic>, <mode>, <motion>)"
// (for example "chromatic(C, major, ascending)"), "custom", or
// "SpellingRule(unset)".
func (r SpellingRule) String() string {
	switch r.kind {
	case kindSharps:
		return "sharps"
	case kindFlats:
		return "flats"
	case kindChromatic:
		return "chromatic(" + r.tonic.String() + ", " + r.mode.String() + ", " + r.motion.String() + ")"
	case kindCustom:
		return "custom"
	default:
		return "SpellingRule(unset)"
	}
}

// Spell returns the name of the class c chosen by rule.
//
// Errors, in this order: ErrInvalid when s, c or rule is unset;
// ErrModulusMismatch when the modulus of c is not N; ErrNotApplicable when
// rule does not apply to s; ErrInadmissible when the tonic of rule is not
// admissible; ErrNoSpelling when a predefined rule gives no admissible name;
// ErrBadSpelling when a custom rule returns a bad name.
func (s *System) Spell(c pitchclass.Class, rule SpellingRule) (spelling.Spelling, error) {
	if err := s.checkClass(c, rule); err != nil {
		return spelling.Spelling{}, err
	}

	sp, err := s.prepare(rule)
	if err != nil {
		return spelling.Spelling{}, err
	}

	return sp.spell(int64(c.Number()))
}

// MustSpell is like Spell but panics on error.
func (s *System) MustSpell(c pitchclass.Class, rule SpellingRule) spelling.Spelling {
	return must(s.Spell(c, rule))
}

// PitchAt returns the pitch with the name Spell(c, rule) and the Octave o.
// Errors are those of Spell; in addition, an unset o gives an error that
// matches ErrInvalid.
func (s *System) PitchAt(c pitchclass.Class, o octave.Octave, rule SpellingRule) (pitch.Pitch, error) {
	if err := s.checkClassInputs(c, rule); err != nil {
		return pitch.Pitch{}, err
	}

	if !o.IsValid() {
		return pitch.Pitch{}, errorf(ErrInvalid, "unset octave")
	}

	if err := s.checkModulus(c); err != nil {
		return pitch.Pitch{}, err
	}

	name, err := s.Spell(c, rule)
	if err != nil {
		return pitch.Pitch{}, err
	}

	return pitch.MustNew(name, o), nil
}

// MustPitchAt is like PitchAt but panics on error.
func (s *System) MustPitchAt(c pitchclass.Class, o octave.Octave, rule SpellingRule) pitch.Pitch {
	return must(s.PitchAt(c, o, rule))
}

// checkClass checks that s, c and rule are set and that c has the modulus N.
func (s *System) checkClass(c pitchclass.Class, rule SpellingRule) error {
	if err := s.checkClassInputs(c, rule); err != nil {
		return err
	}

	return s.checkModulus(c)
}

// checkClassInputs checks that s, c and rule are set.
func (s *System) checkClassInputs(c pitchclass.Class, rule SpellingRule) error {
	if !s.IsValid() {
		return errorf(ErrInvalid, "unset system")
	}

	if !c.IsValid() {
		return errorf(ErrInvalid, "unset class")
	}

	if !rule.IsValid() {
		return errorf(ErrInvalid, "unset spelling rule")
	}

	return nil
}

// speller applies a spelling rule to the classes of one system.
type speller struct {
	s     *System
	rule  SpellingRule
	scale [12]scaleName // names of the chromatic scale by class
}

// scaleName is a name of a chromatic scale; a is an integer alteration value
// that may lie outside the range of package alteration.
type scaleName struct {
	l letter.Letter
	a int64
}

// prepare checks rule against s (ErrInvalid, ErrNotApplicable,
// ErrInadmissible) and returns its speller.
func (s *System) prepare(rule SpellingRule) (speller, error) {
	if !rule.IsValid() {
		return speller{}, errorf(ErrInvalid, "unset spelling rule")
	}

	sp := speller{s: s, rule: rule}

	if rule.kind != kindChromatic {
		return sp, nil
	}

	if !s.tradition {
		return speller{}, errorf(ErrNotApplicable, "%v in %v without the traditional layout", rule, s.name)
	}

	if _, ok := s.shift(rule.tonic); !ok {
		return speller{}, errorf(ErrInadmissible, "tonic %v in %v", rule.tonic, s.name)
	}

	sp.scale = s.chromaticScale(rule)

	return sp, nil
}

// spell returns the name of the class kappa.
func (sp speller) spell(kappa int64) (spelling.Spelling, error) {
	s := sp.s

	switch sp.rule.kind {
	case kindSharps, kindFlats:
		name, ok := s.nearest(kappa, sp.rule.kind == kindSharps)
		if !ok {
			return spelling.Spelling{}, errorf(ErrNoSpelling, "class %d of %v by %v", kappa, s.name, sp.rule)
		}

		return name, nil
	case kindChromatic:
		sn := sp.scale[kappa]
		if sn.a < -maxMagnitude || sn.a > maxMagnitude {
			return spelling.Spelling{}, errorf(ErrNoSpelling, "class %d of %v by %v", kappa, s.name, sp.rule)
		}

		name := spelling.MustNew(sn.l, alteration.MustFromInt(int(sn.a)))
		if _, ok := s.shift(name); !ok {
			return spelling.Spelling{}, errorf(ErrNoSpelling, "%v is not admissible in %v", name, s.name)
		}

		return name, nil
	default: // kindCustom
		c := pitchclass.MustNew(int(kappa), s.n)

		name := sp.rule.choose(s, c)
		if !name.IsValid() {
			return spelling.Spelling{}, errorf(ErrBadSpelling, "unset name for class %v", c)
		}

		sh, ok := s.shift(name)
		if !ok {
			return spelling.Spelling{}, errorf(ErrBadSpelling, "%v for class %v is not admissible", name, c)
		}

		if floorMod(sh, int64(s.n)) != kappa {
			return spelling.Spelling{}, errorf(ErrBadSpelling, "%v is not of class %v", name, c)
		}

		return name, nil
	}
}

// nearest returns the admissible name of the class kappa chosen by the rule
// "sharps" (sharps is true) or "flats"; ok is false when there is none.
func (s *System) nearest(kappa int64, sharps bool) (spelling.Spelling, bool) {
	var (
		best     spelling.Spelling
		bestM    int64
		found    bool
		preferUp = sharps
	)

	better := func(m int64) bool {
		if !found {
			return true
		}

		am, ab := max(m, -m), max(bestM, -bestM)
		if am != ab {
			return am < ab
		}

		if m == bestM {
			return false
		}

		return (m > 0) == preferUp
	}

	for i := range letter.Count {
		l := s.b.Shift(i)
		r := kappa - int64(s.offsets[l.Ordinal()])

		for _, sp := range s.shifts {
			start, period, ok := s.residue(sp, r)
			if !ok {
				continue
			}

			// The smallest point >= max(lo, 0) and the largest point <= min(hi, 0).
			up := max(sp.lo, 0)
			up += floorMod(start-up, period)

			down := min(sp.hi, 0)
			down -= floorMod(down-start, period)

			for _, m := range [2]int64{up, down} {
				if m < sp.lo || m > sp.hi || !better(m) {
					continue
				}

				best = spelling.MustNew(l, alteration.MustFromFraction(int(m), s.c))
				bestM, found = m, true
			}
		}
	}

	return best, found
}

// chromaticScale returns the names of the chromatic scale of rule by class in
// s, a system with the traditional layout (c = 1, N = 12) and an admissible
// tonic, which therefore has an integer alteration value.
func (s *System) chromaticScale(rule SpellingRule) [12]scaleName {
	major := [letter.Count]int64{0, 2, 4, 5, 7, 9, 11}
	minor := [letter.Count]int64{0, 2, 3, 5, 7, 8, 10}

	mode := rule.mode
	if rule.motion == Descending {
		mode = Major // the descending minor scale is written as the major one
	}

	template := major
	if mode == Minor {
		template = minor
	}

	tl := rule.tonic.Letter()
	a0, _ := rule.tonic.Alteration().Fraction()

	var degrees [letter.Count]scaleName

	dist := int64(0) // steps of the system from the tonic letter up to the degree letter

	for i := range letter.Count {
		l := tl.Shift(i)
		degrees[i] = scaleName{l: l, a: int64(a0) + template[i] - dist}
		dist += int64(s.steps[l.Ordinal()])
	}

	names := make([]scaleName, 0, 12)
	names = append(names, degrees[:]...)

	for i := range letter.Count {
		next := int64(12)
		if i+1 < letter.Count {
			next = template[i+1]
		}

		if next-template[i] != 2 {
			continue
		}

		j := (i + 1) % letter.Count

		switch {
		case rule.motion == Ascending && mode == Major && i == 5:
			names = append(names, scaleName{l: degrees[6].l, a: degrees[6].a - 1})
		case rule.motion == Ascending && mode == Minor && i == 0:
			names = append(names, scaleName{l: degrees[1].l, a: degrees[1].a - 1})
		case rule.motion == Ascending:
			names = append(names, scaleName{l: degrees[i].l, a: degrees[i].a + 1})
		case j == 4:
			names = append(names, scaleName{l: degrees[3].l, a: degrees[3].a + 1})
		default:
			names = append(names, scaleName{l: degrees[j].l, a: degrees[j].a - 1})
		}
	}

	var byClass [12]scaleName

	for _, sn := range names {
		kappa := floorMod(int64(s.offsets[sn.l.Ordinal()])+sn.a, 12)
		byClass[kappa] = sn
	}

	return byClass
}
