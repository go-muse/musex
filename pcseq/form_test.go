package pcseq_test

import (
	"reflect"
	"testing"

	"github.com/go-muse/musex/pcseq"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormLabelModelEnumsAndValues(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		kind pcseq.FormKind
		text string
	}{
		{kind: pcseq.P, text: "P"},
		{kind: pcseq.I, text: "I"},
		{kind: pcseq.R, text: "R"},
		{kind: pcseq.RI, text: "RI"},
	} {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()
			assert.True(t, tt.kind.IsValid())
			assert.True(t, tt.kind.Equal(tt.kind))
			assert.Zero(t, tt.kind.Compare(tt.kind))
			assert.Equal(t, tt.text, tt.kind.String())
		})
	}
	var invalid pcseq.FormKind
	assert.False(t, invalid.IsValid())
	assert.False(t, invalid.Equal(pcseq.FormKind(0)))
	assert.Zero(t, invalid.Compare(pcseq.FormKind(99)))
	assert.Negative(t, invalid.Compare(pcseq.P))
	assert.Equal(t, "FormKind(invalid)", invalid.String())
	assert.Negative(t, pcseq.P.Compare(pcseq.I))
	assert.Negative(t, pcseq.R.Compare(pcseq.RI))

	label := mustLabel(t, pcseq.RI, 7, 12)
	assert.True(t, label.IsValid())
	assert.Equal(t, pcseq.RI, label.Kind())
	assert.Equal(t, 7, label.Index())
	assert.Equal(t, 12, label.Modulus())
	assert.Equal(t, "RI7", label.String())
	same := mustLabel(t, pcseq.RI, 7, 12)
	assert.True(t, label.Equal(same))
	assert.Equal(t, label, same)
	assert.Zero(t, label.Compare(same))
	assert.Negative(t, mustLabel(t, pcseq.I, 11, 12).Compare(label))
	assert.Negative(t, mustLabel(t, pcseq.RI, 6, 12).Compare(label))
	assert.Negative(t, mustLabel(t, pcseq.RI, 7, 11).Compare(label))
	keys := map[pcseq.FormLabel]int{label: 1}
	keys[same] = 2
	assert.Len(t, keys, 1)
	assert.True(t, reflect.TypeFor[pcseq.FormLabel]().Comparable())

	var zero pcseq.FormLabel
	assert.False(t, zero.IsValid())
	assert.Zero(t, zero.Index())
	assert.Zero(t, zero.Modulus())
	assert.Equal(t, "FormLabel(invalid)", zero.String())
	assert.False(t, zero.Equal(pcseq.FormLabel{}))
	assert.Zero(t, zero.Compare(pcseq.FormLabel{}))
	assert.Negative(t, zero.Compare(label))

	for _, tt := range []struct {
		name     string
		kind     pcseq.FormKind
		index    int
		modulus  int
		expected error
	}{
		{name: "invalid modulus", kind: pcseq.P, index: 0, modulus: 0, expected: pcseq.ErrInvalidModulus},
		{name: "large modulus", kind: pcseq.P, index: 0, modulus: 65_537, expected: pcseq.ErrModulusTooLarge},
		{name: "invalid kind", kind: pcseq.FormKind(0), index: 0, modulus: 12, expected: pcseq.ErrInvalidFormLabel},
		{name: "negative index", kind: pcseq.P, index: -1, modulus: 12, expected: pcseq.ErrInvalidFormLabel},
		{name: "index at modulus", kind: pcseq.P, index: 12, modulus: 12, expected: pcseq.ErrInvalidFormLabel},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := pcseq.NewFormLabel(tt.kind, tt.index, tt.modulus)
			assert.ErrorIs(t, err, tt.expected)
			assert.Equal(t, pcseq.FormLabel{}, got)
		})
	}
}

func TestRowFormPrimeInversionRetrograde(t *testing.T) {
	t.Parallel()
	row := mustRow(t, 12, 3, 7, 11, 1, 5, 0, 2, 10, 6, 4, 8, 9)
	for _, tt := range []struct {
		name     string
		kind     pcseq.FormKind
		expected []int
	}{
		{name: "P0", kind: pcseq.P, expected: []int{0, 4, 8, 10, 2, 9, 11, 7, 3, 1, 5, 6}},
		{name: "I0", kind: pcseq.I, expected: []int{0, 8, 4, 2, 10, 3, 1, 5, 9, 11, 7, 6}},
		{name: "R0", kind: pcseq.R, expected: []int{6, 5, 1, 3, 7, 11, 9, 2, 10, 8, 4, 0}},
		{name: "RI0", kind: pcseq.RI, expected: []int{6, 7, 11, 9, 5, 1, 3, 10, 2, 4, 8, 0}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			form, err := row.Form(mustLabel(t, tt.kind, 0, 12))
			require.NoError(t, err)
			assert.Equal(t, tt.expected, form.Numbers())
			labels, err := row.LabelsOf(form)
			require.NoError(t, err)
			assert.Equal(t, []string{tt.name}, labelStrings(labels))
		})
	}

	primeThree, err := row.Form(mustLabel(t, pcseq.P, 3, 12))
	require.NoError(t, err)
	assert.Equal(t, row, primeThree, "P3 is the source because its first class is 3")
	labels, err := row.LabelsOf(primeThree)
	require.NoError(t, err)
	assert.Contains(t, labelStrings(labels), "P3")

	for _, kind := range []pcseq.FormKind{pcseq.P, pcseq.I, pcseq.R, pcseq.RI} {
		form, err := row.Form(mustLabel(t, kind, 5, 12))
		require.NoError(t, err)
		labels, err := row.LabelsOf(form)
		require.NoError(t, err)
		require.Len(t, labels, 1)
		roundTrip, err := row.Form(labels[0])
		require.NoError(t, err)
		assert.Equal(t, form, roundTrip)
	}

	notForm := mustRow(t, 12, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 11, 10)
	labels, err = row.LabelsOf(notForm)
	require.NoError(t, err)
	assert.NotNil(t, labels)
	assert.Empty(t, labels)

	_, err = row.Form(mustLabel(t, pcseq.P, 0, 19))
	assert.ErrorIs(t, err, pcseq.ErrModulusMismatch)
	_, err = row.Form(pcseq.FormLabel{})
	assert.ErrorIs(t, err, pcseq.ErrInvalidFormLabel)
	_, err = pcseq.Row{}.Form(mustLabel(t, pcseq.P, 0, 12))
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
	_, err = row.LabelsOf(mustRow(t, 19, identityNumbers(19)...))
	assert.ErrorIs(t, err, pcseq.ErrModulusMismatch)
	_, err = row.LabelsOf(pcseq.Row{})
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)

	_, baseErr := row.Form(pcseq.FormLabel{})
	assert.PanicsWithValue(t, baseErr, func() { row.MustForm(pcseq.FormLabel{}) })
	_, baseErr = row.LabelsOf(pcseq.Row{})
	assert.PanicsWithValue(t, baseErr, func() { row.MustLabelsOf(pcseq.Row{}) })
}

func TestRowFormSymmetryAndRecognition(t *testing.T) {
	t.Parallel()
	op21 := mustRow(t, 12, 0, 3, 2, 1, 5, 4, 10, 11, 7, 8, 9, 6)
	retro, err := op21.Retrograde()
	require.NoError(t, err)
	primeSix, err := op21.Transpose(6)
	require.NoError(t, err)
	assert.Equal(t, primeSix, retro)
	r0, err := op21.Form(mustLabel(t, pcseq.R, 0, 12))
	require.NoError(t, err)
	p6, err := op21.Form(mustLabel(t, pcseq.P, 6, 12))
	require.NoError(t, err)
	assert.Equal(t, p6, r0)
	labels, err := op21.LabelsOf(r0)
	require.NoError(t, err)
	assert.Equal(t, []string{"P6", "R0"}, labelStrings(labels))

	omtP0 := mustRow(t, 12, 9, 6, 7, 8, 4, 5, 11, 10, 2, 1, 0, 3)
	labels, err = op21.LabelsOf(omtP0)
	require.NoError(t, err)
	assert.Equal(t, []string{"I9", "RI3"}, labelStrings(labels))
}
