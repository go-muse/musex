package instrument

import (
	"iter"
	"strings"
)

// Catalog is an immutable ordered collection of presets with different names.
// Names are compared exactly, with case and spaces. A catalog is not a
// registry: the package keeps no mutable list of all presets, and a preset
// works the same whether it is in a catalog or not.
//
// A Catalog is used through a pointer and never changes; With and Without
// return a new catalog. It holds its own copies of the presets it receives
// and returns new copies, so assigning through a pointer given to or received
// from a catalog does not change the catalog. The sets of settings inside the
// copies are shared, as sets are immutable. The nil pointer and a pointer to
// the zero value are the empty catalog.
type Catalog struct {
	presets []Preset
}

// NewCatalog returns a catalog of presets in the given order; without
// arguments, the empty catalog.
//
// Errors, in this order: ErrUnset for the first unset preset; ErrDuplicateName
// for the first repeated name, at its second occurrence.
func NewCatalog(presets ...*Preset) (*Catalog, error) {
	const op = "new catalog"

	for _, p := range presets {
		if !p.IsValid() {
			return nil, unsetError(op, "preset")
		}
	}

	c := &Catalog{presets: make([]Preset, 0, len(presets))}
	for _, p := range presets {
		if c.index(p.name) >= 0 {
			return nil, duplicateNameError(op, p.name)
		}

		c.presets = append(c.presets, *p)
	}

	return c, nil
}

// MustNewCatalog is like NewCatalog but panics with its error on failure.
func MustNewCatalog(presets ...*Preset) *Catalog {
	return must(NewCatalog(presets...))
}

// Len returns the number of presets; 0 for the empty catalog.
func (c *Catalog) Len() int {
	return len(c.list())
}

// All returns the presets in order, each as a new copy. The sequence can be
// stopped at any element and iterated again; for the empty catalog it is
// empty.
func (c *Catalog) All() iter.Seq[*Preset] {
	list := c.list()

	return func(yield func(*Preset) bool) {
		for i := range list {
			p := list[i]
			if !yield(&p) {
				return
			}
		}
	}
}

// Names returns the names of the presets in order.
func (c *Catalog) Names() iter.Seq[string] {
	list := c.list()

	return func(yield func(string) bool) {
		for i := range list {
			if !yield(list[i].name) {
				return
			}
		}
	}
}

// Lookup returns a new copy of the preset with exactly this name and true,
// or nil and false when there is none. The comparison is exact: another case
// of letters or an extra space gives no preset.
func (c *Catalog) Lookup(name string) (*Preset, bool) {
	i := c.index(name)
	if i < 0 {
		return nil, false
	}

	p := c.presets[i]

	return &p, true
}

// Find is like Lookup but fails with ErrNotFound when there is no preset with
// this name.
func (c *Catalog) Find(name string) (*Preset, error) {
	p, ok := c.Lookup(name)
	if !ok {
		return nil, notFoundError("find", name)
	}

	return p, nil
}

// MustFind is like Find but panics with its error on failure.
func (c *Catalog) MustFind(name string) *Preset {
	return must(c.Find(name))
}

// With returns a new catalog with p added at the end. For the empty catalog,
// including nil, it is a catalog of p alone.
//
// Errors, in this order: ErrUnset when p is unset; ErrDuplicateName when the
// catalog already has a preset with the name of p.
func (c *Catalog) With(p *Preset) (*Catalog, error) {
	const op = "catalog with"

	if !p.IsValid() {
		return nil, unsetError(op, "preset")
	}

	if c.index(p.name) >= 0 {
		return nil, duplicateNameError(op, p.name)
	}

	list := c.list()
	presets := make([]Preset, len(list), len(list)+1)
	copy(presets, list)

	return &Catalog{presets: append(presets, *p)}, nil
}

// MustWith is like With but panics with its error on failure.
func (c *Catalog) MustWith(p *Preset) *Catalog {
	return must(c.With(p))
}

// Without returns a new catalog without the preset with this name; the other
// presets keep their order. It fails with ErrNotFound when there is no such
// preset.
func (c *Catalog) Without(name string) (*Catalog, error) {
	i := c.index(name)
	if i < 0 {
		return nil, notFoundError("catalog without", name)
	}

	presets := make([]Preset, 0, len(c.presets)-1)
	presets = append(presets, c.presets[:i]...)
	presets = append(presets, c.presets[i+1:]...)

	return &Catalog{presets: presets}, nil
}

// MustWithout is like Without but panics with its error on failure.
func (c *Catalog) MustWithout(name string) *Catalog {
	return must(c.Without(name))
}

// Equal reports whether c and other have the same number of presets and
// equal presets (Preset.Equal) in the same order. The empty catalogs,
// including nil, are equal.
func (c *Catalog) Equal(other *Catalog) bool {
	a, b := c.list(), other.list()
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if !a[i].Equal(&b[i]) {
			return false
		}
	}

	return true
}

// String returns a debugging form with the names of all presets in order, or
// "Catalog(empty)". It is not a notation, has no parser and does not affect
// equality.
func (c *Catalog) String() string {
	list := c.list()
	if len(list) == 0 {
		return "Catalog(empty)"
	}

	var b strings.Builder
	b.WriteString("Catalog(")
	for i := range list {
		if i > 0 {
			b.WriteString(", ")
		}

		b.WriteString(list[i].name)
	}

	b.WriteString(")")

	return b.String()
}

// list returns the presets of c; nil for the empty catalog.
func (c *Catalog) list() []Preset {
	if c == nil {
		return nil
	}

	return c.presets
}

// index returns the index of the preset with this name, or −1.
func (c *Catalog) index(name string) int {
	list := c.list()
	for i := range list {
		if list[i].name == name {
			return i
		}
	}

	return -1
}
