package spelling_test

import (
	"encoding/json"
	jsonv2 "encoding/json/v2"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// jsonCodec is one of the two JSON libraries of the standard library. The
// contract of the package is the same for both.
type jsonCodec struct {
	name string
	// sortsKeys reports whether marshaling sorts the keys of a map, so that a
	// map with several keys can be compared with an exact literal.
	sortsKeys bool
	marshal   func(v any) ([]byte, error)
	unmarshal func(data []byte, v any) error
}

func jsonCodecs() []jsonCodec {
	return []jsonCodec{
		{
			name:      "encoding/json",
			sortsKeys: true,
			marshal:   json.Marshal,
			unmarshal: json.Unmarshal,
		},
		{
			name: "encoding/json/v2",
			marshal: func(v any) ([]byte, error) {
				return jsonv2.Marshal(v)
			},
			unmarshal: func(data []byte, v any) error {
				return jsonv2.Unmarshal(data, v)
			},
		},
	}
}

// assertJSON compares JSON text: exactly when the codec sorts map keys or the
// text has at most one key per object, and with JSONEq otherwise.
func assertJSON(t *testing.T, codec jsonCodec, exact bool, want string, got []byte) {
	t.Helper()

	if exact || codec.sortsKeys {
		assert.Equal(t, want, string(got))

		return
	}

	assert.JSONEq(t, want, string(got))
}

type spellingField struct {
	S spelling.Spelling `json:"s"`
}

type spellingDocument struct {
	Tonic     spelling.Spelling            `json:"tonic"`
	Spellings []spelling.Spelling          `json:"spellings"`
	By        map[spelling.Spelling]string `json:"by"`
	Inner     spellingField                `json:"inner"`
}

// K-28: a name as a JSON value, a field, an array element, and an object key.
func TestJSONRoundTrip(t *testing.T) {
	t.Parallel()

	for _, codec := range jsonCodecs() {
		t.Run(codec.name, func(t *testing.T) {
			t.Parallel()

			for _, tc := range formNames() {
				x := tc.build()
				form := formOf(tc.l, tc.a)

				data, err := codec.marshal(x)
				require.NoError(t, err, tc.name)
				assert.Equal(t, `"`+form+`"`, string(data), "value %s", tc.name)

				again, err := codec.marshal(x)
				require.NoError(t, err)
				assert.Equal(t, data, again, "writing again gives the same JSON")

				var value spelling.Spelling

				require.NoError(t, codec.unmarshal(data, &value), tc.name)
				assertName(t, tc.l, tc.a, value)

				data, err = codec.marshal(spellingField{S: x})
				require.NoError(t, err)
				assert.Equal(t, `{"s":"`+form+`"}`, string(data), "field %s", tc.name)

				field := spellingField{S: spelling.DFlat()}
				require.NoError(t, codec.unmarshal(data, &field))
				assertName(t, tc.l, tc.a, field.S)

				data, err = codec.marshal([]spelling.Spelling{x})
				require.NoError(t, err)
				assert.Equal(t, `["`+form+`"]`, string(data), "element %s", tc.name)

				var elements []spelling.Spelling

				require.NoError(t, codec.unmarshal(data, &elements))
				require.Len(t, elements, 1)
				assertName(t, tc.l, tc.a, elements[0])

				data, err = codec.marshal(map[spelling.Spelling]int{x: 1})
				require.NoError(t, err)
				assert.Equal(t, `{"`+form+`":1}`, string(data), "key %s", tc.name)

				var keys map[spelling.Spelling]int

				require.NoError(t, codec.unmarshal(data, &keys))
				require.Len(t, keys, 1)

				for k, v := range keys {
					assertName(t, tc.l, tc.a, k)
					assert.Equal(t, 1, v)
				}
			}
		})
	}
}

// K-28, K-30: composite data with set and unset names.
func TestJSONDocument(t *testing.T) {
	t.Parallel()

	doc := spellingDocument{
		Tonic:     spelling.CSharp(),
		Spellings: []spelling.Spelling{spelling.DFlat(), {}, spelling.MustNew(letter.B, frac(-7, 2))},
		By: map[spelling.Spelling]string{
			spelling.BFlat():                       "B",
			spelling.C():                           "C",
			{}:                                     "none",
			spelling.MustNew(letter.C, frac(1, 2)): "half",
		},
		Inner: spellingField{},
	}
	want := `{"tonic":"C+1","spellings":["D-1","unset","B-7/2"],` +
		`"by":{"B-1":"B","C":"C","C+1/2":"half","unset":"none"},"inner":{"s":"unset"}}`

	for _, codec := range jsonCodecs() {
		t.Run(codec.name, func(t *testing.T) {
			t.Parallel()

			data, err := codec.marshal(doc)
			require.NoError(t, err)
			assertJSON(t, codec, false, want, data)

			var got spellingDocument

			require.NoError(t, codec.unmarshal(data, &got))
			assert.Equal(t, doc, got)
			assert.False(t, got.Spellings[1].IsValid(), "the unset element stays unset")
			assert.False(t, got.Inner.S.IsValid(), "the unset field stays unset")
			assert.Equal(t, "none", got.By[spelling.Spelling{}], "the unset key stays unset")
		})
	}

	t.Run("encoding/json/v2 deterministic", func(t *testing.T) {
		t.Parallel()

		data, err := jsonv2.Marshal(doc, jsonv2.Deterministic(true))
		require.NoError(t, err)
		assert.JSONEq(t, want, string(data))

		again, err := jsonv2.Marshal(doc, jsonv2.Deterministic(true))
		require.NoError(t, err)
		assert.Equal(t, data, again)
	})

	t.Run("example of the architecture", func(t *testing.T) {
		t.Parallel()

		type key struct {
			Tonic     spelling.Spelling            `json:"tonic"`
			Spellings []spelling.Spelling          `json:"spellings"`
			By        map[spelling.Spelling]string `json:"by"`
		}

		in := key{
			Tonic:     spelling.CSharp(),
			Spellings: []spelling.Spelling{spelling.DFlat(), {}},
			By:        map[spelling.Spelling]string{spelling.BFlat(): "B"},
		}

		for _, codec := range jsonCodecs() {
			data, err := codec.marshal(in)
			require.NoError(t, err, codec.name)
			assert.Equal(t, `{"tonic":"C+1","spellings":["D-1","unset"],"by":{"B-1":"B"}}`, string(data), codec.name)

			var out key

			require.NoError(t, codec.unmarshal(data, &out), codec.name)
			assert.Equal(t, in, out, codec.name)
		}
	})
}

// K-30: the unset name is written as the mark "unset" in every position, and
// the mark is read back as the unset name.
func TestJSONUnset(t *testing.T) {
	t.Parallel()

	var unset spelling.Spelling

	for _, codec := range jsonCodecs() {
		t.Run(codec.name, func(t *testing.T) {
			t.Parallel()

			cases := []struct {
				name string
				v    any
				want string
			}{
				{"value", unset, `"unset"`},
				{"zero literal", spelling.Spelling{}, `"unset"`},
				{"field", spellingField{}, `{"s":"unset"}`},
				{"element", []spelling.Spelling{unset, spelling.C()}, `["unset","C"]`},
				{"key", map[spelling.Spelling]int{unset: 1}, `{"unset":1}`},
			}

			for _, tc := range cases {
				data, err := codec.marshal(tc.v)
				require.NoError(t, err, tc.name)
				assert.Equal(t, tc.want, string(data), tc.name)

				again, err := codec.marshal(tc.v)
				require.NoError(t, err)
				assert.Equal(t, data, again, "writing again: %s", tc.name)
			}

			value := spelling.C()
			require.NoError(t, codec.unmarshal([]byte(`"unset"`), &value))
			assertUnset(t, value)

			field := spellingField{S: spelling.DFlat()}
			require.NoError(t, codec.unmarshal([]byte(`{"s":"unset"}`), &field))
			assertUnset(t, field.S)

			var keys map[spelling.Spelling]int

			require.NoError(t, codec.unmarshal([]byte(`{"unset":1}`), &keys))
			assert.Equal(t, map[spelling.Spelling]int{{}: 1}, keys)
		})
	}

	t.Run("direct", func(t *testing.T) {
		t.Parallel()

		data, err := unset.MarshalJSON()
		require.NoError(t, err)
		assert.Equal(t, `"unset"`, string(data))

		receiver := spelling.DFlat()
		require.NoError(t, receiver.UnmarshalJSON([]byte(`"unset"`)))
		assertUnset(t, receiver)
	})
}

// K-29, K-30: JSON null is read as the unset name, not C, and is not an error.
func TestJSONNull(t *testing.T) {
	t.Parallel()

	for _, codec := range jsonCodecs() {
		t.Run(codec.name, func(t *testing.T) {
			t.Parallel()

			for _, before := range []spelling.Spelling{spelling.DFlat(), spelling.C()} {
				field := spellingField{S: before}
				require.NoError(t, codec.unmarshal([]byte(`{"s":null}`), &field), before.String())
				assertUnset(t, field.S)
				assert.False(t, field.S.Equal(spelling.C()))
			}
		})
	}

	t.Run("direct", func(t *testing.T) {
		t.Parallel()

		for _, before := range []spelling.Spelling{spelling.DFlat(), spelling.C()} {
			receiver := before
			require.NoError(t, receiver.UnmarshalJSON([]byte("null")), before.String())
			assertUnset(t, receiver)
		}
	})
}

// K-29: JSON that is not a string with a machine form is a reading error; the
// field keeps its value.
func TestJSONInvalid(t *testing.T) {
	t.Parallel()

	inputs := []struct {
		name  string
		value string
		class formErrorClass
	}{
		{"number", `1`, formOnly},
		{"true", `true`, formOnly},
		{"false", `false`, formOnly},
		{"array", `[]`, formOnly},
		{"array of a form", `["C+1"]`, formOnly},
		{"object", `{}`, formOnly},
		{"object with parts", `{"letter":"C"}`, formOnly},
		{"empty string", `""`, formOnly},
		{"notation", `"C#"`, formOnly},
		{"escaped notation", `"C\u0023"`, formOnly},
		{"escaped unset with a space", `"unset\u0020"`, formOnly},
		{"zero denominator", `"C+1/0"`, formZeroDenominator},
		{"out of range", `"C+16777217"`, formOutOfRange},
	}

	for _, codec := range jsonCodecs() {
		t.Run(codec.name, func(t *testing.T) {
			t.Parallel()

			for _, in := range inputs {
				field := spellingField{S: spelling.DFlat()}
				err := codec.unmarshal([]byte(`{"s":`+in.value+`}`), &field)

				if assert.Error(t, err, in.name) {
					assertInvalidForm(t, in.class, err)
				}

				assert.Equal(t, spelling.DFlat(), field.S, "the field must keep its value: %s", in.name)
				assert.NotEqual(t, letter.C, field.S.Letter(), in.name)
			}
		})
	}

	t.Run("direct", func(t *testing.T) {
		t.Parallel()

		direct := []struct {
			name string
			data []byte
		}{
			{"nil", nil},
			{"empty", []byte("")},
			{"number", []byte(`1`)},
			{"true", []byte(`true`)},
			{"array", []byte(`[]`)},
			{"object", []byte(`{}`)},
			{"empty string", []byte(`""`)},
			{"notation", []byte(`"C#"`)},
			{"escaped notation", []byte(`"C\u0023"`)},
			{"unquoted form", []byte(`C+1`)},
			{"unterminated string", []byte(`"C+1`)},
			// JSON has no \x escape; the method itself must reject it.
			{"invalid escape", []byte(`"C\x2B1"`)},
			{"two values", []byte(`"C+1" "D"`)},
		}

		for _, tc := range direct {
			receiver := spelling.DFlat()
			err := receiver.UnmarshalJSON(tc.data)
			assertInvalidForm(t, formOnly, err)
			assert.Equal(t, spelling.DFlat(), receiver, "the receiver must keep its value: %s", tc.name)
		}
	})
}

// The architecture, finding 12 of review v2: a JSON string is read by its
// decoded content, so escaped characters are allowed. Canonical form is
// checked on the decoded content, not on the JSON text.
func TestJSONEscapedString(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		data string
		want spelling.Spelling
	}{
		{"escaped plus", `"C\u002B1"`, spelling.CSharp()},
		{"escaped plus in lower case", `"C\u002b1"`, spelling.CSharp()},
		{"escaped letter and plus", `"\u0043\u002B1"`, spelling.CSharp()},
		{"escaped slash", `"B-7\u002F2"`, spelling.MustNew(letter.B, frac(-7, 2))},
		{"escaped unset", `"\u0075nset"`, spelling.Spelling{}},
		{"escaped solidus", `"B-7\/2"`, spelling.MustNew(letter.B, frac(-7, 2))},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Contains(t, tc.data, `\`, "the input must contain a JSON escape")

			for _, codec := range jsonCodecs() {
				value := spelling.DFlat()
				require.NoError(t, codec.unmarshal([]byte(tc.data), &value), codec.name)
				assert.Equal(t, tc.want, value, codec.name)
				assert.Equal(t, tc.want.IsValid(), value.IsValid(), codec.name)

				field := spellingField{S: spelling.DFlat()}
				require.NoError(t, codec.unmarshal([]byte(`{"s":`+tc.data+`}`), &field), codec.name)
				assert.Equal(t, tc.want, field.S, codec.name)
			}

			receiver := spelling.DFlat()
			require.NoError(t, receiver.UnmarshalJSON([]byte(tc.data)), "direct")
			assert.Equal(t, tc.want, receiver, "direct")
		})
	}

	t.Run("escaped plus, by parts", func(t *testing.T) {
		t.Parallel()

		receiver := spelling.DFlat()
		require.NoError(t, receiver.UnmarshalJSON([]byte(`"C\u002B1"`)))
		assertName(t, letter.C, alteration.Sharp(), receiver)
	})
}

// The architecture, type Direction: the numbers of the constants are stable,
// and JSON encoders write them as numbers. The test checks the observed JSON,
// so a text form of Direction in JSON would make it fail.
func TestDirectionJSON(t *testing.T) {
	t.Parallel()

	type directionField struct {
		D spelling.Direction `json:"d"`
	}

	cases := []struct {
		name string
		d    spelling.Direction
		json string
	}{
		{"DirectionUnset", spelling.DirectionUnset, "0"},
		{"DirectionLowered", spelling.DirectionLowered, "1"},
		{"DirectionNatural", spelling.DirectionNatural, "2"},
		{"DirectionRaised", spelling.DirectionRaised, "3"},
	}

	for _, codec := range jsonCodecs() {
		t.Run(codec.name, func(t *testing.T) {
			t.Parallel()

			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					t.Parallel()

					data, err := codec.marshal(tc.d)
					require.NoError(t, err, tc.json)
					assert.Equal(t, tc.json, string(data), "value %v", tc.d)

					var value spelling.Direction
					if tc.d == spelling.DirectionUnset {
						value = spelling.DirectionRaised
					}

					require.NoError(t, codec.unmarshal([]byte(tc.json), &value), tc.json)
					assert.Equal(t, tc.d, value, "read %s", tc.json)

					data, err = codec.marshal(directionField{D: tc.d})
					require.NoError(t, err, tc.json)
					assert.Equal(t, `{"d":`+tc.json+`}`, string(data), "field %v", tc.d)

					field := directionField{D: spelling.DirectionRaised}
					if tc.d == spelling.DirectionRaised {
						field.D = spelling.DirectionUnset
					}

					require.NoError(t, codec.unmarshal([]byte(`{"d":`+tc.json+`}`), &field), tc.json)
					assert.Equal(t, tc.d, field.D, "read field %s", tc.json)
				})
			}
		})
	}
}
