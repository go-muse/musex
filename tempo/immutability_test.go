package tempo_test

import (
	"sync"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempo/term"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInputsAreKept covers С-58 and architecture 3.2: no operation changes
// its receiver or arguments, on success, on failure, or in a Must panic.
func TestInputsAreKept(t *testing.T) {
	t.Parallel()

	x := tm(dq, fr(200, 3))
	a := ap(q, fi(108))
	r := rg(e, fi(132), fi(144))
	s := sec(24, 37)
	en := term.EnglishWikipedia()
	i := tempo.MustIndicationOf(en.MustParse("Allegro")).MustWithTempo(tq(120))
	bpm := fr(185, 2)
	d := dur(3, 8)

	before := []any{x, a, r, s, en, i, bpm, d}

	_, _ = x.WithUnit(e)
	_, _ = x.Mul(fi(2))
	_, _ = x.Mul(fi(-1))
	_, _ = x.Quantize(fi(1), halfUp)
	_, _ = x.Quantize(fi(1), roundingMode(0))
	_, _ = x.SnapToScale(halfUp)
	_, _ = x.Modulate(q, d)
	_, _ = x.Seconds(d)
	_, _ = x.WithBPM(bpm)
	_, _ = x.TermsIn(en)
	_, _ = a.WithUnit(e)
	_, _ = a.Tempo()
	_, _ = r.WithUnit(q)
	_, _ = r.MidTempo()
	_, _ = r.Contains(fi(-1))
	_, _ = s.Add(s)
	_, _ = s.Sub(sec(1, 1))
	_, _ = s.TimeDuration(halfUp)
	_, _ = i.WithoutTerm()
	_, _ = i.WithTerm(en.MustParse("Largo"))
	_, _ = i.WithApprox(a)
	_, _ = tempo.New(q, bpm)
	_, _ = tempo.NearestScaleDivision(bpm, halfUp)
	_, _ = tempo.ParseIndication("Allegro, 4=60", en)
	_, _ = capturePanic(func() { x.MustMul(fi(0)) })
	_, _ = capturePanic(func() { s.MustSub(sec(1, 1)) })
	_, _ = capturePanic(func() { i.MustWithoutMarking().MustWithoutTerm() })

	assert.Equal(t, before, []any{x, a, r, s, en, i, bpm, d})
}

// TestResultSlicesBelongToCaller covers architecture 3.2: ScaleDivisions and
// TermsIn return new slices.
func TestResultSlicesBelongToCaller(t *testing.T) {
	t.Parallel()

	got := tq(130).MustTermsIn(term.EnglishWikipedia())
	require.Len(t, got, 3)
	got[0] = term.Term{}
	assert.Equal(t, term.EnglishWikipedia().MustParse("Allegro"), tq(130).MustTermsIn(term.EnglishWikipedia())[0])

	got = ap(q, fi(130)).MustTermsIn(term.EnglishWikipedia())
	got[0] = term.Term{}
	assert.Equal(t, term.EnglishWikipedia().MustParse("Allegro"), ap(q, fi(130)).MustTermsIn(term.EnglishWikipedia())[0])

	assert.Nil(t, tq(30).MustTermsIn(term.RussianWikipedia()), "an empty answer is nil")
}

// TestMarshalTextNewSlice covers architecture 3.2: MarshalText returns a new
// slice; AppendText does not change dst[:len(dst)].
func TestMarshalTextNewSlice(t *testing.T) {
	t.Parallel()

	x := tq(60)

	b, err := x.MarshalText()
	require.NoError(t, err)

	b[0] = '8'
	assert.Equal(t, "4=60", x.String())

	b2, err := x.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, "4=60", string(b2))

	buf := make([]byte, 1, 32)
	buf[0] = 'X'
	out, err := sec(1, 3).AppendText(buf)
	require.NoError(t, err)
	assert.Equal(t, "X1/3", string(out))
	assert.Equal(t, "X", string(buf))
}

// TestConcurrentReads covers architecture 3.2, "Конкурентность": shared
// values, built-in lists and the scale are used from many goroutines at once
// without synchronization; run with -race.
func TestConcurrentReads(t *testing.T) {
	t.Parallel()

	shared := tm(dq, fr(200, 3))
	r := rg(e, fi(132), fi(144))
	l := customList()

	const workers = 16

	type outcome struct {
		texts []string
		err   error
	}

	results := make(chan outcome, workers)

	var wg sync.WaitGroup

	for range workers {
		wg.Go(func() {
			var o outcome

			note := func(s string, err error) {
				if err != nil && o.err == nil {
					o.err = err
				}

				o.texts = append(o.texts, s)
			}

			s, err := shared.Seconds(dur(1, 1))
			note(s.String(), err)

			ns, err := s.TimeDuration(halfEven)
			note(ns.String(), err)

			q2, err := shared.Quantize(fi(1), halfUp)
			note(q2.String(), err)

			mid, err := r.MidTempo()
			note(mid.String(), err)

			n, err := tempo.NearestScaleDivision(fi(110), halfEven)
			note(fraction.FromInt(n).String(), err)

			ws, err := shared.TermsIn(term.EnglishWikipedia())
			note(termNamesNoT(ws), err)

			ind, err := tempo.ParseIndication(`"4=60", 4=120`, l)
			note(ind.String(), err)

			results <- o
		})
	}

	wg.Wait()
	close(results)

	want := []string{
		"12/5", "2.4s", "4.=67", "8=138", "112/1",
		"Adagietto;Andante;Marcia moderato;", `"4=60", 4=120`,
	}

	for o := range results {
		require.NoError(t, o.err)
		assert.Equal(t, want, o.texts)
	}
}

// termNamesNoT joins the names of terms with ";" after each name.
func termNamesNoT(terms []term.Term) string {
	var out string
	for _, w := range terms {
		out += w.String() + ";"
	}

	return out
}
