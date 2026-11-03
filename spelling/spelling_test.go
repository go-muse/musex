package spelling_test

import (
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// K-1: a name is created from any letter of the seven and any value, and
// returns exactly its parts.
func TestNew(t *testing.T) {
	t.Parallel()

	var cases []nameCase
	for _, l := range letter.All() {
		for _, a := range setS() {
			cases = append(cases, nc(l, a.a))
		}
	}

	cases = append(cases,
		nc(letter.C, whole(maxMagnitude)),
		nc(letter.C, whole(-maxMagnitude)),
		nc(letter.C, frac(1, maxDenominator)),
		nc(letter.C, frac(-1, maxDenominator)),
	)

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := spelling.New(tc.l, tc.a)
			require.NoError(t, err)
			assertName(t, tc.l, tc.a, got)
		})
	}
}

// K-2: the 35 ready-made names equal the names created from their parts and
// differ from each other.
func TestReadyMadeNames(t *testing.T) {
	t.Parallel()

	cases := predefinedCases()
	require.Len(t, cases, 35)

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.fn()
			want := spelling.MustNew(tc.l, tc.a)

			assertName(t, tc.l, tc.a, got)
			assert.True(t, got.Equal(want), "Equal to the name created from the parts")
			assert.Equal(t, want, got, "== to the name created from the parts")
		})
	}

	t.Run("pairwise different", func(t *testing.T) {
		t.Parallel()

		for i, x := range cases {
			for j, y := range cases {
				if i == j {
					continue
				}

				assert.Falsef(t, x.fn().Equal(y.fn()), "%s must differ from %s", x.name, y.name)
			}
		}
	})
}

// K-3: names are equal only by spelling: letter and value.
func TestEqual(t *testing.T) {
	t.Parallel()

	t.Run("different names", func(t *testing.T) {
		t.Parallel()

		pairs := []struct {
			name string
			x, y spelling.Spelling
		}{
			{"C sharp and D flat", spelling.CSharp(), spelling.DFlat()},
			{"B sharp and C", spelling.BSharp(), spelling.C()},
			{"E sharp and F", spelling.ESharp(), spelling.F()},
			{"C flat and B", spelling.CFlat(), spelling.B()},
			{"C double sharp and D", spelling.CDoubleSharp(), spelling.D()},
			{"F double flat and E flat", spelling.FDoubleFlat(), spelling.EFlat()},
			{"C+1/2 and C+1/3", spelling.MustNew(letter.C, frac(1, 2)), spelling.MustNew(letter.C, frac(1, 3))},
			{"C and D", spelling.C(), spelling.D()},
		}

		for _, tc := range pairs {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				assert.False(t, tc.x.Equal(tc.y))
				assert.False(t, tc.y.Equal(tc.x))
				assert.NotEqual(t, tc.x, tc.y, "==")
			})
		}
	})

	t.Run("one name obtained in different ways", func(t *testing.T) {
		t.Parallel()

		ways := []struct {
			name string
			s    spelling.Spelling
		}{
			{"ready-made C sharp", spelling.CSharp()},
			{"created (C, +1)", spelling.MustNew(letter.C, alteration.Sharp())},
			{"created (C, 2/2)", spelling.MustNew(letter.C, frac(2, 2))},
			{"C raised by 1", spelling.C().MustRaise()},
			{"C double sharp lowered by 1", spelling.CDoubleSharp().MustLower()},
			{"C flat with the sharp", spelling.CFlat().WithAlteration(alteration.Sharp())},
		}

		for _, x := range ways {
			for _, y := range ways {
				assert.Truef(t, x.s.Equal(y.s), "%s must equal %s", x.name, y.name)
				assert.Equalf(t, x.s, y.s, "%s must be == %s", x.name, y.name)
			}
		}
	})

	t.Run("2/4 and 1/2", func(t *testing.T) {
		t.Parallel()

		x := spelling.MustNew(letter.C, frac(2, 4))
		y := spelling.MustNew(letter.C, frac(1, 2))

		assert.True(t, x.Equal(y))
		assert.Equal(t, x, y)
	})

	t.Run("properties on N", func(t *testing.T) {
		t.Parallel()

		for _, x := range setN() {
			for _, y := range setN() {
				xs, ys := x.build(), y.build()
				want := x.l == y.l && x.a.Equal(y.a)

				assert.Equalf(t, want, xs.Equal(ys), "%s.Equal(%s)", x.name, y.name)
				assert.Equalf(t, xs.Equal(ys), ys.Equal(xs), "symmetry for %s and %s", x.name, y.name)
				assert.Equalf(t, xs.Equal(ys), xs == ys, "Equal and == for %s and %s", x.name, y.name)
			}

			xs := x.build()
			assert.Truef(t, xs.Equal(x.build()), "reflexivity for %s", x.name)
		}
	})

	t.Run("map key", func(t *testing.T) {
		t.Parallel()

		m := map[spelling.Spelling]int{}
		m[spelling.CSharp()]++
		m[spelling.MustNew(letter.C, alteration.Sharp())]++
		m[spelling.C().MustRaise()]++
		m[spelling.DFlat()]++

		assert.Len(t, m, 2)
		assert.Equal(t, 3, m[spelling.CSharp()])
		assert.Equal(t, 1, m[spelling.DFlat()])
	})
}

// K-4: raising and lowering by one keep the letter and change the value by one.
func TestRaiseLower(t *testing.T) {
	t.Parallel()

	t.Run("chains", func(t *testing.T) {
		t.Parallel()

		up := []struct {
			name string
			from spelling.Spelling
			want nameCase
		}{
			{"C up", spelling.C(), nc(letter.C, alteration.Sharp())},
			{"C sharp up", spelling.CSharp(), nc(letter.C, alteration.DoubleSharp())},
			{"C double sharp up", spelling.CDoubleSharp(), nc(letter.C, whole(3))},
			{"D flat up", spelling.DFlat(), nc(letter.D, alteration.Natural())},
			{"A+1/2 up", spelling.MustNew(letter.A, frac(1, 2)), nc(letter.A, frac(3, 2))},
		}
		for _, tc := range up {
			got, err := tc.from.Raise()
			if assert.NoError(t, err, tc.name) {
				assertName(t, tc.want.l, tc.want.a, got)
			}
		}

		down := []struct {
			name string
			from spelling.Spelling
			want nameCase
		}{
			{"C down", spelling.C(), nc(letter.C, alteration.Flat())},
			{"C flat down", spelling.CFlat(), nc(letter.C, alteration.DoubleFlat())},
			{"D down", spelling.D(), nc(letter.D, alteration.Flat())},
			{"D-3/2 down", spelling.MustNew(letter.D, frac(-3, 2)), nc(letter.D, frac(-5, 2))},
		}
		for _, tc := range down {
			got, err := tc.from.Lower()
			if assert.NoError(t, err, tc.name) {
				assertName(t, tc.want.l, tc.want.a, got)
			}
		}
	})

	t.Run("no enharmonic change", func(t *testing.T) {
		t.Parallel()

		got := spelling.CSharp().MustRaise()
		assert.False(t, got.Equal(spelling.D()))
		assertName(t, letter.C, alteration.DoubleSharp(), got)
	})

	for _, tc := range setN() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			x := tc.build()

			up, err := x.Raise()
			require.NoError(t, err)
			assertName(t, tc.l, tc.a.MustRaise(), up)

			down, err := x.Lower()
			require.NoError(t, err)
			assertName(t, tc.l, tc.a.MustLower(), down)

			back, err := up.Lower()
			require.NoError(t, err)
			assert.True(t, back.Equal(x), "lower after raise")

			back, err = down.Raise()
			require.NoError(t, err)
			assert.True(t, back.Equal(x), "raise after lower")
		})
	}
}

// K-5: raising and lowering by a whole number n.
func TestRaiseByLowerBy(t *testing.T) {
	t.Parallel()

	t.Run("examples", func(t *testing.T) {
		t.Parallel()

		got, err := spelling.C().RaiseBy(3)
		require.NoError(t, err)
		assertName(t, letter.C, whole(3), got)

		got, err = spelling.CSharp().RaiseBy(-2)
		require.NoError(t, err)
		assertName(t, letter.C, alteration.Flat(), got)

		got, err = spelling.CDoubleSharp().LowerBy(4)
		require.NoError(t, err)
		assertName(t, letter.C, alteration.DoubleFlat(), got)

		got, err = spelling.MustNew(letter.A, frac(1, 2)).RaiseBy(2)
		require.NoError(t, err)
		assertName(t, letter.A, frac(5, 2), got)
	})

	for _, tc := range setN() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			x := tc.build()

			for n := -5; n <= 5; n++ {
				up, err := x.RaiseBy(n)
				require.NoError(t, err, "RaiseBy(%d)", n)
				assertName(t, tc.l, tc.a.MustRaiseBy(n), up)

				down, err := x.LowerBy(n)
				require.NoError(t, err, "LowerBy(%d)", n)
				assertName(t, tc.l, tc.a.MustLowerBy(n), down)

				opposite, err := x.RaiseBy(-n)
				require.NoError(t, err, "RaiseBy(%d)", -n)
				assert.Truef(t, down.Equal(opposite), "LowerBy(%d) must equal RaiseBy(%d)", n, -n)

				back, err := up.LowerBy(n)
				require.NoError(t, err)
				assert.Truef(t, back.Equal(x), "LowerBy(%d) after RaiseBy(%d)", n, n)
			}

			zero, err := x.RaiseBy(0)
			require.NoError(t, err)
			assert.True(t, zero.Equal(x), "RaiseBy(0)")

			zero, err = x.LowerBy(0)
			require.NoError(t, err)
			assert.True(t, zero.Equal(x), "LowerBy(0)")

			one, err := x.RaiseBy(1)
			require.NoError(t, err)
			raised, err := x.Raise()
			require.NoError(t, err)
			assert.True(t, one.Equal(raised), "RaiseBy(1) must equal Raise")

			one, err = x.LowerBy(1)
			require.NoError(t, err)
			lowered, err := x.Lower()
			require.NoError(t, err)
			assert.True(t, one.Equal(lowered), "LowerBy(1) must equal Lower")
		})
	}
}

// K-6: the natural name keeps the letter and takes the natural.
func TestNatural(t *testing.T) {
	t.Parallel()

	examples := []struct {
		name string
		from spelling.Spelling
		want letter.Letter
	}{
		{"C sharp", spelling.CSharp(), letter.C},
		{"D flat", spelling.DFlat(), letter.D},
		{"A+1/2", spelling.MustNew(letter.A, frac(1, 2)), letter.A},
		{"E+3", spelling.MustNew(letter.E, whole(3)), letter.E},
		{"C", spelling.C(), letter.C},
	}
	for _, tc := range examples {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assertName(t, tc.want, alteration.Natural(), tc.from.Natural())
		})
	}

	for _, tc := range setN() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			natural := tc.build().Natural()
			assertName(t, tc.l, alteration.Natural(), natural)
			assert.True(t, natural.Natural().Equal(natural), "the natural name of a natural name is itself")
		})
	}
}

// K-7: the same letter with another value.
func TestWithAlteration(t *testing.T) {
	t.Parallel()

	t.Run("examples", func(t *testing.T) {
		t.Parallel()

		assertName(t, letter.C, alteration.Flat(), spelling.CSharp().WithAlteration(alteration.Flat()))
		assertName(t, letter.C, frac(1, 2), spelling.C().WithAlteration(frac(1, 2)))
		assertName(t, letter.B, frac(-7, 2), spelling.BSharp().WithAlteration(frac(-7, 2)))
	})

	for _, tc := range setN() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			x := tc.build()

			for _, a := range setS() {
				assertName(t, tc.l, a.a, x.WithAlteration(a.a))
			}

			assert.True(t, x.WithAlteration(tc.a).Equal(x), "the same value gives the same name")
			assert.True(t, x.WithAlteration(alteration.Natural()).Equal(x.Natural()), "the natural gives the natural name")
		})
	}
}

// K-8: shifting a name by an alteration value keeps the letter.
func TestRaiseByAlterationLowerByAlteration(t *testing.T) {
	t.Parallel()

	half := frac(1, 2)

	t.Run("examples", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name string
			got  func() (spelling.Spelling, error)
			want nameCase
		}{
			{"C sharp up 1/2", func() (spelling.Spelling, error) { return spelling.CSharp().RaiseByAlteration(half) }, nc(letter.C, frac(3, 2))},
			{"A+1/2 up 1/2", func() (spelling.Spelling, error) {
				return spelling.MustNew(letter.A, half).RaiseByAlteration(half)
			}, nc(letter.A, alteration.Sharp())},
			{"D flat up sharp", func() (spelling.Spelling, error) {
				return spelling.DFlat().RaiseByAlteration(alteration.Sharp())
			}, nc(letter.D, alteration.Natural())},
			{"C+1/3 up 1/2", func() (spelling.Spelling, error) {
				return spelling.MustNew(letter.C, frac(1, 3)).RaiseByAlteration(half)
			}, nc(letter.C, frac(5, 6))},
			{"C down 1/2", func() (spelling.Spelling, error) { return spelling.C().LowerByAlteration(half) }, nc(letter.C, frac(-1, 2))},
			{"C sharp down 1/2", func() (spelling.Spelling, error) { return spelling.CSharp().LowerByAlteration(half) }, nc(letter.C, frac(1, 2))},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				got, err := tc.got()
				require.NoError(t, err)
				assertName(t, tc.want.l, tc.want.a, got)
			})
		}
	})

	for _, tc := range setN() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			x := tc.build()

			for _, d := range setS() {
				up, err := x.RaiseByAlteration(d.a)
				require.NoError(t, err, "up %s", d.name)
				assertName(t, tc.l, tc.a.MustAdd(d.a), up)

				down, err := x.LowerByAlteration(d.a)
				require.NoError(t, err, "down %s", d.name)
				assertName(t, tc.l, tc.a.MustSub(d.a), down)

				back, err := up.LowerByAlteration(d.a)
				require.NoError(t, err)
				assert.Truef(t, back.Equal(x), "down %s after up %s", d.name, d.name)

				viaUp, err := x.RaiseByAlteration(alteration.Natural().MustSub(d.a))
				require.NoError(t, err)
				assert.Truef(t, down.Equal(viaUp), "down %s must equal up by the opposite value", d.name)

				if d.a.IsInteger() {
					n, _ := d.a.Fraction()

					byInt, err := x.RaiseBy(n)
					require.NoError(t, err)
					assert.Truef(t, up.Equal(byInt), "up %s must equal RaiseBy(%d)", d.name, n)

					byInt, err = x.LowerBy(n)
					require.NoError(t, err)
					assert.Truef(t, down.Equal(byInt), "down %s must equal LowerBy(%d)", d.name, n)
				}
			}

			same, err := x.RaiseByAlteration(alteration.Natural())
			require.NoError(t, err)
			assert.True(t, same.Equal(x), "up by the natural")

			sharp, err := x.RaiseByAlteration(alteration.Sharp())
			require.NoError(t, err)
			raised, err := x.Raise()
			require.NoError(t, err)
			assert.True(t, sharp.Equal(raised), "up by the sharp must equal Raise")
		})
	}
}

// K-21: both parts at once equal the parts taken separately.
func TestParts(t *testing.T) {
	t.Parallel()

	for _, tc := range concatNames(setN(), boundaryNames()) {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			x, err := spelling.New(tc.l, tc.a)
			require.NoError(t, err)

			l, a := x.Parts()
			assert.Equal(t, tc.l, l, "letter part")
			assertAlteration(t, tc.a, a)
			assert.Equal(t, x.Letter(), l, "Parts and Letter")
			assert.Equal(t, x.Alteration(), a, "Parts and Alteration")
		})
	}
}

// K-15: operations on valid input give a result without an error or a panic.
func TestOperationsWithoutErrors(t *testing.T) {
	t.Parallel()

	t.Run("operations without an error result", func(t *testing.T) {
		t.Parallel()

		for _, x := range setN() {
			for _, y := range setN() {
				require.NotPanics(t, func() {
					xs, ys := x.build(), y.build()
					_ = xs.Letter()
					_ = xs.Alteration()
					_, _ = xs.Parts()
					_ = xs.Equal(ys)
					_ = xs.Natural()
					_ = xs.Compare(ys)
					_ = xs.LetterDistanceUp(ys)
					_ = xs.LetterDistanceDown(ys)
					_ = xs.IsValid()
					_ = xs.IsNatural()
					_ = xs.IsInteger()
					_ = xs.Direction()
					_ = xs.SameLetter(ys)
					_ = xs.SameAlteration(ys)
					_ = xs.String()

					for _, a := range setS() {
						_ = xs.WithAlteration(a.a)
					}
				}, "%s and %s", x.name, y.name)
			}
		}

		require.NotPanics(t, func() {
			for _, p := range predefinedCases() {
				_ = p.fn()
			}

			_ = spelling.Predefined()
		})
	})

	t.Run("machine form is written without an error", func(t *testing.T) {
		t.Parallel()

		for _, x := range setN() {
			xs := x.build()

			require.NotPanics(t, func() {
				_, err := xs.MarshalText()
				require.NoError(t, err)

				_, err = xs.AppendText(nil)
				require.NoError(t, err)

				_, err = xs.MarshalJSON()
				require.NoError(t, err)
			}, x.name)
		}
	})

	t.Run("operations with an error result on valid input", func(t *testing.T) {
		t.Parallel()

		for _, l := range letter.All() {
			for _, a := range setS() {
				_, err := spelling.New(l, a.a)
				require.NoError(t, err, "New(%v, %s)", l, a.name)
			}
		}

		for _, x := range setN() {
			xs := x.build()

			_, err := xs.Raise()
			require.NoError(t, err, "%s Raise", x.name)

			_, err = xs.Lower()
			require.NoError(t, err, "%s Lower", x.name)

			for n := -5; n <= 5; n++ {
				_, err = xs.RaiseBy(n)
				require.NoError(t, err, "%s RaiseBy(%d)", x.name, n)

				_, err = xs.LowerBy(n)
				require.NoError(t, err, "%s LowerBy(%d)", x.name, n)
			}

			for _, d := range setS() {
				_, err = xs.RaiseByAlteration(d.a)
				require.NoError(t, err, "%s RaiseByAlteration(%s)", x.name, d.name)

				_, err = xs.LowerByAlteration(d.a)
				require.NoError(t, err, "%s LowerByAlteration(%s)", x.name, d.name)
			}

			for _, y := range setN() {
				_, err = xs.AlterationDifference(y.build())
				require.NoError(t, err, "%s AlterationDifference(%s)", x.name, y.name)
			}
		}
	})
}
