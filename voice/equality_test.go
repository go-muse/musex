package voice_test

import (
	"testing"

	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/voice"

	"github.com/stretchr/testify/assert"
)

// ОП-43…ОП-45, П-16, КП-33, КП-34, КП-37: the three equalities of lines.
func TestLineEquality(t *testing.T) {
	t.Parallel()
	melody := line(snd(c4(), q()), snd(d4(), q())).MustWithName("Melody").MustWithPlacement(voice.Soprano)
	tests := []struct {
		name                         string
		a, b                         *voice.Line
		equal, elements, byDurations bool
	}{
		{"same", kp1(), kp1(), true, true, true},
		{"КП-33 dotted and tied", line(snd(c4(), dq())), line(snd(c4(), qe())), false, false, true},
		{"written and value", line(snd(c4(), q()), rst(e8())), line(snd(c4(), val(1, 4)), rst(val(1, 8))), false, false, true},
		{"КП-34 other name", melody, melody.MustWithName("Top"), false, true, true},
		{"other placement", melody, melody.MustWithPlacement(voice.Alto), false, true, true},
		{"other thematic role", melody, melody.MustWithThematic(voice.Solo), false, true, true},
		{"other importance", melody, melody.MustWithImportance(voice.Filler), false, true, true},
		{"source is ignored", line(snd(c4(), q())), line(snd(free(spelling.C(), 4), q())), true, true, true},
		{"other correction", line(snd(c4(), q())), line(snd(withCents(spelling.C(), 4, 1), q())), false, false, false},
		{"enharmonic notes", line(snd(c4(), q())), line(snd(written(spelling.BSharp(), 3), q())), false, false, false},
		{"other fermata", line(snd(c4(), q())), line(snd(c4(), q()).MustWithFermata(true)), false, false, false},
		{"other length", line(snd(c4(), q())), line(snd(c4(), e8())), false, false, false},
		{"sound and rest", line(snd(c4(), q())), line(rst(q())), false, false, false},
		{"prefix", kp1(), kp1().MustSlice(0, 3), false, false, false},
		{"order", line(snd(c4(), q()), snd(d4(), q())), line(snd(d4(), q()), snd(c4(), q())), false, false, false},
		{"one tied sound and two sounds", line(snd(c4(), qe())), line(snd(c4(), q()), snd(c4(), e8())), false, false, false},
		{"empty lines", line(), line(), true, true, true},
		{"КП-37 empty and unset", line(), nil, false, false, false},
		{"КП-37 unset and zero", nil, &voice.Line{}, true, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.equal, tt.a.Equal(tt.b))
			assert.Equal(t, tt.equal, tt.b.Equal(tt.a))
			assert.Equal(t, tt.elements, tt.a.EqualElements(tt.b))
			assert.Equal(t, tt.elements, tt.b.EqualElements(tt.a))
			assert.Equal(t, tt.byDurations, tt.a.EqualDuration(tt.b))
			assert.Equal(t, tt.byDurations, tt.b.EqualDuration(tt.a))
		})
	}
}

// Architecture 6.15: == of *Line is pointer identity; a line equals itself.
func TestLineIdentity(t *testing.T) {
	t.Parallel()
	l := kp1()
	assert.True(t, l.Equal(l))
	other := kp1()
	assert.NotSame(t, l, other)
	assert.True(t, l.Equal(other))
}

// ОП-47: debugging strings are deterministic; only unset forms are part of
// the contract.
func TestStrings(t *testing.T) {
	t.Parallel()
	values := map[string]func() string{
		"line":     func() string { return named(kp4()).String() },
		"empty":    func() string { return line().String() },
		"sound":    func() string { return snd(c4(), qe()).MustWithFermata(true).String() },
		"rest":     func() string { return rst(q()).String() },
		"duration": func() string { return val(1, 3).String() },
		"zones":    func() string { return voice.DefaultZones().String() },
	}
	for name, f := range values {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			s := f()
			assert.NotEmpty(t, s)
			assert.Equal(t, s, f())
		})
	}
	assert.NotEqual(t, line().String(), "Line(unset)")
}
