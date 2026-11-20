package voice_test

import (
	"testing"

	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/instrument"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/voice"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// П-14, КП-36, architecture 4.2: constructors copy their inputs.
func TestConstructorsCopyInputs(t *testing.T) {
	t.Parallel()

	t.Run("New", func(t *testing.T) {
		t.Parallel()
		elements := []voice.Element{snd(c4(), q()), rst(e8())}
		l := voice.MustNew(elements...)
		elements[0] = rst(h())
		assertLine(t, line(snd(c4(), q()), rst(e8())), l)
	})

	t.Run("NewSound", func(t *testing.T) {
		t.Parallel()
		n := c4()
		e := snd(n, q())
		require.NoError(t, n.SetASPN(6))
		got, _ := e.MustNote()
		assertNote(t, c4(), got)
	})

	t.Run("FromPairs", func(t *testing.T) {
		t.Parallel()
		n := c4()
		pairs := []voice.Pair{{Note: &n, Duration: q()}, {Duration: e8()}}
		l := voice.MustFromPairs(pairs...)
		require.NoError(t, n.SetASPN(5))
		pairs[1] = voice.Pair{Note: &n, Duration: h()}
		assertLine(t, line(snd(c4(), q()), rst(e8())), l)
	})

	t.Run("FromEvents", func(t *testing.T) {
		t.Parallel()
		events := []*event.Event{noteEvent(c4(), duration.Quarter, event.Marks{})}
		l := voice.MustFromEvents(events...)
		events[0] = restEvent(duration.Half, event.Marks{})
		assertLine(t, line(snd(c4(), q())), l)
	})

	t.Run("Join", func(t *testing.T) {
		t.Parallel()
		lines := []*voice.Line{line(snd(c4(), q())), line(rst(q()))}
		l := voice.MustJoin(lines...)
		lines[0] = line()
		assertLine(t, line(snd(c4(), q()), rst(q())), l)
	})

	t.Run("Append and Insert", func(t *testing.T) {
		t.Parallel()
		elements := []voice.Element{snd(d4(), q())}
		appended := line(snd(c4(), q())).MustAppend(elements...)
		inserted := line(snd(c4(), q())).MustInsert(0, elements...)
		elements[0] = rst(h())
		assertLine(t, line(snd(c4(), q()), snd(d4(), q())), appended)
		assertLine(t, line(snd(d4(), q()), snd(c4(), q())), inserted)
	})

	t.Run("WithNote", func(t *testing.T) {
		t.Parallel()
		n := d4()
		l := line(snd(c4(), q())).MustWithNote(0, n)
		require.NoError(t, n.SetASPN(2))
		assertLine(t, line(snd(d4(), q())), l)
	})
}

// П-14, КП-36, НФ-3: reading returns copies; changing them does not change
// the line, and each call returns a new slice.
func TestReadingReturnsCopies(t *testing.T) {
	t.Parallel()
	l := line(snd(c4(), q()), rst(q()).MustWithFermata(true), snd(e4(), h()))
	snapshot := line(snd(c4(), q()), rst(q()).MustWithFermata(true), snd(e4(), h()))

	elements := l.MustElements()
	elements[0] = rst(w())
	n, _ := l.MustElements()[0].MustNote()
	assertNote(t, c4(), n)

	got, ok := l.MustElement(0).MustNote()
	require.True(t, ok)
	require.NoError(t, got.SetASPN(7))

	notes := l.MustNotes()
	require.NoError(t, notes[0].SetASPN(7))
	notes[1] = c4()

	sounding, ok := l.MustSoundAt(duration.Origin())
	require.True(t, ok)
	require.NoError(t, sounding.SetASPN(1))

	sounds := l.MustSounds()
	require.NoError(t, sounds[0].Note.SetASPN(3))
	sounds[1].Index = 9

	starts, _ := l.MustStarts()
	starts[1] = pos(9, 1)

	span := l.MustSpan(duration.Origin(), pos(1, 1))
	span[0] = 7

	fermatas := l.MustFermatas()
	fermatas[0].Index = 5

	melodic := l.MustIntervals()
	melodic[0].From = 4

	locations := l.MustLocate(voice.PresetReference(soprano()))
	locations[0].Index = 8

	_, outside := line(snd(nt(spelling.C(), 2), q())).MustWithin(voice.PresetReference(soprano()))
	outside[0] = 3

	events := l.MustEvents()
	events[0] = restEvent(duration.Whole, event.Marks{})

	timings := l.MustSoundSeconds(map60(), 0, rounding.Down)
	timings[0].Index = 6

	assertLine(t, snapshot, l)
	startsAgain, _ := l.MustStarts()
	assert.Equal(t, []string{"0/1", "1/4", "1/2"}, positionStrings(startsAgain))
	assert.Equal(t, []int{0, 1, 2}, l.MustSpan(duration.Origin(), pos(1, 1)))
	assert.Equal(t, 1, l.MustFermatas()[0].Index)
	assert.Equal(t, 0, l.MustIntervals()[0].From)
	assert.Equal(t, 0, l.MustLocate(voice.PresetReference(soprano()))[0].Index)
	assert.Equal(t, 0, l.MustSounds()[0].Index)
	assert.Equal(t, 2, l.MustSounds()[1].Index)
	assertNote(t, c4(), l.MustSounds()[0].Note)
	assert.Equal(t, 0, l.MustSoundSeconds(map60(), 0, rounding.Down)[0].Index)
	assert.True(t, noteEvent(c4(), duration.Quarter, event.Marks{}).Equal(l.MustEvents()[0]))
	_, outsideAgain := line(snd(nt(spelling.C(), 2), q())).MustWithin(voice.PresetReference(soprano()))
	assert.Equal(t, []int{0}, outsideAgain)
}

// П-14, architecture 4.3: changes and building return new lines; the receiver
// does not change on success or failure.
func TestReceiverIsUnchanged(t *testing.T) {
	t.Parallel()
	src := func() *voice.Line {
		return named(line(snd(c4(), q()), rst(e8()).MustWithFermata(true), snd(d4(), qe())))
	}
	ops := map[string]func(l *voice.Line) (*voice.Line, error){
		"Append":        func(l *voice.Line) (*voice.Line, error) { return l.Append(rst(q())) },
		"Insert":        func(l *voice.Line) (*voice.Line, error) { return l.Insert(0, rst(q())) },
		"Replace":       func(l *voice.Line) (*voice.Line, error) { return l.Replace(0, rst(q())) },
		"Remove":        func(l *voice.Line) (*voice.Line, error) { return l.Remove(0) },
		"RemoveRange":   func(l *voice.Line) (*voice.Line, error) { return l.RemoveRange(0, 2) },
		"Slice":         func(l *voice.Line) (*voice.Line, error) { return l.Slice(0, 3) },
		"WithName":      func(l *voice.Line) (*voice.Line, error) { return l.WithName("Melody") },
		"WithPlacement": func(l *voice.Line) (*voice.Line, error) { return l.WithPlacement(voice.Bass) },
		"WithThematic":  func(l *voice.Line) (*voice.Line, error) { return l.WithThematic(voice.Pedal) },
		"WithImportance": func(l *voice.Line) (*voice.Line, error) {
			return l.WithImportance(voice.Filler)
		},
		"WithNote":       func(l *voice.Line) (*voice.Line, error) { return l.WithNote(1, e4()) },
		"WithDuration":   func(l *voice.Line) (*voice.Line, error) { return l.WithDuration(0, h()) },
		"WithFermata":    func(l *voice.Line) (*voice.Line, error) { return l.WithFermata(0, true) },
		"Transpose":      func(l *voice.Line) (*voice.Line, error) { return l.Transpose(interval.MajorThird()) },
		"TransposeIn":    func(l *voice.Line) (*voice.Line, error) { return l.TransposeIn(traditional(), interval.MajorThird()) },
		"Double":         func(l *voice.Line) (*voice.Line, error) { return l.Double(interval.PerfectOctave()) },
		"Imitate":        func(l *voice.Line) (*voice.Line, error) { return l.Imitate(interval.PerfectFifth()) },
		"ImitateAfter":   func(l *voice.Line) (*voice.Line, error) { return l.ImitateAfter(interval.PerfectFifth(), h()) },
		"Invert":         func(l *voice.Line) (*voice.Line, error) { return l.Invert(e4()) },
		"InvertPitch":    func(l *voice.Line) (*voice.Line, error) { return l.InvertPitch(pt(spelling.E(), 4)) },
		"RetrogradeBoth": func(l *voice.Line) (*voice.Line, error) { return l.Retrograde(voice.RetrogradeBoth) },
		"RetrogradePitches": func(l *voice.Line) (*voice.Line, error) {
			return l.Retrograde(voice.RetrogradePitches)
		},
		"RetrogradeRhythm": func(l *voice.Line) (*voice.Line, error) { return l.Retrograde(voice.RetrogradeRhythm) },
		"Split":            func(l *voice.Line) (*voice.Line, error) { a, _, err := l.Split(1); return a, err },
		"failing Insert":   func(l *voice.Line) (*voice.Line, error) { return l.Insert(1, voice.Element{}) },
		"failing Invert":   func(l *voice.Line) (*voice.Line, error) { return l.Invert(noOctave()) },
		"failing WithNote": func(l *voice.Line) (*voice.Line, error) { return l.WithNote(9, e4()) },
	}
	for name, op := range ops {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			l := src()
			got, err := op(l)
			assertLine(t, src(), l)
			if err == nil {
				assert.NotSame(t, l, got, "a new line is returned")
			}
		})
	}
}

// The operations do not change the instruments and presets of a reference.
func TestReferenceInputsAreUnchanged(t *testing.T) {
	t.Parallel()
	preset := soprano()
	inst := instrument.MustNew(soprano())
	l := line(snd(c4(), q()), snd(nt(spelling.A(), 5), q()))
	_ = l.MustTessitura(voice.PresetReference(preset))
	_ = l.MustLocate(voice.InstrumentReference(inst))
	assert.True(t, soprano().Equal(preset))
	assert.True(t, instrument.MustNew(soprano()).Equal(inst))
}

// Notes kept by a line keep the reference to their source (architecture 4.2):
// a copy read later has the same source.
func TestNotesKeepSource(t *testing.T) {
	t.Parallel()
	n := c4()
	l := line(snd(n, q()))
	got := l.MustNotes()
	assert.True(t, n.SameSource(got[0]))
}
