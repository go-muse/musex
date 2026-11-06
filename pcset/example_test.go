package pcset_test

import (
	"fmt"

	"github.com/go-muse/musex/pcset"
	"github.com/go-muse/musex/pitchclass"
)

func Example() {
	cMaj := pcset.MustNew(12, 0, 4, 7)
	fmt.Println(cMaj)
	fmt.Println(cMaj.Transpose(2))
	fmt.Println(cMaj.Invert(7))
	fmt.Println(cMaj.IntervalVector())
	// Output:
	// {0, 4, 7} mod 12
	// {2, 6, 9} mod 12
	// {0, 3, 7} mod 12
	// <0, 0, 1, 1, 1, 0> mod 12
}

func ExampleNew() {
	a, _ := pcset.New(12, 0, 4, 7)
	b, _ := pcset.New(12, 7, 0, 4, 0)
	empty, _ := pcset.Empty(12)
	univ, _ := pcset.Universe(12)

	fmt.Println(a, a == b)
	fmt.Println(empty, empty.IsEmpty())
	fmt.Println(univ.Cardinality(), univ.IsUniverse())
	// Output:
	// {0, 4, 7} mod 12 true
	// {} mod 12 true
	// 12 true
}

func ExampleReduce() {
	s, _ := pcset.Reduce(12, -1, 12, 16, 7)
	fmt.Println(s)
	// Output:
	// {0, 4, 7, 11} mod 12
}

func ExampleFromClasses() {
	c0 := pitchclass.MustNew(0, 12)
	c4 := pitchclass.MustNew(4, 12)
	c7 := pitchclass.MustNew(7, 12)

	s, _ := pcset.FromClasses(c0, c4, c7, c4)
	empty, _ := pcset.FromClassesWithModulus(12)
	fmt.Println(s)
	fmt.Println(empty)
	// Output:
	// {0, 4, 7} mod 12
	// {} mod 12
}

func ExampleFromBitmask() {
	s, _ := pcset.FromBitmask(145, 12)
	fmt.Println(s, s.MustBitmask())

	large := pcset.MustNew(65, 0, 64)
	maskBig := large.MustBitmaskBig()
	fmt.Println(pcset.MustFromBitmaskBig(maskBig, 65))
	// Output:
	// {0, 4, 7} mod 12 145
	// {0, 64} mod 65
}

func ExampleSet_Transpose() {
	cMaj := pcset.MustNew(12, 0, 4, 7)
	fmt.Println(cMaj.Transpose(2))
	fmt.Println(cMaj.Invert(0))
	fmt.Println(cMaj.Multiply(5))
	// Output:
	// {2, 6, 9} mod 12
	// {0, 5, 8} mod 12
	// {0, 8, 11} mod 12
}

func ExampleSet_Apply() {
	cMaj := pcset.MustNew(12, 0, 4, 7)
	op := pitchclass.MustNewOperator(7, 5, 12)
	fmt.Println(cMaj.MustApply(op))

	i4 := pitchclass.MustInversion(4, 12)
	fmt.Println(pcset.MustOperatorFixedClasses(i4))
	fmt.Println(pcset.MustMultiplicationFixedClasses(5, 12))
	// Output:
	// {5, 6, 9} mod 12
	// {2, 8} mod 12
	// {0, 3, 6, 9} mod 12
}

func ExampleSet_Union() {
	a := pcset.MustNew(12, 0, 4, 7)
	b := pcset.MustNew(12, 4, 7, 11)

	fmt.Println(a.Complement())
	fmt.Println(a.MustUnion(b))
	fmt.Println(a.MustIntersection(b))
	fmt.Println(a.MustDifference(b))
	fmt.Println(a.MustSymmetricDifference(b))
	fmt.Println(a.MustIntersection(b).MustIsSubsetOf(a), a.MustIsDisjoint(b))
	// Output:
	// {1, 2, 3, 5, 6, 8, 9, 10, 11} mod 12
	// {0, 4, 7, 11} mod 12
	// {4, 7} mod 12
	// {0} mod 12
	// {0, 11} mod 12
	// true false
}

func ExampleSet_CoveringTranspositionsOf() {
	majorScale := pcset.MustNew(12, 0, 2, 4, 5, 7, 9, 11)
	cMaj := pcset.MustNew(12, 0, 4, 7)
	dMaj := pcset.MustNew(12, 2, 6, 9)
	cMin := pcset.MustNew(12, 0, 3, 7)

	fmt.Println(cMaj.MustCoveringTranspositionsOf(majorScale))
	fmt.Println(dMaj.MustCoveringTranspositionsOf(majorScale))
	fmt.Println(cMaj.MustTranspositionIndicesTo(dMaj))
	fmt.Println(cMaj.MustInversionIndicesTo(cMin))
	// Output:
	// {0, 5, 7} mod 12
	// {2, 7, 9} mod 12
	// {2} mod 12
	// {7} mod 12
}

func ExampleSet_EquivalenceOperatorsTI() {
	chrom := pcset.MustNew(12, 0, 1, 2)
	quartal := pcset.MustNew(12, 0, 5, 10)
	majorScale := pcset.MustNew(12, 0, 2, 4, 5, 7, 9, 11)

	fmt.Println(chrom.MustEquivalenceOperatorsTI(quartal))
	fmt.Println(chrom.MustEquivalenceOperatorsTTO(quartal))
	fmt.Println(
		majorScale.MustCommonToneCountTranspose(7),
		majorScale.MustCommonToneCountTranspose(6),
	)
	// Output:
	// []
	// [T0M5 mod 12 T10M7 mod 12]
	// 6 2
}

func ExampleSet_NormalForm() {
	s := pcset.MustNew(12, 2, 3, 7, 11)
	fmt.Println(len(s.Rotations()))
	fmt.Println(s.MustNormalFormNotation(pitchclass.Decimal, pcset.Rahn))

	div := pcset.MustNew(12, 0, 1, 3, 7, 8)
	fmt.Println(div.MustNormalFormNotation(pitchclass.Decimal, pcset.Rahn))
	fmt.Println(div.MustNormalFormNotation(pitchclass.Decimal, pcset.Forte))
	// Output:
	// 4
	// [11, 2, 3, 7]
	// [7, 8, 0, 1, 3]
	// [0, 1, 3, 7, 8]
}

func ExampleSet_IntervalVector() {
	majorScale := pcset.MustNew(12, 0, 2, 4, 5, 7, 9, 11)
	iv := majorScale.IntervalVector()
	fmt.Println(iv, iv.MustCompactNotation(), iv.Total())

	z1 := pcset.MustNew(12, 0, 1, 4, 6).IntervalVector()
	z2 := pcset.MustNew(12, 0, 1, 3, 7).IntervalVector()
	fmt.Println(z1 == z2, z1)

	hex := pcset.MustNew(12, 0, 1, 2, 3, 4, 5)
	fmt.Println(hex.IntervalVector() == hex.Complement().IntervalVector())
	// Output:
	// <2, 5, 4, 3, 6, 1> mod 12 <254361> 21
	// true <1, 1, 1, 1, 1, 1> mod 12
	// true
}

func ExampleSet_DegreeOfSymmetryTI() {
	cMaj := pcset.MustNew(12, 0, 4, 7)
	aug := pcset.MustNew(12, 0, 4, 8)
	wholeTone := pcset.MustNew(12, 0, 2, 4, 6, 8, 10)
	chrom := pcset.MustNew(12, 0, 1, 2)

	fmt.Println(cMaj.MustDegreeOfSymmetryTI(), cMaj.MustOrbitSizeTI())
	fmt.Println(aug.MustDegreeOfSymmetryTI(), aug.MustOrbitSizeTI())
	fmt.Println(wholeTone.MustDegreeOfSymmetryTI(), wholeTone.MustOrbitSizeTI())
	fmt.Println(chrom.MustInvarianceVector())
	// Output:
	// 1 24
	// 6 4
	// 12 2
	// <1, 1, 0, 0, 7, 7, 4, 4> mod 12
}

func ExampleAll() {
	count3 := 0
	for range pcset.MustAll(3) {
		count3++
	}

	var dyads []string
	for s := range pcset.MustOfCardinality(4, 2) {
		dyads = append(dyads, s.MustNotation(pitchclass.Decimal))
	}

	fmt.Println(count3)
	fmt.Println(dyads)
	fmt.Println(pcset.MustCount(12), pcset.MustCountOfCardinality(12, 6))
	// Output:
	// 8
	// [{0, 1} {0, 2} {0, 3} {1, 2} {1, 3} {2, 3}]
	// 4096 924
}

func ExampleSet_Notation() {
	s := pcset.MustNew(12, 2, 3, 7, 11)
	fmt.Println(s.MustNotation(pitchclass.Decimal), s.MustNotation(pitchclass.TE))
	fmt.Println(s.MustCompactNotation(pitchclass.TE))
	fmt.Println(
		s.MustNormalFormNotation(pitchclass.TE, pcset.Rahn),
		s.MustNormalFormCompactNotation(pitchclass.TE, pcset.Rahn),
	)

	p1 := pcset.MustParse("{2, 3, 7, 11} mod 12")
	p2 := pcset.MustParseNotation("[E237]", pitchclass.TE, 12)
	fmt.Println(p1 == s, p2 == s)
	// Output:
	// {2, 3, 7, 11} {2, 3, 7, E}
	// {237E}
	// [E, 2, 3, 7] [E237]
	// true true
}
