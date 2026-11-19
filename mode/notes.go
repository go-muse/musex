package mode

import (
	"cmp"
	"errors"
	"fmt"
	"math/big"
	"slices"

	"github.com/go-muse/musex/degree"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/scale"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

// NotesFrom builds the notes of the actual degrees, preserving the correction
// and source of first. A note with an octave must have a settings source.
func (m Mode) NotesFrom(first note.Note) ([]note.Note, error) {
	return m.notesFrom(nil, first, false)
}

// NotesFromIn builds the same sequence in an explicit system.
func (m Mode) NotesFromIn(sys *system.System, first note.Note) ([]note.Note, error) {
	return m.notesFrom(sys, first, true)
}

func (m Mode) notesFrom(sys *system.System, first note.Note, explicit bool) (note.Notes, error) {
	if !m.IsValid() || !first.IsValid() || (explicit && !sys.IsValid()) {
		return nil, ErrUnset
	}
	result := make(note.Notes, m.Len())
	for i, prefix := range m.prefixes {
		n, err := transposeNote(sys, first, coordinatesOf(prefix))
		if err != nil {
			return nil, buildNoteError(fmt.Sprintf("degree %d", i+1), err)
		}
		result[i] = n
	}
	return result, nil
}

// DegreesFrom builds degrees with notes from first.
func (m Mode) DegreesFrom(first note.Note) ([]degree.Degree, error) {
	return degreesWithNotes(m.NotesFrom(first))
}

// DegreesFromIn builds degrees with notes from first in an explicit system.
func (m Mode) DegreesFromIn(sys *system.System, first note.Note) ([]degree.Degree, error) {
	return degreesWithNotes(m.NotesFromIn(sys, first))
}

func degreesWithNotes(notes note.Notes, err error) ([]degree.Degree, error) {
	if err != nil {
		return nil, err
	}
	result := make([]degree.Degree, len(notes))
	for i, n := range notes {
		result[i] = degree.MustNewWithNote(i+1, n)
	}
	return result, nil
}

// NoteFrom builds one degree's note at a signed cycle offset.
func (m Mode) NoteFrom(first note.Note, number, cycle int) (note.Note, error) {
	return m.noteFrom(nil, first, number, cycle, false)
}

// NoteFromIn builds one degree's note at a signed cycle offset in sys.
func (m Mode) NoteFromIn(sys *system.System, first note.Note, number, cycle int) (note.Note, error) {
	return m.noteFrom(sys, first, number, cycle, true)
}

func (m Mode) noteFrom(sys *system.System, first note.Note, number, cycle int, explicit bool) (note.Note, error) {
	if !first.IsValid() || (explicit && !sys.IsValid()) {
		return note.Note{}, ErrUnset
	}
	if err := m.checkDegrees(number); err != nil {
		return note.Note{}, err
	}
	if m.IsOpen() && cycle != 0 {
		return note.Note{}, ErrInvalidArgument
	}
	shift := coordinatesOf(m.prefixes[number-1])
	if cycle != 0 {
		shift = shift.add(coordinatesOf(m.cycle).mul(cycle))
	}
	n, err := transposeNote(sys, first, shift)
	if err != nil {
		return note.Note{}, buildNoteError("NoteFrom", err)
	}
	return n, nil
}

// transposeNote builds the final name and octave directly, without needing an
// Interval, a written octave, or an octave displacement to fit in an int.
func transposeNote(sys *system.System, first note.Note, shift coordinates) (note.Note, error) {
	explicit := sys != nil
	if first.HasOctave() && !explicit {
		if !first.IsBound() {
			return note.Note{}, interval.ErrNoSystem
		}
		set, err := first.Settings()
		if err != nil {
			return note.Note{}, err
		}
		sys = set.System()
	}
	start := first.Spelling()
	startCarry, err := noteCarry(sys, start, explicit)
	if err != nil {
		return note.Note{}, err
	}
	target, err := shift.transposeName(start)
	if err != nil {
		return note.Note{}, err
	}
	targetCarry, err := noteCarry(sys, target, explicit)
	if err != nil {
		return note.Note{}, err
	}
	if !first.HasOctave() {
		return note.FromSpelling(target, first.Cents(), first.Source())
	}

	octave, err := first.ASPN()
	if err != nil {
		return note.Note{}, err
	}
	// The source's written octave is its ASPN octave minus its carry. The
	// change in written octave is the number of boundary letters crossed.
	h := sys.Boundary().DistanceUp(start.Letter())
	letterSteps := new(big.Int).Add(shift.letters, big.NewInt(int64(h)))
	writtenDelta, _ := splitLetters(letterSteps)
	final := new(big.Int).SetInt64(int64(octave))
	final.Sub(final, big.NewInt(int64(startCarry)))
	final.Add(final, writtenDelta)
	final.Add(final, big.NewInt(int64(targetCarry)))
	if !final.IsInt64() {
		return note.Note{}, errors.Join(interval.ErrOutOfRange, system.ErrOutOfRange)
	}
	return note.FromASPN(target, int(final.Int64()), first.Cents(), first.Source())
}

func noteCarry(sys *system.System, name spelling.Spelling, explicit bool) (int, error) {
	if sys == nil {
		return 0, nil
	}
	carry, err := sys.Carry(name)
	if err != nil && !explicit && errors.Is(err, system.ErrInadmissible) {
		return 0, errors.Join(note.ErrInadmissible, err)
	}
	return carry, err
}

// Signature contains the index on the line of fifths and the signed names.
type Signature = scale.Signature

// KeySignatureFrom recognizes a standard key signature among the names
// generated from first. The interval of the cycle does not affect the result.
func (m Mode) KeySignatureFrom(first spelling.Spelling) (Signature, bool, error) {
	if !m.IsValid() || !first.IsValid() {
		return Signature{}, false, ErrUnset
	}
	if m.Len() != 7 {
		return Signature{}, false, nil
	}
	type namedFifth struct {
		name  spelling.Spelling
		fifth int
	}
	var names [7]namedFifth
	for i, p := range m.prefixes {
		name, err := interval.Transpose(first, p)
		if err != nil {
			return Signature{}, false, buildNoteError(fmt.Sprintf("degree %d", i+1), err)
		}
		names[i].name = name
	}
	for i := range names {
		q, err := interval.FifthsOf(names[i].name)
		if err != nil {
			if errors.Is(err, interval.ErrNotOnFifths) {
				return Signature{}, false, nil
			}
			return Signature{}, false, buildNoteError("KeySignatureFrom", err)
		}
		names[i].fifth = q
	}
	slices.SortFunc(names[:], func(a, b namedFifth) int { return cmp.Compare(a.fifth, b.fifth) })
	low := names[0].fifth
	for i, name := range names {
		if name.fifth != low+i {
			return Signature{}, false, nil
		}
	}
	result := Signature{Index: low + 1, Signs: make([]spelling.Spelling, 0, 7)}
	for _, pair := range names {
		if pair.name.Alteration().Sign() != 0 {
			result.Signs = append(result.Signs, pair.name)
		}
	}
	slices.SortFunc(result.Signs, func(a, b spelling.Spelling) int {
		return cmp.Compare(interval.MustFifthsOf(a.Natural()), interval.MustFifthsOf(b.Natural()))
	})
	if result.Index < 0 {
		slices.Reverse(result.Signs)
	}
	return result, true, nil
}

// FirstSpellingForSignature returns the first name of a diatonic mode for
// the given key signature index.
func (m Mode) FirstSpellingForSignature(index int) (spelling.Spelling, error) {
	rank, ok, err := m.Brightness()
	if err != nil {
		return spelling.Spelling{}, err
	}
	if !ok {
		return spelling.Spelling{}, ErrInvalidArgument
	}
	fifth := new(big.Int).Add(big.NewInt(int64(index)), big.NewInt(int64(rank-2)))
	if !fifth.IsInt64() {
		return spelling.Spelling{}, errors.Join(ErrOutOfRange, interval.ErrOutOfRange)
	}
	i, err := interval.FromFifths(int(fifth.Int64()))
	if err != nil {
		return spelling.Spelling{}, positionError("FirstSpellingForSignature", err)
	}
	name, err := interval.Transpose(spelling.MustParse("C"), i)
	if err != nil {
		return spelling.Spelling{}, positionError("FirstSpellingForSignature", err)
	}
	return name, nil
}
