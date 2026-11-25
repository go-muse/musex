package tuplet_test

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The tests in this file cover the ABC markers "(p:" and "(p:q:" with an empty
// last part (issue #162; specification tuplet, O-19, P-9, E-4, KP-19, KP-22,
// K-17; architecture 4.14). "(p:" reads exactly as "(p" and "(p:q:" exactly as
// "(p:q" in the same division: the same marker or the same error. A fourth
// part, including an empty one, zero counts, and leading zeros stay
// ErrInvalidText. The formatter never writes the forms with an empty last
// part.

// namedDivision is a division with a distinct subtest name: String gives the
// same "Division(invalid)" for every unknown value.
type namedDivision struct {
	name     string
	division tuplet.Division
}

// knownAndUnknownDivisions lists both known divisions and two unknown ones.
var knownAndUnknownDivisions = []namedDivision{
	{"binary", tuplet.Binary},
	{"ternary", tuplet.Ternary},
	{"division 0", tuplet.Division(0)},
	{"division 9", tuplet.Division(9)},
}

// sentinelOf returns the single tuplet sentinel that err matches, or nil.
func sentinelOf(err error) error {
	for _, s := range tupletSentinels {
		if errors.Is(err, s.err) {
			return s.err
		}
	}

	return nil
}

// assertSameABCOutcome checks that text parses in division exactly as ref:
// the same marker, or an error of the same class with the zero marker.
func assertSameABCOutcome(t *testing.T, text, ref string, division tuplet.Division) {
	t.Helper()

	want, refErr := tuplet.ParseABCMarker(ref, division)
	got, err := tuplet.ParseABCMarker(text, division)

	if refErr == nil {
		if assert.NoError(t, err, "%q parses as %q in %v", text, ref, division) {
			assert.Equal(t, want, got, "%q reads as %q in %v", text, ref, division)
		}

		return
	}

	class := sentinelOf(refErr)
	require.NotNil(t, class, "the error of %q is a tuplet error: %v", ref, refErr)
	assertErr(t, err, class)
	assert.Equal(t, tuplet.ABCMarker{}, got, "%q gives the zero marker on error", text)
}

// The cases of KP-19 for issue #162 with explicit expectations. Every result
// equals the result of the form without the last ":" in the same division,
// the Must variant returns the same marker, the formatter writes the shortest
// form, which never ends with ":", and parsing it returns the same marker.
func TestParseABCMarker_EmptyLastPartIssueCases(t *testing.T) {
	t.Parallel()

	maxText := strconv.Itoa(math.MaxInt)

	tests := []struct {
		name     string
		text     string
		ref      string
		division tuplet.Division
		actual   int
		normal   int
		events   int
		format   string
	}{
		{"triplet empty q simple meter", "(3:", "(3", tuplet.Binary, 3, 2, 3, "(3"},
		{"triplet empty q compound meter", "(3:", "(3", tuplet.Ternary, 3, 2, 3, "(3"},
		{"triplet empty r simple meter", "(3:2:", "(3:2", tuplet.Binary, 3, 2, 3, "(3"},
		{"triplet empty r compound meter", "(3:2:", "(3:2", tuplet.Ternary, 3, 2, 3, "(3"},
		{"quintuplet empty q simple meter", "(5:", "(5", tuplet.Binary, 5, 2, 5, "(5"},
		{"quintuplet empty q compound meter", "(5:", "(5", tuplet.Ternary, 5, 3, 5, "(5"},
		{"quintuplet empty r compound meter", "(5:2:", "(5:2", tuplet.Ternary, 5, 2, 5, "(5:2"},
		{"duplet empty q simple meter", "(2:", "(2", tuplet.Binary, 2, 3, 2, "(2"},
		{"p without a default with explicit q", "(10:3:", "(10:3", tuplet.Binary, 10, 3, 10, "(10:3"},
		{
			"largest int with explicit q", "(" + maxText + ":1:", "(" + maxText + ":1", tuplet.Binary,
			math.MaxInt, 1, math.MaxInt, "(" + maxText + ":1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tuplet.ParseABCMarker(tt.text, tt.division)
			require.NoError(t, err)
			assertTuplet(t, got.Tuplet, tt.actual, tt.normal)
			assert.Equal(t, tt.events, got.Events)

			ref, err := tuplet.ParseABCMarker(tt.ref, tt.division)
			require.NoError(t, err, "the form %q is the reference", tt.ref)
			assert.Equal(t, ref, got, "%q reads as %q", tt.text, tt.ref)

			if assertNoPanic(t, func() { tuplet.MustParseABCMarker(tt.text, tt.division) }) {
				assert.Equal(t, got, tuplet.MustParseABCMarker(tt.text, tt.division))
			}

			formatted, err := tuplet.FormatABCMarker(got, tt.division)
			require.NoError(t, err)
			assert.Equal(t, tt.format, formatted)
			assert.False(t, strings.HasSuffix(formatted, ":"), "the empty last part is never written")

			back, err := tuplet.ParseABCMarker(formatted, tt.division)
			require.NoError(t, err)
			assert.Equal(t, got, back)
		})
	}
}

// The failures of KP-19 for issue #162 and the failures of the same kind:
// "(p:" and "(p:q:" fail exactly as "(p" and "(p:q". The grammar accepts the
// text, so the error is not ErrInvalidText; it quotes the short input, and
// the Must variant panics with an error of the same class.
func TestParseABCMarker_EmptyLastPartErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		text     string
		ref      string
		division tuplet.Division
		want     error
	}{
		{"p without a default simple meter", "(10:", "(10", tuplet.Binary, tuplet.ErrNoDefault},
		{"p without a default compound meter", "(10:", "(10", tuplet.Ternary, tuplet.ErrNoDefault},
		{"p one", "(1:", "(1", tuplet.Binary, tuplet.ErrNoDefault},
		{"largest int without q", "(9223372036854775807:", "(9223372036854775807", tuplet.Binary, tuplet.ErrNoDefault},
		{"empty q zero division", "(3:", "(3", tuplet.Division(0), tuplet.ErrNoDefault},
		{"empty r zero division", "(3:2:", "(3:2", tuplet.Division(0), tuplet.ErrNoDefault},
		{"empty q unknown division", "(5:", "(5", tuplet.Division(3), tuplet.ErrNoDefault},
		{"empty r unknown division", "(3:2:", "(3:2", tuplet.Division(9), tuplet.ErrNoDefault},
		{"p beyond MaxInt", "(99999999999999999999:", "(99999999999999999999", tuplet.Binary, tuplet.ErrOutOfRange},
		{"p just beyond MaxInt", "(9223372036854775808:", "(9223372036854775808", tuplet.Ternary, tuplet.ErrOutOfRange},
		{"q beyond MaxInt", "(3:99999999999999999999:", "(3:99999999999999999999", tuplet.Binary, tuplet.ErrOutOfRange},
		// The range check precedes the division check (architecture 4.2).
		{"p beyond MaxInt zero division", "(99999999999999999999:", "(99999999999999999999", tuplet.Division(0), tuplet.ErrOutOfRange},
		{"q beyond MaxInt zero division", "(3:9223372036854775808:", "(3:9223372036854775808", tuplet.Division(0), tuplet.ErrOutOfRange},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, refErr := tuplet.ParseABCMarker(tt.ref, tt.division)
			require.ErrorIs(t, refErr, tt.want, "the form %q is the reference", tt.ref)

			got, err := tuplet.ParseABCMarker(tt.text, tt.division)
			assertErr(t, err, tt.want)
			assert.Equal(t, tuplet.ABCMarker{}, got)

			if err != nil {
				assert.Contains(t, err.Error(), strconv.Quote(tt.text), "the error quotes the short input in full")
			}

			_ = assertPanicErr(t, func() { tuplet.MustParseABCMarker(tt.text, tt.division) }, tt.want)
		})
	}
}

// K-17: "(p:" gives the same result or the same error as "(p", and "(p:q:" as
// "(p:q", for every p and q in a range around the ABC defaults and at the
// limits of int, in the known divisions and in unknown ones.
func TestParseABCMarker_EmptyLastPartMatchesShorterForm(t *testing.T) {
	t.Parallel()

	counts := make([]string, 0, 16)
	for v := 1; v <= 12; v++ {
		counts = append(counts, strconv.Itoa(v))
	}

	counts = append(counts, "9223372036854775806", "9223372036854775807", "9223372036854775808", "99999999999999999999")

	for _, nd := range knownAndUnknownDivisions {
		division := nd.division

		t.Run(nd.name, func(t *testing.T) {
			t.Parallel()

			for _, p := range counts {
				short := "(" + p
				assertSameABCOutcome(t, short+":", short, division)

				for _, q := range counts {
					full := short + ":" + q
					assertSameABCOutcome(t, full+":", full, division)
				}
			}
		})
	}
}

// "(p:", "(p::" and "(p" are the same marker in the same division (issues
// #158 and #162), and "(p:q:", "(p:q" and "(p:q:p" are the same marker.
func TestParseABCMarker_EmptyPartSpellingsAgree(t *testing.T) {
	t.Parallel()

	for _, division := range []tuplet.Division{tuplet.Binary, tuplet.Ternary} {
		for p := 2; p <= 9; p++ {
			short := "(" + strconv.Itoa(p)

			t.Run(division.String()+"/"+short, func(t *testing.T) {
				t.Parallel()

				want := tuplet.ABCMarker{Tuplet: tuplet.MustNew(p, abcDefaultNormal[division][p]), Events: p}

				for _, text := range []string{short, short + ":", short + "::"} {
					got, err := tuplet.ParseABCMarker(text, division)
					if assert.NoError(t, err, text) {
						assert.Equal(t, want, got, text)
					}
				}

				for q := 1; q <= 9; q++ {
					full := short + ":" + strconv.Itoa(q)
					want := tuplet.ABCMarker{Tuplet: tuplet.MustNew(p, q), Events: p}

					for _, text := range []string{full, full + ":", full + ":" + strconv.Itoa(p)} {
						got, err := tuplet.ParseABCMarker(text, division)
						if assert.NoError(t, err, text) {
							assert.Equal(t, want, got, text)
						}
					}
				}
			})
		}
	}
}

// A fourth part, including an empty one, zero counts, and leading zeros stay
// ErrInvalidText around the empty parts, in every division (P-9, E-4, KP-22).
// The ABC counts keep the canonical grammar: leading zeros are not accepted
// there, unlike the LilyPond counts.
func TestParseABCMarker_EmptyPartsInvalidText(t *testing.T) {
	t.Parallel()

	texts := []string{
		"(:",
		"(0:",
		"(00:",
		"(01:",
		"(03:",
		"(03",
		"(3:0:",
		"(3:00:",
		"(3:02",
		"(3:02:",
		"(3:2:02",
		"(3:2:0",
		"(3:2:2:",
		"(3:2:2:2",
		"(3:::",
		"(3::2:",
		"(3:2::",
		"(3: ",
		"(3 :",
		"( 3:",
		"(3:2: ",
		"(+3:",
		"(-3:",
		"(3:+2:",
		"(:2:",
		"(::",
		"3:",
		"(3;",
		"(٣:",
	}

	for _, nd := range knownAndUnknownDivisions {
		division := nd.division

		for _, text := range texts {
			t.Run(nd.name+"/"+strconv.Quote(text), func(t *testing.T) {
				t.Parallel()

				got, err := tuplet.ParseABCMarker(text, division)
				assertErr(t, err, tuplet.ErrInvalidText)
				assert.Equal(t, tuplet.ABCMarker{}, got)

				if err != nil {
					assert.Contains(t, err.Error(), strconv.Quote(text), "the error quotes the short input in full")
				}

				_ = assertPanicErr(t, func() { tuplet.MustParseABCMarker(text, division) }, tuplet.ErrInvalidText)
			})
		}
	}
}

// The order of the parse checks is the grammar (E-4), then the range of the
// counts (E-5), then the division and the default (E-3) (K-17, architecture
// 4.2), also for the forms with an empty last part.
func TestParseABCMarker_EmptyLastPartCheckOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		text     string
		division tuplet.Division
		want     error
	}{
		{"empty fourth part before division", "(3:2:2:", tuplet.Division(0), tuplet.ErrInvalidText},
		{"empty fourth part before range", "(99999999999999999999:2:2:", tuplet.Binary, tuplet.ErrInvalidText},
		{"three empty parts before division", "(3:::", tuplet.Division(0), tuplet.ErrInvalidText},
		{"leading zero before division", "(03:", tuplet.Division(0), tuplet.ErrInvalidText},
		{"leading zero before default", "(010:", tuplet.Binary, tuplet.ErrInvalidText},
		{"leading zero before range", "(099999999999999999999:", tuplet.Binary, tuplet.ErrInvalidText},
		{"zero count before range", "(99999999999999999999:0:", tuplet.Binary, tuplet.ErrInvalidText},
		{"range before division", "(99999999999999999999:", tuplet.Division(0), tuplet.ErrOutOfRange},
		{"range before missing default", "(99999999999999999999:", tuplet.Binary, tuplet.ErrOutOfRange},
		{"unknown division before missing default", "(10:", tuplet.Division(0), tuplet.ErrNoDefault},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tuplet.ParseABCMarker(tt.text, tt.division)
			assertErr(t, err, tt.want)
			assert.Equal(t, tuplet.ABCMarker{}, got)
		})
	}
}

// The formatter writes only "(p", "(p:q" and "(p:q:r" (P-9): never a form
// with an empty part, so the text never ends with ":" and never contains
// "::". Parsing the written text in the same division returns the marker.
func TestFormatABCMarker_NeverWritesEmptyPart(t *testing.T) {
	t.Parallel()

	values := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 12, math.MaxInt}

	for _, division := range []tuplet.Division{tuplet.Binary, tuplet.Ternary} {
		t.Run(division.String(), func(t *testing.T) {
			t.Parallel()

			for _, p := range values {
				for _, q := range values {
					for _, r := range values {
						m := tuplet.ABCMarker{Tuplet: tuplet.MustNew(p, q), Events: r}

						text, err := tuplet.FormatABCMarker(m, division)
						if !assert.NoError(t, err, "format %v", m) {
							continue
						}

						assert.False(t, strings.HasSuffix(text, ":"), "%v is written as %q", m, text)
						assert.NotContains(t, text, "::", "%v is written as %q", m, text)

						back, err := tuplet.ParseABCMarker(text, division)
						if assert.NoError(t, err, "parse %q, the text of %v", text, m) {
							assert.Equal(t, m, back, "%v written as %q", m, text)
						}
					}
				}
			}
		})
	}
}
