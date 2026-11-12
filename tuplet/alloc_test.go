package tuplet_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
)

// Package-level sinks keep the measured results alive.
var (
	sinkTuplet   tuplet.Tuplet
	sinkFraction fraction.Fraction
	sinkInt      int
	sinkBool     bool
	sinkKind     tuplet.Kind
	sinkDivision tuplet.Division
	sinkString   string
	sinkErr      error
)

// allocPair is a pair of tuplets measured together: small numbers and
// numbers near math.MaxInt (NF-3, architecture 6).
type allocPair struct {
	name string
	a, b tuplet.Tuplet
	// defaultActual has a default normal count for both divisions.
	defaultActual int
}

func allocPairs() []allocPair {
	return []allocPair{
		{name: "small", a: tuplet.MustNew(3, 2), b: tuplet.MustNew(6, 4), defaultActual: 5},
		{
			name: "large",
			a:    tuplet.MustNew(math.MaxInt, math.MaxInt-1),
			b:    tuplet.MustNew(math.MaxInt-1, math.MaxInt),

			defaultActual: math.MaxInt - 1,
		},
	}
}

// TestZeroAllocs covers NF-3 and the rows "0 - measured" of architecture 6:
// the operations do not allocate on the success path, for small numbers and
// for numbers near math.MaxInt. The tests do not run in parallel, as in the
// TestZeroAllocs test of common/fraction.
//
//nolint:paralleltest // testing.AllocsPerRun must not overlap other tests.
func TestZeroAllocs(t *testing.T) {
	for _, p := range allocPairs() {
		a, b := p.a, p.b
		aActual, aNormal := a.MustPair()
		bActual, bNormal := b.MustPair()
		coefficient := a.MustCoefficient()
		reciprocal := b.MustReciprocalCoefficient()

		ops := []struct {
			name string
			f    func()
		}{
			{name: "New", f: func() { sinkTuplet, sinkErr = tuplet.New(aActual, aNormal) }},
			{name: "New second", f: func() { sinkTuplet, sinkErr = tuplet.New(bActual, bNormal) }},
			{name: "Default binary", f: func() { sinkTuplet, sinkErr = tuplet.Default(p.defaultActual, tuplet.Binary) }},
			{name: "Default ternary", f: func() { sinkTuplet, sinkErr = tuplet.Default(p.defaultActual, tuplet.Ternary) }},
			{name: "FromCoefficient", f: func() { sinkTuplet, sinkErr = tuplet.FromCoefficient(coefficient) }},
			{name: "FromReciprocalCoefficient", f: func() { sinkTuplet, sinkErr = tuplet.FromReciprocalCoefficient(reciprocal) }},
			{name: "named", f: func() {
				sinkTuplet = tuplet.Duplet()
				sinkTuplet = tuplet.Triplet()
				sinkTuplet = tuplet.Quadruplet()
				sinkTuplet = tuplet.Quintuplet()
				sinkTuplet = tuplet.Sextuplet()
				sinkTuplet = tuplet.Septuplet()
				sinkTuplet = tuplet.Octuplet()
				sinkTuplet = tuplet.Nonuplet()
				sinkTuplet = tuplet.Decuplet()
			}},
			{name: "IsSet", f: func() { sinkBool = a.IsSet() }},
			{name: "Actual", f: func() { sinkInt, sinkErr = a.Actual() }},
			{name: "Normal", f: func() { sinkInt, sinkErr = a.Normal() }},
			{name: "Pair", f: func() { sinkInt, sinkInt, sinkErr = a.Pair() }},
			{name: "Equal", f: func() { sinkBool = a.Equal(b) }},
			{name: "Compare", f: func() { sinkInt = a.Compare(b) }},
			{name: "EqualCoefficient", f: func() { sinkBool, sinkErr = a.EqualCoefficient(b) }},
			{name: "CompareCoefficient", f: func() { sinkInt, sinkErr = a.CompareCoefficient(b) }},
			{name: "Kind", f: func() { sinkKind, sinkErr = a.Kind() }},
			{name: "IsCompressing", f: func() { sinkBool, sinkErr = a.IsCompressing() }},
			{name: "IsExpanding", f: func() { sinkBool, sinkErr = a.IsExpanding() }},
			{name: "IsNeutral", f: func() { sinkBool, sinkErr = a.IsNeutral() }},
			{name: "NormalDivision", f: func() { sinkDivision, sinkBool, sinkErr = a.NormalDivision() }},
			{name: "DivisionOf", f: func() { sinkDivision, sinkBool = tuplet.DivisionOf(aNormal) }},
			{name: "HasDefaultNormal", f: func() { sinkBool, sinkErr = a.HasDefaultNormal() }},
			{name: "HasDefaultNormalFor", f: func() { sinkBool, sinkErr = b.HasDefaultNormalFor(tuplet.Ternary) }},
			{name: "Coefficient", f: func() { sinkFraction, sinkErr = a.Coefficient() }},
			{name: "ReciprocalCoefficient", f: func() { sinkFraction, sinkErr = a.ReciprocalCoefficient() }},
			{name: "ReducedCoefficient", f: func() { sinkFraction, sinkErr = b.ReducedCoefficient() }},
			{name: "ReducedCoefficient second", f: func() { sinkFraction, sinkErr = a.ReducedCoefficient() }},
		}

		for _, op := range ops {
			t.Run(p.name+"/"+op.name, func(t *testing.T) {
				sinkErr = nil
				op.f()
				assert.NoError(t, sinkErr, "the measured call succeeds")

				sinkErr = nil
				allocs := testing.AllocsPerRun(100, op.f)
				assert.Equal(t, 0.0, allocs, "allocations per call")
			})
		}
	}
}
