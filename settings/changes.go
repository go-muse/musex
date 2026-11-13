package settings

import (
	"strconv"
	"strings"

	"github.com/go-muse/musex/system"
)

// StructureChange is the kind of change of the system when one set of
// settings replaces another, by the admissible names of the two systems.
type StructureChange uint8

// Kinds of change of the structure. The numbers of the constants are stable.
const (
	// StructureUnset is the zero value: not computed.
	StructureUnset StructureChange = iota
	// StructureUnchanged: the systems are Equal.
	StructureUnchanged
	// StructureExtended: the previous admissible names are a strict part of
	// the next ones.
	StructureExtended
	// StructureNarrowed: the next admissible names are a strict part of the
	// previous ones.
	StructureNarrowed
	// StructureSameNames: the systems differ, but admit the same names.
	StructureSameNames
	// StructureIncomparable: neither system admits all names of the other.
	StructureIncomparable
)

// String returns "unchanged", "extended", "narrowed", "same-names" or
// "incomparable", "StructureChange(unset)" for StructureUnset and
// "StructureChange(<n>)" for other values. It is a debugging form.
func (c StructureChange) String() string {
	switch c {
	case StructureUnset:
		return "StructureChange(unset)"
	case StructureUnchanged:
		return "unchanged"
	case StructureExtended:
		return "extended"
	case StructureNarrowed:
		return "narrowed"
	case StructureSameNames:
		return "same-names"
	case StructureIncomparable:
		return "incomparable"
	default:
		return "StructureChange(" + strconv.Itoa(int(c)) + ")"
	}
}

// isDefined reports whether c is one of the five defined kinds.
func (c StructureChange) isDefined() bool {
	return StructureUnchanged <= c && c <= StructureIncomparable
}

// Changes lists the kinds of change when one set of settings replaces
// another. One replacement can have several kinds. The zero Changes is not
// computed: IsValid and None report false.
type Changes struct {
	// Structure is the change of the system by its admissible names.
	Structure StructureChange
	// Sound reports a change of the temperament, of the root number (as a
	// number, without the modulus) or of the reference.
	Sound bool
	// Designations reports a change of the notation, of the numbering or of
	// the catalog.
	Designations bool
}

// IsValid reports whether Structure is one of the five defined kinds.
func (c Changes) IsValid() bool {
	return c.Structure.isDefined()
}

// None reports whether c is valid and has no change: the structure is
// unchanged and neither the sound nor the designations change. For two set
// sets it is true exactly when they are Equal.
func (c Changes) None() bool {
	return c.IsValid() && c.Structure == StructureUnchanged && !c.Sound && !c.Designations
}

// String returns a debugging form such as "Changes(structure: extended,
// sound)", "Changes(none)" or "Changes(unset)".
func (c Changes) String() string {
	if !c.IsValid() {
		if c == (Changes{}) {
			return "Changes(unset)"
		}
		return "Changes(structure: " + c.Structure.String() + ", sound: " +
			strconv.FormatBool(c.Sound) + ", designations: " + strconv.FormatBool(c.Designations) + ")"
	}
	if c.None() {
		return "Changes(none)"
	}
	var parts []string
	if c.Structure != StructureUnchanged {
		parts = append(parts, "structure: "+c.Structure.String())
	}
	if c.Sound {
		parts = append(parts, "sound")
	}
	if c.Designations {
		parts = append(parts, "designations")
	}
	return "Changes(" + strings.Join(parts, ", ") + ")"
}

// Changes returns the kinds of change when next replaces s:
//   - Structure: StructureUnchanged when the systems are Equal; otherwise by
//     system.System.NamesSubsetOf in both directions;
//   - Sound: the temperaments differ (temperaments with different numbers
//     of classes always differ), the root numbers differ as numbers, or the
//     references differ;
//   - Designations: the notations, the numberings or the catalogs differ.
//
// It fails with ErrUnset when s or next is unset.
func (s *Settings) Changes(next *Settings) (Changes, error) {
	if err := checkPair("changes", s, next); err != nil {
		return Changes{}, err
	}
	pt, nt := s.tuning, next.tuning
	return Changes{
		Structure: structureChange(pt.System(), nt.System()),
		Sound: !pt.Temperament().Equal(nt.Temperament()) ||
			pt.RootClass().Number() != nt.RootClass().Number() ||
			!pt.Reference().Equal(nt.Reference()),
		Designations: !s.notation.Equal(next.notation) ||
			!s.numbering.Equal(next.numbering) ||
			!s.catalog.Equal(next.catalog),
	}, nil
}

// MustChanges is like Changes but panics with its error on failure.
func (s *Settings) MustChanges(next *Settings) Changes {
	return must(s.Changes(next))
}

// structureChange returns the kind of change from the valid system prev to
// the valid system next.
func structureChange(prev, next *system.System) StructureChange {
	if prev.Equal(next) {
		return StructureUnchanged
	}
	forward, _ := prev.NamesSubsetOf(next)
	backward, _ := next.NamesSubsetOf(prev)
	switch {
	case forward && backward:
		return StructureSameNames
	case forward:
		return StructureExtended
	case backward:
		return StructureNarrowed
	default:
		return StructureIncomparable
	}
}
