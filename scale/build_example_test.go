package scale_test

import (
	"fmt"

	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/scale"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

// FromTokens reads a scale from text tokens, preserving order and spellings.
func ExampleFromTokens() {
	s := scale.MustFromTokens(settings.Standard(), "C", "D♭", "C♯", "C")
	tokens := s.MustTokens()
	fmt.Println(s.Len(), tokens)
	// Output:
	// 4 [C D♭ C♯ C]
}

// FromSpellings builds a scale of name-only notes.
func ExampleFromSpellings() {
	s := scale.MustFromSpellings([]spelling.Spelling{spelling.C(), spelling.D()}, settings.Standard())
	fmt.Println(s.Len())
	// Output:
	// 2
}

// ChromaticRun builds a playable chromatic exercise with a directional rule.
func ExampleChromaticRun() {
	src := settings.Standard()
	from := pitch.MustNew(spelling.C(), octave.FromASPN(4))
	to := pitch.MustNew(spelling.C(), octave.FromASPN(5))
	rule := system.MustChromaticScale(spelling.C(), system.Major, system.Ascending)
	s := scale.MustChromaticRun(from, to, rule, src)
	tenth, err := s.At(10)
	if err != nil {
		panic(err)
	}
	fmt.Println(s.Len(), tenth.Spelling())
	// Output:
	// 13 B-1
}

// PitchClassSequence projects a scale to an ordered pitch-class sequence.
func ExampleScale_PitchClassSequence() {
	src := settings.Standard()
	s := scale.MustFromTokens(src, "C4", "D♭4", "C♯5", "D4")
	seq := s.MustPitchClassSequence(system.Traditional())
	for i := 0; i < seq.Len(); i++ {
		c, err := seq.At(i)
		if err != nil {
			panic(err)
		}
		fmt.Println(c.Number())
	}
	// Output:
	// 0
	// 1
	// 1
	// 2
}
