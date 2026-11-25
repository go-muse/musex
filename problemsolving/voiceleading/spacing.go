package voiceleading

import (
	"slices"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/system"
)

// VoiceInterval is the interval from voice Lower to voice Upper: the written
// interval with its direction and the signed chromatic distance.
type VoiceInterval struct {
	Lower, Upper int
	Interval     interval.Interval
	Chromatic    system.Step
}

// Spacing describes the intervals between adjacent voices, the span from the
// bass to the top voice, and whether the span does not exceed a perfect
// octave by writing. With fewer than two voices Adjacent is empty, Span is the
// zero value, and WithinOctave is false. Adjacent belongs to the caller.
type Spacing struct {
	Adjacent     []VoiceInterval
	Span         VoiceInterval
	WithinOctave bool
}

// exceeds reports whether i exceeds the limit: a larger generic size, or the
// same size with a larger deviation in the order of interval.Compare.
// Directions are ignored; an unset limit is no limit.
func exceeds(i, limit interval.Interval) bool {
	if !limit.IsValid() || !i.IsValid() {
		return false
	}
	return i.MustUp().Compare(limit.MustUp()) > 0
}

// Spacing returns the description of the voices. Errors, in this order:
// ErrUnset; ErrNoOrder when the order is not LowToHigh.
func (s *Sonority) Spacing() (Spacing, error) {
	if err := s.checkOrder(); err != nil {
		return Spacing{}, err
	}
	out := Spacing{Adjacent: slices.Clone(s.adjacent)}
	if len(s.notes) >= 2 {
		out.Span = s.span
		out.WithinOctave = !exceeds(s.span.Interval, interval.PerfectOctave())
	}
	return out, nil
}

// MustSpacing is like Spacing but panics with its error.
func (s *Sonority) MustSpacing() Spacing { return must(s.Spacing()) }

// SpacingClassification is an independent result of one spacing definition.
// Source and Scope are filled also when the definition is not applicable.
// When not applicable, Reason is the first of "silent-parts", "voice-count",
// "crossed-voices", Class is SpacingClassUnset and Considered is empty.
// Crossed lists all crossed adjacent pairs and Silent the silent parts; both
// are always computed. All slices belong to the caller.
type SpacingClassification struct {
	Definition SpacingDefinition
	Source     string
	Scope      string
	Applicable bool
	Reason     string
	Class      SpacingClass
	Considered []VoiceInterval
	Crossed    []VoiceInterval
	Silent     []int
}

func spacingText(def SpacingDefinition) (source, scope string, ok bool) {
	switch def {
	case SpacingME:
		return "Muzykalnaya entsiklopediya, Akkord",
			"exactly four voices; tenor-alto and alto-soprano; a fourth or narrower is close", true
	case SpacingAM:
		return "Alekseev and Myasoedov, Elementary Music Theory, 69-71",
			"two or more voices; every adjacent pair including the bass; a fourth or narrower is close", true
	case SpacingOMT:
		return "Open Music Theory, glossary: closed and open spacing",
			"two or more voices; closed when the span does not exceed P8", true
	default:
		return "", "", false
	}
}

// crossed returns the adjacent pairs whose upper voice is strictly lower.
func (s *Sonority) crossed() []VoiceInterval {
	out := []VoiceInterval{}
	for _, v := range s.adjacent {
		if v.Chromatic < 0 {
			out = append(out, v)
		}
	}
	return out
}

// ClassifySpacing classifies the voices by one explicitly selected definition.
// A fourth or narrower is a narrow interval, a fifth or wider a wide one.
// SpacingME considers tenor-alto and alto-soprano of exactly four voices,
// SpacingAM every adjacent pair of two or more voices, SpacingOMT the span of
// two or more voices. Inapplicability is a regular result, not an error.
// Errors, in this order: ErrUnset; ErrNoOrder; ErrOption for def.
func (s *Sonority) ClassifySpacing(def SpacingDefinition) (SpacingClassification, error) {
	if err := s.checkOrder(); err != nil {
		return SpacingClassification{}, err
	}
	source, scope, ok := spacingText(def)
	if !ok {
		return SpacingClassification{}, ErrOption
	}
	out := SpacingClassification{
		Definition: def, Source: source, Scope: scope,
		Considered: []VoiceInterval{}, Crossed: s.crossed(), Silent: slices.Clone(s.silent),
	}
	n := len(s.notes)
	switch {
	case len(s.silent) > 0:
		out.Reason = ReasonSilentParts
	case def == SpacingME && n != 4 || n < 2:
		out.Reason = ReasonVoiceCount
	case len(out.Crossed) > 0:
		out.Reason = ReasonCrossedVoices
	}
	if out.Reason != "" {
		return out, nil
	}
	out.Applicable = true
	switch def {
	case SpacingME:
		out.Considered = []VoiceInterval{s.adjacent[1], s.adjacent[2]}
	case SpacingAM:
		out.Considered = slices.Clone(s.adjacent)
	default:
		out.Considered = []VoiceInterval{s.span}
		out.Class = OMTOpenSpacing
		if !exceeds(s.span.Interval, interval.PerfectOctave()) {
			out.Class = OMTClosedSpacing
		}
		return out, nil
	}
	narrow, wide := 0, 0
	for _, v := range out.Considered {
		if v.Interval.Size() <= 4 {
			narrow++
		} else {
			wide++
		}
	}
	switch {
	case wide == 0:
		out.Class = CloseSpacing
	case narrow == 0:
		out.Class = OpenSpacing
	default:
		out.Class = MixedSpacing
	}
	return out, nil
}

// MustClassifySpacing is like ClassifySpacing but panics with its error.
func (s *Sonority) MustClassifySpacing(def SpacingDefinition) SpacingClassification {
	return must(s.ClassifySpacing(def))
}

// DistanceNorm selects a norm of distances. Limits is used only with
// DistanceCustom: one limit for every adjacent pair or one limit per pair from
// the bass upward. interval.Interval{} is no limit.
type DistanceNorm struct {
	Rule   DistanceRule
	Limits []interval.Interval
}

// DistanceViolation is an adjacent pair that exceeds its limit.
type DistanceViolation struct {
	Pair  VoiceInterval
	Limit interval.Interval
}

// DistanceAssessment is an independent result of a distance norm. Remarks are
// used only by DistanceOMTFirstSpecies: pairs beyond P8, which are not
// violations. When not applicable, Reason is "silent-parts" or "voice-count"
// and both lists are empty. All slices belong to the caller.
type DistanceAssessment struct {
	Rule       string
	Source     string
	Scope      string
	Applicable bool
	Reason     string
	Silent     []int
	Violations []DistanceViolation
	Remarks    []DistanceViolation
}

type distanceText struct {
	voices        int // 0 means two or more
	limits        []interval.Interval
	source, scope string
}

func distanceRule(r DistanceRule) (distanceText, bool) {
	p8, p12 := interval.PerfectOctave(), interval.PerfectTwelfth()
	switch r {
	case DistanceOMTSATB:
		return distanceText{voices: 4, limits: []interval.Interval{p12, p8, p8}, source: "Open Music Theory, SATB style",
			scope: "four voices; soprano-alto and alto-tenor at most P8, tenor-bass at most P12"}, true
	case DistanceMT21SATB:
		return distanceText{voices: 4, limits: []interval.Interval{{}, p8, p8}, source: "Music Theory for the 21st-Century Classroom, 26.6",
			scope: "four voices; soprano-alto and alto-tenor at most P8, tenor-bass unlimited"}, true
	case DistanceOMTFirstSpecies:
		return distanceText{voices: 2, limits: []interval.Interval{p12}, source: "Open Music Theory, first species",
			scope: "two voices; at most P12; intervals beyond P8 are remarked"}, true
	case DistanceCustom:
		return distanceText{source: "composer", scope: "adjacent voices within the given limits"}, true
	default:
		return distanceText{}, false
	}
}

// validateNorm checks the rule and the form of its limits.
func validateNorm(norm DistanceNorm) (distanceText, error) {
	text, ok := distanceRule(norm.Rule)
	if !ok {
		return distanceText{}, ErrOption
	}
	if norm.Rule == DistanceCustom && len(norm.Limits) == 0 || norm.Rule != DistanceCustom && len(norm.Limits) > 0 {
		return distanceText{}, ErrOption
	}
	return text, nil
}

// fitsVoices reports whether a validated norm suits n voices.
func (t distanceText) fitsVoices(n int) bool {
	if t.voices != 0 {
		return n == t.voices
	}
	return n >= 2
}

// pairLimits returns the limit of each of n-1 pairs, or false for a custom
// norm whose number of limits is neither 1 nor n-1.
func (t distanceText) pairLimits(norm DistanceNorm, n int) ([]interval.Interval, bool) {
	if norm.Rule != DistanceCustom {
		return t.limits, true
	}
	switch len(norm.Limits) {
	case 1:
		out := make([]interval.Interval, n-1)
		for k := range out {
			out[k] = norm.Limits[0]
		}
		return out, true
	case n - 1:
		return norm.Limits, true
	default:
		return nil, false
	}
}

// CheckDistances applies a norm of distances to the adjacent voices from the
// bass upward. A pair exceeds a limit when its generic size is larger, or the
// size is equal and the deviation larger in the order of interval.Compare;
// directions are ignored, so crossed pairs are measured by absolute interval.
// Errors, in this order: ErrUnset; ErrNoOrder; ErrOption for an unknown rule,
// DistanceCustom without limits, or limits with a built-in rule; then, when
// the norm applies, ErrOption for a custom number of limits that is neither 1
// nor the number of pairs.
func (s *Sonority) CheckDistances(norm DistanceNorm) (*DistanceAssessment, error) {
	if err := s.checkOrder(); err != nil {
		return nil, err
	}
	text, err := validateNorm(norm)
	if err != nil {
		return nil, err
	}
	out := &DistanceAssessment{
		Rule: norm.Rule.String(), Source: text.source, Scope: text.scope,
		Silent: slices.Clone(s.silent), Violations: []DistanceViolation{}, Remarks: []DistanceViolation{},
	}
	n := len(s.notes)
	switch {
	case len(s.silent) > 0:
		out.Reason = ReasonSilentParts
		return out, nil
	case !text.fitsVoices(n):
		out.Reason = ReasonVoiceCount
		return out, nil
	}
	limits, ok := text.pairLimits(norm, n)
	if !ok {
		return nil, ErrOption
	}
	out.Applicable = true
	out.Violations, out.Remarks = distanceFindings(s.adjacent, limits, norm.Rule)
	return out, nil
}

// MustCheckDistances is like CheckDistances but panics with its error.
func (s *Sonority) MustCheckDistances(norm DistanceNorm) *DistanceAssessment {
	return must(s.CheckDistances(norm))
}

func distanceFindings(adjacent []VoiceInterval, limits []interval.Interval, rule DistanceRule) (violations, remarks []DistanceViolation) {
	violations, remarks = []DistanceViolation{}, []DistanceViolation{}
	for k, v := range adjacent {
		if exceeds(v.Interval, limits[k]) {
			violations = append(violations, DistanceViolation{Pair: v, Limit: limits[k]})
		}
		if rule == DistanceOMTFirstSpecies && exceeds(v.Interval, interval.PerfectOctave()) {
			remarks = append(remarks, DistanceViolation{Pair: v, Limit: interval.PerfectOctave()})
		}
	}
	return violations, remarks
}
