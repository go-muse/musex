package voice_test

import (
	"testing"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/voice"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ОП-1…ОП-3, architecture 6.4.
func TestElementReading(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		e       voice.Element
		kind    voice.Kind
		note    note.Note
		hasNote bool
		d       voice.Duration
		length  string
		fermata bool
	}{
		{name: "sound", e: snd(c4(), q()), kind: voice.KindSound, note: c4(), hasNote: true, d: q(), length: "1/4"},
		{name: "tied sound", e: snd(d4(), qe()), kind: voice.KindSound, note: d4(), hasNote: true, d: qe(), length: "3/8"},
		{name: "sound without notation", e: snd(e4(), val(1, 3)), kind: voice.KindSound, note: e4(), hasNote: true, d: val(1, 3), length: "1/3"},
		{name: "sound with fermata", e: snd(c4(), h()).MustWithFermata(true), kind: voice.KindSound, note: c4(), hasNote: true, d: h(), length: "1/2", fermata: true},
		{name: "rest", e: rst(e8()), kind: voice.KindRest, d: e8(), length: "1/8"},
		{name: "rest with fermata", e: rst(val(2, 1)).MustWithFermata(true), kind: voice.KindRest, d: val(2, 1), length: "2/1", fermata: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := tt.e
			assert.True(t, e.IsSet())
			assert.Equal(t, tt.kind, e.Kind())
			assert.Equal(t, tt.kind == voice.KindSound, e.IsSound())
			assert.Equal(t, tt.kind == voice.KindRest, e.IsRest())

			n, ok, err := e.Note()
			require.NoError(t, err)
			assert.Equal(t, tt.hasNote, ok)
			if tt.hasNote {
				assertNote(t, tt.note, n)
			} else {
				assert.Zero(t, n, "a rest has no note")
			}
			mn, mok := e.MustNote()
			assert.Equal(t, ok, mok)
			assert.True(t, n.Equal(mn) || (!ok && !mn.IsValid()))

			d, err := e.Duration()
			require.NoError(t, err)
			assert.True(t, tt.d.Equal(d))
			assert.True(t, tt.d.Equal(e.MustDuration()))

			length, err := e.Length()
			require.NoError(t, err)
			assert.Equal(t, tt.length, length.String())
			assert.Equal(t, tt.length, e.MustLength().String())

			fermata, err := e.Fermata()
			require.NoError(t, err)
			assert.Equal(t, tt.fermata, fermata)
			assert.Equal(t, tt.fermata, e.MustFermata())

			assert.NotEmpty(t, e.String())
			assert.NotEqual(t, "Element(unset)", e.String())
		})
	}
}

// П-2: the correction and the source of the note are kept as given.
func TestSoundKeepsNote(t *testing.T) {
	t.Parallel()
	n := withCents(spelling.C(), 4, 15)
	e := snd(n, q())
	got, ok := e.MustNote()
	require.True(t, ok)
	assert.True(t, n.SameRecord(got), "want %v, got %v", n, got)
	assert.True(t, n.SameSource(got))

	unbound := free(spelling.C(), 4)
	got, ok = snd(unbound, q()).MustNote()
	require.True(t, ok)
	assert.False(t, got.IsBound())
}

// New elements have no fermata (architecture 6.4).
func TestNewElementsHaveNoFermata(t *testing.T) {
	t.Parallel()
	assert.False(t, snd(c4(), q()).MustFermata())
	assert.False(t, rst(q()).MustFermata())
}

func TestElementConstructorFailures(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		n    note.Note
		d    voice.Duration
		want want
	}{
		{"unset note first", note.Note{}, voice.Duration{}, fails(voice.ErrUnset, note.ErrUnset)},
		{"unset duration", c4(), voice.Duration{}, fails(voice.ErrUnset).only()},
		{"unset duration before no octave", noOctave(), voice.Duration{}, fails(voice.ErrUnset).only()},
		{"no octave", noOctave(), q(), fails(voice.ErrNoOctave).only()},
	}
	for _, tt := range tests {
		t.Run("sound: "+tt.name, func(t *testing.T) {
			t.Parallel()
			assertFailure(t, tt.want,
				func() (voice.Element, error) { return voice.NewSound(tt.n, tt.d) },
				func() voice.Element { return voice.MustNewSound(tt.n, tt.d) })
		})
	}

	t.Run("rest: unset duration", func(t *testing.T) {
		t.Parallel()
		assertFailure(t, fails(voice.ErrUnset).only(),
			func() (voice.Element, error) { return voice.NewRest(voice.Duration{}) },
			func() voice.Element { return voice.MustNewRest(voice.Duration{}) })
	})

	// КП-31: a rest with a tied duration.
	t.Run("rest: tied duration", func(t *testing.T) {
		t.Parallel()
		assertFailure(t, fails(voice.ErrTiedRest).only(),
			func() (voice.Element, error) { return voice.NewRest(qe()) },
			func() voice.Element { return voice.MustNewRest(qe()) })
	})
}

// Architecture 6.1: the zero Element is unset.
func TestUnsetElement(t *testing.T) {
	t.Parallel()
	var e voice.Element
	assert.False(t, e.IsSet())
	assert.Equal(t, voice.KindUnset, e.Kind())
	assert.False(t, e.IsSound())
	assert.False(t, e.IsRest())
	assert.Equal(t, "Element(unset)", e.String())
	assert.True(t, e.Equal(voice.Element{}))
	assert.True(t, e.EqualDuration(voice.Element{}))
	assert.False(t, e.Equal(rst(q())))
	assert.False(t, rst(q()).Equal(e))
	assert.False(t, e.EqualDuration(rst(q())))

	w := fails(voice.ErrUnset).only()
	assertFailure(t, w, func() (pair[note.Note, bool], error) { return two(e.Note()) },
		func() pair[note.Note, bool] { return mustTwo(e.MustNote()) })
	assertFailure(t, w, e.Duration, e.MustDuration)
	assertFailure(t, w, e.Length, e.MustLength)
	assertFailure(t, w, e.Fermata, e.MustFermata)
	assertFailure(t, w,
		func() (voice.Element, error) { return e.WithFermata(true) },
		func() voice.Element { return e.MustWithFermata(true) })
	// The element is checked before the argument.
	assertFailure(t, w,
		func() (voice.Element, error) { return e.WithNote(note.Note{}) },
		func() voice.Element { return e.MustWithNote(note.Note{}) })
	assertFailure(t, w,
		func() (voice.Element, error) { return e.WithDuration(voice.Duration{}) },
		func() voice.Element { return e.MustWithDuration(voice.Duration{}) })
}

func TestElementWith(t *testing.T) {
	t.Parallel()

	t.Run("WithNote keeps duration and fermata", func(t *testing.T) {
		t.Parallel()
		e := snd(c4(), qe()).MustWithFermata(true)
		got, err := e.WithNote(d4())
		require.NoError(t, err)
		assertElement(t, snd(d4(), qe()).MustWithFermata(true), got)
		assertElement(t, snd(c4(), qe()).MustWithFermata(true), e)
		assertElement(t, got, e.MustWithNote(d4()))
	})

	t.Run("WithNote makes a rest a sound", func(t *testing.T) {
		t.Parallel()
		e := rst(h()).MustWithFermata(true)
		got := e.MustWithNote(e4())
		assert.True(t, got.IsSound())
		assertElement(t, snd(e4(), h()).MustWithFermata(true), got)
		assert.True(t, e.IsRest())
	})

	t.Run("WithDuration", func(t *testing.T) {
		t.Parallel()
		e := snd(c4(), q()).MustWithFermata(true)
		got, err := e.WithDuration(qe())
		require.NoError(t, err)
		assertElement(t, snd(c4(), qe()).MustWithFermata(true), got)
		assertElement(t, rst(val(1, 5)), rst(q()).MustWithDuration(val(1, 5)))
		assertElement(t, snd(c4(), q()).MustWithFermata(true), e)
	})

	t.Run("WithFermata", func(t *testing.T) {
		t.Parallel()
		e := snd(c4(), q())
		on, err := e.WithFermata(true)
		require.NoError(t, err)
		assert.True(t, on.MustFermata())
		assert.False(t, e.MustFermata())
		assert.False(t, on.MustWithFermata(false).MustFermata())
	})

	failures := []struct {
		name string
		call func() (voice.Element, error)
		must func() voice.Element
		want want
	}{
		{"WithNote unset note", func() (voice.Element, error) { return rst(q()).WithNote(note.Note{}) },
			func() voice.Element { return rst(q()).MustWithNote(note.Note{}) }, fails(voice.ErrUnset, note.ErrUnset)},
		{"WithNote without octave", func() (voice.Element, error) { return snd(c4(), q()).WithNote(noOctave()) },
			func() voice.Element { return snd(c4(), q()).MustWithNote(noOctave()) }, fails(voice.ErrNoOctave).only()},
		{"WithDuration unset", func() (voice.Element, error) { return snd(c4(), q()).WithDuration(voice.Duration{}) },
			func() voice.Element { return snd(c4(), q()).MustWithDuration(voice.Duration{}) }, fails(voice.ErrUnset).only()},
		{"WithDuration tied on a rest", func() (voice.Element, error) { return rst(q()).WithDuration(qe()) },
			func() voice.Element { return rst(q()).MustWithDuration(qe()) }, fails(voice.ErrTiedRest).only()},
	}
	for _, tt := range failures {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertFailure(t, tt.want, tt.call, tt.must)
		})
	}
}

// Architecture 6.4, П-16: Equal compares the kind, the note by record and
// correction, the duration by Duration.Equal and the fermata; EqualDuration
// compares lengths instead of records.
func TestElementEquality(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		a, b          voice.Element
		equal, sameBy bool
	}{
		{"same sound", snd(c4(), q()), snd(c4(), q()), true, true},
		{"source is ignored", snd(c4(), q()), snd(free(spelling.C(), 4), q()), true, true},
		{"other note", snd(c4(), q()), snd(d4(), q()), false, false},
		{"other correction", snd(c4(), q()), snd(withCents(spelling.C(), 4, 10), q()), false, false},
		{"enharmonic note", snd(c4(), q()), snd(written(spelling.BSharp(), 3), q()), false, false},
		{"dotted and tied", snd(c4(), dq()), snd(c4(), qe()), false, true},
		{"written and value", snd(c4(), q()), snd(c4(), val(1, 4)), false, true},
		{"other length", snd(c4(), q()), snd(c4(), e8()), false, false},
		{"fermata", snd(c4(), q()), snd(c4(), q()).MustWithFermata(true), false, false},
		{"sound and rest", snd(c4(), q()), rst(q()), false, false},
		{"same rest", rst(q()), rst(q()), true, true},
		{"rests of one length", rst(q()), rst(val(1, 4)), false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.equal, tt.a.Equal(tt.b))
			assert.Equal(t, tt.equal, tt.b.Equal(tt.a))
			assert.Equal(t, tt.sameBy, tt.a.EqualDuration(tt.b))
			assert.Equal(t, tt.sameBy, tt.b.EqualDuration(tt.a))
		})
	}
}

// Architecture 6.14: the strings of Kind.
func TestKind(t *testing.T) {
	t.Parallel()
	tests := []struct {
		kind  voice.Kind
		valid bool
		text  string
	}{
		{voice.KindUnset, false, "Kind(unset)"},
		{voice.KindSound, true, "sound"},
		{voice.KindRest, true, "rest"},
		{voice.Kind(42), false, "Kind(invalid)"},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.valid, tt.kind.IsValid())
			assert.Equal(t, tt.text, tt.kind.String())
		})
	}
}

// Element length is the length of its duration: written, tied or value.
func TestElementLengthOfDurations(t *testing.T) {
	t.Parallel()
	got := snd(c4(), voice.MustWrittenDuration(wr(duration.Whole, 2))).MustLength()
	assert.Equal(t, "7/4", got.String())
}
