package instrument_test

import (
	"errors"
	"fmt"
	"slices"

	"github.com/go-muse/musex/instrument"
	"github.com/go-muse/musex/notation"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"
	"github.com/go-muse/musex/tuning/temperament"
)

// A range is checked by position in a system, and its bounds are included:
// B♯3 stands at the position of C4, C♭4 at the position of B3.
func ExampleRange_Contains() {
	std := settings.Standard()
	p := std.MustParsePitch
	sys := std.System()

	soprano := instrument.MustNewRange(p("C4"), p("G5"))
	fmt.Println(soprano.MustContains(sys, p("B♯3")), soprano.MustContains(sys, p("C♭4")))
	fmt.Println(soprano.MustWidth(sys), soprano.MustCount(sys))
	fmt.Println(soprano.MustLocate(sys, p("B3")), soprano.MustLocate(sys, p("A5")))
	// Output:
	// true false
	// 19 20
	// below above
}

// A pitch is folded into a range by whole octaves; its name stays.
func ExampleRange_Folds() {
	std := settings.Standard()
	p := std.MustParsePitch
	sys := std.System()

	wide := instrument.MustNewRange(p("C3"), p("G5"))
	fmt.Println(slices.Collect(wide.MustFolds(sys, p("E1"))))
	e, ok := wide.MustFold(sys, p("E1"))
	fmt.Println(e, ok)

	narrow := instrument.MustNewRange(p("C4"), p("E4"))
	_, ok = narrow.MustFold(sys, p("G4"))
	fmt.Println(ok)
	// Output:
	// [E ASPN 3 E ASPN 4 E ASPN 5]
	// E ASPN 3 true
	// false
}

// A preset of your own: a name, a set of settings and a reference range.
func ExampleNewPreset() {
	std := settings.Standard()
	p := std.MustParsePitch

	lead, err := instrument.NewPreset("Lead voice", std, instrument.MustNewRange(p("C4"), p("A5")))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(lead.Name(), lead.HasRange(), lead.Range().Lower(), lead.Range().Upper())

	_, err = instrument.NewPreset("", std, instrument.Range{})
	fmt.Println(errors.Is(err, instrument.ErrEmptyName))
	// Output:
	// Lead voice true C ASPN 4 A ASPN 5
	// true
}

// An instrument is a source of settings: a note that reads it sees the new
// set after a retuning, and the instrument stays the same source.
func ExampleInstrument_Retune() {
	std := settings.Standard()
	p := std.MustParsePitch

	g := instrument.MustFromParts("Lead voice", std, instrument.MustNewRange(p("C4"), p("A5")))
	var src settings.Source = g
	a4 := p("A4")
	fmt.Println(src.Settings().Tuning().MustFrequency(a4))

	err := g.Retune(std.MustWithTuning(tuning.Standard().MustWithReferenceFrequency(415)))
	fmt.Println(err)
	fmt.Println(src.Settings().Tuning().MustFrequency(a4))
	fmt.Println(settings.SameSource(src, g))
	// Output:
	// 440
	// <nil>
	// 415
	// true
}

// A retuning that the rules of settings refuse keeps the state and gives the
// report with the violated rules.
func ExampleInstrument_Retune_incompatible() {
	std := settings.Standard()
	g := instrument.MustFromParts("Lead voice", std, instrument.Range{})

	german := std.MustWithNotation(notation.MustBuiltIn(notation.GermanAMID))
	err := g.Retune(german)
	ie, ok := errors.AsType[*instrument.IncompatibleError](err)
	fmt.Println(errors.Is(err, instrument.ErrIncompatible), errors.Is(err, settings.ErrIncompatible), ok)
	_, designations := ie.Report.Violation(settings.RuleDesignations)
	fmt.Println(ie.Report.Compatible(), designations)
	fmt.Println(g.Settings().Equal(std))
	// Output:
	// true true true
	// false true
	// true
}

// A retuning that would invert the range is refused with its own cause; the
// range and the set can be replaced together instead.
func ExampleInstrument_RetuneWithRange() {
	std := settings.Standard()
	p := std.MustParsePitch

	// B♯3 and C♯4 are one step apart in traditional, but in 19-position B♯
	// stands above C♯.
	g := instrument.MustFromParts("Lead voice", std, instrument.MustNewRange(p("B♯3"), p("C♯4")))
	t19 := tuning.MustNew(system.Positions19(), temperament.EDO19(), tuning.RootFromName(spelling.C()), tuning.A4At440())
	n19 := std.MustWithTuning(t19)

	err := g.Retune(n19)
	fmt.Println(errors.Is(err, instrument.ErrInverted), errors.Is(err, instrument.ErrIncompatible))

	err = g.RetuneWithRange(n19, instrument.MustNewRange(p("C4"), p("G5")))
	fmt.Println(err)
	fmt.Println(g.Settings().System().Name(), g.Range().Lower(), g.Range().Upper())
	// Output:
	// true false
	// <nil>
	// 19-position C ASPN 4 G ASPN 5
}

// A change that depends on the current set: the transformation gets the
// current set and gives the new one; the instrument checks and keeps it.
func ExampleInstrument_RetuneFunc() {
	std := settings.Standard()
	g := instrument.MustFromParts("Lead voice", std, instrument.Range{})

	raise := func(s *settings.Settings) (*settings.Settings, error) {
		t, err := s.Tuning().WithReferenceFrequency(s.Tuning().Reference().Frequency() + 2)
		if err != nil {
			return nil, err
		}
		return s.WithTuning(t)
	}
	g.MustRetuneFunc(raise).MustRetuneFunc(raise)
	fmt.Println(g.Settings().Tuning().Reference().Frequency())
	// Output:
	// 444
}

// The range of an instrument is checked in the system of its current set.
func ExampleInstrument_Contains() {
	std := settings.Standard()
	p := std.MustParsePitch

	g := instrument.MustFromParts("Soprano part", std, instrument.MustNewRange(p("C4"), p("G5")))
	fmt.Println(g.MustContains(p("B♯3")), g.MustContains(p("C♭4")), g.MustLocate(p("A5")))
	b, ok := g.MustFold(p("B5"))
	fmt.Println(b, ok)

	_, err := instrument.MustFromParts("No range", std, instrument.Range{}).Contains(p("C4"))
	fmt.Println(errors.Is(err, instrument.ErrNoRange))
	// Output:
	// true false above
	// B ASPN 4 true
	// true
}

// A catalog of your own: the order is kept and the search is exact.
func ExampleCatalog_Lookup() {
	std := settings.Standard()
	p := std.MustParsePitch

	lead := instrument.MustNewPreset("Lead voice", std, instrument.MustNewRange(p("C4"), p("A5")))
	alto := instrument.MustNewPreset("Alto part", std, instrument.MustNewRange(p("G3"), p("D5")))
	c := instrument.MustNewCatalog(lead, alto)
	fmt.Println(slices.Collect(c.Names()))

	found, ok := c.Lookup("Alto part")
	fmt.Println(found.Name(), ok)
	_, ok = c.Lookup("alto part")
	fmt.Println(ok)

	_, err := c.With(alto)
	fmt.Println(errors.Is(err, instrument.ErrDuplicateName))
	// Output:
	// [Lead voice Alto part]
	// Alto part true
	// false
	// true
}
