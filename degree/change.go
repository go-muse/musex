package degree

import "github.com/go-muse/musex/note"

// The changing methods change the receiver itself and are atomic: after an
// error its number, note, and note source remain unchanged. Their Must
// variants return the same pointer. A nil or unset receiver reports ErrUnset.

// SetNote replaces the note of d with an independent copy of n, preserving
// the number. It reports ErrUnset for an unset receiver or note.
func (d *Degree) SetNote(n note.Note) error {
	if err := d.check("SetNote"); err != nil {
		return err
	}

	if !n.IsValid() {
		return unsetError("SetNote", "unset note")
	}

	d.n = n

	return nil
}

// RemoveNote removes the note of d, preserving its number. Removing an absent
// note succeeds without changing d. An unset receiver reports ErrUnset.
func (d *Degree) RemoveNote() error {
	if err := d.check("RemoveNote"); err != nil {
		return err
	}

	d.n = note.Note{}

	return nil
}

// TransformNote calls t exactly once with an independent copy of the note of
// d and stores its returned note, preserving the number. d remains unchanged
// when t fails or returns an unset note. A panic from t is not recovered.
//
// Errors, in this order: ErrUnset for the receiver or a nil transform;
// ErrNoNote when d has no note; ErrTransformFailed with the error of t;
// ErrUnset for an unset returned note.
func (d *Degree) TransformNote(t Transform) error {
	if err := d.check("TransformNote"); err != nil {
		return err
	}

	if t == nil {
		return unsetError("TransformNote", "nil transform")
	}

	if !d.HasNote() {
		return noNoteError("TransformNote")
	}

	result, err := t(d.n)
	if err != nil {
		return transformError(err)
	}

	if !result.IsValid() {
		return unsetError("TransformNote", "unset note returned by the transform")
	}

	d.n = result

	return nil
}

// SetNumber changes only the number of d, preserving its note and note
// source. It reports ErrUnset for an unset receiver and ErrInvalidNumber for
// a number below one.
func (d *Degree) SetNumber(number int) error {
	if err := d.check("SetNumber"); err != nil {
		return err
	}

	if number < 1 {
		return invalidNumberError("SetNumber", "number is below one")
	}

	d.number = number

	return nil
}
