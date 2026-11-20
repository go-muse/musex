package tendency

import (
	"slices"
	"strconv"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/mode"
)

// Classify classifies a specific interval using only the selected table coverage.
func (s *Scheme) Classify(i interval.Interval) (ClassificationResult, error) {
	if e := s.check(); e != nil {
		return ClassificationResult{}, e
	}
	if e := intervalOK(i); e != nil {
		return ClassificationResult{}, e
	}
	v := i.MustWithDirection(interval.Up)
	r := ClassificationResult{Input: i, Classified: v}
	if !slices.Contains(s.state.info.Tasks, ClassificationTask) {
		r.Answer = s.unsupported()
		return report(r), nil
	}
	id := s.state.info.ID
	if id == OMTCounterpoint || id == MT21Consonance || id == SeredaConsonance {
		v = v.MustSimple()
		r.Classified = v
	}
	text := v.String()
	basic := slices.Contains([]string{"P1", "m2", "M2", "m3", "M3", "P4", "A4", "d5", "P5", "m6", "M6", "m7", "M7", "P8"}, text)
	kind := AcousticUnset
	sub := ConsonanceUnset
	caveat := ""
	if basic {
		kind = Dissonance
		cons := slices.Contains([]string{"P1", "m3", "M3", "P5", "m6", "M6", "P8"}, text)
		switch id {
		case AMConsonance, WPConsonance:
			cons = cons || text == "P4"
		case OMTMelodic:
			cons = cons || text == "P4" || text == "m2" || text == "M2"
		case SeredaConsonance:
			cons = slices.Contains([]string{"m3", "M3", "m6", "M6"}, text)
			if slices.Contains([]string{"P1", "P5", "P8", "A4", "d5"}, text) {
				kind = AcousticUnset
			}
		}
		if cons {
			kind = Consonance
			switch id {
			case AMConsonance:
				switch text {
				case "P1", "P8":
					sub = AbsoluteConsonance
				case "P4", "P5":
					sub = PerfectConsonance
				default:
					sub = ImperfectConsonance
				}
				if text == "P4" {
					caveat = "The fourth is only partly consonant: in strict two-voice writing and above the bass it is dissonant, while upper-voice context can be consonant."
				}
			case OMTCounterpoint, WPConsonance:
				sub = ImperfectConsonance
				if text == "P1" || text == "P4" || text == "P5" || text == "P8" {
					sub = PerfectConsonance
				}
			case SeredaConsonance:
				sub = ImperfectConsonance
			}
		}
	} else if (id == OMTHarmonic || id == OMTMelodic || id == OMTCounterpoint) && v.IsSimple() && v.IsIntegerQuality() && v.IsChromatic() {
		kind = Dissonance
	}
	if kind == AcousticUnset {
		r.Answer = s.missing()
		r.Reasons = append(r.Reasons, sourceReason(string(id), "consonance-dissonance §6", "The selected classification table does not cover this interval.", Fact{"interval", v.String()}))
	} else {
		r.Answer = s.answer(Known, sourceReason(string(id), "consonance-dissonance §6 row "+text, "The selected table explicitly classifies this named interval.", Fact{"interval", text}))
		r.Value = Classification{kind, sub, caveat}
	}
	return report(r), nil
}
func makeModalCandidate(m mode.Mode, from, to Pair, preferred bool, rs []Reason) (ModalCandidate, error) {
	a, e := Between(m, from.First, to.First)
	if e != nil {
		return ModalCandidate{}, e
	}
	b, e := Between(m, from.Second, to.Second)
	if e != nil {
		return ModalCandidate{}, e
	}
	v, e := Between(m, to.First, to.Second)
	if e != nil {
		return ModalCandidate{}, e
	}
	return ModalCandidate{to, a, b, v.Interval, preferred, canonicalReasons(rs)}, nil
}
func modalRuleFor(s *Scheme, p Pair) (*ModalRule, int, error) {
	for i := range s.state.def.Modal {
		r := &s.state.def.Modal[i]
		if !r.Repeat {
			if r.Source == p {
				return r, 0, nil
			}
			continue
		}
		x, e := normalizedPair(p)
		if e != nil {
			return nil, 0, e
		}
		if x == r.Source {
			return r, p.First.Cycle, nil
		}
	}
	return nil, 0, nil
}

// ResolveModal enumerates two-voice target pairs without swapping voice identities.
func (s *Scheme) ResolveModal(m mode.Mode, p Pair) (ModalResolutions, error) {
	if e := s.check(); e != nil {
		return ModalResolutions{}, e
	}
	if e := pairOK(m, p); e != nil {
		return ModalResolutions{}, e
	}
	v, e := Between(m, p.First, p.Second)
	if e != nil {
		return ModalResolutions{}, e
	}
	r := ModalResolutions{Answer: s.applicability(m, ModalResolutionTask), Source: p, Initial: v.Interval}
	if r.Knowledge != Known {
		return report(r), nil
	}
	if !s.state.info.Builtin {
		row, offset, e := modalRuleFor(s, p)
		if e != nil {
			return ModalResolutions{}, e
		}
		if row == nil {
			r.Answer = s.missing()
			return report(r), nil
		}
		r.Answer = s.answer(Known, bases(row.Bases)...)
		for _, o := range row.Options {
			target := o.Target
			if row.Repeat {
				target, e = shiftedPair(target, offset)
				if e != nil {
					return ModalResolutions{}, e
				}
			}
			c, e := makeModalCandidate(m, p, target, o.Preferred, append(bases(row.Bases), bases(o.Bases)...))
			if e != nil {
				return ModalResolutions{}, e
			}
			r.Allowed = append(r.Allowed, c)
		}
		return report(r), nil
	}
	if v.Interval.IsDescending() || v.Interval.Size() > 8 {
		r.Answer = s.answer(OutsideScope, productReason("scope.interval", "AM modal resolution requires an ascending simple initial interval through P8."))
		return report(r), nil
	}
	if supportDegree(p.First.Degree) && supportDegree(p.Second.Degree) {
		r.Answer = s.answer(Known, productReason("resolution.not-required", "Both source voices are supports; no modal resolution is required."))
		return report(r), nil
	}
	choices := func(p Place) ([]Place, error) {
		if supportDegree(p.Degree) {
			return []Place{p}, nil
		}
		t, e := MustBuiltin(AMNeighborTargets).Tendency(m, p)
		if e != nil {
			return nil, e
		}
		v := []Place{}
		for _, g := range t.Goals {
			v = append(v, Place{g.Target.Degree, g.Target.Cycle})
		}
		return v, nil
	}
	first, e := choices(p.First)
	if e != nil {
		return ModalResolutions{}, e
	}
	second, e := choices(p.Second)
	if e != nil {
		return ModalResolutions{}, e
	}
	rs := []Reason{}
	for _, a := range first {
		for _, b := range second {
			q := Pair{a, b}
			c, e := makeModalCandidate(m, p, q, false, nil)
			if e != nil {
				return ModalResolutions{}, e
			}
			check, e := s.checkAMModal(m, p, c)
			if e != nil {
				return ModalResolutions{}, e
			}
			c.Reasons = check.Reasons
			switch check.Decision {
			case Allowed:
				c.Preferred = preferredModal(m, p, c)
				r.Allowed = append(r.Allowed, c)
			case Denied:
				r.Rejected = append(r.Rejected, c)
			default:
				r.Undetermined = append(r.Undetermined, c)
			}
			rs = append(rs, check.Reasons...)
		}
	}
	k := Known
	if len(r.Undetermined) > 0 {
		k = Partial
	}
	r.Answer = s.answer(k, rs...)
	sortModal(&r)
	return report(r), nil
}
func sortModal(r *ModalResolutions) {
	for _, list := range [][]ModalCandidate{r.Allowed, r.Undetermined, r.Rejected} {
		slices.SortFunc(list, func(a, b ModalCandidate) int { return comparePair(a.Target, b.Target) })
	}
}
func preferredModal(m mode.Mode, p Pair, c ModalCandidate) bool {
	for _, x := range []struct {
		from Place
		mv   Movement
	}{{p.First, c.First}, {p.Second, c.Second}} {
		if x.mv.Direction == Stationary {
			continue
		}
		t, e := MustBuiltin(AMNeighborTargets).Tendency(m, x.from)
		if e != nil {
			return false
		}
		ok := false
		for _, g := range t.Goals {
			if g.Preferred && g.Target.Degree == x.mv.To.Degree && g.Target.Cycle == x.mv.To.Cycle {
				ok = true
			}
		}
		if !ok {
			return false
		}
	}
	return true
}
func (s *Scheme) checkAMModal(m mode.Mode, p Pair, c ModalCandidate) (CheckResult, error) {
	initial, e := Between(m, p.First, p.Second)
	if e != nil {
		return CheckResult{}, e
	}
	rs := []Reason{sourceReason("am.modal-resolution", "interval-resolution op-3", "Apply all AM modal resolution conditions to the two identified voices.", Fact{"initial", initial.Interval.String()}, Fact{"final", c.Final.String()})}
	denied := false
	unknown := false
	fail := func(rule, text string) {
		denied = true
		rs = append(rs, sourceReason(rule, "interval-resolution inv-1/inv-6…inv-11", text))
	}
	gap := func(rule, text string) { unknown = true; rs = append(rs, productReason(rule, text)) }
	if supportDegree(p.First.Degree) && supportDegree(p.Second.Degree) {
		fail("resolution.not-required", "Both voices already occupy supports; this task has no resolution.")
	}
	if comparePlace(c.Target.First, c.Target.Second) > 0 {
		fail("am.crossing", "The final first voice is above the second voice.")
	}
	for _, x := range []struct {
		from  Place
		mv    Movement
		voice string
	}{{p.First, c.First, "First"}, {p.Second, c.Second, "Second"}} {
		if !supportDegree(x.mv.To.Degree) {
			fail("am.non-support", "The final "+x.voice+" degree is not an AM support.")
		}
		if supportDegree(x.from.Degree) {
			if x.from != x.mv.To {
				fail("am.non-support", "An initially supported "+x.voice+" voice must remain stationary.")
			}
		} else {
			t, e := MustBuiltin(AMNeighborTargets).Tendency(m, x.from)
			if e != nil {
				return CheckResult{}, e
			}
			ok := false
			for _, g := range t.Goals {
				if g.Target.Degree == x.mv.To.Degree && g.Target.Cycle == x.mv.To.Cycle {
					ok = true
				}
			}
			if !ok {
				fail("am.non-support", "The "+x.voice+" target is not a neighboring support.")
			}
		}
	}
	cr, e := MustBuiltin(AMConsonance).Classify(c.Final)
	if e != nil {
		return CheckResult{}, e
	}
	rs = append(rs, cr.Reasons...)
	if cr.Knowledge != Known {
		gap("am.classification-missing", "The AM table does not classify the final interval.")
	} else if cr.Value.Kind != Consonance {
		fail("am.dissonance", "The final interval is an AM dissonance.")
	}
	init := initial.Interval
	simple := init.MustWithDirection(interval.Up).MustSimple()
	finSimple := c.Final.MustWithDirection(interval.Up).MustSimple()
	movingFirst := c.First.Direction != Stationary
	movingSecond := c.Second.Direction != Stationary
	if init.Size() == 2 && c.Final.IsPerfectUnison() {
		fail("am.second-to-unison", "A simple second may not resolve directly to a unison.")
	}
	if init.Size() == 7 && c.Final.Equal(interval.PerfectOctave()) {
		fail("am.seventh-to-octave", "A simple seventh may not resolve directly to a perfect octave.")
	}
	if movingFirst && movingSecond && c.First.Direction == c.Second.Direction {
		if simple.Equal(interval.PerfectFifth()) && finSimple.Equal(interval.PerfectFifth()) {
			fail("am.parallel-fifths", "Both voices move in the same direction in parallel perfect fifths.")
		}
		if simple.Equal(interval.PerfectOctave()) && finSimple.Equal(interval.PerfectOctave()) {
			fail("am.parallel-octaves", "Both voices move in the same direction in parallel perfect octaves.")
		}
	}
	if init.Equal(interval.AugmentedFourth()) {
		if c.First.Direction != Descending || c.Second.Direction != Ascending {
			fail("am.tritone-motion", "An augmented fourth requires both voices to move outward.")
		}
		if !c.Final.Equal(interval.MinorSixth()) && !c.Final.Equal(interval.MajorSixth()) {
			fail("am.tritone-target", "An augmented fourth requires a minor or major sixth as its target.")
		}
	}
	if init.Equal(interval.DiminishedFifth()) {
		if c.First.Direction != Ascending || c.Second.Direction != Descending {
			fail("am.tritone-motion", "A diminished fifth requires both voices to move inward.")
		}
		if !c.Final.Equal(interval.MinorThird()) && !c.Final.Equal(interval.MajorThird()) {
			fail("am.tritone-target", "A diminished fifth requires a minor or major third as its target.")
		}
	}
	if movingFirst && movingSecond && c.Final.IsPerfectUnison() && c.First.Direction != c.Second.Direction {
		gap("am.converging-unison-gap", "AM applicability to two moving voices converging to unison is not established (gap-1).")
	}
	d := Allowed
	if unknown {
		d = Undetermined
	}
	if denied {
		d = Denied
	}
	k := Known
	if unknown {
		k = Partial
	}
	return report(CheckResult{s.answer(k, rs...), d}), nil
}

// CheckModal checks the same local and joint conditions used for modal enumeration.
func (s *Scheme) CheckModal(m mode.Mode, from, to Pair) (CheckResult, error) {
	if e := s.check(); e != nil {
		return CheckResult{}, e
	}
	if e := pairOK(m, from); e != nil {
		return CheckResult{}, e
	}
	if e := pairOK(m, to); e != nil {
		return CheckResult{}, e
	}
	r, e := s.ResolveModal(m, from)
	if e != nil {
		return CheckResult{}, e
	}
	if r.Knowledge != Known && r.Knowledge != Partial {
		return report(CheckResult{r.Answer, Undetermined}), nil
	}
	if s.state.info.Builtin {
		c, e := makeModalCandidate(m, from, to, false, nil)
		if e != nil {
			return CheckResult{}, e
		}
		return s.checkAMModal(m, from, c)
	}
	for _, c := range r.Allowed {
		if c.Target == to {
			return report(CheckResult{s.answer(Known, c.Reasons...), Allowed}), nil
		}
	}
	return report(CheckResult{s.answer(Known, productReason("resolution.not-listed", "This target pair is absent from the complete user row.")), Denied}), nil
}

// ResolveInterval enumerates exact movement combinations for the specific source interval.
func (s *Scheme) ResolveInterval(i interval.Interval) (IntervalResolutions, error) {
	if e := s.check(); e != nil {
		return IntervalResolutions{}, e
	}
	if e := intervalOK(i); e != nil {
		return IntervalResolutions{}, e
	}
	r := IntervalResolutions{Initial: i}
	if !slices.Contains(s.state.info.Tasks, IntervalResolutionTask) {
		r.Answer = s.unsupported()
		return report(r), nil
	}
	if !s.state.info.Builtin {
		r.Answer = s.missing()
		for _, row := range s.state.def.Intervals {
			if row.Source != i {
				continue
			}
			r.Answer = s.answer(Known, bases(row.Bases)...)
			for _, o := range row.Options {
				v, e := intervalFinal(i, o.Motions)
				if e != nil {
					return IntervalResolutions{}, e
				}
				r.Allowed = append(r.Allowed, IntervalCandidate{o.Motions, v, o.Preferred, canonicalReasons(append(bases(row.Bases), bases(o.Bases)...))})
			}
			break
		}
		return report(r), nil
	}
	r.Answer = s.answer(OutsideScope, productReason("scope.interval", "AM interval resolution covers only ascending simple basic dissonances and its known consonances."))
	if i.IsDescending() || i.Size() > 8 {
		return report(r), nil
	}
	cr, e := MustBuiltin(AMConsonance).Classify(i)
	if e != nil {
		return IntervalResolutions{}, e
	}
	if cr.Knowledge == Known && cr.Value.Kind == Consonance {
		r.Answer = s.answer(Known, productReason("resolution.not-required", "This AM consonance does not require acoustic resolution."))
		return report(r), nil
	}
	text := i.String()
	if !slices.Contains([]string{"m2", "M2", "m7", "M7", "A4", "d5"}, text) {
		return report(r), nil
	}
	moves := []Motions{}
	p1 := interval.PerfectUnison()
	seconds := []interval.Interval{interval.MinorSecond(), interval.MajorSecond()}
	for _, a := range seconds {
		switch i.Size() {
		case 2:
			moves = append(moves, Motions{a.MustWithDirection(interval.Down), p1})
		case 7:
			moves = append(moves, Motions{p1, a.MustWithDirection(interval.Down)})
		default:
			for _, b := range seconds {
				if text == "A4" {
					moves = append(moves, Motions{a.MustWithDirection(interval.Down), b})
				} else {
					moves = append(moves, Motions{a, b.MustWithDirection(interval.Down)})
				}
			}
		}
	}
	rs := []Reason{}
	for _, mv := range moves {
		v, e := intervalFinal(i, mv)
		if e != nil {
			return IntervalResolutions{}, e
		}
		ok := false
		if i.Size() == 2 || text == "d5" {
			ok = v.Equal(interval.MinorThird()) || v.Equal(interval.MajorThird())
		} else {
			ok = v.Equal(interval.MinorSixth()) || v.Equal(interval.MajorSixth())
		}
		rule := "am.interval-resolution"
		why := "The exact voice movements produce the required consonant third or sixth."
		if !ok {
			rule = "am.tritone-target"
			why = "The exact voice movements do not produce the required minor or major third/sixth."
		}
		rr := []Reason{sourceReason(rule, "interval-resolution inv-3/inv-4/inv-6 op-2", why, Fact{"initial", text}, Fact{"final", v.String()}, Fact{"first", mv.First.String()}, Fact{"second", mv.Second.String()})}
		c := IntervalCandidate{mv, v, false, rr}
		if ok {
			r.Allowed = append(r.Allowed, c)
		} else {
			r.Rejected = append(r.Rejected, c)
		}
		rs = append(rs, rr...)
	}
	r.Answer = s.answer(Known, rs...)
	for _, list := range [][]IntervalCandidate{r.Allowed, r.Rejected} {
		slices.SortFunc(list, func(a, b IntervalCandidate) int { return compareMotions(a.Motions, b.Motions) })
	}
	return report(r), nil
}

// CheckInterval checks the two specific movements, rather than their final interval alone.
func (s *Scheme) CheckInterval(i interval.Interval, mv Motions) (CheckResult, error) {
	if e := s.check(); e != nil {
		return CheckResult{}, e
	}
	if e := intervalOK(i); e != nil {
		return CheckResult{}, e
	}
	if e := motionsOK(mv); e != nil {
		return CheckResult{}, e
	}
	if _, e := intervalFinal(i, mv); e != nil {
		return CheckResult{}, e
	}
	r, e := s.ResolveInterval(i)
	if e != nil {
		return CheckResult{}, e
	}
	for _, c := range r.Allowed {
		if c.Motions == mv {
			return report(CheckResult{s.answer(Known, c.Reasons...), Allowed}), nil
		}
	}
	if r.Knowledge != Known {
		return report(CheckResult{r.Answer, Undetermined}), nil
	}
	return report(CheckResult{s.answer(Known, productReason("resolution.not-listed", "The exact voice movements are not listed for source interval "+i.String()+" (size "+strconv.Itoa(i.Size())+").")), Denied}), nil
}
