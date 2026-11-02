package letter_test

import (
	"fmt"
	"math"
	"strconv"
	"testing"

	"github.com/go-muse/musex/letter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// K-1: the list of letters and their number.
func TestAll(t *testing.T) {
	t.Parallel()

	want := oracle()
	got := letter.All()

	require.Len(t, got, oracleCount)
	assert.Equal(t, want[:], got, "canonical order C D E F G A B")
	assert.Equal(t, letter.Count, cap(got), "a fresh slice of length and capacity Count")
	assert.Len(t, got, letter.Count)

	seen := make(map[letter.Letter]int, len(got))
	for i, l := range got {
		if prev, ok := seen[l]; ok {
			assert.Failf(t, "duplicate letter", "index %d repeats index %d", i, prev)
		}
		seen[l] = i
	}
	assert.Len(t, seen, oracleCount, "each letter exactly once")
}

// K-1: Count is an untyped constant equal to 7, usable as an array length
// and in arithmetic of any integer type.
func TestCount(t *testing.T) {
	t.Parallel()

	var table [letter.Count]string
	var asInt64 int64 = letter.Count
	var asUint8 uint8 = letter.Count

	assert.Equal(t, oracleCount, letter.Count)
	assert.Len(t, table, oracleCount)
	assert.Equal(t, int64(oracleCount), asInt64)
	assert.Equal(t, uint8(oracleCount), asUint8)
	assert.Equal(t, letter.Count, len(letter.All()))
}

// K-2: walking the circle upward from fixed letters.
func TestLetter_Cycle_Fixed(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		start letter.Letter
		want  []letter.Letter
	}{
		{
			name:  "from C",
			start: letter.C,
			want:  []letter.Letter{letter.C, letter.D, letter.E, letter.F, letter.G, letter.A, letter.B},
		},
		{
			name:  "from A",
			start: letter.A,
			want:  []letter.Letter{letter.A, letter.B, letter.C, letter.D, letter.E, letter.F, letter.G},
		},
		{
			name:  "from B",
			start: letter.B,
			want:  []letter.Letter{letter.B, letter.C, letter.D, letter.E, letter.F, letter.G, letter.A},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, tc.start.Cycle())
		})
	}
}

// K-2: for every start letter the cycle has seven distinct letters and
// element i is the shift of the start by i. Ownership contract: the result
// is a fresh slice of length and capacity Count, as for All.
func TestLetter_Cycle_AllStarts(t *testing.T) {
	t.Parallel()

	names := oracleNames()
	for ord, start := range oracle() {
		t.Run(names[ord], func(t *testing.T) {
			t.Parallel()

			got := start.Cycle()
			require.Len(t, got, oracleCount)
			assert.Equal(t, letter.Count, cap(got), "a fresh slice of length and capacity Count")

			seen := make(map[letter.Letter]bool, len(got))
			for i, l := range got {
				assert.Equal(t, expectedShift(ord, i), l, "element %d by the oracle", i)
				assert.Equal(t, start.Shift(i), l, "element %d equals Shift(%d)", i, i)
				seen[l] = true
			}
			assert.Len(t, seen, oracleCount, "all elements are distinct")
		})
	}

	t.Run("C cycle equals All", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, letter.All(), letter.C.Cycle())
	})
}

// K-3, K-9: upward distance, fixed examples.
func TestLetter_DistanceUp_Examples(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		from, to letter.Letter
		want     int
	}{
		{name: "C to E", from: letter.C, to: letter.E, want: 2},
		{name: "B to D", from: letter.B, to: letter.D, want: 2},
		{name: "E to F", from: letter.E, to: letter.F, want: 1},
		{name: "G to C", from: letter.G, to: letter.C, want: 3},
		{name: "B to C across the seam", from: letter.B, to: letter.C, want: 1},
		{name: "C to B", from: letter.C, to: letter.B, want: 6},
		{name: "C to C", from: letter.C, to: letter.C, want: 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, tc.from.DistanceUp(tc.to))
		})
	}
}

// K-3, K-9: upward distance for all 49 pairs.
func TestLetter_DistanceUp_AllPairs(t *testing.T) {
	t.Parallel()

	names := oracleNames()
	for i, from := range oracle() {
		t.Run(names[i], func(t *testing.T) {
			t.Parallel()

			for j, to := range oracle() {
				got := from.DistanceUp(to)
				assert.GreaterOrEqual(t, got, 0, "%s to %s", names[i], names[j])
				assert.LessOrEqual(t, got, 6, "%s to %s", names[i], names[j])
				assert.Equal(t, mod7(j-i), got, "%s to %s", names[i], names[j])
			}
		})
	}
}

// K-4, K-9: downward distance, fixed examples.
func TestLetter_DistanceDown_Examples(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		from, to letter.Letter
		want     int
	}{
		{name: "C to E", from: letter.C, to: letter.E, want: 5},
		{name: "E to C", from: letter.E, to: letter.C, want: 2},
		{name: "D to B", from: letter.D, to: letter.B, want: 2},
		{name: "C to B across the seam", from: letter.C, to: letter.B, want: 1},
		{name: "C to C", from: letter.C, to: letter.C, want: 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, tc.from.DistanceDown(tc.to))
		})
	}
}

// K-4, K-9: downward distance for all 49 pairs.
func TestLetter_DistanceDown_AllPairs(t *testing.T) {
	t.Parallel()

	names := oracleNames()
	for i, a := range oracle() {
		t.Run(names[i], func(t *testing.T) {
			t.Parallel()

			for j, b := range oracle() {
				msg := fmt.Sprintf("%s to %s", names[i], names[j])
				down := a.DistanceDown(b)
				up := a.DistanceUp(b)

				assert.Equal(t, mod7(i-j), down, msg)
				assert.Equal(t, b.DistanceUp(a), down, "down(a, b) == up(b, a): "+msg)
				if i == j {
					assert.Equal(t, 0, down, msg)
					assert.Equal(t, 0, up, msg)
				} else {
					assert.Equal(t, oracleCount, up+down, "up + down == 7: "+msg)
				}
			}
		})
	}
}

// K-5: shift, fixed examples.
func TestLetter_Shift_Examples(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		start letter.Letter
		steps int
		want  letter.Letter
	}{
		{name: "C+0", start: letter.C, steps: 0, want: letter.C},
		{name: "C+1", start: letter.C, steps: 1, want: letter.D},
		{name: "B+1 across the seam", start: letter.B, steps: 1, want: letter.C},
		{name: "C+9", start: letter.C, steps: 9, want: letter.E},
		{name: "G+4", start: letter.G, steps: 4, want: letter.D},
		{name: "C+7", start: letter.C, steps: 7, want: letter.C},
		{name: "C+70", start: letter.C, steps: 70, want: letter.C},
		{name: "C-1 across the seam", start: letter.C, steps: -1, want: letter.B},
		{name: "C-7", start: letter.C, steps: -7, want: letter.C},
		{name: "C-8", start: letter.C, steps: -8, want: letter.B},
		{name: "E-2", start: letter.E, steps: -2, want: letter.C},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, tc.start.Shift(tc.steps))
		})
	}
}

// K-5: every letter shifted by every n in [-15, 15] matches the formula and
// shifting back returns the start.
func TestLetter_Shift_SmallRange(t *testing.T) {
	t.Parallel()

	names := oracleNames()
	for ord, start := range oracle() {
		t.Run(names[ord], func(t *testing.T) {
			t.Parallel()

			for n := -15; n <= 15; n++ {
				got := start.Shift(n)
				assert.Equal(t, expectedShift(ord, n), got, "%s shifted by %d", names[ord], n)
				assert.Equal(t, start, got.Shift(-n), "%s shifted by %d and back", names[ord], n)
				assert.Equal(t, start.Shift(n), start.Shift(n+oracleCount), "period 7 at %d", n)
			}
		})
	}
}

// K-5, K-8: a fixed sample of large step counts for every letter.
// The module supports only 64-bit int (human decision O-7).
func TestLetter_Shift_LargeSteps(t *testing.T) {
	t.Parallel()

	samples := []int{
		1_000_000_000,
		-1_000_000_000,
		1_000_000_003,
		-1_000_000_003,
		1 << 31,
		-(1 << 31),
		1<<31 - 1,
		-(1<<31 - 1),
		1 << 40,
		-(1 << 40),
		7 << 40,
		-(7 << 40),
		1<<62 + 5,
		-(1<<62 + 5),
		math.MaxInt - math.MaxInt%7,
		math.MinInt - math.MinInt%7,
	}

	names := oracleNames()
	for _, n := range samples {
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			t.Parallel()

			for ord, start := range oracle() {
				got := start.Shift(n)
				assert.Equal(t, expectedShift(ord, n), got, "%s shifted by %d", names[ord], n)
				assert.Equal(t, start, got.Shift(-n), "%s shifted by %d and back", names[ord], n)
			}
		})
	}

	t.Run("multiples of 7 near the int limits keep the letter", func(t *testing.T) {
		t.Parallel()

		for ord, start := range oracle() {
			assert.Equal(t, start, start.Shift(math.MaxInt-math.MaxInt%7), names[ord])
			assert.Equal(t, start, start.Shift(math.MinInt-math.MinInt%7), names[ord])
		}
	})
}

// K-8: the shift is correct at the limits of int, without overflow.
func TestLetter_Shift_IntLimits(t *testing.T) {
	t.Parallel()

	names := oracleNames()
	for ord, start := range oracle() {
		t.Run(names[ord], func(t *testing.T) {
			t.Parallel()

			for k := range 14 {
				nearMax := math.MaxInt - k
				assert.Equal(t, expectedShift(ord, nearMax), start.Shift(nearMax), "MaxInt-%d", k)
				assert.Equal(t, start, start.Shift(nearMax).Shift(-nearMax), "MaxInt-%d and back", k)

				nearMin := math.MinInt + k
				assert.Equal(t, expectedShift(ord, nearMin), start.Shift(nearMin), "MinInt+%d", k)
				if nearMin != math.MinInt {
					// -math.MinInt overflows in Go itself, so the round trip
					// is checked only for the other values.
					assert.Equal(t, start, start.Shift(nearMin).Shift(-nearMin), "MinInt+%d and back", k)
				}
			}
		})
	}

	t.Run("C at the limits of 64-bit int", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, letter.C, letter.C.Shift(math.MaxInt), "2^63-1 = 0 mod 7")
		assert.Equal(t, letter.B, letter.C.Shift(math.MinInt), "-2^63 = 6 mod 7")
	})
}

// K-5: shifting by the distance reaches the target for all 49 pairs.
func TestLetter_Shift_ByDistance(t *testing.T) {
	t.Parallel()

	names := oracleNames()
	for i, a := range oracle() {
		t.Run(names[i], func(t *testing.T) {
			t.Parallel()

			for j, b := range oracle() {
				msg := fmt.Sprintf("%s to %s", names[i], names[j])
				assert.Equal(t, b, a.Shift(a.DistanceUp(b)), "up: "+msg)
				assert.Equal(t, b, a.Shift(-a.DistanceDown(b)), "down: "+msg)
			}
		})
	}
}

// K-6: letters obtained in different ways are the same letter, checked both
// with Equal and with ==.
func TestLetter_Equal_Examples(t *testing.T) {
	t.Parallel()

	cycleFromA := letter.A.Cycle()
	require.Len(t, cycleFromA, oracleCount)
	all := letter.All()
	require.Len(t, all, oracleCount)

	cases := []struct {
		name     string
		got      letter.Letter
		want     letter.Letter
		wantSame bool
	}{
		{name: "C shifted by 9 is E", got: letter.C.Shift(9), want: letter.E, wantSame: true},
		{name: "B shifted by 1 is C", got: letter.B.Shift(1), want: letter.C, wantSame: true},
		{name: "index 2 of the cycle from A is C from the list", got: cycleFromA[2], want: all[0], wantSame: true},
		{name: "C shifted by 9 is not D", got: letter.C.Shift(9), want: letter.D, wantSame: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.wantSame, tc.got.Equal(tc.want), "Equal")
			assert.Equal(t, tc.wantSame, tc.got == tc.want, "==")
		})
	}
}

// K-6: the seven letters are pairwise distinct; Equal agrees with == and is
// symmetric for valid letters.
func TestLetter_Equal_AllPairs(t *testing.T) {
	t.Parallel()

	names := oracleNames()
	for i, a := range oracle() {
		t.Run(names[i], func(t *testing.T) {
			t.Parallel()

			for j, b := range oracle() {
				msg := fmt.Sprintf("%s and %s", names[i], names[j])
				assert.Equal(t, i == j, a.Equal(b), "Equal: "+msg)
				assert.Equal(t, i == j, a == b, "==: "+msg)
				assert.Equal(t, a == b, a.Equal(b), "Equal agrees with ==: "+msg)
				assert.Equal(t, b.Equal(a), a.Equal(b), "symmetry: "+msg)
			}
		})
	}
}

// K-6: the same letter obtained in different ways is one map key.
func TestLetter_MapKey(t *testing.T) {
	t.Parallel()

	counts := map[letter.Letter]int{}
	counts[letter.C]++
	counts[letter.C.Shift(7)]++
	counts[letter.B.Shift(1)]++
	counts[letter.E.Shift(-2)]++
	counts[letter.MustFromOrdinal(0)]++

	assert.Len(t, counts, 1)
	assert.Equal(t, 5, counts[letter.C])
}

// Q-3: ordinals of the valid letters.
func TestLetter_Ordinal(t *testing.T) {
	t.Parallel()

	names := oracleNames()
	for ord, l := range oracle() {
		t.Run(names[ord], func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, ord, l.Ordinal())
			assert.Equal(t, letter.C.DistanceUp(l), l.Ordinal(), "equals the distance up from C")
		})
	}
}

// Q-3: creating letters from valid ordinals.
func TestFromOrdinal_Valid(t *testing.T) {
	t.Parallel()

	names := oracleNames()
	for ord, want := range oracle() {
		t.Run(names[ord], func(t *testing.T) {
			t.Parallel()

			got, err := letter.FromOrdinal(ord)
			require.NoError(t, err)
			assert.Equal(t, want, got)

			back, err := letter.FromOrdinal(want.Ordinal())
			require.NoError(t, err)
			assert.Equal(t, want, back, "FromOrdinal(l.Ordinal()) == l")
		})
	}
}

// Q-3: the Must wrapper for valid ordinals.
func TestMustFromOrdinal_Valid(t *testing.T) {
	t.Parallel()

	names := oracleNames()
	for ord, want := range oracle() {
		t.Run(names[ord], func(t *testing.T) {
			t.Parallel()

			var got letter.Letter
			require.NotPanics(t, func() { got = letter.MustFromOrdinal(ord) })
			assert.Equal(t, want, got)
		})
	}
}

// Q-4: string form of the valid letters.
func TestLetter_String(t *testing.T) {
	t.Parallel()

	names := oracleNames()
	for ord, l := range oracle() {
		t.Run(names[ord], func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, names[ord], l.String())
			assert.Equal(t, names[ord], fmt.Sprint(l), "fmt uses String")
		})
	}

	t.Run("list printed with fmt", func(t *testing.T) {
		t.Parallel()

		// Same output as fmt.Sprint(letter.All()); the explicit %v verb avoids
		// the staticcheck QF1010 false positive for slices of a uint8-based
		// type whose elements implement fmt.Stringer.
		assert.Equal(t, "[C D E F G A B]", fmt.Sprintf("%v", letter.All()))
	})
}

// The numeric values of the constants are stable internal values C = 1 ... B = 7.
func TestConstantValues(t *testing.T) {
	t.Parallel()

	names := oracleNames()
	for ord, l := range oracle() {
		t.Run(names[ord], func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, uint8(ord+1), uint8(l))
		})
	}
}

// K-14: for all valid letters the operations neither panic nor fail, and
// every result is a valid letter result (not one of the invalid markers).
func TestValidLetters_NoPanics(t *testing.T) {
	t.Parallel()

	names := oracleNames()
	for ord, l := range oracle() {
		t.Run(names[ord], func(t *testing.T) {
			t.Parallel()

			var (
				allLen, cycleLen, ordinal, equals int
				str                               string
				valid                             bool
				distances                         []int
				shifted                           []letter.Letter
				fromOrdinal, fromMust             letter.Letter
				err                               error
			)
			require.NotPanics(t, func() {
				allLen = len(letter.All())
				cycleLen = len(l.Cycle())
				ordinal = l.Ordinal()
				str = l.String()
				valid = l.IsValid()
				for _, other := range oracle() {
					distances = append(distances, l.DistanceUp(other), l.DistanceDown(other))
					if l.Equal(other) {
						equals++
					}
				}
				for n := -15; n <= 15; n++ {
					shifted = append(shifted, l.Shift(n))
				}
				shifted = append(shifted, l.Shift(math.MaxInt), l.Shift(math.MinInt))
				fromOrdinal, err = letter.FromOrdinal(ord)
				fromMust = letter.MustFromOrdinal(ord)
			})

			assert.Equal(t, oracleCount, allLen, "All length")
			assert.Equal(t, oracleCount, cycleLen, "Cycle length")
			assert.Equal(t, ord, ordinal, "Ordinal")
			assert.Equal(t, names[ord], str, "String")
			assert.True(t, valid, "IsValid")
			assert.Equal(t, 1, equals, "Equal is true for exactly one letter")
			assert.Len(t, distances, 2*oracleCount)
			for i, d := range distances {
				assert.GreaterOrEqual(t, d, 0, "distance %d", i)
				assert.LessOrEqual(t, d, 6, "distance %d", i)
			}
			assert.Len(t, shifted, 33)
			for i, s := range shifted {
				assert.True(t, s.IsValid(), "shift result %d is a letter", i)
			}
			require.NoError(t, err)
			assert.Equal(t, l, fromOrdinal, "FromOrdinal")
			assert.Equal(t, l, fromMust, "MustFromOrdinal")
		})
	}
}
