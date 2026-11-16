package degree

import (
	"iter"

	"github.com/go-muse/musex/note"
)

// MustNew is like New but panics with its error.
func MustNew(number int) Degree {
	return must(New(number))
}

// MustNewWithNote is like NewWithNote but panics with its error.
func MustNewWithNote(number int, n note.Note) Degree {
	return must(NewWithNote(number, n))
}

// MustSetNote is like SetNote but panics with its error and returns d.
func (d *Degree) MustSetNote(n note.Note) *Degree {
	if err := d.SetNote(n); err != nil {
		panic(err)
	}

	return d
}

// MustRemoveNote is like RemoveNote but panics with its error and returns d.
func (d *Degree) MustRemoveNote() *Degree {
	if err := d.RemoveNote(); err != nil {
		panic(err)
	}

	return d
}

// MustTransformNote is like TransformNote but panics with its error and
// returns d.
func (d *Degree) MustTransformNote(t Transform) *Degree {
	if err := d.TransformNote(t); err != nil {
		panic(err)
	}

	return d
}

// MustSetNumber is like SetNumber but panics with its error and returns d.
func (d *Degree) MustSetNumber(number int) *Degree {
	if err := d.SetNumber(number); err != nil {
		panic(err)
	}

	return d
}

// MustDecimal is like Decimal but panics with its error.
func (d Degree) MustDecimal() string {
	return must(d.Decimal())
}

// MustRoman is like Roman but panics with its error.
func (d Degree) MustRoman() string {
	return must(d.Roman())
}

// MustParse is like Parse but panics with its error.
func MustParse(text string) Degree {
	return must(Parse(text))
}

// MustInCircle is like InCircle but panics with its error.
func (d Degree) MustInCircle(k int) bool {
	return must(d.InCircle(k))
}

// MustNext is like Next but panics with its error.
func (d Degree) MustNext(k int) (next Degree, wrapped bool) {
	return must3(d.Next(k))
}

// MustPrevious is like Previous but panics with its error.
func (d Degree) MustPrevious(k int) (previous Degree, wrapped bool) {
	return must3(d.Previous(k))
}

// MustShift is like Shift but panics with its error.
func (d Degree) MustShift(k, n int) (shifted Degree, cycles int) {
	return must3(d.Shift(k, n))
}

// MustDistanceUp is like DistanceUp but panics with its error.
func (d Degree) MustDistanceUp(to Degree, k int) int {
	return must(d.DistanceUp(to, k))
}

// MustDistanceDown is like DistanceDown but panics with its error.
func (d Degree) MustDistanceDown(to Degree, k int) int {
	return must(d.DistanceDown(to, k))
}

// MustShortestDistance is like ShortestDistance but panics with its error.
func (d Degree) MustShortestDistance(to Degree, k int) int {
	return must(d.ShortestDistance(to, k))
}

// MustNormalize is like Normalize but panics with its error.
func (d Degree) MustNormalize(k int) (normalized Degree, cycles int) {
	return must3(d.Normalize(k))
}

// MustDenormalize is like Denormalize but panics with its error.
func (d Degree) MustDenormalize(k, cycles int) Degree {
	return must(d.Denormalize(k, cycles))
}

// MustAll is like All but panics with its error.
func MustAll(k int) iter.Seq[Degree] {
	return must(All(k))
}

// must returns value or panics with err.
func must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}

	return value
}

// must3 returns first and second or panics with err.
func must3[A, B any](first A, second B, err error) (A, B) {
	if err != nil {
		panic(err)
	}

	return first, second
}
