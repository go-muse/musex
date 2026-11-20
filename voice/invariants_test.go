package voice_test

import (
	"testing"

	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/voice"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// invariantLines is the set of lines for the properties ИЛ-1…ИЛ-7: empty,
// of one rest, of one sound, with tied sounds, rests and fermatas.
func invariantLines() map[string]*voice.Line {
	return map[string]*voice.Line{
		"empty":     line(),
		"one rest":  line(rst(q())),
		"one sound": line(snd(c4(), q())),
		"КП-1":      kp1(),
		"КП-4":      kp4(),
		"tied sounds and fermatas": line(rst(e8()).MustWithFermata(true), snd(d4(), qe()), snd(g4(), dq()).MustWithFermata(true),
			snd(nt(spelling.A(), 3), voice.MustTiedDuration(tiedQE())), rst(h())),
		"values":           line(snd(c4(), val(1, 3)), rst(val(1, 6)), snd(nt(spelling.BFlat(), 3), val(5, 7)), snd(e4(), w())),
		"repeated pitches": line(snd(e4(), q()), snd(e4(), q()), rst(q()), snd(e4(), h()).MustWithFermata(true)),
	}
}

// ИЛ-1: every element is one sound with one note or a rest; elements do not
// overlap: at every start exactly that element sounds.
func TestOnePitchAtATime(t *testing.T) {
	t.Parallel()
	for name, l := range invariantLines() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			starts, end := l.MustStarts()
			for i, start := range starts {
				e := l.MustElement(i)
				assert.NotEqual(t, e.IsSound(), e.IsRest(), "element %d is a sound or a rest", i)
				at, ok := l.MustAt(start)
				require.True(t, ok)
				assert.Equal(t, i, at, "the element at its own start")
				spanned := l.MustSpan(start, l.MustEndOf(i))
				assert.Equal(t, []int{i}, spanned, "elements do not overlap")
				n, sounding := l.MustSoundAt(start)
				assert.Equal(t, e.IsSound(), sounding)
				if sounding {
					want, _ := e.MustNote()
					assertNote(t, want, n)
				}
			}
			_, ok := l.MustAt(end)
			assert.False(t, ok)
		})
	}
}

// ИЛ-4: the number of melodic intervals is the number of sounds minus one,
// zero without sounds.
func TestNumberOfIntervals(t *testing.T) {
	t.Parallel()
	for name, l := range invariantLines() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			want := max(l.MustSoundCount()-1, 0)
			got := l.MustIntervals()
			assert.Len(t, got, want)
			sounds := l.MustSounds()
			for k, m := range got {
				assert.Equal(t, sounds[k].Index, m.From)
				assert.Equal(t, sounds[k+1].Index, m.To)
				assert.Equal(t, m.To-m.From > 1, m.ThroughRest, "a rest lies between the sounds")
			}
		})
	}
}

// ИЛ-6: the inversion twice around the same axis gives the original
// written pitches; the same retrograde twice gives the original line.
func TestInvolutions(t *testing.T) {
	t.Parallel()
	for name, l := range invariantLines() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			src := named(l)
			for _, axis := range []note.Note{c4(), nt(spelling.FSharp(), 4), nt(spelling.BFlat(), 3)} {
				twice := src.MustInvert(axis).MustInvert(axis)
				assert.True(t, src.EqualElements(twice), "axis %v: want %v, got %v", axis, src, twice)
				twicePitch := src.MustInvertPitch(axis.MustPitch()).MustInvertPitch(axis.MustPitch())
				assert.True(t, src.EqualElements(twicePitch), "axis %v", axis)
			}
			for _, kind := range []voice.RetrogradeKind{voice.RetrogradeBoth, voice.RetrogradePitches, voice.RetrogradeRhythm} {
				twice := src.MustRetrograde(kind).MustRetrograde(kind)
				assert.True(t, src.Equal(twice), "%v: want %v, got %v", kind, src, twice)
			}
		})
	}
}

// П-15: the retrograde keeps the number of elements, sounds and rests and
// the length.
func TestRetrogradeKeepsComposition(t *testing.T) {
	t.Parallel()
	for name, l := range invariantLines() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			for _, kind := range []voice.RetrogradeKind{voice.RetrogradeBoth, voice.RetrogradePitches, voice.RetrogradeRhythm} {
				got := l.MustRetrograde(kind)
				assert.Equal(t, l.MustLen(), got.MustLen())
				assert.Equal(t, l.MustSoundCount(), got.MustSoundCount())
				assert.True(t, l.MustLength().Equal(got.MustLength()))
			}
			both := l.MustRetrograde(voice.RetrogradeBoth)
			n := l.MustLen()
			for i := range n {
				assertElement(t, l.MustElement(i), both.MustElement(n-1-i))
			}
		})
	}
}
