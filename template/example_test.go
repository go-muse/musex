package template_test

import (
	"encoding/json/v2"
	"errors"
	"fmt"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/template"
)

func Example_scenario1CreationAndReading() {
	tmpl := template.MustClosedFromIntervals([]interval.Interval{
		interval.MajorSecond(), interval.MajorSecond(), interval.MinorSecond(),
	})
	cycle, _ := tmpl.Cycle()
	places, _ := tmpl.PlaceIntervals()
	between, _ := tmpl.BetweenPlaces(3, 10)
	fmt.Println(tmpl.Len(), tmpl.String())
	fmt.Println(cycle, places, between)
	// Output:
	// 3 M2 M2 m2
	// P4 [P1 M2 M3] d8
}

func Example_scenario2CatalogEntryForConsumer() {
	entry, ok := template.BuiltIn().ByCode("dorian")
	matches := template.BuiltIn().ByName("Dorian")
	english, _ := entry.EnglishName()
	places, _ := entry.Template().PlaceIntervals()
	fmt.Println(ok, entry.Code(), entry.Template().Len(), english)
	fmt.Println(matches[0].Template().Equal(entry.Template()), places)
	// Output:
	// true dorian 7 Dorian
	// true [P1 M2 m3 P4 P5 M6 m7]
}

func Example_scenario3ArbitraryCycleAndOpenTemplate() {
	obikhod := template.MustClosedFromPlaceIntervals(
		[]interval.Interval{interval.PerfectUnison(), interval.MajorSecond(), interval.MajorThird()},
		interval.PerfectFourth(),
	)
	wide := template.MustClosedFromSpellings(
		[]spelling.Spelling{spelling.C(), spelling.E(), spelling.G(), spelling.C(), spelling.E()},
		interval.PerfectFifteenth(),
	)
	open := template.MustOpenFromSpellings([]spelling.Spelling{spelling.E(), spelling.G(), spelling.A()})
	obikhodCycle, _ := obikhod.Cycle()
	wideCycle, _ := wide.Cycle()
	_, openHasCycle := open.Cycle()
	fmt.Println(obikhodCycle, obikhod.IsOctave())
	fmt.Println(wideCycle, wide.String())
	fmt.Println(open.String(), openHasCycle, open.Len())
	// Output:
	// P4 false
	// P15 M3 m3 P4 M3 m6
	// m3 M2 | false 3
}

func Example_scenario4RotationAndEquality() {
	dorian := template.Ionian().MustRotate(2)
	place, ok := dorian.RotationOf(template.Ionian())
	back, okBack := template.Ionian().RotationOf(dorian)
	fmt.Println(dorian.Equal(template.Dorian()), place, ok)
	fmt.Println(back, okBack, template.Dorian().MustRotate(back).Equal(template.Ionian()))
	// Output:
	// true 2 true
	// 7 true true
}

func Example_scenario5Identification() {
	tmpl := template.MustParse("m2 M2 M2 M2 M2 m2 M2")
	matches := template.BuiltIn().MustIdentifyRotations(tmpl)
	for _, match := range matches {
		fmt.Println(match.Entry.Code(), match.Place)
	}
	// Output:
	// melodic-major 5
	// melodic-minor 2
	// acoustic 6
	// dorian-b2 1
	// lydian-augmented 7
	// locrian-sharp2 4
	// altered 3
}

func Example_scenario6CustomCatalog() {
	custom := template.MustNewEntry("my-a", []string{"Мой лад", "My mode"}, template.MustParse("M2 M2 m6"))
	catalog := template.MustNewCatalog(custom, template.IonianEntry())
	catalog = catalog.MustWith(template.MustNewEntry("my-tri", []string{"Мой трихорд"}, template.MustParse("m3 M2 |")))
	matches := catalog.ByName("my mode")
	identified := catalog.MustIdentify(template.MustParse("m3 M2 |"))
	fmt.Println(catalog.Codes())
	fmt.Println(matches[0].Code(), identified[0].Code())
	// Output:
	// [my-a ionian my-tri]
	// my-a my-tri
}

func Example_scenario7TextAndJSON() {
	tmpl := template.MustParse("M2 m2 M2 M2 m2 A2 m2")
	russian, ok, _ := tmpl.Russian()
	encoded, _ := json.Marshal(tmpl)
	var decoded template.Template
	_ = json.Unmarshal(encoded, &decoded)
	fmt.Println(tmpl.String())
	fmt.Println(russian, ok)
	fmt.Println(string(encoded), decoded.Equal(tmpl))
	// Output:
	// M2 m2 M2 M2 m2 A2 m2
	// б.2 м.2 б.2 б.2 м.2 ув.2 м.2 true
	// "M2 m2 M2 M2 m2 A2 m2" true
}

func Example_scenario8Related() {
	relations := template.BuiltIn().MustRelated(template.Dorian(), 1, template.FromFirst, false)
	for _, relation := range relations {
		fmt.Println(relation.Entry.Code(), relation.Count(), relation.Places)
	}
	// Output:
	// dorian 0 []
	// mixolydian 1 [3]
	// aeolian 1 [6]
	// melodic-minor 1 [7]
	// dorian-b2 1 [2]
	// ukrainian-dorian 1 [4]
	// dorian-b5 1 [5]
}

func Example_errors() {
	_, err := template.ClosedFromPlaceIntervals(
		[]interval.Interval{interval.PerfectUnison()}, interval.Interval{},
	)
	fmt.Println(errors.Is(err, template.ErrUnset))
	_, err = template.Dorian().MismatchingPlaces(template.Aeolian(), template.MeasureUnset)
	fmt.Println(errors.Is(err, template.ErrInvalidArgument))
	_, err = template.MustOpenFromIntervals([]interval.Interval{interval.MinorThird(), interval.MajorSecond()}).Rotate(2)
	fmt.Println(errors.Is(err, template.ErrOpenRotation))
	_, err = template.Parse("M2  M2")
	fmt.Println(errors.Is(err, template.ErrParse))
	// Output:
	// true
	// true
	// true
	// true
}
