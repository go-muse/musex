package cents_test

import (
	jsonv1 "encoding/json"
	jsonv2 "encoding/json/v2"
	"strings"
	"testing"

	"github.com/go-muse/musex/cents"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// C-23, C-24: MarshalText, AppendText and UnmarshalText.
func TestTextMarshaling(t *testing.T) {
	t.Parallel()

	t.Run("MarshalText and AppendText for all S", func(t *testing.T) {
		t.Parallel()

		for _, a := range setS(t) {
			text, err := a.v.MarshalText()
			require.NoError(t, err)
			assert.Equal(t, a.v.String(), string(text), a.name)

			out, err := a.v.AppendText(nil)
			require.NoError(t, err)
			assert.Equal(t, a.v.String(), string(out), a.name)

			// prefix[:len(prefix)] never changes; the spare capacity may
			// receive the text (KR-C16; the capacity cases are in
			// TestAppendTextBuffer).
			prefix := make([]byte, 3, 64)
			copy(prefix, "ab:")
			out, err = a.v.AppendText(prefix)
			require.NoError(t, err)
			assert.Equal(t, "ab:"+a.v.String(), string(out), a.name)
			assert.Equal(t, "ab:", string(prefix), "the prefix is unchanged")

			// A prefix of 1 byte in an array of 3 (explicit capacity: the
			// capacity of []byte("...") is not defined by the language).
			full := make([]byte, 3)
			copy(full, "xyz")
			out, err = a.v.AppendText(full[:1])
			require.NoError(t, err)
			assert.Equal(t, "x"+a.v.String(), string(out), a.name)
			assert.Equal(t, "x", string(full[:1]), "the prefix is unchanged")
		}
	})

	t.Run("UnmarshalText", func(t *testing.T) {
		t.Parallel()

		for _, a := range setS(t) {
			var c cents.Cents
			text := []byte(a.v.String())
			require.NoError(t, c.UnmarshalText(text), a.name)
			assertCents(t, a.v, c, a.name)
			assert.Equal(t, a.v.String(), string(text), "the input is unchanged")
		}

		c := cents.Octave()
		require.NoError(t, c.UnmarshalText([]byte("3:2")))
		assertCents(t, noErr(cents.FromRatio(3, 2))(t), c, "the receiver is overwritten")

		require.NoError(t, c.UnmarshalText([]byte("+30 ¢")))
		assertComp(t, ratC(30, 1), c, "any accepted text")
	})

	t.Run("UnmarshalText errors keep the receiver", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name string
			text []byte
			err  error
		}{
			{"nil", nil, cents.ErrInvalidText},
			{"empty", []byte{}, cents.ErrInvalidText},
			{"bare a/b", []byte("3/2"), cents.ErrInvalidText},
			{"prime", []byte("37:32"), cents.ErrPrimeLimit},
			// The prime of a base is checked whatever its exponent
			// (architecture "Errors", confirmed by the architect).
			{"prime with exponent 0", []byte("2*37^0"), cents.ErrPrimeLimit},
			{"range", []byte("9000000.000001"), cents.ErrOutOfRange},
			{"length", []byte("1" + strings.Repeat("0", 100)), cents.ErrTooComplex},
			{"length of a fraction", []byte("1/" + strings.Repeat("0", 100) + "7 cents"), cents.ErrTooComplex},
		}

		for _, tc := range cases {
			c := cents.SyntonicComma()
			err := c.UnmarshalText(tc.text)
			assertOnlyError(t, err, tc.err, tc.name)
			assertCents(t, cents.SyntonicComma(), c, "%s: the receiver is unchanged", tc.name)
		}
	})
}

// fifthRecord is a struct with a value field for the JSON tests.
type fifthRecord struct {
	Name  string      `json:"name"`
	Value cents.Cents `json:"value"`
}

// C-23, C-24: JSON through the standard text interfaces, encoding/json (v1).
func TestJSONv1(t *testing.T) {
	t.Parallel()

	r32 := noErr(cents.FromRatio(3, 2))(t)
	e119 := noErr(cents.Step(1, 19))(t)

	data, err := jsonv1.Marshal(r32)
	require.NoError(t, err)
	assert.JSONEq(t, `"3:2"`, string(data))

	data, err = jsonv1.Marshal(e119)
	require.NoError(t, err)
	assert.JSONEq(t, `"1200/19 cents"`, string(data))

	data, err = jsonv1.Marshal(fifthRecord{Name: "fifth", Value: r32})
	require.NoError(t, err)
	assert.JSONEq(t, `{"name":"fifth","value":"3:2"}`, string(data))

	data, err = jsonv1.Marshal([]cents.Cents{r32, cents.Zero(), e119})
	require.NoError(t, err)
	assert.JSONEq(t, `["3:2","0","1200/19 cents"]`, string(data))

	data, err = jsonv1.Marshal(map[cents.Cents]string{r32: "fifth", e119: "step"})
	require.NoError(t, err)
	assert.JSONEq(t, `{"3:2":"fifth","1200/19 cents":"step"}`, string(data))

	var rec fifthRecord
	require.NoError(t, jsonv1.Unmarshal([]byte(`{"name":"q","value":"5^(1/4)"}`), &rec))
	assertComp(t, zeroComp().withExp(5, rat(1, 4)), rec.Value)

	var m map[cents.Cents]string
	require.NoError(t, jsonv1.Unmarshal([]byte(`{"3:2":"fifth","6:4":"same","1200/19 cents":"step"}`), &m))
	assert.Len(t, m, 2, "3:2 and 6:4 are one key")
	assert.Equal(t, "step", m[e119])

	t.Run("round trip for all S", func(t *testing.T) {
		t.Parallel()

		set := setS(t)
		values := make([]cents.Cents, 0, len(set))
		for _, a := range set {
			values = append(values, a.v)
		}

		data, err := jsonv1.Marshal(values)
		require.NoError(t, err)

		var back []cents.Cents
		require.NoError(t, jsonv1.Unmarshal(data, &back))
		require.Len(t, back, len(values))

		for i, a := range set {
			assertCents(t, a.v, back[i], a.name)
		}
	})

	t.Run("errors", func(t *testing.T) {
		t.Parallel()

		var c cents.Cents
		assert.ErrorIs(t, jsonv1.Unmarshal([]byte(`"3/2"`), &c), cents.ErrInvalidText)
		assert.ErrorIs(t, jsonv1.Unmarshal([]byte(`"37:32"`), &c), cents.ErrPrimeLimit)

		var rec fifthRecord
		assert.ErrorIs(t, jsonv1.Unmarshal([]byte(`{"value":"cents"}`), &rec), cents.ErrInvalidText)
	})
}

// C-23, C-24: JSON through the standard text interfaces, encoding/json/v2.
// Map keys are not ordered by v2: objects are compared with JSONEq or
// marshaled with Deterministic(true).
func TestJSONv2(t *testing.T) {
	t.Parallel()

	r32 := noErr(cents.FromRatio(3, 2))(t)
	e119 := noErr(cents.Step(1, 19))(t)

	data, err := jsonv2.Marshal(r32)
	require.NoError(t, err)
	assert.Equal(t, `"3:2"`, string(data))

	data, err = jsonv2.Marshal(fifthRecord{Name: "fifth", Value: r32})
	require.NoError(t, err)
	assert.JSONEq(t, `{"name":"fifth","value":"3:2"}`, string(data))

	data, err = jsonv2.Marshal([]cents.Cents{r32, cents.Zero(), e119})
	require.NoError(t, err)
	assert.Equal(t, `["3:2","0","1200/19 cents"]`, string(data))

	data, err = jsonv2.Marshal(map[cents.Cents]string{r32: "fifth", e119: "step"}, jsonv2.Deterministic(true))
	require.NoError(t, err)
	assert.JSONEq(t, `{"3:2":"fifth","1200/19 cents":"step"}`, string(data))

	var rec fifthRecord
	require.NoError(t, jsonv2.Unmarshal([]byte(`{"name":"q","value":"5^(1/4)"}`), &rec))
	assertComp(t, zeroComp().withExp(5, rat(1, 4)), rec.Value)

	var m map[cents.Cents]string
	require.NoError(t, jsonv2.Unmarshal([]byte(`{"3:2":"fifth","1200/19 cents":"step"}`), &m))
	assert.Len(t, m, 2)
	assert.Equal(t, "fifth", m[r32])

	t.Run("round trip for all S", func(t *testing.T) {
		t.Parallel()

		set := setS(t)
		values := make([]cents.Cents, 0, len(set))
		for _, a := range set {
			values = append(values, a.v)
		}

		data, err := jsonv2.Marshal(values)
		require.NoError(t, err)

		var back []cents.Cents
		require.NoError(t, jsonv2.Unmarshal(data, &back))
		require.Len(t, back, len(values))

		for i, a := range set {
			assertCents(t, a.v, back[i], a.name)
		}
	})

	t.Run("errors", func(t *testing.T) {
		t.Parallel()

		var c cents.Cents
		assert.ErrorIs(t, jsonv2.Unmarshal([]byte(`"3/2"`), &c), cents.ErrInvalidText)
		assert.ErrorIs(t, jsonv2.Unmarshal([]byte(`"37:32"`), &c), cents.ErrPrimeLimit)
	})
}
