package tendency

import (
	"cmp"
	"errors"
	"reflect"
	"slices"
	"strings"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/mode"
)

const (
	AMSupports           ID = "am-supports"
	AMMajorRoles         ID = "am-major-roles"
	AMMinorRoles         ID = "am-minor-roles"
	AMNeighborTargets    ID = "am-neighbor-targets"
	AMTextTargets        ID = "am-text-targets"
	SeredaTonal          ID = "sereda-tonal"
	SeredaModal          ID = "sereda-modal"
	AMConsonance         ID = "am-consonance"
	OMTHarmonic          ID = "omt-harmonic"
	OMTMelodic           ID = "omt-melodic"
	OMTCounterpoint      ID = "omt-counterpoint"
	MT21Consonance       ID = "mt21-consonance"
	SeredaConsonance     ID = "sereda-consonance"
	WPConsonance         ID = "wp-consonance"
	AMModalResolution    ID = "am-modal-resolution"
	AMIntervalResolution ID = "am-interval-resolution"
)

var builtinIDs = []ID{AMSupports, AMMajorRoles, AMMinorRoles, AMNeighborTargets, AMTextTargets, SeredaTonal, SeredaModal, AMConsonance, OMTHarmonic, OMTMelodic, OMTCounterpoint, MT21Consonance, SeredaConsonance, WPConsonance, AMModalResolution, AMIntervalResolution}

type schemeState struct {
	info Info
	def  Definition
}

// Builtin returns a fresh handle for one of the sixteen immutable built-in rules.
func Builtin(id ID) (*Scheme, error) {
	if !slices.Contains(builtinIDs, id) {
		return nil, failure("builtin", ErrUnknownScheme)
	}
	task := ClassificationTask
	scope := "Finite named interval table; uncovered qualities remain unknown."
	name := ""
	refs := []string{"consonance-dissonance §6"}
	context := []string{}
	switch id {
	case AMSupports:
		task = SupportTask
		name = "AM major and minor supports"
	case AMMajorRoles:
		task = RoleTask
		name = "AM natural major degree labels"
	case AMMinorRoles:
		task = RoleTask
		name = "AM natural minor degree labels"
	case AMNeighborTargets:
		task = TendencyTask
		name = "AM neighboring supports"
	case AMTextTargets:
		task = TendencyTask
		name = "AM text targets (12 rows)"
		context = []string{"Named targets do not specify direct transitions."}
	case SeredaTonal:
		task = TendencyTask
		name = "Sereda tonal direction"
	case SeredaModal:
		task = TendencyTask
		name = "Sereda modal direction"
	case AMModalResolution:
		task = ModalResolutionTask
		name = "AM modal two-voice resolution"
	case AMIntervalResolution:
		task = IntervalResolutionTask
		name = "AM basic dissonance resolution"
	case AMConsonance:
		name = "AM consonance"
	case OMTHarmonic:
		name = "OMT harmonic consonance"
	case OMTMelodic:
		name = "OMT melodic consonance"
	case OMTCounterpoint:
		name = "OMT counterpoint consonance"
	case MT21Consonance:
		name = "MT21 counterpoint consonance"
	case SeredaConsonance:
		name = "Sereda two-voice consonance"
	case WPConsonance:
		name = "WP consonance"
	}
	if task <= ModalResolutionTask && task != ClassificationTask {
		scope = "Six closed seven-degree AM structures with a perfect-octave cycle."
		refs = []string{"tendency inv-1; stability inv-2; interval-resolution op-3"}
	}
	if id == AMMajorRoles || id == AMMinorRoles {
		scope = "Natural major or natural minor, respectively."
		refs = []string{"tonal-function degree labels"}
	}
	if id == SeredaTonal || id == SeredaModal {
		scope = "Any valid structure whose requested relative quality is supported by modal.Relative."
		refs = []string{"tendency var-5/var-6; modal MOD-1"}
		context = []string{"Neutral relative positions require a companion-voice pair rule; no automatic pair is supplied."}
	}
	if id == AMIntervalResolution {
		scope = "Ascending simple basic seconds, sevenths and tritones; AM consonances require no resolution."
		refs = []string{"interval-resolution inv-3/inv-4/inv-6"}
	}
	return &Scheme{state: &schemeState{info: report(Info{ID: id, Name: name, Description: name + " under explicitly selected rules. " + scope, Tasks: []Task{task}, Scope: scope, RequiredContext: context, References: refs, Builtin: true})}}, nil
}

// Catalog returns matching built-in metadata in catalog order.
func Catalog(q CatalogQuery) ([]Info, error) {
	if q.Task != TaskUnset && !validTask(q.Task) {
		return nil, failure("catalog task", ErrInvalidArgument)
	}
	r := []Info{}
	for _, id := range builtinIDs {
		s := MustBuiltin(id)
		if q.Task != TaskUnset && !slices.Contains(s.state.info.Tasks, q.Task) {
			continue
		}
		if q.Mode.IsValid() {
			ok := false
			for _, t := range s.state.info.Tasks {
				if t != ClassificationTask && t != IntervalResolutionTask && s.scope(q.Mode, t) {
					ok = true
				}
			}
			if !ok {
				continue
			}
		}
		r = append(r, s.Info())
	}
	return r, nil
}

var amStructures = []string{"P1 M2 M3 P4 P5 M6 M7", "P1 M2 M3 P4 P5 m6 M7", "P1 M2 M3 P4 P5 m6 m7", "P1 M2 m3 P4 P5 m6 m7", "P1 M2 m3 P4 P5 m6 M7", "P1 M2 m3 P4 P5 M6 M7"}

func amKind(m mode.Mode) int {
	if !m.IsClosed() || m.Len() != 7 || !m.IsOctave() {
		return -1
	}
	for j, text := range amStructures {
		ok := true
		for i, v := range strings.Fields(text) {
			if !m.MustDegreeInterval(i + 1).Equal(interval.MustParse(v)) {
				ok = false
				break
			}
		}
		if ok {
			return j
		}
	}
	return -1
}
func (s *Scheme) scope(m mode.Mode, t Task) bool {
	if !s.state.info.Builtin {
		return s.state.def.Mode.Equal(m)
	}
	switch s.state.info.ID {
	case SeredaTonal, SeredaModal:
		return true
	case AMMajorRoles:
		return amKind(m) == 0
	case AMMinorRoles:
		return amKind(m) == 3
	default:
		return amKind(m) >= 0
	}
}
func (s *Scheme) applicability(m mode.Mode, t Task) Answer {
	if !slices.Contains(s.state.info.Tasks, t) || t == ClassificationTask || t == IntervalResolutionTask {
		return s.answer(UnsupportedTask, productReason("coverage.task", "The scheme does not provide this modal task."))
	}
	if !s.scope(m, t) {
		return s.answer(OutsideScope, productReason("scope.mode", "The current mode structure is outside this scheme's scope."))
	}
	return s.answer(Known, productReason("scope.mode", "The current mode structure is compatible with this scheme."))
}

// Applicability checks compatibility with a modal task, without promising coverage of each row.
func (s *Scheme) Applicability(m mode.Mode, t Task) (Answer, error) {
	if e := s.check(); e != nil {
		return Answer{}, e
	}
	if e := modeOK(m); e != nil {
		return Answer{}, e
	}
	if !validTask(t) {
		return Answer{}, failure("applicability task", ErrInvalidArgument)
	}
	return s.applicability(m, t), nil
}
func definitionError(e error) error {
	return failure("definition", errors.Join(ErrInvalidDefinition, e))
}
func canonicalBases(v []Basis) ([]Basis, error) {
	if len(v) == 0 {
		return nil, ErrInvalidDefinition
	}
	v = copied(v)
	for i := range v {
		if strings.TrimSpace(v[i].Rule) == "" || strings.TrimSpace(v[i].Explanation) == "" {
			return nil, ErrInvalidDefinition
		}
		slices.Sort(v[i].References)
		v[i].References = slices.Compact(v[i].References)
		if v[i].References == nil {
			v[i].References = []string{}
		}
	}
	slices.SortFunc(v, func(a, b Basis) int {
		if c := cmp.Compare(a.Rule, b.Rule); c != 0 {
			return c
		}
		if c := cmp.Compare(a.Explanation, b.Explanation); c != 0 {
			return c
		}
		return slices.Compare(a.References, b.References)
	})
	return slices.CompactFunc(v, func(a, b Basis) bool { return reflect.DeepEqual(a, b) }), nil
}
func mergedBases(a, b []Basis) []Basis {
	r, e := canonicalBases(append(slices.Clone(a), b...))
	if e != nil {
		panic(e)
	}
	return r
}
func canonicalDirections(v []Motion) ([]Motion, error) {
	for _, d := range v {
		if !validMotion(d) {
			return nil, ErrInvalidArgument
		}
	}
	v = slices.Clone(v)
	slices.Sort(v)
	return slices.Compact(v), nil
}

// New validates and canonicalizes a finite user definition, taking an independent snapshot.
func New(input Definition) (*Scheme, error) {
	d := copied(input)
	if !strings.HasPrefix(string(d.ID), "user/") || strings.TrimSpace(strings.TrimPrefix(string(d.ID), "user/")) == "" || strings.TrimSpace(d.Name) == "" || strings.TrimSpace(d.Description) == "" {
		return nil, definitionError(ErrInvalidArgument)
	}
	modalPart := d.Supports != nil || d.Roles != nil || d.Tendencies != nil || d.Modal != nil
	if !modalPart && d.Intervals == nil {
		return nil, definitionError(ErrInvalidArgument)
	}
	if modalPart {
		if e := modeOK(d.Mode); e != nil {
			return nil, definitionError(e)
		}
	} else if d.Mode.IsValid() {
		return nil, definitionError(ErrInvalidArgument)
	}
	var e error
	for i := range d.Supports {
		r := &d.Supports[i]
		if e = numberOK(d.Mode, r.Degree); e != nil {
			return nil, definitionError(e)
		}
		v := r.Value
		if v.Kind < Unstable || v.Kind > MainSupport || (!v.Ranked && v.Rank != 0) || (v.Ranked && (v.Rank <= 0 || v.Kind == Unstable)) {
			return nil, definitionError(ErrInvalidArgument)
		}
		r.Bases, e = canonicalBases(r.Bases)
		if e != nil {
			return nil, definitionError(e)
		}
	}
	slices.SortFunc(d.Supports, func(a, b SupportRule) int { return cmp.Compare(a.Degree, b.Degree) })
	for i := 1; i < len(d.Supports); {
		a, b := &d.Supports[i-1], d.Supports[i]
		if a.Degree != b.Degree {
			i++
			continue
		}
		if a.Value != b.Value {
			return nil, definitionError(ErrInvalidArgument)
		}
		a.Bases = mergedBases(a.Bases, b.Bases)
		d.Supports = slices.Delete(d.Supports, i, i+1)
	}
	for i := range d.Roles {
		r := &d.Roles[i]
		if e = numberOK(d.Mode, r.Degree); e != nil {
			return nil, definitionError(e)
		}
		r.Bases, e = canonicalBases(r.Bases)
		if e != nil {
			return nil, definitionError(e)
		}
	}
	slices.SortFunc(d.Roles, func(a, b RoleRule) int { return cmp.Compare(a.Degree, b.Degree) })
	for i := 1; i < len(d.Roles); {
		a, b := &d.Roles[i-1], d.Roles[i]
		if a.Degree != b.Degree {
			i++
			continue
		}
		if a.Role != b.Role {
			return nil, definitionError(ErrInvalidArgument)
		}
		a.Bases = mergedBases(a.Bases, b.Bases)
		d.Roles = slices.Delete(d.Roles, i, i+1)
	}
	for i := range d.Tendencies {
		r := &d.Tendencies[i]
		if e = canonicalTendency(d.Mode, r); e != nil {
			return nil, definitionError(e)
		}
	}
	slices.SortFunc(d.Tendencies, func(a, b TendencyRule) int { return cmp.Compare(a.Degree, b.Degree) })
	for i := 1; i < len(d.Tendencies); {
		a, b := &d.Tendencies[i-1], d.Tendencies[i]
		if a.Degree != b.Degree {
			i++
			continue
		}
		if !sameTendency(*a, b) {
			return nil, definitionError(ErrInvalidArgument)
		}
		a.Bases = mergedBases(a.Bases, b.Bases)
		for j := range a.Goals {
			a.Goals[j].Bases = mergedBases(a.Goals[j].Bases, b.Goals[j].Bases)
		}
		d.Tendencies = slices.Delete(d.Tendencies, i, i+1)
	}
	for i := range d.Modal {
		if e = canonicalModal(d.Mode, &d.Modal[i]); e != nil {
			return nil, definitionError(e)
		}
	}
	slices.SortFunc(d.Modal, func(a, b ModalRule) int {
		if c := comparePair(a.Source, b.Source); c != 0 {
			return c
		}
		return cmp.Compare(boolInt(a.Repeat), boolInt(b.Repeat))
	})
	for i := 0; i < len(d.Modal); i++ {
		for j := i + 1; j < len(d.Modal); {
			a, b := &d.Modal[i], d.Modal[j]
			overlap := a.Source == b.Source
			if a.Repeat != b.Repeat {
				rep, exact := *a, b
				if !rep.Repeat {
					rep, exact = b, *a
				}
				x, err := normalizedPair(exact.Source)
				if err != nil {
					return nil, definitionError(err)
				}
				overlap = x == rep.Source
			}
			if !overlap {
				j++
				continue
			}
			if a.Repeat != b.Repeat || !sameModal(*a, b) {
				return nil, definitionError(ErrInvalidArgument)
			}
			a.Bases = mergedBases(a.Bases, b.Bases)
			for k := range a.Options {
				a.Options[k].Bases = mergedBases(a.Options[k].Bases, b.Options[k].Bases)
			}
			d.Modal = slices.Delete(d.Modal, j, j+1)
		}
	}
	for i := range d.Intervals {
		if e = canonicalInterval(&d.Intervals[i]); e != nil {
			return nil, definitionError(e)
		}
	}
	slices.SortFunc(d.Intervals, func(a, b IntervalRule) int { return compareInterval(a.Source, b.Source) })
	for i := 1; i < len(d.Intervals); {
		a, b := &d.Intervals[i-1], d.Intervals[i]
		if a.Source != b.Source {
			i++
			continue
		}
		if !sameInterval(*a, b) {
			return nil, definitionError(ErrInvalidArgument)
		}
		a.Bases = mergedBases(a.Bases, b.Bases)
		for j := range a.Options {
			a.Options[j].Bases = mergedBases(a.Options[j].Bases, b.Options[j].Bases)
		}
		d.Intervals = slices.Delete(d.Intervals, i, i+1)
	}
	tasks := []Task{}
	if d.Supports != nil {
		tasks = append(tasks, SupportTask)
	}
	if d.Roles != nil {
		tasks = append(tasks, RoleTask)
	}
	if d.Tendencies != nil {
		tasks = append(tasks, TendencyTask)
	}
	if d.Modal != nil {
		tasks = append(tasks, ModalResolutionTask)
	}
	if d.Intervals != nil {
		tasks = append(tasks, IntervalResolutionTask)
	}
	return &Scheme{state: &schemeState{info: report(Info{ID: d.ID, Name: d.Name, Description: d.Description, Tasks: tasks, Scope: "Explicit finite user rules; modal rows require the exact current structure."}), def: d}}, nil
}
func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
func sameTendency(a, b TendencyRule) bool {
	a.Bases = nil
	b.Bases = nil
	a.Goals = copied(a.Goals)
	b.Goals = copied(b.Goals)
	for i := range a.Goals {
		a.Goals[i].Bases = nil
	}
	for i := range b.Goals {
		b.Goals[i].Bases = nil
	}
	return reflect.DeepEqual(a, b)
}
func canonicalTendency(m mode.Mode, r *TendencyRule) error {
	if e := numberOK(m, r.Degree); e != nil {
		return e
	}
	var e error
	r.Bases, e = canonicalBases(r.Bases)
	if e != nil {
		return e
	}
	r.Directions, e = canonicalDirections(r.Directions)
	if e != nil {
		return e
	}
	if r.DirectComplete && r.Goals == nil {
		return ErrInvalidArgument
	}
	for i := range r.Goals {
		g := &r.Goals[i]
		if g.Target.Kind < Immediate || g.Target.Kind > Named {
			return ErrInvalidArgument
		}
		if e = numberOK(m, g.Target.Degree); e != nil {
			return e
		}
		if !g.Target.CycleKnown {
			if g.Target.Kind == Immediate || g.Target.Cycle != 0 || g.Direction != MotionUnset {
				return ErrInvalidArgument
			}
		} else {
			mv, err := Between(m, Place{Degree: r.Degree}, Place{Degree: g.Target.Degree, Cycle: g.Target.Cycle})
			if err != nil {
				return err
			}
			if g.Direction != MotionUnset && g.Direction != mv.Direction {
				return ErrInvalidArgument
			}
			g.Direction = mv.Direction
		}
		g.Bases, e = canonicalBases(g.Bases)
		if e != nil {
			return e
		}
	}
	slices.SortFunc(r.Goals, func(a, b GoalRule) int { return compareTarget(a.Target, b.Target) })
	for i := 1; i < len(r.Goals); {
		a, b := &r.Goals[i-1], r.Goals[i]
		if a.Target != b.Target {
			i++
			continue
		}
		if a.Direction != b.Direction || a.Preferred != b.Preferred {
			return ErrInvalidArgument
		}
		a.Bases = mergedBases(a.Bases, b.Bases)
		r.Goals = slices.Delete(r.Goals, i, i+1)
	}
	dirs := []Motion{}
	preferred := 0
	for _, g := range r.Goals {
		if g.Preferred {
			preferred++
		}
		if g.Target.Kind == Immediate {
			dirs = append(dirs, g.Direction)
			if r.Directions != nil && !slices.Contains(r.Directions, g.Direction) {
				return ErrInvalidArgument
			}
		}
	}
	if preferred > 1 {
		return ErrInvalidArgument
	}
	slices.Sort(dirs)
	dirs = slices.Compact(dirs)
	if r.DirectComplete {
		if r.Directions == nil {
			r.Directions = dirs
		} else if !slices.Equal(r.Directions, dirs) && (len(dirs) != 0 || len(r.Directions) != 1) {
			return ErrInvalidArgument
		}
	}
	return nil
}
func normalizedPair(p Pair) (Pair, error) {
	c, e := intDifference(p.Second.Cycle, p.First.Cycle)
	if e != nil {
		return Pair{}, e
	}
	return Pair{Place{Degree: p.First.Degree}, Place{Degree: p.Second.Degree, Cycle: c}}, nil
}
func shiftedPlace(p Place, c int) (Place, error) {
	n, e := intSum(p.Cycle, c)
	if e != nil {
		return Place{}, e
	}
	p.Cycle = n
	return p, nil
}
func shiftedPair(p Pair, c int) (Pair, error) {
	a, e := shiftedPlace(p.First, c)
	if e != nil {
		return Pair{}, e
	}
	b, e := shiftedPlace(p.Second, c)
	return Pair{a, b}, e
}
func canonicalModal(m mode.Mode, r *ModalRule) error {
	if r.Options == nil {
		r.Options = []ModalOption{}
	}
	if e := pairOK(m, r.Source); e != nil {
		return e
	}
	var e error
	r.Bases, e = canonicalBases(r.Bases)
	if e != nil {
		return e
	}
	if r.Repeat && m.IsOpen() {
		return ErrInvalidArgument
	}
	offset := r.Source.First.Cycle
	if r.Repeat {
		r.Source, e = normalizedPair(r.Source)
		if e != nil {
			return e
		}
	}
	for i := range r.Options {
		o := &r.Options[i]
		if e = pairOK(m, o.Target); e != nil {
			return e
		}
		if comparePlace(o.Target.First, o.Target.Second) > 0 && !o.AllowCrossing {
			return ErrInvalidArgument
		}
		if r.Repeat {
			a, err := intDifference(o.Target.First.Cycle, offset)
			if err != nil {
				return err
			}
			b, err := intDifference(o.Target.Second.Cycle, offset)
			if err != nil {
				return err
			}
			o.Target.First.Cycle = a
			o.Target.Second.Cycle = b
		}
		if _, e = makeModalCandidate(m, r.Source, o.Target, false, nil); e != nil {
			return e
		}
		o.Bases, e = canonicalBases(o.Bases)
		if e != nil {
			return e
		}
	}
	slices.SortFunc(r.Options, func(a, b ModalOption) int { return comparePair(a.Target, b.Target) })
	pref := 0
	for i := 0; i < len(r.Options); i++ {
		if i > 0 && r.Options[i-1].Target == r.Options[i].Target {
			a, b := &r.Options[i-1], r.Options[i]
			if a.AllowCrossing != b.AllowCrossing || a.Preferred != b.Preferred {
				return ErrInvalidArgument
			}
			a.Bases = mergedBases(a.Bases, b.Bases)
			r.Options = slices.Delete(r.Options, i, i+1)
			i--
			continue
		}
	}
	for _, o := range r.Options {
		if o.Preferred {
			pref++
		}
	}
	if pref > 1 {
		return ErrInvalidArgument
	}
	return nil
}
func sameModal(a, b ModalRule) bool {
	a.Bases = nil
	b.Bases = nil
	a.Options = copied(a.Options)
	b.Options = copied(b.Options)
	for i := range a.Options {
		a.Options[i].Bases = nil
	}
	for i := range b.Options {
		b.Options[i].Bases = nil
	}
	if len(a.Options) == 0 {
		a.Options = nil
	}
	if len(b.Options) == 0 {
		b.Options = nil
	}
	return reflect.DeepEqual(a, b)
}
func intervalFinal(i interval.Interval, m Motions) (interval.Interval, error) {
	return coord(i).add(coord(m.Second)).sub(coord(m.First)).value()
}
func canonicalInterval(r *IntervalRule) error {
	if r.Options == nil {
		r.Options = []IntervalOption{}
	}
	if e := intervalOK(r.Source); e != nil {
		return e
	}
	var e error
	r.Bases, e = canonicalBases(r.Bases)
	if e != nil {
		return e
	}
	for i := range r.Options {
		o := &r.Options[i]
		if e = motionsOK(o.Motions); e != nil {
			return e
		}
		v, err := intervalFinal(r.Source, o.Motions)
		if err != nil {
			return err
		}
		if v.IsDescending() && !o.AllowCrossing {
			return ErrInvalidArgument
		}
		o.Bases, e = canonicalBases(o.Bases)
		if e != nil {
			return e
		}
	}
	slices.SortFunc(r.Options, func(a, b IntervalOption) int { return compareMotions(a.Motions, b.Motions) })
	for i := 1; i < len(r.Options); {
		a, b := &r.Options[i-1], r.Options[i]
		if a.Motions != b.Motions {
			i++
			continue
		}
		if a.AllowCrossing != b.AllowCrossing || a.Preferred != b.Preferred {
			return ErrInvalidArgument
		}
		a.Bases = mergedBases(a.Bases, b.Bases)
		r.Options = slices.Delete(r.Options, i, i+1)
	}
	n := 0
	for _, o := range r.Options {
		if o.Preferred {
			n++
		}
	}
	if n > 1 {
		return ErrInvalidArgument
	}
	return nil
}
func sameInterval(a, b IntervalRule) bool {
	a.Bases = nil
	b.Bases = nil
	a.Options = copied(a.Options)
	b.Options = copied(b.Options)
	for i := range a.Options {
		a.Options[i].Bases = nil
	}
	for i := range b.Options {
		b.Options[i].Bases = nil
	}
	if len(a.Options) == 0 {
		a.Options = nil
	}
	if len(b.Options) == 0 {
		b.Options = nil
	}
	return reflect.DeepEqual(a, b)
}
func (s *Scheme) missing() Answer {
	return s.answer(MissingRule, productReason("coverage.row", "No rule is defined for this input row."))
}
func (s *Scheme) unsupported() Answer {
	return s.answer(UnsupportedTask, productReason("coverage.task", "This scheme does not provide the requested task."))
}
func (s *Scheme) operation(m mode.Mode, t Task, n int) (Answer, error) {
	if e := s.check(); e != nil {
		return Answer{}, e
	}
	if e := numberOK(m, n); e != nil {
		return Answer{}, e
	}
	return s.applicability(m, t), nil
}
