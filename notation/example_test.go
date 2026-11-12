package notation_test

import (
	"errors"
	"fmt"

	"github.com/go-muse/musex/notation"
	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/octave/number"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

func ExampleBuiltIn() {
	n := notation.MustBuiltIn(notation.EnglishUnicodeID)
	fmt.Println(n.MustFormat(spelling.CSharp()))
	// Output: C♯
}

func ExampleNew() {
	english := notation.MustBuiltIn(notation.EnglishUnicodeID)
	n := notation.MustNew(notation.Definition{
		Name:            "english-with-russian-octaves",
		Rules:           english.NameRules(),
		OctaveStyle:     notation.OctaveStyleName,
		OctaveSeparator: " ",
	})
	p := pitch.MustFromASPN(spelling.C(), 4)
	fmt.Println(n.MustFormatPitch(p, system.Traditional(), notation.OctaveContext{Catalog: naming.Russian()}))
	// Output: C первая
}

func ExampleNotation_Format() {
	n := notation.MustBuiltIn(notation.EnglishUnicodeID)
	text, err := n.Format(spelling.CSharp())
	fmt.Println(text, err)
	// Output: C♯ <nil>
}

func ExampleNotation_Parse() {
	n := notation.MustBuiltIn(notation.EnglishUnicodeID)
	name, err := n.Parse("C♯")
	fmt.Println(name.Equal(spelling.CSharp()), err)
	// Output: true <nil>
}

func ExampleNotation_FormatPitch_number() {
	n := notation.MustBuiltIn(notation.EnglishUnicodeID)
	p := pitch.MustFromASPN(spelling.BSharp(), 4)
	text, err := n.FormatPitch(p, system.Traditional(), notation.OctaveContext{Numbering: number.ASPN()})
	fmt.Println(text, err)
	// Output: B♯3 <nil>
}

func ExampleNotation_FormatPitch_name() {
	english := notation.MustBuiltIn(notation.EnglishUnicodeID)
	n := notation.MustNew(notation.Definition{
		Rules:           english.NameRules(),
		OctaveStyle:     notation.OctaveStyleName,
		OctaveSeparator: " ",
	})
	p := pitch.MustFromASPN(spelling.C(), 4)
	text, err := n.FormatPitch(p, system.Traditional(), notation.OctaveContext{Catalog: naming.Russian()})
	fmt.Println(text, err)
	// Output: C первая <nil>
}

func ExampleConvert() {
	german := notation.MustBuiltIn(notation.GermanAMID)
	english := notation.MustBuiltIn(notation.EnglishUnicodeID)
	text, err := notation.Convert("cis", german, english)
	fmt.Println(text, err)
	// Output: C♯ <nil>
}

func ExampleConvertPitch() {
	german := notation.MustBuiltIn(notation.GermanAMID)
	english := notation.MustBuiltIn(notation.EnglishUnicodeID)
	abc := notation.MustBuiltIn(notation.ABCID)
	aspn := notation.OctaveContext{Numbering: number.ASPN()}
	englishText, err := notation.ConvertPitch("cis¹", german, english, system.Traditional(), notation.OctaveContext{}, aspn)
	fmt.Println(englishText, err)
	abcText, err := notation.ConvertPitch(englishText, english, abc, system.Traditional(), aspn, notation.OctaveContext{})
	fmt.Println(abcText, err)
	// Output:
	// C♯4 <nil>
	// ^C <nil>
}

func ExampleBuiltIn_errors() {
	_, unknownErr := notation.BuiltIn("unknown")
	_, missingUnitErr := notation.BuiltIn(notation.ASCIIUpDownID)
	fmt.Println(errors.Is(unknownErr, notation.ErrUnknownID))
	fmt.Println(errors.Is(missingUnitErr, notation.ErrMissingChromaticUnit))
	// Output:
	// true
	// true
}
