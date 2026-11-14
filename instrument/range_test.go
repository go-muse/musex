package instrument_test

import (
	"math"
	"slices"
	"testing"

	"github.com/go-muse/musex/instrument"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ОП-1, ОП-2, П-1: a range is a pair of pitches; creation checks only that
// both are set, and the bounds are kept as written.
func TestNewRange(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		lower, upper pitch.Pitch
		wantErr      error
	}{
		{name: "soprano", lower: p("C4"), upper: p("G5")},
		{name: "one sound", lower: p("E4"), upper: p("E4")},
		{name: "written across the octave boundary", lower: p("B♯3"), upper: p("C♭5")},
		{name: "inverted in every system is still created", lower: p("G5"), upper: p("C4")},
		{name: "inadmissible in traditional is still created", lower: p("C4"), upper: cHalf4()},
		{name: "unset lower", lower: pitch.Pitch{}, upper: p("G5"), wantErr: instrument.ErrUnsetPitch},
		{name: "unset upper", lower: p("C4"), upper: pitch.Pitch{}, wantErr: instrument.ErrUnsetPitch},
		{name: "both unset", wantErr: instrument.ErrUnsetPitch},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r, err := instrument.NewRange(tt.lower, tt.upper)
			if tt.wantErr != nil {
				assertCause(t, err, tt.wantErr)
				assert.Equal(t, instrument.Range{}, r)
				assert.False(t, r.IsValid())
				assertPanicsWith(t, tt.wantErr, err, func() { instrument.MustNewRange(tt.lower, tt.upper) })
				return
			}
			require.NoError(t, err)
			assert.True(t, r.IsValid())
			assert.True(t, tt.lower.Equal(r.Lower()), "lower %v, want %v", r.Lower(), tt.lower)
			assert.True(t, tt.upper.Equal(r.Upper()), "upper %v, want %v", r.Upper(), tt.upper)
			assert.True(t, tt.lower.SameSpelling(r.Lower()) && tt.lower.SameOctave(r.Lower()), "lower is not rewritten")
			assert.True(t, tt.upper.SameSpelling(r.Upper()) && tt.upper.SameOctave(r.Upper()), "upper is not rewritten")
			assert.Equal(t, r, instrument.MustNewRange(tt.lower, tt.upper))
		})
	}
}

// П-1а: Range{} is the unset range, "no range".
func TestRange_zeroValue(t *testing.T) {
	t.Parallel()
	var r instrument.Range
	assert.False(t, r.IsValid())
	assert.Equal(t, pitch.Pitch{}, r.Lower())
	assert.Equal(t, pitch.Pitch{}, r.Upper())
	assert.True(t, r.Equal(instrument.Range{}))
	assert.Equal(t, "Range(unset)", r.String())
}

// ОП-3, П-2, КП-6, КП-6а, КП-7, К-1, К-2 and the order of causes А 4.9.
func TestRange_Check(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		sys     *system.System
		r       instrument.Range
		wantErr error
	}{
		{name: "КП-6 traditional", sys: trad(), r: rangeKP6()},
		{name: "КП-6 19-position is inverted", sys: pos19(), r: rangeKP6(), wantErr: instrument.ErrInverted},
		{name: "КП-6а traditional", sys: trad(), r: rangeKP6a()},
		{name: "КП-6а 19-position is inverted", sys: pos19(), r: rangeKP6a(), wantErr: instrument.ErrInverted},
		{name: "КП-7 traditional upper inadmissible", sys: trad(), r: rangeKP7(), wantErr: instrument.ErrInadmissible},
		{name: "КП-7 24-quarter-tone", sys: qt24(), r: rangeKP7()},
		{name: "КП-7 lower inadmissible", sys: trad(), r: instrument.MustNewRange(cHalf4(), p("G5")), wantErr: instrument.ErrInadmissible},
		{name: "one sound", sys: trad(), r: rng("E4", "E4")},
		{name: "bounds on one position", sys: trad(), r: rng("B♯3", "C4")},
		{name: "inverted", sys: trad(), r: rng("G5", "C4"), wantErr: instrument.ErrInverted},
		{name: "inadmissible lower before inverted", sys: trad(), r: instrument.MustNewRange(at(cHalf4().Spelling(), 6), p("C4")), wantErr: instrument.ErrInadmissible},
		{name: "inadmissible upper before inverted", sys: trad(), r: instrument.MustNewRange(p("G5"), cHalf4()), wantErr: instrument.ErrInadmissible},
		{name: "unset range", sys: trad(), r: instrument.Range{}, wantErr: instrument.ErrNoRange},
		{name: "nil system", sys: nil, r: rng("C4", "G5"), wantErr: instrument.ErrUnset},
		{name: "zero system", sys: &system.System{}, r: rng("C4", "G5"), wantErr: instrument.ErrUnset},
		{name: "nil system before unset range", sys: nil, r: instrument.Range{}, wantErr: instrument.ErrUnset},
		{name: "unset range in 19-position", sys: pos19(), r: instrument.Range{}, wantErr: instrument.ErrNoRange},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.r.Check(tt.sys)
			if tt.wantErr == nil {
				require.NoError(t, err)
				got := tt.r.MustCheck(tt.sys)
				assert.Equal(t, tt.r, got, "MustCheck returns the receiver")
				return
			}
			assertCause(t, err, tt.wantErr)
			assertPanicsWith(t, tt.wantErr, err, func() { tt.r.MustCheck(tt.sys) })
		})
	}
}

// rangeOps lists every operation of a range in a system (ОП-4…ОП-12, ОП-14,
// ОП-16) with valid other arguments, so that only the receiver decides.
func rangeOps() map[string]func(r instrument.Range, sys *system.System) error {
	other := rng("C4", "E4")
	x := p("D4")
	return map[string]func(r instrument.Range, sys *system.System) error{
		"Contains": func(r instrument.Range, sys *system.System) error {
			_, err := r.Contains(sys, x)
			return err
		},
		"Locate": func(r instrument.Range, sys *system.System) error {
			_, err := r.Locate(sys, x)
			return err
		},
		"Width": func(r instrument.Range, sys *system.System) error {
			_, err := r.Width(sys)
			return err
		},
		"Count": func(r instrument.Range, sys *system.System) error {
			_, err := r.Count(sys)
			return err
		},
		"Folds": func(r instrument.Range, sys *system.System) error {
			_, err := r.Folds(sys, x)
			return err
		},
		"Fold": func(r instrument.Range, sys *system.System) error {
			_, _, err := r.Fold(sys, x)
			return err
		},
		"Intersect": func(r instrument.Range, sys *system.System) error {
			_, _, err := r.Intersect(sys, other)
			return err
		},
		"Intersect other": func(r instrument.Range, sys *system.System) error {
			_, _, err := other.Intersect(sys, r)
			return err
		},
		"Includes": func(r instrument.Range, sys *system.System) error {
			_, err := r.Includes(sys, other)
			return err
		},
		"Includes other": func(r instrument.Range, sys *system.System) error {
			_, err := other.Includes(sys, r)
			return err
		},
		"Hull": func(r instrument.Range, sys *system.System) error {
			_, err := r.Hull(sys, other)
			return err
		},
		"Hull other": func(r instrument.Range, sys *system.System) error {
			_, err := other.Hull(sys, r)
			return err
		},
		"Pitches": func(r instrument.Range, sys *system.System) error {
			_, err := r.Pitches(sys, system.Sharps())
			return err
		},
		"Equivalent": func(r instrument.Range, sys *system.System) error {
			_, err := r.Equivalent(sys, other)
			return err
		},
		"Equivalent other": func(r instrument.Range, sys *system.System) error {
			_, err := other.Equivalent(sys, r)
			return err
		},
	}
}

// К-1, К-2: every operation with positions gives the cause of ОП-3 for an
// invalid range, Е-7 for the unset range and Е-1 for an unset system.
func TestRange_operationsCheckTheRangeFirst(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		sys     *system.System
		r       instrument.Range
		wantErr error
	}{
		{name: "КП-6 in 19-position", sys: pos19(), r: rangeKP6(), wantErr: instrument.ErrInverted},
		{name: "КП-6а in 19-position", sys: pos19(), r: rangeKP6a(), wantErr: instrument.ErrInverted},
		{name: "КП-7 in traditional", sys: trad(), r: rangeKP7(), wantErr: instrument.ErrInadmissible},
		{name: "unset range", sys: trad(), r: instrument.Range{}, wantErr: instrument.ErrNoRange},
		{name: "nil system", sys: nil, r: rng("C4", "G5"), wantErr: instrument.ErrUnset},
	}
	for opName, op := range rangeOps() {
		for _, c := range cases {
			t.Run(opName+"/"+c.name, func(t *testing.T) {
				t.Parallel()
				if c.sys != nil && c.r.IsValid() {
					assertCause(t, c.r.Check(c.sys), c.wantErr)
				}
				assertCause(t, op(c.r, c.sys), c.wantErr)
			})
		}
	}
}

// П-18 step 4: for two ranges, the receiver is checked whole before other.
func TestRange_twoRangesOrder(t *testing.T) {
	t.Parallel()
	ops := map[string]func(a, b instrument.Range, sys *system.System) error{
		"Intersect": func(a, b instrument.Range, sys *system.System) error {
			_, _, err := a.Intersect(sys, b)
			return err
		},
		"Includes": func(a, b instrument.Range, sys *system.System) error {
			_, err := a.Includes(sys, b)
			return err
		},
		"Hull": func(a, b instrument.Range, sys *system.System) error {
			_, err := a.Hull(sys, b)
			return err
		},
		"Equivalent": func(a, b instrument.Range, sys *system.System) error {
			_, err := a.Equivalent(sys, b)
			return err
		},
	}
	cases := []struct {
		name    string
		a, b    instrument.Range
		wantErr error
	}{
		{name: "inverted receiver before unset other", a: rangeKP6(), b: instrument.Range{}, wantErr: instrument.ErrInverted},
		{name: "unset receiver before inverted other", a: instrument.Range{}, b: rangeKP6(), wantErr: instrument.ErrNoRange},
		{name: "valid receiver, inverted other", a: rng("C4", "G5"), b: rangeKP6a(), wantErr: instrument.ErrInverted},
		{name: "valid receiver, unset other", a: rng("C4", "G5"), b: instrument.Range{}, wantErr: instrument.ErrNoRange},
	}
	for opName, op := range ops {
		for _, c := range cases {
			t.Run(opName+"/"+c.name, func(t *testing.T) {
				t.Parallel()
				assertCause(t, op(c.a, c.b, pos19()), c.wantErr)
			})
		}
	}
}

// ОП-4, П-3, КП-1, КП-2, КП-13, Г-2, Г-3, К-1, and the causes of the argument.
func TestRange_Contains(t *testing.T) {
	t.Parallel()
	soprano := rng("C4", "G5")
	tests := []struct {
		name    string
		sys     *system.System
		r       instrument.Range
		x       pitch.Pitch
		want    bool
		wantErr error
	}{
		{name: "КП-1 B♯3 at the position of C4", sys: trad(), r: soprano, x: p("B♯3"), want: true},
		{name: "КП-1 B3 below", sys: trad(), r: soprano, x: p("B3"), want: false},
		{name: "КП-1 C♭4 at the position of B3", sys: trad(), r: soprano, x: p("C♭4"), want: false},
		{name: "КП-1 upper bound G5", sys: trad(), r: soprano, x: p("G5"), want: true},
		{name: "КП-1 A5 above", sys: trad(), r: soprano, x: p("A5"), want: false},
		{name: "lower bound C4", sys: trad(), r: soprano, x: p("C4"), want: true},
		{name: "Г-2 G♯5 next to the upper bound", sys: trad(), r: soprano, x: p("G♯5"), want: false},
		{name: "КП-2 B♯ with Octave 3", sys: trad(), r: soprano, x: at(spelling.BSharp(), 3), want: false},
		{name: "КП-13 A𝄫 with Octave 5 in traditional", sys: trad(), r: soprano, x: at(spelling.ADoubleFlat(), 5), want: true},
		{name: "КП-13 A𝄫 with Octave 5 in 19-position", sys: pos19(), r: soprano, x: at(spelling.ADoubleFlat(), 5), want: false},
		{name: "КП-6 in traditional", sys: trad(), r: rangeKP6(), x: p("C♯4"), want: true},
		{name: "КП-7 in 24-quarter-tone", sys: qt24(), r: rangeKP7(), x: cHalf4(), want: true},
		{name: "unset x", sys: trad(), r: soprano, x: pitch.Pitch{}, wantErr: instrument.ErrUnsetPitch},
		{name: "inadmissible x", sys: trad(), r: soprano, x: cHalf4(), wantErr: instrument.ErrInadmissible},
		{name: "inverted range before unset x", sys: pos19(), r: rangeKP6(), x: pitch.Pitch{}, wantErr: instrument.ErrInverted},
		{name: "inadmissible bound before inadmissible x", sys: trad(), r: rangeKP7(), x: cHalf4(), wantErr: instrument.ErrInadmissible},
		{name: "unset range before unset x", sys: trad(), r: instrument.Range{}, x: pitch.Pitch{}, wantErr: instrument.ErrNoRange},
		{name: "nil system before unset range", sys: nil, r: instrument.Range{}, x: pitch.Pitch{}, wantErr: instrument.ErrUnset},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.r.Contains(tt.sys, tt.x)
			if tt.wantErr != nil {
				assertCause(t, err, tt.wantErr)
				assert.False(t, got)
				assertPanicsWith(t, tt.wantErr, err, func() { tt.r.MustContains(tt.sys, tt.x) })
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want, tt.r.MustContains(tt.sys, tt.x))
		})
	}
}

// ОП-5, КП-8.
func TestRange_Locate(t *testing.T) {
	t.Parallel()
	soprano := rng("C4", "G5")
	tests := []struct {
		name    string
		sys     *system.System
		r       instrument.Range
		x       pitch.Pitch
		want    instrument.Location
		wantErr error
	}{
		{name: "КП-8 B3", sys: trad(), r: soprano, x: p("B3"), want: instrument.Below},
		{name: "КП-8 C4", sys: trad(), r: soprano, x: p("C4"), want: instrument.Inside},
		{name: "КП-8 A5", sys: trad(), r: soprano, x: p("A5"), want: instrument.Above},
		{name: "B♯3 inside", sys: trad(), r: soprano, x: p("B♯3"), want: instrument.Inside},
		{name: "C♭4 below", sys: trad(), r: soprano, x: p("C♭4"), want: instrument.Below},
		{name: "G5 inside", sys: trad(), r: soprano, x: p("G5"), want: instrument.Inside},
		{name: "КП-13 above in 19-position", sys: pos19(), r: soprano, x: at(spelling.ADoubleFlat(), 5), want: instrument.Above},
		{name: "unset x", sys: trad(), r: soprano, x: pitch.Pitch{}, wantErr: instrument.ErrUnsetPitch},
		{name: "inadmissible x", sys: trad(), r: soprano, x: cHalf4(), wantErr: instrument.ErrInadmissible},
		{name: "unset range", sys: trad(), r: instrument.Range{}, x: p("C4"), wantErr: instrument.ErrNoRange},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.r.Locate(tt.sys, tt.x)
			if tt.wantErr != nil {
				assertCause(t, err, tt.wantErr)
				assert.Equal(t, instrument.LocationUnset, got)
				assertPanicsWith(t, tt.wantErr, err, func() { tt.r.MustLocate(tt.sys, tt.x) })
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want, tt.r.MustLocate(tt.sys, tt.x))
			contains, err := tt.r.Contains(tt.sys, tt.x)
			require.NoError(t, err)
			assert.Equal(t, got == instrument.Inside, contains, "Locate agrees with Contains")
		})
	}
}

// ОП-6, ОП-7, КП-3, Г-1, К-2 (unrepresentable width), КР-19 (unrepresentable
// count).
func TestRange_WidthAndCount(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		r         instrument.Range
		wantWidth system.Step
		wantCount int
	}{
		{name: "КП-3 piano 88 keys", r: rng("A0", "C8"), wantWidth: 87, wantCount: 88},
		{name: "КП-3 piano 92 keys", r: rng("F0", "C8"), wantWidth: 91, wantCount: 92},
		{name: "КП-3 piano 85 keys", r: rng("A0", "A7"), wantWidth: 84, wantCount: 85},
		{name: "КП-3 soprano", r: rng("C4", "G5"), wantWidth: 19, wantCount: 20},
		{name: "КП-3 Г-1 one sound", r: rng("E4", "E4"), wantWidth: 0, wantCount: 1},
		{name: "bounds on one position", r: rng("B♯3", "C4"), wantWidth: 0, wantCount: 1},
		{name: "КП-6 in traditional", r: rangeKP6(), wantWidth: 1, wantCount: 2},
		{name: "КП-6а in traditional", r: rangeKP6a(), wantWidth: 0, wantCount: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w, err := tt.r.Width(trad())
			require.NoError(t, err)
			assert.Equal(t, tt.wantWidth, w)
			c, err := tt.r.Count(trad())
			require.NoError(t, err)
			assert.Equal(t, tt.wantCount, c)
			assert.Equal(t, tt.wantWidth, tt.r.MustWidth(trad()))
			assert.Equal(t, tt.wantCount, tt.r.MustCount(trad()))
		})
	}

	t.Run("КП-7 in 24-quarter-tone", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, system.Step(1), rangeKP7().MustWidth(qt24()))
		assert.Equal(t, 2, rangeKP7().MustCount(qt24()))
	})

	t.Run("К-2 unrepresentable width", func(t *testing.T) {
		t.Parallel()
		huge := instrument.MustNewRange(at(spelling.C(), math.MinInt), at(spelling.C(), math.MaxInt))
		require.NoError(t, huge.Check(trad()), "the range is valid")
		in, err := huge.Contains(trad(), p("C4"))
		require.NoError(t, err)
		assert.True(t, in)

		w, err := huge.Width(trad())
		assertDependencyCause(t, err, system.ErrOutOfRange)
		assert.Equal(t, system.Step(0), w)
		assertPanicsWith(t, system.ErrOutOfRange, err, func() { huge.MustWidth(trad()) })

		c, err := huge.Count(trad())
		assertDependencyCause(t, err, system.ErrOutOfRange)
		assert.Equal(t, 0, c)
		assertPanicsWith(t, system.ErrOutOfRange, err, func() { huge.MustCount(trad()) })
	})

	t.Run("КР-19 width math.MaxInt has an unrepresentable count", func(t *testing.T) {
		t.Parallel()
		// 12·768614336404564650 + 7 = math.MaxInt steps from C to G.
		wide := instrument.MustNewRange(at(spelling.C(), 0), at(spelling.G(), 768614336404564650))
		w, err := wide.Width(trad())
		require.NoError(t, err)
		assert.Equal(t, system.Step(math.MaxInt), w)

		c, err := wide.Count(trad())
		assertDependencyCause(t, err, system.ErrOutOfRange)
		assert.Equal(t, 0, c)
	})
}

// ОП-8, ОП-9, П-6, КП-4, КП-5, К-3, Г-4.
func TestRange_FoldsAndFold(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		r           instrument.Range
		x           pitch.Pitch
		wantAll     []pitch.Pitch
		wantNearest pitch.Pitch
		wantFound   bool
	}{
		{name: "КП-4 bass E2", r: rng("F2", "D4"), x: p("E2"), wantAll: []pitch.Pitch{p("E3")}, wantNearest: p("E3"), wantFound: true},
		{name: "КП-4 soprano B5", r: rng("C4", "G5"), x: p("B5"), wantAll: []pitch.Pitch{p("B4")}, wantNearest: p("B4"), wantFound: true},
		{name: "КП-4 E1 into C3–G5", r: rng("C3", "G5"), x: p("E1"), wantAll: []pitch.Pitch{p("E3"), p("E4"), p("E5")}, wantNearest: p("E3"), wantFound: true},
		{name: "КП-4 Г-4 G4 into C4–E4", r: rng("C4", "E4"), x: p("G4"), wantAll: nil, wantFound: false},
		{name: "КП-5 E4 inside", r: rng("C4", "G5"), x: p("E4"), wantAll: []pitch.Pitch{p("E4"), p("E5")}, wantNearest: p("E4"), wantFound: true},
		{
			name: "КП-5 B♯ with Octave 3", r: rng("C4", "G5"), x: at(spelling.BSharp(), 3),
			wantAll:     []pitch.Pitch{at(spelling.BSharp(), 4), at(spelling.BSharp(), 5)},
			wantNearest: at(spelling.BSharp(), 4), wantFound: true,
		},
		{name: "x above, nearest is the last", r: rng("C3", "G5"), x: p("E8"), wantAll: []pitch.Pitch{p("E3"), p("E4"), p("E5")}, wantNearest: p("E5"), wantFound: true},
		{name: "x on the upper bound", r: rng("C4", "G5"), x: p("G5"), wantAll: []pitch.Pitch{p("G4"), p("G5")}, wantNearest: p("G5"), wantFound: true},
		{name: "one sound range, same name", r: rng("E4", "E4"), x: p("E1"), wantAll: []pitch.Pitch{p("E4")}, wantNearest: p("E4"), wantFound: true},
		{name: "one sound range, other name at the position", r: rng("B♯3", "B♯3"), x: p("C2"), wantAll: []pitch.Pitch{p("C4")}, wantNearest: p("C4"), wantFound: true},
		{
			name: "КР-15 highest Octave, no overflow", r: instrument.MustNewRange(at(spelling.D(), math.MaxInt), at(spelling.E(), math.MaxInt)),
			x: at(spelling.C(), 0), wantAll: nil, wantFound: false,
		},
		{
			name: "КР-15 lowest Octave, no overflow", r: instrument.MustNewRange(at(spelling.C(), math.MinInt), at(spelling.D(), math.MinInt)),
			x: at(spelling.E(), 0), wantAll: nil, wantFound: false,
		},
		{
			name: "КР-15 highest Octave, found", r: instrument.MustNewRange(at(spelling.C(), math.MaxInt), at(spelling.G(), math.MaxInt)),
			x: at(spelling.E(), math.MinInt), wantAll: []pitch.Pitch{at(spelling.E(), math.MaxInt)}, wantNearest: at(spelling.E(), math.MaxInt), wantFound: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sys := trad()
			seq, err := tt.r.Folds(sys, tt.x)
			require.NoError(t, err)
			got := slices.Collect(seq)
			assertPitchesEqual(t, tt.wantAll, got)
			assertPitchesEqual(t, got, slices.Collect(seq)) // iterated again
			assertPitchesEqual(t, tt.wantAll, slices.Collect(tt.r.MustFolds(sys, tt.x)))
			for i, f := range got {
				assert.True(t, f.SameSpelling(tt.x), "fold %d keeps the name of x", i)
				in, err := tt.r.Contains(sys, f)
				require.NoError(t, err)
				assert.True(t, in, "fold %d is inside", i)
				if i > 0 {
					assert.Equal(t, -1, sys.MustCompare(got[i-1], f), "folds go up")
				}
			}

			nearest, found, err := tt.r.Fold(sys, tt.x)
			require.NoError(t, err)
			assert.Equal(t, tt.wantFound, found)
			if !tt.wantFound {
				assert.Equal(t, pitch.Pitch{}, nearest)
				assert.Empty(t, got)
			} else {
				assert.True(t, tt.wantNearest.SameSpelling(nearest) && tt.wantNearest.SameOctave(nearest), "nearest %v, want %v", nearest, tt.wantNearest)
				assert.True(t, slices.ContainsFunc(got, nearest.Equal), "the nearest is one of the folds")
			}
			mn, mf := tt.r.MustFold(sys, tt.x)
			assert.Equal(t, found, mf)
			assert.True(t, nearest.Equal(mn) || !found)
		})
	}

	t.Run("x inside is the nearest itself", func(t *testing.T) {
		t.Parallel()
		r := rng("C4", "G5")
		for _, x := range []pitch.Pitch{p("B♯3"), p("C4"), p("E4"), p("G5"), p("F♭5")} {
			got, found, err := r.Fold(trad(), x)
			require.NoError(t, err)
			require.True(t, found)
			assert.True(t, x.SameSpelling(got) && x.SameOctave(got), "%v, want %v", got, x)
		}
	})

	t.Run("the sequence can be stopped", func(t *testing.T) {
		t.Parallel()
		seq, err := rng("C1", "C8").Folds(trad(), p("E4"))
		require.NoError(t, err)
		var first []pitch.Pitch
		for f := range seq {
			first = append(first, f)
			if len(first) == 2 {
				break
			}
		}
		assertPitchesEqual(t, []pitch.Pitch{p("E1"), p("E2")}, first)
	})

	t.Run("causes are reported before the sequence", func(t *testing.T) {
		t.Parallel()
		seq, err := rng("C4", "G5").Folds(trad(), pitch.Pitch{})
		assertCause(t, err, instrument.ErrUnsetPitch)
		assert.Nil(t, seq)
		_, err = rng("C4", "G5").Folds(trad(), cHalf4())
		assertCause(t, err, instrument.ErrInadmissible)
		assertPanicsWith(t, instrument.ErrInadmissible, err, func() { rng("C4", "G5").MustFolds(trad(), cHalf4()) })

		nearest, found, err := rng("C4", "G5").Fold(trad(), pitch.Pitch{})
		assertCause(t, err, instrument.ErrUnsetPitch)
		assert.False(t, found)
		assert.Equal(t, pitch.Pitch{}, nearest)
		assertPanicsWith(t, instrument.ErrUnsetPitch, err, func() { rng("C4", "G5").MustFold(trad(), pitch.Pitch{}) })
	})
}

// assertPitchesEqual compares two lists of pitches by record: the same
// spellings and Octaves in the same order.
func assertPitchesEqual(t *testing.T, want, got []pitch.Pitch) {
	t.Helper()
	if !assert.Len(t, got, len(want), "want %v, got %v", want, got) {
		return
	}
	for i := range want {
		assert.True(t, want[i].SameSpelling(got[i]) && want[i].SameOctave(got[i]), "pitch %d: want %v, got %v", i, want[i], got[i])
	}
}

// assertRange compares a range by record with the expected bounds.
func assertRange(t *testing.T, want, got instrument.Range) {
	t.Helper()
	assert.True(t, want.Equal(got), "want %v, got %v", want, got)
}

// ОП-10, ОП-11, ОП-12, П-19, КП-9, КП-10, КП-10а, КП-11, Г-5.
func TestRange_IntersectIncludesHull(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		a, b          instrument.Range
		wantIntersect instrument.Range
		wantNonEmpty  bool
		wantIncludes  bool
		wantHull      instrument.Range
	}{
		{
			name: "КП-9 alto and tenor", a: rng("G3", "D5"), b: rng("C3", "G4"),
			wantIntersect: rng("G3", "G4"), wantNonEmpty: true, wantIncludes: false, wantHull: rng("C3", "D5"),
		},
		{
			name: "КП-9 Г-5 no common sound", a: rng("C4", "E4"), b: rng("G4", "A4"),
			wantNonEmpty: false, wantIncludes: false, wantHull: rng("C4", "A4"),
		},
		{
			name: "КП-9 Г-5 one common sound", a: rng("C3", "G4"), b: rng("G4", "D5"),
			wantIntersect: rng("G4", "G4"), wantNonEmpty: true, wantIncludes: false, wantHull: rng("C3", "D5"),
		},
		{
			name: "КП-10 bounds of the receiver at equal positions", a: rng("B♯3", "G5"), b: rng("C4", "A5"),
			wantIntersect: rng("B♯3", "G5"), wantNonEmpty: true, wantIncludes: false, wantHull: rng("B♯3", "A5"),
		},
		{
			name: "КП-10 reverse order", a: rng("C4", "A5"), b: rng("B♯3", "G5"),
			wantIntersect: rng("C4", "G5"), wantNonEmpty: true, wantIncludes: true, wantHull: rng("C4", "A5"),
		},
		{
			name: "КП-10а lower from the lowers, upper from the uppers", a: rng("C3", "A♭4"), b: rng("G♯4", "D5"),
			wantIntersect: rng("G♯4", "A♭4"), wantNonEmpty: true, wantIncludes: false, wantHull: rng("C3", "D5"),
		},
		{
			name: "КП-10а reverse order", a: rng("G♯4", "D5"), b: rng("C3", "A♭4"),
			wantIntersect: rng("G♯4", "A♭4"), wantNonEmpty: true, wantIncludes: false, wantHull: rng("C3", "D5"),
		},
		{
			name: "КП-11 includes", a: rng("G3", "D5"), b: rng("G3", "G4"),
			wantIntersect: rng("G3", "G4"), wantNonEmpty: true, wantIncludes: true, wantHull: rng("G3", "D5"),
		},
		{
			name: "КП-11 the range and itself", a: rng("G3", "D5"), b: rng("G3", "D5"),
			wantIntersect: rng("G3", "D5"), wantNonEmpty: true, wantIncludes: true, wantHull: rng("G3", "D5"),
		},
		{
			name: "includes a range written by other names", a: rng("C4", "G5"), b: rng("B♯3", "F𝄪5"),
			wantIntersect: rng("C4", "G5"), wantNonEmpty: true, wantIncludes: true, wantHull: rng("C4", "G5"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sys := trad()
			got, ok, err := tt.a.Intersect(sys, tt.b)
			require.NoError(t, err)
			assert.Equal(t, tt.wantNonEmpty, ok)
			if tt.wantNonEmpty {
				assertRange(t, tt.wantIntersect, got)
				assert.True(t, tt.a.MustIncludes(sys, got), "the intersection is in a")
				assert.True(t, tt.b.MustIncludes(sys, got), "the intersection is in b")
				assert.NoError(t, got.Check(sys))
			} else {
				assert.Equal(t, instrument.Range{}, got)
			}
			mg, mok := tt.a.MustIntersect(sys, tt.b)
			assert.Equal(t, ok, mok)
			assert.True(t, got.Equal(mg))

			inc, err := tt.a.Includes(sys, tt.b)
			require.NoError(t, err)
			assert.Equal(t, tt.wantIncludes, inc)
			assert.Equal(t, tt.wantIncludes, tt.a.MustIncludes(sys, tt.b))
			if tt.wantIncludes {
				assert.True(t, ok && got.MustEquivalent(sys, tt.b), "the intersection with an included range is that range")
			}

			hull, err := tt.a.Hull(sys, tt.b)
			require.NoError(t, err)
			assertRange(t, tt.wantHull, hull)
			assertRange(t, hull, tt.a.MustHull(sys, tt.b))
			assert.True(t, hull.MustIncludes(sys, tt.a), "the hull includes a")
			assert.True(t, hull.MustIncludes(sys, tt.b), "the hull includes b")
		})
	}
}

// ОП-13, П-19, КП-12, К-4 and the order П-18 step 5.
func TestRangeOf(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		sys     *system.System
		ps      []pitch.Pitch
		want    instrument.Range
		wantErr error
	}{
		{name: "КП-12 melody", sys: trad(), ps: []pitch.Pitch{p("C4"), p("E4"), p("G4"), p("A4")}, want: rng("C4", "A4")},
		{name: "КП-12 another order", sys: trad(), ps: []pitch.Pitch{p("A4"), p("C4"), p("G4"), p("E4")}, want: rng("C4", "A4")},
		{name: "КП-12 B♯3 first", sys: trad(), ps: []pitch.Pitch{p("B♯3"), p("C4"), p("E4")}, want: rng("B♯3", "E4")},
		{name: "КП-12 C4 first", sys: trad(), ps: []pitch.Pitch{p("C4"), p("B♯3"), p("E4")}, want: rng("C4", "E4")},
		{name: "КП-12 one pitch", sys: trad(), ps: []pitch.Pitch{p("E4")}, want: rng("E4", "E4")},
		{name: "upper at equal positions: first", sys: trad(), ps: []pitch.Pitch{p("C4"), p("E4"), p("F♭4")}, want: rng("C4", "E4")},
		{name: "upper at equal positions: F♭4 first", sys: trad(), ps: []pitch.Pitch{p("C4"), p("F♭4"), p("E4")}, want: rng("C4", "F♭4")},
		{name: "КП-12 empty", sys: trad(), ps: nil, wantErr: instrument.ErrNoPitches},
		{name: "unset pitch", sys: trad(), ps: []pitch.Pitch{p("C4"), {}}, wantErr: instrument.ErrUnsetPitch},
		{name: "К-4 C+1/2 in traditional", sys: trad(), ps: []pitch.Pitch{p("C4"), cHalf4()}, wantErr: instrument.ErrInadmissible},
		{name: "К-4 unset before inadmissible", sys: trad(), ps: []pitch.Pitch{{}, cHalf4()}, wantErr: instrument.ErrUnsetPitch},
		{name: "К-4 inadmissible before unset", sys: trad(), ps: []pitch.Pitch{cHalf4(), {}}, wantErr: instrument.ErrInadmissible},
		{name: "C+1/2 in 24-quarter-tone", sys: qt24(), ps: []pitch.Pitch{cHalf4(), p("C4")}, want: instrument.MustNewRange(p("C4"), cHalf4())},
		{name: "nil system", sys: nil, ps: []pitch.Pitch{p("C4")}, wantErr: instrument.ErrUnset},
		{name: "nil system before empty", sys: nil, ps: nil, wantErr: instrument.ErrUnset},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := instrument.RangeOf(tt.sys, tt.ps...)
			if tt.wantErr != nil {
				assertCause(t, err, tt.wantErr)
				assert.Equal(t, instrument.Range{}, got)
				assertPanicsWith(t, tt.wantErr, err, func() { instrument.MustRangeOf(tt.sys, tt.ps...) })
				return
			}
			require.NoError(t, err)
			assertRange(t, tt.want, got)
			assertRange(t, tt.want, instrument.MustRangeOf(tt.sys, tt.ps...))
			for _, x := range tt.ps {
				in, err := got.Contains(tt.sys, x)
				require.NoError(t, err)
				assert.True(t, in, "%v is in the range of the set", x)
			}
		})
	}

	t.Run("КП-12 width", func(t *testing.T) {
		t.Parallel()
		r := instrument.MustRangeOf(trad(), p("C4"), p("E4"), p("G4"), p("A4"))
		assert.Equal(t, system.Step(9), r.MustWidth(trad()))
	})
}

// ОП-14, КП-14, К-5.
func TestRange_Pitches(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		sys  *system.System
		r    instrument.Range
		rule system.SpellingRule
		want []pitch.Pitch
	}{
		{name: "КП-14 sharps", sys: trad(), r: rng("C4", "E4"), rule: system.Sharps(), want: []pitch.Pitch{p("C4"), p("C♯4"), p("D4"), p("D♯4"), p("E4")}},
		{name: "КП-14 flats", sys: trad(), r: rng("C4", "E4"), rule: system.Flats(), want: []pitch.Pitch{p("C4"), p("D♭4"), p("D4"), p("E♭4"), p("E4")}},
		{name: "one sound", sys: trad(), r: rng("B♯3", "C4"), rule: system.Sharps(), want: []pitch.Pitch{p("C4")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			seq, err := tt.r.Pitches(tt.sys, tt.rule)
			require.NoError(t, err)
			got := slices.Collect(seq)
			assertPitchesEqual(t, tt.want, got)
			assertPitchesEqual(t, got, slices.Collect(seq))
			assertPitchesEqual(t, tt.want, slices.Collect(tt.r.MustPitches(tt.sys, tt.rule)))
		})
	}

	t.Run("К-5 properties", func(t *testing.T) {
		t.Parallel()
		ranges := []struct {
			sys *system.System
			r   instrument.Range
		}{
			{trad(), rng("A0", "C8")},
			{trad(), rng("B♯3", "C♭5")},
			{trad(), rangeKP6()},
			{pos19(), rng("C4", "G5")},
			{qt24(), rangeKP7()},
		}
		for _, c := range ranges {
			got := slices.Collect(c.r.MustPitches(c.sys, system.Sharps()))
			require.Len(t, got, c.r.MustCount(c.sys), "count of %v", c.r)
			assert.True(t, c.sys.MustEquivalent(got[0], c.r.Lower()), "the first is at the lower bound of %v", c.r)
			assert.True(t, c.sys.MustEquivalent(got[len(got)-1], c.r.Upper()), "the last is at the upper bound of %v", c.r)
			for i := 1; i < len(got); i++ {
				assert.Equal(t, system.Step(1), c.sys.MustDistance(got[i-1], got[i]), "step %d of %v", i, c.r)
			}
		}
	})

	t.Run("the sequence can be stopped", func(t *testing.T) {
		t.Parallel()
		huge := instrument.MustNewRange(at(spelling.C(), math.MinInt), at(spelling.C(), math.MaxInt))
		seq, err := huge.Pitches(trad(), system.Sharps())
		require.NoError(t, err)
		n := 0
		for range seq {
			n++
			if n == 3 {
				break
			}
		}
		assert.Equal(t, 3, n)
	})

	t.Run("К-5 causes of system", func(t *testing.T) {
		t.Parallel()
		r := rng("C4", "E4")
		seq, err := r.Pitches(trad(), system.SpellingRule{})
		assertDependencyCause(t, err, system.ErrInvalid)
		assert.Nil(t, seq)
		assertPanicsWith(t, system.ErrInvalid, err, func() { r.MustPitches(trad(), system.SpellingRule{}) })

		chromatic := system.MustChromaticScale(spelling.C(), system.Major, system.Ascending)
		_, err = rng("C4", "G5").Pitches(pos19(), chromatic)
		assertDependencyCause(t, err, system.ErrNotApplicable)

		halfTonic, err := system.ChromaticScale(cHalf4().Spelling(), system.Major, system.Ascending)
		require.NoError(t, err)
		_, err = r.Pitches(trad(), halfTonic)
		assertDependencyCause(t, err, system.ErrInadmissible)
		assert.NotErrorIs(t, err, instrument.ErrInadmissible, "a cause of system is not Е-4")
	})

	t.Run("the range is checked before the rule", func(t *testing.T) {
		t.Parallel()
		_, err := instrument.Range{}.Pitches(trad(), system.SpellingRule{})
		assertCause(t, err, instrument.ErrNoRange)
		_, err = rangeKP6().Pitches(pos19(), system.SpellingRule{})
		assertCause(t, err, instrument.ErrInverted)
		_, err = rng("C4", "E4").Pitches(nil, system.SpellingRule{})
		assertCause(t, err, instrument.ErrUnset)
	})
}

// ОП-15, ОП-16, П-20, КП-15, А 4.2 (== agrees with Equal; usable as a map key).
func TestRange_EqualAndEquivalent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		a, b      instrument.Range
		wantEqual bool
	}{
		{name: "КП-15 by record", a: rng("B♯3", "G5"), b: rng("C4", "G5"), wantEqual: false},
		{name: "same record", a: rng("C4", "G5"), b: rng("C4", "G5"), wantEqual: true},
		{name: "same record, other construction", a: rng("C4", "G5"), b: instrument.MustRangeOf(trad(), p("G5"), p("C4")), wantEqual: true},
		{name: "two unset", a: instrument.Range{}, b: instrument.Range{}, wantEqual: true},
		{name: "unset and set", a: instrument.Range{}, b: rng("C4", "G5"), wantEqual: false},
		{name: "set and unset", a: rng("C4", "G5"), b: instrument.Range{}, wantEqual: false},
		{name: "swapped bounds", a: rng("C4", "G5"), b: rng("G5", "C4"), wantEqual: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.wantEqual, tt.a.Equal(tt.b))
			assert.Equal(t, tt.wantEqual, tt.b.Equal(tt.a), "symmetric")
			assert.Equal(t, tt.wantEqual, tt.a == tt.b, "== agrees with Equal")
			assert.True(t, tt.a.Equal(tt.a), "reflexive")
		})
	}

	t.Run("map key", func(t *testing.T) {
		t.Parallel()
		m := map[instrument.Range]int{rng("C4", "G5"): 1, {}: 2}
		assert.Equal(t, 1, m[rng("C4", "G5")])
		assert.Equal(t, 2, m[instrument.Range{}])
		_, ok := m[rng("B♯3", "G5")]
		assert.False(t, ok)
	})

	equivalent := []struct {
		name string
		sys  *system.System
		a, b instrument.Range
		want bool
	}{
		{name: "КП-15 traditional", sys: trad(), a: rng("B♯3", "G5"), b: rng("C4", "G5"), want: true},
		{name: "КП-15 19-position", sys: pos19(), a: rng("B♯3", "G5"), b: rng("C4", "G5"), want: false},
		{name: "upper by other names", sys: trad(), a: rng("C4", "G5"), b: rng("C4", "A𝄫5"), want: true},
		{name: "other sounds", sys: trad(), a: rng("C4", "G5"), b: rng("C4", "A5"), want: false},
		{name: "itself", sys: pos19(), a: rng("C4", "G5"), b: rng("C4", "G5"), want: true},
	}
	for _, tt := range equivalent {
		t.Run("Equivalent "+tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.a.Equivalent(tt.sys, tt.b)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want, tt.a.MustEquivalent(tt.sys, tt.b))
			assert.Equal(t, tt.want, tt.b.MustEquivalent(tt.sys, tt.a), "symmetric")
		})
	}
}

// ОП-17, К-26, А 4.10.
func TestRange_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		r    instrument.Range
	}{
		{name: "soprano", r: rng("C4", "G5")},
		{name: "КП-6", r: rangeKP6()},
		{name: "КП-7", r: rangeKP7()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := tt.r.String()
			assert.Contains(t, s, tt.r.Lower().String())
			assert.Contains(t, s, tt.r.Upper().String())
			assert.NotEqual(t, "Range(unset)", s)
		})
	}
	assert.Equal(t, "Range(unset)", instrument.Range{}.String())
}

// А 4.10, КР-18.
func TestLocation_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		l    instrument.Location
		want string
	}{
		{instrument.LocationUnset, "Location(unset)"},
		{instrument.Below, "below"},
		{instrument.Inside, "inside"},
		{instrument.Above, "above"},
		{instrument.Location(4), "Location(4)"},
		{instrument.Location(255), "Location(255)"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.l.String())
		})
	}
	var zero instrument.Location
	assert.Equal(t, instrument.LocationUnset, zero)
}

// Must variants of the range operations: they return the results of the
// operation without the error and panic with its error (dev.md «Must»).
func TestRange_Must(t *testing.T) {
	t.Parallel()
	inverted := rangeKP6()
	sys := pos19()
	other := rng("C4", "E4")
	x := p("D4")
	musts := map[string]func(){
		"MustCheck":      func() { inverted.MustCheck(sys) },
		"MustContains":   func() { inverted.MustContains(sys, x) },
		"MustLocate":     func() { inverted.MustLocate(sys, x) },
		"MustWidth":      func() { inverted.MustWidth(sys) },
		"MustCount":      func() { inverted.MustCount(sys) },
		"MustFolds":      func() { inverted.MustFolds(sys, x) },
		"MustFold":       func() { inverted.MustFold(sys, x) },
		"MustIntersect":  func() { inverted.MustIntersect(sys, other) },
		"MustIncludes":   func() { inverted.MustIncludes(sys, other) },
		"MustHull":       func() { inverted.MustHull(sys, other) },
		"MustPitches":    func() { inverted.MustPitches(sys, system.Sharps()) },
		"MustEquivalent": func() { inverted.MustEquivalent(sys, other) },
	}
	for name, f := range musts {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assertPanicsWith(t, instrument.ErrInverted, nil, f)
		})
	}

	t.Run("no panic on success", func(t *testing.T) {
		t.Parallel()
		r := rng("C4", "G5")
		assertNoPanic(t, func() {
			r.MustCheck(trad())
			r.MustContains(trad(), x)
			r.MustLocate(trad(), x)
			r.MustWidth(trad())
			r.MustCount(trad())
			r.MustFolds(trad(), x)
			r.MustFold(trad(), x)
			r.MustIntersect(trad(), other)
			r.MustIncludes(trad(), other)
			r.MustHull(trad(), other)
			r.MustPitches(trad(), system.Flats())
			r.MustEquivalent(trad(), other)
		})
	})

	t.Run("chains", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, system.Step(19), instrument.MustNewRange(p("C4"), p("G5")).MustCheck(trad()).MustWidth(trad()))
	})
}
