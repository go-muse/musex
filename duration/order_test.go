package duration_test

import (
	"math"
	"math/big"
	"strconv"
	"strings"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tuplet"
)

// orderCase is an input that violates two adjacent steps of one row of the
// table "Order of checks" of architecture 4.2; the earlier step is reported.
type orderCase struct {
	name   string
	run    func() error
	want   error
	causes []error
}

// e drops the results of a call and keeps its error.
func e[T any](_ T, err error) error { return err }

// e2 drops two results of a call and keeps its error.
func e2[T, U any](_ T, _ U, err error) error { return err }

// TestOrderOfChecks covers architecture 4.2, "Order of checks": one case for
// each pair of adjacent steps of every row where both can fail together.
// Pairs that cannot fail together (an unset value has no sign, a range
// error needs set arguments) are covered by the single-step tests. Steps of
// the same class (two E-2 or two E-11 with the same causes) are told apart
// by the causes where they differ.
func TestOrderOfChecks(t *testing.T) {
	t.Parallel()

	q := wr(duration.Quarter, 0)
	edge := edgeWritten()
	quarterEdge := wr(duration.Quarter, 0, bigChain(4)...)
	v := dur(5, 16)
	unsetTuplet := tuplet.Tuplet{}

	// Coprime denominators near 2^200: differences and sums leave the range.
	small := new(big.Rat).SetFrac(big.NewInt(1), near2pow(200, 1)) // 1/(2^200+1)
	large := new(big.Rat).SetFrac(big.NewInt(1), near2pow(200, -1))
	smallDur, largeDur := durRat(small), durRat(large)
	smallPos, largePos := posRat(small), posRat(large)

	maxPos := posRat(ratInt(maxComponent()))
	n1, n2 := math.MaxInt, math.MaxInt-2
	a3 := wr(duration.Quarter, 0, tp(n1, 1), tp(n1, 1), tp(n1, 1))
	b3 := wr(duration.Quarter, 0, tp(n2, 1), tp(n2, 1), tp(n2, 1))
	big4 := strings.Repeat(maxIntText+":1,", 3) + maxIntText + ":1"
	big5 := big4 + "," + maxIntText + ":1"
	n2Str := strconv.Itoa(n2)

	cases := []orderCase{
		// New, NewPosition: E-2 (den == 0) before the sign.
		{"New den before sign", func() error { return e(duration.New(-1, 0)) }, duration.ErrInvalidArgument, []error{fraction.ErrZeroDenominator}},
		{"NewPosition den before sign", func() error { return e(duration.NewPosition(-1, 0)) }, duration.ErrInvalidArgument, []error{fraction.ErrZeroDenominator}},

		// Sum: E-1 before E-11.
		{"Sum unset before range", func() error { return e(duration.Sum(largeDur, smallDur, duration.Duration{})) }, duration.ErrUnset, nil},

		// Longer, Shorter, Shift: E-2 before E-6.
		{"Longer invalid before outside", func() error { return e(duration.Base(15).Longer()) }, duration.ErrInvalidArgument, nil},
		{"Shorter invalid before outside", func() error { return e(duration.Base(0).Shorter()) }, duration.ErrInvalidArgument, nil},
		{"Shift invalid before outside", func() error { return e(duration.Base(0).Shift(math.MaxInt)) }, duration.ErrInvalidArgument, nil},

		// Sub, DurationTo: E-1 before E-3 before E-11.
		{"Sub unset receiver", func() error { return e(duration.Duration{}.Sub(dur(1, 4))) }, duration.ErrUnset, nil},
		{"Sub sign before range", func() error { return e(smallDur.Sub(largeDur)) }, duration.ErrNotPositive, nil},
		{"DurationTo unset", func() error { return e(duration.Position{}.DurationTo(pos(1, 4))) }, duration.ErrUnset, nil},
		{"DurationTo sign before range", func() error { return e(largePos.DurationTo(smallPos)) }, duration.ErrNotPositive, nil},

		// Position.Sub, OffsetTo: E-1 before E-4 before E-11.
		{"Position.Sub sign before range", func() error { return e(smallPos.Sub(largeDur)) }, duration.ErrNegativePosition, nil},
		{"OffsetTo sign before range", func() error { return e(largePos.OffsetTo(smallPos)) }, duration.ErrNegativePosition, nil},
		{"OffsetTo unset argument", func() error { return e(pos(1, 2).OffsetTo(duration.Position{})) }, duration.ErrUnset, nil},

		// Mul, Div, MulInt, DivInt: E-1 before E-2.
		{"Mul unset before argument", func() error { return e(duration.Duration{}.Mul(fraction.Fraction{})) }, duration.ErrUnset, nil},
		{"Div unset before argument", func() error { return e(duration.Duration{}.Div(fraction.Zero())) }, duration.ErrUnset, nil},
		{"MulInt unset before argument", func() error { return e(duration.Duration{}.MulInt(0)) }, duration.ErrUnset, nil},
		{"DivInt unset before argument", func() error { return e(duration.Duration{}.DivInt(-1)) }, duration.ErrUnset, nil},

		// QuoRem: E-1, then the quotient before the remainder.
		{"QuoRem unset", func() error { return e2(duration.Duration{}.QuoRem(dur(1, 4))) }, duration.ErrUnset, nil},
		{"QuoRem quotient before remainder", func() error {
			p, q := near2pow(150, -1), near2pow(150, 1)
			d := durRat(new(big.Rat).SetFrac(new(big.Int).Add(new(big.Int).Lsh(p, 70), big.NewInt(1)), p))

			return e2(d.QuoRem(durRat(new(big.Rat).SetFrac(big.NewInt(1), q))))
		}, duration.ErrOutOfRange, []error{fraction.ErrNotRepresentable}},

		// NewWritten: base, dots, unset tuplet, coefficient, value.
		{"NewWritten base before unset tuplet", func() error { return e(duration.NewWritten(0, 0, unsetTuplet)) }, duration.ErrInvalidArgument, nil},
		{"NewWritten dots before unset tuplet", func() error { return e(duration.NewWritten(duration.Quarter, -1, unsetTuplet)) }, duration.ErrInvalidArgument, nil},
		{"NewWritten dots before coefficient", func() error { return e(duration.NewWritten(duration.Quarter, -1, bigChain(5)...)) }, duration.ErrInvalidArgument, nil},
		{"NewWritten unset tuplet before coefficient", func() error {
			return e(duration.NewWritten(duration.Quarter, 0, append(bigChain(5), unsetTuplet)...))
		}, duration.ErrInvalidArgument, []error{tuplet.ErrUnset}},
		{"NewWritten coefficient before value", func() error { return e(duration.NewWritten(duration.Quarter, 600, bigChain(5)...)) },
			duration.ErrOutOfRange, []error{tuplet.ErrOutOfRange, fraction.ErrOutOfRange}},

		// WithBase: E-1 before E-2.
		{"WithBase unset before invalid", func() error { return e(duration.Written{}.WithBase(0)) }, duration.ErrUnset, nil},

		// ShiftBase: E-1 before E-6 before E-11.
		{"ShiftBase unset before outside", func() error { return e(duration.Written{}.ShiftBase(100)) }, duration.ErrUnset, nil},
		{"ShiftBase outside before range", func() error { return e(edge.ShiftBase(-100)) }, duration.ErrOutsideSeries, nil},
		{"ShiftBase outside before range min int", func() error { return e(edge.ShiftBase(math.MinInt)) }, duration.ErrOutsideSeries, nil},

		// WithDots: E-1 before E-2.
		{"WithDots unset before negative", func() error { return e(duration.Written{}.WithDots(-1)) }, duration.ErrUnset, nil},

		// RemoveDot: E-1 before E-2.
		{"RemoveDot unset", func() error { return e(duration.Written{}.RemoveDot()) }, duration.ErrUnset, nil},

		// WithChain, AddOuterTuplet, AddInnerTuplet: E-1, unset tuplet,
		// coefficient, value.
		{"WithChain unset before tuplet", func() error { return e(duration.Written{}.WithChain(unsetTuplet)) }, duration.ErrUnset, nil},
		{"WithChain unset tuplet before coefficient", func() error { return e(q.WithChain(append(bigChain(5), unsetTuplet)...)) },
			duration.ErrInvalidArgument, []error{tuplet.ErrUnset}},
		{"WithChain coefficient before value", func() error { return e(wr(duration.ThousandTwentyFourth, 0).WithChain(bigChain(5)...)) },
			duration.ErrOutOfRange, []error{tuplet.ErrOutOfRange, fraction.ErrOutOfRange}},
		{"AddOuterTuplet unset before tuplet", func() error { return e(duration.Written{}.AddOuterTuplet(unsetTuplet)) }, duration.ErrUnset, nil},
		{"AddInnerTuplet unset before tuplet", func() error { return e(duration.Written{}.AddInnerTuplet(unsetTuplet)) }, duration.ErrUnset, nil},
		{"AddOuterTuplet coefficient before value", func() error { return e(quarterEdge.AddOuterTuplet(bigTuplet)) },
			duration.ErrOutOfRange, []error{tuplet.ErrOutOfRange, fraction.ErrOutOfRange}},
		{"AddInnerTuplet coefficient before value", func() error { return e(quarterEdge.AddInnerTuplet(bigTuplet)) },
			duration.ErrOutOfRange, []error{tuplet.ErrOutOfRange, fraction.ErrOutOfRange}},

		// RemoveOuterTuplet, RemoveInnerTuplet: E-1 before E-2.
		{"RemoveOuterTuplet unset", func() error { return e(duration.Written{}.RemoveOuterTuplet()) }, duration.ErrUnset, nil},
		{"RemoveInnerTuplet unset", func() error { return e(duration.Written{}.RemoveInnerTuplet()) }, duration.ErrUnset, nil},

		// Subdivide: E-1, E-7, E-6.
		{"Subdivide unset", func() error { return e2(duration.Written{}.Subdivide()) }, duration.ErrUnset, nil},
		{"Subdivide dots before outside", func() error { return e2(wr(duration.ThousandTwentyFourth, 2).Subdivide()) }, duration.ErrNoSubdivision, nil},

		// NewTied: fewer than two parts before unset parts; E-1 before E-11.
		{"NewTied count before unset", func() error { return e(duration.NewTied(duration.Written{})) }, duration.ErrInvalidArgument, nil},
		{"NewTied unset before range", func() error { return e(duration.NewTied(a3, b3, duration.Written{})) }, duration.ErrUnset, nil},

		// Tied.Append, Tied.Join: E-1 before E-11 (receiver set, argument
		// unset; a range failure needs both set).
		{"Append unset receiver", func() error { return e(duration.Tied{}.Append(q)) }, duration.ErrUnset, nil},

		// ToWritten: E-1, unset tuplet, coefficient, E-5.
		{"ToWritten unset before tuplet", func() error { return e(duration.Duration{}.ToWritten(unsetTuplet)) }, duration.ErrUnset, nil},
		{"ToWritten tuplet before coefficient", func() error { return e(dur(1, 3).ToWritten(append(bigChain(5), unsetTuplet)...)) },
			duration.ErrInvalidArgument, []error{tuplet.ErrUnset}},
		{"ToWritten coefficient before not notatable", func() error { return e(dur(1, 3).ToWritten(bigChain(5)...)) },
			duration.ErrOutOfRange, []error{tuplet.ErrOutOfRange, fraction.ErrOutOfRange}},
		{"Tied.ToWritten unset before tuplet", func() error { return e(duration.Tied{}.ToWritten(unsetTuplet)) }, duration.ErrUnset, nil},
		{"Tied.ToWritten coefficient before not notatable", func() error {
			return e(tie(q, wr(duration.Sixteenth, 0)).ToWritten(bigChain(5)...))
		}, duration.ErrOutOfRange, []error{tuplet.ErrOutOfRange, fraction.ErrOutOfRange}},

		// Notate: E-1, unset tuplet, coefficient, E-5, E-9, E-11 of a part.
		{"Notate unset before tuplet", func() error { return e(duration.Duration{}.Notate(unsetTuplet)) }, duration.ErrUnset, nil},
		{"Notate tuplet before coefficient", func() error { return e(v.Notate(append(bigChain(5), unsetTuplet)...)) },
			duration.ErrInvalidArgument, []error{tuplet.ErrUnset}},
		{"Notate coefficient before not notatable", func() error { return e(dur(1, 3).Notate(bigChain(5)...)) },
			duration.ErrOutOfRange, []error{tuplet.ErrOutOfRange, fraction.ErrOutOfRange}},
		{"Notate not notatable before too many", func() error { return e(durRat(new(big.Rat).Add(pow2(200), rat(1, 3))).Notate()) },
			duration.ErrNotNotatable, nil},
		{"NotateWith too many before part range", func() error {
			// v' = 1537/4 to the eighth with one dot: 2049 dotted eighths and
			// a sixteenth; under four MaxInt:1 the sixteenth is out of range.
			chain := bigChain(4)
			d := wr(duration.Quarter, 0, chain...).MustDuration().MustMulInt(1537)

			return e(d.NotateWith(duration.Eighth, 1, chain...))
		}, duration.ErrTooManyParts, nil},

		// NotateWith: E-1, longest, maxDots, unset tuplet, coefficient.
		{"NotateWith unset before longest", func() error { return e(duration.Duration{}.NotateWith(0, -1)) }, duration.ErrUnset, nil},
		{"NotateWith longest before tuplet", func() error { return e(v.NotateWith(0, 0, unsetTuplet)) }, duration.ErrInvalidArgument, nil},
		{"NotateWith maxDots before tuplet", func() error { return e(v.NotateWith(duration.Breve, -1, unsetTuplet)) }, duration.ErrInvalidArgument, nil},
		{"NotateWith maxDots before coefficient", func() error { return e(v.NotateWith(duration.Breve, -1, bigChain(5)...)) },
			duration.ErrInvalidArgument, nil},
		{"NotateWith tuplet before coefficient", func() error { return e(v.NotateWith(duration.Breve, 0, append(bigChain(5), unsetTuplet)...)) },
			duration.ErrInvalidArgument, []error{tuplet.ErrUnset}},
		{"NotateWith coefficient before not notatable", func() error { return e(dur(1, 3).NotateWith(duration.Breve, 0, bigChain(5)...)) },
			duration.ErrOutOfRange, []error{tuplet.ErrOutOfRange, fraction.ErrOutOfRange}},

		// Duration.Quantize: E-1, mode, step.
		{"Quantize unset before mode", func() error { return e(duration.Duration{}.Quantize(fraction.Fraction{}, 0)) }, duration.ErrUnset, nil},
		{"Quantize mode before step", func() error { return e(v.Quantize(fraction.Fraction{}, 0)) }, duration.ErrInvalidArgument, []error{rounding.ErrUnset}},
		{"Quantize unknown mode before step", func() error { return e(v.Quantize(frac(-1, 8), 200)) },
			duration.ErrInvalidArgument, []error{rounding.ErrUnknownMode}},

		// Position.Quantize: E-1, mode, step.
		{"Position.Quantize unset before mode", func() error { return e(duration.Position{}.Quantize(fraction.Zero(), 0)) }, duration.ErrUnset, nil},
		{"Position.Quantize mode before step", func() error { return e(pos(1, 3).Quantize(fraction.Zero(), 0)) },
			duration.ErrInvalidArgument, []error{rounding.ErrUnset}},

		// QuantizeSpan: E-1 (start, length), mode, step, E-8, E-11.
		{"QuantizeSpan start before mode", func() error { return e2(duration.QuantizeSpan(duration.Position{}, v, frac(1, 8), 0)) }, duration.ErrUnset, nil},
		{"QuantizeSpan length before mode", func() error { return e2(duration.QuantizeSpan(pos(1, 3), duration.Duration{}, frac(1, 8), 0)) },
			duration.ErrUnset, nil},
		{"QuantizeSpan mode before step", func() error { return e2(duration.QuantizeSpan(pos(1, 3), v, fraction.Fraction{}, 200)) },
			duration.ErrInvalidArgument, []error{rounding.ErrUnknownMode}},
		{"QuantizeSpan zero before range", func() error {
			// Start and end both round up to 2^255, which is out of range.
			return e2(duration.QuantizeSpan(maxPos, dur(1, 2), frac(2, 1), rounding.Up))
		}, duration.ErrQuantizedToZero, nil},

		// Parse, ParsePosition: grammar of the whole text before numbers.
		{"Parse grammar before range", func() error { return e(duration.Parse(maxCompPlus + "/1x")) }, duration.ErrInvalidText, nil},
		{"Parse grammar after range", func() error { return e(duration.Parse(maxCompPlus + "/" + maxCompPlus + "/1")) }, duration.ErrInvalidText, nil},
		{"ParsePosition grammar before range", func() error { return e(duration.ParsePosition(maxCompPlus + "/0")) }, duration.ErrInvalidText, nil},

		// ParseWritten: grammar, numbers of tuplets, coefficient, value.
		{"ParseWritten grammar before range", func() error { return e(duration.ParseWritten("4[" + maxIntPlusText + ":1]x")) }, duration.ErrInvalidText, nil},
		{"ParseWritten number before coefficient", func() error { return e(duration.ParseWritten("4[" + big5 + "," + maxIntPlusText + ":1]")) },
			duration.ErrOutOfRange, []error{tuplet.ErrOutOfRange}},
		{"ParseWritten coefficient before value", func() error { return e(duration.ParseWritten("4" + strings.Repeat(".", 600) + "[" + big5 + "]")) },
			duration.ErrOutOfRange, []error{tuplet.ErrOutOfRange, fraction.ErrOutOfRange}},
		{"ParseWritten number before value", func() error {
			return e(duration.ParseWritten("4" + strings.Repeat(".", 600) + "[" + maxIntPlusText + ":1]"))
		},
			duration.ErrOutOfRange, []error{tuplet.ErrOutOfRange}},

		// ParseTied: grammar, parts from left to right, the sum.
		{"ParseTied grammar before range", func() error { return e(duration.ParseTied("4[" + maxIntPlusText + ":1]~3")) }, duration.ErrInvalidText, nil},
		{"ParseTied first part value", func() error {
			return e(duration.ParseTied("4" + strings.Repeat(".", 300) + "~4[" + maxIntPlusText + ":1]"))
		},
			duration.ErrOutOfRange, []error{fraction.ErrOutOfRange}},
		{"ParseTied first part number", func() error { return e(duration.ParseTied("4[" + maxIntPlusText + ":1]~4" + strings.Repeat(".", 300))) },
			duration.ErrOutOfRange, []error{tuplet.ErrOutOfRange}},
		{"ParseTied part before sum", func() error {
			n1Text, n2Text := strings.Repeat(maxIntText+":1,", 2)+maxIntText+":1", strings.Repeat(n2Str+":1,", 2)+n2Str+":1"

			return e(duration.ParseTied("4[" + n1Text + "]~4[" + n2Text + "]~4[" + maxIntPlusText + ":1]"))
		}, duration.ErrOutOfRange, []error{tuplet.ErrOutOfRange}},

		// UnmarshalText: a nil receiver before the grammar.
		{"UnmarshalText nil receiver", func() error {
			var target *duration.Tied

			return target.UnmarshalText([]byte("4"))
		}, duration.ErrInvalidArgument, nil},
		{"Base.UnmarshalText nil receiver", func() error {
			var target *duration.Base

			return target.UnmarshalText([]byte("x"))
		}, duration.ErrInvalidArgument, nil},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			err := fast(t, func() error { return c.run() })
			assertErr(t, err, c.want, c.causes...)
		})
	}
}
