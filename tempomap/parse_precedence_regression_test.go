package tempomap_test

import (
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempomap"

	"github.com/stretchr/testify/assert"
)

func requireWholeGrammarPrecedence(t *testing.T, input string, err error) {
	t.Helper()
	requireWholeInputExcerpt(t, input, "", err, tempomap.ErrInvalidText)
	for _, dependency := range []error{
		fraction.ErrOutOfRange,
		duration.ErrOutOfRange,
		duration.ErrNotPositive,
		measure.ErrOutOfRange,
		tempo.ErrOutOfRange,
		tempo.ErrNoTempo,
	} {
		assert.NotErrorIs(t, err, dependency, "late grammar error must suppress an earlier range or semantic cause")
	}
}

func TestRegressionWholeGrammarPrecedesRangeAndSemantics(t *testing.T) {
	t.Parallel()
	tooBig := new(big.Int).Add(fraction.MaxComponentBig(), big.NewInt(1)).String()
	for _, tt := range []struct {
		name      string
		input     string
		parse     func(string) error
		unmarshal func([]byte) error
	}{
		{
			name: "address range before trailing grammar", input: "p:" + tooBig + "/1:junk",
			parse: func(text string) error { _, err := tempomap.ParseAddress(text); return err },
			unmarshal: func(text []byte) error {
				value := at(1, 2)
				before := value.String()
				err := value.UnmarshalText(text)
				assert.Equal(t, before, value.String())
				return err
			},
		},
		{
			name: "change range before trailing grammar", input: "single@p:1/1:jump:4=" + tooBig + ":junk",
			parse: func(text string) error { _, err := tempomap.ParseChange(text); return err },
			unmarshal: func(text []byte) error {
				value := tempomap.MustNewMain(at(0, 1), quarter(60))
				before := value.String()
				err := value.UnmarshalText(text)
				assert.Equal(t, before, value.String())
				return err
			},
		},
		{
			name: "change semantics before trailing grammar", input: "single@p:1/1:modulation:4.@0/1:junk",
			parse: func(text string) error { _, err := tempomap.ParseChange(text); return err },
			unmarshal: func(text []byte) error {
				value := tempomap.MustNewMain(at(0, 1), quarter(60))
				before := value.String()
				err := value.UnmarshalText(text)
				assert.Equal(t, before, value.String())
				return err
			},
		},
		{
			name: "extension range before trailing grammar", input: "hold@p:1/2:" + tooBig + ":junk",
			parse: func(text string) error { _, err := tempomap.ParseExtension(text); return err },
			unmarshal: func(text []byte) error {
				value := tempomap.MustNewExtension(at(1, 2), sec(1, 1))
				before := value.String()
				err := value.UnmarshalText(text)
				assert.Equal(t, before, value.String())
				return err
			},
		},
		{
			name:  "map range before later malformed record",
			input: "tempomap[first@p:0/1:main:4=60;single@p:1/1:jump:4=" + tooBig + ";single@p:2/1:jump:4=bogus]",
			parse: func(text string) error { _, err := tempomap.Parse(text); return err },
			unmarshal: func(text []byte) error {
				value := tempomap.MustNew(quarter(60))
				before := value.String()
				err := value.UnmarshalText(text)
				assert.Equal(t, before, value.String())
				return err
			},
		},
		{
			name:  "map semantics before later malformed record",
			input: "tempomap[first@p:0/1:main:4=60;single@p:1/1:a-tempo;single@p:2/1:jump:4=bogus]",
			parse: func(text string) error { _, err := tempomap.Parse(text); return err },
			unmarshal: func(text []byte) error {
				value := tempomap.MustNew(quarter(60))
				before := value.String()
				err := value.UnmarshalText(text)
				assert.Equal(t, before, value.String())
				return err
			},
		},
		{
			name: "seconds range before malformed bound", input: "s~" + tooBig + "/1@3:bogus",
			parse: func(text string) error { _, err := tempomap.ParseSecondsResult(text); return err },
			unmarshal: func(text []byte) error {
				value := tempomap.MustParseSecondsResult("s:1/1")
				before := value.String()
				err := value.UnmarshalText(text)
				assert.Equal(t, before, value.String())
				return err
			},
		},
		{
			name: "position range before malformed bound", input: "p~" + tooBig + "/1@3:bogus",
			parse: func(text string) error { _, err := tempomap.ParsePositionResult(text); return err },
			unmarshal: func(text []byte) error {
				value := tempomap.MustParsePositionResult("p:1/1")
				before := value.String()
				err := value.UnmarshalText(text)
				assert.Equal(t, before, value.String())
				return err
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			requireWholeGrammarPrecedence(t, tt.input, tt.parse(tt.input))
			requireWholeGrammarPrecedence(t, tt.input, tt.unmarshal([]byte(tt.input)))
		})
	}
}
