package voice

import (
	"fmt"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/instrument"
	"github.com/go-muse/musex/system"
)

// SoundLocation is the location of a sound relative to a reference range:
// the index of its element, below, inside or above the range, and the
// distance in steps of the system to the nearest bound, zero inside.
type SoundLocation struct {
	Index    int
	Location instrument.Location
	Outside  system.Step
}

// Weighting is the weighting of the sounds in the tessitura number.
type Weighting uint8

// Weightings. WeightingUnset is not a weighting.
const (
	// WeightingUnset is the zero value: no weighting.
	WeightingUnset Weighting = iota
	// ByLength weights each sound by its length; a tied sound by its whole
	// length.
	ByLength
	// ByCount weights each sound by one.
	ByCount
)

var weightingNames = [...]string{
	ByLength: "by-length",
	ByCount:  "by-count",
}

// IsValid reports whether w is ByLength or ByCount.
func (w Weighting) IsValid() bool {
	return w == ByLength || w == ByCount
}

// String returns "by-length" or "by-count"; "Weighting(unset)" for
// WeightingUnset and "Weighting(invalid)" for values outside the list.
func (w Weighting) String() string {
	return enumString(w, weightingNames[:], "Weighting")
}

// TessituraKind is the kind of a tessitura: the zone of its number.
type TessituraKind uint8

// Kinds of a tessitura. TessituraUnset is not a kind.
const (
	// TessituraUnset is the zero value: no kind.
	TessituraUnset TessituraKind = iota
	// TessituraLow is a number below the low bound of the zones.
	TessituraLow
	// TessituraMiddle is a number between the bounds, both included.
	TessituraMiddle
	// TessituraHigh is a number above the high bound of the zones.
	TessituraHigh
)

var tessituraKindNames = [...]string{
	TessituraLow:    "low",
	TessituraMiddle: "middle",
	TessituraHigh:   "high",
}

// IsValid reports whether k is one of the three kinds.
func (k TessituraKind) IsValid() bool {
	return k >= TessituraLow && k <= TessituraHigh
}

// String returns "low", "middle" or "high"; "TessituraKind(unset)" for
// TessituraUnset and "TessituraKind(invalid)" for values outside the list.
func (k TessituraKind) String() string {
	return enumString(k, tessituraKindNames[:], "TessituraKind")
}

// Tessitura is the result of a tessitura: the exact number r, the mean
// position of the sounds where the lower bound of the reference range is 0
// and the upper bound is 1, and its kind by the zones. A number below 0 or
// above 1 is allowed: the kind is the mean position, not whether the line
// fits in the range.
type Tessitura struct {
	Number fraction.Fraction
	Kind   TessituraKind
}

// Zones are the bounds of the zones of the tessitura kinds, low and high with
// 0 <= low <= high <= 1. The default zones 1/3 and 2/3 divide the width into
// three equal parts: a rule of the product, not a fact of music theory.
//
// Zones is an immutable value; the bounds are kept as given and compared by
// value. The zero value Zones{} is unset. Compare zones with Equal.
type Zones struct {
	set       bool
	low, high fraction.Fraction
}

// DefaultZones returns the zones 1/3 and 2/3.
func DefaultZones() Zones {
	return Zones{set: true, low: fraction.MustNew(1, 3), high: fraction.MustNew(2, 3)}
}

// NewZones returns the zones with the bounds low and high.
//
// Errors, in this order: ErrUnset for an unset low, then for an unset high
// (with fraction.ErrInvalid); ErrZones unless 0 <= low <= high <= 1.
func NewZones(low, high fraction.Fraction) (Zones, error) {
	if !low.IsValid() {
		return Zones{}, unsetError("low bound of the zones", fraction.ErrInvalid)
	}

	if !high.IsValid() {
		return Zones{}, unsetError("high bound of the zones", fraction.ErrInvalid)
	}

	if low.Sign() < 0 || low.Greater(high) || high.Greater(fraction.One()) {
		return Zones{}, fmt.Errorf("%w: %s and %s", ErrZones, low, high)
	}

	return Zones{set: true, low: low, high: high}, nil
}

// MustNewZones is like NewZones but panics with its error.
func MustNewZones(low, high fraction.Fraction) Zones {
	return must(NewZones(low, high))
}

// IsSet reports whether z is set.
func (z Zones) IsSet() bool {
	return z.set
}

// Low returns the low bound as given; fraction.Fraction{} for unset zones.
func (z Zones) Low() fraction.Fraction {
	return z.low
}

// High returns the high bound as given; fraction.Fraction{} for unset zones.
func (z Zones) High() fraction.Fraction {
	return z.high
}

// Equal reports whether the bounds of z and other are equal by value: the
// zones 1/3, 2/3 and 2/6, 4/6 are equal. Two unset zones are equal; unset
// zones equal no set ones.
func (z Zones) Equal(other Zones) bool {
	if z.set != other.set {
		return false
	}

	return !z.set || (z.low.Equal(other.low) && z.high.Equal(other.high))
}

// String returns a debugging form, such as "Zones(1/3, 2/3)", or
// "Zones(unset)".
func (z Zones) String() string {
	if !z.set {
		return "Zones(unset)"
	}

	return "Zones(" + z.low.String() + ", " + z.high.String() + ")"
}

// kindOf returns the kind of the number r in z.
func (z Zones) kindOf(r fraction.Fraction) TessituraKind {
	switch {
	case r.Less(z.low):
		return TessituraLow
	case r.Greater(z.high):
		return TessituraHigh
	default:
		return TessituraMiddle
	}
}

// inReference chooses the system of an operation with the reference range
// rng and checks rng in it.
func (l *Line) inReference(sounds []int, sys *system.System, explicit bool, rng instrument.Range, refSys *system.System) (*system.System, error) {
	sys, err := l.chooseSystem(sounds, sys, explicit, refSys)
	if err != nil {
		return nil, err
	}

	if err := rng.Check(sys); err != nil {
		return nil, fmt.Errorf("voice: reference range: %w", err)
	}

	return sys, nil
}

// located returns the system, the reference range and the indexes of the
// sounds checked for an operation of locations; nil sounds for a line
// without sounds.
func (l *Line) located(sys *system.System, explicit bool, ref Reference) (*system.System, instrument.Range, []int, error) {
	if !l.IsSet() {
		return nil, instrument.Range{}, nil, errUnsetLine()
	}

	if explicit {
		if err := checkSystem(sys); err != nil {
			return nil, instrument.Range{}, nil, err
		}
	}

	if err := ref.check(); err != nil {
		return nil, instrument.Range{}, nil, err
	}

	rng, refSys, err := ref.resolve()
	if err != nil {
		return nil, instrument.Range{}, nil, err
	}

	sounds := l.soundIndexes()
	if len(sounds) == 0 {
		return nil, rng, nil, nil
	}

	sys, err = l.inReference(sounds, sys, explicit, rng, refSys)
	if err != nil {
		return nil, instrument.Range{}, nil, err
	}

	if err := l.checkSounds(sys, sounds); err != nil {
		return nil, instrument.Range{}, nil, err
	}

	return sys, rng, sounds, nil
}

// Locate returns for each sound in order the index of its element, its
// location relative to the reference range ref (instrument.Range.Locate) and
// its distance in steps to the nearest bound: below, the distance from the
// sound to the lower bound; above, from the upper bound to the sound; zero
// inside. A line without sounds gives an empty slice without choosing a
// system.
//
// Without an explicit system the operation works in the system of the
// preset or the instrument of ref, where unbound sounds are allowed and bound
// sounds must have that system (ErrDifferentSystems); for a range or a
// placement, in the common system of the sounds, as Intervals does.
//
// Errors, in this order: ErrUnset for an unset line, then for an unset ref,
// a nil or unset preset or instrument (with instrument.ErrUnset) or a
// placement outside the list; ErrNoRange when ref has no range; the choice of
// the system; the errors of instrument.Range.Check for the reference range
// (instrument.ErrInadmissible, instrument.ErrInverted); system.ErrInadmissible
// for a sound; system.ErrOutOfRange for a distance. The errors of the sounds
// are in an *ItemError with the index of the sound.
func (l *Line) Locate(ref Reference) ([]SoundLocation, error) {
	return l.locate(nil, false, ref)
}

// MustLocate is like Locate but panics with its error.
func (l *Line) MustLocate(ref Reference) []SoundLocation {
	return must(l.Locate(ref))
}

// LocateIn is like Locate in the explicit system sys: the sources of the
// notes, the preset and the instrument are not read for the system. A nil or
// unset sys reports ErrUnset (with system.ErrInvalid) after the check of the
// line.
func (l *Line) LocateIn(sys *system.System, ref Reference) ([]SoundLocation, error) {
	return l.locate(sys, true, ref)
}

// MustLocateIn is like LocateIn but panics with its error.
func (l *Line) MustLocateIn(sys *system.System, ref Reference) []SoundLocation {
	return must(l.LocateIn(sys, ref))
}

func (l *Line) locate(sys *system.System, explicit bool, ref Reference) ([]SoundLocation, error) {
	sys, rng, sounds, err := l.located(sys, explicit, ref)
	if err != nil {
		return nil, err
	}

	out := make([]SoundLocation, 0, len(sounds))

	for _, i := range sounds {
		x := l.pitchOf(i)

		loc, err := rng.Locate(sys, x)
		if err != nil {
			return nil, itemError(i, fmt.Errorf("voice: location: %w", err))
		}

		var outside system.Step

		switch loc {
		case instrument.Below:
			outside, err = sys.Distance(x, rng.Lower())
		case instrument.Above:
			outside, err = sys.Distance(rng.Upper(), x)
		default:
		}

		if err != nil {
			return nil, itemError(i, fmt.Errorf("voice: distance to the range: %w", err))
		}

		out = append(out, SoundLocation{Index: i, Location: loc, Outside: outside})
	}

	return out, nil
}

// Within reports whether every sound is inside the reference range ref, and
// the indexes of the elements of the sounds outside it in order; true and an
// empty slice for a line without sounds. It computes no distances. The
// errors are those of Locate except system.ErrOutOfRange.
func (l *Line) Within(ref Reference) (bool, []int, error) {
	return l.within(nil, false, ref)
}

// MustWithin is like Within but panics with its error.
func (l *Line) MustWithin(ref Reference) (bool, []int) {
	return must2(l.Within(ref))
}

// WithinIn is like Within in the explicit system sys.
func (l *Line) WithinIn(sys *system.System, ref Reference) (bool, []int, error) {
	return l.within(sys, true, ref)
}

// MustWithinIn is like WithinIn but panics with its error.
func (l *Line) MustWithinIn(sys *system.System, ref Reference) (bool, []int) {
	return must2(l.WithinIn(sys, ref))
}

func (l *Line) within(sys *system.System, explicit bool, ref Reference) (bool, []int, error) {
	sys, rng, sounds, err := l.located(sys, explicit, ref)
	if err != nil {
		return false, nil, err
	}

	outside := []int{}

	for _, i := range sounds {
		loc, err := rng.Locate(sys, l.pitchOf(i))
		if err != nil {
			return false, nil, itemError(i, fmt.Errorf("voice: location: %w", err))
		}

		if loc != instrument.Inside {
			outside = append(outside, i)
		}
	}

	return len(outside) == 0, outside, nil
}

// Tessitura is TessituraBy with ByLength and DefaultZones.
func (l *Line) Tessitura(ref Reference) (Tessitura, error) {
	return l.tessitura(nil, false, ref, ByLength, DefaultZones())
}

// MustTessitura is like Tessitura but panics with its error.
func (l *Line) MustTessitura(ref Reference) Tessitura {
	return must(l.Tessitura(ref))
}

// TessituraIn is TessituraByIn with ByLength and DefaultZones.
func (l *Line) TessituraIn(sys *system.System, ref Reference) (Tessitura, error) {
	return l.tessitura(sys, true, ref, ByLength, DefaultZones())
}

// MustTessituraIn is like TessituraIn but panics with its error.
func (l *Line) MustTessituraIn(sys *system.System, ref Reference) Tessitura {
	return must(l.TessituraIn(sys, ref))
}

// TessituraBy returns the tessitura of the line relative to the reference
// range ref: the exact number r = Σ wᵢ·D(xᵢ) / (W · Σ wᵢ), where D(x) is the
// distance in steps from the lower bound to the sound x, W is the width of
// the range, and wᵢ is the length of the sound (ByLength) or 1 (ByCount);
// rests do not count. The kind is low for r < low, middle for
// low <= r <= high and high for r > high of the zones z. This is a rule of
// the product: the kind is the mean position of the sounds, and a middle
// tessitura does not mean that the line fits in the range. The system is
// chosen as in Locate.
//
// Errors, in this order: ErrUnset for an unset line, then for an unset ref,
// w (WeightingUnset or outside the list) and z; ErrNoRange; ErrNoSounds; the
// choice of the system; the check of the reference range; ErrZeroWidth for a
// range of width 0; system.ErrInadmissible for a sound and
// system.ErrOutOfRange for its distance, in an *ItemError with its index;
// fraction.ErrOutOfRange, without an *ItemError, if one of the final values
// is outside the range of package common/fraction: the product wᵢ·D(xᵢ) of a
// sound, the sum Σ wᵢ·D(xᵢ), the sum Σ wᵢ, the denominator W·Σ wᵢ or r. The
// sums are exact, so this failure depends neither on intermediate sums nor on
// the order of the sounds.
func (l *Line) TessituraBy(ref Reference, w Weighting, z Zones) (Tessitura, error) {
	return l.tessitura(nil, false, ref, w, z)
}

// MustTessituraBy is like TessituraBy but panics with its error.
func (l *Line) MustTessituraBy(ref Reference, w Weighting, z Zones) Tessitura {
	return must(l.TessituraBy(ref, w, z))
}

// TessituraByIn is like TessituraBy in the explicit system sys.
func (l *Line) TessituraByIn(sys *system.System, ref Reference, w Weighting, z Zones) (Tessitura, error) {
	return l.tessitura(sys, true, ref, w, z)
}

// MustTessituraByIn is like TessituraByIn but panics with its error.
func (l *Line) MustTessituraByIn(sys *system.System, ref Reference, w Weighting, z Zones) Tessitura {
	return must(l.TessituraByIn(sys, ref, w, z))
}

func (l *Line) tessitura(sys *system.System, explicit bool, ref Reference, w Weighting, z Zones) (Tessitura, error) {
	rng, sys, sounds, err := l.tessituraInputs(sys, explicit, ref, w, z)
	if err != nil {
		return Tessitura{}, err
	}

	width, err := rng.Width(sys)
	if err != nil {
		return Tessitura{}, fmt.Errorf("voice: width of the reference range: %w", err)
	}

	if width == 0 {
		return Tessitura{}, fmt.Errorf("%w: %s", ErrZeroWidth, rng)
	}

	if err := l.checkSounds(sys, sounds); err != nil {
		return Tessitura{}, err
	}

	distances := make([]system.Step, len(sounds))

	for k, i := range sounds {
		d, err := sys.Distance(rng.Lower(), l.pitchOf(i))
		if err != nil {
			return Tessitura{}, itemError(i, fmt.Errorf("voice: distance from the lower bound: %w", err))
		}

		distances[k] = d
	}

	r, err := l.tessituraNumber(sounds, distances, width, w)
	if err != nil {
		return Tessitura{}, fmt.Errorf("voice: tessitura: %w", err)
	}

	return Tessitura{Number: r, Kind: z.kindOf(r)}, nil
}

// tessituraInputs checks the arguments of a tessitura in order and returns
// the reference range, the system and the sounds.
func (l *Line) tessituraInputs(sys *system.System, explicit bool, ref Reference, w Weighting, z Zones) (instrument.Range, *system.System, []int, error) {
	if !l.IsSet() {
		return instrument.Range{}, nil, nil, errUnsetLine()
	}

	if explicit {
		if err := checkSystem(sys); err != nil {
			return instrument.Range{}, nil, nil, err
		}
	}

	if err := ref.check(); err != nil {
		return instrument.Range{}, nil, nil, err
	}

	if !w.IsValid() {
		return instrument.Range{}, nil, nil, unsetError("weighting "+w.String(), nil)
	}

	if !z.IsSet() {
		return instrument.Range{}, nil, nil, unsetError("zones", nil)
	}

	rng, refSys, err := ref.resolve()
	if err != nil {
		return instrument.Range{}, nil, nil, err
	}

	sounds := l.soundIndexes()
	if len(sounds) == 0 {
		return instrument.Range{}, nil, nil, fmt.Errorf("%w: no tessitura", ErrNoSounds)
	}

	sys, err = l.inReference(sounds, sys, explicit, rng, refSys)
	if err != nil {
		return instrument.Range{}, nil, nil, err
	}

	return rng, sys, sounds, nil
}

// tessituraNumber returns Σ wᵢ·dᵢ / (width · Σ wᵢ) in reduced form.
func (l *Line) tessituraNumber(sounds []int, distances []system.Step, width system.Step, w Weighting) (fraction.Fraction, error) {
	var sum, weights exactSum

	for k, i := range sounds {
		weight := fraction.One()

		if w == ByLength {
			f, err := l.elems[i].dur.length.Fraction()
			if err != nil {
				return fraction.Fraction{}, err
			}

			weight = f
		}

		term, err := weight.MulInt(int(distances[k]))
		if err != nil {
			return fraction.Fraction{}, err
		}

		sum.add(term)
		weights.add(weight)
	}

	num, err := sum.result()
	if err != nil {
		return fraction.Fraction{}, err
	}

	total, err := weights.result()
	if err != nil {
		return fraction.Fraction{}, err
	}

	den, err := total.MulInt(int(width))
	if err != nil {
		return fraction.Fraction{}, err
	}

	return num.Div(den)
}

// exactSum accumulates the exact sum of valid fractions, so that its result
// and its failure depend only on the exact total, as for fraction.Sum. It adds
// in place while every partial sum is in range and keeps no list of the
// elements; once a partial sum is outside the range, it keeps the last partial
// sum in range and the remaining elements, and fraction.Sum decides the result.
// The zero value is the empty sum.
type exactSum struct {
	acc  fraction.Fraction
	tail []fraction.Fraction // nil while every partial sum is in range
}

// add adds the valid fraction f to s.
func (s *exactSum) add(f fraction.Fraction) {
	if s.tail == nil {
		if !s.acc.IsValid() {
			s.acc = f.Reduce()
			return
		}

		if next, err := s.acc.Add(f); err == nil {
			s.acc = next
			return
		}

		s.tail = []fraction.Fraction{s.acc}
	}

	s.tail = append(s.tail, f)
}

// result returns the exact sum, 0 for the empty sum, or an error wrapping
// fraction.ErrOutOfRange if it is outside the range.
func (s *exactSum) result() (fraction.Fraction, error) {
	switch {
	case s.tail != nil:
		return fraction.Sum(s.tail...)
	case !s.acc.IsValid():
		return fraction.Zero(), nil
	default:
		return s.acc, nil
	}
}
