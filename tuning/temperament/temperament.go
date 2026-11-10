package temperament

import (
	"fmt"
	"math"
	"math/bits"
	"slices"

	"github.com/go-muse/musex/cents"
)

const (
	// MaxStepsPerOctave is the maximum supported number of steps per octave (N_max = 65536).
	MaxStepsPerOctave = 65536

	// DefaultStepsPerOctave is the standard 12-tone number of steps per octave.
	DefaultStepsPerOctave = 12

	// DefaultPythagoreanMinFifth is the default starting fifth index k_min = -5 (Db..F#, wolf F#-Db at step 6->1)
	// for the symmetric Pythagorean preset.
	DefaultPythagoreanMinFifth = -5

	// EFlatPythagoreanMinFifth is the starting fifth index k_min = -3 (Eb..G#, wolf G#-Eb at step 8->3)
	// of the Eb..G# Pythagorean preset [PythagoreanEFlatGSharp].
	EFlatPythagoreanMinFifth = -3

	// DefaultMeantoneMinFifth is the default starting fifth index k_min = -3 (Eb..G#, wolf G#-Eb at step 8->3)
	// for regular meantone temperaments.
	DefaultMeantoneMinFifth = -3
)

// octaveCentsFloat is the size of the octave in cents as a float64.
const octaveCentsFloat = 1200.0

var (
	defaultFifthWolfThreshold = cents.MustFromInt(15)
	defaultThirdWolfThreshold = cents.MustFromInt(25)
)

// DefaultFifthWolfThreshold returns the default wolf deviation threshold for
// fifths (15 cents).
func DefaultFifthWolfThreshold() cents.Cents {
	return defaultFifthWolfThreshold
}

// DefaultThirdWolfThreshold returns the default wolf deviation threshold for
// major and minor thirds (25 cents).
func DefaultThirdWolfThreshold() cents.Cents {
	return defaultThirdWolfThreshold
}

// Ratio represents a positive frequency ratio Num/Den (p/q) for constructing a
// temperament from pure intervals.
type Ratio struct {
	// Num is the numerator p of the frequency ratio p/q (must be > 0).
	Num int
	// Den is the denominator q of the frequency ratio p/q (must be > 0).
	Den int
}

// Temperament represents an immutable musical temperament of dimension N
// (1 <= N <= MaxStepsPerOctave) as a strictly increasing table of intra-octave
// step offsets in cents: 0 = T(0) < T(1) < ... < T(N-1) < 1200 cents.
//
// The zero value Temperament{} is invalid (IsValid() == false,
// StepsPerOctave() == 0).
type Temperament struct {
	name    string
	offsets []cents.Cents
}

// makeTemperament wraps an already validated offsets table that the
// Temperament takes ownership of. The table is never modified afterwards.
func makeTemperament(name string, ownedOffsets []cents.Cents) Temperament {
	return Temperament{
		name:    name,
		offsets: ownedOffsets,
	}
}

// New creates a temperament with the given name and intra-octave step offsets
// in cents. It validates 1 <= len(offsets) <= MaxStepsPerOctave and strict
// monotonicity 0 = T(0) < T(1) < ... < T(N-1) < 1200. It copies the input slice
// so subsequent mutations by the caller do not affect the Temperament.
func New(name string, offsets []cents.Cents) (Temperament, error) {
	return newChecked("New", name, slices.Clone(offsets))
}

// MustNew is like [New] but panics on error.
func MustNew(name string, offsets []cents.Cents) Temperament {
	return must(New(name, offsets))
}

// newChecked validates ownedOffsets in place without cloning again and takes
// ownership of the slice; op names the public operation in error messages.
func newChecked(op, name string, ownedOffsets []cents.Cents) (Temperament, error) {
	n := len(ownedOffsets)
	if n < 1 || n > MaxStepsPerOctave {
		return Temperament{}, errorf(ErrInvalidArgument, "%s: step count %d out of range [1, %d]", op, n, MaxStepsPerOctave)
	}
	if !ownedOffsets[0].Equal(cents.Zero()) {
		return Temperament{}, errorf(ErrInvalidArgument, "%s: initial offset T(0) must be 0 cents, got %v", op, ownedOffsets[0])
	}
	for m := 1; m < n; m++ {
		if ownedOffsets[m].Compare(ownedOffsets[m-1]) <= 0 {
			return Temperament{}, errorf(ErrInvalidArgument, "%s: offsets must be strictly increasing at step %d", op, m)
		}
	}
	if ownedOffsets[n-1].Compare(cents.Octave()) >= 0 {
		return Temperament{}, errorf(ErrInvalidArgument, "%s: offset at step %d must be less than 1200 cents", op, n-1)
	}
	return makeTemperament(name, ownedOffsets), nil
}

// FromOffsets is an alias for [New].
func FromOffsets(name string, offsets []cents.Cents) (Temperament, error) {
	return New(name, offsets)
}

// MustFromOffsets is an alias for [MustNew].
func MustFromOffsets(name string, offsets []cents.Cents) Temperament {
	return must(FromOffsets(name, offsets))
}

// FromFloats creates a temperament by converting each float64 offset via
// [cents.FromFloat] and validating 1 <= len(offsets) <= MaxStepsPerOctave and
// strict monotonicity 0 = T(0) < T(1) < ... < T(N-1) < 1200.
func FromFloats(name string, offsets []float64) (Temperament, error) {
	n := len(offsets)
	if n < 1 || n > MaxStepsPerOctave {
		return Temperament{}, errorf(ErrInvalidArgument, "FromFloats: step count %d out of range [1, %d]", n, MaxStepsPerOctave)
	}
	// Reject non-finite values and values that cannot lie in [0, 1200) cents
	// after conversion before converting, so that a value far outside the
	// octave is reported as an invalid argument rather than as a cents range
	// error (priority of ErrInvalidArgument over ErrOutOfRange). The table
	// invariant itself, including T(0) = 0, is checked on the converted cent
	// values, so a float that rounds to 0 on the cents grid is a valid root.
	for m, x := range offsets {
		switch {
		case math.IsNaN(x) || math.IsInf(x, 0):
			return Temperament{}, errorf(ErrInvalidArgument, "FromFloats: non-finite offset %v at step %d", x, m)
		case x <= -octaveCentsFloat || x >= octaveCentsFloat:
			return Temperament{}, errorf(ErrInvalidArgument, "FromFloats: offset %v at step %d outside (-1200, 1200) cents", x, m)
		}
	}
	conv := make([]cents.Cents, n)
	for m, x := range offsets {
		c, err := cents.FromFloat(x)
		if err != nil {
			return Temperament{}, wrapCentsError("FromFloats", err)
		}
		conv[m] = c
	}
	return newChecked("FromFloats", name, conv)
}

// MustFromFloats is like [FromFloats] but panics on error.
func MustFromFloats(name string, offsets []float64) Temperament {
	return must(FromFloats(name, offsets))
}

// FromRatios creates a temperament of dimension N = len(ratios) from pure
// frequency ratios p_m/q_m in [1, 2). It validates argument preconditions first
// (returning [ErrInvalidArgument] if N is not in [1, MaxStepsPerOctave],
// Num <= 0, Den <= 0, p_0 != q_0, p_m/q_m not in [1, 2), or non-increasing
// ratios), and then converts each ratio via [cents.FromRatio] (returning
// [ErrPrimeLimit] if any reduced ratio has a prime factor > 31).
func FromRatios(name string, ratios []Ratio) (Temperament, error) {
	n := len(ratios)
	if n < 1 || n > MaxStepsPerOctave {
		return Temperament{}, errorf(ErrInvalidArgument, "FromRatios: step count %d out of range [1, %d]", n, MaxStepsPerOctave)
	}
	for m, r := range ratios {
		if r.Num <= 0 || r.Den <= 0 {
			return Temperament{}, errorf(ErrInvalidArgument, "FromRatios: non-positive ratio %d:%d at step %d", r.Num, r.Den, m)
		}
		if m == 0 {
			if r.Num != r.Den {
				return Temperament{}, errorf(ErrInvalidArgument, "FromRatios: initial ratio at step 0 must equal 1:1, got %d:%d", r.Num, r.Den)
			}
			continue
		}
		if r.Num < r.Den || r.Num-r.Den >= r.Den {
			return Temperament{}, errorf(ErrInvalidArgument, "FromRatios: ratio %d:%d at step %d outside [1:1, 2:1)", r.Num, r.Den, m)
		}
		prev := ratios[m-1]
		hi1, lo1 := bits.Mul64(uint64(prev.Num), uint64(r.Den))
		hi2, lo2 := bits.Mul64(uint64(r.Num), uint64(prev.Den))
		if hi1 > hi2 || (hi1 == hi2 && lo1 >= lo2) {
			return Temperament{}, errorf(ErrInvalidArgument, "FromRatios: ratios must be strictly increasing at step %d", m)
		}
	}
	conv := make([]cents.Cents, n)
	conv[0] = cents.Zero()
	for m := 1; m < n; m++ {
		c, err := cents.FromRatio(ratios[m].Num, ratios[m].Den)
		if err != nil {
			return Temperament{}, wrapCentsError("FromRatios", err)
		}
		conv[m] = c
	}
	return makeTemperament(name, conv), nil
}

// MustFromRatios is like [FromRatios] but panics on error.
func MustFromRatios(name string, ratios []Ratio) Temperament {
	return must(FromRatios(name, ratios))
}

// IsValid reports whether t is an initialized temperament with
// StepsPerOctave() >= 1.
func (t Temperament) IsValid() bool {
	return len(t.offsets) > 0
}

// Name returns the name of the temperament (or "" for the zero value).
func (t Temperament) Name() string {
	return t.name
}

// StepsPerOctave returns the number of steps per octave N (or 0 for the zero
// value).
func (t Temperament) StepsPerOctave() int {
	return len(t.offsets)
}

// Offsets returns a defensive copy of the N intra-octave step offsets T(0..N-1)
// in cents (or nil for the zero value).
func (t Temperament) Offsets() []cents.Cents {
	if !t.IsValid() {
		return nil
	}
	return slices.Clone(t.offsets)
}

// WithName returns a copy of t with its name replaced by name. It returns
// [ErrInvalidArgument] if t is the zero value.
func (t Temperament) WithName(name string) (Temperament, error) {
	if !t.IsValid() {
		return Temperament{}, errorf(ErrInvalidArgument, "WithName: invalid temperament")
	}
	return Temperament{
		name:    name,
		offsets: t.offsets,
	}, nil
}

// MustWithName is like [Temperament.WithName] but panics if t is the zero
// value.
func (t Temperament) MustWithName(name string) Temperament {
	return must(t.WithName(name))
}

// Offset returns the intra-octave offset T(step) in [0, 1200) cents for step in
// [0, N-1]. It returns [ErrInvalidArgument] if t is invalid or step is outside
// [0, N-1].
func (t Temperament) Offset(step int) (cents.Cents, error) {
	if !t.IsValid() {
		return cents.Cents{}, errorf(ErrInvalidArgument, "Offset: invalid temperament")
	}
	if step < 0 || step >= len(t.offsets) {
		return cents.Cents{}, errorf(ErrInvalidArgument, "Offset(%d): step out of range [0, %d)", step, len(t.offsets))
	}
	return t.offsets[step], nil
}

// MustOffset is like [Temperament.Offset] but panics on error.
func (t Temperament) MustOffset(step int) cents.Cents {
	return must(t.Offset(step))
}

// SplitStep decomposes any integer step k in Z by Euclidean division
// k = N*octaves + m (0 <= m < N), returning the octave number
// octaves = floor(k/N) and the intra-octave offset T(m) in [0, 1200) cents. It
// returns [ErrInvalidArgument] if t is invalid.
func (t Temperament) SplitStep(step int) (octaves int, offset cents.Cents, err error) {
	if !t.IsValid() {
		return 0, cents.Cents{}, errorf(ErrInvalidArgument, "SplitStep: invalid temperament")
	}
	q, m := t.split(step)
	return q, t.offsets[m], nil
}

// split performs the Euclidean division step = N*q + m (0 <= m < N) for a
// valid t.
func (t Temperament) split(step int) (q, m int) {
	n := len(t.offsets)
	q = step / n
	m = step % n
	if m < 0 {
		m += n
		q--
	}
	return q, m
}

// maxOctaves is the number of whole octaves in the range of package cents.
const maxOctaves = cents.MaxMagnitude / 1200

// cumulativeInRange reports whether c(k) = 1200*q + T(m) lies within the range
// of package cents. Since 0 <= T(m) < 1200 and T(0) = 0, this depends only on
// q and m, so the check needs no cents arithmetic.
func cumulativeInRange(q, m int) bool {
	return q >= -maxOctaves && (q < maxOctaves || (q == maxOctaves && m == 0))
}

// octaves returns 1200*q cents for |q| <= 2*maxOctaves.
func octaves(op string, q int) (cents.Cents, error) {
	c, err := cents.FromInt(1200 * q)
	if err != nil {
		return cents.Cents{}, wrapCentsError(op, err)
	}
	return c, nil
}

// MustSplitStep is like [Temperament.SplitStep] but panics on error.
func (t Temperament) MustSplitStep(step int) (octaves int, offset cents.Cents) {
	return must2(t.SplitStep(step))
}

// CumulativeOffset returns the multi-octave offset
// c(step) = 1200*floor(step/N) + T(step mod N) in cents for any integer step in
// Z. It returns [ErrInvalidArgument] if t is invalid, or [ErrOutOfRange] if the
// result exceeds the range of package cents.
func (t Temperament) CumulativeOffset(step int) (cents.Cents, error) {
	const op = "CumulativeOffset"
	if !t.IsValid() {
		return cents.Cents{}, errorf(ErrInvalidArgument, "%s: invalid temperament", op)
	}
	q, m := t.split(step)
	if !cumulativeInRange(q, m) {
		return cents.Cents{}, cumulativeRangeError(op, q, m)
	}
	return t.cumulative(op, q, m)
}

// cumulative computes c(k) = 1200*q + T(m) for an in-range (q, m).
func (t Temperament) cumulative(op string, q, m int) (cents.Cents, error) {
	if q == 0 {
		return t.offsets[m], nil
	}
	shift, err := octaves(op, q)
	if err != nil {
		return cents.Cents{}, err
	}
	if m == 0 {
		return shift, nil
	}
	res, err := shift.Add(t.offsets[m])
	if err != nil {
		return cents.Cents{}, wrapCentsError(op, err)
	}
	return res, nil
}

// cumulativeRangeError returns the error for c(k) = 1200*q + T(m) outside the
// range of package cents. It wraps both [ErrOutOfRange] and
// [cents.ErrOutOfRange], as the error of the equivalent cents arithmetic would.
func cumulativeRangeError(op string, q, m int) error {
	return fmt.Errorf("%w: %s: 1200*%d + T(%d) cents outside [-%d, %d]: %w",
		ErrOutOfRange, op, q, m, cents.MaxMagnitude, cents.MaxMagnitude, cents.ErrOutOfRange)
}

// MustCumulativeOffset is like [Temperament.CumulativeOffset] but panics on
// error.
func (t Temperament) MustCumulativeOffset(step int) cents.Cents {
	return must(t.CumulativeOffset(step))
}

// DirectedInterval returns the signed multi-octave interval
// I_dir(fromStep, toStep) = c(toStep) - c(fromStep) for any integer steps
// fromStep, toStep in Z. It returns [ErrInvalidArgument] if t is invalid, or
// [ErrOutOfRange] if intermediate or final values exceed the range of package
// cents.
func (t Temperament) DirectedInterval(fromStep, toStep int) (cents.Cents, error) {
	const op = "DirectedInterval"
	if !t.IsValid() {
		return cents.Cents{}, errorf(ErrInvalidArgument, "%s: invalid temperament", op)
	}
	qFrom, mFrom := t.split(fromStep)
	if !cumulativeInRange(qFrom, mFrom) {
		return cents.Cents{}, cumulativeRangeError(op, qFrom, mFrom)
	}
	qTo, mTo := t.split(toStep)
	if !cumulativeInRange(qTo, mTo) {
		return cents.Cents{}, cumulativeRangeError(op, qTo, mTo)
	}
	// c(to) - c(from) = 1200*(qTo - qFrom) + (T(mTo) - T(mFrom)), where
	// |qTo - qFrom| <= 2*maxOctaves, so the octave part is an exact int.
	dq := qTo - qFrom
	switch {
	case mTo == mFrom:
		if dq == 0 {
			return cents.Zero(), nil
		}
		return octaves(op, dq)
	case mFrom == 0 && dq == 0:
		return t.offsets[mTo], nil
	}
	within, err := t.offsets[mTo].Sub(t.offsets[mFrom])
	if err != nil {
		return cents.Cents{}, wrapCentsError(op, err)
	}
	if dq == 0 {
		return within, nil
	}
	shift, err := octaves(op, dq)
	if err != nil {
		return cents.Cents{}, err
	}
	res, err := shift.Add(within)
	if err != nil {
		return cents.Cents{}, wrapCentsError(op, err)
	}
	return res, nil
}

// MustDirectedInterval is like [Temperament.DirectedInterval] but panics on
// error.
func (t Temperament) MustDirectedInterval(fromStep, toStep int) cents.Cents {
	return must(t.DirectedInterval(fromStep, toStep))
}

// CyclicInterval returns the ascending intra-octave interval
// I_cyc(fromStep, toStep) in [0, 1200) cents between two intra-octave positions
// fromStep, toStep in [0, N-1]. It returns [ErrInvalidArgument] if t is invalid
// or if fromStep or toStep is outside [0, N-1].
func (t Temperament) CyclicInterval(fromStep, toStep int) (cents.Cents, error) {
	if !t.IsValid() {
		return cents.Cents{}, errorf(ErrInvalidArgument, "CyclicInterval: invalid temperament")
	}
	n := len(t.offsets)
	if fromStep < 0 || fromStep >= n || toStep < 0 || toStep >= n {
		return cents.Cents{}, errorf(ErrInvalidArgument, "CyclicInterval(%d, %d): step out of range [0, %d)", fromStep, toStep, n)
	}
	return t.cyclicInterval(fromStep, toStep)
}

// cyclicInterval computes I_cyc(fromStep, toStep) for already validated
// intra-octave positions.
func (t Temperament) cyclicInterval(fromStep, toStep int) (cents.Cents, error) {
	if fromStep == toStep {
		return cents.Zero(), nil
	}
	if fromStep == 0 {
		return t.offsets[toStep], nil
	}
	if toStep > fromStep {
		res, err := t.offsets[toStep].Sub(t.offsets[fromStep])
		if err != nil {
			return cents.Cents{}, wrapCentsError("CyclicInterval", err)
		}
		return res, nil
	}
	shifted, err := cents.Octave().Add(t.offsets[toStep])
	if err != nil {
		return cents.Cents{}, wrapCentsError("CyclicInterval", err)
	}
	res, err := shifted.Sub(t.offsets[fromStep])
	if err != nil {
		return cents.Cents{}, wrapCentsError("CyclicInterval", err)
	}
	return res, nil
}

// MustCyclicInterval is like [Temperament.CyclicInterval] but panics on error.
func (t Temperament) MustCyclicInterval(fromStep, toStep int) cents.Cents {
	return must(t.CyclicInterval(fromStep, toStep))
}

// StepSizes returns a new slice of N positive elementary step sizes
// delta_T(m) = T(m+1) - T(m) for m in 0..N-1 (where T(N) = 1200 cents). It
// returns [ErrInvalidArgument] if t is invalid.
func (t Temperament) StepSizes() ([]cents.Cents, error) {
	if !t.IsValid() {
		return nil, errorf(ErrInvalidArgument, "StepSizes: invalid temperament")
	}
	n := len(t.offsets)
	sizes := make([]cents.Cents, n)
	for m := range n - 1 {
		d, err := t.offsets[m+1].Sub(t.offsets[m])
		if err != nil {
			return nil, wrapCentsError("StepSizes", err)
		}
		sizes[m] = d
	}
	last, err := cents.Octave().Sub(t.offsets[n-1])
	if err != nil {
		return nil, wrapCentsError("StepSizes", err)
	}
	sizes[n-1] = last
	return sizes, nil
}

// MustStepSizes is like [Temperament.StepSizes] but panics on error.
func (t Temperament) MustStepSizes() []cents.Cents {
	return must(t.StepSizes())
}

// MinStepSize returns the minimum elementary step size min_m delta_T(m). It
// returns [ErrInvalidArgument] if t is invalid.
func (t Temperament) MinStepSize() (cents.Cents, error) {
	sizes, err := t.StepSizes()
	if err != nil {
		return cents.Cents{}, err
	}
	minVal := sizes[0]
	for _, s := range sizes[1:] {
		if s.Compare(minVal) < 0 {
			minVal = s
		}
	}
	return minVal, nil
}

// MustMinStepSize is like [Temperament.MinStepSize] but panics on error.
func (t Temperament) MustMinStepSize() cents.Cents {
	return must(t.MinStepSize())
}

// MaxStepSize returns the maximum elementary step size max_m delta_T(m). It
// returns [ErrInvalidArgument] if t is invalid.
func (t Temperament) MaxStepSize() (cents.Cents, error) {
	sizes, err := t.StepSizes()
	if err != nil {
		return cents.Cents{}, err
	}
	maxVal := sizes[0]
	for _, s := range sizes[1:] {
		if s.Compare(maxVal) > 0 {
			maxVal = s
		}
	}
	return maxVal, nil
}

// MustMaxStepSize is like [Temperament.MaxStepSize] but panics on error.
func (t Temperament) MustMaxStepSize() cents.Cents {
	return must(t.MaxStepSize())
}

// Rotate returns a new temperament of the same dimension N obtained by
// cyclically shifting the root to step r0 = ((steps mod N) + N) mod N:
// T^(steps)(m) = CyclicInterval(r0, (r0+m) mod N). It preserves the original
// name t.Name() unchanged (use [Temperament.WithName] to rename the rotated
// temperament). It returns [ErrInvalidArgument] if t is invalid.
func (t Temperament) Rotate(steps int) (Temperament, error) {
	if !t.IsValid() {
		return Temperament{}, errorf(ErrInvalidArgument, "Rotate: invalid temperament")
	}
	n := len(t.offsets)
	r0 := steps % n
	if r0 < 0 {
		r0 += n
	}
	if r0 == 0 {
		return t, nil
	}
	rotated := make([]cents.Cents, n)
	rotated[0] = cents.Zero()
	for m := 1; m < n; m++ {
		off, err := t.cyclicInterval(r0, (r0+m)%n)
		if err != nil {
			return Temperament{}, err
		}
		rotated[m] = off
	}
	return makeTemperament(t.name, rotated), nil
}

// MustRotate is like [Temperament.Rotate] (preserving t.Name() unchanged) but
// panics on error.
func (t Temperament) MustRotate(steps int) Temperament {
	return must(t.Rotate(steps))
}

// Equal reports whether t and other have the same dimension N and exact equal
// step offsets T_1(m).Equal(T_2(m)) for all m in 0..N-1 (ignoring Name). Two
// zero-value temperaments are equal to each other.
func (t Temperament) Equal(other Temperament) bool {
	if !t.IsValid() && !other.IsValid() {
		return true
	}
	if len(t.offsets) != len(other.offsets) {
		return false
	}
	for m, off := range t.offsets {
		if !off.Equal(other.offsets[m]) {
			return false
		}
	}
	return true
}

// Near reports whether t and other are both valid, have the same dimension N,
// tolerance >= 0 (tolerance.Sign() >= 0), and |T_1(m) - T_2(m)| <= tolerance
// for all m in 0..N-1. It returns false if either temperament is invalid, if
// N_1 != N_2, or if tolerance < 0.
func (t Temperament) Near(other Temperament, tolerance cents.Cents) bool {
	if !t.IsValid() || !other.IsValid() {
		return false
	}
	if len(t.offsets) != len(other.offsets) {
		return false
	}
	if tolerance.Sign() < 0 {
		return false
	}
	for m, off := range t.offsets {
		if off.Equal(other.offsets[m]) {
			continue
		}
		ok, err := off.Near(other.offsets[m], tolerance)
		if err != nil || !ok {
			return false
		}
	}
	return true
}

// IsEqual reports whether t is a valid equal temperament (N-EDO), i.e.,
// T(m).Equal(cents.Step(m, N)) for all m in 0..N-1.
func (t Temperament) IsEqual() bool {
	if !t.IsValid() {
		return false
	}
	n := len(t.offsets)
	for m, off := range t.offsets {
		// cents.Step(m, n) is the rational 1200*m/n cents; compare the reduced
		// fraction of T(m) with it instead of creating the step value.
		num, den, ok := off.Fraction()
		if !ok {
			return false
		}
		stepNum, stepDen := 1200*m, n
		g := int(gcdUint64(uint64(stepNum), uint64(stepDen)))
		if num != stepNum/g || den != stepDen/g {
			return false
		}
	}
	return true
}

// IsRational reports whether t is valid and every step offset T(m) is an exact
// 31-limit rational frequency ratio p_m/q_m in [1, 2) (i.e.,
// _, _, ok := T(m).Ratio() is true for all m in 0..N-1).
func (t Temperament) IsRational() bool {
	if !t.IsValid() {
		return false
	}
	for _, off := range t.offsets {
		if _, _, ok := off.Ratio(); !ok {
			return false
		}
	}
	return true
}

// NearestStep finds the closest intra-octave step index m* in [0, N-1] to the
// given offset x in cents (wrapping x modulo 1200 cents and considering the
// upper octave boundary T(N) = 1200 cents as step 0 of the next octave), and
// returns m* along with the signed deviation delta = x_mod - T(m*) (or
// x_mod - 1200 if the upper octave boundary is strictly closer). On an exact
// tie between adjacent candidates m and m+1 in 0..N, the lower candidate m in
// 0..N-1 is chosen with positive deviation. It returns [ErrInvalidArgument] if
// t is invalid.
func (t Temperament) NearestStep(x cents.Cents) (step int, deviation cents.Cents, err error) {
	if !t.IsValid() {
		return 0, cents.Cents{}, errorf(ErrInvalidArgument, "NearestStep: invalid temperament")
	}
	_, rest := x.SplitOctaves()
	if rest.Equal(cents.Zero()) {
		return 0, cents.Zero(), nil
	}
	cand, dev, err := t.nearestInOctave(rest)
	if err != nil {
		return 0, cents.Cents{}, err
	}
	if cand == len(t.offsets) {
		cand = 0
	}
	return cand, dev, nil
}

// MustNearestStep is like [Temperament.NearestStep] but panics on error.
func (t Temperament) MustNearestStep(x cents.Cents) (step int, deviation cents.Cents) {
	return must2(t.NearestStep(x))
}

// NearestCumulativeStep finds the closest multi-octave step index
// k* = N*q* + m* in Z to the given offset x in cents, and returns k* along with
// the signed deviation delta = x - c(k*) in cents. It uses the same
// tie-breaking rule as [Temperament.NearestStep]. It returns
// [ErrInvalidArgument] if t is invalid.
func (t Temperament) NearestCumulativeStep(x cents.Cents) (step int, deviation cents.Cents, err error) {
	if !t.IsValid() {
		return 0, cents.Cents{}, errorf(ErrInvalidArgument, "NearestCumulativeStep: invalid temperament")
	}
	// |x| <= cents.MaxMagnitude bounds |q| by 7500 octaves, and N is at most
	// MaxStepsPerOctave, so q*N + cand always fits in a 64-bit int.
	q, rest := x.SplitOctaves()
	n := len(t.offsets)
	if rest.Equal(cents.Zero()) {
		return q * n, cents.Zero(), nil
	}
	cand, dev, err := t.nearestInOctave(rest)
	if err != nil {
		return 0, cents.Cents{}, err
	}
	return q*n + cand, dev, nil
}

// MustNearestCumulativeStep is like [Temperament.NearestCumulativeStep] but
// panics on error.
func (t Temperament) MustNearestCumulativeStep(x cents.Cents) (step int, deviation cents.Cents) {
	return must2(t.NearestCumulativeStep(x))
}

// nearestInOctave finds the nearest candidate index in [0, N] to rest in
// (0, 1200) cents, where candidate N represents the upper octave boundary 1200
// cents. On an exact tie between i-1 and i, i-1 is chosen with positive
// deviation.
func (t Temperament) nearestInOctave(rest cents.Cents) (int, cents.Cents, error) {
	n := len(t.offsets)
	lo, hi := 1, n
	for lo < hi {
		mid := int(uint(lo+hi) >> 1)
		cmp := t.offsets[mid].Compare(rest)
		if cmp == 0 {
			return mid, cents.Zero(), nil
		}
		if cmp < 0 {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	// Here T(i-1) < rest and, if i < N, rest < T(i): an exact match has
	// already returned inside the loop.
	i := lo
	dLo, err := rest.Sub(t.offsets[i-1])
	if err != nil {
		return 0, cents.Cents{}, wrapCentsError("NearestStep", err)
	}

	upperBound := cents.Octave()
	if i < n {
		upperBound = t.offsets[i]
	}
	dHi, err := upperBound.Sub(rest)
	if err != nil {
		return 0, cents.Cents{}, wrapCentsError("NearestStep", err)
	}

	if dLo.Compare(dHi) <= 0 {
		return i - 1, dLo, nil
	}
	return i, dHi.Neg(), nil
}

// String returns a human-readable summary of t:
//   - "Temperament(<invalid>)" if !t.IsValid();
//   - "<Name> (<N> steps/octave)" if t.Name() != "";
//   - "Custom (<N> steps/octave)" if t.IsValid() and t.Name() == "".
func (t Temperament) String() string {
	if !t.IsValid() {
		return "Temperament(<invalid>)"
	}
	if t.name != "" {
		return fmt.Sprintf("%s (%d steps/octave)", t.name, len(t.offsets))
	}
	return fmt.Sprintf("Custom (%d steps/octave)", len(t.offsets))
}
