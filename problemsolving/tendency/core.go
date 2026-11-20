package tendency

import (
	"cmp"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"slices"
	"strconv"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/mode"
)

var (
	ErrUnsetScheme       = errors.New("unset tendency scheme")
	ErrUnknownScheme     = errors.New("unknown tendency scheme")
	ErrInvalidDefinition = errors.New("invalid tendency definition")
	ErrInvalidMode       = errors.New("invalid tendency mode")
	ErrInvalidPlace      = errors.New("invalid tendency place")
	ErrInvalidInterval   = errors.New("invalid tendency interval")
	ErrInvalidArgument   = errors.New("invalid tendency argument")
	ErrOutOfRange        = errors.New("tendency out of range")
)

// copyValue handles only public report data. Immutable values from dependency
// packages are copied by value, without inspecting their private representation.
func copyValue(v reflect.Value, nonnil bool) reflect.Value {
	switch v.Kind() {
	case reflect.Slice:
		if v.IsNil() && !nonnil {
			return reflect.Zero(v.Type())
		}
		r := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
		for i := 0; i < v.Len(); i++ {
			r.Index(i).Set(copyValue(v.Index(i), nonnil))
		}
		return r
	case reflect.Struct:
		r := reflect.New(v.Type()).Elem()
		r.Set(v)
		for i := 0; i < v.NumField(); i++ {
			if v.Type().Field(i).IsExported() {
				r.Field(i).Set(copyValue(v.Field(i), nonnil))
			}
		}
		return r
	default:
		return v
	}
}
func copied[T any](v T) T              { return copyValue(reflect.ValueOf(v), false).Interface().(T) }
func report[T any](v T) T              { return copyValue(reflect.ValueOf(v), true).Interface().(T) }
func failure(op string, e error) error { return fmt.Errorf("tendency %s: %w", op, e) }
func modeOK(m mode.Mode) error {
	if !m.IsValid() {
		return failure("mode", ErrInvalidMode)
	}
	return nil
}
func numberOK(m mode.Mode, n int) error {
	if e := modeOK(m); e != nil {
		return e
	}
	if n < 1 || n > m.Len() {
		return failure("degree "+strconv.Itoa(n), ErrInvalidPlace)
	}
	return nil
}
func intSum(a, b int) (int, error) {
	z := new(big.Int).Add(big.NewInt(int64(a)), big.NewInt(int64(b)))
	if !z.IsInt64() {
		return 0, failure("cycle", ErrOutOfRange)
	}
	return int(z.Int64()), nil
}
func intDifference(a, b int) (int, error) {
	z := new(big.Int).Sub(big.NewInt(int64(a)), big.NewInt(int64(b)))
	if !z.IsInt64() {
		return 0, failure("cycle difference", ErrOutOfRange)
	}
	return int(z.Int64()), nil
}

type coordinate struct {
	letters   *big.Int
	chromatic *big.Rat
}

var natural = [7]int64{0, 2, 4, 5, 7, 9, 11}

func baseDistance(n *big.Int) *big.Rat {
	q, r := new(big.Int), new(big.Int)
	q.QuoRem(n, big.NewInt(7), r)
	q.Mul(q, big.NewInt(12))
	q.Add(q, big.NewInt(natural[int(r.Int64())]))
	return new(big.Rat).SetInt(q)
}
func coord(i interval.Interval) coordinate {
	n := big.NewInt(int64(i.Size() - 1))
	x := baseDistance(n)
	a, b := i.Deviation().Fraction()
	x.Add(x, new(big.Rat).SetFrac64(int64(a), int64(b)))
	if i.IsDescending() {
		n.Neg(n)
		x.Neg(x)
	}
	return coordinate{n, x}
}
func (a coordinate) add(b coordinate) coordinate {
	return coordinate{new(big.Int).Add(a.letters, b.letters), new(big.Rat).Add(a.chromatic, b.chromatic)}
}
func (a coordinate) sub(b coordinate) coordinate {
	return coordinate{new(big.Int).Sub(a.letters, b.letters), new(big.Rat).Sub(a.chromatic, b.chromatic)}
}
func (a coordinate) mul(n int) coordinate {
	return coordinate{new(big.Int).Mul(a.letters, big.NewInt(int64(n))), new(big.Rat).Mul(a.chromatic, new(big.Rat).SetInt64(int64(n)))}
}
func (a coordinate) value() (interval.Interval, error) {
	n := new(big.Int).Set(a.letters)
	x := new(big.Rat).Set(a.chromatic)
	down := n.Sign() < 0 || (n.Sign() == 0 && x.Sign() < 0)
	if down {
		n.Neg(n)
		x.Neg(x)
	}
	size := new(big.Int).Add(n, big.NewInt(1))
	if !size.IsInt64() || size.Sign() <= 0 {
		return interval.Interval{}, errors.Join(ErrOutOfRange, interval.ErrOutOfRange)
	}
	d := new(big.Rat).Sub(x, baseDistance(n))
	if !d.Num().IsInt64() || !d.Denom().IsInt64() {
		return interval.Interval{}, errors.Join(ErrOutOfRange, interval.ErrOutOfRange, alteration.ErrOutOfRange)
	}
	alt, e := alteration.FromFraction(int(d.Num().Int64()), int(d.Denom().Int64()))
	if e != nil {
		return interval.Interval{}, errors.Join(ErrOutOfRange, interval.ErrOutOfRange, e)
	}
	i, e := interval.FromDeviation(int(size.Int64()), alt)
	if e != nil {
		return interval.Interval{}, errors.Join(ErrOutOfRange, e)
	}
	if down {
		i, e = i.WithDirection(interval.Down)
	}
	return i, e
}
func placeCoord(m mode.Mode, p Place) coordinate {
	c := coord(m.MustDegreeInterval(p.Degree))
	if cyc, ok := m.Cycle(); ok {
		c = c.add(coord(cyc).mul(p.Cycle))
	}
	return c
}

// ValidatePlace checks the actual degree, cycle, and representability of the place coordinate.
func ValidatePlace(m mode.Mode, p Place) error {
	if e := numberOK(m, p.Degree); e != nil {
		return e
	}
	if m.IsOpen() && p.Cycle != 0 {
		return failure("open mode cycle", ErrInvalidPlace)
	}
	_, e := placeCoord(m, p).value()
	return e
}

// Between computes an exact specific interval between two validated places.
func Between(m mode.Mode, from, to Place) (Movement, error) {
	if e := ValidatePlace(m, from); e != nil {
		return Movement{}, e
	}
	if e := ValidatePlace(m, to); e != nil {
		return Movement{}, e
	}
	i, e := placeCoord(m, to).sub(placeCoord(m, from)).value()
	if e != nil {
		return Movement{}, e
	}
	return Movement{From: from, To: to, Interval: i, Direction: motion(i)}, nil
}
func motion(i interval.Interval) Motion {
	if i.IsPerfectUnison() {
		return Stationary
	}
	if i.IsDescending() {
		return Descending
	}
	return Ascending
}
func comparePlace(a, b Place) int {
	if c := cmp.Compare(a.Cycle, b.Cycle); c != 0 {
		return c
	}
	return cmp.Compare(a.Degree, b.Degree)
}
func comparePair(a, b Pair) int {
	if c := comparePlace(a.First, b.First); c != 0 {
		return c
	}
	return comparePlace(a.Second, b.Second)
}
func compareInterval(a, b interval.Interval) int {
	if c := cmp.Compare(motion(a), motion(b)); c != 0 {
		return c
	}
	if c := cmp.Compare(a.Size(), b.Size()); c != 0 {
		return c
	}
	return a.Deviation().Compare(b.Deviation())
}
func compareMotions(a, b Motions) int {
	if c := compareInterval(a.First, b.First); c != 0 {
		return c
	}
	return compareInterval(a.Second, b.Second)
}
func compareTarget(a, b Target) int {
	if a.CycleKnown != b.CycleKnown {
		if a.CycleKnown {
			return -1
		}
		return 1
	}
	if a.CycleKnown {
		if c := cmp.Compare(a.Cycle, b.Cycle); c != 0 {
			return c
		}
	}
	if c := cmp.Compare(a.Degree, b.Degree); c != 0 {
		return c
	}
	return cmp.Compare(a.Kind, b.Kind)
}
func validTask(t Task) bool     { return t >= SupportTask && t <= IntervalResolutionTask }
func validMotion(t Motion) bool { return t >= Stationary && t <= Descending }
func intervalOK(i interval.Interval) error {
	if !i.IsValid() {
		return failure("interval", ErrInvalidInterval)
	}
	return nil
}
func motionsOK(m Motions) error {
	if e := intervalOK(m.First); e != nil {
		return e
	}
	return intervalOK(m.Second)
}
func pairOK(m mode.Mode, p Pair) error {
	if e := ValidatePlace(m, p.First); e != nil {
		return e
	}
	return ValidatePlace(m, p.Second)
}
func totalKnowledge(v []Knowledge) Knowledge {
	if len(v) == 0 {
		return Known
	}
	k := v[0]
	for _, x := range v {
		if x != k {
			return Partial
		}
	}
	return k
}
func reason(rule string, origin Origin, reference, text string, facts ...Fact) Reason {
	slices.SortFunc(facts, func(a, b Fact) int {
		if c := cmp.Compare(a.Name, b.Name); c != 0 {
			return c
		}
		return cmp.Compare(a.Value, b.Value)
	})
	return Reason{rule, origin, reference, text, facts}
}
func sourceReason(rule, reference, text string, facts ...Fact) Reason {
	return reason(rule, SourceRule, reference, text, facts...)
}
func productReason(rule, text string, facts ...Fact) Reason {
	return reason(rule, ProductRule, "business-spec §4–5", text, facts...)
}
func canonicalReasons(v []Reason) []Reason {
	r := copied(v)
	for i := range r {
		slices.SortFunc(r[i].Facts, func(a, b Fact) int {
			if c := cmp.Compare(a.Name, b.Name); c != 0 {
				return c
			}
			return cmp.Compare(a.Value, b.Value)
		})
	}
	slices.SortFunc(r, func(a, b Reason) int {
		if c := cmp.Compare(a.Rule, b.Rule); c != 0 {
			return c
		}
		if c := cmp.Compare(a.Origin, b.Origin); c != 0 {
			return c
		}
		if c := cmp.Compare(a.Reference, b.Reference); c != 0 {
			return c
		}
		if c := cmp.Compare(a.Explanation, b.Explanation); c != 0 {
			return c
		}
		return cmp.Compare(fmt.Sprint(a.Facts), fmt.Sprint(b.Facts))
	})
	return slices.CompactFunc(r, func(a, b Reason) bool { return reflect.DeepEqual(a, b) })
}
func bases(v []Basis) []Reason {
	r := []Reason{}
	for _, b := range v {
		r = append(r, reason("user/"+b.Rule, UserRule, b.Rule, b.Explanation))
		for _, ref := range b.References {
			r = append(r, reason("user/"+b.Rule, UserRule, ref, b.Explanation))
		}
	}
	return canonicalReasons(r)
}
func (s *Scheme) answer(k Knowledge, r ...Reason) Answer {
	return report(Answer{Scheme: s.state.info.ID, Name: s.state.info.Name, Knowledge: k, Reasons: canonicalReasons(r)})
}
func (s *Scheme) check() error {
	if !s.IsValid() {
		return failure("scheme", ErrUnsetScheme)
	}
	return nil
}

// IsValid reports whether the scheme was successfully constructed.
func (s *Scheme) IsValid() bool { return s != nil && s.state != nil }

// Info returns independent metadata, or Info{} for an unset scheme.
func (s *Scheme) Info() Info {
	if !s.IsValid() {
		return Info{}
	}
	return report(s.state.info)
}

// Definition returns an independent canonical user definition; built-ins and unset schemes return false.
func (s *Scheme) Definition() (Definition, bool) {
	if !s.IsValid() || s.state.info.Builtin {
		return Definition{}, false
	}
	return copied(s.state.def), true
}

// Equal compares semantic rule snapshots, treating two unset schemes as equal.
func (s *Scheme) Equal(o *Scheme) bool {
	if !s.IsValid() || !o.IsValid() {
		return !s.IsValid() && !o.IsValid()
	}
	if !reflect.DeepEqual(s.state.info, o.state.info) {
		return false
	}
	if s.state.info.Builtin {
		return true
	}
	a, b := s.state.def, o.state.def
	if a.Mode.IsValid() != b.Mode.IsValid() || (a.Mode.IsValid() && !a.Mode.Equal(b.Mode)) {
		return false
	}
	a.Mode = mode.Mode{}
	b.Mode = mode.Mode{}
	return reflect.DeepEqual(a, b)
}
