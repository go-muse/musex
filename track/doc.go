// Package track records an instrument-assigned part in a finite measure grid.
//
// A Track is an immutable snapshot with parallel voices, original events,
// barline fermatas and an optional measured clock. Edit and Splice publish
// only valid complete records. Sounds follows ties by note occurrence;
// Entries and At preserve notation, explicit rests and unrecorded silence.
//
// Layout and transformation operations return a Proposal containing a new
// snapshot and independent before/after records. ExtractLine and WriteLine
// bridge notation and the independent melodic lines of package voice.
// Instrument assignment never rebinds notes to a different settings source.
//
// Nil and zero tracks are unset. Concurrent reads and derived edits are safe;
// the assigned instrument and immutable events remain intentionally shared.
package track
