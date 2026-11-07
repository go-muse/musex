package setclass_test

import (
	"math"
	"math/big"
	"math/bits"
	"slices"
	"sync"
	"testing"

	"github.com/go-muse/musex/pcset"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/setclass"
	"github.com/go-muse/musex/setclass/catalog"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConstantsAndSentinelErrors(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 65_536, setclass.MaxModulus)
	assert.Equal(t, pcset.MaxModulus, setclass.MaxModulus)
	assert.Equal(t, 1_024, setclass.MaxOrbitModulus)
	assert.Equal(t, 256, setclass.MaxTTOModulus)
	assert.Equal(t, 64, setclass.MaxBitmaskModulus)
	assert.Equal(t, 20, setclass.MaxAllEnumerationModulus)
	assert.Equal(t, 200_000, setclass.MaxEnumerationSets)

	sentinels := []struct {
		name string
		err  error
	}{
		{name: "ErrInvalidModulus", err: setclass.ErrInvalidModulus},
		{name: "ErrModulusTooLarge", err: setclass.ErrModulusTooLarge},
		{name: "ErrEnumerationLimitExceeded", err: setclass.ErrEnumerationLimitExceeded},
		{name: "ErrInvalidCardinality", err: setclass.ErrInvalidCardinality},
		{name: "ErrInvalidPitchClass", err: setclass.ErrInvalidPitchClass},
		{name: "ErrInvalidBitmask", err: setclass.ErrInvalidBitmask},
		{name: "ErrNonInvertibleMultiplier", err: setclass.ErrNonInvertibleMultiplier},
		{name: "ErrModulusMismatch", err: setclass.ErrModulusMismatch},
		{name: "ErrKindMismatch", err: setclass.ErrKindMismatch},
		{name: "ErrNotModulus12", err: setclass.ErrNotModulus12},
		{name: "ErrInvalidValue", err: setclass.ErrInvalidValue},
		{name: "ErrInvalidKind", err: setclass.ErrInvalidKind},
		{name: "ErrInvalidAlgorithm", err: setclass.ErrInvalidAlgorithm},
		{name: "ErrNotation", err: setclass.ErrNotation},
		{name: "ErrOverflow", err: setclass.ErrOverflow},
	}

	for i, s1 := range sentinels {
		require.Error(t, s1.err, "sentinel %s must be non-nil", s1.name)
		assert.NotEmpty(t, s1.err.Error())
		for j, s2 := range sentinels {
			if i == j {
				assert.ErrorIs(t, s1.err, s2.err)
			} else {
				assert.NotErrorIs(t, s1.err, s2.err, "%s must not match %s", s1.name, s2.name)
			}
		}
	}
}

func TestKindAndAlgorithm(t *testing.T) {
	t.Parallel()

	t.Run("Kind_valid_and_invalid_values", func(t *testing.T) {
		t.Parallel()

		kinds := setclass.Kinds()
		assert.Equal(t, [3]setclass.Kind{setclass.Tn, setclass.TnIn, setclass.TTO}, kinds)

		expectedNames := map[setclass.Kind]string{
			setclass.Tn:   "Tn",
			setclass.TnIn: "TnIn",
			setclass.TTO:  "TTO",
		}

		for i, k := range kinds {
			assert.True(t, k.IsValid())
			assert.True(t, k.Equal(k))
			assert.Equal(t, 0, k.Compare(k))
			assert.Equal(t, expectedNames[k], k.String())

			parsed, err := setclass.ParseKind(k.String())
			require.NoError(t, err)
			assert.Equal(t, k, parsed)
			assert.Equal(t, k, setclass.MustParseKind(k.String()))

			for j, other := range kinds {
				if i < j {
					assert.False(t, k.Equal(other))
					assert.Equal(t, -1, k.Compare(other))
					assert.Equal(t, 1, other.Compare(k))
				}
			}
		}

		var zeroKind setclass.Kind
		invalidKind := setclass.Kind(99)
		for _, inv := range []setclass.Kind{zeroKind, invalidKind} {
			assert.False(t, inv.IsValid())
			assert.False(t, inv.Equal(inv))
			assert.False(t, inv.Equal(setclass.Tn))
			assert.False(t, setclass.Tn.Equal(inv))
			assert.Equal(t, 0, inv.Compare(zeroKind))
			assert.Equal(t, 0, inv.Compare(invalidKind))
			assert.Equal(t, -1, inv.Compare(setclass.Tn))
			assert.Equal(t, 1, setclass.Tn.Compare(inv))
			assert.Equal(t, "Kind(invalid)", inv.String())
		}
		assert.Equal(t, zeroKind, setclass.Kind(0))

		for _, badStr := range []string{"", "Kind(invalid)", "tn", "TN", "Tn/In", "SG(1)", " Tn"} {
			_, err := setclass.ParseKind(badStr)
			require.ErrorIs(t, err, setclass.ErrNotation)
			require.Panics(t, func() {
				_ = setclass.MustParseKind(badStr)
			})
		}
	})

	t.Run("Algorithm_valid_and_invalid_values", func(t *testing.T) {
		t.Parallel()

		algs := setclass.Algorithms()
		assert.Equal(t, [2]setclass.Algorithm{setclass.Rahn, setclass.Forte}, algs)

		expectedNames := map[setclass.Algorithm]string{
			setclass.Rahn:  "Rahn",
			setclass.Forte: "Forte",
		}

		for _, a := range algs {
			assert.True(t, a.IsValid())
			assert.True(t, a.Equal(a))
			assert.Equal(t, 0, a.Compare(a))
			assert.Equal(t, expectedNames[a], a.String())

			parsed, err := setclass.ParseAlgorithm(a.String())
			require.NoError(t, err)
			assert.Equal(t, a, parsed)
			assert.Equal(t, a, setclass.MustParseAlgorithm(a.String()))
		}
		assert.False(t, setclass.Rahn.Equal(setclass.Forte))
		assert.Equal(t, -1, setclass.Rahn.Compare(setclass.Forte))
		assert.Equal(t, 1, setclass.Forte.Compare(setclass.Rahn))

		var zeroAlg setclass.Algorithm
		invalidAlg := setclass.Algorithm(99)
		for _, inv := range []setclass.Algorithm{zeroAlg, invalidAlg} {
			assert.False(t, inv.IsValid())
			assert.False(t, inv.Equal(inv))
			assert.False(t, inv.Equal(setclass.Rahn))
			assert.False(t, setclass.Rahn.Equal(inv))
			assert.Equal(t, 0, inv.Compare(zeroAlg))
			assert.Equal(t, 0, inv.Compare(invalidAlg))
			assert.Equal(t, -1, inv.Compare(setclass.Rahn))
			assert.Equal(t, 1, setclass.Rahn.Compare(inv))
			assert.Equal(t, "Algorithm(invalid)", inv.String())
		}
		assert.Equal(t, zeroAlg, setclass.Algorithm(0))

		for _, badStr := range []string{"", "Algorithm(invalid)", "rahn", "forte", "Collins"} {
			_, err := setclass.ParseAlgorithm(badStr)
			require.ErrorIs(t, err, setclass.ErrNotation)
			require.Panics(t, func() {
				_ = setclass.MustParseAlgorithm(badStr)
			})
		}
	})
}

func TestSubsetProfileEntry(t *testing.T) {
	t.Parallel()

	sc1 := setclass.MustFromNumbers(12, setclass.TnIn, 0, 3, 7)
	sc2 := setclass.MustFromNumbers(12, setclass.TnIn, 0, 4, 8)

	e1 := setclass.SubsetProfileEntry{Subclass: sc1, Count: 2}
	e1Copy := setclass.SubsetProfileEntry{Subclass: sc1, Count: 2}
	e2 := setclass.SubsetProfileEntry{Subclass: sc1, Count: 6}
	e3 := setclass.SubsetProfileEntry{Subclass: sc2, Count: 1}

	assert.True(t, e1.IsValid())
	assert.True(t, e1.Equal(e1Copy))
	assert.Equal(t, e1, e1Copy)
	assert.False(t, e1.Equal(e2))
	assert.Equal(t, 0, e1.Compare(e1Copy))
	assert.Equal(t, -1, e1.Compare(e2))
	assert.Equal(t, 1, e2.Compare(e1))
	assert.Equal(t, -1, e1.Compare(e3))
	assert.Equal(t, 1, e3.Compare(e1))
	assert.Equal(t, "TnIn(0, 3, 7) mod 12: 2", e1.String())

	profileMap := map[setclass.SubsetProfileEntry]string{e1: "triad"}
	assert.Equal(t, "triad", profileMap[e1Copy])

	invalids := []setclass.SubsetProfileEntry{
		{},
		{Subclass: sc1, Count: 0},
		{Subclass: sc1, Count: -3},
		{Subclass: setclass.SetClass{}, Count: 2},
	}
	for _, inv := range invalids {
		assert.False(t, inv.IsValid())
		assert.False(t, inv.Equal(inv))
		assert.False(t, inv.Equal(e1))
		assert.False(t, e1.Equal(inv))
		assert.Equal(t, 0, inv.Compare(setclass.SubsetProfileEntry{}))
		assert.Equal(t, -1, inv.Compare(e1))
		assert.Equal(t, 1, e1.Compare(inv))
		assert.Equal(t, "SubsetProfileEntry(invalid)", inv.String())
	}
}

func TestZeroSetClassAndInvalidOperands(t *testing.T) {
	t.Parallel()

	var zero setclass.SetClass
	valid12 := setclass.MustFromNumbers(12, setclass.TnIn, 0, 3, 7)
	valid19 := setclass.MustFromNumbers(19, setclass.Tn, 0, 3, 7)
	set12 := pcset.MustNew(12, 0, 3, 7)

	// Non-error getters on zero SetClass must never panic and must return zero/invalid values (E-7.2, K-14).
	assert.False(t, zero.IsValid())
	assert.Equal(t, 0, zero.Modulus())
	assert.Equal(t, setclass.Kind(0), zero.Kind())
	assert.False(t, zero.Kind().IsValid())
	assert.Equal(t, 0, zero.Cardinality())
	assert.False(t, zero.IsEmpty())
	assert.False(t, zero.IsUniverse())
	assert.Equal(t, 0, zero.MinSpan())
	assert.False(t, zero.RahnForteDiffer())
	assert.False(t, zero.PrimeForm().IsValid())
	assert.False(t, zero.FortePrimeForm().IsValid())
	assert.Nil(t, zero.PrimeClasses())
	assert.Nil(t, zero.PrimeNumbers())
	assert.Nil(t, zero.PrimeMaskBig())
	assert.False(t, zero.IntervalVector().IsValid())
	assert.Equal(t, 0, zero.TnSymmetry())
	assert.Equal(t, 0, zero.InSymmetry())
	assert.Equal(t, 0, zero.SymmetryDegreeTI())
	assert.Equal(t, 0, zero.OrbitSize())
	assert.False(t, zero.IsTranspositionallySymmetric())
	assert.False(t, zero.IsInversionallySymmetric())
	assert.False(t, zero.IsChiral())
	assert.False(t, zero.IsSelfComplementary())
	assert.False(t, zero.Complement().IsValid())
	assert.False(t, zero.Inversion().IsValid())
	assert.False(t, zero.ToTnIn().IsValid())
	assert.Nil(t, zero.TnClasses())
	assert.False(t, zero.Contains(set12))
	assert.False(t, zero.Contains(pcset.Set{}))
	assert.False(t, valid12.Contains(pcset.Set{}))
	assert.False(t, zero.Equal(zero))
	assert.False(t, zero.Equal(valid12))
	assert.False(t, valid12.Equal(zero))
	assert.Equal(t, setclass.SetClass{}, zero)
	assert.Equal(t, 0, zero.Compare(setclass.SetClass{}))
	assert.Equal(t, -1, zero.Compare(valid12))
	assert.Equal(t, 1, valid12.Compare(zero))
	assert.Equal(t, "SetClass(invalid)", zero.String())
	assert.Equal(t, "SetClass(invalid)", zero.Canonical())

	// Error-returning methods on zero SetClass must return ErrInvalidValue (taking precedence over E-5 and E-6)
	// and Must-variants must panic.
	errChecks := []struct {
		name   string
		call   func() error
		mustFn func()
	}{
		{
			name:   "ForteClass",
			call:   func() error { _, err := zero.ForteClass(); return err },
			mustFn: func() { _ = zero.MustForteClass() },
		},
		{
			name:   "MustForteClass_func",
			call:   func() error { _, err := zero.ForteClass(); return err },
			mustFn: func() { _ = setclass.MustForteClass(zero) },
		},
		{
			name:   "TnType",
			call:   func() error { _, err := zero.TnType(); return err },
			mustFn: func() { _ = zero.MustTnType() },
		},
		{
			name:   "CatalogTnTypes",
			call:   func() error { _, err := zero.CatalogTnTypes(); return err },
			mustFn: func() { _ = zero.MustCatalogTnTypes() },
		},
		{
			name:   "CarterClass",
			call:   func() error { _, _, err := zero.CarterClass(); return err },
			mustFn: func() { _, _ = zero.MustCarterClass() },
		},
		{
			name:   "CarterNumber",
			call:   func() error { _, _, err := zero.CarterNumber(); return err },
			mustFn: func() { _, _ = zero.MustCarterNumber() },
		},
		{
			name:   "ZPair",
			call:   func() error { _, _, err := zero.ZPair(); return err },
			mustFn: func() { _, _ = zero.MustZPair() },
		},
		{
			name:   "ZPartner",
			call:   func() error { _, _, err := zero.ZPartner(); return err },
			mustFn: func() { _, _ = zero.MustZPartner() },
		},
		{
			name:   "CatalogName",
			call:   func() error { _, err := zero.CatalogName(); return err },
			mustFn: func() { _ = zero.MustCatalogName() },
		},
		{
			name:   "CompareWith_zero_receiver",
			call:   func() error { _, err := zero.CompareWith(valid19, setclass.Rahn); return err },
			mustFn: func() { _ = zero.MustCompareWith(valid19, setclass.Rahn) },
		},
		{
			name:   "CompareWith_zero_other",
			call:   func() error { _, err := valid12.CompareWith(zero, setclass.Rahn); return err },
			mustFn: func() { _ = valid12.MustCompareWith(zero, setclass.Rahn) },
		},
		{
			name:   "PrimeFormWith",
			call:   func() error { _, err := zero.PrimeFormWith(setclass.Rahn); return err },
			mustFn: func() { _ = zero.MustPrimeFormWith(setclass.Rahn) },
		},
		{
			name:   "PrimeClassesWith",
			call:   func() error { _, err := zero.PrimeClassesWith(setclass.Rahn); return err },
			mustFn: func() { _ = zero.MustPrimeClassesWith(setclass.Rahn) },
		},
		{
			name:   "PrimeNumbersWith",
			call:   func() error { _, err := zero.PrimeNumbersWith(setclass.Rahn); return err },
			mustFn: func() { _ = zero.MustPrimeNumbersWith(setclass.Rahn) },
		},
		{
			name:   "PrimeMaskBigWith",
			call:   func() error { _, err := zero.PrimeMaskBigWith(setclass.Rahn); return err },
			mustFn: func() { _ = zero.MustPrimeMaskBigWith(setclass.Rahn) },
		},
		{
			name:   "PrimeMask",
			call:   func() error { _, err := zero.PrimeMask(); return err },
			mustFn: func() { _ = zero.MustPrimeMask() },
		},
		{
			name:   "FortePrimeMask",
			call:   func() error { _, err := zero.FortePrimeMask(); return err },
			mustFn: func() { _ = zero.MustFortePrimeMask() },
		},
		{
			name:   "PrimeMaskWith",
			call:   func() error { _, err := zero.PrimeMaskWith(setclass.Rahn); return err },
			mustFn: func() { _ = zero.MustPrimeMaskWith(setclass.Rahn) },
		},
		{
			name:   "CollinsMask",
			call:   func() error { _, err := zero.CollinsMask(); return err },
			mustFn: func() { _ = zero.MustCollinsMask() },
		},
		{
			name:   "TranspositionIndicesTo",
			call:   func() error { _, err := zero.TranspositionIndicesTo(set12); return err },
			mustFn: func() { _ = zero.MustTranspositionIndicesTo(set12) },
		},
		{
			name:   "TranspositionIndicesToWith",
			call:   func() error { _, err := zero.TranspositionIndicesToWith(set12, setclass.Rahn); return err },
			mustFn: func() { _ = zero.MustTranspositionIndicesToWith(set12, setclass.Rahn) },
		},
		{
			name:   "InversionIndicesTo",
			call:   func() error { _, err := zero.InversionIndicesTo(set12); return err },
			mustFn: func() { _ = zero.MustInversionIndicesTo(set12) },
		},
		{
			name:   "InversionIndicesToWith",
			call:   func() error { _, err := zero.InversionIndicesToWith(set12, setclass.Rahn); return err },
			mustFn: func() { _ = zero.MustInversionIndicesToWith(set12, setclass.Rahn) },
		},
		{
			name:   "OperatorsTo",
			call:   func() error { _, err := zero.OperatorsTo(set12); return err },
			mustFn: func() { _ = zero.MustOperatorsTo(set12) },
		},
		{
			name:   "OperatorsToWith",
			call:   func() error { _, err := zero.OperatorsToWith(set12, setclass.Rahn); return err },
			mustFn: func() { _ = zero.MustOperatorsToWith(set12, setclass.Rahn) },
		},
		{
			name:   "OperatorsFrom",
			call:   func() error { _, err := zero.OperatorsFrom(set12); return err },
			mustFn: func() { _ = zero.MustOperatorsFrom(set12) },
		},
		{
			name:   "OperatorsFromWith",
			call:   func() error { _, err := zero.OperatorsFromWith(set12, setclass.Rahn); return err },
			mustFn: func() { _ = zero.MustOperatorsFromWith(set12, setclass.Rahn) },
		},
		{
			name:   "Members",
			call:   func() error { _, err := zero.Members(); return err },
			mustFn: func() { _ = zero.MustMembers() },
		},
		{
			name:   "SymmetryDegreeTTO",
			call:   func() error { _, err := zero.SymmetryDegreeTTO(); return err },
			mustFn: func() { _ = zero.MustSymmetryDegreeTTO() },
		},
		{
			name:   "ToTTO",
			call:   func() error { _, err := zero.ToTTO(); return err },
			mustFn: func() { _ = zero.MustToTTO() },
		},
		{
			name:   "TnInClasses",
			call:   func() error { _, err := zero.TnInClasses(); return err },
			mustFn: func() { _ = zero.MustTnInClasses() },
		},
		{
			name:   "IntervalVectorWith",
			call:   func() error { _, err := zero.IntervalVectorWith(setclass.Rahn); return err },
			mustFn: func() { _ = zero.MustIntervalVectorWith(setclass.Rahn) },
		},
		{
			name:   "InvarianceVector",
			call:   func() error { _, err := zero.InvarianceVector(); return err },
			mustFn: func() { _ = zero.MustInvarianceVector() },
		},
		{
			name:   "Multiply",
			call:   func() error { _, err := zero.Multiply(5); return err },
			mustFn: func() { _ = zero.MustMultiply(5) },
		},
		{
			name:   "M5Partner",
			call:   func() error { _, err := zero.M5Partner(); return err },
			mustFn: func() { _ = zero.MustM5Partner() },
		},
		{
			name:   "IsM5Invariant",
			call:   func() error { _, err := zero.IsM5Invariant(); return err },
			mustFn: func() { _ = zero.MustIsM5Invariant() },
		},
		{
			name:   "IsZRelated_zero_receiver",
			call:   func() error { _, err := zero.IsZRelated(valid19); return err },
			mustFn: func() { _ = zero.MustIsZRelated(valid19) },
		},
		{
			name:   "IsZRelated_zero_other",
			call:   func() error { _, err := valid12.IsZRelated(zero); return err },
			mustFn: func() { _ = valid12.MustIsZRelated(zero) },
		},
		{
			name:   "HasZPartner",
			call:   func() error { _, err := zero.HasZPartner(); return err },
			mustFn: func() { _ = zero.MustHasZPartner() },
		},
		{
			name:   "ZPartners",
			call:   func() error { _, err := zero.ZPartners(); return err },
			mustFn: func() { _ = zero.MustZPartners() },
		},
		{
			name:   "IsSubclassOf_zero_receiver",
			call:   func() error { _, err := zero.IsSubclassOf(valid19); return err },
			mustFn: func() { _ = zero.MustIsSubclassOf(valid19) },
		},
		{
			name:   "IsSubclassOf_zero_other",
			call:   func() error { _, err := valid12.IsSubclassOf(zero); return err },
			mustFn: func() { _ = valid12.MustIsSubclassOf(zero) },
		},
		{
			name:   "IsProperSubclassOf",
			call:   func() error { _, err := zero.IsProperSubclassOf(valid12); return err },
			mustFn: func() { _ = zero.MustIsProperSubclassOf(valid12) },
		},
		{
			name:   "Includes",
			call:   func() error { _, err := zero.Includes(valid19); return err },
			mustFn: func() { _ = zero.MustIncludes(valid19) },
		},
		{
			name:   "StrictlyIncludes",
			call:   func() error { _, err := zero.StrictlyIncludes(valid12); return err },
			mustFn: func() { _ = zero.MustStrictlyIncludes(valid12) },
		},
		{
			name:   "IncludesSet",
			call:   func() error { _, err := zero.IncludesSet(set12); return err },
			mustFn: func() { _ = zero.MustIncludesSet(set12) },
		},
		{
			name:   "IsEmbeddedInSet",
			call:   func() error { _, err := zero.IsEmbeddedInSet(set12); return err },
			mustFn: func() { _ = zero.MustIsEmbeddedInSet(set12) },
		},
		{
			name:   "EmbeddingCount",
			call:   func() error { _, err := zero.EmbeddingCount(valid12); return err },
			mustFn: func() { _ = zero.MustEmbeddingCount(valid12) },
		},
		{
			name:   "EmbeddingCountInSet",
			call:   func() error { _, err := zero.EmbeddingCountInSet(set12); return err },
			mustFn: func() { _ = zero.MustEmbeddingCountInSet(set12) },
		},
		{
			name:   "EmbeddingOperators",
			call:   func() error { _, err := zero.EmbeddingOperators(valid12); return err },
			mustFn: func() { _ = zero.MustEmbeddingOperators(valid12) },
		},
		{
			name:   "EmbeddingOperatorsInSet",
			call:   func() error { _, err := zero.EmbeddingOperatorsInSet(set12); return err },
			mustFn: func() { _ = zero.MustEmbeddingOperatorsInSet(set12) },
		},
		{
			name:   "CountEmbedding",
			call:   func() error { _, err := setclass.CountEmbedding(zero, valid12); return err },
			mustFn: func() { _ = setclass.MustCountEmbedding(zero, valid12) },
		},
		{
			name:   "FindEmbeddingOperators",
			call:   func() error { _, err := setclass.FindEmbeddingOperators(zero, valid12); return err },
			mustFn: func() { _ = setclass.MustFindEmbeddingOperators(zero, valid12) },
		},
		{
			name:   "Subclasses",
			call:   func() error { _, err := zero.Subclasses(); return err },
			mustFn: func() { _ = zero.MustSubclasses() },
		},
		{
			name:   "SubclassesOfCardinality",
			call:   func() error { _, err := zero.SubclassesOfCardinality(2); return err },
			mustFn: func() { _ = zero.MustSubclassesOfCardinality(2) },
		},
		{
			name:   "SubclassProfile",
			call:   func() error { _, err := zero.SubclassProfile(); return err },
			mustFn: func() { _ = zero.MustSubclassProfile() },
		},
		{
			name:   "SubclassProfileOfCardinality",
			call:   func() error { _, err := zero.SubclassProfileOfCardinality(2); return err },
			mustFn: func() { _ = zero.MustSubclassProfileOfCardinality(2) },
		},
		{
			name:   "Superclasses",
			call:   func() error { _, err := zero.Superclasses(); return err },
			mustFn: func() { _ = zero.MustSuperclasses() },
		},
		{
			name:   "SuperclassesOfCardinality",
			call:   func() error { _, err := zero.SuperclassesOfCardinality(4); return err },
			mustFn: func() { _ = zero.MustSuperclassesOfCardinality(4) },
		},
		{
			name:   "Notation",
			call:   func() error { _, err := zero.Notation(pitchclass.TE); return err },
			mustFn: func() { _ = zero.MustNotation(pitchclass.TE) },
		},
		{
			name:   "NotationWith",
			call:   func() error { _, err := zero.NotationWith(pitchclass.TE, setclass.Rahn); return err },
			mustFn: func() { _ = zero.MustNotationWith(pitchclass.TE, setclass.Rahn) },
		},
		{
			name:   "CompactNotation",
			call:   func() error { _, err := zero.CompactNotation(pitchclass.TE); return err },
			mustFn: func() { _ = zero.MustCompactNotation(pitchclass.TE) },
		},
		{
			name:   "CompactNotationWith",
			call:   func() error { _, err := zero.CompactNotationWith(pitchclass.TE, setclass.Rahn); return err },
			mustFn: func() { _ = zero.MustCompactNotationWith(pitchclass.TE, setclass.Rahn) },
		},
		{
			name:   "FormatMorris",
			call:   func() error { _, err := zero.FormatMorris(); return err },
			mustFn: func() { _ = zero.MustFormatMorris() },
		},
		{
			name:   "FormatMorrisWith",
			call:   func() error { _, err := zero.FormatMorrisWith(setclass.Rahn); return err },
			mustFn: func() { _ = zero.MustFormatMorrisWith(setclass.Rahn) },
		},
	}

	for _, tc := range errChecks {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.call()
			require.ErrorIs(t, err, setclass.ErrInvalidValue)
			require.Panics(t, tc.mustFn)
		})
	}

	// Additional Must* panic checks on invalid arguments.
	require.Panics(t, func() { _ = setclass.MustCountSetEmbedding(pcset.Set{}, set12, setclass.TnIn) })
	require.Panics(t, func() { _ = setclass.MustFindSetEmbeddingOperators(pcset.Set{}, set12, setclass.TnIn) })
	require.Panics(t, func() { _ = setclass.MustCountBig(0, setclass.TnIn) })
	require.Panics(t, func() { _ = setclass.MustCountOfCardinalityBig(0, 0, setclass.TnIn) })
	require.Panics(t, func() { _ = setclass.MustOfCardinalityWith(0, 0, setclass.TnIn, setclass.Rahn) })
	require.Panics(t, func() { _ = setclass.MustOfCardinality(0, 0, setclass.TnIn) })
	require.Panics(t, func() { _ = setclass.MustAllWith(0, setclass.TnIn, setclass.Rahn) })
	require.Panics(t, func() { _ = setclass.MustAll(0, setclass.TnIn) })
	require.Panics(t, func() { _ = setclass.MustByIntervalVector(pcset.IntervalVector{}, setclass.TnIn) })
	require.Panics(t, func() { _ = setclass.MustParseNotation("invalid", pitchclass.TE, 12, setclass.TnIn) })
	require.Panics(t, func() { _ = setclass.MustParseMorrisWithKind("invalid", setclass.TnIn) })
	require.Panics(t, func() { _ = setclass.MustFindSetOperators(pcset.Set{}, set12, setclass.TnIn) })
	require.Panics(t, func() { _ = setclass.MustAreZRelatedSets(pcset.Set{}, set12) })
	require.Panics(t, func() { _ = valid12.MustPrimeClassesWith(setclass.Algorithm(0)) })
	require.Panics(t, func() { _, _ = valid19.MustCarterClass() })
	require.Panics(t, func() { _, _ = valid19.MustCarterNumber() })
}

func TestConstructorsAndBasicProperties(t *testing.T) {
	t.Parallel()

	t.Run("FromSet_New_NewTn_NewTnIn_NewTTO_equivalent_inputs", func(t *testing.T) {
		t.Parallel()

		cMajor := pcset.MustNew(12, 0, 4, 7)
		cMajorPerm := pcset.MustNew(12, 7, 0, 4)
		dMinor := pcset.MustNew(12, 2, 5, 9)

		sc1, err := setclass.New(cMajor, setclass.TnIn)
		require.NoError(t, err)
		sc2, err := setclass.FromSet(cMajorPerm, setclass.TnIn)
		require.NoError(t, err)
		sc3, err := setclass.NewTnIn(dMinor)
		require.NoError(t, err)

		assert.Equal(t, sc1, setclass.MustNew(cMajor, setclass.TnIn))
		assert.Equal(t, sc2, setclass.MustFromSet(cMajorPerm, setclass.TnIn))
		assert.Equal(t, sc3, setclass.MustNewTnIn(dMinor))
		assert.True(t, sc1.IsValid())
		assert.True(t, sc1.Equal(sc2))
		assert.True(t, sc1.Equal(sc3))
		assert.Equal(t, sc1, sc2)
		assert.Equal(t, sc1, sc3)
		assert.Equal(t, 12, sc1.Modulus())
		assert.Equal(t, setclass.TnIn, sc1.Kind())
		assert.Equal(t, 3, sc1.Cardinality())
		assert.False(t, sc1.IsEmpty())
		assert.False(t, sc1.IsUniverse())
		assert.Equal(t, 7, sc1.MinSpan())
		assert.False(t, sc1.RahnForteDiffer())
		assert.Equal(t, []int{0, 3, 7}, sc1.PrimeNumbers())

		// Directed Tn distinguishes C major (3-11B: [0, 4, 7]) from D minor (3-11A: [0, 3, 7]).
		tnMaj, err := setclass.NewTn(cMajor)
		require.NoError(t, err)
		tnMin, err := setclass.NewTn(dMinor)
		require.NoError(t, err)
		assert.Equal(t, tnMaj, setclass.MustNewTn(cMajor))
		assert.Equal(t, tnMin, setclass.MustNewTn(dMinor))
		assert.False(t, tnMaj.Equal(tnMin))
		assert.Equal(t, []int{0, 4, 7}, tnMaj.PrimeNumbers())
		assert.Equal(t, []int{0, 3, 7}, tnMin.PrimeNumbers())

		// TTO merges 3-3 ([0, 1, 4]) and 3-11 ([0, 3, 7]) into 3-3/11 with prime form [0, 1, 4].
		ttoMaj, err := setclass.NewTTO(cMajor)
		require.NoError(t, err)
		assert.Equal(t, ttoMaj, setclass.MustNewTTO(cMajor))
		assert.Equal(t, setclass.TTO, ttoMaj.Kind())
		assert.Equal(t, []int{0, 1, 4}, ttoMaj.PrimeNumbers())
	})

	t.Run("FromClasses_FromNumbers_ReduceNumbers_FromBitmask_FromBitmaskBig", func(t *testing.T) {
		t.Parallel()

		c0 := pitchclass.MustNew(0, 12)
		c4 := pitchclass.MustNew(4, 12)
		c7 := pitchclass.MustNew(7, 12)

		fromClasses, err := setclass.FromClasses(12, setclass.TnIn, c7, c0, c4, c0)
		require.NoError(t, err)
		assert.Equal(t, fromClasses, setclass.MustFromClasses(12, setclass.TnIn, c0, c4, c7))

		fromNumbers, err := setclass.FromNumbers(12, setclass.TnIn, 7, 0, 4, 7)
		require.NoError(t, err)
		assert.Equal(t, fromClasses, fromNumbers)
		assert.Equal(t, fromNumbers, setclass.MustFromNumbers(12, setclass.TnIn, 0, 4, 7))

		fromReduce, err := setclass.ReduceNumbers(12, setclass.TnIn, -12, 16, -5, 31)
		require.NoError(t, err)
		assert.Equal(t, fromClasses, fromReduce)
		assert.Equal(t, fromReduce, setclass.MustReduceNumbers(12, setclass.TnIn, -12, 16, -5, 31))

		// Bitmask 145 = 1 + 16 + 128 = {0, 4, 7}.
		fromMask, err := setclass.FromBitmask(12, setclass.TnIn, 145)
		require.NoError(t, err)
		assert.Equal(t, fromClasses, fromMask)
		assert.Equal(t, fromMask, setclass.MustFromBitmask(12, setclass.TnIn, 145))

		fromMaskBig, err := setclass.FromBitmaskBig(12, setclass.TnIn, big.NewInt(145))
		require.NoError(t, err)
		assert.Equal(t, fromClasses, fromMaskBig)
		assert.Equal(t, fromMaskBig, setclass.MustFromBitmaskBig(12, setclass.TnIn, big.NewInt(145)))

		// Map key compatibility.
		m := map[setclass.SetClass]int{fromClasses: 1}
		m[fromMask]++
		m[fromReduce]++
		assert.Len(t, m, 1)
		assert.Equal(t, 3, m[fromNumbers])
	})

	t.Run("Empty_Universe_ChromaticCluster_and_large_moduli", func(t *testing.T) {
		t.Parallel()

		for _, n := range []int{1, 2, 12, 19, 64, 256, 65_536} {
			for _, kind := range []setclass.Kind{setclass.Tn, setclass.TnIn, setclass.TTO} {
				if kind == setclass.TTO && n > setclass.MaxTTOModulus {
					_, err := setclass.Empty(n, kind)
					require.ErrorIs(t, err, setclass.ErrModulusTooLarge)
					continue
				}

				emp, err := setclass.Empty(n, kind)
				require.NoError(t, err)
				assert.Equal(t, emp, setclass.MustEmpty(n, kind))
				assert.True(t, emp.IsEmpty())
				assert.False(t, emp.IsUniverse())
				assert.Equal(t, 0, emp.Cardinality())
				assert.Equal(t, 0, emp.MinSpan())
				assert.Empty(t, emp.PrimeNumbers())
				assert.Equal(t, 1, emp.OrbitSize())

				uni, err := setclass.Universe(n, kind)
				require.NoError(t, err)
				assert.Equal(t, uni, setclass.MustUniverse(n, kind))
				assert.False(t, uni.IsEmpty())
				assert.True(t, uni.IsUniverse())
				assert.Equal(t, n, uni.Cardinality())
				assert.Equal(t, n-1, uni.MinSpan())
				assert.Equal(t, 1, uni.OrbitSize())
				assert.Equal(t, uni, emp.Complement())
				assert.Equal(t, emp, uni.Complement())
			}
		}

		cluster, err := setclass.ChromaticCluster(12, 4, setclass.TnIn)
		require.NoError(t, err)
		assert.Equal(t, cluster, setclass.MustChromaticCluster(12, 4, setclass.TnIn))
		assert.Equal(t, []int{0, 1, 2, 3}, cluster.PrimeNumbers())
		assert.Equal(t, 3, cluster.MinSpan())

		cluster0, err := setclass.ChromaticCluster(12, 0, setclass.TnIn)
		require.NoError(t, err)
		assert.True(t, cluster0.IsEmpty())

		// Large modulus N = 65_536 with FromBitmaskBig and ReduceNumbers (math.MinInt, math.MaxInt).
		bigMask := new(big.Int).SetBit(new(big.Int), 65_535, 1)
		bigMask.SetBit(bigMask, 0, 1)
		scBig, err := setclass.FromBitmaskBig(65_536, setclass.TnIn, bigMask)
		require.NoError(t, err)
		assert.Equal(t, []int{0, 1}, scBig.PrimeNumbers())
		assert.Equal(t, 1, scBig.MinSpan())

		scMinMax, err := setclass.ReduceNumbers(12, setclass.TnIn, math.MinInt, math.MaxInt)
		require.NoError(t, err)
		assert.True(t, scMinMax.IsValid())
		assert.Equal(t, 2, scMinMax.Cardinality())
	})

	t.Run("Constructor_error_cases_and_Must_panics", func(t *testing.T) {
		t.Parallel()

		validSet12 := pcset.MustNew(12, 0, 4, 7)
		validSet300 := pcset.MustNew(300, 0, 4, 7)
		c0Mod12 := pitchclass.MustNew(0, 12)
		c0Mod19 := pitchclass.MustNew(0, 19)

		// Invalid pcset.Set or invalid Kind in New / FromSet / NewTn / NewTnIn / NewTTO.
		_, err := setclass.New(pcset.Set{}, setclass.TnIn)
		require.ErrorIs(t, err, setclass.ErrInvalidValue)
		require.Panics(t, func() { _ = setclass.MustNew(pcset.Set{}, setclass.TnIn) })

		_, err = setclass.FromSet(validSet12, setclass.Kind(99))
		require.ErrorIs(t, err, setclass.ErrInvalidKind)
		require.Panics(t, func() { _ = setclass.MustFromSet(validSet12, setclass.Kind(99)) })

		_, err = setclass.NewTn(pcset.Set{})
		require.ErrorIs(t, err, setclass.ErrInvalidValue)
		require.Panics(t, func() { _ = setclass.MustNewTn(pcset.Set{}) })

		_, err = setclass.NewTnIn(pcset.Set{})
		require.ErrorIs(t, err, setclass.ErrInvalidValue)
		require.Panics(t, func() { _ = setclass.MustNewTnIn(pcset.Set{}) })

		_, err = setclass.NewTTO(pcset.Set{})
		require.ErrorIs(t, err, setclass.ErrInvalidValue)
		require.Panics(t, func() { _ = setclass.MustNewTTO(pcset.Set{}) })

		_, err = setclass.NewTTO(validSet300)
		require.ErrorIs(t, err, setclass.ErrModulusTooLarge)
		require.Panics(t, func() { _ = setclass.MustNewTTO(validSet300) })

		// FromClasses errors: invalid modulus, modulus too large, invalid kind, invalid class, modulus mismatch.
		_, err = setclass.FromClasses(0, setclass.TnIn, c0Mod12)
		require.ErrorIs(t, err, setclass.ErrInvalidModulus)
		require.Panics(t, func() { _ = setclass.MustFromClasses(0, setclass.TnIn, c0Mod12) })

		_, err = setclass.FromClasses(65_537, setclass.TnIn)
		require.ErrorIs(t, err, setclass.ErrModulusTooLarge)
		_, err = setclass.FromClasses(257, setclass.TTO)
		require.ErrorIs(t, err, setclass.ErrModulusTooLarge)
		_, err = setclass.FromClasses(12, setclass.Kind(0), c0Mod12)
		require.ErrorIs(t, err, setclass.ErrInvalidKind)
		_, err = setclass.FromClasses(12, setclass.TnIn, pitchclass.Class{})
		require.ErrorIs(t, err, setclass.ErrInvalidValue)

		// Dual wrapping of ErrInvalidPitchClass and ErrModulusMismatch when c.Modulus() != modulus.
		_, err = setclass.FromClasses(12, setclass.TnIn, c0Mod12, c0Mod19)
		require.ErrorIs(t, err, setclass.ErrInvalidPitchClass)
		require.ErrorIs(t, err, setclass.ErrModulusMismatch)

		// FromNumbers and ReduceNumbers errors.
		_, err = setclass.FromNumbers(-1, setclass.TnIn, 0)
		require.ErrorIs(t, err, setclass.ErrInvalidModulus)
		_, err = setclass.FromNumbers(12, setclass.TnIn, 0, 12)
		require.ErrorIs(t, err, setclass.ErrInvalidPitchClass)
		_, err = setclass.FromNumbers(12, setclass.TnIn, -1)
		require.ErrorIs(t, err, setclass.ErrInvalidPitchClass)
		require.Panics(t, func() { _ = setclass.MustFromNumbers(12, setclass.TnIn, 12) })

		_, err = setclass.ReduceNumbers(0, setclass.TnIn, 1)
		require.ErrorIs(t, err, setclass.ErrInvalidModulus)
		_, err = setclass.ReduceNumbers(65_537, setclass.TnIn, 1)
		require.ErrorIs(t, err, setclass.ErrModulusTooLarge)
		_, err = setclass.ReduceNumbers(12, setclass.Kind(99), 1)
		require.ErrorIs(t, err, setclass.ErrInvalidKind)
		require.Panics(t, func() { _ = setclass.MustReduceNumbers(0, setclass.TnIn, 1) })

		// FromBitmask errors: N < 1, N > 64 (dual ErrOverflow + ErrModulusTooLarge), bits >= N.
		_, err = setclass.FromBitmask(0, setclass.TnIn, 0)
		require.ErrorIs(t, err, setclass.ErrInvalidModulus)
		_, err = setclass.FromBitmask(65, setclass.TnIn, 0)
		require.ErrorIs(t, err, setclass.ErrOverflow)
		require.ErrorIs(t, err, setclass.ErrModulusTooLarge)
		_, err = setclass.FromBitmask(12, setclass.TnIn, 1<<12)
		require.ErrorIs(t, err, setclass.ErrInvalidBitmask)
		require.Panics(t, func() { _ = setclass.MustFromBitmask(12, setclass.TnIn, 1<<12) })

		// FromBitmaskBig errors: nil *big.Int, negative, BitLen > N.
		_, err = setclass.FromBitmaskBig(12, setclass.TnIn, nil)
		require.ErrorIs(t, err, setclass.ErrInvalidValue)
		_, err = setclass.FromBitmaskBig(12, setclass.TnIn, big.NewInt(-1))
		require.ErrorIs(t, err, setclass.ErrInvalidBitmask)
		_, err = setclass.FromBitmaskBig(12, setclass.TnIn, big.NewInt(1<<12))
		require.ErrorIs(t, err, setclass.ErrInvalidBitmask)
		require.Panics(t, func() { _ = setclass.MustFromBitmaskBig(12, setclass.TnIn, nil) })

		// Empty, Universe, ChromaticCluster errors.
		_, err = setclass.Empty(0, setclass.TnIn)
		require.ErrorIs(t, err, setclass.ErrInvalidModulus)
		require.Panics(t, func() { _ = setclass.MustEmpty(0, setclass.TnIn) })

		_, err = setclass.Universe(0, setclass.TnIn)
		require.ErrorIs(t, err, setclass.ErrInvalidModulus)
		require.Panics(t, func() { _ = setclass.MustUniverse(0, setclass.TnIn) })

		_, err = setclass.ChromaticCluster(12, -1, setclass.TnIn)
		require.ErrorIs(t, err, setclass.ErrInvalidCardinality)
		_, err = setclass.ChromaticCluster(12, 13, setclass.TnIn)
		require.ErrorIs(t, err, setclass.ErrInvalidCardinality)
		require.Panics(t, func() { _ = setclass.MustChromaticCluster(12, 13, setclass.TnIn) })
	})
}

func TestPrimeFormsRahnForteAndCollinsMask(t *testing.T) {
	t.Parallel()

	t.Run("Mod12_all_224_TnIn_and_352_Tn_prime_forms_and_discrepancies", func(t *testing.T) {
		t.Parallel()

		var tnInDiffNames []string
		for _, fc := range catalog.ForteClasses() {
			fromRahn := setclass.MustFromNumbers(12, setclass.TnIn, fc.PrimeForm()...)
			fromForte := setclass.MustFromNumbers(12, setclass.TnIn, fc.FortePrimeForm()...)

			// Both Rahn and Forte prime forms must reduce to the exact same SetClass (G-5, K-2(4)).
			assert.Equal(t, fromRahn, fromForte, "ForteClass %s", fc.String())
			assert.True(t, fromRahn.Equal(fromForte))
			assert.Equal(t, fc.PrimeForm(), fromRahn.PrimeNumbers())
			assert.Equal(t, fc.PrimeForm(), fromRahn.PrimeForm().Numbers())
			assert.Equal(t, fc.FortePrimeForm(), fromRahn.FortePrimeForm().Numbers())
			assert.Equal(t, fc.PrimeForm(), fromRahn.MustPrimeNumbersWith(setclass.Rahn))
			assert.Equal(t, fc.FortePrimeForm(), fromRahn.MustPrimeNumbersWith(setclass.Forte))
			assert.Equal(t, fromRahn.PrimeClasses(), fromRahn.MustPrimeClassesWith(setclass.Rahn))
			assert.Equal(t, fromRahn.FortePrimeForm().Classes(), fromRahn.MustPrimeClassesWith(setclass.Forte))
			assert.Equal(t, fc.PrimeMask(), fromRahn.MustPrimeMask())
			assert.Equal(t, fc.FortePrimeMask(), fromRahn.MustFortePrimeMask())
			assert.Equal(t, fc.PrimeMask(), fromRahn.MustPrimeMaskWith(setclass.Rahn))
			assert.Equal(t, fc.FortePrimeMask(), fromRahn.MustPrimeMaskWith(setclass.Forte))
			assert.Equal(t, fc.PrimeMask(), fromRahn.PrimeMaskBig().Uint64())
			assert.Equal(t, fc.FortePrimeMask(), fromRahn.MustPrimeMaskBigWith(setclass.Forte).Uint64())
			assert.Equal(t, fc.MinSpan(), fromRahn.MinSpan())
			assert.Equal(t, fc.RahnForteDiffer(), fromRahn.RahnForteDiffer())

			if fromRahn.RahnForteDiffer() {
				tnInDiffNames = append(tnInDiffNames, fc.String())
			}
		}
		assert.Equal(t, []string{"5-20", "6-Z29", "6-31", "7-Z18", "7-20", "8-26"}, tnInDiffNames)

		var tnDiffNames []string
		var bOnlyDiffNames []string
		for _, tt := range catalog.TnTypes() {
			sc := setclass.MustFromNumbers(12, setclass.Tn, tt.PrimeForm()...)
			scForte := setclass.MustFromNumbers(12, setclass.Tn, tt.FortePrimeForm()...)

			assert.Equal(t, sc, scForte, "TnType %s", tt.String())
			assert.Equal(t, tt.PrimeForm(), sc.PrimeNumbers())
			assert.Equal(t, tt.FortePrimeForm(), sc.FortePrimeForm().Numbers())
			assert.Equal(t, tt.PrimeMask(), sc.MustPrimeMask())
			assert.Equal(t, tt.FortePrimeMask(), sc.MustFortePrimeMask())
			assert.Equal(t, tt.MinSpan(), sc.MinSpan())
			assert.Equal(t, tt.RahnForteDiffer(), sc.RahnForteDiffer())

			if sc.RahnForteDiffer() {
				tnDiffNames = append(tnDiffNames, tt.String())
				if tt.Suffix() == catalog.SuffixB && !tt.ForteClass().RahnForteDiffer() {
					bOnlyDiffNames = append(bOnlyDiffNames, tt.String())
				}
			}
		}
		assert.Len(t, tnDiffNames, 17)
		assert.Equal(t, []string{"5-32B", "6-Z44B", "8-22B", "8-27B", "9-7B", "9-8B", "9-11B"}, bOnlyDiffNames)
	})

	t.Run("Mod12_TTO_2_discrepancies_and_IntervalVectorWith", func(t *testing.T) {
		t.Parallel()

		allTTO := setclass.MustAll(12, setclass.TTO)
		require.Len(t, allTTO, 158)

		type ttoDiff struct {
			name  string
			rahn  []int
			forte []int
		}
		var diffs []ttoDiff
		for _, sc := range allTTO {
			if sc.RahnForteDiffer() {
				diffs = append(diffs, ttoDiff{
					name:  sc.MustCatalogName(),
					rahn:  sc.PrimeNumbers(),
					forte: sc.MustPrimeNumbersWith(setclass.Forte),
				})
			}
		}
		require.Equal(t, []ttoDiff{
			{
				name:  "6-Z24/Z39",
				rahn:  []int{0, 2, 3, 4, 5, 8},
				forte: []int{0, 1, 3, 4, 6, 8},
			},
			{
				name:  "8-Z15/Z29",
				rahn:  []int{0, 1, 2, 3, 5, 6, 7, 9},
				forte: []int{0, 1, 2, 3, 4, 6, 8, 9},
			},
		}, diffs)

		// For 6-Z24/Z39, Rahn prime form belongs to 6-Z39 (<3,3,3,3,2,1>) and Forte prime form to 6-Z24 (<2,3,3,3,3,1>).
		sc6Z24Z39 := setclass.MustFromCatalogName("6-Z24/Z39", setclass.TTO)
		assert.Equal(t, []int{3, 3, 3, 3, 2, 1}, sc6Z24Z39.IntervalVector().Slice())
		assert.Equal(t, []int{3, 3, 3, 3, 2, 1}, sc6Z24Z39.MustIntervalVectorWith(setclass.Rahn).Slice())
		assert.Equal(t, []int{2, 3, 3, 3, 3, 1}, sc6Z24Z39.MustIntervalVectorWith(setclass.Forte).Slice())
	})

	t.Run("RahnForteDiffer_for_small_moduli_N_1_to_10", func(t *testing.T) {
		t.Parallel()

		for n := 1; n <= 8; n++ {
			for _, kind := range setclass.Kinds() {
				for _, sc := range setclass.MustAll(n, kind) {
					assert.False(t, sc.RahnForteDiffer(), "N=%d kind=%s sc=%s", n, kind, sc)
				}
			}
		}

		countDiffer := func(n int, kind setclass.Kind) int {
			cnt := 0
			for _, sc := range setclass.MustAll(n, kind) {
				if sc.RahnForteDiffer() {
					cnt++
				}
			}
			return cnt
		}

		// G-4, K-2(3): N=9 has 1 in Tn and 0 in TnIn; N=10 has 3 in Tn and 1 in TnIn.
		assert.Equal(t, 1, countDiffer(9, setclass.Tn))
		assert.Equal(t, 0, countDiffer(9, setclass.TnIn))
		assert.Equal(t, 3, countDiffer(10, setclass.Tn))
		assert.Equal(t, 1, countDiffer(10, setclass.TnIn))
	})

	t.Run("CollinsMask_properties_and_4_10_counterexample", func(t *testing.T) {
		t.Parallel()

		sc0 := setclass.MustEmpty(12, setclass.TnIn)
		sc12 := setclass.MustUniverse(12, setclass.TnIn)
		assert.Equal(t, uint64(0), sc0.MustCollinsMask())
		assert.Equal(t, uint64(4095), sc12.MustCollinsMask())

		// Counterexample 4-10: prime form (0, 2, 3, 5) has PrimeMask == 45, whereas member {1, 8, 10, 11} gives CollinsMask == 3330.
		sc4_10 := setclass.MustFromForteNotation("4-10", setclass.TnIn)
		assert.Equal(t, []int{0, 2, 3, 5}, sc4_10.PrimeNumbers())
		assert.Equal(t, uint64(45), sc4_10.MustPrimeMask())
		assert.Equal(t, uint64(3330), sc4_10.MustCollinsMask())

		// Verify across all classes in N=12 for all three kinds that CollinsMask equals max Bitmask over Members(),
		// has k set bits, includes bit N-1 for 1 <= k < N, and is unique within (N, Kind).
		for _, kind := range setclass.Kinds() {
			seenMasks := make(map[uint64]setclass.SetClass)
			for _, sc := range setclass.MustAll(12, kind) {
				cm, err := sc.CollinsMask()
				require.NoError(t, err)
				assert.Equal(t, cm, sc.MustCollinsMask())
				assert.Equal(t, sc.Cardinality(), bits.OnesCount64(cm))
				if sc.Cardinality() >= 1 && sc.Cardinality() < 12 {
					assert.NotZero(t, cm&(uint64(1)<<11), "CollinsMask for %s must have MSB 11 set", sc)
				}

				var maxMemberMask uint64
				for _, m := range sc.MustMembers() {
					maxMemberMask = max(maxMemberMask, m.MustBitmask())
				}
				assert.Equal(t, maxMemberMask, cm, "CollinsMask mismatch with orbit max for %s", sc)

				prev, exists := seenMasks[cm]
				assert.False(t, exists, "duplicate CollinsMask %d for %s and %s", cm, prev, sc)
				seenMasks[cm] = sc
			}
		}

		// Boundary N = 64 vs N = 65 for PrimeMask, FortePrimeMask, PrimeMaskWith, CollinsMask (G-9, E-9.1, E-2.1).
		sc64 := setclass.MustFromNumbers(64, setclass.TnIn, 0, 1, 63)
		pm64, err := sc64.PrimeMask()
		require.NoError(t, err)
		assert.Equal(t, uint64(7), pm64)
		cm64, err := sc64.CollinsMask()
		require.NoError(t, err)
		assert.NotZero(t, cm64&(uint64(1)<<63))

		sc65 := setclass.MustFromNumbers(65, setclass.TnIn, 0, 1, 64)
		for _, fn := range []func() (uint64, error){
			sc65.PrimeMask,
			sc65.FortePrimeMask,
			func() (uint64, error) { return sc65.PrimeMaskWith(setclass.Rahn) },
			sc65.CollinsMask,
		} {
			_, err := fn()
			require.ErrorIs(t, err, setclass.ErrOverflow)
			require.ErrorIs(t, err, setclass.ErrModulusTooLarge)
		}
		require.Panics(t, func() { _ = sc65.MustPrimeMask() })
		require.Panics(t, func() { _ = sc65.MustFortePrimeMask() })
		require.Panics(t, func() { _ = sc65.MustPrimeMaskWith(setclass.Rahn) })
		require.Panics(t, func() { _ = sc65.MustCollinsMask() })
		assert.Equal(t, uint64(7), sc65.PrimeMaskBig().Uint64())
	})

	t.Run("Compare_and_CompareWith_ordering", func(t *testing.T) {
		t.Parallel()

		scMod7 := setclass.MustFromNumbers(7, setclass.TnIn, 0, 2, 4)
		scTn := setclass.MustFromNumbers(12, setclass.Tn, 0, 1, 2)
		scTnIn := setclass.MustFromNumbers(12, setclass.TnIn, 0, 1, 2)
		scCard4 := setclass.MustFromNumbers(12, setclass.TnIn, 0, 1, 2, 3)
		scTTO := setclass.MustFromNumbers(12, setclass.TTO, 0, 1, 2)

		ordered := []setclass.SetClass{
			{},
			scMod7,
			scTn,
			scTnIn,
			scCard4,
			scTTO,
		}
		for i := range ordered {
			assert.Equal(t, 0, ordered[i].Compare(ordered[i]))
			for j := i + 1; j < len(ordered); j++ {
				assert.Equal(t, -1, ordered[i].Compare(ordered[j]), "i=%d j=%d", i, j)
				assert.Equal(t, 1, ordered[j].Compare(ordered[i]), "i=%d j=%d", i, j)
			}
		}

		// 5-19 has prime form [0, 1, 3, 6, 7] under both Rahn and Forte;
		// 5-20 has Rahn prime form [0, 1, 5, 6, 8] (> 5-19) and Forte prime form [0, 1, 3, 7, 8] (> 5-19),
		// while 5-Z18 has prime form [0, 1, 4, 5, 7]: under Rahn 5-Z18 < 5-20, whereas under Forte 5-20 ([0,1,3,7,8]) < 5-Z18 ([0,1,4,5,7])!
		sc5_19 := setclass.MustFromForteNotation("5-19", setclass.TnIn)
		sc5_20 := setclass.MustFromForteNotation("5-20", setclass.TnIn)
		sc5_Z18 := setclass.MustFromForteNotation("5-Z18", setclass.TnIn)

		assert.Equal(t, -1, sc5_19.Compare(sc5_20))
		assert.Equal(t, -1, sc5_Z18.Compare(sc5_20))
		assert.Equal(t, -1, sc5_Z18.MustCompareWith(sc5_20, setclass.Rahn))
		assert.Equal(t, 1, sc5_Z18.MustCompareWith(sc5_20, setclass.Forte))
		assert.Equal(t, 0, sc5_20.MustCompareWith(sc5_20, setclass.Forte))

		// Invalid algorithm error in CompareWith and Prime*With.
		_, err := sc5_19.CompareWith(sc5_20, setclass.Algorithm(99))
		require.ErrorIs(t, err, setclass.ErrInvalidAlgorithm)
		_, err = sc5_19.PrimeFormWith(setclass.Algorithm(0))
		require.ErrorIs(t, err, setclass.ErrInvalidAlgorithm)
		_, err = sc5_19.PrimeClassesWith(setclass.Algorithm(0))
		require.ErrorIs(t, err, setclass.ErrInvalidAlgorithm)
		_, err = sc5_19.PrimeNumbersWith(setclass.Algorithm(0))
		require.ErrorIs(t, err, setclass.ErrInvalidAlgorithm)
		_, err = sc5_19.PrimeMaskBigWith(setclass.Algorithm(0))
		require.ErrorIs(t, err, setclass.ErrInvalidAlgorithm)
		_, err = sc5_19.PrimeMaskWith(setclass.Algorithm(0))
		require.ErrorIs(t, err, setclass.ErrInvalidAlgorithm)
	})
}

func TestMembershipEquivalenceOrbitAndSymmetry(t *testing.T) {
	t.Parallel()

	t.Run("Contains_and_AreEquivalent_across_Tn_TnIn_TTO", func(t *testing.T) {
		t.Parallel()

		cMaj := pcset.MustNew(12, 0, 4, 7)
		cMin := pcset.MustNew(12, 0, 3, 7)
		chromatic3 := pcset.MustNew(12, 0, 1, 2)
		quartal3 := pcset.MustNew(12, 0, 2, 7)
		setMod19 := pcset.MustNew(19, 0, 4, 7)

		tn3_11A := setclass.MustFromNumbers(12, setclass.Tn, 0, 3, 7)
		tnIn3_11 := setclass.MustFromNumbers(12, setclass.TnIn, 0, 3, 7)
		tto3_1_9 := setclass.MustFromNumbers(12, setclass.TTO, 0, 1, 2)

		assert.True(t, tn3_11A.Contains(cMin))
		assert.False(t, tn3_11A.Contains(cMaj))
		assert.True(t, tnIn3_11.Contains(cMin))
		assert.True(t, tnIn3_11.Contains(cMaj))
		assert.False(t, tnIn3_11.Contains(chromatic3))
		assert.False(t, tnIn3_11.Contains(setMod19))
		assert.True(t, tto3_1_9.Contains(chromatic3))
		assert.True(t, tto3_1_9.Contains(quartal3))

		// K-4(1): {0, 4, 7} and {0, 3, 7} are not equivalent in Tn, but equivalent in TnIn and TTO.
		eqTn, err := setclass.AreEquivalentTn(cMaj, cMin)
		require.NoError(t, err)
		assert.False(t, eqTn)
		assert.False(t, setclass.MustAreEquivalent(cMaj, cMin, setclass.Tn))
		assert.False(t, setclass.MustAreEquivalentTn(cMaj, cMin))

		eqTnIn, err := setclass.AreEquivalentTnIn(cMaj, cMin)
		require.NoError(t, err)
		assert.True(t, eqTnIn)
		assert.True(t, setclass.MustAreEquivalent(cMaj, cMin, setclass.TnIn))
		assert.True(t, setclass.MustAreEquivalentTnIn(cMaj, cMin))

		eqTTO, err := setclass.AreEquivalentTTO(cMaj, cMin)
		require.NoError(t, err)
		assert.True(t, eqTTO)
		assert.True(t, setclass.MustAreEquivalent(cMaj, cMin, setclass.TTO))
		assert.True(t, setclass.MustAreEquivalentTTO(cMaj, cMin))

		// K-4(2): {0, 1, 2} (3-1) and {0, 2, 7} (3-9) are equivalent only in TTO.
		assert.False(t, setclass.MustAreEquivalentTn(chromatic3, quartal3))
		assert.False(t, setclass.MustAreEquivalentTnIn(chromatic3, quartal3))
		assert.True(t, setclass.MustAreEquivalentTTO(chromatic3, quartal3))

		// Error cases for AreEquivalent*.
		_, err = setclass.AreEquivalent(pcset.Set{}, cMaj, setclass.TnIn)
		require.ErrorIs(t, err, setclass.ErrInvalidValue)
		_, err = setclass.AreEquivalent(cMaj, cMin, setclass.Kind(99))
		require.ErrorIs(t, err, setclass.ErrInvalidKind)
		_, err = setclass.AreEquivalent(cMaj, setMod19, setclass.TnIn)
		require.ErrorIs(t, err, setclass.ErrModulusMismatch)
		_, err = setclass.AreEquivalentTTO(pcset.MustNew(300, 0, 1), pcset.MustNew(300, 0, 1))
		require.ErrorIs(t, err, setclass.ErrModulusTooLarge)

		require.Panics(t, func() { _ = setclass.MustAreEquivalent(cMaj, setMod19, setclass.TnIn) })
		require.Panics(t, func() { _ = setclass.MustAreEquivalentTn(cMaj, setMod19) })
		require.Panics(t, func() { _ = setclass.MustAreEquivalentTnIn(cMaj, setMod19) })
		require.Panics(t, func() { _ = setclass.MustAreEquivalentTTO(cMaj, setMod19) })
	})

	t.Run("OrbitSize_Members_Symmetries_and_LinkingOperators", func(t *testing.T) {
		t.Parallel()

		// K-5(2): 0-1 and 12-1 have orbit size 1 in all 3 groups;
		// 3-12 ({0, 4, 8}) has DS_T=3, DS_I=3, DS_TI=6, DS_TTO=12 and OrbitSize=4 in all 3 groups;
		// 3-11 ({0, 3, 7}) has DS_T=1, DS_I=0, DS_TI=1, DS_TTO=1 and OrbitSize 12 (Tn), 24 (TnIn), 48 (TTO).
		for _, kind := range setclass.Kinds() {
			assert.Equal(t, 1, setclass.MustEmpty(12, kind).OrbitSize())
			assert.Equal(t, 1, setclass.MustUniverse(12, kind).OrbitSize())

			aug := setclass.MustFromNumbers(12, kind, 0, 4, 8)
			assert.Equal(t, 3, aug.TnSymmetry())
			assert.Equal(t, 3, aug.InSymmetry())
			assert.Equal(t, 6, aug.SymmetryDegreeTI())
			assert.Equal(t, 12, aug.MustSymmetryDegreeTTO())
			assert.Equal(t, 4, aug.OrbitSize())
			assert.True(t, aug.IsTranspositionallySymmetric())
			assert.True(t, aug.IsInversionallySymmetric())
			assert.False(t, aug.IsChiral())
		}

		tn3_11A := setclass.MustFromNumbers(12, setclass.Tn, 0, 3, 7)
		tnIn3_11 := setclass.MustFromNumbers(12, setclass.TnIn, 0, 3, 7)
		tto3_3_11 := setclass.MustFromNumbers(12, setclass.TTO, 0, 3, 7)

		assert.Equal(t, 12, tn3_11A.OrbitSize())
		assert.Equal(t, 24, tnIn3_11.OrbitSize())
		assert.Equal(t, 48, tto3_3_11.OrbitSize())
		assert.False(t, tnIn3_11.IsTranspositionallySymmetric())
		assert.False(t, tnIn3_11.IsInversionallySymmetric())
		assert.True(t, tnIn3_11.IsChiral())

		// Verify sum(OrbitSize) == 4096 and Members() invariants across all classes in N=12 for all 3 kinds.
		for _, kind := range setclass.Kinds() {
			totalSets := 0
			for _, sc := range setclass.MustAll(12, kind) {
				members, err := sc.Members()
				require.NoError(t, err)
				assert.Equal(t, members, sc.MustMembers())
				require.Len(t, members, sc.OrbitSize())
				totalSets += len(members)

				for idx, m := range members {
					assert.True(t, sc.Contains(m))
					if idx > 0 {
						assert.Equal(t, -1, members[idx-1].Compare(m))
					}
				}
			}
			assert.Equal(t, 4096, totalSets, "total sets in orbits for kind %s", kind)
		}

		// K-5(3): Linking indices and operators for 3-11 ({0, 3, 7}) and target C major {0, 4, 7} (I_7({0, 3, 7}) = {0, 4, 7}).
		cMaj := pcset.MustNew(12, 0, 4, 7)
		dMin := pcset.MustNew(12, 2, 5, 9) // T_2({0, 3, 7})
		aug := pcset.MustNew(12, 0, 4, 8)  // not in 3-11

		tIdx, err := tnIn3_11.TranspositionIndicesTo(dMin)
		require.NoError(t, err)
		assert.Equal(t, []int{2}, tIdx.Numbers())
		assert.Equal(t, tIdx, tnIn3_11.MustTranspositionIndicesTo(dMin))
		assert.Equal(t, tIdx, tnIn3_11.MustTranspositionIndicesToWith(dMin, setclass.Rahn))

		iIdx, err := tnIn3_11.InversionIndicesTo(cMaj)
		require.NoError(t, err)
		assert.Equal(t, []int{7}, iIdx.Numbers())
		assert.Equal(t, iIdx, tnIn3_11.MustInversionIndicesTo(cMaj))
		assert.Equal(t, iIdx, tnIn3_11.MustInversionIndicesToWith(cMaj, setclass.Rahn))

		// For Tn class 3-11A, InversionIndicesTo returns empty set because Tn group has no inversions.
		assert.True(t, tn3_11A.MustInversionIndicesTo(cMaj).IsEmpty())

		// Non-member target returns empty set / empty operator slice without error.
		assert.True(t, tnIn3_11.MustTranspositionIndicesTo(aug).IsEmpty())
		assert.Empty(t, tnIn3_11.MustOperatorsTo(aug))
		assert.Empty(t, tnIn3_11.MustOperatorsFrom(aug))
		assert.Empty(t, setclass.MustFindSetOperators(dMin, aug, setclass.TnIn))

		// Check OperatorsTo and OperatorsFrom on all members of 3-12 ({0, 4, 8}) across all 3 kinds.
		for _, kind := range setclass.Kinds() {
			scAug := setclass.MustFromNumbers(12, kind, 0, 4, 8)
			expectedStab := scAug.TnSymmetry()
			switch kind {
			case setclass.TnIn:
				expectedStab = scAug.SymmetryDegreeTI()
			case setclass.TTO:
				expectedStab = scAug.MustSymmetryDegreeTTO()
			}

			for _, member := range scAug.MustMembers() {
				opsTo, err := scAug.OperatorsTo(member)
				require.NoError(t, err)
				assert.Equal(t, opsTo, scAug.MustOperatorsTo(member))
				assert.Equal(t, opsTo, scAug.MustOperatorsToWith(member, setclass.Rahn))
				assert.Len(t, opsTo, expectedStab)
				for _, op := range opsTo {
					assert.Equal(t, member, scAug.PrimeForm().MustApply(op))
				}

				opsFrom, err := scAug.OperatorsFrom(member)
				require.NoError(t, err)
				assert.Equal(t, opsFrom, scAug.MustOperatorsFrom(member))
				assert.Equal(t, opsFrom, scAug.MustOperatorsFromWith(member, setclass.Rahn))
				assert.Len(t, opsFrom, expectedStab)
				for _, op := range opsFrom {
					assert.Equal(t, scAug.PrimeForm(), member.MustApply(op))
				}

				findOps, err := setclass.FindSetOperators(scAug.PrimeForm(), member, kind)
				require.NoError(t, err)
				assert.Equal(t, opsTo, findOps)
				assert.Equal(t, opsTo, setclass.MustFindSetOperators(scAug.PrimeForm(), member, kind))
			}
		}

		// Error cases on linking methods: modulus mismatch, invalid set, invalid algorithm, modulus > 1024.
		set19 := pcset.MustNew(19, 0, 3, 7)
		_, err = tnIn3_11.TranspositionIndicesTo(set19)
		require.ErrorIs(t, err, setclass.ErrModulusMismatch)
		_, err = tnIn3_11.InversionIndicesTo(set19)
		require.ErrorIs(t, err, setclass.ErrModulusMismatch)
		_, err = tnIn3_11.OperatorsTo(set19)
		require.ErrorIs(t, err, setclass.ErrModulusMismatch)
		_, err = tnIn3_11.OperatorsFrom(set19)
		require.ErrorIs(t, err, setclass.ErrModulusMismatch)
		_, err = setclass.FindSetOperators(cMaj, set19, setclass.TnIn)
		require.ErrorIs(t, err, setclass.ErrModulusMismatch)

		_, err = tnIn3_11.OperatorsToWith(cMaj, setclass.Algorithm(99))
		require.ErrorIs(t, err, setclass.ErrInvalidAlgorithm)
		_, err = tnIn3_11.OperatorsFromWith(cMaj, setclass.Algorithm(99))
		require.ErrorIs(t, err, setclass.ErrInvalidAlgorithm)
		_, err = tnIn3_11.TranspositionIndicesToWith(cMaj, setclass.Algorithm(99))
		require.ErrorIs(t, err, setclass.ErrInvalidAlgorithm)
		_, err = tnIn3_11.InversionIndicesToWith(cMaj, setclass.Algorithm(99))
		require.ErrorIs(t, err, setclass.ErrInvalidAlgorithm)

		sc1025 := setclass.MustFromNumbers(1_025, setclass.TnIn, 0, 1, 3)
		set1025 := pcset.MustNew(1_025, 0, 1, 3)
		_, err = sc1025.Members()
		require.ErrorIs(t, err, setclass.ErrModulusTooLarge)
		_, err = sc1025.OperatorsTo(set1025)
		require.ErrorIs(t, err, setclass.ErrModulusTooLarge)
		_, err = sc1025.SymmetryDegreeTTO()
		require.ErrorIs(t, err, setclass.ErrModulusTooLarge)
	})

	t.Run("Degenerate_moduli_N_1_and_N_2_and_chirality_emergence_at_N_6", func(t *testing.T) {
		t.Parallel()

		// G-1: N = 1 (2 classes in each group) and N = 2 (3 classes in each group).
		for _, kind := range setclass.Kinds() {
			all1 := setclass.MustAll(1, kind)
			require.Len(t, all1, 2)
			for _, sc := range all1 {
				assert.Equal(t, 1, sc.TnSymmetry())
				assert.Equal(t, 1, sc.InSymmetry())
				assert.Equal(t, 1, sc.SymmetryDegreeTI())
				assert.Equal(t, 1, sc.MustSymmetryDegreeTTO())
				assert.Equal(t, 1, sc.OrbitSize())
				assert.True(t, sc.IsInversionallySymmetric())
				assert.False(t, sc.IsChiral())
				assert.Equal(t, 0, sc.IntervalVector().Len())
			}

			all2 := setclass.MustAll(2, kind)
			require.Len(t, all2, 3)
			for _, sc := range all2 {
				assert.Equal(t, sc.TnSymmetry(), sc.SymmetryDegreeTI())
				assert.Equal(t, sc.TnSymmetry(), sc.MustSymmetryDegreeTTO())
				assert.True(t, sc.IsInversionallySymmetric())
				assert.False(t, sc.IsChiral())
			}
			assert.Equal(t, 1, all2[0].OrbitSize())
			assert.Equal(t, 2, all2[1].OrbitSize())
			assert.Equal(t, 1, all2[2].OrbitSize())
			assert.False(t, all2[0].IsSelfComplementary())
			assert.True(t, all2[1].IsSelfComplementary())
			assert.False(t, all2[2].IsSelfComplementary())
		}

		// G-3: No chiral classes for N <= 5; at N = 6, exactly one chiral pair in k = 3: (0, 1, 3) and (0, 2, 3).
		for n := 1; n <= 5; n++ {
			for _, sc := range setclass.MustAll(n, setclass.TnIn) {
				assert.False(t, sc.IsChiral(), "N=%d sc=%s", n, sc)
			}
		}
		var chiralN6 []setclass.SetClass
		for _, sc := range setclass.MustAll(6, setclass.TnIn) {
			if sc.IsChiral() {
				chiralN6 = append(chiralN6, sc)
			}
		}
		require.Len(t, chiralN6, 1)
		assert.Equal(t, []int{0, 1, 3}, chiralN6[0].PrimeNumbers())
		tnSub := chiralN6[0].TnClasses()
		require.Len(t, tnSub, 2)
		assert.Equal(t, []int{0, 1, 3}, tnSub[0].PrimeNumbers())
		assert.Equal(t, []int{0, 2, 3}, tnSub[1].PrimeNumbers())
	})

	t.Run("Mod12_structural_invariants_and_Morris_InvarianceVector", func(t *testing.T) {
		t.Parallel()

		// K-6(1): Verify all 224 TnIn classes and 352 Tn classes match setclass/catalog.
		for _, fc := range catalog.ForteClasses() {
			sc := setclass.MustFromForteClass(fc)
			assert.Equal(t, fc.IntervalVectorSlice(), sc.IntervalVector().Slice())
			assert.Equal(t, fc.InvarianceVectorSlice(), sc.MustInvarianceVector().Slice())
			assert.Equal(t, fc.TnSymmetry(), sc.TnSymmetry())
			assert.Equal(t, fc.InSymmetry(), sc.InSymmetry())
			assert.Equal(t, fc.DegreeOfSymmetry(), sc.SymmetryDegreeTI())
			assert.Equal(t, fc.OrbitSize(), sc.OrbitSize())
			assert.Equal(t, fc.IsInversionallySymmetric(), sc.IsInversionallySymmetric())
			assert.Equal(t, fc.IsChiral(), sc.IsChiral())
		}

		// K-6(2): For all 158 TTO classes at N = 12, InvarianceVector is identical across all constituent
		// TnIn and Tn subclasses, and the sum of its first 4 components equals SymmetryDegreeTTO.
		for _, tto := range setclass.MustAll(12, setclass.TTO) {
			iv, err := tto.InvarianceVector()
			require.NoError(t, err)
			assert.Equal(t, iv, tto.MustInvarianceVector())
			s := iv.Slice()
			require.Len(t, s, 8)
			assert.Equal(t, tto.MustSymmetryDegreeTTO(), s[0]+s[1]+s[2]+s[3])

			for _, subTnIn := range tto.MustTnInClasses() {
				assert.Equal(t, iv, subTnIn.MustInvarianceVector())
			}
			for _, subTn := range tto.TnClasses() {
				assert.Equal(t, iv, subTn.MustInvarianceVector())
			}
		}

		// K-6(3): InvarianceVector returns ErrNotModulus12 when N != 12.
		for _, n := range []int{1, 7, 19, 24} {
			sc := setclass.MustEmpty(n, setclass.TnIn)
			_, err := sc.InvarianceVector()
			require.ErrorIs(t, err, setclass.ErrNotModulus12)
			require.Panics(t, func() { _ = sc.MustInvarianceVector() })
		}
	})
}

func TestLevelTransitionsInversionMultiplyAndComplement(t *testing.T) {
	t.Parallel()

	t.Run("ToTnIn_ToTTO_TnClasses_TnInClasses_and_Rahn_ordering_at_N_14", func(t *testing.T) {
		t.Parallel()

		// K-7(1, 2): Partition counts at N = 12.
		allTnIn := setclass.MustAll(12, setclass.TnIn)
		require.Len(t, allTnIn, 224)

		totalTnFromTnIn := 0
		achiralCount := 0
		chiralCount := 0
		for _, sc := range allTnIn {
			assert.Equal(t, sc, sc.ToTnIn())
			tnInSubs, err := sc.TnInClasses()
			require.NoError(t, err)
			assert.Equal(t, []setclass.SetClass{sc}, tnInSubs)

			tnSubs := sc.TnClasses()
			totalTnFromTnIn += len(tnSubs)
			if sc.IsChiral() {
				chiralCount++
				require.Len(t, tnSubs, 2)
				assert.Equal(t, sc.PrimeForm(), tnSubs[0].PrimeForm())
				assert.Equal(t, tnSubs[1], tnSubs[0].Inversion())
				assert.Equal(t, tnSubs[0], tnSubs[1].Inversion())
				assert.Equal(t, sc, tnSubs[0].ToTnIn())
				assert.Equal(t, sc, tnSubs[1].ToTnIn())
				assert.Equal(t, catalog.SuffixA, tnSubs[0].MustTnType().Suffix())
				assert.Equal(t, catalog.SuffixB, tnSubs[1].MustTnType().Suffix())
			} else {
				achiralCount++
				require.Len(t, tnSubs, 1)
				assert.Equal(t, tnSubs[0], tnSubs[0].Inversion())
			}
		}
		assert.Equal(t, 96, achiralCount)
		assert.Equal(t, 128, chiralCount)
		assert.Equal(t, 352, totalTnFromTnIn)

		allTTO := setclass.MustAll(12, setclass.TTO)
		require.Len(t, allTTO, 158)
		totalTnInFromTTO := 0
		totalTnFromTTO := 0
		m5InvCount := 0
		m5PairCount := 0
		for _, tto := range allTTO {
			assert.Equal(t, tto, tto.MustToTTO())
			tnInSubs, err := tto.TnInClasses()
			require.NoError(t, err)
			assert.Equal(t, tnInSubs, tto.MustTnInClasses())
			totalTnInFromTTO += len(tnInSubs)
			if len(tnInSubs) == 1 {
				m5InvCount++
			} else {
				m5PairCount++
				require.Len(t, tnInSubs, 2)
			}
			for _, sub := range tnInSubs {
				assert.Equal(t, tto, sub.MustToTTO())
			}

			tnSubs := tto.TnClasses()
			totalTnFromTTO += len(tnSubs)
			for _, sub := range tnSubs {
				assert.Equal(t, tto, sub.MustToTTO())
			}
		}
		assert.Equal(t, 92, m5InvCount)
		assert.Equal(t, 66, m5PairCount)
		assert.Equal(t, 224, totalTnInFromTTO)
		assert.Equal(t, 352, totalTnFromTTO)

		// Crucially verify Rahn packing order for TnClasses() at N = 14, k = 8:
		// P1 = (0, 1, 3, 4, 5, 8, 9, 11) has smaller Rahn key (right-to-left: index 6 is 9 < 10)
		// than its inversion P2 = (0, 1, 2, 4, 5, 8, 10, 11), even though left-to-right P2 < P1!
		// Therefore TnClasses() on the parent TnIn class MUST put [P1]_Tn first and [P2]_Tn second.
		sc14 := setclass.MustFromNumbers(14, setclass.TnIn, 0, 1, 3, 4, 5, 8, 9, 11)
		assert.Equal(t, []int{0, 1, 3, 4, 5, 8, 9, 11}, sc14.PrimeNumbers())
		tnSubs14 := sc14.TnClasses()
		require.Len(t, tnSubs14, 2)
		assert.Equal(t, []int{0, 1, 3, 4, 5, 8, 9, 11}, tnSubs14[0].PrimeNumbers())
		assert.Equal(t, []int{0, 1, 2, 4, 5, 8, 10, 11}, tnSubs14[1].PrimeNumbers())
		assert.Equal(t, 1, tnSubs14[0].Compare(tnSubs14[1]), "left-to-right Compare has P1 > P2, but Rahn key puts P1 first in TnClasses()")
	})

	t.Run("Multiply_M5Partner_and_IsM5Invariant", func(t *testing.T) {
		t.Parallel()

		// K-7(3): For invertible m, sc.Multiply(m).ToTTO() == sc.ToTTO().
		for _, kind := range setclass.Kinds() {
			for _, sc := range setclass.MustAll(12, kind) {
				for _, m := range []int{1, 5, 7, 11, -1, -5, 17} {
					mul, err := sc.Multiply(m)
					require.NoError(t, err)
					assert.Equal(t, mul, sc.MustMultiply(m))
					assert.Equal(t, sc.MustToTTO(), mul.MustToTTO())
				}
			}
		}

		// Non-invertible multipliers return ErrNonInvertibleMultiplier (E-4.3).
		sc3_11 := setclass.MustFromForteNotation("3-11", setclass.TnIn)
		for _, badM := range []int{0, 2, 3, 4, 6, 8, 9, 10, 12, -2} {
			_, err := sc3_11.Multiply(badM)
			require.ErrorIs(t, err, setclass.ErrNonInvertibleMultiplier, "m=%d", badM)
			require.Panics(t, func() { _ = sc3_11.MustMultiply(badM) })
		}

		// M5Partner and IsM5Invariant across all 3 kinds at N = 12.
		m5InvTnIn := 0
		for _, sc := range setclass.MustAll(12, setclass.TnIn) {
			p, err := sc.M5Partner()
			require.NoError(t, err)
			assert.Equal(t, p, sc.MustM5Partner())
			assert.Equal(t, sc.MustMultiply(5), p)

			inv, err := sc.IsM5Invariant()
			require.NoError(t, err)
			assert.Equal(t, inv, sc.MustIsM5Invariant())
			assert.Equal(t, p.Equal(sc), inv)
			if inv {
				m5InvTnIn++
			}
		}
		assert.Equal(t, 92, m5InvTnIn)

		for _, sc := range setclass.MustAll(12, setclass.TTO) {
			assert.Equal(t, sc, sc.MustM5Partner())
			assert.True(t, sc.MustIsM5Invariant())
		}

		for _, sc := range setclass.MustAll(12, setclass.Tn) {
			p := sc.MustM5Partner()
			assert.Equal(t, sc.MustMultiply(5), p)
			assert.Equal(t, p.Equal(sc), sc.MustIsM5Invariant())
		}

		// ErrNotModulus12 for M5Partner and IsM5Invariant when N != 12.
		sc19 := setclass.MustFromNumbers(19, setclass.TnIn, 0, 3, 7)
		_, err := sc19.M5Partner()
		require.ErrorIs(t, err, setclass.ErrNotModulus12)
		_, err = sc19.IsM5Invariant()
		require.ErrorIs(t, err, setclass.ErrNotModulus12)
		require.Panics(t, func() { _ = sc19.MustM5Partner() })
		require.Panics(t, func() { _ = sc19.MustIsM5Invariant() })
	})

	t.Run("Complement_and_IsSelfComplementary_across_Tn_TnIn_TTO", func(t *testing.T) {
		t.Parallel()

		// K-8(1, 2): Involution and exact self-complementary hexachord counts at N = 12:
		// 8 in Tn (6 Babbitt all-combinatorial + 6-14A/6-14B), 20 in TnIn, 18 in TTO.
		var selfCompTn []string
		for _, sc := range setclass.MustOfCardinality(12, 6, setclass.Tn) {
			assert.Equal(t, sc, sc.Complement().Complement())
			if sc.IsSelfComplementary() {
				selfCompTn = append(selfCompTn, sc.MustCatalogName())
			}
		}
		slices.Sort(selfCompTn)
		assert.Equal(t, []string{"6-1", "6-14A", "6-14B", "6-20", "6-32", "6-35", "6-7", "6-8"}, selfCompTn)

		selfCompTnIn := 0
		for _, sc := range setclass.MustOfCardinality(12, 6, setclass.TnIn) {
			assert.Equal(t, sc, sc.Complement().Complement())
			if sc.IsSelfComplementary() {
				selfCompTnIn++
				assert.False(t, sc.MustForteClass().IsZ())
			} else {
				zp, ok := sc.MustZPartner()
				require.True(t, ok)
				assert.Equal(t, zp, sc.Complement())
			}
		}
		assert.Equal(t, 20, selfCompTnIn)

		var selfCompTTO []string
		for _, sc := range setclass.MustOfCardinality(12, 6, setclass.TTO) {
			assert.Equal(t, sc, sc.Complement().Complement())
			if sc.IsSelfComplementary() {
				selfCompTTO = append(selfCompTTO, sc.MustCatalogName())
			}
		}
		assert.Len(t, selfCompTTO, 18)
		for _, zSelfTTO := range []string{"6-Z6/Z38", "6-Z11/Z40", "6-Z19/Z44", "6-1/32", "6-20", "6-35"} {
			assert.Contains(t, selfCompTTO, zSelfTTO)
		}

		// K-8(3): 13 chiral classes preserving suffix A->A and B->B under directed Tn complement,
		// and 115 chiral classes swapping A<->B.
		preserveSuffixExpected := []string{
			"4-12", "4-14", "5-11", "5-26", "5-28",
			"6-14", "6-Z10", "6-Z39",
			"7-11", "7-26", "7-28", "8-12", "8-14",
		}
		var preserveActual []string
		swapCount := 0
		for _, fc := range catalog.ForteClasses() {
			if !fc.IsChiral() {
				continue
			}
			sc := setclass.MustFromForteClass(fc)
			tnSubs := sc.TnClasses()
			require.Len(t, tnSubs, 2)
			compA := tnSubs[0].Complement().MustTnType()
			compB := tnSubs[1].Complement().MustTnType()
			if compA.Suffix() == catalog.SuffixA && compB.Suffix() == catalog.SuffixB {
				preserveActual = append(preserveActual, fc.String())
			} else {
				assert.Equal(t, catalog.SuffixB, compA.Suffix())
				assert.Equal(t, catalog.SuffixA, compB.Suffix())
				swapCount++
			}
		}
		slices.Sort(preserveActual)
		assert.Equal(t, preserveSuffixExpected, preserveActual)
		assert.Equal(t, 115, swapCount)
	})
}

func TestZRelationsAndZPartners(t *testing.T) {
	t.Parallel()

	t.Run("Mod12_Z_relations_across_TnIn_Tn_TTO", func(t *testing.T) {
		t.Parallel()

		// K-9(1): Empty (0-1) and singleton (1-1), as well as chiral branches of the same parent class
		// (3-11A vs 3-11B, 4-Z15A vs 4-Z15B), are NOT Z-related.
		sc0 := setclass.MustEmpty(12, setclass.TnIn)
		sc1 := setclass.MustFromNumbers(12, setclass.TnIn, 0)
		assert.False(t, sc0.MustIsZRelated(sc1))
		assert.False(t, sc0.MustHasZPartner())
		assert.False(t, sc1.MustHasZPartner())

		tn3_11A := setclass.MustFromTnTypeNotation("3-11A")
		tn3_11B := setclass.MustFromTnTypeNotation("3-11B")
		assert.False(t, tn3_11A.MustIsZRelated(tn3_11B))

		tn4_Z15A := setclass.MustFromTnTypeNotation("4-Z15A")
		tn4_Z15B := setclass.MustFromTnTypeNotation("4-Z15B")
		tn4_Z29A := setclass.MustFromTnTypeNotation("4-Z29A")
		tn4_Z29B := setclass.MustFromTnTypeNotation("4-Z29B")
		assert.False(t, tn4_Z15A.MustIsZRelated(tn4_Z15B))
		assert.True(t, tn4_Z15A.MustIsZRelated(tn4_Z29A))
		assert.True(t, tn4_Z15A.MustIsZRelated(tn4_Z29B))
		assert.Equal(t, []setclass.SetClass{tn4_Z29A, tn4_Z29B}, tn4_Z15A.MustZPartners())

		// 5-Z36A has a single Tn Z-partner [5-Z12] because 5-Z12 is achiral.
		tn5_Z36A := setclass.MustFromTnTypeNotation("5-Z36A")
		tn5_Z12 := setclass.MustFromTnTypeNotation("5-Z12")
		assert.Equal(t, []setclass.SetClass{tn5_Z12}, tn5_Z36A.MustZPartners())

		// K-9(2): Counts of HasZPartner() == true at N = 12: 46 in TnIn, 88 in Tn, 20 in TTO.
		countZ := func(kind setclass.Kind) int {
			cnt := 0
			for _, sc := range setclass.MustAll(12, kind) {
				has, err := sc.HasZPartner()
				require.NoError(t, err)
				assert.Equal(t, has, sc.MustHasZPartner())
				partners, err := sc.ZPartners()
				require.NoError(t, err)
				assert.Equal(t, partners, sc.MustZPartners())
				assert.Equal(t, len(partners) > 0, has)
				for _, p := range partners {
					assert.True(t, sc.MustIsZRelated(p))
				}
				if has {
					cnt++
				}
			}
			return cnt
		}
		assert.Equal(t, 46, countZ(setclass.TnIn))
		assert.Equal(t, 72, countZ(setclass.Tn))
		assert.Equal(t, 20, countZ(setclass.TTO))

		// Verify the 9 composite TTO classes where Z-partners merged under M5 have HasZPartner() == false,
		// while ZPair() returns true!
		mergedTTOZ := []string{
			"4-Z15/Z29",
			"5-Z17/Z37",
			"5-Z18/Z38",
			"6-Z6/Z38",
			"6-Z11/Z40",
			"6-Z19/Z44",
			"7-Z17/Z37",
			"7-Z18/Z38",
			"8-Z15/Z29",
		}
		for _, name := range mergedTTOZ {
			sc := setclass.MustFromCatalogName(name, setclass.TTO)
			assert.False(t, sc.MustHasZPartner(), "merged TTO Z-pair %s must have HasZPartner()==false", name)
			assert.Empty(t, sc.MustZPartners())
			_, hasCatalogZPair := sc.MustZPair()
			assert.True(t, hasCatalogZPair, "merged TTO Z-pair %s must still report catalog ZPair", name)
		}

		// ZPartner() is only valid for Kind == TnIn at N = 12; for Tn and TTO it returns ErrKindMismatch (E-6.2).
		_, _, err := tn4_Z15A.ZPartner()
		require.ErrorIs(t, err, setclass.ErrKindMismatch)
		require.Panics(t, func() { _, _ = tn4_Z15A.MustZPartner() })

		tto4Z := setclass.MustFromCatalogName("4-Z15/Z29", setclass.TTO)
		_, _, err = tto4Z.ZPartner()
		require.ErrorIs(t, err, setclass.ErrKindMismatch)

		// AreZRelatedSets at N = 12 and at N = 65_536.
		s4Z15 := pcset.MustNew(12, 0, 1, 4, 6)
		s4Z29 := pcset.MustNew(12, 0, 1, 3, 7)
		s3_11 := pcset.MustNew(12, 0, 3, 7)
		zr, err := setclass.AreZRelatedSets(s4Z15, s4Z29)
		require.NoError(t, err)
		assert.True(t, zr)
		assert.True(t, setclass.MustAreZRelatedSets(s4Z15, s4Z29))
		assert.False(t, setclass.MustAreZRelatedSets(s4Z15, s3_11))
	})

	t.Run("Non_mod12_Z_relations_N_8_N_10_and_N_16_multiples", func(t *testing.T) {
		t.Parallel()

		// G-4, K-9(3): At N = 8, exactly 1 Z-pair in TnIn: (0, 1, 2, 5) and (0, 1, 3, 4) with ICV <2, 1, 2, 1>.
		var z8 []setclass.SetClass
		for _, sc := range setclass.MustAll(8, setclass.TnIn) {
			if sc.MustHasZPartner() {
				z8 = append(z8, sc)
			}
		}
		require.Len(t, z8, 2)
		assert.Equal(t, []int{0, 1, 2, 5}, z8[0].PrimeNumbers())
		assert.Equal(t, []int{0, 1, 3, 4}, z8[1].PrimeNumbers())
		assert.Equal(t, []int{2, 1, 2, 1}, z8[0].IntervalVector().Slice())
		assert.True(t, z8[0].MustIsZRelated(z8[1]))

		// At N = 10, exactly 3 Z-pairs (6 classes) in TnIn.
		z10Count := 0
		for _, sc := range setclass.MustAll(10, setclass.TnIn) {
			if sc.MustHasZPartner() {
				z10Count++
			}
		}
		assert.Equal(t, 6, z10Count)

		// G-7, K-9(3): At N = 16 in TnIn (2250 classes), there are 337 Z-pairs (674 classes with 1 partner),
		// 14 Z-triples (42 classes with 2 partners), and 3 Z-quadruples (12 classes with 3 partners).
		all16 := setclass.MustAll(16, setclass.TnIn)
		require.Len(t, all16, 2250)

		byICV := make(map[string][]setclass.SetClass)
		for _, sc := range all16 {
			if sc.Cardinality() < 2 {
				continue
			}
			key := sc.IntervalVector().String()
			byICV[key] = append(byICV[key], sc)
		}
		pairs, triples, quads := 0, 0, 0
		var sampleQuad []setclass.SetClass
		for _, group := range byICV {
			switch len(group) {
			case 2:
				pairs++
			case 3:
				triples++
			case 4:
				quads++
				if sampleQuad == nil {
					sampleQuad = group
				}
			}
		}
		assert.Equal(t, 337, pairs)
		assert.Equal(t, 14, triples)
		assert.Equal(t, 3, quads)

		// Verify ZPartners() on a Z-quadruple at N = 16 returns the other 3 classes in Compare order.
		require.Len(t, sampleQuad, 4)
		partners := sampleQuad[0].MustZPartners()
		require.Len(t, partners, 3)
		for _, p := range partners {
			assert.True(t, sampleQuad[0].MustIsZRelated(p))
		}

		// K-9(4): Error cases for Z-operations.
		sc24k12 := setclass.MustChromaticCluster(24, 12, setclass.TnIn)
		_, err := sc24k12.ZPartners()
		require.ErrorIs(t, err, setclass.ErrEnumerationLimitExceeded)
		_, err = sc24k12.HasZPartner()
		require.ErrorIs(t, err, setclass.ErrEnumerationLimitExceeded)

		sc65 := setclass.MustFromNumbers(65, setclass.TnIn, 0, 1)
		_, err = sc65.ZPartners()
		require.ErrorIs(t, err, setclass.ErrModulusTooLarge)

		sc12Tn := setclass.MustFromNumbers(12, setclass.Tn, 0, 3, 7)
		sc12TnIn := setclass.MustFromNumbers(12, setclass.TnIn, 0, 3, 7)
		sc19TnIn := setclass.MustFromNumbers(19, setclass.TnIn, 0, 3, 7)
		_, err = sc12Tn.IsZRelated(sc12TnIn)
		require.ErrorIs(t, err, setclass.ErrKindMismatch)
		_, err = sc12TnIn.IsZRelated(sc19TnIn)
		require.ErrorIs(t, err, setclass.ErrModulusMismatch)
		_, err = setclass.AreZRelatedSets(pcset.MustNew(12, 0, 3, 7), pcset.MustNew(19, 0, 3, 7))
		require.ErrorIs(t, err, setclass.ErrModulusMismatch)
		_, err = setclass.AreZRelatedSets(pcset.Set{}, pcset.MustNew(12, 0, 3, 7))
		require.ErrorIs(t, err, setclass.ErrInvalidValue)
	})
}

func TestInclusionEmbeddingsSubclassesAndSuperclasses(t *testing.T) {
	t.Parallel()

	t.Run("Abstract_inclusion_and_O1_trivial_fast_paths_up_to_MaxModulus", func(t *testing.T) {
		t.Parallel()

		// Trivial cardinality fast paths must work at N = 65_536 without ErrModulusTooLarge (OP-17, E-2.3).
		empBig := setclass.MustEmpty(65_536, setclass.TnIn)
		uniBig := setclass.MustUniverse(65_536, setclass.TnIn)
		pairBigA := setclass.MustFromNumbers(65_536, setclass.TnIn, 0, 1)
		pairBigB := setclass.MustFromNumbers(65_536, setclass.TnIn, 0, 2)
		triBig := setclass.MustFromNumbers(65_536, setclass.TnIn, 0, 1, 2)

		assert.True(t, empBig.MustIsSubclassOf(pairBigA))
		assert.True(t, empBig.MustIsProperSubclassOf(pairBigA))
		assert.False(t, empBig.MustIsProperSubclassOf(empBig))
		assert.True(t, pairBigA.MustIsSubclassOf(uniBig))
		assert.True(t, pairBigA.MustIsProperSubclassOf(uniBig))
		assert.True(t, pairBigA.MustIsSubclassOf(pairBigA))
		assert.False(t, pairBigA.MustIsProperSubclassOf(pairBigA))
		assert.False(t, pairBigA.MustIsSubclassOf(pairBigB))
		assert.False(t, triBig.MustIsSubclassOf(pairBigA))
		assert.True(t, uniBig.MustIncludes(pairBigA))
		assert.True(t, uniBig.MustStrictlyIncludes(pairBigA))

		// Non-trivial 0 < |X| < |Y| < N at N = 1025 returns ErrModulusTooLarge.
		sc1025_2 := setclass.MustFromNumbers(1_025, setclass.TnIn, 0, 1)
		sc1025_3 := setclass.MustFromNumbers(1_025, setclass.TnIn, 0, 1, 2)
		_, err := sc1025_2.IsSubclassOf(sc1025_3)
		require.ErrorIs(t, err, setclass.ErrModulusTooLarge)
		_, err = sc1025_3.IncludesSet(pcset.MustNew(1_025, 0, 1))
		require.ErrorIs(t, err, setclass.ErrModulusTooLarge)
		_, err = sc1025_2.IsEmbeddedInSet(pcset.MustNew(1_025, 0, 1, 2))
		require.ErrorIs(t, err, setclass.ErrModulusTooLarge)
	})

	t.Run("Lewin_EMB_published_benchmarks_and_identities", func(t *testing.T) {
		t.Parallel()

		// K-10(1): EMB(3-1, {1, 2, 3, 6, 7, 8, 10}) = 2 (Buchler ch. 1), EMB_Tn(3-11A, 7-20A) = 2, EMB_TnIn(3-11, 7-20) = 4, EMB(3-11, 7-35) = 6.
		sc3_1 := setclass.MustFromForteNotation("3-1", setclass.TnIn)
		sc3_11 := setclass.MustFromForteNotation("3-11", setclass.TnIn)
		sc7_20 := setclass.MustFromForteNotation("7-20", setclass.TnIn)
		sc7_35 := setclass.MustFromForteNotation("7-35", setclass.TnIn)
		setB := pcset.MustNew(12, 1, 2, 3, 6, 7, 8, 10)

		assert.True(t, sc3_1.MustIsEmbeddedInSet(setB))
		assert.True(t, sc7_20.MustIncludesSet(pcset.MustNew(12, 0, 1, 2)))
		assert.Equal(t, 2, sc3_1.MustEmbeddingCountInSet(setB))
		assert.Equal(t, 2, sc3_1.MustEmbeddingCount(sc7_20))
		assert.Equal(t, 2, setclass.MustCountEmbedding(sc3_1, sc7_20))
		assert.Equal(t, 2, setclass.MustCountSetEmbedding(pcset.MustNew(12, 0, 1, 2), setB, setclass.TnIn))
		assert.Equal(t, 4, sc3_11.MustEmbeddingCount(sc7_20))
		assert.Equal(t, 2, setclass.MustFromTnTypeNotation("3-11A").MustEmbeddingCount(setclass.MustFromTnTypeNotation("7-20A")))
		assert.Equal(t, 6, sc3_11.MustEmbeddingCount(sc7_35))
		assert.Equal(t, 0, sc7_35.MustEmbeddingCount(sc3_11))
		assert.Empty(t, sc7_35.MustEmbeddingOperators(sc3_11))

		// Stabilizer identity: |EmbeddingOperators(X, Y)| == EMB(X, Y) * |Stab_G(X)|.
		ops3_1InB := sc3_1.MustEmbeddingOperatorsInSet(setB)
		assert.Len(t, ops3_1InB, 2*sc3_1.SymmetryDegreeTI())
		assert.Equal(t, ops3_1InB, setclass.MustFindSetEmbeddingOperators(sc3_1.PrimeForm(), setB, setclass.TnIn))

		ops3_11In7_35 := sc3_11.MustEmbeddingOperators(sc7_35)
		assert.Len(t, ops3_11In7_35, 6*sc3_11.SymmetryDegreeTI())
		assert.Equal(t, ops3_11In7_35, setclass.MustFindEmbeddingOperators(sc3_11, sc7_35))

		// K-10(1): Webern hexachord 6-20 ((0, 1, 4, 5, 8, 9)) has a single subclass of cardinality 5 (5-21, EMB=6)
		// and a single superclass of cardinality 7 (7-21).
		sc6_20 := setclass.MustFromForteNotation("6-20", setclass.TnIn)
		sc5_21 := setclass.MustFromForteNotation("5-21", setclass.TnIn)
		sc7_21 := setclass.MustFromForteNotation("7-21", setclass.TnIn)

		sub5 := sc6_20.MustSubclassesOfCardinality(5)
		require.Equal(t, []setclass.SetClass{sc5_21}, sub5)
		prof5 := sc6_20.MustSubclassProfileOfCardinality(5)
		require.Equal(t, []setclass.SubsetProfileEntry{{Subclass: sc5_21, Count: 6}}, prof5)

		sup7 := sc6_20.MustSuperclassesOfCardinality(7)
		require.Equal(t, []setclass.SetClass{sc7_21}, sup7)

		// K-10(2): Dyad embedding counts EMB([{0, d}]_TnIn, Y) equal interval vector components m_d(Y).
		iv7_35 := sc7_35.IntervalVector()
		for d := 1; d <= 6; d++ {
			dyad := setclass.MustFromNumbers(12, setclass.TnIn, 0, d)
			assert.Equal(t, iv7_35.MustCount(d), dyad.MustEmbeddingCount(sc7_35), "d=%d", d)
		}

		// K-10(3): Sum of profile counts equals C(|Y|, k) and 2^|Y|.
		fullProf := sc6_20.MustSubclassProfile()
		allSubs := sc6_20.MustSubclasses()
		require.Len(t, fullProf, len(allSubs))
		totalSubsets := 0
		for i, entry := range fullProf {
			assert.True(t, entry.IsValid())
			assert.Equal(t, allSubs[i], entry.Subclass)
			totalSubsets += entry.Count
		}
		assert.Equal(t, 1<<sc6_20.Cardinality(), totalSubsets)

		cUniv := setclass.MustUniverse(12, setclass.TnIn)
		allSups := sc6_20.MustSuperclasses()
		assert.NotEmpty(t, allSups)
		assert.Equal(t, sc6_20, allSups[0])
		assert.Equal(t, cUniv, allSups[len(allSups)-1])

		// K-10(4): Complement reciprocity theorem:
		// X <=_G Y <=> Comp(Y) <=_G Comp(X) and EMB(X, Y)*OrbitSize(Y) == EMB(Comp(Y), Comp(X))*OrbitSize(X).
		sampleClasses := []setclass.SetClass{
			setclass.MustEmpty(12, setclass.TnIn),
			sc3_1,
			sc3_11,
			sc5_21,
			sc6_20,
			sc7_20,
			sc7_35,
			cUniv,
		}
		for _, x := range sampleClasses {
			for _, y := range sampleClasses {
				assert.Equal(t, x.MustIsSubclassOf(y), y.Complement().MustIsSubclassOf(x.Complement()))
				lhs := x.MustEmbeddingCount(y) * y.OrbitSize()
				rhs := y.Complement().MustEmbeddingCount(x.Complement()) * x.OrbitSize()
				assert.Equal(t, lhs, rhs, "reciprocity failed for X=%s Y=%s", x, y)
			}
		}

		// Error cases for Subclasses*, Superclasses*, and Embedding*.
		_, err := sc6_20.SubclassesOfCardinality(-1)
		require.ErrorIs(t, err, setclass.ErrInvalidCardinality)
		_, err = sc6_20.SubclassesOfCardinality(7)
		require.ErrorIs(t, err, setclass.ErrInvalidCardinality)
		_, err = sc6_20.SubclassProfileOfCardinality(7)
		require.ErrorIs(t, err, setclass.ErrInvalidCardinality)
		_, err = sc6_20.SuperclassesOfCardinality(5)
		require.ErrorIs(t, err, setclass.ErrInvalidCardinality)
		_, err = sc6_20.SuperclassesOfCardinality(13)
		require.ErrorIs(t, err, setclass.ErrInvalidCardinality)

		// |Y| > 20 in Subclasses / SubclassProfile and N - |X| > 20 in Superclasses -> ErrEnumerationLimitExceeded (E-3).
		sc24_21 := setclass.MustChromaticCluster(24, 21, setclass.TnIn)
		_, err = sc24_21.Subclasses()
		require.ErrorIs(t, err, setclass.ErrEnumerationLimitExceeded)
		_, err = sc24_21.SubclassProfile()
		require.ErrorIs(t, err, setclass.ErrEnumerationLimitExceeded)
		_, err = sc24_21.SubclassesOfCardinality(10) // C(21, 10) = 352_716 > 200_000
		require.ErrorIs(t, err, setclass.ErrEnumerationLimitExceeded)

		sc24_2 := setclass.MustChromaticCluster(24, 2, setclass.TnIn)
		_, err = sc24_2.Superclasses() // 24 - 2 = 22 > 20
		require.ErrorIs(t, err, setclass.ErrEnumerationLimitExceeded)
		_, err = sc24_2.SuperclassesOfCardinality(13) // C(22, 11) = 705_432 > 200_000
		require.ErrorIs(t, err, setclass.ErrEnumerationLimitExceeded)
	})
}

func TestCatalogBridgeMod12(t *testing.T) {
	t.Parallel()

	t.Run("Bijection_with_224_ForteClass_352_TnType_and_135_CarterClass", func(t *testing.T) {
		t.Parallel()

		for _, fc := range catalog.ForteClasses() {
			sc, err := setclass.FromForteClass(fc)
			require.NoError(t, err)
			assert.Equal(t, sc, setclass.MustFromForteClass(fc))
			assert.Equal(t, sc, setclass.MustFromForteNumber(fc.Cardinality(), fc.Ordinal(), setclass.TnIn))
			assert.Equal(t, sc, setclass.MustFromCatalogName(fc.String(), setclass.TnIn))
			assert.Equal(t, sc, setclass.MustFromForteNotation(fc.String(), setclass.TnIn))

			gotFC, err := sc.ForteClass()
			require.NoError(t, err)
			assert.Equal(t, fc, gotFC)
			assert.Equal(t, fc, sc.MustForteClass())
			assert.Equal(t, fc, setclass.MustForteClass(sc))
			assert.Equal(t, fc.String(), sc.MustCatalogName())
			assert.Equal(t, fc.TnTypes(), sc.MustCatalogTnTypes())

			cc, hasCarter, err := sc.CarterClass()
			require.NoError(t, err)
			mustCC, mustHasCarter := sc.MustCarterClass()
			assert.Equal(t, cc, mustCC)
			assert.Equal(t, hasCarter, mustHasCarter)
			cn, hasCarterNum, err := sc.CarterNumber()
			require.NoError(t, err)
			mustCN, mustHasCarterNum := sc.MustCarterNumber()
			assert.Equal(t, cn, mustCN)
			assert.Equal(t, hasCarterNum, mustHasCarterNum)
			assert.Equal(t, fc.HasCarter(), hasCarter)
			assert.Equal(t, fc.HasCarter(), hasCarterNum)
			if hasCarter {
				expCC, _ := fc.Carter()
				expCN, _ := fc.CarterNumber()
				assert.Equal(t, expCC, cc)
				assert.Equal(t, expCN, cn)
				assert.Equal(t, sc, setclass.MustFromCarterClass(cc))
				assert.Equal(t, sc, setclass.MustFromCarterNumber(fc.Cardinality(), cn))
			}

			zp, hasZPair, err := sc.ZPair()
			require.NoError(t, err)
			assert.Equal(t, fc.IsZ(), hasZPair)
			if hasZPair {
				expZP, _ := fc.ZPair()
				assert.Equal(t, expZP, zp)
			}

			zPart, hasZPart, err := sc.ZPartner()
			require.NoError(t, err)
			assert.Equal(t, fc.IsZ(), hasZPart)
			if hasZPart {
				expZFC, _ := fc.ZPartner()
				assert.Equal(t, setclass.MustFromForteClass(expZFC), zPart)
			}
		}

		for _, tt := range catalog.TnTypes() {
			sc, err := setclass.FromTnType(tt)
			require.NoError(t, err)
			assert.Equal(t, sc, setclass.MustFromTnType(tt))
			assert.Equal(t, sc, setclass.MustFromTnTypeNumber(tt.Cardinality(), tt.Ordinal(), tt.Suffix()))
			assert.Equal(t, sc, setclass.MustFromCatalogName(tt.String(), setclass.Tn))
			assert.Equal(t, sc, setclass.MustFromForteNotation(tt.String(), setclass.Tn))
			assert.Equal(t, sc, setclass.MustFromTnTypeNotation(tt.String()))

			gotTT, err := sc.TnType()
			require.NoError(t, err)
			assert.Equal(t, tt, gotTT)
			assert.Equal(t, tt, sc.MustTnType())
			assert.Equal(t, []catalog.TnType{tt}, sc.MustCatalogTnTypes())
			assert.Equal(t, tt.ForteClass(), sc.MustForteClass())
			assert.Equal(t, tt.String(), sc.MustCatalogName())
		}

		for _, cc := range catalog.CarterClasses() {
			sc, err := setclass.FromCarterClass(cc)
			require.NoError(t, err)
			assert.Equal(t, cc.ForteClass(), sc.MustForteClass())
		}
	})

	t.Run("TTO_composite_Morris_names_and_5_second_component_prime_form_classes", func(t *testing.T) {
		t.Parallel()

		// K-11(2, 3): Verify "6-1/32", "6-1", and "6-32" all parse into the same TTO class [0, 1, 2, 3, 4, 5].
		tto6_1_32 := setclass.MustFromCatalogName("6-1/32", setclass.TTO)
		assert.Equal(t, tto6_1_32, setclass.MustFromCatalogName("6-1", setclass.TTO))
		assert.Equal(t, tto6_1_32, setclass.MustFromCatalogName("6-32", setclass.TTO))
		assert.Equal(t, tto6_1_32, setclass.MustFromForteNotation("6-1/32", setclass.TTO))
		assert.Equal(t, tto6_1_32, setclass.MustFromForteNumber(6, 1, setclass.TTO))
		assert.Equal(t, tto6_1_32, setclass.MustFromForteNumber(6, 32, setclass.TTO))
		assert.Equal(t, "6-1/32", tto6_1_32.MustCatalogName())
		assert.Equal(t, []int{0, 1, 2, 3, 4, 5}, tto6_1_32.PrimeNumbers())

		// K-11(2): Exactly 5 composite TTO classes have their Rahn prime form in the second ForteClass (d2 > d1),
		// while CatalogName() still preserves d1 < d2!
		secondWins := map[string]string{
			"6-Z24/Z39": "6-Z39",
			"6-Z29/Z42": "6-Z42",
			"7-Z17/Z37": "7-Z37",
			"7-Z18/Z38": "7-Z38",
			"8-Z15/Z29": "8-Z29",
		}

		actualSecondWins := make(map[string]string)
		for _, tto := range setclass.MustAll(12, setclass.TTO) {
			name := tto.MustCatalogName()
			parsed, err := setclass.FromCatalogName(name, setclass.TTO)
			require.NoError(t, err, "CatalogName %s", name)
			assert.Equal(t, tto, parsed)

			subs := tto.MustTnInClasses()
			if len(subs) == 2 {
				fc1 := subs[0].MustForteClass()
				fc2 := subs[1].MustForteClass()
				if tto.MustForteClass() == fc2 && fc2.Ordinal() > fc1.Ordinal() {
					actualSecondWins[name] = fc2.String()
				} else if tto.MustForteClass() == fc1 && fc1.Ordinal() > fc2.Ordinal() {
					actualSecondWins[name] = fc1.String()
				}
			}
		}
		assert.Equal(t, secondWins, actualSecondWins)
	})

	t.Run("Catalog_bridge_error_cases_and_dual_error_wrapping", func(t *testing.T) {
		t.Parallel()

		// Invalid catalog structs wrap both setclass.ErrInvalidValue and catalog.ErrInvalidValue (E-6.3).
		_, err := setclass.FromForteClass(catalog.ForteClass{})
		require.ErrorIs(t, err, setclass.ErrInvalidValue)
		require.ErrorIs(t, err, catalog.ErrInvalidValue)
		require.Panics(t, func() { _ = setclass.MustFromForteClass(catalog.ForteClass{}) })

		_, err = setclass.FromTnType(catalog.TnType{})
		require.ErrorIs(t, err, setclass.ErrInvalidValue)
		require.ErrorIs(t, err, catalog.ErrInvalidValue)
		require.Panics(t, func() { _ = setclass.MustFromTnType(catalog.TnType{}) })

		_, err = setclass.FromCarterClass(catalog.CarterClass{})
		require.ErrorIs(t, err, setclass.ErrInvalidValue)
		require.ErrorIs(t, err, catalog.ErrInvalidValue)
		require.Panics(t, func() { _ = setclass.MustFromCarterClass(catalog.CarterClass{}) })

		// Invalid Carter numbers / cardinalities wrap both setclass and catalog markers.
		_, err = setclass.FromCarterNumber(13, 1)
		require.ErrorIs(t, err, setclass.ErrInvalidCardinality)
		require.ErrorIs(t, err, catalog.ErrInvalidCardinality)

		_, err = setclass.FromCarterNumber(5, 1)
		require.ErrorIs(t, err, catalog.ErrNoCarterCatalog)
		require.Panics(t, func() { _ = setclass.MustFromCarterNumber(5, 1) })

		_, err = setclass.FromCarterNumber(3, 99)
		require.ErrorIs(t, err, catalog.ErrInvalidCarterNumber)

		// Invalid ForteNumber / TnTypeNumber / CatalogName / ForteNotation / TnTypeNotation.
		_, err = setclass.FromForteNumber(3, 11, setclass.Tn) // chiral class in Tn requires A/B suffix
		require.ErrorIs(t, err, setclass.ErrNotation)
		require.ErrorIs(t, err, catalog.ErrInvalidSuffix)
		require.Panics(t, func() { _ = setclass.MustFromForteNumber(3, 11, setclass.Tn) })

		_, err = setclass.FromForteNumber(3, 1, setclass.Kind(99))
		require.ErrorIs(t, err, setclass.ErrInvalidKind)

		_, err = setclass.FromTnTypeNumber(3, 1, catalog.SuffixA)
		require.ErrorIs(t, err, catalog.ErrInvalidSuffix)
		require.Panics(t, func() { _ = setclass.MustFromTnTypeNumber(3, 1, catalog.SuffixA) })

		for _, badStr := range []string{"", "4-15", "3-1A", "13-1", "3-99"} {
			_, err = setclass.FromCatalogName(badStr, setclass.TnIn)
			require.ErrorIs(t, err, setclass.ErrNotation)
			require.ErrorIs(t, err, catalog.ErrInvalidNotation)
		}
		require.Panics(t, func() { _ = setclass.MustFromCatalogName("4-15", setclass.TnIn) })
		require.Panics(t, func() { _ = setclass.MustFromForteNotation("4-15", setclass.TnIn) })
		require.Panics(t, func() { _ = setclass.MustFromTnTypeNotation("3-11") })

		// Non-M5-partners in composite TTO notation (e.g. "6-1/2" or reversed "6-32/1") are rejected.
		for _, badTTO := range []string{"6-1/2", "6-32/1", "6-20/20", "3-11A"} {
			_, err = setclass.FromCatalogName(badTTO, setclass.TTO)
			require.ErrorIs(t, err, setclass.ErrNotation)
			require.ErrorIs(t, err, catalog.ErrInvalidNotation)
		}

		// Calling catalog projection methods on modulus != 12 returns ErrNotModulus12;
		// calling TnType() on TnIn or TTO returns ErrKindMismatch.
		sc19 := setclass.MustFromNumbers(19, setclass.TnIn, 0, 3, 7)
		_, err = sc19.ForteClass()
		require.ErrorIs(t, err, setclass.ErrNotModulus12)
		_, err = sc19.TnType()
		require.ErrorIs(t, err, setclass.ErrNotModulus12)
		_, err = sc19.CatalogTnTypes()
		require.ErrorIs(t, err, setclass.ErrNotModulus12)
		_, _, err = sc19.CarterClass()
		require.ErrorIs(t, err, setclass.ErrNotModulus12)
		_, _, err = sc19.CarterNumber()
		require.ErrorIs(t, err, setclass.ErrNotModulus12)
		_, _, err = sc19.ZPair()
		require.ErrorIs(t, err, setclass.ErrNotModulus12)
		_, _, err = sc19.ZPartner()
		require.ErrorIs(t, err, setclass.ErrNotModulus12)
		_, err = sc19.CatalogName()
		require.ErrorIs(t, err, setclass.ErrNotModulus12)

		sc12TnIn := setclass.MustFromForteNotation("3-11", setclass.TnIn)
		_, err = sc12TnIn.TnType()
		require.ErrorIs(t, err, setclass.ErrKindMismatch)
		require.Panics(t, func() { _ = sc12TnIn.MustTnType() })
	})
}

func TestBurnsideCountAndEnumeration(t *testing.T) {
	t.Parallel()

	t.Run("Burnside_counts_and_All_for_N_1_to_20_and_microtonal_24_31", func(t *testing.T) {
		t.Parallel()

		// K-12(1, 2): Exact reference triples (Tn, TnIn, TTO) for N = 1..20.
		expectedCounts := []struct {
			n    int
			tn   int
			tnIn int
			tto  int
		}{
			{n: 1, tn: 2, tnIn: 2, tto: 2},
			{n: 2, tn: 3, tnIn: 3, tto: 3},
			{n: 3, tn: 4, tnIn: 4, tto: 4},
			{n: 4, tn: 6, tnIn: 6, tto: 6},
			{n: 5, tn: 8, tnIn: 8, tto: 6},
			{n: 6, tn: 14, tnIn: 13, tto: 13},
			{n: 7, tn: 20, tnIn: 18, tto: 10},
			{n: 8, tn: 36, tnIn: 30, tto: 24},
			{n: 9, tn: 60, tnIn: 46, tto: 22},
			{n: 10, tn: 108, tnIn: 78, tto: 45},
			{n: 11, tn: 188, tnIn: 126, tto: 30},
			{n: 12, tn: 352, tnIn: 224, tto: 158},
			{n: 13, tn: 632, tnIn: 380, tto: 74},
			{n: 14, tn: 1_182, tnIn: 687, tto: 245},
			{n: 15, tn: 2_192, tnIn: 1_224, tto: 368},
			{n: 16, tn: 4_116, tnIn: 2_250, tto: 693},
			{n: 17, tn: 7_712, tnIn: 4_112, tto: 522},
			{n: 18, tn: 14_602, tnIn: 7_685, tto: 2_637},
			{n: 19, tn: 27_596, tnIn: 14_310, tto: 1_610},
			{n: 20, tn: 52_488, tnIn: 27_012, tto: 7_386},
		}

		for _, tc := range expectedCounts {
			assert.Equal(t, tc.tn, setclass.MustCount(tc.n, setclass.Tn), "N=%d Tn", tc.n)
			assert.Equal(t, tc.tnIn, setclass.MustCount(tc.n, setclass.TnIn), "N=%d TnIn", tc.n)
			assert.Equal(t, tc.tto, setclass.MustCount(tc.n, setclass.TTO), "N=%d TTO", tc.n)
			assert.Equal(t, int64(tc.tn), setclass.MustCountBig(tc.n, setclass.Tn).Int64())
			assert.Equal(t, int64(tc.tnIn), setclass.MustCountBig(tc.n, setclass.TnIn).Int64())
			assert.Equal(t, int64(tc.tto), setclass.MustCountBig(tc.n, setclass.TTO).Int64())

			for _, kind := range setclass.Kinds() {
				sumByCard := 0
				for k := 0; k <= tc.n; k++ {
					ck := setclass.MustCountOfCardinality(tc.n, k, kind)
					assert.Equal(t, int64(ck), setclass.MustCountOfCardinalityBig(tc.n, k, kind).Int64())
					assert.Equal(t, ck, setclass.MustCountOfCardinality(tc.n, tc.n-k, kind))
					sumByCard += ck
				}
				assert.Equal(t, setclass.MustCount(tc.n, kind), sumByCard, "N=%d kind=%s", tc.n, kind)
			}
		}

		// Verify All and AllWith up to N = 12 (and N = 19, 20 for TTO/TnIn).
		for _, n := range []int{1, 2, 5, 6, 8, 10, 12} {
			for _, kind := range setclass.Kinds() {
				allRahn, err := setclass.All(n, kind)
				require.NoError(t, err)
				assert.Equal(t, allRahn, setclass.MustAll(n, kind))
				assert.Equal(t, allRahn, setclass.MustAllWith(n, kind, setclass.Rahn))
				assert.Len(t, allRahn, setclass.MustCount(n, kind))

				allForte, err := setclass.AllWith(n, kind, setclass.Forte)
				require.NoError(t, err)
				assert.Equal(t, allForte, setclass.MustAllWith(n, kind, setclass.Forte))
				assert.Len(t, allForte, len(allRahn))
				for idx := 1; idx < len(allForte); idx++ {
					assert.Equal(t, -1, allForte[idx-1].MustCompareWith(allForte[idx], setclass.Forte))
				}
			}
		}
		assert.Len(t, setclass.MustAll(19, setclass.TTO), 1_610)
		assert.Len(t, setclass.MustAll(20, setclass.TTO), 7_386)

		// Microtonal N = 24 and N = 31 (G-8, K-12(2)).
		assert.Equal(t, 699_252, setclass.MustCount(24, setclass.Tn))
		assert.Equal(t, 352_698, setclass.MustCount(24, setclass.TnIn))
		assert.Equal(t, 94_484, setclass.MustCount(24, setclass.TTO))

		assert.Equal(t, 69_273_668, setclass.MustCount(31, setclass.Tn))
		assert.Equal(t, 34_669_602, setclass.MustCount(31, setclass.TnIn))
		assert.Equal(t, 2_311_470, setclass.MustCount(31, setclass.TTO))
	})

	t.Run("Mod12_cardinality_distribution_OfCardinality_and_ByIntervalVector", func(t *testing.T) {
		t.Parallel()

		// K-12(2): Cardinality distributions for k = 0..12 at N = 12.
		expTn := []int{1, 1, 6, 19, 43, 66, 80, 66, 43, 19, 6, 1, 1}
		expTnIn := []int{1, 1, 6, 12, 29, 38, 50, 38, 29, 12, 6, 1, 1}
		expTTO := []int{1, 1, 5, 9, 21, 25, 34, 25, 21, 9, 5, 1, 1}

		for k := 0; k <= 12; k++ {
			assert.Equal(t, expTn[k], setclass.MustCountOfCardinality(12, k, setclass.Tn))
			assert.Equal(t, expTnIn[k], setclass.MustCountOfCardinality(12, k, setclass.TnIn))
			assert.Equal(t, expTTO[k], setclass.MustCountOfCardinality(12, k, setclass.TTO))

			classesRahn, err := setclass.OfCardinality(12, k, setclass.TnIn)
			require.NoError(t, err)
			assert.Equal(t, classesRahn, setclass.MustOfCardinality(12, k, setclass.TnIn))
			assert.Equal(t, classesRahn, setclass.MustOfCardinalityWith(12, k, setclass.TnIn, setclass.Rahn))
			assert.Len(t, classesRahn, expTnIn[k])

			classesForte, err := setclass.OfCardinalityWith(12, k, setclass.TnIn, setclass.Forte)
			require.NoError(t, err)
			assert.Equal(t, classesForte, setclass.MustOfCardinalityWith(12, k, setclass.TnIn, setclass.Forte))
			assert.Len(t, classesForte, expTnIn[k])
			for i := 1; i < len(classesForte); i++ {
				assert.Equal(t, -1, classesForte[i-1].MustCompareWith(classesForte[i], setclass.Forte))
			}
		}

		// ByIntervalVector for all-interval tetrachord vector <1, 1, 1, 1, 1, 1> returns [4-Z15, 4-Z29] in TnIn,
		// [4-Z15A, 4-Z15B, 4-Z29A, 4-Z29B] in Tn, and [4-Z15/Z29] in TTO.
		ivAllInterval := pcset.MustNew(12, 0, 1, 4, 6).IntervalVector()
		byTnIn, err := setclass.ByIntervalVector(ivAllInterval, setclass.TnIn)
		require.NoError(t, err)
		assert.Equal(t, byTnIn, setclass.MustByIntervalVector(ivAllInterval, setclass.TnIn))
		assert.Len(t, byTnIn, 2)
		assert.Len(t, setclass.MustByIntervalVector(ivAllInterval, setclass.Tn), 4)
		assert.Len(t, setclass.MustByIntervalVector(ivAllInterval, setclass.TTO), 1)
	})

	t.Run("Overflow_and_modulus_boundaries_69_70_71_74_75_256_1024", func(t *testing.T) {
		t.Parallel()

		// G-9, K-12(3), E-9.2:
		// Tn overflows int at N = 70 (N = 69 fits);
		// TnIn overflows int at N = 71 (N = 70 fits);
		// TTO overflows int at N = 75 (N = 74 fits).
		c69Tn, err := setclass.Count(69, setclass.Tn)
		require.NoError(t, err)
		assert.Positive(t, c69Tn)
		_, err = setclass.Count(70, setclass.Tn)
		require.ErrorIs(t, err, setclass.ErrOverflow)
		require.Panics(t, func() { _ = setclass.MustCount(70, setclass.Tn) })

		c70TnIn, err := setclass.Count(70, setclass.TnIn)
		require.NoError(t, err)
		assert.Positive(t, c70TnIn)
		_, err = setclass.Count(71, setclass.TnIn)
		require.ErrorIs(t, err, setclass.ErrOverflow)
		require.Panics(t, func() { _ = setclass.MustCount(71, setclass.TnIn) })

		c74TTO, err := setclass.Count(74, setclass.TTO)
		require.NoError(t, err)
		assert.Positive(t, c74TTO)
		_, err = setclass.Count(75, setclass.TTO)
		require.ErrorIs(t, err, setclass.ErrOverflow)

		// CountOfCardinality overflow at N = 75, k = 37.
		_, err = setclass.CountOfCardinality(75, 37, setclass.Tn)
		require.ErrorIs(t, err, setclass.ErrOverflow)
		require.Panics(t, func() { _ = setclass.MustCountOfCardinality(75, 37, setclass.Tn) })

		// CountBig and CountOfCardinalityBig succeed up to N = 1024 (Tn/TnIn) and N = 256 (TTO).
		big1024, err := setclass.CountBig(1_024, setclass.TnIn)
		require.NoError(t, err)
		assert.True(t, big1024.BitLen() > 1000)

		big256TTO, err := setclass.CountBig(256, setclass.TTO)
		require.NoError(t, err)
		assert.True(t, big256TTO.BitLen() > 240)

		bigCard, err := setclass.CountOfCardinalityBig(1_024, 2, setclass.TnIn)
		require.NoError(t, err)
		assert.Equal(t, int64(512), bigCard.Int64())

		// Threshold errors: N = 21 for All, N = 257 for TTO, N = 1025 for CountBig(TnIn), C(24, 9) > 200_000 for OfCardinality.
		_, err = setclass.All(21, setclass.TnIn)
		require.ErrorIs(t, err, setclass.ErrModulusTooLarge)
		_, err = setclass.AllWith(21, setclass.TnIn, setclass.Rahn)
		require.ErrorIs(t, err, setclass.ErrModulusTooLarge)
		_, err = setclass.AllWith(12, setclass.TnIn, setclass.Algorithm(99))
		require.ErrorIs(t, err, setclass.ErrInvalidAlgorithm)

		_, err = setclass.OfCardinality(24, 9, setclass.TnIn) // C(24, 9) = 1_307_504 > 200_000
		require.ErrorIs(t, err, setclass.ErrEnumerationLimitExceeded)
		_, err = setclass.OfCardinality(65, 2, setclass.TnIn)
		require.ErrorIs(t, err, setclass.ErrModulusTooLarge)
		_, err = setclass.OfCardinalityWith(12, 3, setclass.TnIn, setclass.Algorithm(0))
		require.ErrorIs(t, err, setclass.ErrInvalidAlgorithm)

		_, err = setclass.CountBig(257, setclass.TTO)
		require.ErrorIs(t, err, setclass.ErrModulusTooLarge)
		_, err = setclass.CountBig(1_025, setclass.TnIn)
		require.ErrorIs(t, err, setclass.ErrModulusTooLarge)
		_, err = setclass.Count(0, setclass.TnIn)
		require.ErrorIs(t, err, setclass.ErrInvalidModulus)
		_, err = setclass.Count(12, setclass.Kind(99))
		require.ErrorIs(t, err, setclass.ErrInvalidKind)

		_, err = setclass.ByIntervalVector(pcset.IntervalVector{}, setclass.TnIn)
		require.ErrorIs(t, err, setclass.ErrInvalidValue)
		_, err = setclass.ByIntervalVector(pcset.MustNew(12, 0, 3, 7).IntervalVector(), setclass.Kind(0))
		require.ErrorIs(t, err, setclass.ErrInvalidKind)
	})
}

func TestNotationsAndParsing(t *testing.T) {
	t.Parallel()

	t.Run("Canonical_String_Canonical_and_strict_Parse", func(t *testing.T) {
		t.Parallel()

		// K-13(1): Round-trip Parse(sc.String()) == sc across all classes in N = 12 and sample in N = 65_536.
		for _, kind := range setclass.Kinds() {
			for _, sc := range setclass.MustAll(12, kind) {
				s := sc.String()
				assert.Equal(t, s, sc.Canonical())
				parsed, err := setclass.Parse(s)
				require.NoError(t, err, "String=%q", s)
				assert.Equal(t, sc, parsed)
				assert.Equal(t, sc, setclass.MustParse(s))
			}
		}

		assert.Equal(t, "TnIn(0, 3, 7) mod 12", setclass.MustFromNumbers(12, setclass.TnIn, 0, 4, 7).String())
		assert.Equal(t, "Tn(0, 4, 7) mod 12", setclass.MustFromNumbers(12, setclass.Tn, 0, 4, 7).String())
		assert.Equal(t, "TTO(0, 1, 4) mod 12", setclass.MustFromNumbers(12, setclass.TTO, 0, 4, 7).String())
		assert.Equal(t, "TnIn() mod 12", setclass.MustEmpty(12, setclass.TnIn).String())

		scBig := setclass.MustFromNumbers(65_536, setclass.TnIn, 0, 1, 65_535)
		assert.Equal(t, scBig, setclass.MustParse(scBig.String()))

		// Strict Parse rejects non-Rahn-prime-form tuples and syntax errors (E-8.1, K-13(1)).
		badCanonical := []struct {
			input   string
			wantErr error
		}{
			{input: "", wantErr: setclass.ErrNotation},
			{input: "SetClass(invalid)", wantErr: setclass.ErrNotation},
			{input: "TnIn(0, 4, 7) mod 12", wantErr: setclass.ErrNotation},       // not Rahn prime form (should be 0, 3, 7)
			{input: "TnIn(0, 1, 3, 7, 8) mod 12", wantErr: setclass.ErrNotation}, // Forte prime form of 5-20, not Rahn (0, 1, 5, 6, 8)
			{input: "Tn(1, 4, 8) mod 12", wantErr: setclass.ErrNotation},         // does not start at 0
			{input: "Tn(0, 3, 3) mod 12", wantErr: setclass.ErrNotation},         // duplicate
			{input: "Tn(0, 7, 3) mod 12", wantErr: setclass.ErrNotation},         // not sorted
			{input: "Tn(0, 03, 7) mod 12", wantErr: setclass.ErrNotation},        // leading zero
			{input: "Tn(0, 12) mod 12", wantErr: setclass.ErrNotation},           // >= modulus
			{input: "Unknown(0, 3, 7) mod 12", wantErr: setclass.ErrNotation},    // unknown kind
			{input: "{0, 3, 7} mod 12", wantErr: setclass.ErrNotation},           // missing kind / wrong brackets
			{input: "TnIn(0, 3, 7) mod 0", wantErr: setclass.ErrInvalidModulus},
			{input: "TnIn(0, 3, 7) mod -1", wantErr: setclass.ErrInvalidModulus},
			{input: "TnIn(0, 1) mod 65537", wantErr: setclass.ErrModulusTooLarge},
			{input: "TTO(0, 1) mod 257", wantErr: setclass.ErrModulusTooLarge},
		}
		for _, tc := range badCanonical {
			_, err := setclass.Parse(tc.input)
			require.ErrorIs(t, err, tc.wantErr, "input=%q", tc.input)
			require.Panics(t, func() { _ = setclass.MustParse(tc.input) })
		}
	})

	t.Run("Literary_Notation_CompactNotation_and_ParseNotation", func(t *testing.T) {
		t.Parallel()

		// K-13(2): 8-26 has Rahn prime form (0, 1, 3, 4, 5, 7, 8, 10) and Forte prime form (0, 1, 2, 4, 5, 7, 9, 10).
		sc8_26 := setclass.MustFromForteNotation("8-26", setclass.TnIn)
		assert.Equal(t, "(0, 1, 3, 4, 5, 7, 8, 10)", sc8_26.MustNotation(pitchclass.Decimal))
		assert.Equal(t, "(0, 1, 3, 4, 5, 7, 8, T)", sc8_26.MustNotation(pitchclass.TE))
		assert.Equal(t, "(0, 1, 3, 4, 5, 7, 8, t)", sc8_26.MustNotation(pitchclass.LowerTE))
		assert.Equal(t, "(0, 1, 3, 4, 5, 7, 8, A)", sc8_26.MustNotation(pitchclass.AB))
		assert.Equal(t, "(0, 1, 2, 4, 5, 7, 9, T)", sc8_26.MustNotationWith(pitchclass.TE, setclass.Forte))

		assert.Equal(t, "(0134578T)", sc8_26.MustCompactNotation(pitchclass.TE))
		assert.Equal(t, "(0134578t)", sc8_26.MustCompactNotation(pitchclass.LowerTE))
		assert.Equal(t, "(0134578A)", sc8_26.MustCompactNotation(pitchclass.AB))
		assert.Equal(t, "(0124579T)", sc8_26.MustCompactNotationWith(pitchclass.TE, setclass.Forte))

		emp := setclass.MustEmpty(12, setclass.TnIn)
		assert.Equal(t, "()", emp.MustNotation(pitchclass.TE))
		assert.Equal(t, "()", emp.MustCompactNotation(pitchclass.TE))

		// CompactNotation(Decimal) works for N <= 10 and returns ErrNotation for N > 10 (E-8.2, K-13(2)).
		sc10 := setclass.MustFromNumbers(10, setclass.TnIn, 0, 3, 7)
		assert.Equal(t, "(036)", sc10.MustCompactNotation(pitchclass.Decimal))

		sc12 := setclass.MustFromNumbers(12, setclass.TnIn, 0, 3, 7)
		_, err := sc12.CompactNotation(pitchclass.Decimal)
		require.ErrorIs(t, err, setclass.ErrNotation)

		// Letter styles return ErrNotation when N > 12.
		sc13 := setclass.MustFromNumbers(13, setclass.TnIn, 0, 3, 7)
		_, err = sc13.Notation(pitchclass.TE)
		require.ErrorIs(t, err, setclass.ErrNotation)
		_, err = sc13.CompactNotation(pitchclass.TE)
		require.ErrorIs(t, err, setclass.ErrNotation)

		// K-13(3): ParseNotation accepts both prime forms and arbitrary set members in parentheses "(...)".
		for _, input := range []string{"(0, 3, 7)", "(037)", "(0, 4, 7)", "(2, 6, 9)", "(9, 0, 4)"} {
			parsed, err := setclass.ParseNotation(input, pitchclass.TE, 12, setclass.TnIn)
			require.NoError(t, err, "input=%q", input)
			assert.Equal(t, sc12, parsed)
			assert.Equal(t, sc12, setclass.MustParseNotation(input, pitchclass.TE, 12, setclass.TnIn))
		}
		assert.Equal(t, emp, setclass.MustParseNotation("()", pitchclass.TE, 12, setclass.TnIn))

		// ParseNotation rejects curly/square brackets, invalid style, invalid kind, and out-of-range numbers.
		for _, bad := range []string{"", "{0, 3, 7}", "[0, 3, 7]", "(0, 3, 7]", "(0, 12)", "(037"} {
			_, err = setclass.ParseNotation(bad, pitchclass.TE, 12, setclass.TnIn)
			require.ErrorIs(t, err, setclass.ErrNotation, "bad=%q", bad)
		}
		_, err = setclass.ParseNotation("(037)", pitchclass.Decimal, 12, setclass.TnIn)
		require.ErrorIs(t, err, setclass.ErrNotation)
		_, err = setclass.ParseNotation("(0, 3, 7)", pitchclass.Style(99), 12, setclass.TnIn)
		require.ErrorIs(t, err, setclass.ErrNotation)
		_, err = setclass.ParseNotation("(0, 3, 7)", pitchclass.TE, 0, setclass.TnIn)
		require.ErrorIs(t, err, setclass.ErrInvalidModulus)
		_, err = setclass.ParseNotation("(0, 3, 7)", pitchclass.Decimal, 65_537, setclass.TnIn)
		require.ErrorIs(t, err, setclass.ErrModulusTooLarge)
		_, err = setclass.ParseNotation("(0, 3, 7)", pitchclass.TE, 12, setclass.Kind(99))
		require.ErrorIs(t, err, setclass.ErrInvalidKind)
	})

	t.Run("Morris_notation_FormatMorris_ParseMorris_and_ParseMorrisWithKind", func(t *testing.T) {
		t.Parallel()

		// K-13(4): FormatMorris and ParseMorris across Tn, TnIn, and TTO at N = 12.
		for _, kind := range setclass.Kinds() {
			for _, sc := range setclass.MustAll(12, kind) {
				morris, err := sc.FormatMorris()
				require.NoError(t, err)
				assert.Equal(t, morris, sc.MustFormatMorris())
				assert.Equal(t, morris, sc.MustFormatMorrisWith(setclass.Rahn))

				parsedWithKind, err := setclass.ParseMorrisWithKind(morris, kind)
				require.NoError(t, err, "morris=%q kind=%s", morris, kind)
				assert.Equal(t, sc, parsedWithKind)
				assert.Equal(t, sc, setclass.MustParseMorrisWithKind(morris, kind))

				// ParseMorris auto-detects Tn for chiral A/B labels, TTO for composite c-d1/d2 labels, and TnIn for single c-d labels.
				if kind == setclass.TnIn {
					parsedAuto, err := setclass.ParseMorris(morris)
					require.NoError(t, err)
					assert.Equal(t, sc, parsedAuto)
					assert.Equal(t, sc, setclass.MustParseMorris(morris))
				}
			}
		}

		sc5_20 := setclass.MustFromForteNotation("5-20", setclass.TnIn)
		assert.Equal(t, "5-20[01568]", sc5_20.MustFormatMorris())
		assert.Equal(t, "5-20[01378]", sc5_20.MustFormatMorrisWith(setclass.Forte))
		assert.Equal(t, sc5_20, setclass.MustParseMorris("5-20[01378]"))
		assert.Equal(t, sc5_20, setclass.MustParseMorris("5-20[0, 1, 3, 7, 8]"))

		sc3_11B := setclass.MustFromTnTypeNotation("3-11B")
		assert.Equal(t, "3-11B[047]", sc3_11B.MustFormatMorris())
		assert.Equal(t, sc3_11B, setclass.MustParseMorris("3-11B[047]"))

		sc6_1_32 := setclass.MustFromCatalogName("6-1/32", setclass.TTO)
		assert.Equal(t, "6-1/32[012345]", sc6_1_32.MustFormatMorris())
		assert.Equal(t, sc6_1_32, setclass.MustParseMorris("6-1/32[012345]"))
		assert.Equal(t, sc6_1_32, setclass.MustParseMorris("6-1/32[024579]"))

		// Mismatched set and catalog label in Morris notation returns ErrNotation (E-8.3).
		for _, badMorris := range []string{"", "3-11[012]", "3-11B[037]", "6-1/32[012346]", "3-11(037)", "3-11[037"} {
			_, err := setclass.ParseMorris(badMorris)
			require.ErrorIs(t, err, setclass.ErrNotation, "badMorris=%q", badMorris)
			require.Panics(t, func() { _ = setclass.MustParseMorris(badMorris) })
		}

		// FormatMorris on modulus != 12 returns ErrNotModulus12.
		sc19 := setclass.MustFromNumbers(19, setclass.TnIn, 0, 3, 7)
		_, err := sc19.FormatMorris()
		require.ErrorIs(t, err, setclass.ErrNotModulus12)
	})
}

func TestDefensiveCopiesAndConcurrency(t *testing.T) {
	t.Parallel()

	t.Run("Defensive_copies_of_all_input_and_output_slices_and_big_Int", func(t *testing.T) {
		t.Parallel()

		// G-10: Mutating input slices or *big.Int after construction must not affect SetClass.
		inNums := []int{0, 4, 7}
		scFromNums := setclass.MustFromNumbers(12, setclass.TnIn, inNums...)
		inNums[0] = 11
		assert.Equal(t, []int{0, 3, 7}, scFromNums.PrimeNumbers())

		inClasses := []pitchclass.Class{
			pitchclass.MustNew(0, 12),
			pitchclass.MustNew(4, 12),
			pitchclass.MustNew(7, 12),
		}
		scFromClasses := setclass.MustFromClasses(12, setclass.TnIn, inClasses...)
		inClasses[0] = pitchclass.MustNew(1, 12)
		assert.Equal(t, []int{0, 3, 7}, scFromClasses.PrimeNumbers())

		inBig := big.NewInt(145)
		scFromBig := setclass.MustFromBitmaskBig(12, setclass.TnIn, inBig)
		inBig.SetInt64(0)
		assert.Equal(t, []int{0, 3, 7}, scFromBig.PrimeNumbers())

		// Mutating returned slices or *big.Int must not affect subsequent calls.
		pNums := scFromNums.PrimeNumbers()
		pNums[0] = 99
		assert.Equal(t, []int{0, 3, 7}, scFromNums.PrimeNumbers())

		pClasses := scFromNums.PrimeClasses()
		pClasses[0] = pitchclass.MustNew(11, 12)
		assert.Equal(t, 0, scFromNums.PrimeClasses()[0].Number())

		pBig := scFromNums.PrimeMaskBig()
		pBig.SetInt64(999999)
		assert.Equal(t, int64(137), scFromNums.PrimeMaskBig().Int64())

		members := scFromNums.MustMembers()
		members[0] = pcset.MustEmpty(12)
		assert.Equal(t, 3, scFromNums.MustMembers()[0].Cardinality())

		tnClasses := scFromNums.TnClasses()
		tnClasses[0] = setclass.SetClass{}
		assert.True(t, scFromNums.TnClasses()[0].IsValid())

		tnInClasses := scFromNums.MustTnInClasses()
		tnInClasses[0] = setclass.SetClass{}
		assert.True(t, scFromNums.MustTnInClasses()[0].IsValid())

		catTnTypes := scFromNums.MustCatalogTnTypes()
		catTnTypes[0] = catalog.TnType{}
		assert.True(t, scFromNums.MustCatalogTnTypes()[0].IsValid())

		sc4Z15 := setclass.MustFromForteNotation("4-Z15", setclass.TnIn)
		zParts := sc4Z15.MustZPartners()
		zParts[0] = setclass.SetClass{}
		assert.True(t, sc4Z15.MustZPartners()[0].IsValid())

		subs := scFromNums.MustSubclasses()
		subs[0] = setclass.SetClass{}
		assert.True(t, scFromNums.MustSubclasses()[0].IsValid())

		prof := scFromNums.MustSubclassProfile()
		prof[0].Count = 999
		assert.Equal(t, 1, scFromNums.MustSubclassProfile()[0].Count)

		cntBig := setclass.MustCountBig(12, setclass.TnIn)
		cntBig.SetInt64(0)
		assert.Equal(t, int64(224), setclass.MustCountBig(12, setclass.TnIn).Int64())
	})

	t.Run("Concurrent_access_from_multiple_goroutines_NF3", func(t *testing.T) {
		t.Parallel()

		sc := setclass.MustFromForteNotation("6-Z29", setclass.TnIn)
		super := setclass.MustFromForteNotation("8-26", setclass.TnIn)
		var wg sync.WaitGroup

		for range 8 {
			wg.Go(func() {
				for range 20 {
					_ = sc.PrimeForm()
					_ = sc.FortePrimeForm()
					_ = sc.MustCollinsMask()
					_ = sc.MustMembers()
					_ = sc.Complement()
					_ = sc.TnClasses()
					_ = sc.MustZPartners()
					_ = sc.MustEmbeddingCount(super)
					_ = sc.MustFormatMorris()
					_ = setclass.MustCount(12, setclass.TTO)
				}
			})
		}
		wg.Wait()
	})
}
