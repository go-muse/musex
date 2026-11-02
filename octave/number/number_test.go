package number_test

import (
	"fmt"
	"testing"

	"github.com/go-muse/musex/octave/number"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// K-1, K-14: Builtins lists exactly the five built-in numberings in the fixed
// order; each element equals the numbering returned by its function and only
// it, except the pair Mexican and Middle C = C5. Numberings are compared only
// with Equal and Label.
func TestBuiltins(t *testing.T) {
	t.Parallel()

	want := builtinTable()
	got := number.Builtins()

	require.Len(t, got, builtinCount)
	assert.Equal(t, builtinCount, cap(got), "a fresh slice of length and capacity 5")

	for i, b := range want {
		assert.True(t, got[i].IsDefined(), "element %d is defined", i)
		assert.Equal(t, b.label, got[i].Label(), "label of element %d", i)
	}

	for i := range got {
		for j, b := range want {
			expected := builtinsEqual(i, j)
			desc := fmt.Sprintf("Builtins()[%d] (%s) and %s", i, want[i].label, b.label)

			assert.Equal(t, expected, got[i].Equal(b.get()), desc)
			assert.Equal(t, expected, b.get().Equal(got[i]), desc+", reversed")
			assert.Equal(t, expected, equalSpec(want[i].spec, b.spec), "oracle: "+desc)
		}
	}
}

// K-1: every call of a built-in function returns a numbering equal to the
// result of any other call, with the same label.
func TestBuiltinFunctionsAreStable(t *testing.T) {
	t.Parallel()

	for _, b := range builtinTable() {
		t.Run(b.label, func(t *testing.T) {
			t.Parallel()

			first, second := b.get(), b.get()
			assert.True(t, first.IsDefined())
			assert.True(t, first.Equal(second))
			assert.True(t, second.Equal(first))
			assert.Equal(t, first.Label(), second.Label())
		})
	}
}

// K-2: the labels of the built-in numberings are exactly these strings.
func TestBuiltinLabels(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		get  func() number.Numbering
		want string
	}{
		{name: "ASPN", get: number.ASPN, want: "ASPN"},
		{name: "FrancoBelgian", get: number.FrancoBelgian, want: "Franco-Belgian"},
		{name: "Mexican", get: number.Mexican, want: "Mexican"},
		{name: "MiddleC3", get: number.MiddleC3, want: "Middle C = C3"},
		{name: "MiddleC5", get: number.MiddleC5, want: "Middle C = C5"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, tc.get().Label())
		})
	}
}

// K-2: own numberings have no label, even when they equal a built-in one.
func TestOwnNumberingsHaveNoLabel(t *testing.T) {
	t.Parallel()

	for _, n := range ownCriteria(t) {
		t.Run(n.name, func(t *testing.T) {
			t.Parallel()

			assert.True(t, n.n.IsDefined())
			assert.Empty(t, n.n.Label())
		})
	}

	t.Run("New(4, 4) equals ASPN without its label", func(t *testing.T) {
		t.Parallel()

		mine := own(t, 4, 4).n
		assert.True(t, mine.Equal(number.ASPN()))
		assert.Empty(t, mine.Label())
		assert.Equal(t, "ASPN", number.ASPN().Label())
	})

	t.Run("New(4, 3, 0) equals Franco-Belgian without its label", func(t *testing.T) {
		t.Parallel()

		mine := own(t, 4, 3, 0).n
		assert.True(t, mine.Equal(number.FrancoBelgian()))
		assert.Empty(t, mine.Label())
	})
}

// K-3: the number of an ASPN octave in each built-in numbering.
func TestFromASPN_Builtins(t *testing.T) {
	t.Parallel()

	aspn := []int{-2, -1, 0, 1, 2, 3, 4, 8, 9}
	shift := func(d int) []int {
		out := make([]int, len(aspn))
		for i, a := range aspn {
			out[i] = a + d
		}

		return out
	}

	table := builtinTable()
	cases := []struct {
		name string
		b    builtin
		want []int
	}{
		{name: "ASPN", b: table[idxASPN], want: shift(0)},
		{name: "Franco-Belgian", b: table[idxFrancoBelgian], want: []int{-4, -3, -2, -1, 1, 2, 3, 7, 8}},
		{name: "Mexican", b: table[idxMexican], want: shift(1)},
		{name: "Middle C = C3", b: table[idxMiddleC3], want: shift(-1)},
		{name: "Middle C = C5", b: table[idxMiddleC5], want: shift(1)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			n := tc.b.get()
			for i, a := range aspn {
				desc := fmt.Sprintf("%s.FromASPN(%d)", tc.name, a)
				want := outcome{value: tc.want[i]}

				require.Equal(t, want, tc.b.spec.fromASPN(a), "oracle agrees with the criterion: "+desc)
				got, err := n.FromASPN(a)
				assertOutcome(t, want, got, err, desc)
			}
		})
	}
}

// K-4: the ASPN number of an octave by its number in a built-in numbering.
func TestToASPN_Builtins(t *testing.T) {
	t.Parallel()

	table := builtinTable()
	cases := []struct {
		b    builtin
		x    int
		want int
	}{
		{b: table[idxASPN], x: -2, want: -2},
		{b: table[idxASPN], x: -1, want: -1},
		{b: table[idxASPN], x: 0, want: 0},
		{b: table[idxASPN], x: 1, want: 1},
		{b: table[idxASPN], x: 2, want: 2},
		{b: table[idxASPN], x: 3, want: 3},
		{b: table[idxASPN], x: 4, want: 4},
		{b: table[idxASPN], x: 8, want: 8},
		{b: table[idxASPN], x: 9, want: 9},
		{b: table[idxFrancoBelgian], x: 3, want: 4},
		{b: table[idxFrancoBelgian], x: 1, want: 2},
		{b: table[idxFrancoBelgian], x: -1, want: 1},
		{b: table[idxFrancoBelgian], x: -2, want: 0},
		{b: table[idxFrancoBelgian], x: -3, want: -1},
		{b: table[idxFrancoBelgian], x: -4, want: -2},
		{b: table[idxFrancoBelgian], x: 7, want: 8},
		{b: table[idxFrancoBelgian], x: 8, want: 9},
		{b: table[idxMexican], x: 5, want: 4},
		{b: table[idxMiddleC3], x: 3, want: 4},
		{b: table[idxMiddleC3], x: -2, want: -1},
		{b: table[idxMiddleC5], x: 0, want: -1},
	}

	for _, tc := range cases {
		name := fmt.Sprintf("%s %d", tc.b.label, tc.x)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			want := outcome{value: tc.want}
			require.Equal(t, want, tc.b.spec.toASPN(tc.x), "oracle agrees with the criterion")

			got, err := tc.b.get().ToASPN(tc.x)
			assertOutcome(t, want, got, err, name)
		})
	}
}

// K-5: for every built-in numbering and the own numberings with skipped
// numbers, FromASPN is strictly increasing, never returns a skipped number,
// matches the oracle and ToASPN inverts it.
func TestRoundTrip(t *testing.T) {
	t.Parallel()

	set := append(builtins(), own(t, 4, 4, 0, 1), own(t, 4, 4, 6))

	for _, n := range set {
		t.Run(n.name, func(t *testing.T) {
			t.Parallel()

			prev, havePrev := 0, false
			for a := -20; a <= 20; a++ {
				desc := fmt.Sprintf("%s, ASPN %d", n.name, a)

				x, err := n.n.FromASPN(a)
				if !assert.NoError(t, err, "FromASPN: "+desc) || !assertOutcome(t, n.spec.fromASPN(a), x, err, desc) {
					return
				}
				assert.False(t, n.spec.isSkipped(x), "result is a skipped number: %s", desc)
				if havePrev {
					assert.Greater(t, x, prev, "strictly increasing: %s", desc)
				}
				prev, havePrev = x, true

				back, err := n.n.ToASPN(x)
				assert.NoError(t, err, "ToASPN(FromASPN): "+desc)
				assertOutcome(t, outcome{value: a}, back, err, "ToASPN(FromASPN): "+desc)
			}
		})
	}
}

// K-6: fixed conversions between built-in numberings.
func TestConvert_Examples(t *testing.T) {
	t.Parallel()

	table := builtinTable()
	cases := []struct {
		x        int
		from, to builtin
		want     int
	}{
		{x: 3, from: table[idxMiddleC3], to: table[idxFrancoBelgian], want: 3},
		{x: 0, from: table[idxMiddleC3], to: table[idxFrancoBelgian], want: -1},
		{x: -1, from: table[idxFrancoBelgian], to: table[idxMiddleC3], want: 0},
		{x: 1, from: table[idxFrancoBelgian], to: table[idxMiddleC5], want: 3},
		{x: 4, from: table[idxASPN], to: table[idxMexican], want: 5},
		{x: 5, from: table[idxMexican], to: table[idxMiddleC5], want: 5},
		{x: -2, from: table[idxMiddleC3], to: table[idxFrancoBelgian], want: -3},
	}

	for _, tc := range cases {
		name := fmt.Sprintf("%s %d to %s", tc.from.label, tc.x, tc.to.label)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			want := outcome{value: tc.want}
			require.Equal(t, want, convertSpec(tc.x, tc.from.spec, tc.to.spec), "oracle agrees with the criterion")

			got, err := number.Convert(tc.x, tc.from.get(), tc.to.get())
			assertOutcome(t, want, got, err, name)
		})
	}
}

// K-6: for all pairs of built-in numberings and every used number x of the
// source for ASPN -10...10: conversion matches the oracle, A -> B -> A and
// A -> A return x, conversion to and from ASPN agrees with ToASPN and FromASPN.
func TestConvert_BuiltinPairs(t *testing.T) {
	t.Parallel()

	for _, from := range builtins() {
		for _, to := range builtins() {
			name := from.name + " to " + to.name
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				for a := -10; a <= 10; a++ {
					x := from.spec.fromASPN(a).value
					desc := fmt.Sprintf("%s, x = %d (ASPN %d)", name, x, a)

					y, err := number.Convert(x, from.n, to.n)
					if !assertOutcome(t, convertSpec(x, from.spec, to.spec), y, err, desc) {
						return
					}

					back, err := number.Convert(y, to.n, from.n)
					assertOutcome(t, outcome{value: x}, back, err, "back: "+desc)

					same, err := number.Convert(x, from.n, from.n)
					assertOutcome(t, outcome{value: x}, same, err, "to itself: "+desc)

					toASPN, err := number.Convert(x, from.n, number.ASPN())
					wantASPN, wantErr := from.n.ToASPN(x)
					assertOutcome(t, outcome{value: wantASPN, err: wantErr}, toASPN, err, "to ASPN: "+desc)
					assert.Equal(t, a, toASPN, "to ASPN: "+desc)

					fromASPN, err := number.Convert(a, number.ASPN(), to.n)
					wantNum, wantErr := to.n.FromASPN(a)
					assertOutcome(t, outcome{value: wantNum, err: wantErr}, fromASPN, err, "from ASPN: "+desc)
					assert.Equal(t, to.spec.fromASPN(a), outcome{value: fromASPN}, "from ASPN, oracle: "+desc)
				}
			})
		}
	}
}

// K-7: own numberings given by a shift.
func TestNew_Shift(t *testing.T) {
	t.Parallel()

	t.Run("New(4, 5)", func(t *testing.T) {
		t.Parallel()

		n := own(t, 4, 5)
		for _, c := range []struct{ a, want int }{{4, 5}, {0, 1}, {-3, -2}} {
			desc := fmt.Sprintf("ASPN %d", c.a)
			require.Equal(t, outcome{value: c.want}, n.spec.fromASPN(c.a), "oracle: "+desc)

			got, err := n.n.FromASPN(c.a)
			assertOutcome(t, outcome{value: c.want}, got, err, desc)
		}
	})

	equal := []struct {
		name string
		n    func(t *testing.T) numbering
		b    builtin
	}{
		{name: "New(0, 0) equals ASPN", n: func(t *testing.T) numbering { return own(t, 0, 0) }, b: builtinTable()[idxASPN]},
		{name: "New(4, 3) equals Middle C = C3", n: func(t *testing.T) numbering { return own(t, 4, 3) }, b: builtinTable()[idxMiddleC3]},
		{name: "New(-1, 0) equals Middle C = C5", n: func(t *testing.T) numbering { return own(t, -1, 0) }, b: builtinTable()[idxMiddleC5]},
	}

	for _, tc := range equal {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			n := tc.n(t)
			require.True(t, equalSpec(n.spec, tc.b.spec), "oracle")
			assert.True(t, n.n.Equal(tc.b.get()))
			assert.True(t, tc.b.get().Equal(n.n))
		})
	}
}

// K-8: own numberings with skipped numbers; the order of skipped numbers does
// not matter and an empty list is a shift.
func TestNew_Skipped(t *testing.T) {
	t.Parallel()

	type check struct {
		toASPN bool // false: FromASPN(in); true: ToASPN(in)
		in     int
		want   int
	}

	cases := []struct {
		name    string
		aspn    int
		num     int
		skipped []int
		checks  []check
	}{
		{
			name: "New(4, 4, 0, 1)", aspn: 4, num: 4, skipped: []int{0, 1},
			checks: []check{
				{in: 3, want: 3}, {in: 2, want: 2}, {in: 1, want: -1}, {in: 0, want: -2},
				{toASPN: true, in: -1, want: 1},
			},
		},
		{
			name: "New(4, 4, 6)", aspn: 4, num: 4, skipped: []int{6},
			checks: []check{
				{in: 5, want: 5}, {in: 6, want: 7}, {in: 7, want: 8},
				{toASPN: true, in: 7, want: 6},
			},
		},
		{
			name: "New(4, 4, -10)", aspn: 4, num: 4, skipped: []int{-10},
			checks: []check{{in: -9, want: -9}, {in: -10, want: -11}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			n := own(t, tc.aspn, tc.num, tc.skipped...)
			for _, c := range tc.checks {
				want := outcome{value: c.want}

				var (
					got    int
					err    error
					oracle outcome
					desc   string
				)
				if c.toASPN {
					desc = fmt.Sprintf("ToASPN(%d)", c.in)
					oracle = n.spec.toASPN(c.in)
					got, err = n.n.ToASPN(c.in)
				} else {
					desc = fmt.Sprintf("FromASPN(%d)", c.in)
					oracle = n.spec.fromASPN(c.in)
					got, err = n.n.FromASPN(c.in)
				}
				require.Equal(t, want, oracle, "oracle: "+desc)
				assertOutcome(t, want, got, err, desc)
			}
		})
	}

	t.Run("empty list equals Middle C = C3", func(t *testing.T) {
		t.Parallel()

		for _, n := range []numbering{own(t, 4, 3), own(t, 4, 3, []int{}...), own(t, 4, 3, nil...)} {
			assert.True(t, n.n.Equal(number.MiddleC3()), n.name)
			assert.True(t, number.MiddleC3().Equal(n.n), n.name)
		}
	})

	t.Run("order of skipped numbers does not matter", func(t *testing.T) {
		t.Parallel()

		a, b := own(t, 4, 4, 0, 1), own(t, 4, 4, 1, 0)
		assert.True(t, a.n.Equal(b.n))
		assert.True(t, b.n.Equal(a.n))
		for x := -20; x <= 20; x++ {
			ga, ea := a.n.FromASPN(x)
			gb, eb := b.n.FromASPN(x)
			assertOutcome(t, a.spec.fromASPN(x), ga, ea, fmt.Sprintf("%s.FromASPN(%d)", a.name, x))
			assertOutcome(t, a.spec.fromASPN(x), gb, eb, fmt.Sprintf("%s.FromASPN(%d)", b.name, x))
		}
	})
}

// K-9: numberings are equal when they give the same numbers to all octaves;
// labels and anchor pairs do not matter.
func TestEqual(t *testing.T) {
	t.Parallel()

	fb := builtinTable()[idxFrancoBelgian]
	aspn := builtinTable()[idxASPN]
	c3 := builtinTable()[idxMiddleC3]
	fromBuiltin := func(b builtin) func(t *testing.T) numbering {
		return func(*testing.T) numbering { return numbering{name: b.label, n: b.get(), spec: b.spec} }
	}
	fromNew := func(aspn, num int, skipped ...int) func(t *testing.T) numbering {
		return func(t *testing.T) numbering { return own(t, aspn, num, skipped...) }
	}

	cases := []struct {
		name string
		p, q func(t *testing.T) numbering
		want bool
	}{
		{name: "New(4, 3, 0) and Franco-Belgian", p: fromNew(4, 3, 0), q: fromBuiltin(fb), want: true},
		{name: "New(2, 1, 0) and Franco-Belgian", p: fromNew(2, 1, 0), q: fromBuiltin(fb), want: true},
		{name: "New(1, -1, 0) and Franco-Belgian", p: fromNew(1, -1, 0), q: fromBuiltin(fb), want: true},
		{name: "New(4, 3, 0) and New(2, 1, 0)", p: fromNew(4, 3, 0), q: fromNew(2, 1, 0), want: true},
		{name: "New(4, 4) and ASPN", p: fromNew(4, 4), q: fromBuiltin(aspn), want: true},
		{name: "Franco-Belgian and Middle C = C3", p: fromBuiltin(fb), q: fromBuiltin(c3), want: false},
		{name: "Franco-Belgian and New(4, 3, 1)", p: fromBuiltin(fb), q: fromNew(4, 3, 1), want: false},
		{name: "ASPN and Middle C = C3", p: fromBuiltin(aspn), q: fromBuiltin(c3), want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p, q := tc.p(t), tc.q(t)
			require.Equal(t, tc.want, equalSpec(p.spec, q.spec), "oracle agrees with the criterion")
			assert.Equal(t, tc.want, p.n.Equal(q.n))
			assert.Equal(t, tc.want, q.n.Equal(p.n), "symmetric")
		})
	}
}

// K-9: Equal is reflexive, symmetric and transitive on the built-in
// numberings and the own numberings of K-7, K-8, K-9, and matches the oracle
// for every pair.
func TestEqual_Properties(t *testing.T) {
	t.Parallel()

	set := criteriaSet(t)

	for _, p := range set {
		assert.True(t, p.n.Equal(p.n), "reflexive: %s", p.name)

		for _, q := range set {
			want := equalSpec(p.spec, q.spec)
			assert.Equal(t, want, p.n.Equal(q.n), "%s and %s", p.name, q.name)
			assert.Equal(t, p.n.Equal(q.n), q.n.Equal(p.n), "symmetric: %s and %s", p.name, q.name)

			for _, r := range set {
				if p.n.Equal(q.n) && q.n.Equal(r.n) {
					assert.True(t, p.n.Equal(r.n), "transitive: %s, %s, %s", p.name, q.name, r.name)
				}
			}
		}
	}
}

// K-14: Mexican equals Middle C = C5, their labels differ and both are listed
// in Builtins at their own places.
func TestMexicanEqualsMiddleC5(t *testing.T) {
	t.Parallel()

	mex, c5 := number.Mexican(), number.MiddleC5()
	assert.True(t, mex.Equal(c5))
	assert.True(t, c5.Equal(mex))
	assert.NotEqual(t, mex.Label(), c5.Label())

	all := number.Builtins()
	require.Len(t, all, builtinCount)
	assert.Equal(t, "Mexican", all[idxMexican].Label())
	assert.Equal(t, "Middle C = C5", all[idxMiddleC5].Label())
}

// K-15: numbers far from middle C continue the same rule: Franco-Belgian
// outside -2...7, Mexican below number 1, and ASPN -100 and 100.
func TestFarOctaves(t *testing.T) {
	t.Parallel()

	table := builtinTable()
	type check struct {
		b      builtin
		toASPN bool
		in     int
		want   int
	}

	cases := []check{
		{b: table[idxFrancoBelgian], in: -1, want: -3},
		{b: table[idxFrancoBelgian], in: -5, want: -7},
		{b: table[idxFrancoBelgian], in: 9, want: 8},
		{b: table[idxFrancoBelgian], in: 20, want: 19},
		{b: table[idxFrancoBelgian], toASPN: true, in: -7, want: -5},
		{b: table[idxMexican], in: -1, want: 0},
		{b: table[idxMexican], in: -2, want: -1},
		{b: table[idxMexican], in: -5, want: -4},
		{b: table[idxMexican], toASPN: true, in: 0, want: -1},
		{b: table[idxMexican], toASPN: true, in: -4, want: -5},
	}

	for _, a := range []int{-100, 100} {
		cases = append(cases,
			check{b: table[idxASPN], in: a, want: a},
			check{b: table[idxMexican], in: a, want: a + 1},
			check{b: table[idxMiddleC3], in: a, want: a - 1},
			check{b: table[idxMiddleC5], in: a, want: a + 1},
			check{b: table[idxASPN], toASPN: true, in: a, want: a},
			check{b: table[idxMexican], toASPN: true, in: a + 1, want: a},
			check{b: table[idxMiddleC3], toASPN: true, in: a - 1, want: a},
			check{b: table[idxMiddleC5], toASPN: true, in: a + 1, want: a},
		)
	}
	cases = append(cases,
		check{b: table[idxFrancoBelgian], in: -100, want: -102},
		check{b: table[idxFrancoBelgian], in: 100, want: 99},
	)

	for _, c := range cases {
		op := "FromASPN"
		if c.toASPN {
			op = "ToASPN"
		}
		name := fmt.Sprintf("%s.%s(%d)", c.b.label, op, c.in)

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			want := outcome{value: c.want}
			var got int
			var err error
			if c.toASPN {
				require.Equal(t, want, c.b.spec.toASPN(c.in), "oracle")
				got, err = c.b.get().ToASPN(c.in)
			} else {
				require.Equal(t, want, c.b.spec.fromASPN(c.in), "oracle")
				got, err = c.b.get().FromASPN(c.in)
			}
			assertOutcome(t, want, got, err, name)
		})
	}
}

// K-18: for ASPN -20...20 and all numberings of K-1, K-7, K-8 (and K-9),
// FromASPN, ToASPN and Convert of used numbers give no error, do not panic
// and match the oracle.
func TestUsedNumbersHaveNoErrors(t *testing.T) {
	t.Parallel()

	set := criteriaSet(t)

	for _, from := range set {
		t.Run(from.name, func(t *testing.T) {
			t.Parallel()

			for a := -20; a <= 20; a++ {
				desc := fmt.Sprintf("%s, ASPN %d", from.name, a)

				var x int
				var err error
				require.NotPanics(t, func() { x, err = from.n.FromASPN(a) }, desc)
				if !assert.NoError(t, err, "FromASPN: "+desc) || !assertOutcome(t, from.spec.fromASPN(a), x, err, "FromASPN: "+desc) {
					return
				}

				var back int
				require.NotPanics(t, func() { back, err = from.n.ToASPN(x) }, desc)
				assert.NoError(t, err, "ToASPN: "+desc)
				assertOutcome(t, outcome{value: a}, back, err, "ToASPN: "+desc)

				for _, to := range set {
					var y int
					require.NotPanics(t, func() { y, err = number.Convert(x, from.n, to.n) }, desc)
					if !assert.NoError(t, err, "Convert to "+to.name+": "+desc) ||
						!assertOutcome(t, to.spec.fromASPN(a), y, err, "Convert to "+to.name+": "+desc) {
						return
					}
				}
			}
		})
	}
}
