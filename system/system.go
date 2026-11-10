package system

import (
	"strconv"
	"strings"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/pitchclass"
)

// Definition is the definition of a system: the values from which New creates
// a System. It is a plain value; New copies it and keeps only the shared,
// immutable *Rule.
//
// The zero Definition has no boundary letter and no rule; New rejects it with
// an error that matches ErrInvalid. == compares the fields, Alterations by
// pointer; use System.Equal to compare systems.
type Definition struct {
	// Name is an optional label for debugging and catalogs. It is not checked
	// and does not take part in System.Equal.
	Name string

	// Classes is N, the number of classes, 1..MaxClasses.
	Classes int

	// Boundary is b, the boundary letter: it has offset 0, and the boundary of
	// the octave of positions lies just before it.
	Boundary letter.Letter

	// Offsets holds p(l), the offset of each natural letter l from the boundary
	// letter, in 0..N, indexed by l.Ordinal().
	Offsets [letter.Count]int

	// Chromatic is c, the number of classes by which the alteration value +1
	// shifts a class, 1..MaxChromatic.
	Chromatic int

	// Alterations is A, the rule of admissible alteration values.
	Alterations *Rule
}

// Regularity describes whether a system is regular. A system is regular when
// the steps after C, D, E, F, G, A, B are Large Large Small Large Large Large
// Small and c = Large − Small. It is regular by classes when the steps after
// C, D, F, G, A are congruent modulo N and the steps after E and B are smaller
// by c modulo N.
//
// Regularity is a comparable value; the zero value means "not regular in any
// sense".
type Regularity struct {
	// Regular reports whether the system is regular.
	Regular bool

	// Large and Small are the large and small steps of a regular system (Small
	// can be 0); both are 0 when Regular is false.
	Large, Small int

	// RegularByClasses reports whether the system is regular by classes.
	RegularByClasses bool
}

// System is a musical system: the rules by which an admissible name gets a
// class of positions and a written pitch gets a position, a written octave, an
// order and distances in steps. See the package documentation.
//
// A System is used through a pointer and is immutable after creation; it is
// safe for concurrent use. The nil pointer and a pointer to the zero System
// are the unset system: IsValid reports false, the simple readers return
// values that a valid system never has (see Definition), Equal reports false,
// String returns "System(unset)", and the other methods return an error that
// matches ErrInvalid. A System value is not comparable: == on pointers is
// identity; use Equal.
type System struct {
	valid     bool
	name      string
	n         int
	b         letter.Letter
	offsets   [letter.Count]int
	c         int
	rule      *Rule
	steps     [letter.Count]int
	shifts    []shiftProgression
	regular   Regularity
	tradition bool
}

// shiftProgression is a progression of the rule in units m = a·c: the integers
// lo, lo + stride, ..., hi; stride is 0 for a single value.
type shiftProgression struct {
	lo, hi, stride int64
}

// traditionalSteps are the steps after C, D, E, F, G, A, B of the traditional
// layout.
var traditionalSteps = [letter.Count]int{2, 2, 1, 2, 2, 2, 1}

// New returns the system with the definition def.
//
// It checks def in this order and reports the first failure:
//   - an invalid boundary letter or a nil or unset rule: an error that matches
//     ErrInvalid;
//   - N outside 1..MaxClasses: ErrDefinition and ErrClassCount;
//   - p(b) ≠ 0 or some p(l) outside 0..N: ErrDefinition and ErrOffsets (the
//     value N is allowed: such a letter stands on the boundary of the next
//     octave of positions);
//   - the offsets go down along the circle of letters b, b+1, ..., b+6:
//     ErrDefinition and ErrDescending;
//   - c outside 1..MaxChromatic: ErrDefinition and ErrChromatic;
//   - 0 is not in the rule: ErrDefinition and ErrNoNatural;
//   - some admissible a has a non-integer shift a·c: ErrDefinition and
//     ErrFractionalShift.
//
// The name is not checked.
func New(def Definition) (*System, error) {
	if !def.Boundary.IsValid() {
		return nil, errorf(ErrInvalid, "boundary letter %v", def.Boundary)
	}

	if !def.Alterations.IsValid() {
		return nil, errorf(ErrInvalid, "unset alteration rule")
	}

	n := def.Classes
	if n < 1 || n > MaxClasses {
		return nil, definitionError(ErrClassCount, "N = %d is outside 1..%d", n, MaxClasses)
	}

	b := def.Boundary
	if def.Offsets[b.Ordinal()] != 0 {
		return nil, definitionError(ErrOffsets, "boundary letter %v has offset %d", b, def.Offsets[b.Ordinal()])
	}

	for i, p := range def.Offsets {
		if p < 0 || p > n {
			return nil, definitionError(ErrOffsets, "offset %d of %v is outside 0..%d", p, letter.MustFromOrdinal(i), n)
		}
	}

	for i := range letter.Count - 1 {
		from, to := b.Shift(i), b.Shift(i+1)
		if def.Offsets[to.Ordinal()] < def.Offsets[from.Ordinal()] {
			return nil, definitionError(ErrDescending, "offset of %v is below the offset of %v", to, from)
		}
	}

	c := def.Chromatic
	if c < 1 || c > MaxChromatic {
		return nil, definitionError(ErrChromatic, "c = %d is outside 1..%d", c, MaxChromatic)
	}

	rule := def.Alterations
	if !rule.contains(alteration.Natural()) {
		return nil, definitionError(ErrNoNatural, "rule %v has no 0", rule)
	}

	for _, p := range rule.ps {
		if int64(c)%p.unit != 0 {
			return nil, definitionError(ErrFractionalShift, "progression %v has values a with non-integer a·%d", p, c)
		}
	}

	s := &System{
		valid:   true,
		name:    def.Name,
		n:       n,
		b:       b,
		offsets: def.Offsets,
		c:       c,
		rule:    rule,
		shifts:  make([]shiftProgression, len(rule.ps)),
	}

	for i, p := range rule.ps {
		f := int64(c) / p.unit

		sp := shiftProgression{lo: p.lo * f, hi: p.hi * f}
		if p.den != 0 {
			sp.stride = p.stride() * f
		}

		s.shifts[i] = sp
	}

	for _, l := range letter.All() {
		next := l.Shift(1)

		to := s.offsets[next.Ordinal()]
		if next == b {
			to = n
		}

		s.steps[l.Ordinal()] = to - s.offsets[l.Ordinal()]
	}

	s.regular = s.regularity()
	s.tradition = n == 12 && c == 1 && s.steps == traditionalSteps

	return s, nil
}

// MustNew is like New but panics on error.
func MustNew(def Definition) *System {
	return must(New(def))
}

// FromSteps returns the system defined by steps between letters: steps[i] is
// the step after the letter with Ordinal() = i, from it to the next letter up.
// N is the sum of the steps, and the offsets are the sums of the steps along
// the circle of letters from the boundary letter. It is equivalent to New with
// these N and offsets.
//
// It returns the errors of New on these values; in addition, a sum of all
// steps that does not fit in int matches ErrClassCount, and a partial sum that
// does not fit in int matches ErrOffsets.
func FromSteps(boundary letter.Letter, steps [letter.Count]int, chromatic int, alterations *Rule, name string) (*System, error) {
	if !boundary.IsValid() {
		return nil, errorf(ErrInvalid, "boundary letter %v", boundary)
	}

	if !alterations.IsValid() {
		return nil, errorf(ErrInvalid, "unset alteration rule")
	}

	n, ok := sumInts(steps[:])
	if !ok {
		return nil, definitionError(ErrClassCount, "the sum of steps %v does not fit in int", steps)
	}

	if n < 1 || n > MaxClasses {
		return nil, definitionError(ErrClassCount, "N = %d is outside 1..%d", n, MaxClasses)
	}

	var offsets [letter.Count]int

	p := 0

	for i := 1; i < letter.Count; i++ {
		var ok bool
		if p, ok = addInt(p, steps[boundary.Shift(i-1).Ordinal()]); !ok {
			return nil, definitionError(ErrOffsets, "a partial sum of steps %v does not fit in int", steps)
		}

		offsets[boundary.Shift(i).Ordinal()] = p
	}

	return New(Definition{
		Name:        name,
		Classes:     n,
		Boundary:    boundary,
		Offsets:     offsets,
		Chromatic:   chromatic,
		Alterations: alterations,
	})
}

// MustFromSteps is like FromSteps but panics on error.
func MustFromSteps(boundary letter.Letter, steps [letter.Count]int, chromatic int, alterations *Rule, name string) *System {
	return must(FromSteps(boundary, steps, chromatic, alterations, name))
}

// NewRegular returns the regular system with the steps large large small large
// large large small after C, D, E, F, G, A, B and c = large − small. It is
// FromSteps with these values: N = 5·large + 2·small. A c that does not fit in
// int matches ErrChromatic, and an N that does not fit matches ErrClassCount;
// large <= small gives ErrChromatic when the other values are valid.
func NewRegular(large, small int, boundary letter.Letter, alterations *Rule, name string) (*System, error) {
	c, ok := subInt(large, small)
	if !ok {
		c = 0 // out of range: reported as ErrChromatic in its turn
	}

	steps := [letter.Count]int{large, large, small, large, large, large, small}

	return FromSteps(boundary, steps, c, alterations, name)
}

// MustNewRegular is like NewRegular but panics on error.
func MustNewRegular(large, small int, boundary letter.Letter, alterations *Rule, name string) *System {
	return must(NewRegular(large, small, boundary, alterations, name))
}

// IsValid reports whether s is set.
func (s *System) IsValid() bool {
	return s != nil && s.valid
}

// Definition returns a copy of the definition of s, with the same *Rule. It
// returns the zero Definition for the unset system.
func (s *System) Definition() Definition {
	if !s.IsValid() {
		return Definition{}
	}

	return Definition{
		Name:        s.name,
		Classes:     s.n,
		Boundary:    s.b,
		Offsets:     s.offsets,
		Chromatic:   s.c,
		Alterations: s.rule,
	}
}

// Name returns the name of s. It is "" for the unset system, and can also be
// "" for a valid one: use IsValid to tell them apart.
func (s *System) Name() string {
	if !s.IsValid() {
		return ""
	}

	return s.name
}

// Classes returns N; 0 for the unset system.
func (s *System) Classes() int {
	if !s.IsValid() {
		return 0
	}

	return s.n
}

// Boundary returns the boundary letter b; the zero (invalid) letter for the
// unset system.
func (s *System) Boundary() letter.Letter {
	if !s.IsValid() {
		return 0
	}

	return s.b
}

// Chromatic returns c; 0 for the unset system.
func (s *System) Chromatic() int {
	if !s.IsValid() {
		return 0
	}

	return s.c
}

// Alterations returns the rule A, shared and immutable; nil for the unset
// system.
func (s *System) Alterations() *Rule {
	if !s.IsValid() {
		return nil
	}

	return s.rule
}

// Steps returns the steps after C, D, E, F, G, A, B, indexed by Ordinal(); their
// sum is N. It returns the zero array for the unset system.
func (s *System) Steps() [letter.Count]int {
	if !s.IsValid() {
		return [letter.Count]int{}
	}

	return s.steps
}

// Offset returns p(l), the offset of the natural letter l.
//
// It returns an error that matches ErrInvalid for the unset system or an
// invalid letter.
func (s *System) Offset(l letter.Letter) (int, error) {
	if err := s.checkLetter(l); err != nil {
		return 0, err
	}

	return s.offsets[l.Ordinal()], nil
}

// MustOffset is like Offset but panics on error.
func (s *System) MustOffset(l letter.Letter) int {
	return must(s.Offset(l))
}

// StepAfter returns the step after the letter l, from it to the next letter
// up.
//
// It returns an error that matches ErrInvalid for the unset system or an
// invalid letter.
func (s *System) StepAfter(l letter.Letter) (int, error) {
	if err := s.checkLetter(l); err != nil {
		return 0, err
	}

	return s.steps[l.Ordinal()], nil
}

// MustStepAfter is like StepAfter but panics on error.
func (s *System) MustStepAfter(l letter.Letter) int {
	return must(s.StepAfter(l))
}

// NaturalClass returns the class of the natural letter l: p(l) mod N, with the
// modulus N.
//
// It returns an error that matches ErrInvalid for the unset system or an
// invalid letter.
func (s *System) NaturalClass(l letter.Letter) (pitchclass.Class, error) {
	if err := s.checkLetter(l); err != nil {
		return pitchclass.Class{}, err
	}

	return s.class(int64(s.offsets[l.Ordinal()])), nil
}

// MustNaturalClass is like NaturalClass but panics on error.
func (s *System) MustNaturalClass(l letter.Letter) pitchclass.Class {
	return must(s.NaturalClass(l))
}

// Regularity returns whether s is regular, with its large and small steps, and
// whether it is regular by classes. Irregularity is not an error.
//
// It returns an error that matches ErrInvalid for the unset system.
func (s *System) Regularity() (Regularity, error) {
	if !s.IsValid() {
		return Regularity{}, errorf(ErrInvalid, "regularity of an unset system")
	}

	return s.regular, nil
}

// MustRegularity is like Regularity but panics on error.
func (s *System) MustRegularity() Regularity {
	return must(s.Regularity())
}

// HasTraditionalLayout reports whether s has the traditional layout: N = 12,
// c = 1 and the steps 2 2 1 2 2 2 1 after C, D, E, F, G, A, B. The boundary
// letter and the rule can be any.
//
// It returns an error that matches ErrInvalid for the unset system.
func (s *System) HasTraditionalLayout() (bool, error) {
	if !s.IsValid() {
		return false, errorf(ErrInvalid, "layout of an unset system")
	}

	return s.tradition, nil
}

// MustHasTraditionalLayout is like HasTraditionalLayout but panics on error.
func (s *System) MustHasTraditionalLayout() bool {
	return must(s.HasTraditionalLayout())
}

// NamesSubsetOf reports whether every name admissible in s is admissible in
// other: whether the rule of s is a subset of the rule of other. N, offsets
// and c do not matter.
//
// It returns an error that matches ErrInvalid when s or other is unset.
func (s *System) NamesSubsetOf(other *System) (bool, error) {
	if !s.IsValid() || !other.IsValid() {
		return false, errorf(ErrInvalid, "inclusion of names of unset systems")
	}

	return s.rule.SubsetOf(other.rule)
}

// MustNamesSubsetOf is like NamesSubsetOf but panics on error.
func (s *System) MustNamesSubsetOf(other *System) bool {
	return must(s.NamesSubsetOf(other))
}

// Equal reports whether s and other are set and have the same N, boundary
// letter, offsets of all letters, c and rule (by Rule.Equal). The name is not
// compared.
func (s *System) Equal(other *System) bool {
	if !s.IsValid() || !other.IsValid() {
		return false
	}

	if s == other {
		return true
	}

	return s.n == other.n && s.b == other.b && s.offsets == other.offsets && s.c == other.c &&
		(s.rule == other.rule || s.rule.Equal(other.rule))
}

// String returns a debugging form, for example
//
//	System("traditional"; N=12; b=C; C=0 D=2 E=4 F=5 G=7 A=9 B=11; c=1; A=0 + Z/1 in [-inf, +inf])
//
// The name is quoted with strconv.Quote. The unset system is "System(unset)".
// This is not a notation and cannot be parsed back.
func (s *System) String() string {
	if !s.IsValid() {
		return "System(unset)"
	}

	var b strings.Builder

	b.WriteString("System(")
	b.WriteString(strconv.Quote(s.name))
	b.WriteString("; N=")
	b.WriteString(strconv.Itoa(s.n))
	b.WriteString("; b=")
	b.WriteString(s.b.String())
	b.WriteString(";")

	for _, l := range letter.All() {
		b.WriteString(" ")
		b.WriteString(l.String())
		b.WriteString("=")
		b.WriteString(strconv.Itoa(s.offsets[l.Ordinal()]))
	}

	b.WriteString("; c=")
	b.WriteString(strconv.Itoa(s.c))
	b.WriteString("; A=")
	b.WriteString(s.rule.String())
	b.WriteString(")")

	return b.String()
}

// checkLetter checks that s is set and l is valid.
func (s *System) checkLetter(l letter.Letter) error {
	if !s.IsValid() {
		return errorf(ErrInvalid, "unset system")
	}

	if !l.IsValid() {
		return errorf(ErrInvalid, "letter %v", l)
	}

	return nil
}

// regularity computes the regularity of a new system.
func (s *System) regularity() Regularity {
	st := s.steps
	large, small := st[0], st[2] // after C and after E

	var r Regularity

	if st[1] == large && st[3] == large && st[4] == large && st[5] == large && st[6] == small &&
		large-small == s.c {
		r.Regular, r.Large, r.Small = true, large, small
	}

	n := int64(s.n)
	congruent := func(x, y int64) bool { return floorMod(x-y, n) == 0 }
	l64, small64 := int64(large), int64(large)-int64(s.c)
	r.RegularByClasses = congruent(int64(st[1]), l64) && congruent(int64(st[3]), l64) &&
		congruent(int64(st[4]), l64) && congruent(int64(st[5]), l64) &&
		congruent(int64(st[2]), small64) && congruent(int64(st[6]), small64)

	return r
}
