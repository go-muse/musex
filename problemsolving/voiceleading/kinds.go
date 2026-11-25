package voiceleading

// LeapDefinition selects the smallest generic size that counts as a leap.
// Its zero value is unset; operations that take it reject unset and unknown
// values with ErrOption.
type LeapDefinition uint8

const (
	// LeapUnset is invalid; callers must choose a definition.
	LeapUnset LeapDefinition = iota
	// LeapFromThird follows Wikipedia: a third or wider is a leap.
	LeapFromThird
	// LeapFromFourth follows Alekseev and Myasoedov: a fourth or wider is a
	// leap, and a third is a smooth motion.
	LeapFromFourth
)

// String returns a technical label, or "unknown" for an unknown value.
func (d LeapDefinition) String() string {
	switch d {
	case LeapUnset:
		return "unset"
	case LeapFromThird:
		return "from-third"
	case LeapFromFourth:
		return "from-fourth"
	default:
		return "unknown"
	}
}

func (d LeapDefinition) valid() bool { return d == LeapFromThird || d == LeapFromFourth }

// MelodyRule selects an explicit educational norm for one voice.
type MelodyRule uint8

const (
	// MelodyRuleUnset is invalid; callers must choose a rule.
	MelodyRuleUnset MelodyRule = iota
	// OMTLeapRecovery requires a step in the opposite direction after a leap of
	// a fourth or wider (Open Music Theory, cantus firmus).
	OMTLeapRecovery
	// OMTMelodicIntervals forbids augmented, diminished, seventh, sixth other
	// than an ascending minor sixth, and wider-than-octave motions (Open Music
	// Theory, rules for all species).
	OMTMelodicIntervals
	// FuxAllowedMotions allows only P1, P4, P5, P8, minor and major seconds and
	// thirds, and an ascending minor sixth followed by downward motion
	// (Wikipedia, Counterpoint, after Fux).
	FuxAllowedMotions
	// MT21ChoraleMelody forbids augmented intervals and requires a change of
	// direction after a leap wider than a fourth or a diminished interval
	// (Music Theory for the 21st-Century Classroom, 26.5).
	MT21ChoraleMelody
)

// String returns the rule identifier, "unset", or "unknown".
func (r MelodyRule) String() string {
	switch r {
	case MelodyRuleUnset:
		return "unset"
	case OMTLeapRecovery:
		return "omt-leap-recovery"
	case OMTMelodicIntervals:
		return "omt-melodic-intervals"
	case FuxAllowedMotions:
		return "wp-fux-allowed-motions"
	case MT21ChoraleMelody:
		return "mt21-chorale-melody"
	default:
		return "unknown"
	}
}

// Reasons of melodic violations. Each violation lists its reasons in the
// order in which its rule enumerates them.
const (
	ReasonNoStepBack      = "no-step-back"
	ReasonEndOfLine       = "end-of-line"
	ReasonRestFollows     = "rest-follows"
	ReasonAugmented       = "augmented"
	ReasonDiminished      = "diminished"
	ReasonSeventh         = "seventh"
	ReasonSixth           = "sixth"
	ReasonWiderThanOctave = "wider-than-octave"
	ReasonNotAllowed      = "not-allowed"
	ReasonSameDirection   = "same-direction"
	ReasonRepeat          = "repeat"
)

// SpacingDefinition selects a definition of close and open spacing.
type SpacingDefinition uint8

const (
	// SpacingUnset is invalid; callers must choose a definition.
	SpacingUnset SpacingDefinition = iota
	// SpacingME follows the Muzykalnaya entsiklopediya: exactly four voices,
	// tenor-alto and alto-soprano are considered.
	SpacingME
	// SpacingAM follows Alekseev and Myasoedov: every adjacent pair, including
	// the bass, is considered.
	SpacingAM
	// SpacingOMT follows Open Music Theory: closed when the span does not
	// exceed a perfect octave.
	SpacingOMT
)

// String returns a technical label, or "unknown" for an unknown value.
func (d SpacingDefinition) String() string {
	switch d {
	case SpacingUnset:
		return "unset"
	case SpacingME:
		return "me"
	case SpacingAM:
		return "am"
	case SpacingOMT:
		return "omt"
	default:
		return "unknown"
	}
}

// SpacingClass is the class of a sonority under one spacing definition.
type SpacingClass uint8

const (
	// SpacingClassUnset marks an inapplicable classification.
	SpacingClassUnset SpacingClass = iota
	// CloseSpacing means that every considered interval is a fourth or narrower.
	CloseSpacing
	// OpenSpacing means that every considered interval is a fifth or wider.
	OpenSpacing
	// MixedSpacing means that the considered intervals are of both kinds.
	MixedSpacing
	// OMTClosedSpacing means that the span does not exceed a perfect octave.
	OMTClosedSpacing
	// OMTOpenSpacing means that the span exceeds a perfect octave.
	OMTOpenSpacing
)

// String returns a technical label, or "unknown" for an unknown value.
func (c SpacingClass) String() string {
	switch c {
	case SpacingClassUnset:
		return "unset"
	case CloseSpacing:
		return "close"
	case OpenSpacing:
		return "open"
	case MixedSpacing:
		return "mixed"
	case OMTClosedSpacing:
		return "omt-closed"
	case OMTOpenSpacing:
		return "omt-open"
	default:
		return "unknown"
	}
}

// Reasons why a spacing classification or a norm is not applicable.
const (
	ReasonSilentParts   = "silent-parts"
	ReasonVoiceCount    = "voice-count"
	ReasonCrossedVoices = "crossed-voices"
)

// DistanceRule selects a norm of distances between adjacent voices.
type DistanceRule uint8

const (
	// DistanceUnset is invalid as a norm; in ConnectionConditions it disables
	// the distance condition.
	DistanceUnset DistanceRule = iota
	// DistanceOMTSATB limits soprano-alto and alto-tenor to P8 and tenor-bass
	// to P12 in four voices (Open Music Theory, SATB style).
	DistanceOMTSATB
	// DistanceMT21SATB limits soprano-alto and alto-tenor to P8 in four voices
	// and leaves tenor-bass unlimited (MT21, 26.6).
	DistanceMT21SATB
	// DistanceOMTFirstSpecies limits two voices to P12 and remarks intervals
	// beyond P8 (Open Music Theory, first species).
	DistanceOMTFirstSpecies
	// DistanceCustom uses the composer's limits from DistanceNorm.Limits.
	DistanceCustom
)

// String returns a technical label, or "unknown" for an unknown value.
func (r DistanceRule) String() string {
	switch r {
	case DistanceUnset:
		return "unset"
	case DistanceOMTSATB:
		return "omt-satb"
	case DistanceMT21SATB:
		return "mt21-satb"
	case DistanceOMTFirstSpecies:
		return "omt-first-species"
	case DistanceCustom:
		return "custom"
	default:
		return "unknown"
	}
}

// Reading selects how tone classes are identified: by written name without
// an octave, or by the pitch class of the position in the system.
type Reading uint8

const (
	// BySpelling identifies a class by its written name: C, B sharp, and
	// D double flat differ. It is the default.
	BySpelling Reading = iota
	// BySound identifies a class by the pitch class in the system: C and
	// B sharp coincide in the traditional system.
	BySound
)

// String returns a technical label, or "unknown" for an unknown value.
func (r Reading) String() string {
	switch r {
	case BySpelling:
		return "spelling"
	case BySound:
		return "sound"
	default:
		return "unknown"
	}
}

// DoublingKind describes how the voices of one tone class relate.
type DoublingKind uint8

const (
	// NoDoubling means that one voice takes the class.
	NoDoubling DoublingKind = iota
	// UnisonDoubling means that every pair of voices of the class is a unison.
	UnisonDoubling
	// OctaveDoubling means that every pair of voices of the class lies in
	// different octaves.
	OctaveDoubling
	// UnisonAndOctaveDoubling means that the class has pairs of both kinds.
	UnisonAndOctaveDoubling
)

// String returns a technical label, or "unknown" for an unknown value.
func (k DoublingKind) String() string {
	switch k {
	case NoDoubling:
		return "none"
	case UnisonDoubling:
		return "unison"
	case OctaveDoubling:
		return "octave"
	case UnisonAndOctaveDoubling:
		return "unison-and-octave"
	default:
		return "unknown"
	}
}

// ConnectionKind classifies a connection of two sonorities.
type ConnectionKind uint8

const (
	// ConnectionUnset marks the zero Connection.
	ConnectionUnset ConnectionKind = iota
	// HarmonicConnection means that at least one voice keeps a common tone.
	HarmonicConnection
	// MelodicConnection means that no voice keeps a common tone.
	MelodicConnection
)

// String returns a technical label, or "unknown" for an unknown value.
func (k ConnectionKind) String() string {
	switch k {
	case ConnectionUnset:
		return "unset"
	case HarmonicConnection:
		return "harmonic"
	case MelodicConnection:
		return "melodic"
	default:
		return "unknown"
	}
}

// MT21ConnectionRule identifies the explicitly selected MT21 connection norm.
const MT21ConnectionRule = "mt21-root-position-triads"

// Reasons why the MT21 connection norm is not applicable.
const (
	ReasonFromNotThreeClasses = "from-not-three-classes"
	ReasonToNotThreeClasses   = "to-not-three-classes"
	ReasonFromBassNotDoubled  = "from-bass-not-doubled"
	ReasonToBassNotDoubled    = "to-bass-not-doubled"
	ReasonBassMotion          = "bass-motion-not-covered"
)

// MT21Variant names one variant of the MT21 connection norm.
type MT21Variant uint8

const (
	// MT21VariantUnset marks the zero VariantCheck.
	MT21VariantUnset MT21Variant = iota
	// MT21CommonToneAndSteps is variant 1 of 26.7.2 (bass by a fourth or
	// fifth): the common tone stays in one upper voice, the two others step.
	MT21CommonToneAndSteps
	// MT21SameDirection is variant 2 of 26.7.2 (bass by a fourth or fifth):
	// all three upper voices move in one direction.
	MT21SameDirection
	// MT21TwoCommonTones is 26.7.1 (bass by a third or sixth): two common
	// tones stay and the third upper voice steps.
	MT21TwoCommonTones
)

// String returns a technical label, or "unknown" for an unknown value.
func (v MT21Variant) String() string {
	switch v {
	case MT21VariantUnset:
		return "unset"
	case MT21CommonToneAndSteps:
		return "26.7.2-common-tone"
	case MT21SameDirection:
		return "26.7.2-same-direction"
	case MT21TwoCommonTones:
		return "26.7.1-two-common-tones"
	default:
		return "unknown"
	}
}
