package system_test

import (
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for Rule: OP-11, OP-12, OP-14, OP-15, contract 5.5, criteria K-15, K-16, K-18.

func single(t testing.TB, v alteration.Alteration) system.Progression {
	t.Helper()
	return prog(t, v, a(0))
}

// singles returns n distinct single-value progressions {0}, {1}, …, {n-1}.
func singles(t testing.TB, n int) []system.Progression {
	t.Helper()
	out := make([]system.Progression, 0, n)
	for i := range n {
		out = append(out, single(t, a(i)))
	}
	return out
}

func rule(t testing.TB, ps ...system.Progression) *system.Rule {
	t.Helper()
	r, err := system.NewRule(ps...)
	require.NoError(t, err)
	require.True(t, r.IsValid())
	return r
}

func union(t testing.TB, rules ...*system.Rule) *system.Rule {
	t.Helper()
	r, err := system.Union(rules...)
	require.NoError(t, err)
	require.True(t, r.IsValid())
	return r
}

func TestRule_ReadyRules(t *testing.T) {
	t.Parallel()

	m := maxAlt
	tests := []struct {
		name  string
		build func(t testing.TB) *system.Rule
		progs func(t testing.TB) []system.Progression
		str   string
	}{
		{
			name:  "Integers",
			build: func(testing.TB) *system.Rule { return system.Integers() },
			progs: func(t testing.TB) []system.Progression { return []system.Progression{prog(t, a(0), a(1))} },
			str:   "0 + Z/1 in [-inf, +inf]",
		},
		{
			name:  "Multiples(1)",
			build: func(testing.TB) *system.Rule { return system.MustMultiples(1) },
			progs: func(t testing.TB) []system.Progression { return []system.Progression{prog(t, a(0), a(1))} },
			str:   "0 + Z/1 in [-inf, +inf]",
		},
		{
			name:  "Multiples(2)",
			build: func(testing.TB) *system.Rule { return system.MustMultiples(2) },
			progs: func(t testing.TB) []system.Progression { return []system.Progression{prog(t, a(0), q(1, 2))} },
			str:   "0 + Z/2 in [-inf, +inf]",
		},
		{
			name:  "Multiples(MaxDenominator)",
			build: func(testing.TB) *system.Rule { return system.MustMultiples(system.MaxDenominator) },
			progs: func(t testing.TB) []system.Progression {
				return []system.Progression{prog(t, a(0), q(1, system.MaxDenominator))}
			},
			str: "0 + Z/65536 in [-inf, +inf]",
		},
		{
			name:  "IntegerRange(-2, 2)",
			build: func(testing.TB) *system.Rule { return system.MustIntegerRange(-2, 2) },
			progs: func(t testing.TB) []system.Progression {
				return []system.Progression{upper(t, lower(t, prog(t, a(0), a(1)), a(-2)), a(2))}
			},
			str: "0 + Z/1 in [-2, +2]",
		},
		{
			name:  "IntegerRange(3, 3)",
			build: func(testing.TB) *system.Rule { return system.MustIntegerRange(3, 3) },
			progs: func(t testing.TB) []system.Progression { return []system.Progression{single(t, a(3))} },
			str:   "{+3}",
		},
		{
			name:  "IntegerRange(-M, M) is the integers",
			build: func(testing.TB) *system.Rule { return system.MustIntegerRange(-m, m) },
			progs: func(t testing.TB) []system.Progression { return []system.Progression{prog(t, a(0), a(1))} },
			str:   "0 + Z/1 in [-inf, +inf]",
		},
		{
			name:  "Single(0)",
			build: func(testing.TB) *system.Rule { return system.Single(a(0)) },
			progs: func(t testing.TB) []system.Progression { return []system.Progression{single(t, a(0))} },
			str:   "{0}",
		},
		{
			name:  "Single(1/2)",
			build: func(testing.TB) *system.Rule { return system.Single(q(1, 2)) },
			progs: func(t testing.TB) []system.Progression { return []system.Progression{single(t, q(1, 2))} },
			str:   "{+1/2}",
		},
		{
			name: "Union keeps the order of progressions",
			build: func(t testing.TB) *system.Rule {
				return union(t, system.Integers(), system.Single(q(1, 2)), system.Single(q(-1, 2)))
			},
			progs: func(t testing.TB) []system.Progression {
				return []system.Progression{prog(t, a(0), a(1)), single(t, q(1, 2)), single(t, q(-1, 2))}
			},
			str: "0 + Z/1 in [-inf, +inf] | {+1/2} | {-1/2}",
		},
		{
			name:  "Union drops exact repeats",
			build: func(t testing.TB) *system.Rule { return union(t, system.Integers(), system.MustMultiples(1)) },
			progs: func(t testing.TB) []system.Progression { return []system.Progression{prog(t, a(0), a(1))} },
			str:   "0 + Z/1 in [-inf, +inf]",
		},
		{
			name: "NewRule drops exact repeats and keeps the first",
			build: func(t testing.TB) *system.Rule {
				return rule(t, single(t, a(1)), prog(t, a(0), a(1)), single(t, a(1)), prog(t, a(7), a(1)))
			},
			progs: func(t testing.TB) []system.Progression {
				return []system.Progression{single(t, a(1)), prog(t, a(0), a(1))}
			},
			str: "{+1} | 0 + Z/1 in [-inf, +inf]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := tt.build(t)
			require.NotNil(t, r)
			assert.True(t, r.IsValid())
			assert.Equal(t, tt.progs(t), r.Progressions())
			assert.Equal(t, tt.str, r.String())
		})
	}
}

func TestRule_NewValueOnEachCall(t *testing.T) {
	t.Parallel()

	assert.NotSame(t, system.Integers(), system.Integers())
	assert.NotSame(t, system.Single(a(0)), system.Single(a(0)))
	assert.True(t, system.Integers().Equal(system.Integers()))
}

func TestRule_MaxProgressions(t *testing.T) {
	t.Parallel()

	t.Run("MaxProgressions distinct progressions", func(t *testing.T) {
		t.Parallel()

		ps := singles(t, system.MaxProgressions)
		r := rule(t, ps...)
		assert.Equal(t, ps, r.Progressions())

		u := union(t, rule(t, ps[:4]...), rule(t, ps[4:]...))
		assert.Equal(t, ps, u.Progressions())
	})

	t.Run("one more after removing repeats", func(t *testing.T) {
		t.Parallel()

		ps := singles(t, system.MaxProgressions+1)
		_, err := system.NewRule(ps...)
		requireErr(t, err, system.ErrDefinition, system.ErrRule)
		requirePanicErr(t, system.ErrDefinition, system.ErrRule, func() { system.MustNewRule(ps...) })

		rs := make([]*system.Rule, 0, len(ps))
		for _, p := range ps {
			rs = append(rs, rule(t, p))
		}
		_, err = system.Union(rs...)
		requireErr(t, err, system.ErrDefinition, system.ErrRule)
		requirePanicErr(t, system.ErrDefinition, system.ErrRule, func() { system.MustUnion(rs...) })
	})

	t.Run("repeats are not counted", func(t *testing.T) {
		t.Parallel()

		ps := singles(t, system.MaxProgressions)
		withRepeat := append([]system.Progression{ps[3]}, ps...)
		r := rule(t, withRepeat...)
		want := append([]system.Progression{ps[3]}, ps[:3]...)
		want = append(want, ps[4:]...)
		assert.Equal(t, want, r.Progressions())
	})
}

func TestRule_CreationErrors(t *testing.T) {
	t.Parallel()

	m := maxAlt
	tests := []struct {
		name     string
		call     func() (*system.Rule, error)
		must     func()
		category error
		cause    error
	}{
		{
			name: "NewRule without progressions",
			call: func() (*system.Rule, error) { return system.NewRule() },
			must: func() { system.MustNewRule() }, category: system.ErrDefinition, cause: system.ErrRule,
		},
		{
			name: "NewRule with an unset progression",
			call: func() (*system.Rule, error) {
				return system.NewRule(system.MustNewProgression(a(0), a(1)), system.Progression{})
			},
			must:     func() { system.MustNewRule(system.Progression{}) },
			category: system.ErrInvalid,
		},
		{
			name: "NewRule with an unset progression and too many progressions",
			call: func() (*system.Rule, error) {
				ps := make([]system.Progression, 0, system.MaxProgressions+2)
				for i := range system.MaxProgressions + 1 {
					ps = append(ps, system.MustNewProgression(a(i), a(0)))
				}
				return system.NewRule(append(ps, system.Progression{})...)
			},
			must:     func() { system.MustNewRule(system.Progression{}, system.Progression{}) },
			category: system.ErrInvalid,
		},
		{
			name: "Multiples(0)",
			call: func() (*system.Rule, error) { return system.Multiples(0) },
			must: func() { system.MustMultiples(0) }, category: system.ErrDefinition, cause: system.ErrRule,
		},
		{
			name: "Multiples(-1)",
			call: func() (*system.Rule, error) { return system.Multiples(-1) },
			must: func() { system.MustMultiples(-1) }, category: system.ErrDefinition, cause: system.ErrRule,
		},
		{
			name:     "Multiples(MaxDenominator+1)",
			call:     func() (*system.Rule, error) { return system.Multiples(system.MaxDenominator + 1) },
			must:     func() { system.MustMultiples(system.MaxDenominator + 1) },
			category: system.ErrDefinition, cause: system.ErrRule,
		},
		{
			name: "IntegerRange(1, 0)",
			call: func() (*system.Rule, error) { return system.IntegerRange(1, 0) },
			must: func() { system.MustIntegerRange(1, 0) }, category: system.ErrDefinition, cause: system.ErrRule,
		},
		{
			name: "IntegerRange(-M-1, 0)",
			call: func() (*system.Rule, error) { return system.IntegerRange(-m-1, 0) },
			must: func() { system.MustIntegerRange(-m-1, 0) }, category: system.ErrDefinition, cause: system.ErrRule,
		},
		{
			name: "IntegerRange(0, M+1)",
			call: func() (*system.Rule, error) { return system.IntegerRange(0, m+1) },
			must: func() { system.MustIntegerRange(0, m+1) }, category: system.ErrDefinition, cause: system.ErrRule,
		},
		{
			name: "Union without rules",
			call: func() (*system.Rule, error) { return system.Union() },
			must: func() { system.MustUnion() }, category: system.ErrDefinition, cause: system.ErrRule,
		},
		{
			name: "Union with nil",
			call: func() (*system.Rule, error) { return system.Union(system.Integers(), nil) },
			must: func() { system.MustUnion(nil) }, category: system.ErrInvalid,
		},
		{
			name: "Union with an unset rule",
			call: func() (*system.Rule, error) { return system.Union(&system.Rule{}, system.Integers()) },
			must: func() { system.MustUnion(&system.Rule{}) }, category: system.ErrInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r, err := tt.call()
			requireErr(t, err, tt.category, tt.cause)
			assert.Nil(t, r, "result on error must be nil")
			requirePanicErr(t, tt.category, tt.cause, tt.must)
		})
	}
}

func TestRule_Contains(t *testing.T) {
	t.Parallel()

	m := maxAlt
	tests := []struct {
		name  string
		build func(t testing.TB) *system.Rule
		in    []alteration.Alteration
		out   []alteration.Alteration
	}{
		{
			name:  "integers",
			build: func(testing.TB) *system.Rule { return system.Integers() },
			in:    []alteration.Alteration{a(0), a(1), a(-2), a(m), a(-m)},
			out:   []alteration.Alteration{q(1, 2), q(-1, 3)},
		},
		{
			name:  "halves",
			build: func(testing.TB) *system.Rule { return system.MustMultiples(2) },
			in:    []alteration.Alteration{a(0), q(1, 2), q(-3, 2), a(m)},
			out:   []alteration.Alteration{q(1, 3), q(1, 4)},
		},
		{
			name:  "integers from -2 to 2",
			build: func(testing.TB) *system.Rule { return system.MustIntegerRange(-2, 2) },
			in:    []alteration.Alteration{a(-2), a(-1), a(0), a(1), a(2)},
			out:   []alteration.Alteration{a(-3), a(3), q(1, 2)},
		},
		{
			name:  "single zero",
			build: func(testing.TB) *system.Rule { return system.Single(a(0)) },
			in:    []alteration.Alteration{a(0)},
			out:   []alteration.Alteration{a(1), a(-1), q(1, 2)},
		},
		{
			name: "integers with two quarter tones",
			build: func(t testing.TB) *system.Rule {
				return union(t, system.Integers(), system.Single(q(1, 2)), system.Single(q(-1, 2)))
			},
			in:  []alteration.Alteration{a(0), a(5), q(1, 2), q(-1, 2)},
			out: []alteration.Alteration{q(3, 2), q(-3, 2)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := tt.build(t)
			for _, v := range tt.in {
				got, err := r.Contains(v)
				require.NoError(t, err)
				assert.True(t, got, "%s must contain %s", r, v)
				assert.True(t, r.MustContains(v))
			}
			for _, v := range tt.out {
				got, err := r.Contains(v)
				require.NoError(t, err)
				assert.False(t, got, "%s must not contain %s", r, v)
				assert.False(t, r.MustContains(v))
			}
		})
	}
}

func TestRule_Values(t *testing.T) {
	t.Parallel()

	m := maxAlt
	tests := []struct {
		name  string
		build func(t testing.TB) *system.Rule
		limit alteration.Alteration
		want  []alteration.Alteration
	}{
		{
			name:  "halves, limit 1 (K-15)",
			build: func(testing.TB) *system.Rule { return system.MustMultiples(2) },
			limit: a(1),
			want:  []alteration.Alteration{a(-1), q(-1, 2), a(0), q(1, 2), a(1)},
		},
		{
			name:  "integers from -2 to 2, limit 5 (K-15)",
			build: func(testing.TB) *system.Rule { return system.MustIntegerRange(-2, 2) },
			limit: a(5),
			want:  []alteration.Alteration{a(-2), a(-1), a(0), a(1), a(2)},
		},
		{
			name:  "single zero (K-15)",
			build: func(testing.TB) *system.Rule { return system.Single(a(0)) },
			limit: a(3),
			want:  []alteration.Alteration{a(0)},
		},
		{
			name:  "limit 0",
			build: func(testing.TB) *system.Rule { return system.MustMultiples(2) },
			limit: a(0),
			want:  []alteration.Alteration{a(0)},
		},
		{
			name:  "fractional limit",
			build: func(testing.TB) *system.Rule { return system.Integers() },
			limit: q(3, 2),
			want:  []alteration.Alteration{a(-1), a(0), a(1)},
		},
		{
			name:  "empty",
			build: func(testing.TB) *system.Rule { return system.Single(a(5)) },
			limit: a(1),
			want:  []alteration.Alteration{},
		},
		{
			name: "overlapping progressions without repeats",
			build: func(t testing.TB) *system.Rule {
				return union(t, system.Integers(), system.MustIntegerRange(-2, 2))
			},
			limit: a(3),
			want:  []alteration.Alteration{a(-3), a(-2), a(-1), a(0), a(1), a(2), a(3)},
		},
		{
			name: "ascending whatever the order of progressions",
			build: func(t testing.TB) *system.Rule {
				return union(t, system.Single(q(1, 2)), system.Single(q(-1, 2)), system.Integers())
			},
			limit: a(1),
			want:  []alteration.Alteration{a(-1), q(-1, 2), a(0), q(1, 2), a(1)},
		},
		{
			name: "bounded progressions",
			build: func(t testing.TB) *system.Rule {
				return rule(t, lower(t, prog(t, a(0), a(1)), a(1)), upper(t, prog(t, q(1, 2), a(1)), q(-1, 2)))
			},
			limit: a(2),
			want:  []alteration.Alteration{q(-3, 2), q(-1, 2), a(1), a(2)},
		},
		{
			name:  "limit M on a single value -M",
			build: func(testing.TB) *system.Rule { return system.Single(a(-m)) },
			limit: a(m),
			want:  []alteration.Alteration{a(-m)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := tt.build(t)
			seq, err := r.Values(tt.limit)
			require.NoError(t, err)
			got := collect(t, seq)
			assertAlterations(t, tt.want, got)

			// Repeated traversal gives the same values.
			assertAlterations(t, tt.want, collect(t, seq))
			assertAlterations(t, tt.want, collect(t, r.MustValues(tt.limit)))
		})
	}

	t.Run("lazy and interruptible on a huge set", func(t *testing.T) {
		t.Parallel()

		seq, err := system.Integers().Values(a(m))
		require.NoError(t, err)
		assertAlterations(t, []alteration.Alteration{a(-m), a(-m + 1), a(-m + 2)}, first(t, seq, 3))
		assertAlterations(t, []alteration.Alteration{a(-m)}, first(t, seq, 1))
	})

	t.Run("errors", func(t *testing.T) {
		t.Parallel()

		seq, err := system.Integers().Values(a(-1))
		requireErr(t, err, system.ErrLimit, nil)
		assert.Nil(t, seq)
		_, err = system.Integers().Values(q(-1, 2))
		requireErr(t, err, system.ErrLimit, nil)
		requirePanicErr(t, system.ErrLimit, nil, func() { system.Integers().MustValues(a(-1)) })

		var unset *system.Rule
		_, err = unset.Values(a(-1))
		requireErr(t, err, system.ErrInvalid, nil) // E-1 before E-4
		_, err = (&system.Rule{}).Values(a(1))
		requireErr(t, err, system.ErrInvalid, nil)
		requirePanicErr(t, system.ErrInvalid, nil, func() { unset.MustValues(a(1)) })
	})
}

func assertAlterations(t testing.TB, want, got []alteration.Alteration) {
	t.Helper()
	if !assert.Len(t, got, len(want)) {
		t.Logf("got %v, want %v", got, want)
		return
	}
	for i := range want {
		assert.True(t, want[i].Equal(got[i]), "element %d: got %s, want %s", i, got[i], want[i])
	}
}

func TestRule_ProgressionsOwnership(t *testing.T) {
	t.Parallel()

	r := union(t, system.Integers(), system.Single(q(1, 2)))
	got := r.Progressions()
	require.Len(t, got, 2)
	got[0] = single(t, a(9))
	got[1] = single(t, a(10))

	assert.Equal(t, []system.Progression{prog(t, a(0), a(1)), single(t, q(1, 2))}, r.Progressions())
	assert.Equal(t, "0 + Z/1 in [-inf, +inf] | {+1/2}", r.String())
}

func TestRule_Unset(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		r    *system.Rule
	}{
		{"nil", nil},
		{"zero value", &system.Rule{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := tc.r
			assert.False(t, r.IsValid())
			ps := r.Progressions()
			assert.Empty(t, ps)
			assert.Equal(t, "Rule(unset)", r.String())

			_, err := r.Contains(a(0))
			requireErr(t, err, system.ErrInvalid, nil)
			requirePanicErr(t, system.ErrInvalid, nil, func() { r.MustContains(a(0)) })

			_, err = r.Values(a(1))
			requireErr(t, err, system.ErrInvalid, nil)

			_, err = r.SubsetOf(system.Integers())
			requireErr(t, err, system.ErrInvalid, nil)
			_, err = system.Integers().SubsetOf(r)
			requireErr(t, err, system.ErrInvalid, nil)
			requirePanicErr(t, system.ErrInvalid, nil, func() { r.MustSubsetOf(system.Integers()) })

			assert.False(t, r.Equal(r))
			assert.False(t, r.Equal(system.Integers()))
			assert.False(t, system.Integers().Equal(r))
		})
	}
}
