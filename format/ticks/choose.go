package ticks

import (
	"math/big"
	"strconv"

	"github.com/go-muse/musex/duration"
)

func checkValues(op string, durations []duration.Duration, positions []duration.Position) error {
	for i, d := range durations {
		if !d.IsSet() {
			return invalidError(op, "duration "+strconv.Itoa(i), duration.ErrUnset)
		}
	}
	for i, p := range positions {
		if !p.IsSet() {
			return invalidError(op, "position "+strconv.Itoa(i), duration.ErrUnset)
		}
	}
	return nil
}

// MinResolution returns the smallest resolution D0 at which every duration
// and position is a whole number of ticks: the least common multiple of the
// denominators of the reduced fractions v/w. A set is representable at D
// exactly when D is a multiple of D0. An empty set gives 1. The MIDI quarter
// notation is taken from q (the standard quarter by default). An unset element
// is ErrInvalidArgument with duration.ErrUnset; a result outside int is
// ErrOutOfRange.
func MinResolution(durations []duration.Duration, positions []duration.Position, q ...MIDIQuarter) (Resolution, error) {
	const op = "MinResolution"
	if err := checkValues(op, durations, positions); err != nil {
		return Resolution{}, err
	}
	quarter, err := quarterArgument(op, q)
	if err != nil {
		return Resolution{}, err
	}
	w := quarter.rat()
	lcm := big.NewInt(1)
	add := func(v *big.Rat) {
		den := new(big.Rat).Quo(v, w).Denom()
		gcd := new(big.Int).GCD(nil, nil, lcm, den)
		lcm.Mul(lcm, new(big.Int).Quo(den, gcd))
	}
	for _, d := range durations {
		add(ratOfDuration(d))
	}
	for _, p := range positions {
		add(ratOfPosition(p))
	}
	d, err := intOf(op, lcm)
	if err != nil {
		return Resolution{}, err
	}
	return Resolution{d: d, bb: quarter.bb}, nil
}

// MustMinResolution is like MinResolution but panics with its error.
func MustMinResolution(durations []duration.Duration, positions []duration.Position, q ...MIDIQuarter) Resolution {
	return must(MinResolution(durations, positions, q...))
}

// AtLeast returns the smallest multiple of D that is not less than n, with the
// same MIDI quarter notation. n < 1 is ErrInvalidArgument and a result outside
// int ErrOutOfRange.
func (r Resolution) AtLeast(n int) (Resolution, error) {
	const op = "Resolution.AtLeast"
	if err := r.check(op); err != nil {
		return Resolution{}, err
	}
	if n < 1 {
		return Resolution{}, invalidError(op, "lower resolution is less than 1")
	}
	k := (n-1)/r.d + 1
	product := new(big.Int).Mul(big.NewInt(int64(k)), big.NewInt(int64(r.d)))
	d, err := intOf(op, product)
	if err != nil {
		return Resolution{}, err
	}
	return Resolution{d: d, bb: r.bb}, nil
}

// MustAtLeast is like AtLeast but panics with its error.
func (r Resolution) MustAtLeast(n int) Resolution { return must(r.AtLeast(n)) }

// AtMost returns r if D <= limit and ErrFormatLimit otherwise; limit < 1 is
// ErrInvalidArgument. Typical limits are MaxSMFResolution and
// MaxMusicXMLResolution.
func (r Resolution) AtMost(limit int) (Resolution, error) {
	const op = "Resolution.AtMost"
	if err := r.check(op); err != nil {
		return Resolution{}, err
	}
	if limit < 1 {
		return Resolution{}, invalidError(op, "upper limit is less than 1")
	}
	if r.d > limit {
		return Resolution{}, newError(ErrFormatLimit, op, "resolution above the limit")
	}
	return r, nil
}

// MustAtMost is like AtMost but panics with its error.
func (r Resolution) MustAtMost(limit int) Resolution { return must(r.AtMost(limit)) }

// Represents reports whether every duration and position is a whole number of
// ticks.
func (r Resolution) Represents(durations []duration.Duration, positions []duration.Position) (bool, error) {
	ds, ps, err := r.unrepresentable("Resolution.Represents", durations, positions)
	if err != nil {
		return false, err
	}
	return ds == nil && ps == nil, nil
}

// MustRepresents is like Represents but panics with its error.
func (r Resolution) MustRepresents(durations []duration.Duration, positions []duration.Position) bool {
	return must(r.Represents(durations, positions))
}

// Unrepresentable returns the durations and positions that are not whole
// numbers of ticks, in input order with repetitions; an empty list is nil.
func (r Resolution) Unrepresentable(durations []duration.Duration, positions []duration.Position) ([]duration.Duration, []duration.Position, error) {
	return r.unrepresentable("Resolution.Unrepresentable", durations, positions)
}

// MustUnrepresentable is like Unrepresentable but panics with its error.
func (r Resolution) MustUnrepresentable(durations []duration.Duration, positions []duration.Position) ([]duration.Duration, []duration.Position) {
	return must2(r.Unrepresentable(durations, positions))
}

func (r Resolution) unrepresentable(op string, durations []duration.Duration, positions []duration.Position) ([]duration.Duration, []duration.Position, error) {
	if err := r.check(op); err != nil {
		return nil, nil, err
	}
	if err := checkValues(op, durations, positions); err != nil {
		return nil, nil, err
	}
	var ds []duration.Duration
	var ps []duration.Position
	for _, d := range durations {
		if !r.unitsOf(ratOfDuration(d)).IsInt() {
			ds = append(ds, d)
		}
	}
	for _, p := range positions {
		if !r.unitsOf(ratOfPosition(p)).IsInt() {
			ps = append(ps, p)
		}
	}
	return ds, ps, nil
}
