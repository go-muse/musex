package interval

import (
	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/spelling"
)

// naturalOffsets holds the traditional steps of the natural letters from C,
// indexed by letter.Letter.Ordinal.
var naturalOffsets = [letter.Count]int64{0, 2, 4, 5, 7, 9, 11}

// naturalDeviation holds δ(l, r): the deviation of the natural interval from
// the natural letter l up r letters, indexed by the ordinal of l and r. It is
// -1, 0, or +1; for example, F up 3 letters is A4 (+1) and B up 4 letters is
// d5 (-1).
var naturalDeviation = func() (t [letter.Count][letter.Count]int64) {
	for l := range letter.Count {
		for r := range letter.Count {
			steps := (naturalOffsets[(l+r)%letter.Count] - naturalOffsets[l] + 12) % 12
			t[l][r] = steps - basicSemitones[r]
		}
	}
	return t
}()

// delta returns δ(l, r) for a valid letter l and 0 <= r <= 6.
func delta(l letter.Letter, r int) int64 {
	return naturalDeviation[l.Ordinal()][r]
}

// fifthsPlaces holds the places of the natural letters on the line of fifths
// (F -1, C 0, G 1, D 2, A 3, E 4, B 5), indexed by letter.Letter.Ordinal.
// The same values, indexed by r, are the places of the basic kinds of the
// simple sizes r + 1 (P1 0, M2 2, M3 4, P4 -1, P5 1, M6 3, M7 5).
var fifthsPlaces = [letter.Count]int{0, 2, 4, -1, 1, 3, 5}

// nameInterval returns the interval from x to y, both valid, in the given
// direction, and reports whether its deviation is in range.
func nameInterval(x, y spelling.Spelling, down bool) (Interval, bool) {
	ax, ay := x.Alteration(), y.Alteration()
	if !down {
		r := x.LetterDistanceUp(y)
		d, ok := combine(delta(x.Letter(), r), 1, ay, -1, ax)
		if !ok {
			return Interval{}, false
		}
		if r != 0 {
			return Interval{size: r + 1, dev: d}, true
		}
		// One letter: the same name, an augmented unison up, or an octave
		// up against the motion of the sound.
		switch d.Sign() {
		case 0:
			return PerfectUnison(), true
		case 1:
			return Interval{size: 1, dev: d}, true
		default:
			return Interval{size: 8, dev: d}, true
		}
	}
	r := x.LetterDistanceDown(y)
	d, ok := combine(delta(y.Letter(), r), 1, ax, -1, ay)
	if !ok {
		return Interval{}, false
	}
	if r != 0 {
		return Interval{size: r + 1, dev: d, down: true}, true
	}
	switch d.Sign() {
	case 0:
		return PerfectUnison(), true
	case 1:
		return Interval{size: 1, dev: d, down: true}, true
	default:
		return Interval{size: 8, dev: d, down: true}, true
	}
}

// transposeName returns x, valid, moved by the set interval i, and reports
// whether the alteration value of the result is in range.
func transposeName(x spelling.Spelling, i Interval) (spelling.Spelling, bool) {
	r := (i.size - 1) % 7
	lx := x.Letter()
	var (
		l  letter.Letter
		a  alteration.Alteration
		ok bool
	)
	if !i.down {
		l = lx.Shift(r)
		a, ok = combine(-delta(lx, r), 1, x.Alteration(), 1, i.dev)
	} else {
		l = lx.Shift(-r)
		a, ok = combine(delta(l, r), 1, x.Alteration(), -1, i.dev)
	}
	if !ok {
		return spelling.Spelling{}, false
	}
	// l is a valid letter, so New cannot fail.
	return spelling.MustNew(l, a), true
}

// Between returns the interval from the name x to the name y in the direction
// d. Names are compared by letters and alteration values, without a system.
// For different letters the size is the letter distance in the direction
// plus 1 and the quality is that of the natural interval of the letters,
// shifted by the difference of the alteration values: C to E up is M3, down
// the descending m6. Equal names give P1 in both directions. For one letter
// with different alteration values, the motion of the sound gives an
// augmented unison (C to C sharp up is A1, C to C flat down the descending
// A1), and the motion against it a diminished octave (C to C flat up is d8).
//
// Errors, in this order: ErrUnset when x or y is unset; ErrNoDirection when d
// is neither Up nor Down; ErrOutOfRange, together with
// alteration.ErrOutOfRange, when the deviation is outside the range of
// package alteration.
func Between(x, y spelling.Spelling, d Direction) (Interval, error) {
	const op = "Between"
	if err := checkNames(op, x, y); err != nil {
		return Interval{}, err
	}
	if !d.isSet() {
		return Interval{}, opError(op, ErrNoDirection)
	}
	i, ok := nameInterval(x, y, d == Down)
	if !ok {
		return Interval{}, alterationRangeError(op)
	}
	return i, nil
}

// MustBetween is like Between but panics if Between returns an error. The
// panic value is the error returned by Between.
func MustBetween(x, y spelling.Spelling, d Direction) Interval {
	return must(Between(x, y, d))
}

// checkNames returns ErrUnset for the first unset name.
func checkNames(op string, names ...spelling.Spelling) error {
	for _, s := range names {
		if !s.IsValid() {
			return detailError(op, ErrUnset, "name")
		}
	}
	return nil
}

// BothWays returns the intervals from x to y up and down: for different names
// they are inversions of each other (C and E give M3 and the descending m6);
// for equal names both are P1.
//
// Errors, in this order: ErrUnset when x or y is unset; ErrOutOfRange,
// together with alteration.ErrOutOfRange, when either deviation is outside
// the range of package alteration.
func BothWays(x, y spelling.Spelling) (up, down Interval, err error) {
	const op = "BothWays"
	if err := checkNames(op, x, y); err != nil {
		return Interval{}, Interval{}, err
	}
	up, ok := nameInterval(x, y, false)
	if !ok {
		return Interval{}, Interval{}, alterationRangeError(op)
	}
	down, ok = nameInterval(x, y, true)
	if !ok {
		return Interval{}, Interval{}, alterationRangeError(op)
	}
	return up, down, nil
}

// MustBothWays is like BothWays but panics if BothWays returns an error. The
// panic value is the error returned by BothWays.
func MustBothWays(x, y spelling.Spelling) (up, down Interval) {
	return must2(BothWays(x, y))
}

// Transpose returns the name x moved by the interval i: the letter moves by
// n - 1 letters in the direction of i, and the alteration value is such that
// the interval from x to the result in that direction has the quality of i.
// The octaves of i do not change the name. There is no enharmonic
// substitution: E moved by M3 is G sharp, and B sharp moved by A6 is G with
// three sharps. Names are not checked against any system.
//
// Errors, in this order: ErrUnset when x or i is unset; ErrOutOfRange,
// together with alteration.ErrOutOfRange, when the alteration value of the
// result is outside the range of package alteration.
func Transpose(x spelling.Spelling, i Interval) (spelling.Spelling, error) {
	const op = "Transpose"
	if err := checkNames(op, x); err != nil {
		return spelling.Spelling{}, err
	}
	if !i.IsValid() {
		return spelling.Spelling{}, detailError(op, ErrUnset, "interval")
	}
	s, ok := transposeName(x, i)
	if !ok {
		return spelling.Spelling{}, alterationRangeError(op)
	}
	return s, nil
}

// MustTranspose is like Transpose but panics if Transpose returns an error.
// The panic value is the error returned by Transpose.
func MustTranspose(x spelling.Spelling, i Interval) spelling.Spelling {
	return must(Transpose(x, i))
}

// fifthsOf returns the place of the valid name x on the line of fifths and
// reports whether its alteration value is whole.
func fifthsOf(x spelling.Spelling) (int, bool) {
	a := x.Alteration()
	if !a.IsInteger() {
		return 0, false
	}
	k, _ := a.Fraction()
	return fifthsPlaces[x.Letter().Ordinal()] + 7*k, true
}

// FifthsOf returns the place of the name x on the line of fifths: the place
// of its letter (F -1, C 0, G 1, D 2, A 3, E 4, B 5) plus 7 for each unit of
// its alteration value. F sharp is 6, B flat -2.
//
// Errors, in this order: ErrUnset when x is unset; ErrNotOnFifths when its
// alteration value is fractional.
func FifthsOf(x spelling.Spelling) (int, error) {
	const op = "FifthsOf"
	if err := checkNames(op, x); err != nil {
		return 0, err
	}
	q, ok := fifthsOf(x)
	if !ok {
		return 0, opError(op, ErrNotOnFifths)
	}
	return q, nil
}

// MustFifthsOf is like FifthsOf but panics if FifthsOf returns an error. The
// panic value is the error returned by FifthsOf.
func MustFifthsOf(x spelling.Spelling) int {
	return must(FifthsOf(x))
}

// FifthsBetween returns the difference of the places of y and x on the line
// of fifths: FifthsOf(y) - FifthsOf(x). C to E is 4, F to B is 6.
//
// Errors, in this order: ErrUnset when x or y is unset; ErrNotOnFifths when
// the alteration value of x, then of y, is fractional.
func FifthsBetween(x, y spelling.Spelling) (int, error) {
	const op = "FifthsBetween"
	if err := checkNames(op, x, y); err != nil {
		return 0, err
	}
	qx, ok := fifthsOf(x)
	if !ok {
		return 0, opError(op, ErrNotOnFifths)
	}
	qy, ok := fifthsOf(y)
	if !ok {
		return 0, opError(op, ErrNotOnFifths)
	}
	return qy - qx, nil
}

// MustFifthsBetween is like FifthsBetween but panics if FifthsBetween returns
// an error. The panic value is the error returned by FifthsBetween.
func MustFifthsBetween(x, y spelling.Spelling) int {
	return must(FifthsBetween(x, y))
}

// FromFifths returns the simple ascending interval whose names differ by n
// places on the line of fifths: the interval that Between gives up for any
// names with this difference. The letter distance is (4n) mod 7, and each 7
// places change the quality by one unit: 4 is M3, 6 is A4, -5 is m2, 7 is A1,
// and -7 is d8. Any int n is accepted.
//
// It returns ErrOutOfRange, together with alteration.ErrOutOfRange, when the
// deviation is outside the range of package alteration.
func FromFifths(n int) (Interval, error) {
	const op = "FromFifths"
	a, b := n/7, n%7
	if b < 0 {
		a, b = a-1, b+7
	}
	r := (4 * b) % 7
	size := r + 1
	if r == 0 && n < 0 {
		size = 8
	}
	// n - Q[r] is a multiple of 7, and b - Q[r] is 0 or 7.
	d, err := alteration.FromInt(a + (b-fifthsPlaces[r])/7)
	if err != nil {
		return Interval{}, rangeError(op, err)
	}
	return Interval{size: size, dev: d}, nil
}

// MustFromFifths is like FromFifths but panics if FromFifths returns an error.
// The panic value is the error returned by FromFifths.
func MustFromFifths(n int) Interval {
	return must(FromFifths(n))
}

// Fifths returns the place of i on the line of fifths: by how many places
// moving a name by i moves its place. The basic kinds have the places P1 0,
// m2 -5, M2 2, m3 -3, M3 4, P4 -1, P5 1, m6 -4, M6 3, m7 -2, M7 5, P8 0;
// each unit of deviation adds 7, octaves do not change the place, and a
// descending interval has the opposite place.
//
// Errors, in this order: ErrUnset for the unset interval; ErrNotOnFifths when
// the quality of i is fractional.
func (i Interval) Fifths() (int, error) {
	const op = "Fifths"
	if !i.IsValid() {
		return 0, opError(op, ErrUnset)
	}
	if !i.dev.IsInteger() {
		return 0, opError(op, ErrNotOnFifths)
	}
	k, _ := i.dev.Fraction()
	return i.sign() * (fifthsPlaces[(i.size-1)%7] + 7*k), nil
}

// MustFifths is like Fifths but panics if Fifths returns an error. The panic
// value is the error returned by Fifths.
func (i Interval) MustFifths() int {
	return must(i.Fifths())
}

// Reflect returns the name y that mirrors x around the axis name a by
// spelling: the interval from a down to y equals the interval from a up to x.
// With the axis D, C gives E, F gives B, and F sharp gives B flat. The axis
// gives itself, and reflecting twice gives x back.
//
// Errors, in this order: ErrUnset when a or x is unset; ErrOutOfRange,
// together with alteration.ErrOutOfRange, when a deviation or the alteration
// value of the result is outside the range of package alteration.
func Reflect(axis, x spelling.Spelling) (spelling.Spelling, error) {
	const op = "Reflect"
	if err := checkNames(op, axis, x); err != nil {
		return spelling.Spelling{}, err
	}
	i, ok := nameInterval(axis, x, false)
	if !ok {
		return spelling.Spelling{}, alterationRangeError(op)
	}
	y, ok := transposeName(axis, i.opposite())
	if !ok {
		return spelling.Spelling{}, alterationRangeError(op)
	}
	return y, nil
}

// MustReflect is like Reflect but panics if Reflect returns an error. The
// panic value is the error returned by Reflect.
func MustReflect(axis, x spelling.Spelling) spelling.Spelling {
	return must(Reflect(axis, x))
}
