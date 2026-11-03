package spelling

import (
	"cmp"
	"fmt"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
)

// Spelling is a note name: a letter of the seven-letter set together with an
// alteration value, without an octave. For example, C sharp is the letter C
// with the value +1.
//
// A valid name has a valid letter and any alteration value. The zero value is
// the unset name: it means that no name is given, it is the only value that is
// not a valid name, and it is not C. See the Unset names section of the
// package documentation.
//
// Names are created by New, MustNew, Parse, MustParse, the ready-made names
// such as C and CSharp, Predefined, and the results of operations. Spelling
// has only unexported fields, so a literal of another package can create only
// the unset name.
//
// Spelling is a comparable value type of fixed size: it can be copied freely,
// a copy shares no data with the original, and it can be used as a map key or
// in a switch, for example case spelling.CSharp():. Spelling.Equal is the
// recommended comparison. For valid names, == and Spelling.Equal agree; the
// unset name is equal to itself under == but not under Spelling.Equal.
// Methods never modify the receiver or the arguments, except
// Spelling.UnmarshalText and Spelling.UnmarshalJSON, which write into the
// receiver.
type Spelling struct {
	l letter.Letter
	a alteration.Alteration
}

// New returns the name with the letter l and the alteration value a.
// Any alteration value is accepted, and the parts of the result are exactly l
// and a.
//
// If l is not a valid letter, New returns the unset name and an error that
// matches [letter.ErrInvalid]. The letter is never replaced with C.
func New(l letter.Letter, a alteration.Alteration) (Spelling, error) {
	if !l.IsValid() {
		return Spelling{}, fmt.Errorf("spelling: new name with %v: %w", l, letter.ErrInvalid)
	}

	return Spelling{l: l, a: a}, nil
}

// MustNew is like New but panics if l is not a valid letter. The panic value
// is the error returned by New, which matches [letter.ErrInvalid].
//
// It is intended for literals and values known to be valid, for example
// spelling.MustNew(letter.C, alteration.MustFromFraction(1, 2)).
func MustNew(l letter.Letter, a alteration.Alteration) Spelling {
	return must(New(l, a))
}

// Letter returns the letter of s.
//
// For the unset name, Letter returns the zero letter, which is not a valid
// letter. Check Spelling.IsValid before using the parts of a name.
func (s Spelling) Letter() letter.Letter {
	return s.l
}

// Alteration returns the alteration value of s.
//
// For the unset name, Alteration returns the natural, the zero value of
// alteration.Alteration. The result then makes sense only together with
// Spelling.IsValid: s.Alteration().IsNatural() is true for the unset name,
// although the unset name is not natural. Use Spelling.IsNatural to check
// whether a name is natural.
func (s Spelling) Alteration() alteration.Alteration {
	return s.a
}

// Parts returns the letter and the alteration value of s. It equals
// s.Letter(), s.Alteration().
//
// For the unset name, Parts returns the zero letter, which is not a valid
// letter, and the natural. The parts then make sense only together with
// Spelling.IsValid.
func (s Spelling) Parts() (letter.Letter, alteration.Alteration) {
	return s.l, s.a
}

// IsValid reports whether s is set, that is, whether it is a valid name.
// It returns false only for the unset name, the zero value of Spelling.
func (s Spelling) IsValid() bool {
	return s.l.IsValid()
}

// IsNatural reports whether s is a natural name: a valid name whose
// alteration value is the natural (0). It returns false for the unset name.
func (s Spelling) IsNatural() bool {
	return s.IsValid() && s.a.IsNatural()
}

// IsInteger reports whether s is a valid name whose alteration value is a
// whole number of chromatic units, such as C, C sharp, or E with +3. It returns
// false for fractional values, such as A with +1/2, and for the unset name.
func (s Spelling) IsInteger() bool {
	return s.IsValid() && s.a.IsInteger()
}

// Direction returns the direction of the alteration value of s:
// DirectionRaised for a value above 0, DirectionLowered for a value below 0,
// and DirectionNatural for the natural. For the unset name, it returns
// DirectionUnset, which no valid name has.
func (s Spelling) Direction() Direction {
	if !s.IsValid() {
		return DirectionUnset
	}

	switch s.a.Sign() {
	case 1:
		return DirectionRaised
	case -1:
		return DirectionLowered
	default:
		return DirectionNatural
	}
}

// SameLetter reports whether s and other are both valid names with the same
// letter, whatever their alteration values: C sharp and C flat have the same
// letter, C sharp and D flat do not. It returns false if either name is unset.
func (s Spelling) SameLetter(other Spelling) bool {
	return s.IsValid() && other.IsValid() && s.l == other.l
}

// SameAlteration reports whether s and other are both valid names with the
// same alteration value, whatever their letters: C sharp and D sharp have the
// same value. It returns false if either name is unset.
func (s Spelling) SameAlteration(other Spelling) bool {
	return s.IsValid() && other.IsValid() && s.a == other.a
}

// Natural returns the name with the letter of s and the natural: C sharp gives
// C. For the unset name, it returns the unset name.
func (s Spelling) Natural() Spelling {
	return s.WithAlteration(alteration.Natural())
}

// WithAlteration returns the name with the letter of s and the alteration
// value a: C sharp with a flat gives C flat. For the unset name, it returns
// the unset name, whatever a is.
func (s Spelling) WithAlteration(a alteration.Alteration) Spelling {
	if !s.IsValid() {
		return Spelling{}
	}

	return Spelling{l: s.l, a: a}
}

// Raise returns s raised by one chromatic unit: the same letter with the
// alteration value a + 1. The letter never changes: C sharp raised is C double
// sharp, not D. It equals s.RaiseBy(1) and s.RaiseByAlteration(alteration.Sharp()).
//
// If s is unset, Raise returns the unset name and an error that matches
// [letter.ErrInvalid]. If the result is outside the guaranteed range of
// package alteration, it returns the unset name and an error that matches
// [alteration.ErrOutOfRange].
func (s Spelling) Raise() (Spelling, error) {
	return s.apply("raise", alteration.Alteration.Raise)
}

// MustRaise is like Spelling.Raise but panics in the same cases. The panic
// value is the error returned by Raise.
//
// It is intended for literals and values known to be valid and in range.
func (s Spelling) MustRaise() Spelling {
	return must(s.Raise())
}

// Lower returns s lowered by one chromatic unit: the same letter with the
// alteration value a - 1. It equals s.LowerBy(1).
//
// If s is unset, Lower returns the unset name and an error that matches
// [letter.ErrInvalid]. If the result is outside the guaranteed range of
// package alteration, it returns the unset name and an error that matches
// [alteration.ErrOutOfRange].
func (s Spelling) Lower() (Spelling, error) {
	return s.apply("lower", alteration.Alteration.Lower)
}

// MustLower is like Spelling.Lower but panics in the same cases. The panic
// value is the error returned by Lower.
//
// It is intended for literals and values known to be valid and in range.
func (s Spelling) MustLower() Spelling {
	return must(s.Lower())
}

// RaiseBy returns s raised by n chromatic units: the same letter with the
// alteration value a + n. A negative n lowers s. Any int is accepted,
// including math.MinInt and math.MaxInt; only the exact value of the result
// is checked.
//
// If s is unset, RaiseBy returns the unset name and an error that matches
// [letter.ErrInvalid], whatever n is. If the result is outside the guaranteed
// range of package alteration, it returns the unset name and an error that
// matches [alteration.ErrOutOfRange].
func (s Spelling) RaiseBy(n int) (Spelling, error) {
	return s.apply("raise by", func(a alteration.Alteration) (alteration.Alteration, error) {
		return a.RaiseBy(n)
	})
}

// MustRaiseBy is like Spelling.RaiseBy but panics in the same cases.
// The panic value is the error returned by RaiseBy.
//
// It is intended for literals and values known to be valid and in range.
func (s Spelling) MustRaiseBy(n int) Spelling {
	return must(s.RaiseBy(n))
}

// LowerBy returns s lowered by n chromatic units: the same letter with the
// alteration value a - n. A negative n raises s. Any int is accepted,
// including math.MinInt and math.MaxInt; only the exact value of the result
// is checked. For any n other than math.MinInt, s.LowerBy(n) equals
// s.RaiseBy(-n).
//
// If s is unset, LowerBy returns the unset name and an error that matches
// [letter.ErrInvalid], whatever n is. If the result is outside the guaranteed
// range of package alteration, it returns the unset name and an error that
// matches [alteration.ErrOutOfRange].
func (s Spelling) LowerBy(n int) (Spelling, error) {
	return s.apply("lower by", func(a alteration.Alteration) (alteration.Alteration, error) {
		return a.LowerBy(n)
	})
}

// MustLowerBy is like Spelling.LowerBy but panics in the same cases.
// The panic value is the error returned by LowerBy.
//
// It is intended for literals and values known to be valid and in range.
func (s Spelling) MustLowerBy(n int) Spelling {
	return must(s.LowerBy(n))
}

// RaiseByAlteration returns s raised by the alteration value d: the same
// letter with the value a + d. The value d may be whole or fractional,
// positive or negative: C sharp raised by +1/2 is C with +3/2. For a whole d
// equal to n, it equals s.RaiseBy(n).
//
// If s is unset, RaiseByAlteration returns the unset name and an error that
// matches [letter.ErrInvalid]. If the result is outside the guaranteed range
// of package alteration, it returns the unset name and an error that matches
// [alteration.ErrOutOfRange].
func (s Spelling) RaiseByAlteration(d alteration.Alteration) (Spelling, error) {
	return s.apply("raise by alteration", func(a alteration.Alteration) (alteration.Alteration, error) {
		return a.Add(d)
	})
}

// MustRaiseByAlteration is like Spelling.RaiseByAlteration but panics in the
// same cases. The panic value is the error returned by RaiseByAlteration.
//
// It is intended for literals and values known to be valid and in range.
func (s Spelling) MustRaiseByAlteration(d alteration.Alteration) Spelling {
	return must(s.RaiseByAlteration(d))
}

// LowerByAlteration returns s lowered by the alteration value d: the same
// letter with the value a - d. The value d may be whole or fractional,
// positive or negative: C lowered by +1/2 is C with -1/2.
//
// If s is unset, LowerByAlteration returns the unset name and an error that
// matches [letter.ErrInvalid]. If the result is outside the guaranteed range
// of package alteration, it returns the unset name and an error that matches
// [alteration.ErrOutOfRange].
func (s Spelling) LowerByAlteration(d alteration.Alteration) (Spelling, error) {
	return s.apply("lower by alteration", func(a alteration.Alteration) (alteration.Alteration, error) {
		return a.Sub(d)
	})
}

// MustLowerByAlteration is like Spelling.LowerByAlteration but panics in the
// same cases. The panic value is the error returned by LowerByAlteration.
//
// It is intended for literals and values known to be valid and in range.
func (s Spelling) MustLowerByAlteration(d alteration.Alteration) Spelling {
	return must(s.LowerByAlteration(d))
}

// Equal reports whether s and other are both valid and the same name: the
// same letter and the same alteration value, however they were obtained.
// Names are compared as written, not by sound: C sharp is not equal to D flat,
// and B sharp is not equal to C.
//
// For valid names, s.Equal(other) equals s == other. If s or other is unset,
// Equal returns false, even when both are unset.
func (s Spelling) Equal(other Spelling) bool {
	return s.IsValid() && s == other
}

// Compare compares s and other in the agreed order for sorting. It returns -1
// if s sorts before other, 0 if they are the same, and +1 if s sorts after
// other.
//
// Valid names are ordered by letter in the canonical order C D E F G A B,
// then by alteration value (see [alteration.Alteration.Compare]). This order
// is not an order by pitch: B sharp sorts after C, and C flat sorts before B.
// The unset name sorts before every valid name, and two unset names compare
// as 0.
//
// The order is total, and Compare returns 0 exactly when s == other.
// Compare can be used with slices.SortFunc as spelling.Spelling.Compare.
func (s Spelling) Compare(other Spelling) int {
	switch sv, ov := s.IsValid(), other.IsValid(); {
	case !sv && !ov:
		return 0
	case !sv:
		return -1
	case !ov:
		return 1
	}

	if c := cmp.Compare(s.l.Ordinal(), other.l.Ordinal()); c != 0 {
		return c
	}

	return s.a.Compare(other.a)
}

// LetterDistanceUp returns the number of letter steps upward around the circle
// of letters from the letter of s to the letter of to, in the range 0 to 6.
// Alteration values do not affect it: from C sharp up to E flat it is 2.
// It equals s.Letter().DistanceUp(to.Letter()).
//
// The result is a count of letter steps, not an interval. If s or to is unset,
// LetterDistanceUp returns -1.
func (s Spelling) LetterDistanceUp(to Spelling) int {
	return s.l.DistanceUp(to.l)
}

// LetterDistanceDown returns the number of letter steps downward around the
// circle of letters from the letter of s to the letter of to, in the range 0
// to 6. Alteration values do not affect it: from C sharp down to E flat it is
// 5. It equals s.Letter().DistanceDown(to.Letter()) and to.LetterDistanceUp(s).
//
// The result is a count of letter steps, not an interval. If s or to is unset,
// LetterDistanceDown returns -1.
func (s Spelling) LetterDistanceDown(to Spelling) int {
	return s.l.DistanceDown(to.l)
}

// AlterationDifference returns the alteration value of s minus the alteration
// value of other: how many chromatic units the value of s is higher. Letters
// do not affect it: C sharp minus C flat is +2, and D flat minus C sharp is
// -2. It equals s.Alteration().Sub(other.Alteration()) for valid names.
//
// The difference is not an interval. If s or other is unset,
// AlterationDifference returns the natural and an error that matches
// [letter.ErrInvalid]. If the exact difference is outside the guaranteed range
// of package alteration, it returns the natural and an error that matches
// [alteration.ErrOutOfRange].
func (s Spelling) AlterationDifference(other Spelling) (alteration.Alteration, error) {
	if !s.IsValid() || !other.IsValid() {
		return alteration.Natural(), fmt.Errorf("spelling: alteration difference of %v and %v: %w", s, other, letter.ErrInvalid)
	}

	d, err := s.a.Sub(other.a)
	if err != nil {
		return alteration.Natural(), fmt.Errorf("spelling: alteration difference of %v and %v: %w", s, other, err)
	}

	return d, nil
}

// MustAlterationDifference is like Spelling.AlterationDifference but panics
// in the same cases. The panic value is the error returned by
// AlterationDifference.
//
// It is intended for values known to be valid and in range.
func (s Spelling) MustAlterationDifference(other Spelling) alteration.Alteration {
	return must(s.AlterationDifference(other))
}

// apply returns the name with the letter of s and the value op(s.Alteration()).
// It checks that s is set before calling op and returns the unset name with an
// error on failure.
func (s Spelling) apply(name string, op func(alteration.Alteration) (alteration.Alteration, error)) (Spelling, error) {
	if !s.IsValid() {
		return Spelling{}, fmt.Errorf("spelling: %s: unset name: %w", name, letter.ErrInvalid)
	}

	a, err := op(s.a)
	if err != nil {
		return Spelling{}, fmt.Errorf("spelling: %s %v: %w", name, s, err)
	}

	return Spelling{l: s.l, a: a}, nil
}

// must returns v if err is nil and panics with err otherwise.
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}
