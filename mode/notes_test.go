package mode_test

import (
	"math"
	"strings"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func noteAt(t *testing.T, name string, aspn int, delta cents.Cents, src settings.Source) note.Note {
	t.Helper()
	n, err := note.FromASPN(spelling.MustParse(name), aspn, delta, src)
	require.NoError(t, err)
	return n
}

func TestNotesFromNamesAndOwnership(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, structure, first, want string }{
		{"dorian from D", dorian, "D", "D E F G A B C"},
		{"dorian from C", dorian, "C", "C D E-1 F G A B-1"},
		{"harmonic minor", "M2 m2 M2 M2 m2 A2 m2", "G+1", "G+1 A+1 B C+1 D+1 E F+2"},
		{"seven flats", ionian, "C-1", "C-1 D-1 E-1 F-1 G-1 A-1 B-1"},
		{"pentatonic minor", "m3 M2 M2 m3 M2", "C", "C E-1 F G B-1"},
		{"blues", "m3 M2 A1 m2 m3 M2", "C", "C E-1 F F+1 G B-1"},
		{"whole tone", "M2 M2 M2 M2 M2 d3", "C", "C D E F+1 G+1 A+1"},
		{"bebop major", "M2 M2 m2 M2 A1 m2 M2 m2", "C", "C D E F G G+1 A B"},
		{"two steps", "A4 d5", "C", "C F+1"},
		{"one step", "P8", "D", "D"},
		{"open", "m3 M2 |", "E", "E G A"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			m := mode.MustParse(tt.structure)
			delta := cents.MustFromInt(30)
			first := notes(t, tt.first, delta, nil)[0]
			result, err := m.NotesFrom(first)
			require.NoError(t, err)
			assert.Equal(t, tt.want, noteNames(result))
			for _, n := range result {
				assert.False(t, n.HasOctave())
				assert.Nil(t, n.Source())
				assert.True(t, delta.Equal(n.Cents()))
			}
			ds, err := m.DegreesFrom(first)
			require.NoError(t, err)
			require.Len(t, ds, len(result))
			for i, d := range ds {
				assert.True(t, d.IsValid())
				assert.Equal(t, i+1, d.Number())
				assert.True(t, d.HasNote())
				assert.True(t, result[i].Spelling().Equal(d.Note().Spelling()))
			}
			again, err := m.NoteFrom(first, 1, 0)
			require.NoError(t, err)
			assert.True(t, first.Spelling().Equal(again.Spelling()))
			if cycle, ok := m.Cycle(); ok {
				rebuilt, err := mode.ClosedFromNotes(result, cycle)
				require.NoError(t, err)
				assert.True(t, m.Equal(rebuilt))
			} else {
				assert.True(t, m.Equal(mode.MustOpenFromNotes(result)))
			}
			result[0].MustSetCents(cents.MustFromInt(40))
			ds[0] = mode.MustParse("P8").MustDegree(1)
			assert.True(t, delta.Equal(first.Cents()))
			assert.True(t, delta.Equal(m.MustNotesFrom(first)[0].Cents()))
			assert.False(t, m.MustDegree(1).HasNote())
		})
	}
}

func TestReconstructModeFromNotes(t *testing.T) {
	t.Parallel()
	// Under architecture 3.2 (P-5/F-2), reconstruction restores the original
	// mode when spellings preserve the intervals between consecutive notes.
	// For closed modes, the caller also passes the original cycle.
	for _, tt := range []struct {
		name, original, cycle, spellings, want string
		octaves                                []int
		wantErr                                error
	}{
		{"closed compound edge", "M9 m7", "P15", "C D", "M2 m14", []int{4, 5}, nil},
		{"open compound edge", "M9 |", "", "C D", "M2 |", []int{4, 5}, nil},
		{"repeated spelling", "P8 P8", "P15", "C C", "", []int{4, 5}, mode.ErrInvalidStructure},
		{"compound cycle with simple edge", "M2 m14", "P15", "C D", "M2 m14", []int{4, 4}, nil},
		{"singleton with explicit cycle", "P4", "P4", "C", "P4", []int{4}, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			original := mode.MustParse(tt.original)
			first := noteAt(t, "C", 4, cents.Cents{}, settings.Standard())
			produced, err := original.NotesFrom(first)
			require.NoError(t, err)
			require.Len(t, produced, len(tt.octaves))
			assert.Equal(t, tt.spellings, noteNames(produced))
			for i, n := range produced {
				require.True(t, n.HasOctave(), "note %d", i+1)
				assert.Equal(t, tt.octaves[i], n.MustASPN(), "note %d", i+1)
			}

			var rebuilt mode.Mode
			if tt.cycle == "" {
				rebuilt, err = mode.OpenFromNotes(produced)
			} else {
				cycle := interval.MustParse(tt.cycle)
				originalCycle, ok := original.Cycle()
				require.True(t, ok)
				require.True(t, cycle.Equal(originalCycle))
				rebuilt, err = mode.ClosedFromNotes(produced, cycle)
			}
			if tt.wantErr != nil {
				checkError(t, err, tt.wantErr)
				assert.Zero(t, rebuilt)
				return
			}
			require.NoError(t, err)
			checkMode(t, rebuilt, tt.want)
			assert.Equal(t, tt.original == tt.want, original.Equal(rebuilt))
		})
	}
}

func TestBoundAndExplicitNoteSystems(t *testing.T) {
	t.Parallel()
	m := mode.MustFromCode("dorian")
	src := settings.Standard()
	delta := cents.MustFromInt(30)
	first := noteAt(t, "D", 4, delta, src)
	got, err := m.NotesFrom(first)
	require.NoError(t, err)
	require.Len(t, got, 7)
	assert.Equal(t, "D E F G A B C", noteNames(got))
	for i, n := range got {
		assert.True(t, n.HasOctave())
		assert.Same(t, src, n.Source())
		assert.True(t, delta.Equal(n.Cents()))
		wantOctave := 4
		if i == 6 {
			wantOctave = 5
		}
		assert.Equal(t, wantOctave, n.MustASPN())
	}
	assert.True(t, first.Spelling().Equal(spelling.MustParse("D")))
	assert.Equal(t, 4, first.MustASPN())
	assert.Equal(t, 5, m.MustNoteFrom(first, 2, 1).MustASPN())
	assert.Equal(t, "E", m.MustNoteFrom(first, 2, 1).Spelling().String())
	assert.Equal(t, 3, m.MustNoteFrom(first, 1, -1).MustASPN())

	unattached := noteAt(t, "D", 4, delta, nil)
	noNotes, err := m.NotesFrom(unattached)
	checkError(t, err, mode.ErrNoSystem, interval.ErrNoSystem)
	assert.Nil(t, noNotes)
	one, err := m.NoteFrom(unattached, 1, 0)
	checkError(t, err, mode.ErrNoSystem, interval.ErrNoSystem)
	assert.Zero(t, one)

	sys := system.Traditional()
	got, err = m.NotesFromIn(sys, unattached)
	require.NoError(t, err)
	assert.Equal(t, "D E F G A B C", noteNames(got))
	ds, err := m.DegreesFromIn(sys, unattached)
	require.NoError(t, err)
	require.Len(t, ds, 7)
	assert.Equal(t, 5, ds[6].Note().MustASPN())
	assert.Nil(t, ds[6].Note().Source())
	out, err := m.NoteFromIn(sys, unattached, 7, 0)
	require.NoError(t, err)
	assert.Equal(t, 5, out.MustASPN())
	assert.Nil(t, out.Source())

	restricted := system.MustNewRegular(2, 1, letter.C, system.MustIntegerRange(-1, 1), "Single accidentals")
	bound := noteAt(t, "G+1", 4, delta, src.MustWithSystem(restricted))
	harmonic := mode.MustFromCode("harmonic-minor")
	got, err = harmonic.NotesFrom(bound)
	checkError(t, err, mode.ErrCannotBuildNote, note.ErrInadmissible, system.ErrInadmissible)
	assert.Nil(t, got)
	ds, err = harmonic.DegreesFrom(bound)
	checkError(t, err, mode.ErrCannotBuildNote, note.ErrInadmissible, system.ErrInadmissible)
	assert.Nil(t, ds)
	got, err = harmonic.NotesFromIn(restricted, noteAt(t, "G+1", 4, delta, nil))
	checkError(t, err, mode.ErrCannotBuildNote, system.ErrInadmissible)
	assert.Nil(t, got)
	got, err = m.NotesFromIn(nil, first)
	checkError(t, err, mode.ErrUnset)
	assert.Nil(t, got)
	out, err = m.NoteFromIn(&system.System{}, first, 0, 0)
	checkError(t, err, mode.ErrUnset)
	assert.Zero(t, out)
	out, err = mode.MustParse("m3 M2 |").NoteFromIn(sys, first, 1, -1)
	checkError(t, err, mode.ErrInvalidArgument)
	assert.Zero(t, out)
}

func TestNoteFromArbitraryCyclesAndExactFinalNote(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		structure, first, want string
		number, cycle, octave  int
	}{
		{"M2 M2 m2", "G", "C", 1, 1, 5},
		{"M2 M2 m2", "G", "E", 3, 1, 5},
		{"M2 M2 m2", "G", "A", 3, 2, 5},
		{"M2 M2 m2", "G", "B-1", 1, 3, 5},
		{"M9 m7", "C", "D", 2, 0, 5},
		{"M9 m7", "C", "C", 1, 1, 6},
		{"m3 M2 |", "E", "A", 3, 0, 4},
	} {
		got, err := mode.MustParse(tt.structure).NoteFromIn(system.Traditional(), noteAt(t, tt.first, 4, cents.Cents{}, nil), tt.number, tt.cycle)
		require.NoError(t, err)
		assert.True(t, spelling.MustParse(tt.want).Equal(got.Spelling()))
		assert.Equal(t, tt.octave, got.MustASPN())
	}

	// Two copies of this chromatic unison exceed Interval's deviation limit.
	// Starting at -M leaves the final alteration at +M, which is admissible.
	magnitude := alteration.MaxMagnitude
	first, err := note.FromSpelling(spelling.MustNew(letter.C, alteration.MustFromInt(-magnitude)), cents.Cents{}, nil)
	require.NoError(t, err)
	cycle := interval.MustFromDeviation(1, alteration.MustFromInt(magnitude))
	m := mode.MustClosedFromIntervals([]interval.Interval{cycle})
	got, err := m.NoteFrom(first, 1, 2)
	require.NoError(t, err)
	assert.Equal(t, letter.C, got.Letter())
	assert.True(t, alteration.MustFromInt(magnitude).Equal(got.Alteration()))
	assert.False(t, got.HasOctave())
	got, err = m.NoteFrom(first, 1, 3)
	checkError(t, err, mode.ErrCannotBuildNote, alteration.ErrOutOfRange)
	assert.Zero(t, got)

	// 7*MinInt letter steps cannot be an Interval but the final octave is -1.
	octaveMode := mode.MustParse("P8")
	high := noteAt(t, "C", math.MaxInt, cents.Cents{}, nil)
	got, err = octaveMode.NoteFromIn(system.Traditional(), high, 1, math.MinInt)
	require.NoError(t, err)
	assert.Equal(t, -1, got.MustASPN())
	assert.Equal(t, letter.C, got.Letter())
	got, err = octaveMode.NoteFromIn(system.Traditional(), high, 1, 1)
	checkError(t, err, mode.ErrCannotBuildNote)
	assert.Zero(t, got)
}

func TestKeySignatureFromNames(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		code, first, want string
		index             int
	}{
		{"mixolydian", "F+1", "F+1 C+1 G+1 D+1 A+1", 5},
		{"dorian", "D", "", 0},
		{"aeolian", "C", "B-1 E-1 A-1", -3},
		{"ionian", "C-1", "B-1 E-1 A-1 D-1 G-1 C-1 F-1", -7},
		{"ionian", "G+1", "F+2 C+1 G+1 D+1 A+1 E+1 B+1", 8},
	} {
		t.Run(tt.code+"_"+tt.first, func(t *testing.T) {
			m := mode.MustFromCode(tt.code)
			sign, ok, err := m.KeySignatureFrom(spelling.MustParse(tt.first))
			require.NoError(t, err)
			require.True(t, ok)
			assert.Equal(t, tt.index, sign.Index)
			require.NotNil(t, sign.Signs)
			expected := strings.Fields(tt.want)
			require.Len(t, sign.Signs, len(expected))
			for i, n := range sign.Signs {
				assert.True(t, spelling.MustParse(expected[i]).Equal(n), "sign %d: %s", i, n)
			}
			if len(sign.Signs) > 0 {
				sign.Signs[0] = spelling.Spelling{}
			}
			fresh, again := m.MustKeySignatureFrom(spelling.MustParse(tt.first))
			assert.True(t, again)
			if len(expected) > 0 {
				assert.True(t, spelling.MustParse(expected[0]).Equal(fresh.Signs[0]))
			}
		})
	}
	sig, ok, err := mode.MustParse("M2 M2 m2 M2 M2 M2 |").KeySignatureFrom(spelling.MustParse("C"))
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Zero(t, sig.Index)
	assert.NotNil(t, sig.Signs)

	for _, tt := range []struct {
		m     mode.Mode
		start spelling.Spelling
	}{
		{mode.MustFromCode("harmonic-minor"), spelling.MustParse("A")},
		{mode.MustFromCode("pentatonic-major"), spelling.MustParse("C")},
		{mode.MustFromCode("ionian"), spelling.MustNew(letter.C, alteration.MustFromFraction(1, 2))},
	} {
		sign, ok, err := tt.m.KeySignatureFrom(tt.start)
		require.NoError(t, err)
		assert.False(t, ok)
		assert.Zero(t, sign)
	}
	sign, ok, err := mode.MustFromCode("ionian").KeySignatureFrom(spelling.Spelling{})
	checkError(t, err, mode.ErrUnset)
	assert.False(t, ok)
	assert.Zero(t, sign)
	sharpest := spelling.MustNew(letter.C, alteration.MustFromInt(alteration.MaxMagnitude))
	sign, ok, err = mode.MustFromCode("lydian").KeySignatureFrom(sharpest)
	checkError(t, err, mode.ErrCannotBuildNote, alteration.ErrOutOfRange)
	assert.False(t, ok)
	assert.Zero(t, sign)
}

func TestFirstSpellingForSignature(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		code  string
		index int
		want  string
	}{
		{"ionian", 3, "A"}, {"dorian", 3, "B"}, {"aeolian", 3, "F+1"}, {"lydian", 3, "D"}, {"locrian", 3, "G+1"},
		{"ionian", -4, "A-1"}, {"aeolian", -4, "F"}, {"phrygian", -4, "C"},
		{"dorian", 0, "D"}, {"lydian", 0, "F"}, {"locrian", 0, "B"}, {"mixolydian", 5, "F+1"},
		{"ionian", 8, "G+1"},
	} {
		m := mode.MustFromCode(tt.code)
		got, err := m.FirstSpellingForSignature(tt.index)
		require.NoError(t, err)
		assert.True(t, spelling.MustParse(tt.want).Equal(got), "%s, index %d: %s", tt.code, tt.index, got)
		result, ok := m.MustKeySignatureFrom(got)
		require.True(t, ok)
		assert.Equal(t, tt.index, result.Index)
	}
	for _, m := range []mode.Mode{mode.MustFromCode("harmonic-minor"), mode.MustParse("M2 M2 m2 M2 M2 M2 |")} {
		got, err := m.FirstSpellingForSignature(0)
		checkError(t, err, mode.ErrInvalidArgument)
		assert.Zero(t, got)
	}
	got, err := mode.MustFromCode("locrian").FirstSpellingForSignature(math.MaxInt)
	checkError(t, err, mode.ErrOutOfRange)
	assert.Zero(t, got)
}
