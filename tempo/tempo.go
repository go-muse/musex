package tempo

import (
	"math/big"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tempo/term"
)

// Tempo is an exact tempo marking: a written unit u and beats per minute N,
// such as ♩. = 60 with the unit of a dotted quarter and 60 beats per minute.
// The unit is any written duration of one sign, with dots and tuplets; beats
// per minute are any exact positive fraction. The speed of a tempo is u·N
// whole notes per minute, and a duration d lasts 60·d/(u·N) seconds.
//
// The unit is stored as written and beats per minute in reduced form.
// Equal compares the record: ♩. = 60 and ♪ = 180 are different records of the
// same speed, which EqualSpeed compares. The zero value Tempo{} is the unset
// tempo. Tempo is comparable with ==, which matches Equal, and can be used as
// a map key.
type Tempo struct {
	m mark
}

// New returns the tempo with the given unit and beats per minute. Beats per
// minute need not be whole or lie on the metronome scale. In order of
// checking, it returns an error wrapping ErrInvalidArgument if unit or then
// bpm is unset, and ErrNoTempo if bpm is zero or negative.
func New(unit duration.Written, bpm fraction.Fraction) (Tempo, error) {
	m, err := newMark("New", unit, bpm)
	return Tempo{m: m}, err
}

// MustNew is like New but panics with the returned error.
func MustNew(unit duration.Written, bpm fraction.Fraction) Tempo {
	return must(New(unit, bpm))
}

// quarter is the unit of a tempo given by beats per minute alone: a quarter
// without dots and tuplets.
var quarter = duration.Quarter.MustWritten()

// NewQuarter returns the tempo with the unit of a quarter, which a number of
// beats per minute alone means. In order of checking, it returns an error
// wrapping ErrInvalidArgument if bpm is unset and ErrNoTempo if bpm is zero or
// negative.
func NewQuarter(bpm fraction.Fraction) (Tempo, error) {
	m, err := newMark("NewQuarter", quarter, bpm)
	return Tempo{m: m}, err
}

// MustNewQuarter is like NewQuarter but panics with the returned error.
func MustNewQuarter(bpm fraction.Fraction) Tempo {
	return must(NewQuarter(bpm))
}

// FromSeconds returns the tempo with unit u in which the duration d lasts s
// seconds: 60·d/(u·s) beats per minute. In order of checking, it returns an
// error wrapping ErrUnset if s is unset, ErrInvalidArgument if unit or then d
// is unset, ErrNoSolution if s is zero, and ErrOutOfRange together with
// fraction.ErrOutOfRange if the result is outside the range of package
// common/fraction.
func FromSeconds(unit duration.Written, d duration.Duration, s Seconds) (Tempo, error) {
	const op = "FromSeconds"
	switch {
	case !s.IsSet():
		return Tempo{}, unsetError(op, "time")
	case !unit.IsSet():
		return Tempo{}, invalidError(op, "unset unit")
	case !d.IsSet():
		return Tempo{}, invalidError(op, "unset duration")
	case s.v.Sign() == 0:
		return Tempo{}, simpleError(ErrNoSolution, op, "zero time")
	}
	m, err := markOf(op, unit, quo(mul(ratInt(60), durRat(d)), mul(unitRat(unit), ratOf(s.v))))
	return Tempo{m: m}, err
}

// MustFromSeconds is like FromSeconds but panics with the returned error.
func MustFromSeconds(unit duration.Written, d duration.Duration, s Seconds) Tempo {
	return must(FromSeconds(unit, d, s))
}

// IsSet reports whether t is set, that is, not the unset Tempo{}.
func (t Tempo) IsSet() bool {
	return t.m.isSet()
}

// unset returns an error wrapping ErrUnset for the receiver of op.
func (Tempo) unset(op string) error {
	return unsetError(op, "receiver")
}

// Unit returns the written unit of t. It returns an error wrapping ErrUnset if
// t is unset.
func (t Tempo) Unit() (duration.Written, error) {
	if !t.IsSet() {
		return duration.Written{}, t.unset("Tempo.Unit")
	}
	return t.m.unit, nil
}

// MustUnit is like Unit but panics with the returned error.
func (t Tempo) MustUnit() duration.Written {
	return must(t.Unit())
}

// UnitDuration returns the duration u of the unit of t. It returns an error
// wrapping ErrUnset if t is unset.
func (t Tempo) UnitDuration() (duration.Duration, error) {
	if !t.IsSet() {
		return duration.Duration{}, t.unset("Tempo.UnitDuration")
	}
	return t.m.unitDuration(), nil
}

// MustUnitDuration is like UnitDuration but panics with the returned error.
func (t Tempo) MustUnitDuration() duration.Duration {
	return must(t.UnitDuration())
}

// BPM returns the beats per minute N of t in reduced form. It returns an
// error wrapping ErrUnset if t is unset.
func (t Tempo) BPM() (fraction.Fraction, error) {
	if !t.IsSet() {
		return fraction.Fraction{}, t.unset("Tempo.BPM")
	}
	return t.m.bpm, nil
}

// MustBPM is like BPM but panics with the returned error.
func (t Tempo) MustBPM() fraction.Fraction {
	return must(t.BPM())
}

// Speed returns the speed u·N of t in whole notes per minute: 45/2 for
// ♩. = 60. In order of checking, it returns an error wrapping ErrUnset if t is
// unset and ErrOutOfRange together with fraction.ErrOutOfRange if the result
// is outside the range of package common/fraction.
func (t Tempo) Speed() (fraction.Fraction, error) {
	const op = "Tempo.Speed"
	if !t.IsSet() {
		return fraction.Fraction{}, t.unset(op)
	}
	return t.m.speedFraction(op)
}

// MustSpeed is like Speed but panics with the returned error.
func (t Tempo) MustSpeed() fraction.Fraction {
	return must(t.Speed())
}

// QuartersPerMinute returns 4·u·N, the number of quarters per minute: 90 for
// ♩. = 60. The errors are those of Speed.
func (t Tempo) QuartersPerMinute() (fraction.Fraction, error) {
	const op = "Tempo.QuartersPerMinute"
	if !t.IsSet() {
		return fraction.Fraction{}, t.unset(op)
	}
	return t.m.quartersPerMinute(op)
}

// MustQuartersPerMinute is like QuartersPerMinute but panics with the returned
// error.
func (t Tempo) MustQuartersPerMinute() fraction.Fraction {
	return must(t.QuartersPerMinute())
}

// Equal reports whether t and other are the same record: units equal as
// duration.Written.Equal compares them and equal beats per minute, so ♩ = 60
// equals ♩ = 120/2. Two unset tempos are equal. Equal matches ==.
func (t Tempo) Equal(other Tempo) bool {
	return t == other
}

// Compare is a total order for sorting that returns −1, 0, or +1: by speed,
// and for equal speeds by unit as duration.Written.Compare orders them, so
// ♪ = 180 < ♩ = 90 < ♩. = 60. The unset tempo precedes every set one. Compare
// returns 0 exactly when t == other.
func (t Tempo) Compare(other Tempo) int {
	return compareMarks(t.m, other.m)
}

// checkPair returns the error of an unset receiver or argument.
func (t Tempo) checkPair(op string, other Tempo) error {
	if !t.IsSet() {
		return t.unset(op)
	}
	if !other.IsSet() {
		return unsetError(op, "argument")
	}
	return nil
}

// EqualSpeed reports whether t and other have equal speeds: ♩. = 60, ♪ = 180,
// and ♩ = 90 do. It returns an error wrapping ErrUnset if t or then other is
// unset.
func (t Tempo) EqualSpeed(other Tempo) (bool, error) {
	if err := t.checkPair("Tempo.EqualSpeed", other); err != nil {
		return false, err
	}
	return t.m.speed().Cmp(other.m.speed()) == 0, nil
}

// MustEqualSpeed is like EqualSpeed but panics with the returned error.
func (t Tempo) MustEqualSpeed(other Tempo) bool {
	return must(t.EqualSpeed(other))
}

// CompareSpeed returns +1 if t is faster than other, −1 if it is slower, and 0
// for equal speeds. It returns an error wrapping ErrUnset if t or then other is
// unset.
func (t Tempo) CompareSpeed(other Tempo) (int, error) {
	if err := t.checkPair("Tempo.CompareSpeed", other); err != nil {
		return 0, err
	}
	return t.m.speed().Cmp(other.m.speed()), nil
}

// MustCompareSpeed is like CompareSpeed but panics with the returned error.
func (t Tempo) MustCompareSpeed(other Tempo) int {
	return must(t.CompareSpeed(other))
}

// SpeedRatio returns how many times t is faster than other, (u·N)/(u′·N′):
// 4/3 for ♩ = 120 against ♩. = 60. In order of checking, it returns an error
// wrapping ErrUnset if t or then other is unset, and ErrOutOfRange together
// with fraction.ErrOutOfRange if the result is outside the range of package
// common/fraction.
func (t Tempo) SpeedRatio(other Tempo) (fraction.Fraction, error) {
	const op = "Tempo.SpeedRatio"
	if err := t.checkPair(op, other); err != nil {
		return fraction.Fraction{}, err
	}
	return toFraction(op, quo(t.m.speed(), other.m.speed()))
}

// MustSpeedRatio is like SpeedRatio but panics with the returned error.
func (t Tempo) MustSpeedRatio(other Tempo) fraction.Fraction {
	return must(t.SpeedRatio(other))
}

// WithUnit returns the tempo of the same speed with unit u: N·u/u′ beats per
// minute, so ♩. = 60 becomes ♪ = 180 and ♩ = 100 becomes ♩. = 200/3. In order
// of checking, it returns an error wrapping ErrUnset if t is unset,
// ErrInvalidArgument if unit is unset, and ErrOutOfRange together with
// fraction.ErrOutOfRange if the result is outside the range of package
// common/fraction.
func (t Tempo) WithUnit(unit duration.Written) (Tempo, error) {
	const op = "Tempo.WithUnit"
	if !t.IsSet() {
		return Tempo{}, t.unset(op)
	}
	m, err := t.m.withUnit(op, unit)
	return Tempo{m: m}, err
}

// MustWithUnit is like WithUnit but panics with the returned error.
func (t Tempo) MustWithUnit(unit duration.Written) Tempo {
	return must(t.WithUnit(unit))
}

// WithBPM returns the tempo with the unit of t and beats per minute bpm. In
// order of checking, it returns an error wrapping ErrUnset if t is unset,
// ErrInvalidArgument if bpm is unset, and ErrNoTempo if bpm is zero or
// negative.
func (t Tempo) WithBPM(bpm fraction.Fraction) (Tempo, error) {
	const op = "Tempo.WithBPM"
	if !t.IsSet() {
		return Tempo{}, t.unset(op)
	}
	m, err := newMark(op, t.m.unit, bpm)
	return Tempo{m: m}, err
}

// MustWithBPM is like WithBPM but panics with the returned error.
func (t Tempo) MustWithBPM(bpm fraction.Fraction) Tempo {
	return must(t.WithBPM(bpm))
}

// Mul returns the tempo k times as fast with the same unit, N·k beats per
// minute: 3/2 makes ♩ = 60 into ♩ = 90. In order of checking, it returns an
// error wrapping ErrUnset if t is unset, ErrInvalidArgument if k is unset,
// zero, or negative, and ErrOutOfRange together with fraction.ErrOutOfRange if
// the result is outside the range of package common/fraction.
func (t Tempo) Mul(k fraction.Fraction) (Tempo, error) {
	const op = "Tempo.Mul"
	if !t.IsSet() {
		return Tempo{}, t.unset(op)
	}
	if unsetFraction(k) || k.Sign() <= 0 {
		return Tempo{}, invalidError(op, "factor unset or not positive")
	}
	return t.scale(op, ratOf(k))
}

// MustMul is like Mul but panics with the returned error.
func (t Tempo) MustMul(k fraction.Fraction) Tempo {
	return must(t.Mul(k))
}

// scale returns the set tempo t with beats per minute multiplied by k > 0.
func (t Tempo) scale(op string, k *big.Rat) (Tempo, error) {
	m, err := markOf(op, t.m.unit, mul(ratOf(t.m.bpm), k))
	return Tempo{m: m}, err
}

// Double returns the tempo twice as fast, the doppio movimento: Mul(2). In
// order of checking, it returns an error wrapping ErrUnset if t is unset and
// ErrOutOfRange together with fraction.ErrOutOfRange if the result is outside
// the range of package common/fraction.
func (t Tempo) Double() (Tempo, error) {
	const op = "Tempo.Double"
	if !t.IsSet() {
		return Tempo{}, t.unset(op)
	}
	return t.scale(op, ratInt(2))
}

// MustDouble is like Double but panics with the returned error.
func (t Tempo) MustDouble() Tempo {
	return must(t.Double())
}

// Halve returns the tempo half as fast: Mul(1/2). The errors are those of
// Double.
func (t Tempo) Halve() (Tempo, error) {
	const op = "Tempo.Halve"
	if !t.IsSet() {
		return Tempo{}, t.unset(op)
	}
	return t.scale(op, big.NewRat(1, 2))
}

// MustHalve is like Halve but panics with the returned error.
func (t Tempo) MustHalve() Tempo {
	return must(t.Halve())
}

// Modulate returns the tempo after a metric modulation "new newUnit = old
// old": the unit newUnit lasts as long as the duration old lasted in t, so the
// result is newUnit = N·u/w beats per minute, where w is the duration of old.
// For ♩ = 120, a dotted quarter equal to the old quarter gives ♩. = 120. old
// can be any duration, written, tied, or without notation. In order of
// checking, it returns an error wrapping ErrUnset if t is unset,
// ErrInvalidArgument if newUnit or then old is unset, and ErrOutOfRange
// together with fraction.ErrOutOfRange if the result is outside the range of
// package common/fraction.
func (t Tempo) Modulate(newUnit duration.Written, old duration.Duration) (Tempo, error) {
	const op = "Tempo.Modulate"
	switch {
	case !t.IsSet():
		return Tempo{}, t.unset(op)
	case !newUnit.IsSet():
		return Tempo{}, invalidError(op, "unset new unit")
	case !old.IsSet():
		return Tempo{}, invalidError(op, "unset old duration")
	}
	m, err := markOf(op, newUnit, quo(t.m.speed(), durRat(old)))
	return Tempo{m: m}, err
}

// MustModulate is like Modulate but panics with the returned error.
func (t Tempo) MustModulate(newUnit duration.Written, old duration.Duration) Tempo {
	return must(t.Modulate(newUnit, old))
}

// ModulationTo returns the duration w of t that equals the unit of next, so
// that the change can be written "new unit = old w": w = N·u/N′. From ♩ = 120
// to ♩. = 120 it is 1/4. Writing w as a note is an operation of package
// duration. In order of checking, it returns an error wrapping ErrUnset if t
// or then next is unset, and ErrOutOfRange together with
// fraction.ErrOutOfRange if the result is outside the range of package
// common/fraction.
func (t Tempo) ModulationTo(next Tempo) (duration.Duration, error) {
	const op = "Tempo.ModulationTo"
	if err := t.checkPair(op, next); err != nil {
		return duration.Duration{}, err
	}
	return toDuration(op, quo(t.m.speed(), ratOf(next.m.bpm)))
}

// MustModulationTo is like ModulationTo but panics with the returned error.
func (t Tempo) MustModulationTo(next Tempo) duration.Duration {
	return must(t.ModulationTo(next))
}

// Seconds returns the time that the duration d lasts in t: 60·d/(u·N)
// seconds, so a half lasts 2 seconds at ♩ = 60. In order of checking, it
// returns an error wrapping ErrUnset if t is unset, ErrInvalidArgument if d is
// unset, and ErrOutOfRange together with fraction.ErrOutOfRange if the result
// is outside the range of package common/fraction.
func (t Tempo) Seconds(d duration.Duration) (Seconds, error) {
	const op = "Tempo.Seconds"
	if !t.IsSet() {
		return Seconds{}, t.unset(op)
	}
	if !d.IsSet() {
		return Seconds{}, invalidError(op, "unset duration")
	}
	return toSeconds(op, quo(mul(ratInt(60), durRat(d)), t.m.speed()))
}

// MustSeconds is like Seconds but panics with the returned error.
func (t Tempo) MustSeconds(d duration.Duration) Seconds {
	return must(t.Seconds(d))
}

// UnitSeconds returns the time of one unit, 60/N seconds. In order of
// checking, it returns an error wrapping ErrUnset if t is unset and
// ErrOutOfRange together with fraction.ErrOutOfRange if the result is outside
// the range of package common/fraction.
func (t Tempo) UnitSeconds() (Seconds, error) {
	const op = "Tempo.UnitSeconds"
	if !t.IsSet() {
		return Seconds{}, t.unset(op)
	}
	return toSeconds(op, quo(ratInt(60), ratOf(t.m.bpm)))
}

// MustUnitSeconds is like UnitSeconds but panics with the returned error.
func (t Tempo) MustUnitSeconds() Seconds {
	return must(t.UnitSeconds())
}

// UnitsPerSecond returns the number of units per second, N/60. The errors are
// those of UnitSeconds.
func (t Tempo) UnitsPerSecond() (fraction.Fraction, error) {
	const op = "Tempo.UnitsPerSecond"
	if !t.IsSet() {
		return fraction.Fraction{}, t.unset(op)
	}
	return toFraction(op, quo(ratOf(t.m.bpm), ratInt(60)))
}

// MustUnitsPerSecond is like UnitsPerSecond but panics with the returned
// error.
func (t Tempo) MustUnitsPerSecond() fraction.Fraction {
	return must(t.UnitsPerSecond())
}

// minutes returns the time s in minutes, s/60.
func minutes(s Seconds) *big.Rat {
	return quo(ratOf(s.v), ratInt(60))
}

// checkTime returns the error of an unset receiver or an unset time.
func (t Tempo) checkTime(op string, s Seconds) error {
	if !t.IsSet() {
		return t.unset(op)
	}
	if !s.IsSet() {
		return unsetError(op, "time")
	}
	return nil
}

// DurationIn returns the duration that lasts s seconds in t: s·u·N/60. In
// order of checking, it returns an error wrapping ErrUnset if t or then s is
// unset, ErrNoSolution if s is zero, and ErrOutOfRange together with
// fraction.ErrOutOfRange if the result is outside the range of package
// common/fraction.
func (t Tempo) DurationIn(s Seconds) (duration.Duration, error) {
	const op = "Tempo.DurationIn"
	if err := t.checkTime(op, s); err != nil {
		return duration.Duration{}, err
	}
	if s.v.Sign() == 0 {
		return duration.Duration{}, simpleError(ErrNoSolution, op, "zero time")
	}
	return toDuration(op, mul(minutes(s), t.m.speed()))
}

// MustDurationIn is like DurationIn but panics with the returned error.
func (t Tempo) MustDurationIn(s Seconds) duration.Duration {
	return must(t.DurationIn(s))
}

// UnitsIn returns the number of units in s seconds, N·s/60; zero seconds give
// zero units. In order of checking, it returns an error wrapping ErrUnset if t
// or then s is unset, and ErrOutOfRange together with fraction.ErrOutOfRange
// if the result is outside the range of package common/fraction.
func (t Tempo) UnitsIn(s Seconds) (fraction.Fraction, error) {
	const op = "Tempo.UnitsIn"
	if err := t.checkTime(op, s); err != nil {
		return fraction.Fraction{}, err
	}
	return toFraction(op, mul(ratOf(t.m.bpm), minutes(s)))
}

// MustUnitsIn is like UnitsIn but panics with the returned error.
func (t Tempo) MustUnitsIn(s Seconds) fraction.Fraction {
	return must(t.UnitsIn(s))
}

// SecondsOfUnits returns the time of k units, 60·k/N seconds; zero units take
// zero seconds. In order of checking, it returns an error wrapping ErrUnset if
// t is unset, ErrInvalidArgument if k is unset, ErrNoSolution if k is
// negative, and ErrOutOfRange together with fraction.ErrOutOfRange if the
// result is outside the range of package common/fraction.
func (t Tempo) SecondsOfUnits(k fraction.Fraction) (Seconds, error) {
	const op = "Tempo.SecondsOfUnits"
	switch {
	case !t.IsSet():
		return Seconds{}, t.unset(op)
	case unsetFraction(k):
		return Seconds{}, invalidError(op, "unset number of units")
	case k.Sign() < 0:
		return Seconds{}, simpleError(ErrNoSolution, op, "negative number of units")
	}
	return toSeconds(op, quo(mul(ratInt(60), ratOf(k)), ratOf(t.m.bpm)))
}

// MustSecondsOfUnits is like SecondsOfUnits but panics with the returned
// error.
func (t Tempo) MustSecondsOfUnits(k fraction.Fraction) Seconds {
	return must(t.SecondsOfUnits(k))
}

// Quantize returns the tempo with the unit of t and beats per minute moved to
// a node k·step of the grid of multiples of step, chosen by mode; a value on a
// node is unchanged. There is no default mode. ♩. = 200/3 with step 1 becomes
// ♩. = 67 with HalfUp and ♩. = 66 with Down. In order of checking, it returns
// an error wrapping ErrUnset if t is unset, ErrInvalidArgument together with
// rounding.ErrUnset or rounding.ErrUnknownMode if mode is unset or unknown,
// ErrInvalidArgument if step is unset, zero, or negative, ErrQuantizedToZero
// if the node is zero, and ErrOutOfRange together with fraction.ErrOutOfRange
// if the result is outside the range of package common/fraction.
func (t Tempo) Quantize(step fraction.Fraction, mode rounding.Mode) (Tempo, error) {
	const op = "Tempo.Quantize"
	if !t.IsSet() {
		return Tempo{}, t.unset(op)
	}
	if err := checkMode(op, mode); err != nil {
		return Tempo{}, err
	}
	if unsetFraction(step) || step.Sign() <= 0 {
		return Tempo{}, invalidError(op, "step unset or not positive")
	}
	_, v, err := mode.RatMultiple(ratOf(t.m.bpm), ratOf(step))
	if err != nil {
		return Tempo{}, modeError(op, mode)
	}
	if v.Sign() == 0 {
		return Tempo{}, simpleError(ErrQuantizedToZero, op, "node zero")
	}
	m, err := markOf(op, t.m.unit, v)
	return Tempo{m: m}, err
}

// MustQuantize is like Quantize but panics with the returned error.
func (t Tempo) MustQuantize(step fraction.Fraction, mode rounding.Mode) Tempo {
	return must(t.Quantize(step, mode))
}

// SnapToScale returns the tempo with the unit of t and beats per minute
// replaced by the division of the metronome scale that NearestScaleDivision
// chooses by mode: ♩. = 62 becomes ♩. = 63 with HalfUp. In order of checking,
// it returns an error wrapping ErrUnset if t is unset, ErrInvalidArgument
// together with rounding.ErrUnset or rounding.ErrUnknownMode if mode is unset
// or unknown, and ErrOffScale if beats per minute are outside 40..208.
func (t Tempo) SnapToScale(mode rounding.Mode) (Tempo, error) {
	const op = "Tempo.SnapToScale"
	if !t.IsSet() {
		return Tempo{}, t.unset(op)
	}
	if err := checkMode(op, mode); err != nil {
		return Tempo{}, err
	}
	div, err := nearest(op, t.m.bpm, mode)
	if err != nil {
		return Tempo{}, err
	}
	return Tempo{m: mark{unit: t.m.unit, bpm: fraction.FromInt(div)}}, nil
}

// MustSnapToScale is like SnapToScale but panics with the returned error.
func (t Tempo) MustSnapToScale(mode rounding.Mode) Tempo {
	return must(t.SnapToScale(mode))
}

// TermsIn returns the terms of l whose guides contain the beats per minute of
// t, as l.At does; the unit is not taken into account. If there are none, it
// returns nil without an error. In order of checking, it returns an error
// wrapping ErrUnset if t is unset and ErrInvalidArgument if l is unset.
func (t Tempo) TermsIn(l term.List) ([]term.Term, error) {
	const op = "Tempo.TermsIn"
	if !t.IsSet() {
		return nil, t.unset(op)
	}
	return termsAt(op, l, t.m.bpm)
}

// MustTermsIn is like TermsIn but panics with the returned error.
func (t Tempo) MustTermsIn(l term.List) []term.Term {
	return must(t.TermsIn(l))
}

// termsAt returns the terms of l at the positive number beats.
func termsAt(op string, l term.List, beats fraction.Fraction) ([]term.Term, error) {
	if !l.IsSet() {
		return nil, invalidError(op, "unset list")
	}
	terms, _ := l.At(beats)
	return terms, nil
}
