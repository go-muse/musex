package system_test

import (
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaletteMustAndNil(t *testing.T) {
	t.Parallel()

	groups := system.Traditional().MustPaletteWithin(alteration.MustFromInt(1))
	require.Len(t, groups, 12)

	assert.Panics(t, func() { system.Traditional().MustPaletteWithin(alteration.MustFromInt(-1)) })
	assert.Panics(t, func() { system.Traditional().MustPalette(system.SpellingRule{}) })

	var nilSys *system.System
	_, err := nilSys.PaletteWithin(alteration.MustFromInt(1))
	assert.ErrorIs(t, err, system.ErrInvalid)
}
