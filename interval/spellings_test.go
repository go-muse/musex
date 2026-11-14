package interval_test

import (
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T-3: intervals between names, both ways, transposition of a name, the line
// of fifths and the named inversion (OP-21...OP-28; P-11, P-12, P-16, P-17;
// table 2.4; KP-1, KP-3, KP-9, KP-10; architecture 4.8).

// ---------------------------------------------------------------------------
// Between and BothWays (OP-21, OP-22, P-11; KP-1).

func TestSpellingsBetween(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		x, y     spelling.Spelling
		up, down string
	}{
		{"C E", spelling.C(), spelling.E(), "M3", "↓m6"},
		{"E C", spelling.E(), spelling.C(), "m6", "↓M3"},
		{"F B", spelling.F(), spelling.B(), "A4", "↓d5"},
		{"B F", spelling.B(), spelling.F(), "d5", "↓A4"},
		{"C C", spelling.C(), spelling.C(), "P1", "P1"},
		{"C C sharp", spelling.C(), spelling.CSharp(), "A1", "↓d8"},
		{"C C flat", spelling.C(), spelling.CFlat(), "d8", "↓A1"},
		{"C D double flat", spelling.C(), spelling.DDoubleFlat(), "d2", "↓A7"},
		{"G sharp A flat", spelling.GSharp(), spelling.AFlat(), "d2", "↓A7"},
		{"A flat G sharp", spelling.AFlat(), spelling.GSharp(), "A7", "↓d2"},
		{"C E half flat", spelling.C(), eHalfFlat(), "n3", "↓n6"},
		{"C C half sharp", spelling.C(), sp(letter.C, 1, 2), "{+1/2}1", "↓{-1/2}8"},
		{"C half sharp C", sp(letter.C, 1, 2), spelling.C(), "{-1/2}8", "↓{+1/2}1"},
		{"B C", spelling.B(), spelling.C(), "m2", "↓M7"},
		{"C B", spelling.C(), spelling.B(), "M7", "↓m2"},
		{"D double sharp E flat", spelling.DDoubleSharp(), spelling.EFlat(), "dd2", "↓AA7"},
		{"B sharp F flat", spelling.BSharp(), spelling.FFlat(), "ddd5", "↓AAA4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			up, err := interval.Between(tt.x, tt.y, interval.Up)
			require.NoError(t, err)
			assert.True(t, up.Equal(iv(tt.up)), "up: %v", up)

			down, err := interval.Between(tt.x, tt.y, interval.Down)
			require.NoError(t, err)
			assert.True(t, down.Equal(iv(tt.down)), "down: %v", down)

			bu, bd, err := interval.BothWays(tt.x, tt.y)
			require.NoError(t, err)
			assert.True(t, bu.Equal(up), "BothWays up: %v", bu)
			assert.True(t, bd.Equal(down), "BothWays down: %v", bd)
		})
	}
}

// TestSpellingsNatural covers table 2.4: the interval up from a natural name
// by r letters.
func TestSpellingsNatural(t *testing.T) {
	t.Parallel()

	table := map[letter.Letter][7]string{
		letter.C: {"P1", "M2", "M3", "P4", "P5", "M6", "M7"},
		letter.D: {"P1", "M2", "m3", "P4", "P5", "M6", "m7"},
		letter.E: {"P1", "m2", "m3", "P4", "P5", "m6", "m7"},
		letter.F: {"P1", "M2", "M3", "A4", "P5", "M6", "M7"},
		letter.G: {"P1", "M2", "M3", "P4", "P5", "M6", "m7"},
		letter.A: {"P1", "M2", "m3", "P4", "P5", "m6", "m7"},
		letter.B: {"P1", "m2", "m3", "P4", "d5", "m6", "m7"},
	}

	for l, row := range table {
		for r, want := range row {
			x := spelling.MustNew(l, alteration.Natural())
			y := spelling.MustNew(l.Shift(r), alteration.Natural())

			got, err := interval.Between(x, y, interval.Up)
			require.NoError(t, err)
			assert.True(t, got.Equal(iv(want)), "%v up to %v: %v, want %s", x, y, got, want)

			// The same interval from x transposes x to y.
			name, err := interval.Transpose(x, iv(want))
			require.NoError(t, err)
			assert.True(t, name.Equal(y), "%v by %s: %v", x, want, name)
		}
	}
}

func TestSpellingsBetweenErrors(t *testing.T) {
	t.Parallel()

	cMinus := spelling.MustNew(letter.C, ai(-mag))
	cPlus := spelling.MustNew(letter.C, ai(mag))
	fMinus := spelling.MustNew(letter.F, ai(-mag))

	tests := []struct {
		name string
		x, y spelling.Spelling
		d    interval.Direction
		want error
	}{
		{"unset x", spelling.Spelling{}, spelling.C(), interval.Up, interval.ErrUnset},
		{"unset y", spelling.C(), spelling.Spelling{}, interval.Down, interval.ErrUnset},
		{"unset name before unset direction", spelling.Spelling{}, spelling.C(), interval.DirectionUnset, interval.ErrUnset},
		{"unset direction", spelling.C(), spelling.E(), interval.DirectionUnset, interval.ErrNoDirection},
		{"direction out of the enumeration", spelling.C(), spelling.E(), interval.Direction(7), interval.ErrNoDirection},
		{"unset direction for equal names", spelling.C(), spelling.C(), interval.DirectionUnset, interval.ErrNoDirection},
		{"unset direction before out of range", cMinus, cPlus, interval.DirectionUnset, interval.ErrNoDirection},
		{"unison beyond M", cMinus, cPlus, interval.Up, interval.ErrOutOfRange},
		{"descending unison beyond M", cPlus, cMinus, interval.Down, interval.ErrOutOfRange},
		{"octave beyond -M", cPlus, cMinus, interval.Up, interval.ErrOutOfRange},
		{"augmented fourth beyond M", fMinus, spelling.MustNew(letter.B, ai(mag)), interval.Up, interval.ErrOutOfRange},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := interval.Between(tt.x, tt.y, tt.d)
			assertErr(t, err, tt.want)
			assert.Equal(t, interval.Interval{}, got)
		})
	}

	t.Run("BothWays", func(t *testing.T) {
		t.Parallel()

		for _, p := range [][2]spelling.Spelling{{{}, spelling.C()}, {spelling.C(), {}}} {
			up, down, err := interval.BothWays(p[0], p[1])
			assertErr(t, err, interval.ErrUnset)
			assert.Equal(t, interval.Interval{}, up)
			assert.Equal(t, interval.Interval{}, down)
		}

		up, down, err := interval.BothWays(cMinus, cPlus)
		assertErr(t, err, interval.ErrOutOfRange)
		assert.Equal(t, interval.Interval{}, up)
		assert.Equal(t, interval.Interval{}, down)
	})

	t.Run("in range at the edge", func(t *testing.T) {
		t.Parallel()

		got, err := interval.Between(spelling.C(), cPlus, interval.Up)
		require.NoError(t, err)
		assert.True(t, got.Equal(aug(mag, 1)))

		got, err = interval.Between(cPlus, spelling.C(), interval.Up)
		require.NoError(t, err)
		assert.True(t, got.Equal(dim(mag, 8)))

		got, err = interval.Between(spelling.F(), spelling.MustNew(letter.B, ai(mag-1)), interval.Up)
		require.NoError(t, err)
		assert.True(t, got.Equal(aug(mag, 4)), "A4 with deviation M")
	})
}

// ---------------------------------------------------------------------------
// Transpose (OP-23, P-12; KP-3).

func TestSpellingsTranspose(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		x    spelling.Spelling
		i    string
		want spelling.Spelling
	}{
		{"E by M3", spelling.E(), "M3", spelling.GSharp()},
		{"F by down A4", spelling.F(), "↓A4", spelling.CFlat()},
		{"A by m6", spelling.A(), "m6", spelling.F()},
		{"B sharp by A6", spelling.BSharp(), "A6", sp(letter.G, 3, 1)},
		{"C by A1", spelling.C(), "A1", spelling.CSharp()},
		{"C by down A1", spelling.C(), "↓A1", spelling.CFlat()},
		{"C by d2", spelling.C(), "d2", spelling.DDoubleFlat()},
		{"C by dd2", spelling.C(), "dd2", sp(letter.D, -3, 1)},
		{"E by M10", spelling.E(), "M10", spelling.GSharp()},
		{"C by down m2", spelling.C(), "↓m2", spelling.B()},
		{"G flat by down d5", spelling.GFlat(), "↓d5", spelling.C()},
		{"B flat by down M7", spelling.BFlat(), "↓M7", spelling.CFlat()},
		{"C by n3", spelling.C(), "n3", eHalfFlat()},
		{"C by P1", spelling.C(), "P1", spelling.C()},
		{"C by P8", spelling.C(), "P8", spelling.C()},
		{"C by d8", spelling.C(), "d8", spelling.CFlat()},
		{"C by down d8", spelling.C(), "↓d8", spelling.CSharp()},
		{"C by A15", spelling.C(), "A15", spelling.CSharp()},
		{"C by down M3", spelling.C(), "↓M3", spelling.AFlat()},
		{"C by unison +1/2", spelling.C(), "{+1/2}1", sp(letter.C, 1, 2)},
		{"C by M7", spelling.C(), "M7", spelling.B()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := interval.Transpose(tt.x, iv(tt.i))
			require.NoError(t, err)
			assert.True(t, got.Equal(tt.want), "got %v, want %v", got, tt.want)
		})
	}

	t.Run("octaves do not change the name", func(t *testing.T) {
		t.Parallel()

		assert.True(t, interval.MustTranspose(spelling.C(), basicSize(interval.MaxSize)).Equal(spelling.B()))
		assert.True(t, interval.MustTranspose(spelling.C(), basicSize(interval.MaxSize).MustDown()).Equal(spelling.DFlat()))
	})
}

func TestSpellingsTransposeErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		x      spelling.Spelling
		i      interval.Interval
		want   error
		causes []error
	}{
		{"unset name", spelling.Spelling{}, iv("M3"), interval.ErrUnset, nil},
		{"unset interval", spelling.C(), interval.Interval{}, interval.ErrUnset, nil},
		{"alteration beyond M", spelling.MustNew(letter.C, ai(mag)), iv("A1"),
			interval.ErrOutOfRange, []error{alteration.ErrOutOfRange}},
		{"alteration beyond -M", spelling.MustNew(letter.C, ai(-mag)), iv("↓A1"),
			interval.ErrOutOfRange, []error{alteration.ErrOutOfRange}},
		{"alteration beyond M by a large quality", spelling.B(), aug(mag, 5),
			interval.ErrOutOfRange, []error{alteration.ErrOutOfRange}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := interval.Transpose(tt.x, tt.i)
			assertErr(t, err, tt.want, tt.causes...)
			assert.Equal(t, spelling.Spelling{}, got)
		})
	}

	t.Run("in range at the edge", func(t *testing.T) {
		t.Parallel()

		got, err := interval.Transpose(spelling.C(), aug(mag, 4))
		require.NoError(t, err)
		assert.True(t, got.Equal(spelling.MustNew(letter.F, ai(mag))))
	})
}

// ---------------------------------------------------------------------------
// The line of fifths (OP-24...OP-27, P-16; table 2.5; KP-9).

func TestSpellingsFifthsOf(t *testing.T) {
	t.Parallel()

	tests := []struct {
		x    spelling.Spelling
		want int
	}{
		{spelling.F(), -1},
		{spelling.C(), 0},
		{spelling.G(), 1},
		{spelling.D(), 2},
		{spelling.A(), 3},
		{spelling.E(), 4},
		{spelling.B(), 5},
		{spelling.FSharp(), 6},
		{spelling.BFlat(), -2},
		{spelling.CSharp(), 7},
		{spelling.FFlat(), -8},
		{spelling.CDoubleSharp(), 14},
		{spelling.FDoubleFlat(), -15},
		{spelling.MustNew(letter.C, ai(mag)), 7 * mag},
		{spelling.MustNew(letter.F, ai(-mag)), -1 - 7*mag},
	}

	for _, tt := range tests {
		t.Run(tt.x.String(), func(t *testing.T) {
			t.Parallel()

			got, err := interval.FifthsOf(tt.x)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}

	t.Run("errors", func(t *testing.T) {
		t.Parallel()

		got, err := interval.FifthsOf(spelling.Spelling{})
		assertErr(t, err, interval.ErrUnset)
		assert.Zero(t, got)

		got, err = interval.FifthsOf(eHalfFlat())
		assertErr(t, err, interval.ErrNotOnFifths)
		assert.Zero(t, got)
	})
}

func TestSpellingsFifthsBetween(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		x, y spelling.Spelling
		want int
		up   string
	}{
		{"C E", spelling.C(), spelling.E(), 4, "M3"},
		{"F B", spelling.F(), spelling.B(), 6, "A4"},
		{"B F", spelling.B(), spelling.F(), -6, "d5"},
		{"C C sharp", spelling.C(), spelling.CSharp(), 7, "A1"},
		{"C D flat", spelling.C(), spelling.DFlat(), -5, "m2"},
		{"C C flat", spelling.C(), spelling.CFlat(), -7, "d8"},
		{"C C", spelling.C(), spelling.C(), 0, "P1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := interval.FifthsBetween(tt.x, tt.y)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)

			i, err := interval.FromFifths(got)
			require.NoError(t, err)
			assert.True(t, i.Equal(iv(tt.up)), "FromFifths(%d) = %v", got, i)
		})
	}

	t.Run("errors", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			x, y spelling.Spelling
			want error
		}{
			{spelling.Spelling{}, spelling.C(), interval.ErrUnset},
			{spelling.C(), spelling.Spelling{}, interval.ErrUnset},
			{eHalfFlat(), spelling.Spelling{}, interval.ErrUnset},
			{spelling.C(), eHalfFlat(), interval.ErrNotOnFifths},
			{eHalfFlat(), spelling.C(), interval.ErrNotOnFifths},
		}

		for _, c := range cases {
			got, err := interval.FifthsBetween(c.x, c.y)
			assertErr(t, err, c.want)
			assert.Zero(t, got)
		}
	})
}

func TestSpellingsFromFifths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		q    int
		want interval.Interval
	}{
		{0, iv("P1")},
		{-5, iv("m2")},
		{2, iv("M2")},
		{-3, iv("m3")},
		{4, iv("M3")},
		{-1, iv("P4")},
		{6, iv("A4")},
		{-6, iv("d5")},
		{1, iv("P5")},
		{-4, iv("m6")},
		{3, iv("M6")},
		{-2, iv("m7")},
		{5, iv("M7")},
		{7, iv("A1")},
		{-7, iv("d8")},
		{9, iv("A2")},
		{11, iv("A3")},
		{14, iv("AA1")},
		{-14, iv("dd8")},
		{-13, iv("dd5")},
		{-8, iv("d4")},
		{7 * mag, aug(mag, 1)},
		{-7 * mag, dim(mag, 8)},
		{7*mag + 5, aug(mag, 7)},
		{-7*mag + 1, dim(mag, 5)},
	}

	for _, tt := range tests {
		t.Run(tt.want.String(), func(t *testing.T) {
			t.Parallel()

			got, err := interval.FromFifths(tt.q)
			require.NoError(t, err)
			assert.True(t, got.Equal(tt.want), "FromFifths(%d) = %v", tt.q, got)
			assert.True(t, got.IsAscending() && got.IsSimple())
		})
	}

	t.Run("out of range", func(t *testing.T) {
		t.Parallel()

		for _, q := range []int{7*mag + 7, 7*mag + 6, -7*mag - 7, -7*mag - 5, maxInt, minInt, maxInt - 1, minInt + 1} {
			got, err := interval.FromFifths(q)
			assertErr(t, err, interval.ErrOutOfRange)
			assert.Equal(t, interval.Interval{}, got, "%d", q)
		}
	})
}

func TestSpellingsIntervalFifths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		i    string
		want int
	}{
		{"P1", 0}, {"m2", -5}, {"M2", 2}, {"m3", -3}, {"M3", 4}, {"P4", -1}, {"A4", 6}, {"d5", -6},
		{"P5", 1}, {"m6", -4}, {"M6", 3}, {"m7", -2}, {"M7", 5}, {"P8", 0}, {"A1", 7}, {"d8", -7},
		{"M10", 4}, {"↓M3", -4}, {"↓P8", 0}, {"↓A1", -7}, {"P15", 0}, {"AA4", 13}, {"↓dd5", 13},
	}

	for _, tt := range tests {
		t.Run(tt.i, func(t *testing.T) {
			t.Parallel()

			got, err := iv(tt.i).Fifths()
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}

	t.Run("edges", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, 7*mag, aug(mag, 1).MustFifths())
		assert.Equal(t, 1-7*mag, dim(mag, 5).MustFifths())
		assert.Equal(t, 5, basicSize(interval.MaxSize).MustFifths())
	})

	t.Run("errors", func(t *testing.T) {
		t.Parallel()

		for _, c := range []struct {
			i    interval.Interval
			want error
		}{
			{interval.Interval{}, interval.ErrUnset},
			{iv("n3"), interval.ErrNotOnFifths},
			{iv("↓n6"), interval.ErrNotOnFifths},
			{iv("{+1/2}1"), interval.ErrNotOnFifths},
		} {
			got, err := c.i.Fifths()
			assertErr(t, err, c.want)
			assert.Zero(t, got)
		}
	})
}

// ---------------------------------------------------------------------------
// Named inversion (OP-28, P-17; KP-10).

func TestSpellingsReflect(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		axis, x spelling.Spelling
		want    spelling.Spelling
	}{
		{"C", spelling.D(), spelling.C(), spelling.E()},
		{"E", spelling.D(), spelling.E(), spelling.C()},
		{"F", spelling.D(), spelling.F(), spelling.B()},
		{"F sharp", spelling.D(), spelling.FSharp(), spelling.BFlat()},
		{"D", spelling.D(), spelling.D(), spelling.D()},
		{"D sharp", spelling.D(), spelling.DSharp(), spelling.DFlat()},
		{"axis C, G", spelling.C(), spelling.G(), spelling.F()},
		{"axis C, E half flat", spelling.C(), eHalfFlat(), sp(letter.A, -1, 2)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := interval.Reflect(tt.axis, tt.x)
			require.NoError(t, err)
			assert.True(t, got.Equal(tt.want), "got %v, want %v", got, tt.want)
		})
	}

	t.Run("errors", func(t *testing.T) {
		t.Parallel()

		for _, c := range []struct {
			axis, x spelling.Spelling
			want    error
		}{
			{spelling.Spelling{}, spelling.C(), interval.ErrUnset},
			{spelling.D(), spelling.Spelling{}, interval.ErrUnset},
			{spelling.MustNew(letter.D, ai(mag)), spelling.D(), interval.ErrOutOfRange},
			{spelling.MustNew(letter.C, ai(-mag)), spelling.MustNew(letter.C, ai(mag)), interval.ErrOutOfRange},
		} {
			got, err := interval.Reflect(c.axis, c.x)
			assertErr(t, err, c.want)
			assert.Equal(t, spelling.Spelling{}, got)
		}
	})
}

// ---------------------------------------------------------------------------
// Properties on the sweep of F-2: 35 predefined names × 57 predefined
// intervals × both directions (architecture 4.8, "Properties").

func TestSpellingsProperties(t *testing.T) {
	t.Parallel()

	names := spelling.Predefined()
	require.Len(t, names, 35)

	intervals := predefinedBoth()

	t.Run("transpose", func(t *testing.T) {
		t.Parallel()

		for _, x := range names {
			assert.True(t, interval.MustTranspose(x, interval.PerfectUnison()).Equal(x), "P1 is the identity")

			qx := interval.MustFifthsOf(x)

			for _, i := range intervals {
				y, err := interval.Transpose(x, i)
				require.NoError(t, err, "%v by %v", x, i)

				back, err := interval.Transpose(y, i.MustOpposite())
				require.NoError(t, err)
				assert.True(t, back.Equal(x), "%v by %v and back: %v", x, i, back)

				got, err := interval.Between(x, y, i.Direction())
				require.NoError(t, err)
				assert.True(t, got.Equal(expectedBetween(i)), "Between(%v, %v, %v) = %v for %v", x, y, i.Direction(), got, i)

				qy, err := interval.FifthsOf(y)
				require.NoError(t, err)
				assert.Equal(t, qx+i.MustFifths(), qy, "P-16: %v by %v", x, i)

				// The letter moves by n − 1 letter steps in the direction.
				steps := (i.Size() - 1) % 7
				if i.IsDescending() {
					steps = -steps
				}

				assert.Equal(t, x.Letter().Shift(steps), y.Letter(), "%v by %v", x, i)
			}
		}
	})

	t.Run("between", func(t *testing.T) {
		t.Parallel()

		for _, x := range names {
			for _, y := range names {
				up, down, err := interval.BothWays(x, y)
				require.NoError(t, err)
				assert.True(t, up.Equal(interval.MustBetween(x, y, interval.Up)))
				assert.True(t, down.Equal(interval.MustBetween(x, y, interval.Down)))
				assert.True(t, up.IsSimple() && down.IsSimple())

				if x.Equal(y) {
					assert.True(t, up.IsPerfectUnison() && down.IsPerfectUnison(), "%v %v", x, y)

					continue
				}

				assert.True(t, up.IsAscending(), "%v %v", x, y)
				assert.True(t, down.IsDescending(), "%v %v", x, y)
				assert.True(t, up.MustInvert().Equal(down.MustUp()), "OP-22: inversions of each other: %v %v", x, y)
				assert.True(t, down.Equal(interval.MustBetween(y, x, interval.Up).MustOpposite()), "%v %v", x, y)
				assert.True(t, interval.MustTranspose(x, up).Equal(y), "%v %v", x, y)
				assert.True(t, interval.MustTranspose(x, down).Equal(y), "%v %v", x, y)

				f, err := interval.FromFifths(interval.MustFifthsBetween(x, y))
				require.NoError(t, err)
				assert.True(t, f.Equal(up), "OP-26: %v %v", x, y)
				assert.Equal(t, interval.MustFifthsOf(y)-interval.MustFifthsOf(x), interval.MustFifthsBetween(x, y))
			}
		}
	})

	t.Run("reflect", func(t *testing.T) {
		t.Parallel()

		for _, a := range names {
			assert.True(t, interval.MustReflect(a, a).Equal(a), "the axis is fixed: %v", a)

			for _, x := range names {
				y, err := interval.Reflect(a, x)
				require.NoError(t, err)
				assert.True(t, interval.MustReflect(a, y).Equal(x), "involution: %v %v", a, x)

				upToX := interval.MustBetween(a, x, interval.Up)
				downToY := interval.MustBetween(a, y, interval.Down)
				assert.True(t, downToY.EqualUndirected(upToX), "OP-28: a %v, x %v, y %v", a, x, y)
			}
		}
	})
}

// expectedBetween is the interval between a name and its transposition by i
// (architecture 4.8, "Properties"): the simple part of i, except a simple
// part of size 8 with d ≥ 0, which names without octaves do not tell from a
// unison of the same deviation.
func expectedBetween(i interval.Interval) interval.Interval {
	s := i.MustSimple()
	if s.Size() == 8 && s.Deviation().Sign() >= 0 {
		return interval.MustFromDeviation(1, s.Deviation()).MustWithDirection(i.Direction())
	}

	return s
}

// TestSpellingsSum checks P-9 for names: the transposition by a sum is the
// transposition by the parts in turn.
func TestSpellingsSum(t *testing.T) {
	t.Parallel()

	intervals := predefinedBoth()
	names := []spelling.Spelling{spelling.C(), spelling.FSharp(), spelling.BFlat(), spelling.EDoubleSharp(), eHalfFlat()}

	for _, x := range names {
		for _, i := range intervals {
			xi := interval.MustTranspose(x, i)

			for _, j := range intervals {
				s, err := i.Add(j)
				require.NoError(t, err)

				got, err := interval.Transpose(x, s)
				require.NoError(t, err)
				assert.True(t, got.Equal(interval.MustTranspose(xi, j)), "%v by %v + %v", x, i, j)
			}
		}
	}
}
