// Package event represents a single written element of a voice: a note,
// a nonempty chord, or a rest. Use [New] to combine its [Content], [Duration],
// and [Marks]. The resulting [Event] is immutable. Its With methods return
// new events, and changes to returned notes, chords, and slices leave the
// event's record intact. Note copies retain their settings sources, which
// the event package does not read.
//
// A duration can preserve a written value with dots and tuplets, hold an
// exact positive fraction of a whole note, or describe a rest in measures.
// [Event.Length] reports the occupied length in whole notes. A grace note or
// chord occupies zero and retains its written duration. For a rest in
// measures, the caller supplies the measure lengths.
//
// Ties mark individual note occurrences in written order, including
// repeated notes. A fermata records a hold, and an arpeggio records a sign
// on a chord. Their performance, the continuation of ties, and positions in
// a voice are resolved by the caller.
//
// The zero [Content] and [Duration], the zero [Event], and a nil *Event
// represent unset values. Read and replacement methods report errors for
// these values except for the kind, set, and record equality queries. Must
// functions and methods panic with the error their regular operation returns.
package event
