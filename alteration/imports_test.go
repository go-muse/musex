package alteration_test

import (
	"go/build"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// K-18: the package imports only the standard library: no musex packages and
// no third-party modules. NF-7: no arbitrary-precision numbers (math/big).
// The check reads the import list of the non-test files of the package
// directory with go/build.
func TestPackageImports(t *testing.T) {
	t.Parallel()

	pkg, err := build.ImportDir(".", 0)
	require.NoError(t, err)
	require.Equal(t, "alteration", pkg.Name)
	require.NotEmpty(t, pkg.GoFiles, "the package has non-test Go files")

	for _, path := range pkg.Imports {
		assert.False(t, strings.HasPrefix(path, "github.com/go-muse/musex"), "musex package imported: %s", path)

		first, _, _ := strings.Cut(path, "/")
		assert.NotContains(t, first, ".", "non-standard package imported: %s", path)
	}

	assert.NotContains(t, pkg.Imports, "math/big", "no arbitrary-precision numbers")
}
