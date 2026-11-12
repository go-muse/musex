package tuplet_test

import (
	jsonv1 "encoding/json"
	jsonv2 "encoding/json/v2"
	"math"
	"testing"

	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestString covers O-16 and architecture 4.12: canonical n:m output and the
// special string of the unset tuplet.
func TestString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		tp    tuplet.Tuplet
		text  string
		short string
	}{
		{tp: tuplet.MustNew(3, 2), text: "3:2", short: "3"},
		{tp: tuplet.MustNew(6, 4), text: "6:4", short: "6"},
		{tp: tuplet.MustNew(1, 1), text: "1:1", short: "1"},
		{tp: tuplet.MustNew(10, 8), text: "10:8", short: "10"},
		{tp: tuplet.MustNew(math.MaxInt, 1), text: "9223372036854775807:1", short: "9223372036854775807"},
		{
			tp:    tuplet.MustNew(math.MaxInt, math.MaxInt),
			text:  "9223372036854775807:9223372036854775807",
			short: "9223372036854775807",
		},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.text, tt.tp.String())

			short, err := tt.tp.ShortString()
			require.NoError(t, err)
			assert.Equal(t, tt.short, short)
			assert.Equal(t, tt.short, tt.tp.MustShortString())

			marshaled, err := tt.tp.MarshalText()
			require.NoError(t, err)
			assert.Equal(t, tt.text, string(marshaled))
		})
	}

	assert.Equal(t, "Tuplet(unset)", tuplet.Tuplet{}.String())
}

// TestParse_RoundTrip covers K-11 and G-2: Parse(t.String()) == t and
// UnmarshalText(MarshalText(t)) gives t for every set tuplet.
func TestParse_RoundTrip(t *testing.T) {
	t.Parallel()

	for _, tp := range sampleTuplets() {
		t.Run(tp.String(), func(t *testing.T) {
			t.Parallel()

			parsed, err := tuplet.Parse(tp.String())
			require.NoError(t, err)
			assert.Equal(t, tp, parsed)

			text, err := tp.MarshalText()
			require.NoError(t, err)

			var got tuplet.Tuplet
			require.NoError(t, got.UnmarshalText(text))
			assert.Equal(t, tp, got)
		})
	}
}

// TestParse covers architecture 4.12: accepted canonical input.
func TestParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		text           string
		actual, normal int
	}{
		{text: "3:2", actual: 3, normal: 2},
		{text: "6:4", actual: 6, normal: 4},
		{text: "1:1", actual: 1, normal: 1},
		{text: "10:20", actual: 10, normal: 20},
		{text: "9223372036854775807:9223372036854775807", actual: math.MaxInt, normal: math.MaxInt},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()

			got, err := tuplet.Parse(tt.text)
			require.NoError(t, err)
			assertTuplet(t, got, tt.actual, tt.normal)
		})
	}
}

// TestParse_Errors covers K-11, E-4, E-5 and the order E-4 -> E-5.
func TestParse_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		text string
		want error
	}{
		{text: "", want: tuplet.ErrInvalidText},
		{text: " 3:2", want: tuplet.ErrInvalidText},
		{text: "3:2 ", want: tuplet.ErrInvalidText},
		{text: "3 :2", want: tuplet.ErrInvalidText},
		{text: "3: 2", want: tuplet.ErrInvalidText},
		{text: "+3:2", want: tuplet.ErrInvalidText},
		{text: "-3:2", want: tuplet.ErrInvalidText},
		{text: "3:-2", want: tuplet.ErrInvalidText},
		{text: "03:2", want: tuplet.ErrInvalidText},
		{text: "3:02", want: tuplet.ErrInvalidText},
		{text: "3:", want: tuplet.ErrInvalidText},
		{text: ":2", want: tuplet.ErrInvalidText},
		{text: ":", want: tuplet.ErrInvalidText},
		{text: "3", want: tuplet.ErrInvalidText},
		{text: "3:2:1", want: tuplet.ErrInvalidText},
		{text: "0:2", want: tuplet.ErrInvalidText},
		{text: "3:0", want: tuplet.ErrInvalidText},
		{text: "a:b", want: tuplet.ErrInvalidText},
		{text: "3/2", want: tuplet.ErrInvalidText},
		{text: "3:2.0", want: tuplet.ErrInvalidText},
		{text: "３:２", want: tuplet.ErrInvalidText},
		{text: "Triplet", want: tuplet.ErrInvalidText},
		{text: "triplet", want: tuplet.ErrInvalidText},
		{text: "дуплет", want: tuplet.ErrInvalidText},
		{text: "Tuplet(unset)", want: tuplet.ErrInvalidText},
		{text: "3 pour 2", want: tuplet.ErrInvalidText},
		{text: "(3:2:3", want: tuplet.ErrInvalidText},
		{text: "9223372036854775808:2", want: tuplet.ErrOutOfRange},
		{text: "2:9223372036854775808", want: tuplet.ErrOutOfRange},
		{text: "99999999999999999999999:1", want: tuplet.ErrOutOfRange},
		{text: "9223372036854775808:0", want: tuplet.ErrInvalidText},
		{text: "9223372036854775808:2:1", want: tuplet.ErrInvalidText},
		{text: "09223372036854775808:2", want: tuplet.ErrInvalidText},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()

			got, err := tuplet.Parse(tt.text)
			assertErr(t, err, tt.want)
			assert.Equal(t, tuplet.Tuplet{}, got)

			original := tuplet.Quintuplet()
			target := original
			err = target.UnmarshalText([]byte(tt.text))
			assertErr(t, err, tt.want)
			assert.Equal(t, original, target, "UnmarshalText keeps the receiver on error")
		})
	}
}

// TestParseShort covers K-11 and architecture 4.12: the short form with a
// division gives Default(n, division).
func TestParseShort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		text           string
		division       tuplet.Division
		actual, normal int
	}{
		{text: "3", division: tuplet.Binary, actual: 3, normal: 2},
		{text: "2", division: tuplet.Ternary, actual: 2, normal: 3},
		{text: "5", division: tuplet.Ternary, actual: 5, normal: 3},
		{text: "7", division: tuplet.Binary, actual: 7, normal: 4},
		{text: "9223372036854775807", division: tuplet.Binary, actual: math.MaxInt, normal: 1 << 62},
	}

	for _, tt := range tests {
		t.Run(tt.text+"/"+tt.division.String(), func(t *testing.T) {
			t.Parallel()

			got, err := tuplet.ParseShort(tt.text, tt.division)
			require.NoError(t, err)
			assertTuplet(t, got, tt.actual, tt.normal)
			assert.Equal(t, tuplet.MustDefault(tt.actual, tt.division), got)
			assert.Equal(t, got, tuplet.MustParseShort(tt.text, tt.division))
		})
	}
}

// TestParseShort_RoundTrip covers architecture 4.12: parsing the short
// output gives Default(n, d), not necessarily the original tuplet.
func TestParseShort_RoundTrip(t *testing.T) {
	t.Parallel()

	for _, tp := range []tuplet.Tuplet{tuplet.Triplet(), tuplet.Sextuplet(), tuplet.MustNew(5, 3), tuplet.MustNew(7, 8)} {
		for _, division := range []tuplet.Division{tuplet.Binary, tuplet.Ternary} {
			want, wantErr := tuplet.Default(tp.MustActual(), division)

			got, err := tuplet.ParseShort(tp.MustShortString(), division)
			if wantErr != nil {
				assertErr(t, err, tuplet.ErrNoDefault)

				continue
			}

			require.NoError(t, err)
			assert.Equal(t, want, got)
		}
	}
}

// TestParseShort_Errors covers E-3, E-4, E-5 and the order
// E-4 -> E-5 -> E-3 (unknown division) -> E-3 (no default).
func TestParseShort_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		text     string
		division tuplet.Division
		want     error
	}{
		{name: "pair", text: "3:2", division: tuplet.Binary, want: tuplet.ErrInvalidText},
		{name: "leading zero", text: "03", division: tuplet.Binary, want: tuplet.ErrInvalidText},
		{name: "zero", text: "0", division: tuplet.Binary, want: tuplet.ErrInvalidText},
		{name: "empty", text: "", division: tuplet.Binary, want: tuplet.ErrInvalidText},
		{name: "sign", text: "+3", division: tuplet.Binary, want: tuplet.ErrInvalidText},
		{name: "minus", text: "-3", division: tuplet.Binary, want: tuplet.ErrInvalidText},
		{name: "space", text: " 3", division: tuplet.Binary, want: tuplet.ErrInvalidText},
		{name: "name", text: "triplet", division: tuplet.Binary, want: tuplet.ErrInvalidText},
		{name: "overflow", text: "9223372036854775808", division: tuplet.Binary, want: tuplet.ErrOutOfRange},
		{name: "no default for one", text: "1", division: tuplet.Binary, want: tuplet.ErrNoDefault},
		{name: "binary member", text: "4", division: tuplet.Binary, want: tuplet.ErrNoDefault},
		{name: "ternary member", text: "6", division: tuplet.Ternary, want: tuplet.ErrNoDefault},
		{name: "unknown division", text: "3", division: tuplet.Division(0), want: tuplet.ErrNoDefault},
		{name: "unknown division 3", text: "5", division: tuplet.Division(3), want: tuplet.ErrNoDefault},
		{name: "grammar before range", text: "09223372036854775808", division: tuplet.Binary, want: tuplet.ErrInvalidText},
		{name: "grammar before division", text: "03", division: tuplet.Division(0), want: tuplet.ErrInvalidText},
		{name: "range before division", text: "9223372036854775808", division: tuplet.Division(0), want: tuplet.ErrOutOfRange},
		{name: "unknown division before no default", text: "4", division: tuplet.Division(0), want: tuplet.ErrNoDefault},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tuplet.ParseShort(tt.text, tt.division)
			assertErr(t, err, tt.want)
			assert.Equal(t, tuplet.Tuplet{}, got)
		})
	}
}

// TestParseName covers K-2 and architecture 4.12: English names, alternative
// forms and Russian names give the named tuplet for the count.
func TestParseName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		text string
		want tuplet.Tuplet
	}{
		{text: "duplet", want: tuplet.Duplet()},
		{text: "triplet", want: tuplet.Triplet()},
		{text: "quadruplet", want: tuplet.Quadruplet()},
		{text: "quintuplet", want: tuplet.Quintuplet()},
		{text: "sextuplet", want: tuplet.Sextuplet()},
		{text: "septuplet", want: tuplet.Septuplet()},
		{text: "octuplet", want: tuplet.Octuplet()},
		{text: "nonuplet", want: tuplet.Nonuplet()},
		{text: "decuplet", want: tuplet.Decuplet()},
		{text: "quartole", want: tuplet.Quadruplet()},
		{text: "quintolet", want: tuplet.Quintuplet()},
		{text: "pentuplet", want: tuplet.Quintuplet()},
		{text: "sextolet", want: tuplet.Sextuplet()},
		{text: "sestole", want: tuplet.Sextuplet()},
		{text: "septolet", want: tuplet.Septuplet()},
		{text: "septimole", want: tuplet.Septuplet()},
		{text: "дуоль", want: tuplet.Duplet()},
		{text: "триоль", want: tuplet.Triplet()},
		{text: "квартоль", want: tuplet.Quadruplet()},
		{text: "квинтоль", want: tuplet.Quintuplet()},
		{text: "секстоль", want: tuplet.Sextuplet()},
		{text: "септоль", want: tuplet.Septuplet()},
		{text: "октоль", want: tuplet.Octuplet()},
		{text: "нонемоль", want: tuplet.Nonuplet()},
		{text: "децимоль", want: tuplet.Decuplet()},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()

			got, err := tuplet.ParseName(tt.text)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want, tuplet.MustParseName(tt.text))
		})
	}
}

// TestParseName_Names covers K-2: parsing every name returned by Name and
// RussianName of the named tuplets gives the same pair.
func TestParseName_Names(t *testing.T) {
	t.Parallel()

	for _, nt := range namedTuplets {
		name, ok, err := nt.fn().Name()
		require.NoError(t, err)
		require.True(t, ok)

		russian, ok, err := nt.fn().RussianName()
		require.NoError(t, err)
		require.True(t, ok)

		fromName, err := tuplet.ParseName(name)
		require.NoError(t, err)
		assert.Equal(t, nt.fn(), fromName)

		fromRussian, err := tuplet.ParseName(russian)
		require.NoError(t, err)
		assert.Equal(t, nt.fn(), fromRussian)
	}
}

// TestParseName_Errors covers K-11, E-4 and architecture 4.11, 4.12.
func TestParseName_Errors(t *testing.T) {
	t.Parallel()

	for _, text := range []string{
		"", "Triplet", "TRIPLET", "Триоль", "дуплет", "триплет", " triplet", "triplet ",
		"3:2", "3", "tuplet", "undecuplet", "Tuplet(unset)", "triplets",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()

			got, err := tuplet.ParseName(text)
			assertErr(t, err, tuplet.ErrInvalidText)
			assert.Equal(t, tuplet.Tuplet{}, got)
		})
	}
}

// TestAppendText covers architecture 3.2 and 4.12: AppendText appends the
// canonical text without changing b[:len(b)]; on the unset tuplet it returns
// b unchanged with E-2.
func TestAppendText(t *testing.T) {
	t.Parallel()

	t.Run("set", func(t *testing.T) {
		t.Parallel()

		b := make([]byte, 3, 32)
		copy(b, "ab,")
		backing := b[:cap(b)]
		before := append([]byte(nil), backing...)

		got, err := tuplet.Sextuplet().AppendText(b)
		require.NoError(t, err)
		assert.Equal(t, "ab,6:4", string(got))
		assert.Equal(t, "ab,", string(b))
		assert.Equal(t, before[:3], backing[:3], "existing bytes are unchanged")

		got, err = tuplet.Triplet().AppendText(nil)
		require.NoError(t, err)
		assert.Equal(t, "3:2", string(got))
	})

	t.Run("unset", func(t *testing.T) {
		t.Parallel()

		b := make([]byte, 3, 32)
		copy(b, "ab,")
		backing := b[:cap(b)]
		before := append([]byte(nil), backing...)

		got, err := tuplet.Tuplet{}.AppendText(b)
		assertErr(t, err, tuplet.ErrUnset)
		assert.Equal(t, "ab,", string(got))
		assert.Len(t, got, 3)
		assert.Equal(t, before, backing, "the buffer is unchanged")
	})
}

// TestMarshalText_NewSlice covers architecture 3.2: MarshalText returns a new
// slice each time.
func TestMarshalText_NewSlice(t *testing.T) {
	t.Parallel()

	first, err := tuplet.Triplet().MarshalText()
	require.NoError(t, err)

	first[0] = 'x'

	second, err := tuplet.Triplet().MarshalText()
	require.NoError(t, err)
	assert.Equal(t, "3:2", string(second))
}

// TestUnmarshalText covers architecture 3.2 and 4.12: success sets the
// receiver; the input is not changed or retained; a nil receiver gives E-6.
func TestUnmarshalText(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		text := []byte("6:4")

		var got tuplet.Tuplet
		require.NoError(t, got.UnmarshalText(text))
		assert.Equal(t, tuplet.Sextuplet(), got)
		assert.Equal(t, "6:4", string(text), "the input is unchanged")

		copy(text, "9:8")
		assert.Equal(t, tuplet.Sextuplet(), got, "the input is not retained")
	})

	t.Run("overwrites a set receiver", func(t *testing.T) {
		t.Parallel()

		got := tuplet.Triplet()
		require.NoError(t, got.UnmarshalText([]byte("5:4")))
		assert.Equal(t, tuplet.Quintuplet(), got)
	})

	t.Run("nil receiver", func(t *testing.T) {
		t.Parallel()

		var nilTarget *tuplet.Tuplet

		assertErr(t, nilTarget.UnmarshalText([]byte("3:2")), tuplet.ErrInvalidArgument)
		assertErr(t, nilTarget.UnmarshalText([]byte("bad")), tuplet.ErrInvalidArgument)
		assertErr(t, nilTarget.UnmarshalText(nil), tuplet.ErrInvalidArgument)
	})
}

// jsonHolder is a structure with a required and an optional tuplet field.
type jsonHolder struct {
	Required tuplet.Tuplet `json:"required"`
	Optional tuplet.Tuplet `json:"optional,omitzero"`
}

// TestJSON covers architecture 4.12 and risk R-T5: a field is encoded as the
// canonical string; an unset field is an error with ErrUnset unless omitzero
// omits it.
func TestJSON(t *testing.T) {
	t.Parallel()

	codecs := []struct {
		name      string
		marshal   func(any) ([]byte, error)
		unmarshal func([]byte, any) error
	}{
		{
			name:      "v1",
			marshal:   jsonv1.Marshal,
			unmarshal: jsonv1.Unmarshal,
		},
		{
			name:      "v2",
			marshal:   func(v any) ([]byte, error) { return jsonv2.Marshal(v) },
			unmarshal: func(data []byte, v any) error { return jsonv2.Unmarshal(data, v) },
		},
	}

	for _, codec := range codecs {
		t.Run(codec.name, func(t *testing.T) {
			t.Parallel()

			data, err := codec.marshal(jsonHolder{Required: tuplet.Triplet(), Optional: tuplet.Sextuplet()})
			require.NoError(t, err)
			assert.JSONEq(t, `{"required":"3:2","optional":"6:4"}`, string(data))

			data, err = codec.marshal(jsonHolder{Required: tuplet.Triplet()})
			require.NoError(t, err)
			assert.JSONEq(t, `{"required":"3:2"}`, string(data))

			_, err = codec.marshal(jsonHolder{})
			require.Error(t, err)
			assert.ErrorIs(t, err, tuplet.ErrUnset)

			var decoded jsonHolder
			require.NoError(t, codec.unmarshal([]byte(`{"required":"5:4","optional":"6:4"}`), &decoded))
			assert.Equal(t, jsonHolder{Required: tuplet.Quintuplet(), Optional: tuplet.Sextuplet()}, decoded)

			err = codec.unmarshal([]byte(`{"required":"03:2"}`), &decoded)
			require.Error(t, err)
			assert.ErrorIs(t, err, tuplet.ErrInvalidText)
		})
	}
}
