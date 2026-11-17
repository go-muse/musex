package degree_test

import (
	"errors"
	"fmt"
	"slices"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/degree"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
)

func ExampleDegree_creationAndText() {
	d := degree.Fourth()
	fmt.Println(d.MustDecimal(), d.MustRoman())
	parsed := degree.MustParse("LIII")
	fmt.Println(parsed.Number(), parsed.HasNote())
	// Output:
	// 4 IV
	// 53 false
}

func ExampleDegree_circularCounting() {
	d := degree.Seventh()
	next, wrapped := d.MustNext(7)
	previous, previousWrapped := d.MustPrevious(7)
	shifted, cycles := d.MustShift(7, -9)
	up := d.MustDistanceUp(degree.Fifth(), 7)
	down := d.MustDistanceDown(degree.Fifth(), 7)
	shortest := d.MustShortestDistance(degree.Fifth(), 7)
	fmt.Println(next.Number(), wrapped)
	fmt.Println(previous.Number(), previousWrapped)
	fmt.Println(shifted.Number(), cycles)
	fmt.Println(up, down, shortest)
	for x := range degree.MustAll(3) {
		fmt.Println(x.Number())
		if x.Number() == 2 {
			break
		}
	}
	// Output:
	// 1 true
	// 6 false
	// 5 -1
	// 5 2 -2
	// 1
	// 2
}

func ExampleDegree_normalizeAndDenormalize() {
	continuous := degree.MustNew(8)
	normalized, cycles := continuous.MustNormalize(7)
	fmt.Println(normalized.Number(), cycles)
	fmt.Println(normalized.MustDenormalize(7, cycles).Number())
	// Output:
	// 1 1
	// 8
}

func ExampleDegree_noteAccess() {
	n, err := note.FromASPN(spelling.A(), 4, cents.Zero(), settings.Standard())
	if err != nil {
		panic(err)
	}
	d := degree.MustNewWithNote(4, n)
	read := d.Note()
	fmt.Println(read.MustFormatName())
	fmt.Printf("%.4f\n", read.MustFrequency())
	// Output:
	// A
	// 440.0000
}

func ExampleDegree_changeNoteAndNumber() {
	n, err := note.FromSpelling(spelling.B(), cents.Zero(), settings.Standard())
	if err != nil {
		panic(err)
	}
	if err := n.Lower(); err != nil {
		panic(err)
	}
	d := degree.MustNew(4)
	d.MustSetNote(n)
	d.MustTransformNote(func(x note.Note) (note.Note, error) {
		if err := x.Raise(); err != nil {
			return note.Note{}, err
		}
		return x, nil
	})
	fmt.Println(d.Number(), d.Note().MustFormatName())
	d.MustRemoveNote()
	fmt.Println(d.Number(), d.HasNote())
	d.MustSetNote(n).MustSetNumber(5)
	fmt.Println(d.Number(), d.Note().MustFormatName())
	// Output:
	// 4 B
	// 4 false
	// 5 B♭
}

func ExampleDegree_compareAndSort() {
	without := degree.MustNew(4)
	n, err := note.FromSpelling(spelling.B(), cents.Zero(), nil)
	if err != nil {
		panic(err)
	}
	with := degree.MustNewWithNote(4, n)
	fmt.Println(without.Equal(degree.MustNew(4)))
	fmt.Println(without.SameNumber(with), without.SameNote(with), without.Compare(with))
	values := []degree.Degree{degree.Fifth(), with, without}
	slices.SortFunc(values, degree.Degree.Compare)
	for _, d := range values {
		fmt.Println(d.MustDecimal())
	}
	// Output:
	// true
	// true false -1
	// 4
	// 4
	// 5
}

func ExampleDegree_independentCopies() {
	n, err := note.FromSpelling(spelling.B(), cents.Zero(), nil)
	if err != nil {
		panic(err)
	}
	original := degree.MustNewWithNote(4, n)
	clone := original
	cloneNote := clone.Note()
	if err := cloneNote.Raise(); err != nil {
		panic(err)
	}
	clone.MustSetNote(cloneNote).MustSetNumber(5)
	fmt.Println(original.Number(), original.Note().Equal(n))
	fmt.Println(clone.Number(), clone.HasNote())
	original.MustRemoveNote()
	fmt.Println(original.HasNote(), clone.HasNote())
	// Output:
	// 4 true
	// 5 true
	// false true
}

func ExampleDegree_transformError() {
	n, err := note.FromSpelling(spelling.B(), cents.Zero(), nil)
	if err != nil {
		panic(err)
	}
	d := degree.MustNewWithNote(4, n)
	before := d
	cause := note.ErrNoOctave
	err = d.TransformNote(func(note.Note) (note.Note, error) {
		return note.Note{}, cause
	})
	fmt.Println(errors.Is(err, degree.ErrTransformFailed), errors.Is(err, cause))
	fmt.Println(d.Equal(before), d.Number())
	// Output:
	// true true
	// true 4
}
