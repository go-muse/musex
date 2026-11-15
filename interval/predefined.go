package interval

import "github.com/go-muse/musex/alteration"

// ready returns the ascending interval of size n with the whole deviation d.
// It is used only with the constant arguments of the ready-made intervals.
func ready(n, d int) Interval {
	return Interval{size: n, dev: alteration.MustFromInt(d)}
}

// PerfectUnison returns the ascending perfect unison (P1).
func PerfectUnison() Interval { return ready(1, 0) }

// AugmentedUnison returns the ascending augmented unison (A1).
func AugmentedUnison() Interval { return ready(1, 1) }

// DoublyAugmentedUnison returns the ascending doubly augmented unison (AA1).
func DoublyAugmentedUnison() Interval { return ready(1, 2) }

// DiminishedSecond returns the ascending diminished second (d2).
func DiminishedSecond() Interval { return ready(2, -2) }

// MinorSecond returns the ascending minor second (m2).
func MinorSecond() Interval { return ready(2, -1) }

// MajorSecond returns the ascending major second (M2).
func MajorSecond() Interval { return ready(2, 0) }

// AugmentedSecond returns the ascending augmented second (A2).
func AugmentedSecond() Interval { return ready(2, 1) }

// DiminishedThird returns the ascending diminished third (d3).
func DiminishedThird() Interval { return ready(3, -2) }

// MinorThird returns the ascending minor third (m3).
func MinorThird() Interval { return ready(3, -1) }

// MajorThird returns the ascending major third (M3).
func MajorThird() Interval { return ready(3, 0) }

// AugmentedThird returns the ascending augmented third (A3).
func AugmentedThird() Interval { return ready(3, 1) }

// DiminishedFourth returns the ascending diminished fourth (d4).
func DiminishedFourth() Interval { return ready(4, -1) }

// PerfectFourth returns the ascending perfect fourth (P4).
func PerfectFourth() Interval { return ready(4, 0) }

// AugmentedFourth returns the ascending augmented fourth (A4).
func AugmentedFourth() Interval { return ready(4, 1) }

// DoublyAugmentedFourth returns the ascending doubly augmented fourth (AA4).
func DoublyAugmentedFourth() Interval { return ready(4, 2) }

// DoublyDiminishedFifth returns the ascending doubly diminished fifth (dd5).
func DoublyDiminishedFifth() Interval { return ready(5, -2) }

// DiminishedFifth returns the ascending diminished fifth (d5).
func DiminishedFifth() Interval { return ready(5, -1) }

// PerfectFifth returns the ascending perfect fifth (P5).
func PerfectFifth() Interval { return ready(5, 0) }

// AugmentedFifth returns the ascending augmented fifth (A5).
func AugmentedFifth() Interval { return ready(5, 1) }

// DiminishedSixth returns the ascending diminished sixth (d6).
func DiminishedSixth() Interval { return ready(6, -2) }

// MinorSixth returns the ascending minor sixth (m6).
func MinorSixth() Interval { return ready(6, -1) }

// MajorSixth returns the ascending major sixth (M6).
func MajorSixth() Interval { return ready(6, 0) }

// AugmentedSixth returns the ascending augmented sixth (A6).
func AugmentedSixth() Interval { return ready(6, 1) }

// DiminishedSeventh returns the ascending diminished seventh (d7).
func DiminishedSeventh() Interval { return ready(7, -2) }

// MinorSeventh returns the ascending minor seventh (m7).
func MinorSeventh() Interval { return ready(7, -1) }

// MajorSeventh returns the ascending major seventh (M7).
func MajorSeventh() Interval { return ready(7, 0) }

// AugmentedSeventh returns the ascending augmented seventh (A7).
func AugmentedSeventh() Interval { return ready(7, 1) }

// DoublyDiminishedOctave returns the ascending doubly diminished octave (dd8).
func DoublyDiminishedOctave() Interval { return ready(8, -2) }

// DiminishedOctave returns the ascending diminished octave (d8).
func DiminishedOctave() Interval { return ready(8, -1) }

// PerfectOctave returns the ascending perfect octave (P8).
func PerfectOctave() Interval { return ready(8, 0) }

// AugmentedOctave returns the ascending augmented octave (A8).
func AugmentedOctave() Interval { return ready(8, 1) }

// DoublyAugmentedOctave returns the ascending doubly augmented octave (AA8).
func DoublyAugmentedOctave() Interval { return ready(8, 2) }

// DiminishedNinth returns the ascending diminished ninth (d9).
func DiminishedNinth() Interval { return ready(9, -2) }

// MinorNinth returns the ascending minor ninth (m9).
func MinorNinth() Interval { return ready(9, -1) }

// MajorNinth returns the ascending major ninth (M9).
func MajorNinth() Interval { return ready(9, 0) }

// AugmentedNinth returns the ascending augmented ninth (A9).
func AugmentedNinth() Interval { return ready(9, 1) }

// DiminishedTenth returns the ascending diminished tenth (d10).
func DiminishedTenth() Interval { return ready(10, -2) }

// MinorTenth returns the ascending minor tenth (m10).
func MinorTenth() Interval { return ready(10, -1) }

// MajorTenth returns the ascending major tenth (M10).
func MajorTenth() Interval { return ready(10, 0) }

// AugmentedTenth returns the ascending augmented tenth (A10).
func AugmentedTenth() Interval { return ready(10, 1) }

// DiminishedEleventh returns the ascending diminished eleventh (d11).
func DiminishedEleventh() Interval { return ready(11, -1) }

// PerfectEleventh returns the ascending perfect eleventh (P11).
func PerfectEleventh() Interval { return ready(11, 0) }

// AugmentedEleventh returns the ascending augmented eleventh (A11).
func AugmentedEleventh() Interval { return ready(11, 1) }

// DiminishedTwelfth returns the ascending diminished twelfth (d12).
func DiminishedTwelfth() Interval { return ready(12, -1) }

// PerfectTwelfth returns the ascending perfect twelfth (P12).
func PerfectTwelfth() Interval { return ready(12, 0) }

// AugmentedTwelfth returns the ascending augmented twelfth (A12).
func AugmentedTwelfth() Interval { return ready(12, 1) }

// DiminishedThirteenth returns the ascending diminished thirteenth (d13).
func DiminishedThirteenth() Interval { return ready(13, -2) }

// MinorThirteenth returns the ascending minor thirteenth (m13).
func MinorThirteenth() Interval { return ready(13, -1) }

// MajorThirteenth returns the ascending major thirteenth (M13).
func MajorThirteenth() Interval { return ready(13, 0) }

// AugmentedThirteenth returns the ascending augmented thirteenth (A13).
func AugmentedThirteenth() Interval { return ready(13, 1) }

// DiminishedFourteenth returns the ascending diminished fourteenth (d14).
func DiminishedFourteenth() Interval { return ready(14, -2) }

// MinorFourteenth returns the ascending minor fourteenth (m14).
func MinorFourteenth() Interval { return ready(14, -1) }

// MajorFourteenth returns the ascending major fourteenth (M14).
func MajorFourteenth() Interval { return ready(14, 0) }

// AugmentedFourteenth returns the ascending augmented fourteenth (A14).
func AugmentedFourteenth() Interval { return ready(14, 1) }

// DiminishedFifteenth returns the ascending diminished fifteenth (d15).
func DiminishedFifteenth() Interval { return ready(15, -1) }

// PerfectFifteenth returns the ascending perfect fifteenth (P15).
func PerfectFifteenth() Interval { return ready(15, 0) }

// AugmentedFifteenth returns the ascending augmented fifteenth (A15).
func AugmentedFifteenth() Interval { return ready(15, 1) }

// Predefined returns the 57 ready-made ascending intervals, in the order of
// size and, for one size, of deviation from the smallest: P1, A1, AA1, d2,
// m2, ..., d15, P15, A15. For each size from 1 to 15 they are the basic kinds
// with the augmented and diminished intervals (except the diminished unison),
// together with the doubly augmented unison, fourth, and octave and the doubly
// diminished fifth and octave.
//
// Each call returns a new slice with length and capacity 57. The caller owns
// it and may modify or retain it; this does not affect the ready-made
// intervals or later calls.
func Predefined() []Interval {
	return []Interval{
		PerfectUnison(), AugmentedUnison(), DoublyAugmentedUnison(), DiminishedSecond(),
		MinorSecond(), MajorSecond(), AugmentedSecond(), DiminishedThird(),
		MinorThird(), MajorThird(), AugmentedThird(), DiminishedFourth(),
		PerfectFourth(), AugmentedFourth(), DoublyAugmentedFourth(), DoublyDiminishedFifth(),
		DiminishedFifth(), PerfectFifth(), AugmentedFifth(), DiminishedSixth(),
		MinorSixth(), MajorSixth(), AugmentedSixth(), DiminishedSeventh(),
		MinorSeventh(), MajorSeventh(), AugmentedSeventh(), DoublyDiminishedOctave(),
		DiminishedOctave(), PerfectOctave(), AugmentedOctave(), DoublyAugmentedOctave(),
		DiminishedNinth(), MinorNinth(), MajorNinth(), AugmentedNinth(),
		DiminishedTenth(), MinorTenth(), MajorTenth(), AugmentedTenth(),
		DiminishedEleventh(), PerfectEleventh(), AugmentedEleventh(), DiminishedTwelfth(),
		PerfectTwelfth(), AugmentedTwelfth(), DiminishedThirteenth(), MinorThirteenth(),
		MajorThirteenth(), AugmentedThirteenth(), DiminishedFourteenth(), MinorFourteenth(),
		MajorFourteenth(), AugmentedFourteenth(), DiminishedFifteenth(), PerfectFifteenth(),
		AugmentedFifteenth(),
	}
}
