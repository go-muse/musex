package spelling_test

import (
	"bytes"
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// inputOperation calls one operation of the package, including the operations
// that fail and the Must variants that panic, on the inputs x, y, d, and n.
type inputOperation struct {
	name string
	call func(x, y spelling.Spelling, d alteration.Alteration, n int)
}

func inputOperations() []inputOperation {
	return []inputOperation{
		{"Letter", func(x, _ spelling.Spelling, _ alteration.Alteration, _ int) { _ = x.Letter() }},
		{"Alteration", func(x, _ spelling.Spelling, _ alteration.Alteration, _ int) { _ = x.Alteration() }},
		{"Parts", func(x, _ spelling.Spelling, _ alteration.Alteration, _ int) { _, _ = x.Parts() }},
		{"IsValid", func(x, _ spelling.Spelling, _ alteration.Alteration, _ int) { _ = x.IsValid() }},
		{"IsNatural", func(x, _ spelling.Spelling, _ alteration.Alteration, _ int) { _ = x.IsNatural() }},
		{"IsInteger", func(x, _ spelling.Spelling, _ alteration.Alteration, _ int) { _ = x.IsInteger() }},
		{"Direction", func(x, _ spelling.Spelling, _ alteration.Alteration, _ int) { _ = x.Direction() }},
		{"SameLetter", func(x, y spelling.Spelling, _ alteration.Alteration, _ int) { _ = x.SameLetter(y) }},
		{"SameAlteration", func(x, y spelling.Spelling, _ alteration.Alteration, _ int) { _ = x.SameAlteration(y) }},
		{"Natural", func(x, _ spelling.Spelling, _ alteration.Alteration, _ int) { _ = x.Natural() }},
		{"WithAlteration", func(x, _ spelling.Spelling, d alteration.Alteration, _ int) { _ = x.WithAlteration(d) }},
		{"Raise", func(x, _ spelling.Spelling, _ alteration.Alteration, _ int) { _, _ = x.Raise() }},
		{"MustRaise", func(x, _ spelling.Spelling, _ alteration.Alteration, _ int) { _ = x.MustRaise() }},
		{"Lower", func(x, _ spelling.Spelling, _ alteration.Alteration, _ int) { _, _ = x.Lower() }},
		{"MustLower", func(x, _ spelling.Spelling, _ alteration.Alteration, _ int) { _ = x.MustLower() }},
		{"RaiseBy", func(x, _ spelling.Spelling, _ alteration.Alteration, n int) { _, _ = x.RaiseBy(n) }},
		{"MustRaiseBy", func(x, _ spelling.Spelling, _ alteration.Alteration, n int) { _ = x.MustRaiseBy(n) }},
		{"LowerBy", func(x, _ spelling.Spelling, _ alteration.Alteration, n int) { _, _ = x.LowerBy(n) }},
		{"MustLowerBy", func(x, _ spelling.Spelling, _ alteration.Alteration, n int) { _ = x.MustLowerBy(n) }},
		{"RaiseByAlteration", func(x, _ spelling.Spelling, d alteration.Alteration, _ int) { _, _ = x.RaiseByAlteration(d) }},
		{"MustRaiseByAlteration", func(x, _ spelling.Spelling, d alteration.Alteration, _ int) { _ = x.MustRaiseByAlteration(d) }},
		{"LowerByAlteration", func(x, _ spelling.Spelling, d alteration.Alteration, _ int) { _, _ = x.LowerByAlteration(d) }},
		{"MustLowerByAlteration", func(x, _ spelling.Spelling, d alteration.Alteration, _ int) { _ = x.MustLowerByAlteration(d) }},
		{"Equal", func(x, y spelling.Spelling, _ alteration.Alteration, _ int) { _ = x.Equal(y) }},
		{"Compare", func(x, y spelling.Spelling, _ alteration.Alteration, _ int) { _ = x.Compare(y) }},
		{"LetterDistanceUp", func(x, y spelling.Spelling, _ alteration.Alteration, _ int) { _ = x.LetterDistanceUp(y) }},
		{"LetterDistanceDown", func(x, y spelling.Spelling, _ alteration.Alteration, _ int) { _ = x.LetterDistanceDown(y) }},
		{"AlterationDifference", func(x, y spelling.Spelling, _ alteration.Alteration, _ int) { _, _ = x.AlterationDifference(y) }},
		{"MustAlterationDifference", func(x, y spelling.Spelling, _ alteration.Alteration, _ int) { _ = x.MustAlterationDifference(y) }},
		{"String", func(x, _ spelling.Spelling, _ alteration.Alteration, _ int) { _ = x.String() }},
		{"MarshalText", func(x, _ spelling.Spelling, _ alteration.Alteration, _ int) { _, _ = x.MarshalText() }},
		{"AppendText", func(x, _ spelling.Spelling, _ alteration.Alteration, _ int) { _, _ = x.AppendText(nil) }},
		{"MarshalJSON", func(x, _ spelling.Spelling, _ alteration.Alteration, _ int) { _, _ = x.MarshalJSON() }},
		{"New", func(x, _ spelling.Spelling, d alteration.Alteration, _ int) { _, _ = spelling.New(x.Letter(), d) }},
		{"MustNew", func(x, _ spelling.Spelling, d alteration.Alteration, _ int) { _ = spelling.MustNew(x.Letter(), d) }},
	}
}

// K-12: inputs are not changed by any operation, including operations that
// return an error and Must variants that panic.
func TestInputsAreNotChanged(t *testing.T) {
	t.Parallel()

	var unset spelling.Spelling

	inputs := concatNames(setN(), boundaryNames())
	others := []namedSpelling{
		{"C sharp", spelling.CSharp()},
		{"unset", unset},
		{"(C, -M)", spelling.MustNew(letter.C, whole(-maxMagnitude))},
	}
	values := []alteration.Alteration{frac(1, 2), alteration.Sharp(), whole(maxMagnitude)}
	numbers := []int{1, math.MaxInt, math.MinInt}

	for _, op := range inputOperations() {
		t.Run(op.name, func(t *testing.T) {
			t.Parallel()

			check := func(x spelling.Spelling, wantLetter letter.Letter, wantValue alteration.Alteration) {
				t.Helper()

				for _, y := range others {
					for _, d := range values {
						for _, n := range numbers {
							x0, y0, d0, n0 := x, y.s, d, n
							yLetter, yValue := y.s.Letter(), y.s.Alteration()

							catchPanic(func() { op.call(x, y.s, d, n) })

							assert.Equal(t, x0, x, "the receiver")
							assert.Equal(t, wantLetter, x.Letter(), "the letter of the receiver")
							assert.True(t, wantValue.Equal(x.Alteration()), "the value of the receiver")
							assert.Equal(t, y0, y.s, "the argument")
							assert.Equal(t, yLetter, y.s.Letter(), "the letter of the argument")
							assert.True(t, yValue.Equal(y.s.Alteration()), "the value of the argument")
							assert.Equal(t, d0, d, "the alteration argument")
							assert.True(t, d0.Equal(d), "the alteration argument by Equal")
							assert.Equal(t, n0, n, "the int argument")
						}
					}
				}
			}

			for _, in := range inputs {
				check(in.build(), in.l, in.a)
			}

			check(unset, letter.Letter(0), alteration.Natural())
		})
	}

	t.Run("invalid letters given to New", func(t *testing.T) {
		t.Parallel()

		for _, v := range []letter.Letter{letter.Letter(0), letter.Letter(8), letter.Letter(42), letter.Letter(255)} {
			l := v
			a := frac(1, 2)

			_, _ = spelling.New(l, a)
			catchPanic(func() { _ = spelling.MustNew(l, a) })

			assert.True(t, v == l, "the letter must be unchanged, compared as a value")
			assert.Equal(t, frac(1, 2), a)
		}
	})
}

// K-12, finding N2-4: the reading methods do not change their input and do
// not keep it.
func TestReadingDoesNotKeepInput(t *testing.T) {
	t.Parallel()

	t.Run("UnmarshalText", func(t *testing.T) {
		t.Parallel()

		for _, in := range []string{"C+1", "unset", "C#", ""} {
			text := []byte(in)
			before := bytes.Clone(text)

			var s spelling.Spelling

			err := s.UnmarshalText(text)
			assert.True(t, bytes.Equal(before, text), "the input %q must not change", in)

			if err != nil {
				continue
			}

			want := s
			wantForm := s.String()

			for i := range text {
				text[i] = 'x'
			}

			assert.Equal(t, want, s, "the name read from %q must not depend on the input afterwards", in)
			assert.Equal(t, wantForm, s.String())
		}
	})

	t.Run("UnmarshalJSON", func(t *testing.T) {
		t.Parallel()

		for _, in := range []string{`"C+1"`, `"unset"`, `null`, `"C#"`, `1`, ``} {
			data := []byte(in)
			before := bytes.Clone(data)

			var s spelling.Spelling

			err := s.UnmarshalJSON(data)
			assert.True(t, bytes.Equal(before, data), "the input %q must not change", in)

			if err != nil {
				continue
			}

			want := s

			for i := range data {
				data[i] = 'x'
			}

			assert.Equal(t, want, s, "the name read from %q must not depend on the input afterwards", in)
		}
	})

	t.Run("read through encoding/json", func(t *testing.T) {
		t.Parallel()

		data := []byte(`{"s":"C+1"}`)
		before := bytes.Clone(data)

		var v struct {
			S spelling.Spelling `json:"s"`
		}

		require.NoError(t, json.Unmarshal(data, &v))
		assert.True(t, bytes.Equal(before, data))

		for i := range data {
			data[i] = ' '
		}

		assert.True(t, v.S.Equal(spelling.CSharp()))
	})
}

// K-12, finding N2-4: AppendText does not change the bytes already in the
// buffer and appends the form.
func TestAppendTextKeepsPrefix(t *testing.T) {
	t.Parallel()

	var unset spelling.Spelling

	cases := []namedSpelling{
		{"C sharp", spelling.CSharp()},
		{"B-7/2", spelling.MustNew(letter.B, frac(-7, 2))},
		{"unset", unset},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			buf := make([]byte, 0, 64)
			buf = append(buf, "key="...)
			before := bytes.Clone(buf)

			got, err := tc.s.AppendText(buf)
			require.NoError(t, err)

			assert.True(t, bytes.Equal(before, buf), "the bytes already in the buffer must not change")
			assert.Equal(t, "key="+tc.s.String(), string(got))
			assert.True(t, strings.HasPrefix(string(got), "key="))

			full := []byte("ab")
			got, err = tc.s.AppendText(full[:2:2])
			require.NoError(t, err)
			assert.Equal(t, "ab", string(full), "a full buffer must not change")
			assert.Equal(t, "ab"+tc.s.String(), string(got))
		})
	}
}
