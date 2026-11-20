package ticks

import (
	"errors"
	"math/big"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempomap"
)

// This file selects integers from exact irrational values of a musex tempo
// map. A target that is linear in one exact value of tempomap is rounded by
// the step methods of tempomap, which are correctly rounded without a
// precision limit. A target that is a fractional-linear function of one exact
// value is decided by exact comparisons, which are also step methods. Only the
// product of two irrational values has no proven bound; it is decided by
// intervals of increasing precision and may fail with ErrOutOfRange.

// placesSchedule is the fixed sequence of decimal places requested from
// tempomap, in order.
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

// span is an interval of positions of a tempo map without holds, with its
// exact seconds available through tempomap.
type span struct {
	m        tempomap.Map
	from, to tempomap.Address
}

func secondsStep(op string, step *big.Rat) (tempo.Seconds, error) {
	f, err := fraction.FromBigRat(step)
	if err != nil {
		return tempo.Seconds{}, rangeError(op, "step outside common/fraction")
	}
	s, err := tempo.SecondsFromFraction(f)
	if err != nil {
		return tempo.Seconds{}, dependencyError(op, err)
	}
	return s, nil
}

// steps returns k = mode(S/step). An index outside int is reported by the
// error of tempomap in overflow, not as a failure.
func (s span) steps(op string, step *big.Rat, mode rounding.Mode) (k int, overflow, err error) {
	h, err := secondsStep(op, step)
	if err != nil {
		return 0, nil, err
	}
	k, err = s.m.BetweenSteps(s.from, s.to, h, mode)
	if err != nil {
		if errors.Is(err, tempomap.ErrOutOfRange) && errors.Is(err, fraction.ErrNotRepresentable) {
			return 0, err, nil
		}
		return 0, nil, dependencyError(op, err)
	}
	return k, nil, nil
}

// compare returns the sign of S - y for the exact seconds S of the span and a
// rational y > 0.
func (s span) compare(op string, y *big.Rat) (int, error) {
	d, overflow, err := s.steps(op, y, rounding.Down)
	if err != nil {
		return 0, err
	}
	if overflow != nil {
		return 1, nil
	}
	if d == 0 {
		return -1, nil
	}
	u, overflow, err := s.steps(op, y, rounding.Up)
	if err != nil {
		return 0, err
	}
	if overflow != nil || u >= 2 {
		return 1, nil
	}
	return 0, nil
}

// interval returns the seconds of the span rounded to places down and up, and
// whether they are exact.
func (s span) interval(op string, places int) (lo, hi *big.Rat, err error) {
	down, err := s.m.Between(s.from, s.to, places, rounding.Down)
	if err != nil {
		return nil, nil, dependencyError(op, err)
	}
	up, err := s.m.Between(s.from, s.to, places, rounding.Up)
	if err != nil {
		return nil, nil, dependencyError(op, err)
	}
	return ratOfSeconds(down.MustValue()), ratOfSeconds(up.MustValue()), nil
}

// halfUpInterval returns the interval of the exact seconds from the result of
// Between with places and HalfUp, and whether it is exact.
func (s span) halfUpInterval(places int) (lo, hi *big.Rat, exact bool, err error) {
	result, err := s.m.Between(s.from, s.to, places, rounding.HalfUp)
	if err != nil {
		return nil, nil, false, err
	}
	value := ratOfSeconds(result.MustValue())
	if !result.MustApproximate() {
		return value, new(big.Rat).Set(value), true, nil
	}
	half := new(big.Rat).SetFrac(big.NewInt(1), new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(places)), nil))
	half.Quo(half, big.NewRat(2, 1))
	return new(big.Rat).Sub(value, half), new(big.Rat).Add(value, half), false, nil
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
func stretchedTempo(op string, s span, m, e *big.Rat, mode rounding.Mode) (MIDITempo, error) {
	limitError := func() error { return newError(ErrFormatLimit, op, "tempo outside 1..16777215 microseconds") }
	upper := tempoLimit()
	if m.Cmp(upper) >= 0 {
		return MIDITempo{}, limitError()
	}
	f := func(seconds *big.Rat) *big.Rat {
		v := new(big.Rat).Mul(m, e)
		v.Quo(v, seconds)
		return v.Add(v, m)
	}
	y0 := new(big.Rat).Mul(m, e)
	y0.Quo(y0, new(big.Rat).Sub(upper, m))
	sign, err := s.compare(op, y0)
	if err != nil {
		return MIDITempo{}, err
	}
	if sign <= 0 {
		return MIDITempo{}, limitError()
	}
	lo, hi, err := s.interval(op, placesSchedule[0])
	if err != nil {
		return MIDITempo{}, err
	}
	if lo.Cmp(y0) < 0 {
		lo = y0
	}
	nLo, err := mode.RatInt(f(hi))
	if err != nil {
		panic(err)
	}
	nHi, err := mode.RatInt(f(lo))
	if err != nil {
		panic(err)
	}
	// Binary search for the first boundary that f(S) does not pass.
	low, high := new(big.Int).Set(nLo), new(big.Int).Set(nHi)
	for low.Cmp(high) < 0 {
		mid := new(big.Int).Add(low, high)
		mid.Rsh(mid, 1)
		b, tieUp := boundary(mode, mid)
		y := new(big.Rat).Mul(m, e)
		y.Quo(y, new(big.Rat).Sub(b, m))
		sign, err := s.compare(op, y)
		if err != nil {
			return MIDITempo{}, err
		}
		if sign < 0 || (sign == 0 && tieUp) {
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

// productTempo rounds c*T*(S+E)/S once with mode for the exact seconds T and S
// of two spans, both irrational.
func productTempo(op string, part, stretch span, c, e *big.Rat, mode rounding.Mode) (MIDITempo, error) {
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
