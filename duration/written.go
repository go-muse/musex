package duration

import (
	"cmp"
	"encoding/binary"
	"math/big"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/tuplet"
)

// Written is a written duration: a base, a number of dots, and a chain of
// tuplets listed from the outermost to the innermost. Its duration is
// b·(2 − 1/2^k)·C, where b is the duration of the base, k the number of dots,
// and C the coefficient of the chain (1 for an empty chain).
//
// The notation is stored as given: a quarter in 3:2 and a quarter in 6:4 are
// different written durations of equal duration, and so are the chains
// [3:2, 5:4] and [5:4, 3:2]. Equal compares the notation and matches ==;
// EqualDuration compares durations.
//
// The zero value Written{} is the unset written duration. Written is
// comparable with == and can be used as a map key.
type Written struct {
	base  Base
	dots  int
	chain string // packed chain, see packChain
	value fraction.Fraction
}

// tupletBytes is the size of one packed tuplet: the actual and the normal
// count as big-endian 64-bit integers.
const tupletBytes = 16

// minDotsOutOfRange is the smallest number of dots for which the duration of
// a written duration is outside the range of package common/fraction for
// every base and chain. The duration is 2^e·(2^(k+1) − 1)·M/(2^k·N) with
// e ≤ 3 and a reduced coefficient M/N whose binary power is at most 2^254.
// The odd factor 2^(k+1) − 1 cancels no power of two, so the reduced
// denominator is at least 2^(k−257), which exceeds fraction.MaxComponent for
// k ≥ 512.
const minDotsOutOfRange = 512

// packChain returns the canonical immutable packing of a chain of set
// tuplets: equal notations give equal strings.
func packChain(chain []tuplet.Tuplet) string {
	if len(chain) == 0 {
		return ""
	}
	b := make([]byte, 0, len(chain)*tupletBytes)
	for _, t := range chain {
		actual, _ := t.Actual()
		normal, _ := t.Normal()
		b = binary.BigEndian.AppendUint64(b, uint64(actual))
		b = binary.BigEndian.AppendUint64(b, uint64(normal))
	}
	return string(b)
}

// unpackChain returns a new slice with the tuplets of a packed chain, or nil
// for an empty chain.
func unpackChain(s string) []tuplet.Tuplet {
	if s == "" {
		return nil
	}
	chain := make([]tuplet.Tuplet, len(s)/tupletBytes)
	for i := range chain {
		p := s[i*tupletBytes : (i+1)*tupletBytes]
		actual := uint64At(p[:8])
		normal := uint64At(p[8:])
		chain[i] = tuplet.MustNew(int(actual), int(normal))
	}
	return chain
}

// uint64At decodes a big-endian 64-bit integer from the 8 bytes of s.
func uint64At(s string) uint64 {
	var v uint64
	for i := range 8 {
		v = v<<8 | uint64(s[i])
	}
	return v
}

// checkChain returns an error wrapping ErrInvalidArgument and tuplet.ErrUnset
// for the first unset tuplet of chain.
func checkChain(op string, chain []tuplet.Tuplet) error {
	for i, t := range chain {
		if !t.IsSet() {
			return unsetTupletError(op, i)
		}
	}
	return nil
}

// chainCoefficient returns the coefficient value of a chain of set tuplets,
// or an error wrapping ErrOutOfRange, tuplet.ErrOutOfRange, and
// fraction.ErrOutOfRange if it is outside the range of package
// common/fraction.
func chainCoefficient(op string, chain []tuplet.Tuplet) (*big.Rat, error) {
	if !chainInRange(chain) {
		return nil, chainRangeError(op)
	}
	c, err := tuplet.ChainCoefficient(chain...)
	if err != nil {
		return nil, chainRangeError(op)
	}
	return ratOf(c), nil
}

// chainInRange reports whether both products of the counts of a chain of set
// tuplets are at most fraction.MaxComponent. The products do not decrease as
// factors are added, so the check stops at the first product out of range and
// its cost does not grow with the rest of a long chain.
func chainInRange(chain []tuplet.Tuplet) bool {
	limit := fraction.MaxComponentBig()
	actual, normal := big.NewInt(1), big.NewInt(1)
	for _, t := range chain {
		a, _ := t.Actual()
		n, _ := t.Normal()
		actual.Mul(actual, big.NewInt(int64(a)))
		normal.Mul(normal, big.NewInt(int64(n)))
		if actual.Cmp(limit) > 0 || normal.Cmp(limit) > 0 {
			return false
		}
	}
	return true
}

// checkedCoefficient checks chain for unset tuplets and returns its
// coefficient value.
func checkedCoefficient(op string, chain []tuplet.Tuplet) (*big.Rat, error) {
	if err := checkChain(op, chain); err != nil {
		return nil, err
	}
	return chainCoefficient(op, chain)
}

// writtenValue returns the duration of a valid base with dots ≥ 0 under a
// chain with coefficient coef, or an error wrapping ErrOutOfRange and
// fraction.ErrOutOfRange if it is outside the range of package
// common/fraction.
func writtenValue(op string, base Base, dots int, coef *big.Rat) (fraction.Fraction, error) {
	if dots >= minDotsOutOfRange {
		return fraction.Fraction{}, rangeError(op)
	}
	r := pow2Rat(base.exp())
	r.Mul(r, dotFactor(dots))
	r.Mul(r, coef)
	f, ok := fromRat(r)
	if !ok {
		return fraction.Fraction{}, rangeError(op)
	}
	return f, nil
}

// makeWritten checks the parts of a written duration in the order of
// NewWritten and returns it.
func makeWritten(op string, base Base, dots int, chain []tuplet.Tuplet) (Written, error) {
	if !base.IsValid() {
		return Written{}, invalidBaseError(op, base)
	}
	if dots < 0 {
		return Written{}, invalidError(op, "negative number of dots")
	}
	coef, err := checkedCoefficient(op, chain)
	if err != nil {
		return Written{}, err
	}
	v, err := writtenValue(op, base, dots, coef)
	if err != nil {
		return Written{}, err
	}
	return Written{base: base, dots: dots, chain: packChain(chain), value: v}, nil
}

// NewWritten returns the written duration with base, dots ≥ 0 dots, and the
// chain of tuplets listed from the outermost to the innermost; an empty chain
// means no tuplets. The number of dots is not limited by a rule, only by the
// range of package common/fraction. In order of checking, it returns an error
// wrapping ErrInvalidArgument if base is not a base, if dots < 0, or if a
// tuplet of the chain is unset (together with tuplet.ErrUnset), and
// ErrOutOfRange if the chain coefficient (together with tuplet.ErrOutOfRange
// and fraction.ErrOutOfRange) or the duration (together with
// fraction.ErrOutOfRange) is outside the range of package common/fraction.
// The chain is only read.
func NewWritten(base Base, dots int, chain ...tuplet.Tuplet) (Written, error) {
	return makeWritten("NewWritten", base, dots, chain)
}

// MustNewWritten is like NewWritten but panics with the returned error.
func MustNewWritten(base Base, dots int, chain ...tuplet.Tuplet) Written {
	return must(NewWritten(base, dots, chain...))
}

// IsSet reports whether w is a set written duration.
func (w Written) IsSet() bool {
	return w.base != 0
}

// unset returns the error for an unset receiver.
func (w Written) unset(op string) error {
	return unsetError(op, "receiver")
}

// Base returns the base of w. It returns an error wrapping ErrUnset if w is
// unset.
func (w Written) Base() (Base, error) {
	if !w.IsSet() {
		return 0, w.unset("Written.Base")
	}
	return w.base, nil
}

// MustBase is like Base but panics with the returned error.
func (w Written) MustBase() Base {
	return must(w.Base())
}

// Dots returns the number of dots of w. It returns an error wrapping ErrUnset
// if w is unset.
func (w Written) Dots() (int, error) {
	if !w.IsSet() {
		return 0, w.unset("Written.Dots")
	}
	return w.dots, nil
}

// MustDots is like Dots but panics with the returned error.
func (w Written) MustDots() int {
	return must(w.Dots())
}

// Chain returns a new slice with the tuplets of w as given, from the
// outermost to the innermost, or nil for an empty chain. The slice belongs to
// the caller. It returns an error wrapping ErrUnset if w is unset.
func (w Written) Chain() ([]tuplet.Tuplet, error) {
	if !w.IsSet() {
		return nil, w.unset("Written.Chain")
	}
	return unpackChain(w.chain), nil
}

// MustChain is like Chain but panics with the returned error.
func (w Written) MustChain() []tuplet.Tuplet {
	return must(w.Chain())
}

// Coefficient returns the coefficient of the chain of w as
// tuplet.ChainCoefficient does, with the unreduced stored pair
// (m1·m2·…, n1·n2·…); for an empty chain it is 1/1. It returns an error
// wrapping ErrUnset if w is unset.
func (w Written) Coefficient() (fraction.Fraction, error) {
	if !w.IsSet() {
		return fraction.Fraction{}, w.unset("Written.Coefficient")
	}
	c, err := tuplet.ChainCoefficient(unpackChain(w.chain)...)
	if err != nil {
		return fraction.Fraction{}, chainRangeError("Written.Coefficient")
	}
	return c, nil
}

// MustCoefficient is like Coefficient but panics with the returned error.
func (w Written) MustCoefficient() fraction.Fraction {
	return must(w.Coefficient())
}

// Duration returns the duration of w. It returns an error wrapping ErrUnset if
// w is unset.
func (w Written) Duration() (Duration, error) {
	if !w.IsSet() {
		return Duration{}, w.unset("Written.Duration")
	}
	return Duration{v: w.value}, nil
}

// MustDuration is like Duration but panics with the returned error.
func (w Written) MustDuration() Duration {
	return must(w.Duration())
}

// HasDots reports whether w has at least one dot. It returns an error
// wrapping ErrUnset if w is unset.
func (w Written) HasDots() (bool, error) {
	if !w.IsSet() {
		return false, w.unset("Written.HasDots")
	}
	return w.dots > 0, nil
}

// MustHasDots is like HasDots but panics with the returned error.
func (w Written) MustHasDots() bool {
	return must(w.HasDots())
}

// HasTuplets reports whether the chain of w is not empty. It returns an error
// wrapping ErrUnset if w is unset.
func (w Written) HasTuplets() (bool, error) {
	if !w.IsSet() {
		return false, w.unset("Written.HasTuplets")
	}
	return w.chain != "", nil
}

// MustHasTuplets is like HasTuplets but panics with the returned error.
func (w Written) MustHasTuplets() bool {
	return must(w.HasTuplets())
}

// IsBasic reports whether w is a base without dots and tuplets. It returns an
// error wrapping ErrUnset if w is unset.
func (w Written) IsBasic() (bool, error) {
	if !w.IsSet() {
		return false, w.unset("Written.IsBasic")
	}
	return w.dots == 0 && w.chain == "", nil
}

// MustIsBasic is like IsBasic but panics with the returned error.
func (w Written) MustIsBasic() bool {
	return must(w.IsBasic())
}

// Equal reports whether w and other have the same notation: the same base,
// the same number of dots, and the same tuplets of the chain in the same
// order, compared by notation as tuplet.Tuplet.Equal does. It matches ==; the
// unset written duration is equal only to the unset written duration.
func (w Written) Equal(other Written) bool {
	return w == other
}

// Compare is a total order for sorting that returns −1, 0, or +1: by
// duration; for equal durations by the number of tuplets (fewer first), then
// by base (longer first), then by the number of dots (fewer first), and then
// by the tuplets of the chain in order, as tuplet.Tuplet.Compare does. The
// unset written duration precedes every set one. Compare returns 0 if and
// only if w == other.
func (w Written) Compare(other Written) int {
	if c := w.value.Compare(other.value); c != 0 || !w.IsSet() {
		return c
	}
	if c := cmp.Compare(len(w.chain), len(other.chain)); c != 0 {
		return c
	}
	if c := cmp.Compare(w.base, other.base); c != 0 {
		return c
	}
	if c := cmp.Compare(w.dots, other.dots); c != 0 {
		return c
	}
	if w.chain == other.chain {
		return 0
	}
	a, b := unpackChain(w.chain), unpackChain(other.chain)
	for i := range a {
		if c := a[i].Compare(b[i]); c != 0 {
			return c
		}
	}
	return 0
}

// durations checks that w and other are set.
func (w Written) durations(op string, other Written) error {
	if !w.IsSet() {
		return w.unset(op)
	}
	if !other.IsSet() {
		return unsetError(op, "argument")
	}
	return nil
}

// EqualDuration reports whether w and other have equal durations. It returns
// an error wrapping ErrUnset if w or other is unset.
func (w Written) EqualDuration(other Written) (bool, error) {
	if err := w.durations("Written.EqualDuration", other); err != nil {
		return false, err
	}
	return w.value == other.value, nil
}

// MustEqualDuration is like EqualDuration but panics with the returned error.
func (w Written) MustEqualDuration(other Written) bool {
	return must(w.EqualDuration(other))
}

// CompareDuration compares the durations of w and other and returns −1, 0, or
// +1. It returns an error wrapping ErrUnset if w or other is unset.
func (w Written) CompareDuration(other Written) (int, error) {
	if err := w.durations("Written.CompareDuration", other); err != nil {
		return 0, err
	}
	return w.value.Compare(other.value), nil
}

// MustCompareDuration is like CompareDuration but panics with the returned
// error.
func (w Written) MustCompareDuration(other Written) int {
	return must(w.CompareDuration(other))
}

// WithBase returns w with base b; the dots and the chain are kept. It returns
// an error wrapping ErrUnset if w is unset, ErrInvalidArgument if b is not a
// base, and ErrOutOfRange if the new duration is outside the range of package
// common/fraction.
func (w Written) WithBase(b Base) (Written, error) {
	const op = "Written.WithBase"
	if !w.IsSet() {
		return Written{}, w.unset(op)
	}
	return makeWritten(op, b, w.dots, unpackChain(w.chain))
}

// MustWithBase is like WithBase but panics with the returned error.
func (w Written) MustWithBase(b Base) Written {
	return must(w.WithBase(b))
}

// ShiftBase returns w with its base steps steps longer (steps > 0) or shorter
// (steps < 0); the dots and the chain are kept. It returns an error wrapping
// ErrUnset if w is unset, ErrOutsideSeries if the base would be beyond either
// end of the series, and ErrOutOfRange if the new duration is outside the
// range of package common/fraction.
func (w Written) ShiftBase(steps int) (Written, error) {
	const op = "Written.ShiftBase"
	if !w.IsSet() {
		return Written{}, w.unset(op)
	}
	b, err := w.base.shiftValid(op, steps)
	if err != nil {
		return Written{}, err
	}
	return makeWritten(op, b, w.dots, unpackChain(w.chain))
}

// MustShiftBase is like ShiftBase but panics with the returned error.
func (w Written) MustShiftBase(steps int) Written {
	return must(w.ShiftBase(steps))
}

// WithDots returns w with dots dots. It returns an error wrapping ErrUnset if
// w is unset, ErrInvalidArgument if dots < 0, and ErrOutOfRange if the new
// duration is outside the range of package common/fraction.
func (w Written) WithDots(dots int) (Written, error) {
	const op = "Written.WithDots"
	if !w.IsSet() {
		return Written{}, w.unset(op)
	}
	return makeWritten(op, w.base, dots, unpackChain(w.chain))
}

// MustWithDots is like WithDots but panics with the returned error.
func (w Written) MustWithDots(dots int) Written {
	return must(w.WithDots(dots))
}

// AddDot returns w with one more dot. It returns an error wrapping ErrUnset if
// w is unset and ErrOutOfRange if the new duration is outside the range of
// package common/fraction.
func (w Written) AddDot() (Written, error) {
	const op = "Written.AddDot"
	if !w.IsSet() {
		return Written{}, w.unset(op)
	}
	return makeWritten(op, w.base, w.dots+1, unpackChain(w.chain))
}

// MustAddDot is like AddDot but panics with the returned error.
func (w Written) MustAddDot() Written {
	return must(w.AddDot())
}

// RemoveDot returns w with one dot fewer. It returns an error wrapping
// ErrUnset if w is unset and ErrInvalidArgument if w has no dots.
func (w Written) RemoveDot() (Written, error) {
	const op = "Written.RemoveDot"
	if !w.IsSet() {
		return Written{}, w.unset(op)
	}
	if w.dots == 0 {
		return Written{}, invalidError(op, "no dots to remove")
	}
	return makeWritten(op, w.base, w.dots-1, unpackChain(w.chain))
}

// MustRemoveDot is like RemoveDot but panics with the returned error.
func (w Written) MustRemoveDot() Written {
	return must(w.RemoveDot())
}

// WithChain returns w with its chain replaced by chain, listed from the
// outermost to the innermost; no arguments give no tuplets. It returns an
// error wrapping ErrUnset if w is unset, ErrInvalidArgument together with
// tuplet.ErrUnset if a tuplet is unset, and ErrOutOfRange if the chain
// coefficient or the new duration is outside the range of package
// common/fraction. The chain is only read.
func (w Written) WithChain(chain ...tuplet.Tuplet) (Written, error) {
	const op = "Written.WithChain"
	if !w.IsSet() {
		return Written{}, w.unset(op)
	}
	return makeWritten(op, w.base, w.dots, chain)
}

// MustWithChain is like WithChain but panics with the returned error.
func (w Written) MustWithChain(chain ...tuplet.Tuplet) Written {
	return must(w.WithChain(chain...))
}

// AddOuterTuplet returns w with t added as the outermost tuplet, at the start
// of the chain. It returns an error wrapping ErrUnset if w is unset,
// ErrInvalidArgument together with tuplet.ErrUnset if t is unset, and
// ErrOutOfRange if the chain coefficient or the new duration is outside the
// range of package common/fraction.
func (w Written) AddOuterTuplet(t tuplet.Tuplet) (Written, error) {
	const op = "Written.AddOuterTuplet"
	if !w.IsSet() {
		return Written{}, w.unset(op)
	}
	return makeWritten(op, w.base, w.dots, append([]tuplet.Tuplet{t}, unpackChain(w.chain)...))
}

// MustAddOuterTuplet is like AddOuterTuplet but panics with the returned
// error.
func (w Written) MustAddOuterTuplet(t tuplet.Tuplet) Written {
	return must(w.AddOuterTuplet(t))
}

// AddInnerTuplet returns w with t added as the innermost tuplet, at the end of
// the chain. It returns an error wrapping ErrUnset if w is unset,
// ErrInvalidArgument together with tuplet.ErrUnset if t is unset, and
// ErrOutOfRange if the chain coefficient or the new duration is outside the
// range of package common/fraction.
func (w Written) AddInnerTuplet(t tuplet.Tuplet) (Written, error) {
	const op = "Written.AddInnerTuplet"
	if !w.IsSet() {
		return Written{}, w.unset(op)
	}
	return makeWritten(op, w.base, w.dots, append(unpackChain(w.chain), t))
}

// MustAddInnerTuplet is like AddInnerTuplet but panics with the returned
// error.
func (w Written) MustAddInnerTuplet(t tuplet.Tuplet) Written {
	return must(w.AddInnerTuplet(t))
}

// removeTuplet removes the first (outer) or the last tuplet of the chain.
func (w Written) removeTuplet(op string, outer bool) (Written, error) {
	if !w.IsSet() {
		return Written{}, w.unset(op)
	}
	chain := unpackChain(w.chain)
	if len(chain) == 0 {
		return Written{}, invalidError(op, "no tuplets to remove")
	}
	if outer {
		chain = chain[1:]
	} else {
		chain = chain[:len(chain)-1]
	}
	return makeWritten(op, w.base, w.dots, chain)
}

// RemoveOuterTuplet returns w without its outermost tuplet. It returns an
// error wrapping ErrUnset if w is unset, ErrInvalidArgument if the chain is
// empty, and ErrOutOfRange if the new duration is outside the range of
// package common/fraction.
func (w Written) RemoveOuterTuplet() (Written, error) {
	return w.removeTuplet("Written.RemoveOuterTuplet", true)
}

// MustRemoveOuterTuplet is like RemoveOuterTuplet but panics with the returned
// error.
func (w Written) MustRemoveOuterTuplet() Written {
	return must(w.RemoveOuterTuplet())
}

// RemoveInnerTuplet returns w without its innermost tuplet. It returns an
// error wrapping ErrUnset if w is unset, ErrInvalidArgument if the chain is
// empty, and ErrOutOfRange if the new duration is outside the range of
// package common/fraction.
func (w Written) RemoveInnerTuplet() (Written, error) {
	return w.removeTuplet("Written.RemoveInnerTuplet", false)
}

// MustRemoveInnerTuplet is like RemoveInnerTuplet but panics with the returned
// error.
func (w Written) MustRemoveInnerTuplet() Written {
	return must(w.RemoveInnerTuplet())
}

// RemoveTuplets returns w with an empty chain; for an empty chain it returns
// an equal value without an error. It returns an error wrapping ErrUnset if w
// is unset and ErrOutOfRange if the new duration is outside the range of
// package common/fraction.
func (w Written) RemoveTuplets() (Written, error) {
	const op = "Written.RemoveTuplets"
	if !w.IsSet() {
		return Written{}, w.unset(op)
	}
	return makeWritten(op, w.base, w.dots, nil)
}

// MustRemoveTuplets is like RemoveTuplets but panics with the returned error.
func (w Written) MustRemoveTuplets() Written {
	return must(w.RemoveTuplets())
}

// Subdivide returns the normal subdivision of w: a base without dots divides
// into 2 parts and a base with one dot into 3 parts, each part being the base
// one step shorter, without dots, under the same chain. So a quarter gives 2
// eighths and a dotted quarter 3 eighths, and count·part equals the duration
// of w. In order of checking, it returns an error wrapping ErrUnset if w is
// unset, ErrNoSubdivision if w has two or more dots, ErrOutsideSeries if the
// part would be shorter than the 1024th, and ErrOutOfRange if the duration of
// the part is outside the range of package common/fraction.
func (w Written) Subdivide() (count int, part Written, err error) {
	const op = "Written.Subdivide"
	if !w.IsSet() {
		return 0, Written{}, w.unset(op)
	}
	if w.dots >= 2 {
		return 0, Written{}, simpleError(ErrNoSubdivision, op)
	}
	b, err := w.base.shiftValid(op, -1)
	if err != nil {
		return 0, Written{}, err
	}
	part, err = makeWritten(op, b, 0, unpackChain(w.chain))
	if err != nil {
		return 0, Written{}, err
	}
	return 2 + w.dots, part, nil
}

// MustSubdivide is like Subdivide but panics with the returned error.
func (w Written) MustSubdivide() (count int, part Written) {
	return must2(w.Subdivide())
}
