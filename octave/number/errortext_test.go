package number_test

import (
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/go-muse/musex/octave/number"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests of the bounded description of a numbering in the errors of
// FromASPN, ToASPN, Convert and their Must variants (architecture of task
// 4.3, KR-O15, properties T1-T5).
//
// A numbering with a million skipped numbers gives an error text of more than
// seven megabytes until the description is bounded. So no assertion prints the
// text of an error of a big numbering: T4 and the length (T1, with require)
// come first, and the substrings are checked with short messages.

const (
	// errTextLimit is the bound T1 on the length of the text of an error.
	errTextLimit = 512
	// allocLimit is the bound T5 on the bytes allocated to create one error.
	allocLimit = 65536
	// tailLimit is the bound of the format F3 on the tail of a long
	// description.
	tailLimit = 40
	// bigCount is M of the architecture: the count of skipped numbers of a
	// big numbering.
	bigCount = 1 << 20
	// separator is the text between the descriptions of from and to in the
	// text of a Convert overflow error (format F1).
	separator = " to "
)

// side tells which description in the text of an error is checked.
type side int

const (
	// sideOnly is the only description: the text ends with it.
	sideOnly side = iota
	// sideFrom is the description of from in a Convert overflow error.
	sideFrom
	// sideTo is the description of to in a Convert overflow error: it ends
	// the text.
	sideTo
)

func (s side) String() string {
	switch s {
	case sideOnly:
		return "only description"
	case sideFrom:
		return "from"
	case sideTo:
		return "to"
	}

	return "unknown side"
}

// tailProblem is the oracle of T3. It selects the description of the side in
// the text msg and checks the format F3 of a long description with the
// leading part p4 and the count k of skipped numbers. It returns "" when the
// description is acceptable, otherwise the reason. The reason never contains
// msg.
//
// The part of the text with the description is the whole text, or the text
// before the last separator for from, or the text after the last separator
// for to. The description of to starts that part; on other parts it starts at
// the last occurrence of p4. Everything after p4 is the tail.
func tailProblem(msg string, s side, p4 string, k int) string {
	part := msg
	if s != sideOnly {
		j := strings.LastIndex(msg, separator)
		if j < 0 {
			return "no separator between the descriptions"
		}

		if s == sideFrom {
			part = msg[:j]
		} else {
			part = msg[j+len(separator):]
		}
	}

	start := 0
	if s == sideTo {
		if !strings.HasPrefix(part, p4) {
			return "the description of to does not start the part after the last separator"
		}
	} else if start = strings.LastIndex(part, p4); start < 0 {
		return "no description start"
	}

	tail := part[start+len(p4):]
	count := strconv.Itoa(k)

	switch {
	case len(tail) > tailLimit:
		return fmt.Sprintf("the tail of %d bytes is longer than %d bytes", len(tail), tailLimit)
	case strings.Contains(tail, separator):
		return "the tail contains the separator"
	case !strings.Contains(tail, count):
		return fmt.Sprintf("the tail %q does not contain the count %s", tail, count)
	case strings.ContainsAny(strings.Replace(tail, count, "", 1), "0123456789"):
		return fmt.Sprintf("the tail %q has digits other than the count %s", tail, count)
	}

	return ""
}

// The oracle accepts the texts that follow the format and rejects the texts
// of the mutants. The texts are made up for the oracle and are not the format
// of the package. The accepted limitations (L1-L3) are the texts that cannot
// be told from a correct tail by the format.
func TestTailProblem(t *testing.T) {
	t.Parallel()

	const (
		skippedPrefix  = "number: number is skipped in the numbering: number 10 in "
		overflowPrefix = "number: number out of range: number -9223372036854775808 converted from "
		p4a            = "ASPN 100 -> 200, skip [10 20 30 40"
		p4b            = "ASPN 100 -> 200, skip [1 2 3 4"
		k5             = " ...] (5 skipped)"
	)

	p4w := strings.TrimSuffix(
		number.MustNew(minInt, minInt+bigCount, minInt, minInt+1, minInt+2, minInt+3).String(), "]")

	tests := []struct {
		name   string
		msg    string
		side   side
		p4     string
		k      int
		accept bool
	}{
		{"C1 one description", skippedPrefix + p4a + k5, sideOnly, p4a, 5, true},
		{
			"C2 short from with the start of to", overflowPrefix + "ASPN 100 -> 200, skip [10 20 30 40] to " + p4a + k5,
			sideTo, p4a, 5, true,
		},
		{"C3 k equals the fifth skipped number", skippedPrefix + p4b + " 5 skipped ...]", sideOnly, p4b, 5, true},
		{"C4 equal starts, from", overflowPrefix + p4a + k5 + " to " + p4a + " ...] (6 skipped)", sideFrom, p4a, 5, true},
		{"C5 equal starts, to", overflowPrefix + p4a + k5 + " to " + p4a + " ...] (6 skipped)", sideTo, p4a, 6, true},
		{"C6 tail of 40 bytes", skippedPrefix + p4a + k5 + strings.Repeat("x", 23), sideOnly, p4a, 5, true},
		{"C7 other wording", skippedPrefix + p4a + " …] (total 5)", sideOnly, p4a, 5, true},
		{"C8 start in the text before", "context " + p4a + " in " + p4a + k5, sideOnly, p4a, 5, true},
		{"C9 short to", overflowPrefix + p4a + k5 + " to Franco-Belgian", sideFrom, p4a, 5, true},
		{"C10 long numbers", skippedPrefix + p4w + " ...] (1048576 skipped)", sideOnly, p4w, bigCount, true},

		{"R1 full form", skippedPrefix + p4a + " 60]", sideOnly, p4a, 5, false},
		{"R2 five skipped numbers", skippedPrefix + p4a + " 60 ...] (5 skipped)", sideOnly, p4a, 5, false},
		{"R3 five skipped numbers without the count", skippedPrefix + p4a + " 60 ...]", sideOnly, p4a, 5, false},
		{"R4 no count", skippedPrefix + p4a + " ...]", sideOnly, p4a, 5, false},
		{"R5 extra number", skippedPrefix + p4a + " ...] (5 of 60 skipped)", sideOnly, p4a, 5, false},
		{"R6 count twice", skippedPrefix + p4a + " ...] (5 skipped, 5 total)", sideOnly, p4a, 5, false},
		{"R7 tail of 41 bytes", skippedPrefix + p4a + k5 + strings.Repeat("x", 24), sideOnly, p4a, 5, false},
		{"R8 long tail", skippedPrefix + p4a + k5 + strings.Repeat("x", 50), sideOnly, p4a, 5, false},
		{"R9 separator in the tail", skippedPrefix + p4a + k5 + " to annotation", sideOnly, p4a, 5, false},
		{
			"R10 separator in the tail of to", overflowPrefix + "ASPN 100 -> 200, skip [10 20 30 40] to " + p4a + k5 + " to annotation",
			sideTo, p4a, 5, false,
		},
		{"R11 extra separator in from", overflowPrefix + p4a + k5 + " to x" + " to ASPN", sideFrom, p4a, 5, false},
		{"R12 text before the description of to", overflowPrefix + "ASPN to numbering " + p4a + k5, sideTo, p4a, 5, false},
		{
			"R13 three skipped numbers", skippedPrefix + "ASPN 100 -> 200, skip [10 20 30 ...] (5 skipped)",
			sideOnly, p4a, 5, false,
		},
		{"R14 no separator", overflowPrefix + p4a + k5, sideTo, p4a, 5, false},

		{"L1 full form when k is the fifth skipped number", skippedPrefix + p4b + " 5]", sideOnly, p4b, 5, true},
		{"L2 five skipped numbers when k is the fifth skipped number", skippedPrefix + p4b + " 5 ...]", sideOnly, p4b, 5, true},
		{"L3 text after the description", skippedPrefix + p4a + k5 + " (ASPN overflow)", sideOnly, p4a, 5, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			problem := tailProblem(tc.msg, tc.side, tc.p4, tc.k)
			if tc.accept {
				assert.Empty(t, problem, "the oracle must accept the text")
			} else {
				assert.NotEmpty(t, problem, "the oracle must reject the text")
			}
		})
	}
}

// The leading part of a long description is the anchor pair and the four
// smallest skipped numbers, whatever the order of the arguments of New.
func TestLeadingPart(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		n    numbering
		want string
	}{
		{"consecutive", own(t, 0, -1, 0, 1, 2, 3, 4), "ASPN 0 -> -1, skip [0 1 2 3"},
		{"ascending", own(t, 100, 200, 10, 20, 30, 40, 60), "ASPN 100 -> 200, skip [10 20 30 40"},
		{"descending", own(t, 100, 200, 60, 40, 30, 20, 10), "ASPN 100 -> 200, skip [10 20 30 40"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, leadingPart(tc.n))
		})
	}
}

// isLong reports whether the description of n in an error is the long one:
// n has more than four skipped numbers.
func isLong(n numbering) bool {
	return len(n.spec.skipped) > 4
}

// leadingPart is P4 of the architecture for a numbering with more than four
// skipped numbers: the form of String up to the closing bracket, with the
// anchor pair and the four smallest skipped numbers.
func leadingPart(n numbering) string {
	s := n.spec

	return strings.TrimSuffix(number.MustNew(s.aspn, s.num, s.skipped[:4]...).String(), "]")
}

// consecutive returns count numbers first, first+1, and so on.
func consecutive(first, count int) []int {
	s := make([]int, count)
	for i := range s {
		s[i] = first + i
	}

	return s
}

// bigNumbering creates a numbering with a long list of skipped numbers. Unlike
// own, it names the numbering by a short name: the list is never printed.
func bigNumbering(t testing.TB, name string, aspn, num int, skipped []int) numbering {
	t.Helper()

	n, err := number.New(aspn, num, skipped...)
	require.NoError(t, err, "New for %s", name)

	return numbering{name: name, n: n, spec: newSpec(aspn, num, skipped...)}
}

// textCheck is a description in the text of an error that is checked: the
// numbering described and the side of the text.
type textCheck struct {
	n    numbering
	side side
}

// errCase is a failing call with the property checks of its error.
type errCase struct {
	name string
	want error
	// oracle is the expected outcome by the mathematical definition; it is
	// not checked for the calls with an invalid definition.
	oracle outcome
	call   func() (int, error)
	// must is the Must variant of the call.
	must  func()
	texts []textCheck
}

func toASPNCase(name string, n numbering, x int, want error) errCase {
	return errCase{
		name:   name,
		want:   want,
		oracle: n.spec.toASPN(x),
		call:   func() (int, error) { return n.n.ToASPN(x) },
		must:   func() { n.n.MustToASPN(x) },
		texts:  []textCheck{{n, sideOnly}},
	}
}

func fromASPNCase(name string, n numbering, aspn int, want error) errCase {
	return errCase{
		name:   name,
		want:   want,
		oracle: n.spec.fromASPN(aspn),
		call:   func() (int, error) { return n.n.FromASPN(aspn) },
		must:   func() { n.n.MustFromASPN(aspn) },
		texts:  []textCheck{{n, sideOnly}},
	}
}

// convertCase is a Convert call that fails with ErrSkipped (the text
// describes from) or with ErrOutOfRange (the text describes from and to).
func convertCase(name string, x int, from, to numbering, want error) errCase {
	texts := []textCheck{{from, sideOnly}}
	if want == number.ErrOutOfRange {
		texts = []textCheck{{from, sideFrom}, {to, sideTo}}
	}

	return errCase{
		name:   name,
		want:   want,
		oracle: convertSpec(x, from.spec, to.spec),
		call:   func() (int, error) { return number.Convert(x, from.n, to.n) },
		must:   func() { number.MustConvert(x, from.n, to.n) },
		texts:  texts,
	}
}

// undefinedCase is a Convert call with an undefined numbering: the error is
// returned before the numberings are described.
func undefinedCase(name string, x int, from, to number.Numbering) errCase {
	return errCase{
		name:   name,
		want:   number.ErrUndefined,
		oracle: outcome{err: number.ErrUndefined},
		call:   func() (int, error) { return number.Convert(x, from, to) },
		must:   func() { number.MustConvert(x, from, to) },
	}
}

// invalidCase is a New call with an invalid definition.
func invalidCase(name string, aspn, num int, skipped []int) errCase {
	return errCase{
		name: name,
		want: number.ErrInvalidDefinition,
		call: func() (int, error) {
			_, err := number.New(aspn, num, skipped...)

			return 0, err
		},
		must: func() { number.MustNew(aspn, num, skipped...) },
	}
}

// thresholdCases are the calls with an error on an own numbering around the
// threshold of four skipped numbers (architecture, "Boundary cases").
func thresholdCases(n numbering) []errCase {
	aspn := builtins()[idxASPN]
	prefix := n.name + " "

	return []errCase{
		toASPNCase(prefix+"ToASPN(10)", n, 10, number.ErrSkipped),
		fromASPNCase(prefix+"FromASPN(MaxInt)", n, maxInt, number.ErrOutOfRange),
		toASPNCase(prefix+"ToASPN(MinInt)", n, minInt, number.ErrOutOfRange),
		convertCase(prefix+"Convert(MaxInt, ASPN, n)", maxInt, aspn, n, number.ErrOutOfRange),
		convertCase(prefix+"Convert(MinInt, n, ASPN)", minInt, n, aspn, number.ErrOutOfRange),
		convertCase(prefix+"Convert(10, n, ASPN)", 10, n, aspn, number.ErrSkipped),
	}
}

// errorTextCases are the failing calls of the architecture table "Boundary
// cases" for the text of errors (task 4.3). Big, BigMax and W are created
// once and shared by the subtests: numberings are immutable.
func errorTextCases(t *testing.T) []errCase {
	t.Helper()

	var (
		table  = builtins()
		aspn   = table[idxASPN]
		franco = table[idxFrancoBelgian]
		middC3 = table[idxMiddleC3]
		middC5 = table[idxMiddleC5]

		big    = bigNumbering(t, "Big", 0, -1, consecutive(0, bigCount))
		bigMax = bigNumbering(t, "BigMax", 0, maxInt, consecutive(0, bigCount))
		w      = bigNumbering(t, "W", minInt, minInt+bigCount, consecutive(minInt, bigCount))

		k4    = own(t, 100, 200, 10, 20, 30, 40)
		k5    = own(t, 100, 200, 10, 20, 30, 40, 60)
		k5rev = own(t, 100, 200, 60, 40, 30, 20, 10)
		k6    = own(t, 100, 200, 10, 20, 30, 40, 60, 70)
	)

	k5rev.name = "New(100, 200, 60, 40, 30, 20, 10)"

	wide4 := own(t, 0, 0, minInt, minInt+1, minInt+2, minInt+3)
	wide4.name = "New(0, 0, MinInt..MinInt+3)"
	noSkips := own(t, 0, maxInt)

	repeated := append(consecutive(0, bigCount), 5)

	cases := []errCase{
		toASPNCase("Big ToASPN(0)", big, 0, number.ErrSkipped),
		fromASPNCase("Big FromASPN(MaxInt)", big, maxInt, number.ErrOutOfRange),
		convertCase("Convert(MaxInt, ASPN, Big)", maxInt, aspn, big, number.ErrOutOfRange),
		convertCase("Convert(0, Big, ASPN)", 0, big, aspn, number.ErrSkipped),
		toASPNCase("BigMax ToASPN(MinInt)", bigMax, minInt, number.ErrOutOfRange),
		convertCase("Convert(MinInt, BigMax, ASPN)", minInt, bigMax, aspn, number.ErrOutOfRange),
		convertCase("Convert(MinInt, BigMax, Big)", minInt, bigMax, big, number.ErrOutOfRange),
		convertCase("Convert(MinInt, k=4, k=5): short from, long to", minInt, k4, k5, number.ErrOutOfRange),
		convertCase("Convert(MinInt, k=5, k=6): equal starts", minInt, k5, k6, number.ErrOutOfRange),

		toASPNCase("W ToASPN(MinInt)", w, minInt, number.ErrSkipped),
		fromASPNCase("W FromASPN(MaxInt)", w, maxInt, number.ErrOutOfRange),
		convertCase("Convert(MinInt, W, ASPN)", minInt, w, aspn, number.ErrSkipped),
		convertCase("Convert(MaxInt, ASPN, W)", maxInt, aspn, w, number.ErrOutOfRange),
		convertCase("Convert(MinInt+M, W, Middle C = C3)", minInt+bigCount, w, middC3, number.ErrOutOfRange),

		toASPNCase("New(0, 0, MinInt..MinInt+3) ToASPN(MinInt)", wide4, minInt, number.ErrSkipped),
		fromASPNCase("New(0, MaxInt) FromASPN(1)", noSkips, 1, number.ErrOutOfRange),

		toASPNCase("Franco-Belgian ToASPN(0)", franco, 0, number.ErrSkipped),
		fromASPNCase("Middle C = C3 FromASPN(MinInt)", middC3, minInt, number.ErrOutOfRange),
		toASPNCase("Middle C = C3 ToASPN(MaxInt)", middC3, maxInt, number.ErrOutOfRange),
		convertCase("Convert(MaxInt, Middle C = C3, Middle C = C5)", maxInt, middC3, middC5, number.ErrOutOfRange),
		convertCase("Convert(0, Franco-Belgian, ASPN)", 0, franco, aspn, number.ErrSkipped),

		undefinedCase("Convert(0, Big, undefined)", 0, big.n, number.Numbering{}),
		undefinedCase("Convert(0, undefined, Big)", 0, number.Numbering{}, big.n),

		invalidCase("New(0, 0, 0..M-1): anchor among skipped", 0, 0, consecutive(0, bigCount)),
		invalidCase("New(0, -1, 0..M-1, 5): repeated", 0, -1, repeated),
	}

	for _, n := range []numbering{k4, k5, k5rev} {
		cases = append(cases, thresholdCases(n)...)
	}

	return cases
}

// checkErrorText checks T4, T1, T2 and T3 on the error of a failing call.
// T4 does not print the text; T1 comes next, so that a text of seven
// megabytes is not inspected or printed.
func checkErrorText(t *testing.T, tc errCase, err error) {
	t.Helper()

	require.Error(t, err, "%s: an error is expected", tc.name)

	assert.True(t, errors.Is(err, tc.want), "T4: %s must match %v", tc.name, tc.want)
	for _, other := range allErrors() {
		if other != tc.want {
			assert.False(t, errors.Is(err, other), "T4: %s must not match %v", tc.name, other)
		}
	}

	msg := err.Error()
	require.LessOrEqual(t, len(msg), errTextLimit, "T1: length of the error text of %s", tc.name)

	for _, check := range tc.texts {
		if isLong(check.n) {
			assertBoundedDescription(t, tc.name, msg, check)
		} else {
			assertWholeDescription(t, tc.name, msg, check)
		}
	}
}

// assertWholeDescription checks T2: the description of a built-in numbering
// or of an own one with at most four skipped numbers is its String, in the
// place of the format F1.
func assertWholeDescription(t *testing.T, name, msg string, check textCheck) {
	t.Helper()

	want := check.n.n.String()

	var ok bool
	switch check.side {
	case sideOnly:
		ok = strings.HasSuffix(msg, want)
	case sideTo:
		ok = strings.HasSuffix(msg, separator+want)
	case sideFrom:
		ok = strings.Contains(msg, want+separator)
	}

	assert.True(t, ok, "T2: the error text of %s must have the whole description %q (%s)", name, want, check.side)
}

// assertBoundedDescription checks T3 for a numbering with more than four
// skipped numbers, after the precondition on the data: the record of the
// count k must differ from the records of the anchor pair and of the five
// smallest skipped numbers.
func assertBoundedDescription(t *testing.T, name, msg string, check textCheck) {
	t.Helper()

	s := check.n.spec
	k := len(s.skipped)
	count := strconv.Itoa(k)

	for _, v := range append([]int{s.aspn, s.num}, s.skipped[:5]...) {
		require.NotEqual(t, count, strings.TrimPrefix(strconv.Itoa(v), "-"),
			"T3 precondition: the record of the count %d of %s equals the record of %d", k, name, v)
	}

	assert.Empty(t, tailProblem(msg, check.side, leadingPart(check.n), k),
		"T3: bounded description of %s (%s, count %d)", name, check.side, k)
}

// T1-T4 on all paths of the architecture table "Where it applies" and the
// Must variants: the error text is short and describes a big numbering by its
// leading part and the count, a short numbering whole; the classification and
// the result are the same as for small numberings.
func TestErrorText(t *testing.T) {
	t.Parallel()

	for _, tc := range errorTextCases(t) {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.oracle.err != nil {
				require.Equal(t, outcome{err: tc.want}, tc.oracle, "oracle")
			}

			t.Run("error", func(t *testing.T) {
				t.Parallel()

				got, err := tc.call()
				assert.Zero(t, got, "%s: the result on error", tc.name)
				checkErrorText(t, tc, err)
			})

			t.Run("Must", func(t *testing.T) {
				t.Parallel()

				v, panicked := panicValue(tc.must)
				require.True(t, panicked, "%s: Must must panic", tc.name)

				err, ok := v.(error)
				require.True(t, ok, "%s: the panic value (%T) must be an error", tc.name, v)
				checkErrorText(t, tc, err)
			})
		})
	}
}

// T5: creating the error of a numbering with a million skipped numbers
// allocates a bounded amount of memory, so the text is not built whole and
// then cut. The numberings are created before the measurement. The test and
// its subtests are not parallel: parallel tests are paused while it runs and
// do not add to the allocation counter.
//
// The count is not deterministic (the pools of fmt, the race detector and the
// coverage change it), so only the upper bound is checked.
func TestErrorAllocations(t *testing.T) {
	const calls = 8

	var (
		aspn   = builtins()[idxASPN]
		big    = bigNumbering(t, "Big", 0, -1, consecutive(0, bigCount))
		bigMax = bigNumbering(t, "BigMax", 0, maxInt, consecutive(0, bigCount))
	)

	tests := []struct {
		name string
		op   func() error
	}{
		{"Big ToASPN(0)", func() error { _, err := big.n.ToASPN(0); return err }},
		{"Big FromASPN(MaxInt)", func() error { _, err := big.n.FromASPN(maxInt); return err }},
		{"Convert(0, Big, ASPN)", func() error { _, err := number.Convert(0, big.n, aspn.n); return err }},
		{"Convert(MaxInt, ASPN, Big)", func() error { _, err := number.Convert(maxInt, aspn.n, big.n); return err }},
		{"Convert(MinInt, BigMax, ASPN)", func() error { _, err := number.Convert(minInt, bigMax.n, aspn.n); return err }},
		{"Convert(MinInt, BigMax, Big)", func() error { _, err := number.Convert(minInt, bigMax.n, big.n); return err }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Error(t, tc.op(), "%s: warm-up call must fail", tc.name)

			var (
				before, after runtime.MemStats
				succeeded     int
			)

			runtime.ReadMemStats(&before)
			for range calls {
				if tc.op() == nil {
					succeeded++
				}
			}
			runtime.ReadMemStats(&after)

			perCall := (after.TotalAlloc - before.TotalAlloc) / calls
			t.Logf("T5 %s: %d bytes per error", tc.name, perCall)

			require.Zero(t, succeeded, "%s: calls without an error", tc.name)
			assert.LessOrEqual(t, perCall, uint64(allocLimit), "T5: bytes allocated per error of %s", tc.name)
		})
	}
}
