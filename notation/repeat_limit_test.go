package notation_test

import (
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/notation"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/octave/number"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These regression tests must only execute after the repeat guard is implemented.
// All generated strings are bounded to approximately 64 KiB; int extremes are
// exercised only by short numeric formats, never by repeated formats.
type repeatBoundary struct {
	name       string
	style      notation.OctaveStyle
	written    int
	base, mark string
	count      int
}

func repeatBoundaries() []repeatBoundary {
	return []repeatBoundary{
		{"helmholtz lower", notation.OctaveStyleHelmholtz, 2 - 32768, "C", "͵", 32768},
		{"helmholtz upper", notation.OctaveStyleHelmholtz, 3 + 21845, "c", "′", 21845},
		{"english lower", notation.OctaveStyleHelmholtzEnglish, 2 - 32768, "C", "͵", 32768},
		{"english upper", notation.OctaveStyleHelmholtzEnglish, 3 + 21845, "c", "′", 21845},
		{"lilypond lower", notation.OctaveStyleLilyPond, 3 - 65536, "c", ",", 65536},
		{"lilypond upper", notation.OctaveStyleLilyPond, 3 + 65536, "c", "'", 65536},
		{"abc lower", notation.OctaveStyleABC, 4 - 65536, "C", ",", 65536},
		{"abc upper", notation.OctaveStyleABC, 5 + 65536, "c", "'", 65536},
	}
}

func repeatNotation(t *testing.T, style notation.OctaveStyle, forms ...notation.NameForm) notation.Notation {
	t.Helper()
	rules := notation.MustBuiltIn(notation.EnglishUnicodeID).NameRules()
	if len(forms) != 0 {
		var err error
		rules, err = notation.NewNameRules(notation.NameRulesDefinition{Names: forms})
		require.NoError(t, err)
	}
	n, err := notation.New(notation.Definition{Rules: rules, OctaveStyle: style})
	require.NoError(t, err)
	return n
}

func repeatPanic(t *testing.T, target error, f func()) {
	t.Helper()
	var recovered any
	func() { defer func() { recovered = recover() }(); f() }()
	require.NotNil(t, recovered, "Must wrapper must panic")
	err, ok := recovered.(error)
	require.True(t, ok, "panic must contain an error")
	assert.ErrorIs(t, err, target)
}

func repeatError(t *testing.T, err error) {
	t.Helper()
	require.ErrorIs(t, err, notation.ErrOctaveRepeatLimit)
	for _, other := range []error{notation.ErrUndefined, notation.ErrParse, notation.ErrNoOctave, notation.ErrUnwritableName, notation.ErrIncompatibleSystem, octave.ErrOutOfRange, number.ErrOutOfRange, system.ErrOutOfRange} {
		assert.NotErrorIs(t, err, other)
	}
}

// K-1, K-2 and K-5: exact canonical bytes on both sides of all four styles.
func TestOctaveRepeatBoundaries(t *testing.T) {
	require.Equal(t, 65536, notation.MaxOctaveRepeatBytes)
	for _, tc := range repeatBoundaries() {
		t.Run(tc.name, func(t *testing.T) {
			n := repeatNotation(t, tc.style)
			if tc.style == notation.OctaveStyleLilyPond {
				n = notation.MustBuiltIn(notation.LilyPondEnglishID)
			}
			if tc.style == notation.OctaveStyleABC {
				n = notation.MustBuiltIn(notation.ABCID)
			}
			before := n
			p := pitch.MustFromASPN(spelling.C(), tc.written)
			expected := tc.base + strings.Repeat(tc.mark, tc.count)
			text, err := n.FormatPitch(p, system.Traditional(), notation.OctaveContext{})
			require.NoError(t, err)
			assert.Equal(t, expected, text)
			assert.Len(t, text, len(tc.base)+len(tc.mark)*tc.count)
			assert.Equal(t, expected, n.MustFormatPitch(p, system.Traditional(), notation.OctaveContext{}))
			back, err := n.ParsePitch(expected, system.Traditional(), notation.OctaveContext{})
			require.NoError(t, err)
			assert.True(t, back.Equal(p))
			assert.True(t, n.MustParsePitch(expected, system.Traditional(), notation.OctaveContext{}).Equal(p))

			w := tc.written + 1
			if tc.mark == "," || tc.mark == "͵" {
				w = tc.written - 1
			}
			beyond := pitch.MustFromASPN(spelling.C(), w)
			text, err = n.FormatPitch(beyond, system.Traditional(), notation.OctaveContext{})
			repeatError(t, err)
			assert.Empty(t, text)
			repeatPanic(t, notation.ErrOctaveRepeatLimit, func() { n.MustFormatPitch(beyond, system.Traditional(), notation.OctaveContext{}) })
			input := expected + tc.mark
			back, err = n.ParsePitch(input, system.Traditional(), notation.OctaveContext{})
			repeatError(t, err)
			assert.Equal(t, pitch.Pitch{}, back)
			assert.NotContains(t, err.Error(), input)
			repeatPanic(t, notation.ErrOctaveRepeatLimit, func() { n.MustParsePitch(input, system.Traditional(), notation.OctaveContext{}) })
			assert.True(t, before.Equal(n), "format and parse leave the notation unchanged")
		})
	}
}

// K-3: name bytes, including embedded marks and Unicode case changes, are free.
func TestOctaveRepeatCustomNames(t *testing.T) {
	longName := "É" + strings.Repeat("x", notation.MaxOctaveRepeatBytes)
	longAlias := "Ö" + strings.Repeat("y", notation.MaxOctaveRepeatBytes)
	for _, tc := range repeatBoundaries() {
		for _, text := range []string{"@É͵′,'tail", longName} {
			t.Run(tc.name+"/"+strconv.Itoa(len(text)), func(t *testing.T) {
				n := repeatNotation(t, tc.style, notation.NameForm{Name: spelling.C(), Text: text, Aliases: []string{longAlias}})
				base := text
				if tc.base == "c" {
					base = strings.Replace(text, "É", "é", 1)
				}
				expected := base + strings.Repeat(tc.mark, tc.count)
				p := pitch.MustFromASPN(spelling.C(), tc.written)
				got, err := n.FormatPitch(p, system.Traditional(), notation.OctaveContext{})
				require.NoError(t, err)
				assert.Equal(t, expected, got)
				back, err := n.ParsePitch(expected, system.Traditional(), notation.OctaveContext{})
				require.NoError(t, err)
				assert.True(t, back.Equal(p))
				alias := longAlias
				if tc.base == "c" {
					alias = strings.Replace(alias, "Ö", "ö", 1)
				}
				back, err = n.ParsePitch(alias+strings.Repeat(tc.mark, tc.count), system.Traditional(), notation.OctaveContext{})
				require.NoError(t, err)
				assert.True(t, back.Equal(p))
				w := tc.written + 1
				if tc.mark == "," || tc.mark == "͵" {
					w = tc.written - 1
				}
				got, err = n.FormatPitch(pitch.MustFromASPN(spelling.C(), w), system.Traditional(), notation.OctaveContext{})
				repeatError(t, err)
				assert.Empty(t, got)
				back, err = n.ParsePitch(alias+strings.Repeat(tc.mark, tc.count+1), system.Traditional(), notation.OctaveContext{})
				repeatError(t, err)
				assert.Equal(t, pitch.Pitch{}, back)
				assert.Equal(t, text, n.MustFormat(spelling.C()))
				assert.True(t, n.MustParse(longAlias).Equal(spelling.C()))
				assert.Equal(t, text, notation.MustConvert(longAlias, n, n))
			})
		}
	}
}

func TestOctaveRepeatZeroAndFixedEnglishForms(t *testing.T) {
	text := "@É" + strings.Repeat("x", notation.MaxOctaveRepeatBytes)
	for _, tc := range []struct {
		style notation.OctaveStyle
		w     int
		first string
	}{
		{notation.OctaveStyleHelmholtz, 2, "É"},
		{notation.OctaveStyleHelmholtz, 3, "é"},
		{notation.OctaveStyleHelmholtzEnglish, 0, "ÉÉÉ"},
		{notation.OctaveStyleHelmholtzEnglish, 1, "ÉÉ"},
		{notation.OctaveStyleHelmholtzEnglish, 2, "É"},
		{notation.OctaveStyleHelmholtzEnglish, 3, "é"},
		{notation.OctaveStyleLilyPond, 3, "é"},
		{notation.OctaveStyleABC, 4, "É"},
		{notation.OctaveStyleABC, 5, "é"},
	} {
		t.Run(tc.style.String()+"/"+strconv.Itoa(tc.w), func(t *testing.T) {
			n := repeatNotation(t, tc.style, notation.NameForm{Name: spelling.C(), Text: text})
			p := pitch.MustFromASPN(spelling.C(), tc.w)
			expected := strings.Replace(text, "É", tc.first, 1)
			got, err := n.FormatPitch(p, system.Traditional(), notation.OctaveContext{})
			require.NoError(t, err)
			assert.Equal(t, expected, got)
			back, err := n.ParsePitch(expected, system.Traditional(), notation.OctaveContext{})
			require.NoError(t, err)
			assert.True(t, back.Equal(p))
		})
	}
}

// K-5: canonical ABC register and algebraic cancellation control the budget.
func TestOctaveRepeatABCCanonicalBudget(t *testing.T) {
	n := notation.MustBuiltIn(notation.ABCID)
	for _, tc := range []struct {
		name, input, canonical string
		w                      int
	}{
		{"uppercase loses one mark", "C" + strings.Repeat("'", 65537), "c" + strings.Repeat("'", 65536), 65541},
		{"cancellation", "C" + strings.Repeat(",'", 32769), "C", 4},
		{"cancellation at lower boundary", "C" + strings.Repeat(",", 65537) + "'", "C" + strings.Repeat(",", 65536), -65532},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, err := n.ParsePitch(tc.input, system.Traditional(), notation.OctaveContext{})
			require.NoError(t, err)
			assert.True(t, p.Equal(pitch.MustFromASPN(spelling.C(), tc.w)))
			assert.Equal(t, tc.canonical, n.MustFormatPitch(p, system.Traditional(), notation.OctaveContext{}))
		})
	}
}

// K-4: conversion rejects either stage and preserves source error precedence.
func TestOctaveRepeatConversion(t *testing.T) {
	num := notation.MustBuiltIn(notation.EnglishUnicodeID)
	ctx := notation.OctaveContext{Numbering: number.ASPN()}
	for _, tc := range repeatBoundaries() {
		t.Run(tc.name, func(t *testing.T) {
			n := repeatNotation(t, tc.style)
			w := tc.written + 1
			if tc.mark == "," || tc.mark == "͵" {
				w = tc.written - 1
			}
			for _, c := range []struct {
				input          string
				from, to       notation.Notation
				fromCtx, toCtx notation.OctaveContext
			}{
				{"C" + strconv.Itoa(w), num, n, ctx, notation.OctaveContext{}},
				{tc.base + strings.Repeat(tc.mark, tc.count+1), n, num, notation.OctaveContext{}, ctx},
			} {
				got, err := notation.ConvertPitch(c.input, c.from, c.to, system.Traditional(), c.fromCtx, c.toCtx)
				repeatError(t, err)
				assert.Empty(t, got)
				repeatPanic(t, notation.ErrOctaveRepeatLimit, func() { notation.MustConvertPitch(c.input, c.from, c.to, system.Traditional(), c.fromCtx, c.toCtx) })
			}
			_, err := notation.ConvertPitch("invalid", num, n, system.Traditional(), ctx, notation.OctaveContext{})
			assert.ErrorIs(t, err, notation.ErrParse)
			assert.NotErrorIs(t, err, notation.ErrOctaveRepeatLimit)
		})
	}
}

func TestOctaveRepeatErrorPrecedence(t *testing.T) {
	n := notation.MustBuiltIn(notation.LilyPondEnglishID)
	p := pitch.MustFromASPN(spelling.C(), 65540)
	for _, tc := range []struct {
		name string
		n    notation.Notation
		p    pitch.Pitch
		sys  *system.System
		want error
	}{
		{"undefined notation", notation.Notation{}, p, system.Traditional(), notation.ErrUndefined},
		{"undefined pitch", n, pitch.Pitch{}, system.Traditional(), notation.ErrUndefined},
		{"undefined system", n, p, nil, notation.ErrUndefined},
		{"inadmissible name", n, pitch.MustFromASPN(spelling.C().WithAlteration(alteration.MustFromFraction(1, 2)), 65540), system.Traditional(), system.ErrInadmissible},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.n.FormatPitch(tc.p, tc.sys, notation.OctaveContext{})
			assert.Empty(t, got)
			assert.ErrorIs(t, err, tc.want)
			assert.NotErrorIs(t, err, notation.ErrOctaveRepeatLimit)
		})
	}
	input := "cqs" + strings.Repeat("'", 65537)
	back, err := n.ParsePitch(input, system.Traditional(), notation.OctaveContext{})
	repeatError(t, err)
	assert.NotErrorIs(t, err, system.ErrInadmissible)
	assert.Equal(t, pitch.Pitch{}, back)
	back, err = n.ParsePitch(input+"?", system.Traditional(), notation.OctaveContext{})
	assert.ErrorIs(t, err, notation.ErrParse)
	assert.NotErrorIs(t, err, notation.ErrOctaveRepeatLimit)
	assert.Equal(t, pitch.Pitch{}, back)
}

// K-6: only numeric styles format int extremes; repeated styles are reviewed
// statically for arithmetic overflow and early rejection.
func TestOctaveRepeatDoesNotLimitNumericStyles(t *testing.T) {
	for _, tc := range []struct {
		style notation.OctaveStyle
		w     int
		want  string
	}{
		{notation.OctaveStyleNumber, math.MinInt, "C" + strconv.Itoa(math.MinInt)},
		{notation.OctaveStyleNumber, math.MaxInt, "C" + strconv.Itoa(math.MaxInt)},
		{notation.OctaveStyleHelmholtzNumbered, math.MaxInt, "c" + strconv.Itoa(math.MaxInt-3)},
	} {
		t.Run(tc.style.String()+"/"+strconv.Itoa(tc.w), func(t *testing.T) {
			n := repeatNotation(t, tc.style)
			ctx := notation.OctaveContext{Numbering: number.ASPN()}
			p := pitch.MustFromASPN(spelling.C(), tc.w)
			got, err := n.FormatPitch(p, system.Traditional(), ctx)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
			back, err := n.ParsePitch(got, system.Traditional(), ctx)
			require.NoError(t, err)
			assert.True(t, back.Equal(p))
		})
	}
}

func TestOctaveRepeatDoesNotLimitOtherText(t *testing.T) {
	text := "É" + strings.Repeat("x", 65537)
	rules := notation.MustNewNameRules(notation.NameRulesDefinition{Names: []notation.NameForm{{Name: spelling.C(), Text: text}}})
	catalogText := strings.Repeat("octave", 11000)
	cat := naming.MustNew(map[int]string{4: catalogText})
	sep := strings.Repeat(":", 65537)
	for _, tc := range []struct {
		style notation.OctaveStyle
		w     int
		sep   string
		want  string
	}{
		{notation.OctaveStyleNumber, 4, sep, text + sep + "4"},
		{notation.OctaveStyleName, 4, sep, text + sep + catalogText},
		{notation.OctaveStyleAM, 4, "", "é" + text[len("É"):] + "¹"},
		{notation.OctaveStyleHelmholtzNumbered, 4, "", "é" + text[len("É"):] + "1"},
	} {
		t.Run(tc.style.String(), func(t *testing.T) {
			n := notation.MustNew(notation.Definition{Name: text, Rules: rules, OctaveStyle: tc.style, OctaveSeparator: tc.sep})
			ctx := notation.OctaveContext{Numbering: number.ASPN(), Catalog: cat}
			p := pitch.MustFromASPN(spelling.C(), tc.w)
			got, err := n.FormatPitch(p, system.Traditional(), ctx)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
			back, err := n.ParsePitch(got, system.Traditional(), ctx)
			require.NoError(t, err)
			assert.True(t, back.Equal(p))
		})
	}
}
