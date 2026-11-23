package duration

import (
	"math/big"
	"slices"
	"strconv"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/tuplet"
)

// ReadingsOptions bounds the search of [Readings].
//
// Allowed lists the tuplet ratios a reading may use; an empty list means no
// tuplets. Longest is the longest allowed base; the zero value means Breve.
// MaxDots is the largest number of dots of a written sign, at least 0.
// MaxParts is the largest number of written values in one reading; 0 means
// [MaxParts]. MaxDepth is the largest nesting depth of groups; 0 (the default)
// means no nesting. MaxReadings is the largest number of readings returned;
// 0 means 256.
type ReadingsOptions struct {
	Allowed     []tuplet.Tuplet
	Longest     Base
	MaxDots     int
	MaxParts    int
	MaxDepth    int
	MaxReadings int
}

// Reading is one exact rhythmic reading of a sequence of sounding durations:
// an ordered cover by written values outside groups and by exactly filled
// groups. Every input duration has exactly one written value, in order,
// inside its chain of groups; the written value times the chain coefficient
// equals the sounding duration exactly. Build readings with [Readings]; the
// zero value is not a valid reading.
type Reading struct {
	contents []Content
}

// Contents returns a deep copy of the items of the reading, in order: written
// values outside groups and groups. It returns an error wrapping
// ErrInvalidArgument if r is not a valid reading. The caller owns the result.
func (r Reading) Contents() ([]Content, error) {
	if len(r.contents) == 0 {
		return nil, invalidError("Reading.Contents", "unset reading")
	}
	return cloneContents(r.contents), nil
}

// MustContents is like Contents but panics with the returned error.
func (r Reading) MustContents() []Content { return must(r.Contents()) }

// Equal reports whether r and other have equal items in order.
func (r Reading) Equal(other Reading) bool {
	if len(r.contents) != len(other.contents) {
		return false
	}
	for i, a := range r.contents {
		b := other.contents[i]
		if a.IsGroup() != b.IsGroup() {
			return false
		}
		if a.IsGroup() {
			if !a.group.Equal(*b.group) {
				return false
			}
		} else if a.written != b.written {
			return false
		}
	}
	return true
}

// maxSearchNodes bounds the work of one Readings call.
const maxSearchNodes = 100000

// defaultMaxReadings is the result bound when MaxReadings is 0.
const defaultMaxReadings = 256

// Readings returns the exact rhythmic readings of the ordered sounding
// durations ds under opts.
//
// A reading covers ds in order by items: a written value outside any group,
// or a group of an allowed ratio. A group of ratio t covers exactly t.Actual()
// consecutive input durations, one written value each under the chain extended
// by t, and is filled exactly: the sum of their sounding durations is the
// group span, and the written unit of the group is that sum divided by
// t.Normal(), required to be one written sign. Durations that need no tuplet
// stay outside groups. With MaxDepth > 0 an item of a group may be a nested
// group over a consecutive part of the inputs, up to that depth.
//
// The readings are ordered deterministically: fewer groups first, then
// lexicographically by item kind (written values before groups), ratio, unit,
// and contents. At most MaxReadings readings are returned. No reading is an
// empty result, not an error.
//
// In order of checking, it returns an error wrapping ErrInvalidArgument for
// an empty list, an unset duration, a negative MaxDots, MaxDepth or
// MaxReadings, a MaxParts above MaxParts, or an unset tuplet of Allowed
// (together with tuplet.ErrUnset); ErrInvalidArgument if Longest is set but is
// not a base; and ErrTooManyParts if the search exceeds its work bound. The
// inputs and opts are only read.
func Readings(ds []Duration, opts ReadingsOptions) ([]Reading, error) {
	const op = "Readings"
	if err := checkList(op, ds); err != nil {
		return nil, err
	}
	if opts.MaxDots < 0 || opts.MaxDepth < 0 || opts.MaxReadings < 0 {
		return nil, invalidError(op, "negative bound")
	}
	if opts.MaxParts < 0 || opts.MaxParts > MaxParts {
		return nil, invalidError(op, "MaxParts out of range")
	}
	longest := opts.Longest
	if longest == 0 {
		longest = Breve
	}
	if !longest.IsValid() {
		return nil, invalidBaseError(op, longest)
	}
	for i, t := range opts.Allowed {
		if !t.IsSet() {
			return nil, unsetTupletError(op, i)
		}
	}
	maxParts := opts.MaxParts
	if maxParts == 0 {
		maxParts = MaxParts
	}
	maxReadings := opts.MaxReadings
	if maxReadings == 0 {
		maxReadings = defaultMaxReadings
	}

	s := &readingSearch{
		ds:       ds,
		allowed:  slices.Clone(opts.Allowed),
		longest:  longest,
		maxDots:  opts.MaxDots,
		maxParts: maxParts,
		maxDepth: opts.MaxDepth,
		cache:    make(map[string][]coverResult),
	}

	covers := s.coverAny(0, 0, nil)
	if s.nodes > maxSearchNodes {
		return nil, simpleError(ErrTooManyParts, op)
	}

	readings := make([]Reading, 0, len(covers))
	for _, c := range covers {
		if len(c.items) > 0 && c.parts <= maxParts {
			readings = append(readings, Reading{contents: c.items})
		}
	}
	slices.SortFunc(readings, compareReadings)
	if len(readings) > maxReadings {
		readings = readings[:maxReadings]
	}
	return readings, nil
}

// MustReadings is like Readings but panics with the returned error.
func MustReadings(ds []Duration, opts ReadingsOptions) []Reading {
	return must(Readings(ds, opts))
}

type readingSearch struct {
	ds       []Duration
	allowed  []tuplet.Tuplet
	longest  Base
	maxDots  int
	maxParts int
	maxDepth int
	nodes    int
	cache    map[string][]coverResult
}

type coverResult struct {
	items []Content
	used  int
	parts int
}

// coverAny returns the covers of ds[i:] by any number of items under chain at
// the given depth.
func (s *readingSearch) coverAny(i, depth int, chain []tuplet.Tuplet) []coverResult {
	return s.cover(i, depth, -1, chain)
}

// coverExactly returns the covers of a prefix of ds[i:] by exactly n items
// under chain at the given depth.
func (s *readingSearch) coverExactly(i, depth, n int, chain []tuplet.Tuplet) []coverResult {
	return s.cover(i, depth, n, chain)
}

// cover returns the covers of ds[i:] by exactly n items, or by any number of
// items when n < 0. Each result records the inputs consumed and the written
// values used. Covers are cached by position, depth, count, and chain.
func (s *readingSearch) cover(i, depth, n int, chain []tuplet.Tuplet) []coverResult {
	s.nodes++
	if s.nodes > maxSearchNodes {
		return nil
	}
	if n < 0 {
		n = -1
	}
	if n == 0 {
		return []coverResult{{}}
	}
	if i >= len(s.ds) {
		if n < 0 {
			return []coverResult{{}}
		}
		return nil
	}

	key := strconv.Itoa(i) + "/" + strconv.Itoa(depth) + "/" + strconv.Itoa(n) + "/" + packChain(chain)
	if cached, ok := s.cache[key]; ok {
		return cached
	}

	var out []coverResult

	// One written value for ds[i] under chain, then the remaining items.
	if w, ok := s.signFor(s.ds[i], chain); ok {
		for _, rest := range s.cover(i+1, depth, n-1, chain) {
			out = append(out, coverResult{
				items: append([]Content{WrittenContent(w)}, rest.items...),
				used:  1 + rest.used,
				parts: 1 + rest.parts,
			})
		}
	}

	// A group of an allowed ratio: one item of this cover, whose contents are
	// exactly the actual count parts at depth+1. A group at depth d is nested
	// d levels; MaxDepth bounds that level, so top-level groups are allowed
	// with MaxDepth 0 and nesting needs MaxDepth > 0.
	if depth <= s.maxDepth {
		for _, t := range s.allowed {
			actual, normal := t.MustPair()
			innerChain := append(slices.Clone(chain), t)
			inner := s.coverExactly(i, depth+1, actual, innerChain)
			for _, c := range inner {
				if c.used == 0 {
					continue
				}
				g, ok := s.groupOf(t, normal, i, c)
				if !ok {
					continue
				}
				for _, rest := range s.cover(i+c.used, depth, n-1, chain) {
					out = append(out, coverResult{
						items: append([]Content{GroupContent(g)}, rest.items...),
						used:  c.used + rest.used,
						parts: c.parts + rest.parts,
					})
				}
			}
		}
	}

	s.cache[key] = out
	return out
}

// groupOf builds the exactly filled group of ratio t over the inputs consumed
// by the content cover c, with the unit derived from their sounding sum.
func (s *readingSearch) groupOf(t tuplet.Tuplet, normal, i int, c coverResult) (Group, bool) {
	sum := fraction.FromInt(0)
	for k := 0; k < c.used; k++ {
		var err error
		sum, err = sum.Add(s.ds[i+k].v)
		if err != nil {
			return Group{}, false
		}
	}
	unitValue, err := sum.Div(fraction.FromInt(normal))
	if err != nil {
		return Group{}, false
	}
	unit, ok := s.sign(unitValue, nil)
	if !ok {
		return Group{}, false
	}
	g, err := NewGroup(t, unit, c.items...)
	if err != nil {
		return Group{}, false
	}
	return g, true
}

// signFor returns the single written sign of d under chain, within the base
// and dot bounds.
func (s *readingSearch) signFor(d Duration, chain []tuplet.Tuplet) (Written, bool) {
	coef, err := checkedCoefficient("Readings", chain)
	if err != nil {
		return Written{}, false
	}
	v := ratOf(d.v)
	v.Quo(v, coef)
	f, ok := fromRat(v)
	if !ok {
		return Written{}, false
	}
	return s.sign(f, chain)
}

// sign returns the single written sign of value under chain, within the base
// and dot bounds, or false.
func (s *readingSearch) sign(value fraction.Fraction, chain []tuplet.Tuplet) (Written, bool) {
	v := ratOf(value)
	if !isPow2(v.Denom()) {
		return Written{}, false
	}
	shift := v.Num().TrailingZeroBits()
	a := new(big.Int).Rsh(v.Num(), shift)
	a.Add(a, big.NewInt(1))
	if !isPow2(a) {
		return Written{}, false
	}
	k := a.BitLen() - 2
	if k > s.maxDots {
		return Written{}, false
	}
	j := int(shift) - (v.Denom().BitLen() - 1)
	b, ok := baseOfExp(j + k)
	if !ok || b.exp() > s.longest.exp() {
		return Written{}, false
	}
	w, err := makeWritten("Readings", b, k, chain)
	if err != nil {
		return Written{}, false
	}
	return w, true
}

// groupCount counts the groups of a cover, recursively.
func groupCount(items []Content) int {
	n := 0
	for _, c := range items {
		if c.IsGroup() {
			n += 1 + groupCount(c.group.contents)
		}
	}
	return n
}

// compareReadings orders readings: fewer groups first, then lexicographically
// by item kind, ratio, unit, and contents.
func compareReadings(a, b Reading) int {
	if c := cmpInt(groupCount(a.contents), groupCount(b.contents)); c != 0 {
		return c
	}
	return compareContents(a.contents, b.contents)
}

func cmpInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func compareContents(a, b []Content) int {
	for i := 0; i < len(a) && i < len(b); i++ {
		if c := compareContent(a[i], b[i]); c != 0 {
			return c
		}
	}
	return cmpInt(len(a), len(b))
}

func compareContent(a, b Content) int {
	if a.IsGroup() != b.IsGroup() {
		if a.IsGroup() {
			return 1
		}
		return -1
	}
	if !a.IsGroup() {
		return a.written.Compare(b.written)
	}
	ga, gb := a.group, b.group
	if c := cmpInt(ga.t.MustActual(), gb.t.MustActual()); c != 0 {
		return c
	}
	if c := cmpInt(ga.t.MustNormal(), gb.t.MustNormal()); c != 0 {
		return c
	}
	if c := ga.unit.Compare(gb.unit); c != 0 {
		return c
	}
	return compareContents(ga.contents, gb.contents)
}
