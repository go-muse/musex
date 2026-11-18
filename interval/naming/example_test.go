package naming_test

import (
	"fmt"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/interval/naming"
)

func ExampleCatalog() {
	down := interval.MinorThird().MustDown()
	short, ok := naming.Russian().Specific(down).MustName()
	full, _ := naming.English().Specific(down).MustNameFull()
	fmt.Println(short, ok)
	fmt.Println(full)
	fmt.Println(down)
	// Output:
	// м.3 true
	// minor third
	// -m3
}
