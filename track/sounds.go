package track

import (
	"slices"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/event"
)

func (t *Track) filter(vs []VoiceID) (map[VoiceID]bool, error) {
	if !t.IsSet() {
		return nil, ErrUnset
	}
	out := map[VoiceID]bool{}
	if len(vs) == 0 {
		for _, r := range t.records {
			out[r.id] = true
		}
		return out, nil
	}
	for _, v := range vs {
		if _, err := t.voice(v); err != nil {
			return nil, err
		}
		if out[v] {
			return nil, ErrInvalidArgument
		}
		out[v] = true
	}
	return out, nil
}

// Sounds joins tied occurrences without calculating frequencies.
func (t *Track) Sounds(voices ...VoiceID) ([]Sound, error) {
	selected, err := t.filter(voices)
	if err != nil {
		return nil, err
	}
	var out []Sound
	order := map[VoiceID]int{}
	for ri, r := range t.records {
		order[r.id] = ri
		if !selected[r.id] {
			continue
		}
		incoming := map[Occurrence]int{}
		for i, e := range r.entries {
			if e.Event.MustMarks().Grace || e.Event.Kind() == event.KindRest {
				continue
			}
			ns := e.Event.MustContent().MustNotes()
			active := make(map[int]int, len(ns))
			for ni, n := range ns {
				occ := Occurrence{e.Place, ni}
				si, ok := incoming[occ]
				if !ok {
					si = len(out)
					out = append(out, Sound{Note: n, Time: e.Time})
				} else {
					out[si].Time.End = e.Time.End
				}
				out[si].Parts = append(out[si].Parts, occ)
				active[ni] = si
			}
			ties := e.Event.MustMarks().Ties
			if len(ties) == 0 {
				continue
			}
			j := nextMain(r.entries, i)
			if j < 0 {
				continue
			}
			next := r.entries[j]
			pairs, err := matching(ns, next.Event.MustContent().MustNotes(), ties)
			if err != nil {
				return nil, err
			}
			for from, to := range pairs {
				incoming[Occurrence{next.Place, to}] = active[from]
			}
		}
	}
	slices.SortStableFunc(out, func(a, b Sound) int {
		if c := a.Time.Start.Compare(b.Time.Start); c != 0 {
			return c
		}
		x, y := a.Parts[0], b.Parts[0]
		if order[x.Place.Voice] != order[y.Place.Voice] {
			return order[x.Place.Voice] - order[y.Place.Voice]
		}
		if x.Place.Bar != y.Place.Bar {
			return compareInt(x.Place.Bar, y.Place.Bar)
		}
		if x.Place.Index != y.Place.Index {
			return compareInt(x.Place.Index, y.Place.Index)
		}
		return compareInt(x.NoteIndex, y.NoteIndex)
	})
	return out, nil
}

// SoundingAt selects ordinary sounds in half-open intervals.
func (t *Track) SoundingAt(p duration.Position, voices ...VoiceID) ([]Sound, error) {
	ss, err := t.Sounds(voices...)
	if err != nil {
		return nil, err
	}
	if err = t.position(p, VoiceID{}); err != nil {
		return nil, err
	}
	var out []Sound
	for _, s := range ss {
		if p.Compare(s.Time.Start) >= 0 && p.Compare(s.Time.End) < 0 {
			out = append(out, s)
		}
	}
	return out, nil
}
