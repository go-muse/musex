package duration

import (
	"iter"
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
//
// FreeContents lets a group hold any number, at least one, of consecutive
// items whose written sum under the chain of the group is the actual count
// times the unit, instead of exactly the actual count of items. Such a group
// may end after any input duration; the durations after it are read as
// written values outside groups or as further groups, nested up to MaxDepth.
// It is off by default. With it on, the set of readings only grows: every
// reading found without it is also found with it under the same other
// options, unless the larger search exceeds its work bound, and a
// quarter and an eighth sounding 1/6 and 1/12 followed by a sounding quarter
// gain a 3:2 group in eighths followed by a plain quarter. Imported rhythms
// that do not split into equal parts gain readings as well.
type ReadingsOptions struct {
	Allowed      []tuplet.Tuplet
	Longest      Base
	MaxDots      int
	MaxParts     int
	MaxDepth     int
	MaxReadings  int
	FreeContents bool
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
		} else if !a.written.Equal(b.written) {
			return false
		}
	}
	return true
}

// maxSearchCovers bounds the number of covers produced by one Readings call.
const maxSearchCovers = 100000

// defaultMaxReadings is the result bound when MaxReadings is 0.
const defaultMaxReadings = 256

// Readings returns the exact rhythmic readings of the ordered sounding
// durations ds under opts.
//
// A reading covers ds in order by items: a written value outside any group,
// or a group of an allowed ratio. By default a group of ratio t holds exactly
// t.Actual() items over consecutive input durations; with FreeContents it
// holds any number, at least one, of items over consecutive input durations
// whose written values under the chain of the group sum to t.Actual() times
// the unit, and it may end after any input duration, with the rest of the
// input read by further items. Each written value of a group carries the
// chain of the unit extended by t, as [NewGroup] requires; the unit of a
// nested group carries that chain too. A group is filled exactly: the sum of
// the sounding durations of its contents is the group span, and its written
// unit is that sum divided by t.Normal(), required to be one written sign
// under the enclosing chain. Durations that need no tuplet stay outside
// groups. With MaxDepth > 0 an item of a group may be a nested group over a
// consecutive part of the inputs, up to that depth.
//
// The search is exhaustive within opts: every exact reading is found, and no
// two returned readings are equal by [Reading.Equal]. The readings are
// ordered deterministically: fewer groups first, then lexicographically by
// item kind (written values before groups), ratio, unit, and contents. At
// most MaxReadings readings are returned, the first ones in that order. No
// reading is an empty result, not an error.
//
// The work of a call is bounded independently of the result count: covers
// longer than MaxParts are pruned during the search, and at most a fixed
// number of covers is produced, the contents of groups included; exceeding
// that bound fails with ErrTooManyParts rather than exhausting time or
// memory. The number of readings grows quickly with the length of the input,
// especially with FreeContents and nesting, so a long input is better split
// into parts, for example by measures, and read part by part; otherwise the
// search reaches its bound.
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
		ds:           ds,
		allowed:      slices.Clone(opts.Allowed),
		longest:      longest,
		maxDots:      opts.MaxDots,
		maxParts:     maxParts,
		maxDepth:     opts.MaxDepth,
		freeContents: opts.FreeContents,
		cache:        make(map[string][]coverResult),
		groups:       make(map[string][]groupCover),
		spans:        make(map[string][]groupSpan),
		prefix:       prefixSums(ds),
	}

	covers := s.coverAll(0, len(ds), 0, nil)
	if s.exceeded {
		return nil, simpleError(ErrTooManyParts, op)
	}

	readings := make([]Reading, 0, len(covers))
	for _, c := range covers {
		if c.items != nil && c.parts <= maxParts {
			readings = append(readings, Reading{contents: c.items.slice()})
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
	ds           []Duration
	allowed      []tuplet.Tuplet
	longest      Base
	maxDots      int
	maxParts     int
	maxDepth     int
	freeContents bool
	work         int
	exceeded     bool
	cache        map[string][]coverResult
	groups       map[string][]groupCover
	spans        map[string][]groupSpan
	prefix       []*big.Rat
}

// prefixSums returns the sounding sums of the prefixes of ds: element k is
// the sum of ds[:k], so the sum of ds[i:j] is element j minus element i.
func prefixSums(ds []Duration) []*big.Rat {
	out := make([]*big.Rat, len(ds)+1)
	out[0] = new(big.Rat)
	for k, d := range ds {
		out[k+1] = new(big.Rat).Add(out[k], ratOf(d.v))
	}
	return out
}

// spanSum returns the sounding sum of ds[i:j] in a new *big.Rat.
func (s *readingSearch) spanSum(i, j int) *big.Rat {
	return new(big.Rat).Sub(s.prefix[j], s.prefix[i])
}

// produced records one produced cover toward the work bound.
func (s *readingSearch) produced() bool {
	s.work++
	if s.work > maxSearchCovers {
		s.exceeded = true
		return false
	}
	return true
}

// coverResult is one cover found by the search: its items, the inputs it
// consumes, and the written values it uses.
type coverResult struct {
	items *itemList
	used  int
	parts int
}

// itemList is an immutable list of the items of a cover. Covers that end alike
// share their tails, so producing a cover costs one node rather than a copy
// of the whole cover; the search works within its bound in time and memory.
type itemList struct {
	head Content
	tail *itemList
	n    int
}

// prepend returns the list of c followed by l.
func (l *itemList) prepend(c Content) *itemList {
	n := 1
	if l != nil {
		n += l.n
	}
	return &itemList{head: c, tail: l, n: n}
}

// slice returns the items of l in order, in a new slice.
func (l *itemList) slice() []Content {
	if l == nil {
		return nil
	}
	out := make([]Content, 0, l.n)
	for ; l != nil; l = l.tail {
		out = append(out, l.head)
	}
	return out
}

// coverAll returns the covers of all of ds[i:end] by any number of items under
// chain at the given depth.
func (s *readingSearch) coverAll(i, end, depth int, chain []tuplet.Tuplet) []coverResult {
	return s.cover(i, end, depth, -1, chain)
}

// coverExactly returns the covers of a prefix of ds[i:end] by exactly n items
// under chain at the given depth.
func (s *readingSearch) coverExactly(i, end, depth, n int, chain []tuplet.Tuplet) []coverResult {
	return s.cover(i, end, depth, n, chain)
}

// cover returns the covers of a prefix of ds[i:end] by exactly n items, or of
// all of ds[i:end] by any number of items when n < 0. Each result records the
// inputs consumed and the written values used. Covers are cached by position,
// end, depth, count, and chain.
func (s *readingSearch) cover(i, end, depth, n int, chain []tuplet.Tuplet) []coverResult {
	if s.exceeded {
		return nil
	}
	if n < 0 {
		n = -1
	}
	if n == 0 {
		return []coverResult{{}}
	}
	if i >= end {
		if n < 0 {
			return []coverResult{{}}
		}
		return nil
	}

	key := strconv.Itoa(i) + "/" + strconv.Itoa(end) + "/" + strconv.Itoa(depth) + "/" +
		strconv.Itoa(n) + "/" + packChain(chain)
	if cached, ok := s.cache[key]; ok {
		return cached
	}

	var out []coverResult

	// One written value for ds[i] under chain, then the remaining items.
	if w, ok := s.signFor(s.ds[i], chain); ok {
		item := WrittenContent(w)
		for _, rest := range s.cover(i+1, end, depth, n-1, chain) {
			if 1+rest.parts > s.maxParts || !s.produced() {
				continue
			}
			out = append(out, coverResult{
				items: rest.items.prepend(item),
				used:  1 + rest.used,
				parts: 1 + rest.parts,
			})
		}
	}

	// A group of an allowed ratio: one item of this cover. A group at depth d
	// is nested d levels; MaxDepth bounds that level, so top-level groups are
	// allowed with MaxDepth 0 and nesting needs MaxDepth > 0.
	if depth <= s.maxDepth {
		for _, t := range s.allowed {
			for g := range s.groupsAt(t, i, end, depth, chain) {
				for _, rest := range s.cover(i+g.used, end, depth, n-1, chain) {
					if g.parts+rest.parts > s.maxParts || !s.produced() {
						continue
					}
					out = append(out, coverResult{
						items: rest.items.prepend(g.item),
						used:  g.used + rest.used,
						parts: g.parts + rest.parts,
					})
				}
			}
		}
	}

	s.cache[key] = out
	return out
}

// groupCover is one exactly filled group starting at a position: the group
// as an item, the inputs it consumes, and the written values it uses. The
// item is shared by every cover that holds the group; covers are only read.
type groupCover struct {
	item  Content
	used  int
	parts int
}

// groupSpan is a candidate span ds[i:end] of a group with free contents
// together with its unit: the sounding sum of the span divided by the normal
// count of the ratio is one written sign under the enclosing chain.
type groupSpan struct {
	end  int
	unit Written
	span fraction.Fraction
}

// groupsAt yields the exactly filled groups of ratio t that start at ds[i],
// end at most at end, and sit under chain at the given depth.
//
// By default the contents are exactly t.Actual() items at depth+1, and the
// unit is derived from their sounding sum. With FreeContents the contents are
// any number of items covering all of ds[i:j], for every j in (i, end] whose
// span has a unit; a free group therefore may end before the input does, and
// the caller covers the items after it. In both cases the unit is one written
// sign under chain and the sounding sum of the contents is t.Normal() units,
// so the written sum of the contents under the chain of the group is
// t.Actual() units: the group is filled exactly.
func (s *readingSearch) groupsAt(t tuplet.Tuplet, i, end, depth int, chain []tuplet.Tuplet) iter.Seq[groupCover] {
	return func(yield func(groupCover) bool) {
		innerChain := append(slices.Clone(chain), t)
		if !s.freeContents {
			for _, g := range s.fixedGroups(t, i, end, depth, chain, innerChain) {
				if !yield(g) {
					return
				}
			}
			return
		}
		for _, sp := range s.spansAt(t, i, chain, innerChain) {
			if sp.end > end || s.exceeded {
				return
			}
			for _, g := range s.freeGroups(t, sp, i, depth, innerChain) {
				if !yield(g) {
					return
				}
			}
		}
	}
}

// fixedGroups returns the groups of ratio t whose contents are exactly
// t.Actual() items covering a prefix of ds[i:end]. The groups are cached, so
// each content cover becomes a group once.
func (s *readingSearch) fixedGroups(t tuplet.Tuplet, i, end, depth int, chain, innerChain []tuplet.Tuplet) []groupCover {
	key := strconv.Itoa(i) + "/" + strconv.Itoa(end) + "/" + strconv.Itoa(depth) + "/" + packChain(innerChain)
	if cached, ok := s.groups[key]; ok {
		return cached
	}
	actual, normal := t.MustPair()
	coef, err := checkedCoefficient("Readings", chain)
	var out []groupCover
	if err == nil {
		for _, c := range s.coverExactly(i, end, depth+1, actual, innerChain) {
			if c.used == 0 || c.parts > s.maxParts {
				continue
			}
			unit, ok := s.unitOf(s.spanSum(i, i+c.used), normal, coef, chain)
			if !ok {
				continue
			}
			span, err := t.Span(unit.value)
			if err != nil {
				continue
			}
			out = append(out, groupOf(t, unit, span, c))
		}
	}
	s.groups[key] = out
	return out
}

// spansAt returns, in increasing order of end, the spans ds[i:end] that a
// group of ratio t under chain can fill exactly with free contents: those
// whose unit is one written sign within the base and dot bounds. Every
// written sign is shorter than twice the longest base, and the sum of a span
// grows with its end, so only the spans below that limit are examined. The
// spans are cached by position and chain.
func (s *readingSearch) spansAt(t tuplet.Tuplet, i int, chain, innerChain []tuplet.Tuplet) []groupSpan {
	key := strconv.Itoa(i) + "/" + packChain(innerChain)
	if cached, ok := s.spans[key]; ok {
		return cached
	}
	var out []groupSpan
	coef, err := checkedCoefficient("Readings", chain)
	if err == nil {
		// scale turns a sounding sum into the plain value of the unit:
		// sum / normal / coefficient of the enclosing chain.
		scale := new(big.Rat).Mul(coef, new(big.Rat).SetInt64(int64(t.MustNormal())))
		limit := new(big.Rat).Mul(ratOf(s.longest.fraction()), scale)
		limit.Add(limit, limit)
		limit.Add(limit, s.prefix[i])
		scale.Inv(scale)
		last, _ := slices.BinarySearchFunc(s.prefix[i+1:], limit, (*big.Rat).Cmp)
		plain := new(big.Rat)
		for j := i + 1; j <= i+last; j++ {
			plain.Sub(s.prefix[j], s.prefix[i])
			plain.Mul(plain, scale)
			unit, ok := s.signRat(plain, chain)
			if !ok {
				continue
			}
			span, err := t.Span(unit.value)
			if err != nil {
				continue
			}
			out = append(out, groupSpan{end: j, unit: unit, span: span})
		}
	}
	s.spans[key] = out
	return out
}

// freeGroups returns the groups of ratio t with the unit of sp whose contents
// are any cover of all of ds[i:sp.end] at depth+1. The groups are cached, so
// each content cover becomes a group once.
func (s *readingSearch) freeGroups(t tuplet.Tuplet, sp groupSpan, i, depth int, innerChain []tuplet.Tuplet) []groupCover {
	key := strconv.Itoa(i) + "/" + strconv.Itoa(sp.end) + "/" + strconv.Itoa(depth) + "/" + packChain(innerChain)
	if cached, ok := s.groups[key]; ok {
		return cached
	}
	var out []groupCover
	for _, c := range s.coverAll(i, sp.end, depth+1, innerChain) {
		if c.used == 0 || c.parts > s.maxParts {
			continue
		}
		out = append(out, groupOf(t, sp.unit, sp.span, c))
	}
	s.groups[key] = out
	return out
}

// unitOf returns the written unit of a group under chain, with coefficient
// coef, whose contents sound sum and whose ratio has the given normal count:
// the sounding unit sum/normal written as one sign under chain, or false.
func (s *readingSearch) unitOf(sum *big.Rat, normal int, coef *big.Rat, chain []tuplet.Tuplet) (Written, bool) {
	plain := new(big.Rat).Quo(sum, new(big.Rat).SetInt64(int64(normal)))
	plain.Quo(plain, coef)
	return s.signRat(plain, chain)
}

// groupOf returns the group of ratio t in unit, with the given span, over
// the content cover c, as an item of covers. The search builds the group
// exactly as NewGroup would accept it: the unit is written under the
// enclosing chain, every written content carries that chain extended by t
// and every nested unit carries it too, and the sounding sum of the contents
// is the span. The contents are not copied: covers and their items are
// immutable and only read, and Reading.Contents returns a deep copy.
func groupOf(t tuplet.Tuplet, unit Written, span fraction.Fraction, c coverResult) groupCover {
	g := &Group{t: t, unit: unit, contents: c.items.slice(), span: span, sounding: span}
	return groupCover{item: Content{group: g}, used: c.used, parts: c.parts}
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
	return s.signRat(v, chain)
}

// signRat returns the single written sign whose plain value, without the
// chain coefficient, is v, written under chain within the base and dot
// bounds, or false. It only reads v.
func (s *readingSearch) signRat(v *big.Rat, chain []tuplet.Tuplet) (Written, bool) {
	if v.Sign() <= 0 || !isPow2(v.Denom()) {
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
