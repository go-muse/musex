package scale_test

import (
	"testing"

	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/scale"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func FuzzCircularListProperties(f *testing.F) {
	f.Add([]byte{0, 1, 2}, int64(7))
	f.Add([]byte{0, 0, 1, 0}, int64(-8))
	f.Add([]byte{}, int64(0))
	f.Fuzz(func(t *testing.T, data []byte, distance int64) {
		if len(data) > 64 {
			return
		}
		palette := names("C D E F G A B")
		ns := make([]note.Note, len(data))
		for i, b := range data {
			ns[i] = palette[int(b)%len(palette)]
		}
		s := scale.MustNew(ns...)
		reverse, err := s.Reverse()
		require.NoError(t, err)
		twice, err := reverse.Reverse()
		require.NoError(t, err)
		assert.True(t, s.Equal(twice))
		unique, err := s.Unique()
		require.NoError(t, err)
		union, err := s.Union(s)
		require.NoError(t, err)
		assert.True(t, unique.Equal(union))
		difference, err := s.Difference(s)
		require.NoError(t, err)
		assert.True(t, difference.IsEmpty())
		subset, err := unique.IsSubsetOf(s)
		require.NoError(t, err)
		assert.True(t, subset)
		if len(ns) == 0 {
			return
		}
		index := int(uint64(distance) % uint64(len(ns)))
		moved, err := s.Move(index, int(distance))
		require.NoError(t, err)
		// Reduce first, so the oracle itself cannot overflow.
		expected := (index + int(distance)%len(ns) + len(ns)) % len(ns)
		assert.Equal(t, expected, moved.Index)
		for _, d := range []int{len(ns), -len(ns)} {
			got, err := s.Move(index, d)
			require.NoError(t, err)
			assert.Equal(t, index, got.Index)
			assert.True(t, ns[index].Equal(got.Note))
		}
		rotation, err := s.Rotate(index)
		require.NoError(t, err)
		restored, err := rotation.Rotate((len(ns) - index) % len(ns))
		require.NoError(t, err)
		assert.True(t, s.Equal(restored))
		reflected, err := s.MirrorIndex(index, len(ns)-1)
		require.NoError(t, err)
		again, err := s.MirrorIndex(reflected.Index, len(ns)-1)
		require.NoError(t, err)
		assert.Equal(t, index, again.Index)
	})
}
