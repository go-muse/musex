package interval

import (
	"iter"
	"math/big"
	"slices"
	"strconv"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

// Generic is an immutable full letter size, without quality or direction.
// Zero is unset. It is comparable and safe to copy and read concurrently.
type Generic struct{ size int }

// NewGeneric constructs a size in 1..MaxSize, otherwise ErrInvalidArgument.
func NewGeneric(size int) (Generic, error) {
	if size < 1 {
		return Generic{}, ErrInvalidArgument
	}
	return Generic{size}, nil
}

// MustNewGeneric panics with the error returned by NewGeneric.
func MustNewGeneric(size int) Generic { return must(NewGeneric(size)) }

// IsValid reports whether the size is set.
func (g Generic) IsValid() bool { return g.size > 0 }

// Size returns the full size, or zero when unset.
func (g Generic) Size() int { return g.size }

// String returns the decimal full size or "Generic(unset)".
func (g Generic) String() string {
	if !g.IsValid() {
		return "Generic(unset)"
	}
	return strconv.Itoa(g.size)
}

// Generic extracts the full size, including zero for an unset interval.
func (i Interval) Generic() Generic { return Generic{i.size} }

// Palette is an immutable choice of deviations. Zero is DefaultPalette.
// Copies may share private immutable slices; inputs are always snapshotted.
// Palette does not define public equality.
type Palette struct {
	kind   uint8
	limit  int
	values []alteration.Alteration
}

// DefaultPalette contains basic qualities, up to double alterations in the
// perfect group and single alterations in the major-minor group.
func DefaultPalette() Palette { return Palette{} }

// Through selects basic qualities and multiplicities through k, inclusive.
func Through(k int) (Palette, error) {
	if k < 0 || k > alteration.MaxMagnitude {
		return Palette{}, ErrInvalidArgument
	}
	return Palette{kind: 1, limit: k}, nil
}

// MustThrough panics with the error returned by Through.
func MustThrough(k int) Palette { return must(Through(k)) }
func sortedDeviations(v []alteration.Alteration) []alteration.Alteration {
	v = slices.Clone(v)
	slices.SortFunc(v, alteration.Alteration.Compare)
	return slices.Compact(v)
}

// Deviations selects exactly the given deviations; no arguments selects none.
func Deviations(values ...alteration.Alteration) Palette {
	return Palette{kind: 2, values: sortedDeviations(values)}
}

// WithDeviations adds exact deviations to a snapshot of this palette.
func (p Palette) WithDeviations(values ...alteration.Alteration) Palette {
	p.values = sortedDeviations(append(slices.Clone(p.values), values...))
	return p
}
func (p Palette) limits(perfect bool) (int, int, bool) {
	if p.kind == 2 {
		return 0, 0, false
	}
	k := p.limit
	if p.kind == 0 {
		k = 1
		if perfect {
			k = 2
		}
	}
	lo := -k
	if !perfect {
		lo = max(-alteration.MaxMagnitude, -1-k)
	}
	return lo, k, true
}
func (p Palette) contains(i Interval) bool {
	lo, hi, base := p.limits(perfectSize(i.size))
	if base && i.dev.IsInteger() {
		n, _ := i.dev.Fraction()
		if n >= lo && n <= hi {
			return true
		}
	}
	_, ok := slices.BinarySearchFunc(p.values, i.dev, alteration.Alteration.Compare)
	return ok
}
func checkDirection(dir Direction) error {
	if dir == DirectionUnset {
		return ErrNoDirection
	}
	if !dir.isSet() {
		return ErrInvalidArgument
	}
	return nil
}

// Variants lazily enumerates unique admissible qualities in increasing order.
// The sequence is repeatable, concurrent-readable, and supports early stop.
func (g Generic) Variants(dir Direction, p Palette) (iter.Seq[Interval], error) {
	if !g.IsValid() {
		return nil, ErrUnset
	}
	if err := checkDirection(dir); err != nil {
		return nil, err
	}
	return func(yield func(Interval) bool) {
		lo, hi, base := p.limits(perfectSize(g.size))
		if g.size == 1 {
			lo = max(lo, 0)
		}
		j := 0
		emit := func(d alteration.Alteration) bool {
			if g.size == 1 && d.Sign() < 0 {
				return true
			}
			return yield(newInterval(g.size, d, dir == Down))
		}
		if base {
			for n := lo; n <= hi; n++ {
				d := alteration.MustFromInt(n)
				for j < len(p.values) && p.values[j].Compare(d) < 0 {
					if !emit(p.values[j]) {
						return
					}
					j++
				}
				if !emit(d) {
					return
				}
				if j < len(p.values) && p.values[j] == d {
					j++
				}
			}
		}
		for ; j < len(p.values); j++ {
			if !emit(p.values[j]) {
				return
			}
		}
	}, nil
}

// MustVariants panics with the error returned by Variants.
func (g Generic) MustVariants(dir Direction, p Palette) iter.Seq[Interval] {
	return must(g.Variants(dir, p))
}

// PitchVariant pairs an interval with its written endpoint.
type PitchVariant struct {
	Interval Interval
	Pitch    pitch.Pitch
}

// FromPitch enumerates variants whose endpoints are representable and
// admitted. Invalid initial inputs fail before a nonnil iterator is returned.
func (g Generic) FromPitch(sys *system.System, start pitch.Pitch, dir Direction, p Palette) (iter.Seq[PitchVariant], error) {
	if !g.IsValid() || !sys.IsValid() || !start.IsValid() {
		return nil, ErrUnset
	}
	seq, err := g.Variants(dir, p)
	if err != nil {
		return nil, err
	}
	if !admits(sys, start.Spelling()) {
		return nil, system.ErrInadmissible
	}
	return func(yield func(PitchVariant) bool) {
		for i := range seq {
			end, err := TransposePitch(sys, start, i)
			if err == nil && !yield(PitchVariant{i, end}) {
				return
			}
		}
	}, nil
}

// MustFromPitch panics with the error returned by FromPitch.
func (g Generic) MustFromPitch(sys *system.System, start pitch.Pitch, dir Direction, p Palette) iter.Seq[PitchVariant] {
	return must(g.FromPitch(sys, start, dir, p))
}
func checkResolve(c Chromatic, start spelling.Spelling, dir Direction) error {
	if !c.IsValid() {
		return ErrUnset
	}
	if err := checkDirection(dir); err != nil {
		return err
	}
	if !start.IsValid() {
		if !isRegular(c.sys) {
			return ErrStartRequired
		}
	} else if !admits(c.sys, start) {
		return system.ErrInadmissible
	}
	return nil
}
func ratInt(n int) *big.Rat { return new(big.Rat).SetInt64(int64(n)) }
func ratAlteration(a alteration.Alteration) *big.Rat {
	n, d := a.Fraction()
	return new(big.Rat).SetFrac64(int64(n), int64(d))
}
func fromRat(r *big.Rat) (alteration.Alteration, bool) {
	if !r.Num().IsInt64() || !r.Denom().IsInt64() {
		return alteration.Alteration{}, false
	}
	a, err := alteration.FromFraction(int(r.Num().Int64()), int(r.Denom().Int64()))
	return a, err == nil
}

// resolveBase gives the unsigned natural-letter distance and the deviation
// of those natural letters from the basic quality, for one residue.
func resolveBase(sys *system.System, start spelling.Spelling, dir Direction, r int) (int64, int64) {
	l := letter.C
	if start.IsValid() {
		l = start.Letter()
		if dir == Down {
			l = l.Shift(-r)
		}
	}
	d := int64(0)
	if start.IsValid() {
		d = delta(l, r)
	}
	return stepsAfter(sys, l, r), d
}

// Resolve recovers the exact quality for this full size and written direction.
// Intermediate arithmetic is exact even beyond int. start is required only
// for irregular systems. An incompatible unison direction is invalid.
func (g Generic) Resolve(c Chromatic, start spelling.Spelling, dir Direction) (Interval, error) {
	if !g.IsValid() {
		return Interval{}, ErrUnset
	}
	if err := checkResolve(c, start, dir); err != nil {
		return Interval{}, err
	}
	return g.resolve(c, start, dir)
}
func (g Generic) resolve(c Chromatic, start spelling.Spelling, dir Direction) (Interval, error) {
	r := (g.size - 1) % 7
	q := (g.size - 1) / 7
	local, delta := resolveBase(c.sys, start, dir, r)
	base := new(big.Int).Mul(big.NewInt(int64(q)), big.NewInt(int64(c.sys.Classes())))
	base.Add(base, big.NewInt(local))
	target := big.NewInt(int64(c.steps))
	if dir == Down {
		target.Neg(target)
	}
	difference := new(big.Rat).SetInt(new(big.Int).Sub(target, base))
	difference.Quo(difference, ratInt(c.sys.Chromatic()))
	dRat := new(big.Rat).Add(difference, new(big.Rat).SetInt64(delta))
	d, ok := fromRat(dRat)
	if !ok {
		return Interval{}, alterationRangeError("Resolve")
	}
	if g.size == 1 && d.Sign() < 0 {
		return Interval{}, ErrInvalidArgument
	}
	if start.IsValid() {
		change := new(big.Rat).Set(difference)
		if dir == Down {
			change.Neg(change)
		}
		a, ok := fromRat(change.Add(change, ratAlteration(start.Alteration())))
		if !ok {
			return Interval{}, alterationRangeError("Resolve")
		}
		lr := r
		if dir == Down {
			lr = -lr
		}
		if !admits(c.sys, spelling.MustNew(start.Letter().Shift(lr), a)) {
			return Interval{}, system.ErrInadmissible
		}
	}
	return newInterval(g.size, d, dir == Down), nil
}

// MustResolve panics with the error returned by Resolve.
func (g Generic) MustResolve(c Chromatic, start spelling.Spelling, dir Direction) Interval {
	return must(g.Resolve(c, start, dir))
}

func (p Palette) extremes() (alteration.Alteration, alteration.Alteration, bool) {
	lo, hi, ok := p.limits(false)
	minD, maxD := alteration.Natural(), alteration.Natural()
	if ok {
		lo2, hi2, _ := p.limits(true)
		minD = alteration.MustFromInt(min(lo, lo2))
		maxD = alteration.MustFromInt(max(hi, hi2))
	}
	for _, d := range p.values {
		if !ok || d.Compare(minD) < 0 {
			minD = d
		}
		if !ok || d.Compare(maxD) > 0 {
			maxD = d
		}
		ok = true
	}
	return minD, maxD, ok
}
func floorRat(r *big.Rat) *big.Int { return new(big.Int).Div(r.Num(), r.Denom()) }
func ceilRat(r *big.Rat) *big.Int  { return new(big.Int).Neg(floorRat(new(big.Rat).Neg(r))) }

// ForShift enumerates successful Resolve results in increasing full size,
// filtering by p. It bounds the octave window analytically, never scanning
// 1..MaxSize. Validation precedes iteration; candidate failures are skipped.
func ForShift(c Chromatic, start spelling.Spelling, dir Direction, p Palette) (iter.Seq[Interval], error) {
	if err := checkResolve(c, start, dir); err != nil {
		return nil, err
	}
	minD, maxD, nonempty := p.extremes()
	if !nonempty {
		return func(func(Interval) bool) {}, nil
	}
	first, last := big.NewInt(int64((MaxSize-1)/7)), big.NewInt(-1)
	for r := range 7 {
		local, delta := resolveBase(c.sys, start, dir, r)
		t := ratInt(int(c.steps))
		if dir == Down {
			t.Neg(t)
		}
		t.Sub(t, new(big.Rat).SetInt64(local))
		t.Add(t, new(big.Rat).Mul(new(big.Rat).SetInt64(delta), ratInt(c.sys.Chromatic())))
		lower := new(big.Rat).Sub(t, new(big.Rat).Mul(ratAlteration(maxD), ratInt(c.sys.Chromatic())))
		lower.Quo(lower, ratInt(c.sys.Classes()))
		upper := new(big.Rat).Sub(t, new(big.Rat).Mul(ratAlteration(minD), ratInt(c.sys.Chromatic())))
		upper.Quo(upper, ratInt(c.sys.Classes()))
		l, h := ceilRat(lower), floorRat(upper)
		if l.Cmp(first) < 0 {
			first = l
		}
		if h.Cmp(last) > 0 {
			last = h
		}
	}
	if first.Sign() < 0 {
		first.SetInt64(0)
	}
	maxQ := big.NewInt(int64((MaxSize - 1) / 7))
	if last.Cmp(maxQ) > 0 {
		last = maxQ
	}
	if first.Cmp(last) > 0 {
		return func(func(Interval) bool) {}, nil
	}
	a, b := int(first.Int64()), int(last.Int64())
	return func(yield func(Interval) bool) {
		for q := a; q <= b; q++ {
			for r := range 7 {
				if q == (MaxSize-1)/7 && r > (MaxSize-1)%7 {
					return
				}
				g := Generic{7*q + r + 1}
				i, err := g.resolve(c, start, dir)
				if err == nil && p.contains(i) && !yield(i) {
					return
				}
			}
		}
	}, nil
}

// MustForShift panics with the error returned by ForShift.
func MustForShift(c Chromatic, start spelling.Spelling, dir Direction, p Palette) iter.Seq[Interval] {
	return must(ForShift(c, start, dir, p))
}
