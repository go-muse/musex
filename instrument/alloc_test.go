package instrument_test

import (
	"testing"

	"github.com/go-muse/musex/instrument"

	"github.com/stretchr/testify/assert"
)

// А 4.11, НФ-3: the listed operations do not allocate on the successful path
// with the standard sets and built-in systems.
//
// The tests are not parallel: testing.AllocsPerRun counts the allocations of
// the whole process.
func TestAllocations(t *testing.T) {
	g := instrument.MustFromParts("Lead", h0(), rng("C4", "G5"))
	twin := instrument.MustFromParts("Lead", h0(), rng("C4", "G5"))
	r := rng("C4", "G5")
	other := rng("C4", "G5")
	ps := preset("Lead", h0(), rng("C4", "G5"))
	psTwin := preset("Lead", h0(), rng("C4", "G5"))
	sys := trad()
	inside, above := p("E4"), p("A5")

	tests := []struct {
		name string
		f    func()
	}{
		{name: "Instrument.Settings", f: func() { _ = g.Settings() }},
		{name: "Instrument.IsValid", f: func() { _ = g.IsValid() }},
		{name: "Instrument.Name", f: func() { _ = g.Name() }},
		{name: "Instrument.Range", f: func() { _ = g.Range() }},
		{name: "Instrument.HasRange", f: func() { _ = g.HasRange() }},
		{name: "Instrument.Same", f: func() { _ = g.Same(twin) }},
		{name: "Instrument.Contains", f: func() { _, _ = g.Contains(inside) }},
		{name: "Instrument.Contains outside", f: func() { _, _ = g.Contains(above) }},
		{name: "Instrument.Locate", f: func() { _, _ = g.Locate(above) }},
		{name: "Instrument.Equal itself", f: func() { _ = g.Equal(g) }},
		{name: "Instrument.Equal another", f: func() { _ = g.Equal(twin) }},
		{name: "Range.Contains", f: func() { _, _ = r.Contains(sys, inside) }},
		{name: "Range.Locate", f: func() { _, _ = r.Locate(sys, above) }},
		{name: "Range.Check", f: func() { _ = r.Check(sys) }},
		{name: "Range.Equal", f: func() { _ = r.Equal(other) }},
		{name: "Preset.Equal", f: func() { _ = ps.Equal(psTwin) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.f() // warm up
			assert.Zero(t, testing.AllocsPerRun(100, tt.f))
		})
	}
}
