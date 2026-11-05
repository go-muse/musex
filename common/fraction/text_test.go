package fraction_test

import (
	"encoding"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"testing"

	"github.com/go-muse/musex/common/fraction"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Compile-time interface implementation assertions.
var (
	_ fmt.Stringer             = fraction.Fraction{}
	_ encoding.TextMarshaler   = fraction.Fraction{}
	_ encoding.TextAppender    = fraction.Fraction{}
	_ encoding.TextUnmarshaler = (*fraction.Fraction)(nil)
)

// TestFormatting verifies criterion K-23 (String, ReducedString, CompactString, MixedString, Decimal, FormatDecimal).
func TestFormatting(t *testing.T) {
	t.Parallel()

	t.Run("String_ReducedString_CompactString_MixedString", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name        string
			f           fraction.Fraction
			wantStr     string
			wantReduced string
			wantCompact string
			wantMixed   string
		}{
			{
				name:        "invalid_0/0",
				f:           fraction.Fraction{},
				wantStr:     "0/0",
				wantReduced: "0/0",
				wantCompact: "0/0",
				wantMixed:   "0/0",
			},
			{
				name:        "6/12",
				f:           fraction.MustNew(6, 12),
				wantStr:     "6/12",
				wantReduced: "1/2",
				wantCompact: "1/2",
				wantMixed:   "1/2",
			},
			{
				name:        "1/-2",
				f:           fraction.MustNew(1, -2),
				wantStr:     "1/-2",
				wantReduced: "-1/2",
				wantCompact: "-1/2",
				wantMixed:   "-1/2",
			},
			{
				name:        "0/3",
				f:           fraction.MustNew(0, 3),
				wantStr:     "0/3",
				wantReduced: "0/1",
				wantCompact: "0",
				wantMixed:   "0",
			},
			{
				name:        "2/1",
				f:           fraction.MustNew(2, 1),
				wantStr:     "2/1",
				wantReduced: "2/1",
				wantCompact: "2",
				wantMixed:   "2",
			},
			{
				name:        "6/3",
				f:           fraction.MustNew(6, 3),
				wantStr:     "6/3",
				wantReduced: "2/1",
				wantCompact: "2",
				wantMixed:   "2",
			},
			{
				name:        "-4/2",
				f:           fraction.MustNew(-4, 2),
				wantStr:     "-4/2",
				wantReduced: "-2/1",
				wantCompact: "-2",
				wantMixed:   "-2",
			},
			{
				name:        "5/4",
				f:           fraction.MustNew(5, 4),
				wantStr:     "5/4",
				wantReduced: "5/4",
				wantCompact: "5/4",
				wantMixed:   "1 1/4",
			},
			{
				name:        "-5/4",
				f:           fraction.MustNew(-5, 4),
				wantStr:     "-5/4",
				wantReduced: "-5/4",
				wantCompact: "-5/4",
				wantMixed:   "-1 1/4",
			},
			{
				name:        "18/8",
				f:           fraction.MustNew(18, 8),
				wantStr:     "18/8",
				wantReduced: "9/4",
				wantCompact: "9/4",
				wantMixed:   "2 1/4",
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tc.wantStr, tc.f.String())
				assert.Equal(t, tc.wantReduced, tc.f.ReducedString())
				assert.Equal(t, tc.wantCompact, tc.f.CompactString())
				assert.Equal(t, tc.wantMixed, tc.f.MixedString())
			})
		}
	})

	t.Run("Decimal_and_FormatDecimal", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name   string
			f      fraction.Fraction
			digits int
			mode   fraction.RoundingMode
			want   string
		}{
			{
				name:   "1/8_3digits",
				f:      fraction.MustNew(1, 8),
				digits: 3,
				mode:   fraction.RoundHalfAwayFromZero,
				want:   "0.125",
			},
			{
				name:   "-1/8_3digits_negative_zero_integer_part",
				f:      fraction.MustNew(-1, 8),
				digits: 3,
				mode:   fraction.RoundHalfAwayFromZero,
				want:   "-0.125",
			},
			{
				name:   "-1/200_3digits_leading_zeros_in_frac",
				f:      fraction.MustNew(-1, 200),
				digits: 3,
				mode:   fraction.RoundHalfAwayFromZero,
				want:   "-0.005",
			},
			{
				name:   "1/3_4digits",
				f:      fraction.MustNew(1, 3),
				digits: 4,
				mode:   fraction.RoundHalfAwayFromZero,
				want:   "0.3333",
			},
			{
				name:   "2/3_4digits_HalfAway",
				f:      fraction.MustNew(2, 3),
				digits: 4,
				mode:   fraction.RoundHalfAwayFromZero,
				want:   "0.6667",
			},
			{
				name:   "2/3_4digits_Trunc",
				f:      fraction.MustNew(2, 3),
				digits: 4,
				mode:   fraction.RoundTrunc,
				want:   "0.6666",
			},
			{
				name:   "-1/10000_2digits_no_minus_zero",
				f:      fraction.MustNew(-1, 10000),
				digits: 2,
				mode:   fraction.RoundHalfAwayFromZero,
				want:   "0.00",
			},
			{
				name:   "0/3_0digits",
				f:      fraction.MustNew(0, 3),
				digits: 0,
				mode:   fraction.RoundHalfAwayFromZero,
				want:   "0",
			},
			{
				name:   "3/2_0digits_HalfAway",
				f:      fraction.MustNew(3, 2),
				digits: 0,
				mode:   fraction.RoundHalfAwayFromZero,
				want:   "2",
			},
			{
				name:   "1/2_0digits_HalfToEven",
				f:      fraction.MustNew(1, 2),
				digits: 0,
				mode:   fraction.RoundHalfToEven,
				want:   "0",
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				got, err := tc.f.FormatDecimal(tc.digits, tc.mode)
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
				assert.Equal(t, tc.want, tc.f.MustFormatDecimal(tc.digits, tc.mode))

				if tc.mode == fraction.RoundHalfAwayFromZero {
					gotDec, err := tc.f.Decimal(tc.digits)
					require.NoError(t, err)
					assert.Equal(t, tc.want, gotDec)
					assert.Equal(t, tc.want, tc.f.MustDecimal(tc.digits))
				}
			})
		}

		// Invalid digits or mode -> ErrNotRepresentable.
		oneHalf := fraction.MustNew(1, 2)
		for _, badDigits := range []int{-1, -10, fraction.MaxDigits + 1, 1000} {
			_, err := oneHalf.Decimal(badDigits)
			require.ErrorIs(t, err, fraction.ErrNotRepresentable)
			_, err = oneHalf.FormatDecimal(badDigits, fraction.RoundHalfAwayFromZero)
			require.ErrorIs(t, err, fraction.ErrNotRepresentable)

			assertMustPanicsWithError(t, func() {
				_ = oneHalf.MustDecimal(badDigits)
			}, fraction.ErrNotRepresentable)
			assertMustPanicsWithError(t, func() {
				_ = oneHalf.MustFormatDecimal(badDigits, fraction.RoundHalfAwayFromZero)
			}, fraction.ErrNotRepresentable)
		}

		_, err := oneHalf.FormatDecimal(2, fraction.RoundingMode(99))
		require.ErrorIs(t, err, fraction.ErrNotRepresentable)
	})
}

// TestTextInterfacesAndParse verifies criterion K-23 (Parse, MustParse, MarshalText, AppendText, UnmarshalText, error length limit).
func TestTextInterfacesAndParse(t *testing.T) {
	t.Parallel()

	t.Run("Parse_valid_inputs", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			input   string
			wantNum int
			wantDen int
		}{
			{input: "6/12", wantNum: 6, wantDen: 12},
			{input: "1/-2", wantNum: 1, wantDen: -2},
			{input: "-2/-4", wantNum: -2, wantDen: -4},
			{input: "+3/+4", wantNum: 3, wantDen: 4},
			{input: "0/3", wantNum: 0, wantDen: 3},
			{input: "5", wantNum: 5, wantDen: 1},
			{input: "-3", wantNum: -3, wantDen: 1},
			{input: "+42", wantNum: 42, wantDen: 1},
			{input: "0", wantNum: 0, wantDen: 1},
			{input: "-0", wantNum: 0, wantDen: 1},
			{input: "1 1/4", wantNum: 5, wantDen: 4},
			{input: "-1 1/4", wantNum: -5, wantDen: 4},
			{input: "2 2/8", wantNum: 18, wantDen: 8},
			{input: "+2 2/8", wantNum: 18, wantDen: 8},
			{input: "0 3/8", wantNum: 3, wantDen: 8},
			{input: "-0 1/4", wantNum: -1, wantDen: 4},
			{input: "7.1", wantNum: 71, wantDen: 10},
			{input: "-0.25", wantNum: -1, wantDen: 4},
			{input: "+0.125", wantNum: 1, wantDen: 8},
			{input: "-0.00", wantNum: 0, wantDen: 1},
			{input: "2.00", wantNum: 2, wantDen: 1},
		}

		for _, tc := range cases {
			t.Run(tc.input, func(t *testing.T) {
				t.Parallel()

				got, err := fraction.Parse(tc.input)
				require.NoError(t, err)
				assertExactPairInt(t, got, tc.wantNum, tc.wantDen)
				assert.Equal(t, got, fraction.MustParse(tc.input))

				var unmarshaled fraction.Fraction
				require.NoError(t, unmarshaled.UnmarshalText([]byte(tc.input)))
				assert.Equal(t, got, unmarshaled)
			})
		}
	})

	t.Run("round_trip_on_S_and_Marshal_Append", func(t *testing.T) {
		t.Parallel()

		for _, p := range sampleSetS() {
			f := fraction.MustNew(p.num, p.den)
			s := f.String()

			parsed, err := fraction.Parse(s)
			require.NoError(t, err)
			assert.True(t, parsed.EqualExact(f))
			assert.Equal(t, f, parsed)

			textBytes, err := f.MarshalText()
			require.NoError(t, err)
			assert.Equal(t, s, string(textBytes))

			prefix := []byte("prefix:")
			buf := make([]byte, len(prefix), len(prefix)+32)
			copy(buf, prefix)
			appended, err := f.AppendText(buf)
			require.NoError(t, err)
			assert.Equal(t, "prefix:"+s, string(appended))
			assert.Equal(t, "prefix:", string(buf), "AppendText must not modify existing prefix bytes")
		}
	})

	t.Run("Parse_ErrInvalidText", func(t *testing.T) {
		t.Parallel()

		invalidInputs := []string{
			"",
			" ",
			" 1/2",
			"1/2 ",
			"\t1/2",
			"1/2\n",
			"1/",
			"/2",
			"/",
			"1/2/3",
			"1 / 2",
			"1  1/4",
			"1 -1/4",
			"1 +1/4",
			"1 1/-4",
			"1 1/+4",
			"1 1",
			"7.",
			".5",
			".",
			"+.",
			"-.",
			"1.2.3",
			"1.-5",
			"1.+5",
			"abc",
			"1/2a",
			"--1/2",
			"++1/2",
		}

		for _, input := range invalidInputs {
			_, err := fraction.Parse(input)
			require.ErrorIs(t, err, fraction.ErrInvalidText, "input=%q", input)
			assert.False(t, errors.Is(err, fraction.ErrZeroDenominator), "input=%q", input)
			assert.False(t, errors.Is(err, fraction.ErrOutOfRange), "input=%q", input)

			assertMustPanicsWithError(t, func() {
				_ = fraction.MustParse(input)
			}, fraction.ErrInvalidText)
		}
	})

	t.Run("Parse_ErrZeroDenominator", func(t *testing.T) {
		t.Parallel()

		zeroDenInputs := []string{
			"1/0",
			"0/0",
			"-5/0",
			"1/-0",
			"1/+0",
			"1 1/0",
			"-2 0/0",
			"0 1/000",
		}

		for _, input := range zeroDenInputs {
			_, err := fraction.Parse(input)
			require.ErrorIs(t, err, fraction.ErrZeroDenominator, "input=%q", input)
			assert.False(t, errors.Is(err, fraction.ErrInvalidText), "input=%q must wrap only ErrZeroDenominator", input)
			assert.False(t, errors.Is(err, fraction.ErrOutOfRange), "input=%q", input)

			assertMustPanicsWithError(t, func() {
				_ = fraction.MustParse(input)
			}, fraction.ErrZeroDenominator)
		}
	})

	t.Run("Parse_ErrOutOfRange", func(t *testing.T) {
		t.Parallel()

		zeros101 := strings.Repeat("0", fraction.MaxDigits+1)
		nines80 := strings.Repeat("9", 80)

		outOfRangeInputs := []string{
			zeros101,
			zeros101 + "/2",
			"1/" + zeros101, // len(DIGITS) > MaxDigits is checked before b == 0!
			"0." + zeros101,
			zeros101 + ".5",
			"1 " + zeros101 + "/2",
			"1 1/" + zeros101,
			nines80,
			nines80 + "/1",
			"1/" + nines80,
			"1 1/" + nines80,
			fraction.MaxComponentString + " 1/2",
		}

		for _, input := range outOfRangeInputs {
			_, err := fraction.Parse(input)
			require.ErrorIs(t, err, fraction.ErrOutOfRange, "input len=%d", len(input))
			assert.False(t, errors.Is(err, fraction.ErrInvalidText))
			assert.False(t, errors.Is(err, fraction.ErrZeroDenominator))
		}
	})

	t.Run("UnmarshalText_receiver_preservation_and_nil_receiver", func(t *testing.T) {
		t.Parallel()

		var nilPtr *fraction.Fraction
		err := nilPtr.UnmarshalText([]byte("1/2"))
		require.ErrorIs(t, err, fraction.ErrInvalid)

		orig := fraction.MustNew(6, 12)
		target := orig
		err = target.UnmarshalText([]byte("invalid"))
		require.ErrorIs(t, err, fraction.ErrInvalidText)
		assert.Equal(t, orig, target, "receiver must remain unchanged on UnmarshalText error")

		err = target.UnmarshalText([]byte("1/0"))
		require.ErrorIs(t, err, fraction.ErrZeroDenominator)
		assert.Equal(t, orig, target)
	})

	t.Run("long_invalid_inputs_bounded_error_message", func(t *testing.T) {
		t.Parallel()

		for _, size := range []int{1024, 1024 * 1024} {
			longSyntax := strings.Repeat("a", size)
			_, err := fraction.Parse(longSyntax)
			require.ErrorIs(t, err, fraction.ErrInvalidText)
			assert.LessOrEqual(t, len(err.Error()), 256, "error message must be bounded on %d-byte input", size)

			longDigits := strings.Repeat("9", size)
			_, err = fraction.Parse(longDigits)
			require.ErrorIs(t, err, fraction.ErrOutOfRange)
			assert.LessOrEqual(t, len(err.Error()), 256, "error message must be bounded on %d-byte digits input", size)
		}
	})
}

// TestSentinelErrorsAndAllMustWrappers verifies criteria K-24 (complete) and K-26.
func TestSentinelErrorsAndAllMustWrappers(t *testing.T) {
	t.Parallel()

	t.Run("pairwise_distinct_sentinel_errors", func(t *testing.T) {
		t.Parallel()

		sentinels := []struct {
			name string
			err  error
		}{
			{name: "ErrInvalid", err: fraction.ErrInvalid},
			{name: "ErrZeroDenominator", err: fraction.ErrZeroDenominator},
			{name: "ErrDivisionByZero", err: fraction.ErrDivisionByZero},
			{name: "ErrNotRepresentable", err: fraction.ErrNotRepresentable},
			{name: "ErrOutOfRange", err: fraction.ErrOutOfRange},
			{name: "ErrInvalidText", err: fraction.ErrInvalidText},
		}

		for i, a := range sentinels {
			require.NotNil(t, a.err, "%s must not be nil", a.name)
			for j, b := range sentinels {
				if i == j {
					assert.ErrorIs(t, a.err, b.err)
				} else {
					assert.False(t, errors.Is(a.err, b.err), "%s must not match %s", a.name, b.name)
				}
			}
		}
	})

	t.Run("all_Must_wrappers_panic_on_invalid_and_error_inputs", func(t *testing.T) {
		t.Parallel()

		var z fraction.Fraction
		v := fraction.MustNew(1, 2)

		// MarshalText and AppendText on invalid fraction.
		mBytes, err := z.MarshalText()
		require.ErrorIs(t, err, fraction.ErrInvalid)
		assert.Nil(t, mBytes)

		prefix := []byte("keep")
		appBytes, err := z.AppendText(prefix)
		require.ErrorIs(t, err, fraction.ErrInvalid)
		assert.Equal(t, prefix, appBytes)

		_, err = z.Decimal(2)
		require.ErrorIs(t, err, fraction.ErrInvalid)
		_, err = z.FormatDecimal(2, fraction.RoundHalfAwayFromZero)
		require.ErrorIs(t, err, fraction.ErrInvalid)

		// Verify every Must* method on invalid fraction panics with ErrInvalid.
		mustInvalidCalls := []struct {
			name string
			fn   func()
		}{
			{"MustWithDenominator", func() { _ = z.MustWithDenominator(2) }},
			{"MustWithBigDenominator", func() { _ = z.MustWithBigDenominator(big.NewInt(2)) }},
			{"MustNeg", func() { _ = z.MustNeg() }},
			{"MustAbs", func() { _ = z.MustAbs() }},
			{"MustReciprocal", func() { _ = z.MustReciprocal() }},
			{"MustInc", func() { _ = z.MustInc() }},
			{"MustDec", func() { _ = z.MustDec() }},
			{"MustStepNumerator", func() { _ = z.MustStepNumerator(1) }},
			{"MustPow", func() { _ = z.MustPow(2) }},
			{"MustAdd_recv", func() { _ = z.MustAdd(v) }},
			{"MustAdd_arg", func() { _ = v.MustAdd(z) }},
			{"MustSub_recv", func() { _ = z.MustSub(v) }},
			{"MustSub_arg", func() { _ = v.MustSub(z) }},
			{"MustMul_recv", func() { _ = z.MustMul(v) }},
			{"MustMul_arg", func() { _ = v.MustMul(z) }},
			{"MustDiv_recv", func() { _ = z.MustDiv(v) }},
			{"MustDiv_arg", func() { _ = v.MustDiv(z) }},
			{"MustAddInt", func() { _ = z.MustAddInt(1) }},
			{"MustSubInt", func() { _ = z.MustSubInt(1) }},
			{"MustMulInt", func() { _ = z.MustMulInt(2) }},
			{"MustDivInt", func() { _ = z.MustDivInt(2) }},
			{"MustAddWith", func() { _ = z.MustAddWith(v, fraction.DenomReduced) }},
			{"MustSubWith", func() { _ = z.MustSubWith(v, fraction.DenomReduced) }},
			{"MustMulWith", func() { _ = z.MustMulWith(v, fraction.DenomReduced) }},
			{"MustDivWith", func() { _ = z.MustDivWith(v, fraction.DenomReduced) }},
			{"MustAddInDenominator", func() { _ = z.MustAddInDenominator(v, 4) }},
			{"MustSubInDenominator", func() { _ = z.MustSubInDenominator(v, 4) }},
			{"MustMulInDenominator", func() { _ = z.MustMulInDenominator(v, 4) }},
			{"MustDivInDenominator", func() { _ = z.MustDivInDenominator(v, 4) }},
			{"MustSplit", func() { _, _ = z.MustSplit() }},
			{"MustSplitBig", func() { _, _ = z.MustSplitBig() }},
			{"MustSplitFloor", func() { _, _ = z.MustSplitFloor() }},
			{"MustSplitFloorBig", func() { _, _ = z.MustSplitFloorBig() }},
			{"MustQuoRem", func() { _, _ = z.MustQuoRem(v) }},
			{"MustQuoRemBig", func() { _, _ = z.MustQuoRemBig(v) }},
			{"MustDivModFloor", func() { _, _ = z.MustDivModFloor(v) }},
			{"MustDivModFloorBig", func() { _, _ = z.MustDivModFloorBig(v) }},
			{"MustScaleIntExact", func() { _, _ = z.MustScaleIntExact(2) }},
			{"MustIsMultipleOf", func() { _ = z.MustIsMultipleOf(v) }},
			{"MustInProgression", func() { _, _ = z.MustInProgression(v, v) }},
			{"MustRoundToInt", func() { _ = z.MustRoundToInt(fraction.RoundHalfAwayFromZero) }},
			{"MustRoundToBigInt", func() { _ = z.MustRoundToBigInt(fraction.RoundHalfAwayFromZero) }},
			{"MustRoundToDenominator", func() { _ = z.MustRoundToDenominator(4, fraction.RoundHalfAwayFromZero) }},
			{"MustRoundToBigDenominator", func() { _ = z.MustRoundToBigDenominator(big.NewInt(4), fraction.RoundHalfAwayFromZero) }},
			{"MustDecimal", func() { _ = z.MustDecimal(2) }},
			{"MustFormatDecimal", func() { _ = z.MustFormatDecimal(2, fraction.RoundHalfAwayFromZero) }},
		}

		for _, tc := range mustInvalidCalls {
			assertMustPanicsWithError(t, tc.fn, fraction.ErrInvalid)
		}
	})
}

// TestConcurrencyAndImmutability verifies criterion K-27 (concurrent safety and operand immutability).
func TestConcurrencyAndImmutability(t *testing.T) {
	t.Parallel()

	s := sampleSetS()
	fractions := make([]fraction.Fraction, len(s))
	for i, p := range s {
		fractions[i] = fraction.MustNew(p.num, p.den)
	}

	const workers = 16
	var wg sync.WaitGroup
	errCh := make(chan error, workers)

	for w := range workers {
		wg.Go(func() {
			for i, a := range fractions {
				b := fractions[(i+w)%len(fractions)]

				_ = a.Reduce()
				_ = a.Equal(b)
				_ = a.EqualExact(b)
				_ = a.Compare(b)
				_ = a.CompareExact(b)
				_ = a.IsValid()
				_ = a.IsZero()
				_ = a.IsOne()
				_ = a.IsInteger()
				_ = a.IsProper()
				_ = a.IsReduced()
				_ = a.Sign()

				if _, err := a.Add(b); err != nil {
					errCh <- err
					return
				}
				if _, err := a.Sub(b); err != nil {
					errCh <- err
					return
				}
				if _, err := a.Mul(b); err != nil {
					errCh <- err
					return
				}
				if !b.IsZero() {
					if _, err := a.Div(b); err != nil {
						errCh <- err
						return
					}
					if _, _, err := a.QuoRem(b); err != nil {
						errCh <- err
						return
					}
				}
				if _, _, err := a.Split(); err != nil {
					errCh <- err
					return
				}
				if _, _, err := a.SplitFloor(); err != nil {
					errCh <- err
					return
				}
				if _, err := a.RoundToInt(fraction.RoundHalfAwayFromZero); err != nil {
					errCh <- err
					return
				}
				if _, err := a.Decimal(3); err != nil {
					errCh <- err
					return
				}
				str := a.String()
				parsed, err := fraction.Parse(str)
				if err != nil {
					errCh <- err
					return
				}
				if !parsed.EqualExact(a) {
					errCh <- fmt.Errorf("round-trip mismatch: got %v, want %v", parsed, a)
					return
				}
			}
		})
	}

	wg.Wait()
	close(errCh)
	for err := range errCh {
		require.NoError(t, err)
	}

	// Verify none of the shared fractions were mutated.
	for i, p := range s {
		assertExactPairInt(t, fractions[i], p.num, p.den)
	}
}

// TestZeroAllocs verifies criterion K-27 (0 heap allocations on core value operations).
// Not parallel: testing.AllocsPerRun measures process-wide heap allocations.
func TestZeroAllocs(t *testing.T) {
	a := fraction.MustNew(6, 12)
	b := fraction.MustNew(1, 3)

	allocs := testing.AllocsPerRun(100, func() {
		z := fraction.Zero()
		o := fraction.One()
		m := fraction.MinusOne()
		f, _ := fraction.New(3, 4)
		r, _ := fraction.NewReduced(6, 12)
		fi := fraction.FromInt(5)

		_, _ = a.Numerator()
		_, _ = a.Denominator()
		_, _, _ = a.Pair()
		_, _, _ = a.ReducedPair()
		_, _ = fi.Int()
		red := a.Reduce()
		_, _ = a.WithDenominator(24)

		_ = a.Equal(b)
		_ = a.EqualExact(b)
		_ = a.Compare(b)
		_ = a.CompareExact(b)
		_ = a.IsValid()
		_ = z.IsZero()
		_ = o.IsOne()
		_ = m.IsInteger()
		_ = f.IsProper()
		_ = r.IsReduced()
		_ = a.Sign()

		_, _ = a.Neg()
		_, _ = a.Abs()
		_, _ = a.Reciprocal()
		_, _ = a.Inc()
		_, _ = a.Dec()
		_, _ = a.StepNumerator(1)
		_, _ = a.Add(b)
		_, _ = a.Sub(b)
		_, _ = a.Mul(b)
		_, _ = a.Div(b)
		_, _ = a.AddWith(b, fraction.DenomUnreducedLCM)
		_ = red
	})

	assert.Equal(t, 0.0, allocs)
}

// FuzzParse verifies criterion K-23 and K-27 on arbitrary string inputs.
func FuzzParse(f *testing.F) {
	seeds := []string{
		"0/1",
		"6/12",
		"1/-2",
		"-2/-4",
		"5",
		"-3",
		"1 1/4",
		"-1 1/4",
		"2 2/8",
		"7.1",
		"-0.25",
		"",
		"1/0",
		"0/0",
		"1 1/0",
		"1 -1/4",
		"7.",
		".5",
		"abc",
		fraction.MaxComponentString + "/1",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		got, err := fraction.Parse(input)
		if err != nil {
			assert.False(t, got.IsValid())
			assert.LessOrEqual(t, len(err.Error()), 256)

			matchCount := 0
			for _, target := range []error{
				fraction.ErrInvalidText,
				fraction.ErrZeroDenominator,
				fraction.ErrOutOfRange,
			} {
				if errors.Is(err, target) {
					matchCount++
				}
			}
			assert.Equal(t, 1, matchCount, "Parse error %v must wrap exactly one sentinel error", err)
			return
		}

		require.True(t, got.IsValid())
		roundTrip, err := fraction.Parse(got.String())
		require.NoError(t, err)
		assert.True(t, roundTrip.EqualExact(got))
	})
}

// FuzzArithmetic verifies criteria K-13..K-22 against math/big.Rat on arbitrary 64-bit integer pairs.
func FuzzArithmetic(f *testing.F) {
	f.Add(int64(1), int64(3), int64(1), int64(6))
	f.Add(int64(6), int64(12), int64(1), int64(-2))
	f.Add(int64(0), int64(3), int64(-13), int64(12))
	f.Add(int64(14), int64(12), int64(3), int64(4))

	f.Fuzz(func(t *testing.T, n1, d1, n2, d2 int64) {
		if d1 == 0 || d2 == 0 {
			return
		}

		a, err := fraction.New(int(n1), int(d1))
		require.NoError(t, err)
		b, err := fraction.New(int(n2), int(d2))
		require.NoError(t, err)

		ra := new(big.Rat).SetFrac(big.NewInt(n1), big.NewInt(d1))
		rb := new(big.Rat).SetFrac(big.NewInt(n2), big.NewInt(d2))

		assert.Equal(t, ra.Cmp(rb), a.Compare(b))
		assert.Equal(t, ra.Cmp(rb) == 0, a.Equal(b))

		addGot, err := a.Add(b)
		require.NoError(t, err)
		addRat, ok := addGot.BigRat()
		require.True(t, ok)
		assert.Equal(t, 0, addRat.Cmp(new(big.Rat).Add(ra, rb)))

		subGot, err := a.Sub(b)
		require.NoError(t, err)
		subRat, ok := subGot.BigRat()
		require.True(t, ok)
		assert.Equal(t, 0, subRat.Cmp(new(big.Rat).Sub(ra, rb)))

		mulGot, err := a.Mul(b)
		require.NoError(t, err)
		mulRat, ok := mulGot.BigRat()
		require.True(t, ok)
		assert.Equal(t, 0, mulRat.Cmp(new(big.Rat).Mul(ra, rb)))

		if n2 != 0 {
			divGot, err := a.Div(b)
			require.NoError(t, err)
			divRat, ok := divGot.BigRat()
			require.True(t, ok)
			assert.Equal(t, 0, divRat.Cmp(new(big.Rat).Quo(ra, rb)))
		}
	})
}
