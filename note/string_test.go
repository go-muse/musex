package note_test

import (
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
)

// The debug string (architecture 6.12; OP-53; K-31): "<record>, <cents>,
// <bound|unbound>", the record by pitch.Pitch.String or
// spelling.Spelling.String, the cents by Text(WithUnit|WithPlus);
// "Note(unset)" for the unset note.

func TestString(t *testing.T) {
	t.Parallel()

	opts := cents.WithUnit | cents.WithPlus
	third := cents.MustFromFraction(1, 3)
	tests := []struct {
		name   string
		input  spelling.Spelling
		aspn   int
		hasOct bool
		c      cents.Cents
		bound  bool
	}{
		{"K-31 C♯4 +30 bound", spelling.CSharp(), 4, true, ct(30), true},
		{"no octave unbound", spelling.DFlat(), 0, false, ct(0), false},
		{"fractional cents", sp(letter.E, 1, 2), -1, true, third, false},
		{"negative cents bound", spelling.B(), 2, true, ct(-7), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			src := newCounting(std())
			var s settings.Source
			if tt.bound {
				s = src
			}
			var n note.Note
			record := tt.input.String()
			if tt.hasOct {
				n = mk(t, tt.input, tt.aspn, tt.c, s)
				record = pt(tt.input, tt.aspn).String()
			} else {
				n = mkName(t, tt.input, tt.c, s)
			}
			link := "unbound"
			if tt.bound {
				link = "bound"
			}
			before := src.reads()
			assert.Equal(t, record+", "+tt.c.Text(opts)+", "+link, n.String())
			assert.Equal(t, before, src.reads(), "String does not read the source")
		})
	}

	t.Run("K-31 example", func(t *testing.T) {
		t.Parallel()
		n := mk(t, spelling.CSharp(), 4, ct(30), std())
		assert.Equal(t, "C+1 ASPN 4, +30 cents, bound", n.String())
		m := mk(t, spelling.CSharp(), 4, ct(30), nil)
		assert.True(t, n.Equal(m), "the string does not affect equality")
	})

	t.Run("unset", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "Note(unset)", note.Note{}.String())
	})
}
