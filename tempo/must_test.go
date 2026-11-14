package tempo_test

import (
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempo/term"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mustCase pairs an operation with its Must variant (architecture 4.3,
// "Must"; Т-25). call runs the base operation with good or bad arguments and
// returns its results without the error; must runs the Must variant with the
// same arguments.
type mustCase struct {
	name string
	call func(good bool) ([]any, error)
	must func(good bool) []any
}

// m1 builds a mustCase for an operation with one result.
func m1[T any](name string, call func(bool) (T, error), must func(bool) T) mustCase {
	return mustCase{
		name: name,
		call: func(g bool) ([]any, error) { v, err := call(g); return []any{v}, err },
		must: func(g bool) []any { return []any{must(g)} },
	}
}

// m2 builds a mustCase for an operation with two results.
func m2[T, U any](name string, call func(bool) (T, U, error), must func(bool) (T, U)) mustCase {
	return mustCase{
		name: name,
		call: func(g bool) ([]any, error) { v, w, err := call(g); return []any{v, w}, err },
		must: func(g bool) []any { v, w := must(g); return []any{v, w} },
	}
}

// pick returns good if g, else bad.
func pick[T any](g bool, good, bad T) T {
	if g {
		return good
	}

	return bad
}

func functionMustCases() []mustCase {
	n := func(g bool) fraction.Fraction { return pick(g, fi(60), fi(0)) }
	s := func(g bool) tempo.Seconds { return pick(g, sec(60, 1), tempo.ZeroSeconds()) }
	en := term.EnglishWikipedia()

	return []mustCase{
		m1("New", func(g bool) (tempo.Tempo, error) { return tempo.New(q, n(g)) },
			func(g bool) tempo.Tempo { return tempo.MustNew(q, n(g)) }),
		m1("NewQuarter", func(g bool) (tempo.Tempo, error) { return tempo.NewQuarter(n(g)) },
			func(g bool) tempo.Tempo { return tempo.MustNewQuarter(n(g)) }),
		m1("FromSeconds", func(g bool) (tempo.Tempo, error) { return tempo.FromSeconds(q, dur(1, 4), s(g)) },
			func(g bool) tempo.Tempo { return tempo.MustFromSeconds(q, dur(1, 4), s(g)) }),
		m1("FromBars", func(g bool) (tempo.Tempo, error) { return tempo.FromBars(q, dur(1, 2), fi(30), s(g)) },
			func(g bool) tempo.Tempo { return tempo.MustFromBars(q, dur(1, 2), fi(30), s(g)) }),
		m1("FromBarsPerMinute", func(g bool) (tempo.Tempo, error) { return tempo.FromBarsPerMinute(q, dur(3, 4), n(g)) },
			func(g bool) tempo.Tempo { return tempo.MustFromBarsPerMinute(q, dur(3, 4), n(g)) }),
		m1("Parse", func(g bool) (tempo.Tempo, error) { return tempo.Parse(pick(g, "4=60", "4=0")) },
			func(g bool) tempo.Tempo { return tempo.MustParse(pick(g, "4=60", "4=0")) }),
		m1("BarUnits", func(g bool) (fraction.Fraction, error) {
			return tempo.BarUnits(dur(1, 1), pick(g, dur(1, 4), duration.Duration{}))
		},
			func(g bool) fraction.Fraction {
				return tempo.MustBarUnits(dur(1, 1), pick(g, dur(1, 4), duration.Duration{}))
			}),
		m1("BarLength", func(g bool) (duration.Duration, error) { return tempo.BarLength(dur(1, 4), n(g)) },
			func(g bool) duration.Duration { return tempo.MustBarLength(dur(1, 4), n(g)) }),
		m1("UnitLength", func(g bool) (duration.Duration, error) { return tempo.UnitLength(dur(1, 2), n(g)) },
			func(g bool) duration.Duration { return tempo.MustUnitLength(dur(1, 2), n(g)) }),
		m1("SolveBars", func(g bool) (fraction.Fraction, error) { return tempo.SolveBars(fi(2), n(g), s(true)) },
			func(g bool) fraction.Fraction { return tempo.MustSolveBars(fi(2), n(g), s(true)) }),
		m1("SolveBarUnits", func(g bool) (fraction.Fraction, error) { return tempo.SolveBarUnits(fi(30), n(g), s(true)) },
			func(g bool) fraction.Fraction { return tempo.MustSolveBarUnits(fi(30), n(g), s(true)) }),
		m1("SolveBPM", func(g bool) (fraction.Fraction, error) { return tempo.SolveBPM(fi(30), fi(2), s(g)) },
			func(g bool) fraction.Fraction { return tempo.MustSolveBPM(fi(30), fi(2), s(g)) }),
		m1("SolveSeconds", func(g bool) (tempo.Seconds, error) { return tempo.SolveSeconds(fi(30), fi(2), n(g)) },
			func(g bool) tempo.Seconds { return tempo.MustSolveSeconds(fi(30), fi(2), n(g)) }),
		m1("NewSeconds", func(g bool) (tempo.Seconds, error) { return tempo.NewSeconds(1, pick(g, 3, 0)) },
			func(g bool) tempo.Seconds { return tempo.MustNewSeconds(1, pick(g, 3, 0)) }),
		m1("SecondsFromFraction", func(g bool) (tempo.Seconds, error) { return tempo.SecondsFromFraction(pick(g, fi(1), fi(-1))) },
			func(g bool) tempo.Seconds { return tempo.MustSecondsFromFraction(pick(g, fi(1), fi(-1))) }),
		m1("SecondsFromMinutes", func(g bool) (tempo.Seconds, error) { return tempo.SecondsFromMinutes(pick(g, fi(1), maxF())) },
			func(g bool) tempo.Seconds { return tempo.MustSecondsFromMinutes(pick(g, fi(1), maxF())) }),
		m1("SecondsFromTimeDuration",
			func(g bool) (tempo.Seconds, error) {
				return tempo.SecondsFromTimeDuration(pick(g, time.Second, -time.Second))
			},
			func(g bool) tempo.Seconds {
				return tempo.MustSecondsFromTimeDuration(pick(g, time.Second, -time.Second))
			}),
		m1("ParseSeconds", func(g bool) (tempo.Seconds, error) { return tempo.ParseSeconds(pick(g, "1/3", "1/0")) },
			func(g bool) tempo.Seconds { return tempo.MustParseSeconds(pick(g, "1/3", "1/0")) }),
		m1("NewApprox", func(g bool) (tempo.Approx, error) { return tempo.NewApprox(q, n(g)) },
			func(g bool) tempo.Approx { return tempo.MustNewApprox(q, n(g)) }),
		m1("ParseApprox", func(g bool) (tempo.Approx, error) { return tempo.ParseApprox(pick(g, "4≈108", "4=108")) },
			func(g bool) tempo.Approx { return tempo.MustParseApprox(pick(g, "4≈108", "4=108")) }),
		m1("NewRange", func(g bool) (tempo.Range, error) { return tempo.NewRange(q, fi(50), pick(g, fi(60), fi(40))) },
			func(g bool) tempo.Range { return tempo.MustNewRange(q, fi(50), pick(g, fi(60), fi(40))) }),
		m1("RangeOf", func(g bool) (tempo.Range, error) {
			return tempo.RangeOf(en.MustParse(pick(g, "Allegro", "Larghissimo")), q)
		}, func(g bool) tempo.Range { return tempo.MustRangeOf(en.MustParse(pick(g, "Allegro", "Larghissimo")), q) }),
		m1("ParseRange", func(g bool) (tempo.Range, error) { return tempo.ParseRange(pick(g, "8=132-144", "8=144-132")) },
			func(g bool) tempo.Range { return tempo.MustParseRange(pick(g, "8=132-144", "8=144-132")) }),
		m1("ScaleDivision", func(g bool) (int, error) { return tempo.ScaleDivision(pick(g, 21, 39)) },
			func(g bool) int { return tempo.MustScaleDivision(pick(g, 21, 39)) }),
		m1("IsScaleDivision", func(g bool) (bool, error) { return tempo.IsScaleDivision(pick(g, fi(100), fi(0))) },
			func(g bool) bool { return tempo.MustIsScaleDivision(pick(g, fi(100), fi(0))) }),
		m1("ScaleIndex", func(g bool) (int, error) { return tempo.ScaleIndex(pick(g, fi(100), fi(110))) },
			func(g bool) int { return tempo.MustScaleIndex(pick(g, fi(100), fi(110))) }),
		m1("NextScaleDivision", func(g bool) (int, error) { return tempo.NextScaleDivision(pick(g, fi(110), fi(208))) },
			func(g bool) int { return tempo.MustNextScaleDivision(pick(g, fi(110), fi(208))) }),
		m1("PrevScaleDivision", func(g bool) (int, error) { return tempo.PrevScaleDivision(pick(g, fi(110), fi(40))) },
			func(g bool) int { return tempo.MustPrevScaleDivision(pick(g, fi(110), fi(40))) }),
		{
			name: "ScaleNeighbors",
			call: func(g bool) ([]any, error) {
				lo, hi, c, err := tempo.ScaleNeighbors(pick(g, fi(110), fi(30)))
				return []any{lo, hi, c}, err
			},
			must: func(g bool) []any {
				lo, hi, c := tempo.MustScaleNeighbors(pick(g, fi(110), fi(30)))
				return []any{lo, hi, c}
			},
		},
		m1("NearestScaleDivision",
			func(g bool) (int, error) {
				return tempo.NearestScaleDivision(fi(110), pick(g, halfEven, roundingMode(0)))
			},
			func(g bool) int { return tempo.MustNearestScaleDivision(fi(110), pick(g, halfEven, roundingMode(0))) }),
		m1("IndicationOf", func(g bool) (tempo.Indication, error) {
			return tempo.IndicationOf(pick(g, en.MustParse("Allegro"), term.Term{}))
		}, func(g bool) tempo.Indication {
			return tempo.MustIndicationOf(pick(g, en.MustParse("Allegro"), term.Term{}))
		}),
		m1("IndicationOfTempo", func(g bool) (tempo.Indication, error) { return tempo.IndicationOfTempo(pick(g, tq(60), tempo.Tempo{})) },
			func(g bool) tempo.Indication { return tempo.MustIndicationOfTempo(pick(g, tq(60), tempo.Tempo{})) }),
		m1("IndicationOfApprox",
			func(g bool) (tempo.Indication, error) {
				return tempo.IndicationOfApprox(pick(g, ap(q, fi(1)), tempo.Approx{}))
			},
			func(g bool) tempo.Indication {
				return tempo.MustIndicationOfApprox(pick(g, ap(q, fi(1)), tempo.Approx{}))
			}),
		m1("IndicationOfRange",
			func(g bool) (tempo.Indication, error) {
				return tempo.IndicationOfRange(pick(g, rg(q, fi(1), fi(2)), tempo.Range{}))
			},
			func(g bool) tempo.Indication {
				return tempo.MustIndicationOfRange(pick(g, rg(q, fi(1), fi(2)), tempo.Range{}))
			}),
		m1("ParseIndication",
			func(g bool) (tempo.Indication, error) {
				return tempo.ParseIndication(pick(g, "Allegro, 4=120", "Fast"), en)
			},
			func(g bool) tempo.Indication { return tempo.MustParseIndication(pick(g, "Allegro, 4=120", "Fast"), en) }),
	}
}

func tempoMustCases() []mustCase {
	x := func(g bool) tempo.Tempo { return pick(g, tm(dq, fi(60)), tempo.Tempo{}) }
	o := tq(120)
	s := sec(60, 1)

	return []mustCase{
		m1("Tempo.Unit", func(g bool) (duration.Written, error) { return x(g).Unit() },
			func(g bool) duration.Written { return x(g).MustUnit() }),
		m1("Tempo.UnitDuration", func(g bool) (duration.Duration, error) { return x(g).UnitDuration() },
			func(g bool) duration.Duration { return x(g).MustUnitDuration() }),
		m1("Tempo.BPM", func(g bool) (fraction.Fraction, error) { return x(g).BPM() },
			func(g bool) fraction.Fraction { return x(g).MustBPM() }),
		m1("Tempo.Speed", func(g bool) (fraction.Fraction, error) { return x(g).Speed() },
			func(g bool) fraction.Fraction { return x(g).MustSpeed() }),
		m1("Tempo.QuartersPerMinute", func(g bool) (fraction.Fraction, error) { return x(g).QuartersPerMinute() },
			func(g bool) fraction.Fraction { return x(g).MustQuartersPerMinute() }),
		m1("Tempo.EqualSpeed", func(g bool) (bool, error) { return x(g).EqualSpeed(o) },
			func(g bool) bool { return x(g).MustEqualSpeed(o) }),
		m1("Tempo.CompareSpeed", func(g bool) (int, error) { return x(g).CompareSpeed(o) },
			func(g bool) int { return x(g).MustCompareSpeed(o) }),
		m1("Tempo.SpeedRatio", func(g bool) (fraction.Fraction, error) { return x(g).SpeedRatio(o) },
			func(g bool) fraction.Fraction { return x(g).MustSpeedRatio(o) }),
		m1("Tempo.WithUnit", func(g bool) (tempo.Tempo, error) { return x(g).WithUnit(e) },
			func(g bool) tempo.Tempo { return x(g).MustWithUnit(e) }),
		m1("Tempo.WithBPM", func(g bool) (tempo.Tempo, error) { return x(g).WithBPM(fi(90)) },
			func(g bool) tempo.Tempo { return x(g).MustWithBPM(fi(90)) }),
		m1("Tempo.Mul", func(g bool) (tempo.Tempo, error) { return x(g).Mul(fr(3, 2)) },
			func(g bool) tempo.Tempo { return x(g).MustMul(fr(3, 2)) }),
		m1("Tempo.Double", func(g bool) (tempo.Tempo, error) { return x(g).Double() },
			func(g bool) tempo.Tempo { return x(g).MustDouble() }),
		m1("Tempo.Halve", func(g bool) (tempo.Tempo, error) { return x(g).Halve() },
			func(g bool) tempo.Tempo { return x(g).MustHalve() }),
		m1("Tempo.Modulate", func(g bool) (tempo.Tempo, error) { return x(g).Modulate(q, dur(1, 8)) },
			func(g bool) tempo.Tempo { return x(g).MustModulate(q, dur(1, 8)) }),
		m1("Tempo.ModulationTo", func(g bool) (duration.Duration, error) { return x(g).ModulationTo(o) },
			func(g bool) duration.Duration { return x(g).MustModulationTo(o) }),
		m1("Tempo.Seconds", func(g bool) (tempo.Seconds, error) { return x(g).Seconds(dur(1, 8)) },
			func(g bool) tempo.Seconds { return x(g).MustSeconds(dur(1, 8)) }),
		m1("Tempo.UnitSeconds", func(g bool) (tempo.Seconds, error) { return x(g).UnitSeconds() },
			func(g bool) tempo.Seconds { return x(g).MustUnitSeconds() }),
		m1("Tempo.UnitsPerSecond", func(g bool) (fraction.Fraction, error) { return x(g).UnitsPerSecond() },
			func(g bool) fraction.Fraction { return x(g).MustUnitsPerSecond() }),
		m1("Tempo.DurationIn", func(g bool) (duration.Duration, error) { return x(g).DurationIn(s) },
			func(g bool) duration.Duration { return x(g).MustDurationIn(s) }),
		m1("Tempo.UnitsIn", func(g bool) (fraction.Fraction, error) { return x(g).UnitsIn(s) },
			func(g bool) fraction.Fraction { return x(g).MustUnitsIn(s) }),
		m1("Tempo.SecondsOfUnits", func(g bool) (tempo.Seconds, error) { return x(g).SecondsOfUnits(fi(3)) },
			func(g bool) tempo.Seconds { return x(g).MustSecondsOfUnits(fi(3)) }),
		m1("Tempo.BarUnits", func(g bool) (fraction.Fraction, error) { return x(g).BarUnits(dur(3, 4)) },
			func(g bool) fraction.Fraction { return x(g).MustBarUnits(dur(3, 4)) }),
		m1("Tempo.Bars", func(g bool) (fraction.Fraction, error) { return x(g).Bars(dur(3, 4), s) },
			func(g bool) fraction.Fraction { return x(g).MustBars(dur(3, 4), s) }),
		m1("Tempo.SecondsOfBars", func(g bool) (tempo.Seconds, error) { return x(g).SecondsOfBars(dur(3, 4), fi(30)) },
			func(g bool) tempo.Seconds { return x(g).MustSecondsOfBars(dur(3, 4), fi(30)) }),
		m2("Tempo.BarFor", func(g bool) (fraction.Fraction, duration.Duration, error) { return x(g).BarFor(fi(30), s) },
			func(g bool) (fraction.Fraction, duration.Duration) { return x(g).MustBarFor(fi(30), s) }),
		m1("Tempo.BarsPerMinute", func(g bool) (fraction.Fraction, error) { return x(g).BarsPerMinute(dur(3, 4)) },
			func(g bool) fraction.Fraction { return x(g).MustBarsPerMinute(dur(3, 4)) }),
		m1("Tempo.Quantize", func(g bool) (tempo.Tempo, error) { return x(g).Quantize(fi(7), halfUp) },
			func(g bool) tempo.Tempo { return x(g).MustQuantize(fi(7), halfUp) }),
		m1("Tempo.SnapToScale", func(g bool) (tempo.Tempo, error) { return x(g).SnapToScale(halfUp) },
			func(g bool) tempo.Tempo { return x(g).MustSnapToScale(halfUp) }),
		m1("Tempo.TermsIn", func(g bool) ([]term.Term, error) { return x(g).TermsIn(term.EnglishWikipedia()) },
			func(g bool) []term.Term { return x(g).MustTermsIn(term.EnglishWikipedia()) }),
	}
}

func secondsMustCases() []mustCase {
	x := func(g bool) tempo.Seconds { return pick(g, sec(90, 1), tempo.Seconds{}) }

	return []mustCase{
		m1("Seconds.Fraction", func(g bool) (fraction.Fraction, error) { return x(g).Fraction() },
			func(g bool) fraction.Fraction { return x(g).MustFraction() }),
		m1("Seconds.Minutes", func(g bool) (fraction.Fraction, error) { return x(g).Minutes() },
			func(g bool) fraction.Fraction { return x(g).MustMinutes() }),
		m1("Seconds.Add", func(g bool) (tempo.Seconds, error) { return x(g).Add(sec(1, 3)) },
			func(g bool) tempo.Seconds { return x(g).MustAdd(sec(1, 3)) }),
		m1("Seconds.Sub", func(g bool) (tempo.Seconds, error) { return sec(1, 1).Sub(pick(g, sec(1, 3), sec(2, 1))) },
			func(g bool) tempo.Seconds { return sec(1, 1).MustSub(pick(g, sec(1, 3), sec(2, 1))) }),
		m1("Seconds.TimeDuration", func(g bool) (time.Duration, error) { return sec(1, 3).TimeDuration(pick(g, up, roundingMode(0))) },
			func(g bool) time.Duration { return sec(1, 3).MustTimeDuration(pick(g, up, roundingMode(0))) }),
	}
}

func approxMustCases() []mustCase {
	x := func(g bool) tempo.Approx { return pick(g, ap(dq, fi(60)), tempo.Approx{}) }
	o := ap(e, fi(180))

	return []mustCase{
		m1("Approx.Unit", func(g bool) (duration.Written, error) { return x(g).Unit() },
			func(g bool) duration.Written { return x(g).MustUnit() }),
		m1("Approx.UnitDuration", func(g bool) (duration.Duration, error) { return x(g).UnitDuration() },
			func(g bool) duration.Duration { return x(g).MustUnitDuration() }),
		m1("Approx.BPM", func(g bool) (fraction.Fraction, error) { return x(g).BPM() },
			func(g bool) fraction.Fraction { return x(g).MustBPM() }),
		m1("Approx.Speed", func(g bool) (fraction.Fraction, error) { return x(g).Speed() },
			func(g bool) fraction.Fraction { return x(g).MustSpeed() }),
		m1("Approx.QuartersPerMinute", func(g bool) (fraction.Fraction, error) { return x(g).QuartersPerMinute() },
			func(g bool) fraction.Fraction { return x(g).MustQuartersPerMinute() }),
		m1("Approx.EqualSpeed", func(g bool) (bool, error) { return x(g).EqualSpeed(o) },
			func(g bool) bool { return x(g).MustEqualSpeed(o) }),
		m1("Approx.CompareSpeed", func(g bool) (int, error) { return x(g).CompareSpeed(o) },
			func(g bool) int { return x(g).MustCompareSpeed(o) }),
		m1("Approx.WithUnit", func(g bool) (tempo.Approx, error) { return x(g).WithUnit(e) },
			func(g bool) tempo.Approx { return x(g).MustWithUnit(e) }),
		m1("Approx.Tempo", func(g bool) (tempo.Tempo, error) { return x(g).Tempo() },
			func(g bool) tempo.Tempo { return x(g).MustTempo() }),
		m1("Approx.TermsIn", func(g bool) ([]term.Term, error) { return x(g).TermsIn(term.EnglishWikipedia()) },
			func(g bool) []term.Term { return x(g).MustTermsIn(term.EnglishWikipedia()) }),
	}
}

func rangeMustCases() []mustCase {
	x := func(g bool) tempo.Range { return pick(g, rg(e, fi(132), fi(144)), tempo.Range{}) }

	return []mustCase{
		m1("Range.Unit", func(g bool) (duration.Written, error) { return x(g).Unit() },
			func(g bool) duration.Written { return x(g).MustUnit() }),
		m1("Range.UnitDuration", func(g bool) (duration.Duration, error) { return x(g).UnitDuration() },
			func(g bool) duration.Duration { return x(g).MustUnitDuration() }),
		m1("Range.Lower", func(g bool) (fraction.Fraction, error) { return x(g).Lower() },
			func(g bool) fraction.Fraction { return x(g).MustLower() }),
		m1("Range.Upper", func(g bool) (fraction.Fraction, error) { return x(g).Upper() },
			func(g bool) fraction.Fraction { return x(g).MustUpper() }),
		m1("Range.LowerTempo", func(g bool) (tempo.Tempo, error) { return x(g).LowerTempo() },
			func(g bool) tempo.Tempo { return x(g).MustLowerTempo() }),
		m1("Range.UpperTempo", func(g bool) (tempo.Tempo, error) { return x(g).UpperTempo() },
			func(g bool) tempo.Tempo { return x(g).MustUpperTempo() }),
		m1("Range.MidTempo", func(g bool) (tempo.Tempo, error) { return x(g).MidTempo() },
			func(g bool) tempo.Tempo { return x(g).MustMidTempo() }),
		m1("Range.EqualSpeed", func(g bool) (bool, error) { return x(g).EqualSpeed(rg(q, fi(66), fi(72))) },
			func(g bool) bool { return x(g).MustEqualSpeed(rg(q, fi(66), fi(72))) }),
		m1("Range.WithUnit", func(g bool) (tempo.Range, error) { return x(g).WithUnit(q) },
			func(g bool) tempo.Range { return x(g).MustWithUnit(q) }),
		m1("Range.Contains", func(g bool) (bool, error) { return x(true).Contains(pick(g, fi(138), fi(0))) },
			func(g bool) bool { return x(true).MustContains(pick(g, fi(138), fi(0))) }),
		m1("Range.ContainsTempo", func(g bool) (bool, error) { return x(true).ContainsTempo(pick(g, tq(69), tempo.Tempo{})) },
			func(g bool) bool { return x(true).MustContainsTempo(pick(g, tq(69), tempo.Tempo{})) }),
	}
}

func indicationMustCases() []mustCase {
	en := term.EnglishWikipedia()
	both := tempo.MustIndicationOf(en.MustParse("Allegro")).MustWithTempo(tq(120))
	x := func(g bool) tempo.Indication { return pick(g, both, tempo.Indication{}) }

	return []mustCase{
		m2("Indication.Term", func(g bool) (term.Term, bool, error) { return x(g).Term() },
			func(g bool) (term.Term, bool) { return x(g).MustTerm() }),
		m1("Indication.Kind", func(g bool) (tempo.Kind, error) { return x(g).Kind() },
			func(g bool) tempo.Kind { return x(g).MustKind() }),
		m2("Indication.Tempo", func(g bool) (tempo.Tempo, bool, error) { return x(g).Tempo() },
			func(g bool) (tempo.Tempo, bool) { return x(g).MustTempo() }),
		m2("Indication.Approx", func(g bool) (tempo.Approx, bool, error) { return x(g).Approx() },
			func(g bool) (tempo.Approx, bool) { return x(g).MustApprox() }),
		m2("Indication.Range", func(g bool) (tempo.Range, bool, error) { return x(g).Range() },
			func(g bool) (tempo.Range, bool) { return x(g).MustRange() }),
		m1("Indication.WithTerm", func(g bool) (tempo.Indication, error) { return x(g).WithTerm(en.MustParse("Presto")) },
			func(g bool) tempo.Indication { return x(g).MustWithTerm(en.MustParse("Presto")) }),
		m1("Indication.WithTempo", func(g bool) (tempo.Indication, error) { return x(g).WithTempo(tq(1)) },
			func(g bool) tempo.Indication { return x(g).MustWithTempo(tq(1)) }),
		m1("Indication.WithApprox", func(g bool) (tempo.Indication, error) { return x(g).WithApprox(ap(q, fi(1))) },
			func(g bool) tempo.Indication { return x(g).MustWithApprox(ap(q, fi(1))) }),
		m1("Indication.WithRange", func(g bool) (tempo.Indication, error) { return x(g).WithRange(rg(q, fi(1), fi(2))) },
			func(g bool) tempo.Indication { return x(g).MustWithRange(rg(q, fi(1), fi(2))) }),
		m1("Indication.WithoutTerm", func(g bool) (tempo.Indication, error) { return x(g).WithoutTerm() },
			func(g bool) tempo.Indication { return x(g).MustWithoutTerm() }),
		m1("Indication.WithoutMarking", func(g bool) (tempo.Indication, error) { return x(g).WithoutMarking() },
			func(g bool) tempo.Indication { return x(g).MustWithoutMarking() }),
	}
}

// TestMust covers architecture 4.3, "Must": every Must variant of 4.2 (101
// in all, architecture 10, Р-TP-1) returns the results of its base operation
// on success and panics with the same error value on failure.
func TestMust(t *testing.T) {
	t.Parallel()

	groups := map[string]struct {
		cases []mustCase
		count int
	}{
		"functions":  {functionMustCases(), 35},
		"Tempo":      {tempoMustCases(), 29},
		"Seconds":    {secondsMustCases(), 5},
		"Approx":     {approxMustCases(), 10},
		"Range":      {rangeMustCases(), 11},
		"Indication": {indicationMustCases(), 11},
	}

	for name, g := range groups {
		require.Len(t, g.cases, g.count, name)

		for _, tc := range g.cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				want, err := tc.call(true)
				require.NoError(t, err)

				var got []any

				_, panicked := capturePanic(func() { got = tc.must(true) })
				require.False(t, panicked)
				assert.Equal(t, want, got)

				_, baseErr := tc.call(false)
				require.Error(t, baseErr)

				mustErr := panicError(t, func() { tc.must(false) })
				require.Error(t, mustErr)
				assert.Equal(t, baseErr.Error(), mustErr.Error())

				for _, s := range append(slices.Clone(tempoSentinels), dependencySentinels...) {
					assert.Equal(t, errors.Is(baseErr, s.err), errors.Is(mustErr, s.err), s.name)
				}
			})
		}
	}
}

// TestMethodSets covers architecture 4.2 ("no exported identifiers beyond
// those listed") and 4.3 ("exactly one Must variant with the same parameters
// and results without error") for the methods of the types of tempo.
func TestMethodSets(t *testing.T) {
	t.Parallel()

	common := func(extra ...string) []string {
		return append([]string{"AppendText", "Equal", "IsSet", "MarshalText", "String"}, extra...)
	}

	want := map[reflect.Type][]string{
		reflect.TypeFor[tempo.Tempo](): common(
			"BPM", "BarFor", "BarUnits", "Bars", "BarsPerMinute", "Compare", "CompareSpeed", "Double", "DurationIn",
			"EqualSpeed", "Halve", "Modulate", "ModulationTo", "Mul", "QuartersPerMinute", "Quantize", "Seconds",
			"SecondsOfBars", "SecondsOfUnits", "SnapToScale", "Speed", "SpeedRatio", "TermsIn", "Unit", "UnitDuration",
			"UnitSeconds", "UnitsIn", "UnitsPerSecond", "WithBPM", "WithUnit",
		),
		reflect.TypeFor[tempo.Seconds](): common("Add", "Compare", "Fraction", "Minutes", "Sub", "TimeDuration"),
		reflect.TypeFor[tempo.Approx](): common(
			"BPM", "Compare", "CompareSpeed", "EqualSpeed", "QuartersPerMinute", "Speed", "TermsIn", "Tempo", "Unit",
			"UnitDuration", "WithUnit",
		),
		reflect.TypeFor[tempo.Range](): common(
			"Compare", "Contains", "ContainsTempo", "EqualSpeed", "Lower", "LowerTempo", "MidTempo", "Unit",
			"UnitDuration", "Upper", "UpperTempo", "WithUnit",
		),
		reflect.TypeFor[tempo.Indication](): common(
			"Approx", "Kind", "Range", "Tempo", "Term", "WithApprox", "WithRange", "WithTempo", "WithTerm",
			"WithoutMarking", "WithoutTerm",
		),
		reflect.TypeFor[tempo.Kind](): {"Equal", "IsValid", "String"},
	}

	withUnmarshal := map[reflect.Type]bool{
		reflect.TypeFor[tempo.Tempo](): true, reflect.TypeFor[tempo.Seconds](): true,
		reflect.TypeFor[tempo.Approx](): true, reflect.TypeFor[tempo.Range](): true,
	}

	for typ, methods := range want {
		var expected []string

		for _, m := range methods {
			expected = append(expected, m)

			if base, ok := mustBase(typ, m); ok {
				expected = append(expected, "Must"+base)
			}
		}

		slices.Sort(expected)

		var got []string
		for i := range typ.NumMethod() {
			got = append(got, typ.Method(i).Name)
		}

		assert.Equal(t, expected, got, "methods of %s", typ)

		ptr := reflect.PointerTo(typ)
		extra := 0

		if withUnmarshal[typ] {
			extra = 1
			_, ok := ptr.MethodByName("UnmarshalText")
			assert.True(t, ok, "%s has UnmarshalText on the pointer", typ)
		}

		assert.Equal(t, typ.NumMethod()+extra, ptr.NumMethod(), "pointer methods of %s", typ)

		checkMustPairs(t, typ)
	}
}

// mustBase reports whether the method name of typ returns an error and has a
// Must variant by architecture 4.3.
func mustBase(typ reflect.Type, name string) (string, bool) {
	if slices.Contains([]string{"MarshalText", "AppendText", "Equal", "IsSet", "String", "Compare", "IsValid"}, name) {
		return "", false
	}

	if typ == reflect.TypeFor[tempo.Kind]() {
		return "", false
	}

	return name, true
}

// checkMustPairs checks that every method returning an error, except
// MarshalText and AppendText, has a Must variant with the same parameters and
// the results without the error, and that every Must method has a base.
func checkMustPairs(t *testing.T, typ reflect.Type) {
	t.Helper()

	errType := reflect.TypeFor[error]()

	for i := range typ.NumMethod() {
		m := typ.Method(i)
		ft := m.Type

		if base, ok := strings.CutPrefix(m.Name, "Must"); ok {
			_, found := typ.MethodByName(base)
			assert.True(t, found, "%s.%s has no base method %s", typ, m.Name, base)

			continue
		}

		if ft.NumOut() == 0 || ft.Out(ft.NumOut()-1) != errType ||
			slices.Contains([]string{"MarshalText", "AppendText"}, m.Name) {
			continue
		}

		mm, found := typ.MethodByName("Must" + m.Name)
		if !assert.True(t, found, "%s.%s has no Must variant", typ, m.Name) {
			continue
		}

		mt := mm.Type
		if !assert.Equal(t, ft.NumIn(), mt.NumIn(), "%s.Must%s parameters", typ, m.Name) {
			continue
		}

		for j := range ft.NumIn() {
			assert.Equal(t, ft.In(j), mt.In(j), "%s.Must%s parameter %d", typ, m.Name, j)
		}

		if !assert.Equal(t, ft.NumOut()-1, mt.NumOut(), "%s.Must%s results", typ, m.Name) {
			continue
		}

		for j := range mt.NumOut() {
			assert.Equal(t, ft.Out(j), mt.Out(j), "%s.Must%s result %d", typ, m.Name, j)
		}
	}
}
