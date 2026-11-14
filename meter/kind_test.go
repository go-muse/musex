package meter_test

import (
	"testing"

	"github.com/go-muse/musex/meter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// kindCase is the expected kind of a meter in both traditions (ИР-9, ИР-10,
// ИР-10а, ИР-11). A zero kind means E-3 for Kind, RussianKind, BeatDivision,
// KindName, and KindRussianName; beats −1 means E-3 for BeatCount; a zero
// division means E-5 for BeatDivision.
type kindCase struct {
	name, text string
	kind       meter.Kind
	beats      int
	russian    meter.RussianKind
	div, sub   int
	en, ru     string // "" means no name
}

const (
	simple       = meter.KindSimple
	compound     = meter.KindCompound
	asymmetrical = meter.KindAsymmetrical
	undetermined = meter.KindUndetermined
	rSimple      = meter.RussianSimple
	rUniform     = meter.RussianCompoundUniform
	rMixed       = meter.RussianCompoundMixed
	rUndet       = meter.RussianUndetermined
)

// kindCases are КП-9 to КП-13, КП-26а, КП-30, КП-32, КП-41, КП-50, КП-51 and
// further cases of the rules ИР-9 to ИР-11.
var kindCases = []kindCase{
	// КП-9.
	{"КП-9 2/4", "2/4", simple, 2, rSimple, 2, 4, "simple duple", "простой двухдольный"},
	{"КП-9 3/4", "3/4", simple, 3, rSimple, 2, 4, "simple triple", "простой трёхдольный"},
	{"КП-9 4/4", "4/4", simple, 4, rUniform, 2, 4, "simple quadruple", "простой четырёхдольный"},
	{"КП-9 2/2", "2/2", simple, 2, rSimple, 2, 4, "simple duple", "простой двухдольный"},
	{"КП-9 3/8", "3/8", simple, 3, rSimple, 2, 4, "simple triple", "простой трёхдольный"},
	{"КП-9 6/8", "6/8", compound, 2, rUniform, 3, 6, "compound duple", "сложный двухдольный"},
	{"КП-9 9/8", "9/8", compound, 3, rUniform, 3, 6, "compound triple", "сложный трёхдольный"},
	{"КП-9 12/8", "12/8", compound, 4, rUniform, 3, 6, "compound quadruple", "сложный четырёхдольный"},
	{"КП-9 6/4", "6/4", compound, 2, rUniform, 3, 6, "compound duple", "сложный двухдольный"},
	{"КП-9 15/8", "15/8", compound, 5, rUniform, 3, 6, "compound", "сложный"},

	// КП-10, КП-11.
	{"КП-10 5/8[2+3]", "5/8[2+3]", asymmetrical, 2, rMixed, 0, 0, "asymmetrical", "асимметричный"},
	{"КП-10 7/8[3+2+2]", "7/8[3+2+2]", asymmetrical, 3, rMixed, 0, 0, "asymmetrical", "асимметричный"},
	{"КП-10 (3+2+3)/8", "(3+2+3)/8", asymmetrical, 3, rMixed, 0, 0, "asymmetrical", "асимметричный"},
	{"КП-10 5/4[3+2]", "5/4[3+2]", asymmetrical, 2, rMixed, 0, 0, "asymmetrical", "асимметричный"},
	{"КП-10 5/8", "5/8", asymmetrical, -1, rMixed, 0, 0, "asymmetrical", "асимметричный"},
	{"КП-10 7/8", "7/8", asymmetrical, -1, rMixed, 0, 0, "asymmetrical", "асимметричный"},
	{"КП-10 5/8+2/8", "5/8+2/8", asymmetrical, -1, rMixed, 0, 0, "asymmetrical", "асимметричный"},
	{"КП-10 6/8[2+2+2]", "6/8[2+2+2]", simple, 3, rUniform, 2, 4, "simple triple", "простой трёхдольный"},
	{"КП-10 (2+2+2+2)/8", "(2+2+2+2)/8", simple, 4, rUniform, 2, 4, "simple quadruple", "простой четырёхдольный"},
	{"КП-10 (2+2)/4", "(2+2)/4", simple, 4, rUniform, 2, 4, "simple quadruple", "простой четырёхдольный"},
	{"КП-10 (3+3)/8", "(3+3)/8", compound, 2, rUniform, 3, 6, "compound duple", "сложный двухдольный"},
	{"КП-10 2/4+3/8", "2/4+3/8", asymmetrical, 2, rMixed, 0, 0, "asymmetrical", "асимметричный"},
	{"КП-10 2/4+2/4", "2/4+2/4", simple, 4, rUniform, 2, 4, "simple quadruple", "простой четырёхдольный"},
	{"КП-10 ИР-10а 2/4+2/8", "2/4+2/8", asymmetrical, 2, rMixed, 0, 0, "asymmetrical", "асимметричный"},
	{"КП-10 3/8+3/4", "3/8+3/4", asymmetrical, 2, rMixed, 0, 0, "asymmetrical", "асимметричный"},

	// КП-12.
	{"КП-12 1/4", "1/4", undetermined, 1, rUndet, 0, 0, "", ""},
	{"КП-12 (4+4)/8", "(4+4)/8", undetermined, 2, rUndet, 0, 0, "", ""},
	{"КП-12 4/4[3+1]", "4/4[3+1]", asymmetrical, 2, rUndet, 0, 0, "asymmetrical", "асимметричный"},
	{"КП-12 5/8[5]", "5/8[5]", undetermined, 1, rUndet, 0, 0, "", ""},

	// КП-30: no grouping and no kind without it.
	{"КП-30 8/8", "8/8", 0, -1, 0, 0, 0, "", ""},
	{"КП-30 10/8", "10/8", 0, -1, 0, 0, 0, "", ""},
	{"КП-30 5/8+2/4", "5/8+2/4", 0, -1, 0, 0, 0, "", ""},
	{"14/8", "14/8", 0, -1, 0, 0, 0, "", ""},

	// Other cases of the rules.
	{"11/8 without grouping", "11/8", asymmetrical, -1, rMixed, 0, 0, "asymmetrical", "асимметричный"},
	{"13/8 without grouping", "13/8", asymmetrical, -1, rMixed, 0, 0, "asymmetrical", "асимметричный"},
	{"25/8 without grouping", "25/8", asymmetrical, -1, rMixed, 0, 0, "asymmetrical", "асимметричный"},
	{"C", "C", simple, 4, rUniform, 2, 4, "simple quadruple", "простой четырёхдольный"},
	{"C|", "C|", simple, 2, rSimple, 2, 4, "simple duple", "простой двухдольный"},
	{"C[3+1]", "C[3+1]", asymmetrical, 2, rUndet, 0, 0, "asymmetrical", "асимметричный"},
	{"КП-50 4/3", "4/3", simple, 4, rUniform, 2, 4, "simple quadruple", "простой четырёхдольный"},
	{"КП-41 3/1", "3/1", simple, 3, rSimple, 2, 4, "simple triple", "простой трёхдольный"},
	{"1/1", "1/1", undetermined, 1, rUndet, 0, 0, "", ""},
	{"КП-51 6/6", "6/6", compound, 2, rUniform, 3, 6, "compound duple", "сложный двухдольный"},
	{"КП-51 3/6", "3/6", simple, 3, rSimple, 2, 4, "simple triple", "простой трёхдольный"},
	{"simple of six beats", "12/8[2+2+2+2+2+2]", simple, 6, rUniform, 2, 4, "simple", "простой"},
	{"6/4[2+2+2]", "6/4[2+2+2]", simple, 3, rUniform, 2, 4, "simple triple", "простой трёхдольный"},
	{"9/8[2+2+2+3]", "9/8[2+2+2+3]", asymmetrical, 4, rMixed, 0, 0, "asymmetrical", "асимметричный"},
	{"8/8[4+4]", "8/8[4+4]", undetermined, 2, rUndet, 0, 0, "", ""},
	{"8/8[3+3+2]", "8/8[3+3+2]", asymmetrical, 3, rMixed, 0, 0, "asymmetrical", "асимметричный"},
	{"2/4[1+1]", "2/4[1+1]", undetermined, 2, rUndet, 0, 0, "", ""},
	{"3/4[2+1]", "3/4[2+1]", asymmetrical, 2, rUndet, 0, 0, "asymmetrical", "асимметричный"},
	{"(3+3)/6", "(3+3)/6", compound, 2, rUniform, 3, 6, "compound duple", "сложный двухдольный"},
	{"ИР-10а 2/4+2/3", "2/4+2/3", asymmetrical, 2, rMixed, 0, 0, "asymmetrical", "асимметричный"},
	{"3/8+3/8", "3/8+3/8", compound, 2, rUniform, 3, 6, "compound duple", "сложный двухдольный"},
	{"1/4+1/4", "1/4+1/4", undetermined, 2, rUndet, 0, 0, "", ""},
	{"18/8", "18/8", compound, 6, rUniform, 3, 6, "compound", "сложный"},
	{"2/4+2/4+2/4", "2/4+2/4+2/4", simple, 3, rUniform, 2, 4, "simple triple", "простой трёхдольный"},
}

// TestKind covers О-9, ИР-9, ИР-10, ИР-10а, П-8, Т-6, КП-9 to КП-12, КП-30,
// КП-50, КП-51.
func TestKind(t *testing.T) {
	t.Parallel()

	for _, c := range kindCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			m := mt(c.text)

			kind, err := m.Kind()
			russian, rerr := m.RussianKind()

			if c.kind == 0 {
				assertErr(t, err, meter.ErrGroupingRequired)
				assert.Equal(t, meter.Kind(0), kind)
				assertErr(t, rerr, meter.ErrGroupingRequired)
				assert.Equal(t, meter.RussianKind(0), russian)
			} else {
				require.NoError(t, err)
				assert.Equal(t, c.kind, kind, "Kind")
				require.NoError(t, rerr)
				assert.Equal(t, c.russian, russian, "RussianKind")
			}

			beats, err := m.BeatCount()
			if c.beats < 0 {
				assertErr(t, err, meter.ErrGroupingRequired)
			} else {
				require.NoError(t, err)
				assert.Equal(t, c.beats, beats, "the number of beats")
			}
		})
	}
}

// TestBeatDivision covers О-9, ИР-11, Т-6, КП-13, КП-30, КП-32.
func TestBeatDivision(t *testing.T) {
	t.Parallel()

	for _, c := range kindCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			div, sub, err := mt(c.text).BeatDivision()

			switch {
			case c.kind == 0:
				assertErr(t, err, meter.ErrGroupingRequired)
			case c.div == 0:
				assertErr(t, err, meter.ErrUndefined)
			default:
				require.NoError(t, err)
				assert.Equal(t, c.div, div)
				assert.Equal(t, c.sub, sub)

				return
			}

			assert.Zero(t, div)
			assert.Zero(t, sub)
		})
	}
}

// TestKindName covers О-20, П-13, Т-16, КП-26а and architecture 4.14: the full
// name of the kind in the English-language tradition.
func TestKindName(t *testing.T) {
	t.Parallel()

	for _, c := range kindCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			m := mt(c.text)

			en, enOK, err := m.KindName()
			ru, ruOK, rerr := m.KindRussianName()

			if c.kind == 0 {
				assertErr(t, err, meter.ErrGroupingRequired)
				assertErr(t, rerr, meter.ErrGroupingRequired)
				assert.Empty(t, en)
				assert.Empty(t, ru)
				assert.False(t, enOK)
				assert.False(t, ruOK)

				return
			}

			require.NoError(t, err)
			require.NoError(t, rerr)
			assert.Equal(t, c.en, en)
			assert.Equal(t, c.en != "", enOK)
			assert.Equal(t, c.ru, ru)
			assert.Equal(t, c.ru != "", ruOK)
		})
	}
}

// TestKindConsistency covers ИР-9 to ИР-11 on the corpus: BeatDivision is
// defined exactly for simple and compound kinds, both kinds fail together
// with E-3, and a kind without beats occurs only without grouping.
func TestKindConsistency(t *testing.T) {
	t.Parallel()

	for _, m := range corpus() {
		t.Run(m.String(), func(t *testing.T) {
			t.Parallel()

			kind, err := m.Kind()
			_, rerr := m.RussianKind()
			div, sub, derr := m.BeatDivision()
			beats, berr := m.BeatCount()
			grouped := noErr(m.HasGrouping())

			assert.Equal(t, sentinelName(err), sentinelName(rerr), "both kinds fail together")

			if err != nil {
				assertErr(t, err, meter.ErrGroupingRequired)
				assertErr(t, derr, meter.ErrGroupingRequired)
				assert.False(t, grouped)

				return
			}

			assert.True(t, kind.IsValid())

			switch kind {
			case meter.KindSimple:
				require.NoError(t, derr)
				assert.Equal(t, [2]int{2, 4}, [2]int{div, sub})
			case meter.KindCompound:
				require.NoError(t, derr)
				assert.Equal(t, [2]int{3, 6}, [2]int{div, sub})
			default:
				assertErr(t, derr, meter.ErrUndefined)
			}

			if grouped {
				require.NoError(t, berr)
				assert.Positive(t, beats)
			} else {
				assert.Equal(t, meter.KindAsymmetrical, kind, "ИР-9 rule 1")
				assertErr(t, berr, meter.ErrGroupingRequired)
			}
		})
	}
}
