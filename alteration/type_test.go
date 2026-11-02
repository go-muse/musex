package alteration_test

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/go-muse/musex/alteration"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NF-7, K-22 (test part), architecture "Type Alteration": the type is a
// comparable struct of at most 16 bytes with unexported fields of integer
// types only — no pointers, slices, maps, interfaces or math/big inside.
// This observes the public type through reflection; field names are not
// inspected.
func TestTypeLayout(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeFor[alteration.Alteration]()

	assert.LessOrEqual(t, unsafe.Sizeof(alteration.Alteration{}), uintptr(16), "fixed size of at most 16 bytes")
	assert.True(t, typ.Comparable(), "usable with == and as a map key")
	require.Equal(t, reflect.Struct, typ.Kind())

	integerKinds := []reflect.Kind{
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
	}

	for i := range typ.NumField() {
		field := typ.Field(i)
		assert.False(t, field.IsExported(), "field %d is unexported", i)
		assert.Contains(t, integerKinds, field.Type.Kind(), "field %d has an integer type", i)
	}
}

// Architecture "Type Alteration": assignment is a full copy; changing the
// copy through the API leaves the original value as it was.
func TestCopy(t *testing.T) {
	t.Parallel()

	x := alteration.Sharp()
	y := x

	y, err := y.Raise()
	require.NoError(t, err)

	assert.Equal(t, frac{2, 1}, fractionOf(y))
	assert.Equal(t, frac{1, 1}, fractionOf(x), "the original is still a sharp")
	assertSame(t, alteration.Sharp(), x)
}

// Architecture "Type Alteration": methods with value receivers are available
// on *Alteration as well.
func TestPointerMethodSet(t *testing.T) {
	t.Parallel()

	a := alteration.Sharp()
	p := &a

	up, err := p.Raise()
	require.NoError(t, err)
	assertSame(t, alteration.DoubleSharp(), up)
	assertSame(t, alteration.Sharp(), a, "the value behind the pointer is unchanged")
	assert.True(t, p.Equal(alteration.Sharp()))
	assert.Equal(t, "+1", p.String())
}
