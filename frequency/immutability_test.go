package frequency_test

import (
	"testing"

	"github.com/go-muse/musex/frequency"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCriterionComponents_ReturnIndependentSlice(t *testing.T) {
	oneCent := centsCriterion(1)
	oneHertz := frequency.MustWithinHertz(1)
	criterion := frequency.MustAllOf(oneCent, oneHertz)

	components := criterion.Components()
	require.Len(t, components, 2)
	components[0] = frequency.Strict()
	components = components[:0]
	assert.Empty(t, components)

	fresh := criterion.Components()
	require.Len(t, fresh, 2)
	requireCriterionEqual(t, oneCent, fresh[0])
	requireCriterionEqual(t, oneHertz, fresh[1])
	near, err := criterion.Near(440, 441)
	require.NoError(t, err)
	assert.False(t, near)
	assert.Equal(t, "all(cents(1), hertz(1))", criterion.String())
}

func TestCompositeConstruction_DoesNotRetainOrModifyInput(t *testing.T) {
	for _, tt := range []struct {
		name string
		call func(...frequency.Criterion) (frequency.Criterion, error)
		text string
	}{
		{"all", frequency.AllOf, "all(cents(1), hertz(1))"},
		{"any", frequency.AnyOf, "any(cents(1), hertz(1))"},
	} {
		t.Run(tt.name+" success", func(t *testing.T) {
			oneCent := centsCriterion(1)
			oneHertz := frequency.MustWithinHertz(1)
			input := []frequency.Criterion{oneCent, oneHertz}
			criterion, err := tt.call(input...)
			require.NoError(t, err)
			requireCriterionEqual(t, oneCent, input[0])
			requireCriterionEqual(t, oneHertz, input[1])

			input[0] = frequency.Strict()
			input[1] = frequency.Criterion{}
			assert.Equal(t, tt.text, criterion.String())
			components := criterion.Components()
			require.Len(t, components, 2)
			requireCriterionEqual(t, oneCent, components[0])
			requireCriterionEqual(t, oneHertz, components[1])
		})
	}
}

func TestCompositeConstruction_FailureDoesNotModifyInput(t *testing.T) {
	for _, tt := range []struct {
		name string
		call func(...frequency.Criterion) (frequency.Criterion, error)
	}{
		{"all", frequency.AllOf},
		{"any", frequency.AnyOf},
	} {
		t.Run(tt.name, func(t *testing.T) {
			input := []frequency.Criterion{frequency.Strict(), {}}
			criterion, err := tt.call(input...)
			assert.False(t, criterion.IsValid())
			requireErrorChain(t, err, frequency.ErrUnset, nil)
			requireCriterionEqual(t, frequency.Strict(), input[0])
			assert.False(t, input[1].IsValid())
		})
	}
}
