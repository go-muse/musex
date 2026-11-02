package naming_test

import (
	"testing"

	"github.com/go-muse/musex/octave/naming"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Must wrappers on valid input: MustNew does not
// panic and returns a new non-nil catalog equal to the result of New. The
// panics on invalid input are checked in TestNewDuplicateName (K-8).
func TestMustNewValid(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		pairs func() map[int]string
	}{
		{name: "K-7 pairs", pairs: customPairs},
		{name: "K-7 exact strings", pairs: func() map[int]string { return map[int]string{3: "малая октава", 5: " x "} }},
		{name: "nil map", pairs: func() map[int]string { return nil }},
		{name: "empty map", pairs: func() map[int]string { return map[int]string{} }},
		{name: "empty names only", pairs: func() map[int]string { return map[int]string{1: "", 2: ""} }},
		{name: "Russian oracle", pairs: russianOracle},
		{name: "English oracle", pairs: englishOracle},
		{name: "German oracle", pairs: germanOracle},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var must *naming.Catalog
			require.NotPanics(t, func() { must = naming.MustNew(tc.pairs()) })
			require.NotNil(t, must)

			c, err := naming.New(tc.pairs())
			require.NoError(t, err)
			require.NotNil(t, c)

			assert.NotSame(t, c, must)
			assert.True(t, must.Equal(c))

			for aspn, name := range tc.pairs() {
				assert.Equal(t, name, must.Name(aspn), "ASPN %d", aspn)
			}
		})
	}
}

// Must wrappers on valid input: MustASPN does not panic and returns the same
// number as ASPN for every name of the built-in catalogs (K-5) and of the
// custom catalog of K-7. The panics on unknown names are checked in
// TestUnknownName (K-6) and TestZeroAndNil (K-11).
func TestMustASPNValid(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		catalog func() *naming.Catalog
		oracle  map[int]string
	}{
		{name: "Russian", catalog: naming.Russian, oracle: russianOracle()},
		{name: "English", catalog: naming.English, oracle: englishOracle()},
		{name: "German", catalog: naming.German, oracle: germanOracle()},
		{name: "custom", catalog: func() *naming.Catalog { return naming.MustNew(customPairs()) }, oracle: customPairs()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := tc.catalog()
			require.NotNil(t, c)

			for want, name := range tc.oracle {
				var got int
				if assert.NotPanics(t, func() { got = c.MustASPN(name) }, "name %q", name) {
					assert.Equal(t, want, got, "name %q", name)
				}

				aspn, err := c.ASPN(name)
				if assert.NoError(t, err, "name %q", name) {
					assert.Equal(t, aspn, got, "name %q", name)
				}
			}
		})
	}
}
