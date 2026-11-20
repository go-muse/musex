// Package tendency answers finite musical questions under explicitly selected rules.
package tendency

import (
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/mode"
)

// ID identifies an explicitly selected built-in or user scheme.
type ID string

// Task identifies the kind of musical question.
type Task uint8

const (
	TaskUnset Task = iota
	SupportTask
	RoleTask
	TendencyTask
	ClassificationTask
	ModalResolutionTask
	IntervalResolutionTask
)

// Knowledge records rule coverage independently of a musical decision.
type Knowledge uint8

const (
	KnowledgeUnset Knowledge = iota
	Known
	Partial
	MissingRule
	MissingContext
	OutsideScope
	UnsupportedTask
)

// Decision distinguishes confirmed permission, a proved violation, and uncertainty.
type Decision uint8

const (
	DecisionUnset Decision = iota
	Allowed
	Denied
	Undetermined
)

// Motion is a symbolic movement direction; its zero value is unspecified.
type Motion uint8

const (
	MotionUnset Motion = iota
	Stationary
	Ascending
	Descending
)

// Origin identifies source, product, or user provenance.
type Origin uint8

const (
	OriginUnset Origin = iota
	SourceRule
	ProductRule
	UserRule
)

// Place identifies a degree and a signed cycle; open modes require cycle zero.
type Place struct {
	Degree int
	Cycle  int
}

// Pair preserves the identities of two voices without sorting them.
type Pair struct {
	First  Place
	Second Place
}

// Motions contains two specific movement intervals; P1 means no movement.
type Motions struct {
	First  interval.Interval
	Second interval.Interval
}

// Fact is a named input or condition attached to a reason.
type Fact struct {
	Name  string
	Value string
}

// Reason explains an applied rule with its origin, reference, and sorted facts.
type Reason struct {
	Rule        string
	Origin      Origin
	Reference   string
	Explanation string
	Facts       []Fact
}

// Answer identifies the selected scheme and the completeness of its answer.
type Answer struct {
	Scheme    ID
	Name      string
	Knowledge Knowledge
	Reasons   []Reason
}

// Movement connects two exact places with a specific symbolic interval.
type Movement struct {
	From      Place
	To        Place
	Interval  interval.Interval
	Direction Motion
}

// Scheme is an immutable rule snapshot. Construct it with Builtin or New; nil and zero values are unset. Copies share immutable state and are safe for concurrent reads.
type Scheme struct {
	state *schemeState
	_     [0]func()
}

// Info describes a scheme and its scope. Returned slices belong to the caller.
type Info struct {
	ID              ID
	Name            string
	Description     string
	Tasks           []Task
	Scope           string
	RequiredContext []string
	References      []string
	Builtin         bool
}

// CatalogQuery optionally filters built-ins by task and compatible modal structure.
type CatalogQuery struct {
	Task Task
	Mode mode.Mode
}

// SupportKind distinguishes unstable degrees, relative supports, and the main support.
type SupportKind uint8

const (
	SupportUnset SupportKind = iota
	Unstable
	RelativeSupport
	MainSupport
)

// SupportValue describes support and an optional positive rank; smaller ranks are stronger.
type SupportValue struct {
	Kind   SupportKind
	Rank   int
	Ranked bool
}

// SupportResult answers the support question for one actual degree.
type SupportResult struct {
	Answer
	Degree int
	Value  SupportValue
}

// SupportGroup collects known degrees of one support kind.
type SupportGroup struct {
	Kind    SupportKind
	Degrees []int
}

// RankGroup collects known supports of equal strength.
type RankGroup struct {
	Rank    int
	Degrees []int
}

// SupportAnalysis keeps every row, support groups, and ranked groups.
type SupportAnalysis struct {
	Answer
	Rows   []SupportResult
	Groups []SupportGroup
	Order  []RankGroup
}

// RankResult compares two known support ranks with Order -1, 0, or +1.
type RankResult struct {
	Answer
	First  int
	Second int
	Order  int
}

// RoleResult contains a degree label; a known empty label means no label.
type RoleResult struct {
	Answer
	Degree int
	Role   string
}

// RoleGroup collects degrees with the same case-sensitive label.
type RoleGroup struct {
	Role    string
	Degrees []int
}

// RoleAnalysis keeps every degree row and known label groups.
type RoleAnalysis struct {
	Answer
	Rows   []RoleResult
	Groups []RoleGroup
}

// TargetKind distinguishes an immediate move, an ultimate goal, and a named goal.
type TargetKind uint8

const (
	TargetUnset TargetKind = iota
	Immediate
	Ultimate
	Named
)

// Target describes a goal degree and optionally its exact cycle. Unknown cycles have Cycle zero.
type Target struct {
	Kind       TargetKind
	Degree     int
	Cycle      int
	CycleKnown bool
}

// Goal preserves a target kind, movement when known, preference, and reasons.
type Goal struct {
	Target    Target
	Direction Motion
	Interval  interval.Interval
	Preferred bool
	Reasons   []Reason
}

// TendencyResult reports direction, goal-list, and direct-transition coverage separately.
type TendencyResult struct {
	Answer
	Source             Place
	DirectionKnowledge Knowledge
	Directions         []Motion
	GoalKnowledge      Knowledge
	DirectKnowledge    Knowledge
	Goals              []Goal
}

// TendencyAnalysis keeps the result for every degree in the requested cycle.
type TendencyAnalysis struct {
	Answer
	Rows []TendencyResult
}

// AcousticKind identifies consonance or dissonance under one selected table.
type AcousticKind uint8

const (
	AcousticUnset AcousticKind = iota
	Consonance
	Dissonance
)

// ConsonanceKind records a subtype only when the selected table establishes it.
type ConsonanceKind uint8

const (
	ConsonanceUnset ConsonanceKind = iota
	AbsoluteConsonance
	PerfectConsonance
	ImperfectConsonance
)

// Classification contains the table classification, optional subtype, and context caveat.
type Classification struct {
	Kind    AcousticKind
	Subtype ConsonanceKind
	Caveat  string
}

// ClassificationResult preserves the input and the ascending interval actually classified.
type ClassificationResult struct {
	Answer
	Input      interval.Interval
	Classified interval.Interval
	Value      Classification
}

// ModalCandidate preserves both voices, exact movements, and the final interval.
type ModalCandidate struct {
	Target    Pair
	First     Movement
	Second    Movement
	Final     interval.Interval
	Preferred bool
	Reasons   []Reason
}

// ModalResolutions partitions generated candidates into allowed, uncertain, and rejected lists.
type ModalResolutions struct {
	Answer
	Source       Pair
	Initial      interval.Interval
	Allowed      []ModalCandidate
	Undetermined []ModalCandidate
	Rejected     []ModalCandidate
}

// IntervalCandidate preserves both specific movements even when final intervals coincide.
type IntervalCandidate struct {
	Motions   Motions
	Final     interval.Interval
	Preferred bool
	Reasons   []Reason
}

// IntervalResolutions partitions concrete voice-movement combinations by decision.
type IntervalResolutions struct {
	Answer
	Initial      interval.Interval
	Allowed      []IntervalCandidate
	Undetermined []IntervalCandidate
	Rejected     []IntervalCandidate
}

// CheckResult keeps rule coverage and the musical decision separate.
type CheckResult struct {
	Answer
	Decision Decision
}

// Basis provides mandatory user rule identity and explanation with optional references.
type Basis struct {
	Rule        string
	Explanation string
	References  []string
}

// SupportRule defines one degree support value.
type SupportRule struct {
	Degree int
	Value  SupportValue
	Bases  []Basis
}

// RoleRule defines one case-sensitive degree label, including an empty label.
type RoleRule struct {
	Degree int
	Role   string
	Bases  []Basis
}

// GoalRule defines one target with an optional direction assertion.
type GoalRule struct {
	Target    Target
	Direction Motion
	Preferred bool
	Bases     []Basis
}

// TendencyRule defines one degree row; nil lists and complete empty lists have different meanings.
type TendencyRule struct {
	Degree         int
	Directions     []Motion
	Goals          []GoalRule
	DirectComplete bool
	Bases          []Basis
}

// ModalOption defines a permitted target pair and explicitly allows crossing if required.
type ModalOption struct {
	Target        Pair
	AllowCrossing bool
	Preferred     bool
	Bases         []Basis
}

// ModalRule defines a finite source pair; Repeat enables joint cycle translation.
type ModalRule struct {
	Source  Pair
	Repeat  bool
	Options []ModalOption
	Bases   []Basis
}

// IntervalOption defines permitted specific voice movements, including optional crossing.
type IntervalOption struct {
	Motions       Motions
	AllowCrossing bool
	Preferred     bool
	Bases         []Basis
}

// IntervalRule defines a complete finite row keyed by the full specific source interval.
type IntervalRule struct {
	Source  interval.Interval
	Options []IntervalOption
	Bases   []Basis
}

// Definition contains finite user rules. Nil sections are unsupported; present empty sections have no known rows. New validates and copies all data.
type Definition struct {
	ID          ID
	Name        string
	Description string
	Mode        mode.Mode
	Supports    []SupportRule
	Roles       []RoleRule
	Tendencies  []TendencyRule
	Modal       []ModalRule
	Intervals   []IntervalRule
}

// Keep selects which voices must remain stationary during filtering.
type Keep uint8

const (
	KeepNone Keep = iota
	KeepFirst
	KeepSecond
	KeepBoth
)

// GoalFilter intersects goal constraints; nil masks disable a constraint and empty masks reject all.
type GoalFilter struct {
	Targets    []Place
	Directions []Motion
	Support    *Scheme
}

// ModalFilter intersects final-place, direction, stationary-voice, and support constraints.
type ModalFilter struct {
	FirstTargets     []Place
	SecondTargets    []Place
	FirstDirections  []Motion
	SecondDirections []Motion
	Keep             Keep
	Support          *Scheme
}

// IntervalFilter intersects movement directions and stationary-voice requirements.
type IntervalFilter struct {
	FirstDirections  []Motion
	SecondDirections []Motion
	Keep             Keep
}

// Excluded retains an excluded value together with the filter reasons.
type Excluded[T any] struct {
	Value   T
	Reasons []Reason
}

// Selection owns independent original, matched, uncertain, and excluded values.
type Selection[R any, V any] struct {
	Original  R
	Matched   []V
	Uncertain []V
	Excluded  []Excluded[V]
	Reasons   []Reason
}

// GoalMatch selects exact target identity or a coarser grouping by degree number.
type GoalMatch uint8

const (
	ExactGoal GoalMatch = iota
	DegreeGoal
)

// GoalKey is a target comparison key; DegreeGoal leaves only Degree set.
type GoalKey struct {
	Target Target
}

// Shared partitions scheme indices into present, absent, and unknown sets.
type Shared[T any] struct {
	Value     T
	Present   []int
	Absent    []int
	Unknown   []int
	Preferred []int
}

// Comparison preserves scheme order, complete answers, and shared-value partitions.
type Comparison[R any, V any] struct {
	Schemes []Info
	Answers []R
	Items   []Shared[V]
}
