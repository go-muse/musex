package frequency_test

import (
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/frequency"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCriterionEqual_ControlExamples(t *testing.T) {
	oneCent := centsCriterion(1)
	oneHertz := frequency.MustWithinHertz(1)
	tests := []struct {
		name string
		c, d frequency.Criterion
		want bool
	}{
		{"K-28a same cents", oneCent, centsCriterion(1), true},
		{"K-28b different cents", oneCent, centsCriterion(2), false},
		{"K-28v strict and zero hertz", frequency.Strict(), frequency.MustWithinHertz(0), false},
		{"K-28g component order", frequency.MustAllOf(oneCent, oneHertz), frequency.MustAllOf(oneHertz, oneCent), false},
		{"K-28d all and any", frequency.MustAllOf(oneCent), frequency.MustAnyOf(oneCent), false},
		{"K-28d composite and component", frequency.MustAllOf(oneCent), oneCent, false},
		{"K-28e both unset", frequency.Criterion{}, frequency.Criterion{}, true},
		{"K-28e unset and strict", frequency.Criterion{}, frequency.Strict(), false},
		{"negative zero hertz", frequency.MustWithinHertz(negativeZero()), frequency.MustWithinHertz(0), true},
		{"negative zero relative", frequency.MustWithinRelative(negativeZero()), frequency.MustWithinRelative(0), true},
		{"same ratio cents", frequency.MustWithinCents(cents.MustFromRatio(81, 80)), frequency.MustWithinCents(cents.MustFromRatio(81, 80)), true},
		{"ratio and approximate cents differ", frequency.MustWithinCents(cents.MustFromRatio(3, 2)), frequency.MustWithinCents(cents.MustFromRatioFloat(1.5)), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.c.Equal(tt.d))
			assert.Equal(t, tt.want, tt.d.Equal(tt.c))
			assert.Equal(t, tt.want, tt.c.String() == tt.d.String())
		})
	}
}

func TestCriterionEqual_PropertiesAndTextBijection(t *testing.T) {
	criteria := definedCriteria()
	criteria = append(criteria, struct {
		name string
		c    frequency.Criterion
	}{"unset", frequency.Criterion{}})
	for _, first := range criteria {
		for _, second := range criteria {
			t.Run(first.name+"/"+second.name, func(t *testing.T) {
				t.Parallel()
				got := first.c.Equal(second.c)
				assert.Equal(t, first.c.String() == second.c.String(), got)
				assert.Equal(t, second.c.Equal(first.c), got)
				assert.True(t, first.c.Equal(first.c))
			})
		}
	}

	for _, first := range criteria {
		for _, second := range criteria {
			for _, third := range criteria {
				if first.c.Equal(second.c) && second.c.Equal(third.c) {
					assert.True(t, first.c.Equal(third.c), "Equal must be transitive")
				}
			}
		}
	}
}

func TestCriterionEqual_SingleComponent(t *testing.T) {
	component := centsCriterion(5)
	for _, composite := range []frequency.Criterion{
		frequency.MustAllOf(component),
		frequency.MustAnyOf(component),
	} {
		assert.False(t, composite.Equal(component))
		assert.NotEqual(t, component.String(), composite.String())
		for _, pair := range frequencyPairs() {
			want, err := component.Near(pair.f1, pair.f2)
			require.NoError(t, err)
			got, err := composite.Near(pair.f1, pair.f2)
			require.NoError(t, err)
			assert.Equal(t, want, got)
		}
	}
}
