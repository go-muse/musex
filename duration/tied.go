package duration

import (
	"cmp"
	"encoding/binary"
	"math/big"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/tuplet"
)

// Tied is a tied duration: two or more written durations under a tie, in
// order. Each part has its own chain of tuplets, and the duration is the sum
// of the durations of the parts: a half tied to a quarter lasts 3/4.
//
// A single sign is a Written, not a Tied. Equal compares the parts by
// notation and matches ==; EqualDuration compares durations. The zero value
// Tied{} is the unset tied duration. Tied is comparable with == and can be
// used as a map key.
type Tied struct {
	n     int    // number of parts
	parts string // packed parts, see packParts
	value fraction.Fraction
}

// partHeaderBytes is the size of the header of a packed part: the base, the
// number of dots, and the number of tuplets.
const partHeaderBytes = 1 + 8 + 8

// appendPart appends the canonical immutable packing of a set written
// duration to b: the base, the number of dots, the number of tuplets, and the
// packed chain.
func appendPart(b []byte, w Written) []byte {
	b = append(b, byte(w.base))
	b = binary.BigEndian.AppendUint64(b, uint64(w.dots))
	b = binary.BigEndian.AppendUint64(b, uint64(len(w.chain)/tupletBytes))
	return append(b, w.chain...)
}

// unpackParts returns a new slice with the parts of a packed tied duration.
func unpackParts(s string, n int) []Written {
	parts := make([]Written, 0, n)
	for len(s) > 0 {
		base := Base(s[0])
		dots := int(uint64At(s[1:9]))
		size := int(uint64At(s[9:17])) * tupletBytes
		chainText := s[partHeaderBytes : partHeaderBytes+size]
		s = s[partHeaderBytes+size:]
		chain := unpackChain(chainText)
		coef, _ := chainCoefficient("Tied.Parts", chain)
		v, _ := writtenValue("Tied.Parts", base, dots, coef)
		parts = append(parts, Written{base: base, dots: dots, chain: chainText, value: v})
	}
	return parts
}

// makeTied returns the tied duration of set parts with the exact sum of
// their durations.
func makeTied(op string, parts []Written) (Tied, error) {
	sum := new(big.Rat)
	size := 0
	for _, w := range parts {
		sum.Add(sum, ratOf(w.value))
		size += partHeaderBytes + len(w.chain)
	}
	v, ok := fromRat(sum)
	if !ok {
		return Tied{}, rangeError(op)
	}
	b := make([]byte, 0, size)
	for _, w := range parts {
		b = appendPart(b, w)
	}
	return Tied{n: len(parts), parts: string(b), value: v}, nil
}

// NewTied returns the tied duration of parts in order. In order of checking,
// it returns an error wrapping ErrInvalidArgument for fewer than two parts,
// ErrUnset for the first unset part, and ErrOutOfRange if the sum of the
// durations is outside the range of package common/fraction. The list is only
// read.
func NewTied(parts ...Written) (Tied, error) {
	const op = "NewTied"
	if len(parts) < 2 {
		return Tied{}, invalidError(op, "fewer than two parts")
	}
	for _, w := range parts {
		if !w.IsSet() {
			return Tied{}, unsetError(op, "part")
		}
	}
	return makeTied(op, parts)
}

// MustNewTied is like NewTied but panics with the returned error.
func MustNewTied(parts ...Written) Tied {
	return must(NewTied(parts...))
}

// IsSet reports whether t is a set tied duration.
func (t Tied) IsSet() bool {
	return t.n != 0
}

// unset returns the error for an unset receiver.
func (t Tied) unset(op string) error {
	return unsetError(op, "receiver")
}

// Len returns the number of parts of t. It returns an error wrapping ErrUnset
// if t is unset.
func (t Tied) Len() (int, error) {
	if !t.IsSet() {
		return 0, t.unset("Tied.Len")
	}
	return t.n, nil
}

// MustLen is like Len but panics with the returned error.
func (t Tied) MustLen() int {
	return must(t.Len())
}

// Parts returns a new slice with the parts of t in order. The slice belongs
// to the caller. It returns an error wrapping ErrUnset if t is unset.
func (t Tied) Parts() ([]Written, error) {
	if !t.IsSet() {
		return nil, t.unset("Tied.Parts")
	}
	return unpackParts(t.parts, t.n), nil
}

// MustParts is like Parts but panics with the returned error.
func (t Tied) MustParts() []Written {
	return must(t.Parts())
}

// Duration returns the duration of t, the sum of the durations of its parts.
// It returns an error wrapping ErrUnset if t is unset.
func (t Tied) Duration() (Duration, error) {
	if !t.IsSet() {
		return Duration{}, t.unset("Tied.Duration")
	}
	return Duration{v: t.value}, nil
}

// MustDuration is like Duration but panics with the returned error.
func (t Tied) MustDuration() Duration {
	return must(t.Duration())
}

// Equal reports whether t and other have the same parts in the same order,
// compared by notation as Written.Equal does. It matches ==; the unset tied
// duration is equal only to the unset tied duration.
func (t Tied) Equal(other Tied) bool {
	return t == other
}

// Compare is a total order for sorting that returns −1, 0, or +1: by
// duration; for equal durations by the number of parts (fewer first), and
// then by the parts in order, as Written.Compare does. The unset tied
// duration precedes every set one. Compare returns 0 if and only if
// t == other.
func (t Tied) Compare(other Tied) int {
	if c := t.value.Compare(other.value); c != 0 || !t.IsSet() {
		return c
	}
	if c := cmp.Compare(t.n, other.n); c != 0 {
		return c
	}
	if t.parts == other.parts {
		return 0
	}
	a, b := unpackParts(t.parts, t.n), unpackParts(other.parts, other.n)
	for i := range a {
		if c := a[i].Compare(b[i]); c != 0 {
			return c
		}
	}
	return 0
}

// durations checks that t and other are set.
func (t Tied) durations(op string, other Tied) error {
	if !t.IsSet() {
		return t.unset(op)
	}
	if !other.IsSet() {
		return unsetError(op, "argument")
	}
	return nil
}

// EqualDuration reports whether t and other have equal durations. It returns
// an error wrapping ErrUnset if t or other is unset.
func (t Tied) EqualDuration(other Tied) (bool, error) {
	if err := t.durations("Tied.EqualDuration", other); err != nil {
		return false, err
	}
	return t.value == other.value, nil
}

// MustEqualDuration is like EqualDuration but panics with the returned error.
func (t Tied) MustEqualDuration(other Tied) bool {
	return must(t.EqualDuration(other))
}

// CompareDuration compares the durations of t and other and returns −1, 0, or
// +1. It returns an error wrapping ErrUnset if t or other is unset.
func (t Tied) CompareDuration(other Tied) (int, error) {
	if err := t.durations("Tied.CompareDuration", other); err != nil {
		return 0, err
	}
	return t.value.Compare(other.value), nil
}

// MustCompareDuration is like CompareDuration but panics with the returned
// error.
func (t Tied) MustCompareDuration(other Tied) int {
	return must(t.CompareDuration(other))
}

// joined returns the tied duration with the packed parts of t followed by
// extra parts with the given count and total duration.
func (t Tied) joined(op string, n int, parts string, value fraction.Fraction) (Tied, error) {
	sum := ratOf(t.value)
	v, ok := fromRat(sum.Add(sum, ratOf(value)))
	if !ok {
		return Tied{}, rangeError(op)
	}
	return Tied{n: t.n + n, parts: t.parts + parts, value: v}, nil
}

// Append returns a new tied duration with w added after the parts of t. It
// returns an error wrapping ErrUnset if t or w is unset and ErrOutOfRange if
// the new duration is outside the range of package common/fraction.
func (t Tied) Append(w Written) (Tied, error) {
	const op = "Tied.Append"
	if !t.IsSet() {
		return Tied{}, t.unset(op)
	}
	if !w.IsSet() {
		return Tied{}, unsetError(op, "argument")
	}
	return t.joined(op, 1, string(appendPart(nil, w)), w.value)
}

// MustAppend is like Append but panics with the returned error.
func (t Tied) MustAppend(w Written) Tied {
	return must(t.Append(w))
}

// Join returns a new tied duration with the parts of t followed by the parts
// of other. It returns an error wrapping ErrUnset if t or other is unset and
// ErrOutOfRange if the new duration is outside the range of package
// common/fraction.
func (t Tied) Join(other Tied) (Tied, error) {
	const op = "Tied.Join"
	if err := t.durations(op, other); err != nil {
		return Tied{}, err
	}
	return t.joined(op, other.n, other.parts, other.value)
}

// MustJoin is like Join but panics with the returned error.
func (t Tied) MustJoin(other Tied) Tied {
	return must(t.Join(other))
}

// ToWritten returns the single sign with the duration of t under chain, as
// Duration.ToWritten does: a half tied to a quarter gives a dotted half. In
// order of checking, it returns an error wrapping ErrUnset if t is unset,
// ErrInvalidArgument together with tuplet.ErrUnset if a tuplet is unset,
// ErrOutOfRange if the chain coefficient is outside the range of package
// common/fraction, and ErrNotNotatable if the duration is not one sign under
// the chain. The chain is only read.
func (t Tied) ToWritten(chain ...tuplet.Tuplet) (Written, error) {
	const op = "Tied.ToWritten"
	if !t.IsSet() {
		return Written{}, t.unset(op)
	}
	return Duration{v: t.value}.toWritten(op, chain)
}

// MustToWritten is like ToWritten but panics with the returned error.
func (t Tied) MustToWritten(chain ...tuplet.Tuplet) Written {
	return must(t.ToWritten(chain...))
}
