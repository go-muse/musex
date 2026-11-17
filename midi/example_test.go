package midi_test

import (
	"errors"
	"fmt"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/midi"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

func ExampleNew() {
	c4 := pitch.MustFromASPN(spelling.C(), 4)
	a4 := pitch.MustFromASPN(spelling.A(), 4)
	fmt.Println(midi.MustFromPitch(c4, system.Traditional()).MustInt())
	fmt.Println(midi.MustFromPitch(a4, system.Traditional()).MustInt())
	fmt.Println(midi.Min().MustInt(), midi.Max().MustInt(), midi.Count())

	_, err := midi.New(128)
	fmt.Println(errors.Is(err, midi.ErrInvalidNumber))
	_, err = midi.MustNew(60).Shift(68)
	fmt.Println(errors.Is(err, midi.ErrOutOfRange))

	// Output:
	// 60
	// 69
	// 0 127 128
	// true
	// true
}

func ExampleNumber_Parts() {
	number := midi.MustNew(60)
	o, class, position := number.MustParts()
	fmt.Println(o.MustASPN(), class.Number(), class.Modulus(), position)
	fmt.Println(midi.MustFromParts(o, class).MustInt())

	// Output:
	// 4 0 12 48
	// 60
}

func ExampleFromPitch_enharmonic() {
	bSharp3 := pitch.MustFromASPN(spelling.BSharp(), 4)
	c4 := pitch.MustFromASPN(spelling.C(), 4)
	cFlat4 := pitch.MustFromASPN(spelling.CFlat(), 3)
	b3 := pitch.MustFromASPN(spelling.B(), 3)

	fmt.Println(midi.MustFromPitch(bSharp3, system.Traditional()).MustInt())
	fmt.Println(midi.MustFromPitch(c4, system.Traditional()).MustInt())
	fmt.Println(midi.MustFromPitch(cFlat4, system.Traditional()).MustInt())
	fmt.Println(midi.MustFromPitch(b3, system.Traditional()).MustInt())
	fmt.Println(bSharp3.Equal(c4), cFlat4.Equal(b3))

	// Output:
	// 60
	// 60
	// 59
	// 59
	// false false
}

func ExampleNumber_Pitch() {
	number := midi.MustNew(61)
	sharp := number.MustPitch(system.Traditional(), system.Sharps())
	flat := number.MustPitch(system.Traditional(), system.Flats())
	fmt.Println(sharp.Spelling(), sharp.MustASPN())
	fmt.Println(flat.Spelling(), flat.MustASPN())

	// Output:
	// C+1 4
	// D-1 4
}

func ExampleNumber_Note() {
	number := midi.MustNew(60)
	natural := number.MustNote(settings.Standard(), system.Sharps(), cents.Cents{})
	corrected := number.MustNote(settings.Standard(), system.Sharps(), cents.MustFromInt(30))
	fmt.Println(natural.Cents().Float64(), corrected.Cents().Float64())

	modal := system.MustChromaticScale(spelling.C(), system.Major, system.Ascending)
	_, err := number.Note(settings.Standard(), modal, cents.Cents{})
	fmt.Println(errors.Is(err, midi.ErrNotApplicable), errors.Is(err, system.ErrNotApplicable))

	_, err = number.Note(settings.Standard(), system.SpellingRule{}, cents.Cents{})
	fmt.Println(errors.Is(err, midi.ErrNoSpellingRule))

	// Output:
	// 0 30
	// true false
	// true
}

func ExampleFromNoteIn() {
	unbound := note.MustFromPitch(pitch.MustFromASPN(spelling.A(), 4), cents.Zero(), nil)
	number, err := midi.FromNoteIn(settings.Standard(), unbound)
	fmt.Println(number.MustInt(), err)
	fmt.Println(unbound.IsBound())

	// Output:
	// 69 <nil>
	// false
}

func ExampleNumber_Pitch_positionOctave() {
	bSharp3 := pitch.MustFromASPN(spelling.BSharp(), 4) // Written register 3, position octave 4.
	cFlat4 := pitch.MustFromASPN(spelling.CFlat(), 3)   // Written register 4, position octave 3.
	fmt.Println("B♯3:", bSharp3.MustASPN(), midi.MustFromPitch(bSharp3, system.Traditional()).MustInt())
	fmt.Println("C♭4:", cFlat4.MustASPN(), midi.MustFromPitch(cFlat4, system.Traditional()).MustInt())

	_, err := midi.FromPitch(pitch.MustFromASPN(spelling.C(), 4), system.Positions19())
	fmt.Println(errors.Is(err, midi.ErrNotTraditionalC))

	// Output:
	// B♯3: 4 60
	// C♭4: 3 59
	// true
}
