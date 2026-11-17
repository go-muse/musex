package midi_test

import (
	"fmt"
	"math"
	"sync"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/midi"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/require"
)

func TestNumber_AllValuesAndRoundTrip(t *testing.T) {
	t.Parallel()

	requireNumberValue(t, 0, midi.Min())
	requireNumberValue(t, 127, midi.Max())
	require.Equal(t, 128, midi.Count())

	for value := 0; value < 128; value++ {
		t.Run(fmt.Sprintf("MIDI_%d", value), func(t *testing.T) {
			t.Parallel()

			number, err := midi.New(value)
			require.NoError(t, err)
			requireNumberValue(t, value, number)
			require.Equal(t, value, number.MustInt())
			require.True(t, number.Equal(number))
			require.Equal(t, fmt.Sprintf("MIDI %d", value), number.String())

			o, class, position, err := number.Parts()
			require.NoError(t, err)
			require.True(t, o.IsValid())
			require.Equal(t, value/12-1, o.MustASPN())
			require.True(t, class.IsValid())
			require.Equal(t, value%12, class.Number())
			require.Equal(t, 12, class.Modulus())
			require.Equal(t, value-12, position)

			mustOctave, mustClass, mustPosition := number.MustParts()
			require.True(t, mustOctave.Equal(o))
			require.True(t, mustClass.Equal(class))
			require.Equal(t, position, mustPosition)

			assembled, err := midi.FromParts(o, class)
			require.NoError(t, err)
			requireNumberValue(t, value, assembled)
			requireNumberValue(t, value, midi.MustFromParts(o, class))
		})
	}
}

func TestNumber_AllSequence(t *testing.T) {
	t.Parallel()

	collect := func() []int {
		values := make([]int, 0, midi.Count())
		for number := range midi.All() {
			value, err := number.Int()
			require.NoError(t, err)
			values = append(values, value)
		}
		return values
	}

	expected := make([]int, 128)
	for i := range expected {
		expected[i] = i
	}
	require.Equal(t, expected, collect())
	require.Equal(t, expected, collect(), "All must be repeatable")

	prefix := make([]int, 0, 3)
	for number := range midi.All() {
		value, err := number.Int()
		require.NoError(t, err)
		prefix = append(prefix, value)
		if len(prefix) == 3 {
			break
		}
	}
	require.Equal(t, []int{0, 1, 2}, prefix, "All must support early termination")
}

func TestNumber_AllSequenceParallel(t *testing.T) {
	t.Parallel()

	const workers = 8
	seq := midi.All()
	results := make([][]int, workers)
	var wg sync.WaitGroup
	for i := range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			values := make([]int, 0, midi.Count())
			for number := range seq {
				value, err := number.Int()
				if err != nil {
					return
				}
				values = append(values, value)
			}
			results[i] = values
		}()
	}
	wg.Wait()

	for i, values := range results {
		require.Len(t, values, 128, "worker %d", i)
		for value, actual := range values {
			require.Equal(t, value, actual, "worker %d", i)
		}
	}
}

func TestNumber_NewRejectsValuesOutsideRange(t *testing.T) {
	t.Parallel()

	for _, value := range []int{math.MinInt, -1, 128, math.MaxInt} {
		t.Run(fmt.Sprintf("value_%d", value), func(t *testing.T) {
			t.Parallel()

			number, err := midi.New(value)
			assertMidiError(t, err, midi.ErrInvalidNumber)
			require.Equal(t, midi.Number{}, number)
			requirePanicError(t, err, func() { midi.MustNew(value) })
		})
	}
}

func TestNumber_ZeroValue(t *testing.T) {
	t.Parallel()

	var zero midi.Number
	validZero := midi.Min()
	require.False(t, zero.IsValid())
	require.True(t, validZero.IsValid())
	require.True(t, zero == midi.Number{})
	require.False(t, zero.Equal(zero))
	require.False(t, zero.Equal(validZero))
	require.False(t, validZero.Equal(zero))
	require.Equal(t, 0, zero.Compare(zero))
	require.Equal(t, -1, zero.Compare(validZero))
	require.Equal(t, 1, validZero.Compare(zero))
	require.Equal(t, "Number(invalid)", zero.String())
	require.NotEqual(t, zero.String(), validZero.String())

	value, err := zero.Int()
	assertMidiError(t, err, midi.ErrInvalidNumber)
	require.Zero(t, value)
	requirePanicError(t, err, func() { zero.MustInt() })

	numberOperations := []struct {
		name string
		call func() (midi.Number, error)
		must func()
	}{
		{"Shift", func() (midi.Number, error) { return zero.Shift(0) }, func() { zero.MustShift(0) }},
		{"ShiftOctaves", func() (midi.Number, error) { return zero.ShiftOctaves(0) }, func() { zero.MustShiftOctaves(0) }},
		{"Next", zero.Next, func() { zero.MustNext() }},
		{"Previous", zero.Previous, func() { zero.MustPrevious() }},
		{"Up", zero.Up, func() { zero.MustUp() }},
		{"Down", zero.Down, func() { zero.MustDown() }},
	}
	for _, operation := range numberOperations {
		t.Run(operation.name, func(t *testing.T) {
			result, err := operation.call()
			assertMidiError(t, err, midi.ErrInvalidNumber)
			require.Equal(t, midi.Number{}, result)
			requirePanicError(t, err, operation.must)
		})
	}

	difference, err := zero.Sub(validZero)
	assertMidiError(t, err, midi.ErrInvalidNumber)
	require.Zero(t, difference)
	requirePanicError(t, err, func() { zero.MustSub(validZero) })
	difference, err = validZero.Sub(zero)
	assertMidiError(t, err, midi.ErrInvalidNumber)
	require.Zero(t, difference)
	requirePanicError(t, err, func() { validZero.MustSub(zero) })

	o, class, position, err := zero.Parts()
	assertMidiError(t, err, midi.ErrInvalidNumber)
	require.False(t, o.IsValid())
	require.False(t, class.IsValid())
	require.Zero(t, position)
	requirePanicError(t, err, func() { zero.MustParts() })

	written, err := zero.Pitch(system.Traditional(), system.Sharps())
	assertMidiError(t, err, midi.ErrInvalidNumber)
	require.False(t, written.IsValid())
	requirePanicError(t, err, func() { zero.MustPitch(system.Traditional(), system.Sharps()) })

	created, err := zero.Note(settings.Standard(), system.Sharps(), cents.Zero())
	assertMidiError(t, err, midi.ErrInvalidNumber)
	require.False(t, created.IsValid())
	requirePanicError(t, err, func() { zero.MustNote(settings.Standard(), system.Sharps(), cents.Zero()) })
}

func TestNumber_ShiftsAndDifference(t *testing.T) {
	t.Parallel()

	for value := 0; value < 128; value++ {
		number := midi.MustNew(value)
		for k := -value; k <= 127-value; k++ {
			shifted, err := number.Shift(k)
			require.NoError(t, err, "MIDI %d Shift(%d)", value, k)
			requireNumberValue(t, value+k, shifted)
			requireNumberValue(t, value+k, number.MustShift(k))
		}

		for _, k := range []int{-value - 1, 128 - value} {
			shifted, err := number.Shift(k)
			assertMidiError(t, err, midi.ErrOutOfRange)
			require.Equal(t, midi.Number{}, shifted, "MIDI %d Shift(%d)", value, k)
			requirePanicError(t, err, func() { number.MustShift(k) })
		}

		q := value / 12
		r := value % 12
		maxOctaveShift := (127-r)/12 - q
		for k := -q; k <= maxOctaveShift; k++ {
			shifted, err := number.ShiftOctaves(k)
			require.NoError(t, err, "MIDI %d ShiftOctaves(%d)", value, k)
			requireNumberValue(t, value+12*k, shifted)
			requireNumberValue(t, value+12*k, number.MustShiftOctaves(k))
			_, class, _, err := shifted.Parts()
			require.NoError(t, err)
			require.Equal(t, r, class.Number())
		}
		for _, k := range []int{-q - 1, maxOctaveShift + 1} {
			shifted, err := number.ShiftOctaves(k)
			assertMidiError(t, err, midi.ErrOutOfRange)
			require.Equal(t, midi.Number{}, shifted, "MIDI %d ShiftOctaves(%d)", value, k)
			requirePanicError(t, err, func() { number.MustShiftOctaves(k) })
		}

		next, nextErr := number.Next()
		previous, previousErr := number.Previous()
		up, upErr := number.Up()
		down, downErr := number.Down()
		switch value {
		case 0:
			requireNumberValue(t, 1, next)
			requireNumberValue(t, 1, number.MustNext())
			assertMidiError(t, previousErr, midi.ErrOutOfRange)
			require.Equal(t, midi.Number{}, previous)
			requirePanicError(t, previousErr, func() { number.MustPrevious() })
			requireNumberValue(t, 12, up)
			requireNumberValue(t, 12, number.MustUp())
			assertMidiError(t, downErr, midi.ErrOutOfRange)
			require.Equal(t, midi.Number{}, down)
			requirePanicError(t, downErr, func() { number.MustDown() })
		case 127:
			assertMidiError(t, nextErr, midi.ErrOutOfRange)
			require.Equal(t, midi.Number{}, next)
			requirePanicError(t, nextErr, func() { number.MustNext() })
			requireNumberValue(t, 126, previous)
			requireNumberValue(t, 126, number.MustPrevious())
			assertMidiError(t, upErr, midi.ErrOutOfRange)
			require.Equal(t, midi.Number{}, up)
			requirePanicError(t, upErr, func() { number.MustUp() })
			requireNumberValue(t, 115, down)
			requireNumberValue(t, 115, number.MustDown())
		default:
			requireNumberValue(t, value+1, next)
			requireNumberValue(t, value+1, number.MustNext())
			requireNumberValue(t, value-1, previous)
			requireNumberValue(t, value-1, number.MustPrevious())
			if value <= 115 {
				requireNumberValue(t, value+12, up)
				requireNumberValue(t, value+12, number.MustUp())
			} else {
				assertMidiError(t, upErr, midi.ErrOutOfRange)
				require.Equal(t, midi.Number{}, up)
				requirePanicError(t, upErr, func() { number.MustUp() })
			}
			if value >= 12 {
				requireNumberValue(t, value-12, down)
				requireNumberValue(t, value-12, number.MustDown())
			} else {
				assertMidiError(t, downErr, midi.ErrOutOfRange)
				require.Equal(t, midi.Number{}, down)
				requirePanicError(t, downErr, func() { number.MustDown() })
			}
		}

		for otherValue := 0; otherValue < 128; otherValue++ {
			other := midi.MustNew(otherValue)
			difference, err := number.Sub(other)
			require.NoError(t, err)
			require.Equal(t, value-otherValue, difference)
			require.Equal(t, value-otherValue, number.MustSub(other))
		}
	}
}

func TestNumber_ExtremeShiftArguments(t *testing.T) {
	t.Parallel()

	for _, value := range []int{0, 60, 127} {
		number := midi.MustNew(value)
		for _, k := range []int{math.MinInt, math.MaxInt} {
			t.Run(fmt.Sprintf("MIDI_%d_k_%d", value, k), func(t *testing.T) {
				shifted, err := number.Shift(k)
				assertMidiError(t, err, midi.ErrOutOfRange)
				require.Equal(t, midi.Number{}, shifted)
				requirePanicError(t, err, func() { number.MustShift(k) })

				shifted, err = number.ShiftOctaves(k)
				assertMidiError(t, err, midi.ErrOutOfRange)
				require.Equal(t, midi.Number{}, shifted)
				requirePanicError(t, err, func() { number.MustShiftOctaves(k) })
			})
		}
	}
}

func TestNumber_EqualityAndCompare(t *testing.T) {
	t.Parallel()

	for left := 0; left < 128; left++ {
		for right := 0; right < 128; right++ {
			a := midi.MustNew(left)
			b := midi.MustNew(right)
			require.Equal(t, left == right, a.Equal(b))
			require.Equal(t, left == right, a == b)
			expected := 0
			if left < right {
				expected = -1
			} else if left > right {
				expected = 1
			}
			require.Equal(t, expected, a.Compare(b))
		}
	}

	original := midi.MustNew(60)
	copied := original
	require.True(t, copied.Equal(original))
	shifted := copied.MustShift(1)
	requireNumberValue(t, 61, shifted)
	requireNumberValue(t, 60, copied)
	requireNumberValue(t, 60, original)

	byNumber := map[midi.Number]string{original: "C4"}
	require.Equal(t, "C4", byNumber[shifted.MustShift(-1)])
}

func TestNumber_FromPartsErrors(t *testing.T) {
	t.Parallel()

	validOctave := octave.FromASPN(4)
	validClass := pitchclass.MustNew(3, 12)
	wrongModulus := pitchclass.MustNew(3, 19)

	tests := []struct {
		name       string
		octave     octave.Octave
		class      pitchclass.Class
		midiError  error
		dependency error
	}{
		{"unset octave and class", octave.Octave{}, pitchclass.Class{}, midi.ErrUnset, octave.ErrInvalid},
		{"unset octave before invalid class", octave.Octave{}, validClass, midi.ErrUnset, octave.ErrInvalid},
		{"unset class", validOctave, pitchclass.Class{}, midi.ErrInvalidClass, pitchclass.ErrInvalidValue},
		{"wrong modulus", validOctave, wrongModulus, midi.ErrInvalidClass, pitchclass.ErrModulusMismatch},
		{"wrong modulus before out of range", octave.FromASPN(10), wrongModulus, midi.ErrInvalidClass, pitchclass.ErrModulusMismatch},
		{"below range", octave.FromASPN(-2), validClass, midi.ErrOutOfRange, nil},
		{"above range", octave.FromASPN(9), pitchclass.MustNew(8, 12), midi.ErrOutOfRange, nil},
		{"minimum octave", octave.FromASPN(math.MinInt), validClass, midi.ErrOutOfRange, nil},
		{"maximum octave", octave.FromASPN(math.MaxInt), validClass, midi.ErrOutOfRange, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			number, err := midi.FromParts(tt.octave, tt.class)
			assertMidiError(t, err, tt.midiError)
			if tt.dependency != nil {
				require.ErrorIs(t, err, tt.dependency)
			}
			require.Equal(t, midi.Number{}, number)
			requirePanicError(t, err, func() { midi.MustFromParts(tt.octave, tt.class) })
		})
	}
}
