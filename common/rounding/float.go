package rounding

import (
	"math"
	"math/big"
	"strconv"
)

// Float rounds x to an integer-valued float64 using the float64 input rules of
// the package. NaN and infinities return ErrInvalidArgument. A zero result is
// always positive zero.
func (m Mode) Float(x float64) (float64, error) {
	if err := validateMode(m); err != nil {
		return 0, err
	}
	if !isFinite(x) {
		return 0, ErrInvalidArgument
	}
	return m.roundFloat(x), nil
}

// MustFloat calls Float and panics with its error on failure.
func (m Mode) MustFloat(x float64) float64 {
	v, err := m.Float(x)
	if err != nil {
		panic(err)
	}
	return v
}

// FloatInt rounds x to an integer and returns it as int. NaN and infinities
// return ErrInvalidArgument; a result outside int returns ErrOutOfRange.
func (m Mode) FloatInt(x float64) (int, error) {
	v, err := m.Float(x)
	if err != nil {
		return 0, err
	}
	if v < float64(math.MinInt) || v >= float64(1<<63) {
		return 0, ErrOutOfRange
	}
	return int(v), nil
}

// MustFloatInt calls FloatInt and panics with its error on failure.
func (m Mode) MustFloatInt(x float64) int {
	v, err := m.FloatInt(x)
	if err != nil {
		panic(err)
	}
	return v
}

// FloatMultiple rounds x to a multiple of step using exact arithmetic on the
// package's rational interpretation of both inputs. Non-finite x or step and
// step <= 0 return ErrInvalidArgument. An overflowing result returns
// ErrOutOfRange. A zero result is positive zero.
func (m Mode) FloatMultiple(x, step float64) (float64, error) {
	if err := validateMode(m); err != nil {
		return 0, err
	}
	if !isFinite(x) || !isFinite(step) || step <= 0 {
		return 0, ErrInvalidArgument
	}
	stepValue := floatValue(step)
	t := new(big.Rat).Quo(floatValue(x), stepValue)
	k := roundRat(m, t)
	v := new(big.Rat).Mul(new(big.Rat).SetInt(k), stepValue)
	return floatFromRat(v)
}

// MustFloatMultiple calls FloatMultiple and panics with its error on failure.
func (m Mode) MustFloatMultiple(x, step float64) float64 {
	v, err := m.FloatMultiple(x, step)
	if err != nil {
		panic(err)
	}
	return v
}

// FloatSteps returns the integer node index selected by rounding x to a
// multiple of step. Input failures match FloatMultiple; an index outside int
// returns ErrOutOfRange even when the corresponding float64 value is valid.
func (m Mode) FloatSteps(x, step float64) (int, error) {
	if err := validateMode(m); err != nil {
		return 0, err
	}
	if !isFinite(x) || !isFinite(step) || step <= 0 {
		return 0, ErrInvalidArgument
	}
	t := new(big.Rat).Quo(floatValue(x), floatValue(step))
	return intFromBig(roundRat(m, t))
}

// MustFloatSteps calls FloatSteps and panics with its error on failure.
func (m Mode) MustFloatSteps(x, step float64) int {
	k, err := m.FloatSteps(x, step)
	if err != nil {
		panic(err)
	}
	return k
}

// FloatPlaces rounds x to places decimal digits and returns the nearest
// float64 to that exact decimal result. Negative places round to tens,
// hundreds, and so on. Non-finite x and places outside [-MaxPlaces, MaxPlaces]
// return ErrInvalidArgument; overflow returns ErrOutOfRange. Zero is positive.
func (m Mode) FloatPlaces(x float64, places int) (float64, error) {
	if err := validateMode(m); err != nil {
		return 0, err
	}
	if !isFinite(x) || places < -MaxPlaces || places > MaxPlaces {
		return 0, ErrInvalidArgument
	}
	k := roundRat(m, scaleDecimal(floatValue(x), places))
	return floatFromRat(scaleDecimal(new(big.Rat).SetInt(k), -places))
}

// MustFloatPlaces calls FloatPlaces and panics with its error on failure.
func (m Mode) MustFloatPlaces(x float64, places int) float64 {
	v, err := m.FloatPlaces(x, places)
	if err != nil {
		panic(err)
	}
	return v
}

// FloatDecimal rounds x to places decimal digits and returns a plain decimal
// string without an exponent. Input failures match FloatPlaces; this operation
// itself cannot overflow. A zero result never has a minus sign.
func (m Mode) FloatDecimal(x float64, places int) (string, error) {
	if err := validateMode(m); err != nil {
		return "", err
	}
	if !isFinite(x) || places < -MaxPlaces || places > MaxPlaces {
		return "", ErrInvalidArgument
	}
	k := roundRat(m, scaleDecimal(floatValue(x), places))
	return formatDecimal(k, places), nil
}

// MustFloatDecimal calls FloatDecimal and panics with its error on failure.
func (m Mode) MustFloatDecimal(x float64, places int) string {
	s, err := m.FloatDecimal(x, places)
	if err != nil {
		panic(err)
	}
	return s
}

// FloatFraction rounds x to the exact grid 1/den and returns the nearest
// float64 to the selected node. Non-finite x and den <= 0 return
// ErrInvalidArgument. The operation cannot overflow and returns positive zero
// for zero.
func (m Mode) FloatFraction(x float64, den int) (float64, error) {
	if err := validateMode(m); err != nil {
		return 0, err
	}
	if !isFinite(x) || den <= 0 {
		return 0, ErrInvalidArgument
	}
	t := new(big.Rat).Mul(floatValue(x), new(big.Rat).SetInt64(int64(den)))
	k := roundRat(m, t)
	return floatFromRat(new(big.Rat).SetFrac(k, big.NewInt(int64(den))))
}

// MustFloatFraction calls FloatFraction and panics with its error on failure.
func (m Mode) MustFloatFraction(x float64, den int) float64 {
	v, err := m.FloatFraction(x, den)
	if err != nil {
		panic(err)
	}
	return v
}

// FloatNumerator returns the integer numerator k selected by rounding x to the
// exact grid 1/den. Input failures match FloatFraction; k outside int returns
// ErrOutOfRange.
func (m Mode) FloatNumerator(x float64, den int) (int, error) {
	if err := validateMode(m); err != nil {
		return 0, err
	}
	if !isFinite(x) || den <= 0 {
		return 0, ErrInvalidArgument
	}
	t := new(big.Rat).Mul(floatValue(x), new(big.Rat).SetInt64(int64(den)))
	return intFromBig(roundRat(m, t))
}

// MustFloatNumerator calls FloatNumerator and panics with its error on failure.
func (m Mode) MustFloatNumerator(x float64, den int) int {
	v, err := m.FloatNumerator(x, den)
	if err != nil {
		panic(err)
	}
	return v
}

func (m Mode) roundFloat(x float64) float64 {
	if x == 0 {
		return 0
	}
	if math.Abs(x) >= 1<<52 {
		return x
	}
	lower := math.Floor(x)
	if lower == x {
		return x
	}
	midpoint := lower + 0.5
	cmp := 0
	if x < midpoint {
		cmp = -1
	} else if x > midpoint {
		cmp = 1
	}
	side := Positive
	if lower < 0 {
		side = Negative
	}
	if m.choose(cmp, side, math.Mod(lower, 2) == 0) == Lower {
		return lower
	}
	return lower + 1
}

func floatValue(x float64) *big.Rat {
	if math.Abs(x) >= 1<<53 {
		return new(big.Rat).SetFloat64(x)
	}
	v, _ := new(big.Rat).SetString(strconv.FormatFloat(x, 'e', -1, 64))
	return v
}

func floatFromRat(v *big.Rat) (float64, error) {
	f, _ := v.Float64()
	if math.IsInf(f, 0) {
		return 0, ErrOutOfRange
	}
	if f == 0 {
		return 0, nil
	}
	return f, nil
}

func intFromBig(v *big.Int) (int, error) {
	if !v.IsInt64() {
		return 0, ErrOutOfRange
	}
	return int(v.Int64()), nil
}

func isFinite(x float64) bool {
	return !math.IsNaN(x) && !math.IsInf(x, 0)
}
