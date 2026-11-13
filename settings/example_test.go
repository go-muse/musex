package settings_test

import (
	"errors"
	"fmt"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/notation"
	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/octave/number"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"
	"github.com/go-muse/musex/tuning/temperament"
)

// The standard settings: the traditional system, 12-EDO with A4 = 440 Hz,
// English names with ASPN octave numbers and English octave names.
func ExampleStandard() {
	s := settings.Standard()
	fmt.Println(s.IsValid(), s.System().Name())
	fmt.Println(s.Notation().Name(), s.Numbering().Label(), s.Catalog().Name(4))
	fmt.Println(s.Tuning().Equal(tuning.Standard()))
	fmt.Println(s.Equal(settings.Standard()))
	// Output:
	// true traditional
	// english-unicode ASPN one-line
	// true
	// true
}

// Creating a set from a tuning and designations, and a refused creation.
func ExampleNew() {
	s, err := settings.New(tuning.Standard(), notation.MustBuiltIn(notation.GermanAMID), number.FrancoBelgian(), naming.Russian())
	fmt.Println(err)
	fmt.Println(s.MustFormatPitch(pitch.MustFromASPN(spelling.CSharp(), 4)))

	// The ups-and-downs chromatic unit 3 does not match c = 1 of the
	// traditional system.
	_, err = settings.New(tuning.Standard(), notation.MustBuiltIn(notation.ASCIIUpDownID, 3), number.ASPN(), nil)
	fmt.Println(errors.Is(err, settings.ErrInconsistent), errors.Is(err, notation.ErrIncompatibleSystem))
	// Output:
	// <nil>
	// cis¹
	// true true
}

// Replacing one part gives a new set; the receiver does not change.
func ExampleSettings_WithNotation() {
	s := settings.Standard()
	german := s.MustWithNotation(notation.MustBuiltIn(notation.GermanAMID))
	c4s := pitch.MustFromASPN(spelling.CSharp(), 4)
	fmt.Println(german.MustFormatPitch(c4s), s.MustFormatPitch(c4s))
	fmt.Println(s.Equal(settings.Standard()), german.Equal(s))
	// Output:
	// cis¹ C♯4
	// true false
}

// Changing the system keeps the temperament, the root number, the reference
// and the designations; a system with another number of classes is refused.
func ExampleSettings_WithSystem() {
	s := settings.Standard()
	_, err := s.WithSystem(system.QuarterTones24())
	fmt.Println(errors.Is(err, tuning.ErrClassCount))
	// Output:
	// true
}

// Writing and reading pitches and names with the parts of a set.
func ExampleSettings_FormatPitch() {
	s := settings.Standard()
	fmt.Println(s.MustFormatPitch(pitch.MustFromASPN(spelling.CSharp(), 4)))

	// B♯3 is the pitch (B♯, Octave 4) in the traditional system.
	p := s.MustParsePitch("B♯3")
	fmt.Println(p.Spelling() == spelling.BSharp(), p.MustASPN())

	// A notation without an octave writes names but no pitches.
	italian := s.MustWithNotation(notation.MustBuiltIn(notation.RomanceItalianID))
	fmt.Println(italian.MustFormat(spelling.CSharp()), italian.MustExpressesPitch(p))
	_, err := italian.FormatPitch(p)
	fmt.Println(errors.Is(err, notation.ErrNoOctave))
	// Output:
	// C♯4
	// true 4
	// do diesis false
	// true
}

// A quarter-tone name is not admissible in the traditional system.
func ExampleSettings_Format() {
	s := settings.Standard()
	cq := spelling.C().WithAlteration(alteration.MustFromFraction(1, 2))
	fmt.Println(s.MustExpresses(cq))
	_, err := s.Format(cq)
	fmt.Println(errors.Is(err, settings.ErrInadmissible), errors.Is(err, system.ErrInadmissible))
	// Output:
	// false
	// true true
}

// A set is its own source; equal sets are one source.
func ExampleSameSource() {
	s := settings.Standard()
	german := s.MustWithNotation(notation.MustBuiltIn(notation.GermanAMID))
	var src settings.Source = s
	fmt.Println(src.Settings() == s)
	fmt.Println(settings.SameSource(settings.Standard(), settings.Standard()))
	fmt.Println(settings.SameSource(s, german), settings.SameSource(s, nil))
	// Output:
	// true
	// true
	// false false
}

// Checking whether a set can replace another, with reasons and kinds of change.
func ExampleSettings_CheckReplacement() {
	s := settings.Standard()
	t24 := tuning.MustNew(system.QuarterTones24(), temperament.Equal24(),
		tuning.RootFromName(spelling.C()), tuning.A4At440())
	wide := s.MustWithTuning(t24)
	fmt.Println(s.MustCheckReplacement(wide).Compatible())

	back := wide.MustCheckReplacement(s)
	v, ok := back.Violation(settings.RuleStructure)
	fmt.Println(back.Compatible(), ok, wide.System().MustAdmits(v.Name), s.System().MustAdmits(v.Name))
	_, ok = back.Violation(settings.RuleDesignations)
	fmt.Println(ok, errors.Is(back.Err(), settings.ErrIncompatible))
	// Output:
	// true
	// false true true false
	// false true
}

// The kinds of a change from one set to another.
func ExampleSettings_Changes() {
	s := settings.Standard()
	t24 := tuning.MustNew(system.QuarterTones24(), temperament.Equal24(),
		tuning.RootFromName(spelling.C()), tuning.A4At440())
	ch := s.MustChanges(s.MustWithTuning(t24))
	fmt.Println(ch.Structure == settings.StructureExtended, ch.Sound, ch.Designations)
	fmt.Println(s.MustChanges(settings.Standard()).None())
	// Output:
	// true true false
	// true
}
