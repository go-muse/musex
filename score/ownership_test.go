package score_test

import (
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/meter"
	"github.com/go-muse/musex/score"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempomap"
	"github.com/go-muse/musex/track"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sparseBar = 100_000_000

// Includes kinds absent from Sounds: explicit rests, a boundary grace, and a
// multibar sign. A removed far cell must not survive in an internal index.
func sparseRecord() *track.Track {
	p := emptyPart(4, 4, sparseBar+2)
	v := p.MustVoices()[0]
	grace := event.MustNew(event.MustNoteContent(tone(p, "B4", 1, 8, event.Marks{}).MustContent().MustNote()), event.MustWrittenDuration(duration.MustNewWritten(duration.Eighth, 0)), event.Marks{Grace: true})
	p = p.MustEdit(track.Edit{Bars: []track.BarEdit{
		{Voice: v, Bar: 1, Events: []*event.Event{tone(p, "C4", 1, 4, event.Marks{Ties: []int{0}}), tone(p, "C4", 1, 4, event.Marks{}), event.MustNew(event.RestContent(), event.MustValueDuration(fraction.MustNew(1, 2)), event.Marks{}), grace}},
		{Voice: v, Bar: 2, Events: []*event.Event{tone(p, "D4", 1, 4, event.Marks{})}},
		{Voice: v, Bar: 3, Events: []*event.Event{rest(3, true)}},
		{Voice: v, Bar: sparseBar, Events: []*event.Event{tone(p, "E4", 1, 4, event.Marks{})}},
		{Voice: v, Bar: sparseBar - 1, Events: []*event.Event{tone(p, "F4", 1, 4, event.Marks{})}},
	}, Barlines: []track.BarlineEdit{{AfterBar: 2, Fermata: true}}})
	return p.MustSplice(v, sparseBar-1, 0, 1)
}

func captureRecord(p *track.Track) map[int][]track.Entry {
	out := make(map[int][]track.Entry)
	for _, bar := range []int{1, 2, 3, 4, 5, sparseBar - 1, sparseBar} {
		entries := p.MustEntries(p.MustVoices()[0], bar)
		for i := range entries {
			clone := *entries[i].Event
			entries[i].Event = &clone
		}
		out[bar] = entries
	}
	return out
}

func assertRecord(t *testing.T, s *score.Score, id score.PartID, want map[int][]track.Entry) {
	t.Helper()
	p := s.MustPart(id)
	for bar, expected := range want {
		got := p.MustEntries(p.MustVoices()[0], bar)
		require.Len(t, got, len(expected), "bar %d", bar)
		for i, e := range expected {
			assertKey(t, e.Place, got[i].Place)
			assert.Equal(t, e.Time, got[i].Time)
			assertEvent(t, e.Event, got[i].Event)
		}
	}
	assert.Equal(t, []int{2}, p.MustFermataBars())
	assert.Len(t, s.MustSounds(), 3)
}

func overwriteRecord(p *track.Track) {
	for _, bar := range []int{1, 2, 3, sparseBar} {
		for _, e := range p.MustEntries(p.MustVoices()[0], bar) {
			*e.Event = event.Event{}
		}
	}
	*p = track.Track{}
}

func TestInputEventAndTrackShellIsolation(t *testing.T) {
	t.Parallel()
	for _, operation := range []string{"Add", "Replace", "AddWithoutClock", "ReplaceWithoutClock"} {
		t.Run(operation, func(t *testing.T) {
			p := sparseRecord()
			if operation == "AddWithoutClock" || operation == "ReplaceWithoutClock" {
				local := track.Clock{Map: tempomap.MustNewMeasured(p.MustGrid(), bpm(90)), BarHolds: []track.BarHold{{AfterBar: 2, Seconds: tempo.MustNewSeconds(1, 1)}}}
				p = p.MustEdit(track.Edit{Clock: &local})
			}
			want := captureRecord(p)
			inst := p.MustInstrument()
			s := score.MustNew(clock60())
			var id score.PartID
			switch operation {
			case "Add":
				s, id = s.MustAdd(p)
			case "Replace":
				s, id = s.MustAdd(emptyPart(4, 4, sparseBar+2))
				s = s.MustReplace(id, p)
			case "AddWithoutClock":
				inc := s.MustAddWithoutClock(p)
				s, id = inc.Score, inc.Part
			case "ReplaceWithoutClock":
				s, id = s.MustAdd(emptyPart(4, 4, sparseBar+2))
				inc := s.MustReplaceWithoutClock(id, p)
				s = inc.Score
			}
			derived := s.MustWithMap(clock60().MustWithExtension(tempomap.MustNewExtension(at(1, 2), tempo.MustNewSeconds(1, 1))))
			overwriteRecord(p)
			assertRecord(t, s, id, want)
			assertRecord(t, derived, id, want)
			assert.Same(t, inst, s.MustPart(id).MustInstrument())
			assert.Same(t, inst, s.MustSounds()[0].Sound.Note.Source())
			require.NoError(t, inst.SetName("Live piano"))
			assert.Equal(t, "Live piano", derived.MustPart(id).MustInstrument().Name())
		})
	}
}

func TestOutputEventAndTrackShellIsolation(t *testing.T) {
	t.Parallel()
	for _, source := range []string{"Part", "ExtractWithMap", "Before", "After", "At", "BarView.Event", "BarView.Source"} {
		t.Run(source, func(t *testing.T) {
			p := sparseRecord()
			want := captureRecord(p)
			s, ids := assemble(clock60(), p)
			id := ids[0]
			proposal := s.MustChangeMeter(2, 2, meter.MustNew(3, 4))
			derived := proposal.Score
			derivedWant := captureRecord(derived.MustPart(id))
			switch source {
			case "Part":
				overwriteRecord(s.MustPart(id))
			case "ExtractWithMap":
				overwriteRecord(s.MustExtractWithMap(id))
			case "Before":
				overwriteRecord(proposal.Parts[0].Before)
			case "After":
				overwriteRecord(proposal.Parts[0].After)
			case "At":
				rows := s.MustAt(pos(0, 1))
				*rows[0].Voices[0].Entry.Event = event.Event{}
			case "BarView.Event":
				rows := s.MustBarView(id, p.MustVoices()[0], 1)
				for _, e := range rows {
					*e.Event = event.Event{}
				}
			case "BarView.Source":
				rows := s.MustBarView(id, p.MustVoices()[0], 4)
				require.Len(t, rows, 1)
				*rows[0].Source.Event = event.Event{}
			}
			assertRecord(t, s, id, want)
			assertRecord(t, derived, id, derivedWant)
			// Re-read the particular surface, not only the Track adapter.
			assert.True(t, s.MustAt(pos(0, 1))[0].Voices[0].Entry.Event.IsSet())
			assert.Equal(t, 3, s.MustBarView(id, p.MustVoices()[0], 4)[0].Source.Event.MustDuration().MustMeasureCount())
		})
	}
}

func TestSecondVoiceEventShellIsolation(t *testing.T) {
	t.Parallel()
	for _, source := range []string{"Add", "Replace", "Part", "ExtractWithMap", "Before", "After", "At", "BarView.Event", "BarView.Source"} {
		t.Run(source, func(t *testing.T) {
			p := emptyPart(4, 4, 2)
			p, secondVoice := p.MustAddVoice()
			p = write(p, 1, tone(p, "C4", 1, 4, event.Marks{}))
			second := tone(p, "E4", 1, 2, event.Marks{})
			p = p.MustSplice(secondVoice, 1, 0, 0, second)
			want := *second
			s, ids := assemble(clock60(), p)
			id := ids[0]
			if source == "Replace" {
				s, ids = assemble(clock60(), emptyPart(4, 4, 2))
				id = ids[0]
				s = s.MustReplace(id, p)
			}
			proposal := s.MustChangeMeter(1, 1, meter.MustNew(3, 4))
			switch source {
			case "Add", "Replace":
				*p.MustEntries(secondVoice, 1)[0].Event = event.Event{}
			case "Part":
				*s.MustPart(id).MustEntries(secondVoice, 1)[0].Event = event.Event{}
			case "ExtractWithMap":
				*s.MustExtractWithMap(id).MustEntries(secondVoice, 1)[0].Event = event.Event{}
			case "Before":
				*proposal.Parts[0].Before.MustEntries(secondVoice, 1)[0].Event = event.Event{}
			case "After":
				*proposal.Parts[0].After.MustEntries(secondVoice, 1)[0].Event = event.Event{}
			case "At":
				rows := s.MustAt(pos(0, 1))
				*rows[0].Voices[1].Entry.Event = event.Event{}
			case "BarView.Event":
				*s.MustBarView(id, secondVoice, 1)[0].Event = event.Event{}
			case "BarView.Source":
				*s.MustBarView(id, secondVoice, 1)[0].Source.Event = event.Event{}
			}
			for _, snapshot := range []*score.Score{s, proposal.Score} {
				assertKeys(t, []track.VoiceID{p.MustVoices()[0], secondVoice}, snapshot.MustPart(id).MustVoices())
				assertEvent(t, &want, snapshot.MustPart(id).MustEntries(secondVoice, 1)[0].Event)
				rows := snapshot.MustAt(pos(1, 4))
				require.Len(t, rows[0].Voices, 2)
				assert.True(t, rows[0].Voices[1].Found)
				assertEvent(t, &want, rows[0].Voices[1].Entry.Event)
				assertEvent(t, &want, snapshot.MustBarView(id, secondVoice, 1)[0].Source.Event)
				assert.Len(t, snapshot.MustSounds(), 2)
			}
		})
	}
}

func TestNestedReturnedContainersAreIndependent(t *testing.T) {
	t.Parallel()
	p := emptyPart(4, 4, 2)
	p = write(p, 1, tone(p, "C4", 1, 2, event.Marks{Ties: []int{0}}), tone(p, "C4", 1, 2, event.Marks{}))
	clock := clock60().MustWithExtension(tempomap.MustNewExtension(at(2, 1), tempo.MustNewSeconds(1, 1)))
	s, ids := assemble(clock, p, p)
	parts := s.MustParts()
	parts[0] = score.PartID{}
	assertKeys(t, ids, s.MustParts())
	slice := s.MustAt(pos(0, 1))
	slice[0].Voices[0].Voice = track.VoiceID{}
	slice[0].Voices = nil
	assertKey(t, p.MustVoices()[0], s.MustAt(pos(0, 1))[0].Voices[0].Voice)
	sounds := s.MustSounds()
	sounds[0].Sound.Parts[0].Place.Bar = 999
	assert.Equal(t, 1, s.MustSounds()[0].Sound.Parts[0].Place.Bar)
	timed := s.MustSoundSeconds(3, rounding.HalfUp)
	timed[0].Sound.Sound.Parts[0].Place.Bar = 999
	assert.Equal(t, 1, s.MustSoundSeconds(3, rounding.HalfUp)[0].Sound.Sound.Parts[0].Place.Bar)
	proposal := s.MustChangeMeter(1, 1, meter.MustNew(4, 4))
	userCopy := proposal
	userCopy.Parts[0].Part = score.PartID{} // User struct copies share their own slice.
	assert.Zero(t, proposal.Parts[0].Part)
	for i := range proposal.Marks {
		for j := range proposal.Marks[i].Bindings {
			b := &proposal.Marks[i].Bindings[j]
			b.Part = score.PartID{}
			if b.AfterBar != nil {
				*b.AfterBar = 999
			}
		}
	}
	repeated := s.MustChangeMeter(1, 1, meter.MustNew(4, 4))
	assertKey(t, ids[0], repeated.Parts[0].Part)
	assert.Equal(t, 2, *repeated.Marks[1].Bindings[0].AfterBar)
	assert.True(t, repeated.Score.MustMap().Equal(clock))
	lines := s.MustCommonBarlines(pos(0, 1), pos(2, 1), ids...)
	*lines[0].Boundaries[0].NextBar = 999
	*lines[1].Boundaries[0].PreviousBar = 999
	*lines[1].Boundaries[0].NextBar = 999
	repeatedLines := s.MustCommonBarlines(pos(0, 1), pos(2, 1), ids...)
	assert.Equal(t, 1, *repeatedLines[0].Boundaries[0].NextBar)
	assert.Equal(t, 1, *repeatedLines[1].Boundaries[0].PreviousBar)
	copyValue := *s
	derived := s.MustWithMap(clock60())
	*s = score.Score{}
	assert.True(t, copyValue.IsSet())
	assertKeys(t, ids, copyValue.MustParts())
	assertKeys(t, ids, derived.MustParts())
}

func TestGraceAtBarlineIsOnlyInNotation(t *testing.T) {
	t.Parallel()
	p := sparseRecord()
	s, ids := assemble(clock60(), p)
	row := s.MustAt(pos(1, 1))
	require.Len(t, row, 1)
	assert.Equal(t, 2, row[0].Voices[0].Entry.Place.Bar)
	view := s.MustBarView(ids[0], p.MustVoices()[0], 1)
	require.Len(t, view, 4)
	assert.True(t, view[3].Event.MustMarks().Grace)
	assertPos(t, pos(1, 1), view[3].Time.Start)
	assertPos(t, view[3].Time.Start, view[3].Time.End)
	active := s.MustSoundingAt(pos(1, 1))
	require.Len(t, active, 1)
	assert.Equal(t, 2, active[0].Sound.Parts[0].Place.Bar)
}
