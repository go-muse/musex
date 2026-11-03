package spelling_test

import (
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The type Direction: its values, their numbers, and their text.
func TestDirectionType(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		d      spelling.Direction
		number uint8
		text   string
	}{
		{"unset", spelling.DirectionUnset, 0, "unset"},
		{"lowered", spelling.DirectionLowered, 1, "lowered"},
		{"natural", spelling.DirectionNatural, 2, "natural"},
		{"raised", spelling.DirectionRaised, 3, "raised"},
		{"other value", spelling.Direction(9), 9, "Direction(9)"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.number, uint8(tc.d), "the number of the constant")
			assert.Equal(t, tc.text, tc.d.String())
			assert.Equal(t, tc.text, fmt.Sprint(tc.d))
		})
	}

	t.Run("zero value", func(t *testing.T) {
		t.Parallel()

		var d spelling.Direction

		assert.Equal(t, spelling.DirectionUnset, d)
	})

	t.Run("comparable", func(t *testing.T) {
		t.Parallel()

		typ := reflect.TypeFor[spelling.Direction]()
		assert.Equal(t, reflect.Uint8, typ.Kind())
		assert.True(t, typ.Comparable())

		m := map[spelling.Direction]int{spelling.DirectionRaised: 1, spelling.DirectionLowered: 2}
		assert.Equal(t, 1, m[spelling.CSharp().Direction()])
		assert.Equal(t, 2, m[spelling.DFlat().Direction()])
	})
}

// Properties of the type Spelling from the architecture: a comparable struct
// without exported fields, at most 24 bytes, with the encoding interfaces on
// value receivers and the reading interfaces on the pointer.
func TestSpellingType(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeFor[spelling.Spelling]()

	assert.Equal(t, reflect.Struct, typ.Kind())
	assert.True(t, typ.Comparable())

	for i := range typ.NumField() {
		assert.Falsef(t, typ.Field(i).IsExported(), "field %s must not be exported", typ.Field(i).Name)
	}

	assert.LessOrEqual(t, unsafe.Sizeof(spelling.Spelling{}), uintptr(24))

	var value any = spelling.CSharp()

	_, ok := value.(fmt.Stringer)
	assert.True(t, ok, "fmt.Stringer")
	_, ok = value.(encoding.TextMarshaler)
	assert.True(t, ok, "encoding.TextMarshaler")
	_, ok = value.(encoding.TextAppender)
	assert.True(t, ok, "encoding.TextAppender")
	_, ok = value.(json.Marshaler)
	assert.True(t, ok, "json.Marshaler")

	var pointer any = new(spelling.Spelling)

	_, ok = pointer.(encoding.TextUnmarshaler)
	assert.True(t, ok, "encoding.TextUnmarshaler")
	_, ok = pointer.(json.Unmarshaler)
	assert.True(t, ok, "json.Unmarshaler")
}

// Copying a name gives an independent value.
func TestSpellingCopy(t *testing.T) {
	t.Parallel()

	x := spelling.CSharp()
	y := x
	y = y.MustRaise()

	assertName(t, letter.C, alteration.DoubleSharp(), y)
	assertName(t, letter.C, alteration.Sharp(), x)

	z := x
	require.NoError(t, z.UnmarshalText([]byte("B-7/2")))
	assertName(t, letter.B, frac(-7, 2), z)
	assertName(t, letter.C, alteration.Sharp(), x)
}

// K-20: no action through the public API changes the ready-made names for
// other calls.
func TestReadyMadeNamesCannotBeChanged(t *testing.T) {
	t.Parallel()

	for _, tc := range predefinedCases() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := tc.fn()
			s = s.MustRaise()
			s = s.MustRaiseBy(5)
			s = s.MustRaiseByAlteration(frac(1, 2))
			s = s.WithAlteration(whole(7))
			s = s.Natural()
			require.NoError(t, s.UnmarshalText([]byte("G-3/2")))
			require.NoError(t, s.UnmarshalJSON([]byte("null")))

			p := new(spelling.Spelling)
			*p = tc.fn()
			assert.True(t, p.IsValid())
			*p = spelling.MustNew(letter.G, whole(9))
			require.NoError(t, p.UnmarshalJSON([]byte(`"A+1/3"`)))

			other := tc.fn()
			assert.True(t, other.Equal(tc.fn()))
			other = spelling.D()
			assertName(t, letter.D, alteration.Natural(), other)

			again := tc.fn()
			assertName(t, tc.l, tc.a, again)
			assert.True(t, again.Equal(spelling.MustNew(tc.l, tc.a)))
		})
	}

	t.Run("through Predefined", func(t *testing.T) {
		t.Parallel()

		list := spelling.Predefined()
		for i := range list {
			require.NoError(t, list[i].UnmarshalText([]byte("unset")))
		}

		for _, tc := range predefinedCases() {
			assertName(t, tc.l, tc.a, tc.fn())
		}

		for i, s := range spelling.Predefined() {
			assert.Truef(t, s.IsValid(), "element %d", i)
		}
	})
}
