package tempo

import (
	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tempo/term"
)

// Kind is the kind of the metronome marking of an indication. Its zero value
// is NoMark, a valid value that means that the indication has no marking; it
// is not an unset state. Values greater than RangeMark are not kinds: IsValid
// reports false for them.
type Kind uint8

const (
	// NoMark means that the indication has no metronome marking. It is a
	// valid value.
	NoMark Kind = iota
	// ExactMark is an exact marking, a Tempo.
	ExactMark
	// ApproxMark is an approximate marking, an Approx.
	ApproxMark
	// RangeMark is a range, a Range.
	RangeMark
)

// kindNames holds the strings of the valid kinds, indexed by kind.
var kindNames = [...]string{"none", "exact", "approx", "range"}

// IsValid reports whether k is NoMark, ExactMark, ApproxMark, or RangeMark.
func (k Kind) IsValid() bool {
	return k <= RangeMark
}

// Equal reports whether k and other are the same value; it matches ==.
func (k Kind) Equal(other Kind) bool {
	return k == other
}

// String returns "none", "exact", "approx", or "range", and "Kind(invalid)"
// for a value that is not a kind.
func (k Kind) String() string {
	if !k.IsValid() {
		return "Kind(invalid)"
	}
	return kindNames[k]
}

// Indication is a tempo indication as written above a score: a term of a list,
// a metronome marking of any kind, or both, such as "Allegro, ♩ = 120". The
// marking sets the speed and the term the character; whether the number fits
// the guide of the term is not checked, so "Allegro, ♩ = 60" is a valid
// indication. Seconds are computed only from an exact marking, which Tempo
// reads.
//
// The zero value Indication{} is the unset indication; a set indication has a
// term, a marking, or both. Its String is the empty string, which no set
// indication prints. Indication is comparable with ==, which matches Equal,
// and can be used as a map key.
type Indication struct {
	term  term.Term
	kind  Kind
	unit  duration.Written  // unit of the marking; Written{} without one
	lower fraction.Fraction // number of an exact or approximate marking, or lower number of a range
	upper fraction.Fraction // upper number of a range; Fraction{} otherwise
}

// withMark returns i with the marking replaced by the marking of kind k.
func (i Indication) withMark(k Kind, unit duration.Written, lower, upper fraction.Fraction) Indication {
	i.kind, i.unit, i.lower, i.upper = k, unit, lower, upper
	return i
}

// IndicationOf returns the indication of the term t alone, without a marking.
// It returns an error wrapping ErrInvalidArgument if t is unset.
func IndicationOf(t term.Term) (Indication, error) {
	if !t.IsSet() {
		return Indication{}, invalidError("IndicationOf", "unset term")
	}
	return Indication{term: t}, nil
}

// MustIndicationOf is like IndicationOf but panics with the returned error.
func MustIndicationOf(t term.Term) Indication {
	return must(IndicationOf(t))
}

// IndicationOfTempo returns the indication of the exact marking m alone,
// without a term. It returns an error wrapping ErrUnset if m is unset.
func IndicationOfTempo(m Tempo) (Indication, error) {
	return Indication{}.setTempo("IndicationOfTempo", m)
}

// MustIndicationOfTempo is like IndicationOfTempo but panics with the returned
// error.
func MustIndicationOfTempo(m Tempo) Indication {
	return must(IndicationOfTempo(m))
}

// IndicationOfApprox returns the indication of the approximate marking a
// alone, without a term. It returns an error wrapping ErrUnset if a is unset.
func IndicationOfApprox(a Approx) (Indication, error) {
	return Indication{}.setApprox("IndicationOfApprox", a)
}

// MustIndicationOfApprox is like IndicationOfApprox but panics with the
// returned error.
func MustIndicationOfApprox(a Approx) Indication {
	return must(IndicationOfApprox(a))
}

// IndicationOfRange returns the indication of the range r alone, without a
// term. It returns an error wrapping ErrUnset if r is unset.
func IndicationOfRange(r Range) (Indication, error) {
	return Indication{}.setRange("IndicationOfRange", r)
}

// MustIndicationOfRange is like IndicationOfRange but panics with the returned
// error.
func MustIndicationOfRange(r Range) Indication {
	return must(IndicationOfRange(r))
}

// setTempo returns i with the exact marking m.
func (i Indication) setTempo(op string, m Tempo) (Indication, error) {
	if !m.IsSet() {
		return Indication{}, unsetError(op, "marking")
	}
	return i.withMark(ExactMark, m.m.unit, m.m.bpm, fraction.Fraction{}), nil
}

// setApprox returns i with the approximate marking a.
func (i Indication) setApprox(op string, a Approx) (Indication, error) {
	if !a.IsSet() {
		return Indication{}, unsetError(op, "marking")
	}
	return i.withMark(ApproxMark, a.m.unit, a.m.bpm, fraction.Fraction{}), nil
}

// setRange returns i with the range r.
func (i Indication) setRange(op string, r Range) (Indication, error) {
	if !r.IsSet() {
		return Indication{}, unsetError(op, "marking")
	}
	return i.withMark(RangeMark, r.unit, r.lower, r.upper), nil
}

// IsSet reports whether i is set, that is, not the unset Indication{}.
func (i Indication) IsSet() bool {
	return i.term.IsSet() || i.kind != NoMark
}

// unset returns an error wrapping ErrUnset for the receiver of op.
func (Indication) unset(op string) error {
	return unsetError(op, "receiver")
}

// Term returns the term of i and true, or Term{} and false without an error if
// i has no term. It returns an error wrapping ErrUnset if i is unset.
func (i Indication) Term() (term.Term, bool, error) {
	if !i.IsSet() {
		return term.Term{}, false, i.unset("Indication.Term")
	}
	return i.term, i.term.IsSet(), nil
}

// MustTerm is like Term but panics with the returned error.
func (i Indication) MustTerm() (term.Term, bool) {
	return must2(i.Term())
}

// Kind returns the kind of the marking of i, NoMark if it has none. It returns
// an error wrapping ErrUnset if i is unset.
func (i Indication) Kind() (Kind, error) {
	if !i.IsSet() {
		return NoMark, i.unset("Indication.Kind")
	}
	return i.kind, nil
}

// MustKind is like Kind but panics with the returned error.
func (i Indication) MustKind() Kind {
	return must(i.Kind())
}

// Tempo returns the exact marking of i and true, or Tempo{} and false without
// an error if i has no exact marking: "Moderato, ♩ ≈ 108" has none. It
// returns an error wrapping ErrUnset if i is unset.
func (i Indication) Tempo() (Tempo, bool, error) {
	if !i.IsSet() {
		return Tempo{}, false, i.unset("Indication.Tempo")
	}
	if i.kind != ExactMark {
		return Tempo{}, false, nil
	}
	return Tempo{m: mark{unit: i.unit, bpm: i.lower}}, true, nil
}

// MustTempo is like Tempo but panics with the returned error.
func (i Indication) MustTempo() (Tempo, bool) {
	return must2(i.Tempo())
}

// Approx returns the approximate marking of i and true, or Approx{} and false
// without an error if i has none. It returns an error wrapping ErrUnset if i is
// unset.
func (i Indication) Approx() (Approx, bool, error) {
	if !i.IsSet() {
		return Approx{}, false, i.unset("Indication.Approx")
	}
	if i.kind != ApproxMark {
		return Approx{}, false, nil
	}
	return Approx{m: mark{unit: i.unit, bpm: i.lower}}, true, nil
}

// MustApprox is like Approx but panics with the returned error.
func (i Indication) MustApprox() (Approx, bool) {
	return must2(i.Approx())
}

// Range returns the range of i and true, or Range{} and false without an
// error if i has none. It returns an error wrapping ErrUnset if i is unset.
func (i Indication) Range() (Range, bool, error) {
	if !i.IsSet() {
		return Range{}, false, i.unset("Indication.Range")
	}
	if i.kind != RangeMark {
		return Range{}, false, nil
	}
	return Range{unit: i.unit, lower: i.lower, upper: i.upper}, true, nil
}

// MustRange is like Range but panics with the returned error.
func (i Indication) MustRange() (Range, bool) {
	return must2(i.Range())
}

// WithTerm returns i with the term t, replacing or adding it. In order of
// checking, it returns an error wrapping ErrUnset if i is unset and
// ErrInvalidArgument if t is unset.
func (i Indication) WithTerm(t term.Term) (Indication, error) {
	const op = "Indication.WithTerm"
	if !i.IsSet() {
		return Indication{}, i.unset(op)
	}
	if !t.IsSet() {
		return Indication{}, invalidError(op, "unset term")
	}
	i.term = t
	return i, nil
}

// MustWithTerm is like WithTerm but panics with the returned error.
func (i Indication) MustWithTerm(t term.Term) Indication {
	return must(i.WithTerm(t))
}

// WithTempo returns i with the exact marking m, replacing a marking of any
// kind or adding it. It returns an error wrapping ErrUnset if i or then m is
// unset.
func (i Indication) WithTempo(m Tempo) (Indication, error) {
	const op = "Indication.WithTempo"
	if !i.IsSet() {
		return Indication{}, i.unset(op)
	}
	return i.setTempo(op, m)
}

// MustWithTempo is like WithTempo but panics with the returned error.
func (i Indication) MustWithTempo(m Tempo) Indication {
	return must(i.WithTempo(m))
}

// WithApprox returns i with the approximate marking a, replacing a marking of
// any kind or adding it. It returns an error wrapping ErrUnset if i or then a
// is unset.
func (i Indication) WithApprox(a Approx) (Indication, error) {
	const op = "Indication.WithApprox"
	if !i.IsSet() {
		return Indication{}, i.unset(op)
	}
	return i.setApprox(op, a)
}

// MustWithApprox is like WithApprox but panics with the returned error.
func (i Indication) MustWithApprox(a Approx) Indication {
	return must(i.WithApprox(a))
}

// WithRange returns i with the range r, replacing a marking of any kind or
// adding it. It returns an error wrapping ErrUnset if i or then r is unset.
func (i Indication) WithRange(r Range) (Indication, error) {
	const op = "Indication.WithRange"
	if !i.IsSet() {
		return Indication{}, i.unset(op)
	}
	return i.setRange(op, r)
}

// MustWithRange is like WithRange but panics with the returned error.
func (i Indication) MustWithRange(r Range) Indication {
	return must(i.WithRange(r))
}

// WithoutTerm returns i without its term. In order of checking, it returns an
// error wrapping ErrUnset if i is unset and ErrInvalidArgument if i has no
// marking, so that nothing would remain.
func (i Indication) WithoutTerm() (Indication, error) {
	const op = "Indication.WithoutTerm"
	if !i.IsSet() {
		return Indication{}, i.unset(op)
	}
	if i.kind == NoMark {
		return Indication{}, invalidError(op, "indication without a marking")
	}
	i.term = term.Term{}
	return i, nil
}

// MustWithoutTerm is like WithoutTerm but panics with the returned error.
func (i Indication) MustWithoutTerm() Indication {
	return must(i.WithoutTerm())
}

// WithoutMarking returns i without its marking. In order of checking, it
// returns an error wrapping ErrUnset if i is unset and ErrInvalidArgument if i
// has no term, so that nothing would remain.
func (i Indication) WithoutMarking() (Indication, error) {
	const op = "Indication.WithoutMarking"
	if !i.IsSet() {
		return Indication{}, i.unset(op)
	}
	if !i.term.IsSet() {
		return Indication{}, invalidError(op, "indication without a term")
	}
	return i.withMark(NoMark, duration.Written{}, fraction.Fraction{}, fraction.Fraction{}), nil
}

// MustWithoutMarking is like WithoutMarking but panics with the returned error.
func (i Indication) MustWithoutMarking() Indication {
	return must(i.WithoutMarking())
}

// Equal reports whether i and other are the same record: equal terms, as
// term.Term.Equal compares them, or both without a term, and equal markings
// of the same kind, or both without a marking. Two unset indications are
// equal. Equal matches ==. Indications have no order.
func (i Indication) Equal(other Indication) bool {
	return i == other
}
