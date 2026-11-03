package octave_test

import (
	"fmt"
	"reflect"
	"slices"
	"testing"

	"github.com/go-muse/musex/octave"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The type implements fmt.Stringer and fmt.GoStringer with value receivers.
var (
	_ fmt.Stringer   = octave.Octave{}
	_ fmt.GoStringer = octave.Octave{}
)

// KR-OC1, KR-OC2, K-30 (type part): Octave is a comparable struct with
// unexported fields only, at most two machine words.
func TestTypeShape(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeFor[octave.Octave]()
	require.Equal(t, reflect.Struct, typ.Kind())
	assert.True(t, typ.Comparable(), "comparable")
	assert.LessOrEqual(t, typ.Size(), uintptr(16), "size in bytes")
	for i := range typ.NumField() {
		f := typ.Field(i)
		assert.False(t, f.IsExported(), "field %s is exported", f.Name)
	}
}

// KR-OC1, architecture "Summary": the methods of Octave are exactly the 25 of
// the contract, all with value receivers (the method set of the value type).
func TestMethodSet(t *testing.T) {
	t.Parallel()

	want := []string{
		"ASPN", "Above", "Below", "Between", "Compare", "Distance", "Down",
		"Equal", "GoString", "IsValid", "MustASPN", "MustDistance", "MustDown",
		"MustNumber", "MustShift", "MustShiftDown", "MustSub", "MustUp", "Name",
		"Number", "Shift", "ShiftDown", "String", "Sub", "Up",
	}

	typ := reflect.TypeFor[octave.Octave]()
	got := make([]string, 0, typ.NumMethod())
	for i := range typ.NumMethod() {
		got = append(got, typ.Method(i).Name)
	}
	slices.Sort(got)
	assert.Equal(t, want, got)
}

// Architecture "Copying": assignment copies an octave; the copy is equal to
// the original by == and Equal and behaves the same way.
func TestCopy(t *testing.T) {
	t.Parallel()

	for _, a := range []int{minInt, -1, 0, 4, maxInt} {
		orig := oct(a)
		cp := orig
		assert.True(t, orig == cp, "O(%d) copy ==", a)
		assert.True(t, orig.Equal(cp), "O(%d) copy Equal", a)
		assert.Equal(t, orig.String(), cp.String(), "O(%d) copy String", a)
		assert.Equal(t, a, cp.MustASPN(), "O(%d) copy ASPN", a)

		moved := cp.MustShift(0)
		assert.True(t, orig == moved, "O(%d) Shift(0) of the copy", a)
	}

	var z octave.Octave
	zc := z
	assert.True(t, z == zc, "copy of Z ==")
	assert.False(t, zc.IsValid(), "copy of Z is invalid")
}
