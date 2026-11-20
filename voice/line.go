package voice

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/note"
)

// Line is a melodic line: a voice as a line of the texture. It has elements
// in order, sounds and rests following each other, a name and three roles.
// A line sounds at most one pitch at a time.
//
// Use a line through *Line. A line is immutable: the operations that change
// or build a line return a new *Line and never change the receiver, its
// elements or their arguments, both on success and on failure. Reads return
// copies. All methods are safe for concurrent use.
//
// nil and &Line{} are the unset line, which is not the empty line: create the
// empty line with New(). On an unset line only IsSet, the equalities and
// String work; the other methods report ErrUnset.
type Line struct {
	set        bool
	elems      []Element
	starts     []duration.Position
	end        duration.Position
	name       string
	placement  Placement
	thematic   ThematicRole
	importance Importance
}

// Pair is a note or a rest with a duration, an input of FromPairs. A nil
// Note is a rest. FromPairs reads the note once and keeps a copy.
type Pair struct {
	Note     *note.Note // nil for a rest
	Duration Duration
}

// SoundSpan is a sound of a line: the index of its element, a copy of its
// note, and its start and end in positions of the line.
type SoundSpan struct {
	Index      int
	Note       note.Note
	Start, End duration.Position
}

// Fermata is an element with a fermata: its index and the position of its
// end, the address of a hold of a tempo map.
type Fermata struct {
	Index int
	End   duration.Position
}

// newLine returns a set line of elems, which it owns, with the name and the
// roles of meta, computing the starts of the elements. A partial sum out of
// the range of package common/fraction reports the error of package
// duration, which also matches fraction.ErrOutOfRange.
func newLine(elems []Element, meta *Line) (*Line, error) {
	starts := make([]duration.Position, len(elems))
	pos := duration.Origin()

	for i, e := range elems {
		starts[i] = pos

		next, err := pos.Add(e.dur.length)
		if err != nil {
			return nil, fmt.Errorf("voice: start of element %d: %w", i+1, err)
		}

		pos = next
	}

	l := &Line{set: true, elems: elems, starts: starts, end: pos}
	if meta != nil {
		l.name = meta.name
		l.placement = meta.placement
		l.thematic = meta.thematic
		l.importance = meta.importance
	}

	return l, nil
}

// withSameRhythm returns a copy of l with elems, which have the kinds and the
// lengths of the elements of l in the same order, so the starts are shared.
func (l *Line) withSameRhythm(elems []Element) *Line {
	c := *l
	c.elems = elems

	return &c
}

// clone returns a shallow copy of l that shares its immutable data.
func (l *Line) clone() *Line {
	c := *l

	return &c
}

// New returns a set line of copies of elements in order, with the name ""
// and no roles. Without arguments it returns the empty line.
//
// Errors, in this order: ErrUnset for the first unset element, in an
// *ItemError with its index; a start out of the range of package
// common/fraction (duration.ErrOutOfRange with fraction.ErrOutOfRange,
// without an *ItemError).
func New(elements ...Element) (*Line, error) {
	for i, e := range elements {
		if !e.IsSet() {
			return nil, itemError(i, unsetError("element", nil))
		}
	}

	return newLine(append([]Element(nil), elements...), nil)
}

// MustNew is like New but panics with its error.
func MustNew(elements ...Element) *Line {
	return must(New(elements...))
}

// FromPairs returns a line of pairs in order without fermatas: a pair with a
// nil Note is a rest, other pairs are sounds with a copy of *Note.
//
// Errors: for each pair in order, the errors of NewRest or NewSound in an
// *ItemError with the index of the pair; then a start out of range, as in
// New, without an *ItemError.
func FromPairs(pairs ...Pair) (*Line, error) {
	elems := make([]Element, len(pairs))

	for i, p := range pairs {
		var (
			e   Element
			err error
		)

		if p.Note == nil {
			e, err = NewRest(p.Duration)
		} else {
			e, err = NewSound(*p.Note, p.Duration)
		}

		if err != nil {
			return nil, itemError(i, err)
		}

		elems[i] = e
	}

	return newLine(elems, nil)
}

// MustFromPairs is like FromPairs but panics with its error.
func MustFromPairs(pairs ...Pair) *Line {
	return must(FromPairs(pairs...))
}

// Join returns a line of the elements of lines in order, with the name and
// the roles of the first line. Without arguments it returns the empty line
// without a name and roles.
//
// Errors, in this order: ErrUnset for the first unset line, in an
// *ItemError with its index; a start out of range, as in New.
func Join(lines ...*Line) (*Line, error) {
	total := 0

	for i, l := range lines {
		if !l.IsSet() {
			return nil, itemError(i, unsetError("line", nil))
		}

		total += len(l.elems)
	}

	elems := make([]Element, 0, total)
	for _, l := range lines {
		elems = append(elems, l.elems...)
	}

	var meta *Line
	if len(lines) > 0 {
		meta = lines[0]
	}

	return newLine(elems, meta)
}

// MustJoin is like Join but panics with its error.
func MustJoin(lines ...*Line) *Line {
	return must(Join(lines...))
}

// IsSet reports whether l is a set line; false for nil and &Line{}.
func (l *Line) IsSet() bool {
	return l != nil && l.set
}

// Len returns the number of elements. An unset line reports ErrUnset.
func (l *Line) Len() (int, error) {
	if !l.IsSet() {
		return 0, errUnsetLine()
	}

	return len(l.elems), nil
}

// MustLen is like Len but panics with its error.
func (l *Line) MustLen() int {
	return must(l.Len())
}

// IsEmpty reports whether l has no elements. An unset line reports ErrUnset.
func (l *Line) IsEmpty() (bool, error) {
	if !l.IsSet() {
		return false, errUnsetLine()
	}

	return len(l.elems) == 0, nil
}

// MustIsEmpty is like IsEmpty but panics with its error.
func (l *Line) MustIsEmpty() bool {
	return must(l.IsEmpty())
}

// SoundCount returns the number of sounds; a tied sound counts once.
// An unset line reports ErrUnset.
func (l *Line) SoundCount() (int, error) {
	return l.count(KindSound)
}

// MustSoundCount is like SoundCount but panics with its error.
func (l *Line) MustSoundCount() int {
	return must(l.SoundCount())
}

// RestCount returns the number of rests. An unset line reports ErrUnset.
func (l *Line) RestCount() (int, error) {
	return l.count(KindRest)
}

// MustRestCount is like RestCount but panics with its error.
func (l *Line) MustRestCount() int {
	return must(l.RestCount())
}

func (l *Line) count(k Kind) (int, error) {
	if !l.IsSet() {
		return 0, errUnsetLine()
	}

	n := 0

	for _, e := range l.elems {
		if e.kind == k {
			n++
		}
	}

	return n, nil
}

// checkIndex reports ErrIndex unless 0 <= i < len(l.elems).
func (l *Line) checkIndex(i int) error {
	if i < 0 || i >= len(l.elems) {
		return indexError("element %d of %d", i, len(l.elems))
	}

	return nil
}

// checkGap reports ErrIndex unless 0 <= i <= len(l.elems).
func (l *Line) checkGap(i int) error {
	if i < 0 || i > len(l.elems) {
		return indexError("position %d of %d elements", i, len(l.elems))
	}

	return nil
}

// checkRange reports ErrIndex unless 0 <= i <= j <= len(l.elems).
func (l *Line) checkRange(i, j int) error {
	if i < 0 || j > len(l.elems) || i > j {
		return indexError("range [%d, %d) of %d elements", i, j, len(l.elems))
	}

	return nil
}

// Element returns a copy of the element i.
//
// Errors, in this order: ErrUnset for an unset line; ErrIndex unless
// 0 <= i < Len.
func (l *Line) Element(i int) (Element, error) {
	if !l.IsSet() {
		return Element{}, errUnsetLine()
	}

	if err := l.checkIndex(i); err != nil {
		return Element{}, err
	}

	return l.elems[i], nil
}

// MustElement is like Element but panics with its error.
func (l *Line) MustElement(i int) Element {
	return must(l.Element(i))
}

// Elements returns a new slice of copies of all elements in order; a non-nil
// empty slice for the empty line. An unset line reports ErrUnset.
func (l *Line) Elements() ([]Element, error) {
	if !l.IsSet() {
		return nil, errUnsetLine()
	}

	return append(make([]Element, 0, len(l.elems)), l.elems...), nil
}

// MustElements is like Elements but panics with its error.
func (l *Line) MustElements() []Element {
	return must(l.Elements())
}

// Notes returns a new list of copies of the notes of the sounds in order;
// a tied sound gives one note. An unset line reports ErrUnset.
func (l *Line) Notes() (note.Notes, error) {
	if !l.IsSet() {
		return nil, errUnsetLine()
	}

	notes := note.Notes{}

	for _, e := range l.elems {
		if e.kind == KindSound {
			notes = append(notes, e.note)
		}
	}

	return notes, nil
}

// MustNotes is like Notes but panics with its error.
func (l *Line) MustNotes() note.Notes {
	return must(l.Notes())
}

// Name returns the name of l, "" when it has none. An unset line reports
// ErrUnset.
func (l *Line) Name() (string, error) {
	if !l.IsSet() {
		return "", errUnsetLine()
	}

	return l.name, nil
}

// MustName is like Name but panics with its error.
func (l *Line) MustName() string {
	return must(l.Name())
}

// Placement returns the role by place; PlacementUnset when it is not set.
// An unset line reports ErrUnset.
func (l *Line) Placement() (Placement, error) {
	if !l.IsSet() {
		return PlacementUnset, errUnsetLine()
	}

	return l.placement, nil
}

// MustPlacement is like Placement but panics with its error.
func (l *Line) MustPlacement() Placement {
	return must(l.Placement())
}

// Thematic returns the thematic role; ThematicUnset when it is not set.
// An unset line reports ErrUnset.
func (l *Line) Thematic() (ThematicRole, error) {
	if !l.IsSet() {
		return ThematicUnset, errUnsetLine()
	}

	return l.thematic, nil
}

// MustThematic is like Thematic but panics with its error.
func (l *Line) MustThematic() ThematicRole {
	return must(l.Thematic())
}

// Importance returns the role by importance; ImportanceUnset when it is not
// set. An unset line reports ErrUnset.
func (l *Line) Importance() (Importance, error) {
	if !l.IsSet() {
		return ImportanceUnset, errUnsetLine()
	}

	return l.importance, nil
}

// MustImportance is like Importance but panics with its error.
func (l *Line) MustImportance() Importance {
	return must(l.Importance())
}

// Equal reports whether l and other have the same number of elements, equal
// elements in order (Element.Equal), the same name and the same three roles.
// Two unset lines are equal; an unset line equals no set line, not even the
// empty one. The sources of the notes do not take part.
func (l *Line) Equal(other *Line) bool {
	if !l.equalElementsBy(other, Element.Equal) {
		return false
	}

	if !l.IsSet() {
		return true
	}

	return l.name == other.name && l.placement == other.placement &&
		l.thematic == other.thematic && l.importance == other.importance
}

// EqualElements is like Equal without the name and the roles.
func (l *Line) EqualElements(other *Line) bool {
	return l.equalElementsBy(other, Element.Equal)
}

// EqualDuration is like EqualElements but compares the durations of the
// elements by value (Element.EqualDuration): a dotted quarter equals a
// quarter tied to an eighth.
func (l *Line) EqualDuration(other *Line) bool {
	return l.equalElementsBy(other, Element.EqualDuration)
}

func (l *Line) equalElementsBy(other *Line, equal func(a, b Element) bool) bool {
	if l.IsSet() != other.IsSet() {
		return false
	}

	if !l.IsSet() {
		return true
	}

	if len(l.elems) != len(other.elems) {
		return false
	}

	for i := range l.elems {
		if !equal(l.elems[i], other.elems[i]) {
			return false
		}
	}

	return true
}

// String returns a debugging form with the name, the roles and the elements,
// or "Line(unset)". It is not a notation and has no parser.
func (l *Line) String() string {
	if !l.IsSet() {
		return "Line(unset)"
	}

	var b strings.Builder

	b.WriteString("Line(")
	b.WriteString(strconv.Quote(l.name))
	b.WriteString("; ")
	b.WriteString(l.placement.String())
	b.WriteString(", ")
	b.WriteString(l.thematic.String())
	b.WriteString(", ")
	b.WriteString(l.importance.String())
	b.WriteString("; [")

	for i, e := range l.elems {
		if i > 0 {
			b.WriteString(", ")
		}

		b.WriteString(e.String())
	}

	b.WriteString("])")

	return b.String()
}
