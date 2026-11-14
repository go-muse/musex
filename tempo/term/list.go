package term

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/go-muse/musex/common/fraction"
)

// Entry is one term of a user row: its name and optional meanings.
type Entry struct {
	// Name is the name of the term: non-empty valid UTF-8 text. Any
	// characters are allowed, including "=", "≈", commas, quotation marks,
	// spaces at the ends, and control characters; the name is stored as
	// given.
	Name string
	// English is the English meaning; "" means that there is none.
	English string
	// Russian is the Russian meaning; "" means that there is none.
	Russian string
}

// Row is one place of a user list: one or more terms that share a group and
// a guide.
type Row struct {
	// Entries lists the terms of the place in order; there must be at least
	// one.
	Entries []Entry
	// Group is the group of the terms; NoGroup means no group.
	Group Group
	// Guide is the guide of the terms; nil means no guide. A non-nil guide
	// must have at least one bound.
	Guide *Guide
}

// List is an ordered list of tempo terms: one of the eight built-in lists or a
// list made by the user with [NewList]. Each row of a list is a place numbered
// from 1, and a place holds one or more terms that share a group and a guide.
// The names of a list are pairwise different up to case.
//
// The zero value List{} is the unset list; only IsSet, Equal, and String
// accept it without an error. List is comparable with ==, which matches Equal
// and compares the content: the rows in order with their terms, meanings,
// groups, and guides, and the synonyms of parsing. A list can be used as a map
// key. Copies share immutable data.
type List struct {
	pack string // "" for the unset list
}

// NewList returns the list with the given rows in order; the place of a row is
// its number from 1. The rows, their entries, and their guides are only read
// and not retained. A user list has no synonyms of parsing, and its rows need
// not be ordered by their guides. It returns an error wrapping
// ErrInvalidArgument if there are no rows, a row has no entries, a name is
// empty or not valid UTF-8, two names are equal up to case (as
// strings.EqualFold compares them), a group is not NoGroup, Slow, Moderate, or
// Fast, or a guide is not nil and has no bounds.
func NewList(rows ...Row) (List, error) {
	return newList("NewList", rows, nil)
}

// MustNewList is like NewList but panics with the returned error.
func MustNewList(rows ...Row) List {
	return must(NewList(rows...))
}

// newList checks rows, packs them with synonyms, and returns the list.
func newList(op string, rows []Row, synonyms []synonym) (List, error) {
	if len(rows) == 0 {
		return List{}, invalidError(op, "no rows")
	}
	var words []packWord
	seen := make(map[string]struct{})
	for i, row := range rows {
		place := i + 1
		if len(row.Entries) == 0 {
			return List{}, invalidError(op, "row "+strconv.Itoa(place)+" has no entries")
		}
		if !row.Group.IsValid() {
			return List{}, invalidError(op, "row "+strconv.Itoa(place)+" has an invalid group")
		}
		var guide *Guide
		if row.Guide != nil {
			g := *row.Guide
			if !g.IsValid() {
				return List{}, invalidError(op, "row "+strconv.Itoa(place)+" has a guide without bounds")
			}
			guide = &g
		}
		for _, e := range row.Entries {
			if e.Name == "" {
				return List{}, invalidError(op, "row "+strconv.Itoa(place)+" has an empty name")
			}
			if !utf8.ValidString(e.Name) {
				return List{}, invalidError(op, "row "+strconv.Itoa(place)+" has a name that is not valid UTF-8")
			}
			key := foldKey(e.Name)
			if _, dup := seen[key]; dup {
				return List{}, invalidError(op, "row "+strconv.Itoa(place)+" repeats a name up to case")
			}
			seen[key] = struct{}{}
			words = append(words, packWord{
				place:   place,
				group:   row.Group,
				guide:   guide,
				name:    e.Name,
				english: e.English,
				russian: e.Russian,
			})
		}
	}
	return List{pack: pack(words, len(rows), synonyms)}, nil
}

// foldKey returns a key of the valid UTF-8 text s such that two texts have the
// same key exactly when strings.EqualFold reports them equal: every rune is
// replaced by the smallest rune of its simple case-folding orbit.
func foldKey(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		m := r
		for f := unicode.SimpleFold(r); f != r; f = unicode.SimpleFold(f) {
			m = min(m, f)
		}
		b.WriteRune(m)
	}
	return b.String()
}

// IsSet reports whether l is set, that is, not the unset List{}.
func (l List) IsSet() bool {
	return l.pack != ""
}

// words returns the number of terms of the set list l.
func (l List) words() int {
	return u64At(l.pack, headerWords)
}

// unset returns an error wrapping ErrUnset for the receiver of op.
func (List) unset(op string) error {
	return unsetError(op, "receiver")
}

// Len returns the number of places of l. It returns an error wrapping ErrUnset
// if l is unset.
func (l List) Len() (int, error) {
	if !l.IsSet() {
		return 0, l.unset("List.Len")
	}
	return u64At(l.pack, headerPlaces), nil
}

// MustLen is like Len but panics with the returned error.
func (l List) MustLen() int {
	return must(l.Len())
}

// Place returns the terms of the given place, numbered from 1, in order, as a
// new slice. In order of checking, it returns an error wrapping ErrUnset if l
// is unset and ErrInvalidArgument if place is outside 1..Len.
func (l List) Place(place int) ([]Term, error) {
	const op = "List.Place"
	if !l.IsSet() {
		return nil, l.unset(op)
	}
	if place < 1 || place > u64At(l.pack, headerPlaces) {
		return nil, invalidError(op, "place "+strconv.Itoa(place)+" outside the list")
	}
	var out []Term
	for i := range l.words() {
		if recordAt(l.pack, i).place == place {
			out = append(out, Term{list: l, index: i})
		}
	}
	return out, nil
}

// MustPlace is like Place but panics with the returned error.
func (l List) MustPlace(place int) []Term {
	return must(l.Place(place))
}

// Terms returns all terms of l by place, and within a place in the order of
// its row, as a new slice. It returns an error wrapping ErrUnset if l is
// unset.
func (l List) Terms() ([]Term, error) {
	if !l.IsSet() {
		return nil, l.unset("List.Terms")
	}
	out := make([]Term, l.words())
	for i := range out {
		out[i] = Term{list: l, index: i}
	}
	return out, nil
}

// MustTerms is like Terms but panics with the returned error.
func (l List) MustTerms() []Term {
	return must(l.Terms())
}

// Parse returns the term of l whose name equals name up to case, as
// strings.EqualFold compares them, without trimming spaces. In the list of
// the English Wikipedia it also accepts "Vivo" as Vivace. In order of
// checking, it returns an error wrapping ErrUnset if l is unset, and
// ErrInvalidText if no term has the name, including for the empty string and
// for text that is not valid UTF-8. The error quotes an excerpt of name.
func (l List) Parse(name string) (Term, error) {
	if !l.IsSet() {
		return Term{}, l.unset("List.Parse")
	}
	if i, ok := l.find(name); ok {
		return Term{list: l, index: i}, nil
	}
	return Term{}, textError(name)
}

// MustParse is like Parse but panics with the returned error.
func (l List) MustParse(name string) Term {
	return must(l.Parse(name))
}

// find returns the index of the term of the set list l named name up to case,
// including the synonyms of parsing.
func (l List) find(name string) (int, bool) {
	if name == "" || !utf8.ValidString(name) {
		return 0, false
	}
	for i := range l.words() {
		if strings.EqualFold(recordAt(l.pack, i).name, name) {
			return i, true
		}
	}
	for _, s := range synonymsOf(l.pack) {
		if strings.EqualFold(s.alias, name) {
			return s.index, true
		}
	}
	return 0, false
}

// OfGroup returns the terms of group g in the order of l, as a new slice;
// NoGroup gives the terms without a group. If there are none, it returns nil
// without an error. In order of checking, it returns an error wrapping
// ErrUnset if l is unset and ErrInvalidArgument if g is not NoGroup, Slow,
// Moderate, or Fast.
func (l List) OfGroup(g Group) ([]Term, error) {
	const op = "List.OfGroup"
	if !l.IsSet() {
		return nil, l.unset(op)
	}
	if !g.IsValid() {
		return nil, invalidError(op, "invalid group")
	}
	var out []Term
	for i := range l.words() {
		if recordAt(l.pack, i).group == g {
			out = append(out, Term{list: l, index: i})
		}
	}
	return out, nil
}

// MustOfGroup is like OfGroup but panics with the returned error.
func (l List) MustOfGroup(g Group) []Term {
	return must(l.OfGroup(g))
}

// At returns all terms of l whose guide contains beats, bounds included, in
// the order of l, as a new slice. The number is compared as it is, without a
// duration; terms without a guide are not returned. If no guide contains the
// number, it returns nil without an error. In order of checking, it returns an
// error wrapping ErrUnset if l is unset, ErrInvalidArgument if beats is unset,
// and ErrNoTempo if beats is zero or negative.
func (l List) At(beats fraction.Fraction) ([]Term, error) {
	const op = "List.At"
	if !l.IsSet() {
		return nil, l.unset(op)
	}
	if !beats.IsValid() {
		return nil, invalidError(op, "unset number of beats")
	}
	if beats.Sign() <= 0 {
		return nil, noTempoError(op)
	}
	var out []Term
	for i := range l.words() {
		if g, ok := recordAt(l.pack, i).guide(); ok && g.contains(beats) {
			out = append(out, Term{list: l, index: i})
		}
	}
	return out, nil
}

// MustAt is like At but panics with the returned error.
func (l List) MustAt(beats fraction.Fraction) []Term {
	return must(l.At(beats))
}

// Equal reports whether l and other have the same content: equal rows in
// order with the same terms, meanings, groups, and guides, and the same
// synonyms of parsing. A user list is therefore never equal to the list of the
// English Wikipedia, which parses "Vivo". Two unset lists are equal. Equal
// matches ==.
func (l List) Equal(other List) bool {
	return l == other
}

// String returns a debugging form that is not parsed: "List[" followed by the
// places separated by "; ", each with the names of its terms separated by ", ",
// and "]", such as "List[Larghissimo; Adagissimo, Grave; Largo; ...]". For the
// unset list it returns "List(unset)".
func (l List) String() string {
	if !l.IsSet() {
		return "List(unset)"
	}
	var b strings.Builder
	b.WriteString("List[")
	prev := 0
	for i := range l.words() {
		rec := recordAt(l.pack, i)
		switch {
		case i == 0:
		case rec.place != prev:
			b.WriteString("; ")
		default:
			b.WriteString(", ")
		}
		prev = rec.place
		b.WriteString(rec.name)
	}
	b.WriteByte(']')
	return b.String()
}
