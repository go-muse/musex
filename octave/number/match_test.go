package number_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/octave/number"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// labels returns the labels of the numberings in order.
func labels(numberings []number.Numbering) []string {
	result := make([]string, 0, len(numberings))
	for _, n := range numberings {
		result = append(result, n.Label())
	}

	return result
}

// MatchBuiltins returns every built-in numbering consistent with all the
// observations, in Builtins order, with labels preserved.
func TestMatchBuiltins(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		obs  []number.Observation
		want []string
	}{
		{
			name: "no observations",
			obs:  nil,
			want: []string{"ASPN", "Franco-Belgian", "Mexican", "Middle C = C3", "Middle C = C5"},
		},
		{
			name: "middle C as 3 is ambiguous",
			obs:  []number.Observation{{ASPN: 4, Number: 3}},
			want: []string{"Franco-Belgian", "Middle C = C3"},
		},
		{
			name: "second observation resolves the ambiguity",
			obs:  []number.Observation{{ASPN: 4, Number: 3}, {ASPN: 1, Number: 0}},
			want: []string{"Middle C = C3"},
		},
		{
			name: "middle C as 5 keeps both labels",
			obs:  []number.Observation{{ASPN: 4, Number: 5}},
			want: []string{"Mexican", "Middle C = C5"},
		},
		{
			name: "middle C as 4 is ASPN",
			obs:  []number.Observation{{ASPN: 4, Number: 4}},
			want: []string{"ASPN"},
		},
		{
			name: "contradiction across octaves",
			obs:  []number.Observation{{ASPN: 4, Number: 3}, {ASPN: 2, Number: 2}},
			want: []string{},
		},
		{
			name: "contradiction at one octave",
			obs:  []number.Observation{{ASPN: 4, Number: 3}, {ASPN: 4, Number: 4}},
			want: []string{},
		},
		{
			name: "identical repeats are harmless",
			obs:  []number.Observation{{ASPN: 4, Number: 3}, {ASPN: 4, Number: 3}},
			want: []string{"Franco-Belgian", "Middle C = C3"},
		},
		{
			name: "skipped number never matches",
			obs:  []number.Observation{{ASPN: 2, Number: 0}},
			want: []string{},
		},
		{
			name: "zero below the Franco-Belgian skip",
			obs:  []number.Observation{{ASPN: 0, Number: -2}},
			want: []string{"Franco-Belgian"},
		},
		{
			name: "largest int, identity",
			obs:  []number.Observation{{ASPN: math.MaxInt, Number: math.MaxInt}},
			want: []string{"ASPN"},
		},
		{
			name: "largest int, minus one",
			obs:  []number.Observation{{ASPN: math.MaxInt, Number: math.MaxInt - 1}},
			want: []string{"Franco-Belgian", "Middle C = C3"},
		},
		{
			name: "smallest int, plus one",
			obs:  []number.Observation{{ASPN: math.MinInt, Number: math.MinInt + 1}},
			want: []string{"Mexican", "Middle C = C5"},
		},
		{
			name: "smallest int, identity",
			obs:  []number.Observation{{ASPN: math.MinInt, Number: math.MinInt}},
			want: []string{"ASPN"},
		},
		{
			name: "smallest int, shifted results overflow",
			obs:  []number.Observation{{ASPN: math.MinInt, Number: 0}},
			want: []string{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := number.MatchBuiltins(tc.obs...)
			require.NotNil(t, got, "the result is never nil")
			assert.Equal(t, tc.want, labels(got))

			for _, n := range got {
				assert.True(t, n.IsDefined())
				assert.NotEmpty(t, n.Label())
			}
		})
	}
}

// The matches are equal to the built-in functions and keep their labels:
// Mexican and Middle C = C5 are both returned although they are equal.
func TestMatchBuiltinsLabels(t *testing.T) {
	t.Parallel()

	got := number.MatchBuiltins(number.Observation{ASPN: 4, Number: 5})
	require.Len(t, got, 2)

	assert.True(t, got[0].Equal(number.Mexican()))
	assert.True(t, got[1].Equal(number.MiddleC5()))
	assert.True(t, got[0].Equal(got[1]), "the two numberings give the same numbers")
	assert.NotEqual(t, got[0].Label(), got[1].Label(), "the labels differ")
}

// The caller owns the result: modifying it does not affect later calls.
func TestMatchBuiltinsOwnership(t *testing.T) {
	t.Parallel()

	got := number.MatchBuiltins()
	require.Len(t, got, 5)

	got[0] = number.Numbering{}

	again := number.MatchBuiltins()
	require.Len(t, again, 5)
	assert.Equal(t, "ASPN", again[0].Label())
}

// Matches is the per-numbering check: a defined numbering matches the
// observations it satisfies, contradictory observations never match, and an
// undefined numbering matches nothing.
func TestMatches(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		n    number.Numbering
		obs  []number.Observation
		want bool
	}{
		{name: "defined without observations", n: number.ASPN(), obs: nil, want: true},
		{name: "undefined without observations", n: number.Numbering{}, obs: nil, want: false},
		{name: "undefined with observations", n: number.Numbering{}, obs: []number.Observation{{ASPN: 4, Number: 4}}, want: false},
		{name: "satisfied", n: number.MiddleC3(), obs: []number.Observation{{ASPN: 4, Number: 3}}, want: true},
		{name: "own numbering", n: number.MustNew(4, 3), obs: []number.Observation{{ASPN: 4, Number: 3}}, want: true},
		{name: "wrong number", n: number.MiddleC3(), obs: []number.Observation{{ASPN: 4, Number: 4}}, want: false},
		{name: "contradiction", n: number.MiddleC3(), obs: []number.Observation{{ASPN: 4, Number: 3}, {ASPN: 4, Number: 4}}, want: false},
		{name: "identical repeats", n: number.MiddleC3(), obs: []number.Observation{{ASPN: 4, Number: 3}, {ASPN: 4, Number: 3}}, want: true},
		{name: "skipped number", n: number.FrancoBelgian(), obs: []number.Observation{{ASPN: 1, Number: 0}}, want: false},
		{name: "out of range result", n: number.MiddleC3(), obs: []number.Observation{{ASPN: math.MinInt, Number: 0}}, want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, tc.n.Matches(tc.obs...))
		})
	}
}

// Matches does not change the numbering.
func TestMatchesKeepsNumbering(t *testing.T) {
	t.Parallel()

	n := number.FrancoBelgian()
	before := number.FrancoBelgian()

	n.Matches(number.Observation{ASPN: 4, Number: 3}, number.Observation{ASPN: 1, Number: -1})

	assert.True(t, n.Equal(before))
	assert.Equal(t, before.String(), n.String())
}
