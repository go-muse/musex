package tempo_test

import (
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempo/term"
)

// TestCheckOrder covers architecture 4.3, "Порядок проверок": for inputs
// with two or more violations the first by the row of the operation is
// returned. Each case names the expected winner.
func TestCheckOrder(t *testing.T) {
	t.Parallel()

	var (
		unsetT tempo.Tempo
		unsetS tempo.Seconds
		unsetR tempo.Range
		unsetI tempo.Indication
		nofr   fraction.Fraction
		nodur  duration.Duration
		nowr   duration.Written
	)

	bigT := tm(q, maxF())
	zero := tempo.ZeroSeconds()
	m60 := sec(60, 1)

	tests := []struct {
		name string
		call func() error
		want errClass
	}{
		{"WithBPM Е-1 before Е-2", func() error { _, err := unsetT.WithBPM(nofr); return err }, unsetClass},
		{"WithBPM Е-2 before Е-3", func() error { _, err := tq(1).WithBPM(nofr); return err }, argClass},
		{"WithBPM Е-3", func() error { _, err := tq(1).WithBPM(fi(-1)); return err }, noTempoClass},
		{"EqualSpeed both unset", func() error { _, err := unsetT.EqualSpeed(unsetT); return err }, unsetClass},
		{"CompareSpeed argument unset", func() error { _, err := bigT.CompareSpeed(unsetT); return err }, unsetClass},
		{"SpeedRatio Е-1 before Е-11", func() error { _, err := tm(breve, maxF()).SpeedRatio(unsetT); return err }, unsetClass},
		{"SpeedRatio Е-11", func() error { _, err := tm(breve, maxF()).SpeedRatio(tm(q, invMaxF())); return err }, fracRange},
		{"ModulationTo Е-1 before Е-11", func() error { _, err := unsetT.ModulationTo(tm(q, invMaxF())); return err }, unsetClass},
		{"ModulationTo Е-11", func() error { _, err := bigT.ModulationTo(tm(q, invMaxF())); return err }, fracRange},
		{"Seconds.Add Е-1 before Е-11", func() error { _, err := secF(maxF()).Add(unsetS); return err }, unsetClass},
		{"WithUnit Е-1 before Е-2", func() error { _, err := unsetT.WithUnit(nowr); return err }, unsetClass},
		{"WithUnit Е-2 before Е-11", func() error { _, err := bigT.WithUnit(nowr); return err }, argClass},
		{"WithUnit Е-11", func() error { _, err := bigT.WithUnit(e); return err }, fracRange},
		{"Tempo.Seconds Е-1 before Е-2", func() error { _, err := unsetT.Seconds(nodur); return err }, unsetClass},
		{"Tempo.Seconds Е-2", func() error { _, err := tm(q, invMaxF()).Seconds(nodur); return err }, argClass},
		{"Tempo.Seconds Е-11", func() error { _, err := tm(q, invMaxF()).Seconds(dur(1, 1)); return err }, fracRange},
		{"BarUnits method Е-1 before Е-2", func() error { _, err := unsetT.BarUnits(nodur); return err }, unsetClass},
		{"BarsPerMinute Е-1 before Е-2", func() error { _, err := unsetT.BarsPerMinute(nodur); return err }, unsetClass},
		{"Mul Е-1 before Е-2", func() error { _, err := unsetT.Mul(fi(0)); return err }, unsetClass},
		{"Mul Е-2 before Е-11", func() error { _, err := bigT.Mul(fi(0)); return err }, argClass},
		{"Mul Е-11", func() error { _, err := bigT.Mul(fi(2)); return err }, fracRange},
		{"С-43 Double Е-11", func() error { _, err := bigT.Double(); return err }, fracRange},
		{"Halve Е-11", func() error { _, err := tm(q, invMaxF()).Halve(); return err }, fracRange},
		{"Modulate Е-1 before Е-2", func() error { _, err := unsetT.Modulate(nowr, nodur); return err }, unsetClass},
		{"Modulate Е-2 unit before old", func() error { _, err := bigT.Modulate(nowr, nodur); return err }, argClass},
		{"Modulate Е-2 old before Е-11", func() error { _, err := bigT.Modulate(q, nodur); return err }, argClass},
		{"Modulate Е-11", func() error { _, err := bigT.Modulate(q, dur(1, 8)); return err }, fracRange},
		{"DurationIn Е-1 receiver", func() error { _, err := unsetT.DurationIn(zero); return err }, unsetClass},
		{"DurationIn Е-1 time", func() error { _, err := tq(1).DurationIn(unsetS); return err }, unsetClass},
		{"DurationIn Е-4 before Е-11", func() error { _, err := bigT.DurationIn(zero); return err }, noSolution},
		{"DurationIn Е-11", func() error { _, err := bigT.DurationIn(secF(maxF())); return err }, fracRange},
		{"UnitsIn Е-1", func() error { _, err := unsetT.UnitsIn(unsetS); return err }, unsetClass},
		{"UnitsIn Е-11", func() error { _, err := bigT.UnitsIn(secF(maxF())); return err }, fracRange},
		{"SecondsOfUnits Е-1 before Е-2", func() error { _, err := unsetT.SecondsOfUnits(nofr); return err }, unsetClass},
		{"SecondsOfUnits Е-2 before Е-4", func() error { _, err := tq(1).SecondsOfUnits(nofr); return err }, argClass},
		{"SecondsOfUnits Е-4 before Е-11", func() error { _, err := tm(q, invMaxF()).SecondsOfUnits(fi(-1)); return err }, noSolution},
		{"SecondsOfUnits Е-11", func() error { _, err := tm(q, invMaxF()).SecondsOfUnits(maxF()); return err }, fracRange},
		{"UnitSeconds Е-11", func() error { _, err := tm(q, invMaxF()).UnitSeconds(); return err }, fracRange},
		{"UnitsPerSecond Е-11", func() error { _, err := tm(q, invMaxF()).UnitsPerSecond(); return err }, fracRange},
		{"QuartersPerMinute Е-11", func() error { _, err := tm(h, maxF()).QuartersPerMinute(); return err }, fracRange},
		{"FromSeconds Е-1 before Е-2", func() error { _, err := tempo.FromSeconds(nowr, nodur, unsetS); return err }, unsetClass},
		{"FromSeconds Е-2 unit", func() error { _, err := tempo.FromSeconds(nowr, nodur, zero); return err }, argClass},
		{"FromSeconds Е-2 duration before Е-4", func() error { _, err := tempo.FromSeconds(q, nodur, zero); return err }, argClass},
		{"FromSeconds Е-11", func() error {
			_, err := tempo.FromSeconds(q, durF(maxF()), secF(invMaxF()))
			return err
		}, fracRange},
		{"Bars Е-1 time before Е-2", func() error { _, err := tq(1).Bars(nodur, unsetS); return err }, unsetClass},
		{"Bars Е-2 before Е-4", func() error { _, err := tq(1).Bars(nodur, zero); return err }, argClass},
		{"SecondsOfBars Е-2 length before bars", func() error { _, err := tq(1).SecondsOfBars(nodur, fi(-1)); return err }, argClass},
		{"SecondsOfBars Е-2 bars before Е-4", func() error { _, err := tq(1).SecondsOfBars(dur(1, 1), nofr); return err }, argClass},
		{"BarFor Е-1 before Е-2", func() error { _, _, err := tq(1).BarFor(nofr, unsetS); return err }, unsetClass},
		{"BarFor Е-2 before Е-4", func() error { _, _, err := tq(1).BarFor(nofr, zero); return err }, argClass},
		{"BarFor Е-4 bars", func() error { _, _, err := tq(1).BarFor(fi(-1), m60); return err }, noSolution},
		{"BarFor Е-4 time", func() error { _, _, err := tq(1).BarFor(fi(1), zero); return err }, noSolution},
		{"BarFor Е-4 before Е-11", func() error { _, _, err := bigT.BarFor(fi(0), secF(maxF())); return err }, noSolution},
		{"FromBars Е-1 before Е-2", func() error { _, err := tempo.FromBars(nowr, nodur, nofr, unsetS); return err }, unsetClass},
		{"FromBars Е-2 unit before Е-4", func() error { _, err := tempo.FromBars(nowr, dur(1, 1), fi(0), zero); return err }, argClass},
		{"FromBars Е-2 length", func() error { _, err := tempo.FromBars(q, nodur, fi(0), zero); return err }, argClass},
		{"FromBars Е-2 bars", func() error { _, err := tempo.FromBars(q, dur(1, 1), nofr, zero); return err }, argClass},
		{"FromBarsPerMinute Е-2 before Е-4", func() error { _, err := tempo.FromBarsPerMinute(nowr, nodur, fi(0)); return err }, argClass},
		{"FromBarsPerMinute Е-4 before Е-11", func() error {
			_, err := tempo.FromBarsPerMinute(q, durF(maxF()), fi(-1))
			return err
		}, noSolution},
		{"BarUnits function Е-2 before Е-11", func() error { _, err := tempo.BarUnits(durF(maxF()), nodur); return err }, argClass},
		{"BarLength Е-2 first before r", func() error { _, err := tempo.BarLength(nodur, nofr); return err }, argClass},
		{"BarLength Е-2 r before Е-4", func() error { _, err := tempo.BarLength(nodur, fi(0)); return err }, argClass},
		{"UnitLength Е-2 before Е-4", func() error { _, err := tempo.UnitLength(dur(1, 2), nofr); return err }, argClass},
		{"UnitLength Е-4 before Е-11", func() error { _, err := tempo.UnitLength(durF(invMaxF()), fi(0)); return err }, noSolution},
		{"SolveBars Е-1 before Е-2", func() error { _, err := tempo.SolveBars(nofr, nofr, unsetS); return err }, unsetClass},
		{"SolveBars Е-2 before Е-4", func() error { _, err := tempo.SolveBars(fi(0), nofr, zero); return err }, argClass},
		{"SolveBarUnits Е-2 before Е-4", func() error { _, err := tempo.SolveBarUnits(nofr, fi(0), zero); return err }, argClass},
		{"SolveBPM Е-1 before Е-2", func() error { _, err := tempo.SolveBPM(nofr, fi(0), unsetS); return err }, unsetClass},
		{"SolveBPM Е-4 before Е-11", func() error { _, err := tempo.SolveBPM(maxF(), maxF(), zero); return err }, noSolution},
		{"SolveSeconds Е-2 before Е-4", func() error { _, err := tempo.SolveSeconds(fi(0), fi(0), nofr); return err }, argClass},
		{"SolveSeconds Е-4 before Е-11", func() error { _, err := tempo.SolveSeconds(maxF(), maxF(), fi(0)); return err }, noSolution},
		{"Quantize Е-1 before Е-2", func() error { _, err := unsetT.Quantize(nofr, roundingMode(0)); return err }, unsetClass},
		{"Quantize Е-2 mode before step", func() error { _, err := tq(1).Quantize(nofr, roundingMode(0)); return err }, modeUnset},
		{"Quantize Е-2 unknown mode before step", func() error { _, err := tq(1).Quantize(fi(-1), roundingMode(42)); return err }, modeUnknown},
		{"Quantize Е-2 step before Е-8", func() error { _, err := tm(q, fr(1, 4)).Quantize(fi(0), down); return err }, argClass},
		{"SnapToScale Е-1 before Е-2", func() error { _, err := unsetT.SnapToScale(roundingMode(0)); return err }, unsetClass},
		{"SnapToScale Е-2 before Е-6", func() error { _, err := tq(30).SnapToScale(roundingMode(0)); return err }, modeUnset},
		{"TermsIn Е-1 before Е-2", func() error { _, err := unsetT.TermsIn(term.List{}); return err }, unsetClass},
		{"Range.Contains Е-1 before Е-2", func() error { _, err := unsetR.Contains(nofr); return err }, unsetClass},
		{"Range.Contains Е-2 before Е-3", func() error { _, err := rg(q, fi(1), fi(2)).Contains(nofr); return err }, argClass},
		{"Range.ContainsTempo both unset", func() error { _, err := unsetR.ContainsTempo(unsetT); return err }, unsetClass},
		{"MidTempo Е-11", func() error {
			_, err := rg(q, invMaxF(), fbig(big.NewInt(2), fraction.MaxComponentBig())).MidTempo()
			return err
		}, fracRange},
		{"IsScaleDivision Е-2", func() error { _, err := tempo.IsScaleDivision(nofr); return err }, argClass},
		{"ScaleIndex Е-3 before Е-6", func() error { _, err := tempo.ScaleIndex(fi(-110)); return err }, noTempoClass},
		{"Next Е-3 before Е-6", func() error { _, err := tempo.NextScaleDivision(fi(-300)); return err }, noTempoClass},
		{"Prev Е-3 before Е-6", func() error { _, err := tempo.PrevScaleDivision(fi(0)); return err }, noTempoClass},
		{"ScaleNeighbors Е-2 before Е-3", func() error { _, _, _, err := tempo.ScaleNeighbors(nofr); return err }, argClass},
		{"SecondsFromFraction Е-2", func() error { _, err := tempo.SecondsFromFraction(nofr); return err }, argClass},
		{"Seconds.Sub Е-1 before Е-5", func() error { _, err := unsetS.Sub(sec(1, 1)); return err }, unsetClass},
		{"Seconds.Sub Е-5 before Е-11", func() error { _, err := secF(invMaxF()).Sub(secF(maxF())); return err }, negTime},
		{"Seconds.Minutes Е-11", func() error { _, err := secF(invMaxF()).Minutes(); return err }, fracRange},
		{"TimeDuration Е-1 before Е-2", func() error { _, err := unsetS.TimeDuration(roundingMode(0)); return err }, unsetClass},
		{"TimeDuration Е-2 before Е-9", func() error { _, err := secF(maxF()).TimeDuration(roundingMode(0)); return err }, modeUnset},
		{"WithTerm Е-1 before Е-2", func() error { _, err := unsetI.WithTerm(term.Term{}); return err }, unsetClass},
		{"WithoutTerm Е-1", func() error { _, err := unsetI.WithoutTerm(); return err }, unsetClass},
		{"IndicationOfTempo Е-1", func() error { _, err := tempo.IndicationOfTempo(unsetT); return err }, unsetClass},
		{"ParseSeconds Е-10 before Е-11", func() error { _, err := tempo.ParseSeconds(maxPlusOne() + "/0"); return err }, textClass},
		{"ParseSeconds Е-11 numerator", func() error { _, err := tempo.ParseSeconds(maxPlusOne() + "/" + maxPlusOne()); return err }, fracRange},
		{"ParseApprox Е-11 unit before Е-3", func() error { _, err := tempo.ParseApprox("4" + rep(".", 300) + "≈0"); return err }, unitValueErr},
		{"ParseIndication Е-2 before Е-10", func() error { _, err := tempo.ParseIndication("\xff", term.List{}); return err }, argClass},
		{"UnmarshalText Е-2 before the grammar", func() error { return (*tempo.Range)(nil).UnmarshalText([]byte("x")) }, argClass},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.want.check(t, tt.call())
		})
	}
}
