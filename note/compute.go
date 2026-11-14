package note

import (
	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"
)

// The computations by settings read the source of the note, or use an
// explicit source (the In forms) or an explicit tuning (the InTuning forms),
// at the moment of the call. They never change the note and its binding.
// An unbound note needs an explicit context: without it they report
// ErrNoSource.

// opContext is the context of an operation: the own settings of the notes, an
// explicit source or an explicit tuning.
type opContext struct {
	src     settings.Source
	t       *tuning.Tuning
	in      bool // an explicit source
	inTuned bool // an explicit tuning
}

// ownContext is the context of the own settings of the notes.
func ownContext() opContext { return opContext{} }

// sourceContext is the context of an explicit source.
func sourceContext(src settings.Source) opContext { return opContext{src: src, in: true} }

// tuningContext is the context of an explicit tuning.
func tuningContext(t *tuning.Tuning) opContext { return opContext{t: t, inTuned: true} }

// check reports ErrUnset for a nil explicit source or an unset explicit
// tuning.
func (c opContext) check(op string) error {
	switch {
	case c.in:
		return checkSource(op, c.src)
	case c.inTuned:
		_, err := tuningFrame(op, c.t)

		return err
	default:
		return nil
	}
}

// frame returns the frame of the note n in the context: its own settings
// (ErrNoSource, ErrUnsetSettings), the settings of the explicit source
// (ErrUnsetSettings) or the explicit tuning.
func (c opContext) frame(r *reader, n Note) (frame, error) {
	switch {
	case c.in:
		return sourceFrame(r, c.src)
	case c.inTuned:
		return frame{sys: c.t.System(), t: c.t}, nil
	default:
		return ownFrame(r, n)
	}
}

// prepare checks n and the context for a computation over one note and
// returns its frame: ErrUnset, ErrNoOctave when withOctave, ErrNoSource,
// ErrUnsetSettings, then ErrInadmissible when admissible.
func (n Note) prepare(op string, c opContext, withOctave, admissible bool) (frame, error) {
	if err := checkNote(op, n); err != nil {
		return frame{}, err
	}

	if err := c.check(op); err != nil {
		return frame{}, err
	}

	if withOctave {
		if err := checkOctave(op, n); err != nil {
			return frame{}, err
		}
	}

	f, err := c.frame(newReader(op), n)
	if err != nil {
		return frame{}, err
	}

	if admissible {
		if err := admit(op, f.sys, n.name); err != nil {
			return frame{}, err
		}
	}

	return f, nil
}

// Settings returns the current settings of the source of n: a snapshot read
// at the call; the settings are shared and immutable.
//
// Errors, in this order: ErrUnset for the note; ErrNoSource;
// ErrUnsetSettings.
func (n Note) Settings() (*settings.Settings, error) {
	f, err := n.prepare("Settings", ownContext(), false, false)
	if err != nil {
		return nil, err
	}

	return f.set, nil
}

// MustSettings is like Settings but panics with its error.
func (n Note) MustSettings() *settings.Settings {
	return must(n.Settings())
}

func (n Note) isAdmissible(op string, c opContext) (bool, error) {
	f, err := n.prepare(op, c, false, false)
	if err != nil {
		return false, err
	}

	ok, err := f.sys.Admits(n.name)
	if err != nil {
		return false, depError(op, err)
	}

	return ok, nil
}

// IsAdmissible reports whether the name of n is admissible in the system of
// its settings (system.System.Admits). An inadmissible name is the answer
// false, not an error.
//
// Errors, in this order: ErrUnset for the note; ErrNoSource;
// ErrUnsetSettings.
func (n Note) IsAdmissible() (bool, error) {
	return n.isAdmissible("IsAdmissible", ownContext())
}

// MustIsAdmissible is like IsAdmissible but panics with its error.
func (n Note) MustIsAdmissible() bool {
	return must(n.IsAdmissible())
}

// IsAdmissibleIn is like IsAdmissible in the settings of the explicit source
// src. A nil src reports ErrUnset.
func (n Note) IsAdmissibleIn(src settings.Source) (bool, error) {
	return n.isAdmissible("IsAdmissibleIn", sourceContext(src))
}

// MustIsAdmissibleIn is like IsAdmissibleIn but panics with its error.
func (n Note) MustIsAdmissibleIn(src settings.Source) bool {
	return must(n.IsAdmissibleIn(src))
}

func (n Note) class(op string, c opContext) (pitchclass.Class, error) {
	f, err := n.prepare(op, c, false, true)
	if err != nil {
		return pitchclass.Class{}, err
	}

	class, err := f.sys.Class(n.name)
	if err != nil {
		return pitchclass.Class{}, depError(op, err)
	}

	return class, nil
}

// Class returns the class of the name of n in the system of its settings
// (system.System.Class), for a note with or without an octave.
//
// Errors, in this order: ErrUnset for the note; ErrNoSource;
// ErrUnsetSettings; ErrInadmissible.
func (n Note) Class() (pitchclass.Class, error) {
	return n.class("Class", ownContext())
}

// MustClass is like Class but panics with its error.
func (n Note) MustClass() pitchclass.Class {
	return must(n.Class())
}

// ClassIn is like Class in the settings of the explicit source src. A nil src
// reports ErrUnset.
func (n Note) ClassIn(src settings.Source) (pitchclass.Class, error) {
	return n.class("ClassIn", sourceContext(src))
}

// MustClassIn is like ClassIn but panics with its error.
func (n Note) MustClassIn(src settings.Source) pitchclass.Class {
	return must(n.ClassIn(src))
}

// ClassInTuning is like Class in the system of the explicit tuning t. A nil or
// unset t reports ErrUnset.
func (n Note) ClassInTuning(t *tuning.Tuning) (pitchclass.Class, error) {
	return n.class("ClassInTuning", tuningContext(t))
}

// MustClassInTuning is like ClassInTuning but panics with its error.
func (n Note) MustClassInTuning(t *tuning.Tuning) pitchclass.Class {
	return must(n.ClassInTuning(t))
}

// writtenOctave returns the frame and the written octave of n.
func (n Note) writtenOctave(op string, c opContext) (frame, octave.Octave, error) {
	f, err := n.prepare(op, c, true, true)
	if err != nil {
		return frame{}, octave.Octave{}, err
	}

	w, err := f.sys.WrittenOctave(n.pitch())
	if err != nil {
		return frame{}, octave.Octave{}, depError(op, err)
	}

	return f, w, nil
}

// WrittenOctave returns the written octave w of n in the system of its
// settings (system.System.WrittenOctave): B♯ in octave 4 is written in
// octave 3.
//
// Errors, in this order: ErrUnset for the note; ErrNoOctave; ErrNoSource;
// ErrUnsetSettings; ErrInadmissible; ErrOutOfRange (also
// system.ErrOutOfRange).
func (n Note) WrittenOctave() (octave.Octave, error) {
	_, w, err := n.writtenOctave("WrittenOctave", ownContext())

	return w, err
}

// MustWrittenOctave is like WrittenOctave but panics with its error.
func (n Note) MustWrittenOctave() octave.Octave {
	return must(n.WrittenOctave())
}

// WrittenOctaveIn is like WrittenOctave in the settings of the explicit source
// src. A nil src reports ErrUnset.
func (n Note) WrittenOctaveIn(src settings.Source) (octave.Octave, error) {
	_, w, err := n.writtenOctave("WrittenOctaveIn", sourceContext(src))

	return w, err
}

// MustWrittenOctaveIn is like WrittenOctaveIn but panics with its error.
func (n Note) MustWrittenOctaveIn(src settings.Source) octave.Octave {
	return must(n.WrittenOctaveIn(src))
}

// WrittenASPN returns the ASPN number of the written octave of n; see
// WrittenOctave.
func (n Note) WrittenASPN() (int, error) {
	_, w, err := n.writtenOctave("WrittenASPN", ownContext())
	if err != nil {
		return 0, err
	}

	return w.MustASPN(), nil
}

// MustWrittenASPN is like WrittenASPN but panics with its error.
func (n Note) MustWrittenASPN() int {
	return must(n.WrittenASPN())
}

func (n Note) writtenNumber(op string, c opContext) (int, error) {
	f, w, err := n.writtenOctave(op, c)
	if err != nil {
		return 0, err
	}

	num, err := w.Number(f.set.Numbering())
	if err != nil {
		return 0, depError(op, err)
	}

	return num, nil
}

// WrittenNumber returns the visible number of the written octave of n in the
// numbering of its settings (octave.Octave.Number); see WrittenOctave. A
// number that does not fit in an int reports ErrOutOfRange (also
// octave.ErrOutOfRange and number.ErrOutOfRange).
func (n Note) WrittenNumber() (int, error) {
	return n.writtenNumber("WrittenNumber", ownContext())
}

// MustWrittenNumber is like WrittenNumber but panics with its error.
func (n Note) MustWrittenNumber() int {
	return must(n.WrittenNumber())
}

// WrittenNumberIn is like WrittenNumber in the settings of the explicit
// source src. A nil src reports ErrUnset.
func (n Note) WrittenNumberIn(src settings.Source) (int, error) {
	return n.writtenNumber("WrittenNumberIn", sourceContext(src))
}

// MustWrittenNumberIn is like WrittenNumberIn but panics with its error.
func (n Note) MustWrittenNumberIn(src settings.Source) int {
	return must(n.WrittenNumberIn(src))
}

func (n Note) writtenName(op string, c opContext) (string, error) {
	f, w, err := n.writtenOctave(op, c)
	if err != nil {
		return "", err
	}

	return w.Name(f.set.Catalog()), nil
}

// WrittenName returns the name of the written octave of n in the catalog of
// its settings (octave.Octave.Name); an empty name is a valid answer. See
// WrittenOctave.
func (n Note) WrittenName() (string, error) {
	return n.writtenName("WrittenName", ownContext())
}

// MustWrittenName is like WrittenName but panics with its error.
func (n Note) MustWrittenName() string {
	return must(n.WrittenName())
}

// WrittenNameIn is like WrittenName in the settings of the explicit source
// src. A nil src reports ErrUnset.
func (n Note) WrittenNameIn(src settings.Source) (string, error) {
	return n.writtenName("WrittenNameIn", sourceContext(src))
}

// MustWrittenNameIn is like WrittenNameIn but panics with its error.
func (n Note) MustWrittenNameIn(src settings.Source) string {
	return must(n.WrittenNameIn(src))
}

func (n Note) step(op string, c opContext) (system.Step, error) {
	f, err := n.prepare(op, c, true, true)
	if err != nil {
		return 0, err
	}

	k, err := f.tuning().StepOf(n.pitch())
	if err != nil {
		return 0, depError(op, err)
	}

	return k, nil
}

// Step returns the number of the step of n from the reference of the tuning
// of its settings (tuning.Tuning.StepOf); the correction does not take part.
//
// Errors, in this order: ErrUnset for the note; ErrNoOctave; ErrNoSource;
// ErrUnsetSettings; ErrInadmissible; ErrOutOfRange (also the error of tuning).
func (n Note) Step() (system.Step, error) {
	return n.step("Step", ownContext())
}

// MustStep is like Step but panics with its error.
func (n Note) MustStep() system.Step {
	return must(n.Step())
}

// StepIn is like Step in the settings of the explicit source src. A nil src
// reports ErrUnset.
func (n Note) StepIn(src settings.Source) (system.Step, error) {
	return n.step("StepIn", sourceContext(src))
}

// MustStepIn is like StepIn but panics with its error.
func (n Note) MustStepIn(src settings.Source) system.Step {
	return must(n.StepIn(src))
}

// StepInTuning is like Step in the explicit tuning t. A nil or unset t
// reports ErrUnset.
func (n Note) StepInTuning(t *tuning.Tuning) (system.Step, error) {
	return n.step("StepInTuning", tuningContext(t))
}

// MustStepInTuning is like StepInTuning but panics with its error.
func (n Note) MustStepInTuning(t *tuning.Tuning) system.Step {
	return must(n.StepInTuning(t))
}

// height returns h(x) + c: the exact height of the note in cents from the
// reference of the tuning t.
func (n Note) height(op string, t *tuning.Tuning) (cents.Cents, error) {
	h, err := t.CentsFromReference(n.pitch())
	if err != nil {
		return cents.Cents{}, depError(op, err)
	}

	h, err = h.Add(n.c)
	if err != nil {
		return cents.Cents{}, depError(op, err)
	}

	return h, nil
}

func (n Note) centsFromReference(op string, c opContext) (cents.Cents, error) {
	f, err := n.prepare(op, c, true, true)
	if err != nil {
		return cents.Cents{}, err
	}

	return n.height(op, f.tuning())
}

// CentsFromReference returns the height of n in cents from the reference of
// the tuning of its settings, with the correction: h(x) + c, exactly
// (tuning.Tuning.CentsFromReference).
//
// Errors, in this order: ErrUnset for the note; ErrNoOctave; ErrNoSource;
// ErrUnsetSettings; ErrInadmissible; ErrOutOfRange (also the error of tuning
// or cents).
func (n Note) CentsFromReference() (cents.Cents, error) {
	return n.centsFromReference("CentsFromReference", ownContext())
}

// MustCentsFromReference is like CentsFromReference but panics with its
// error.
func (n Note) MustCentsFromReference() cents.Cents {
	return must(n.CentsFromReference())
}

// CentsFromReferenceIn is like CentsFromReference in the settings of the
// explicit source src. A nil src reports ErrUnset.
func (n Note) CentsFromReferenceIn(src settings.Source) (cents.Cents, error) {
	return n.centsFromReference("CentsFromReferenceIn", sourceContext(src))
}

// MustCentsFromReferenceIn is like CentsFromReferenceIn but panics with its
// error.
func (n Note) MustCentsFromReferenceIn(src settings.Source) cents.Cents {
	return must(n.CentsFromReferenceIn(src))
}

// CentsFromReferenceInTuning is like CentsFromReference in the explicit tuning
// t. A nil or unset t reports ErrUnset.
func (n Note) CentsFromReferenceInTuning(t *tuning.Tuning) (cents.Cents, error) {
	return n.centsFromReference("CentsFromReferenceInTuning", tuningContext(t))
}

// MustCentsFromReferenceInTuning is like CentsFromReferenceInTuning but panics
// with its error.
func (n Note) MustCentsFromReferenceInTuning(t *tuning.Tuning) cents.Cents {
	return must(n.CentsFromReferenceInTuning(t))
}

// sound returns the frequency of the note n in the tuning t: the frequency of
// its written pitch shifted by its correction, rounded once.
func (n Note) sound(op string, t *tuning.Tuning) (float64, error) {
	f, err := t.ShiftedFrequency(n.pitch(), n.c)
	if err != nil {
		return 0, depError(op, err)
	}

	return f, nil
}

func (n Note) frequency(op string, c opContext) (float64, error) {
	f, err := n.prepare(op, c, true, true)
	if err != nil {
		return 0, err
	}

	return n.sound(op, f.tuning())
}

// Frequency returns the frequency of n in hertz: the frequency of its written
// pitch in the tuning of its settings shifted by its correction, with one
// rounding (tuning.Tuning.ShiftedFrequency). With a correction of 0 it equals
// PitchFrequency bit for bit. The note does not store the frequency: each call
// reads the source, so a retuned source changes the result.
//
// Errors, in this order: ErrUnset for the note; ErrNoOctave; ErrNoSource;
// ErrUnsetSettings; ErrInadmissible; ErrOutOfRange (also the error of tuning).
func (n Note) Frequency() (float64, error) {
	return n.frequency("Frequency", ownContext())
}

// MustFrequency is like Frequency but panics with its error.
func (n Note) MustFrequency() float64 {
	return must(n.Frequency())
}

// FrequencyIn is like Frequency in the settings of the explicit source src.
// A nil src reports ErrUnset.
func (n Note) FrequencyIn(src settings.Source) (float64, error) {
	return n.frequency("FrequencyIn", sourceContext(src))
}

// MustFrequencyIn is like FrequencyIn but panics with its error.
func (n Note) MustFrequencyIn(src settings.Source) float64 {
	return must(n.FrequencyIn(src))
}

// FrequencyInTuning is like Frequency in the explicit tuning t; it does not
// copy the tuning. A nil or unset t reports ErrUnset.
func (n Note) FrequencyInTuning(t *tuning.Tuning) (float64, error) {
	return n.frequency("FrequencyInTuning", tuningContext(t))
}

// MustFrequencyInTuning is like FrequencyInTuning but panics with its error.
func (n Note) MustFrequencyInTuning(t *tuning.Tuning) float64 {
	return must(n.FrequencyInTuning(t))
}

func (n Note) pitchFrequency(op string, c opContext) (float64, error) {
	f, err := n.prepare(op, c, true, true)
	if err != nil {
		return 0, err
	}

	freq, err := f.tuning().Frequency(n.pitch())
	if err != nil {
		return 0, depError(op, err)
	}

	return freq, nil
}

// PitchFrequency returns the frequency of the written pitch of n without the
// correction (tuning.Tuning.Frequency). The errors are those of Frequency.
func (n Note) PitchFrequency() (float64, error) {
	return n.pitchFrequency("PitchFrequency", ownContext())
}

// MustPitchFrequency is like PitchFrequency but panics with its error.
func (n Note) MustPitchFrequency() float64 {
	return must(n.PitchFrequency())
}

// PitchFrequencyIn is like PitchFrequency in the settings of the explicit
// source src. A nil src reports ErrUnset.
func (n Note) PitchFrequencyIn(src settings.Source) (float64, error) {
	return n.pitchFrequency("PitchFrequencyIn", sourceContext(src))
}

// MustPitchFrequencyIn is like PitchFrequencyIn but panics with its error.
func (n Note) MustPitchFrequencyIn(src settings.Source) float64 {
	return must(n.PitchFrequencyIn(src))
}

// PitchFrequencyInTuning is like PitchFrequency in the explicit tuning t. A nil
// or unset t reports ErrUnset.
func (n Note) PitchFrequencyInTuning(t *tuning.Tuning) (float64, error) {
	return n.pitchFrequency("PitchFrequencyInTuning", tuningContext(t))
}

// MustPitchFrequencyInTuning is like PitchFrequencyInTuning but panics with
// its error.
func (n Note) MustPitchFrequencyInTuning(t *tuning.Tuning) float64 {
	return must(n.PitchFrequencyInTuning(t))
}

func (n Note) isExpressible(op string, c opContext) (bool, error) {
	f, err := n.prepare(op, c, false, false)
	if err != nil {
		return false, err
	}

	return expresses(f.set, n), nil
}

// IsExpressible reports whether n is expressible by its settings: its name
// for a note without an octave (settings.Settings.Expresses), its written
// pitch for a note with an octave (settings.Settings.ExpressesPitch). The
// correction does not take part. An inadmissible or inexpressible note is the
// answer false, not an error.
//
// Errors, in this order: ErrUnset for the note; ErrNoSource;
// ErrUnsetSettings.
func (n Note) IsExpressible() (bool, error) {
	return n.isExpressible("IsExpressible", ownContext())
}

// MustIsExpressible is like IsExpressible but panics with its error.
func (n Note) MustIsExpressible() bool {
	return must(n.IsExpressible())
}

// IsExpressibleIn is like IsExpressible in the settings of the explicit source
// src. A nil src reports ErrUnset.
func (n Note) IsExpressibleIn(src settings.Source) (bool, error) {
	return n.isExpressible("IsExpressibleIn", sourceContext(src))
}

// MustIsExpressibleIn is like IsExpressibleIn but panics with its error.
func (n Note) MustIsExpressibleIn(src settings.Source) bool {
	return must(n.IsExpressibleIn(src))
}

func (n Note) format(op string, c opContext, nameOnly bool) (string, error) {
	f, err := n.prepare(op, c, false, false)
	if err != nil {
		return "", err
	}

	var text string

	if n.oct.IsValid() && !nameOnly {
		text, err = f.set.FormatPitch(n.pitch())
	} else {
		text, err = f.set.Format(n.name)
	}

	if err != nil {
		return "", formatError(op, err)
	}

	return text, nil
}

// Format returns the text of n in the notation of its settings: the name for
// a note without an octave (settings.Settings.Format), the written pitch for a
// note with an octave (settings.Settings.FormatPitch). The correction is not
// written.
//
// Errors, in this order: ErrUnset for the note; ErrNoSource;
// ErrUnsetSettings; ErrInadmissible (also settings.ErrInadmissible); the
// other errors of settings in the order of its contract, ErrOutOfRange for a
// result out of range. For octaves extremely far from the middle,
// settings.Settings.FormatPitch can panic; Format does not recover it.
func (n Note) Format() (string, error) {
	return n.format("Format", ownContext(), false)
}

// MustFormat is like Format but panics with its error.
func (n Note) MustFormat() string {
	return must(n.Format())
}

// FormatIn is like Format in the settings of the explicit source src. A nil
// src reports ErrUnset.
func (n Note) FormatIn(src settings.Source) (string, error) {
	return n.format("FormatIn", sourceContext(src), false)
}

// MustFormatIn is like FormatIn but panics with its error.
func (n Note) MustFormatIn(src settings.Source) string {
	return must(n.FormatIn(src))
}

// FormatName returns the text of the name of n without an octave in the
// notation of its settings (settings.Settings.Format), for a note with or
// without an octave. The errors are those of Format.
func (n Note) FormatName() (string, error) {
	return n.format("FormatName", ownContext(), true)
}

// MustFormatName is like FormatName but panics with its error.
func (n Note) MustFormatName() string {
	return must(n.FormatName())
}

// FormatNameIn is like FormatName in the settings of the explicit source src.
// A nil src reports ErrUnset.
func (n Note) FormatNameIn(src settings.Source) (string, error) {
	return n.format("FormatNameIn", sourceContext(src), true)
}

// MustFormatNameIn is like FormatNameIn but panics with its error.
func (n Note) MustFormatNameIn(src settings.Source) string {
	return must(n.FormatNameIn(src))
}
