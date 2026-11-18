package mode_test

import (
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/scale"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func scaleName(name string) note.Note {
	return note.MustFromSpelling(spelling.MustParse(name), cents.Zero(), nil)
}
func scalePitch(name string, aspn int, src settings.Source) note.Note {
	return note.MustFromASPN(spelling.MustParse(name), aspn, cents.Zero(), src)
}
func checkBuilt(t *testing.T, s *scale.Scale, namesText string, aspns ...int) {
	t.Helper()
	require.NotNil(t, s)
	want := strings.Fields(namesText)
	require.Len(t, s.Notes(), len(want))
	for i, n := range s.Notes() {
		assert.True(t, spelling.MustParse(want[i]).Equal(n.Spelling()), "note %d: %s", i, n)
		if len(aspns) == 0 {
			assert.False(t, n.HasOctave())
		} else {
			assert.Equal(t, aspns[i], n.MustASPN())
		}
	}
}

// O-2, K-1/4/40/45 and compatibility of released note and degree APIs.
func TestFromTonicReturnsIndependentScale(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, structure, first, want string }{{"dorian", "M2 m2 M2 M2 M2 m2 M2", "D", "D E F G A B C"}, {"ionian", "M2 M2 m2 M2 M2 M2 m2", "A", "A B C+1 D E F+1 G+1"}, {"open", "m3 M2 |", "E", "E G A"}, {"quartal cycle", "M2 M2 m2", "G", "G A B"}, {"one note", "P8", "C", "C"}} {
		t.Run(tt.name, func(t *testing.T) {
			m := mode.MustParse(tt.structure)
			tonic := scaleName(tt.first)
			s, err := m.FromTonic(tonic)
			require.NoError(t, err)
			checkBuilt(t, s, tt.want)
			old, err := m.NotesFrom(tonic)
			require.NoError(t, err)
			require.Len(t, old, s.Len())
			for i, n := range old {
				assert.True(t, n.Equal(s.Notes()[i]))
			}
			ds, err := m.DegreesFrom(tonic)
			require.NoError(t, err)
			require.Len(t, ds, s.Len())
			one, err := m.NoteFrom(tonic, 1, 0)
			require.NoError(t, err)
			assert.True(t, one.Equal(s.Notes()[0]))
			require.NoError(t, s.Set(0, scaleName("F")))
			unchanged, err := m.FromTonic(tonic)
			require.NoError(t, err)
			checkBuilt(t, unchanged, tt.want)
			require.NoError(t, m.RebuildFromCode("dorian"))
			checkBuilt(t, unchanged, tt.want)
		})
	}
	m := mode.MustFromCode("ionian")
	src := settings.Standard()
	tonic := scalePitch("A", 4, src)
	tonic.MustSetCents(cents.MustFromInt(30))
	built, err := m.FromTonic(tonic)
	require.NoError(t, err)
	checkBuilt(t, built, "A B C+1 D E F+1 G+1", 4, 4, 5, 5, 5, 5, 5)
	explicit, err := m.FromTonicIn(system.Traditional(), tonic)
	require.NoError(t, err)
	assert.True(t, built.Equal(explicit))
	for _, n := range built.Notes() {
		assert.True(t, n.Cents().Equal(tonic.Cents()))
		assert.True(t, settings.SameSource(src, n.Source()))
	}
	// Both exported names are the same type, including assignments without conversion.
	sig := mode.Signature{Index: 2, Signs: []spelling.Spelling{spelling.MustParse("F+1"), spelling.MustParse("C+1")}}
	shared := toScaleSignature(sig)
	back := toModeSignature(shared)
	assert.Equal(t, sig.Index, back.Index)
	assert.Equal(t, reflect.TypeFor[scale.Signature](), reflect.TypeFor[mode.Signature]())
	assert.False(t, reflect.TypeFor[mode.Signature]().Comparable())
}

func TestFiniteCountAndRangeBuilders(t *testing.T) {
	t.Parallel()
	m := mode.MustParse("M2 M2 m2")
	tonic := scalePitch("G", 4, nil)
	forward := mode.BuildOptions{Direction: scale.Forward}
	got, err := m.FromTonicCountIn(system.Traditional(), tonic, 7, forward)
	require.NoError(t, err)
	checkBuilt(t, got, "G A B C D E F", 4, 4, 4, 5, 5, 5, 5)
	next, err := got.Next(6)
	require.NoError(t, err)
	assert.True(t, tonic.Equal(next.Note))
	got, err = m.FromTonicOctavesIn(system.Traditional(), tonic, 1, forward)
	require.NoError(t, err)
	checkBuilt(t, got, "G A B C D E F G", 4, 4, 4, 5, 5, 5, 5, 5)
	round, err := got.OneRound(0, scale.Forward)
	require.NoError(t, err)
	assert.True(t, got.Equal(round))
	assert.Equal(t, 8, round.Len())
	wrapped, err := got.Next(7)
	require.NoError(t, err)
	assert.True(t, tonic.Equal(wrapped.Note))
	stripped := forward
	stripped.StripOctaves = true
	got, err = m.FromTonicOctavesIn(system.Traditional(), tonic, 1, stripped)
	require.NoError(t, err)
	checkBuilt(t, got, "G A B C D E F G")
	got, err = m.FromTonicToIn(system.Traditional(), tonic, scalePitch("F+1", 5, nil).MustPitch(), forward)
	require.NoError(t, err)
	checkBuilt(t, got, "G A B C D E F", 4, 4, 4, 5, 5, 5, 5)
	got, err = m.FromTonicToIn(system.Traditional(), tonic, tonic.MustPitch(), forward)
	require.NoError(t, err)
	checkBuilt(t, got, "G", 4)
	start := mode.BuildOptions{Start: mode.Start{Index: 1, Cycle: -1}, Direction: scale.Forward}
	got, err = m.FromTonicCountIn(system.Traditional(), tonic, 4, start)
	require.NoError(t, err)
	checkBuilt(t, got, "E F+1 G A", 4, 4, 4, 4)
	backward := mode.BuildOptions{Direction: scale.Backward}
	got, err = m.FromTonicCountIn(system.Traditional(), tonic, 4, backward)
	require.NoError(t, err)
	checkBuilt(t, got, "G F+1 E D", 4, 4, 4, 4)
	// An octave is measured from the selected first note, not from the tonic.
	selected := mode.BuildOptions{Start: mode.Start{Index: 1}, Direction: scale.Forward}
	got, err = mode.MustFromCode("ionian").FromTonicOctavesIn(system.Traditional(), scalePitch("C", 4, nil), 1, selected)
	require.NoError(t, err)
	checkBuilt(t, got, "D E F G A B C D", 4, 4, 4, 4, 4, 4, 5, 5)
	got, err = mode.MustFromCode("ionian").FromTonicOctavesIn(system.Traditional(), scalePitch("C", 5, nil), 1, backward)
	require.NoError(t, err)
	checkBuilt(t, got, "C B A G F E D C", 5, 4, 4, 4, 4, 4, 4, 4)
	open := mode.MustParse("m3 M2 |")
	e := scalePitch("E", 4, nil)
	last := mode.BuildOptions{Start: mode.Start{Index: 2}, Direction: scale.Backward}
	got, err = open.FromTonicCountIn(system.Traditional(), e, 3, last)
	require.NoError(t, err)
	checkBuilt(t, got, "A G E", 4, 4, 4)
	got, err = open.FromTonicCountIn(system.Traditional(), e, 4, last)
	assert.Nil(t, got)
	assert.ErrorIs(t, err, mode.ErrInsufficientNotes)
	got, err = open.FromTonicToIn(system.Traditional(), e, scalePitch("E", 5, nil).MustPitch(), forward)
	require.NoError(t, err)
	checkBuilt(t, got, "E G A", 4, 4, 4)
	got, err = open.FromTonicToIn(system.Traditional(), e, e.MustPitch(), last)
	require.NoError(t, err)
	checkBuilt(t, got, "A G E", 4, 4, 4)
	got, err = open.FromTonicCount(scaleName("E"), 3, mode.BuildOptions{Direction: scale.Forward})
	require.NoError(t, err)
	checkBuilt(t, got, "E G A")
	// Own-context variants agree with explicit ones without changing the borrowed source.
	bound := scalePitch("G", 4, settings.Standard())
	own, err := m.FromTonicCount(bound, 7, forward)
	require.NoError(t, err)
	in, err := m.FromTonicCountIn(system.Traditional(), bound, 7, forward)
	require.NoError(t, err)
	assert.True(t, own.Equal(in))
	own, err = m.FromTonicTo(bound, scalePitch("G", 5, nil).MustPitch(), forward)
	require.NoError(t, err)
	in, err = m.FromTonicToIn(system.Traditional(), bound, scalePitch("G", 5, nil).MustPitch(), forward)
	require.NoError(t, err)
	assert.True(t, own.Equal(in))
	own, err = m.FromTonicOctaves(bound, 2, forward)
	require.NoError(t, err)
	in, err = m.FromTonicOctavesIn(system.Traditional(), bound, 2, forward)
	require.NoError(t, err)
	assert.True(t, own.Equal(in))
}

func TestRangeChecksAllLetterPhasesAndOpenRemainder(t *testing.T) {
	t.Parallel()
	irregular := system.MustFromSteps(letter.C, [letter.Count]int{3, 1, 1, 2, 2, 2, 1}, 1, system.Integers(), "scale phase fixture")
	tonic := scalePitch("C", 4, nil)
	options := mode.BuildOptions{Direction: scale.Forward}
	closed := mode.MustParse("d2")
	// Directed d2 moves one letter and subtracts two alterations. Positions are 0,1,0:
	// the first cycle advances but the second letter phase returns to the boundary.
	counted, err := closed.FromTonicCountIn(irregular, tonic, 3, options)
	require.NoError(t, err)
	checkBuilt(t, counted, "C D-2 E-4", 4, 4, 4)
	ps := counted.Notes()
	assert.Less(t, irregular.MustCompare(ps[0].MustPitch(), ps[1].MustPitch()), 0)
	assert.Greater(t, irregular.MustCompare(ps[1].MustPitch(), ps[2].MustPitch()), 0)
	got, err := closed.FromTonicToIn(irregular, tonic, tonic.MustPitch(), options)
	assert.Nil(t, got)
	assert.ErrorIs(t, err, mode.ErrNonMonotone)
	got, err = closed.FromTonicOctavesIn(irregular, tonic, 1, options)
	assert.Nil(t, got)
	assert.ErrorIs(t, err, mode.ErrNonMonotone)
	// The finite open remainder must be filtered in full, including re-entry.
	got, err = mode.MustParse("d2 d2 |").FromTonicToIn(irregular, tonic, tonic.MustPitch(), options)
	require.NoError(t, err)
	checkBuilt(t, got, "C E-4", 4, 4)
	got, err = mode.MustParse("d2 d2 |").FromTonicToIn(irregular, tonic, tonic.MustPitch(), mode.BuildOptions{Start: mode.Start{Index: 2}, Direction: scale.Backward})
	require.NoError(t, err)
	checkBuilt(t, got, "E-4 C", 4, 4)
	// A zero positional cycle fails both range directions but is legal for count.
	for _, direction := range []scale.Direction{scale.Forward, scale.Backward} {
		got, err = closed.FromTonicToIn(system.Traditional(), tonic, tonic.MustPitch(), mode.BuildOptions{Direction: direction})
		assert.Nil(t, got)
		assert.ErrorIs(t, err, mode.ErrNonMonotone)
	}
}

func TestBuilderRepresentabilityAndErrorOrder(t *testing.T) {
	t.Parallel()
	m := mode.MustParse("P8")
	forward := mode.BuildOptions{Direction: scale.Forward}
	tonic := scalePitch("C", 4, nil)
	for _, tt := range []struct {
		name  string
		run   func() (*scale.Scale, error)
		want  error
		cause error
	}{
		{"zero mode first", func() (*scale.Scale, error) { return mode.Mode{}.FromTonic(note.Note{}) }, mode.ErrUnset, nil},
		{"unset tonic", func() (*scale.Scale, error) { return m.FromTonic(note.Note{}) }, mode.ErrUnset, note.ErrUnset},
		{"unset explicit system", func() (*scale.Scale, error) { return m.FromTonicIn(nil, tonic) }, mode.ErrUnset, system.ErrInvalid},
		{"negative count", func() (*scale.Scale, error) { return m.FromTonicCount(tonic, -1, forward) }, mode.ErrInvalidArgument, nil},
		{"unset direction", func() (*scale.Scale, error) { return m.FromTonicCount(tonic, 0, mode.BuildOptions{}) }, mode.ErrInvalidArgument, nil},
		{"unknown direction", func() (*scale.Scale, error) { return m.FromTonicCount(tonic, 1, mode.BuildOptions{Direction: 2}) }, mode.ErrInvalidArgument, nil},
		{"negative start", func() (*scale.Scale, error) {
			return m.FromTonicCount(tonic, 1, mode.BuildOptions{Start: mode.Start{Index: -1}, Direction: scale.Forward})
		}, mode.ErrIndex, nil},
		{"past start", func() (*scale.Scale, error) {
			return m.FromTonicCount(tonic, 1, mode.BuildOptions{Start: mode.Start{Index: 1}, Direction: scale.Forward})
		}, mode.ErrIndex, nil},
		{"open cycle", func() (*scale.Scale, error) {
			return mode.MustParse("M2 |").FromTonicCount(scaleName("C"), 1, mode.BuildOptions{Start: mode.Start{Cycle: 1}, Direction: scale.Forward})
		}, mode.ErrInvalidArgument, nil},
		{"no system", func() (*scale.Scale, error) { return m.FromTonic(tonic) }, mode.ErrNoSystem, interval.ErrNoSystem},
		{"no octave", func() (*scale.Scale, error) {
			return m.FromTonicOctavesIn(system.Traditional(), scaleName("C"), 1, forward)
		}, nil, note.ErrNoOctave},
		{"zero octaves", func() (*scale.Scale, error) { return m.FromTonicOctaves(tonic, 0, forward) }, mode.ErrInvalidArgument, nil},
		{"negative octaves", func() (*scale.Scale, error) { return m.FromTonicOctaves(tonic, -1, forward) }, mode.ErrInvalidArgument, nil},
		{"unset end", func() (*scale.Scale, error) {
			return m.FromTonicToIn(system.Traditional(), tonic, pitch.Pitch{}, forward)
		}, mode.ErrUnset, pitch.ErrInvalid},
		{"against direction", func() (*scale.Scale, error) {
			return m.FromTonicToIn(system.Traditional(), scalePitch("G", 4, nil), scalePitch("F", 4, nil).MustPitch(), forward)
		}, mode.ErrInvalidArgument, nil},
		{"end octave overflow", func() (*scale.Scale, error) {
			return m.FromTonicOctavesIn(system.Traditional(), scalePitch("C", math.MaxInt, nil), 1, forward)
		}, mode.ErrCannotBuildNote, octave.ErrOutOfRange},
		{"inside note overflow", func() (*scale.Scale, error) {
			return m.FromTonicCountIn(system.Traditional(), scalePitch("C", math.MaxInt, nil), 2, forward)
		}, mode.ErrCannotBuildNote, nil},
		{"unrepresentable result length", func() (*scale.Scale, error) {
			return m.FromTonicOctavesIn(system.Traditional(), scalePitch("C", 0, nil), math.MaxInt, forward)
		}, mode.ErrOutOfRange, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.run()
			assert.Nil(t, got)
			require.Error(t, err)
			if tt.want != nil {
				assert.ErrorIs(t, err, tt.want)
			}
			if tt.cause != nil {
				assert.ErrorIs(t, err, tt.cause)
			}
		})
	}
	// The next mathematical note is outside the requested range and outside int octaves.
	high := scalePitch("C", math.MaxInt, nil)
	got, err := m.FromTonicToIn(system.Traditional(), high, high.MustPitch(), forward)
	require.NoError(t, err)
	checkBuilt(t, got, "C", math.MaxInt)
	low := scalePitch("C", math.MinInt, nil)
	got, err = m.FromTonicToIn(system.Traditional(), low, low.MustPitch(), mode.BuildOptions{Direction: scale.Backward})
	require.NoError(t, err)
	checkBuilt(t, got, "C", math.MinInt)
	got, err = m.FromTonicCountIn(system.Traditional(), high, 2, mode.BuildOptions{Start: mode.Start{Cycle: math.MinInt}, Direction: scale.Forward})
	require.NoError(t, err)
	checkBuilt(t, got, "C C", -1, 0)
	magnitude := alteration.MaxMagnitude
	unison := mode.MustClosedFromIntervals([]interval.Interval{interval.MustFromDeviation(1, alteration.MustFromInt(magnitude))})
	start := note.MustFromSpelling(spelling.MustNew(letter.C, alteration.MustFromInt(-magnitude)), cents.Zero(), nil)
	got, err = unison.FromTonicCount(start, 1, mode.BuildOptions{Start: mode.Start{Cycle: 2}, Direction: scale.Forward})
	require.NoError(t, err)
	assert.True(t, alteration.MustFromInt(magnitude).Equal(got.Notes()[0].Alteration()))
	got, err = unison.FromTonicCount(start, 2, mode.BuildOptions{Start: mode.Start{Cycle: 2}, Direction: scale.Forward})
	assert.Nil(t, got)
	assert.ErrorIs(t, err, mode.ErrCannotBuildNote)
	assert.ErrorIs(t, err, alteration.ErrOutOfRange)
	assert.Equal(t, "mode: index out of range", mode.ErrIndex.Error())
	assert.Equal(t, "mode: insufficient mode notes", mode.ErrInsufficientNotes.Error())
	assert.Equal(t, "mode: non-monotone mode expansion", mode.ErrNonMonotone.Error())
}

func toScaleSignature(sig mode.Signature) scale.Signature { return sig }
func toModeSignature(sig scale.Signature) mode.Signature  { return sig }

func TestRangeDoesNotBuildInadmissibleOutsideNote(t *testing.T) {
	t.Parallel()
	natural := system.MustFromSteps(letter.C, [letter.Count]int{2, 2, 1, 2, 2, 2, 1}, 1, system.Single(alteration.Natural()), "naturals only")
	tonic := scalePitch("C", 4, nil)
	options := mode.BuildOptions{Direction: scale.Forward}
	m := mode.MustParse("A1")
	got, err := m.FromTonicToIn(natural, tonic, tonic.MustPitch(), options)
	require.NoError(t, err)
	checkBuilt(t, got, "C", 4)
	got, err = m.FromTonicToIn(natural, tonic, scalePitch("D", 4, nil).MustPitch(), options)
	assert.Nil(t, got)
	assert.ErrorIs(t, err, mode.ErrCannotBuildNote)
	assert.ErrorIs(t, err, system.ErrInadmissible)
}
