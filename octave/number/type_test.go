package number_test

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/go-muse/musex/octave/number"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// K-19, architecture "Summary": the public operations have exactly these
// signatures; octaves are given by ASPN numbers of type int and methods have
// value receivers (method expressions on number.Numbering). The assignments
// fail to compile if a signature differs.
var (
	_ func() number.Numbering   = number.ASPN
	_ func() number.Numbering   = number.FrancoBelgian
	_ func() number.Numbering   = number.Mexican
	_ func() number.Numbering   = number.MiddleC3
	_ func() number.Numbering   = number.MiddleC5
	_ func() []number.Numbering = number.Builtins

	_ func(aspn, num int, skipped ...int) (number.Numbering, error) = number.New
	_ func(aspn, num int, skipped ...int) number.Numbering          = number.MustNew

	_ func(x int, from, to number.Numbering) (int, error) = number.Convert
	_ func(x int, from, to number.Numbering) int          = number.MustConvert

	_ func(n number.Numbering, aspn int) (int, error)       = number.Numbering.FromASPN
	_ func(n number.Numbering, aspn int) int                = number.Numbering.MustFromASPN
	_ func(n number.Numbering, num int) (int, error)        = number.Numbering.ToASPN
	_ func(n number.Numbering, num int) int                 = number.Numbering.MustToASPN
	_ func(n number.Numbering, other number.Numbering) bool = number.Numbering.Equal
	_ func(n number.Numbering) bool                         = number.Numbering.IsDefined
	_ func(n number.Numbering) string                       = number.Numbering.Label
	_ func(n number.Numbering) string                       = number.Numbering.String

	_ error = number.ErrSkipped
	_ error = number.ErrInvalidDefinition
	_ error = number.ErrOutOfRange
	_ error = number.ErrUndefined

	_ fmt.Stringer                          = number.Numbering{}
	_ interface{ ToASPN(int) (int, error) } = number.Numbering{}
)

// The contract and the boundary cases assume a 64-bit int (docs/rules/dev.md:
// only 64-bit systems are supported).
func TestPlatformIntIs64Bit(t *testing.T) {
	t.Parallel()

	require.Equal(t, 64, strconv.IntSize)
}

// Architecture "Type Numbering": a struct with unexported fields only, not
// comparable (== does not compile, the type cannot be a map key). Observed
// through reflect on the public type, without reading the implementation.
func TestNumberingType(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeFor[number.Numbering]()

	require.Equal(t, reflect.Struct, typ.Kind())
	assert.False(t, typ.Comparable(), "Numbering must not be comparable")

	for i := range typ.NumField() {
		f := typ.Field(i)
		assert.False(t, f.IsExported(), "exported field %s", f.Name)
	}
}

// Architecture "Type Numbering": methods with value receivers are also
// available on *Numbering.
func TestNumberingPointerMethods(t *testing.T) {
	t.Parallel()

	n := number.FrancoBelgian()
	p := &n

	assert.Equal(t, "Franco-Belgian", p.Label())
	assert.True(t, p.IsDefined())
	assert.True(t, p.Equal(number.FrancoBelgian()))
	assert.Equal(t, "Franco-Belgian", p.String())

	x, err := p.FromASPN(1)
	assertOutcome(t, outcome{value: -1}, x, err, "FromASPN(1)")
	a, err := p.ToASPN(-1)
	assertOutcome(t, outcome{value: 1}, a, err, "ToASPN(-1)")
}
