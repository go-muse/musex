package scale_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/scale"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListTransformations(t *testing.T) {
	t.Parallel()
	s := list("G A B")
	for _, tt := range []struct {
		name, want string
		run        func() (*scale.Scale, error)
	}{
		{"rotate", "A B G", func() (*scale.Scale, error) { return s.Rotate(1) }},
		{"rotate no index", "G A B", func() (*scale.Scale, error) { return s.Rotate() }},
		{"reverse", "B A G", s.Reverse},
		{"slice", "A B", func() (*scale.Scale, error) { return s.Slice(1, 3) }},
		{"empty slice", "", func() (*scale.Scale, error) { return s.Slice(3, 3) }},
		{"repeat", "G A B G A B", func() (*scale.Scale, error) { return s.Repeat(2) }},
		{"zero repeat", "", func() (*scale.Scale, error) { return s.Repeat(0) }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.run()
			require.NoError(t, err)
			checkScale(t, names(tt.want), got)
			assert.NotSame(t, s, got)
			require.NoError(t, got.Clear())
			checkScale(t, names("G A B"), s)
		})
	}
	for _, tt := range []struct {
		name string
		run  func() (*scale.Scale, error)
		want error
	}{
		{"rotation index", func() (*scale.Scale, error) { return s.Rotate(3) }, scale.ErrIndex},
		{"too many indices", func() (*scale.Scale, error) { return s.Rotate(0, 1) }, scale.ErrInvalidArgument},
		{"negative repeat", func() (*scale.Scale, error) { return s.Repeat(-1) }, scale.ErrInvalidArgument},
		{"size overflow", func() (*scale.Scale, error) { return s.Repeat(math.MaxInt) }, scale.ErrOutOfRange},
		{"reversed slice", func() (*scale.Scale, error) { return s.Slice(2, 1) }, scale.ErrIndex},
		{"negative slice", func() (*scale.Scale, error) { return s.Slice(-1, 0) }, scale.ErrIndex},
		{"oversize slice", func() (*scale.Scale, error) { return s.Slice(0, 4) }, scale.ErrIndex},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.run()
			assert.Nil(t, got)
			assert.ErrorIs(t, err, tt.want)
			checkScale(t, names("G A B"), s)
		})
	}
	empty := scale.MustNew()
	_, err := empty.Rotate(0)
	assert.ErrorIs(t, err, scale.ErrEmpty)
	checkScale(t, names(""), empty.MustRotate())
	checkScale(t, names(""), empty.MustReverse())
	checkScale(t, names(""), empty.MustSlice(0, 0))
	ns := []note.Note{at("G", 4, cents.Zero(), nil), at("A", 4, cents.Zero(), nil), at("B", 4, cents.Zero(), nil)}
	repeated, err := scale.MustNew(ns...).Repeat(2)
	require.NoError(t, err)
	checkScale(t, append(append([]note.Note{}, ns...), ns...), repeated)
}

func TestOctavesAndTransposition(t *testing.T) {
	t.Parallel()
	a, b := settings.Standard(), settings.Standard().MustWithNumbering(settings.Standard().Numbering())
	original := []note.Note{at("G", 4, cents.MustFromInt(30), a), at("A", 4, cents.MustFromInt(10), b), at("B", 4, cents.Zero(), nil)}
	s := scale.MustNew(original...)
	want := []note.Note{at("A", 4, cents.MustFromInt(30), a), at("B", 4, cents.MustFromInt(10), b), at("C+1", 5, cents.Zero(), nil)}
	got, err := s.TransposeIn(system.Traditional(), interval.MustParse("M2"))
	require.NoError(t, err)
	checkScale(t, want, got)
	own := scale.MustNew(original[:2]...)
	got, err = own.Transpose(interval.MustParse("M2"))
	require.NoError(t, err)
	checkScale(t, want[:2], got)
	target := at("A", 4, cents.MustFromInt(99), settings.Standard())
	got, err = s.TransposeToIn(system.Traditional(), target, interval.Direction(127))
	require.NoError(t, err)
	checkScale(t, want, got)
	got, err = own.TransposeTo(target, interval.Down)
	require.NoError(t, err)
	checkScale(t, want[:2], got)
	got, err = list("G A B").TransposeTo(names("A")[0], interval.Up)
	require.NoError(t, err)
	checkScale(t, names("A B C+1"), got)
	got, err = s.WithoutOctaves()
	require.NoError(t, err)
	for i, n := range got.Notes() {
		assert.False(t, n.HasOctave())
		assert.True(t, n.Spelling().Equal(original[i].Spelling()))
		assert.True(t, n.Cents().Equal(original[i].Cents()))
		checkSource(t, original[i].Source(), n.Source())
	}
	got, err = scale.MustNew(at("G", 4, cents.Zero(), nil), at("G", 5, cents.Zero(), nil)).WithoutOctaves()
	require.NoError(t, err)
	checkScale(t, names("G G"), got)
	checkScale(t, names("G"), got.MustUnique())
	got, err = s.ShiftOctaves(2)
	require.NoError(t, err)
	for i, n := range got.Notes() {
		assert.Equal(t, original[i].MustASPN()+2, n.MustASPN())
		assert.True(t, n.Cents().Equal(original[i].Cents()))
		checkSource(t, original[i].Source(), n.Source())
	}
	for _, tt := range []struct {
		name  string
		run   func() (*scale.Scale, error)
		cause error
	}{
		{"mixed octave shift zero", func() (*scale.Scale, error) { return list("G").ShiftOctaves(0) }, note.ErrNoOctave},
		{"missing system", func() (*scale.Scale, error) {
			return scale.MustNew(at("G", 4, cents.Zero(), nil)).Transpose(interval.MustParse("M2"))
		}, interval.ErrNoSystem},
		{"unset interval", func() (*scale.Scale, error) { return s.Transpose(interval.Interval{}) }, scale.ErrUnset},
		{"empty unset interval", func() (*scale.Scale, error) { return scale.MustNew().Transpose(interval.Interval{}) }, scale.ErrUnset},
		{"empty explicit system", func() (*scale.Scale, error) { return scale.MustNew().TransposeIn(nil, interval.MustParse("M2")) }, scale.ErrUnset},
		{"empty target", func() (*scale.Scale, error) { return scale.MustNew().TransposeTo(target, interval.Up) }, scale.ErrEmpty},
		{"mixed target", func() (*scale.Scale, error) { return s.TransposeToIn(system.Traditional(), names("A")[0], interval.Up) }, interval.ErrNoOctave},
		{"shift overflow", func() (*scale.Scale, error) {
			return scale.MustNew(at("C", math.MaxInt, cents.Zero(), nil)).ShiftOctaves(1)
		}, octave.ErrOutOfRange},
	} {
		t.Run(tt.name, func(t *testing.T) { got, err := tt.run(); assert.Nil(t, got); assert.ErrorIs(t, err, tt.cause) })
	}
	checkScale(t, original, s)
	// A late overflow must not expose an already transposed prefix.
	overflow := scale.MustNew(at("C", 4, cents.Zero(), nil), at("B", math.MaxInt, cents.Zero(), nil))
	got, err = overflow.TransposeIn(system.Traditional(), interval.MustParse("M2"))
	assert.Nil(t, got)
	require.Error(t, err)
	checkScale(t, []note.Note{at("C", 4, cents.Zero(), nil), at("B", math.MaxInt, cents.Zero(), nil)}, overflow)
}

func TestEnharmonicAlternativesAndReplacement(t *testing.T) {
	t.Parallel()
	src := settings.Standard()
	original := at("C+1", 4, cents.MustFromInt(30), src)
	s := scale.MustNew(original, at("G", 4, cents.Zero(), nil))
	for _, explicit := range []bool{false, true} {
		t.Run(map[bool]string{false: "own", true: "explicit"}[explicit], func(t *testing.T) {
			var got *scale.Scale
			var err error
			if explicit {
				got, err = s.WithEnharmonicIn(src, 0, spelling.MustParse("D-1"))
			} else {
				got, err = s.WithEnharmonic(0, spelling.MustParse("D-1"))
			}
			require.NoError(t, err)
			checkScale(t, []note.Note{at("D-1", 4, cents.MustFromInt(30), src), at("G", 4, cents.Zero(), nil)}, got)
			seq, err := s.Enharmonics(0, alteration.MustFromInt(1))
			if explicit {
				seq, err = s.EnharmonicsIn(src, 0, alteration.MustFromInt(1))
			}
			require.NoError(t, err)
			values := collect(seq, 100)
			require.NotEmpty(t, values)
			hasDB := false
			expected := collect(original.MustEnharmonicsIn(src, alteration.MustFromInt(1)), 100)
			checkNotes(t, expected, values)
			checkNotes(t, values, collect(seq, 100))
			for _, n := range values {
				hasDB = hasDB || n.Spelling().Equal(spelling.MustParse("D-1"))
				assert.Equal(t, 4, n.MustASPN())
				assert.True(t, n.Cents().Equal(original.Cents()))
				assert.True(t, settings.SameSource(src, n.Source()))
			}
			assert.True(t, hasDB)
		})
	}
	got, err := s.WithEnharmonic(0, spelling.MustParse("D"))
	assert.Nil(t, got)
	assert.ErrorIs(t, err, scale.ErrNotEquivalent)
	_, err = s.Enharmonics(2, alteration.MustFromInt(1))
	assert.ErrorIs(t, err, scale.ErrIndex)
	_, err = s.EnharmonicsIn(nil, 0, alteration.MustFromInt(1))
	assert.ErrorIs(t, err, scale.ErrUnset)
	_, err = s.Enharmonics(0, alteration.MustFromInt(-1))
	require.Error(t, err)
	_, err = scale.MustNew(at("C+1", 4, cents.Zero(), nil)).WithEnharmonic(0, spelling.MustParse("D-1"))
	assert.ErrorIs(t, err, note.ErrNoSource)
	unbound := list("B+1")
	got, err = unbound.WithEnharmonicIn(src, 0, spelling.MustParse("C"))
	require.NoError(t, err)
	checkScale(t, names("C"), got)
	// Numeric Octave differs from written register: B-sharp at numeric Octave4 is C4.
	boundary := at("B+1", 4, cents.Zero(), nil)
	require.True(t, system.Traditional().MustEquivalent(boundary.MustPitch(), at("C", 4, cents.Zero(), nil).MustPitch()))
	got, err = scale.MustNew(boundary).WithEnharmonicIn(src, 0, spelling.MustParse("C"))
	require.NoError(t, err)
	checkScale(t, []note.Note{at("C", 4, cents.Zero(), nil)}, got)
	written := note.MustFromWrittenASPN(src, spelling.MustParse("B+1"), 4, cents.Zero())
	require.Equal(t, 5, written.MustASPN())
	got, err = scale.MustNew(written).WithEnharmonicIn(src, 0, spelling.MustParse("C"))
	require.NoError(t, err)
	checkScale(t, []note.Note{at("C", 5, cents.Zero(), src)}, got)
	checkScale(t, []note.Note{original, at("G", 4, cents.Zero(), nil)}, s)
}
