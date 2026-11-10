package system

import (
	"iter"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/spelling"
)

// Admits reports whether name is admissible in s: whether its alteration value
// belongs to the rule of s. Every letter is admissible.
//
// It returns an error that matches ErrInvalid when s or name is unset.
func (s *System) Admits(name spelling.Spelling) (bool, error) {
	if err := s.checkName(name); err != nil {
		return false, err
	}

	_, ok := s.shift(name)

	return ok, nil
}

// MustAdmits is like Admits but panics on error.
func (s *System) MustAdmits(name spelling.Spelling) bool {
	return must(s.Admits(name))
}

// Class returns the class of name: (p(l) + a·c) mod N, with the modulus N.
//
// It returns an error that matches ErrInvalid when s or name is unset, and an
// error that matches ErrInadmissible when name is not admissible.
func (s *System) Class(name spelling.Spelling) (pitchclass.Class, error) {
	sh, err := s.admittedShift(name)
	if err != nil {
		return pitchclass.Class{}, err
	}

	return s.class(sh), nil
}

// MustClass is like Class but panics on error.
func (s *System) MustClass(name spelling.Spelling) pitchclass.Class {
	return must(s.Class(name))
}

// Carry returns k = ⌊(p(l) + a·c)/N⌋: by how many octaves of positions the
// name leaves the octave of its natural letter. In the traditional system B♯
// has the carry 1, C♭ −1 and C 0. The written octave of a pitch is its Octave
// minus the carry of its name.
//
// It returns an error that matches ErrInvalid when s or name is unset, and an
// error that matches ErrInadmissible when name is not admissible.
func (s *System) Carry(name spelling.Spelling) (int, error) {
	sh, err := s.admittedShift(name)
	if err != nil {
		return 0, err
	}

	return int(floorDiv(sh, int64(s.n))), nil
}

// MustCarry is like Carry but panics on error.
func (s *System) MustCarry(name spelling.Spelling) int {
	return must(s.Carry(name))
}

// ShiftedClass returns the class of the name (l, a + n): the class of name
// shifted by n·c modulo N.
//
// Errors, in this order: ErrInvalid when s or name is unset; ErrInadmissible
// when name is not admissible; ErrOutOfRange when a + n is outside the range
// of package alteration; ErrInadmissible when a + n is not admissible.
func (s *System) ShiftedClass(name spelling.Spelling, n int) (pitchclass.Class, error) {
	return s.shiftedClass(name, func(sp spelling.Spelling) (spelling.Spelling, error) {
		return sp.RaiseBy(n)
	})
}

// MustShiftedClass is like ShiftedClass but panics on error.
func (s *System) MustShiftedClass(name spelling.Spelling, n int) pitchclass.Class {
	return must(s.ShiftedClass(name, n))
}

// ShiftedClassByAlteration returns the class of the name (l, a + d). The change
// d can be fractional; for the result to be admissible, d·c is an integer.
// Errors are those of ShiftedClass.
func (s *System) ShiftedClassByAlteration(name spelling.Spelling, d alteration.Alteration) (pitchclass.Class, error) {
	return s.shiftedClass(name, func(sp spelling.Spelling) (spelling.Spelling, error) {
		return sp.RaiseByAlteration(d)
	})
}

// MustShiftedClassByAlteration is like ShiftedClassByAlteration but panics on
// error.
func (s *System) MustShiftedClassByAlteration(name spelling.Spelling, d alteration.Alteration) pitchclass.Class {
	return must(s.ShiftedClassByAlteration(name, d))
}

// EquivalentNames reports whether x and y are enharmonically equivalent in s:
// whether they have the same class.
//
// It returns an error that matches ErrInvalid when s, x or y is unset, and
// then an error that matches ErrInadmissible when x or y is not admissible.
func (s *System) EquivalentNames(x, y spelling.Spelling) (bool, error) {
	if err := s.checkName(x); err != nil {
		return false, err
	}

	if !y.IsValid() {
		return false, errorf(ErrInvalid, "unset name")
	}

	sx, err := s.admittedShift(x)
	if err != nil {
		return false, err
	}

	sy, err := s.admittedShift(y)
	if err != nil {
		return false, err
	}

	n := int64(s.n)

	return floorMod(sx, n) == floorMod(sy, n), nil
}

// MustEquivalentNames is like EquivalentNames but panics on error.
func (s *System) MustEquivalentNames(x, y spelling.Spelling) bool {
	return must(s.EquivalentNames(x, y))
}

// Spellings returns all admissible names of the class c with |a| <= limit,
// each exactly once, ordered by letter C to B and, within a letter, by
// increasing alteration value (the order of spelling.Spelling.Compare). The
// sequence is lazy, can be stopped at any element and iterated again; it keeps
// O(MaxProgressions) state. A class without such names gives an empty
// sequence.
//
// Errors, in this order: ErrInvalid when s or c is unset; ErrModulusMismatch
// when the modulus of c is not N; ErrLimit when limit is negative.
func (s *System) Spellings(c pitchclass.Class, limit alteration.Alteration) (iter.Seq[spelling.Spelling], error) {
	if !s.IsValid() {
		return nil, errorf(ErrInvalid, "unset system")
	}

	if !c.IsValid() {
		return nil, errorf(ErrInvalid, "unset class")
	}

	if err := s.checkModulus(c); err != nil {
		return nil, err
	}

	if limit.Sign() < 0 {
		return nil, errorf(ErrLimit, "limit %v", limit)
	}

	return s.spellings(int64(c.Number()), limit), nil
}

// MustSpellings is like Spellings but panics on error.
func (s *System) MustSpellings(c pitchclass.Class, limit alteration.Alteration) iter.Seq[spelling.Spelling] {
	return must(s.Spellings(c, limit))
}

// EnharmonicNames returns the admissible names with the class of name and
// |a| <= limit, in the order of Spellings; it includes name itself when its
// |a| <= limit.
//
// Errors, in this order: ErrInvalid when s or name is unset; ErrLimit when
// limit is negative; ErrInadmissible when name is not admissible.
func (s *System) EnharmonicNames(name spelling.Spelling, limit alteration.Alteration) (iter.Seq[spelling.Spelling], error) {
	if err := s.checkName(name); err != nil {
		return nil, err
	}

	if limit.Sign() < 0 {
		return nil, errorf(ErrLimit, "limit %v", limit)
	}

	sh, err := s.admittedShift(name)
	if err != nil {
		return nil, err
	}

	return s.spellings(floorMod(sh, int64(s.n)), limit), nil
}

// MustEnharmonicNames is like EnharmonicNames but panics on error.
func (s *System) MustEnharmonicNames(name spelling.Spelling, limit alteration.Alteration) iter.Seq[spelling.Spelling] {
	return must(s.EnharmonicNames(name, limit))
}

// checkName checks that s and name are set.
func (s *System) checkName(name spelling.Spelling) error {
	if !s.IsValid() {
		return errorf(ErrInvalid, "unset system")
	}

	if !name.IsValid() {
		return errorf(ErrInvalid, "unset name")
	}

	return nil
}

// checkModulus checks that the set class c has the modulus N.
func (s *System) checkModulus(c pitchclass.Class) error {
	if c.Modulus() != s.n {
		return errorf(ErrModulusMismatch, "class %v of a system with N = %d", c, s.n)
	}

	return nil
}

// shift returns the shift p(l) + a·c of the set name and whether the name is
// admissible. It costs O(MaxProgressions) and does not allocate.
func (s *System) shift(name spelling.Spelling) (int64, bool) {
	m, ok := s.alterationShift(name.Alteration())
	if !ok {
		return 0, false
	}

	return int64(s.offsets[name.Letter().Ordinal()]) + m, true
}

// alterationShift returns m = a·c and whether a is admissible.
func (s *System) alterationShift(a alteration.Alteration) (int64, bool) {
	p, q := a.Fraction()

	c := int64(s.c)
	if c%int64(q) != 0 {
		return 0, false // every admissible value has a denominator dividing c
	}

	m := int64(p) * (c / int64(q))

	for _, sp := range s.shifts {
		if m < sp.lo || m > sp.hi {
			continue
		}

		if sp.stride == 0 || (m-sp.lo)%sp.stride == 0 {
			return m, true
		}
	}

	return 0, false
}

// admittedShift returns the shift of name after the checks of an unset system
// or name (ErrInvalid) and of admissibility (ErrInadmissible).
func (s *System) admittedShift(name spelling.Spelling) (int64, error) {
	if err := s.checkName(name); err != nil {
		return 0, err
	}

	sh, ok := s.shift(name)
	if !ok {
		return 0, errorf(ErrInadmissible, "%v in %v", name, s.name)
	}

	return sh, nil
}

// class returns the class of the shift sh.
func (s *System) class(sh int64) pitchclass.Class {
	return pitchclass.MustNew(int(floorMod(sh, int64(s.n))), s.n)
}

// shiftedClass implements ShiftedClass and ShiftedClassByAlteration.
func (s *System) shiftedClass(name spelling.Spelling,
	change func(spelling.Spelling) (spelling.Spelling, error),
) (pitchclass.Class, error) {
	if _, err := s.admittedShift(name); err != nil {
		return pitchclass.Class{}, err
	}

	t, err := change(name)
	if err != nil {
		return pitchclass.Class{}, errorf(ErrOutOfRange, "changed alteration of %v", name)
	}

	sh, ok := s.shift(t)
	if !ok {
		return pitchclass.Class{}, errorf(ErrInadmissible, "%v in %v", t, s.name)
	}

	return s.class(sh), nil
}

// spellings returns the admissible names of the class kappa with |a| <= limit.
func (s *System) spellings(kappa int64, limit alteration.Alteration) iter.Seq[spelling.Spelling] {
	lp, lq := limit.Fraction()
	bound := floorDiv(int64(lp)*int64(s.c), int64(lq)) // |m| <= bound

	return func(yield func(spelling.Spelling) bool) {
		for _, l := range letter.All() {
			var (
				streams [MaxProgressions]intStream
				n       int
			)

			r := kappa - int64(s.offsets[l.Ordinal()])

			for _, sp := range s.shifts {
				if st, ok := s.residueStream(sp, r, -bound, bound); ok {
					streams[n] = st
					n++
				}
			}

			if !mergeStreams(streams[:n], func(m int64) bool {
				return yield(spelling.MustNew(l, alteration.MustFromFraction(int(m), s.c)))
			}) {
				return
			}
		}
	}
}

// residueStream returns the stream of the points m of sp with m ≡ r (mod N)
// and lower <= m <= upper; ok is false when there are none.
func (s *System) residueStream(sp shiftProgression, r, lower, upper int64) (intStream, bool) {
	start, period, ok := s.residue(sp, r)
	if !ok {
		return intStream{}, false
	}

	lo, hi := max(sp.lo, lower), min(sp.hi, upper)
	lo += floorMod(start-lo, period)
	hi -= floorMod(hi-start, period)

	if lo > hi {
		return intStream{}, false
	}

	return intStream{cur: lo, end: hi, step: period}, true
}

// residue returns the points m of the lattice of sp with m ≡ r (mod N) as
// start + period·Z; ok is false when there are none. Bounds of sp are not
// applied. For a single value the period is 1 and only start is meaningful to
// the caller after applying the bounds.
func (s *System) residue(sp shiftProgression, r int64) (start, period int64, ok bool) {
	n := int64(s.n)

	if sp.stride == 0 {
		if floorMod(sp.lo-r, n) != 0 {
			return 0, 0, false
		}

		return sp.lo, 1, true
	}

	j, e, ok := solveCongruence(sp.stride, r-sp.lo, n)
	if !ok {
		return 0, 0, false
	}

	return sp.lo + j*sp.stride, sp.stride * e, true
}

// intStream yields cur, cur + step, ..., end.
type intStream struct {
	cur, end, step int64
}

// mergeStreams yields the union of the increasing streams in increasing order
// without repeats. It reports false when yield stopped the iteration.
func mergeStreams(streams []intStream, yield func(int64) bool) bool {
	var (
		last    int64
		started bool
	)

	for {
		best := -1

		for i := range streams {
			if streams[i].cur > streams[i].end {
				continue
			}

			if best < 0 || streams[i].cur < streams[best].cur {
				best = i
			}
		}

		if best < 0 {
			return true
		}

		m := streams[best].cur
		streams[best].cur += streams[best].step

		if started && m == last {
			continue
		}

		if !yield(m) {
			return false
		}

		last, started = m, true
	}
}
