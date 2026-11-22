package score

import (
	"errors"
	"slices"

	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/tempomap"
	"github.com/go-muse/musex/track"
)

// Selection chooses voices within one participation; empty voices means all.
type Selection struct {
	Part   PartID
	Voices []track.VoiceID
}

// VoiceSlice distinguishes an empty record from an explicit event.
type VoiceSlice struct {
	Voice track.VoiceID
	Entry track.Entry
	Found bool
}

// PartSlice identifies a participation whose grid may have already ended.
type PartSlice struct {
	Part   PartID
	Ended  bool
	Voices []VoiceSlice
}

// Sound identifies an independently sustained note in one participation.
type Sound struct {
	Part  PartID
	Sound track.Sound
}

// TimedSound preserves the map's rounded values and precision guarantees.
type TimedSound struct {
	Sound      Sound
	Start, End tempomap.SecondsResult
}

func (s *Score) selectParts(sel []Selection) ([]Selection, error) {
	if !s.IsSet() {
		return nil, ErrUnset
	}
	chosen := map[PartID][]track.VoiceID{}
	for _, x := range sel {
		i, err := s.index(x.Part)
		if err != nil {
			return nil, err
		}
		if _, ok := chosen[x.Part]; ok {
			return nil, ErrInvalidArgument
		}
		vs := s.parts[i].record.MustVoices()
		seen := map[track.VoiceID]bool{}
		for _, v := range x.Voices {
			if !slices.Contains(vs, v) {
				return nil, errors.Join(ErrVoice, track.ErrVoice)
			}
			if seen[v] {
				return nil, ErrInvalidArgument
			}
			seen[v] = true
		}
		chosen[x.Part] = x.Voices
	}
	var out []Selection
	for _, p := range s.parts {
		vs, ok := chosen[p.id]
		if len(sel) > 0 && !ok {
			continue
		}
		all := p.record.MustVoices()
		var selected []track.VoiceID
		for _, v := range all {
			if len(vs) == 0 || slices.Contains(vs, v) {
				selected = append(selected, v)
			}
		}
		out = append(out, Selection{p.id, selected})
	}
	return out, nil
}
func (s *Score) checkPosition(p duration.Position) error {
	if !s.IsSet() {
		return ErrUnset
	}
	if !p.IsSet() {
		return errors.Join(ErrInvalidArgument, duration.ErrUnset)
	}
	e, err := s.End()
	if err != nil {
		return err
	}
	if p.Compare(e) > 0 {
		return ErrPosition
	}
	return nil
}

// At returns the record at a position in score and voice order.
func (s *Score) At(p duration.Position, selection ...Selection) ([]PartSlice, error) {
	if !s.IsSet() {
		return nil, ErrUnset
	}
	if !p.IsSet() {
		return nil, errors.Join(ErrInvalidArgument, duration.ErrUnset)
	}
	sel, err := s.selectParts(selection)
	if err != nil {
		return nil, err
	}
	if err := s.checkPosition(p); err != nil {
		return nil, err
	}
	end, err := s.End()
	if err != nil {
		return nil, err
	}
	if p.Equal(end) {
		return nil, nil
	}
	var out []PartSlice
	for _, x := range sel {
		i, _ := s.index(x.Part)
		t := s.parts[i].record
		e, _, err := t.MustGrid().End()
		if err != nil {
			return nil, classify(err)
		}
		row := PartSlice{Part: x.Part, Ended: p.Compare(e) >= 0}
		if !row.Ended {
			for _, v := range x.Voices {
				entry, ok, err := t.At(v, p)
				if err != nil {
					return nil, classify(err)
				}
				row.Voices = append(row.Voices, VoiceSlice{v, cloneEntry(entry), ok})
			}
		}
		out = append(out, row)
	}
	return out, nil
}

// Sounds returns sustained notes ordered by attack, participation, and voice.
func (s *Score) Sounds(selection ...Selection) ([]Sound, error) {
	sel, err := s.selectParts(selection)
	if err != nil {
		return nil, err
	}
	var out []Sound
	for _, x := range sel {
		i, _ := s.index(x.Part)
		ss, err := s.parts[i].record.Sounds(x.Voices...)
		if err != nil {
			return nil, classify(err)
		}
		for _, v := range ss {
			v.Parts = slices.Clone(v.Parts)
			out = append(out, Sound{x.Part, v})
		}
	}
	slices.SortStableFunc(out, func(a, b Sound) int { return a.Sound.Time.Start.Compare(b.Sound.Time.Start) })
	return out, nil
}

// SoundingAt returns all notes sounding on the half-open musical interval.
func (s *Score) SoundingAt(p duration.Position, selection ...Selection) ([]Sound, error) {
	if !s.IsSet() {
		return nil, ErrUnset
	}
	if !p.IsSet() {
		return nil, errors.Join(ErrInvalidArgument, duration.ErrUnset)
	}
	ss, err := s.Sounds(selection...)
	if err != nil {
		return nil, err
	}
	if err := s.checkPosition(p); err != nil {
		return nil, err
	}
	var out []Sound
	for _, v := range ss {
		if p.Compare(v.Sound.Time.Start) >= 0 && p.Compare(v.Sound.Time.End) < 0 {
			out = append(out, v)
		}
	}
	return out, nil
}

// SoundSeconds uses the common map's End side at both attack and release.
func (s *Score) SoundSeconds(places int, mode rounding.Mode, selection ...Selection) ([]TimedSound, error) {
	if !s.IsSet() {
		return nil, ErrUnset
	}
	if places < 0 || places > tempomap.MaxPlaces {
		return nil, errors.Join(ErrInvalidArgument, tempomap.ErrInvalidArgument)
	}
	if _, err := mode.Name(); err != nil {
		return nil, errors.Join(ErrInvalidArgument, err)
	}
	ss, err := s.Sounds(selection...)
	if err != nil {
		return nil, err
	}
	var out []TimedSound
	for _, v := range ss {
		a, err := s.clock.End(tempomap.MustAt(v.Sound.Time.Start), places, mode)
		if err != nil {
			return nil, classify(err)
		}
		b, err := s.clock.End(tempomap.MustAt(v.Sound.Time.End), places, mode)
		if err != nil {
			return nil, classify(err)
		}
		out = append(out, TimedSound{v, a, b})
	}
	return out, nil
}

// ViewEntry relates a displayed bar event to its complete original record.
type ViewEntry struct {
	Source   track.Entry
	Bar      int
	Event    *event.Event
	Time     track.Span
	Expanded bool
}

// BarView reads one bar, expanding a multibar rest only for this display.
func (s *Score) BarView(id PartID, v track.VoiceID, bar int) ([]ViewEntry, error) {
	i, err := s.index(id)
	if err != nil {
		return nil, err
	}
	p := s.parts[i]
	entries, err := p.record.Entries(v, bar)
	if err != nil {
		if errors.Is(err, track.ErrVoice) {
			return nil, errors.Join(ErrVoice, err)
		}
		return nil, &Problems{Issues: trackIssues(id, err)}
	}
	g := p.record.MustGrid()
	var out []ViewEntry
	for _, cell := range p.cells {
		if cell.Voice != v || cell.Bar > bar {
			continue
		}
		es := entries
		if cell.Bar != bar {
			es, err = p.record.Entries(v, cell.Bar)
			if err != nil {
				return nil, classify(err)
			}
		}
		for _, e := range es {
			d := e.Event.MustDuration()
			if d.Kind() == event.DurationMeasures && d.MustMeasureCount() > 1 {
				count := d.MustMeasureCount()
				if bar-cell.Bar >= count {
					continue
				}
				marks := e.Event.MustMarks()
				marks.Fermata = marks.Fermata && bar-cell.Bar == count-1
				rest, err := event.New(event.RestContent(), event.MustMeasureDuration(1), marks)
				if err != nil {
					return nil, err
				}
				b, err := g.Bar(bar)
				if err != nil {
					return nil, classify(err)
				}
				out = append(out, ViewEntry{cloneEntry(e), bar, rest, track.Span{Start: b.Start, End: b.End}, true})
			} else if cell.Bar == bar {
				out = append(out, ViewEntry{cloneEntry(e), bar, cloneEvent(e.Event), e.Time, false})
			}
		}
	}
	return out, nil
}
