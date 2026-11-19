package scalechoice_test

import (
	"fmt"
	"slices"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/problemsolving/scalechoice"
	"github.com/go-muse/musex/search"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/template"
	"github.com/go-muse/musex/tuning"
	"github.com/go-muse/musex/tuning/temperament"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func names(text string) []spelling.Spelling {
	out := make([]spelling.Spelling, 0)
	for _, token := range strings.Fields(text) {
		out = append(out, spelling.MustParse(token))
	}
	return out
}
func notes(text string) []note.Note {
	out := make([]note.Note, 0)
	for _, name := range names(text) {
		out = append(out, note.MustFromSpelling(name, cents.Zero(), nil))
	}
	return out
}
func pitched(text string, aspn int, source settings.Source) []note.Note {
	out := make([]note.Note, 0)
	for _, name := range names(text) {
		out = append(out, note.MustFromPitch(pitch.MustFromASPN(name, aspn), cents.Zero(), source))
	}
	return out
}
func opts() scalechoice.Options {
	return scalechoice.Options{Condition: search.Containment, Comparison: scalechoice.BySpelling}
}
func full() scalechoice.Options { o := opts(); o.Condition = search.FullMatch; return o }
func positionOptions() scalechoice.Options {
	o := full()
	o.Comparison = scalechoice.ByPosition
	o.System = system.Traditional()
	return o
}
func nameText(ns []spelling.Spelling) string {
	words := make([]string, len(ns))
	for index, name := range ns {
		words[index] = name.String()
	}
	return strings.Join(words, " ")
}
func notesText(ns []note.Note) string {
	out := make([]spelling.Spelling, len(ns))
	for index, n := range ns {
		out[index] = n.Spelling()
	}
	return nameText(out)
}
func toneText(ts []scalechoice.Tone) string {
	words := make([]string, len(ts))
	for index, tone := range ts {
		words[index] = nameText(tone.Names)
	}
	return strings.Join(words, " ")
}
func classes(ts []scalechoice.Tone) []int {
	out := make([]int, 0, len(ts))
	for _, tone := range ts {
		out = append(out, tone.Class.Number())
	}
	return out
}
func readingKeys(rs []scalechoice.Reading) []string {
	out := make([]string, 0, len(rs))
	for _, r := range rs {
		out = append(out, r.Entry().Code()+"/"+r.Tonic().String())
	}
	return out
}
func allReadingKeys(a scalechoice.Analysis) []string {
	out := make([]string, 0)
	for _, group := range a.Groups() {
		out = append(out, readingKeys(group.Readings)...)
	}
	return out
}
func catalog(codes ...string) template.Catalog {
	built := template.BuiltIn()
	es := make([]template.Entry, len(codes))
	for index, code := range codes {
		e, ok := built.ByCode(code)
		if !ok {
			panic(code)
		}
		es[index] = e
	}
	return template.MustNewCatalog(es...)
}
func customCatalog(code string, places []interval.Interval, cycle interval.Interval) template.Catalog {
	tmpl := template.MustClosedFromPlaceIntervals(places, cycle)
	return template.MustNewCatalog(template.MustNewEntry(code, []string{code, "alias for " + code}, tmpl))
}
func oneToneCatalog() template.Catalog {
	return customCatalog("one", []interval.Interval{interval.PerfectUnison()}, interval.PerfectOctave())
}
func irregular() *system.System {
	return system.MustNew(system.Definition{Name: "irregular", Classes: 12, Boundary: letter.C, Offsets: [letter.Count]int{0, 3, 4, 5, 7, 9, 11}, Chromatic: 1, Alterations: system.Integers()})
}
func boundaryA() *system.System {
	return system.MustNew(system.Definition{Name: "boundary A", Classes: 12, Boundary: letter.A, Offsets: [letter.Count]int{3, 5, 7, 8, 10, 0, 2}, Chromatic: 1, Alterations: system.Integers()})
}
func halfFlatE() spelling.Spelling {
	return spelling.MustNew(letter.E, alteration.MustFromFraction(-1, 2))
}

type countedSource struct {
	current atomic.Pointer[settings.Settings]
	reads   atomic.Int64
}

func source(sys *system.System) *countedSource {
	s := &countedSource{}
	// Replacing only the system would retain the standard 12-class
	// temperament. Build a matching N-EDO tuning for every requested system.
	t := tuning.MustNew(sys, temperament.MustEDO(sys.Classes()),
		tuning.RootFromName(spelling.C()), tuning.A4At440())
	s.current.Store(settings.Standard().MustWithTuning(t))
	return s
}
func (s *countedSource) Settings() *settings.Settings { s.reads.Add(1); return s.current.Load() }
func (s *countedSource) reset()                       { s.reads.Store(0) }

func checkNotes(t *testing.T, want, got []note.Note) {
	t.Helper()
	require.Len(t, got, len(want))
	for index, n := range got {
		assert.True(t, n.Equal(want[index]), "place %d: %v versus %v", index, n, want[index])
		assert.True(t, n.Cents().Equal(want[index].Cents()))
		assert.False(t, n.IsBound(), "snapshots have no live source")
	}
}
func checkMotion(t *testing.T, m scalechoice.Motion, named string, steps int, sys *system.System) {
	t.Helper()
	require.NoError(t, m.NamedError)
	require.NoError(t, m.ChromaticError)
	assert.True(t, interval.MustParse(named).Equal(m.Named), "got %v want %s", m.Named, named)
	assert.True(t, m.Chromatic.IsValid())
	require.NotNil(t, m.Chromatic.System())
	assert.True(t, sys.Equal(m.Chromatic.System()))
	assert.Equal(t, steps, m.Chromatic.Steps())
}
func checkMotionError(t *testing.T, m scalechoice.Motion, want error) {
	t.Helper()
	assert.ErrorIs(t, m.NamedError, want)
	assert.ErrorIs(t, m.ChromaticError, want)
	assert.False(t, m.Named.IsValid())
	assert.False(t, m.Chromatic.IsValid())
}
func mustPanic(t *testing.T, want error, call func()) {
	t.Helper()
	defer func() {
		value := recover()
		err, ok := value.(error)
		require.True(t, ok, "panic should be an error, got %T", value)
		assert.ErrorIs(t, err, want)
	}()
	call()
}
func checkPartition(t *testing.T, p scalechoice.Partition, kept, dropped []int) {
	t.Helper()
	assert.Equal(t, kept, slices.Collect(slices.Values(p.Kept)))
	assert.Equal(t, dropped, slices.Collect(slices.Values(p.Dropped)))
}
func profileCodes(c template.Catalog) []string {
	out := make([]string, 0)
	for _, e := range c.Entries() {
		out = append(out, e.Code())
	}
	return out
}
func describe(a scalechoice.Analysis) string {
	var groups []string
	for _, group := range a.Groups() {
		groups = append(groups, fmt.Sprint(toneText(group.Tones), readingKeys(group.Readings)))
	}
	return fmt.Sprint(notesText(a.Notes()), groups, toneText(a.Common()), allReadingKeys(a))
}
