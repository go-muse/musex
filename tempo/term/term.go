package term

// Term is one tempo term of a list, such as Allegro of the list of the English
// Wikipedia. A term is obtained from its list with List.Parse, List.Place,
// List.Terms, List.OfGroup, or List.At; it knows its list, name, place, group,
// guide, and meanings. Terms with the same name in different lists are
// different terms.
//
// The zero value Term{} is the unset term; only IsSet, Equal, ComparePlace,
// and String accept it without an error, and its String is the empty string,
// which no name equals. Term is comparable with ==, which matches Equal, and
// can be used as a map key.
type Term struct {
	list  List
	index int // index of the word in the list, from 0
}

// IsSet reports whether t is set, that is, not the unset Term{}.
func (t Term) IsSet() bool {
	return t.list.IsSet()
}

// record returns the record of the set term t.
func (t Term) record() record {
	return recordAt(t.list.pack, t.index)
}

// unset returns an error wrapping ErrUnset for the receiver of op.
func (Term) unset(op string) error {
	return unsetError(op, "receiver")
}

// Name returns the name of t as the list stores it. It returns an error
// wrapping ErrUnset if t is unset.
func (t Term) Name() (string, error) {
	if !t.IsSet() {
		return "", t.unset("Term.Name")
	}
	return t.record().name, nil
}

// MustName is like Name but panics with the returned error.
func (t Term) MustName() string {
	return must(t.Name())
}

// List returns the list of t. It returns an error wrapping ErrUnset if t is
// unset.
func (t Term) List() (List, error) {
	if !t.IsSet() {
		return List{}, t.unset("Term.List")
	}
	return t.list, nil
}

// MustList is like List but panics with the returned error.
func (t Term) MustList() List {
	return must(t.List())
}

// Place returns the place of t in its list, numbered from 1. It returns an
// error wrapping ErrUnset if t is unset.
func (t Term) Place() (int, error) {
	if !t.IsSet() {
		return 0, t.unset("Term.Place")
	}
	return t.record().place, nil
}

// MustPlace is like Place but panics with the returned error.
func (t Term) MustPlace() int {
	return must(t.Place())
}

// Group returns the group of t, which is NoGroup for a term without a group.
// It returns an error wrapping ErrUnset if t is unset.
func (t Term) Group() (Group, error) {
	if !t.IsSet() {
		return NoGroup, t.unset("Term.Group")
	}
	return t.record().group, nil
}

// MustGroup is like Group but panics with the returned error.
func (t Term) MustGroup() Group {
	return must(t.Group())
}

// Guide returns the guide of t and true, or Guide{} and false without an error
// if t has no guide. It returns an error wrapping ErrUnset if t is unset.
func (t Term) Guide() (Guide, bool, error) {
	if !t.IsSet() {
		return Guide{}, false, t.unset("Term.Guide")
	}
	g, ok := t.record().guide()
	return g, ok, nil
}

// MustGuide is like Guide but panics with the returned error.
func (t Term) MustGuide() (Guide, bool) {
	return must2(t.Guide())
}

// English returns the English meaning of t and true, or "" and false without
// an error if t has none. It returns an error wrapping ErrUnset if t is unset.
func (t Term) English() (string, bool, error) {
	if !t.IsSet() {
		return "", false, t.unset("Term.English")
	}
	m := t.record().english
	return m, m != "", nil
}

// MustEnglish is like English but panics with the returned error.
func (t Term) MustEnglish() (string, bool) {
	return must2(t.English())
}

// Russian returns the Russian meaning of t and true, or "" and false without
// an error if t has none. It returns an error wrapping ErrUnset if t is unset.
func (t Term) Russian() (string, bool, error) {
	if !t.IsSet() {
		return "", false, t.unset("Term.Russian")
	}
	m := t.record().russian
	return m, m != "", nil
}

// MustRussian is like Russian but panics with the returned error.
func (t Term) MustRussian() (string, bool) {
	return must2(t.Russian())
}

// Equal reports whether t and other are the same term: equal lists, as
// List.Equal compares them, and the same name. Two unset terms are equal.
// Equal matches ==.
func (t Term) Equal(other Term) bool {
	return t == other
}

// ComparePlace compares the places of t and other in their list and returns
// −1, 0, or +1; terms of the same place compare as 0. It compares places, not
// tempos. The unset term precedes every set term and two unset terms compare
// as 0, without an error. It returns an error wrapping ErrInvalidArgument if
// t and other are set terms of different lists.
func (t Term) ComparePlace(other Term) (int, error) {
	switch {
	case !t.IsSet() && !other.IsSet():
		return 0, nil
	case !t.IsSet():
		return -1, nil
	case !other.IsSet():
		return +1, nil
	case t.list != other.list:
		return 0, invalidError("Term.ComparePlace", "terms of different lists")
	}
	a, b := t.record().place, other.record().place
	switch {
	case a < b:
		return -1, nil
	case a > b:
		return +1, nil
	default:
		return 0, nil
	}
}

// MustComparePlace is like ComparePlace but panics with the returned error.
func (t Term) MustComparePlace(other Term) int {
	return must(t.ComparePlace(other))
}

// String returns the name of t, and the empty string for the unset term.
// List.Parse(t.String()) on the list of t returns t.
func (t Term) String() string {
	if !t.IsSet() {
		return ""
	}
	return t.record().name
}

// MarshalText implements encoding.TextMarshaler with the name of t. It returns
// nil and an error wrapping ErrUnset if t is unset. Term does not implement
// encoding.TextUnmarshaler: a name is parsed only in a named list, with
// List.Parse.
func (t Term) MarshalText() ([]byte, error) {
	if !t.IsSet() {
		return nil, t.unset("Term.MarshalText")
	}
	return []byte(t.record().name), nil
}

// AppendText implements encoding.TextAppender: it appends the name of t to dst
// without modifying dst[:len(dst)]. It returns dst unchanged and an error
// wrapping ErrUnset if t is unset.
func (t Term) AppendText(dst []byte) ([]byte, error) {
	if !t.IsSet() {
		return dst, t.unset("Term.AppendText")
	}
	return append(dst, t.record().name...), nil
}
