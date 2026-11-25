package voiceleading

import (
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/spelling"
)

// VariantCheck reports whether one variant of the MT21 norm is satisfied.
type VariantCheck struct {
	Variant   MT21Variant
	Satisfied bool
}

// ConnectionAssessment is an independent result of the MT21 connection norm.
// Connection always holds the description. When not applicable, Reason names
// the first failed scope condition, Variants is empty and Conforms is false.
// Variants belongs to the caller.
type ConnectionAssessment struct {
	Rule       string
	Source     string
	Scope      string
	Applicable bool
	Reason     string
	Connection Connection
	Variants   []VariantCheck
	Conforms   bool
}

const (
	mt21Source = "Music Theory for the 21st-Century Classroom, 26.7.1, 26.7.2"
	mt21Scope  = "four sounding voices; root-position triads with a doubled bass, as asserted by the composer"
)

// CheckMT21Connection applies the explicitly selected MT21 norm for connecting
// root-position triads with a doubled bass. Selecting it asserts that both
// chords are such triads; the root is not checked. The scope requires four
// voices, three classes by spelling in each sonority, and the bass class in an
// upper voice of each. For a bass motion by a fourth or fifth (simple size)
// the variants are MT21CommonToneAndSteps and MT21SameDirection; by a third or
// sixth, MT21TwoCommonTones; other bass motions are not covered. The
// connection conforms when at least one variant is satisfied.
// Errors are those of Connect.
func CheckMT21Connection(from, to *Sonority) (*ConnectionAssessment, error) {
	c, err := Connect(from, to)
	if err != nil {
		return nil, err
	}
	out := &ConnectionAssessment{Rule: MT21ConnectionRule, Source: mt21Source, Scope: mt21Scope, Connection: c, Variants: []VariantCheck{}}
	if out.Reason = mt21Scope4(from.notes, to.notes, c.Moves[0].Interval); out.Reason != "" {
		return out, nil
	}
	out.Applicable = true
	out.Variants = mt21Variants(c.Moves)
	for _, v := range out.Variants {
		out.Conforms = out.Conforms || v.Satisfied
	}
	return out, nil
}

// MustCheckMT21Connection is like CheckMT21Connection but panics with its error.
func MustCheckMT21Connection(from, to *Sonority) *ConnectionAssessment {
	return must(CheckMT21Connection(from, to))
}

// mt21Scope4 returns the first reason why two sonorities with the given bass
// motion are outside the norm, or "".
func mt21Scope4(from, to []note.Note, bass interval.Interval) string {
	switch {
	case len(from) != 4:
		return ReasonVoiceCount
	case classCount(from) != 3:
		return ReasonFromNotThreeClasses
	case classCount(to) != 3:
		return ReasonToNotThreeClasses
	case !bassDoubled(from):
		return ReasonFromBassNotDoubled
	case !bassDoubled(to):
		return ReasonToBassNotDoubled
	case !bassMotionCovered(bass):
		return ReasonBassMotion
	}
	return ""
}

// classCount returns the number of distinct classes by spelling.
func classCount(notes []note.Note) int {
	var seen []spelling.Spelling
	for _, n := range notes {
		if !containsName(seen, n.Spelling()) {
			seen = append(seen, n.Spelling())
		}
	}
	return len(seen)
}

func containsName(names []spelling.Spelling, name spelling.Spelling) bool {
	for _, x := range names {
		if x.Equal(name) {
			return true
		}
	}
	return false
}

// bassDoubled reports whether an upper voice takes the class of voice 0.
func bassDoubled(notes []note.Note) bool {
	for _, n := range notes[1:] {
		if n.Spelling().Equal(notes[0].Spelling()) {
			return true
		}
	}
	return false
}

func bassMotionCovered(bass interval.Interval) bool {
	n := bass.SimpleSize()
	return n >= 3 && n <= 6
}

// mt21Variants checks the variants of the bass motion of four moves.
func mt21Variants(moves []VoiceMove) []VariantCheck {
	upper := moves[1:]
	kept, steps := 0, 0
	for _, mv := range upper {
		switch {
		case mv.IsRepeat():
			kept++
		case mv.IsStep():
			steps++
		}
	}
	switch moves[0].Interval.SimpleSize() {
	case 4, 5:
		same := true
		for _, mv := range upper {
			same = same && !mv.Interval.IsPerfectUnison() && mv.Interval.Direction() == upper[0].Interval.Direction()
		}
		return []VariantCheck{
			{Variant: MT21CommonToneAndSteps, Satisfied: kept == 1 && steps == 2},
			{Variant: MT21SameDirection, Satisfied: same},
		}
	default:
		return []VariantCheck{{Variant: MT21TwoCommonTones, Satisfied: kept == 2 && steps == 1}}
	}
}
