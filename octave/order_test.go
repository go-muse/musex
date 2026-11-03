package octave_test

import (
	"cmp"
	"fmt"
	"slices"
	"testing"

	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/octave/number"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// namedOctave is an octave obtained in a described way.
type namedOctave struct {
	name string
	o    octave.Octave
}

// middleCWays returns the octave of middle C obtained in the ways of K-6.
func middleCWays(t *testing.T) []namedOctave {
	t.Helper()

	get := func(name string, o octave.Octave, err error) namedOctave {
		require.NoError(t, err, name)

		return namedOctave{name: name, o: o}
	}

	byFB, errFB := octave.FromNumber(3, fb())
	byMex, errMex := octave.FromNumber(5, mex())
	byRu, errRu := octave.FromName("первая", naming.Russian())
	byEn, errEn := octave.FromName("one-line", naming.English())
	shifted, errShift := oct(3).Shift(1)
	down, errDown := oct(5).ShiftDown(1)
	up, errUp := oct(3).Up()

	return []namedOctave{
		get("O(3, FB)", byFB, errFB),
		{name: "O(4)", o: oct(4)},
		get("O(5, Mex)", byMex, errMex),
		get("P(первая, Ru)", byRu, errRu),
		get("P(one-line, En)", byEn, errEn),
		get("Shift(O(3), 1)", shifted, errShift),
		get("ShiftDown(O(5), 1)", down, errDown),
		get("Up(O(3))", up, errUp),
	}
}

// K-6, C-6, G-9, NF-4: one octave obtained in different ways is Equal and ==
// pairwise; different octaves are not equal.
func TestEqual(t *testing.T) {
	t.Parallel()

	t.Run("same octave", func(t *testing.T) {
		t.Parallel()

		ways := middleCWays(t)
		for _, a := range ways {
			for _, b := range ways {
				assert.True(t, a.o.Equal(b.o), "%s Equal %s", a.name, b.name)
				assert.True(t, a.o == b.o, "%s == %s", a.name, b.name)
			}
		}
	})

	cases := []struct {
		name string
		a, b octave.Octave
	}{
		{name: "O(4) and O(5)", a: oct(4), b: oct(5)},
		{name: "O(4) and O(4, FB)", a: oct(4), b: octave.MustFromNumber(4, fb())},
		{name: "O(MIN) and O(MAX)", a: oct(minInt), b: oct(maxInt)},
		{name: "O(-1) and O(1)", a: oct(-1), b: oct(1)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.False(t, tc.a.Equal(tc.b), "Equal")
			assert.False(t, tc.b.Equal(tc.a), "Equal reversed")
			assert.False(t, tc.a == tc.b, "==")
			assert.True(t, tc.a != tc.b, "!=")
		})
	}
}

// K-6: Equal is reflexive, symmetric and transitive on the octaves of K-1
// (every a in -20...20 in each numbering of K-1), and == agrees with Equal on
// octaves. The expected equality is equality of the ASPN numbers given by the
// formulas of the numberings.
func TestEqualIsEquivalence(t *testing.T) {
	t.Parallel()

	type entry struct {
		desc string
		aspn int
		o    octave.Octave
	}

	var all []entry
	for _, nc := range criteriaNumberings() {
		for _, a := range span(-20, 20) {
			x, ok := nc.number(a)
			require.True(t, ok, "oracle")
			o, err := octave.FromNumber(x, nc.n)
			require.NoError(t, err, "FromNumber(%d, %s)", x, nc.name)
			all = append(all, entry{desc: fmt.Sprintf("O(%d, %s)", x, nc.name), aspn: a, o: o})
		}
	}

	for _, x := range all {
		assert.True(t, x.o.Equal(x.o), "reflexive: %s", x.desc)
		for _, y := range all {
			want := x.aspn == y.aspn
			assert.Equal(t, want, x.o.Equal(y.o), "%s Equal %s", x.desc, y.desc)
			assert.Equal(t, x.o.Equal(y.o), y.o.Equal(x.o), "symmetric: %s, %s", x.desc, y.desc)
			assert.Equal(t, want, x.o == y.o, "%s == %s", x.desc, y.desc)
		}
	}

	for _, x := range all {
		for _, y := range all {
			if !x.o.Equal(y.o) {
				continue
			}
			for _, z := range all {
				if y.o.Equal(z.o) {
					assert.True(t, x.o.Equal(z.o), "transitive: %s, %s, %s", x.desc, y.desc, z.desc)
				}
			}
		}
	}
}

// K-6, NF-4: two records of one octave obtained in different ways are one
// map key.
func TestOctaveAsMapKey(t *testing.T) {
	t.Parallel()

	m := map[octave.Octave]int{}
	m[octave.MustFromNumber(3, fb())] = 1
	m[octave.MustFromName("первая", naming.Russian())] = 2
	require.Len(t, m, 1)
	assert.Equal(t, 2, m[oct(4)])

	set := map[octave.Octave]bool{}
	for _, w := range middleCWays(t) {
		set[w.o] = true
	}
	assert.Len(t, set, 1, "all ways to the octave of middle C give one key")
}

// K-7, C-7, G-2: comparison of octaves, including the limits of the axis.
func TestCompare(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		a, b octave.Octave
		want int
	}{
		{name: "O(3) O(4)", a: oct(3), b: oct(4), want: -1},
		{name: "O(4) O(4)", a: oct(4), b: oct(4), want: 0},
		{name: "O(5) O(4)", a: oct(5), b: oct(4), want: 1},
		{name: "O(MIN) O(MAX)", a: oct(minInt), b: oct(maxInt), want: -1},
		{name: "O(MAX) O(MIN)", a: oct(maxInt), b: oct(minInt), want: 1},
		{name: "O(MIN) O(MIN)", a: oct(minInt), b: oct(minInt), want: 0},
		{name: "O(MAX) O(MAX)", a: oct(maxInt), b: oct(maxInt), want: 0},
		{name: "O(MAX) O(-1)", a: oct(maxInt), b: oct(-1), want: 1},
		{name: "O(MIN) O(0)", a: oct(minInt), b: oct(0), want: -1},
		{name: "O(3, FB) O(4)", a: octave.MustFromNumber(3, fb()), b: oct(4), want: 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var got int
			require.NotPanics(t, func() { got = tc.a.Compare(tc.b) })
			assert.Equal(t, tc.want, got)
		})
	}
}

// K-7, C-7: sorting with slices.SortFunc and Compare.
func TestSort(t *testing.T) {
	t.Parallel()

	xs := octs(5, -1, 4, maxInt, minInt, 0)
	slices.SortFunc(xs, octave.Octave.Compare)
	want := octs(minInt, -1, 0, 4, 5, maxInt)
	assert.Equal(t, want, xs, "sorted %s", describe(xs))
}

// K-7: for all a, b in -10...10, Compare agrees with the order of the ASPN
// numbers, with the sign of Sub, with the order of the numbers in each of the
// five built-in numberings, and Compare == 0 exactly when Equal.
func TestCompareConsistency(t *testing.T) {
	t.Parallel()

	for _, a := range span(-10, 10) {
		for _, b := range span(-10, 10) {
			got := oct(a).Compare(oct(b))
			assert.Equal(t, cmp.Compare(a, b), got, "Compare(O(%d), O(%d))", a, b)

			sub, err := oct(a).Sub(oct(b))
			require.NoError(t, err, "Sub(O(%d), O(%d))", a, b)
			assert.Equal(t, sign(sub), got, "sign of Sub(O(%d), O(%d))", a, b)

			for _, nc := range builtinNumberings() {
				na, nb := oct(a).MustNumber(nc.n), oct(b).MustNumber(nc.n)
				assert.Equal(t, cmp.Compare(na, nb), got, "Compare(O(%d), O(%d)) vs numbers in %s", a, b, nc.name)
			}

			assert.Equal(t, got == 0, oct(a).Equal(oct(b)), "Compare == 0 vs Equal: O(%d), O(%d)", a, b)
		}
	}
}

// K-8, C-8: above and below.
func TestAboveBelow(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		a, b         octave.Octave
		above, below bool
	}{
		{name: "O(5) O(4)", a: oct(5), b: oct(4), above: true},
		{name: "O(4) O(4)", a: oct(4), b: oct(4)},
		{name: "O(4) O(5)", a: oct(4), b: oct(5), below: true},
		{name: "O(MAX) O(MIN)", a: oct(maxInt), b: oct(minInt), above: true},
		{name: "O(MIN) O(MAX)", a: oct(minInt), b: oct(maxInt), below: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.above, tc.a.Above(tc.b), "Above")
			assert.Equal(t, tc.below, tc.a.Below(tc.b), "Below")
		})
	}

	t.Run("agrees with Compare", func(t *testing.T) {
		t.Parallel()

		for _, a := range span(-5, 5) {
			for _, b := range span(-5, 5) {
				c := oct(a).Compare(oct(b))
				assert.Equal(t, c == 1, oct(a).Above(oct(b)), "Above(O(%d), O(%d))", a, b)
				assert.Equal(t, c == -1, oct(a).Below(oct(b)), "Below(O(%d), O(%d))", a, b)
				assert.Equal(t, a > b, oct(a).Above(oct(b)), "Above(O(%d), O(%d)) vs ASPN", a, b)
				assert.Equal(t, a < b, oct(a).Below(oct(b)), "Below(O(%d), O(%d)) vs ASPN", a, b)
			}
		}
	})
}

// K-9, C-9, G-8: inside a segment, ends included, in any order of the ends.
func TestBetween(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		o, a, b octave.Octave
		want    bool
	}{
		{name: "O(4) in [O(2), O(6)]", o: oct(4), a: oct(2), b: oct(6), want: true},
		{name: "O(2) in [O(2), O(6)]", o: oct(2), a: oct(2), b: oct(6), want: true},
		{name: "O(6) in [O(2), O(6)]", o: oct(6), a: oct(2), b: oct(6), want: true},
		{name: "O(6) in [O(6), O(2)]", o: oct(6), a: oct(6), b: oct(2), want: true},
		{name: "O(4) in [O(6), O(2)]", o: oct(4), a: oct(6), b: oct(2), want: true},
		{name: "O(4) in [O(4), O(4)]", o: oct(4), a: oct(4), b: oct(4), want: true},
		{name: "O(7) in [O(2), O(6)]", o: oct(7), a: oct(2), b: oct(6), want: false},
		{name: "O(1) in [O(2), O(6)]", o: oct(1), a: oct(2), b: oct(6), want: false},
		{name: "O(7) in [O(6), O(2)]", o: oct(7), a: oct(6), b: oct(2), want: false},
		{name: "O(5) in [O(4), O(4)]", o: oct(5), a: oct(4), b: oct(4), want: false},
		{name: "O(MIN) in [O(MIN), O(MAX)]", o: oct(minInt), a: oct(minInt), b: oct(maxInt), want: true},
		{name: "O(MAX) in [O(MIN), O(MAX)]", o: oct(maxInt), a: oct(minInt), b: oct(maxInt), want: true},
		{name: "O(MAX) in [O(MAX), O(MIN)]", o: oct(maxInt), a: oct(maxInt), b: oct(minInt), want: true},
		{name: "O(0) in [O(MIN), O(MAX)]", o: oct(0), a: oct(minInt), b: oct(maxInt), want: true},
		{name: "O(MIN) in [O(MIN+1), O(MAX)]", o: oct(minInt), a: oct(minInt + 1), b: oct(maxInt), want: false},
		{name: "O(4) in [O(3, FB), O(3, FB)]", o: oct(4), a: octave.MustFromNumber(3, fb()), b: octave.MustFromNumber(3, fb()), want: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var got bool
			require.NotPanics(t, func() { got = tc.o.Between(tc.a, tc.b) })
			assert.Equal(t, tc.want, got)
		})
	}

	t.Run("agrees with ASPN numbers", func(t *testing.T) {
		t.Parallel()

		for _, x := range span(-4, 4) {
			for _, a := range span(-4, 4) {
				for _, b := range span(-4, 4) {
					want := min(a, b) <= x && x <= max(a, b)
					assert.Equal(t, want, oct(x).Between(oct(a), oct(b)), "O(%d) in [O(%d), O(%d)]", x, a, b)
				}
			}
		}
	})
}

// K-18, C-18: every octave obtained without an error is valid, including the
// limits, octaves of own numberings and catalogs, and results of movement;
// the zero value is not.
func TestIsValid(t *testing.T) {
	t.Parallel()

	var valid []namedOctave
	for _, a := range append(span(-20, 20), edgeASPN()...) {
		valid = append(valid, namedOctave{name: fmt.Sprintf("O(%d)", a), o: oct(a)})
	}
	valid = append(valid, middleCWays(t)...)

	own := number.MustNew(4, 4, 0, 1)
	for _, x := range []int{minInt + 2, -1, 2, 4, maxInt} {
		o, err := octave.FromNumber(x, own)
		require.NoError(t, err)
		valid = append(valid, namedOctave{name: fmt.Sprintf("O(%d, New(4, 4, 0, 1))", x), o: o})
	}

	cat := octave.MustNewCatalog(map[octave.Octave]string{oct(minInt): "lo", oct(12): "high"})
	for _, s := range []string{"lo", "high"} {
		o, err := octave.FromName(s, cat)
		require.NoError(t, err)
		valid = append(valid, namedOctave{name: "P(" + s + ", own)", o: o})
	}

	mine := octave.MustNewNumbering(oct(4), 1)
	o, err := octave.FromNumber(0, mine)
	require.NoError(t, err)
	valid = append(valid, namedOctave{name: "O(0, NewNumbering(O(4), 1))", o: o})

	for _, step := range []struct {
		name string
		f    func() (octave.Octave, error)
	}{
		{name: "O(MAX).Shift(MIN)", f: func() (octave.Octave, error) { return oct(maxInt).Shift(minInt) }},
		{name: "O(-1).ShiftDown(MIN)", f: func() (octave.Octave, error) { return oct(-1).ShiftDown(minInt) }},
		{name: "O(MAX-1).Up()", f: func() (octave.Octave, error) { return oct(maxInt - 1).Up() }},
		{name: "O(MIN+1).Down()", f: func() (octave.Octave, error) { return oct(minInt + 1).Down() }},
	} {
		got, err := step.f()
		require.NoError(t, err, step.name)
		valid = append(valid, namedOctave{name: step.name, o: got})
	}

	seq, err := octave.Range(oct(maxInt-2), oct(maxInt))
	require.NoError(t, err)
	for _, x := range collect(seq, 10) {
		valid = append(valid, namedOctave{name: "element of Range(O(MAX-2), O(MAX))", o: x})
	}

	for _, v := range valid {
		assert.True(t, v.o.IsValid(), "%s.IsValid()", v.name)
	}

	var z octave.Octave
	assert.False(t, z.IsValid(), "var z octave.Octave")
	assert.False(t, octave.Octave{}.IsValid(), "octave.Octave{}")
}

// K-24, E-5: the invalid value in equality and order: not equal to anything,
// including itself, while == reports two invalid values equal; lower than
// every octave; two invalid values compare as 0; Above, Below and Between are
// false with an invalid value in any position.
func TestInvalidValueOrder(t *testing.T) {
	t.Parallel()

	var z1, z2 octave.Octave
	z3 := octave.Octave{}

	t.Run("Equal and ==", func(t *testing.T) {
		t.Parallel()

		assert.False(t, z1.Equal(z1), "Z.Equal(Z)")
		assert.False(t, z1.Equal(z2), "Z.Equal(other Z)")
		assert.True(t, z1 == z2, "Z == Z: the contract keeps == true")
		assert.True(t, z1 == z3, "var Z == Octave{}")
		for _, a := range []int{0, 4, minInt, maxInt} {
			assert.False(t, z1.Equal(oct(a)), "Z.Equal(O(%d))", a)
			assert.False(t, oct(a).Equal(z1), "O(%d).Equal(Z)", a)
			assert.False(t, z1 == oct(a), "Z == O(%d)", a)
		}
	})

	t.Run("Compare", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, 0, z1.Compare(z2), "Z.Compare(Z)")
		for _, a := range []int{minInt, 0, 4, maxInt} {
			assert.Equal(t, -1, z1.Compare(oct(a)), "Z.Compare(O(%d))", a)
			assert.Equal(t, 1, oct(a).Compare(z1), "O(%d).Compare(Z)", a)
		}

		xs := []octave.Octave{oct(4), z1, oct(minInt), z2, oct(0)}
		slices.SortFunc(xs, octave.Octave.Compare)
		assert.Equal(t, []octave.Octave{{}, {}, oct(minInt), oct(0), oct(4)}, xs, "sorted %s", describe(xs))
	})

	t.Run("Above and Below", func(t *testing.T) {
		t.Parallel()

		assert.False(t, z1.Above(z2), "Z.Above(Z)")
		assert.False(t, z1.Below(z2), "Z.Below(Z)")
		for _, a := range []int{minInt, 0, 4, maxInt} {
			assert.False(t, z1.Above(oct(a)), "Z.Above(O(%d))", a)
			assert.False(t, z1.Below(oct(a)), "Z.Below(O(%d))", a)
			assert.False(t, oct(a).Above(z1), "O(%d).Above(Z)", a)
			assert.False(t, oct(a).Below(z1), "O(%d).Below(Z)", a)
		}
	})

	t.Run("Between", func(t *testing.T) {
		t.Parallel()

		values := []struct {
			name string
			o    octave.Octave
		}{
			{name: "Z", o: z1},
			{name: "O(MIN)", o: oct(minInt)},
			{name: "O(4)", o: oct(4)},
			{name: "O(MAX)", o: oct(maxInt)},
		}
		for _, x := range values {
			for _, a := range values {
				for _, b := range values {
					if x.o.IsValid() && a.o.IsValid() && b.o.IsValid() {
						continue
					}
					assert.False(t, x.o.Between(a.o, b.o), "%s in [%s, %s]", x.name, a.name, b.name)
				}
			}
		}
	})
}
