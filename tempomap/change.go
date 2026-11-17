package tempomap

import (
	"strings"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tempo"
)

// Change is a tempo change record. Constructors set SlotSingle; WithSlot
// selects SlotFirst or SlotSecond for the initial pair.
type Change struct {
	at      Address
	slot    Slot
	kind    ChangeKind
	value   tempo.Tempo
	law     Law
	newUnit duration.Written
	old     duration.Duration
}

func newChange(at Address, kind ChangeKind) (Change, error) {
	if !at.IsSet() {
		return Change{}, simpleError(ErrUnset, "NewChange", "address")
	}
	return Change{at: at, slot: SlotSingle, kind: kind}, nil
}

// NewMain constructs a new main-tempo change.
func NewMain(at Address, value tempo.Tempo) (Change, error) {
	c, err := newChange(at, ChangeMain)
	if err != nil {
		return Change{}, err
	}
	if !value.IsSet() {
		return Change{}, dependencyArgumentError("NewMain", tempo.ErrUnset)
	}
	c.value = value
	return c, nil
}

// MustNewMain is like NewMain but panics with its error.
func MustNewMain(at Address, value tempo.Tempo) Change { return must(NewMain(at, value)) }

// NewJump constructs a jump deviation.
func NewJump(at Address, value tempo.Tempo) (Change, error) {
	c, err := newChange(at, ChangeJump)
	if err != nil {
		return Change{}, err
	}
	if !value.IsSet() {
		return Change{}, dependencyArgumentError("NewJump", tempo.ErrUnset)
	}
	c.value = value
	return c, nil
}

// MustNewJump is like NewJump but panics with its error.
func MustNewJump(at Address, value tempo.Tempo) Change { return must(NewJump(at, value)) }

// NewGradual constructs a gradual deviation with an explicit law.
func NewGradual(at Address, law Law) (Change, error) {
	c, err := newChange(at, ChangeGradual)
	if err != nil {
		return Change{}, err
	}
	if !law.IsValid() {
		return Change{}, simpleError(ErrInvalidArgument, "NewGradual", "invalid law")
	}
	c.law = law
	return c, nil
}

// MustNewGradual is like NewGradual but panics with its error.
func MustNewGradual(at Address, law Law) Change { return must(NewGradual(at, law)) }

// NewATempo constructs an immediate return to the current main tempo.
func NewATempo(at Address) (Change, error) { return newChange(at, ChangeATempo) }

// MustNewATempo is like NewATempo but panics with its error.
func MustNewATempo(at Address) Change { return must(NewATempo(at)) }

// NewTempoI constructs an immediate return to the initial tempo.
func NewTempoI(at Address) (Change, error) { return newChange(at, ChangeTempoI) }

// MustNewTempoI is like NewTempoI but panics with its error.
func MustNewTempoI(at Address) Change { return must(NewTempoI(at)) }

// NewModulation constructs a metric modulation.
func NewModulation(at Address, newUnit duration.Written, old duration.Duration) (Change, error) {
	c, err := newChange(at, ChangeModulation)
	if err != nil {
		return Change{}, err
	}
	if !newUnit.IsSet() {
		return Change{}, dependencyArgumentError("NewModulation", duration.ErrUnset)
	}
	if !old.IsSet() {
		return Change{}, dependencyArgumentError("NewModulation", duration.ErrUnset)
	}
	c.newUnit, c.old = newUnit, old
	return c, nil
}

// MustNewModulation is like NewModulation but panics with its error.
func MustNewModulation(at Address, newUnit duration.Written, old duration.Duration) Change {
	return must(NewModulation(at, newUnit, old))
}

// IsSet reports whether c is set.
func (c Change) IsSet() bool { return c.at.IsSet() && c.slot.IsValid() && c.kind.IsValid() }

// Address returns the change address.
func (c Change) Address() (Address, error) {
	if !c.IsSet() {
		return Address{}, simpleError(ErrUnset, "Change.Address", "receiver")
	}
	return c.at, nil
}

// MustAddress is like Address but panics with its error.
func (c Change) MustAddress() Address { return must(c.Address()) }

// Slot returns the record slot.
func (c Change) Slot() (Slot, error) {
	if !c.IsSet() {
		return 0, simpleError(ErrUnset, "Change.Slot", "receiver")
	}
	return c.slot, nil
}

// MustSlot is like Slot but panics with its error.
func (c Change) MustSlot() Slot { return must(c.Slot()) }

// Kind returns the change kind.
func (c Change) Kind() (ChangeKind, error) {
	if !c.IsSet() {
		return 0, simpleError(ErrUnset, "Change.Kind", "receiver")
	}
	return c.kind, nil
}

// MustKind is like Kind but panics with its error.
func (c Change) MustKind() ChangeKind { return must(c.Kind()) }

// Tempo returns the explicit tempo for main and jump changes.
func (c Change) Tempo() (tempo.Tempo, bool, error) {
	if !c.IsSet() {
		return tempo.Tempo{}, false, simpleError(ErrUnset, "Change.Tempo", "receiver")
	}
	if c.kind != ChangeMain && c.kind != ChangeJump {
		return tempo.Tempo{}, false, nil
	}
	return c.value, true, nil
}

// MustTempo is like Tempo but panics with its error.
func (c Change) MustTempo() (tempo.Tempo, bool) { return must2(c.Tempo()) }

// Law returns the law for a gradual change.
func (c Change) Law() (Law, bool, error) {
	if !c.IsSet() {
		return 0, false, simpleError(ErrUnset, "Change.Law", "receiver")
	}
	if c.kind != ChangeGradual {
		return 0, false, nil
	}
	return c.law, true, nil
}

// MustLaw is like Law but panics with its error.
func (c Change) MustLaw() (Law, bool) { return must2(c.Law()) }

// Modulation returns the new unit and old duration for a modulation.
func (c Change) Modulation() (newUnit duration.Written, old duration.Duration, ok bool, err error) {
	if !c.IsSet() {
		return duration.Written{}, duration.Duration{}, false, simpleError(ErrUnset, "Change.Modulation", "receiver")
	}
	if c.kind != ChangeModulation {
		return duration.Written{}, duration.Duration{}, false, nil
	}
	return c.newUnit, c.old, true, nil
}

// MustModulation is like Modulation but panics with its error.
func (c Change) MustModulation() (newUnit duration.Written, old duration.Duration, ok bool) {
	return must3(c.Modulation())
}

// WithSlot returns c with only its slot changed.
func (c Change) WithSlot(slot Slot) (Change, error) {
	if !c.IsSet() {
		return Change{}, simpleError(ErrUnset, "Change.WithSlot", "receiver")
	}
	if !slot.IsValid() {
		return Change{}, simpleError(ErrInvalidArgument, "Change.WithSlot", "invalid slot")
	}
	c.slot = slot
	return c, nil
}

// MustWithSlot is like WithSlot but panics with its error.
func (c Change) MustWithSlot(slot Slot) Change { return must(c.WithSlot(slot)) }

// Equal reports whether c and other are the same record.
func (c Change) Equal(other Change) bool { return c == other }

// String returns the canonical change text or Change(unset).
func (c Change) String() string {
	if !c.IsSet() {
		return "Change(unset)"
	}
	var b strings.Builder
	b.WriteString(c.slot.String())
	b.WriteByte('@')
	b.WriteString(c.at.String())
	b.WriteByte(':')
	b.WriteString(c.kind.String())
	switch c.kind {
	case ChangeMain, ChangeJump:
		b.WriteByte(':')
		b.WriteString(c.value.String())
	case ChangeGradual:
		b.WriteByte(':')
		b.WriteString(c.law.String())
	case ChangeModulation:
		b.WriteByte(':')
		b.WriteString(c.newUnit.String())
		b.WriteByte('@')
		b.WriteString(c.old.String())
	}
	return b.String()
}

// MarshalText implements encoding.TextMarshaler.
func (c Change) MarshalText() ([]byte, error) {
	if !c.IsSet() {
		return nil, simpleError(ErrUnset, "Change.MarshalText", "receiver")
	}
	return []byte(c.String()), nil
}

// AppendText implements encoding.TextAppender.
func (c Change) AppendText(dst []byte) ([]byte, error) {
	if !c.IsSet() {
		return dst, simpleError(ErrUnset, "Change.AppendText", "receiver")
	}
	return append(dst, c.String()...), nil
}

// UnmarshalText implements encoding.TextUnmarshaler and leaves c unchanged on error.
func (c *Change) UnmarshalText(text []byte) error {
	if c == nil {
		return simpleError(ErrInvalidArgument, "Change.UnmarshalText", "nil receiver")
	}
	v, err := ParseChange(string(text))
	if err != nil {
		return err
	}
	*c = v
	return nil
}

// Extension is an explicit fermata hold in seconds at one address.
type Extension struct {
	at      Address
	seconds tempo.Seconds
}

// NewExtension constructs an extension, including an explicit zero hold.
func NewExtension(at Address, seconds tempo.Seconds) (Extension, error) {
	if !at.IsSet() {
		return Extension{}, simpleError(ErrUnset, "NewExtension", "address")
	}
	if !seconds.IsSet() {
		return Extension{}, dependencyArgumentError("NewExtension", tempo.ErrUnset)
	}
	if ratOfSeconds(seconds).Sign() < 0 {
		return Extension{}, causedError(ErrInvalidArgument, "NewExtension", "negative hold", tempo.ErrNegativeTime)
	}
	return Extension{at: at, seconds: seconds}, nil
}

// MustNewExtension is like NewExtension but panics with its error.
func MustNewExtension(at Address, seconds tempo.Seconds) Extension {
	return must(NewExtension(at, seconds))
}

// IsSet reports whether e is set.
func (e Extension) IsSet() bool { return e.at.IsSet() && e.seconds.IsSet() }

// Address returns the extension address.
func (e Extension) Address() (Address, error) {
	if !e.IsSet() {
		return Address{}, simpleError(ErrUnset, "Extension.Address", "receiver")
	}
	return e.at, nil
}

// MustAddress is like Address but panics with its error.
func (e Extension) MustAddress() Address { return must(e.Address()) }

// Seconds returns the exact hold duration.
func (e Extension) Seconds() (tempo.Seconds, error) {
	if !e.IsSet() {
		return tempo.Seconds{}, simpleError(ErrUnset, "Extension.Seconds", "receiver")
	}
	return e.seconds, nil
}

// MustSeconds is like Seconds but panics with its error.
func (e Extension) MustSeconds() tempo.Seconds { return must(e.Seconds()) }

// Equal reports whether e and other are the same record.
func (e Extension) Equal(other Extension) bool { return e == other }

// String returns the canonical extension text or Extension(unset).
func (e Extension) String() string {
	if !e.IsSet() {
		return "Extension(unset)"
	}
	return "hold@" + e.at.String() + ":" + e.seconds.String()
}

// MarshalText implements encoding.TextMarshaler.
func (e Extension) MarshalText() ([]byte, error) {
	if !e.IsSet() {
		return nil, simpleError(ErrUnset, "Extension.MarshalText", "receiver")
	}
	return []byte(e.String()), nil
}

// AppendText implements encoding.TextAppender.
func (e Extension) AppendText(dst []byte) ([]byte, error) {
	if !e.IsSet() {
		return dst, simpleError(ErrUnset, "Extension.AppendText", "receiver")
	}
	return append(dst, e.String()...), nil
}

// UnmarshalText implements encoding.TextUnmarshaler and leaves e unchanged on error.
func (e *Extension) UnmarshalText(text []byte) error {
	if e == nil {
		return simpleError(ErrInvalidArgument, "Extension.UnmarshalText", "nil receiver")
	}
	v, err := ParseExtension(string(text))
	if err != nil {
		return err
	}
	*e = v
	return nil
}
