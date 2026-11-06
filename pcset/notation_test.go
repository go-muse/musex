package pcset_test

import (
	"testing"

	"github.com/go-muse/musex/pcset"
	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringAndParseCanonical(t *testing.T) {
	t.Parallel()

	t.Run("canonical_examples_and_roundtrip", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name string
			set  pcset.Set
			text string
		}{
			{name: "empty_mod_12", set: pcset.MustEmpty(12), text: "{} mod 12"},
			{name: "singleton_mod_1", set: pcset.MustNew(1, 0), text: "{0} mod 1"},
			{name: "c_major_mod_12", set: pcset.MustNew(12, 0, 4, 7), text: "{0, 4, 7} mod 12"},
			{name: "multi_digit_mod_12", set: pcset.MustNew(12, 0, 10, 11), text: "{0, 10, 11} mod 12"},
			{name: "boundary_mod_65536", set: pcset.MustNew(65_536, 0, 65_535), text: "{0, 65535} mod 65536"},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tc.text, tc.set.String())
				parsed, err := pcset.Parse(tc.text)
				require.NoError(t, err)
				assert.Equal(t, tc.set, parsed)
				assert.Equal(t, tc.set, pcset.MustParse(tc.text))
			})
		}

		var zero pcset.Set
		assert.Equal(t, "Set(invalid)", zero.String())

		// Exhaustive round-trip for all sets of N in {1, 2, 12} and samples at {19, 31, 65, 65536}.
		for _, mod := range []int{1, 2, 12} {
			for s := range pcset.MustAll(mod) {
				str := s.String()
				parsed, err := pcset.Parse(str)
				require.NoError(t, err)
				assert.Equal(t, s, parsed)
			}
		}

		for _, s := range []pcset.Set{
			pcset.MustNew(19, 0, 6, 11, 18),
			pcset.MustNew(31, 0, 10, 18, 30),
			pcset.MustNew(65, 0, 32, 64),
			pcset.MustNew(65_536, 0, 1, 32_768, 65_535),
		} {
			parsed, err := pcset.Parse(s.String())
			require.NoError(t, err)
			assert.Equal(t, s, parsed)
		}
	})

	t.Run("parse_modulus_too_large", func(t *testing.T) {
		t.Parallel()

		for _, input := range []string{
			"{} mod 65537",
			"{0} mod 100000",
			"{0, 65536} mod 65537",
		} {
			s, err := pcset.Parse(input)
			requireSentinelError(t, err, pcset.ErrModulusTooLarge)
			assert.False(t, s.IsValid())
			requirePanicsWithSentinel(t, func() { _ = pcset.MustParse(input) }, pcset.ErrModulusTooLarge)
		}
	})

	t.Run("parse_invalid_notation", func(t *testing.T) {
		t.Parallel()

		badInputs := []string{
			"",
			"Set(invalid)",
			"{0, 4, 7}",
			"[0, 4, 7] mod 12",
			"{4, 0} mod 12",
			"{0, 0} mod 12",
			"{12} mod 12",
			"{-1} mod 12",
			"{01} mod 12",
			"{+1} mod 12",
			"{-0} mod 12",
			"{} mod 0",
			"{} mod -12",
			"{} mod 012",
			"{} mod +12",
			"{} mod 99999999999999999999999999999",
			"{0,4,7} mod 12",
			"{0,  4, 7} mod 12",
			"{ 0, 4, 7} mod 12",
			"{0, 4, 7 } mod 12",
			" {0, 4, 7} mod 12",
			"{0, 4, 7} mod 12 ",
			"{0, 4, 7}  mod 12",
			"{0, 4, 7} mod  12",
			"{, 0} mod 12",
			"{0, } mod 12",
			"{70000} mod 65536",
			"{T} mod 12",
		}

		for _, input := range badInputs {
			s, err := pcset.Parse(input)
			requireSentinelError(t, err, pcset.ErrNotation)
			assert.False(t, s.IsValid())
			requirePanicsWithSentinel(t, func() { _ = pcset.MustParse(input) }, pcset.ErrNotation)
		}
	})
}

func TestLiteraryNotationsAndParseNotation(t *testing.T) {
	t.Parallel()

	t.Run("k26_canonical_formatting", func(t *testing.T) {
		t.Parallel()

		s := pcset.MustNew(12, 0, 4, 10, 11)

		assert.Equal(t, "{0, 4, 10, 11}", res(s.Notation(pitchclass.Decimal)).get(t))
		assert.Equal(t, "{0, 4, 10, 11}", s.MustNotation(pitchclass.Decimal))
		assert.Equal(t, "{0, 4, T, E}", res(s.Notation(pitchclass.TE)).get(t))
		assert.Equal(t, "{0, 4, T, E}", s.MustNotation(pitchclass.TE))
		assert.Equal(t, "{0, 4, t, e}", res(s.Notation(pitchclass.LowerTE)).get(t))
		assert.Equal(t, "{0, 4, t, e}", s.MustNotation(pitchclass.LowerTE))
		assert.Equal(t, "{0, 4, A, B}", res(s.Notation(pitchclass.AB)).get(t))
		assert.Equal(t, "{0, 4, A, B}", s.MustNotation(pitchclass.AB))

		assert.Equal(t, "{04TE}", res(s.CompactNotation(pitchclass.TE)).get(t))
		assert.Equal(t, "{04TE}", s.MustCompactNotation(pitchclass.TE))
		assert.Equal(t, "{04te}", res(s.CompactNotation(pitchclass.LowerTE)).get(t))
		assert.Equal(t, "{04te}", s.MustCompactNotation(pitchclass.LowerTE))
		assert.Equal(t, "{04AB}", res(s.CompactNotation(pitchclass.AB)).get(t))
		assert.Equal(t, "{04AB}", s.MustCompactNotation(pitchclass.AB))

		// CompactNotation in Decimal for N > 10 is forbidden even if all elements < 10.
		for _, mod := range []int{11, 12, 19} {
			for _, candidate := range []pcset.Set{
				pcset.MustEmpty(mod),
				pcset.MustReduce(mod, 0, 4, 7),
			} {
				_, err := candidate.CompactNotation(pitchclass.Decimal)
				requireSentinelError(t, err, pcset.ErrNotation)
				requirePanicsWithSentinel(t, func() {
					_ = candidate.MustCompactNotation(pitchclass.Decimal)
				}, pcset.ErrNotation)

				_, err = candidate.NormalFormCompactNotation(pitchclass.Decimal, pcset.Rahn)
				requireSentinelError(t, err, pcset.ErrNotation)
				requirePanicsWithSentinel(t, func() {
					_ = candidate.MustNormalFormCompactNotation(pitchclass.Decimal, pcset.Rahn)
				}, pcset.ErrNotation)
			}
		}

		// For N <= 10, CompactNotation in Decimal succeeds (G-12).
		s10 := pcset.MustNew(10, 0, 4, 7)
		assert.Equal(t, "{047}", res(s10.CompactNotation(pitchclass.Decimal)).get(t))
		assert.Equal(t, "{047}", s10.MustCompactNotation(pitchclass.Decimal))
		assert.Equal(t, "[470]", res(s10.NormalFormCompactNotation(pitchclass.Decimal, pcset.Rahn)).get(t))
		assert.Equal(t, "[470]", s10.MustNormalFormCompactNotation(pitchclass.Decimal, pcset.Rahn))
		assert.Equal(t, "{}", pcset.MustEmpty(10).MustCompactNotation(pitchclass.Decimal))
		assert.Equal(t, "[]", pcset.MustEmpty(10).MustNormalFormCompactNotation(pitchclass.Decimal, pcset.Rahn))
	})

	t.Run("normal_form_notations_rahn_and_forte", func(t *testing.T) {
		t.Parallel()

		s := pcset.MustNew(12, 2, 3, 7, 11)
		assert.Equal(t, "[11, 2, 3, 7]", res(s.NormalFormNotation(pitchclass.Decimal, pcset.Rahn)).get(t))
		assert.Equal(t, "[11, 2, 3, 7]", s.MustNormalFormNotation(pitchclass.Decimal, pcset.Rahn))
		assert.Equal(t, "[E, 2, 3, 7]", res(s.NormalFormNotation(pitchclass.TE, pcset.Rahn)).get(t))
		assert.Equal(t, "[E, 2, 3, 7]", s.MustNormalFormNotation(pitchclass.TE, pcset.Rahn))
		assert.Equal(t, "[E237]", res(s.NormalFormCompactNotation(pitchclass.TE, pcset.Rahn)).get(t))
		assert.Equal(t, "[E237]", s.MustNormalFormCompactNotation(pitchclass.TE, pcset.Rahn))

		div := pcset.MustNew(12, 0, 1, 3, 7, 8)
		assert.Equal(t, "[7, 8, 0, 1, 3]", div.MustNormalFormNotation(pitchclass.TE, pcset.Rahn))
		assert.Equal(t, "[78013]", div.MustNormalFormCompactNotation(pitchclass.TE, pcset.Rahn))
		assert.Equal(t, "[0, 1, 3, 7, 8]", div.MustNormalFormNotation(pitchclass.TE, pcset.Forte))
		assert.Equal(t, "[01378]", div.MustNormalFormCompactNotation(pitchclass.TE, pcset.Forte))

		empty := pcset.MustEmpty(12)
		assert.Equal(t, "{}", empty.MustNotation(pitchclass.TE))
		assert.Equal(t, "{}", empty.MustCompactNotation(pitchclass.TE))
		assert.Equal(t, "[]", empty.MustNormalFormNotation(pitchclass.TE, pcset.Rahn))
		assert.Equal(t, "[]", empty.MustNormalFormCompactNotation(pitchclass.TE, pcset.Forte))
	})

	t.Run("parse_notation_and_roundtrips", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name    string
			input   string
			style   pitchclass.Style
			modulus int
			want    pcset.Set
		}{
			{
				name:    "spaced_te_braces",
				input:   "{0, 4, T, E}",
				style:   pitchclass.TE,
				modulus: 12,
				want:    pcset.MustNew(12, 0, 4, 10, 11),
			},
			{
				name:    "compact_te_braces",
				input:   "{04TE}",
				style:   pitchclass.TE,
				modulus: 12,
				want:    pcset.MustNew(12, 0, 4, 10, 11),
			},
			{
				name:    "spaced_te_brackets",
				input:   "[E, 2, 3, 7]",
				style:   pitchclass.TE,
				modulus: 12,
				want:    pcset.MustNew(12, 2, 3, 7, 11),
			},
			{
				name:    "compact_te_brackets",
				input:   "[E237]",
				style:   pitchclass.TE,
				modulus: 12,
				want:    pcset.MustNew(12, 2, 3, 7, 11),
			},
			{
				name:    "unsorted_with_duplicates",
				input:   "{7, 0, 4, 0}",
				style:   pitchclass.TE,
				modulus: 12,
				want:    pcset.MustNew(12, 0, 4, 7),
			},
			{
				name:    "singleton_10_decimal_mod_12_g12",
				input:   "{10}",
				style:   pitchclass.Decimal,
				modulus: 12,
				want:    pcset.MustNew(12, 10),
			},
			{
				name:    "singleton_10_brackets_decimal_mod_12_g12",
				input:   "[10]",
				style:   pitchclass.Decimal,
				modulus: 12,
				want:    pcset.MustNew(12, 10),
			},
			{
				name:    "compact_047_decimal_mod_10_g12",
				input:   "{047}",
				style:   pitchclass.Decimal,
				modulus: 10,
				want:    pcset.MustNew(10, 0, 4, 7),
			},
			{
				name:    "empty_braces",
				input:   "{}",
				style:   pitchclass.TE,
				modulus: 12,
				want:    pcset.MustEmpty(12),
			},
			{
				name:    "empty_brackets",
				input:   "[]",
				style:   pitchclass.Decimal,
				modulus: 19,
				want:    pcset.MustEmpty(19),
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				got, err := pcset.ParseNotation(tc.input, tc.style, tc.modulus)
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
				assert.Equal(t, tc.want, pcset.MustParseNotation(tc.input, tc.style, tc.modulus))
			})
		}

		// Round-trip across all 4 styles and both algorithms for N = 12.
		sampleSets := []pcset.Set{
			pcset.MustEmpty(12),
			pcset.MustNew(12, 10),
			pcset.MustNew(12, 0, 4, 7),
			pcset.MustNew(12, 0, 1, 3, 7, 8),
			pcset.MustNew(12, 0, 4, 10, 11),
			pcset.MustUniverse(12),
		}
		styles := []pitchclass.Style{
			pitchclass.Decimal,
			pitchclass.TE,
			pitchclass.LowerTE,
			pitchclass.AB,
		}
		algs := []pcset.Algorithm{pcset.Rahn, pcset.Forte}

		for _, s := range sampleSets {
			for _, st := range styles {
				notStr := s.MustNotation(st)
				assert.Equal(t, s, pcset.MustParseNotation(notStr, st, 12))

				for _, alg := range algs {
					nfStr := s.MustNormalFormNotation(st, alg)
					assert.Equal(t, s, pcset.MustParseNotation(nfStr, st, 12))
				}

				if st != pitchclass.Decimal {
					compStr := s.MustCompactNotation(st)
					assert.Equal(t, s, pcset.MustParseNotation(compStr, st, 12))

					for _, alg := range algs {
						nfCompStr := s.MustNormalFormCompactNotation(st, alg)
						assert.Equal(t, s, pcset.MustParseNotation(nfCompStr, st, 12))
					}
				}
			}
		}
	})

	t.Run("notation_and_parse_notation_errors", func(t *testing.T) {
		t.Parallel()

		var zero pcset.Set
		valid12 := pcset.MustNew(12, 0, 4, 7)
		valid19 := pcset.MustNew(19, 0, 4, 7)
		badStyle := pitchclass.Style(99)
		badAlg := pcset.Algorithm(99)

		// Precedence on zero Set{}: ErrInvalidValue wins over badAlgorithm and badStyle.
		_, err := zero.Notation(badStyle)
		requireSentinelError(t, err, pcset.ErrInvalidValue)
		requirePanicsWithSentinel(t, func() { _ = zero.MustNotation(badStyle) }, pcset.ErrInvalidValue)

		_, err = zero.CompactNotation(badStyle)
		requireSentinelError(t, err, pcset.ErrInvalidValue)
		requirePanicsWithSentinel(t, func() { _ = zero.MustCompactNotation(badStyle) }, pcset.ErrInvalidValue)

		_, err = zero.NormalFormNotation(badStyle, badAlg)
		requireSentinelError(t, err, pcset.ErrInvalidValue)
		requirePanicsWithSentinel(t, func() { _ = zero.MustNormalFormNotation(badStyle, badAlg) }, pcset.ErrInvalidValue)

		_, err = zero.NormalFormCompactNotation(badStyle, badAlg)
		requireSentinelError(t, err, pcset.ErrInvalidValue)
		requirePanicsWithSentinel(t, func() { _ = zero.MustNormalFormCompactNotation(badStyle, badAlg) }, pcset.ErrInvalidValue)

		// Precedence on valid Set: ErrInvalidAlgorithm wins over badStyle / ErrNotation.
		_, err = valid12.NormalFormNotation(badStyle, badAlg)
		requireSentinelError(t, err, pcset.ErrInvalidAlgorithm)
		requirePanicsWithSentinel(t, func() { _ = valid12.MustNormalFormNotation(badStyle, badAlg) }, pcset.ErrInvalidAlgorithm)

		_, err = valid12.NormalFormCompactNotation(badStyle, badAlg)
		requireSentinelError(t, err, pcset.ErrInvalidAlgorithm)
		requirePanicsWithSentinel(t, func() { _ = valid12.MustNormalFormCompactNotation(badStyle, badAlg) }, pcset.ErrInvalidAlgorithm)

		// Invalid style or letter style at N > 12 -> ErrNotation.
		for _, st := range []pitchclass.Style{badStyle, pitchclass.TE, pitchclass.LowerTE, pitchclass.AB} {
			_, err := valid19.Notation(st)
			requireSentinelError(t, err, pcset.ErrNotation)
			requirePanicsWithSentinel(t, func() { _ = valid19.MustNotation(st) }, pcset.ErrNotation)

			_, err = valid19.CompactNotation(st)
			requireSentinelError(t, err, pcset.ErrNotation)
			requirePanicsWithSentinel(t, func() { _ = valid19.MustCompactNotation(st) }, pcset.ErrNotation)

			_, err = valid19.NormalFormNotation(st, pcset.Rahn)
			requireSentinelError(t, err, pcset.ErrNotation)
			requirePanicsWithSentinel(t, func() { _ = valid19.MustNormalFormNotation(st, pcset.Rahn) }, pcset.ErrNotation)

			_, err = valid19.NormalFormCompactNotation(st, pcset.Rahn)
			requireSentinelError(t, err, pcset.ErrNotation)
			requirePanicsWithSentinel(t, func() { _ = valid19.MustNormalFormCompactNotation(st, pcset.Rahn) }, pcset.ErrNotation)
		}

		// ParseNotation modulus errors take precedence.
		for _, tc := range []struct {
			name    string
			modulus int
			wantErr error
		}{
			{name: "zero_modulus", modulus: 0, wantErr: pcset.ErrInvalidModulus},
			{name: "negative_modulus", modulus: -12, wantErr: pcset.ErrInvalidModulus},
			{name: "modulus_65537", modulus: 65_537, wantErr: pcset.ErrModulusTooLarge},
		} {
			t.Run("parse_notation_"+tc.name, func(t *testing.T) {
				t.Parallel()

				_, err := pcset.ParseNotation("{0}", badStyle, tc.modulus)
				requireSentinelError(t, err, tc.wantErr)
				requirePanicsWithSentinel(t, func() {
					_ = pcset.MustParseNotation("{0}", badStyle, tc.modulus)
				}, tc.wantErr)
			})
		}

		// ParseNotation ErrNotation cases.
		badNotations := []struct {
			name    string
			input   string
			style   pitchclass.Style
			modulus int
		}{
			{name: "empty_string", input: "", style: pitchclass.TE, modulus: 12},
			{name: "mismatched_brace_bracket", input: "{0, 4, 7]", style: pitchclass.TE, modulus: 12},
			{name: "mismatched_bracket_brace", input: "[0, 4, 7}", style: pitchclass.TE, modulus: 12},
			{name: "no_delimiters", input: "0, 4, 7", style: pitchclass.TE, modulus: 12},
			{name: "no_space_after_comma", input: "{0,4,7}", style: pitchclass.Decimal, modulus: 12},
			{name: "extra_spaces_around_elements", input: " { 0 , 4 , 7 } ", style: pitchclass.Decimal, modulus: 12},
			{name: "empty_brackets_with_spaces", input: "[   ]", style: pitchclass.Decimal, modulus: 19},
			{name: "compact_decimal_at_mod_12", input: "{047}", style: pitchclass.Decimal, modulus: 12},
			{name: "compact_decimal_at_mod_11", input: "{047}", style: pitchclass.Decimal, modulus: 11},
			{name: "symbol_E_at_mod_11", input: "{0E}", style: pitchclass.TE, modulus: 11},
			{name: "symbol_E_comma_at_mod_11", input: "{0, E}", style: pitchclass.TE, modulus: 11},
			{name: "letter_style_at_mod_13", input: "{0, 4, 7}", style: pitchclass.TE, modulus: 13},
			{name: "letter_style_at_mod_19", input: "{0, 4, 7}", style: pitchclass.AB, modulus: 19},
			{name: "invalid_style_99", input: "{0, 4, 7}", style: badStyle, modulus: 12},
			{name: "empty_element_between_commas", input: "{0, , 4}", style: pitchclass.Decimal, modulus: 12},
			{name: "leading_comma", input: "{, 0}", style: pitchclass.Decimal, modulus: 12},
			{name: "trailing_comma", input: "{0, }", style: pitchclass.Decimal, modulus: 12},
			{name: "space_inside_compact", input: "{0 4 7}", style: pitchclass.TE, modulus: 12},
			{name: "leading_zero_singleton_mod_12", input: "{01}", style: pitchclass.Decimal, modulus: 12},
			{name: "out_of_range_mod_12", input: "{12}", style: pitchclass.Decimal, modulus: 12},
		}

		for _, tc := range badNotations {
			t.Run("bad_"+tc.name, func(t *testing.T) {
				t.Parallel()

				s, err := pcset.ParseNotation(tc.input, tc.style, tc.modulus)
				requireSentinelError(t, err, pcset.ErrNotation)
				assert.False(t, s.IsValid())
				requirePanicsWithSentinel(t, func() {
					_ = pcset.MustParseNotation(tc.input, tc.style, tc.modulus)
				}, pcset.ErrNotation)
			})
		}
	})
}
