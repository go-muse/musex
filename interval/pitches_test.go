package interval_test

import (
	"errors"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T-4: intervals between written pitches, transposition and named inversion
// of a pitch in an explicit system (OP-29...OP-31; P-13, P-14, P-17; KP-2,
// KP-4, KP-10; architecture 4.9).

// ---------------------------------------------------------------------------
// BetweenPitches (OP-29, P-13; KP-2).

func TestPitchesBetween(t *testing.T) {
	t.Parallel()

	tr, ba, s19 := trad(), boundaryA(), pos19()

	tests := []struct {
		name  string
		sys   *system.System
		x, y  pitch.Pitch
		want  string
		steps system.Step // sys.Distance(x, y), the positions
	}{
		{"C4 E5", tr, pt(spelling.C(), 4), pt(spelling.E(), 5), "M10", 16},
		{"E5 C4", tr, pt(spelling.E(), 5), pt(spelling.C(), 4), "↓M10", -16},
		{"B3 C4", tr, pt(spelling.B(), 3), pt(spelling.C(), 4), "m2", 1},
		{"B sharp 3 C4: same position, up by writing", tr, wp(tr, spelling.BSharp(), 3), pt(spelling.C(), 4), "d2", 0},
		{"A3 C3", tr, pt(spelling.A(), 3), pt(spelling.C(), 3), "↓M6", -9},
		{"C4 C4", tr, pt(spelling.C(), 4), pt(spelling.C(), 4), "P1", 0},
		{"C4 C sharp 4", tr, pt(spelling.C(), 4), pt(spelling.CSharp(), 4), "A1", 1},
		{"C4 C flat 4", tr, pt(spelling.C(), 4), wp(tr, spelling.CFlat(), 4), "↓A1", -1},
		{"C4 C5", tr, pt(spelling.C(), 4), pt(spelling.C(), 5), "P8", 12},
		{"C4 C6", tr, pt(spelling.C(), 4), pt(spelling.C(), 6), "P15", 24},
		{"C4 C flat 5", tr, pt(spelling.C(), 4), wp(tr, spelling.CFlat(), 5), "d8", 11},
		{"C4 C sharp 5", tr, pt(spelling.C(), 4), pt(spelling.CSharp(), 5), "A8", 13},
		{"C4 D double flat 4: same position", tr, pt(spelling.C(), 4), pt(spelling.DDoubleFlat(), 4), "d2", 0},
		{"F4 B5", tr, pt(spelling.F(), 4), pt(spelling.B(), 5), "A11", 18},
		{"C4 E half flat 4 in 24", qt24(), pt(spelling.C(), 4), pt(eHalfFlat(), 4), "n3", 7},
		{"boundary A: A3 C3", ba, wp(ba, spelling.A(), 3), wp(ba, spelling.C(), 3), "m3", 3},
		{"boundary A: G3 A4", ba, wp(ba, spelling.G(), 3), wp(ba, spelling.A(), 4), "M2", 2},
		{"19: B sharp 3 C4", s19, wp(s19, spelling.BSharp(), 3), wp(s19, spelling.C(), 4), "d2", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := interval.BetweenPitches(tt.sys, tt.x, tt.y)
			require.NoError(t, err)
			assert.True(t, got.Equal(iv(tt.want)), "got %v", got)

			back, err := interval.BetweenPitches(tt.sys, tt.y, tt.x)
			require.NoError(t, err)
			assert.True(t, back.Equal(got.MustOpposite()), "the other way: %v", back)

			dist, err := tt.sys.Distance(tt.x, tt.y)
			require.NoError(t, err)
			assert.Equal(t, tt.steps, dist, "the positions of the example")
		})
	}

	t.Run("written octaves of B sharp and C flat", func(t *testing.T) {
		t.Parallel()

		// B♯3 is (B♯, Octave 4) and C♭4 is (C♭, Octave 3) in the
		// traditional system; in the 19-position system B♯3 has Octave 3.
		assert.True(t, wp(tr, spelling.BSharp(), 3).Equal(pt(spelling.BSharp(), 4)))
		assert.True(t, wp(tr, spelling.CFlat(), 4).Equal(pt(spelling.CFlat(), 3)))
		assert.True(t, wp(s19, spelling.BSharp(), 3).Equal(pt(spelling.BSharp(), 3)))
	})

	t.Run("edges of the size", func(t *testing.T) {
		t.Parallel()

		far := pitch.MustNew(spelling.B(), octave.FromASPN(q7-1))
		got, err := interval.BetweenPitches(tr, pt(spelling.C(), 0), far)
		require.NoError(t, err)
		assert.True(t, got.Equal(basicSize(interval.MaxSize)), "C0 up to B(MaxInt/7-1)")

		got, err = interval.BetweenPitches(tr, far, pt(spelling.C(), 0))
		require.NoError(t, err)
		assert.True(t, got.Equal(basicSize(interval.MaxSize).MustDown()))
	})
}

func TestPitchesBetweenErrors(t *testing.T) {
	t.Parallel()

	tr := trad()
	c4, e4 := pt(spelling.C(), 4), pt(spelling.E(), 4)
	eh4 := pt(eHalfFlat(), 4)

	tests := []struct {
		name   string
		sys    *system.System
		x, y   pitch.Pitch
		want   error
		causes []error
	}{
		{"nil system", nil, c4, e4, interval.ErrUnset, nil},
		{"unset system", new(system.System), c4, e4, interval.ErrUnset, nil},
		{"unset x", tr, pitch.Pitch{}, e4, interval.ErrUnset, nil},
		{"unset y", tr, c4, pitch.Pitch{}, interval.ErrUnset, nil},
		{"nil system before an inadmissible name", nil, eh4, e4, interval.ErrUnset, nil},
		{"unset y before an inadmissible x", tr, eh4, pitch.Pitch{}, interval.ErrUnset, nil},
		{"inadmissible x", tr, eh4, e4, nil, []error{system.ErrInadmissible}},
		{"inadmissible y", tr, c4, eh4, nil, []error{system.ErrInadmissible}},
		{"written octave out of int", tr, pitch.MustNew(spelling.BSharp(), octave.FromASPN(minInt)), c4,
			interval.ErrOutOfRange, []error{system.ErrOutOfRange}},
		{"difference of written octaves out of int", tr, pitch.MustNew(spelling.C(), octave.FromASPN(minInt)),
			pitch.MustNew(spelling.C(), octave.FromASPN(maxInt)), interval.ErrOutOfRange, []error{octave.ErrOutOfRange}},
		{"size beyond MaxSize", tr, pt(spelling.B(), 0), pitch.MustNew(spelling.B(), octave.FromASPN(q7)),
			interval.ErrOutOfRange, nil},
		{"descending size beyond MaxSize", tr, pitch.MustNew(spelling.B(), octave.FromASPN(q7)), pt(spelling.B(), 0),
			interval.ErrOutOfRange, nil},
		{"deviation beyond M", tr, pt(spelling.MustNew(letter.C, ai(-mag)), 4), pt(spelling.MustNew(letter.C, ai(mag)), 4),
			interval.ErrOutOfRange, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := interval.BetweenPitches(tt.sys, tt.x, tt.y)
			assertErr(t, err, tt.want, tt.causes...)
			assert.Equal(t, interval.Interval{}, got)
		})
	}
}

// ---------------------------------------------------------------------------
// TransposePitch (OP-30, P-14; KP-4).

func TestPitchesTranspose(t *testing.T) {
	t.Parallel()

	tr, ba, s19 := trad(), boundaryA(), pos19()

	tests := []struct {
		name  string
		sys   *system.System
		x     pitch.Pitch
		i     string
		want  pitch.Pitch
		steps system.Step
	}{
		{"C4 by M10", tr, pt(spelling.C(), 4), "M10", pt(spelling.E(), 5), 16},
		{"B3 by m2", tr, pt(spelling.B(), 3), "m2", pt(spelling.C(), 4), 1},
		{"C4 by down P8", tr, pt(spelling.C(), 4), "↓P8", pt(spelling.C(), 3), -12},
		{"B sharp 3 by d2", tr, wp(tr, spelling.BSharp(), 3), "d2", pt(spelling.C(), 4), 0},
		{"C flat 4 by A1", tr, wp(tr, spelling.CFlat(), 4), "A1", pt(spelling.C(), 4), 1},
		{"B3 by A1", tr, pt(spelling.B(), 3), "A1", wp(tr, spelling.BSharp(), 3), 1},
		{"C4 by down d2", tr, pt(spelling.C(), 4), "↓d2", wp(tr, spelling.BSharp(), 3), 0},
		{"C4 by down A1", tr, pt(spelling.C(), 4), "↓A1", wp(tr, spelling.CFlat(), 4), -1},
		{"C4 by down m2", tr, pt(spelling.C(), 4), "↓m2", pt(spelling.B(), 3), -1},
		{"C4 by down M7", tr, pt(spelling.C(), 4), "↓M7", pt(spelling.DFlat(), 3), -11},
		{"A4 by M3", tr, pt(spelling.A(), 4), "M3", pt(spelling.CSharp(), 5), 4},
		{"C4 by P15", tr, pt(spelling.C(), 4), "P15", pt(spelling.C(), 6), 24},
		{"C4 by P1", tr, pt(spelling.C(), 4), "P1", pt(spelling.C(), 4), 0},
		{"C4 by dd2", tr, pt(spelling.C(), 4), "dd2", wp(tr, sp(letter.D, -3, 1), 4), -1},
		{"C4 by n3 in 24", qt24(), pt(spelling.C(), 4), "n3", pt(eHalfFlat(), 4), 7},
		{"boundary A: A3 by m3", ba, wp(ba, spelling.A(), 3), "m3", wp(ba, spelling.C(), 3), 3},
		{"boundary A: G3 by M2", ba, wp(ba, spelling.G(), 3), "M2", wp(ba, spelling.A(), 4), 2},
		{"19: B sharp 3 by m2", s19, wp(s19, spelling.BSharp(), 3), "m2", wp(s19, spelling.CSharp(), 4), 2},
		{"19: C4 by A4", s19, wp(s19, spelling.C(), 4), "A4", wp(s19, spelling.FSharp(), 4), 9},
		{"19: C4 by d5", s19, wp(s19, spelling.C(), 4), "d5", wp(s19, spelling.GFlat(), 4), 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := interval.TransposePitch(tt.sys, tt.x, iv(tt.i))
			require.NoError(t, err)
			assert.True(t, got.Equal(tt.want), "got %v, want %v", got, tt.want)

			dist, err := tt.sys.Distance(tt.x, got)
			require.NoError(t, err)
			assert.Equal(t, tt.steps, dist, "the position moves by the steps of the interval")
		})
	}
}

func TestPitchesTransposeErrors(t *testing.T) {
	t.Parallel()

	tr := trad()
	c4 := pt(spelling.C(), 4)
	cPlus4 := pt(spelling.MustNew(letter.C, ai(mag)), 4)

	tests := []struct {
		name    string
		sys     *system.System
		x       pitch.Pitch
		i       interval.Interval
		want    error
		causes  []error
		anyOf   []error // at least one of these causes
		notWant []error
	}{
		{name: "nil system", sys: nil, x: c4, i: iv("M3"), want: interval.ErrUnset},
		{name: "unset system", sys: new(system.System), x: c4, i: iv("M3"), want: interval.ErrUnset},
		{name: "unset pitch", sys: tr, x: pitch.Pitch{}, i: iv("M3"), want: interval.ErrUnset},
		{name: "unset interval", sys: tr, x: c4, i: interval.Interval{}, want: interval.ErrUnset},
		{name: "unset interval before an inadmissible pitch", sys: tr, x: pt(eHalfFlat(), 4), i: interval.Interval{}, want: interval.ErrUnset},
		{name: "inadmissible pitch", sys: tr, x: pt(eHalfFlat(), 4), i: iv("M3"), causes: []error{system.ErrInadmissible}},
		{name: "alteration of the result beyond M", sys: tr, x: cPlus4, i: iv("A1"),
			want: interval.ErrOutOfRange, causes: []error{alteration.ErrOutOfRange}},
		{name: "alteration out of range before inadmissibility of the result", sys: tr, x: cPlus4, i: iv("{+1/2}1"),
			want: interval.ErrOutOfRange, causes: []error{alteration.ErrOutOfRange}, notWant: []error{system.ErrInadmissible}},
		{name: "inadmissible result", sys: tr, x: c4, i: iv("n3"), causes: []error{system.ErrInadmissible}},
		{name: "inadmissible result in 31", sys: sf31(), x: c4, i: iv("n3"), causes: []error{system.ErrInadmissible}},
		{name: "octave of the result beyond int", sys: tr, x: pitch.MustNew(spelling.C(), octave.FromASPN(maxInt)), i: iv("P8"),
			want: interval.ErrOutOfRange, anyOf: []error{octave.ErrOutOfRange, system.ErrOutOfRange}},
		{name: "octave of the result below int", sys: tr, x: pitch.MustNew(spelling.C(), octave.FromASPN(minInt)), i: iv("↓m2"),
			want: interval.ErrOutOfRange, anyOf: []error{octave.ErrOutOfRange, system.ErrOutOfRange}},
		{name: "written octave of the pitch beyond int", sys: tr, x: pitch.MustNew(spelling.BSharp(), octave.FromASPN(minInt)), i: iv("P1"),
			want: interval.ErrOutOfRange, anyOf: []error{octave.ErrOutOfRange, system.ErrOutOfRange}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := interval.TransposePitch(tt.sys, tt.x, tt.i)
			assertErr(t, err, tt.want, tt.causes...)
			assert.Equal(t, pitch.Pitch{}, got)

			if len(tt.anyOf) > 0 {
				assert.True(t, errors.Is(err, tt.anyOf[0]) || errors.Is(err, tt.anyOf[1]), "cause of %v", err)
			}

			for _, e := range tt.notWant {
				assert.NotErrorIs(t, err, e)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ReflectPitch (OP-31, P-17; KP-10).

func TestPitchesReflect(t *testing.T) {
	t.Parallel()

	tr := trad()
	axis := pt(spelling.D(), 4)

	tests := []struct {
		name string
		x    pitch.Pitch
		want pitch.Pitch
	}{
		{"C4", pt(spelling.C(), 4), pt(spelling.E(), 4)},
		{"E5", pt(spelling.E(), 5), pt(spelling.C(), 3)},
		{"F sharp 4", pt(spelling.FSharp(), 4), pt(spelling.BFlat(), 3)},
		{"B3", pt(spelling.B(), 3), pt(spelling.F(), 4)},
		{"D4", pt(spelling.D(), 4), pt(spelling.D(), 4)},
		{"D sharp 4", pt(spelling.DSharp(), 4), pt(spelling.DFlat(), 4)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := interval.ReflectPitch(tr, axis, tt.x)
			require.NoError(t, err)
			assert.True(t, got.Equal(tt.want), "got %v, want %v", got, tt.want)
		})
	}

	t.Run("errors", func(t *testing.T) {
		t.Parallel()

		dPlus := pt(spelling.MustNew(letter.D, ai(mag)), 4)

		cases := []struct {
			name   string
			sys    *system.System
			a, x   pitch.Pitch
			want   error
			causes []error
		}{
			{"nil system", nil, axis, pt(spelling.C(), 4), interval.ErrUnset, nil},
			{"unset axis", tr, pitch.Pitch{}, pt(spelling.C(), 4), interval.ErrUnset, nil},
			{"unset pitch", tr, axis, pitch.Pitch{}, interval.ErrUnset, nil},
			{"inadmissible axis", tr, pt(eHalfFlat(), 4), pt(spelling.C(), 4), nil, []error{system.ErrInadmissible}},
			{"inadmissible pitch", tr, axis, pt(eHalfFlat(), 4), nil, []error{system.ErrInadmissible}},
			{"alteration beyond M", tr, dPlus, pt(spelling.D(), 4), interval.ErrOutOfRange, nil},
		}

		for _, c := range cases {
			got, err := interval.ReflectPitch(c.sys, c.a, c.x)
			assertErr(t, err, c.want, c.causes...)
			assert.Equal(t, pitch.Pitch{}, got, c.name)
		}
	})
}

// ---------------------------------------------------------------------------
// Properties on the sweep of F-2: written octaves 2...5 in the traditional,
// 19-position and 24-quarter-tone systems (P-13, P-14, P-17; architecture
// 4.9, "Properties").

func TestPitchesProperties(t *testing.T) {
	t.Parallel()

	intervals := predefinedBoth()

	for _, sys := range []*system.System{trad(), pos19(), qt24()} {
		t.Run(sys.Name(), func(t *testing.T) {
			t.Parallel()

			var pitches []pitch.Pitch

			for _, name := range spelling.Predefined() {
				for w := 2; w <= 5; w++ {
					pitches = append(pitches, wp(sys, name, w))
				}
			}

			for _, x := range pitches {
				for _, i := range intervals {
					y, err := interval.TransposePitch(sys, x, i)
					require.NoError(t, err, "%v by %v", x, i)

					assert.True(t, y.Spelling().Equal(interval.MustTranspose(x.Spelling(), i)), "the name: %v by %v", x, i)

					got, err := interval.BetweenPitches(sys, x, y)
					require.NoError(t, err)
					assert.True(t, got.Equal(i), "BetweenPitches(%v, %v) = %v, want %v", x, y, got, i)

					back, err := interval.TransposePitch(sys, y, i.MustOpposite())
					require.NoError(t, err)
					assert.True(t, back.Equal(x), "%v by %v and back: %v", x, i, back)

					dist, err := sys.Distance(x, y)
					require.NoError(t, err)
					assert.Equal(t, i.MustSignedSteps(sys, x.Spelling()), dist, "%v by %v", x, i)
				}

				assert.True(t, interval.MustTransposePitch(sys, x, interval.PerfectUnison()).Equal(x), "P1 is the identity")

				for k := 0; k <= 3; k++ {
					shifted, err := x.Shift(k)
					require.NoError(t, err)
					assert.True(t, interval.MustTransposePitch(sys, x, interval.PerfectOctave().MustMul(k)).Equal(shifted),
						"%d octaves of %v", k, x)
				}
			}

			// Pairs and the named inversion on a part of the sweep.
			axes := pitches
			for k, a := range axes {
				if k%5 != 0 {
					continue
				}

				assert.True(t, interval.MustReflectPitch(sys, a, a).Equal(a), "the axis is fixed")

				for _, x := range pitches {
					ax := interval.MustBetweenPitches(sys, a, x)
					assert.True(t, interval.MustBetweenPitches(sys, x, a).Equal(ax.MustOpposite()), "%v %v", a, x)

					y, err := interval.ReflectPitch(sys, a, x)
					require.NoError(t, err)
					assert.True(t, interval.MustReflectPitch(sys, a, y).Equal(x), "involution: %v %v", a, x)
					assert.True(t, interval.MustBetweenPitches(sys, a, y).Equal(ax.MustOpposite()), "P-17: %v %v", a, x)
				}
			}
		})
	}
}
