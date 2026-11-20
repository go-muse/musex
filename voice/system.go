package voice

import (
	"fmt"

	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/system"
)

// soundIndexes returns the indexes of the sounds of l in order.
func (l *Line) soundIndexes() []int {
	out := make([]int, 0, len(l.elems))

	for i, e := range l.elems {
		if e.kind == KindSound {
			out = append(out, i)
		}
	}

	return out
}

// pitchOf returns the written pitch of the sound i, which always has one.
func (l *Line) pitchOf(i int) pitch.Pitch {
	p, _ := l.elems[i].note.Pitch()

	return p
}

// checkSystem reports ErrUnset with system.ErrInvalid for a nil or unset
// explicit system.
func checkSystem(sys *system.System) error {
	if sys == nil || !sys.IsValid() {
		return unsetError("system", system.ErrInvalid)
	}

	return nil
}

// noteSystem returns the system of the current settings of the bound note n.
func noteSystem(n note.Note) (*system.System, error) {
	set, err := n.Settings()
	if err != nil {
		return nil, fmt.Errorf("voice: settings of %s: %w", n, err)
	}

	return set.System(), nil
}

// checkBound reports ErrNoSystem in an *ItemError for the first unbound sound.
func (l *Line) checkBound(sounds []int) error {
	for _, i := range sounds {
		if !l.elems[i].note.IsBound() {
			return itemError(i, fmt.Errorf("%w: unbound %s", ErrNoSystem, l.elems[i].note))
		}
	}

	return nil
}

// compareSystem chooses the system of a comparison of the pitches of the
// sounds when no system is given. With refSys, the system of the preset or
// the instrument of the reference range, unbound sounds are allowed and every
// bound sound must have that system. Otherwise every sound must be bound and
// all sounds must have the system of the first one. The settings of all the
// sounds concerned are read before the systems are compared.
func (l *Line) compareSystem(sounds []int, refSys *system.System) (*system.System, error) {
	if refSys == nil {
		if err := l.checkBound(sounds); err != nil {
			return nil, err
		}
	}

	systems := make([]*system.System, len(sounds))

	for k, i := range sounds {
		n := l.elems[i].note
		if !n.IsBound() {
			continue
		}

		sys, err := noteSystem(n)
		if err != nil {
			return nil, itemError(i, err)
		}

		systems[k] = sys
	}

	want := refSys
	if want == nil {
		want = systems[0]
	}

	for k, i := range sounds {
		if systems[k] != nil && !systems[k].Equal(want) {
			return nil, itemError(i, fmt.Errorf("%w: %s and %s", ErrDifferentSystems, systems[k], want))
		}
	}

	return want, nil
}

// checkSounds reports system.ErrInadmissible in an *ItemError for the first
// sound whose name is not admissible in sys.
func (l *Line) checkSounds(sys *system.System, sounds []int) error {
	for _, i := range sounds {
		if err := checkAdmits(sys, l.pitchOf(i), "sound"); err != nil {
			return itemError(i, err)
		}
	}

	return nil
}

// checkAdmits reports system.ErrInadmissible, and no other sentinel, when the
// name of p is not admissible in sys.
func checkAdmits(sys *system.System, p pitch.Pitch, what string) error {
	ok, err := sys.AdmitsPitch(p)
	if err != nil {
		return fmt.Errorf("voice: %s %s: %w", what, p, err)
	}

	if !ok {
		return fmt.Errorf("voice: %s %s in %s: %w", what, p, sys, system.ErrInadmissible)
	}

	return nil
}
