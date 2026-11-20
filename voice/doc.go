// Package voice provides the melodic line: a voice as a line of the texture of
// a polyphonic work, such as the soprano of a chorale or the subject of a
// fugue.
//
// A Line is not the voice of a track in the sense of MusicXML, which is a
// layer of the notation in measures: that notion belongs to package track.
// The names of the roles by place coincide with presets of package
// instrument, such as "Soprano Voice", but a role is not a preset.
//
// # Elements
//
// A line is monophonic. Its elements follow each other in order, and each
// is a sound with one note with an octave or a rest, with a Duration and an
// optional fermata sign. Chords, grace notes and rests in measures are never
// elements: FromEvents reports them as ErrChord, ErrGrace and
// ErrMeasureDuration.
//
// A Duration has one of three forms, kept as given: a written duration, a
// value without notation, or a tied duration of two or more written parts.
// A rule of the product: a sound under a tie is one element with a tied
// duration, not several elements with tie signs as in the voice of a track.
// Events converts it to tied events and FromEvents back.
//
//	std := settings.Standard()
//	c4 := note.MustFromASPN(spelling.C(), 4, cents.Zero(), std)
//	q := voice.MustWrittenDuration(duration.MustNewWritten(duration.Quarter, 0))
//	line := voice.MustNew(voice.MustNewSound(c4, q), voice.MustNewRest(q)).
//		MustWithName("Melody").MustWithPlacement(voice.Soprano)
//
// # Positions without measures
//
// Positions are exact offsets of duration.Position from the start of the
// line, without measures: the start of an element is the sum of the lengths
// of the previous elements. A moment at the end of an element belongs to the
// next element, and a moment at or after the end of the line has no element:
// At reports false, which is a result, not an error. Starts are computed when
// a line is created, so StartOf is O(1) and At is O(log n).
//
// # Seconds
//
// A rule of the product: the seconds of the sounds and the elements are taken
// from a tempo map of package tempomap addressed by positions, with the
// position of the start of the line on the map and an explicit number of
// decimal places and rounding mode. A fermata of an element adds no hold by
// itself: the hold is an extension of the map at the end of the element,
// whose position Fermatas gives. Approximate seconds of gradual tempo
// changes keep the approximation flag and the error bound of tempomap.
//
// # Pitch and the system of an operation
//
// Melodic intervals, the range, the extremes, the ambitus, the locations and
// the tessitura compare the written pitches of the sounds. A rule of the
// product: an interval between two sounds separated by a rest is reported too,
// with the flag ThroughRest.
//
// Every such operation works in one system. The forms with the suffix In take
// it explicitly and read no source for it. Without it, a reference range
// given by a preset or an instrument gives its system, and bound sounds must
// have that system (ErrDifferentSystems), while unbound sounds are allowed.
// Otherwise every sound must be bound (ErrNoSystem) and all of them must have
// one system (ErrDifferentSystems). The building operations (transposition,
// doubling, imitation, inversion) move each sound in the system of its own
// settings, or in the explicit system; sounds of different systems are then
// allowed, but an unbound sound without an explicit system reports
// ErrNoSystem.
//
// # Tessitura
//
// A rule of the product: the tessitura of a line relative to a reference
// range is an exact number, the mean position of its sounds weighted by
// their lengths or by one, where the lower bound of the range is 0 and the
// upper bound is 1, and a kind given by zones. The default zones 1/3 and 2/3
// are a rule of the product, not a fact of music theory, and can be replaced
// by NewZones. The kind is the mean position of the sounds: a middle
// tessitura does not mean that the line fits in the range; use Within or
// Locate for that.
//
// # Building lines
//
// Transpose, Double, Imitate, ImitateAfter, Invert, InvertPitch and
// Retrograde build new lines. Doubling, imitation and inversion give lines
// with the thematic roles Doubling, Imitating and Inversion, an empty name and
// no other roles; transposition and retrograde keep the name and the roles.
// The constructions are exact, by intervals, not by the degrees of a mode.
//
// # Values, ownership and concurrency
//
// The zero value of a line, nil or &Line{}, is the unset line, not the empty
// line, which New() creates. All types are immutable: the operations that
// change or build a line return a new *Line, and the receiver and the
// arguments never change, on success or failure. Reads return copies, and a
// line never exposes references to its data, so all methods are safe for
// concurrent use. The immutability of the notes is shallow: a note keeps the
// reference to its source of settings, which operations that need a system
// read at the moment of the call.
//
// # Errors
//
// Every error matches at most one sentinel of the package with errors.Is,
// and the causes of the dependencies are kept. A failure caused by one
// element of an input list or of the line is an *ItemError with its index:
//
//	if ie, ok := errors.AsType[*voice.ItemError](err); ok {
//		fmt.Println(ie.Index)
//	}
//
// Every function and method with an error has a Must variant that panics
// with the same error.
package voice
