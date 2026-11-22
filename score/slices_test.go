package score_test

import (
	"testing"

	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/score"
	"github.com/go-muse/musex/track"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// K-3: the boundary belongs to the next written event.
func TestVerticalSlices(t *testing.T) {
	t.Parallel()
	a, b := emptyPart(4, 4, 1), emptyPart(4, 4, 1)
	ac, ad := tone(a, "C4", 1, 2, event.Marks{}), tone(a, "D4", 1, 2, event.Marks{})
	a = write(a, 1, ac, ad)
	be, bf, bg, ba := tone(b, "E4", 1, 4, event.Marks{}), tone(b, "F4", 1, 4, event.Marks{}), tone(b, "G4", 1, 4, event.Marks{}), tone(b, "A4", 1, 4, event.Marks{})
	b = write(b, 1, be, bf, bg, ba)
	s, ids := assemble(clock60(), a, b)
	for _, tc := range []struct {
		name   string
		n, d   int
		a, b   *event.Event
		ai, bi int
	}{
		{"quarter", 1, 4, ac, bf, 0, 1}, {"half", 1, 2, ad, bg, 1, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := s.At(pos(tc.n, tc.d), score.Selection{Part: ids[1]}, score.Selection{Part: ids[0]})
			require.NoError(t, err)
			require.Len(t, got, 2)
			assert.Equal(t, ids[0], got[0].Part)
			assert.Equal(t, ids[1], got[1].Part)
			for i, want := range []*event.Event{tc.a, tc.b} {
				require.Len(t, got[i].Voices, 1)
				assert.True(t, got[i].Voices[0].Found)
				assertEvent(t, want, got[i].Voices[0].Entry.Event)
			}
			assert.Equal(t, tc.ai, got[0].Voices[0].Entry.Place.Index)
			assert.Equal(t, tc.bi, got[1].Voices[0].Entry.Place.Index)
		})
	}
}

// K-4: tied notation remains visible while a sound has one sustained span.
func TestTiesDuplicateParticipationsAndVoiceOrdering(t *testing.T) {
	t.Parallel()
	a := emptyPart(4, 4, 1)
	first := tone(a, "C4", 1, 2, event.Marks{Ties: []int{0}})
	second := tone(a, "C4", 1, 2, event.Marks{})
	a = write(a, 1, first, second)
	b := emptyPart(4, 4, 1)
	b = write(b, 1, tone(b, "C4", 1, 1, event.Marks{}))
	s, ids := assemble(clock60(), a, b, a)
	sounds, err := s.Sounds()
	require.NoError(t, err)
	require.Len(t, sounds, 3)
	for i, sound := range sounds {
		assert.Equal(t, ids[i], sound.Part)
		assertPos(t, pos(0, 1), sound.Sound.Time.Start)
		assertPos(t, pos(1, 1), sound.Sound.Time.End)
	}
	require.Len(t, sounds[0].Sound.Parts, 2)
	assert.Equal(t, 0, sounds[0].Sound.Parts[0].Place.Index)
	assert.Equal(t, 1, sounds[0].Sound.Parts[1].Place.Index)
	assert.Equal(t, 0, sounds[0].Sound.Parts[0].NoteIndex)
	slice := s.MustAt(pos(3, 4))
	assertEvent(t, second, slice[0].Voices[0].Entry.Event)
	active := s.MustSoundingAt(pos(3, 4))
	require.Len(t, active, 3)
	assert.Len(t, active[0].Sound.Parts, 2)
	assert.Empty(t, s.MustSoundingAt(pos(1, 1)))
	// A reverse filter does not reverse track or voice order.
	p := emptyPart(4, 4, 1)
	p, v2 := p.MustAddVoice()
	v1 := p.MustVoices()[0]
	p = p.MustSplice(v2, 1, 0, 0, tone(p, "E4", 1, 4, event.Marks{}))
	p = p.MustSplice(v1, 1, 0, 0, tone(p, "D4", 1, 4, event.Marks{}))
	x, xids := assemble(clock60(), p)
	selected := x.MustAt(pos(0, 1), score.Selection{Part: xids[0], Voices: []track.VoiceID{v2, v1}})
	assert.Equal(t, v1, selected[0].Voices[0].Voice)
	assert.Equal(t, v2, selected[0].Voices[1].Voice)
	ordered := x.MustSounds(score.Selection{Part: xids[0], Voices: []track.VoiceID{v2, v1}})
	require.Len(t, ordered, 2)
	assert.Equal(t, v1, ordered[0].Sound.Parts[0].Place.Voice)
	assert.Equal(t, v2, ordered[1].Sound.Parts[0].Place.Voice)
}

// K-5: silence is not an inserted rest, and a short grid has ended.
func TestSilenceRestAndFiniteEnds(t *testing.T) {
	t.Parallel()
	a := emptyPart(4, 4, 1)
	r := event.MustNew(event.RestContent(), event.MustValueDuration(fraction.MustNew(1, 4)), event.Marks{})
	a = write(a, 1, r)
	b := emptyPart(4, 4, 2)
	b = write(b, 2, tone(b, "G4", 1, 1, event.Marks{}))
	s, ids := assemble(clock60(), a, b)
	start := s.MustAt(pos(0, 1))
	assert.True(t, start[0].Voices[0].Found)
	assertEvent(t, r, start[0].Voices[0].Entry.Event)
	assert.False(t, start[1].Voices[0].Found)
	assert.Zero(t, start[1].Voices[0].Entry)
	gap := s.MustAt(pos(1, 2))
	assert.False(t, gap[0].Ended)
	assert.False(t, gap[0].Voices[0].Found)
	short := s.MustAt(pos(3, 2))
	assert.True(t, short[0].Ended)
	assert.Empty(t, short[0].Voices)
	assert.False(t, short[1].Ended)
	assert.True(t, short[1].Voices[0].Found)
	assertPos(t, pos(2, 1), s.MustEnd())
	assert.Empty(t, s.MustAt(pos(2, 1)))
	got, err := s.At(pos(9, 4))
	require.ErrorIs(t, err, score.ErrPosition)
	assert.Nil(t, got)
	assert.Len(t, s.MustPart(ids[0]).MustEntries(a.MustVoices()[0], 1), 1)
	assert.Empty(t, s.MustSounds(score.Selection{Part: ids[0]}))
}

func TestPartialChordTiePreservesOccurrencesAndAttacks(t *testing.T) {
	t.Parallel()
	p := emptyPart(4, 4, 1)
	c := note.MustParsePitch(p.MustInstrument(), "C4")
	e := note.MustParsePitch(p.MustInstrument(), "E4")
	g := note.MustParsePitch(p.MustInstrument(), "G4")
	first := event.MustNew(event.MustChordContent(chord.MustNew(c, c, e)), event.MustValueDuration(fraction.MustNew(1, 2)), event.Marks{Ties: []int{1}, Arpeggio: true})
	second := event.MustNew(event.MustChordContent(chord.MustNew(c, g)), event.MustValueDuration(fraction.MustNew(1, 2)), event.Marks{})
	p = write(p, 1, first, second)
	s, ids := assemble(clock60(), p)
	got := s.MustSounds()
	require.Len(t, got, 4)
	for _, sound := range got {
		assert.Equal(t, ids[0], sound.Part)
	}
	assertPos(t, pos(1, 2), got[0].Sound.Time.End)
	assertPos(t, pos(1, 1), got[1].Sound.Time.End)
	require.Len(t, got[1].Sound.Parts, 2)
	assert.Equal(t, 1, got[1].Sound.Parts[0].NoteIndex)
	assert.Equal(t, 0, got[1].Sound.Parts[1].NoteIndex)
	assertPos(t, pos(1, 2), got[3].Sound.Time.Start)
	assert.Equal(t, 1, got[3].Sound.Parts[0].NoteIndex)
	assert.Len(t, s.MustSoundingAt(pos(3, 4)), 2)
	assertEvent(t, second, s.MustAt(pos(3, 4))[0].Voices[0].Entry.Event)
	assert.True(t, s.MustPart(ids[0]).MustEntries(p.MustVoices()[0], 1)[0].Event.MustMarks().Arpeggio)
}
