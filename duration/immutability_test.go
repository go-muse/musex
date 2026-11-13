package duration_test

import (
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
)

// TestInputSlicesNotRetained covers architecture 3.2: variadic lists are
// only read during the call; changing them afterwards does not change any
// value, and the call does not change them.
func TestInputSlicesNotRetained(t *testing.T) {
	t.Parallel()

	triplet, quintuplet := tuplet.Triplet(), tuplet.Quintuplet()

	chain := []tuplet.Tuplet{triplet, quintuplet}
	w := duration.MustNewWritten(duration.Quarter, 1, chain...)
	chain[0] = tuplet.Septuplet()
	assert.Equal(t, []tuplet.Tuplet{triplet, quintuplet}, w.MustChain(), "NewWritten")

	chain = []tuplet.Tuplet{triplet}
	w2 := wr(duration.Half, 0).MustWithChain(chain...)
	chain[0] = quintuplet
	assert.Equal(t, []tuplet.Tuplet{triplet}, w2.MustChain(), "WithChain")

	chain = []tuplet.Tuplet{triplet}
	single := dur(1, 3).MustToWritten(chain...)
	parts := dur(5, 12).MustNotate(chain...)
	partsWith := dur(5, 12).MustNotateWith(duration.Whole, 1, chain...)
	chain[0] = quintuplet
	assert.Equal(t, wr(duration.Half, 0, triplet), single, "ToWritten")
	assert.Equal(t, []tuplet.Tuplet{triplet}, parts[0].MustChain(), "Notate")
	assert.Equal(t, []tuplet.Tuplet{triplet}, partsWith[0].MustChain(), "NotateWith")

	list := []duration.Written{wr(duration.Half, 0), wr(duration.Quarter, 0)}
	tt := duration.MustNewTied(list...)
	list[0] = wr(duration.Whole, 0)
	assert.Equal(t, []duration.Written{wr(duration.Half, 0), wr(duration.Quarter, 0)}, tt.MustParts(), "NewTied")

	values := []duration.Duration{dur(3, 8), dur(1, 8), dur(1, 4)}
	original := []duration.Duration{dur(3, 8), dur(1, 8), dur(1, 4)}
	sum := duration.MustSum(values...)
	minimum := duration.MustMin(values...)
	maximum := duration.MustMax(values...)
	assert.Equal(t, original, values, "Sum, Min, and Max do not reorder or change the list")

	values[0] = dur(9, 1)
	assert.Equal(t, dur(3, 4), sum)
	assert.Equal(t, dur(1, 8), minimum)
	assert.Equal(t, dur(3, 8), maximum)

	// Failing calls do not change their lists either.
	failing := []duration.Duration{dur(3, 8), {}, dur(1, 4)}
	_, _ = duration.Sum(failing...)
	_, _ = duration.Min(failing...)
	assert.Equal(t, []duration.Duration{dur(3, 8), {}, dur(1, 4)}, failing)

	badChain := []tuplet.Tuplet{triplet, {}}
	_, _ = duration.NewWritten(duration.Quarter, 0, badChain...)
	_, _ = dur(1, 3).Notate(badChain...)
	assert.Equal(t, []tuplet.Tuplet{triplet, {}}, badChain)
}

// TestResultsOwned covers architecture 3.2: result slices are new on every
// call and changing them does not affect values or later calls.
func TestResultsOwned(t *testing.T) {
	t.Parallel()

	w := wr(duration.Quarter, 0, tuplet.Triplet())
	c1 := w.MustChain()
	c2 := w.MustChain()
	c1[0] = tuplet.Septuplet()
	assert.Equal(t, tuplet.Triplet(), c2[0], "two results do not share storage")
	assert.Equal(t, []tuplet.Tuplet{tuplet.Triplet()}, w.MustChain())

	tt := tie(wr(duration.Half, 0), wr(duration.Quarter, 0))
	p1 := tt.MustParts()
	p1[1] = wr(duration.Whole, 0)
	p1 = append(p1, wr(duration.Whole, 0))
	assert.Len(t, p1, 3)
	assert.Equal(t, 2, tt.MustLen())
	assert.Equal(t, dur(3, 4), tt.MustDuration())

	b := duration.Bases()
	b[5] = duration.Base(0)
	assert.Equal(t, duration.Quarter, duration.Bases()[5])

	text, err := dur(3, 8).MarshalText()
	assert.NoError(t, err)
	text[0] = '9'
	assert.Equal(t, "3/8", dur(3, 8).String())

	text2, err := dur(3, 8).MarshalText()
	assert.NoError(t, err)
	assert.Equal(t, "3/8", string(text2), "MarshalText returns a new slice")
}

// TestValuesUnchanged covers ID-7, KP-43, and architecture 3.2: no
// operation changes its receiver or arguments, on success or on failure.
func TestValuesUnchanged(t *testing.T) {
	t.Parallel()

	d, d0 := dur(5, 16), dur(5, 16)
	o, o0 := dur(1, 4), dur(1, 4)
	w, w0 := wr(duration.Quarter, 1, tuplet.Triplet()), wr(duration.Quarter, 1, tuplet.Triplet())
	tt, tt0 := tie(wr(duration.Half, 0), wr(duration.Quarter, 0)), tie(wr(duration.Half, 0), wr(duration.Quarter, 0))
	p, p0 := pos(7, 24), pos(7, 24)
	f, f0 := frac(6, 16), frac(6, 16)
	step, step0 := frac(2, 16), frac(2, 16)
	tup, tup0 := tp(6, 4), tp(6, 4)
	b, b0 := duration.Half, duration.Half

	ops := []func(){
		func() { _, _ = d.Add(o) }, func() { _, _ = d.Sub(o) }, func() { _, _ = o.Sub(d) },
		func() { _, _ = d.Mul(f) }, func() { _, _ = d.Div(f) }, func() { _, _ = d.MulInt(3) }, func() { _, _ = d.DivInt(3) },
		func() { _, _ = d.Ratio(o) }, func() { _, _, _ = d.QuoRem(o) }, func() { _, _ = d.ToWritten(tup) }, func() { _, _ = d.ToWritten() },
		func() { _, _ = d.Notate(tup) }, func() { _, _ = d.NotateWith(b, 1, tup) }, func() { _, _ = d.Quantize(step, rounding.HalfEven) },
		func() { _, _ = d.Quantize(step, rounding.Mode(0)) }, func() { _, _ = d.MarshalText() }, func() { _ = d.String() },
		func() { _, _ = w.WithBase(b) }, func() { _, _ = w.ShiftBase(1) }, func() { _, _ = w.ShiftBase(100) }, func() { _, _ = w.WithDots(3) },
		func() { _, _ = w.AddDot() }, func() { _, _ = w.RemoveDot() }, func() { _, _ = w.WithChain(tup) }, func() { _, _ = w.AddOuterTuplet(tup) },
		func() { _, _ = w.AddInnerTuplet(tup) }, func() { _, _ = w.RemoveOuterTuplet() }, func() { _, _ = w.RemoveInnerTuplet() },
		func() { _, _ = w.RemoveTuplets() }, func() { _, _, _ = w.Subdivide() }, func() { _, _ = w.EqualDuration(w) }, func() { _, _ = w.Chain() },
		func() { _, _ = tt.Append(w) }, func() { _, _ = tt.Join(tt) }, func() { _, _ = tt.ToWritten(tup) }, func() { _, _ = tt.Parts() },
		func() { _, _ = p.Add(o) }, func() { _, _ = p.Sub(d) }, func() { _, _ = p.OffsetTo(p) }, func() { _, _ = p.DurationTo(p) },
		func() { _, _ = p.Quantize(step, rounding.Up) }, func() { _, _, _ = duration.QuantizeSpan(p, d, step, rounding.HalfEven) },
		func() { _, _ = duration.NewWritten(b, 2, tup) }, func() { _, _ = duration.NewTied(w, w) }, func() { _, _ = b.Shift(2) },
		func() { _, _ = duration.FromFraction(f) }, func() { _, _ = duration.PositionFromFraction(f) }, func() { _, _ = b.Written() },
		func() { _, _ = d.AppendText(nil) }, func() { _, _ = w.AppendText(nil) }, func() { _, _ = tt.AppendText(nil) },
		func() { _, _ = p.AppendText(nil) }, func() { _, _ = b.AppendText(nil) },
		func() { capturePanic(func() { w.MustRemoveTuplets().MustRemoveDot().MustRemoveDot() }) },
		func() { capturePanic(func() { d.MustToWritten() }) },
	}

	for _, op := range ops {
		op()
	}

	assert.Equal(t, d0, d)
	assert.Equal(t, o0, o)
	assert.Equal(t, w0, w)
	assert.Equal(t, tt0, tt)
	assert.Equal(t, p0, p)
	assert.True(t, f.EqualExact(f0), "the fraction keeps its pair")
	assert.True(t, step.EqualExact(step0))
	assert.Equal(t, tup0, tup)
	assert.Equal(t, b0, b)
	assert.Equal(t, fraction.MustNew(6, 16), f)
}

// TestUnmarshalTextInput covers architecture 3.2: UnmarshalText neither
// changes nor retains its text.
func TestUnmarshalTextInput(t *testing.T) {
	t.Parallel()

	text := []byte("4.[3:2]~8")

	var tt duration.Tied
	assert.NoError(t, tt.UnmarshalText(text))
	assert.Equal(t, "4.[3:2]~8", string(text))

	copy(text, "2~4~8~16")
	assert.Equal(t, "4.[3:2]~8", tt.String(), "the value does not share the text")

	durText := []byte("6/16")

	var d duration.Duration
	assert.NoError(t, d.UnmarshalText(durText))
	durText[0] = '1'
	assert.Equal(t, dur(3, 8), d)
	assert.Equal(t, "1/16", string(durText))

	var w duration.Written
	bad := []byte("4[]")
	assert.Error(t, w.UnmarshalText(bad))
	assert.Equal(t, "4[]", string(bad))
}
