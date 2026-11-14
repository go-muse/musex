package interval_test

import (
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T-5: intervals between notes and transposition of a note (OP-32, OP-33;
// P-15; KP-15; F-4, F-7; architecture 4.10).

// sameNote reports whether got has the record and correction of want and
// the source of src (note.Equal ignores the source).
func sameNote(t *testing.T, want, got, src note.Note) {
	t.Helper()

	assert.True(t, got.Equal(want), "got %v, want %v", got, want)
	assert.Equal(t, want.HasOctave(), got.HasOctave())
	assert.True(t, got.SameCents(src), "the correction is kept")
	assert.Equal(t, src.IsBound(), got.IsBound())

	if src.IsBound() {
		assert.True(t, got.SameSource(src), "the source is kept")
	}
}

// unchanged checks that the input note keeps its record, correction and
// source (P-1, F-7).
func unchanged(t *testing.T, before, after note.Note) {
	t.Helper()

	assert.True(t, after.Equal(before), "the input note changed: %v", after)
	assert.True(t, after.SameCents(before))
	assert.Equal(t, before.IsBound(), after.IsBound())

	if before.IsBound() {
		assert.True(t, after.SameSource(before))
	}
}

// ---------------------------------------------------------------------------
// BetweenNotes and BetweenNotesIn (OP-32).

func TestNotesBetween(t *testing.T) {
	t.Parallel()

	src := std()
	c4 := noteAt(t, pt(spelling.C(), 4), 30, src)
	e4 := noteAt(t, pt(spelling.E(), 4), -20, src)

	t.Run("cents do not take part (KP-15)", func(t *testing.T) {
		t.Parallel()

		for _, d := range []interval.Direction{interval.Up, interval.Down, interval.DirectionUnset, interval.Direction(9)} {
			got, err := interval.BetweenNotes(c4, e4, d)
			require.NoError(t, err, "direction %v is not used with octaves", d)
			assert.True(t, got.Equal(iv("M3")), "got %v", got)

			got, err = interval.BetweenNotes(e4, c4, d)
			require.NoError(t, err)
			assert.True(t, got.Equal(iv("↓M3")), "got %v", got)
		}
	})

	t.Run("two sources with equal systems", func(t *testing.T) {
		t.Parallel()

		x := noteAt(t, pt(spelling.BSharp(), 4), 0, std())
		y := noteAt(t, pt(spelling.C(), 4), 0, std())

		got, err := interval.BetweenNotes(x, y, interval.DirectionUnset)
		require.NoError(t, err)
		assert.True(t, got.Equal(iv("d2")), "B♯3 → C4 by writing: %v", got)
	})

	t.Run("notes without octave", func(t *testing.T) {
		t.Parallel()

		x := noteName(t, spelling.C(), 5, src)
		y := noteName(t, spelling.E(), -7, nil)

		got, err := interval.BetweenNotes(x, y, interval.Up)
		require.NoError(t, err)
		assert.True(t, got.Equal(iv("M3")))

		got, err = interval.BetweenNotes(x, y, interval.Down)
		require.NoError(t, err)
		assert.True(t, got.Equal(iv("↓m6")))
	})

	t.Run("notes without octave: the source is not read", func(t *testing.T) {
		t.Parallel()

		off := newSwitch(set24())
		x := noteName(t, spelling.C(), 0, off)
		y := noteName(t, eHalfFlat(), 0, off)
		off.set(nil)

		got, err := interval.BetweenNotes(x, y, interval.Up)
		require.NoError(t, err)
		assert.True(t, got.Equal(iv("n3")))
	})

	t.Run("unbound notes need an explicit system (KP-15)", func(t *testing.T) {
		t.Parallel()

		x := noteAt(t, pt(spelling.C(), 4), 0, nil)
		y := noteAt(t, pt(spelling.E(), 4), 0, nil)

		got, err := interval.BetweenNotes(x, y, interval.Up)
		assertErr(t, err, interval.ErrNoSystem)
		assert.Equal(t, interval.Interval{}, got)

		got, err = interval.BetweenNotesIn(trad(), x, y, interval.Up)
		require.NoError(t, err)
		assert.True(t, got.Equal(iv("M3")))
	})

	t.Run("different systems (KP-15)", func(t *testing.T) {
		t.Parallel()

		x := noteAt(t, pt(spelling.C(), 4), 0, std())
		y := noteAt(t, pt(spelling.E(), 4), 0, set19())

		got, err := interval.BetweenNotes(x, y, interval.Up)
		assertErr(t, err, interval.ErrDifferentSystems)
		assert.Equal(t, interval.Interval{}, got)

		got, err = interval.BetweenNotesIn(pos19(), x, y, interval.Up)
		require.NoError(t, err)
		assert.True(t, got.Equal(iv("M3")))
	})

	t.Run("the system of the sets", func(t *testing.T) {
		t.Parallel()

		s := set19()
		x := noteAt(t, wp(pos19(), spelling.BSharp(), 3), 0, s)
		y := noteAt(t, wp(pos19(), spelling.C(), 4), 0, s)

		got, err := interval.BetweenNotes(x, y, interval.DirectionUnset)
		require.NoError(t, err)
		assert.True(t, got.Equal(iv("d2")))
	})

	t.Run("explicit system: sources are not read", func(t *testing.T) {
		t.Parallel()

		off := newSwitch(std())
		x := noteAt(t, pt(spelling.C(), 4), 0, off)
		y := noteAt(t, pt(spelling.E(), 4), 0, off)
		off.set(nil)

		got, err := interval.BetweenNotesIn(trad(), x, y, interval.DirectionUnset)
		require.NoError(t, err)
		assert.True(t, got.Equal(iv("M3")))

		got, err = interval.BetweenNotesIn(trad(), noteName(t, spelling.C(), 0, nil), noteName(t, spelling.E(), 0, nil), interval.Down)
		require.NoError(t, err)
		assert.True(t, got.Equal(iv("↓m6")))
	})
}

func TestNotesBetweenErrors(t *testing.T) {
	t.Parallel()

	src := std()
	c4 := noteAt(t, pt(spelling.C(), 4), 0, src)
	e4 := noteAt(t, pt(spelling.E(), 4), 0, src)
	e := noteName(t, spelling.E(), 0, src)
	c := noteName(t, spelling.C(), 0, src)
	free4 := noteAt(t, pt(spelling.C(), 4), 0, nil)

	unsetSrc := newSwitch(std())
	cOff := noteAt(t, pt(spelling.C(), 4), 0, unsetSrc)
	eOff := noteAt(t, pt(spelling.E(), 4), 0, unsetSrc)
	unsetSrc.set(nil)

	narrowed := newSwitch(set24())
	cNar := noteAt(t, pt(spelling.C(), 4), 0, narrowed)
	ehNar := noteAt(t, pt(eHalfFlat(), 4), 0, narrowed)
	narrowed.set(std())

	cMinus := noteAt(t, pt(spelling.MustNew(letter.C, ai(-mag)), 4), 0, src)
	cPlus := noteAt(t, pt(spelling.MustNew(letter.C, ai(mag)), 4), 0, src)

	tests := []struct {
		name    string
		x, y    note.Note
		d       interval.Direction
		want    error
		causes  []error
		notWant []error
	}{
		{name: "unset x", x: note.Note{}, y: e4, d: interval.Up, want: interval.ErrUnset},
		{name: "unset y", x: c4, y: note.Note{}, d: interval.Up, want: interval.ErrUnset},
		{name: "unset before no octave", x: note.Note{}, y: e, d: interval.DirectionUnset, want: interval.ErrUnset},
		{name: "one note without octave (KP-15)", x: c4, y: e, d: interval.Up, want: interval.ErrNoOctave},
		{name: "one note without octave and no direction (KP-15)", x: c4, y: e, d: interval.DirectionUnset, want: interval.ErrNoOctave},
		{name: "the other note without octave", x: c, y: e4, d: interval.Down, want: interval.ErrNoOctave},
		{name: "no octave before no system", x: free4, y: e, d: interval.Up, want: interval.ErrNoOctave},
		{name: "no direction without octaves", x: c, y: e, d: interval.DirectionUnset, want: interval.ErrNoDirection},
		{name: "direction out of the enumeration", x: c, y: e, d: interval.Direction(3), want: interval.ErrNoDirection},
		{name: "unbound x", x: free4, y: e4, d: interval.Up, want: interval.ErrNoSystem},
		{name: "unbound y", x: c4, y: noteAt(t, pt(spelling.E(), 4), 0, nil), d: interval.Up, want: interval.ErrNoSystem},
		{name: "unbound before unset settings", x: free4, y: eOff, d: interval.Up, want: interval.ErrNoSystem},
		{name: "unset settings", x: cOff, y: eOff, d: interval.Up,
			causes: []error{note.ErrUnsetSettings, settings.ErrUnset}},
		{name: "unset settings of y", x: c4, y: eOff, d: interval.Up,
			causes: []error{note.ErrUnsetSettings, settings.ErrUnset}},
		{name: "unset settings before different systems", x: noteAt(t, pt(spelling.C(), 4), 0, set19()), y: eOff, d: interval.Up,
			causes: []error{note.ErrUnsetSettings, settings.ErrUnset}},
		{name: "different systems", x: c4, y: noteAt(t, pt(spelling.E(), 4), 0, set19()), d: interval.Up, want: interval.ErrDifferentSystems},
		{name: "name not admissible in the system of the sets", x: cNar, y: ehNar, d: interval.Up,
			causes: []error{note.ErrInadmissible, system.ErrInadmissible}},
		{name: "deviation beyond M", x: cMinus, y: cPlus, d: interval.Up, want: interval.ErrOutOfRange},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := interval.BetweenNotes(tt.x, tt.y, tt.d)
			assertErr(t, err, tt.want, tt.causes...)
			assert.Equal(t, interval.Interval{}, got)

			for _, e := range tt.notWant {
				assert.NotErrorIs(t, err, e)
			}
		})
	}
}

func TestNotesBetweenInErrors(t *testing.T) {
	t.Parallel()

	src := std()
	c4 := noteAt(t, pt(spelling.C(), 4), 0, src)
	e4 := noteAt(t, pt(spelling.E(), 4), 0, src)
	e := noteName(t, spelling.E(), 0, src)
	c := noteName(t, spelling.C(), 0, nil)
	eh := noteName(t, eHalfFlat(), 0, nil)
	eh4 := noteAt(t, pt(eHalfFlat(), 4), 0, nil)

	tests := []struct {
		name   string
		sys    *system.System
		x, y   note.Note
		d      interval.Direction
		want   error
		causes []error
	}{
		{name: "nil system", sys: nil, x: c4, y: e4, d: interval.Up, want: interval.ErrUnset},
		{name: "unset system", sys: new(system.System), x: c4, y: e4, d: interval.Up, want: interval.ErrUnset},
		{name: "unset note", sys: trad(), x: note.Note{}, y: e4, d: interval.Up, want: interval.ErrUnset},
		{name: "one note without octave", sys: trad(), x: c4, y: e, d: interval.Up, want: interval.ErrNoOctave},
		{name: "no direction without octaves", sys: trad(), x: c, y: e, d: interval.DirectionUnset, want: interval.ErrNoDirection},
		{name: "names without octave checked in the explicit system", sys: trad(), x: c, y: eh, d: interval.Up,
			causes: []error{system.ErrInadmissible}},
		{name: "names with octave checked in the explicit system", sys: trad(), x: c4, y: eh4, d: interval.Up,
			causes: []error{system.ErrInadmissible}},
		{name: "no direction before inadmissible names", sys: trad(), x: c, y: eh, d: interval.DirectionUnset,
			want: interval.ErrNoDirection},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := interval.BetweenNotesIn(tt.sys, tt.x, tt.y, tt.d)
			assertErr(t, err, tt.want, tt.causes...)
			assert.NotErrorIs(t, err, note.ErrInadmissible, "an explicit system gives causes of system")
			assert.Equal(t, interval.Interval{}, got)
		})
	}

	t.Run("names without octave in a system that admits them", func(t *testing.T) {
		t.Parallel()

		got, err := interval.BetweenNotesIn(qt24(), c, eh, interval.Up)
		require.NoError(t, err)
		assert.True(t, got.Equal(iv("n3")))
	})
}

// ---------------------------------------------------------------------------
// TransposeNote and TransposeNoteIn (OP-33).

func TestNotesTranspose(t *testing.T) {
	t.Parallel()

	t.Run("KP-15: cents and source are kept, the input does not change", func(t *testing.T) {
		t.Parallel()

		src := std()
		x := noteAt(t, pt(spelling.C(), 4), 30, src)
		before := x

		got, err := interval.TransposeNote(x, iv("M3"))
		require.NoError(t, err)
		sameNote(t, noteAt(t, pt(spelling.E(), 4), 30, src), got, x)
		unchanged(t, before, x)
		assert.True(t, x.Equal(noteAt(t, pt(spelling.C(), 4), 30, src)), "x is still C4 +30")
	})

	tests := []struct {
		name string
		src  func() settings.Source
		x    func(t *testing.T, src settings.Source) note.Note
		i    string
		want func(t *testing.T, src settings.Source) note.Note
	}{
		{"without octave by down m2 (KP-15)", func() settings.Source { return std() },
			func(t *testing.T, s settings.Source) note.Note { return noteName(t, spelling.C(), 12, s) }, "↓m2",
			func(t *testing.T, s settings.Source) note.Note { return noteName(t, spelling.B(), 12, s) }},
		{"unbound without octave by n3: no system needed", func() settings.Source { return nil },
			func(t *testing.T, s settings.Source) note.Note { return noteName(t, spelling.C(), -3, s) }, "n3",
			func(t *testing.T, s settings.Source) note.Note { return noteName(t, eHalfFlat(), -3, s) }},
		{"B sharp 3 by d2", func() settings.Source { return std() },
			func(t *testing.T, s settings.Source) note.Note {
				return noteAt(t, wp(trad(), spelling.BSharp(), 3), 0, s)
			}, "d2",
			func(t *testing.T, s settings.Source) note.Note { return noteAt(t, pt(spelling.C(), 4), 0, s) }},
		{"C4 by down A1", func() settings.Source { return std() },
			func(t *testing.T, s settings.Source) note.Note { return noteAt(t, pt(spelling.C(), 4), 7, s) }, "↓A1",
			func(t *testing.T, s settings.Source) note.Note {
				return noteAt(t, wp(trad(), spelling.CFlat(), 4), 7, s)
			}},
		{"19: C4 by A4", func() settings.Source { return set19() },
			func(t *testing.T, s settings.Source) note.Note { return noteAt(t, pt(spelling.C(), 4), 0, s) }, "A4",
			func(t *testing.T, s settings.Source) note.Note { return noteAt(t, pt(spelling.FSharp(), 4), 0, s) }},
		{"24: C4 by n3", func() settings.Source { return set24() },
			func(t *testing.T, s settings.Source) note.Note { return noteAt(t, pt(spelling.C(), 4), 0, s) }, "n3",
			func(t *testing.T, s settings.Source) note.Note { return noteAt(t, pt(eHalfFlat(), 4), 0, s) }},
		{"C4 by down P15", func() settings.Source { return std() },
			func(t *testing.T, s settings.Source) note.Note { return noteAt(t, pt(spelling.C(), 4), 1, s) }, "↓P15",
			func(t *testing.T, s settings.Source) note.Note { return noteAt(t, pt(spelling.C(), 2), 1, s) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			src := tt.src()
			x := tt.x(t, src)
			before := x

			got, err := interval.TransposeNote(x, iv(tt.i))
			require.NoError(t, err)
			sameNote(t, tt.want(t, src), got, x)
			unchanged(t, before, x)
		})
	}
}

func TestNotesTransposeIn(t *testing.T) {
	t.Parallel()

	t.Run("unbound with octave in an explicit system (KP-15)", func(t *testing.T) {
		t.Parallel()

		x := noteAt(t, pt(spelling.C(), 4), 30, nil)
		got, err := interval.TransposeNoteIn(trad(), x, iv("M3"))
		require.NoError(t, err)
		sameNote(t, noteAt(t, pt(spelling.E(), 4), 30, nil), got, x)
	})

	t.Run("unbound in 24 by n3", func(t *testing.T) {
		t.Parallel()

		x := noteAt(t, pt(spelling.C(), 4), 0, nil)
		got, err := interval.TransposeNoteIn(qt24(), x, iv("n3"))
		require.NoError(t, err)
		sameNote(t, noteAt(t, pt(eHalfFlat(), 4), 0, nil), got, x)
	})

	t.Run("19 written octaves in the explicit system", func(t *testing.T) {
		t.Parallel()

		x := noteAt(t, wp(pos19(), spelling.BSharp(), 3), 0, nil)
		got, err := interval.TransposeNoteIn(pos19(), x, iv("m2"))
		require.NoError(t, err)
		sameNote(t, noteAt(t, wp(pos19(), spelling.CSharp(), 4), 0, nil), got, x)
	})

	t.Run("bound without octave", func(t *testing.T) {
		t.Parallel()

		src := std()
		x := noteName(t, spelling.F(), 4, src)
		got, err := interval.TransposeNoteIn(trad(), x, iv("↓A4"))
		require.NoError(t, err)
		sameNote(t, noteName(t, spelling.CFlat(), 4, src), got, x)
	})
}

func TestNotesTransposeErrors(t *testing.T) {
	t.Parallel()

	src := std()
	c4 := noteAt(t, pt(spelling.C(), 4), 0, src)
	c := noteName(t, spelling.C(), 0, src)

	off := newSwitch(std())
	c4Off := noteAt(t, pt(spelling.C(), 4), 0, off)
	cOff := noteName(t, spelling.C(), 0, off)
	off.set(nil)

	narrowed := newSwitch(set24())
	ehNar := noteAt(t, pt(eHalfFlat(), 4), 0, narrowed)
	narrowed.set(std())

	cPlus4 := noteAt(t, pt(spelling.MustNew(letter.C, ai(mag)), 4), 0, src)
	cPlus := noteName(t, spelling.MustNew(letter.C, ai(mag)), 0, src)
	cTop := noteAt(t, pitch.MustNew(spelling.C(), octave.FromASPN(maxInt)), 0, src)

	tests := []struct {
		name   string
		x      note.Note
		i      interval.Interval
		want   error
		causes []error
	}{
		{"unset note", note.Note{}, iv("M3"), interval.ErrUnset, nil},
		{"unset interval", c4, interval.Interval{}, interval.ErrUnset, nil},
		{"unbound with octave", noteAt(t, pt(spelling.C(), 4), 0, nil), iv("M3"), interval.ErrNoSystem, nil},
		{"unset settings with octave", c4Off, iv("M3"), nil, []error{note.ErrUnsetSettings, settings.ErrUnset}},
		{"unset settings without octave: replacing the record", cOff, iv("M3"), nil, []error{note.ErrUnsetSettings, settings.ErrUnset}},
		{"name of the note not admissible in its set", ehNar, iv("M2"), nil, []error{note.ErrInadmissible, system.ErrInadmissible}},
		{"result not admissible in the set (with octave)", c4, iv("n3"), nil, []error{note.ErrInadmissible, system.ErrInadmissible}},
		{"result not admissible in the set (without octave)", c, iv("n3"), nil, []error{note.ErrInadmissible, system.ErrInadmissible}},
		{"alteration beyond M with octave", cPlus4, iv("A1"), interval.ErrOutOfRange, []error{alteration.ErrOutOfRange}},
		{"alteration beyond M without octave", cPlus, iv("A1"), interval.ErrOutOfRange, []error{alteration.ErrOutOfRange}},
		{"octave beyond int", cTop, iv("P8"), interval.ErrOutOfRange, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			before := tt.x
			got, err := interval.TransposeNote(tt.x, tt.i)
			assertErr(t, err, tt.want, tt.causes...)
			assert.Equal(t, note.Note{}, got)

			if before.IsValid() {
				unchanged(t, before, tt.x)
			}
		})
	}
}

func TestNotesTransposeInErrors(t *testing.T) {
	t.Parallel()

	src := std()
	c4 := noteAt(t, pt(spelling.C(), 4), 0, src)

	off := newSwitch(std())
	c4Off := noteAt(t, pt(spelling.C(), 4), 0, off)
	off.set(nil)

	tests := []struct {
		name    string
		sys     *system.System
		x       note.Note
		i       interval.Interval
		want    error
		causes  []error
		notWant []error
	}{
		{name: "nil system", sys: nil, x: c4, i: iv("M3"), want: interval.ErrUnset},
		{name: "unset note", sys: trad(), x: note.Note{}, i: iv("M3"), want: interval.ErrUnset},
		{name: "unset interval", sys: trad(), x: c4, i: interval.Interval{}, want: interval.ErrUnset},
		{name: "name of the note not admissible in the explicit system", sys: trad(), x: noteAt(t, pt(eHalfFlat(), 4), 0, nil), i: iv("M2"),
			causes: []error{system.ErrInadmissible}, notWant: []error{note.ErrInadmissible}},
		{name: "result not admissible in the explicit system", sys: trad(), x: noteAt(t, pt(spelling.C(), 4), 0, nil), i: iv("n3"),
			causes: []error{system.ErrInadmissible}, notWant: []error{note.ErrInadmissible}},
		{name: "result without octave not admissible in the explicit system", sys: trad(), x: noteName(t, spelling.C(), 0, nil), i: iv("n3"),
			causes: []error{system.ErrInadmissible}, notWant: []error{note.ErrInadmissible}},
		{name: "result admissible in the explicit system, not in the set of the note", sys: qt24(), x: c4, i: iv("n3"),
			causes: []error{note.ErrInadmissible, system.ErrInadmissible}},
		{name: "unset settings when replacing the record", sys: trad(), x: c4Off, i: iv("M3"),
			causes: []error{note.ErrUnsetSettings, settings.ErrUnset}},
		{name: "alteration beyond M", sys: trad(), x: noteAt(t, pt(spelling.MustNew(letter.C, ai(mag)), 4), 0, nil), i: iv("A1"),
			want: interval.ErrOutOfRange, causes: []error{alteration.ErrOutOfRange}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := interval.TransposeNoteIn(tt.sys, tt.x, tt.i)
			assertErr(t, err, tt.want, tt.causes...)
			assert.Equal(t, note.Note{}, got)

			for _, e := range tt.notWant {
				assert.NotErrorIs(t, err, e)
			}
		})
	}
}

// TestNotesSweep checks F-7 and P-15 on a sweep: the interval from a note to
// its transposition is the interval of the transposition, the correction
// and the source are kept, the input note does not change.
func TestNotesSweep(t *testing.T) {
	t.Parallel()

	src := std()
	intervals := predefinedBoth()

	for _, name := range []spelling.Spelling{spelling.C(), spelling.FSharp(), spelling.BFlat(), spelling.BSharp()} {
		x := noteAt(t, wp(trad(), name, 4), 25, src)
		xName := noteName(t, name, -25, src)

		for _, i := range intervals {
			before := x

			y, err := interval.TransposeNote(x, i)
			require.NoError(t, err, "%v by %v", x, i)
			unchanged(t, before, x)
			assert.True(t, y.SameCents(x) && y.SameSource(x), "%v by %v", x, i)

			got, err := interval.BetweenNotes(x, y, interval.DirectionUnset)
			require.NoError(t, err)
			assert.True(t, got.Equal(i), "%v by %v: %v", x, i, got)

			p, err := y.Pitch()
			require.NoError(t, err)

			xp, err := x.Pitch()
			require.NoError(t, err)
			assert.True(t, p.Equal(interval.MustTransposePitch(trad(), xp, i)))

			yName, err := interval.TransposeNote(xName, i)
			require.NoError(t, err)
			assert.False(t, yName.HasOctave())
			assert.True(t, yName.Spelling().Equal(interval.MustTranspose(name, i)))
			assert.True(t, yName.SameCents(xName) && yName.SameSource(xName))
		}
	}
}
