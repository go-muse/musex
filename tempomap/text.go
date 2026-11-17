package tempomap

import (
	"strings"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/tempo"
)

func parseSlot(text string) (Slot, bool) {
	switch text {
	case "single":
		return SlotSingle, true
	case "first":
		return SlotFirst, true
	case "second":
		return SlotSecond, true
	default:
		return 0, false
	}
}

func parseChangeKind(text string) (ChangeKind, bool) {
	switch text {
	case "main":
		return ChangeMain, true
	case "jump":
		return ChangeJump, true
	case "gradual":
		return ChangeGradual, true
	case "a-tempo":
		return ChangeATempo, true
	case "tempo-i":
		return ChangeTempoI, true
	case "modulation":
		return ChangeModulation, true
	default:
		return 0, false
	}
}

func splitAddressText(text string) (address, rest string, ok bool) {
	if strings.HasPrefix(text, "p:") {
		i := strings.IndexByte(text[2:], ':')
		if i < 0 {
			return "", "", false
		}
		i += 2
		return text[:i], text[i+1:], true
	}
	if strings.HasPrefix(text, "m:") {
		first := strings.IndexByte(text[2:], ':')
		if first < 0 {
			return "", "", false
		}
		first += 2
		second := strings.IndexByte(text[first+1:], ':')
		if second < 0 {
			return "", "", false
		}
		second += first + 1
		return text[:second], text[second+1:], true
	}
	return "", "", false
}

// ParseChange parses one complete change record.
func ParseChange(text string) (Change, error) {
	slotText, rest, ok := strings.Cut(text, "@")
	if !ok {
		return Change{}, invalidTextError(text)
	}
	slot, ok := parseSlot(slotText)
	if !ok {
		return Change{}, invalidTextError(text)
	}
	addressText, rest, ok := splitAddressText(rest)
	if !ok {
		return Change{}, invalidTextError(text)
	}
	at, err := ParseAddress(addressText)
	if err != nil {
		return Change{}, err
	}
	kindText, data, hasData := strings.Cut(rest, ":")
	kind, ok := parseChangeKind(kindText)
	if !ok {
		return Change{}, invalidTextError(text)
	}
	var c Change
	switch kind {
	case ChangeMain, ChangeJump:
		if !hasData {
			return Change{}, invalidTextError(text)
		}
		value, parseErr := tempo.Parse(data)
		if parseErr != nil {
			return Change{}, nestedTextError(text, parseErr)
		}
		if kind == ChangeMain {
			c, err = NewMain(at, value)
		} else {
			c, err = NewJump(at, value)
		}
	case ChangeGradual:
		if !hasData {
			return Change{}, invalidTextError(text)
		}
		law, parseErr := ParseLaw(data)
		if parseErr != nil {
			return Change{}, invalidTextError(text)
		}
		c, err = NewGradual(at, law)
	case ChangeATempo:
		if hasData {
			return Change{}, invalidTextError(text)
		}
		c, err = NewATempo(at)
	case ChangeTempoI:
		if hasData {
			return Change{}, invalidTextError(text)
		}
		c, err = NewTempoI(at)
	case ChangeModulation:
		unitText, oldText, found := strings.Cut(data, "@")
		if !hasData || !found {
			return Change{}, invalidTextError(text)
		}
		unit, parseErr := duration.ParseWritten(unitText)
		if parseErr != nil {
			return Change{}, nestedTextError(text, parseErr)
		}
		old, parseErr := duration.Parse(oldText)
		if parseErr != nil {
			return Change{}, nestedTextError(text, parseErr)
		}
		c, err = NewModulation(at, unit, old)
	}
	if err != nil {
		return Change{}, err
	}
	return c.WithSlot(slot)
}

// MustParseChange is like ParseChange but panics with its error.
func MustParseChange(text string) Change { return must(ParseChange(text)) }

// ParseExtension parses one complete hold record.
func ParseExtension(text string) (Extension, error) {
	rest, ok := strings.CutPrefix(text, "hold@")
	if !ok {
		return Extension{}, invalidTextError(text)
	}
	addressText, secondsText, ok := splitAddressText(rest)
	if !ok {
		return Extension{}, invalidTextError(text)
	}
	at, err := ParseAddress(addressText)
	if err != nil {
		return Extension{}, err
	}
	seconds, err := tempo.ParseSeconds(secondsText)
	if err != nil {
		return Extension{}, nestedTextError(text, err)
	}
	return NewExtension(at, seconds)
}

// MustParseExtension is like ParseExtension but panics with its error.
func MustParseExtension(text string) Extension { return must(ParseExtension(text)) }

// Parse parses a complete map. A measure map may be parsed without a grid;
// passing one additionally validates every address against it.
func Parse(text string, grid ...measure.Grid) (Map, error) {
	const op = "Parse"
	if len(grid) > 1 {
		return Map{}, simpleError(ErrInvalidArgument, op, "too many grids")
	}
	if len(grid) == 1 && !grid[0].IsSet() {
		return Map{}, dependencyArgumentError(op, measure.ErrUnset)
	}
	body, ok := strings.CutPrefix(text, "tempomap[")
	if !ok || !strings.HasSuffix(body, "]") {
		return Map{}, invalidTextError(text)
	}
	body = strings.TrimSuffix(body, "]")
	if body == "" {
		return Map{}, simpleError(ErrNoInitialTempo, op, "empty map")
	}
	items := strings.Split(body, ";")
	changes := make([]Change, 0, len(items))
	extensions := make([]Extension, 0)
	seenExtension := false
	for _, item := range items {
		if strings.HasPrefix(item, "hold@") {
			seenExtension = true
			extension, err := ParseExtension(item)
			if err != nil {
				return Map{}, err
			}
			extensions = append(extensions, extension)
			continue
		}
		if seenExtension {
			return Map{}, invalidTextError(text)
		}
		change, err := ParseChange(item)
		if err != nil {
			return Map{}, err
		}
		changes = append(changes, change)
	}
	if len(changes) == 0 {
		return Map{}, simpleError(ErrNoInitialTempo, op, "map has no initial change")
	}
	addressing := changes[0].at.addressing
	if addressing == PositionAddressing && len(grid) != 0 {
		return Map{}, simpleError(ErrInvalidArgument, op, "positional map takes no grid")
	}
	m := Map{addressing: addressing, changes: changes, extensions: extensions}
	idx, err := m.buildIndex(grid)
	if err != nil {
		return Map{}, err
	}
	if addressing == PositionAddressing {
		m.index = idx
	}
	return m, nil
}

// MustParse is like Parse but panics with its error.
func MustParse(text string, grid ...measure.Grid) Map { return must(Parse(text, grid...)) }
