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
// Copying a Mode by assignment gives a value that can be changed independently.
// Mutating methods replace the receiver only after validating the new
// structure. Read-only calls may run concurrently with one another.
package mode
