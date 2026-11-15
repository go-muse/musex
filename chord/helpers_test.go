package chord_test

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"
	"github.com/go-muse/musex/tuning/temperament"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The helpers build every expectation from business-spec (section 6.1 and
// its conventions in 1.3) and architecture (section 5) through the public API
// of package note and its dependencies. No expectation is copied from the
// implementation of package chord.

// sentinels lists the four sentinels of package chord (architecture 5.4).
var sentinels = []error{
	chord.ErrUnset,
	chord.ErrNotFound,
	chord.ErrIndex,
	chord.ErrRuleFailed,
}

// ---------------------------------------------------------------------------
// Sets of settings and sources (business-spec 1.3).

// std returns the standard settings. Two results are equal sets, so
// settings.SameSource takes them for one source (note П-5а, Р-13); where two
// notes with the standard settings must be told apart, use stdSource.
func std() *settings.Settings { return settings.Standard() }

// stdSource returns a new source that gives the standard settings. Each result
// is its own pointer, so two results are two different sources.
func stdSource() settings.Source { return newCounting() }

// set19 returns "set 19" (note #requirements 6.1): the 19-position system,
// 19-EDO rooted at C, A4 at 440 Hz, otherwise standard.
func set19() *settings.Settings {
	t19 := tuning.MustNew(system.Positions19(), temperament.Equal19(),
		tuning.RootFromName(spelling.C()), tuning.A4At440())
	return std().MustWithTuning(t19)
}

// countingSource is a source of settings that counts its reads (KP-4). It is
// a pointer, so its dynamic type is comparable, and it is safe for concurrent
// calls, as settings.Source requires.
type countingSource struct {
	set   *settings.Settings
	calls atomic.Int64
}

func newCounting() *countingSource { return &countingSource{set: std()} }

func (s *countingSource) Settings() *settings.Settings {
	s.calls.Add(1)
	return s.set
}

func (s *countingSource) reads() int64 { return s.calls.Load() }

// ---------------------------------------------------------------------------
// Notes (business-spec 1.3): a name and the ASPN number of its Octave, a
// correction in cents, a source (nil: unbound).

// pn returns the note name with the Octave aspn, c cents, bound to src.
func pn(name spelling.Spelling, aspn, c int, src settings.Source) note.Note {
	return note.MustFromPitch(pitch.MustFromASPN(name, aspn), cents.MustFromInt(c), src)
}

// nn returns the note name without an octave, 0 cents, bound to src.
func nn(name spelling.Spelling, src settings.Source) note.Note {
	return note.MustFromSpelling(name, cents.Zero(), src)
}

// on returns the note name with the Octave whose ASPN number is aspn, 0 cents,
// unbound; it reaches the octaves at the ends of int (K-31).
func on(name spelling.Spelling, aspn int) note.Note {
	return note.MustFromOctave(name, octave.FromASPN(aspn), cents.Zero(), nil)
}

// Unbound notes of the examples.
var (
	c4    = pn(spelling.C(), 4, 0, nil)
	d4    = pn(spelling.D(), 4, 0, nil)
	e4    = pn(spelling.E(), 4, 0, nil)
	f4    = pn(spelling.F(), 4, 0, nil)
	g4    = pn(spelling.G(), 4, 0, nil)
	b4    = pn(spelling.B(), 4, 0, nil)
	c5    = pn(spelling.C(), 5, 0, nil)
	e3    = pn(spelling.E(), 3, 0, nil)
	e5    = pn(spelling.E(), 5, 0, nil)
	e6    = pn(spelling.E(), 6, 0, nil)
	g5    = pn(spelling.G(), 5, 0, nil)
	c4p5  = pn(spelling.C(), 4, 5, nil)
	c4p10 = pn(spelling.C(), 4, 10, nil)
	c4p30 = pn(spelling.C(), 4, 30, nil)
	c5p30 = pn(spelling.C(), 5, 30, nil)
	bs4   = pn(spelling.BSharp(), 4, 0, nil)
	cNo   = nn(spelling.C(), nil)
	eNo   = nn(spelling.E(), nil)
	csNo  = nn(spelling.CSharp(), nil)
	cMax  = on(spelling.C(), math.MaxInt)
	cMin  = on(spelling.C(), math.MinInt)
)

// list returns its arguments as a slice; it keeps the tables short.
func list(notes ...note.Note) []note.Note { return notes }

// mk returns the chord with the notes; it stops the test if creation fails.
func mk(t *testing.T, notes ...note.Note) chord.Chord {
	t.Helper()
	c, err := chord.New(notes...)
	require.NoError(t, err, "test setup: chord.New")
	return c
}

// ---------------------------------------------------------------------------
// Checks of a chord.

// sameNote reports whether got is the same note as want: equal records and
// corrections (note.Note.Equal) and the same source, or both unbound. Two
// unset notes are the same here (an unchanged list of inputs).
func sameNote(want, got note.Note) bool {
	if !want.IsValid() || !got.IsValid() {
		return want.IsValid() == got.IsValid()
	}
	if !want.Equal(got) || want.IsBound() != got.IsBound() {
		return false
	}
	return !want.IsBound() || want.SameSource(got)
}

// assertNotes checks that got holds the notes of want in the same order,
// each with the same source (business-spec 6.1: "the chord is the same").
func assertNotes(t *testing.T, want, got []note.Note, msgAndArgs ...any) bool {
	t.Helper()
	if !assert.Len(t, got, len(want), msgAndArgs...) {
		t.Logf("notes %v, want %v", got, want)
		return false
	}
	ok := true
	for i := range want {
		if !sameNote(want[i], got[i]) {
			ok = false
			assert.Fail(t, fmt.Sprintf("note %d: got %v (bound %t), want %v (bound %t); notes %v, want %v",
				i, got[i], got[i].IsBound(), want[i], want[i].IsBound(), got, want), msgAndArgs...)
		}
	}
	return ok
}

// assertChord checks that c is set and holds the notes of want in the record
// order with their sources.
func assertChord(t *testing.T, want []note.Note, c chord.Chord, msgAndArgs ...any) bool {
	t.Helper()
	ok := assert.True(t, c.IsValid(), "chord is set")
	ok = assert.Equal(t, len(want), c.Len(), "Len") && ok
	ok = assert.Equal(t, len(want) == 0, c.IsEmpty(), "IsEmpty") && ok
	return assertNotes(t, want, c.Notes(), msgAndArgs...) && ok
}

// assertUnsetChord checks that c is the unset chord.
func assertUnsetChord(t *testing.T, c chord.Chord) bool {
	t.Helper()
	ok := assert.False(t, c.IsValid(), "chord is unset")
	ok = assert.True(t, c.IsEmpty(), "unset chord is empty") && ok
	ok = assert.Zero(t, c.Len(), "unset chord has no notes") && ok
	return assert.Equal(t, "Chord(unset)", c.String()) && ok
}

// state is what a chord holds before an operation: whether it is set and its
// notes with their sources.
type state struct {
	valid bool
	notes []note.Note
}

func stateOf(c chord.Chord) state { return state{valid: c.IsValid(), notes: c.Notes()} }

// assertState checks that c still holds s (business-spec 6.1: "the chord is
// the same").
func assertState(t *testing.T, s state, c chord.Chord, msgAndArgs ...any) bool {
	t.Helper()
	if !s.valid {
		return assertUnsetChord(t, c)
	}
	return assertChord(t, s.notes, c, msgAndArgs...)
}

// ---------------------------------------------------------------------------
// Checks of errors (architecture 5.4).

// assertOnly checks that err matches target and no other sentinel of chord.
func assertOnly(t *testing.T, err, target error, msgAndArgs ...any) bool {
	t.Helper()
	if !assert.ErrorIs(t, err, target, msgAndArgs...) {
		return false
	}
	ok := true
	for _, s := range sentinels {
		if s != target {
			ok = assert.NotErrorIs(t, err, s, msgAndArgs...) && ok
		}
	}
	return ok
}

// assertNoChordSentinel checks that err matches none of the sentinels of chord
// (P-10a: the causes of note at the octave shifts).
func assertNoChordSentinel(t *testing.T, err error, msgAndArgs ...any) bool {
	t.Helper()
	ok := assert.Error(t, err, msgAndArgs...)
	for _, s := range sentinels {
		ok = assert.NotErrorIs(t, err, s, msgAndArgs...) && ok
	}
	return ok
}

// assertRuleFailed checks E-4: the error matches ErrRuleFailed and the error of
// the rule, and no other sentinel of chord.
func assertRuleFailed(t *testing.T, err, ruleErr error, msgAndArgs ...any) bool {
	t.Helper()
	ok := assertOnly(t, err, chord.ErrRuleFailed, msgAndArgs...)
	return assert.ErrorIs(t, err, ruleErr, msgAndArgs...) && ok
}

// recovered calls f and returns the value of its panic, if any.
func recovered(f func()) (v any, panicked bool) {
	defer func() {
		v = recover()
	}()
	panicked = true
	f()
	panicked = false
	return v, panicked
}

// assertPanicsWith checks that f panics with an error matching target.
func assertPanicsWith(t *testing.T, target error, f func(), msgAndArgs ...any) bool {
	t.Helper()
	v, panicked := recovered(f)
	if !assert.True(t, panicked, msgAndArgs...) {
		return false
	}
	err, isErr := v.(error)
	if !assert.True(t, isErr, "panic value %#v is an error", v) {
		return false
	}
	return assert.ErrorIs(t, err, target, msgAndArgs...)
}

// ---------------------------------------------------------------------------
// Rules of the user (architecture 5.8).

// errRule is the error of the failing test rules.
var errRule = errors.New("chord_test: the rule failed")

// hasOctave is the condition "the note has an octave" (K-32).
func hasOctave(n note.Note) (bool, error) { return n.HasOctave(), nil }

// up is the transform "move the note one octave up" (K-28).
func up(n note.Note) (note.Note, error) {
	err := n.Up()
	return n, err
}

// byRecord is the agreed order of notes as a comparison rule.
func byRecord(a, b note.Note) (int, error) { return a.Compare(b), nil }

// reverseByRecord is the reverse agreed order of notes (K-26).
func reverseByRecord(a, b note.Note) (int, error) { return b.Compare(a), nil }

// recorder records the notes given to a rule, in the order of the calls. It is safe
// for concurrent use.
type recorder struct {
	mu    sync.Mutex
	notes []note.Note
}

func (l *recorder) add(n note.Note) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.notes = append(l.notes, n)
	return len(l.notes)
}

func (l *recorder) got() []note.Note {
	l.mu.Lock()
	defer l.mu.Unlock()
	return slices.Clone(l.notes)
}

func (l *recorder) calls() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.notes)
}

// cond returns a condition that records its notes and answers with f. Before
// answering it changes the copy it got, which must not change the chord.
func (l *recorder) cond(f func(note.Note) bool) chord.Condition {
	return func(n note.Note) (bool, error) {
		l.add(n)
		ok := f(n)
		_ = n.SetCents(cents.MustFromInt(7)) // changes only the copy given to the rule
		return ok, nil
	}
}

// condFailAt returns a condition that records its notes, answers with f and
// fails with errRule at the k-th call (from 1).
func (l *recorder) condFailAt(k int, f func(note.Note) bool) chord.Condition {
	return func(n note.Note) (bool, error) {
		if l.add(n) == k {
			return false, errRule
		}
		return f(n), nil
	}
}

// transform returns a transform that records its notes and applies f.
func (l *recorder) transform(f chord.Transform) chord.Transform {
	return func(n note.Note) (note.Note, error) {
		l.add(n)
		return f(n)
	}
}

// transformFailAt returns a transform that records its notes, applies f and
// fails with errRule at the k-th call (from 1).
func (l *recorder) transformFailAt(k int, f chord.Transform) chord.Transform {
	return func(n note.Note) (note.Note, error) {
		if l.add(n) == k {
			return n, errRule
		}
		return f(n)
	}
}

// yes and no are constant answers of conditions.
func yes(note.Note) bool { return true }
func no(note.Note) bool  { return false }

// ---------------------------------------------------------------------------
// Oracles of the composition (business-spec P-6, P-10; architecture 5.9,
// 5.10), built only from note.Note.Equal, SameRecord and Compare.

// count returns the multiplicity of x in notes by eq.
func count(notes []note.Note, x note.Note, eq func(a, b note.Note) bool) int {
	k := 0
	for _, n := range notes {
		if eq(n, x) {
			k++
		}
	}
	return k
}

func equalNotes(a, b note.Note) bool { return a.Equal(b) }

func sameRecordNotes(a, b note.Note) bool { return a.SameRecord(b) }

// sameComposition reports whether every note has one multiplicity in a and b.
func sameComposition(a, b []note.Note, eq func(x, y note.Note) bool) bool {
	if len(a) != len(b) {
		return false
	}
	for _, x := range a {
		if count(a, x, eq) != count(b, x, eq) {
			return false
		}
	}
	return true
}

// includes reports whether a holds every note of b with its multiplicity.
func includes(a, b []note.Note) bool {
	for _, x := range b {
		if count(a, x, equalNotes) < count(b, x, equalNotes) {
			return false
		}
	}
	return true
}

// common reports whether a note of a occurs in b.
func common(a, b []note.Note) bool {
	for _, x := range a {
		if count(b, x, equalNotes) > 0 {
			return true
		}
	}
	return false
}

// compareNotes is the agreed order of chords for set chords (P-6): the notes
// in the agreed order of notes, note by note, a shorter prefix first.
func compareNotes(a, b []note.Note) int {
	x := slices.Clone(a)
	y := slices.Clone(b)
	slices.SortStableFunc(x, note.Note.Compare)
	slices.SortStableFunc(y, note.Note.Compare)
	for i := range min(len(x), len(y)) {
		if r := x[i].Compare(y[i]); r != 0 {
			return r
		}
	}
	switch {
	case len(x) < len(y):
		return -1
	case len(x) > len(y):
		return 1
	default:
		return 0
	}
}

// unionOf is the expected Union (P-10): a in place; then, in the record order
// of b, each occurrence of b beyond the multiplicity in a.
func unionOf(a, b []note.Note) []note.Note {
	out := slices.Clone(a)
	for i, y := range b {
		j := count(b[:i+1], y, equalNotes) // y is the j-th occurrence in b
		if j > count(a, y, equalNotes) {
			out = append(out, y)
		}
	}
	return out
}

// intersectOf is the expected Intersect (P-10): the first occurrences of a,
// as many as the smaller multiplicity.
func intersectOf(a, b []note.Note) []note.Note {
	out := []note.Note{}
	for i, x := range a {
		j := count(a[:i+1], x, equalNotes)
		if j <= count(b, x, equalNotes) {
			out = append(out, x)
		}
	}
	return out
}

// subtractOf is the expected Subtract (P-10): for each note of b in its
// record order, the first remaining occurrence is removed from a.
func subtractOf(a, b []note.Note) []note.Note {
	out := slices.Clone(a)
	for _, y := range b {
		if i := slices.IndexFunc(out, y.Equal); i >= 0 {
			out = slices.Delete(out, i, i+1)
		}
	}
	if out == nil {
		out = []note.Note{}
	}
	return out
}

// distinctOf is the expected DistinctNotes: the first occurrences in order.
func distinctOf(a []note.Note) []note.Note {
	out := []note.Note{}
	for i, x := range a {
		if count(a[:i], x, equalNotes) == 0 {
			out = append(out, x)
		}
	}
	return out
}

// pool is a small set of notes for the property tests (plan, "Properties"):
// repeats by record with different sources, a correction, a note without an
// octave, a note that differs only by name.
func pool() []note.Note {
	return []note.Note{c4, c4p5, pn(spelling.C(), 4, 0, std()), cNo, e4, bs4}
}

// sequences returns all sequences of notes of p up to length n, in a fixed
// order, starting with the empty one.
func sequences(p []note.Note, n int) [][]note.Note {
	out := [][]note.Note{{}}
	prev := [][]note.Note{{}}
	for range n {
		var next [][]note.Note
		for _, s := range prev {
			for _, x := range p {
				next = append(next, append(slices.Clone(s), x))
			}
		}
		out = append(out, next...)
		prev = next
	}
	return out
}

// ---------------------------------------------------------------------------
// Long chords for the stability of the sorts (architecture 5.6, 5.10; ОП-5,
// ОП-35, К-27). They are longer than 12 notes, where a sort of a short slice
// may behave as a stable one even when it is not.

// longOrder is a shuffled order of 20 indexes into the records of longBound.
var longOrder = []int{3, 1, 0, 2, 4, 1, 3, 0, 2, 1, 4, 3, 0, 1, 2, 3, 4, 0, 1, 2}

// longBound returns 20 notes of 5 records (C, C4, C4+5, E4, G4), each bound to
// its own source with the standard settings: notes of one record are equal by
// record but told apart by their sources.
func longBound() []note.Note {
	records := []func(src settings.Source) note.Note{
		func(src settings.Source) note.Note { return nn(spelling.C(), src) },
		func(src settings.Source) note.Note { return pn(spelling.C(), 4, 0, src) },
		func(src settings.Source) note.Note { return pn(spelling.C(), 4, 5, src) },
		func(src settings.Source) note.Note { return pn(spelling.E(), 4, 0, src) },
		func(src settings.Source) note.Note { return pn(spelling.G(), 4, 0, src) },
	}
	out := make([]note.Note, 0, len(longOrder))
	for _, k := range longOrder {
		out = append(out, records[k](stdSource()))
	}
	return out
}

// longNames returns 20 unbound notes with the names C, E, G in several octaves,
// with and without octaves and corrections: a rule that compares only the
// names (byName) finds many of them equal, and they stay distinguishable.
func longNames() []note.Note {
	return list(
		pn(spelling.G(), 5, 0, nil), c4, eNo, pn(spelling.C(), 6, 5, nil), e4,
		cNo, g4, pn(spelling.E(), 3, 0, nil), c5, pn(spelling.G(), 3, 7, nil),
		c4p5, e5, pn(spelling.C(), 2, 0, nil), nn(spelling.G(), nil), pn(spelling.E(), 6, 9, nil),
		c4p30, pn(spelling.G(), 6, 0, nil), pn(spelling.E(), 4, 3, nil), c5p30, pn(spelling.G(), 4, 5, nil),
	)
}

// byName compares only the names of the notes, by the agreed order of notes
// without an octave: C5, C4+5 and C are equal by it.
func byName(a, b note.Note) (int, error) {
	return nn(a.Spelling(), nil).Compare(nn(b.Spelling(), nil)), nil
}

// stableSorted returns a copy of notes sorted stably by cmp: the expectation
// of a stable sort by its definition.
func stableSorted(notes []note.Note, cmp func(a, b note.Note) int) []note.Note {
	out := slices.Clone(notes)
	slices.SortStableFunc(out, cmp)
	return out
}
