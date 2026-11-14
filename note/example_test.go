package note_test

import (
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/frequency"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"
)

// Examples of architecture 6.14 and 6.15 (NF-6). They use only the public API.

// A note bound to the standard settings computes its frequency through the
// tuning of the set it reads at the call.
func Example() {
	std := settings.Standard()
	c4 := note.MustFromPitch(pitch.MustFromASPN(spelling.C(), 4), cents.Zero(), std)
	fmt.Printf("%.4f\n", c4.MustFrequency())
	fmt.Println(c4.MustFormat())
	// Output:
	// 261.6256
	// C4
}

// A changing method changes the note itself; its Must variant returns the same
// note, so Must variants chain (K-42).
func ExampleNote_MustRaise() {
	std := settings.Standard()
	n := note.MustFromPitch(pitch.MustFromASPN(spelling.C(), 4), cents.MustFromInt(30), std)
	n.MustRaise().MustUp().MustSetCents(cents.MustFromInt(-10)) // n is now C♯5 −10 cents
	fmt.Printf("%.4f\n", n.MustFrequency())
	fmt.Println(n)
	// Output:
	// 551.1724
	// C+1 ASPN 5, -10 cents, bound
}

// A copy by assignment is independent; only the source is shared (K-43).
func ExampleNote_MustMove() {
	std := settings.Standard()
	root := note.MustFromPitch(pitch.MustFromASPN(spelling.C(), 4), cents.Zero(), std)
	third := root                      // an independent copy; the source is shared
	third.MustMove(4, system.Sharps()) // third is E4, root is still C4
	fmt.Println(root.MustFormat(), third.MustFormat(), root.SameSource(third))
	// Output:
	// C4 E4 true
}

// A failed change leaves the note as it was (K-44).
func ExampleNote_Raise() {
	narrow := settings.Standard().MustWithSystem(
		system.MustNewRegular(2, 1, letter.C, system.MustIntegerRange(-1, 1), "narrow"))
	n := note.MustFromPitch(pitch.MustFromASPN(spelling.CSharp(), 4), cents.MustFromInt(30), narrow)
	if err := n.Raise(); err != nil { // C𝄪 is not admissible: n is unchanged
		fmt.Println(errors.Is(err, note.ErrInadmissible))
	}
	fmt.Println(n.MustFormat(), n.Cents())
	// Output:
	// true
	// C♯4 30
}

// A chain from creation needs an addressable note: a variable or new(…).
// Elements of a slice change in place; a map element goes through a variable.
func ExampleNote_MustLower() {
	std := settings.Standard()
	c4 := pitch.MustFromASPN(spelling.C(), 4)

	p := new(note.MustFromPitch(c4, cents.Zero(), std)).MustRaise() // p is *note.Note
	fmt.Println(p.MustFormat())

	notes := []note.Note{note.MustFromPitch(c4, cents.Zero(), std)}
	for _, x := range notes {
		x.MustLower() // changes a copy only
	}
	fmt.Println(notes[0].MustFormat())
	notes[0].MustLower() // changes the element
	fmt.Println(notes[0].MustFormat())

	m := map[string]note.Note{"k": note.MustFromPitch(c4, cents.Zero(), std)}
	v := m["k"]
	v.MustLower()
	m["k"] = v // a map element is not addressable
	fmt.Println(m["k"].MustFormat())
	// Output:
	// C♯4
	// C4
	// C♭5
	// C♭5
}

// An unbound note has no set: an explicit tuning or source computes for one
// call without binding the note (K-19).
func ExampleNote_FrequencyInTuning() {
	free := note.MustFromPitch(pitch.MustFromASPN(spelling.C(), 4), cents.Zero(), nil)
	fmt.Printf("%.4f\n", free.MustFrequencyInTuning(tuning.Standard()))
	fmt.Printf("%.4f\n", free.MustFrequencyIn(settings.Standard()))
	_, err := free.Frequency()
	fmt.Println(errors.Is(err, note.ErrNoSource), free.IsBound())
	// Output:
	// 261.6256
	// 261.6256
	// true false
}

// Two ways to raise: by sound, keeping the Octave, as package pitch does, and
// by record, keeping the written octave (K-11).
func ExampleNote_RaiseWritten() {
	std := settings.Standard()
	b3 := note.MustFromOctave(spelling.B(), octave.FromASPN(3), cents.Zero(), std)
	x, y := b3, b3
	x.MustRaise()        // B♯ in Octave 3, written "B♯2"
	y.MustRaiseWritten() // B♯ in Octave 4, written "B♯3"
	fmt.Println(x.MustFormat(), y.MustFormat())
	// Output:
	// B♯2 B♯3
}

// tunable is a source whose set its owner replaces whole, as an instrument
// does.
type tunable struct {
	set atomic.Pointer[settings.Settings]
}

func (s *tunable) Settings() *settings.Settings { return s.set.Load() }

// The note reads its source at each call: retuning the source changes the
// frequency of the same note (K-24).
func ExampleNote_Frequency() {
	src := &tunable{}
	src.set.Store(settings.Standard())
	a4 := note.MustFromPitch(pitch.MustFromASPN(spelling.A(), 4), cents.Zero(), src)
	fmt.Printf("%.1f\n", a4.MustFrequency())
	src.set.Store(settings.Standard().MustWithTuning(tuning.Standard().MustWithReferenceFrequency(415)))
	fmt.Printf("%.1f\n", a4.MustFrequency())
	// Output:
	// 440.0
	// 415.0
}

// SetSource moves a note to another source, keeping its name and Octave (K-49).
func ExampleNote_SetSource() {
	n := note.MustFromPitch(pitch.MustFromASPN(spelling.C(), 4), cents.MustFromInt(5), settings.Standard())
	before := n
	other := settings.Standard().MustWithTuning(tuning.Standard().MustWithReferenceFrequency(415))
	if err := n.SetSource(other); err != nil {
		fmt.Println(err)
	}
	fmt.Println(n.Equal(before), n.SameSource(before))
	// Output:
	// true false
}

// Equal compares the record and the cents; == also compares the source.
func ExampleNote_Equal() {
	a := note.MustFromPitch(pitch.MustFromASPN(spelling.C(), 4), cents.Zero(), settings.Standard())
	b := note.MustFromPitch(pitch.MustFromASPN(spelling.C(), 4), cents.Zero(), settings.Standard())
	fmt.Println(a.Equal(b), a == b)
	// Output:
	// true false
}

// Comparison by position needs one system; comparison by sound works across
// sets by frequencies, with an optional criterion (K-4, K-39).
func ExampleNote_CompareSoundBy() {
	std := settings.Standard()
	c := note.MustFromPitch(pitch.MustFromASPN(spelling.C(), 4), cents.MustFromInt(150), std)
	cs := note.MustFromPitch(pitch.MustFromASPN(spelling.CSharp(), 4), cents.Zero(), std)
	fmt.Println(c.MustComparePosition(cs), c.MustCompareSound(cs))

	a440 := note.MustFromPitch(pitch.MustFromASPN(spelling.A(), 4), cents.Zero(), std)
	a441 := note.MustFromPitch(pitch.MustFromASPN(spelling.A(), 4), cents.Zero(),
		std.MustWithTuning(tuning.Standard().MustWithReferenceFrequency(441)))
	fmt.Println(a440.MustSoundsSame(a441), a440.MustCompareSoundBy(a441, frequency.MustWithinCents(cents.MustFromInt(5))))
	// Output:
	// -1 1
	// false 0
}

// Offset tells where a measured frequency is relative to the sound of a note
// and by how many cents (K-40).
func ExampleNote_Offset() {
	a4 := note.MustFromPitch(pitch.MustFromASPN(spelling.A(), 4), cents.Zero(), settings.Standard())
	where, dev := a4.MustOffset(441)
	fmt.Println(where, dev.Decimal(6, 0))
	// Output:
	// 1 3.930158
}
