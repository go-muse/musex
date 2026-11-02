package number_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-muse/musex/octave/number"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// E-1...E-4: the four sentinel errors exist and are distinct; none of them
// matches another with errors.Is.
func TestSentinelErrorsAreDistinct(t *testing.T) {
	t.Parallel()

	all := allErrors()
	for i, a := range all {
		require.Error(t, a, "sentinel %d is not nil", i)
		for j, b := range all {
			assert.Equal(t, i == j, errors.Is(a, b), "errors.Is(%v, %v)", a, b)
		}
	}
}

// K-10: a skipped number has no octave: ToASPN and Convert from the numbering
// give ErrSkipped and nothing else, with the value 0; no nearest number or
// other octave is substituted.
func TestSkippedNumber(t *testing.T) {
	t.Parallel()

	t.Run("Franco-Belgian ToASPN(0)", func(t *testing.T) {
		t.Parallel()

		got, err := number.FrancoBelgian().ToASPN(0)
		assertOutcome(t, outcome{err: number.ErrSkipped}, got, err, "FrancoBelgian().ToASPN(0)")
	})

	for _, to := range builtins() {
		name := "Convert 0 from Franco-Belgian to " + to.name
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := number.Convert(0, number.FrancoBelgian(), to.n)
			assertOutcome(t, outcome{err: number.ErrSkipped}, got, err, name)
		})
	}

	cases := []struct {
		aspn, num int
		skipped   []int
		x         int
	}{
		{aspn: 4, num: 4, skipped: []int{0, 1}, x: 0},
		{aspn: 4, num: 4, skipped: []int{0, 1}, x: 1},
		{aspn: 4, num: 4, skipped: []int{6}, x: 6},
		{aspn: 4, num: 4, skipped: []int{-10}, x: -10},
	}

	for _, tc := range cases {
		name := fmt.Sprintf("%s number %d", ownName(tc.aspn, tc.num, tc.skipped), tc.x)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			n := own(t, tc.aspn, tc.num, tc.skipped...)
			want := outcome{err: number.ErrSkipped}
			require.Equal(t, want, n.spec.toASPN(tc.x), "oracle")

			got, err := n.n.ToASPN(tc.x)
			assertOutcome(t, want, got, err, "ToASPN")

			for _, to := range builtins() {
				got, err := number.Convert(tc.x, n.n, to.n)
				assertOutcome(t, want, got, err, "Convert to "+to.name)
			}
		})
	}
}

// K-11: an invalid definition of an own numbering gives ErrInvalidDefinition
// and the zero (undefined) value; limits of int are valid inputs.
func TestNew_Invalid(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		aspn, num int
		skipped   []int
	}{
		{name: "anchor number skipped", aspn: 4, num: 0, skipped: []int{0}},
		{name: "anchor number skipped among others", aspn: 4, num: 3, skipped: []int{5, 3}},
		{name: "repeated skipped number", aspn: 4, num: 3, skipped: []int{0, 0}},
		{name: "anchor number MaxInt skipped", aspn: maxInt, num: maxInt, skipped: []int{maxInt}},
		{name: "repeat not adjacent", aspn: 0, num: 0, skipped: []int{2, 1, 2}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			n, err := number.New(tc.aspn, tc.num, tc.skipped...)
			assertOnlyError(t, err, number.ErrInvalidDefinition, ownName(tc.aspn, tc.num, tc.skipped))
			assert.False(t, n.IsDefined(), "the result is the zero value")
		})
	}

	t.Run("limits of int are valid", func(t *testing.T) {
		t.Parallel()

		n, err := number.New(minInt, maxInt, minInt, maxInt-1)
		require.NoError(t, err)
		assert.True(t, n.IsDefined())

		spec := newSpec(minInt, maxInt, minInt, maxInt-1)
		for _, v := range edgeValues() {
			got, err := n.FromASPN(v)
			assertOutcome(t, spec.fromASPN(v), got, err, fmt.Sprintf("FromASPN(%d)", v))
			got, err = n.ToASPN(v)
			assertOutcome(t, spec.toASPN(v), got, err, fmt.Sprintf("ToASPN(%d)", v))
		}
	})
}

// edgeCase is one row of the architecture table "Boundary cases".
type edgeCase struct {
	name   string
	call   func() (int, error)
	oracle outcome
	want   outcome
}

// K-12: the architecture table of boundary cases at the limits of int. Every
// expected value is stated explicitly and also checked against the oracle.
// No operation returns a value wrapped around the limits.
func TestBoundaryCases(t *testing.T) {
	t.Parallel()

	tbl := builtinTable()
	aspn, fb, mex, c3, c5 := tbl[idxASPN], tbl[idxFrancoBelgian], tbl[idxMexican], tbl[idxMiddleC3], tbl[idxMiddleC5]

	val := func(v int) outcome { return outcome{value: v} }
	oor := outcome{err: number.ErrOutOfRange}
	skip := outcome{err: number.ErrSkipped}

	from := func(b builtin, a int, want outcome) edgeCase {
		return edgeCase{
			name:   fmt.Sprintf("%s.FromASPN(%d)", b.label, a),
			call:   func() (int, error) { return b.get().FromASPN(a) },
			oracle: b.spec.fromASPN(a),
			want:   want,
		}
	}
	to := func(b builtin, x int, want outcome) edgeCase {
		return edgeCase{
			name:   fmt.Sprintf("%s.ToASPN(%d)", b.label, x),
			call:   func() (int, error) { return b.get().ToASPN(x) },
			oracle: b.spec.toASPN(x),
			want:   want,
		}
	}
	ownFrom := func(s spec, a int, want outcome) edgeCase {
		return edgeCase{
			name: fmt.Sprintf("%s.FromASPN(%d)", ownName(s.aspn, s.num, s.skipped), a),
			call: func() (int, error) {
				n, err := number.New(s.aspn, s.num, s.skipped...)
				if err != nil {
					return 0, fmt.Errorf("unexpected error of New: %w", err)
				}

				return n.FromASPN(a)
			},
			oracle: s.fromASPN(a),
			want:   want,
		}
	}
	ownTo := func(s spec, x int, want outcome) edgeCase {
		return edgeCase{
			name: fmt.Sprintf("%s.ToASPN(%d)", ownName(s.aspn, s.num, s.skipped), x),
			call: func() (int, error) {
				n, err := number.New(s.aspn, s.num, s.skipped...)
				if err != nil {
					return 0, fmt.Errorf("unexpected error of New: %w", err)
				}

				return n.ToASPN(x)
			},
			oracle: s.toASPN(x),
			want:   want,
		}
	}
	convert := func(x int, src, dst spec, want outcome) edgeCase {
		return edgeCase{
			name: fmt.Sprintf("Convert(%d, %s, %s)", x, ownName(src.aspn, src.num, src.skipped), ownName(dst.aspn, dst.num, dst.skipped)),
			call: func() (int, error) {
				fn, err := number.New(src.aspn, src.num, src.skipped...)
				if err != nil {
					return 0, fmt.Errorf("unexpected error of New: %w", err)
				}
				tn, err := number.New(dst.aspn, dst.num, dst.skipped...)
				if err != nil {
					return 0, fmt.Errorf("unexpected error of New: %w", err)
				}

				return number.Convert(x, fn, tn)
			},
			oracle: convertSpec(x, src, dst),
			want:   want,
		}
	}
	convertBuiltins := func(x int, src, dst builtin, want outcome) edgeCase {
		return edgeCase{
			name:   fmt.Sprintf("Convert(%d, %s, %s)", x, src.label, dst.label),
			call:   func() (int, error) { return number.Convert(x, src.get(), dst.get()) },
			oracle: convertSpec(x, src.spec, dst.spec),
			want:   want,
		}
	}

	shiftMax := newSpec(0, maxInt)
	maxToMin := newSpec(maxInt, minInt)
	skipMax := newSpec(0, 0, maxInt)
	skipMin := newSpec(0, 0, minInt)
	zeroToMin := newSpec(0, minInt)
	oneToMin := newSpec(1, minInt)

	cases := []edgeCase{
		from(aspn, minInt, val(minInt)),
		from(aspn, maxInt, val(maxInt)),
		to(aspn, minInt, val(minInt)),
		to(aspn, maxInt, val(maxInt)),

		from(c3, minInt, oor),
		from(c3, maxInt, val(maxInt-1)),
		to(c3, maxInt, oor),
		to(c3, minInt, val(minInt+1)),

		from(c5, maxInt, oor),
		from(c5, minInt, val(minInt+1)),
		to(c5, minInt, oor),
		to(c5, maxInt, val(maxInt-1)),
		from(mex, maxInt, oor),
		from(mex, minInt, val(minInt+1)),
		to(mex, minInt, oor),
		to(mex, maxInt, val(maxInt-1)),

		from(fb, minInt, oor),
		from(fb, minInt+1, oor),
		from(fb, minInt+2, val(minInt)),
		from(fb, maxInt, val(maxInt-1)),
		to(fb, maxInt, oor),
		to(fb, minInt, val(minInt+2)),
		to(fb, 0, skip),

		ownFrom(shiftMax, 1, oor),
		ownFrom(shiftMax, 0, val(maxInt)),
		ownTo(shiftMax, maxInt, val(0)),

		ownFrom(maxToMin, maxInt, val(minInt)),
		ownFrom(maxToMin, 0, oor),
		ownTo(maxToMin, minInt, val(maxInt)),

		ownFrom(skipMax, maxInt-1, val(maxInt-1)),
		ownFrom(skipMax, maxInt, oor),
		ownTo(skipMax, maxInt, skip),

		ownFrom(skipMin, minInt+1, val(minInt+1)),
		ownFrom(skipMin, minInt, oor),
		ownTo(skipMin, minInt, skip),

		ownTo(zeroToMin, maxInt, oor),

		convert(maxInt, zeroToMin, zeroToMin, val(maxInt)),
		convert(maxInt, zeroToMin, oneToMin, val(maxInt-1)),
		convertBuiltins(maxInt, c3, c5, oor),
		convertBuiltins(maxInt, c5, c3, val(maxInt-2)),
		{
			name:   "Convert(MinInt, New(0, MaxInt), ASPN)",
			call:   func() (int, error) { return number.Convert(minInt, number.MustNew(0, maxInt), number.ASPN()) },
			oracle: convertSpec(minInt, shiftMax, aspn.spec),
			want:   oor,
		},
		{
			name:   "Convert(0, Franco-Belgian, ASPN)",
			call:   func() (int, error) { return number.Convert(0, number.FrancoBelgian(), number.ASPN()) },
			oracle: convertSpec(0, fb.spec, aspn.spec),
			want:   skip,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, tc.oracle, "oracle agrees with the architecture table")

			got, err := tc.call()
			assertOutcome(t, tc.want, got, err, tc.name)
		})
	}
}

// K-12: near the limits of int and around zero, FromASPN, ToASPN and Convert
// between all numberings of the boundary set match the exact oracle: either
// the exact value or ErrOutOfRange (ErrSkipped for skipped numbers).
func TestBoundaryOracle(t *testing.T) {
	t.Parallel()

	set := edgeSet(t)
	values := edgeValues()

	for _, n := range set {
		t.Run(n.name, func(t *testing.T) {
			t.Parallel()

			for _, v := range values {
				got, err := n.n.FromASPN(v)
				assertOutcome(t, n.spec.fromASPN(v), got, err, fmt.Sprintf("FromASPN(%d)", v))

				got, err = n.n.ToASPN(v)
				assertOutcome(t, n.spec.toASPN(v), got, err, fmt.Sprintf("ToASPN(%d)", v))

				for _, to := range set {
					got, err := number.Convert(v, n.n, to.n)
					assertOutcome(t, convertSpec(v, n.spec, to.spec), got, err, fmt.Sprintf("Convert(%d) to %s", v, to.name))
				}
			}
		})
	}
}

// Equal at the limits of int: comparing shifts num - aspn in int would
// overflow; each pair is checked in both directions and against the oracle.
func TestEqual_Limits(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		p, q spec
		want bool
	}{
		{name: "New(0, MaxInt) and New(1, MinInt)", p: newSpec(0, maxInt), q: newSpec(1, minInt), want: false},
		{name: "New(MinInt, 0) and New(0, MinInt)", p: newSpec(minInt, 0), q: newSpec(0, minInt), want: false},
		{name: "New(0, MinInt) and New(1, MinInt+1)", p: newSpec(0, minInt), q: newSpec(1, minInt+1), want: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, equalSpec(tc.p, tc.q), "oracle")
			require.Equal(t, tc.want, equalSpec(tc.q, tc.p), "oracle, reversed")

			p := own(t, tc.p.aspn, tc.p.num)
			q := own(t, tc.q.aspn, tc.q.num)
			assert.Equal(t, tc.want, p.n.Equal(q.n))
			assert.Equal(t, tc.want, q.n.Equal(p.n), "reversed")
		})
	}

	t.Run("boundary set against the oracle", func(t *testing.T) {
		t.Parallel()

		set := edgeSet(t)
		for _, p := range set {
			for _, q := range set {
				assert.Equal(t, equalSpec(p.spec, q.spec), p.n.Equal(q.n), "%s and %s", p.name, q.name)
			}
		}
	})
}

// The order of checks with several causes: undefined before skipped, skipped
// before out of range.
func TestErrorOrder(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		call func() (int, error)
		want error
	}{
		{
			name: "Convert(0, Franco-Belgian, zero)",
			call: func() (int, error) { return number.Convert(0, number.FrancoBelgian(), number.Numbering{}) },
			want: number.ErrUndefined,
		},
		{
			name: "Convert(MaxInt, zero, Middle C = C5)",
			call: func() (int, error) { return number.Convert(maxInt, number.Numbering{}, number.MiddleC5()) },
			want: number.ErrUndefined,
		},
		{
			name: "zero ToASPN(0)",
			call: func() (int, error) { return number.Numbering{}.ToASPN(0) },
			want: number.ErrUndefined,
		},
		{
			name: "New(0, 0, MaxInt).ToASPN(MaxInt)",
			call: func() (int, error) { return number.MustNew(0, 0, maxInt).ToASPN(maxInt) },
			want: number.ErrSkipped,
		},
		{
			name: "Convert(MaxInt, New(0, 0, MaxInt), Middle C = C5)",
			call: func() (int, error) { return number.Convert(maxInt, number.MustNew(0, 0, maxInt), number.MiddleC5()) },
			want: number.ErrSkipped,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.call()
			assertOutcome(t, outcome{err: tc.want}, got, err, tc.name)
		})
	}
}
