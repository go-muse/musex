// Package instrument provides the instrument: the owner and operator of a set
// of settings, with presets, catalogs of presets and a reference range.
//
// # Settings, instrument and note
//
// Three roles are separate. A set of settings (package settings) is an
// immutable value: a system, a tuning and designations. An Instrument owns
// its current set, gives it whole as a settings.Source and replaces it whole
// after a check. A note (package note) reads its source at the moment of each
// operation; it knows nothing of instruments, and this package knows nothing
// of notes: it does not hold, register or check them. An instrument does not
// compute frequencies, positions or texts either; packages tuning, system and
// settings do that through the set.
//
//	g := instrument.MustNew(instrument.Violin())
//	var src settings.Source = g // a note can hold g as its source
//	g.MustRetuneTuning(tuning.Standard().MustWithReferenceFrequency(415))
//	// src.Settings() now gives the set with A4 at 415 Hz.
//
// # Presets and catalogs
//
// A Preset is an immutable name, set of settings and range, which can be
// absent. A Catalog is an immutable ordered collection of presets with
// different names and exact lookup by name; there is no registry, and a
// preset of the caller works exactly as a built-in one. Both are used through
// pointers and never change: With methods return new values. A catalog keeps
// and returns its own copies of presets. A preset is not a source: its set is
// read with SettingsSet, and a note is tied to an instrument, not to a preset.
//
// The 39 built-in presets — strings, woodwinds, brass, keyboards, pitched
// percussion and six voices — have English names (the exact text of lookup),
// new standard settings and a range from the domain knowledge of the library.
// Each function, such as Violin or BassVoice, and Builtins return new values
// on every call, so the built-in presets cannot be changed. The range of a
// built-in preset is a sounding range: for transposing instruments, such as
// the clarinet in B-flat, the double bass or the xylophone, it lists sounding,
// not written, pitches. Transposition of written pitches is not provided.
//
// # Ranges
//
// A Range is a reference range of two pitches, both bounds included. It holds
// no system: every operation with positions takes the system explicitly, so
// the same pair can be valid in one system and inverted in another. Pitches
// are compared by position, never by written octave numbers: in the
// traditional system B♯ at Octave 4, written B♯3, is at the position of C4 and
// is in a range from C4, while C♭ at Octave 3, written C♭4, is below it.
// A correction in cents does not take part: the operations take pitches, not
// notes. The range is only a reference: a pitch outside it is still a valid
// pitch, and nothing is forbidden because of a range. A fold of a pitch into a
// range keeps its name and changes its Octave; an empty intersection and a
// missing fold are answers, not errors.
//
// # Instruments and retuning
//
// An Instrument is created from a preset (New) or from parts (FromParts) and
// does not remember the preset. Its state — the name, the set and the range —
// is always correct. The name and the range can be changed (SetName,
// SetRange, RemoveRange). The set is replaced whole by retuning: with a ready
// set (Retune), by a transformation of the current set (RetuneFunc), by one
// part of the set (RetuneTuning, RetuneSystem, RetuneNotation,
// RetuneNumbering, RetuneCatalog, RetuneDesignations) or with a new range in
// one action (RetuneWithRange). A retuning succeeds when the next set is
// compatible with the current one by settings.Settings.CheckReplacement and
// the range, if any, is valid in the next system; otherwise it fails and the
// state stays whole. There is no separate operation to change the reference
// note: replace the tuning. Notes are not taken into account, so a refusal can
// be false for the notes actually in use; a new instrument can still be
// created with the needed set. CheckRetune and CheckRetuneWithRange give the
// same answer without a change.
//
// Same reports whether two values are one instance and agrees with
// settings.SameSource; Equal compares states. Retuning does not change
// identity. Clone gives an independent instrument with the same state, and
// State gives the state of one moment as a preset.
//
// # Unset values
//
// The zero Range, the nil and zero Preset, the nil and zero Catalog and the
// nil and zero Instrument never panic. The unset range means "no range": in a
// preset or an instrument it is the absence of a range, and an operation of a
// range fails with ErrNoRange. The unset catalog is empty. The unset preset
// and the unset instrument read as empty values; the unset instrument gives
// the unset set as a source; their other operations fail with ErrUnset.
// Debugging strings name the parts of set values and mark unset ones:
// "Range(unset)", "Preset(unset)", "Catalog(empty)", "Instrument(unset)".
// Nothing is substituted silently: no standard settings for an unset set, no
// swapped bounds for an inverted range and no range from a name.
//
// # Errors
//
// Errors are classified with errors.Is; their texts are not part of the API.
// Each error with a cause of this package matches exactly one of ErrUnset,
// ErrEmptyName, ErrUnsetPitch, ErrInadmissible, ErrInverted,
// ErrIncompatible, ErrNoRange, ErrDuplicateName, ErrNotFound and
// ErrNoPitches; an error of a dependency (below) matches none of them unless
// its own cause is one. ErrInadmissible also matches system.ErrInadmissible. A refused retuning is an *IncompatibleError
// with the report of package settings; it matches ErrIncompatible and
// settings.ErrIncompatible:
//
//	if ie, ok := errors.AsType[*instrument.IncompatibleError](err); ok {
//		fmt.Println(ie.Report) // violated rules and counterexamples
//	}
//
// Errors of dependencies, such as system.ErrOutOfRange for a width that does
// not fit in int, the errors of system.System.Walk, of the With methods of
// settings and of a transformation, are returned with context and keep their
// causes. The first cause found is reported, in this order:
//  1. ErrUnset: the instrument, the preset, the set, the system or the
//     transformation, in the order of the inputs;
//  2. ErrEmptyName;
//  3. ErrNoRange;
//  4. the validity of a range in the system: ErrInadmissible for the lower
//     bound, then for the upper bound, then ErrInverted; for two ranges the
//     receiver first, then the other;
//  5. pitch arguments: ErrUnsetPitch, then ErrInadmissible; RangeOf reports
//     ErrNoPitches first and then checks the pitches in order;
//  6. catalogs: ErrUnset for the first unset preset, then ErrDuplicateName
//     for the first repeated name; ErrNotFound;
//  7. errors of dependencies.
//
// Retuning reports ErrUnset, then ErrIncompatible, then the validity of the
// range in the next system (ErrInadmissible, ErrInverted); a transformation
// fails first with its own error. On error a function returns zero results
// and changes nothing; inputs are never changed. Every function and method
// that can fail has a Must variant that panics with the same error, which
// errors.Is recognizes after recover; the Must variants of the changing
// methods return the same instrument, so they can be chained:
//
//	g.MustSetName("Lead").MustRetune(set)
//
// # Concurrency
//
// Ranges, presets and catalogs are immutable and safe for concurrent use.
// All methods of an Instrument are safe for concurrent use without external
// synchronization. Readers — Settings, Name, Range, HasRange, IsValid, State,
// Clone, Same, Equal, String, the checks and the operations of the range —
// take no lock and never wait for a change in progress, including its check of
// compatibility; a reader sees the previous state or the next one whole,
// never a mixture. Name, Range and Settings are separate reads; State gives
// one moment. Changing methods run one at a time; a check and the replacement
// it allows belong to one state, and concurrent RetuneFunc calls lose no
// change.
//
// A transformation given to RetuneFunc runs exactly once, under the mutex of
// the instrument. It may read any instrument, but it must not call a changing
// method of any instrument, this one or another: that deadlocks, always for
// this instrument and for two instruments when two goroutines do it for each
// other's. The violation cannot be detected; change several instruments one
// after another outside transformations instead.
//
// # Ownership and copies
//
// Presets and instruments hold the given *settings.Settings and return the
// same pointer, without copies, so reading the set of an instrument does not
// allocate. Never assign through such a pointer (*s = …): that would change
// the set for every holder without a check and race with readers. The same
// holds for *Preset and *Catalog pointers. An Instrument holds a mutex and
// must not be copied (go vet reports such copies); use Clone.
//
// # Cost
//
// Settings, IsValid, Name, Range, HasRange and Same of an instrument do not
// allocate; neither do Contains, Locate, Check and Equal of a range and
// Contains and Locate of an instrument on success for the built-in systems,
// nor Equal of instruments and presets with built-in parts. Errors allocate
// only on failure. State, Clone and catalog lookups allocate one value; each
// built-in function allocates a preset and standard settings. Sequences of
// folds and pitches are lazy.
package instrument
