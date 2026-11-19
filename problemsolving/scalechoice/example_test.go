package scalechoice_test

import (
	"fmt"

	"github.com/go-muse/musex/problemsolving/scalechoice"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

func ExampleAnalyzeSet() {
	a := scalechoice.MustAnalyzeSet(notes("C D E G A"), scalechoice.Classical(), opts())
	fmt.Println(len(a.Groups()))
	fmt.Println(a.MustAfterAdding(spelling.F()).Kept)
	fmt.Println(a.MustAfterAdding(spelling.F(), spelling.B()).Kept)
	// Output:
	// 3
	// [0 1]
	// [0]
}

func ExampleReading_DegreesOf() {
	o := full()
	o.Center = spelling.D()
	a := scalechoice.MustAnalyzeSet(notes("C D E F G A B"), scalechoice.Classical(), o)
	r := a.Groups()[0].Readings[0]
	fmt.Println(r.Entry().Code(), r.Tonic())
	fmt.Println(r.MustDegreesOf(spelling.D()), r.MustDegreesOf(spelling.F()))
	// Output:
	// dorian D
	// [1] [3]
}

func ExampleAnalysis_ReplaceAll() {
	o := full()
	o.Center = spelling.C()
	o.System = system.Traditional()
	line := scalechoice.MustAnalyzePhrase(pitched("C D A B E F G", 4, nil), scalechoice.Classical(), o)
	trial := line.MustReplaceAll(spelling.F(), spelling.FSharp())
	fmt.Println(trial.Readings.Removed[0].Entry().Code(), trial.Readings.Added[0].Entry().Code())
	for _, edge := range trial.Neighbors {
		fmt.Println(edge.Before.Chromatic.Steps(), edge.After.Chromatic.Steps())
	}
	// Output:
	// ionian lydian
	// 1 2
	// 2 1
}
