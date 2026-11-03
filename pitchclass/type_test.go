package pitchclass_test

import (
	"reflect"
	"testing"

	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/assert"
)

// Architecture, types Class and Operator (KR-PC1, NF-2): comparable structs
// with unexported fields only, usable as map keys; Style is uint8 with
// Decimal as its zero value.
func TestTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		typ  reflect.Type
	}{
		{name: "Class", typ: reflect.TypeFor[pitchclass.Class]()},
		{name: "Operator", typ: reflect.TypeFor[pitchclass.Operator]()},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, reflect.Struct, tc.typ.Kind())
			assert.True(t, tc.typ.Comparable(), "comparable, usable as a map key")
			for i := range tc.typ.NumField() {
				f := tc.typ.Field(i)
				assert.False(t, f.IsExported(), "field %s is unexported", f.Name)
			}
		})
	}

	t.Run("Style", func(t *testing.T) {
		t.Parallel()

		typ := reflect.TypeFor[pitchclass.Style]()
		assert.Equal(t, reflect.Uint8, typ.Kind())

		var zero pitchclass.Style
		assert.Equal(t, pitchclass.Decimal, zero, "the zero Style is Decimal")

		seen := map[pitchclass.Style]bool{}
		for _, st := range styles() {
			assert.False(t, seen[st], "%v is distinct", st)
			seen[st] = true
		}
		assert.Len(t, seen, 4)
	})

	t.Run("zero values", func(t *testing.T) {
		t.Parallel()

		var c pitchclass.Class
		var f pitchclass.Operator
		assert.Equal(t, pitchclass.Class{}, c)
		assert.Equal(t, pitchclass.Operator{}, f)
	})
}
