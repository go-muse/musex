package track

import (
	"errors"
	"slices"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/voice"
)

// ExtractLine measures context-dependent rests and materializes only leading
// and internal gaps. It never includes an unrecorded trailing grid tail.
func (t *Track) ExtractLine(v VoiceID, mode LineMode) (Extraction, error) {
	ri, err := t.voice(v)
	if err != nil {
		return Extraction{}, err
	}
	if mode != Exact && mode != PrincipalOnly {
		return Extraction{}, ErrInvalidArgument
	}
	out := Extraction{}
	var es []*event.Event
	var origins []Entry
	pos := duration.Origin()
	for _, en := range t.records[ri].entries {
		if en.Event.MustMarks().Grace && mode == PrincipalOnly {
			out.Skipped = append(out.Skipped, en)
			continue
		}
		if en.Time.Start.Compare(pos) > 0 {
			d, err := pos.DurationTo(en.Time.Start)
			if err != nil {
				return Extraction{}, err
			}
			e, err := rest(d, event.Marks{})
			if err != nil {
				return Extraction{}, err
			}
			out.Gaps = append(out.Gaps, Span{pos, en.Time.Start})
			es = append(es, e)
			origins = append(origins, en)
		}
		e := en.Event
		if e.MustDuration().Kind() == event.DurationMeasures {
			out.Measured = append(out.Measured, en)
			d, err := en.Time.Start.DurationTo(en.Time.End)
			if err != nil {
				return Extraction{}, err
			}
			e, err = rest(d, e.MustMarks())
			if err != nil {
				return Extraction{}, err
			}
		}
		es = append(es, e)
		origins = append(origins, en)
		pos = en.Time.End
	}
	line, err := voice.FromEvents(es...)
	if err != nil {
		if ie, ok := errors.AsType[*voice.ItemError](err); ok && ie.Index >= 0 && ie.Index < len(origins) {
			return Extraction{}, problem(err, origins[ie.Index].Place)
		}
		return Extraction{}, err
	}
	out.Line = line
	end, err := t.VoiceEnd(v)
	if err != nil {
		return Extraction{}, err
	}
	if end.Equal(duration.Origin()) {
		return out, nil
	}
	for _, b := range t.MustFermataBars() {
		p, err := t.grid.BarEnd(b)
		if err != nil {
			return Extraction{}, err
		}
		if p.Compare(end) <= 0 {
			out.FermataBars = append(out.FermataBars, b)
		}
	}
	if t.clock.Map.IsSet() {
		m, err := t.PerformanceMap()
		if err != nil {
			return Extraction{}, err
		}
		for _, ex := range m.MustExtensions() {
			p, _, _ := ex.MustAddress().Position()
			if p.Compare(end) <= 0 {
				out.Extensions = append(out.Extensions, p)
			}
		}
		slices.SortFunc(out.Extensions, func(a, b duration.Position) int { return a.Compare(b) })
	}
	return out, nil
}

// WriteLine expands each element's notation before using the common layout.
func (t *Track) WriteLine(v VoiceID, at measure.Address, line *voice.Line, o WriteOptions) (Proposal, error) {
	if _, err := t.voice(v); err != nil {
		return Proposal{}, err
	}
	if !validOptions(o) {
		return Proposal{}, ErrInvalidArgument
	}
	if !line.IsSet() {
		return Proposal{}, errors.Join(ErrInvalidArgument, voice.ErrUnset)
	}
	elems, err := line.Elements()
	if err != nil {
		return Proposal{}, err
	}
	var es []*event.Event
	var inputs []int
	for i, el := range elems {
		one, err := voice.New(el)
		if err != nil {
			return Proposal{}, err
		}
		xs, err := one.Events()
		if err != nil {
			return Proposal{}, err
		}
		es = append(es, xs...)
		for range xs {
			inputs = append(inputs, i)
		}
	}
	p, err := t.WriteEvents(v, at, es, o)
	if err != nil {
		if ps, ok := errors.AsType[*Problems](err); ok {
			for i := range ps.Issues {
				k := ps.Issues[i].InputIndex
				if k >= 0 && k < len(inputs) {
					ps.Issues[i].InputIndex = inputs[k]
				}
			}
		}
		return Proposal{}, err
	}
	// One element can expand to several original notation pieces. Group those
	// changes so its provenance remains visible as one independent operation.
	merged := map[int]int{}
	var cs []Change
	for _, ch := range p.Changes {
		if len(ch.Input) == 0 {
			cs = append(cs, ch)
			continue
		}
		idx := inputs[ch.Input[0]]
		if j, ok := merged[idx]; ok {
			cs[j].From = append(cs[j].From, ch.From...)
			cs[j].To = append(cs[j].To, ch.To...)
		} else {
			merged[idx] = len(cs)
			ch.Input = []int{idx}
			cs = append(cs, ch)
		}
	}
	p.Changes = cs
	sortChanges(p.Changes)
	return p, nil
}
