package setclass_test

import (
	"math/bits"
	"testing"

	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/setclass"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func FuzzParseCanonicalRoundTrip(f *testing.F) {
	f.Add(12, uint8(setclass.TnIn), uint64(0b000010010001), "TnIn(0, 3, 7) mod 12")
	f.Add(12, uint8(setclass.Tn), uint64(0b000010010001), "Tn(0, 4, 7) mod 12")
	f.Add(12, uint8(setclass.TTO), uint64(0b000000000111), "TTO(0, 1, 2) mod 12")
	f.Add(1, uint8(setclass.TnIn), uint64(0), "TnIn() mod 1")
	f.Add(19, uint8(setclass.Tn), uint64(0b10101), "TnIn(0, 4, 7) mod 12")

	f.Fuzz(func(t *testing.T, modulus int, kindRaw uint8, mask uint64, rawStr string) {
		// Arbitrary string parsing must never panic.
		if parsed, err := setclass.Parse(rawStr); err == nil {
			assert.True(t, parsed.IsValid())
			assert.Equal(t, rawStr, parsed.String())
		}

		kinds := setclass.Kinds()
		kind := kinds[int(kindRaw)%len(kinds)]
		maxMod := 64
		n := (int(uint(modulus)%uint(maxMod)) + 1)
		if n < 64 {
			mask &= (uint64(1) << n) - 1
		}

		sc, err := setclass.FromBitmask(n, kind, mask)
		require.NoError(t, err)

		canon := sc.String()
		assert.Equal(t, canon, sc.Canonical())
		roundTrip, err := setclass.Parse(canon)
		require.NoError(t, err)
		assert.Equal(t, sc, roundTrip)
	})
}

func FuzzParseNotationRoundTrip(f *testing.F) {
	f.Add(10, uint8(setclass.TnIn), uint8(pitchclass.Decimal), uint64(0b010010001), "(0, 3, 7)")
	f.Add(12, uint8(setclass.Tn), uint8(pitchclass.TE), uint64(0b110000010001), "(0, 4, T, E)")
	f.Add(12, uint8(setclass.TTO), uint8(pitchclass.LowerTE), uint64(0b10100001011), "(01378)")
	f.Add(12, uint8(setclass.TnIn), uint8(pitchclass.AB), uint64(0), "()")

	styles := []pitchclass.Style{
		pitchclass.Decimal,
		pitchclass.TE,
		pitchclass.LowerTE,
		pitchclass.AB,
	}

	f.Fuzz(func(t *testing.T, modulus int, kindRaw uint8, styleRaw uint8, mask uint64, rawStr string) {
		kinds := setclass.Kinds()
		kind := kinds[int(kindRaw)%len(kinds)]
		style := styles[int(styleRaw)%len(styles)]
		n := (int(uint(modulus)%12) + 1)
		mask &= (uint64(1) << n) - 1

		// Arbitrary string parsing must never panic.
		_, _ = setclass.ParseNotation(rawStr, style, n, kind)

		sc, err := setclass.FromBitmask(n, kind, mask)
		require.NoError(t, err)

		for _, alg := range setclass.Algorithms() {
			lit, err := sc.NotationWith(style, alg)
			require.NoError(t, err)
			parsedLit, err := setclass.ParseNotation(lit, style, n, kind)
			require.NoError(t, err)
			assert.Equal(t, sc, parsedLit)

			if style != pitchclass.Decimal || n <= 10 {
				comp, err := sc.CompactNotationWith(style, alg)
				require.NoError(t, err)
				parsedComp, err := setclass.ParseNotation(comp, style, n, kind)
				require.NoError(t, err)
				assert.Equal(t, sc, parsedComp)
			}
		}
	})
}

func FuzzParseMorrisRoundTrip(f *testing.F) {
	f.Add(uint8(setclass.TnIn), uint16(0b000010010001), "3-11[037]")
	f.Add(uint8(setclass.Tn), uint16(0b000010010001), "3-11B[047]")
	f.Add(uint8(setclass.TTO), uint16(0b000000111111), "6-1/32[012345]")
	f.Add(uint8(setclass.TnIn), uint16(0), "3-11[012]")

	f.Fuzz(func(t *testing.T, kindRaw uint8, mask12 uint16, rawStr string) {
		// Arbitrary input must never panic.
		_, _ = setclass.ParseMorris(rawStr)

		kinds := setclass.Kinds()
		kind := kinds[int(kindRaw)%len(kinds)]
		mask := uint64(mask12 & 0xFFF)

		sc, err := setclass.FromBitmask(12, kind, mask)
		require.NoError(t, err)

		for _, alg := range setclass.Algorithms() {
			morris, err := sc.FormatMorrisWith(alg)
			require.NoError(t, err)

			parsedWithKind, err := setclass.ParseMorrisWithKind(morris, kind)
			require.NoError(t, err)
			assert.Equal(t, sc, parsedWithKind)
		}
	})
}

func FuzzOrbitAndAlgebraicInvariants(f *testing.F) {
	f.Add(12, uint8(setclass.TnIn), uint32(0b000010010001))
	f.Add(12, uint8(setclass.Tn), uint32(0b000101100101))
	f.Add(12, uint8(setclass.TTO), uint32(0b000100100101))
	f.Add(1, uint8(setclass.TnIn), uint32(1))
	f.Add(16, uint8(setclass.TnIn), uint32(0b00101101))

	f.Fuzz(func(t *testing.T, modulus int, kindRaw uint8, mask32 uint32) {
		kinds := setclass.Kinds()
		kind := kinds[int(kindRaw)%len(kinds)]
		n := (int(uint(modulus)%16) + 1)
		mask := uint64(mask32) & ((uint64(1) << n) - 1)

		sc, err := setclass.FromBitmask(n, kind, mask)
		require.NoError(t, err)

		// Idempotence of prime form reduction under both Rahn and Forte.
		fromRahn, err := setclass.New(sc.PrimeForm(), kind)
		require.NoError(t, err)
		assert.Equal(t, sc, fromRahn)

		fromForte, err := setclass.New(sc.FortePrimeForm(), kind)
		require.NoError(t, err)
		assert.Equal(t, sc, fromForte)

		// Involution of Complement and Inversion.
		assert.Equal(t, sc, sc.Complement().Complement())
		assert.Equal(t, n-sc.Cardinality(), sc.Complement().Cardinality())
		assert.Equal(t, sc, sc.Inversion().Inversion())

		// IntervalVector total equals k*(k-1)/2.
		k := sc.Cardinality()
		assert.Equal(t, k*(k-1)/2, sc.IntervalVector().Total())

		// Orbit-stabilizer theorem and Members() consistency.
		members, err := sc.Members()
		require.NoError(t, err)
		assert.Len(t, members, sc.OrbitSize())

		switch kind {
		case setclass.Tn:
			assert.Equal(t, n, sc.OrbitSize()*sc.TnSymmetry())
		case setclass.TnIn:
			groupOrder := 2 * n
			if n <= 2 {
				groupOrder = n
			}
			assert.Equal(t, groupOrder, sc.OrbitSize()*sc.SymmetryDegreeTI())
		case setclass.TTO:
			tot := pitchclass.MustTotient(n)
			assert.Equal(t, n*tot, sc.OrbitSize()*sc.MustSymmetryDegreeTTO())
		}

		// CollinsMask equals maximum bitmask over Members().
		cm, err := sc.CollinsMask()
		require.NoError(t, err)
		assert.Equal(t, k, bits.OnesCount64(cm))
		var maxMember uint64
		for _, m := range members {
			maxMember = max(maxMember, m.MustBitmask())
		}
		assert.Equal(t, maxMember, cm)
	})
}

func FuzzCatalogBijectionMod12(f *testing.F) {
	f.Add(uint8(setclass.TnIn), uint16(0b000010010001))
	f.Add(uint8(setclass.Tn), uint16(0b000101100011))
	f.Add(uint8(setclass.TTO), uint16(0b000101011101))

	f.Fuzz(func(t *testing.T, kindRaw uint8, mask12 uint16) {
		kinds := setclass.Kinds()
		kind := kinds[int(kindRaw)%len(kinds)]
		mask := uint64(mask12 & 0xFFF)

		sc, err := setclass.FromBitmask(12, kind, mask)
		require.NoError(t, err)

		fc, err := sc.ForteClass()
		require.NoError(t, err)
		assert.True(t, fc.IsValid())

		name, err := sc.CatalogName()
		require.NoError(t, err)
		fromName, err := setclass.FromCatalogName(name, kind)
		require.NoError(t, err)
		assert.Equal(t, sc, fromName)

		catTnTypes, err := sc.CatalogTnTypes()
		require.NoError(t, err)
		tnClasses := sc.TnClasses()
		require.Len(t, catTnTypes, len(tnClasses))
		for i, subTn := range tnClasses {
			assert.Equal(t, catTnTypes[i], subTn.MustTnType())
		}
	})
}

func FuzzEmbeddingAndBurnsideIdentities(f *testing.F) {
	f.Add(8, uint8(setclass.TnIn), uint16(0b01011011), 3)
	f.Add(10, uint8(setclass.Tn), uint16(0b0101011011), 2)
	f.Add(12, uint8(setclass.TTO), uint16(0b000100100101), 2)

	f.Fuzz(func(t *testing.T, modulus int, kindRaw uint8, mask16 uint16, subCard int) {
		kinds := setclass.Kinds()
		kind := kinds[int(kindRaw)%len(kinds)]
		n := (int(uint(modulus)%10) + 1)
		mask := uint64(mask16) & ((uint64(1) << n) - 1)

		sc, err := setclass.FromBitmask(n, kind, mask)
		require.NoError(t, err)

		yCard := sc.Cardinality()
		k := int(uint(subCard) % uint(yCard+1))

		// Profile sum for cardinality k equals binomial(yCard, k).
		prof, err := sc.SubclassProfileOfCardinality(k)
		require.NoError(t, err)
		sumCount := 0
		for _, entry := range prof {
			assert.Equal(t, entry.Count, entry.Subclass.MustEmbeddingCount(sc))
			sumCount += entry.Count
		}
		binom := 1
		for i := 1; i <= k; i++ {
			binom = binom * (yCard - i + 1) / i
		}
		assert.Equal(t, binom, sumCount)

		// OfCardinality length matches Burnside CountOfCardinality.
		classesK, err := setclass.OfCardinality(n, k, kind)
		require.NoError(t, err)
		assert.Len(t, classesK, setclass.MustCountOfCardinality(n, k, kind))
	})
}
