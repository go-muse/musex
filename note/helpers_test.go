package note_test

import (
	"slices"
	"sync/atomic"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/notation"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/octave/number"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"
	"github.com/go-muse/musex/tuning/temperament"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The helpers build every expectation through the public API of the
// dependencies (business-spec 6.1 conventions; plan, Sh1 "Expectations"),
// never through package note.

// sentinels lists the eight sentinels of package note (architecture 6.2.1).
var sentinels = []error{
	note.ErrUnset,
	note.ErrNoSource,
	note.ErrUnsetSettings,
	note.ErrNoOctave,
	note.ErrInadmissible,
	note.ErrDifferentSystems,
	note.ErrInexpressible,
	note.ErrOutOfRange,
}

// ---------------------------------------------------------------------------
// Sets of settings (business-spec 6.1, conventions).

// std returns the standard settings: a new pointer to an equal set on each
// call (settings, Standard).
func std() *settings.Settings { return settings.Standard() }

// tuning19 is the tuning of "set 19": the 19-position system, 19-EDO rooted
// at C, A4 at 440 Hz.
func tuning19() *tuning.Tuning {
	return tuning.MustNew(system.Positions19(), temperament.Equal19(),
		tuning.RootFromName(spelling.C()), tuning.A4At440())
}

// tuning24 is the tuning of "set 24": the 24-quarter-tone system, 24-EDO
// rooted at C, A4 at 440 Hz.
func tuning24() *tuning.Tuning {
	return tuning.MustNew(system.QuarterTones24(), temperament.Equal24(),
		tuning.RootFromName(spelling.C()), tuning.A4At440())
}

// set19 is "set 19": the standard set with the 19-position tuning.
func set19() *settings.Settings { return std().MustWithTuning(tuning19()) }

// set24 is "set 24": the 24-quarter-tone tuning and the lilypond-english
// notation, otherwise standard.
func set24() *settings.Settings {
	return std().MustWithTuning(tuning24()).MustWithNotation(notation.MustBuiltIn("lilypond-english"))
}

// systemK25 is the system of K-25: traditional layout, alterations -1, 0, +1.
func systemK25() *system.System {
	return system.MustNewRegular(2, 1, letter.C, system.MustIntegerRange(-1, 1), "k25")
}

// setK25 is the standard set with the system of K-25.
func setK25() *settings.Settings { return std().MustWithSystem(systemK25()) }

// stdNotation is the standard set with the built-in notation id.
func stdNotation(id string) *settings.Settings {
	return std().MustWithNotation(notation.MustBuiltIn(id))
}

// stdFB is the standard set with the Franco-Belgian octave numbering.
func stdFB() *settings.Settings { return std().MustWithNumbering(number.FrancoBelgian()) }

// stdAt is the standard set with A4 at freq hertz.
func stdAt(freq float64) *settings.Settings {
	return std().MustWithTuning(tuning.Standard().MustWithReferenceFrequency(freq))
}

// ---------------------------------------------------------------------------
// Sources (plan, Sh1 "Sets and sources"). All are pointers to structs, so their
// dynamic types are comparable and notes holding them support ==; all are
// safe for concurrent calls, as settings.Source requires.

// countingSource counts its reads.
type countingSource struct {
	set   *settings.Settings
	calls atomic.Int64
}

func newCounting(set *settings.Settings) *countingSource { return &countingSource{set: set} }

func (s *countingSource) Settings() *settings.Settings {
	s.calls.Add(1)
	return s.set
}

func (s *countingSource) reads() int64 { return s.calls.Load() }

// alternatingSource returns a and b in turn, starting with a (K-24a).
type alternatingSource struct {
	a, b  *settings.Settings
	calls atomic.Int64
}

func (s *alternatingSource) Settings() *settings.Settings {
	if s.calls.Add(1)%2 == 1 {
		return s.a
	}
	return s.b
}

// unsetSource always returns the unset set (K-29).
type unsetSource struct{ _ byte }

func (*unsetSource) Settings() *settings.Settings { return nil }

// switchSource returns its current set; the set can be replaced at any time
// (K-24, concurrency, notes whose name becomes inadmissible).
type switchSource struct {
	p atomic.Pointer[settings.Settings]
}

func newSwitch(set *settings.Settings) *switchSource {
	s := &switchSource{}
	s.p.Store(set)
	return s
}

func (s *switchSource) Settings() *settings.Settings { return s.p.Load() }

func (s *switchSource) set(set *settings.Settings) { s.p.Store(set) }

// nilSettingsSource is a non-nil Source holding a nil *settings.Settings:
// a source with an unset set, not a missing source (architecture 6.2.3, E-1).
func nilSettingsSource() settings.Source { return (*settings.Settings)(nil) }

// ---------------------------------------------------------------------------
// Names, pitches, cents.

// sp returns the name with letter l and the alteration num/den.
func sp(l letter.Letter, num, den int) spelling.Spelling {
	return spelling.MustNew(l, alteration.MustFromFraction(num, den))
}

// spi returns the name with letter l and the integer alteration a.
func spi(l letter.Letter, a int) spelling.Spelling {
	return spelling.MustNew(l, alteration.MustFromInt(a))
}

// pt returns the pitch name with the ASPN number aspn of its Octave.
func pt(name spelling.Spelling, aspn int) pitch.Pitch { return pitch.MustFromASPN(name, aspn) }

// ct returns n cents.
func ct(n int) cents.Cents { return cents.MustFromInt(n) }

// oct returns the Octave with the ASPN number aspn.
func oct(aspn int) octave.Octave { return octave.FromASPN(aspn) }

// alt returns the alteration num/den.
func alt(num, den int) alteration.Alteration { return alteration.MustFromFraction(num, den) }

// tiny is 10^-15 cent: exact heights differ by it, while the frequencies
// after the single rounding coincide (checked by requireTinyMerges).
func tiny() cents.Cents { return cents.MustFromFraction(1, 1_000_000_000_000_000) }

// requireTinyMerges checks the precondition of the exact-path tests through
// package tuning: C4 and C4 shifted by tiny have the same float64 frequency.
func requireTinyMerges(t *testing.T) {
	t.Helper()
	tu := tuning.Standard()
	c4 := pt(spelling.C(), 4)
	require.Equal(t, tu.MustShiftedFrequency(c4, cents.Zero()), tu.MustShiftedFrequency(c4, tiny()),
		"precondition: C4 and C4 + 1e-15 cent have one float64 frequency")
	require.NotZero(t, tiny().Sign())
}

// classRule is a custom spelling rule of the traditional system that names
// class k as C with the alteration k: C, C♯, C𝄪, C+3, ... It gives names that
// are admissible in the traditional system but not in the system of K-25.
func classRule() system.SpellingRule {
	return system.MustCustomSpelling(func(_ *system.System, c pitchclass.Class) spelling.Spelling {
		return spi(letter.C, c.Number())
	})
}

// badRule is a custom spelling rule that returns the unset name
// (system.ErrBadSpelling).
func badRule() system.SpellingRule {
	return system.MustCustomSpelling(func(_ *system.System, _ pitchclass.Class) spelling.Spelling {
		return spelling.Spelling{}
	})
}

// ---------------------------------------------------------------------------
// Notes. The creation functions are under test themselves (create_test.go);
// here they only prepare inputs, and require stops the test if they fail.

// mk returns the note (name, Octave with ASPN aspn) + c cents bound to src
// (nil: unbound).
func mk(t *testing.T, name spelling.Spelling, aspn int, c cents.Cents, src settings.Source) note.Note {
	t.Helper()
	n, err := note.FromPitch(pt(name, aspn), c, src)
	require.NoError(t, err, "test setup: FromPitch(%v, %d, %v)", name, aspn, c)
	return n
}

// mkName returns the note without octave name + c cents bound to src.
func mkName(t *testing.T, name spelling.Spelling, c cents.Cents, src settings.Source) note.Note {
	t.Helper()
	n, err := note.FromSpelling(name, c, src)
	require.NoError(t, err, "test setup: FromSpelling(%v, %v)", name, c)
	return n
}

// inadmissibleNote returns the note C𝄪4 bound to a source that, after the
// creation, returns the set of K-25 where C𝄪 is not admissible (business-spec
// P-4: possible only with a source without compatibility checks).
func inadmissibleNote(t *testing.T) (note.Note, *switchSource) {
	t.Helper()
	src := newSwitch(std())
	n := mk(t, spelling.CDoubleSharp(), 4, ct(0), src)
	src.set(setK25())
	return n, src
}

// inadmissibleNameNote is like inadmissibleNote, without octave.
func inadmissibleNameNote(t *testing.T) (note.Note, *switchSource) {
	t.Helper()
	src := newSwitch(std())
	n := mkName(t, spelling.CDoubleSharp(), ct(0), src)
	src.set(setK25())
	return n, src
}

// want describes an expected note through dependency values.
type want struct {
	name     spelling.Spelling
	hasOct   bool
	aspn     int // ASPN number of the Octave when hasOct
	cents    cents.Cents
	bound    bool
	src      settings.Source // compared by == when bound
	checkSrc bool
}

// wp is the expectation of a note with octave.
func wp(name spelling.Spelling, aspn int, c cents.Cents) want {
	return want{name: name, hasOct: true, aspn: aspn, cents: c}
}

// wn is the expectation of a note without octave.
func wn(name spelling.Spelling, c cents.Cents) want {
	return want{name: name, cents: c}
}

// boundTo sets the expected source.
func (w want) boundTo(src settings.Source) want {
	w.bound, w.src, w.checkSrc = true, src, true
	return w
}

// free marks the expected note as unbound.
func (w want) free() want {
	w.bound, w.src, w.checkSrc = false, nil, true
	return w
}

// assertNote checks a note part by part: name, octave, cents and, when the
// expectation says so, its source (by ==, the identity of the source value).
func assertNote(t *testing.T, w want, got note.Note, msgAndArgs ...any) bool {
	t.Helper()
	ok := assert.True(t, got.IsValid(), msgAndArgs...)
	ok = assert.Truef(t, got.Spelling().Equal(w.name), "name: want %v, got %v", w.name, got.Spelling()) && ok
	ok = assert.Equal(t, w.hasOct, got.HasOctave(), msgAndArgs...) && ok
	if w.hasOct && got.HasOctave() {
		o, err := got.Octave()
		ok = assert.NoError(t, err, msgAndArgs...) && ok
		ok = assert.Truef(t, o.Equal(oct(w.aspn)), "octave: want %v, got %v", oct(w.aspn), o) && ok
	}
	ok = assert.Truef(t, got.Cents().Equal(w.cents), "cents: want %v, got %v", w.cents, got.Cents()) && ok
	if w.checkSrc {
		ok = assert.Equal(t, w.bound, got.IsBound(), msgAndArgs...) && ok
		ok = assert.Truef(t, got.Source() == w.src, "source: want %v, got %v", w.src, got.Source()) && ok
	}
	return ok
}

// assertSame checks that got is the note before: the same parts and the same
// source (==, plan: "the note did not change").
func assertSame(t *testing.T, before, got note.Note, msgAndArgs ...any) bool {
	t.Helper()
	return assert.True(t, before == got, msgAndArgs...)
}

// ---------------------------------------------------------------------------
// Errors.

// assertErr checks that err matches the note sentinel want (nil: no note
// sentinel) and every error in extra, and no other note sentinel
// (architecture 6.2.1: an error matches at most one sentinel of note).
func assertErr(t *testing.T, err, want error, extra ...error) bool {
	t.Helper()
	if !assert.Error(t, err) {
		return false
	}
	ok := true
	for _, s := range sentinels {
		if s == want {
			ok = assert.ErrorIsf(t, err, s, "want sentinel %v", s) && ok
		} else {
			ok = assert.NotErrorIsf(t, err, s, "unexpected sentinel %v in %v", s, err) && ok
		}
	}
	for _, e := range extra {
		ok = assert.ErrorIsf(t, err, e, "want also %v", e) && ok
	}
	return ok
}

// panicValue runs f and returns the value it panicked with, or nil.
func panicValue(f func()) (v any) {
	defer func() { v = recover() }()
	f()
	return nil
}

// assertPanicErr checks that f panics with an error value matching want and
// extra, as a Must variant does (architecture 6.11).
func assertPanicErr(t *testing.T, f func(), want error, extra ...error) bool {
	t.Helper()
	v := panicValue(f)
	if !assert.NotNil(t, v, "want a panic") {
		return false
	}
	err, isErr := v.(error)
	if !assert.True(t, isErr, "the panic value is an error: %v", v) {
		return false
	}
	return assertErr(t, err, want, extra...)
}

// assertNoPanic checks that f does not panic.
func assertNoPanic(t *testing.T, f func()) bool {
	t.Helper()
	return assert.Nil(t, panicValue(f), "unexpected panic")
}

// collect returns the notes of a sequence.
func collect(seq func(func(note.Note) bool)) []note.Note {
	var out []note.Note
	for n := range seq {
		out = append(out, n)
	}
	return out
}

// assertNotes checks a list of notes against expectations, in order.
func assertNotes(t *testing.T, ws []want, got []note.Note) bool {
	t.Helper()
	if !assert.Len(t, got, len(ws)) {
		return false
	}
	ok := true
	for i := range ws {
		ok = assertNote(t, ws[i], got[i], "note %d", i) && ok
	}
	return ok
}

// sameNotes reports whether two slices hold == notes in the same order.
func sameNotes(a, b []note.Note) bool { return slices.Equal(a, b) }

// cHalf is C +1/2: admissible in the 24-quarter-tone system only.
func cHalf() spelling.Spelling { return sp(letter.C, 1, 2) }
