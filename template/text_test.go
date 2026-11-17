package template_test

import (
	"encoding/json/v2"
	"errors"
	"strings"
	"testing"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateText(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		tmpl    template.Template
		short   string
		russian string
		hasRus  bool
	}{
		{name: "ionian", tmpl: template.Ionian(), short: "M2 M2 m2 M2 M2 M2 m2", russian: "б.2 б.2 м.2 б.2 б.2 б.2 м.2", hasRus: true},
		{name: "harmonic minor", tmpl: template.HarmonicMinor(), short: "M2 m2 M2 M2 m2 A2 m2", russian: "б.2 м.2 б.2 б.2 м.2 ув.2 м.2", hasRus: true},
		{name: "blues", tmpl: template.Blues(), short: "m3 M2 A1 m2 m3 M2", russian: "м.3 б.2 ув.1 м.2 м.3 б.2", hasRus: true},
		{name: "open trichord", tmpl: opened("m3 M2"), short: "m3 M2 |", russian: "м.3 б.2 |", hasRus: true},
		{name: "neutral intervals", tmpl: closed("M2 n2 n2 M2 M2 n2 n2"), short: "M2 n2 n2 M2 M2 n2 n2", hasRus: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.short, tc.tmpl.String())
			russian, ok, err := tc.tmpl.Russian()
			require.NoError(t, err)
			assert.Equal(t, tc.hasRus, ok)
			assert.Equal(t, tc.russian, russian)
			parsed, err := template.Parse(tc.short)
			require.NoError(t, err)
			assert.True(t, parsed.Equal(tc.tmpl))
			if tc.hasRus {
				parsed, err = template.Parse(tc.russian)
				require.NoError(t, err)
				assert.True(t, parsed.Equal(tc.tmpl))
			}
		})
	}
}

func TestParseAcceptedForms(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		text string
		want template.Template
	}{
		{name: "short ionian", text: "M2 M2 m2 M2 M2 M2 m2", want: template.Ionian()},
		{name: "OMT ionian", text: "ma2 ma2 mi2 ma2 ma2 ma2 mi2", want: template.Ionian()},
		{name: "Russian dotted ionian", text: "б.2 б.2 м.2 б.2 б.2 б.2 м.2", want: template.Ionian()},
		{name: "Russian compact ionian", text: "б2 б2 м2 б2 б2 б2 м2", want: template.Ionian()},
		{name: "single octave", text: "P8", want: closed("P8")},
		{name: "arbitrary cycle", text: "M2 M2", want: closed("M2 M2")},
		{name: "open", text: "M2 M2 |", want: opened("M2 M2")},
		{name: "Russian open", text: "м.3 б.2 |", want: opened("m3 M2")},
		{name: "fractional deviations", text: "{-1/3}2 {-2/3}7", want: closed("{-1/3}2 {-2/3}7")},
		{name: "integer deviations", text: "{-5}3 {+4}6", want: closed("{-5}3 {+4}6")},
		{name: "plus quality", text: "+2", want: closed("A2")},
		{name: "degree quality", text: "°2", want: closed("d2")},
		{name: "o quality", text: "o2", want: closed("d2")},
		{name: "full perfect", text: "perf8", want: closed("P8")},
		{name: "full major", text: "maj2", want: closed("M2")},
		{name: "full minor", text: "min2", want: closed("m2")},
		{name: "full augmented", text: "aug1", want: closed("A1")},
		{name: "full diminished", text: "dim2", want: closed("d2")},
		{name: "neutral", text: "n2", want: closed("n2")},
		{name: "Russian doubly augmented", text: "дв.ув.1", want: closed("AA1")},
		{name: "Russian doubly diminished", text: "двум5", want: closed("dd5")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := template.Parse(tc.text)
			require.NoError(t, err)
			requireTemplate(t, tc.want, got)
			assert.Equal(t, tc.want.String(), got.String())
		})
	}
}

func TestParseErrors(t *testing.T) {
	t.Parallel()
	parseErrors := []string{
		"", "M2,M2", "M2  M2", " M2 m7", "M2 m7 ", "M2-M2", "2-2-1", "2 2 1",
		"|", "| M2", "M2 | M2", "M2|", "M2 | |", "M2\tM2", "M2\nM2", "TT",
	}
	for _, text := range parseErrors {
		t.Run(strings.ReplaceAll(text, " ", "space"), func(t *testing.T) {
			t.Parallel()
			got, err := template.Parse(text)
			require.ErrorIs(t, err, template.ErrParse)
			assert.False(t, got.IsValid())
		})
	}

	t.Run("dependency parse cause", func(t *testing.T) {
		t.Parallel()
		got, err := template.Parse("M2 not-an-interval")
		require.ErrorIs(t, err, template.ErrParse)
		assert.ErrorIs(t, err, interval.ErrParse)
		assert.False(t, got.IsValid())
		got, err = template.Parse("{+16777217}2")
		require.ErrorIs(t, err, template.ErrParse)
		assert.ErrorIs(t, err, interval.ErrOutOfRange)
		assert.False(t, got.IsValid())
	})

	t.Run("parsed but invalid", func(t *testing.T) {
		t.Parallel()
		for _, text := range []string{"M2 -M2 P8", "M2 P1", "M2 P1 |", "-P1"} {
			got, err := template.Parse(text)
			require.ErrorIs(t, err, template.ErrInvalidInterval)
			assert.False(t, got.IsValid())
			assert.False(t, errors.Is(err, template.ErrParse))
		}
	})

	t.Run("bounded error", func(t *testing.T) {
		t.Parallel()
		input := strings.Repeat("ж", 1000)
		_, err := template.Parse(input)
		require.Error(t, err)
		assert.LessOrEqual(t, len(err.Error()), 512)
		assert.NotContains(t, err.Error(), input)
	})
}

func TestTextMarshalAppendAndUnmarshal(t *testing.T) {
	t.Parallel()
	tmpl := opened("m3 M2")
	marshaled, err := tmpl.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, "m3 M2 |", string(marshaled))

	prefix := []byte("prefix:")
	appended, err := tmpl.AppendText(slicesClone(prefix))
	require.NoError(t, err)
	assert.Equal(t, "prefix:m3 M2 |", string(appended))
	assert.Equal(t, "prefix:", string(prefix))

	var decoded template.Template
	require.NoError(t, decoded.UnmarshalText(marshaled))
	assert.True(t, decoded.Equal(tmpl))
	before := decoded
	require.ErrorIs(t, decoded.UnmarshalText([]byte("M2  M2")), template.ErrParse)
	assert.True(t, decoded.Equal(before))
	require.ErrorIs(t, decoded.UnmarshalText([]byte("M2 P1")), template.ErrInvalidInterval)
	assert.True(t, decoded.Equal(before))
	var nilReceiver *template.Template
	require.ErrorIs(t, nilReceiver.UnmarshalText(marshaled), template.ErrUnset)

	var zero template.Template
	marshaled, err = zero.MarshalText()
	require.ErrorIs(t, err, template.ErrUnset)
	assert.Nil(t, marshaled)
	original := []byte("keep")
	got, err := zero.AppendText(original)
	require.ErrorIs(t, err, template.ErrUnset)
	assert.Equal(t, original, got)
	assert.Equal(t, "keep", string(original))
}

func slicesClone[S ~[]E, E any](input S) S {
	return append(S(nil), input...)
}

func TestTemplateJSONUsingStandardEncoder(t *testing.T) {
	t.Parallel()
	tmpl := template.HarmonicMinor()
	encoded, err := json.Marshal(tmpl)
	require.NoError(t, err)
	assert.JSONEq(t, `"M2 m2 M2 M2 m2 A2 m2"`, string(encoded))
	var decoded template.Template
	require.NoError(t, json.Unmarshal(encoded, &decoded))
	assert.True(t, decoded.Equal(tmpl))

	encoded, err = json.Marshal([]template.Template{template.Ionian(), opened("m3 M2")})
	require.NoError(t, err)
	assert.JSONEq(t, `["M2 M2 m2 M2 M2 M2 m2","m3 M2 |"]`, string(encoded))
	var list []template.Template
	require.NoError(t, json.Unmarshal(encoded, &list))
	require.Len(t, list, 2)
	assert.True(t, list[0].Equal(template.Ionian()))
	assert.True(t, list[1].Equal(opened("m3 M2")))

	before := decoded
	err = json.Unmarshal([]byte(`"M2  M2"`), &decoded)
	require.ErrorIs(t, err, template.ErrParse)
	assert.True(t, decoded.Equal(before))
	_, err = json.Marshal(template.Template{})
	require.ErrorIs(t, err, template.ErrUnset)
}

func TestTextRoundTripControls(t *testing.T) {
	t.Parallel()
	for _, tc := range controlTemplates() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tmpl := tc.template()
			parsed, err := template.Parse(tmpl.String())
			require.NoError(t, err)
			assert.True(t, parsed.Equal(tmpl))
			russian, ok, err := tmpl.Russian()
			require.NoError(t, err)
			if ok {
				parsed, err = template.Parse(russian)
				require.NoError(t, err)
				assert.True(t, parsed.Equal(tmpl))
			}
		})
	}
}
