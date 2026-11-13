package settings_test

import (
	"testing"

	"github.com/go-muse/musex/notation"
	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/octave/number"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"

	"github.com/stretchr/testify/assert"
)

// Sinks keep the measured results alive.
var (
	sinkBool     bool
	sinkSettings *settings.Settings
	sinkSystem   *system.System
	sinkNotation notation.Notation
	sinkNumber   number.Numbering
	sinkCatalog  *naming.Catalog
	sinkContext  notation.OctaveContext
	sinkTuning   *tuning.Tuning
)

// TestAllocations covers NF-3 and A 5.13. testing.AllocsPerRun measures the
// whole process, so these tests do not run in parallel.
func TestAllocations(t *testing.T) {
	std := settings.Standard()
	other := settings.Standard()
	german := std.MustWithDesignations(builtin(notation.GermanAMID), number.FrancoBelgian(), naming.Russian())
	german2 := settings.Standard().MustWithDesignations(builtin(notation.GermanAMID), number.FrancoBelgian(), naming.Russian())
	ptr := &pointerSource{s: std}

	for _, tc := range []struct {
		name string
		max  float64
		f    func()
	}{
		{name: "IsValid", max: 0, f: func() { sinkBool = std.IsValid() }},
		{name: "System", max: 0, f: func() { sinkSystem = std.System() }},
		{name: "Notation", max: 0, f: func() { sinkNotation = std.Notation() }},
		{name: "Numbering", max: 0, f: func() { sinkNumber = std.Numbering() }},
		{name: "Settings", max: 0, f: func() { sinkSettings = std.Settings() }},
		{name: "Equal standard", max: 0, f: func() { sinkBool = std.Equal(other) }},
		{name: "Equal built-in parts", max: 0, f: func() { sinkBool = german.Equal(german2) }},
		{name: "Equal unequal", max: 0, f: func() { sinkBool = std.Equal(german) }},
		{name: "SameSource sets", max: 0, f: func() { sinkBool = settings.SameSource(std, other) }},
		{name: "SameSource pointers", max: 0, f: func() { sinkBool = settings.SameSource(ptr, ptr) }},
		{name: "Standard", max: 1, f: func() { sinkSettings = settings.Standard() }},
		{name: "Tuning", max: 1, f: func() { sinkTuning = std.Tuning() }},
		{name: "Catalog", max: 1, f: func() { sinkCatalog = std.Catalog() }},
		{name: "OctaveContext", max: 1, f: func() { sinkContext = std.OctaveContext() }},
	} {
		got := testing.AllocsPerRun(100, tc.f)
		assert.LessOrEqual(t, got, tc.max, "%s allocates %v times", tc.name, got)
	}
}
