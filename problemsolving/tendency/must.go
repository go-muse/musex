package tendency

import (
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/mode"
)

func must[T any](v T, e error) T {
	if e != nil {
		panic(e)
	}
	return v
}

// MustBuiltin is the panic-on-error form of Builtin. It panics with the original error.
func MustBuiltin(id ID) *Scheme { return must(Builtin(id)) }

// MustCatalog is the panic-on-error form of Catalog. It panics with the original error.
func MustCatalog(query CatalogQuery) []Info { return must(Catalog(query)) }

// MustNew is the panic-on-error form of New. It panics with the original error.
func MustNew(def Definition) *Scheme { return must(New(def)) }

// MustApplicability is the panic-on-error form of Applicability. It panics with the original error.
func (s *Scheme) MustApplicability(m mode.Mode, task Task) Answer {
	return must(s.Applicability(m, task))
}

// MustBetween is the panic-on-error form of Between. It panics with the original error.
func MustBetween(m mode.Mode, from, to Place) Movement { return must(Between(m, from, to)) }

// MustSupport is the panic-on-error form of Support. It panics with the original error.
func (s *Scheme) MustSupport(m mode.Mode, number int) SupportResult {
	return must(s.Support(m, number))
}

// MustSupports is the panic-on-error form of Supports. It panics with the original error.
func (s *Scheme) MustSupports(m mode.Mode) SupportAnalysis { return must(s.Supports(m)) }

// MustCompareSupport is the panic-on-error form of CompareSupport. It panics with the original error.
func (s *Scheme) MustCompareSupport(m mode.Mode, first, second int) RankResult {
	return must(s.CompareSupport(m, first, second))
}

// MustRole is the panic-on-error form of Role. It panics with the original error.
func (s *Scheme) MustRole(m mode.Mode, number int) RoleResult { return must(s.Role(m, number)) }

// MustRoles is the panic-on-error form of Roles. It panics with the original error.
func (s *Scheme) MustRoles(m mode.Mode) RoleAnalysis { return must(s.Roles(m)) }

// MustTendency is the panic-on-error form of Tendency. It panics with the original error.
func (s *Scheme) MustTendency(m mode.Mode, source Place) TendencyResult {
	return must(s.Tendency(m, source))
}

// MustTendencies is the panic-on-error form of Tendencies. It panics with the original error.
func (s *Scheme) MustTendencies(m mode.Mode, cycle int) TendencyAnalysis {
	return must(s.Tendencies(m, cycle))
}

// MustClassify is the panic-on-error form of Classify. It panics with the original error.
func (s *Scheme) MustClassify(value interval.Interval) ClassificationResult {
	return must(s.Classify(value))
}

// MustResolveModal is the panic-on-error form of ResolveModal. It panics with the original error.
func (s *Scheme) MustResolveModal(m mode.Mode, source Pair) ModalResolutions {
	return must(s.ResolveModal(m, source))
}

// MustResolveInterval is the panic-on-error form of ResolveInterval. It panics with the original error.
func (s *Scheme) MustResolveInterval(source interval.Interval) IntervalResolutions {
	return must(s.ResolveInterval(source))
}

// MustCheckTransition is the panic-on-error form of CheckTransition. It panics with the original error.
func (s *Scheme) MustCheckTransition(m mode.Mode, from, to Place) CheckResult {
	return must(s.CheckTransition(m, from, to))
}

// MustCheckModal is the panic-on-error form of CheckModal. It panics with the original error.
func (s *Scheme) MustCheckModal(m mode.Mode, from, to Pair) CheckResult {
	return must(s.CheckModal(m, from, to))
}

// MustCheckInterval is the panic-on-error form of CheckInterval. It panics with the original error.
func (s *Scheme) MustCheckInterval(from interval.Interval, moves Motions) CheckResult {
	return must(s.CheckInterval(from, moves))
}

// MustSelectGoals is the panic-on-error form of SelectGoals. It panics with the original error.
func (s *Scheme) MustSelectGoals(m mode.Mode, source Place, filter GoalFilter) Selection[TendencyResult, Goal] {
	return must(s.SelectGoals(m, source, filter))
}

// MustSelectModal is the panic-on-error form of SelectModal. It panics with the original error.
func (s *Scheme) MustSelectModal(m mode.Mode, source Pair, filter ModalFilter) Selection[ModalResolutions, ModalCandidate] {
	return must(s.SelectModal(m, source, filter))
}

// MustSelectInterval is the panic-on-error form of SelectInterval. It panics with the original error.
func (s *Scheme) MustSelectInterval(source interval.Interval, filter IntervalFilter) Selection[IntervalResolutions, IntervalCandidate] {
	return must(s.SelectInterval(source, filter))
}

// MustCompareSupports is the panic-on-error form of CompareSupports. It panics with the original error.
func MustCompareSupports(m mode.Mode, number int, schemes []*Scheme) Comparison[SupportResult, SupportValue] {
	return must(CompareSupports(m, number, schemes))
}

// MustCompareRoles is the panic-on-error form of CompareRoles. It panics with the original error.
func MustCompareRoles(m mode.Mode, number int, schemes []*Scheme) Comparison[RoleResult, string] {
	return must(CompareRoles(m, number, schemes))
}

// MustCompareTendencies is the panic-on-error form of CompareTendencies. It panics with the original error.
func MustCompareTendencies(m mode.Mode, source Place, schemes []*Scheme, match GoalMatch) Comparison[TendencyResult, GoalKey] {
	return must(CompareTendencies(m, source, schemes, match))
}

// MustCompareClassifications is the panic-on-error form of CompareClassifications. It panics with the original error.
func MustCompareClassifications(value interval.Interval, schemes []*Scheme) Comparison[ClassificationResult, Classification] {
	return must(CompareClassifications(value, schemes))
}

// MustCompareModalResolutions is the panic-on-error form of CompareModalResolutions. It panics with the original error.
func MustCompareModalResolutions(m mode.Mode, source Pair, schemes []*Scheme) Comparison[ModalResolutions, Pair] {
	return must(CompareModalResolutions(m, source, schemes))
}

// MustCompareIntervalResolutions is the panic-on-error form of CompareIntervalResolutions. It panics with the original error.
func MustCompareIntervalResolutions(source interval.Interval, schemes []*Scheme) Comparison[IntervalResolutions, Motions] {
	return must(CompareIntervalResolutions(source, schemes))
}

// MustValidatePlace is the panic-on-error form of ValidatePlace. It panics with the original error.
func MustValidatePlace(m mode.Mode, p Place) {
	if e := ValidatePlace(m, p); e != nil {
		panic(e)
	}
}
