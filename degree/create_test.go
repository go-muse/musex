package degree_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/degree"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()
	for _, number := range []int{1, 2, 7, 8, 53, math.MaxInt} {
		t.Run(fmt.Sprintf("number_%d", number), func(t *testing.T) {
			t.Parallel()
			d, err := degree.New(number)
			require.NoError(t, err)
			assertDegreeState(t, d, number, false, note.Note{})
			must := degree.MustNew(number)
			assertSameDegree(t, d, must)
		})
	}
}

func TestNewRejectsInvalidNumbers(t *testing.T) {
	t.Parallel()
	for _, number := range []int{math.MinInt, -1, 0} {
		t.Run(fmt.Sprintf("number_%d", number), func(t *testing.T) {
			t.Parallel()
			d, err := degree.New(number)
			assertDegreeError(t, err, degree.ErrInvalidNumber)
			assertUnsetDegree(t, d)
			_ = assertMustPanic(t, degree.ErrInvalidNumber, func() { degree.MustNew(number) })
		})
	}
}

func TestNewWithNoteAcceptsEveryNoteFormAndCopiesInput(t *testing.T) {
	t.Parallel()
	correction := cents.MustFromInt(30)
	bound := degreeTestASPNNote(t, spelling.C(), 4, correction, settings.Standard())
	tests := []struct {
		name       string
		note       note.Note
		sameSource bool
	}{
		{"without octave", degreeTestNote(t, spelling.B(), cents.Zero(), nil), false},
		{"with octave", degreeTestASPNNote(t, spelling.D(), 4, cents.Zero(), nil), false},
		{"with correction", degreeTestNote(t, spelling.B(), correction, nil), false},
		{"unbound", degreeTestASPNNote(t, spelling.A(), 4, cents.Zero(), nil), false},
		{"bound", bound, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := tt.note
			d, err := degree.NewWithNote(4, input)
			require.NoError(t, err)
			assertDegreeState(t, d, 4, true, input)
			must := degree.MustNewWithNote(4, input)
			assertSameDegree(t, d, must)

			original := input
			input.MustRaise()
			assertDegreeState(t, d, 4, true, original)
			assert.Equal(t, tt.sameSource, original.SameSource(d.Note()))
		})
	}
}

func TestNewWithNoteRejectsUnsetNote(t *testing.T) {
	t.Parallel()
	d, err := degree.NewWithNote(4, note.Note{})
	assertDegreeError(t, err, degree.ErrUnset)
	assertUnsetDegree(t, d)
	_ = assertMustPanic(t, degree.ErrUnset, func() { degree.MustNewWithNote(4, note.Note{}) })
}

func TestPredefinedDegrees(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		number int
		new    func() degree.Degree
	}{
		{"First", 1, degree.First},
		{"Second", 2, degree.Second},
		{"Third", 3, degree.Third},
		{"Fourth", 4, degree.Fourth},
		{"Fifth", 5, degree.Fifth},
		{"Sixth", 6, degree.Sixth},
		{"Seventh", 7, degree.Seventh},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			first := tt.new()
			assertDegreeState(t, first, tt.number, false, note.Note{})
			first.MustSetNote(degreeTestNote(t, spelling.D(), cents.Zero(), nil))
			first.MustSetNumber(53)
			second := tt.new()
			assertDegreeState(t, second, tt.number, false, note.Note{})
		})
	}
}

func TestDegreeString(t *testing.T) {
	t.Parallel()
	assertUnsetDegree(t, degree.Degree{})
	without := degree.MustNew(4)
	assert.Equal(t, "Degree(4, no note)", without.String())
	n := degreeTestASPNNote(t, spelling.C(), 4, cents.MustFromInt(30), settings.Standard())
	n.MustRaise()
	with := degree.MustNewWithNote(4, n)
	assert.Equal(t, "Degree(4, "+n.String()+")", with.String())
}
