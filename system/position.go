package system

import (
	"cmp"
	"iter"
	"slices"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/spelling"
)

// Step is a signed number of steps of a system: a difference of positions or a
// shift. It is not a semitone and not a number of steps from a reference
// pitch. Step(0) is zero steps.
type Step int

// AdmitsPitch reports whether the name of p is admissible in s.
//
// It returns an error that matches ErrInvalid when s or p is unset.
func (s *System) AdmitsPitch(p pitch.Pitch) (bool, error) {
	if err := s.checkPitches(p); err != nil {
		return false, err
	}

	_, ok := s.shift(p.Spelling())

	return ok, nil
}

// MustAdmitsPitch is like AdmitsPitch but panics on error.
func (s *System) MustAdmitsPitch(p pitch.Pitch) bool {
	return must(s.AdmitsPitch(p))
}

// PitchClass returns the class of the name of p.
//
// It returns an error that matches ErrInvalid when s or p is unset, and an
// error that matches ErrInadmissible when the name of p is not admissible.
func (s *System) PitchClass(p pitch.Pitch) (pitchclass.Class, error) {
	_, kappa, err := s.position(p)
	if err != nil {
		return pitchclass.Class{}, err
	}

	return pitchclass.MustNew(int(kappa), s.n), nil
}

// MustPitchClass is like PitchClass but panics on error.
func (s *System) MustPitchClass(p pitch.Pitch) pitchclass.Class {
	return must(s.PitchClass(p))
}

// Compare compares x and y by position and returns −1, 0 or +1. It compares
// the Octaves first and then the classes, so it never overflows.
//
// It returns an error that matches ErrInvalid when s, x or y is unset, and then
// an error that matches ErrInadmissible when a name is not admissible.
func (s *System) Compare(x, y pitch.Pitch) (int, error) {
	if err := s.checkPitches(x, y); err != nil {
		return 0, err
	}

	ox, kx, err := s.position(x)
	if err != nil {
		return 0, err
	}

	oy, ky, err := s.position(y)
	if err != nil {
		return 0, err
	}

	return comparePositions(ox, kx, oy, ky), nil
}

// MustCompare is like Compare but panics on error.
func (s *System) MustCompare(x, y pitch.Pitch) int {
	return must(s.Compare(x, y))
}

// Between reports whether x is neither below the lower nor above the upper of
// a and b by position. Both ends are included, and their order does not
// matter.
//
// It returns an error that matches ErrInvalid when s or a pitch is unset, and
// then an error that matches ErrInadmissible when a name is not admissible.
func (s *System) Between(x, a, b pitch.Pitch) (bool, error) {
	if err := s.checkPitches(x, a, b); err != nil {
		return false, err
	}

	ox, kx, err := s.position(x)
	if err != nil {
		return false, err
	}

	oa, ka, err := s.position(a)
	if err != nil {
		return false, err
	}

	ob, kb, err := s.position(b)
	if err != nil {
		return false, err
	}

	if comparePositions(oa, ka, ob, kb) > 0 {
		oa, ka, ob, kb = ob, kb, oa, ka
	}

	return comparePositions(oa, ka, ox, kx) <= 0 && comparePositions(ox, kx, ob, kb) <= 0, nil
}

// MustBetween is like Between but panics on error.
func (s *System) MustBetween(x, a, b pitch.Pitch) bool {
	return must(s.Between(x, a, b))
}

// Sorted returns a new slice with the pitches of ps in increasing order of
// position; pitches at the same position keep their order (the sort is
// stable). ps is neither changed nor kept. An empty ps gives an empty result.
//
// It returns an error that matches ErrInvalid when s or some pitch is unset,
// and otherwise an error that matches ErrInadmissible when some name is not
// admissible.
func (s *System) Sorted(ps []pitch.Pitch) ([]pitch.Pitch, error) {
	if err := s.checkPitches(ps...); err != nil {
		return nil, err
	}

	type keyed struct {
		p     pitch.Pitch
		o     int
		kappa int64
	}

	keys := make([]keyed, len(ps))

	for i, p := range ps {
		o, kappa, err := s.position(p)
		if err != nil {
			return nil, err
		}

		keys[i] = keyed{p: p, o: o, kappa: kappa}
	}

	slices.SortStableFunc(keys, func(x, y keyed) int {
		return comparePositions(x.o, x.kappa, y.o, y.kappa)
	})

	out := make([]pitch.Pitch, len(ps))
	for i, k := range keys {
		out[i] = k.p
	}

	return out, nil
}

// MustSorted is like Sorted but panics on error.
func (s *System) MustSorted(ps []pitch.Pitch) []pitch.Pitch {
	return must(s.Sorted(ps))
}

// Distance returns the distance in steps from from to to: the position of to
// minus the position of from, (Octave(to) − Octave(from))·N + class(to) −
// class(from). Distance(y, x) = −Distance(x, y).
//
// Errors, in this order: ErrInvalid when s or a pitch is unset;
// ErrInadmissible when a name is not admissible; ErrOutOfRange when the
// result does not fit in int. Only the result is checked: an intermediate
// product outside int is not an error.
func (s *System) Distance(from, to pitch.Pitch) (Step, error) {
	if err := s.checkPitches(from, to); err != nil {
		return 0, err
	}

	of, kf, err := s.position(from)
	if err != nil {
		return 0, err
	}

	ot, kt, err := s.position(to)
	if err != nil {
		return 0, err
	}

	d, ok := subInt(ot, of)
	if !ok {
		return 0, errorf(ErrOutOfRange, "distance from %v to %v", from, to)
	}

	r, ok := mulAdd(int64(d), int64(s.n), kt-kf)
	if !ok {
		return 0, errorf(ErrOutOfRange, "distance from %v to %v", from, to)
	}

	return Step(r), nil
}

// MustDistance is like Distance but panics on error.
func (s *System) MustDistance(from, to pitch.Pitch) Step {
	return must(s.Distance(from, to))
}

// Equivalent reports whether x and y are enharmonically equivalent in s:
// whether they have the same position. Errors are those of Compare.
func (s *System) Equivalent(x, y pitch.Pitch) (bool, error) {
	c, err := s.Compare(x, y)
	if err != nil {
		return false, err
	}

	return c == 0, nil
}

// MustEquivalent is like Equivalent but panics on error.
func (s *System) MustEquivalent(x, y pitch.Pitch) bool {
	return must(s.Equivalent(x, y))
}

// EnharmonicPitches returns the pitches (name, Octave of p) for the names of
// EnharmonicNames(p.Spelling(), limit), in the same order; all of them have
// the position of p, and p itself is included when its |a| <= limit.
//
// Errors, in this order: ErrInvalid when s or p is unset; ErrLimit when limit
// is negative; ErrInadmissible when the name of p is not admissible.
func (s *System) EnharmonicPitches(p pitch.Pitch, limit alteration.Alteration) (iter.Seq[pitch.Pitch], error) {
	if err := s.checkPitches(p); err != nil {
		return nil, err
	}

	names, err := s.EnharmonicNames(p.Spelling(), limit)
	if err != nil {
		return nil, err
	}

	o := p.Octave()

	return func(yield func(pitch.Pitch) bool) {
		for name := range names {
			if !yield(pitch.MustNew(name, o)) {
				return
			}
		}
	}, nil
}

// MustEnharmonicPitches is like EnharmonicPitches but panics on error.
func (s *System) MustEnharmonicPitches(p pitch.Pitch, limit alteration.Alteration) iter.Seq[pitch.Pitch] {
	return must(s.EnharmonicPitches(p, limit))
}

// Move returns the pitch at the position of p plus n steps, with the name
// chosen by rule; the name is chosen again even for n = 0.
//
// Errors, in this order: ErrInvalid when s, p or rule is unset;
// ErrNotApplicable when rule does not apply to s; ErrInadmissible when the
// name of p or the tonic of rule is not admissible; ErrNoSpelling or
// ErrBadSpelling when rule gives no name; ErrOutOfRange when the Octave of the
// result does not fit in int.
func (s *System) Move(p pitch.Pitch, n Step, rule SpellingRule) (pitch.Pitch, error) {
	if err := s.checkPitches(p); err != nil {
		return pitch.Pitch{}, err
	}

	sp, err := s.prepare(rule)
	if err != nil {
		return pitch.Pitch{}, err
	}

	o, kappa, err := s.position(p)
	if err != nil {
		return pitch.Pitch{}, err
	}

	nn := int64(s.n)
	q, r := floorDiv(int64(n), nn), floorMod(int64(n), nn)

	kappa += r
	if kappa >= nn {
		kappa -= nn
		q++
	}

	name, err := sp.spell(kappa)
	if err != nil {
		return pitch.Pitch{}, err
	}

	to, ok := addInt(o, int(q))
	if !ok {
		return pitch.Pitch{}, errorf(ErrOutOfRange, "%v moved by %d steps", p, n)
	}

	return pitch.MustNew(name, octave.FromASPN(to)), nil
}

// MustMove is like Move but panics on error.
func (s *System) MustMove(p pitch.Pitch, n Step, rule SpellingRule) pitch.Pitch {
	return must(s.Move(p, n, rule))
}

// Walk returns the pitches at the positions from the position of from to the
// position of to, both included, by one step up or down, with names chosen by
// rule; from and to at the same position give one pitch.
//
// All errors are reported before the sequence is returned, in this order:
// ErrInvalid when s, from, to or rule is unset; ErrNotApplicable when rule
// does not apply to s; ErrInadmissible when the name of from or to, or the
// tonic of rule, is not admissible; ErrNoSpelling or ErrBadSpelling when rule
// gives no name for some class on the way. ErrOutOfRange does not occur. The
// rule is applied in advance to each class on the way, at most N times; the
// returned sequence is lazy, can be stopped at any element, and yields the
// same pitches when iterated again.
func (s *System) Walk(from, to pitch.Pitch, rule SpellingRule) (iter.Seq[pitch.Pitch], error) {
	if err := s.checkPitches(from, to); err != nil {
		return nil, err
	}

	sp, err := s.prepare(rule)
	if err != nil {
		return nil, err
	}

	of, kf, err := s.position(from)
	if err != nil {
		return nil, err
	}

	ot, kt, err := s.position(to)
	if err != nil {
		return nil, err
	}

	nn := int64(s.n)
	dir := int64(comparePositions(ot, kt, of, kf))

	count := nn // number of distinct classes on the way, at most N
	if d, ok := subInt(ot, of); ok {
		if dist, ok := mulAdd(int64(d), nn, kt-kf); ok && dist > -nn && dist < nn {
			count = min(nn, max(dist, -dist)+1)
		}
	}

	names := make([]spelling.Spelling, count)
	for i := range names {
		kappa := floorMod(kf+dir*int64(i), nn)

		name, err := sp.spell(kappa)
		if err != nil {
			return nil, err
		}

		names[i] = name
	}

	return func(yield func(pitch.Pitch) bool) {
		o, kappa := of, kf

		for i := int64(0); ; i++ {
			if !yield(pitch.MustNew(names[i%count], octave.FromASPN(o))) {
				return
			}

			if o == ot && kappa == kt {
				return
			}

			kappa += dir

			switch {
			case kappa >= nn:
				kappa, o = 0, o+1
			case kappa < 0:
				kappa, o = nn-1, o-1
			}
		}
	}, nil
}

// MustWalk is like Walk but panics on error.
func (s *System) MustWalk(from, to pitch.Pitch, rule SpellingRule) iter.Seq[pitch.Pitch] {
	return must(s.Walk(from, to, rule))
}

// checkPitches checks that s and every pitch are set.
func (s *System) checkPitches(ps ...pitch.Pitch) error {
	if !s.IsValid() {
		return errorf(ErrInvalid, "unset system")
	}

	for _, p := range ps {
		if !p.IsValid() {
			return errorf(ErrInvalid, "unset pitch")
		}
	}

	return nil
}

// position returns the Octave (ASPN) and the class of the pitch p, or an error
// that matches ErrInvalid for an unset s or name, or ErrInadmissible. The
// caller has checked that p is set.
func (s *System) position(p pitch.Pitch) (o int, kappa int64, err error) {
	sh, err := s.admittedShift(p.Spelling())
	if err != nil {
		return 0, 0, err
	}

	return p.Octave().MustASPN(), floorMod(sh, int64(s.n)), nil
}

// comparePositions compares the positions (ox, kx) and (oy, ky).
func comparePositions(ox int, kx int64, oy int, ky int64) int {
	if c := cmp.Compare(ox, oy); c != 0 {
		return c
	}

	return cmp.Compare(kx, ky)
}
