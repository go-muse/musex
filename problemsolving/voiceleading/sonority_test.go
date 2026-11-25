package voiceleading_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/problemsolving/voiceleading"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Expectations: business-spec RS-1...RS-5 and architecture 6, 7 and 13.

func TestNewSonorityFailureOrder(t *testing.T) {
	t.Parallel()
	half := func(t *testing.T) note.Note {
		t.Helper()
		n, err := note.FromPitch(pitch.MustFromASPN(spelling.MustParse("C+1/2"), 4), cents.Zero(), nil)
		require.NoError(t, err)
		return n
	}
	cases := []struct {
		name   string
		notes  func(t *testing.T) chord.Chord
		sys    *system.System
		target error
		chord  int
		note   int
		reads  bool
	}{
		{"an unset chord precedes an unset system", func(*testing.T) chord.Chord { return chord.Chord{} }, new(system.System), chord.ErrUnset, 0, -1, false},
		{"an unset system precedes a note without an octave", func(t *testing.T) chord.Chord {
			return chord.MustNew(noteOf(t, "C4", nil), nameOf(t, "E"))
		}, new(system.System), system.ErrInvalid, -1, -1, false},
		{"a note without an octave precedes the choice of a system", func(t *testing.T) chord.Chord {
			return chord.MustNew(noteOf(t, "C4", nil), nameOf(t, "E"))
		}, nil, voiceleading.ErrNoOctave, 0, 1, false},
		{"a note without an octave precedes admissibility", func(t *testing.T) chord.Chord {
			return chord.MustNew(half(t), nameOf(t, "E"))
		}, traditionalSystem(), voiceleading.ErrNoOctave, 0, 1, false},
		{"one note without a system", func(t *testing.T) chord.Chord {
			return chord.MustNew(noteOf(t, "C4", nil))
		}, nil, voiceleading.ErrNoSystem, 0, 0, false},
		{"an inadmissible note under the explicit system", func(t *testing.T) chord.Chord {
			return chord.MustNew(noteOf(t, "C4", nil), half(t))
		}, traditionalSystem(), system.ErrInadmissible, 0, 1, false},
		{"an adjacent interval out of range names the upper note", func(t *testing.T) chord.Chord {
			return chord.MustNew(aspnNote(t, "C", math.MinInt), aspnNote(t, "C", math.MaxInt))
		}, traditionalSystem(), interval.ErrOutOfRange, 0, 1, false},
		{"the span is measured after the adjacent intervals", func(t *testing.T) chord.Chord {
			return chord.MustNew(aspnNote(t, "C", 0), aspnNote(t, "C", math.MaxInt/16), aspnNote(t, "C", math.MaxInt/8))
		}, traditionalSystem(), system.ErrOutOfRange, 0, 2, false},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			notes := tt.notes(t)
			s, err := voiceleading.NewSonority(notes, tt.sys)
			assert.Nil(t, s)
			require.Error(t, err)
			assert.ErrorIs(t, err, tt.target)
			if tt.chord >= 0 {
				assertChordError(t, err, tt.target, tt.chord, tt.note)
			} else {
				assertNoChordError(t, err)
			}
			switch tt.target {
			case chord.ErrUnset, system.ErrInvalid:
				assertSentinel(t, err, voiceleading.ErrUnset)
			case voiceleading.ErrNoOctave, voiceleading.ErrNoSystem:
				assertSentinel(t, err, tt.target)
			default:
				assertSentinel(t, err, nil)
			}
			assertMustPanics(t, err, func() { voiceleading.MustNewSonority(notes, tt.sys) })
		})
	}
}

func TestNewSonorityChoosesTheSystemFromItsNotes(t *testing.T) {
	t.Parallel()
	t.Run("one read per note in order", func(t *testing.T) {
		t.Parallel()
		trace := &sourceTrace{}
		sources := make([]*testSource, 3)
		notes := make([]note.Note, 3)
		for i, token := range []string{"C4", "E4", "G4"} {
			sources[i] = &testSource{current: settings.Standard(), id: i}
			notes[i] = noteOf(t, token, sources[i])
			sources[i].trace = trace
			sources[i].reset(settings.Standard())
		}
		s, err := voiceleading.NewSonority(chord.MustNew(notes...), nil)
		require.NoError(t, err)
		assert.Equal(t, []int{0, 1, 2}, trace.values)
		sys, ok, err := s.System()
		require.NoError(t, err)
		assert.True(t, ok)
		assert.True(t, traditionalSystem().Equal(sys))
		for _, src := range sources {
			src.reset(nil)
		}
		_ = s.MustSpacing()
		_ = s.MustClassifySpacing(voiceleading.SpacingAM)
		_ = s.MustCheckDistances(voiceleading.DistanceNorm{Rule: voiceleading.DistanceMT21SATB})
		_ = s.MustDoublings(voiceleading.BySound, namesChord(t, "C", "E", "G"))
		_ = s.MustNotes()
		for _, src := range sources {
			assert.Zero(t, src.count())
		}
	})
	t.Run("an unbound note fails before any read", func(t *testing.T) {
		t.Parallel()
		good := &testSource{current: settings.Standard()}
		failing := &testSource{current: settings.Standard()}
		notes := chord.MustNew(noteOf(t, "C4", good), noteOf(t, "E4", nil), noteOf(t, "G4", failing))
		good.reset(settings.Standard())
		failing.reset(nil)
		s, err := voiceleading.NewSonority(notes, nil)
		assert.Nil(t, s)
		assertChordError(t, err, voiceleading.ErrNoSystem, 0, 1)
		assertSentinel(t, err, voiceleading.ErrNoSystem)
		assert.Zero(t, good.count())
		assert.Zero(t, failing.count())
	})
	t.Run("all reads precede the comparison of systems", func(t *testing.T) {
		t.Parallel()
		one := &testSource{current: settings.Standard()}
		two := &testSource{current: settings.Standard()}
		three := &testSource{current: settings.Standard()}
		notes := chord.MustNew(noteOf(t, "C4", one), noteOf(t, "E4", two), noteOf(t, "G4", three))
		one.reset(settings.Standard())
		two.reset(settings19(t))
		three.reset(nil)
		s, err := voiceleading.NewSonority(notes, nil)
		assert.Nil(t, s)
		assertChordError(t, err, note.ErrUnsetSettings, 0, 2)
		assertSentinel(t, err, nil)
	})
	t.Run("different systems", func(t *testing.T) {
		t.Parallel()
		one := &testSource{current: settings.Standard()}
		two := &testSource{current: settings.Standard()}
		notes := chord.MustNew(noteOf(t, "C4", one), noteOf(t, "E4", two))
		one.reset(settings.Standard())
		two.reset(settings19(t))
		s, err := voiceleading.NewSonority(notes, nil)
		assert.Nil(t, s)
		assertChordError(t, err, voiceleading.ErrDifferentSystems, 0, 1)
		assertSentinel(t, err, voiceleading.ErrDifferentSystems)
	})
	t.Run("an explicit system never reads sources", func(t *testing.T) {
		t.Parallel()
		src := &testSource{current: settings.Standard()}
		notes := chord.MustNew(noteOf(t, "C4", src), noteOf(t, "D4", src))
		src.reset(nil)
		s := voiceleading.MustNewSonority(notes, system.Positions19())
		assert.Zero(t, src.count())
		assert.Equal(t, system.Step(3), s.MustSpacing().Adjacent[0].Chromatic)
	})
}

func TestStandaloneSonority(t *testing.T) {
	t.Parallel()
	t.Run("notes keep the order of record, repeats, corrections and sources", func(t *testing.T) {
		t.Parallel()
		src := &testSource{current: settings.Standard()}
		tuned, err := note.FromPitch(writtenPitch(t, "C3"), cents.MustParse("10"), src)
		require.NoError(t, err)
		s := voiceleading.MustNewSonority(chord.MustNew(noteOf(t, "E4", nil), tuned, noteOf(t, "E4", nil)), traditionalSystem())
		assert.Equal(t, 3, s.MustLen())
		notes := s.MustNotes()
		assertNotesAre(t, []string{"E4", "C3", "E4"}, notes)
		assert.True(t, notes[1].SameSource(tuned))
		assert.True(t, notes[1].Equal(tuned))
		assert.True(t, notes[1].SameCents(tuned))
		parts, ok, err := s.Parts()
		require.NoError(t, err)
		assert.Nil(t, parts)
		assert.False(t, ok)
		silent, err := s.Silent()
		require.NoError(t, err)
		assert.Empty(t, silent)
		assert.Equal(t, voiceleading.LowToHigh, s.MustOrder())
		// Corrections take no part in measurements.
		assertVoiceInterval(t, s.MustSpacing().Adjacent[0], 0, 1, "-M10", -16)
	})
	t.Run("an empty chord is a set sonority without voices", func(t *testing.T) {
		t.Parallel()
		for _, sys := range []*system.System{nil, traditionalSystem()} {
			s, err := voiceleading.NewSonority(chord.MustNew(), sys)
			require.NoError(t, err)
			assert.True(t, s.IsSet())
			assert.Equal(t, 0, s.MustLen())
			assert.Empty(t, s.MustNotes())
			got, ok := s.MustSystem()
			assert.Equal(t, sys != nil, ok)
			if sys == nil {
				assert.Nil(t, got)
			} else {
				assert.True(t, sys.Equal(got))
			}
			spacing := s.MustSpacing()
			assert.Empty(t, spacing.Adjacent)
			assert.Equal(t, voiceleading.VoiceInterval{}, spacing.Span)
			assert.False(t, spacing.WithinOctave)
		}
	})
	t.Run("an explicit system is kept with one note", func(t *testing.T) {
		t.Parallel()
		s := voiceleading.MustNewSonority(chordOf(t, "C4"), system.Positions19())
		sys, ok := s.MustSystem()
		assert.True(t, ok)
		assert.True(t, system.Positions19().Equal(sys))
	})
	t.Run("equality", func(t *testing.T) {
		t.Parallel()
		src := &testSource{current: settings.Standard()}
		bound := voiceleading.MustNewSonority(chord.MustNew(noteOf(t, "C4", src), noteOf(t, "E4", src)), traditionalSystem())
		plain := sonorityOf(t, "C4", "E4")
		assert.True(t, bound.Equal(plain), "sources take no part in equality")
		assert.True(t, plain.Equal(bound))
		assert.True(t, plain.Equal(plain))
		assert.False(t, plain.Equal(sonorityOf(t, "C4", "F4")))
		assert.False(t, plain.Equal(sonorityOf(t, "E4", "C4")))
		assert.False(t, plain.Equal(voiceleading.MustNewSonority(chordOf(t, "C4", "E4"), system.Positions19())))
		tuned, err := note.FromPitch(writtenPitch(t, "E4"), cents.MustParse("5"), nil)
		require.NoError(t, err)
		assert.False(t, plain.Equal(voiceleading.MustNewSonority(chord.MustNew(noteOf(t, "C4", nil), tuned), traditionalSystem())))
		assert.False(t, voiceleading.MustNewSonority(chord.MustNew(), nil).Equal(voiceleading.MustNewSonority(chord.MustNew(), traditionalSystem())))
		assert.True(t, voiceleading.MustNewSonority(chord.MustNew(), nil).Equal(voiceleading.MustNewSonority(chord.MustNew(), nil)))
		// A sonority of parts is never equal to a standalone one with the same notes.
		a := analyzeParts(t, voiceleading.LowToHigh, traditionalSystem(), line(t, "C4"), line(t, "E4"))
		parts := a.MustSonorities()[0].Voices
		assert.False(t, parts.Equal(plain))
		assert.False(t, plain.Equal(parts))
		var unset *voiceleading.Sonority
		assert.True(t, unset.Equal(nil))
		assert.True(t, unset.Equal(new(voiceleading.Sonority)))
		assert.True(t, new(voiceleading.Sonority).Equal(unset))
		assert.False(t, unset.Equal(plain))
		assert.False(t, plain.Equal(unset))
	})
}

func TestSonoritySpacing(t *testing.T) {
	t.Parallel()
	type adjacent struct {
		iv        string
		chromatic int
	}
	cases := []struct {
		name     string
		tokens   []string
		adjacent []adjacent
		span     adjacent
		within   bool
	}{
		{"C3 E4 G4 C5", []string{"C3", "E4", "G4", "C5"}, []adjacent{{"M10", 16}, {"m3", 3}, {"P4", 5}}, adjacent{"P15", 24}, false},
		{"C4 E4 G4 C5 spans an octave", []string{"C4", "E4", "G4", "C5"}, []adjacent{{"M3", 4}, {"m3", 3}, {"P4", 5}}, adjacent{"P8", 12}, true},
		{"RS-1 B#3 to C5 is a diminished ninth of 12 positions", []string{"B#3", "E4", "G4", "C5"}, []adjacent{{"d4", 4}, {"m3", 3}, {"P4", 5}}, adjacent{"d9", 12}, false},
		{"a crossed pair gives a descending interval", []string{"C3", "G4", "E4", "C5"}, []adjacent{{"P12", 19}, {"-m3", -3}, {"m6", 8}}, adjacent{"P15", 24}, false},
		{"an augmented octave exceeds the octave", []string{"C4", "C#5"}, []adjacent{{"A8", 13}}, adjacent{"A8", 13}, false},
		{"a descending span within an octave", []string{"C5", "C4"}, []adjacent{{"-P8", -12}}, adjacent{"-P8", -12}, true},
		{"a unison", []string{"C4", "C4"}, []adjacent{{"P1", 0}}, adjacent{"P1", 0}, true},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := sonorityOf(t, tt.tokens...)
			got, err := s.Spacing()
			require.NoError(t, err)
			require.Len(t, got.Adjacent, len(tt.adjacent))
			for k, w := range tt.adjacent {
				assertVoiceInterval(t, got.Adjacent[k], k, k+1, w.iv, w.chromatic, "pair", k)
			}
			assertVoiceInterval(t, got.Span, 0, len(tt.tokens)-1, tt.span.iv, tt.span.chromatic, "span")
			assert.Equal(t, tt.within, got.WithinOctave)
			assert.Equal(t, got, s.MustSpacing())
		})
	}
	one := sonorityOf(t, "C4").MustSpacing()
	assert.Empty(t, one.Adjacent)
	assert.Equal(t, voiceleading.VoiceInterval{}, one.Span)
	assert.False(t, one.WithinOctave)
}

// silentSonority is a part sonority C3, (rest), G4, E4: part 1 is silent and
// the sounding voices 1 and 2 are crossed.
func silentSonority(t *testing.T, order voiceleading.Order) *voiceleading.Sonority {
	t.Helper()
	a := analyzeParts(t, order, traditionalSystem(), line(t, "C3"), line(t, "-"), line(t, "G4"), line(t, "E4"))
	ps := a.MustSonorities()
	require.Len(t, ps, 1)
	return ps[0].Voices
}

func TestSpacingNeedsTheNominalOrder(t *testing.T) {
	t.Parallel()
	s := silentSonority(t, voiceleading.Unordered)
	assert.Equal(t, voiceleading.Unordered, s.MustOrder())
	_, err := s.Spacing()
	assertSentinel(t, err, voiceleading.ErrNoOrder)
	assertMustPanics(t, err, func() { s.MustSpacing() })
	// The order is checked before the definition and the norm.
	c, err := s.ClassifySpacing(voiceleading.SpacingUnset)
	assert.Equal(t, voiceleading.SpacingClassification{}, c)
	assertSentinel(t, err, voiceleading.ErrNoOrder)
	d, err := s.CheckDistances(voiceleading.DistanceNorm{})
	assert.Nil(t, d)
	assertSentinel(t, err, voiceleading.ErrNoOrder)
	// Doublings need no order.
	_, err = s.Doublings(voiceleading.BySpelling, chord.Chord{})
	require.NoError(t, err)
}

const (
	meSource  = "Muzykalnaya entsiklopediya, Akkord"
	meScope   = "exactly four voices; tenor-alto and alto-soprano; a fourth or narrower is close"
	amSource  = "Alekseev and Myasoedov, Elementary Music Theory, 69-71"
	amScope   = "two or more voices; every adjacent pair including the bass; a fourth or narrower is close"
	omtSource = "Open Music Theory, glossary: closed and open spacing"
	omtScope  = "two or more voices; closed when the span does not exceed P8"
)

func TestClassifySpacing(t *testing.T) {
	t.Parallel()
	// Business-spec RS-2: the whole table.
	cases := []struct {
		name        string
		tokens      []string
		me, am, omt voiceleading.SpacingClass
	}{
		{"C3 E4 G4 C5", []string{"C3", "E4", "G4", "C5"}, voiceleading.CloseSpacing, voiceleading.MixedSpacing, voiceleading.OMTOpenSpacing},
		{"C4 E4 G4 C5", []string{"C4", "E4", "G4", "C5"}, voiceleading.CloseSpacing, voiceleading.CloseSpacing, voiceleading.OMTClosedSpacing},
		{"C3 C4 G4 E5", []string{"C3", "C4", "G4", "E5"}, voiceleading.OpenSpacing, voiceleading.OpenSpacing, voiceleading.OMTOpenSpacing},
		{"C3 E4 C5 E5", []string{"C3", "E4", "C5", "E5"}, voiceleading.MixedSpacing, voiceleading.MixedSpacing, voiceleading.OMTOpenSpacing},
		{"C4 E4 G4", []string{"C4", "E4", "G4"}, voiceleading.SpacingClassUnset, voiceleading.CloseSpacing, voiceleading.OMTClosedSpacing},
		{"sevenths and compounds are wide for ME", []string{"C3", "C4", "B4", "D6"}, voiceleading.OpenSpacing, voiceleading.OpenSpacing, voiceleading.OMTOpenSpacing},
		{"unisons, seconds and fourths are close", []string{"C4", "C4", "D4", "G4"}, voiceleading.CloseSpacing, voiceleading.CloseSpacing, voiceleading.OMTClosedSpacing},
		{"two voices for AM and OMT", []string{"C4", "G4"}, voiceleading.SpacingClassUnset, voiceleading.OpenSpacing, voiceleading.OMTClosedSpacing},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := sonorityOf(t, tt.tokens...)
			spacing := s.MustSpacing()
			n := len(tt.tokens)
			definitions := []struct {
				def           voiceleading.SpacingDefinition
				class         voiceleading.SpacingClass
				source, scope string
				considered    []voiceleading.VoiceInterval
				applicable    bool
			}{
				{voiceleading.SpacingME, tt.me, meSource, meScope, nil, n == 4},
				{voiceleading.SpacingAM, tt.am, amSource, amScope, spacing.Adjacent, true},
				{voiceleading.SpacingOMT, tt.omt, omtSource, omtScope, []voiceleading.VoiceInterval{spacing.Span}, true},
			}
			if n == 4 {
				definitions[0].considered = []voiceleading.VoiceInterval{spacing.Adjacent[1], spacing.Adjacent[2]}
			}
			for _, d := range definitions {
				got, err := s.ClassifySpacing(d.def)
				require.NoError(t, err)
				assert.Equal(t, d.def, got.Definition)
				assert.Equal(t, d.source, got.Source, "%v", d.def)
				assert.Equal(t, d.scope, got.Scope, "%v", d.def)
				assert.Equal(t, d.applicable, got.Applicable, "%v", d.def)
				assert.Equal(t, d.class, got.Class, "%v", d.def)
				assert.Empty(t, got.Crossed)
				assert.Empty(t, got.Silent)
				if d.applicable {
					assert.Empty(t, got.Reason)
					assert.Equal(t, d.considered, got.Considered, "%v", d.def)
				} else {
					assert.Equal(t, voiceleading.ReasonVoiceCount, got.Reason)
					assert.Empty(t, got.Considered)
				}
				assert.Equal(t, got, s.MustClassifySpacing(d.def))
			}
		})
	}
}

func TestClassifySpacingInapplicability(t *testing.T) {
	t.Parallel()
	crossed := sonorityOf(t, "C3", "G4", "E4", "C5")
	crossedPair := crossed.MustSpacing().Adjacent[1]
	assertVoiceInterval(t, crossedPair, 1, 2, "-m3", -3)
	silent := silentSonority(t, voiceleading.LowToHigh)
	silentPair := silent.MustSpacing().Adjacent[1]
	assertVoiceInterval(t, silentPair, 1, 2, "-m3", -3)
	threeCrossed := sonorityOf(t, "C4", "G4", "E4")
	threePair := threeCrossed.MustSpacing().Adjacent[1]
	assertVoiceInterval(t, threePair, 1, 2, "-m3", -3)
	cases := []struct {
		name    string
		s       *voiceleading.Sonority
		def     voiceleading.SpacingDefinition
		reason  string
		crossed []voiceleading.VoiceInterval
		silent  []int
	}{
		{"RS-5 a crossed tenor and alto for ME", crossed, voiceleading.SpacingME, voiceleading.ReasonCrossedVoices, []voiceleading.VoiceInterval{crossedPair}, []int{}},
		{"RS-5 a crossed tenor and alto for AM", crossed, voiceleading.SpacingAM, voiceleading.ReasonCrossedVoices, []voiceleading.VoiceInterval{crossedPair}, []int{}},
		{"RS-5 a crossed tenor and alto for OMT", crossed, voiceleading.SpacingOMT, voiceleading.ReasonCrossedVoices, []voiceleading.VoiceInterval{crossedPair}, []int{}},
		{"silent parts come first", silent, voiceleading.SpacingAM, voiceleading.ReasonSilentParts, []voiceleading.VoiceInterval{silentPair}, []int{1}},
		{"silent parts precede the voice count", silent, voiceleading.SpacingME, voiceleading.ReasonSilentParts, []voiceleading.VoiceInterval{silentPair}, []int{1}},
		{"the voice count precedes crossings", threeCrossed, voiceleading.SpacingME, voiceleading.ReasonVoiceCount, []voiceleading.VoiceInterval{threePair}, []int{}},
		{"one voice for AM", sonorityOf(t, "C4"), voiceleading.SpacingAM, voiceleading.ReasonVoiceCount, nil, []int{}},
		{"one voice for OMT", sonorityOf(t, "C4"), voiceleading.SpacingOMT, voiceleading.ReasonVoiceCount, nil, []int{}},
		{"no voices", voiceleading.MustNewSonority(chord.MustNew(), nil), voiceleading.SpacingOMT, voiceleading.ReasonVoiceCount, nil, []int{}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.s.ClassifySpacing(tt.def)
			require.NoError(t, err)
			assert.Equal(t, tt.def, got.Definition)
			assert.False(t, got.Applicable)
			assert.Equal(t, tt.reason, got.Reason)
			assert.Equal(t, voiceleading.SpacingClassUnset, got.Class)
			assert.Empty(t, got.Considered)
			assert.NotEmpty(t, got.Source)
			assert.NotEmpty(t, got.Scope)
			if tt.crossed == nil {
				assert.Empty(t, got.Crossed)
			} else {
				assert.Equal(t, tt.crossed, got.Crossed)
			}
			assert.ElementsMatch(t, tt.silent, got.Silent)
			assert.Equal(t, len(tt.silent), len(got.Silent))
		})
	}
	sourceOf := map[voiceleading.SpacingDefinition][2]string{
		voiceleading.SpacingME: {meSource, meScope}, voiceleading.SpacingAM: {amSource, amScope}, voiceleading.SpacingOMT: {omtSource, omtScope},
	}
	for def, want := range sourceOf {
		got := sonorityOf(t, "C4").MustClassifySpacing(def)
		assert.Equal(t, want[0], got.Source)
		assert.Equal(t, want[1], got.Scope)
	}
	for _, def := range []voiceleading.SpacingDefinition{voiceleading.SpacingUnset, voiceleading.SpacingDefinition(9)} {
		got, err := crossed.ClassifySpacing(def)
		assert.Equal(t, voiceleading.SpacingClassification{}, got)
		assertSentinel(t, err, voiceleading.ErrOption)
		assertMustPanics(t, err, func() { crossed.MustClassifySpacing(def) })
	}
}

const (
	omtSATBScope  = "four voices; soprano-alto and alto-tenor at most P8, tenor-bass at most P12"
	mt21SATBScope = "four voices; soprano-alto and alto-tenor at most P8, tenor-bass unlimited"
	firstScope    = "two voices; at most P12; intervals beyond P8 are remarked"
	customScope   = "adjacent voices within the given limits"
)

type wantDistance struct {
	lower int
	iv    string
	limit string
}

func assertDistanceViolations(t *testing.T, s *voiceleading.Sonority, want []wantDistance, got []voiceleading.DistanceViolation, kind string) {
	t.Helper()
	require.Len(t, got, len(want), kind)
	adjacent := s.MustSpacing().Adjacent
	for i, w := range want {
		assert.Equal(t, adjacent[w.lower], got[i].Pair, "%s %d pair", kind, i)
		assertIntervalText(t, w.iv, got[i].Pair.Interval, kind, i)
		assertIntervalText(t, w.limit, got[i].Limit, kind, i, "limit")
	}
}

func TestCheckDistances(t *testing.T) {
	t.Parallel()
	omt := voiceleading.DistanceNorm{Rule: voiceleading.DistanceOMTSATB}
	mt21 := voiceleading.DistanceNorm{Rule: voiceleading.DistanceMT21SATB}
	first := voiceleading.DistanceNorm{Rule: voiceleading.DistanceOMTFirstSpecies}
	custom := func(limits ...interval.Interval) voiceleading.DistanceNorm {
		return voiceleading.DistanceNorm{Rule: voiceleading.DistanceCustom, Limits: limits}
	}
	p8, p12, p5 := interval.PerfectOctave(), interval.PerfectTwelfth(), interval.PerfectFifth()
	quarter := func(t *testing.T) *voiceleading.Sonority {
		t.Helper()
		e := pitch.MustFromASPN(spelling.MustParse("E-1/2"), 4)
		n, err := note.FromPitch(e, cents.Zero(), nil)
		require.NoError(t, err)
		return voiceleading.MustNewSonority(chord.MustNew(noteOf(t, "C4", nil), n), system.QuarterTones24())
	}
	cases := []struct {
		name       string
		s          func(t *testing.T) *voiceleading.Sonority
		norm       voiceleading.DistanceNorm
		rule       string
		source     string
		scope      string
		violations []wantDistance
		remarks    []wantDistance
	}{
		{"RS-4 E5 C4 G3 C3 by OMT", func(t *testing.T) *voiceleading.Sonority { return sonorityOf(t, "C3", "G3", "C4", "E5") },
			omt, "omt-satb", "Open Music Theory, SATB style", omtSATBScope, []wantDistance{{2, "M10", "P8"}}, nil},
		{"RS-4 E5 C4 G3 C3 by MT21", func(t *testing.T) *voiceleading.Sonority { return sonorityOf(t, "C3", "G3", "C4", "E5") },
			mt21, "mt21-satb", "Music Theory for the 21st-Century Classroom, 26.6", mt21SATBScope, []wantDistance{{2, "M10", "P8"}}, nil},
		{"RS-4 C3 C4 G4 E5 by OMT", func(t *testing.T) *voiceleading.Sonority { return sonorityOf(t, "C3", "C4", "G4", "E5") },
			omt, "omt-satb", "Open Music Theory, SATB style", omtSATBScope, nil, nil},
		{"RS-4 C3 C4 G4 E5 by MT21", func(t *testing.T) *voiceleading.Sonority { return sonorityOf(t, "C3", "C4", "G4", "E5") },
			mt21, "mt21-satb", "Music Theory for the 21st-Century Classroom, 26.6", mt21SATBScope, nil, nil},
		{"RS-4 C3 E3 B#3 C5 a diminished ninth", func(t *testing.T) *voiceleading.Sonority { return sonorityOf(t, "C3", "E3", "B#3", "C5") },
			omt, "omt-satb", "Open Music Theory, SATB style", omtSATBScope, []wantDistance{{2, "d9", "P8"}}, nil},
		{"RS-4 C2 C4 G4 E5 by OMT", func(t *testing.T) *voiceleading.Sonority { return sonorityOf(t, "C2", "C4", "G4", "E5") },
			omt, "omt-satb", "Open Music Theory, SATB style", omtSATBScope, []wantDistance{{0, "P15", "P12"}}, nil},
		{"RS-4 C2 C4 G4 E5 by MT21", func(t *testing.T) *voiceleading.Sonority { return sonorityOf(t, "C2", "C4", "G4", "E5") },
			mt21, "mt21-satb", "Music Theory for the 21st-Century Classroom, 26.6", mt21SATBScope, nil, nil},
		{"an octave does not exceed an octave", func(t *testing.T) *voiceleading.Sonority { return sonorityOf(t, "C3", "C4", "C5", "C6") },
			omt, "omt-satb", "Open Music Theory, SATB style", omtSATBScope, nil, nil},
		{"an augmented octave exceeds an octave", func(t *testing.T) *voiceleading.Sonority { return sonorityOf(t, "C3", "C4", "C5", "C#6") },
			omt, "omt-satb", "Open Music Theory, SATB style", omtSATBScope, []wantDistance{{2, "A8", "P8"}}, nil},
		{"a crossed pair is measured by its size", func(t *testing.T) *voiceleading.Sonority { return sonorityOf(t, "C3", "G4", "E4", "C5") },
			omt, "omt-satb", "Open Music Theory, SATB style", omtSATBScope, nil, nil},
		{"a crossed wide pair violates", func(t *testing.T) *voiceleading.Sonority { return sonorityOf(t, "C3", "G3", "E5", "C4") },
			mt21, "mt21-satb", "Music Theory for the 21st-Century Classroom, 26.6", mt21SATBScope, []wantDistance{{1, "M13", "P8"}, {2, "-M10", "P8"}}, nil},
		{"first species beyond a twelfth", func(t *testing.T) *voiceleading.Sonority { return sonorityOf(t, "C3", "E5") },
			first, "omt-first-species", "Open Music Theory, first species", firstScope, []wantDistance{{0, "M17", "P12"}}, []wantDistance{{0, "M17", "P8"}}},
		{"first species at a twelfth is remarked only", func(t *testing.T) *voiceleading.Sonority { return sonorityOf(t, "C3", "G4") },
			first, "omt-first-species", "Open Music Theory, first species", firstScope, nil, []wantDistance{{0, "P12", "P8"}}},
		{"first species at a tenth is remarked only", func(t *testing.T) *voiceleading.Sonority { return sonorityOf(t, "C4", "E5") },
			first, "omt-first-species", "Open Music Theory, first species", firstScope, nil, []wantDistance{{0, "M10", "P8"}}},
		{"first species at an octave", func(t *testing.T) *voiceleading.Sonority { return sonorityOf(t, "C4", "C5") },
			first, "omt-first-species", "Open Music Theory, first species", firstScope, nil, nil},
		{"first species of crossed voices", func(t *testing.T) *voiceleading.Sonority { return sonorityOf(t, "C5", "C3") },
			first, "omt-first-species", "Open Music Theory, first species", firstScope, []wantDistance{{0, "-P15", "P12"}}, []wantDistance{{0, "-P15", "P8"}}},
		{"one custom limit for every pair", func(t *testing.T) *voiceleading.Sonority { return sonorityOf(t, "C3", "E4", "G4", "C5") },
			custom(p5), "custom", "composer", customScope, []wantDistance{{0, "M10", "P5"}}, nil},
		{"one custom limit for a close chord", func(t *testing.T) *voiceleading.Sonority { return sonorityOf(t, "C4", "E4", "G4", "C5") },
			custom(p5), "custom", "composer", customScope, nil, nil},
		{"a custom limit is reported as given", func(t *testing.T) *voiceleading.Sonority { return sonorityOf(t, "C3", "E4", "G4", "C5") },
			custom(p5.MustDown()), "custom", "composer", customScope, []wantDistance{{0, "M10", "-P5"}}, nil},
		{"a custom pair without a limit", func(t *testing.T) *voiceleading.Sonority { return sonorityOf(t, "C2", "C4", "G4", "E5") },
			custom(interval.Interval{}, p8, p8), "custom", "composer", customScope, nil, nil},
		{"custom limits per pair", func(t *testing.T) *voiceleading.Sonority { return sonorityOf(t, "C3", "G3", "C4", "E5") },
			custom(p12, p8, p8), "custom", "composer", customScope, []wantDistance{{2, "M10", "P8"}}, nil},
		{"custom limits on two voices", func(t *testing.T) *voiceleading.Sonority { return sonorityOf(t, "C4", "A4") },
			custom(p5), "custom", "composer", customScope, []wantDistance{{0, "M6", "P5"}}, nil},
		{"a neutral third is below a major third", quarter, custom(interval.MajorThird()), "custom", "composer", customScope, nil, nil},
		{"a neutral third exceeds a minor third", quarter, custom(interval.MinorThird()), "custom", "composer", customScope, []wantDistance{{0, "n3", "m3"}}, nil},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := tt.s(t)
			got, err := s.CheckDistances(tt.norm)
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tt.rule, got.Rule)
			assert.Equal(t, tt.source, got.Source)
			assert.Equal(t, tt.scope, got.Scope)
			assert.True(t, got.Applicable)
			assert.Empty(t, got.Reason)
			assert.Empty(t, got.Silent)
			assertDistanceViolations(t, s, tt.violations, got.Violations, "violation")
			assertDistanceViolations(t, s, tt.remarks, got.Remarks, "remark")
			assert.Equal(t, got, s.MustCheckDistances(tt.norm))
		})
	}
}

func TestCheckDistancesInapplicabilityAndOptions(t *testing.T) {
	t.Parallel()
	three := sonorityOf(t, "C4", "E4", "G4")
	five := sonorityOf(t, "C3", "G3", "C4", "E4", "G4")
	one := sonorityOf(t, "C4")
	silent := silentSonority(t, voiceleading.LowToHigh)
	p8 := interval.PerfectOctave()
	inapplicable := []struct {
		name   string
		s      *voiceleading.Sonority
		norm   voiceleading.DistanceNorm
		reason string
		silent []int
	}{
		{"OMT SATB with three voices", three, voiceleading.DistanceNorm{Rule: voiceleading.DistanceOMTSATB}, voiceleading.ReasonVoiceCount, []int{}},
		{"MT21 SATB with five voices", five, voiceleading.DistanceNorm{Rule: voiceleading.DistanceMT21SATB}, voiceleading.ReasonVoiceCount, []int{}},
		{"first species with three voices", three, voiceleading.DistanceNorm{Rule: voiceleading.DistanceOMTFirstSpecies}, voiceleading.ReasonVoiceCount, []int{}},
		{"a custom norm with one voice precedes the count of limits", one, voiceleading.DistanceNorm{Rule: voiceleading.DistanceCustom, Limits: []interval.Interval{p8, p8, p8}}, voiceleading.ReasonVoiceCount, []int{}},
		{"silent parts precede the voice count", silent, voiceleading.DistanceNorm{Rule: voiceleading.DistanceOMTSATB}, voiceleading.ReasonSilentParts, []int{1}},
		{"silent parts with a custom norm", silent, voiceleading.DistanceNorm{Rule: voiceleading.DistanceCustom, Limits: []interval.Interval{p8}}, voiceleading.ReasonSilentParts, []int{1}},
	}
	for _, tt := range inapplicable {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.s.CheckDistances(tt.norm)
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.False(t, got.Applicable)
			assert.Equal(t, tt.reason, got.Reason)
			assert.Equal(t, tt.norm.Rule.String(), got.Rule)
			assert.NotEmpty(t, got.Source)
			assert.NotEmpty(t, got.Scope)
			assert.Empty(t, got.Violations)
			assert.Empty(t, got.Remarks)
			assert.Equal(t, len(tt.silent), len(got.Silent))
			assert.ElementsMatch(t, tt.silent, got.Silent)
		})
	}
	four := sonorityOf(t, "C3", "E4", "G4", "C5")
	invalid := []struct {
		name string
		s    *voiceleading.Sonority
		norm voiceleading.DistanceNorm
	}{
		{"an unset rule", four, voiceleading.DistanceNorm{}},
		{"an unknown rule", four, voiceleading.DistanceNorm{Rule: voiceleading.DistanceRule(99)}},
		{"a custom norm without limits", four, voiceleading.DistanceNorm{Rule: voiceleading.DistanceCustom}},
		{"limits with a built-in norm", four, voiceleading.DistanceNorm{Rule: voiceleading.DistanceOMTSATB, Limits: []interval.Interval{p8}}},
		{"limits with a built-in norm precede inapplicability", three, voiceleading.DistanceNorm{Rule: voiceleading.DistanceMT21SATB, Limits: []interval.Interval{p8, p8, p8}}},
		{"two limits for three pairs", four, voiceleading.DistanceNorm{Rule: voiceleading.DistanceCustom, Limits: []interval.Interval{p8, p8}}},
		{"four limits for three pairs", four, voiceleading.DistanceNorm{Rule: voiceleading.DistanceCustom, Limits: []interval.Interval{p8, p8, p8, p8}}},
	}
	for _, tt := range invalid {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.s.CheckDistances(tt.norm)
			assert.Nil(t, got)
			assertSentinel(t, err, voiceleading.ErrOption)
			assertMustPanics(t, err, func() { tt.s.MustCheckDistances(tt.norm) })
		})
	}
	labels := map[voiceleading.DistanceRule]string{
		voiceleading.DistanceUnset: "unset", voiceleading.DistanceOMTSATB: "omt-satb", voiceleading.DistanceMT21SATB: "mt21-satb",
		voiceleading.DistanceOMTFirstSpecies: "omt-first-species", voiceleading.DistanceCustom: "custom", voiceleading.DistanceRule(77): "unknown",
	}
	for rule, label := range labels {
		assert.Equal(t, label, rule.String())
	}
	// A new assessment is returned on every call.
	norm := voiceleading.DistanceNorm{Rule: voiceleading.DistanceOMTSATB}
	assert.NotSame(t, four.MustCheckDistances(norm), four.MustCheckDistances(norm))
}

func TestPartSonoritySpacingUsesSoundingVoices(t *testing.T) {
	t.Parallel()
	a := analyzeParts(t, voiceleading.LowToHigh, traditionalSystem(),
		line(t, "C3", "C3"), line(t, "E4", "-"), line(t, "G4", "G4"), line(t, "C5", "C5"))
	ps := a.MustSonorities()
	require.Len(t, ps, 2)
	full := ps[0].Voices
	assert.Equal(t, voiceleading.CloseSpacing, full.MustClassifySpacing(voiceleading.SpacingME).Class)
	assert.Equal(t, voiceleading.MixedSpacing, full.MustClassifySpacing(voiceleading.SpacingAM).Class)
	assert.Equal(t, voiceleading.OMTOpenSpacing, full.MustClassifySpacing(voiceleading.SpacingOMT).Class)
	partial := ps[1].Voices
	spacing := partial.MustSpacing()
	require.Len(t, spacing.Adjacent, 2)
	assertVoiceInterval(t, spacing.Adjacent[0], 0, 1, "P12", 19)
	assertVoiceInterval(t, spacing.Adjacent[1], 1, 2, "P4", 5)
	assert.False(t, partial.MustClassifySpacing(voiceleading.SpacingAM).Applicable)
	assert.False(t, partial.MustCheckDistances(voiceleading.DistanceNorm{Rule: voiceleading.DistanceCustom, Limits: []interval.Interval{interval.PerfectOctave()}}).Applicable)
}
