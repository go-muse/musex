package fraction

import "math/big"

// Range constants for fraction components and parsed decimal digit sequences.
const (
	// MaxComponentBits is the bit width of the magnitude of a numerator or
	// denominator component (255 magnitude bits plus sign in a 256-bit signed
	// representation).
	MaxComponentBits = 255

	// MaxComponent is the maximum supported magnitude of a numerator or
	// denominator component: 2^255 - 1.
	MaxComponent = (1 << MaxComponentBits) - 1

	// MaxComponentString is the decimal string representation of MaxComponent.
	MaxComponentString = "57896044618658097711785492504343953926634992332820282019728792003956564819967"

	// MaxDigits is the maximum number of digits accepted in a single digit
	// group by Parse and the maximum fractional digit count accepted by
	// Decimal and FormatDecimal.
	MaxDigits = 100
)

// MaxComponentBig returns a new *big.Int equal to [MaxComponent] (2^255 - 1).
func MaxComponentBig() *big.Int {
	return u256Max.toBigInt(false)
}

const (
	flagNumNeg uint8 = 1 << 0
	flagDenNeg uint8 = 1 << 1
)

// Fraction is an immutable exact signed rational number that stores both its
// original numerator and its original denominator in a 256-bit signed range.
//
// The zero value Fraction{} represents an invalid, unset fraction (0/0) for
// which [Fraction.IsValid] returns false. Fraction is comparable with == (by
// stored numerator and denominator pair) and can be used as a map key.
type Fraction struct {
	num   u256
	den   u256
	flags uint8
}

func makeFraction(num, den u256, numNeg, denNeg bool) Fraction {
	var flags uint8
	if numNeg && !num.isZero() {
		flags |= flagNumNeg
	}
	if denNeg && !den.isZero() {
		flags |= flagDenNeg
	}
	return Fraction{
		num:   num,
		den:   den,
		flags: flags,
	}
}

func (f Fraction) numNeg() bool {
	return (f.flags & flagNumNeg) != 0
}

func (f Fraction) denNeg() bool {
	return (f.flags & flagDenNeg) != 0
}

// Zero returns the canonical reduced zero fraction 0/1.
func Zero() Fraction {
	return Fraction{num: u256Zero, den: u256One, flags: 0}
}

// One returns the canonical reduced unit fraction 1/1.
func One() Fraction {
	return Fraction{num: u256One, den: u256One, flags: 0}
}

// MinusOne returns the canonical reduced negative unit fraction -1/1.
func MinusOne() Fraction {
	return Fraction{num: u256One, den: u256One, flags: flagNumNeg}
}

// New constructs a Fraction preserving the supplied numerator and denominator
// pair (including signs and common factors). It returns [ErrZeroDenominator]
// if den == 0.
func New(num, den int) (Fraction, error) {
	if den == 0 {
		return Fraction{}, ErrZeroDenominator
	}
	nMag, nNeg := u256FromInt(num)
	dMag, dNeg := u256FromInt(den)
	return makeFraction(nMag, dMag, nNeg, dNeg), nil
}

// MustNew is like [New] but panics with the returned error if den == 0.
func MustNew(num, den int) Fraction {
	f, err := New(num, den)
	if err != nil {
		panic(err)
	}
	return f
}

// NewReduced constructs a Fraction reduced to canonical form (positive
// denominator, coprime numerator and denominator, 0/1 for zero). It returns
// [ErrZeroDenominator] if den == 0.
func NewReduced(num, den int) (Fraction, error) {
	if den == 0 {
		return Fraction{}, ErrZeroDenominator
	}
	if num == 0 {
		return Zero(), nil
	}
	nMag, nNeg := u256FromInt(num)
	dMag, dNeg := u256FromInt(den)
	g := gcdUint64(nMag[0], dMag[0])
	return makeFraction(u256{nMag[0] / g}, u256{dMag[0] / g}, nNeg != dNeg, false), nil
}

// MustNewReduced is like [NewReduced] but panics with the returned error if
// den == 0.
func MustNewReduced(num, den int) Fraction {
	f, err := NewReduced(num, den)
	if err != nil {
		panic(err)
	}
	return f
}

// FromInt constructs the reduced integer fraction n/1.
func FromInt(n int) Fraction {
	nMag, nNeg := u256FromInt(n)
	return makeFraction(nMag, u256One, nNeg, false)
}

// FromBigInt constructs a Fraction preserving the supplied *big.Int numerator
// and denominator pair. It returns [ErrInvalid] if num == nil,
// [ErrZeroDenominator] if den == nil or den.Sign() == 0, and [ErrOutOfRange]
// if |num| or |den| exceeds [MaxComponent].
func FromBigInt(num, den *big.Int) (Fraction, error) {
	if num == nil {
		return Fraction{}, ErrInvalid
	}
	if den == nil || den.Sign() == 0 {
		return Fraction{}, ErrZeroDenominator
	}
	nMag, nNeg, okN := u256FromBigInt(num)
	if !okN {
		return Fraction{}, ErrOutOfRange
	}
	dMag, dNeg, okD := u256FromBigInt(den)
	if !okD {
		return Fraction{}, ErrOutOfRange
	}
	return makeFraction(nMag, dMag, nNeg, dNeg), nil
}

// MustFromBigInt is like [FromBigInt] but panics on error.
func MustFromBigInt(num, den *big.Int) Fraction {
	f, err := FromBigInt(num, den)
	if err != nil {
		panic(err)
	}
	return f
}

// FromBigIntReduced constructs a Fraction reduced to canonical form from num
// and den. Reduction is performed in arbitrary precision before checking
// [MaxComponent]. It returns [ErrInvalid] if num == nil, [ErrZeroDenominator]
// if den == nil or den.Sign() == 0, and [ErrOutOfRange] if the reduced pair
// exceeds [MaxComponent].
func FromBigIntReduced(num, den *big.Int) (Fraction, error) {
	if num == nil {
		return Fraction{}, ErrInvalid
	}
	if den == nil || den.Sign() == 0 {
		return Fraction{}, ErrZeroDenominator
	}
	if num.Sign() == 0 {
		return Zero(), nil
	}
	nMag, nNeg, okN := u256FromBigInt(num)
	dMag, dNeg, okD := u256FromBigInt(den)
	if okN && okD {
		return makeFraction(nMag, dMag, nNeg, dNeg).Reduce(), nil
	}
	absN := new(big.Int).Abs(num)
	absD := new(big.Int).Abs(den)
	g := new(big.Int).GCD(nil, nil, absN, absD)
	absN.Quo(absN, g)
	absD.Quo(absD, g)
	nRed, _, okN := u256FromBigInt(absN)
	if !okN {
		return Fraction{}, ErrOutOfRange
	}
	dRed, _, okD := u256FromBigInt(absD)
	if !okD {
		return Fraction{}, ErrOutOfRange
	}
	resNeg := (num.Sign() < 0) != (den.Sign() < 0)
	return makeFraction(nRed, dRed, resNeg, false), nil
}

// MustFromBigIntReduced is like [FromBigIntReduced] but panics on error.
func MustFromBigIntReduced(num, den *big.Int) Fraction {
	f, err := FromBigIntReduced(num, den)
	if err != nil {
		panic(err)
	}
	return f
}

// FromBigRat constructs a reduced Fraction from r. It returns [ErrInvalid] if
// r == nil and [ErrOutOfRange] if |r.Num()| or r.Denom() exceeds
// [MaxComponent].
func FromBigRat(r *big.Rat) (Fraction, error) {
	if r == nil {
		return Fraction{}, ErrInvalid
	}
	nMag, nNeg, okN := u256FromBigInt(r.Num())
	if !okN {
		return Fraction{}, ErrOutOfRange
	}
	dMag, _, okD := u256FromBigInt(r.Denom())
	if !okD {
		return Fraction{}, ErrOutOfRange
	}
	return makeFraction(nMag, dMag, nNeg, false), nil
}

// MustFromBigRat is like [FromBigRat] but panics on error.
func MustFromBigRat(r *big.Rat) Fraction {
	f, err := FromBigRat(r)
	if err != nil {
		panic(err)
	}
	return f
}

// Numerator returns the stored numerator as an int and true if f is valid and
// the numerator fits in int; otherwise it returns 0, false.
func (f Fraction) Numerator() (int, bool) {
	if !f.IsValid() {
		return 0, false
	}
	return u256ToInt(f.num, f.numNeg())
}

// Denominator returns the stored denominator as an int and true if f is valid
// and the denominator fits in int; otherwise it returns 0, false.
func (f Fraction) Denominator() (int, bool) {
	if !f.IsValid() {
		return 0, false
	}
	return u256ToInt(f.den, f.denNeg())
}

// Pair returns the stored numerator and denominator as int values and true if
// f is valid and both components fit in int; otherwise it returns 0, 0, false.
func (f Fraction) Pair() (num, den int, ok bool) {
	if !f.IsValid() {
		return 0, 0, false
	}
	n, okN := u256ToInt(f.num, f.numNeg())
	if !okN {
		return 0, 0, false
	}
	d, okD := u256ToInt(f.den, f.denNeg())
	if !okD {
		return 0, 0, false
	}
	return n, d, true
}

// ReducedNumerator returns the numerator of the reduced canonical form as an
// int and true if f is valid and the reduced numerator fits in int; otherwise
// it returns 0, false.
func (f Fraction) ReducedNumerator() (int, bool) {
	if !f.IsValid() {
		return 0, false
	}
	red := f.Reduce()
	return u256ToInt(red.num, red.numNeg())
}

// ReducedDenominator returns the positive denominator of the reduced canonical
// form as an int and true if f is valid and the reduced denominator fits in
// int; otherwise it returns 0, false.
func (f Fraction) ReducedDenominator() (int, bool) {
	if !f.IsValid() {
		return 0, false
	}
	red := f.Reduce()
	return u256ToInt(red.den, false)
}

// ReducedPair returns the reduced canonical numerator and denominator as int
// values and true if f is valid and both fit in int; otherwise it returns
// 0, 0, false.
func (f Fraction) ReducedPair() (num, den int, ok bool) {
	if !f.IsValid() {
		return 0, 0, false
	}
	red := f.Reduce()
	n, okN := u256ToInt(red.num, red.numNeg())
	if !okN {
		return 0, 0, false
	}
	d, okD := u256ToInt(red.den, false)
	if !okD {
		return 0, 0, false
	}
	return n, d, true
}

// Int returns the exact integer value of f and true if f is valid, represents
// an exact integer (reduced denominator == 1), and fits in int; otherwise it
// returns 0, false.
func (f Fraction) Int() (int, bool) {
	if !f.IsValid() {
		return 0, false
	}
	if f.num.isZero() {
		return 0, true
	}
	quo, rem := u256DivMod(f.num, f.den)
	if !rem.isZero() {
		return 0, false
	}
	return u256ToInt(quo, f.numNeg() != f.denNeg())
}

// BigNumerator returns a new *big.Int with the stored numerator and true if f
// is valid; otherwise nil, false.
func (f Fraction) BigNumerator() (*big.Int, bool) {
	if !f.IsValid() {
		return nil, false
	}
	return f.num.toBigInt(f.numNeg()), true
}

// BigDenominator returns a new *big.Int with the stored denominator and true if
// f is valid; otherwise nil, false.
func (f Fraction) BigDenominator() (*big.Int, bool) {
	if !f.IsValid() {
		return nil, false
	}
	return f.den.toBigInt(f.denNeg()), true
}

// BigPair returns new *big.Int instances for the stored numerator and
// denominator and true if f is valid; otherwise nil, nil, false.
func (f Fraction) BigPair() (num, den *big.Int, ok bool) {
	if !f.IsValid() {
		return nil, nil, false
	}
	return f.num.toBigInt(f.numNeg()), f.den.toBigInt(f.denNeg()), true
}

// ReducedBigNumerator returns a new *big.Int with the reduced canonical
// numerator and true if f is valid; otherwise nil, false.
func (f Fraction) ReducedBigNumerator() (*big.Int, bool) {
	if !f.IsValid() {
		return nil, false
	}
	red := f.Reduce()
	return red.num.toBigInt(red.numNeg()), true
}

// ReducedBigDenominator returns a new *big.Int with the reduced canonical
// positive denominator and true if f is valid; otherwise nil, false.
func (f Fraction) ReducedBigDenominator() (*big.Int, bool) {
	if !f.IsValid() {
		return nil, false
	}
	red := f.Reduce()
	return red.den.toBigInt(false), true
}

// ReducedBigPair returns new *big.Int instances for the reduced canonical
// numerator and denominator and true if f is valid; otherwise nil, nil, false.
func (f Fraction) ReducedBigPair() (num, den *big.Int, ok bool) {
	if !f.IsValid() {
		return nil, nil, false
	}
	red := f.Reduce()
	return red.num.toBigInt(red.numNeg()), red.den.toBigInt(false), true
}

// BigInt returns a new *big.Int with the exact integer value of f and true if
// f is valid and represents an exact integer; otherwise nil, false.
func (f Fraction) BigInt() (*big.Int, bool) {
	if !f.IsValid() {
		return nil, false
	}
	if f.num.isZero() {
		return new(big.Int), true
	}
	quo, rem := u256DivMod(f.num, f.den)
	if !rem.isZero() {
		return nil, false
	}
	return quo.toBigInt(f.numNeg() != f.denNeg()), true
}

// BigRat returns a new *big.Rat equal to f and true if f is valid; otherwise
// nil, false.
func (f Fraction) BigRat() (*big.Rat, bool) {
	if !f.IsValid() {
		return nil, false
	}
	red := f.Reduce()
	r := new(big.Rat).SetFrac(red.num.toBigInt(red.numNeg()), red.den.toBigInt(false))
	return r, true
}

// Reduce returns a new Fraction in canonical reduced form (positive
// denominator, coprime numerator and denominator, 0/1 for zero). For an
// invalid Fraction{}, it returns Fraction{}.
func (f Fraction) Reduce() Fraction {
	if !f.IsValid() {
		return Fraction{}
	}
	if f.num.isZero() {
		return Zero()
	}
	resNeg := f.numNeg() != f.denNeg()
	g := u256GCD(f.num, f.den)
	if g.isOne() {
		return makeFraction(f.num, f.den, resNeg, false)
	}
	nRed, _ := u256DivMod(f.num, g)
	dRed, _ := u256DivMod(f.den, g)
	return makeFraction(nRed, dRed, resNeg, false)
}

func (f Fraction) withTargetDenU256(targetDen u256) (Fraction, error) {
	red := f.Reduce()
	k, rem := u256DivMod(targetDen, red.den)
	if !rem.isZero() {
		return Fraction{}, ErrNotRepresentable
	}
	prod := u256Mul(red.num, k)
	if !prod.fitsU256Max() {
		return Fraction{}, ErrOutOfRange
	}
	return makeFraction(prod.low256(), targetDen, red.numNeg(), false), nil
}

// WithDenominator returns a new Fraction numerically equal to f with stored
// denominator den. It returns [ErrInvalid] if f is invalid,
// [ErrZeroDenominator] if den <= 0, [ErrNotRepresentable] if f cannot be
// expressed with denominator den without remainder, and [ErrOutOfRange] if the
// resulting numerator exceeds [MaxComponent].
func (f Fraction) WithDenominator(den int) (Fraction, error) {
	if !f.IsValid() {
		return Fraction{}, ErrInvalid
	}
	if den <= 0 {
		return Fraction{}, ErrZeroDenominator
	}
	targetDen, _ := u256FromInt(den)
	return f.withTargetDenU256(targetDen)
}

// MustWithDenominator is like [Fraction.WithDenominator] but panics on error.
func (f Fraction) MustWithDenominator(den int) Fraction {
	res, err := f.WithDenominator(den)
	if err != nil {
		panic(err)
	}
	return res
}

// WithBigDenominator returns a new Fraction numerically equal to f with stored
// denominator den. It returns [ErrInvalid] if f is invalid,
// [ErrZeroDenominator] if den == nil or den.Sign() <= 0, [ErrNotRepresentable]
// if f cannot be expressed with denominator den without remainder, and
// [ErrOutOfRange] if den or the resulting numerator exceeds [MaxComponent].
func (f Fraction) WithBigDenominator(den *big.Int) (Fraction, error) {
	if !f.IsValid() {
		return Fraction{}, ErrInvalid
	}
	if den == nil || den.Sign() <= 0 {
		return Fraction{}, ErrZeroDenominator
	}
	targetDen, _, ok := u256FromBigInt(den)
	if ok {
		return f.withTargetDenU256(targetDen)
	}
	// den > MaxComponent: check divisibility (Step 3) before range (Step 4).
	red := f.Reduce()
	rem := new(big.Int).Mod(den, red.den.toBigInt(false))
	if rem.Sign() != 0 {
		return Fraction{}, ErrNotRepresentable
	}
	return Fraction{}, ErrOutOfRange
}

// MustWithBigDenominator is like [Fraction.WithBigDenominator] but panics on
// error.
func (f Fraction) MustWithBigDenominator(den *big.Int) Fraction {
	res, err := f.WithBigDenominator(den)
	if err != nil {
		panic(err)
	}
	return res
}

// Equal reports whether f and other are both valid and represent the same
// rational number (n1 * d2 == n2 * d1).
func (f Fraction) Equal(other Fraction) bool {
	if !f.IsValid() || !other.IsValid() {
		return false
	}
	return f.Compare(other) == 0
}

// EqualExact reports whether f and other are both valid and have identical
// stored numerators and denominators (n1 == n2 && d1 == d2).
func (f Fraction) EqualExact(other Fraction) bool {
	if !f.IsValid() || !other.IsValid() {
		return false
	}
	return f == other
}

// Compare compares f and other by numeric value, returning -1 if f < other,
// 0 if f == other numerically, and +1 if f > other. Invalid Fraction{} is
// ordered before all valid fractions and equal to itself.
func (f Fraction) Compare(other Fraction) int {
	fValid := f.IsValid()
	oValid := other.IsValid()
	if !fValid || !oValid {
		if !fValid && !oValid {
			return 0
		}
		if !fValid {
			return -1
		}
		return 1
	}
	s1 := f.Sign()
	s2 := other.Sign()
	if s1 != s2 {
		if s1 < s2 {
			return -1
		}
		return 1
	}
	if s1 == 0 {
		return 0
	}
	lhs := u256Mul(f.num, other.den)
	rhs := u256Mul(other.num, f.den)
	c := u512Cmp(lhs, rhs)
	if s1 < 0 {
		return -c
	}
	return c
}

// CompareExact defines a total deterministic order on Fraction values: first by
// numeric value ([Fraction.Compare]), then by signed stored denominator, and
// finally by signed stored numerator. CompareExact(other) == 0 if and only if
// f == other.
func (f Fraction) CompareExact(other Fraction) int {
	if c := f.Compare(other); c != 0 {
		return c
	}
	if !f.IsValid() && !other.IsValid() {
		return 0
	}
	if c := cmpSignedU256(f.den, f.denNeg(), other.den, other.denNeg()); c != 0 {
		return c
	}
	return cmpSignedU256(f.num, f.numNeg(), other.num, other.numNeg())
}

// Less reports whether f and other are both valid and f < other numerically.
func (f Fraction) Less(other Fraction) bool {
	if !f.IsValid() || !other.IsValid() {
		return false
	}
	return f.Compare(other) < 0
}

// LessOrEqual reports whether f and other are both valid and f <= other
// numerically.
func (f Fraction) LessOrEqual(other Fraction) bool {
	if !f.IsValid() || !other.IsValid() {
		return false
	}
	return f.Compare(other) <= 0
}

// Greater reports whether f and other are both valid and f > other numerically.
func (f Fraction) Greater(other Fraction) bool {
	if !f.IsValid() || !other.IsValid() {
		return false
	}
	return f.Compare(other) > 0
}

// GreaterOrEqual reports whether f and other are both valid and f >= other
// numerically.
func (f Fraction) GreaterOrEqual(other Fraction) bool {
	if !f.IsValid() || !other.IsValid() {
		return false
	}
	return f.Compare(other) >= 0
}

// IsValid reports whether f is a valid fraction (stored denominator != 0).
func (f Fraction) IsValid() bool {
	return !f.den.isZero()
}

// IsZero reports whether f is valid and numerically equal to 0.
func (f Fraction) IsZero() bool {
	return !f.den.isZero() && f.num.isZero()
}

// IsOne reports whether f is valid and numerically equal to 1 (n == d).
func (f Fraction) IsOne() bool {
	return !f.den.isZero() && f.num == f.den && (f.numNeg() == f.denNeg())
}

// IsInteger reports whether f is valid and its numeric value is an exact
// integer (|n| mod |d| == 0).
func (f Fraction) IsInteger() bool {
	if !f.IsValid() {
		return false
	}
	if f.num.isZero() || f.den.isOne() {
		return true
	}
	if u256Cmp(f.num, f.den) < 0 {
		return false
	}
	_, rem := u256DivMod(f.num, f.den)
	return rem.isZero()
}

// IsProper reports whether f is valid and |n| < |d|.
func (f Fraction) IsProper() bool {
	if !f.IsValid() {
		return false
	}
	return u256Cmp(f.num, f.den) < 0
}

// IsReduced reports whether f is valid and its stored pair is already in
// canonical reduced form (d > 0 and gcd(|n|, d) == 1, with 0/1 for zero).
func (f Fraction) IsReduced() bool {
	if !f.IsValid() || f.denNeg() {
		return false
	}
	if f.num.isZero() {
		return f.den.isOne()
	}
	if f.den.isOne() {
		return true
	}
	g := u256GCD(f.num, f.den)
	return g.isOne()
}

// Sign returns -1 if f < 0, 0 if f == 0 or f is invalid, and +1 if f > 0.
func (f Fraction) Sign() int {
	if !f.IsValid() || f.num.isZero() {
		return 0
	}
	if f.numNeg() == f.denNeg() {
		return 1
	}
	return -1
}
