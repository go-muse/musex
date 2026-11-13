package tuning_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"

	"github.com/stretchr/testify/require"
)

func TestConcurrentSharedValues(t *testing.T) {
	t.Parallel()
	std := tuning.Standard()
	pyth := newPythagoreanTuning(t, tuning.RootFromClass(pitchclass.MustNew(0, 12)))
	sharedReference := tuning.A4At440()
	sharedRoot := tuning.RootFromClass(pitchclass.MustNew(0, 12))
	wantNearestDeviation := cents.MustBetween(440, 445)

	const (
		workers    = 12
		iterations = 25
	)
	errs := make(chan error, workers*iterations)
	var wg sync.WaitGroup
	for worker := range workers {
		wg.Go(func() {
			for range iterations {
				if got := std.MustFrequency(aspn(spelling.A(), 4)); got != 440 {
					errs <- fmt.Errorf("worker %d: A4 = %v", worker, got)
					return
				}
				if got := std.MustFrequencyOfStep(12); got != 880 {
					errs <- fmt.Errorf("worker %d: step 12 = %v", worker, got)
					return
				}
				step, deviation, err := std.Nearest(445)
				if err != nil || step != 0 || !wantNearestDeviation.Equal(deviation) || !cents.MustFromFloat(19.562175).Equal(deviation) {
					errs <- fmt.Errorf("worker %d: nearest = %v, %v, %v", worker, step, deviation, err)
					return
				}
				if got := pyth.MustCentsBetween(aspn(spelling.A(), 4), aspn(spelling.C(), 5)); !cents.MustFromRatio(32, 27).Equal(got) {
					errs <- fmt.Errorf("worker %d: Pythagorean cents = %v", worker, got)
					return
				}
				changed := std.MustWithReferenceFrequency(435).MustWithRoot(tuning.RootFromName(spelling.A()))
				if changed.MustFrequency(aspn(spelling.A(), 4)) != 435 || std.MustFrequency(aspn(spelling.A(), 4)) != 440 {
					errs <- fmt.Errorf("worker %d: With methods mutated an input", worker)
					return
				}
				classes, err := std.ChangedClasses(tuning.RootFromName(spelling.A()))
				if err != nil || classes == nil || len(classes) != 0 {
					errs <- fmt.Errorf("worker %d: changed classes = %v, %v", worker, classes, err)
					return
				}
				if !sharedReference.Equal(tuning.A4At440()) || !sharedRoot.Equal(tuning.RootFromClass(pitchclass.MustNew(0, 12))) {
					errs <- fmt.Errorf("worker %d: shared values changed", worker)
					return
				}
				refs := tuning.References()
				if len(refs) != 8 || !refs[0].Equal(sharedReference) {
					errs <- fmt.Errorf("worker %d: reference catalog = %v", worker, refs)
					return
				}
				refs[0] = tuning.Reference{}
				if !tuning.A4At440().Equal(sharedReference) {
					errs <- fmt.Errorf("worker %d: catalog result aliases shared state", worker)
					return
				}
				fresh := tuning.Standard()
				if !fresh.Equal(std) || fresh == std {
					errs <- fmt.Errorf("worker %d: Standard returns invalid shared state", worker)
					return
				}
				if _, err := std.PitchOfStep(1, system.Sharps()); err != nil {
					errs <- fmt.Errorf("worker %d: PitchOfStep: %w", worker, err)
					return
				}
			}
		})
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}
}
