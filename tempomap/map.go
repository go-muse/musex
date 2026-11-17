package tempomap

import (
	"math/big"
	"slices"
	"strings"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/tempo"
)

// Map is an immutable tempo map. Measure maps take their grid per operation.
type Map struct {
	addressing Addressing
	changes    []Change
	extensions []Extension
	index      *mapIndex
}

type mapIndex struct {
	changes    []indexedChange
	extensions []indexedExtension
	events     []timeEvent
	positioned bool
}

type indexedChange struct {
	change            Change
	position          duration.Position
	before            State
	after             State
	startSpeed        fraction.Fraction
	startForm         logForm
	approximateBefore bool
}

type indexedExtension struct {
	extension Extension
	position  duration.Position
	before    *big.Rat
}

// New constructs a positional map with the initial main tempo at the origin.
func New(initial tempo.Tempo) (Map, error) {
	if !initial.IsSet() {
		return Map{}, dependencyArgumentError("New", tempo.ErrUnset)
	}
	if _, err := initial.Speed(); err != nil {
		return Map{}, tempoRangeError("New", err)
	}
	at, _ := At(duration.Origin())
	change, _ := NewMain(at, initial)
	change, _ = change.WithSlot(SlotFirst)
	m := Map{addressing: PositionAddressing, changes: []Change{change}}
	idx, err := m.buildIndex(nil)
	if err != nil {
		return Map{}, err
	}
	m.index = idx
	return m, nil
}

// MustNew is like New but panics with its error.
func MustNew(initial tempo.Tempo) Map { return must(New(initial)) }

// NewMeasured constructs a measure map at the first bar of grid.
func NewMeasured(grid measure.Grid, initial tempo.Tempo) (Map, error) {
	if !grid.IsSet() {
		return Map{}, dependencyArgumentError("NewMeasured", measure.ErrUnset)
	}
	if !initial.IsSet() {
		return Map{}, dependencyArgumentError("NewMeasured", tempo.ErrUnset)
	}
	first, err := grid.FirstBar()
	if err != nil {
		return Map{}, measureOperationError("NewMeasured", err)
	}
	bar, err := measure.NewAddress(first, duration.Origin())
	if err != nil {
		return Map{}, measureOperationError("NewMeasured", err)
	}
	at, _ := AtBar(bar)
	if _, err = grid.PositionOf(bar); err != nil {
		return Map{}, measureOperationError("NewMeasured", err)
	}
	if _, err = initial.Speed(); err != nil {
		return Map{}, tempoRangeError("NewMeasured", err)
	}
	change, _ := NewMain(at, initial)
	change, _ = change.WithSlot(SlotFirst)
	m := Map{addressing: MeasureAddressing, changes: []Change{change}}
	if _, err := m.buildIndex([]measure.Grid{grid}); err != nil {
		return Map{}, err
	}
	m.index = nil
	return m, nil
}

// MustNewMeasured is like NewMeasured but panics with its error.
func MustNewMeasured(grid measure.Grid, initial tempo.Tempo) Map {
	return must(NewMeasured(grid, initial))
}

// IsSet reports whether m is set.
func (m Map) IsSet() bool {
	return m.addressing.IsValid() && len(m.changes) > 0 && m.changes[0].IsSet()
}

// Addressing returns the map's address representation.
func (m Map) Addressing() (Addressing, error) {
	if !m.IsSet() {
		return 0, simpleError(ErrUnset, "Map.Addressing", "receiver")
	}
	return m.addressing, nil
}

// MustAddressing is like Addressing but panics with its error.
func (m Map) MustAddressing() Addressing { return must(m.Addressing()) }

// Initial returns the initial exact tempo.
func (m Map) Initial() (tempo.Tempo, error) {
	if !m.IsSet() {
		return tempo.Tempo{}, simpleError(ErrUnset, "Map.Initial", "receiver")
	}
	return m.changes[0].value, nil
}

// MustInitial is like Initial but panics with its error.
func (m Map) MustInitial() tempo.Tempo { return must(m.Initial()) }

// Changes returns a caller-owned copy of all changes in record order.
func (m Map) Changes() ([]Change, error) {
	if !m.IsSet() {
		return nil, simpleError(ErrUnset, "Map.Changes", "receiver")
	}
	return slices.Clone(m.changes), nil
}

// MustChanges is like Changes but panics with its error.
func (m Map) MustChanges() []Change { return must(m.Changes()) }

// Extensions returns a caller-owned copy of all extensions in address order.
func (m Map) Extensions() ([]Extension, error) {
	if !m.IsSet() {
		return nil, simpleError(ErrUnset, "Map.Extensions", "receiver")
	}
	return slices.Clone(m.extensions), nil
}

// MustExtensions is like Extensions but panics with its error.
func (m Map) MustExtensions() []Extension { return must(m.Extensions()) }

// Unfinished reports whether the final change is an unfinished gradual deviation.
func (m Map) Unfinished() (bool, error) {
	if !m.IsSet() {
		return false, simpleError(ErrUnset, "Map.Unfinished", "receiver")
	}
	return m.changes[len(m.changes)-1].kind == ChangeGradual, nil
}

// MustUnfinished is like Unfinished but panics with its error.
func (m Map) MustUnfinished() bool { return must(m.Unfinished()) }

// Equal compares the complete record, including slots and explicit zero holds.
func (m Map) Equal(other Map) bool {
	if m.IsSet() != other.IsSet() {
		return false
	}
	if !m.IsSet() {
		return true
	}
	if m.addressing != other.addressing || len(m.changes) != len(other.changes) || len(m.extensions) != len(other.extensions) {
		return false
	}
	for i := range m.changes {
		if !m.changes[i].Equal(other.changes[i]) {
			return false
		}
	}
	for i := range m.extensions {
		if !m.extensions[i].Equal(other.extensions[i]) {
			return false
		}
	}
	return true
}

// String returns the canonical map text or Map(unset).
func (m Map) String() string {
	if !m.IsSet() {
		return "Map(unset)"
	}
	var b strings.Builder
	b.WriteString("tempomap[")
	for i, c := range m.changes {
		if i > 0 {
			b.WriteByte(';')
		}
		b.WriteString(c.String())
	}
	for _, e := range m.extensions {
		if len(m.changes) > 0 {
			b.WriteByte(';')
		}
		b.WriteString(e.String())
	}
	b.WriteByte(']')
	return b.String()
}

// MarshalText implements encoding.TextMarshaler.
func (m Map) MarshalText() ([]byte, error) {
	if !m.IsSet() {
		return nil, simpleError(ErrUnset, "Map.MarshalText", "receiver")
	}
	return []byte(m.String()), nil
}

// AppendText implements encoding.TextAppender.
func (m Map) AppendText(dst []byte) ([]byte, error) {
	if !m.IsSet() {
		return dst, simpleError(ErrUnset, "Map.AppendText", "receiver")
	}
	return append(dst, m.String()...), nil
}

// UnmarshalText parses a map without a grid and leaves m unchanged on error.
func (m *Map) UnmarshalText(text []byte) error {
	if m == nil {
		return simpleError(ErrInvalidArgument, "Map.UnmarshalText", "nil receiver")
	}
	v, err := Parse(string(text))
	if err != nil {
		return err
	}
	*m = v
	return nil
}

func (m Map) gridArgument(op string, grid []measure.Grid) (measure.Grid, error) {
	if m.addressing == PositionAddressing {
		if len(grid) != 0 {
			return measure.Grid{}, simpleError(ErrInvalidArgument, op, "positional map takes no grid")
		}
		return measure.Grid{}, nil
	}
	if len(grid) != 1 {
		return measure.Grid{}, simpleError(ErrInvalidArgument, op, "measure map requires exactly one grid")
	}
	if !grid[0].IsSet() {
		return measure.Grid{}, dependencyArgumentError(op, measure.ErrUnset)
	}
	return grid[0], nil
}

func addressPosition(op string, at Address, addressing Addressing, grid measure.Grid) (duration.Position, error) {
	if at.addressing != addressing {
		return duration.Position{}, simpleError(ErrInvalidArgument, op, "addressing mismatch")
	}
	if addressing == PositionAddressing {
		return at.position, nil
	}
	p, err := grid.PositionOf(at.bar)
	if err != nil {
		return duration.Position{}, measureOperationError(op+" "+at.String(), err)
	}
	return p, nil
}

func (m Map) buildIndex(grid []measure.Grid) (*mapIndex, error) {
	const op = "tempomap.buildIndex"
	var g measure.Grid
	var err error
	if m.addressing == PositionAddressing {
		if len(grid) != 0 {
			return nil, simpleError(ErrInvalidArgument, op, "positional map takes no grid")
		}
	} else if len(grid) == 1 {
		g = grid[0]
		if !g.IsSet() {
			return nil, dependencyArgumentError(op, measure.ErrUnset)
		}
	} else if len(grid) > 1 {
		return nil, simpleError(ErrInvalidArgument, op, "too many grids")
	}
	if len(m.changes) == 0 || m.changes[0].slot != SlotFirst || m.changes[0].kind != ChangeMain {
		return nil, simpleError(ErrNoInitialTempo, op, "first change must be the initial main tempo")
	}
	idx := &mapIndex{positioned: m.addressing == PositionAddressing || len(grid) == 1}
	idx.changes = make([]indexedChange, len(m.changes))
	positions := make([]duration.Position, len(m.changes))
	for i, c := range m.changes {
		if !c.IsSet() || c.at.addressing != m.addressing || !c.slot.IsValid() || !c.kind.IsValid() {
			return nil, simpleError(ErrInvalidArgument, op, "invalid change record")
		}
		if i == 0 && (c.slot != SlotFirst || c.kind != ChangeMain) {
			return nil, simpleError(ErrNoInitialTempo, op, "invalid initial change")
		}
		if i > 0 {
			cmp := c.at.Compare(m.changes[i-1].at)
			if cmp < 0 {
				return nil, simpleError(ErrDuplicateAddress, op, "changes out of order")
			}
			if cmp == 0 {
				if i != 1 || m.changes[0].slot != SlotFirst || c.slot != SlotSecond || c.kind != ChangeGradual {
					return nil, simpleError(ErrDuplicateAddress, op, "duplicate change address")
				}
			} else if c.slot != SlotSingle {
				return nil, simpleError(ErrInvalidAddress, op, "paired slot outside the initial address")
			}
		}
		if i > 1 && c.at.Compare(m.changes[0].at) == 0 {
			return nil, simpleError(ErrDuplicateAddress, op, "third change at initial address")
		}
		if i > 0 && c.at.Compare(m.changes[0].at) == 0 && c.slot != SlotSecond {
			return nil, simpleError(ErrDuplicateAddress, op, "invalid initial pair")
		}
		if idx.positioned {
			positions[i], err = addressPosition(op, c.at, m.addressing, g)
			if err != nil {
				return nil, err
			}
		}
	}
	if idx.positioned {
		if positions[0].Compare(duration.Origin()) != 0 {
			return nil, simpleError(ErrInvalidAddress, op, "initial change is not at the origin")
		}
	} else if m.addressing == MeasureAddressing {
		bar, _ := m.changes[0].at.bar.Bar()
		offset, _ := m.changes[0].at.bar.Offset()
		if (bar != 0 && bar != 1) || offset.Compare(duration.Origin()) != 0 {
			return nil, simpleError(ErrInvalidAddress, op, "initial measure address must begin bar zero or one")
		}
	}

	state := State{current: m.changes[0].value, main: m.changes[0].value, set: true}
	initial := state.current
	for i, c := range m.changes {
		if i > 0 {
			previous := m.changes[i-1]
			if previous.kind == ChangeGradual && (c.kind == ChangeGradual || c.kind == ChangeModulation) {
				return nil, simpleError(ErrInvalidSequence, op, "gradual change has no exact following tempo")
			}
			before := state
			switch c.kind {
			case ChangeMain:
				state = State{current: c.value, main: c.value, set: true}
			case ChangeJump:
				state.current, state.deviating = c.value, true
			case ChangeGradual:
				state.deviating = true
			case ChangeATempo:
				if !before.deviating {
					return nil, simpleError(ErrNoDeviation, op, "a tempo without a deviation")
				}
				state.current, state.deviating = state.main, false
			case ChangeTempoI:
				state = State{current: initial, main: initial, set: true}
			case ChangeModulation:
				modulated, modErr := before.current.Modulate(c.newUnit, c.old)
				if modErr != nil {
					return nil, tempoRangeError(op, modErr)
				}
				state = State{current: modulated, main: modulated, set: true}
			}
		}
		speed, speedErr := state.current.Speed()
		if speedErr != nil {
			return nil, tempoRangeError(op, speedErr)
		}
		idx.changes[i] = indexedChange{change: c, position: positions[i], before: func() State {
			if i == 0 {
				return State{}
			}
			return idx.changes[i-1].after
		}(), after: state, startSpeed: speed}
	}

	idx.extensions = make([]indexedExtension, len(m.extensions))
	extPositions := make([]duration.Position, len(m.extensions))
	for i, e := range m.extensions {
		if !e.IsSet() || e.at.addressing != m.addressing {
			return nil, simpleError(ErrInvalidArgument, op, "invalid extension")
		}
		if i > 0 {
			cmp := e.at.Compare(m.extensions[i-1].at)
			if cmp < 0 {
				return nil, simpleError(ErrDuplicateAddress, op, "extensions out of order")
			}
			if cmp == 0 {
				return nil, simpleError(ErrDuplicateAddress, op, "duplicate extension address")
			}
		}
		if idx.positioned {
			extPositions[i], err = addressPosition(op, e.at, m.addressing, g)
			if err != nil {
				return nil, err
			}
		}
		idx.extensions[i] = indexedExtension{extension: e, position: extPositions[i]}
	}
	if idx.positioned {
		prepareTimeline(idx)
	}
	return idx, nil
}

func (m Map) positionIndex(grid []measure.Grid) (*mapIndex, error) {
	if m.addressing == PositionAddressing && m.index != nil {
		if len(grid) != 0 {
			return nil, simpleError(ErrInvalidArgument, "tempomap.positionIndex", "positional map takes no grid")
		}
		return m.index, nil
	}
	return m.buildIndex(grid)
}

// Validate checks all record, address, state, and sequence invariants.
func (m Map) Validate(grid ...measure.Grid) (unfinished bool, err error) {
	if !m.IsSet() {
		return false, simpleError(ErrUnset, "Map.Validate", "receiver")
	}
	if _, err = m.gridArgument("Map.Validate", grid); err != nil {
		return false, err
	}
	if _, err = m.buildIndex(grid); err != nil {
		return false, err
	}
	return m.changes[len(m.changes)-1].kind == ChangeGradual, nil
}

// MustValidate is like Validate but panics with its error.
func (m Map) MustValidate(grid ...measure.Grid) bool { return must(m.Validate(grid...)) }
