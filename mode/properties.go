package mode

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/go-muse/musex/interval"
)

// Inclination describes the quality of a mode's unique third, if it has one.
type Inclination uint8

const (
	InclinationUndefined Inclination = iota
	Major
	Minor
)

// IsValid reports whether i is a defined Inclination value.
func (i Inclination) IsValid() bool { return i <= Minor }

// String returns the English name of the inclination.
func (i Inclination) String() string {
	switch i {
	case InclinationUndefined:
		return "undefined"
	case Major:
		return "major"
	case Minor:
		return "minor"
	default:
		return fmt.Sprintf("Inclination(%d)", uint8(i))
	}
}

// IsSecondwise reports whether m is closed and has seven adjacent seconds.
func (m Mode) IsSecondwise() bool {
	if !m.IsClosed() || m.Len() != 7 {
		return false
	}
	for j := 1; j <= 7; j++ {
		if m.structure.MustAdjacentIntervalAfter(j).Size() != 2 {
			return false
		}
	}
	return true
}

// Inclination classifies the unique third from the first degree, if any.
func (m Mode) Inclination() (Inclination, error) {
	if !m.IsValid() {
		return InclinationUndefined, ErrUnset
	}
	var third interval.Interval
	for _, p := range m.prefixes {
		if p.Size() != 3 {
			continue
		}
		if third.IsValid() {
			return InclinationUndefined, nil
		}
		third = p
	}
	switch {
	case third.Equal(interval.MajorThird()):
		return Major, nil
	case third.Equal(interval.MinorThird()):
		return Minor, nil
	default:
		return InclinationUndefined, nil
	}
}

// IsDiatonic reports whether m has one of the seven diatonic structures.
func (m Mode) IsDiatonic() (bool, error) {
	_, ok, err := m.Brightness()
	return ok, err
}

// Brightness ranks diatonic structures from 1 (brightest) to 7 (darkest).
func (m Mode) Brightness() (int, bool, error) {
	if !m.IsValid() {
		return 0, false, ErrUnset
	}
	if !m.IsClosed() || m.Len() != 7 {
		return 0, false, nil
	}
	tone, semitone := interval.MajorSecond(), interval.MinorSecond()
	diatonic := [...]interval.Interval{tone, tone, semitone, tone, tone, tone, semitone}
	for rank, start := range [...]int{3, 0, 4, 1, 5, 2, 6} {
		match := true
		for j := range diatonic {
			if !m.structure.MustAdjacentIntervalAfter(j + 1).Equal(diatonic[(start+j)%len(diatonic)]) {
				match = false
				break
			}
		}
		if match {
			return rank + 1, true, nil
		}
	}
	return 0, false, nil
}

// CompareBrightness sorts brighter modes first. Both must be diatonic.
func (m Mode) CompareBrightness(other Mode) (int, error) {
	if !m.IsValid() || !other.IsValid() {
		return 0, ErrUnset
	}
	first, firstOK := mustPair(m.Brightness())
	second, secondOK := mustPair(other.Brightness())
	if !firstOK || !secondOK {
		return 0, ErrIncomparable
	}
	return cmp.Compare(first, second), nil
}

// Tetrachord describes three adjacent intervals and any recognized names.
type Tetrachord struct {
	Intervals   [3]interval.Interval
	RussianName string
	EnglishName string
}

// TetrachordPair divides a seven-degree closed mode into two tetrachords.
type TetrachordPair struct {
	Lower     Tetrachord
	Connector interval.Interval
	Upper     Tetrachord
}

// SeparatedByTone reports whether the connector is a major second.
func (p TetrachordPair) SeparatedByTone() bool {
	return p.Connector.Equal(interval.MajorSecond())
}

// Tetrachords divides a seven-degree closed mode into two groups and a connector.
func (m Mode) Tetrachords() (TetrachordPair, error) {
	if err := m.checkClosed(); err != nil {
		return TetrachordPair{}, err
	}
	if m.Len() != 7 {
		return TetrachordPair{}, ErrWrongDegreeCount
	}
	a := m.AdjacentIntervals()
	return TetrachordPair{tetrachord([3]interval.Interval(a[:3])), a[3], tetrachord([3]interval.Interval(a[4:]))}, nil
}

func tetrachord(a [3]interval.Interval) Tetrachord {
	tone, semi, augmented := interval.MajorSecond(), interval.MinorSecond(), interval.AugmentedSecond()
	result := Tetrachord{Intervals: a}
	switch a {
	case [3]interval.Interval{semi, tone, tone}:
		result.RussianName, result.EnglishName = "фригийский", "Upper Minor"
	case [3]interval.Interval{tone, semi, tone}:
		result.RussianName, result.EnglishName = "дорийский", "Minor"
	case [3]interval.Interval{tone, tone, semi}:
		result.RussianName, result.EnglishName = "ионийский", "Major"
	case [3]interval.Interval{tone, tone, tone}:
		result.RussianName = "лидийский"
	case [3]interval.Interval{semi, augmented, semi}:
		result.EnglishName = "Harmonic"
	}
	return result
}

// leastPeriod uses the longest proper prefix that is also a suffix. It is
// linear even when a long common prefix occurs at many different offsets.
func leastPeriod[T comparable](sequence []T) int {
	length := len(sequence)
	prefix := make([]int, length)
	for i := 1; i < length; i++ {
		j := prefix[i-1]
		for j > 0 && sequence[i] != sequence[j] {
			j = prefix[j-1]
		}
		if sequence[i] == sequence[j] {
			j++
		}
		prefix[i] = j
	}
	p := length - prefix[length-1]
	if length%p != 0 {
		return length
	}
	return p
}

// Period returns the least positive number of degrees in a repeating part.
func (m Mode) Period() (int, error) {
	if err := m.checkClosed(); err != nil {
		return 0, err
	}
	return leastPeriod(m.AdjacentIntervals()), nil
}

// DistinctRotations counts the structurally distinct rotations.
func (m Mode) DistinctRotations() (int, error) { return m.Period() }

// MinimalPart returns the shortest repeating part and its repetition count.
func (m Mode) MinimalPart() ([]interval.Interval, int, error) {
	if err := m.checkClosed(); err != nil {
		return nil, 0, err
	}
	adjacent := m.AdjacentIntervals()
	p := leastPeriod(adjacent)
	return slices.Clip(adjacent[:p]), m.Len() / p, nil
}

// IsPalindrome reports whether the adjacent intervals read the same backward.
func (m Mode) IsPalindrome() bool {
	if !m.IsValid() {
		return false
	}
	a := m.AdjacentIntervals()
	for i := 0; i < len(a)/2; i++ {
		if !a[i].Equal(a[len(a)-1-i]) {
			return false
		}
	}
	return true
}
