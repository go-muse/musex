package duration_test

import (
	"fmt"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tuplet"
)

// SubdivideTuplet divides a written duration into the parts of a tuplet.
func ExampleWritten_SubdivideTuplet() {
	quarter := duration.MustNewWritten(duration.Quarter, 0)
	count, part := quarter.MustSubdivideTuplet(tuplet.Triplet())
	fmt.Println(count, part.MustDuration().MustFraction())
	// Output:
	// 3 1/12
}

// NewGroup builds a written tuplet group with contents and exposes its spans.
func ExampleNewGroup() {
	triplet := tuplet.Triplet()
	eighth := duration.MustNewWritten(duration.Eighth, 0)
	g := duration.MustNewGroup(triplet, eighth,
		duration.WrittenContent(duration.MustNewWritten(duration.Quarter, 0, triplet)),
		duration.WrittenContent(duration.MustNewWritten(duration.Eighth, 0, triplet)),
	)
	fmt.Println(g.MustSpan(), g.MustShortfall())
	for _, span := range g.MustSpans() {
		fmt.Println(span[0], span[1])
	}
	// Output:
	// 1/4 0/1
	// 0/1 1/6
	// 1/6 1/4
}

// Readings finds the exact rhythmic readings of a sequence of sounding
// durations.
func ExampleReadings() {
	ds := make([]duration.Duration, 6)
	for i := range ds {
		ds[i] = duration.MustNew(1, 12)
	}
	readings := duration.MustReadings(ds, duration.ReadingsOptions{
		Allowed: []tuplet.Tuplet{tuplet.Triplet(), tuplet.Sextuplet()},
	})
	fmt.Println(len(readings))
	for _, item := range readings[0].MustContents() {
		fmt.Println(item.MustGroup().MustTuplet())
	}
	// Output:
	// 2
	// 6:4
}

// With FreeContents a group may hold any number of values that fill it
// exactly: a quarter and an eighth sounding 1/6 and 1/12 have no reading with
// fixed contents, and one 3:2 group in eighths with free contents.
func ExampleReadings_freeContents() {
	ds := []duration.Duration{duration.MustNew(1, 6), duration.MustNew(1, 12)}
	allowed := []tuplet.Tuplet{tuplet.Triplet()}

	fixed := duration.MustReadings(ds, duration.ReadingsOptions{Allowed: allowed})
	fmt.Println(len(fixed))

	free := duration.MustReadings(ds, duration.ReadingsOptions{Allowed: allowed, FreeContents: true})
	fmt.Println(len(free))

	g := free[0].MustContents()[0].MustGroup()
	fmt.Println(g.MustTuplet(), g.MustUnit())
	for _, c := range g.MustContents() {
		fmt.Println(c.MustWritten())
	}
	// Output:
	// 0
	// 1
	// 3:2 8
	// 4[3:2]
	// 8[3:2]
}
