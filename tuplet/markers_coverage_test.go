package tuplet_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
)

func TestFormatMustVariants(t *testing.T) {
	t.Parallel()

	m := tuplet.ABCMarker{Tuplet: tuplet.MustNew(3, 4), Events: 2}
	assert.Equal(t, "(3:4:2", tuplet.MustFormatABCMarker(m, tuplet.Binary))
	assert.Panics(t, func() { tuplet.MustFormatABCMarker(tuplet.ABCMarker{}, tuplet.Binary) })

	assert.Equal(t, `\tuplet 3/2`, tuplet.MustFormatLilyPondTuplet(tuplet.Triplet()))
	assert.Panics(t, func() { tuplet.MustFormatLilyPondTuplet(tuplet.Tuplet{}) })

	assert.Equal(t, "*2/3", tuplet.MustFormatLilyPondMultiplier(fraction.MustNew(2, 3)))
	assert.Panics(t, func() { tuplet.MustFormatLilyPondMultiplier(fraction.Fraction{}) })

	assert.Equal(t, tuplet.Triplet(), tuplet.MustParseLilyPondTuplet(`\tuplet 3/2`))
	assert.Equal(t, fraction.MustNew(2, 3), tuplet.MustParseLilyPondMultiplier("*2/3"))
	assert.Equal(t, tuplet.MustParseABCMarker("(3", tuplet.Binary), tuplet.MustParseABCMarker("(3", tuplet.Binary))
}

func TestUnapplyChainRange(t *testing.T) {
	t.Parallel()

	// A chain whose flattened counts exceed the fraction range is rejected.
	var chain []tuplet.Tuplet
	for range 5 {
		chain = append(chain, tuplet.MustNew(math.MaxInt, 1))
	}
	_, err := tuplet.UnapplyChain(fraction.One(), chain...)
	assert.ErrorIs(t, err, tuplet.ErrOutOfRange)
}
