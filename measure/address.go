package measure

import (
	"strconv"
	"strings"

	"github.com/go-muse/musex/duration"
)

// Address is a moment as a bar number and an offset from the start of the
// bar: (2, 1/8) is an eighth after the start of bar 2. The pickup is bar 0.
//
// An address does not know a grid: (2, 5/4) is a valid value even though
// in 4/4 its offset is not less than the length of bar 2. Whether an address
// is valid in a grid is reported by Grid.Contains, and Grid.Normalize turns
// any offset into a valid address of the same moment.
//
// The zero value Address{} is the unset address. Only IsSet, Equal, Compare,
// and String accept it without an error; String returns "Address(unset)".
// Address is comparable with ==, which matches Equal, and can be used as a
// map key.
type Address struct {
	bar    int
	offset duration.Position
}

// NewAddress returns the address of bar bar at offset offset from the start
// of the bar.
//
// Errors, in this order: ErrInvalidArgument if bar is negative;
// ErrInvalidArgument together with duration.ErrUnset if offset is unset.
//
// Example: NewAddress(2, duration.MustNewPosition(1, 8)) is "2:1/8".
func NewAddress(bar int, offset duration.Position) (Address, error) {
	const op = "NewAddress"
	if bar < 0 {
		return Address{}, errInvalid(op, "negative bar number")
	}
	if !offset.IsSet() {
		return Address{}, errUnsetTime(op, "offset")
	}
	return Address{bar: bar, offset: offset}, nil
}

// MustNewAddress is like NewAddress but panics with the error of NewAddress.
func MustNewAddress(bar int, offset duration.Position) Address {
	return must(NewAddress(bar, offset))
}

// ParseAddress parses the canonical text of an address, "<bar>:<position>",
// such as "2:1/8" or "0:0/1". The bar is "0" or a decimal number without a
// leading zero; the position is parsed by duration.ParsePosition, so its
// non-canonical forms, such as "2/16", are accepted. Spaces are not
// accepted.
//
// Errors, in this order: ErrInvalidText for text outside the grammar,
// including "Address(unset)"; ErrOutOfRange for a bar number greater than
// math.MaxInt; ErrOutOfRange together with duration.ErrOutOfRange and
// fraction.ErrOutOfRange for a position outside the range; ErrInvalidText
// for any other rejected position. The message quotes the text as described
// in the package documentation.
func ParseAddress(text string) (Address, error) {
	barText, posText, ok := strings.Cut(text, ":")
	if !ok || posText == "" || strings.ContainsAny(posText, ";:") {
		return Address{}, parseError(classInvalidText, text)
	}
	if barText != "0" && !isCount(barText) {
		return Address{}, parseError(classInvalidText, text)
	}
	bar, ok := parseCount(barText)
	if !ok {
		return Address{}, parseError(classNumberRange, text)
	}
	offset, err := duration.ParsePosition(posText)
	if err != nil {
		return Address{}, parseError(nestedClass(err), text)
	}
	return Address{bar: bar, offset: offset}, nil
}

// MustParseAddress is like ParseAddress but panics with the error of
// ParseAddress.
func MustParseAddress(text string) Address {
	return must(ParseAddress(text))
}

// IsSet reports whether a is not the unset address Address{}.
func (a Address) IsSet() bool {
	return a.offset.IsSet()
}

// Bar returns the bar number of a; 0 is the pickup.
//
// Errors: ErrUnset for the unset address.
func (a Address) Bar() (int, error) {
	if !a.IsSet() {
		return 0, errUnset("Address.Bar", "address")
	}
	return a.bar, nil
}

// MustBar is like Bar but panics with the error of Bar.
func (a Address) MustBar() int {
	return must(a.Bar())
}

// Offset returns the offset of a from the start of its bar.
//
// Errors: ErrUnset for the unset address.
func (a Address) Offset() (duration.Position, error) {
	if !a.IsSet() {
		return duration.Position{}, errUnset("Address.Offset", "address")
	}
	return a.offset, nil
}

// MustOffset is like Offset but panics with the error of Offset.
func (a Address) MustOffset() duration.Position {
	return must(a.Offset())
}

// Equal reports whether a and other have the same bar and offset; two unset
// addresses are equal, and the unset address equals only itself. It is the
// same as ==.
func (a Address) Equal(other Address) bool {
	return a == other
}

// Compare returns -1, 0, or +1 as a is before, equal to, or after other: by
// bar number and, at an equal bar, by offset. The unset address is before
// every set address, and two unset addresses compare as 0. It suits
// slices.SortFunc(addresses, Address.Compare).
func (a Address) Compare(other Address) int {
	switch {
	case !a.IsSet() && !other.IsSet():
		return 0
	case !a.IsSet():
		return -1
	case !other.IsSet():
		return 1
	case a.bar < other.bar:
		return -1
	case a.bar > other.bar:
		return 1
	default:
		return a.offset.Compare(other.offset)
	}
}

// String returns the canonical text "<bar>:<position>", such as "2:1/8", or
// "Address(unset)" for the unset address.
func (a Address) String() string {
	if !a.IsSet() {
		return "Address(unset)"
	}
	return strconv.Itoa(a.bar) + ":" + a.offset.String()
}

// MarshalText returns the canonical text of a, as String does.
//
// Errors: ErrUnset for the unset address.
func (a Address) MarshalText() ([]byte, error) {
	if !a.IsSet() {
		return nil, errUnset("Address.MarshalText", "address")
	}
	return []byte(a.String()), nil
}

// AppendText appends the canonical text of a to dst and returns the extended
// slice.
//
// Errors: ErrUnset for the unset address; dst is then returned unchanged.
func (a Address) AppendText(dst []byte) ([]byte, error) {
	if !a.IsSet() {
		return dst, errUnset("Address.AppendText", "address")
	}
	return append(dst, a.String()...), nil
}

// UnmarshalText parses text as ParseAddress does and stores the result in
// *a. On an error *a is not changed.
//
// Errors: ErrInvalidArgument for a nil receiver; otherwise those of
// ParseAddress.
func (a *Address) UnmarshalText(text []byte) error {
	if a == nil {
		return errInvalid("Address.UnmarshalText", "nil receiver")
	}
	v, err := ParseAddress(string(text))
	if err != nil {
		return err
	}
	*a = v
	return nil
}
