package ticks

import (
	"errors"
	"math/big"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/internal/timecalc"
	"github.com/go-muse/musex/tempomap"
)

// This file selects integers from exact time forms supplied by the shared
// kernel. Linear targets use its correctly rounded step operation; a
// fractional-linear target uses exact comparisons. Only the combination of
// two irrational forms uses bounded interval refinement and may fail with
// ErrOutOfRange when none of the scheduled precisions proves the rounding.

// placesSchedule is the fixed sequence of decimal places requested from
// the shared time kernel, in order.
var placesSchedule = [...]int{20, 40, 60, 76}

// evalFunc returns, for one number of decimal places, a closed interval
// [lo, hi] that contains the exact target value. exact reports lo == hi ==
// the exact value. hi == nil reports an interval unbounded above: the
// precision does not decide.
type evalFunc func(places int) (lo, hi *big.Rat, exact bool, err error)

// roundCertified returns the integer that mode selects from the exact target
// value, if it is at most limit.
func roundCertified(eval evalFunc, mode rounding.Mode, limit *big.Int) (*big.Int, error) {
	const op = "ticks.roundCertified"
	for _, places := range placesSchedule {
		lo, hi, exact, err := eval(places)
		if err != nil {
			if errors.Is(err, tempomap.ErrOutOfRange) {
				return nil, newError(ErrOutOfRange, op, "precision exhausted", err, fraction.ErrOutOfRange)
			}
			return nil, err
		}
		lower, err := mode.RatInt(lo)
		if err != nil {
			panic(err)
		}
		decided := exact
		if !decided && hi != nil {
			upper, err := mode.RatInt(hi)
			if err != nil {
				panic(err)
			}
			decided = lower.Cmp(upper) == 0
		}
		if !decided {
			if lower.Cmp(limit) > 0 {
				return nil, newError(ErrFormatLimit, op, "result above the limit")
			}
			continue
		}
		if lower.Cmp(limit) > 0 {
			return nil, newError(ErrFormatLimit, op, "result above the limit")
		}
		return lower, nil
	}
	return nil, rangeError(op, "rounding not decided within 76 decimal places")
}

// span retains the exact time form of an original source step. Its computed
// endpoints remain unrestricted rationals throughout the export calculation.
type span struct{ form timecalc.LogForm }

func (s span) steps(_ string, step *big.Rat, mode rounding.Mode) (k int, overflow, err error) {
	n := timecalc.RoundLogSteps(s.form, step, mode)
	if !n.IsInt64() {
		return 0, errors.Join(tempomap.ErrOutOfRange, fraction.ErrNotRepresentable), nil
	}
	return int(n.Int64()), nil, nil
}

func (s span) halfUpInterval(places int) (lo, hi *big.Rat, exact bool, err error) {
	lo, hi, exact = timecalc.DecimalInterval(s.form, places)
	return lo, hi, exact, nil
}

// tempoLimit returns M + 1 for the Set Tempo limit M.
func tempoLimit() *big.Rat { return big.NewRat(MaxMicrosecondsPerQuarter+1, 1) }

// linearTempo rounds c*S once with mode for the exact seconds S of the span
// and a rational c > 0. An index outside int means a tempo above the limit.
func linearTempo(op string, s span, c *big.Rat, mode rounding.Mode) (MIDITempo, error) {
	k, overflow, err := s.steps(op, new(big.Rat).Inv(c), mode)
	if err != nil {
		return MIDITempo{}, err
	}
	if overflow != nil {
		return MIDITempo{}, newError(ErrFormatLimit, op, "tempo outside 1..16777215 microseconds", overflow)
	}
	if k <= 0 || k > MaxMicrosecondsPerQuarter {
		return MIDITempo{}, newError(ErrFormatLimit, op, "tempo outside 1..16777215 microseconds")
	}
	return MIDITempo{us: int32(k)}, nil
}

// boundary returns the value b between the results j and j+1 of mode and
// whether the exact value b itself rounds to j+1.
func boundary(mode rounding.Mode, j *big.Int) (*big.Rat, bool) {
	jr := new(big.Rat).SetInt(j)
	switch mode {
	case rounding.Down, rounding.TowardZero:
		return jr.Add(jr, big.NewRat(1, 1)), true
	case rounding.Up, rounding.AwayFromZero:
		return jr, false
	case rounding.HalfUp, rounding.HalfAwayFromZero:
		return jr.Add(jr, big.NewRat(1, 2)), true
	case rounding.HalfDown, rounding.HalfTowardZero:
		return jr.Add(jr, big.NewRat(1, 2)), false
	default: // HalfEven: the tie goes to the even neighbor.
		return jr.Add(jr, big.NewRat(1, 2)), j.Bit(0) == 1
	}
}

// stretchedTempo rounds f(S) = m + m*E/S once with mode for the exact seconds
// S of the span, a rational tempo m > 0 and a hold E > 0.
func stretchedTempo(op string, s timeSum, m, e *big.Rat, mode rounding.Mode) (MIDITempo, error) {
	limitError := func() error { return newError(ErrFormatLimit, op, "tempo outside 1..16777215 microseconds") }
	// Compare m*(S+E)/S with a rational boundary. Only one logarithmic
	// form is involved, so every decision uses the exact weighted kernel.
	compare := func(b *big.Rat) (int, error) {
		if b.Cmp(m) <= 0 {
			return 1, nil
		}
		y := new(big.Rat).Quo(new(big.Rat).Mul(m, e), new(big.Rat).Sub(b, m))
		sign, err := s.compare(op, y)
		return -sign, err
	}
	sign, err := compare(tempoLimit())
	if err != nil {
		return MIDITempo{}, err
	}
	if sign >= 0 {
		return MIDITempo{}, limitError()
	}
	low, high := big.NewInt(0), big.NewInt(MaxMicrosecondsPerQuarter+1)
	for low.Cmp(high) < 0 {
		mid := new(big.Int).Rsh(new(big.Int).Add(low, high), 1)
		b, tieUp := boundary(mode, mid)
		sign, err := compare(b)
		if err != nil {
			return MIDITempo{}, err
		}
		if sign > 0 || (sign == 0 && tieUp) {
			low.Add(mid, big.NewInt(1))
		} else {
			high.Set(mid)
		}
	}
	if low.Sign() <= 0 || low.Cmp(big.NewInt(MaxMicrosecondsPerQuarter)) > 0 {
		return MIDITempo{}, limitError()
	}
	return MIDITempo{us: int32(low.Int64())}, nil
}

// productTempo rounds c*T*(S+E)/S once for the exact source-step time T
// and weighted staircase time S, both irrational.
func productTempo(op string, part span, stretch timeSum, c, e *big.Rat, mode rounding.Mode) (MIDITempo, error) {
	eval := func(places int) (lo, hi *big.Rat, exact bool, err error) {
		tLo, tHi, tExact, err := part.halfUpInterval(places)
		if err != nil {
			return nil, nil, false, err
		}
		sLo, sHi, sExact, err := stretch.halfUpInterval(places)
		if err != nil {
			return nil, nil, false, err
		}
		if tLo.Sign() < 0 {
			tLo = new(big.Rat)
		}
		g := func(t, s *big.Rat) *big.Rat {
			v := new(big.Rat).Add(s, e)
			v.Quo(v, s)
			v.Mul(v, t)
			return v.Mul(v, c)
		}
		lo = g(tLo, sHi)
		if sLo.Sign() <= 0 {
			return lo, nil, false, nil
		}
		hi = g(tHi, sLo)
		return lo, hi, tExact && sExact, nil
	}
	n, err := roundCertified(eval, mode, big.NewInt(MaxMicrosecondsPerQuarter))
	if err != nil {
		if errors.Is(err, ErrOutOfRange) || errors.Is(err, ErrFormatLimit) {
			return MIDITempo{}, err
		}
		return MIDITempo{}, dependencyError(op, err)
	}
	if n.Sign() == 0 {
		return MIDITempo{}, newError(ErrFormatLimit, op, "tempo rounded to zero")
	}
	return MIDITempo{us: int32(n.Int64())}, nil
}

// weightedSpan is a positive contribution to the base time on the final tick grid.
type weightedSpan struct {
	source span
	weight *big.Rat
}
type timeSum struct {
	parts    []weightedSpan
	exact    *big.Rat
	constant *big.Rat // rational contribution when parts contain irrational times
}

func (s timeSum) compare(_ string, y *big.Rat) (int, error) {
	if s.exact != nil {
		return s.exact.Cmp(y), nil
	}
	form := timecalc.Constant(s.constant)
	for _, p := range s.parts {
		form = timecalc.Add(form, timecalc.Scale(p.source.form, p.weight))
	}
	return timecalc.CompareLog(form, y), nil
}

func (s timeSum) halfUpInterval(places int) (lo, hi *big.Rat, exact bool, err error) {
	if s.exact != nil {
		return new(big.Rat).Set(s.exact), new(big.Rat).Set(s.exact), true, nil
	}
	lo, hi = new(big.Rat).Set(s.constant), new(big.Rat).Set(s.constant)
	exact = true
	for _, p := range s.parts {
		l, h, x, err := p.source.halfUpInterval(places)
		if err != nil {
			return nil, nil, false, err
		}
		if l.Sign() < 0 {
			l = new(big.Rat)
		}
		lo.Add(lo, new(big.Rat).Mul(p.weight, l))
		hi.Add(hi, new(big.Rat).Mul(p.weight, h))
		exact = exact && x
	}
	return lo, hi, exact, nil
}
