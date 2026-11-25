package voiceleading_test

import (
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/problemsolving/voiceleading"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/voice"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Expectations: business-spec UD-1...UD-3 and architecture 8.

type wantClass struct {
	name     string // spelling in compact form
	voices   []int
	doubling voiceleading.DoublingKind
}

// assertToneClasses compares tone classes field by field with the presence
// rule: a sound class is set only for the reading by sound.
func assertToneClasses(t *testing.T, reading voiceleading.Reading, want []wantClass, got []voiceleading.ToneClass, kind string) {
	t.Helper()
	require.Len(t, got, len(want), kind)
	for i, w := range want {
		g := got[i]
		assert.True(t, spellingOf(t, w.name).Equal(g.Name), "%s %d: want %s, got %s", kind, i, w.name, g.Name)
		assert.Equal(t, w.voices, g.Voices, "%s %d voices", kind, i)
		assert.NotNil(t, g.Voices, "%s %d voices must be a non-nil slice", kind, i)
		assert.Equal(t, w.doubling, g.Doubling, "%s %d doubling", kind, i)
		if reading == voiceleading.BySound {
			want := traditionalSystem().MustClass(spellingOf(t, w.name))
			assert.True(t, want.Equal(g.Class), "%s %d class: want %v, got %v", kind, i, want, g.Class)
		} else {
			assert.False(t, g.Class.IsValid(), "%s %d: no sound class by spelling", kind, i)
			assert.Equal(t, pitchclass.Class{}, g.Class)
		}
	}
}

func TestDoublingsWithoutComposition(t *testing.T) {
	t.Parallel()
	spellingReading, sound := voiceleading.BySpelling, voiceleading.BySound
	none, unison, octave, both := voiceleading.NoDoubling, voiceleading.UnisonDoubling, voiceleading.OctaveDoubling, voiceleading.UnisonAndOctaveDoubling
	cases := []struct {
		name        string
		tokens      []string
		reading     voiceleading.Reading
		classes     []wantClass
		bassDoubled bool
	}{
		{"UD-1 C3 E4 G4 C5", []string{"C3", "E4", "G4", "C5"}, spellingReading, []wantClass{{"C", []int{0, 3}, octave}, {"E", []int{1}, none}, {"G", []int{2}, none}}, true},
		{"UD-1 C3 E4 G4 C5 by sound", []string{"C3", "E4", "G4", "C5"}, sound, []wantClass{{"C", []int{0, 3}, octave}, {"E", []int{1}, none}, {"G", []int{2}, none}}, true},
		{"UD-1 C3 G3 E4 E4", []string{"C3", "G3", "E4", "E4"}, spellingReading, []wantClass{{"C", []int{0}, none}, {"G", []int{1}, none}, {"E", []int{2, 3}, unison}}, false},
		{"UD-1 C3 E4 G4 B#4 by spelling", []string{"C3", "E4", "G4", "B#4"}, spellingReading, []wantClass{{"C", []int{0}, none}, {"E", []int{1}, none}, {"G", []int{2}, none}, {"B#", []int{3}, none}}, false},
		{"UD-1 C3 E4 G4 B#4 by sound", []string{"C3", "E4", "G4", "B#4"}, sound, []wantClass{{"C", []int{0, 3}, octave}, {"E", []int{1}, none}, {"G", []int{2}, none}}, true},
		{"UD-1 C3 B#3 C4 E4 by spelling", []string{"C3", "B#3", "C4", "E4"}, spellingReading, []wantClass{{"C", []int{0, 2}, octave}, {"B#", []int{1}, none}, {"E", []int{3}, none}}, true},
		{"UD-1 C3 B#3 C4 E4 by sound", []string{"C3", "B#3", "C4", "E4"}, sound, []wantClass{{"C", []int{0, 1, 2}, both}, {"E", []int{3}, none}}, true},
		{"the name by sound is the voice with the lowest number", []string{"B#3", "C3", "E4"}, sound, []wantClass{{"B#", []int{0, 1}, octave}, {"E", []int{2}, none}}, true},
		{"enharmonic unisons by sound", []string{"E4", "Fb4"}, sound, []wantClass{{"E", []int{0, 1}, unison}}, true},
		{"enharmonic unisons by spelling", []string{"E4", "Fb4"}, spellingReading, []wantClass{{"E", []int{0}, none}, {"Fb", []int{1}, none}}, false},
		{"one voice", []string{"C4"}, spellingReading, []wantClass{{"C", []int{0}, none}}, false},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := sonorityOf(t, tt.tokens...)
			d, err := s.Doublings(tt.reading, chord.Chord{})
			require.NoError(t, err)
			assert.Equal(t, tt.reading, d.Reading)
			assertToneClasses(t, tt.reading, tt.classes, d.Classes, "class")
			assert.True(t, d.HasBass)
			assert.Equal(t, tt.bassDoubled, d.BassDoubled)
			assert.False(t, d.Expected)
			assert.Empty(t, d.Missing)
			assert.Empty(t, d.Extra)
			assert.False(t, d.Complete)
			var doubled []wantClass
			for _, c := range tt.classes {
				if c.doubling != none {
					doubled = append(doubled, c)
				}
			}
			assertToneClasses(t, tt.reading, doubled, d.Doubled(), "doubled")
			assert.NotNil(t, d.Doubled())
			assert.Equal(t, d, s.MustDoublings(tt.reading, chord.Chord{}))
		})
	}
	empty := voiceleading.MustNewSonority(chord.MustNew(), nil)
	for _, reading := range []voiceleading.Reading{voiceleading.BySpelling, voiceleading.BySound} {
		d, err := empty.Doublings(reading, chord.Chord{})
		require.NoError(t, err)
		assert.Empty(t, d.Classes)
		assert.False(t, d.HasBass)
		assert.False(t, d.BassDoubled)
	}
}

func TestDoublingsWithComposition(t *testing.T) {
	t.Parallel()
	spellingReading, sound := voiceleading.BySpelling, voiceleading.BySound
	none, octave := voiceleading.NoDoubling, voiceleading.OctaveDoubling
	cases := []struct {
		name        string
		tokens      []string
		reading     voiceleading.Reading
		expected    func(t *testing.T) chord.Chord
		classes     []wantClass
		missing     []wantClass
		extra       []wantClass
		complete    bool
		bassDoubled bool
	}{
		{"UD-2 C3 C4 E4 C5 misses G and triples C", []string{"C3", "C4", "E4", "C5"}, spellingReading,
			func(t *testing.T) chord.Chord { return namesChord(t, "C", "E", "G") },
			[]wantClass{{"C", []int{0, 1, 3}, octave}, {"E", []int{2}, none}}, []wantClass{{"G", []int{}, none}}, nil, false, true},
		{"UD-2 C3 C4 E4 C5 by sound", []string{"C3", "C4", "E4", "C5"}, sound,
			func(t *testing.T) chord.Chord { return namesChord(t, "C", "E", "G") },
			[]wantClass{{"C", []int{0, 1, 3}, octave}, {"E", []int{2}, none}}, []wantClass{{"G", []int{}, none}}, nil, false, true},
		{"UD-2 C3 E4 G4 B#4 by spelling has an extra B#", []string{"C3", "E4", "G4", "B#4"}, spellingReading,
			func(t *testing.T) chord.Chord { return namesChord(t, "C", "E", "G") },
			[]wantClass{{"C", []int{0}, none}, {"E", []int{1}, none}, {"G", []int{2}, none}, {"B#", []int{3}, none}}, nil, []wantClass{{"B#", []int{3}, none}}, true, false},
		{"UD-2 C3 E4 G4 B#4 by sound has no extra class", []string{"C3", "E4", "G4", "B#4"}, sound,
			func(t *testing.T) chord.Chord { return namesChord(t, "C", "E", "G") },
			[]wantClass{{"C", []int{0, 3}, octave}, {"E", []int{1}, none}, {"G", []int{2}, none}}, nil, nil, true, true},
		{"UD-3 repeats and octaves of the composition collapse", []string{"C3", "E4", "G4", "C5"}, spellingReading,
			func(t *testing.T) chord.Chord {
				return chord.MustNew(noteOf(t, "C4", nil), nameOf(t, "E"), noteOf(t, "C5", nil), noteOf(t, "G3", nil), nameOf(t, "E"))
			},
			[]wantClass{{"C", []int{0, 3}, octave}, {"E", []int{1}, none}, {"G", []int{2}, none}}, nil, nil, true, true},
		{"the missing classes follow the order of the composition", []string{"E4"}, spellingReading,
			func(t *testing.T) chord.Chord { return namesChord(t, "G", "C", "E", "Bb") },
			[]wantClass{{"E", []int{0}, none}}, []wantClass{{"G", []int{}, none}, {"C", []int{}, none}, {"Bb", []int{}, none}}, nil, false, false},
		{"a missing class by sound is named by its first note in the composition", []string{"E4"}, sound,
			func(t *testing.T) chord.Chord { return namesChord(t, "Ab", "G#", "E") },
			[]wantClass{{"E", []int{0}, none}}, []wantClass{{"Ab", []int{}, none}}, nil, false, false},
		{"by spelling an enharmonic composition is distinct", []string{"G#4"}, spellingReading,
			func(t *testing.T) chord.Chord { return namesChord(t, "Ab") },
			[]wantClass{{"G#", []int{0}, none}}, []wantClass{{"Ab", []int{}, none}}, []wantClass{{"G#", []int{0}, none}}, false, false},
		{"extra classes follow the order of the voices", []string{"D3", "C4", "F#4", "A4"}, spellingReading,
			func(t *testing.T) chord.Chord { return namesChord(t, "D", "F#", "A") },
			[]wantClass{{"D", []int{0}, none}, {"C", []int{1}, none}, {"F#", []int{2}, none}, {"A", []int{3}, none}}, nil, []wantClass{{"C", []int{1}, none}}, true, false},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := sonorityOf(t, tt.tokens...)
			d, err := s.Doublings(tt.reading, tt.expected(t))
			require.NoError(t, err)
			assert.True(t, d.Expected)
			assertToneClasses(t, tt.reading, tt.classes, d.Classes, "class")
			assertToneClasses(t, tt.reading, tt.missing, d.Missing, "missing")
			assertToneClasses(t, tt.reading, tt.extra, d.Extra, "extra")
			assert.Equal(t, tt.complete, d.Complete)
			assert.Equal(t, tt.bassDoubled, d.BassDoubled)
		})
	}
}

func TestDoublingsCompositionIsReadByNameOnly(t *testing.T) {
	t.Parallel()
	src := &testSource{current: settings.Standard()}
	tuned, err := note.FromSpelling(spelling.G(), cents.MustParse("25"), src)
	require.NoError(t, err)
	src.reset(nil)
	s := sonorityOf(t, "C3", "E4", "G4", "C5")
	for _, reading := range []voiceleading.Reading{voiceleading.BySpelling, voiceleading.BySound} {
		d, err := s.Doublings(reading, chord.MustNew(nameOf(t, "C"), nameOf(t, "E"), tuned))
		require.NoError(t, err)
		assert.True(t, d.Complete)
		assert.Empty(t, d.Missing)
		assert.Empty(t, d.Extra)
	}
	assert.Zero(t, src.count())
	// By spelling a name is not checked against the system of the sonority.
	half, err := note.FromSpelling(spelling.MustParse("C+1/2"), cents.Zero(), nil)
	require.NoError(t, err)
	d, err := s.Doublings(voiceleading.BySpelling, chord.MustNew(nameOf(t, "C"), half))
	require.NoError(t, err)
	assertToneClasses(t, voiceleading.BySpelling, []wantClass{{"C+1/2", []int{}, voiceleading.NoDoubling}}, d.Missing, "missing")
}

func TestDoublingsFailureOrder(t *testing.T) {
	t.Parallel()
	s := sonorityOf(t, "C3", "E4", "G4", "C5")
	half, err := note.FromSpelling(spelling.MustParse("C+1/2"), cents.Zero(), nil)
	require.NoError(t, err)
	t.Run("an unset receiver", func(t *testing.T) {
		t.Parallel()
		for _, unset := range []*voiceleading.Sonority{nil, new(voiceleading.Sonority)} {
			d, err := unset.Doublings(voiceleading.Reading(9), chord.MustNew())
			assert.Equal(t, voiceleading.Doublings{}, d)
			assertSentinel(t, err, voiceleading.ErrUnset)
		}
	})
	t.Run("an unknown reading precedes an empty composition", func(t *testing.T) {
		t.Parallel()
		d, err := s.Doublings(voiceleading.Reading(9), chord.MustNew())
		assert.Equal(t, voiceleading.Doublings{}, d)
		assertSentinel(t, err, voiceleading.ErrOption)
		assertNoChordError(t, err)
		assertMustPanics(t, err, func() { s.MustDoublings(voiceleading.Reading(9), chord.Chord{}) })
	})
	t.Run("an empty composition", func(t *testing.T) {
		t.Parallel()
		for _, reading := range []voiceleading.Reading{voiceleading.BySpelling, voiceleading.BySound} {
			d, err := s.Doublings(reading, chord.MustNew())
			assert.Equal(t, voiceleading.Doublings{}, d)
			assertChordError(t, err, voiceleading.ErrOption, 0, -1)
			assertSentinel(t, err, voiceleading.ErrOption)
		}
	})
	t.Run("an inadmissible name by sound", func(t *testing.T) {
		t.Parallel()
		d, err := s.Doublings(voiceleading.BySound, chord.MustNew(nameOf(t, "C"), nameOf(t, "E"), half))
		assert.Equal(t, voiceleading.Doublings{}, d)
		assertChordError(t, err, system.ErrInadmissible, 0, 2)
		assertSentinel(t, err, nil)
	})
	t.Run("a sonority without a system cannot read a composition by sound", func(t *testing.T) {
		t.Parallel()
		empty := voiceleading.MustNewSonority(chord.MustNew(), nil)
		d, err := empty.Doublings(voiceleading.BySound, namesChord(t, "C", "E", "G"))
		assert.Equal(t, voiceleading.Doublings{}, d)
		assertSentinel(t, err, voiceleading.ErrNoSystem)
		assertNoChordError(t, err)
		assertNoPartError(t, err)
		got, err := empty.Doublings(voiceleading.BySpelling, namesChord(t, "C", "E", "G"))
		require.NoError(t, err)
		assertToneClasses(t, voiceleading.BySpelling, []wantClass{{"C", []int{}, voiceleading.NoDoubling}, {"E", []int{}, voiceleading.NoDoubling}, {"G", []int{}, voiceleading.NoDoubling}}, got.Missing, "missing")
	})
}

func TestDoublingsOfPartSonorities(t *testing.T) {
	t.Parallel()
	t.Run("an unordered sonority declares no bass", func(t *testing.T) {
		t.Parallel()
		a := analyzeParts(t, voiceleading.Unordered, traditionalSystem(), line(t, "C3"), line(t, "E4"), line(t, "C5"))
		s := a.MustSonorities()[0].Voices
		d := s.MustDoublings(voiceleading.BySpelling, chord.Chord{})
		assertToneClasses(t, voiceleading.BySpelling, []wantClass{{"C", []int{0, 2}, voiceleading.OctaveDoubling}, {"E", []int{1}, voiceleading.NoDoubling}}, d.Classes, "class")
		assert.False(t, d.HasBass)
		assert.False(t, d.BassDoubled)
	})
	t.Run("voices follow the sounding parts", func(t *testing.T) {
		t.Parallel()
		a := analyzeParts(t, voiceleading.LowToHigh, traditionalSystem(), line(t, "C3"), line(t, "-"), line(t, "G4"), line(t, "C5"))
		d := a.MustSonorities()[0].Voices.MustDoublings(voiceleading.BySound, namesChord(t, "C", "E", "G"))
		assertToneClasses(t, voiceleading.BySound, []wantClass{{"C", []int{0, 2}, voiceleading.OctaveDoubling}, {"G", []int{1}, voiceleading.NoDoubling}}, d.Classes, "class")
		assertToneClasses(t, voiceleading.BySound, []wantClass{{"E", []int{}, voiceleading.NoDoubling}}, d.Missing, "missing")
		assert.True(t, d.HasBass)
		assert.True(t, d.BassDoubled)
	})
	t.Run("one sounding voice in an analysis without a system", func(t *testing.T) {
		t.Parallel()
		a := analyzeParts(t, voiceleading.LowToHigh, nil, line(t, "C4", "-"), line(t, "-", "G4"))
		ps := a.MustSonorities()
		require.Len(t, ps, 2)
		for i, want := range []struct{ part, element int }{{0, 0}, {1, 1}} {
			s := ps[i].Voices
			d, err := s.Doublings(voiceleading.BySpelling, namesChord(t, "C", "G"))
			require.NoError(t, err)
			assert.Len(t, d.Classes, 1)
			_, err = s.Doublings(voiceleading.BySound, chord.Chord{})
			assertPartError(t, err, voiceleading.ErrNoSystem, want.part, want.element)
			assertSentinel(t, err, voiceleading.ErrNoSystem)
			// The order of failures puts the reading before the unmeasured voice.
			_, err = s.Doublings(voiceleading.Reading(5), chord.Chord{})
			assertSentinel(t, err, voiceleading.ErrOption)
			_, err = s.Doublings(voiceleading.BySound, chord.MustNew())
			assertChordError(t, err, voiceleading.ErrOption, 0, -1)
		}
	})
	t.Run("one sounding voice with an inadmissible spelling", func(t *testing.T) {
		t.Parallel()
		halfSound := voice.MustNewSound(note.MustFromPitch(pitch.MustFromASPN(spelling.MustParse("C+1/2"), 4), cents.Zero(), nil), voice.MustValueDuration(quarterValue()))
		a := analyzeParts(t, voiceleading.LowToHigh, traditionalSystem(), voice.MustNew(halfSound), line(t, "-"))
		s := a.MustSonorities()[0].Voices
		_, err := s.Doublings(voiceleading.BySpelling, chord.Chord{})
		require.NoError(t, err)
		_, err = s.Doublings(voiceleading.BySound, chord.Chord{})
		assertPartError(t, err, system.ErrInadmissible, 0, 0)
		assertSentinel(t, err, nil)
	})
}
