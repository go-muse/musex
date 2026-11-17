package tempomap_test

import (
	"fmt"

	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tempomap"
)

func ExampleMap() {
	m := tempomap.MustNew(quarter(60)).
		MustWithChange(tempomap.MustNewJump(at(2, 1), quarter(120)))
	start := m.MustStart(at(3, 1), 0, rounding.HalfUp)
	position := m.MustPosition(sec(10, 1), 0, rounding.HalfUp)
	fmt.Println(start)
	fmt.Println(position)
	// Output:
	// s:10/1
	// p:3/1
}

func ExampleMap_gradual() {
	gradual := tempomap.MustNewGradual(tempomap.MustAt(duration.Origin()), tempomap.LinearSpeed).
		MustWithSlot(tempomap.SlotSecond)
	m := tempomap.MustNew(quarter(60)).
		MustWithChange(gradual).
		MustWithChange(tempomap.MustNewMain(at(1, 1), quarter(120)))
	seconds := m.MustStart(at(1, 1), 3, rounding.HalfUp)
	fmt.Println(m.MustSpeed(at(1, 2)))
	fmt.Println(m.MustTempoAt(at(1, 2), duration.Quarter.MustWritten()))
	fmt.Println(seconds)
	// Output:
	// 45/2
	// 4=90
	// s~2773/1000@3:1/2000
}

func ExampleMap_fermata() {
	m := tempomap.MustNew(quarter(60)).MustWithExtension(
		tempomap.MustNewExtension(at(1, 2), sec(1, 1)),
	)
	fmt.Println(m.MustStart(at(1, 2), 3, rounding.HalfUp))
	fmt.Println(m.MustEnd(at(1, 2), 3, rounding.HalfUp))
	fmt.Println(m.MustBetween(at(1, 2), at(3, 4), 3, rounding.HalfUp))
	// Output:
	// s:2/1
	// s:3/1
	// s:2/1
}

func ExampleMap_measures() {
	grid := grid34Pickup()
	m := tempomap.MustNewMeasured(grid, quarter(60))
	address, position := m.MustAddressAt(grid, sec(9, 2), 3, rounding.HalfUp)
	fmt.Println(address)
	fmt.Println(position)
	// Output:
	// 2:1/8
	// p:9/8
}
