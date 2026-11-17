package mode_test

import (
	"errors"
	"iter"
	"reflect"
	"slices"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/degree"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func capturePanic(call func()) (recovered any) {
	defer func() { recovered = recover() }()
	call()
	return nil
}

type mustCase struct {
	name           string
	normal, must   any
	valid, invalid []any
	category       error
}

func callAPI(call any, args []any) []reflect.Value {
	values := make([]reflect.Value, len(args))
	for i, arg := range args {
		values[i] = reflect.ValueOf(arg)
	}
	return reflect.ValueOf(call).Call(values)
}

func compareResult(t *testing.T, want, got any) {
	t.Helper()
	switch w := want.(type) {
	case mode.Mode:
		assert.Equal(t, observe(w), observe(got.(mode.Mode)))
	case mode.Repository:
		wc, wok := w.Catalog()
		gc, gok := got.(mode.Repository).Catalog()
		assert.Equal(t, wok, gok)
		assert.True(t, wc.Equal(gc))
	case iter.Seq[mode.Mode]:
		all := slices.Collect(w)
		must := slices.Collect(got.(iter.Seq[mode.Mode]))
		require.Len(t, must, len(all))
		for i, m := range all {
			assert.Equal(t, observe(m), observe(must[i]))
		}
	case iter.Seq[mode.IntervalAt]:
		assert.Equal(t, intervalRows(slices.Collect(w)), intervalRows(slices.Collect(got.(iter.Seq[mode.IntervalAt]))))
	case degree.Degree:
		assert.True(t, w.Equal(got.(degree.Degree)))
	case []degree.Degree:
		ds := got.([]degree.Degree)
		assert.Equal(t, w == nil, ds == nil)
		require.Len(t, ds, len(w))
		for i, d := range w {
			assert.True(t, d.Equal(ds[i]))
		}
	case []template.Entry:
		es := got.([]template.Entry)
		assert.Equal(t, w == nil, es == nil)
		require.Len(t, es, len(w))
		for i, e := range w {
			assert.True(t, e.Equal(es[i]))
		}
	case note.Note:
		actual := got.(note.Note)
		assert.True(t, w.Equal(actual))
		assert.Equal(t, w.Source(), actual.Source())
	case []note.Note:
		ns := got.([]note.Note)
		assert.Equal(t, w == nil, ns == nil)
		require.Len(t, ns, len(w))
		for i, n := range w {
			assert.True(t, n.Equal(ns[i]))
			assert.Equal(t, n.Source(), ns[i].Source())
		}
	case []mode.Relation:
		actual := got.([]mode.Relation)
		assert.Equal(t, w == nil, actual == nil)
		require.Len(t, actual, len(w))
		for i, r := range w {
			assert.True(t, r.Entry.Equal(actual[i].Entry))
			assert.Equal(t, r.Degree, actual[i].Degree)
			assert.Equal(t, r.Degrees, actual[i].Degrees)
		}
	case []mode.Rotation:
		assert.Equal(t, rotationMatches(w), rotationMatches(got.([]mode.Rotation)))
	default:
		assert.Equal(t, want, got)
	}
}

func checkMustSignature(t *testing.T, normal, must any, mutator bool) {
	t.Helper()
	n, m := reflect.TypeOf(normal), reflect.TypeOf(must)
	require.Equal(t, n.NumIn(), m.NumIn())
	for i := range n.NumIn() {
		assert.Equal(t, n.In(i), m.In(i))
	}
	require.Positive(t, n.NumOut())
	assert.Equal(t, reflect.TypeFor[error](), n.Out(n.NumOut()-1))
	if mutator {
		require.Equal(t, 1, m.NumOut())
		assert.Equal(t, reflect.TypeFor[*mode.Mode](), m.Out(0))
	} else {
		require.Equal(t, n.NumOut()-1, m.NumOut())
		for i := range m.NumOut() {
			assert.Equal(t, n.Out(i), m.Out(i))
		}
	}
}

func checkMustFailure(t *testing.T, tt mustCase) {
	t.Helper()
	results := callAPI(tt.normal, tt.invalid)
	for _, v := range results[:len(results)-1] {
		assert.True(t, v.IsZero(), "failure returned a nonzero %s: %v", v.Type(), v)
	}
	base, ok := results[len(results)-1].Interface().(error)
	require.True(t, ok, "normal variant did not return an error")
	checkError(t, base, tt.category)
	panicked := capturePanic(func() { callAPI(tt.must, tt.invalid) })
	failure, ok := panicked.(error)
	require.True(t, ok, "Must panicked with %T: %v", panicked, panicked)
	checkError(t, failure, tt.category)
	// The independently executed wrapper must preserve the original error
	// chain, not merely panic with the mode sentinel for the same category.
	for _, cause := range []error{
		template.ErrUnset, template.ErrInvalidArgument, template.ErrInvalidInterval,
		template.ErrEmptyInput, template.ErrParse, template.ErrOpenRotation,
		template.ErrIncomparable, template.ErrOutOfRange, interval.ErrParse,
		interval.ErrUnset, interval.ErrOutOfRange,
	} {
		assert.Equal(t, errors.Is(base, cause), errors.Is(failure, cause), "chain membership for %v", cause)
	}
}

func TestEveryMustFunctionAndReader(t *testing.T) {
	t.Parallel()
	zero := mode.Mode{}
	uninit := mode.Repository{}
	m := mode.MustFromCode("ionian")
	other := mode.MustFromCode("dorian")
	e := entry("custom", ionian, "Test mode")
	c := template.MustNewCatalog(e)
	r := mode.MustNewRepository(&c)
	sys := system.Traditional()
	first := noteAt(t, "C", 4, cents.Cents{}, nil)
	noOctave := notes(t, "C", cents.Cents{}, nil)[0]
	cName := spelling.MustParse("C")
	intervalM2 := interval.MustParse("M2")
	cycle := interval.MustParse("P8")
	own := intervals("P1 M2")
	ps := []system.Step{2, 2, 1, 2, 2, 2, 1}
	plan := mode.LetterPlan{Degrees: []int{1, 2}, Cycle: 3}
	ns := []note.Note{first, noteAt(t, "D", 4, cents.Cents{}, nil)}
	for _, tt := range []mustCase{
		{"FromTemplate", mode.FromTemplate, mode.MustFromTemplate, []any{e.Template()}, []any{template.Template{}}, mode.ErrUnset},
		{"FromEntry", mode.FromEntry, mode.MustFromEntry, []any{e}, []any{template.Entry{}}, mode.ErrUnset},
		{"FromCode", mode.FromCode, mode.MustFromCode, []any{"ionian"}, []any{"no such code"}, mode.ErrNoEntry},
		{"FromCodeIn", mode.FromCodeIn, mode.MustFromCodeIn, []any{&c, "custom"}, []any{(*template.Catalog)(nil), "ionian"}, mode.ErrUnset},
		{"FromName", mode.FromName, mode.MustFromName, []any{"Dorian"}, []any{"blues scale"}, mode.ErrAmbiguousName},
		{"FromNameIn", mode.FromNameIn, mode.MustFromNameIn, []any{&c, "TEST MODE"}, []any{(*template.Catalog)(nil), "Dorian"}, mode.ErrUnset},
		{"ClosedFromIntervals", mode.ClosedFromIntervals, mode.MustClosedFromIntervals, []any{intervals("P8")}, []any{[]interval.Interval(nil)}, mode.ErrEmptyInput},
		{"OpenFromIntervals", mode.OpenFromIntervals, mode.MustOpenFromIntervals, []any{intervals("M2")}, []any{intervals("-M2")}, mode.ErrInvalidStructure},
		{"ClosedFromDegreeIntervals", mode.ClosedFromDegreeIntervals, mode.MustClosedFromDegreeIntervals, []any{own, cycle}, []any{own, interval.Interval{}}, mode.ErrUnset},
		{"OpenFromDegreeIntervals", mode.OpenFromDegreeIntervals, mode.MustOpenFromDegreeIntervals, []any{own}, []any{intervals("P1")}, mode.ErrEmptyInput},
		{"ClosedFromNotes", mode.ClosedFromNotes, mode.MustClosedFromNotes, []any{ns, cycle}, []any{ns, interval.Interval{}}, mode.ErrUnset},
		{"FromNotesInOctave", mode.FromNotesInOctave, mode.MustFromNotesInOctave, []any{ns}, []any{[]note.Note(nil)}, mode.ErrEmptyInput},
		{"OpenFromNotes", mode.OpenFromNotes, mode.MustOpenFromNotes, []any{ns}, []any{ns[:1]}, mode.ErrEmptyInput},
		{"FromNumbers", mode.FromNumbers, mode.MustFromNumbers, []any{sys, ps}, []any{sys, ps[:1]}, mode.ErrWrongDegreeCount},
		{"FromNumbersFrom", mode.FromNumbersFrom, mode.MustFromNumbersFrom, []any{sys, ps, cName}, []any{sys, ps, spelling.Spelling{}}, mode.ErrUnset},
		{"FromNumbersWithPlan", mode.FromNumbersWithPlan, mode.MustFromNumbersWithPlan, []any{sys, []system.Step{2, 2}, plan}, []any{sys, ps, plan}, mode.ErrWrongDegreeCount},
		{"FromNumbersFromWithPlan", mode.FromNumbersFromWithPlan, mode.MustFromNumbersFromWithPlan, []any{sys, []system.Step{2, 2}, cName, plan}, []any{sys, ps, cName, mode.LetterPlan{Degrees: []int{2, 3}, Cycle: 4}}, mode.ErrInvalidArgument},
		{"Parse", mode.Parse, mode.MustParse, []any{"m3 M2 |"}, []any{"M2  P8"}, mode.ErrParse},
		{"FromDegree", mode.FromDegree, mode.MustFromDegree, []any{m, 2}, []any{m, 8}, mode.ErrInvalidArgument},
		{"Rotations", mode.Rotations, mode.MustRotations, []any{m}, []any{mode.MustParse("M2 |")}, mode.ErrOpenMode},
		{"Degree", mode.Mode.Degree, mode.Mode.MustDegree, []any{m, 7}, []any{zero, 7}, mode.ErrUnset},
		{"Degrees", mode.Mode.Degrees, mode.Mode.MustDegrees, []any{m}, []any{zero}, mode.ErrUnset},
		{"DegreeInterval", mode.Mode.DegreeInterval, mode.Mode.MustDegreeInterval, []any{m, 8}, []any{zero, 8}, mode.ErrUnset},
		{"AdjacentIntervalAfter", mode.Mode.AdjacentIntervalAfter, mode.Mode.MustAdjacentIntervalAfter, []any{m, 8}, []any{zero, 8}, mode.ErrUnset},
		{"BetweenDegrees", mode.Mode.BetweenDegrees, mode.Mode.MustBetweenDegrees, []any{m, 7, 2}, []any{zero, 7, 2}, mode.ErrUnset},
		{"Segment", mode.Mode.Segment, mode.Mode.MustSegment, []any{m, 7, 9}, []any{zero, 7, 9}, mode.ErrUnset},
		{"Next", mode.Mode.Next, mode.Mode.MustNext, []any{m, 7}, []any{zero, 7}, mode.ErrUnset},
		{"Previous", mode.Mode.Previous, mode.Mode.MustPrevious, []any{m, 1}, []any{zero, 1}, mode.ErrUnset},
		{"Shift", mode.Mode.Shift, mode.Mode.MustShift, []any{m, 1, -1}, []any{zero, 1, -1}, mode.ErrUnset},
		{"DistanceUp", mode.Mode.DistanceUp, mode.Mode.MustDistanceUp, []any{m, 3, 6}, []any{zero, 3, 6}, mode.ErrUnset},
		{"DistanceDown", mode.Mode.DistanceDown, mode.Mode.MustDistanceDown, []any{m, 3, 6}, []any{zero, 3, 6}, mode.ErrUnset},
		{"ShortestDistance", mode.Mode.ShortestDistance, mode.Mode.MustShortestDistance, []any{m, 6, 3}, []any{zero, 6, 3}, mode.ErrUnset},
		{"NotesFrom", mode.Mode.NotesFrom, mode.Mode.MustNotesFrom, []any{m, noOctave}, []any{zero, first}, mode.ErrUnset},
		{"NotesFromIn", mode.Mode.NotesFromIn, mode.Mode.MustNotesFromIn, []any{m, sys, first}, []any{zero, sys, first}, mode.ErrUnset},
		{"DegreesFrom", mode.Mode.DegreesFrom, mode.Mode.MustDegreesFrom, []any{m, noOctave}, []any{zero, first}, mode.ErrUnset},
		{"DegreesFromIn", mode.Mode.DegreesFromIn, mode.Mode.MustDegreesFromIn, []any{m, sys, first}, []any{zero, sys, first}, mode.ErrUnset},
		{"NoteFrom", mode.Mode.NoteFrom, mode.Mode.MustNoteFrom, []any{m, noOctave, 2, -1}, []any{zero, first, 2, -1}, mode.ErrUnset},
		{"NoteFromIn", mode.Mode.NoteFromIn, mode.Mode.MustNoteFromIn, []any{m, sys, first, 2, -1}, []any{zero, sys, first, 2, -1}, mode.ErrUnset},
		{"KeySignatureFrom", mode.Mode.KeySignatureFrom, mode.Mode.MustKeySignatureFrom, []any{m, cName}, []any{zero, cName}, mode.ErrUnset},
		{"FirstSpellingForSignature", mode.Mode.FirstSpellingForSignature, mode.Mode.MustFirstSpellingForSignature, []any{m, 3}, []any{zero, 3}, mode.ErrUnset},
		{"Difference", mode.Mode.Difference, mode.Mode.MustDifference, []any{m, other}, []any{zero, other}, mode.ErrUnset},
		{"MismatchingDegrees", mode.Mode.MismatchingDegrees, mode.Mode.MustMismatchingDegrees, []any{m, other, mode.Adjacent}, []any{zero, other, mode.Adjacent}, mode.ErrUnset},
		{"SameInSystem", mode.Mode.SameInSystem, mode.Mode.MustSameInSystem, []any{m, sys, other}, []any{zero, sys, other}, mode.ErrUnset},
		{"SameInSystemFrom", mode.Mode.SameInSystemFrom, mode.Mode.MustSameInSystemFrom, []any{m, sys, cName, other, cName}, []any{zero, sys, cName, other, cName}, mode.ErrUnset},
		{"Inclination", mode.Mode.Inclination, mode.Mode.MustInclination, []any{m}, []any{zero}, mode.ErrUnset},
		{"IsDiatonic", mode.Mode.IsDiatonic, mode.Mode.MustIsDiatonic, []any{m}, []any{zero}, mode.ErrUnset},
		{"OnlyDiatonicIntervals", mode.Mode.OnlyDiatonicIntervals, mode.Mode.MustOnlyDiatonicIntervals, []any{m}, []any{zero}, mode.ErrUnset},
		{"Brightness", mode.Mode.Brightness, mode.Mode.MustBrightness, []any{m}, []any{zero}, mode.ErrUnset},
		{"CompareBrightness", mode.Mode.CompareBrightness, mode.Mode.MustCompareBrightness, []any{m, other}, []any{zero, other}, mode.ErrUnset},
		{"Intervals", mode.Mode.Intervals, mode.Mode.MustIntervals, []any{m}, []any{zero}, mode.ErrUnset},
		{"DegreesForInterval", mode.Mode.DegreesForInterval, mode.Mode.MustDegreesForInterval, []any{m, intervalM2}, []any{zero, intervalM2}, mode.ErrUnset},
		{"ChromaticIntervals", mode.Mode.ChromaticIntervals, mode.Mode.MustChromaticIntervals, []any{mode.MustFromCode("harmonic-minor")}, []any{zero}, mode.ErrUnset},
		{"Tetrachords", mode.Mode.Tetrachords, mode.Mode.MustTetrachords, []any{m}, []any{zero}, mode.ErrUnset},
		{"Period", mode.Mode.Period, mode.Mode.MustPeriod, []any{m}, []any{zero}, mode.ErrUnset},
		{"DistinctRotations", mode.Mode.DistinctRotations, mode.Mode.MustDistinctRotations, []any{m}, []any{zero}, mode.ErrUnset},
		{"MinimalPart", mode.Mode.MinimalPart, mode.Mode.MustMinimalPart, []any{m}, []any{zero}, mode.ErrUnset},
		{"Numbers", mode.Mode.Numbers, mode.Mode.MustNumbers, []any{m, sys}, []any{zero, sys}, mode.ErrUnset},
		{"NumbersFrom", mode.Mode.NumbersFrom, mode.Mode.MustNumbersFrom, []any{m, sys, cName}, []any{zero, sys, cName}, mode.ErrUnset},
		{"CycleIn", mode.Mode.CycleIn, mode.Mode.MustCycleIn, []any{m, sys}, []any{zero, sys}, mode.ErrUnset},
		{"CycleInFrom", mode.Mode.CycleInFrom, mode.Mode.MustCycleInFrom, []any{m, sys, cName}, []any{zero, sys, cName}, mode.ErrUnset},
		{"SymmetryIn", mode.Mode.SymmetryIn, mode.Mode.MustSymmetryIn, []any{m, sys}, []any{zero, sys}, mode.ErrUnset},
		{"SymmetryInFrom", mode.Mode.SymmetryInFrom, mode.Mode.MustSymmetryInFrom, []any{m, sys, cName}, []any{zero, sys, cName}, mode.ErrUnset},
		{"NewRepository", mode.NewRepository, mode.MustNewRepository, []any{&c}, []any{(*template.Catalog)(nil)}, mode.ErrUnset},
		{"Identify", mode.Repository.Identify, mode.Repository.MustIdentify, []any{r, m}, []any{uninit, m}, mode.ErrUnset},
		{"IdentifyRotations", mode.Repository.IdentifyRotations, mode.Repository.MustIdentifyRotations, []any{r, other}, []any{uninit, other}, mode.ErrUnset},
		{"Related", mode.Repository.Related, mode.Repository.MustRelated, []any{r, other, 1, mode.FromFirst, true}, []any{uninit, other, 1, mode.FromFirst, true}, mode.ErrUnset},
		{"IdentifyInSystem", mode.Repository.IdentifyInSystem, mode.Repository.MustIdentifyInSystem, []any{r, m, sys}, []any{uninit, m, sys}, mode.ErrUnset},
		{"IdentifyInSystemFrom", mode.Repository.IdentifyInSystemFrom, mode.Repository.MustIdentifyInSystemFrom, []any{r, m, sys, cName}, []any{uninit, m, sys, cName}, mode.ErrUnset},
		{"IdentifyRotationsInSystem", mode.Repository.IdentifyRotationsInSystem, mode.Repository.MustIdentifyRotationsInSystem, []any{r, other, sys}, []any{uninit, other, sys}, mode.ErrUnset},
		{"IdentifyRotationsInSystemFrom", mode.Repository.IdentifyRotationsInSystemFrom, mode.Repository.MustIdentifyRotationsInSystemFrom, []any{r, other, sys, cName}, []any{uninit, other, sys, cName}, mode.ErrUnset},
		{"IntervalsString", mode.Mode.IntervalsString, mode.Mode.MustIntervalsString, []any{m}, []any{zero}, mode.ErrUnset},
		{"RussianIntervals", mode.Mode.RussianIntervals, mode.Mode.MustRussianIntervals, []any{m}, []any{zero}, mode.ErrUnset},
	} {
		t.Run(tt.name, func(t *testing.T) {
			checkMustSignature(t, tt.normal, tt.must, false)
			want := callAPI(tt.normal, tt.valid)
			require.True(t, want[len(want)-1].IsNil(), "valid input did not succeed: %v", want[len(want)-1])
			got := callAPI(tt.must, tt.valid)
			for i, v := range got {
				compareResult(t, want[i].Interface(), v.Interface())
			}
			checkMustFailure(t, tt)
		})
	}
}

func TestEveryMustMutator(t *testing.T) {
	t.Parallel()
	e := entry("custom", dorian, "Test Dorian")
	c := template.MustNewCatalog(e)
	for _, tt := range []struct {
		name         string
		normal, must any
		args         []any
	}{
		{"Rebuild", (*mode.Mode).Rebuild, (*mode.Mode).MustRebuild, []any{e.Template()}},
		{"RebuildFromEntry", (*mode.Mode).RebuildFromEntry, (*mode.Mode).MustRebuildFromEntry, []any{e}},
		{"RebuildFromCode", (*mode.Mode).RebuildFromCode, (*mode.Mode).MustRebuildFromCode, []any{"dorian"}},
		{"RebuildFromCodeIn", (*mode.Mode).RebuildFromCodeIn, (*mode.Mode).MustRebuildFromCodeIn, []any{&c, "custom"}},
		{"RebuildFromName", (*mode.Mode).RebuildFromName, (*mode.Mode).MustRebuildFromName, []any{"Dorian"}},
		{"RebuildFromNameIn", (*mode.Mode).RebuildFromNameIn, (*mode.Mode).MustRebuildFromNameIn, []any{&c, "TEST DORIAN"}},
		{"Rotate", (*mode.Mode).Rotate, (*mode.Mode).MustRotate, []any{2}},
		{"SetDegreeInterval", (*mode.Mode).SetDegreeInterval, (*mode.Mode).MustSetDegreeInterval, []any{7, interval.MustParse("m7")}},
		{"AddDegreeInterval", (*mode.Mode).AddDegreeInterval, (*mode.Mode).MustAddDegreeInterval, []any{interval.MustParse("A1")}},
		{"RemoveDegree", (*mode.Mode).RemoveDegree, (*mode.Mode).MustRemoveDegree, []any{4}},
		{"Mirror", (*mode.Mode).Mirror, (*mode.Mode).MustMirror, nil},
		{"Repeat", (*mode.Mode).Repeat, (*mode.Mode).MustRepeat, []any{2}},
		{"ReduceToMinimalPart", (*mode.Mode).ReduceToMinimalPart, (*mode.Mode).MustReduceToMinimalPart, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			checkMustSignature(t, tt.normal, tt.must, true)
			var zero mode.Mode
			for _, receiver := range []*mode.Mode{nil, &zero} {
				args := append([]any{receiver}, tt.args...)
				checkMustFailure(t, mustCase{normal: tt.normal, must: tt.must, invalid: args, category: mode.ErrUnset})
				assert.False(t, zero.IsValid())
			}
			var candidate mode.Mode
			var invalid []any
			category := mode.ErrInvalidArgument
			switch tt.name {
			case "Rebuild", "RebuildFromEntry", "RebuildFromCodeIn", "RebuildFromNameIn":
				category = mode.ErrUnset
			case "RebuildFromCode":
				category, invalid = mode.ErrNoEntry, []any{"no such code"}
			case "RebuildFromName":
				category, invalid = mode.ErrAmbiguousName, []any{"blues scale"}
			case "Rotate", "RemoveDegree", "Repeat":
				invalid = []any{0}
			case "SetDegreeInterval":
				invalid = []any{1, interval.MustParse("M2")}
			case "AddDegreeInterval":
				invalid = []any{interval.MustParse("M2")}
			case "Mirror":
				candidate = mode.MustOpenFromIntervals(extremeSeconds()[:3])
				category = mode.ErrOutOfRange
			case "ReduceToMinimalPart":
				candidate = mode.MustParse("M2 |")
				category = mode.ErrOpenMode
			}
			if !candidate.IsValid() {
				candidate = mode.MustFromName("Ionian")
			}
			if category == mode.ErrUnset {
				invalid = slices.Clone(tt.args)
				invalid[0] = reflect.Zero(reflect.TypeOf(invalid[0])).Interface()
			}
			before := observe(candidate)
			checkMustFailure(t, mustCase{normal: tt.normal, must: tt.must, invalid: append([]any{&candidate}, invalid...), category: category})
			assert.Equal(t, before, observe(candidate))

			want := mode.MustFromName("Ionian")
			got := want
			require.True(t, callAPI(tt.normal, append([]any{&want}, tt.args...))[0].IsNil())
			actual := callAPI(tt.must, append([]any{&got}, tt.args...))
			assert.Same(t, &got, actual[0].Interface())
			assert.Equal(t, observe(want), observe(got))
		})
	}
}
