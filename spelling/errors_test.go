package spelling_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// shiftCase is an operation that returns a name, with its Must variant and
// the expected result: an out-of-range error or the name want.
type shiftCase struct {
	name    string
	op      func() (spelling.Spelling, error)
	must    func() spelling.Spelling
	wantErr bool
	want    nameCase
}

func checkShiftCase(t *testing.T, tc shiftCase) {
	t.Helper()

	got, err := tc.op()

	if tc.wantErr {
		assertOutOfRange(t, err)
		assertUnset(t, got)

		mustErr := panicError(t, func() { _ = tc.must() })
		assert.ErrorIs(t, mustErr, alteration.ErrOutOfRange)
		assert.NotErrorIs(t, mustErr, letter.ErrInvalid)

		return
	}

	require.NoError(t, err)
	assertName(t, tc.want.l, tc.want.a, got)

	var mustGot spelling.Spelling

	require.NotPanics(t, func() { mustGot = tc.must() })
	assert.Equal(t, got, mustGot, "the Must variant must return the same name")
}

// K-9: creating a name from an invalid letter is an invalid-letter error, not
// a name with the letter C.
func TestNewInvalidLetter(t *testing.T) {
	t.Parallel()

	letters := []struct {
		name string
		l    letter.Letter
	}{
		{"zero", letter.Letter(0)},
		{"8", letter.Letter(8)},
		{"42", letter.Letter(42)},
		{"255", letter.Letter(255)},
		{"B+1", letter.B + 1},
	}
	values := []alteration.Alteration{alteration.Natural(), alteration.Sharp(), frac(1, 2)}

	for _, lc := range letters {
		for _, a := range values {
			t.Run(lc.name+" "+a.String(), func(t *testing.T) {
				t.Parallel()

				got, err := spelling.New(lc.l, a)
				assertInvalidLetter(t, err)
				assertUnset(t, got)
				assert.NotEqual(t, letter.C, got.Letter(), "the letter must not be C")
				assert.False(t, got.Equal(spelling.C().WithAlteration(a)), "the result must not be C")

				mustErr := panicError(t, func() { _ = spelling.MustNew(lc.l, a) })
				assert.ErrorIs(t, mustErr, letter.ErrInvalid)
				assert.NotErrorIs(t, mustErr, alteration.ErrOutOfRange)
			})
		}
	}
}

// K-10 items 1-3 and the boundary table of the architecture: raising and
// lowering by one and by n beyond the guaranteed range.
func TestRaiseLowerOutOfRange(t *testing.T) {
	t.Parallel()

	top := spelling.MustNew(letter.C, whole(maxMagnitude))
	bottom := spelling.MustNew(letter.C, whole(-maxMagnitude))
	nearTop := spelling.MustNew(letter.C, frac(2*maxMagnitude-1, 2))
	nearBottom := spelling.MustNew(letter.C, frac(-(2*maxMagnitude-1), 2))
	edges := edgeNames()
	edgeTop, edgeBottom := edges[0].build(), edges[1].build()

	raise := func(s spelling.Spelling) (func() (spelling.Spelling, error), func() spelling.Spelling) {
		return s.Raise, s.MustRaise
	}
	lower := func(s spelling.Spelling) (func() (spelling.Spelling, error), func() spelling.Spelling) {
		return s.Lower, s.MustLower
	}
	raiseBy := func(s spelling.Spelling, n int) (func() (spelling.Spelling, error), func() spelling.Spelling) {
		return func() (spelling.Spelling, error) { return s.RaiseBy(n) },
			func() spelling.Spelling { return s.MustRaiseBy(n) }
	}
	lowerBy := func(s spelling.Spelling, n int) (func() (spelling.Spelling, error), func() spelling.Spelling) {
		return func() (spelling.Spelling, error) { return s.LowerBy(n) },
			func() spelling.Spelling { return s.MustLowerBy(n) }
	}

	newCase := func(name string, op func() (spelling.Spelling, error), must func() spelling.Spelling) shiftCase {
		return shiftCase{name: name, op: op, must: must, wantErr: true}
	}
	okCase := func(name string, want nameCase, op func() (spelling.Spelling, error), must func() spelling.Spelling) shiftCase {
		return shiftCase{name: name, op: op, must: must, want: want}
	}

	var cases []shiftCase

	op, must := raise(top)
	cases = append(cases, newCase("(C, M) up 1", op, must))
	op, must = lower(bottom)
	cases = append(cases, newCase("(C, -M) down 1", op, must))
	op, must = raise(nearTop)
	cases = append(cases, newCase("(C, M-1/2) up 1", op, must))
	op, must = lower(nearBottom)
	cases = append(cases, newCase("(C, -(M-1/2)) down 1", op, must))
	op, must = raise(edgeTop)
	cases = append(cases, newCase("(C, (MQ-1)/Q) up 1", op, must))
	op, must = lower(edgeBottom)
	cases = append(cases, newCase("(C, -(MQ-1)/Q) down 1", op, must))

	for _, n := range []int{1, 2, math.MaxInt} {
		op, must = raiseBy(top, n)
		cases = append(cases, newCase("(C, M) up n="+itoa(n), op, must))
		op, must = lowerBy(bottom, n)
		cases = append(cases, newCase("(C, -M) down n="+itoa(n), op, must))
		op, must = raiseBy(bottom, -n)
		cases = append(cases, newCase("(C, -M) up -n="+itoa(-n), op, must))
		op, must = lowerBy(top, -n)
		cases = append(cases, newCase("(C, M) down -n="+itoa(-n), op, must))
	}

	for _, n := range []int{maxMagnitude + 1, math.MaxInt, math.MinInt} {
		op, must = raiseBy(spelling.C(), n)
		cases = append(cases, newCase("C up "+itoa(n), op, must))
		op, must = lowerBy(spelling.C(), n)
		cases = append(cases, newCase("C down "+itoa(n), op, must))
	}

	op, must = raiseBy(spelling.C(), maxMagnitude)
	cases = append(cases, okCase("C up M", nc(letter.C, whole(maxMagnitude)), op, must))
	op, must = lowerBy(spelling.C(), maxMagnitude)
	cases = append(cases, okCase("C down M", nc(letter.C, whole(-maxMagnitude)), op, must))
	op, must = raiseBy(bottom, 2*maxMagnitude)
	cases = append(cases, okCase("(C, -M) up 2M", nc(letter.C, whole(maxMagnitude)), op, must))
	op, must = lowerBy(top, 2*maxMagnitude)
	cases = append(cases, okCase("(C, M) down 2M", nc(letter.C, whole(-maxMagnitude)), op, must))
	op, must = raise(spelling.MustNew(letter.C, whole(maxMagnitude-1)))
	cases = append(cases, okCase("(C, M-1) up 1", nc(letter.C, whole(maxMagnitude)), op, must))
	op, must = lower(spelling.MustNew(letter.C, whole(-(maxMagnitude - 1))))
	cases = append(cases, okCase("(C, -(M-1)) down 1", nc(letter.C, whole(-maxMagnitude)), op, must))

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			checkShiftCase(t, tc)
		})
	}
}

// K-10 item 4 and the edges of finding N2-1: raising and lowering fail exactly
// when the same alteration operation fails; otherwise the value is the value
// of that operation and the letter is kept.
func TestRaiseLowerMatchAlteration(t *testing.T) {
	t.Parallel()

	for _, tc := range concatNames(setN(), boundaryNames(), edgeNames()) {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			x := tc.build()

			type step struct {
				name   string
				got    func() (spelling.Spelling, error)
				must   func() spelling.Spelling
				oracle func() (alteration.Alteration, error)
			}

			steps := []step{
				{"Raise", x.Raise, x.MustRaise, tc.a.Raise},
				{"Lower", x.Lower, x.MustLower, tc.a.Lower},
			}

			for n := -5; n <= 5; n++ {
				steps = append(steps,
					step{
						"RaiseBy(" + itoa(n) + ")",
						func() (spelling.Spelling, error) { return x.RaiseBy(n) },
						func() spelling.Spelling { return x.MustRaiseBy(n) },
						func() (alteration.Alteration, error) { return tc.a.RaiseBy(n) },
					},
					step{
						"LowerBy(" + itoa(n) + ")",
						func() (spelling.Spelling, error) { return x.LowerBy(n) },
						func() spelling.Spelling { return x.MustLowerBy(n) },
						func() (alteration.Alteration, error) { return tc.a.LowerBy(n) },
					},
				)
			}

			for _, s := range steps {
				want, wantErr := s.oracle()
				got, err := s.got()

				if wantErr != nil {
					if assertOutOfRange(t, err) {
						assertUnset(t, got)
					}

					assertPanicsWithErrorIs(t, alteration.ErrOutOfRange, func() { _ = s.must() })

					continue
				}

				if assert.NoError(t, err, s.name) {
					assertName(t, tc.l, want, got)
				}

				assert.NotPanics(t, func() { assert.Equal(t, got, s.must(), s.name) })
			}
		})
	}
}

// K-22: shifting by a value beyond the guaranteed range.
func TestRaiseByAlterationOutOfRange(t *testing.T) {
	t.Parallel()

	half := frac(1, 2)
	top := spelling.MustNew(letter.C, whole(maxMagnitude))
	bottom := spelling.MustNew(letter.C, whole(-maxMagnitude))
	nearTop := spelling.MustNew(letter.C, frac(2*maxMagnitude-1, 2))
	smallest := spelling.MustNew(letter.C, frac(1, maxDenominator))

	up := func(s spelling.Spelling, d alteration.Alteration) (func() (spelling.Spelling, error), func() spelling.Spelling) {
		return func() (spelling.Spelling, error) { return s.RaiseByAlteration(d) },
			func() spelling.Spelling { return s.MustRaiseByAlteration(d) }
	}
	down := func(s spelling.Spelling, d alteration.Alteration) (func() (spelling.Spelling, error), func() spelling.Spelling) {
		return func() (spelling.Spelling, error) { return s.LowerByAlteration(d) },
			func() spelling.Spelling { return s.MustLowerByAlteration(d) }
	}

	var cases []shiftCase

	add := func(name string, wantErr bool, want nameCase, op func() (spelling.Spelling, error), must func() spelling.Spelling) {
		cases = append(cases, shiftCase{name: name, op: op, must: must, wantErr: wantErr, want: want})
	}

	op, must := up(top, half)
	add("(C, M) up 1/2", true, nameCase{}, op, must)
	op, must = down(bottom, half)
	add("(C, -M) down 1/2", true, nameCase{}, op, must)
	op, must = up(nearTop, alteration.Sharp())
	add("(C, M-1/2) up sharp", true, nameCase{}, op, must)
	op, must = up(bottom, frac(-1, 2))
	add("(C, -M) up -1/2", true, nameCase{}, op, must)
	op, must = up(smallest, frac(1, maxDenominator-1))
	add("(C, 1/Q) up 1/(Q-1)", true, nameCase{}, op, must)

	op, must = up(smallest, frac(maxDenominator-1, maxDenominator))
	add("(C, 1/Q) up (Q-1)/Q", false, nc(letter.C, alteration.Sharp()), op, must)
	op, must = up(bottom, whole(maxMagnitude))
	add("(C, -M) up M", false, nc(letter.C, alteration.Natural()), op, must)
	op, must = up(nearTop, half)
	add("(C, M-1/2) up 1/2", false, nc(letter.C, whole(maxMagnitude)), op, must)

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			checkShiftCase(t, tc)
		})
	}

	t.Run("C sharp from (C, 1/Q)", func(t *testing.T) {
		t.Parallel()

		got, err := smallest.RaiseByAlteration(frac(maxDenominator-1, maxDenominator))
		require.NoError(t, err)
		assert.True(t, got.Equal(spelling.CSharp()))
	})
}

// K-22: shifting fails exactly when the alteration addition or subtraction
// fails; otherwise the value is that of the alteration operation.
func TestRaiseByAlterationMatchAlteration(t *testing.T) {
	t.Parallel()

	for _, tc := range concatNames(setN(), boundaryNames()) {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			x := tc.build()

			for _, d := range setS() {
				checks := []struct {
					name   string
					got    func() (spelling.Spelling, error)
					must   func() spelling.Spelling
					oracle func() (alteration.Alteration, error)
				}{
					{
						"up " + d.name,
						func() (spelling.Spelling, error) { return x.RaiseByAlteration(d.a) },
						func() spelling.Spelling { return x.MustRaiseByAlteration(d.a) },
						func() (alteration.Alteration, error) { return tc.a.Add(d.a) },
					},
					{
						"down " + d.name,
						func() (spelling.Spelling, error) { return x.LowerByAlteration(d.a) },
						func() spelling.Spelling { return x.MustLowerByAlteration(d.a) },
						func() (alteration.Alteration, error) { return tc.a.Sub(d.a) },
					},
				}

				for _, c := range checks {
					want, wantErr := c.oracle()
					got, err := c.got()

					if wantErr != nil {
						if assertOutOfRange(t, err) {
							assertUnset(t, got)
						}

						assertPanicsWithErrorIs(t, alteration.ErrOutOfRange, func() { _ = c.must() })

						continue
					}

					if assert.NoError(t, err, c.name) {
						assertName(t, tc.l, want, got)
					}

					assert.NotPanics(t, func() { assert.Equal(t, got, c.must(), c.name) })
				}
			}
		})
	}
}

// K-27: the difference of values beyond the guaranteed range.
func TestAlterationDifferenceOutOfRange(t *testing.T) {
	t.Parallel()

	top := spelling.MustNew(letter.C, whole(maxMagnitude))
	bottom := spelling.MustNew(letter.C, whole(-maxMagnitude))

	failing := []struct {
		name string
		x, y spelling.Spelling
	}{
		{"(C, M) minus (C, -M)", top, bottom},
		{"(C, -M) minus (C, M)", bottom, top},
		{"(C, 1/Q) minus (C, 1/(Q-1))", spelling.MustNew(letter.C, frac(1, maxDenominator)), spelling.MustNew(letter.C, frac(1, maxDenominator-1))},
	}

	for _, tc := range failing {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.x.AlterationDifference(tc.y)
			assertOutOfRange(t, err)
			assertAlteration(t, alteration.Natural(), got)

			mustErr := panicError(t, func() { _ = tc.x.MustAlterationDifference(tc.y) })
			assert.ErrorIs(t, mustErr, alteration.ErrOutOfRange)
			assert.NotErrorIs(t, mustErr, letter.ErrInvalid)
		})
	}

	succeeding := []struct {
		name string
		x, y spelling.Spelling
		want alteration.Alteration
	}{
		{"(C, M) minus (C, M-1)", top, spelling.MustNew(letter.C, whole(maxMagnitude-1)), alteration.Sharp()},
		{"(C, M) minus (C, M)", top, top, alteration.Natural()},
	}

	for _, tc := range succeeding {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.x.AlterationDifference(tc.y)
			require.NoError(t, err)
			assertAlteration(t, tc.want, got)

			require.NotPanics(t, func() { assertAlteration(t, tc.want, tc.x.MustAlterationDifference(tc.y)) })
		})
	}
}

// K-27: the difference fails exactly when the alteration subtraction fails.
func TestAlterationDifferenceMatchAlteration(t *testing.T) {
	t.Parallel()

	names := concatNames(setN(), boundaryNames())

	for _, x := range names {
		t.Run(x.name, func(t *testing.T) {
			t.Parallel()

			xs := x.build()

			for _, y := range names {
				ys := y.build()
				want, wantErr := x.a.Sub(y.a)
				got, err := xs.AlterationDifference(ys)

				if wantErr != nil {
					if assertOutOfRange(t, err) {
						assertAlteration(t, alteration.Natural(), got)
					}

					assertPanicsWithErrorIs(t, alteration.ErrOutOfRange, func() { _ = xs.MustAlterationDifference(ys) })

					continue
				}

				if assert.NoError(t, err, "minus %s", y.name) {
					assertAlteration(t, want, got)
				}

				assert.NotPanics(t, func() { assertAlteration(t, want, xs.MustAlterationDifference(ys)) })
			}
		})
	}
}
