package tuplet_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// namedTuplet is a row of table 4.4 of the architecture (table 2.3 of the
// requirements).
type namedTuplet struct {
	fn             func() tuplet.Tuplet
	actual, normal int
	name, russian  string
}

var namedTuplets = []namedTuplet{
	{fn: tuplet.Duplet, actual: 2, normal: 3, name: "duplet", russian: "дуоль"},
	{fn: tuplet.Triplet, actual: 3, normal: 2, name: "triplet", russian: "триоль"},
	{fn: tuplet.Quadruplet, actual: 4, normal: 3, name: "quadruplet", russian: "квартоль"},
	{fn: tuplet.Quintuplet, actual: 5, normal: 4, name: "quintuplet", russian: "квинтоль"},
	{fn: tuplet.Sextuplet, actual: 6, normal: 4, name: "sextuplet", russian: "секстоль"},
	{fn: tuplet.Septuplet, actual: 7, normal: 4, name: "septuplet", russian: "септоль"},
	{fn: tuplet.Octuplet, actual: 8, normal: 6, name: "octuplet", russian: "октоль"},
	{fn: tuplet.Nonuplet, actual: 9, normal: 8, name: "nonuplet", russian: "нонемоль"},
	{fn: tuplet.Decuplet, actual: 10, normal: 8, name: "decuplet", russian: "децимоль"},
}

// TestNamedTuplets covers O-2, K-2, K-3 and KP-15: pairs and names of the nine
// named tuplets.
func TestNamedTuplets(t *testing.T) {
	t.Parallel()

	for _, nt := range namedTuplets {
		t.Run(nt.name, func(t *testing.T) {
			t.Parallel()

			got := nt.fn()
			assertTuplet(t, got, nt.actual, nt.normal)
			assert.Equal(t, got, nt.fn(), "repeated calls return the same value")

			name, ok, err := got.Name()
			require.NoError(t, err)
			assert.True(t, ok)
			assert.Equal(t, nt.name, name)

			russian, ok, err := got.RussianName()
			require.NoError(t, err)
			assert.True(t, ok)
			assert.Equal(t, nt.russian, russian)
		})
	}
}

// TestNamedTuplets_DefaultPair covers K-3 and architecture 4.4: a named tuplet
// is Default(n, Binary), or Default(n, Ternary) when there is no binary
// default; the pair also follows rule P-5 computed independently.
func TestNamedTuplets_DefaultPair(t *testing.T) {
	t.Parallel()

	for _, nt := range namedTuplets {
		t.Run(nt.name, func(t *testing.T) {
			t.Parallel()

			want, err := tuplet.Default(nt.actual, tuplet.Binary)
			if err != nil {
				assertErr(t, err, tuplet.ErrNoDefault)

				want, err = tuplet.Default(nt.actual, tuplet.Ternary)
				require.NoError(t, err)
			}

			assert.Equal(t, want, nt.fn())

			normal, ok := oracleDefault(nt.actual, tuplet.Binary)
			if !ok {
				normal, ok = oracleDefault(nt.actual, tuplet.Ternary)
			}

			require.True(t, ok)
			assert.Equal(t, normal, nt.normal)

			has, err := nt.fn().HasDefaultNormal()
			require.NoError(t, err)
			assert.True(t, has)
		})
	}
}

// TestNames covers KP-15 and architecture 4.11: names depend only on actual;
// there is no name for actual 1 and actual 11 or more.
func TestNames(t *testing.T) {
	t.Parallel()

	byActual := map[int]namedTuplet{}
	for _, nt := range namedTuplets {
		byActual[nt.actual] = nt
	}

	actuals := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 100, math.MaxInt}
	normals := []int{1, 2, 3, 4, 11, math.MaxInt}

	for _, actual := range actuals {
		for _, normal := range normals {
			t.Run(itoa(actual)+":"+itoa(normal), func(t *testing.T) {
				t.Parallel()

				tp := tuplet.MustNew(actual, normal)
				nt, named := byActual[actual]

				name, ok, err := tp.Name()
				require.NoError(t, err)
				assert.Equal(t, named, ok)
				assert.Equal(t, nt.name, name)

				russian, ok, err := tp.RussianName()
				require.NoError(t, err)
				assert.Equal(t, named, ok)
				assert.Equal(t, nt.russian, russian)
			})
		}
	}
}

// TestNames_AnyNormal covers architecture 4.11 examples: 5:3 and 7:11 are a
// quintuplet and a septuplet.
func TestNames_AnyNormal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		tp            tuplet.Tuplet
		name, russian string
	}{
		{tp: tuplet.MustNew(5, 3), name: "quintuplet", russian: "квинтоль"},
		{tp: tuplet.MustNew(7, 11), name: "septuplet", russian: "септоль"},
	}

	for _, tt := range tests {
		t.Run(tt.tp.String(), func(t *testing.T) {
			t.Parallel()

			name, ok := tt.tp.MustName()
			assert.True(t, ok)
			assert.Equal(t, tt.name, name)

			russian, ok := tt.tp.MustRussianName()
			assert.True(t, ok)
			assert.Equal(t, tt.russian, russian)
		})
	}
}
