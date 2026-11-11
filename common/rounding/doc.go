// Package rounding provides exact rounding on uniform grids and neighbor
// selection on arbitrary grids.
//
// A grid is an ordered set of nodes. For a value strictly between two adjacent
// nodes, the lower and upper neighbors are the only possible choices. A Mode
// defines that choice; there is no default mode, so every consumer selects its
// own policy. A value already on a node remains unchanged. On a non-uniform
// grid the caller compares exact distances, defines the grid's zero, and
// supplies the parity of the lower neighbor's index. Choose is not called for
// a value on a node.
//
// The nine modes have the following results on an integer grid. Negative and
// positive zero are not distinguished in results.
//
//	Mode                -2.5  -1.5  -0.7  -0.5   0.5   0.7   1.5   2.5
//	Down                  -3    -2    -1    -1     0     0     1     2
//	Up                    -2    -1     0     0     1     1     2     3
//	TowardZero            -2    -1     0     0     0     0     1     2
//	AwayFromZero          -3    -2    -1    -1     1     1     2     3
//	HalfUp                -2    -1    -1     0     1     1     2     3
//	HalfDown              -3    -2    -1    -1     0     1     1     2
//	HalfAwayFromZero      -3    -2    -1    -1     1     1     2     3
//	HalfTowardZero        -2    -1    -1     0     0     1     1     2
//	HalfEven              -2    -2    -1     0     0     1     2     2
//
// Names differ between rounding libraries. Here HalfUp resolves a tie toward
// positive infinity and Up always rounds toward positive infinity. In
// java.math.RoundingMode and Python's decimal module, HALF_UP instead means
// ties away from zero (HalfAwayFromZero here), and UP means away from zero
// (AwayFromZero here).
//
// Uniform-grid operations cover integers, multiples of a step, exact thirds
// and other 1/den grids, and decimal places from -MaxPlaces through MaxPlaces.
// Negative decimal places select tens, hundreds, and larger nodes. Decimal
// strings never use an exponent and never give zero a minus sign.
//
// A finite float64 below 2^53 in magnitude is interpreted as its shortest
// decimal representation, not its exact binary value. Starting at 2^53, its
// exact integer value is used. Thus HalfUp.FloatDecimal(0.145, 2) is "0.15",
// while strconv.FormatFloat(0.145, 'f', 2, 64) is "0.14". For another example,
// FloatDecimal(1e23, 0) is "99999999999999991611392". To round the exact binary
// value of any float64, convert it with new(big.Rat).SetFloat64 and use the Rat
// methods. Results converted back to float64 use one IEEE 754 round-to-nearest,
// ties-to-even operation. Intermediate choices use exact arithmetic.
//
// These modes correspond to existing MuseX policies without importing their
// packages: common/fraction's five rules are HalfAwayFromZero, HalfEven, Down,
// Up, and TowardZero; cents uses HalfAwayFromZero; tuning/temperament resolves
// ties like HalfDown.
//
// Operations return ErrUnset, ErrUnknownMode, ErrInvalidArgument, or
// ErrOutOfRange, in that precedence order. Errors.Is recognizes exactly one of
// these sentinels in each returned error. Must variants call the corresponding
// operation and panic with its error.
//
// Canonical names support configuration through Parse, Name, and the standard
// text interfaces. Mode also works with JSON through those interfaces. Since
// an unset Mode cannot be encoded, optional fields can use omitzero. JSON null
// is handled by the JSON decoder rather than UnmarshalText: encoding/json/v2
// produces the zero Mode, while encoding/json leaves an existing field
// unchanged.
//
// The package has no mutable package state or caches. Operations do not change
// or retain their inputs. Independent calls are safe concurrently, including
// calls sharing read-only big.Int and big.Rat inputs. A caller must synchronize
// mutation of an input concurrent with a call, and concurrent UnmarshalText
// calls targeting the same Mode variable. Mode, Side, and Neighbor are small
// comparable value types. Allocation costs are constant for mode, neighbor,
// integer, and whole-float operations; arbitrary-precision operations allocate
// their new results and bounded intermediate values.
package rounding
