package modal_test

import (
	"strings"
	"testing"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/mode"

	"github.com/stretchr/testify/require"
)

// These fixtures describe specific intervals, without a sounding system.
func degreeIntervals(t *testing.T, text string) []interval.Interval {
	t.Helper()
	var result []interval.Interval
	for _, token := range strings.Fields(text) {
		value, err := interval.Parse(token)
		require.NoError(t, err, "fixture interval %s", token)
		result = append(result, value)
	}
	return result
}

func closedMode(t *testing.T, degrees, cycle string) mode.Mode {
	t.Helper()
	values := degreeIntervals(t, degrees)
	cycles := degreeIntervals(t, cycle)
	require.Len(t, cycles, 1)
	m, err := mode.ClosedFromDegreeIntervals(values, cycles[0])
	require.NoError(t, err, "closed fixture %s / %s", degrees, cycle)
	return m
}

func openMode(t *testing.T, degrees string) mode.Mode {
	t.Helper()
	m, err := mode.OpenFromDegreeIntervals(degreeIntervals(t, degrees))
	require.NoError(t, err, "open fixture %s", degrees)
	return m
}

func dorianMode(t *testing.T) mode.Mode {
	t.Helper()
	return closedMode(t, "P1 M2 m3 P4 P5 M6 m7", "P8")
}
