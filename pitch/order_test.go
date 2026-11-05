package pitch_test

import (
	"fmt"
	"math"
	"slices"
	"testing"

	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/octave/number"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
)

func TestEqualAndMapKey(t *testing.T) {
	t.Parallel()

	t.Run("equivalent_pitches_equal_and_share_map_key", func(t *testing.T) {
		t.Parallel()

		equiv := []pitch.Pitch{
			pa(spelling.CSharp(), 4),
			pitch.MustNew(spelling.CSharp(), o(4)),
			pitch.MustNew(spelling.CSharp(), octave.MustFromNumber(3, number.FrancoBelgian())),
			pitch.MustNew(spelling.CSharp(), octave.MustFromName("первая", naming.Russian())),
			pa(spelling.C(), 4).MustRaise(),
			pa(spelling.CSharp(), 3).MustUp(),
			pa(spelling.CSharp(), 5).MustShiftDown(1),
		}

		m := make(map[pitch.Pitch]int)
		for idx, p := range equiv {
			m[p] = idx + 1
			for jdx, q := range equiv {
				msg := fmt.Sprintf("i=%d j=%d", idx, jdx)
				assert.True(t, p.Equal(q), msg)
				assert.True(t, p == q, msg)
				assert.Equal(t, 0, p.Compare(q), msg)
			}
		}
		assert.Len(t, m, 1)
		assert.Equal(t, len(equiv), m[pa(spelling.CSharp(), 4)])
	})

	t.Run("distinct_pitches_not_equal_and_distinct_map_keys", func(t *testing.T) {
		t.Parallel()

		pairs := []struct {
			name string
			p    pitch.Pitch
			q    pitch.Pitch
		}{
			{
				name: "enharmonic_CSharp_4_and_DFlat_4",
				p:    pa(spelling.CSharp(), 4),
				q:    pa(spelling.DFlat(), 4),
			},
			{
				name: "boundary_same_octave_BSharp_4_and_C_4",
				p:    pa(spelling.BSharp(), 4),
				q:    pa(spelling.C(), 4),
			},
			{
				name: "boundary_different_octave_BSharp_3_and_C_4",
				p:    pa(spelling.BSharp(), 3),
				q:    pa(spelling.C(), 4),
			},
			{
				name: "boundary_different_octave_CFlat_4_and_B_3",
				p:    pa(spelling.CFlat(), 4),
				q:    pa(spelling.B(), 3),
			},
			{
				name: "different_octave_C_4_and_C_5",
				p:    pa(spelling.C(), 4),
				q:    pa(spelling.C(), 5),
			},
			{
				name: "fractional_C_plus_half_4_and_C_4",
				p:    pa(cPlusHalf(), 4),
				q:    pa(spelling.C(), 4),
			},
		}

		for _, tc := range pairs {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				assert.False(t, tc.p.Equal(tc.q))
				assert.False(t, tc.q.Equal(tc.p))
				assert.False(t, tc.p == tc.q)

				m := map[pitch.Pitch]string{
					tc.p: "first",
					tc.q: "second",
				}
				assert.Len(t, m, 2)
				assert.Equal(t, "first", m[tc.p])
				assert.Equal(t, "second", m[tc.q])
			})
		}
	})

	t.Run("equivalence_relation_and_zero_value_semantics", func(t *testing.T) {
		t.Parallel()

		var z1, z2 pitch.Pitch
		assert.False(t, z1.Equal(z2))
		assert.True(t, z1 == z2)
		assert.False(t, z1.Equal(pa(spelling.C(), 4)))
		assert.False(t, pa(spelling.C(), 4).Equal(z1))
		assert.False(t, z1 == pa(spelling.C(), 4))

		m := map[pitch.Pitch]string{
			z1:                   "zero1",
			z2:                   "zero2",
			pa(spelling.C(), 4):  "c4",
			pa(spelling.C(), -1): "c-1",
		}
		assert.Len(t, m, 3)
		assert.Equal(t, "zero2", m[pitch.Pitch{}])

		// Reflexivity, symmetry, and transitivity on valid pitches.
		groups := [][]pitch.Pitch{
			{
				pa(spelling.C(), 4),
				pitch.MustNew(spelling.C(), o(4)),
				pa(spelling.CSharp(), 4).Natural(),
			},
			{
				pa(spelling.FSharp(), 2),
				pitch.MustNew(spelling.FSharp(), o(2)),
				pa(spelling.FSharp(), 3).MustDown(),
			},
			{
				pa(cPlusHalf(), math.MinInt),
				pitch.MustNew(cPlusHalf(), o(math.MinInt)),
			},
		}

		for g1, grp1 := range groups {
			for g2, grp2 := range groups {
				for _, x := range grp1 {
					assert.True(t, x.Equal(x))
					for _, y := range grp2 {
						wantEqual := g1 == g2
						assert.Equal(t, wantEqual, x.Equal(y))
						assert.Equal(t, wantEqual, y.Equal(x))
						assert.Equal(t, wantEqual, x == y)
						for _, z := range grp2 {
							if x.Equal(y) && y.Equal(z) {
								assert.True(t, x.Equal(z))
							}
						}
					}
				}
			}
		}
	})
}

func TestCompareAndSort(t *testing.T) {
	t.Parallel()

	t.Run("cross_octave_and_within_octave_comparisons", func(t *testing.T) {
		t.Parallel()

		// Cross-octave order is determined strictly by octave.
		assert.Equal(t, -1, pa(spelling.BDoubleSharp(), 3).Compare(pa(spelling.CDoubleFlat(), 4)))
		assert.Equal(t, 1, pa(spelling.CDoubleFlat(), 4).Compare(pa(spelling.BDoubleSharp(), 3)))
		assert.Equal(t, -1, pa(spelling.B(), math.MinInt).Compare(pa(spelling.C(), math.MaxInt)))
		assert.Equal(t, 1, pa(spelling.C(), math.MaxInt).Compare(pa(spelling.B(), math.MinInt)))

		// Within the same octave, Compare matches Spelling.Compare (not acoustic pitch).
		assert.Equal(t, -1, pa(spelling.C(), 4).Compare(pa(spelling.BSharp(), 4)))
		assert.Equal(t, 1, pa(spelling.BSharp(), 4).Compare(pa(spelling.C(), 4)))
		assert.Equal(t, -1, pa(spelling.CSharp(), 4).Compare(pa(spelling.DFlat(), 4)))
		assert.Equal(t, 1, pa(spelling.DFlat(), 4).Compare(pa(spelling.CSharp(), 4)))

		// Zero value precedes every valid pitch and equals itself in Compare.
		var z pitch.Pitch
		assert.Equal(t, 0, z.Compare(pitch.Pitch{}))
		assert.Equal(t, -1, z.Compare(pa(spelling.C(), math.MinInt)))
		assert.Equal(t, 1, pa(spelling.C(), math.MinInt).Compare(z))
	})

	t.Run("ordered_chain_within_octave_4", func(t *testing.T) {
		t.Parallel()

		chain := []pitch.Pitch{
			pa(spelling.CDoubleFlat(), 4),
			pa(spelling.CFlat(), 4),
			pa(spelling.C(), 4),
			pa(cPlusHalf(), 4),
			pa(spelling.CSharp(), 4),
			pa(spelling.CDoubleSharp(), 4),
			pa(spelling.DDoubleFlat(), 4),
			pa(spelling.DFlat(), 4),
			pa(spelling.D(), 4),
			pa(spelling.E(), 4),
			pa(spelling.F(), 4),
			pa(spelling.G(), 4),
			pa(spelling.A(), 4),
			pa(spelling.BFlat(), 4),
			pa(spelling.B(), 4),
			pa(spelling.BSharp(), 4),
			pa(spelling.BDoubleSharp(), 4),
		}

		for i := range chain {
			assert.Equal(t, 0, chain[i].Compare(chain[i]))
			for j := i + 1; j < len(chain); j++ {
				assert.Equal(t, -1, chain[i].Compare(chain[j]), "i=%d (%v) j=%d (%v)", i, chain[i], j, chain[j])
				assert.Equal(t, 1, chain[j].Compare(chain[i]), "i=%d (%v) j=%d (%v)", i, chain[i], j, chain[j])
			}
		}
	})

	t.Run("total_order_laws_and_sort_permutations", func(t *testing.T) {
		t.Parallel()

		var ordered []pitch.Pitch
		ordered = append(ordered, pitch.Pitch{})

		octaves := []int{math.MinInt, 3, 4, math.MaxInt}
		namesInOrder := []spelling.Spelling{
			spelling.CDoubleFlat(),
			spelling.C(),
			cPlusHalf(),
			spelling.CSharp(),
			spelling.DFlat(),
			spelling.B(),
			spelling.BSharp(),
			spelling.BDoubleSharp(),
		}
		for _, a := range octaves {
			for _, s := range namesInOrder {
				ordered = append(ordered, pa(s, a))
			}
		}

		for i, p := range ordered {
			for j, q := range ordered {
				got := p.Compare(q)
				switch {
				case i < j:
					assert.Equal(t, -1, got, "i=%d j=%d", i, j)
					assert.False(t, p == q)
				case i == j:
					assert.Equal(t, 0, got, "i=%d j=%d", i, j)
					assert.True(t, p == q)
				default:
					assert.Equal(t, 1, got, "i=%d j=%d", i, j)
					assert.False(t, p == q)
				}
				assert.Equal(t, -q.Compare(p), got, "antisymmetry i=%d j=%d", i, j)

				for k, r := range ordered {
					if p.Compare(q) < 0 && q.Compare(r) < 0 {
						assert.Equal(t, -1, p.Compare(r), "transitivity i=%d j=%d k=%d", i, j, k)
					}
				}
			}
		}

		// Reversed copy and rotated copy must sort back to ordered.
		reversed := slices.Clone(ordered)
		slices.Reverse(reversed)
		slices.SortFunc(reversed, pitch.Pitch.Compare)
		assert.Equal(t, ordered, reversed)

		interleaved := make([]pitch.Pitch, 0, len(ordered))
		for i := 0; i < len(ordered); i += 2 {
			interleaved = append(interleaved, ordered[i])
		}
		for i := 1; i < len(ordered); i += 2 {
			interleaved = append(interleaved, ordered[i])
		}
		slices.SortFunc(interleaved, pitch.Pitch.Compare)
		assert.Equal(t, ordered, interleaved)
	})
}
