// Package tempomap provides immutable tempo maps: exact tempo changes and
// explicit fermata holds at positions or measure addresses, conversions between
// seconds and musical time, and correctly rounded approximations of gradual
// tempo changes.
//
// A map starts with a main tempo. A main change replaces both the current and
// main tempo; a jump changes only the current tempo; a gradual change starts a
// deviation with an explicitly selected law; a tempo returns to the main tempo
// and requires a current deviation; Tempo I restores the initial tempo as main;
// and a metric modulation computes a new main tempo. The only gradual law is
// LinearSpeed, linear whole-notes-per-minute speed by position. This is a
// product law, not a quantitative rule supplied by music theory for every
// accelerando or ritardando. Unit notation is not interpolated: TempoAt chooses
// the unit of an exact equivalent marking explicitly.
//
// # Records and editing
//
// Map is immutable. WithChange, WithoutChange, WithExtension, and
// WithoutExtension return new maps and leave their inputs unchanged on success
// and failure. A change is selected by its address and Slot. SlotFirst is the
// initial main tempo and cannot be removed. SlotSecond permits exactly one
// initial pair, main tempo followed by gradual deviation at the same initial
// address. This pair starts a gradual segment at the origin and creates no
// zero-length constant segment. Other duplicate change addresses are rejected.
// Extensions form a separate ordered collection and may share an address with
// a change; an explicit zero hold is preserved by record equality but has no
// effect on the timeline.
//
// A final gradual change may remain unfinished for editing. Validate reports
// that state without treating it as malformed. Calculations on or after an
// unfinished gradual start fail with ErrUnfinished because neither its final
// speed nor its seconds are defined. Consecutive gradual changes and a gradual
// change immediately followed by modulation fail with ErrInvalidSequence.
//
// # Seconds and fermata holds
//
// Start returns seconds at the beginning of a moment, including earlier holds
// but not the hold at the moment itself. End additionally includes that hold.
// Between(from,to) is Start(to)-Start(from), so it includes the hold at from
// and excludes the hold at to. For sounding notes, use End at both boundaries:
// adjacent notes do not overlap when their shared position has a hold. Position
// and AddressAt perform the reverse conversion; any time inside a hold maps to
// its exact position. Holds do not move musical positions or the measure grid.
//
// Measure maps do not store a grid. Operations that need positions or address
// validation receive a measure.Grid for each call. ToPositions and ToMeasures
// convert the complete record without embedding a grid. Equal compares records;
// EqualTimeline compares speed functions and holds after conversion to
// positions, ignoring addressing and record distinctions that do not affect
// musical time.
//
// # Approximation
//
// Constant-speed conversions and gradual segments with equal endpoint speeds
// are exact rational values. Other gradual segments use logarithms and
// exponentials. Start, End, Between, Position, AddressAt, and Segment.Seconds
// evaluate the complete symbolic expression and round it once, never once per
// segment. The caller always supplies decimal places and one of the nine modes
// of package common/rounding; there is no default. SecondsResult and
// PositionResult preserve the rounded value, approximation flag, decimal
// places, and guaranteed error bound. Directed modes have bound 10^-D and
// nearest modes 10^-D/2. Exact results have zero bound and do not depend on D,
// although D and mode are still validated.
//
// The implementation uses exact big.Int and big.Rat arithmetic, directed
// big.Float intervals, bounded series, exact multiplicative zero detection, and
// a computable precision limit Pmax. No float64, NaN, infinity, or unbounded
// refinement is used. Values and bounds must fit common/fraction: D=76 still
// represents both approximation bounds, whereas an approximate operation with
// D>=77 fails with ErrOutOfRange even if its rounded value would fit. An exact
// operation remains valid at those precisions.
//
// # Text
//
// String, MarshalText, and AppendText emit canonical text; the corresponding
// Parse functions and UnmarshalText read it back. The grammar distinguishes
// positional addresses (p:) from measure addresses (m:), exact results (s: and
// p:) from approximate results (s~ and p~), and all change kinds and slots.
// Approximate text stores D and the error bound, not the rounding mode. Unset
// values print diagnostic placeholders which parsing rejects. State and Segment
// have diagnostic strings only. Parse errors quote at most the first 64 bytes
// of input without splitting a UTF-8 character and remain at most 512 bytes.
//
// # Unset values and errors
//
// The zero values Map{}, Address{}, Change{}, Extension{}, State{}, Segment{},
// SecondsResult{}, and PositionResult{} are unset, not empty or initial values.
// Only IsSet, Equal, Address.Compare, and String work on them without error;
// all other operations on them return ErrUnset. Enumerations use zero and
// unknown values as invalid, not defaults. Errors match one of eleven package
// sentinels with errors.Is: ErrUnset, ErrInvalidArgument, ErrNoInitialTempo,
// ErrDuplicateAddress, ErrInvalidAddress, ErrNoDeviation, ErrInvalidSequence,
// ErrUnfinished, ErrNegativeTime, ErrOutOfRange, and ErrInvalidText. Relevant
// sentinel causes of fraction, rounding, duration, measure, and tempo remain
// visible as well. Validation order follows receiver and argument order, then
// record, address, state, sequence, unfinished, time, and range checks as
// documented by the operations. Failing calls return zero results; AppendText
// leaves its destination prefix unchanged and UnmarshalText leaves its receiver
// unchanged. Every fallible operation other than text marshaling has a Must
// variant that panics with the same error value.
//
// # Concurrency and ownership
//
// Values have no mutable global state, caches, pools, or goroutines. All methods
// except UnmarshalText have value receivers and do not modify inputs. Maps may
// share immutable internal slices and indices; Changes, Extensions, Segments,
// and MarshalText return caller-owned slices. Concurrent reads are safe without
// synchronization. As usual, UnmarshalText must not write through a pointer
// while another goroutine reads the same value.
package tempomap
