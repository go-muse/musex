package ticks_test

import (
	"testing"

	"github.com/go-muse/musex/format/ticks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDeltas checks О-10, ИТ-5, ТК-6, КП-6, КП-22, КП-29.
func TestDeltas(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name     string
		absolute []int
		deltas   []int
	}{
		{name: "KP-6", absolute: []int{0, 0, 96, 192, 384}, deltas: []int{0, 0, 96, 96, 192}},
		{name: "first delta is the first tick", absolute: []int{10, 15}, deltas: []int{10, 5}},
		{name: "KP-29 largest delta", absolute: []int{0, 268435455}, deltas: []int{0, 268435455}},
		{name: "largest first delta", absolute: []int{268435455}, deltas: []int{268435455}},
		{name: "single zero", absolute: []int{0}, deltas: []int{0}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			in := append([]int(nil), tt.absolute...)
			got, err := ticks.Deltas(in)
			require.NoError(t, err)
			assert.Equal(t, tt.deltas, got)
			assert.Equal(t, tt.absolute, in, "input changed")
			assert.Equal(t, tt.deltas, ticks.MustDeltas(tt.absolute))

			back := append([]int(nil), tt.deltas...)
			abs, err := ticks.Absolute(back)
			require.NoError(t, err)
			assert.Equal(t, tt.absolute, abs)
			assert.Equal(t, tt.deltas, back, "input changed")
			assert.Equal(t, tt.absolute, ticks.MustAbsolute(tt.deltas))

			// The result is a new slice.
			got[0] = -100
			assert.Equal(t, tt.absolute, in)
		})
	}
	for _, empty := range [][]int{nil, {}} {
		got, err := ticks.Deltas(empty)
		require.NoError(t, err)
		assert.Nil(t, got)
		got, err = ticks.Absolute(empty)
		require.NoError(t, err)
		assert.Nil(t, got)
	}
}

func TestDeltaErrors(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name     string
		absolute []int
		own      error
	}{
		{name: "KP-22 decreasing", absolute: []int{0, 96, 48}, own: ticks.ErrOrder},
		{name: "KP-22 delta above the limit", absolute: []int{0, 268435456}, own: ticks.ErrFormatLimit},
		{name: "first delta above the limit", absolute: []int{268435456}, own: ticks.ErrFormatLimit},
		{name: "negative tick", absolute: []int{-1}, own: ticks.ErrInvalidArgument},
		{name: "negative is checked before the order inside an element", absolute: []int{10, -1}, own: ticks.ErrInvalidArgument},
		{name: "elements are checked in order: order first", absolute: []int{10, 5, -1}, own: ticks.ErrOrder},
		{name: "elements are checked in order: limit first", absolute: []int{0, 268435456, 1}, own: ticks.ErrFormatLimit},
		{name: "elements are checked in order: negative first", absolute: []int{-1, 268435456}, own: ticks.ErrInvalidArgument},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ticks.Deltas(tt.absolute)
			requireTicksError(t, err, tt.own)
			assert.Nil(t, got)
			requireMustPanic(t, err, tt.own, func() { ticks.MustDeltas(tt.absolute) })
		})
	}
	for _, tt := range []struct {
		name   string
		deltas []int
		own    error
	}{
		{name: "negative delta", deltas: []int{0, -1}, own: ticks.ErrInvalidArgument},
		{name: "delta above the limit", deltas: []int{268435456}, own: ticks.ErrFormatLimit},
		{name: "elements are checked in order", deltas: []int{268435456, -1}, own: ticks.ErrFormatLimit},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ticks.Absolute(tt.deltas)
			requireTicksError(t, err, tt.own)
			assert.Nil(t, got)
			requireMustPanic(t, err, tt.own, func() { ticks.MustAbsolute(tt.deltas) })
		})
	}
}

func TestIsDelta(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		delta int
		want  bool
	}{
		{delta: -1, want: false},
		{delta: 0, want: true},
		{delta: 1, want: true},
		{delta: ticks.MaxDelta, want: true},
		{delta: ticks.MaxDelta + 1, want: false},
	} {
		assert.Equal(t, tt.want, ticks.IsDelta(tt.delta), "%d", tt.delta)
	}
}
