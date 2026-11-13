package duration_test

import (
	"testing"

	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
)

// TestUnsetAccepted covers KP-42, T-9, and architecture 4.2: IsSet, Equal,
// Compare, and String accept unset values without an error.
func TestUnsetAccepted(t *testing.T) {
	t.Parallel()

	var (
		ud duration.Duration
		uw duration.Written
		ut duration.Tied
		up duration.Position
	)

	assert.False(t, ud.IsSet())
	assert.False(t, uw.IsSet())
	assert.False(t, ut.IsSet())
	assert.False(t, up.IsSet())

	assert.True(t, ud.Equal(duration.Duration{}))
	assert.True(t, uw.Equal(duration.Written{}))
	assert.True(t, ut.Equal(duration.Tied{}))
	assert.True(t, up.Equal(duration.Position{}))

	d, w, tt, p := dur(1, 4), wr(duration.Quarter, 0), tie(wr(duration.Half, 0), wr(duration.Quarter, 0)), duration.Origin()

	assert.False(t, ud.Equal(d))
	assert.False(t, d.Equal(ud))
	assert.False(t, uw.Equal(w))
	assert.False(t, w.Equal(uw))
	assert.False(t, ut.Equal(tt))
	assert.False(t, tt.Equal(ut))
	assert.False(t, up.Equal(p), "the unset position is not the origin")
	assert.False(t, p.Equal(up))

	assert.Equal(t, 0, ud.Compare(duration.Duration{}))
	assert.Equal(t, -1, ud.Compare(d))
	assert.Equal(t, 1, d.Compare(ud))
	assert.Equal(t, 0, uw.Compare(duration.Written{}))
	assert.Equal(t, -1, uw.Compare(w))
	assert.Equal(t, 1, w.Compare(uw))
	assert.Equal(t, 0, ut.Compare(duration.Tied{}))
	assert.Equal(t, -1, ut.Compare(tt))
	assert.Equal(t, 1, tt.Compare(ut))
	assert.Equal(t, 0, up.Compare(duration.Position{}))
	assert.Equal(t, -1, up.Compare(p), "the unset position precedes the origin")
	assert.Equal(t, 1, p.Compare(up))

	assert.Equal(t, "Duration(unset)", ud.String())
	assert.Equal(t, "Written(unset)", uw.String())
	assert.Equal(t, "Tied(unset)", ut.String())
	assert.Equal(t, "Position(unset)", up.String())
}

// TestUnsetRejected covers KP-42, KP-46, T-9, and architecture 4.2: every
// other operation with an unset receiver or argument returns E-1 and zero
// results.
func TestUnsetRejected(t *testing.T) {
	t.Parallel()

	var (
		ud duration.Duration
		uw duration.Written
		ut duration.Tied
		up duration.Position
	)

	d, w, tt, p := dur(1, 4), wr(duration.Quarter, 0), tie(wr(duration.Half, 0), wr(duration.Quarter, 0)), pos(1, 4)
	step, mode := frac(1, 8), rounding.HalfEven

	// Each check reports whether the other results are zero, and the error.
	checks := []struct {
		name string
		run  func() (bool, error)
	}{
		// Functions with unset arguments.
		{"BaseOf", func() (bool, error) { b, err := duration.BaseOf(ud); return b == 0, err }},
		{"KP-46 Sum", func() (bool, error) { r, err := duration.Sum(d, ud); return r == duration.Duration{}, err }},
		{"Min", func() (bool, error) { r, err := duration.Min(ud, d); return r == duration.Duration{}, err }},
		{"Max", func() (bool, error) { r, err := duration.Max(d, ud); return r == duration.Duration{}, err }},
		{"NewTied", func() (bool, error) { r, err := duration.NewTied(w, uw); return r == duration.Tied{}, err }},
		{"QuantizeSpan start", func() (bool, error) {
			s, l, err := duration.QuantizeSpan(up, d, step, mode)
			return s == duration.Position{} && l == duration.Duration{}, err
		}},
		{"QuantizeSpan length", func() (bool, error) {
			s, l, err := duration.QuantizeSpan(p, ud, step, mode)
			return s == duration.Position{} && l == duration.Duration{}, err
		}},

		// Duration.
		{"Duration.Fraction", func() (bool, error) { f, err := ud.Fraction(); return !f.IsValid(), err }},
		{"Duration.Add receiver", func() (bool, error) { r, err := ud.Add(d); return r == duration.Duration{}, err }},
		{"Duration.Add argument", func() (bool, error) { r, err := d.Add(ud); return r == duration.Duration{}, err }},
		{"Duration.Sub receiver", func() (bool, error) { r, err := ud.Sub(d); return r == duration.Duration{}, err }},
		{"Duration.Sub argument", func() (bool, error) { r, err := d.Sub(ud); return r == duration.Duration{}, err }},
		{"Duration.Mul", func() (bool, error) { r, err := ud.Mul(step); return r == duration.Duration{}, err }},
		{"Duration.MulInt", func() (bool, error) { r, err := ud.MulInt(2); return r == duration.Duration{}, err }},
		{"Duration.Div", func() (bool, error) { r, err := ud.Div(step); return r == duration.Duration{}, err }},
		{"Duration.DivInt", func() (bool, error) { r, err := ud.DivInt(2); return r == duration.Duration{}, err }},
		{"Duration.Ratio receiver", func() (bool, error) { r, err := ud.Ratio(d); return !r.IsValid(), err }},
		{"Duration.Ratio argument", func() (bool, error) { r, err := d.Ratio(ud); return !r.IsValid(), err }},
		{"Duration.QuoRem receiver", func() (bool, error) { q, r, err := ud.QuoRem(d); return q == 0 && !r.IsValid(), err }},
		{"Duration.QuoRem argument", func() (bool, error) { q, r, err := d.QuoRem(ud); return q == 0 && !r.IsValid(), err }},
		{"Duration.ToWritten", func() (bool, error) { r, err := ud.ToWritten(); return r == duration.Written{}, err }},
		{"Duration.Notate", func() (bool, error) { r, err := ud.Notate(); return r == nil, err }},
		{"Duration.NotateWith", func() (bool, error) { r, err := ud.NotateWith(duration.Breve, 0); return r == nil, err }},
		{"Duration.Quantize", func() (bool, error) { r, err := ud.Quantize(step, mode); return r == duration.Duration{}, err }},
		{"Duration.MarshalText", func() (bool, error) { r, err := ud.MarshalText(); return r == nil, err }},

		// Written.
		{"Written.Base", func() (bool, error) { r, err := uw.Base(); return r == 0, err }},
		{"Written.Dots", func() (bool, error) { r, err := uw.Dots(); return r == 0, err }},
		{"Written.Chain", func() (bool, error) { r, err := uw.Chain(); return r == nil, err }},
		{"Written.Coefficient", func() (bool, error) { r, err := uw.Coefficient(); return !r.IsValid(), err }},
		{"Written.Duration", func() (bool, error) { r, err := uw.Duration(); return r == duration.Duration{}, err }},
		{"Written.HasDots", func() (bool, error) { r, err := uw.HasDots(); return !r, err }},
		{"Written.HasTuplets", func() (bool, error) { r, err := uw.HasTuplets(); return !r, err }},
		{"Written.IsBasic", func() (bool, error) { r, err := uw.IsBasic(); return !r, err }},
		{"Written.EqualDuration receiver", func() (bool, error) { r, err := uw.EqualDuration(w); return !r, err }},
		{"Written.EqualDuration argument", func() (bool, error) { r, err := w.EqualDuration(uw); return !r, err }},
		{"Written.EqualDuration both", func() (bool, error) { r, err := uw.EqualDuration(uw); return !r, err }},
		{"Written.CompareDuration receiver", func() (bool, error) { r, err := uw.CompareDuration(w); return r == 0, err }},
		{"Written.CompareDuration argument", func() (bool, error) { r, err := w.CompareDuration(uw); return r == 0, err }},
		{"Written.WithBase", func() (bool, error) { r, err := uw.WithBase(duration.Half); return r == duration.Written{}, err }},
		{"Written.ShiftBase", func() (bool, error) { r, err := uw.ShiftBase(0); return r == duration.Written{}, err }},
		{"Written.WithDots", func() (bool, error) { r, err := uw.WithDots(1); return r == duration.Written{}, err }},
		{"Written.AddDot", func() (bool, error) { r, err := uw.AddDot(); return r == duration.Written{}, err }},
		{"Written.RemoveDot", func() (bool, error) { r, err := uw.RemoveDot(); return r == duration.Written{}, err }},
		{"Written.WithChain", func() (bool, error) { r, err := uw.WithChain(); return r == duration.Written{}, err }},
		{"Written.AddOuterTuplet", func() (bool, error) {
			r, err := uw.AddOuterTuplet(tuplet.Triplet())
			return r == duration.Written{}, err
		}},
		{"Written.AddInnerTuplet", func() (bool, error) {
			r, err := uw.AddInnerTuplet(tuplet.Triplet())
			return r == duration.Written{}, err
		}},
		{"Written.RemoveOuterTuplet", func() (bool, error) { r, err := uw.RemoveOuterTuplet(); return r == duration.Written{}, err }},
		{"Written.RemoveInnerTuplet", func() (bool, error) { r, err := uw.RemoveInnerTuplet(); return r == duration.Written{}, err }},
		{"Written.RemoveTuplets", func() (bool, error) { r, err := uw.RemoveTuplets(); return r == duration.Written{}, err }},
		{"Written.Subdivide", func() (bool, error) { c, r, err := uw.Subdivide(); return c == 0 && r == duration.Written{}, err }},
		{"Written.MarshalText", func() (bool, error) { r, err := uw.MarshalText(); return r == nil, err }},

		// Tied.
		{"Tied.Len", func() (bool, error) { r, err := ut.Len(); return r == 0, err }},
		{"Tied.Parts", func() (bool, error) { r, err := ut.Parts(); return r == nil, err }},
		{"Tied.Duration", func() (bool, error) { r, err := ut.Duration(); return r == duration.Duration{}, err }},
		{"Tied.EqualDuration receiver", func() (bool, error) { r, err := ut.EqualDuration(tt); return !r, err }},
		{"Tied.EqualDuration argument", func() (bool, error) { r, err := tt.EqualDuration(ut); return !r, err }},
		{"Tied.CompareDuration receiver", func() (bool, error) { r, err := ut.CompareDuration(tt); return r == 0, err }},
		{"Tied.CompareDuration argument", func() (bool, error) { r, err := tt.CompareDuration(ut); return r == 0, err }},
		{"Tied.Append receiver", func() (bool, error) { r, err := ut.Append(w); return r == duration.Tied{}, err }},
		{"Tied.Append argument", func() (bool, error) { r, err := tt.Append(uw); return r == duration.Tied{}, err }},
		{"Tied.Join receiver", func() (bool, error) { r, err := ut.Join(tt); return r == duration.Tied{}, err }},
		{"Tied.Join argument", func() (bool, error) { r, err := tt.Join(ut); return r == duration.Tied{}, err }},
		{"Tied.ToWritten", func() (bool, error) { r, err := ut.ToWritten(); return r == duration.Written{}, err }},
		{"Tied.MarshalText", func() (bool, error) { r, err := ut.MarshalText(); return r == nil, err }},

		// Position.
		{"Position.Fraction", func() (bool, error) { r, err := up.Fraction(); return !r.IsValid(), err }},
		{"Position.Add receiver", func() (bool, error) { r, err := up.Add(d); return r == duration.Position{}, err }},
		{"KP-46 Position.Add argument", func() (bool, error) { r, err := p.Add(ud); return r == duration.Position{}, err }},
		{"Position.Sub receiver", func() (bool, error) { r, err := up.Sub(d); return r == duration.Position{}, err }},
		{"Position.Sub argument", func() (bool, error) { r, err := p.Sub(ud); return r == duration.Position{}, err }},
		{"Position.OffsetTo receiver", func() (bool, error) { r, err := up.OffsetTo(p); return r == duration.Position{}, err }},
		{"Position.OffsetTo argument", func() (bool, error) { r, err := p.OffsetTo(up); return r == duration.Position{}, err }},
		{"Position.DurationTo receiver", func() (bool, error) { r, err := up.DurationTo(p); return r == duration.Duration{}, err }},
		{"Position.DurationTo argument", func() (bool, error) { r, err := p.DurationTo(up); return r == duration.Duration{}, err }},
		{"Position.Quantize", func() (bool, error) { r, err := up.Quantize(step, mode); return r == duration.Position{}, err }},
		{"Position.MarshalText", func() (bool, error) { r, err := up.MarshalText(); return r == nil, err }},
	}

	for _, c := range checks {
		zero, err := c.run()
		assertErr(t, err, duration.ErrUnset)
		assert.True(t, zero, "%s: zero results on failure", c.name)
	}

	// AppendText returns dst unchanged.
	for _, v := range []interface {
		AppendText([]byte) ([]byte, error)
	}{ud, uw, ut, up} {
		out, err := v.AppendText([]byte("x"))
		assertErr(t, err, duration.ErrUnset)
		assert.Equal(t, []byte("x"), out)
	}

	// Parsing rejects the special strings (KP-36).
	for _, s := range []string{"Duration(unset)", "Written(unset)", "Tied(unset)", "Position(unset)"} {
		_, err := duration.Parse(s)
		assertErr(t, err, duration.ErrInvalidText)

		_, err = duration.ParsePosition(s)
		assertErr(t, err, duration.ErrInvalidText)

		_, err = duration.ParseWritten(s)
		assertErr(t, err, duration.ErrInvalidText)

		_, err = duration.ParseTied(s)
		assertErr(t, err, duration.ErrInvalidText)
	}

}
