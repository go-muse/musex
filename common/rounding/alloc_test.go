package rounding_test

import (
	"testing"

	"github.com/go-muse/musex/common/rounding"

	"github.com/stretchr/testify/assert"
)

var (
	allocBool     bool
	allocMode     rounding.Mode
	allocModes    []rounding.Mode
	allocString   string
	allocBytes    []byte
	allocNeighbor rounding.Neighbor
	allocInt      int
	allocInt2     int
	allocFloat    float64
	allocErr      error
)

func TestAllocationBudget(t *testing.T) {
	buffer := make([]byte, 0, 32)
	tests := []struct {
		name string
		want float64
		call func()
	}{
		{"IsValid", 0, func() { allocBool = rounding.HalfEven.IsValid() }},
		{"Equal", 0, func() { allocBool = rounding.HalfEven.Equal(rounding.HalfUp) }},
		{"IsDirected", 0, func() { allocBool = rounding.Down.IsDirected() }},
		{"IsNearest", 0, func() { allocBool = rounding.HalfEven.IsNearest() }},
		{"Mirror", 0, func() { allocMode = rounding.HalfUp.Mirror() }},
		{"Side_IsValid", 0, func() { allocBool = rounding.Positive.IsValid() }},
		{"Side_String", 0, func() { allocString = rounding.Positive.String() }},
		{"Neighbor_String", 0, func() { allocString = rounding.Lower.String() }},
		{"Mode_String", 0, func() { allocString = rounding.HalfEven.String() }},
		{"Mode_String_unset", 0, func() { allocString = rounding.Mode(0).String() }},
		{"Name", 0, func() { allocString, allocErr = rounding.HalfEven.Name() }},
		{"MustName", 0, func() { allocString = rounding.HalfEven.MustName() }},
		{"Parse", 0, func() { allocMode, allocErr = rounding.Parse("half-even") }},
		{"MustParse", 0, func() { allocMode = rounding.MustParse("half-even") }},
		{"MarshalText", 1, func() { allocBytes, allocErr = rounding.HalfEven.MarshalText() }},
		{"AppendText", 0, func() { allocBytes, allocErr = rounding.HalfEven.AppendText(buffer) }},
		{"Modes", 1, func() { allocModes = rounding.Modes() }},
		{"Choose", 0, func() { allocNeighbor, allocErr = rounding.HalfEven.Choose(0, rounding.Positive, false) }},
		{"ChooseIndex", 0, func() { allocInt, allocErr = rounding.HalfEven.ChooseIndex(3, 0) }},
		{"QuoRem", 0, func() { allocInt, allocInt2, allocErr = rounding.HalfEven.QuoRem(7, 2) }},
		{"IntMultiple", 0, func() { allocInt, allocErr = rounding.HalfEven.IntMultiple(1200, 480) }},
		{"Float", 0, func() { allocFloat, allocErr = rounding.HalfEven.Float(2.5) }},
		{"FloatInt", 0, func() { allocInt, allocErr = rounding.HalfEven.FloatInt(2.5) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, testing.AllocsPerRun(100, tt.call))
		})
	}
}
