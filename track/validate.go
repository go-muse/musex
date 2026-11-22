package track

import (
	"errors"
	"slices"

	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/note"
)

func (t *Track) validate(previous ...*Track) error {
	var issues []Issue
	for ri := range t.records {
		r := &t.records[ri]
		r.entries = nil
		bars := make([]int, 0, len(r.bars))
		for b := range r.bars {
			bars = append(bars, b)
		}
		slices.Sort(bars)
		coveredThrough := -1
		for _, b := range bars {
			es := r.bars[b]
			bp := Place{r.id, b, -1}
			ok, _ := t.grid.HasBar(b)
			if !ok {
				for j := range es {
					issues = append(issues, issue(ErrLostRecord, Place{r.id, b, j}))
				}
				continue
			}
			bar, err := t.grid.Bar(b)
			if err != nil {
				return err
			}
			pos := bar.Start
			if b <= coveredThrough && len(es) > 0 {
				issues = append(issues, issue(ErrOccupied, bp))
			}
			for j, e := range es {
				p := Place{r.id, b, j}
				if !e.IsSet() {
					issues = append(issues, issue(errors.Join(ErrInvalidArgument, event.ErrUnset), p))
					continue
				}
				d := e.MustDuration()
				m := e.MustMarks()
				end := pos
				if d.Kind() == event.DurationMeasures {
					count := d.MustMeasureCount()
					n, _, _ := t.grid.BarCount()
					if j != 0 || len(es) != 1 {
						issues = append(issues, issue(ErrMeasureRest, p))
					}
					if count-1 > n-b {
						cause := ErrNeedBars
						if len(previous) > 0 {
							old := previous[0]
							oi, err := old.voice(r.id)
							if err == nil {
								prior := old.records[oi].bars[b]
								if j < len(prior) && prior[j] == e {
									cause = ErrLostRecord
								}
							}
						}
						x := issue(cause, p)
						x.BarsNeeded = count - 1 - (n - b)
						issues = append(issues, x)
						continue
					}
					last := b + count - 1
					if _, err := t.grid.RangeLength(b, last); err != nil {
						return err
					}
					runs, err := t.grid.MeterRuns(b, last)
					if err != nil {
						return err
					}
					if len(runs) > 1 {
						issues = append(issues, issue(ErrMeasureRest, p))
					}
					coveredThrough = max(coveredThrough, last)
					end, err = t.grid.BarEnd(last)
					if err != nil {
						return err
					}
				} else if !m.Grace {
					end, err = pos.Add(d.MustValue())
					if err != nil {
						return err
					}
				}
				r.entries = append(r.entries, Entry{p, e, Span{pos, end}})
				pos = end
			}
			if len(es) > 0 && es[0].IsSet() && es[0].MustDuration().Kind() != event.DurationMeasures && pos.Compare(bar.End) > 0 {
				x := issue(ErrOverflow, bp)
				x.Amount, err = bar.End.DurationTo(pos)
				if err != nil {
					return err
				}
				issues = append(issues, x)
			}
		}
	}
	for _, r := range t.records {
		issues = append(issues, t.checkLinks(r)...)
	}
	for b := range t.fermatas {
		if ok, _ := t.grid.HasBar(b); !ok {
			x := issue(ErrLostRecord, Place{Bar: b, Index: -1})
			x.AfterBar = new(b)
			x.Kind = BarlineFermata
			issues = append(issues, x)
		}
	}
	ci, err := t.validateClock()
	if err != nil {
		return err
	}
	issues = append(issues, ci...)
	if len(issues) == 0 {
		return nil
	}
	rank := map[VoiceID]int{}
	for i, r := range t.records {
		rank[r.id] = i + 1
	}
	slices.SortStableFunc(issues, func(a, b Issue) int {
		if rank[a.Place.Voice] != rank[b.Place.Voice] {
			return rank[a.Place.Voice] - rank[b.Place.Voice]
		}
		if a.Place.Bar != b.Place.Bar {
			return compareInt(a.Place.Bar, b.Place.Bar)
		}
		if a.Place.Index != b.Place.Index {
			return compareInt(a.Place.Index, b.Place.Index)
		}
		return compareInt(a.NoteIndex, b.NoteIndex)
	})
	return &Problems{issues}
}
func compareInt(a, b int) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}
func nextMain(es []Entry, i int) int {
	for j := i + 1; j < len(es); j++ {
		if !es[j].Event.MustMarks().Grace {
			return j
		}
	}
	return -1
}
func matching(src, dst []note.Note, ties []int) (map[int]int, error) {
	out := map[int]int{}
	used := make([]bool, len(dst))
	for _, i := range ties {
		found := -1
		equal := false
		for j, n := range dst {
			if src[i].Equal(n) {
				equal = true
				if !used[j] {
					found = j
					break
				}
			}
		}
		if found < 0 {
			if equal {
				return out, ErrTieMultiplicity
			}
			return out, ErrTiePitch
		}
		used[found] = true
		out[i] = found
	}
	return out, nil
}
func (t *Track) checkLinks(r record) []Issue {
	var out []Issue
	end, _, _ := t.grid.End()
	for i, e := range r.entries {
		m := e.Event.MustMarks()
		if !m.Grace && len(m.Ties) == 0 {
			continue
		}
		j := nextMain(r.entries, i)
		var dest *Entry
		if j >= 0 {
			dest = &r.entries[j]
		}
		if m.Grace && (dest == nil || dest.Event.Kind() == event.KindRest || !dest.Time.Start.Equal(e.Time.End)) {
			out = append(out, issue(ErrGraceTarget, e.Place))
			continue
		}
		if len(m.Ties) == 0 {
			continue
		}
		var cause error
		switch {
		case dest == nil:
			if e.Time.End.Equal(end) {
				cause = ErrTieEnd
			} else {
				cause = ErrTieGap
			}
		case !e.Time.End.Equal(dest.Time.Start):
			cause = ErrTieGap
		case dest.Event.Kind() == event.KindRest:
			cause = ErrTieRest
		}
		if cause != nil {
			for _, ni := range m.Ties {
				x := issue(cause, e.Place)
				x.NoteIndex = ni
				out = append(out, x)
			}
			continue
		}
		src := e.Event.MustContent().MustNotes()
		dst := dest.Event.MustContent().MustNotes()
		used := make([]bool, len(dst))
		for _, ni := range m.Ties {
			found := -1
			eq := false
			for k, n := range dst {
				if src[ni].Equal(n) {
					eq = true
					if !used[k] {
						found = k
						break
					}
				}
			}
			if found < 0 {
				cause = ErrTiePitch
				if eq {
					cause = ErrTieMultiplicity
				}
				x := issue(cause, e.Place)
				x.NoteIndex = ni
				out = append(out, x)
			} else {
				used[found] = true
			}
		}
	}
	return out
}
