package degree_test

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/degree"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDegreeNumberFormattingControlExamples(t *testing.T) {
	t.Parallel()
	tests := []struct {
		number  int
		decimal string
		roman   string
	}{
		{1, "1", "I"},
		{4, "4", "IV"},
		{7, "7", "VII"},
		{8, "8", "VIII"},
		{9, "9", "IX"},
		{12, "12", "XII"},
		{14, "14", "XIV"},
		{40, "40", "XL"},
		{53, "53", "LIII"},
		{90, "90", "XC"},
		{400, "400", "CD"},
		{900, "900", "CM"},
		{1994, "1994", "MCMXCIV"},
		{3999, "3999", "MMMCMXCIX"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("number_%d", tt.number), func(t *testing.T) {
			t.Parallel()
			d := degree.MustNew(tt.number)
			decimal, err := d.Decimal()
			require.NoError(t, err)
			assert.Equal(t, tt.decimal, decimal)
			roman, err := d.Roman()
			require.NoError(t, err)
			assert.Equal(t, tt.roman, roman)
			assert.Equal(t, decimal, d.MustDecimal())
			assert.Equal(t, roman, d.MustRoman())

			withNote := degree.MustNewWithNote(tt.number, degreeTestNote(t, spelling.B(), cents.Zero(), nil))
			assert.Equal(t, decimal, withNote.MustDecimal())
			assert.Equal(t, roman, withNote.MustRoman())
		})
	}
}

func TestDegreeFormattingAtIntegerAndRomanLimits(t *testing.T) {
	t.Parallel()
	d := degree.MustNew(math.MaxInt)
	assert.Equal(t, strconv.Itoa(math.MaxInt), d.MustDecimal())
	roman, err := d.Roman()
	assertDegreeError(t, err, degree.ErrRomanOutOfRange)
	assert.Empty(t, roman)
	_ = assertMustPanic(t, degree.ErrRomanOutOfRange, func() { d.MustRoman() })

	fourThousand := degree.MustNew(4000)
	assert.Equal(t, "4000", fourThousand.MustDecimal())
	roman, err = fourThousand.Roman()
	assertDegreeError(t, err, degree.ErrRomanOutOfRange)
	assert.Empty(t, roman)
}

func TestDegreeFormattingUnset(t *testing.T) {
	t.Parallel()
	var d degree.Degree
	decimal, err := d.Decimal()
	assertDegreeError(t, err, degree.ErrUnset)
	assert.Empty(t, decimal)
	_ = assertMustPanic(t, degree.ErrUnset, func() { d.MustDecimal() })
	roman, err := d.Roman()
	assertDegreeError(t, err, degree.ErrUnset)
	assert.Empty(t, roman)
	_ = assertMustPanic(t, degree.ErrUnset, func() { d.MustRoman() })
}

func TestDegreeParseCanonicalForms(t *testing.T) {
	t.Parallel()
	tests := []struct {
		text   string
		number int
	}{
		{"4", 4},
		{"IV", 4},
		{"VIII", 8},
		{"XII", 12},
		{"53", 53},
		{"LIII", 53},
		{"3999", 3999},
		{"MMMCMXCIX", 3999},
		{"4000", 4000},
		{strconv.Itoa(math.MaxInt), math.MaxInt},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()
			d, err := degree.Parse(tt.text)
			require.NoError(t, err)
			assertDegreeState(t, d, tt.number, false, note.Note{})
			must := degree.MustParse(tt.text)
			assertSameDegree(t, d, must)
		})
	}
}

func TestDegreeParseRejectsNonCanonicalText(t *testing.T) {
	t.Parallel()
	oversized := strings.Repeat("9", 40)
	for _, text := range []string{
		"", "iv", "Iv", "IIII", "IL", "VX", "MMMM", "4 ", " 4", "+4", "-4", "04", "000000000000000000000000000000000000000", "4.0", "IV4", "Ⅳ", "４", "I I", "IIV", "VV", "IC", "XD", "CMCM",
		oversized + "I", oversized + " ", oversized + "Ⅳ",
	} {
		t.Run(fmt.Sprintf("%q", text), func(t *testing.T) {
			t.Parallel()
			d, err := degree.Parse(text)
			assertDegreeError(t, err, degree.ErrParse)
			assertUnsetDegree(t, d)
			_ = assertMustPanic(t, degree.ErrParse, func() { degree.MustParse(text) })
		})
	}
}

func TestDegreeParseNumericErrorCategories(t *testing.T) {
	t.Parallel()
	d, err := degree.Parse("0")
	assertDegreeError(t, err, degree.ErrInvalidNumber)
	assertUnsetDegree(t, d)
	_ = assertMustPanic(t, degree.ErrInvalidNumber, func() { degree.MustParse("0") })

	oversized := strings.Repeat("9", 40)
	for _, text := range []string{"99999999999999999999", "9223372036854775808", oversized} {
		t.Run(fmt.Sprintf("%q", text), func(t *testing.T) {
			t.Parallel()
			d, err := degree.Parse(text)
			assertDegreeError(t, err, degree.ErrOutOfRange)
			assertUnsetDegree(t, d)
			_ = assertMustPanic(t, degree.ErrOutOfRange, func() { degree.MustParse(text) })
		})
	}
}

func TestDegreeFormattingRoundTripEveryRomanNumber(t *testing.T) {
	t.Parallel()
	for number := 1; number <= degree.MaxRomanNumber; number++ {
		d := degree.MustNew(number)
		decimal, err := d.Decimal()
		require.NoError(t, err, "Decimal(%d)", number)
		fromDecimal, err := degree.Parse(decimal)
		require.NoError(t, err, "Parse(%q)", decimal)
		assertDegreeState(t, fromDecimal, number, false, note.Note{})

		roman, err := d.Roman()
		require.NoError(t, err, "Roman(%d)", number)
		fromRoman, err := degree.Parse(roman)
		require.NoError(t, err, "Parse(%q)", roman)
		assertDegreeState(t, fromRoman, number, false, note.Note{})
	}
}
