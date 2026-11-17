package tempomap_test

import (
	"math/big"
	"strconv"
	"strings"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempomap"

	"github.com/stretchr/testify/assert"
)

func requireWholeInputExcerpt(t *testing.T, input, nested string, err error, own error, dependencies ...error) {
	t.Helper()
	requireTempomapError(t, err, own, dependencies...)
	assert.LessOrEqual(t, len(err.Error()), 512)
	assert.Contains(t, err.Error(), wholeInputExcerpt(input))
	if nested != "" {
		assert.NotContains(t, err.Error(), strconv.Quote(nested), "error quoted only a nested item")
	}
	for _, dependency := range []error{
		fraction.ErrInvalidText,
		fraction.ErrOutOfRange,
		duration.ErrInvalidText,
		duration.ErrOutOfRange,
		duration.ErrUnset,
		tempo.ErrInvalidText,
		tempo.ErrOutOfRange,
		tempo.ErrUnset,
		tempo.ErrNoTempo,
	} {
		assert.NotContains(t, err.Error(), dependency.Error(), "dependency text leaked into the parse error")
	}
}

func wholeInputExcerpt(input string) string {
	if len(input) <= 64 {
		return strconv.Quote(input)
	}
	return strconv.Quote(input[:64]) + "..."
}

func TestRegressionParseErrorsQuoteWholePublicInput(t *testing.T) {
	t.Parallel()
	tooBig := new(big.Int).Add(fraction.MaxComponentBig(), big.NewInt(1)).String()
	longPlaces := strings.Repeat("9", 80)
	duplicateMap := "tempomap[first@p:0/1:main:4=60;single@p:1/1:jump:4=90;single@p:1/1:jump:4=120]"

	for _, tt := range []struct {
		name         string
		input        string
		nested       string
		own          error
		dependencies []error
		parse        func(string) error
		unmarshal    func([]byte) error
	}{
		{
			name: "change nested grammar", input: "single@p:1/1:jump:4=bogus", nested: "4=bogus", own: tempomap.ErrInvalidText,
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
			name: "extension nested grammar", input: "hold@p:1/2:bogus", nested: "bogus", own: tempomap.ErrInvalidText,
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
			name: "map semantic error", input: "tempomap[first@p:0/1:main:4=60;single@p:1/1:a-tempo]", nested: "single@p:1/1:a-tempo", own: tempomap.ErrNoDeviation,
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
			name: "map nested grammar", input: "tempomap[first@p:0/1:main:4=60;single@p:1/1:jump:4=bogus]", nested: "4=bogus", own: tempomap.ErrInvalidText,
			parse: func(text string) error { _, err := tempomap.Parse(text); return err },
			unmarshal: func(text []byte) error {
				value := tempomap.MustNew(quarter(60))
				err := value.UnmarshalText(text)
				return err
			},
		},
		{
			name: "long change range", input: "single@p:1/1:jump:4=" + tooBig, own: tempomap.ErrOutOfRange, dependencies: []error{fraction.ErrOutOfRange},
			parse: func(text string) error { _, err := tempomap.ParseChange(text); return err },
			unmarshal: func(text []byte) error {
				value := tempomap.MustNewMain(at(0, 1), quarter(60))
				err := value.UnmarshalText(text)
				return err
			},
		},
		{
			name: "long extension range", input: "hold@p:1/2:" + tooBig, own: tempomap.ErrOutOfRange, dependencies: []error{fraction.ErrOutOfRange},
			parse: func(text string) error { _, err := tempomap.ParseExtension(text); return err },
			unmarshal: func(text []byte) error {
				value := tempomap.MustNewExtension(at(1, 2), sec(1, 1))
				err := value.UnmarshalText(text)
				return err
			},
		},
		{
			name: "long map semantic error", input: duplicateMap, nested: "single@p:1/1:jump:4=120", own: tempomap.ErrDuplicateAddress,
			parse: func(text string) error { _, err := tempomap.Parse(text); return err },
			unmarshal: func(text []byte) error {
				value := tempomap.MustNew(quarter(60))
				err := value.UnmarshalText(text)
				return err
			},
		},
		{
			name: "long map nested range", input: "tempomap[first@p:0/1:main:4=60;single@p:1/1:jump:4=" + tooBig + "]", own: tempomap.ErrOutOfRange, dependencies: []error{fraction.ErrOutOfRange},
			parse: func(text string) error { _, err := tempomap.Parse(text); return err },
			unmarshal: func(text []byte) error {
				value := tempomap.MustNew(quarter(60))
				err := value.UnmarshalText(text)
				return err
			},
		},
		{
			name: "seconds places", input: "s~1/1@1001:1/1", own: tempomap.ErrInvalidArgument,
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
			name: "position long places", input: "p~1/1@" + longPlaces + ":1/1", own: tempomap.ErrInvalidArgument,
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
			requireWholeInputExcerpt(t, tt.input, tt.nested, tt.parse(tt.input), tt.own, tt.dependencies...)
			requireWholeInputExcerpt(t, tt.input, tt.nested, tt.unmarshal([]byte(tt.input)), tt.own, tt.dependencies...)
		})
	}
}
