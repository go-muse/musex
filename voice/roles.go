package voice

import (
	"fmt"

	"github.com/go-muse/musex/instrument"
)

// Placement is the role of a line by its place in the texture. The roles are
// labels of the user: the package never derives them from the notes. Their
// names coincide with presets of package instrument, but they are not
// presets.
type Placement uint8

// Roles by place. PlacementUnset is the role "not set".
const (
	// PlacementUnset is the zero value: the role is not set.
	PlacementUnset Placement = iota
	// Soprano is the soprano line.
	Soprano
	// Alto is the alto line.
	Alto
	// Tenor is the tenor line.
	Tenor
	// Bass is the bass line.
	Bass
	// Upper is the upper line.
	Upper
	// Middle is a middle line.
	Middle
	// Lower is the lower line.
	Lower
)

// ThematicRole is the role of a line in the thematic process.
type ThematicRole uint8

// Thematic roles. ThematicUnset is the role "not set".
const (
	// ThematicUnset is the zero value: the role is not set.
	ThematicUnset ThematicRole = iota
	// Leading is the leading line.
	Leading
	// Solo is a leading line performed by one singer or player.
	Solo
	// Accompanying is an accompanying line.
	Accompanying
	// Imitating is an imitating line.
	Imitating
	// Counterpoint is a counterpoint.
	Counterpoint
	// Doubling is a doubling.
	Doubling
	// Inversion is an inversion (contrary motion).
	Inversion
	// Pedal is a pedal.
	Pedal
)

// Importance is the role of a line by its importance in the texture. It is a
// label of the user: the package neither derives nor checks it.
type Importance uint8

// Roles by importance. ImportanceUnset is the role "not set".
const (
	// ImportanceUnset is the zero value: the role is not set.
	ImportanceUnset Importance = iota
	// Principal is a principal (obbligato) line.
	Principal
	// Subsidiary is a subsidiary line.
	Subsidiary
	// Supplementary is a supplementary line.
	Supplementary
	// Filler is a filler line.
	Filler
	// Harmonic is a harmonic line.
	Harmonic
)

// roleNames holds the English and Russian names of the values of a role kind,
// indexed by value; index 0 is the unset value.
type roleNames struct {
	english []string
	russian []string
}

var (
	placementNames = roleNames{
		english: []string{"", "soprano", "alto", "tenor", "bass", "upper", "middle", "lower"},
		russian: []string{"", "сопрано", "альт", "тенор", "бас", "верхний", "средний", "нижний"},
	}
	thematicNames = roleNames{
		english: []string{
			"", "leading", "solo", "accompanying", "imitating",
			"counterpoint", "doubling", "inversion", "pedal",
		},
		russian: []string{
			"", "ведущий", "сольный", "сопровождающий", "имитирующий",
			"контрапункт", "удвоение", "противодвижение", "педаль",
		},
	}
	importanceNames = roleNames{
		english: []string{"", "principal", "subsidiary", "supplementary", "filler", "harmonic"},
		russian: []string{"", "главный", "побочный", "дополняющий", "заполняющий", "гармонический"},
	}
)

// isValid reports whether v is a set value of the kind.
func (n roleNames) isValid(v uint8) bool {
	return v > 0 && int(v) < len(n.english)
}

func (n roleNames) name(v uint8) string {
	if !n.isValid(v) {
		return ""
	}

	return n.english[v]
}

func (n roleNames) russianName(v uint8) string {
	if !n.isValid(v) {
		return ""
	}

	return n.russian[v]
}

// parse returns the value whose English or Russian name equals text exactly.
func (n roleNames) parse(text, kind string) (uint8, error) {
	for v := 1; v < len(n.english); v++ {
		if text == n.english[v] || text == n.russian[v] {
			return uint8(v), nil
		}
	}

	return 0, fmt.Errorf("%w: %q is not a name of %s", ErrUnknownName, text, kind)
}

// values returns all set values of the kind in the order of declaration.
func values[T ~uint8](n roleNames) []T {
	out := make([]T, 0, len(n.english)-1)
	for v := 1; v < len(n.english); v++ {
		out = append(out, T(v))
	}

	return out
}

// roleString returns the name of v, or the unset or invalid form.
func roleString(n roleNames, v uint8, typ string) string {
	switch {
	case v == 0:
		return typ + "(unset)"
	case !n.isValid(v):
		return typ + "(invalid)"
	default:
		return n.english[v]
	}
}

// Placements returns a new slice of all roles by place in the order of
// declaration, without PlacementUnset.
func Placements() []Placement {
	return values[Placement](placementNames)
}

// ParsePlacement returns the role by place whose English or Russian name
// equals text exactly, such as "tenor" or "тенор". The case, spaces and
// normalization are not changed. Another text reports ErrUnknownName.
func ParsePlacement(text string) (Placement, error) {
	v, err := placementNames.parse(text, "a role by place")

	return Placement(v), err
}

// MustParsePlacement is like ParsePlacement but panics with its error.
func MustParsePlacement(text string) Placement {
	return must(ParsePlacement(text))
}

// IsValid reports whether p is a set role in the list.
func (p Placement) IsValid() bool {
	return placementNames.isValid(uint8(p))
}

// Name returns the English name, such as "soprano"; "" for PlacementUnset and
// values outside the list.
func (p Placement) Name() string {
	return placementNames.name(uint8(p))
}

// RussianName returns the Russian name, such as "сопрано"; "" for
// PlacementUnset and values outside the list.
func (p Placement) RussianName() string {
	return placementNames.russianName(uint8(p))
}

// String returns Name; "Placement(unset)" for PlacementUnset and
// "Placement(invalid)" for values outside the list.
func (p Placement) String() string {
	return roleString(placementNames, uint8(p), "Placement")
}

// IsOuter reports whether p is an outer line: Soprano, Bass, Upper or Lower.
func (p Placement) IsOuter() bool {
	return p == Soprano || p == Bass || p == Upper || p == Lower
}

// IsInner reports whether p is an inner line: Alto, Tenor or Middle.
func (p Placement) IsInner() bool {
	return p == Alto || p == Tenor || p == Middle
}

// Range returns the reference range of the role and true: the range of the
// built-in preset instrument.SopranoVoice, AltoVoice, TenorVoice or BassVoice
// for Soprano, Alto, Tenor and Bass. Other roles, PlacementUnset and values
// outside the list have no reference range: instrument.Range{} and false.
func (p Placement) Range() (instrument.Range, bool) {
	var preset *instrument.Preset

	switch p {
	case Soprano:
		preset = instrument.SopranoVoice()
	case Alto:
		preset = instrument.AltoVoice()
	case Tenor:
		preset = instrument.TenorVoice()
	case Bass:
		preset = instrument.BassVoice()
	default:
		return instrument.Range{}, false
	}

	return preset.Range(), true
}

// ThematicRoles returns a new slice of all thematic roles in the order of
// declaration, without ThematicUnset.
func ThematicRoles() []ThematicRole {
	return values[ThematicRole](thematicNames)
}

// ParseThematicRole returns the thematic role whose English or Russian name
// equals text exactly, such as "doubling" or "удвоение". Another text,
// including a name of another kind of role, reports ErrUnknownName.
func ParseThematicRole(text string) (ThematicRole, error) {
	v, err := thematicNames.parse(text, "a thematic role")

	return ThematicRole(v), err
}

// MustParseThematicRole is like ParseThematicRole but panics with its error.
func MustParseThematicRole(text string) ThematicRole {
	return must(ParseThematicRole(text))
}

// IsValid reports whether r is a set role in the list.
func (r ThematicRole) IsValid() bool {
	return thematicNames.isValid(uint8(r))
}

// Name returns the English name, such as "imitating"; "" for ThematicUnset and
// values outside the list.
func (r ThematicRole) Name() string {
	return thematicNames.name(uint8(r))
}

// RussianName returns the Russian name, such as "имитирующий"; "" for
// ThematicUnset and values outside the list.
func (r ThematicRole) RussianName() string {
	return thematicNames.russianName(uint8(r))
}

// String returns Name; "ThematicRole(unset)" for ThematicUnset and
// "ThematicRole(invalid)" for values outside the list.
func (r ThematicRole) String() string {
	return roleString(thematicNames, uint8(r), "ThematicRole")
}

// Importances returns a new slice of all roles by importance in the order of
// declaration, without ImportanceUnset.
func Importances() []Importance {
	return values[Importance](importanceNames)
}

// ParseImportance returns the role by importance whose English or Russian
// name equals text exactly, such as "principal" or "главный". Another text
// reports ErrUnknownName.
func ParseImportance(text string) (Importance, error) {
	v, err := importanceNames.parse(text, "a role by importance")

	return Importance(v), err
}

// MustParseImportance is like ParseImportance but panics with its error.
func MustParseImportance(text string) Importance {
	return must(ParseImportance(text))
}

// IsValid reports whether i is a set role in the list.
func (i Importance) IsValid() bool {
	return importanceNames.isValid(uint8(i))
}

// Name returns the English name, such as "principal"; "" for ImportanceUnset
// and values outside the list.
func (i Importance) Name() string {
	return importanceNames.name(uint8(i))
}

// RussianName returns the Russian name, such as "главный"; "" for
// ImportanceUnset and values outside the list.
func (i Importance) RussianName() string {
	return importanceNames.russianName(uint8(i))
}

// String returns Name; "Importance(unset)" for ImportanceUnset and
// "Importance(invalid)" for values outside the list.
func (i Importance) String() string {
	return roleString(importanceNames, uint8(i), "Importance")
}

// RetrogradeKind is the kind of a retrograde.
type RetrogradeKind uint8

// Kinds of a retrograde. RetrogradeUnset is not a kind.
const (
	// RetrogradeUnset is the zero value: no kind.
	RetrogradeUnset RetrogradeKind = iota
	// RetrogradePitches keeps the rhythm and puts the notes of the sounds in
	// reverse order.
	RetrogradePitches
	// RetrogradeRhythm reverses the rhythm and keeps the order of the notes
	// of the sounds.
	RetrogradeRhythm
	// RetrogradeBoth reverses the order of the elements.
	RetrogradeBoth
)

var retrogradeKindNames = [...]string{
	RetrogradePitches: "pitches",
	RetrogradeRhythm:  "rhythm",
	RetrogradeBoth:    "both",
}

// IsValid reports whether k is one of the three kinds.
func (k RetrogradeKind) IsValid() bool {
	return k >= RetrogradePitches && k <= RetrogradeBoth
}

// String returns "pitches", "rhythm" or "both"; "RetrogradeKind(unset)" for
// RetrogradeUnset and "RetrogradeKind(invalid)" for values outside the list.
func (k RetrogradeKind) String() string {
	return enumString(k, retrogradeKindNames[:], "RetrogradeKind")
}

// enumString returns names[v], or the unset or invalid form of typ.
func enumString[T ~uint8](v T, names []string, typ string) string {
	switch {
	case v == 0:
		return typ + "(unset)"
	case int(v) < len(names) && names[v] != "":
		return names[v]
	default:
		return typ + "(invalid)"
	}
}
