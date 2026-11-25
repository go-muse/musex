package voiceleading

import "github.com/go-muse/musex/interval"

// MelodyViolation explains one move that breaks a melodic norm. Index is the
// move's index in Melody.Moves. Next and HasNext are set exactly when a reason
// concerns the following move (no-step-back, same-direction, repeat);
// otherwise Next is the zero Move. Reasons is nonempty and follows the order
// of the rule. Reasons belongs to the caller.
type MelodyViolation struct {
	Index   int
	Move    Move
	Next    Move
	HasNext bool
	Reasons []string
}

// MelodyAssessment is an independent result of an explicit melodic norm.
// A norm applies to any voice, so Applicable is always true; an empty or
// silent line has no violations. Violations belongs to the caller.
type MelodyAssessment struct {
	Rule       string
	Source     string
	Scope      string
	Applicable bool
	Violations []MelodyViolation
}

// melodyRuleText returns the source and scope of a rule, or false.
func melodyRuleText(rule MelodyRule) (source, scope string, ok bool) {
	switch rule {
	case OMTLeapRecovery:
		return "Open Music Theory, cantus firmus",
			"one voice; a leap of a fourth or wider is followed by a step in the opposite direction within the same fragment", true
	case OMTMelodicIntervals:
		return "Open Music Theory, rules for all species",
			"one voice; no augmented, diminished, seventh, sixth other than an ascending minor sixth, or wider than an octave", true
	case FuxAllowedMotions:
		return "Wikipedia, Counterpoint, considerations for all species",
			"one voice; only P1, P4, P5, P8, minor and major seconds and thirds, and an ascending minor sixth followed by downward motion", true
	case MT21ChoraleMelody:
		return "Music Theory for the 21st-Century Classroom, 26.5",
			"one voice; no augmented intervals; the direction changes after a leap wider than a fourth or a diminished interval", true
	default:
		return "", "", false
	}
}

// Check applies one explicitly selected melodic norm to every move that is not
// through a rest. The following move of a move is the next move of the same
// fragment; a move through a rest ends a fragment.
//
//   - OMTLeapRecovery: a move of a fourth or wider, whatever the leap
//     definition, must be followed by a step in the opposite direction:
//     otherwise "no-step-back"; without a following move "end-of-line" at the
//     end of the line or "rest-follows" before a rest.
//   - OMTMelodicIntervals: reasons "augmented", "diminished", "seventh",
//     "sixth" (except an ascending minor sixth) and "wider-than-octave", for
//     every move including unisons and seconds.
//   - FuxAllowedMotions: any motion other than P1, P4, P5, P8, minor and major
//     seconds and thirds and an ascending minor sixth is "not-allowed"; after an
//     ascending minor sixth the following move must go down, otherwise
//     "repeat" or "same-direction". Without a following move nothing is due.
//   - MT21ChoraleMelody: "augmented" for an augmented move; after a move of a
//     fifth or wider or a diminished move, the following move must change
//     direction, otherwise "repeat" or "same-direction".
//
// Errors, in this order: ErrUnset; ErrOption for an unknown rule.
func (m *Melody) Check(rule MelodyRule) (*MelodyAssessment, error) {
	if err := m.check(); err != nil {
		return nil, err
	}
	source, scope, ok := melodyRuleText(rule)
	if !ok {
		return nil, ErrOption
	}
	out := &MelodyAssessment{Rule: rule.String(), Source: source, Scope: scope, Applicable: true, Violations: []MelodyViolation{}}
	for k, mv := range m.moves {
		if mv.ThroughRest {
			continue
		}
		var next Move
		hasNext := k+1 < len(m.moves) && !m.moves[k+1].ThroughRest
		if hasNext {
			next = m.moves[k+1]
		}
		var reasons []string
		var usesNext bool
		switch rule {
		case OMTLeapRecovery:
			reasons, usesNext = m.leapRecovery(mv, next, hasNext)
		case OMTMelodicIntervals:
			reasons = melodicIntervals(mv)
		case FuxAllowedMotions:
			reasons, usesNext = fuxMotions(mv, next, hasNext)
		case MT21ChoraleMelody:
			reasons, usesNext = mt21Melody(mv, next, hasNext)
		}
		if len(reasons) == 0 {
			continue
		}
		v := MelodyViolation{Index: k, Move: mv, Reasons: reasons}
		if usesNext {
			v.Next, v.HasNext = next, true
		}
		out.Violations = append(out.Violations, v)
	}
	return out, nil
}

// MustCheck is like Check but panics with its error.
func (m *Melody) MustCheck(rule MelodyRule) *MelodyAssessment { return must(m.Check(rule)) }

func (m *Melody) leapRecovery(mv, next Move, hasNext bool) ([]string, bool) {
	if mv.Interval.Size() < 4 {
		return nil, false
	}
	if hasNext {
		if next.IsStep() && next.Interval.Direction() != mv.Interval.Direction() {
			return nil, false
		}
		return []string{ReasonNoStepBack}, true
	}
	if after := mv.To.Index + 1; after < len(m.elements) && m.elements[after].Element.IsRest() {
		return []string{ReasonRestFollows}, false
	}
	return []string{ReasonEndOfLine}, false
}

func melodicIntervals(mv Move) []string {
	q, n := mv.Interval.Quality(), mv.Interval.Size()
	var out []string
	if q == interval.Augmented {
		out = append(out, ReasonAugmented)
	}
	if q == interval.Diminished {
		out = append(out, ReasonDiminished)
	}
	if n == 7 {
		out = append(out, ReasonSeventh)
	}
	if n == 6 && (q != interval.Minor || mv.Interval.Direction() != interval.Up) {
		out = append(out, ReasonSixth)
	}
	if n > 8 {
		out = append(out, ReasonWiderThanOctave)
	}
	return out
}

func fuxMotions(mv, next Move, hasNext bool) ([]string, bool) {
	q, n, dir := mv.Interval.Quality(), mv.Interval.Size(), mv.Interval.Direction()
	ascendingMinorSixth := n == 6 && q == interval.Minor && dir == interval.Up
	allowed := q == interval.Perfect && (n == 1 || n == 4 || n == 5 || n == 8) ||
		(n == 2 || n == 3) && (q == interval.Major || q == interval.Minor) || ascendingMinorSixth
	if !allowed {
		return []string{ReasonNotAllowed}, false
	}
	if !ascendingMinorSixth || !hasNext {
		return nil, false
	}
	if next.IsRepeat() {
		return []string{ReasonRepeat}, true
	}
	if next.Interval.Direction() == interval.Up {
		return []string{ReasonSameDirection}, true
	}
	return nil, false
}

func mt21Melody(mv, next Move, hasNext bool) ([]string, bool) {
	q, n := mv.Interval.Quality(), mv.Interval.Size()
	var out []string
	if q == interval.Augmented {
		out = append(out, ReasonAugmented)
	}
	if !hasNext || n < 5 && q != interval.Diminished {
		return out, false
	}
	switch {
	case next.IsRepeat():
		return append(out, ReasonRepeat), true
	case next.Interval.Direction() == mv.Interval.Direction():
		return append(out, ReasonSameDirection), true
	default:
		return out, false
	}
}
