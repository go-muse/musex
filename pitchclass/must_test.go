package pitchclass_test

import (
	"iter"
	"math/big"
	"testing"

	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mustCase is one call of an operation with an error and of its Must
// variant on the same inputs. want is nil for a successful call, otherwise
// the only sentinel error both report.
type mustCase struct {
	op   string // the Must variant, e.g. "MustNew"
	name string
	want error
	call func() (any, error)
	must func() any
}

// mustCases returns, for each of the 37 operations with an error, calls on
// valid inputs and on inputs of every error the operation reports, including
// the order of checks when several causes apply (architecture, "Errors").
func mustCases(t testing.TB) []mustCase {
	t.Helper()

	c12, d12, c19 := newClass(t, 1, 12), newClass(t, 4, 12), newClass(t, 1, 19)
	f12, f19 := pitchclass.MustTransposition(5, 12), pitchclass.MustTransposition(5, 19)
	var zc pitchclass.Class
	var zo pitchclass.Operator

	errInvalidClass := pitchclass.ErrInvalidClass
	errModulus := pitchclass.ErrInvalidModulus
	errMismatch := pitchclass.ErrModulusMismatch
	errNotInvertible := pitchclass.ErrNotInvertible
	errValue := pitchclass.ErrInvalidValue
	errNotation := pitchclass.ErrNotation
	errOverflow := pitchclass.ErrOverflow

	class := func(f func() (pitchclass.Class, error)) func() (any, error) {
		return func() (any, error) { return f() }
	}
	classMust := func(f func() pitchclass.Class) func() any { return func() any { return f() } }
	operator := func(f func() (pitchclass.Operator, error)) func() (any, error) {
		return func() (any, error) { return f() }
	}
	operatorMust := func(f func() pitchclass.Operator) func() any { return func() any { return f() } }
	integer := func(f func() (int, error)) func() (any, error) { return func() (any, error) { return f() } }
	integerMust := func(f func() int) func() any { return func() any { return f() } }

	newCase := func(n, s int, want error) mustCase {
		return mustCase{
			op: "MustNew", name: "New", want: want,
			call: class(func() (pitchclass.Class, error) { return pitchclass.New(n, s) }),
			must: classMust(func() pitchclass.Class { return pitchclass.MustNew(n, s) }),
		}
	}
	reduceCase := func(x, s int, want error) mustCase {
		return mustCase{
			op: "MustReduce", name: "Reduce", want: want,
			call: class(func() (pitchclass.Class, error) { return pitchclass.Reduce(x, s) }),
			must: classMust(func() pitchclass.Class { return pitchclass.MustReduce(x, s) }),
		}
	}
	allCase := func(s int, want error) mustCase {
		return mustCase{
			op: "MustAll", name: "All", want: want,
			call: func() (any, error) { return pitchclass.All(s) },
			must: func() any { return pitchclass.MustAll(s) },
		}
	}
	parseCase := func(in string, want error) mustCase {
		return mustCase{
			op: "MustParse", name: "Parse " + in, want: want,
			call: class(func() (pitchclass.Class, error) { return pitchclass.Parse(in) }),
			must: classMust(func() pitchclass.Class { return pitchclass.MustParse(in) }),
		}
	}
	parseNotationCase := func(in string, st pitchclass.Style, s int, want error) mustCase {
		return mustCase{
			op: "MustParseNotation", name: "ParseNotation " + in, want: want,
			call: class(func() (pitchclass.Class, error) { return pitchclass.ParseNotation(in, st, s) }),
			must: classMust(func() pitchclass.Class { return pitchclass.MustParseNotation(in, st, s) }),
		}
	}
	pairCase := func(op string, x, y pitchclass.Class, want error) mustCase {
		var call func() (any, error)
		var must func() any

		switch op {
		case "MustIntervalTo":
			call, must = integer(func() (int, error) { return x.IntervalTo(y) }), integerMust(func() int { return x.MustIntervalTo(y) })
		case "MustIntervalClass (method)":
			call, must = integer(func() (int, error) { return x.IntervalClass(y) }), integerMust(func() int { return x.MustIntervalClass(y) })
		case "MustIndexSum":
			call, must = integer(func() (int, error) { return x.IndexSum(y) }), integerMust(func() int { return x.MustIndexSum(y) })
		case "MustInvertAround":
			call = class(func() (pitchclass.Class, error) { return x.InvertAround(y) })
			must = classMust(func() pitchclass.Class { return x.MustInvertAround(y) })
		case "MustInversionSwapping":
			call = operator(func() (pitchclass.Operator, error) { return pitchclass.InversionSwapping(x, y) })
			must = operatorMust(func() pitchclass.Operator { return pitchclass.MustInversionSwapping(x, y) })
		default:
			t.Fatalf("unknown pair operation %s", op)
		}

		return mustCase{op: op, name: op, want: want, call: call, must: must}
	}
	swappingCase := func(c, x, y pitchclass.Class, want error) mustCase {
		return mustCase{
			op: "MustInvertSwapping", name: "InvertSwapping", want: want,
			call: class(func() (pitchclass.Class, error) { return c.InvertSwapping(x, y) }),
			must: classMust(func() pitchclass.Class { return c.MustInvertSwapping(x, y) }),
		}
	}
	notationCase := func(c pitchclass.Class, st pitchclass.Style, want error) mustCase {
		return mustCase{
			op: "MustNotation", name: "Notation", want: want,
			call: func() (any, error) { return c.Notation(st) },
			must: func() any { return c.MustNotation(st) },
		}
	}
	bySizeCase := func(op string, i, s int, want error) mustCase {
		var call func() (any, error)
		var must func() any

		switch op {
		case "MustReduceInterval":
			call, must = integer(func() (int, error) { return pitchclass.ReduceInterval(i, s) }), integerMust(func() int { return pitchclass.MustReduceInterval(i, s) })
		case "MustComplementInterval":
			call, must = integer(func() (int, error) { return pitchclass.ComplementInterval(i, s) }), integerMust(func() int { return pitchclass.MustComplementInterval(i, s) })
		case "MustIntervalClass":
			call, must = integer(func() (int, error) { return pitchclass.IntervalClass(i, s) }), integerMust(func() int { return pitchclass.MustIntervalClass(i, s) })
		case "MustInverseMultiplier":
			call, must = integer(func() (int, error) { return pitchclass.InverseMultiplier(i, s) }), integerMust(func() int { return pitchclass.MustInverseMultiplier(i, s) })
		default:
			t.Fatalf("unknown operation %s", op)
		}

		return mustCase{op: op, name: op, want: want, call: call, must: must}
	}
	fixedCase := func(i, s int, want error) mustCase {
		return mustCase{
			op: "MustInversionFixedClasses", name: "InversionFixedClasses", want: want,
			call: func() (any, error) { return pitchclass.InversionFixedClasses(i, s) },
			must: func() any { return pitchclass.MustInversionFixedClasses(i, s) },
		}
	}
	invertibleCase := func(m, s int, want error) mustCase {
		return mustCase{
			op: "MustInvertible", name: "Invertible", want: want,
			call: func() (any, error) { return pitchclass.Invertible(m, s) },
			must: func() any { return pitchclass.MustInvertible(m, s) },
		}
	}
	multipliersCase := func(s int, want error) mustCase {
		return mustCase{
			op: "MustInvertibleMultipliers", name: "InvertibleMultipliers", want: want,
			call: func() (any, error) { return pitchclass.InvertibleMultipliers(s) },
			must: func() any { return pitchclass.MustInvertibleMultipliers(s) },
		}
	}
	modulusCase := func(op string, s int, want error) mustCase {
		var call func() (any, error)
		var must func() any

		switch op {
		case "MustTotient":
			call, must = integer(func() (int, error) { return pitchclass.Totient(s) }), integerMust(func() int { return pitchclass.MustTotient(s) })
		case "MustIdentity":
			call = operator(func() (pitchclass.Operator, error) { return pitchclass.Identity(s) })
			must = operatorMust(func() pitchclass.Operator { return pitchclass.MustIdentity(s) })
		case "MustTranspositions":
			call = func() (any, error) { return pitchclass.Transpositions(s) }
			must = func() any { return pitchclass.MustTranspositions(s) }
		case "MustTranspositionsInversions":
			call = func() (any, error) { return pitchclass.TranspositionsInversions(s) }
			must = func() any { return pitchclass.MustTranspositionsInversions(s) }
		case "MustOperators":
			call = func() (any, error) { return pitchclass.Operators(s) }
			must = func() any { return pitchclass.MustOperators(s) }
		case "MustCountTranspositions":
			call, must = integer(func() (int, error) { return pitchclass.CountTranspositions(s) }), integerMust(func() int { return pitchclass.MustCountTranspositions(s) })
		case "MustCountTranspositionsInversions":
			call = integer(func() (int, error) { return pitchclass.CountTranspositionsInversions(s) })
			must = integerMust(func() int { return pitchclass.MustCountTranspositionsInversions(s) })
		case "MustCountOperators":
			call, must = integer(func() (int, error) { return pitchclass.CountOperators(s) }), integerMust(func() int { return pitchclass.MustCountOperators(s) })
		case "MustCountTranspositionsInversionsBig":
			call = func() (any, error) { return pitchclass.CountTranspositionsInversionsBig(s) }
			must = func() any { return pitchclass.MustCountTranspositionsInversionsBig(s) }
		case "MustCountOperatorsBig":
			call = func() (any, error) { return pitchclass.CountOperatorsBig(s) }
			must = func() any { return pitchclass.MustCountOperatorsBig(s) }
		default:
			t.Fatalf("unknown operation %s", op)
		}

		return mustCase{op: op, name: op, want: want, call: call, must: must}
	}
	indexCase := func(op string, a, s int, want error) mustCase {
		var call func() (any, error)
		var must func() any

		switch op {
		case "MustTransposition":
			call = operator(func() (pitchclass.Operator, error) { return pitchclass.Transposition(a, s) })
			must = operatorMust(func() pitchclass.Operator { return pitchclass.MustTransposition(a, s) })
		case "MustInversion":
			call = operator(func() (pitchclass.Operator, error) { return pitchclass.Inversion(a, s) })
			must = operatorMust(func() pitchclass.Operator { return pitchclass.MustInversion(a, s) })
		case "MustMultiplication":
			call = operator(func() (pitchclass.Operator, error) { return pitchclass.Multiplication(a, s) })
			must = operatorMust(func() pitchclass.Operator { return pitchclass.MustMultiplication(a, s) })
		default:
			t.Fatalf("unknown operation %s", op)
		}

		return mustCase{op: op, name: op, want: want, call: call, must: must}
	}
	newOperatorCase := func(m, n, s int, want error) mustCase {
		return mustCase{
			op: "MustNewOperator", name: "NewOperator", want: want,
			call: operator(func() (pitchclass.Operator, error) { return pitchclass.NewOperator(m, n, s) }),
			must: operatorMust(func() pitchclass.Operator { return pitchclass.MustNewOperator(m, n, s) }),
		}
	}
	aroundCase := func(a pitchclass.Class, want error) mustCase {
		return mustCase{
			op: "MustInversionAround", name: "InversionAround", want: want,
			call: operator(func() (pitchclass.Operator, error) { return pitchclass.InversionAround(a) }),
			must: operatorMust(func() pitchclass.Operator { return pitchclass.MustInversionAround(a) }),
		}
	}
	parseOperatorCase := func(in string, want error) mustCase {
		return mustCase{
			op: "MustParseOperator", name: "ParseOperator " + in, want: want,
			call: operator(func() (pitchclass.Operator, error) { return pitchclass.ParseOperator(in) }),
			must: operatorMust(func() pitchclass.Operator { return pitchclass.MustParseOperator(in) }),
		}
	}
	applyCase := func(f pitchclass.Operator, c pitchclass.Class, want error) mustCase {
		return mustCase{
			op: "MustApply", name: "Apply", want: want,
			call: class(func() (pitchclass.Class, error) { return f.Apply(c) }),
			must: classMust(func() pitchclass.Class { return f.MustApply(c) }),
		}
	}
	composeCase := func(f, g pitchclass.Operator, want error) mustCase {
		return mustCase{
			op: "MustCompose", name: "Compose", want: want,
			call: operator(func() (pitchclass.Operator, error) { return f.Compose(g) }),
			must: operatorMust(func() pitchclass.Operator { return f.MustCompose(g) }),
		}
	}

	cases := []mustCase{
		newCase(1, 12, nil), newCase(12, 12, errInvalidClass), newCase(0, 0, errInvalidClass),
		reduceCase(13, 12, nil), reduceCase(1, 0, errInvalidClass), reduceCase(minInt, -1, errInvalidClass),
		allCase(12, nil), allCase(0, errModulus),
		parseCase("1 mod 12", nil), parseCase("12 mod 12", errNotation), parseCase("0 mod 0", errNotation),
		parseNotationCase("E", pitchclass.TE, 12, nil), parseNotationCase("E", pitchclass.TE, 11, errNotation),
		parseNotationCase("1", pitchclass.Decimal, 0, errModulus), parseNotationCase("1", pitchclass.Style(9), 12, errNotation),
		parseNotationCase("x", pitchclass.Style(9), -1, errModulus),

		pairCase("MustIntervalTo", c12, d12, nil), pairCase("MustIntervalTo", c12, c19, errMismatch),
		pairCase("MustIntervalTo", zc, c12, errValue), pairCase("MustIntervalTo", zc, c19, errValue), pairCase("MustIntervalTo", c12, zc, errValue),
		pairCase("MustIntervalClass (method)", c12, d12, nil), pairCase("MustIntervalClass (method)", c19, c12, errMismatch),
		pairCase("MustIntervalClass (method)", c19, zc, errValue),
		pairCase("MustIndexSum", c12, d12, nil), pairCase("MustIndexSum", c12, c19, errMismatch), pairCase("MustIndexSum", zc, zc, errValue),
		pairCase("MustInvertAround", c12, d12, nil), pairCase("MustInvertAround", c12, c19, errMismatch), pairCase("MustInvertAround", zc, c19, errValue),
		swappingCase(c12, c12, d12, nil), swappingCase(c12, c19, d12, errMismatch), swappingCase(c19, zc, d12, errValue),
		notationCase(newClass(t, 10, 12), pitchclass.TE, nil), notationCase(newClass(t, 3, 19), pitchclass.TE, errNotation),
		notationCase(c12, pitchclass.Style(9), errNotation), notationCase(zc, pitchclass.Style(9), errValue),

		bySizeCase("MustReduceInterval", 14, 12, nil), bySizeCase("MustReduceInterval", 14, 0, errModulus),
		bySizeCase("MustComplementInterval", 14, 12, nil), bySizeCase("MustComplementInterval", minInt, -1, errModulus),
		bySizeCase("MustIntervalClass", 7, 12, nil), bySizeCase("MustIntervalClass", 7, minInt, errModulus),
		fixedCase(4, 12, nil), fixedCase(1, 12, nil), fixedCase(4, 0, errModulus),

		invertibleCase(5, 12, nil), invertibleCase(2, 12, nil), invertibleCase(5, 0, errModulus),
		bySizeCase("MustInverseMultiplier", 5, 12, nil), bySizeCase("MustInverseMultiplier", 2, 12, errNotInvertible),
		bySizeCase("MustInverseMultiplier", 2, 0, errModulus),
		multipliersCase(12, nil), multipliersCase(-1, errModulus),
		modulusCase("MustTotient", 12, nil), modulusCase("MustTotient", 0, errModulus),

		newOperatorCase(7, 5, 12, nil), newOperatorCase(2, 0, 12, errNotInvertible), newOperatorCase(1, 0, 0, errModulus),
		newOperatorCase(2, 0, 0, errModulus),
		modulusCase("MustIdentity", 12, nil), modulusCase("MustIdentity", 0, errModulus),
		indexCase("MustTransposition", 3, 12, nil), indexCase("MustTransposition", 3, -1, errModulus),
		indexCase("MustInversion", 8, 12, nil), indexCase("MustInversion", 8, 0, errModulus),
		indexCase("MustMultiplication", 7, 12, nil), indexCase("MustMultiplication", 2, 12, errNotInvertible),
		indexCase("MustMultiplication", 2, 0, errModulus),
		aroundCase(d12, nil), aroundCase(zc, errValue),
		pairCase("MustInversionSwapping", c12, d12, nil), pairCase("MustInversionSwapping", c12, c19, errMismatch),
		pairCase("MustInversionSwapping", zc, c19, errValue), pairCase("MustInversionSwapping", c19, zc, errValue),
		parseOperatorCase("T5M7 mod 12", nil), parseOperatorCase("T0M2 mod 12", errNotInvertible),
		parseOperatorCase("T12M2 mod 12", errNotation), parseOperatorCase("T0M12 mod 12", errNotation),
		parseOperatorCase("Operator(invalid)", errNotation),

		applyCase(f12, c12, nil), applyCase(f12, c19, errMismatch), applyCase(zo, c12, errValue),
		applyCase(f19, zc, errValue), applyCase(zo, zc, errValue),
		composeCase(f12, f12, nil), composeCase(f12, f19, errMismatch), composeCase(zo, f19, errValue),
		composeCase(f19, zo, errValue), composeCase(zo, zo, errValue),

		modulusCase("MustTranspositions", 12, nil), modulusCase("MustTranspositions", 0, errModulus),
		modulusCase("MustTranspositionsInversions", 12, nil), modulusCase("MustTranspositionsInversions", -1, errModulus),
		modulusCase("MustOperators", 12, nil), modulusCase("MustOperators", minInt, errModulus),
		modulusCase("MustCountTranspositions", 12, nil), modulusCase("MustCountTranspositions", 0, errModulus),
		modulusCase("MustCountTranspositions", maxInt, nil),
		modulusCase("MustCountTranspositionsInversions", 12, nil), modulusCase("MustCountTranspositionsInversions", 0, errModulus),
		modulusCase("MustCountTranspositionsInversions", maxInt, errOverflow),
		modulusCase("MustCountOperators", 12, nil), modulusCase("MustCountOperators", -1, errModulus),
		modulusCase("MustCountOperators", maxInt, errOverflow),
		modulusCase("MustCountTranspositionsInversionsBig", maxInt, nil), modulusCase("MustCountTranspositionsInversionsBig", 0, errModulus),
		modulusCase("MustCountOperatorsBig", maxInt, nil), modulusCase("MustCountOperatorsBig", 0, errModulus),
	}

	return cases
}

// mustOperations are the 37 Must variants of the architecture ("Must
// variants"); the method IntervalClass is listed apart from the function.
func mustOperations() []string {
	return []string{
		"MustNew", "MustReduce", "MustAll", "MustParse", "MustParseNotation",
		"MustIntervalTo", "MustIntervalClass (method)", "MustIndexSum", "MustInvertAround", "MustInvertSwapping", "MustNotation",
		"MustReduceInterval", "MustComplementInterval", "MustIntervalClass",
		"MustInversionFixedClasses",
		"MustInvertible", "MustInverseMultiplier", "MustInvertibleMultipliers", "MustTotient",
		"MustNewOperator", "MustIdentity", "MustTransposition", "MustInversion", "MustInversionAround",
		"MustInversionSwapping", "MustMultiplication", "MustParseOperator",
		"MustApply", "MustCompose",
		"MustTranspositions", "MustTranspositionsInversions", "MustOperators",
		"MustCountTranspositions", "MustCountTranspositionsInversions", "MustCountOperators",
		"MustCountTranspositionsInversionsBig", "MustCountOperatorsBig",
	}
}

// K-30, dev.md "Must": the cases cover each of the 37 Must variants with a
// valid call and with at least one error.
func TestMustCasesCoverAllOperations(t *testing.T) {
	t.Parallel()

	ops := mustOperations()
	require.Len(t, ops, 37)

	valid, failing := map[string]bool{}, map[string]bool{}
	for _, c := range mustCases(t) {
		require.Contains(t, ops, c.op, "unknown Must variant in the cases")
		if c.want == nil {
			valid[c.op] = true
		} else {
			failing[c.op] = true
		}
	}

	for _, op := range ops {
		assert.True(t, valid[op], "%s has a valid call", op)
		assert.True(t, failing[op], "%s has a failing call", op)
	}
}

// K-30, E-1...E-7: each operation with an error reports exactly one
// sentinel error and the documented zero result; its Must variant panics
// with an error that matches the same sentinel, and on valid inputs returns
// the same result without a panic.
func TestMustVariants(t *testing.T) {
	t.Parallel()

	for _, c := range mustCases(t) {
		t.Run(c.name+" "+errName(c.want), func(t *testing.T) {
			t.Parallel()

			var (
				got any
				err error
			)
			require.NotPanics(t, func() { got, err = c.call() }, "the variant with an error never panics")

			if c.want != nil {
				assertOnlyError(t, err, c.want, c.name)
				assertZeroResult(t, got, c.name)
				assertPanicsWithErrorIs(t, c.want, func() { c.must() }, c.op)

				return
			}

			require.NoError(t, err, c.name)

			var fromMust any
			require.NotPanics(t, func() { fromMust = c.must() }, c.op)
			assertSameResult(t, got, fromMust, c.op)
		})
	}
}

// errName names a sentinel error for subtest names.
func errName(err error) string {
	if err == nil {
		return "ok"
	}

	return err.Error()
}

// assertZeroResult checks the value returned with an error (architecture,
// "Errors"): the zero Class or Operator, 0, false, "", nil slice or
// *big.Int, a non-nil empty iterator.
func assertZeroResult(t *testing.T, v any, desc string) {
	t.Helper()

	switch x := v.(type) {
	case pitchclass.Class:
		assert.Equal(t, pitchclass.Class{}, x, desc)
	case pitchclass.Operator:
		assert.Equal(t, pitchclass.Operator{}, x, desc)
	case int:
		assert.Zero(t, x, desc)
	case bool:
		assert.False(t, x, desc)
	case string:
		assert.Empty(t, x, desc)
	case []pitchclass.Class:
		assert.Nil(t, x, desc)
	case *big.Int:
		assert.Nil(t, x, desc)
	case iter.Seq[pitchclass.Class]:
		assertEmptySeq(t, x, desc)
	case iter.Seq[pitchclass.Operator]:
		assertEmptySeq(t, x, desc)
	case iter.Seq[int]:
		assertEmptySeq(t, x, desc)
	default:
		assert.Fail(t, "unexpected result type", "%s: %T", desc, v)
	}
}

// assertSameResult checks that the Must variant returned the same result as
// the variant with an error; iterators are compared by their elements.
func assertSameResult(t *testing.T, want, got any, desc string) {
	t.Helper()

	const limit = 100

	switch w := want.(type) {
	case iter.Seq[pitchclass.Class]:
		g, ok := got.(iter.Seq[pitchclass.Class])
		require.True(t, ok, desc)
		assert.Equal(t, take(w, limit), take(g, limit), desc)
	case iter.Seq[pitchclass.Operator]:
		g, ok := got.(iter.Seq[pitchclass.Operator])
		require.True(t, ok, desc)
		assert.Equal(t, take(w, limit), take(g, limit), desc)
	case iter.Seq[int]:
		g, ok := got.(iter.Seq[int])
		require.True(t, ok, desc)
		assert.Equal(t, take(w, limit), take(g, limit), desc)
	case *big.Int:
		g, ok := got.(*big.Int)
		require.True(t, ok, desc)
		require.NotNil(t, w, desc)
		require.NotNil(t, g, desc)
		assert.Zero(t, w.Cmp(g), "%s: %v and %v", desc, w, g)
	default:
		assert.Equal(t, want, got, desc)
	}
}
