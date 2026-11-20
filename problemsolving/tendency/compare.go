package tendency

import (
	"cmp"
	"slices"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/mode"
)

func schemesOK(ss []*Scheme) error {
	seen := map[ID]bool{}
	for _, s := range ss {
		if e := s.check(); e != nil {
			return e
		}
		if seen[s.state.info.ID] {
			return failure("duplicate comparison scheme ID", ErrInvalidArgument)
		}
		seen[s.state.info.ID] = true
	}
	return nil
}
func comparison[R any, V comparable](ss []*Scheme, query func(*Scheme) (R, error), values func(R) []V, status func(*Scheme, R, V) (Decision, bool, error), order func(V, V) int) (Comparison[R, V], error) {
	r := Comparison[R, V]{}
	if e := schemesOK(ss); e != nil {
		return r, e
	}
	all := map[V]bool{}
	for _, s := range ss {
		a, e := query(s)
		if e != nil {
			return Comparison[R, V]{}, e
		}
		r.Schemes = append(r.Schemes, s.Info())
		r.Answers = append(r.Answers, a)
		for _, v := range values(a) {
			all[v] = true
		}
	}
	for v := range all {
		item := Shared[V]{Value: v}
		for i, s := range ss {
			d, p, e := status(s, r.Answers[i], v)
			if e != nil {
				return Comparison[R, V]{}, e
			}
			switch d {
			case Allowed:
				item.Present = append(item.Present, i)
				if p {
					item.Preferred = append(item.Preferred, i)
				}
			case Denied:
				item.Absent = append(item.Absent, i)
			default:
				item.Unknown = append(item.Unknown, i)
			}
		}
		r.Items = append(r.Items, item)
	}
	slices.SortFunc(r.Items, func(a, b Shared[V]) int { return order(a.Value, b.Value) })
	return report(r), nil
}

// CompareSupports compares one support question across explicitly selected schemes.
func CompareSupports(m mode.Mode, n int, ss []*Scheme) (Comparison[SupportResult, SupportValue], error) {
	if e := numberOK(m, n); e != nil {
		return Comparison[SupportResult, SupportValue]{}, e
	}
	return comparison(ss, func(s *Scheme) (SupportResult, error) { return s.Support(m, n) }, func(r SupportResult) []SupportValue {
		if r.Knowledge == Known {
			return []SupportValue{r.Value}
		}
		return nil
	}, func(_ *Scheme, r SupportResult, v SupportValue) (Decision, bool, error) {
		if r.Knowledge != Known {
			return Undetermined, false, nil
		}
		if r.Value == v {
			return Allowed, false, nil
		}
		return Denied, false, nil
	}, func(a, b SupportValue) int {
		if c := cmp.Compare(a.Kind, b.Kind); c != 0 {
			return c
		}
		if c := cmp.Compare(boolInt(a.Ranked), boolInt(b.Ranked)); c != 0 {
			return c
		}
		return cmp.Compare(a.Rank, b.Rank)
	})
}

// CompareRoles compares one label question across explicitly selected schemes.
func CompareRoles(m mode.Mode, n int, ss []*Scheme) (Comparison[RoleResult, string], error) {
	if e := numberOK(m, n); e != nil {
		return Comparison[RoleResult, string]{}, e
	}
	return comparison(ss, func(s *Scheme) (RoleResult, error) { return s.Role(m, n) }, func(r RoleResult) []string {
		if r.Knowledge == Known {
			return []string{r.Role}
		}
		return nil
	}, func(_ *Scheme, r RoleResult, v string) (Decision, bool, error) {
		if r.Knowledge != Known {
			return Undetermined, false, nil
		}
		if r.Role == v {
			return Allowed, false, nil
		}
		return Denied, false, nil
	}, cmp.Compare[string])
}
func goalKey(t Target, match GoalMatch) GoalKey {
	if match == DegreeGoal {
		return GoalKey{Target{Degree: t.Degree}}
	}
	if !t.CycleKnown {
		t.Cycle = 0
	}
	return GoalKey{t}
}

// CompareTendencies compares goal mentions with exact or degree-only identity.
func CompareTendencies(m mode.Mode, p Place, ss []*Scheme, match GoalMatch) (Comparison[TendencyResult, GoalKey], error) {
	if e := ValidatePlace(m, p); e != nil {
		return Comparison[TendencyResult, GoalKey]{}, e
	}
	if match > DegreeGoal {
		return Comparison[TendencyResult, GoalKey]{}, failure("goal match", ErrInvalidArgument)
	}
	return comparison(ss, func(s *Scheme) (TendencyResult, error) { return s.Tendency(m, p) }, func(r TendencyResult) []GoalKey {
		v := []GoalKey{}
		for _, g := range r.Goals {
			v = append(v, goalKey(g.Target, match))
		}
		return v
	}, func(_ *Scheme, r TendencyResult, k GoalKey) (Decision, bool, error) {
		found, pref := false, false
		for _, g := range r.Goals {
			if goalKey(g.Target, match) == k {
				found = true
				pref = pref || g.Preferred
			}
		}
		if found {
			return Allowed, pref, nil
		}
		complete := r.GoalKnowledge == Known
		if match == ExactGoal {
			if k.Target.Kind == Immediate {
				complete = r.DirectKnowledge == Known
			} else {
				complete = r.GoalKnowledge == Known
				if r.Scheme == AMNeighborTargets || r.Scheme == SeredaTonal || r.Scheme == SeredaModal {
					complete = false
				}
			}
		}
		if complete {
			return Denied, false, nil
		}
		return Undetermined, false, nil
	}, func(a, b GoalKey) int { return compareTarget(a.Target, b.Target) })
}

// CompareClassifications compares independent table classifications for one input interval.
func CompareClassifications(i interval.Interval, ss []*Scheme) (Comparison[ClassificationResult, Classification], error) {
	if e := intervalOK(i); e != nil {
		return Comparison[ClassificationResult, Classification]{}, e
	}
	return comparison(ss, func(s *Scheme) (ClassificationResult, error) { return s.Classify(i) }, func(r ClassificationResult) []Classification {
		if r.Knowledge == Known {
			return []Classification{r.Value}
		}
		return nil
	}, func(_ *Scheme, r ClassificationResult, v Classification) (Decision, bool, error) {
		if r.Knowledge != Known {
			return Undetermined, false, nil
		}
		if r.Value == v {
			return Allowed, false, nil
		}
		return Denied, false, nil
	}, func(a, b Classification) int {
		if c := cmp.Compare(a.Kind, b.Kind); c != 0 {
			return c
		}
		if c := cmp.Compare(a.Subtype, b.Subtype); c != 0 {
			return c
		}
		return cmp.Compare(a.Caveat, b.Caveat)
	})
}

// CompareModalResolutions compares full target pairs and reuses each scheme’s checking rules.
func CompareModalResolutions(m mode.Mode, p Pair, ss []*Scheme) (Comparison[ModalResolutions, Pair], error) {
	if e := pairOK(m, p); e != nil {
		return Comparison[ModalResolutions, Pair]{}, e
	}
	return comparison(ss, func(s *Scheme) (ModalResolutions, error) { return s.ResolveModal(m, p) }, func(r ModalResolutions) []Pair {
		v := []Pair{}
		for _, list := range [][]ModalCandidate{r.Allowed, r.Undetermined} {
			for _, c := range list {
				v = append(v, c.Target)
			}
		}
		return v
	}, func(s *Scheme, r ModalResolutions, v Pair) (Decision, bool, error) {
		c, e := s.CheckModal(m, p, v)
		pref := false
		for _, x := range r.Allowed {
			if x.Target == v {
				pref = x.Preferred
			}
		}
		return c.Decision, pref, e
	}, comparePair)
}

// CompareIntervalResolutions compares exact movement pairs, preserving uncertain and forbidden cases.
func CompareIntervalResolutions(i interval.Interval, ss []*Scheme) (Comparison[IntervalResolutions, Motions], error) {
	if e := intervalOK(i); e != nil {
		return Comparison[IntervalResolutions, Motions]{}, e
	}
	return comparison(ss, func(s *Scheme) (IntervalResolutions, error) { return s.ResolveInterval(i) }, func(r IntervalResolutions) []Motions {
		v := []Motions{}
		for _, list := range [][]IntervalCandidate{r.Allowed, r.Undetermined} {
			for _, c := range list {
				v = append(v, c.Motions)
			}
		}
		return v
	}, func(s *Scheme, r IntervalResolutions, v Motions) (Decision, bool, error) {
		c, e := s.CheckInterval(i, v)
		pref := false
		for _, x := range r.Allowed {
			if x.Motions == v {
				pref = x.Preferred
			}
		}
		return c.Decision, pref, e
	}, compareMotions)
}
