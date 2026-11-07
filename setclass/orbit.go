package setclass

import (
	"fmt"
	"slices"

	"github.com/go-muse/musex/pcset"
	"github.com/go-muse/musex/pitchclass"
)

// Contains reports whether the pitch-class set s belongs to the equivalence class sc.
// It returns false without error if sc or s is invalid, or if they have different moduli or cardinalities.
func (sc SetClass) Contains(s pcset.Set) bool {
	if !sc.IsValid() || !s.IsValid() || sc.prime.Modulus() != s.Modulus() || sc.prime.Cardinality() != s.Cardinality() {
		return false
	}
	return sc.prime.Equal(computePrimeForm(s, sc.kind, Rahn))
}

func validateSetPairAndKind(a, b pcset.Set, kind Kind, checkOrbitBound bool) error {
	if !a.IsValid() || !b.IsValid() {
		return fmt.Errorf("%w: invalid pcset.Set", ErrInvalidValue)
	}
	if !kind.IsValid() {
		return fmt.Errorf("%w: %v", ErrInvalidKind, kind)
	}
	n := a.Modulus()
	if n != b.Modulus() {
		return fmt.Errorf("%w: %d != %d", ErrModulusMismatch, n, b.Modulus())
	}
	if (kind == TTO && n > MaxTTOModulus) || (checkOrbitBound && n > MaxOrbitModulus) {
		return fmt.Errorf("%w: modulus %d exceeds limit", ErrModulusTooLarge, n)
	}
	return nil
}

func validateClassPair(a, b SetClass) error {
	if !a.IsValid() || !b.IsValid() {
		return fmt.Errorf("%w: invalid SetClass", ErrInvalidValue)
	}
	if a.prime.Modulus() != b.prime.Modulus() {
		return fmt.Errorf("%w: %d != %d", ErrModulusMismatch, a.prime.Modulus(), b.prime.Modulus())
	}
	if a.kind != b.kind {
		return fmt.Errorf("%w: %v != %v", ErrKindMismatch, a.kind, b.kind)
	}
	return nil
}

func (sc SetClass) validateOrbitReceiver() (int, error) {
	if !sc.IsValid() {
		return 0, fmt.Errorf("%w: invalid SetClass", ErrInvalidValue)
	}
	n := sc.prime.Modulus()
	if n > MaxOrbitModulus {
		return 0, fmt.Errorf("%w: modulus %d > %d", ErrModulusTooLarge, n, MaxOrbitModulus)
	}
	return n, nil
}

func filterGroupOperators(source pcset.Set, kind Kind, pred func(img pcset.Set) bool) []pitchclass.Operator {
	n := source.Modulus()
	out := []pitchclass.Operator{}
	switch kind {
	case Tn:
		for t := range n {
			if pred(source.Transpose(t)) {
				out = append(out, pitchclass.MustTransposition(t, n))
			}
		}
	case TnIn:
		for t := range n {
			if pred(source.Transpose(t)) {
				out = append(out, pitchclass.MustTransposition(t, n))
			}
		}
		if n > 2 {
			for t := range n {
				if pred(source.Invert(t)) {
					out = append(out, pitchclass.MustInversion(t, n))
				}
			}
		}
	default: // TTO
		for m := range pitchclass.MustInvertibleMultipliers(n) {
			ms := source.Multiply(m)
			for t := range n {
				if pred(ms.Transpose(t)) {
					out = append(out, pitchclass.MustNewOperator(m, t, n))
				}
			}
		}
	}
	return out
}

// AreEquivalent reports whether two pitch-class sets a and b belong to the same SetClass of kind.
func AreEquivalent(a, b pcset.Set, kind Kind) (bool, error) {
	if err := validateSetPairAndKind(a, b, kind, false); err != nil {
		return false, err
	}
	return MustFromSet(a, kind).Contains(b), nil
}

// MustAreEquivalent is like [AreEquivalent] but panics on error.
func MustAreEquivalent(a, b pcset.Set, kind Kind) bool {
	eq, err := AreEquivalent(a, b, kind)
	if err != nil {
		panic(err)
	}
	return eq
}

// AreEquivalentTn reports whether a and b are transpositionally equivalent (in Tn).
func AreEquivalentTn(a, b pcset.Set) (bool, error) {
	return AreEquivalent(a, b, Tn)
}

// MustAreEquivalentTn is like [AreEquivalentTn] but panics on error.
func MustAreEquivalentTn(a, b pcset.Set) bool {
	eq, err := AreEquivalentTn(a, b)
	if err != nil {
		panic(err)
	}
	return eq
}

// AreEquivalentTnIn reports whether a and b are transpositionally or inversionally equivalent (in TnIn).
func AreEquivalentTnIn(a, b pcset.Set) (bool, error) {
	return AreEquivalent(a, b, TnIn)
}

// MustAreEquivalentTnIn is like [AreEquivalentTnIn] but panics on error.
func MustAreEquivalentTnIn(a, b pcset.Set) bool {
	eq, err := AreEquivalentTnIn(a, b)
	if err != nil {
		panic(err)
	}
	return eq
}

// AreEquivalentTTO reports whether a and b are equivalent under the twelve-tone operator group (in TTO).
func AreEquivalentTTO(a, b pcset.Set) (bool, error) {
	return AreEquivalent(a, b, TTO)
}

// MustAreEquivalentTTO is like [AreEquivalentTTO] but panics on error.
func MustAreEquivalentTTO(a, b pcset.Set) bool {
	eq, err := AreEquivalentTTO(a, b)
	if err != nil {
		panic(err)
	}
	return eq
}

func (sc SetClass) validateTargetWith(target pcset.Set, alg Algorithm) (pcset.Set, error) {
	if !sc.IsValid() || !target.IsValid() {
		return pcset.Set{}, fmt.Errorf("%w: invalid SetClass or pcset.Set", ErrInvalidValue)
	}
	if !alg.IsValid() {
		return pcset.Set{}, fmt.Errorf("%w: %v", ErrInvalidAlgorithm, alg)
	}
	n := sc.prime.Modulus()
	if n != target.Modulus() {
		return pcset.Set{}, fmt.Errorf("%w: %d != %d", ErrModulusMismatch, n, target.Modulus())
	}
	if n > MaxOrbitModulus {
		return pcset.Set{}, fmt.Errorf("%w: modulus %d > %d", ErrModulusTooLarge, n, MaxOrbitModulus)
	}
	return sc.MustPrimeFormWith(alg), nil
}

// TranspositionIndicesTo returns the set of transposition indices {n in Z_N : T_n(P) = target}
// mapping the Rahn prime form P of sc to target (1 <= N <= 1024).
// If target does not belong to sc, it returns the empty set of modulus N without error.
func (sc SetClass) TranspositionIndicesTo(target pcset.Set) (pcset.Set, error) {
	return sc.TranspositionIndicesToWith(target, Rahn)
}

// MustTranspositionIndicesTo is like [SetClass.TranspositionIndicesTo] but panics on error.
func (sc SetClass) MustTranspositionIndicesTo(target pcset.Set) pcset.Set {
	s, err := sc.TranspositionIndicesTo(target)
	if err != nil {
		panic(err)
	}
	return s
}

// TranspositionIndicesToWith returns the set of transposition indices {n in Z_N : T_n(P) = target}
// mapping the prime form P of sc under alg to target (1 <= N <= 1024).
func (sc SetClass) TranspositionIndicesToWith(target pcset.Set, alg Algorithm) (pcset.Set, error) {
	pf, err := sc.validateTargetWith(target, alg)
	if err != nil {
		return pcset.Set{}, err
	}
	if !sc.Contains(target) {
		return pcset.MustEmpty(pf.Modulus()), nil
	}
	return pf.MustTranspositionIndicesTo(target), nil
}

// MustTranspositionIndicesToWith is like [SetClass.TranspositionIndicesToWith] but panics on error.
func (sc SetClass) MustTranspositionIndicesToWith(target pcset.Set, alg Algorithm) pcset.Set {
	s, err := sc.TranspositionIndicesToWith(target, alg)
	if err != nil {
		panic(err)
	}
	return s
}

// InversionIndicesTo returns the set of inversion indices {n in Z_N : I_n(P) = target}
// mapping the Rahn prime form P of sc to target (1 <= N <= 1024).
// If target does not belong to sc, it returns the empty set of modulus N without error.
func (sc SetClass) InversionIndicesTo(target pcset.Set) (pcset.Set, error) {
	return sc.InversionIndicesToWith(target, Rahn)
}

// MustInversionIndicesTo is like [SetClass.InversionIndicesTo] but panics on error.
func (sc SetClass) MustInversionIndicesTo(target pcset.Set) pcset.Set {
	s, err := sc.InversionIndicesTo(target)
	if err != nil {
		panic(err)
	}
	return s
}

// InversionIndicesToWith returns the set of inversion indices {n in Z_N : I_n(P) = target}
// mapping the prime form P of sc under alg to target (1 <= N <= 1024).
func (sc SetClass) InversionIndicesToWith(target pcset.Set, alg Algorithm) (pcset.Set, error) {
	pf, err := sc.validateTargetWith(target, alg)
	if err != nil {
		return pcset.Set{}, err
	}
	if !sc.Contains(target) {
		return pcset.MustEmpty(pf.Modulus()), nil
	}
	return pf.MustInversionIndicesTo(target), nil
}

// MustInversionIndicesToWith is like [SetClass.InversionIndicesToWith] but panics on error.
func (sc SetClass) MustInversionIndicesToWith(target pcset.Set, alg Algorithm) pcset.Set {
	s, err := sc.InversionIndicesToWith(target, alg)
	if err != nil {
		panic(err)
	}
	return s
}

// FindSetOperators returns all operators f in the equivalence group kind such that f(source) = target,
// sorted by [pitchclass.Operator.Compare].
func FindSetOperators(source, target pcset.Set, kind Kind) ([]pitchclass.Operator, error) {
	if err := validateSetPairAndKind(source, target, kind, true); err != nil {
		return nil, err
	}
	return filterGroupOperators(source, kind, target.Equal), nil
}

// MustFindSetOperators is like [FindSetOperators] but panics on error.
func MustFindSetOperators(source, target pcset.Set, kind Kind) []pitchclass.Operator {
	ops, err := FindSetOperators(source, target, kind)
	if err != nil {
		panic(err)
	}
	return ops
}

// OperatorsTo returns all operators f in sc.Kind() mapping the Rahn prime form P of sc to target (f(P) = target),
// sorted by [pitchclass.Operator.Compare].
func (sc SetClass) OperatorsTo(target pcset.Set) ([]pitchclass.Operator, error) {
	return sc.OperatorsToWith(target, Rahn)
}

// MustOperatorsTo is like [SetClass.OperatorsTo] but panics on error.
func (sc SetClass) MustOperatorsTo(target pcset.Set) []pitchclass.Operator {
	ops, err := sc.OperatorsTo(target)
	if err != nil {
		panic(err)
	}
	return ops
}

// OperatorsToWith returns all operators f in sc.Kind() mapping the prime form P of sc under alg to target (f(P) = target),
// sorted by [pitchclass.Operator.Compare].
func (sc SetClass) OperatorsToWith(target pcset.Set, alg Algorithm) ([]pitchclass.Operator, error) {
	pf, err := sc.validateTargetWith(target, alg)
	if err != nil {
		return nil, err
	}
	return FindSetOperators(pf, target, sc.kind)
}

// MustOperatorsToWith is like [SetClass.OperatorsToWith] but panics on error.
func (sc SetClass) MustOperatorsToWith(target pcset.Set, alg Algorithm) []pitchclass.Operator {
	ops, err := sc.OperatorsToWith(target, alg)
	if err != nil {
		panic(err)
	}
	return ops
}

// OperatorsFrom returns all operators f in sc.Kind() mapping source to the Rahn prime form P of sc (f(source) = P),
// sorted by [pitchclass.Operator.Compare].
func (sc SetClass) OperatorsFrom(source pcset.Set) ([]pitchclass.Operator, error) {
	return sc.OperatorsFromWith(source, Rahn)
}

// MustOperatorsFrom is like [SetClass.OperatorsFrom] but panics on error.
func (sc SetClass) MustOperatorsFrom(source pcset.Set) []pitchclass.Operator {
	ops, err := sc.OperatorsFrom(source)
	if err != nil {
		panic(err)
	}
	return ops
}

// OperatorsFromWith returns all operators f in sc.Kind() mapping source to the prime form P of sc under alg (f(source) = P),
// sorted by [pitchclass.Operator.Compare].
func (sc SetClass) OperatorsFromWith(source pcset.Set, alg Algorithm) ([]pitchclass.Operator, error) {
	pf, err := sc.validateTargetWith(source, alg)
	if err != nil {
		return nil, err
	}
	return FindSetOperators(source, pf, sc.kind)
}

// MustOperatorsFromWith is like [SetClass.OperatorsFromWith] but panics on error.
func (sc SetClass) MustOperatorsFromWith(source pcset.Set, alg Algorithm) []pitchclass.Operator {
	ops, err := sc.OperatorsFromWith(source, alg)
	if err != nil {
		panic(err)
	}
	return ops
}

// Members returns a fresh slice of all distinct member sets in the orbit of sc,
// sorted by [pcset.Set.Compare] (1 <= N <= 1024 for Tn/TnIn; 1 <= N <= 256 for TTO).
func (sc SetClass) Members() ([]pcset.Set, error) {
	n, err := sc.validateOrbitReceiver()
	if err != nil {
		return nil, err
	}
	period := n / sc.TnSymmetry()
	tnClasses := sc.TnClasses()
	out := make([]pcset.Set, 0, len(tnClasses)*period)
	for _, tc := range tnClasses {
		for t := range period {
			out = append(out, tc.prime.Transpose(t))
		}
	}
	slices.SortFunc(out, pcset.Set.Compare)
	return out, nil
}

// MustMembers is like [SetClass.Members] but panics on error.
func (sc SetClass) MustMembers() []pcset.Set {
	m, err := sc.Members()
	if err != nil {
		panic(err)
	}
	return m
}

// TnSymmetry returns the transpositional symmetry degree DS_T in [1, N] of sc, or 0 if !sc.IsValid().
func (sc SetClass) TnSymmetry() int {
	if !sc.IsValid() {
		return 0
	}
	return sc.prime.MustDegreeOfTranspositionSymmetry()
}

// InSymmetry returns the inversional symmetry degree DS_I in [0, N] of sc, or 0 if !sc.IsValid().
func (sc SetClass) InSymmetry() int {
	if !sc.IsValid() {
		return 0
	}
	return sc.prime.MustDegreeOfInversionSymmetry()
}

// SymmetryDegreeTI returns the stabilizer size DS_{T/I} of a member set of sc in the dihedral group Tn/In:
// DS_T for N <= 2 (where inversions coincide with transpositions), and DS_T + DS_I for N >= 3.
// For an invalid SetClass, it returns 0.
func (sc SetClass) SymmetryDegreeTI() int {
	if !sc.IsValid() {
		return 0
	}
	if sc.prime.Modulus() <= 2 {
		return sc.prime.MustDegreeOfTranspositionSymmetry()
	}
	return sc.prime.MustDegreeOfTranspositionSymmetry() + sc.prime.MustDegreeOfInversionSymmetry()
}

// SymmetryDegreeTTO returns the stabilizer size DS_{TTO} of a member set of sc in the full TTO group (1 <= N <= 256).
func (sc SetClass) SymmetryDegreeTTO() (int, error) {
	if !sc.IsValid() {
		return 0, fmt.Errorf("%w: invalid SetClass", ErrInvalidValue)
	}
	n := sc.prime.Modulus()
	if n > MaxTTOModulus {
		return 0, fmt.Errorf("%w: modulus %d > %d", ErrModulusTooLarge, n, MaxTTOModulus)
	}
	tnPrime := computePrimeForm(sc.prime, Tn, Rahn)
	countM := 0
	for m := range pitchclass.MustInvertibleMultipliers(n) {
		if computePrimeForm(sc.prime.Multiply(m), Tn, Rahn).Equal(tnPrime) {
			countM++
		}
	}
	return sc.prime.MustDegreeOfTranspositionSymmetry() * countM, nil
}

// MustSymmetryDegreeTTO is like [SetClass.SymmetryDegreeTTO] but panics on error.
func (sc SetClass) MustSymmetryDegreeTTO() int {
	d, err := sc.SymmetryDegreeTTO()
	if err != nil {
		panic(err)
	}
	return d
}

// OrbitSize returns the exact number of distinct pitch-class sets in the equivalence class sc,
// or 0 if !sc.IsValid().
func (sc SetClass) OrbitSize() int {
	if !sc.IsValid() {
		return 0
	}
	n := sc.prime.Modulus()
	switch sc.kind {
	case Tn:
		return n / sc.TnSymmetry()
	case TnIn:
		if n <= 2 {
			return n / sc.TnSymmetry()
		}
		return (2 * n) / sc.SymmetryDegreeTI()
	default: // TTO
		phi := pitchclass.MustTotient(n)
		return (n * phi) / sc.MustSymmetryDegreeTTO()
	}
}

// IsTranspositionallySymmetric reports whether sc has non-trivial transpositional symmetry (DS_T > 1).
func (sc SetClass) IsTranspositionallySymmetric() bool {
	return sc.IsValid() && sc.TnSymmetry() > 1
}

// IsInversionallySymmetric reports whether sc is achiral (inversionally symmetric: DS_I > 0).
func (sc SetClass) IsInversionallySymmetric() bool {
	return sc.IsValid() && sc.InSymmetry() > 0
}

// IsChiral reports whether sc is chiral (enantiomorphic: N >= 3 and DS_I == 0).
func (sc SetClass) IsChiral() bool {
	return sc.IsValid() && sc.prime.Modulus() >= 3 && sc.InSymmetry() == 0
}

// ToTnIn returns the TnIn SetClass corresponding to sc:
//   - for Kind == Tn: the enclosing TnIn SetClass;
//   - for Kind == TnIn: sc itself;
//   - for Kind == TTO: the TnIn SetClass of sc's canonical Rahn prime form.
//
// For an invalid SetClass, ToTnIn returns SetClass{}.
func (sc SetClass) ToTnIn() SetClass {
	if !sc.IsValid() {
		return SetClass{}
	}
	if sc.kind == TnIn {
		return sc
	}
	return SetClass{
		kind:  TnIn,
		prime: computePrimeForm(sc.prime, TnIn, Rahn),
	}
}

// ToTTO returns the enclosing TTO SetClass of sc (1 <= N <= 256).
func (sc SetClass) ToTTO() (SetClass, error) {
	return FromSet(sc.prime, TTO)
}

// MustToTTO is like [SetClass.ToTTO] but panics on error.
func (sc SetClass) MustToTTO() SetClass {
	tto, err := sc.ToTTO()
	if err != nil {
		panic(err)
	}
	return tto
}

// TnClasses decomposes sc into its constituent transpositional (Tn) SetClasses:
//   - for Kind == Tn: a 1-element slice [sc];
//   - for Kind == TnIn: 1 element if achiral, or 2 elements [[P]_{Tn}, [I_0(P)]_{Tn}] in Rahn packing order if chiral;
//   - for Kind == TTO: all constituent Tn SetClasses sorted by [SetClass.Compare].
//
// For an invalid SetClass, TnClasses returns nil.
func (sc SetClass) TnClasses() []SetClass {
	if !sc.IsValid() {
		return nil
	}
	switch sc.kind {
	case Tn:
		return []SetClass{sc}
	case TnIn:
		c1 := SetClass{kind: Tn, prime: sc.prime}
		if !sc.IsChiral() {
			return []SetClass{c1}
		}
		c2 := SetClass{kind: Tn, prime: computePrimeForm(sc.prime.Invert(0), Tn, Rahn)}
		return []SetClass{c1, c2}
	default: // TTO
		n := sc.prime.Modulus()
		var out []SetClass
		for m := range pitchclass.MustInvertibleMultipliers(n) {
			out = append(out, SetClass{
				kind:  Tn,
				prime: computePrimeForm(sc.prime.Multiply(m), Tn, Rahn),
			})
		}
		slices.SortFunc(out, SetClass.Compare)
		return slices.Compact(out)
	}
}

// TnInClasses decomposes sc into its constituent TnIn SetClasses sorted by [SetClass.Compare] (1 <= N <= 256).
func (sc SetClass) TnInClasses() ([]SetClass, error) {
	if _, err := sc.ToTTO(); err != nil {
		return nil, err
	}
	if sc.kind != TTO {
		return []SetClass{sc.ToTnIn()}, nil
	}
	n := sc.prime.Modulus()
	var out []SetClass
	for m := range pitchclass.MustInvertibleMultipliers(n) {
		out = append(out, SetClass{
			kind:  TnIn,
			prime: computePrimeForm(sc.prime.Multiply(m), TnIn, Rahn),
		})
	}
	slices.SortFunc(out, SetClass.Compare)
	return slices.Compact(out), nil
}

// MustTnInClasses is like [SetClass.TnInClasses] but panics on error.
func (sc SetClass) MustTnInClasses() []SetClass {
	list, err := sc.TnInClasses()
	if err != nil {
		panic(err)
	}
	return list
}

// IntervalVector returns the floor(N/2)-element interval-class vector of sc's Rahn prime form,
// or the zero [pcset.IntervalVector] if !sc.IsValid().
func (sc SetClass) IntervalVector() pcset.IntervalVector {
	if !sc.IsValid() {
		return pcset.IntervalVector{}
	}
	return sc.prime.IntervalVector()
}

// IntervalVectorWith returns the floor(N/2)-element interval-class vector of sc's prime form under alg.
func (sc SetClass) IntervalVectorWith(alg Algorithm) (pcset.IntervalVector, error) {
	pf, err := sc.PrimeFormWith(alg)
	if err != nil {
		return pcset.IntervalVector{}, err
	}
	return pf.IntervalVector(), nil
}

// MustIntervalVectorWith is like [SetClass.IntervalVectorWith] but panics on error.
func (sc SetClass) MustIntervalVectorWith(alg Algorithm) pcset.IntervalVector {
	iv, err := sc.IntervalVectorWith(alg)
	if err != nil {
		panic(err)
	}
	return iv
}

// InvarianceVector returns Robert Morris's 8-element invariance vector for sc at N = 12.
// For N != 12, it returns [ErrNotModulus12].
func (sc SetClass) InvarianceVector() (pcset.InvarianceVector, error) {
	if !sc.IsValid() {
		return pcset.InvarianceVector{}, fmt.Errorf("%w: invalid SetClass", ErrInvalidValue)
	}
	if sc.prime.Modulus() != 12 {
		return pcset.InvarianceVector{}, fmt.Errorf("%w: modulus %d != 12", ErrNotModulus12, sc.prime.Modulus())
	}
	return sc.prime.MustInvarianceVector(), nil
}

// MustInvarianceVector is like [SetClass.InvarianceVector] but panics on error.
func (sc SetClass) MustInvarianceVector() pcset.InvarianceVector {
	iv, err := sc.InvarianceVector()
	if err != nil {
		panic(err)
	}
	return iv
}

// Complement returns the abstract complement set class [Z_N \ P]_G of the same Kind and Modulus as sc,
// or SetClass{} if !sc.IsValid().
func (sc SetClass) Complement() SetClass {
	if !sc.IsValid() {
		return SetClass{}
	}
	return SetClass{
		kind:  sc.kind,
		prime: computePrimeForm(sc.prime.Complement(), sc.kind, Rahn),
	}
}

// IsSelfComplementary reports whether sc is equal to its abstract complement in sc.Kind().
func (sc SetClass) IsSelfComplementary() bool {
	if !sc.IsValid() {
		return false
	}
	if sc.prime.Cardinality()*2 != sc.prime.Modulus() {
		return false
	}
	return sc == sc.Complement()
}

// Inversion returns the inversional image [I_0(P)]_G of sc in the same Kind and Modulus,
// or SetClass{} if !sc.IsValid(). For Kind == TnIn and Kind == TTO, Inversion always returns sc.
func (sc SetClass) Inversion() SetClass {
	if !sc.IsValid() {
		return SetClass{}
	}
	if sc.kind != Tn {
		return sc
	}
	return SetClass{
		kind:  Tn,
		prime: computePrimeForm(sc.prime.Invert(0), Tn, Rahn),
	}
}

// Multiply returns the multiplicative image [M_m(P)]_G of sc in the same Kind and Modulus
// for an invertible multiplier m (gcd(m mod N, N) == 1).
// If m is not invertible modulo N, Multiply returns [ErrNonInvertibleMultiplier].
func (sc SetClass) Multiply(m int) (SetClass, error) {
	if !sc.IsValid() {
		return SetClass{}, fmt.Errorf("%w: invalid SetClass", ErrInvalidValue)
	}
	n := sc.prime.Modulus()
	if !pitchclass.MustInvertible(m, n) {
		return SetClass{}, fmt.Errorf("%w: multiplier %d modulo %d", ErrNonInvertibleMultiplier, m, n)
	}
	if sc.kind == TTO {
		return sc, nil
	}
	return SetClass{
		kind:  sc.kind,
		prime: computePrimeForm(sc.prime.Multiply(m), sc.kind, Rahn),
	}, nil
}

// MustMultiply is like [SetClass.Multiply] but panics on error.
func (sc SetClass) MustMultiply(m int) SetClass {
	res, err := sc.Multiply(m)
	if err != nil {
		panic(err)
	}
	return res
}

// M5Partner returns the M_5 multiplicative image [M_5(P)]_G of sc at N = 12.
// For N != 12, M5Partner returns [ErrNotModulus12].
func (sc SetClass) M5Partner() (SetClass, error) {
	if !sc.IsValid() {
		return SetClass{}, fmt.Errorf("%w: invalid SetClass", ErrInvalidValue)
	}
	if sc.prime.Modulus() != 12 {
		return SetClass{}, fmt.Errorf("%w: modulus %d != 12", ErrNotModulus12, sc.prime.Modulus())
	}
	return sc.Multiply(5)
}

// MustM5Partner is like [SetClass.M5Partner] but panics on error.
func (sc SetClass) MustM5Partner() SetClass {
	res, err := sc.M5Partner()
	if err != nil {
		panic(err)
	}
	return res
}

// IsM5Invariant reports whether sc equals its M_5 image in sc.Kind() at N = 12.
// For N != 12, IsM5Invariant returns [ErrNotModulus12].
func (sc SetClass) IsM5Invariant() (bool, error) {
	p, err := sc.M5Partner()
	if err != nil {
		return false, err
	}
	return sc == p, nil
}

// MustIsM5Invariant is like [SetClass.IsM5Invariant] but panics on error.
func (sc SetClass) MustIsM5Invariant() bool {
	inv, err := sc.IsM5Invariant()
	if err != nil {
		panic(err)
	}
	return inv
}

// IsZRelated reports whether sc and other are Z-related set classes of the same Modulus and Kind:
// they have the same cardinality, identical interval-class vectors, are distinct (sc != other),
// and do not belong to the same TnIn class (sc.ToTnIn() != other.ToTnIn()).
func (sc SetClass) IsZRelated(other SetClass) (bool, error) {
	if err := validateClassPair(sc, other); err != nil {
		return false, err
	}
	if sc.prime.Cardinality() != other.prime.Cardinality() || sc == other || sc.ToTnIn() == other.ToTnIn() {
		return false, nil
	}
	return sc.prime.IntervalVector().Equal(other.prime.IntervalVector()), nil
}

// MustIsZRelated is like [SetClass.IsZRelated] but panics on error.
func (sc SetClass) MustIsZRelated(other SetClass) bool {
	z, err := sc.IsZRelated(other)
	if err != nil {
		panic(err)
	}
	return z
}

// AreZRelatedSets reports whether two pitch-class sets a and b of the same modulus N
// have identical interval-class vectors while belonging to distinct TnIn set classes.
func AreZRelatedSets(a, b pcset.Set) (bool, error) {
	if !a.IsValid() || !b.IsValid() {
		return false, fmt.Errorf("%w: invalid pcset.Set", ErrInvalidValue)
	}
	if a.Modulus() != b.Modulus() {
		return false, fmt.Errorf("%w: %d != %d", ErrModulusMismatch, a.Modulus(), b.Modulus())
	}
	if a.Cardinality() != b.Cardinality() || !a.IntervalVector().Equal(b.IntervalVector()) {
		return false, nil
	}
	return !computePrimeForm(a, TnIn, Rahn).Equal(computePrimeForm(b, TnIn, Rahn)), nil
}

// MustAreZRelatedSets is like [AreZRelatedSets] but panics on error.
func MustAreZRelatedSets(a, b pcset.Set) bool {
	z, err := AreZRelatedSets(a, b)
	if err != nil {
		panic(err)
	}
	return z
}

// ZPartners returns all Z-related set classes of the same Modulus and Kind as sc,
// sorted by [SetClass.Compare].
// At N = 12, ZPartners runs in O(1) via [github.com/go-muse/musex/setclass/catalog].
// At N != 12, it requires 1 <= N <= 64 and C(N, k) <= [MaxEnumerationSets].
func (sc SetClass) ZPartners() ([]SetClass, error) {
	if !sc.IsValid() {
		return nil, fmt.Errorf("%w: invalid SetClass", ErrInvalidValue)
	}
	n := sc.prime.Modulus()
	k := sc.prime.Cardinality()
	if n == 12 {
		fc := sc.MustForteClass()
		zp, ok := fc.ZPartner()
		if !ok {
			return []SetClass{}, nil
		}
		partnerTnIn := MustFromForteClass(zp)
		switch sc.kind {
		case TnIn:
			return []SetClass{partnerTnIn}, nil
		case Tn:
			return partnerTnIn.TnClasses(), nil
		default: // TTO
			partnerTTO := partnerTnIn.MustToTTO()
			if partnerTTO == sc {
				return []SetClass{}, nil
			}
			return []SetClass{partnerTTO}, nil
		}
	}
	if n > MaxBitmaskModulus {
		return nil, fmt.Errorf("%w: modulus %d > %d", ErrModulusTooLarge, n, MaxBitmaskModulus)
	}
	if _, ok := binomialBounded(n, k, MaxEnumerationSets); !ok {
		return nil, fmt.Errorf("%w: C(%d, %d) > %d", ErrEnumerationLimitExceeded, n, k, MaxEnumerationSets)
	}
	if k <= 2 || k >= n-2 || n <= 7 {
		return []SetClass{}, nil
	}
	candidates := MustByIntervalVector(sc.IntervalVector(), sc.kind)
	out := []SetClass{}
	scTnIn := sc.ToTnIn()
	for _, c := range candidates {
		if c != sc && c.ToTnIn() != scTnIn {
			out = append(out, c)
		}
	}
	return out, nil
}

// MustZPartners is like [SetClass.ZPartners] but panics on error.
func (sc SetClass) MustZPartners() []SetClass {
	z, err := sc.ZPartners()
	if err != nil {
		panic(err)
	}
	return z
}

// HasZPartner reports whether sc has at least one Z-partner in sc.Kind() and sc.Modulus().
func (sc SetClass) HasZPartner() (bool, error) {
	partners, err := sc.ZPartners()
	if err != nil {
		return false, err
	}
	return len(partners) > 0, nil
}

// MustHasZPartner is like [SetClass.HasZPartner] but panics on error.
func (sc SetClass) MustHasZPartner() bool {
	has, err := sc.HasZPartner()
	if err != nil {
		panic(err)
	}
	return has
}
