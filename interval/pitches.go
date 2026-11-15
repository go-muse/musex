package interval

import (
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

// inadmissibleFunc returns the error for a name that is not admissible:
// a cause of package system for an explicit system or a cause of package note
// for the system of the settings of a note.
type inadmissibleFunc func(op, what string) error

// checkSystem returns ErrUnset when sys is nil or unset.
func checkSystem(op string, sys *system.System) error {
	if !sys.IsValid() {
		return detailError(op, ErrUnset, "system")
	}
	return nil
}

// checkPitches returns ErrUnset for the first unset pitch.
func checkPitches(op string, ps ...pitch.Pitch) error {
	for _, p := range ps {
		if !p.IsValid() {
			return detailError(op, ErrUnset, "pitch")
		}
	}
	return nil
}

// admits reports whether the valid name is admissible in the valid system.
func admits(sys *system.System, name spelling.Spelling) bool {
	ok, err := sys.Admits(name)
	return err == nil && ok
}

// checkAdmitted returns the error of inadm for the first name of ps that is
// not admissible in sys.
func checkAdmitted(op string, sys *system.System, inadm inadmissibleFunc, what string, names ...spelling.Spelling) error {
	for _, s := range names {
		if !admits(sys, s) {
			return inadm(op, what)
		}
	}
	return nil
}

// pitchInterval returns the interval from x to y in sys. The system and the
// pitches are valid, and their names are admissible.
func pitchInterval(op string, sys *system.System, x, y pitch.Pitch) (Interval, error) {
	wx, err := sys.WrittenOctave(x)
	if err != nil {
		return Interval{}, rangeError(op, err)
	}
	wy, err := sys.WrittenOctave(y)
	if err != nil {
		return Interval{}, rangeError(op, err)
	}
	dw, err := wy.Sub(wx)
	if err != nil {
		return Interval{}, rangeError(op, err)
	}
	sx, sy := x.Spelling(), y.Spelling()
	b := sys.Boundary()
	hx, hy := int64(b.DistanceUp(sx.Letter())), int64(b.DistanceUp(sy.Letter()))
	steps := mul64(int64(dw), 7).add(i128Of(hy - hx))
	ax, ay := sx.Alteration(), sy.Alteration()
	switch steps.sign() {
	case 0:
		d, ok := combine(0, 1, ay, -1, ax)
		if !ok {
			return Interval{}, alterationRangeError(op)
		}
		if d.Sign() >= 0 {
			return newInterval(1, d, false), nil
		}
		d, ok = combine(0, 1, ax, -1, ay)
		if !ok {
			return Interval{}, alterationRangeError(op)
		}
		return newInterval(1, d, true), nil
	case 1:
		n, ok := fitsSize(steps)
		if !ok {
			return Interval{}, detailError(op, ErrOutOfRange, "size greater than MaxSize")
		}
		d, ok := combine(delta(sx.Letter(), (n-1)%7), 1, ay, -1, ax)
		if !ok {
			return Interval{}, alterationRangeError(op)
		}
		return newInterval(n, d, false), nil
	default:
		n, ok := fitsSize(steps.neg())
		if !ok {
			return Interval{}, detailError(op, ErrOutOfRange, "size greater than MaxSize")
		}
		d, ok := combine(delta(sy.Letter(), (n-1)%7), 1, ax, -1, ay)
		if !ok {
			return Interval{}, alterationRangeError(op)
		}
		return newInterval(n, d, true), nil
	}
}

// transposePitch returns x moved by i in sys. The system, the pitch, and the
// interval are valid, and the name of x is admissible; inadm reports the name
// of the result that is not admissible.
func transposePitch(op string, sys *system.System, x pitch.Pitch, i Interval, inadm inadmissibleFunc) (pitch.Pitch, error) {
	sx := x.Spelling()
	name, ok := transposeName(sx, i)
	if !ok {
		return pitch.Pitch{}, alterationRangeError(op)
	}
	if !admits(sys, name) {
		return pitch.Pitch{}, inadm(op, "name of the result")
	}
	wx, err := sys.WrittenOctave(x)
	if err != nil {
		return pitch.Pitch{}, rangeError(op, err)
	}
	h := sys.Boundary().DistanceUp(sx.Letter())
	q, r := (i.size-1)/7, (i.size-1)%7
	if !i.down {
		if h+r >= 7 {
			q++
		}
		wx, err = wx.Shift(q)
	} else {
		if h-r < 0 {
			q++
		}
		wx, err = wx.ShiftDown(q)
	}
	if err != nil {
		return pitch.Pitch{}, rangeError(op, err)
	}
	p, err := sys.FromWritten(name, wx)
	if err != nil {
		return pitch.Pitch{}, rangeError(op, err)
	}
	return p, nil
}

// BetweenPitches returns the interval from the written pitch x to the written
// pitch y in the system sys. Each pitch has the letter coordinate
// u = 7·w + h, where w is its written octave in sys and h the letter distance
// up from the boundary letter of sys to its letter. For L = u(y) - u(x) > 0
// the interval is ascending with the size L + 1, for L < 0 descending with
// the size 1 - L; its quality is given by the letters and the alteration
// values, as in Between. For L = 0 it is P1 for equal alteration values and
// otherwise an augmented unison up or down. The direction follows the
// writing, not the position: in the traditional system B sharp 3 to C4 is d2
// up, although both have the same position. The system is never chosen by
// default.
//
// Errors, in this order: ErrUnset when sys is nil or unset or when x or y is
// unset; system.ErrInadmissible when the name of x, then of y, is not
// admissible in sys; ErrOutOfRange when a written octave does not fit in an
// int (with system.ErrOutOfRange), the difference of the written octaves does
// not fit in an int (with octave.ErrOutOfRange), the size is greater than
// MaxSize, or the deviation is outside the range of package alteration (with
// alteration.ErrOutOfRange).
func BetweenPitches(sys *system.System, x, y pitch.Pitch) (Interval, error) {
	const op = "BetweenPitches"
	if err := checkSystem(op, sys); err != nil {
		return Interval{}, err
	}
	if err := checkPitches(op, x, y); err != nil {
		return Interval{}, err
	}
	if err := checkAdmitted(op, sys, inadmissibleError, "name", x.Spelling(), y.Spelling()); err != nil {
		return Interval{}, err
	}
	return pitchInterval(op, sys, x, y)
}

// MustBetweenPitches is like BetweenPitches but panics if BetweenPitches
// returns an error. The panic value is the error returned by BetweenPitches.
func MustBetweenPitches(sys *system.System, x, y pitch.Pitch) Interval {
	return must(BetweenPitches(sys, x, y))
}

// TransposePitch returns the written pitch x moved by the interval i in the
// system sys. Its name is Transpose of the name of x; its written octave is
// the written octave of x plus the octaves crossed by moving the letter
// coordinate by n - 1 in the direction of i; sys gives its octave. The
// interval from x to the result is i, and moving by the opposite of i gives x
// back. In the traditional system C4 moved by M10 is E5, and C4 moved down by
// d2 is B sharp 3.
//
// Errors, in this order: ErrUnset when sys is nil or unset or when x or i is
// unset; system.ErrInadmissible when the name of x is not admissible in sys;
// ErrOutOfRange, together with alteration.ErrOutOfRange, when the alteration
// value of the result is outside the range of package alteration;
// system.ErrInadmissible when the name of the result is not admissible;
// ErrOutOfRange when a written octave or the octave does not fit in an int
// (with system.ErrOutOfRange or octave.ErrOutOfRange).
func TransposePitch(sys *system.System, x pitch.Pitch, i Interval) (pitch.Pitch, error) {
	const op = "TransposePitch"
	if err := checkSystem(op, sys); err != nil {
		return pitch.Pitch{}, err
	}
	if err := checkPitches(op, x); err != nil {
		return pitch.Pitch{}, err
	}
	if !i.IsValid() {
		return pitch.Pitch{}, detailError(op, ErrUnset, "interval")
	}
	if err := checkAdmitted(op, sys, inadmissibleError, "name", x.Spelling()); err != nil {
		return pitch.Pitch{}, err
	}
	return transposePitch(op, sys, x, i, inadmissibleError)
}

// MustTransposePitch is like TransposePitch but panics if TransposePitch
// returns an error. The panic value is the error returned by TransposePitch.
func MustTransposePitch(sys *system.System, x pitch.Pitch, i Interval) pitch.Pitch {
	return must(TransposePitch(sys, x, i))
}

// ReflectPitch returns the written pitch that mirrors x around the axis pitch
// a in the system sys: a moved by the opposite of the interval from a to x.
// In the traditional system with the axis D4, C4 gives E4 and F sharp 4 gives
// B flat 3.
//
// Errors, in this order: ErrUnset when sys is nil or unset or when a or x is
// unset; system.ErrInadmissible when the name of a, then of x, is not
// admissible; then the errors of BetweenPitches for the interval and of
// TransposePitch for the result.
func ReflectPitch(sys *system.System, axis, x pitch.Pitch) (pitch.Pitch, error) {
	const op = "ReflectPitch"
	if err := checkSystem(op, sys); err != nil {
		return pitch.Pitch{}, err
	}
	if err := checkPitches(op, axis, x); err != nil {
		return pitch.Pitch{}, err
	}
	if err := checkAdmitted(op, sys, inadmissibleError, "name", axis.Spelling(), x.Spelling()); err != nil {
		return pitch.Pitch{}, err
	}
	i, err := pitchInterval(op, sys, axis, x)
	if err != nil {
		return pitch.Pitch{}, err
	}
	return transposePitch(op, sys, axis, i.opposite(), inadmissibleError)
}

// MustReflectPitch is like ReflectPitch but panics if ReflectPitch returns an
// error. The panic value is the error returned by ReflectPitch.
func MustReflectPitch(sys *system.System, axis, x pitch.Pitch) pitch.Pitch {
	return must(ReflectPitch(sys, axis, x))
}
