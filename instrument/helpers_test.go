package instrument_test

import (
	"errors"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/instrument"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/notation"
	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/octave/number"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"
	"github.com/go-muse/musex/tuning/temperament"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The fixtures below follow business-spec 6.1. Every function returns a new
// value, so no test shares a pointer with another one; the values themselves
// are immutable.

// p parses a pitch written as the standard settings write it: "B♯3" is B♯
// with Octave 4 and "C♭4" is C♭ with Octave 3 (settings К-11).
func p(text string) pitch.Pitch {
	return settings.Standard().MustParsePitch(text)
}

// at returns the pitch with the spelling s and the Octave whose ASPN number is
// o, for pitches whose Octave is named explicitly in business-spec 6.1.
func at(s spelling.Spelling, o int) pitch.Pitch {
	return pitch.MustFromASPN(s, o)
}

// cHalf4 is C with the alteration +1/2 and Octave 4: inadmissible in the
// traditional system, admissible in 24-quarter-tone (КП-7).
func cHalf4() pitch.Pitch {
	return at(spelling.MustNew(letter.C, alteration.MustFromFraction(1, 2)), 4)
}

func trad() *system.System  { return system.Traditional() }
func pos19() *system.System { return system.Positions19() }
func qt24() *system.System  { return system.QuarterTones24() }

// h0 is Н-0: the standard settings.
func h0() *settings.Settings { return settings.Standard() }

// h415 is Н-415: Н-0 with the reference frequency 415 Hz.
func h415() *settings.Settings {
	return settings.Standard().MustWithTuning(tuning.Standard().MustWithReferenceFrequency(415))
}

// hDe is Н-нем: Н-0 with the notation german-am.
func hDe() *settings.Settings {
	return settings.Standard().MustWithNotation(notation.MustBuiltIn(notation.GermanAMID))
}

// tuning19 is the tuning of Н-19: 19-position, 19-EDO, root C, A4 = 440 Hz.
func tuning19() *tuning.Tuning {
	return tuning.MustNew(system.Positions19(), temperament.EDO19(), tuning.RootFromName(spelling.C()), tuning.A4At440())
}

// h19 is Н-19: the tuning19 with the designations of Н-0.
func h19() *settings.Settings {
	return settings.Standard().MustWithTuning(tuning19())
}

// h24 is Н-24: 24-quarter-tone, 24-EDO, root C, A4 = 440 Hz, the notation
// lilypond-english and the other designations of Н-0.
func h24() *settings.Settings {
	t := tuning.MustNew(system.QuarterTones24(), temperament.EDO24(), tuning.RootFromName(spelling.C()), tuning.A4At440())
	return settings.MustNew(t, notation.MustBuiltIn(notation.LilyPondEnglishID), number.ASPN(), naming.English())
}

// rangeKP6 is the range of КП-6: from B♯ with Octave 4 to C♯ with Octave 4.
// It is valid in traditional and inverted in 19-position.
func rangeKP6() instrument.Range {
	return instrument.MustNewRange(at(spelling.BSharp(), 4), at(spelling.CSharp(), 4))
}

// rangeKP6a is the range of КП-6а: from D♭4 to C♯4. It is valid in
// traditional and inverted in 19-position.
func rangeKP6a() instrument.Range {
	return instrument.MustNewRange(p("D♭4"), p("C♯4"))
}

// rangeKP7 is the range of КП-7: from C4 to C with the alteration +1/2 and
// Octave 4. Its upper bound is inadmissible in traditional.
func rangeKP7() instrument.Range {
	return instrument.MustNewRange(p("C4"), cHalf4())
}

// rng returns the range between two pitches written by the standard settings.
func rng(lower, upper string) instrument.Range {
	return instrument.MustNewRange(p(lower), p(upper))
}

// preset returns a new preset with the given name, set and range.
func preset(name string, set *settings.Settings, r instrument.Range) *instrument.Preset {
	return instrument.MustNewPreset(name, set, r)
}

// sentinels are the ten own causes of the package, Е-1…Е-10.
var sentinels = []error{
	instrument.ErrUnset,
	instrument.ErrEmptyName,
	instrument.ErrUnsetPitch,
	instrument.ErrInadmissible,
	instrument.ErrInverted,
	instrument.ErrIncompatible,
	instrument.ErrNoRange,
	instrument.ErrDuplicateName,
	instrument.ErrNotFound,
	instrument.ErrNoPitches,
}

// assertCause checks that err matches the own cause want and none of the other
// own causes (К-23). For Е-4 and Е-6 it also checks the cause of the
// dependency that the contract adds (А 4.3).
func assertCause(t *testing.T, err, want error) bool {
	t.Helper()
	if !assert.Error(t, err) {
		return false
	}
	ok := assert.ErrorIs(t, err, want)
	for _, s := range sentinels {
		if s != want {
			ok = assert.NotErrorIs(t, err, s, "unexpected own cause") && ok
		}
	}
	switch want {
	case instrument.ErrInadmissible:
		ok = assert.ErrorIs(t, err, system.ErrInadmissible) && ok
	case instrument.ErrIncompatible:
		ok = assert.ErrorIs(t, err, settings.ErrIncompatible) && ok
		_, isIncompatible := errors.AsType[*instrument.IncompatibleError](err)
		ok = assert.True(t, isIncompatible, "Е-6 is an *IncompatibleError") && ok
	default:
		ok = assert.NotErrorIs(t, err, system.ErrInadmissible) && ok
		ok = assert.NotErrorIs(t, err, settings.ErrIncompatible) && ok
	}
	return ok
}

// assertDependencyCause checks an error Е-11: it matches the cause of a
// dependency and none of the own causes of the package (А 4.3).
func assertDependencyCause(t *testing.T, err, dep error) bool {
	t.Helper()
	if !assert.Error(t, err) {
		return false
	}
	ok := assert.ErrorIs(t, err, dep)
	for _, s := range sentinels {
		ok = assert.NotErrorIs(t, err, s, "Е-11 matches no own cause") && ok
	}
	return ok
}

// causeOf returns the own cause that err matches, or nil.
func causeOf(err error) error {
	for _, s := range sentinels {
		if errors.Is(err, s) {
			return s
		}
	}
	return nil
}

// assertSameAnswer checks that two errors give the same answer: both nil, or
// both with the same own cause and the same causes of dependencies.
func assertSameAnswer(t *testing.T, want, got error) {
	t.Helper()
	if want == nil {
		assert.NoError(t, got)
		return
	}
	require.Error(t, got)
	assert.Equal(t, causeOf(want), causeOf(got), "own cause")
	for _, dep := range []error{
		settings.ErrIncompatible, settings.ErrUnset, settings.ErrInconsistent,
		system.ErrInadmissible, system.ErrInvalid, system.ErrOutOfRange,
		tuning.ErrClassCount, tuning.ErrInadmissible,
	} {
		assert.Equal(t, errors.Is(want, dep), errors.Is(got, dep), "dependency cause %v", dep)
	}
}

// recovered calls f and returns the value of its panic, if any.
func recovered(f func()) (value any, panicked bool) {
	panicked = true
	defer func() {
		if panicked {
			value = recover()
		}
	}()
	f()
	panicked = false
	return nil, false
}

// assertPanicsWith checks that f panics with an error whose own cause is want
// and that has the text of the error of the original operation (dev.md
// «Must»: the variant panics with the error of the operation it wraps).
func assertPanicsWith(t *testing.T, want, original error, f func()) {
	t.Helper()
	v, panicked := recovered(f)
	if !assert.True(t, panicked, "the Must variant panics") {
		return
	}
	err, isErr := v.(error)
	if !assert.True(t, isErr, "the panic value is an error, got %T", v) {
		return
	}
	assert.ErrorIs(t, err, want)
	if original != nil {
		assert.EqualError(t, err, original.Error())
	}
}

// assertNoPanic checks that f does not panic.
func assertNoPanic(t *testing.T, f func()) {
	t.Helper()
	v, panicked := recovered(f)
	assert.False(t, panicked, "unexpected panic: %v", v)
}

// presetOf builds the expected preset of a state for comparisons by Equal.
func presetOf(t *testing.T, name string, set *settings.Settings, r instrument.Range) *instrument.Preset {
	t.Helper()
	ps, err := instrument.NewPreset(name, set, r)
	require.NoError(t, err)
	return ps
}

// assertState checks the state of g against the expected preset: the name,
// the set by Equal and the range by Equal, read both one by one and as one
// state.
func assertState(t *testing.T, want *instrument.Preset, g *instrument.Instrument) {
	t.Helper()
	assert.Equal(t, want.Name(), g.Name(), "name")
	assert.True(t, want.SettingsSet().Equal(g.Settings()), "set: want %v, got %v", want.SettingsSet(), g.Settings())
	assert.Equal(t, want.HasRange(), g.HasRange(), "has range")
	assert.True(t, want.Range().Equal(g.Range()), "range: want %v, got %v", want.Range(), g.Range())
	st, err := g.State()
	if assert.NoError(t, err) {
		assert.True(t, want.Equal(st), "state: want %v, got %v", want, st)
	}
}
