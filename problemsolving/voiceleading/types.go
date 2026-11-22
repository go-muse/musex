package voiceleading

import (
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/voice"
)

// Order declares the nominal order of input parts; it never sorts their notes.
type Order uint8

const (
	// Unordered leaves the relative roles of the parts unspecified.
	Unordered Order = iota
	// LowToHigh declares that input parts are ordered from lowest to highest.
	LowToHigh
)

// Options selects a common system and an optional nominal part order.
// A nil System selects the system from the notes that need measurement.
// A non-nil System is borrowed and must remain immutable.
type Options struct {
	System *system.System
	Order  Order
}

// Pair identifies input slots, with 0 <= First < Second < PartCount.
// Names, line equality, and line pointer identity do not identify parts.
type Pair struct{ First, Second int }

// Motion describes the written melodic movements of two parts.
type Motion uint8

const (
	// MotionUnclassified marks a transition interrupted by silence.
	MotionUnclassified Motion = iota
	// MotionStationary means that both melodic intervals are perfect unisons.
	MotionStationary
	// MotionOblique means that exactly one part remains stationary.
	MotionOblique
	// MotionContrary means that the two written directions differ.
	MotionContrary
	// MotionDirect means equal directions with different full generic sizes.
	MotionDirect
	// MotionParallel means equal directions and equal full generic sizes.
	MotionParallel
)

// IsDirect reports whether the motion is direct, including parallel motion.
func (m Motion) IsDirect() bool { return m == MotionDirect || m == MotionParallel }

// String returns a technical label, or "unknown" for an unknown value.
func (m Motion) String() string {
	switch m {
	case MotionUnclassified:
		return "unclassified"
	case MotionStationary:
		return "stationary"
	case MotionOblique:
		return "oblique"
	case MotionContrary:
		return "contrary"
	case MotionDirect:
		return "direct"
	case MotionParallel:
		return "parallel"
	default:
		return "unknown"
	}
}

// PerfectSet selects perfect interval families. Zero is not a valid filter.
type PerfectSet uint8

const (
	// PerfectFifths selects sizes 5, 12, 19, and their octave extensions.
	PerfectFifths PerfectSet = 1
	// PerfectOctaves selects sizes 1, 8, 15, and their octave extensions.
	PerfectOctaves PerfectSet = 2
	// PerfectBoth selects both families.
	PerfectBoth PerfectSet = PerfectFifths | PerfectOctaves
)

// ParallelOptions selects strict parallels. Without Compound or Unisons,
// only P5 and P8 qualify. Both ends must have the same full size.
type ParallelOptions struct {
	Perfect  PerfectSet
	Compound bool
	Unisons  bool
}

// DirectOrigin selects the origin condition for a direct arrival.
type DirectOrigin uint8

const (
	// OriginUnset is invalid; callers must choose an interpretation.
	OriginUnset DirectOrigin = iota
	// AnyOrigin accepts every starting interval.
	AnyOrigin
	// DifferentFamily excludes a perfect origin in the target's family.
	DifferentFamily
)

// ElementAt preserves the original element index and its complete time span.
// Element is immutable; its Note method returns an independent note copy.
type ElementAt struct {
	Index      int
	Element    voice.Element
	Start, End duration.Position
}

// Vertical is one half-open pairwise segment [Start, End).
// Specific and Chromatic run from First to Second. If either part rests,
// Sounding is false, Specific is unset, and Chromatic is zero.
type Vertical struct {
	Pair          Pair
	Start, End    duration.Position
	First, Second ElementAt
	Sounding      bool
	Specific      interval.Interval
	Chromatic     system.Step
}

// Transition connects adjacent pairwise verticals. If either end contains a
// rest, both melodic intervals are unset, both distances zero, and the motion
// unclassified. Otherwise intervals retain full written sizes and directions.
type Transition struct {
	From, To                        Vertical
	First, Second                   interval.Interval
	FirstChromatic, SecondChromatic system.Step
	Motion                          Motion
}

// Crossing is a maximal run where the nominally lower part is strictly higher.
// Verticals belongs to the caller; copying Crossing shares that caller's slice.
type Crossing struct {
	Pair       Pair
	Start, End duration.Position
	Verticals  []Vertical
}

// Overlap records a new entry past the neighbor's previous height.
// Both flags may be true in one transition.
type Overlap struct {
	Transition               Transition
	FirstAbovePreviousSecond bool
	SecondBelowPreviousFirst bool
}

// OMTFirstSpeciesRule identifies the explicitly selected educational rule.
const OMTFirstSpeciesRule = "omt-first-species-consecutive-perfects"

// ScopeIssue explains why one sounding vertical is outside the rule's scope.
type ScopeIssue struct {
	Vertical Vertical
	Reason   string
}

// Violation explains a consecutive pair of perfect intervals in one family.
type Violation struct {
	Transition Transition
	Reason     string
}

// Assessment is an independent result of an explicit rule check. Issues and
// Violations belong to the caller. If Issues is nonempty, Violations is empty.
type Assessment struct {
	Rule       string
	Scope      string
	Applicable bool
	Issues     []ScopeIssue
	Violations []Violation
}
