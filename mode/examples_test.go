package mode_test

import (
	"fmt"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

func ExampleMode_NotesFrom() {
	m := mode.MustFromName("Dorian")
	first, err := note.FromSpelling(spelling.MustParse("D"), cents.Cents{}, nil)
	if err != nil {
		panic(err)
	}
	for _, n := range m.MustNotesFrom(first) {
		fmt.Println(n.Spelling())
	}
	// Output:
	// D
	// E
	// F
	// G
	// A
	// B
	// C
}

func ExampleMode_DegreeInterval() {
	m := mode.MustFromCode("ionian")
	fmt.Println(m.Len(), m.MustDegree(3).Number())
	fmt.Println(m.MustDegreeInterval(9), m.MustBetweenDegrees(4, 7))
	next, wrapped := m.MustNext(7)
	fmt.Println(next.Number(), wrapped)
	// Output:
	// 7 3
	// M9 A4
	// 1 true
}

func ExampleMode_MismatchingDegrees() {
	a, b := mode.MustFromCode("dorian"), mode.MustFromCode("aeolian")
	fmt.Println(a.Equal(b))
	fmt.Println(a.MustMismatchingDegrees(b, mode.FromFirst).Degrees)
	fmt.Println(a.MustMismatchingDegrees(b, mode.Adjacent).Degrees)
	// Output:
	// false
	// [6]
	// [5 6]
}

func ExampleRepository_Identify() {
	m := mode.MustParse("M2 m2 M2 M2 M2 m2 M2")
	repo := mode.BuiltInRepository()
	for _, e := range repo.MustIdentify(m) {
		fmt.Println(e.Code())
	}
	for _, r := range repo.MustIdentifyRotations(m) {
		if r.Entry.Code() == "ionian" {
			fmt.Println(r.Entry.Code(), r.Degree)
		}
	}
	// Output:
	// dorian
	// ionian 2
}

func ExampleMode_SetDegreeInterval() {
	m := mode.MustFromName("Dorian")
	original := m
	m.MustSetDegreeInterval(6, interval.MinorSixth())
	fmt.Println(m)
	fmt.Println(original.Changed())
	m.MustSetDegreeInterval(6, interval.MajorSixth())
	fmt.Println(m.Changed())
	// Output:
	// Dorian (edited): M2 m2 M2 M2 m2 M2 M2
	// false
	// false
}

func ExampleRotations() {
	m := mode.MustParse("M2 M2 m2")
	for candidate := range mode.MustRotations(m) {
		fmt.Println(candidate)
	}
	// Output:
	// M2 M2 m2
	// M2 m2 M2
	// m2 M2 M2
}

func ExampleMode_Brightness() {
	m := mode.MustFromCode("lydian")
	rank, known := m.MustBrightness()
	fmt.Println(m.IsSecondwise(), m.MustIsDiatonic(), rank, known)
	fmt.Println(m.MustTetrachords().SeparatedByTone())
	// Output:
	// true true 1 true
	// false
}

func ExampleMode_Numbers() {
	sys := system.Traditional()
	m := mode.MustParse("M2 M2 m2")
	p := m.MustNumbers(sys)
	cycle := m.MustCycleIn(sys)
	fmt.Println(p.Adjacent, p.Cycle)
	fmt.Println(cycle.Cycles, cycle.Octaves)
	// Output:
	// [2 2 1] 5
	// 12 5
}

func ExampleMode_KeySignatureFrom() {
	m := mode.MustFromCode("aeolian")
	signature, ok := m.MustKeySignatureFrom(spelling.MustParse("C"))
	fmt.Println(signature.Index, ok)
	fmt.Println(signature.Signs)
	fmt.Println(m.MustFirstSpellingForSignature(signature.Index))
	// Output:
	// -3 true
	// [B-1 E-1 A-1]
	// C
}
