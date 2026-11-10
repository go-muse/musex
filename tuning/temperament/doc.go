// Package temperament provides mathematical models of musical temperaments and
// tuning tables expressed as exact cent offsets within a 2:1 octave.
//
// A [Temperament] is an ordered sequence of N step offsets T(0), T(1), ..., T(N-1)
// in [cents.Cents] for an octave divided into N steps (1 <= N <= [MaxStepsPerOctave]),
// satisfying the anchor condition T(0) = 0 and strict monotonicity
// 0 = T(0) < T(1) < ... < T(N-1) < 1200 cents. Multi-octave step indices s in Z
// are mapped via Euclidean division s = q*N + m (with 0 <= m < N) to cumulative
// offsets T_cum(s) = 1200*q + T(m).
//
// All offsets and intervals are represented exactly in the 31-prime limit vector
// space of [github.com/go-muse/musex/cents] without intermediate floating-point
// rounding: pure superparticular and rational intervals (such as 3:2, 5:4, 7:4),
// rational fractions of commas (such as 1/4 syntonic comma or 1/6 Pythagorean
// comma), and equal divisions of the octave (k * 1200 / N) preserve their exact
// algebraic identities.
//
// The package is independent of acoustic reference frequencies in Hz, written
// note spellings, and musical system degree mappings. It covers five major
// families of tuning systems and provides 31 built-in presets (available
// individually or via [Presets]):
//
//   - Equal divisions of the octave (N-EDO): [EDO] and presets [Equal12] ([Equal],
//     [EDO12]), [Equal19] ([EDO19]), [Equal22] ([EDO22]), [Equal24] ([EDO24]),
//     [Equal31] ([EDO31]), and [Equal53] ([EDO53]).
//   - Pythagorean tunings (3-limit): [NewPythagorean], [NewPythagoreanWithWolf],
//     [Pythagorean], and [PythagoreanEFlatGSharp].
//   - Just intonation (5-limit and 7-limit): [Just5Limit] ([Just]),
//     [Just5LimitSymmetric1], [Just5LimitSymmetric2], and [Just7Limit].
//   - Regular meantone temperaments: [NewMeantone], [NewMeantoneWithRange],
//     [NewMeantoneFromFifth], [NewMeantoneFromFifthWithRange], and presets
//     [MeantoneQuarterComma] ([Meantone]), [MeantoneThirdComma],
//     [MeantoneFifthComma], [MeantoneSixthComma], and [MeantoneTwoSeventhComma].
//   - Well and circulating temperaments: [FromFifths], [FromFifthsWithStart],
//     [FromFifthDeviations], [FromFifthDeviationsWithStart], [FromGenerator],
//     and presets [WerckmeisterIII], [WerckmeisterIV], [WerckmeisterV],
//     [KirnbergerI], [KirnbergerII], [KirnbergerIII], [Vallotti], [YoungI],
//     [YoungII] ([Young]), [Kellner], [NeidhardtI], [NeidhardtII],
//     [NeidhardtIII], and [NeidhardtDorf1732].
//
// Arbitrary temperaments can also be constructed from cent values ([New],
// [FromOffsets]), floating-point cent offsets ([FromFloats]), or frequency
// ratios ([FromRatios]). Beyond step and interval queries ([Temperament.Offset],
// [Temperament.SplitStep], [Temperament.CumulativeOffset],
// [Temperament.DirectedInterval], [Temperament.CyclicInterval],
// [Temperament.StepSizes], [Temperament.Rotate], [Temperament.NearestStep],
// [Temperament.NearestCumulativeStep]), the package supports structural
// comparison ([Temperament.Equal], [Temperament.Near], [Temperament.IsEqual],
// [Temperament.IsRational]) and interval/wolf analysis ([Temperament.AnalyzeEDO],
// [Temperament.AnalyzeIntervals], [Temperament.AnalyzeFifths],
// [Temperament.AnalyzeMajorThirds], [Temperament.AnalyzeMinorThirds]).
//
// A [Temperament] is an immutable value type: it holds only its name and its
// offsets table, which is never modified after construction, and the package
// keeps no mutable state. A Temperament is therefore safe for concurrent use by
// multiple goroutines. The zero value of [Temperament] represents an invalid,
// uninitialized temperament ([Temperament.IsValid] is false).
//
// The built-in presets are built once at package initialization. For valid
// arguments, reading [Temperament.StepsPerOctave], [Temperament.Name], and
// [Temperament.Offset], and comparing with [Temperament.Equal] do not
// allocate; returning an error may allocate. Step and interval queries
// ([Temperament.SplitStep], [Temperament.CumulativeOffset],
// [Temperament.DirectedInterval], [Temperament.CyclicInterval]), nearest-step
// searches ([Temperament.NearestStep], [Temperament.NearestCumulativeStep]),
// and the predicates [Temperament.IsEqual], [Temperament.IsRational], and
// [Temperament.Near] allocate memory only inside package cents, when creating
// and comparing cent values. Constructors and all other operations, such as
// [Temperament.Offsets], [Temperament.StepSizes], [Temperament.Rotate], the
// Analyze methods, [Presets], and [Temperament.String], allocate the values
// they return; their number of allocations is not specified. The package
// caches nothing between calls.
package temperament
