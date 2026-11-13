package settings_test

import (
	"sync"
	"testing"

	"github.com/go-muse/musex/notation"
	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/octave/number"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// snapshot collects the observable results of all operations on a shared
// pair of sets. Two snapshots of the same sets are equal.
type snapshot struct {
	valid                    bool
	str                      string
	equal                    bool
	tuningEqual, catalogEq   bool
	notationEq, numberingEq  bool
	systemEqual              bool
	ctxEqual                 bool
	sourceSame, sameSource   bool
	expresses, expressesP    bool
	format, formatPitch      string
	parsed                   spelling.Spelling
	parsedPitch              pitch.Pitch
	withTuning, withSystem   bool
	withNotation, withNumber bool
	withCatalog, withDesign  bool
	replacement, structure   settings.Report
	designations, systems    settings.Report
	changes                  settings.Changes
}

// take runs every operation of the package on prev and next.
func take(prev, next *settings.Settings, ref *settings.Settings) snapshot {
	var s snapshot
	s.valid = prev.IsValid()
	s.str = prev.String()
	s.equal = prev.Equal(next)
	s.tuningEqual = prev.Tuning().Equal(ref.Tuning())
	s.catalogEq = prev.Catalog().Equal(ref.Catalog())
	s.notationEq = prev.Notation().Equal(ref.Notation())
	s.numberingEq = prev.Numbering().Equal(ref.Numbering())
	s.systemEqual = prev.System().Equal(ref.System())
	ctx := prev.OctaveContext()
	s.ctxEqual = ctx.Numbering.Equal(ref.Numbering()) && ctx.Catalog.Equal(ref.Catalog())
	var src settings.Source = prev
	s.sourceSame = src.Settings() == prev
	s.sameSource = settings.SameSource(prev, next)
	s.expresses = prev.MustExpresses(spelling.CSharp())
	s.expressesP = prev.MustExpressesPitch(at(spelling.BSharp(), 4))
	s.format = prev.MustFormat(spelling.CSharp())
	s.formatPitch = prev.MustFormatPitch(at(spelling.BSharp(), 4))
	s.parsed = prev.MustParse(s.format)
	s.parsedPitch = prev.MustParsePitch(s.formatPitch)
	s.withTuning = prev.MustWithTuning(next.Tuning()).Tuning().Equal(next.Tuning())
	s.withSystem = prev.MustWithSystem(system.Traditional()).System().Equal(system.Traditional())
	s.withNotation = prev.MustWithNotation(next.Notation()).Notation().Equal(next.Notation())
	s.withNumber = prev.MustWithNumbering(number.Mexican()).Numbering().Equal(number.Mexican())
	s.withCatalog = prev.MustWithCatalog(naming.German()).Catalog().Equal(naming.German())
	designed := prev.MustWithDesignations(next.Notation(), next.Numbering(), next.Catalog())
	s.withDesign = designed.Notation().Equal(next.Notation()) && designed.Numbering().Equal(next.Numbering()) &&
		designed.Catalog().Equal(next.Catalog()) && designed.Tuning().Equal(prev.Tuning())
	s.replacement = prev.MustCheckReplacement(next)
	s.structure = prev.MustCheckStructure(next)
	s.designations = prev.MustCheckDesignations(next)
	s.systems = settings.MustCheckSystems(prev.System(), next.System())
	s.changes = prev.MustChanges(next)
	return s
}

// TestConcurrentUse covers K-21, G-6 and P-5: concurrent calls of all
// operations on shared sets give the sequential results and change no input.
// Run it with -race.
func TestConcurrentUse(t *testing.T) {
	t.Parallel()
	tu := tuning.Standard()
	cat := naming.English()
	prev := newSet(t, tu, builtin(notation.EnglishUnicodeID), number.ASPN(), cat)
	nextTuning := tuningFor(t, system.Positions19())
	next := newSet(t, nextTuning, builtin(notation.ABCID), number.FrancoBelgian(), naming.Russian())
	ref := settings.Standard()

	pairs := []struct{ prev, next *settings.Settings }{
		{prev: prev, next: next},
		{prev: prev, next: prev},
		{prev: settings.Standard(), next: next},
	}
	want := make([]snapshot, len(pairs))
	for i, p := range pairs {
		want[i] = take(p.prev, p.next, ref)
	}
	require.True(t, want[0].valid)
	require.True(t, want[0].withDesign)

	const workers = 16
	results := make([][]snapshot, workers)
	var wg sync.WaitGroup
	for w := range workers {
		wg.Go(func() {
			got := make([]snapshot, 0, len(pairs)*4)
			for range 4 {
				for _, p := range pairs {
					got = append(got, take(p.prev, p.next, ref))
				}
			}
			results[w] = got
		})
	}
	wg.Wait()

	for w, got := range results {
		for i, s := range got {
			assert.Equal(t, want[i%len(pairs)], s, "worker %d, call %d", w, i)
		}
	}
	assert.True(t, tuning.Standard().Equal(tu), "the tuning passed to New is unchanged")
	assert.True(t, naming.English().Equal(cat), "the catalog passed to New is unchanged")
	assert.True(t, tuningFor(t, system.Positions19()).Equal(nextTuning))
	assert.True(t, settings.Standard().Equal(prev))
	assert.True(t, ref.Equal(settings.Standard()))
	assert.True(t, settings.MustNew(nextTuning, builtin(notation.ABCID), number.FrancoBelgian(), naming.Russian()).Equal(next))
}

// TestConcurrentSameSource covers K-21 for SameSource with shared sources.
func TestConcurrentSameSource(t *testing.T) {
	t.Parallel()
	std := settings.Standard()
	ptr := &pointerSource{t: t, s: std}
	sources := []settings.Source{std, settings.Standard(), ptr, valueSource{id: 1, s: std}, sliceSource{sets: []*settings.Settings{std}}, nil}
	want := make([][]bool, len(sources))
	for i, a := range sources {
		want[i] = make([]bool, len(sources))
		for j, b := range sources {
			want[i][j] = settings.SameSource(a, b)
		}
	}
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for i, a := range sources {
				for j, b := range sources {
					assert.Equal(t, want[i][j], settings.SameSource(a, b))
				}
			}
		})
	}
	wg.Wait()
}
