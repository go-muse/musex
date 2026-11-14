package note

import (
	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/frequency"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"
)

// Comparisons have three levels:
//   - without settings (Equal, SameRecord, Compare and the like): any notes,
//     bound to any sources or unbound; the source never takes part;
//   - by position (Equivalent, ComparePosition, Between, Distance,
//     SortedByPosition): the systems of the settings of all notes must be
//     equal; the correction does not take part;
//   - by sound (CompareSound, SoundsSame, CentsTo, SortedBySound, Offset):
//     each note in its own settings, which may differ. Strict comparisons of
//     notes with equal tunings compare the exact heights h + c; otherwise the
//     frequencies are compared, with a criterion of package frequency when
//     one is given.

// Equal reports whether n and other are set and have equal records and
// corrections: equal names for notes without an octave, equal written pitches
// for notes with an octave (a note with an octave never equals a note without
// one), and equal corrections. The source does not take part. Two unset notes
// are not equal.
func (n Note) Equal(other Note) bool {
	return n.SameRecord(other) && n.c.Equal(other.c)
}

// SameRecord reports whether n and other are set and have equal records,
// regardless of the corrections.
func (n Note) SameRecord(other Note) bool {
	if !n.name.IsValid() || !other.name.IsValid() {
		return false
	}

	if n.oct.IsValid() != other.oct.IsValid() {
		return false
	}

	return n.name.Equal(other.name) && (!n.oct.IsValid() || n.oct.Equal(other.oct))
}

// SameSpelling reports whether n and other are set and have one name,
// regardless of the octaves.
func (n Note) SameSpelling(other Note) bool {
	return n.name.IsValid() && other.name.IsValid() && n.name.Equal(other.name)
}

// SameLetter reports whether n and other are set and have one letter.
func (n Note) SameLetter(other Note) bool {
	return n.name.IsValid() && other.name.IsValid() && n.name.SameLetter(other.name)
}

// SameAlteration reports whether n and other are set and have one value of
// the alteration.
func (n Note) SameAlteration(other Note) bool {
	return n.name.IsValid() && other.name.IsValid() && n.name.SameAlteration(other.name)
}

// SameOctave reports whether n and other both have an octave and their
// Octaves are equal.
func (n Note) SameOctave(other Note) bool {
	return n.HasOctave() && other.HasOctave() && n.oct.Equal(other.oct)
}

// SameCents reports whether n and other are set and have equal corrections.
func (n Note) SameCents(other Note) bool {
	return n.name.IsValid() && other.name.IsValid() && n.c.Equal(other.c)
}

// SameSource reports whether n and other are both bound to one source
// (settings.SameSource). It never reads the sources.
func (n Note) SameSource(other Note) bool {
	return n.src != nil && other.src != nil && settings.SameSource(n.src, other.src)
}

// Compare returns -1, 0 or +1 by an agreed order of records: the unset note
// first, notes without an octave before notes with one, then by
// spelling.Spelling.Compare or pitch.Pitch.Compare, then by the correction.
// Two unset notes compare as 0. The order is neither equality nor sound.
func (n Note) Compare(other Note) int {
	switch {
	case !n.name.IsValid() && !other.name.IsValid():
		return 0
	case !n.name.IsValid():
		return -1
	case !other.name.IsValid():
		return 1
	}

	switch {
	case !n.oct.IsValid() && other.oct.IsValid():
		return -1
	case n.oct.IsValid() && !other.oct.IsValid():
		return 1
	}

	var r int
	if n.oct.IsValid() {
		r = n.pitch().Compare(other.pitch())
	} else {
		r = n.name.Compare(other.name)
	}

	if r != 0 {
		return r
	}

	return n.c.Compare(other.c)
}

// octaves checks n and other for OctaveSub and OctaveDistance.
func (n Note) octaves(op string, other Note) error {
	notes := [2]Note{n, other}

	if err := checkNotes(op, notes[:]); err != nil {
		return err
	}

	return checkOctaves(op, notes[:])
}

// OctaveSub returns how many octaves the Octave of n is above the Octave of
// other (pitch.Pitch.OctaveSub).
//
// Errors, in this order: ErrUnset for a note; ErrNoOctave; ErrOutOfRange
// (also octave.ErrOutOfRange).
func (n Note) OctaveSub(other Note) (int, error) {
	const op = "OctaveSub"

	if err := n.octaves(op, other); err != nil {
		return 0, err
	}

	d, err := n.oct.Sub(other.oct)
	if err != nil {
		return 0, depError(op, err)
	}

	return d, nil
}

// MustOctaveSub is like OctaveSub but panics with its error.
func (n Note) MustOctaveSub(other Note) int {
	return must(n.OctaveSub(other))
}

// OctaveDistance returns the distance between the Octaves of n and other in
// octaves, without a direction (pitch.Pitch.OctaveDistance). The errors are
// those of OctaveSub.
func (n Note) OctaveDistance(other Note) (int, error) {
	const op = "OctaveDistance"

	if err := n.octaves(op, other); err != nil {
		return 0, err
	}

	d, err := n.oct.Distance(other.oct)
	if err != nil {
		return 0, depError(op, err)
	}

	return d, nil
}

// MustOctaveDistance is like OctaveDistance but panics with its error.
func (n Note) MustOctaveDistance(other Note) int {
	return must(n.OctaveDistance(other))
}

// checkNotes reports ErrUnset for the first unset note.
func checkNotes(op string, notes []Note) error {
	for _, x := range notes {
		if err := checkNote(op, x); err != nil {
			return err
		}
	}

	return nil
}

// checkOctaves reports ErrNoOctave for the first note without an octave.
func checkOctaves(op string, notes []Note) error {
	for _, x := range notes {
		if err := checkOctave(op, x); err != nil {
			return err
		}
	}

	return nil
}

// frames returns the frame of each note in the context c: ErrNoSource for
// the first unbound note, then ErrUnsetSettings while reading the sources in
// order (each source once), then ErrInadmissible for the first inadmissible
// name.
func frames(op string, r *reader, c context, notes []Note) ([]frame, error) {
	fs := make([]frame, len(notes))

	switch {
	case c.in || c.inTuned:
		f, err := c.frame(r, Note{})
		if err != nil {
			return nil, err
		}

		for i := range fs {
			fs[i] = f
		}
	default:
		for _, x := range notes {
			if x.src == nil {
				return nil, noSourceError(op)
			}
		}

		for i, x := range notes {
			set, err := r.read(x.src)
			if err != nil {
				return nil, err
			}

			fs[i] = frame{set: set, sys: set.System()}
		}
	}

	for i, x := range notes {
		if err := admit(op, fs[i].sys, x.name); err != nil {
			return nil, err
		}
	}

	return fs, nil
}

// positionSystem returns the one system of the notes for an operation by
// position: the checks of frames, then ErrDifferentSystems.
func positionSystem(op string, r *reader, c context, notes []Note) (*system.System, error) {
	fs, err := frames(op, r, c, notes)
	if err != nil {
		return nil, err
	}

	sys := fs[0].sys

	for _, f := range fs[1:] {
		if !sys.Equal(f.sys) {
			return nil, differentSystemsError(op, sys, f.sys)
		}
	}

	return sys, nil
}

// position checks the notes for an operation by position and returns their
// system. Every note must have an octave unless names is true; then all the
// notes must have an octave, or none.
func position(op string, c context, notes []Note, names bool) (*system.System, error) {
	if err := checkNotes(op, notes); err != nil {
		return nil, err
	}

	if err := c.check(op); err != nil {
		return nil, err
	}

	if names {
		for _, x := range notes[1:] {
			if x.oct.IsValid() != notes[0].oct.IsValid() {
				return nil, noOctaveError(op, x.name)
			}
		}
	} else if err := checkOctaves(op, notes); err != nil {
		return nil, err
	}

	return positionSystem(op, newReader(op), c, notes)
}

func (n Note) equivalent(op string, c context, other Note) (bool, error) {
	sys, err := position(op, c, []Note{n, other}, true)
	if err != nil {
		return false, err
	}

	var ok bool
	if n.oct.IsValid() {
		ok, err = sys.Equivalent(n.pitch(), other.pitch())
	} else {
		ok, err = sys.EquivalentNames(n.name, other.name)
	}

	if err != nil {
		return false, depError(op, err)
	}

	return ok, nil
}

// Equivalent reports whether n and other are at one position
// (system.System.Equivalent): B♯3 and C4 are equivalent. Two notes without an
// octave are equivalent when their names are of one class
// (system.System.EquivalentNames). The correction does not take part.
//
// Errors, in this order: ErrUnset for a note; ErrNoOctave for a note with an
// octave and a note without one; ErrNoSource; ErrUnsetSettings;
// ErrInadmissible; ErrDifferentSystems.
func (n Note) Equivalent(other Note) (bool, error) {
	return n.equivalent("Equivalent", ownContext(), other)
}

// MustEquivalent is like Equivalent but panics with its error.
func (n Note) MustEquivalent(other Note) bool {
	return must(n.Equivalent(other))
}

// EquivalentIn is like Equivalent in the settings of the explicit source src
// for both notes. A nil src reports ErrUnset.
func (n Note) EquivalentIn(src settings.Source, other Note) (bool, error) {
	return n.equivalent("EquivalentIn", sourceContext(src), other)
}

// MustEquivalentIn is like EquivalentIn but panics with its error.
func (n Note) MustEquivalentIn(src settings.Source, other Note) bool {
	return must(n.EquivalentIn(src, other))
}

// EquivalentInTuning is like Equivalent in the system of the explicit tuning
// t. A nil or unset t reports ErrUnset.
func (n Note) EquivalentInTuning(t *tuning.Tuning, other Note) (bool, error) {
	return n.equivalent("EquivalentInTuning", tuningContext(t), other)
}

// MustEquivalentInTuning is like EquivalentInTuning but panics with its error.
func (n Note) MustEquivalentInTuning(t *tuning.Tuning, other Note) bool {
	return must(n.EquivalentInTuning(t, other))
}

func (n Note) comparePosition(op string, c context, other Note) (int, error) {
	sys, err := position(op, c, []Note{n, other}, false)
	if err != nil {
		return 0, err
	}

	r, err := sys.Compare(n.pitch(), other.pitch())
	if err != nil {
		return 0, depError(op, err)
	}

	return r, nil
}

// ComparePosition returns -1, 0 or +1 when the position of n is lower than,
// equal to or higher than the position of other (system.System.Compare). The
// correction does not take part.
//
// Errors, in this order: ErrUnset for a note; ErrNoOctave; ErrNoSource;
// ErrUnsetSettings; ErrInadmissible; ErrDifferentSystems.
func (n Note) ComparePosition(other Note) (int, error) {
	return n.comparePosition("ComparePosition", ownContext(), other)
}

// MustComparePosition is like ComparePosition but panics with its error.
func (n Note) MustComparePosition(other Note) int {
	return must(n.ComparePosition(other))
}

// ComparePositionIn is like ComparePosition in the settings of the explicit
// source src for both notes. A nil src reports ErrUnset.
func (n Note) ComparePositionIn(src settings.Source, other Note) (int, error) {
	return n.comparePosition("ComparePositionIn", sourceContext(src), other)
}

// MustComparePositionIn is like ComparePositionIn but panics with its error.
func (n Note) MustComparePositionIn(src settings.Source, other Note) int {
	return must(n.ComparePositionIn(src, other))
}

// ComparePositionInTuning is like ComparePosition in the system of the
// explicit tuning t. A nil or unset t reports ErrUnset.
func (n Note) ComparePositionInTuning(t *tuning.Tuning, other Note) (int, error) {
	return n.comparePosition("ComparePositionInTuning", tuningContext(t), other)
}

// MustComparePositionInTuning is like ComparePositionInTuning but panics with
// its error.
func (n Note) MustComparePositionInTuning(t *tuning.Tuning, other Note) int {
	return must(n.ComparePositionInTuning(t, other))
}

func (n Note) between(op string, c context, a, b Note) (bool, error) {
	sys, err := position(op, c, []Note{n, a, b}, false)
	if err != nil {
		return false, err
	}

	ok, err := sys.Between(n.pitch(), a.pitch(), b.pitch())
	if err != nil {
		return false, depError(op, err)
	}

	return ok, nil
}

// Between reports whether the position of n is not lower than the lower and
// not higher than the higher of the positions of a and b, both included
// (system.System.Between). The errors are those of ComparePosition.
func (n Note) Between(a, b Note) (bool, error) {
	return n.between("Between", ownContext(), a, b)
}

// MustBetween is like Between but panics with its error.
func (n Note) MustBetween(a, b Note) bool {
	return must(n.Between(a, b))
}

// BetweenIn is like Between in the settings of the explicit source src for
// all notes. A nil src reports ErrUnset.
func (n Note) BetweenIn(src settings.Source, a, b Note) (bool, error) {
	return n.between("BetweenIn", sourceContext(src), a, b)
}

// MustBetweenIn is like BetweenIn but panics with its error.
func (n Note) MustBetweenIn(src settings.Source, a, b Note) bool {
	return must(n.BetweenIn(src, a, b))
}

// BetweenInTuning is like Between in the system of the explicit tuning t. A
// nil or unset t reports ErrUnset.
func (n Note) BetweenInTuning(t *tuning.Tuning, a, b Note) (bool, error) {
	return n.between("BetweenInTuning", tuningContext(t), a, b)
}

// MustBetweenInTuning is like BetweenInTuning but panics with its error.
func (n Note) MustBetweenInTuning(t *tuning.Tuning, a, b Note) bool {
	return must(n.BetweenInTuning(t, a, b))
}

func (n Note) distance(op string, c context, to Note) (system.Step, error) {
	sys, err := position(op, c, []Note{n, to}, false)
	if err != nil {
		return 0, err
	}

	d, err := sys.Distance(n.pitch(), to.pitch())
	if err != nil {
		return 0, depError(op, err)
	}

	return d, nil
}

// Distance returns the number of steps from the position of n to the
// position of to: P(to) - P(n) (system.System.Distance). The errors are those
// of ComparePosition, then ErrOutOfRange (also system.ErrOutOfRange).
func (n Note) Distance(to Note) (system.Step, error) {
	return n.distance("Distance", ownContext(), to)
}

// MustDistance is like Distance but panics with its error.
func (n Note) MustDistance(to Note) system.Step {
	return must(n.Distance(to))
}

// DistanceIn is like Distance in the settings of the explicit source src for
// both notes. A nil src reports ErrUnset.
func (n Note) DistanceIn(src settings.Source, to Note) (system.Step, error) {
	return n.distance("DistanceIn", sourceContext(src), to)
}

// MustDistanceIn is like DistanceIn but panics with its error.
func (n Note) MustDistanceIn(src settings.Source, to Note) system.Step {
	return must(n.DistanceIn(src, to))
}

// DistanceInTuning is like Distance in the system of the explicit tuning t. A
// nil or unset t reports ErrUnset.
func (n Note) DistanceInTuning(t *tuning.Tuning, to Note) (system.Step, error) {
	return n.distance("DistanceInTuning", tuningContext(t), to)
}

// MustDistanceInTuning is like DistanceInTuning but panics with its error.
func (n Note) MustDistanceInTuning(t *tuning.Tuning, to Note) system.Step {
	return must(n.DistanceInTuning(t, to))
}

// sounding prepares notes for an operation by sound: the checks of the notes
// (ErrUnset, then the criterion when crit is not nil), ErrNoOctave, then the
// frames of the notes.
func sounding(op string, c context, crit *frequency.Criterion, notes []Note) ([]frame, error) {
	if err := checkNotes(op, notes); err != nil {
		return nil, err
	}

	if err := c.check(op); err != nil {
		return nil, err
	}

	if crit != nil && !crit.IsValid() {
		return nil, unsetCriterionError(op)
	}

	if err := checkOctaves(op, notes); err != nil {
		return nil, err
	}

	return frames(op, newReader(op), c, notes)
}

// isStrict reports whether crit is absent or a strict criterion.
func isStrict(crit *frequency.Criterion) bool {
	return crit == nil || crit.Kind() == frequency.KindStrict
}

// exactDistance returns the exact distance in cents from the sound of x to
// the sound of y in one tuning t: h(y) + c(y) - (h(x) + c(x)).
func exactDistance(op string, t *tuning.Tuning, x, y Note) (cents.Cents, error) {
	d, err := t.CentsBetween(x.pitch(), y.pitch())
	if err != nil {
		return cents.Cents{}, depError(op, err)
	}

	d, err = d.Add(y.c)
	if err != nil {
		return cents.Cents{}, depError(op, err)
	}

	d, err = d.Sub(x.c)
	if err != nil {
		return cents.Cents{}, depError(op, err)
	}

	return d, nil
}

// soundPair is a pair of notes prepared for a comparison by sound.
type soundPair struct {
	x, y   Note
	tx, ty *tuning.Tuning
}

// exact reports whether the pair is compared by exact heights: a strict
// comparison of notes with equal tunings.
func (p soundPair) exact(crit *frequency.Criterion) bool {
	return isStrict(crit) && p.tx.Equal(p.ty)
}

// frequencies returns the frequencies of the two notes.
func (p soundPair) frequencies(op string) (float64, float64, error) {
	fx, err := p.x.sound(op, p.tx)
	if err != nil {
		return 0, 0, err
	}

	fy, err := p.y.sound(op, p.ty)
	if err != nil {
		return 0, 0, err
	}

	return fx, fy, nil
}

// pair prepares n and other for a comparison by sound.
func (n Note) pair(op string, c context, crit *frequency.Criterion, other Note) (soundPair, error) {
	fs, err := sounding(op, c, crit, []Note{n, other})
	if err != nil {
		return soundPair{}, err
	}

	tx := fs[0].tuning()

	ty := tx
	if !c.in && !c.inTuned {
		ty = fs[1].tuning()
	}

	return soundPair{x: n, y: other, tx: tx, ty: ty}, nil
}

func (n Note) compareSound(op string, c context, crit *frequency.Criterion, other Note) (int, error) {
	p, err := n.pair(op, c, crit, other)
	if err != nil {
		return 0, err
	}

	if p.exact(crit) {
		d, err := exactDistance(op, p.tx, n, other)
		if err != nil {
			return 0, err
		}

		return -d.Sign(), nil
	}

	fx, fy, err := p.frequencies(op)
	if err != nil {
		return 0, err
	}

	var r int
	if crit == nil {
		r, err = frequency.Compare(fx, fy)
	} else {
		r, err = crit.Compare(fx, fy)
	}

	if err != nil {
		return 0, depError(op, err)
	}

	return r, nil
}

// CompareSound returns -1, 0 or +1 when n sounds lower than, the same as or
// higher than other, strictly. Each note is in its own settings; with equal
// tunings the exact heights h + c are compared, otherwise the frequencies.
//
// Errors, in this order: ErrUnset for a note; ErrNoOctave; ErrNoSource;
// ErrUnsetSettings; ErrInadmissible; ErrOutOfRange (also the error of tuning
// or cents).
func (n Note) CompareSound(other Note) (int, error) {
	return n.compareSound("CompareSound", ownContext(), nil, other)
}

// MustCompareSound is like CompareSound but panics with its error.
func (n Note) MustCompareSound(other Note) int {
	return must(n.CompareSound(other))
}

// CompareSoundIn is like CompareSound in the settings of the explicit source
// src for both notes; the comparison is then always exact. A nil src reports
// ErrUnset.
func (n Note) CompareSoundIn(src settings.Source, other Note) (int, error) {
	return n.compareSound("CompareSoundIn", sourceContext(src), nil, other)
}

// MustCompareSoundIn is like CompareSoundIn but panics with its error.
func (n Note) MustCompareSoundIn(src settings.Source, other Note) int {
	return must(n.CompareSoundIn(src, other))
}

// CompareSoundInTuning is like CompareSound in the explicit tuning t for both
// notes. A nil or unset t reports ErrUnset.
func (n Note) CompareSoundInTuning(t *tuning.Tuning, other Note) (int, error) {
	return n.compareSound("CompareSoundInTuning", tuningContext(t), nil, other)
}

// MustCompareSoundInTuning is like CompareSoundInTuning but panics with its
// error.
func (n Note) MustCompareSoundInTuning(t *tuning.Tuning, other Note) int {
	return must(n.CompareSoundInTuning(t, other))
}

// CompareSoundBy is like CompareSound with the criterion c of package
// frequency: 0 when the frequencies are near by c, otherwise their strict
// order (frequency.Criterion.Compare). A strict criterion behaves as
// CompareSound. An unset criterion reports ErrUnset, also frequency.ErrUnset.
func (n Note) CompareSoundBy(other Note, c frequency.Criterion) (int, error) {
	return n.compareSound("CompareSoundBy", ownContext(), &c, other)
}

// MustCompareSoundBy is like CompareSoundBy but panics with its error.
func (n Note) MustCompareSoundBy(other Note, c frequency.Criterion) int {
	return must(n.CompareSoundBy(other, c))
}

// CompareSoundByIn is like CompareSoundBy in the settings of the explicit
// source src for both notes. A nil src reports ErrUnset.
func (n Note) CompareSoundByIn(src settings.Source, other Note, c frequency.Criterion) (int, error) {
	return n.compareSound("CompareSoundByIn", sourceContext(src), &c, other)
}

// MustCompareSoundByIn is like CompareSoundByIn but panics with its error.
func (n Note) MustCompareSoundByIn(src settings.Source, other Note, c frequency.Criterion) int {
	return must(n.CompareSoundByIn(src, other, c))
}

// CompareSoundByInTuning is like CompareSoundBy in the explicit tuning t for
// both notes. A nil or unset t reports ErrUnset.
func (n Note) CompareSoundByInTuning(t *tuning.Tuning, other Note, c frequency.Criterion) (int, error) {
	return n.compareSound("CompareSoundByInTuning", tuningContext(t), &c, other)
}

// MustCompareSoundByInTuning is like CompareSoundByInTuning but panics with
// its error.
func (n Note) MustCompareSoundByInTuning(t *tuning.Tuning, other Note, c frequency.Criterion) int {
	return must(n.CompareSoundByInTuning(t, other, c))
}

func (n Note) soundsSame(op string, c context, crit *frequency.Criterion, other Note) (bool, error) {
	p, err := n.pair(op, c, crit, other)
	if err != nil {
		return false, err
	}

	if p.exact(crit) {
		d, err := exactDistance(op, p.tx, n, other)
		if err != nil {
			return false, err
		}

		return d.Sign() == 0, nil
	}

	fx, fy, err := p.frequencies(op)
	if err != nil {
		return false, err
	}

	var ok bool
	if crit == nil {
		ok, err = frequency.Equal(fx, fy)
	} else {
		ok, err = crit.Near(fx, fy)
	}

	if err != nil {
		return false, depError(op, err)
	}

	return ok, nil
}

// SoundsSame reports whether n and other sound the same, strictly; see
// CompareSound. A4 +100 cents and A♯4 sound the same.
func (n Note) SoundsSame(other Note) (bool, error) {
	return n.soundsSame("SoundsSame", ownContext(), nil, other)
}

// MustSoundsSame is like SoundsSame but panics with its error.
func (n Note) MustSoundsSame(other Note) bool {
	return must(n.SoundsSame(other))
}

// SoundsSameIn is like SoundsSame in the settings of the explicit source src
// for both notes. A nil src reports ErrUnset.
func (n Note) SoundsSameIn(src settings.Source, other Note) (bool, error) {
	return n.soundsSame("SoundsSameIn", sourceContext(src), nil, other)
}

// MustSoundsSameIn is like SoundsSameIn but panics with its error.
func (n Note) MustSoundsSameIn(src settings.Source, other Note) bool {
	return must(n.SoundsSameIn(src, other))
}

// SoundsSameInTuning is like SoundsSame in the explicit tuning t for both
// notes. A nil or unset t reports ErrUnset.
func (n Note) SoundsSameInTuning(t *tuning.Tuning, other Note) (bool, error) {
	return n.soundsSame("SoundsSameInTuning", tuningContext(t), nil, other)
}

// MustSoundsSameInTuning is like SoundsSameInTuning but panics with its error.
func (n Note) MustSoundsSameInTuning(t *tuning.Tuning, other Note) bool {
	return must(n.SoundsSameInTuning(t, other))
}

// SoundsSameBy reports whether the frequencies of n and other are near by the
// criterion c (frequency.Criterion.Near); a strict criterion behaves as
// SoundsSame. An unset criterion reports ErrUnset, also frequency.ErrUnset.
func (n Note) SoundsSameBy(other Note, c frequency.Criterion) (bool, error) {
	return n.soundsSame("SoundsSameBy", ownContext(), &c, other)
}

// MustSoundsSameBy is like SoundsSameBy but panics with its error.
func (n Note) MustSoundsSameBy(other Note, c frequency.Criterion) bool {
	return must(n.SoundsSameBy(other, c))
}

// SoundsSameByIn is like SoundsSameBy in the settings of the explicit source
// src for both notes. A nil src reports ErrUnset.
func (n Note) SoundsSameByIn(src settings.Source, other Note, c frequency.Criterion) (bool, error) {
	return n.soundsSame("SoundsSameByIn", sourceContext(src), &c, other)
}

// MustSoundsSameByIn is like SoundsSameByIn but panics with its error.
func (n Note) MustSoundsSameByIn(src settings.Source, other Note, c frequency.Criterion) bool {
	return must(n.SoundsSameByIn(src, other, c))
}

// SoundsSameByInTuning is like SoundsSameBy in the explicit tuning t for both
// notes. A nil or unset t reports ErrUnset.
func (n Note) SoundsSameByInTuning(t *tuning.Tuning, other Note, c frequency.Criterion) (bool, error) {
	return n.soundsSame("SoundsSameByInTuning", tuningContext(t), &c, other)
}

// MustSoundsSameByInTuning is like SoundsSameByInTuning but panics with its
// error.
func (n Note) MustSoundsSameByInTuning(t *tuning.Tuning, other Note, c frequency.Criterion) bool {
	return must(n.SoundsSameByInTuning(t, other, c))
}

func (n Note) centsTo(op string, c context, other Note) (cents.Cents, error) {
	p, err := n.pair(op, c, nil, other)
	if err != nil {
		return cents.Cents{}, err
	}

	if p.exact(nil) {
		return exactDistance(op, p.tx, n, other)
	}

	fx, fy, err := p.frequencies(op)
	if err != nil {
		return cents.Cents{}, err
	}

	d, err := cents.Between(fx, fy)
	if err != nil {
		return cents.Cents{}, depError(op, err)
	}

	return d, nil
}

// CentsTo returns the distance in cents from the sound of n to the sound of
// other: exactly, h(other) + c(other) - (h(n) + c(n)), when the tunings are
// equal; otherwise cents.Between of the frequencies, on the grid of 10⁻⁶
// cent. The errors are those of CompareSound.
func (n Note) CentsTo(other Note) (cents.Cents, error) {
	return n.centsTo("CentsTo", ownContext(), other)
}

// MustCentsTo is like CentsTo but panics with its error.
func (n Note) MustCentsTo(other Note) cents.Cents {
	return must(n.CentsTo(other))
}

// CentsToIn is like CentsTo in the settings of the explicit source src for
// both notes; the result is then always exact. A nil src reports ErrUnset.
func (n Note) CentsToIn(src settings.Source, other Note) (cents.Cents, error) {
	return n.centsTo("CentsToIn", sourceContext(src), other)
}

// MustCentsToIn is like CentsToIn but panics with its error.
func (n Note) MustCentsToIn(src settings.Source, other Note) cents.Cents {
	return must(n.CentsToIn(src, other))
}

// CentsToInTuning is like CentsTo in the explicit tuning t for both notes. A
// nil or unset t reports ErrUnset.
func (n Note) CentsToInTuning(t *tuning.Tuning, other Note) (cents.Cents, error) {
	return n.centsTo("CentsToInTuning", tuningContext(t), other)
}

// MustCentsToInTuning is like CentsToInTuning but panics with its error.
func (n Note) MustCentsToInTuning(t *tuning.Tuning, other Note) cents.Cents {
	return must(n.CentsToInTuning(t, other))
}

func (n Note) offset(op string, c context, crit *frequency.Criterion, f float64) (int, cents.Cents, error) {
	if err := checkNote(op, n); err != nil {
		return 0, cents.Cents{}, err
	}

	if err := c.check(op); err != nil {
		return 0, cents.Cents{}, err
	}

	if crit != nil && !crit.IsValid() {
		return 0, cents.Cents{}, unsetCriterionError(op)
	}

	if err := checkOctave(op, n); err != nil {
		return 0, cents.Cents{}, err
	}

	if !cents.IsValidFrequency(f) {
		return 0, cents.Cents{}, invalidFrequencyError(op, f)
	}

	fr, err := c.frame(newReader(op), n)
	if err != nil {
		return 0, cents.Cents{}, err
	}

	if err := admit(op, fr.sys, n.name); err != nil {
		return 0, cents.Cents{}, err
	}

	fn, err := n.sound(op, fr.tuning())
	if err != nil {
		return 0, cents.Cents{}, err
	}

	var r int
	if crit == nil {
		r, err = frequency.Compare(f, fn)
	} else {
		r, err = crit.Compare(f, fn)
	}

	if err != nil {
		return 0, cents.Cents{}, depError(op, err)
	}

	d, err := frequency.CentsBetween(fn, f)
	if err != nil {
		return 0, cents.Cents{}, depError(op, err)
	}

	return r, d, nil
}

// Offset tells where the frequency f in hertz is relative to the sound of n:
// -1 lower, 0 the same, +1 higher (frequency.Compare), and returns the
// deviation of f from the sound of n in cents (frequency.CentsBetween).
//
// Errors, in this order: ErrUnset for the note; ErrNoOctave; an invalid f
// (frequency.ErrInvalidFrequency and cents.ErrInvalidArgument, without a
// sentinel of note), checked before the source is read; ErrNoSource;
// ErrUnsetSettings; ErrInadmissible; ErrOutOfRange.
func (n Note) Offset(f float64) (int, cents.Cents, error) {
	return n.offset("Offset", ownContext(), nil, f)
}

// MustOffset is like Offset but panics with its error.
func (n Note) MustOffset(f float64) (int, cents.Cents) {
	return must3(n.Offset(f))
}

// OffsetIn is like Offset in the settings of the explicit source src; the
// binding of n does not change and its own source is not read. A nil src
// reports ErrUnset.
func (n Note) OffsetIn(src settings.Source, f float64) (int, cents.Cents, error) {
	return n.offset("OffsetIn", sourceContext(src), nil, f)
}

// MustOffsetIn is like OffsetIn but panics with its error.
func (n Note) MustOffsetIn(src settings.Source, f float64) (int, cents.Cents) {
	return must3(n.OffsetIn(src, f))
}

// OffsetInTuning is like Offset in the explicit tuning t. A nil or unset t
// reports ErrUnset.
func (n Note) OffsetInTuning(t *tuning.Tuning, f float64) (int, cents.Cents, error) {
	return n.offset("OffsetInTuning", tuningContext(t), nil, f)
}

// MustOffsetInTuning is like OffsetInTuning but panics with its error.
func (n Note) MustOffsetInTuning(t *tuning.Tuning, f float64) (int, cents.Cents) {
	return must3(n.OffsetInTuning(t, f))
}

// OffsetBy is like Offset with the criterion c of package frequency: 0 when
// f is within c of the sound of n (frequency.Criterion.Compare). An unset
// criterion reports ErrUnset, also frequency.ErrUnset.
func (n Note) OffsetBy(f float64, c frequency.Criterion) (int, cents.Cents, error) {
	return n.offset("OffsetBy", ownContext(), &c, f)
}

// MustOffsetBy is like OffsetBy but panics with its error.
func (n Note) MustOffsetBy(f float64, c frequency.Criterion) (int, cents.Cents) {
	return must3(n.OffsetBy(f, c))
}

// OffsetByIn is like OffsetBy in the settings of the explicit source src. A
// nil src reports ErrUnset.
func (n Note) OffsetByIn(src settings.Source, f float64, c frequency.Criterion) (int, cents.Cents, error) {
	return n.offset("OffsetByIn", sourceContext(src), &c, f)
}

// MustOffsetByIn is like OffsetByIn but panics with its error.
func (n Note) MustOffsetByIn(src settings.Source, f float64, c frequency.Criterion) (int, cents.Cents) {
	return must3(n.OffsetByIn(src, f, c))
}

// OffsetByInTuning is like OffsetBy in the explicit tuning t. A nil or unset t
// reports ErrUnset.
func (n Note) OffsetByInTuning(t *tuning.Tuning, f float64, c frequency.Criterion) (int, cents.Cents, error) {
	return n.offset("OffsetByInTuning", tuningContext(t), &c, f)
}

// MustOffsetByInTuning is like OffsetByInTuning but panics with its error.
func (n Note) MustOffsetByInTuning(t *tuning.Tuning, f float64, c frequency.Criterion) (int, cents.Cents) {
	return must3(n.OffsetByInTuning(t, f, c))
}
