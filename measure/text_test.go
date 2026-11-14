package measure_test

import (
	"encoding/json/v2"
	"errors"
	"math/big"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/meter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Group Т-8, text: texts of enumerations and records, the text and the
// parsing of grids and addresses, JSON, and the texts of parse errors
// (architecture 4.3, 4.13, 4.16; О-22; П-16; КП-38, КП-46).

// canonicalGrids are texts of architecture 4.16 and КП-38 with the
// operations that build the same record.
var canonicalGrids = []string{
	"4/4",
	"3/4;pickup:1/4",
	"4/4;3:3/4;4:length:1/2;bars:8",
	"4/4;5:length:1/4:end",
	"4/4;5:4/4",
	"C",
	"C|;2:C",
	"2/4+3/8;pickup:1/8",
	"(3+2+3)/8;2:7/8[3+2+2];3:5/8[2+3]",
	"4/4;3:3/4;3:length:1/2",
	"4/4;pickup:1/1;1:length:1/2:end;2:length:5/2;bars:2",
	"4/4;pickup:1/4;3:3/4;4:length:1/2;bars:9223372036854775807",
	"4/3;9223372036854775807:4/4",
	"4/4;1:length:1/3",
	"6/8;pickup:1/8;3:9/8;5:length:1/16:end;7:length:7/4;bars:12",
}

// TestTextGrid covers О-22, П-16, and КП-38: the canonical text of a grid
// by String, MarshalText, and AppendText, and parsing back to an equal grid.
func TestTextGrid(t *testing.T) {
	t.Parallel()

	for _, text := range canonicalGrids {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := gridOf(t, text)
			assert.Equal(t, text, g.String())

			b, err := g.MarshalText()
			require.NoError(t, err)
			assert.Equal(t, text, string(b))

			out, err := g.AppendText([]byte("grid="))
			require.NoError(t, err)
			assert.Equal(t, "grid="+text, string(out))

			p, err := measure.Parse(text)
			require.NoError(t, err)
			assert.True(t, p.Equal(g), "Parse(String()) equals by record")
			assert.Equal(t, recordOf(t, g), recordOf(t, p))
			assert.Equal(t, text, p.String())

			var u measure.Grid
			require.NoError(t, u.UnmarshalText(b))
			assert.True(t, u.Equal(g), "UnmarshalText(MarshalText()) equals by record")
		})
	}
}

// TestTextGridDistinct covers П-16 and КП-38: the texts of grids that
// differ by record differ.
func TestTextGridDistinct(t *testing.T) {
	t.Parallel()
	defer guard(t)

	texts := append([]string{}, canonicalGrids...)
	for _, c := range equalCases() {
		texts = append(texts, c.a, c.b)
	}

	grids := make([]measure.Grid, len(texts))
	for i, s := range texts {
		grids[i] = gridOf(t, s)
	}

	for i := range grids {
		for j := range grids {
			same := grids[i].Equal(grids[j])
			assert.Equal(t, same, grids[i].String() == grids[j].String(), "%s and %s", texts[i], texts[j])
		}
	}
}

// TestTextGridNonCanonical covers architecture 4.16: nested texts accept
// the noncanonical forms of meter and duration.
func TestTextGridNonCanonical(t *testing.T) {
	t.Parallel()

	tests := []struct {
		text string
		want string
	}{
		{"3/4;pickup:2/8", "3/4;pickup:1/4"},
		{"¢", "C|"},
		{"4/4;5:¢", "4/4;5:C|"},
		{"6/8[3+3]", "6/8"},
		{"4/4;3:length:2/4", "4/4;3:length:1/2"},
		{"4/4;3:length:4/4:end", "4/4;3:length:1/1:end"},
		{"4/4;3:6/8[3+3]", "4/4;3:6/8"},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g, err := measure.Parse(tt.text)
			require.NoError(t, err)
			assert.Equal(t, tt.want, g.String())
			assert.True(t, g.Equal(gridOf(t, tt.want)))
		})
	}
}

// excerpt is the quotation of a rejected input of architecture 4.16: the
// whole input up to 64 bytes, otherwise its longest prefix of whole UTF-8
// decoding units within 64 bytes followed by "...".
func excerpt(x string) string {
	if len(x) <= 64 {
		return strconv.Quote(x)
	}

	k := 0
	for k < len(x) {
		_, size := utf8.DecodeRuneInString(x[k:])
		if k+size > 64 {
			break
		}

		k += size
	}

	return strconv.Quote(x[:k]) + "..."
}

// assertParseMessage checks the text of a parse error: it quotes the
// excerpt of the input and is at most 512 bytes long.
func assertParseMessage(t *testing.T, err error, input string) {
	t.Helper()

	if err == nil {
		return
	}

	msg := err.Error()
	assert.Contains(t, msg, excerpt(input))
	assert.LessOrEqual(t, len(msg), 512, "the error text is at most 512 bytes")
}

// rangeCauses returns the range sentinels of the dependencies that a parse
// error of a dependency wraps (architecture 4.2, Е-10 of a parse).
func rangeCauses(err error) []error {
	var out []error

	for _, s := range []error{meter.ErrOutOfRange, duration.ErrOutOfRange, fraction.ErrOutOfRange} {
		if errors.Is(err, s) {
			out = append(out, s)
		}
	}

	return out
}

// hugeDigits is 2^256 in decimal: greater than fraction.MaxComponent.
func hugeDigits() string { return pow2(256).String() }

// meterOutOfRange is a sum of signatures whose size is outside the range of
// common/fraction: five parts 1/p with pairwise coprime Mersenne
// denominators.
func meterOutOfRange() string {
	parts := make([]string, 0, 5)
	for _, p := range []int{61, 59, 53, 47, 43} {
		d := new(big.Int).Sub(pow2(p), big.NewInt(1))
		parts = append(parts, "1/"+d.String())
	}

	return strings.Join(parts, "+")
}

// TestTextParseErrors covers О-22, КП-46, and the order of parse failures
// of architecture 4.16: Е-9 for the form, Е-10 for a number greater than
// math.MaxInt, Е-9 for the canonical order, nested texts from left to right,
// then building: Е-3 and Е-4. A parse error wraps only its sentinel and the
// named range sentinels.
func TestTextParseErrors(t *testing.T) {
	t.Parallel()

	huge := hugeDigits()
	big9 := "99999999999999999999"

	_, durationErr := duration.Parse("1/" + huge)
	_, meterRangeErr := meter.Parse("3/" + big9)
	_, meterSumErr := meter.Parse(meterOutOfRange())

	tests := []struct {
		name   string
		text   string
		err    error
		causes []error
	}{
		{"empty", "", measure.ErrInvalidText, nil},
		{"КП-46 unset", "Grid(unset)", measure.ErrInvalidText, nil},
		{"semicolon", ";", measure.ErrInvalidText, nil},
		{"trailing semicolon", "4/4;", measure.ErrInvalidText, nil},
		{"leading semicolon", ";4/4", measure.ErrInvalidText, nil},
		{"leading space", " 4/4", measure.ErrInvalidText, nil},
		{"trailing space", "4/4 ", measure.ErrInvalidText, nil},
		{"space inside", "4/4; 3:3/4", measure.ErrInvalidText, nil},
		{"invalid meter", "x", measure.ErrInvalidText, nil},
		{"meter unset text", "Meter(unset)", measure.ErrInvalidText, nil},
		{"meter too many units", "4/4;3:999999999999999999999/4", measure.ErrInvalidText, nil},
		{"empty pickup", "4/4;pickup:", measure.ErrInvalidText, nil},
		{"pickup zero", "4/4;pickup:0/1", measure.ErrInvalidText, nil},
		{"pickup integer form", "4/4;pickup:1", measure.ErrInvalidText, nil},
		{"pickup negative", "4/4;pickup:-1/4", measure.ErrInvalidText, nil},
		{"pickup twice", "4/4;pickup:1/4;pickup:1/4", measure.ErrInvalidText, nil},
		{"pickup after a mark", "4/4;3:3/4;pickup:1/4", measure.ErrInvalidText, nil},
		{"bars before pickup", "4/4;bars:4;pickup:1/4", measure.ErrInvalidText, nil},
		{"bars twice", "4/4;bars:4;bars:4", measure.ErrInvalidText, nil},
		{"bars zero", "4/4;bars:0", measure.ErrInvalidText, nil},
		{"bars leading zero", "4/4;bars:04", measure.ErrInvalidText, nil},
		{"bars sign", "4/4;bars:+4", measure.ErrInvalidText, nil},
		{"bars empty", "4/4;bars:", measure.ErrInvalidText, nil},
		{"bars after", "4/4;bars:4;3:3/4", measure.ErrInvalidText, nil},
		{"mark at bar 0", "4/4;0:3/4", measure.ErrInvalidText, nil},
		{"mark at bar 1", "4/4;1:3/4", measure.ErrInvalidText, nil},
		{"mark leading zero", "4/4;03:3/4", measure.ErrInvalidText, nil},
		{"mark without bar", "4/4;:3/4", measure.ErrInvalidText, nil},
		{"mark without meter", "4/4;3:", measure.ErrInvalidText, nil},
		{"mark with suffix", "4/4;3:3/4:end", measure.ErrInvalidText, nil},
		{"duplicate mark", "4/4;3:3/4;3:2/4", measure.ErrInvalidText, nil},
		{"decreasing marks", "4/4;5:3/4;3:2/4", measure.ErrInvalidText, nil},
		{"length before mark", "4/4;3:length:1/2;3:3/4", measure.ErrInvalidText, nil},
		{"duplicate length", "4/4;3:length:1/2;3:length:1/4", measure.ErrInvalidText, nil},
		{"decreasing lengths", "4/4;4:length:1/2;3:length:1/4", measure.ErrInvalidText, nil},
		{"length at bar 0", "4/4;0:length:1/2", measure.ErrInvalidText, nil},
		{"length without value", "4/4;3:length:", measure.ErrInvalidText, nil},
		{"length start", "4/4;3:length:1/2:start", measure.ErrInvalidText, nil},
		{"length end twice", "4/4;3:length:1/2:end:end", measure.ErrInvalidText, nil},
		{"length without bar", "4/4;length:1/2", measure.ErrInvalidText, nil},
		{"upper case keyword", "4/4;3:LENGTH:1/2", measure.ErrInvalidText, nil},
		{"upper case pickup", "4/4;PICKUP:1/4", measure.ErrInvalidText, nil},
		{"length integer form", "4/4;3:length:1", measure.ErrInvalidText, nil},
		{"form before a big number", "4/4;" + big9 + ":3/4;bars", measure.ErrInvalidText, nil},
		{"order before a nested range", "4/4;5:3/" + big9 + ";3:2/4", measure.ErrInvalidText, nil},
		{"nested left to right: text first", "4/4;pickup:x;3:3/" + big9, measure.ErrInvalidText, nil},
		{"nested before building", "4/4;pickup:5/4;3:x", measure.ErrInvalidText, nil},
		{"mark number > MaxInt", "4/4;" + big9 + ":3/4", measure.ErrOutOfRange, nil},
		{"MaxInt + 1", "4/4;9223372036854775808:3/4", measure.ErrOutOfRange, nil},
		{"bars > MaxInt", "4/4;bars:9223372036854775808", measure.ErrOutOfRange, nil},
		{"length number > MaxInt", "4/4;" + big9 + ":length:1/2", measure.ErrOutOfRange, nil},
		{"number before order", "4/4;5:3/4;" + big9 + ":2/4;3:2/4", measure.ErrOutOfRange, nil},
		{"pickup out of range", "4/4;pickup:1/" + huge, measure.ErrOutOfRange, rangeCauses(durationErr)},
		{"length out of range", "4/4;3:length:1/" + huge, measure.ErrOutOfRange, rangeCauses(durationErr)},
		{"meter out of range", "3/" + big9, measure.ErrOutOfRange, rangeCauses(meterRangeErr)},
		{"mark out of range", "4/4;3:3/" + big9, measure.ErrOutOfRange, rangeCauses(meterRangeErr)},
		{"meter sum out of range", meterOutOfRange(), measure.ErrOutOfRange, rangeCauses(meterSumErr)},
		{"nested left to right: range first", "4/4;pickup:1/" + huge + ";3:x", measure.ErrOutOfRange, rangeCauses(durationErr)},
		{"КП-40 pickup longer", "4/4;pickup:5/4", measure.ErrLongerThanMeter, nil},
		{"pickup by the meter of bar 1", "2/4;pickup:3/4;3:4/4", measure.ErrLongerThanMeter, nil},
		{"length to the end longer", "4/4;5:length:5/4:end", measure.ErrLongerThanMeter, nil},
		{"length to the end by its meter", "4/4;3:3/4;3:length:7/8:end", measure.ErrLongerThanMeter, nil},
		{"pickup before count", "4/4;pickup:5/4;3:3/4;bars:2", measure.ErrLongerThanMeter, nil},
		{"count below a mark", "4/4;3:3/4;bars:2", measure.ErrOutsideGrid, nil},
		{"count below a length", "4/4;3:length:1/2;bars:2", measure.ErrOutsideGrid, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g, err := measure.Parse(tt.text)
			assertErrExact(t, err, tt.err, tt.causes...)
			assert.False(t, g.IsSet())
			assertParseMessage(t, err, tt.text)

			before := gridOf(t, "3/4;pickup:1/4")
			u := before
			uerr := u.UnmarshalText([]byte(tt.text))
			assertErrExact(t, uerr, tt.err, tt.causes...)
			assert.True(t, u.Equal(before), "UnmarshalText does not change the receiver on error")

			assertParseMessage(t, uerr, tt.text)
		})
	}

	t.Run("range causes exist", func(t *testing.T) {
		t.Parallel()

		assert.Contains(t, rangeCauses(durationErr), fraction.ErrOutOfRange)
		assert.Contains(t, rangeCauses(meterRangeErr), meter.ErrOutOfRange)
		assert.Contains(t, rangeCauses(meterSumErr), fraction.ErrOutOfRange)
	})
}

// TestTextParseLimits covers architecture 4.16: math.MaxInt is a valid
// number of a bar and of the count; a mark of bar 1 by the first meter and
// an explicit length of bar 1 are valid.
func TestTextParseLimits(t *testing.T) {
	t.Parallel()

	tests := []struct {
		text  string
		check func(t *testing.T, g measure.Grid)
	}{
		{"4/4;bars:9223372036854775807", func(t *testing.T, g measure.Grid) {
			n, finite, err := g.BarCount()
			require.NoError(t, err)
			assert.True(t, finite)
			assert.Equal(t, 9223372036854775807, n)
		}},
		{"4/4;9223372036854775807:3/4", func(t *testing.T, g measure.Grid) {
			has, err := g.HasMark(9223372036854775807)
			require.NoError(t, err)
			assert.True(t, has)
		}},
		{"4/4;1:length:1/2:end", func(t *testing.T, g measure.Grid) {
			assertRecord(t, "4/4;1:length:1/2:end", g)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g, err := measure.Parse(tt.text)
			require.NoError(t, err)
			tt.check(t, g)
		})
	}
}

// TestTextAddress covers О-21, О-22, and КП-38: the text of an address and
// parsing it back.
func TestTextAddress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		address measure.Address
		text    string
	}{
		{ad(2, "1/8"), "2:1/8"},
		{ad(0, "0"), "0:0/1"},
		{ad(1, "5/4"), "1:5/4"},
		{ad(9223372036854775807, "3/7"), "9223372036854775807:3/7"},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			assert.Equal(t, tt.text, tt.address.String())

			b, err := tt.address.MarshalText()
			require.NoError(t, err)
			assert.Equal(t, tt.text, string(b))

			out, err := tt.address.AppendText([]byte("at "))
			require.NoError(t, err)
			assert.Equal(t, "at "+tt.text, string(out))

			a, err := measure.ParseAddress(tt.text)
			require.NoError(t, err)
			assert.Equal(t, tt.address, a)

			var u measure.Address
			require.NoError(t, u.UnmarshalText(b))
			assert.Equal(t, tt.address, u)
		})
	}

	t.Run("noncanonical position", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		a, err := measure.ParseAddress("2:2/16")
		require.NoError(t, err)
		assert.Equal(t, ad(2, "1/8"), a)
		assert.Equal(t, "2:1/8", a.String())
	})
}

// TestTextAddressErrors covers КП-46 and the order of failures of
// architecture 4.16 for addresses: Е-9 for the form, Е-10 for a number
// greater than math.MaxInt, then the position: Е-10 with
// duration.ErrOutOfRange and fraction.ErrOutOfRange, otherwise Е-9.
func TestTextAddressErrors(t *testing.T) {
	t.Parallel()

	huge := hugeDigits()

	tests := []struct {
		name   string
		text   string
		err    error
		causes []error
	}{
		{"empty", "", measure.ErrInvalidText, nil},
		{"КП-46 unset", "Address(unset)", measure.ErrInvalidText, nil},
		{"bar only", "2", measure.ErrInvalidText, nil},
		{"no position", "2:", measure.ErrInvalidText, nil},
		{"no bar", ":1/8", measure.ErrInvalidText, nil},
		{"leading zero", "02:1/8", measure.ErrInvalidText, nil},
		{"negative bar", "-1:0/1", measure.ErrInvalidText, nil},
		{"sign", "+2:0/1", measure.ErrInvalidText, nil},
		{"two colons", "2:1/8:3", measure.ErrInvalidText, nil},
		{"letters", "a:0/1", measure.ErrInvalidText, nil},
		{"space", " 2:1/8", measure.ErrInvalidText, nil},
		{"negative position", "2:-1/8", measure.ErrInvalidText, nil},
		{"integer position", "2:0", measure.ErrInvalidText, nil},
		{"zero denominator", "2:1/0", measure.ErrInvalidText, nil},
		{"form before number", "99999999999999999999:", measure.ErrInvalidText, nil},
		{"bar > MaxInt", "99999999999999999999:0/1", measure.ErrOutOfRange, nil},
		{"MaxInt + 1", "9223372036854775808:0/1", measure.ErrOutOfRange, nil},
		{"number before position", "9223372036854775808:x", measure.ErrOutOfRange, nil},
		{"position out of range", "2:1/" + huge, measure.ErrOutOfRange, []error{duration.ErrOutOfRange, fraction.ErrOutOfRange}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			a, err := measure.ParseAddress(tt.text)
			assertErrExact(t, err, tt.err, tt.causes...)
			assert.Equal(t, measure.Address{}, a)
			assertParseMessage(t, err, tt.text)

			u := ad(3, "1/4")
			uerr := u.UnmarshalText([]byte(tt.text))
			assertErrExact(t, uerr, tt.err, tt.causes...)
			assert.Equal(t, ad(3, "1/4"), u, "UnmarshalText does not change the receiver on error")
		})
	}
}

// TestTextErrorMessage covers the text of parse errors of architecture
// 4.16: the excerpt of long inputs, the limit of 512 bytes, no length of the
// input, and a dependence only on the first 67 bytes within one class.
func TestTextErrorMessage(t *testing.T) {
	t.Parallel()

	t.Run("long input", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		x := "4/4;" + strings.Repeat("x", 119)
		_, err := measure.Parse(x)
		require.Error(t, err)

		msg := err.Error()
		assert.Contains(t, msg, strconv.Quote(x[:64])+"...")
		assert.NotContains(t, msg, x[:65])
		assert.NotContains(t, msg, "123", "the length of the input is not written")
		assert.LessOrEqual(t, len(msg), 512)
	})

	t.Run("UTF-8 prefix", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		x := strings.Repeat("a", 63) + "é" + strings.Repeat("b", 10)
		_, err := measure.Parse(x)
		require.Error(t, err)
		assert.Contains(t, err.Error(), strconv.Quote(x[:63])+"...")

		_, err = measure.ParseAddress(x)
		require.Error(t, err)
		assert.Contains(t, err.Error(), strconv.Quote(x[:63])+"...")
	})

	t.Run("exactly 64 bytes", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		x := strings.Repeat("z", 64)
		_, err := measure.Parse(x)
		require.Error(t, err)
		assert.Contains(t, err.Error(), strconv.Quote(x))
		assert.NotContains(t, err.Error(), strconv.Quote(x)+"...")
	})

	// longPickup is 62 bytes: a pickup longer than the meter and marks.
	longPickup := "4/4;pickup:5/4;2:3/4;3:2/4;4:3/4;5:2/4;6:3/4;7:2/4;8:3/4;9:2/4"

	pairs := []struct {
		name string
		a, b string
		err  error
	}{
		{"invalid text", "4/4;" + strings.Repeat("x", 100), "4/4;" + strings.Repeat("x", 63) + strings.Repeat("y;", 40), measure.ErrInvalidText},
		{"number > MaxInt", "4/4;bars:" + strings.Repeat("9", 80), "4/4;bars:" + strings.Repeat("9", 58) + strings.Repeat("8", 30), measure.ErrOutOfRange},
		{"nested range", "4/4;pickup:1/" + strings.Repeat("9", 90), "4/4;pickup:1/" + strings.Repeat("9", 54) + strings.Repeat("7", 60), measure.ErrOutOfRange},
		{"longer than the meter", longPickup + ";10:3/4;11:2/4;12:3/4", longPickup + ";10:3/8;20:5/4", measure.ErrLongerThanMeter},
		{"address", strings.Repeat("7", 70) + ":0/1", strings.Repeat("7", 67) + "9:0/1", measure.ErrOutOfRange},
	}

	for _, tt := range pairs {
		t.Run("first 67 bytes: "+tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			require.Equal(t, tt.a[:67], tt.b[:67])

			parse := func(s string) error {
				if tt.name == "address" {
					_, err := measure.ParseAddress(s)

					return err
				}

				_, err := measure.Parse(s)

				return err
			}

			errA, errB := parse(tt.a), parse(tt.b)
			assertErr(t, errA, tt.err)
			assertErr(t, errB, tt.err)

			for _, s := range dependencySentinels {
				require.Equal(t, errors.Is(errA, s.err), errors.Is(errB, s.err), "the same class: %s", s.name)
			}

			assert.Equal(t, errA.Error(), errB.Error())
			assert.LessOrEqual(t, len(errA.Error()), 512)
		})
	}
}

// TestTextUnmarshalNil covers architecture 4.16: UnmarshalText of a nil
// receiver fails with Е-2.
func TestTextUnmarshalNil(t *testing.T) {
	t.Parallel()
	defer guard(t)

	var g *measure.Grid
	assertErr(t, g.UnmarshalText([]byte("4/4")), measure.ErrInvalidArgument)

	var a *measure.Address
	assertErr(t, a.UnmarshalText([]byte("2:1/8")), measure.ErrInvalidArgument)
}

// TestTextUnmarshalReplace covers architecture 4.16: UnmarshalText replaces
// a set value and does not keep the input.
func TestTextUnmarshalReplace(t *testing.T) {
	t.Parallel()
	defer guard(t)

	g := gridOf(t, "3/4;pickup:1/4")
	input := []byte("4/4;3:3/4;bars:4")
	require.NoError(t, g.UnmarshalText(input))

	copy(input, "XXXXXXXXXXXXXXXX")
	assert.Equal(t, "4/4;3:3/4;bars:4", g.String())

	a := ad(1, "0")
	addr := []byte("2:1/8")
	require.NoError(t, a.UnmarshalText(addr))

	copy(addr, "XXXXX")
	assert.Equal(t, ad(2, "1/8"), a)
}

// TestTextJSON covers architecture 4.16: a grid and an address as fields of
// a struct in JSON; an unset field fails to encode unless it is omitted with
// omitzero.
func TestTextJSON(t *testing.T) {
	t.Parallel()
	defer guard(t)

	type doc struct {
		Grid measure.Grid    `json:"grid"`
		At   measure.Address `json:"at"`
	}

	in := doc{Grid: gridOf(t, "3/4;pickup:1/4;3:2/4"), At: ad(2, "1/8")}
	b, err := json.Marshal(in)
	require.NoError(t, err)
	assert.JSONEq(t, `{"grid":"3/4;pickup:1/4;3:2/4","at":"2:1/8"}`, string(b))

	var out doc
	require.NoError(t, json.Unmarshal(b, &out))
	assert.True(t, out.Grid.Equal(in.Grid))
	assert.Equal(t, in.At, out.At)

	_, err = json.Marshal(doc{At: ad(1, "0")})
	assert.Error(t, err, "an unset grid does not encode")

	_, err = json.Marshal(doc{Grid: in.Grid})
	assert.Error(t, err, "an unset address does not encode")

	type optional struct {
		Grid measure.Grid    `json:"grid,omitzero"`
		At   measure.Address `json:"at,omitzero"`
	}

	b, err = json.Marshal(optional{})
	require.NoError(t, err)
	assert.JSONEq(t, `{}`, string(b))

	err = json.Unmarshal([]byte(`{"grid":"4/4;","at":"2:1/8"}`), &out)
	require.Error(t, err)
	assert.ErrorIs(t, err, measure.ErrInvalidText)
}

// TestTextEnums covers О-22 and architecture 4.3: the constants, their
// texts, IsValid, and Equal of Alignment, Kind, and Level.
func TestTextEnums(t *testing.T) {
	t.Parallel()

	t.Run("Alignment", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, measure.Alignment(1), measure.AlignStart)
		assert.Equal(t, measure.Alignment(2), measure.AlignEnd)

		for _, tt := range []struct {
			v     measure.Alignment
			text  string
			valid bool
		}{
			{measure.AlignStart, "start", true},
			{measure.AlignEnd, "end", true},
			{0, "Alignment(invalid)", false},
			{3, "Alignment(invalid)", false},
			{255, "Alignment(invalid)", false},
		} {
			assert.Equal(t, tt.text, tt.v.String())
			assert.Equal(t, tt.valid, tt.v.IsValid())
			assert.True(t, tt.v.Equal(tt.v))
			assert.False(t, tt.v.Equal(tt.v+1))
		}
	})

	t.Run("Kind", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, measure.Kind(1), measure.KindFull)
		assert.Equal(t, measure.Kind(2), measure.KindIncomplete)
		assert.Equal(t, measure.Kind(3), measure.KindExtended)

		for _, tt := range []struct {
			v     measure.Kind
			text  string
			valid bool
		}{
			{measure.KindFull, "full", true},
			{measure.KindIncomplete, "incomplete", true},
			{measure.KindExtended, "extended", true},
			{0, "Kind(invalid)", false},
			{4, "Kind(invalid)", false},
		} {
			assert.Equal(t, tt.text, tt.v.String())
			assert.Equal(t, tt.valid, tt.v.IsValid())
			assert.True(t, tt.v.Equal(tt.v))
			assert.False(t, tt.v.Equal(tt.v+1))
		}
	})

	t.Run("Level", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, measure.Level(1), measure.LevelBars)
		assert.Equal(t, measure.Level(2), measure.LevelGroups)
		assert.Equal(t, measure.Level(3), measure.LevelBeats)
		assert.Equal(t, measure.Level(4), measure.LevelUnits)

		for _, tt := range []struct {
			v     measure.Level
			text  string
			valid bool
		}{
			{measure.LevelBars, "bars", true},
			{measure.LevelGroups, "groups", true},
			{measure.LevelBeats, "beats", true},
			{measure.LevelUnits, "units", true},
			{0, "Level(invalid)", false},
			{5, "Level(invalid)", false},
		} {
			assert.Equal(t, tt.text, tt.v.String())
			assert.Equal(t, tt.valid, tt.v.IsValid())
			assert.True(t, tt.v.Equal(tt.v))
			assert.False(t, tt.v.Equal(tt.v+1))
		}
	})

	t.Run("limits", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, 1024, measure.MaxParts)
		assert.Equal(t, 1024, measure.MaxAlternation)
	})
}

// TestTextRecords covers architecture 4.13: the texts of Mark,
// ExplicitLength, MeterRun, and Part, and their Equal, the same as ==.
func TestTextRecords(t *testing.T) {
	t.Parallel()

	t.Run("String", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "3:3/4", measure.Mark{Bar: 3, Meter: mt("3/4")}.String())
		assert.Equal(t, "9:C|", measure.Mark{Bar: 9, Meter: mt("C|")}.String())
		assert.Equal(t, "0:Meter(unset)", measure.Mark{}.String())
		assert.Equal(t, "4:length:1/2", measure.ExplicitLength{Bar: 4, Duration: dur("1/2"), Alignment: measure.AlignStart}.String())
		assert.Equal(t, "5:length:1/4:end", measure.ExplicitLength{Bar: 5, Duration: dur("1/4"), Alignment: measure.AlignEnd}.String())
		assert.Equal(t, "5:length:1/4:Alignment(invalid)", measure.ExplicitLength{Bar: 5, Duration: dur("1/4")}.String())
		assert.Equal(t, "0:length:Duration(unset):Alignment(invalid)", measure.ExplicitLength{}.String())
		assert.Equal(t, "3-5:3/4", measure.MeterRun{First: 3, Last: 5, Meter: mt("3/4")}.String())
		assert.Equal(t, "0-0:Meter(unset)", measure.MeterRun{}.String())
		assert.Equal(t, "1/4@2:0/1", measure.Part{Address: ad(2, "0"), Start: pos("5/4"), Duration: dur("1/4")}.String())
		assert.Equal(t, "Duration(unset)@Address(unset)", measure.Part{}.String())
	})

	t.Run("Equal", func(t *testing.T) {
		t.Parallel()

		marks := []measure.Mark{{}, {Bar: 3, Meter: mt("3/4")}, {Bar: 3, Meter: mt("6/8")}, {Bar: 4, Meter: mt("3/4")}}
		for _, a := range marks {
			for _, b := range marks {
				assert.Equal(t, a == b, a.Equal(b), "%s and %s", a, b)
			}
		}

		lengths := []measure.ExplicitLength{
			{}, {Bar: 4, Duration: dur("1/2"), Alignment: measure.AlignStart},
			{Bar: 4, Duration: dur("1/2"), Alignment: measure.AlignEnd}, {Bar: 4, Duration: dur("1/4"), Alignment: measure.AlignStart},
		}
		for _, a := range lengths {
			for _, b := range lengths {
				assert.Equal(t, a == b, a.Equal(b), "%s and %s", a, b)
			}
		}

		runs := []measure.MeterRun{{}, {First: 1, Last: 2, Meter: mt("4/4")}, {First: 1, Last: 3, Meter: mt("4/4")}, {First: 1, Last: 2, Meter: mt("C")}}
		for _, a := range runs {
			for _, b := range runs {
				assert.Equal(t, a == b, a.Equal(b), "%s and %s", a, b)
			}
		}

		parts := []measure.Part{
			{}, {Address: ad(2, "0"), Start: pos("1"), Duration: dur("1/4")},
			{Address: ad(2, "0"), Start: pos("1"), Duration: dur("1/2")}, {Address: ad(1, "0"), Start: pos("1"), Duration: dur("1/4")},
		}
		for _, a := range parts {
			for _, b := range parts {
				assert.Equal(t, a == b, a.Equal(b), "%s and %s", a, b)
			}
		}

		bars := []measure.Bar{
			{},
			barOf{1, "4/4", "1", "0", "1", measure.KindFull, false, false, measure.AlignStart, "0", true, false}.value(),
			barOf{1, "4/4", "1", "0", "1", measure.KindFull, false, false, measure.AlignStart, "0", true, true}.value(),
			barOf{1, "C", "1", "0", "1", measure.KindFull, false, false, measure.AlignStart, "0", true, false}.value(),
		}
		for _, a := range bars {
			for _, b := range bars {
				assert.Equal(t, a == b, a.Equal(b))
			}
		}
	})

	t.Run("Part.End", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		end, err := measure.Part{Address: ad(2, "1/4"), Start: pos("5/4"), Duration: dur("1/8")}.End()
		require.NoError(t, err)
		assertText(t, "11/8", end)

		_, err = measure.Part{Duration: dur("1/8")}.End()
		assertErr(t, err, measure.ErrInvalidArgument, duration.ErrUnset)

		_, err = measure.Part{Start: pos("1")}.End()
		assertErr(t, err, measure.ErrInvalidArgument, duration.ErrUnset)

		e, err := measure.Part{}.End()
		assertErr(t, err, measure.ErrInvalidArgument, duration.ErrUnset)
		assert.False(t, e.IsSet())
	})
}
