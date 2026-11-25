package voiceleading

import (
	"slices"

	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/spelling"
)

// ToneClass is one tone class of a sonority in the selected reading. Name is
// the written name without an octave: the class itself by spelling, or the
// name of its lowest-numbered voice by sound. Class is the pitch class by
// sound and the zero value by spelling. Voices lists the voice numbers in
// increasing order and belongs to the caller.
type ToneClass struct {
	Name     spelling.Spelling
	Class    pitchclass.Class
	Voices   []int
	Doubling DoublingKind
}

// Doublings describes the tone classes of a sonority in one reading, in order
// of first appearance by voice, and optionally compares them with an expected
// chord composition. HasBass is true for a LowToHigh sonority with voices;
// BassDoubled reports that two or more voices take the class of voice 0.
// Without an expected composition Expected and Complete are false and Missing
// and Extra are empty. All slices belong to the caller.
type Doublings struct {
	Reading     Reading
	Classes     []ToneClass
	HasBass     bool
	BassDoubled bool
	Expected    bool
	Missing     []ToneClass
	Extra       []ToneClass
	Complete    bool
}

// Doubled returns the classes taken by two or more voices, in the order of
// Classes, in a new slice with independent Voices.
func (d Doublings) Doubled() []ToneClass {
	out := []ToneClass{}
	for _, c := range d.Classes {
		if c.Doubling != NoDoubling {
			out = append(out, cloneToneClass(c))
		}
	}
	return out
}

func cloneToneClass(c ToneClass) ToneClass {
	c.Voices = slices.Clone(c.Voices)
	return c
}

// toneKey identifies a class in one reading.
type toneKey struct {
	name  spelling.Spelling
	class pitchclass.Class
}

func (k toneKey) equal(other toneKey, reading Reading) bool {
	if reading == BySound {
		return k.class.Equal(other.class)
	}
	return k.name.Equal(other.name)
}

// Doublings returns the tone classes of the voices in the selected reading,
// their doublings, and, for a set expected chord, the missing and extra
// classes. The expected chord is read only by the names of its notes: octaves,
// corrections and sources are ignored, and repeated classes collapse to their
// first occurrence. By spelling two voices of a class form a unison when their
// written pitches are equal and an octave otherwise; by sound, a unison when
// their positions are equal. No nominal order is required; an unordered part
// sonority declares no bass.
//
// Errors, in this order: ErrUnset; ErrOption for an unknown reading; ErrOption
// in a ChordError{0, -1} for a set but empty expected chord; by sound, the
// PartError of an unmeasured single voice of a part sonority (ErrNoSystem or
// system.ErrInadmissible); by sound with an expected chord, ErrNoSystem for a
// sonority without a system, then system.ErrInadmissible in a ChordError{0, k}
// for an expected name.
func (s *Sonority) Doublings(reading Reading, expected chord.Chord) (Doublings, error) {
	if err := s.check(); err != nil {
		return Doublings{}, err
	}
	if reading != BySpelling && reading != BySound {
		return Doublings{}, ErrOption
	}
	if expected.IsValid() && expected.IsEmpty() {
		return Doublings{}, chordErr(0, -1, ErrOption)
	}
	if reading == BySound && s.soundErr != nil {
		return Doublings{}, s.soundErr
	}
	var wanted []toneKey
	if expected.IsValid() {
		var err error
		if wanted, err = s.expectedKeys(reading, expected); err != nil {
			return Doublings{}, err
		}
	}
	keys := make([]toneKey, len(s.notes))
	for v, n := range s.notes {
		keys[v].name = n.Spelling()
		if reading == BySound {
			c, err := s.sys.PitchClass(s.pitches[v])
			if err != nil {
				return Doublings{}, cause("class by sound", err)
			}
			keys[v].class = c
		}
	}
	out := Doublings{Reading: reading, Classes: []ToneClass{}, Missing: []ToneClass{}, Extra: []ToneClass{}}
	classOf := make([]int, len(s.notes))
	var classKeys []toneKey
	for v := range s.notes {
		i := slices.IndexFunc(classKeys, func(k toneKey) bool { return k.equal(keys[v], reading) })
		if i < 0 {
			i = len(classKeys)
			classKeys = append(classKeys, keys[v])
			out.Classes = append(out.Classes, ToneClass{Name: keys[v].name, Class: keys[v].class, Voices: []int{}})
		}
		classOf[v] = i
		out.Classes[i].Voices = append(out.Classes[i].Voices, v)
	}
	for i := range out.Classes {
		out.Classes[i].Doubling = s.doubling(out.Classes[i].Voices, reading)
	}
	out.HasBass = s.order == LowToHigh && len(s.notes) > 0
	out.BassDoubled = out.HasBass && len(out.Classes[classOf[0]].Voices) >= 2
	if wanted == nil {
		return out, nil
	}
	out.Expected = true
	for _, w := range wanted {
		if !slices.ContainsFunc(classKeys, func(k toneKey) bool { return k.equal(w, reading) }) {
			out.Missing = append(out.Missing, ToneClass{Name: w.name, Class: w.class, Voices: []int{}})
		}
	}
	for i, k := range classKeys {
		if !slices.ContainsFunc(wanted, func(w toneKey) bool { return w.equal(k, reading) }) {
			out.Extra = append(out.Extra, cloneToneClass(out.Classes[i]))
		}
	}
	out.Complete = len(out.Missing) == 0
	return out, nil
}

// MustDoublings is like Doublings but panics with its error.
func (s *Sonority) MustDoublings(reading Reading, expected chord.Chord) Doublings {
	return must(s.Doublings(reading, expected))
}

// expectedKeys returns the distinct classes of an expected chord in the order
// of first occurrence, each named after its first note.
func (s *Sonority) expectedKeys(reading Reading, expected chord.Chord) ([]toneKey, error) {
	if reading == BySound && s.sys == nil {
		return nil, ErrNoSystem
	}
	out := []toneKey{}
	for k, n := range expected.Notes() {
		key := toneKey{name: n.Spelling()}
		if reading == BySound {
			if err := admit(s.sys, key.name); err != nil {
				return nil, chordErr(0, k, err)
			}
			c, err := s.sys.Class(key.name)
			if err != nil {
				return nil, chordErr(0, k, err)
			}
			key.class = c
		}
		if !slices.ContainsFunc(out, func(o toneKey) bool { return o.equal(key, reading) }) {
			out = append(out, key)
		}
	}
	return out, nil
}

func (s *Sonority) doubling(voices []int, reading Reading) DoublingKind {
	unison, octave := false, false
	for a := range voices {
		for b := a + 1; b < len(voices); b++ {
			x, y := s.pitches[voices[a]], s.pitches[voices[b]]
			same := x.Equal(y)
			if reading == BySound {
				same = s.sys.MustCompare(x, y) == 0
			}
			if same {
				unison = true
			} else {
				octave = true
			}
		}
	}
	switch {
	case unison && octave:
		return UnisonAndOctaveDoubling
	case unison:
		return UnisonDoubling
	case octave:
		return OctaveDoubling
	default:
		return NoDoubling
	}
}
