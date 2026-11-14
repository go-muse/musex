package note

import (
	"errors"

	"github.com/go-muse/musex/settings"
)

// expresses reports whether the record of n (its name, or its written pitch)
// is expressible by set. The set and the note are set, so there is no error.
func expresses(set *settings.Settings, n Note) bool {
	var (
		ok  bool
		err error
	)

	if n.oct.IsValid() {
		ok, err = set.ExpressesPitch(n.pitch())
	} else {
		ok, err = set.Expresses(n.name)
	}

	return ok && err == nil
}

// transition is the result of the checks of a change of the source of n.
type transition struct {
	next       Note    // the note after the change, when it is possible
	violations []error // the violated conditions, in order
}

// planTransition reads the previous and the new settings and checks the
// change of the source of n to src: keeping the Octave, or keeping the
// written octave when written is true. ErrUnset and ErrUnsetSettings are
// returned as errors; the other conditions are collected: no source, an
// inadmissible name, lost expressibility, a result out of range. When all
// is true, every violated condition is collected; otherwise the collection
// stops at the first one.
func planTransition(op string, n *Note, src settings.Source, written, all bool) (transition, error) {
	cur, err := n.current(op)
	if err != nil {
		return transition{}, err
	}

	if err := checkSource(op, src); err != nil {
		return transition{}, err
	}

	r := newReader(op)

	var prev *settings.Settings

	if cur.src != nil {
		prev, err = r.read(cur.src)
		if err != nil {
			return transition{}, err
		}
	}

	next, err := r.read(src)
	if err != nil {
		return transition{}, err
	}

	var tr transition

	add := func(err error) bool {
		tr.violations = append(tr.violations, err)

		return all
	}

	res := cur
	res.src = src

	if !written || !cur.oct.IsValid() {
		inadmissible := admit(op, next.System(), cur.name)
		if inadmissible != nil && !add(inadmissible) {
			return tr, nil
		}

		if prev != nil {
			if expresses(prev, cur) && !expresses(next, cur) && !add(inexpressibleError(op)) {
				return tr, nil
			}
		}

		tr.next = res

		return tr, nil
	}

	if prev == nil {
		if !add(noSourceError(op)) {
			return tr, nil
		}
	} else if err := admit(op, prev.System(), cur.name); err != nil && !add(err) {
		return tr, nil
	}

	inadmissible := admit(op, next.System(), cur.name)
	if inadmissible != nil && !add(inadmissible) {
		return tr, nil
	}

	if prev == nil || len(tr.violations) > 0 {
		return tr, nil
	}

	p := cur.pitch()

	w, err := prev.System().WrittenOctave(p)
	if err != nil {
		add(depError(op, err))

		return tr, nil
	}

	np, err := next.System().FromWritten(cur.name, w)
	if err != nil {
		add(depError(op, err))

		return tr, nil
	}

	res = res.withPitch(np)

	if expresses(prev, cur) && !expresses(next, res) {
		add(inexpressibleError(op))

		return tr, nil
	}

	tr.next = res

	return tr, nil
}

// changeSource changes the source of n to src as planned by planTransition.
func (n *Note) changeSource(op string, src settings.Source, written bool) error {
	tr, err := planTransition(op, n, src, written, false)
	if err != nil {
		return err
	}

	if len(tr.violations) > 0 {
		return tr.violations[0]
	}

	*n = tr.next

	return nil
}

// SetSource binds n to the source src, keeping the name, the Octave and the
// correction. With another system the written octave of the note can change;
// use SetSourceWritten to keep it. Changing to the same source keeps the
// record and the correction, and the source of n becomes src.
//
// Errors, in this order: ErrUnset for the note or a nil src;
// ErrUnsetSettings for the settings of a bound note, then for the new
// settings (a source that is one source with the source of n is read once);
// ErrInadmissible when the name is not admissible in the new system;
// ErrInexpressible when a bound note is expressible by its settings and would
// not be expressible by the new ones. On an error n does not change. A bound
// note whose source returns unset settings can change its source only through
// Unbind and then SetSource.
func (n *Note) SetSource(src settings.Source) error {
	return n.changeSource("SetSource", src, false)
}

// MustSetSource is like SetSource but panics with its error; it returns n.
func (n *Note) MustSetSource(src settings.Source) *Note {
	return n.mustDo(n.SetSource(src))
}

// SetSourceWritten binds n to the source src, keeping the name, the
// correction and the written octave w of a note with an octave: w is computed
// in the previous system, and the Octave becomes w + k in the new system
// (system.System.FromWritten). A note without an octave changes as with
// SetSource.
//
// Errors, in this order: ErrUnset for the note or a nil src; ErrNoSource for
// an unbound note with an octave; ErrUnsetSettings for the previous, then the
// new settings; ErrInadmissible in the previous, then the new system;
// ErrOutOfRange (also system.ErrOutOfRange) when w or the new Octave does not
// fit in an int; ErrInexpressible.
func (n *Note) SetSourceWritten(src settings.Source) error {
	const op = "SetSourceWritten"

	if n != nil && n.name.IsValid() && src != nil && n.src == nil && n.oct.IsValid() {
		return noSourceError(op)
	}

	return n.changeSource(op, src, true)
}

// MustSetSourceWritten is like SetSourceWritten but panics with its error; it
// returns n.
func (n *Note) MustSetSourceWritten(src settings.Source) *Note {
	return n.mustDo(n.SetSourceWritten(src))
}

// Unbind makes n unbound, keeping its record and correction; an unbound note
// does not change. The source is not read.
//
// Errors: ErrUnset for the note.
func (n *Note) Unbind() error {
	cur, err := n.current("Unbind")
	if err != nil {
		return err
	}

	cur.src = nil
	*n = cur

	return nil
}

// MustUnbind is like Unbind but panics with its error; it returns n.
func (n *Note) MustUnbind() *Note {
	return n.mustDo(n.Unbind())
}

// SourceCheck is the result of a check of a change of the source of a note:
// whether the change is possible and, when it is not, all the violated
// conditions. The zero SourceCheck means "not checked": Possible is false,
// Violations is empty and Err reports ErrUnset.
//
// A SourceCheck is immutable, is copied by assignment and is safe for
// concurrent use. == is not part of its API.
type SourceCheck struct {
	checked    bool
	violations []error
}

// Possible reports whether the check was done and found no violated
// condition.
func (c SourceCheck) Possible() bool {
	return c.checked && len(c.violations) == 0
}

// Violations returns the violated conditions, in this order: no source
// (ErrNoSource), an inadmissible name (ErrInadmissible), lost expressibility
// (ErrInexpressible), a result out of range (ErrOutOfRange). Each one matches
// exactly one sentinel of note. The result is a new slice; changing it does
// not change c.
func (c SourceCheck) Violations() []error {
	out := make([]error, len(c.violations))
	copy(out, c.violations)

	return out
}

// Err returns nil when the change is possible, the joined violated conditions
// (errors.Join) otherwise, and an error that matches ErrUnset for the zero
// SourceCheck. Err returns the result of the check as data; it has no Must
// variant.
func (c SourceCheck) Err() error {
	if !c.checked {
		return unsetError("SourceCheck", "not checked")
	}

	if len(c.violations) == 0 {
		return nil
	}

	return errors.Join(c.violations...)
}

// check runs planTransition and collects all the violated conditions.
func (n Note) check(op string, src settings.Source, written bool) (SourceCheck, error) {
	tr, err := planTransition(op, &n, src, written, true)
	if err != nil {
		return SourceCheck{}, err
	}

	return SourceCheck{checked: true, violations: tr.violations}, nil
}

// CheckSetSource checks whether SetSource(src) is possible, without changing
// n. It reports an error only for ErrUnset (the note, a nil src) and
// ErrUnsetSettings (the settings of n, then of src); every other reason is a
// violated condition of the result: an inadmissible name and lost
// expressibility.
func (n Note) CheckSetSource(src settings.Source) (SourceCheck, error) {
	return n.check("CheckSetSource", src, false)
}

// MustCheckSetSource is like CheckSetSource but panics with its error.
func (n Note) MustCheckSetSource(src settings.Source) SourceCheck {
	return must(n.CheckSetSource(src))
}

// CheckSetSourceWritten checks whether SetSourceWritten(src) is possible,
// without changing n; see CheckSetSource. For an unbound note with an octave
// the violated condition is no source (ErrNoSource); the conditions that need
// the previous settings (the written octave, a result out of range, lost
// expressibility) are then not checked. Lost expressibility is checked only
// for a result that fits in range.
func (n Note) CheckSetSourceWritten(src settings.Source) (SourceCheck, error) {
	return n.check("CheckSetSourceWritten", src, true)
}

// MustCheckSetSourceWritten is like CheckSetSourceWritten but panics with its
// error.
func (n Note) MustCheckSetSourceWritten(src settings.Source) SourceCheck {
	return must(n.CheckSetSourceWritten(src))
}
