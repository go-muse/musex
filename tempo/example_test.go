package tempo_test

import (
	"errors"
	"fmt"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempo/term"
)

// A tempo marking, its speed, the two equalities, and another unit.
func Example() {
	f := fraction.FromInt
	dq := duration.MustNewWritten(duration.Quarter, 1)
	e := duration.Eighth.MustWritten()

	t := tempo.MustNew(dq, f(60))
	fmt.Println(t, t.MustSpeed(), t.MustQuartersPerMinute())
	u := tempo.MustNew(e, f(180))
	fmt.Println(t.Equal(u), t.MustEqualSpeed(u), t.MustWithUnit(e))
	// Output:
	// 4.=60 45/2 90/1
	// false true 8=180
}

// Seconds of a duration and system time with an explicit rounding mode.
func ExampleTempo_Seconds() {
	half := duration.Half.MustDuration()
	fmt.Println(tempo.MustNewQuarter(fraction.FromInt(60)).MustSeconds(half))
	s := tempo.MustNewQuarter(fraction.MustNew(185, 2)).MustSeconds(duration.Quarter.MustDuration())
	fmt.Println(s, s.MustTimeDuration(rounding.HalfUp))
	// Output:
	// 2/1
	// 24/37 648.648649ms
}

// A metric modulation: the new dotted quarter equals the old quarter.
func ExampleTempo_Modulate() {
	q := duration.Quarter.MustWritten()
	dq := duration.MustNewWritten(duration.Quarter, 1)
	fmt.Println(tempo.MustNewQuarter(fraction.FromInt(120)).MustModulate(dq, q.MustDuration()))
	// Output:
	// 4.=120
}

// The formula of bars, beats, bpm and time.
func ExampleTempo_Bars() {
	whole := duration.MustNew(1, 1)
	twoMinutes := tempo.MustSecondsFromMinutes(fraction.FromInt(2))
	fmt.Println(tempo.MustNewQuarter(fraction.FromInt(120)).MustBars(whole, twoMinutes))
	// Output:
	// 60/1
}

// An approximate tempo and the explicit exact tempo of its number.
func ExampleApprox() {
	q := duration.Quarter.MustWritten()
	a := tempo.MustNewApprox(q, fraction.FromInt(108))
	fmt.Println(a, a.MustTempo())
	// Output:
	// 4≈108 4=108
}

// A tempo range, its exact middle, and whether a tempo is in it by speed.
func ExampleRange() {
	f := fraction.FromInt
	q := duration.Quarter.MustWritten()
	e := duration.Eighth.MustWritten()

	r := tempo.MustNewRange(e, f(132), f(144))
	fmt.Println(r, r.MustMidTempo(), r.MustMidTempo().MustWithUnit(q))
	fmt.Println(r.MustContainsTempo(tempo.MustNewQuarter(f(69))))
	// Output:
	// 8=132-144 8=138 4=69
	// true
}

// The metronome scale: both neighbors and the division by a mode.
func ExampleScaleNeighbors() {
	f := fraction.FromInt
	lo, hi, cmp := tempo.MustScaleNeighbors(f(110))
	fmt.Println(lo, hi, cmp, tempo.MustNearestScaleDivision(f(110), rounding.HalfEven))
	// Output:
	// 108 112 0 112
}

// Words of a list, an indication, and its text.
func ExampleIndication() {
	f := fraction.FromInt
	q := duration.Quarter.MustWritten()

	l := term.EnglishWikipedia()
	allegro := l.MustParse("allegro")
	fmt.Println(allegro, allegro.MustPlace(), allegro.MustGroup())
	fmt.Println(l.MustAt(f(115)), l.MustParse("Vivo"))

	ind := tempo.MustIndicationOf(allegro).MustWithTempo(tempo.MustNewQuarter(f(120)))
	fmt.Println(ind, tempo.MustParseIndication("Allegro, 4=120", l).Equal(ind))
	fmt.Println(tempo.MustRangeOf(allegro, q))
	// Output:
	// Allegro 15 fast
	// [Moderato Allegretto] Vivace
	// Allegro, 4=120 true
	// 4=120-156
}

// A user list with names that need quotes in the text of an indication.
func ExampleParseIndication() {
	f := fraction.FromInt
	mine := term.MustNewList(term.Row{Entries: []term.Entry{{Name: "4=60"}, {Name: "Скоро, но не слишком"}}})
	odd := tempo.MustIndicationOf(mine.MustParse("4=60"))
	fmt.Println(odd, tempo.MustIndicationOfTempo(tempo.MustNewQuarter(f(60))))
	soon := tempo.MustIndicationOf(mine.MustParse("скоро, но не слишком")).MustWithTempo(tempo.MustNewQuarter(f(120)))
	fmt.Println(soon, tempo.MustParseIndication(soon.String(), mine).Equal(soon))
	// Output:
	// "4=60" 4=60
	// Скоро, но не слишком, 4=120 true
}

// Failures are distinguished with errors.Is, including the causes in the
// packages below.
func Example_errors() {
	_, err := tempo.NewQuarter(fraction.Zero())
	fmt.Println(errors.Is(err, tempo.ErrNoTempo))

	_, _, _, err = tempo.ScaleNeighbors(fraction.FromInt(250))
	fmt.Println(errors.Is(err, tempo.ErrOffScale))

	_, err = tempo.MustNewQuarter(fraction.MustNew(1, 4)).Quantize(fraction.One(), rounding.Down)
	fmt.Println(errors.Is(err, tempo.ErrQuantizedToZero))

	_, err = tempo.ZeroSeconds().TimeDuration(rounding.Mode(0))
	fmt.Println(errors.Is(err, tempo.ErrInvalidArgument), errors.Is(err, rounding.ErrUnset))

	_, err = tempo.NewRange(duration.Quarter.MustWritten(), fraction.FromInt(132), fraction.FromInt(120))
	fmt.Println(errors.Is(err, tempo.ErrInvalidArgument))

	_, err = term.EnglishWikipedia().Parse("Fast")
	fmt.Println(errors.Is(err, term.ErrInvalidText))

	var t0 tempo.Tempo
	_, err = t0.BPM()
	fmt.Println(t0, errors.Is(err, tempo.ErrUnset))
	// Output:
	// true
	// true
	// true
	// true true
	// true
	// true
	// Tempo(unset) true
}
