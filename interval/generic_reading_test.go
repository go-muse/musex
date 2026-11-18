package interval_test

import (
	"fmt"
	"math"
	"slices"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenericValue(t *testing.T) {
	t.Parallel()
	var zero interval.Generic
	assert.False(t, zero.IsValid())
	assert.Zero(t, zero.Size())
	assert.Equal(t, "Generic(unset)", zero.String())
	assert.Equal(t, zero, interval.Interval{}.Generic())
	for _, size := range []int{1, 3, 8, 10, 15, 16, interval.MaxSize} {
		g, err := interval.NewGeneric(size)
		require.NoError(t, err)
		assert.True(t, g.IsValid())
		assert.Equal(t, size, g.Size())
		assert.Equal(t, fmt.Sprint(size), g.String())
		assert.Equal(t, g, interval.MustNewGeneric(size))
		assert.Equal(t, g, basicSize(size).Generic())
		assert.Equal(t, g, basicSize(size).MustDown().Generic())
		keys := map[interval.Generic]int{g: size}
		assert.Equal(t, size, keys[g])
	}
	for _, size := range []int{0, -1, math.MinInt} {
		g, err := interval.NewGeneric(size)
		assert.ErrorIs(t, err, interval.ErrInvalidArgument)
		assert.Equal(t, zero, g)
	}
	assert.ErrorIs(t, panicError(t, func() { interval.MustNewGeneric(0) }), interval.ErrInvalidArgument)
}

func TestGenericPalettes(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		size int
		dir  interval.Direction
		p    interval.Palette
		want []interval.Interval
	}{
		{"default third", 3, interval.Up, interval.DefaultPalette(), ivs("d3", "m3", "M3", "A3")},
		{"zero palette", 3, interval.Up, interval.Palette{}, ivs("d3", "m3", "M3", "A3")},
		{"default tenth", 10, interval.Down, interval.DefaultPalette(), ivs("↓d10", "↓m10", "↓M10", "↓A10")},
		{"default fifth", 5, interval.Up, interval.DefaultPalette(), ivs("dd5", "d5", "P5", "A5", "AA5")},
		{"double octave", 15, interval.Down, interval.DefaultPalette(), ivs("↓dd15", "↓d15", "↓P15", "↓A15", "↓AA15")},
		{"unison upward", 1, interval.Up, interval.DefaultPalette(), ivs("P1", "A1", "AA1")},
		{"unison downward canonical zero", 1, interval.Down, interval.DefaultPalette(), ivs("P1", "↓A1", "↓AA1")},
		{"basic third", 3, interval.Up, interval.MustThrough(0), ivs("m3", "M3")},
		{"basic fifth", 5, interval.Up, interval.MustThrough(0), ivs("P5")},
		{"double third", 3, interval.Up, interval.MustThrough(2), ivs("dd3", "d3", "m3", "M3", "A3", "AA3")},
		{"explicit rational", 3, interval.Up, interval.Deviations(ai(0), alt(-1, 2), ai(-1), ai(-1)), ivs("m3", "n3", "M3")},
		{"explicit empty", 3, interval.Up, interval.Deviations(), nil},
		{"negative unison omitted", 1, interval.Up, interval.Deviations(ai(-2), ai(-1), alt(-1, 2)), nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g := interval.MustNewGeneric(tc.size)
			seq, err := g.Variants(tc.dir, tc.p)
			require.NoError(t, err)
			require.NotNil(t, seq)
			assert.Equal(t, tc.want, slices.Collect(seq))
			assert.Equal(t, tc.want, slices.Collect(seq), "repeated traversal")
			for range seq {
				break
			}
			assert.Equal(t, tc.want, slices.Collect(seq), "early stop does not consume the sequence")
			assert.Equal(t, tc.want, slices.Collect(g.MustVariants(tc.dir, tc.p)))
		})
	}
	for _, k := range []int{-1, alteration.MaxMagnitude + 1, math.MaxInt} {
		_, err := interval.Through(k)
		assert.ErrorIs(t, err, interval.ErrInvalidArgument)
	}
	_, err := interval.Through(alteration.MaxMagnitude)
	require.NoError(t, err)
	assert.ErrorIs(t, panicError(t, func() { interval.MustThrough(-1) }), interval.ErrInvalidArgument)
	values := []alteration.Alteration{ai(-1), alt(-1, 2)}
	p := interval.Deviations(values...)
	extended := p.WithDeviations(ai(0))
	values[0] = ai(4)
	g := interval.MustNewGeneric(3)
	assert.Equal(t, ivs("m3", "n3"), slices.Collect(g.MustVariants(interval.Up, p)))
	assert.Equal(t, ivs("m3", "n3", "M3"), slices.Collect(g.MustVariants(interval.Up, extended)))
	assert.Equal(t, ivs("d3", "m3", "n3", "M3", "A3"), slices.Collect(g.MustVariants(interval.Up, interval.Palette{}.WithDeviations(alt(-1, 2)))))
	for _, tc := range []struct {
		g    interval.Generic
		dir  interval.Direction
		want error
	}{
		{interval.Generic{}, interval.Up, interval.ErrUnset},
		{g, interval.DirectionUnset, interval.ErrNoDirection},
		{g, interval.Direction(99), interval.ErrInvalidArgument},
	} {
		seq, err := tc.g.Variants(tc.dir, interval.Deviations())
		assert.ErrorIs(t, err, tc.want)
		assert.Nil(t, seq)
	}
	assert.ErrorIs(t, panicError(t, func() { g.MustVariants(interval.DirectionUnset, p) }), interval.ErrNoDirection)
}

func TestGenericFromPitch(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name  string
		size  int
		sys   *system.System
		start pitch.Pitch
		dir   interval.Direction
		p     interval.Palette
		wantI []interval.Interval
		wantP []pitch.Pitch
	}{
		{"third from D", 3, trad(), wp(trad(), spelling.D(), 4), interval.Up, interval.DefaultPalette(), ivs("d3", "m3", "M3", "A3"), []pitch.Pitch{wp(trad(), spelling.FFlat(), 4), wp(trad(), spelling.F(), 4), wp(trad(), spelling.FSharp(), 4), wp(trad(), spelling.FDoubleSharp(), 4)}},
		{"fifth from C", 5, trad(), wp(trad(), spelling.C(), 4), interval.Up, interval.DefaultPalette(), ivs("dd5", "d5", "P5", "A5", "AA5"), []pitch.Pitch{wp(trad(), spelling.GDoubleFlat(), 4), wp(trad(), spelling.GFlat(), 4), wp(trad(), spelling.G(), 4), wp(trad(), spelling.GSharp(), 4), wp(trad(), spelling.GDoubleSharp(), 4)}},
		{"traditional filters fractional", 3, trad(), wp(trad(), spelling.C(), 4), interval.Up, interval.Deviations(ai(-1), alt(-1, 2), ai(0)), ivs("m3", "M3"), []pitch.Pitch{wp(trad(), spelling.EFlat(), 4), wp(trad(), spelling.E(), 4)}},
		{"quarter tones retain neutral", 3, qt24(), wp(qt24(), spelling.C(), 4), interval.Up, interval.Deviations(ai(-1), alt(-1, 2), ai(0)), ivs("m3", "n3", "M3"), []pitch.Pitch{wp(qt24(), spelling.EFlat(), 4), wp(qt24(), eHalfFlat(), 4), wp(qt24(), spelling.E(), 4)}},
		{"no admissible candidates", 3, trad(), wp(trad(), spelling.C(), 4), interval.Up, interval.Deviations(alt(-1, 2)), nil, nil},
		{"descending compound", 10, trad(), wp(trad(), spelling.C(), 4), interval.Down, interval.MustThrough(0), ivs("↓m10", "↓M10"), []pitch.Pitch{wp(trad(), spelling.A(), 2), wp(trad(), spelling.AFlat(), 2)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			seq, err := interval.MustNewGeneric(tc.size).FromPitch(tc.sys, tc.start, tc.dir, tc.p)
			require.NoError(t, err)
			require.NotNil(t, seq)
			check := func() {
				got := slices.Collect(seq)
				require.Len(t, got, len(tc.wantI))
				for index, v := range got {
					assert.True(t, tc.wantI[index].Equal(v.Interval))
					assert.True(t, tc.wantP[index].Equal(v.Pitch))
				}
			}
			check()
			for range seq {
				break
			}
			check()
			assert.Equal(t, slices.Collect(seq), slices.Collect(interval.MustNewGeneric(tc.size).MustFromPitch(tc.sys, tc.start, tc.dir, tc.p)))
		})
	}
	g := interval.MustNewGeneric(3)
	start := wp(trad(), spelling.C(), 4)
	for _, tc := range []struct {
		g     interval.Generic
		sys   *system.System
		start pitch.Pitch
		dir   interval.Direction
		want  error
	}{
		{interval.Generic{}, trad(), start, interval.Up, interval.ErrUnset},
		{g, nil, start, interval.Up, interval.ErrUnset},
		{g, trad(), pitch.Pitch{}, interval.Up, interval.ErrUnset},
		{g, trad(), start, interval.DirectionUnset, interval.ErrNoDirection},
		{g, trad(), start, interval.Direction(99), interval.ErrInvalidArgument},
		{g, trad(), wp(qt24(), eHalfFlat(), 4), interval.Up, system.ErrInadmissible},
	} {
		seq, err := tc.g.FromPitch(tc.sys, tc.start, tc.dir, interval.Deviations())
		assert.ErrorIs(t, err, tc.want)
		assert.Nil(t, seq)
	}
	assert.ErrorIs(t, panicError(t, func() { g.MustFromPitch(nil, start, interval.Up, interval.Palette{}) }), interval.ErrUnset)
}

func TestGenericResolve(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name        string
		size, steps int
		sys         *system.System
		start       spelling.Spelling
		dir         interval.Direction
		want        interval.Interval
	}{
		{"augmented second", 2, 3, trad(), spelling.C(), interval.Up, iv("A2")},
		{"minor third", 3, 3, trad(), spelling.C(), interval.Up, iv("m3")},
		{"neutral", 3, 7, qt24(), spelling.C(), interval.Up, iv("n3")},
		{"regular without start", 3, 3, trad(), noStart, interval.Up, iv("m3")},
		{"regular down without start", 3, -3, trad(), noStart, interval.Down, iv("↓m3")},
		{"irregular upward", 2, 1, irregular(), spelling.D(), interval.Up, iv("M2")},
		{"irregular downward", 2, -3, irregular(), spelling.D(), interval.Down, iv("↓M2")},
		{"written up with negative shift", 2, -1, trad(), spelling.C(), interval.Up, iv("dd2")},
		{"descending augmented unison", 1, -1, trad(), spelling.C(), interval.Down, iv("↓A1")},
		{"canonical zero unison", 1, 0, trad(), spelling.C(), interval.Down, iv("P1")},
		{"compound", 10, 15, trad(), spelling.C(), interval.Up, iv("m10")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := interval.MustNewChromatic(tc.sys, tc.steps)
			g := interval.MustNewGeneric(tc.size)
			got, err := g.Resolve(c, tc.start, tc.dir)
			require.NoError(t, err)
			assert.True(t, tc.want.Equal(got), "got %v, want %v", got, tc.want)
			assert.Equal(t, got, g.MustResolve(c, tc.start, tc.dir))
			st, err := got.SignedSteps(tc.sys, tc.start)
			require.NoError(t, err)
			assert.Equal(t, tc.steps, int(st))
		})
	}
	for _, sys := range []*system.System{trad(), qt24(), irregular(), boundaryA()} {
		for _, dir := range []interval.Direction{interval.Up, interval.Down} {
			for _, size := range []int{1, 2, 3, 5, 8, 10, 15} {
				for _, original := range slices.Collect(interval.MustNewGeneric(size).MustVariants(dir, interval.DefaultPalette())) {
					c, err := original.Chromatic(sys, spelling.D())
					require.NoError(t, err)
					got, err := interval.MustNewGeneric(size).Resolve(c, spelling.D(), dir)
					require.NoError(t, err)
					assert.True(t, original.Equal(got), "%v in %v", original, sys)
				}
			}
		}
	}
	g := interval.MustNewGeneric(3)
	for _, tc := range []struct {
		g           interval.Generic
		c           interval.Chromatic
		start       spelling.Spelling
		dir         interval.Direction
		want, cause error
	}{
		{interval.Generic{}, interval.Chromatic{}, noStart, interval.DirectionUnset, interval.ErrUnset, nil},
		{g, interval.Chromatic{}, noStart, interval.Up, interval.ErrUnset, nil},
		{g, interval.MustNewChromatic(trad(), 3), noStart, interval.DirectionUnset, interval.ErrNoDirection, nil},
		{g, interval.MustNewChromatic(trad(), 3), noStart, interval.Direction(99), interval.ErrInvalidArgument, nil},
		{g, interval.MustNewChromatic(irregular(), 3), noStart, interval.Up, interval.ErrStartRequired, nil},
		{g, interval.MustNewChromatic(trad(), 3), eHalfFlat(), interval.Up, system.ErrInadmissible, nil},
		{interval.MustNewGeneric(1), interval.MustNewChromatic(trad(), -1), spelling.C(), interval.Up, interval.ErrInvalidArgument, nil},
		{g, interval.MustNewChromatic(trad(), math.MinInt), noStart, interval.Up, interval.ErrOutOfRange, alteration.ErrOutOfRange},
		{interval.MustNewGeneric(math.MaxInt), interval.MustNewChromatic(trad(), math.MaxInt), noStart, interval.Up, interval.ErrOutOfRange, alteration.ErrOutOfRange},
	} {
		got, err := tc.g.Resolve(tc.c, tc.start, tc.dir)
		assert.ErrorIs(t, err, tc.want)
		if tc.cause != nil {
			assert.ErrorIs(t, err, tc.cause)
		}
		assert.Equal(t, interval.Interval{}, got)
	}
	assert.ErrorIs(t, panicError(t, func() { g.MustResolve(interval.Chromatic{}, noStart, interval.Up) }), interval.ErrUnset)
}

func TestForShift(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name  string
		steps int
		dir   interval.Direction
		p     interval.Palette
		want  []interval.Interval
	}{
		{"tritone", 6, interval.Up, interval.MustThrough(2), ivs("AA3", "A4", "d5", "dd6")},
		{"major third only", 4, interval.Up, interval.Deviations(ai(-1), alt(-1, 2), ai(0)), ivs("d4", "M3")},
		{"empty", 3, interval.Up, interval.Deviations(), nil},
		{"written up negative", -1, interval.Up, interval.MustThrough(2), ivs("dd2")},
		{"zero", 0, interval.Down, interval.MustThrough(0), ivs("P1")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			seq, err := interval.ForShift(interval.MustNewChromatic(trad(), tc.steps), spelling.C(), tc.dir, tc.p)
			require.NoError(t, err)
			require.NotNil(t, seq)
			// All candidates are sorted by full generic size.
			want := slices.Clone(tc.want)
			slices.SortFunc(want, func(a, b interval.Interval) int { return a.Size() - b.Size() })
			assert.Equal(t, want, slices.Collect(seq))
			assert.Equal(t, want, slices.Collect(seq))
			assert.Equal(t, want, slices.Collect(interval.MustForShift(interval.MustNewChromatic(trad(), tc.steps), spelling.C(), tc.dir, tc.p)))
		})
	}
	for _, tc := range []struct {
		c     interval.Chromatic
		start spelling.Spelling
		dir   interval.Direction
		want  error
	}{
		{interval.Chromatic{}, noStart, interval.Up, interval.ErrUnset},
		{interval.MustNewChromatic(trad(), 3), noStart, interval.DirectionUnset, interval.ErrNoDirection},
		{interval.MustNewChromatic(trad(), 3), noStart, interval.Direction(99), interval.ErrInvalidArgument},
		{interval.MustNewChromatic(irregular(), 3), noStart, interval.Up, interval.ErrStartRequired},
		{interval.MustNewChromatic(trad(), 3), eHalfFlat(), interval.Up, system.ErrInadmissible},
	} {
		seq, err := interval.ForShift(tc.c, tc.start, tc.dir, interval.Deviations())
		assert.ErrorIs(t, err, tc.want)
		assert.Nil(t, seq)
	}
	assert.ErrorIs(t, panicError(t, func() { interval.MustForShift(interval.Chromatic{}, noStart, interval.Up, interval.Palette{}) }), interval.ErrUnset)
	// Huge signed values must have finite analytic windows; stop at the first result.
	for _, steps := range []int{math.MinInt, math.MaxInt} {
		for _, dir := range []interval.Direction{interval.Up, interval.Down} {
			seq, err := interval.ForShift(interval.MustNewChromatic(trad(), steps), noStart, dir, interval.MustThrough(0))
			require.NoError(t, err)
			var want []interval.Interval
			if steps == math.MaxInt && dir == interval.Up {
				size := 5 + 7*((math.MaxInt-7)/12)
				want = []interval.Interval{basicSize(size)}
			}
			if steps == math.MinInt && dir == interval.Down {
				magnitude := uint64(math.MaxInt) + 1
				size := 6 + 7*int((magnitude-8)/12)
				want = []interval.Interval{interval.MustFromDeviation(size, ai(-1)).MustDown()}
			}
			assert.Equal(t, want, slices.Collect(seq))
			for range seq {
				break
			}
			assert.Equal(t, want, slices.Collect(seq))
		}
	}
}

func TestGenericExactExtremesAndSkippedCandidates(t *testing.T) {
	t.Parallel()
	// Seven naturals at one position are legal. A very large generic size is
	// representable even though multiplying its letter coordinate by 12 is not.
	sys := system.MustNew(system.Definition{Name: "one position", Classes: 1, Boundary: letter.C, Offsets: [letter.Count]int{}, Chromatic: 1, Alterations: system.Integers()})
	g := interval.MustNewGeneric(math.MaxInt)
	octaves := (math.MaxInt - 1) / 7
	for _, dir := range []interval.Direction{interval.Up, interval.Down} {
		steps := octaves
		if dir == interval.Down {
			steps = -steps
		}
		got, err := g.Resolve(interval.MustNewChromatic(sys, steps), spelling.C(), dir)
		require.NoError(t, err)
		want := basicSize(math.MaxInt).MustWithDirection(dir)
		assert.True(t, want.Equal(got))
	}
	// Fractional mathematical deviation can exist, but the final spelling may
	// still be forbidden by the system's alteration rule.
	_, err := interval.MustNewGeneric(3).Resolve(interval.MustNewChromatic(system.QuarterTones24(), 7), spelling.C(), interval.Up)
	require.NoError(t, err)
	restricted := system.MustNew(system.Definition{Name: "integer 24", Classes: 24, Boundary: letter.C, Offsets: [letter.Count]int{0, 4, 8, 10, 14, 18, 22}, Chromatic: 2, Alterations: system.Integers()})
	_, err = interval.MustNewGeneric(3).Resolve(interval.MustNewChromatic(restricted, 7), spelling.C(), interval.Up)
	assert.ErrorIs(t, err, system.ErrInadmissible)
	start := wp(trad(), spelling.MustNew(letter.C, ai(mag)), 4)
	seq, err := interval.MustNewGeneric(1).FromPitch(trad(), start, interval.Up, interval.DefaultPalette())
	require.NoError(t, err)
	got := slices.Collect(seq)
	require.Len(t, got, 1, "unrepresentable raised endings are filtered")
	assert.True(t, iv("P1").Equal(got[0].Interval))
	assert.True(t, start.Equal(got[0].Pitch))
	// Search at +6 in a traditional system uses the selected palette, while
	// irregular and quarter-tone searches preserve their own exact distances.
	for _, tc := range []struct {
		sys   *system.System
		steps int
		start spelling.Spelling
		dir   interval.Direction
		p     interval.Palette
		want  []interval.Interval
	}{
		{qt24(), 7, spelling.C(), interval.Up, interval.Deviations(alt(-1, 2)), ivs("n3")},
		{irregular(), 1, spelling.D(), interval.Up, interval.MustThrough(0), ivs("M2", "m3")},
		{irregular(), -3, spelling.D(), interval.Down, interval.MustThrough(0), ivs("↓M2", "↓m3")},
	} {
		seq, err := interval.ForShift(interval.MustNewChromatic(tc.sys, tc.steps), tc.start, tc.dir, tc.p)
		require.NoError(t, err)
		assert.Equal(t, tc.want, slices.Collect(seq))
	}
}

func TestPaletteMagnitudeBoundaryIsLazy(t *testing.T) {
	t.Parallel()
	seq, err := interval.MustNewGeneric(3).Variants(interval.Up, interval.MustThrough(alteration.MaxMagnitude))
	require.NoError(t, err)
	seen := false
	for candidate := range seq {
		assert.Equal(t, ai(-alteration.MaxMagnitude), candidate.Deviation())
		assert.Equal(t, 3, candidate.Size())
		seen = true
		break
	}
	assert.True(t, seen)
}
