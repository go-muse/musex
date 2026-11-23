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
	// ErrUnset indicates an unset input: an unset profile, a nil or unset
	// system, an unset note or the zero pitch-class set.
	ErrUnset = errors.New("sonance: unset value")

	// ErrInvalidArgument indicates a profile of the wrong length or with a
	// non-finite weight, an interval outside the profile's range, or a
	// pitch-class set whose modulus is not 12.
	ErrInvalidArgument = errors.New("sonance: no such profile or parameter")

	// ErrNoOctave indicates a note without an octave.
	ErrNoOctave = errors.New("sonance: note without octave")

	// ErrUnsupportedSystem indicates a system that is not twelve classes
	// with chromatic unit 1.
	ErrUnsupportedSystem = errors.New("sonance: system is not 12-tone equal-tempered")

	// ErrInadmissible indicates a note name that the system does not admit,
	// wrapping system.ErrInadmissible.
	ErrInadmissible = errors.New("sonance: inadmissible note name")

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

func inadmissibleError(op string, cause error) error {
	return fmt.Errorf("sonance: %s: %w: %w", op, ErrInadmissible, cause)
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func finite(weights []float64) bool {
	for _, w := range weights {
		if math.IsNaN(w) || math.IsInf(w, 0) {
			return false
		}
	}
	return true
}

// Profile is an immutable set of seven consonance weights, one per
// interval class 0 to 6, for pitch-class sets. The zero Profile is unset.
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

	if len(weights) != 7 || !finite(weights) {
		return Profile{}, opError(op, ErrInvalidArgument)
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

// Equal reports whether p and other carry the same seven weights.
func (p Profile) Equal(other Profile) bool { return p.set == other.set && p.weights == other.weights }

// PairProfile is an immutable set of twelve consonance weights, one per
// tempered distance modulo 12 (0 to 11 semitones), for pairs of actual
// pitches. Unlike Profile it distinguishes an interval from its inversion,
// so a profile can weight a fourth differently from a fifth. The zero
// PairProfile is unset.
type PairProfile struct {
	weights [12]float64
	set     bool
}

// HuronPairProfile returns Huron's published weights spread over the twelve
// tempered distances: each interval class weight applies to both of its
// members (1 and 11, 2 and 10, 3 and 9, 4 and 8, 5 and 7), the tritone 6
// keeps its weight and unisons and octaves (0) carry weight zero. It scores
// actual pitch pairs like HuronProfile scores sets; use NewPairProfile to
// distinguish inversions.
func HuronPairProfile() PairProfile {
	ic := HuronProfile().weights
	return PairProfile{weights: [12]float64{
		ic[0], ic[1], ic[2], ic[3], ic[4], ic[5], ic[6], ic[5], ic[4], ic[3], ic[2], ic[1],
	}, set: true}
}

// NewPairProfile returns the profile of the twelve weights of the tempered
// distances 0 to 11 semitones, in that order.
//
// Errors: ErrInvalidArgument for a slice of length other than twelve or
// with a non-finite weight.
func NewPairProfile(weights []float64) (PairProfile, error) {
	const op = "NewPairProfile"

	if len(weights) != 12 || !finite(weights) {
		return PairProfile{}, opError(op, ErrInvalidArgument)
	}

	var p PairProfile
	copy(p.weights[:], weights)
	p.set = true

	return p, nil
}

// MustNewPairProfile is like NewPairProfile but panics with its error.
func MustNewPairProfile(weights []float64) PairProfile { return must(NewPairProfile(weights)) }

// IsValid reports whether the profile was constructed.
func (p PairProfile) IsValid() bool { return p.set }

// Weights returns a copy of the twelve weights, or nil for an unset
// profile.
func (p PairProfile) Weights() []float64 {
	if !p.set {
		return nil
	}
	return append([]float64(nil), p.weights[:]...)
}

// Weight returns the weight of the tempered distance semis, 0 to 11.
//
// Errors: ErrUnset for an unset profile; ErrInvalidArgument for semis
// outside 0..11.
func (p PairProfile) Weight(semis int) (float64, error) {
	const op = "Weight"

	if !p.set {
		return 0, opError(op, ErrUnset)
	}
	if semis < 0 || semis > 11 {
		return 0, opError(op, ErrInvalidArgument)
	}

	return p.weights[semis], nil
}

// MustWeight is like Weight but panics with its error.
func (p PairProfile) MustWeight(semis int) float64 { return must(p.Weight(semis)) }

// Equal reports whether p and other carry the same twelve weights.
func (p PairProfile) Equal(other PairProfile) bool {
	return p.set == other.set && p.weights == other.weights
}

// Contribution is one counted interval of a score: how many pairs of that
// interval were counted, with which weight, and their product.
type Contribution struct {
	interval int
	count    int
	weight   float64
}

// Interval returns the counted interval: the interval class 0 to 6 for a
// set score, the tempered distance modulo 12 (0 to 11 semitones) for a
// note-pair score.
func (c Contribution) Interval() int { return c.interval }

// Count returns the number of counted pairs of the interval.
func (c Contribution) Count() int { return c.count }

// Weight returns the profile weight of the interval.
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

// Contributions returns a copy of the contributions in interval order,
// including intervals with zero count: seven for a set score, twelve for a
// note-pair score.
func (s Score) Contributions() []Contribution {
	return append([]Contribution(nil), s.contributions...)
}

// build sums the counts under the weights, one contribution per interval.
func build(weights []float64, counts []int) Score {
	out := Score{contributions: make([]Contribution, len(weights))}
	for i, w := range weights {
		c := Contribution{interval: i, count: counts[i], weight: w}
		out.contributions[i] = c
		out.value += c.Value()
	}

	return out
}

// ScoreSet returns the aggregate dyadic consonance of the twelve-class
// pitch-class set s under the interval-class profile p, counting its
// interval-class vector. IC0 is always zero for a set.
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
	counts := make([]int, 7)
	for ic := 1; ic <= 6; ic++ {
		counts[ic] = v.MustCount(ic)
	}

	return build(p.weights[:], counts), nil
}

// MustScoreSet is like ScoreSet but panics with its error.
func MustScoreSet(s pcset.Set, p Profile) Score { return must(ScoreSet(s, p)) }

// ScoreNotes returns the aggregate dyadic consonance of the actual pitch
// pairs of notes in sys under the pair profile p. Every unordered pair of
// positions counts once at its tempered distance modulo 12, including
// unisons and octaves and pairs of equal pitch classes at different
// octaves; cents do not take part. An empty or single-note input gives the
// zero score with zero counts.
//
// Errors, in this order: ErrUnset for an unset profile, a nil or unset
// system or an unset note; ErrUnsupportedSystem; ErrNoOctave;
// ErrInadmissible together with system.ErrInadmissible for a name the
// system does not admit; ErrOutOfRange together with system.ErrOutOfRange.
func ScoreNotes(sys *system.System, notes []note.Note, p PairProfile) (Score, error) {
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
		ok, err := sys.Admits(n.Spelling())
		if err != nil {
			return Score{}, opError(op, ErrUnset)
		}
		if !ok {
			return Score{}, inadmissibleError(op, system.ErrInadmissible)
		}
	}

	counts := make([]int, 12)
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
			counts[d%12]++
		}
	}

	return build(p.weights[:], counts), nil
}

// MustScoreNotes is like ScoreNotes but panics with its error.
func MustScoreNotes(sys *system.System, notes []note.Note, p PairProfile) Score {
	return must(ScoreNotes(sys, notes, p))
}
