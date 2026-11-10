package system_test

import (
	"testing"

	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for the built-in systems: OP-4, OP-41, PK-3 (edition 2), ODV-8, spec 2.6,
// contract 5.1 and 5.6, criteria K-1, K-29.

// builtinRow is one row of table 2.6 of the spec (contract 5.6).
type builtinRow struct {
	get     func() *system.System
	name    string
	n       int
	offsets [letter.Count]int
	c       int
	// den is d of A = (1/d)Z; 1 means A = Z.
	den          int
	regular      bool
	large, small int
}

// builtinTable returns the 20 built-in systems in the order of table 2.6.
func builtinTable() []builtinRow {
	return []builtinRow{
		{system.Traditional, "traditional", 12, offs(0, 2, 4, 5, 7, 9, 11), 1, 1, true, 2, 1},
		{system.HeptatonicSecondGenerated8, "8-heptatonic-2nd-generated", 8, offs(0, 1, 2, 3, 4, 6, 7), 1, 1, false, 0, 0},
		{system.Blackwood15, "15-blackwood", 15, offs(0, 2, 5, 6, 8, 11, 14), 1, 1, false, 0, 0},
		{system.UpsAndDowns15, "15-ups-and-downs", 15, offs(0, 3, 6, 6, 9, 12, 15), 3, 3, true, 3, 0},
		{system.Porcupine15, "15-porcupine", 15, offs(0, 2, 4, 6, 8, 11, 13), 1, 1, false, 0, 0},
		{system.Blackwood17, "17-blackwood", 17, offs(0, 3, 6, 7, 10, 13, 16), 2, 1, true, 3, 1},
		{system.QuarterTones17, "17-quarter-tone", 17, offs(0, 3, 6, 7, 10, 13, 16), 2, 2, true, 3, 1},
		{system.Positions19, "19-position", 19, offs(0, 3, 6, 8, 11, 14, 17), 1, 1, true, 3, 2},
		{system.UpsAndDowns22, "22-ups-and-downs", 22, offs(0, 4, 8, 9, 13, 17, 21), 3, 3, true, 4, 1},
		{system.QuarterTones22, "22-quarter-tone", 22, offs(0, 4, 8, 9, 13, 17, 21), 2, 2, false, 0, 0},
		{system.Porcupine22, "22-porcupine", 22, offs(0, 3, 6, 9, 12, 16, 19), 1, 1, false, 0, 0},
		{system.QuarterTones24, "24-quarter-tone", 24, offs(0, 4, 8, 10, 14, 18, 22), 2, 2, true, 4, 2},
		{system.SharpsAndFlats31, "31-sharps-and-flats", 31, offs(0, 5, 10, 13, 18, 23, 28), 2, 1, true, 5, 3},
		{system.Fokker31, "31-fokker", 31, offs(0, 5, 10, 13, 18, 23, 28), 2, 2, true, 5, 3},
		{system.ExtendedPythagorean41, "41-extended-pythagorean", 41, offs(0, 7, 14, 17, 24, 31, 38), 4, 1, true, 7, 3},
		{system.UpsAndDowns41, "41-ups-and-downs", 41, offs(0, 7, 14, 17, 24, 31, 38), 4, 4, true, 7, 3},
		{system.ExtendedPythagorean53, "53-extended-pythagorean", 53, offs(0, 9, 18, 22, 31, 40, 49), 5, 1, true, 9, 4},
		{system.UpsAndDowns53, "53-ups-and-downs", 53, offs(0, 9, 18, 22, 31, 40, 49), 5, 5, true, 9, 4},
		{system.ArelEzgiUzdilek53, "53-arel-ezgi-uzdilek", 53, offs(0, 9, 18, 22, 31, 40, 49), 4, 4, false, 0, 0},
		{system.ManeriSims72, "72-maneri-sims", 72, offs(0, 12, 24, 30, 42, 54, 66), 6, 6, true, 12, 6},
	}
}

func TestBuiltins_Attributes(t *testing.T) {
	t.Parallel()

	for i, row := range builtinTable() {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()

			s := row.get()
			require.NotNil(t, s, "row %d", i+1)
			assert.True(t, s.IsValid())
			assert.Same(t, s, row.get(), "each call returns the same system")
			assert.Equal(t, row.name, s.Name())
			assert.Equal(t, row.n, s.Classes())
			assert.Equal(t, letter.C, s.Boundary())
			assert.Equal(t, row.offsets, s.Definition().Offsets)
			assert.Equal(t, row.c, s.Chromatic())

			want := system.MustMultiples(row.den)
			assert.True(t, want.Equal(s.Alterations()), "rule %s, want %s", s.Alterations(), want)

			// RegularByClasses is not a column of table 2.6: it follows from the steps by
			// musical-system.inv-8. A regular system is regular by classes; each irregular
			// built-in system (No. 2, 3, 5, 10, 11, 19) has steps after C, D, F, G, A that
			// differ mod N, or steps after E and B that are not c less, so it is not.
			wantReg := system.Regularity{Regular: row.regular, RegularByClasses: row.regular}
			if row.regular {
				wantReg.Large, wantReg.Small = row.large, row.small
			}
			assert.Equal(t, wantReg, s.MustRegularity())

			assert.True(t, s.Equal(system.MustNew(s.Definition())), "the system equals a system from its definition")
		})
	}
}

func TestBuiltins_BoundaryLetterOnN(t *testing.T) {
	t.Parallel()

	// No. 4: B stands on N = 15 (P-2); its natural name carries 1.
	s := system.UpsAndDowns15()
	assert.Equal(t, 15, s.MustOffset(letter.B))
	assert.Equal(t, 1, s.MustCarry(spi(letter.B, 0)))
	assert.Equal(t, cls(0, 15), s.MustNaturalClass(letter.B))
}

func TestBuiltins_PairwiseNotEqual(t *testing.T) {
	t.Parallel()

	// Every two rows of table 2.6 differ in N, offsets, c or A; in particular the rows
	// with equal offsets (No. 6-7, 9-10, 13-14, 15-16, 17-19) are not equal (contract 5.6).
	rows := builtinTable()
	for i, x := range rows {
		assert.True(t, x.get().Equal(x.get()), "%s equals itself", x.name)
		for j, y := range rows {
			if i == j {
				continue
			}
			assert.False(t, x.get().Equal(y.get()), "%s must not equal %s", x.name, y.name)
		}
	}
}

func TestBuiltins_Sequence(t *testing.T) {
	t.Parallel()

	rows := builtinTable()
	want := make([]*system.System, 0, len(rows))
	for _, row := range rows {
		want = append(want, row.get())
	}
	require.Len(t, want, 20)

	seq := system.Builtins()
	got := collect(t, seq)
	require.Len(t, got, len(want))
	for i := range want {
		assert.Same(t, want[i], got[i], "built-in %d (%s)", i+1, rows[i].name)
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

	for _, row := range builtinTable() {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()

			got, ok := system.LookupBuiltin(row.name)
			assert.True(t, ok)
			assert.Same(t, row.get(), got)
		})
	}

	for _, input := range []string{
		"Traditional", "traditional ", " traditional", "", "31-position", "53-Arel-Ezgi-Uzdilek", "24-quarter-tones",
	} {
		t.Run("not found: "+input, func(t *testing.T) {
			t.Parallel()

			got, ok := system.LookupBuiltin(input)
			assert.False(t, ok)
			assert.Nil(t, got)
		})
	}
}
