package measure_test

import (
	"math"
	"slices"
	"testing"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/measure"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Group Т-3: addresses and positions (architecture 4.7; О-9, О-10, О-21;
// ИС-6, ИС-7; П-9; КП-3, КП-4, КП-5, КП-11, КП-12, КП-37, КП-41, КП-42,
// КП-49, КП-50, КП-51, КП-53, КП-63, КП-64).

// TestAddressValue covers О-21: NewAddress, Bar, Offset, IsSet, Equal, and
// == of addresses; an address does not know a grid.
func TestAddressValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		bar    int
		offset string
	}{
		{"bar 2, 1/8", 2, "1/8"},
		{"bar 0, 0", 0, "0"},
		{"offset longer than any bar", 2, "5/4"},
		{"max int", math.MaxInt, "3/7"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			a, err := measure.NewAddress(tt.bar, pos(tt.offset))
			require.NoError(t, err)
			assert.True(t, a.IsSet())

			bar, err := a.Bar()
			require.NoError(t, err)
			assert.Equal(t, tt.bar, bar)

			off, err := a.Offset()
			require.NoError(t, err)
			assertText(t, canonicalRat(tt.offset), off)

			same := measure.MustNewAddress(tt.bar, pos(tt.offset))
			assert.True(t, a.Equal(same))
			assert.Equal(t, same, a, "== matches Equal")

			otherBar := tt.bar + 1
			if tt.bar == math.MaxInt {
				otherBar = tt.bar - 1
			}

			other := measure.MustNewAddress(otherBar, pos(tt.offset))
			assert.False(t, a.Equal(other))
			assert.NotEqual(t, other, a)
		})
	}

	t.Run("unreduced offset", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		a := measure.MustNewAddress(2, duration.MustNewPosition(2, 16))
		assert.Equal(t, ad(2, "1/8"), a)
	})
}

// TestAddressErrors covers О-21, КП-39, and architecture 4.2: Е-2 for a
// negative bar, then for an unset offset together with duration.ErrUnset.
func TestAddressErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		bar     int
		offset  duration.Position
		causes  []error
		without []error
	}{
		{"КП-39 bar −1", -1, pos("0"), nil, nil},
		{"min int", math.MinInt, pos("1/8"), nil, nil},
		{"unset offset", 2, duration.Position{}, []error{duration.ErrUnset}, nil},
		{"bar −1 before unset offset", -1, duration.Position{}, nil, []error{duration.ErrUnset}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			a, err := measure.NewAddress(tt.bar, tt.offset)
			assertErr(t, err, measure.ErrInvalidArgument, tt.causes...)
			assertNoCause(t, err, tt.without...)
			assert.Equal(t, measure.Address{}, a)
			assert.False(t, a.IsSet())
		})
	}
}

// TestAddressCompare covers О-21 and КП-37: addresses order by bar and then
// by offset; Compare suits slices.SortFunc.
func TestAddressCompare(t *testing.T) {
	t.Parallel()
	defer guard(t)

	ordered := []measure.Address{ad(0, "3/4"), ad(2, "0"), ad(2, "1/8"), ad(2, "5/4"), ad(3, "0"), ad(math.MaxInt, "0")}

	for i, a := range ordered {
		for j, b := range ordered {
			want := 0

			switch {
			case i < j:
				want = -1
			case i > j:
				want = 1
			}

			assert.Equal(t, want, a.Compare(b), "%s vs %s", a, b)
		}
	}

	shuffled := []measure.Address{ordered[4], ordered[1], ordered[5], ordered[3], ordered[0], ordered[2]}
	slices.SortFunc(shuffled, measure.Address.Compare)
	assert.Equal(t, ordered, shuffled)

	assert.True(t, ad(2, "1/8").Equal(ad(2, "1/8")), "КП-37 equal")
	assert.Equal(t, 0, ad(2, "1/8").Compare(measure.MustNewAddress(2, duration.MustNewPosition(3, 24))))
}

// locateCase is a position of a grid and its address.
type locateCase struct {
	name    string
	grid    string
	pos     string
	address measure.Address
}

// locateCases are the examples of ИС-6 and ИС-7.
func locateCases() []locateCase {
	return []locateCase{
		{"КП-3 0", "3/4;pickup:1/4", "0", ad(0, "0")},
		{"КП-3 1/4", "3/4;pickup:1/4", "1/4", ad(1, "0")},
		{"КП-3 1", "3/4;pickup:1/4", "1", ad(2, "0")},
		{"КП-3 9/8", "3/4;pickup:1/4", "9/8", ad(2, "1/8")},
		{"КП-4 3/4", "3/4;pickup:1/4", "3/4", ad(1, "1/2")},
		{"КП-49 4/4 at 1", "4/4", "1", ad(2, "0")},
		{"КП-49 pickup end", "3/4;pickup:1/4", "1/4", ad(1, "0")},
		{"КП-50 0 without pickup", "4/4", "0", ad(1, "0")},
		{"КП-5 11/4", "4/4;3:3/4", "11/4", ad(4, "0")},
		{"КП-5 9/4", "4/4;3:3/4", "9/4", ad(3, "1/4")},
		{"КП-51 0", "4/4;pickup:1/1", "0", ad(0, "0")},
		{"КП-51 1", "4/4;pickup:1/1", "1", ad(1, "0")},
		{"КП-53 N = 1 at 0", "4/4;bars:1", "0", ad(1, "0")},
		{"КП-63 pickup removed", "3/4", "0", ad(1, "0")},
		{"just before the end of a finite grid", "4/4;bars:4", "31/8", ad(4, "7/8")},
		{"explicit length", "4/4;3:length:1/2", "21/8", ad(4, "1/8")},
		{"extended bar", "4/4;2:length:3/2", "9/4", ad(2, "5/4")},
		{"to the end", "4/4;5:length:1/4:end", "33/8", ad(5, "1/8")},
		{"triplet offset", "3/4", "5/6", ad(2, "1/12")},
		{"far", "4/4", "1000001/2", ad(500001, "1/2")},
	}
}

// TestLocate covers О-9, ИС-6, and the examples of locateCases: Locate
// gives the address, and PositionOf gives the position back.
func TestLocate(t *testing.T) {
	t.Parallel()

	for _, tt := range locateCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := gridOf(t, tt.grid)

			a, err := g.Locate(pos(tt.pos))
			require.NoError(t, err)
			assert.Equal(t, tt.address, a)

			p, err := g.PositionOf(tt.address)
			require.NoError(t, err)
			assertText(t, canonicalRat(tt.pos), p)

			ok, err := g.Contains(tt.address)
			require.NoError(t, err)
			assert.True(t, ok)
		})
	}
}

// TestLocateBarAt covers О-9 and КП-11: BarAt gives the information of the
// bar that contains a position, equal to Bar of the located bar.
func TestLocateBarAt(t *testing.T) {
	t.Parallel()

	for _, tt := range locateCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := gridOf(t, tt.grid)

			got, err := g.BarAt(pos(tt.pos))
			require.NoError(t, err)

			bar, err := tt.address.Bar()
			require.NoError(t, err)

			want, err := g.Bar(bar)
			require.NoError(t, err)
			assert.Equal(t, want, got)
		})
	}

	t.Run("КП-11 3/4 at 1/4", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		b, err := gridOf(t, "3/4").BarAt(pos("1/4"))
		require.NoError(t, err)
		assertText(t, "0/1", b.Start)
		assertText(t, "3/4", b.End)
		assert.Equal(t, 1, b.Number)
	})
}

// TestLocateRemaining covers О-9 and КП-11: the rest of the bar from a
// position to its end, always greater than zero.
func TestLocateRemaining(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		grid string
		pos  string
		want string
	}{
		{"КП-11 3/4 at 1/4", "3/4", "1/4", "1/2"},
		{"at a bar line", "3/4", "3/4", "3/4"},
		{"pickup", "3/4;pickup:1/4", "1/8", "1/8"},
		{"explicit length", "4/4;3:length:1/2", "17/8", "3/8"},
		{"last bar of a finite grid", "4/4;bars:4", "7/2", "1/2"},
		{"after a change", "4/4;3:3/4", "5/2", "1/4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			r, err := gridOf(t, tt.grid).Remaining(pos(tt.pos))
			require.NoError(t, err)
			assertText(t, canonicalRat(tt.want), r)
		})
	}
}

// TestLocateOnBarline covers О-9, КП-11, and КП-53: a position is on a bar
// line at the start of a bar or at the end of a finite grid.
func TestLocateOnBarline(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		grid string
		pos  string
		want bool
	}{
		{"КП-11 1/4", "3/4", "1/4", false},
		{"КП-11 3/4", "3/4", "3/4", true},
		{"КП-11 0", "3/4", "0", true},
		{"КП-53 end of N = 1", "4/4;bars:1", "1", true},
		{"КП-41 end of N = 4", "4/4;bars:4", "4", true},
		{"pickup start", "3/4;pickup:1/4", "0", true},
		{"pickup end", "3/4;pickup:1/4", "1/4", true},
		{"inside the pickup", "3/4;pickup:1/4", "1/8", false},
		{"after explicit length", "4/4;3:length:1/2", "5/2", true},
		{"inside an extended bar", "4/4;2:length:3/2", "2", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			got, err := gridOf(t, tt.grid).OnBarline(pos(tt.pos))
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestLocateErrors covers architecture 4.7 and КП-39, КП-41: Е-2 for an
// unset position together with duration.ErrUnset; Е-4 for a position at
// or after the end of a finite grid (OnBarline: after the end).
func TestLocateErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		grid     string
		pos      duration.Position
		err      error
		onBar    error // nil: OnBarline succeeds
		causes   []error
		onBarRes bool
	}{
		{"КП-39 unset", "4/4", duration.Position{}, measure.ErrInvalidArgument, measure.ErrInvalidArgument, []error{duration.ErrUnset}, false},
		{"КП-41 at the end 4", "4/4;bars:4", pos("4"), measure.ErrOutsideGrid, nil, nil, true},
		{"КП-41 after the end 5", "4/4;bars:4", pos("5"), measure.ErrOutsideGrid, measure.ErrOutsideGrid, nil, false},
		{"КП-53 after the end", "4/4;bars:1", pos("5/4"), measure.ErrOutsideGrid, measure.ErrOutsideGrid, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := gridOf(t, tt.grid)

			a, err := g.Locate(tt.pos)
			assertErr(t, err, tt.err, tt.causes...)
			assert.Equal(t, measure.Address{}, a)

			b, err := g.BarAt(tt.pos)
			assertErr(t, err, tt.err, tt.causes...)
			assert.Equal(t, measure.Bar{}, b)

			r, err := g.Remaining(tt.pos)
			assertErr(t, err, tt.err, tt.causes...)
			assert.False(t, r.IsSet())

			on, err := g.OnBarline(tt.pos)
			if tt.onBar == nil {
				require.NoError(t, err)
				assert.Equal(t, tt.onBarRes, on)

				return
			}

			assertErr(t, err, tt.onBar, tt.causes...)
			assert.False(t, on)
		})
	}
}

// TestLocatePositionOf covers О-10, ИС-7, П-9, КП-4, КП-5, КП-41, and
// КП-42: the position of an address; a missing bar is Е-4 and an offset at
// or after the length of the bar is Е-5.
func TestLocatePositionOf(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		grid    string
		address measure.Address
		want    string
		err     error
	}{
		{"КП-4 (0, 0)", "3/4;pickup:1/4", ad(0, "0"), "0", nil},
		{"КП-4 (1, 0)", "3/4;pickup:1/4", ad(1, "0"), "1/4", nil},
		{"КП-4 (2, 0)", "3/4;pickup:1/4", ad(2, "0"), "1", nil},
		{"КП-4 (2, 1/8)", "3/4;pickup:1/4", ad(2, "1/8"), "9/8", nil},
		{"КП-4 (1, 1/2)", "3/4;pickup:1/4", ad(1, "1/2"), "3/4", nil},
		{"КП-5 (3, 1/4)", "4/4;3:3/4", ad(3, "1/4"), "9/4", nil},
		{"extended bar", "4/4;2:length:3/2", ad(2, "5/4"), "9/4", nil},
		{"КП-41 (0, 0) without pickup", "4/4", ad(0, "0"), "", measure.ErrOutsideGrid},
		{"after N", "4/4;bars:4", ad(5, "0"), "", measure.ErrOutsideGrid},
		{"missing bar before a long offset", "4/4;bars:4", ad(5, "7"), "", measure.ErrOutsideGrid},
		{"КП-42 4/4 (1, 1)", "4/4", ad(1, "1"), "", measure.ErrOutsideBar},
		{"КП-42 3/4 (1, 3/4)", "3/4", ad(1, "3/4"), "", measure.ErrOutsideBar},
		{"pickup (0, 1/4)", "3/4;pickup:1/4", ad(0, "1/4"), "", measure.ErrOutsideBar},
		{"explicit length (3, 1/2)", "4/4;3:length:1/2", ad(3, "1/2"), "", measure.ErrOutsideBar},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			p, err := gridOf(t, tt.grid).PositionOf(tt.address)

			if tt.err != nil {
				assertErr(t, err, tt.err)
				assert.False(t, p.IsSet())

				return
			}

			require.NoError(t, err)
			assertText(t, canonicalRat(tt.want), p)
		})
	}
}

// TestLocateContains covers О-10, ИС-7, and КП-64: whether an address is
// valid in a grid is a result, not a failure.
func TestLocateContains(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		grid    string
		address measure.Address
		want    bool
	}{
		{"КП-64 (1, 1/2)", "4/4;bars:4", ad(1, "1/2"), true},
		{"КП-64 (1, 1)", "4/4;bars:4", ad(1, "1"), false},
		{"КП-64 (5, 0)", "4/4;bars:4", ad(5, "0"), false},
		{"КП-64 (0, 0)", "4/4;bars:4", ad(0, "0"), false},
		{"pickup (0, 0)", "4/4;pickup:1/4", ad(0, "0"), true},
		{"pickup (0, 1/4)", "4/4;pickup:1/4", ad(0, "1/4"), false},
		{"open max int", "4/4", ad(math.MaxInt, "1/2"), true},
		{"N (4, 7/8)", "4/4;bars:4", ad(4, "7/8"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			got, err := gridOf(t, tt.grid).Contains(tt.address)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestLocateNormalize covers О-10, П-9, КП-12, and КП-41: an address of an
// existing bar with any offset becomes the valid address of the same
// position.
func TestLocateNormalize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		grid    string
		address measure.Address
		want    measure.Address
		err     error
	}{
		{"КП-12 (2, 5/4)", "4/4", ad(2, "5/4"), ad(3, "1/4"), nil},
		{"КП-12 (1, 0)", "4/4", ad(1, "0"), ad(1, "0"), nil},
		{"КП-12 (0, 1/4)", "3/4;pickup:1/4", ad(0, "1/4"), ad(1, "0"), nil},
		{"КП-12 (0, 1)", "3/4;pickup:1/4", ad(0, "1"), ad(2, "0"), nil},
		{"across a change", "4/4;3:3/4", ad(2, "7/4"), ad(4, "0"), nil},
		{"across an explicit length", "4/4;3:length:1/2", ad(3, "5/8"), ad(4, "1/8"), nil},
		{"valid stays", "4/4;bars:4", ad(4, "7/8"), ad(4, "7/8"), nil},
		{"into the last bar", "4/4;bars:4", ad(2, "15/8"), ad(3, "7/8"), nil},
		{"КП-41 (4, 1) at the end", "4/4;bars:4", ad(4, "1"), measure.Address{}, measure.ErrOutsideGrid},
		{"bar 0 without pickup", "4/4", ad(0, "0"), measure.Address{}, measure.ErrOutsideGrid},
		{"bar after N", "4/4;bars:4", ad(5, "0"), measure.Address{}, measure.ErrOutsideGrid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			got, err := gridOf(t, tt.grid).Normalize(tt.address)

			if tt.err != nil {
				assertErr(t, err, tt.err)
				assert.Equal(t, measure.Address{}, got)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
