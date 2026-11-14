package interval_test

import (
	"errors"
	"iter"
	"slices"
	"testing"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
)

// T-7, Must variants (OP-45; architecture 4.2, "Must"): the same results on
// success; a panic with the same error value otherwise.

// mustCase pairs a function with its Must variant on one input. Both return
// their results as a slice; base also returns the error.
type mustCase struct {
	name string
	base func() ([]any, error)
	must func() []any
}

// normalize turns sequences into slices so results can be compared.
func normalize(v any) any {
	switch s := v.(type) {
	case iter.Seq[interval.Interval]:
		if s == nil {
			return []interval.Interval(nil)
		}

		return slices.Collect(s)
	case iter.Seq[interval.Respelling]:
		if s == nil {
			return []interval.Respelling(nil)
		}

		return slices.Collect(s)
	default:
		return v
	}
}

func checkMust(t *testing.T, c mustCase) {
	t.Helper()

	want, err := c.base()
	if err == nil {
		var got []any

		noPanic(t, func() { got = c.must() })

		if assert.Len(t, got, len(want), c.name) {
			for k := range want {
				assert.Equal(t, normalize(want[k]), normalize(got[k]), "%s: result %d", c.name, k)
			}
		}

		return
	}

	perr := panicError(t, func() { c.must() })
	if perr == nil {
		return
	}

	assert.Equal(t, err.Error(), perr.Error(), "%s: the panic value is the error", c.name)

	for _, s := range sentinels {
		assert.Equal(t, errors.Is(err, s.err), errors.Is(perr, s.err), "%s: %s", c.name, s.name)
	}
}

func r1[T any](v T, err error) ([]any, error)         { return []any{v}, err }
func r2[T, U any](a T, b U, err error) ([]any, error) { return []any{a, b}, err }
func m1[T any](v T) []any                             { return []any{v} }
func m2[T, U any](a T, b U) []any                     { return []any{a, b} }

// mustCases lists every Must variant of 4.1 with a successful and a failing
// input.
func mustCases(t *testing.T) []mustCase {
	t.Helper()

	var (
		u        interval.Interval
		m3       = iv("M3")
		a8       = iv("A8")
		tr       = trad()
		irr      = irregular()
		c, e     = spelling.C(), spelling.E()
		c4, e4   = pt(spelling.C(), 4), pt(spelling.E(), 4)
		src      = std()
		nc4      = noteAt(t, c4, 10, src)
		ne4      = noteAt(t, e4, 0, src)
		ne       = noteName(t, spelling.E(), 0, src)
		free4    = noteAt(t, c4, 0, nil)
		eh       = eHalfFlat()
		twoSharp = ai(2)
	)

	return []mustCase{
		{"MustNew ok", func() ([]any, error) { return r1(interval.New(interval.Major, 3)) },
			func() []any { return m1(interval.MustNew(interval.Major, 3)) }},
		{"MustNew error", func() ([]any, error) { return r1(interval.New(interval.Perfect, 3)) },
			func() []any { return m1(interval.MustNew(interval.Perfect, 3)) }},
		{"MustNewMultiple ok", func() ([]any, error) { return r1(interval.NewMultiple(2, interval.Augmented, 4)) },
			func() []any { return m1(interval.MustNewMultiple(2, interval.Augmented, 4)) }},
		{"MustNewMultiple error", func() ([]any, error) { return r1(interval.NewMultiple(mag+1, interval.Augmented, 4)) },
			func() []any { return m1(interval.MustNewMultiple(mag+1, interval.Augmented, 4)) }},
		{"MustFromDeviation ok", func() ([]any, error) { return r1(interval.FromDeviation(3, alt(-1, 2))) },
			func() []any { return m1(interval.MustFromDeviation(3, alt(-1, 2))) }},
		{"MustFromDeviation error", func() ([]any, error) { return r1(interval.FromDeviation(1, ai(-1))) },
			func() []any { return m1(interval.MustFromDeviation(1, ai(-1))) }},
		{"MustSimple ok", func() ([]any, error) { return r1(iv("M10").Simple()) }, func() []any { return m1(iv("M10").MustSimple()) }},
		{"MustSimple error", func() ([]any, error) { return r1(u.Simple()) }, func() []any { return m1(u.MustSimple()) }},
		{"MustWithDirection ok", func() ([]any, error) { return r1(m3.WithDirection(interval.Down)) },
			func() []any { return m1(m3.MustWithDirection(interval.Down)) }},
		{"MustWithDirection error", func() ([]any, error) { return r1(m3.WithDirection(interval.DirectionUnset)) },
			func() []any { return m1(m3.MustWithDirection(interval.DirectionUnset)) }},
		{"MustUp ok", func() ([]any, error) { return r1(iv("↓m6").Up()) }, func() []any { return m1(iv("↓m6").MustUp()) }},
		{"MustUp error", func() ([]any, error) { return r1(u.Up()) }, func() []any { return m1(u.MustUp()) }},
		{"MustDown ok", func() ([]any, error) { return r1(m3.Down()) }, func() []any { return m1(m3.MustDown()) }},
		{"MustDown error", func() ([]any, error) { return r1(u.Down()) }, func() []any { return m1(u.MustDown()) }},
		{"MustOpposite ok", func() ([]any, error) { return r1(m3.Opposite()) }, func() []any { return m1(m3.MustOpposite()) }},
		{"MustOpposite error", func() ([]any, error) { return r1(u.Opposite()) }, func() []any { return m1(u.MustOpposite()) }},
		{"MustAddOctaves ok", func() ([]any, error) { return r1(m3.AddOctaves(2)) }, func() []any { return m1(m3.MustAddOctaves(2)) }},
		{"MustAddOctaves error", func() ([]any, error) { return r1(m3.AddOctaves(-1)) }, func() []any { return m1(m3.MustAddOctaves(-1)) }},
		{"MustExpand ok", func() ([]any, error) { return r1(m3.Expand()) }, func() []any { return m1(m3.MustExpand()) }},
		{"MustExpand error", func() ([]any, error) { return r1(aug(mag, 4).Expand()) }, func() []any { return m1(aug(mag, 4).MustExpand()) }},
		{"MustContract ok", func() ([]any, error) { return r1(m3.Contract()) }, func() []any { return m1(m3.MustContract()) }},
		{"MustContract error", func() ([]any, error) { return r1(iv("P1").Contract()) }, func() []any { return m1(iv("P1").MustContract()) }},
		{"MustExpandBy ok", func() ([]any, error) { return r1(m3.ExpandBy(2)) }, func() []any { return m1(m3.MustExpandBy(2)) }},
		{"MustExpandBy error", func() ([]any, error) { return r1(m3.ExpandBy(maxInt)) }, func() []any { return m1(m3.MustExpandBy(maxInt)) }},
		{"MustContractBy ok", func() ([]any, error) { return r1(m3.ContractBy(2)) }, func() []any { return m1(m3.MustContractBy(2)) }},
		{"MustContractBy error", func() ([]any, error) { return r1(m3.ContractBy(minInt)) }, func() []any { return m1(m3.MustContractBy(minInt)) }},
		{"MustExpandByAlteration ok", func() ([]any, error) { return r1(m3.ExpandByAlteration(alt(1, 2))) },
			func() []any { return m1(m3.MustExpandByAlteration(alt(1, 2))) }},
		{"MustExpandByAlteration error", func() ([]any, error) { return r1(u.ExpandByAlteration(alt(1, 2))) },
			func() []any { return m1(u.MustExpandByAlteration(alt(1, 2))) }},
		{"MustContractByAlteration ok", func() ([]any, error) { return r1(m3.ContractByAlteration(alt(1, 2))) },
			func() []any { return m1(m3.MustContractByAlteration(alt(1, 2))) }},
		{"MustContractByAlteration error", func() ([]any, error) { return r1(iv("P1").ContractByAlteration(alt(1, 2))) },
			func() []any { return m1(iv("P1").MustContractByAlteration(alt(1, 2))) }},
		{"MustInvert ok", func() ([]any, error) { return r1(m3.Invert()) }, func() []any { return m1(m3.MustInvert()) }},
		{"MustInvert error", func() ([]any, error) { return r1(a8.Invert()) }, func() []any { return m1(a8.MustInvert()) }},
		{"MustInversionClass ok", func() ([]any, error) { return r2(iv("M10").InversionClass()) },
			func() []any { return m2(iv("M10").MustInversionClass()) }},
		{"MustInversionClass error", func() ([]any, error) { return r2(a8.InversionClass()) },
			func() []any { return m2(a8.MustInversionClass()) }},
		{"MustAdd ok", func() ([]any, error) { return r1(m3.Add(iv("m3"))) }, func() []any { return m1(m3.MustAdd(iv("m3"))) }},
		{"MustAdd error", func() ([]any, error) { return r1(m3.Add(u)) }, func() []any { return m1(m3.MustAdd(u)) }},
		{"MustSub ok", func() ([]any, error) { return r1(m3.Sub(iv("P5"))) }, func() []any { return m1(m3.MustSub(iv("P5"))) }},
		{"MustSub error", func() ([]any, error) { return r1(u.Sub(m3)) }, func() []any { return m1(u.MustSub(m3)) }},
		{"MustMul ok", func() ([]any, error) { return r1(m3.Mul(-3)) }, func() []any { return m1(m3.MustMul(-3)) }},
		{"MustMul error", func() ([]any, error) { return r1(iv("P8").Mul(maxInt)) }, func() []any { return m1(iv("P8").MustMul(maxInt)) }},
		{"MustBetween ok", func() ([]any, error) { return r1(interval.Between(c, e, interval.Down)) },
			func() []any { return m1(interval.MustBetween(c, e, interval.Down)) }},
		{"MustBetween error", func() ([]any, error) { return r1(interval.Between(c, e, interval.DirectionUnset)) },
			func() []any { return m1(interval.MustBetween(c, e, interval.DirectionUnset)) }},
		{"MustBothWays ok", func() ([]any, error) { return r2(interval.BothWays(c, e)) },
			func() []any { return m2(interval.MustBothWays(c, e)) }},
		{"MustBothWays error", func() ([]any, error) { return r2(interval.BothWays(c, spelling.Spelling{})) },
			func() []any { return m2(interval.MustBothWays(c, spelling.Spelling{})) }},
		{"MustTranspose ok", func() ([]any, error) { return r1(interval.Transpose(e, m3)) },
			func() []any { return m1(interval.MustTranspose(e, m3)) }},
		{"MustTranspose error", func() ([]any, error) { return r1(interval.Transpose(e, u)) },
			func() []any { return m1(interval.MustTranspose(e, u)) }},
		{"MustFifthsOf ok", func() ([]any, error) { return r1(interval.FifthsOf(spelling.FSharp())) },
			func() []any { return m1(interval.MustFifthsOf(spelling.FSharp())) }},
		{"MustFifthsOf error", func() ([]any, error) { return r1(interval.FifthsOf(eh)) },
			func() []any { return m1(interval.MustFifthsOf(eh)) }},
		{"MustFifthsBetween ok", func() ([]any, error) { return r1(interval.FifthsBetween(c, e)) },
			func() []any { return m1(interval.MustFifthsBetween(c, e)) }},
		{"MustFifthsBetween error", func() ([]any, error) { return r1(interval.FifthsBetween(c, eh)) },
			func() []any { return m1(interval.MustFifthsBetween(c, eh)) }},
		{"MustFromFifths ok", func() ([]any, error) { return r1(interval.FromFifths(-6)) },
			func() []any { return m1(interval.MustFromFifths(-6)) }},
		{"MustFromFifths error", func() ([]any, error) { return r1(interval.FromFifths(maxInt)) },
			func() []any { return m1(interval.MustFromFifths(maxInt)) }},
		{"MustFifths ok", func() ([]any, error) { return r1(iv("↓M3").Fifths()) }, func() []any { return m1(iv("↓M3").MustFifths()) }},
		{"MustFifths error", func() ([]any, error) { return r1(iv("n3").Fifths()) }, func() []any { return m1(iv("n3").MustFifths()) }},
		{"MustReflect ok", func() ([]any, error) { return r1(interval.Reflect(spelling.D(), spelling.FSharp())) },
			func() []any { return m1(interval.MustReflect(spelling.D(), spelling.FSharp())) }},
		{"MustReflect error", func() ([]any, error) { return r1(interval.Reflect(spelling.Spelling{}, c)) },
			func() []any { return m1(interval.MustReflect(spelling.Spelling{}, c)) }},
		{"MustBetweenPitches ok", func() ([]any, error) { return r1(interval.BetweenPitches(tr, c4, e4)) },
			func() []any { return m1(interval.MustBetweenPitches(tr, c4, e4)) }},
		{"MustBetweenPitches error", func() ([]any, error) { return r1(interval.BetweenPitches(tr, c4, pt(eh, 4))) },
			func() []any { return m1(interval.MustBetweenPitches(tr, c4, pt(eh, 4))) }},
		{"MustTransposePitch ok", func() ([]any, error) { return r1(interval.TransposePitch(tr, c4, iv("M10"))) },
			func() []any { return m1(interval.MustTransposePitch(tr, c4, iv("M10"))) }},
		{"MustTransposePitch error", func() ([]any, error) { return r1(interval.TransposePitch(nil, c4, m3)) },
			func() []any { return m1(interval.MustTransposePitch(nil, c4, m3)) }},
		{"MustReflectPitch ok", func() ([]any, error) { return r1(interval.ReflectPitch(tr, pt(spelling.D(), 4), c4)) },
			func() []any { return m1(interval.MustReflectPitch(tr, pt(spelling.D(), 4), c4)) }},
		{"MustReflectPitch error", func() ([]any, error) { return r1(interval.ReflectPitch(tr, pitch.Pitch{}, c4)) },
			func() []any { return m1(interval.MustReflectPitch(tr, pitch.Pitch{}, c4)) }},
		{"MustBetweenNotes ok", func() ([]any, error) { return r1(interval.BetweenNotes(nc4, ne4, interval.DirectionUnset)) },
			func() []any { return m1(interval.MustBetweenNotes(nc4, ne4, interval.DirectionUnset)) }},
		{"MustBetweenNotes error", func() ([]any, error) { return r1(interval.BetweenNotes(nc4, ne, interval.Up)) },
			func() []any { return m1(interval.MustBetweenNotes(nc4, ne, interval.Up)) }},
		{"MustBetweenNotesIn ok", func() ([]any, error) { return r1(interval.BetweenNotesIn(tr, free4, ne4, interval.Up)) },
			func() []any { return m1(interval.MustBetweenNotesIn(tr, free4, ne4, interval.Up)) }},
		{"MustBetweenNotesIn error", func() ([]any, error) { return r1(interval.BetweenNotesIn(nil, free4, ne4, interval.Up)) },
			func() []any { return m1(interval.MustBetweenNotesIn(nil, free4, ne4, interval.Up)) }},
		{"MustTransposeNote ok", func() ([]any, error) { return r1(interval.TransposeNote(nc4, m3)) },
			func() []any { return m1(interval.MustTransposeNote(nc4, m3)) }},
		{"MustTransposeNote error", func() ([]any, error) { return r1(interval.TransposeNote(free4, m3)) },
			func() []any { return m1(interval.MustTransposeNote(free4, m3)) }},
		{"MustTransposeNoteIn ok", func() ([]any, error) { return r1(interval.TransposeNoteIn(tr, free4, m3)) },
			func() []any { return m1(interval.MustTransposeNoteIn(tr, free4, m3)) }},
		{"MustTransposeNoteIn error", func() ([]any, error) { return r1(interval.TransposeNoteIn(tr, note.Note{}, m3)) },
			func() []any { return m1(interval.MustTransposeNoteIn(tr, note.Note{}, m3)) }},
		{"MustSteps ok", func() ([]any, error) { return r1(iv("A4").Steps(tr, noStart)) },
			func() []any { return m1(iv("A4").MustSteps(tr, noStart)) }},
		{"MustSteps error", func() ([]any, error) { return r1(iv("M2").Steps(irr, noStart)) },
			func() []any { return m1(iv("M2").MustSteps(irr, noStart)) }},
		{"MustSignedSteps ok", func() ([]any, error) { return r1(iv("↓M2").SignedSteps(irr, spelling.D())) },
			func() []any { return m1(iv("↓M2").MustSignedSteps(irr, spelling.D())) }},
		{"MustSignedSteps error", func() ([]any, error) { return r1(iv("n3").SignedSteps(tr, noStart)) },
			func() []any { return m1(iv("n3").MustSignedSteps(tr, noStart)) }},
		{"MustFromSteps ok", func() ([]any, error) { return r1(interval.FromSteps(qt24(), 3, 7, noStart)) },
			func() []any { return m1(interval.MustFromSteps(qt24(), 3, 7, noStart)) }},
		{"MustFromSteps error", func() ([]any, error) { return r1(interval.FromSteps(tr, 1, -1, noStart)) },
			func() []any { return m1(interval.MustFromSteps(tr, 1, -1, noStart)) }},
		{"MustEnharmonicallyEqual ok", func() ([]any, error) {
			return r1(interval.EnharmonicallyEqual(tr, iv("A4"), noStart, iv("d5"), noStart))
		}, func() []any { return m1(interval.MustEnharmonicallyEqual(tr, iv("A4"), noStart, iv("d5"), noStart)) }},
		{"MustEnharmonicallyEqual error", func() ([]any, error) {
			return r1(interval.EnharmonicallyEqual(irr, iv("A4"), noStart, iv("d5"), noStart))
		}, func() []any { return m1(interval.MustEnharmonicallyEqual(irr, iv("A4"), noStart, iv("d5"), noStart)) }},
		{"MustAllFromSteps ok", func() ([]any, error) { return r1(interval.AllFromSteps(tr, 6, 2, noStart)) },
			func() []any { return m1(interval.MustAllFromSteps(tr, 6, 2, noStart)) }},
		{"MustAllFromSteps error", func() ([]any, error) { return r1(interval.AllFromSteps(tr, 6, -1, noStart)) },
			func() []any { return m1(interval.MustAllFromSteps(tr, 6, -1, noStart)) }},
		{"MustEnharmonics ok", func() ([]any, error) { return r1(iv("A4").Enharmonics(tr, 2, noStart)) },
			func() []any { return m1(iv("A4").MustEnharmonics(tr, 2, noStart)) }},
		{"MustEnharmonics error", func() ([]any, error) { return r1(iv("A4").Enharmonics(nil, 2, noStart)) },
			func() []any { return m1(iv("A4").MustEnharmonics(nil, 2, noStart)) }},
		{"MustRespellings ok", func() ([]any, error) { return r1(interval.Respellings(tr, c4, e4, twoSharp, interval.ReplaceBoth)) },
			func() []any { return m1(interval.MustRespellings(tr, c4, e4, twoSharp, interval.ReplaceBoth)) }},
		{"MustRespellings error", func() ([]any, error) { return r1(interval.Respellings(tr, c4, e4, twoSharp, interval.ReplaceUnset)) },
			func() []any { return m1(interval.MustRespellings(tr, c4, e4, twoSharp, interval.ReplaceUnset)) }},
		{"MustParse ok", func() ([]any, error) { return r1(interval.Parse("-m6")) }, func() []any { return m1(interval.MustParse("-m6")) }},
		{"MustParse error", func() ([]any, error) { return r1(interval.Parse("TT")) }, func() []any { return m1(interval.MustParse("TT")) }},
		{"MustName ok", func() ([]any, error) { return r2(iv("dd5").Name()) }, func() []any { return m2(iv("dd5").MustName()) }},
		{"MustName without a name", func() ([]any, error) { return r2(iv("M16").Name()) }, func() []any { return m2(iv("M16").MustName()) }},
		{"MustName error", func() ([]any, error) { return r2(u.Name()) }, func() []any { return m2(u.MustName()) }},
		{"MustRussianName ok", func() ([]any, error) { return r2(iv("A4").RussianName()) },
			func() []any { return m2(iv("A4").MustRussianName()) }},
		{"MustRussianName error", func() ([]any, error) { return r2(u.RussianName()) }, func() []any { return m2(u.MustRussianName()) }},
		{"MustRussianAbbreviation ok", func() ([]any, error) { return r2(iv("dd5").RussianAbbreviation()) },
			func() []any { return m2(iv("dd5").MustRussianAbbreviation()) }},
		{"MustRussianAbbreviation error", func() ([]any, error) { return r2(u.RussianAbbreviation()) },
			func() []any { return m2(u.MustRussianAbbreviation()) }},
		{"MustParseName ok", func() ([]any, error) { return r1(interval.ParseName("ув.4")) },
			func() []any { return m1(interval.MustParseName("ув.4")) }},
		{"MustParseName error", func() ([]any, error) { return r1(interval.ParseName("major fifth")) },
			func() []any { return m1(interval.MustParseName("major fifth")) }},
	}
}

func TestMustVariants(t *testing.T) {
	t.Parallel()

	for _, c := range mustCases(t) {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			checkMust(t, c)
		})
	}
}

// TestMustChains covers the chains of 4.2: Must results feed further calls.
func TestMustChains(t *testing.T) {
	t.Parallel()

	assert.True(t, interval.MajorThird().MustDown().MustInvert().Equal(iv("↓m6")))
	assert.True(t, interval.MustParse("aug4").MustAdd(interval.PerfectFourth()).MustSimple().Equal(iv("M7")))
	assert.Equal(t, system.Step(6), interval.MustNew(interval.Augmented, 4).MustSteps(trad(), noStart))

	err := panicError(t, func() { interval.MajorThird().MustAdd(interval.Interval{}).MustInvert() })
	assertErr(t, err, interval.ErrUnset)
}
