package tempomap_test

import (
	"testing"

	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/meter"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempomap"

	"github.com/stretchr/testify/assert"
)

func TestRegressionPositionSubtractsActiveStartExtensionOnce(t *testing.T) {
	t.Parallel()
	t.Run("constant origin", func(t *testing.T) {
		t.Parallel()
		m := tempomap.MustNew(quarter(60)).MustWithExtension(
			tempomap.MustNewExtension(tempomap.MustAt(duration.Origin()), sec(1, 1)),
		)
		checkPositionResult(t, m.MustPosition(sec(3, 2), 3, rounding.HalfUp), "1/8", false, 0, "0/1")
	})

	t.Run("constant jump", func(t *testing.T) {
		t.Parallel()
		m := newJumpMap().MustWithExtension(tempomap.MustNewExtension(at(2, 1), sec(1, 1)))
		// Start(2)=8 and End(2)=9. At 9.5 only half a second remains
		// on the q=120 segment, so p=2+30*(1/2)/60=9/4.
		checkPositionResult(t, m.MustPosition(sec(19, 2), 3, rounding.HalfUp), "9/4", false, 0, "0/1")
	})

	t.Run("initial gradual", func(t *testing.T) {
		t.Parallel()
		m := newGradualMap(pos(1, 1), quarter(120)).MustWithExtension(
			tempomap.MustNewExtension(tempomap.MustAt(duration.Origin()), sec(1, 1)),
		)
		// One second is the extension; the remaining second is the KP-10 input.
		checkPositionResult(t, m.MustPosition(sec(2, 1), 6, rounding.HalfUp), "284025/1000000", true, 6, "1/2000000")
	})
}

func TestRegressionAddressAtSubtractsActiveStartExtensionOnce(t *testing.T) {
	t.Parallel()
	grid := measure.MustNew(meter.FourFour())

	t.Run("constant origin", func(t *testing.T) {
		t.Parallel()
		m := tempomap.MustNewMeasured(grid, quarter(60)).MustWithExtension(
			tempomap.MustNewExtension(bar(1, 0, 1), sec(1, 1)), grid,
		)
		address, position := m.MustAddressAt(grid, sec(3, 2), 3, rounding.HalfUp)
		assert.True(t, measure.MustNewAddress(1, pos(1, 8)).Equal(address))
		checkPositionResult(t, position, "1/8", false, 0, "0/1")
	})

	t.Run("constant jump", func(t *testing.T) {
		t.Parallel()
		m := tempomap.MustNewMeasured(grid, quarter(60)).
			MustWithChange(tempomap.MustNewJump(bar(2, 0, 1), quarter(120)), grid).
			MustWithExtension(tempomap.MustNewExtension(bar(2, 0, 1), sec(1, 1)), grid)
		address, position := m.MustAddressAt(grid, sec(11, 2), 3, rounding.HalfUp)
		assert.True(t, measure.MustNewAddress(2, pos(1, 4)).Equal(address))
		checkPositionResult(t, position, "5/4", false, 0, "0/1")
	})

	t.Run("initial gradual", func(t *testing.T) {
		t.Parallel()
		m := tempomap.MustNewMeasured(grid, quarter(60)).
			MustWithChange(tempomap.MustNewGradual(bar(1, 0, 1), tempomap.LinearSpeed).MustWithSlot(tempomap.SlotSecond), grid).
			MustWithChange(tempomap.MustNewMain(bar(2, 0, 1), quarter(120)), grid).
			MustWithExtension(tempomap.MustNewExtension(bar(1, 0, 1), sec(1, 1)), grid)
		address, position := m.MustAddressAt(grid, sec(2, 1), 3, rounding.HalfUp)
		expectedAddress := measure.MustNewAddress(1, pos(284, 1000))
		assert.True(t, expectedAddress.Equal(address))
		checkPositionResult(t, position, "284/1000", true, 3, "1/2000")
	})
}

func TestRegressionUnfinishedGradualRejectsInverseAtAndAfterStart(t *testing.T) {
	t.Parallel()
	m := tempomap.MustNew(quarter(60)).
		MustWithChange(tempomap.MustNewGradual(at(1, 1), tempomap.LinearSpeed)).
		MustWithExtension(tempomap.MustNewExtension(at(1, 1), sec(1, 1)))
	assert.True(t, m.MustUnfinished())
	checkPositionResult(t, m.MustPosition(sec(7, 2), 3, rounding.HalfUp), "7/8", false, 0, "0/1")

	for _, seconds := range []tempo.Seconds{
		sec(4, 1), // Start(1), the unfinished gradual boundary.
		sec(9, 2), // Inside the extension at the same boundary.
		sec(5, 1), // End(1), still inside the inclusive extension interval.
		sec(6, 1), // After the boundary and its extension.
	} {
		position, err := m.Position(seconds, 3, rounding.HalfUp)
		assert.False(t, position.IsSet())
		requireTempomapError(t, err, tempomap.ErrUnfinished)
	}
}

func TestRegressionMeasuredUnfinishedGradualRejectsAddressAt(t *testing.T) {
	t.Parallel()
	grid := measure.MustNew(meter.FourFour())
	m := tempomap.MustNewMeasured(grid, quarter(60)).
		MustWithChange(tempomap.MustNewGradual(bar(2, 0, 1), tempomap.LinearSpeed), grid).
		MustWithExtension(tempomap.MustNewExtension(bar(2, 0, 1), sec(1, 1)), grid)

	address, position := m.MustAddressAt(grid, sec(7, 2), 3, rounding.HalfUp)
	assert.True(t, measure.MustNewAddress(1, pos(7, 8)).Equal(address))
	checkPositionResult(t, position, "7/8", false, 0, "0/1")

	for _, seconds := range []tempo.Seconds{sec(4, 1), sec(9, 2), sec(5, 1), sec(6, 1)} {
		address, position, err := m.AddressAt(grid, seconds, 3, rounding.HalfUp)
		assert.False(t, address.IsSet())
		assert.False(t, position.IsSet())
		requireTempomapError(t, err, tempomap.ErrUnfinished)
	}
}

func TestRegressionInitialUnfinishedGradualRejectsZeroInverse(t *testing.T) {
	t.Parallel()
	m := tempomap.MustNew(quarter(60)).MustWithChange(
		tempomap.MustNewGradual(tempomap.MustAt(duration.Origin()), tempomap.LinearSpeed).MustWithSlot(tempomap.SlotSecond),
	)
	position, err := m.Position(tempo.ZeroSeconds(), 3, rounding.HalfUp)
	assert.False(t, position.IsSet())
	requireTempomapError(t, err, tempomap.ErrUnfinished)

	grid := measure.MustNew(meter.FourFour())
	measured := tempomap.MustNewMeasured(grid, quarter(60)).MustWithChange(
		tempomap.MustNewGradual(bar(1, 0, 1), tempomap.LinearSpeed).MustWithSlot(tempomap.SlotSecond), grid,
	)
	address, position, err := measured.AddressAt(grid, tempo.ZeroSeconds(), 3, rounding.HalfUp)
	assert.False(t, address.IsSet())
	assert.False(t, position.IsSet())
	requireTempomapError(t, err, tempomap.ErrUnfinished)
}
