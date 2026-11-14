package interval_test

import (
	"iter"
	"slices"
	"testing"
	"time"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T-6: positional size, signed steps, quality by size, enharmonic equality,
// intervals of a size and enharmonic respellings (OP-34...OP-39, OP-46;
// P-18, P-19, P-24; table 2.6; KP-11...KP-14, KP-17; architecture 4.11).

// collect returns the elements of a sequence; nil for a nil sequence.
func collect[T any](seq iter.Seq[T]) []T {
	if seq == nil {
		return nil
	}

	return slices.Collect(seq)
}

// assertIntervals checks that got lists exactly want, in order.
func assertIntervals(t *testing.T, want, got []interval.Interval) {
	t.Helper()

	if !assert.Len(t, got, len(want), "got %v, want %v", got, want) {
		return
	}

	for k := range want {
		assert.True(t, got[k].Equal(want[k]), "element %d: got %v, want %v", k, got[k], want[k])
	}
}

// ---------------------------------------------------------------------------
// Steps and SignedSteps (OP-34, OP-35, P-18; table 2.6; KP-11).

func TestPositionalStepsTable(t *testing.T) {
	t.Parallel()

	systems := []*system.System{trad(), pos19(), qt24(), sf31()}
	table := []struct {
		i     string
		steps [4]system.Step
	}{
		{"P1", [4]system.Step{0, 0, 0, 0}},
		{"m2", [4]system.Step{1, 2, 2, 3}},
		{"M2", [4]system.Step{2, 3, 4, 5}},
		{"m3", [4]system.Step{3, 5, 6, 8}},
		{"M3", [4]system.Step{4, 6, 8, 10}},
		{"P4", [4]system.Step{5, 8, 10, 13}},
		{"A4", [4]system.Step{6, 9, 12, 15}},
		{"d5", [4]system.Step{6, 10, 12, 16}},
		{"P5", [4]system.Step{7, 11, 14, 18}},
		{"m6", [4]system.Step{8, 13, 16, 21}},
		{"M6", [4]system.Step{9, 14, 18, 23}},
		{"m7", [4]system.Step{10, 16, 20, 26}},
		{"M7", [4]system.Step{11, 17, 22, 28}},
		{"P8", [4]system.Step{12, 19, 24, 31}},
	}
	starts := []spelling.Spelling{spelling.C(), spelling.FSharp(), spelling.BFlat(), spelling.E(), spelling.GDoubleSharp()}

	for _, row := range table {
		for k, sys := range systems {
			t.Run(row.i+" in "+sys.Name(), func(t *testing.T) {
				t.Parallel()

				i := iv(row.i)
				want := row.steps[k]

				got, err := i.Steps(sys, noStart)
				require.NoError(t, err)
				assert.Equal(t, want, got, "without a start name")

				signed, err := i.SignedSteps(sys, noStart)
				require.NoError(t, err)
				assert.Equal(t, want, signed)

				down := i.MustDown()
				assert.Equal(t, want, down.MustSteps(sys, noStart), "the size does not depend on the direction")

				if !i.IsPerfectUnison() {
					assert.Equal(t, -want, down.MustSignedSteps(sys, noStart), "the signed steps of a descending interval")
				}

				for _, x := range starts {
					assert.Equal(t, want, i.MustSteps(sys, x), "from %v: a regular system", x)
					assert.Equal(t, want, down.MustSteps(sys, x), "down from %v", x)
				}
			})
		}
	}
}

func TestPositionalSteps(t *testing.T) {
	t.Parallel()

	tr, s19, s24, irr := trad(), pos19(), qt24(), irregular()

	tests := []struct {
		name   string
		sys    *system.System
		i      string
		start  spelling.Spelling
		steps  system.Step
		signed system.Step
	}{
		{"d2", tr, "d2", noStart, 0, 0},
		{"dd2", tr, "dd2", noStart, -1, -1},
		{"A1", tr, "A1", noStart, 1, 1},
		{"M10", tr, "M10", noStart, 16, 16},
		{"P15", tr, "P15", noStart, 24, 24},
		{"A8", tr, "A8", noStart, 13, 13},
		{"d8", tr, "d8", noStart, 11, 11},
		{"descending M3", tr, "↓M3", noStart, 4, -4},
		{"descending dd2", tr, "↓dd2", noStart, -1, 1},
		{"descending A1", tr, "↓A1", spelling.C(), 1, -1},
		{"M10 from B", tr, "M10", spelling.B(), 16, 16},
		{"19: A5", s19, "A5", noStart, 12, 12},
		{"19: d8", s19, "d8", noStart, 18, 18},
		{"19: d2", s19, "d2", noStart, 1, 1},
		{"24: n3", s24, "n3", noStart, 7, 7},
		{"24: n3 from C", s24, "n3", spelling.C(), 7, 7},
		{"24: descending n3 from C", s24, "↓n3", spelling.C(), 7, -7},
		{"24: unison +1/2", s24, "{+1/2}1", noStart, 1, 1},
		{"irregular: M2 from C", irr, "M2", spelling.C(), 3, 3},
		{"irregular: M2 from D", irr, "M2", spelling.D(), 1, 1},
		{"irregular: M2 from E", irr, "M2", spelling.E(), 2, 2},
		{"irregular: descending M2 from D (lower name C)", irr, "↓M2", spelling.D(), 3, -3},
		{"irregular: P5 from C", irr, "P5", spelling.C(), 7, 7},
		{"irregular: P5 from D", irr, "P5", spelling.D(), 6, 6},
		{"irregular: m7 from D", irr, "m7", spelling.D(), 9, 9},
		{"irregular: descending m7 from C (lower name D)", irr, "↓m7", spelling.C(), 9, -9},
		{"irregular: m2 from B", irr, "m2", spelling.B(), 1, 1},
		{"irregular: M9 from C", irr, "M9", spelling.C(), 15, 15},
		{"irregular: P1 from G", irr, "P1", spelling.G(), 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			i := iv(tt.i)

			got, err := i.Steps(tt.sys, tt.start)
			require.NoError(t, err)
			assert.Equal(t, tt.steps, got)

			signed, err := i.SignedSteps(tt.sys, tt.start)
			require.NoError(t, err)
			assert.Equal(t, tt.signed, signed)
		})
	}

	t.Run("inversions add to N in the irregular system", func(t *testing.T) {
		t.Parallel()

		// P-18: M2 from C and m7 from D; ↓M2 from D and ↓m7 from C.
		assert.Equal(t, system.Step(12), iv("M2").MustSteps(irr, spelling.C())+iv("m7").MustSteps(irr, spelling.D()))
		assert.Equal(t, system.Step(12), iv("↓M2").MustSteps(irr, spelling.D())+iv("↓m7").MustSteps(irr, spelling.C()))
	})
}

func TestPositionalStepsErrors(t *testing.T) {
	t.Parallel()

	tr, irr := trad(), irregular()
	cPlus := spelling.MustNew(letter.C, ai(mag))

	tests := []struct {
		name   string
		sys    *system.System
		i      interval.Interval
		start  spelling.Spelling
		want   error
		causes []error
	}{
		{"nil system", nil, iv("M3"), noStart, interval.ErrUnset, nil},
		{"unset system", new(system.System), iv("M3"), noStart, interval.ErrUnset, nil},
		{"unset interval", tr, interval.Interval{}, noStart, interval.ErrUnset, nil},
		{"unset interval before the start name requirement", irr, interval.Interval{}, noStart, interval.ErrUnset, nil},
		{"irregular without start (KP-11)", irr, iv("M2"), noStart, interval.ErrStartRequired, nil},
		{"irregular without start, fractional: E-8 before E-9", irr, iv("n3"), noStart, interval.ErrStartRequired, nil},
		{"irregular without start, even for P1", irr, iv("P1"), noStart, interval.ErrStartRequired, nil},
		{"inadmissible start", tr, iv("M3"), eHalfFlat(), nil, []error{system.ErrInadmissible}},
		{"inadmissible start in the irregular system", irr, iv("M2"), eHalfFlat(), nil, []error{system.ErrInadmissible}},
		{"n3 without start in the traditional system (KP-11)", tr, iv("n3"), noStart, interval.ErrFractionalSteps, nil},
		{"descending n3 without start", tr, iv("↓n3"), noStart, interval.ErrFractionalSteps, nil},
		{"31: third -1/4 without start", sf31(), iv("{-1/4}3"), noStart, interval.ErrFractionalSteps, nil},
		{"n3 from C in the traditional system (KP-11)", tr, iv("n3"), spelling.C(), nil, []error{system.ErrInadmissible}},
		{"alteration of the built name beyond M", tr, iv("A1"), cPlus, interval.ErrOutOfRange, []error{alteration.ErrOutOfRange}},
		{"size beyond int", tr, basicSize(interval.MaxSize), noStart, interval.ErrOutOfRange, nil},
		{"size beyond int from a start", tr, basicSize(interval.MaxSize), spelling.C(), interval.ErrOutOfRange, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.i.Steps(tt.sys, tt.start)
			assertErr(t, err, tt.want, tt.causes...)
			assert.Zero(t, got)

			signed, err := tt.i.SignedSteps(tt.sys, tt.start)
			assertErr(t, err, tt.want, tt.causes...)
			assert.Zero(t, signed)
		})
	}
}

// ---------------------------------------------------------------------------
// FromSteps (OP-36, P-19; KP-12).

func TestPositionalFromSteps(t *testing.T) {
	t.Parallel()

	tr, irr := trad(), irregular()

	tests := []struct {
		name  string
		sys   *system.System
		n     int
		s     system.Step
		start spelling.Spelling
		want  string
	}{
		{"4 and 6", tr, 4, 6, noStart, "A4"},
		{"3 and 2", tr, 3, 2, noStart, "d3"},
		{"5 and 6", tr, 5, 6, noStart, "d5"},
		{"2 and 0", tr, 2, 0, noStart, "d2"},
		{"1 and 1", tr, 1, 1, noStart, "A1"},
		{"10 and 16", tr, 10, 16, noStart, "M10"},
		{"1 and 0", tr, 1, 0, noStart, "P1"},
		{"8 and 12", tr, 8, 12, noStart, "P8"},
		{"2 and -1", tr, 2, -1, noStart, "dd2"},
		{"3 and 4 from C", tr, 3, 4, spelling.C(), "M3"},
		{"3 and 4 from B sharp", tr, 3, 4, spelling.BSharp(), "M3"},
		{"24: 3 and 7", qt24(), 3, 7, noStart, "n3"},
		{"24: 3 and 7 from C", qt24(), 3, 7, spelling.C(), "n3"},
		{"24: 3 and 9", qt24(), 3, 9, noStart, "{+1/2}3"},
		{"31: 3 and 9", sf31(), 3, 9, noStart, "n3"},
		{"31: 2 and 4", sf31(), 2, 4, noStart, "n2"},
		{"31: 5 and 17", sf31(), 5, 17, noStart, "{-1/2}5"},
		{"irregular: 2 and 3 from C", irr, 2, 3, spelling.C(), "M2"},
		{"irregular: 2 and 2 from D", irr, 2, 2, spelling.D(), "A2"},
		{"irregular: 3 and 3 from D", irr, 3, 3, spelling.D(), "M3"},
		{"irregular: 4 and 3 from D", irr, 4, 3, spelling.D(), "d4"},
		{"irregular: 4 and 4 from D", irr, 4, 4, spelling.D(), "P4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := interval.FromSteps(tt.sys, tt.n, tt.s, tt.start)
			require.NoError(t, err)
			assert.True(t, got.Equal(iv(tt.want)), "got %v, want %s", got, tt.want)
		})
	}

	t.Run("edge of the deviation", func(t *testing.T) {
		t.Parallel()

		got, err := interval.FromSteps(tr, 3, system.Step(4+mag), noStart)
		require.NoError(t, err)
		assert.True(t, got.Equal(aug(mag, 3)))

		got, err = interval.FromSteps(tr, 5, system.Step(7-mag), noStart)
		require.NoError(t, err)
		assert.True(t, got.Equal(dim(mag, 5)))
	})
}

func TestPositionalFromStepsErrors(t *testing.T) {
	t.Parallel()

	tr, irr := trad(), irregular()

	tests := []struct {
		name   string
		sys    *system.System
		n      int
		s      system.Step
		start  spelling.Spelling
		want   error
		causes []error
	}{
		{"nil system", nil, 3, 4, noStart, interval.ErrUnset, nil},
		{"unset system", new(system.System), 3, 4, noStart, interval.ErrUnset, nil},
		{"nil system before size 0", nil, 0, 4, noStart, interval.ErrUnset, nil},
		{"size 0", tr, 0, 0, noStart, interval.ErrInvalidArgument, nil},
		{"negative size", tr, -2, 0, noStart, interval.ErrInvalidArgument, nil},
		{"size MinInt", tr, minInt, 0, noStart, interval.ErrInvalidArgument, nil},
		{"unison of negative size (KP-12)", tr, 1, -1, noStart, interval.ErrInvalidArgument, nil},
		{"unison of negative size from a start", tr, 1, -1, spelling.C(), interval.ErrInvalidArgument, nil},
		{"size 0 before the start name requirement", irr, 0, 3, noStart, interval.ErrInvalidArgument, nil},
		{"irregular without start", irr, 2, 3, noStart, interval.ErrStartRequired, nil},
		{"inadmissible start", tr, 3, 4, eHalfFlat(), nil, []error{system.ErrInadmissible}},
		{"31: inadmissible built name (KP-12)", sf31(), 3, 9, spelling.C(), nil, []error{system.ErrInadmissible}},
		{"size of the main kind beyond int", tr, interval.MaxSize, 0, noStart, interval.ErrOutOfRange, nil},
		{"deviation beyond M", tr, 3, system.Step(4 + mag + 1), noStart, interval.ErrOutOfRange, nil},
		{"deviation beyond -M", tr, 5, system.Step(7 - mag - 1), noStart, interval.ErrOutOfRange, nil},
		{"alteration of the built name beyond M", tr, 2, system.Step(2 + mag + 1), spelling.C(), interval.ErrOutOfRange, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := interval.FromSteps(tt.sys, tt.n, tt.s, tt.start)
			assertErr(t, err, tt.want, tt.causes...)
			assert.Equal(t, interval.Interval{}, got)
		})
	}
}

// ---------------------------------------------------------------------------
// EnharmonicallyEqual (OP-37; KP-13).

func TestPositionalEnharmonicallyEqual(t *testing.T) {
	t.Parallel()

	tr, s19, irr := trad(), pos19(), irregular()

	tests := []struct {
		name   string
		sys    *system.System
		i      string
		iStart spelling.Spelling
		j      string
		jStart spelling.Spelling
		want   bool
	}{
		{"A4 d5", tr, "A4", noStart, "d5", noStart, true},
		{"M3 d4", tr, "M3", noStart, "d4", noStart, true},
		{"m3 A2", tr, "m3", noStart, "A2", noStart, true},
		{"m2 A1", tr, "m2", noStart, "A1", noStart, true},
		{"M6 d7", tr, "M6", noStart, "d7", noStart, true},
		{"M3 m3", tr, "M3", noStart, "m3", noStart, false},
		{"M3 and descending d4", tr, "M3", noStart, "↓d4", noStart, true},
		{"P8 P1", tr, "P8", noStart, "P1", noStart, false},
		{"A7 P8", tr, "A7", spelling.C(), "P8", spelling.D(), true},
		{"19: A4 d5", s19, "A4", noStart, "d5", noStart, false},
		{"19: A1 d2", s19, "A1", noStart, "d2", noStart, true},
		{"24: n3 M3", qt24(), "n3", noStart, "M3", noStart, false},
		{"24: n3 and third -1/2 from C", qt24(), "n3", noStart, "n3", spelling.C(), true},
		{"irregular: M2 from C, m3 from E", irr, "M2", spelling.C(), "m3", spelling.E(), true},
		{"irregular: M2 from C, M2 from D", irr, "M2", spelling.C(), "M2", spelling.D(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := interval.EnharmonicallyEqual(tt.sys, iv(tt.i), tt.iStart, iv(tt.j), tt.jStart)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)

			got, err = interval.EnharmonicallyEqual(tt.sys, iv(tt.j), tt.jStart, iv(tt.i), tt.iStart)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got, "symmetric")
		})
	}

	t.Run("errors", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name   string
			sys    *system.System
			i      interval.Interval
			iStart spelling.Spelling
			j      interval.Interval
			jStart spelling.Spelling
			want   error
			causes []error
		}{
			{"nil system", nil, iv("A4"), noStart, iv("d5"), noStart, interval.ErrUnset, nil},
			{"unset i", tr, interval.Interval{}, noStart, iv("d5"), noStart, interval.ErrUnset, nil},
			{"unset j", tr, iv("A4"), noStart, interval.Interval{}, noStart, interval.ErrUnset, nil},
			{"unset j before the start of i", irr, iv("A4"), noStart, interval.Interval{}, noStart, interval.ErrUnset, nil},
			{"irregular without starts", irr, iv("M2"), noStart, iv("m3"), noStart, interval.ErrStartRequired, nil},
			{"irregular without the start of j", irr, iv("M2"), spelling.C(), iv("m3"), noStart, interval.ErrStartRequired, nil},
			{"start required before an inadmissible start", irr, iv("M2"), eHalfFlat(), iv("m3"), noStart, interval.ErrStartRequired, nil},
			{"inadmissible start of j", tr, iv("M2"), spelling.C(), iv("m3"), eHalfFlat(), nil, []error{system.ErrInadmissible}},
			{"inadmissible start before E-9", tr, iv("n3"), noStart, iv("m3"), eHalfFlat(), nil, []error{system.ErrInadmissible}},
			{"fractional steps", tr, iv("n3"), noStart, iv("M3"), noStart, interval.ErrFractionalSteps, nil},
			{"inadmissible built name", tr, iv("n3"), spelling.C(), iv("M3"), noStart, nil, []error{system.ErrInadmissible}},
			{"size beyond int", tr, basicSize(interval.MaxSize), noStart, iv("M3"), noStart, interval.ErrOutOfRange, nil},
		}

		for _, c := range cases {
			got, err := interval.EnharmonicallyEqual(c.sys, c.i, c.iStart, c.j, c.jStart)
			assertErr(t, err, c.want, c.causes...)
			assert.False(t, got, c.name)
		}
	})
}

// ---------------------------------------------------------------------------
// AllFromSteps and Enharmonics (OP-38, OP-39, P-19; KP-14).

func TestPositionalAllFromSteps(t *testing.T) {
	t.Parallel()

	tr, irr := trad(), irregular()

	tests := []struct {
		name  string
		sys   *system.System
		s     system.Step
		k     int
		start spelling.Spelling
		want  []string
	}{
		{"6, K = 2", tr, 6, 2, noStart, []string{"AA3", "A4", "d5", "dd6"}},
		{"3, K = 1", tr, 3, 1, noStart, []string{"A2", "m3"}},
		{"0, K = 2", tr, 0, 2, noStart, []string{"P1", "d2"}},
		{"12, K = 1", tr, 12, 1, noStart, []string{"A7", "P8", "d9"}},
		{"4, K = 0", tr, 4, 0, noStart, []string{"M3"}},
		{"-1, K = 2", tr, -1, 2, noStart, []string{"dd2"}},
		{"13, K = 0", tr, 13, 0, noStart, []string{"m9"}},
		{"6, K = 2 from C", tr, 6, 2, spelling.C(), []string{"AA3", "A4", "d5", "dd6"}},
		{"19: 9, K = 2", pos19(), 9, 2, noStart, []string{"A4", "dd5"}},
		{"24: 7, K = 1", qt24(), 7, 1, noStart, []string{"n3"}},
		{"31: 9, K = 1", sf31(), 9, 1, noStart, []string{"n3"}},
		{"31: 9, K = 1 from C: E½♭ is not admissible", sf31(), 9, 1, spelling.C(), nil},
		{"31: 9, K = 2 from C", sf31(), 9, 2, spelling.C(), []string{"AA2", "dd4"}},
		{"irregular: 3, K = 1 from C", irr, 3, 1, spelling.C(), []string{"M2", "m3"}},
		{"irregular: 3, K = 1 from D", irr, 3, 1, spelling.D(), []string{"M3", "d4"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seq, err := interval.AllFromSteps(tt.sys, tt.s, tt.k, tt.start)
			require.NoError(t, err)
			require.NotNil(t, seq)

			got := collect(seq)
			assertIntervals(t, ivs(tt.want...), got)
			assertIntervals(t, ivs(tt.want...), collect(seq))

			for _, i := range got {
				assert.True(t, i.IsAscending())
				assert.Equal(t, tt.s, i.MustSteps(tt.sys, tt.start), "%v", i)
			}
		})
	}

	t.Run("the sequence can be stopped", func(t *testing.T) {
		t.Parallel()

		seq := interval.MustAllFromSteps(tr, 6, 2, noStart)

		var got []interval.Interval

		for i := range seq {
			got = append(got, i)
			if len(got) == 2 {
				break
			}
		}

		assertIntervals(t, ivs("AA3", "A4"), got)
	})

	t.Run("errors", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name   string
			sys    *system.System
			k      int
			start  spelling.Spelling
			want   error
			causes []error
		}{
			{"nil system", nil, 1, noStart, interval.ErrUnset, nil},
			{"nil system before a negative limit", nil, -1, noStart, interval.ErrUnset, nil},
			{"negative limit", tr, -1, noStart, interval.ErrInvalidArgument, nil},
			{"limit MinInt", tr, minInt, noStart, interval.ErrInvalidArgument, nil},
			{"negative limit before the start name requirement", irr, -1, noStart, interval.ErrInvalidArgument, nil},
			{"irregular without start (KP-14)", irr, 1, noStart, interval.ErrStartRequired, nil},
			{"inadmissible start", tr, 1, eHalfFlat(), nil, []error{system.ErrInadmissible}},
		}

		for _, c := range cases {
			seq, err := interval.AllFromSteps(c.sys, 3, c.k, c.start)
			assertErr(t, err, c.want, c.causes...)
			assert.Nil(t, seq, c.name)
		}
	})
}

func TestPositionalEnharmonics(t *testing.T) {
	t.Parallel()

	tr, irr := trad(), irregular()

	tests := []struct {
		name  string
		sys   *system.System
		i     string
		k     int
		start spelling.Spelling
		want  []string
	}{
		{"A4, K = 2", tr, "A4", 2, noStart, []string{"AA3", "A4", "d5", "dd6"}},
		{"descending M3, K = 1", tr, "↓M3", 1, noStart, []string{"M3", "d4"}},
		{"A4, K = 0: the interval itself is not within the limit", tr, "A4", 0, noStart, nil},
		{"M3 from C, K = 1", tr, "M3", 1, spelling.C(), []string{"M3", "d4"}},
		{"irregular: M2 from C, K = 1", irr, "M2", 1, spelling.C(), []string{"M2", "m3"}},
		{"irregular: descending M2 from D, K = 1 (lower name C)", irr, "↓M2", 1, spelling.D(), []string{"M2", "m3"}},
		{"irregular: M2 from D, K = 1", irr, "M2", 1, spelling.D(), []string{"A1", "M2", "d3"}},
		{"24: n3, K = 1", qt24(), "n3", 1, noStart, []string{"n3"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seq, err := iv(tt.i).Enharmonics(tt.sys, tt.k, tt.start)
			require.NoError(t, err)
			require.NotNil(t, seq)
			assertIntervals(t, ivs(tt.want...), collect(seq))
		})
	}

	t.Run("errors", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name   string
			sys    *system.System
			i      interval.Interval
			k      int
			start  spelling.Spelling
			want   error
			causes []error
		}{
			{"nil system", nil, iv("A4"), 1, noStart, interval.ErrUnset, nil},
			{"unset interval", tr, interval.Interval{}, 1, noStart, interval.ErrUnset, nil},
			{"negative limit", tr, iv("A4"), -1, noStart, interval.ErrInvalidArgument, nil},
			{"unset interval before a negative limit", tr, interval.Interval{}, -1, noStart, interval.ErrUnset, nil},
			{"irregular without start", irr, iv("M2"), 1, noStart, interval.ErrStartRequired, nil},
			{"negative limit before the start requirement", irr, iv("M2"), -1, noStart, interval.ErrInvalidArgument, nil},
			{"inadmissible start", tr, iv("M3"), 1, eHalfFlat(), nil, []error{system.ErrInadmissible}},
			{"fractional steps", tr, iv("n3"), 1, noStart, interval.ErrFractionalSteps, nil},
			{"inadmissible built name", tr, iv("n3"), 1, spelling.C(), nil, []error{system.ErrInadmissible}},
			{"size beyond int", tr, basicSize(interval.MaxSize), 1, noStart, interval.ErrOutOfRange, nil},
		}

		for _, c := range cases {
			seq, err := c.i.Enharmonics(c.sys, c.k, c.start)
			assertErr(t, err, c.want, c.causes...)
			assert.Nil(t, seq, c.name)
		}
	})
}

// TestPositionalAllFromStepsSweep checks P-19 on a sweep: every listed
// interval has the size, is within the limit, the list is ordered by size
// and contains every interval of the predefined set that qualifies.
func TestPositionalAllFromStepsSweep(t *testing.T) {
	t.Parallel()

	for _, sys := range []*system.System{trad(), pos19(), qt24(), sf31()} {
		t.Run(sys.Name(), func(t *testing.T) {
			t.Parallel()

			candidates := interval.Predefined()
			for n := 1; n <= 22; n++ {
				for _, d := range []alteration.Alteration{alt(-1, 2), alt(1, 2), ai(-3), ai(3), ai(-4)} {
					if n == 1 && d.Sign() < 0 {
						continue
					}

					candidates = append(candidates, interval.MustFromDeviation(n, d))
				}
			}

			for s := system.Step(-3); s <= system.Step(40); s++ {
				for k := 0; k <= 3; k++ {
					got := collect(interval.MustAllFromSteps(sys, s, k, noStart))

					for idx, i := range got {
						assert.Equal(t, s, i.MustSteps(sys, noStart), "%v", i)
						assert.LessOrEqual(t, remoteness(i).Compare(ai(k)), 0, "%v within %d", i, k)

						if idx > 0 {
							assert.Less(t, got[idx-1].Size(), i.Size(), "ordered by size")
						}
					}

					for _, c := range candidates {
						steps, err := c.Steps(sys, noStart)
						if err != nil || steps != s || remoteness(c).Compare(ai(k)) > 0 {
							continue
						}

						assert.True(t, slices.ContainsFunc(got, c.Equal), "%v of size %d within %d is listed", c, s, k)
					}
				}
			}
		})
	}
}

// within runs f in a goroutine and returns its result; the test fails at
// once if f does not return in 5 seconds (NF-2: the time does not grow with
// the size of an interval and the number of octaves; architecture 6: the
// window of AllFromSteps starts by division, not by a scan from size 1).
func within[T any](t *testing.T, f func() T) T {
	t.Helper()

	type outcome struct {
		v     T
		panic any
	}

	done := make(chan outcome, 1)

	go func() {
		var o outcome

		defer func() {
			o.panic = recover()
			done <- o
		}()

		o.v = f()
	}()

	select {
	case o := <-done:
		if o.panic != nil {
			t.Fatalf("panic: %v", o.panic)
		}

		return o.v
	case <-time.After(5 * time.Second):
		t.Fatal("the call did not return in 5 seconds")

		var zero T

		return zero
	}
}

func TestPositionalLargeValues(t *testing.T) {
	t.Parallel()

	tr := trad()

	const k = 1_000_000_000_000_000 // octaves

	t.Run("AllFromSteps far from size 1", func(t *testing.T) {
		t.Parallel()

		got := within(t, func() []interval.Interval {
			return collect(interval.MustAllFromSteps(tr, system.Step(12*k), 0, noStart))
		})
		assertIntervals(t, []interval.Interval{interval.MustFromDeviation(7*k+1, ai(0))}, got)
	})

	t.Run("Steps and FromSteps of a large size", func(t *testing.T) {
		t.Parallel()

		big := interval.MustFromDeviation(7*k+3, ai(0)) // a major third and k octaves
		steps := within(t, func() system.Step { return big.MustSteps(tr, spelling.C()) })
		assert.Equal(t, system.Step(12*k+4), steps)

		back := within(t, func() interval.Interval { return interval.MustFromSteps(tr, 7*k+3, steps, spelling.C()) })
		assert.True(t, back.Equal(big))
	})

	t.Run("transposition of a pitch by many octaves", func(t *testing.T) {
		t.Parallel()

		big := interval.MustFromDeviation(7*k+3, ai(0))
		got := within(t, func() pitch.Pitch { return interval.MustTransposePitch(tr, pt(spelling.C(), 0), big) })
		assert.True(t, got.Equal(pt(spelling.E(), k)))

		i := within(t, func() interval.Interval { return interval.MustBetweenPitches(tr, pt(spelling.C(), 0), got) })
		assert.True(t, i.Equal(big))
	})

	t.Run("arithmetic of MaxSize", func(t *testing.T) {
		t.Parallel()

		big := basicSize(interval.MaxSize)
		got := within(t, func() interval.Interval { return big.MustAdd(iv("↓P8")).MustSub(iv("↓P8")) })
		assert.True(t, got.Equal(big))

		name := within(t, func() spelling.Spelling { return interval.MustTranspose(spelling.C(), big) })
		assert.True(t, name.Equal(spelling.B()))
	})
}

// remoteness is ρ of P-19 and architecture 4.11.
func remoteness(i interval.Interval) alteration.Alteration {
	d := i.Deviation()

	if i.Group() == interval.PerfectGroup {
		if d.Sign() < 0 {
			return alteration.Natural().MustSub(d)
		}

		return d
	}

	switch {
	case d.Sign() > 0:
		return d
	case d.Compare(ai(-1)) < 0:
		return ai(-1).MustSub(d)
	default:
		return alteration.Natural()
	}
}

// ---------------------------------------------------------------------------
// Properties of the positional size on a sweep (P-18, P-19; architecture
// 4.11, "Properties").

func TestPositionalProperties(t *testing.T) {
	t.Parallel()

	starts := []spelling.Spelling{spelling.C(), spelling.D(), spelling.E(), spelling.FSharp(), spelling.BFlat(), spelling.B()}

	for _, sys := range []*system.System{trad(), pos19(), qt24(), sf31(), irregular(), boundaryA()} {
		t.Run(sys.Name(), func(t *testing.T) {
			t.Parallel()

			n := system.Step(sys.Classes())
			c := system.Step(sys.Chromatic())
			reg, err := sys.Regularity()
			require.NoError(t, err)

			for _, x := range starts {
				for _, i := range predefinedBoth() {
					s := i.MustSteps(sys, x)
					y := interval.MustTranspose(x, i)

					if i.IsDescending() {
						assert.Equal(t, s, i.MustUp().MustSteps(sys, y), "lower name of %v from %v", i, x)
						assert.Equal(t, -s, i.MustSignedSteps(sys, x))
						assert.True(t, interval.MustFromSteps(sys, i.Size(), s, y).Equal(i.MustUp()), "FromSteps of %v from %v", i, x)
					} else {
						assert.Equal(t, s, i.MustSignedSteps(sys, x))
						assert.True(t, interval.MustFromSteps(sys, i.Size(), s, x).Equal(i), "FromSteps of %v from %v", i, x)
					}

					if reg.Regular {
						assert.Equal(t, s, i.MustSteps(sys, noStart), "regular: %v from %v", i, x)
					}

					if i.IsAscending() {
						assert.Equal(t, s+c, i.MustExpand().MustSteps(sys, x), "a quality step adds c: %v from %v", i, x)
					}

					assert.Equal(t, s+n, i.MustAddOctaves(1).MustSteps(sys, x), "an octave adds N: %v from %v", i, x)

					if inv, err := i.Invert(); err == nil && i.IsSimple() {
						assert.Equal(t, n, s+inv.MustSteps(sys, y), "P-18: %v from %v and its inversion from %v", i, x, y)
					}
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Respellings (OP-46, P-24; KP-17).

// expectedRespellings builds the list of P-24 from sys.EnharmonicPitches.
func expectedRespellings(t *testing.T, sys *system.System, x, y pitch.Pitch, limit alteration.Alteration, which interval.Replace) []interval.Respelling {
	t.Helper()

	froms := []pitch.Pitch{x}
	tos := []pitch.Pitch{y}

	if which == interval.ReplaceFrom || which == interval.ReplaceBoth {
		froms = collect(sys.MustEnharmonicPitches(x, limit))
	}

	if which == interval.ReplaceTo || which == interval.ReplaceBoth {
		tos = collect(sys.MustEnharmonicPitches(y, limit))
	}

	var out []interval.Respelling

	for _, f := range froms {
		for _, to := range tos {
			out = append(out, interval.Respelling{From: f, To: to, Interval: interval.MustBetweenPitches(sys, f, to)})
		}
	}

	return out
}

func assertRespellings(t *testing.T, want, got []interval.Respelling) {
	t.Helper()

	if !assert.Len(t, got, len(want)) {
		return
	}

	for k := range want {
		assert.True(t, got[k].From.Equal(want[k].From), "%d: From %v, want %v", k, got[k].From, want[k].From)
		assert.True(t, got[k].To.Equal(want[k].To), "%d: To %v, want %v", k, got[k].To, want[k].To)
		assert.True(t, got[k].Interval.Equal(want[k].Interval), "%d: Interval %v, want %v", k, got[k].Interval, want[k].Interval)
	}
}

func TestPositionalRespellings(t *testing.T) {
	t.Parallel()

	tr := trad()
	c4, e4 := pt(spelling.C(), 4), pt(spelling.E(), 4)
	two := alteration.DoubleSharp()

	t.Run("KP-17: replace the end", func(t *testing.T) {
		t.Parallel()

		seq, err := interval.Respellings(tr, c4, e4, two, interval.ReplaceTo)
		require.NoError(t, err)

		got := collect(seq)
		assertRespellings(t, expectedRespellings(t, tr, c4, e4, two, interval.ReplaceTo), got)

		byEnd := map[spelling.Spelling]interval.Interval{}
		for _, r := range got {
			assert.True(t, r.From.Equal(c4))
			byEnd[r.To.Spelling()] = r.Interval
		}

		require.Len(t, byEnd, 3)
		assert.True(t, byEnd[spelling.DDoubleSharp()].Equal(iv("AA2")))
		assert.True(t, byEnd[spelling.E()].Equal(iv("M3")))
		assert.True(t, byEnd[spelling.FFlat()].Equal(iv("d4")))
	})

	t.Run("KP-17: replace both", func(t *testing.T) {
		t.Parallel()

		seq, err := interval.Respellings(tr, c4, e4, two, interval.ReplaceBoth)
		require.NoError(t, err)

		got := collect(seq)
		require.Len(t, got, 9)
		assertRespellings(t, expectedRespellings(t, tr, c4, e4, two, interval.ReplaceBoth), got)

		type pair struct{ from, to spelling.Spelling }

		want := map[pair]interval.Interval{
			{spelling.C(), spelling.DDoubleSharp()}:           iv("AA2"),
			{spelling.C(), spelling.E()}:                      iv("M3"),
			{spelling.C(), spelling.FFlat()}:                  iv("d4"),
			{spelling.DDoubleFlat(), spelling.DDoubleSharp()}: aug(4, 1),
			{spelling.DDoubleFlat(), spelling.E()}:            iv("AA2"),
			{spelling.DDoubleFlat(), spelling.FFlat()}:        iv("M3"),
			{spelling.BSharp(), spelling.DDoubleSharp()}:      iv("M3"),
			{spelling.BSharp(), spelling.E()}:                 iv("d4"),
			{spelling.BSharp(), spelling.FFlat()}:             dim(3, 5),
		}

		for _, r := range got {
			w, ok := want[pair{r.From.Spelling(), r.To.Spelling()}]
			if assert.True(t, ok, "pair %v %v", r.From, r.To) {
				assert.True(t, r.Interval.Equal(w), "%v → %v: %v, want %v", r.From, r.To, r.Interval, w)
			}
		}

		// B♯ is B♯3, of Octave 4.
		assert.True(t, slices.ContainsFunc(got, func(r interval.Respelling) bool {
			return r.From.Equal(wp(tr, spelling.BSharp(), 3))
		}))
	})

	t.Run("replace the start", func(t *testing.T) {
		t.Parallel()

		seq, err := interval.Respellings(tr, c4, e4, two, interval.ReplaceFrom)
		require.NoError(t, err)
		assertRespellings(t, expectedRespellings(t, tr, c4, e4, two, interval.ReplaceFrom), collect(seq))
	})

	t.Run("the positions are kept", func(t *testing.T) {
		t.Parallel()

		pairs := [][2]pitch.Pitch{
			{c4, e4}, {pt(spelling.FSharp(), 3), pt(spelling.BFlat(), 5)}, {pt(spelling.B(), 4), pt(spelling.C(), 4)},
		}

		for _, p := range pairs {
			dist := tr.MustDistance(p[0], p[1])

			for _, which := range []interval.Replace{interval.ReplaceFrom, interval.ReplaceTo, interval.ReplaceBoth} {
				for _, l := range []alteration.Alteration{alteration.Natural(), alteration.Sharp(), two, ai(3)} {
					got := collect(interval.MustRespellings(tr, p[0], p[1], l, which))
					assertRespellings(t, expectedRespellings(t, tr, p[0], p[1], l, which), got)

					for _, r := range got {
						assert.Equal(t, dist, tr.MustDistance(r.From, r.To))
						assert.True(t, r.Interval.Equal(interval.MustBetweenPitches(tr, r.From, r.To)))
					}
				}
			}
		}
	})

	t.Run("the original pair outside the limit is not listed", func(t *testing.T) {
		t.Parallel()

		got := collect(interval.MustRespellings(tr, c4, pt(spelling.DDoubleSharp(), 4), alteration.Sharp(), interval.ReplaceTo))
		require.Len(t, got, 2)

		for _, r := range got {
			assert.False(t, r.To.Equal(pt(spelling.DDoubleSharp(), 4)))
		}
	})

	t.Run("24: quarter tones", func(t *testing.T) {
		t.Parallel()

		s24 := qt24()
		x, y := pt(spelling.C(), 4), pt(eHalfFlat(), 4)
		got := collect(interval.MustRespellings(s24, x, y, alteration.Sharp(), interval.ReplaceTo))
		assertRespellings(t, expectedRespellings(t, s24, x, y, alteration.Sharp(), interval.ReplaceTo), got)
	})
}

func TestPositionalRespellingsErrors(t *testing.T) {
	t.Parallel()

	tr := trad()
	c4, e4 := pt(spelling.C(), 4), pt(spelling.E(), 4)
	fMinus := pt(spelling.MustNew(letter.F, ai(-mag)), 4)

	tests := []struct {
		name   string
		sys    *system.System
		x, y   pitch.Pitch
		limit  alteration.Alteration
		which  interval.Replace
		want   error
		causes []error
	}{
		{"nil system", nil, c4, e4, ai(1), interval.ReplaceTo, interval.ErrUnset, nil},
		{"unset system", new(system.System), c4, e4, ai(1), interval.ReplaceTo, interval.ErrUnset, nil},
		{"unset x", tr, pitch.Pitch{}, e4, ai(1), interval.ReplaceTo, interval.ErrUnset, nil},
		{"unset y", tr, c4, pitch.Pitch{}, ai(1), interval.ReplaceTo, interval.ErrUnset, nil},
		{"unset choice", tr, c4, e4, ai(1), interval.ReplaceUnset, interval.ErrUnset, nil},
		{"choice out of the enumeration", tr, c4, e4, ai(1), interval.Replace(4), interval.ErrUnset, nil},
		{"unset choice before a negative limit", tr, c4, e4, ai(-1), interval.ReplaceUnset, interval.ErrUnset, nil},
		{"negative limit", tr, c4, e4, ai(-1), interval.ReplaceBoth, interval.ErrInvalidArgument, nil},
		{"negative limit -1/2", tr, c4, e4, alt(-1, 2), interval.ReplaceFrom, interval.ErrInvalidArgument, nil},
		{"negative limit before an inadmissible name", tr, pt(eHalfFlat(), 4), e4, ai(-1), interval.ReplaceTo, interval.ErrInvalidArgument, nil},
		{"inadmissible x", tr, pt(eHalfFlat(), 4), e4, ai(1), interval.ReplaceTo, nil, []error{system.ErrInadmissible}},
		{"inadmissible y", tr, c4, pt(eHalfFlat(), 4), ai(1), interval.ReplaceFrom, nil, []error{system.ErrInadmissible}},
		{"interval of a pair out of range", tr, fMinus, pt(spelling.B(), 4), ai(0), interval.ReplaceTo, interval.ErrOutOfRange, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seq, err := interval.Respellings(tt.sys, tt.x, tt.y, tt.limit, tt.which)
			assertErr(t, err, tt.want, tt.causes...)
			assert.Nil(t, seq)
		})
	}

	t.Run("the pair out of range is not listed when its start is replaced", func(t *testing.T) {
		t.Parallel()

		seq, err := interval.Respellings(tr, fMinus, pt(spelling.B(), 4), ai(1), interval.ReplaceFrom)
		require.NoError(t, err)

		for _, r := range collect(seq) {
			assert.True(t, r.Interval.IsValid())
			assert.False(t, r.From.Equal(fMinus))
		}
	})
}
