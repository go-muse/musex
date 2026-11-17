package degree_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/degree"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDegreeSentinelsDoNotWrapEachOther(t *testing.T) {
	t.Parallel()
	sentinels := degreeErrorSentinels()
	for i, left := range sentinels {
		for j, right := range sentinels {
			if i == j {
				assert.ErrorIs(t, left, right)
				continue
			}
			assert.NotErrorIs(t, left, right)
		}
	}
}

func TestDegreeNewWithNoteErrorPrecedenceAndZeroResults(t *testing.T) {
	t.Parallel()
	valid := degreeTestNote(t, spelling.B(), cents.Zero(), nil)
	tests := []struct {
		name     string
		number   int
		note     note.Note
		expected error
	}{
		{"unset note before invalid number", 0, note.Note{}, degree.ErrUnset},
		{"invalid number with valid note", 0, valid, degree.ErrInvalidNumber},
		{"negative number with valid note", -1, valid, degree.ErrInvalidNumber},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d, err := degree.NewWithNote(tt.number, tt.note)
			assertDegreeError(t, err, tt.expected)
			assertUnsetDegree(t, d)
			_ = assertMustPanic(t, tt.expected, func() { degree.MustNewWithNote(tt.number, tt.note) })
		})
	}
}

func TestDegreeMutatorsRejectUnsetReceiver(t *testing.T) {
	t.Parallel()
	valid := degreeTestNote(t, spelling.B(), cents.Zero(), nil)
	identity := func(n note.Note) (note.Note, error) { return n, nil }
	tests := []struct {
		name string
		call func(*degree.Degree) error
		must func(*degree.Degree)
	}{
		{"SetNote", func(d *degree.Degree) error { return d.SetNote(valid) }, func(d *degree.Degree) { d.MustSetNote(valid) }},
		{"RemoveNote", func(d *degree.Degree) error { return d.RemoveNote() }, func(d *degree.Degree) { d.MustRemoveNote() }},
		{"TransformNote", func(d *degree.Degree) error { return d.TransformNote(identity) }, func(d *degree.Degree) { d.MustTransformNote(identity) }},
		{"SetNumber", func(d *degree.Degree) error { return d.SetNumber(1) }, func(d *degree.Degree) { d.MustSetNumber(1) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := degree.Degree{}
			err := tt.call(&d)
			assertDegreeError(t, err, degree.ErrUnset)
			assertUnsetDegree(t, d)
			_ = assertMustPanic(t, degree.ErrUnset, func() { tt.must(&d) })
			assertUnsetDegree(t, d)
		})
	}
}

func TestDegreeInCircleErrorsAndZeroResult(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		d        degree.Degree
		k        int
		expected error
	}{
		{"unset before invalid k", degree.Degree{}, 0, degree.ErrUnset},
		{"zero k", degree.First(), 0, degree.ErrInvalidNumber},
		{"negative k", degree.First(), -1, degree.ErrInvalidNumber},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.d.InCircle(tt.k)
			assertDegreeError(t, err, tt.expected)
			assert.False(t, got)
			_ = assertMustPanic(t, tt.expected, func() { tt.d.MustInCircle(tt.k) })
		})
	}
}

func TestDegreeNextAndPreviousErrorsAndZeroResults(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		d        degree.Degree
		k        int
		expected error
	}{
		{"unset before invalid k", degree.Degree{}, 0, degree.ErrUnset},
		{"zero k", degree.First(), 0, degree.ErrInvalidNumber},
		{"negative k", degree.First(), -1, degree.ErrInvalidNumber},
		{"number above k", degree.MustNew(8), 7, degree.ErrInvalidNumber},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			next, wrapped, err := tt.d.Next(tt.k)
			assertDegreeError(t, err, tt.expected)
			assertUnsetDegree(t, next)
			assert.False(t, wrapped)
			_ = assertMustPanic(t, tt.expected, func() { tt.d.MustNext(tt.k) })

			previous, wrapped, err := tt.d.Previous(tt.k)
			assertDegreeError(t, err, tt.expected)
			assertUnsetDegree(t, previous)
			assert.False(t, wrapped)
			_ = assertMustPanic(t, tt.expected, func() { tt.d.MustPrevious(tt.k) })
		})
	}
}

func TestDegreeShiftErrorsAndZeroResults(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		d        degree.Degree
		k        int
		expected error
	}{
		{"unset before invalid k", degree.Degree{}, 0, degree.ErrUnset},
		{"zero k", degree.First(), 0, degree.ErrInvalidNumber},
		{"negative k", degree.First(), -1, degree.ErrInvalidNumber},
		{"number above k", degree.MustNew(8), 7, degree.ErrInvalidNumber},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			shifted, cycles, err := tt.d.Shift(tt.k, math.MaxInt)
			assertDegreeError(t, err, tt.expected)
			assertUnsetDegree(t, shifted)
			assert.Zero(t, cycles)
			_ = assertMustPanic(t, tt.expected, func() { tt.d.MustShift(tt.k, math.MaxInt) })
		})
	}
}

func TestDegreeDistanceErrorsAndZeroResults(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		from     degree.Degree
		to       degree.Degree
		k        int
		expected error
	}{
		{"unset receiver before other errors", degree.Degree{}, degree.Degree{}, 0, degree.ErrUnset},
		{"unset argument before invalid k", degree.First(), degree.Degree{}, 0, degree.ErrUnset},
		{"zero k", degree.First(), degree.Second(), 0, degree.ErrInvalidNumber},
		{"negative k", degree.First(), degree.Second(), -1, degree.ErrInvalidNumber},
		{"receiver above k", degree.MustNew(8), degree.First(), 7, degree.ErrInvalidNumber},
		{"argument above k", degree.First(), degree.MustNew(8), 7, degree.ErrInvalidNumber},
		{"both above k", degree.MustNew(8), degree.MustNew(9), 7, degree.ErrInvalidNumber},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			up, err := tt.from.DistanceUp(tt.to, tt.k)
			assertDegreeError(t, err, tt.expected)
			assert.Zero(t, up)
			_ = assertMustPanic(t, tt.expected, func() { tt.from.MustDistanceUp(tt.to, tt.k) })

			down, err := tt.from.DistanceDown(tt.to, tt.k)
			assertDegreeError(t, err, tt.expected)
			assert.Zero(t, down)
			_ = assertMustPanic(t, tt.expected, func() { tt.from.MustDistanceDown(tt.to, tt.k) })

			shortest, err := tt.from.ShortestDistance(tt.to, tt.k)
			assertDegreeError(t, err, tt.expected)
			assert.Zero(t, shortest)
			_ = assertMustPanic(t, tt.expected, func() { tt.from.MustShortestDistance(tt.to, tt.k) })
		})
	}
}

func TestDegreeNormalizeErrorsAndZeroResults(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		d        degree.Degree
		k        int
		expected error
	}{
		{"unset before invalid k", degree.Degree{}, 0, degree.ErrUnset},
		{"zero k", degree.First(), 0, degree.ErrInvalidNumber},
		{"negative k", degree.First(), -1, degree.ErrInvalidNumber},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			normalized, cycles, err := tt.d.Normalize(tt.k)
			assertDegreeError(t, err, tt.expected)
			assertUnsetDegree(t, normalized)
			assert.Zero(t, cycles)
			_ = assertMustPanic(t, tt.expected, func() { tt.d.MustNormalize(tt.k) })
		})
	}
}

func TestDegreeDenormalizeErrorsAndZeroResults(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		d        degree.Degree
		k        int
		cycles   int
		expected error
	}{
		{"unset before other errors", degree.Degree{}, 0, -1, degree.ErrUnset},
		{"zero k", degree.First(), 0, 0, degree.ErrInvalidNumber},
		{"negative k", degree.First(), -1, 0, degree.ErrInvalidNumber},
		{"number above k", degree.MustNew(8), 7, 0, degree.ErrInvalidNumber},
		{"negative cycles", degree.First(), 7, -1, degree.ErrInvalidNumber},
		{"invalid number before overflow", degree.MustNew(8), 7, math.MaxInt, degree.ErrInvalidNumber},
		{"negative cycles before overflow", degree.MustNew(7), 7, -1, degree.ErrInvalidNumber},
		{"one cycle above last representable", degree.MustNew(7), 7, 1317624576693539401, degree.ErrOutOfRange},
		{"large multiplication overflow", degree.MustNew(math.MaxInt), math.MaxInt, 2, degree.ErrOutOfRange},
		{"large addition overflow", degree.MustNew(math.MaxInt), math.MaxInt, 1, degree.ErrOutOfRange},
		{"k one cycles overflow", degree.First(), 1, math.MaxInt, degree.ErrOutOfRange},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d, err := tt.d.Denormalize(tt.k, tt.cycles)
			assertDegreeError(t, err, tt.expected)
			assertUnsetDegree(t, d)
			_ = assertMustPanic(t, tt.expected, func() { tt.d.MustDenormalize(tt.k, tt.cycles) })
		})
	}
}

func TestDegreeAllRejectsInvalidCircleSize(t *testing.T) {
	t.Parallel()
	for _, k := range []int{math.MinInt, -1, 0} {
		t.Run(fmt.Sprintf("k_%d", k), func(t *testing.T) {
			t.Parallel()
			sequence, err := degree.All(k)
			assertDegreeError(t, err, degree.ErrInvalidNumber)
			assert.Nil(t, sequence)
			_ = assertMustPanic(t, degree.ErrInvalidNumber, func() { degree.MustAll(k) })
		})
	}
}

func TestDegreeOrdinaryOperationsDoNotPanicOnInvalidInput(t *testing.T) {
	t.Parallel()
	identity := func(n note.Note) (note.Note, error) { return n, nil }
	assert.NotPanics(t, func() {
		_, _ = degree.New(0)
		_, _ = degree.NewWithNote(0, note.Note{})
		d := degree.MustNew(4)
		_ = d.SetNote(note.Note{})
		_ = d.TransformNote(identity)
		_ = d.SetNumber(0)
		_, _ = d.Decimal()
		_, _ = d.Roman()
		_, _ = degree.Parse("not a degree")
		_, _ = d.InCircle(0)
		_, _, _ = d.Next(0)
		_, _, _ = d.Previous(0)
		_, _, _ = d.Shift(0, math.MinInt)
		_, _ = d.DistanceUp(degree.Degree{}, 0)
		_, _ = d.DistanceDown(degree.Degree{}, 0)
		_, _ = d.ShortestDistance(degree.Degree{}, 0)
		_, _, _ = d.Normalize(0)
		_, _ = d.Denormalize(0, -1)
		_, _ = degree.All(0)
	})
}

func TestDegreeTransformErrorDoesNotAcquireOtherDegreeSentinel(t *testing.T) {
	t.Parallel()
	d := degree.MustNewWithNote(4, degreeTestNote(t, spelling.B(), cents.Zero(), nil))
	cause := fmt.Errorf("wrapped cause: %w", note.ErrNoOctave)
	err := d.TransformNote(func(note.Note) (note.Note, error) {
		return note.Note{}, cause
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, degree.ErrTransformFailed)
	assert.ErrorIs(t, err, note.ErrNoOctave)
	for _, sentinel := range degreeErrorSentinels() {
		if sentinel == degree.ErrTransformFailed {
			continue
		}
		assert.NotErrorIs(t, err, sentinel)
	}
}
