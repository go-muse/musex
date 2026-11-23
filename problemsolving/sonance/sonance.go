package sonance

import (
	"errors"
	"fmt"
	"math"

	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pcset"
	"github.com/go-muse/musex/system"
)

// Sentinel errors of the package.
var (
	// ErrUnset indicates an unset input: a nil or unset system, an unset
	// note or the zero pitch-class set.
	ErrUnset = errors.New("sonance: unset value")

	// ErrInvalidArgument indicates a profile without exactly seven finite
	// weights, an interval class outside 0..6, or a pitch-class set whose
	// modulus is not 12.
	ErrInvalidArgument = errors.New("sonance: no such profile or parameter")

	// ErrNoOctave indicates a note without an octave.
	ErrNoOctave = errors.New("sonance: note without octave")

	// ErrUnsupportedSystem indicates a system that is not twelve classes
	// with chromatic unit 1.
	ErrUnsupportedSystem = errors.New("sonance: system is not 12-tone equal-tempered")

	// ErrOutOfRange indicates a distance that does not fit, wrapping
	// system.ErrOutOfRange.
	ErrOutOfRange = errors.New("sonance: result out of range")
)

func opError(op string, sentinel error) error {
	return fmt.Errorf("sonance: %s: %w", op, sentinel)
}

func rangeError(op string, cause error) error {
	return fmt.Errorf("sonance: %s: %w: %w", op, ErrOutOfRange, cause)
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// Profile is an immutable set of seven consonance weights, one per
// interval class 0 to 6. The zero Profile is unset.
type Profile struct {
	weights [7]float64
	set     bool
}

// HuronProfile returns the published aggregate dyadic consonance profile
// of Huron (1994): weights 0, -1.428, -0.582, +0.594, +0.386, +1.240,
// -0.453 for the interval classes 0 to 6.
func HuronProfile() Profile {
	return Profile{weights: [7]float64{0, -1.428, -0.582, 0.594, 0.386, 1.240, -0.453}, set: true}
}

// NewProfile returns the profile of the seven weights of the interval
// classes 0 to 6, in that order.
//
// Errors: ErrInvalidArgument for a slice of length other than seven or
// with a non-finite weight.
func NewProfile(weights []float64) (Profile, error) {
	const op = "NewProfile"

	if len(weights) != 7 {
		return Profile{}, opError(op, ErrInvalidArgument)
	}
	for _, w := range weights {
		if math.IsNaN(w) || math.IsInf(w, 0) {
			return Profile{}, opError(op, ErrInvalidArgument)
		}
	}

	var p Profile
	copy(p.weights[:], weights)
	p.set = true

	return p, nil
}

// MustNewProfile is like NewProfile but panics with its error.
func MustNewProfile(weights []float64) Profile { return must(NewProfile(weights)) }

// IsValid reports whether the profile was constructed.
func (p Profile) IsValid() bool { return p.set }

// Weights returns a copy of the seven weights, or nil for an unset profile.
func (p Profile) Weights() []float64 {
	if !p.set {
		return nil
	}
	return append([]float64(nil), p.weights[:]...)
}

// Weight returns the weight of the interval class ic, 0 to 6.
//
// Errors: ErrUnset for an unset profile; ErrInvalidArgument for ic outside
// 0..6.
func (p Profile) Weight(ic int) (float64, error) {
	const op = "Weight"

	if !p.set {
		return 0, opError(op, ErrUnset)
	}
	if ic < 0 || ic > 6 {
		return 0, opError(op, ErrInvalidArgument)
	}

	return p.weights[ic], nil
}

// MustWeight is like Weight but panics with its error.
func (p Profile) MustWeight(ic int) float64 { return must(p.Weight(ic)) }

// Contribution is one interval class of a score: how many pairs of that
// class were counted, with which weight, and their product.
type Contribution struct {
	ic     int
	count  int
	weight float64
}

// IntervalClass returns the interval class, 0 to 6.
func (c Contribution) IntervalClass() int { return c.ic }

// Count returns the number of counted pairs of the class.
func (c Contribution) Count() int { return c.count }

// Weight returns the profile weight of the class.
func (c Contribution) Weight() float64 { return c.weight }

// Value returns Count times Weight.
func (c Contribution) Value() float64 { return float64(c.count) * c.weight }

// Score is an aggregate dyadic consonance value with its contributions.
// The zero Score has value 0 and no contributions.
type Score struct {
	contributions []Contribution
	value         float64
}

// Value returns the sum of the contributions.
func (s Score) Value() float64 { return s.value }

// Contributions returns a copy of the seven contributions, interval class
// 0 to 6 in order, including classes with zero count.
func (s Score) Contributions() []Contribution {
	return append([]Contribution(nil), s.contributions...)
}

// build sums the counts of the seven classes under the set profile p.
func build(p Profile, counts [7]int) Score {
	out := Score{contributions: make([]Contribution, 7)}
	for ic := 0; ic < 7; ic++ {
		c := Contribution{ic: ic, count: counts[ic], weight: p.weights[ic]}
		out.contributions[ic] = c
		out.value += c.Value()
	}

	return out
}

// ScoreSet returns the aggregate dyadic consonance of the twelve-class
// pitch-class set s under the profile p, counting its interval-class
// vector. IC0 is always zero for a set.
//
// Errors, in this order: ErrUnset for an unset profile or the zero set;
// ErrInvalidArgument for a set whose modulus is not 12.
func ScoreSet(s pcset.Set, p Profile) (Score, error) {
	const op = "ScoreSet"

	if !p.set {
		return Score{}, opError(op, ErrUnset)
	}
	if !s.IsValid() {
		return Score{}, opError(op, ErrUnset)
	}
	if s.Modulus() != 12 {
		return Score{}, opError(op, ErrInvalidArgument)
	}

	v := s.IntervalVector()
	var counts [7]int
	for ic := 1; ic <= 6; ic++ {
		counts[ic] = v.MustCount(ic)
	}

	return build(p, counts), nil
}

// MustScoreSet is like ScoreSet but panics with its error.
func MustScoreSet(s pcset.Set, p Profile) Score { return must(ScoreSet(s, p)) }

// ScoreNotes returns the aggregate dyadic consonance of the actual pitch
// pairs of notes in sys under the profile p. Every unordered pair of
// positions counts once, including unisons and octaves; cents do not take
// part. An empty or single-note input gives the zero score with zero
// counts.
//
// Errors, in this order: ErrUnset for an unset profile, a nil or unset
// system or an unset note; ErrUnsupportedSystem; ErrNoOctave;
// ErrOutOfRange together with system.ErrOutOfRange.
func ScoreNotes(sys *system.System, notes []note.Note, p Profile) (Score, error) {
	const op = "ScoreNotes"

	if !p.set {
		return Score{}, opError(op, ErrUnset)
	}
	if !sys.IsValid() {
		return Score{}, opError(op, ErrUnset)
	}
	if sys.Classes() != 12 || sys.Chromatic() != 1 {
		return Score{}, opError(op, ErrUnsupportedSystem)
	}
	for _, n := range notes {
		if !n.IsValid() {
			return Score{}, opError(op, ErrUnset)
		}
		if !n.HasOctave() {
			return Score{}, opError(op, ErrNoOctave)
		}
	}

	var counts [7]int
	for i := range notes {
		for j := i + 1; j < len(notes); j++ {
			pi, _ := notes[i].Pitch()
			pj, _ := notes[j].Pitch()
			s, err := sys.Distance(pi, pj)
			if err != nil {
				return Score{}, rangeError(op, err)
			}
			d := int(s)
			if d < 0 {
				d = -d
			}
			class := d % 12
			if class > 6 {
				class = 12 - class
			}
			counts[class]++
		}
	}

	return build(p, counts), nil
}

// MustScoreNotes is like ScoreNotes but panics with its error.
func MustScoreNotes(sys *system.System, notes []note.Note, p Profile) Score {
	return must(ScoreNotes(sys, notes, p))
}
