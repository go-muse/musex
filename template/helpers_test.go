package template_test

import (
	"math/big"
	"strings"
	"testing"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func intervals(text string) []interval.Interval {
	parts := strings.Fields(text)
	result := make([]interval.Interval, len(parts))
	for i, part := range parts {
		result[i] = interval.MustParse(part)
	}
	return result
}

func closed(text string) template.Template {
	return template.MustClosedFromIntervals(intervals(text))
}

func opened(text string) template.Template {
	return template.MustOpenFromIntervals(intervals(text))
}

func names(text string) []spelling.Spelling {
	parts := strings.Fields(text)
	result := make([]spelling.Spelling, len(parts))
	replacer := strings.NewReplacer("𝄪", "+2", "𝄫", "-2", "♯", "+1", "♭", "-1", "♮", "")
	for i, part := range parts {
		result[i] = spelling.MustParse(replacer.Replace(part))
	}
	return result
}

func intervalTexts(values []interval.Interval) []string {
	result := make([]string, len(values))
	for i, value := range values {
		result[i] = value.String()
	}
	return result
}

func entryCodes(entries []template.Entry) []string {
	result := make([]string, len(entries))
	for i, entry := range entries {
		result[i] = entry.Code()
	}
	return result
}

func rotationCodes(rotations []template.Rotation) []string {
	result := make([]string, len(rotations))
	for i, rotation := range rotations {
		result[i] = rotation.Entry.Code()
	}
	return result
}

func rotationPlaces(rotations []template.Rotation) []int {
	result := make([]int, len(rotations))
	for i, rotation := range rotations {
		result[i] = rotation.Place
	}
	return result
}

func requireInterval(t testing.TB, want string, got interval.Interval) {
	t.Helper()
	require.True(t, got.Equal(interval.MustParse(want)), "interval = %s, want %s", got, want)
}

func requireTemplate(t testing.TB, want, got template.Template) {
	t.Helper()
	require.True(t, got.Equal(want), "template = %s, want %s", got, want)
}

func requirePanicValue(t testing.TB, want error, call func()) {
	t.Helper()
	require.PanicsWithValue(t, want, call)
}

// linearCoordinate is the independent (t,v) oracle from architecture 4.4.4.
// The tests use arbitrary precision only for expectations, never as product code.
type linearCoordinate struct {
	t *big.Int
	v *big.Rat
}

func newCoordinate(t int64, v *big.Rat) linearCoordinate {
	return linearCoordinate{t: big.NewInt(t), v: new(big.Rat).Set(v)}
}

func baseChromatic(size int64) int64 {
	base := [...]int64{0, 2, 4, 5, 7, 9, 11}
	return 12*((size-1)/7) + base[(size-1)%7]
}

func coordinateOf(value interval.Interval) linearCoordinate {
	size := int64(value.Size())
	sign := int64(1)
	if value.IsDescending() {
		sign = -1
	}
	t := big.NewInt(sign * (size - 1))
	num, den := value.Deviation().Fraction()
	v := new(big.Rat).SetFrac(big.NewInt(int64(num)), big.NewInt(int64(den)))
	v.Add(v, new(big.Rat).SetInt64(baseChromatic(size)))
	if sign < 0 {
		v.Neg(v)
	}
	return linearCoordinate{t: t, v: v}
}

func addCoordinates(left, right linearCoordinate) linearCoordinate {
	return linearCoordinate{
		t: new(big.Int).Add(left.t, right.t),
		v: new(big.Rat).Add(left.v, right.v),
	}
}

func subCoordinates(left, right linearCoordinate) linearCoordinate {
	return linearCoordinate{
		t: new(big.Int).Sub(left.t, right.t),
		v: new(big.Rat).Sub(left.v, right.v),
	}
}

func mulCoordinate(value linearCoordinate, multiplier int64) linearCoordinate {
	return linearCoordinate{
		t: new(big.Int).Mul(value.t, big.NewInt(multiplier)),
		v: new(big.Rat).Mul(value.v, new(big.Rat).SetInt64(multiplier)),
	}
}

func coordinatesOf(values []interval.Interval) []linearCoordinate {
	result := make([]linearCoordinate, len(values))
	for i, value := range values {
		result[i] = coordinateOf(value)
	}
	return result
}

func placeCoordinate(adjacent []linearCoordinate, isClosed bool, place int) linearCoordinate {
	if place == 1 {
		return newCoordinate(0, new(big.Rat))
	}
	prefix := newCoordinate(0, new(big.Rat))
	k := len(adjacent)
	if !isClosed {
		k++
	}
	remainder := (place - 1) % k
	for i := range remainder {
		prefix = addCoordinates(prefix, adjacent[i])
	}
	if !isClosed {
		return prefix
	}
	cycle := newCoordinate(0, new(big.Rat))
	for _, value := range adjacent {
		cycle = addCoordinates(cycle, value)
	}
	return addCoordinates(prefix, mulCoordinate(cycle, int64((place-1)/k)))
}

func requireCoordinate(t testing.TB, want linearCoordinate, got interval.Interval) {
	t.Helper()
	require.True(t, got.IsValid(), "invalid interval, want (%s,%s)", want.t, want.v)
	actual := coordinateOf(got)
	assert.Zero(t, actual.t.Cmp(want.t), "coordinate t = %s, want %s", actual.t, want.t)
	assert.Zero(t, actual.v.Cmp(want.v), "coordinate v = %s, want %s", actual.v, want.v)
}

type controlTemplate struct {
	name   string
	text   string
	isOpen bool
	value  func() template.Template
}

func controlTemplates() []controlTemplate {
	return []controlTemplate{
		{name: "ionian", text: "M2 M2 m2 M2 M2 M2 m2"},
		{name: "single octave", text: "P8"},
		{name: "single fifth", text: "P5"},
		{name: "fifth fourth", text: "P5 P4"},
		{name: "blues", text: "m3 M2 A1 m2 m3 M2"},
		{name: "diminished second", text: "d2 A7"},
		{name: "mixed sizes", text: "A3 dd2 A2 P4"},
		{name: "obikhod", text: "M2 M2 m2"},
		{name: "seven major seconds", text: "M2 M2 M2 M2 M2 M2 M2"},
		{name: "two fourths", text: "P4 P4"},
		{name: "two fifths", text: "P5 P5"},
		{name: "open trichord", text: "m3 M2", isOpen: true},
		{name: "open fifth", text: "P5", isOpen: true},
	}
}

func (c controlTemplate) template() template.Template {
	if c.value != nil {
		return c.value()
	}
	if c.isOpen {
		return opened(c.text)
	}
	return closed(c.text)
}
