package scale_test

import (
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"github.com/go-muse/musex/scale"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Package boundaries are public metadata; no implementation bodies are opened.
func TestPackageDependencyDirectionAndClosedAggregate(t *testing.T) {
	t.Parallel()
	typ := reflect.TypeFor[scale.Scale]()
	for i := range typ.NumField() {
		assert.False(t, typ.Field(i).IsExported(), "Scale state must be private")
	}
	for _, tt := range []struct {
		pkg                 string
		required, forbidden []string
	}{
		{"scale", nil, []string{"github.com/go-muse/musex/mode", "github.com/go-muse/musex/template"}},
		{"mode", []string{"github.com/go-muse/musex/scale"}, nil},
	} {
		t.Run(tt.pkg, func(t *testing.T) {
			cmd := exec.CommandContext(t.Context(), "go", "list", "-f", `{{join .Imports "\n"}}`, "github.com/go-muse/musex/"+tt.pkg)
			metadata, err := cmd.Output()
			require.NoError(t, err, "could not read import metadata")
			imports := strings.Fields(string(metadata))
			for _, pkg := range tt.required {
				assert.Contains(t, imports, pkg)
			}
			for _, pkg := range tt.forbidden {
				assert.NotContains(t, imports, pkg)
			}
		})
	}
	cmd := exec.CommandContext(t.Context(), "go", "doc", "-short", "github.com/go-muse/musex/scale")
	declarations, err := cmd.Output()
	require.NoError(t, err, "could not read exported declaration metadata")
	for _, pattern := range []string{`(?m)^type ModeSource\b`, `(?m)^func (Must)?FromMode\w*\(`, `(?m)^type BuildOptions\b`, `(?m)^type ModeStart\b`} {
		assert.NotRegexp(t, pattern, string(declarations), "mode construction must belong to mode")
	}
}
