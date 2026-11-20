package tendency

import (
	"slices"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/mode"
)

type filterVerdict struct {
	fail, unknown bool
	reasons       []Reason
}

func (v *filterVerdict) add(pass, known bool, rule, text string) {
	if !known {
		v.unknown = true
	} else if !pass {
		v.fail = true
	}
	v.reasons = append(v.reasons, productReason(rule, text))
}
func (v *filterVerdict) targets(t Target, mask []Place) {
	if mask == nil {
		return
	}
	pass := false
	degree := false
	for _, p := range mask {
		if p.Degree == t.Degree {
			degree = true
			if t.CycleKnown && p.Cycle == t.Cycle {
				pass = true
			}
		}
	}
	known := t.CycleKnown || !degree || len(mask) == 0
	v.add(pass, known, "filter.target", "Compare the target against the requested finite place mask.")
}
func (v *filterVerdict) direction(d Motion, mask []Motion) {
	if mask == nil {
		return
	}
	v.add(slices.Contains(mask, d), d != MotionUnset || len(mask) == 0, "filter.direction", "Compare the movement direction against the requested direction mask.")
}
func (v *filterVerdict) support(s *Scheme, m mode.Mode, n int) error {
	if s == nil {
		return nil
	}
	r, e := s.Support(m, n)
	if e != nil {
		return e
	}
	v.add(r.Value.Kind == MainSupport || r.Value.Kind == RelativeSupport, r.Knowledge == Known, "filter.support", "Check whether the selected support scheme confirms a final support.")
	return nil
}
func validateMasks(m mode.Mode, places ...[]Place) error {
	for _, mask := range places {
		for _, p := range mask {
			if e := ValidatePlace(m, p); e != nil {
				return e
			}
		}
	}
	return nil
}
func validateDirections(masks ...[]Motion) error {
	for _, mask := range masks {
		for _, d := range mask {
			if !validMotion(d) {
				return failure("direction filter", ErrInvalidArgument)
			}
		}
	}
	return nil
}
func validateSupport(s *Scheme) error {
	if s != nil {
		return s.check()
	}
	return nil
}
func validateKeep(k Keep) error {
	if k > KeepBoth {
		return failure("keep filter", ErrInvalidArgument)
	}
	return nil
}
func selectValue[R, V any](r *Selection[R, V], v V, filter filterVerdict, uncertain bool) {
	r.Reasons = append(r.Reasons, filter.reasons...)
	if filter.fail {
		r.Excluded = append(r.Excluded, Excluded[V]{copied(v), canonicalReasons(filter.reasons)})
	} else if uncertain || filter.unknown {
		r.Uncertain = append(r.Uncertain, copied(v))
	} else {
		r.Matched = append(r.Matched, copied(v))
	}
}

// SelectGoals filters listed goals without strengthening their original musical claims.
func (s *Scheme) SelectGoals(m mode.Mode, p Place, f GoalFilter) (Selection[TendencyResult, Goal], error) {
	if e := s.check(); e != nil {
		return Selection[TendencyResult, Goal]{}, e
	}
	if e := ValidatePlace(m, p); e != nil {
		return Selection[TendencyResult, Goal]{}, e
	}
	if e := validateMasks(m, f.Targets); e != nil {
		return Selection[TendencyResult, Goal]{}, e
	}
	if e := validateDirections(f.Directions); e != nil {
		return Selection[TendencyResult, Goal]{}, e
	}
	if e := validateSupport(f.Support); e != nil {
		return Selection[TendencyResult, Goal]{}, e
	}
	original, e := s.Tendency(m, p)
	if e != nil {
		return Selection[TendencyResult, Goal]{}, e
	}
	r := Selection[TendencyResult, Goal]{Original: original}
	for _, g := range original.Goals {
		v := filterVerdict{}
		v.targets(g.Target, f.Targets)
		v.direction(g.Direction, f.Directions)
		if e := v.support(f.Support, m, g.Target.Degree); e != nil {
			return Selection[TendencyResult, Goal]{}, e
		}
		selectValue(&r, g, v, false)
	}
	r.Reasons = canonicalReasons(r.Reasons)
	return report(r), nil
}

// SelectModal filters allowed and uncertain candidates while preserving the original report.
func (s *Scheme) SelectModal(m mode.Mode, p Pair, f ModalFilter) (Selection[ModalResolutions, ModalCandidate], error) {
	if e := s.check(); e != nil {
		return Selection[ModalResolutions, ModalCandidate]{}, e
	}
	if e := pairOK(m, p); e != nil {
		return Selection[ModalResolutions, ModalCandidate]{}, e
	}
	if e := validateMasks(m, f.FirstTargets, f.SecondTargets); e != nil {
		return Selection[ModalResolutions, ModalCandidate]{}, e
	}
	if e := validateDirections(f.FirstDirections, f.SecondDirections); e != nil {
		return Selection[ModalResolutions, ModalCandidate]{}, e
	}
	if e := validateKeep(f.Keep); e != nil {
		return Selection[ModalResolutions, ModalCandidate]{}, e
	}
	if e := validateSupport(f.Support); e != nil {
		return Selection[ModalResolutions, ModalCandidate]{}, e
	}
	original, e := s.ResolveModal(m, p)
	if e != nil {
		return Selection[ModalResolutions, ModalCandidate]{}, e
	}
	r := Selection[ModalResolutions, ModalCandidate]{Original: original}
	for listIndex, list := range [][]ModalCandidate{original.Allowed, original.Undetermined} {
		for _, c := range list {
			v := filterVerdict{}
			v.targets(Target{Immediate, c.Target.First.Degree, c.Target.First.Cycle, true}, f.FirstTargets)
			v.targets(Target{Immediate, c.Target.Second.Degree, c.Target.Second.Cycle, true}, f.SecondTargets)
			v.direction(c.First.Direction, f.FirstDirections)
			v.direction(c.Second.Direction, f.SecondDirections)
			if f.Keep == KeepFirst || f.Keep == KeepBoth {
				v.add(c.First.Direction == Stationary, true, "filter.keep", "Keep the first voice stationary.")
			}
			if f.Keep == KeepSecond || f.Keep == KeepBoth {
				v.add(c.Second.Direction == Stationary, true, "filter.keep", "Keep the second voice stationary.")
			}
			if e := v.support(f.Support, m, c.Target.First.Degree); e != nil {
				return Selection[ModalResolutions, ModalCandidate]{}, e
			}
			if e := v.support(f.Support, m, c.Target.Second.Degree); e != nil {
				return Selection[ModalResolutions, ModalCandidate]{}, e
			}
			selectValue(&r, c, v, listIndex == 1)
		}
	}
	r.Reasons = canonicalReasons(r.Reasons)
	return report(r), nil
}

// SelectInterval filters voice movements while preserving musical uncertainty.
func (s *Scheme) SelectInterval(i interval.Interval, f IntervalFilter) (Selection[IntervalResolutions, IntervalCandidate], error) {
	if e := s.check(); e != nil {
		return Selection[IntervalResolutions, IntervalCandidate]{}, e
	}
	if e := intervalOK(i); e != nil {
		return Selection[IntervalResolutions, IntervalCandidate]{}, e
	}
	if e := validateDirections(f.FirstDirections, f.SecondDirections); e != nil {
		return Selection[IntervalResolutions, IntervalCandidate]{}, e
	}
	if e := validateKeep(f.Keep); e != nil {
		return Selection[IntervalResolutions, IntervalCandidate]{}, e
	}
	original, e := s.ResolveInterval(i)
	if e != nil {
		return Selection[IntervalResolutions, IntervalCandidate]{}, e
	}
	r := Selection[IntervalResolutions, IntervalCandidate]{Original: original}
	for listIndex, list := range [][]IntervalCandidate{original.Allowed, original.Undetermined} {
		for _, c := range list {
			v := filterVerdict{}
			v.direction(motion(c.Motions.First), f.FirstDirections)
			v.direction(motion(c.Motions.Second), f.SecondDirections)
			if f.Keep == KeepFirst || f.Keep == KeepBoth {
				v.add(motion(c.Motions.First) == Stationary, true, "filter.keep", "Keep the first voice stationary.")
			}
			if f.Keep == KeepSecond || f.Keep == KeepBoth {
				v.add(motion(c.Motions.Second) == Stationary, true, "filter.keep", "Keep the second voice stationary.")
			}
			selectValue(&r, c, v, listIndex == 1)
		}
	}
	r.Reasons = canonicalReasons(r.Reasons)
	return report(r), nil
}
