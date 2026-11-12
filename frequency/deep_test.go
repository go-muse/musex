package frequency_test

import (
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"
	"testing"

	"github.com/go-muse/musex/frequency"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCriterion_DeepNesting(t *testing.T) {
	old := debug.SetMaxStack(1 << 20)
	defer debug.SetMaxStack(old)

	const depth = 100_000
	build := func() frequency.Criterion {
		criterion := frequency.Strict()
		for range depth {
			criterion = frequency.MustAnyOf(criterion)
		}
		return criterion
	}
	criterion := build()
	other := build()
	text := strings.Repeat("any(", depth) + "strict" + strings.Repeat(")", depth)

	parsed, err := frequency.Parse(text)
	require.NoError(t, err)
	require.True(t, criterion.Equal(parsed))

	var unmarshaled frequency.Criterion
	require.NoError(t, unmarshaled.UnmarshalText([]byte(text)))
	require.True(t, criterion.Equal(unmarshaled))
	require.True(t, criterion.Equal(other))

	assert.True(t, criterion.String() == text)
	marshaled, err := criterion.MarshalText()
	require.NoError(t, err)
	assert.True(t, string(marshaled) == text)
	appended, err := criterion.AppendText(nil)
	require.NoError(t, err)
	assert.True(t, string(appended) == text)

	assert.True(t, fmt.Sprint(criterion) == text)
	assert.True(t, fmt.Sprintf("%q", criterion) == strconv.Quote(text))
	assert.True(t, fmt.Sprintf("%d", criterion) == "%!d(frequency.Criterion="+text+")")
	assert.True(t, fmt.Sprintf("%*v", len(text)+2, criterion) == "  "+text)
	assert.True(t, fmt.Sprintf("%-*v|", len(text)+2, criterion) == text+"  |")

	near, err := criterion.Near(440, 440)
	require.NoError(t, err)
	assert.True(t, near)
	near, err = criterion.Near(440, 441)
	require.NoError(t, err)
	assert.False(t, near)
	comparison, err := criterion.Compare(440, 441)
	require.NoError(t, err)
	assert.Equal(t, -1, comparison)
	assert.Len(t, criterion.Components(), 1)

	invalidTolerance := strings.Repeat("any(", depth) + "hertz(-1)" + strings.Repeat(")", depth)
	parsed, err = frequency.Parse(invalidTolerance)
	assert.False(t, parsed.IsValid())
	requireErrorChain(t, err, frequency.ErrInvalidTolerance, nil)

	invalidText := text[:len(text)-1]
	parsed, err = frequency.Parse(invalidText)
	assert.False(t, parsed.IsValid())
	requireErrorChain(t, err, frequency.ErrInvalidText, nil)
}
