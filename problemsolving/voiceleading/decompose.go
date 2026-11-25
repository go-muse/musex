package voiceleading

import (
	"slices"

	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/voice"
)

// Simultaneity is one sonority of a sequence with its duration. An empty
// chord is a rest in every voice.
type Simultaneity struct {
	Notes    chord.Chord
	Duration voice.Duration
}

// Ambiguity marks voices of sonority Index that have equal positions (a
// unison or an enharmonic equivalent): which voice takes which note does not
// follow from the notes. Voices belongs to the caller.
type Ambiguity struct {
	Index  int
	Voices []int
}

// Decomposition holds the lines of a decomposed sequence from the bass upward
// and the places where voice membership is ambiguous. Both slices belong to
// the caller; the lines are new.
type Decomposition struct {
	Lines       []*voice.Line
	Ambiguities []Ambiguity
}

// Decompose splits a sequence of sonorities into lines from the bass upward:
// line k takes the k-th lowest note of every sonority with its duration, and
// an empty sonority gives a rest in every line. The notes of a sonority are
// ordered by position, then by written pitch (written octave, then letter
// distance up from the boundary of the system), then by note.Note.Compare,
// then by record order. The decomposition assumes that voices do not cross;
// crossed voices are not restored. Every sonority attacks anew; equal adjacent
// notes are not tied. The lines are ready for Analyze with LowToHigh.
//
// With at least two voices a system is needed: a non-nil sys, which must admit
// every note, or a system chosen from all notes of all sonorities in order
// (bindings, then settings read once per note, then equality, then
// admissibility).
//
// Errors, in this order: ErrEmpty for an empty sequence; ErrUnset with
// system.ErrInvalid for an unset explicit system; for each sonority, ErrUnset
// with chord.ErrUnset in a ChordError{i, -1}, ErrNoOctave in a
// ChordError{i, k}, and a duration failure of package voice in a
// ChordError{i, -1}; ErrNoVoices when every sonority is empty; ErrVoiceCount
// in a ChordError{i, -1} for the first sonority with another number of notes;
// with two or more voices ErrNoSystem, a settings failure,
// ErrDifferentSystems, system.ErrInadmissible and a written-octave failure in
// a ChordError{i, k}; a failure to build a line.
func Decompose(seq []Simultaneity, sys *system.System) (Decomposition, error) {
	if len(seq) == 0 {
		return Decomposition{}, ErrEmpty
	}
	if sys != nil && !sys.IsValid() {
		return Decomposition{}, unset("system", system.ErrInvalid)
	}
	chords, n, err := checkSequence(seq)
	if err != nil {
		return Decomposition{}, err
	}
	out := Decomposition{Ambiguities: []Ambiguity{}}
	if n >= 2 {
		if sys == nil {
			if sys, err = chooseChordSystem(chords, 0); err != nil {
				return Decomposition{}, err
			}
		}
		if err := admitNotes(sys, chords, 0); err != nil {
			return Decomposition{}, err
		}
		for i, ns := range chords {
			for k, x := range ns {
				if _, err := sys.WrittenOctave(x.MustPitch()); err != nil {
					return Decomposition{}, chordErr(i, k, err)
				}
			}
		}
		for i, ns := range chords {
			sortVoices(sys, ns)
			out.Ambiguities = append(out.Ambiguities, ambiguities(sys, i, ns)...)
		}
	}
	if out.Lines, err = buildLines(seq, chords, n); err != nil {
		return Decomposition{}, err
	}
	return out, nil
}

// MustDecompose is like Decompose but panics with its error.
func MustDecompose(seq []Simultaneity, sys *system.System) Decomposition {
	return must(Decompose(seq, sys))
}

// DecomposeInOrder splits a sequence of sonorities given by voice from the
// bass upward: line k takes the k-th note in record order of every sonority.
// Crossings are kept; no system is needed and no source is read. Durations,
// rests, attacks and errors are as in Decompose, without the system errors.
func DecomposeInOrder(seq []Simultaneity) ([]*voice.Line, error) {
	if len(seq) == 0 {
		return nil, ErrEmpty
	}
	chords, n, err := checkSequence(seq)
	if err != nil {
		return nil, err
	}
	return buildLines(seq, chords, n)
}

// MustDecomposeInOrder is like DecomposeInOrder but panics with its error.
func MustDecomposeInOrder(seq []Simultaneity) []*voice.Line { return must(DecomposeInOrder(seq)) }

// checkSequence validates every sonority and returns copies of their notes
// and the number of voices.
func checkSequence(seq []Simultaneity) ([][]note.Note, int, error) {
	chords := make([][]note.Note, len(seq))
	for i, s := range seq {
		if !s.Notes.IsValid() {
			return nil, 0, chordErr(i, -1, unset("chord", chord.ErrUnset))
		}
		chords[i] = s.Notes.Notes()
		for k, x := range chords[i] {
			if !x.HasOctave() {
				return nil, 0, chordErr(i, k, ErrNoOctave)
			}
		}
		var err error
		if len(chords[i]) == 0 {
			_, err = voice.NewRest(s.Duration)
		} else {
			_, err = voice.NewSound(chords[i][0], s.Duration)
		}
		if err != nil {
			return nil, 0, chordErr(i, -1, err)
		}
	}
	n := 0
	for _, ns := range chords {
		if len(ns) > 0 {
			n = len(ns)
			break
		}
	}
	if n == 0 {
		return nil, 0, ErrNoVoices
	}
	for i, ns := range chords {
		if len(ns) > 0 && len(ns) != n {
			return nil, 0, chordErr(i, -1, ErrVoiceCount)
		}
	}
	return chords, n, nil
}

// sortVoices orders admitted notes from the bass upward, stably.
func sortVoices(sys *system.System, ns []note.Note) {
	boundary := sys.Boundary()
	slices.SortStableFunc(ns, func(x, y note.Note) int {
		px, py := x.MustPitch(), y.MustPitch()
		if c := sys.MustCompare(px, py); c != 0 {
			return c
		}
		if c := sys.MustWrittenOctave(px).Compare(sys.MustWrittenOctave(py)); c != 0 {
			return c
		}
		if c := boundary.DistanceUp(x.Letter()) - boundary.DistanceUp(y.Letter()); c != 0 {
			return c
		}
		return x.Compare(y)
	})
}

// ambiguities returns the groups of two or more sorted voices with equal
// positions.
func ambiguities(sys *system.System, i int, ns []note.Note) []Ambiguity {
	var out []Ambiguity
	for start := 0; start < len(ns); {
		end := start + 1
		for end < len(ns) && sys.MustCompare(ns[start].MustPitch(), ns[end].MustPitch()) == 0 {
			end++
		}
		if end-start >= 2 {
			voices := make([]int, 0, end-start)
			for v := start; v < end; v++ {
				voices = append(voices, v)
			}
			out = append(out, Ambiguity{Index: i, Voices: voices})
		}
		start = end
	}
	return out
}

// buildLines builds n lines from validated sonorities.
func buildLines(seq []Simultaneity, chords [][]note.Note, n int) ([]*voice.Line, error) {
	lines := make([]*voice.Line, n)
	for k := range n {
		elems := make([]voice.Element, len(seq))
		for i, s := range seq {
			var err error
			if len(chords[i]) == 0 {
				elems[i], err = voice.NewRest(s.Duration)
			} else {
				elems[i], err = voice.NewSound(chords[i][k], s.Duration)
			}
			if err != nil {
				return nil, cause("line element", err)
			}
		}
		line, err := voice.New(elems...)
		if err != nil {
			return nil, cause("line", err)
		}
		lines[k] = line
	}
	return lines, nil
}
