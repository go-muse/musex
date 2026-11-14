package tempo_test

import (
	"math/big"
	"slices"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempo/term"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestApprox covers С-19 and architecture 4.10: a separate kind, reading,
// equality and order, another unit, the explicit exact tempo, and words.
func TestApprox(t *testing.T) {
	t.Parallel()

	a, err := tempo.NewApprox(q, fi(108))
	require.NoError(t, err)
	assert.True(t, a.IsSet())
	assert.Equal(t, q, a.MustUnit())
	assertDur(t, dur(1, 4), a.MustUnitDuration())
	assert.Equal(t, fi(108), a.MustBPM())
	assertFrac(t, fi(27), a.MustSpeed())
	assertFrac(t, fi(108), a.MustQuartersPerMinute())

	got, err := a.Tempo()
	require.NoError(t, err)
	assert.Equal(t, tq(108), got, "С-19: the named number as an exact tempo")

	assert.Equal(t, ap(e, fi(180)), ap(dq, fi(60)).MustWithUnit(e), "С-19: ♩. ≈ 60 -> ♪ ≈ 180")
	assert.Equal(t, ap(q, fi(60)), ap(q, fraction.MustNew(120, 2)), "the number is stored reduced")
	assert.Equal(t, fi(60), ap(q, fraction.MustNew(120, 2)).MustBPM())

	assert.True(t, ap(dq, fi(60)).MustEqualSpeed(ap(e, fi(180))))
	assert.False(t, ap(dq, fi(60)).Equal(ap(e, fi(180))))
	assert.Equal(t, 1, ap(q, fi(120)).MustCompareSpeed(ap(dq, fi(60))))
	assert.Equal(t, 0, ap(q, fi(90)).MustCompareSpeed(ap(dq, fi(60))))

	sorted := []tempo.Approx{{}, ap(e, fi(180)), ap(q, fi(90)), ap(dq, fi(60)), ap(q, fi(120))}
	for i, x := range sorted {
		for j, y := range sorted {
			want := 0
			if i < j {
				want = -1
			} else if i > j {
				want = 1
			}

			assert.Equal(t, want, x.Compare(y), "%s vs %s", x, y)
			assert.Equal(t, x == y, x.Equal(y))
		}
	}

	shuffled := slices.Clone(sorted)
	slices.Reverse(shuffled)
	slices.SortFunc(shuffled, tempo.Approx.Compare)
	assert.Equal(t, sorted, shuffled)

	words := ap(dq, fi(60)).MustTermsIn(term.EnglishWikipedia())
	assert.Equal(t, []string{"Largo", "Larghetto", "Lento", "Adagio", "Adagietto", "Andante"}, termNames(t, words))

	m := map[tempo.Approx]int{ap(q, fi(60)): 1}
	m[ap(q, fraction.MustNew(120, 2))]++
	assert.Equal(t, 2, m[ap(q, fi(60))])
}

// TestApproxErrors covers С-30 and architecture 4.10: as New.
func TestApproxErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		call func() (tempo.Approx, error)
		want errClass
	}{
		{"unset unit", func() (tempo.Approx, error) { return tempo.NewApprox(duration.Written{}, fi(108)) }, argClass},
		{"unset number", func() (tempo.Approx, error) { return tempo.NewApprox(q, fraction.Fraction{}) }, argClass},
		{"С-30 zero", func() (tempo.Approx, error) { return tempo.NewApprox(q, fi(0)) }, noTempoClass},
		{"С-30 negative", func() (tempo.Approx, error) { return tempo.NewApprox(q, fi(-108)) }, noTempoClass},
		{"order: unit before number", func() (tempo.Approx, error) { return tempo.NewApprox(duration.Written{}, fi(0)) }, argClass},
		{"С-43 WithUnit", func() (tempo.Approx, error) { return ap(q, maxF()).WithUnit(e) }, fracRange},
		{"WithUnit unset unit", func() (tempo.Approx, error) { return ap(q, maxF()).WithUnit(duration.Written{}) }, argClass},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.call()
			tt.want.check(t, err)
			assert.Equal(t, tempo.Approx{}, got)
		})
	}

	_, err := tm(breve, maxF()).Speed()
	fracRange.check(t, err)

	_, err = ap(breve, maxF()).Speed()
	fracRange.check(t, err)

	_, err = ap(h, maxF()).QuartersPerMinute()
	fracRange.check(t, err)

	words, err := ap(q, fi(60)).TermsIn(term.List{})
	argClass.check(t, err)
	assert.Nil(t, words)
}

// TestRange covers С-20, С-25 and architecture 4.11: bounds, middle without
// rounding, another unit, containing a number and a tempo, equality and
// order.
func TestRange(t *testing.T) {
	t.Parallel()

	r, err := tempo.NewRange(e, fi(132), fi(144))
	require.NoError(t, err)
	assert.True(t, r.IsSet())
	assert.Equal(t, e, r.MustUnit())
	assertDur(t, dur(1, 8), r.MustUnitDuration())
	assert.Equal(t, fi(132), r.MustLower())
	assert.Equal(t, fi(144), r.MustUpper())
	assert.Equal(t, tm(e, fi(132)), r.MustLowerTempo())
	assert.Equal(t, tm(e, fi(144)), r.MustUpperTempo())

	mid, err := r.MidTempo()
	require.NoError(t, err)
	assert.Equal(t, tm(e, fi(138)), mid, "С-20")
	assert.Equal(t, tq(69), mid.MustWithUnit(q), "С-20 in quarters")

	assert.Equal(t, tq(43), rg(q, fi(40), fi(46)).MustMidTempo(), "С-20")
	assert.Equal(t, tm(q, fr(251, 2)), rg(q, fi(120), fi(131)).MustMidTempo(), "С-20")
	assert.Equal(t, tq(126), rg(q, fi(120), fi(131)).MustMidTempo().MustQuantize(fi(1), halfEven), "С-20")

	assert.Equal(t, rg(q, fi(66), fi(72)), r.MustWithUnit(q), "С-20 another unit")
	assert.Equal(t, rg(q, fi(60), fi(70)), rg(q, fraction.MustNew(120, 2), fraction.MustNew(140, 2)), "stored reduced")
	assert.Equal(t, fi(60), rg(q, fraction.MustNew(120, 2), fi(70)).MustLower())

	t.Run("С-20 contains a number", func(t *testing.T) {
		t.Parallel()

		for b, want := range map[int]bool{138: true, 132: true, 144: true, 131: false, 145: false, 1: false} {
			got, err := r.Contains(fi(b))
			require.NoError(t, err)
			assert.Equal(t, want, got, "%d", b)
		}

		assert.True(t, r.MustContains(fr(265, 2)))
		assert.False(t, r.MustContains(fr(263, 2)))
	})

	t.Run("С-20 contains a tempo by speed", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			t    tempo.Tempo
			want bool
		}{
			{tq(69), true}, {tq(65), false}, {tq(66), true}, {tq(72), true}, {tq(73), false},
			{tm(e, fi(132)), true}, {tm(dq, fi(44)), true}, {tm(dq, fi(48)), true}, {tm(dq, fi(49)), false},
		}

		for _, tt := range tests {
			got, err := r.ContainsTempo(tt.t)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got, "%s", tt.t)
		}
	})

	t.Run("equality and order", func(t *testing.T) {
		t.Parallel()

		assert.True(t, r.MustEqualSpeed(rg(q, fi(66), fi(72))))
		assert.False(t, r.MustEqualSpeed(rg(q, fi(66), fi(73))))
		assert.False(t, r.Equal(rg(q, fi(66), fi(72))))

		sorted := []tempo.Range{
			{}, rg(e, fi(120), fi(140)), rg(q, fi(60), fi(70)), rg(q, fi(60), fi(80)), rg(q, fi(61), fi(62)),
		}

		for i, x := range sorted {
			for j, y := range sorted {
				want := 0
				if i < j {
					want = -1
				} else if i > j {
					want = 1
				}

				assert.Equal(t, want, x.Compare(y), "%s vs %s", x, y)
				assert.Equal(t, x == y, x.Equal(y))
			}
		}

		m := map[tempo.Range]int{rg(q, fi(60), fi(70)): 1}
		m[rg(q, fraction.MustNew(120, 2), fi(70))]++
		assert.Equal(t, 2, m[rg(q, fi(60), fi(70))])
	})

	t.Run("С-25 range of a guide", func(t *testing.T) {
		t.Parallel()

		allegro := term.EnglishWikipedia().MustParse("Allegro")

		got, err := tempo.RangeOf(allegro, q)
		require.NoError(t, err)
		assert.Equal(t, rg(q, fi(120), fi(156)), got)
		assert.Equal(t, tq(138), got.MustMidTempo())

		assert.Equal(t, rg(e, fi(120), fi(144)), tempo.MustRangeOf(term.RussianWikipedia().MustParse("allegro"), e))
		assert.Equal(t, rg(q, fi(80), fi(100)), tempo.MustRangeOf(term.ElectronicMetronome().MustParse("andante"), q))
	})
}

// TestRangeErrors covers С-38, С-40, С-44 and architecture 4.3, 4.11: Е-2,
// Е-3, Е-7, Е-11 and the order of checks of NewRange, RangeOf, Contains.
func TestRangeErrors(t *testing.T) {
	t.Parallel()

	mine := term.MustNewList(
		term.Row{Entries: []term.Entry{{Name: "Ровно"}}, Guide: new(term.MustBetween(fi(100), fi(100)))},
		term.Row{Entries: []term.Entry{{Name: "Без ориентира"}}},
	)
	en := term.EnglishWikipedia()

	tests := []struct {
		name string
		call func() (tempo.Range, error)
		want errClass
	}{
		{"С-38 132-120", func() (tempo.Range, error) { return tempo.NewRange(q, fi(132), fi(120)) }, argClass},
		{"С-38 120-120", func() (tempo.Range, error) { return tempo.NewRange(q, fi(120), fi(120)) }, argClass},
		{"С-30 lower 0", func() (tempo.Range, error) { return tempo.NewRange(q, fi(0), fi(120)) }, noTempoClass},
		{"С-30 both negative", func() (tempo.Range, error) { return tempo.NewRange(q, fi(-5), fi(-3)) }, noTempoClass},
		{"unset unit", func() (tempo.Range, error) { return tempo.NewRange(duration.Written{}, fi(1), fi(2)) }, argClass},
		{"unset lower", func() (tempo.Range, error) { return tempo.NewRange(q, fraction.Fraction{}, fi(2)) }, argClass},
		{"unset upper", func() (tempo.Range, error) { return tempo.NewRange(q, fi(1), fraction.Fraction{}) }, argClass},
		{"order: unit before Е-3", func() (tempo.Range, error) { return tempo.NewRange(duration.Written{}, fi(0), fi(-1)) }, argClass},
		{"order: unset upper before lower Е-3", func() (tempo.Range, error) { return tempo.NewRange(q, fi(0), fraction.Fraction{}) }, argClass},
		{"order: upper Е-3 before lower >= upper", func() (tempo.Range, error) { return tempo.NewRange(q, fi(5), fi(0)) }, noTempoClass},
		{"С-40 Larghissimo", func() (tempo.Range, error) { return tempo.RangeOf(en.MustParse("Larghissimo"), q) }, noRange},
		{"С-40 Prestissimo", func() (tempo.Range, error) { return tempo.RangeOf(en.MustParse("Prestissimo"), q) }, noRange},
		{"open bound of list 3", func() (tempo.Range, error) {
			return tempo.RangeOf(term.JapaneseMetronome().MustParse("Prestissimo"), q)
		}, noRange},
		{"equal bounds of list 6", func() (tempo.Range, error) {
			return tempo.RangeOf(term.MaelzelMetronome().MustParse("largo"), q)
		}, noRange},
		{"С-40 Ровно", func() (tempo.Range, error) { return tempo.RangeOf(mine.MustParse("Ровно"), q) }, noRange},
		{"С-40 no guide", func() (tempo.Range, error) { return tempo.RangeOf(mine.MustParse("Без ориентира"), q) }, noRange},
		{"RangeOf unset term", func() (tempo.Range, error) { return tempo.RangeOf(term.Term{}, q) }, argClass},
		{"RangeOf unset unit", func() (tempo.Range, error) { return tempo.RangeOf(en.MustParse("Allegro"), duration.Written{}) }, argClass},
		{"order: unit before no range", func() (tempo.Range, error) {
			return tempo.RangeOf(en.MustParse("Larghissimo"), duration.Written{})
		}, argClass},
		{"С-43 WithUnit", func() (tempo.Range, error) { return rg(q, fi(1), maxF()).WithUnit(e) }, fracRange},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.call()
			tt.want.check(t, err)
			assert.Equal(t, tempo.Range{}, got)
		})
	}

	r := rg(e, fi(132), fi(144))

	for _, b := range []fraction.Fraction{fi(0), fi(-5)} {
		got, err := r.Contains(b)
		noTempoClass.check(t, err)
		assert.False(t, got)
	}

	_, err := r.Contains(fraction.Fraction{})
	argClass.check(t, err)

	mid, err := rg(q, invMaxF(), fbig(big.NewInt(2), fraction.MaxComponentBig())).MidTempo()
	fracRange.check(t, err)
	assert.Equal(t, tempo.Tempo{}, mid)
}

// termNames returns the names of terms, nil for nil.
func termNames(t *testing.T, terms []term.Term) []string {
	t.Helper()

	if terms == nil {
		return nil
	}

	out := make([]string, 0, len(terms))
	for _, w := range terms {
		out = append(out, w.MustName())
	}

	return out
}
