package system_test

import (
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for Progression: OP-10, contract 5.4, criteria K-16 (rule part).

func prog(t testing.TB, start, step alteration.Alteration) system.Progression {
	t.Helper()
	p, err := system.NewProgression(start, step)
	require.NoError(t, err)
	require.True(t, p.IsValid())
	return p
}

func lower(t testing.TB, p system.Progression, lo alteration.Alteration) system.Progression {
	t.Helper()
	r, err := p.WithLower(lo)
	require.NoError(t, err)
	return r
}

func upper(t testing.TB, p system.Progression, hi alteration.Alteration) system.Progression {
	t.Helper()
	r, err := p.WithUpper(hi)
	require.NoError(t, err)
	return r
}

func TestProgression_Readers(t *testing.T) {
	t.Parallel()

	m := maxAlt
	tests := []struct {
		name         string
		build        func(t testing.TB) system.Progression
		step         alteration.Alteration
		den          int
		first, last  alteration.Alteration
		boundedBelow bool
		boundedAbove bool
		str          string
	}{
		{
			name:  "integers",
			build: func(t testing.TB) system.Progression { return prog(t, a(0), a(1)) },
			step:  a(1), den: 1, first: a(-m), last: a(m),
			str: "0 + Z/1 in [-inf, +inf]",
		},
		{
			name:  "half plus integers",
			build: func(t testing.TB) system.Progression { return prog(t, q(1, 2), a(1)) },
			step:  a(1), den: 1, first: q(-2*m+1, 2), last: q(2*m-1, 2),
			str: "+1/2 + Z/1 in [-inf, +inf]",
		},
		{
			name:  "halves",
			build: func(t testing.TB) system.Progression { return prog(t, a(0), q(1, 2)) },
			step:  q(1, 2), den: 2, first: a(-m), last: a(m),
			str: "0 + Z/2 in [-inf, +inf]",
		},
		{
			name:  "quarter plus halves",
			build: func(t testing.TB) system.Progression { return prog(t, q(1, 4), q(1, 2)) },
			step:  q(1, 2), den: 2, first: q(-4*m+1, 4), last: q(4*m-1, 4),
			str: "+1/4 + Z/2 in [-inf, +inf]",
		},
		{
			name:  "third plus halves",
			build: func(t testing.TB) system.Progression { return prog(t, q(1, 3), q(1, 2)) },
			step:  q(1, 2), den: 2, first: q(-3*m+1, 3), last: q(6*m-1, 6),
			str: "+1/3 + Z/2 in [-inf, +inf]",
		},
		{
			name:  "smallest step 1/MaxDenominator",
			build: func(t testing.TB) system.Progression { return prog(t, a(0), q(1, system.MaxDenominator)) },
			step:  q(1, system.MaxDenominator), den: system.MaxDenominator, first: a(-m), last: a(m),
			str: "0 + Z/65536 in [-inf, +inf]",
		},
		{
			name:  "single natural",
			build: func(t testing.TB) system.Progression { return prog(t, a(0), a(0)) },
			step:  a(0), den: 0, first: a(0), last: a(0),
			boundedBelow: true, boundedAbove: true,
			str: "{0}",
		},
		{
			name:  "single half",
			build: func(t testing.TB) system.Progression { return prog(t, q(1, 2), a(0)) },
			step:  a(0), den: 0, first: q(1, 2), last: q(1, 2),
			boundedBelow: true, boundedAbove: true,
			str: "{+1/2}",
		},
		{
			name:  "integers from 1",
			build: func(t testing.TB) system.Progression { return lower(t, prog(t, a(0), a(1)), a(1)) },
			step:  a(1), den: 1, first: a(1), last: a(m),
			boundedBelow: true,
			str:          "0 + Z/1 in [+1, +inf]",
		},
		{
			name: "integers from -3 to -1",
			build: func(t testing.TB) system.Progression {
				return upper(t, lower(t, prog(t, a(-3), a(1)), a(-3)), a(-1))
			},
			step: a(1), den: 1, first: a(-3), last: a(-1),
			boundedBelow: true, boundedAbove: true,
			str: "0 + Z/1 in [-3, -1]",
		},
		{
			name:  "integers from -2 to 2",
			build: func(t testing.TB) system.Progression { return upper(t, lower(t, prog(t, a(0), a(1)), a(-2)), a(2)) },
			step:  a(1), den: 1, first: a(-2), last: a(2),
			boundedBelow: true, boundedAbove: true,
			str: "0 + Z/1 in [-2, +2]",
		},
		{
			name:  "lower bound -M+1 is a bound",
			build: func(t testing.TB) system.Progression { return lower(t, prog(t, a(0), a(1)), a(-m+1)) },
			step:  a(1), den: 1, first: a(-m + 1), last: a(m),
			boundedBelow: true,
			str:          "0 + Z/1 in [-16777215, +inf]",
		},
		{
			name:  "upper bound M-1 is a bound",
			build: func(t testing.TB) system.Progression { return upper(t, prog(t, a(0), a(1)), a(m-1)) },
			step:  a(1), den: 1, first: a(-m), last: a(m - 1),
			boundedAbove: true,
			str:          "0 + Z/1 in [-inf, +16777215]",
		},
		{
			name:  "bounds at -M and M are no bounds",
			build: func(t testing.TB) system.Progression { return upper(t, lower(t, prog(t, a(0), a(1)), a(-m)), a(m)) },
			step:  a(1), den: 1, first: a(-m), last: a(m),
			str: "0 + Z/1 in [-inf, +inf]",
		},
		{
			name:  "one point left by bounds becomes single",
			build: func(t testing.TB) system.Progression { return upper(t, lower(t, prog(t, a(0), a(1)), a(2)), a(2)) },
			step:  a(0), den: 0, first: a(2), last: a(2),
			boundedBelow: true, boundedAbove: true,
			str: "{+2}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := tt.build(t)
			assert.True(t, p.IsValid())
			assert.True(t, tt.step.Equal(p.Step()), "Step = %s, want %s", p.Step(), tt.step)
			assert.Equal(t, tt.den, p.Den())
			assert.True(t, tt.first.Equal(p.First()), "First = %s, want %s", p.First(), tt.first)
			assert.True(t, tt.last.Equal(p.Last()), "Last = %s, want %s", p.Last(), tt.last)
			assert.Equal(t, tt.boundedBelow, p.BoundedBelow(), "BoundedBelow")
			assert.Equal(t, tt.boundedAbove, p.BoundedAbove(), "BoundedAbove")
			assert.Equal(t, tt.str, p.String())
		})
	}
}

func TestProgression_Canonical(t *testing.T) {
	t.Parallel()

	m := maxAlt
	tests := []struct {
		name string
		x, y func(t testing.TB) system.Progression
	}{
		{
			name: "integers from different starts",
			x:    func(t testing.TB) system.Progression { return prog(t, a(0), a(1)) },
			y:    func(t testing.TB) system.Progression { return prog(t, a(5), a(1)) },
		},
		{
			name: "half plus integers from different starts",
			x:    func(t testing.TB) system.Progression { return prog(t, q(1, 2), a(1)) },
			y:    func(t testing.TB) system.Progression { return prog(t, q(11, 2), a(1)) },
		},
		{
			name: "start shifted by the step",
			x:    func(t testing.TB) system.Progression { return prog(t, q(1, 4), q(1, 2)) },
			y:    func(t testing.TB) system.Progression { return prog(t, q(7, 4), q(1, 2)) },
		},
		{
			name: "thirds from a third and from zero",
			x:    func(t testing.TB) system.Progression { return prog(t, q(1, 3), q(1, 3)) },
			y:    func(t testing.TB) system.Progression { return prog(t, a(0), q(1, 3)) },
		},
		{
			name: "lower bound off the lattice",
			x:    func(t testing.TB) system.Progression { return lower(t, prog(t, a(0), q(1, 2)), q(1, 3)) },
			y:    func(t testing.TB) system.Progression { return lower(t, prog(t, q(1, 2), q(1, 2)), q(1, 2)) },
		},
		{
			name: "upper bound off the lattice",
			x:    func(t testing.TB) system.Progression { return upper(t, prog(t, a(0), a(1)), q(5, 2)) },
			y:    func(t testing.TB) system.Progression { return upper(t, prog(t, a(0), a(1)), a(2)) },
		},
		{
			name: "lower bound -M is no bound",
			x:    func(t testing.TB) system.Progression { return lower(t, prog(t, a(0), a(1)), a(-m)) },
			y:    func(t testing.TB) system.Progression { return prog(t, a(0), a(1)) },
		},
		{
			name: "upper bound M is no bound",
			x:    func(t testing.TB) system.Progression { return upper(t, prog(t, a(0), a(1)), a(m)) },
			y:    func(t testing.TB) system.Progression { return prog(t, a(0), a(1)) },
		},
		{
			name: "single point from bounds equals single value",
			x:    func(t testing.TB) system.Progression { return upper(t, lower(t, prog(t, a(0), a(1)), a(2)), a(2)) },
			y:    func(t testing.TB) system.Progression { return prog(t, a(2), a(0)) },
		},
		{
			name: "top point from a lower bound equals single value",
			x:    func(t testing.TB) system.Progression { return lower(t, prog(t, a(0), a(1)), a(m)) },
			y:    func(t testing.TB) system.Progression { return prog(t, a(m), a(0)) },
		},
		{
			name: "repeated lower bounds narrow",
			x:    func(t testing.TB) system.Progression { return lower(t, lower(t, prog(t, a(0), a(1)), a(3)), a(0)) },
			y:    func(t testing.TB) system.Progression { return lower(t, prog(t, a(0), a(1)), a(3)) },
		},
		{
			name: "repeated upper bounds narrow",
			x:    func(t testing.TB) system.Progression { return upper(t, upper(t, prog(t, a(0), a(1)), a(-3)), a(0)) },
			y:    func(t testing.TB) system.Progression { return upper(t, prog(t, a(0), a(1)), a(-3)) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			x, y := tt.x(t), tt.y(t)
			assert.True(t, x == y, "x == y: %s vs %s", x, y)
			assert.True(t, x.Equal(y))
			assert.True(t, y.Equal(x))
			assert.Equal(t, x.String(), y.String())

			// A comparable value is usable as a map key (contract 5.3).
			set := map[system.Progression]int{x: 1}
			assert.Equal(t, 1, set[y])
		})
	}
}

func TestProgression_NotEqual(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		x, y func(t testing.TB) system.Progression
	}{
		{
			name: "integers and half plus integers",
			x:    func(t testing.TB) system.Progression { return prog(t, a(0), a(1)) },
			y:    func(t testing.TB) system.Progression { return prog(t, q(1, 2), a(1)) },
		},
		{
			name: "integers and halves",
			x:    func(t testing.TB) system.Progression { return prog(t, a(0), a(1)) },
			y:    func(t testing.TB) system.Progression { return prog(t, a(0), q(1, 2)) },
		},
		{
			name: "integers and bounded integers",
			x:    func(t testing.TB) system.Progression { return prog(t, a(0), a(1)) },
			y:    func(t testing.TB) system.Progression { return lower(t, prog(t, a(0), a(1)), a(-maxAlt+1)) },
		},
		{
			name: "two single values",
			x:    func(t testing.TB) system.Progression { return prog(t, a(0), a(0)) },
			y:    func(t testing.TB) system.Progression { return prog(t, a(1), a(0)) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			x, y := tt.x(t), tt.y(t)
			assert.False(t, x == y)
			assert.False(t, x.Equal(y))
			assert.False(t, y.Equal(x))
		})
	}
}

func TestProgression_Contains(t *testing.T) {
	t.Parallel()

	m := maxAlt
	tests := []struct {
		name  string
		build func(t testing.TB) system.Progression
		in    []alteration.Alteration
		out   []alteration.Alteration
	}{
		{
			name:  "integers",
			build: func(t testing.TB) system.Progression { return prog(t, a(0), a(1)) },
			in:    []alteration.Alteration{a(0), a(1), a(-1), a(m), a(-m), a(12345)},
			out:   []alteration.Alteration{q(1, 2), q(-1, 2), q(1, 3), q(2*m-1, 2)},
		},
		{
			name:  "half plus integers",
			build: func(t testing.TB) system.Progression { return prog(t, q(1, 2), a(1)) },
			in:    []alteration.Alteration{q(1, 2), q(-1, 2), q(-2*m+1, 2), q(2*m-1, 2)},
			out:   []alteration.Alteration{a(0), a(m), a(-m), q(1, 4)},
		},
		{
			name:  "thirds",
			build: func(t testing.TB) system.Progression { return prog(t, a(0), q(1, 3)) },
			in:    []alteration.Alteration{a(0), q(1, 3), q(-2, 3), a(1), a(-m)},
			out:   []alteration.Alteration{q(1, 2), q(1, 6)},
		},
		{
			name:  "smallest step",
			build: func(t testing.TB) system.Progression { return prog(t, a(0), q(1, system.MaxDenominator)) },
			in:    []alteration.Alteration{q(1, system.MaxDenominator), q(-3, 4), a(7)},
			out:   []alteration.Alteration{q(1, 3)},
		},
		{
			name:  "non-negative halves",
			build: func(t testing.TB) system.Progression { return lower(t, prog(t, a(0), q(1, 2)), a(0)) },
			in:    []alteration.Alteration{a(0), q(1, 2), a(m)},
			out:   []alteration.Alteration{q(-1, 2), a(-1), q(1, 4)},
		},
		{
			name:  "integers from -2 to 2",
			build: func(t testing.TB) system.Progression { return upper(t, lower(t, prog(t, a(0), a(1)), a(-2)), a(2)) },
			in:    []alteration.Alteration{a(-2), a(0), a(2)},
			out:   []alteration.Alteration{a(-3), a(3), q(1, 2)},
		},
		{
			name:  "single half",
			build: func(t testing.TB) system.Progression { return prog(t, q(1, 2), a(0)) },
			in:    []alteration.Alteration{q(1, 2)},
			out:   []alteration.Alteration{a(0), q(3, 2), q(-1, 2)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := tt.build(t)
			for _, v := range tt.in {
				assert.True(t, p.Contains(v), "%s must contain %s", p, v)
			}
			for _, v := range tt.out {
				assert.False(t, p.Contains(v), "%s must not contain %s", p, v)
			}
		})
	}
}

func TestNewProgression_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		start, step alteration.Alteration
	}{
		{name: "step 2", start: a(0), step: a(2)},
		{name: "step 2/3", start: a(0), step: q(2, 3)},
		{name: "step 3/2", start: a(0), step: q(3, 2)},
		{name: "negative step -1", start: a(0), step: a(-1)},
		{name: "negative step -1/2", start: a(0), step: q(-1, 2)},
		{name: "lcm of denominators above MaxDenominator", start: q(1, 3), step: q(1, system.MaxDenominator)},
		{name: "lcm of 5 and 16384 above MaxDenominator", start: q(1, 5), step: q(1, 16384)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p, err := system.NewProgression(tt.start, tt.step)
			requireErr(t, err, system.ErrDefinition, system.ErrRule)
			assert.True(t, p == system.Progression{}, "result on error must be the zero value")
			assert.False(t, p.IsValid())

			requirePanicErr(t, system.ErrDefinition, system.ErrRule, func() {
				system.MustNewProgression(tt.start, tt.step)
			})
		})
	}
}

func TestNewProgression_ValidSteps(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		start, step alteration.Alteration
	}{
		{name: "start 1/3 step 1/3", start: q(1, 3), step: q(1, 3)},
		{name: "start 1/3 step 1/2", start: q(1, 3), step: q(1, 2)},
		{name: "start 1/MaxDenominator step 1/2", start: q(1, system.MaxDenominator), step: q(1, 2)},
		{name: "start M step 1", start: a(maxAlt), step: a(1)},
		{name: "start -M step 0", start: a(-maxAlt), step: a(0)},
		{name: "start 1/3 step 0", start: q(1, 3), step: a(0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p, err := system.NewProgression(tt.start, tt.step)
			require.NoError(t, err)
			assert.True(t, p.IsValid())
			assert.True(t, p.Contains(tt.start), "the start is a point of the progression")
			assert.True(t, p == system.MustNewProgression(tt.start, tt.step))
		})
	}
}

func TestProgression_BoundErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		build func(t testing.TB) (system.Progression, error)
	}{
		{
			name: "lower bound above upper bound",
			build: func(t testing.TB) (system.Progression, error) {
				return lower(t, prog(t, a(0), a(1)), a(1)).WithUpper(a(0))
			},
		},
		{
			name: "upper bound below lower bound",
			build: func(t testing.TB) (system.Progression, error) {
				return upper(t, prog(t, a(0), a(1)), a(0)).WithLower(a(1))
			},
		},
		{
			name: "single half with bounds 0 and 1/4",
			build: func(t testing.TB) (system.Progression, error) {
				return lower(t, prog(t, q(1, 2), a(0)), a(0)).WithUpper(q(1, 4))
			},
		},
		{
			name: "no integer between 1/2 and 3/4",
			build: func(t testing.TB) (system.Progression, error) {
				return lower(t, prog(t, a(0), a(1)), q(1, 2)).WithUpper(q(3, 4))
			},
		},
		{
			name: "single half above its value",
			build: func(t testing.TB) (system.Progression, error) {
				return prog(t, q(1, 2), a(0)).WithLower(a(1))
			},
		},
		{
			name: "half plus integers above their top point",
			build: func(t testing.TB) (system.Progression, error) {
				return prog(t, q(1, 2), a(1)).WithLower(a(maxAlt))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p, err := tt.build(t)
			requireErr(t, err, system.ErrDefinition, system.ErrRule)
			assert.True(t, p == system.Progression{})
		})
	}

	t.Run("Must variants panic with the same error", func(t *testing.T) {
		t.Parallel()

		ints := prog(t, a(0), a(1))
		requirePanicErr(t, system.ErrDefinition, system.ErrRule, func() {
			ints.MustWithLower(a(1)).MustWithUpper(a(0))
		})
		requirePanicErr(t, system.ErrDefinition, system.ErrRule, func() {
			ints.MustWithUpper(a(0)).MustWithLower(a(1))
		})
		assert.True(t, ints.MustWithLower(a(1)) == lower(t, ints, a(1)))
		assert.True(t, ints.MustWithUpper(a(1)) == upper(t, ints, a(1)))
	})
}

func TestProgression_Unset(t *testing.T) {
	t.Parallel()

	var p system.Progression

	assert.False(t, p.IsValid())
	assert.Equal(t, -1, p.Den())
	assert.True(t, p.Step().IsNatural())
	assert.True(t, p.First().IsNatural())
	assert.True(t, p.Last().IsNatural())
	assert.False(t, p.BoundedBelow())
	assert.False(t, p.BoundedAbove())
	assert.False(t, p.Contains(a(0)))
	assert.False(t, p.Equal(p), "unset progression is not equal even to itself")
	assert.False(t, p.Equal(prog(t, a(0), a(0))))
	assert.False(t, prog(t, a(0), a(0)).Equal(p))
	assert.Equal(t, "Progression(unset)", p.String())

	_, err := p.WithLower(a(0))
	requireErr(t, err, system.ErrInvalid, nil)
	_, err = p.WithUpper(a(0))
	requireErr(t, err, system.ErrInvalid, nil)
	requirePanicErr(t, system.ErrInvalid, nil, func() { p.MustWithLower(a(0)) })
	requirePanicErr(t, system.ErrInvalid, nil, func() { p.MustWithUpper(a(0)) })
}

func TestProgression_ValueSemantics(t *testing.T) {
	t.Parallel()

	ints := prog(t, a(0), a(1))
	before := ints
	_ = lower(t, ints, a(3))
	_ = upper(t, ints, a(3))
	assert.True(t, ints == before, "WithLower and WithUpper do not change the receiver")
	assert.Equal(t, "0 + Z/1 in [-inf, +inf]", ints.String())
}
