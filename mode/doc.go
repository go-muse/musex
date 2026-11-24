// Package mode models open and cyclic musical modes in terms of their written
// intervals. A mode has degrees without attached notes. A caller supplies the
// starting note, spelling, musical system, or catalog when an operation needs
// that context.
//
// In an explicit musical system, StepsBetween and StepsBetweenFrom measure the
// signed distance in positions between two degree places of any cycles, and
// the offset methods of Positions list the degrees at an offset and walk the
// occupied offsets or every offset of a range in either direction.
//
// Modes can also be built from the numbers of positions of their adjacent
// intervals in a system. SuggestLetterPlan suggests the letter plan that a
// spelling rule of package system gives for a one-octave mode, and
// FromNumbersFromWithRule builds that mode directly; the caller can instead
// edit the suggested plan and pass it to FromNumbersFromWithPlan.
// OpenFromNumbers and OpenFromNumbersFrom build an open mode from numbers and
// the sizes of its degrees, and Mode.LetterPlan reads the plan of any mode,
// which rebuilds the mode together with its numbers. Repository.IdentifyNumbers
// and Repository.IdentifyNumbersFrom find catalog entries by their numbers
// without building a mode.
//
// Copying a Mode by assignment gives a value that can be changed independently.
// Mutating methods replace the receiver only after validating the new
// structure. Read-only calls may run concurrently with one another.
package mode
