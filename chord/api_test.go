package chord_test

import (
	"fmt"
	"iter"
	"reflect"
	"slices"
	"testing"

	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/note"

	"github.com/stretchr/testify/assert"
)

// T-7, the properties of the API: architecture 5.1 (the complete list of the
// exported identifiers, "no others"), 5.2 (not comparable; receivers), 5.8
// (rules as named function types).

func TestAPINotComparable(t *testing.T) {
	t.Parallel()
	// КР-4: == does not compile; the type is not comparable.
	assert.False(t, reflect.TypeFor[chord.Chord]().Comparable())
	type holder struct{ c chord.Chord }
	assert.False(t, reflect.TypeFor[holder]().Comparable(), "a struct with a Chord field is not comparable")
	assert.Panics(t, func() {
		var a, b any = chord.MustNew(c4), chord.MustNew(c4)
		_ = a == b // compiles and panics at run time (architecture 5.2)
	})
}

func TestAPIStringer(t *testing.T) {
	t.Parallel()
	assert.True(t, reflect.TypeFor[chord.Chord]().Implements(reflect.TypeFor[fmt.Stringer]()), "Chord implements fmt.Stringer by value")
}

// valueMethods are the methods of Chord with a value receiver (architecture
// 5.1): reading, comparison, the string form and their Must variants.
var valueMethods = map[string]any{
	"IsValid":               chord.Chord.IsValid,
	"IsEmpty":               chord.Chord.IsEmpty,
	"Len":                   chord.Chord.Len,
	"Notes":                 chord.Chord.Notes,
	"All":                   chord.Chord.All,
	"Values":                chord.Chord.Values,
	"At":                    chord.Chord.At,
	"MustAt":                chord.Chord.MustAt,
	"SortedNotes":           chord.Chord.SortedNotes,
	"DistinctNotes":         chord.Chord.DistinctNotes,
	"HasRepeats":            chord.Chord.HasRepeats,
	"HasNotesWithOctave":    chord.Chord.HasNotesWithOctave,
	"HasNotesWithoutOctave": chord.Chord.HasNotesWithoutOctave,
	"Contains":              chord.Chord.Contains,
	"Count":                 chord.Chord.Count,
	"Find":                  chord.Chord.Find,
	"Indices":               chord.Chord.Indices,
	"ContainsAll":           chord.Chord.ContainsAll,
	"ContainsAny":           chord.Chord.ContainsAny,
	"ContainsRecord":        chord.Chord.ContainsRecord,
	"ContainsSpelling":      chord.Chord.ContainsSpelling,
	"ContainsFunc":          chord.Chord.ContainsFunc,
	"MustContainsFunc":      chord.Chord.MustContainsFunc,
	"CountFunc":             chord.Chord.CountFunc,
	"MustCountFunc":         chord.Chord.MustCountFunc,
	"IndicesFunc":           chord.Chord.IndicesFunc,
	"MustIndicesFunc":       chord.Chord.MustIndicesFunc,
	"NotesFunc":             chord.Chord.NotesFunc,
	"MustNotesFunc":         chord.Chord.MustNotesFunc,
	"Equal":                 chord.Chord.Equal,
	"EqualInOrder":          chord.Chord.EqualInOrder,
	"SameRecords":           chord.Chord.SameRecords,
	"Includes":              chord.Chord.Includes,
	"HasCommonNotes":        chord.Chord.HasCommonNotes,
	"Compare":               chord.Chord.Compare,
	"String":                chord.Chord.String,
}

// pointerMethods are the changing methods of *Chord and their Must variants
// (architecture 5.1, 5.10).
var pointerMethods = map[string]any{
	"Add":               (*chord.Chord).Add,
	"MustAdd":           (*chord.Chord).MustAdd,
	"AddMissing":        (*chord.Chord).AddMissing,
	"MustAddMissing":    (*chord.Chord).MustAddMissing,
	"Insert":            (*chord.Chord).Insert,
	"MustInsert":        (*chord.Chord).MustInsert,
	"Remove":            (*chord.Chord).Remove,
	"MustRemove":        (*chord.Chord).MustRemove,
	"RemoveAll":         (*chord.Chord).RemoveAll,
	"MustRemoveAll":     (*chord.Chord).MustRemoveAll,
	"RemoveAt":          (*chord.Chord).RemoveAt,
	"MustRemoveAt":      (*chord.Chord).MustRemoveAt,
	"RemoveFunc":        (*chord.Chord).RemoveFunc,
	"MustRemoveFunc":    (*chord.Chord).MustRemoveFunc,
	"KeepFunc":          (*chord.Chord).KeepFunc,
	"MustKeepFunc":      (*chord.Chord).MustKeepFunc,
	"Replace":           (*chord.Chord).Replace,
	"MustReplace":       (*chord.Chord).MustReplace,
	"ReplaceAll":        (*chord.Chord).ReplaceAll,
	"MustReplaceAll":    (*chord.Chord).MustReplaceAll,
	"ReplaceAt":         (*chord.Chord).ReplaceAt,
	"MustReplaceAt":     (*chord.Chord).MustReplaceAt,
	"ReplaceFunc":       (*chord.Chord).ReplaceFunc,
	"MustReplaceFunc":   (*chord.Chord).MustReplaceFunc,
	"ReplaceAtFunc":     (*chord.Chord).ReplaceAtFunc,
	"MustReplaceAtFunc": (*chord.Chord).MustReplaceAtFunc,
	"Clear":             (*chord.Chord).Clear,
	"MustClear":         (*chord.Chord).MustClear,
	"RemoveRepeats":     (*chord.Chord).RemoveRepeats,
	"MustRemoveRepeats": (*chord.Chord).MustRemoveRepeats,
	"Union":             (*chord.Chord).Union,
	"MustUnion":         (*chord.Chord).MustUnion,
	"Intersect":         (*chord.Chord).Intersect,
	"MustIntersect":     (*chord.Chord).MustIntersect,
	"Subtract":          (*chord.Chord).Subtract,
	"MustSubtract":      (*chord.Chord).MustSubtract,
	"Sort":              (*chord.Chord).Sort,
	"MustSort":          (*chord.Chord).MustSort,
	"SortFunc":          (*chord.Chord).SortFunc,
	"MustSortFunc":      (*chord.Chord).MustSortFunc,
	"Reverse":           (*chord.Chord).Reverse,
	"MustReverse":       (*chord.Chord).MustReverse,
	"Shift":             (*chord.Chord).Shift,
	"MustShift":         (*chord.Chord).MustShift,
	"ShiftDown":         (*chord.Chord).ShiftDown,
	"MustShiftDown":     (*chord.Chord).MustShiftDown,
	"Up":                (*chord.Chord).Up,
	"MustUp":            (*chord.Chord).MustUp,
	"Down":              (*chord.Chord).Down,
	"MustDown":          (*chord.Chord).MustDown,
}

// Signatures of architecture 5.1. Each conversion compiles only when the
// signature is exactly the one of the contract.
var signatures = []any{
	(func(...note.Note) (chord.Chord, error))(chord.New),
	(func(...note.Note) chord.Chord)(chord.MustNew),

	(func(chord.Chord) bool)(chord.Chord.IsValid),
	(func(chord.Chord) bool)(chord.Chord.IsEmpty),
	(func(chord.Chord) int)(chord.Chord.Len),
	(func(chord.Chord) []note.Note)(chord.Chord.Notes),
	(func(chord.Chord) iter.Seq2[int, note.Note])(chord.Chord.All),
	(func(chord.Chord) iter.Seq[note.Note])(chord.Chord.Values),
	(func(chord.Chord, int) (note.Note, error))(chord.Chord.At),
	(func(chord.Chord, int) note.Note)(chord.Chord.MustAt),
	(func(chord.Chord) []note.Note)(chord.Chord.SortedNotes),
	(func(chord.Chord) []note.Note)(chord.Chord.DistinctNotes),
	(func(chord.Chord) bool)(chord.Chord.HasRepeats),
	(func(chord.Chord) bool)(chord.Chord.HasNotesWithOctave),
	(func(chord.Chord) bool)(chord.Chord.HasNotesWithoutOctave),

	(func(chord.Chord, note.Note) bool)(chord.Chord.Contains),
	(func(chord.Chord, note.Note) int)(chord.Chord.Count),
	(func(chord.Chord, note.Note) (int, note.Note, bool))(chord.Chord.Find),
	(func(chord.Chord, note.Note) []int)(chord.Chord.Indices),
	(func(chord.Chord, ...note.Note) bool)(chord.Chord.ContainsAll),
	(func(chord.Chord, ...note.Note) bool)(chord.Chord.ContainsAny),
	(func(chord.Chord, note.Note) bool)(chord.Chord.ContainsRecord),
	(func(chord.Chord, note.Note) bool)(chord.Chord.ContainsSpelling),
	(func(chord.Chord, chord.Condition) (bool, error))(chord.Chord.ContainsFunc),
	(func(chord.Chord, chord.Condition) bool)(chord.Chord.MustContainsFunc),
	(func(chord.Chord, chord.Condition) (int, error))(chord.Chord.CountFunc),
	(func(chord.Chord, chord.Condition) int)(chord.Chord.MustCountFunc),
	(func(chord.Chord, chord.Condition) ([]int, error))(chord.Chord.IndicesFunc),
	(func(chord.Chord, chord.Condition) []int)(chord.Chord.MustIndicesFunc),
	(func(chord.Chord, chord.Condition) ([]note.Note, error))(chord.Chord.NotesFunc),
	(func(chord.Chord, chord.Condition) []note.Note)(chord.Chord.MustNotesFunc),

	(func(chord.Chord, chord.Chord) bool)(chord.Chord.Equal),
	(func(chord.Chord, chord.Chord) bool)(chord.Chord.EqualInOrder),
	(func(chord.Chord, chord.Chord) bool)(chord.Chord.SameRecords),
	(func(chord.Chord, chord.Chord) bool)(chord.Chord.Includes),
	(func(chord.Chord, chord.Chord) bool)(chord.Chord.HasCommonNotes),
	(func(chord.Chord, chord.Chord) int)(chord.Chord.Compare),

	(func(*chord.Chord, ...note.Note) error)((*chord.Chord).Add),
	(func(*chord.Chord, ...note.Note) *chord.Chord)((*chord.Chord).MustAdd),
	(func(*chord.Chord, ...note.Note) error)((*chord.Chord).AddMissing),
	(func(*chord.Chord, ...note.Note) *chord.Chord)((*chord.Chord).MustAddMissing),
	(func(*chord.Chord, int, ...note.Note) error)((*chord.Chord).Insert),
	(func(*chord.Chord, int, ...note.Note) *chord.Chord)((*chord.Chord).MustInsert),
	(func(*chord.Chord, note.Note) error)((*chord.Chord).Remove),
	(func(*chord.Chord, note.Note) *chord.Chord)((*chord.Chord).MustRemove),
	(func(*chord.Chord, note.Note) error)((*chord.Chord).RemoveAll),
	(func(*chord.Chord, note.Note) *chord.Chord)((*chord.Chord).MustRemoveAll),
	(func(*chord.Chord, int) error)((*chord.Chord).RemoveAt),
	(func(*chord.Chord, int) *chord.Chord)((*chord.Chord).MustRemoveAt),
	(func(*chord.Chord, chord.Condition) error)((*chord.Chord).RemoveFunc),
	(func(*chord.Chord, chord.Condition) *chord.Chord)((*chord.Chord).MustRemoveFunc),
	(func(*chord.Chord, chord.Condition) error)((*chord.Chord).KeepFunc),
	(func(*chord.Chord, chord.Condition) *chord.Chord)((*chord.Chord).MustKeepFunc),
	(func(*chord.Chord, note.Note, note.Note) error)((*chord.Chord).Replace),
	(func(*chord.Chord, note.Note, note.Note) *chord.Chord)((*chord.Chord).MustReplace),
	(func(*chord.Chord, note.Note, note.Note) error)((*chord.Chord).ReplaceAll),
	(func(*chord.Chord, note.Note, note.Note) *chord.Chord)((*chord.Chord).MustReplaceAll),
	(func(*chord.Chord, int, note.Note) error)((*chord.Chord).ReplaceAt),
	(func(*chord.Chord, int, note.Note) *chord.Chord)((*chord.Chord).MustReplaceAt),
	(func(*chord.Chord, chord.Transform) error)((*chord.Chord).ReplaceFunc),
	(func(*chord.Chord, chord.Transform) *chord.Chord)((*chord.Chord).MustReplaceFunc),
	(func(*chord.Chord, int, chord.Transform) error)((*chord.Chord).ReplaceAtFunc),
	(func(*chord.Chord, int, chord.Transform) *chord.Chord)((*chord.Chord).MustReplaceAtFunc),
	(func(*chord.Chord) error)((*chord.Chord).Clear),
	(func(*chord.Chord) *chord.Chord)((*chord.Chord).MustClear),
	(func(*chord.Chord) error)((*chord.Chord).RemoveRepeats),
	(func(*chord.Chord) *chord.Chord)((*chord.Chord).MustRemoveRepeats),
	(func(*chord.Chord, chord.Chord) error)((*chord.Chord).Union),
	(func(*chord.Chord, chord.Chord) *chord.Chord)((*chord.Chord).MustUnion),
	(func(*chord.Chord, chord.Chord) error)((*chord.Chord).Intersect),
	(func(*chord.Chord, chord.Chord) *chord.Chord)((*chord.Chord).MustIntersect),
	(func(*chord.Chord, chord.Chord) error)((*chord.Chord).Subtract),
	(func(*chord.Chord, chord.Chord) *chord.Chord)((*chord.Chord).MustSubtract),
	(func(*chord.Chord) error)((*chord.Chord).Sort),
	(func(*chord.Chord) *chord.Chord)((*chord.Chord).MustSort),
	(func(*chord.Chord, chord.Comparison) error)((*chord.Chord).SortFunc),
	(func(*chord.Chord, chord.Comparison) *chord.Chord)((*chord.Chord).MustSortFunc),
	(func(*chord.Chord) error)((*chord.Chord).Reverse),
	(func(*chord.Chord) *chord.Chord)((*chord.Chord).MustReverse),
	(func(*chord.Chord, int) error)((*chord.Chord).Shift),
	(func(*chord.Chord, int) *chord.Chord)((*chord.Chord).MustShift),
	(func(*chord.Chord, int) error)((*chord.Chord).ShiftDown),
	(func(*chord.Chord, int) *chord.Chord)((*chord.Chord).MustShiftDown),
	(func(*chord.Chord) error)((*chord.Chord).Up),
	(func(*chord.Chord) *chord.Chord)((*chord.Chord).MustUp),
	(func(*chord.Chord) error)((*chord.Chord).Down),
	(func(*chord.Chord) *chord.Chord)((*chord.Chord).MustDown),

	(func(chord.Chord) string)(chord.Chord.String),
}

func TestAPIMethodSets(t *testing.T) {
	t.Parallel()
	names := func(typ reflect.Type) []string {
		var out []string
		for i := range typ.NumMethod() {
			out = append(out, typ.Method(i).Name)
		}
		slices.Sort(out)
		return out
	}
	sortedKeys := func(ms ...map[string]any) []string {
		var out []string
		for _, m := range ms {
			for k := range m {
				out = append(out, k)
			}
		}
		slices.Sort(out)
		return out
	}

	// RS-7: 36 methods of Chord, 86 of *Chord. Changing methods are not in
	// the method set of Chord.
	assert.Len(t, valueMethods, 36)
	assert.Len(t, pointerMethods, 50)
	assert.Equal(t, sortedKeys(valueMethods), names(reflect.TypeFor[chord.Chord]()), "the methods of Chord")
	assert.Equal(t, sortedKeys(valueMethods, pointerMethods), names(reflect.TypeFor[*chord.Chord]()), "the methods of *Chord")

	// The method expressions have the types of the methods.
	for name, f := range valueMethods {
		m, ok := reflect.TypeFor[chord.Chord]().MethodByName(name)
		if assert.True(t, ok, name) {
			assert.Equal(t, m.Type, reflect.TypeOf(f), name)
		}
	}
	for name, f := range pointerMethods {
		m, ok := reflect.TypeFor[*chord.Chord]().MethodByName(name)
		if assert.True(t, ok, name) {
			assert.Equal(t, m.Type, reflect.TypeOf(f), name)
		}
	}
	assert.Len(t, signatures, 2+36+50)
	for _, s := range signatures {
		assert.NotNil(t, s)
	}
}

func TestAPIRuleTypes(t *testing.T) {
	t.Parallel()
	// Architecture 5.8: a function literal and a value of an unnamed
	// function type with the same signature are passed without conversion.
	var cond chord.Condition = func(n note.Note) (bool, error) { return n.HasOctave(), nil }
	var tr chord.Transform = func(n note.Note) (note.Note, error) { return n, nil }
	var cmp chord.Comparison = func(a, b note.Note) (int, error) { return a.Compare(b), nil }
	unnamed := func(n note.Note) (bool, error) { return !n.HasOctave(), nil }
	c := chord.MustNew(c4, eNo)
	assert.Equal(t, 1, c.MustCountFunc(cond))
	assert.Equal(t, 1, c.MustCountFunc(unnamed))
	assert.NotNil(t, tr)
	assert.NotNil(t, cmp)

	for _, typ := range []reflect.Type{reflect.TypeFor[chord.Condition](), reflect.TypeFor[chord.Transform](), reflect.TypeFor[chord.Comparison]()} {
		assert.Equal(t, reflect.Func, typ.Kind(), typ.String())
	}
}

func TestAPIErrors(t *testing.T) {
	t.Parallel()
	for _, err := range []error{chord.ErrUnset, chord.ErrNotFound, chord.ErrIndex, chord.ErrRuleFailed} {
		assert.Error(t, err)
	}
}
