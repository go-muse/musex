package meter_test

import (
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/meter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// enumValue is one value of an enumeration of architecture 4.3 with its
// expected behavior. Name and RussianName are nil for Form.
type enumValue struct {
	name       string
	isValid    bool
	str        string
	equalSelf  func() bool
	equalOther func() bool
	stringer   func() string
	validFn    func() bool
	nameFn     func() (string, bool, error)
	russianFn  func() (string, bool, error)
	en, ru     string // "" means no name
}

// TestEnumerations covers О-19, О-20, П-13, Т-16, КП-26а and architecture 4.3,
// 4.14: IsValid, Equal, String, and the names of every constant, of 0, and of
// a value outside the list.
func TestEnumerations(t *testing.T) {
	t.Parallel()

	var values []enumValue

	addForm := func(v meter.Form, valid bool, str string) {
		values = append(values, enumValue{
			name: "Form " + str, isValid: valid, str: str,
			equalSelf: func() bool { return v.Equal(v) }, equalOther: func() bool { return v.Equal(v + 1) },
			stringer: v.String, validFn: v.IsValid,
		})
	}
	addSymbol := func(v meter.Symbol, valid bool, str, en, ru string) {
		values = append(values, enumValue{
			name: "Symbol " + str, isValid: valid, str: str,
			equalSelf: func() bool { return v.Equal(v) }, equalOther: func() bool { return v.Equal(v + 1) },
			stringer: v.String, validFn: v.IsValid, nameFn: v.Name, russianFn: v.RussianName, en: en, ru: ru,
		})
	}
	addKind := func(v meter.Kind, valid bool, str, en, ru string) {
		values = append(values, enumValue{
			name: "Kind " + str, isValid: valid, str: str,
			equalSelf: func() bool { return v.Equal(v) }, equalOther: func() bool { return v.Equal(v + 1) },
			stringer: v.String, validFn: v.IsValid, nameFn: v.Name, russianFn: v.RussianName, en: en, ru: ru,
		})
	}
	addRussian := func(v meter.RussianKind, valid bool, str, en, ru string) {
		values = append(values, enumValue{
			name: "RussianKind " + str, isValid: valid, str: str,
			equalSelf: func() bool { return v.Equal(v) }, equalOther: func() bool { return v.Equal(v + 1) },
			stringer: v.String, validFn: v.IsValid, nameFn: v.Name, russianFn: v.RussianName, en: en, ru: ru,
		})
	}
	addWeight := func(v meter.Weight, valid bool, str, en, ru string) {
		values = append(values, enumValue{
			name: "Weight " + str, isValid: valid, str: str,
			equalSelf: func() bool { return v.Equal(v) }, equalOther: func() bool { return v.Equal(v + 1) },
			stringer: v.String, validFn: v.IsValid, nameFn: v.Name, russianFn: v.RussianName, en: en, ru: ru,
		})
	}

	addForm(meter.FormSimple, true, "simple")
	addForm(meter.FormSymbol, true, "symbol")
	addForm(meter.FormAdditive, true, "additive")
	addForm(meter.FormSum, true, "sum")
	addForm(0, false, "Form(invalid)")
	addForm(5, false, "Form(invalid)")
	addForm(255, false, "Form(invalid)")

	addSymbol(meter.SymbolCommon, true, "C", "common time", "")
	addSymbol(meter.SymbolCut, true, "C|", "cut time", "alla breve")
	addSymbol(0, false, "Symbol(invalid)", "", "")
	addSymbol(3, false, "Symbol(invalid)", "", "")

	addKind(meter.KindSimple, true, "simple", "simple", "простой")
	addKind(meter.KindCompound, true, "compound", "compound", "сложный")
	addKind(meter.KindAsymmetrical, true, "asymmetrical", "asymmetrical", "асимметричный")
	addKind(meter.KindUndetermined, true, "undetermined", "", "")
	addKind(0, false, "Kind(invalid)", "", "")
	addKind(5, false, "Kind(invalid)", "", "")

	addRussian(meter.RussianSimple, true, "simple", "", "простой")
	addRussian(meter.RussianCompoundUniform, true, "compound-uniform", "", "сложный однородный")
	addRussian(meter.RussianCompoundMixed, true, "compound-mixed", "", "сложный смешанный")
	addRussian(meter.RussianUndetermined, true, "undetermined", "", "")
	addRussian(0, false, "RussianKind(invalid)", "", "")
	addRussian(5, false, "RussianKind(invalid)", "", "")

	addWeight(meter.WeightStrong, true, "strong", "", "сильная")
	addWeight(meter.WeightRelativelyStrong, true, "relatively-strong", "", "относительно сильная")
	addWeight(meter.WeightWeak, true, "weak", "", "слабая")
	addWeight(meter.WeightOffUnit, true, "off-unit", "", "")
	addWeight(0, false, "Weight(invalid)", "", "")
	addWeight(5, false, "Weight(invalid)", "", "")

	for i, v := range values {
		t.Run(v.name+" "+itoa(i), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, v.isValid, v.validFn(), "IsValid")
			assert.Equal(t, v.str, v.stringer(), "String")
			assert.True(t, v.equalSelf(), "Equal to itself")
			assert.False(t, v.equalOther(), "Equal to another value")

			if v.nameFn == nil {
				return
			}

			for _, n := range []struct {
				what string
				f    func() (string, bool, error)
				want string
			}{
				{"Name", v.nameFn, v.en},
				{"RussianName", v.russianFn, v.ru},
			} {
				name, ok, err := n.f()
				if !v.isValid {
					assertErr(t, err, meter.ErrInvalidArgument)
					assert.Empty(t, name, n.what)
					assert.False(t, ok, n.what)

					continue
				}

				require.NoError(t, err, n.what)
				assert.Equal(t, n.want, name, n.what)
				assert.Equal(t, n.want != "", ok, n.what)
			}
		})
	}
}

// TestEnumerationOrder covers architecture 4.3 and КР-M10: constants are
// numbered from 1 in the listed order, and 0 is not valid.
func TestEnumerationOrder(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []int{1, 2, 3, 4}, []int{int(meter.FormSimple), int(meter.FormSymbol), int(meter.FormAdditive), int(meter.FormSum)})
	assert.Equal(t, []int{1, 2}, []int{int(meter.SymbolCommon), int(meter.SymbolCut)})
	assert.Equal(t, []int{1, 2, 3, 4}, []int{
		int(meter.KindSimple), int(meter.KindCompound), int(meter.KindAsymmetrical), int(meter.KindUndetermined),
	})
	assert.Equal(t, []int{1, 2, 3, 4}, []int{
		int(meter.RussianSimple), int(meter.RussianCompoundUniform), int(meter.RussianCompoundMixed), int(meter.RussianUndetermined),
	})
	assert.Equal(t, []int{1, 2, 3, 4}, []int{
		int(meter.WeightStrong), int(meter.WeightRelativelyStrong), int(meter.WeightWeak), int(meter.WeightOffUnit),
	})
}

// TestSpan covers architecture 4.7 and КР-M5: Equal matches ==, End is the
// start plus the duration, and String joins the texts of duration.
func TestSpan(t *testing.T) {
	t.Parallel()

	t.Run("Equal and String", func(t *testing.T) {
		t.Parallel()

		a, b, c := sp("3/8@1/4"), sp("3/8@1/4"), sp("3/8@3/8")
		assert.True(t, a.Equal(b))
		assert.Equal(t, a == b, a.Equal(b))
		assert.False(t, a.Equal(c))
		assert.False(t, a.Equal(meter.Span{}))
		assert.True(t, meter.Span{}.Equal(meter.Span{}))

		assert.Equal(t, "3/8@1/4", a.String())
		assert.Equal(t, "3/8@0/1", sp("3/8@0").String())
		assert.Equal(t, "1/1@2/1", sp("1@2").String())
		assert.Equal(t, "Duration(unset)@Position(unset)", meter.Span{}.String())
	})

	t.Run("End", func(t *testing.T) {
		t.Parallel()

		end, err := sp("3/8@1/4").End()
		require.NoError(t, err)
		assert.Equal(t, posT("5/8"), end)
		assert.Equal(t, posT("3/8"), sp("3/8@0").MustEnd())
	})

	edge := posRat(new(big.Rat).SetInt(maxComponent()))

	for _, c := range []struct {
		name   string
		s      meter.Span
		want   error
		causes []error
	}{
		{"unset", meter.Span{}, meter.ErrInvalidArgument, []error{duration.ErrUnset}},
		{"unset start", meter.Span{Duration: durT("1/4")}, meter.ErrInvalidArgument, []error{duration.ErrUnset}},
		{"unset duration", meter.Span{Start: posT("1/4")}, meter.ErrInvalidArgument, []error{duration.ErrUnset}},
		{"end out of range", meter.Span{Start: edge, Duration: durT("1")}, meter.ErrOutOfRange, []error{fraction.ErrOutOfRange}},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			end, err := c.s.End()
			assertErr(t, err, c.want, c.causes...)
			assert.Equal(t, duration.Position{}, end)
		})
	}
}

// TestNamesKP26a covers КП-26а and Т-16 with the meters of the example:
// kinds in both traditions, weights at positions of 4/4, and the symbols.
func TestNamesKP26a(t *testing.T) {
	t.Parallel()

	type names struct {
		en, ru string // "" means no name
	}

	check := func(t *testing.T, get func() (string, bool, error), want string) {
		t.Helper()

		name, ok, err := get()
		require.NoError(t, err)
		assert.Equal(t, want, name)
		assert.Equal(t, want != "", ok)
	}

	for _, c := range []struct {
		text string
		want names
	}{
		{"6/8", names{"", "сложный однородный"}},
		{"5/8", names{"", "сложный смешанный"}},
	} {
		t.Run("RussianKind of "+c.text, func(t *testing.T) {
			t.Parallel()

			k := mt(c.text).MustRussianKind()
			check(t, k.Name, c.want.en)
			check(t, k.RussianName, c.want.ru)
		})
	}

	for _, c := range []struct {
		at   string
		want names
	}{
		{"0", names{"", "сильная"}},
		{"1/2", names{"", "относительно сильная"}},
	} {
		t.Run("Weight of 4/4 at "+c.at, func(t *testing.T) {
			t.Parallel()

			w := meter.FourFour().MustWeightAt(posT(c.at))
			check(t, w.Name, c.want.en)
			check(t, w.RussianName, c.want.ru)
		})
	}

	for _, c := range []struct {
		m    meter.Meter
		want names
	}{
		{meter.CommonTime(), names{"common time", ""}},
		{meter.CutTime(), names{"cut time", "alla breve"}},
	} {
		t.Run("Symbol of "+c.m.String(), func(t *testing.T) {
			t.Parallel()

			s, ok := c.m.MustSymbol()
			require.True(t, ok)
			check(t, s.Name, c.want.en)
			check(t, s.RussianName, c.want.ru)
		})
	}
}
