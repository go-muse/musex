package template_test

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIssueRotationAndDirectConstructionRoundTrip(t *testing.T) {
	t.Parallel()

	original, err := template.ClosedFromIntervals(intervals("{-1/257}2 {+1/257}2 {+1/263}2"))
	require.NoError(t, err)
	rotated, err := original.Rotate(2)
	require.NoError(t, err)

	const rotatedText = "{+1/257}2 {+1/263}2 {-1/257}2"
	assert.Equal(t, rotatedText, rotated.String())
	direct, err := template.ClosedFromIntervals(intervals(rotatedText))
	require.NoError(t, err)
	requireTemplate(t, rotated, direct)
	requireTemplateTextRoundTrip(t, rotatedText, direct)

	coordinates := coordinatesOf(direct.AdjacentIntervals())
	cycle, ok := direct.Cycle()
	require.True(t, ok)
	requireCoordinate(t, sumTemplateCoordinates(coordinates), cycle)

	// The third place is 1/257 + 1/263 = 520/67591 above the second
	// place's natural quality; denominator 67591 exceeds 65536.
	requireUnrepresentablePlace(t, direct, 3)
	_, err = direct.BetweenPlaces(1, 3)
	requireTemplateRangeError(t, err, true)

	place, err := direct.PlaceInterval(1)
	require.NoError(t, err)
	requireCoordinate(t, placeCoordinate(coordinates, true, 1), place)
	place, err = direct.PlaceInterval(2)
	require.NoError(t, err)
	requireCoordinate(t, placeCoordinate(coordinates, true, 2), place)
	adjacent, err := direct.AdjacentIntervalAfter(3)
	require.NoError(t, err)
	requireCoordinate(t, coordinates[2], adjacent)
	between, err := direct.BetweenPlaces(1, 2)
	require.NoError(t, err)
	requireCoordinate(t, coordinates[0], between)
	between, err = direct.BetweenPlaces(2, 3)
	require.NoError(t, err)
	requireCoordinate(t, coordinates[1], between)

	rotatedAgain, err := direct.Rotate(2)
	require.NoError(t, err)
	originalThird, err := original.Rotate(3)
	require.NoError(t, err)
	requireTemplate(t, originalThird, rotatedAgain)
	cycleAgain, ok := rotatedAgain.Cycle()
	require.True(t, ok)
	requireCoordinate(t, sumTemplateCoordinates(coordinates), cycleAgain)
}

func TestClosedIntermediateRangeBoundaries(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		text string
	}{
		{
			name: "intermediate alteration magnitude",
			// Place 3 has deviation 2*MaxMagnitude. The octave-sized closing
			// interval makes the complete cycle size 10 with deviation MaxMagnitude.
			text: "{+16777216}2 {+16777216}2 {-16777216}8",
		},
		{
			name: "intermediate alteration denominator",
			// Place 3 has denominator 65536*65535 after exact addition.
			// The complete cycle has deviation 2 + 1/65535 and is representable.
			text: "{+1/65536}2 {+1/65535}2 {-1/65536}2",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			adjacent := intervals(tc.text)
			closed, err := template.ClosedFromIntervals(adjacent)
			require.NoError(t, err)
			requireTemplateTextRoundTrip(t, tc.text, closed)

			coordinates := coordinatesOf(adjacent)
			cycle, ok := closed.Cycle()
			require.True(t, ok)
			requireCoordinate(t, sumTemplateCoordinates(coordinates), cycle)
			place, err := closed.PlaceInterval(2)
			require.NoError(t, err)
			requireCoordinate(t, placeCoordinate(coordinates, true, 2), place)
			requireUnrepresentablePlace(t, closed, 3)

			opened, err := template.OpenFromIntervals(adjacent)
			requireTemplateRangeError(t, err, true)
			assert.Equal(t, template.Template{}, opened)
			opened, err = template.Parse(tc.text + " |")
			requireTemplateRangeError(t, err, true)
			assert.Equal(t, template.Template{}, opened)
			assert.False(t, errors.Is(err, template.ErrParse))
		})
	}
}

func TestFullCycleRangeRejections(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name            string
		text            string
		alterationCause bool
	}{
		{
			name:            "alteration magnitude",
			text:            "{+16777216}2 {+16777216}2",
			alterationCause: true,
		},
		{
			name:            "alteration denominator",
			text:            "{+1/65536}2 {+1/65535}2",
			alterationCause: true,
		},
		{
			name: "interval size",
			text: fmt.Sprintf("M%d M2", int64(math.MaxInt)),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			calls := []struct {
				name string
				call func() (template.Template, error)
			}{
				{name: "closed constructor", call: func() (template.Template, error) {
					return template.ClosedFromIntervals(intervals(tc.text))
				}},
				{name: "open constructor", call: func() (template.Template, error) {
					return template.OpenFromIntervals(intervals(tc.text))
				}},
				{name: "closed parse", call: func() (template.Template, error) {
					return template.Parse(tc.text)
				}},
				{name: "open parse", call: func() (template.Template, error) {
					return template.Parse(tc.text + " |")
				}},
			}
			for _, call := range calls {
				t.Run(call.name, func(t *testing.T) {
					got, err := call.call()
					requireTemplateRangeError(t, err, tc.alterationCause)
					assert.Equal(t, template.Template{}, got)
					assert.False(t, errors.Is(err, template.ErrParse))
					assert.LessOrEqual(t, len(err.Error()), 512)
				})
			}
		})
	}
}

func TestSpellingsWithUnrepresentableIntermediatePlace(t *testing.T) {
	t.Parallel()

	// Independently, C-D and D-E have deviations 1/257 and 1/263.
	// E-F includes its natural minor-second deviation -1, and F-C closes
	// with -1/257. Thus C-E has denominator 257*263=67591, while the
	// complete C-D-E-F-C cycle is a perfect octave.
	spellings := names("C-1/257 D E+1/263 F")
	const expectedText = "{+1/257}2 {+1/263}2 {-264/263}2 {-1/257}5"
	expected, err := template.ClosedFromIntervals(intervals(expectedText))
	require.NoError(t, err)
	requireTemplateTextRoundTrip(t, expectedText, expected)

	fromSpellings, err := template.ClosedFromSpellings(spellings, interval.PerfectOctave())
	require.NoError(t, err)
	requireTemplate(t, expected, fromSpellings)
	fromOctave, err := template.FromSpellingsInOctave(spellings)
	require.NoError(t, err)
	requireTemplate(t, expected, fromOctave)
	requireTemplateTextRoundTrip(t, expectedText, fromSpellings)
	requireTemplateTextRoundTrip(t, expectedText, fromOctave)

	coordinates := coordinatesOf(expected.AdjacentIntervals())
	cycle, ok := fromOctave.Cycle()
	require.True(t, ok)
	requireCoordinate(t, sumTemplateCoordinates(coordinates), cycle)
	requireInterval(t, "P8", cycle)
	requireUnrepresentablePlace(t, fromOctave, 3)
	place, err := fromOctave.PlaceInterval(4)
	require.NoError(t, err)
	requireCoordinate(t, placeCoordinate(coordinates, true, 4), place)

	opened, err := template.OpenFromSpellings(spellings)
	requireTemplateRangeError(t, err, true)
	assert.Equal(t, template.Template{}, opened)
}

func TestExistingErrorCategoriesAndPrecedence(t *testing.T) {
	t.Parallel()

	parseCases := []struct {
		name string
		text string
		want error
	}{
		{name: "empty", text: "", want: template.ErrParse},
		{name: "malformed", text: "M2  M2", want: template.ErrParse},
		{name: "descending", text: "M2 -M2 P8", want: template.ErrInvalidInterval},
		{name: "unison", text: "M2 P1", want: template.ErrInvalidInterval},
	}
	for _, tc := range parseCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := template.Parse(tc.text)
			require.ErrorIs(t, err, tc.want)
			assert.Equal(t, template.Template{}, got)
			if !errors.Is(tc.want, template.ErrParse) {
				assert.False(t, errors.Is(err, template.ErrParse))
			}
		})
	}

	t.Run("unset precedes invalid interval and range", func(t *testing.T) {
		t.Parallel()
		input := []interval.Interval{{}, interval.PerfectUnison()}
		input = append(input, intervals("{+16777216}2 {+16777216}2")...)
		got, err := template.ClosedFromIntervals(input)
		require.ErrorIs(t, err, template.ErrUnset)
		assert.ErrorIs(t, err, interval.ErrUnset)
		assert.False(t, errors.Is(err, template.ErrInvalidInterval))
		assert.False(t, errors.Is(err, template.ErrOutOfRange))
		assert.Equal(t, template.Template{}, got)
	})

	t.Run("empty input precedes derived checks", func(t *testing.T) {
		t.Parallel()
		got, err := template.ClosedFromIntervals(nil)
		require.ErrorIs(t, err, template.ErrEmptyInput)
		assert.Equal(t, template.Template{}, got)
	})

	t.Run("invalid interval precedes range", func(t *testing.T) {
		t.Parallel()
		input := append([]interval.Interval{interval.PerfectUnison()}, intervals("{+16777216}2 {+16777216}2")...)
		got, err := template.ClosedFromIntervals(input)
		require.ErrorIs(t, err, template.ErrInvalidInterval)
		assert.False(t, errors.Is(err, template.ErrOutOfRange))
		assert.Equal(t, template.Template{}, got)
	})
}

func TestLongParseErrorsAreBounded(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
	}{
		{name: "NUL", input: strings.Repeat("\x00", 1000)},
		{name: "control bytes", input: strings.Repeat("\x01\x02\x1b\x7f", 250)},
		{name: "invalid UTF-8", input: string(bytes.Repeat([]byte{0xff, 0xfe, 0x80}, 400))},
		{name: "multibyte", input: strings.Repeat("ж🎼", 400)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, parseErr := template.Parse(tc.input)
			requireBoundedParseError(t, parseErr)
			assert.Equal(t, template.Template{}, got)

			receiver := template.Ionian()
			before := receiver
			unmarshalErr := receiver.UnmarshalText([]byte(tc.input))
			requireBoundedParseError(t, unmarshalErr)
			requireTemplate(t, before, receiver)

			func() {
				defer func() {
					recovered := recover()
					require.NotNil(t, recovered)
					panicErr, ok := recovered.(error)
					require.True(t, ok, "MustParse panic has type %T, want error", recovered)
					assert.IsType(t, parseErr, panicErr)
					assert.Equal(t, parseErr.Error(), panicErr.Error())
					requireBoundedParseError(t, panicErr)
				}()
				_ = template.MustParse(tc.input)
			}()
		})
	}
}

func TestUnmarshalTextOwnFallbackBoundary(t *testing.T) {
	t.Parallel()

	// Independent byte calculation from the accepted plan: each 1000-NUL
	// dependency excerpt occupies 261 bytes, each 52-NUL excerpt 210 bytes.
	// Two causes reduce Parse by 102 bytes to exactly 511; UnmarshalText's
	// former wrapper would instead produce 536 bytes and must fall back.
	input := strings.Repeat("\x00", 52)
	got, parseErr := template.Parse(input)
	requireBoundedParseError(t, parseErr)
	assert.Equal(t, template.Template{}, got)
	assert.Len(t, parseErr.Error(), 511)
	requireParseMultiWrapTree(t, parseErr, input, 0)

	receiver := template.Ionian()
	before := receiver
	unmarshalErr := receiver.UnmarshalText([]byte(input))
	requireBoundedParseError(t, unmarshalErr)
	requireTemplate(t, before, receiver)
	requireParseMultiWrapTree(t, unmarshalErr, input, 2)
}

func TestParseDependencyCausesRemainDistinct(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		input      string
		first      error
		second     error
		additional error
		unwanted   []error
	}{
		{
			name:       "out of range first parse",
			input:      "{+1/65537}3",
			first:      interval.ErrOutOfRange,
			second:     interval.ErrParse,
			additional: alteration.ErrOutOfRange,
			unwanted:   []error{interval.ErrInvalidArgument},
		},
		{
			name:     "invalid argument first parse",
			input:    "M5",
			first:    interval.ErrInvalidArgument,
			second:   interval.ErrParse,
			unwanted: []error{interval.ErrOutOfRange, alteration.ErrOutOfRange},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, firstErr := interval.Parse(tc.input)
			require.ErrorIs(t, firstErr, tc.first)
			_, secondErr := interval.ParseName(tc.input)
			require.ErrorIs(t, secondErr, tc.second)

			got, err := template.Parse(tc.input)
			requireBoundedParseError(t, err)
			assert.Equal(t, template.Template{}, got)
			assert.ErrorIs(t, err, tc.first)
			assert.ErrorIs(t, err, tc.second)
			if tc.additional != nil {
				assert.ErrorIs(t, err, tc.additional)
			}
			for _, unwanted := range tc.unwanted {
				assert.False(t, errors.Is(err, unwanted), "unexpected cause %v", unwanted)
			}
		})
	}
}

func TestFallbackErrorTree(t *testing.T) {
	t.Parallel()

	input := strings.Repeat("\x00", 1000)
	got, err := template.Parse(input)
	requireBoundedParseError(t, err)
	assert.Equal(t, template.Template{}, got)
	requireParseMultiWrapTree(t, err, input, 1)
}

func requireTemplateTextRoundTrip(t testing.TB, text string, want template.Template) {
	t.Helper()
	assert.Equal(t, text, want.String())
	marshaled, err := want.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, text, string(marshaled))

	parsed, err := template.Parse(text)
	require.NoError(t, err)
	requireTemplate(t, want, parsed)
	var decoded template.Template
	require.NoError(t, decoded.UnmarshalText(marshaled))
	requireTemplate(t, want, decoded)
}

func sumTemplateCoordinates(coordinates []linearCoordinate) linearCoordinate {
	total := newCoordinate(0, new(big.Rat))
	for _, coordinate := range coordinates {
		total = addCoordinates(total, coordinate)
	}
	return total
}

func requireUnrepresentablePlace(t testing.TB, tmpl template.Template, place int) {
	t.Helper()
	got, err := tmpl.PlaceInterval(place)
	requireTemplateRangeError(t, err, true)
	assert.False(t, got.IsValid())
	places, err := tmpl.PlaceIntervals()
	requireTemplateRangeError(t, err, true)
	assert.Nil(t, places)
}

func requireTemplateRangeError(t testing.TB, err error, alterationCause bool) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, template.ErrOutOfRange)
	assert.ErrorIs(t, err, interval.ErrOutOfRange)
	if alterationCause {
		assert.ErrorIs(t, err, alteration.ErrOutOfRange)
	}
}

func requireBoundedParseError(t testing.TB, err error) {
	t.Helper()
	require.Error(t, err)
	assert.LessOrEqual(t, len(err.Error()), 512)
	assert.True(t, utf8.ValidString(err.Error()), "error text is not valid UTF-8: %q", err.Error())
	assert.ErrorIs(t, err, template.ErrParse)
	assert.ErrorIs(t, err, interval.ErrParse)
}

func requireParseMultiWrapTree(t testing.TB, err error, input string, singleWrappers int) {
	t.Helper()
	current := err
	for range singleWrappers {
		single, ok := current.(interface{ Unwrap() error })
		require.True(t, ok, "error %T does not expose Unwrap() error", current)
		_, isMulti := current.(interface{ Unwrap() []error })
		assert.False(t, isMulti, "single-wrapper error unexpectedly also exposes multi-unwrap")
		current = single.Unwrap()
		require.NotNil(t, current)
	}

	multi, ok := current.(interface{ Unwrap() []error })
	require.True(t, ok, "error %T does not expose Unwrap() []error", current)
	branches := multi.Unwrap()
	require.Len(t, branches, 3)

	_, parseErr := interval.Parse(input)
	require.Error(t, parseErr)
	_, parseNameErr := interval.ParseName(input)
	require.Error(t, parseNameErr)

	templateBranches := 0
	dependencyBranches := make([]error, 0, 2)
	for _, branch := range branches {
		require.NotNil(t, branch)
		if errors.Is(branch, template.ErrParse) {
			templateBranches++
		} else {
			dependencyBranches = append(dependencyBranches, branch)
		}
	}
	assert.Equal(t, 1, templateBranches, "multi-wrap must contain exactly one template.ErrParse branch")
	require.Len(t, dependencyBranches, 2)
	assert.Equal(t, parseErr, dependencyBranches[0])
	assert.Equal(t, parseErr.Error(), dependencyBranches[0].Error())
	assert.Equal(t, parseNameErr, dependencyBranches[1])
	assert.Equal(t, parseNameErr.Error(), dependencyBranches[1].Error())
}
