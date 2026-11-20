package voice

import (
	"fmt"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/system"
)

// builtRole is the name and the roles of a built line.
type builtRole int

const (
	keepRoles builtRole = iota // the name and the roles of the line
	asDoubling
	asImitating
	asInversion
)

// meta returns a line with the name and the roles of a line built from l.
func (l *Line) meta(role builtRole) *Line {
	switch role {
	case asDoubling:
		return &Line{thematic: Doubling}
	case asImitating:
		return &Line{thematic: Imitating}
	case asInversion:
		return &Line{thematic: Inversion}
	default:
		return l
	}
}

// mapSounds returns a copy of the elements of l with the note of each sound
// replaced by f; an error of f is wrapped in an *ItemError with the index of
// the sound.
func (l *Line) mapSounds(f func(n note.Note) (note.Note, error)) ([]Element, error) {
	elems := make([]Element, len(l.elems))

	for i, e := range l.elems {
		if e.kind == KindSound {
			n, err := f(e.note)
			if err != nil {
				return nil, itemError(i, err)
			}

			e.note = n
		}

		elems[i] = e
	}

	return elems, nil
}

// transposed checks the arguments of a transposition and returns the
// elements of l with every sound moved by i.
func (l *Line) transposed(sys *system.System, explicit bool, i interval.Interval) ([]Element, error) {
	if !l.IsSet() {
		return nil, errUnsetLine()
	}

	if explicit {
		if err := checkSystem(sys); err != nil {
			return nil, err
		}
	}

	if !i.IsValid() {
		return nil, unsetError("interval", interval.ErrUnset)
	}

	return l.transposeSounds(sys, explicit, i)
}

func (l *Line) transposeSounds(sys *system.System, explicit bool, i interval.Interval) ([]Element, error) {
	if !explicit {
		if err := l.checkBound(l.soundIndexes()); err != nil {
			return nil, err
		}
	}

	return l.mapSounds(func(n note.Note) (note.Note, error) {
		var (
			moved note.Note
			err   error
		)

		if explicit {
			moved, err = interval.TransposeNoteIn(sys, n, i)
		} else {
			moved, err = interval.TransposeNote(n, i)
		}

		if err != nil {
			return note.Note{}, fmt.Errorf("voice: transpose %s by %s: %w", n, i, err)
		}

		return moved, nil
	})
}

// build returns the line of the elements of l transposed by i with the name
// and the roles given by role.
func (l *Line) build(sys *system.System, explicit bool, i interval.Interval, role builtRole) (*Line, error) {
	elems, err := l.transposed(sys, explicit, i)
	if err != nil {
		return nil, err
	}

	out := l.withSameRhythm(elems)
	out.setMeta(l.meta(role))

	return out, nil
}

// setMeta copies the name and the roles of meta to l.
func (l *Line) setMeta(meta *Line) {
	l.name = meta.name
	l.placement = meta.placement
	l.thematic = meta.thematic
	l.importance = meta.importance
}

// Transpose returns the line with every sound moved by the interval i:
// a new note with the correction and the source of the original one
// (interval.TransposeNote), in the system of the settings of that sound.
// Sounds of different systems are allowed. Rests, durations, fermatas, the
// name and the roles are kept.
//
// Errors, in this order: ErrUnset for an unset line, then for an unset i
// (with interval.ErrUnset); ErrNoSystem for the first unbound sound; then for
// each sound the errors of packages interval, note and system, such as
// note.ErrInadmissible with system.ErrInadmissible for a name of the result
// that is not admissible. The errors of the sounds are in an *ItemError with
// the index of the sound.
func (l *Line) Transpose(i interval.Interval) (*Line, error) {
	return l.build(nil, false, i, keepRoles)
}

// MustTranspose is like Transpose but panics with its error.
func (l *Line) MustTranspose(i interval.Interval) *Line {
	return must(l.Transpose(i))
}

// TransposeIn is like Transpose in the explicit system sys for every sound
// (interval.TransposeNoteIn): unbound sounds are allowed, and a bound sound
// also checks the new name in its own set. A nil or unset sys reports
// ErrUnset (with system.ErrInvalid) after the check of the line.
func (l *Line) TransposeIn(sys *system.System, i interval.Interval) (*Line, error) {
	return l.build(sys, true, i, keepRoles)
}

// MustTransposeIn is like TransposeIn but panics with its error.
func (l *Line) MustTransposeIn(sys *system.System, i interval.Interval) *Line {
	return must(l.TransposeIn(sys, i))
}

// Double returns a doubling of the line at the interval i: the elements of
// Transpose, the name "", no roles by place and by importance, and the
// thematic role Doubling. The errors are those of Transpose.
func (l *Line) Double(i interval.Interval) (*Line, error) {
	return l.build(nil, false, i, asDoubling)
}

// MustDouble is like Double but panics with its error.
func (l *Line) MustDouble(i interval.Interval) *Line {
	return must(l.Double(i))
}

// DoubleIn is like Double in the explicit system sys.
func (l *Line) DoubleIn(sys *system.System, i interval.Interval) (*Line, error) {
	return l.build(sys, true, i, asDoubling)
}

// MustDoubleIn is like DoubleIn but panics with its error.
func (l *Line) MustDoubleIn(sys *system.System, i interval.Interval) *Line {
	return must(l.DoubleIn(sys, i))
}

// Imitate returns an exact imitation of the line at the interval i without a
// rest before its entry: the elements of Transpose, the name "", no roles by
// place and by importance, and the thematic role Imitating. The errors are
// those of Transpose.
func (l *Line) Imitate(i interval.Interval) (*Line, error) {
	return l.build(nil, false, i, asImitating)
}

// MustImitate is like Imitate but panics with its error.
func (l *Line) MustImitate(i interval.Interval) *Line {
	return must(l.Imitate(i))
}

// ImitateIn is like Imitate in the explicit system sys.
func (l *Line) ImitateIn(sys *system.System, i interval.Interval) (*Line, error) {
	return l.build(sys, true, i, asImitating)
}

// MustImitateIn is like ImitateIn but panics with its error.
func (l *Line) MustImitateIn(sys *system.System, i interval.Interval) *Line {
	return must(l.ImitateIn(sys, i))
}

// ImitateAfter is like Imitate with a rest of the duration shift without a
// fermata before the entry: the length is the length of the line plus the
// shift.
//
// Errors, in this order: ErrUnset for an unset line, then for an unset i and
// an unset shift; ErrTiedRest for a tied shift; the errors of Transpose; a
// start out of the range of package common/fraction
// (duration.ErrOutOfRange with fraction.ErrOutOfRange), without an
// *ItemError.
func (l *Line) ImitateAfter(i interval.Interval, shift Duration) (*Line, error) {
	return l.imitateAfter(nil, false, i, shift)
}

// MustImitateAfter is like ImitateAfter but panics with its error.
func (l *Line) MustImitateAfter(i interval.Interval, shift Duration) *Line {
	return must(l.ImitateAfter(i, shift))
}

// ImitateAfterIn is like ImitateAfter in the explicit system sys.
func (l *Line) ImitateAfterIn(sys *system.System, i interval.Interval, shift Duration) (*Line, error) {
	return l.imitateAfter(sys, true, i, shift)
}

// MustImitateAfterIn is like ImitateAfterIn but panics with its error.
func (l *Line) MustImitateAfterIn(sys *system.System, i interval.Interval, shift Duration) *Line {
	return must(l.ImitateAfterIn(sys, i, shift))
}

func (l *Line) imitateAfter(sys *system.System, explicit bool, i interval.Interval, shift Duration) (*Line, error) {
	if !l.IsSet() {
		return nil, errUnsetLine()
	}

	if explicit {
		if err := checkSystem(sys); err != nil {
			return nil, err
		}
	}

	if !i.IsValid() {
		return nil, unsetError("interval", interval.ErrUnset)
	}

	rest, err := NewRest(shift)
	if err != nil {
		return nil, err
	}

	elems, err := l.transposeSounds(sys, explicit, i)
	if err != nil {
		return nil, err
	}

	return newLine(append([]Element{rest}, elems...), l.meta(asImitating))
}

// Invert returns an exact inversion of the line around the written pitch of
// the axis note: every sound x is replaced by y = interval.ReflectPitch(sys,
// a, x), the axis moved by the opposite of the interval from the axis to x,
// in the system of the settings of that sound. The new note is a copy of x
// with the written pitch y (note.Note.ReplaceWithPitch): its correction and
// source are kept. Durations, rests and fermatas are kept; the result has the
// name "", no roles by place and by importance, and the thematic role
// Inversion. The source of the axis is not read.
//
// Errors, in this order: ErrUnset for an unset line, then for an unset axis
// (with note.ErrUnset); ErrNoOctave for an axis without an octave;
// ErrNoSystem for the first unbound sound; then for each sound, in an
// *ItemError with its index: the errors of reading its settings,
// system.ErrInadmissible when the name of the axis or of the sound is not
// admissible in its system, the other errors of interval.ReflectPitch, and
// note.ErrInadmissible when the new name is not admissible in the set of the
// note.
func (l *Line) Invert(axis note.Note) (*Line, error) {
	return l.invertNote(nil, false, axis)
}

// MustInvert is like Invert but panics with its error.
func (l *Line) MustInvert(axis note.Note) *Line {
	return must(l.Invert(axis))
}

// InvertIn is like Invert in the explicit system sys for every sound.
// A name of the axis that is not admissible in sys reports
// system.ErrInadmissible without an *ItemError, even for a line without
// sounds. A nil or unset sys reports ErrUnset (with system.ErrInvalid) after
// the check of the line.
func (l *Line) InvertIn(sys *system.System, axis note.Note) (*Line, error) {
	return l.invertNote(sys, true, axis)
}

// MustInvertIn is like InvertIn but panics with its error.
func (l *Line) MustInvertIn(sys *system.System, axis note.Note) *Line {
	return must(l.InvertIn(sys, axis))
}

// InvertPitch is like Invert with the axis given as a written pitch. An unset
// axis reports ErrUnset with pitch.ErrInvalid.
func (l *Line) InvertPitch(axis pitch.Pitch) (*Line, error) {
	return l.invertPitch(nil, false, axis)
}

// MustInvertPitch is like InvertPitch but panics with its error.
func (l *Line) MustInvertPitch(axis pitch.Pitch) *Line {
	return must(l.InvertPitch(axis))
}

// InvertPitchIn is like InvertIn with the axis given as a written pitch.
func (l *Line) InvertPitchIn(sys *system.System, axis pitch.Pitch) (*Line, error) {
	return l.invertPitch(sys, true, axis)
}

// MustInvertPitchIn is like InvertPitchIn but panics with its error.
func (l *Line) MustInvertPitchIn(sys *system.System, axis pitch.Pitch) *Line {
	return must(l.InvertPitchIn(sys, axis))
}

func (l *Line) invertNote(sys *system.System, explicit bool, axis note.Note) (*Line, error) {
	if !l.IsSet() {
		return nil, errUnsetLine()
	}

	if explicit {
		if err := checkSystem(sys); err != nil {
			return nil, err
		}
	}

	if !axis.IsValid() {
		return nil, unsetError("axis", note.ErrUnset)
	}

	if !axis.HasOctave() {
		return nil, fmt.Errorf("%w: axis %s", ErrNoOctave, axis)
	}

	a, err := axis.Pitch()
	if err != nil {
		return nil, fmt.Errorf("voice: axis: %w", err)
	}

	return l.invert(sys, explicit, a)
}

func (l *Line) invertPitch(sys *system.System, explicit bool, axis pitch.Pitch) (*Line, error) {
	if !l.IsSet() {
		return nil, errUnsetLine()
	}

	if explicit {
		if err := checkSystem(sys); err != nil {
			return nil, err
		}
	}

	if !axis.IsValid() {
		return nil, unsetError("axis", pitch.ErrInvalid)
	}

	return l.invert(sys, explicit, axis)
}

// invert reflects every sound around the checked axis a.
func (l *Line) invert(sys *system.System, explicit bool, a pitch.Pitch) (*Line, error) {
	if explicit {
		if err := checkAdmits(sys, a, "axis"); err != nil {
			return nil, err
		}
	} else if err := l.checkBound(l.soundIndexes()); err != nil {
		return nil, err
	}

	elems, err := l.mapSounds(func(n note.Note) (note.Note, error) {
		s := sys
		if !explicit {
			var err error
			if s, err = noteSystem(n); err != nil {
				return note.Note{}, err
			}
		}

		x, err := n.Pitch()
		if err != nil {
			return note.Note{}, fmt.Errorf("voice: pitch of %s: %w", n, err)
		}

		y, err := interval.ReflectPitch(s, a, x)
		if err != nil {
			return note.Note{}, fmt.Errorf("voice: reflect %s around %s: %w", x, a, err)
		}

		if err := n.ReplaceWithPitch(y); err != nil {
			return note.Note{}, fmt.Errorf("voice: inverted %s: %w", y, err)
		}

		return n, nil
	})
	if err != nil {
		return nil, err
	}

	out := l.withSameRhythm(elems)
	out.setMeta(l.meta(asInversion))

	return out, nil
}

// Retrograde returns a retrograde of the line of the kind: RetrogradeBoth
// puts the elements in reverse order; RetrogradePitches keeps the rhythm
// (the kind, the duration and the fermata of each element) and puts the notes
// of the sounds into the sounds in reverse order; RetrogradeRhythm reverses
// the rhythm and puts the notes of the sounds into the sounds in their
// original order. A tied duration is reversed as a whole. The name and the
// roles are kept.
//
// Errors, in this order: ErrUnset for an unset line, then for RetrogradeUnset
// or a kind outside the list; for RetrogradeBoth and RetrogradeRhythm a start
// out of the range of package common/fraction (duration.ErrOutOfRange with
// fraction.ErrOutOfRange).
func (l *Line) Retrograde(kind RetrogradeKind) (*Line, error) {
	if !l.IsSet() {
		return nil, errUnsetLine()
	}

	if !kind.IsValid() {
		return nil, unsetError("retrograde kind "+kind.String(), nil)
	}

	n := len(l.elems)
	reversed := make([]Element, n)

	for i, e := range l.elems {
		reversed[n-1-i] = e
	}

	switch kind {
	case RetrogradeBoth:
		return newLine(reversed, l)
	case RetrogradePitches:
		return l.withSameRhythm(withNotesOf(l.elems, reversed)), nil
	default:
		return newLine(withNotesOf(reversed, l.elems), l)
	}
}

// MustRetrograde is like Retrograde but panics with its error.
func (l *Line) MustRetrograde(kind RetrogradeKind) *Line {
	return must(l.Retrograde(kind))
}

// withNotesOf returns a copy of rhythm whose sounds get the notes of the
// sounds of source in order; both have the same number of sounds.
func withNotesOf(rhythm, source []Element) []Element {
	notes := make([]note.Note, 0, len(source))

	for _, e := range source {
		if e.kind == KindSound {
			notes = append(notes, e.note)
		}
	}

	out := make([]Element, len(rhythm))
	k := 0

	for i, e := range rhythm {
		if e.kind == KindSound {
			e.note = notes[k]
			k++
		}

		out[i] = e
	}

	return out
}
