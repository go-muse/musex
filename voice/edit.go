package voice

import (
	"github.com/go-muse/musex/note"
)

// checkElements reports ErrUnset in an *ItemError with the index in the input
// list for the first unset element.
func checkElements(elements []Element) error {
	for i, e := range elements {
		if !e.IsSet() {
			return itemError(i, unsetError("element", nil))
		}
	}

	return nil
}

// spliced returns a new slice of the elements of l with the elements [i, j)
// replaced by insert.
func (l *Line) spliced(i, j int, insert []Element) []Element {
	out := make([]Element, 0, len(l.elems)-(j-i)+len(insert))
	out = append(out, l.elems[:i]...)
	out = append(out, insert...)

	return append(out, l.elems[j:]...)
}

// Append returns the line with elements added after its elements.
//
// Errors, in this order: ErrUnset for an unset line, then for the first unset
// element, in an *ItemError with its index in elements; a start out of the
// range of package common/fraction (duration.ErrOutOfRange with
// fraction.ErrOutOfRange).
func (l *Line) Append(elements ...Element) (*Line, error) {
	if !l.IsSet() {
		return nil, errUnsetLine()
	}

	return l.Insert(len(l.elems), elements...)
}

// MustAppend is like Append but panics with its error.
func (l *Line) MustAppend(elements ...Element) *Line {
	return must(l.Append(elements...))
}

// Insert returns the line with elements inserted before the element i;
// i = Len inserts them at the end.
//
// Errors, in this order: ErrUnset for an unset line; ErrIndex unless
// 0 <= i <= Len; the errors of the elements and the overflow, as in Append.
func (l *Line) Insert(i int, elements ...Element) (*Line, error) {
	if !l.IsSet() {
		return nil, errUnsetLine()
	}

	if err := l.checkGap(i); err != nil {
		return nil, err
	}

	if err := checkElements(elements); err != nil {
		return nil, err
	}

	return newLine(l.spliced(i, i, elements), l)
}

// MustInsert is like Insert but panics with its error.
func (l *Line) MustInsert(i int, elements ...Element) *Line {
	return must(l.Insert(i, elements...))
}

// Replace returns the line with the element i replaced by e.
//
// Errors, in this order: ErrUnset for an unset line; ErrIndex unless
// 0 <= i < Len; ErrUnset for an unset e; the overflow, as in Append.
func (l *Line) Replace(i int, e Element) (*Line, error) {
	if !l.IsSet() {
		return nil, errUnsetLine()
	}

	if err := l.checkIndex(i); err != nil {
		return nil, err
	}

	if !e.IsSet() {
		return nil, unsetError("element", nil)
	}

	return newLine(l.spliced(i, i+1, []Element{e}), l)
}

// MustReplace is like Replace but panics with its error.
func (l *Line) MustReplace(i int, e Element) *Line {
	return must(l.Replace(i, e))
}

// Remove returns the line without the element i.
//
// Errors, in this order: ErrUnset for an unset line; ErrIndex unless
// 0 <= i < Len; the overflow, as in Append.
func (l *Line) Remove(i int) (*Line, error) {
	if !l.IsSet() {
		return nil, errUnsetLine()
	}

	if err := l.checkIndex(i); err != nil {
		return nil, err
	}

	return newLine(l.spliced(i, i+1, nil), l)
}

// MustRemove is like Remove but panics with its error.
func (l *Line) MustRemove(i int) *Line {
	return must(l.Remove(i))
}

// RemoveRange returns the line without the elements [i, j); for i = j it has
// the same elements.
//
// Errors, in this order: ErrUnset for an unset line; ErrIndex unless
// 0 <= i <= j <= Len; the overflow, as in Append.
func (l *Line) RemoveRange(i, j int) (*Line, error) {
	if !l.IsSet() {
		return nil, errUnsetLine()
	}

	if err := l.checkRange(i, j); err != nil {
		return nil, err
	}

	return newLine(l.spliced(i, j, nil), l)
}

// MustRemoveRange is like RemoveRange but panics with its error.
func (l *Line) MustRemoveRange(i, j int) *Line {
	return must(l.RemoveRange(i, j))
}

// Slice returns the line of the elements [i, j) with the name and the roles
// of l.
//
// Errors, in this order: ErrUnset for an unset line; ErrIndex unless
// 0 <= i <= j <= Len; the overflow, as in Append.
func (l *Line) Slice(i, j int) (*Line, error) {
	if !l.IsSet() {
		return nil, errUnsetLine()
	}

	if err := l.checkRange(i, j); err != nil {
		return nil, err
	}

	return newLine(append([]Element(nil), l.elems[i:j]...), l)
}

// MustSlice is like Slice but panics with its error.
func (l *Line) MustSlice(i, j int) *Line {
	return must(l.Slice(i, j))
}

// Split returns the lines of the elements [0, i) and [i, Len), both with the
// name and the roles of l.
//
// Errors, in this order: ErrUnset for an unset line; ErrIndex unless
// 0 <= i <= Len; the overflow, as in Append.
func (l *Line) Split(i int) (*Line, *Line, error) {
	if !l.IsSet() {
		return nil, nil, errUnsetLine()
	}

	if err := l.checkGap(i); err != nil {
		return nil, nil, err
	}

	head, err := newLine(append([]Element(nil), l.elems[:i]...), l)
	if err != nil {
		return nil, nil, err
	}

	tail, err := newLine(append([]Element(nil), l.elems[i:]...), l)
	if err != nil {
		return nil, nil, err
	}

	return head, tail, nil
}

// MustSplit is like Split but panics with its error.
func (l *Line) MustSplit(i int) (*Line, *Line) {
	return must2(l.Split(i))
}

// WithName returns the line with the name, any string including "". It does
// not change the elements. An unset line reports ErrUnset.
func (l *Line) WithName(name string) (*Line, error) {
	if !l.IsSet() {
		return nil, errUnsetLine()
	}

	out := l.clone()
	out.name = name

	return out, nil
}

// MustWithName is like WithName but panics with its error.
func (l *Line) MustWithName(name string) *Line {
	return must(l.WithName(name))
}

// WithPlacement returns the line with the role by place p; PlacementUnset
// removes the role.
//
// Errors, in this order: ErrUnset for an unset line, then for p outside the
// list.
func (l *Line) WithPlacement(p Placement) (*Line, error) {
	if !l.IsSet() {
		return nil, errUnsetLine()
	}

	if p != PlacementUnset && !p.IsValid() {
		return nil, unsetError("placement "+p.String(), nil)
	}

	out := l.clone()
	out.placement = p

	return out, nil
}

// MustWithPlacement is like WithPlacement but panics with its error.
func (l *Line) MustWithPlacement(p Placement) *Line {
	return must(l.WithPlacement(p))
}

// WithThematic returns the line with the thematic role r; ThematicUnset
// removes the role.
//
// Errors, in this order: ErrUnset for an unset line, then for r outside the
// list.
func (l *Line) WithThematic(r ThematicRole) (*Line, error) {
	if !l.IsSet() {
		return nil, errUnsetLine()
	}

	if r != ThematicUnset && !r.IsValid() {
		return nil, unsetError("thematic role "+r.String(), nil)
	}

	out := l.clone()
	out.thematic = r

	return out, nil
}

// MustWithThematic is like WithThematic but panics with its error.
func (l *Line) MustWithThematic(r ThematicRole) *Line {
	return must(l.WithThematic(r))
}

// WithImportance returns the line with the role by importance r;
// ImportanceUnset removes the role.
//
// Errors, in this order: ErrUnset for an unset line, then for r outside the
// list.
func (l *Line) WithImportance(r Importance) (*Line, error) {
	if !l.IsSet() {
		return nil, errUnsetLine()
	}

	if r != ImportanceUnset && !r.IsValid() {
		return nil, unsetError("importance "+r.String(), nil)
	}

	out := l.clone()
	out.importance = r

	return out, nil
}

// MustWithImportance is like WithImportance but panics with its error.
func (l *Line) MustWithImportance(r Importance) *Line {
	return must(l.WithImportance(r))
}

// withElement returns the line with the element i changed by f.
func (l *Line) withElement(i int, f func(e Element) (Element, error)) (*Line, error) {
	if !l.IsSet() {
		return nil, errUnsetLine()
	}

	if err := l.checkIndex(i); err != nil {
		return nil, err
	}

	e, err := f(l.elems[i])
	if err != nil {
		return nil, err
	}

	elems := l.spliced(i, i+1, []Element{e})
	if e.dur.length.Equal(l.elems[i].dur.length) {
		return l.withSameRhythm(elems), nil
	}

	return newLine(elems, l)
}

// WithNote returns the line with the element i changed by Element.WithNote:
// a rest becomes a sound.
//
// Errors, in this order: ErrUnset for an unset line; ErrIndex unless
// 0 <= i < Len; the errors of Element.WithNote.
func (l *Line) WithNote(i int, n note.Note) (*Line, error) {
	return l.withElement(i, func(e Element) (Element, error) { return e.WithNote(n) })
}

// MustWithNote is like WithNote but panics with its error.
func (l *Line) MustWithNote(i int, n note.Note) *Line {
	return must(l.WithNote(i, n))
}

// WithDuration returns the line with the element i changed by
// Element.WithDuration.
//
// Errors, in this order: ErrUnset for an unset line; ErrIndex unless
// 0 <= i < Len; the errors of Element.WithDuration; the overflow, as in
// Append.
func (l *Line) WithDuration(i int, d Duration) (*Line, error) {
	return l.withElement(i, func(e Element) (Element, error) { return e.WithDuration(d) })
}

// MustWithDuration is like WithDuration but panics with its error.
func (l *Line) MustWithDuration(i int, d Duration) *Line {
	return must(l.WithDuration(i, d))
}

// WithFermata returns the line with the element i changed by
// Element.WithFermata.
//
// Errors, in this order: ErrUnset for an unset line; ErrIndex unless
// 0 <= i < Len.
func (l *Line) WithFermata(i int, on bool) (*Line, error) {
	return l.withElement(i, func(e Element) (Element, error) { return e.WithFermata(on) })
}

// MustWithFermata is like WithFermata but panics with its error.
func (l *Line) MustWithFermata(i int, on bool) *Line {
	return must(l.WithFermata(i, on))
}
