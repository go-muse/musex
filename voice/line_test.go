package voice_test

import (
	"testing"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/voice"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ОП-4, ОП-6, П-1: a new line has the elements in order, no name and no
// roles.
func TestNewLine(t *testing.T) {
	t.Parallel()
	elements := []voice.Element{snd(c4(), q()), rst(e8()), snd(d4(), qe()).MustWithFermata(true), rst(val(1, 3))}
	l, err := voice.New(elements...)
	require.NoError(t, err)
	assert.True(t, l.IsSet())
	assert.Equal(t, 4, l.MustLen())
	assert.False(t, l.MustIsEmpty())
	assert.Equal(t, 2, l.MustSoundCount())
	assert.Equal(t, 2, l.MustRestCount())
	got := l.MustElements()
	require.Len(t, got, len(elements))
	for i := range elements {
		assertElement(t, elements[i], got[i])
		assertElement(t, elements[i], l.MustElement(i))
	}
	assertNotes(t, []note.Note{c4(), d4()}, l.MustNotes())
	assert.Empty(t, l.MustName())
	assert.Equal(t, voice.PlacementUnset, l.MustPlacement())
	assert.Equal(t, voice.ThematicUnset, l.MustThematic())
	assert.Equal(t, voice.ImportanceUnset, l.MustImportance())
	assert.NotEmpty(t, l.String())
	assert.NotEqual(t, "Line(unset)", l.String())
	assert.Equal(t, l.String(), l.String())
}

// С-13: the empty line, a line of one rest and a line of one sound.
func TestBoundaryLines(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                   string
		l                      *voice.Line
		n, sounds, rests       int
		empty                  bool
		notes                  []note.Note
		length                 string
		elementsAreEmptyNonNil bool
	}{
		{name: "empty", l: line(), empty: true, notes: []note.Note{}, length: "0/1"},
		{name: "one rest", l: line(rst(q())), n: 1, rests: 1, notes: []note.Note{}, length: "1/4"},
		{name: "one sound", l: line(snd(c4(), q())), n: 1, sounds: 1, notes: []note.Note{c4()}, length: "1/4"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n, err := tt.l.Len()
			require.NoError(t, err)
			assert.Equal(t, tt.n, n)
			empty, err := tt.l.IsEmpty()
			require.NoError(t, err)
			assert.Equal(t, tt.empty, empty)
			sounds, err := tt.l.SoundCount()
			require.NoError(t, err)
			assert.Equal(t, tt.sounds, sounds)
			rests, err := tt.l.RestCount()
			require.NoError(t, err)
			assert.Equal(t, tt.rests, rests)

			elements, err := tt.l.Elements()
			require.NoError(t, err)
			require.NotNil(t, elements, "a successful empty list is not nil")
			assert.Len(t, elements, tt.n)

			notes, err := tt.l.Notes()
			require.NoError(t, err)
			require.NotNil(t, notes, "a successful empty list is not nil")
			assertNotes(t, tt.notes, notes)

			length, err := tt.l.Length()
			require.NoError(t, err)
			assert.Equal(t, tt.length, length.String())
		})
	}

	empty := voice.MustNew()
	assert.True(t, empty.IsSet())
	assert.Equal(t, "0/1", empty.MustLength().String())
	assert.True(t, duration.Origin().Equal(empty.MustEnd()))
}

// A tied sound is one element and one note (П-3).
func TestTiedSoundIsOneElement(t *testing.T) {
	t.Parallel()
	l := kp4()
	assert.Equal(t, 4, l.MustLen())
	assert.Equal(t, 3, l.MustSoundCount())
	assertNotes(t, []note.Note{c4(), d4(), e4()}, l.MustNotes())
	tied, ok := l.MustElement(0).MustDuration().MustTied()
	require.True(t, ok)
	assert.True(t, tiedQE().Equal(tied))
}

// ОП-5: pairs of a note or a rest and a duration, without fermatas.
func TestFromPairs(t *testing.T) {
	t.Parallel()
	c, d := c4(), d4()
	l, err := voice.FromPairs(
		voice.Pair{Note: &c, Duration: q()},
		voice.Pair{Duration: e8()},
		voice.Pair{Note: &d, Duration: qe()},
	)
	require.NoError(t, err)
	assertLine(t, line(snd(c4(), q()), rst(e8()), snd(d4(), qe())), l)
	assertLine(t, l, voice.MustFromPairs(voice.Pair{Note: &c, Duration: q()}, voice.Pair{Duration: e8()}, voice.Pair{Note: &d, Duration: qe()}))
	for _, e := range l.MustElements() {
		assert.False(t, e.MustFermata())
	}

	empty, err := voice.FromPairs()
	require.NoError(t, err)
	assertLine(t, voice.MustNew(), empty)
}

func TestFromPairsFailures(t *testing.T) {
	t.Parallel()
	c := c4()
	unset := note.Note{}
	octaveless := noOctave()
	tests := []struct {
		name  string
		pairs []voice.Pair
		want  want
	}{
		{"unset note", []voice.Pair{{Note: &c, Duration: q()}, {Note: &unset, Duration: q()}},
			failsAt(1, voice.ErrUnset, note.ErrUnset)},
		{"unset sound duration", []voice.Pair{{Note: &c, Duration: voice.Duration{}}}, failsAt(0, voice.ErrUnset).only()},
		{"unset rest duration", []voice.Pair{{Note: &c, Duration: q()}, {Duration: q()}, {}}, failsAt(2, voice.ErrUnset).only()},
		{"note without octave", []voice.Pair{{Note: &octaveless, Duration: q()}}, failsAt(0, voice.ErrNoOctave).only()},
		{"tied rest", []voice.Pair{{Duration: q()}, {Duration: qe()}}, failsAt(1, voice.ErrTiedRest).only()},
		{"first failing pair wins", []voice.Pair{{Duration: qe()}, {Note: &unset, Duration: q()}}, failsAt(0, voice.ErrTiedRest).only()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertFailure(t, tt.want,
				func() (*voice.Line, error) { return voice.FromPairs(tt.pairs...) },
				func() *voice.Line { return voice.MustFromPairs(tt.pairs...) })
		})
	}
}

func TestNewFailures(t *testing.T) {
	t.Parallel()
	assertFailure(t, failsAt(2, voice.ErrUnset).only(),
		func() (*voice.Line, error) {
			return voice.New(snd(c4(), q()), rst(q()), voice.Element{}, voice.Element{})
		},
		func() *voice.Line { return voice.MustNew(snd(c4(), q()), rst(q()), voice.Element{}, voice.Element{}) })
}

// ОП-38: Join keeps the name and the roles of the first line.
func TestJoin(t *testing.T) {
	t.Parallel()
	first := named(line(snd(c4(), q())))
	second := line(rst(e8()), snd(d4(), e8())).MustWithName("Second").MustWithPlacement(voice.Bass)
	third := line()
	got, err := voice.Join(first, second, third)
	require.NoError(t, err)
	want := named(line(snd(c4(), q()), rst(e8()), snd(d4(), e8())))
	assertLine(t, want, got)
	assertLine(t, want, voice.MustJoin(first, second, third))
	assert.Equal(t, "1/2", got.MustLength().String())

	empty, err := voice.Join()
	require.NoError(t, err)
	assertLine(t, voice.MustNew(), empty)
	assert.Empty(t, empty.MustName())
	assert.Equal(t, voice.PlacementUnset, empty.MustPlacement())

	// The empty first line still gives its name and roles.
	assertLine(t, line(snd(c4(), q())).MustWithName("Empty"),
		voice.MustJoin(line().MustWithName("Empty"), line(snd(c4(), q())).MustWithName("Other")))

	assertFailure(t, failsAt(1, voice.ErrUnset).only(),
		func() (*voice.Line, error) { return voice.Join(first, nil, &voice.Line{}) },
		func() *voice.Line { return voice.MustJoin(first, nil, &voice.Line{}) })
	assertFailure(t, failsAt(0, voice.ErrUnset).only(),
		func() (*voice.Line, error) { return voice.Join(&voice.Line{}) },
		func() *voice.Line { return voice.MustJoin(&voice.Line{}) })
}

// КП-37, П-1: the zero value is an unset line.
func TestUnsetLine(t *testing.T) {
	t.Parallel()
	for name, l := range map[string]*voice.Line{"nil": nil, "zero": {}} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.False(t, l.IsSet())
			assert.Equal(t, "Line(unset)", l.String())
			assert.True(t, l.Equal(nil))
			assert.True(t, l.Equal(&voice.Line{}))
			assert.True(t, l.EqualElements(nil))
			assert.True(t, l.EqualDuration(&voice.Line{}))
			assert.False(t, l.Equal(voice.MustNew()), "the unset line is not the empty line")
			assert.False(t, voice.MustNew().Equal(l))
			assert.False(t, l.EqualElements(voice.MustNew()))
			assert.False(t, l.EqualDuration(voice.MustNew()))

			w := fails(voice.ErrUnset).only()
			assertFailure(t, w, l.Len, l.MustLen)
			assertFailure(t, w, l.IsEmpty, l.MustIsEmpty)
			assertFailure(t, w, l.SoundCount, l.MustSoundCount)
			assertFailure(t, w, l.RestCount, l.MustRestCount)
			assertFailure(t, w, func() (voice.Element, error) { return l.Element(0) },
				func() voice.Element { return l.MustElement(0) })
			assertFailure(t, w, l.Elements, l.MustElements)
			assertFailure(t, w, l.Notes, l.MustNotes)
			assertFailure(t, w, l.Name, l.MustName)
			assertFailure(t, w, l.Placement, l.MustPlacement)
			assertFailure(t, w, l.Thematic, l.MustThematic)
			assertFailure(t, w, l.Importance, l.MustImportance)
			assertFailure(t, w, l.Length, l.MustLength)
			assertFailure(t, w, l.End, l.MustEnd)
			assertFailure(t, w, l.Events, l.MustEvents)
			assertFailure(t, w, l.Sounds, l.MustSounds)
			assertFailure(t, w, l.Fermatas, l.MustFermatas)
		})
	}
}

// Element reads a copy by index; out of range is Е-2 (КП-35).
func TestElementByIndex(t *testing.T) {
	t.Parallel()
	l := kp1()
	got, err := l.Element(3)
	require.NoError(t, err)
	assertElement(t, snd(f4(), h()), got)

	for _, i := range []int{-1, 4, 100} {
		assertFailure(t, fails(voice.ErrIndex).only(),
			func() (voice.Element, error) { return l.Element(i) },
			func() voice.Element { return l.MustElement(i) })
	}
	// КП-35: number 3 in a line of three elements.
	three := line(snd(c4(), q()), rst(q()), snd(d4(), q()))
	assertFailure(t, fails(voice.ErrIndex).only(),
		func() (voice.Element, error) { return three.Element(3) },
		func() voice.Element { return three.MustElement(3) })
	assertFailure(t, fails(voice.ErrIndex).only(),
		func() (voice.Element, error) { return line().Element(0) },
		func() voice.Element { return line().MustElement(0) })
}

// Names and roles are read as set (ОП-6, ОП-39).
func TestNameAndRoles(t *testing.T) {
	t.Parallel()
	l := named(kp1())
	assert.Equal(t, "Melody", l.MustName())
	assert.Equal(t, voice.Soprano, l.MustPlacement())
	assert.Equal(t, voice.Leading, l.MustThematic())
	assert.Equal(t, voice.Principal, l.MustImportance())

	name, err := l.Name()
	require.NoError(t, err)
	assert.Equal(t, "Melody", name)
	p, err := l.Placement()
	require.NoError(t, err)
	assert.Equal(t, voice.Soprano, p)
	r, err := l.Thematic()
	require.NoError(t, err)
	assert.Equal(t, voice.Leading, r)
	i, err := l.Importance()
	require.NoError(t, err)
	assert.Equal(t, voice.Principal, i)
}

// The three roles are independent (П-12): any combination is kept.
func TestRolesAreIndependent(t *testing.T) {
	t.Parallel()
	l := kp1().MustWithPlacement(voice.Bass).MustWithThematic(voice.Leading).MustWithImportance(voice.Filler)
	assert.Equal(t, voice.Bass, l.MustPlacement())
	assert.Equal(t, voice.Leading, l.MustThematic())
	assert.Equal(t, voice.Filler, l.MustImportance())
}

// The name does not identify a line and any string is allowed.
func TestAnyName(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"", " ", "Мелодия", "Soprano Voice", "a\nb"} {
		assert.Equal(t, name, kp1().MustWithName(name).MustName())
	}
}

// Events of a line are not required to create it: a line is independent of
// event, but a line built from events equals one built from elements.
func TestLineFromEventsEqualsLineFromElements(t *testing.T) {
	t.Parallel()
	got := voice.MustFromEvents(noteEvent(c4(), duration.Quarter, event.Marks{}), restEvent(duration.Eighth, event.Marks{}))
	assertLine(t, line(snd(c4(), q()), rst(e8())), got)
}
