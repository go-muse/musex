package instrument

import (
	"iter"
	"sync"
	"sync/atomic"

	"github.com/go-muse/musex/notation"
	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/octave/number"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"
)

// Instrument is an instance of an instrument: an object with identity that
// owns its current set of settings, gives it as a settings.Source and replaces
// it whole after a check. It also has a name and a reference range, which can
// be absent.
//
// The state of an instrument — the name, the set and the range — is always
// correct: the name is not empty, the set is set, and the range is absent or
// valid in the system of the set. An operation that would break this fails and
// changes nothing.
//
// An Instrument is used through a pointer. It holds a mutex and must not be
// copied; Clone makes an independent copy. The nil pointer and a pointer to the
// zero value are the unset instrument: IsValid reports false, Name returns "",
// Range returns the unset Range, Settings returns the unset set (nil), String
// returns "Instrument(unset)", Same and Equal follow their rules, and every
// other operation fails with ErrUnset. The unset instrument never becomes set.
//
// All methods are safe for concurrent use without external synchronization,
// with one restriction: a transformation given to RetuneFunc must not call a
// changing method of any instrument (see RetuneFunc).
type Instrument struct {
	// mu serializes the changing methods.
	mu sync.Mutex
	// state is the current snapshot; it is never changed after it is stored.
	state atomic.Pointer[Preset]
}

var _ settings.Source = (*Instrument)(nil)

// New returns a new instrument with the name, the set and the range of p.
// It keeps its own copy of the value of p: assigning through p later does not
// change the instrument. It fails with ErrUnset when p is unset.
func New(p *Preset) (*Instrument, error) {
	if !p.IsValid() {
		return nil, unsetError("new instrument", "preset")
	}

	return newInstrument(*p), nil
}

// MustNew is like New but panics with its error on failure.
func MustNew(p *Preset) *Instrument {
	return must(New(p))
}

// FromParts returns a new instrument with a name, a set of settings and a
// range; the unset Range gives an instrument without a range. It is like
// New(NewPreset(name, set, r)) and fails with the errors of NewPreset.
func FromParts(name string, set *settings.Settings, r Range) (*Instrument, error) {
	p, err := NewPreset(name, set, r)
	if err != nil {
		return nil, err
	}

	return newInstrument(*p), nil
}

// MustFromParts is like FromParts but panics with its error on failure.
func MustFromParts(name string, set *settings.Settings, r Range) *Instrument {
	return must(FromParts(name, set, r))
}

// IsValid reports whether i is set.
func (i *Instrument) IsValid() bool {
	return i.load() != nil
}

// Name returns the current name; "" for the unset instrument.
func (i *Instrument) Name() string {
	s := i.load()
	if s == nil {
		return ""
	}

	return s.name
}

// Range returns the current range; the unset Range when the instrument has no
// range or is unset.
func (i *Instrument) Range() Range {
	s := i.load()
	if s == nil {
		return Range{}
	}

	return s.rng
}

// HasRange reports whether the instrument currently has a range.
func (i *Instrument) HasRange() bool {
	return i.Range().IsValid()
}

// Settings returns the current set of settings whole: the same pointer that
// the instrument holds, without allocation and without waiting for a change in
// progress. The unset instrument returns the unset set (nil). It implements
// settings.Source. The set must not be overwritten through the pointer.
//
// Name, Range, HasRange and Settings are separate reads: two calls in a row
// can see different states. State gives the state of one moment whole.
func (i *Instrument) Settings() *settings.Settings {
	s := i.load()
	if s == nil {
		return nil
	}

	return s.set
}

// State returns the state of one moment as a new preset: the name, the set and
// the range. Assigning through the result does not change the instrument.
// It fails with ErrUnset when i is unset.
func (i *Instrument) State() (*Preset, error) {
	s := i.load()
	if s == nil {
		return nil, unsetError("state", "instrument")
	}

	st := *s

	return &st, nil
}

// MustState is like State but panics with its error on failure.
func (i *Instrument) MustState() *Preset {
	return must(i.State())
}

// Clone returns a new instrument with the state of one moment of i. The copy
// is independent: changing one does not change the other, and they are
// different sources. It fails with ErrUnset when i is unset.
func (i *Instrument) Clone() (*Instrument, error) {
	s := i.load()
	if s == nil {
		return nil, unsetError("clone", "instrument")
	}

	c := &Instrument{}
	c.state.Store(s)

	return c, nil
}

// MustClone is like Clone but panics with its error on failure.
func (i *Instrument) MustClone() *Instrument {
	return must(i.Clone())
}

// Same reports whether i and other are the same instance: i == other. All nil
// pointers are one unset instance, and every separate Instrument{} is its own
// instance. The answer always agrees with settings.SameSource(i, other).
// Retuning does not change identity.
func (i *Instrument) Same(other *Instrument) bool {
	return i == other
}

// Equal reports whether the states of i and other are equal as presets
// (Preset.Equal): two unset instruments are equal, an unset and a set one are
// not. When i == other it is true without reading the state. Otherwise each
// state is read once, each at its own moment. Equal states do not make two
// instruments one source; see Same.
func (i *Instrument) Equal(other *Instrument) bool {
	if i == other {
		return true
	}

	return i.load().Equal(other.load())
}

// String returns a debugging form with the name, the set and the range or
// "no range" of one moment, or "Instrument(unset)". It is not a notation, has
// no parser and does not affect equality.
func (i *Instrument) String() string {
	s := i.load()
	if s == nil {
		return "Instrument(unset)"
	}

	return "Instrument(" + s.name + ", " + s.set.String() + ", " + rangeString(s.rng) + ")"
}

// CheckRetune reports whether Retune(set) would replace the set now: nil when
// it would, otherwise the same error with the same cause. It does not change
// the instrument and does not wait for changes in progress, so its answer can
// become stale when other goroutines change the instrument.
func (i *Instrument) CheckRetune(set *settings.Settings) error {
	const op = "check retune"

	s := i.load()
	if s == nil {
		return unsetError(op, "instrument")
	}

	_, err := retuned(op, s, set)

	return err
}

// MustCheckRetune is like CheckRetune but panics with its error on failure;
// it returns i.
func (i *Instrument) MustCheckRetune(set *settings.Settings) *Instrument {
	return i.mustDo(i.CheckRetune(set))
}

// CheckRetuneWithRange reports whether RetuneWithRange(set, r) would replace
// the set and the range now, as CheckRetune does for Retune.
func (i *Instrument) CheckRetuneWithRange(set *settings.Settings, r Range) error {
	const op = "check retune with range"

	s := i.load()
	if s == nil {
		return unsetError(op, "instrument")
	}

	_, err := retunedWithRange(op, s, set, r)

	return err
}

// MustCheckRetuneWithRange is like CheckRetuneWithRange but panics with its
// error on failure; it returns i.
func (i *Instrument) MustCheckRetuneWithRange(set *settings.Settings, r Range) *Instrument {
	return i.mustDo(i.CheckRetuneWithRange(set, r))
}

// Contains reports whether x is in the range of the instrument in the system
// of its current set, as Range.Contains does; the range and the set are read
// from one moment.
//
// Errors, in this order: ErrUnset when i is unset; ErrNoRange when it has no
// range; ErrUnsetPitch when x is unset; ErrInadmissible when the name of x is
// not admissible in the system.
func (i *Instrument) Contains(x pitch.Pitch) (bool, error) {
	s := i.load()
	if s == nil {
		return false, unsetError("contains", "instrument")
	}

	return s.rng.Contains(s.set.System(), x)
}

// MustContains is like Contains but panics with its error on failure.
func (i *Instrument) MustContains(x pitch.Pitch) bool {
	return must(i.Contains(x))
}

// Locate returns where x is relative to the range of the instrument, as
// Range.Locate does. Errors are those of Contains.
func (i *Instrument) Locate(x pitch.Pitch) (Location, error) {
	s := i.load()
	if s == nil {
		return LocationUnset, unsetError("locate", "instrument")
	}

	return s.rng.Locate(s.set.System(), x)
}

// MustLocate is like Locate but panics with its error on failure.
func (i *Instrument) MustLocate(x pitch.Pitch) Location {
	return must(i.Locate(x))
}

// Width returns the width of the range of the instrument, as Range.Width does.
// Errors, in this order: ErrUnset; ErrNoRange; system.ErrOutOfRange.
func (i *Instrument) Width() (system.Step, error) {
	s := i.load()
	if s == nil {
		return 0, unsetError("width", "instrument")
	}

	return s.rng.Width(s.set.System())
}

// MustWidth is like Width but panics with its error on failure.
func (i *Instrument) MustWidth() system.Step {
	return must(i.Width())
}

// Count returns the number of sounds of the range of the instrument, as
// Range.Count does. Errors are those of Width.
func (i *Instrument) Count() (int, error) {
	s := i.load()
	if s == nil {
		return 0, unsetError("count", "instrument")
	}

	return s.rng.Count(s.set.System())
}

// MustCount is like Count but panics with its error on failure.
func (i *Instrument) MustCount() int {
	return must(i.Count())
}

// Folds returns all folds of x into the range of the instrument, as
// Range.Folds does. The sequence uses the range and the system of the moment
// of the call; a later retuning does not change it. Errors are those of
// Contains.
func (i *Instrument) Folds(x pitch.Pitch) (iter.Seq[pitch.Pitch], error) {
	s := i.load()
	if s == nil {
		return nil, unsetError("folds", "instrument")
	}

	return s.rng.Folds(s.set.System(), x)
}

// MustFolds is like Folds but panics with its error on failure.
func (i *Instrument) MustFolds(x pitch.Pitch) iter.Seq[pitch.Pitch] {
	return must(i.Folds(x))
}

// Fold returns the nearest fold of x into the range of the instrument, as
// Range.Fold does. Errors are those of Contains.
func (i *Instrument) Fold(x pitch.Pitch) (pitch.Pitch, bool, error) {
	s := i.load()
	if s == nil {
		return pitch.Pitch{}, false, unsetError("fold", "instrument")
	}

	return s.rng.Fold(s.set.System(), x)
}

// MustFold is like Fold but panics with its error on failure.
func (i *Instrument) MustFold(x pitch.Pitch) (pitch.Pitch, bool) {
	p, ok, err := i.Fold(x)
	if err != nil {
		panic(err)
	}

	return p, ok
}

// Pitches returns the pitches of the range of the instrument by rule, as
// Range.Pitches does. The sequence uses the range and the system of the moment
// of the call. Errors, in this order: ErrUnset; ErrNoRange; the errors of
// system.System.Walk.
func (i *Instrument) Pitches(rule system.SpellingRule) (iter.Seq[pitch.Pitch], error) {
	s := i.load()
	if s == nil {
		return nil, unsetError("pitches", "instrument")
	}

	return s.rng.Pitches(s.set.System(), rule)
}

// MustPitches is like Pitches but panics with its error on failure.
func (i *Instrument) MustPitches(rule system.SpellingRule) iter.Seq[pitch.Pitch] {
	return must(i.Pitches(rule))
}

// Retune replaces the current set of settings by set whole; the name and the
// range stay. It succeeds when both conditions hold:
//   - set is compatible with the current set by settings.Settings.CheckReplacement;
//   - the range, if there is one, is valid in the system of set.
//
// Errors, in this order: ErrUnset when i is unset, then when set is unset;
// ErrIncompatible as an *IncompatibleError with the report; ErrInverted when
// the range would be inverted in the new system. On failure the state stays
// whole. A set equal to the current one is compatible. After success Settings
// returns the pointer set.
func (i *Instrument) Retune(set *settings.Settings) error {
	const op = "retune"

	return i.change(op, func(cur *Preset) (*Preset, error) {
		return retuned(op, cur, set)
	})
}

// MustRetune is like Retune but panics with its error on failure; it returns i.
func (i *Instrument) MustRetune(set *settings.Settings) *Instrument {
	return i.mustDo(i.Retune(set))
}

// RetuneWithRange replaces the set by set and the range by r in one action;
// the unset Range leaves the instrument without a range. The current range
// does not matter. A reader sees either the previous set and range or the new
// ones, never a mixture and never a state without a range in between.
//
// Errors, in this order: ErrUnset when i is unset, then when set is unset;
// ErrIncompatible; for a set range in the system of set, ErrInadmissible for
// the lower bound, then for the upper bound, then ErrInverted. On failure the
// state stays whole.
func (i *Instrument) RetuneWithRange(set *settings.Settings, r Range) error {
	const op = "retune with range"

	return i.change(op, func(cur *Preset) (*Preset, error) {
		return retunedWithRange(op, cur, set, r)
	})
}

// MustRetuneWithRange is like RetuneWithRange but panics with its error on
// failure; it returns i.
func (i *Instrument) MustRetuneWithRange(set *settings.Settings, r Range) *Instrument {
	return i.mustDo(i.RetuneWithRange(set, r))
}

// RetuneFunc replaces the current set by the result of f, as Retune does,
// in one indivisible action: f receives the current set (the pointer that
// Settings returns), and no other change happens between that call and the
// replacement. Concurrent calls of RetuneFunc lose no change: each f receives
// the result of the previous one.
//
// f is called exactly once, synchronously, in the calling goroutine, while the
// mutex of the instrument is held. Inside f it is allowed to read this and any
// other instrument: readers do not take the mutex. Inside f it is forbidden to
// call a changing method (the Retune methods, SetName, SetRange, RemoveRange)
// of any instrument, directly or through other calls. Doing so is a deadlock:
// for the same instrument the call never returns, because f waits for the
// mutex held by its own call; for another instrument, when two goroutines do
// it for each other's instruments. When all goroutines wait, the Go runtime
// stops the program with "all goroutines are asleep - deadlock!". The violation
// cannot be detected reliably. To change several instruments, read what is
// needed, then call the changing methods one after another outside f. A panic
// of f is not recovered: the mutex is released, the state stays and the panic
// goes on. f must not keep the received pointer to overwrite the set.
//
// Errors, in this order: ErrUnset when i is unset, then when f is nil; the
// error of f, with its cause (the result of f is then not used); then those of
// Retune for the result, ErrUnset when f returns an unset set.
func (i *Instrument) RetuneFunc(f func(*settings.Settings) (*settings.Settings, error)) error {
	const op = "retune func"

	return i.change(op, func(cur *Preset) (*Preset, error) {
		if f == nil {
			return nil, unsetError(op, "transformation")
		}

		next, err := f(cur.set)
		if err != nil {
			return nil, wrapError(op, err)
		}

		return retuned(op, cur, next)
	})
}

// MustRetuneFunc is like RetuneFunc but panics with its error on failure;
// it returns i.
func (i *Instrument) MustRetuneFunc(f func(*settings.Settings) (*settings.Settings, error)) *Instrument {
	return i.mustDo(i.RetuneFunc(f))
}

// RetuneTuning replaces the tuning of the current set, as RetuneFunc does with
// the transformation settings.Settings.WithTuning(t). An unset t is reported
// by package settings (settings.ErrUnset), not by ErrUnset of this package,
// which reports only an unset instrument. Errors, in this order: ErrUnset;
// the error of settings; ErrIncompatible; ErrInverted.
func (i *Instrument) RetuneTuning(t *tuning.Tuning) error {
	return i.retunePart("retune tuning", func(s *settings.Settings) (*settings.Settings, error) {
		return s.WithTuning(t)
	})
}

// MustRetuneTuning is like RetuneTuning but panics with its error on failure;
// it returns i.
func (i *Instrument) MustRetuneTuning(t *tuning.Tuning) *Instrument {
	return i.mustDo(i.RetuneTuning(t))
}

// RetuneSystem replaces the system of the current set, as RetuneTuning does
// with settings.Settings.WithSystem(sys): the tuning is translated to sys with
// the same number of classes; another number of classes fails with the error
// of package tuning (tuning.ErrClassCount).
func (i *Instrument) RetuneSystem(sys *system.System) error {
	return i.retunePart("retune system", func(s *settings.Settings) (*settings.Settings, error) {
		return s.WithSystem(sys)
	})
}

// MustRetuneSystem is like RetuneSystem but panics with its error on failure;
// it returns i.
func (i *Instrument) MustRetuneSystem(sys *system.System) *Instrument {
	return i.mustDo(i.RetuneSystem(sys))
}

// RetuneNotation replaces the notation of the current set, as RetuneTuning
// does with settings.Settings.WithNotation(n).
func (i *Instrument) RetuneNotation(n notation.Notation) error {
	return i.retunePart("retune notation", func(s *settings.Settings) (*settings.Settings, error) {
		return s.WithNotation(n)
	})
}

// MustRetuneNotation is like RetuneNotation but panics with its error on
// failure; it returns i.
func (i *Instrument) MustRetuneNotation(n notation.Notation) *Instrument {
	return i.mustDo(i.RetuneNotation(n))
}

// RetuneNumbering replaces the octave numbering of the current set, as
// RetuneTuning does with settings.Settings.WithNumbering(num).
func (i *Instrument) RetuneNumbering(num number.Numbering) error {
	return i.retunePart("retune numbering", func(s *settings.Settings) (*settings.Settings, error) {
		return s.WithNumbering(num)
	})
}

// MustRetuneNumbering is like RetuneNumbering but panics with its error on
// failure; it returns i.
func (i *Instrument) MustRetuneNumbering(num number.Numbering) *Instrument {
	return i.mustDo(i.RetuneNumbering(num))
}

// RetuneCatalog replaces the catalog of octave names of the current set, as
// RetuneTuning does with settings.Settings.WithCatalog(c); nil is the empty
// catalog. It is not the Catalog of presets of this package.
func (i *Instrument) RetuneCatalog(c *naming.Catalog) error {
	return i.retunePart("retune catalog", func(s *settings.Settings) (*settings.Settings, error) {
		return s.WithCatalog(c)
	})
}

// MustRetuneCatalog is like RetuneCatalog but panics with its error on
// failure; it returns i.
func (i *Instrument) MustRetuneCatalog(c *naming.Catalog) *Instrument {
	return i.mustDo(i.RetuneCatalog(c))
}

// RetuneDesignations replaces all designations of the current set — the
// notation, the octave numbering and the catalog of octave names — as
// RetuneTuning does with settings.Settings.WithDesignations(n, num, c).
func (i *Instrument) RetuneDesignations(n notation.Notation, num number.Numbering, c *naming.Catalog) error {
	return i.retunePart("retune designations", func(s *settings.Settings) (*settings.Settings, error) {
		return s.WithDesignations(n, num, c)
	})
}

// MustRetuneDesignations is like RetuneDesignations but panics with its error
// on failure; it returns i.
func (i *Instrument) MustRetuneDesignations(n notation.Notation, num number.Numbering, c *naming.Catalog) *Instrument {
	return i.mustDo(i.RetuneDesignations(n, num, c))
}

// SetName replaces the name. Errors, in this order: ErrUnset when i is unset;
// ErrEmptyName.
func (i *Instrument) SetName(name string) error {
	const op = "set name"

	return i.change(op, func(cur *Preset) (*Preset, error) {
		if name == "" {
			return nil, emptyNameError(op)
		}

		return &Preset{name: name, set: cur.set, rng: cur.rng}, nil
	})
}

// MustSetName is like SetName but panics with its error on failure; it
// returns i.
func (i *Instrument) MustSetName(name string) *Instrument {
	return i.mustDo(i.SetName(name))
}

// SetRange replaces the range; the unset Range removes it, as RemoveRange
// does. The range is checked in the system of the current set; the set is not
// checked for compatibility.
//
// Errors, in this order: ErrUnset when i is unset; ErrInadmissible for the
// lower bound, then for the upper bound; ErrInverted.
func (i *Instrument) SetRange(r Range) error {
	const op = "set range"

	return i.change(op, func(cur *Preset) (*Preset, error) {
		if err := rangeIn(op, cur.set, r); err != nil {
			return nil, err
		}

		return &Preset{name: cur.name, set: cur.set, rng: r}, nil
	})
}

// MustSetRange is like SetRange but panics with its error on failure; it
// returns i.
func (i *Instrument) MustSetRange(r Range) *Instrument {
	return i.mustDo(i.SetRange(r))
}

// RemoveRange leaves the instrument without a range; removing again succeeds.
// It fails with ErrUnset when i is unset.
func (i *Instrument) RemoveRange() error {
	return i.change("remove range", func(cur *Preset) (*Preset, error) {
		return &Preset{name: cur.name, set: cur.set}, nil
	})
}

// MustRemoveRange is like RemoveRange but panics with its error on failure;
// it returns i.
func (i *Instrument) MustRemoveRange() *Instrument {
	return i.mustDo(i.RemoveRange())
}

// newInstrument returns an instrument with a copy of a correct preset as its
// first snapshot.
func newInstrument(p Preset) *Instrument {
	i := &Instrument{}
	i.state.Store(&p)

	return i
}

// load returns the current snapshot; nil for the unset instrument.
func (i *Instrument) load() *Preset {
	if i == nil {
		return nil
	}

	return i.state.Load()
}

// change runs one change: under the mutex it reads the current snapshot,
// builds the next one and stores it. A failure of build stores nothing, and
// the deferred unlock releases the mutex even when build panics.
func (i *Instrument) change(op string, build func(cur *Preset) (*Preset, error)) error {
	if i == nil {
		return unsetError(op, "instrument")
	}

	i.mu.Lock()
	defer i.mu.Unlock()

	cur := i.state.Load()
	if cur == nil {
		return unsetError(op, "instrument")
	}

	next, err := build(cur)
	if err != nil {
		return err
	}

	i.state.Store(next)

	return nil
}

// retunePart runs RetuneFunc with a replacement of one part of the set; its
// failure is an error of package settings with context.
func (i *Instrument) retunePart(op string, with func(*settings.Settings) (*settings.Settings, error)) error {
	return i.change(op, func(cur *Preset) (*Preset, error) {
		next, err := with(cur.set)
		if err != nil {
			return nil, wrapError(op, err)
		}

		return retuned(op, cur, next)
	})
}

// mustDo panics with err or returns i.
func (i *Instrument) mustDo(err error) *Instrument {
	if err != nil {
		panic(err)
	}

	return i
}

// retuned returns the snapshot of cur with set, or the reason why set cannot
// replace the current set: ErrUnset, ErrIncompatible, then the validity of the
// current range in the system of set.
func retuned(op string, cur *Preset, set *settings.Settings) (*Preset, error) {
	if err := compatible(op, cur, set); err != nil {
		return nil, err
	}

	if err := rangeIn(op, set, cur.rng); err != nil {
		return nil, err
	}

	return &Preset{name: cur.name, set: set, rng: cur.rng}, nil
}

// retunedWithRange returns the snapshot of cur with set and r, or the reason:
// ErrUnset, ErrIncompatible, then the validity of r in the system of set.
func retunedWithRange(op string, cur *Preset, set *settings.Settings, r Range) (*Preset, error) {
	if err := compatible(op, cur, set); err != nil {
		return nil, err
	}

	if err := rangeIn(op, set, r); err != nil {
		return nil, err
	}

	return &Preset{name: cur.name, set: set, rng: r}, nil
}

// compatible checks that set is set and can replace the set of cur.
func compatible(op string, cur *Preset, set *settings.Settings) error {
	if !set.IsValid() {
		return unsetError(op, "settings")
	}

	rep, err := cur.set.CheckReplacement(set)
	if err != nil {
		return wrapError(op, err)
	}

	if !rep.Compatible() {
		return &IncompatibleError{Report: rep}
	}

	return nil
}
