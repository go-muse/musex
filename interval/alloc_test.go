package interval_test

import (
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/require"
)

// T-7, allocations (architecture 6): "0" rows on the successful path —
// creation, predefined intervals, reading, properties, comparisons (row 1);
// the arithmetic, names and written pitches of 4.6, 4.8, 4.9 on ordinary
// values (row 3); Steps, FromSteps, EnharmonicallyEqual (positional row).
// The tests do not call t.Parallel: testing.AllocsPerRun counts the
// allocations of the whole process, so they run in the sequential phase of
// the package tests (as note/alloc_test.go).

// Sinks keep the results of the measured calls.
var (
	sinkInterval interval.Interval
	sinkSecond   interval.Interval
	sinkBool     bool
	sinkInt      int
	sinkAlt      alteration.Alteration
	sinkDir      interval.Direction
	sinkQuality  interval.Quality
	sinkGroup    interval.Group
	sinkSpelling spelling.Spelling
	sinkPitch    pitch.Pitch
	sinkStep     system.Step
	sinkErr      error
)

func assertNoAllocs(t *testing.T, name string, f func()) {
	t.Helper()

	allocs := testing.AllocsPerRun(100, f)
	require.Zero(t, allocs, "%s allocates", name)
}

func TestAllocsCreation(t *testing.T) {
	assertNoAllocs(t, "New", func() { sinkInterval, sinkErr = interval.New(interval.Major, 3) })
	assertNoAllocs(t, "NewMultiple", func() { sinkInterval, sinkErr = interval.NewMultiple(2, interval.Diminished, 5) })
	assertNoAllocs(t, "FromDeviation", func() { sinkInterval, sinkErr = interval.FromDeviation(3, sinkAlt) })
	assertNoAllocs(t, "MustNew", func() { sinkInterval = interval.MustNew(interval.Minor, 6) })

	for _, p := range predefinedTable {
		assertNoAllocs(t, p.name, func() { sinkInterval = p.fn() })
	}
}

func TestAllocsReading(t *testing.T) {
	values := []interval.Interval{iv("M3"), iv("↓dd12"), iv("n3"), iv("{+1/4}17"), iv("P1"), {}}

	readers := map[string]func(i interval.Interval){
		"IsValid":          func(i interval.Interval) { sinkBool = i.IsValid() },
		"Direction":        func(i interval.Interval) { sinkDir = i.Direction() },
		"IsAscending":      func(i interval.Interval) { sinkBool = i.IsAscending() },
		"IsDescending":     func(i interval.Interval) { sinkBool = i.IsDescending() },
		"Size":             func(i interval.Interval) { sinkInt = i.Size() },
		"Deviation":        func(i interval.Interval) { sinkAlt = i.Deviation() },
		"Quality":          func(i interval.Interval) { sinkQuality = i.Quality() },
		"Multiplicity":     func(i interval.Interval) { sinkInt = i.Multiplicity() },
		"IsIntegerQuality": func(i interval.Interval) { sinkBool = i.IsIntegerQuality() },
		"Group":            func(i interval.Interval) { sinkGroup = i.Group() },
		"SimpleSize":       func(i interval.Interval) { sinkInt = i.SimpleSize() },
		"Octaves":          func(i interval.Interval) { sinkInt = i.Octaves() },
	}

	for name, f := range propertyReaders() {
		readers[name] = func(i interval.Interval) { sinkBool = f(i) }
	}

	for _, v := range values {
		for name, read := range readers {
			assertNoAllocs(t, name, func() { read(v) })
		}

		if v.IsValid() {
			assertNoAllocs(t, "Simple", func() { sinkInterval, sinkErr = v.Simple() })
		}
	}
}

func TestAllocsCompare(t *testing.T) {
	a, b := iv("M3"), iv("↓d4")

	assertNoAllocs(t, "Equal", func() { sinkBool = a.Equal(b) })
	assertNoAllocs(t, "SameSize", func() { sinkBool = a.SameSize(b) })
	assertNoAllocs(t, "SameQuality", func() { sinkBool = a.SameQuality(b) })
	assertNoAllocs(t, "SameDirection", func() { sinkBool = a.SameDirection(b) })
	assertNoAllocs(t, "EqualUndirected", func() { sinkBool = a.EqualUndirected(b) })
	assertNoAllocs(t, "Compare", func() { sinkInt = a.Compare(b) })
}

func TestAllocsArithmetic(t *testing.T) {
	i, j := iv("M10"), iv("↓m3")

	assertNoAllocs(t, "WithDirection", func() { sinkInterval, sinkErr = i.WithDirection(interval.Down) })
	assertNoAllocs(t, "Up", func() { sinkInterval, sinkErr = j.Up() })
	assertNoAllocs(t, "Down", func() { sinkInterval, sinkErr = i.Down() })
	assertNoAllocs(t, "Opposite", func() { sinkInterval, sinkErr = i.Opposite() })
	assertNoAllocs(t, "AddOctaves", func() { sinkInterval, sinkErr = i.AddOctaves(2) })
	assertNoAllocs(t, "Expand", func() { sinkInterval, sinkErr = i.Expand() })
	assertNoAllocs(t, "Contract", func() { sinkInterval, sinkErr = i.Contract() })
	assertNoAllocs(t, "ExpandBy", func() { sinkInterval, sinkErr = i.ExpandBy(2) })
	assertNoAllocs(t, "ContractBy", func() { sinkInterval, sinkErr = i.ContractBy(2) })
	assertNoAllocs(t, "ExpandByAlteration", func() { sinkInterval, sinkErr = i.ExpandByAlteration(alt(1, 2)) })
	assertNoAllocs(t, "ContractByAlteration", func() { sinkInterval, sinkErr = i.ContractByAlteration(alt(1, 2)) })
	assertNoAllocs(t, "Invert", func() { sinkInterval, sinkErr = i.Invert() })
	assertNoAllocs(t, "InversionClass", func() { sinkInterval, sinkSecond, sinkErr = i.InversionClass() })
	assertNoAllocs(t, "Add", func() { sinkInterval, sinkErr = i.Add(j) })
	assertNoAllocs(t, "Sub", func() { sinkInterval, sinkErr = i.Sub(j) })
	assertNoAllocs(t, "Mul", func() { sinkInterval, sinkErr = j.Mul(5) })
}

func TestAllocsSpellings(t *testing.T) {
	x, y, i := spelling.FSharp(), spelling.BFlat(), iv("↓A4")

	assertNoAllocs(t, "Between", func() { sinkInterval, sinkErr = interval.Between(x, y, interval.Down) })
	assertNoAllocs(t, "BothWays", func() { sinkInterval, sinkSecond, sinkErr = interval.BothWays(x, y) })
	assertNoAllocs(t, "Transpose", func() { sinkSpelling, sinkErr = interval.Transpose(x, i) })
	assertNoAllocs(t, "FifthsOf", func() { sinkInt, sinkErr = interval.FifthsOf(x) })
	assertNoAllocs(t, "FifthsBetween", func() { sinkInt, sinkErr = interval.FifthsBetween(x, y) })
	assertNoAllocs(t, "FromFifths", func() { sinkInterval, sinkErr = interval.FromFifths(-9) })
	assertNoAllocs(t, "Fifths", func() { sinkInt, sinkErr = i.Fifths() })
	assertNoAllocs(t, "Reflect", func() { sinkSpelling, sinkErr = interval.Reflect(y, x) })
}

func TestAllocsPitches(t *testing.T) {
	tr := trad()
	x, y, i := pt(spelling.BFlat(), 3), wp(tr, spelling.BSharp(), 4), iv("M10")

	assertNoAllocs(t, "BetweenPitches", func() { sinkInterval, sinkErr = interval.BetweenPitches(tr, x, y) })
	assertNoAllocs(t, "TransposePitch", func() { sinkPitch, sinkErr = interval.TransposePitch(tr, x, i) })
	assertNoAllocs(t, "ReflectPitch", func() { sinkPitch, sinkErr = interval.ReflectPitch(tr, x, y) })
}

func TestAllocsPositional(t *testing.T) {
	tr, irr := trad(), irregular()
	i, j := iv("A4"), iv("↓d5")

	assertNoAllocs(t, "Steps", func() { sinkStep, sinkErr = i.Steps(tr, noStart) })
	assertNoAllocs(t, "Steps from a start", func() { sinkStep, sinkErr = j.Steps(irr, spelling.D()) })
	assertNoAllocs(t, "SignedSteps", func() { sinkStep, sinkErr = j.SignedSteps(tr, spelling.C()) })
	assertNoAllocs(t, "FromSteps", func() { sinkInterval, sinkErr = interval.FromSteps(tr, 4, 6, noStart) })
	assertNoAllocs(t, "FromSteps from a start", func() { sinkInterval, sinkErr = interval.FromSteps(irr, 2, 3, spelling.C()) })
	assertNoAllocs(t, "EnharmonicallyEqual", func() {
		sinkBool, sinkErr = interval.EnharmonicallyEqual(tr, i, noStart, j, noStart)
	})
}
