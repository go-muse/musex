package voice_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/voice"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// НФ-3, architecture 4.3: any methods of a line may be called from several
// goroutines at once. Run with -race (make test).
func TestConcurrentUse(t *testing.T) {
	t.Parallel()
	shared := named(line(snd(c4(), qe()), rst(e8()).MustWithFermata(true), snd(e4(), q()), snd(g4(), h())))
	ref := voice.InstrumentReference(sopranoInstrument())
	snapshot := named(line(snd(c4(), qe()), rst(e8()).MustWithFermata(true), snd(e4(), q()), snd(g4(), h())))

	wantTessitura := snapshot.MustTessitura(ref)
	wantIntervals := intervalStrings(snapshot.MustIntervals())
	wantTransposed := snapshot.MustTranspose(interval.MajorSecond())
	wantRetrograde := snapshot.MustRetrograde(voice.RetrogradeBoth)
	wantTimings := snapshot.MustSoundSeconds(map60(), 3, rounding.HalfUp)

	checks := []func() error{
		func() error {
			if got := shared.MustElements(); len(got) != 4 || !got[0].Equal(snd(c4(), qe())) {
				return fmt.Errorf("elements: %v", got)
			}
			return nil
		},
		func() error {
			if got := shared.MustNotes(); len(got) != 3 || !got[2].Equal(g4()) {
				return fmt.Errorf("notes: %v", got)
			}
			return nil
		},
		func() error {
			starts, end := shared.MustStarts()
			if len(starts) != 4 || end.String() != "5/4" {
				return fmt.Errorf("starts: %v %v", starts, end)
			}
			return nil
		},
		func() error {
			if i, ok := shared.MustAt(pos(1, 2)); !ok || i != 2 {
				return fmt.Errorf("at: %d %v", i, ok)
			}
			return nil
		},
		func() error {
			if got := intervalStrings(shared.MustIntervals()); fmt.Sprint(got) != fmt.Sprint(wantIntervals) {
				return fmt.Errorf("intervals: %v", got)
			}
			return nil
		},
		func() error {
			if got := shared.MustAmbitus().String(); got != "P5" {
				return fmt.Errorf("ambitus: %v", got)
			}
			return nil
		},
		func() error {
			if got := shared.MustTessitura(ref); got.Kind != wantTessitura.Kind || !got.Number.Equal(wantTessitura.Number) {
				return fmt.Errorf("tessitura: %v", got)
			}
			return nil
		},
		func() error {
			if got := shared.MustTranspose(interval.MajorSecond()); !got.Equal(wantTransposed) {
				return fmt.Errorf("transpose: %v", got)
			}
			return nil
		},
		func() error {
			if got := shared.MustRetrograde(voice.RetrogradeBoth); !got.Equal(wantRetrograde) {
				return fmt.Errorf("retrograde: %v", got)
			}
			return nil
		},
		func() error {
			if got := voice.MustFromEvents(shared.MustEvents()...); !got.EqualElements(snapshot) {
				return fmt.Errorf("events: %v", got)
			}
			return nil
		},
		func() error {
			got := shared.MustSoundSeconds(map60(), 3, rounding.HalfUp)
			if len(got) != len(wantTimings) || !got[2].End.Equal(wantTimings[2].End) {
				return fmt.Errorf("seconds: %v", got)
			}
			return nil
		},
		func() error {
			if got := shared.MustAppend(rst(q())).MustLen(); got != 5 {
				return fmt.Errorf("append: %d", got)
			}
			return nil
		},
		func() error {
			if !shared.Equal(snapshot) || shared.String() == "" {
				return fmt.Errorf("equal: %v", shared)
			}
			return nil
		},
	}

	const workers = 8
	failures := make(chan error, workers*len(checks)*4)
	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			for range 4 {
				for _, check := range checks {
					if err := safely(check); err != nil {
						failures <- err
					}
				}
			}
		})
	}
	wg.Wait()
	close(failures)
	for err := range failures {
		assert.NoError(t, err)
	}
	require.True(t, shared.Equal(snapshot), "the line is unchanged")
}

// safely runs check and turns a panic into an error: a failing Must in a
// worker must not stop the test binary.
func safely(check func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	return check()
}
