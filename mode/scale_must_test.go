package mode_test

import (
	"math"
	"reflect"
	"testing"

	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/scale"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type scaleBuilderSource struct {
	settings  *settings.Settings
	reads     int
	forbidden bool
}

func (s *scaleBuilderSource) Settings() *settings.Settings {
	s.reads++
	if s.forbidden {
		panic("unexpected source read")
	}
	return s.settings
}

func TestAllScaleBuilderMustForms(t *testing.T) {
	t.Parallel()
	m := mode.MustFromCode("ionian")
	tonic := scalePitch("C", 4, settings.Standard())
	options := mode.BuildOptions{Direction: scale.Forward}
	cases := []struct {
		name string
		args []any
	}{
		{"FromTonic", []any{tonic}}, {"FromTonicIn", []any{system.Traditional(), tonic}},
		{"FromTonicCount", []any{tonic, 3, options}}, {"FromTonicCountIn", []any{system.Traditional(), tonic, 3, options}},
		{"FromTonicTo", []any{tonic, scalePitch("C", 5, nil).MustPitch(), options}}, {"FromTonicToIn", []any{system.Traditional(), tonic, scalePitch("C", 5, nil).MustPitch(), options}},
		{"FromTonicOctaves", []any{tonic, 1, options}}, {"FromTonicOctavesIn", []any{system.Traditional(), tonic, 1, options}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			values := make([]reflect.Value, len(tt.args))
			for i, a := range tt.args {
				values[i] = reflect.ValueOf(a)
			}
			ordinary := reflect.ValueOf(m).MethodByName(tt.name)
			require.True(t, ordinary.IsValid())
			assert.Equal(t, reflect.TypeFor[*scale.Scale](), ordinary.Type().Out(0))
			result := ordinary.Call(values)
			require.True(t, result[1].IsNil())
			want := result[0].Interface().(*scale.Scale)
			must := reflect.ValueOf(m).MethodByName("Must" + tt.name)
			require.True(t, must.IsValid())
			got := must.Call(values)[0].Interface().(*scale.Scale)
			assert.True(t, want.Equal(got))
			var empty mode.Mode
			unset := reflect.ValueOf(empty).MethodByName(tt.name).Call(values)
			assert.Nil(t, unset[0].Interface())
			assert.ErrorIs(t, unset[1].Interface().(error), mode.ErrUnset)
			func() {
				defer func() { p := recover(); e, ok := p.(error); require.True(t, ok); assert.ErrorIs(t, e, mode.ErrUnset) }()
				reflect.ValueOf(empty).MethodByName("Must" + tt.name).Call(values)
				t.Fatal("Must must panic")
			}()
		})
	}
	var nilMode *mode.Mode
	assert.Panics(t, func() { _, _ = nilMode.FromTonic(tonic) })
	pointer := &m
	s, err := pointer.FromTonic(tonic)
	require.NoError(t, err)
	assert.Equal(t, 7, s.Len())
	assert.True(t, reflect.TypeFor[mode.Start]().Comparable())
	assert.True(t, reflect.TypeFor[mode.BuildOptions]().Comparable())
	copy := options
	copy.Start.Index = 3
	assert.Zero(t, options.Start.Index)
}

func TestZeroCountIgnoresStartAndSources(t *testing.T) {
	t.Parallel()
	src := &scaleBuilderSource{settings: settings.Standard()}
	tonic := scalePitch("G", 4, src)
	src.forbidden = true
	before := src.reads
	options := mode.BuildOptions{Start: mode.Start{Index: math.MinInt, Cycle: math.MaxInt}, Direction: scale.Forward, StripOctaves: true}
	open := mode.MustParse("m3 M2 |")
	got, err := open.FromTonicCount(tonic, 0, options)
	require.NoError(t, err)
	checkBuilt(t, got, "")
	assert.Equal(t, before, src.reads)
	got, err = open.FromTonicCountIn(system.Traditional(), tonic, 0, options)
	require.NoError(t, err)
	checkBuilt(t, got, "")
	assert.Equal(t, before, src.reads)
	got, err = open.FromTonicCountIn(nil, tonic, 0, options)
	assert.Nil(t, got)
	assert.ErrorIs(t, err, mode.ErrUnset)
	got, err = open.FromTonicCount(note.Note{}, 0, options)
	assert.Nil(t, got)
	assert.ErrorIs(t, err, mode.ErrUnset)
	// No future source calls are hidden by an eagerly created scale or iterator.
	naked := scaleName("G")
	s, err := mode.MustParse("M2 M2 m2").FromTonic(naked)
	require.NoError(t, err)
	require.NoError(t, s.Append(scaleName("C")))
	assert.Equal(t, before, src.reads)
}

// Compile-time signatures guarantee all eight ordinary methods accept exactly the contract types.
var (
	_ func(mode.Mode, note.Note) (*scale.Scale, error)                                                 = mode.Mode.FromTonic
	_ func(mode.Mode, *system.System, note.Note) (*scale.Scale, error)                                 = mode.Mode.FromTonicIn
	_ func(mode.Mode, note.Note, int, mode.BuildOptions) (*scale.Scale, error)                         = mode.Mode.FromTonicCount
	_ func(mode.Mode, *system.System, note.Note, int, mode.BuildOptions) (*scale.Scale, error)         = mode.Mode.FromTonicCountIn
	_ func(mode.Mode, note.Note, pitch.Pitch, mode.BuildOptions) (*scale.Scale, error)                 = mode.Mode.FromTonicTo
	_ func(mode.Mode, *system.System, note.Note, pitch.Pitch, mode.BuildOptions) (*scale.Scale, error) = mode.Mode.FromTonicToIn
	_ func(mode.Mode, note.Note, int, mode.BuildOptions) (*scale.Scale, error)                         = mode.Mode.FromTonicOctaves
	_ func(mode.Mode, *system.System, note.Note, int, mode.BuildOptions) (*scale.Scale, error)         = mode.Mode.FromTonicOctavesIn
	_ func(mode.Mode, note.Note) *scale.Scale                                                          = mode.Mode.MustFromTonic
	_ func(mode.Mode, *system.System, note.Note) *scale.Scale                                          = mode.Mode.MustFromTonicIn
	_ func(mode.Mode, note.Note, int, mode.BuildOptions) *scale.Scale                                  = mode.Mode.MustFromTonicCount
	_ func(mode.Mode, *system.System, note.Note, int, mode.BuildOptions) *scale.Scale                  = mode.Mode.MustFromTonicCountIn
	_ func(mode.Mode, note.Note, pitch.Pitch, mode.BuildOptions) *scale.Scale                          = mode.Mode.MustFromTonicTo
	_ func(mode.Mode, *system.System, note.Note, pitch.Pitch, mode.BuildOptions) *scale.Scale          = mode.Mode.MustFromTonicToIn
	_ func(mode.Mode, note.Note, int, mode.BuildOptions) *scale.Scale                                  = mode.Mode.MustFromTonicOctaves
	_ func(mode.Mode, *system.System, note.Note, int, mode.BuildOptions) *scale.Scale                  = mode.Mode.MustFromTonicOctavesIn
)
