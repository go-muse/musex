// Package modal analyzes the relative and absolute modal positions of degrees
// in a mode. Relative positions classify the full specific interval from the
// mode's first degree: major and augmented are high, perfect is neutral, and
// minor and diminished are low. Neutral and other fractional qualities are
// unsupported. Relative analysis accepts both open and closed modes.
//
// Absolute analysis requires a closed perfect-octave cycle. For each degree it
// counts one upward interval to every degree, including its own perfect unison
// exactly once. Minor and diminished intervals contribute +1 to the score,
// major and augmented -1, and perfect 0, regardless of integral multiplicity.
// Equal scores form a group; ranks are dense and start at 1, brightest first.
//
// Relations generalize relative positions to any source degree: RelationsFrom
// lists the full upward intervals and categories from one degree, taking lower
// targets in the next cycle of a closed mode; RelationTo gives one relation and
// RelationTable yields the rows from every degree. CompareRelativePositions and
// ComparePositions match the degrees of two modes by number and flag changed
// intervals, categories, scores, and ranks.
//
// Functions do not change or retain their input mode. Results are caller-owned
// snapshots, independent between calls and groups. Slice-containing results
// share elements on assignment, as usual in Go. Concurrent calls on an unchanged
// mode are safe; callers synchronize mutation of modes and returned slices.
// Errors are atomic and preserve dependency causes for errors.Is. Must variants
// call the corresponding ordinary function and panic with the returned error.
package modal
