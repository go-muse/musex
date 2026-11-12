package tuplet_test

import (
	"encoding"
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
)

// Compile-time checks of architecture 4.1: Tuplet implements fmt.Stringer,
// encoding.TextMarshaler and encoding.TextAppender, and *Tuplet implements
// encoding.TextUnmarshaler.
var (
	_ fmt.Stringer             = tuplet.Tuplet{}
	_ encoding.TextMarshaler   = tuplet.Tuplet{}
	_ encoding.TextAppender    = tuplet.Tuplet{}
	_ encoding.TextUnmarshaler = (*tuplet.Tuplet)(nil)
)

// TestValueType covers NF-2 and architecture 3.1: Tuplet is a comparable
// value of 16 bytes on a 64-bit platform without pointers, slices, maps or
// interfaces.
func TestValueType(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeFor[tuplet.Tuplet]()

	assert.Equal(t, reflect.Struct, typ.Kind())
	assert.True(t, typ.Comparable(), "Tuplet is comparable")
	assert.Equal(t, uintptr(16), unsafe.Sizeof(tuplet.Tuplet{}), "size of Tuplet")

	for i := range typ.NumField() {
		field := typ.Field(i)

		assert.False(t, field.IsExported(), "field %s is unexported", field.Name)
		assert.False(t, containsReference(field.Type), "field %s holds no references", field.Name)
	}
}

// containsReference reports whether values of typ hold pointers, slices,
// maps, channels, functions, interfaces or strings.
func containsReference(typ reflect.Type) bool {
	switch typ.Kind() {
	case reflect.Pointer, reflect.UnsafePointer, reflect.Slice, reflect.Map, reflect.Chan,
		reflect.Func, reflect.Interface, reflect.String:
		return true
	case reflect.Array:
		return containsReference(typ.Elem())
	case reflect.Struct:
		for i := range typ.NumField() {
			if containsReference(typ.Field(i).Type) {
				return true
			}
		}

		return false
	default:
		return false
	}
}
