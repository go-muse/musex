package tuning_test

import (
	"testing"

	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"
	"github.com/go-muse/musex/tuning/temperament"

	"github.com/stretchr/testify/assert"
)

var (
	allocPitch       pitch.Pitch
	allocSystem      *system.System
	allocTemperament temperament.Temperament
	allocClass       pitchclass.Class
	allocReference   tuning.Reference
	allocSpelling    spelling.Spelling
	allocInt         int
	allocBool        bool
	allocFloat       float64
)

func TestZeroAllocationReaders(t *testing.T) {
	ref := tuning.A4At440()
	tune := tuning.Standard()
	root := tuning.RootFromClass(pitchclass.MustNew(0, 12))
	rootName := tuning.RootFromName(spelling.C())

	t.Run("Reference", func(t *testing.T) {
		allocs := testing.AllocsPerRun(200, func() {
			allocPitch = ref.Pitch()
			allocFloat = ref.Frequency()
			allocBool = ref.IsValid()
		})
		assert.Zero(t, allocs)
	})

	t.Run("Tuning", func(t *testing.T) {
		allocs := testing.AllocsPerRun(200, func() {
			allocSystem = tune.System()
			allocTemperament = tune.Temperament()
			allocClass = tune.RootClass()
			allocReference = tune.Reference()
			allocInt = tune.Classes()
			allocBool = tune.IsValid()
		})
		assert.Zero(t, allocs)
	})

	t.Run("Root", func(t *testing.T) {
		allocs := testing.AllocsPerRun(200, func() {
			allocClass, allocBool = root.Class()
			allocSpelling, allocBool = rootName.Name()
			allocBool = root.IsValid()
		})
		assert.Zero(t, allocs)
	})
}

func TestZeroAllocationValueEquality(t *testing.T) {
	ref := tuning.A4At440()
	otherRef := tuning.A4At440()
	root := tuning.RootFromClass(pitchclass.MustNew(0, 12))
	otherRoot := tuning.RootFromClass(pitchclass.MustNew(0, 12))

	allocs := testing.AllocsPerRun(200, func() {
		allocBool = ref.Equal(otherRef)
		allocBool = root.Equal(otherRoot)
	})
	assert.Zero(t, allocs)
}
