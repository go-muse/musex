package system_test

import (
	"testing"

	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for the built-in systems: OP-4, OP-41, spec 2.6, contract 5.6, criteria K-1, K-29.

func TestBuiltins_Attributes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		get        func() *system.System
		sysName    string
		n          int
		offsets    [7]int
		c          int
		rule       *system.Rule
		regularity system.Regularity
	}{
		{
			name: "Traditional", get: system.Traditional, sysName: "traditional",
			n: 12, offsets: offs(0, 2, 4, 5, 7, 9, 11), c: 1, rule: system.Integers(),
			regularity: system.Regularity{Regular: true, Large: 2, Small: 1, RegularByClasses: true},
		},
		{
			name: "Positions19", get: system.Positions19, sysName: "19-position",
			n: 19, offsets: offs(0, 3, 6, 8, 11, 14, 17), c: 1, rule: system.Integers(),
			regularity: system.Regularity{Regular: true, Large: 3, Small: 2, RegularByClasses: true},
		},
		{
			name: "QuarterTones24", get: system.QuarterTones24, sysName: "24-quarter-tone",
			n: 24, offsets: offs(0, 4, 8, 10, 14, 18, 22), c: 2, rule: system.MustMultiples(2),
			regularity: system.Regularity{Regular: true, Large: 4, Small: 2, RegularByClasses: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := tt.get()
			require.NotNil(t, s)
			assert.True(t, s.IsValid())
			assert.Same(t, s, tt.get(), "each call returns the same system")
			assert.Equal(t, tt.sysName, s.Name())
			assert.Equal(t, tt.n, s.Classes())
			assert.Equal(t, letter.C, s.Boundary())
			assert.Equal(t, tt.offsets, s.Definition().Offsets)
			assert.Equal(t, tt.c, s.Chromatic())
			assert.True(t, tt.rule.Equal(s.Alterations()), "rule %s, want %s", s.Alterations(), tt.rule)
			assert.Equal(t, tt.regularity, s.MustRegularity())

			def := s.Definition()
			assert.True(t, s.Equal(system.MustNew(def)), "the system equals a system from its definition")
		})
	}
}

func TestBuiltins_Sequence(t *testing.T) {
	t.Parallel()

	seq := system.Builtins()
	want := []*system.System{system.Traditional(), system.Positions19(), system.QuarterTones24()}

	got := collect(t, seq)
	require.Len(t, got, len(want))
	for i := range want {
		assert.Same(t, want[i], got[i], "built-in %d", i)
	}

	// Repeated traversal and early stop.
	again := collect(t, seq)
	require.Len(t, again, len(want))
	for i := range want {
		assert.Same(t, want[i], again[i])
	}
	stopped := first(t, seq, 1)
	require.Len(t, stopped, 1)
	assert.Same(t, want[0], stopped[0])
}

func TestLookupBuiltin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  *system.System
	}{
		{name: "traditional", input: "traditional", want: system.Traditional()},
		{name: "19-position", input: "19-position", want: system.Positions19()},
		{name: "24-quarter-tone", input: "24-quarter-tone", want: system.QuarterTones24()},
		{name: "case matters", input: "Traditional"},
		{name: "trailing space", input: "traditional "},
		{name: "leading space", input: " traditional"},
		{name: "empty", input: ""},
		{name: "unknown", input: "31-position"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, ok := system.LookupBuiltin(tt.input)
			if tt.want == nil {
				assert.False(t, ok)
				assert.Nil(t, got)
				return
			}
			assert.True(t, ok)
			assert.Same(t, tt.want, got)
		})
	}
}
