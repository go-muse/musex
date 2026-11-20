package voice

import (
	"fmt"

	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tempomap"
)

// Timing is the start and the end in seconds of an element: the index of the
// element and two results of package tempomap with their approximation flags,
// precision and error bounds.
type Timing struct {
	Index      int
	Start, End tempomap.SecondsResult
}

// checkMap reports ErrUnset for an unset map and ErrMeasureMap for a map
// addressed by measures, with the check of start between them.
func checkMap(m tempomap.Map, start *duration.Position) error {
	if !m.IsSet() {
		return unsetError("tempo map", tempomap.ErrUnset)
	}

	if start != nil {
		if err := checkPosition(*start, "start of the line"); err != nil {
			return err
		}
	}

	addressing, err := m.Addressing()
	if err != nil {
		return fmt.Errorf("voice: tempo map: %w", err)
	}

	if addressing != tempomap.PositionAddressing {
		return fmt.Errorf("%w: %s", ErrMeasureMap, addressing)
	}

	return nil
}

// secondsAt returns the seconds at the end of the hold at the position p of
// the map.
func secondsAt(m tempomap.Map, p duration.Position, places int, mode rounding.Mode) (tempomap.SecondsResult, error) {
	at, err := tempomap.At(p)
	if err != nil {
		return tempomap.SecondsResult{}, fmt.Errorf("voice: address %s: %w", p, err)
	}

	r, err := m.End(at, places, mode)
	if err != nil {
		return tempomap.SecondsResult{}, fmt.Errorf("voice: seconds at %s: %w", p, err)
	}

	return r, nil
}

// timing returns the seconds of the element i of a line that starts at the
// position start of the map: first the positions start + p and start + q,
// then the seconds of the start and of the end.
func (l *Line) timing(i int, m tempomap.Map, start duration.Position, places int, mode rounding.Mode) (Timing, error) {
	ps, err := shift(start, l.starts[i])
	if err != nil {
		return Timing{}, err
	}

	pe, err := shift(start, l.endOf(i))
	if err != nil {
		return Timing{}, err
	}

	s, err := secondsAt(m, ps, places, mode)
	if err != nil {
		return Timing{}, err
	}

	e, err := secondsAt(m, pe, places, mode)
	if err != nil {
		return Timing{}, err
	}

	return Timing{Index: i, Start: s, End: e}, nil
}

// SoundSeconds returns the start and the end in seconds of each sound in
// order by the tempo map m addressed by positions, for a line that starts at
// the origin of the map. Use SoundSecondsFrom for another start.
//
// The start of an element at p is m.End(p) and its end at q is m.End(q):
// the next sound starts after a hold at the end of the previous one, and the
// sound before a hold lasts until its end. A fermata of an element adds no
// hold itself: holds are extensions of the map. The number of decimal places
// and the rounding mode are required and are checked by package tempomap,
// also for a line without sounds and for exact results.
//
// Errors, in this order: ErrUnset for an unset line, then for an unset m
// (with tempomap.ErrUnset); ErrMeasureMap for a map addressed by measures;
// the errors of m.End at the start of the line, such as
// tempomap.ErrInvalidArgument for places or mode and tempomap.ErrUnfinished;
// then for each sound in an *ItemError with its index: a position out of
// range and the errors of package tempomap.
func (l *Line) SoundSeconds(m tempomap.Map, places int, mode rounding.Mode) ([]Timing, error) {
	if !l.IsSet() {
		return nil, errUnsetLine()
	}

	if err := checkMap(m, nil); err != nil {
		return nil, err
	}

	return l.soundSeconds(m, duration.Origin(), places, mode)
}

// MustSoundSeconds is like SoundSeconds but panics with its error.
func (l *Line) MustSoundSeconds(m tempomap.Map, places int, mode rounding.Mode) []Timing {
	return must(l.SoundSeconds(m, places, mode))
}

// SoundSecondsFrom is like SoundSeconds for a line that starts at the
// position start of the map. An unset start reports ErrUnset (with
// duration.ErrUnset) after the check of m.
func (l *Line) SoundSecondsFrom(m tempomap.Map, start duration.Position, places int, mode rounding.Mode) ([]Timing, error) {
	if !l.IsSet() {
		return nil, errUnsetLine()
	}

	if err := checkMap(m, &start); err != nil {
		return nil, err
	}

	return l.soundSeconds(m, start, places, mode)
}

// MustSoundSecondsFrom is like SoundSecondsFrom but panics with its error.
func (l *Line) MustSoundSecondsFrom(m tempomap.Map, start duration.Position, places int, mode rounding.Mode) []Timing {
	return must(l.SoundSecondsFrom(m, start, places, mode))
}

func (l *Line) soundSeconds(m tempomap.Map, start duration.Position, places int, mode rounding.Mode) ([]Timing, error) {
	if _, err := secondsAt(m, start, places, mode); err != nil {
		return nil, err
	}

	out := []Timing{}

	for i, e := range l.elems {
		if e.kind != KindSound {
			continue
		}

		t, err := l.timing(i, m, start, places, mode)
		if err != nil {
			return nil, itemError(i, err)
		}

		out = append(out, t)
	}

	return out, nil
}

// ElementSeconds returns the start and the end in seconds of the element i,
// a sound or a rest, by the rules of SoundSeconds. Its errors are not
// wrapped in an *ItemError.
//
// Errors, in this order: ErrUnset for an unset line; ErrIndex unless
// 0 <= i < Len; ErrUnset for an unset m; ErrMeasureMap; a position out of
// range; the errors of package tempomap for the start, then for the end.
func (l *Line) ElementSeconds(i int, m tempomap.Map, places int, mode rounding.Mode) (Timing, error) {
	return l.elementSeconds(i, m, nil, places, mode)
}

// MustElementSeconds is like ElementSeconds but panics with its error.
func (l *Line) MustElementSeconds(i int, m tempomap.Map, places int, mode rounding.Mode) Timing {
	return must(l.ElementSeconds(i, m, places, mode))
}

// ElementSecondsFrom is like ElementSeconds for a line that starts at the
// position start of the map. An unset start reports ErrUnset (with
// duration.ErrUnset) after the check of m.
func (l *Line) ElementSecondsFrom(i int, m tempomap.Map, start duration.Position, places int, mode rounding.Mode) (Timing, error) {
	return l.elementSeconds(i, m, &start, places, mode)
}

// MustElementSecondsFrom is like ElementSecondsFrom but panics with its
// error.
func (l *Line) MustElementSecondsFrom(i int, m tempomap.Map, start duration.Position, places int, mode rounding.Mode) Timing {
	return must(l.ElementSecondsFrom(i, m, start, places, mode))
}

func (l *Line) elementSeconds(i int, m tempomap.Map, start *duration.Position, places int, mode rounding.Mode) (Timing, error) {
	if !l.IsSet() {
		return Timing{}, errUnsetLine()
	}

	if err := l.checkIndex(i); err != nil {
		return Timing{}, err
	}

	if err := checkMap(m, start); err != nil {
		return Timing{}, err
	}

	s := duration.Origin()
	if start != nil {
		s = *start
	}

	return l.timing(i, m, s, places, mode)
}
