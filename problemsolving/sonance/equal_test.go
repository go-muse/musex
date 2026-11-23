package sonance_test

import (
	"fmt"
	"testing"

	"github.com/go-muse/musex/problemsolving/sonance"

	"github.com/stretchr/testify/assert"
)

// huronWeights are the published weights of HuronProfile for the interval
// classes 0 to 6.
var huronWeights = []float64{0, -1.428, -0.582, 0.594, 0.386, 1.240, -0.453}

// huronPairWeights are the weights of HuronPairProfile for the tempered
// distances 0 to 11: each interval class weight on both of its members.
var huronPairWeights = []float64{0, -1.428, -0.582, 0.594, 0.386, 1.240, -0.453, 1.240, 0.386, 0.594, -0.582, -1.428}

// countingWeights returns the weights 1, 2, ..., size.
func countingWeights(size int) []float64 {
	w := make([]float64, size)
	for i := range w {
		w[i] = float64(i + 1)
	}
	return w
}

// withChanged returns a copy of w with the weight at i increased by 0.5.
func withChanged(w []float64, i int) []float64 {
	c := append([]float64(nil), w...)
	c[i] += 0.5
	return c
}

// TestProfileEqual covers Profile.Equal for every combination of set and
// unset profiles (issue #151): two unset profiles are equal, an unset profile
// equals no set profile, including the profile of zero weights, and two set
// profiles are equal exactly when their seven weights are equal.
func TestProfileEqual(t *testing.T) {
	t.Parallel()

	type tc struct {
		name string
		a, b func() sonance.Profile
		want bool
	}
	unsetP := func() sonance.Profile { return sonance.Profile{} }
	of := func(w []float64) func() sonance.Profile {
		return func() sonance.Profile { return sonance.MustNewProfile(w) }
	}

	cases := []tc{
		{"two unset profiles", unsetP, unsetP, true},
		{"unset and Huron", unsetP, sonance.HuronProfile, false},
		{"unset and zero weights", unsetP, of(make([]float64, 7)), false},
		{"unset and counting weights", unsetP, of(countingWeights(7)), false},
		{"Huron twice", sonance.HuronProfile, sonance.HuronProfile, true},
		{"Huron and its weights constructed", sonance.HuronProfile, of(huronWeights), true},
		{"zero weights twice", of(make([]float64, 7)), of(make([]float64, 7)), true},
		{"counting weights twice", of(countingWeights(7)), of(countingWeights(7)), true},
		{"Huron and counting weights", sonance.HuronProfile, of(countingWeights(7)), false},
	}
	for i := range 7 {
		cases = append(cases, tc{
			fmt.Sprintf("weight of interval class %d differs", i),
			of(countingWeights(7)), of(withChanged(countingWeights(7), i)), false,
		})
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			a, b := c.a(), c.b()
			assert.Equal(t, c.want, a.Equal(b), "a.Equal(b)")
			assert.Equal(t, c.want, b.Equal(a), "b.Equal(a): symmetry")
			assert.True(t, a.Equal(a), "a.Equal(a): reflexivity")
			assert.True(t, b.Equal(b), "b.Equal(b): reflexivity")
		})
	}
}

// TestPairProfileEqual covers PairProfile.Equal for every combination of set
// and unset profiles (issue #151): two unset profiles are equal, an unset
// profile equals no set profile, including the profile of zero weights, and
// two set profiles are equal exactly when their twelve weights are equal.
func TestPairProfileEqual(t *testing.T) {
	t.Parallel()

	type tc struct {
		name string
		a, b func() sonance.PairProfile
		want bool
	}
	unsetP := func() sonance.PairProfile { return sonance.PairProfile{} }
	of := func(w []float64) func() sonance.PairProfile {
		return func() sonance.PairProfile { return sonance.MustNewPairProfile(w) }
	}

	cases := []tc{
		{"two unset profiles", unsetP, unsetP, true},
		{"unset and Huron", unsetP, sonance.HuronPairProfile, false},
		{"unset and zero weights", unsetP, of(make([]float64, 12)), false},
		{"unset and counting weights", unsetP, of(countingWeights(12)), false},
		{"Huron twice", sonance.HuronPairProfile, sonance.HuronPairProfile, true},
		{"Huron and its weights constructed", sonance.HuronPairProfile, of(huronPairWeights), true},
		{"zero weights twice", of(make([]float64, 12)), of(make([]float64, 12)), true},
		{"counting weights twice", of(countingWeights(12)), of(countingWeights(12)), true},
		{"Huron and counting weights", sonance.HuronPairProfile, of(countingWeights(12)), false},
	}
	for i := range 12 {
		cases = append(cases, tc{
			fmt.Sprintf("weight of distance %d differs", i),
			of(countingWeights(12)), of(withChanged(countingWeights(12), i)), false,
		})
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			a, b := c.a(), c.b()
			assert.Equal(t, c.want, a.Equal(b), "a.Equal(b)")
			assert.Equal(t, c.want, b.Equal(a), "b.Equal(a): symmetry")
			assert.True(t, a.Equal(a), "a.Equal(a): reflexivity")
			assert.True(t, b.Equal(b), "b.Equal(b): reflexivity")
		})
	}
}
