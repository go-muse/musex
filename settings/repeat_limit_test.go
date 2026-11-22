package settings_test

import (
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/notation"
	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/octave/number"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"
	"github.com/go-muse/musex/tuning/temperament"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func repeatSet(t *testing.T, style notation.OctaveStyle, forms ...notation.NameForm) *settings.Settings {
	t.Helper()
	rules := notation.MustBuiltIn(notation.EnglishUnicodeID).NameRules()
	if len(forms) != 0 {
		rules = notation.MustNewNameRules(notation.NameRulesDefinition{Names: forms})
	}
	n := notation.MustNew(notation.Definition{Rules: rules, OctaveStyle: style})
	return settings.Standard().MustWithNotation(n)
}

func repeatSetPanic(t *testing.T, f func()) {
	t.Helper()
	var value any
	func() { defer func() { value = recover() }(); f() }()
	require.NotNil(t, value)
	err, ok := value.(error)
	require.True(t, ok)
	assert.ErrorIs(t, err, notation.ErrOctaveRepeatLimit)
}

// K-1, K-2, K-4, K-5: same boundary for recognition, rendering, and parsing.
func TestPitchOctaveRepeatBudget(t *testing.T) {
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
			s := repeatSet(t, tc.style)
			before := s.MustWithNotation(s.Notation())
			for _, over := range []bool{false, true} {
				w, count := tc.w, tc.count
				if over {
					w += tc.direction
					count++
				}
				p := pitch.MustFromASPN(spelling.C(), w)
				input := tc.base + strings.Repeat(tc.mark, count)
				ok, err := s.ExpressesPitch(p)
				require.NoError(t, err)
				assert.Equal(t, !over, ok)
				assert.Equal(t, !over, s.MustExpressesPitch(p))
				text, err := s.FormatPitch(p)
				if over {
					assert.Empty(t, text)
					require.ErrorIs(t, err, notation.ErrOctaveRepeatLimit)
					assert.NotErrorIs(t, err, settings.ErrInadmissible)
					repeatSetPanic(t, func() { s.MustFormatPitch(p) })
				} else {
					require.NoError(t, err)
					assert.Equal(t, input, text)
					assert.Equal(t, input, s.MustFormatPitch(p))
				}
				back, err := s.ParsePitch(input)
				if over {
					assert.Equal(t, pitch.Pitch{}, back)
					require.ErrorIs(t, err, notation.ErrOctaveRepeatLimit)
					assert.NotErrorIs(t, err, settings.ErrInadmissible)
					repeatSetPanic(t, func() { s.MustParsePitch(input) })
				} else {
					require.NoError(t, err)
					assert.True(t, p.Equal(back))
					assert.True(t, s.MustExpressesPitch(back))
					assert.True(t, s.MustParsePitch(input).Equal(p))
				}
			}
			assert.True(t, before.Equal(s))
		})
	}
}

// K-3: no byte subtraction for long canonical names or aliases in ExpressesPitch.
func TestPitchOctaveRepeatLongNames(t *testing.T) {
	for _, style := range []notation.OctaveStyle{notation.OctaveStyleHelmholtz, notation.OctaveStyleHelmholtzEnglish, notation.OctaveStyleLilyPond, notation.OctaveStyleABC} {
		t.Run(style.String(), func(t *testing.T) {
			name := "É" + strings.Repeat("x", 65537)
			alias := "Ö" + strings.Repeat("y", 65537)
			s := repeatSet(t, style, notation.NameForm{Name: spelling.C(), Text: name, Aliases: []string{alias}})
			w, mark, count := 21848, "′", 21845
			if style == notation.OctaveStyleLilyPond {
				w, mark, count = 65539, "'", 65536
			}
			if style == notation.OctaveStyleABC {
				w, mark, count = 65541, "'", 65536
			}
			p := pitch.MustFromASPN(spelling.C(), w)
			assert.True(t, s.MustExpressesPitch(p))
			got, err := s.FormatPitch(p)
			require.NoError(t, err)
			assert.Equal(t, "é"+name[len("É"):]+strings.Repeat(mark, count), got)
			back, err := s.ParsePitch("ö" + alias[len("Ö"):] + strings.Repeat(mark, count))
			require.NoError(t, err)
			assert.True(t, back.Equal(p))
			assert.True(t, s.MustExpressesPitch(back))
			assert.True(t, s.MustExpresses(spelling.C()))
			assert.Equal(t, name, s.MustFormat(spelling.C()))
		})
	}
}

// K-5 and K-6: these queries do not generate octave strings, even at int limits.
func TestPitchOctaveRepeatExtremeExpressibility(t *testing.T) {
	t.Parallel()
	for _, style := range []notation.OctaveStyle{notation.OctaveStyleHelmholtz, notation.OctaveStyleHelmholtzEnglish, notation.OctaveStyleLilyPond, notation.OctaveStyleABC} {
		t.Run(style.String(), func(t *testing.T) {
			t.Parallel()
			s := repeatSet(t, style)
			for _, w := range []int{math.MinInt, math.MaxInt} {
				p := pitch.MustFromASPN(spelling.C(), w)
				ok, err := s.ExpressesPitch(p)
				require.NoError(t, err)
				assert.False(t, ok)
				assert.False(t, s.MustExpressesPitch(p))
			}
		})
	}
}

// The written octave, rather than pitch's octave, determines each budget.
func TestPitchOctaveRepeatSystemCarry(t *testing.T) {
	for _, tc := range []struct {
		name   spelling.Spelling
		octave int
		sys    *system.System
		want   bool
	}{
		{spelling.BSharp(), 65540, system.Traditional(), true},
		{spelling.BSharp(), 65541, system.Traditional(), false},
		{spelling.BSharp(), 65540, system.Positions19(), false},
		{spelling.CFlat(), -65534, system.Traditional(), true},
		{spelling.CFlat(), -65535, system.Traditional(), false},
	} {
		t.Run(tc.name.String()+"/"+strconv.Itoa(tc.octave)+"/"+tc.sys.Name(), func(t *testing.T) {
			s := repeatSet(t, notation.OctaveStyleLilyPond).MustWithTuning(tuningFor(t, tc.sys))
			p := pitch.MustFromASPN(tc.name, tc.octave)
			assert.Equal(t, tc.want, s.MustExpressesPitch(p))
			text, err := s.FormatPitch(p)
			if !tc.want {
				assert.Empty(t, text)
				assert.ErrorIs(t, err, notation.ErrOctaveRepeatLimit)
				return
			}
			require.NoError(t, err)
			back, err := s.ParsePitch(text)
			require.NoError(t, err)
			assert.True(t, back.Equal(p))
		})
	}
}

func TestPitchOctaveRepeatParseOrderAndABC(t *testing.T) {
	lily := settings.Standard().MustWithNotation(notation.MustBuiltIn(notation.LilyPondEnglishID))
	input := "cqs" + strings.Repeat("'", 65537)
	back, err := lily.ParsePitch(input)
	require.ErrorIs(t, err, notation.ErrOctaveRepeatLimit)
	assert.NotErrorIs(t, err, settings.ErrInadmissible)
	assert.Equal(t, pitch.Pitch{}, back)
	text, err := lily.FormatPitch(pitch.MustFromASPN(spelling.C().WithAlteration(alteration.MustFromFraction(1, 2)), 65540))
	assert.Empty(t, text)
	assert.ErrorIs(t, err, settings.ErrInadmissible)
	assert.NotErrorIs(t, err, notation.ErrOctaveRepeatLimit)
	var zero *settings.Settings
	text, err = zero.FormatPitch(pitch.MustFromASPN(spelling.C(), 65540))
	assert.Empty(t, text)
	assert.ErrorIs(t, err, settings.ErrUnset)
	back, err = zero.ParsePitch(input)
	assert.Equal(t, pitch.Pitch{}, back)
	assert.ErrorIs(t, err, settings.ErrUnset)
	abc := settings.Standard().MustWithNotation(notation.MustBuiltIn(notation.ABCID))
	for _, input := range []string{"C" + strings.Repeat("'", 65537), "C" + strings.Repeat(",'", 32769)} {
		back, err := abc.ParsePitch(input)
		require.NoError(t, err)
		assert.True(t, abc.MustExpressesPitch(back))
	}
}

// K-7: compatibility compares musical domains, not rendering resource ranges.
func TestOctaveRepeatSemanticCompatibility(t *testing.T) {
	numeric := settings.Standard()
	for _, style := range []notation.OctaveStyle{notation.OctaveStyleHelmholtz, notation.OctaveStyleHelmholtzEnglish, notation.OctaveStyleLilyPond, notation.OctaveStyleABC} {
		t.Run(style.String(), func(t *testing.T) {
			repeated := repeatSet(t, style)
			p := pitch.MustFromASPN(spelling.C(), 65542)
			assert.True(t, numeric.MustExpressesPitch(p))
			assert.False(t, repeated.MustExpressesPitch(p))
			for _, pair := range [][2]*settings.Settings{{numeric, repeated}, {repeated, numeric}} {
				report, err := pair[0].CheckReplacement(pair[1])
				require.NoError(t, err)
				assert.True(t, report.Compatible())
				assert.True(t, pair[0].MustCheckReplacement(pair[1]).Compatible())
				report, err = pair[0].CheckDesignations(pair[1])
				require.NoError(t, err)
				assert.True(t, report.Compatible())
				assert.True(t, pair[0].MustCheckDesignations(pair[1]).Compatible())
			}
		})
	}
}

// A singleton name with a large carry makes the semantic witness technically
// unrenderable in the old set, without constructing a large octave catalog.
func TestOctaveRepeatSemanticWitnessOutsideBudget(t *testing.T) {
	name := spelling.C().WithAlteration(alteration.MustFromInt(1 << 20))
	prev := repeatSet(t, notation.OctaveStyleLilyPond, notation.NameForm{Name: name, Text: "C"})
	nextSystem := system.MustNew(system.Definition{
		Classes: 12, Boundary: letter.C, Offsets: [letter.Count]int{0, 2, 4, 5, 7, 9, 11},
		Chromatic: 2, Alterations: system.Integers(),
	})
	tu := tuning.MustNew(nextSystem, temperament.Equal12(), tuning.RootFromName(spelling.C()), tuning.A4At440())
	next := prev.MustWithTuning(tu).MustWithNotation(notation.MustNew(notation.Definition{Rules: prev.Notation().NameRules(), OctaveStyle: notation.OctaveStyleAM}))
	for _, report := range []settings.Report{prev.MustCheckReplacement(next), prev.MustCheckDesignations(next)} {
		assert.False(t, report.Compatible())
		v, found := report.Violation(settings.RuleDesignations)
		require.True(t, found)
		assert.True(t, v.Pitch.IsValid())
		assert.False(t, v.Name.IsValid())
		expected := pitch.MustFromASPN(name, 9+174762)
		assert.True(t, v.Pitch.Equal(expected))
		assert.False(t, prev.MustExpressesPitch(v.Pitch), "the witness expresses semantic loss, not a promise to render")
		assert.False(t, next.MustExpressesPitch(v.Pitch))
		assert.ErrorIs(t, report.Err(), settings.ErrIncompatible)
	}
}

// Number, Name and finite AM still answer by their own unchanged conditions.
func TestOctaveRepeatUnaffectedStyles(t *testing.T) {
	for _, tc := range []struct {
		style notation.OctaveStyle
		w     int
		want  bool
	}{
		{notation.OctaveStyleNumber, math.MaxInt, true},
		{notation.OctaveStyleHelmholtzNumbered, math.MaxInt, true},
		{notation.OctaveStyleAM, 8, true},
		{notation.OctaveStyleAM, 9, false},
	} {
		s := repeatSet(t, tc.style)
		assert.Equal(t, tc.want, s.MustExpressesPitch(pitch.MustFromASPN(spelling.C(), tc.w)))
	}
	named := settings.Standard().MustWithNotation(notation.MustNew(notation.Definition{Rules: notation.MustBuiltIn(notation.EnglishUnicodeID).NameRules(), OctaveStyle: notation.OctaveStyleName, OctaveSeparator: " "}))
	named = named.MustWithNumbering(number.ASPN()).MustWithCatalog(naming.MustNew(map[int]string{65542: "high"}))
	assert.True(t, named.MustExpressesPitch(pitch.MustFromASPN(spelling.C(), 65542)))
	assert.Equal(t, "C high", named.MustFormatPitch(pitch.MustFromASPN(spelling.C(), 65542)))
}
