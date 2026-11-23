package system_test

import (
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Palette groups the spellings the rules produce by class, preserving the
// policy order.
func TestPaletteAcceptance(t *testing.T) {
	t.Parallel()

	sys := system.Traditional()
	groups, err := sys.Palette(system.Flats(), system.Sharps())
	require.NoError(t, err)
	require.Len(t, groups, 12)

	total := 0
	for _, g := range groups {
		total += len(g.Spellings)
	}
	assert.Equal(t, 17, total)

	// The class of C♯/D♭ offers D♭ before C♯.
	g := groups[1]
	require.Len(t, g.Spellings, 2)
	assert.True(t, g.Spellings[0].Equal(spelling.DFlat()))
	assert.True(t, g.Spellings[1].Equal(spelling.CSharp()))

	// PaletteWithin offers every admissible spelling within the limit.
	within, err := sys.PaletteWithin(alteration.MustFromInt(1))
	require.NoError(t, err)
	require.Len(t, within, 12)
	var hasBSharp bool
	for _, sp := range within[0].Spellings {
		if sp.Equal(spelling.BSharp()) {
			hasBSharp = true
		}
	}
	assert.True(t, hasBSharp, "class of C offers B♯ within limit 1")
}
