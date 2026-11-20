package voice

import (
	"fmt"

	"github.com/go-muse/musex/instrument"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/system"
)

// MelodicInterval is the interval from a sound to the next sound of a line:
// the indexes of their elements, the interval with its direction, and
// whether at least one rest lies between them.
type MelodicInterval struct {
	From, To    int
	Interval    interval.Interval
	ThroughRest bool
}

// chooseSystem returns the system of a comparison of the sounds: sys when it
// is explicit, otherwise the system chosen from refSys and the sounds.
func (l *Line) chooseSystem(sounds []int, sys *system.System, explicit bool, refSys *system.System) (*system.System, error) {
	if explicit {
		return sys, nil
	}

	return l.compareSystem(sounds, refSys)
}

// Intervals returns for each pair of neighboring sounds in order the indexes
// of their elements, the interval from the first to the second with its
// direction by the writing (interval.BetweenPitches), and whether a rest lies
// between them. A line of n sounds has n - 1 intervals; a repeated pitch is
// P1. The corrections of the notes do not take part.
//
// The system is the common system of the settings of the sounds: every
// sound must be bound (ErrNoSystem) and all must have one system
// (ErrDifferentSystems). Use IntervalsIn for an explicit system. A line
// without sounds gives an empty slice.
//
// Errors, in this order: ErrUnset for an unset line; the choice of the
// system; system.ErrInadmissible for a name of a sound that is not admissible
// in it; interval.ErrOutOfRange for an interval that cannot be represented.
// The errors of the sounds are in an *ItemError with the index of the sound,
// for an interval the index of its second sound.
func (l *Line) Intervals() ([]MelodicInterval, error) {
	return l.intervals(nil, false)
}

// MustIntervals is like Intervals but panics with its error.
func (l *Line) MustIntervals() []MelodicInterval {
	return must(l.Intervals())
}

// IntervalsIn is like Intervals in the explicit system sys: the sources of
// the notes are not read. A nil or unset sys reports ErrUnset (with
// system.ErrInvalid) after the check of the line.
func (l *Line) IntervalsIn(sys *system.System) ([]MelodicInterval, error) {
	return l.intervals(sys, true)
}

// MustIntervalsIn is like IntervalsIn but panics with its error.
func (l *Line) MustIntervalsIn(sys *system.System) []MelodicInterval {
	return must(l.IntervalsIn(sys))
}

func (l *Line) intervals(sys *system.System, explicit bool) ([]MelodicInterval, error) {
	if !l.IsSet() {
		return nil, errUnsetLine()
	}

	if explicit {
		if err := checkSystem(sys); err != nil {
			return nil, err
		}
	}

	sounds := l.soundIndexes()
	if len(sounds) == 0 {
		return []MelodicInterval{}, nil
	}

	sys, err := l.chooseSystem(sounds, sys, explicit, nil)
	if err != nil {
		return nil, err
	}

	if err := l.checkSounds(sys, sounds); err != nil {
		return nil, err
	}

	out := make([]MelodicInterval, 0, len(sounds)-1)

	for k := 1; k < len(sounds); k++ {
		from, to := sounds[k-1], sounds[k]

		iv, err := interval.BetweenPitches(sys, l.pitchOf(from), l.pitchOf(to))
		if err != nil {
			return nil, itemError(to, fmt.Errorf("voice: interval from element %d: %w", from, err))
		}

		out = append(out, MelodicInterval{From: from, To: to, Interval: iv, ThroughRest: to-from > 1})
	}

	return out, nil
}

// extremes returns the system of the comparison and the indexes of the
// lowest and the highest sound by position; of sounds at one position the
// first in the order of the line.
func (l *Line) extremes(sys *system.System, explicit bool) (*system.System, int, int, error) {
	if !l.IsSet() {
		return nil, 0, 0, errUnsetLine()
	}

	if explicit {
		if err := checkSystem(sys); err != nil {
			return nil, 0, 0, err
		}
	}

	sounds := l.soundIndexes()
	if len(sounds) == 0 {
		return nil, 0, 0, fmt.Errorf("%w: no extremes", ErrNoSounds)
	}

	sys, err := l.chooseSystem(sounds, sys, explicit, nil)
	if err != nil {
		return nil, 0, 0, err
	}

	if err := l.checkSounds(sys, sounds); err != nil {
		return nil, 0, 0, err
	}

	lowest, highest := sounds[0], sounds[0]

	for _, i := range sounds[1:] {
		p := l.pitchOf(i)

		below, err := sys.Compare(p, l.pitchOf(lowest))
		if err != nil {
			return nil, 0, 0, itemError(i, fmt.Errorf("voice: compare: %w", err))
		}

		above, err := sys.Compare(p, l.pitchOf(highest))
		if err != nil {
			return nil, 0, 0, itemError(i, fmt.Errorf("voice: compare: %w", err))
		}

		if below < 0 {
			lowest = i
		}

		if above > 0 {
			highest = i
		}
	}

	return sys, lowest, highest, nil
}

// Range returns the range of the line from its lowest to its highest sound
// by position in the system of the comparison; of sounds at one position the
// first in the order of the line is taken. The bounds are the written pitches
// of these sounds.
//
// Errors, in this order: ErrUnset for an unset line; ErrNoSounds for a line
// without sounds; the choice of the system and the check of the sounds, as in
// Intervals.
func (l *Line) Range() (instrument.Range, error) {
	return l.lineRange(nil, false)
}

// MustRange is like Range but panics with its error.
func (l *Line) MustRange() instrument.Range {
	return must(l.Range())
}

// RangeIn is like Range in the explicit system sys.
func (l *Line) RangeIn(sys *system.System) (instrument.Range, error) {
	return l.lineRange(sys, true)
}

// MustRangeIn is like RangeIn but panics with its error.
func (l *Line) MustRangeIn(sys *system.System) instrument.Range {
	return must(l.RangeIn(sys))
}

func (l *Line) lineRange(sys *system.System, explicit bool) (instrument.Range, error) {
	_, lowest, highest, err := l.extremes(sys, explicit)
	if err != nil {
		return instrument.Range{}, err
	}

	r, err := instrument.NewRange(l.pitchOf(lowest), l.pitchOf(highest))
	if err != nil {
		return instrument.Range{}, fmt.Errorf("voice: range: %w", err)
	}

	return r, nil
}

// Extremes returns the indexes of the elements of the lowest and the highest
// sound, by the rule of Range; for one sound they are equal. The errors are
// those of Range.
func (l *Line) Extremes() (lowest, highest int, err error) {
	_, lowest, highest, err = l.extremes(nil, false)

	return lowest, highest, err
}

// MustExtremes is like Extremes but panics with its error.
func (l *Line) MustExtremes() (lowest, highest int) {
	return must2(l.Extremes())
}

// ExtremesIn is like Extremes in the explicit system sys.
func (l *Line) ExtremesIn(sys *system.System) (lowest, highest int, err error) {
	_, lowest, highest, err = l.extremes(sys, true)

	return lowest, highest, err
}

// MustExtremesIn is like ExtremesIn but panics with its error.
func (l *Line) MustExtremesIn(sys *system.System) (lowest, highest int) {
	return must2(l.ExtremesIn(sys))
}

// Ambitus returns the interval from the lower bound of Range to its upper
// bound by interval.BetweenPitches as is: its direction follows the writing.
// It is P1 for one sound or sounds of one record, and it can be descending
// for enharmonically shifted extremes, such as C♭4 and B♯3 (dd2 down).
//
// Errors: those of Range, then interval.ErrOutOfRange, without an
// *ItemError.
func (l *Line) Ambitus() (interval.Interval, error) {
	return l.ambitus(nil, false)
}

// MustAmbitus is like Ambitus but panics with its error.
func (l *Line) MustAmbitus() interval.Interval {
	return must(l.Ambitus())
}

// AmbitusIn is like Ambitus in the explicit system sys.
func (l *Line) AmbitusIn(sys *system.System) (interval.Interval, error) {
	return l.ambitus(sys, true)
}

// MustAmbitusIn is like AmbitusIn but panics with its error.
func (l *Line) MustAmbitusIn(sys *system.System) interval.Interval {
	return must(l.AmbitusIn(sys))
}

func (l *Line) ambitus(sys *system.System, explicit bool) (interval.Interval, error) {
	sys, lowest, highest, err := l.extremes(sys, explicit)
	if err != nil {
		return interval.Interval{}, err
	}

	iv, err := interval.BetweenPitches(sys, l.pitchOf(lowest), l.pitchOf(highest))
	if err != nil {
		return interval.Interval{}, fmt.Errorf("voice: ambitus: %w", err)
	}

	return iv, nil
}
