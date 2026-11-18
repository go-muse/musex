package mode

import (
	"errors"
	"math/big"
	"slices"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

// PitchComparison selects exact written equality or equality of positions.
type PitchComparison uint8

const (
	// WrittenPitch compares the spelling and registered octave.
	WrittenPitch PitchComparison = iota + 1
	// PositionalPitch permits enharmonic spellings at the same position.
	PositionalPitch
)

// SpellingMembership lists matching degree numbers, independently per input.
// Slices are sorted, unique, and owned by the caller; both may be empty.
type SpellingMembership struct{ First, Second []int }

// PitchMembership lists matching places by Cycle then Number. Slices are
// independent and caller-owned; membership of both requires both nonempty.
type PitchMembership struct{ First, Second []DegreePlace }

func (m Mode) placeCoordinates(p DegreePlace) coordinates {
	c := coordinatesOf(m.prefixes[p.Number-1])
	if p.Cycle != 0 {
		c = c.add(coordinatesOf(m.cycle).mul(p.Cycle))
	}
	return c
}
func (m Mode) placePitch(sys *system.System, tonic pitch.Pitch, p DegreePlace) (pitch.Pitch, error) {
	shift := m.placeCoordinates(p)
	start := tonic.Spelling()
	startCarry, err := sys.Carry(start)
	if err != nil {
		return pitch.Pitch{}, queryError(err)
	}
	target, err := shift.transposeName(start)
	if err != nil {
		return pitch.Pitch{}, queryError(err)
	}
	targetCarry, err := sys.Carry(target)
	if err != nil {
		return pitch.Pitch{}, queryError(err)
	}
	letterSteps := new(big.Int).Add(shift.letters, big.NewInt(int64(sys.Boundary().DistanceUp(start.Letter()))))
	writtenDelta, _ := splitLetters(letterSteps)
	final := big.NewInt(int64(tonic.MustASPN()))
	final.Sub(final, big.NewInt(int64(startCarry)))
	final.Add(final, writtenDelta)
	final.Add(final, big.NewInt(int64(targetCarry)))
	if !final.IsInt64() {
		return pitch.Pitch{}, errors.Join(ErrOutOfRange, interval.ErrOutOfRange, system.ErrOutOfRange)
	}
	return pitch.FromASPN(target, int(final.Int64()))
}
func queryError(err error) error {
	if errors.Is(err, interval.ErrOutOfRange) || errors.Is(err, system.ErrOutOfRange) {
		return errors.Join(ErrOutOfRange, err)
	}
	return err
}
func (m Mode) eachCycle(w CycleWindow, visit func(int) error) error {
	if m.IsOpen() {
		if w.contains(0) {
			return visit(0)
		}
		return nil
	}
	for c := w.First; ; c++ {
		if err := visit(c); err != nil {
			return err
		}
		if c == w.Last {
			return nil
		}
	}
}

// MatchSpellings checks exact letters and alterations in the selected cycles,
// without treating missing notes as invalid inputs. Octave-cycle spelling
// membership is independent of the cycle. Open modes intersect the window
// with {0}. All initial arguments are validated even for an empty intersection.
func (m Mode) MatchSpellings(tonic, a, b spelling.Spelling, cycles CycleWindow) (SpellingMembership, error) {
	if !m.IsValid() || !tonic.IsValid() || !a.IsValid() || !b.IsValid() {
		return SpellingMembership{}, ErrUnset
	}
	if !cycles.valid() {
		return SpellingMembership{}, ErrInvalidArgument
	}
	first, second := make([]bool, m.Len()), make([]bool, m.Len())
	scan := func(c int) error {
		for j := range m.Len() {
			name, err := m.placeCoordinates(DegreePlace{j + 1, c}).transposeName(tonic)
			if err != nil {
				return queryError(err)
			}
			first[j] = first[j] || name.Equal(a)
			second[j] = second[j] || name.Equal(b)
		}
		return nil
	}
	var err error
	if m.IsOctave() {
		err = scan(0)
	} else {
		err = m.eachCycle(cycles, scan)
	}
	if err != nil {
		return SpellingMembership{}, err
	}
	result := SpellingMembership{First: make([]int, 0), Second: make([]int, 0)}
	for j := range m.Len() {
		if first[j] {
			result.First = append(result.First, j+1)
		}
		if second[j] {
			result.Second = append(result.Second, j+1)
		}
	}
	return result, nil
}

// MustMatchSpellings panics with the error returned by MatchSpellings.
func (m Mode) MustMatchSpellings(tonic, a, b spelling.Spelling, cycles CycleWindow) SpellingMembership {
	return must(m.MatchSpellings(tonic, a, b, cycles))
}

// MatchPitches compares registered notes against the requested mode places in
// an explicit system. Invalid or unrepresentable constructions return errors,
// with their original causes. Exact and positional comparisons are distinct.
func (m Mode) MatchPitches(sys *system.System, tonic, a, b pitch.Pitch, by PitchComparison, cycles CycleWindow) (PitchMembership, error) {
	if !m.IsValid() || !sys.IsValid() || !tonic.IsValid() || !a.IsValid() || !b.IsValid() {
		return PitchMembership{}, ErrUnset
	}
	if !cycles.valid() || (by != WrittenPitch && by != PositionalPitch) {
		return PitchMembership{}, ErrInvalidArgument
	}
	for _, p := range []pitch.Pitch{tonic, a, b} {
		ok, err := sys.AdmitsPitch(p)
		if err != nil {
			return PitchMembership{}, err
		}
		if !ok {
			return PitchMembership{}, system.ErrInadmissible
		}
	}
	result := PitchMembership{First: make([]DegreePlace, 0), Second: make([]DegreePlace, 0)}
	compare := func(p, target pitch.Pitch) (bool, error) {
		if by == WrittenPitch {
			return p.Equal(target), nil
		}
		return sys.Equivalent(p, target)
	}
	scan := func(c int) error {
		for j := range m.Len() {
			place := DegreePlace{j + 1, c}
			p, err := m.placePitch(sys, tonic, place)
			if err != nil {
				return err
			}
			x, err := compare(p, a)
			if err != nil {
				return err
			}
			y, err := compare(p, b)
			if err != nil {
				return err
			}
			if x {
				result.First = append(result.First, place)
			}
			if y {
				result.Second = append(result.Second, place)
			}
		}
		return nil
	}
	if !m.IsOctave() {
		if err := m.eachCycle(cycles, scan); err != nil {
			return PitchMembership{}, err
		}
		return result, nil
	}
	// One octave cycle changes the registered octave by exactly one. Find
	// candidate cycles analytically and also validate the window's extremes.
	for _, c := range []int{cycles.First, cycles.Last} {
		for j := range m.Len() {
			if _, err := m.placePitch(sys, tonic, DegreePlace{j + 1, c}); err != nil {
				return PitchMembership{}, err
			}
		}
	}
	// Anchor the arithmetic inside the validated window: cycle zero may
	// require an unrepresentable pitch even when every requested place fits.
	// Keep the displacement and the resulting cycle exact until the final
	// range check, since their individual terms need not fit in an int.
	candidates := make([]int, 0, 2*m.Len())
	for j := range m.Len() {
		base, err := m.placePitch(sys, tonic, DegreePlace{j + 1, cycles.First})
		if err != nil {
			return PitchMembership{}, err
		}
		for _, target := range []pitch.Pitch{a, b} {
			diff := new(big.Int).Sub(big.NewInt(int64(target.MustASPN())), big.NewInt(int64(base.MustASPN())))
			diff.Add(diff, big.NewInt(int64(cycles.First)))
			if diff.IsInt64() {
				c := int(diff.Int64())
				if cycles.contains(c) {
					candidates = append(candidates, c)
				}
			}
		}
	}
	slices.Sort(candidates)
	for _, c := range slices.Compact(candidates) {
		if err := scan(c); err != nil {
			return PitchMembership{}, err
		}
	}
	return result, nil
}

// MustMatchPitches panics with the error returned by MatchPitches.
func (m Mode) MustMatchPitches(sys *system.System, tonic, a, b pitch.Pitch, by PitchComparison, cycles CycleWindow) PitchMembership {
	return must(m.MatchPitches(sys, tonic, a, b, by, cycles))
}
