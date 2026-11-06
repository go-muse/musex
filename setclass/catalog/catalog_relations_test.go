package catalog_test

import (
	"sync"
	"testing"

	"github.com/go-muse/musex/setclass/catalog"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestForteClass_Complement_And_SelfComplementary(t *testing.T) {
	t.Parallel()

	rows := parseForteCatalogCSV(t)
	all := catalog.ForteClasses()
	require.Len(t, all, 224)

	var selfCompLabels []string

	for idx, row := range rows {
		fc := all[idx]
		comp := fc.Complement()
		require.True(t, comp.IsValid())
		assert.Equal(t, row.Complement, comp.String(), "Complement of %s", row.Forte)
		assert.Equal(t, 12-fc.Cardinality(), comp.Cardinality())
		assert.True(t, comp.Complement().Equal(fc), "Complement must be an involution for %s", row.Forte)
		assert.Equal(t, comp.Equal(fc), fc.IsSelfComplementary())

		if fc.IsSelfComplementary() {
			selfCompLabels = append(selfCompLabels, fc.String())
		}

		// Generalized Babbitt hexachord theorem for interval vectors
		icv := fc.IntervalVector()
		compICV := comp.IntervalVector()
		c := fc.Cardinality()
		for k := 0; k < 5; k++ {
			assert.Equal(t, icv[k]+12-2*c, compICV[k], "Babbitt theorem ic=%d for %s", k+1, row.Forte)
		}
		assert.Equal(t, icv[5]+6-c, compICV[5], "Babbitt theorem ic=6 for %s", row.Forte)

		if c == 6 {
			assert.Equal(t, icv, compICV, "hexachord %s and its complement must share ICV", row.Forte)
			if fc.IsZ() {
				zp, ok := fc.ZPartner()
				require.True(t, ok)
				assert.False(t, fc.IsSelfComplementary(), "Z-hexachord %s is not self-complementary", row.Forte)
				assert.True(t, comp.Equal(zp), "Z-hexachord %s complement must equal ZPartner %s", row.Forte, zp)
			} else {
				assert.True(t, fc.IsSelfComplementary(), "non-Z hexachord %s must be self-complementary", row.Forte)
			}
		}
	}

	wantSelfComp20 := []string{
		"6-1", "6-2", "6-5", "6-7", "6-8",
		"6-9", "6-14", "6-15", "6-16", "6-18",
		"6-20", "6-21", "6-22", "6-27", "6-30",
		"6-31", "6-32", "6-33", "6-34", "6-35",
	}
	assert.Equal(t, wantSelfComp20, selfCompLabels)
}

func TestTnType_Complement_Inversion_And_SelfComplementary(t *testing.T) {
	t.Parallel()

	rows := parseTnTypesCatalogCSV(t)
	all := catalog.TnTypes()
	require.Len(t, all, 352)

	var selfCompTnLabels []string
	var sameSuffixChiralA []string
	swappedSuffixChiralACount := 0

	for idx, row := range rows {
		tn := all[idx]
		comp := tn.Complement()
		require.True(t, comp.IsValid())
		assert.Equal(t, row.Complement, comp.String(), "Complement of TnType %s", row.TnType)
		assert.Equal(t, 12-tn.Cardinality(), comp.Cardinality())
		assert.True(t, comp.Complement().Equal(tn), "TnType.Complement must be an involution for %s", row.TnType)
		assert.True(t, comp.ForteClass().Equal(tn.ForteClass().Complement()))
		assert.Equal(t, comp.Equal(tn), tn.IsSelfComplementary())

		if tn.IsSelfComplementary() {
			selfCompTnLabels = append(selfCompTnLabels, tn.String())
		}

		if tn.Suffix() == catalog.SuffixA {
			if comp.Suffix() == catalog.SuffixA {
				sameSuffixChiralA = append(sameSuffixChiralA, tn.ForteClass().String())
			} else {
				assert.Equal(t, catalog.SuffixB, comp.Suffix())
				swappedSuffixChiralACount++
			}
		}

		// Inversion check
		inv := tn.Inversion()
		require.True(t, inv.IsValid())
		assert.True(t, inv.Inversion().Equal(tn), "TnType.Inversion must be an involution for %s", row.TnType)
		assert.True(t, inv.ForteClass().Equal(tn.ForteClass()))
		switch tn.Suffix() {
		case catalog.SuffixNone:
			assert.True(t, inv.Equal(tn), "achiral TnType %s must equal its inversion", row.TnType)
		case catalog.SuffixA:
			tnB, ok := tn.ForteClass().TnTypeB()
			require.True(t, ok)
			assert.True(t, inv.Equal(tnB))
		case catalog.SuffixB:
			assert.True(t, inv.Equal(tn.ForteClass().TnTypeA()))
		}
	}

	wantSelfComp8 := []string{
		"6-1", "6-7", "6-8", "6-14A", "6-14B", "6-20", "6-32", "6-35",
	}
	assert.Equal(t, wantSelfComp8, selfCompTnLabels)

	// Explicitly check 6-30A <-> 6-30B (chiral hexachord that swaps branches under complement)
	t630A := catalog.MustParseTnType("6-30A")
	t630B := catalog.MustParseTnType("6-30B")
	assert.False(t, t630A.IsSelfComplementary())
	assert.False(t, t630B.IsSelfComplementary())
	assert.Equal(t, t630B, t630A.Complement())
	assert.Equal(t, t630A, t630B.Complement())

	// Exact 13 chiral ForteClasses preserving suffix A <-> A and B <-> B under TnType.Complement (Г-6, К-14)
	wantSameSuffix13 := []string{
		"4-12", "4-14",
		"5-11", "5-26", "5-28",
		"6-Z10", "6-14", "6-Z39",
		"7-11", "7-26", "7-28",
		"8-12", "8-14",
	}
	assert.Equal(t, wantSameSuffix13, sameSuffixChiralA)
	assert.Equal(t, 115, swappedSuffixChiralACount)
}

func TestZRelations_And_ZPairs(t *testing.T) {
	t.Parallel()

	rows := parseForteCatalogCSV(t)
	all := catalog.ForteClasses()
	require.Len(t, all, 224)

	zClassCount := 0
	nonZClassCount := 0

	for idx, row := range rows {
		fc := all[idx]
		zp, ok := fc.ZPartner()
		zpair, pairOk := fc.ZPair()

		if row.IsZ {
			zClassCount++
			require.True(t, ok, "%s must have ZPartner", row.Forte)
			require.True(t, pairOk, "%s must have ZPair", row.Forte)
			assert.Equal(t, row.ZPartner, zp.String())
			assert.False(t, zp.Equal(fc))
			assert.Equal(t, fc.Cardinality(), zp.Cardinality())
			assert.Equal(t, fc.IntervalVector(), zp.IntervalVector())
			assert.NotEqual(t, fc.PrimeMask(), zp.PrimeMask())

			zpBack, backOk := zp.ZPartner()
			require.True(t, backOk)
			assert.True(t, zpBack.Equal(fc))

			assert.True(t, zpair.IsValid())
			assert.True(t, zpair.First().Equal(fc) || zpair.Second().Equal(fc))
			assert.True(t, zpair.First().Equal(zp) || zpair.Second().Equal(zp))
		} else {
			nonZClassCount++
			assert.False(t, ok, "non-Z class %s must not have ZPartner", row.Forte)
			assert.False(t, zp.IsValid())
			assert.False(t, pairOk, "non-Z class %s must not have ZPair", row.Forte)
			assert.False(t, zpair.IsValid())
		}
	}

	assert.Equal(t, 46, zClassCount)
	assert.Equal(t, 178, nonZClassCount)

	// Verify all 23 ZPairs
	pairs := catalog.ZPairs()
	require.Len(t, pairs, 23)
	for _, zp := range pairs {
		require.True(t, zp.IsValid())
		first := zp.First()
		second := zp.Second()
		require.True(t, first.IsValid())
		require.True(t, second.IsValid())
		assert.Equal(t, -1, first.Compare(second))
		assert.Less(t, first.Ordinal(), second.Ordinal())
		assert.Equal(t, first.Cardinality(), second.Cardinality())
		assert.Equal(t, first.Cardinality(), zp.Cardinality())
		assert.Equal(t, first.IntervalVector(), second.IntervalVector())
		assert.Equal(t, first.IntervalVector(), zp.IntervalVector())
		assert.Equal(t, first.IntervalVectorSlice(), zp.IntervalVectorSlice())
		assert.Equal(t, first.String()+"/"+second.String(), zp.String())

		// Mutation protection on ZPair.IntervalVectorSlice
		s := zp.IntervalVectorSlice()
		orig0 := s[0]
		s[0] = 999
		assert.Equal(t, orig0, zp.IntervalVectorSlice()[0])
	}

	// Asymmetric chirality Z-pairs: 5-Z12/5-Z36 and 7-Z12/7-Z36 (Г-4)
	c5Z12 := catalog.MustParseForteClass("5-Z12")
	c5Z36 := catalog.MustParseForteClass("5-Z36")
	assert.True(t, c5Z12.IsInversionallySymmetric())
	assert.Equal(t, 1, c5Z12.TnTypeCount())
	assert.Equal(t, 12, c5Z12.OrbitSize())
	assert.True(t, c5Z36.IsChiral())
	assert.Equal(t, 2, c5Z36.TnTypeCount())
	assert.Equal(t, 24, c5Z36.OrbitSize())

	c7Z12 := catalog.MustParseForteClass("7-Z12")
	c7Z36 := catalog.MustParseForteClass("7-Z36")
	assert.True(t, c7Z12.IsInversionallySymmetric())
	assert.Equal(t, 1, c7Z12.TnTypeCount())
	assert.Equal(t, 12, c7Z12.OrbitSize())
	assert.True(t, c7Z36.IsChiral())
	assert.Equal(t, 2, c7Z36.TnTypeCount())
	assert.Equal(t, 24, c7Z36.OrbitSize())

	// 4 historical Forte packing exception Z-pairs (Г-7):
	// 6-Z28/6-Z49, 7-Z12/7-Z36, 7-Z17/7-Z37, 7-Z18/7-Z38
	for _, pairLabels := range [][2]string{
		{"6-Z28", "6-Z49"},
		{"7-Z12", "7-Z36"},
		{"7-Z17", "7-Z37"},
		{"7-Z18", "7-Z38"},
	} {
		a := catalog.MustParseForteClass(pairLabels[0])
		b := catalog.MustParseForteClass(pairLabels[1])
		zp, ok := a.ZPartner()
		require.True(t, ok)
		assert.Equal(t, b, zp)
		assert.GreaterOrEqual(t, a.MinSpan(), b.MinSpan())
	}

	// Verify TnType.ZPartner on all 352 TnTypes
	for _, tn := range catalog.TnTypes() {
		zp, ok := tn.ZPartner()
		fcZP, fcOk := tn.ForteClass().ZPartner()
		assert.Equal(t, fcOk, ok)
		assert.Equal(t, fcZP, zp)
	}
}

func TestM5Partner_And_IsM5Invariant(t *testing.T) {
	t.Parallel()

	rows := parseForteCatalogCSV(t)
	all := catalog.ForteClasses()
	require.Len(t, all, 224)

	m5InvariantCount := 0
	m5NonInvariantCount := 0

	for idx, row := range rows {
		fc := all[idx]
		m5 := fc.M5Partner()
		require.True(t, m5.IsValid())
		assert.Equal(t, row.M5Image, m5.String(), "M5Partner of %s", row.Forte)
		assert.Equal(t, fc.Cardinality(), m5.Cardinality())
		assert.True(t, m5.M5Partner().Equal(fc), "M5Partner must be an involution for %s", row.Forte)
		assert.Equal(t, m5.Equal(fc), fc.IsM5Invariant())

		if fc.IsM5Invariant() {
			m5InvariantCount++
		} else {
			m5NonInvariantCount++
		}
	}

	assert.Equal(t, 92, m5InvariantCount, "exact 92 M5-invariant ForteClasses")
	assert.Equal(t, 132, m5NonInvariantCount, "exact 132 M5-non-invariant ForteClasses in 66 pairs")

	spotPairs := [][2]string{
		{"2-1", "2-5"},
		{"3-1", "3-9"},
		{"5-1", "5-35"},
		{"6-1", "6-32"},
		{"7-1", "7-35"},
	}
	for _, p := range spotPairs {
		a := catalog.MustParseForteClass(p[0])
		b := catalog.MustParseForteClass(p[1])
		assert.Equal(t, b, a.M5Partner())
		assert.Equal(t, a, b.M5Partner())
		assert.False(t, a.IsM5Invariant())
		assert.False(t, b.IsM5Invariant())
	}

	for _, tn := range catalog.TnTypes() {
		assert.Equal(t, tn.ForteClass().M5Partner(), tn.M5Partner())
		assert.Equal(t, tn.ForteClass().IsM5Invariant(), tn.IsM5Invariant())
	}
}

func TestForteClass_TnType_Bijection_And_Carter_Bijection(t *testing.T) {
	t.Parallel()

	rows := parseForteCatalogCSV(t)
	all := catalog.ForteClasses()
	require.Len(t, all, 224)

	carterCount := 0
	for idx, row := range rows {
		fc := all[idx]
		tns := fc.TnTypes()
		assert.Len(t, tns, row.TnTypesCount)

		tnA := fc.TnTypeA()
		require.True(t, tnA.IsValid())
		assert.True(t, tnA.ForteClass().Equal(fc))
		assert.Equal(t, fc.PrimeForm(), tnA.PrimeForm())
		assert.Equal(t, fc.FortePrimeForm(), tnA.FortePrimeForm())
		assert.Equal(t, fc.PrimeMask(), tnA.PrimeMask())
		assert.Equal(t, fc.FortePrimeMask(), tnA.FortePrimeMask())
		assert.Equal(t, tns[0], tnA)

		tnB, hasB := fc.TnTypeB()
		if fc.IsChiral() {
			require.True(t, hasB)
			require.True(t, tnB.IsValid())
			assert.True(t, tnB.ForteClass().Equal(fc))
			assert.Equal(t, catalog.SuffixA, tnA.Suffix())
			assert.Equal(t, catalog.SuffixB, tnB.Suffix())
			assert.Equal(t, tns[1], tnB)
		} else {
			assert.False(t, hasB)
			assert.False(t, tnB.IsValid())
			assert.Equal(t, catalog.SuffixNone, tnA.Suffix())
		}

		// Carter catalog check against CSV row
		assert.Equal(t, row.HasCarter, fc.HasCarter())
		cNum, hasNum := fc.CarterNumber()
		cc, hasCC := fc.Carter()
		assert.Equal(t, row.HasCarter, hasNum)
		assert.Equal(t, row.HasCarter, hasCC)
		if row.HasCarter {
			carterCount++
			assert.Equal(t, row.Carter, cNum)
			require.True(t, cc.IsValid())
			assert.Equal(t, fc.Cardinality(), cc.Cardinality())
			assert.Equal(t, row.Carter, cc.Number())
			assert.True(t, cc.ForteClass().Equal(fc))
			assert.Equal(t, cc, catalog.MustCarterByNumber(fc.Cardinality(), row.Carter))
		} else {
			assert.Equal(t, 0, cNum)
			assert.False(t, cc.IsValid())
		}

		for _, tn := range tns {
			assert.Equal(t, fc.HasCarter(), tn.HasCarter())
			tnCNum, tnHasNum := tn.CarterNumber()
			tnCC, tnHasCC := tn.Carter()
			assert.Equal(t, hasNum, tnHasNum)
			assert.Equal(t, cNum, tnCNum)
			assert.Equal(t, hasCC, tnHasCC)
			assert.Equal(t, cc, tnCC)
		}
	}

	assert.Equal(t, 135, carterCount)
}

func TestCanonicalOrdering_Compare_Equal_And_MapKeys(t *testing.T) {
	t.Parallel()

	t.Run("ForteClass_ordering_and_map", func(t *testing.T) {
		t.Parallel()

		all := catalog.ForteClasses()
		seen := make(map[catalog.ForteClass]int, len(all))
		for i, fc := range all {
			seen[fc] = i
			assert.True(t, fc.Equal(fc))
			assert.Equal(t, 0, fc.Compare(fc))
			if i > 0 {
				prev := all[i-1]
				assert.Equal(t, -1, prev.Compare(fc))
				assert.Equal(t, 1, fc.Compare(prev))
				assert.False(t, prev.Equal(fc))
				assert.True(t, prev != fc)
			}
		}
		assert.Len(t, seen, 224)
	})

	t.Run("TnType_ordering_and_map", func(t *testing.T) {
		t.Parallel()

		all := catalog.TnTypes()
		seen := make(map[catalog.TnType]int, len(all))
		for i, tn := range all {
			seen[tn] = i
			assert.True(t, tn.Equal(tn))
			assert.Equal(t, 0, tn.Compare(tn))
			if i > 0 {
				prev := all[i-1]
				assert.Equal(t, -1, prev.Compare(tn))
				assert.Equal(t, 1, tn.Compare(prev))
				assert.False(t, prev.Equal(tn))
				assert.True(t, prev != tn)
			}
		}
		assert.Len(t, seen, 352)
	})

	t.Run("CarterClass_ordering_and_map", func(t *testing.T) {
		t.Parallel()

		all := catalog.CarterClasses()
		seen := make(map[catalog.CarterClass]int, len(all))
		for i, cc := range all {
			seen[cc] = i
			assert.True(t, cc.Equal(cc))
			assert.Equal(t, 0, cc.Compare(cc))
			if i > 0 {
				prev := all[i-1]
				assert.Equal(t, -1, prev.Compare(cc))
				assert.Equal(t, 1, cc.Compare(prev))
				assert.False(t, prev.Equal(cc))
				assert.True(t, prev != cc)
			}
		}
		assert.Len(t, seen, 135)
	})

	t.Run("ZPair_ordering_and_map", func(t *testing.T) {
		t.Parallel()

		all := catalog.ZPairs()
		seen := make(map[catalog.ZPair]int, len(all))
		for i, zp := range all {
			seen[zp] = i
			assert.True(t, zp.Equal(zp))
			assert.Equal(t, 0, zp.Compare(zp))
			if i > 0 {
				prev := all[i-1]
				assert.Equal(t, -1, prev.Compare(zp))
				assert.Equal(t, 1, zp.Compare(prev))
				assert.False(t, prev.Equal(zp))
				assert.True(t, prev != zp)
			}
		}
		assert.Len(t, seen, 23)
	})
}

func TestConcurrentCatalogAccess(t *testing.T) {
	t.Parallel()

	const workers = 8
	var wg sync.WaitGroup
	for w := range workers {
		wg.Go(func() {
			for i := range 20 {
				_ = w
				_ = i
				fc := catalog.MustParseForteClass("6-Z29")
				_ = fc.PrimeForm()
				_ = fc.FortePrimeForm()
				_ = fc.IntervalVector()
				_ = fc.InvarianceVector()
				_ = fc.Complement()
				_, _ = fc.ZPartner()
				_, _ = fc.ZPair()
				_ = fc.M5Partner()
				_, _ = fc.Carter()

				tn := catalog.MustParseTnType("5-20B")
				_ = tn.PrimeForm()
				_ = tn.Inversion()
				_ = tn.Complement()

				_, _ = catalog.ForteClassByAnyPrimeMask(843)
				_, _ = catalog.TnTypeByAnyPrimeMask(419)
				_, _ = catalog.CarterByNumber(3, 6)
				_ = catalog.ForteClasses()
				_ = catalog.TnTypes()
				_ = catalog.CarterClasses()
				_ = catalog.ZPairs()
			}
		})
	}
	wg.Wait()
}

func BenchmarkParseForteClass(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, err := catalog.ParseForteClass("6-Z29")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkForteClassByPrimeMask(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, err := catalog.ForteClassByPrimeMask(137)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkForteClassInvariants(b *testing.B) {
	fc := catalog.MustParseForteClass("6-Z29")
	b.ReportAllocs()
	for b.Loop() {
		_ = fc.Cardinality()
		_ = fc.Ordinal()
		_ = fc.IsZ()
		_ = fc.PrimeMask()
		_ = fc.FortePrimeMask()
		_ = fc.RahnForteDiffer()
		_ = fc.IntervalVector()
		_ = fc.InvarianceVector()
		_ = fc.TnSymmetry()
		_ = fc.InSymmetry()
		_ = fc.DegreeOfSymmetry()
		_ = fc.OrbitSize()
		_ = fc.MinSpan()
		_ = fc.Complement()
		_, _ = fc.ZPartner()
		_ = fc.M5Partner()
	}
}

func BenchmarkTnTypeRelations(b *testing.B) {
	tn := catalog.MustParseTnType("6-14A")
	b.ReportAllocs()
	for b.Loop() {
		_ = tn.ForteClass()
		_ = tn.Suffix()
		_ = tn.PrimeMask()
		_ = tn.Inversion()
		_ = tn.Complement()
		_ = tn.IsSelfComplementary()
		_ = tn.M5Partner()
	}
}
