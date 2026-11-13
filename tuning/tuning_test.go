package tuning_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"
	"github.com/go-muse/musex/tuning/temperament"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStandard(t *testing.T) {
	t.Parallel()
	std := tuning.Standard()
	require.NotNil(t, std)
	assert.True(t, std.IsValid())
	assert.True(t, system.Traditional().Equal(std.System()))
	assert.True(t, temperament.Equal12().Equal(std.Temperament()))
	assert.Equal(t, pitchclass.MustNew(0, 12), std.RootClass())
	assert.True(t, tuning.A4At440().Equal(std.Reference()))
	assert.Equal(t, 12, std.Classes())
	assert.Equal(t, 440.0, std.MustFrequency(aspn(spelling.A(), 4)))
	assert.Equal(t, 880.0, std.MustFrequency(aspn(spelling.A(), 5)))

	other := tuning.Standard()
	assert.NotSame(t, std, other)
	assert.True(t, std.Equal(other))
	*other = tuning.Tuning{}
	assert.True(t, tuning.Standard().Equal(std), "overwriting one Standard result does not affect later calls")
}

func TestNewAndReaders(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		root tuning.Root
		want pitchclass.Class
	}{
		{name: "root class", root: tuning.RootFromClass(pitchclass.MustNew(0, 12)), want: pitchclass.MustNew(0, 12)},
		{name: "root name", root: tuning.RootFromName(spelling.A()), want: pitchclass.MustNew(9, 12)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tune := newTestTuning(t, system.Traditional(), temperament.PythagoreanEFlatGSharp(), tc.root, tuning.A4At440())
			assert.True(t, tune.IsValid())
			assert.True(t, system.Traditional().Equal(tune.System()))
			assert.True(t, temperament.PythagoreanEFlatGSharp().Equal(tune.Temperament()))
			assert.Equal(t, tc.want, tune.RootClass())
			assert.True(t, tuning.A4At440().Equal(tune.Reference()))
			assert.Equal(t, 12, tune.Classes())
		})
	}

	tune := newTestTuningAt(t, system.Traditional(), temperament.Equal12(), tuning.RootFromClass(pitchclass.MustNew(0, 12)), aspn(spelling.BFlat(), 4), 466.1638)
	assert.True(t, tune.Reference().Pitch().Equal(aspn(spelling.BFlat(), 4)))
	assert.Equal(t, 466.1638, tune.Reference().Frequency())
}

func TestNewErrorPrecedence(t *testing.T) {
	t.Parallel()
	validRoot := tuning.RootFromClass(pitchclass.MustNew(0, 12))
	validRef := tuning.A4At440()
	halfAltered := aspn(spelling.MustParse("C+1/2"), 4)

	for _, tc := range []struct {
		name string
		call func() (*tuning.Tuning, error)
		want error
	}{
		{
			name: "nil system before every other problem",
			call: func() (*tuning.Tuning, error) {
				return tuning.NewAt(nil, temperament.Temperament{}, tuning.RootFromClass(pitchclass.MustNew(12, 19)), halfAltered, -1)
			},
			want: tuning.ErrUnset,
		},
		{
			name: "zero system",
			call: func() (*tuning.Tuning, error) {
				return tuning.New(&system.System{}, temperament.Equal12(), validRoot, validRef)
			},
			want: tuning.ErrUnset,
		},
		{
			name: "unset temperament",
			call: func() (*tuning.Tuning, error) {
				return tuning.New(system.Traditional(), temperament.Temperament{}, validRoot, validRef)
			},
			want: tuning.ErrUnset,
		},
		{
			name: "unset root",
			call: func() (*tuning.Tuning, error) {
				return tuning.New(system.Traditional(), temperament.Equal12(), tuning.Root{}, validRef)
			},
			want: tuning.ErrUnset,
		},
		{
			name: "unset reference",
			call: func() (*tuning.Tuning, error) {
				return tuning.New(system.Traditional(), temperament.Equal12(), validRoot, tuning.Reference{})
			},
			want: tuning.ErrUnset,
		},
		{
			name: "class count before root modulus",
			call: func() (*tuning.Tuning, error) {
				return tuning.New(system.Positions19(), temperament.Equal12(), tuning.RootFromClass(pitchclass.MustNew(0, 19)), validRef)
			},
			want: tuning.ErrClassCount,
		},
		{
			name: "root class modulus",
			call: func() (*tuning.Tuning, error) {
				return tuning.New(system.Traditional(), temperament.Equal12(), tuning.RootFromClass(pitchclass.MustNew(12, 19)), validRef)
			},
			want: tuning.ErrInvalidArgument,
		},
		{
			name: "root name before reference pitch",
			call: func() (*tuning.Tuning, error) {
				return tuning.New(system.Traditional(), temperament.Equal12(), tuning.RootFromName(spelling.MustParse("C+1/2")), tuning.MustNewReference(halfAltered, 440))
			},
			want: tuning.ErrInadmissible,
		},
		{
			name: "reference pitch",
			call: func() (*tuning.Tuning, error) {
				return tuning.New(system.Traditional(), temperament.Equal12(), validRoot, tuning.MustNewReference(halfAltered, 440))
			},
			want: tuning.ErrInadmissible,
		},
		{
			name: "frequency last in NewAt",
			call: func() (*tuning.Tuning, error) {
				return tuning.NewAt(system.Traditional(), temperament.Equal12(), validRoot, aspn(spelling.A(), 4), -1)
			},
			want: tuning.ErrInvalidArgument,
		},
		{
			name: "K-22 first cause only",
			call: func() (*tuning.Tuning, error) {
				return tuning.NewAt(nil, temperament.Equal12(), tuning.RootFromClass(pitchclass.MustNew(12, 19)), halfAltered, -1)
			},
			want: tuning.ErrUnset,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := tc.call()
			requireTuningError(t, err, tc.want)
			requireNoDependencyError(t, err)
			assert.Nil(t, got)
		})
	}
}

func TestTuningZeroValuesAndReaders(t *testing.T) {
	t.Parallel()
	var nilTuning *tuning.Tuning
	for _, tc := range []struct {
		name string
		tune *tuning.Tuning
	}{
		{name: "nil", tune: nilTuning},
		{name: "zero", tune: &tuning.Tuning{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.False(t, tc.tune.IsValid())
			assert.Nil(t, tc.tune.System())
			assert.False(t, tc.tune.Temperament().IsValid())
			assert.False(t, tc.tune.RootClass().IsValid())
			assert.False(t, tc.tune.Reference().IsValid())
			assert.Zero(t, tc.tune.Classes())
			assert.Equal(t, "Tuning(unset)", tc.tune.String())
			assert.True(t, tc.tune.Equal(nilTuning))
			assert.True(t, tc.tune.Equal(&tuning.Tuning{}))
			assert.False(t, tc.tune.Equal(tuning.Standard()))
			assert.False(t, tuning.Standard().Equal(tc.tune))
		})
	}
}

func TestTuningEqual(t *testing.T) {
	t.Parallel()
	std := tuning.Standard()
	same := newTestTuning(t, system.Traditional(), temperament.Equal12(), tuning.RootFromClass(pitchclass.MustNew(0, 12)), tuning.A4At440())
	assert.True(t, std.Equal(same))

	renamedTemp := temperament.Equal12().MustWithName("renamed equal temperament")
	renamed := newTestTuning(t, system.Traditional(), renamedTemp, tuning.RootFromClass(pitchclass.MustNew(0, 12)), tuning.A4At440())
	assert.True(t, std.Equal(renamed), "temperament names do not affect equality")

	def := system.Traditional().Definition()
	def.Name = "renamed traditional"
	renamedSystem := system.MustNew(def)
	withRenamedSystem := newTestTuning(t, renamedSystem, temperament.Equal12(), tuning.RootFromClass(pitchclass.MustNew(0, 12)), tuning.A4At440())
	assert.True(t, std.Equal(withRenamedSystem), "system names do not affect equality")

	withRoot := std.MustWithRoot(tuning.RootFromClass(pitchclass.MustNew(9, 12)))
	assert.False(t, std.Equal(withRoot))
	withFrequency := std.MustWithReferenceFrequency(math.Nextafter(440, 441))
	assert.False(t, std.Equal(withFrequency))
	withTemperament := std.MustWithTemperament(temperament.PythagoreanEFlatGSharp())
	assert.False(t, std.Equal(withTemperament))
	withSystem := std.MustWithSystem(traditionalSystemWithAlterations(-1, 1, "limited alterations"))
	assert.False(t, std.Equal(withSystem))
}

func TestTuningString(t *testing.T) {
	t.Parallel()
	assert.Equal(t, `Tuning(system "traditional"; temperament "12-EDO"; root 0 mod 12; reference A ASPN 4 = 440 Hz)`, tuning.Standard().String())
	assert.Equal(t, "Tuning(unset)", (&tuning.Tuning{}).String())
}
