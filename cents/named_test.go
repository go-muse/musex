package cents_test

import (
	"cmp"
	"testing"

	"github.com/go-muse/musex/cents"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// namedSpec is a named value: its function, name, definition and the
// column "Cents (6 digits)" of the model table.
type namedSpec struct {
	name  string
	get   func() cents.Cents
	def   comp
	cents string
}

// namedSpecs returns the 19 named small values in increasing order
// (table of the domain model; values checked with Python decimal).
func namedSpecs() []namedSpec {
	return []namedSpec{
		{"schisma", cents.Schisma, ratio(32805, 32768), "1.953721"},
		{"septimal kleisma", cents.SeptimalKleisma, ratio(225, 224), "7.711523"},
		{"kleisma", cents.Kleisma, ratio(15625, 15552), "8.107279"},
		{"Ptolemy's comma", cents.PtolemysComma, ratio(100, 99), "17.399484"},
		{"diaschisma", cents.Diaschisma, ratio(2048, 2025), "19.552569"},
		{"syntonic comma", cents.SyntonicComma, ratio(81, 80), "21.506290"},
		{"Holdrian comma", cents.HoldrianComma, stepC(1, 53), "22.641509"},
		{"Pythagorean comma", cents.PythagoreanComma, ratio(531441, 524288), "23.460010"},
		{"septimal comma", cents.SeptimalComma, ratio(64, 63), "27.264092"},
		{"lesser diesis", cents.LesserDiesis, ratio(128, 125), "41.058858"},
		{"septimal quarter tone", cents.SeptimalQuarterTone, ratio(36, 35), "48.770381"},
		{"undecimal comma", cents.UndecimalComma, ratio(33, 32), "53.272943"},
		{"greater diesis", cents.GreaterDiesis, ratio(648, 625), "62.565148"},
		{"tridecimal comma", cents.TridecimalComma, ratio(27, 26), "65.337341"},
		{"minor chroma", cents.MinorChroma, ratio(25, 24), "70.672427"},
		{"limma", cents.Limma, ratio(256, 243), "90.224996"},
		{"major chroma", cents.MajorChroma, ratio(135, 128), "92.178716"},
		{"apotome", cents.Apotome, ratio(2187, 2048), "113.685006"},
		{"large limma", cents.LargeLimma, ratio(27, 25), "133.237575"},
	}
}

// unitSpecs returns the 8 named units in increasing order (architecture,
// "Named values and units").
func unitSpecs() []namedSpec {
	return []namedSpec{
		{"MIDI tuning unit", cents.MIDITuningUnit, ratC(25, 4096), "0.006104"},
		{"jot", cents.Jot, ratC(1200, 30103), "0.039863"},
		{"millioctave", cents.Millioctave, ratC(6, 5), "1.200000"},
		{"centitone", cents.Centitone, ratC(2, 1), "2.000000"},
		// 1200*log2(10^(1/1000)) = 3.9863137...: Python decimal, 120 digits.
		{"savart", cents.Savart, ratC(6, 5).withExp(5, rat(1, 1000)), "3.986314"},
		{"eptameride", cents.Eptameride, ratC(1200, 301), "3.986711"},
		{"meride", cents.Meride, ratC(1200, 43), "27.906977"},
		{"semitone", cents.Semitone, ratC(100, 1), "100.000000"},
	}
}

// K-8: the 19 named values equal their definitions, have the decimal view
// of the model table, are pairwise different and strictly increasing; a
// repeated call gives an equal value.
func TestNamedValues(t *testing.T) {
	t.Parallel()

	specs := namedSpecs()
	require.Len(t, specs, 19)

	for _, s := range specs {
		t.Run(s.name, func(t *testing.T) {
			t.Parallel()

			v := s.get()
			assertComp(t, s.def, v)
			assert.Equal(t, s.cents, v.Decimal(6, 0))
			assert.Equal(t, s.cents, decimalOracle(s.def, 6, 0), "oracle")
			assertCents(t, v, s.get(), "repeated call")
		})
	}

	t.Run("definitions", func(t *testing.T) {
		t.Parallel()

		defs := []cents.Cents{
			noErr(cents.FromRatio(32805, 32768))(t), noErr(cents.FromRatio(225, 224))(t), noErr(cents.FromRatio(15625, 15552))(t),
			noErr(cents.FromRatio(100, 99))(t), noErr(cents.FromRatio(2048, 2025))(t), noErr(cents.FromRatio(81, 80))(t),
			noErr(cents.Step(1, 53))(t), noErr(cents.FromRatio(531441, 524288))(t), noErr(cents.FromRatio(64, 63))(t),
			noErr(cents.FromRatio(128, 125))(t), noErr(cents.FromRatio(36, 35))(t), noErr(cents.FromRatio(33, 32))(t),
			noErr(cents.FromRatio(648, 625))(t), noErr(cents.FromRatio(27, 26))(t), noErr(cents.FromRatio(25, 24))(t),
			noErr(cents.FromRatio(256, 243))(t), noErr(cents.FromRatio(135, 128))(t), noErr(cents.FromRatio(2187, 2048))(t),
			noErr(cents.FromRatio(27, 25))(t),
		}

		for i, s := range specs {
			assertCents(t, defs[i], s.get(), s.name)
		}
	})

	t.Run("pairwise different and increasing", func(t *testing.T) {
		t.Parallel()

		for i, a := range specs {
			for j, b := range specs {
				assert.Equal(t, i == j, a.get() == b.get(), "%s, %s", a.name, b.name)
				assert.Equal(t, cmp.Compare(i, j), a.get().Compare(b.get()), "%s vs %s", a.name, b.name)
			}
		}
	})

	t.Run("list", func(t *testing.T) {
		t.Parallel()

		list := cents.NamedValues()
		require.Len(t, list, 19)
		assert.Equal(t, 19, cap(list))

		for i, s := range specs {
			assert.Equal(t, s.name, list[i].Name, "position %d", i)
			assertCents(t, s.get(), list[i].Value, s.name)
		}
	})
}

// C-21: the 8 named units, their values and the list in increasing order.
func TestNamedUnits(t *testing.T) {
	t.Parallel()

	specs := unitSpecs()

	for _, s := range specs {
		t.Run(s.name, func(t *testing.T) {
			t.Parallel()

			v := s.get()
			assertComp(t, s.def, v)
			assert.Equal(t, s.cents, v.Decimal(6, 0))
			assertCents(t, v, s.get(), "repeated call")
		})
	}

	t.Run("exact decimal of the MIDI tuning unit", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "0.006103515625", cents.MIDITuningUnit().String())
		assertCents(t, noErr(cents.Step(1, 196608))(t), cents.MIDITuningUnit())
		assertCents(t, noErr(cents.Step(1, 30103))(t), cents.Jot())
		assertCents(t, noErr(cents.Step(1, 1000))(t), cents.Millioctave())
		assertCents(t, noErr(cents.Step(1, 600))(t), cents.Centitone())
		assertCents(t, noErr(cents.Step(1, 301))(t), cents.Eptameride())
		assertCents(t, noErr(cents.Step(1, 43))(t), cents.Meride())
		assertCents(t, noErr(cents.Step(1, 12))(t), cents.Semitone())
		assertCents(t, noErr(cents.Parse("10^(1/1000)"))(t), cents.Savart())
	})

	t.Run("list", func(t *testing.T) {
		t.Parallel()

		list := cents.NamedUnits()
		require.Len(t, list, 8)
		assert.Equal(t, 8, cap(list))

		for i, s := range specs {
			assert.Equal(t, s.name, list[i].Name, "position %d", i)
			assertCents(t, s.get(), list[i].Value, s.name)

			if i > 0 {
				assert.Equal(t, 1, list[i].Value.Compare(list[i-1].Value), "increasing at %d", i)
			}
		}
	})
}
