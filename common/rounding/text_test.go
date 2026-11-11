package rounding_test

import (
	jsonv1 "encoding/json"
	jsonv2 "encoding/json/v2"
	"fmt"
	"testing"

	"github.com/go-muse/musex/common/rounding"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModeString(t *testing.T) {
	t.Parallel()
	for i, mode := range append(append([]rounding.Mode{0}, testModes...), 10, 255) {
		want := "unset"
		if i > 0 && i <= len(testModeNames) {
			want = testModeNames[i-1]
		} else if i > len(testModeNames) {
			want = fmt.Sprintf("Mode(%d)", mode)
		}
		t.Run(fmt.Sprintf("mode_%d", mode), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, want, mode.String())
		})
	}
}

func TestSideAndNeighborStrings(t *testing.T) {
	t.Parallel()
	for value := -128; value <= 127; value++ {
		side := rounding.Side(value)
		t.Run(fmt.Sprintf("side_%d", value), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, value == -1 || value == 1, side.IsValid())
		})
	}
	for _, tt := range []struct {
		value rounding.Side
		want  string
	}{{rounding.Negative, "negative"}, {rounding.Positive, "positive"}, {0, "Side(0)"}, {-2, "Side(-2)"}, {127, "Side(127)"}} {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.value.String())
		})
	}
	for _, tt := range []struct {
		value rounding.Neighbor
		want  string
	}{{rounding.Lower, "lower"}, {rounding.Upper, "upper"}, {0, "Neighbor(0)"}, {3, "Neighbor(3)"}, {255, "Neighbor(255)"}} {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.value.String())
		})
	}
}

func TestParse(t *testing.T) {
	t.Parallel()
	for i, name := range testModeNames {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, err := rounding.Parse(name)
			require.NoError(t, err)
			assert.Equal(t, testModes[i], got)
		})
	}
	for _, name := range []string{"HALF-EVEN", " half-even", "half-even ", "", "unset", "Mode(1)", "1", "half_even", "halfeven", "½", "half-even\x00"} {
		t.Run(fmt.Sprintf("invalid_%q", name), func(t *testing.T) {
			t.Parallel()
			got, err := rounding.Parse(name)
			checkError(t, err, rounding.ErrInvalidArgument)
			assert.Equal(t, rounding.Mode(0), got)
		})
	}
}

func TestModeUnmarshalText(t *testing.T) {
	t.Parallel()
	for i, name := range testModeNames {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			mode := rounding.Mode(0)
			text := []byte(name)
			before := append([]byte(nil), text...)
			require.NoError(t, mode.UnmarshalText(text))
			assert.Equal(t, testModes[i], mode)
			assert.Equal(t, before, text)
		})
	}
	for _, text := range [][]byte{[]byte("HALF-EVEN"), []byte(" half-even"), []byte("half-even "), {}, []byte("unset"), []byte("Mode(1)"), []byte("1"), []byte("half_even"), []byte("halfeven"), []byte("½"), nil} {
		t.Run(fmt.Sprintf("invalid_%q", text), func(t *testing.T) {
			t.Parallel()
			mode := rounding.HalfUp
			before := append([]byte(nil), text...)
			checkError(t, mode.UnmarshalText(text), rounding.ErrInvalidArgument)
			assert.Equal(t, rounding.HalfUp, mode)
			assert.Equal(t, before, text)
		})
	}
}

func TestModeNameAndMarshalText(t *testing.T) {
	t.Parallel()
	for i, mode := range testModes {
		t.Run(testModeNames[i], func(t *testing.T) {
			t.Parallel()
			name, err := mode.Name()
			require.NoError(t, err)
			assert.Equal(t, testModeNames[i], name)
			text, err := mode.MarshalText()
			require.NoError(t, err)
			assert.Equal(t, testModeNames[i], string(text))
			text[0] = 'X'
			again, err := mode.MarshalText()
			require.NoError(t, err)
			assert.Equal(t, testModeNames[i], string(again))
		})
	}
	for _, tt := range []struct {
		mode rounding.Mode
		want error
	}{{0, rounding.ErrUnset}, {10, rounding.ErrUnknownMode}, {255, rounding.ErrUnknownMode}} {
		t.Run(fmt.Sprintf("error_%d", tt.mode), func(t *testing.T) {
			t.Parallel()
			name, err := tt.mode.Name()
			checkError(t, err, tt.want)
			assert.Empty(t, name)
			text, err := tt.mode.MarshalText()
			checkError(t, err, tt.want)
			assert.Nil(t, text)
		})
	}
}

func TestModeAppendText(t *testing.T) {
	t.Parallel()
	t.Run("reuses capacity", func(t *testing.T) {
		t.Parallel()
		buffer := make([]byte, 2, 64)
		copy(buffer, "xy")
		got, err := rounding.HalfEven.AppendText(buffer)
		require.NoError(t, err)
		assert.Equal(t, "xyhalf-even", string(got))
		assert.Equal(t, "xy", string(buffer))
		assert.True(t, &got[0] == &buffer[0])
	})
	t.Run("allocates when needed", func(t *testing.T) {
		t.Parallel()
		buffer := []byte("xy")
		got, err := rounding.Down.AppendText(buffer)
		require.NoError(t, err)
		assert.Equal(t, "xydown", string(got))
		assert.Equal(t, "xy", string(buffer))
		assert.True(t, &got[0] != &buffer[0])
	})
	for _, tt := range []struct {
		mode rounding.Mode
		want error
	}{{0, rounding.ErrUnset}, {10, rounding.ErrUnknownMode}} {
		t.Run(fmt.Sprintf("error_%d", tt.mode), func(t *testing.T) {
			t.Parallel()
			buffer := make([]byte, 2, 64)
			copy(buffer, "xy")
			got, err := tt.mode.AppendText(buffer)
			checkError(t, err, tt.want)
			assert.Nil(t, got)
			assert.Equal(t, "xy", string(buffer))
		})
	}
}

func TestModeJSON(t *testing.T) {
	t.Parallel()
	type payload struct {
		Mode rounding.Mode `json:"mode"`
	}
	for _, tt := range []struct {
		name      string
		marshal   func(any) ([]byte, error)
		unmarshal func([]byte, any) error
	}{
		{"json", jsonv1.Marshal, jsonv1.Unmarshal},
		{"json_v2", func(v any) ([]byte, error) { return jsonv2.Marshal(v) }, func(data []byte, v any) error { return jsonv2.Unmarshal(data, v) }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			encoded, err := tt.marshal(payload{Mode: rounding.HalfEven})
			require.NoError(t, err)
			assert.JSONEq(t, `{"mode":"half-even"}`, string(encoded))
			var decoded payload
			require.NoError(t, tt.unmarshal(encoded, &decoded))
			assert.Equal(t, rounding.HalfEven, decoded.Mode)
			_, err = tt.marshal(payload{})
			checkError(t, err, rounding.ErrUnset)
		})
	}
}
