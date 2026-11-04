package octave_test

import (
	"errors"
	"fmt"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/octave/number"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The tests of this file check the text of the errors of FromName, FromNumber
// and Octave.Number and of their Must variants (КР-OC16, architecture
// "Текст ошибок"). Properties T1...T6 are those of that subsection:
//
//	T1  the whole text is at most errTextMaxLen bytes;
//	T2  the error of the subpackage (u) is the whole suffix of the text;
//	T3  the context p of this package (text without u) is at most
//	    errTextMaxContext bytes and does not repeat the input;
//	T4  classification and results are as before;
//	T5  one failed call allocates at most errTextMaxAlloc bytes;
//	T6  full inputs that differ only beyond the excerpt of a diagnostic are
//	    still distinguished by the successful operations.
//
// The tests analyse only the context of this package (p) and never the text
// of the wrapped errors; the exact words of the context are not checked.
// The text of an error of a large input is not printed before a check that
// bounds its length, and a panic of a large input is recovered without
// formatting its value: before the fix such a text is megabytes long.
const (
	errTextM          = 1 << 20 // M of the architecture: skipped numbers of the large cases
	errTextKiB        = 1 << 10
	errTextShortName  = 32 // the longest name that a diagnostic of package naming keeps whole
	errTextMaxName    = 96 // the longest name that the context may not repeat as strconv.Quote(name)
	errTextMaxLen     = 512
	errTextMaxContext = 96
	errTextMaxAlloc   = 64 * 1024
	errTextAllocCalls = 8
)

// errTextCase is a failing call of one of the six paths with the properties
// to check. call is the ordinary variant: it returns whether the other result
// is the zero value (Octave{} or 0) and the error. must calls the Must
// variant. direct calls the subpackage with the same arguments and returns u;
// it is nil when the error is a sentinel of this package without u. forbid
// holds the strings that the context p must not contain.
type errTextCase struct {
	name   string
	want   cause
	call   func() (zero bool, err error)
	must   func()
	direct func() error
	forbid []string
}

// errTextRepeat is R(unit, n) of the architecture: n bytes of a one-byte
// unit, and floor(n / len(unit)) whole runes of a longer unit.
func errTextRepeat(unit string, n int) string {
	return strings.Repeat(unit, n/len(unit))
}

// errTextUnit is the repeated unit of a long name.
type errTextUnit struct {
	label string
	unit  string
	text  string // the unit as strconv.Quote writes it, without quotes
}

func errTextUnits() []errTextUnit {
	return []errTextUnit{
		{label: "nul", unit: "\x00", text: `\x00`},
		{label: "ff", unit: "\xff", text: `\xff`},
		{label: "e-diaeresis", unit: "ё", text: "ё"},
		{label: "clef", unit: "𝄞", text: "𝄞"},
	}
}

// errTextP4 is P4(n) of the architecture: the beginning of the bounded
// description of an own numbering with more than four skipped numbers, built
// from the pair and the four smallest skipped numbers.
func errTextP4(aspn, num int, skipped4 ...int) string {
	return strings.TrimSuffix(number.MustNew(aspn, num, skipped4...).String(), "]")
}

// errTextNameCase returns the case of a name that is not in the catalog. unit
// is the repeated unit of a long name, or "".
func errTextNameCase(label, name string, c *naming.Catalog, unit string) errTextCase {
	var forbid []string
	if len(name) <= errTextMaxName {
		forbid = append(forbid, strconv.Quote(name))
	}
	if unit != "" && len(name) > errTextShortName {
		forbid = append(forbid, strings.Trim(strconv.Quote(unit), `"`))
	}

	return errTextCase{
		name: label,
		want: causeUnknownName,
		call: func() (bool, error) {
			o, err := octave.FromName(name, c)

			return o == (octave.Octave{}), err
		},
		must: func() {
			_ = octave.MustFromName(name, c)
		},
		direct: func() error {
			_, err := c.ASPN(name)

			return err
		},
		forbid: forbid,
	}
}

// errTextFromNumberCase returns the case of FromNumber(num, n).
func errTextFromNumberCase(label string, num int, n number.Numbering, want cause, forbid ...string) errTextCase {
	return errTextCase{
		name: label,
		want: want,
		call: func() (bool, error) {
			o, err := octave.FromNumber(num, n)

			return o == (octave.Octave{}), err
		},
		must: func() {
			_ = octave.MustFromNumber(num, n)
		},
		direct: func() error {
			_, err := n.ToASPN(num)

			return err
		},
		forbid: forbid,
	}
}

// errTextNumberCase returns the case of o.Number(n) for a valid octave o.
func errTextNumberCase(label string, o octave.Octave, n number.Numbering, want cause, forbid ...string) errTextCase {
	return errTextCase{
		name: label,
		want: want,
		call: func() (bool, error) {
			v, err := o.Number(n)

			return v == 0, err
		},
		must: func() {
			_ = o.MustNumber(n)
		},
		direct: func() error {
			_, err := n.FromASPN(o.MustASPN())

			return err
		},
		forbid: forbid,
	}
}

// errTextDescribe describes a recovered panic value or an error without
// printing its text: the type, and for an error its length and the sentinels
// it matches.
func errTextDescribe(v any) string {
	err, ok := v.(error)
	if !ok {
		return fmt.Sprintf("value of type %T", v)
	}

	var matches []string
	for _, s := range allSentinels() {
		if errors.Is(err, s) {
			matches = append(matches, sentinelName(s))
		}
	}

	return fmt.Sprintf("error of type %T, text length %d, matches %v", err, len(err.Error()), matches)
}

// errTextResult calls f and fails the test with a bounded message if f
// panics: the value of the panic is not formatted.
func errTextResult[T any](t *testing.T, desc string, f func() (T, error)) (T, error) {
	t.Helper()

	var (
		result T
		err    error
	)

	v, panicked := panicValue(func() {
		result, err = f()
	})
	if panicked {
		require.Failf(t, "unexpected panic", "%s: %s", desc, errTextDescribe(v))
	}

	return result, err
}

// errTextRequireNoError fails the test if err is not nil without printing
// its text.
func errTextRequireNoError(t *testing.T, err error, desc string) {
	t.Helper()

	if err != nil {
		require.Failf(t, "unexpected error", "%s: %s", desc, errTextDescribe(err))
	}
}

// errTextAssertCause checks T4: err matches exactly the sentinel errors of
// want. The checks are boolean and print no text.
func errTextAssertCause(t *testing.T, err error, want cause, desc string) {
	t.Helper()

	for _, s := range allSentinels() {
		assert.Equal(t, slices.Contains(want.match, s), errors.Is(err, s), "%s: errors.Is(%s)", desc, sentinelName(s))
	}
}

// errTextAssertText checks T1, T2 and T3 of the error err of the case tc.
// The text is not printed, and p is not computed, before the check that
// bounds their length.
func errTextAssertText(t *testing.T, tc errTextCase, err error, desc string) {
	t.Helper()

	require.Error(t, err, "%s: expected an error", desc)

	var u error
	if tc.direct != nil {
		u = tc.direct()
		require.Error(t, u, "%s: the direct call returned no error", desc)

		for _, s := range tc.want.match {
			if s != octave.ErrOutOfRange {
				require.ErrorIs(t, u, s, "%s: cause of the direct call", desc)
			}
		}
	}

	msg := err.Error()
	require.LessOrEqual(t, len(msg), errTextMaxLen, "%s: length of the text (T1)", desc)

	if u == nil {
		return
	}

	require.True(t, strings.HasSuffix(msg, u.Error()), "%s: the wrapped error is not the suffix of the text (T2)", desc)

	p := strings.TrimSuffix(msg, u.Error())
	require.LessOrEqual(t, len(p), errTextMaxContext, "%s: length of the context (T3)", desc)

	for _, f := range tc.forbid {
		assert.NotContains(t, p, f, "%s: the context repeats the input (T3)", desc)
	}
}

// errTextCheckError checks the ordinary variant of tc: T4, then T1...T3.
func errTextCheckError(t *testing.T, tc errTextCase) {
	t.Helper()

	var (
		zero bool
		err  error
	)

	v, panicked := panicValue(func() {
		zero, err = tc.call()
	})
	require.False(t, panicked, "%s: unexpected panic: %s", tc.name, errTextDescribe(v))

	errTextAssertCause(t, err, tc.want, tc.name)
	assert.True(t, zero, "%s: the other result is not the zero value", tc.name)
	errTextAssertText(t, tc, err, tc.name)
}

// errTextCheckMust checks the Must variant of tc: the panic value is the
// error, then T4 and T1...T3 for it.
func errTextCheckMust(t *testing.T, tc errTextCase) {
	t.Helper()

	desc := tc.name + " (Must)"

	v, panicked := panicValue(tc.must)
	require.True(t, panicked, "%s: expected a panic", desc)

	err, ok := v.(error)
	require.True(t, ok, "%s: the panic value has type %T, not an error", desc, v)

	errTextAssertCause(t, err, tc.want, desc)
	errTextAssertText(t, tc, err, desc)
}

// errTextRun runs the ordinary and the Must variants of every case as
// parallel subtests.
func errTextRun(t *testing.T, cases []errTextCase) {
	t.Helper()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("error", func(t *testing.T) {
				t.Parallel()
				errTextCheckError(t, tc)
			})
			t.Run("Must", func(t *testing.T) {
				t.Parallel()
				errTextCheckMust(t, tc)
			})
		})
	}
}

// TestErrorTextLongNames: FromName and MustFromName for names of 1 KiB and
// 1 MiB (T1...T4); the context does not repeat the unit of the name.
func TestErrorTextLongNames(t *testing.T) {
	t.Parallel()

	catalogs := []struct {
		label string
		c     *naming.Catalog
	}{
		{label: "Russian", c: naming.Russian()},
		{label: "nil", c: nil},
	}
	sizes := []struct {
		label string
		n     int
	}{
		{label: "1KiB", n: errTextKiB},
		{label: "1MiB", n: errTextM},
	}

	var cases []errTextCase
	for _, u := range errTextUnits() {
		for _, size := range sizes {
			name := errTextRepeat(u.unit, size.n)
			for _, cat := range catalogs {
				cases = append(cases, errTextNameCase(u.label+"-"+size.label+"-"+cat.label, name, cat.c, u.unit))
			}
		}
	}

	errTextRun(t, cases)
}

// TestErrorTextShortNames: names of up to 33 bytes stay out of the context
// (T3: no strconv.Quote(name) in p) and keep the whole diagnostic (T2).
func TestErrorTextShortNames(t *testing.T) {
	t.Parallel()

	ru := naming.Russian()
	nul32 := errTextRepeat("\x00", errTextShortName)
	// 33 bytes: 31 letters that are not in the words of the context, and a
	// two-byte rune.
	z33 := errTextRepeat("z", errTextShortName-1) + "ё"

	errTextRun(t, []errTextCase{
		errTextNameCase("empty", "", ru, ""),
		errTextNameCase("x", "x", ru, ""),
		errTextNameCase("Первая", "Первая", ru, ""),
		errTextNameCase("nul-32", nul32, ru, "\x00"),
		errTextNameCase("z-31-and-e-diaeresis-33", z33, ru, ""),
	})
}

// TestErrorTextLargeNumberings: FromNumber and Octave.Number with own
// numberings of 2^20 skipped numbers (T1...T4); the context does not repeat
// the beginning of the bounded description of the numbering (T3).
func TestErrorTextLargeNumberings(t *testing.T) {
	t.Parallel()

	big := number.MustNew(0, -1, span(0, errTextM-1)...)
	bigMax := number.MustNew(0, maxInt, span(0, errTextM-1)...)
	w := number.MustNew(minInt, minInt+errTextM, span(minInt, minInt+errTextM-1)...)
	n1 := number.MustNew(100, 200, 10, 20, 30, 40, 60)

	p4Big := errTextP4(0, -1, 0, 1, 2, 3)
	p4BigMax := errTextP4(0, maxInt, 0, 1, 2, 3)
	p4W := errTextP4(minInt, minInt+errTextM, minInt, minInt+1, minInt+2, minInt+3)
	p4N1 := errTextP4(100, 200, 10, 20, 30, 40)

	errTextRun(t, []errTextCase{
		errTextFromNumberCase("FromNumber(0, Big)", 0, big, causeSkipped, p4Big),
		errTextFromNumberCase("FromNumber(MIN, BigMax)", minInt, bigMax, causeNumberOutOfRange, p4BigMax),
		errTextNumberCase("O(MAX).Number(Big)", oct(maxInt), big, causeNumberOutOfRange, p4Big),
		errTextFromNumberCase("FromNumber(MIN, W)", minInt, w, causeSkipped, p4W),
		errTextNumberCase("O(MAX).Number(W)", oct(maxInt), w, causeNumberOutOfRange, p4W),
		errTextFromNumberCase("FromNumber(60, n1)", 60, n1, causeSkipped, p4N1),
		errTextNumberCase("O(MAX).Number(n1)", oct(maxInt), n1, causeNumberOutOfRange, p4N1),
	})
}

// TestErrorTextUndefinedNumbering: the undefined numbering gives the
// sentinel error itself as the wrapped error; the context does not describe
// the numbering.
func TestErrorTextUndefinedNumbering(t *testing.T) {
	t.Parallel()

	u := number.Numbering{}

	errTextRun(t, []errTextCase{
		errTextFromNumberCase("FromNumber(MIN, U)", minInt, u, causeUndefined, u.String()),
		errTextNumberCase("O(MIN).Number(U)", oct(minInt), u, causeUndefined, u.String()),
	})
}

// TestErrorTextShortNumberings: built-in numberings and own numberings with
// up to four skipped numbers keep the whole diagnostic (T2) and stay out of
// the context (T3: no n.String() in p).
func TestErrorTextShortNumberings(t *testing.T) {
	t.Parallel()

	own4 := number.MustNew(4, 4, 0, 1)
	own4Wide := number.MustNew(0, 0, minInt, minInt+1, minInt+2, minInt+3)
	own0 := number.MustNew(4, 5)

	errTextRun(t, []errTextCase{
		errTextFromNumberCase("FromNumber(0, FB)", 0, fb(), causeSkipped, fb().String()),
		errTextFromNumberCase("FromNumber(MAX, C3)", maxInt, c3(), causeNumberOutOfRange, c3().String()),
		errTextNumberCase("O(MAX).Number(C5)", oct(maxInt), c5(), causeNumberOutOfRange, c5().String()),
		errTextNumberCase("O(MIN).Number(FB)", oct(minInt), fb(), causeNumberOutOfRange, fb().String()),
		errTextFromNumberCase("FromNumber(0, New(4, 4, 0, 1))", 0, own4, causeSkipped, own4.String()),
		errTextFromNumberCase("FromNumber(MIN, New(0, 0, MIN...MIN+3))", minInt, own4Wide, causeSkipped, own4Wide.String()),
		errTextFromNumberCase("FromNumber(MIN, New(4, 5))", minInt, own0, causeNumberOutOfRange, own0.String()),
		errTextNumberCase("O(MAX).Number(New(4, 5))", oct(maxInt), own0, causeNumberOutOfRange, own0.String()),
	})
}

// TestErrorTextInvalidOctave: an invalid octave is found before the
// numbering is looked at (T4): ErrInvalid and the zero result, whatever the
// numbering. The error is a sentinel of this package without context.
func TestErrorTextInvalidOctave(t *testing.T) {
	t.Parallel()

	big := number.MustNew(0, -1, span(0, errTextM-1)...)

	cases := []errTextCase{
		{
			name: "Z.Number(Big)",
			want: causeInvalid,
			call: func() (bool, error) {
				v, err := octave.Octave{}.Number(big)

				return v == 0, err
			},
			must: func() {
				_ = octave.Octave{}.MustNumber(big)
			},
		},
		{
			name: "Z.Number(U)",
			want: causeInvalid,
			call: func() (bool, error) {
				v, err := octave.Octave{}.Number(number.Numbering{})

				return v == 0, err
			},
			must: func() {
				_ = octave.Octave{}.MustNumber(number.Numbering{})
			},
		},
	}

	errTextRun(t, cases)
}

// TestErrorTextFullNamesFound (T6): the bounded diagnostic does not bound
// the data of the search. Names that differ only beyond the excerpt of a
// diagnostic are found and distinguished, and Name returns the full strings.
func TestErrorTextFullNamesFound(t *testing.T) {
	t.Parallel()

	prefixes := []struct {
		label  string
		prefix string
	}{
		{label: "prefix-32", prefix: errTextRepeat("\x00", errTextShortName)},
		{label: "prefix-1MiB-1", prefix: errTextRepeat("\x00", errTextM-1)},
	}

	for _, pc := range prefixes {
		t.Run(pc.label, func(t *testing.T) {
			t.Parallel()

			nameX, nameY, nameZ := pc.prefix+"x", pc.prefix+"y", pc.prefix+"z"

			c, err := octave.NewCatalog(map[octave.Octave]string{oct(1): nameX, oct(2): nameY})
			errTextRequireNoError(t, err, "NewCatalog")

			found := []struct {
				label string
				name  string
				want  octave.Octave
			}{
				{label: "x", name: nameX, want: oct(1)},
				{label: "y", name: nameY, want: oct(2)},
			}
			for _, f := range found {
				t.Run("found-"+f.label, func(t *testing.T) {
					t.Parallel()

					got, err := errTextResult(t, f.label+" FromName", func() (octave.Octave, error) {
						return octave.FromName(f.name, c)
					})
					errTextRequireNoError(t, err, f.label+" FromName")
					assert.Equal(t, f.want, got, "%s: FromName", f.label)

					got, err = errTextResult(t, f.label+" MustFromName", func() (octave.Octave, error) {
						return octave.MustFromName(f.name, c), nil
					})
					errTextRequireNoError(t, err, f.label+" MustFromName")
					assert.Equal(t, f.want, got, "%s: MustFromName", f.label)

					full, err := errTextResult(t, f.label+" Name", func() (string, error) {
						return f.want.Name(c), nil
					})
					errTextRequireNoError(t, err, f.label+" Name")
					assert.True(t, full == f.name, "%s: Name is not the full name: length %d, want %d", f.label, len(full), len(f.name))
				})
			}

			errTextRun(t, []errTextCase{errTextNameCase("not-found-z", nameZ, c, "\x00")})
		})
	}
}

// TestErrorTextFullNumberingsFound (T6): own numberings whose bounded
// descriptions are the same still translate every number exactly, and the
// large numbering is translated exactly in both directions.
func TestErrorTextFullNumberingsFound(t *testing.T) {
	t.Parallel()

	n1 := number.MustNew(100, 200, 10, 20, 30, 40, 60)
	n2 := number.MustNew(100, 200, 10, 20, 30, 40, 70)
	big := number.MustNew(0, -1, span(0, errTextM-1)...)

	cases := []struct {
		name string
		num  int
		n    number.Numbering
		want octave.Octave
	}{
		{name: "FromNumber(60, n2)", num: 60, n: n2, want: oct(n2.MustToASPN(60))},
		{name: "FromNumber(70, n1)", num: 70, n: n1, want: oct(n1.MustToASPN(70))},
		{name: "FromNumber(-1, Big)", num: -1, n: big, want: oct(0)},
		{name: "FromNumber(1048576, Big)", num: errTextM, n: big, want: oct(1)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := errTextResult(t, tc.name, func() (octave.Octave, error) {
				return octave.FromNumber(tc.num, tc.n)
			})
			errTextRequireNoError(t, err, tc.name)
			assert.Equal(t, tc.want, got, "%s: FromNumber", tc.name)

			got, err = errTextResult(t, tc.name+" Must", func() (octave.Octave, error) {
				return octave.MustFromNumber(tc.num, tc.n), nil
			})
			errTextRequireNoError(t, err, tc.name+" Must")
			assert.Equal(t, tc.want, got, "%s: MustFromNumber", tc.name)

			num, err := errTextResult(t, tc.name+" Number", func() (int, error) {
				return tc.want.Number(tc.n)
			})
			errTextRequireNoError(t, err, tc.name+" Number")
			assert.Equal(t, tc.num, num, "%s: Number", tc.name)

			num, err = errTextResult(t, tc.name+" MustNumber", func() (int, error) {
				return tc.want.MustNumber(tc.n), nil
			})
			errTextRequireNoError(t, err, tc.name+" MustNumber")
			assert.Equal(t, tc.num, num, "%s: MustNumber", tc.name)
		})
	}
}

// TestErrorTextOracle checks the oracle of T3 against literals: the
// beginning of a bounded description and the forms of the repeated units.
func TestErrorTextOracle(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "ASPN 0 -> -1, skip [0 1 2 3", errTextP4(0, -1, 0, 1, 2, 3))
	assert.Equal(t, "ASPN 100 -> 200, skip [10 20 30 40", errTextP4(100, 200, 10, 20, 30, 40))

	for _, u := range errTextUnits() {
		assert.Equal(t, u.text, strings.Trim(strconv.Quote(u.unit), `"`), "unit %s", u.label)
	}

	assert.Len(t, errTextRepeat("\x00", errTextKiB), errTextKiB)
	assert.Len(t, errTextRepeat("ё", errTextM), errTextM)
	assert.Len(t, errTextRepeat("𝄞", errTextM), errTextM)
}

// TestErrorTextAllocations (T5): one failed operation allocates at most
// errTextMaxAlloc bytes on average, whatever the size of the input. The test
// is sequential by design: TotalAlloc counts all goroutines of the process,
// and parallel tests are not run while a sequential test runs. The inputs are
// prepared before the measurement.
func TestErrorTextAllocations(t *testing.T) {
	name := errTextRepeat("\x00", errTextM)
	ru := naming.Russian()
	big := number.MustNew(0, -1, span(0, errTextM-1)...)
	bigMax := number.MustNew(0, maxInt, span(0, errTextM-1)...)

	cases := []struct {
		name string
		call func() error
	}{
		{name: "FromName(nul-1MiB, Russian)", call: func() error {
			_, err := octave.FromName(name, ru)

			return err
		}},
		{name: "FromNumber(0, Big)", call: func() error {
			_, err := octave.FromNumber(0, big)

			return err
		}},
		{name: "FromNumber(MIN, BigMax)", call: func() error {
			_, err := octave.FromNumber(minInt, bigMax)

			return err
		}},
		{name: "O(MAX).Number(Big)", call: func() error {
			_, err := oct(maxInt).Number(big)

			return err
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				before, after runtime.MemStats
				failures      int
			)

			// Both parts are recovered: a panic with a large error must not
			// reach the test runner, which would print the whole value. One
			// recover frame wraps the whole measured loop, so no per-call
			// wrapper allocates inside the measurement; the only extra
			// allocation is the closure of that frame, made once before the
			// first ReadMemStats of the loop.
			v, panicked := panicValue(func() { // warm-up of the pools of package fmt
				if tc.call() == nil {
					failures++
				}
			})
			require.False(t, panicked, "%s: unexpected panic in the warm-up call: %s", tc.name, errTextDescribe(v))

			v, panicked = panicValue(func() {
				runtime.ReadMemStats(&before)
				for range errTextAllocCalls {
					if tc.call() == nil {
						failures++
					}
				}
				runtime.ReadMemStats(&after)
			})
			require.False(t, panicked, "%s: unexpected panic in a measured call: %s", tc.name, errTextDescribe(v))

			perCall := (after.TotalAlloc - before.TotalAlloc) / errTextAllocCalls
			t.Logf("%s: %d bytes per failed call", tc.name, perCall)

			assert.Zero(t, failures, "%s: calls without an error", tc.name)
			assert.LessOrEqual(t, perCall, uint64(errTextMaxAlloc), "%s: bytes per failed call (T5)", tc.name)
		})
	}
}
