package duration

import (
	"strconv"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/tuplet"
)

// Content is one item of a group's contents: a written duration or a nested
// group. Build one with [WrittenContent] or [GroupContent]; the zero value is
// not valid content.
//
// The contents are written durations: whether a value is played as a note or
// a rest belongs to package event, not to duration.
type Content struct {
	written Written
	group   *Group
}

// WrittenContent returns the content of the written duration w.
func WrittenContent(w Written) Content { return Content{written: w} }

// GroupContent returns the content of a copy of the nested group g.
func GroupContent(g Group) Content {
	if !g.IsSet() {
		return Content{}
	}
	return Content{group: g.clone()}
}

// IsValid reports whether c is a written duration or a set nested group.
func (c Content) IsValid() bool {
	if c.group != nil {
		return c.group.IsSet()
	}
	return c.written.IsSet()
}

// IsGroup reports whether c is a nested group rather than a written duration.
func (c Content) IsGroup() bool { return c.group != nil }

// Written returns the written duration of c. It returns an error wrapping
// ErrInvalidArgument if c is a nested group or is not valid.
func (c Content) Written() (Written, error) {
	if c.group != nil || !c.written.IsSet() {
		return Written{}, invalidError("Content.Written", "content is not a written duration")
	}
	return c.written, nil
}

// MustWritten is like Written but panics with the returned error.
func (c Content) MustWritten() Written { return must(c.Written()) }

// Group returns a copy of the nested group of c. It returns an error wrapping
// ErrInvalidArgument if c is a written duration or is not valid.
func (c Content) Group() (Group, error) {
	if c.group == nil || !c.group.IsSet() {
		return Group{}, invalidError("Content.Group", "content is not a group")
	}
	return *c.group.clone(), nil
}

// MustGroup is like Group but panics with the returned error.
func (c Content) MustGroup() Group { return must(c.Group()) }

// sounding returns the sounding duration of a valid content.
func (c Content) sounding() fraction.Fraction {
	if c.group != nil {
		return c.group.sounding
	}
	return c.written.value
}

// clone returns a deep copy of a valid content.
func (c Content) clone() Content {
	if c.group != nil {
		return Content{group: c.group.clone()}
	}
	return Content{written: c.written}
}

// Group is a written tuplet group: a tuplet, a written unit, and contents
// whose sounding durations fill at most the group's span. The span is
// normal·unit, the total sounding duration of the group; the shortfall is the
// part of the span the contents do not fill. Adjacent and nested groups stay
// distinguishable: one 6:4 group is not two 3:2 groups over the same notes.
//
// Build a group with [NewGroup]; the zero value is the unset group. A Group
// is immutable; copies made by assignment share contents only for reading,
// and the accessors return owned copies.
type Group struct {
	t        tuplet.Tuplet
	unit     Written
	contents []Content
	span     fraction.Fraction
	sounding fraction.Fraction
}

// NewGroup returns the group of the tuplet t in the written unit, with the
// given contents. The contents are given under the chain of the group: the
// chain of the unit extended by t. Every written content must carry exactly
// that chain, and the unit of every nested group must carry it too; anything
// else is rejected, so a group labeled 3:2 cannot hold values written outside
// 3:2 or under a different tuplet.
//
// The span of the group is t.Normal() times the value of the unit; the
// sounding durations of the contents must not exceed it. An empty content
// list is allowed: the shortfall is then the whole span.
//
// In order of checking, it returns an error wrapping ErrInvalidArgument
// together with tuplet.ErrUnset if t is unset, ErrUnset if the unit is unset,
// ErrOutOfRange if the span or a sounding sum is outside the range of package
// common/fraction, ErrInvalidArgument if a content is not valid or its chain
// does not match the group, and ErrInvalidArgument if the sounding sum of the
// contents exceeds the span. The contents are copied; later changes to them
// do not affect the group.
func NewGroup(t tuplet.Tuplet, unit Written, contents ...Content) (Group, error) {
	return newGroup("NewGroup", t, unit, contents, true)
}

// newGroup applies the checks of NewGroup for op and returns the group. With
// copyContents false the group takes the contents slice without copying it;
// the caller then guarantees that neither the slice nor its groups change.
func newGroup(op string, t tuplet.Tuplet, unit Written, contents []Content, copyContents bool) (Group, error) {
	if !t.IsSet() {
		return Group{}, unsetTupletError(op, 0)
	}
	if !unit.IsSet() {
		return Group{}, unsetError(op, "unit")
	}

	span, err := t.Span(unit.value)
	if err != nil {
		return Group{}, rangeError(op)
	}

	contentChain := packChain(append(unpackChain(unit.chain), t))

	sum := fraction.FromInt(0)
	for i, c := range contents {
		if !c.IsValid() {
			return Group{}, invalidError(op, "content "+strconv.Itoa(i))
		}
		if c.IsGroup() {
			if c.group.unit.chain != contentChain {
				return Group{}, invalidError(op, "nested group unit chain does not match content "+strconv.Itoa(i))
			}
		} else if c.written.chain != contentChain {
			return Group{}, invalidError(op, "content chain does not match content "+strconv.Itoa(i))
		}
		sum, err = sum.Add(c.sounding())
		if err != nil {
			return Group{}, rangeError(op)
		}
	}
	if sum.Compare(span) > 0 {
		return Group{}, invalidError(op, "contents exceed the group span")
	}

	if copyContents {
		contents = cloneContents(contents)
	} else if len(contents) == 0 {
		contents = nil
	}
	return Group{
		t:        t,
		unit:     unit,
		contents: contents,
		span:     span,
		sounding: sum,
	}, nil
}

// MustNewGroup is like NewGroup but panics with the returned error.
func MustNewGroup(t tuplet.Tuplet, unit Written, contents ...Content) Group {
	return must(NewGroup(t, unit, contents...))
}

func cloneContents(contents []Content) []Content {
	if len(contents) == 0 {
		return nil
	}
	out := make([]Content, len(contents))
	for i, c := range contents {
		out[i] = c.clone()
	}
	return out
}

func (g Group) clone() *Group {
	if !g.IsSet() {
		return &Group{}
	}
	cp := g
	cp.contents = cloneContents(g.contents)
	return &cp
}

// IsSet reports whether g is a set group rather than the zero value.
func (g Group) IsSet() bool { return g.t.IsSet() }

func (g Group) unset(op string) error { return unsetError(op, "receiver") }

// Tuplet returns the tuplet of g. It returns an error wrapping ErrUnset if g
// is unset.
func (g Group) Tuplet() (tuplet.Tuplet, error) {
	if !g.IsSet() {
		return tuplet.Tuplet{}, g.unset("Group.Tuplet")
	}
	return g.t, nil
}

// MustTuplet is like Tuplet but panics with the returned error.
func (g Group) MustTuplet() tuplet.Tuplet { return must(g.Tuplet()) }

// Unit returns the written unit of g. It returns an error wrapping ErrUnset
// if g is unset.
func (g Group) Unit() (Written, error) {
	if !g.IsSet() {
		return Written{}, g.unset("Group.Unit")
	}
	return g.unit, nil
}

// MustUnit is like Unit but panics with the returned error.
func (g Group) MustUnit() Written { return must(g.Unit()) }

// Contents returns a deep copy of the contents of g, in order. It returns an
// error wrapping ErrUnset if g is unset. The caller owns the result.
func (g Group) Contents() ([]Content, error) {
	if !g.IsSet() {
		return nil, g.unset("Group.Contents")
	}
	return cloneContents(g.contents), nil
}

// MustContents is like Contents but panics with the returned error.
func (g Group) MustContents() []Content { return must(g.Contents()) }

// Span returns the total sounding duration of g: the normal count of its
// tuplet times the value of its unit. It returns an error wrapping ErrUnset
// if g is unset.
func (g Group) Span() (fraction.Fraction, error) {
	if !g.IsSet() {
		return fraction.Fraction{}, g.unset("Group.Span")
	}
	return g.span, nil
}

// MustSpan is like Span but panics with the returned error.
func (g Group) MustSpan() fraction.Fraction { return must(g.Span()) }

// Shortfall returns the part of the span of g that its contents do not fill:
// Span minus the sum of the sounding durations of the contents. It is zero
// for an exactly filled group. It returns an error wrapping ErrUnset if g is
// unset.
func (g Group) Shortfall() (fraction.Fraction, error) {
	if !g.IsSet() {
		return fraction.Fraction{}, g.unset("Group.Shortfall")
	}
	return g.span.Sub(g.sounding)
}

// MustShortfall is like Shortfall but panics with the returned error.
func (g Group) MustShortfall() fraction.Fraction { return must(g.Shortfall()) }

// Spans returns the half-open sounding spans of the contents of g relative to
// the group start, in order: the span of content i is [start, end) with
// end − start equal to its sounding duration, and the start of the next
// content is end. A 3:2 group in eighths holding a quarter and an eighth
// gives [0, 1/6) and [1/6, 1/4). It returns an error wrapping ErrUnset if g is
// unset. The caller owns the result.
func (g Group) Spans() ([][2]fraction.Fraction, error) {
	if !g.IsSet() {
		return nil, g.unset("Group.Spans")
	}
	spans := make([][2]fraction.Fraction, 0, len(g.contents))
	start := fraction.FromInt(0)
	for _, c := range g.contents {
		end, err := start.Add(c.sounding())
		if err != nil {
			// Defensive: every set group is built by newGroup, which adds
			// the same contents in the same order from zero and rejects a
			// sum outside the range, so each end here was already computed
			// there without error. The branch keeps the error explicit
			// should a group ever be built otherwise.
			return nil, rangeError("Group.Spans")
		}
		spans = append(spans, [2]fraction.Fraction{start, end})
		start = end
	}
	return spans, nil
}

// MustSpans is like Spans but panics with the returned error.
func (g Group) MustSpans() [][2]fraction.Fraction { return must(g.Spans()) }

// Equal reports whether g and other are the same group: the same tuplet as
// given, the same written unit, and equal contents in order, nested groups
// compared recursively. The unset group is equal only to the unset group.
func (g Group) Equal(other Group) bool {
	if !g.IsSet() || !other.IsSet() {
		return g.IsSet() == other.IsSet()
	}
	if !g.t.Equal(other.t) || !g.unit.Equal(other.unit) || len(g.contents) != len(other.contents) {
		return false
	}
	for i, a := range g.contents {
		b := other.contents[i]
		if a.IsGroup() != b.IsGroup() {
			return false
		}
		if a.IsGroup() {
			if !a.group.Equal(*b.group) {
				return false
			}
		} else if !a.written.Equal(b.written) {
			return false
		}
	}
	return true
}
