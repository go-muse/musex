package note

import (
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"
)

// reader reads the sources of one operation. Each source is read once: a
// source that is one source with an already read one (settings.SameSource)
// gives the same snapshot.
type reader struct {
	op   string
	srcs []settings.Source
	sets []*settings.Settings
}

// newReader returns a reader for the operation op.
func newReader(op string) *reader {
	return &reader{op: op}
}

// read returns the snapshot of src, reading it at most once per operation.
// Unset settings report ErrUnsetSettings.
func (r *reader) read(src settings.Source) (*settings.Settings, error) {
	for i, s := range r.srcs {
		if settings.SameSource(s, src) {
			return r.sets[i], nil
		}
	}

	set := src.Settings()
	if !set.IsValid() {
		return nil, unsetSettingsError(r.op)
	}

	r.srcs = append(r.srcs, src)
	r.sets = append(r.sets, set)

	return set, nil
}

// admit reports ErrInadmissible when name is not admissible in sys. The
// system and the name are set.
func admit(op string, sys *system.System, name spelling.Spelling) error {
	ok, err := sys.Admits(name)
	if err != nil {
		return depError(op, err)
	}

	if !ok {
		return inadmissibleError(op, name, sys)
	}

	return nil
}

// frame is the context of a computation over one or several notes: the
// settings (unset for an explicit tuning), the system and, when it is needed,
// the tuning.
type frame struct {
	set *settings.Settings
	sys *system.System
	t   *tuning.Tuning
}

// tuning returns the tuning of the frame: the explicit tuning, or a copy of
// the tuning of the settings made on the first call.
func (f *frame) tuning() *tuning.Tuning {
	if f.t == nil {
		f.t = f.set.Tuning()
	}

	return f.t
}

// tuningFrame returns the frame of an explicit tuning t, or ErrUnset when t
// is nil or unset.
func tuningFrame(op string, t *tuning.Tuning) (frame, error) {
	if t == nil || !t.IsValid() {
		return frame{}, unsetError(op, "tuning")
	}

	return frame{sys: t.System(), t: t}, nil
}

// checkSource reports ErrUnset for a nil explicit source.
func checkSource(op string, src settings.Source) error {
	if src == nil {
		return unsetError(op, "source")
	}

	return nil
}

// checkNote reports ErrUnset for the unset note.
func checkNote(op string, n Note) error {
	if !n.name.IsValid() {
		return unsetError(op, "note")
	}

	return nil
}

// checkOctave reports ErrNoOctave for a note without an octave.
func checkOctave(op string, n Note) error {
	if !n.oct.IsValid() {
		return noOctaveError(op, n.name)
	}

	return nil
}

// ownFrame returns the frame of the settings of the note n: ErrNoSource for
// an unbound note, ErrUnsetSettings for unset settings.
func ownFrame(r *reader, n Note) (frame, error) {
	if n.src == nil {
		return frame{}, noSourceError(r.op)
	}

	set, err := r.read(n.src)
	if err != nil {
		return frame{}, err
	}

	return frame{set: set, sys: set.System()}, nil
}

// sourceFrame returns the frame of the settings of an explicit source.
func sourceFrame(r *reader, src settings.Source) (frame, error) {
	set, err := r.read(src)
	if err != nil {
		return frame{}, err
	}

	return frame{set: set, sys: set.System()}, nil
}
