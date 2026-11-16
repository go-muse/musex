// Package template provides immutable mode templates and catalogs of named
// templates. A Template records only how many places a mode has and the named
// intervals between neighboring places. It is not a mode: it contains no
// degrees, tonic, note names, musical system, or compositional analysis.
//
// # Kinds and places
//
// A closed template has k places and k ascending adjacent intervals, none of
// which is a perfect unison. The last interval closes the pattern by leading
// from place k to place 1 of the next cycle. The cycle is their exact sum and
// need not be an octave. Template.IsOctave reports whether it is a perfect
// octave. An open template has k >= 2 places and k-1 adjacent intervals; it has
// no closing interval, cycle, or continuation beyond its last place.
//
// Places are numbered from 1. Template.PlaceInterval gives the interval from
// the first place to a place, and Template.BetweenPlaces gives a directed
// interval between any two places. A closed template also accepts virtual
// places above k: place k+1 is the first place raised by one cycle, and larger
// places continue by whole cycles. An open template accepts only its actual
// places and can rotate only from place 1.
//
// ClosedFromIntervals and OpenFromIntervals construct from adjacent intervals.
// ClosedFromPlaceIntervals and OpenFromPlaceIntervals construct from intervals
// measured from the first place. ClosedFromSpellings, OpenFromSpellings, and
// FromSpellingsInOctave construct from names read upward in order. The kind is
// explicit in every constructor name; a closed constructor always requires an
// explicit cycle.
//
// MaxPlaces limits the number of actual places to 65536. Virtual place numbers
// of closed templates are limited only by int. Arithmetic uses exact extended
// intermediate coordinates, so an operation fails only when a required final
// interval cannot be represented by package interval.
//
// # Rotation, equality, and comparison
//
// Template.Rotate starts a closed pattern at another place. Template.RotationOf
// reports whether one template is a rotation of another and returns the
// smallest matching place. Template.Equal compares kind and adjacent intervals;
// names and catalog codes are not template data. Template.Compare supplies a
// total order for sorting: unset, then closed, then open, followed by place
// count and adjacent intervals. It is not a musical ordering.
//
// Template.MismatchingPlaces compares templates of the same kind and length.
// FromFirst compares intervals from the first place and includes a closed
// template's cycle as place k+1. Adjacent compares neighboring intervals and
// includes the closing interval. The result contains the mismatching places in
// ascending order.
//
// # Entries and catalogs
//
// An Entry adds an exact code, an ordered nonempty list of names, and a set
// template. A Catalog is an immutable ordered collection with distinct exact
// codes; its zero value is a valid empty catalog. NewCatalog, Catalog.With,
// Catalog.Without, Catalog.Merge, and Catalog.Filter create independent values
// without changing their inputs. Catalog.All returns a repeatable sequence over
// an immutable snapshot.
//
// Catalog.ByCode uses exact code matching. Catalog.ByName compares names after
// Unicode case folding, replacement of musical accidental signs, normalization
// of dashes, removal of Latin combining marks, special handling of Russian ё,
// and whitespace trimming and collapsing. Original names remain unchanged.
// Names containing Cyrillic are reported as Russian; all others are reported as
// English for catalog purposes.
//
// Catalog.Identify finds equal templates. Catalog.IdentifyRotations recognizes
// rotations. Catalog.Related finds entries within a requested number of
// mismatches under either measure, optionally considering every rotation, and
// returns detailed Relation values in stable count and catalog order.
//
// BuiltIn returns the accepted 81-entry catalog. Each entry also has a direct
// pair of functions, such as Ionian and IonianEntry. No template, including a
// major or minor one, is treated as a reference standard for the others.
//
// # Text
//
// Template.String writes adjacent intervals in the short notation of package
// interval, separated by one ASCII space; an open template ends in " |". Parse
// reads that notation and Russian abbreviations with strict spacing.
// Template.MarshalText, Template.AppendText, and Template.UnmarshalText provide
// the machine form used as a JSON string by standard encoders. Template.Russian
// writes Russian abbreviations when every interval has one and otherwise
// reports that no such representation exists.
//
// # Errors, zero values, and concurrency
//
// Errors are classified with errors.Is. Each package error matches exactly one
// sentinel of this package and also preserves a dependency cause where present.
// Operations validate inputs in the order documented by their constructors and
// return no partial result. Every operation with an error result has a Must
// wrapper, except MarshalText, AppendText, and UnmarshalText; a Must wrapper
// panics with the same error value and otherwise returns the same results.
//
// The zero Template and Entry are unset and equal to no value under Equal.
// Their read-only methods return zero values, false, or nil as documented. The
// zero Catalog is empty and equals any other empty catalog. All input slices
// needed by a value are copied, and all returned slices are newly allocated;
// successful empty results are non-nil except documented unset reads.
// Templates, entries, and catalogs are immutable and safe for concurrent reads.
// Callers synchronize concurrent writes to one UnmarshalText receiver or to one
// AppendText buffer.
package template
