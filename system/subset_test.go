package system_test

import (
	"fmt"
	"testing"

	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for Rule.SubsetOf and Rule.Equal: OP-13, P-11, P-12, contract 5.5 and 6.3,
// criterion K-14 (rules), risk R-3 (enumeration oracle).

func TestRule_SubsetOfAndEqual(t *testing.T) {
	t.Parallel()

	m := maxAlt
	tests := []struct {
		name   string
		x, y   func(t testing.TB) *system.Rule
		subset bool // x ⊆ y
		super  bool // y ⊆ x
	}{
		{
			name:   "integers and halves",
			x:      func(testing.TB) *system.Rule { return system.Integers() },
			y:      func(testing.TB) *system.Rule { return system.MustMultiples(2) },
			subset: true,
		},
		{
			name: "halves and integers with half plus integers (K-14)",
			x:    func(testing.TB) *system.Rule { return system.MustMultiples(2) },
			y: func(t testing.TB) *system.Rule {
				return rule(t, prog(t, a(0), a(1)), prog(t, q(1, 2), a(1)))
			},
			subset: true, super: true,
		},
		{
			name:   "integers and integers from -M to M (K-14)",
			x:      func(testing.TB) *system.Rule { return system.Integers() },
			y:      func(testing.TB) *system.Rule { return system.MustIntegerRange(-m, m) },
			subset: true, super: true,
		},
		{
			name:  "integers and integers from -M+1 to M",
			x:     func(testing.TB) *system.Rule { return system.Integers() },
			y:     func(testing.TB) *system.Rule { return system.MustIntegerRange(-m+1, m) },
			super: true,
		},
		{
			name:  "integers and integers from -M to M-1",
			x:     func(testing.TB) *system.Rule { return system.Integers() },
			y:     func(testing.TB) *system.Rule { return system.MustIntegerRange(-m, m-1) },
			super: true,
		},
		{
			name: "half plus integers and integers",
			x:    func(t testing.TB) *system.Rule { return rule(t, prog(t, q(1, 2), a(1))) },
			y:    func(testing.TB) *system.Rule { return system.Integers() },
		},
		{
			name:   "integers from -2 to 2 and integers",
			x:      func(testing.TB) *system.Rule { return system.MustIntegerRange(-2, 2) },
			y:      func(testing.TB) *system.Rule { return system.Integers() },
			subset: true,
		},
		{
			name:   "single zero and anything with zero",
			x:      func(testing.TB) *system.Rule { return system.Single(a(0)) },
			y:      func(testing.TB) *system.Rule { return system.MustIntegerRange(-1, 1) },
			subset: true,
		},
		{
			name:   "thirds and sixths",
			x:      func(testing.TB) *system.Rule { return system.MustMultiples(3) },
			y:      func(testing.TB) *system.Rule { return system.MustMultiples(6) },
			subset: true,
		},
		{
			name: "sixths and thirds with sixth plus thirds",
			x:    func(testing.TB) *system.Rule { return system.MustMultiples(6) },
			y: func(t testing.TB) *system.Rule {
				return rule(t, prog(t, a(0), q(1, 3)), prog(t, q(1, 6), q(1, 3)))
			},
			subset: true, super: true,
		},
		{
			name: "sixths and thirds with halves",
			x:    func(testing.TB) *system.Rule { return system.MustMultiples(6) },
			y: func(t testing.TB) *system.Rule {
				return union(t, system.MustMultiples(3), system.MustMultiples(2))
			},
			super: true,
		},
		{
			name: "integers 0..10 and two adjacent ranges",
			x:    func(testing.TB) *system.Rule { return system.MustIntegerRange(0, 10) },
			y: func(t testing.TB) *system.Rule {
				return union(t, system.MustIntegerRange(0, 4), system.MustIntegerRange(5, 10))
			},
			subset: true, super: true,
		},
		{
			name: "integers 0..10 and two ranges with a gap",
			x:    func(testing.TB) *system.Rule { return system.MustIntegerRange(0, 10) },
			y: func(t testing.TB) *system.Rule {
				return union(t, system.MustIntegerRange(0, 4), system.MustIntegerRange(6, 10))
			},
			super: true,
		},
		{
			name: "non-negative halves and non-negative integers with half plus integers",
			x:    func(t testing.TB) *system.Rule { return rule(t, lower(t, prog(t, a(0), q(1, 2)), a(0))) },
			y: func(t testing.TB) *system.Rule {
				return rule(t, lower(t, prog(t, a(0), a(1)), a(0)), prog(t, q(1, 2), a(1)))
			},
			subset: true,
		},
		{
			name: "halves from -1 and non-negative integers with half plus integers",
			x:    func(t testing.TB) *system.Rule { return rule(t, lower(t, prog(t, a(0), q(1, 2)), a(-1))) },
			y: func(t testing.TB) *system.Rule {
				return rule(t, lower(t, prog(t, a(0), a(1)), a(0)), prog(t, q(1, 2), a(1)))
			},
		},
		{
			name:   "same rule written twice",
			x:      func(testing.TB) *system.Rule { return system.Integers() },
			y:      func(t testing.TB) *system.Rule { return union(t, system.Integers(), system.MustIntegerRange(-2, 2)) },
			subset: true, super: true,
		},
		{
			name: "integers with -M and M versus integers in between",
			x: func(t testing.TB) *system.Rule {
				return union(t, system.MustIntegerRange(-m+1, m-1), system.Single(a(-m)), system.Single(a(m)))
			},
			y:      func(testing.TB) *system.Rule { return system.Integers() },
			subset: true, super: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			x, y := tt.x(t), tt.y(t)

			got, err := x.SubsetOf(y)
			require.NoError(t, err)
			assert.Equal(t, tt.subset, got, "x ⊆ y")
			assert.Equal(t, tt.subset, x.MustSubsetOf(y))

			got, err = y.SubsetOf(x)
			require.NoError(t, err)
			assert.Equal(t, tt.super, got, "y ⊆ x")

			eq := tt.subset && tt.super
			assert.Equal(t, eq, x.Equal(y), "x.Equal(y)")
			assert.Equal(t, eq, y.Equal(x), "y.Equal(x)")

			// Reflexivity.
			assert.True(t, x.MustSubsetOf(x))
			assert.True(t, x.Equal(x))
		})
	}
}

// ---- Enumeration oracle (R-3).

// oracleDen is the common denominator of all values of oracle progressions.
const oracleDen = 12

// oracleWindow bounds all bounds and single values: they lie in [-oracleWindow, oracleWindow].
const oracleWindow = 3

// oracleProg describes a progression with all values in units of 1/oracleDen.
type oracleProg struct {
	start  int  // start, in units
	step   int  // 1/d in units (oracleDen/d), 0 for a single value
	lo, hi *int // inclusive bounds in units, nil for none
}

func ip(v int) *int { return &v }

func (p oracleProg) String() string {
	b := func(v *int) string {
		if v == nil {
			return "none"
		}
		return fmt.Sprintf("%d/%d", *v, oracleDen)
	}
	return fmt.Sprintf("{start %d/%d step %d/%d lo %s hi %s}", p.start, oracleDen, p.step, oracleDen, b(p.lo), b(p.hi))
}

func (p oracleProg) contains(k int) bool {
	if p.lo != nil && k < *p.lo {
		return false
	}
	if p.hi != nil && k > *p.hi {
		return false
	}
	if p.step == 0 {
		return k == p.start
	}
	return (k-p.start)%p.step == 0
}

func (p oracleProg) build(t testing.TB) system.Progression {
	t.Helper()
	r := prog(t, q(p.start, oracleDen), q(p.step, oracleDen))
	if p.lo != nil {
		r = lower(t, r, q(*p.lo, oracleDen))
	}
	if p.hi != nil {
		r = upper(t, r, q(*p.hi, oracleDen))
	}
	return r
}

// oraclePool returns progressions whose bounds and single values lie inside the window.
func oraclePool() []oracleProg {
	const u = oracleDen
	return []oracleProg{
		{start: 0, step: u},                                // Z
		{start: 0, step: u / 2},                            // Z/2
		{start: u / 2, step: u},                            // 1/2 + Z
		{start: 0, step: u / 3},                            // Z/3
		{start: u / 6, step: u / 3},                        // 1/6 + Z/3
		{start: u / 4, step: u / 2},                        // 1/4 + Z/2
		{start: 0, step: u / 6},                            // Z/6
		{start: 0, step: u, lo: ip(-2 * u), hi: ip(3 * u)}, // Z in [-2, 3]
		{start: 0, step: u / 2, lo: ip(0)},                 // Z/2 in [0, +inf]
		{start: 0, step: u, hi: ip(u)},                     // Z in [-inf, 1]
		{start: 0, step: u / 3, lo: ip(-u), hi: ip(u)},     // Z/3 in [-1, 1]
		{start: 0},      // {0}
		{start: u / 2},  // {1/2}
		{start: -3 * u}, // {-3}
	}
}

type oracleRule []oracleProg

func (r oracleRule) contains(k int) bool {
	for _, p := range r {
		if p.contains(k) {
			return true
		}
	}
	return false
}

func (r oracleRule) build(t testing.TB) *system.Rule {
	t.Helper()
	ps := make([]system.Progression, 0, len(r))
	for _, p := range r {
		ps = append(ps, p.build(t))
	}
	return rule(t, ps...)
}

// subsetOf decides r ⊆ other on the window extended by two periods on each side:
// outside the window all progressions are unbounded and periodic with period 1,
// and every unbounded progression reaches the extreme lattice points of [-M, M].
func (r oracleRule) subsetOf(other oracleRule) bool {
	const lim = (oracleWindow + 2) * oracleDen
	for k := -lim; k <= lim; k++ {
		if r.contains(k) && !other.contains(k) {
			return false
		}
	}
	return true
}

func oracleRules() []oracleRule {
	pool := oraclePool()
	out := make([]oracleRule, 0, len(pool)*(len(pool)+1)/2)
	for i := range pool {
		out = append(out, oracleRule{pool[i]})
		for j := i + 1; j < len(pool); j++ {
			out = append(out, oracleRule{pool[i], pool[j]})
		}
	}
	return out
}

func TestRule_SubsetOfOracle(t *testing.T) {
	t.Parallel()

	rules := oracleRules()
	built := make([]*system.Rule, len(rules))
	for i, r := range rules {
		built[i] = r.build(t)
	}

	for i, x := range rules {
		for j, y := range rules {
			want := x.subsetOf(y)
			got, err := built[i].SubsetOf(built[j])
			require.NoError(t, err)
			if !assert.Equal(t, want, got, "%v ⊆ %v", x, y) {
				continue
			}
			wantEq := want && y.subsetOf(x)
			assert.Equal(t, wantEq, built[i].Equal(built[j]), "%v = %v", x, y)
		}
	}
}

func TestRule_ContainsOracle(t *testing.T) {
	t.Parallel()

	const lim = (oracleWindow + 2) * oracleDen
	for _, r := range oracleRules() {
		built := r.build(t)
		for k := -lim; k <= lim; k++ {
			got, err := built.Contains(q(k, oracleDen))
			require.NoError(t, err)
			assert.Equal(t, r.contains(k), got, "%v contains %d/%d", r, k, oracleDen)
		}
	}
}
