package setclass_test

import (
	"errors"
	"fmt"

	"github.com/go-muse/musex/pcset"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/setclass"
)

func ExampleSetClass() {
	// C major {0, 4, 7} and A minor {9, 0, 4} in Z_12:
	cMaj := pcset.MustNew(12, 0, 4, 7)
	aMin := pcset.MustNew(12, 9, 0, 4)

	// Under Tn equivalence these are distinct directed Tn-types (3-11B and 3-11A):
	tnMaj := setclass.MustNewTn(cMaj)
	tnMin := setclass.MustNewTn(aMin)
	fmt.Println("Tn equal:", tnMaj.Equal(tnMin))
	fmt.Println("Tn types:", tnMaj.MustTnType().String(), tnMin.MustTnType().String())

	// Under TnIn equivalence they belong to the same Forte class 3-11 with prime form (0, 3, 7):
	tnInMaj := tnMaj.ToTnIn()
	tnInMin := tnMin.ToTnIn()
	fmt.Println("TnIn equal:", tnInMaj.Equal(tnInMin))
	fmt.Println("Canonical:", tnInMaj.String())
	fmt.Println("Morris:", tnInMaj.MustFormatMorris())
	fmt.Println("Literary:", tnInMaj.MustNotation(pitchclass.TE))
	fmt.Println("Compact:", tnInMaj.MustCompactNotation(pitchclass.TE))

	// Compare PrimeMask and CollinsMask for class 4-10 (prime form (0, 2, 3, 5)):
	sc4_10 := setclass.MustFromForteNotation("4-10", setclass.TnIn)
	fmt.Println("4-10 PrimeMask:", sc4_10.MustPrimeMask(), "CollinsMask:", sc4_10.MustCollinsMask())

	// Embedding count of triad class 3-11 in diatonic heptachord 7-35:
	diatonic := setclass.MustFromForteNotation("7-35", setclass.TnIn)
	fmt.Println("EMB(3-11, 7-35):", tnInMaj.MustEmbeddingCount(diatonic))
	// Output:
	// Tn equal: false
	// Tn types: 3-11B 3-11A
	// TnIn equal: true
	// Canonical: TnIn(0, 3, 7) mod 12
	// Morris: 3-11[037]
	// Literary: (0, 3, 7)
	// Compact: (037)
	// 4-10 PrimeMask: 45 CollinsMask: 3330
	// EMB(3-11, 7-35): 6
}

func ExampleSetClass_errors() {
	// 1. The zero value of SetClass is invalid: non-error getters return zero values,
	// whereas error-returning methods return ErrInvalidValue:
	var zero setclass.SetClass
	fmt.Println("Zero valid:", zero.IsValid(), "PrimeForm valid:", zero.PrimeForm().IsValid())
	_, err := zero.PrimeMask()
	fmt.Println("Zero PrimeMask error:", errors.Is(err, setclass.ErrInvalidValue))

	// 2. Exceeding the TTO modulus limit (N > 256):
	_, err = setclass.Empty(300, setclass.TTO)
	fmt.Println("TTO modulus > 256:", errors.Is(err, setclass.ErrModulusTooLarge))

	// 3. Exceeding the combinatorial enumeration limit C(64, 10) > 200_000:
	_, err = setclass.OfCardinality(64, 10, setclass.TnIn)
	fmt.Println("Enumeration limit:", errors.Is(err, setclass.ErrEnumerationLimitExceeded))

	// 4. Non-invertible multiplier m = 6 modulo 12 (gcd(6, 12) = 6 > 1):
	sc := setclass.MustFromForteNotation("3-11", setclass.TnIn)
	_, err = sc.Multiply(6)
	fmt.Println("Non-invertible multiplier:", errors.Is(err, setclass.ErrNonInvertibleMultiplier))

	// 5. Mismatch between pitch-class set and catalog label in Morris notation:
	_, err = setclass.ParseMorris("3-11[012]")
	fmt.Println("Morris mismatch:", errors.Is(err, setclass.ErrNotation))
	// Output:
	// Zero valid: false PrimeForm valid: false
	// Zero PrimeMask error: true
	// TTO modulus > 256: true
	// Enumeration limit: true
	// Non-invertible multiplier: true
	// Morris mismatch: true
}

func ExampleSetClass_PrimeFormWith() {
	// Forte class 5-20 has different prime forms under Rahn (1980) and Forte (1973),
	// yet both represent the exact same SetClass:
	fromRahn := setclass.MustFromNumbers(12, setclass.TnIn, 0, 1, 5, 6, 8)
	fromForte := setclass.MustFromNumbers(12, setclass.TnIn, 0, 1, 3, 7, 8)

	fmt.Println("Equal:", fromRahn.Equal(fromForte))
	fmt.Println("RahnForteDiffer:", fromRahn.RahnForteDiffer())
	fmt.Println("Rahn:", fromRahn.MustNotationWith(pitchclass.TE, setclass.Rahn))
	fmt.Println("Forte:", fromRahn.MustNotationWith(pitchclass.TE, setclass.Forte))
	// Output:
	// Equal: true
	// RahnForteDiffer: true
	// Rahn: (0, 1, 5, 6, 8)
	// Forte: (0, 1, 3, 7, 8)
}

func ExampleSetClass_ZPartners() {
	// All-interval tetrachord 4-Z15 in TnIn, Tn, and TTO:
	tnIn := setclass.MustFromForteNotation("4-Z15", setclass.TnIn)
	zPart, ok := tnIn.MustZPartner()
	fmt.Println("TnIn ZPartner:", ok, zPart.MustCatalogName())

	// In directed Tn, 4-Z15A has two Z-partners: 4-Z29A and 4-Z29B:
	tnA := setclass.MustFromTnTypeNotation("4-Z15A")
	for _, p := range tnA.MustZPartners() {
		fmt.Println("Tn Z-partner of 4-Z15A:", p.MustCatalogName())
	}

	// In TTO, 4-Z15 and 4-Z29 merge under M5 into 4-Z15/Z29, so HasZPartner is false:
	tto := tnIn.MustToTTO()
	fmt.Println("TTO CatalogName:", tto.MustCatalogName(), "HasZPartner:", tto.MustHasZPartner())
	// Output:
	// TnIn ZPartner: true 4-Z29
	// Tn Z-partner of 4-Z15A: 4-Z29A
	// Tn Z-partner of 4-Z15A: 4-Z29B
	// TTO CatalogName: 4-Z15/Z29 HasZPartner: false
}

func ExampleCount() {
	// Exact Burnside-Polya set-class counts in Z_12 and Z_24:
	fmt.Println("N=12:",
		setclass.MustCount(12, setclass.Tn),
		setclass.MustCount(12, setclass.TnIn),
		setclass.MustCount(12, setclass.TTO),
	)
	fmt.Println("N=24:",
		setclass.MustCount(24, setclass.Tn),
		setclass.MustCount(24, setclass.TnIn),
		setclass.MustCount(24, setclass.TTO),
	)
	// Output:
	// N=12: 352 224 158
	// N=24: 699252 352698 94484
}
