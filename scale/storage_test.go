package scale_test

import (
	"reflect"
	"sync"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/scale"
	"github.com/go-muse/musex/settings"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// O-1/4/27, G-1...G-7; K-12/34/36/38/40.
func TestStorageZeroNilAndOwnership(t *testing.T) {
	t.Parallel()
	var absent *scale.Scale
	assert.False(t, absent.IsValid())
	assert.False(t, absent.IsEmpty())
	assert.Zero(t, absent.Len())
	assert.Nil(t, absent.Notes())
	assert.Nil(t, absent.Clone())
	assert.False(t, reflect.TypeFor[scale.Scale]().Comparable())
	assert.False(t, reflect.TypeFor[scale.Signature]().Comparable())
	for _, s := range []*scale.Scale{new(scale.Scale), scale.MustNew(), scale.MustNew([]note.Note(nil)...)} {
		assert.True(t, s.IsValid())
		assert.True(t, s.IsEmpty())
		assert.Zero(t, s.Len())
		require.NotNil(t, s.Notes())
		require.NotNil(t, s.Clone())
		_, err := s.First()
		assert.ErrorIs(t, err, scale.ErrEmpty)
		_, err = s.Last()
		assert.ErrorIs(t, err, scale.ErrEmpty)
		_, err = s.At(0)
		assert.ErrorIs(t, err, scale.ErrIndex)
	}
	src := &source{set: settings.Standard()}
	g := at("G", 4, cents.MustFromInt(30), src)
	src.forbidden = true
	input := []note.Note{g, names("A")[0], g}
	s, err := scale.New(input...)
	require.NoError(t, err)
	saved := s.Notes()
	input[0] = names("F")[0]
	output := s.Notes()
	output[0].MustSetCents(cents.Zero())
	checkScale(t, saved, s)
	cloned := s.Clone()
	valueCopy := *s
	require.NoError(t, cloned.Set(0, names("F")[0]))
	require.NoError(t, valueCopy.Clear())
	checkScale(t, saved, s)
	pointerCopy := s
	require.NoError(t, pointerCopy.Clear())
	assert.True(t, s.IsEmpty())
	assert.Equal(t, int64(1), src.calls.Load()) // Only the note constructor reads its source.
	bad, err := scale.New(g, note.Note{})
	require.Nil(t, bad)
	assert.ErrorIs(t, err, scale.ErrUnset)
	assert.ErrorIs(t, err, note.ErrUnset)
}

func TestStorageMutationsAndAtomicFailures(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		edit func(*scale.Scale) error
		want string
	}{
		{"replace", func(s *scale.Scale) error { return s.Replace(names("C D")...) }, "C D"},
		{"set", func(s *scale.Scale) error { return s.Set(1, names("C")[0]) }, "G C B"},
		{"insert first", func(s *scale.Scale) error { return s.Insert(0, names("C D")...) }, "C D G A B"},
		{"insert last", func(s *scale.Scale) error { return s.Insert(3, names("C")...) }, "G A B C"},
		{"append", func(s *scale.Scale) error { return s.Append(names("C D")...) }, "G A B C D"},
		{"remove", func(s *scale.Scale) error { return s.Remove(1) }, "G B"},
		{"delete half open", func(s *scale.Scale) error { return s.Delete(1, 3) }, "G"},
		{"delete empty", func(s *scale.Scale) error { return s.Delete(3, 3) }, "G A B"},
		{"insert empty", func(s *scale.Scale) error { return s.Insert(2) }, "G A B"},
		{"clear", func(s *scale.Scale) error { return s.Clear() }, ""},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s := list("G A B")
			copy := *s
			require.NoError(t, tt.edit(s))
			checkScale(t, names(tt.want), s)
			checkScale(t, names("G A B"), &copy)
		})
	}
	failures := []struct {
		name string
		edit func(*scale.Scale) error
		want error
	}{
		{"replace invalid", func(s *scale.Scale) error { return s.Replace(names("C")[0], note.Note{}) }, scale.ErrUnset},
		{"set invalid", func(s *scale.Scale) error { return s.Set(1, note.Note{}) }, scale.ErrUnset},
		{"set index", func(s *scale.Scale) error { return s.Set(3, names("C")[0]) }, scale.ErrIndex},
		{"insert invalid", func(s *scale.Scale) error { return s.Insert(1, names("C")[0], note.Note{}) }, scale.ErrUnset},
		{"append invalid", func(s *scale.Scale) error { return s.Append(names("C")[0], note.Note{}) }, scale.ErrUnset},
		{"insert negative", func(s *scale.Scale) error { return s.Insert(-1) }, scale.ErrIndex},
		{"insert beyond", func(s *scale.Scale) error { return s.Insert(4) }, scale.ErrIndex},
		{"remove negative", func(s *scale.Scale) error { return s.Remove(-1) }, scale.ErrIndex},
		{"remove beyond", func(s *scale.Scale) error { return s.Remove(3) }, scale.ErrIndex},
		{"delete reverse", func(s *scale.Scale) error { return s.Delete(2, 1) }, scale.ErrIndex},
		{"delete negative", func(s *scale.Scale) error { return s.Delete(-1, 1) }, scale.ErrIndex},
		{"delete beyond", func(s *scale.Scale) error { return s.Delete(0, 4) }, scale.ErrIndex},
	}
	for _, tt := range failures {
		t.Run(tt.name, func(t *testing.T) {
			s := list("G A B")
			err := tt.edit(s)
			assert.ErrorIs(t, err, tt.want)
			checkScale(t, names("G A B"), s)
		})
	}
	s := list("G A B")
	first, err := s.First()
	require.NoError(t, err)
	assert.True(t, first.Equal(names("G")[0]))
	last, err := s.Last()
	require.NoError(t, err)
	assert.True(t, last.Equal(names("B")[0]))
	for _, i := range []int{-1, 3} {
		n, err := s.At(i)
		assert.ErrorIs(t, err, scale.ErrIndex)
		assert.False(t, n.IsValid())
	}
	empty := scale.MustNew()
	require.NoError(t, empty.Insert(0, names("G")...))
	require.NoError(t, empty.Delete(0, 1))
	require.NoError(t, empty.Delete(0, 0))
}

func TestLargeListAndConcurrentIndependentCopies(t *testing.T) {
	t.Parallel()
	ns := make([]note.Note, 1024)
	palette := names("G A B")
	for i := range ns {
		ns[i] = palette[i%3]
	}
	s := scale.MustNew(ns...)
	require.Equal(t, 1024, s.Len())
	checkScale(t, ns, s)
	a, b := *s, *s
	done := make(chan error, 2)
	var wg sync.WaitGroup
	wg.Go(func() { done <- a.Set(0, names("C")[0]) })
	wg.Go(func() { done <- b.Set(0, names("D")[0]) })
	wg.Wait()
	for range 2 {
		require.NoError(t, <-done)
	}
	checkScale(t, ns, s)
	assert.True(t, a.MustAt(0).Equal(names("C")[0]))
	assert.True(t, b.MustAt(0).Equal(names("D")[0]))
}

func TestMutationInputsAreCopied(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, initial, want string
		mutate              func(*scale.Scale, []note.Note) error
	}{
		{"replace", "G A B", "C D", func(s *scale.Scale, ns []note.Note) error { return s.Replace(ns...) }},
		{"insert", "G A B", "G C D A B", func(s *scale.Scale, ns []note.Note) error { return s.Insert(1, ns...) }},
		{"append", "G A B", "G A B C D", func(s *scale.Scale, ns []note.Note) error { return s.Append(ns...) }},
		{"insert into empty", "", "C D", func(s *scale.Scale, ns []note.Note) error { return s.Insert(0, ns...) }},
		{"append into empty", "", "C D", func(s *scale.Scale, ns []note.Note) error { return s.Append(ns...) }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s := list(tt.initial)
			saved := *s
			input := names("C D")
			require.NoError(t, tt.mutate(s, input))
			input[0] = names("F")[0]
			input[1].MustSetCents(cents.MustFromInt(40))
			checkScale(t, names(tt.want), s)
			checkScale(t, names(tt.initial), &saved)
		})
	}
}
