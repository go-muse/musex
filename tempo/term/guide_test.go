package term_test

import (
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/tempo/term"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGroup covers architecture 4.4, "Group".
func TestGroup(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []term.Group{term.Slow, term.Moderate, term.Fast}, term.Groups())

	g := term.Groups()
	g[0] = term.Fast
	assert.Equal(t, term.Slow, term.Groups()[0], "Groups returns a new slice")

	assert.Equal(t, term.NoGroup, term.Group(0), "the zero group is NoGroup")

	tests := []struct {
		g     term.Group
		valid bool
		text  string
	}{
		{term.NoGroup, true, "none"},
		{term.Slow, true, "slow"},
		{term.Moderate, true, "moderate"},
		{term.Fast, true, "fast"},
		{term.Group(4), false, "Group(invalid)"},
		{term.Group(255), false, "Group(invalid)"},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.valid, tt.g.IsValid())
			assert.Equal(t, tt.text, tt.g.String())

			for _, other := range []term.Group{term.NoGroup, term.Slow, term.Moderate, term.Fast, term.Group(4)} {
				assert.Equal(t, tt.g == other, tt.g.Equal(other), "%d.Equal(%d)", tt.g, other)
			}
		})
	}
}

// TestGuideCreate covers architecture 4.4, "Guide": bounds, open bounds,
// equal bounds, reduced storage, and the Е-2 conditions (С-42 through the
// constructors).
func TestGuideCreate(t *testing.T) {
	t.Parallel()

	type bounds struct {
		lo, hi     fraction.Fraction
		loOK, hiOK bool
	}

	tests := []struct {
		name string
		make func() (term.Guide, error)
		want bounds
		text string
	}{
		{"Between 40-60", func() (term.Guide, error) { return term.Between(fi(40), fi(60)) },
			bounds{fi(40), fi(60), true, true}, "40-60"},
		{"Between equal bounds", func() (term.Guide, error) { return term.Between(fi(100), fi(100)) },
			bounds{fi(100), fi(100), true, true}, "100-100"},
		{"Between fractions", func() (term.Guide, error) { return term.Between(fr(1, 2), fr(3, 2)) },
			bounds{fr(1, 2), fr(3, 2), true, true}, "1/2-3/2"},
		{"Between unreduced is stored reduced",
			func() (term.Guide, error) { return term.Between(fraction.MustNew(120, 2), fraction.MustNew(240, 2)) },
			bounds{fi(60), fi(120), true, true}, "60-120"},
		{"AtLeast", func() (term.Guide, error) { return term.AtLeast(fi(200)) },
			bounds{fi(200), fraction.Fraction{}, true, false}, ">=200"},
		{"AtMost", func() (term.Guide, error) { return term.AtMost(fi(24)) },
			bounds{fraction.Fraction{}, fi(24), false, true}, "<=24"},
		{"AtLeast fraction", func() (term.Guide, error) { return term.AtLeast(fr(185, 2)) },
			bounds{fr(185, 2), fraction.Fraction{}, true, false}, ">=185/2"},
		{"AtMost unreduced", func() (term.Guide, error) { return term.AtMost(fraction.MustNew(48, 2)) },
			bounds{fraction.Fraction{}, fi(24), false, true}, "<=24"},
		{"Between max", func() (term.Guide, error) { return term.Between(invMaxF(), maxF()) },
			bounds{invMaxF(), maxF(), true, true}, "1/" + fraction.MaxComponentString + "-" + fraction.MaxComponentString},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			g, err := tt.make()
			require.NoError(t, err)
			assert.True(t, g.IsValid())

			lo, ok := g.Lower()
			assert.Equal(t, tt.want.loOK, ok)
			assert.Equal(t, tt.want.lo, lo, "Lower %s", lo)

			hi, ok := g.Upper()
			assert.Equal(t, tt.want.hiOK, ok)
			assert.Equal(t, tt.want.hi, hi, "Upper %s", hi)

			assert.Equal(t, tt.text, g.String())
		})
	}
}

// TestGuideCreateErrors covers the Е-2 conditions of term.Between, AtLeast,
// and AtMost (architecture 4.3, 4.4; С-42): unset fraction, bound <= 0, lower
// greater than upper. The result is Guide{}.
func TestGuideCreateErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		make func() (term.Guide, error)
	}{
		{"Between unset lower", func() (term.Guide, error) { return term.Between(fraction.Fraction{}, fi(60)) }},
		{"Between unset upper", func() (term.Guide, error) { return term.Between(fi(40), fraction.Fraction{}) }},
		{"Between both unset", func() (term.Guide, error) { return term.Between(fraction.Fraction{}, fraction.Fraction{}) }},
		{"С-42 Between 60-40", func() (term.Guide, error) { return term.Between(fi(60), fi(40)) }},
		{"С-42 Between zero lower", func() (term.Guide, error) { return term.Between(fi(0), fi(40)) }},
		{"Between zero both", func() (term.Guide, error) { return term.Between(fi(0), fi(0)) }},
		{"Between negative", func() (term.Guide, error) { return term.Between(fi(-5), fi(40)) }},
		{"Between negative upper", func() (term.Guide, error) { return term.Between(fi(-5), fi(-1)) }},
		{"AtLeast unset", func() (term.Guide, error) { return term.AtLeast(fraction.Fraction{}) }},
		{"AtLeast zero", func() (term.Guide, error) { return term.AtLeast(fi(0)) }},
		{"AtLeast negative", func() (term.Guide, error) { return term.AtLeast(fr(-1, 2)) }},
		{"AtMost unset", func() (term.Guide, error) { return term.AtMost(fraction.Fraction{}) }},
		{"AtMost zero", func() (term.Guide, error) { return term.AtMost(fi(0)) }},
		{"AtMost negative", func() (term.Guide, error) { return term.AtMost(fi(-24)) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			g, err := tt.make()
			assertErr(t, err, term.ErrInvalidArgument)
			assert.Equal(t, term.Guide{}, g)
		})
	}
}

// TestGuideZero covers architecture 3.1, 4.4: the zero Guide has no bounds
// and is not valid.
func TestGuideZero(t *testing.T) {
	t.Parallel()

	var g term.Guide

	assert.False(t, g.IsValid())

	lo, ok := g.Lower()
	assert.False(t, ok)
	assert.Equal(t, fraction.Fraction{}, lo)

	hi, ok := g.Upper()
	assert.False(t, ok)
	assert.Equal(t, fraction.Fraction{}, hi)

	assert.Equal(t, "Guide(none)", g.String())
	assert.True(t, g.Equal(term.Guide{}))
	assert.False(t, g.Equal(between(40, 60)))

	got, err := g.Contains(fi(50))
	assertErr(t, err, term.ErrInvalidArgument)
	assert.False(t, got)
}

// TestGuideContains covers architecture 4.4: bounds are included (П-11,
// П-18), an open bound does not limit, Е-2 and Е-3, and their order:
// guide without bounds -> unset number -> number <= 0.
func TestGuideContains(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		g     term.Guide
		beats fraction.Fraction
		want  bool
	}{
		{"inside", between(40, 60), fi(50), true},
		{"lower bound included", between(40, 60), fi(40), true},
		{"upper bound included", between(40, 60), fi(60), true},
		{"below", between(40, 60), fi(39), false},
		{"above", between(40, 60), fi(61), false},
		{"just above by a fraction", between(40, 60), fr(121, 2), false},
		{"fraction inside", between(40, 60), fr(81, 2), true},
		{"equal bounds", between(100, 100), fi(100), true},
		{"equal bounds miss", between(100, 100), fr(201, 2), false},
		{"open upper includes its bound", atLeast(200), fi(200), true},
		{"open upper huge", atLeast(200), maxF(), true},
		{"open upper below", atLeast(200), fi(199), false},
		{"open lower includes its bound", atMost(24), fi(24), true},
		{"open lower tiny", atMost(24), invMaxF(), true},
		{"open lower above", atMost(24), fr(49, 2), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.g.Contains(tt.beats)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want, tt.g.MustContains(tt.beats))
		})
	}

	errTests := []struct {
		name  string
		g     term.Guide
		beats fraction.Fraction
		want  error
	}{
		{"unset number", between(40, 60), fraction.Fraction{}, term.ErrInvalidArgument},
		{"zero", between(40, 60), fi(0), term.ErrNoTempo},
		{"negative", atMost(24), fi(-5), term.ErrNoTempo},
		{"order: no bounds before unset number", term.Guide{}, fraction.Fraction{}, term.ErrInvalidArgument},
		{"order: no bounds before number <= 0", term.Guide{}, fi(-5), term.ErrInvalidArgument},
		{"order: unset number before number <= 0 (one case)", atLeast(1), fraction.Fraction{}, term.ErrInvalidArgument},
	}

	for _, tt := range errTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.g.Contains(tt.beats)
			assertErr(t, err, tt.want)
			assert.False(t, got)
		})
	}
}

// TestGuideEqual covers architecture 3.1, 4.4: Equal compares bounds and
// openness and matches ==.
func TestGuideEqual(t *testing.T) {
	t.Parallel()

	values := []term.Guide{
		{}, between(40, 60), between(40, 61), between(41, 60), between(100, 100),
		atLeast(40), atMost(40), atLeast(60), atMost(60),
		term.MustBetween(fraction.MustNew(80, 2), fraction.MustNew(120, 2)),
	}

	for _, a := range values {
		for _, b := range values {
			assert.Equal(t, a == b, a.Equal(b), "%s vs %s", a, b)
		}
	}

	assert.True(t, between(40, 60).Equal(values[9]), "bounds are compared by value")
	assert.Equal(t, between(40, 60), values[9])
	assert.False(t, atLeast(40).Equal(between(40, 40)))

	m := map[term.Guide]int{between(40, 60): 1}
	m[values[9]]++
	assert.Len(t, m, 1, "a Guide is usable as a map key")
	assert.Equal(t, 2, m[between(40, 60)])
}
