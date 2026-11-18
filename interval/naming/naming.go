// Package naming provides explicit, immutable Russian and English interval
// dictionaries. Dictionary names omit direction; interval.Parse and String
// remain the directed canonical representation. Legacy interval name methods
// retain their historical behavior.
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
