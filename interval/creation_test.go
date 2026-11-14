package interval_test

import (
	"math"
	"slices"
	"strings"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/interval"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T-1: creation, predefined intervals, reading, properties, directions,
// comparison and enumerations (OP-1...OP-10, OP-18...OP-20; P-2...P-5, P-7,
// P-10, P-23; architecture 4.3...4.5, 4.7).

// ---------------------------------------------------------------------------
// Creation (OP-1, OP-2; architecture 4.4.1).

func TestCreationNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		q    interval.Quality
		n    int
		dev  alteration.Alteration
	}{
		{"perfect unison", interval.Perfect, 1, ai(0)},
		{"augmented unison", interval.Augmented, 1, ai(1)},
		{"minor second", interval.Minor, 2, ai(-1)},
		{"major second", interval.Major, 2, ai(0)},
		{"diminished second", interval.Diminished, 2, ai(-2)},
		{"neutral second", interval.Neutral, 2, alt(-1, 2)},
		{"major third", interval.Major, 3, ai(0)},
		{"neutral third", interval.Neutral, 3, alt(-1, 2)},
		{"perfect fourth", interval.Perfect, 4, ai(0)},
		{"augmented fourth", interval.Augmented, 4, ai(1)},
		{"diminished fourth", interval.Diminished, 4, ai(-1)},
		{"diminished fifth", interval.Diminished, 5, ai(-1)},
		{"minor seventh", interval.Minor, 7, ai(-1)},
		{"perfect octave", interval.Perfect, 8, ai(0)},
		{"diminished octave", interval.Diminished, 8, ai(-1)},
		{"augmented octave", interval.Augmented, 8, ai(1)},
		{"major ninth", interval.Major, 9, ai(0)},
		{"neutral tenth", interval.Neutral, 10, alt(-1, 2)},
		{"perfect eleventh", interval.Perfect, 11, ai(0)},
		{"diminished twelfth", interval.Diminished, 12, ai(-1)},
		{"perfect fifteenth", interval.Perfect, 15, ai(0)},
		{"major sixteenth", interval.Major, 16, ai(0)},
		{"perfect twenty-second", interval.Perfect, 22, ai(0)},
		{"major of MaxSize", interval.Major, interval.MaxSize, ai(0)},
		{"diminished of MaxSize", interval.Diminished, interval.MaxSize, ai(-2)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := interval.New(tt.q, tt.n)
			require.NoError(t, err)
			assert.True(t, got.IsValid())
			assert.True(t, got.IsAscending())
			assert.Equal(t, interval.Up, got.Direction())
			assert.Equal(t, tt.n, got.Size())
			assert.Equal(t, tt.dev, got.Deviation())
			assert.True(t, got.Equal(interval.MustFromDeviation(tt.n, tt.dev)))
		})
	}
}

func TestCreationNewErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		q    interval.Quality
		n    int
		want error
	}{
		{"unset quality", interval.QualityUnset, 3, interval.ErrUnset},
		{"unset quality before size 0", interval.QualityUnset, 0, interval.ErrUnset},
		{"size 0", interval.Major, 0, interval.ErrInvalidArgument},
		{"negative size", interval.Perfect, -1, interval.ErrInvalidArgument},
		{"size MinInt", interval.Perfect, minInt, interval.ErrInvalidArgument},
		{"fractional quality", interval.Fractional, 3, interval.ErrInvalidArgument},
		{"quality out of the enumeration", interval.Quality(8), 3, interval.ErrInvalidArgument},
		{"quality 255", interval.Quality(255), 4, interval.ErrInvalidArgument},
		{"perfect third", interval.Perfect, 3, interval.ErrInvalidArgument},
		{"perfect seventh", interval.Perfect, 7, interval.ErrInvalidArgument},
		{"perfect ninth", interval.Perfect, 9, interval.ErrInvalidArgument},
		{"major fifth", interval.Major, 5, interval.ErrInvalidArgument},
		{"major unison", interval.Major, 1, interval.ErrInvalidArgument},
		{"minor fourth", interval.Minor, 4, interval.ErrInvalidArgument},
		{"minor octave", interval.Minor, 8, interval.ErrInvalidArgument},
		{"neutral fourth", interval.Neutral, 4, interval.ErrInvalidArgument},
		{"neutral unison", interval.Neutral, 1, interval.ErrInvalidArgument},
		{"neutral fifteenth", interval.Neutral, 15, interval.ErrInvalidArgument},
		{"diminished unison", interval.Diminished, 1, interval.ErrInvalidArgument},
		{"perfect of MaxSize", interval.Perfect, interval.MaxSize, interval.ErrInvalidArgument},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := interval.New(tt.q, tt.n)
			assertErr(t, err, tt.want)
			assert.Equal(t, interval.Interval{}, got)
		})
	}
}

func TestCreationNewMultiple(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		k    int
		q    interval.Quality
		n    int
		dev  alteration.Alteration
	}{
		{"augmented unison", 1, interval.Augmented, 1, ai(1)},
		{"doubly augmented unison", 2, interval.Augmented, 1, ai(2)},
		{"doubly augmented fourth", 2, interval.Augmented, 4, ai(2)},
		{"triply augmented third", 3, interval.Augmented, 3, ai(3)},
		{"doubly diminished fifth", 2, interval.Diminished, 5, ai(-2)},
		{"doubly diminished third", 2, interval.Diminished, 3, ai(-3)},
		{"diminished octave", 1, interval.Diminished, 8, ai(-1)},
		{"doubly diminished octave", 2, interval.Diminished, 8, ai(-2)},
		{"diminished second", 1, interval.Diminished, 2, ai(-2)},
		{"augmented of multiplicity 5", 5, interval.Augmented, 6, ai(5)},
		{"augmented of multiplicity M", mag, interval.Augmented, 4, ai(mag)},
		{"diminished perfect of multiplicity M", mag, interval.Diminished, 5, ai(-mag)},
		{"diminished major of multiplicity M-1", mag - 1, interval.Diminished, 3, ai(-mag)},
		{"augmented of MaxSize", 3, interval.Augmented, interval.MaxSize, ai(3)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := interval.NewMultiple(tt.k, tt.q, tt.n)
			require.NoError(t, err)
			assert.True(t, got.IsAscending())
			assert.Equal(t, tt.n, got.Size())
			assert.Equal(t, tt.dev, got.Deviation())
			assert.Equal(t, tt.k, got.Multiplicity())
			assert.Equal(t, tt.q, got.Quality())
		})
	}

	// New(q, n) equals NewMultiple(1, q, n) for Augmented and Diminished.
	for n := 1; n <= 16; n++ {
		for _, q := range []interval.Quality{interval.Augmented, interval.Diminished} {
			a, errA := interval.New(q, n)
			b, errB := interval.NewMultiple(1, q, n)

			if n == 1 && q == interval.Diminished {
				assertErr(t, errA, interval.ErrInvalidArgument)
				assertErr(t, errB, interval.ErrInvalidArgument)

				continue
			}

			require.NoError(t, errA)
			require.NoError(t, errB)
			assert.True(t, a.Equal(b), "%v %d", q, n)
		}
	}
}

func TestCreationNewMultipleErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		k      int
		q      interval.Quality
		n      int
		want   error
		causes []error
	}{
		{"unset quality", 2, interval.QualityUnset, 4, interval.ErrUnset, nil},
		{"unset quality before everything", 0, interval.QualityUnset, 0, interval.ErrUnset, nil},
		{"size 0", 2, interval.Augmented, 0, interval.ErrInvalidArgument, nil},
		{"multiplicity 0", 0, interval.Augmented, 4, interval.ErrInvalidArgument, nil},
		{"negative multiplicity", -1, interval.Diminished, 5, interval.ErrInvalidArgument, nil},
		{"multiplicity MinInt", minInt, interval.Augmented, 4, interval.ErrInvalidArgument, nil},
		{"perfect", 2, interval.Perfect, 5, interval.ErrInvalidArgument, nil},
		{"major", 1, interval.Major, 3, interval.ErrInvalidArgument, nil},
		{"minor", 1, interval.Minor, 3, interval.ErrInvalidArgument, nil},
		{"neutral", 1, interval.Neutral, 3, interval.ErrInvalidArgument, nil},
		{"fractional", 1, interval.Fractional, 3, interval.ErrInvalidArgument, nil},
		{"quality out of the enumeration", 1, interval.Quality(42), 3, interval.ErrInvalidArgument, nil},
		{"diminished unison", 2, interval.Diminished, 1, interval.ErrInvalidArgument, nil},
		{"diminished unison before out of range", mag + 1, interval.Diminished, 1, interval.ErrInvalidArgument, nil},
		{"size 0 before out of range", maxInt, interval.Augmented, 0, interval.ErrInvalidArgument, nil},
		{"augmented beyond M", mag + 1, interval.Augmented, 4,
			interval.ErrOutOfRange, []error{alteration.ErrOutOfRange}},
		{"augmented of MaxInt", maxInt, interval.Augmented, 4,
			interval.ErrOutOfRange, []error{alteration.ErrOutOfRange}},
		{"diminished perfect beyond M", mag + 1, interval.Diminished, 5,
			interval.ErrOutOfRange, []error{alteration.ErrOutOfRange}},
		{"diminished major of multiplicity M", mag, interval.Diminished, 3,
			interval.ErrOutOfRange, []error{alteration.ErrOutOfRange}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := interval.NewMultiple(tt.k, tt.q, tt.n)
			assertErr(t, err, tt.want, tt.causes...)
			assert.Equal(t, interval.Interval{}, got)
		})
	}
}

func TestCreationFromDeviation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		n    int
		dev  alteration.Alteration
		q    interval.Quality
	}{
		{"perfect unison", 1, ai(0), interval.Perfect},
		{"unison a quarter wider", 1, alt(1, 4), interval.Fractional},
		{"augmented unison", 1, ai(1), interval.Augmented},
		{"neutral third", 3, alt(-1, 2), interval.Neutral},
		{"third a third wider than major", 3, alt(1, 3), interval.Fractional},
		{"fifth a half narrower", 5, alt(-1, 2), interval.Fractional},
		{"deviation M", 4, ai(mag), interval.Augmented},
		{"deviation -M", 6, ai(-mag), interval.Diminished},
		{"size MaxSize", interval.MaxSize, ai(-1), interval.Minor},
		{"zero alteration value", 3, alteration.Alteration{}, interval.Major},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := interval.FromDeviation(tt.n, tt.dev)
			require.NoError(t, err)
			assert.True(t, got.IsAscending())
			assert.Equal(t, tt.n, got.Size())
			assert.Equal(t, tt.dev, got.Deviation())
			assert.Equal(t, tt.q, got.Quality())
		})
	}
}

func TestCreationFromDeviationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		n    int
		dev  alteration.Alteration
	}{
		{"size 0", 0, ai(0)},
		{"negative size", -3, ai(0)},
		{"size MinInt", minInt, ai(1)},
		{"unison narrower by 1", 1, ai(-1)},
		{"unison narrower by 1/2", 1, alt(-1, 2)},
		{"unison narrower by M", 1, ai(-mag)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := interval.FromDeviation(tt.n, tt.dev)
			assertErr(t, err, interval.ErrInvalidArgument)
			assert.Equal(t, interval.Interval{}, got)
		})
	}
}

// ---------------------------------------------------------------------------
// Predefined intervals (OP-3, R-20; architecture 4.4.2).

// predefinedTable is table 4.4.2 in its order: by size, then by deviation.
var predefinedTable = []struct {
	name     string
	fn       func() interval.Interval
	notation string
}{
	{"PerfectUnison", interval.PerfectUnison, "P1"},
	{"AugmentedUnison", interval.AugmentedUnison, "A1"},
	{"DoublyAugmentedUnison", interval.DoublyAugmentedUnison, "AA1"},
	{"DiminishedSecond", interval.DiminishedSecond, "d2"},
	{"MinorSecond", interval.MinorSecond, "m2"},
	{"MajorSecond", interval.MajorSecond, "M2"},
	{"AugmentedSecond", interval.AugmentedSecond, "A2"},
	{"DiminishedThird", interval.DiminishedThird, "d3"},
	{"MinorThird", interval.MinorThird, "m3"},
	{"MajorThird", interval.MajorThird, "M3"},
	{"AugmentedThird", interval.AugmentedThird, "A3"},
	{"DiminishedFourth", interval.DiminishedFourth, "d4"},
	{"PerfectFourth", interval.PerfectFourth, "P4"},
	{"AugmentedFourth", interval.AugmentedFourth, "A4"},
	{"DoublyAugmentedFourth", interval.DoublyAugmentedFourth, "AA4"},
	{"DoublyDiminishedFifth", interval.DoublyDiminishedFifth, "dd5"},
	{"DiminishedFifth", interval.DiminishedFifth, "d5"},
	{"PerfectFifth", interval.PerfectFifth, "P5"},
	{"AugmentedFifth", interval.AugmentedFifth, "A5"},
	{"DiminishedSixth", interval.DiminishedSixth, "d6"},
	{"MinorSixth", interval.MinorSixth, "m6"},
	{"MajorSixth", interval.MajorSixth, "M6"},
	{"AugmentedSixth", interval.AugmentedSixth, "A6"},
	{"DiminishedSeventh", interval.DiminishedSeventh, "d7"},
	{"MinorSeventh", interval.MinorSeventh, "m7"},
	{"MajorSeventh", interval.MajorSeventh, "M7"},
	{"AugmentedSeventh", interval.AugmentedSeventh, "A7"},
	{"DoublyDiminishedOctave", interval.DoublyDiminishedOctave, "dd8"},
	{"DiminishedOctave", interval.DiminishedOctave, "d8"},
	{"PerfectOctave", interval.PerfectOctave, "P8"},
	{"AugmentedOctave", interval.AugmentedOctave, "A8"},
	{"DoublyAugmentedOctave", interval.DoublyAugmentedOctave, "AA8"},
	{"DiminishedNinth", interval.DiminishedNinth, "d9"},
	{"MinorNinth", interval.MinorNinth, "m9"},
	{"MajorNinth", interval.MajorNinth, "M9"},
	{"AugmentedNinth", interval.AugmentedNinth, "A9"},
	{"DiminishedTenth", interval.DiminishedTenth, "d10"},
	{"MinorTenth", interval.MinorTenth, "m10"},
	{"MajorTenth", interval.MajorTenth, "M10"},
	{"AugmentedTenth", interval.AugmentedTenth, "A10"},
	{"DiminishedEleventh", interval.DiminishedEleventh, "d11"},
	{"PerfectEleventh", interval.PerfectEleventh, "P11"},
	{"AugmentedEleventh", interval.AugmentedEleventh, "A11"},
	{"DiminishedTwelfth", interval.DiminishedTwelfth, "d12"},
	{"PerfectTwelfth", interval.PerfectTwelfth, "P12"},
	{"AugmentedTwelfth", interval.AugmentedTwelfth, "A12"},
	{"DiminishedThirteenth", interval.DiminishedThirteenth, "d13"},
	{"MinorThirteenth", interval.MinorThirteenth, "m13"},
	{"MajorThirteenth", interval.MajorThirteenth, "M13"},
	{"AugmentedThirteenth", interval.AugmentedThirteenth, "A13"},
	{"DiminishedFourteenth", interval.DiminishedFourteenth, "d14"},
	{"MinorFourteenth", interval.MinorFourteenth, "m14"},
	{"MajorFourteenth", interval.MajorFourteenth, "M14"},
	{"AugmentedFourteenth", interval.AugmentedFourteenth, "A14"},
	{"DiminishedFifteenth", interval.DiminishedFifteenth, "d15"},
	{"PerfectFifteenth", interval.PerfectFifteenth, "P15"},
	{"AugmentedFifteenth", interval.AugmentedFifteenth, "A15"},
}

// createdBy builds the interval of a notation without sign of direction
// through MustNew or MustNewMultiple, as table 4.4.2 requires.
func createdBy(notation string) interval.Interval {
	at := strings.IndexAny(notation, "0123456789")
	sign := notation[:at]
	_, n := devOf(notation)

	switch {
	case sign == "P":
		return interval.MustNew(interval.Perfect, n)
	case sign == "M":
		return interval.MustNew(interval.Major, n)
	case sign == "m":
		return interval.MustNew(interval.Minor, n)
	case strings.Trim(sign, "A") == "":
		return interval.MustNewMultiple(len(sign), interval.Augmented, n)
	default:
		return interval.MustNewMultiple(len(sign), interval.Diminished, n)
	}
}

func TestCreationPredefined(t *testing.T) {
	t.Parallel()

	require.Len(t, predefinedTable, 57)

	for _, tt := range predefinedTable {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.fn()
			assert.True(t, got.IsValid())
			assert.True(t, got.IsAscending())
			assert.True(t, got.Equal(createdBy(tt.notation)), "%s is %s", tt.name, tt.notation)
			assert.True(t, got.Equal(iv(tt.notation)), "%s is %s", tt.name, tt.notation)
			assert.True(t, got.Equal(tt.fn()), "every call gives the same interval")
		})
	}
}

func TestCreationPredefinedSlice(t *testing.T) {
	t.Parallel()

	all := interval.Predefined()
	require.Len(t, all, 57)
	assert.Equal(t, 57, cap(all))

	for k, tt := range predefinedTable {
		assert.True(t, all[k].Equal(tt.fn()), "Predefined()[%d] is %s", k, tt.name)
	}

	// The order of 4.4.2: by size, at one size by deviation ascending.
	for k := 1; k < len(all); k++ {
		prev, cur := all[k-1], all[k]
		if prev.Size() == cur.Size() {
			assert.Equal(t, -1, prev.Deviation().Compare(cur.Deviation()), "order at %d", k)
		} else {
			assert.Less(t, prev.Size(), cur.Size(), "order at %d", k)
		}
	}

	// The slice belongs to the caller.
	all[0] = interval.Interval{}
	all[56] = interval.MajorThird()
	again := interval.Predefined()
	assert.True(t, again[0].Equal(interval.PerfectUnison()))
	assert.True(t, again[56].Equal(interval.AugmentedFifteenth()))
	assert.True(t, interval.PerfectUnison().IsValid())

	// No predefined tritone (R-20): A4 and d5 are both present, nothing else
	// of their simple sizes with deviation between.
	assert.True(t, slices.ContainsFunc(again, interval.AugmentedFourth().Equal))
	assert.True(t, slices.ContainsFunc(again, interval.DiminishedFifth().Equal))
}

// ---------------------------------------------------------------------------
// Reading (OP-4...OP-8; architecture 4.5.1).

func TestReadingParts(t *testing.T) {
	t.Parallel()

	maxOctaves := (interval.MaxSize - maxSizeSimple) / 7

	tests := []struct {
		name    string
		i       interval.Interval
		dir     interval.Direction
		size    int
		dev     alteration.Alteration
		q       interval.Quality
		mult    int
		integer bool
		group   interval.Group
		simple  int
		octaves int
	}{
		{"P1", iv("P1"), interval.Up, 1, ai(0), interval.Perfect, 0, true, interval.PerfectGroup, 1, 0},
		{"P1 made descending is ascending (P-4)", iv("↓P1"), interval.Up, 1, ai(0), interval.Perfect, 0, true, interval.PerfectGroup, 1, 0},
		{"A1", iv("A1"), interval.Up, 1, ai(1), interval.Augmented, 1, true, interval.PerfectGroup, 1, 0},
		{"descending A1", iv("↓A1"), interval.Down, 1, ai(1), interval.Augmented, 1, true, interval.PerfectGroup, 1, 0},
		{"AA1", iv("AA1"), interval.Up, 1, ai(2), interval.Augmented, 2, true, interval.PerfectGroup, 1, 0},
		{"unison +1/2", iv("{+1/2}1"), interval.Up, 1, alt(1, 2), interval.Fractional, 0, false, interval.PerfectGroup, 1, 0},
		{"descending unison +1/2", iv("↓{+1/2}1"), interval.Down, 1, alt(1, 2), interval.Fractional, 0, false, interval.PerfectGroup, 1, 0},
		{"m2", iv("m2"), interval.Up, 2, ai(-1), interval.Minor, 0, true, interval.MajorMinorGroup, 2, 0},
		{"M2", iv("M2"), interval.Up, 2, ai(0), interval.Major, 0, true, interval.MajorMinorGroup, 2, 0},
		{"d2", iv("d2"), interval.Up, 2, ai(-2), interval.Diminished, 1, true, interval.MajorMinorGroup, 2, 0},
		{"dd2", iv("dd2"), interval.Up, 2, ai(-3), interval.Diminished, 2, true, interval.MajorMinorGroup, 2, 0},
		{"A2", iv("A2"), interval.Up, 2, ai(1), interval.Augmented, 1, true, interval.MajorMinorGroup, 2, 0},
		{"n3", iv("n3"), interval.Up, 3, alt(-1, 2), interval.Neutral, 0, false, interval.MajorMinorGroup, 3, 0},
		{"descending n3", iv("↓n3"), interval.Down, 3, alt(-1, 2), interval.Neutral, 0, false, interval.MajorMinorGroup, 3, 0},
		{"third +1/2", iv("{+1/2}3"), interval.Up, 3, alt(1, 2), interval.Fractional, 0, false, interval.MajorMinorGroup, 3, 0},
		{"third -3/2", iv("{-3/2}3"), interval.Up, 3, alt(-3, 2), interval.Fractional, 0, false, interval.MajorMinorGroup, 3, 0},
		{"third -1/4", iv("{-1/4}3"), interval.Up, 3, alt(-1, 4), interval.Fractional, 0, false, interval.MajorMinorGroup, 3, 0},
		{"third -3/4", iv("{-3/4}3"), interval.Up, 3, alt(-3, 4), interval.Fractional, 0, false, interval.MajorMinorGroup, 3, 0},
		{"M3", iv("M3"), interval.Up, 3, ai(0), interval.Major, 0, true, interval.MajorMinorGroup, 3, 0},
		{"descending M3", iv("↓M3"), interval.Down, 3, ai(0), interval.Major, 0, true, interval.MajorMinorGroup, 3, 0},
		{"AAA3", iv("AAA3"), interval.Up, 3, ai(3), interval.Augmented, 3, true, interval.MajorMinorGroup, 3, 0},
		{"fourth +1/2 is not neutral", iv("{+1/2}4"), interval.Up, 4, alt(1, 2), interval.Fractional, 0, false, interval.PerfectGroup, 4, 0},
		{"AA4", iv("AA4"), interval.Up, 4, ai(2), interval.Augmented, 2, true, interval.PerfectGroup, 4, 0},
		{"AAAA4", iv("AAAA4"), interval.Up, 4, ai(4), interval.Augmented, 4, true, interval.PerfectGroup, 4, 0},
		{"d5", iv("d5"), interval.Up, 5, ai(-1), interval.Diminished, 1, true, interval.PerfectGroup, 5, 0},
		{"dd5", iv("dd5"), interval.Up, 5, ai(-2), interval.Diminished, 2, true, interval.PerfectGroup, 5, 0},
		{"fifth -1/2 is not neutral", iv("{-1/2}5"), interval.Up, 5, alt(-1, 2), interval.Fractional, 0, false, interval.PerfectGroup, 5, 0},
		{"m6", iv("m6"), interval.Up, 6, ai(-1), interval.Minor, 0, true, interval.MajorMinorGroup, 6, 0},
		{"third -5", iv("{-5}3"), interval.Up, 3, ai(-5), interval.Diminished, 4, true, interval.MajorMinorGroup, 3, 0},
		{"P8", iv("P8"), interval.Up, 8, ai(0), interval.Perfect, 0, true, interval.PerfectGroup, 8, 0},
		{"descending P8", iv("↓P8"), interval.Down, 8, ai(0), interval.Perfect, 0, true, interval.PerfectGroup, 8, 0},
		{"d8", iv("d8"), interval.Up, 8, ai(-1), interval.Diminished, 1, true, interval.PerfectGroup, 8, 0},
		{"A8", iv("A8"), interval.Up, 8, ai(1), interval.Augmented, 1, true, interval.PerfectGroup, 8, 0},
		{"M9", iv("M9"), interval.Up, 9, ai(0), interval.Major, 0, true, interval.MajorMinorGroup, 2, 1},
		{"M10", iv("M10"), interval.Up, 10, ai(0), interval.Major, 0, true, interval.MajorMinorGroup, 3, 1},
		{"n10", iv("n10"), interval.Up, 10, alt(-1, 2), interval.Neutral, 0, false, interval.MajorMinorGroup, 3, 1},
		{"A11", iv("A11"), interval.Up, 11, ai(1), interval.Augmented, 1, true, interval.PerfectGroup, 4, 1},
		{"d12", iv("d12"), interval.Up, 12, ai(-1), interval.Diminished, 1, true, interval.PerfectGroup, 5, 1},
		{"P15", iv("P15"), interval.Up, 15, ai(0), interval.Perfect, 0, true, interval.PerfectGroup, 8, 1},
		{"d15", iv("d15"), interval.Up, 15, ai(-1), interval.Diminished, 1, true, interval.PerfectGroup, 8, 1},
		{"M16", iv("M16"), interval.Up, 16, ai(0), interval.Major, 0, true, interval.MajorMinorGroup, 2, 2},
		{"M17", iv("M17"), interval.Up, 17, ai(0), interval.Major, 0, true, interval.MajorMinorGroup, 3, 2},
		{"P22", iv("P22"), interval.Up, 22, ai(0), interval.Perfect, 0, true, interval.PerfectGroup, 8, 2},
		{"descending P22", iv("↓P22"), interval.Down, 22, ai(0), interval.Perfect, 0, true, interval.PerfectGroup, 8, 2},
		{"augmented fourth of multiplicity M", aug(mag, 4), interval.Up, 4, ai(mag), interval.Augmented, mag, true, interval.PerfectGroup, 4, 0},
		{"diminished fifth of multiplicity M", dim(mag, 5), interval.Up, 5, ai(-mag), interval.Diminished, mag, true, interval.PerfectGroup, 5, 0},
		{"third with deviation -M", interval.MustFromDeviation(3, ai(-mag)), interval.Up, 3, ai(-mag), interval.Diminished, mag - 1, true, interval.MajorMinorGroup, 3, 0},
		{"major of MaxSize", basicSize(interval.MaxSize), interval.Up, interval.MaxSize, ai(0), interval.Major, 0, true, interval.MajorMinorGroup, maxSizeSimple, maxOctaves},
		{"descending minor of MaxSize", interval.MustFromDeviation(interval.MaxSize, ai(-1)).MustDown(), interval.Down, interval.MaxSize, ai(-1), interval.Minor, 0, true, interval.MajorMinorGroup, maxSizeSimple, maxOctaves},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			i := tt.i
			assert.True(t, i.IsValid())
			assert.Equal(t, tt.dir, i.Direction())
			assert.Equal(t, tt.dir == interval.Up, i.IsAscending())
			assert.Equal(t, tt.dir == interval.Down, i.IsDescending())
			assert.Equal(t, tt.size, i.Size())
			assert.Equal(t, tt.dev, i.Deviation())
			assert.Equal(t, tt.q, i.Quality())
			assert.Equal(t, tt.mult, i.Multiplicity())
			assert.Equal(t, tt.integer, i.IsIntegerQuality())
			assert.Equal(t, tt.group, i.Group())
			assert.Equal(t, tt.simple, i.SimpleSize())
			assert.Equal(t, tt.octaves, i.Octaves())
		})
	}
}

func TestReadingSimple(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		i       string
		simple  string
		octaves int
	}{
		{"P1", "P1", "P1", 0},
		{"P8", "P8", "P8", 0},
		{"M9", "M9", "M2", 1},
		{"M10", "M10", "M3", 1},
		{"P15", "P15", "P8", 1},
		{"d15", "d15", "d8", 1},
		{"A15", "A15", "A8", 1},
		{"M17", "M17", "M3", 2},
		{"P22", "P22", "P8", 2},
		{"descending M10", "↓M10", "↓M3", 1},
		{"descending A8", "↓A8", "↓A8", 0},
		{"n10", "n10", "n3", 1},
		{"third -1/4 of two octaves", "{-1/4}17", "{-1/4}3", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			i := iv(tt.i)
			got, err := i.Simple()
			require.NoError(t, err)
			assert.True(t, got.Equal(iv(tt.simple)), "Simple(%s) = %v", tt.i, got)
			assert.Equal(t, tt.octaves, i.Octaves())
			assert.True(t, got.MustAddOctaves(i.Octaves()).Equal(i), "P-5: Simple + Octaves gives the interval")
		})
	}

	t.Run("property on the predefined and MaxSize", func(t *testing.T) {
		t.Parallel()

		set := append(predefinedBoth(), basicSize(interval.MaxSize), basicSize(interval.MaxSize).MustDown(),
			aug(2, interval.MaxSize-1))
		for _, i := range set {
			s, err := i.Simple()
			require.NoError(t, err)
			assert.True(t, s.IsSimple(), "%v", i)
			assert.Equal(t, i.SimpleSize(), s.Size(), "%v", i)
			assert.Equal(t, i.Deviation(), s.Deviation(), "%v", i)
			assert.Equal(t, i.Direction(), s.Direction(), "%v", i)
			assert.True(t, s.MustAddOctaves(i.Octaves()).Equal(i), "%v", i)
		}
	})

	t.Run("unset", func(t *testing.T) {
		t.Parallel()

		got, err := interval.Interval{}.Simple()
		assertErr(t, err, interval.ErrUnset)
		assert.Equal(t, interval.Interval{}, got)
	})
}

// TestReadingUnset covers the table of 4.2 for the unset interval (P-23).
func TestReadingUnset(t *testing.T) {
	t.Parallel()

	var u interval.Interval

	assert.False(t, u.IsValid())
	assert.Equal(t, interval.DirectionUnset, u.Direction())
	assert.False(t, u.IsAscending())
	assert.False(t, u.IsDescending())
	assert.Equal(t, 0, u.Size())
	assert.Equal(t, alteration.Natural(), u.Deviation())
	assert.Equal(t, interval.QualityUnset, u.Quality())
	assert.Equal(t, 0, u.Multiplicity())
	assert.False(t, u.IsIntegerQuality())
	assert.Equal(t, interval.GroupUnset, u.Group())
	assert.Equal(t, 0, u.SimpleSize())
	assert.Equal(t, 0, u.Octaves())

	for name, f := range propertyReaders() {
		assert.False(t, f(u), name)
	}

	assert.False(t, u.Equal(u))
	assert.False(t, u.Equal(interval.PerfectUnison()))
	assert.False(t, interval.PerfectUnison().Equal(u))
	assert.False(t, u.SameSize(u))
	assert.False(t, u.SameQuality(u))
	assert.False(t, u.SameDirection(u))
	assert.False(t, u.EqualUndirected(u))
	assert.Equal(t, 0, u.Compare(interval.Interval{}))
	assert.Equal(t, -1, u.Compare(iv("↓P22")))
	assert.Equal(t, 1, iv("↓P22").Compare(u))
	assert.Equal(t, "Interval(unset)", u.String())
	assert.True(t, u == interval.Interval{}, "the unset value is the zero value")
	assert.False(t, u.Equal(interval.PerfectUnison()), "the unset interval is not P1")
}

// propertyReaders are the properties of 4.5.2.
func propertyReaders() map[string]func(interval.Interval) bool {
	return map[string]func(interval.Interval) bool{
		"IsPerfectUnison":      interval.Interval.IsPerfectUnison,
		"IsSimple":             interval.Interval.IsSimple,
		"IsCompound":           interval.Interval.IsCompound,
		"IsBeyondDoubleOctave": interval.Interval.IsBeyondDoubleOctave,
		"IsBasic":              interval.Interval.IsBasic,
		"IsDerived":            interval.Interval.IsDerived,
		"IsDiatonic":           interval.Interval.IsDiatonic,
		"IsChromatic":          interval.Interval.IsChromatic,
		"IsNarrow":             interval.Interval.IsNarrow,
		"IsWide":               interval.Interval.IsWide,
		"IsTritone":            interval.Interval.IsTritone,
	}
}

// ---------------------------------------------------------------------------
// Properties (OP-9, P-7, R-14; architecture 4.5.2; KP-8).

func TestReadingProperties(t *testing.T) {
	t.Parallel()

	type flags struct {
		unison, simple, compound, beyond, basic, derived, diatonic, chromatic, narrow, wide, tritone bool
	}

	const (
		T = true
		F = false
	)

	tests := []struct {
		i string
		f flags
	}{
		{"P1", flags{T, T, F, F, T, F, T, F, T, F, F}},
		{"↓P1", flags{T, T, F, F, T, F, T, F, T, F, F}},
		{"A1", flags{F, T, F, F, F, T, F, T, T, F, F}},
		{"↓A1", flags{F, T, F, F, F, T, F, T, T, F, F}},
		{"AA1", flags{F, T, F, F, F, T, F, T, T, F, F}},
		{"{+1/2}1", flags{F, T, F, F, F, T, F, F, T, F, F}},
		{"m2", flags{F, T, F, F, T, F, T, F, T, F, F}},
		{"d2", flags{F, T, F, F, F, T, F, T, T, F, F}},
		{"A2", flags{F, T, F, F, F, T, F, T, T, F, F}},
		{"n3", flags{F, T, F, F, F, T, F, F, T, F, F}},
		{"↓n3", flags{F, T, F, F, F, T, F, F, T, F, F}},
		{"M3", flags{F, T, F, F, T, F, T, F, T, F, F}},
		{"↓M3", flags{F, T, F, F, T, F, T, F, T, F, F}},
		{"d4", flags{F, T, F, F, F, T, F, T, T, F, F}},
		{"P4", flags{F, T, F, F, T, F, T, F, T, F, F}},
		{"A4", flags{F, T, F, F, T, F, T, F, T, F, T}},
		{"↓A4", flags{F, T, F, F, T, F, T, F, T, F, T}},
		{"AA4", flags{F, T, F, F, F, T, F, T, T, F, F}},
		{"d5", flags{F, T, F, F, T, F, T, F, F, T, T}},
		{"↓d5", flags{F, T, F, F, T, F, T, F, F, T, T}},
		{"dd5", flags{F, T, F, F, F, T, F, T, F, T, F}},
		{"P5", flags{F, T, F, F, T, F, T, F, F, T, F}},
		{"{-1/2}5", flags{F, T, F, F, F, T, F, F, F, T, F}},
		{"m6", flags{F, T, F, F, T, F, T, F, F, T, F}},
		{"M7", flags{F, T, F, F, T, F, T, F, F, T, F}},
		{"d8", flags{F, T, F, F, F, T, F, T, F, T, F}},
		{"P8", flags{F, T, F, F, T, F, T, F, F, T, F}},
		{"↓P8", flags{F, T, F, F, T, F, T, F, F, T, F}},
		{"A8", flags{F, T, F, F, F, T, F, T, F, T, F}},
		{"M9", flags{F, F, T, F, T, F, T, F, F, F, F}},
		{"M10", flags{F, F, T, F, T, F, T, F, F, F, F}},
		{"A11", flags{F, F, T, F, T, F, T, F, F, F, F}},
		{"↓A11", flags{F, F, T, F, T, F, T, F, F, F, F}},
		{"d12", flags{F, F, T, F, T, F, T, F, F, F, F}},
		{"A12", flags{F, F, T, F, F, T, F, T, F, F, F}},
		{"{+1/2}9", flags{F, F, T, F, F, T, F, F, F, F, F}},
		{"P15", flags{F, F, T, F, T, F, T, F, F, F, F}},
		{"A15", flags{F, F, T, F, F, T, F, T, F, F, F}},
		{"M16", flags{F, F, T, T, T, F, T, F, F, F, F}},
		{"A18", flags{F, F, T, T, T, F, T, F, F, F, F}},
		{"AA18", flags{F, F, T, T, F, T, F, T, F, F, F}},
	}

	for _, tt := range tests {
		t.Run(tt.i, func(t *testing.T) {
			t.Parallel()

			i := iv(tt.i)
			assert.Equal(t, tt.f.unison, i.IsPerfectUnison(), "IsPerfectUnison")
			assert.Equal(t, tt.f.simple, i.IsSimple(), "IsSimple")
			assert.Equal(t, tt.f.compound, i.IsCompound(), "IsCompound")
			assert.Equal(t, tt.f.beyond, i.IsBeyondDoubleOctave(), "IsBeyondDoubleOctave")
			assert.Equal(t, tt.f.basic, i.IsBasic(), "IsBasic")
			assert.Equal(t, tt.f.derived, i.IsDerived(), "IsDerived")
			assert.Equal(t, tt.f.diatonic, i.IsDiatonic(), "IsDiatonic")
			assert.Equal(t, tt.f.chromatic, i.IsChromatic(), "IsChromatic")
			assert.Equal(t, tt.f.narrow, i.IsNarrow(), "IsNarrow")
			assert.Equal(t, tt.f.wide, i.IsWide(), "IsWide")
			assert.Equal(t, tt.f.tritone, i.IsTritone(), "IsTritone")
		})
	}

	t.Run("major of MaxSize", func(t *testing.T) {
		t.Parallel()

		i := basicSize(interval.MaxSize)
		assert.True(t, i.IsCompound())
		assert.True(t, i.IsBeyondDoubleOctave())
		assert.True(t, i.IsBasic())
		assert.True(t, i.IsDiatonic())
		assert.False(t, i.IsNarrow())
		assert.False(t, i.IsWide())
		assert.False(t, i.IsTritone())
	})

	// The relations of P-7 on a sweep: basic and derived split the valid
	// intervals; diatonic is basic; chromatic is integer and not diatonic;
	// simple and compound split them; for a compound interval basic,
	// diatonic and chromatic are those of its simple part.
	t.Run("relations", func(t *testing.T) {
		t.Parallel()

		set := predefinedBoth()
		for n := 1; n <= 24; n++ {
			for _, d := range []alteration.Alteration{alt(-1, 2), alt(1, 2), alt(-3, 2), ai(3), ai(-4)} {
				if n == 1 && d.Sign() < 0 {
					continue
				}

				set = append(set, interval.MustFromDeviation(n, d))
			}
		}

		for _, i := range set {
			assert.NotEqual(t, i.IsBasic(), i.IsDerived(), "%v", i)
			assert.Equal(t, i.IsBasic(), i.IsDiatonic(), "%v", i)
			assert.Equal(t, i.IsIntegerQuality() && !i.IsDiatonic(), i.IsChromatic(), "%v", i)
			assert.NotEqual(t, i.IsSimple(), i.IsCompound(), "%v", i)
			assert.Equal(t, i.Size() <= 4, i.IsNarrow(), "%v", i)
			assert.Equal(t, i.Size() >= 5 && i.Size() <= 8, i.IsWide(), "%v", i)
			assert.Equal(t, i.Size() > 15, i.IsBeyondDoubleOctave(), "%v", i)

			s := i.MustSimple()
			assert.Equal(t, s.IsBasic(), i.IsBasic(), "%v", i)
			assert.Equal(t, s.IsChromatic(), i.IsChromatic(), "%v", i)
		}
	})
}

// ---------------------------------------------------------------------------
// Directions (OP-1, OP-2, OP-10, P-4; architecture 4.6).

func TestCreationDirection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		i, up, down, opposite string
	}{
		{"M3", "M3", "↓M3", "↓M3"},
		{"↓M3", "M3", "↓M3", "M3"},
		{"P1", "P1", "P1", "P1"},
		{"A1", "A1", "↓A1", "↓A1"},
		{"↓A1", "A1", "↓A1", "A1"},
		{"{+1/2}1", "{+1/2}1", "↓{+1/2}1", "↓{+1/2}1"},
		{"P8", "P8", "↓P8", "↓P8"},
		{"↓n3", "n3", "↓n3", "n3"},
		{"d2", "d2", "↓d2", "↓d2"},
		{"M17", "M17", "↓M17", "↓M17"},
	}

	for _, tt := range tests {
		t.Run(tt.i, func(t *testing.T) {
			t.Parallel()

			i := iv(tt.i)

			up, err := i.Up()
			require.NoError(t, err)
			assert.True(t, up.Equal(iv(tt.up)), "Up = %v", up)

			down, err := i.Down()
			require.NoError(t, err)
			assert.True(t, down.Equal(iv(tt.down)), "Down = %v", down)

			opp, err := i.Opposite()
			require.NoError(t, err)
			assert.True(t, opp.Equal(iv(tt.opposite)), "Opposite = %v", opp)

			wu, err := i.WithDirection(interval.Up)
			require.NoError(t, err)
			assert.True(t, wu.Equal(up))

			wd, err := i.WithDirection(interval.Down)
			require.NoError(t, err)
			assert.True(t, wd.Equal(down))

			assert.True(t, opp.MustOpposite().Equal(i), "the opposite of the opposite")
			assert.True(t, i.SameSize(opp) && i.EqualUndirected(opp))
		})
	}

	t.Run("unison has no direction", func(t *testing.T) {
		t.Parallel()

		p1 := interval.PerfectUnison()
		for _, i := range []interval.Interval{p1.MustDown(), p1.MustOpposite(), p1.MustWithDirection(interval.Down)} {
			assert.Equal(t, interval.Up, i.Direction())
			assert.True(t, i.Equal(p1))
			assert.True(t, i == p1, "== agrees with Equal for set values")
		}
	})
}

func TestCreationDirectionErrors(t *testing.T) {
	t.Parallel()

	ops := map[string]func(interval.Interval) (interval.Interval, error){
		"Up":       interval.Interval.Up,
		"Down":     interval.Interval.Down,
		"Opposite": interval.Interval.Opposite,
		"WithDirection(Up)": func(i interval.Interval) (interval.Interval, error) {
			return i.WithDirection(interval.Up)
		},
		"WithDirection(unset)": func(i interval.Interval) (interval.Interval, error) {
			return i.WithDirection(interval.DirectionUnset)
		},
	}

	for name, op := range ops {
		got, err := op(interval.Interval{})
		assertErr(t, err, interval.ErrUnset)
		assert.Equal(t, interval.Interval{}, got, name)
	}

	for _, d := range []interval.Direction{interval.DirectionUnset, interval.Direction(3), interval.Direction(9), interval.Direction(255)} {
		for _, i := range ivs("M3", "↓M3", "P1", "A1") {
			got, err := i.WithDirection(d)
			assertErr(t, err, interval.ErrNoDirection)
			assert.Equal(t, interval.Interval{}, got)
		}
	}
}

// ---------------------------------------------------------------------------
// Comparison (OP-18...OP-20, P-3, P-10; architecture 4.7).

// compareSet is a sweep of set intervals for the comparison relations.
func compareSet() []interval.Interval {
	set := predefinedBoth()
	set = append(set, both(ivs("n3", "{+1/2}1", "{+1/2}3", "{+1/2}4", "{+1/2}5", "{-1/4}3", "n6", "n10", "AAAA4", "M16"))...)

	return set
}

func TestCompareEqual(t *testing.T) {
	t.Parallel()

	t.Run("examples", func(t *testing.T) {
		t.Parallel()

		assert.False(t, iv("A4").Equal(iv("d5")), "A4 and d5 differ (P-3)")
		assert.False(t, iv("M3").Equal(iv("↓M3")), "M3 and ↓M3 differ (P-3)")
		assert.True(t, interval.MustNew(interval.Major, 3).Equal(interval.MajorThird()))
		assert.True(t, interval.MustFromDeviation(1, ai(0)).MustDown().Equal(interval.PerfectUnison()))
		assert.False(t, iv("P8").Equal(iv("P1")))
		assert.False(t, iv("A1").Equal(iv("↓A1")))
	})

	t.Run("equivalence and ==", func(t *testing.T) {
		t.Parallel()

		set := compareSet()
		for _, a := range set {
			assert.True(t, a.Equal(a), "reflexive: %v", a)

			for _, b := range set {
				eq := a.Equal(b)
				assert.Equal(t, eq, b.Equal(a), "symmetric: %v %v", a, b)
				assert.Equal(t, eq, a == b, "== agrees with Equal: %v %v", a, b)
				assert.Equal(t, eq, a.Direction() == b.Direction() && a.Size() == b.Size() &&
					a.Deviation() == b.Deviation(), "P-3: %v %v", a, b)
			}
		}
	})

	t.Run("map key", func(t *testing.T) {
		t.Parallel()

		m := map[interval.Interval]int{}
		m[interval.PerfectUnison()]++
		m[interval.MustNew(interval.Perfect, 1)]++
		m[interval.MustFromDeviation(1, ai(0)).MustDown()]++
		m[interval.MajorThird()]++
		m[interval.MajorThird().MustDown().MustUp()]++
		assert.Len(t, m, 2)
		assert.Equal(t, 3, m[interval.PerfectUnison()])
		assert.Equal(t, 2, m[interval.MajorThird()])
	})
}

func TestCompareParts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                                     string
		a, b                                     string
		sameSize, sameQuality, sameDir, undirect bool
	}{
		{"M3 and M3", "M3", "M3", true, true, true, true},
		{"M3 and down M3", "M3", "↓M3", true, true, false, true},
		{"M3 and m3", "M3", "m3", true, false, true, false},
		{"M3 and M6", "M3", "M6", false, true, true, false},
		{"M3 and down M10", "M3", "↓M10", false, true, false, false},
		{"A4 and d5", "A4", "d5", false, false, true, false},
		{"A4 and AA4", "A4", "AA4", true, false, true, false},
		{"d5 and d3", "d5", "d3", false, true, true, false},
		{"dd5 and dd3", "dd5", "dd3", false, true, true, false},
		{"P1 and M3", "P1", "M3", false, false, true, false},
		{"P1 and down M3", "P1", "↓M3", false, false, false, false},
		{"A1 and down A1", "A1", "↓A1", true, true, false, true},
		{"down A1 and down M3", "↓A1", "↓M3", false, false, true, false},
		{"n3 and n6", "n3", "n6", false, true, true, false},
		{"n3 and third -1/2 down", "n3", "↓n3", true, true, false, true},
		{"fractional of the same group and deviation", "{+1/2}4", "{+1/2}5", false, true, true, false},
		{"fractional of different groups", "{+1/2}4", "{+1/2}3", false, false, true, false},
		{"fractional of different deviations", "{+1/4}3", "{+1/2}3", true, false, true, false},
		{"P5 and P8", "P5", "P8", false, true, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			a, b := iv(tt.a), iv(tt.b)
			for _, p := range [][2]interval.Interval{{a, b}, {b, a}} {
				assert.Equal(t, tt.sameSize, p[0].SameSize(p[1]), "SameSize")
				assert.Equal(t, tt.sameQuality, p[0].SameQuality(p[1]), "SameQuality")
				assert.Equal(t, tt.sameDir, p[0].SameDirection(p[1]), "SameDirection")
				assert.Equal(t, tt.undirect, p[0].EqualUndirected(p[1]), "EqualUndirected")
			}
		})
	}

	t.Run("with unset", func(t *testing.T) {
		t.Parallel()

		var u interval.Interval

		for _, i := range ivs("P1", "M3", "↓A1") {
			assert.False(t, i.SameSize(u))
			assert.False(t, i.SameQuality(u))
			assert.False(t, i.SameDirection(u))
			assert.False(t, i.EqualUndirected(u))
			assert.False(t, u.SameSize(i))
			assert.False(t, u.SameQuality(i))
			assert.False(t, u.SameDirection(i))
			assert.False(t, u.EqualUndirected(i))
		}
	})
}

func TestCompareOrder(t *testing.T) {
	t.Parallel()

	t.Run("example of P-10", func(t *testing.T) {
		t.Parallel()

		want := ivs("↓M3", "↓m3", "↓A1", "P1", "A1", "d2", "m2", "M2")
		for k := range want {
			for l := range want {
				assert.Equal(t, cmpInt(k, l), want[k].Compare(want[l]), "%d %d", k, l)
			}
		}

		got := []interval.Interval{want[5], want[2], {}, want[7], want[0], want[3], want[6], want[1], want[4]}
		slices.SortFunc(got, interval.Interval.Compare)
		assert.Equal(t, interval.Interval{}, got[0], "the unset interval first")

		for k, w := range want {
			assert.True(t, got[k+1].Equal(w), "position %d: %v", k+1, got[k+1])
		}
	})

	t.Run("relation", func(t *testing.T) {
		t.Parallel()

		set := compareSet()
		for _, a := range set {
			for _, b := range set {
				c := a.Compare(b)
				assert.Contains(t, []int{-1, 0, 1}, c)
				assert.Equal(t, -c, b.Compare(a), "antisymmetric: %v %v", a, b)
				assert.Equal(t, a.Equal(b), c == 0, "Compare = 0 iff Equal: %v %v", a, b)

				// P-10: by the signed number of letter steps, then by the
				// signed deviation.
				ta, tb := signedSteps(a), signedSteps(b)
				if ta != tb {
					assert.Equal(t, cmpInt(ta, tb), c, "%v %v", a, b)
				} else {
					assert.Equal(t, signedDev(a).Compare(signedDev(b)), c, "%v %v", a, b)
				}
			}
		}
	})

	t.Run("edges", func(t *testing.T) {
		t.Parallel()

		big := basicSize(interval.MaxSize)
		assert.Equal(t, 1, big.Compare(basicSize(interval.MaxSize-1)))
		assert.Equal(t, -1, big.MustDown().Compare(basicSize(interval.MaxSize-1).MustDown()))
		assert.Equal(t, -1, big.MustDown().Compare(big))
		assert.Equal(t, 1, aug(mag, 4).Compare(dim(mag, 4)))
	})
}

func cmpInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

// signedSteps is σ(n − 1) of 4.7, for sizes far below math.MaxInt.
func signedSteps(i interval.Interval) int {
	if i.IsDescending() {
		return -(i.Size() - 1)
	}

	return i.Size() - 1
}

// signedDev is σ·d of 4.7.
func signedDev(i interval.Interval) alteration.Alteration {
	if i.IsDescending() {
		return alteration.Natural().MustSub(i.Deviation())
	}

	return i.Deviation()
}

// ---------------------------------------------------------------------------
// Enumerations (architecture 4.3).

func TestCreationEnumerations(t *testing.T) {
	t.Parallel()

	t.Run("Direction", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, []uint8{0, 1, 2}, []uint8{uint8(interval.DirectionUnset), uint8(interval.Up), uint8(interval.Down)})
		assert.Equal(t, "unset", interval.DirectionUnset.String())
		assert.Equal(t, "up", interval.Up.String())
		assert.Equal(t, "down", interval.Down.String())
		assert.Equal(t, "Direction(3)", interval.Direction(3).String())
		assert.Equal(t, "Direction(255)", interval.Direction(255).String())
	})

	t.Run("Quality", func(t *testing.T) {
		t.Parallel()

		qs := []interval.Quality{
			interval.QualityUnset, interval.Perfect, interval.Major, interval.Minor,
			interval.Augmented, interval.Diminished, interval.Neutral, interval.Fractional,
		}
		names := []string{"unset", "perfect", "major", "minor", "augmented", "diminished", "neutral", "fractional"}

		for k, q := range qs {
			assert.Equal(t, uint8(k), uint8(q))
			assert.Equal(t, names[k], q.String())
		}

		assert.Equal(t, "Quality(8)", interval.Quality(8).String())
		assert.Equal(t, "Quality(200)", interval.Quality(200).String())
	})

	t.Run("Group", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, []uint8{0, 1, 2}, []uint8{uint8(interval.GroupUnset), uint8(interval.PerfectGroup), uint8(interval.MajorMinorGroup)})
		assert.Equal(t, "unset", interval.GroupUnset.String())
		assert.Equal(t, "perfect", interval.PerfectGroup.String())
		assert.Equal(t, "major-minor", interval.MajorMinorGroup.String())
		assert.Equal(t, "Group(3)", interval.Group(3).String())
	})

	t.Run("Replace", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, []uint8{0, 1, 2, 3}, []uint8{
			uint8(interval.ReplaceUnset), uint8(interval.ReplaceFrom),
			uint8(interval.ReplaceTo), uint8(interval.ReplaceBoth),
		})
		assert.Equal(t, "unset", interval.ReplaceUnset.String())
		assert.Equal(t, "from", interval.ReplaceFrom.String())
		assert.Equal(t, "to", interval.ReplaceTo.String())
		assert.Equal(t, "both", interval.ReplaceBoth.String())
		assert.Equal(t, "Replace(4)", interval.Replace(4).String())
	})

	t.Run("MaxSize", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, math.MaxInt, interval.MaxSize)
	})
}
