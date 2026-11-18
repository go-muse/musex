package mode

import (
	"errors"
	"math/big"
	"unsafe"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/scale"
	"github.com/go-muse/musex/system"
)

const maxScaleNotes = int(^uint(0)>>1) / int(unsafe.Sizeof(note.Note{}))

// Start addresses one actual degree and a signed structural cycle.
type Start struct {
	Index int
	Cycle int
}

// BuildOptions selects the first note, direction, and output octave policy.
type BuildOptions struct {
	Start        Start
	Direction    scale.Direction
	StripOctaves bool
}

var (
	ErrIndex             = errors.New("mode: index out of range")
	ErrInsufficientNotes = errors.New("mode: insufficient mode notes")
	ErrNonMonotone       = errors.New("mode: non-monotone mode expansion")
)

func (m Mode) scaleInputs(sys *system.System, tonic note.Note, explicit bool) error {
	if !m.IsValid() {
		return ErrUnset
	}
	if explicit && !sys.IsValid() {
		return errors.Join(ErrUnset, system.ErrInvalid)
	}
	if !tonic.IsValid() {
		return errors.Join(ErrUnset, note.ErrUnset)
	}
	return nil
}

// FromTonic builds one independent note per actual degree.
func (m Mode) FromTonic(tonic note.Note) (*scale.Scale, error) { return m.fromTonic(nil, tonic, false) }

// FromTonicIn builds one note per degree in an explicit musical system.
func (m Mode) FromTonicIn(sys *system.System, tonic note.Note) (*scale.Scale, error) {
	return m.fromTonic(sys, tonic, true)
}
func (m Mode) fromTonic(sys *system.System, tonic note.Note, explicit bool) (*scale.Scale, error) {
	if err := m.scaleInputs(sys, tonic, explicit); err != nil {
		return nil, err
	}
	ns, err := m.notesFrom(sys, tonic, explicit)
	if err != nil {
		return nil, err
	}
	return scale.New(ns...)
}
func (m Mode) startCheck(options BuildOptions) error {
	if options.Start.Index < 0 || options.Start.Index >= m.Len() {
		return ErrIndex
	}
	if m.IsOpen() && options.Start.Cycle != 0 {
		return ErrInvalidArgument
	}
	return nil
}

type scaleCursor struct {
	index int
	shift coordinates
}

func (m Mode) scaleCursor(start Start) scaleCursor {
	shift := coordinatesOf(m.prefixes[start.Index])
	if m.IsClosed() {
		shift = shift.add(coordinatesOf(m.cycle).mul(start.Cycle))
	}
	return scaleCursor{index: start.Index, shift: shift}
}
func (m Mode) advance(cursor *scaleCursor, direction scale.Direction, adjacent []coordinates) bool {
	if direction == scale.Forward {
		if m.IsOpen() && cursor.index == m.Len()-1 {
			return false
		}
		cursor.shift = cursor.shift.add(adjacent[cursor.index])
		cursor.index = (cursor.index + 1) % m.Len()
	} else {
		if m.IsOpen() && cursor.index == 0 {
			return false
		}
		cursor.index = (cursor.index + m.Len() - 1) % m.Len()
		cursor.shift = cursor.shift.sub(adjacent[cursor.index])
	}
	return true
}
func (m Mode) scaleAdjacent() []coordinates {
	intervals := m.AdjacentIntervals()
	result := make([]coordinates, len(intervals))
	for i, value := range intervals {
		result[i] = coordinatesOf(value)
	}
	return result
}
func finishScale(ns []note.Note, strip bool) (*scale.Scale, error) {
	if strip {
		for i := range ns {
			if err := ns[i].RemoveOctave(); err != nil {
				return nil, err
			}
		}
	}
	return scale.New(ns...)
}

// FromTonicCount builds exactly count notes in structural order.
func (m Mode) FromTonicCount(tonic note.Note, count int, options BuildOptions) (*scale.Scale, error) {
	return m.fromTonicCount(nil, tonic, count, options, false)
}
func (m Mode) FromTonicCountIn(sys *system.System, tonic note.Note, count int, options BuildOptions) (*scale.Scale, error) {
	return m.fromTonicCount(sys, tonic, count, options, true)
}
func (m Mode) fromTonicCount(sys *system.System, tonic note.Note, count int, options BuildOptions, explicit bool) (*scale.Scale, error) {
	if err := m.scaleInputs(sys, tonic, explicit); err != nil {
		return nil, err
	}
	if count < 0 || !options.Direction.IsValid() {
		return nil, ErrInvalidArgument
	}
	if count == 0 {
		return scale.New()
	}
	if err := m.startCheck(options); err != nil {
		return nil, err
	}
	available := m.Len() - options.Start.Index
	if options.Direction == scale.Backward {
		available = options.Start.Index + 1
	}
	if m.IsOpen() && count > available {
		return nil, ErrInsufficientNotes
	}
	cursor := m.scaleCursor(options.Start)
	adjacent := m.scaleAdjacent()
	if count > maxScaleNotes {
		return nil, ErrOutOfRange
	}
	ns := make([]note.Note, count)
	for i := range ns {
		n, err := transposeNote(sys, tonic, cursor.shift)
		if err != nil {
			return nil, buildNoteError("FromTonicCount", err)
		}
		ns[i] = n
		if i+1 < count {
			m.advance(&cursor, options.Direction, adjacent)
		}
	}
	return finishScale(ns, options.StripOctaves)
}

// FromTonicTo builds the notes inside an inclusive pitch range.
func (m Mode) FromTonicTo(tonic note.Note, end pitch.Pitch, options BuildOptions) (*scale.Scale, error) {
	return m.fromTonicRange(nil, tonic, end, 0, options, false, false)
}
func (m Mode) FromTonicToIn(sys *system.System, tonic note.Note, end pitch.Pitch, options BuildOptions) (*scale.Scale, error) {
	return m.fromTonicRange(sys, tonic, end, 0, options, true, false)
}

// FromTonicOctaves bounds the expansion by octaves from the selected start.
func (m Mode) FromTonicOctaves(tonic note.Note, octaves int, options BuildOptions) (*scale.Scale, error) {
	return m.fromTonicRange(nil, tonic, pitch.Pitch{}, octaves, options, false, true)
}
func (m Mode) FromTonicOctavesIn(sys *system.System, tonic note.Note, octaves int, options BuildOptions) (*scale.Scale, error) {
	return m.fromTonicRange(sys, tonic, pitch.Pitch{}, octaves, options, true, true)
}

// shiftedPosition evaluates the mathematical position before constructing a
// spelling. It can safely reject a trial note beyond the requested range.
func shiftedPosition(sys *system.System, tonic note.Note, shift coordinates) (*big.Rat, error) {
	start := tonic.Spelling()
	carry, err := sys.Carry(start)
	if err != nil {
		return nil, err
	}
	octave, err := tonic.ASPN()
	if err != nil {
		return nil, err
	}
	letters := new(big.Int).Add(shift.letters, big.NewInt(int64(sys.Boundary().DistanceUp(start.Letter()))))
	delta, _ := splitLetters(letters)
	written := new(big.Int).Sub(big.NewInt(int64(octave)), big.NewInt(int64(carry)))
	written.Add(written, delta).Mul(written, big.NewInt(int64(sys.Classes())))
	_, rem := splitLetters(shift.letters)
	target := start.Letter().Shift(rem)
	alteration := new(big.Rat).Add(rational(start.Alteration()), shift.chromatic)
	alteration.Sub(alteration, naturalDistance(shift.letters, start.Letter().Ordinal()))
	alteration.Mul(alteration, new(big.Rat).SetInt64(int64(sys.Chromatic())))
	alteration.Add(alteration, new(big.Rat).SetInt64(int64(sys.MustOffset(target))))
	return alteration.Add(alteration, new(big.Rat).SetInt(written)), nil
}
func pitchPosition(sys *system.System, p pitch.Pitch) (*big.Rat, error) {
	c, err := sys.Class(p.Spelling())
	if err != nil {
		return nil, err
	}
	octave, err := p.ASPN()
	if err != nil {
		return nil, err
	}
	total := new(big.Int).Mul(big.NewInt(int64(octave)), big.NewInt(int64(sys.Classes())))
	total.Add(total, big.NewInt(int64(c.Number())))
	return new(big.Rat).SetInt(total), nil
}
func (m Mode) checkExpansion(sys *system.System, tonic note.Note, adjacent []coordinates) error {
	if m.IsOpen() {
		return nil
	}
	cursor := m.scaleCursor(Start{})
	previous, err := shiftedPosition(sys, tonic, cursor.shift)
	if err != nil {
		return err
	}
	// Seven cycles include every possible starting-letter phase.
	for range 7 * m.Len() {
		m.advance(&cursor, scale.Forward, adjacent)
		next, err := shiftedPosition(sys, tonic, cursor.shift)
		if err != nil {
			return err
		}
		if next.Cmp(previous) <= 0 {
			return ErrNonMonotone
		}
		previous = next
	}
	return nil
}

// expansionSize counts each residue of the finite letter-phase period with
// exact arithmetic, rejecting an unrepresentable result before allocation.
func (m Mode) expansionSize(sys *system.System, tonic note.Note, start Start, end *big.Rat, direction scale.Direction, adjacent []coordinates) error {
	cursor := m.scaleCursor(start)
	period := coordinatesOf(m.cycle).mul(7)
	first, err := shiftedPosition(sys, tonic, cursor.shift)
	if err != nil {
		return err
	}
	shifted, err := shiftedPosition(sys, tonic, cursor.shift.add(period))
	if err != nil {
		return err
	}
	delta := new(big.Rat).Sub(shifted, first)
	total := new(big.Int)
	limit := new(big.Int).SetUint64(uint64(maxScaleNotes))
	for range 7 * m.Len() {
		at, err := shiftedPosition(sys, tonic, cursor.shift)
		if err != nil {
			return err
		}
		remaining := new(big.Rat).Sub(end, at)
		if remaining.Sign()*int(direction) >= 0 {
			quotient := new(big.Rat).Quo(remaining, delta)
			if direction == scale.Backward {
				quotient.Neg(quotient)
			}
			count := new(big.Int).Quo(quotient.Num(), quotient.Denom())
			count.Add(count, big.NewInt(1))
			total.Add(total, count)
			if total.Cmp(limit) > 0 {
				return ErrOutOfRange
			}
		}
		m.advance(&cursor, direction, adjacent)
	}
	return nil
}

func (m Mode) fromTonicRange(sys *system.System, tonic note.Note, end pitch.Pitch, octaves int, options BuildOptions, explicit, byOctaves bool) (*scale.Scale, error) {
	if err := m.scaleInputs(sys, tonic, explicit); err != nil {
		return nil, err
	}
	if !byOctaves && !end.IsValid() {
		return nil, errors.Join(ErrUnset, pitch.ErrInvalid)
	}
	if !options.Direction.IsValid() || byOctaves && octaves <= 0 {
		return nil, ErrInvalidArgument
	}
	if err := m.startCheck(options); err != nil {
		return nil, err
	}
	if !tonic.HasOctave() {
		return nil, note.ErrNoOctave
	}
	if !explicit {
		if !tonic.IsBound() {
			return nil, buildNoteError("FromTonicRange", interval.ErrNoSystem)
		}
		set, err := tonic.Settings()
		if err != nil {
			return nil, buildNoteError("FromTonicRange", err)
		}
		sys = set.System()
	}
	cursor := m.scaleCursor(options.Start)
	first, err := transposeNote(sys, tonic, cursor.shift)
	if err != nil {
		return nil, buildNoteError("FromTonicRange", err)
	}
	if byOctaves {
		boundary := first
		shift := octaves
		if options.Direction == scale.Backward {
			shift = -octaves
		}
		if err := boundary.Shift(shift); err != nil {
			return nil, buildNoteError("FromTonicOctaves", err)
		}
		end = boundary.MustPitch()
	}
	startPosition, err := pitchPosition(sys, first.MustPitch())
	if err != nil {
		return nil, buildNoteError("FromTonicRange", err)
	}
	endPosition, err := pitchPosition(sys, end)
	if err != nil {
		return nil, buildNoteError("FromTonicRange", err)
	}
	sign := int(options.Direction)
	if endPosition.Cmp(startPosition)*sign < 0 {
		return nil, ErrInvalidArgument
	}
	adjacent := m.scaleAdjacent()
	if err := m.checkExpansion(sys, tonic, adjacent); err != nil {
		return nil, err
	}
	if m.IsClosed() {
		if err := m.expansionSize(sys, tonic, options.Start, endPosition, options.Direction, adjacent); err != nil {
			return nil, err
		}
	}
	ns := make([]note.Note, 0)
	for {
		position, err := shiftedPosition(sys, tonic, cursor.shift)
		if err != nil {
			return nil, buildNoteError("FromTonicRange", err)
		}
		pastEnd := position.Cmp(endPosition)*sign > 0
		if pastEnd && m.IsClosed() {
			break
		}
		if !pastEnd && position.Cmp(startPosition)*sign >= 0 {
			n, err := transposeNote(sys, tonic, cursor.shift)
			if err != nil {
				return nil, buildNoteError("FromTonicRange", err)
			}
			if len(ns) == maxScaleNotes {
				return nil, ErrOutOfRange
			}
			ns = append(ns, n)
		}
		if !m.advance(&cursor, options.Direction, adjacent) {
			break
		}
	}
	return finishScale(ns, options.StripOctaves)
}

// MustFromTonic calls FromTonic and panics with its error.
func (m Mode) MustFromTonic(tonic note.Note) *scale.Scale { return must(m.FromTonic(tonic)) }

// MustFromTonicIn calls FromTonicIn and panics with its error.
func (m Mode) MustFromTonicIn(sys *system.System, tonic note.Note) *scale.Scale {
	return must(m.FromTonicIn(sys, tonic))
}

// MustFromTonicCount calls FromTonicCount and panics with its error.
func (m Mode) MustFromTonicCount(tonic note.Note, count int, options BuildOptions) *scale.Scale {
	return must(m.FromTonicCount(tonic, count, options))
}

// MustFromTonicCountIn calls FromTonicCountIn and panics with its error.
func (m Mode) MustFromTonicCountIn(sys *system.System, tonic note.Note, count int, options BuildOptions) *scale.Scale {
	return must(m.FromTonicCountIn(sys, tonic, count, options))
}

// MustFromTonicTo calls FromTonicTo and panics with its error.
func (m Mode) MustFromTonicTo(tonic note.Note, end pitch.Pitch, options BuildOptions) *scale.Scale {
	return must(m.FromTonicTo(tonic, end, options))
}

// MustFromTonicToIn calls FromTonicToIn and panics with its error.
func (m Mode) MustFromTonicToIn(sys *system.System, tonic note.Note, end pitch.Pitch, options BuildOptions) *scale.Scale {
	return must(m.FromTonicToIn(sys, tonic, end, options))
}

// MustFromTonicOctaves calls FromTonicOctaves and panics with its error.
func (m Mode) MustFromTonicOctaves(tonic note.Note, octaves int, options BuildOptions) *scale.Scale {
	return must(m.FromTonicOctaves(tonic, octaves, options))
}

// MustFromTonicOctavesIn calls FromTonicOctavesIn and panics with its error.
func (m Mode) MustFromTonicOctavesIn(sys *system.System, tonic note.Note, octaves int, options BuildOptions) *scale.Scale {
	return must(m.FromTonicOctavesIn(sys, tonic, octaves, options))
}
