package notation_test

import (
	"testing"

	"github.com/go-muse/musex/notation"
	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/octave/number"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/require"
)

func FuzzNotationNameRoundTrip(f *testing.F) {
	for _, seed := range []string{"", "C", "C♯", "C♮", "c-sharp", "heses", "до-диез", "^C", "__D", "cqs", "\xff", "C♯ "} {
		f.Add(seed, uint8(0))
	}
	expectations := expectedBuiltIns(f)
	expectations = append(expectations, expectedUpDown(f, 3), expectedUpDown(f, 4))
	f.Fuzz(func(t *testing.T, text string, selector uint8) {
		expected := expectations[int(selector)%len(expectations)]
		var n notation.Notation
		if expected.id == notation.ASCIIUpDownID {
			n = notation.MustBuiltIn(expected.id, expected.chromaticUnit)
		} else {
			n = notation.MustBuiltIn(expected.id)
		}
		name, err := n.Parse(text)
		if err != nil {
			require.ErrorIs(t, err, notation.ErrParse)
			require.False(t, name.IsValid())
			return
		}
		canonical, ok := expected.forms[name]
		require.True(t, ok)
		require.Equal(t, canonical, n.MustFormat(name))
		roundTrip, err := n.Parse(canonical)
		require.NoError(t, err)
		require.True(t, roundTrip.Equal(name))
	})
}

func FuzzNotationPitchRoundTrip(f *testing.F) {
	for _, seed := range []string{"", "C4", "C04", "C-12", "C,',", "cis¹", "cqs'", "C первая", "C ", "c''''", "\xff"} {
		f.Add(seed, uint8(0))
	}
	english := notation.MustBuiltIn(notation.EnglishUnicodeID)
	named := customNotation(f, english.NameRules(), notation.OctaveStyleName, " ")
	cases := []struct {
		n   notation.Notation
		ctx notation.OctaveContext
	}{
		{english, notation.OctaveContext{Numbering: number.ASPN()}},
		{english, notation.OctaveContext{Numbering: number.FrancoBelgian()}},
		{named, notation.OctaveContext{Catalog: naming.Russian()}},
	}
	for _, style := range []notation.OctaveStyle{
		notation.OctaveStyleAM,
		notation.OctaveStyleHelmholtz,
		notation.OctaveStyleHelmholtzEnglish,
		notation.OctaveStyleHelmholtzNumbered,
		notation.OctaveStyleLilyPond,
		notation.OctaveStyleABC,
	} {
		cases = append(cases, struct {
			n   notation.Notation
			ctx notation.OctaveContext
		}{customNotation(f, english.NameRules(), style, ""), notation.OctaveContext{}})
	}
	f.Fuzz(func(t *testing.T, text string, selector uint8) {
		tc := cases[int(selector)%len(cases)]
		p, err := tc.n.ParsePitch(text, system.Traditional(), tc.ctx)
		if err != nil {
			require.False(t, p.IsValid())
			return
		}
		canonical, err := tc.n.FormatPitch(p, system.Traditional(), tc.ctx)
		require.NoError(t, err)
		roundTrip, err := tc.n.ParsePitch(canonical, system.Traditional(), tc.ctx)
		require.NoError(t, err)
		require.True(t, roundTrip.Equal(p))
	})
}
