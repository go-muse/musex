package ticks_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/format/ticks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMinResolution checks О-9, ИТ-3, П-13, ТК-4, КП-3, КП-27.
func TestMinResolution(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name      string
		durations []duration.Duration
		positions []duration.Position
		q         []ticks.MIDIQuarter
		want      int
	}{
		{name: "KP-3 quarter, eighth, triplet eighth", durations: []duration.Duration{dur(1, 4), dur(1, 8), dur(1, 12)}, want: 6},
		{name: "KP-3 sixteenth and quintuplet sixteenth", durations: []duration.Duration{dur(1, 16), dur(1, 20)}, want: 20},
		{name: "KP-3 four values", durations: []duration.Duration{dur(1, 4), dur(1, 8), dur(1, 12), dur(1, 20)}, want: 30},
		{name: "KP-3 septuplet sixteenth", durations: []duration.Duration{dur(1, 28)}, want: 7},
		{name: "KP-27 empty set", want: 1},
		{name: "positions count", positions: []duration.Position{pos(3, 8)}, want: 2},
		{name: "position zero", positions: []duration.Position{duration.Origin()}, want: 1},
		{name: "durations and positions", durations: []duration.Duration{dur(1, 8)}, positions: []duration.Position{pos(1, 12)}, want: 6},
		{name: "written half quarter", durations: []duration.Duration{dur(1, 4)}, q: []ticks.MIDIQuarter{ticks.MustNewMIDIQuarter(16)}, want: 2},
		{name: "explicit standard quarter", durations: []duration.Duration{dur(1, 12)}, q: []ticks.MIDIQuarter{ticks.StandardQuarter()}, want: 3},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r, err := ticks.MinResolution(tt.durations, tt.positions, tt.q...)
			require.NoError(t, err)
			assert.Equal(t, tt.want, r.MustPerQuarter())
			wantQuarter := ticks.StandardQuarter()
			if len(tt.q) == 1 {
				wantQuarter = tt.q[0]
			}
			assert.True(t, r.MustQuarter().Equal(wantQuarter))
			assert.True(t, r.Equal(ticks.MustMinResolution(tt.durations, tt.positions, tt.q...)))
			ok, err := r.Represents(tt.durations, tt.positions)
			require.NoError(t, err)
			assert.True(t, ok)
		})
	}
}

// TestMinResolutionChain checks the chain of 4.5 and КП-3 with lower bounds.
func TestMinResolutionChain(t *testing.T) {
	t.Parallel()
	set := []duration.Duration{dur(1, 4), dur(1, 8), dur(1, 12)}
	r := ticks.MustMinResolution(set, nil).MustAtLeast(480).MustAtMost(ticks.MaxSMFResolution)
	assert.Equal(t, 480, r.MustPerQuarter())
	assert.Equal(t, 483, ticks.MustMinResolution([]duration.Duration{dur(1, 28)}, nil).MustAtLeast(480).MustPerQuarter())
	assert.Equal(t, 7, res(7).MustAtLeast(1).MustPerQuarter())
	assert.Equal(t, 14, res(7).MustAtLeast(8).MustPerQuarter())
	assert.Equal(t, 6, res(6).MustAtLeast(6).MustPerQuarter())
	// The written quarter of the receiver is kept.
	assert.True(t, resQ(2, 16).MustAtLeast(5).Equal(resQ(6, 16)))
	assert.True(t, resQ(480, 16).MustAtMost(480).Equal(resQ(480, 16)))
	// math.MaxInt is a multiple of 7.
	assert.Equal(t, math.MaxInt, res(7).MustAtLeast(math.MaxInt).MustPerQuarter())
}

func TestChooseErrors(t *testing.T) {
	t.Parallel()
	primes := []duration.Duration{dur(1, 1000003), dur(1, 1000033), dur(1, 1000037), dur(1, 1000039)}
	for _, tt := range []struct {
		name string
		call func() error
		must func()
		own  error
		deps []error
	}{
		{
			name: "unset duration",
			call: func() error {
				_, err := ticks.MinResolution([]duration.Duration{dur(1, 4), {}}, nil)
				return err
			},
			must: func() { ticks.MustMinResolution([]duration.Duration{dur(1, 4), {}}, nil) },
			own:  ticks.ErrInvalidArgument,
			deps: []error{duration.ErrUnset},
		},
		{
			name: "unset position",
			call: func() error { _, err := ticks.MinResolution(nil, []duration.Position{{}}); return err },
			must: func() { ticks.MustMinResolution(nil, []duration.Position{{}}) },
			own:  ticks.ErrInvalidArgument,
			deps: []error{duration.ErrUnset},
		},
		{
			name: "two written quarters",
			call: func() error {
				_, err := ticks.MinResolution(nil, nil, ticks.StandardQuarter(), ticks.StandardQuarter())
				return err
			},
			must: func() { ticks.MustMinResolution(nil, nil, ticks.StandardQuarter(), ticks.StandardQuarter()) },
			own:  ticks.ErrInvalidArgument,
		},
		{
			name: "unset written quarter",
			call: func() error { _, err := ticks.MinResolution(nil, nil, ticks.MIDIQuarter{}); return err },
			must: func() { ticks.MustMinResolution(nil, nil, ticks.MIDIQuarter{}) },
			own:  ticks.ErrUnset,
		},
		{
			name: "least common multiple outside int",
			call: func() error { _, err := ticks.MinResolution(primes, nil); return err },
			must: func() { ticks.MustMinResolution(primes, nil) },
			own:  ticks.ErrOutOfRange,
		},
		{
			name: "at least zero",
			call: func() error { _, err := res(6).AtLeast(0); return err },
			must: func() { res(6).MustAtLeast(0) },
			own:  ticks.ErrInvalidArgument,
		},
		{
			name: "at least outside int",
			call: func() error { _, err := res(2).AtLeast(math.MaxInt); return err },
			must: func() { res(2).MustAtLeast(math.MaxInt) },
			own:  ticks.ErrOutOfRange,
		},
		{
			name: "at most zero",
			call: func() error { _, err := res(6).AtMost(0); return err },
			must: func() { res(6).MustAtMost(0) },
			own:  ticks.ErrInvalidArgument,
		},
		{
			name: "above the SMF limit",
			call: func() error { _, err := res(40000).AtMost(ticks.MaxSMFResolution); return err },
			must: func() { res(40000).MustAtMost(ticks.MaxSMFResolution) },
			own:  ticks.ErrFormatLimit,
		},
		{
			name: "KP-30 above the MusicXML recommendation",
			call: func() error { _, err := res(20000).AtMost(ticks.MaxMusicXMLResolution); return err },
			must: func() { res(20000).MustAtMost(ticks.MaxMusicXMLResolution) },
			own:  ticks.ErrFormatLimit,
		},
		{
			name: "unset duration in Represents",
			call: func() error { _, err := res(6).Represents([]duration.Duration{{}}, nil); return err },
			must: func() { res(6).MustRepresents([]duration.Duration{{}}, nil) },
			own:  ticks.ErrInvalidArgument,
			deps: []error{duration.ErrUnset},
		},
		{
			name: "unset position in Unrepresentable",
			call: func() error { _, _, err := res(6).Unrepresentable(nil, []duration.Position{{}}); return err },
			must: func() { res(6).MustUnrepresentable(nil, []duration.Position{{}}) },
			own:  ticks.ErrInvalidArgument,
			deps: []error{duration.ErrUnset},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.call()
			requireTicksError(t, err, tt.own, tt.deps...)
			requireMustPanic(t, err, tt.own, tt.must)
		})
	}
	assert.Equal(t, 32767, res(32767).MustAtMost(ticks.MaxSMFResolution).MustPerQuarter())
	assert.Equal(t, 16383, res(16383).MustAtMost(ticks.MaxMusicXMLResolution).MustPerQuarter())
}

// TestRepresentsAndUnrepresentable checks П-13, КП-42.
func TestRepresentsAndUnrepresentable(t *testing.T) {
	t.Parallel()
	set := []duration.Duration{dur(1, 4), dur(1, 8), dur(1, 12)}
	ok, err := res(480).Represents(set, nil)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.True(t, res(480).MustRepresents(set, nil))
	ok, err = res(100).Represents(set, nil)
	require.NoError(t, err)
	assert.False(t, ok)
	assert.False(t, res(100).MustRepresents(set, nil))

	ds, ps, err := res(100).Unrepresentable(set, nil)
	require.NoError(t, err)
	require.Len(t, ds, 1)
	assertEqualValue(t, dur(1, 12), ds[0])
	assert.Nil(t, ps)

	// Input order and repeats are kept.
	dsIn := []duration.Duration{dur(1, 12), dur(1, 4), dur(1, 7), dur(1, 12)}
	psIn := []duration.Position{pos(1, 3), duration.Origin(), pos(5, 12)}
	ds, ps = res(100).MustUnrepresentable(dsIn, psIn)
	require.Len(t, ds, 3)
	assertEqualValue(t, dur(1, 12), ds[0])
	assertEqualValue(t, dur(1, 7), ds[1])
	assertEqualValue(t, dur(1, 12), ds[2])
	require.Len(t, ps, 2)
	assertEqualValue(t, pos(1, 3), ps[0])
	assertEqualValue(t, pos(5, 12), ps[1])

	ds, ps, err = res(480).Unrepresentable(set, []duration.Position{pos(3, 4)})
	require.NoError(t, err)
	assert.Nil(t, ds)
	assert.Nil(t, ps)
	ds, ps, err = res(480).Unrepresentable(nil, nil)
	require.NoError(t, err)
	assert.Nil(t, ds)
	assert.Nil(t, ps)
	assert.True(t, res(1).MustRepresents(nil, nil))

	// With w = 1/2 the units are d/w·D.
	assert.True(t, resQ(1, 16).MustRepresents([]duration.Duration{dur(1, 2)}, nil))
	assert.False(t, resQ(1, 16).MustRepresents([]duration.Duration{dur(1, 4)}, nil))
}
