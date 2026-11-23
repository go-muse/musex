package sonance_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/problemsolving/sonance"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
)

// TestDistanceOverflow covers the out-of-range path at the edges of int.
func TestDistanceOverflow(t *testing.T) {
	t.Parallel()

	ns := []note.Note{
		note.MustFromASPN(spelling.C(), math.MaxInt, cents.Zero(), nil),
		note.MustFromASPN(spelling.C(), math.MinInt, cents.Zero(), nil),
	}
	_, err := sonance.ScoreNotes(system.Traditional(), ns, sonance.HuronPairProfile())
	assert.ErrorIs(t, err, sonance.ErrOutOfRange)
	assert.ErrorIs(t, err, system.ErrOutOfRange)
	assert.Panics(t, func() { sonance.MustScoreNotes(system.Traditional(), ns, sonance.HuronPairProfile()) })
}
