package note_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The type Note (architecture 6.1, KR-1, KR-2; P-10, P-12; R-B).

// changingMethods lists the changing methods of architecture 6.6.2 and 6.7.
var changingMethods = []string{
	"Respell", "RespellWritten", "RespellWrittenIn",
	"SetOctave", "SetASPN", "SetWrittenOctave", "SetWrittenASPN", "SetWrittenNumber",
	"SetWrittenOctaveIn", "SetWrittenNumberIn", "RemoveOctave",
	"ReplaceWithPitch", "ReplaceWithSpelling", "SetCents",
	"Raise", "Lower", "RaiseBy", "LowerBy", "RaiseByAlteration", "LowerByAlteration", "SetAlteration",
	"RaiseWritten", "LowerWritten", "RaiseWrittenBy", "LowerWrittenBy",
	"RaiseWrittenByAlteration", "LowerWrittenByAlteration", "SetAlterationWritten",
	"RaiseWrittenByAlterationIn", "SetAlterationWrittenIn",
	"Naturalize", "NaturalizeWritten", "NaturalizeWrittenIn",
	"Shift", "ShiftDown", "Up", "Down", "Move", "MoveIn",
	"SetSource", "SetSourceWritten", "Unbind",
}

// readingMethods lists methods that read the note or give new notes
// (architecture 6.5, 6.7, 6.8, 6.9, 6.10, 6.12).
var readingMethods = []string{
	"IsValid", "HasOctave", "Spelling", "Letter", "Alteration", "Cents", "Pitch", "Octave", "ASPN",
	"MustPitch", "MustOctave", "MustASPN", "IsNatural", "IsInteger", "Direction", "Source", "IsBound", "String",
	"CheckSetSource", "CheckSetSourceWritten", "MustCheckSetSource", "MustCheckSetSourceWritten",
	"Settings", "MustSettings",
	"IsAdmissible", "IsAdmissibleIn", "Class", "ClassIn", "ClassInTuning",
	"WrittenOctave", "WrittenOctaveIn", "WrittenASPN", "WrittenNumber", "WrittenNumberIn", "WrittenName", "WrittenNameIn",
	"Step", "StepIn", "StepInTuning", "CentsFromReference", "CentsFromReferenceIn", "CentsFromReferenceInTuning",
	"Frequency", "FrequencyIn", "FrequencyInTuning", "PitchFrequency", "PitchFrequencyIn", "PitchFrequencyInTuning",
	"IsExpressible", "IsExpressibleIn", "Format", "FormatIn", "FormatName", "FormatNameIn",
	"Enharmonics", "EnharmonicsIn", "Range", "Walk", "WalkIn",
	"Equal", "SameRecord", "SameSpelling", "SameLetter", "SameAlteration", "SameOctave", "SameCents", "SameSource",
	"Compare", "OctaveSub", "OctaveDistance",
	"Equivalent", "EquivalentIn", "EquivalentInTuning", "ComparePosition", "ComparePositionIn", "ComparePositionInTuning",
	"Between", "BetweenIn", "BetweenInTuning", "Distance", "DistanceIn", "DistanceInTuning",
	"CompareSound", "CompareSoundIn", "CompareSoundInTuning", "CompareSoundBy", "CompareSoundByIn", "CompareSoundByInTuning",
	"SoundsSame", "SoundsSameIn", "SoundsSameInTuning", "SoundsSameBy", "SoundsSameByIn", "SoundsSameByInTuning",
	"CentsTo", "CentsToIn", "CentsToInTuning",
	"Offset", "OffsetBy", "OffsetIn", "OffsetByIn", "OffsetInTuning", "OffsetByInTuning",
}

// TestMethodSets checks KR-2: the value type has no changing method, the
// pointer type has all of them; reading methods are in both sets; every
// reading method that reports an error has a Must variant.
func TestMethodSets(t *testing.T) {
	t.Parallel()

	value := reflect.TypeFor[note.Note]()
	pointer := reflect.TypeFor[*note.Note]()
	for _, name := range changingMethods {
		for _, m := range []string{name, "Must" + name} {
			_, onValue := value.MethodByName(m)
			_, onPointer := pointer.MethodByName(m)
			assert.False(t, onValue, "%s must not be a method of Note", m)
			assert.True(t, onPointer, "%s must be a method of *Note", m)
		}
		pm, ok := pointer.MethodByName(name)
		if assert.True(t, ok, name) {
			require.Equal(t, 1, pm.Type.NumOut(), "%s returns only an error", name)
			assert.Equal(t, reflect.TypeFor[error](), pm.Type.Out(0), name)
		}
		mm, ok := pointer.MethodByName("Must" + name)
		if assert.True(t, ok, "Must"+name) {
			require.Equal(t, 1, mm.Type.NumOut())
			assert.Equal(t, pointer, mm.Type.Out(0), "Must%s returns *Note", name)
		}
	}
	errType := reflect.TypeFor[error]()
	for _, name := range readingMethods {
		vm, onValue := value.MethodByName(name)
		_, onPointer := pointer.MethodByName(name)
		assert.True(t, onValue, "%s must be a method of Note", name)
		assert.True(t, onPointer, "%s must be a method of *Note", name)
		if !onValue || strings.HasPrefix(name, "Must") {
			continue
		}
		out := vm.Type.NumOut()
		if out > 0 && vm.Type.Out(out-1) == errType {
			_, hasMust := value.MethodByName("Must" + name)
			assert.True(t, hasMust, "Must%s", name)
		}
	}
}

// TestComparable checks that Note is comparable and == is not the equality by
// record: it sees the source (P-12, R-B; risk R-N-4).
func TestComparable(t *testing.T) {
	t.Parallel()

	assert.True(t, reflect.TypeFor[note.Note]().Comparable())

	a := mk(t, spelling.C(), 4, ct(0), settings.Standard())
	b := mk(t, spelling.C(), 4, ct(0), settings.Standard())
	assert.False(t, a == b, "two separate standard sets differ by ==")
	assert.True(t, a.Equal(b))

	c := a
	assert.True(t, a == c, "a copy is == while nothing changes")

	m := map[note.Note]int{a: 1}
	m[c]++
	assert.Equal(t, 2, m[a])
}

// TestZeroValue checks that Note{} is the unset note (architecture 6.1).
func TestZeroValue(t *testing.T) {
	t.Parallel()

	var n note.Note
	assert.False(t, n.IsValid())
	assert.False(t, n.HasOctave())
	assert.False(t, n.IsBound())
	assert.True(t, n.Cents().Equal(ct(0)))
	assert.True(t, n == note.Note{})
	_, err := n.Frequency()
	assertErr(t, err, note.ErrUnset)
}

// TestNilPointerReads checks that a reading method through a nil *Note panics
// by the rules of Go, before the method is entered (architecture 6.1).
func TestNilPointerReads(t *testing.T) {
	t.Parallel()

	var p *note.Note
	assert.NotNil(t, panicValue(func() { _ = p.IsValid() }))
	assert.NotNil(t, panicValue(func() { _, _ = p.Frequency() }))
	assert.NotNil(t, panicValue(func() { _ = p.String() }))
}

// TestStringer checks that Note implements fmt.Stringer by value.
func TestStringer(t *testing.T) {
	t.Parallel()

	var _ fmt.Stringer = note.Note{}
	n := mk(t, spelling.C(), 4, ct(0), nil)
	assert.Equal(t, n.String(), fmt.Sprint(n))
}
