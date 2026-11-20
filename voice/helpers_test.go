package voice_test

import (
	"errors"
	"math/big"
	"sync/atomic"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/instrument"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/meter"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempomap"
	"github.com/go-muse/musex/tuning"
	"github.com/go-muse/musex/tuning/temperament"
	"github.com/go-muse/musex/voice"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The fixtures below build fresh values on every call: tests share no
// mutable state and run in parallel.

// ---- Notes ----

// nt returns a note bound to the standard settings (traditional system);
// aspn is the ASPN octave of the position.
func nt(name spelling.Spelling, aspn int) note.Note {
	return note.MustFromASPN(name, aspn, cents.Zero(), settings.Standard())
}

// written returns a note bound to the standard settings with the written
// octave aspn, so that C flat 4 and B sharp 3 keep their written octaves.
func written(name spelling.Spelling, aspn int) note.Note {
	return note.MustFromWrittenASPN(settings.Standard(), name, aspn, cents.Zero())
}

// free returns an unbound note: it has no settings source.
func free(name spelling.Spelling, aspn int) note.Note {
	return note.MustFromASPN(name, aspn, cents.Zero(), nil)
}

// withCents returns a bound note with a cent correction.
func withCents(name spelling.Spelling, aspn, c int) note.Note {
	return note.MustFromASPN(name, aspn, cents.MustFromInt(c), settings.Standard())
}

// noOctave returns a bound note without an octave.
func noOctave() note.Note {
	return note.MustFromSpelling(spelling.C(), cents.Zero(), settings.Standard())
}

// halfSharpName is C half sharp: not admissible in the traditional system.
func halfSharpName() spelling.Spelling {
	return spelling.MustNew(letter.C, alteration.MustFromFraction(1, 2))
}

// halfSharpC returns the unbound note C half sharp 4.
func halfSharpC() note.Note {
	return note.MustFromASPN(halfSharpName(), 4, cents.Zero(), nil)
}

// settings19 returns settings of the 19-position system (КП-44, КП-45).
func settings19() *settings.Settings {
	t := tuning.MustNew(system.Positions19(), temperament.EDO19(), tuning.RootFromName(spelling.C()), tuning.A4At440())
	return settings.Standard().MustWithTuning(t)
}

// settings24 returns settings of the quarter-tone system with 24 classes.
func settings24() *settings.Settings {
	t := tuning.MustNew(system.QuarterTones24(), temperament.EDO24(), tuning.RootFromName(spelling.C()), tuning.A4At440())
	return settings.Standard().MustWithTuning(t)
}

// n19 returns a note bound to settings of the 19-position system.
func n19(name spelling.Spelling, aspn int) note.Note {
	return note.MustFromASPN(name, aspn, cents.Zero(), settings19())
}

// switchSource is a settings source whose settings can be removed after a
// note was created with it.
type switchSource struct {
	current atomic.Pointer[settings.Settings]
}

func (s *switchSource) Settings() *settings.Settings {
	return s.current.Load()
}

// brokenNote returns a bound note whose source no longer returns settings:
// reading them fails with note.ErrUnsetSettings.
func brokenNote(name spelling.Spelling, aspn int) note.Note {
	src := &switchSource{}
	src.current.Store(settings.Standard())
	n := note.MustFromASPN(name, aspn, cents.Zero(), src)
	src.current.Store(nil)
	return n
}

func c4() note.Note { return nt(spelling.C(), 4) }
func d4() note.Note { return nt(spelling.D(), 4) }
func e4() note.Note { return nt(spelling.E(), 4) }
func f4() note.Note { return nt(spelling.F(), 4) }
func g4() note.Note { return nt(spelling.G(), 4) }

func traditional() *system.System { return system.Traditional() }

func pt(name spelling.Spelling, aspn int) pitch.Pitch {
	return pitch.MustFromASPN(name, aspn)
}

// ---- Durations ----

func wr(base duration.Base, dots int) duration.Written {
	return duration.MustNewWritten(base, dots)
}

func q() voice.Duration  { return voice.MustWrittenDuration(wr(duration.Quarter, 0)) }
func e8() voice.Duration { return voice.MustWrittenDuration(wr(duration.Eighth, 0)) }
func h() voice.Duration  { return voice.MustWrittenDuration(wr(duration.Half, 0)) }
func w() voice.Duration  { return voice.MustWrittenDuration(wr(duration.Whole, 0)) }

// dq is a dotted quarter.
func dq() voice.Duration { return voice.MustWrittenDuration(wr(duration.Quarter, 1)) }

func tiedQE() duration.Tied {
	return duration.MustNewTied(wr(duration.Quarter, 0), wr(duration.Eighth, 0))
}

// qe is a quarter tied to an eighth.
func qe() voice.Duration { return voice.MustTiedDuration(tiedQE()) }

// val is a duration without notation.
func val(num, den int) voice.Duration {
	return voice.MustValueDuration(duration.MustNew(num, den))
}

func snd(n note.Note, d voice.Duration) voice.Element { return voice.MustNewSound(n, d) }
func rst(d voice.Duration) voice.Element              { return voice.MustNewRest(d) }

func line(elements ...voice.Element) *voice.Line { return voice.MustNew(elements...) }

func pos(num, den int) duration.Position { return duration.MustNewPosition(num, den) }

func frac(num, den int) fraction.Fraction { return fraction.MustNew(num, den) }

// kp1 is the line of КП-1: C4 quarter, D4 eighth, E4 eighth, F4 half.
func kp1() *voice.Line {
	return line(snd(c4(), q()), snd(d4(), e8()), snd(e4(), e8()), snd(f4(), h()))
}

// kp4 is the line of КП-4: C4 quarter tied to an eighth, D4 eighth,
// quarter rest, E4 quarter.
func kp4() *voice.Line {
	return line(snd(c4(), qe()), snd(d4(), e8()), rst(q()), snd(e4(), q()))
}

// cde is C4, D4, E4 in quarters (КП-10…КП-13).
func cde() *voice.Line {
	return line(snd(c4(), q()), snd(d4(), q()), snd(e4(), q()))
}

// named returns l with a name and all three roles set.
func named(l *voice.Line) *voice.Line {
	return l.MustWithName("Melody").MustWithPlacement(voice.Soprano).
		MustWithThematic(voice.Leading).MustWithImportance(voice.Principal)
}

// ---- Overflow fixtures (architecture 6.1) ----

// bigDenominators returns coprime p, q of about 2^200: 1/p, (p−1)/p, 1/q
// and (q−1)/q are representable, while 1/p + 1/q has a denominator of about
// 2^400, beyond the range of common/fraction.
func bigDenominators() (p, q *big.Int) {
	p = new(big.Int).Lsh(big.NewInt(1), 200)
	p.Add(p, big.NewInt(1))
	q = new(big.Int).Add(p, big.NewInt(2))
	return p, q
}

func bigDuration(num, den *big.Int) duration.Duration {
	return duration.MustFromFraction(fraction.MustFromBigInt(num, den))
}

// overflowDurations returns 1/p, (p−1)/p, 1/q as durations without notation.
func overflowDurations() (invP, restP, invQ duration.Duration) {
	p, q := bigDenominators()
	one := big.NewInt(1)
	return bigDuration(one, p), bigDuration(new(big.Int).Sub(p, one), p), bigDuration(one, q)
}

// overflowLine returns the representable line of rests 1/p, (p−1)/p, 1/q:
// its starts are 0, 1/p, 1 and its end 1 + 1/q.
func overflowLine() *voice.Line {
	invP, restP, invQ := overflowDurations()
	return line(rst(voice.MustValueDuration(invP)), snd(c4(), voice.MustValueDuration(restP)),
		rst(voice.MustValueDuration(invQ)))
}

// overflowWant is the failure of a sum out of the range of common/fraction.
func overflowWant() want {
	return fails(duration.ErrOutOfRange, fraction.ErrOutOfRange)
}

// ---- Ranges and tempo maps ----

func soprano() *instrument.Preset { return instrument.SopranoVoice() }

func bass() *instrument.Preset { return instrument.BassVoice() }

func newRange(lower, upper pitch.Pitch) instrument.Range {
	return instrument.MustNewRange(lower, upper)
}

func quarter(bpm int) tempo.Tempo { return tempo.MustNewQuarter(fraction.FromInt(bpm)) }

func at(num, den int) tempomap.Address { return tempomap.MustAt(pos(num, den)) }

// map60 is ♩ = 60 from position 0.
func map60() tempomap.Map { return tempomap.MustNew(quarter(60)) }

// map60Hold is the map of КП-25: ♩ = 60 and a hold of 2 s at position 3/4.
func map60Hold() tempomap.Map {
	return map60().MustWithExtension(tempomap.MustNewExtension(at(3, 4), tempo.MustNewSeconds(2, 1)))
}

// mapGradual is the map of КП-40 (tempomap КП-7): ♩ = 60 at 0, a linear
// gradual change at 0 and ♩ = 120 at 1.
func mapGradual() tempomap.Map {
	return map60().
		MustWithChange(tempomap.MustNewGradual(at(0, 1), tempomap.LinearSpeed).MustWithSlot(tempomap.SlotSecond)).
		MustWithChange(tempomap.MustNewMain(at(1, 1), quarter(120)))
}

// mapUnfinished0 is ♩ = 60 with a gradual change at 0 that has no end.
func mapUnfinished0() tempomap.Map {
	return map60().MustWithChange(tempomap.MustNewGradual(at(0, 1), tempomap.LinearSpeed).MustWithSlot(tempomap.SlotSecond))
}

// mapUnfinishedHalf is ♩ = 60 with a gradual change at 1/2 that has no end:
// End fails from position 1/2 on.
func mapUnfinishedHalf() tempomap.Map {
	return map60().MustWithChange(tempomap.MustNewGradual(at(1, 2), tempomap.LinearSpeed))
}

// mapMeasured is a map addressed by measures.
func mapMeasured() tempomap.Map {
	return tempomap.MustNewMeasured(measure.MustNew(meter.FourFour()), quarter(60))
}

func seconds(num, den int) tempo.Seconds { return tempo.MustNewSeconds(num, den) }

// assertExactSeconds checks an exact seconds result.
func assertExactSeconds(t *testing.T, want tempo.Seconds, got tempomap.SecondsResult) {
	t.Helper()
	value, err := got.Value()
	require.NoError(t, err)
	assert.True(t, want.Equal(value), "want %v s, got %v", want, got)
	approximate, err := got.Approximate()
	require.NoError(t, err)
	assert.False(t, approximate, "an exact result must not be approximate: %v", got)
}

// ---- Events ----

func noteEvent(n note.Note, base duration.Base, marks event.Marks) *event.Event {
	return event.MustNew(event.MustNoteContent(n), event.MustWrittenDuration(wr(base, 0)), marks)
}

func valueNoteEvent(n note.Note, num, den int, marks event.Marks) *event.Event {
	return event.MustNew(event.MustNoteContent(n), event.MustValueDuration(frac(num, den)), marks)
}

func restEvent(base duration.Base, marks event.Marks) *event.Event {
	return event.MustNew(event.RestContent(), event.MustWrittenDuration(wr(base, 0)), marks)
}

func tie() event.Marks { return event.Marks{Ties: []int{0}} }

// ---- Errors ----

// voiceErrors lists the sentinels of the package (Е-1…Е-18).
func voiceErrors() []error {
	return []error{
		voice.ErrUnset, voice.ErrIndex, voice.ErrNoOctave, voice.ErrChord,
		voice.ErrMeasureDuration, voice.ErrTiedRest, voice.ErrGrace, voice.ErrNoSounds,
		voice.ErrNoRange, voice.ErrZeroWidth, voice.ErrZones, voice.ErrUnknownName,
		voice.ErrBrokenTie, voice.ErrTiedValue, voice.ErrFermataInTie, voice.ErrNoSystem,
		voice.ErrDifferentSystems, voice.ErrMeasureMap,
	}
}

// dependencyErrors lists causes of dependencies that must be absent when the
// order of checks selects another failure first.
func dependencyErrors() []error {
	return []error{
		note.ErrUnset, note.ErrUnsetSettings, note.ErrInadmissible,
		duration.ErrUnset, duration.ErrOutOfRange, fraction.ErrInvalid, fraction.ErrOutOfRange,
		interval.ErrUnset, interval.ErrOutOfRange, pitch.ErrInvalid, system.ErrInvalid,
		system.ErrInadmissible, system.ErrOutOfRange, tempomap.ErrUnset, tempomap.ErrInvalidArgument,
		tempomap.ErrUnfinished, rounding.ErrUnset, rounding.ErrUnknownMode, instrument.ErrUnset,
		instrument.ErrInadmissible, instrument.ErrInverted, event.ErrUnset,
	}
}

const noItem = -1

// want describes an expected failure.
type want struct {
	is   []error // causes the error must match
	not  []error // causes the error must not match
	item int     // Index of *voice.ItemError, or noItem
}

// fails expects an error without *voice.ItemError.
func fails(is ...error) want { return want{is: is, item: noItem} }

// failsAt expects an error wrapped in *voice.ItemError with the index.
func failsAt(index int, is ...error) want { return want{is: is, item: index} }

// without additionally forbids the causes.
func (w want) without(causes ...error) want {
	w.not = append(append([]error(nil), w.not...), causes...)
	return w
}

// only additionally forbids every listed dependency cause not in w.is: the
// failure must come from the first check alone.
func (w want) only() want {
	var extra []error
	for _, dep := range dependencyErrors() {
		if !containsError(w.is, dep) {
			extra = append(extra, dep)
		}
	}
	return w.without(extra...)
}

func containsError(list []error, target error) bool {
	for _, err := range list {
		if err == target {
			return true
		}
	}
	return false
}

// checkErr checks err against w: every listed cause, no other sentinel of
// the package, never the system sentinels of interval (architecture 6.2,
// КР-V6), and the *voice.ItemError index.
func checkErr(t *testing.T, err error, w want) {
	t.Helper()
	require.Error(t, err)
	for _, cause := range w.is {
		assert.ErrorIs(t, err, cause)
	}
	for _, sentinel := range voiceErrors() {
		if !containsError(w.is, sentinel) {
			assert.NotErrorIs(t, err, sentinel)
		}
	}
	assert.NotErrorIs(t, err, interval.ErrNoSystem)
	assert.NotErrorIs(t, err, interval.ErrDifferentSystems)
	for _, cause := range w.not {
		assert.NotErrorIs(t, err, cause)
	}
	itemErr, ok := errors.AsType[*voice.ItemError](err)
	if w.item == noItem {
		assert.False(t, ok, "unexpected *voice.ItemError in %v", err)
		return
	}
	require.True(t, ok, "want *voice.ItemError, got %v", err)
	assert.Equal(t, w.item, itemErr.Index)
}

// recovered runs f and returns the value it panicked with, or nil.
func recovered(f func()) (value any) {
	defer func() { value = recover() }()
	f()
	return nil
}

// checkPanic checks that f panics with an error that satisfies w.
func checkPanic(t *testing.T, w want, f func()) {
	t.Helper()
	value := recovered(f)
	err, ok := value.(error)
	require.True(t, ok, "Must must panic with an error, got %T (%v)", value, value)
	checkErr(t, err, w)
}

// assertFailure checks a failing operation and its Must variant: zero
// results, the expected causes, and a panic with the same causes.
func assertFailure[T any](t *testing.T, w want, call func() (T, error), must func() T) {
	t.Helper()
	got, err := call()
	assert.Zero(t, got, "a failure must return zero results")
	checkErr(t, err, w)
	checkPanic(t, w, func() { must() })
}

// pair holds two results of one call.
type pair[A, B any] struct {
	A A
	B B
}

func two[A, B any](a A, b B, err error) (pair[A, B], error) {
	return pair[A, B]{A: a, B: b}, err
}

func mustTwo[A, B any](a A, b B) pair[A, B] {
	return pair[A, B]{A: a, B: b}
}

// ---- Comparisons ----

func assertLine(t *testing.T, want, got *voice.Line) {
	t.Helper()
	require.NotNil(t, got)
	assert.True(t, want.Equal(got), "want %v, got %v", want, got)
}

func assertNote(t *testing.T, want, got note.Note) {
	t.Helper()
	assert.True(t, want.Equal(got), "want %v, got %v", want, got)
}

func assertNotes(t *testing.T, want, got []note.Note) {
	t.Helper()
	require.Len(t, got, len(want))
	for i := range want {
		assert.True(t, want[i].Equal(got[i]), "note %d: want %v, got %v", i, want[i], got[i])
	}
}

func assertElement(t *testing.T, want, got voice.Element) {
	t.Helper()
	assert.True(t, want.Equal(got), "want %v, got %v", want, got)
}

func positionStrings(ps []duration.Position) []string {
	out := make([]string, len(ps))
	for i, p := range ps {
		out[i] = p.String()
	}
	return out
}

func intervalStrings(ms []voice.MelodicInterval) []string {
	out := make([]string, len(ms))
	for i, m := range ms {
		out[i] = m.Interval.String()
	}
	return out
}
