package tendency

import (
	"cmp"
	"slices"
	"strconv"
	"strings"

	"github.com/go-muse/musex/modal"
	"github.com/go-muse/musex/mode"
)

// Support answers support for one actual degree under the selected scheme.
func (s *Scheme) Support(m mode.Mode, n int) (SupportResult, error) {
	a, e := s.operation(m, SupportTask, n)
	if e != nil {
		return SupportResult{}, e
	}
	r := SupportResult{Answer: a, Degree: n}
	if a.Knowledge != Known {
		return report(r), nil
	}
	if s.state.info.Builtin {
		v := SupportValue{Kind: Unstable}
		switch n {
		case 1:
			v = SupportValue{MainSupport, 1, true}
		case 5:
			v = SupportValue{RelativeSupport, 2, true}
		case 3:
			v = SupportValue{RelativeSupport, 3, true}
		}
		r.Value = v
		r.Reasons = []Reason{sourceReason("am.supports", "stability inv-2/inv-3", "AM assigns I as the main support, III and V as relative supports, ranked I, V, III.", Fact{"degree", strconv.Itoa(n)})}
	} else {
		r.Answer = s.missing()
		for _, row := range s.state.def.Supports {
			if row.Degree == n {
				r.Answer = s.answer(Known, bases(row.Bases)...)
				r.Value = row.Value
				break
			}
		}
	}
	return report(r), nil
}

// Supports returns every support row, known groups, and ranked groups.
func (s *Scheme) Supports(m mode.Mode) (SupportAnalysis, error) {
	if e := s.check(); e != nil {
		return SupportAnalysis{}, e
	}
	if e := modeOK(m); e != nil {
		return SupportAnalysis{}, e
	}
	r := SupportAnalysis{Groups: []SupportGroup{{Kind: MainSupport}, {Kind: RelativeSupport}, {Kind: Unstable}}}
	ks := []Knowledge{}
	rs := []Reason{}
	ranks := map[int][]int{}
	for n := 1; n <= m.Len(); n++ {
		x, e := s.Support(m, n)
		if e != nil {
			return SupportAnalysis{}, e
		}
		r.Rows = append(r.Rows, x)
		ks = append(ks, x.Knowledge)
		rs = append(rs, x.Reasons...)
		if x.Knowledge == Known {
			for i := range r.Groups {
				if r.Groups[i].Kind == x.Value.Kind {
					r.Groups[i].Degrees = append(r.Groups[i].Degrees, n)
				}
			}
			if x.Value.Ranked {
				ranks[x.Value.Rank] = append(ranks[x.Value.Rank], n)
			}
		}
	}
	for rank, ns := range ranks {
		r.Order = append(r.Order, RankGroup{rank, ns})
	}
	slices.SortFunc(r.Order, func(a, b RankGroup) int { return cmp.Compare(a.Rank, b.Rank) })
	r.Answer = s.answer(totalKnowledge(ks), rs...)
	return report(r), nil
}

// CompareSupport compares ranked supports; unknown ranking remains explicitly unknown.
func (s *Scheme) CompareSupport(m mode.Mode, a, b int) (RankResult, error) {
	if e := s.check(); e != nil {
		return RankResult{}, e
	}
	if e := numberOK(m, a); e != nil {
		return RankResult{}, e
	}
	if e := numberOK(m, b); e != nil {
		return RankResult{}, e
	}
	x, e := s.Support(m, a)
	if e != nil {
		return RankResult{}, e
	}
	y, e := s.Support(m, b)
	if e != nil {
		return RankResult{}, e
	}
	k := totalKnowledge([]Knowledge{x.Knowledge, y.Knowledge})
	r := RankResult{First: a, Second: b}
	rs := append(x.Reasons, y.Reasons...)
	if k == Known {
		if x.Value.Ranked && y.Value.Ranked {
			r.Order = cmp.Compare(x.Value.Rank, y.Value.Rank)
		} else {
			k = MissingRule
			rs = append(rs, productReason("coverage.row", "Both supports need known ranks for an order comparison."))
		}
	}
	r.Answer = s.answer(k, rs...)
	return report(r), nil
}

// Role returns the case-sensitive label of one actual degree.
func (s *Scheme) Role(m mode.Mode, n int) (RoleResult, error) {
	a, e := s.operation(m, RoleTask, n)
	if e != nil {
		return RoleResult{}, e
	}
	r := RoleResult{Answer: a, Degree: n}
	if a.Knowledge != Known {
		return report(r), nil
	}
	if s.state.info.Builtin {
		labels := []string{"T", "", "M", "S", "D", "SM", ""}
		r.Role = labels[n-1]
		if s.state.info.ID == AMMinorRoles {
			r.Role = strings.ToLower(r.Role)
		}
		r.Reasons = []Reason{sourceReason("am.roles", "tonal-function functional degree letters", "The selected natural major/minor table supplies this degree label.", Fact{"degree", strconv.Itoa(n)}, Fact{"role", r.Role})}
	} else {
		r.Answer = s.missing()
		for _, row := range s.state.def.Roles {
			if row.Degree == n {
				r.Answer = s.answer(Known, bases(row.Bases)...)
				r.Role = row.Role
				break
			}
		}
	}
	return report(r), nil
}

// Roles returns every label row and sorted known groups.
func (s *Scheme) Roles(m mode.Mode) (RoleAnalysis, error) {
	if e := s.check(); e != nil {
		return RoleAnalysis{}, e
	}
	if e := modeOK(m); e != nil {
		return RoleAnalysis{}, e
	}
	r := RoleAnalysis{}
	ks := []Knowledge{}
	rs := []Reason{}
	groups := map[string][]int{}
	for n := 1; n <= m.Len(); n++ {
		x, e := s.Role(m, n)
		if e != nil {
			return RoleAnalysis{}, e
		}
		r.Rows = append(r.Rows, x)
		ks = append(ks, x.Knowledge)
		rs = append(rs, x.Reasons...)
		if x.Knowledge == Known {
			groups[x.Role] = append(groups[x.Role], n)
		}
	}
	for role, ns := range groups {
		r.Groups = append(r.Groups, RoleGroup{role, ns})
	}
	slices.SortFunc(r.Groups, func(a, b RoleGroup) int { return cmp.Compare(a.Role, b.Role) })
	r.Answer = s.answer(totalKnowledge(ks), rs...)
	return report(r), nil
}
func neighbor(m mode.Mode, p Place, d Motion) (Place, bool, error) {
	n := p.Degree
	if d == Ascending {
		n++
	} else {
		n--
	}
	c := p.Cycle
	if n < 1 {
		if m.IsOpen() {
			return Place{}, false, nil
		}
		var e error
		c, e = intSum(c, -1)
		if e != nil {
			return Place{}, false, e
		}
		n = m.Len()
	}
	if n > m.Len() {
		if m.IsOpen() {
			return Place{}, false, nil
		}
		var e error
		c, e = intSum(c, 1)
		if e != nil {
			return Place{}, false, e
		}
		n = 1
	}
	q := Place{n, c}
	if e := ValidatePlace(m, q); e != nil {
		return Place{}, false, e
	}
	return q, true, nil
}
func supportDegree(n int) bool { return n == 1 || n == 3 || n == 5 }
func finishTendency(s *Scheme, r TendencyResult) TendencyResult {
	r.Knowledge = totalKnowledge([]Knowledge{r.DirectionKnowledge, r.GoalKnowledge, r.DirectKnowledge})
	rs := r.Reasons
	for _, g := range r.Goals {
		rs = append(rs, g.Reasons...)
	}
	r.Reasons = canonicalReasons(rs)
	slices.Sort(r.Directions)
	r.Directions = slices.Compact(r.Directions)
	slices.SortFunc(r.Goals, func(a, b Goal) int { return compareTarget(a.Target, b.Target) })
	return report(r)
}

// Tendency answers one source place while preserving three independent coverage states.
func (s *Scheme) Tendency(m mode.Mode, p Place) (TendencyResult, error) {
	if e := s.check(); e != nil {
		return TendencyResult{}, e
	}
	if e := ValidatePlace(m, p); e != nil {
		return TendencyResult{}, e
	}
	a := s.applicability(m, TendencyTask)
	r := TendencyResult{Answer: a, Source: p, DirectionKnowledge: a.Knowledge, GoalKnowledge: a.Knowledge, DirectKnowledge: a.Knowledge}
	if a.Knowledge != Known {
		return report(r), nil
	}
	if !s.state.info.Builtin {
		r.Answer = s.missing()
		r.DirectionKnowledge = MissingRule
		r.GoalKnowledge = MissingRule
		r.DirectKnowledge = MissingRule
		for _, row := range s.state.def.Tendencies {
			if row.Degree != p.Degree {
				continue
			}
			r.Reasons = bases(row.Bases)
			if row.Directions != nil || row.DirectComplete {
				r.DirectionKnowledge = Known
				r.Directions = slices.Clone(row.Directions)
			}
			if row.Goals != nil {
				r.GoalKnowledge = Known
			}
			if row.DirectComplete {
				r.DirectKnowledge = Known
			}
			for _, g := range row.Goals {
				goal := Goal{Target: g.Target, Direction: g.Direction, Preferred: g.Preferred, Reasons: bases(g.Bases)}
				if g.Target.CycleKnown {
					c, e := intSum(p.Cycle, g.Target.Cycle)
					if e != nil {
						return TendencyResult{}, e
					}
					goal.Target.Cycle = c
					v, e := Between(m, p, Place{goal.Target.Degree, c})
					if e != nil {
						return TendencyResult{}, e
					}
					goal.Interval = v.Interval
					goal.Direction = v.Direction
				}
				r.Goals = append(r.Goals, goal)
			}
			break
		}
		return finishTendency(s, r), nil
	}
	switch s.state.info.ID {
	case AMNeighborTargets:
		r.Reasons = []Reason{sourceReason("am.neighbor-targets", "tendency inv-1/var-11", "Only immediately neighboring AM supports are direct goals.", Fact{"am-structure", amStructures[amKind(m)]})}
		if supportDegree(p.Degree) {
			break
		}
		for _, d := range []Motion{Descending, Ascending} {
			q, ok, e := neighbor(m, p, d)
			if e != nil {
				return TendencyResult{}, e
			}
			if ok && supportDegree(q.Degree) {
				v, e := Between(m, p, q)
				if e != nil {
					return TendencyResult{}, e
				}
				r.Goals = append(r.Goals, Goal{Target: Target{Immediate, q.Degree, q.Cycle, true}, Direction: d, Interval: v.Interval, Reasons: copied(r.Reasons)})
				r.Directions = append(r.Directions, d)
			}
		}
		if len(r.Goals) == 1 {
			r.Goals[0].Preferred = true
		} else if len(r.Goals) == 2 {
			a, b := r.Goals[0], r.Goals[1]
			c := a.Interval.Deviation().Compare(b.Interval.Deviation())
			idx := -1
			if c < 0 {
				idx = 0
			} else if c > 0 {
				idx = 1
			} else if a.Target.Degree == 1 {
				idx = 0
			} else if b.Target.Degree == 1 {
				idx = 1
			}
			if idx >= 0 {
				r.Goals[idx].Preferred = true
			}
		}
		for i := range r.Goals {
			r.Goals[i].Reasons = append(r.Goals[i].Reasons, productReason("product.neighbor-preference", "Prefer the narrower neighboring second, then degree I on a tie."))
		}
	case AMTextTargets:
		r.DirectionKnowledge = MissingRule
		r.DirectKnowledge = MissingRule
		kind := amKind(m)
		targets := []int{}
		preferred := 0
		switch {
		case kind == 0 || kind == 3:
			switch p.Degree {
			case 2:
				targets = []int{1, 3}
				preferred = 1
			case 4:
				targets = []int{3, 5}
				if kind == 0 {
					preferred = 3
				}
			case 6:
				targets = []int{5}
				preferred = 5
			case 7:
				targets = []int{1}
				preferred = 1
			}
		case kind == 1 && p.Degree == 6:
			targets = []int{5}
			preferred = 5
		case kind == 2 && p.Degree == 7:
			targets = []int{1}
			preferred = 1
		case kind == 4 && p.Degree == 7:
			targets = []int{1}
			preferred = 1
		case kind == 5 && p.Degree == 6:
			targets = []int{5, 1}
			preferred = 1
		}
		if len(targets) == 0 {
			r.GoalKnowledge = MissingRule
			r.Reasons = s.missing().Reasons
		} else {
			r.Reasons = []Reason{sourceReason("am.text-targets", "tendency-am-targets.csv: "+amStructures[kind]+" degree "+strconv.Itoa(p.Degree), "The named target row does not specify direct transitions.")}
			for _, n := range targets {
				g := Goal{Target: Target{Kind: Named, Degree: n}, Preferred: n == preferred, Reasons: copied(r.Reasons)}
				if kind == 5 && p.Degree == 6 && n == 1 {
					g.Target.Kind = Ultimate
					g.Reasons = append(g.Reasons, sourceReason("am.text-targets", "tendency var-10 / AM L1622", "The final tonic is reached through raised degree VII.", Fact{"via-degree", "7"}))
				}
				r.Goals = append(r.Goals, g)
			}
		}
	case SeredaTonal, SeredaModal:
		pos, e := modal.Relative(m, p.Degree)
		if e != nil {
			r.DirectionKnowledge = OutsideScope
			r.GoalKnowledge = OutsideScope
			r.DirectKnowledge = OutsideScope
			r.Reasons = []Reason{sourceReason("modal.unsupported-quality", "modal MOD-1", "The relative interval quality is not supported by modal.Relative.", Fact{"interval", m.MustDegreeInterval(p.Degree).String()})}
			break
		}
		if pos.Category == modal.Neutral {
			r.DirectionKnowledge = MissingContext
			r.GoalKnowledge = MissingContext
			r.DirectKnowledge = MissingContext
			r.Reasons = []Reason{sourceReason("context.neutral-pair", "tendency gap-4", "A neutral relative position requires a companion-voice pair rule.", Fact{"category", pos.Category.String()})}
			break
		}
		d := Ascending
		if pos.Category == modal.Low {
			d = Descending
		}
		rule := "sereda.tonal-direction"
		if s.state.info.ID == SeredaModal {
			rule = "sereda.modal-direction"
			if d == Ascending {
				d = Descending
			} else {
				d = Ascending
			}
		}
		r.Directions = []Motion{d}
		r.Reasons = []Reason{sourceReason(rule, "tendency var-5/var-6", "The selected Sereda rule maps the relative category to this direction.", Fact{"category", pos.Category.String()}), productReason("product.direction-to-neighbor", "The direction is realized by the adjacent place in that direction.")}
		q, ok, e := neighbor(m, p, d)
		if e != nil {
			return TendencyResult{}, e
		}
		if ok {
			v, e := Between(m, p, q)
			if e != nil {
				return TendencyResult{}, e
			}
			r.Goals = []Goal{{Target: Target{Immediate, q.Degree, q.Cycle, true}, Direction: d, Interval: v.Interval, Reasons: copied(r.Reasons)}}
		} else {
			r.Reasons = append(r.Reasons, productReason("boundary.no-neighbor", "The open structure has no neighbor in the required direction."))
		}
	}
	return finishTendency(s, r), nil
}

// Tendencies answers every degree in the requested signed cycle.
func (s *Scheme) Tendencies(m mode.Mode, c int) (TendencyAnalysis, error) {
	if e := s.check(); e != nil {
		return TendencyAnalysis{}, e
	}
	if e := modeOK(m); e != nil {
		return TendencyAnalysis{}, e
	}
	r := TendencyAnalysis{}
	ks := []Knowledge{}
	rs := []Reason{}
	for n := 1; n <= m.Len(); n++ {
		v, e := s.Tendency(m, Place{n, c})
		if e != nil {
			return TendencyAnalysis{}, e
		}
		r.Rows = append(r.Rows, v)
		ks = append(ks, v.Knowledge)
		rs = append(rs, v.Reasons...)
	}
	r.Answer = s.answer(totalKnowledge(ks), rs...)
	return report(r), nil
}

// CheckTransition checks a direct move; named or ultimate goals alone do not prove permission.
func (s *Scheme) CheckTransition(m mode.Mode, from, to Place) (CheckResult, error) {
	if e := s.check(); e != nil {
		return CheckResult{}, e
	}
	if e := ValidatePlace(m, from); e != nil {
		return CheckResult{}, e
	}
	if e := ValidatePlace(m, to); e != nil {
		return CheckResult{}, e
	}
	r, e := s.Tendency(m, from)
	if e != nil {
		return CheckResult{}, e
	}
	a := r.Answer
	d := Undetermined
	for _, g := range r.Goals {
		if g.Target.Kind == Immediate && g.Target.CycleKnown && g.Target.Degree == to.Degree && g.Target.Cycle == to.Cycle && r.DirectKnowledge == Known {
			d = Allowed
		}
	}
	if d != Allowed && r.DirectKnowledge == Known {
		d = Denied
	}
	if r.DirectionKnowledge == Known {
		// A known empty direction set excludes every movement.
		admitted := false
		if len(r.Directions) > 0 {
			v, e := Between(m, from, to)
			if e != nil {
				return CheckResult{}, e
			}
			admitted = slices.Contains(r.Directions, v.Direction)
		}
		if !admitted {
			d = Denied
			a.Reasons = append(a.Reasons, productReason("filter.direction", "The proposed transition has no direction admitted by the rule."))
		}
	}
	return report(CheckResult{a, d}), nil
}
