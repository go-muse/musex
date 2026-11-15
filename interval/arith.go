package interval

import "github.com/go-muse/musex/alteration"

// AddOctaves returns i widened by k >= 0 octaves: the size n + 7k with the
// same quality and direction. P1 widened by one octave is P8, and A1 is A8.
//
// Errors, in this order: ErrUnset for the unset interval; ErrInvalidArgument
// for k < 0; ErrOutOfRange when n + 7k is greater than MaxSize.
func (i Interval) AddOctaves(k int) (Interval, error) {
	const op = "AddOctaves"
	if !i.IsValid() {
		return Interval{}, opError(op, ErrUnset)
	}
	if k < 0 {
		return Interval{}, detailError(op, ErrInvalidArgument, "negative number of octaves")
	}
	if k > (MaxSize-i.size)/7 {
		return Interval{}, detailError(op, ErrOutOfRange, "size greater than MaxSize")
	}
	return newInterval(i.size+7*k, i.dev, i.down), nil
}

// MustAddOctaves is like AddOctaves but panics if AddOctaves returns an error.
// The panic value is the error returned by AddOctaves.
func (i Interval) MustAddOctaves(k int) Interval {
	return must(i.AddOctaves(k))
}

// changeQuality returns i with the deviation d. The sign of the exact result
// is given by negative when the result is out of range, so that a unison
// narrower than perfect is reported before the range.
func (i Interval) changeQuality(op string, d alteration.Alteration, ok, negative bool) (Interval, error) {
	if !i.IsValid() {
		return Interval{}, opError(op, ErrUnset)
	}
	if ok {
		negative = d.Sign() < 0
	}
	if i.size == 1 && negative {
		return Interval{}, detailError(op, ErrInvalidArgument, "diminished unison")
	}
	if !ok {
		return Interval{}, alterationRangeError(op)
	}
	return newInterval(i.size, d, i.down), nil
}

// Expand returns i one chromatic unit wider: the deviation d + 1, in the
// same direction. M3 becomes A3, m3 becomes M3.
//
// Errors, in this order: ErrUnset for the unset interval; ErrInvalidArgument
// for a unison that would become narrower than perfect; ErrOutOfRange,
// together with alteration.ErrOutOfRange, when the deviation is outside the
// range of package alteration.
func (i Interval) Expand() (Interval, error) {
	return i.expandBy("Expand", 1)
}

// MustExpand is like Expand but panics if Expand returns an error. The panic
// value is the error returned by Expand.
func (i Interval) MustExpand() Interval {
	return must(i.Expand())
}

// Contract returns i one chromatic unit narrower: the deviation d - 1, in the
// same direction. M3 becomes m3; P1 has no narrower unison. The errors are
// those of Expand.
func (i Interval) Contract() (Interval, error) {
	return i.contractBy("Contract", 1)
}

// MustContract is like Contract but panics if Contract returns an error. The
// panic value is the error returned by Contract.
func (i Interval) MustContract() Interval {
	return must(i.Contract())
}

// ExpandBy returns i widened by k chromatic units: the deviation d + k; a
// negative k narrows it. Any int k is accepted, and the result is exact.
// The errors are those of Expand.
func (i Interval) ExpandBy(k int) (Interval, error) {
	return i.expandBy("ExpandBy", k)
}

// MustExpandBy is like ExpandBy but panics if ExpandBy returns an error. The
// panic value is the error returned by ExpandBy.
func (i Interval) MustExpandBy(k int) Interval {
	return must(i.ExpandBy(k))
}

// ContractBy returns i narrowed by k chromatic units: the deviation d - k; a
// negative k widens it. Any int k is accepted, including math.MinInt, and the
// result is exact. The errors are those of Expand.
func (i Interval) ContractBy(k int) (Interval, error) {
	return i.contractBy("ContractBy", k)
}

// MustContractBy is like ContractBy but panics if ContractBy returns an error.
// The panic value is the error returned by ContractBy.
func (i Interval) MustContractBy(k int) Interval {
	return must(i.ContractBy(k))
}

// expandBy implements ExpandBy for the operation op.
func (i Interval) expandBy(op string, k int) (Interval, error) {
	d, err := i.dev.RaiseBy(k)
	// Out of range, |d + k| > MaxMagnitude >= |d|, so d + k has the sign of k.
	return i.changeQuality(op, d, err == nil, k < 0)
}

// contractBy implements ContractBy for the operation op.
func (i Interval) contractBy(op string, k int) (Interval, error) {
	d, err := i.dev.LowerBy(k)
	// Out of range, |d - k| > MaxMagnitude >= |d|, so d - k has the sign of -k.
	return i.changeQuality(op, d, err == nil, k > 0)
}

// ExpandByAlteration returns i widened by the alteration value d: the
// deviation of i plus d, which can be fractional. The errors are those of Expand.
func (i Interval) ExpandByAlteration(d alteration.Alteration) (Interval, error) {
	return i.byAlteration("ExpandByAlteration", d, 1)
}

// MustExpandByAlteration is like ExpandByAlteration but panics if
// ExpandByAlteration returns an error. The panic value is the error returned
// by ExpandByAlteration.
func (i Interval) MustExpandByAlteration(d alteration.Alteration) Interval {
	return must(i.ExpandByAlteration(d))
}

// ContractByAlteration returns i narrowed by the alteration value d: the
// deviation of i minus d. The errors are those of Expand.
func (i Interval) ContractByAlteration(d alteration.Alteration) (Interval, error) {
	return i.byAlteration("ContractByAlteration", d, -1)
}

// MustContractByAlteration is like ContractByAlteration but panics if
// ContractByAlteration returns an error. The panic value is the error
// returned by ContractByAlteration.
func (i Interval) MustContractByAlteration(d alteration.Alteration) Interval {
	return must(i.ContractByAlteration(d))
}

// byAlteration returns i with the deviation d + s·a, s = ±1.
func (i Interval) byAlteration(op string, a alteration.Alteration, s int64) (Interval, error) {
	pd, qd := fractionOf(i.dev)
	pa, qa := fractionOf(a)
	num := pd*qa + s*pa*qd
	d, ok := alterationOf(i128Of(num), qd*qa)
	return i.changeQuality(op, d, ok, num < 0)
}

// Invert returns the inversion of the simple part (n1, d) of i: the size
// 9 - n1 with the deviation -d in the perfect group or -1 - d in the
// major-minor group, in the same direction; a resulting perfect unison is
// ascending. M3 inverts to m6, A4 to d5, P1 to P8, P8 to P1, M10 to m6, and
// P15 to P1.
//
// Errors, in this order: ErrUnset for the unset interval; ErrNoInversion when
// the simple part is an octave with a positive deviation (an augmented
// octave), whose inversion would be a diminished unison; ErrOutOfRange,
// together with alteration.ErrOutOfRange, when the deviation of the inversion
// is outside the range of package alteration.
func (i Interval) Invert() (Interval, error) {
	return i.invert("Invert")
}

// invert implements Invert for the operation op.
func (i Interval) invert(op string) (Interval, error) {
	if !i.IsValid() {
		return Interval{}, opError(op, ErrUnset)
	}
	n1 := simpleSizeOf(i.size)
	if n1 == 8 && i.dev.Sign() > 0 {
		return Interval{}, opError(op, ErrNoInversion)
	}
	base := alteration.Natural()
	if !perfectSize(n1) {
		base = alteration.Flat()
	}
	d, err := base.Sub(i.dev)
	if err != nil {
		return Interval{}, rangeError(op, err)
	}
	return newInterval(9-n1, d, i.down), nil
}

// MustInvert is like Invert but panics if Invert returns an error. The panic
// value is the error returned by Invert.
func (i Interval) MustInvert() Interval {
	return must(i.Invert())
}

// InversionClass returns the simple part of i and its inversion. The errors
// are those of Invert.
func (i Interval) InversionClass() (simple, inversion Interval, err error) {
	inv, err := i.invert("InversionClass")
	if err != nil {
		return Interval{}, Interval{}, err
	}
	return newInterval(simpleSizeOf(i.size), i.dev, i.down), inv, nil
}

// MustInversionClass is like InversionClass but panics if InversionClass
// returns an error. The panic value is the error returned by InversionClass.
func (i Interval) MustInversionClass() (simple, inversion Interval) {
	return must2(i.InversionClass())
}

// basicSemitones holds B[r]: the steps of the traditional system spanned by
// the basic kind of the simple sizes 1 to 7 (r = n - 1 from 0 to 6).
var basicSemitones = [7]int64{0, 2, 4, 5, 7, 9, 11}

// linear accumulates an integer combination of intervals in their linear
// coordinates. An interval (σ, n, d) with n - 1 = 7q + r has the letter
// coordinate t = σ(7q + r) and the traditional step coordinate
// v = σ(12q + B[r] + d). The sum of intervals is the interval whose
// coordinates are the sums of their coordinates; these coordinates are a
// device of calculation, not a positional size in some system.
//
// The fields hold the parts of the combination: oct = Σ kσq, rem = Σ kσr,
// and the excess Σ kσ(B[r] + d) as the fraction exc/den. All values are exact.
type linear struct {
	oct i128
	rem i128
	exc i128
	den int64
}

// newLinear returns the empty combination, which is the perfect unison.
func newLinear() linear {
	return linear{den: 1}
}

// addTerm adds k copies of the set interval i to the combination.
func (l *linear) addTerm(k int, i Interval) {
	s := int64(i.sign())
	nm1 := int64(i.size - 1)
	q, r := nm1/7, nm1%7
	l.oct = l.oct.add(mul64(int64(k), s*q))
	l.rem = l.rem.add(mul64(int64(k), s*r))
	p, den := fractionOf(i.dev)
	term := mul64(int64(k), s*(basicSemitones[r]*den+p))
	l.exc = l.exc.mul(den).add(term.mul(l.den))
	l.den *= den
}

// result returns the interval with the coordinates of the combination.
func (l *linear) result(op string) (Interval, error) {
	c, r := l.rem.floorDivMod(7)
	oct := l.oct.add(c)
	// W = excess - 12c: the step coordinate is 12·oct + W.
	w := l.exc.sub(c.mul(12 * l.den))
	var (
		nm1  i128
		num  i128
		down bool
	)
	switch {
	case oct.sign() == 0 && r == 0:
		num, down = w, w.sign() < 0
		if down {
			num = w.neg()
		}
	case oct.sign() > 0 || (oct.sign() == 0 && r > 0):
		nm1 = oct.mul(7).add(i128Of(r))
		num = w.sub(i128Of(basicSemitones[r] * l.den))
	default:
		down = true
		if r == 0 {
			nm1 = oct.neg().mul(7)
			num = w.neg()
		} else {
			rho := 7 - r
			nm1 = oct.neg().sub(i128Of(1)).mul(7).add(i128Of(rho))
			num = w.neg().add(i128Of((12 - basicSemitones[rho]) * l.den))
		}
	}
	n, ok := fitsSize(nm1)
	if !ok {
		return Interval{}, detailError(op, ErrOutOfRange, "size greater than MaxSize")
	}
	d, ok := alterationOf(num, l.den)
	if !ok {
		return Interval{}, alterationRangeError(op)
	}
	return newInterval(n, d, down), nil
}

// Add returns the sum of i and j: the interval s such that moving any name or
// pitch by s equals moving it by i and then by j. Addition is commutative and
// associative, P1 is its neutral element, and the sum of an interval and its
// opposite is P1: M3 + m3 = P5, P5 + P4 = P8, m2 + descending A1 = d2.
// The result is exact whenever it is representable.
//
// Errors, in this order: ErrUnset when i or j is unset; ErrOutOfRange when
// the size of the sum is greater than MaxSize or its deviation is outside the
// range of package alteration (then also alteration.ErrOutOfRange).
func (i Interval) Add(j Interval) (Interval, error) {
	return i.combine("Add", j, 1)
}

// MustAdd is like Add but panics if Add returns an error. The panic value is
// the error returned by Add.
func (i Interval) MustAdd(j Interval) Interval {
	return must(i.Add(j))
}

// Sub returns the difference of i and j: the sum of i and the opposite of j.
// M9 - P8 = M2, P8 - d8 = A1. The errors are those of Add.
func (i Interval) Sub(j Interval) (Interval, error) {
	return i.combine("Sub", j, -1)
}

// MustSub is like Sub but panics if Sub returns an error. The panic value is
// the error returned by Sub.
func (i Interval) MustSub(j Interval) Interval {
	return must(i.Sub(j))
}

// combine returns i + k·j for k = ±1.
func (i Interval) combine(op string, j Interval, k int) (Interval, error) {
	if !i.IsValid() {
		return Interval{}, detailError(op, ErrUnset, "receiver")
	}
	if !j.IsValid() {
		return Interval{}, detailError(op, ErrUnset, "argument")
	}
	l := newLinear()
	l.addTerm(1, i)
	l.addTerm(k, j)
	return l.result(op)
}

// Mul returns the sum of k copies of i; for k < 0, of |k| copies of its
// opposite; for k = 0, P1. 2 × P5 = M9, 3 × M3 = A7. Any int k is accepted,
// including math.MinInt, and the result is exact whenever it is
// representable.
//
// Errors, in this order: ErrUnset for the unset interval; ErrOutOfRange as
// for Add.
func (i Interval) Mul(k int) (Interval, error) {
	const op = "Mul"
	if !i.IsValid() {
		return Interval{}, opError(op, ErrUnset)
	}
	l := newLinear()
	l.addTerm(k, i)
	return l.result(op)
}

// MustMul is like Mul but panics if Mul returns an error. The panic value is
// the error returned by Mul.
func (i Interval) MustMul(k int) Interval {
	return must(i.Mul(k))
}
