package interval

import (
	"iter"
	"math"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

// stepsAfter returns S(l, r): the sum of the steps of sys after r letters
// going up from the letter l.
func stepsAfter(sys *system.System, l letter.Letter, r int) int64 {
	var s int64
	for j := range r {
		st, _ := sys.StepAfter(l.Shift(j))
		s += int64(st)
	}
	return s
}

// isRegular reports whether the valid system sys is regular.
func isRegular(sys *system.System) bool {
	reg, err := sys.Regularity()
	return err == nil && reg.Regular
}

// gcd64 returns the greatest common divisor of |a| and b > 0.
func gcd64(a, b int64) int64 {
	if a < 0 {
		a = -a
	}
	for a != 0 {
		a, b = b%a, a
	}
	return b
}

// scaledSteps returns a·c as a whole number of steps and reports whether it
// is whole; a is the exact fraction num/den with den > 0.
func scaledSteps(num, den int64, c int64) (i128, bool) {
	g := gcd64(num, den)
	num, den = num/g, den/g
	if c%den != 0 {
		return i128{}, false
	}
	return mul64(num, c/den), true
}

// sizeCalc computes the positional size of an interval in stages, so that a
// failure of one category is found for every input before the next category
// is checked.
type sizeCalc struct {
	sys   *system.System
	i     Interval
	start spelling.Spelling
	end   spelling.Spelling
}

// checkStart reports ErrStartRequired for a missing start name in a system
// that is not regular.
func (sc *sizeCalc) checkStart(op string) error {
	if !sc.start.IsValid() && !isRegular(sc.sys) {
		return opError(op, ErrStartRequired)
	}
	return nil
}

// checkStartAdmitted reports a start name that is not admissible.
func (sc *sizeCalc) checkStartAdmitted(op string) error {
	if sc.start.IsValid() && !admits(sc.sys, sc.start) {
		return inadmissibleError(op, "start name")
	}
	return nil
}

// checkWhole reports ErrFractionalSteps when, without a start name, d·c is
// not a whole number.
func (sc *sizeCalc) checkWhole(op string) error {
	if sc.start.IsValid() {
		return nil
	}
	p, q := fractionOf(sc.i.dev)
	if _, ok := scaledSteps(p, q, int64(sc.sys.Chromatic())); !ok {
		return opError(op, ErrFractionalSteps)
	}
	return nil
}

// buildEnd computes the name reached from the start name by the interval.
func (sc *sizeCalc) buildEnd(op string) error {
	if !sc.start.IsValid() {
		return nil
	}
	z, ok := transposeName(sc.start, sc.i)
	if !ok {
		return alterationRangeError(op)
	}
	sc.end = z
	return nil
}

// checkEndAdmitted reports an end name that is not admissible.
func (sc *sizeCalc) checkEndAdmitted(op string) error {
	if sc.start.IsValid() && !admits(sc.sys, sc.end) {
		return inadmissibleError(op, "name reached from the start name")
	}
	return nil
}

// size returns the positional size after the other stages succeeded.
func (sc *sizeCalc) size(op string) (system.Step, error) {
	sys, i := sc.sys, sc.i
	nm1 := int64(i.size - 1)
	q, r := nm1/7, int(nm1%7)
	n := int64(sys.Classes())
	c := int64(sys.Chromatic())
	total := mul64(q, n)
	if !sc.start.IsValid() {
		p, den := fractionOf(i.dev)
		dc, _ := scaledSteps(p, den, c)
		total = total.add(i128Of(stepsAfter(sys, letter.C, r))).add(dc)
	} else {
		lower, upper := sc.start, sc.end
		if i.down {
			lower, upper = upper, lower
		}
		pu, qu := fractionOf(upper.Alteration())
		pl, ql := fractionOf(lower.Alteration())
		shift, ok := scaledSteps(pu*ql-pl*qu, qu*ql, c)
		if !ok {
			return 0, opError(op, ErrFractionalSteps)
		}
		total = total.add(i128Of(stepsAfter(sys, lower.Letter(), r))).add(shift)
	}
	v, ok := total.intValue()
	if !ok {
		return 0, detailError(op, ErrOutOfRange, "number of steps does not fit in an int")
	}
	return system.Step(v), nil
}

// sizeStages is the number of checks of sizeCalc before the size itself.
const sizeStages = 5

// stage runs the check number k of sc, in the order of categories.
func (sc *sizeCalc) stage(op string, k int) error {
	switch k {
	case 0:
		return sc.checkStart(op)
	case 1:
		return sc.checkStartAdmitted(op)
	case 2:
		return sc.checkWhole(op)
	case 3:
		return sc.buildEnd(op)
	default:
		return sc.checkEndAdmitted(op)
	}
}

// positionalSize returns the positional size of the valid interval i in the
// valid system sys from the optional start name.
func positionalSize(op string, sys *system.System, i Interval, start spelling.Spelling) (system.Step, spelling.Spelling, error) {
	sc := sizeCalc{sys: sys, i: i, start: start}
	for k := range sizeStages {
		if err := sc.stage(op, k); err != nil {
			return 0, spelling.Spelling{}, err
		}
	}
	s, err := sc.size(op)
	if err != nil {
		return 0, spelling.Spelling{}, err
	}
	return s, sc.end, nil
}

// checkSystemInterval returns ErrUnset for an unset system, then for an unset
// interval.
func checkSystemInterval(op string, sys *system.System, is ...Interval) error {
	if err := checkSystem(op, sys); err != nil {
		return err
	}
	for _, i := range is {
		if !i.IsValid() {
			return detailError(op, ErrUnset, "interval")
		}
	}
	return nil
}

// Steps returns the positional size of i in the system sys: the number of
// steps of the system by which the upper note of the interval is above the
// lower one, regardless of the direction of i. The size can be zero or
// negative: in the traditional system d2 is 0 and dd2 is -1.
//
// The start name is the name from which i is laid off in its direction:
// the lower name of an ascending interval and the upper name of a descending
// one. With a start name x the size is computed from x and the name
// Transpose(x, i): the steps of sys between their letters plus the difference
// of their alteration values times the chromatic unit c of sys. Both names
// must be admissible in sys. The zero spelling.Spelling{} means that no start
// name is given; this is allowed only in a regular system, where the size does
// not depend on the start: the size of the basic kind of the size of i plus
// d·c. Names are then not checked.
//
// Errors, in this order: ErrUnset when sys is nil or unset or when i is unset;
// ErrStartRequired when no start name is given and sys is not regular;
// system.ErrInadmissible when the start name is not admissible;
// ErrFractionalSteps when no start name is given and d·c is not whole (the
// neutral third in the traditional system); ErrOutOfRange, together with
// alteration.ErrOutOfRange, when the alteration value of the name reached
// from the start name is out of range; system.ErrInadmissible when that name
// is not admissible; ErrOutOfRange when the size does not fit in an int.
func (i Interval) Steps(sys *system.System, start spelling.Spelling) (system.Step, error) {
	const op = "Steps"
	if err := checkSystemInterval(op, sys, i); err != nil {
		return 0, err
	}
	s, _, err := positionalSize(op, sys, i, start)
	return s, err
}

// MustSteps is like Steps but panics if Steps returns an error. The panic
// value is the error returned by Steps.
func (i Interval) MustSteps(sys *system.System, start spelling.Spelling) system.Step {
	return must(i.Steps(sys, start))
}

// SignedSteps returns the shift in steps of i in the system sys: the
// positional size for an ascending interval and its negation for a
// descending one, that is, by how many steps moving a pitch by i changes its
// position. The errors are those of Steps; the negated size that does not fit
// in an int is also ErrOutOfRange.
func (i Interval) SignedSteps(sys *system.System, start spelling.Spelling) (system.Step, error) {
	const op = "SignedSteps"
	if err := checkSystemInterval(op, sys, i); err != nil {
		return 0, err
	}
	s, _, err := positionalSize(op, sys, i, start)
	if err != nil {
		return 0, err
	}
	if !i.down {
		return s, nil
	}
	if s == math.MinInt {
		return 0, detailError(op, ErrOutOfRange, "number of steps does not fit in an int")
	}
	return -s, nil
}

// MustSignedSteps is like SignedSteps but panics if SignedSteps returns an
// error. The panic value is the error returned by SignedSteps.
func (i Interval) MustSignedSteps(sys *system.System, start spelling.Spelling) system.Step {
	return must(i.SignedSteps(sys, start))
}

// stepsFailure classifies why no interval of a size has a positional size.
type stepsFailure uint8

const (
	// stepsOK means that the interval exists.
	stepsOK stepsFailure = iota
	// stepsBaseRange means that the size of the basic kind does not fit in
	// an int.
	stepsBaseRange
	// stepsRange means that the deviation or the alteration value of the end
	// name is out of range.
	stepsRange
	// stepsInadmissible means that the end name is not admissible.
	stepsInadmissible
)

// intervalForSteps returns the ascending interval of size n >= 1 with the
// positional size s in the valid system sys from the optional start name,
// which is valid and admissible when given; without it sys is regular.
func intervalForSteps(sys *system.System, n int, s system.Step, start spelling.Spelling) (Interval, stepsFailure) {
	nm1 := int64(n - 1)
	q, r := nm1/7, int(nm1%7)
	c := int64(sys.Chromatic())
	l := letter.C
	if start.IsValid() {
		l = start.Letter()
	}
	base := mul64(q, int64(sys.Classes())).add(i128Of(stepsAfter(sys, l, r)))
	if _, ok := base.int64(); !ok {
		return Interval{}, stepsBaseRange
	}
	diff := i128Of(int64(s)).sub(base)
	if !start.IsValid() {
		d, ok := alterationOf(diff, c)
		if !ok {
			return Interval{}, stepsRange
		}
		return Interval{size: n, dev: d}, stepsOK
	}
	d, ok := alterationOf(diff.add(i128Of(delta(l, r)*c)), c)
	if !ok {
		return Interval{}, stepsRange
	}
	diff64, ok := diff.int64()
	if !ok {
		return Interval{}, stepsRange
	}
	px, qx := fractionOf(start.Alteration())
	az, ok := alterationOf(mul64(diff64, qx).add(mul64(px, c)), qx*c)
	if !ok {
		return Interval{}, stepsRange
	}
	if !admits(sys, spelling.MustNew(l.Shift(r), az)) {
		return Interval{}, stepsInadmissible
	}
	return Interval{size: n, dev: d}, stepsOK
}

// FromSteps returns the ascending interval of size n with the positional size
// s in the system sys: its deviation is the exact quotient (s - size of the
// basic kind) / c and can be fractional (in the 24-position system the third
// of size 7 is the neutral third). The start name is the lower name; it is
// optional only in a regular system. With a start name x, the name reached
// from x by the result must be admissible in sys.
//
// Errors, in this order: ErrUnset when sys is nil or unset; ErrInvalidArgument
// for n < 1 and for a unison with s < 0; ErrStartRequired when no start name
// is given and sys is not regular; system.ErrInadmissible when the start name
// is not admissible; ErrOutOfRange when the size of the basic kind does not
// fit in an int or when the deviation or the alteration value of the reached
// name is out of range (then with alteration.ErrOutOfRange);
// system.ErrInadmissible when the reached name is not admissible.
func FromSteps(sys *system.System, size int, steps system.Step, start spelling.Spelling) (Interval, error) {
	const op = "FromSteps"
	if err := checkSystem(op, sys); err != nil {
		return Interval{}, err
	}
	if size < 1 {
		return Interval{}, detailError(op, ErrInvalidArgument, "size less than 1")
	}
	if size == 1 && steps < 0 {
		return Interval{}, detailError(op, ErrInvalidArgument, "diminished unison")
	}
	sc := sizeCalc{sys: sys, start: start}
	if err := sc.checkStart(op); err != nil {
		return Interval{}, err
	}
	if err := sc.checkStartAdmitted(op); err != nil {
		return Interval{}, err
	}
	i, fail := intervalForSteps(sys, size, steps, start)
	switch fail {
	case stepsBaseRange:
		return Interval{}, detailError(op, ErrOutOfRange, "size of the basic kind does not fit in an int")
	case stepsRange:
		return Interval{}, alterationRangeError(op)
	case stepsInadmissible:
		return Interval{}, inadmissibleError(op, "name reached from the start name")
	}
	return i, nil
}

// MustFromSteps is like FromSteps but panics if FromSteps returns an error.
// The panic value is the error returned by FromSteps.
func MustFromSteps(sys *system.System, size int, steps system.Step, start spelling.Spelling) Interval {
	return must(FromSteps(sys, size, steps, start))
}

// EnharmonicallyEqual reports whether i from the start name iStart and j from
// the start name jStart have the same positional size in the system sys.
// In the traditional system A4 and d5 are enharmonically equal; in the
// 19-position system they are not.
//
// The errors are those of Steps for both intervals, checked category by
// category: ErrUnset for sys, i, and j; ErrStartRequired; the start names;
// ErrFractionalSteps; ErrOutOfRange for the reached names; the reached names;
// ErrOutOfRange for the sizes.
func EnharmonicallyEqual(sys *system.System, i Interval, iStart spelling.Spelling, j Interval, jStart spelling.Spelling) (bool, error) {
	const op = "EnharmonicallyEqual"
	if err := checkSystemInterval(op, sys, i, j); err != nil {
		return false, err
	}
	a := sizeCalc{sys: sys, i: i, start: iStart}
	b := sizeCalc{sys: sys, i: j, start: jStart}
	for k := range sizeStages {
		if err := a.stage(op, k); err != nil {
			return false, err
		}
		if err := b.stage(op, k); err != nil {
			return false, err
		}
	}
	x, err := a.size(op)
	if err != nil {
		return false, err
	}
	y, err := b.size(op)
	if err != nil {
		return false, err
	}
	return x == y, nil
}

// MustEnharmonicallyEqual is like EnharmonicallyEqual but panics if
// EnharmonicallyEqual returns an error. The panic value is the error returned
// by EnharmonicallyEqual.
func MustEnharmonicallyEqual(sys *system.System, i Interval, iStart spelling.Spelling, j Interval, jStart spelling.Spelling) bool {
	return must(EnharmonicallyEqual(sys, i, iStart, j, jStart))
}

// remotenessAtMost reports whether the remoteness of the ascending interval i
// from the basic kinds is at most k >= 0: |d| in the perfect group; in the
// major-minor group d for d > 0, -1 - d for d < -1, and 0 otherwise. For
// whole qualities the remoteness is the multiplicity.
func remotenessAtMost(i Interval, k int) bool {
	if k > alteration.MaxMagnitude {
		return true
	}
	limit := alteration.MustFromInt(k)
	d := i.dev
	if perfectSize(i.size) {
		if d.Sign() < 0 {
			d = alteration.Natural().MustSub(d)
		}
		return d.Compare(limit) <= 0
	}
	switch {
	case d.Sign() > 0:
		return d.Compare(limit) <= 0
	case d.Compare(alteration.Flat()) < 0:
		// -1 - d > 0 and in range: d >= -MaxMagnitude.
		return alteration.Flat().MustSub(d).Compare(limit) <= 0
	default:
		return true
	}
}

// stepsWindow returns the sequence of the ascending intervals with the
// positional size s and remoteness at most k in the valid system sys from
// the optional start name. Only the sizes whose basic kind is within
// (k + 1)·c steps of s are visited, starting from a size found by division.
func stepsWindow(sys *system.System, s system.Step, k int, start spelling.Spelling) iter.Seq[Interval] {
	n := int64(sys.Classes())
	c := int64(sys.Chromatic())
	span := mul64(int64(min(k, alteration.MaxMagnitude+1))+1, c)
	hi := i128Of(int64(s)).add(span)
	lo := i128Of(int64(s)).sub(span)
	q0, _ := lo.sub(i128Of(c)).floorDivMod(n)
	q0 = q0.sub(i128Of(1))
	const maxQ = (MaxSize - 1) / 7
	first, ok := q0.int64()
	switch {
	case q0.sign() < 0:
		first = 0
	case !ok || first > maxQ:
		return func(func(Interval) bool) {}
	}
	return func(yield func(Interval) bool) {
		for q := first; q <= maxQ; q++ {
			if mul64(q, n).sub(i128Of(c)).cmp(hi) > 0 {
				return
			}
			for r := range int64(7) {
				nm1 := 7*q + r
				if nm1 > MaxSize-1 {
					return
				}
				i, fail := intervalForSteps(sys, int(nm1)+1, s, start)
				if fail != stepsOK || (i.size == 1 && i.dev.Sign() < 0) || !remotenessAtMost(i, k) {
					continue
				}
				if !yield(i) {
					return
				}
			}
		}
	}
}

// AllFromSteps returns the ascending intervals with the positional size s in
// the system sys whose remoteness from the basic kinds is at most
// maxMultiplicity, in the order of size. The remoteness is |d| in the perfect
// group and, in the major-minor group, d for d > 0, -1 - d for d < -1, and 0
// otherwise; for whole qualities it is the multiplicity. Unisons narrower than
// perfect are not included. The start name is the lower name; it is optional
// only in a regular system. With a start name, only intervals whose upper name
// is representable and admissible in sys are included; without it, names are
// not checked. In the traditional system the size 6 with the limit 2 gives
// AA3, A4, d5, and dd6.
//
// All errors are reported before the sequence is returned: ErrUnset when sys
// is nil or unset; ErrInvalidArgument for maxMultiplicity < 0;
// ErrStartRequired when no start name is given and sys is not regular;
// system.ErrInadmissible when the start name is not admissible. The sequence
// is finite and can be iterated many times and concurrently; each iteration
// gives the same intervals.
func AllFromSteps(sys *system.System, steps system.Step, maxMultiplicity int, start spelling.Spelling) (iter.Seq[Interval], error) {
	const op = "AllFromSteps"
	if err := checkSystem(op, sys); err != nil {
		return nil, err
	}
	if maxMultiplicity < 0 {
		return nil, detailError(op, ErrInvalidArgument, "negative limit of multiplicity")
	}
	sc := sizeCalc{sys: sys, start: start}
	if err := sc.checkStart(op); err != nil {
		return nil, err
	}
	if err := sc.checkStartAdmitted(op); err != nil {
		return nil, err
	}
	return stepsWindow(sys, steps, maxMultiplicity, start), nil
}

// MustAllFromSteps is like AllFromSteps but panics if AllFromSteps returns an
// error. The panic value is the error returned by AllFromSteps.
func MustAllFromSteps(sys *system.System, steps system.Step, maxMultiplicity int, start spelling.Spelling) iter.Seq[Interval] {
	return must(AllFromSteps(sys, steps, maxMultiplicity, start))
}

// Enharmonics returns the ascending intervals enharmonically equal to i in the
// system sys: AllFromSteps for the positional size of i from its lower name,
// which is the start name for an ascending i and the name reached from the
// start name for a descending one. Without a start name, no name is used.
// The ascending form of i is included when it passes the limit.
//
// Errors, in this order: ErrUnset when sys is nil or unset or when i is unset;
// ErrInvalidArgument for maxMultiplicity < 0; then the errors of Steps.
func (i Interval) Enharmonics(sys *system.System, maxMultiplicity int, start spelling.Spelling) (iter.Seq[Interval], error) {
	const op = "Enharmonics"
	if err := checkSystemInterval(op, sys, i); err != nil {
		return nil, err
	}
	if maxMultiplicity < 0 {
		return nil, detailError(op, ErrInvalidArgument, "negative limit of multiplicity")
	}
	s, end, err := positionalSize(op, sys, i, start)
	if err != nil {
		return nil, err
	}
	lower := start
	if i.down {
		lower = end
	}
	return stepsWindow(sys, s, maxMultiplicity, lower), nil
}

// MustEnharmonics is like Enharmonics but panics if Enharmonics returns an
// error. The panic value is the error returned by Enharmonics.
func (i Interval) MustEnharmonics(sys *system.System, maxMultiplicity int, start spelling.Spelling) iter.Seq[Interval] {
	return must(i.Enharmonics(sys, maxMultiplicity, start))
}

// Respelling is one pair of Respellings: the start and end pitches and the
// interval from From to To.
type Respelling struct {
	From, To pitch.Pitch
	Interval Interval
}

// sideOf returns the pitches of one side of Respellings: p itself when it is
// kept, or its enharmonic pitches within the limit when it is replaced.
func sideOf(sys *system.System, p pitch.Pitch, limit alteration.Alteration, replace bool) iter.Seq[pitch.Pitch] {
	if !replace {
		return func(yield func(pitch.Pitch) bool) { yield(p) }
	}
	seq, err := sys.EnharmonicPitches(p, limit)
	if err != nil {
		// The checks before make this impossible.
		return func(func(pitch.Pitch) bool) {}
	}
	return seq
}

// sideStats holds the extreme letter coordinates and alteration values of the
// pitches of one side, with the pitches that have them.
type sideStats struct {
	minU, maxU i128
	minA, maxA alteration.Alteration
	extremes   [4]pitch.Pitch
	dens       []int64
	manyDens   bool
}

// maxSideDenominators bounds the number of distinct denominators collected
// for a side; beyond it the quick check gives up.
const maxSideDenominators = 16

// collectSide computes the statistics of a side and reports a written octave
// that does not fit in an int.
func collectSide(op string, sys *system.System, seq iter.Seq[pitch.Pitch]) (sideStats, error) {
	var st sideStats
	first := true
	b := sys.Boundary()
	for p := range seq {
		w, err := sys.WrittenOctave(p)
		if err != nil {
			return sideStats{}, rangeError(op, err)
		}
		wn, err := w.ASPN()
		if err != nil {
			return sideStats{}, rangeError(op, err)
		}
		s := p.Spelling()
		u := mul64(int64(wn), 7).add(i128Of(int64(b.DistanceUp(s.Letter()))))
		a := s.Alteration()
		if first || u.cmp(st.minU) < 0 {
			st.minU, st.extremes[0] = u, p
		}
		if first || u.cmp(st.maxU) > 0 {
			st.maxU, st.extremes[1] = u, p
		}
		if first || a.Compare(st.minA) < 0 {
			st.minA, st.extremes[2] = a, p
		}
		if first || a.Compare(st.maxA) > 0 {
			st.maxA, st.extremes[3] = a, p
		}
		first = false
		_, den := fractionOf(a)
		if !st.manyDens && !containsDen(st.dens, den) {
			if len(st.dens) == maxSideDenominators {
				st.manyDens = true
			} else {
				st.dens = append(st.dens, den)
			}
		}
	}
	return st, nil
}

// containsDen reports whether dens contains d.
func containsDen(dens []int64, d int64) bool {
	for _, x := range dens {
		if x == d {
			return true
		}
	}
	return false
}

// pairsSafe reports whether every pair of the sides certainly has a
// representable interval: the letter distances, the alteration differences,
// and their denominators are within the limits.
func pairsSafe(x, y sideStats) bool {
	maxL := y.maxU.sub(x.minU)
	if l := x.maxU.sub(y.minU); l.cmp(maxL) > 0 {
		maxL = l
	}
	if maxL.cmp(i128Of(MaxSize-1)) > 0 {
		return false
	}
	// |d| <= |a_y - a_x| + 1 for every pair.
	for _, pair := range [2][2]alteration.Alteration{{y.maxA, x.minA}, {x.maxA, y.minA}} {
		d, ok := combine(1, 1, pair[0], -1, pair[1])
		if !ok || d.Compare(alteration.MustFromInt(alteration.MaxMagnitude)) > 0 {
			return false
		}
	}
	if x.manyDens || y.manyDens {
		return false
	}
	for _, dx := range x.dens {
		for _, dy := range y.dens {
			if dx/gcd64(dx, dy)*dy > alteration.MaxDenominator {
				return false
			}
		}
	}
	return true
}

// checkPairs reports the first pair of the sides without a representable
// interval. The pairs of extreme pitches are checked first.
func checkPairs(op string, sys *system.System, xs, ys iter.Seq[pitch.Pitch], x, y sideStats) error {
	for _, f := range x.extremes {
		for _, t := range y.extremes {
			if _, err := pitchInterval(op, sys, f, t); err != nil {
				return err
			}
		}
	}
	for f := range xs {
		for t := range ys {
			if _, err := pitchInterval(op, sys, f, t); err != nil {
				return err
			}
		}
	}
	return nil
}

// Respellings returns the enharmonic respellings of the interval from the
// written pitch x to the written pitch y in the system sys: the pairs of
// pitches at the same positions as x and y, with the interval of each pair.
// The replaced pitch runs over system.System.EnharmonicPitches with the limit
// of alteration values; the kept pitch stays. The pairs are in the order of
// the start pitch, then of the end pitch, each in the order of
// EnharmonicPitches; the original pair is included when its replaced pitches
// pass the limit. In the traditional system with the limit 2, replacing the
// end pitch of C4 to E4 gives C4 to D double sharp 4 (AA2), C4 to E4 (M3),
// and C4 to F flat 4 (d4).
//
// All errors are reported before the sequence is returned: ErrUnset when sys
// is nil or unset, when x or y is unset, or when which is not ReplaceFrom,
// ReplaceTo, or ReplaceBoth; ErrInvalidArgument for a negative limit;
// system.ErrInadmissible when the name of x, then of y, is not admissible;
// ErrOutOfRange when the interval of some pair is not representable. The
// sequence can be iterated many times and concurrently; each iteration gives
// the same pairs.
func Respellings(sys *system.System, x, y pitch.Pitch, limit alteration.Alteration, which Replace) (iter.Seq[Respelling], error) {
	const op = "Respellings"
	if err := checkSystem(op, sys); err != nil {
		return nil, err
	}
	if err := checkPitches(op, x, y); err != nil {
		return nil, err
	}
	if which != ReplaceFrom && which != ReplaceTo && which != ReplaceBoth {
		return nil, detailError(op, ErrUnset, "choice of the replaced pitches")
	}
	if limit.Sign() < 0 {
		return nil, detailError(op, ErrInvalidArgument, "negative limit of alteration")
	}
	if err := checkAdmitted(op, sys, inadmissibleError, "name", x.Spelling(), y.Spelling()); err != nil {
		return nil, err
	}
	xs := sideOf(sys, x, limit, which != ReplaceTo)
	ys := sideOf(sys, y, limit, which != ReplaceFrom)
	sx, err := collectSide(op, sys, xs)
	if err != nil {
		return nil, err
	}
	sy, err := collectSide(op, sys, ys)
	if err != nil {
		return nil, err
	}
	if !pairsSafe(sx, sy) {
		if err := checkPairs(op, sys, xs, ys, sx, sy); err != nil {
			return nil, err
		}
	}
	return func(yield func(Respelling) bool) {
		for f := range xs {
			for t := range ys {
				i, _ := pitchInterval(op, sys, f, t)
				if !yield(Respelling{From: f, To: t, Interval: i}) {
					return
				}
			}
		}
	}, nil
}

// MustRespellings is like Respellings but panics if Respellings returns an
// error. The panic value is the error returned by Respellings.
func MustRespellings(sys *system.System, x, y pitch.Pitch, limit alteration.Alteration, which Replace) iter.Seq[Respelling] {
	return must(Respellings(sys, x, y, limit, which))
}
