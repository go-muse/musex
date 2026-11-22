package voiceleading

import "github.com/go-muse/musex/interval"

func validPerfect(perfect PerfectSet) bool {
	return perfect == PerfectFifths || perfect == PerfectOctaves || perfect == PerfectBoth
}

func family(i interval.Interval) PerfectSet {
	if !i.IsValid() || i.Quality() != interval.Perfect {
		return 0
	}
	switch (i.Size() - 1) % 7 {
	case 4:
		return PerfectFifths
	case 0:
		return PerfectOctaves
	default:
		return 0
	}
}

// consecutive reports a selected perfect family shared by both verticals.
func consecutive(t Transition, perfect PerfectSet) PerfectSet {
	if !t.From.Sounding || !t.To.Sounding {
		return 0
	}
	f := family(t.From.Specific)
	if f == family(t.To.Specific) && f&perfect != 0 {
		return f
	}
	return 0
}

func (a *Analysis) find(accept func(Transition) bool) []Transition {
	out := []Transition{}
	for _, p := range a.data {
		for _, t := range p.transitions {
			if accept(t) {
				out = append(out, t)
			}
		}
	}
	return out
}

// FindMotion finds transitions in pair/time order. MotionDirect includes
// MotionParallel; other valid kinds match exactly. MotionUnclassified and
// unknown values return ErrOption after checking that the receiver is set.
func (a *Analysis) FindMotion(kind Motion) ([]Transition, error) {
	if err := a.check(); err != nil {
		return nil, err
	}
	if kind < MotionStationary || kind > MotionParallel {
		return nil, ErrOption
	}
	return a.find(func(t Transition) bool { return t.Motion == kind || kind == MotionDirect && t.Motion.IsDirect() }), nil
}

// MustFindMotion is like FindMotion but panics with its error.
func (a *Analysis) MustFindMotion(kind Motion) []Transition { return must(a.FindMotion(kind)) }

// FindParallels finds parallel motion between equal full perfect intervals in
// selected families. Compound adds octave extensions; Unisons independently
// adds P1 to the octave family. The zero Perfect filter returns ErrOption.
func (a *Analysis) FindParallels(opts ParallelOptions) ([]Transition, error) {
	if err := a.check(); err != nil {
		return nil, err
	}
	if !validPerfect(opts.Perfect) {
		return nil, ErrOption
	}
	return a.find(func(t Transition) bool {
		if t.Motion != MotionParallel || consecutive(t, opts.Perfect) == 0 || t.From.Specific.Size() != t.To.Specific.Size() {
			return false
		}
		n := t.To.Specific.Size()
		return n == 5 || n == 8 || n == 1 && opts.Unisons || n > 8 && opts.Compound
	}), nil
}

// MustFindParallels is like FindParallels but panics with its error.
func (a *Analysis) MustFindParallels(opts ParallelOptions) []Transition {
	return must(a.FindParallels(opts))
}

// FindConsecutive finds adjacent perfect intervals of one selected family,
// including unisons, compound intervals, stationary and contrary motion.
func (a *Analysis) FindConsecutive(perfect PerfectSet) ([]Transition, error) {
	if err := a.check(); err != nil {
		return nil, err
	}
	if !validPerfect(perfect) {
		return nil, ErrOption
	}
	return a.find(func(t Transition) bool { return consecutive(t, perfect) != 0 }), nil
}

// MustFindConsecutive is like FindConsecutive but panics with its error.
func (a *Analysis) MustFindConsecutive(perfect PerfectSet) []Transition {
	return must(a.FindConsecutive(perfect))
}

// FindDirect finds direct or parallel arrivals in a selected perfect family,
// including unisons and compound intervals. DifferentFamily excludes perfect
// origins in that same family. Validation order is receiver, perfect, origin.
func (a *Analysis) FindDirect(perfect PerfectSet, origin DirectOrigin) ([]Transition, error) {
	if err := a.check(); err != nil {
		return nil, err
	}
	if !validPerfect(perfect) {
		return nil, ErrOption
	}
	if origin != AnyOrigin && origin != DifferentFamily {
		return nil, ErrOption
	}
	return a.find(func(t Transition) bool {
		f := family(t.To.Specific)
		return t.Motion.IsDirect() && f&perfect != 0 && (origin == AnyOrigin || family(t.From.Specific) != f)
	}), nil
}

// MustFindDirect is like FindDirect but panics with its error.
func (a *Analysis) MustFindDirect(perfect PerfectSet, origin DirectOrigin) []Transition {
	return must(a.FindDirect(perfect, origin))
}

func (a *Analysis) checkOrder() error {
	if err := a.check(); err != nil {
		return err
	}
	if a.order != LowToHigh {
		return ErrNoOrder
	}
	return nil
}

// FindCrossings returns maximal contiguous crossed runs for every pair, in
// pair/time order. Rest, equal height, and restored order end a run. Requires
// LowToHigh even when the music is empty. Nested Verticals slices are fresh.
func (a *Analysis) FindCrossings() ([]Crossing, error) {
	if err := a.checkOrder(); err != nil {
		return nil, err
	}
	out := []Crossing{}
	for _, p := range a.data {
		var run *Crossing
		for _, v := range p.verticals {
			if !v.Sounding || v.Chromatic >= 0 {
				run = nil
				continue
			}
			if run == nil {
				out = append(out, Crossing{Pair: p.pair, Start: v.Start, Verticals: []Vertical{}})
				run = &out[len(out)-1]
			}
			run.End = v.End
			run.Verticals = append(run.Verticals, v)
		}
	}
	return out, nil
}

// MustFindCrossings is like FindCrossings but panics with its error.
func (a *Analysis) MustFindCrossings() []Crossing { return must(a.FindCrossings()) }

// FindOverlaps checks only adjacent nominal parts. A part must start on its
// own side of the neighbor (equality allowed) and finish strictly past that
// neighbor's previous height. Existing crossings alone are not new overlaps.
// Requires LowToHigh even when the music is empty.
func (a *Analysis) FindOverlaps() ([]Overlap, error) {
	if err := a.checkOrder(); err != nil {
		return nil, err
	}
	out := []Overlap{}
	for _, p := range a.data {
		if p.pair.Second != p.pair.First+1 {
			continue
		}
		for _, t := range p.transitions {
			if !t.From.Sounding || !t.To.Sounding || t.From.Chromatic < 0 {
				continue
			}
			// Every pitch was admitted during Analyze. Compare is total on these
			// immutable pitches and, unlike subtracting heights, cannot overflow.
			first := a.sys.MustCompare(elementPitch(t.To.First), elementPitch(t.From.Second)) > 0
			second := a.sys.MustCompare(elementPitch(t.To.Second), elementPitch(t.From.First)) < 0
			if first || second {
				out = append(out, Overlap{Transition: t, FirstAbovePreviousSecond: first, SecondBelowPreviousFirst: second})
			}
		}
	}
	return out, nil
}

// MustFindOverlaps is like FindOverlaps but panics with its error.
func (a *Analysis) MustFindOverlaps() []Overlap { return must(a.FindOverlaps()) }

// CheckOMTFirstSpecies explicitly applies only OMT's consecutive-perfect-family
// rule to one pair. Sounding elements must have equal original starts and ends.
// Asynchronous boundaries produce Issues and no partial Violations, not an
// error. Empty sounding material is applicable with no findings. No nominal
// order is needed; this does not check all rules of first-species counterpoint.
func (a *Analysis) CheckOMTFirstSpecies(pair Pair) (*Assessment, error) {
	p, err := a.pair(pair)
	if err != nil {
		return nil, err
	}
	out := &Assessment{
		Rule:       OMTFirstSpeciesRule,
		Scope:      "two-part first-species counterpoint, note against note; consecutive perfect families only",
		Applicable: true,
		Issues:     []ScopeIssue{}, Violations: []Violation{},
	}
	for _, v := range p.verticals {
		if v.Sounding && (!v.First.Start.Equal(v.Second.Start) || !v.First.End.Equal(v.Second.End)) {
			out.Issues = append(out.Issues, ScopeIssue{Vertical: v, Reason: "asynchronous-note-boundaries"})
		}
	}
	if len(out.Issues) > 0 {
		out.Applicable = false
		return out, nil
	}
	for _, t := range p.transitions {
		f := consecutive(t, PerfectBoth)
		if f == 0 {
			continue
		}
		reason := "consecutive-perfect-fifths"
		if f == PerfectOctaves {
			reason = "consecutive-perfect-octaves"
		}
		out.Violations = append(out.Violations, Violation{Transition: t, Reason: reason})
	}
	return out, nil
}

// MustCheckOMTFirstSpecies is like CheckOMTFirstSpecies but panics with its error.
func (a *Analysis) MustCheckOMTFirstSpecies(pair Pair) *Assessment {
	return must(a.CheckOMTFirstSpecies(pair))
}
