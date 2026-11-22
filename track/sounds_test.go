package track_test

import (
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/track"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPartialChordTiesAndSoundOrdering(t *testing.T) {
	t.Parallel()
	p, v := fixture(2)
	p, w := p.MustAddVoice()
	c, e, g := parsed("C4"), parsed("E4"), parsed("G4")
	first := chordEvent(written(duration.Half, 0), event.Marks{Ties: []int{2}, Arpeggio: true}, c, e, g)
	last := tone("G4", duration.Quarter, event.Marks{})
	low := tone("D3", duration.Whole, event.Marks{})
	p = p.MustEdit(track.Edit{Bars: []track.BarEdit{{Voice: v, Bar: 1, Events: []*event.Event{first, last}}, {Voice: w, Bar: 1, Events: []*event.Event{low}}}})
	sounds := p.MustSounds()
	require.Len(t, sounds, 4)
	for i, n := range []note.Note{c, e, g, parsed("D3")} {
		assertNote(t, n, sounds[i].Note)
	}
	assertSpan(t, sounds[0].Time, 0, 1, 1, 2)
	assertSpan(t, sounds[1].Time, 0, 1, 1, 2)
	assertSpan(t, sounds[2].Time, 0, 1, 3, 4)
	assertSpan(t, sounds[3].Time, 0, 1, 1, 1)
	assertOccurrences(t, []track.Occurrence{{Place: track.Place{Voice: v, Bar: 1, Index: 0}, NoteIndex: 2}, {Place: track.Place{Voice: v, Bar: 1, Index: 1}, NoteIndex: 0}}, sounds[2].Parts)
	assertOccurrences(t, []track.Occurrence{{Place: track.Place{Voice: w, Bar: 1, Index: 0}, NoteIndex: 0}}, sounds[3].Parts)
	at, ok := p.MustAt(v, position(1, 2))
	require.True(t, ok)
	assert.Same(t, last, at.Event)
	now := p.MustSoundingAt(position(1, 2), v)
	require.Len(t, now, 1)
	assertSpan(t, now[0].Time, 0, 1, 3, 4)
	assert.Empty(t, p.MustSoundingAt(position(3, 4), v))
	assert.Empty(t, p.MustSoundingAt(position(2, 1)))
	assert.Len(t, p.MustSounds(w, v), 4)
	assert.True(t, w == p.MustSounds(w)[0].Parts[0].Place.Voice)
	duplicate, err := p.Sounds(v, v)
	assert.Error(t, err)
	assert.Empty(t, duplicate)
	_, err = p.SoundingAt(position(1, 4), track.VoiceID{})
	assert.ErrorIs(t, err, track.ErrVoice)
	_, err = p.SoundingAt(position(9, 4))
	issues(t, err, track.ErrPlace)
	_, err = p.SoundingAt(duration.Position{})
	assert.Error(t, err)
	assert.True(t, p.MustEntries(v, 1)[0].Event.MustMarks().Arpeggio)

	// Returned occurrences and notes are owned by this report.
	sounds[2].Parts[0].NoteIndex = 99
	sounds[2].Note.MustSetASPN(7)
	clean := p.MustSounds(v)
	require.Len(t, clean, 3)
	assert.Equal(t, 2, clean[2].Parts[0].NoteIndex)
	assertNote(t, g, clean[2].Note)
}

func TestRepeatedOccurrencesMatchInWrittenOrder(t *testing.T) {
	t.Parallel()
	p, v := fixture(1)
	c := parsed("C4")
	first := chordEvent(written(duration.Quarter, 0), event.Marks{Ties: []int{1}}, c, c)
	second := chordEvent(written(duration.Quarter, 0), event.Marks{}, c, c)
	p = p.MustSplice(v, 1, 0, 0, first, second)
	s := p.MustSounds(v)
	require.Len(t, s, 3)
	assertSpan(t, s[0].Time, 0, 1, 1, 4)
	assertSpan(t, s[1].Time, 0, 1, 1, 2)
	assertSpan(t, s[2].Time, 1, 4, 1, 2)
	assertOccurrences(t, []track.Occurrence{{Place: track.Place{Voice: v, Bar: 1, Index: 0}, NoteIndex: 1}, {Place: track.Place{Voice: v, Bar: 1, Index: 1}, NoteIndex: 0}}, s[1].Parts)
	assert.Equal(t, 1, s[2].Parts[0].NoteIndex)
	all := chordEvent(written(duration.Quarter, 0), event.Marks{Ties: []int{0, 1}}, c, c)
	two := p.MustSplice(v, 1, 0, 1, all).MustSounds(v)
	require.Len(t, two, 2)
	assert.Equal(t, 0, two[0].Parts[1].NoteIndex)
	assert.Equal(t, 1, two[1].Parts[1].NoteIndex)
}

func TestTieValidationIsAtomicAndAddressed(t *testing.T) {
	t.Parallel()
	c := parsed("C4")
	raised := c
	raised.MustSetCents(cents.MustFromInt(1))
	for _, tc := range []struct {
		name      string
		bars      int
		events    []*event.Event
		later     []*event.Event
		cause     error
		noteIndex int
	}{
		{"end of grid", 1, []*event.Event{tone("C4", duration.Whole, event.Marks{Ties: []int{0}})}, nil, track.ErrTieEnd, 0},
		{"empty tail", 2, []*event.Event{tone("C4", duration.Quarter, event.Marks{Ties: []int{0}})}, nil, track.ErrTieGap, 0},
		{"rest", 1, []*event.Event{tone("C4", duration.Quarter, event.Marks{Ties: []int{0}}), rest(written(duration.Quarter, 0), false)}, nil, track.ErrTieRest, 0},
		{"enharmonic spelling", 1, []*event.Event{tone("C♯4", duration.Quarter, event.Marks{Ties: []int{0}}), tone("D♭4", duration.Quarter, event.Marks{})}, nil, track.ErrTiePitch, 0},
		{"different cents", 1, []*event.Event{sounded(c, written(duration.Quarter, 0), event.Marks{Ties: []int{0}}), sounded(raised, written(duration.Quarter, 0), event.Marks{})}, nil, track.ErrTiePitch, 0},
		{"missing duplicate", 1, []*event.Event{chordEvent(written(duration.Quarter, 0), event.Marks{Ties: []int{0, 1}}, c, c), tone("C4", duration.Quarter, event.Marks{})}, nil, track.ErrTieMultiplicity, 1},
		{"interbar gap", 2, []*event.Event{tone("C4", duration.Half, event.Marks{Ties: []int{0}})}, []*event.Event{tone("C4", duration.Quarter, event.Marks{})}, track.ErrTieGap, 0},
		{"missing octave", 1, []*event.Event{nameTone("C", duration.Quarter, event.Marks{Ties: []int{0}}), tone("C4", duration.Quarter, event.Marks{})}, nil, track.ErrTiePitch, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, v := fixture(tc.bars)
			ed := track.Edit{Bars: []track.BarEdit{{Voice: v, Bar: 1, Events: tc.events}}}
			if tc.later != nil {
				ed.Bars = append(ed.Bars, track.BarEdit{Voice: v, Bar: 2, Events: tc.later})
			}
			next, err := p.Edit(ed)
			assert.Nil(t, next)
			is := issues(t, err, tc.cause)
			assertPlace(t, track.Place{Voice: v, Bar: 1, Index: 0}, is[0].Place)
			assert.Equal(t, tc.noteIndex, is[0].NoteIndex)
			assert.Empty(t, p.MustEntries(v, 1))
		})
	}
	p, v := fixture(2)
	tied := nameTone("C", duration.Whole, event.Marks{Ties: []int{0}})
	next := nameTone("C", duration.Quarter, event.Marks{})
	p = p.MustEdit(track.Edit{Bars: []track.BarEdit{{Voice: v, Bar: 1, Events: []*event.Event{tied}}, {Voice: v, Bar: 2, Events: []*event.Event{next}}}})
	s := p.MustSounds()
	require.Len(t, s, 1)
	assert.False(t, s[0].Note.HasOctave())
	assertSpan(t, s[0].Time, 0, 1, 5, 4)
}

func TestInterbarTieCanBePublishedAsOneEdit(t *testing.T) {
	t.Parallel()
	p, v := fixture(2)
	first := tone("G3", duration.Whole, event.Marks{Ties: []int{0}})
	last := tone("G3", duration.Half, event.Marks{})
	bad, err := p.Splice(v, 1, 0, 0, first)
	assert.Nil(t, bad)
	assert.ErrorIs(t, err, track.ErrTieGap)
	p = p.MustEdit(track.Edit{Bars: []track.BarEdit{{Voice: v, Bar: 1, Events: []*event.Event{first}}, {Voice: v, Bar: 2, Events: []*event.Event{last}}}})
	s := p.MustSounds()
	require.Len(t, s, 1)
	assertSpan(t, s[0].Time, 0, 1, 3, 2)
	assertOccurrences(t, []track.Occurrence{{Place: track.Place{Voice: v, Bar: 1, Index: 0}, NoteIndex: 0}, {Place: track.Place{Voice: v, Bar: 2, Index: 0}, NoteIndex: 0}}, s[0].Parts)
	at, ok := p.MustAt(v, position(1, 1))
	require.True(t, ok)
	assert.Equal(t, 2, at.Place.Bar)
	bad, err = p.Splice(v, 2, 0, 1)
	assert.Nil(t, bad)
	assert.Error(t, err)
	assert.Len(t, p.MustSounds()[0].Parts, 2)
}

func TestGraceGroupsPreserveOrderButDoNotSound(t *testing.T) {
	t.Parallel()
	p, v := fixture(2)
	g1 := tone("C4", duration.Eighth, event.Marks{Grace: true})
	g2 := tone("D4", duration.Sixteenth, event.Marks{Grace: true, Ties: []int{0}, Fermata: true})
	main := tone("D4", duration.Quarter, event.Marks{})
	p = p.MustSplice(v, 1, 0, 0, g1, g2, main)
	e := p.MustEntries(v, 1)
	require.Len(t, e, 3)
	assert.Same(t, g1, e[0].Event)
	assert.Same(t, g2, e[1].Event)
	assertSpan(t, e[0].Time, 0, 1, 0, 1)
	assertSpan(t, e[1].Time, 0, 1, 0, 1)
	at, ok := p.MustAt(v, duration.Origin())
	require.True(t, ok)
	assert.Same(t, main, at.Event)
	s := p.MustSounds()
	require.Len(t, s, 1)
	assertNote(t, parsed("D4"), s[0].Note)
	assertOccurrences(t, []track.Occurrence{{Place: track.Place{Voice: v, Bar: 1, Index: 2}, NoteIndex: 0}}, s[0].Parts)
	assert.True(t, e[1].Event.MustDuration().Equal(written(duration.Sixteenth, 0)))
	for _, events := range [][]*event.Event{{g1}, {g2, tone("E4", duration.Quarter, event.Marks{})}, {g1, rest(written(duration.Quarter, 0), false)}} {
		base, bv := fixture(1)
		next, err := base.Splice(bv, 1, 0, 0, events...)
		assert.Nil(t, next)
		assert.Error(t, err)
	}
	base, bv := fixture(2)
	whole := tone("D4", duration.Whole, event.Marks{Ties: []int{0}})
	base = base.MustEdit(track.Edit{Bars: []track.BarEdit{{Voice: bv, Bar: 1, Events: []*event.Event{whole, g1, g2}}, {Voice: bv, Bar: 2, Events: []*event.Event{main}}}})
	be := base.MustEntries(bv, 1)
	require.Len(t, be, 3)
	assertSpan(t, be[1].Time, 1, 1, 1, 1)
	bs := base.MustSounds()
	require.Len(t, bs, 1)
	assertSpan(t, bs[0].Time, 0, 1, 5, 4)
	// Isolate the orphaned grace group from a separate main-note tie failure.
	bad, err := base.Edit(track.Edit{Bars: []track.BarEdit{
		{Voice: bv, Bar: 1, Events: []*event.Event{tone("D4", duration.Whole, event.Marks{}), g1, g2}},
		{Voice: bv, Bar: 2},
	}})
	assert.Nil(t, bad)
	issues(t, err, track.ErrGraceTarget)
}
