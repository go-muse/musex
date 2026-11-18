package mode_test

import (
	"fmt"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

func ExampleMode_ResolveRole() {
	major := mode.MustFromCode("ionian")
	tonic := pitch.MustFromASPN(spelling.C(), 4)
	shift := interval.MustNewChromatic(system.Traditional(), 3)
	result := major.MustResolveRole(tonic, mode.DegreePlace{Number: 1}, mode.DegreePlace{Number: 3}, shift, mode.AllowTargetAlteration)
	fmt.Println(result.Status == mode.RoleMismatch, result.HasCandidate)
	fmt.Println(result.Expected.Interval, result.Expected.Shift)
	fmt.Println(result.Candidate.Pitch.Spelling(), result.Candidate.Interval, result.Candidate.Shift)
	// Output:
	// true true
	// M3 +4
	// E-1 m3 +3
}

func ExampleMode_ResolveRole_needContext() {
	var unknown mode.Mode
	result := unknown.MustResolveRole(pitch.Pitch{}, mode.DegreePlace{Number: 1}, mode.DegreePlace{Number: 3}, interval.Chromatic{}, mode.StrictRole)
	fmt.Println(result.Status == mode.NeedContext, result.HasCandidate)
	// Output:
	// true false
}
