package pcset

import (
	"cmp"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
)

// IntervalVector is an immutable interval-class vector <m_1, m_2, ..., m_L> of
// length L = floor(N/2) for a pitch-class set in Z_N (1 <= N <= [MaxModulus]).
//
// IntervalVector is a value type that supports the built-in == operator and can
// be used as a map key. Its zero value IntervalVector{} is invalid
// (IsValid() == false).
type IntervalVector struct {
	n      int
	counts string
}

// IntervalVector returns the interval-class vector of s. For the zero value
// Set{}, it returns the invalid zero IntervalVector{}.
func (s Set) IntervalVector() IntervalVector {
	if !s.IsValid() {
		return IntervalVector{}
	}

	l := s.n / 2
	if l == 0 {
		return IntervalVector{n: s.n}
	}

	k := s.Cardinality()
	if k <= 1 {
		return IntervalVector{n: s.n}
	}

	raw := make([]uint32, l)

	if k <= l {
		accumulatePairIntervals(s.Numbers(), s.n, raw)
	} else {
		compNums := s.Complement().Numbers()
		accumulatePairIntervals(compNums, s.n, raw)

		delta := uint32(2*k - s.n)
		halfLimit := (s.n - 1) / 2

		for d := 1; d <= halfLimit; d++ {
			raw[d-1] += delta
		}

		if s.n%2 == 0 {
			raw[l-1] += uint32(k - l)
		}
	}

	return makeIntervalVector(s.n, raw)
}

// IsValid reports whether v is a valid interval-class vector (not the zero
// value IntervalVector{}).
func (v IntervalVector) IsValid() bool {
	return v.n >= 1
}

// Modulus returns the modulus N of v, or 0 if v is IntervalVector{}.
func (v IntervalVector) Modulus() int {
	return v.n
}

// Len returns the number of interval classes floor(N/2) in v, or 0 if v is
// IntervalVector{} or has Modulus() == 1.
func (v IntervalVector) Len() int {
	return v.n / 2
}

// Count returns the multiplicity m_d of interval class d (1 <= d <= v.Len()).
//
// Count checks errors in the following order:
//  1. !v.IsValid(): [ErrInvalidValue]
//  2. intervalClass < 1 || intervalClass > v.Len(): [ErrInvalidIntervalClass]
func (v IntervalVector) Count(intervalClass int) (int, error) {
	if !v.IsValid() {
		return 0, errInvalidValue()
	}

	l := v.n / 2
	if intervalClass < 1 || intervalClass > l {
		return 0, fmt.Errorf("%w: interval class %d out of range [1, %d]", ErrInvalidIntervalClass, intervalClass, l)
	}

	return v.countAtUnchecked(intervalClass), nil
}

// MustCount is like [IntervalVector.Count] but panics on error.
func (v IntervalVector) MustCount(intervalClass int) int {
	c, err := v.Count(intervalClass)
	if err != nil {
		panic(err)
	}

	return c
}

// Total returns the sum of all interval-class counts sum_{d=1}^{floor(N/2)} m_d,
// which equals C(|X|, 2). For IntervalVector{}, it returns 0.
func (v IntervalVector) Total() int {
	if !v.IsValid() {
		return 0
	}

	total := 0
	l := v.n / 2

	for d := 1; d <= l; d++ {
		total += v.countAtUnchecked(d)
	}

	return total
}

// Slice returns a newly allocated slice [m_1, m_2, ..., m_{floor(N/2)}]. For
// N == 1 it returns a non-nil empty slice []int{}; for IntervalVector{} it
// returns nil.
func (v IntervalVector) Slice() []int {
	if !v.IsValid() {
		return nil
	}

	l := v.n / 2
	res := make([]int, l)

	for d := 1; d <= l; d++ {
		res[d-1] = v.countAtUnchecked(d)
	}

	return res
}

// Equal reports whether v and other are both valid interval-class vectors with
// the same modulus and identical counts. If either operand is IntervalVector{},
// Equal returns false.
func (v IntervalVector) Equal(other IntervalVector) bool {
	return v.IsValid() && other.IsValid() && v == other
}

// Compare compares v and other and returns -1, 0, or +1:
//  1. by Modulus() ascending (so IntervalVector{} with Modulus() == 0 is less than any valid vector);
//  2. lexicographically by components m_1, m_2, ..., m_{floor(N/2)}.
func (v IntervalVector) Compare(other IntervalVector) int {
	if c := cmp.Compare(v.n, other.n); c != 0 {
		return c
	}

	l := v.n / 2
	for d := 1; d <= l; d++ {
		if c := cmp.Compare(v.countAtUnchecked(d), other.countAtUnchecked(d)); c != 0 {
			return c
		}
	}

	return 0
}

// String returns the canonical representation "<m_1, m_2, ..., m_L> mod N"
// (or "<> mod 1" when N == 1), or "IntervalVector(invalid)" for IntervalVector{}.
func (v IntervalVector) String() string {
	if !v.IsValid() {
		return "IntervalVector(invalid)"
	}

	return v.formatDelimited() + " mod " + strconv.Itoa(v.n)
}

// Notation returns the delimited literary representation "<m_1, m_2, ..., m_L>"
// without the modulus suffix (or "<>" when N == 1).
//
// It returns [ErrInvalidValue] if !v.IsValid().
func (v IntervalVector) Notation() (string, error) {
	if !v.IsValid() {
		return "", errInvalidValue()
	}

	return v.formatDelimited(), nil
}

// MustNotation is like [IntervalVector.Notation] but panics on error.
func (v IntervalVector) MustNotation() string {
	s, err := v.Notation()
	if err != nil {
		panic(err)
	}

	return s
}

// CompactNotation returns the compact representation "<m_1m_2...m_L>" without
// separators when every component m_d is in 0..9.
//
// CompactNotation checks errors in the following order:
//  1. !v.IsValid(): [ErrInvalidValue]
//  2. any component m_d >= 10: [ErrNotation]
func (v IntervalVector) CompactNotation() (string, error) {
	if !v.IsValid() {
		return "", errInvalidValue()
	}

	l := v.n / 2
	buf := make([]byte, 0, l+2)
	buf = append(buf, '<')

	for d := 1; d <= l; d++ {
		c := v.countAtUnchecked(d)
		if c >= 10 {
			return "", fmt.Errorf("%w: interval count %d >= 10 cannot be formatted in compact notation", ErrNotation, c)
		}

		buf = append(buf, byte('0'+c))
	}

	buf = append(buf, '>')

	return string(buf), nil
}

// MustCompactNotation is like [IntervalVector.CompactNotation] but panics on error.
func (v IntervalVector) MustCompactNotation() string {
	s, err := v.CompactNotation()
	if err != nil {
		panic(err)
	}

	return s
}

func (v IntervalVector) countAtUnchecked(d int) int {
	offset := (d - 1) * 4
	if offset+4 > len(v.counts) {
		return 0
	}

	return int(binary.LittleEndian.Uint32([]byte(v.counts[offset : offset+4])))
}

func (v IntervalVector) formatDelimited() string {
	l := v.n / 2
	if l == 0 {
		return "<>"
	}

	var b strings.Builder
	b.Grow(l*3 + 2)
	b.WriteByte('<')

	for d := 1; d <= l; d++ {
		if d > 1 {
			b.WriteString(", ")
		}

		b.WriteString(strconv.Itoa(v.countAtUnchecked(d)))
	}

	b.WriteByte('>')

	return b.String()
}

func accumulatePairIntervals(nums []int, n int, raw []uint32) {
	for i := range nums {
		xi := nums[i]
		for j := i + 1; j < len(nums); j++ {
			diff := nums[j] - xi
			if diff > n-diff {
				diff = n - diff
			}

			raw[diff-1]++
		}
	}
}

func makeIntervalVector(n int, raw []uint32) IntervalVector {
	end := len(raw)
	for end > 0 && raw[end-1] == 0 {
		end--
	}

	buf := make([]byte, end*4)
	for i := range end {
		binary.LittleEndian.PutUint32(buf[i*4:(i+1)*4], raw[i])
	}

	return IntervalVector{n: n, counts: string(buf)}
}
