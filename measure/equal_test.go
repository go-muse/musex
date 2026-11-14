package measure_test

import (
	"testing"

	"github.com/go-muse/musex/measure"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Group Т-8, equalities: Equal, EqualMarkup, and EqualBarLengths
// (architecture 4.14, 4.15; О-19, О-20; ИС-11; П-13; КП-35, КП-36, КП-51,
// КП-68).

// equalCase is a pair of grids and their three equalities.
type equalCase struct {
	name    string
	a, b    string
	record  bool
	markup  bool
	lengths bool
}

// equalCases are the pairs of КП-35 and КП-51 and further pairs.
func equalCases() []equalCase {
	return []equalCase{
		{"КП-35 repeated time signature", "4/4", "4/4;5:4/4", false, true, true},
		{"КП-35 explicit length equal to the meter", "4/4", "4/4;3:length:1/1", false, true, true},
		{"КП-35 3/4 and 6/8", "3/4", "6/8", false, false, true},
		{"КП-35 pickup and a short bar 1", "3/4;pickup:1/4", "3/4;1:length:1/4", false, false, false},
		{"КП-35 alignment", "4/4;5:length:1/4", "4/4;5:length:1/4:end", false, false, true},
		{"КП-35 open and finite", "4/4", "4/4;bars:4", false, false, false},
		{"КП-51 whole-bar pickup", "4/4;pickup:1/1", "4/4", false, false, false},
		{"same record", "4/4;pickup:1/4;3:3/4;4:length:1/2:end;bars:8", "4/4;pickup:1/4;3:3/4;4:length:1/2:end;bars:8", true, true, true},
		{"C and 4/4", "C", "4/4", false, false, true},
		{"different counts", "4/4;pickup:1/4;bars:2", "4/4;pickup:1/4;bars:3", false, false, false},
		{"different pickups", "4/4;pickup:1/4", "4/4;pickup:1/8", false, false, false},
		{"lengths equal after a change", "4/4;3:3/4", "4/4;3:6/8", false, false, true},
		{"repeats and equal lengths", "4/4;3:3/4", "4/4;3:3/4;5:3/4;4:length:3/4", false, true, true},
		{"repeated mark with an explicit length", "4/4;3:4/4;3:length:1/2", "4/4;3:length:1/2", false, true, true},
		{"explicit length after a pickup of a whole bar", "4/4;pickup:1/1;1:length:1/2", "4/4;pickup:1/1;1:length:1/2:end", false, false, true},
		{"to the end of a whole bar", "4/4;5:length:1/1:end", "4/4", false, true, true},
		{"different bar later", "4/4;9:3/4", "4/4;10:3/4", false, false, false},
	}
}

// TestEqualGrids covers О-20, ИС-11, П-13, and КП-35, КП-51: the three
// equalities of grids; equality by record implies equality by markup, which
// implies equality of bar lengths.
func TestEqualGrids(t *testing.T) {
	t.Parallel()

	for _, tt := range equalCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			a, b := gridOf(t, tt.a), gridOf(t, tt.b)

			assert.Equal(t, tt.record, a.Equal(b), "Equal")
			assert.Equal(t, tt.record, b.Equal(a), "Equal is symmetric")

			markup, err := a.EqualMarkup(b)
			require.NoError(t, err)
			assert.Equal(t, tt.markup, markup, "EqualMarkup")

			markup, err = b.EqualMarkup(a)
			require.NoError(t, err)
			assert.Equal(t, tt.markup, markup, "EqualMarkup is symmetric")

			lengths, err := a.EqualBarLengths(b)
			require.NoError(t, err)
			assert.Equal(t, tt.lengths, lengths, "EqualBarLengths")

			lengths, err = b.EqualBarLengths(a)
			require.NoError(t, err)
			assert.Equal(t, tt.lengths, lengths, "EqualBarLengths is symmetric")

			assert.True(t, a.Equal(a))

			self, err := a.EqualMarkup(a)
			require.NoError(t, err)
			assert.True(t, self)

			self, err = a.EqualBarLengths(a)
			require.NoError(t, err)
			assert.True(t, self)
		})
	}
}

// TestEqualCopies covers 3.2: a copy of a grid by assignment and a grid
// built by the same operations are equal by record.
func TestEqualCopies(t *testing.T) {
	t.Parallel()
	defer guard(t)

	a := gridOf(t, "4/4;pickup:1/4;3:3/4;4:length:1/2;bars:8")
	b := a
	assert.True(t, a.Equal(b))
	assert.True(t, a.Equal(gridOf(t, "4/4;pickup:1/4;3:3/4;4:length:1/2;bars:8")))

	c := build(t, mt("4/4"), withBarCount(8), withLength(4, "1/2", measure.AlignStart), withMark(3, "3/4"), withPickup("1/4"))
	assert.True(t, a.Equal(c), "the order of the operations does not matter for the same record")
}

// TestEqualSimplify covers О-19, П-13, and КП-36, КП-68: Simplify keeps the
// equality by markup and changes the record only where there are repeats.
func TestEqualSimplify(t *testing.T) {
	t.Parallel()

	for _, text := range []string{
		"4/4;5:4/4;9:C;3:length:1/1;4:length:1/2",
		"4/4;pickup:1/4;3:3/4;4:length:1/2:end;bars:8",
		"3/4;2:3/4;3:3/4;4:2/4;5:2/4",
		"4/4;5:length:1/1:end",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := gridOf(t, text)
			s, err := g.Simplify()
			require.NoError(t, err)

			markup, err := s.EqualMarkup(g)
			require.NoError(t, err)
			assert.True(t, markup, "КП-36: equal by markup")

			again, err := s.Simplify()
			require.NoError(t, err)
			assert.True(t, again.Equal(s), "Simplify of a simplified grid gives the same record")
		})
	}

	t.Run("КП-36 differs by record", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		g := gridOf(t, "4/4;5:4/4;9:C;3:length:1/1;4:length:1/2")
		s, err := g.Simplify()
		require.NoError(t, err)
		assert.False(t, s.Equal(g))
		assert.True(t, s.Equal(gridOf(t, "4/4;9:C;4:length:1/2")))
	})

	t.Run("КП-68 open other", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		joined, err := gridOf(t, "4/4;bars:2").Append(gridOf(t, "4/4"))
		require.NoError(t, err)

		four := gridOf(t, "4/4")
		assert.False(t, joined.Equal(four), "different by record")

		markup, err := joined.EqualMarkup(four)
		require.NoError(t, err)
		assert.True(t, markup, "equal by markup")
	})

	t.Run("КП-68 pickup", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		joined, err := gridOf(t, "4/4;bars:2").Append(gridOf(t, "3/4;pickup:1/4;bars:2"))
		require.NoError(t, err)
		assert.True(t, joined.Equal(gridOf(t, "4/4;3:3/4;3:length:1/4:end;bars:5")))
	})
}
