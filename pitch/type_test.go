package pitch_test

import (
	"encoding"
	"fmt"
	"reflect"
	"testing"

	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPitchTypeProperties(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeFor[pitch.Pitch]()
	ptrTyp := reflect.TypeFor[*pitch.Pitch]()

	require.Equal(t, reflect.Struct, typ.Kind())
	assert.True(t, typ.Comparable(), "Pitch must be comparable for == and map keys")
	assert.LessOrEqual(t, typ.Size(), uintptr(40), "Pitch struct size must not exceed 40 bytes")
	assert.Positive(t, typ.NumField(), "Pitch struct must have unexported fields")

	for i := range typ.NumField() {
		field := typ.Field(i)
		assert.False(t, field.IsExported(), "field %s must be unexported to enforce invariants", field.Name)
	}

	// Pitch implements fmt.Stringer and fmt.GoStringer, and does not implement fmt.Formatter
	// or text/JSON marshaling interfaces.
	assert.True(t, typ.Implements(reflect.TypeFor[fmt.Stringer]()))
	assert.True(t, typ.Implements(reflect.TypeFor[fmt.GoStringer]()))
	assert.False(t, typ.Implements(reflect.TypeFor[fmt.Formatter]()))
	assert.False(t, ptrTyp.Implements(reflect.TypeFor[fmt.Formatter]()))

	assert.False(t, typ.Implements(reflect.TypeFor[encoding.TextMarshaler]()))
	assert.False(t, ptrTyp.Implements(reflect.TypeFor[encoding.TextMarshaler]()))
	assert.False(t, typ.Implements(reflect.TypeFor[encoding.TextUnmarshaler]()))
	assert.False(t, ptrTyp.Implements(reflect.TypeFor[encoding.TextUnmarshaler]()))

	for _, methodName := range []string{"MarshalJSON", "UnmarshalJSON", "MarshalText", "UnmarshalText"} {
		_, hasVal := typ.MethodByName(methodName)
		_, hasPtr := ptrTyp.MethodByName(methodName)
		assert.False(t, hasVal, "Pitch must not define %s", methodName)
		assert.False(t, hasPtr, "*Pitch must not define %s", methodName)
	}
}

func TestPitchValueCopySemantics(t *testing.T) {
	t.Parallel()

	original := pa(spelling.FSharp(), 4)
	copied := original

	assert.Equal(t, original, copied)
	assert.True(t, original.Equal(copied))

	// Reassigning derived values to copied leaves original untouched.
	copied = copied.MustRaise().MustUp()
	assert.Equal(t, pa(spelling.FSharp(), 4), original)
	assert.Equal(t, pa(spelling.FDoubleSharp(), 5), copied)
}
