package cents_test

import (
	"encoding"
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/go-muse/musex/cents"

	"github.com/stretchr/testify/assert"
)

// The type Cents is a comparable struct with unexported fields only;
// Quotient is not comparable (architecture, "Type Cents", "Quotient").
func TestTypeProperties(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeFor[cents.Cents]()
	assert.Equal(t, reflect.Struct, typ.Kind())
	assert.True(t, typ.Comparable(), "Cents is comparable: == and map keys")

	for i := range typ.NumField() {
		assert.False(t, typ.Field(i).IsExported(), "Cents field %s", typ.Field(i).Name)
	}

	quo := reflect.TypeFor[cents.Quotient]()
	assert.Equal(t, reflect.Struct, quo.Kind())
	assert.False(t, quo.Comparable(), "Quotient is not comparable")

	for i := range quo.NumField() {
		assert.False(t, quo.Field(i).IsExported(), "Quotient field %s", quo.Field(i).Name)
	}

	assert.Equal(t, reflect.Uint8, reflect.TypeFor[cents.Kind]().Kind())
	assert.Equal(t, reflect.Uint8, reflect.TypeFor[cents.Options]().Kind())
}

// Power and NamedValue have exactly the fields of the summary.
func TestExportedStructs(t *testing.T) {
	t.Parallel()

	fields := func(typ reflect.Type) map[string]reflect.Type {
		out := map[string]reflect.Type{}
		for i := range typ.NumField() {
			out[typ.Field(i).Name] = typ.Field(i).Type
		}

		return out
	}

	assert.Equal(t, map[string]reflect.Type{
		"Base": reflect.TypeFor[int](),
		"Num":  reflect.TypeFor[*big.Int](),
		"Den":  reflect.TypeFor[*big.Int](),
	}, fields(reflect.TypeFor[cents.Power]()))

	assert.Equal(t, map[string]reflect.Type{
		"Name":  reflect.TypeFor[string](),
		"Value": reflect.TypeFor[cents.Cents](),
	}, fields(reflect.TypeFor[cents.NamedValue]()))
}

// Cents implements fmt.Stringer, fmt.Formatter, encoding.TextMarshaler and
// encoding.TextAppender by value; only *Cents implements
// encoding.TextUnmarshaler.
func TestInterfaces(t *testing.T) {
	t.Parallel()

	var (
		c cents.Cents

		stringer  fmt.Stringer           = c
		formatter fmt.Formatter          = c
		marshaler encoding.TextMarshaler = c
		appender  encoding.TextAppender  = c

		unmarshaler encoding.TextUnmarshaler = &c
	)

	assert.NotNil(t, stringer)
	assert.NotNil(t, formatter)
	assert.NotNil(t, marshaler)
	assert.NotNil(t, appender)
	assert.NotNil(t, unmarshaler)

	typ := reflect.TypeFor[cents.Cents]()
	assert.False(t, typ.Implements(reflect.TypeFor[encoding.TextUnmarshaler]()),
		"UnmarshalText has a pointer receiver")
	assert.True(t, reflect.PointerTo(typ).Implements(reflect.TypeFor[encoding.TextUnmarshaler]()))

	for _, iface := range []reflect.Type{
		reflect.TypeFor[fmt.Stringer](),
		reflect.TypeFor[fmt.Formatter](),
		reflect.TypeFor[encoding.TextMarshaler](),
		reflect.TypeFor[encoding.TextAppender](),
	} {
		assert.True(t, typ.Implements(iface), "Cents implements %s", iface)
	}

	assert.True(t, reflect.TypeFor[cents.Quotient]().Implements(reflect.TypeFor[fmt.Stringer]()))
}

// Constants of the summary.
func TestConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 9000000, cents.MaxMagnitude)
	assert.Equal(t, 31, cents.MaxPrime)
	assert.Equal(t, 1000000, cents.GridDenominator)
	assert.Equal(t, 100, cents.MaxDigits)
	assert.Equal(t, 64, cents.MaxFactors)

	assert.Equal(t, cents.Kind(0), cents.KindDecimal)
	assert.Equal(t, cents.Kind(1), cents.KindFraction)
	assert.Equal(t, cents.Kind(2), cents.KindRatio)
	assert.Equal(t, cents.Kind(3), cents.KindPower)

	var zeroKind cents.Kind
	assert.Equal(t, cents.KindDecimal, zeroKind, "the zero Kind is KindDecimal")

	assert.Equal(t, cents.Options(1), cents.Typographic)
	assert.Equal(t, cents.Options(2), cents.WithUnit)
	assert.Equal(t, cents.Options(4), cents.WithPlus)
}

// Kind.String: the four names and "Kind(n)" for other values.
func TestKindString(t *testing.T) {
	t.Parallel()

	cases := []struct {
		kind cents.Kind
		want string
	}{
		{cents.KindDecimal, "decimal"},
		{cents.KindFraction, "fraction"},
		{cents.KindRatio, "ratio"},
		{cents.KindPower, "power"},
		{cents.Kind(4), "Kind(4)"},
		{cents.Kind(9), "Kind(9)"},
		{cents.Kind(255), "Kind(255)"},
	}

	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, tc.kind.String())
		})
	}
}

// The five errors are distinct values.
func TestErrorsAreDistinct(t *testing.T) {
	t.Parallel()

	errs := allErrors()
	for i, a := range errs {
		assert.Error(t, a)

		for j, b := range errs {
			assert.Equal(t, i == j, a == b, "%v and %v", a, b)
		}
	}
}
