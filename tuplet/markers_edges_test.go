package tuplet_test

import (
	"math"
	"strconv"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
)

func TestUnapplyChainErrors(t *testing.T) {
	t.Parallel()

	_, err := tuplet.UnapplyChain(fraction.MustNewReduced(1, 8), tuplet.Tuplet{})
	assert.ErrorIs(t, err, tuplet.ErrUnset)

	_, err = tuplet.UnapplyChain(fraction.Fraction{}, tuplet.Triplet())
	assert.ErrorIs(t, err, tuplet.ErrInvalidArgument)

	assert.Panics(t, func() { tuplet.MustUnapplyChain(fraction.Fraction{}, tuplet.Triplet()) })
}

func TestABCMarkerErrors(t *testing.T) {
	t.Parallel()

	// "(3::" is a valid marker equal to "(3" (abc 2.1, section 4.13; issue
	// #158), and "(3:" and "(3:2:" read as "(3" and "(3:2" (issue #162), so
	// they are not listed here. The rows are KP-22.
	for _, text := range []string{
		"", "3", "(0", "(01", "(3:0", "(3:2:0", "(3:2:2:2", "(-3", "( 3",
		"(:", "(0:", "(01:", "(3:0:", "(3:02", "(3:02:", "(3:2:2:", "(3:::", "(3::2:", "(3: ",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()

			_, err := tuplet.ParseABCMarker(text, tuplet.Binary)
			assert.ErrorIs(t, err, tuplet.ErrInvalidText)
		})
	}

	_, err := tuplet.ParseABCMarker("(3", tuplet.Division(9))
	assert.ErrorIs(t, err, tuplet.ErrNoDefault)

	_, err = tuplet.ParseABCMarker("(99999999999999999999:2", tuplet.Binary)
	assert.ErrorIs(t, err, tuplet.ErrOutOfRange)

	_, err = tuplet.ParseABCMarker("(99999999999999999999", tuplet.Binary)
	assert.ErrorIs(t, err, tuplet.ErrOutOfRange)

	_, err = tuplet.FormatABCMarker(tuplet.ABCMarker{}, tuplet.Binary)
	assert.ErrorIs(t, err, tuplet.ErrUnset)

	_, err = tuplet.FormatABCMarker(tuplet.ABCMarker{Tuplet: tuplet.Triplet(), Events: 0}, tuplet.Binary)
	assert.ErrorIs(t, err, tuplet.ErrInvalidCount)
}

func TestLilyPondErrors(t *testing.T) {
	t.Parallel()

	// `\tuplet 03/2` reads as `\tuplet 3/2` since issue #162, so it is not
	// listed here. The rows are KP-22.
	for _, text := range []string{
		"", "3/2", "\\tuplet 3", "\\tuplet 0/2", "\\tuplet 3/0", "\\tuplet 3/2 ", "\\tuplet -3/2",
		"\\tuplet 00/2", "\\tuplet 3/00", "\\tuplet /2", "\\tuplet 3/", "\\tuplet +03/2",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()

			_, err := tuplet.ParseLilyPondTuplet(text)
			assert.ErrorIs(t, err, tuplet.ErrInvalidText)
		})
	}

	_, err := tuplet.ParseLilyPondTuplet(`\tuplet 99999999999999999999/2`)
	assert.ErrorIs(t, err, tuplet.ErrOutOfRange)

	for _, text := range []string{
		"", "2/3", "*0/3", "*2/0", "*2/3 ",
		"*0", "*00", "*00/3", "*2/00", "*", "*/3", "*2/", "*-02/3",
	} {
		_, err := tuplet.ParseLilyPondMultiplier(text)
		assert.ErrorIs(t, err, tuplet.ErrInvalidText, text)
	}

	_, err = tuplet.ParseLilyPondMultiplier("*09223372036854775808")
	assert.ErrorIs(t, err, tuplet.ErrOutOfRange)

	_, err = tuplet.FormatLilyPondTuplet(tuplet.Tuplet{})
	assert.ErrorIs(t, err, tuplet.ErrUnset)

	_, err = tuplet.FormatLilyPondMultiplier(fraction.Fraction{})
	assert.ErrorIs(t, err, tuplet.ErrInvalidArgument)

	_, err = tuplet.FormatLilyPondMultiplier(fraction.MustNew(-2, 3))
	assert.ErrorIs(t, err, tuplet.ErrInvalidCount)
}

// The ABC short-form defaults follow ABC 2.1, not tuplet.Default.
func TestABCDefaultsDifferFromDefault(t *testing.T) {
	t.Parallel()

	m := tuplet.MustParseABCMarker("(5", tuplet.Binary)
	assert.Equal(t, tuplet.MustNew(5, 2), m.Tuplet)

	d := tuplet.MustDefault(5, tuplet.Binary)
	assert.Equal(t, tuplet.MustNew(5, 4), d)

	// Counts up to MaxInt are accepted when explicit.
	big := tuplet.MustParseABCMarker("(3:"+strconv.Itoa(math.MaxInt), tuplet.Binary)
	assert.Equal(t, math.MaxInt, big.Tuplet.MustNormal())
}
