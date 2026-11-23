// Package naming provides explicit, immutable Russian and English interval
// dictionaries. Dictionary names omit direction; interval.Parse and String
// remain the directed canonical representation. Legacy interval name methods
// retain their historical behavior.
//
// Catalog.Entries enumerates the whole dictionary in a documented
// deterministic order, and Entry.IsGeneric, Entry.Generic and Entry.Interval
// read the subject of an entry. Every enumerated entry has a full name; the
// short form is present exactly when the language gives one.
package naming

import (
	"strings"
	"unicode"

	"github.com/go-muse/musex/interval"
)

// Catalog selects a language. Zero is unset. Values can be copied and read
// concurrently; Go equality is not a public comparison contract.
type Catalog struct{ language uint8 }

// Entry binds a catalog to one immutable generic or specific interval.
// Zero and entries with an unset subject return interval.ErrUnset.
type Entry struct {
	catalog   Catalog
	generic   interval.Generic
	specific  interval.Interval
	isGeneric bool
}

// Russian selects the Russian dictionary.
func Russian() Catalog { return Catalog{1} }

// English selects the English dictionary.
func English() Catalog { return Catalog{2} }

// Generic selects an interval size without quality.
func (c Catalog) Generic(g interval.Generic) Entry {
	return Entry{catalog: c, generic: g, isGeneric: true}
}

// Specific selects an interval without including its direction in its name.
func (c Catalog) Specific(i interval.Interval) Entry { return Entry{catalog: c, specific: i} }
func (e Entry) valid() bool {
	return e.catalog.language != 0 && ((e.isGeneric && e.generic.IsValid()) || (!e.isGeneric && e.specific.IsValid()))
}

// IsGeneric reports whether the entry selects a generic size rather than a
// specific interval. The zero entry reports false.
func (e Entry) IsGeneric() bool { return e.isGeneric }

// Generic returns the selected size, or the unset size for a specific entry
// and for the zero entry.
func (e Entry) Generic() interval.Generic { return e.generic }

// Interval returns the selected specific interval, or the unset interval for
// a generic entry and for the zero entry. The direction of the subject is
// preserved; the dictionary names omit it.
func (e Entry) Interval() interval.Interval { return e.specific }

// genericName derives size words from the same dictionary as the legacy
// names, so additions and corrections have a single source of truth.
func (c Catalog) genericName(g interval.Generic) (string, bool, error) {
	if c.language == 0 || !g.IsValid() {
		return "", false, interval.ErrUnset
	}
	i := interval.MustFromDeviation(g.Size(), interval.PerfectUnison().Deviation())
	var s string
	var ok bool
	var err error
	if c.language == 1 {
		s, ok, err = i.RussianName()
	} else {
		s, ok, err = i.Name()
	}
	if !ok || err != nil {
		return "", ok, err
	}
	_, name, _ := strings.Cut(s, " ")
	return name, true, nil
}

// Name returns the short dictionary form and whether it exists. A valid
// subject outside the dictionary returns "", false, nil.
func (e Entry) Name() (string, bool, error) {
	if !e.valid() {
		return "", false, interval.ErrUnset
	}
	if e.isGeneric {
		return e.catalog.genericName(e.generic)
	}
	if e.catalog.language == 1 {
		return e.specific.RussianAbbreviation()
	}
	_, ok, err := e.specific.Name()
	if !ok || err != nil {
		return "", ok, err
	}
	return e.specific.MustUp().String(), true, nil
}

// NameFull returns the full dictionary form, with the same absence rules.
func (e Entry) NameFull() (string, bool, error) {
	if !e.valid() {
		return "", false, interval.ErrUnset
	}
	if e.isGeneric {
		return e.catalog.genericName(e.generic)
	}
	if e.catalog.language == 1 {
		return e.specific.RussianName()
	}
	return e.specific.Name()
}
func mustName(s string, ok bool, err error) (string, bool) {
	if err != nil {
		panic(err)
	}
	return s, ok
}

// MustName panics with the error returned by Name.
func (e Entry) MustName() (string, bool) { return mustName(e.Name()) }

// MustNameFull panics with the error returned by NameFull.
func (e Entry) MustNameFull() (string, bool) { return mustName(e.NameFull()) }

// ParseGeneric reads full size words 1..15, ignoring case. Extra spaces,
// unknown words and other-language words return interval.ErrParse.
func (c Catalog) ParseGeneric(text string) (interval.Generic, error) {
	if c.language == 0 {
		return interval.Generic{}, interval.ErrUnset
	}
	for n := 1; n <= 15; n++ {
		g := interval.MustNewGeneric(n)
		name, _, _ := c.genericName(g)
		if strings.EqualFold(text, name) {
			return g, nil
		}
	}
	return interval.Generic{}, interval.ErrParse
}

// MustParseGeneric panics with the error returned by ParseGeneric.
func (c Catalog) MustParseGeneric(text string) interval.Generic {
	g, err := c.ParseGeneric(text)
	if err != nil {
		panic(err)
	}
	return g
}

// ParseInterval reads this language's dictionary forms and legacy synonyms.
// English short forms distinguish M and m exactly. Direction signs are not
// dictionary syntax. The returned interval is ascending.
func (c Catalog) ParseInterval(text string) (interval.Interval, error) {
	if c.language == 0 {
		return interval.Interval{}, interval.ErrUnset
	}
	if strings.HasPrefix(text, "+") || strings.HasPrefix(text, "-") {
		return interval.Interval{}, interval.ErrParse
	}
	if c.language == 2 {
		if i, err := interval.Parse(text); err == nil {
			name, ok, _ := c.Specific(i).Name()
			if ok && name == text {
				return i, nil
			}
		}
	}
	russian := false
	for _, r := range text {
		if unicode.Is(unicode.Cyrillic, r) {
			russian = true
			break
		}
	}
	if russian != (c.language == 1) {
		return interval.Interval{}, interval.ErrParse
	}
	return interval.ParseName(text)
}

// MustParseInterval panics with the error returned by ParseInterval.
func (c Catalog) MustParseInterval(text string) interval.Interval {
	i, err := c.ParseInterval(text)
	if err != nil {
		panic(err)
	}
	return i
}

// Entries returns the whole vocabulary of the catalog as a new slice, in a
// deterministic order: the generic sizes 1 to 15 in ascending order, then
// the named specific intervals in ascending size; within one size the
// qualities follow the order perfect, major, minor, neutral, then augmented
// and diminished in multiplicity order 1, 2, 3, augmented before diminished
// at each multiplicity. Only the entries with a full name in this language
// are included, so the neutral third is present in Russian without a short
// form, and the neutral second is present in English with both forms.
//
// Every returned entry has a full name (Entry.NameFull); its short form
// (Entry.Name) is present exactly when the language gives one. The specific
// subjects are ascending; direction belongs to the canonical interval text,
// not to the dictionary. The enumeration agrees with the naming and parsing
// APIs: ParseInterval of a full name returns the subject of the entry, and
// ParseGeneric of a generic full name returns its size.
//
// Errors: interval.ErrUnset for the zero catalog.
func (c Catalog) Entries() ([]Entry, error) {
	if c.language == 0 {
		return nil, interval.ErrUnset
	}

	out := make([]Entry, 0, 128)

	for n := 1; n <= 15; n++ {
		out = append(out, c.Generic(interval.MustNewGeneric(n)))
	}

	for n := 1; n <= 15; n++ {
		for _, q := range []interval.Quality{interval.Perfect, interval.Major, interval.Minor, interval.Neutral} {
			i, err := interval.New(q, n)
			if err != nil {
				continue
			}

			e := c.Specific(i)
			if _, ok := e.MustNameFull(); ok {
				out = append(out, e)
			}
		}

		for mult := 1; mult <= 3; mult++ {
			for _, q := range []interval.Quality{interval.Augmented, interval.Diminished} {
				i, err := interval.NewMultiple(mult, q, n)
				if err != nil {
					continue
				}

				e := c.Specific(i)
				if _, ok := e.MustNameFull(); ok {
					out = append(out, e)
				}
			}
		}
	}

	return out, nil
}

// MustEntries panics with the error returned by Entries.
func (c Catalog) MustEntries() []Entry {
	entries, err := c.Entries()
	if err != nil {
		panic(err)
	}
	return entries
}
