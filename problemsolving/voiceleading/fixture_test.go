package voiceleading_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/problemsolving/voiceleading"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/voice"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var pair = voiceleading.Pair{First: 0, Second: 1}

func pos(n, d int) duration.Position { return duration.MustNewPosition(n, d) }

func sound(t *testing.T, name string, octave, numerator, denominator int, src settings.Source) voice.Element {
	t.Helper()
	if strings.HasSuffix(name, "#") {
		name = fmt.Sprintf("%s+%d", name[:1], len(name)-1)
	}
	if strings.HasSuffix(name, "b") {
		name = fmt.Sprintf("%s-%d", name[:1], len(name)-1)
	}
	n, err := note.FromPitch(pitch.MustFromASPN(spelling.MustParse(name), octave), cents.Zero(), src)
	require.NoError(t, err)
	e, err := voice.NewSound(n, voice.MustValueDuration(duration.MustNew(numerator, denominator)))
	require.NoError(t, err)
	return e
}

func rest(n, d int) voice.Element {
	return voice.MustNewRest(voice.MustValueDuration(duration.MustNew(n, d)))
}

// Each compact fixture token is a natural or altered spelling followed by one ASPN digit.
// A dash denotes a quarter rest. Extreme octaves use sound directly.
func line(t *testing.T, tokens ...string) *voice.Line {
	t.Helper()
	es := make([]voice.Element, len(tokens))
	for i, token := range tokens {
		if token == "-" {
			es[i] = rest(1, 4)
			continue
		}
		require.GreaterOrEqual(t, len(token), 2)
		es[i] = sound(t, token[:len(token)-1], int(token[len(token)-1]-'0'), 1, 4, nil)
	}
	return voice.MustNew(es...)
}

func analyze(t *testing.T, low, high []string) *voiceleading.Analysis {
	t.Helper()
	a, err := voiceleading.Analyze([]*voice.Line{line(t, low...), line(t, high...)}, voiceleading.Options{System: system.Traditional(), Order: voiceleading.LowToHigh})
	require.NoError(t, err)
	require.NotNil(t, a)
	return a
}

func assertInterval(t *testing.T, got interval.Interval, size int, quality interval.Quality, direction interval.Direction) {
	t.Helper()
	assert.True(t, got.IsValid())
	assert.Equal(t, size, got.Size())
	assert.Equal(t, quality, got.Quality())
	assert.Equal(t, direction, got.Direction())
}

func assertPosition(t *testing.T, want, got duration.Position) {
	t.Helper()
	assert.True(t, want.Equal(got), "position: want %s, got %s", want, got)
}

func assertVertical(t *testing.T, want, got voiceleading.Vertical) {
	t.Helper()
	assert.Equal(t, want.Pair, got.Pair)
	assertPosition(t, want.Start, got.Start)
	assertPosition(t, want.End, got.End)
	for _, e := range []struct{ want, got voiceleading.ElementAt }{{want.First, got.First}, {want.Second, got.Second}} {
		assert.Equal(t, e.want.Index, e.got.Index)
		assert.True(t, e.want.Element.Equal(e.got.Element))
		assertPosition(t, e.want.Start, e.got.Start)
		assertPosition(t, e.want.End, e.got.End)
	}
	assert.Equal(t, want.Sounding, got.Sounding)
	assert.Equal(t, want.Chromatic, got.Chromatic)
	assert.Equal(t, want.Specific.IsValid(), got.Specific.IsValid())
	if want.Specific.IsValid() {
		assert.True(t, want.Specific.Equal(got.Specific))
	}
}

func assertTransition(t *testing.T, want, got voiceleading.Transition) {
	t.Helper()
	assertVertical(t, want.From, got.From)
	assertVertical(t, want.To, got.To)
	assert.Equal(t, want.Motion, got.Motion)
	assert.Equal(t, want.FirstChromatic, got.FirstChromatic)
	assert.Equal(t, want.SecondChromatic, got.SecondChromatic)
	for _, v := range []struct{ want, got interval.Interval }{{want.First, got.First}, {want.Second, got.Second}} {
		assert.Equal(t, v.want.IsValid(), v.got.IsValid())
		if v.want.IsValid() {
			assert.True(t, v.want.Equal(v.got))
		}
	}
}
