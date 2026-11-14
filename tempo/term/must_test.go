package term_test

import (
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/go-muse/musex/common/fraction"
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

func res(v ...any) []any { return v }

func termMustCases() []mustCase {
	en := term.EnglishWikipedia()
	largo := en.MustParse("Largo")
	other := term.RussianWikipedia().MustParse("largo")

	list := func(good bool) term.List {
		if good {
			return en
		}

		return term.List{}
	}
	word := func(good bool) term.Term {
		if good {
			return largo
		}

		return term.Term{}
	}
	num := func(good bool) fraction.Fraction {
		if good {
			return fi(40)
		}

		return fi(0)
	}
	rows := func(good bool) []term.Row {
		if good {
			return customRows()
		}

		return nil
	}

	return []mustCase{
		{"Between",
			func(g bool) ([]any, error) { v, err := term.Between(num(g), fi(60)); return res(v), err },
			func(g bool) []any { return res(term.MustBetween(num(g), fi(60))) }},
		{"AtLeast",
			func(g bool) ([]any, error) { v, err := term.AtLeast(num(g)); return res(v), err },
			func(g bool) []any { return res(term.MustAtLeast(num(g))) }},
		{"AtMost",
			func(g bool) ([]any, error) { v, err := term.AtMost(num(g)); return res(v), err },
			func(g bool) []any { return res(term.MustAtMost(num(g))) }},
		{"Guide.Contains",
			func(g bool) ([]any, error) { v, err := between(40, 60).Contains(num(g)); return res(v), err },
			func(g bool) []any { return res(between(40, 60).MustContains(num(g))) }},
		{"NewList",
			func(g bool) ([]any, error) { v, err := term.NewList(rows(g)...); return res(v), err },
			func(g bool) []any { return res(term.MustNewList(rows(g)...)) }},
		{"List.Len",
			func(g bool) ([]any, error) { v, err := list(g).Len(); return res(v), err },
			func(g bool) []any { return res(list(g).MustLen()) }},
		{"List.Place",
			func(g bool) ([]any, error) { v, err := list(g).Place(2); return res(v), err },
			func(g bool) []any { return res(list(g).MustPlace(2)) }},
		{"List.Terms",
			func(g bool) ([]any, error) { v, err := list(g).Terms(); return res(v), err },
			func(g bool) []any { return res(list(g).MustTerms()) }},
		{"List.Parse",
			func(g bool) ([]any, error) { v, err := list(g).Parse("vivo"); return res(v), err },
			func(g bool) []any { return res(list(g).MustParse("vivo")) }},
		{"List.OfGroup",
			func(g bool) ([]any, error) { v, err := list(g).OfGroup(term.Fast); return res(v), err },
			func(g bool) []any { return res(list(g).MustOfGroup(term.Fast)) }},
		{"List.At",
			func(g bool) ([]any, error) { v, err := en.At(num(g)); return res(v), err },
			func(g bool) []any { return res(en.MustAt(num(g))) }},
		{"Term.Name",
			func(g bool) ([]any, error) { v, err := word(g).Name(); return res(v), err },
			func(g bool) []any { return res(word(g).MustName()) }},
		{"Term.List",
			func(g bool) ([]any, error) { v, err := word(g).List(); return res(v), err },
			func(g bool) []any { return res(word(g).MustList()) }},
		{"Term.Place",
			func(g bool) ([]any, error) { v, err := word(g).Place(); return res(v), err },
			func(g bool) []any { return res(word(g).MustPlace()) }},
		{"Term.Group",
			func(g bool) ([]any, error) { v, err := word(g).Group(); return res(v), err },
			func(g bool) []any { return res(word(g).MustGroup()) }},
		{"Term.Guide",
			func(g bool) ([]any, error) { v, ok, err := word(g).Guide(); return res(v, ok), err },
			func(g bool) []any { v, ok := word(g).MustGuide(); return res(v, ok) }},
		{"Term.English",
			func(g bool) ([]any, error) { v, ok, err := word(g).English(); return res(v, ok), err },
			func(g bool) []any { v, ok := word(g).MustEnglish(); return res(v, ok) }},
		{"Term.Russian",
			func(g bool) ([]any, error) { v, ok, err := word(g).Russian(); return res(v, ok), err },
			func(g bool) []any { v, ok := word(g).MustRussian(); return res(v, ok) }},
		{"Term.ComparePlace",
			func(g bool) ([]any, error) {
				o := en.MustParse("Allegro")
				if !g {
					o = other
				}

				v, err := largo.ComparePlace(o)

				return res(v), err
			},
			func(g bool) []any {
				o := en.MustParse("Allegro")
				if !g {
					o = other
				}

				return res(largo.MustComparePlace(o))
			}},
	}
}

// TestMust covers architecture 4.3, "Must": every Must variant of 4.1 returns
// the results of its base operation on success and panics with the same error
// value on failure.
func TestMust(t *testing.T) {
	t.Parallel()

	cases := termMustCases()
	require.Len(t, cases, 19, "19 Must variants in tempo/term (architecture 10, Р-TP-1)")

	for _, tc := range cases {
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

			for _, s := range termSentinels {
				assert.Equal(t, errors.Is(baseErr, s.err), errors.Is(mustErr, s.err), s.name)
			}
		})
	}
}

// TestMethodSets covers architecture 4.1 ("no exported identifiers beyond
// those listed") and 4.3 ("exactly one Must variant with the same parameters
// and results without error") for the methods of the types of tempo/term.
func TestMethodSets(t *testing.T) {
	t.Parallel()

	want := map[reflect.Type][]string{
		reflect.TypeFor[term.Group](): {"Equal", "IsValid", "String"},
		reflect.TypeFor[term.Guide](): {"Contains", "Equal", "IsValid", "Lower", "MustContains", "String", "Upper"},
		reflect.TypeFor[term.List](): {
			"At", "Equal", "IsSet", "Len", "MustAt", "MustLen", "MustOfGroup", "MustParse", "MustPlace",
			"MustTerms", "OfGroup", "Parse", "Place", "String", "Terms",
		},
		reflect.TypeFor[term.Term](): {
			"AppendText", "ComparePlace", "English", "Equal", "Group", "Guide", "IsSet", "List", "MarshalText",
			"MustComparePlace", "MustEnglish", "MustGroup", "MustGuide", "MustList", "MustName", "MustPlace",
			"MustRussian", "Name", "Place", "Russian", "String",
		},
		reflect.TypeFor[term.Entry](): nil,
		reflect.TypeFor[term.Row]():   nil,
	}

	for typ, methods := range want {
		var got []string
		for i := range typ.NumMethod() {
			got = append(got, typ.Method(i).Name)
		}

		assert.Equal(t, methods, got, "methods of %s", typ)

		ptr := reflect.PointerTo(typ)
		assert.Equal(t, typ.NumMethod(), ptr.NumMethod(), "%s has no pointer-receiver methods", typ)

		checkMustPairs(t, typ)
	}
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
