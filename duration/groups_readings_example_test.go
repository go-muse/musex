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
