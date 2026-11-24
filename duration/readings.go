package duration

import (
	"math/big"
	"slices"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/tuplet"
)

// ReadingsOptions bounds the search of [Readings].
//
// Allowed lists the tuplet ratios a reading may use; an empty list means no
// tuplets. It is a set by [tuplet.Tuplet.Equal]: a repeated ratio gives the
// same readings as one occurrence, and the order of the list does not change
// the result, which follows the order of [Readings]. Longest is the longest
// allowed base; the zero value means Breve. MaxDots is the largest number of
// dots of a written sign, at least 0. MaxParts is the largest number of
// written values in one reading; 0 means [MaxParts]. A reading has one written
// value per input duration, so an input longer than MaxParts has no reading.
// MaxDepth is the largest nesting depth of groups; 0 (the default) means no
// nesting. MaxReadings is the largest number of readings returned; 0 means
// 256. The memory of the result is proportional to MaxReadings times the
// length of the input, since each returned reading holds its own items.
//
// FreeContents lets a group hold any number, at least one, of consecutive
// items whose written sum under the chain of the group is the actual count
// times the unit, instead of exactly the actual count of items. Such a group
// may end after any input duration; the durations after it are read as
// written values outside groups or as further groups, nested up to MaxDepth.
// It is off by default. With it on, the set of all readings only grows: every
// reading found without it is also one with it under the same other options,
// and a quarter and an eighth sounding 1/6 and 1/12 followed by a sounding
// quarter gain a 3:2 group in eighths followed by a plain quarter. Imported
// rhythms that do not split into equal parts gain readings as well. Two
// limits apply to the returned list: the larger search may exceed its work
// bound and fail with [ErrSearchLimit], and the first MaxReadings readings
// may differ, since new readings can precede old ones in the order of
// [Readings]. For six sounding 1/12 with 3:2 and 6:4 and MaxReadings 2, the
// list without FreeContents is one 6:4 group in eighths, then two 3:2 groups
// in eighths; with it, one 3:2 group in quarters, then one 6:4 group in
// eighths.
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

// maxSearchSteps bounds the work of one Readings call. Every step of the
// search counts toward it: a computed cover entry, a ratio tried, a span
// examined, a group considered, a content cover turned into a group (one step
// per input it covers), a produced cover, a chain of ratios or a written sign
// built, and each item visited while choosing the first readings in order.
// Work that grows with the depth of a chain of ratios is charged by
// [chainSteps], so the cost of a step is bounded independently of the input
// and the options, and a call ends within bounded time and memory.
const maxSearchSteps = 200000

// chainTupletsPerStep is the number of tuplets of a chain that the search
// may process for one step: packing a chain, comparing it with the chain of
// a group, or building a sign under it.
const chainTupletsPerStep = 16

// chainSteps returns the steps charged for work over a chain of depth d.
func chainSteps(d int) int { return 1 + d/chainTupletsPerStep }

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
// under the enclosing chain. A duration that can be written as one sign
// without a tuplet may stand outside groups, and it may also be read inside
// a group when a group over it is filled exactly. With MaxDepth > 0 an item
// of a group may be a nested group over a consecutive part of the inputs, up
// to that depth.
//
// Every reading has exactly len(ds) written values, one per input duration,
// so an input longer than MaxParts has no reading: the result is empty
// without a search. An input of exactly MaxParts durations is read.
//
// The search is exhaustive within opts: every exact reading is found, and no
// two returned readings are equal by [Reading.Equal]. The readings are
// ordered deterministically: fewer groups first, then lexicographically by
// item kind (written values before groups), ratio, unit, and contents. At
// most MaxReadings readings are returned, the first ones in that order. No
// reading is an empty result, not an error.
//
// The work of a call is bounded independently of the input and the result
// count: every step of the search counts toward a fixed bound, including the
// branches that produce no cover, such as spans that no unit fills, groups
// over durations that no sign writes, and nested ratios at every depth, with
// work over long chains of ratios charged in proportion to their depth. The
// first MaxReadings readings are chosen by comparing the found covers, which
// also counts toward the bound, and only those are built. Exceeding the bound
// fails with ErrSearchLimit rather than exhausting time or memory; within it,
// a call returns its readings or an empty result. Building the returned
// readings is outside the bound: each reading holds its own items, so the
// memory of the result is proportional to the number of returned readings
// times the length of the input, and MaxReadings limits it. The number of
// readings grows quickly with the length of the input, the number of ratios,
// and the nesting depth, especially with FreeContents, so a long input is
// better split into parts, for example by measures, and read part by part;
// otherwise the search reaches its bound.
//
// In order of checking, it returns an error wrapping ErrInvalidArgument for
// an empty list, an unset duration, a negative MaxDots, MaxDepth or
// MaxReadings, a MaxParts above MaxParts, or an unset tuplet of Allowed
// (together with tuplet.ErrUnset); ErrInvalidArgument if Longest is set but is
// not a base; and ErrSearchLimit if the search exceeds its work bound. The
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
	if len(ds) > maxParts {
		return []Reading{}, nil
	}

	s := &readingSearch{
		ds:           ds,
		allowed:      distinctTuplets(opts.Allowed),
		longest:      longest,
		maxDots:      opts.MaxDots,
		maxDepth:     opts.MaxDepth,
		freeContents: opts.FreeContents,
		prefix:       prefixSums(ds),
		covers:       make(map[coverKey][]coverResult),
		fixed:        make(map[groupKey][]groupCover),
		free:         make(map[groupKey][]groupCover),
		spans:        make(map[groupKey][]groupSpan),
	}

	covers := s.cover(0, len(ds), -1, s.root())
	if s.exceeded {
		return nil, simpleError(ErrSearchLimit, op)
	}

	first := s.first(covers, maxReadings)
	if s.exceeded {
		return nil, simpleError(ErrSearchLimit, op)
	}
	readings := make([]Reading, len(first))
	for k, c := range first {
		readings[k] = Reading{contents: c.items.slice()}
	}
	return readings, nil
}

// first returns the first n of the covers in the order of Readings.
// It keeps at most n candidates and compares covers as item lists, so only
// the returned covers are built into readings; every item visited by a
// comparison counts toward the work bound.
func (s *readingSearch) first(covers []coverResult, n int) []coverResult {
	best := make([]coverResult, 0, min(n, len(covers)))
	for _, c := range covers {
		if len(best) == n {
			if s.compareCovers(c, best[n-1]) >= 0 {
				continue
			}
			best = best[:n-1]
		}
		k, _ := slices.BinarySearchFunc(best, c, s.compareCovers)
		best = slices.Insert(best, k, c)
		if s.exceeded {
			return nil
		}
	}
	return best
}

// compareCovers orders covers in the order of Readings: fewer groups first,
// nested ones included, then lexicographically by items. Covers that share a
// tail of items stop comparing at the shared node.
func (s *readingSearch) compareCovers(a, b coverResult) int {
	if !s.step(1) {
		return 0
	}
	if c := cmpInt(a.groups, b.groups); c != 0 {
		return c
	}
	la, lb := a.items, b.items
	for la != lb {
		if la == nil {
			return -1
		}
		if lb == nil {
			return 1
		}
		if c := s.compareContent(la.head, lb.head); c != 0 {
			return c
		}
		la, lb = la.tail, lb.tail
	}
	return 0
}

// MustReadings is like Readings but panics with the returned error.
func MustReadings(ds []Duration, opts ReadingsOptions) []Reading {
	return must(Readings(ds, opts))
}

// distinctTuplets returns the tuplets of ts without repeats by
// [tuplet.Tuplet.Equal], in order of first occurrence: Allowed is a set.
func distinctTuplets(ts []tuplet.Tuplet) []tuplet.Tuplet {
	out := make([]tuplet.Tuplet, 0, len(ts))
	for _, t := range ts {
		if !slices.ContainsFunc(out, t.Equal) {
			out = append(out, t)
		}
	}
	return out
}

type readingSearch struct {
	ds           []Duration
	allowed      []tuplet.Tuplet
	longest      Base
	maxDots      int
	maxDepth     int
	freeContents bool
	prefix       []*big.Rat
	work         int
	exceeded     bool
	chain        *chainInfo
	covers       map[coverKey][]coverResult
	fixed        map[groupKey][]groupCover
	free         map[groupKey][]groupCover
	spans        map[groupKey][]groupSpan
}

// chainInfo is a chain of tuplets met by the search, built once: its packed
// form, the products of the actual and normal counts, its coefficient, its
// depth, the written signs under it, and the chains extended by each allowed
// ratio. A nil coef means that a product is outside the range of package
// common/fraction, as for checkedCoefficient, so nothing can be written under
// the chain or under its extensions. Every written value under the chain
// shares its packed form.
type chainInfo struct {
	packed   string
	actual   *big.Int
	normal   *big.Int
	coef     *big.Rat
	depth    int
	signs    map[[2]int]Written
	children []*chainInfo
}

// root returns the empty chain of the search.
func (s *readingSearch) root() *chainInfo {
	if s.chain == nil {
		s.chain = &chainInfo{packed: packChain(nil), actual: big.NewInt(1), normal: big.NewInt(1), coef: big.NewRat(1, 1)}
	}
	return s.chain
}

// child returns the chain c extended by the allowed ratio at index k, or nil
// if the work bound is exceeded. Its products and coefficient follow from
// those of c; packing it costs work proportional to its depth, which is
// charged by chainSteps.
func (s *readingSearch) child(c *chainInfo, k int) *chainInfo {
	if c.children == nil {
		c.children = make([]*chainInfo, len(s.allowed))
	}
	if c.children[k] != nil {
		return c.children[k]
	}
	if !s.step(chainSteps(c.depth + 1)) {
		return nil
	}
	t := s.allowed[k]
	actual, normal := t.MustPair()
	ch := &chainInfo{packed: c.packed + packChain([]tuplet.Tuplet{t}), depth: c.depth + 1}
	if c.coef != nil {
		limit := fraction.MaxComponentBig()
		ch.actual = new(big.Int).Mul(c.actual, big.NewInt(int64(actual)))
		ch.normal = new(big.Int).Mul(c.normal, big.NewInt(int64(normal)))
		if ch.actual.Cmp(limit) <= 0 && ch.normal.Cmp(limit) <= 0 {
			ch.coef = new(big.Rat).SetFrac(ch.normal, ch.actual)
		}
	}
	c.children[k] = ch
	return ch
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

// step records n steps of work toward the bound and reports whether the
// search may go on. Once the bound is exceeded, every later step fails and
// every search function returns at once.
func (s *readingSearch) step(n int) bool {
	if s.exceeded {
		return false
	}
	s.work += n
	if s.work > maxSearchSteps {
		s.exceeded = true
		return false
	}
	return true
}

// coverResult is one cover found by the search: its items, the inputs it
// consumes, and the number of its groups, nested ones included. Every input
// has exactly one written value, so the number of written values of a cover is
// the number of inputs it consumes.
type coverResult struct {
	items  *itemList
	used   int
	groups int
}

// itemList is an immutable list of the items of a cover. Covers that end alike
// share their tails, so producing a cover costs one node rather than a copy
// of the whole cover.
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

// coverKey identifies a cover entry; the depth is the length of the chain.
type coverKey struct {
	i, end, n int
	chain     *chainInfo
}

// cover returns the covers of a prefix of ds[i:end] by exactly n items, or of
// all of ds[i:end] by any number of items when n < 0, under chain. Each
// result records its items and the inputs consumed. Covers are cached by
// position, end, count, and chain.
func (s *readingSearch) cover(i, end, n int, chain *chainInfo) []coverResult {
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
	key := coverKey{i: i, end: end, n: n, chain: chain}
	if cached, ok := s.covers[key]; ok {
		return cached
	}
	if !s.step(1) {
		return nil
	}

	var out []coverResult

	// One written value for ds[i] under chain, then the remaining items.
	if w, ok := s.signFor(i, chain); ok {
		item := WrittenContent(w)
		for _, rest := range s.cover(i+1, end, n-1, chain) {
			if !s.step(1) {
				return nil
			}
			out = append(out, coverResult{items: rest.items.prepend(item), used: 1 + rest.used, groups: rest.groups})
		}
	}

	// A group of an allowed ratio: one item of this cover. A group under a
	// chain of depth d is nested d levels; MaxDepth bounds that level, so
	// top-level groups are allowed with MaxDepth 0 and nesting needs
	// MaxDepth > 0.
	if chain.depth <= s.maxDepth && chain.coef != nil {
		for k := range s.allowed {
			if !s.step(1) {
				return nil
			}
			for _, g := range s.groupsAt(k, i, end, chain) {
				if !s.step(1) {
					return nil
				}
				for _, rest := range s.cover(i+g.used, end, n-1, chain) {
					if !s.step(1) {
						return nil
					}
					out = append(out, coverResult{
						items:  rest.items.prepend(g.item),
						used:   g.used + rest.used,
						groups: g.groups + rest.groups,
					})
				}
			}
		}
	}

	s.covers[key] = out
	return out
}

// groupCover is one exactly filled group starting at a position: the group
// as an item, the inputs it consumes, and the number of groups it holds,
// itself included. The item is shared by every cover that holds the group;
// covers are only read.
type groupCover struct {
	item   Content
	used   int
	groups int
}

// groupKey identifies cached groups or spans: the start, the end (the bound
// of fixed contents, or the exact end of free contents), and the chain of the
// contents.
type groupKey struct {
	i, end int
	chain  *chainInfo
}

// groupSpan is a span ds[i:end] that a group with free contents can fill
// exactly, with its unit: the sounding sum of the span divided by the normal
// count of the ratio is one written sign under the enclosing chain.
type groupSpan struct {
	end  int
	unit Written
}

// groupsAt returns the exactly filled groups of the allowed ratio at index k
// that start at ds[i], end at most at end, and sit under chain.
//
// By default the contents are exactly t.Actual() items, and the unit is
// derived from their sounding sum. With FreeContents the contents are any
// number of items covering all of ds[i:j], for every j in (i, end] whose span
// has a unit; a free group therefore may end before the input does, and the
// caller covers the items after it. In both cases the unit is one written
// sign under chain and the sounding sum of the contents is t.Normal() units,
// so the written sum of the contents under the chain of the group is
// t.Actual() units: the group is filled exactly.
func (s *readingSearch) groupsAt(k, i, end int, chain *chainInfo) []groupCover {
	inner := s.child(chain, k)
	if inner == nil {
		return nil
	}
	if !s.freeContents {
		return s.fixedGroups(k, i, end, chain, inner)
	}
	var out []groupCover
	for _, sp := range s.spansAt(k, i, chain, inner) {
		if sp.end > end {
			break
		}
		if !s.step(1) {
			return nil
		}
		out = append(out, s.freeGroups(k, sp, i, inner)...)
	}
	return out
}

// fixedGroups returns the groups of the allowed ratio at index k whose
// contents are exactly t.Actual() items under inner covering a prefix of
// ds[i:end]. The groups are cached, so each content cover becomes a group
// once.
func (s *readingSearch) fixedGroups(k, i, end int, chain, inner *chainInfo) []groupCover {
	key := groupKey{i: i, end: end, chain: inner}
	if cached, ok := s.fixed[key]; ok {
		return cached
	}
	t := s.allowed[k]
	actual, normal := t.MustPair()
	var out []groupCover
	// Each item covers at least one input, so fewer inputs than the actual
	// count leave no fixed contents.
	if end-i < actual {
		s.fixed[key] = out
		return out
	}
	for _, c := range s.cover(i, end, actual, inner) {
		if !s.step((1 + c.used) * chainSteps(inner.depth)) {
			return nil
		}
		sum := new(big.Rat).Sub(s.prefix[i+c.used], s.prefix[i])
		plain := sum.Quo(sum, new(big.Rat).SetInt64(int64(normal)))
		plain.Quo(plain, chain.coef)
		unit, ok := s.signRat(plain, chain)
		if !ok {
			continue
		}
		if g, ok := groupOf(t, unit, c); ok {
			out = append(out, g)
		}
	}
	s.fixed[key] = out
	return out
}

// spansAt returns, in increasing order of end, the spans ds[i:end] that a
// group of the allowed ratio at index k under chain can fill exactly with
// free contents: those whose unit is one written sign within the base and
// dot bounds. Every written sign is shorter than twice the longest base, and
// the sum of a span grows with its end, so only the spans below that limit
// are examined, one step each. The spans are cached by position and chain.
func (s *readingSearch) spansAt(k, i int, chain, inner *chainInfo) []groupSpan {
	key := groupKey{i: i, chain: inner}
	if cached, ok := s.spans[key]; ok {
		return cached
	}
	// scale turns a sounding sum into the plain value of the unit: sum /
	// normal / coefficient of the enclosing chain.
	scale := new(big.Rat).Mul(chain.coef, new(big.Rat).SetInt64(int64(s.allowed[k].MustNormal())))
	limit := new(big.Rat).Mul(ratOf(s.longest.fraction()), scale)
	limit.Add(limit, limit)
	limit.Add(limit, s.prefix[i])
	scale.Inv(scale)
	last, _ := slices.BinarySearchFunc(s.prefix[i+1:], limit, (*big.Rat).Cmp)
	var out []groupSpan
	plain := new(big.Rat)
	for j := i + 1; j <= i+last; j++ {
		if !s.step(1) {
			return nil
		}
		plain.Sub(s.prefix[j], s.prefix[i])
		plain.Mul(plain, scale)
		if unit, ok := s.signRat(plain, chain); ok {
			out = append(out, groupSpan{end: j, unit: unit})
		}
	}
	s.spans[key] = out
	return out
}

// freeGroups returns the groups of the allowed ratio at index k with the unit
// of sp whose contents are any cover of all of ds[i:sp.end] under inner. The
// groups are cached, so each content cover becomes a group once.
func (s *readingSearch) freeGroups(k int, sp groupSpan, i int, inner *chainInfo) []groupCover {
	key := groupKey{i: i, end: sp.end, chain: inner}
	if cached, ok := s.free[key]; ok {
		return cached
	}
	var out []groupCover
	for _, c := range s.cover(i, sp.end, -1, inner) {
		if !s.step((1 + c.used) * chainSteps(inner.depth)) {
			return nil
		}
		if g, ok := groupOf(s.allowed[k], sp.unit, c); ok {
			out = append(out, g)
		}
	}
	s.free[key] = out
	return out
}

// groupOf returns the group of ratio t in unit over the content cover c, as
// an item of covers. It applies every check of [NewGroup] but does not copy
// the contents: covers and their items are immutable and only read, and
// Reading.Contents returns a deep copy.
func groupOf(t tuplet.Tuplet, unit Written, c coverResult) (groupCover, bool) {
	g, err := newGroup("Readings", t, unit, c.items.slice(), false)
	if err != nil {
		return groupCover{}, false
	}
	return groupCover{item: Content{group: &g}, used: c.used, groups: 1 + c.groups}, true
}

// signFor returns the single written sign of ds[i] under chain, within the
// base and dot bounds.
func (s *readingSearch) signFor(i int, chain *chainInfo) (Written, bool) {
	if chain.coef == nil {
		return Written{}, false
	}
	v := new(big.Rat).Sub(s.prefix[i+1], s.prefix[i])
	return s.signRat(v.Quo(v, chain.coef), chain)
}

// signRat returns the single written sign whose plain value, without the
// chain coefficient, is v, written under chain within the base and dot
// bounds, or false. It only reads v. The written values are cached per chain.
func (s *readingSearch) signRat(v *big.Rat, chain *chainInfo) (Written, bool) {
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
	key := [2]int{int(b), k}
	if w, ok := chain.signs[key]; ok {
		return w, w.IsSet()
	}
	if !s.step(chainSteps(chain.depth)) {
		return Written{}, false
	}
	// As makeWritten, with the coefficient and the packed form of the chain
	// already at hand.
	var w Written
	if value, err := writtenValue("Readings", b, k, chain.coef); err == nil {
		w = Written{base: b, dots: k, chain: chain.packed, value: value}
	}
	if chain.signs == nil {
		chain.signs = make(map[[2]int]Written)
	}
	chain.signs[key] = w
	return w, w.IsSet()
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

// compareContents compares item slices lexicographically by compareContent;
// a shorter slice that is a prefix of the other comes first.
func (s *readingSearch) compareContents(a, b []Content) int {
	for i := 0; i < len(a) && i < len(b); i++ {
		if c := s.compareContent(a[i], b[i]); c != 0 {
			return c
		}
	}
	return cmpInt(len(a), len(b))
}

// compareContent orders items: written values before groups, written values
// by Written.Compare, and groups by actual count, normal count, unit, and
// contents. Each item visited counts one step toward the work bound; a group
// shared by both items compares equal at once.
func (s *readingSearch) compareContent(a, b Content) int {
	if !s.step(1) {
		return 0
	}
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
	if ga == gb {
		return 0
	}
	if c := cmpInt(ga.t.MustActual(), gb.t.MustActual()); c != 0 {
		return c
	}
	if c := cmpInt(ga.t.MustNormal(), gb.t.MustNormal()); c != 0 {
		return c
	}
	if c := ga.unit.Compare(gb.unit); c != 0 {
		return c
	}
	return s.compareContents(ga.contents, gb.contents)
}
