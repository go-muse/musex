package tempo_test

import (
	"testing"

	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempo/term"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestKind covers architecture 4.13: Kind values, validity, equality, text.
func TestKind(t *testing.T) {
	t.Parallel()

	assert.Equal(t, tempo.NoMark, tempo.Kind(0))

	tests := []struct {
		k     tempo.Kind
		valid bool
		text  string
	}{
		{tempo.NoMark, true, "none"},
		{tempo.ExactMark, true, "exact"},
		{tempo.ApproxMark, true, "approx"},
		{tempo.RangeMark, true, "range"},
		{tempo.Kind(4), false, "Kind(invalid)"},
		{tempo.Kind(255), false, "Kind(invalid)"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.valid, tt.k.IsValid(), tt.text)
		assert.Equal(t, tt.text, tt.k.String())

		for _, o := range []tempo.Kind{tempo.NoMark, tempo.ExactMark, tempo.ApproxMark, tempo.RangeMark, tempo.Kind(4)} {
			assert.Equal(t, tt.k == o, tt.k.Equal(o))
		}
	}
}

// TestIndication covers С-28 and architecture 4.13: a term, a marking of
// any kind, or both; reading; replacing; the number is not checked against
// the guide of the term.
func TestIndication(t *testing.T) {
	t.Parallel()

	en := term.EnglishWikipedia()
	allegro := en.MustParse("Allegro")
	moderato := en.MustParse("Moderato")

	t.Run("С-28 Allegro, ♩ = 120", func(t *testing.T) {
		t.Parallel()

		i, err := tempo.IndicationOf(allegro)
		require.NoError(t, err)
		i, err = i.WithTempo(tq(120))
		require.NoError(t, err)

		assert.True(t, i.IsSet())

		w, ok, err := i.Term()
		require.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, allegro, w)
		assert.Equal(t, tempo.ExactMark, i.MustKind())

		m, ok, err := i.Tempo()
		require.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, tq(120), m)

		a, ok, err := i.Approx()
		require.NoError(t, err)
		assert.False(t, ok)
		assert.Equal(t, tempo.Approx{}, a)

		r, ok, err := i.Range()
		require.NoError(t, err)
		assert.False(t, ok)
		assert.Equal(t, tempo.Range{}, r)

		assert.Equal(t, "Allegro, 4=120", i.String())
	})

	t.Run("С-28 Moderato, ♩ ≈ 108", func(t *testing.T) {
		t.Parallel()

		i := tempo.MustIndicationOf(moderato).MustWithApprox(ap(q, fi(108)))
		assert.Equal(t, tempo.ApproxMark, i.MustKind())

		_, ok := i.MustTempo()
		assert.False(t, ok, "no exact tempo")

		a, ok := i.MustApprox()
		assert.True(t, ok)
		assert.Equal(t, ap(q, fi(108)), a)
		assert.Equal(t, "Moderato, 4≈108", i.String())
	})

	t.Run("С-28 only a term", func(t *testing.T) {
		t.Parallel()

		i := tempo.MustIndicationOf(en.MustParse("Andante"))
		assert.Equal(t, tempo.NoMark, i.MustKind())

		_, ok := i.MustTempo()
		assert.False(t, ok)
		_, ok = i.MustApprox()
		assert.False(t, ok)
		_, ok = i.MustRange()
		assert.False(t, ok)
		assert.Equal(t, "Andante", i.String())
	})

	t.Run("С-28 only a tempo", func(t *testing.T) {
		t.Parallel()

		i, err := tempo.IndicationOfTempo(tq(60))
		require.NoError(t, err)

		w, ok, err := i.Term()
		require.NoError(t, err)
		assert.False(t, ok)
		assert.Equal(t, term.Term{}, w)
		assert.Equal(t, tempo.ExactMark, i.MustKind())
		assert.Equal(t, "4=60", i.String())
	})

	t.Run("С-28 the number is not checked", func(t *testing.T) {
		t.Parallel()

		i, err := tempo.MustIndicationOf(allegro).WithTempo(tq(60))
		require.NoError(t, err)
		assert.Equal(t, "Allegro, 4=60", i.String())
	})

	t.Run("range and approx only", func(t *testing.T) {
		t.Parallel()

		i, err := tempo.IndicationOfRange(rg(e, fi(132), fi(144)))
		require.NoError(t, err)
		assert.Equal(t, tempo.RangeMark, i.MustKind())

		r, ok := i.MustRange()
		assert.True(t, ok)
		assert.Equal(t, rg(e, fi(132), fi(144)), r)
		assert.Equal(t, "8=132-144", i.String())
		assert.Equal(t, "Allegro, 8=132-144", i.MustWithTerm(allegro).String())

		a, err := tempo.IndicationOfApprox(ap(dq, fi(60)))
		require.NoError(t, err)
		assert.Equal(t, "4.≈60", a.String())
	})

	t.Run("replace and remove", func(t *testing.T) {
		t.Parallel()

		i := tempo.MustIndicationOfApprox(ap(q, fi(108)))
		i = i.MustWithTempo(tq(100))
		assert.Equal(t, tempo.ExactMark, i.MustKind())
		_, ok := i.MustApprox()
		assert.False(t, ok, "the marking is replaced")

		i = i.MustWithRange(rg(q, fi(60), fi(70)))
		assert.Equal(t, tempo.RangeMark, i.MustKind())
		_, ok = i.MustTempo()
		assert.False(t, ok)

		i = i.MustWithTerm(allegro).MustWithTerm(moderato)
		w, _ := i.MustTerm()
		assert.Equal(t, moderato, w, "the term is replaced")

		noTerm := i.MustWithoutTerm()
		_, ok = noTerm.MustTerm()
		assert.False(t, ok)
		assert.Equal(t, tempo.MustIndicationOfRange(rg(q, fi(60), fi(70))), noTerm)

		noMark := i.MustWithoutMarking()
		assert.Equal(t, tempo.NoMark, noMark.MustKind())
		assert.Equal(t, tempo.MustIndicationOf(moderato), noMark)
	})

	t.Run("remove what is absent", func(t *testing.T) {
		t.Parallel()

		// 4.13: Е-2 only if neither a term nor a marking would remain, so
		// removing the absent part returns the same value without an error.
		onlyMark := tempo.MustIndicationOfTempo(tq(60))
		got, err := onlyMark.WithoutTerm()
		require.NoError(t, err)
		assert.Equal(t, onlyMark, got)

		onlyTerm := tempo.MustIndicationOf(allegro)
		got, err = onlyTerm.WithoutMarking()
		require.NoError(t, err)
		assert.Equal(t, onlyTerm, got)
	})

	t.Run("equality", func(t *testing.T) {
		t.Parallel()

		values := []tempo.Indication{
			{},
			tempo.MustIndicationOf(allegro),
			tempo.MustIndicationOf(term.RussianWikipedia().MustParse("allegro")),
			tempo.MustIndicationOfTempo(tq(120)),
			tempo.MustIndicationOfApprox(ap(q, fi(120))),
			tempo.MustIndicationOfRange(rg(q, fi(120), fi(130))),
			tempo.MustIndicationOf(allegro).MustWithTempo(tq(120)),
			tempo.MustIndicationOf(allegro).MustWithApprox(ap(q, fi(120))),
		}

		for _, a := range values {
			for _, b := range values {
				assert.Equal(t, a == b, a.Equal(b), "%q vs %q", a, b)
			}
		}

		same := tempo.MustIndicationOf(term.EnglishWikipedia().MustParse("ALLEGRO")).MustWithTempo(tempo.MustNewQuarter(fi(120)))
		assert.Equal(t, values[6], same)
		assert.True(t, values[6].Equal(same))

		m := map[tempo.Indication]int{values[6]: 1}
		m[same]++
		assert.Equal(t, 2, m[values[6]])
	})
}

// TestIndicationErrors covers С-46 and architecture 4.3, 4.13: unset term —
// Е-2; unset marking — Е-1; nothing left — Е-2; and the order of checks.
func TestIndicationErrors(t *testing.T) {
	t.Parallel()

	allegro := term.EnglishWikipedia().MustParse("Allegro")
	onlyTerm := tempo.MustIndicationOf(allegro)
	onlyMark := tempo.MustIndicationOfTempo(tq(60))

	tests := []struct {
		name string
		call func() (tempo.Indication, error)
		want errClass
	}{
		{"IndicationOf unset term", func() (tempo.Indication, error) { return tempo.IndicationOf(term.Term{}) }, argClass},
		{"IndicationOfTempo unset", func() (tempo.Indication, error) { return tempo.IndicationOfTempo(tempo.Tempo{}) }, unsetClass},
		{"IndicationOfApprox unset", func() (tempo.Indication, error) { return tempo.IndicationOfApprox(tempo.Approx{}) }, unsetClass},
		{"IndicationOfRange unset", func() (tempo.Indication, error) { return tempo.IndicationOfRange(tempo.Range{}) }, unsetClass},
		{"WithTerm unset term", func() (tempo.Indication, error) { return onlyMark.WithTerm(term.Term{}) }, argClass},
		{"WithTempo unset", func() (tempo.Indication, error) { return onlyTerm.WithTempo(tempo.Tempo{}) }, unsetClass},
		{"WithApprox unset", func() (tempo.Indication, error) { return onlyTerm.WithApprox(tempo.Approx{}) }, unsetClass},
		{"WithRange unset", func() (tempo.Indication, error) { return onlyTerm.WithRange(tempo.Range{}) }, unsetClass},
		{"С-46 WithoutTerm of a term", func() (tempo.Indication, error) { return onlyTerm.WithoutTerm() }, argClass},
		{"С-46 WithoutMarking of a marking", func() (tempo.Indication, error) { return onlyMark.WithoutMarking() }, argClass},
		{"order: unset receiver before unset term",
			func() (tempo.Indication, error) { return tempo.Indication{}.WithTerm(term.Term{}) }, unsetClass},
		{"unset receiver of WithTempo", func() (tempo.Indication, error) { return tempo.Indication{}.WithTempo(tq(1)) }, unsetClass},
		{"order: unset receiver and unset argument",
			func() (tempo.Indication, error) { return tempo.Indication{}.WithTempo(tempo.Tempo{}) }, unsetClass},
		{"unset receiver of WithoutTerm", func() (tempo.Indication, error) { return tempo.Indication{}.WithoutTerm() }, unsetClass},
		{"unset receiver of WithoutMarking",
			func() (tempo.Indication, error) { return tempo.Indication{}.WithoutMarking() }, unsetClass},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.call()
			tt.want.check(t, err)
			assert.Equal(t, tempo.Indication{}, got)
		})
	}
}
