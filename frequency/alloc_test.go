package frequency_test

import (
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/frequency"

	"github.com/stretchr/testify/assert"
)

var (
	criterionSink  frequency.Criterion
	boolSink       bool
	intSink        int
	floatSink      float64
	centsSink      cents.Cents
	kindSink       frequency.Kind
	stringSink     string
	componentsSink []frequency.Criterion
)

func TestZeroAllocations(t *testing.T) {
	check := func(name string, call func()) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			assert.Zero(t, testing.AllocsPerRun(1000, call))
		})
	}

	check("Equal", func() { boolSink = frequency.MustEqual(440, 440) })
	check("Compare", func() { intSink = frequency.MustCompare(440, 441) })
	check("Equal with error result", func() {
		var err error
		boolSink, err = frequency.Equal(440, 440)
		if err != nil {
			panic(err)
		}
	})
	check("Compare with error result", func() {
		var err error
		intSink, err = frequency.Compare(440, 441)
		if err != nil {
			panic(err)
		}
	})
	check("Strict", func() { criterionSink = frequency.Strict() })
	check("Strict.Near", func() { boolSink = frequency.Strict().MustNear(440, 440) })
	check("Strict.Compare", func() { intSink = frequency.Strict().MustCompare(440, 441) })
	check("Strict.Near with error result", func() {
		var err error
		boolSink, err = frequency.Strict().Near(440, 440)
		if err != nil {
			panic(err)
		}
	})
	check("Strict.Compare with error result", func() {
		var err error
		intSink, err = frequency.Strict().Compare(440, 441)
		if err != nil {
			panic(err)
		}
	})
	check("WithinHertz", func() { criterionSink = frequency.MustWithinHertz(1) })
	check("WithinRelative", func() { criterionSink = frequency.MustWithinRelative(0.01) })
	check("WithinHertz with error result", func() {
		var err error
		criterionSink, err = frequency.WithinHertz(1)
		if err != nil {
			panic(err)
		}
	})
	check("WithinRelative with error result", func() {
		var err error
		criterionSink, err = frequency.WithinRelative(0.01)
		if err != nil {
			panic(err)
		}
	})

	criterion := frequency.MustWithinHertz(1)
	check("IsValid", func() { boolSink = criterion.IsValid() })
	check("Kind", func() { kindSink = criterion.Kind() })
	check("Tolerance", func() { floatSink, boolSink = criterion.Tolerance() })
	check("CentsTolerance", func() { centsSink, boolSink = criterion.CentsTolerance() })
	check("Components noncomposite", func() { componentsSink = criterion.Components() })
	check("Criterion.Equal noncomposite", func() { boolSink = criterion.Equal(frequency.Strict()) })
	check("Criterion.Equal one noncomposite", func() {
		boolSink = criterion.Equal(frequency.MustAllOf(frequency.Strict()))
	})

	for kind := frequency.Kind(0); kind <= frequency.KindAnyOf; kind++ {
		check("Kind methods "+kind.String(), func() {
			boolSink = kind.IsValid()
			boolSink = kind.Equal(frequency.KindStrict)
			boolSink = kind.HasTolerance()
			boolSink = kind.IsComposite()
			stringSink = kind.String()
		})
	}
}
