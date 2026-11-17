package degree_test

import (
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/degree"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoteReadsReturnIndependentCopies(t *testing.T) {
	t.Parallel()
	n := degreeTestNote(t, spelling.B(), cents.Zero(), settings.Standard())
	d := degree.MustNewWithNote(4, n)
	first := d.Note()
	second := d.Note()
	require.True(t, first.IsValid())
	require.True(t, second.IsValid())
	assert.True(t, first.Equal(second))
	assert.True(t, first.SameSource(second))

	first.MustLower()
	assertDegreeState(t, d, 4, true, n)
	assert.True(t, n.Equal(second))
	second.MustRaise()
	assertDegreeState(t, d, 4, true, n)
}

func TestDegreeAssignmentCopiesAreIndependentInBothDirections(t *testing.T) {
	t.Parallel()
	correction := cents.MustFromInt(30)
	tests := []struct {
		name       string
		note       note.Note
		sameSource bool
	}{
		{"without octave", degreeTestNote(t, spelling.B(), cents.Zero(), nil), false},
		{"with octave", degreeTestASPNNote(t, spelling.D(), 4, cents.Zero(), nil), false},
		{"with correction", degreeTestNote(t, spelling.B(), correction, nil), false},
		{"unbound", degreeTestASPNNote(t, spelling.A(), 4, cents.Zero(), nil), false},
		{"bound", degreeTestASPNNote(t, spelling.C(), 4, correction, settings.Standard()), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			original := degree.MustNewWithNote(4, tt.note)
			clone := original
			originalNote := original.Note()
			assert.Equal(t, tt.sameSource, originalNote.SameSource(clone.Note()))

			clone.MustSetNote(degreeTestNote(t, spelling.C(), cents.Zero(), nil))
			clone.MustSetNumber(5)
			assertDegreeState(t, original, 4, true, originalNote)
			assert.Equal(t, 5, clone.Number())

			original.MustRemoveNote()
			original.MustSetNumber(2)
			assertDegreeState(t, original, 2, false, note.Note{})
			assert.Equal(t, 5, clone.Number())
			assert.True(t, clone.HasNote())
		})
	}
}

func TestPredefinedAndCircleValuesAreIndependent(t *testing.T) {
	t.Parallel()
	first := degree.Second()
	second := degree.Second()
	first.MustSetNote(degreeTestNote(t, spelling.D(), cents.Zero(), nil))
	assertDegreeState(t, second, 2, false, note.Note{})

	sequence := degree.MustAll(2)
	var yielded []degree.Degree
	for d := range sequence {
		yielded = append(yielded, d)
	}
	require.Len(t, yielded, 2)
	yielded[0].MustSetNumber(53)
	yielded[1].MustSetNote(degreeTestNote(t, spelling.B(), cents.Zero(), nil))
	assertDegreeState(t, degree.First(), 1, false, note.Note{})
	assertDegreeState(t, degree.Second(), 2, false, note.Note{})

	var repeated []degree.Degree
	for d := range sequence {
		repeated = append(repeated, d)
	}
	require.Len(t, repeated, 2)
	assertDegreeState(t, repeated[0], 1, false, note.Note{})
	assertDegreeState(t, repeated[1], 2, false, note.Note{})
}
