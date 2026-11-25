package voiceleading

import (
	"slices"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pitch"
)

// PartSonority is the state of every part of an analysis during [Start, End):
// Elements[p] is the element of part p that sounds or rests at Start, with
// its original index and complete boundaries. Voices is a new standalone
// snapshot of the sounding parts in input-slot order with the order and
// system of the analysis. Elements belongs to the caller; Voices is an
// immutable borrowed snapshot.
type PartSonority struct {
	Start, End duration.Position
	Elements   []ElementAt
	Voices     *Sonority
}

// jointEnd returns the smallest end of all parts.
func (a *Analysis) jointEnd() duration.Position {
	end := duration.Origin()
	for p, elems := range a.parts {
		e := duration.Origin()
		if len(elems) > 0 {
			e = elems[len(elems)-1].End
		}
		if p == 0 || e.Compare(end) < 0 {
			end = e
		}
	}
	return end
}

// moments returns the distinct element starts of all parts before the joint
// end, in increasing order, and the joint end.
func (a *Analysis) moments() ([]duration.Position, duration.Position) {
	end := a.jointEnd()
	var starts []duration.Position
	for _, elems := range a.parts {
		for _, e := range elems {
			if e.Start.Compare(end) < 0 {
				starts = append(starts, e.Start)
			}
		}
	}
	slices.SortFunc(starts, duration.Position.Compare)
	starts = slices.CompactFunc(starts, duration.Position.Equal)
	return starts, end
}

// elementAt returns the element of part p containing at, which must lie
// before the end of the part.
func (a *Analysis) elementAt(p int, at duration.Position) ElementAt {
	elems := a.parts[p]
	i, found := slices.BinarySearchFunc(elems, at, func(e ElementAt, x duration.Position) int { return e.Start.Compare(x) })
	if !found {
		i--
	}
	return elems[i]
}

// partSonority builds the sonority of the moment [start, end). Its voices
// repeat the pairwise measurements and never fail or read sources.
func (a *Analysis) partSonority(start, end duration.Position) PartSonority {
	ps := PartSonority{Start: start, End: end, Elements: make([]ElementAt, len(a.parts))}
	s := &Sonority{set: true, sys: a.sys, order: a.order, hasParts: true, parts: []int{}, silent: []int{}, notes: []note.Note{}, pitches: []pitch.Pitch{}}
	for p := range a.parts {
		e := a.elementAt(p, start)
		ps.Elements[p] = e
		if !e.Element.IsSound() {
			s.silent = append(s.silent, p)
			continue
		}
		n, _ := e.Element.MustNote()
		s.parts = append(s.parts, p)
		s.notes = append(s.notes, n)
		s.pitches = append(s.pitches, n.MustPitch())
	}
	s.adjacent = []VoiceInterval{}
	switch n := len(s.parts); {
	case n >= 2:
		for k := 1; k < n; k++ {
			s.adjacent = append(s.adjacent, a.voiceInterval(s.parts, k-1, k, start))
		}
		s.span = a.voiceInterval(s.parts, 0, n-1, start)
	case n == 1:
		e := ps.Elements[s.parts[0]]
		if a.sys == nil {
			s.soundErr = &PartError{Part: s.parts[0], Element: e.Index, Err: ErrNoSystem}
		} else if err := admit(a.sys, s.notes[0].Spelling()); err != nil {
			s.soundErr = &PartError{Part: s.parts[0], Element: e.Index, Err: err}
		}
	}
	ps.Voices = s
	return ps
}

// voiceInterval reads the measured pairwise vertical of two sounding voices.
func (a *Analysis) voiceInterval(parts []int, lower, upper int, at duration.Position) VoiceInterval {
	vs := a.data[a.index[Pair{First: parts[lower], Second: parts[upper]}]].verticals
	i, found := verticalIndex(vs, at)
	if !found {
		i--
	}
	return VoiceInterval{Lower: lower, Upper: upper, Interval: vs[i].Specific, Chromatic: vs[i].Chromatic}
}

// Sonorities returns the sonorities of all parts at every moment in time
// order: a moment starts at an element start of any part before the joint end
// of all parts. A part sonority never fails and reads no sources; it does not
// require a nominal order, but spacing and connections of its voices do.
// Without moments the result is empty. Errors: ErrUnset.
func (a *Analysis) Sonorities() ([]PartSonority, error) {
	if err := a.check(); err != nil {
		return nil, err
	}
	starts, end := a.moments()
	out := make([]PartSonority, len(starts))
	for i, start := range starts {
		out[i] = a.partSonority(start, momentEnd(starts, i, end))
	}
	return out, nil
}

// MustSonorities is like Sonorities but panics with its error.
func (a *Analysis) MustSonorities() []PartSonority { return must(a.Sonorities()) }

func momentEnd(starts []duration.Position, i int, end duration.Position) duration.Position {
	if i+1 < len(starts) {
		return starts[i+1]
	}
	return end
}

// SonorityAt returns the sonority containing at with its complete boundaries.
// At or after the joint end, or without moments, it returns the zero value,
// false, nil. Errors, in this order: ErrUnset; ErrUnset with duration.ErrUnset
// for an unset position.
func (a *Analysis) SonorityAt(at duration.Position) (PartSonority, bool, error) {
	if err := a.check(); err != nil {
		return PartSonority{}, false, err
	}
	if !at.IsSet() {
		return PartSonority{}, false, unset("position", duration.ErrUnset)
	}
	starts, end := a.moments()
	i, found := slices.BinarySearchFunc(starts, at, duration.Position.Compare)
	if !found {
		i--
	}
	if i < 0 || at.Compare(end) >= 0 {
		return PartSonority{}, false, nil
	}
	return a.partSonority(starts[i], momentEnd(starts, i, end)), true, nil
}

// MustSonorityAt is like SonorityAt but panics with its error.
func (a *Analysis) MustSonorityAt(at duration.Position) (PartSonority, bool) {
	return must2(a.SonorityAt(at))
}
