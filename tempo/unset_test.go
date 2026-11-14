package tempo_test

import (
	"testing"
	"time"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempo/term"

	"github.com/stretchr/testify/assert"
)

// TestUnsetValues covers С-29 and architecture 4.3, "Незаданные значения":
// IsSet, Equal (only with the unset value of the same type), Compare (before
// any set value) and String (special strings) work without an error.
func TestUnsetValues(t *testing.T) {
	t.Parallel()

	assert.False(t, tempo.Tempo{}.IsSet())
	assert.False(t, tempo.Approx{}.IsSet())
	assert.False(t, tempo.Range{}.IsSet())
	assert.False(t, tempo.Seconds{}.IsSet())
	assert.False(t, tempo.Indication{}.IsSet())

	assert.True(t, tempo.Tempo{}.Equal(tempo.Tempo{}))
	assert.False(t, tempo.Tempo{}.Equal(tq(60)))
	assert.False(t, tq(60).Equal(tempo.Tempo{}))
	assert.True(t, tempo.Approx{}.Equal(tempo.Approx{}))
	assert.False(t, tempo.Approx{}.Equal(ap(q, fi(1))))
	assert.True(t, tempo.Range{}.Equal(tempo.Range{}))
	assert.False(t, rg(q, fi(1), fi(2)).Equal(tempo.Range{}))
	assert.True(t, tempo.Seconds{}.Equal(tempo.Seconds{}))
	assert.False(t, tempo.Seconds{}.Equal(tempo.ZeroSeconds()))
	assert.True(t, tempo.Indication{}.Equal(tempo.Indication{}))
	assert.False(t, tempo.Indication{}.Equal(tempo.MustIndicationOfTempo(tq(1))))

	assert.Equal(t, 0, tempo.Tempo{}.Compare(tempo.Tempo{}))
	assert.Equal(t, -1, tempo.Tempo{}.Compare(tm(q, invMaxF())))
	assert.Equal(t, 1, tm(q, invMaxF()).Compare(tempo.Tempo{}))
	assert.Equal(t, 0, tempo.Approx{}.Compare(tempo.Approx{}))
	assert.Equal(t, -1, tempo.Approx{}.Compare(ap(q, invMaxF())))
	assert.Equal(t, 1, ap(q, invMaxF()).Compare(tempo.Approx{}))
	assert.Equal(t, 0, tempo.Range{}.Compare(tempo.Range{}))
	assert.Equal(t, -1, tempo.Range{}.Compare(rg(q, invMaxF(), fi(1))))
	assert.Equal(t, 1, rg(q, invMaxF(), fi(1)).Compare(tempo.Range{}))
	assert.Equal(t, 0, tempo.Seconds{}.Compare(tempo.Seconds{}))
	assert.Equal(t, -1, tempo.Seconds{}.Compare(tempo.ZeroSeconds()))
	assert.Equal(t, 1, tempo.ZeroSeconds().Compare(tempo.Seconds{}))

	assert.Equal(t, "Tempo(unset)", tempo.Tempo{}.String())
	assert.Equal(t, "Approx(unset)", tempo.Approx{}.String())
	assert.Equal(t, "Range(unset)", tempo.Range{}.String())
	assert.Equal(t, "Seconds(unset)", tempo.Seconds{}.String())
	assert.Empty(t, tempo.Indication{}.String(), "КР-TP-16")

	for _, s := range []string{"Tempo(unset)", "Approx(unset)", "Range(unset)", "Seconds(unset)"} {
		_, err := tempo.Parse(s)
		textClass.check(t, err)
		_, err = tempo.ParseApprox(s)
		textClass.check(t, err)
		_, err = tempo.ParseRange(s)
		textClass.check(t, err)
		_, err = tempo.ParseSeconds(s)
		textClass.check(t, err)
		_, err = tempo.ParseIndication(s, term.EnglishWikipedia())
		textClass.check(t, err)
	}

	_, err := tempo.ParseIndication("", term.EnglishWikipedia())
	textClass.check(t, err)
}

// TestUnsetOperations covers С-32 and architecture 4.3: every other
// operation with an unset receiver or an unset argument of tempo is Е-1 with
// the zero results of 4.3, MarshalText and AppendText included.
func TestUnsetOperations(t *testing.T) {
	t.Parallel()

	var (
		x  tempo.Tempo
		a  tempo.Approx
		r  tempo.Range
		s  tempo.Seconds
		i  tempo.Indication
		ok = tq(60)
	)

	type c = func() (any, error)

	tests := []struct {
		name string
		call c
		zero any
	}{
		{"Tempo.Unit", func() (any, error) { return x.Unit() }, duration.Written{}},
		{"Tempo.UnitDuration", func() (any, error) { return x.UnitDuration() }, duration.Duration{}},
		{"Tempo.BPM", func() (any, error) { return x.BPM() }, fraction.Fraction{}},
		{"Tempo.Speed", func() (any, error) { return x.Speed() }, fraction.Fraction{}},
		{"Tempo.QuartersPerMinute", func() (any, error) { return x.QuartersPerMinute() }, fraction.Fraction{}},
		{"Tempo.EqualSpeed", func() (any, error) { return x.EqualSpeed(ok) }, false},
		{"Tempo.EqualSpeed argument", func() (any, error) { return ok.EqualSpeed(x) }, false},
		{"Tempo.CompareSpeed", func() (any, error) { return x.CompareSpeed(ok) }, 0},
		{"Tempo.CompareSpeed argument", func() (any, error) { return ok.CompareSpeed(x) }, 0},
		{"Tempo.SpeedRatio", func() (any, error) { return x.SpeedRatio(ok) }, fraction.Fraction{}},
		{"Tempo.SpeedRatio argument", func() (any, error) { return ok.SpeedRatio(x) }, fraction.Fraction{}},
		{"Tempo.WithUnit", func() (any, error) { return x.WithUnit(q) }, tempo.Tempo{}},
		{"Tempo.WithBPM", func() (any, error) { return x.WithBPM(fi(1)) }, tempo.Tempo{}},
		{"Tempo.Mul", func() (any, error) { return x.Mul(fi(2)) }, tempo.Tempo{}},
		{"Tempo.Double", func() (any, error) { return x.Double() }, tempo.Tempo{}},
		{"Tempo.Halve", func() (any, error) { return x.Halve() }, tempo.Tempo{}},
		{"Tempo.Modulate", func() (any, error) { return x.Modulate(q, dur(1, 4)) }, tempo.Tempo{}},
		{"Tempo.ModulationTo", func() (any, error) { return x.ModulationTo(ok) }, duration.Duration{}},
		{"Tempo.ModulationTo argument", func() (any, error) { return ok.ModulationTo(x) }, duration.Duration{}},
		{"Tempo.Seconds", func() (any, error) { return x.Seconds(dur(1, 4)) }, tempo.Seconds{}},
		{"Tempo.UnitSeconds", func() (any, error) { return x.UnitSeconds() }, tempo.Seconds{}},
		{"Tempo.UnitsPerSecond", func() (any, error) { return x.UnitsPerSecond() }, fraction.Fraction{}},
		{"Tempo.DurationIn", func() (any, error) { return x.DurationIn(sec(1, 1)) }, duration.Duration{}},
		{"Tempo.DurationIn argument", func() (any, error) { return ok.DurationIn(s) }, duration.Duration{}},
		{"Tempo.UnitsIn", func() (any, error) { return x.UnitsIn(sec(1, 1)) }, fraction.Fraction{}},
		{"Tempo.UnitsIn argument", func() (any, error) { return ok.UnitsIn(s) }, fraction.Fraction{}},
		{"Tempo.SecondsOfUnits", func() (any, error) { return x.SecondsOfUnits(fi(1)) }, tempo.Seconds{}},
		{"Tempo.BarUnits", func() (any, error) { return x.BarUnits(dur(1, 1)) }, fraction.Fraction{}},
		{"Tempo.Bars", func() (any, error) { return x.Bars(dur(1, 1), sec(1, 1)) }, fraction.Fraction{}},
		{"Tempo.SecondsOfBars", func() (any, error) { return x.SecondsOfBars(dur(1, 1), fi(1)) }, tempo.Seconds{}},
		{"Tempo.BarFor", func() (any, error) {
			rr, l, err := x.BarFor(fi(1), sec(1, 1))
			return []any{rr, l}, err
		}, []any{fraction.Fraction{}, duration.Duration{}}},
		{"Tempo.BarFor argument", func() (any, error) {
			rr, l, err := ok.BarFor(fi(1), s)
			return []any{rr, l}, err
		}, []any{fraction.Fraction{}, duration.Duration{}}},
		{"Tempo.BarsPerMinute", func() (any, error) { return x.BarsPerMinute(dur(1, 1)) }, fraction.Fraction{}},
		{"Tempo.Quantize", func() (any, error) { return x.Quantize(fi(1), down) }, tempo.Tempo{}},
		{"Tempo.SnapToScale", func() (any, error) { return x.SnapToScale(down) }, tempo.Tempo{}},
		{"Tempo.TermsIn", func() (any, error) { return x.TermsIn(term.Default()) }, []term.Term(nil)},
		{"Tempo.MarshalText", func() (any, error) { return x.MarshalText() }, []byte(nil)},
		{"Seconds.Fraction", func() (any, error) { return s.Fraction() }, fraction.Fraction{}},
		{"Seconds.Minutes", func() (any, error) { return s.Minutes() }, fraction.Fraction{}},
		{"Seconds.Add", func() (any, error) { return s.Add(sec(1, 1)) }, tempo.Seconds{}},
		{"Seconds.Add argument", func() (any, error) { return sec(1, 1).Add(s) }, tempo.Seconds{}},
		{"Seconds.Sub", func() (any, error) { return s.Sub(sec(1, 1)) }, tempo.Seconds{}},
		{"Seconds.Sub argument", func() (any, error) { return sec(1, 1).Sub(s) }, tempo.Seconds{}},
		{"Seconds.TimeDuration", func() (any, error) { return s.TimeDuration(down) }, time.Duration(0)},
		{"Seconds.MarshalText", func() (any, error) { return s.MarshalText() }, []byte(nil)},
		{"FromSeconds", func() (any, error) { return tempo.FromSeconds(q, dur(1, 4), s) }, tempo.Tempo{}},
		{"FromBars", func() (any, error) { return tempo.FromBars(q, dur(1, 4), fi(1), s) }, tempo.Tempo{}},
		{"SolveBars", func() (any, error) { return tempo.SolveBars(fi(1), fi(1), s) }, fraction.Fraction{}},
		{"SolveBarUnits", func() (any, error) { return tempo.SolveBarUnits(fi(1), fi(1), s) }, fraction.Fraction{}},
		{"SolveBPM", func() (any, error) { return tempo.SolveBPM(fi(1), fi(1), s) }, fraction.Fraction{}},
		{"Approx.Unit", func() (any, error) { return a.Unit() }, duration.Written{}},
		{"Approx.UnitDuration", func() (any, error) { return a.UnitDuration() }, duration.Duration{}},
		{"Approx.BPM", func() (any, error) { return a.BPM() }, fraction.Fraction{}},
		{"Approx.Speed", func() (any, error) { return a.Speed() }, fraction.Fraction{}},
		{"Approx.QuartersPerMinute", func() (any, error) { return a.QuartersPerMinute() }, fraction.Fraction{}},
		{"Approx.EqualSpeed", func() (any, error) { return a.EqualSpeed(ap(q, fi(1))) }, false},
		{"Approx.EqualSpeed argument", func() (any, error) { return ap(q, fi(1)).EqualSpeed(a) }, false},
		{"Approx.CompareSpeed", func() (any, error) { return a.CompareSpeed(ap(q, fi(1))) }, 0},
		{"Approx.WithUnit", func() (any, error) { return a.WithUnit(q) }, tempo.Approx{}},
		{"Approx.Tempo", func() (any, error) { return a.Tempo() }, tempo.Tempo{}},
		{"Approx.TermsIn", func() (any, error) { return a.TermsIn(term.Default()) }, []term.Term(nil)},
		{"Approx.MarshalText", func() (any, error) { return a.MarshalText() }, []byte(nil)},
		{"Range.Unit", func() (any, error) { return r.Unit() }, duration.Written{}},
		{"Range.UnitDuration", func() (any, error) { return r.UnitDuration() }, duration.Duration{}},
		{"Range.Lower", func() (any, error) { return r.Lower() }, fraction.Fraction{}},
		{"Range.Upper", func() (any, error) { return r.Upper() }, fraction.Fraction{}},
		{"Range.LowerTempo", func() (any, error) { return r.LowerTempo() }, tempo.Tempo{}},
		{"Range.UpperTempo", func() (any, error) { return r.UpperTempo() }, tempo.Tempo{}},
		{"Range.MidTempo", func() (any, error) { return r.MidTempo() }, tempo.Tempo{}},
		{"Range.EqualSpeed", func() (any, error) { return r.EqualSpeed(rg(q, fi(1), fi(2))) }, false},
		{"Range.EqualSpeed argument", func() (any, error) { return rg(q, fi(1), fi(2)).EqualSpeed(r) }, false},
		{"Range.WithUnit", func() (any, error) { return r.WithUnit(q) }, tempo.Range{}},
		{"Range.Contains", func() (any, error) { return r.Contains(fi(1)) }, false},
		{"Range.ContainsTempo", func() (any, error) { return r.ContainsTempo(ok) }, false},
		{"Range.ContainsTempo argument", func() (any, error) { return rg(q, fi(1), fi(2)).ContainsTempo(x) }, false},
		{"Range.MarshalText", func() (any, error) { return r.MarshalText() }, []byte(nil)},
		{"Indication.Kind", func() (any, error) { return i.Kind() }, tempo.NoMark},
		{"Indication.WithTerm", func() (any, error) { return i.WithTerm(term.Default().MustParse("Largo")) }, tempo.Indication{}},
		{"Indication.WithTempo", func() (any, error) { return i.WithTempo(ok) }, tempo.Indication{}},
		{"Indication.WithApprox", func() (any, error) { return i.WithApprox(ap(q, fi(1))) }, tempo.Indication{}},
		{"Indication.WithRange", func() (any, error) { return i.WithRange(rg(q, fi(1), fi(2))) }, tempo.Indication{}},
		{"Indication.WithoutTerm", func() (any, error) { return i.WithoutTerm() }, tempo.Indication{}},
		{"Indication.WithoutMarking", func() (any, error) { return i.WithoutMarking() }, tempo.Indication{}},
		{"Indication.MarshalText", func() (any, error) { return i.MarshalText() }, []byte(nil)},
		{"Indication.Term", func() (any, error) { w, b, err := i.Term(); return []any{w, b}, err }, []any{term.Term{}, false}},
		{"Indication.Tempo", func() (any, error) { v, b, err := i.Tempo(); return []any{v, b}, err }, []any{tempo.Tempo{}, false}},
		{"Indication.Approx", func() (any, error) { v, b, err := i.Approx(); return []any{v, b}, err }, []any{tempo.Approx{}, false}},
		{"Indication.Range", func() (any, error) { v, b, err := i.Range(); return []any{v, b}, err }, []any{tempo.Range{}, false}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.call()
			unsetClass.check(t, err)
			assert.Equal(t, tt.zero, got)
		})
	}

	t.Run("AppendText returns dst unchanged", func(t *testing.T) {
		t.Parallel()

		for name, f := range map[string]func([]byte) ([]byte, error){
			"Tempo": x.AppendText, "Approx": a.AppendText, "Range": r.AppendText, "Seconds": s.AppendText,
			"Indication": i.AppendText,
		} {
			dst := []byte("keep")
			got, err := f(dst)
			unsetClass.check(t, err)
			assert.Equal(t, "keep", string(got), name)
		}
	})
}

// TestOtherPackagesUnset covers С-31 and architecture 4.3: an unset value of
// another package on input is Е-2 without a dependency sentinel.
func TestOtherPackagesUnset(t *testing.T) {
	t.Parallel()

	ok := tq(60)

	tests := []struct {
		name string
		call func() error
	}{
		{"WithUnit", func() error { _, err := ok.WithUnit(duration.Written{}); return err }},
		{"WithBPM", func() error { _, err := ok.WithBPM(fraction.Fraction{}); return err }},
		{"Mul", func() error { _, err := ok.Mul(fraction.Fraction{}); return err }},
		{"Modulate unit", func() error { _, err := ok.Modulate(duration.Written{}, dur(1, 4)); return err }},
		{"Modulate old", func() error { _, err := ok.Modulate(q, duration.Duration{}); return err }},
		{"Seconds", func() error { _, err := ok.Seconds(duration.Duration{}); return err }},
		{"SecondsOfUnits", func() error { _, err := ok.SecondsOfUnits(fraction.Fraction{}); return err }},
		{"BarUnits", func() error { _, err := ok.BarUnits(duration.Duration{}); return err }},
		{"Bars", func() error { _, err := ok.Bars(duration.Duration{}, sec(1, 1)); return err }},
		{"SecondsOfBars length", func() error { _, err := ok.SecondsOfBars(duration.Duration{}, fi(1)); return err }},
		{"SecondsOfBars bars", func() error { _, err := ok.SecondsOfBars(dur(1, 1), fraction.Fraction{}); return err }},
		{"BarFor", func() error { _, _, err := ok.BarFor(fraction.Fraction{}, sec(1, 1)); return err }},
		{"BarsPerMinute", func() error { _, err := ok.BarsPerMinute(duration.Duration{}); return err }},
		{"Quantize", func() error { _, err := ok.Quantize(fraction.Fraction{}, down); return err }},
		{"TermsIn", func() error { _, err := ok.TermsIn(term.List{}); return err }},
		{"FromSeconds unit", func() error { _, err := tempo.FromSeconds(duration.Written{}, dur(1, 4), sec(1, 1)); return err }},
		{"FromSeconds duration", func() error { _, err := tempo.FromSeconds(q, duration.Duration{}, sec(1, 1)); return err }},
		{"FromBars", func() error { _, err := tempo.FromBars(q, duration.Duration{}, fi(1), sec(1, 1)); return err }},
		{"FromBarsPerMinute", func() error { _, err := tempo.FromBarsPerMinute(q, dur(1, 1), fraction.Fraction{}); return err }},
		{"SolveBars", func() error { _, err := tempo.SolveBars(fraction.Fraction{}, fi(1), sec(1, 1)); return err }},
		{"SolveBarUnits", func() error { _, err := tempo.SolveBarUnits(fi(1), fraction.Fraction{}, sec(1, 1)); return err }},
		{"SolveBPM", func() error { _, err := tempo.SolveBPM(fi(1), fraction.Fraction{}, sec(1, 1)); return err }},
		{"SolveSeconds", func() error { _, err := tempo.SolveSeconds(fraction.Fraction{}, fi(1), fi(1)); return err }},
		{"Approx.WithUnit", func() error { _, err := ap(q, fi(1)).WithUnit(duration.Written{}); return err }},
		{"Approx.TermsIn", func() error { _, err := ap(q, fi(1)).TermsIn(term.List{}); return err }},
		{"Range.WithUnit", func() error { _, err := rg(q, fi(1), fi(2)).WithUnit(duration.Written{}); return err }},
		{"RangeOf", func() error { _, err := tempo.RangeOf(term.Term{}, q); return err }},
		{"IsScaleDivision", func() error { _, err := tempo.IsScaleDivision(fraction.Fraction{}); return err }},
		{"IndicationOf", func() error { _, err := tempo.IndicationOf(term.Term{}); return err }},
		{"WithTerm", func() error { _, err := tempo.MustIndicationOfTempo(ok).WithTerm(term.Term{}); return err }},
		{"ParseIndication", func() error { _, err := tempo.ParseIndication("4=60", term.List{}); return err }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			argClass.check(t, tt.call())
		})
	}
}
