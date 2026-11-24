package tempomap_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tempomap"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKP39InputsAndReturnedSlicesAreIndependent(t *testing.T) {
	t.Parallel()
	grid := grid34Pickup()
	base := tempomap.MustNewMeasured(grid, quarter(60))
	beforeMap := base.String()
	beforeGrid := grid.String()
	change := tempomap.MustNewJump(bar(1, 1, 2), quarter(120))
	beforeChange := change.String()
	changed := base.MustWithChange(change, grid)
	assert.Equal(t, beforeMap, base.String())
	assert.Equal(t, beforeGrid, grid.String())
	assert.Equal(t, beforeChange, change.String())

	rejected, err := base.WithChange(tempomap.MustNewATempo(bar(1, 1, 4)), grid)
	assert.False(t, rejected.IsSet())
	require.Error(t, err)
	assert.Equal(t, beforeMap, base.String())
	assert.Equal(t, beforeGrid, grid.String())

	changes := changed.MustChanges()
	require.NotEmpty(t, changes)
	changes[0] = tempomap.Change{}
	freshChanges := changed.MustChanges()
	assert.True(t, freshChanges[0].IsSet())
	assert.NotEqual(t, changes[0], freshChanges[0])

	withHold := changed.MustWithExtension(tempomap.MustNewExtension(bar(1, 1, 4), sec(1, 1)), grid)
	extensions := withHold.MustExtensions()
	require.NotEmpty(t, extensions)
	extensions[0] = tempomap.Extension{}
	assert.True(t, withHold.MustExtensions()[0].IsSet())

	segments := withHold.MustSegments(grid)
	require.NotEmpty(t, segments)
	segments[0] = tempomap.Segment{}
	assert.True(t, withHold.MustSegments(grid)[0].IsSet())

	marshaled, err := withHold.MarshalText()
	require.NoError(t, err)
	canonical := withHold.String()
	for i := range marshaled {
		marshaled[i] = 'x'
	}
	assert.Equal(t, canonical, withHold.String())
	fresh, err := withHold.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, canonical, string(fresh))

	prefix := make([]byte, len("prefix"), 128)
	copy(prefix, "prefix")
	appended, err := withHold.AppendText(prefix)
	require.NoError(t, err)
	assert.Equal(t, "prefix", string(prefix))
	assert.Equal(t, "prefix"+canonical, string(appended))

	assert.Equal(t, canonical, withHold.String())
	assert.True(t, changed.MustWithoutExtension(bar(100, 0, 1), grid).Equal(changed))
}

func TestConcurrentReadOnlyOperations(t *testing.T) {
	m := tempomap.MustNew(quarter(60)).
		MustWithChange(tempomap.MustNewGradual(tempomap.MustAt(duration.Origin()), tempomap.LinearSpeed).MustWithSlot(tempomap.SlotSecond)).
		MustWithChange(tempomap.MustNewMain(at(1, 1), quarter(120))).
		MustWithChange(tempomap.MustNewJump(at(2, 1), quarter(90))).
		MustWithExtension(tempomap.MustNewExtension(at(1, 2), sec(1, 4)))
	canonical := m.String()

	// Several workers released together are enough for the race detector to
	// see overlapping access to the shared map; each iteration is costly.
	const workers = 8
	const iterations = 10
	var wg sync.WaitGroup
	errorsCh := make(chan error, workers)
	start := make(chan struct{})
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for range iterations {
				if got := m.String(); got != canonical {
					errorsCh <- fmt.Errorf("String changed: %q", got)
					return
				}
				if !m.MustStart(at(3, 2), 6, rounding.HalfUp).IsSet() {
					errorsCh <- fmt.Errorf("Start returned an unset result")
					return
				}
				if !m.MustEnd(at(3, 2), 6, rounding.HalfUp).IsSet() {
					errorsCh <- fmt.Errorf("End returned an unset result")
					return
				}
				if !m.MustPosition(sec(2, 1), 6, rounding.HalfUp).IsSet() {
					errorsCh <- fmt.Errorf("Position returned an unset result")
					return
				}
				if !m.MustSpeed(at(1, 2)).IsValid() {
					errorsCh <- fmt.Errorf("Speed returned an invalid fraction")
					return
				}
				if len(m.MustChanges()) != 4 || len(m.MustExtensions()) != 1 || len(m.MustSegments()) != 3 {
					errorsCh <- fmt.Errorf("unexpected record or segment count")
					return
				}
				if !m.MustEqualTimeline(m) {
					errorsCh <- fmt.Errorf("map is not equal to its own timeline")
					return
				}
			}
		}()
	}
	close(start)
	wg.Wait()
	close(errorsCh)
	for err := range errorsCh {
		assert.NoError(t, err)
	}
}

func TestConcurrentIndependentUnmarshalReceivers(t *testing.T) {
	t.Parallel()
	const text = "tempomap[first@p:0/1:main:4=60]"
	var wg sync.WaitGroup
	results := make([]tempomap.Map, 16)
	errorsCh := make(chan error, len(results))
	for i := range results {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := results[i].UnmarshalText([]byte(text)); err != nil {
				errorsCh <- err
			}
		}()
	}
	wg.Wait()
	close(errorsCh)
	for err := range errorsCh {
		assert.NoError(t, err)
	}
	for _, result := range results {
		assert.Equal(t, text, result.String())
	}
}
