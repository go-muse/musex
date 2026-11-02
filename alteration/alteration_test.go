package alteration_test

import (
	"math/big"
	"slices"
	"strconv"
	"testing"

	"github.com/go-muse/musex/alteration"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// K-1: the five ready values have exact values, are pairwise different and
// equal the values created from the integers 0, 1, -1, 2, -2.
func TestSigns(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		value alteration.Alteration
		n     int
	}{
		{name: "natural", value: alteration.Natural(), n: 0},
		{name: "sharp", value: alteration.Sharp(), n: 1},
		{name: "flat", value: alteration.Flat(), n: -1},
		{name: "double sharp", value: alteration.DoubleSharp(), n: 2},
		{name: "double flat", value: alteration.DoubleFlat(), n: -2},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, frac{tc.n, 1}, fractionOf(tc.value))

			fromInt, err := alteration.FromInt(tc.n)
			require.NoError(t, err)
			assert.True(t, tc.value.Equal(fromInt))
			assertSame(t, fromInt, tc.value)
		})
	}

	t.Run("pairwise different", func(t *testing.T) {
		t.Parallel()

		for i, a := range cases {
			for j, b := range cases {
				assert.Equal(t, i == j, a.value.Equal(b.value), "%s.Equal(%s)", a.name, b.name)
				assert.Equal(t, i == j, a.value == b.value, "%s == %s", a.name, b.name)
			}
		}
	})

	t.Run("natural is the zero value", func(t *testing.T) {
		t.Parallel()

		var zero alteration.Alteration

		assertSame(t, zero, alteration.Natural())
		assert.True(t, alteration.Natural().Equal(zero))
	})
}

// K-2: values from the integers -10...10; +3 raises and -4 lowers.
func TestFromInt(t *testing.T) {
	t.Parallel()

	t.Run("named", func(t *testing.T) {
		t.Parallel()

		three, err := alteration.FromInt(3)
		require.NoError(t, err)
		assert.Equal(t, frac{3, 1}, fractionOf(three))
		assert.Equal(t, 1, three.Sign(), "+3 raises")
		assert.True(t, three.IsInteger())

		minusFour, err := alteration.FromInt(-4)
		require.NoError(t, err)
		assert.Equal(t, frac{-4, 1}, fractionOf(minusFour))
		assert.Equal(t, -1, minusFour.Sign(), "-4 lowers")
		assert.True(t, minusFour.IsInteger())
	})

	for n := -10; n <= 10; n++ {
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			t.Parallel()

			a, err := alteration.FromInt(n)
			require.NoError(t, err)
			assert.Equal(t, frac{n, 1}, fractionOf(a))
			assert.True(t, a.IsInteger())
			assert.Equal(t, sign(n), a.Sign())
		})
	}
}

// K-3: values from fractions are normalized: the sign goes to the numerator,
// the fraction is reduced; equal numbers give equal values.
func TestFromFraction(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   frac
		want frac
	}{
		{name: "one half", in: frac{1, 2}, want: frac{1, 2}},
		{name: "two quarters", in: frac{2, 4}, want: frac{1, 2}},
		{name: "negative denominator", in: frac{1, -2}, want: frac{-1, 2}},
		{name: "negative numerator", in: frac{-1, 2}, want: frac{-1, 2}},
		{name: "both negative", in: frac{-2, -4}, want: frac{1, 2}},
		{name: "zero over five", in: frac{0, 5}, want: frac{0, 1}},
		{name: "zero over minus three", in: frac{0, -3}, want: frac{0, 1}},
		{name: "four halves", in: frac{4, 2}, want: frac{2, 1}},
		{name: "six quarters", in: frac{6, 4}, want: frac{3, 2}},
		{name: "one third", in: frac{1, 3}, want: frac{1, 3}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			a, err := alteration.FromFraction(tc.in.num, tc.in.den)
			require.NoError(t, err)
			assert.Equal(t, tc.want, fractionOf(a))
		})
	}

	t.Run("equal numbers give equal values", func(t *testing.T) {
		t.Parallel()

		half := value(t, frac{1, 2})
		twoQuarters := value(t, frac{2, 4})
		assert.True(t, half.Equal(twoQuarters))
		assertSame(t, twoQuarters, half)

		assertSame(t, alteration.DoubleSharp(), value(t, frac{4, 2}))
		assertSame(t, alteration.Natural(), value(t, frac{0, 5}))
		assertSame(t, alteration.Natural(), value(t, frac{0, -3}))
	})

	t.Run("all p and q in -12...12", func(t *testing.T) {
		t.Parallel()

		// Values grouped by the exact number: every group holds one value,
		// different groups hold different values.
		groups := make(map[string][]alteration.Alteration)

		for p := -12; p <= 12; p++ {
			for q := -12; q <= 12; q++ {
				if q == 0 {
					continue
				}

				a, err := alteration.FromFraction(p, q)
				require.NoError(t, err, "FromFraction(%d, %d)", p, q)

				num, den := a.Fraction()
				assert.Equal(t, expectedFrac(t, ratOf(p, q)), frac{num, den}, "FromFraction(%d, %d)", p, q)
				assert.Positive(t, den, "FromFraction(%d, %d): positive denominator", p, q)
				assert.Equal(t, 1, gcd(num, den), "FromFraction(%d, %d): lowest terms", p, q)

				key := ratOf(p, q).RatString()
				groups[key] = append(groups[key], a)
			}
		}

		representatives := make(map[alteration.Alteration]string, len(groups))
		for key, group := range groups {
			for _, a := range group[1:] {
				assertSame(t, group[0], a, "%s: equal numbers give equal values", key)
				assert.True(t, a.Equal(group[0]), "%s: Equal", key)
			}
			if prev, ok := representatives[group[0]]; ok {
				assert.Failf(t, "different numbers give equal values", "%s and %s", prev, key)
			}
			representatives[group[0]] = key
		}
	})
}

// K-4: raising and lowering by one chromatic unit.
func TestRaiseLower(t *testing.T) {
	t.Parallel()

	t.Run("raise chain", func(t *testing.T) {
		t.Parallel()

		a := alteration.Flat()
		for _, want := range []frac{{0, 1}, {1, 1}, {2, 1}, {3, 1}} {
			var err error
			a, err = a.Raise()
			require.NoError(t, err)
			assert.Equal(t, want, fractionOf(a))
		}
	})

	t.Run("lower chain", func(t *testing.T) {
		t.Parallel()

		a := alteration.Sharp()
		for _, want := range []frac{{0, 1}, {-1, 1}, {-2, 1}} {
			var err error
			a, err = a.Lower()
			require.NoError(t, err)
			assert.Equal(t, want, fractionOf(a))
		}
	})

	t.Run("fractions", func(t *testing.T) {
		t.Parallel()

		up, err := value(t, frac{1, 2}).Raise()
		require.NoError(t, err)
		assert.Equal(t, frac{3, 2}, fractionOf(up))

		up, err = value(t, frac{-1, 2}).Raise()
		require.NoError(t, err)
		assert.Equal(t, frac{1, 2}, fractionOf(up))

		down, err := value(t, frac{1, 2}).Lower()
		require.NoError(t, err)
		assert.Equal(t, frac{-1, 2}, fractionOf(down))
	})

	for _, f := range setS() {
		t.Run(f.String(), func(t *testing.T) {
			t.Parallel()

			a := value(t, f)

			up, err := a.Raise()
			require.NoError(t, err)
			assertExact(t, ratAdd(f.rat(), ratInt(1)), up)

			down, err := a.Lower()
			require.NoError(t, err)
			assertExact(t, ratSub(f.rat(), ratInt(1)), down)

			back, err := up.Lower()
			require.NoError(t, err)
			assertSame(t, a, back, "Lower(Raise(a)) = a")

			back, err = down.Raise()
			require.NoError(t, err)
			assertSame(t, a, back, "Raise(Lower(a)) = a")

			byOne, err := a.RaiseBy(1)
			require.NoError(t, err)
			assertSame(t, byOne, up, "Raise() = RaiseBy(1)")

			bySharp, err := a.Add(alteration.Sharp())
			require.NoError(t, err)
			assertSame(t, bySharp, up, "Raise() = Add(Sharp())")

			byOne, err = a.LowerBy(1)
			require.NoError(t, err)
			assertSame(t, byOne, down, "Lower() = LowerBy(1)")

			bySharp, err = a.Sub(alteration.Sharp())
			require.NoError(t, err)
			assertSame(t, bySharp, down, "Lower() = Sub(Sharp())")
		})
	}
}

// K-5: raising and lowering by an integer n.
func TestRaiseByLowerBy(t *testing.T) {
	t.Parallel()

	examples := []struct {
		name string
		a    frac
		up   bool
		n    int
		want frac
	}{
		{name: "natural up 3", a: frac{0, 1}, up: true, n: 3, want: frac{3, 1}},
		{name: "sharp up -2", a: frac{1, 1}, up: true, n: -2, want: frac{-1, 1}},
		{name: "double sharp down 4", a: frac{2, 1}, n: 4, want: frac{-2, 1}},
		{name: "one half up 2", a: frac{1, 2}, up: true, n: 2, want: frac{5, 2}},
	}

	for _, tc := range examples {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			a := value(t, tc.a)

			var (
				got alteration.Alteration
				err error
			)
			if tc.up {
				got, err = a.RaiseBy(tc.n)
			} else {
				got, err = a.LowerBy(tc.n)
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, fractionOf(got))
		})
	}

	for _, f := range setS() {
		t.Run(f.String(), func(t *testing.T) {
			t.Parallel()

			a := value(t, f)

			for n := -5; n <= 5; n++ {
				up, err := a.RaiseBy(n)
				require.NoError(t, err, "RaiseBy(%d)", n)
				assertExact(t, ratAdd(f.rat(), ratInt(n)), up, "RaiseBy(%d)", n)

				down, err := a.LowerBy(n)
				require.NoError(t, err, "LowerBy(%d)", n)
				assertExact(t, ratSub(f.rat(), ratInt(n)), down, "LowerBy(%d)", n)

				opposite, err := a.RaiseBy(-n)
				require.NoError(t, err, "RaiseBy(%d)", -n)
				assertSame(t, opposite, down, "LowerBy(%d) = RaiseBy(%d)", n, -n)

				back, err := up.LowerBy(n)
				require.NoError(t, err, "LowerBy(%d)", n)
				assertSame(t, a, back, "LowerBy(RaiseBy(a, %d), %d) = a", n, n)
			}

			same, err := a.RaiseBy(0)
			require.NoError(t, err)
			assertSame(t, a, same, "RaiseBy(0) = a")

			raised, err := a.Raise()
			require.NoError(t, err)
			byOne, err := a.RaiseBy(1)
			require.NoError(t, err)
			assertSame(t, byOne, raised, "Raise() = RaiseBy(1)")
		})
	}
}

// K-6: raising and lowering by a value of alteration, fractional or whole.
func TestAddSub(t *testing.T) {
	t.Parallel()

	examples := []struct {
		name string
		got  func(t *testing.T) (alteration.Alteration, error)
		want frac
	}{
		{
			name: "natural up 1/2",
			got: func(t *testing.T) (alteration.Alteration, error) {
				return alteration.Natural().Add(value(t, frac{1, 2}))
			},
			want: frac{1, 2},
		},
		{
			name: "1/2 up 1/2",
			got: func(t *testing.T) (alteration.Alteration, error) {
				return value(t, frac{1, 2}).Add(value(t, frac{1, 2}))
			},
			want: frac{1, 1},
		},
		{
			name: "1/2 up 1/3",
			got: func(t *testing.T) (alteration.Alteration, error) {
				return value(t, frac{1, 2}).Add(value(t, frac{1, 3}))
			},
			want: frac{5, 6},
		},
		{
			name: "sharp down 1/2",
			got: func(t *testing.T) (alteration.Alteration, error) {
				return alteration.Sharp().Sub(value(t, frac{1, 2}))
			},
			want: frac{1, 2},
		},
		{
			name: "flat up 3/2",
			got: func(t *testing.T) (alteration.Alteration, error) {
				return alteration.Flat().Add(value(t, frac{3, 2}))
			},
			want: frac{1, 2},
		},
	}

	for _, tc := range examples {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.got(t)
			require.NoError(t, err)
			assert.Equal(t, tc.want, fractionOf(got))
		})
	}

	t.Run("1/2 up 1/2 is a sharp and whole", func(t *testing.T) {
		t.Parallel()

		half := value(t, frac{1, 2})
		got, err := half.Add(half)
		require.NoError(t, err)
		assertSame(t, alteration.Sharp(), got)
		assert.True(t, got.IsInteger())
	})

	for _, fa := range setS() {
		t.Run(fa.String(), func(t *testing.T) {
			t.Parallel()

			a := value(t, fa)

			same, err := a.Add(alteration.Natural())
			require.NoError(t, err)
			assertSame(t, a, same, "Add(Natural()) = a")

			for _, fd := range setS() {
				d := value(t, fd)

				sum, err := a.Add(d)
				require.NoError(t, err, "Add%s", fd)
				assertExact(t, ratAdd(fa.rat(), fd.rat()), sum, "Add%s", fd)

				diff, err := a.Sub(d)
				require.NoError(t, err, "Sub%s", fd)
				assertExact(t, ratSub(fa.rat(), fd.rat()), diff, "Sub%s", fd)

				back, err := sum.Sub(d)
				require.NoError(t, err, "Sub%s", fd)
				assertSame(t, a, back, "Sub(Add(a, %s), %s) = a", fd, fd)

				if fd.rat().IsInt() {
					n := int(fd.rat().Num().Int64())

					byN, err := a.RaiseBy(n)
					require.NoError(t, err, "RaiseBy(%d)", n)
					assertSame(t, byN, sum, "Add%s = RaiseBy(%d)", fd, n)

					byN, err = a.LowerBy(n)
					require.NoError(t, err, "LowerBy(%d)", n)
					assertSame(t, byN, diff, "Sub%s = LowerBy(%d)", fd, n)
				}
			}
		})
	}
}

// K-7: equality does not depend on how the value was obtained.
func TestEqual(t *testing.T) {
	t.Parallel()

	t.Run("one half three ways", func(t *testing.T) {
		t.Parallel()

		half := value(t, frac{1, 2})
		quarters := value(t, frac{2, 4})
		raised, err := alteration.Natural().Add(half)
		require.NoError(t, err)

		ways := []alteration.Alteration{half, quarters, raised}
		for i, a := range ways {
			for j, b := range ways {
				assert.True(t, a.Equal(b), "way %d equals way %d", i, j)
				assertSame(t, b, a, "way %d == way %d", i, j)
			}
		}
	})

	t.Run("sharp four ways", func(t *testing.T) {
		t.Parallel()

		fromInt, err := alteration.FromInt(1)
		require.NoError(t, err)
		raised, err := alteration.Natural().RaiseBy(1)
		require.NoError(t, err)

		ways := []alteration.Alteration{alteration.Sharp(), fromInt, value(t, frac{2, 2}), raised}
		for i, a := range ways {
			for j, b := range ways {
				assert.True(t, a.Equal(b), "way %d equals way %d", i, j)
				assertSame(t, b, a, "way %d == way %d", i, j)
			}
		}
	})

	t.Run("different values", func(t *testing.T) {
		t.Parallel()

		assert.False(t, value(t, frac{1, 2}).Equal(value(t, frac{1, 3})))
		assert.False(t, alteration.Natural().Equal(alteration.Sharp()))
		assert.False(t, alteration.Sharp().Equal(alteration.Natural()))
	})

	t.Run("one map key", func(t *testing.T) {
		t.Parallel()

		raised, err := alteration.Natural().Add(value(t, frac{1, 2}))
		require.NoError(t, err)

		keys := map[alteration.Alteration]string{value(t, frac{1, 2}): "half"}
		keys[value(t, frac{2, 4})] = "two quarters"
		keys[raised] = "raised"

		assert.Len(t, keys, 1)
		assert.Equal(t, "raised", keys[value(t, frac{-3, -6})])
	})

	t.Run("pairs of S", func(t *testing.T) {
		t.Parallel()

		for _, fa := range setS() {
			a := value(t, fa)
			assert.True(t, a.Equal(a), "%s reflexive", fa)

			for _, fb := range setS() {
				b := value(t, fb)
				want := fa.rat().Cmp(fb.rat()) == 0

				assert.Equal(t, want, a.Equal(b), "%s.Equal(%s)", fa, fb)
				assert.Equal(t, a.Equal(b), b.Equal(a), "symmetric for %s and %s", fa, fb)
				assert.Equal(t, a.Equal(b), a == b, "Equal and == agree for %s and %s", fa, fb)

				diff, err := a.Sub(b)
				require.NoError(t, err)
				assert.Equal(t, a.Equal(b), diff.IsNatural(), "equal iff the difference is natural: %s, %s", fa, fb)
			}
		}
	})
}

// K-8: the difference of two values (Sub).
func TestDifference(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		a, b frac
		want frac
	}{
		{name: "sharp minus flat", a: frac{1, 1}, b: frac{-1, 1}, want: frac{2, 1}},
		{name: "flat minus sharp", a: frac{-1, 1}, b: frac{1, 1}, want: frac{-2, 1}},
		{name: "1/2 minus natural", a: frac{1, 2}, b: frac{0, 1}, want: frac{1, 2}},
		{name: "1/2 minus 1/3", a: frac{1, 2}, b: frac{1, 3}, want: frac{1, 6}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := value(t, tc.a).Sub(value(t, tc.b))
			require.NoError(t, err)
			assert.Equal(t, tc.want, fractionOf(got))
		})
	}

	t.Run("sharp minus flat is a double sharp", func(t *testing.T) {
		t.Parallel()

		got, err := alteration.Sharp().Sub(alteration.Flat())
		require.NoError(t, err)
		assertSame(t, alteration.DoubleSharp(), got)
	})

	for _, fa := range setS() {
		t.Run(fa.String(), func(t *testing.T) {
			t.Parallel()

			a := value(t, fa)

			self, err := a.Sub(value(t, fa))
			require.NoError(t, err)
			assert.True(t, self.IsNatural(), "a - a is natural")
			assertSame(t, alteration.Natural(), self, "a - a is natural")

			for _, fb := range setS() {
				b := value(t, fb)

				ab, err := a.Sub(b)
				require.NoError(t, err)
				ba, err := b.Sub(a)
				require.NoError(t, err)

				negated, err := alteration.Natural().Sub(ba)
				require.NoError(t, err)
				assertSame(t, negated, ab, "a - b = 0 - (b - a) for %s, %s", fa, fb)

				restored, err := b.Add(ab)
				require.NoError(t, err)
				assertSame(t, a, restored, "b + (a - b) = a for %s, %s", fa, fb)
			}
		})
	}
}

// K-9: the order by magnitude.
func TestCompare(t *testing.T) {
	t.Parallel()

	t.Run("strictly ascending S", func(t *testing.T) {
		t.Parallel()

		ordered := values(t, setS())
		for i, a := range ordered {
			for j, b := range ordered {
				assert.Equal(t, sign(i-j), a.Compare(b), "compare index %d with %d", i, j)
			}
		}
	})

	t.Run("pairs of S", func(t *testing.T) {
		t.Parallel()

		for _, fa := range setS() {
			a := value(t, fa)

			for _, fb := range setS() {
				b := value(t, fb)
				got := a.Compare(b)

				assert.Contains(t, []int{-1, 0, 1}, got, "exactly one outcome for %s, %s", fa, fb)
				assert.Equal(t, fa.rat().Cmp(fb.rat()), got, "%s.Compare(%s)", fa, fb)
				assert.Equal(t, got == 0, a.Equal(b), "equal outcome iff Equal for %s, %s", fa, fb)
				assert.Equal(t, -got, b.Compare(a), "antisymmetric for %s, %s", fa, fb)

				diff, err := a.Sub(b)
				require.NoError(t, err)
				assert.Equal(t, diff.Sign(), got, "sign of the difference for %s, %s", fa, fb)
			}
		}
	})

	t.Run("sort", func(t *testing.T) {
		t.Parallel()

		want := values(t, setS())

		// A fixed permutation of S: no randomness in tests.
		shuffled := make([]alteration.Alteration, 0, len(want))
		for _, i := range []int{6, 0, 9, 3, 10, 1, 8, 4, 2, 7, 5} {
			shuffled = append(shuffled, want[i])
		}
		require.Len(t, shuffled, len(want))

		slices.SortFunc(shuffled, alteration.Alteration.Compare)
		assert.Equal(t, fractions(want), fractions(shuffled))
	})
}

// fractions returns the exact values of a slice of values.
func fractions(as []alteration.Alteration) []frac {
	out := make([]frac, len(as))
	for i, a := range as {
		out[i] = fractionOf(a)
	}

	return out
}

// K-10: natural, whole and direction.
func TestPredicates(t *testing.T) {
	t.Parallel()

	halfDownHalf := func(t *testing.T) alteration.Alteration {
		half := value(t, frac{1, 2})
		got, err := half.Sub(half)
		require.NoError(t, err)

		return got
	}

	cases := []struct {
		name      string
		value     func(t *testing.T) alteration.Alteration
		isNatural bool
		isInteger bool
		sign      int
	}{
		{name: "natural", value: func(*testing.T) alteration.Alteration { return alteration.Natural() }, isNatural: true, isInteger: true, sign: 0},
		{name: "0/7", value: func(t *testing.T) alteration.Alteration { return value(t, frac{0, 7}) }, isNatural: true, isInteger: true, sign: 0},
		{name: "1/2 down 1/2", value: halfDownHalf, isNatural: true, isInteger: true, sign: 0},
		{name: "sharp", value: func(*testing.T) alteration.Alteration { return alteration.Sharp() }, isInteger: true, sign: 1},
		{name: "flat", value: func(*testing.T) alteration.Alteration { return alteration.Flat() }, isInteger: true, sign: -1},
		{name: "double sharp", value: func(*testing.T) alteration.Alteration { return alteration.DoubleSharp() }, isInteger: true, sign: 1},
		{name: "double flat", value: func(*testing.T) alteration.Alteration { return alteration.DoubleFlat() }, isInteger: true, sign: -1},
		{name: "+3", value: func(t *testing.T) alteration.Alteration { return value(t, frac{3, 1}) }, isInteger: true, sign: 1},
		{name: "4/2", value: func(t *testing.T) alteration.Alteration { return value(t, frac{4, 2}) }, isInteger: true, sign: 1},
		{name: "1/2", value: func(t *testing.T) alteration.Alteration { return value(t, frac{1, 2}) }, sign: 1},
		{name: "-1/2", value: func(t *testing.T) alteration.Alteration { return value(t, frac{-1, 2}) }, sign: -1},
		{name: "1/3", value: func(t *testing.T) alteration.Alteration { return value(t, frac{1, 3}) }, sign: 1},
		{name: "3/2", value: func(t *testing.T) alteration.Alteration { return value(t, frac{3, 2}) }, sign: 1},
		{name: "-7/2", value: func(t *testing.T) alteration.Alteration { return value(t, frac{-7, 2}) }, sign: -1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			a := tc.value(t)
			assert.Equal(t, tc.isNatural, a.IsNatural(), "IsNatural")
			assert.Equal(t, tc.isInteger, a.IsInteger(), "IsInteger")
			assert.Equal(t, tc.sign, a.Sign(), "Sign")
			assert.Equal(t, a.IsNatural(), a.Equal(alteration.Natural()), "IsNatural iff Equal(Natural())")
			assert.Equal(t, a.IsNatural(), a.Sign() == 0, "IsNatural iff Sign() == 0")
		})
	}

	t.Run("S", func(t *testing.T) {
		t.Parallel()

		for _, f := range setS() {
			a := value(t, f)
			r := f.rat()

			assert.Equal(t, r.Sign() == 0, a.IsNatural(), "%s.IsNatural()", f)
			assert.Equal(t, r.IsInt(), a.IsInteger(), "%s.IsInteger()", f)
			assert.Equal(t, r.Sign(), a.Sign(), "%s.Sign()", f)
		}
	})
}

// K-11: the exact value as a fraction in lowest terms.
func TestFraction(t *testing.T) {
	t.Parallel()

	for _, f := range setS() {
		t.Run(f.String(), func(t *testing.T) {
			t.Parallel()

			a := value(t, f)
			num, den := a.Fraction()

			assert.Positive(t, den, "positive denominator")
			assert.Equal(t, 1, gcd(num, den), "lowest terms")
			assert.Equal(t, expectedFrac(t, f.rat()), frac{num, den})

			back, err := alteration.FromFraction(a.Fraction())
			require.NoError(t, err)
			assertSame(t, a, back, "FromFraction(a.Fraction()) = a")
			assertSame(t, a, alteration.MustFromFraction(a.Fraction()), "MustFromFraction(a.Fraction()) = a")
		})
	}

	t.Run("natural and integers", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, frac{0, 1}, fractionOf(alteration.Natural()))

		for n := -10; n <= 10; n++ {
			a, err := alteration.FromInt(n)
			require.NoError(t, err)
			assert.Equal(t, frac{n, 1}, fractionOf(a), "FromInt(%d)", n)
		}
	})
}

// K-12: multiplication by an integer.
func TestMul(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		a    frac
		k    int
		want frac
	}{
		{name: "1/2 times 2", a: frac{1, 2}, k: 2, want: frac{1, 1}},
		{name: "-3/2 times 4", a: frac{-3, 2}, k: 4, want: frac{-6, 1}},
		{name: "1/3 times 3", a: frac{1, 3}, k: 3, want: frac{1, 1}},
		{name: "1/2 times 3", a: frac{1, 2}, k: 3, want: frac{3, 2}},
		{name: "double flat times -1", a: frac{-2, 1}, k: -1, want: frac{2, 1}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := value(t, tc.a).Mul(tc.k)
			require.NoError(t, err)
			assert.Equal(t, tc.want, fractionOf(got))
		})
	}

	t.Run("sharps", func(t *testing.T) {
		t.Parallel()

		got, err := value(t, frac{1, 2}).Mul(2)
		require.NoError(t, err)
		assertSame(t, alteration.Sharp(), got)

		got, err = alteration.DoubleFlat().Mul(-1)
		require.NoError(t, err)
		assertSame(t, alteration.DoubleSharp(), got)
	})

	for _, f := range setS() {
		t.Run(f.String(), func(t *testing.T) {
			t.Parallel()

			a := value(t, f)

			for k := -4; k <= 4; k++ {
				ak, err := a.Mul(k)
				require.NoError(t, err, "Mul(%d)", k)
				assertExact(t, ratMul(f.rat(), k), ak, "Mul(%d)", k)

				for m := -3; m <= 3; m++ {
					left, err := ak.Mul(m)
					require.NoError(t, err, "Mul(%d).Mul(%d)", k, m)
					right, err := a.Mul(k * m)
					require.NoError(t, err, "Mul(%d)", k*m)
					assertSame(t, right, left, "(a * %d) * %d = a * %d", k, m, k*m)
				}
			}

			zero, err := a.Mul(0)
			require.NoError(t, err)
			assertSame(t, alteration.Natural(), zero, "a * 0 is natural")

			one, err := a.Mul(1)
			require.NoError(t, err)
			assertSame(t, a, one, "a * 1 = a")

			minusOne, err := a.Mul(-1)
			require.NoError(t, err)
			negated, err := alteration.Natural().Sub(a)
			require.NoError(t, err)
			assertSame(t, negated, minusOne, "a * -1 = 0 - a")
		})
	}
}

// K-16: values with many signs are ordinary values without errors.
func TestLargeValues(t *testing.T) {
	t.Parallel()

	large := []frac{{3, 1}, {-7, 2}, {100, 1}, {-100, 1}}

	for _, f := range large {
		t.Run(f.String(), func(t *testing.T) {
			t.Parallel()

			a := value(t, f)
			assert.Equal(t, expectedFrac(t, f.rat()), fractionOf(a))

			up, err := a.Raise()
			require.NoError(t, err)
			assertExact(t, ratAdd(f.rat(), ratInt(1)), up, "Raise")

			down, err := a.Lower()
			require.NoError(t, err)
			assertExact(t, ratSub(f.rat(), ratInt(1)), down, "Lower")

			up, err = a.RaiseBy(5)
			require.NoError(t, err)
			assertExact(t, ratAdd(f.rat(), ratInt(5)), up, "RaiseBy(5)")

			down, err = a.LowerBy(5)
			require.NoError(t, err)
			assertExact(t, ratSub(f.rat(), ratInt(5)), down, "LowerBy(5)")

			for _, g := range slices.Concat(large, setS()) {
				b := value(t, g)
				assert.Equal(t, f.rat().Cmp(g.rat()), a.Compare(b), "%s.Compare(%s)", f, g)
				assert.Equal(t, f.rat().Cmp(g.rat()) == 0, a.Equal(b), "%s.Equal(%s)", f, g)
			}
		})
	}

	t.Run("integers", func(t *testing.T) {
		t.Parallel()

		hundred, err := alteration.FromInt(100)
		require.NoError(t, err)
		minusHundred, err := alteration.FromInt(-100)
		require.NoError(t, err)

		up, err := hundred.Raise()
		require.NoError(t, err)
		assert.Equal(t, frac{101, 1}, fractionOf(up), "+100 up 1 = +101")

		down, err := minusHundred.Lower()
		require.NoError(t, err)
		assert.Equal(t, frac{-101, 1}, fractionOf(down), "-100 down 1 = -101")
	})
}

// K-20: operations without errors neither fail nor panic on S; operations
// with an error return nil on S and on the sets of K-4...K-12.
func TestOperationsTotal(t *testing.T) {
	t.Parallel()

	for _, fa := range setS() {
		t.Run(fa.String(), func(t *testing.T) {
			t.Parallel()

			a := value(t, fa)

			require.NotPanics(t, func() {
				_ = alteration.Natural()
				_ = alteration.Sharp()
				_ = alteration.Flat()
				_ = alteration.DoubleSharp()
				_ = alteration.DoubleFlat()
				_, _ = a.Fraction()
				_ = a.Sign()
				_ = a.IsNatural()
				_ = a.IsInteger()
				_ = a.String()
			})

			var err error

			require.NotPanics(t, func() {
				_, err = a.Raise()
			})
			require.NoError(t, err, "Raise")

			require.NotPanics(t, func() {
				_, err = a.Lower()
			})
			require.NoError(t, err, "Lower")

			for n := -5; n <= 5; n++ {
				require.NotPanics(t, func() { _, err = a.RaiseBy(n) })
				require.NoError(t, err, "RaiseBy(%d)", n)
				require.NotPanics(t, func() { _, err = a.LowerBy(n) })
				require.NoError(t, err, "LowerBy(%d)", n)
			}

			for k := -12; k <= 12; k++ {
				require.NotPanics(t, func() { _, err = a.Mul(k) })
				require.NoError(t, err, "Mul(%d)", k)
			}

			for _, fb := range setS() {
				b := value(t, fb)

				require.NotPanics(t, func() {
					_ = a.Equal(b)
					_ = a.Compare(b)
				})
				require.NotPanics(t, func() { _, err = a.Add(b) })
				require.NoError(t, err, "Add%s", fb)
				require.NotPanics(t, func() { _, err = a.Sub(b) })
				require.NoError(t, err, "Sub%s", fb)
			}
		})
	}

	t.Run("constructors", func(t *testing.T) {
		t.Parallel()

		var err error

		for n := -10; n <= 10; n++ {
			require.NotPanics(t, func() { _, err = alteration.FromInt(n) })
			require.NoError(t, err, "FromInt(%d)", n)
		}

		for p := -12; p <= 12; p++ {
			for q := -12; q <= 12; q++ {
				if q == 0 {
					continue
				}
				require.NotPanics(t, func() { _, err = alteration.FromFraction(p, q) })
				require.NoError(t, err, "FromFraction(%d, %d)", p, q)
			}
		}
	})
}

// The exact values used by the tests come from math/big, not from the
// package; this guards the oracle helper against an int64 overflow in the
// largest K-12 product (|(-7/2) * 4 * 3| = 42).
func TestOracleSelfCheck(t *testing.T) {
	t.Parallel()

	assert.Equal(t, frac{-42, 1}, expectedFrac(t, ratMul(ratMul(ratOf(-7, 2), 4), 3)))
	assert.True(t, inRange(ratInt(oracleMaxMagnitude)))
	assert.False(t, inRange(ratInt(oracleMaxMagnitude+1)))
	assert.True(t, inRange(ratOf(1, oracleMaxDenominator)))
	assert.False(t, inRange(ratOf(1, oracleMaxDenominator+1)))
	assert.Equal(t, "+1/2", contractString(big.NewRat(1, 2)))
	assert.Equal(t, "-7/2", contractString(big.NewRat(-7, 2)))
	assert.Equal(t, "0", contractString(new(big.Rat)))
	assert.Equal(t, "+3", contractString(big.NewRat(3, 1)))
}
