package system_test

import (
	"fmt"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaletteEdges(t *testing.T) {
	t.Parallel()

	var nilSys *system.System
	_, err := nilSys.Palette(system.Sharps())
	assert.ErrorIs(t, err, system.ErrInvalid)

	_, err = system.Traditional().Palette(system.SpellingRule{})
	assert.ErrorIs(t, err, system.ErrInvalid)

	// No rules give empty groups in class order.
	groups, err := system.Traditional().Palette()
	require.NoError(t, err)
	require.Len(t, groups, 12)
	for i, g := range groups {
		assert.Equal(t, i, g.Class.Number())
		assert.Empty(t, g.Spellings)
	}

	// A chromatic rule does not apply to a non-traditional system.
	_, err = system.Positions19().Palette(system.MustChromaticScale(spelling.C(), system.Major, system.Ascending))
	assert.ErrorIs(t, err, system.ErrNotApplicable)

	_, err = system.Traditional().PaletteWithin(alteration.MustFromInt(-1))
	assert.ErrorIs(t, err, system.ErrLimit)

	// The result is owned by the caller.
	g1 := system.Traditional().MustPalette(system.Sharps())
	g2 := system.Traditional().MustPalette(system.Sharps())
	g1[0].Spellings[0] = spelling.Spelling{}
	assert.True(t, g2[0].Spellings[0].IsValid())
}

func ExampleSystem_Palette() {
	groups := system.Traditional().MustPalette(system.Flats(), system.Sharps())
	total := 0
	for _, g := range groups {
		total += len(g.Spellings)
	}
	fmt.Println(len(groups), total)
	fmt.Println(groups[1].Spellings[0], groups[1].Spellings[1])
	// Output:
	// 12 17
	// D-1 C+1
}
