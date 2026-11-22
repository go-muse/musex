package note_test

import (
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/notation"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func repeatNoteSet(t *testing.T, style notation.OctaveStyle) *settings.Settings {
	t.Helper()
	n := notation.MustNew(notation.Definition{Rules: notation.MustBuiltIn(notation.EnglishUnicodeID).NameRules(), OctaveStyle: style})
	return settings.Standard().MustWithNotation(n)
}

func repeatNoteError(t *testing.T, err error) {
	t.Helper()
	require.ErrorIs(t, err, notation.ErrOctaveRepeatLimit)
	for _, other := range []error{note.ErrUnset, note.ErrNoSource, note.ErrUnsetSettings, note.ErrNoOctave, note.ErrInadmissible, note.ErrDifferentSystems, note.ErrInexpressible, note.ErrOutOfRange, settings.ErrInadmissible, notation.ErrParse} {
		assert.NotErrorIs(t, err, other)
	}
}

func repeatNotePanic(t *testing.T, f func()) {
	t.Helper()
	var value any
	func() { defer func() { value = recover() }(); f() }()
	require.NotNil(t, value)
	err, ok := value.(error)
	require.True(t, ok)
	repeatNoteError(t, err)
}

// K-4 and K-5: the public note API propagates the dependency error unchanged
// in classification, including Must wrappers and explicit-source operations.
func TestNoteOctaveRepeatBudget(t *testing.T) {
	for _, tc := range []struct {
		name                string
		style               notation.OctaveStyle
		w, direction, count int
		base, mark          string
	}{
		{"helmholtz lower", notation.OctaveStyleHelmholtz, -32766, -1, 32768, "C", "͵"},
		{"helmholtz upper", notation.OctaveStyleHelmholtz, 21848, 1, 21845, "c", "′"},
		{"english lower", notation.OctaveStyleHelmholtzEnglish, -32766, -1, 32768, "C", "͵"},
		{"english upper", notation.OctaveStyleHelmholtzEnglish, 21848, 1, 21845, "c", "′"},
		{"lilypond lower", notation.OctaveStyleLilyPond, -65533, -1, 65536, "c", ","},
		{"lilypond upper", notation.OctaveStyleLilyPond, 65539, 1, 65536, "c", "'"},
		{"abc lower", notation.OctaveStyleABC, -65532, -1, 65536, "C", ","},
		{"abc upper", notation.OctaveStyleABC, 65541, 1, 65536, "c", "'"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := repeatNoteSet(t, tc.style)
			for _, over := range []bool{false, true} {
				w, count := tc.w, tc.count
				if over {
					w += tc.direction
					count++
				}
				p := pitch.MustFromASPN(spelling.C(), w)
				n, err := note.FromPitch(p, cents.MustFromInt(7), s)
				require.NoError(t, err)
				before := n
				explicit, err := note.FromPitch(p, cents.Zero(), settings.Standard())
				require.NoError(t, err)
				input := tc.base + strings.Repeat(tc.mark, count)
				for _, format := range []func() (string, error){n.Format, func() (string, error) { return explicit.FormatIn(s) }} {
					got, err := format()
					if over {
						assert.Empty(t, got)
						repeatNoteError(t, err)
					} else {
						require.NoError(t, err)
						assert.Equal(t, input, got)
					}
				}
				for _, expresses := range []func() (bool, error){n.IsExpressible, func() (bool, error) { return explicit.IsExpressibleIn(s) }} {
					ok, err := expresses()
					require.NoError(t, err)
					assert.Equal(t, !over, ok)
				}
				assert.Equal(t, !over, n.MustIsExpressible())
				assert.Equal(t, !over, explicit.MustIsExpressibleIn(s))
				parsed, err := note.ParsePitch(s, input)
				if over {
					repeatNoteError(t, err)
					assert.Equal(t, note.Note{}, parsed)
					repeatNotePanic(t, func() { n.MustFormat() })
					repeatNotePanic(t, func() { explicit.MustFormatIn(s) })
					repeatNotePanic(t, func() { note.MustParsePitch(s, input) })
				} else {
					require.NoError(t, err)
					assert.True(t, parsed.MustPitch().Equal(p))
					assert.True(t, parsed.MustIsExpressible())
					assert.True(t, note.MustParsePitch(s, input).MustPitch().Equal(p))
					assert.Equal(t, input, n.MustFormat())
					assert.Equal(t, input, explicit.MustFormatIn(s))
				}
				assert.Equal(t, before, n, "queries do not mutate the note or its cents/source")
				assert.Equal(t, "C", n.MustFormatName())
				assert.Equal(t, "C", explicit.MustFormatNameIn(s))
				assert.NotEmpty(t, n.String())
			}
		})
	}
}

func TestNoteOctaveRepeatExtremeQueries(t *testing.T) {
	t.Parallel()
	for _, style := range []notation.OctaveStyle{notation.OctaveStyleHelmholtz, notation.OctaveStyleHelmholtzEnglish, notation.OctaveStyleLilyPond, notation.OctaveStyleABC} {
		t.Run(style.String(), func(t *testing.T) {
			t.Parallel()
			s := repeatNoteSet(t, style)
			for _, w := range []int{math.MinInt, math.MaxInt} {
				n, err := note.FromPitch(pitch.MustFromASPN(spelling.C(), w), cents.Zero(), s)
				require.NoError(t, err)
				ok, err := n.IsExpressible()
				require.NoError(t, err)
				assert.False(t, ok)
				assert.False(t, n.MustIsExpressible())
				ok, err = n.IsExpressibleIn(s)
				require.NoError(t, err)
				assert.False(t, ok)
				assert.False(t, n.MustIsExpressibleIn(s))
			}
		})
	}
}

func TestNoteOctaveRepeatSourceTransitions(t *testing.T) {
	numeric := settings.Standard()
	repeated := repeatNoteSet(t, notation.OctaveStyleLilyPond)
	p := pitch.MustFromASPN(spelling.C(), 65540)
	for _, written := range []bool{false, true} {
		t.Run(strconv.FormatBool(written), func(t *testing.T) {
			n := note.MustFromPitch(p, cents.Zero(), numeric)
			before := n
			check, err := n.CheckSetSource(repeated)
			if written {
				check, err = n.CheckSetSourceWritten(repeated)
			}
			require.NoError(t, err)
			assert.False(t, check.Possible())
			assert.ErrorIs(t, check.Err(), note.ErrInexpressible)
			assert.Equal(t, before, n)
			if written {
				err = n.SetSourceWritten(repeated)
			} else {
				err = n.SetSource(repeated)
			}
			assert.ErrorIs(t, err, note.ErrInexpressible)
			assert.Equal(t, before, n)

			// A previously unrenderable pitch does not introduce a new transition rule.
			n = note.MustFromPitch(p, cents.Zero(), repeated)
			if written {
				err = n.SetSourceWritten(numeric)
			} else {
				err = n.SetSource(numeric)
			}
			require.NoError(t, err)
			assert.True(t, n.MustIsExpressible())
		})
	}
}

func TestNoteOctaveRepeatParsingAndOrdinaryErrors(t *testing.T) {
	s := settings.Standard().MustWithNotation(notation.MustBuiltIn(notation.LilyPondEnglishID))
	input := "cqs" + strings.Repeat("'", 65537)
	n, err := note.ParsePitch(s, input)
	repeatNoteError(t, err)
	assert.Equal(t, note.Note{}, n)
	for _, tc := range []struct {
		name string
		src  settings.Source
		want error
	}{
		{"no source", nil, note.ErrUnset},
		{"unset settings", new(settings.Settings), note.ErrUnsetSettings},
	} {
		t.Run(tc.name, func(t *testing.T) {
			n, err := note.ParsePitch(tc.src, input)
			assert.Equal(t, note.Note{}, n)
			assert.ErrorIs(t, err, tc.want)
			assert.NotErrorIs(t, err, notation.ErrOctaveRepeatLimit)
		})
	}
	unbound := note.MustFromPitch(pitch.MustFromASPN(spelling.C(), 65540), cents.Zero(), nil)
	text, err := unbound.Format()
	assert.Empty(t, text)
	assert.ErrorIs(t, err, note.ErrNoSource)
	assert.NotErrorIs(t, err, notation.ErrOctaveRepeatLimit)
	text, err = (note.Note{}).FormatIn(s)
	assert.Empty(t, text)
	assert.ErrorIs(t, err, note.ErrUnset)
	assert.NotErrorIs(t, err, notation.ErrOctaveRepeatLimit)
	abc := settings.Standard().MustWithNotation(notation.MustBuiltIn(notation.ABCID))
	for _, input := range []string{"C" + strings.Repeat("'", 65537), "C" + strings.Repeat(",'", 32769)} {
		parsed, err := note.ParsePitch(abc, input)
		require.NoError(t, err)
		assert.True(t, parsed.MustIsExpressible())
	}
}
