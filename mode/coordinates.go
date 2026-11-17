package mode

import (
	"errors"
	"math/big"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/spelling"
)

// Coordinates describe signed letter steps and chromatic distance. They let
// operations range-check the requested result after all terms have canceled.
type coordinates struct {
	letters   *big.Int
	chromatic *big.Rat
}

const lettersPerOctave = 7
const semitonesPerOctave = 12

var naturalOffsets = [...]int64{0, 2, 4, 5, 7, 9, 11}

func rational(a alteration.Alteration) *big.Rat {
	n, d := a.Fraction()
	return new(big.Rat).SetFrac64(int64(n), int64(d))
}

func alterationOf(q *big.Rat) (alteration.Alteration, error) {
	if !q.Num().IsInt64() || !q.Denom().IsInt64() {
		return alteration.Alteration{}, alteration.ErrOutOfRange
	}
	return alteration.FromFraction(int(q.Num().Int64()), int(q.Denom().Int64()))
}

// splitLetters uses floor division so the residue also works for a descent.
func splitLetters(steps *big.Int) (*big.Int, int) {
	rem := new(big.Int)
	octaves, _ := new(big.Int).DivMod(steps, big.NewInt(lettersPerOctave), rem)
	return octaves, int(rem.Int64())
}

func naturalDistance(steps *big.Int, start int) *big.Rat {
	octaves, rem := splitLetters(steps)
	end := (start + rem) % lettersPerOctave
	if start+rem >= lettersPerOctave {
		octaves.Add(octaves, big.NewInt(1))
	}
	octaves.Mul(octaves, big.NewInt(semitonesPerOctave))
	octaves.Add(octaves, big.NewInt(naturalOffsets[end]-naturalOffsets[start]))
	return new(big.Rat).SetInt(octaves)
}

func coordinatesOf(i interval.Interval) coordinates {
	letters := big.NewInt(int64(i.Size() - 1))
	chromatic := naturalDistance(letters, 0)
	chromatic.Add(chromatic, rational(i.Deviation()))
	if i.IsDescending() {
		letters.Neg(letters)
		chromatic.Neg(chromatic)
	}
	return coordinates{letters, chromatic}
}

func (a coordinates) add(b coordinates) coordinates {
	return coordinates{
		new(big.Int).Add(a.letters, b.letters),
		new(big.Rat).Add(a.chromatic, b.chromatic),
	}
}

func (a coordinates) sub(b coordinates) coordinates {
	return coordinates{
		new(big.Int).Sub(a.letters, b.letters),
		new(big.Rat).Sub(a.chromatic, b.chromatic),
	}
}

func (a coordinates) mul(n int) coordinates {
	return coordinates{
		new(big.Int).Mul(a.letters, big.NewInt(int64(n))),
		new(big.Rat).Mul(a.chromatic, new(big.Rat).SetInt64(int64(n))),
	}
}

func (a coordinates) equal(b coordinates) bool {
	return a.letters.Cmp(b.letters) == 0 && a.chromatic.Cmp(b.chromatic) == 0
}

// Interval tables move upward in the written order. Their exact coordinates
// have nonnegative letter steps, including for augmented unisons.
func (a coordinates) deviation() *big.Rat {
	return new(big.Rat).Sub(a.chromatic, naturalDistance(a.letters, 0))
}

func (a coordinates) ascendingInterval() (interval.Interval, error) {
	size := new(big.Int).Add(a.letters, big.NewInt(1))
	if !size.IsInt64() {
		return interval.Interval{}, errors.Join(ErrOutOfRange, interval.ErrOutOfRange)
	}
	d, err := alterationOf(a.deviation())
	if err != nil {
		return interval.Interval{}, errors.Join(ErrOutOfRange, interval.ErrOutOfRange, err)
	}
	return interval.FromDeviation(int(size.Int64()), d)
}

func (a coordinates) qualities() (diatonic, chromatic bool) {
	d := a.deviation()
	if !d.IsInt() {
		return false, false
	}
	value, err := alterationOf(d)
	if err != nil {
		// The whole deviations that give basic intervals are -1, 0, and 1.
		return false, true
	}
	_, rem := splitLetters(a.letters)
	size := rem + 1
	if size == 1 && a.letters.Sign() != 0 {
		size = 8
	}
	i := interval.MustFromDeviation(size, value)
	return i.IsDiatonic(), i.IsChromatic()
}

func (a coordinates) transposeName(start spelling.Spelling) (spelling.Spelling, error) {
	letter := start.Letter()
	_, rem := splitLetters(a.letters)
	accidental := new(big.Rat).Add(rational(start.Alteration()), a.chromatic)
	accidental.Sub(accidental, naturalDistance(a.letters, letter.Ordinal()))
	value, err := alterationOf(accidental)
	if err != nil {
		return spelling.Spelling{}, errors.Join(interval.ErrOutOfRange, err)
	}
	return spelling.New(letter.Shift(rem), value)
}
