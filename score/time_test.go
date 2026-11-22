package score_test

import (
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/score"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempomap"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// K-6, K-7: both attack and release use the end side of one shared hold.
func TestSharedPerformanceClock(t *testing.T) {
	t.Parallel()
	t.Run("tempo change and inverse", func(t *testing.T) {
		clock := clock60().MustWithChange(tempomap.MustNewMain(at(1, 1), bpm(120)))
		a, b := emptyPart(4, 4, 2), emptyPart(4, 4, 2)
		a = write(a, 1, tone(a, "C4", 1, 1, event.Marks{}))
		a = write(a, 2, tone(a, "D4", 1, 1, event.Marks{}))
		b = write(b, 2, tone(b, "E4", 1, 1, event.Marks{}))
		s, _ := assemble(clock, a, b)
		got, err := s.SoundSeconds(3, rounding.HalfUp)
		require.NoError(t, err)
		require.Len(t, got, 3)
		assertSeconds(t, got[0].Start, 0, 1)
		assertSeconds(t, got[0].End, 4, 1)
		for _, r := range got[1:] {
			assertSeconds(t, r.Start, 4, 1)
			assertSeconds(t, r.End, 6, 1)
		}
		for _, tc := range []struct{ seconds, n, d int }{{4, 1, 1}, {6, 2, 1}, {8, 3, 1}} {
			r := s.MustMap().MustPosition(tempo.MustNewSeconds(tc.seconds, 1), 3, rounding.HalfUp)
			assertPos(t, pos(tc.n, tc.d), r.MustValue())
		}
		_, err = s.At(pos(3, 1))
		require.ErrorIs(t, err, score.ErrPosition)
	})
	for _, seconds := range []int{0, 1} {
		t.Run(string(rune('0'+seconds))+" second hold", func(t *testing.T) {
			a, b, c := emptyPart(4, 4, 1), emptyPart(4, 4, 1), emptyPart(4, 4, 1)
			a = write(a, 1, tone(a, "C4", 1, 2, event.Marks{Fermata: true}))
			b = write(b, 1, event.MustNew(event.RestContent(), event.MustValueDuration(fraction.MustNew(1, 2)), event.Marks{Fermata: true}), tone(b, "D4", 1, 2, event.Marks{}))
			c = write(c, 1, tone(c, "E4", 1, 1, event.Marks{}))
			clock := clock60()
			if seconds != 0 {
				clock = clock.MustWithExtension(tempomap.MustNewExtension(at(1, 2), tempo.MustNewSeconds(seconds, 1)))
			}
			s, ids := assemble(clock, a, b, c)
			got := s.MustSoundSeconds(3, rounding.HalfUp)
			require.Len(t, got, 3)
			assert.Equal(t, ids[0], got[0].Sound.Part)
			assertSeconds(t, got[0].End, 2+seconds, 1)
			assert.Equal(t, ids[2], got[1].Sound.Part)
			assertSeconds(t, got[1].End, 4+seconds, 1)
			assert.Equal(t, ids[1], got[2].Sound.Part)
			assertSeconds(t, got[2].Start, 2+seconds, 1)
			if seconds != 0 {
				assertPos(t, pos(1, 2), s.MustMap().MustPosition(tempo.MustNewSeconds(5, 2), 3, rounding.HalfUp).MustValue())
			}
		})
	}
}

// K-15: no partial result when a later sound cannot be timed.
func TestTimingFailuresAndPrecision(t *testing.T) {
	t.Parallel()
	p := emptyPart(4, 4, 2)
	p = write(p, 1, tone(p, "C4", 1, 4, event.Marks{}))
	p = write(p, 2, tone(p, "D4", 1, 4, event.Marks{}))
	unfinished := clock60().MustWithChange(tempomap.MustNewGradual(at(1, 2), tempomap.LinearSpeed))
	s, _ := assemble(unfinished, p)
	got, err := s.SoundSeconds(3, rounding.HalfUp)
	require.ErrorIs(t, err, tempomap.ErrUnfinished)
	assert.Nil(t, got)
	empty := score.MustNew(unfinished)
	assert.Empty(t, empty.MustSoundSeconds(77, rounding.HalfUp))
	for _, tc := range []struct {
		name   string
		places int
		mode   rounding.Mode
		cause  error
	}{
		{"negative precision", -1, rounding.HalfUp, tempomap.ErrInvalidArgument},
		{"excess precision", 1001, rounding.HalfUp, tempomap.ErrInvalidArgument},
		{"unset mode", 3, rounding.Mode(0), rounding.ErrUnset},
		{"unknown mode", 3, rounding.Mode(255), rounding.ErrUnknownMode},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, e := empty.SoundSeconds(tc.places, tc.mode)
			require.ErrorIs(t, e, tc.cause)
			assert.Nil(t, out)
		})
	}
	linear := clock60().MustWithChange(tempomap.MustNewGradual(at(0, 1), tempomap.LinearSpeed).MustWithSlot(tempomap.SlotSecond)).MustWithChange(tempomap.MustNewMain(at(1, 1), bpm(120)))
	p = emptyPart(4, 4, 1)
	p = write(p, 1, tone(p, "C4", 1, 1, event.Marks{}))
	s, _ = assemble(linear, p)
	timed := s.MustSoundSeconds(3, rounding.HalfUp)
	require.Len(t, timed, 1)
	assert.True(t, timed[0].End.MustApproximate())
	assert.True(t, timed[0].End.MustValue().Equal(tempo.MustNewSeconds(2773, 1000)))
	assert.True(t, timed[0].End.MustErrorBound().Equal(tempo.MustNewSeconds(1, 2000)))
	places, approx := timed[0].End.MustPlaces()
	assert.True(t, approx)
	assert.Equal(t, 3, places)
	down := s.MustSoundSeconds(3, rounding.Down)
	assert.True(t, down[0].End.MustValue().Equal(tempo.MustNewSeconds(693, 250)))
	got, err = s.SoundSeconds(77, rounding.HalfUp)
	require.ErrorIs(t, err, score.ErrOutOfRange)
	require.ErrorIs(t, err, fraction.ErrOutOfRange)
	assert.Nil(t, got)
	// An exact result remains exact at a precision unused by approximation.
	assertSeconds(t, s.MustWithMap(clock60()).MustSoundSeconds(1000, rounding.HalfUp)[0].End, 4, 1)
	// M+4 seconds is outside the exact range, though all map records are valid.
	maximum := tempo.MustSecondsFromFraction(fraction.MustParse(fraction.MaxComponentString))
	longHold := clock60().MustWithExtension(tempomap.MustNewExtension(at(0, 1), maximum))
	got, err = s.MustWithMap(longHold).SoundSeconds(3, rounding.HalfUp)
	require.ErrorIs(t, err, score.ErrOutOfRange)
	require.ErrorIs(t, err, tempomap.ErrOutOfRange)
	require.ErrorIs(t, err, fraction.ErrOutOfRange)
	assert.Nil(t, got)
}

func TestInitialAndTerminalHolds(t *testing.T) {
	t.Parallel()
	p := emptyPart(4, 4, 1)
	p = write(p, 1, tone(p, "C4", 1, 1, event.Marks{}))
	clock := clock60().MustWithExtension(tempomap.MustNewExtension(at(0, 1), tempo.MustNewSeconds(1, 1))).MustWithExtension(tempomap.MustNewExtension(at(1, 1), tempo.MustNewSeconds(2, 1)))
	s, _ := assemble(clock, p)
	got := s.MustSoundSeconds(0, rounding.HalfEven)
	require.Len(t, got, 1)
	assertSeconds(t, got[0].Start, 1, 1)
	assertSeconds(t, got[0].End, 7, 1)
	assertPos(t, duration.Origin(), s.MustMap().MustPosition(tempo.MustNewSeconds(1, 2), 3, rounding.HalfUp).MustValue())
}
