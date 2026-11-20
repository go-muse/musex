package voice

import (
	"fmt"
	"sort"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/note"
)

// endOf returns the end of the element i: the start of the next element or
// the end of the line.
func (l *Line) endOf(i int) duration.Position {
	if i+1 < len(l.starts) {
		return l.starts[i+1]
	}

	return l.end
}

// shift returns the position s + p. A result out of the range of package
// common/fraction reports the error of package duration, which also matches
// fraction.ErrOutOfRange.
func shift(s, p duration.Position) (duration.Position, error) {
	f, err := p.Fraction()
	if err != nil {
		return duration.Position{}, fmt.Errorf("voice: position: %w", err)
	}

	if f.IsZero() {
		return s, nil
	}

	d, err := duration.FromFraction(f)
	if err != nil {
		return duration.Position{}, fmt.Errorf("voice: position: %w", err)
	}

	r, err := s.Add(d)
	if err != nil {
		return duration.Position{}, fmt.Errorf("voice: position %s + %s: %w", s, p, err)
	}

	return r, nil
}

// checkPosition reports ErrUnset with duration.ErrUnset for an unset t.
func checkPosition(t duration.Position, what string) error {
	if !t.IsSet() {
		return unsetError(what, duration.ErrUnset)
	}

	return nil
}

// StartOf returns the start of the element i: the sum of the lengths of the
// elements 0 to i-1.
//
// Errors, in this order: ErrUnset for an unset line; ErrIndex unless
// 0 <= i < Len.
func (l *Line) StartOf(i int) (duration.Position, error) {
	if !l.IsSet() {
		return duration.Position{}, errUnsetLine()
	}

	if err := l.checkIndex(i); err != nil {
		return duration.Position{}, err
	}

	return l.starts[i], nil
}

// MustStartOf is like StartOf but panics with its error.
func (l *Line) MustStartOf(i int) duration.Position {
	return must(l.StartOf(i))
}

// EndOf returns the end of the element i: its start plus its length.
//
// Errors, in this order: ErrUnset for an unset line; ErrIndex unless
// 0 <= i < Len.
func (l *Line) EndOf(i int) (duration.Position, error) {
	if !l.IsSet() {
		return duration.Position{}, errUnsetLine()
	}

	if err := l.checkIndex(i); err != nil {
		return duration.Position{}, err
	}

	return l.endOf(i), nil
}

// MustEndOf is like EndOf but panics with its error.
func (l *Line) MustEndOf(i int) duration.Position {
	return must(l.EndOf(i))
}

// Starts returns a new slice of the starts of all elements in order and the
// end of the line. An unset line reports ErrUnset.
func (l *Line) Starts() ([]duration.Position, duration.Position, error) {
	if !l.IsSet() {
		return nil, duration.Position{}, errUnsetLine()
	}

	return append(make([]duration.Position, 0, len(l.starts)), l.starts...), l.end, nil
}

// MustStarts is like Starts but panics with its error.
func (l *Line) MustStarts() ([]duration.Position, duration.Position) {
	return must2(l.Starts())
}

// Length returns the length of the line, the exact sum of the lengths of its
// elements, in reduced form; 0/1 for the empty line. An unset line reports
// ErrUnset.
func (l *Line) Length() (fraction.Fraction, error) {
	if !l.IsSet() {
		return fraction.Fraction{}, errUnsetLine()
	}

	f, err := l.end.Fraction()
	if err != nil {
		return fraction.Fraction{}, fmt.Errorf("voice: length: %w", err)
	}

	return f.Reduce(), nil
}

// MustLength is like Length but panics with its error.
func (l *Line) MustLength() fraction.Fraction {
	return must(l.Length())
}

// End returns the position of the end of the line, equal to its length;
// duration.Origin() for the empty line. An unset line reports ErrUnset.
func (l *Line) End() (duration.Position, error) {
	if !l.IsSet() {
		return duration.Position{}, errUnsetLine()
	}

	return l.end, nil
}

// MustEnd is like End but panics with its error.
func (l *Line) MustEnd() duration.Position {
	return must(l.End())
}

// at returns the index of the element with start <= t < end, if any.
func (l *Line) at(t duration.Position) (int, bool) {
	if len(l.elems) == 0 || t.Compare(l.end) >= 0 {
		return 0, false
	}

	k := sort.Search(len(l.starts), func(i int) bool {
		return l.starts[i].Compare(t) > 0
	})

	return k - 1, true
}

// At returns the index of the element with start <= t < end and true.
// A moment at the end of an element belongs to the next element. For t at or
// after the end of the line, and for the empty line, it returns 0 and false
// without an error. It takes O(log Len).
//
// Errors, in this order: ErrUnset for an unset line, then for an unset t
// (with duration.ErrUnset).
func (l *Line) At(t duration.Position) (int, bool, error) {
	if !l.IsSet() {
		return 0, false, errUnsetLine()
	}

	if err := checkPosition(t, "moment"); err != nil {
		return 0, false, err
	}

	k, ok := l.at(t)

	return k, ok, nil
}

// MustAt is like At but panics with its error.
func (l *Line) MustAt(t duration.Position) (int, bool) {
	return must2(l.At(t))
}

// SoundAt returns a copy of the note sounding at t and true. A rest and a
// moment outside the line are silence: note.Note{} and false without an
// error. The errors are those of At.
func (l *Line) SoundAt(t duration.Position) (note.Note, bool, error) {
	k, ok, err := l.At(t)
	if err != nil || !ok || l.elems[k].kind != KindSound {
		return note.Note{}, false, err
	}

	return l.elems[k].note, true, nil
}

// MustSoundAt is like SoundAt but panics with its error.
func (l *Line) MustSoundAt(t duration.Position) (note.Note, bool) {
	return must2(l.SoundAt(t))
}

// Span returns a new slice of the indexes of the elements that overlap the
// half-open interval [a, b) in order: start < b and end > a. For a = b it
// returns the element at the moment a, if any. After the end of the line it
// returns an empty slice.
//
// Errors, in this order: ErrUnset for an unset line, then for an unset a and
// an unset b (with duration.ErrUnset); ErrIndex when a > b.
func (l *Line) Span(a, b duration.Position) ([]int, error) {
	if !l.IsSet() {
		return nil, errUnsetLine()
	}

	if err := checkPosition(a, "span start"); err != nil {
		return nil, err
	}

	if err := checkPosition(b, "span end"); err != nil {
		return nil, err
	}

	if a.Compare(b) > 0 {
		return nil, indexError("span start %s after its end %s", a, b)
	}

	out := []int{}

	if a.Equal(b) {
		if k, ok := l.at(a); ok {
			out = append(out, k)
		}

		return out, nil
	}

	first := sort.Search(len(l.elems), func(i int) bool {
		return l.endOf(i).Compare(a) > 0
	})

	for i := first; i < len(l.elems) && l.starts[i].Compare(b) < 0; i++ {
		out = append(out, i)
	}

	return out, nil
}

// MustSpan is like Span but panics with its error.
func (l *Line) MustSpan(a, b duration.Position) []int {
	return must(l.Span(a, b))
}

// Sounds returns for each sound in order the index of its element, a copy of
// its note, its start and its end. A tied sound is one sound. An unset line
// reports ErrUnset.
func (l *Line) Sounds() ([]SoundSpan, error) {
	if !l.IsSet() {
		return nil, errUnsetLine()
	}

	out := []SoundSpan{}

	for i, e := range l.elems {
		if e.kind == KindSound {
			out = append(out, SoundSpan{Index: i, Note: e.note, Start: l.starts[i], End: l.endOf(i)})
		}
	}

	return out, nil
}

// MustSounds is like Sounds but panics with its error.
func (l *Line) MustSounds() []SoundSpan {
	return must(l.Sounds())
}

// Fermatas returns for each element with a fermata in order its index and
// the position of its end: the address of a hold of a tempo map that starts
// with the line. An unset line reports ErrUnset.
func (l *Line) Fermatas() ([]Fermata, error) {
	if !l.IsSet() {
		return nil, errUnsetLine()
	}

	return l.fermatasFrom(duration.Origin())
}

// MustFermatas is like Fermatas but panics with its error.
func (l *Line) MustFermatas() []Fermata {
	return must(l.Fermatas())
}

// FermatasFrom is like Fermatas for a line that starts at the position start
// of a tempo map: the ends are start + end.
//
// Errors, in this order: ErrUnset for an unset line, then for an unset start
// (with duration.ErrUnset); a position out of the range of package
// common/fraction (duration.ErrOutOfRange with fraction.ErrOutOfRange).
func (l *Line) FermatasFrom(start duration.Position) ([]Fermata, error) {
	if !l.IsSet() {
		return nil, errUnsetLine()
	}

	if err := checkPosition(start, "start of the line"); err != nil {
		return nil, err
	}

	return l.fermatasFrom(start)
}

// MustFermatasFrom is like FermatasFrom but panics with its error.
func (l *Line) MustFermatasFrom(start duration.Position) []Fermata {
	return must(l.FermatasFrom(start))
}

func (l *Line) fermatasFrom(start duration.Position) ([]Fermata, error) {
	out := []Fermata{}

	for i, e := range l.elems {
		if !e.fermata {
			continue
		}

		end, err := shift(start, l.endOf(i))
		if err != nil {
			return nil, err
		}

		out = append(out, Fermata{Index: i, End: end})
	}

	return out, nil
}
