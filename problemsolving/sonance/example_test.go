package sonance_test

import (
	"fmt"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pcset"
	"github.com/go-muse/musex/problemsolving/sonance"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

// The C major triad scores 2.220 under Huron's published profile: one minor
// third, one major third and one perfect fourth/fifth class.
func ExampleScoreSet() {
	s := sonance.MustScoreSet(pcset.MustNew(12, 0, 4, 7), sonance.HuronProfile())
	fmt.Printf("%.3f\n", s.Value())
	for _, c := range s.Contributions() {
		if c.Count() > 0 {
			fmt.Printf("IC%d %d %.3f\n", c.Interval(), c.Count(), c.Value())
		}
	}
	// Output:
	// 2.220
	// IC3 1 0.594
	// IC4 1 0.386
	// IC5 1 1.240
}

// Doubling the bass adds a unison pair (weight zero in Huron's profile) and
// one more third and fifth pair each.
func ExampleScoreNotes() {
	ns := []note.Note{
		note.MustFromASPN(spelling.C(), 4, cents.Zero(), nil),
		note.MustFromASPN(spelling.C(), 4, cents.Zero(), nil),
		note.MustFromASPN(spelling.E(), 4, cents.Zero(), nil),
		note.MustFromASPN(spelling.G(), 4, cents.Zero(), nil),
	}
	s := sonance.MustScoreNotes(system.Traditional(), ns, sonance.HuronPairProfile())
	fmt.Printf("%.3f\n", s.Value())
	// Output:
	// 3.846
}

// A pair profile over the actual tempered interval distinguishes a fourth
// from a fifth: here the fourth weighs 1 and the fifth 2.
func ExampleNewPairProfile() {
	weights := make([]float64, 12)
	weights[5] = 1
	weights[7] = 2
	p := sonance.MustNewPairProfile(weights)
	sys := system.Traditional()
	fourth := sonance.MustScoreNotes(sys, []note.Note{
		note.MustFromASPN(spelling.C(), 4, cents.Zero(), nil),
		note.MustFromASPN(spelling.F(), 4, cents.Zero(), nil),
	}, p)
	fifth := sonance.MustScoreNotes(sys, []note.Note{
		note.MustFromASPN(spelling.C(), 4, cents.Zero(), nil),
		note.MustFromASPN(spelling.G(), 4, cents.Zero(), nil),
	}, p)
	fmt.Printf("%.3f %.3f\n", fourth.Value(), fifth.Value())
	// Output:
	// 1.000 2.000
}

// Two voicings of the C major triad, C4 E4 G4 and E4 G4 C5, contain the same
// pair classes once each, so they score alike under Huron's pair profile.
func ExampleScoreNotes_voicings() {
	sys := system.Traditional()
	rootPosition := sonance.MustScoreNotes(sys, []note.Note{
		note.MustFromASPN(spelling.C(), 4, cents.Zero(), nil),
		note.MustFromASPN(spelling.E(), 4, cents.Zero(), nil),
		note.MustFromASPN(spelling.G(), 4, cents.Zero(), nil),
	}, sonance.HuronPairProfile())
	firstInversion := sonance.MustScoreNotes(sys, []note.Note{
		note.MustFromASPN(spelling.E(), 4, cents.Zero(), nil),
		note.MustFromASPN(spelling.G(), 4, cents.Zero(), nil),
		note.MustFromASPN(spelling.C(), 5, cents.Zero(), nil),
	}, sonance.HuronPairProfile())
	fmt.Printf("%.3f %.3f\n", rootPosition.Value(), firstInversion.Value())
	// Output:
	// 2.220 2.220
}
