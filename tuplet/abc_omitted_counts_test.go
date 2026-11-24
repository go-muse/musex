package tuplet_test

import (
	"strconv"
	"testing"

	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The tests in this file cover the ABC marker "(p::" with both the normal
// count q and the event count r omitted (issue #158). The abc 2.1 standard,
// section 4.13, names "(3::" as equivalent to "(3", "(3:2", and "(3:2:3", so
// "(p::" reads exactly as "(p" in the same division: q takes the ABC default
// of p for that division and r equals p.

// abcDefaultNormal is the ABC 2.1 default normal count q of "(p" for p = 2
// through 9, by division: Binary is the context of a simple meter, Ternary of
// a compound meter (domain knowledge topic tuplet, section 6, column
// "abc 2.1"; specification tuplet, P-9 and KP-18).
var abcDefaultNormal = map[tuplet.Division]map[int]int{
	tuplet.Binary:  {2: 3, 3: 2, 4: 3, 5: 2, 6: 2, 7: 2, 8: 3, 9: 2},
	tuplet.Ternary: {2: 3, 3: 2, 4: 3, 5: 3, 6: 2, 7: 3, 8: 3, 9: 3},
}

// "(p::" reads exactly as "(p" in the same division for every p with an ABC
// default, gives the default tuplet over p events, and formats back to the
// shortest form "(p".
func TestParseABCMarker_BothCountsOmitted(t *testing.T) {
	t.Parallel()

	for _, division := range []tuplet.Division{tuplet.Binary, tuplet.Ternary} {
		for p := 2; p <= 9; p++ {
			short := "(" + strconv.Itoa(p)
			text := short + "::"

			t.Run(division.String()+"/"+text, func(t *testing.T) {
				t.Parallel()

				want, err := tuplet.ParseABCMarker(short, division)
				require.NoError(t, err, "the short form %q is the reference", short)

				got, err := tuplet.ParseABCMarker(text, division)
				require.NoError(t, err)
				assert.Equal(t, want, got, "%q reads as %q", text, short)

				assert.Equal(t, tuplet.MustNew(p, abcDefaultNormal[division][p]), got.Tuplet)
				assert.Equal(t, p, got.Events, "the event count defaults to p")

				// The formatter keeps writing the shortest form, and parsing
				// it in the same division returns the same marker.
				formatted, err := tuplet.FormatABCMarker(got, division)
				require.NoError(t, err)
				assert.Equal(t, short, formatted)

				back, err := tuplet.ParseABCMarker(formatted, division)
				require.NoError(t, err)
				assert.Equal(t, got, back)
			})
		}
	}
}

// The cases named by issue #158, with explicit expectations: "(3::" is 3:2
// over 3 events in both divisions; "(5::" is 5:2 in a simple meter and 5:3 in
// a compound meter, over 5 events. Every spelling of the same marker with
// omitted parts reads as the same marker.
func TestParseABCMarker_BothCountsOmittedIssueCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		text     string
		division tuplet.Division
		want     tuplet.Tuplet
		events   int
		same     []string
	}{
		{
			name:     "triplet simple meter",
			text:     "(3::",
			division: tuplet.Binary,
			want:     tuplet.Triplet(),
			events:   3,
			same:     []string{"(3", "(3:2", "(3:2:3", "(3::3"},
		},
		{
			name:     "triplet compound meter",
			text:     "(3::",
			division: tuplet.Ternary,
			want:     tuplet.Triplet(),
			events:   3,
			same:     []string{"(3", "(3:2", "(3:2:3", "(3::3"},
		},
		{
			name:     "quintuplet simple meter",
			text:     "(5::",
			division: tuplet.Binary,
			want:     tuplet.MustNew(5, 2),
			events:   5,
			same:     []string{"(5", "(5:2", "(5:2:5", "(5::5"},
		},
		{
			name:     "quintuplet compound meter",
			text:     "(5::",
			division: tuplet.Ternary,
			want:     tuplet.MustNew(5, 3),
			events:   5,
			same:     []string{"(5", "(5:3", "(5:3:5", "(5::5"},
		},
		{
			name:     "duplet compound meter",
			text:     "(2::",
			division: tuplet.Ternary,
			want:     tuplet.Duplet(),
			events:   2,
			same:     []string{"(2", "(2:3", "(2:3:2", "(2::2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tuplet.ParseABCMarker(tt.text, tt.division)
			require.NoError(t, err)
			assert.Equal(t, tuplet.ABCMarker{Tuplet: tt.want, Events: tt.events}, got)

			for _, other := range tt.same {
				m, err := tuplet.ParseABCMarker(other, tt.division)
				if assert.NoError(t, err, other) {
					assert.Equal(t, got, m, "%q and %q are the same marker", tt.text, other)
				}
			}

			assert.Equal(t, got, tuplet.MustParseABCMarker(tt.text, tt.division), "the Must variant returns the same marker")
		})
	}
}

// "(p::" fails exactly as "(p" when no default exists: a p without an ABC
// default and an unknown division give ErrNoDefault, a p beyond math.MaxInt
// gives ErrOutOfRange. The grammar accepts the text, so the error is not
// ErrInvalidText.
func TestParseABCMarker_BothCountsOmittedNoDefault(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		p        string
		division tuplet.Division
		wantErr  error
	}{
		{"p 1 simple meter", "1", tuplet.Binary, tuplet.ErrNoDefault},
		{"p 1 compound meter", "1", tuplet.Ternary, tuplet.ErrNoDefault},
		{"p 10 simple meter", "10", tuplet.Binary, tuplet.ErrNoDefault},
		{"p 10 compound meter", "10", tuplet.Ternary, tuplet.ErrNoDefault},
		{"p 11 simple meter", "11", tuplet.Binary, tuplet.ErrNoDefault},
		{"p 99 compound meter", "99", tuplet.Ternary, tuplet.ErrNoDefault},
		{"zero division", "3", tuplet.Division(0), tuplet.ErrNoDefault},
		{"unknown division", "3", tuplet.Division(9), tuplet.ErrNoDefault},
		{"unknown division p 5", "5", tuplet.Division(3), tuplet.ErrNoDefault},
		{"p beyond MaxInt", "99999999999999999999", tuplet.Binary, tuplet.ErrOutOfRange},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			short := "(" + tt.p
			text := short + "::"

			_, shortErr := tuplet.ParseABCMarker(short, tt.division)
			require.ErrorIs(t, shortErr, tt.wantErr, "the short form %q is the reference", short)

			_, err := tuplet.ParseABCMarker(text, tt.division)
			require.Error(t, err)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.NotErrorIs(t, err, tuplet.ErrInvalidText, "%q matches the grammar", text)
			assert.Contains(t, err.Error(), strconv.Quote(text), "the error quotes the short input in full")

			assert.Panics(t, func() { tuplet.MustParseABCMarker(text, tt.division) })
		})
	}
}

// The forms "(p:" and "(p:q:", a fourth part, and other malformed spellings
// around the empty parts stay rejected as text that does not match the
// grammar, in both divisions.
func TestParseABCMarker_OmittedCountsStillRejected(t *testing.T) {
	t.Parallel()

	texts := []string{
		"(3:",
		"(5:",
		"(3:2:",
		"(5:2:",
		"(10:3:",
		"(3:2:2:2",
		"(3::2:2",
		"(3::2:",
		"(3:::",
		"(3:::2",
		"(3::0",
		"(3::02",
		"(0::",
		"(03::",
		"(::",
		"(::3",
		"(3:: ",
		" (3::",
		"(3: :",
		"3::",
	}

	for _, division := range []tuplet.Division{tuplet.Binary, tuplet.Ternary} {
		for _, text := range texts {
			t.Run(division.String()+"/"+strconv.Quote(text), func(t *testing.T) {
				t.Parallel()

				_, err := tuplet.ParseABCMarker(text, division)
				require.Error(t, err)
				assert.ErrorIs(t, err, tuplet.ErrInvalidText)
				assert.Contains(t, err.Error(), strconv.Quote(text), "the error quotes the short input in full")
			})
		}
	}
}
