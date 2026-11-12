package pcseq_test

import (
	"testing"

	"github.com/go-muse/musex/pcseq"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/setclass"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSequenceCoreClassConstructorErrors(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name   string
		call   func() (pcseq.Sequence, error)
		target error
	}{
		{
			name: "FromClasses rejects modulus above the package limit",
			call: func() (pcseq.Sequence, error) {
				return pcseq.FromClasses(mustClass(t, 0, 65_537))
			},
			target: pcseq.ErrModulusTooLarge,
		},
		{
			name: "FromClassesWithModulus rejects zero modulus",
			call: func() (pcseq.Sequence, error) {
				return pcseq.FromClassesWithModulus(0)
			},
			target: pcseq.ErrInvalidModulus,
		},
		{
			name: "FromClassesWithModulus rejects excessive modulus",
			call: func() (pcseq.Sequence, error) {
				return pcseq.FromClassesWithModulus(65_537)
			},
			target: pcseq.ErrModulusTooLarge,
		},
		{
			name: "FromClassesWithModulus rejects invalid class",
			call: func() (pcseq.Sequence, error) {
				return pcseq.FromClassesWithModulus(12, pitchclass.Class{})
			},
			target: pcseq.ErrInvalidValue,
		},
		{
			name: "FromClassesWithModulus rejects another modulus",
			call: func() (pcseq.Sequence, error) {
				return pcseq.FromClassesWithModulus(12, mustClass(t, 0, 19))
			},
			target: pcseq.ErrModulusMismatch,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.call()
			assert.ErrorIs(t, err, tt.target)
			assert.Equal(t, pcseq.Sequence{}, got)
		})
	}
}

func TestRowCoreCompareContractEdges(t *testing.T) {
	t.Parallel()
	base := mustRow(t, 3, 0, 1, 2)
	for _, tt := range []struct {
		name string
		a    pcseq.Row
		b    pcseq.Row
		want int
	}{
		{name: "same", a: base, b: mustRow(t, 3, 0, 1, 2), want: 0},
		{name: "different modulus", a: mustRow(t, 2, 0, 1), b: base, want: -1},
		{name: "lexicographic", a: base, b: mustRow(t, 3, 0, 2, 1), want: -1},
		{name: "zero before valid", a: pcseq.Row{}, b: base, want: -1},
		{name: "valid after zero", a: base, b: pcseq.Row{}, want: 1},
		{name: "two zeros", a: pcseq.Row{}, b: pcseq.Row{}, want: 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.a.Compare(tt.b))
			assert.Equal(t, -tt.want, tt.b.Compare(tt.a))
		})
	}
}

func TestRowFormCompareContractEdges(t *testing.T) {
	t.Parallel()
	base := mustLabel(t, pcseq.RI, 7, 12)
	for _, tt := range []struct {
		name string
		a    pcseq.FormLabel
		b    pcseq.FormLabel
		want int
	}{
		{name: "same", a: base, b: mustLabel(t, pcseq.RI, 7, 12), want: 0},
		{name: "modulus", a: mustLabel(t, pcseq.RI, 7, 11), b: base, want: -1},
		{name: "kind", a: mustLabel(t, pcseq.R, 7, 12), b: base, want: -1},
		{name: "index", a: mustLabel(t, pcseq.RI, 6, 12), b: base, want: -1},
		{name: "zero before valid", a: pcseq.FormLabel{}, b: base, want: -1},
		{name: "valid after zero", a: base, b: pcseq.FormLabel{}, want: 1},
		{name: "two zeros", a: pcseq.FormLabel{}, b: pcseq.FormLabel{}, want: 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.a.Compare(tt.b))
			assert.Equal(t, -tt.want, tt.b.Compare(tt.a))
		})
	}
}

func TestRowFormEnumCompareContractEdges(t *testing.T) {
	t.Parallel()
	assert.Positive(t, pcseq.P.Compare(pcseq.FormKind(0)))
	assert.Negative(t, pcseq.FormKind(0).Compare(pcseq.P))
	assert.Positive(t, pcseq.I.Compare(pcseq.P))
	assert.Negative(t, pcseq.P.Compare(pcseq.I))
	assert.Zero(t, pcseq.FormKind(99).Compare(pcseq.FormKind(0)))

	assert.Positive(t, pcseq.TIR.Compare(pcseq.RowClassKind(0)))
	assert.Negative(t, pcseq.RowClassKind(0).Compare(pcseq.TIR))
	assert.Positive(t, pcseq.TIRRotations.Compare(pcseq.TIR))
	assert.Negative(t, pcseq.TIR.Compare(pcseq.TIRRotations))
	assert.Zero(t, pcseq.RowClassKind(99).Compare(pcseq.RowClassKind(0)))

	assert.Positive(t, pcseq.DerivationT.Compare(pcseq.DerivationFamily(0)))
	assert.Negative(t, pcseq.DerivationFamily(0).Compare(pcseq.DerivationT))
	assert.Positive(t, pcseq.DerivationI.Compare(pcseq.DerivationT))
	assert.Negative(t, pcseq.DerivationT.Compare(pcseq.DerivationI))
	assert.Zero(t, pcseq.DerivationFamily(99).Compare(pcseq.DerivationFamily(0)))
}

func TestDerivationCompareAndApplyContractEdges(t *testing.T) {
	t.Parallel()
	base, err := pcseq.NewDerivationOperator(pcseq.DerivationT, 1, 4)
	require.NoError(t, err)
	lowerModulus, err := pcseq.NewDerivationOperator(pcseq.DerivationT, 1, 3)
	require.NoError(t, err)
	higherFamily, err := pcseq.NewDerivationOperator(pcseq.DerivationI, 1, 4)
	require.NoError(t, err)
	higherIndex, err := pcseq.NewDerivationOperator(pcseq.DerivationT, 2, 4)
	require.NoError(t, err)
	same, err := pcseq.NewDerivationOperator(pcseq.DerivationT, 1, 4)
	require.NoError(t, err)
	for _, tt := range []struct {
		name string
		a    pcseq.DerivationOperator
		b    pcseq.DerivationOperator
		want int
	}{
		{name: "same", a: base, b: same, want: 0},
		{name: "modulus", a: base, b: lowerModulus, want: 1},
		{name: "family", a: base, b: higherFamily, want: -1},
		{name: "index", a: base, b: higherIndex, want: -1},
		{name: "zero before valid", a: pcseq.DerivationOperator{}, b: base, want: -1},
		{name: "valid after zero", a: base, b: pcseq.DerivationOperator{}, want: 1},
		{name: "two zeros", a: pcseq.DerivationOperator{}, b: pcseq.DerivationOperator{}, want: 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.a.Compare(tt.b))
			assert.Equal(t, -tt.want, tt.b.Compare(tt.a))
		})
	}

	for _, family := range []pcseq.DerivationFamily{
		pcseq.DerivationT,
		pcseq.DerivationI,
		pcseq.DerivationTR,
		pcseq.DerivationIR,
	} {
		t.Run(family.String()+" empty", func(t *testing.T) {
			t.Parallel()
			op, err := pcseq.NewDerivationOperator(family, 0, 12)
			require.NoError(t, err)
			got, err := op.Apply(mustSequence(t, 12))
			require.NoError(t, err)
			assert.True(t, got.IsValid())
			assert.True(t, got.IsEmpty())
		})
	}
}

func TestRowClassModelCompareContract(t *testing.T) {
	t.Parallel()
	row2 := mustRow(t, 2, 0, 1)
	tir2, err := pcseq.NewRowClass(row2, pcseq.TIR)
	require.NoError(t, err)
	rotations2, err := pcseq.NewRowClass(row2, pcseq.TIRRotations)
	require.NoError(t, err)
	tto2, err := pcseq.NewRowClass(row2, pcseq.TTOR)
	require.NoError(t, err)
	tir3, err := pcseq.NewRowClass(mustRow(t, 3, 0, 1, 2), pcseq.TIR)
	require.NoError(t, err)
	chromatic4, err := pcseq.NewRowClass(mustRow(t, 4, 0, 1, 2, 3), pcseq.TIR)
	require.NoError(t, err)
	other4, err := pcseq.NewRowClass(mustRow(t, 4, 0, 2, 1, 3), pcseq.TIR)
	require.NoError(t, err)

	assert.Negative(t, tir2.Compare(rotations2))
	assert.Positive(t, rotations2.Compare(tir2))
	assert.Negative(t, rotations2.Compare(tto2))
	assert.Negative(t, tir2.Compare(tir3))
	assert.Positive(t, tir3.Compare(tir2))
	assert.False(t, chromatic4.Equal(other4))
	assert.NotZero(t, chromatic4.Compare(other4))
	assert.Equal(t, -chromatic4.Compare(other4), other4.Compare(chromatic4))
	assert.Negative(t, pcseq.RowClass{}.Compare(tir2))
	assert.Positive(t, tir2.Compare(pcseq.RowClass{}))
	assert.Zero(t, pcseq.RowClass{}.Compare(pcseq.RowClass{}))
}

func TestRowClassModelContainsNonMembers(t *testing.T) {
	t.Parallel()
	source := mustRow(t, 4, 0, 1, 2, 3)
	other := mustRow(t, 4, 0, 2, 1, 3)
	for _, kind := range []pcseq.RowClassKind{pcseq.TIR, pcseq.TIRRotations, pcseq.TTOR} {
		t.Run(kind.String(), func(t *testing.T) {
			t.Parallel()
			class, err := pcseq.NewRowClass(source, kind)
			require.NoError(t, err)
			assert.True(t, class.Contains(source))
			assert.False(t, class.Contains(other))
			equivalent, err := pcseq.AreEquivalentRows(source, other, kind)
			require.NoError(t, err)
			assert.False(t, equivalent)
		})
	}
}

func TestRowClassModelRotationCanonicalEdges(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		row  pcseq.Row
	}{
		{name: "N4 interleaved", row: mustRow(t, 4, 1, 0, 3, 2)},
		{name: "N5 interleaved", row: mustRow(t, 5, 2, 0, 4, 1, 3)},
		{name: "N6 interleaved", row: mustRow(t, 6, 1, 0, 5, 2, 4, 3)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			class, err := pcseq.NewRowClass(tt.row, pcseq.TIRRotations)
			require.NoError(t, err)
			members, err := class.Members()
			require.NoError(t, err)
			require.NotEmpty(t, members)
			unique := make(map[pcseq.Row]struct{}, len(members))
			for i, member := range members {
				unique[member] = struct{}{}
				assert.True(t, class.Contains(member))
				if i > 0 {
					assert.Negative(t, members[i-1].Compare(member))
				}
			}
			assert.Len(t, unique, class.Size())
			assert.Equal(t, members[0], class.Representative())
			fromMember, err := pcseq.NewRowClass(members[len(members)/2], pcseq.TIRRotations)
			require.NoError(t, err)
			assert.True(t, class.Equal(fromMember))
			assert.Equal(t, class.Representative(), fromMember.Representative())
		})
	}
}

func TestCombinatorialSmallModuliAndErrors(t *testing.T) {
	t.Parallel()
	row := mustRow(t, 2, 0, 1)
	forms, err := row.CombinatorialForms()
	require.NoError(t, err)
	assert.Equal(t, []string{"P1", "I1", "RI0"}, labelStrings(forms))
	rotations, err := row.CombinatorialRotations()
	require.NoError(t, err)
	require.Len(t, rotations, 2)
	assert.Equal(t, 0, rotations[0].Rotation)
	assert.Equal(t, []string{"I1"}, labelStrings(rotations[0].Forms))
	assert.Equal(t, 1, rotations[1].Rotation)
	assert.Equal(t, []string{"I0"}, labelStrings(rotations[1].Forms))

	one := mustRow(t, 1, 0)
	oneForms, err := one.CombinatorialForms()
	require.NoError(t, err)
	assert.NotNil(t, oneForms)
	assert.Empty(t, oneForms)
	oneRotations, err := one.CombinatorialRotations()
	require.NoError(t, err)
	assert.NotNil(t, oneRotations)
	assert.Empty(t, oneRotations)

	_, err = pcseq.Row{}.CombinatorialForms()
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
	_, err = pcseq.Row{}.CombinatorialRotations()
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
	_, err = row.IsCombinatorialPair(mustLabel(t, pcseq.P, 0, 2), pcseq.FormLabel{})
	assert.ErrorIs(t, err, pcseq.ErrInvalidFormLabel)
	_, err = row.IsCombinatorialPair(pcseq.FormLabel{}, pcseq.FormLabel{})
	assert.ErrorIs(t, err, pcseq.ErrInvalidFormLabel)
	_, err = pcseq.Row{}.IsCombinatorialPair(mustLabel(t, pcseq.P, 0, 2), mustLabel(t, pcseq.R, 0, 2))
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)

	odd := mustRow(t, 3, 0, 1, 2)
	_, err = odd.HexachordAnalysis()
	assert.ErrorIs(t, err, pcseq.ErrNotModulus12)
	_, err = pcseq.Row{}.HexachordAnalysis()
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
}

func TestSequenceCoreContentClassEmpty(t *testing.T) {
	t.Parallel()
	empty := mustSequence(t, 12)
	for _, kind := range []setclass.Kind{setclass.Tn, setclass.TnIn, setclass.TTO} {
		t.Run(kind.String(), func(t *testing.T) {
			t.Parallel()
			class, err := empty.ContentClass(kind)
			require.NoError(t, err)
			assert.True(t, class.IsValid())
			assert.True(t, class.IsEmpty())
			assert.Equal(t, 12, class.Modulus())
		})
	}
}

func TestSequenceStringCompactNotationEdges(t *testing.T) {
	t.Parallel()
	_, err := pcseq.Sequence{}.CompactNotation(pitchclass.TE)
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
	compact, err := mustSequence(t, 10).CompactNotation(pitchclass.Decimal)
	require.NoError(t, err)
	assert.Equal(t, "⟨⟩", compact)
	compact, err = mustSequence(t, 12).CompactNotation(pitchclass.TE)
	require.NoError(t, err)
	assert.Equal(t, "⟨⟩", compact)
	_, err = mustSequence(t, 12).CompactNotation(pitchclass.Decimal)
	assert.ErrorIs(t, err, pcseq.ErrNotation)
}

func TestRowStringNotationEdges(t *testing.T) {
	t.Parallel()
	_, err := pcseq.Row{}.Notation(pitchclass.Decimal)
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
	notation, err := mustRow(t, 1, 0).Notation(pitchclass.Decimal)
	require.NoError(t, err)
	assert.Equal(t, "⟨0⟩", notation)
}

func TestParseNotationContractEdges(t *testing.T) {
	t.Parallel()
	single, err := pcseq.ParseNotation("⟨11⟩", pitchclass.Decimal, 12)
	require.NoError(t, err)
	assert.Equal(t, []int{11}, single.Numbers())
	_, err = pcseq.ParseNotation("⟨12⟩", pitchclass.Decimal, 12)
	assert.ErrorIs(t, err, pcseq.ErrNotation)
	_, err = pcseq.ParseNotation("⟨0, 12⟩", pitchclass.Decimal, 12)
	assert.ErrorIs(t, err, pcseq.ErrNotation)
}

func TestRowStringParseNotationErrors(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name    string
		text    string
		style   pitchclass.Style
		modulus int
		target  error
	}{
		{name: "invalid modulus", text: "⟨0⟩", style: pitchclass.Decimal, modulus: 0, target: pcseq.ErrInvalidModulus},
		{name: "large modulus", text: "⟨0⟩", style: pitchclass.Decimal, modulus: 65_537, target: pcseq.ErrModulusTooLarge},
		{name: "invalid style", text: "⟨0⟩", style: pitchclass.Style(99), modulus: 12, target: pcseq.ErrNotation},
		{name: "invalid text", text: "⟨T⟩", style: pitchclass.Decimal, modulus: 12, target: pcseq.ErrNotation},
		{name: "not a row", text: "⟨0, 0⟩", style: pitchclass.Decimal, modulus: 2, target: pcseq.ErrNotRow},
		{name: "wrong separator", text: "⟨0,1⟩", style: pitchclass.Decimal, modulus: 2, target: pcseq.ErrNotation},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			row, err := pcseq.ParseRowNotation(tt.text, tt.style, tt.modulus)
			assert.ErrorIs(t, err, tt.target)
			assert.Equal(t, pcseq.Row{}, row)
		})
	}
}

func TestRowFormParseAdditionalErrors(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"",
		"P",
		"RI",
		"RＩ0",
		"P999999999999999999999999999",
		"P+1",
		"RI00",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			label, err := pcseq.ParseFormLabel(text, 12)
			assert.ErrorIs(t, err, pcseq.ErrNotation)
			assert.Equal(t, pcseq.FormLabel{}, label)
		})
	}
}
