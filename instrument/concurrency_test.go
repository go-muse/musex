package instrument_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/go-muse/musex/instrument"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The tests of this file check К-24 and П-15 under the race detector. Worker
// goroutines never call require or t.FailNow: they count violations in atomic
// counters, and the test goroutine checks the counters after Wait. Readers run
// until the writers have finished; nothing waits on time.

const (
	writerRounds = 200
	readerCount  = 4
)

// race runs the readers in a loop while the writers make their rounds, and
// returns when every goroutine has finished. Each reader runs at least once.
func race(rounds int, writers []func(round int), readers ...func()) {
	var done atomic.Bool
	var readersWG, writersWG sync.WaitGroup
	for _, read := range readers {
		for range readerCount {
			readersWG.Go(func() {
				for {
					read()
					if done.Load() {
						return
					}
				}
			})
		}
	}
	for _, write := range writers {
		writersWG.Go(func() {
			for round := range rounds {
				write(round)
			}
		})
	}
	writersWG.Wait()
	done.Store(true)
	readersWG.Wait()
}

// К-24: readers of the source during alternating retunings Н-0 ↔ Н-415 get only
// sets equal to Н-0 or Н-415.
func TestConcurrency_readersSeeWholeSets(t *testing.T) {
	t.Parallel()
	n0, n415 := h0(), h415()
	g := instrument.MustFromParts("Lead", h0(), rng("C4", "G5"))
	var src settings.Source = g
	var violations, writeErrors, reads atomic.Int64

	race(writerRounds,
		[]func(int){func(round int) {
			next := n415
			if round%2 == 1 {
				next = n0
			}
			if g.Retune(next) != nil {
				writeErrors.Add(1)
			}
		}},
		func() {
			reads.Add(1)
			s := src.Settings()
			if !s.Equal(n0) && !s.Equal(n415) {
				violations.Add(1)
			}
		},
	)

	assert.Zero(t, violations.Load(), "a set that is neither Н-0 nor Н-415")
	assert.Zero(t, writeErrors.Load())
	assert.Positive(t, reads.Load())
	assert.True(t, g.Settings().Equal(n0), "an even number of rounds ends with Н-0")
}

// К-24: concurrent RetuneFunc calls whose transformations depend on the
// current set lose no change.
func TestConcurrency_RetuneFuncLosesNothing(t *testing.T) {
	t.Parallel()
	const goroutines, perGoroutine = 16, 8
	g := instrument.MustFromParts("Lead", h0(), rng("C4", "G5"))
	var failures atomic.Int64
	var wg sync.WaitGroup
	for range goroutines {
		wg.Go(func() {
			for range perGoroutine {
				err := g.RetuneFunc(func(s *settings.Settings) (*settings.Settings, error) {
					next, err := s.Tuning().WithReferenceFrequency(s.Tuning().Reference().Frequency() + 1)
					if err != nil {
						return nil, err
					}
					return s.WithTuning(next)
				})
				if err != nil {
					failures.Add(1)
				}
			}
		})
	}
	wg.Wait()

	require.Zero(t, failures.Load())
	assert.InDelta(t, 440.0+goroutines*perGoroutine, g.Settings().Tuning().Reference().Frequency(), 1e-9)
}

// К-24: the range operations of an instrument during retunings and changes of
// the range fail only for causes that exist in some state. Every state here
// has a valid range, so no operation fails.
func TestConcurrency_rangeOperations(t *testing.T) {
	t.Parallel()
	n0, n415 := h0(), h415()
	upper, lower := rng("C4", "G5"), rng("C3", "G4") // both 19 steps wide
	g := instrument.MustFromParts("Lead", h0(), upper)
	x := p("E4") // inside both ranges
	var violations, writeErrors atomic.Int64

	race(writerRounds,
		[]func(int){
			func(round int) {
				next := n415
				if round%2 == 1 {
					next = n0
				}
				if g.Retune(next) != nil {
					writeErrors.Add(1)
				}
			},
			func(round int) {
				next := lower
				if round%2 == 1 {
					next = upper
				}
				if g.SetRange(next) != nil {
					writeErrors.Add(1)
				}
			},
		},
		func() {
			if in, err := g.Contains(x); err != nil || !in {
				violations.Add(1)
			}
			if loc, err := g.Locate(x); err != nil || loc != instrument.Inside {
				violations.Add(1)
			}
			if w, err := g.Width(); err != nil || w != 19 {
				violations.Add(1)
			}
			if c, err := g.Count(); err != nil || c != 20 {
				violations.Add(1)
			}
			if f, ok, err := g.Fold(p("E1")); err != nil || !ok || !f.SameSpelling(x) {
				violations.Add(1)
			}
			if seq, err := g.Folds(x); err != nil {
				violations.Add(1)
			} else {
				n := 0
				for range seq {
					n++
				}
				if n != 2 {
					violations.Add(1)
				}
			}
			if seq, err := g.Pitches(system.Sharps()); err != nil {
				violations.Add(1)
			} else {
				n := 0
				for range seq {
					n++
				}
				if n != 20 {
					violations.Add(1)
				}
			}
		},
	)

	assert.Zero(t, violations.Load(), "an answer that no state gives")
	assert.Zero(t, writeErrors.Load())
}

// matchesOne reports whether a state is valid (П-9) and equal to one of the
// allowed states.
func matchesOne(st *instrument.Preset, allowed []*instrument.Preset) bool {
	if !st.IsValid() || st.Name() == "" || !st.SettingsSet().IsValid() {
		return false
	}
	if st.HasRange() && st.Range().Check(st.SettingsSet().System()) != nil {
		return false
	}
	for _, a := range allowed {
		if a.Equal(st) {
			return true
		}
	}
	return false
}

// К-24: the state of a clone and the state as a preset, taken during the chain
// "remove the range → retune to Н-19 → set the range" and back, are always one
// of the states of the original and are valid.
func TestConcurrency_chainStates(t *testing.T) {
	t.Parallel()
	allowed := []*instrument.Preset{
		preset("Lead", h0(), rangeKP6()),
		preset("Lead", h0(), instrument.Range{}),
		preset("Lead", h19(), instrument.Range{}),
		preset("Lead", h19(), rng("C4", "G5")),
	}
	n0, n19 := h0(), h19()
	g := instrument.MustFromParts("Lead", h0(), rangeKP6())
	var violations, writeErrors atomic.Int64

	race(writerRounds,
		[]func(int){func(round int) {
			var err error
			if round%2 == 0 {
				err = g.RemoveRange()
				if err == nil {
					err = g.Retune(n19)
				}
				if err == nil {
					err = g.SetRange(rng("C4", "G5"))
				}
			} else {
				err = g.RemoveRange()
				if err == nil {
					err = g.Retune(n0)
				}
				if err == nil {
					err = g.SetRange(rangeKP6())
				}
			}
			if err != nil {
				writeErrors.Add(1)
			}
		}},
		func() {
			st, err := g.State()
			if err != nil || !matchesOne(st, allowed) {
				violations.Add(1)
			}
		},
		func() {
			c, err := g.Clone()
			if err != nil {
				violations.Add(1)
				return
			}
			st, err := c.State()
			if err != nil || !matchesOne(st, allowed) {
				violations.Add(1)
			}
		},
	)

	assert.Zero(t, violations.Load(), "a state that the original never had")
	assert.Zero(t, writeErrors.Load())
	assertState(t, allowed[0], g)
}

// К-24, П-12а: during alternating RetuneWithRange between (Н-0, B♯3–C♯4) and
// (Н-19, C4–G5) the state as a preset is only one of these two pairs.
func TestConcurrency_RetuneWithRange(t *testing.T) {
	t.Parallel()
	allowed := []*instrument.Preset{
		preset("Lead", h0(), rangeKP6()),
		preset("Lead", h19(), rng("C4", "G5")),
	}
	n0, n19 := h0(), h19()
	g := instrument.MustFromParts("Lead", h0(), rangeKP6())
	var violations, writeErrors atomic.Int64

	race(writerRounds,
		[]func(int){func(round int) {
			var err error
			if round%2 == 0 {
				err = g.RetuneWithRange(n19, rng("C4", "G5"))
			} else {
				err = g.RetuneWithRange(n0, rangeKP6())
			}
			if err != nil {
				writeErrors.Add(1)
			}
		}},
		func() {
			st, err := g.State()
			if err != nil || !matchesOne(st, allowed) {
				violations.Add(1)
			}
		},
		func() {
			// A range operation reads one state: in both states C4 is inside.
			if in, err := g.Contains(p("C4")); err != nil || !in {
				violations.Add(1)
			}
		},
	)

	assert.Zero(t, violations.Load(), "a set with the range of the other state")
	assert.Zero(t, writeErrors.Load())
	assertState(t, allowed[0], g)
}

// П-15, А 4.8.1: all methods can be called at once from different goroutines;
// readers of other instruments and failing changes do not disturb anything.
func TestConcurrency_mixedOperations(t *testing.T) {
	t.Parallel()
	g := instrument.MustFromParts("Lead", h0(), rng("C4", "G5"))
	other := instrument.MustFromParts("Other", h415(), instrument.Range{})
	var violations atomic.Int64

	race(writerRounds,
		[]func(int){
			func(round int) {
				if round%2 == 0 {
					_ = g.SetName("Second")
				} else {
					_ = g.SetName("Lead")
				}
			},
			func(int) {
				if g.Retune(hDe()) == nil { // always incompatible
					violations.Add(1)
				}
				if g.SetName("") == nil {
					violations.Add(1)
				}
			},
			func(int) {
				_ = other.RetuneFunc(func(s *settings.Settings) (*settings.Settings, error) {
					_ = g.Settings() // reading another instrument inside a transformation is allowed
					return s, nil
				})
			},
		},
		func() {
			name := g.Name()
			if name != "Lead" && name != "Second" {
				violations.Add(1)
			}
			if !g.Equal(g) || !g.Same(g) || g.Equal(other) {
				violations.Add(1)
			}
			if g.CheckRetune(hDe()) == nil {
				violations.Add(1)
			}
			_ = g.String()
			if !other.Settings().Equal(h415()) {
				violations.Add(1)
			}
		},
	)

	assert.Zero(t, violations.Load())
	assertState(t, preset("Lead", h0(), rng("C4", "G5")), g)
}
