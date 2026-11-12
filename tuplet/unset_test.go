package tuplet_test

import (
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
)

// unsetCase is an operation on an unset tuplet: call returns the results
// that architecture 4.2 requires to be zero values on error, and the error.
type unsetCase struct {
	name string
	call func() ([]any, error)
}

// unsetCases lists every operation of summary 4.1 that receives a tuplet,
// except IsSet, Equal, Compare and String, with the unset tuplet as the
// receiver, an argument, or a chain element (K-12, E-2, P-3).
func unsetCases() []unsetCase {
	var u tuplet.Tuplet

	set := tuplet.Triplet()
	d := frac(1, 8)

	return []unsetCase{
		{name: "Actual", call: func() ([]any, error) { v, err := u.Actual(); return []any{v}, err }},
		{name: "Normal", call: func() ([]any, error) { v, err := u.Normal(); return []any{v}, err }},
		{name: "Pair", call: func() ([]any, error) { a, n, err := u.Pair(); return []any{a, n}, err }},
		{name: "Kind", call: func() ([]any, error) { v, err := u.Kind(); return []any{v}, err }},
		{name: "IsCompressing", call: func() ([]any, error) { v, err := u.IsCompressing(); return []any{v}, err }},
		{name: "IsExpanding", call: func() ([]any, error) { v, err := u.IsExpanding(); return []any{v}, err }},
		{name: "IsNeutral", call: func() ([]any, error) { v, err := u.IsNeutral(); return []any{v}, err }},
		{name: "NormalDivision", call: func() ([]any, error) {
			v, ok, err := u.NormalDivision()
			return []any{v, ok}, err
		}},
		{name: "HasDefaultNormal", call: func() ([]any, error) { v, err := u.HasDefaultNormal(); return []any{v}, err }},
		{name: "HasDefaultNormalFor", call: func() ([]any, error) {
			v, err := u.HasDefaultNormalFor(tuplet.Binary)
			return []any{v}, err
		}},
		{name: "Coefficient", call: func() ([]any, error) { v, err := u.Coefficient(); return []any{v}, err }},
		{name: "ReciprocalCoefficient", call: func() ([]any, error) {
			v, err := u.ReciprocalCoefficient()
			return []any{v}, err
		}},
		{name: "ReducedCoefficient", call: func() ([]any, error) { v, err := u.ReducedCoefficient(); return []any{v}, err }},
		{name: "EqualCoefficient receiver", call: func() ([]any, error) {
			v, err := u.EqualCoefficient(set)
			return []any{v}, err
		}},
		{name: "EqualCoefficient argument", call: func() ([]any, error) {
			v, err := set.EqualCoefficient(u)
			return []any{v}, err
		}},
		{name: "EqualCoefficient both", call: func() ([]any, error) {
			v, err := u.EqualCoefficient(u)
			return []any{v}, err
		}},
		{name: "CompareCoefficient receiver", call: func() ([]any, error) {
			v, err := u.CompareCoefficient(set)
			return []any{v}, err
		}},
		{name: "CompareCoefficient argument", call: func() ([]any, error) {
			v, err := set.CompareCoefficient(u)
			return []any{v}, err
		}},
		{name: "CompareCoefficient both", call: func() ([]any, error) {
			v, err := u.CompareCoefficient(u)
			return []any{v}, err
		}},
		{name: "Reduce", call: func() ([]any, error) { v, err := u.Reduce(); return []any{v}, err }},
		{name: "Scale", call: func() ([]any, error) { v, err := u.Scale(2); return []any{v}, err }},
		{name: "Apply", call: func() ([]any, error) { v, err := u.Apply(d); return []any{v}, err }},
		{name: "Unapply", call: func() ([]any, error) { v, err := u.Unapply(d); return []any{v}, err }},
		{name: "Span", call: func() ([]any, error) { v, err := u.Span(d); return []any{v}, err }},
		{name: "Nest receiver", call: func() ([]any, error) { v, err := u.Nest(set); return []any{v}, err }},
		{name: "Nest argument", call: func() ([]any, error) { v, err := set.Nest(u); return []any{v}, err }},
		{name: "Flatten only", call: func() ([]any, error) { v, err := tuplet.Flatten(u); return []any{v}, err }},
		{name: "Flatten first", call: func() ([]any, error) { v, err := tuplet.Flatten(u, set, set); return []any{v}, err }},
		{name: "Flatten middle", call: func() ([]any, error) { v, err := tuplet.Flatten(set, u, set); return []any{v}, err }},
		{name: "Flatten last", call: func() ([]any, error) { v, err := tuplet.Flatten(set, set, u); return []any{v}, err }},
		{name: "ChainCoefficient first", call: func() ([]any, error) {
			v, err := tuplet.ChainCoefficient(u, set)
			return []any{v}, err
		}},
		{name: "ChainCoefficient last", call: func() ([]any, error) {
			v, err := tuplet.ChainCoefficient(set, set, u)
			return []any{v}, err
		}},
		{name: "ApplyChain first", call: func() ([]any, error) {
			v, err := tuplet.ApplyChain(d, u, set)
			return []any{v}, err
		}},
		{name: "ApplyChain last", call: func() ([]any, error) {
			v, err := tuplet.ApplyChain(d, set, u)
			return []any{v}, err
		}},
		{name: "Name", call: func() ([]any, error) { v, ok, err := u.Name(); return []any{v, ok}, err }},
		{name: "RussianName", call: func() ([]any, error) { v, ok, err := u.RussianName(); return []any{v, ok}, err }},
		{name: "ShortString", call: func() ([]any, error) { v, err := u.ShortString(); return []any{v}, err }},
		{name: "MarshalText", call: func() ([]any, error) { v, err := u.MarshalText(); return []any{v}, err }},
	}
}

// zeroResult is the zero value of a result as listed in architecture 4.2.
func zeroResult(v any) any {
	switch v.(type) {
	case int:
		return 0
	case bool:
		return false
	case string:
		return ""
	case tuplet.Kind:
		return tuplet.Kind(0)
	case tuplet.Division:
		return tuplet.Division(0)
	case tuplet.Tuplet:
		return tuplet.Tuplet{}
	case fraction.Fraction:
		return fraction.Fraction{}
	default:
		return v
	}
}

// TestUnset covers K-12, E-2 and architecture 4.2: every operation except
// IsSet, Equal, Compare and String returns ErrUnset and zero values when the
// receiver, an argument or a chain element is unset.
func TestUnset(t *testing.T) {
	t.Parallel()

	for _, tc := range unsetCases() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			results, err := tc.call()
			assertErr(t, err, tuplet.ErrUnset)

			for i, v := range results {
				if b, isBytes := v.([]byte); isBytes {
					assert.Nil(t, b, "result %d", i)

					continue
				}

				assert.Equal(t, zeroResult(v), v, "result %d", i)
			}
		})
	}
}

// TestUnset_Allowed covers P-3: IsSet, Equal, Compare and String work on the
// unset tuplet without errors.
func TestUnset_Allowed(t *testing.T) {
	t.Parallel()

	var u tuplet.Tuplet

	assert.False(t, u.IsSet())
	assert.True(t, u.Equal(tuplet.Tuplet{}))
	assert.False(t, u.Equal(tuplet.Triplet()))
	assert.Equal(t, 0, u.Compare(tuplet.Tuplet{}))
	assert.Negative(t, u.Compare(tuplet.MustNew(1, 1)))
	assert.Positive(t, tuplet.MustNew(1, 1).Compare(u))
	assert.Equal(t, "Tuplet(unset)", u.String())
}

// TestUnset_AppendText covers E-2 for AppendText: b is returned unchanged.
func TestUnset_AppendText(t *testing.T) {
	t.Parallel()

	b := []byte("prefix")

	got, err := tuplet.Tuplet{}.AppendText(b)
	assertErr(t, err, tuplet.ErrUnset)
	assert.Equal(t, "prefix", string(got))
}

// TestUnset_Must covers K-12 with K-14: Must variants panic with ErrUnset on
// the unset tuplet.
func TestUnset_Must(t *testing.T) {
	t.Parallel()

	var u tuplet.Tuplet

	set := tuplet.Triplet()
	d := frac(1, 8)

	calls := map[string]func(){
		"MustActual":                func() { u.MustActual() },
		"MustNormal":                func() { u.MustNormal() },
		"MustPair":                  func() { u.MustPair() },
		"MustKind":                  func() { u.MustKind() },
		"MustIsCompressing":         func() { u.MustIsCompressing() },
		"MustIsExpanding":           func() { u.MustIsExpanding() },
		"MustIsNeutral":             func() { u.MustIsNeutral() },
		"MustNormalDivision":        func() { u.MustNormalDivision() },
		"MustHasDefaultNormal":      func() { u.MustHasDefaultNormal() },
		"MustHasDefaultNormalFor":   func() { u.MustHasDefaultNormalFor(tuplet.Ternary) },
		"MustCoefficient":           func() { u.MustCoefficient() },
		"MustReciprocalCoefficient": func() { u.MustReciprocalCoefficient() },
		"MustReducedCoefficient":    func() { u.MustReducedCoefficient() },
		"MustEqualCoefficient":      func() { set.MustEqualCoefficient(u) },
		"MustCompareCoefficient":    func() { u.MustCompareCoefficient(set) },
		"MustReduce":                func() { u.MustReduce() },
		"MustScale":                 func() { u.MustScale(3) },
		"MustApply":                 func() { u.MustApply(d) },
		"MustUnapply":               func() { u.MustUnapply(d) },
		"MustSpan":                  func() { u.MustSpan(d) },
		"MustNest":                  func() { set.MustNest(u) },
		"MustFlatten":               func() { tuplet.MustFlatten(set, u) },
		"MustChainCoefficient":      func() { tuplet.MustChainCoefficient(u) },
		"MustApplyChain":            func() { tuplet.MustApplyChain(d, set, u) },
		"MustName":                  func() { u.MustName() },
		"MustRussianName":           func() { u.MustRussianName() },
		"MustShortString":           func() { u.MustShortString() },
	}

	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_ = assertPanicErr(t, call, tuplet.ErrUnset)
		})
	}
}
