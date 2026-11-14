package tempo

import (
	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tempo/term"
)

// Approx is an approximate tempo marking: a written unit and a named number,
// such as ♩ ≈ 108. Which sign of approximation the record used — c., ca.,
// circa, or ≈ — is not stored. An approximate marking is a kind of its own: it
// never equals an exact marking with the same number, has no tolerance and no
// "contains" operation, and becomes a Tempo only through the explicit
// operation Tempo, which takes the named number.
//
// The zero value Approx{} is the unset marking. Approx is comparable with ==,
// which matches Equal, and can be used as a map key.
type Approx struct {
	m mark
}

// NewApprox returns the approximate marking with the given unit and named
// number. In order of checking, it returns an error wrapping
// ErrInvalidArgument if unit or then bpm is unset, and ErrNoTempo if bpm is
// zero or negative.
func NewApprox(unit duration.Written, bpm fraction.Fraction) (Approx, error) {
	m, err := newMark("NewApprox", unit, bpm)
	return Approx{m: m}, err
}

// MustNewApprox is like NewApprox but panics with the returned error.
func MustNewApprox(unit duration.Written, bpm fraction.Fraction) Approx {
	return must(NewApprox(unit, bpm))
}

// IsSet reports whether a is set, that is, not the unset Approx{}.
func (a Approx) IsSet() bool {
	return a.m.isSet()
}

// unset returns an error wrapping ErrUnset for the receiver of op.
func (Approx) unset(op string) error {
	return unsetError(op, "receiver")
}

// Unit returns the written unit of a. It returns an error wrapping ErrUnset
// if a is unset.
func (a Approx) Unit() (duration.Written, error) {
	if !a.IsSet() {
		return duration.Written{}, a.unset("Approx.Unit")
	}
	return a.m.unit, nil
}

// MustUnit is like Unit but panics with the returned error.
func (a Approx) MustUnit() duration.Written {
	return must(a.Unit())
}

// UnitDuration returns the duration of the unit of a. It returns an error
// wrapping ErrUnset if a is unset.
func (a Approx) UnitDuration() (duration.Duration, error) {
	if !a.IsSet() {
		return duration.Duration{}, a.unset("Approx.UnitDuration")
	}
	return a.m.unitDuration(), nil
}

// MustUnitDuration is like UnitDuration but panics with the returned error.
func (a Approx) MustUnitDuration() duration.Duration {
	return must(a.UnitDuration())
}

// BPM returns the named number of a in reduced form. It returns an error
// wrapping ErrUnset if a is unset.
func (a Approx) BPM() (fraction.Fraction, error) {
	if !a.IsSet() {
		return fraction.Fraction{}, a.unset("Approx.BPM")
	}
	return a.m.bpm, nil
}

// MustBPM is like BPM but panics with the returned error.
func (a Approx) MustBPM() fraction.Fraction {
	return must(a.BPM())
}

// Speed returns the speed of the named number, u·N whole notes per minute. In
// order of checking, it returns an error wrapping ErrUnset if a is unset and
// ErrOutOfRange together with fraction.ErrOutOfRange if the result is outside
// the range of package common/fraction.
func (a Approx) Speed() (fraction.Fraction, error) {
	const op = "Approx.Speed"
	if !a.IsSet() {
		return fraction.Fraction{}, a.unset(op)
	}
	return a.m.speedFraction(op)
}

// MustSpeed is like Speed but panics with the returned error.
func (a Approx) MustSpeed() fraction.Fraction {
	return must(a.Speed())
}

// QuartersPerMinute returns 4·u·N for the named number. The errors are those
// of Speed.
func (a Approx) QuartersPerMinute() (fraction.Fraction, error) {
	const op = "Approx.QuartersPerMinute"
	if !a.IsSet() {
		return fraction.Fraction{}, a.unset(op)
	}
	return a.m.quartersPerMinute(op)
}

// MustQuartersPerMinute is like QuartersPerMinute but panics with the returned
// error.
func (a Approx) MustQuartersPerMinute() fraction.Fraction {
	return must(a.QuartersPerMinute())
}

// Equal reports whether a and other are the same record: equal units and equal
// named numbers. Two unset markings are equal. Equal matches ==.
func (a Approx) Equal(other Approx) bool {
	return a == other
}

// Compare is a total order for sorting that returns −1, 0, or +1: by the
// speed of the named number, and for equal speeds by unit as
// duration.Written.Compare orders them. The unset marking precedes every set
// one. Compare returns 0 exactly when a == other.
func (a Approx) Compare(other Approx) int {
	return compareMarks(a.m, other.m)
}

// checkPair returns the error of an unset receiver or argument.
func (a Approx) checkPair(op string, other Approx) error {
	if !a.IsSet() {
		return a.unset(op)
	}
	if !other.IsSet() {
		return unsetError(op, "argument")
	}
	return nil
}

// EqualSpeed reports whether the named numbers of a and other give equal
// speeds. It returns an error wrapping ErrUnset if a or then other is unset.
func (a Approx) EqualSpeed(other Approx) (bool, error) {
	if err := a.checkPair("Approx.EqualSpeed", other); err != nil {
		return false, err
	}
	return a.m.speed().Cmp(other.m.speed()) == 0, nil
}

// MustEqualSpeed is like EqualSpeed but panics with the returned error.
func (a Approx) MustEqualSpeed(other Approx) bool {
	return must(a.EqualSpeed(other))
}

// CompareSpeed returns +1 if the named number of a gives a faster speed than
// that of other, −1 if slower, and 0 for equal speeds. It returns an error
// wrapping ErrUnset if a or then other is unset.
func (a Approx) CompareSpeed(other Approx) (int, error) {
	if err := a.checkPair("Approx.CompareSpeed", other); err != nil {
		return 0, err
	}
	return a.m.speed().Cmp(other.m.speed()), nil
}

// MustCompareSpeed is like CompareSpeed but panics with the returned error.
func (a Approx) MustCompareSpeed(other Approx) int {
	return must(a.CompareSpeed(other))
}

// WithUnit returns the approximate marking of the same speed with unit u: ♩. ≈
// 60 becomes ♪ ≈ 180. In order of checking, it returns an error wrapping
// ErrUnset if a is unset, ErrInvalidArgument if unit is unset, and
// ErrOutOfRange together with fraction.ErrOutOfRange if the result is outside
// the range of package common/fraction.
func (a Approx) WithUnit(unit duration.Written) (Approx, error) {
	const op = "Approx.WithUnit"
	if !a.IsSet() {
		return Approx{}, a.unset(op)
	}
	m, err := a.m.withUnit(op, unit)
	return Approx{m: m}, err
}

// MustWithUnit is like WithUnit but panics with the returned error.
func (a Approx) MustWithUnit(unit duration.Written) Approx {
	return must(a.WithUnit(unit))
}

// Tempo returns the exact marking with the unit and the named number of a: ♩
// ≈ 108 gives ♩ = 108. This is the explicit choice of the named number; no
// other number is ever substituted. It returns an error wrapping ErrUnset if
// a is unset.
func (a Approx) Tempo() (Tempo, error) {
	if !a.IsSet() {
		return Tempo{}, a.unset("Approx.Tempo")
	}
	return Tempo(a), nil
}

// MustTempo is like Tempo but panics with the returned error.
func (a Approx) MustTempo() Tempo {
	return must(a.Tempo())
}

// TermsIn returns the terms of l whose guides contain the named number of a,
// as l.At does; the unit is not taken into account. If there are none, it
// returns nil without an error. In order of checking, it returns an error
// wrapping ErrUnset if a is unset and ErrInvalidArgument if l is unset.
func (a Approx) TermsIn(l term.List) ([]term.Term, error) {
	const op = "Approx.TermsIn"
	if !a.IsSet() {
		return nil, a.unset(op)
	}
	return termsAt(op, l, a.m.bpm)
}

// MustTermsIn is like TermsIn but panics with the returned error.
func (a Approx) MustTermsIn(l term.List) []term.Term {
	return must(a.TermsIn(l))
}
