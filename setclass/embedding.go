package setclass

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/go-muse/musex/pcset"
	"github.com/go-muse/musex/pitchclass"
)

// SubsetProfileEntry represents an entry in the subset-class profile of a set class Y:
// an abstract subclass Subclass and its Lewin embedding multiplicity Count = EMB(Subclass, Y) >= 1.
// The zero value SubsetProfileEntry{} is invalid (IsValid() == false).
type SubsetProfileEntry struct {
	Subclass SetClass
	Count    int
}

// IsValid reports whether e contains a valid Subclass and a positive Count.
func (e SubsetProfileEntry) IsValid() bool {
	return e.Subclass.IsValid() && e.Count > 0
}

// Equal reports whether e and other are both valid and equal.
func (e SubsetProfileEntry) Equal(other SubsetProfileEntry) bool {
	return e.IsValid() && other.IsValid() && e == other
}

// Compare orders SubsetProfileEntries by Subclass.Compare, then by Count.
// Any invalid entry is ordered before all valid entries; two invalid entries return 0.
func (e SubsetProfileEntry) Compare(other SubsetProfileEntry) int {
	ev, ov := e.IsValid(), other.IsValid()
	if !ev && !ov {
		return 0
	}
	if !ev {
		return -1
	}
	if !ov {
		return 1
	}
	if c := e.Subclass.Compare(other.Subclass); c != 0 {
		return c
	}
	return cmp.Compare(e.Count, other.Count)
}

// String returns "<Subclass>: <Count>", or "SubsetProfileEntry(invalid)" if !e.IsValid().
func (e SubsetProfileEntry) String() string {
	if !e.IsValid() {
		return "SubsetProfileEntry(invalid)"
	}
	return fmt.Sprintf("%s: %d", e.Subclass.String(), e.Count)
}

func binomialBounded(n, k, limit int) (int, bool) {
	if k > n-k {
		k = n - k
	}
	res := 1
	for i := 1; i <= k; i++ {
		res = (res * (n - k + i)) / i
		if res > limit {
			return 0, false
		}
	}
	return res, true
}

func (sc SetClass) hasMemberSubsetOf(superSet pcset.Set) bool {
	n := sc.prime.Modulus()
	period := n / sc.TnSymmetry()
	for _, tc := range sc.TnClasses() {
		for t := range period {
			if tc.prime.Transpose(t).MustIsSubsetOf(superSet) {
				return true
			}
		}
	}
	return false
}

// IsSubclassOf reports whether sc is an abstract subclass of super (sc subseteq_G super):
// at least one member set A in sc is a literal subset of super.PrimeForm().
// Trivial cardinality cases (|sc| == 0, |super| == N, sc == super, |sc| > |super|, or |sc| == |super| && sc != super)
// are answered in O(1) for any valid modulus N. Non-trivial cases (0 < |sc| < |super| < N) require N <= [MaxOrbitModulus].
func (sc SetClass) IsSubclassOf(super SetClass) (bool, error) {
	if err := validateClassPair(sc, super); err != nil {
		return false, err
	}
	return sc.IsEmbeddedInSet(super.prime)
}

// MustIsSubclassOf is like [SetClass.IsSubclassOf] but panics on error.
func (sc SetClass) MustIsSubclassOf(super SetClass) bool {
	ok, err := sc.IsSubclassOf(super)
	if err != nil {
		panic(err)
	}
	return ok
}

// IsProperSubclassOf reports whether sc is a proper abstract subclass of super (sc subset_G super):
// sc.IsSubclassOf(super) && sc.Cardinality() < super.Cardinality().
func (sc SetClass) IsProperSubclassOf(super SetClass) (bool, error) {
	sub, err := sc.IsSubclassOf(super)
	if err != nil {
		return false, err
	}
	return sub && sc.prime.Cardinality() < super.prime.Cardinality(), nil
}

// MustIsProperSubclassOf is like [SetClass.IsProperSubclassOf] but panics on error.
func (sc SetClass) MustIsProperSubclassOf(super SetClass) bool {
	ok, err := sc.IsProperSubclassOf(super)
	if err != nil {
		panic(err)
	}
	return ok
}

// Includes reports whether sc abstractly includes sub (sub subseteq_G sc).
func (sc SetClass) Includes(sub SetClass) (bool, error) {
	return sub.IsSubclassOf(sc)
}

// MustIncludes is like [SetClass.Includes] but panics on error.
func (sc SetClass) MustIncludes(sub SetClass) bool {
	ok, err := sc.Includes(sub)
	if err != nil {
		panic(err)
	}
	return ok
}

// StrictlyIncludes reports whether sc properly abstractly includes sub (sub subset_G sc).
func (sc SetClass) StrictlyIncludes(sub SetClass) (bool, error) {
	return sub.IsProperSubclassOf(sc)
}

// MustStrictlyIncludes is like [SetClass.StrictlyIncludes] but panics on error.
func (sc SetClass) MustStrictlyIncludes(sub SetClass) bool {
	ok, err := sc.StrictlyIncludes(sub)
	if err != nil {
		panic(err)
	}
	return ok
}

// IncludesSet reports whether sc abstractly includes the pitch-class set sub ([sub]_G subseteq_G sc).
func (sc SetClass) IncludesSet(sub pcset.Set) (bool, error) {
	if err := validateSetPairAndKind(sc.prime, sub, sc.kind, false); err != nil {
		return false, err
	}
	subClass := SetClass{kind: sc.kind, prime: computePrimeForm(sub, sc.kind, Rahn)}
	return subClass.IsSubclassOf(sc)
}

// MustIncludesSet is like [SetClass.IncludesSet] but panics on error.
func (sc SetClass) MustIncludesSet(sub pcset.Set) bool {
	ok, err := sc.IncludesSet(sub)
	if err != nil {
		panic(err)
	}
	return ok
}

// IsEmbeddedInSet reports whether at least one member set A in sc is a literal subset of super.
func (sc SetClass) IsEmbeddedInSet(super pcset.Set) (bool, error) {
	if err := validateSetPairAndKind(sc.prime, super, sc.kind, false); err != nil {
		return false, err
	}
	n := sc.prime.Modulus()
	kx, ky := sc.prime.Cardinality(), super.Cardinality()
	if kx > ky {
		return false, nil
	}
	if kx == 0 || ky == n {
		return true, nil
	}
	if kx == ky {
		return sc.Contains(super), nil
	}
	if n > MaxOrbitModulus {
		return false, fmt.Errorf("%w: modulus %d > %d", ErrModulusTooLarge, n, MaxOrbitModulus)
	}
	return sc.hasMemberSubsetOf(super), nil
}

// MustIsEmbeddedInSet is like [SetClass.IsEmbeddedInSet] but panics on error.
func (sc SetClass) MustIsEmbeddedInSet(super pcset.Set) bool {
	ok, err := sc.IsEmbeddedInSet(super)
	if err != nil {
		panic(err)
	}
	return ok
}

// EmbeddingCountInSet returns David Lewin's embedding number EMB(sc, super) = |{A in sc : A subseteq super}|
// (1 <= N <= 1024 for Tn/TnIn; 1 <= N <= 256 for TTO).
func (sc SetClass) EmbeddingCountInSet(super pcset.Set) (int, error) {
	if err := validateSetPairAndKind(sc.prime, super, sc.kind, true); err != nil {
		return 0, err
	}
	n := sc.prime.Modulus()
	kx, ky := sc.prime.Cardinality(), super.Cardinality()
	if kx > ky {
		return 0, nil
	}
	if kx == 0 {
		return 1, nil
	}
	if ky == n {
		return sc.OrbitSize(), nil
	}
	period := n / sc.TnSymmetry()
	count := 0
	for _, tc := range sc.TnClasses() {
		for t := range period {
			if tc.prime.Transpose(t).MustIsSubsetOf(super) {
				count++
			}
		}
	}
	return count, nil
}

// MustEmbeddingCountInSet is like [SetClass.EmbeddingCountInSet] but panics on error.
func (sc SetClass) MustEmbeddingCountInSet(super pcset.Set) int {
	c, err := sc.EmbeddingCountInSet(super)
	if err != nil {
		panic(err)
	}
	return c
}

// EmbeddingCount returns David Lewin's embedding number EMB(sc, super) = |{A in sc : A subseteq super.PrimeForm()}|.
func (sc SetClass) EmbeddingCount(super SetClass) (int, error) {
	if err := validateClassPair(sc, super); err != nil {
		return 0, err
	}
	return sc.EmbeddingCountInSet(super.prime)
}

// MustEmbeddingCount is like [SetClass.EmbeddingCount] but panics on error.
func (sc SetClass) MustEmbeddingCount(super SetClass) int {
	c, err := sc.EmbeddingCount(super)
	if err != nil {
		panic(err)
	}
	return c
}

// CountEmbedding returns sub.EmbeddingCount(super).
func CountEmbedding(sub, super SetClass) (int, error) {
	return sub.EmbeddingCount(super)
}

// MustCountEmbedding is like [CountEmbedding] but panics on error.
func MustCountEmbedding(sub, super SetClass) int {
	c, err := CountEmbedding(sub, super)
	if err != nil {
		panic(err)
	}
	return c
}

// CountSetEmbedding returns the number of distinct sets in [sub]_kind that are literal subsets of super.
func CountSetEmbedding(sub, super pcset.Set, kind Kind) (int, error) {
	sc, err := FromSet(sub, kind)
	if err != nil {
		return 0, err
	}
	return sc.EmbeddingCountInSet(super)
}

// MustCountSetEmbedding is like [CountSetEmbedding] but panics on error.
func MustCountSetEmbedding(sub, super pcset.Set, kind Kind) int {
	c, err := CountSetEmbedding(sub, super, kind)
	if err != nil {
		panic(err)
	}
	return c
}

// FindSetEmbeddingOperators returns all operators f in kind such that f(sub) is a subset of super,
// sorted by [pitchclass.Operator.Compare].
func FindSetEmbeddingOperators(sub, super pcset.Set, kind Kind) ([]pitchclass.Operator, error) {
	if err := validateSetPairAndKind(sub, super, kind, true); err != nil {
		return nil, err
	}
	return filterGroupOperators(sub, kind, func(img pcset.Set) bool {
		return img.MustIsSubsetOf(super)
	}), nil
}

// MustFindSetEmbeddingOperators is like [FindSetEmbeddingOperators] but panics on error.
func MustFindSetEmbeddingOperators(sub, super pcset.Set, kind Kind) []pitchclass.Operator {
	ops, err := FindSetEmbeddingOperators(sub, super, kind)
	if err != nil {
		panic(err)
	}
	return ops
}

// EmbeddingOperatorsInSet returns all operators f in sc.Kind() mapping sc.PrimeForm() into a subset of super,
// sorted by [pitchclass.Operator.Compare].
func (sc SetClass) EmbeddingOperatorsInSet(super pcset.Set) ([]pitchclass.Operator, error) {
	if !sc.IsValid() || !super.IsValid() {
		return nil, fmt.Errorf("%w: invalid SetClass or pcset.Set", ErrInvalidValue)
	}
	return FindSetEmbeddingOperators(sc.prime, super, sc.kind)
}

// MustEmbeddingOperatorsInSet is like [SetClass.EmbeddingOperatorsInSet] but panics on error.
func (sc SetClass) MustEmbeddingOperatorsInSet(super pcset.Set) []pitchclass.Operator {
	ops, err := sc.EmbeddingOperatorsInSet(super)
	if err != nil {
		panic(err)
	}
	return ops
}

// EmbeddingOperators returns all operators f in sc.Kind() mapping sc.PrimeForm() into a subset of super.PrimeForm(),
// sorted by [pitchclass.Operator.Compare].
func (sc SetClass) EmbeddingOperators(super SetClass) ([]pitchclass.Operator, error) {
	if err := validateClassPair(sc, super); err != nil {
		return nil, err
	}
	return sc.EmbeddingOperatorsInSet(super.prime)
}

// MustEmbeddingOperators is like [SetClass.EmbeddingOperators] but panics on error.
func (sc SetClass) MustEmbeddingOperators(super SetClass) []pitchclass.Operator {
	ops, err := sc.EmbeddingOperators(super)
	if err != nil {
		panic(err)
	}
	return ops
}

// FindEmbeddingOperators returns sub.EmbeddingOperators(super).
func FindEmbeddingOperators(sub, super SetClass) ([]pitchclass.Operator, error) {
	return sub.EmbeddingOperators(super)
}

// MustFindEmbeddingOperators is like [FindEmbeddingOperators] but panics on error.
func MustFindEmbeddingOperators(sub, super SetClass) []pitchclass.Operator {
	ops, err := FindEmbeddingOperators(sub, super)
	if err != nil {
		panic(err)
	}
	return ops
}

// SubclassProfileOfCardinality returns the subset-class profile of sc at the given cardinality k in [0, |sc|],
// sorted by [SetClass.Compare]. It requires N <= [MaxOrbitModulus] and C(|sc|, k) <= [MaxEnumerationSets].
func (sc SetClass) SubclassProfileOfCardinality(cardinality int) ([]SubsetProfileEntry, error) {
	n, err := sc.validateOrbitReceiver()
	if err != nil {
		return nil, err
	}
	ky := sc.prime.Cardinality()
	if cardinality < 0 || cardinality > ky {
		return nil, fmt.Errorf("%w: cardinality %d out of range [0, %d]", ErrInvalidCardinality, cardinality, ky)
	}
	if _, ok := binomialBounded(ky, cardinality, MaxEnumerationSets); !ok {
		return nil, fmt.Errorf("%w: C(%d, %d) > %d", ErrEnumerationLimitExceeded, ky, cardinality, MaxEnumerationSets)
	}
	if cardinality == 0 {
		return []SubsetProfileEntry{{Subclass: MustEmpty(n, sc.kind), Count: 1}}, nil
	}
	if cardinality == ky {
		return []SubsetProfileEntry{{Subclass: sc, Count: 1}}, nil
	}
	elems := sc.prime.Numbers()
	idx := make([]int, cardinality)
	for i := range cardinality {
		idx[i] = i
	}
	subNums := make([]int, cardinality)
	counts := make(map[SetClass]int)
	for {
		for i, pos := range idx {
			subNums[i] = elems[pos]
		}
		subSet := pcset.MustNew(n, subNums...)
		subClass := SetClass{
			kind:  sc.kind,
			prime: computePrimeForm(subSet, sc.kind, Rahn),
		}
		counts[subClass]++

		i := cardinality - 1
		for i >= 0 && idx[i] == ky-cardinality+i {
			i--
		}
		if i < 0 {
			break
		}
		idx[i]++
		for j := i + 1; j < cardinality; j++ {
			idx[j] = idx[j-1] + 1
		}
	}
	out := make([]SubsetProfileEntry, 0, len(counts))
	for sub, c := range counts {
		out = append(out, SubsetProfileEntry{Subclass: sub, Count: c})
	}
	slices.SortFunc(out, SubsetProfileEntry.Compare)
	return out, nil
}

// MustSubclassProfileOfCardinality is like [SetClass.SubclassProfileOfCardinality] but panics on error.
func (sc SetClass) MustSubclassProfileOfCardinality(cardinality int) []SubsetProfileEntry {
	prof, err := sc.SubclassProfileOfCardinality(cardinality)
	if err != nil {
		panic(err)
	}
	return prof
}

// SubclassesOfCardinality returns all abstract subclasses of sc of the given cardinality k in [0, |sc|],
// sorted by [SetClass.Compare].
func (sc SetClass) SubclassesOfCardinality(cardinality int) ([]SetClass, error) {
	prof, err := sc.SubclassProfileOfCardinality(cardinality)
	if err != nil {
		return nil, err
	}
	out := make([]SetClass, len(prof))
	for i, e := range prof {
		out[i] = e.Subclass
	}
	return out, nil
}

// MustSubclassesOfCardinality is like [SetClass.SubclassesOfCardinality] but panics on error.
func (sc SetClass) MustSubclassesOfCardinality(cardinality int) []SetClass {
	subs, err := sc.SubclassesOfCardinality(cardinality)
	if err != nil {
		panic(err)
	}
	return subs
}

// SubclassProfile returns the complete subset-class profile of sc across all cardinalities 0..|sc|,
// sorted by [SetClass.Compare]. It requires N <= [MaxOrbitModulus] and |sc| <= [MaxAllEnumerationModulus] (20).
func (sc SetClass) SubclassProfile() ([]SubsetProfileEntry, error) {
	if _, err := sc.validateOrbitReceiver(); err != nil {
		return nil, err
	}
	ky := sc.prime.Cardinality()
	if ky > MaxAllEnumerationModulus {
		return nil, fmt.Errorf("%w: cardinality %d > %d", ErrEnumerationLimitExceeded, ky, MaxAllEnumerationModulus)
	}
	var out []SubsetProfileEntry
	for k := range ky + 1 {
		prof := sc.MustSubclassProfileOfCardinality(k)
		out = append(out, prof...)
	}
	return out, nil
}

// MustSubclassProfile is like [SetClass.SubclassProfile] but panics on error.
func (sc SetClass) MustSubclassProfile() []SubsetProfileEntry {
	prof, err := sc.SubclassProfile()
	if err != nil {
		panic(err)
	}
	return prof
}

// Subclasses returns all abstract subclasses of sc across all cardinalities 0..|sc|,
// sorted by [SetClass.Compare]. It requires N <= [MaxOrbitModulus] and |sc| <= [MaxAllEnumerationModulus] (20).
func (sc SetClass) Subclasses() ([]SetClass, error) {
	prof, err := sc.SubclassProfile()
	if err != nil {
		return nil, err
	}
	out := make([]SetClass, len(prof))
	for i, e := range prof {
		out[i] = e.Subclass
	}
	return out, nil
}

// MustSubclasses is like [SetClass.Subclasses] but panics on error.
func (sc SetClass) MustSubclasses() []SetClass {
	subs, err := sc.Subclasses()
	if err != nil {
		panic(err)
	}
	return subs
}

// SuperclassesOfCardinality returns all abstract superclasses of sc of the given cardinality k in [|sc|, N],
// sorted by [SetClass.Compare]. It requires N <= [MaxOrbitModulus] and C(N - |sc|, k - |sc|) <= [MaxEnumerationSets].
func (sc SetClass) SuperclassesOfCardinality(cardinality int) ([]SetClass, error) {
	n, err := sc.validateOrbitReceiver()
	if err != nil {
		return nil, err
	}
	kx := sc.prime.Cardinality()
	if cardinality < kx || cardinality > n {
		return nil, fmt.Errorf("%w: cardinality %d out of range [%d, %d]", ErrInvalidCardinality, cardinality, kx, n)
	}
	rem := n - kx
	add := cardinality - kx
	if _, ok := binomialBounded(rem, add, MaxEnumerationSets); !ok {
		return nil, fmt.Errorf("%w: C(%d, %d) > %d", ErrEnumerationLimitExceeded, rem, add, MaxEnumerationSets)
	}
	if add == 0 {
		return []SetClass{sc}, nil
	}
	if cardinality == n {
		return []SetClass{MustUniverse(n, sc.kind)}, nil
	}
	compElems := sc.prime.Complement().Numbers()
	baseElems := sc.prime.Numbers()
	supNums := make([]int, cardinality)
	copy(supNums, baseElems)
	idx := make([]int, add)
	for i := range add {
		idx[i] = i
	}
	seen := make(map[SetClass]struct{})
	for {
		for i, pos := range idx {
			supNums[kx+i] = compElems[pos]
		}
		supSet := pcset.MustNew(n, supNums...)
		supClass := SetClass{
			kind:  sc.kind,
			prime: computePrimeForm(supSet, sc.kind, Rahn),
		}
		seen[supClass] = struct{}{}

		i := add - 1
		for i >= 0 && idx[i] == rem-add+i {
			i--
		}
		if i < 0 {
			break
		}
		idx[i]++
		for j := i + 1; j < add; j++ {
			idx[j] = idx[j-1] + 1
		}
	}
	out := make([]SetClass, 0, len(seen))
	for sup := range seen {
		out = append(out, sup)
	}
	slices.SortFunc(out, SetClass.Compare)
	return out, nil
}

// MustSuperclassesOfCardinality is like [SetClass.SuperclassesOfCardinality] but panics on error.
func (sc SetClass) MustSuperclassesOfCardinality(cardinality int) []SetClass {
	sups, err := sc.SuperclassesOfCardinality(cardinality)
	if err != nil {
		panic(err)
	}
	return sups
}

// Superclasses returns all abstract superclasses of sc across all cardinalities |sc|..N,
// sorted by [SetClass.Compare]. It requires N <= [MaxOrbitModulus] and N - |sc| <= [MaxAllEnumerationModulus] (20).
func (sc SetClass) Superclasses() ([]SetClass, error) {
	n, err := sc.validateOrbitReceiver()
	if err != nil {
		return nil, err
	}
	kx := sc.prime.Cardinality()
	if n-kx > MaxAllEnumerationModulus {
		return nil, fmt.Errorf("%w: N - |sc| = %d > %d", ErrEnumerationLimitExceeded, n-kx, MaxAllEnumerationModulus)
	}
	var out []SetClass
	for k := kx; k <= n; k++ {
		sups := sc.MustSuperclassesOfCardinality(k)
		out = append(out, sups...)
	}
	return out, nil
}

// MustSuperclasses is like [SetClass.Superclasses] but panics on error.
func (sc SetClass) MustSuperclasses() []SetClass {
	sups, err := sc.Superclasses()
	if err != nil {
		panic(err)
	}
	return sups
}
