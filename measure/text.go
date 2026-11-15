package measure

import (
	"math"
	"strconv"
	"strings"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/meter"
)

// unsetGridText is the string of the unset grid.
const unsetGridText = "Grid(unset)"

// String returns the canonical text of g, as described for Parse, such as
// "4/4;3:3/4;4:length:1/2;bars:8", or "Grid(unset)" for the unset grid.
func (g Grid) String() string {
	if !g.IsSet() {
		return unsetGridText
	}
	return string(g.appendText(nil))
}

// appendText appends the canonical text of the set grid g to dst.
func (g Grid) appendText(dst []byte) []byte {
	dst = append(dst, g.marks[0].Meter.String()...)
	if g.pickup.IsSet() {
		dst = append(dst, ";pickup:"...)
		dst = append(dst, g.pickup.String()...)
	}
	mi, li := 1, 0
	for mi < len(g.marks) || li < len(g.lengths) {
		if mi < len(g.marks) && (li == len(g.lengths) || g.marks[mi].Bar <= g.lengths[li].Bar) {
			dst = append(dst, ';')
			dst = append(dst, g.marks[mi].String()...)
			mi++
			continue
		}
		dst = append(dst, ';')
		dst = append(dst, g.lengths[li].String()...)
		li++
	}
	if g.count > 0 {
		dst = append(dst, ";bars:"...)
		dst = strconv.AppendInt(dst, int64(g.count), 10)
	}
	return dst
}

// MarshalText returns the canonical text of g, as String does.
//
// Errors: ErrUnset for the unset grid.
func (g Grid) MarshalText() ([]byte, error) {
	if !g.IsSet() {
		return nil, errUnsetGrid("Grid.MarshalText")
	}
	return g.appendText(nil), nil
}

// AppendText appends the canonical text of g to dst and returns the extended
// slice.
//
// Errors: ErrUnset for the unset grid; dst is then returned unchanged.
func (g Grid) AppendText(dst []byte) ([]byte, error) {
	if !g.IsSet() {
		return dst, errUnsetGrid("Grid.AppendText")
	}
	return g.appendText(dst), nil
}

// UnmarshalText parses text as Parse does and stores the result in *g. On an
// error *g is not changed.
//
// Errors: ErrInvalidArgument for a nil receiver; otherwise those of Parse.
func (g *Grid) UnmarshalText(text []byte) error {
	if g == nil {
		return errInvalid("Grid.UnmarshalText", "nil receiver")
	}
	v, err := Parse(string(text))
	if err != nil {
		return err
	}
	*g = v
	return nil
}

// isCount reports whether s is a decimal number without a leading zero,
// at least 1.
func isCount(s string) bool {
	if s == "" || s[0] < '1' || s[0] > '9' {
		return false
	}
	for i := 1; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// parseCount returns the value of the decimal digits s and true, or 0 and
// false if the value is greater than math.MaxInt.
func parseCount(s string) (int, bool) {
	if len(s) > len(strconv.Itoa(math.MaxInt)) {
		return 0, false
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, false
	}
	return int(n), true
}

// textItem is a record of the text of a grid after a bar number.
type textItem struct {
	barText string
	bar     int
	isMark  bool
	value   string // the text of the meter or of the duration
	end     bool   // the explicit length is aligned to the end
}

// Parse parses the canonical text of a grid:
//
//	grid     = meter [ ";pickup:" duration ] *( ";" item ) [ ";bars:" count ]
//	item     = count ":" meter                          ; time signature from bar k >= 2
//	         / count ":length:" duration [ ":end" ]     ; explicit length of bar k >= 1
//	count    = nzdigit *digit
//
// The first meter is the mark of bar 1. A meter is parsed by meter.Parse and a
// duration by duration.Parse, so their non-canonical forms are accepted:
// "3/4;pickup:2/8" is the same grid as "3/4;pickup:1/4". The text of a
// meter or a duration contains no ";" or ":". Records follow in the
// canonical order: bar numbers strictly increase, except that the mark of a
// bar precedes its explicit length. Spaces are not accepted.
//
// Errors, in this order: ErrInvalidText for text outside the grammar,
// including "Grid(unset)"; ErrOutOfRange for the first bar number or number
// of bars greater than math.MaxInt; ErrInvalidText for records outside the
// canonical order; for nested texts from left to right, ErrOutOfRange
// together with those of meter.ErrOutOfRange, duration.ErrOutOfRange, and
// fraction.ErrOutOfRange that the failure of the nested parser wraps, or
// ErrInvalidText for any other failure; ErrLongerThanMeter for the pickup,
// then for explicit lengths aligned to the end in the order of the text;
// ErrOutsideGrid for a number of bars less than the number of a bar with a
// record. The message quotes the text as described in the package
// documentation.
//
// Example: Parse("4/4;3:3/4;4:length:1/2;bars:8") is the grid of 4/4 with
// 3/4 from bar 3, bar 4 of the length 1/2, and 8 bars.
func Parse(text string) (Grid, error) {
	const op = "Parse"
	fields := strings.Split(text, ";")
	meterText := fields[0]
	if meterText == "" || strings.Contains(meterText, ":") {
		return Grid{}, parseError(classInvalidText, text)
	}
	rest := fields[1:]
	pickupText := ""
	if len(rest) > 0 {
		if v, ok := strings.CutPrefix(rest[0], "pickup:"); ok {
			if v == "" || strings.Contains(v, ":") {
				return Grid{}, parseError(classInvalidText, text)
			}
			pickupText = v
			rest = rest[1:]
		}
	}
	countText := ""
	if n := len(rest); n > 0 {
		if v, ok := strings.CutPrefix(rest[n-1], "bars:"); ok {
			if !isCount(v) {
				return Grid{}, parseError(classInvalidText, text)
			}
			countText = v
			rest = rest[:n-1]
		}
	}
	items := make([]textItem, 0, len(rest))
	for _, f := range rest {
		parts := strings.Split(f, ":")
		if !isCount(parts[0]) {
			return Grid{}, parseError(classInvalidText, text)
		}
		it := textItem{barText: parts[0]}
		switch {
		case len(parts) == 2 && parts[1] != "":
			it.isMark = true
			it.value = parts[1]
		case len(parts) == 3 && parts[1] == "length" && parts[2] != "":
			it.value = parts[2]
		case len(parts) == 4 && parts[1] == "length" && parts[2] != "" && parts[3] == "end":
			it.value = parts[2]
			it.end = true
		default:
			return Grid{}, parseError(classInvalidText, text)
		}
		items = append(items, it)
	}
	// Numbers greater than math.MaxInt, from left to right.
	for i := range items {
		k, ok := parseCount(items[i].barText)
		if !ok {
			return Grid{}, parseError(classNumberRange, text)
		}
		items[i].bar = k
	}
	count := 0
	if countText != "" {
		k, ok := parseCount(countText)
		if !ok {
			return Grid{}, parseError(classNumberRange, text)
		}
		count = k
	}
	// The canonical order.
	for i, it := range items {
		if it.isMark && it.bar < 2 {
			return Grid{}, parseError(classInvalidText, text)
		}
		if i == 0 {
			continue
		}
		prev := items[i-1]
		if it.bar < prev.bar || (it.bar == prev.bar && (it.isMark || !prev.isMark)) {
			return Grid{}, parseError(classInvalidText, text)
		}
	}
	// Nested texts from left to right.
	m1, err := meter.Parse(meterText)
	if err != nil {
		return Grid{}, parseError(nestedClass(err), text)
	}
	var pickup duration.Duration
	if pickupText != "" {
		pickup, err = duration.Parse(pickupText)
		if err != nil {
			return Grid{}, parseError(nestedClass(err), text)
		}
	}
	marks := []Mark{{Bar: 1, Meter: m1}}
	var lengths []ExplicitLength
	for _, it := range items {
		if it.isMark {
			m, err := meter.Parse(it.value)
			if err != nil {
				return Grid{}, parseError(nestedClass(err), text)
			}
			marks = append(marks, Mark{Bar: it.bar, Meter: m})
			continue
		}
		d, err := duration.Parse(it.value)
		if err != nil {
			return Grid{}, parseError(nestedClass(err), text)
		}
		align := AlignStart
		if it.end {
			align = AlignEnd
		}
		lengths = append(lengths, ExplicitLength{Bar: it.bar, Duration: d, Alignment: align})
	}
	// The value: the rules of the record, then the number of bars.
	g, err := build(op, pickup, marks, lengths, 0)
	if err != nil {
		return Grid{}, parseError(classLonger, text)
	}
	if count == 0 {
		return g, nil
	}
	if g.lastRecordBar() > count {
		return Grid{}, parseError(classOutsideGrid, text)
	}
	g, _ = build(op, pickup, marks, lengths, count)
	return g, nil
}

// MustParse is like Parse but panics with the error of Parse.
func MustParse(text string) Grid {
	return must(Parse(text))
}
