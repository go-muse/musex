package template

import (
	"math"
	"math/big"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/spelling"
)

// coordinates are the exact linear coordinates of an interval: a signed letter
// distance t and a signed rational chromatic distance v. Every operation
// returns newly owned values; no coordinate shares mutable numbers.
type coordinates struct {
	t *big.Int
	v *big.Rat
}

var naturalBase = [...]int64{0, 2, 4, 5, 7, 9, 11}

func zeroCoordinates() coordinates {
	return coordinates{t: new(big.Int), v: new(big.Rat)}
}

// coordinatesOf converts a set interval to exact coordinates.
func coordinatesOf(i interval.Interval) coordinates {
	size := i.Size()
	t := size - 1
	v := naturalBaseAt(size)
	num, den := i.Deviation().Fraction()
	v.Add(v, new(big.Rat).SetFrac64(int64(num), int64(den)))
	if i.IsDescending() {
		t = -t
		v.Neg(v)
	}
	return coordinates{t: big.NewInt(int64(t)), v: v}
}

// naturalBaseAt returns the exact chromatic coordinate of the basic kind of n.
func naturalBaseAt(n int) *big.Rat {
	octaves, remainder := (n-1)/7, (n-1)%7
	value := new(big.Int).Mul(big.NewInt(int64(octaves)), big.NewInt(12))
	value.Add(value, big.NewInt(naturalBase[remainder]))
	return new(big.Rat).SetInt(value)
}

func addCoordinates(a, b coordinates) coordinates {
	return coordinates{
		t: new(big.Int).Add(a.t, b.t),
		v: new(big.Rat).Add(a.v, b.v),
	}
}

func subCoordinates(a, b coordinates) coordinates {
	return coordinates{
		t: new(big.Int).Sub(a.t, b.t),
		v: new(big.Rat).Sub(a.v, b.v),
	}
}

func multiplyCoordinates(a coordinates, multiplier int) coordinates {
	m := big.NewInt(int64(multiplier))
	return coordinates{
		t: new(big.Int).Mul(a.t, m),
		v: new(big.Rat).Mul(a.v, new(big.Rat).SetInt(m)),
	}
}

func equalCoordinates(a, b coordinates) bool {
	return a.t.Cmp(b.t) == 0 && a.v.Cmp(b.v) == 0
}

// isAdjacent reports whether coordinates denote an ascending interval other
// than the perfect unison, without requiring that interval to be representable.
func isAdjacent(c coordinates) bool {
	return c.t.Sign() > 0 || (c.t.Sign() == 0 && c.v.Sign() > 0)
}

// intervalFromCoordinates materializes only the final result of an exact
// calculation. Intermediate values may exceed both int and alteration ranges.
func intervalFromCoordinates(c coordinates, op string) (interval.Interval, error) {
	if c.t.Sign() == 0 {
		deviation := new(big.Rat).Abs(c.v)
		alterationValue, err := alterationFromRational(deviation)
		if err != nil {
			return interval.Interval{}, rangeError(op, err)
		}
		result, err := interval.FromDeviation(1, alterationValue)
		if err != nil {
			return interval.Interval{}, rangeError(op, err)
		}
		if c.v.Sign() < 0 {
			result, err = result.Down()
			if err != nil {
				return interval.Interval{}, rangeError(op, err)
			}
		}
		return result, nil
	}

	absolute := new(big.Int).Abs(c.t)
	size := new(big.Int).Add(absolute, big.NewInt(1))
	if !size.IsInt64() || size.Sign() <= 0 || size.Cmp(big.NewInt(int64(math.MaxInt))) > 0 {
		return interval.Interval{}, rangeError(op, interval.ErrOutOfRange)
	}
	n := int(size.Int64())
	deviation := new(big.Rat).Set(c.v)
	if c.t.Sign() < 0 {
		deviation.Neg(deviation)
	}
	deviation.Sub(deviation, naturalBaseAt(n))
	alterationValue, err := alterationFromRational(deviation)
	if err != nil {
		return interval.Interval{}, rangeError(op, err)
	}
	result, err := interval.FromDeviation(n, alterationValue)
	if err != nil {
		return interval.Interval{}, rangeError(op, err)
	}
	if c.t.Sign() < 0 {
		result, err = result.Down()
		if err != nil {
			return interval.Interval{}, rangeError(op, err)
		}
	}
	return result, nil
}

// alterationFromRational converts a final rational deviation to alteration.
func alterationFromRational(value *big.Rat) (alteration.Alteration, error) {
	num, den := value.Num(), value.Denom()
	if !num.IsInt64() || !den.IsInt64() {
		return alteration.Alteration{}, alteration.ErrOutOfRange
	}
	return alteration.FromFraction(int(num.Int64()), int(den.Int64()))
}

// coordinatesBetweenSpellings computes interval.Between(from, to, Up) exactly
// without rejecting an intermediate result outside the alteration range.
func coordinatesBetweenSpellings(from, to spelling.Spelling) coordinates {
	distance := from.LetterDistanceUp(to)
	difference := rationalDifference(to.Alteration(), from.Alteration())
	if distance != 0 {
		natural := interval.MustBetween(from.Natural(), to.Natural(), interval.Up)
		result := coordinatesOf(natural)
		result.v.Add(result.v, difference)
		return result
	}
	if difference.Sign() > 0 {
		return coordinates{t: new(big.Int), v: difference}
	}
	// Moving upward against the alteration gives a diminished octave.
	return coordinates{
		t: big.NewInt(7),
		v: new(big.Rat).Add(new(big.Rat).SetInt64(12), difference),
	}
}

func rationalDifference(a, b alteration.Alteration) *big.Rat {
	aNum, aDen := a.Fraction()
	bNum, bDen := b.Fraction()
	return new(big.Rat).Sub(
		new(big.Rat).SetFrac64(int64(aNum), int64(aDen)),
		new(big.Rat).SetFrac64(int64(bNum), int64(bDen)),
	)
}

// sumCoordinates returns the exact sum of intervals in their given order.
func sumCoordinates(intervals []interval.Interval) coordinates {
	total := zeroCoordinates()
	for _, current := range intervals {
		total = addCoordinates(total, coordinatesOf(current))
	}
	return total
}

// prefixCoordinates returns the exact sum of the first count intervals.
func prefixCoordinates(intervals []interval.Interval, count int) coordinates {
	total := zeroCoordinates()
	for _, current := range intervals[:count] {
		total = addCoordinates(total, coordinatesOf(current))
	}
	return total
}
