package spelling_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// formNames are the names whose machine form is checked: the set N, the 35
// ready-made names, and the boundary names of K-10 item 4.
func formNames() []nameCase {
	var ready []nameCase
	for _, p := range predefinedCases() {
		ready = append(ready, nc(p.l, p.a))
	}

	return concatNames(setN(), ready, boundaryNames())
}

// K-28: the exact text forms named by the architecture.
func TestStringForms(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		s    spelling.Spelling
		want string
	}{
		{"C", spelling.C(), "C"},
		{"B", spelling.B(), "B"},
		{"C sharp", spelling.CSharp(), "C+1"},
		{"D flat", spelling.DFlat(), "D-1"},
		{"C double sharp", spelling.CDoubleSharp(), "C+2"},
		{"B double flat", spelling.BDoubleFlat(), "B-2"},
		{"E+3", spelling.MustNew(letter.E, whole(3)), "E+3"},
		{"C+1/2", spelling.MustNew(letter.C, frac(1, 2)), "C+1/2"},
		{"D-3/2", spelling.MustNew(letter.D, frac(-3, 2)), "D-3/2"},
		{"B-7/2", spelling.MustNew(letter.B, frac(-7, 2)), "B-7/2"},
		{"(C, M)", spelling.MustNew(letter.C, whole(maxMagnitude)), "C+16777216"},
		{"(C, -M)", spelling.MustNew(letter.C, whole(-maxMagnitude)), "C-16777216"},
		{"(C, 1/Q)", spelling.MustNew(letter.C, frac(1, maxDenominator)), "C+1/65536"},
		{"(C, -1/Q)", spelling.MustNew(letter.C, frac(-1, maxDenominator)), "C-1/65536"},
	}

	suffix := map[string]string{"DoubleFlat": "-2", "Flat": "-1", "": "", "Sharp": "+1", "DoubleSharp": "+2"}
	for _, p := range predefinedCases() {
		letterName, sign := p.name[:1], p.name[1:]
		want, ok := suffix[sign]
		require.True(t, ok, p.name)
		cases = append(cases, struct {
			name string
			s    spelling.Spelling
			want string
		}{p.name + "()", p.fn(), letterName + want})
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, tc.s.String(), "String")
			assert.Equal(t, tc.want, fmt.Sprint(tc.s), "fmt.Sprint")

			text, err := tc.s.MarshalText()
			require.NoError(t, err)
			assert.Equal(t, tc.want, string(text), "MarshalText")

			appended, err := tc.s.AppendText(nil)
			require.NoError(t, err)
			assert.Equal(t, tc.want, string(appended), "AppendText")

			data, err := tc.s.MarshalJSON()
			require.NoError(t, err)
			assert.Equal(t, `"`+tc.want+`"`, string(data), "MarshalJSON")
		})
	}
}

// K-28: the form follows the grammar, is read back as the same name, and
// differs for different names.
func TestFormRoundTrip(t *testing.T) {
	t.Parallel()

	names := formNames()

	for _, tc := range names {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			x := tc.build()
			form := x.String()

			assert.Equal(t, formOf(tc.l, tc.a), form, "the form by the grammar")
			assert.Equal(t, form, x.String(), "writing again gives the same form")

			text, err := x.MarshalText()
			require.NoError(t, err)
			assert.Equal(t, form, string(text))

			appended, err := x.AppendText(nil)
			require.NoError(t, err)
			assert.Equal(t, form, string(appended))

			parsed, err := spelling.Parse(form)
			require.NoError(t, err)
			assertName(t, tc.l, tc.a, parsed)
			assert.True(t, parsed.Equal(x))

			var read spelling.Spelling

			require.NoError(t, read.UnmarshalText([]byte(form)))
			assertName(t, tc.l, tc.a, read)

			read = spelling.DFlat()
			require.NoError(t, read.UnmarshalJSON([]byte(`"`+form+`"`)))
			assertName(t, tc.l, tc.a, read)
		})
	}

	t.Run("different names have different forms", func(t *testing.T) {
		t.Parallel()

		seen := map[string]nameCase{}

		for _, tc := range names {
			x := tc.build()
			form := x.String()

			if prev, ok := seen[form]; ok {
				assert.Truef(t, prev.build().Equal(x), "the form %q is shared by %s and %s", form, prev.name, tc.name)
			}

			seen[form] = tc
			assert.NotEqual(t, "unset", form)
		}
	})
}

// invalidForms are the strings that are not the machine form of any name,
// with the secondary alteration error each reading error must carry
// (the architecture, "Reading errors").
func invalidForms() []struct {
	text  string
	class formErrorClass
} {
	return []struct {
		text  string
		class formErrorClass
	}{
		{"", formOnly},
		{" C", formOnly},
		{"C ", formOnly},
		{"X+1", formOnly},
		{"c+1", formOnly},
		{"H", formOnly},
		{"C+", formOnly},
		{"+1", formOnly},
		{"C1", formOnly},
		{"C+1/", formOnly},
		{"C+1x", formOnly},
		{"C+1/2/3", formOnly},
		{"C+0", formOnly},
		{"C-0", formOnly},
		{"C0", formOnly},
		{"C+01", formOnly},
		{"C+1/00", formOnly},
		{"C+1/02", formOnly},
		{"C++1", formOnly},
		{"C+-1", formOnly},
		{"C+1/+2", formOnly},
		{"C+1/-2", formOnly},
		{"C+٣", formOnly},
		{"C#", formOnly},
		{"C♯", formOnly},
		{"Cis", formOnly},
		{"до-диез", formOnly},
		{"Unset", formOnly},
		{"unset ", formOnly},
		{" unset", formOnly},
		{"C+1/1", formOnly},
		{"C+2/4", formOnly},
		{"C-4/2", formOnly},
		{"C+1/0", formZeroDenominator},
		{"C-7/0", formZeroDenominator},
		// A zero denominator wins over a numerator that does not fit in int.
		{"C+99999999999999999999/0", formZeroDenominator},
		{"C+16777217", formOutOfRange},
		{"C-16777217", formOutOfRange},
		{"C+1/65537", formOutOfRange},
		{"C+33554434/2", formOutOfRange},
		{"C+16777217/1", formOutOfRange},
		{"C+99999999999999999999", formOutOfRange},
		{"C+1/99999999999999999999", formOutOfRange},
	}
}

// K-29: reading a string that is not a machine form is an error; the result
// is the unset name and the receiver keeps its value.
func TestParseInvalid(t *testing.T) {
	t.Parallel()

	for _, tc := range invalidForms() {
		t.Run(fmt.Sprintf("%q", tc.text), func(t *testing.T) {
			t.Parallel()

			got, err := spelling.Parse(tc.text)
			assertInvalidForm(t, tc.class, err)
			assertUnset(t, got)

			assertPanicsWithErrorIs(t, spelling.ErrInvalidForm, func() { _ = spelling.MustParse(tc.text) })

			receiver := spelling.DFlat()
			err = receiver.UnmarshalText([]byte(tc.text))
			assertInvalidForm(t, tc.class, err)
			assert.Equal(t, spelling.DFlat(), receiver, "UnmarshalText must keep the receiver")

			quoted, err := json.Marshal(tc.text)
			require.NoError(t, err)

			receiver = spelling.DFlat()
			err = receiver.UnmarshalJSON(quoted)
			assertInvalidForm(t, tc.class, err)
			assert.Equal(t, spelling.DFlat(), receiver, "UnmarshalJSON must keep the receiver")
		})
	}
}

// K-29: empty input of the text methods.
func TestUnmarshalTextEmpty(t *testing.T) {
	t.Parallel()

	for name, text := range map[string][]byte{"nil": nil, "empty": {}} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			receiver := spelling.DFlat()
			err := receiver.UnmarshalText(text)
			assertInvalidForm(t, formOnly, err)
			assert.Equal(t, spelling.DFlat(), receiver)
		})
	}
}

// K-30: the mark of the unset name is read as the unset name, not C.
func TestParseUnset(t *testing.T) {
	t.Parallel()

	got, err := spelling.Parse("unset")
	require.NoError(t, err)
	assertUnset(t, got)
	assert.False(t, got.Equal(spelling.C()))
	assert.False(t, got.IsNatural())

	require.NotPanics(t, func() { assertUnset(t, spelling.MustParse("unset")) })

	receiver := spelling.DFlat()
	require.NoError(t, receiver.UnmarshalText([]byte("unset")))
	assertUnset(t, receiver)

	receiver = spelling.C()
	require.NoError(t, receiver.UnmarshalText([]byte("unset")))
	assertUnset(t, receiver)
}

// K-36: names read from the machine form are set.
func TestParsedNamesAreSet(t *testing.T) {
	t.Parallel()

	for _, tc := range formNames() {
		got, err := spelling.Parse(tc.name)
		require.NoError(t, err, tc.name)
		assert.Truef(t, got.IsValid(), "the name read from %q", tc.name)
	}
}

// The letter-only check of K-29: no reading error is reported as an invalid
// letter, and no failed reading yields a name with the letter C.
func TestParseNeverYieldsC(t *testing.T) {
	t.Parallel()

	for _, tc := range invalidForms() {
		got, err := spelling.Parse(tc.text)
		require.Error(t, err, "%q", tc.text)
		assert.NotEqualf(t, letter.C, got.Letter(), "%q", tc.text)
		assert.Falsef(t, got.Equal(spelling.C().WithAlteration(alteration.Natural())), "%q", tc.text)
	}
}
