package tuning_test

import (
	"math"
	"strconv"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/tuning"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewReference(t *testing.T) {
	t.Parallel()

	t.Run("valid values", func(t *testing.T) {
		t.Parallel()
		for _, tc := range []struct {
			name string
			note pitch.Pitch
			freq float64
		}{
			{name: "A4 at 440", note: aspn(spelling.A(), 4), freq: 440},
			{name: "altered reference", note: aspn(spelling.BFlat(), 4), freq: 466.1638},
			{name: "minimum normal frequency", note: aspn(spelling.C(), 0), freq: cents.MinNormalFrequency},
			{name: "maximum frequency", note: aspn(spelling.C(), 0), freq: cents.MaxFrequency},
		} {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				ref, err := tuning.NewReference(tc.note, tc.freq)
				require.NoError(t, err)
				assert.True(t, ref.IsValid())
				assert.True(t, tc.note.Equal(ref.Pitch()))
				assert.Equal(t, tc.freq, ref.Frequency())
			})
		}
	})

	t.Run("invalid values", func(t *testing.T) {
		t.Parallel()
		for _, tc := range []struct {
			name string
			note pitch.Pitch
			freq float64
			want error
		}{
			{name: "unset pitch", note: pitch.Pitch{}, freq: 440, want: tuning.ErrUnset},
			{name: "zero", note: aspn(spelling.A(), 4), freq: 0, want: tuning.ErrInvalidArgument},
			{name: "negative", note: aspn(spelling.A(), 4), freq: -440, want: tuning.ErrInvalidArgument},
			{name: "subnormal", note: aspn(spelling.A(), 4), freq: math.SmallestNonzeroFloat64, want: tuning.ErrInvalidArgument},
			{name: "NaN", note: aspn(spelling.A(), 4), freq: math.NaN(), want: tuning.ErrInvalidArgument},
			{name: "positive infinity", note: aspn(spelling.A(), 4), freq: math.Inf(1), want: tuning.ErrInvalidArgument},
			{name: "negative infinity", note: aspn(spelling.A(), 4), freq: math.Inf(-1), want: tuning.ErrInvalidArgument},
		} {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				ref, err := tuning.NewReference(tc.note, tc.freq)
				requireTuningError(t, err, tc.want)
				requireNoDependencyError(t, err)
				assert.False(t, ref.IsValid())
				assert.Equal(t, tuning.Reference{}, ref)
			})
		}
	})
}

func TestReferenceZeroValue(t *testing.T) {
	t.Parallel()
	var zero tuning.Reference

	assert.False(t, zero.IsValid())
	assert.False(t, zero.Pitch().IsValid())
	assert.Zero(t, zero.Frequency())
	assert.Equal(t, "Reference(unset)", zero.String())
	assert.True(t, zero.Equal(tuning.Reference{}))
	assert.True(t, zero == tuning.Reference{})
	assert.False(t, zero.Equal(tuning.A4At440()))
	assert.False(t, tuning.A4At440().Equal(zero))
}

func TestReferenceEqualAndComparable(t *testing.T) {
	t.Parallel()
	base := tuning.A4At440()
	same := tuning.MustNewReference(aspn(spelling.A(), 4), 440)
	differentNote := tuning.MustNewReference(aspn(spelling.BFlat(), 4), 440)
	differentFrequency := tuning.MustNewReference(aspn(spelling.A(), 4), math.Nextafter(440, 441))

	assert.True(t, base.Equal(same))
	assert.True(t, base == same)
	assert.False(t, base.Equal(differentNote))
	assert.False(t, base == differentNote)
	assert.False(t, base.Equal(differentFrequency))
	assert.False(t, base == differentFrequency)

	byReference := map[tuning.Reference]string{base: "standard"}
	assert.Equal(t, "standard", byReference[same])
}

func TestReferenceString(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		ref  tuning.Reference
		want string
	}{
		{name: "integer frequency", ref: tuning.A4At440(), want: "A ASPN 4 = 440 Hz"},
		{name: "decimal frequency", ref: tuning.MustNewReference(aspn(spelling.C(), 5), 523.25), want: "C ASPN 5 = 523.25 Hz"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, tc.ref.String())
			assert.Equal(t, tc.ref.Pitch().String()+" = "+strconv.FormatFloat(tc.ref.Frequency(), 'g', -1, 64)+" Hz", tc.ref.String())
		})
	}
}

func TestReferenceWithFrequency(t *testing.T) {
	t.Parallel()
	base := tuning.A4At440()
	changed, err := base.WithFrequency(435)
	require.NoError(t, err)
	assert.Equal(t, 435.0, changed.Frequency())
	assert.True(t, base.Pitch().Equal(changed.Pitch()))
	assert.Equal(t, 440.0, base.Frequency(), "the input reference is immutable")

	t.Run("extreme replacement does not calculate a ratio", func(t *testing.T) {
		t.Parallel()
		ref := tuning.MustNewReference(aspn(spelling.A(), 4), cents.MinNormalFrequency)
		changed, err := ref.WithFrequency(cents.MaxFrequency)
		require.NoError(t, err)
		assert.Equal(t, cents.MaxFrequency, changed.Frequency())
	})

	t.Run("errors", func(t *testing.T) {
		t.Parallel()
		var zero tuning.Reference
		_, err := zero.WithFrequency(440)
		requireTuningError(t, err, tuning.ErrUnset)
		_, err = base.WithFrequency(0)
		requireTuningError(t, err, tuning.ErrInvalidArgument)
		assert.Equal(t, 440.0, base.Frequency())
	})
}

func TestReferenceFrequencyRatio(t *testing.T) {
	t.Parallel()
	ratio, err := tuning.A4At440().FrequencyRatio(435)
	require.NoError(t, err)
	assert.Equal(t, 435.0/440.0, ratio)
	assert.Equal(t, 0.988636, math.Round(ratio*1e6)/1e6)

	t.Run("overflow", func(t *testing.T) {
		t.Parallel()
		ref := tuning.MustNewReference(aspn(spelling.A(), 4), cents.MinNormalFrequency)
		_, err := ref.FrequencyRatio(cents.MaxFrequency)
		requireTuningError(t, err, tuning.ErrOutOfRange)
		assert.ErrorIs(t, err, cents.ErrOutOfRange)
		assert.NotErrorIs(t, err, tuning.ErrInvalidArgument)
	})

	t.Run("underflow", func(t *testing.T) {
		t.Parallel()
		ref := tuning.MustNewReference(aspn(spelling.A(), 4), cents.MaxFrequency)
		_, err := ref.FrequencyRatio(cents.MinNormalFrequency)
		requireTuningError(t, err, tuning.ErrOutOfRange)
		assert.ErrorIs(t, err, cents.ErrOutOfRange)
	})

	t.Run("input errors", func(t *testing.T) {
		t.Parallel()
		var zero tuning.Reference
		_, err := zero.FrequencyRatio(440)
		requireTuningError(t, err, tuning.ErrUnset)
		_, err = tuning.A4At440().FrequencyRatio(0)
		requireTuningError(t, err, tuning.ErrInvalidArgument)
	})
}

func TestReferenceCentsTo(t *testing.T) {
	t.Parallel()
	got, err := tuning.A4At440().CentsTo(tuning.A4At435())
	require.NoError(t, err)
	requireCentsEqual(t, cents.MustBetween(440, 435), got)
	assert.Equal(t, "-19.785747", got.Decimal(6, 0), "K-25")

	t.Run("different notes are incomparable", func(t *testing.T) {
		t.Parallel()
		_, err := tuning.A4At440().CentsTo(tuning.C5At523())
		requireTuningError(t, err, tuning.ErrIncomparable)
		requireNoDependencyError(t, err)
	})

	t.Run("unset", func(t *testing.T) {
		t.Parallel()
		_, err := tuning.Reference{}.CentsTo(tuning.A4At440())
		requireTuningError(t, err, tuning.ErrUnset)
		_, err = tuning.A4At440().CentsTo(tuning.Reference{})
		requireTuningError(t, err, tuning.ErrUnset)
	})
}

func TestReadyReferences(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		make func() tuning.Reference
		note pitch.Pitch
		freq float64
	}{
		{name: "A4 at 440", make: tuning.A4At440, note: aspn(spelling.A(), 4), freq: 440},
		{name: "A4 at 435", make: tuning.A4At435, note: aspn(spelling.A(), 4), freq: 435},
		{name: "A4 at 442", make: tuning.A4At442, note: aspn(spelling.A(), 4), freq: 442},
		{name: "A4 at 444", make: tuning.A4At444, note: aspn(spelling.A(), 4), freq: 444},
		{name: "A4 at 453", make: tuning.A4At453, note: aspn(spelling.A(), 4), freq: 453},
		{name: "A4 at 466", make: tuning.A4At466, note: aspn(spelling.A(), 4), freq: 466},
		{name: "C5 at 523", make: tuning.C5At523, note: aspn(spelling.C(), 5), freq: 523},
		{name: "C0 at 16", make: tuning.C0At16, note: aspn(spelling.C(), 0), freq: 16},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ref := tc.make()
			require.True(t, ref.IsValid())
			assert.True(t, tc.note.Equal(ref.Pitch()))
			assert.Equal(t, tc.freq, ref.Frequency())
		})
	}

	refs := tuning.References()
	require.Len(t, refs, 8)
	for i, makeRef := range []func() tuning.Reference{
		tuning.A4At440,
		tuning.A4At435,
		tuning.A4At442,
		tuning.A4At444,
		tuning.A4At453,
		tuning.A4At466,
		tuning.C5At523,
		tuning.C0At16,
	} {
		assert.True(t, makeRef().Equal(refs[i]), "reference %d is out of catalog order", i)
	}

	refs[0] = tuning.Reference{}
	refs[7] = tuning.A4At440()
	fresh := tuning.References()
	require.Len(t, fresh, 8)
	assert.True(t, tuning.A4At440().Equal(fresh[0]))
	assert.True(t, tuning.C0At16().Equal(fresh[7]))
}
