package template

import (
	"iter"
	"strings"
)

// Catalog is an immutable ordered set of entries with pairwise distinct exact
// codes. The zero Catalog is a valid empty catalog. Copies may share immutable
// internal data, and concurrent reads and iterations are safe.
type Catalog struct {
	entries []Entry
	byCode  map[string]int
}

// NewCatalog creates a catalog in the order of entries. With no arguments it
// returns an empty catalog.
func NewCatalog(entries ...Entry) (Catalog, error) {
	for _, entry := range entries {
		if !entry.IsValid() {
			return Catalog{}, opError("NewCatalog", ErrUnset)
		}
	}
	seen := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		if _, exists := seen[entry.code]; exists {
			return Catalog{}, opError("NewCatalog", ErrDuplicateCode)
		}
		seen[entry.code] = struct{}{}
	}
	return catalogFromEntries(entries), nil
}

// MustNewCatalog is NewCatalog without the error result. It panics with the
// same error.
func MustNewCatalog(entries ...Entry) Catalog {
	return must(NewCatalog(entries...))
}

func catalogFromEntries(entries []Entry) Catalog {
	owned := append([]Entry{}, entries...)
	index := make(map[string]int, len(owned))
	for i, entry := range owned {
		index[entry.code] = i
	}
	return Catalog{entries: owned, byCode: index}
}

// Len returns the number of entries.
func (c Catalog) Len() int {
	return len(c.entries)
}

// All returns a repeatable, interruptible sequence over the immutable snapshot
// of the catalog at the time of the call. Independent iterations may run
// concurrently.
func (c Catalog) All() iter.Seq[Entry] {
	entries := c.entries
	return func(yield func(Entry) bool) {
		for _, entry := range entries {
			if !yield(entry) {
				return
			}
		}
	}
}

// Entries returns a new slice of entries in catalog order.
func (c Catalog) Entries() []Entry {
	return append([]Entry{}, c.entries...)
}

// Codes returns a new slice of exact entry codes in catalog order.
func (c Catalog) Codes() []string {
	result := make([]string, len(c.entries))
	for i, entry := range c.entries {
		result[i] = entry.code
	}
	return result
}

// Templates returns a new slice of entry templates in catalog order.
func (c Catalog) Templates() []Template {
	result := make([]Template, len(c.entries))
	for i, entry := range c.entries {
		result[i] = entry.tmpl
	}
	return result
}

// ByCode returns the entry with the exact code, or (Entry{}, false).
func (c Catalog) ByCode(code string) (Entry, bool) {
	index, ok := c.byCode[code]
	if !ok {
		return Entry{}, false
	}
	return c.entries[index], true
}

// ByName returns every entry with a name equal to name after normalization, in
// catalog order. Codes are not searched, and each entry occurs at most once.
func (c Catalog) ByName(name string) []Entry {
	key := normalizeName(name)
	result := make([]Entry, 0)
	for _, entry := range c.entries {
		for _, normalized := range entry.normalized {
			if normalized == key {
				result = append(result, entry)
				break
			}
		}
	}
	return result
}

// With returns a new catalog with entries appended in argument order. Existing
// and newly added codes must all be distinct.
func (c Catalog) With(entries ...Entry) (Catalog, error) {
	for _, entry := range entries {
		if !entry.IsValid() {
			return Catalog{}, opError("With", ErrUnset)
		}
	}
	seen := make(map[string]struct{}, len(c.entries)+len(entries))
	for _, entry := range c.entries {
		seen[entry.code] = struct{}{}
	}
	for _, entry := range entries {
		if _, exists := seen[entry.code]; exists {
			return Catalog{}, opError("With", ErrDuplicateCode)
		}
		seen[entry.code] = struct{}{}
	}
	combined := make([]Entry, 0, len(c.entries)+len(entries))
	combined = append(combined, c.entries...)
	combined = append(combined, entries...)
	return catalogFromEntries(combined), nil
}

// MustWith is With without the error result. It panics with the same error.
func (c Catalog) MustWith(entries ...Entry) Catalog {
	return must(c.With(entries...))
}

// Without returns a new catalog without entries having codes. Codes are removed
// in argument order. An unknown code, including a repeated code in one call,
// returns ErrUnknownCode without a partial result.
func (c Catalog) Without(codes ...string) (Catalog, error) {
	remaining := make(map[string]int, len(c.byCode))
	for code, index := range c.byCode {
		remaining[code] = index
	}
	for _, code := range codes {
		if _, exists := remaining[code]; !exists {
			return Catalog{}, opError("Without", ErrUnknownCode)
		}
		delete(remaining, code)
	}
	entries := make([]Entry, 0, len(remaining))
	for _, entry := range c.entries {
		if _, exists := remaining[entry.code]; exists {
			entries = append(entries, entry)
		}
	}
	return catalogFromEntries(entries), nil
}

// MustWithout is Without without the error result. It panics with the same
// error.
func (c Catalog) MustWithout(codes ...string) Catalog {
	return must(c.Without(codes...))
}

// Merge returns a new catalog containing the receiver's entries followed by
// other's entries. Any duplicate exact code returns ErrDuplicateCode.
func (c Catalog) Merge(other Catalog) (Catalog, error) {
	result, err := c.With(other.entries...)
	if err != nil {
		return Catalog{}, opError("Merge", ErrDuplicateCode)
	}
	return result, nil
}

// MustMerge is Merge without the error result. It panics with the same error.
func (c Catalog) MustMerge(other Catalog) Catalog {
	return must(c.Merge(other))
}

// Filter returns entries accepted by predicate in their original order. The
// predicate is called once per entry. A nil predicate returns ErrUnset; a panic
// in the predicate is not recovered.
func (c Catalog) Filter(predicate func(Entry) bool) (Catalog, error) {
	if predicate == nil {
		return Catalog{}, opError("Filter", ErrUnset)
	}
	entries := make([]Entry, 0)
	for _, entry := range c.entries {
		if predicate(entry) {
			entries = append(entries, entry)
		}
	}
	return catalogFromEntries(entries), nil
}

// MustFilter is Filter without the error result. It panics with the same error.
func (c Catalog) MustFilter(predicate func(Entry) bool) Catalog {
	return must(c.Filter(predicate))
}

// Equal reports whether both catalogs have equal entries in the same order.
// The zero catalog equals every other empty catalog.
func (c Catalog) Equal(other Catalog) bool {
	if len(c.entries) != len(other.entries) {
		return false
	}
	for i, entry := range c.entries {
		if !entry.Equal(other.entries[i]) {
			return false
		}
	}
	return true
}

// String returns a debug representation consisting of entry codes.
func (c Catalog) String() string {
	return "Catalog[" + strings.Join(c.Codes(), ", ") + "]"
}
