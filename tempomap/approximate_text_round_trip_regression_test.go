package tempomap_test

import (
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempomap"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	decimalQ75   = "1000000000000000000000000000000000000000000000000000000000000000000000000000"
	decimalQ76   = "10000000000000000000000000000000000000000000000000000000000000000000000000000"
	decimalQ77   = "100000000000000000000000000000000000000000000000000000000000000000000000000000"
	decimal2Q75  = "2000000000000000000000000000000000000000000000000000000000000000000000000000"
	decimal2Q76  = "20000000000000000000000000000000000000000000000000000000000000000000000000000"
	decimal2Q77  = "200000000000000000000000000000000000000000000000000000000000000000000000000000"
	maxComponent = "57896044618658097711785492504343953926634992332820282019728792003956564819967"
	maxPlusOne   = "57896044618658097711785492504343953926634992332820282019728792003956564819968"
	threeG       = "173688133855974293135356477513031861779904976998460846059186376011869694459904"
	sixG         = "347376267711948586270712955026063723559809953996921692118372752023739388919808"
)

type approximateTextFixture struct {
	name      string
	input     string
	canonical string
	value     string
	bound     string
	places    int
}

type approximateResultAdapter[T any, V any] struct {
	approximatePrefix       string
	exactPrefix             string
	unsetText               string
	dependencyRangeSentinel error

	parse       func(string) (T, error)
	mustParse   func(string) T
	equal       func(T, T) bool
	isSet       func(T) bool
	approximate func(T) (bool, error)
	places      func(T) (int, bool, error)
	value       func(T) (V, error)
	errorBound  func(T) (V, error)
	mustValue   func(string) V
	valueEqual  func(V, V) bool
	stringFn    func(T) string
	marshalText func(T) ([]byte, error)
	appendText  func(T, []byte) ([]byte, error)
	unmarshal   func(*T, []byte) error
}

func secondsApproximateResultAdapter() approximateResultAdapter[tempomap.SecondsResult, tempo.Seconds] {
	return approximateResultAdapter[tempomap.SecondsResult, tempo.Seconds]{
		approximatePrefix:       "s~",
		exactPrefix:             "s:",
		unsetText:               "SecondsResult(unset)",
		dependencyRangeSentinel: tempo.ErrOutOfRange,
		parse:                   tempomap.ParseSecondsResult,
		mustParse:               tempomap.MustParseSecondsResult,
		equal:                   tempomap.SecondsResult.Equal,
		isSet:                   tempomap.SecondsResult.IsSet,
		approximate:             tempomap.SecondsResult.Approximate,
		places:                  tempomap.SecondsResult.Places,
		value:                   tempomap.SecondsResult.Value,
		errorBound:              tempomap.SecondsResult.ErrorBound,
		mustValue:               tempo.MustParseSeconds,
		valueEqual:              tempo.Seconds.Equal,
		stringFn:                tempomap.SecondsResult.String,
		marshalText:             tempomap.SecondsResult.MarshalText,
		appendText:              tempomap.SecondsResult.AppendText,
		unmarshal:               (*tempomap.SecondsResult).UnmarshalText,
	}
}

func positionApproximateResultAdapter() approximateResultAdapter[tempomap.PositionResult, duration.Position] {
	return approximateResultAdapter[tempomap.PositionResult, duration.Position]{
		approximatePrefix:       "p~",
		exactPrefix:             "p:",
		unsetText:               "PositionResult(unset)",
		dependencyRangeSentinel: duration.ErrOutOfRange,
		parse:                   tempomap.ParsePositionResult,
		mustParse:               tempomap.MustParsePositionResult,
		equal:                   tempomap.PositionResult.Equal,
		isSet:                   tempomap.PositionResult.IsSet,
		approximate:             tempomap.PositionResult.Approximate,
		places:                  tempomap.PositionResult.Places,
		value:                   tempomap.PositionResult.Value,
		errorBound:              tempomap.PositionResult.ErrorBound,
		mustValue:               duration.MustParsePosition,
		valueEqual:              duration.Position.Equal,
		stringFn:                tempomap.PositionResult.String,
		marshalText:             tempomap.PositionResult.MarshalText,
		appendText:              tempomap.PositionResult.AppendText,
		unmarshal:               (*tempomap.PositionResult).UnmarshalText,
	}
}

func approximateTextFixtures(prefix string) []approximateTextFixture {
	approximate := prefix + "~"

	return []approximateTextFixture{
		{
			name:      "KP-28d.3-1_and_2_D76_directed",
			input:     approximate + "10/1@76:1/" + decimalQ76,
			canonical: approximate + decimalQ77 + "/" + decimalQ76 + "@76:1/" + decimalQ76,
			value:     "10/1",
			bound:     "1/" + decimalQ76,
			places:    76,
		},
		{
			name:      "KP-28d.3-3_D76_nearest",
			input:     approximate + "10/1@76:1/" + decimal2Q76,
			canonical: approximate + decimalQ77 + "/" + decimalQ76 + "@76:1/" + decimal2Q76,
			value:     "10/1",
			bound:     "1/" + decimal2Q76,
			places:    76,
		},
		{
			name:      "KP-28d.3-4_D75_directed",
			input:     approximate + "10/1@75:1/" + decimalQ75,
			canonical: approximate + decimalQ76 + "/" + decimalQ75 + "@75:1/" + decimalQ75,
			value:     "10/1",
			bound:     "1/" + decimalQ75,
			places:    75,
		},
		{
			name:      "KP-28d.3-4_D75_nearest",
			input:     approximate + "10/1@75:1/" + decimal2Q75,
			canonical: approximate + decimalQ76 + "/" + decimalQ75 + "@75:1/" + decimal2Q75,
			value:     "10/1",
			bound:     "1/" + decimal2Q75,
			places:    75,
		},
		{
			name:      "KP-28d.3-5_unreduced_value_and_directed_bound",
			input:     approximate + "10/" + decimalQ77 + "@76:10/" + decimalQ77,
			canonical: approximate + "1/" + decimalQ76 + "@76:1/" + decimalQ76,
			value:     "1/" + decimalQ76,
			bound:     "1/" + decimalQ76,
			places:    76,
		},
		{
			name:      "KP-28d.3-5_unreduced_value_and_nearest_bound",
			input:     approximate + "10/" + decimalQ77 + "@76:10/" + decimal2Q77,
			canonical: approximate + "1/" + decimalQ76 + "@76:1/" + decimal2Q76,
			value:     "1/" + decimalQ76,
			bound:     "1/" + decimal2Q76,
			places:    76,
		},
		{
			name:      "Additional_max_component_value",
			input:     approximate + maxComponent + "/1@0:1/1",
			canonical: approximate + maxComponent + "/1@0:1/1",
			value:     maxComponent + "/1",
			bound:     "1/1",
			places:    0,
		},
		{
			name:      "Additional_common_non_decimal_divisor",
			input:     approximate + threeG + "/" + sixG + "@1:1/10",
			canonical: approximate + "5/10@1:1/10",
			value:     "1/2",
			bound:     "1/10",
			places:    1,
		},
	}
}

func TestSecondsResultApproximateTextRoundTrips(t *testing.T) {
	t.Parallel()
	testApproximateResultTextRoundTrips(t, secondsApproximateResultAdapter(), approximateTextFixtures("s"))
}

func TestPositionResultApproximateTextRoundTrips(t *testing.T) {
	t.Parallel()
	testApproximateResultTextRoundTrips(t, positionApproximateResultAdapter(), approximateTextFixtures("p"))
}

func testApproximateResultTextRoundTrips[T any, V any](
	t *testing.T,
	adapter approximateResultAdapter[T, V],
	fixtures []approximateTextFixture,
) {
	t.Helper()

	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			t.Parallel()
			assertApproximateResultTextRoundTrip(t, adapter, fixture)
		})
	}
}

func assertApproximateResultTextRoundTrip[T any, V any](
	t *testing.T,
	adapter approximateResultAdapter[T, V],
	fixture approximateTextFixture,
	calculated ...T,
) {
	t.Helper()
	is := assert.New(t)
	must := require.New(t)

	parsed, err := adapter.parse(fixture.input)
	must.NoError(err)
	assertApproximateResultMetadata(t, adapter, parsed, fixture)

	original := parsed
	if len(calculated) == 1 {
		original = calculated[0]
		must.True(adapter.equal(parsed, original), "calculated result differs from its independently parsed fixture")
		assertApproximateResultMetadata(t, adapter, original, fixture)
	}

	is.Equal(fixture.canonical, adapter.stringFn(original))

	reparsed, err := adapter.parse(adapter.stringFn(original))
	must.NoError(err)
	is.True(adapter.equal(original, reparsed), "Parse(String()) did not preserve the result")
	is.Equal(fixture.canonical, adapter.stringFn(reparsed))
	assertApproximateResultMetadata(t, adapter, reparsed, fixture)

	mustParsed := adapter.mustParse(fixture.canonical)
	is.True(adapter.equal(original, mustParsed), "MustParse(String()) did not preserve the result")
	assertApproximateResultMetadata(t, adapter, mustParsed, fixture)

	marshaled, err := adapter.marshalText(original)
	must.NoError(err)
	is.Equal([]byte(fixture.canonical), marshaled)

	seed, err := adapter.parse(adapter.exactPrefix + "7/1")
	must.NoError(err)
	target := new(T)
	*target = seed
	must.NoError(adapter.unmarshal(target, marshaled))
	is.True(adapter.equal(original, *target), "UnmarshalText(MarshalText()) did not preserve the result")
	is.Equal(fixture.canonical, adapter.stringFn(*target))
	assertApproximateResultMetadata(t, adapter, *target, fixture)

	prefix := []byte("regression-prefix:")
	appended, err := adapter.appendText(original, append([]byte(nil), prefix...))
	must.NoError(err)
	must.Equal(string(prefix)+fixture.canonical, string(appended))

	fromAppend, err := adapter.parse(string(appended[len(prefix):]))
	must.NoError(err)
	is.True(adapter.equal(original, fromAppend), "the text added by AppendText did not parse to an equal result")
	is.Equal(fixture.canonical, adapter.stringFn(fromAppend))
	assertApproximateResultMetadata(t, adapter, fromAppend, fixture)
}

func assertApproximateResultMetadata[T any, V any](
	t *testing.T,
	adapter approximateResultAdapter[T, V],
	result T,
	fixture approximateTextFixture,
) {
	t.Helper()
	is := assert.New(t)
	must := require.New(t)

	is.True(adapter.isSet(result))

	approximate, err := adapter.approximate(result)
	must.NoError(err)
	is.True(approximate)

	places, ok, err := adapter.places(result)
	must.NoError(err)
	is.True(ok)
	is.Equal(fixture.places, places)

	value, err := adapter.value(result)
	must.NoError(err)
	is.True(adapter.valueEqual(adapter.mustValue(fixture.value), value), "unexpected result value: %v", value)

	bound, err := adapter.errorBound(result)
	must.NoError(err)
	is.True(adapter.valueEqual(adapter.mustValue(fixture.bound), bound), "unexpected error bound: %v", bound)
}

type approximateErrorFixture struct {
	name         string
	input        string
	want         error
	dependencies []error
	rejected     []error
}

func TestSecondsResultApproximateTextErrors(t *testing.T) {
	t.Parallel()
	testApproximateResultTextErrors(t, secondsApproximateResultAdapter())
}

func TestPositionResultApproximateTextErrors(t *testing.T) {
	t.Parallel()
	testApproximateResultTextErrors(t, positionApproximateResultAdapter())
}

func testApproximateResultTextErrors[T any, V any](t *testing.T, adapter approximateResultAdapter[T, V]) {
	t.Helper()

	directed76 := "1/" + decimalQ76
	approximate := func(value, places, bound string) string {
		return adapter.approximatePrefix + value + "@" + places + ":" + bound
	}
	rangeDependencies := []error{fraction.ErrOutOfRange, adapter.dependencyRangeSentinel}
	rejectRange := []error{tempomap.ErrOutOfRange, fraction.ErrOutOfRange, adapter.dependencyRangeSentinel}
	rejectInvalidText := []error{tempomap.ErrInvalidText}

	fixtures := []approximateErrorFixture{
		{
			name:         "KP-28d.3-6_reduced_value_out_of_range",
			input:        approximate(maxPlusOne+"/1", "76", directed76),
			want:         tempomap.ErrOutOfRange,
			dependencies: rangeDependencies,
			rejected:     rejectInvalidText,
		},
		{
			name:         "KP-28d.3-6_reduced_bound_range_precedes_semantic_mismatch",
			input:        approximate("10/1", "76", "1/"+maxPlusOne),
			want:         tempomap.ErrOutOfRange,
			dependencies: rangeDependencies,
			rejected:     rejectInvalidText,
		},
		{
			name:         "Additional_D77_directed_bound_out_of_range",
			input:        approximate("10/1", "77", "1/"+decimalQ77),
			want:         tempomap.ErrOutOfRange,
			dependencies: rangeDependencies,
			rejected:     rejectInvalidText,
		},
		{
			name:         "Additional_D77_nearest_bound_out_of_range",
			input:        approximate("10/1", "77", "1/"+decimal2Q77),
			want:         tempomap.ErrOutOfRange,
			dependencies: rangeDependencies,
			rejected:     rejectInvalidText,
		},
		{
			name:     "KP-28d.3-7_value_off_decimal_grid",
			input:    approximate("1/3", "76", directed76),
			want:     tempomap.ErrInvalidText,
			rejected: rejectRange,
		},
		{
			name:     "KP-28d.3-7_representable_wrong_bound",
			input:    approximate("10/1", "76", "1/"+decimalQ75),
			want:     tempomap.ErrInvalidText,
			rejected: rejectRange,
		},
		{
			name:     "KP-28d.3-7_malformed_prefix",
			input:    "x" + adapter.approximatePrefix[1:] + "10/1@76:" + directed76,
			want:     tempomap.ErrInvalidText,
			rejected: rejectRange,
		},
		{
			name:     "KP-28d.3-7_missing_places_separator",
			input:    adapter.approximatePrefix + "10/1:" + directed76,
			want:     tempomap.ErrInvalidText,
			rejected: rejectRange,
		},
		{
			name:     "KP-28d.3-7_missing_bound_separator",
			input:    adapter.approximatePrefix + "10/1@76",
			want:     tempomap.ErrInvalidText,
			rejected: rejectRange,
		},
		{
			name:     "KP-28d.3-7_extra_separator",
			input:    approximate("10/1", "76", directed76) + ":" + directed76,
			want:     tempomap.ErrInvalidText,
			rejected: rejectRange,
		},
		{
			name:     "KP-28d.3-7_negative_value",
			input:    approximate("-1/1", "76", directed76),
			want:     tempomap.ErrInvalidText,
			rejected: rejectRange,
		},
		{
			name:     "KP-28d.3-7_negative_bound",
			input:    approximate("10/1", "76", "-"+directed76),
			want:     tempomap.ErrInvalidText,
			rejected: rejectRange,
		},
		{
			name:     "KP-28d.3-7_unset_string",
			input:    adapter.unsetText,
			want:     tempomap.ErrInvalidText,
			rejected: rejectRange,
		},
		{
			name:     "KP-28d.3-7_approximate_as_exact",
			input:    adapter.exactPrefix + "10/1@76:" + directed76,
			want:     tempomap.ErrInvalidText,
			rejected: rejectRange,
		},
		{
			name:     "KP-28d.3-7_leading_zero_in_places",
			input:    approximate("10/1", "076", directed76),
			want:     tempomap.ErrInvalidText,
			rejected: rejectRange,
		},
		{
			name:     "KP-28d.3-7_signed_places",
			input:    approximate("10/1", "-1", directed76),
			want:     tempomap.ErrInvalidText,
			rejected: rejectRange,
		},
		{
			name:     "KP-28d.3-7_decimal_places_token",
			input:    approximate("10/1", "76.0", directed76),
			want:     tempomap.ErrInvalidText,
			rejected: rejectRange,
		},
		{
			name:  "KP-28d.3-7_places_above_maximum",
			input: approximate("10/1", "1001", directed76),
			want:  tempomap.ErrInvalidArgument,
			rejected: []error{
				tempomap.ErrInvalidText,
				tempomap.ErrOutOfRange,
				fraction.ErrOutOfRange,
			},
		},
		{
			name:         "KP-28d.3-7_exact_unreduced_pair_keeps_written_range_limit",
			input:        adapter.exactPrefix + "10/" + decimalQ77,
			want:         tempomap.ErrOutOfRange,
			dependencies: rangeDependencies,
			rejected:     rejectInvalidText,
		},
		{
			name:     "Priority_bound_grammar_over_deferred_value_range",
			input:    approximate(maxPlusOne+"/1", "76", "not-a-position"),
			want:     tempomap.ErrInvalidText,
			rejected: rejectRange,
		},
		{
			name:     "Priority_invalid_places_over_reduced_value_range",
			input:    approximate(maxPlusOne+"/1", "bad", directed76),
			want:     tempomap.ErrInvalidText,
			rejected: rejectRange,
		},
		{
			name:  "Priority_excess_places_over_reduced_value_range",
			input: approximate(maxPlusOne+"/1", "1001", directed76),
			want:  tempomap.ErrInvalidArgument,
			rejected: []error{
				tempomap.ErrInvalidText,
				tempomap.ErrOutOfRange,
				fraction.ErrOutOfRange,
			},
		},
		{
			name:         "Priority_value_range_over_bound_range",
			input:        approximate(maxPlusOne+"/1", "76", "1/"+maxPlusOne),
			want:         tempomap.ErrOutOfRange,
			dependencies: rangeDependencies,
			rejected:     rejectInvalidText,
		},
		{
			name:         "Priority_value_range_over_node_semantic_error",
			input:        approximate(maxPlusOne+"/3", "76", directed76),
			want:         tempomap.ErrOutOfRange,
			dependencies: rangeDependencies,
			rejected:     rejectInvalidText,
		},
		{
			name:         "Priority_bound_range_over_node_semantic_error",
			input:        approximate("1/3", "76", "1/"+maxPlusOne),
			want:         tempomap.ErrOutOfRange,
			dependencies: rangeDependencies,
			rejected:     rejectInvalidText,
		},
	}

	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			t.Parallel()
			assertApproximateResultTextError(t, adapter, fixture)
		})
	}
}

func assertApproximateResultTextError[T any, V any](
	t *testing.T,
	adapter approximateResultAdapter[T, V],
	fixture approximateErrorFixture,
) {
	t.Helper()
	is := assert.New(t)
	must := require.New(t)

	got, err := adapter.parse(fixture.input)
	must.Error(err)
	var zero T
	is.True(adapter.equal(got, zero), "a rejected parse returned a non-zero result")
	assertResultErrorChain(is, err, fixture)
	is.LessOrEqual(len(err.Error()), 512, "parse error text exceeds its contract limit")

	seed, seedErr := adapter.parse(adapter.exactPrefix + "7/1")
	must.NoError(seedErr)
	target := new(T)
	*target = seed

	unmarshalErr := adapter.unmarshal(target, []byte(fixture.input))
	must.Error(unmarshalErr)
	assertResultErrorChain(is, unmarshalErr, fixture)
	is.True(adapter.equal(seed, *target), "UnmarshalText changed its receiver on error")
	is.LessOrEqual(len(unmarshalErr.Error()), 512, "unmarshal error text exceeds its contract limit")

	panicValue := capturedPanic(func() {
		adapter.mustParse(fixture.input)
	})
	must.NotNil(panicValue, "MustParse did not panic with the base parse error")
	panicErr, ok := panicValue.(error)
	must.True(ok, "MustParse panic has type %T, want error", panicValue)
	is.EqualError(panicErr, err.Error(), "MustParse did not panic with the base parse error")
	assertResultErrorChain(is, panicErr, fixture)
}

func assertResultErrorChain(is *assert.Assertions, err error, fixture approximateErrorFixture) {
	is.ErrorIs(err, fixture.want)
	for _, dependency := range fixture.dependencies {
		is.ErrorIs(err, dependency)
	}
	for _, rejected := range fixture.rejected {
		is.NotErrorIs(err, rejected)
	}
}

func capturedPanic(fn func()) (recovered any) {
	defer func() {
		recovered = recover()
	}()
	fn()
	return nil
}

func TestMapCalculationApproximateResultsTextRoundTrips(t *testing.T) {
	t.Parallel()

	q60 := tempo.MustNewQuarter(fraction.FromInt(60))
	q120 := tempo.MustNewQuarter(fraction.FromInt(120))
	gradual := tempomap.MustNewGradual(tempomap.MustAt(duration.Origin()), tempomap.LinearSpeed).
		MustWithSlot(tempomap.SlotSecond)
	m := tempomap.MustNew(q60).
		MustWithChange(gradual).
		MustWithChange(tempomap.MustNewMain(tempomap.MustAt(duration.MustNewPosition(1, 1)), q120))

	secondsAdapter := secondsApproximateResultAdapter()
	secondsCases := []struct {
		name      string
		mode      rounding.Mode
		value     string
		bound     string
		canonical string
	}{
		{
			name:      "KP-8_half_up",
			mode:      rounding.HalfUp,
			value:     "2773/1000",
			bound:     "1/2000",
			canonical: "s~2773/1000@3:1/2000",
		},
		{
			name:      "KP-8_down",
			mode:      rounding.Down,
			value:     "2772/1000",
			bound:     "1/1000",
			canonical: "s~2772/1000@3:1/1000",
		},
	}
	for _, testCase := range secondsCases {
		t.Run("Seconds/"+testCase.name, func(t *testing.T) {
			t.Parallel()
			result := m.MustStart(tempomap.MustAt(duration.MustNewPosition(1, 1)), 3, testCase.mode)
			assertApproximateResultTextRoundTrip(t, secondsAdapter, approximateTextFixture{
				input:     testCase.canonical,
				canonical: testCase.canonical,
				value:     testCase.value,
				bound:     testCase.bound,
				places:    3,
			}, result)
		})
	}

	positionAdapter := positionApproximateResultAdapter()
	positionCases := []struct {
		name      string
		mode      rounding.Mode
		value     string
		bound     string
		canonical string
	}{
		{
			name:      "KP-10_half_up",
			mode:      rounding.HalfUp,
			value:     "284/1000",
			bound:     "1/2000",
			canonical: "p~284/1000@3:1/2000",
		},
		{
			name:      "KP-10_down",
			mode:      rounding.Down,
			value:     "284/1000",
			bound:     "1/1000",
			canonical: "p~284/1000@3:1/1000",
		},
	}
	for _, testCase := range positionCases {
		t.Run("Position/"+testCase.name, func(t *testing.T) {
			t.Parallel()
			result := m.MustPosition(tempo.MustNewSeconds(1, 1), 3, testCase.mode)
			assertApproximateResultTextRoundTrip(t, positionAdapter, approximateTextFixture{
				input:     testCase.canonical,
				canonical: testCase.canonical,
				value:     testCase.value,
				bound:     testCase.bound,
				places:    3,
			}, result)
		})
	}
}

func TestResultUnmarshalTextNilReceiver(t *testing.T) {
	t.Parallel()

	t.Run("Seconds", func(t *testing.T) {
		t.Parallel()
		var result *tempomap.SecondsResult
		err := result.UnmarshalText([]byte("s:0/1"))
		require.ErrorIs(t, err, tempomap.ErrInvalidArgument)
	})

	t.Run("Position", func(t *testing.T) {
		t.Parallel()
		var result *tempomap.PositionResult
		err := result.UnmarshalText([]byte("p:0/1"))
		require.ErrorIs(t, err, tempomap.ErrInvalidArgument)
	})
}
