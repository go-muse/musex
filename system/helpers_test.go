package system_test

import (
	"iter"
	"slices"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// maxAlt is M, the largest magnitude of a guaranteed alteration value (2^24).
const maxAlt = alteration.MaxMagnitude

// ---- Value helpers.

// a returns the whole alteration value n.
func a(n int) alteration.Alteration { return alteration.MustFromInt(n) }

// q returns the alteration value num/den.
func q(num, den int) alteration.Alteration { return alteration.MustFromFraction(num, den) }

// spi returns the name with letter l and whole alteration value n.
func spi(l letter.Letter, n int) spelling.Spelling { return spelling.MustNew(l, a(n)) }

// spq returns the name with letter l and alteration value num/den.
func spq(l letter.Letter, num, den int) spelling.Spelling { return spelling.MustNew(l, q(num, den)) }

// dHalfSharp returns D+1/2, which is admissible in the 24-class system only.
func dHalfSharp() spelling.Spelling { return spq(letter.D, 1, 2) }

// o returns the octave with ASPN number n.
func o(n int) octave.Octave { return octave.FromASPN(n) }

// pt returns the pitch with name s and positional Octave n (not the written register).
func pt(s spelling.Spelling, n int) pitch.Pitch { return pitch.MustNew(s, o(n)) }

// cls returns the class n mod modulus.
func cls(n, modulus int) pitchclass.Class { return pitchclass.MustNew(n, modulus) }

// offs builds an offsets array indexed by letter.Letter.Ordinal (C, D, E, F, G, A, B).
func offs(c, d, e, f, g, aa, b int) [letter.Count]int {
	return [letter.Count]int{c, d, e, f, g, aa, b}
}

// collect drains a sequence returned by the package.
func collect[T any](t testing.TB, seq iter.Seq[T]) []T {
	t.Helper()
	require.NotNil(t, seq, "sequence must not be nil")
	return slices.Collect(seq)
}

// first returns at most n first elements of seq and stops the iteration early.
func first[T any](t testing.TB, seq iter.Seq[T], n int) []T {
	t.Helper()
	require.NotNil(t, seq, "sequence must not be nil")
	out := make([]T, 0, n)
	if n == 0 {
		return out
	}
	for v := range seq {
		out = append(out, v)
		if len(out) == n {
			break
		}
	}
	return out
}

// ---- Systems.

// newSystem creates a system and requires success.
func newSystem(t testing.TB, def system.Definition) *system.System {
	t.Helper()
	s, err := system.New(def)
	require.NoError(t, err)
	require.NotNil(t, s)
	return s
}

// tradOffsets are the letter offsets of the traditional system.
func tradOffsets() [letter.Count]int { return offs(0, 2, 4, 5, 7, 9, 11) }

// tradLayoutDef is a definition with the traditional layout and the given rule A.
func tradLayoutDef(rule *system.Rule) system.Definition {
	return system.Definition{
		Name: "trad-layout", Classes: 12, Boundary: letter.C,
		Offsets: tradOffsets(), Chromatic: 1, Alterations: rule,
	}
}

// defSA is SA of the spec: traditional steps with the boundary letter A.
func defSA() system.Definition {
	return system.Definition{
		Name: "SA", Classes: 12, Boundary: letter.A,
		Offsets: offs(3, 5, 7, 8, 10, 0, 2), Chromatic: 1, Alterations: system.Integers(),
	}
}

// defS14 is S14 of the spec: equal steps 2 from C, N = 14.
func defS14() system.Definition {
	return system.Definition{
		Name: "S14", Classes: 14, Boundary: letter.C,
		Offsets: offs(0, 2, 4, 6, 8, 10, 12), Chromatic: 1, Alterations: system.Integers(),
	}
}

// defS12Irregular is the irregular 12-class system C0 D3 E4 F5 G7 A9 B11.
func defS12Irregular() system.Definition {
	return system.Definition{
		Name: "S12-irregular", Classes: 12, Boundary: letter.C,
		Offsets: offs(0, 3, 4, 5, 7, 9, 11), Chromatic: 1, Alterations: system.Integers(),
	}
}

// defS2 is S2 of the spec: N = 2, C0, the other letters on 1.
func defS2() system.Definition {
	return system.Definition{
		Name: "S2", Classes: 2, Boundary: letter.C,
		Offsets: offs(0, 1, 1, 1, 1, 1, 1), Chromatic: 1, Alterations: system.Integers(),
	}
}

// requireDefinition checks a definition field by field; the rules are compared
// with Rule.Equal because they may be different pointers.
func requireDefinition(t testing.TB, want, got system.Definition) {
	t.Helper()
	require.NotNil(t, got.Alterations)
	assert.True(t, want.Alterations.Equal(got.Alterations), "rule %s, want %s", got.Alterations, want.Alterations)
	want.Alterations, got.Alterations = nil, nil
	assert.Equal(t, want, got)
}

// ---- Reference model: the formulas of the spec (P-5…P-10), computed by the tests.

// model is an independent description of a system for computing expected values.
type model struct {
	name    string
	build   func(t testing.TB) *system.System
	n       int
	offsets [letter.Count]int
	c       int
}

func modelT() model {
	return model{
		name: "T", n: 12, offsets: tradOffsets(), c: 1,
		build: func(testing.TB) *system.System { return system.Traditional() },
	}
}

func modelS19() model {
	return model{
		name: "S19", n: 19, offsets: offs(0, 3, 6, 8, 11, 14, 17), c: 1,
		build: func(testing.TB) *system.System { return system.Positions19() },
	}
}

func modelS24() model {
	return model{
		name: "S24", n: 24, offsets: offs(0, 4, 8, 10, 14, 18, 22), c: 2,
		build: func(testing.TB) *system.System { return system.QuarterTones24() },
	}
}

func modelSA() model {
	d := defSA()
	return model{
		name: "SA", n: d.Classes, offsets: d.Offsets, c: d.Chromatic,
		build: func(t testing.TB) *system.System { return newSystem(t, defSA()) },
	}
}

func modelS14() model {
	d := defS14()
	return model{
		name: "S14", n: d.Classes, offsets: d.Offsets, c: d.Chromatic,
		build: func(t testing.TB) *system.System { return newSystem(t, defS14()) },
	}
}

func modelS12Irregular() model {
	d := defS12Irregular()
	return model{
		name: "S12-irregular", n: d.Classes, offsets: d.Offsets, c: d.Chromatic,
		build: func(t testing.TB) *system.System { return newSystem(t, defS12Irregular()) },
	}
}

func modelS2() model {
	d := defS2()
	return model{
		name: "S2", n: d.Classes, offsets: d.Offsets, c: d.Chromatic,
		build: func(t testing.TB) *system.System { return newSystem(t, defS2()) },
	}
}

// floorDiv returns ⌊x/n⌋ for n > 0.
func floorDiv(x, n int) int {
	d := x / n
	if x%n != 0 && x < 0 {
		d--
	}
	return d
}

// shift returns s = p(l) + a·c for a name whose alteration denominator divides c.
func (m model) shift(t testing.TB, s spelling.Spelling) int {
	t.Helper()
	num, den := s.Alteration().Fraction()
	require.Zero(t, m.c%den, "model %s: a·c must be whole for %s", m.name, s)
	return m.offsets[s.Letter().Ordinal()] + num*(m.c/den)
}

// carry returns k = ⌊(p(l) + a·c)/N⌋ (P-7).
func (m model) carry(t testing.TB, s spelling.Spelling) int {
	t.Helper()
	return floorDiv(m.shift(t, s), m.n)
}

// class returns κ = (p(l) + a·c) mod N (P-5).
func (m model) class(t testing.TB, s spelling.Spelling) int {
	t.Helper()
	sh := m.shift(t, s)
	return sh - floorDiv(sh, m.n)*m.n
}

// position returns P = Octave·N + κ (P-8); only for octaves where it fits in int.
func (m model) position(t testing.TB, p pitch.Pitch) int {
	t.Helper()
	return p.Octave().MustASPN()*m.n + m.class(t, p.Spelling())
}

// names returns all names with whole alteration values from -limit to limit,
// letters C…B, then by alteration.
func namesUpTo(limit int) []spelling.Spelling {
	out := make([]spelling.Spelling, 0, letter.Count*(2*limit+1))
	for _, l := range letter.All() {
		for n := -limit; n <= limit; n++ {
			out = append(out, spi(l, n))
		}
	}
	return out
}

// ---- Errors.

type namedErr struct {
	name string
	err  error
}

// categories are the nine error categories E-1…E-9 of the contract.
func categories() []namedErr {
	return []namedErr{
		{"ErrInvalid", system.ErrInvalid},
		{"ErrDefinition", system.ErrDefinition},
		{"ErrModulusMismatch", system.ErrModulusMismatch},
		{"ErrLimit", system.ErrLimit},
		{"ErrNotApplicable", system.ErrNotApplicable},
		{"ErrInadmissible", system.ErrInadmissible},
		{"ErrNoSpelling", system.ErrNoSpelling},
		{"ErrBadSpelling", system.ErrBadSpelling},
		{"ErrOutOfRange", system.ErrOutOfRange},
	}
}

// causes are the seven causes of E-2.
func causes() []namedErr {
	return []namedErr{
		{"ErrClassCount", system.ErrClassCount},
		{"ErrOffsets", system.ErrOffsets},
		{"ErrDescending", system.ErrDescending},
		{"ErrChromatic", system.ErrChromatic},
		{"ErrNoNatural", system.ErrNoNatural},
		{"ErrFractionalShift", system.ErrFractionalShift},
		{"ErrRule", system.ErrRule},
	}
}

// requireErr checks that err matches exactly one category, the wanted one, and,
// for E-2 (ErrDefinition), exactly one cause, the wanted one. For other categories
// cause must be nil and err must match no cause.
func requireErr(t testing.TB, err, category, cause error) {
	t.Helper()
	require.Error(t, err)
	for _, c := range categories() {
		if c.err == category {
			assert.ErrorIs(t, err, c.err, "error %q must match %s", err, c.name)
		} else {
			assert.NotErrorIs(t, err, c.err, "error %q must not match %s", err, c.name)
		}
	}
	for _, c := range causes() {
		if cause != nil && c.err == cause {
			assert.ErrorIs(t, err, c.err, "error %q must match cause %s", err, c.name)
		} else {
			assert.NotErrorIs(t, err, c.err, "error %q must not match cause %s", err, c.name)
		}
	}
}

// requirePanicErr runs f, requires a panic whose value is an error, and checks that
// error as requireErr does.
func requirePanicErr(t testing.TB, category, cause error, f func()) {
	t.Helper()
	var recovered any
	panicked := true
	func() {
		defer func() { recovered = recover() }()
		f()
		panicked = false
	}()
	require.True(t, panicked, "expected a panic")
	err, ok := recovered.(error)
	require.True(t, ok, "panic value must be an error, got %T (%v)", recovered, recovered)
	requireErr(t, err, category, cause)
}
