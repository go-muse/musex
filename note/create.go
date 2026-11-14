package note

import (
	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"
)

// newNote returns a note with the given parts, checking the name in the
// settings of src when src is not nil (ErrUnsetSettings, ErrInadmissible).
// A nil src gives an unbound note.
func newNote(op string, name spelling.Spelling, o octave.Octave, c cents.Cents, src settings.Source) (Note, error) {
	if src != nil {
		r := newReader(op)

		f, err := sourceFrame(r, src)
		if err != nil {
			return Note{}, err
		}

		if err := admit(op, f.sys, name); err != nil {
			return Note{}, err
		}
	}

	return Note{name: name, oct: o, c: c, src: src}, nil
}

// FromSpelling returns the note without an octave with name, correction c and
// source src. A nil src gives an unbound note; with a source the note is bound
// to it, and name must be admissible in the system of its settings. The zero
// cents.Cents is a correction of 0 cents.
//
// Errors, in this order: ErrUnset for an unset name; ErrUnsetSettings;
// ErrInadmissible.
func FromSpelling(name spelling.Spelling, c cents.Cents, src settings.Source) (Note, error) {
	const op = "FromSpelling"

	if !name.IsValid() {
		return Note{}, unsetError(op, "name")
	}

	return newNote(op, name, octave.Octave{}, c, src)
}

// MustFromSpelling is like FromSpelling but panics with the error of
// FromSpelling.
func MustFromSpelling(name spelling.Spelling, c cents.Cents, src settings.Source) Note {
	return must(FromSpelling(name, c, src))
}

// FromPitch returns the note with the written pitch p, correction c and
// source src, as FromSpelling.
//
// Errors, in this order: ErrUnset for an unset pitch; ErrUnsetSettings;
// ErrInadmissible.
func FromPitch(p pitch.Pitch, c cents.Cents, src settings.Source) (Note, error) {
	const op = "FromPitch"

	if !p.IsValid() {
		return Note{}, unsetError(op, "pitch")
	}

	name, o := p.Parts()

	return newNote(op, name, o, c, src)
}

// MustFromPitch is like FromPitch but panics with the error of FromPitch.
func MustFromPitch(p pitch.Pitch, c cents.Cents, src settings.Source) Note {
	return must(FromPitch(p, c, src))
}

// FromOctave returns the note with name in the Octave o, correction c and
// source src, as FromSpelling.
//
// Errors, in this order: ErrUnset for an unset name or octave;
// ErrUnsetSettings; ErrInadmissible.
func FromOctave(name spelling.Spelling, o octave.Octave, c cents.Cents, src settings.Source) (Note, error) {
	const op = "FromOctave"

	if !name.IsValid() {
		return Note{}, unsetError(op, "name")
	}

	if !o.IsValid() {
		return Note{}, unsetError(op, "octave")
	}

	return newNote(op, name, o, c, src)
}

// MustFromOctave is like FromOctave but panics with the error of FromOctave.
func MustFromOctave(name spelling.Spelling, o octave.Octave, c cents.Cents, src settings.Source) Note {
	return must(FromOctave(name, o, c, src))
}

// FromASPN returns the note with name in the Octave with ASPN number aspn,
// correction c and source src, as FromSpelling. Every int is an ASPN number,
// so there is no error out of range.
//
// Errors, in this order: ErrUnset for an unset name; ErrUnsetSettings;
// ErrInadmissible.
func FromASPN(name spelling.Spelling, aspn int, c cents.Cents, src settings.Source) (Note, error) {
	const op = "FromASPN"

	if !name.IsValid() {
		return Note{}, unsetError(op, "name")
	}

	return newNote(op, name, octave.FromASPN(aspn), c, src)
}

// MustFromASPN is like FromASPN but panics with the error of FromASPN.
func MustFromASPN(name spelling.Spelling, aspn int, c cents.Cents, src settings.Source) Note {
	return must(FromASPN(name, aspn, c, src))
}

// fromWritten returns the note bound to src with name written in the written
// octave given by written: Octave = w + k, the carry of the name in the
// system of the settings (system.System.FromWritten).
func fromWritten(op string, src settings.Source, name spelling.Spelling, c cents.Cents,
	written func(set *settings.Settings) (octave.Octave, error),
) (Note, error) {
	r := newReader(op)

	f, err := sourceFrame(r, src)
	if err != nil {
		return Note{}, err
	}

	if err := admit(op, f.sys, name); err != nil {
		return Note{}, err
	}

	w, err := written(f.set)
	if err != nil {
		return Note{}, depError(op, err)
	}

	p, err := f.sys.FromWritten(name, w)
	if err != nil {
		return Note{}, depError(op, err)
	}

	return Note{c: c, src: src}.withPitch(p), nil
}

// FromWritten returns the note bound to src with name written in the written
// octave w: its Octave is w plus the carry of the name in the system of the
// settings of src.
//
// Errors, in this order: ErrUnset for a nil src, an unset name or an unset
// w; ErrUnsetSettings; ErrInadmissible; ErrOutOfRange when the Octave does not
// fit in an int (also system.ErrOutOfRange).
func FromWritten(src settings.Source, name spelling.Spelling, w octave.Octave, c cents.Cents) (Note, error) {
	const op = "FromWritten"

	if err := checkSource(op, src); err != nil {
		return Note{}, err
	}

	if !name.IsValid() {
		return Note{}, unsetError(op, "name")
	}

	if !w.IsValid() {
		return Note{}, unsetError(op, "written octave")
	}

	return fromWritten(op, src, name, c, func(*settings.Settings) (octave.Octave, error) { return w, nil })
}

// MustFromWritten is like FromWritten but panics with the error of
// FromWritten.
func MustFromWritten(src settings.Source, name spelling.Spelling, w octave.Octave, c cents.Cents) Note {
	return must(FromWritten(src, name, w, c))
}

// FromWrittenASPN is like FromWritten with the written octave given by its
// ASPN number.
func FromWrittenASPN(src settings.Source, name spelling.Spelling, aspn int, c cents.Cents) (Note, error) {
	const op = "FromWrittenASPN"

	if err := checkSource(op, src); err != nil {
		return Note{}, err
	}

	if !name.IsValid() {
		return Note{}, unsetError(op, "name")
	}

	w := octave.FromASPN(aspn)

	return fromWritten(op, src, name, c, func(*settings.Settings) (octave.Octave, error) { return w, nil })
}

// MustFromWrittenASPN is like FromWrittenASPN but panics with the error of
// FromWrittenASPN.
func MustFromWrittenASPN(src settings.Source, name spelling.Spelling, aspn int, c cents.Cents) Note {
	return must(FromWrittenASPN(src, name, aspn, c))
}

// FromWrittenNumber is like FromWritten with the written octave given by its
// visible number num in the numbering of the settings of src.
//
// Errors, in this order: ErrUnset for a nil src or an unset name;
// ErrUnsetSettings; ErrInadmissible; for num: number.ErrSkipped (without a
// sentinel of note) or ErrOutOfRange (also octave.ErrOutOfRange and
// number.ErrOutOfRange); ErrOutOfRange for the Octave (also
// system.ErrOutOfRange).
func FromWrittenNumber(src settings.Source, name spelling.Spelling, num int, c cents.Cents) (Note, error) {
	const op = "FromWrittenNumber"

	if err := checkSource(op, src); err != nil {
		return Note{}, err
	}

	if !name.IsValid() {
		return Note{}, unsetError(op, "name")
	}

	return fromWritten(op, src, name, c, func(set *settings.Settings) (octave.Octave, error) {
		return octave.FromNumber(num, set.Numbering())
	})
}

// MustFromWrittenNumber is like FromWrittenNumber but panics with the error
// of FromWrittenNumber.
func MustFromWrittenNumber(src settings.Source, name spelling.Spelling, num int, c cents.Cents) Note {
	return must(FromWrittenNumber(src, name, num, c))
}

// Parse parses text as a name in the notation of the settings of src
// (settings.Settings.Parse) and returns the note without an octave bound to
// src, with a correction of 0 cents.
//
// Errors, in this order: ErrUnset for a nil src; ErrUnsetSettings; the errors
// of settings in the order of its contract, without a sentinel of note, except
// that an inadmissible name also matches ErrInadmissible.
func Parse(src settings.Source, text string) (Note, error) {
	const op = "Parse"

	if err := checkSource(op, src); err != nil {
		return Note{}, err
	}

	set, err := newReader(op).read(src)
	if err != nil {
		return Note{}, err
	}

	name, err := set.Parse(text)
	if err != nil {
		return Note{}, parseError(op, err)
	}

	return Note{name: name, src: src}, nil
}

// MustParse is like Parse but panics with the error of Parse.
func MustParse(src settings.Source, text string) Note {
	return must(Parse(src, text))
}

// ParsePitch parses text as a written pitch by the notation, the system and
// the numbering of the settings of src (settings.Settings.ParsePitch) and
// returns the note with an octave bound to src, with a correction of 0 cents.
// The errors are those of Parse.
func ParsePitch(src settings.Source, text string) (Note, error) {
	const op = "ParsePitch"

	if err := checkSource(op, src); err != nil {
		return Note{}, err
	}

	set, err := newReader(op).read(src)
	if err != nil {
		return Note{}, err
	}

	p, err := set.ParsePitch(text)
	if err != nil {
		return Note{}, parseError(op, err)
	}

	return Note{src: src}.withPitch(p), nil
}

// MustParsePitch is like ParsePitch but panics with the error of ParsePitch.
func MustParsePitch(src settings.Source, text string) Note {
	return must(ParsePitch(src, text))
}

// fromStep returns the note at step k of the tuning t, named by rule, with
// correction c and source src.
func fromStep(op string, t *tuning.Tuning, k system.Step, rule system.SpellingRule, c cents.Cents,
	src settings.Source,
) (Note, error) {
	p, err := t.PitchOfStep(k, rule)
	if err != nil {
		return Note{}, depError(op, err)
	}

	return Note{c: c, src: src}.withPitch(p), nil
}

// FromStep returns the note at step k from the reference of the tuning of
// the settings of src (tuning.Tuning.PitchOfStep), named by rule, with
// correction c, bound to src.
//
// Errors, in this order: ErrUnset for a nil src or an unset rule;
// ErrUnsetSettings; the errors of tuning: tuning.ErrSpelling without a
// sentinel of note, ErrOutOfRange (also tuning.ErrOutOfRange).
func FromStep(src settings.Source, k system.Step, rule system.SpellingRule, c cents.Cents) (Note, error) {
	const op = "FromStep"

	if err := checkSource(op, src); err != nil {
		return Note{}, err
	}

	if !rule.IsValid() {
		return Note{}, unsetError(op, "spelling rule")
	}

	set, err := newReader(op).read(src)
	if err != nil {
		return Note{}, err
	}

	return fromStep(op, set.Tuning(), k, rule, c, src)
}

// MustFromStep is like FromStep but panics with the error of FromStep.
func MustFromStep(src settings.Source, k system.Step, rule system.SpellingRule, c cents.Cents) Note {
	return must(FromStep(src, k, rule, c))
}

// FromStepInTuning is like FromStep in the explicit tuning t; the note is
// unbound. A nil or unset t reports ErrUnset.
func FromStepInTuning(t *tuning.Tuning, k system.Step, rule system.SpellingRule, c cents.Cents) (Note, error) {
	const op = "FromStepInTuning"

	if _, err := tuningFrame(op, t); err != nil {
		return Note{}, err
	}

	if !rule.IsValid() {
		return Note{}, unsetError(op, "spelling rule")
	}

	return fromStep(op, t, k, rule, c, nil)
}

// MustFromStepInTuning is like FromStepInTuning but panics with the error of
// FromStepInTuning.
func MustFromStepInTuning(t *tuning.Tuning, k system.Step, rule system.SpellingRule, c cents.Cents) Note {
	return must(FromStepInTuning(t, k, rule, c))
}

// nearest returns the nearest note to freq in t, named by rule, with the
// deviation of freq as its correction.
func nearest(op string, t *tuning.Tuning, freq float64, mode rounding.Mode, rule system.SpellingRule,
	src settings.Source,
) (Note, error) {
	p, delta, err := t.NearestPitchBy(freq, mode, rule)
	if err != nil {
		return Note{}, depError(op, err)
	}

	return Note{c: delta, src: src}.withPitch(p), nil
}

// Nearest returns the note nearest to the frequency freq in hertz in the
// tuning of the settings of src, a tuner: the step is chosen with
// tuning.NearestMode, the name by rule, and the correction is the deviation of
// freq from the written pitch (tuning.Tuning.NearestPitch). The note is bound
// to src.
//
// Errors, in this order: ErrUnset for a nil src or an unset rule;
// ErrUnsetSettings; the errors of tuning in the order of its contract:
// tuning.ErrInvalidArgument for an invalid frequency, ErrOutOfRange (also
// tuning.ErrTooComplex), tuning.ErrSpelling, ErrOutOfRange (also
// tuning.ErrOutOfRange).
func Nearest(src settings.Source, freq float64, rule system.SpellingRule) (Note, error) {
	return nearestIn("Nearest", src, freq, tuning.NearestMode, rule)
}

// MustNearest is like Nearest but panics with the error of Nearest.
func MustNearest(src settings.Source, freq float64, rule system.SpellingRule) Note {
	return must(Nearest(src, freq, rule))
}

// NearestBy is like Nearest with the rounding mode mode for the choice of the
// neighbor step. An unset mode reports ErrUnset; a mode outside the list of
// rounding reports tuning.ErrInvalidArgument.
func NearestBy(src settings.Source, freq float64, mode rounding.Mode, rule system.SpellingRule) (Note, error) {
	const op = "NearestBy"

	if err := checkSource(op, src); err != nil {
		return Note{}, err
	}

	if mode == 0 {
		return Note{}, unsetError(op, "rounding mode")
	}

	return nearestIn(op, src, freq, mode, rule)
}

// MustNearestBy is like NearestBy but panics with the error of NearestBy.
func MustNearestBy(src settings.Source, freq float64, mode rounding.Mode, rule system.SpellingRule) Note {
	return must(NearestBy(src, freq, mode, rule))
}

// nearestIn is the common part of Nearest and NearestBy.
func nearestIn(op string, src settings.Source, freq float64, mode rounding.Mode, rule system.SpellingRule) (Note, error) {
	if err := checkSource(op, src); err != nil {
		return Note{}, err
	}

	if !rule.IsValid() {
		return Note{}, unsetError(op, "spelling rule")
	}

	set, err := newReader(op).read(src)
	if err != nil {
		return Note{}, err
	}

	return nearest(op, set.Tuning(), freq, mode, rule, src)
}

// NearestInTuning is like Nearest in the explicit tuning t; the note is
// unbound. A nil or unset t reports ErrUnset.
func NearestInTuning(t *tuning.Tuning, freq float64, rule system.SpellingRule) (Note, error) {
	return nearestInTuning("NearestInTuning", t, freq, tuning.NearestMode, rule)
}

// MustNearestInTuning is like NearestInTuning but panics with the error of
// NearestInTuning.
func MustNearestInTuning(t *tuning.Tuning, freq float64, rule system.SpellingRule) Note {
	return must(NearestInTuning(t, freq, rule))
}

// NearestByInTuning is like NearestBy in the explicit tuning t; the note is
// unbound. A nil or unset t reports ErrUnset.
func NearestByInTuning(t *tuning.Tuning, freq float64, mode rounding.Mode, rule system.SpellingRule) (Note, error) {
	const op = "NearestByInTuning"

	if _, err := tuningFrame(op, t); err != nil {
		return Note{}, err
	}

	if mode == 0 {
		return Note{}, unsetError(op, "rounding mode")
	}

	return nearestInTuning(op, t, freq, mode, rule)
}

// MustNearestByInTuning is like NearestByInTuning but panics with the error
// of NearestByInTuning.
func MustNearestByInTuning(t *tuning.Tuning, freq float64, mode rounding.Mode, rule system.SpellingRule) Note {
	return must(NearestByInTuning(t, freq, mode, rule))
}

// nearestInTuning is the common part of NearestInTuning and
// NearestByInTuning.
func nearestInTuning(op string, t *tuning.Tuning, freq float64, mode rounding.Mode, rule system.SpellingRule) (Note, error) {
	if _, err := tuningFrame(op, t); err != nil {
		return Note{}, err
	}

	if !rule.IsValid() {
		return Note{}, unsetError(op, "spelling rule")
	}

	return nearest(op, t, freq, mode, rule, nil)
}
