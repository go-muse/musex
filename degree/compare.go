package degree

// Equal reports whether both degrees are set, have the same number, and
// either both have no note or have notes equal by note.Note.Equal. Sources of
// settings do not take part. The unset degree equals no degree, not even
// another unset one.
func (d Degree) Equal(other Degree) bool {
	if !d.IsValid() || !other.IsValid() || d.number != other.number {
		return false
	}

	if d.HasNote() != other.HasNote() {
		return false
	}

	return !d.HasNote() || d.n.Equal(other.n)
}

// SameNumber reports whether both degrees are set and have the same number.
// Notes do not take part.
func (d Degree) SameNumber(other Degree) bool {
	return d.IsValid() && other.IsValid() && d.number == other.number
}

// SameNote reports whether both degrees have notes equal by note.Note.Equal.
// Two degrees without notes do not have the same note.
func (d Degree) SameNote(other Degree) bool {
	return d.HasNote() && other.HasNote() && d.n.Equal(other.n)
}

// Compare returns -1, 0, or +1 by an agreed order: the unset degree first,
// then by number, a degree without a note before one with a note, and finally
// by note.Note.Compare. Two unset degrees compare as 0 without being equal.
// For set degrees, Compare returns 0 exactly when Equal is true.
func (d Degree) Compare(other Degree) int {
	switch {
	case !d.IsValid() && !other.IsValid():
		return 0
	case !d.IsValid():
		return -1
	case !other.IsValid():
		return 1
	case d.number < other.number:
		return -1
	case d.number > other.number:
		return 1
	case !d.HasNote() && other.HasNote():
		return -1
	case d.HasNote() && !other.HasNote():
		return 1
	case d.HasNote():
		return d.n.Compare(other.n)
	default:
		return 0
	}
}
