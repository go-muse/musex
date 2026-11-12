package frequency_test

import (
	"math"
	"reflect"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/frequency"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCriterionCreation_InvalidTolerance(t *testing.T) {
	tests := []struct {
		name string
		call func() (frequency.Criterion, error)
	}{
		{
			"K-23 negative cents",
			func() (frequency.Criterion, error) { return frequency.WithinCents(cents.MustFromInt(-1)) },
		},
		{"K-23 negative hertz", func() (frequency.Criterion, error) { return frequency.WithinHertz(-0.5) }},
		{"K-23 nan hertz", func() (frequency.Criterion, error) { return frequency.WithinHertz(math.NaN()) }},
		{"K-23 infinite hertz", func() (frequency.Criterion, error) { return frequency.WithinHertz(math.Inf(1)) }},
		{"K-23 negative relative", func() (frequency.Criterion, error) { return frequency.WithinRelative(-0.01) }},
		{"K-23 infinite relative", func() (frequency.Criterion, error) { return frequency.WithinRelative(math.Inf(1)) }},
		{"negative infinite hertz", func() (frequency.Criterion, error) { return frequency.WithinHertz(math.Inf(-1)) }},
		{"negative infinite relative", func() (frequency.Criterion, error) { return frequency.WithinRelative(math.Inf(-1)) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			criterion, err := tt.call()
			assert.False(t, criterion.IsValid())
			requireErrorChain(t, err, frequency.ErrInvalidTolerance, nil)
		})
	}
}

func TestCriterionCreation_NumericBoundaries(t *testing.T) {
	negativeZero := math.Copysign(0, -1)
	for _, tt := range []struct {
		name string
		c    frequency.Criterion
		kind frequency.Kind
		want float64
	}{
		{"K-24 negative zero hertz", frequency.MustWithinHertz(negativeZero), frequency.KindHertz, 0},
		{"negative zero relative", frequency.MustWithinRelative(negativeZero), frequency.KindRelative, 0},
		{"subnormal hertz", frequency.MustWithinHertz(math.SmallestNonzeroFloat64), frequency.KindHertz, math.SmallestNonzeroFloat64},
		{"maximum hertz", frequency.MustWithinHertz(math.MaxFloat64), frequency.KindHertz, math.MaxFloat64},
		{"subnormal relative", frequency.MustWithinRelative(math.SmallestNonzeroFloat64), frequency.KindRelative, math.SmallestNonzeroFloat64},
		{"one relative", frequency.MustWithinRelative(1), frequency.KindRelative, 1},
		{"maximum relative", frequency.MustWithinRelative(math.MaxFloat64), frequency.KindRelative, math.MaxFloat64},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.True(t, tt.c.IsValid())
			assert.True(t, tt.kind.Equal(tt.c.Kind()))
			got, ok := tt.c.Tolerance()
			require.True(t, ok)
			assert.Equal(t, tt.want, got)
			assert.False(t, math.Signbit(got))
		})
	}

	zeroCents, err := frequency.WithinCents(cents.Cents{})
	require.NoError(t, err)
	tolerance, ok := zeroCents.CentsTolerance()
	require.True(t, ok)
	assert.True(t, tolerance.Equal(cents.Zero()))
}

func TestCriterionCreation_Composites(t *testing.T) {
	t.Run("K-27 empty all", func(t *testing.T) {
		criterion, err := frequency.AllOf()
		assert.False(t, criterion.IsValid())
		requireErrorChain(t, err, frequency.ErrNoComponents, nil)
	})
	t.Run("empty any", func(t *testing.T) {
		criterion, err := frequency.AnyOf([]frequency.Criterion{}...)
		assert.False(t, criterion.IsValid())
		requireErrorChain(t, err, frequency.ErrNoComponents, nil)
	})
	t.Run("nil slice", func(t *testing.T) {
		criterion, err := frequency.AllOf(nil...)
		assert.False(t, criterion.IsValid())
		requireErrorChain(t, err, frequency.ErrNoComponents, nil)
	})
	t.Run("K-27 unset first", func(t *testing.T) {
		criterion, err := frequency.AnyOf(frequency.Criterion{}, centsCriterion(1))
		assert.False(t, criterion.IsValid())
		requireErrorChain(t, err, frequency.ErrUnset, nil)
	})
	t.Run("first unset by index", func(t *testing.T) {
		criterion, err := frequency.AllOf(centsCriterion(1), frequency.Criterion{}, frequency.Criterion{})
		assert.False(t, criterion.IsValid())
		requireErrorChain(t, err, frequency.ErrUnset, nil)
	})
	t.Run("empty precedes unset", func(t *testing.T) {
		criterion, err := frequency.AnyOf()
		assert.False(t, criterion.IsValid())
		requireErrorChain(t, err, frequency.ErrNoComponents, nil)
	})

	component := centsCriterion(1)
	all := frequency.MustAllOf(component, component)
	components := all.Components()
	require.Len(t, components, 2)
	requireCriterionEqual(t, component, components[0])
	requireCriterionEqual(t, component, components[1])

	nested := frequency.MustAnyOf(frequency.MustAllOf(frequency.Strict()), component)
	assert.True(t, nested.IsValid())
	assert.True(t, nested.Kind().IsComposite())
}

func TestCriterionReading_ControlExamples(t *testing.T) {
	negativeZero := frequency.MustWithinHertz(math.Copysign(0, -1))
	nested := frequency.MustAnyOf(centsCriterion(1), frequency.MustAllOf(frequency.MustWithinHertz(1)))
	tests := []struct {
		name         string
		c            frequency.Criterion
		valid        bool
		kind         frequency.Kind
		tolerance    float64
		hasTolerance bool
		cents        cents.Cents
		hasCents     bool
		components   []frequency.Criterion
	}{
		{name: "K-29 unset", c: frequency.Criterion{}, valid: false, kind: frequency.Kind(0)},
		{name: "K-29 strict", c: frequency.Strict(), valid: true, kind: frequency.KindStrict},
		{name: "K-29 cents", c: centsCriterion(5), valid: true, kind: frequency.KindCents, cents: cents.MustFromInt(5), hasCents: true},
		{name: "K-29 hertz", c: frequency.MustWithinHertz(1.5), valid: true, kind: frequency.KindHertz, tolerance: 1.5, hasTolerance: true},
		{name: "K-29 relative", c: frequency.MustWithinRelative(0.01), valid: true, kind: frequency.KindRelative, tolerance: 0.01, hasTolerance: true},
		{name: "K-29 negative zero", c: negativeZero, valid: true, kind: frequency.KindHertz, tolerance: 0, hasTolerance: true},
		{name: "K-29 nested", c: nested, valid: true, kind: frequency.KindAnyOf, components: []frequency.Criterion{centsCriterion(1), frequency.MustAllOf(frequency.MustWithinHertz(1))}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.valid, tt.c.IsValid())
			assert.True(t, tt.kind.Equal(tt.c.Kind()))
			tolerance, ok := tt.c.Tolerance()
			assert.Equal(t, tt.hasTolerance, ok)
			assert.Equal(t, tt.tolerance, tolerance)
			if tt.name == "K-29 negative zero" {
				assert.False(t, math.Signbit(tolerance))
			}
			centsTolerance, ok := tt.c.CentsTolerance()
			assert.Equal(t, tt.hasCents, ok)
			assert.True(t, centsTolerance.Equal(tt.cents))
			components := tt.c.Components()
			if tt.components == nil {
				assert.Nil(t, components)
				return
			}
			require.Len(t, components, len(tt.components))
			for i := range tt.components {
				requireCriterionEqual(t, tt.components[i], components[i])
			}
		})
	}
}

func TestKinds(t *testing.T) {
	expected := []frequency.Kind{
		frequency.KindStrict,
		frequency.KindCents,
		frequency.KindHertz,
		frequency.KindRelative,
		frequency.KindAllOf,
		frequency.KindAnyOf,
	}
	got := frequency.Kinds()
	require.Len(t, got, 6)
	assert.Equal(t, 6, cap(got))
	for i, kind := range expected {
		assert.True(t, kind.Equal(got[i]))
	}
	got[0] = frequency.Kind(255)
	got = got[:0]
	assert.Empty(t, got)
	fresh := frequency.Kinds()
	require.Len(t, fresh, 6)
	assert.Equal(t, 6, cap(fresh))
	for i, kind := range expected {
		assert.True(t, kind.Equal(fresh[i]))
	}
}

func TestKind_Methods(t *testing.T) {
	tests := []struct {
		kind         frequency.Kind
		number       int
		text         string
		valid        bool
		hasTolerance bool
		composite    bool
	}{
		{frequency.Kind(0), 0, "unset", false, false, false},
		{frequency.KindStrict, 1, "strict", true, false, false},
		{frequency.KindCents, 2, "cents", true, true, false},
		{frequency.KindHertz, 3, "hertz", true, true, false},
		{frequency.KindRelative, 4, "relative", true, true, false},
		{frequency.KindAllOf, 5, "all", true, false, true},
		{frequency.KindAnyOf, 6, "any", true, false, true},
		{frequency.Kind(7), 7, "Kind(7)", false, false, false},
		{frequency.Kind(255), 255, "Kind(255)", false, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.number, int(tt.kind))
			assert.Equal(t, tt.text, tt.kind.String())
			assert.Equal(t, tt.valid, tt.kind.IsValid())
			assert.Equal(t, tt.hasTolerance, tt.kind.HasTolerance())
			assert.Equal(t, tt.composite, tt.kind.IsComposite())
			assert.True(t, tt.kind.Equal(frequency.Kind(tt.number)))
			assert.False(t, tt.kind.Equal(frequency.Kind(tt.number+1)))
		})
	}
}

func TestCriterion_NotComparable(t *testing.T) {
	assert.False(t, reflect.TypeFor[frequency.Criterion]().Comparable())
	var unset = frequency.Criterion{}
	assert.False(t, unset.IsValid())
	assert.Equal(t, "unset", unset.String())
}
