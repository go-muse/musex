package midi_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/go-muse/musex/midi"
	"github.com/go-muse/musex/notation"
	"github.com/go-muse/musex/octave/naming"
	octavenumber "github.com/go-muse/musex/octave/number"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"
	"github.com/go-muse/musex/tuning/temperament"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func requirePanicError(t *testing.T, expected error, operation func()) {
	t.Helper()
	defer func() {
		recovered := recover()
		require.NotNil(t, recovered, "Must variant must panic")
		assert.Equal(t, expected, recovered, "Must variant must panic with the same error value")
	}()
	operation()
}

func assertMidiError(t *testing.T, err error, expected error) {
	t.Helper()
	require.Error(t, err)
	sentinels := []error{
		midi.ErrInvalidNumber,
		midi.ErrUnset,
		midi.ErrNoOctave,
		midi.ErrNoSource,
		midi.ErrUnsetSettings,
		midi.ErrNotTraditionalC,
		midi.ErrInadmissible,
		midi.ErrOutOfRange,
		midi.ErrInvalidClass,
		midi.ErrNoSpellingRule,
		midi.ErrNotApplicable,
	}
	for _, sentinel := range sentinels {
		assert.Equal(t, sentinel == expected, errors.Is(err, sentinel),
			"unexpected MIDI sentinel match for %v", sentinel)
	}
}

func requireNumberValue(t *testing.T, expected int, number midi.Number) {
	t.Helper()
	require.True(t, number.IsValid(), "MIDI %d must be valid", expected)
	actual, err := number.Int()
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func settingsForSystem(sys *system.System) *settings.Settings {
	temp := temperament.Equal12()
	if sys.Classes() == 19 {
		temp = temperament.Equal19()
	}
	tun := tuning.MustNew(sys, temp, tuning.RootFromName(spelling.C()), tuning.A4At440())
	return settings.MustNew(
		tun,
		notation.MustBuiltIn(notation.EnglishUnicodeID),
		octavenumber.ASPN(),
		naming.English(),
	)
}

type countingSource struct {
	mu        sync.RWMutex
	current   *settings.Settings
	calls     int
	snapshots []*settings.Settings
}

func newCountingSource(initial *settings.Settings) *countingSource {
	return &countingSource{current: initial}
}

func (s *countingSource) Settings() *settings.Settings {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	s.snapshots = append(s.snapshots, s.current)
	return s.current
}

func (s *countingSource) Replace(next *settings.Settings) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.current = next
}

func (s *countingSource) State() (int, []*settings.Settings) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.calls, append([]*settings.Settings(nil), s.snapshots...)
}
