package tempomap

import (
	"strings"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/measure"
)

// MaxPlaces is the largest accepted number of decimal places.
const MaxPlaces = 1000

// Addressing selects the address representation of a whole map.
type Addressing uint8

const (
	PositionAddressing Addressing = iota + 1
	MeasureAddressing
)

// IsValid reports whether a is a defined addressing mode.
func (a Addressing) IsValid() bool { return a == PositionAddressing || a == MeasureAddressing }

// Equal reports whether a and other are identical.
func (a Addressing) Equal(other Addressing) bool { return a == other }

// String returns the canonical addressing name or a diagnostic invalid value.
func (a Addressing) String() string {
	switch a {
	case PositionAddressing:
		return "positions"
	case MeasureAddressing:
		return "measures"
	default:
		return "Addressing(invalid)"
	}
}

// Slot identifies a change at its address, including the initial pair.
type Slot uint8

const (
	SlotSingle Slot = iota + 1
	SlotFirst
	SlotSecond
)

// IsValid reports whether s is a defined slot.
func (s Slot) IsValid() bool { return s >= SlotSingle && s <= SlotSecond }

// Equal reports whether s and other are identical.
func (s Slot) Equal(other Slot) bool { return s == other }

// String returns the canonical slot name or a diagnostic invalid value.
func (s Slot) String() string {
	switch s {
	case SlotSingle:
		return "single"
	case SlotFirst:
		return "first"
	case SlotSecond:
		return "second"
	default:
		return "Slot(invalid)"
	}
}

// ChangeKind is the kind of a tempo change.
type ChangeKind uint8

const (
	ChangeMain ChangeKind = iota + 1
	ChangeJump
	ChangeGradual
	ChangeATempo
	ChangeTempoI
	ChangeModulation
)

// IsValid reports whether k is a defined change kind.
func (k ChangeKind) IsValid() bool { return k >= ChangeMain && k <= ChangeModulation }

// Equal reports whether k and other are identical.
func (k ChangeKind) Equal(other ChangeKind) bool { return k == other }

// String returns the canonical change-kind name or a diagnostic invalid value.
func (k ChangeKind) String() string {
	switch k {
	case ChangeMain:
		return "main"
	case ChangeJump:
		return "jump"
	case ChangeGradual:
		return "gradual"
	case ChangeATempo:
		return "a-tempo"
	case ChangeTempoI:
		return "tempo-i"
	case ChangeModulation:
		return "modulation"
	default:
		return "ChangeKind(invalid)"
	}
}

// Law is the explicitly selected law of a gradual change.
type Law uint8

const (
	// LinearSpeed selects linear whole-notes-per-minute speed by position.
	LinearSpeed Law = iota + 1
)

// IsValid reports whether l is a defined gradual law.
func (l Law) IsValid() bool { return l == LinearSpeed }

// Equal reports whether l and other are identical.
func (l Law) Equal(other Law) bool { return l == other }

// String returns the canonical law name or a diagnostic invalid value.
func (l Law) String() string {
	if l == LinearSpeed {
		return "linear-speed"
	}
	return "Law(invalid)"
}

// ParseLaw parses the canonical law name.
func ParseLaw(text string) (Law, error) {
	if text == LinearSpeed.String() {
		return LinearSpeed, nil
	}
	return 0, invalidTextError(text)
}

// MustParseLaw is like ParseLaw but panics with its error.
func MustParseLaw(text string) Law { return must(ParseLaw(text)) }

// MarshalText implements encoding.TextMarshaler.
func (l Law) MarshalText() ([]byte, error) {
	if !l.IsValid() {
		return nil, simpleError(ErrInvalidArgument, "Law.MarshalText", "invalid law")
	}
	return []byte(l.String()), nil
}

// AppendText implements encoding.TextAppender.
func (l Law) AppendText(dst []byte) ([]byte, error) {
	if !l.IsValid() {
		return dst, simpleError(ErrInvalidArgument, "Law.AppendText", "invalid law")
	}
	return append(dst, l.String()...), nil
}

// UnmarshalText implements encoding.TextUnmarshaler and leaves l unchanged on error.
func (l *Law) UnmarshalText(text []byte) error {
	if l == nil {
		return simpleError(ErrInvalidArgument, "Law.UnmarshalText", "nil receiver")
	}
	v, err := ParseLaw(string(text))
	if err != nil {
		return err
	}
	*l = v
	return nil
}

// Address is either a position or a measure address.
type Address struct {
	addressing Addressing
	position   duration.Position
	bar        measure.Address
}

// At constructs a positional address.
func At(position duration.Position) (Address, error) {
	if !position.IsSet() {
		return Address{}, dependencyArgumentError("At", duration.ErrUnset)
	}
	return Address{addressing: PositionAddressing, position: position}, nil
}

// MustAt is like At but panics with its error.
func MustAt(position duration.Position) Address { return must(At(position)) }

// AtBar constructs a measure address.
func AtBar(address measure.Address) (Address, error) {
	if !address.IsSet() {
		return Address{}, dependencyArgumentError("AtBar", measure.ErrUnset)
	}
	return Address{addressing: MeasureAddressing, bar: address}, nil
}

// MustAtBar is like AtBar but panics with its error.
func MustAtBar(address measure.Address) Address { return must(AtBar(address)) }

// ParseAddress parses p:position or m:bar:offset.
func ParseAddress(text string) (Address, error) {
	address, err := parseAddress(text)
	if err != nil {
		return Address{}, normalizeTextError(text, err)
	}
	return address, nil
}

func parseAddress(text string) (Address, error) {
	if strings.HasPrefix(text, "p:") {
		p, err := duration.ParsePosition(text[2:])
		if err != nil {
			return Address{}, nestedTextError(text, err)
		}
		return At(p)
	}
	if strings.HasPrefix(text, "m:") {
		a, err := measure.ParseAddress(text[2:])
		if err != nil {
			return Address{}, nestedTextError(text, err)
		}
		return AtBar(a)
	}
	return Address{}, invalidTextError(text)
}

// MustParseAddress is like ParseAddress but panics with its error.
func MustParseAddress(text string) Address { return must(ParseAddress(text)) }

// IsSet reports whether a is set.
func (a Address) IsSet() bool {
	switch a.addressing {
	case PositionAddressing:
		return a.position.IsSet()
	case MeasureAddressing:
		return a.bar.IsSet()
	default:
		return false
	}
}

// Addressing returns the address branch.
func (a Address) Addressing() (Addressing, error) {
	if !a.IsSet() {
		return 0, simpleError(ErrUnset, "Address.Addressing", "receiver")
	}
	return a.addressing, nil
}

// MustAddressing is like Addressing but panics with its error.
func (a Address) MustAddressing() Addressing { return must(a.Addressing()) }

// Position returns the positional value and true, or an unset position and false
// for a measure address.
func (a Address) Position() (duration.Position, bool, error) {
	if !a.IsSet() {
		return duration.Position{}, false, simpleError(ErrUnset, "Address.Position", "receiver")
	}
	if a.addressing != PositionAddressing {
		return duration.Position{}, false, nil
	}
	return a.position, true, nil
}

// MustPosition is like Position but panics with its error.
func (a Address) MustPosition() (duration.Position, bool) { return must2(a.Position()) }

// Bar returns the measure value and true, or an unset measure address and false
// for a positional address.
func (a Address) Bar() (measure.Address, bool, error) {
	if !a.IsSet() {
		return measure.Address{}, false, simpleError(ErrUnset, "Address.Bar", "receiver")
	}
	if a.addressing != MeasureAddressing {
		return measure.Address{}, false, nil
	}
	return a.bar, true, nil
}

// MustBar is like Bar but panics with its error.
func (a Address) MustBar() (measure.Address, bool) { return must2(a.Bar()) }

// Equal reports whether a and other contain the same address.
func (a Address) Equal(other Address) bool { return a == other }

// Compare defines the total order: unset, positions by value, then measure addresses.
func (a Address) Compare(other Address) int {
	if !a.IsSet() || !other.IsSet() {
		switch {
		case a.IsSet() == other.IsSet():
			return 0
		case !a.IsSet():
			return -1
		default:
			return 1
		}
	}
	if a.addressing != other.addressing {
		if a.addressing == PositionAddressing {
			return -1
		}
		return 1
	}
	if a.addressing == PositionAddressing {
		return a.position.Compare(other.position)
	}
	return a.bar.Compare(other.bar)
}

// String returns the canonical address text or Address(unset).
func (a Address) String() string {
	if !a.IsSet() {
		return "Address(unset)"
	}
	if a.addressing == PositionAddressing {
		return "p:" + a.position.String()
	}
	return "m:" + a.bar.String()
}

// MarshalText implements encoding.TextMarshaler.
func (a Address) MarshalText() ([]byte, error) {
	if !a.IsSet() {
		return nil, simpleError(ErrUnset, "Address.MarshalText", "receiver")
	}
	return []byte(a.String()), nil
}

// AppendText implements encoding.TextAppender.
func (a Address) AppendText(dst []byte) ([]byte, error) {
	if !a.IsSet() {
		return dst, simpleError(ErrUnset, "Address.AppendText", "receiver")
	}
	return append(dst, a.String()...), nil
}

// UnmarshalText implements encoding.TextUnmarshaler and leaves a unchanged on error.
func (a *Address) UnmarshalText(text []byte) error {
	if a == nil {
		return simpleError(ErrInvalidArgument, "Address.UnmarshalText", "nil receiver")
	}
	v, err := ParseAddress(string(text))
	if err != nil {
		return err
	}
	*a = v
	return nil
}
