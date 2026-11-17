package degree_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/degree"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetNoteAndReplace(t *testing.T) {
	t.Parallel()
	d := degree.MustNew(4)
	b := degreeTestNote(t, spelling.B(), cents.Zero(), settings.Standard())
	original := b
	require.NoError(t, d.SetNote(b))
	assertDegreeState(t, d, 4, true, b)
	assert.True(t, b.Equal(original))

	b.MustRaise()
	assertDegreeState(t, d, 4, true, original)

	flat := degreeTestNote(t, spelling.B(), cents.Zero(), nil)
	flat.MustLower()
	require.NoError(t, d.SetNote(flat))
	assertDegreeState(t, d, 4, true, flat)

	returned := d.MustSetNote(original)
	assert.Same(t, &d, returned)
	assertDegreeState(t, d, 4, true, original)
}

func TestSetNoteRejectsUnsetNoteAtomically(t *testing.T) {
	t.Parallel()
	n := degreeTestNote(t, spelling.B(), cents.Zero(), settings.Standard())
	d := degree.MustNewWithNote(4, n)
	before := d
	err := d.SetNote(note.Note{})
	assertDegreeError(t, err, degree.ErrUnset)
	assertSameDegree(t, before, d)
	assert.True(t, before.Note().SameSource(d.Note()))
	assertMustPanic(t, err, func() { d.MustSetNote(note.Note{}) })
	assertSameDegree(t, before, d)
}

func TestRemoveNoteIsIdempotentAndPreservesNumber(t *testing.T) {
	t.Parallel()
	n := degreeTestNote(t, spelling.B(), cents.Zero(), nil)
	d := degree.MustNewWithNote(4, n)
	require.NoError(t, d.RemoveNote())
	assertDegreeState(t, d, 4, false, note.Note{})
	require.NoError(t, d.RemoveNote())
	assertDegreeState(t, d, 4, false, note.Note{})
	returned := d.MustRemoveNote()
	assert.Same(t, &d, returned)
	assertDegreeState(t, d, 4, false, note.Note{})
}

func TestTransformNoteUsesOneIndependentCopy(t *testing.T) {
	t.Parallel()
	flat := degreeTestNote(t, spelling.B(), cents.Zero(), settings.Standard())
	flat.MustLower()
	d := degree.MustNewWithNote(4, flat)
	before := d
	calls := 0
	err := d.TransformNote(func(n note.Note) (note.Note, error) {
		calls++
		assert.True(t, flat.Equal(n))
		require.NoError(t, n.Raise())
		assertSameDegree(t, before, d)
		return n, nil
	})
	require.NoError(t, err)
	assert.Equal(t, 1, calls)
	natural := degreeTestNote(t, spelling.B(), cents.Zero(), settings.Standard())
	assertDegreeState(t, d, 4, true, natural)
	assert.True(t, flat.SameSource(d.Note()))
}

func TestTransformNoteOctaveShift(t *testing.T) {
	t.Parallel()
	d4 := degreeTestASPNNote(t, spelling.D(), 4, cents.Zero(), nil)
	d := degree.MustNewWithNote(2, d4)
	returned := d.MustTransformNote(func(n note.Note) (note.Note, error) {
		require.NoError(t, n.Up())
		return n, nil
	})
	assert.Same(t, &d, returned)
	d5 := degreeTestASPNNote(t, spelling.D(), 5, cents.Zero(), nil)
	assertDegreeState(t, d, 2, true, d5)
}

func TestTransformNoteCopiesReturnedExternalNote(t *testing.T) {
	t.Parallel()
	d := degree.MustNewWithNote(4, degreeTestNote(t, spelling.B(), cents.Zero(), nil))
	replacement := degreeTestNote(t, spelling.C(), cents.Zero(), nil)
	require.NoError(t, d.TransformNote(func(note.Note) (note.Note, error) {
		return replacement, nil
	}))
	expected := replacement
	replacement.MustRaise()
	assertDegreeState(t, d, 4, true, expected)
}

func TestTransformNoteInputErrorsAndPrecedence(t *testing.T) {
	t.Parallel()
	valid := degreeTestNote(t, spelling.B(), cents.Zero(), nil)
	identity := func(n note.Note) (note.Note, error) { return n, nil }
	tests := []struct {
		name      string
		d         degree.Degree
		transform degree.Transform
		expected  error
	}{
		{"unset receiver", degree.Degree{}, identity, degree.ErrUnset},
		{"nil transform", degree.MustNewWithNote(4, valid), nil, degree.ErrUnset},
		{"nil transform before absent note", degree.MustNew(4), nil, degree.ErrUnset},
		{"absent note", degree.MustNew(4), identity, degree.ErrNoNote},
		{"unset result", degree.MustNewWithNote(4, valid), func(note.Note) (note.Note, error) {
			return note.Note{}, nil
		}, degree.ErrUnset},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := tt.d
			before := d
			calls := 0
			transform := tt.transform
			if transform != nil {
				original := transform
				transform = func(n note.Note) (note.Note, error) {
					calls++
					return original(n)
				}
			}
			err := d.TransformNote(transform)
			assertDegreeError(t, err, tt.expected)
			assertSameDegree(t, before, d)
			if tt.name == "unset result" {
				assert.Equal(t, 1, calls)
			} else {
				assert.Zero(t, calls)
			}
			assertMustPanic(t, err, func() { d.MustTransformNote(transform) })
			assertSameDegree(t, before, d)
		})
	}
}

func TestTransformNoteFailurePreservesCauseAndDegree(t *testing.T) {
	t.Parallel()
	n := degreeTestNote(t, spelling.B(), cents.Zero(), settings.Standard())
	d := degree.MustNewWithNote(4, n)
	before := d
	cause := errors.New("user transform failed")
	calls := 0
	transform := func(n note.Note) (note.Note, error) {
		calls++
		return note.Note{}, cause
	}
	err := d.TransformNote(transform)
	assertDegreeError(t, err, degree.ErrTransformFailed)
	assert.ErrorIs(t, err, cause)
	assertSameDegree(t, before, d)
	assert.True(t, before.Note().SameSource(d.Note()))
	assert.Equal(t, 1, calls)
	assertMustPanic(t, err, func() { d.MustTransformNote(transform) })
	assert.Equal(t, 2, calls)
	assertSameDegree(t, before, d)
}

func TestTransformNoteErrorWinsOverReturnedNote(t *testing.T) {
	t.Parallel()
	d := degree.MustNewWithNote(4, degreeTestNote(t, spelling.B(), cents.Zero(), nil))
	before := d
	cause := note.ErrNoOctave
	replacement := degreeTestNote(t, spelling.C(), cents.Zero(), nil)
	err := d.TransformNote(func(note.Note) (note.Note, error) {
		return replacement, cause
	})
	assertDegreeError(t, err, degree.ErrTransformFailed)
	assert.ErrorIs(t, err, cause)
	assert.NotErrorIs(t, err, degree.ErrUnset)
	assertSameDegree(t, before, d)
}

func TestTransformNotePanicPropagatesAndLeavesDegreeUnchanged(t *testing.T) {
	t.Parallel()
	d := degree.MustNewWithNote(4, degreeTestNote(t, spelling.B(), cents.Zero(), nil))
	before := d
	panicValue := errors.New("transform panic")
	assert.PanicsWithValue(t, panicValue, func() {
		_ = d.TransformNote(func(note.Note) (note.Note, error) {
			panic(panicValue)
		})
	})
	assertSameDegree(t, before, d)
}

func TestSetNumberPreservesNoteAndSource(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		d       degree.Degree
		hasNote bool
	}{
		{"without note", degree.MustNew(4), false},
		{"with note", degree.MustNewWithNote(4, degreeTestNote(t, spelling.B(), cents.Zero(), settings.Standard())), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := tt.d
			n := d.Note()
			require.NoError(t, d.SetNumber(5))
			assert.Equal(t, 5, d.Number())
			assert.Equal(t, tt.hasNote, d.HasNote())
			if tt.hasNote {
				assert.True(t, n.Equal(d.Note()))
				assert.True(t, n.SameSource(d.Note()))
			}
			returned := d.MustSetNumber(1)
			assert.Same(t, &d, returned)
			assert.Equal(t, 1, d.Number())
		})
	}
}

func TestSetNumberRejectsInvalidNumbersAtomically(t *testing.T) {
	t.Parallel()
	n := degreeTestNote(t, spelling.B(), cents.Zero(), settings.Standard())
	for _, number := range []int{-1, 0} {
		t.Run(fmt.Sprintf("number_%d", number), func(t *testing.T) {
			t.Parallel()
			d := degree.MustNewWithNote(4, n)
			before := d
			err := d.SetNumber(number)
			assertDegreeError(t, err, degree.ErrInvalidNumber)
			assertSameDegree(t, before, d)
			assert.True(t, before.Note().SameSource(d.Note()))
			assertMustPanic(t, err, func() { d.MustSetNumber(number) })
			assertSameDegree(t, before, d)
		})
	}
}
