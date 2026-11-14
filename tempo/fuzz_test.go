package tempo_test

import (
	"errors"
	"testing"

	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempo/term"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fuzzSeeds are seeds for every parse grammar of architecture 4.14.
var fuzzSeeds = []string{
	"4=60", "4.=60", "4=185/2", "4=92.5", "8[3:2]=360", "breve=1/1000", "4≈108", "4.≈60", "8=132-144", "4=120.5-131",
	"0/1", "1/3", "2/1", "4/2", "Allegro", "Allegro, 4=120", `"4=60", 4=120`, "Скоро, но не слишком, 4=120",
	`"\"Presto\""`, "Vivo, 4≈170", "", "4=0", "4=~108", "4~108", "4=01", "4=1/0", "4=132-120", "Tempo(unset)",
	"4[9223372036854775808:1]=60", "4=0.000001", "\xff", "x, , 4=60", `"x=\t\\y"`,
}

// sentinelsOf returns the sentinels of tempo matched by err.
func sentinelsOf(err error) []error {
	var out []error

	for _, s := range tempoSentinels {
		if errors.Is(err, s.err) {
			out = append(out, s.err)
		}
	}

	return out
}

// checkFuzzError asserts that err matches exactly one sentinel of tempo and
// that it is one of allowed, with a bounded text.
func checkFuzzError(t *testing.T, err error, allowed ...error) {
	t.Helper()

	got := sentinelsOf(err)
	if assert.Len(t, got, 1, "exactly one sentinel of tempo: %v", err) {
		assert.Contains(t, allowed, got[0], "unexpected kind: %v", err)
	}

	assert.LessOrEqual(t, len(err.Error()), errorTextLimit)
}

// FuzzParse covers architecture 4.14 for Parse: no panic; a success prints
// back to a text that parses to an equal value; a failure is Е-3, Е-10 or
// Е-11.
func FuzzParse(f *testing.F) {
	for _, s := range fuzzSeeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, x string) {
		v, err := tempo.Parse(x)
		if err != nil {
			checkFuzzError(t, err, tempo.ErrNoTempo, tempo.ErrInvalidText, tempo.ErrOutOfRange)
			assert.Equal(t, tempo.Tempo{}, v)

			return
		}

		back, err := tempo.Parse(v.String())
		require.NoError(t, err)
		assert.Equal(t, v, back)
	})
}

// FuzzParseApprox covers architecture 4.14 for ParseApprox.
func FuzzParseApprox(f *testing.F) {
	for _, s := range fuzzSeeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, x string) {
		v, err := tempo.ParseApprox(x)
		if err != nil {
			checkFuzzError(t, err, tempo.ErrNoTempo, tempo.ErrInvalidText, tempo.ErrOutOfRange)
			assert.Equal(t, tempo.Approx{}, v)

			return
		}

		back, err := tempo.ParseApprox(v.String())
		require.NoError(t, err)
		assert.Equal(t, v, back)
	})
}

// FuzzParseRange covers architecture 4.14 for ParseRange.
func FuzzParseRange(f *testing.F) {
	for _, s := range fuzzSeeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, x string) {
		v, err := tempo.ParseRange(x)
		if err != nil {
			checkFuzzError(t, err, tempo.ErrInvalidArgument, tempo.ErrNoTempo, tempo.ErrInvalidText, tempo.ErrOutOfRange)
			assert.Equal(t, tempo.Range{}, v)

			return
		}

		back, err := tempo.ParseRange(v.String())
		require.NoError(t, err)
		assert.Equal(t, v, back)
	})
}

// FuzzParseSeconds covers architecture 4.14 for ParseSeconds.
func FuzzParseSeconds(f *testing.F) {
	for _, s := range fuzzSeeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, x string) {
		v, err := tempo.ParseSeconds(x)
		if err != nil {
			checkFuzzError(t, err, tempo.ErrInvalidText, tempo.ErrOutOfRange)
			assert.Equal(t, tempo.Seconds{}, v)

			return
		}

		back, err := tempo.ParseSeconds(v.String())
		require.NoError(t, err)
		assert.Equal(t, v, back)
	})
}

// FuzzParseIndication covers architecture 4.14 for ParseIndication with a
// built-in list and a user list of unusual names: a success prints back to
// a text that parses to an equal value; a failure is Е-2, Е-3, Е-10 or Е-11.
func FuzzParseIndication(f *testing.F) {
	for _, s := range fuzzSeeds {
		f.Add(s)
	}

	lists := []term.List{term.EnglishWikipedia(), customList()}

	f.Fuzz(func(t *testing.T, x string) {
		for _, l := range lists {
			v, err := tempo.ParseIndication(x, l)
			if err != nil {
				checkFuzzError(t, err, tempo.ErrInvalidArgument, tempo.ErrNoTempo, tempo.ErrInvalidText, tempo.ErrOutOfRange)
				assert.Equal(t, tempo.Indication{}, v)

				continue
			}

			back, err := tempo.ParseIndication(v.String(), l)
			require.NoError(t, err, "%q", v.String())
			assert.Equal(t, v, back)
		}
	})
}
