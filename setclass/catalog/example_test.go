package catalog_test

import (
	"fmt"

	"github.com/go-muse/musex/setclass/catalog"
)

func ExampleParseForteClass() {
	fc, err := catalog.ParseForteClass("4-Z15")
	if err != nil {
		panic(err)
	}
	fmt.Println(fc.String(), fc.Cardinality(), fc.Ordinal(), fc.IsZ())
	// Output: 4-Z15 4 15 true
}

func ExampleParseTnType() {
	tnA := catalog.MustParseTnType("3-11A")
	tnB := tnA.Inversion()
	fmt.Println(tnA.String(), tnA.PrimeForm(), tnB.String(), tnB.PrimeForm())
	// Output: 3-11A [0 3 7] 3-11B [0 4 7]
}

func ExampleForteClass_PrimeForm() {
	fc := catalog.MustParseForteClass("5-20")
	fmt.Println("differ:", fc.RahnForteDiffer())
	fmt.Println("rahn:", fc.PrimeForm(), fc.PrimeMask())
	fmt.Println("forte:", fc.FortePrimeForm(), fc.FortePrimeMask())
	// Output:
	// differ: true
	// rahn: [0 1 5 6 8] 355
	// forte: [0 1 3 7 8] 395
}

func ExampleForteClass_Complement() {
	triad := catalog.MustParseForteClass("3-11")
	hex := catalog.MustParseForteClass("6-14")
	zHex := catalog.MustParseForteClass("6-Z29")
	fmt.Println(triad.Complement().String(), triad.IsSelfComplementary())
	fmt.Println(hex.Complement().String(), hex.IsSelfComplementary())
	fmt.Println(zHex.Complement().String(), zHex.IsSelfComplementary())
	// Output:
	// 9-11 false
	// 6-14 true
	// 6-Z50 false
}

func ExampleForteClass_ZPartner() {
	fc := catalog.MustParseForteClass("4-Z15")
	zp, ok := fc.ZPartner()
	pair, _ := fc.ZPair()
	fmt.Println(ok, zp.String(), pair.String(), pair.IntervalVector())
	// Output: true 4-Z29 4-Z15/4-Z29 [1 1 1 1 1 1]
}

func ExampleForteClass_M5Partner() {
	chromatic := catalog.MustParseForteClass("5-1")
	diatonic := chromatic.M5Partner()
	fmt.Println(diatonic.String(), chromatic.IsM5Invariant())
	// Output: 5-35 false
}

func ExampleCarterByNumber() {
	cc := catalog.MustCarterByNumber(3, 6)
	fc := cc.ForteClass()
	num, ok := fc.CarterNumber()
	fmt.Println(cc.String(), fc.String(), ok, num)
	// Output: Carter(3, 6): 3-11 3-11 true 6
}

func ExampleTnType_Complement() {
	tn614A := catalog.MustParseTnType("6-14A")
	tn630A := catalog.MustParseTnType("6-30A")
	fmt.Println(tn614A.Complement().String(), tn614A.IsSelfComplementary())
	fmt.Println(tn630A.Complement().String(), tn630A.IsSelfComplementary())
	// Output:
	// 6-14A true
	// 6-30B false
}
