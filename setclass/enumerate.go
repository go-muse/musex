package setclass

import (
	"fmt"
	"math"
	"math/big"
	"math/bits"
	"slices"

	"github.com/go-muse/musex/pcset"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/setclass/catalog"
)

func gcdInt(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func validateBoundedModulusKindAlg(modulus, maxModulus int, kind Kind, alg Algorithm) error {
	if modulus < 1 {
		return fmt.Errorf("%w: %d", ErrInvalidModulus, modulus)
	}
	if !kind.IsValid() {
		return fmt.Errorf("%w: %v", ErrInvalidKind, kind)
	}
	if !alg.IsValid() {
		return fmt.Errorf("%w: %v", ErrInvalidAlgorithm, alg)
	}
	if (kind == TTO && modulus > MaxTTOModulus) || modulus > maxModulus {
		return fmt.Errorf("%w: modulus %d exceeds limit", ErrModulusTooLarge, modulus)
	}
	return nil
}

func validateEnumerationParams(modulus, maxModulus, cardinality, minCard, maxCard, limit int, kind Kind, alg Algorithm) error {
	if err := validateBoundedModulusKindAlg(modulus, maxModulus, kind, alg); err != nil {
		return err
	}
	if cardinality < minCard || cardinality > maxCard {
		return fmt.Errorf("%w: cardinality %d out of range [%d, %d]", ErrInvalidCardinality, cardinality, minCard, maxCard)
	}
	if limit > 0 {
		if _, ok := binomialBounded(maxCard-minCard, cardinality-minCard, limit); !ok {
			return fmt.Errorf("%w: C(%d, %d) > %d", ErrEnumerationLimitExceeded, maxCard-minCard, cardinality-minCard, limit)
		}
	}
	return nil
}

func bigCountToInt(b *big.Int, err error) (int, error) {
	if err != nil {
		return 0, err
	}
	if !b.IsInt64() || b.Int64() > int64(math.MaxInt) {
		return 0, fmt.Errorf("%w: count exceeds math.MaxInt", ErrOverflow)
	}
	return int(b.Int64()), nil
}

// CountBig returns the exact number of set classes of kind across all cardinalities 0..modulus in Z_modulus
// using Burnside's lemma (1 <= modulus <= 1024 for Tn/TnIn; 1 <= modulus <= 256 for TTO).
func CountBig(modulus int, kind Kind) (*big.Int, error) {
	if err := validateBoundedModulusKindAlg(modulus, MaxOrbitModulus, kind, Rahn); err != nil {
		return nil, err
	}
	sum := new(big.Int)
	if kind == Tn || (kind == TnIn && modulus <= 2) {
		for d := 1; d <= modulus; d++ {
			if modulus%d != 0 {
				continue
			}
			phi := int64(pitchclass.MustTotient(modulus / d))
			term := big.NewInt(phi)
			term.Lsh(term, uint(d))
			sum.Add(sum, term)
		}
		return sum.Div(sum, big.NewInt(int64(modulus))), nil
	}
	if kind == TnIn {
		for d := 1; d <= modulus; d++ {
			if modulus%d != 0 {
				continue
			}
			phi := int64(pitchclass.MustTotient(modulus / d))
			term := big.NewInt(phi)
			term.Lsh(term, uint(d))
			sum.Add(sum, term)
		}
		if modulus%2 == 1 {
			inv := big.NewInt(int64(modulus))
			inv.Lsh(inv, uint((modulus+1)/2))
			sum.Add(sum, inv)
		} else {
			inv := big.NewInt(int64(3 * modulus))
			inv.Lsh(inv, uint(modulus/2-1))
			sum.Add(sum, inv)
		}
		return sum.Div(sum, big.NewInt(int64(2*modulus))), nil
	}
	// TTO: iterate over all operators x -> (m*x + t) mod N
	var visited [MaxTTOModulus]bool
	groupOrder := int64(0)
	for m := range pitchclass.MustInvertibleMultipliers(modulus) {
		for t := range modulus {
			groupOrder++
			clear(visited[:modulus])
			cycles := 0
			for i := range modulus {
				if visited[i] {
					continue
				}
				cycles++
				for cur := i; !visited[cur]; cur = (m*cur + t) % modulus {
					visited[cur] = true
				}
			}
			term := big.NewInt(1)
			term.Lsh(term, uint(cycles))
			sum.Add(sum, term)
		}
	}
	return sum.Div(sum, big.NewInt(groupOrder)), nil
}

// MustCountBig is like [CountBig] but panics on error.
func MustCountBig(modulus int, kind Kind) *big.Int {
	c, err := CountBig(modulus, kind)
	if err != nil {
		panic(err)
	}
	return c
}

// Count returns the exact number of set classes of kind across all cardinalities 0..modulus in Z_modulus as an int.
// If the exact count exceeds [math.MaxInt], Count returns [ErrOverflow].
func Count(modulus int, kind Kind) (int, error) {
	return bigCountToInt(CountBig(modulus, kind))
}

// MustCount is like [Count] but panics on error.
func MustCount(modulus int, kind Kind) int {
	c, err := Count(modulus, kind)
	if err != nil {
		panic(err)
	}
	return c
}

// CountOfCardinalityBig returns the exact number of set classes of kind and cardinality in Z_modulus
// using Polya-Burnside cycle index enumeration (1 <= modulus <= 1024 for Tn/TnIn; 1 <= modulus <= 256 for TTO).
func CountOfCardinalityBig(modulus, cardinality int, kind Kind) (*big.Int, error) {
	if err := validateEnumerationParams(modulus, MaxOrbitModulus, cardinality, 0, modulus, 0, kind, Rahn); err != nil {
		return nil, err
	}
	k := cardinality
	if k > modulus-k {
		k = modulus - k
	}
	if k <= 1 {
		return big.NewInt(1), nil
	}
	sum := new(big.Int)
	if kind == Tn || (kind == TnIn && modulus <= 2) {
		g := gcdInt(modulus, k)
		for l := 1; l <= g; l++ {
			if g%l != 0 {
				continue
			}
			phi := big.NewInt(int64(pitchclass.MustTotient(l)))
			binom := new(big.Int).Binomial(int64(modulus/l), int64(k/l))
			sum.Add(sum, phi.Mul(phi, binom))
		}
		return sum.Div(sum, big.NewInt(int64(modulus))), nil
	}
	if kind == TnIn {
		g := gcdInt(modulus, k)
		for l := 1; l <= g; l++ {
			if g%l != 0 {
				continue
			}
			phi := big.NewInt(int64(pitchclass.MustTotient(l)))
			binom := new(big.Int).Binomial(int64(modulus/l), int64(k/l))
			sum.Add(sum, phi.Mul(phi, binom))
		}
		inv := new(big.Int)
		switch {
		case modulus%2 == 1:
			inv.Binomial(int64((modulus-1)/2), int64(k/2))
		case k%2 == 0:
			inv.Binomial(int64(modulus/2), int64(k/2))
		default:
			inv.Binomial(int64((modulus-2)/2), int64((k-1)/2))
		}
		inv.Mul(inv, big.NewInt(int64(modulus)))
		sum.Add(sum, inv)
		return sum.Div(sum, big.NewInt(int64(2*modulus))), nil
	}
	// TTO: group operators by cycle-length partition to run polynomial DP only once per distinct cycle type.
	var visited [MaxTTOModulus]bool
	type cycleProfile [MaxTTOModulus + 1]uint16
	profileCounts := make(map[cycleProfile]int64)
	groupOrder := int64(0)
	for m := range pitchclass.MustInvertibleMultipliers(modulus) {
		for t := range modulus {
			groupOrder++
			clear(visited[:modulus])
			var prof cycleProfile
			for i := range modulus {
				if visited[i] {
					continue
				}
				length := 0
				for cur := i; !visited[cur]; cur = (m*cur + t) % modulus {
					visited[cur] = true
					length++
				}
				if length <= k {
					prof[length]++
				}
			}
			profileCounts[prof]++
		}
	}
	dp := make([]*big.Int, k+1)
	for i := range dp {
		dp[i] = new(big.Int)
	}
	tmp := new(big.Int)
	for prof, mult := range profileCounts {
		dp[0].SetInt64(1)
		for i := 1; i <= k; i++ {
			dp[i].SetInt64(0)
		}
		maxDeg := 0
		for length := 1; length <= k; length++ {
			cnt := int(prof[length])
			if cnt == 0 {
				continue
			}
			for range cnt {
				maxDeg = min(k, maxDeg+length)
				for deg := maxDeg; deg >= length; deg-- {
					dp[deg].Add(dp[deg], dp[deg-length])
				}
			}
		}
		if dp[k].Sign() > 0 {
			tmp.Mul(dp[k], big.NewInt(mult))
			sum.Add(sum, tmp)
		}
	}
	return sum.Div(sum, big.NewInt(groupOrder)), nil
}

// MustCountOfCardinalityBig is like [CountOfCardinalityBig] but panics on error.
func MustCountOfCardinalityBig(modulus, cardinality int, kind Kind) *big.Int {
	c, err := CountOfCardinalityBig(modulus, cardinality, kind)
	if err != nil {
		panic(err)
	}
	return c
}

// CountOfCardinality returns the exact number of set classes of kind and cardinality in Z_modulus as an int.
// If the exact count exceeds [math.MaxInt], CountOfCardinality returns [ErrOverflow].
func CountOfCardinality(modulus, cardinality int, kind Kind) (int, error) {
	return bigCountToInt(CountOfCardinalityBig(modulus, cardinality, kind))
}

// MustCountOfCardinality is like [CountOfCardinality] but panics on error.
func MustCountOfCardinality(modulus, cardinality int, kind Kind) int {
	c, err := CountOfCardinality(modulus, cardinality, kind)
	if err != nil {
		panic(err)
	}
	return c
}

// isCandidatePrimeMask reports whether mask (which always has bit 0 set and 2 <= k < n <= 64)
// has its external wrap-around gap (n - lastBit) >= every internal gap between consecutive bits.
// Every canonical Rahn prime form satisfies this condition, so filtering by it preserves all orbits.
func isCandidatePrimeMask(mask uint64, n, k int) bool {
	last := 63 - bits.LeadingZeros64(mask)
	extGap := n - last
	if extGap*k < n {
		return false
	}
	prev := 0
	for rem := mask &^ 1; rem != 0; rem &= rem - 1 {
		cur := bits.TrailingZeros64(rem)
		if cur-prev > extGap {
			return false
		}
		prev = cur
	}
	return true
}

func enumerateOfCardinalityRahn(modulus, cardinality int, kind Kind) []SetClass {
	if cardinality == 0 {
		return []SetClass{MustEmpty(modulus, kind)}
	}
	if cardinality == modulus {
		return []SetClass{MustUniverse(modulus, kind)}
	}
	if cardinality == 1 {
		return []SetClass{MustChromaticCluster(modulus, 1, kind)}
	}
	if modulus == 12 {
		switch kind {
		case Tn:
			tts := catalog.MustTnTypesOfCardinality(cardinality)
			out := make([]SetClass, len(tts))
			for i, tt := range tts {
				out[i] = MustFromTnType(tt)
			}
			slices.SortFunc(out, SetClass.Compare)
			return out
		case TnIn:
			fcs := catalog.MustForteClassesOfCardinality(cardinality)
			out := make([]SetClass, len(fcs))
			for i, fc := range fcs {
				out[i] = MustFromForteClass(fc)
			}
			slices.SortFunc(out, SetClass.Compare)
			return out
		default: // TTO
			fcs := catalog.MustForteClassesOfCardinality(cardinality)
			out := make([]SetClass, 0, len(fcs))
			for _, fc := range fcs {
				out = append(out, MustFromForteClass(fc).MustToTTO())
			}
			slices.SortFunc(out, SetClass.Compare)
			return slices.Compact(out)
		}
	}
	if cardinality > modulus/2 {
		compClasses := enumerateOfCardinalityRahn(modulus, modulus-cardinality, kind)
		out := make([]SetClass, len(compClasses))
		for i, c := range compClasses {
			out[i] = c.Complement()
		}
		slices.SortFunc(out, SetClass.Compare)
		return out
	}
	seen := make(map[SetClass]struct{})
	m := (uint64(1) << uint(cardinality-1)) - 1
	limit := uint64(1) << uint(modulus-1)
	for m < limit {
		mask := (m << 1) | 1
		if isCandidatePrimeMask(mask, modulus, cardinality) {
			s := pcset.MustFromBitmask(mask, modulus)
			sc := SetClass{
				kind:  kind,
				prime: computePrimeForm(s, kind, Rahn),
			}
			seen[sc] = struct{}{}
		}
		c := m & -m
		r := m + c
		m = (((r ^ m) >> 2) / c) | r
	}
	out := make([]SetClass, 0, len(seen))
	for sc := range seen {
		out = append(out, sc)
	}
	slices.SortFunc(out, SetClass.Compare)
	return out
}

func sortByAlgorithm(classes []SetClass, alg Algorithm) {
	if alg == Rahn || len(classes) <= 1 {
		return
	}
	type keyed struct {
		sc   SetClass
		nums []int
	}
	items := make([]keyed, len(classes))
	for i, sc := range classes {
		items[i] = keyed{
			sc:   sc,
			nums: computePrimeForm(sc.prime, sc.kind, Forte).Numbers(),
		}
	}
	slices.SortFunc(items, func(a, b keyed) int {
		return slices.Compare(a.nums, b.nums)
	})
	for i, it := range items {
		classes[i] = it.sc
	}
}

// OfCardinalityWith returns all set classes of kind and cardinality in Z_modulus,
// sorted by [SetClass.CompareWith](..., alg).
// It requires 1 <= modulus <= 64, 0 <= cardinality <= modulus, and C(modulus, cardinality) <= [MaxEnumerationSets].
func OfCardinalityWith(modulus, cardinality int, kind Kind, alg Algorithm) ([]SetClass, error) {
	if err := validateEnumerationParams(modulus, MaxBitmaskModulus, cardinality, 0, modulus, MaxEnumerationSets, kind, alg); err != nil {
		return nil, err
	}
	out := enumerateOfCardinalityRahn(modulus, cardinality, kind)
	sortByAlgorithm(out, alg)
	return out, nil
}

// MustOfCardinalityWith is like [OfCardinalityWith] but panics on error.
func MustOfCardinalityWith(modulus, cardinality int, kind Kind, alg Algorithm) []SetClass {
	list, err := OfCardinalityWith(modulus, cardinality, kind, alg)
	if err != nil {
		panic(err)
	}
	return list
}

// OfCardinality returns all set classes of kind and cardinality in Z_modulus,
// sorted by [SetClass.Compare] (1 <= modulus <= 64, C(modulus, cardinality) <= [MaxEnumerationSets]).
func OfCardinality(modulus, cardinality int, kind Kind) ([]SetClass, error) {
	return OfCardinalityWith(modulus, cardinality, kind, Rahn)
}

// MustOfCardinality is like [OfCardinality] but panics on error.
func MustOfCardinality(modulus, cardinality int, kind Kind) []SetClass {
	list, err := OfCardinality(modulus, cardinality, kind)
	if err != nil {
		panic(err)
	}
	return list
}

// AllWith returns all set classes of kind across all cardinalities 0..modulus in Z_modulus (1 <= modulus <= 20),
// sorted by [SetClass.CompareWith](..., alg).
func AllWith(modulus int, kind Kind, alg Algorithm) ([]SetClass, error) {
	if err := validateBoundedModulusKindAlg(modulus, MaxAllEnumerationModulus, kind, alg); err != nil {
		return nil, err
	}
	half := modulus / 2
	byCard := make([][]SetClass, modulus+1)
	total := 0
	for k := 0; k <= half; k++ {
		byCard[k] = enumerateOfCardinalityRahn(modulus, k, kind)
		total += len(byCard[k])
	}
	for k := half + 1; k <= modulus; k++ {
		src := byCard[modulus-k]
		dst := make([]SetClass, len(src))
		for i, c := range src {
			dst[i] = c.Complement()
		}
		slices.SortFunc(dst, SetClass.Compare)
		byCard[k] = dst
		total += len(dst)
	}
	out := make([]SetClass, 0, total)
	for k := 0; k <= modulus; k++ {
		sortByAlgorithm(byCard[k], alg)
		out = append(out, byCard[k]...)
	}
	return out, nil
}

// MustAllWith is like [AllWith] but panics on error.
func MustAllWith(modulus int, kind Kind, alg Algorithm) []SetClass {
	list, err := AllWith(modulus, kind, alg)
	if err != nil {
		panic(err)
	}
	return list
}

// All returns all set classes of kind across all cardinalities 0..modulus in Z_modulus (1 <= modulus <= 20),
// sorted by [SetClass.Compare].
func All(modulus int, kind Kind) ([]SetClass, error) {
	return AllWith(modulus, kind, Rahn)
}

// MustAll is like [All] but panics on error.
func MustAll(modulus int, kind Kind) []SetClass {
	list, err := All(modulus, kind)
	if err != nil {
		panic(err)
	}
	return list
}

// ByIntervalVector returns all set classes of kind in Z_{iv.Modulus()} whose [SetClass.IntervalVector] equals iv,
// sorted by [SetClass.Compare].
func ByIntervalVector(iv pcset.IntervalVector, kind Kind) ([]SetClass, error) {
	if !iv.IsValid() {
		return nil, fmt.Errorf("%w: invalid pcset.IntervalVector", ErrInvalidValue)
	}
	n := iv.Modulus()
	total := iv.Total()
	out := []SetClass{}
	for k := 0; k <= n; k++ {
		if k*(k-1)/2 != total {
			continue
		}
		classes, err := OfCardinality(n, k, kind)
		if err != nil {
			return nil, err
		}
		for _, sc := range classes {
			if sc.IntervalVector().Equal(iv) {
				out = append(out, sc)
			}
		}
	}
	return out, nil
}

// MustByIntervalVector is like [ByIntervalVector] but panics on error.
func MustByIntervalVector(iv pcset.IntervalVector, kind Kind) []SetClass {
	list, err := ByIntervalVector(iv, kind)
	if err != nil {
		panic(err)
	}
	return list
}
