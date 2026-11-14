package term

// Group is the tempo group of a term: slow, moderate, or fast tempos, or no
// group. Its zero value is NoGroup, a valid value that means that the term
// belongs to no group; it is not an unset state. Values greater than Fast are
// not groups: IsValid reports false for them.
type Group uint8

const (
	// NoGroup means that the term has no group. It is a valid value.
	NoGroup Group = iota
	// Slow is the group of slow tempos.
	Slow
	// Moderate is the group of moderate tempos.
	Moderate
	// Fast is the group of fast tempos.
	Fast
)

// groupNames holds the strings of the valid groups, indexed by group.
var groupNames = [...]string{"none", "slow", "moderate", "fast"}

// Groups returns the three groups Slow, Moderate, and Fast in this order, as a
// new slice that belongs to the caller. NoGroup is not listed.
func Groups() []Group {
	return []Group{Slow, Moderate, Fast}
}

// IsValid reports whether g is NoGroup, Slow, Moderate, or Fast.
func (g Group) IsValid() bool {
	return g <= Fast
}

// Equal reports whether g and other are the same value; it matches ==.
func (g Group) Equal(other Group) bool {
	return g == other
}

// String returns "none", "slow", "moderate", or "fast", and "Group(invalid)"
// for a value that is not a group.
func (g Group) String() string {
	if !g.IsValid() {
		return "Group(invalid)"
	}
	return groupNames[g]
}
