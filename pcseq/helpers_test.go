package pcseq_test

import (
	"fmt"
	"testing"

	"github.com/go-muse/musex/pcseq"
	"github.com/go-muse/musex/pcset"
	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/require"
)

func mustSequence(t *testing.T, modulus int, numbers ...int) pcseq.Sequence {
	t.Helper()
	s, err := pcseq.New(modulus, numbers...)
	require.NoError(t, err)
	return s
}

func mustReduceSequence(t *testing.T, modulus int, numbers ...int) pcseq.Sequence {
	t.Helper()
	s, err := pcseq.Reduce(modulus, numbers...)
	require.NoError(t, err)
	return s
}

func mustRow(t *testing.T, modulus int, numbers ...int) pcseq.Row {
	t.Helper()
	r, err := pcseq.NewRow(modulus, numbers...)
	require.NoError(t, err)
	return r
}

func mustClass(t *testing.T, number, modulus int) pitchclass.Class {
	t.Helper()
	c, err := pitchclass.New(number, modulus)
	require.NoError(t, err)
	return c
}

func mustSet(t *testing.T, modulus int, numbers ...int) pcset.Set {
	t.Helper()
	s, err := pcset.New(modulus, numbers...)
	require.NoError(t, err)
	return s
}

func mustLabel(t *testing.T, kind pcseq.FormKind, index, modulus int) pcseq.FormLabel {
	t.Helper()
	label, err := pcseq.NewFormLabel(kind, index, modulus)
	require.NoError(t, err)
	return label
}

func mustOperator(t *testing.T, multiplier, index, modulus int) pitchclass.Operator {
	t.Helper()
	op, err := pitchclass.NewOperator(multiplier, index, modulus)
	require.NoError(t, err)
	return op
}

func identityNumbers(modulus int) []int {
	numbers := make([]int, modulus)
	for i := range modulus {
		numbers[i] = i
	}
	return numbers
}

func classNumbers(classes []pitchclass.Class) []int {
	if classes == nil {
		return nil
	}
	numbers := make([]int, len(classes))
	for i, class := range classes {
		numbers[i] = class.Number()
	}
	return numbers
}

func labelStrings(labels []pcseq.FormLabel) []string {
	if labels == nil {
		return nil
	}
	result := make([]string, len(labels))
	for i, label := range labels {
		result[i] = label.String()
	}
	return result
}

func derivationStrings(operators []pcseq.DerivationOperator) []string {
	if operators == nil {
		return nil
	}
	result := make([]string, len(operators))
	for i, operator := range operators {
		result[i] = operator.String()
	}
	return result
}

func derivationGroupsStrings(groups [][]pcseq.DerivationOperator) []string {
	if groups == nil {
		return nil
	}
	result := make([]string, len(groups))
	for i, group := range groups {
		result[i] = fmt.Sprint(derivationStrings(group))
	}
	return result
}
