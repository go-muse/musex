package note_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/octave/number"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Changing methods (architecture 6.6; OP-12...OP-24; P-6, P-10, KP-3;
// K-6, K-11, K-12, K-12a, K-20, K-25, K-27a, K-33, K-42...K-46).

// changeCase is one call of a changing method and its Must variant on a
// fresh note: the plain form returns only an error, the Must form returns the
// receiver (architecture 6.6.1).
type changeCase struct {
	name  string
	start func(t *testing.T) note.Note
	op    func(n *note.Note) error
	must  func(n *note.Note) *note.Note
	want  want
	err   error   // the note sentinel on failure
	extra []error // further causes; non-nil alone means a failure without a note sentinel
}

func (c changeCase) fails() bool { return c.err != nil || c.extra != nil }

// runChangeCases runs each case twice: with the plain method and with its
// Must variant. On success the note has the expected parts and Must returns
// the same pointer; on failure the note is unchanged (P-10) and Must panics
// with the same error.
func runChangeCases(t *testing.T, cases []changeCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			n := tc.start(t)
			m := n
			before := n

			err := tc.op(&n)
			if tc.fails() {
				assertErr(t, err, tc.err, tc.extra...)
				assertSame(t, before, n, "a failed change leaves the note unchanged")
				assertPanicErr(t, func() { tc.must(&m) }, tc.err, tc.extra...)
				assertSame(t, before, m, "a failed Must change leaves the note unchanged")
				return
			}
			require.NoError(t, err)
			assertNote(t, tc.want, n)

			got := tc.must(&m)
			assert.Same(t, &m, got, "Must returns the receiver, not a copy")
			assertNote(t, tc.want, m)
		})
	}
}

// Note builders for the tables.

func onPitch(name spelling.Spelling, aspn int, c cents.Cents, src settings.Source) func(*testing.T) note.Note {
	return func(t *testing.T) note.Note { return mk(t, name, aspn, c, src) }
}

func onName(name spelling.Spelling, c cents.Cents, src settings.Source) func(*testing.T) note.Note {
	return func(t *testing.T) note.Note { return mkName(t, name, c, src) }
}

func onUnset() func(*testing.T) note.Note {
	return func(*testing.T) note.Note { return note.Note{} }
}

// onInadmissible is C𝄪4, bound to a source that now gives the set of K-25.
func onInadmissible() func(*testing.T) note.Note {
	return func(t *testing.T) note.Note { n, _ := inadmissibleNote(t); return n }
}

// onInadmissibleName is C𝄪 without octave, bound like onInadmissible.
func onInadmissibleName() func(*testing.T) note.Note {
	return func(t *testing.T) note.Note { n, _ := inadmissibleNameNote(t); return n }
}

// onStale is the note bound to a source that now gives the unset set.
func onStale(name spelling.Spelling, aspn int) func(*testing.T) note.Note {
	return func(t *testing.T) note.Note {
		src := newSwitch(std())
		n := mk(t, name, aspn, ct(0), src)
		src.set(nil)
		return n
	}
}

// onStaleName is like onStale, without octave.
func onStaleName(name spelling.Spelling) func(*testing.T) note.Note {
	return func(t *testing.T) note.Note {
		src := newSwitch(std())
		n := mkName(t, name, ct(0), src)
		src.set(nil)
		return n
	}
}

// TestRespell covers OP-12, OP-13 and OP-17 (Respell, RespellWritten,
// RespellWrittenIn, ReplaceWithPitch, ReplaceWithSpelling).
func TestRespell(t *testing.T) {
	t.Parallel()

	s, k25, s24 := std(), setK25(), set24()
	respell := func(name spelling.Spelling) (func(*note.Note) error, func(*note.Note) *note.Note) {
		return func(n *note.Note) error { return n.Respell(name) },
			func(n *note.Note) *note.Note { return n.MustRespell(name) }
	}
	written := func(name spelling.Spelling) (func(*note.Note) error, func(*note.Note) *note.Note) {
		return func(n *note.Note) error { return n.RespellWritten(name) },
			func(n *note.Note) *note.Note { return n.MustRespellWritten(name) }
	}
	writtenIn := func(src settings.Source, name spelling.Spelling) (func(*note.Note) error, func(*note.Note) *note.Note) {
		return func(n *note.Note) error { return n.RespellWrittenIn(src, name) },
			func(n *note.Note) *note.Note { return n.MustRespellWrittenIn(src, name) }
	}
	withPitch := func(p pitch.Pitch) (func(*note.Note) error, func(*note.Note) *note.Note) {
		return func(n *note.Note) error { return n.ReplaceWithPitch(p) },
			func(n *note.Note) *note.Note { return n.MustReplaceWithPitch(p) }
	}
	withName := func(name spelling.Spelling) (func(*note.Note) error, func(*note.Note) *note.Note) {
		return func(n *note.Note) error { return n.ReplaceWithSpelling(name) },
			func(n *note.Note) *note.Note { return n.MustReplaceWithSpelling(name) }
	}
	c := func(name string, start func(*testing.T) note.Note,
		f func() (func(*note.Note) error, func(*note.Note) *note.Note), w want, err error, extra ...error,
	) changeCase {
		op, must := f()
		return changeCase{name: name, start: start, op: op, must: must, want: w, err: err, extra: extra}
	}
	call := func(f func(spelling.Spelling) (func(*note.Note) error, func(*note.Note) *note.Note), name spelling.Spelling) func() (func(*note.Note) error, func(*note.Note) *note.Note) {
		return func() (func(*note.Note) error, func(*note.Note) *note.Note) { return f(name) }
	}
	callIn := func(src settings.Source, name spelling.Spelling) func() (func(*note.Note) error, func(*note.Note) *note.Note) {
		return func() (func(*note.Note) error, func(*note.Note) *note.Note) { return writtenIn(src, name) }
	}
	callPitch := func(p pitch.Pitch) func() (func(*note.Note) error, func(*note.Note) *note.Note) {
		return func() (func(*note.Note) error, func(*note.Note) *note.Note) { return withPitch(p) }
	}

	runChangeCases(t, []changeCase{
		// OP-12: the Octave is kept.
		c("Respell keeps the Octave", onPitch(spelling.C(), 4, ct(30), s), call(respell, spelling.BSharp()),
			wp(spelling.BSharp(), 4, ct(30)).boundTo(s), nil),
		c("Respell without octave", onName(spelling.CSharp(), ct(30), s), call(respell, spelling.DFlat()),
			wn(spelling.DFlat(), ct(30)).boundTo(s), nil),
		c("Respell unbound does not check the name", onPitch(spelling.C(), 4, ct(0), nil), call(respell, cHalf()),
			wp(cHalf(), 4, ct(0)).free(), nil),
		c("Respell does not check the input name", onInadmissible(), call(respell, spelling.D()),
			wp(spelling.D(), 4, ct(0)), nil),
		c("Respell unset name", onPitch(spelling.C(), 4, ct(0), s), call(respell, spelling.Spelling{}), want{}, note.ErrUnset),
		c("Respell unset note", onUnset(), call(respell, spelling.C()), want{}, note.ErrUnset),
		c("Respell unset settings", onStale(spelling.C(), 4), call(respell, spelling.D()), want{}, note.ErrUnsetSettings, settings.ErrUnset),
		c("Respell inadmissible result", onPitch(spelling.C(), 4, ct(0), k25), call(respell, spelling.CDoubleSharp()), want{},
			note.ErrInadmissible, system.ErrInadmissible),

		// OP-13: the written octave is kept (K-12a).
		c("K-12a RespellWritten B3 +10 to B♯", onPitch(spelling.B(), 3, ct(10), s), call(written, spelling.BSharp()),
			wp(spelling.BSharp(), 4, ct(10)).boundTo(s), nil),
		c("RespellWritten C4 to B♯ is B♯ in Octave 5", onPitch(spelling.C(), 4, ct(0), s), call(written, spelling.BSharp()),
			wp(spelling.BSharp(), 5, ct(0)).boundTo(s), nil),
		c("RespellWritten C4 to C♭ is C♭ in Octave 3", onPitch(spelling.C(), 4, ct(-3), s), call(written, spelling.CFlat()),
			wp(spelling.CFlat(), 3, ct(-3)).boundTo(s), nil),
		c("RespellWritten unbound with octave", onPitch(spelling.C(), 4, ct(0), nil), call(written, spelling.BSharp()), want{}, note.ErrNoSource),
		c("RespellWritten unbound without octave works like Respell", onName(spelling.CSharp(), ct(5), nil), call(written, spelling.DFlat()),
			wn(spelling.DFlat(), ct(5)).free(), nil),
		c("RespellWritten bound without octave", onName(spelling.CSharp(), ct(5), s), call(written, spelling.DFlat()),
			wn(spelling.DFlat(), ct(5)).boundTo(s), nil),
		c("RespellWritten bound without octave checks the result", onName(spelling.C(), ct(0), k25), call(written, spelling.CDoubleSharp()),
			want{}, note.ErrInadmissible, system.ErrInadmissible),
		c("RespellWritten without octave does not check the input", onInadmissibleName(), call(written, spelling.D()),
			wn(spelling.D(), ct(0)), nil),
		c("Respell without octave unset settings", onStaleName(spelling.C()), call(respell, spelling.D()), want{}, note.ErrUnsetSettings, settings.ErrUnset),
		c("RespellWritten unset name before no source", onPitch(spelling.C(), 4, ct(0), nil), call(written, spelling.Spelling{}), want{}, note.ErrUnset),
		c("RespellWritten unset settings", onStale(spelling.C(), 4), call(written, spelling.D()), want{}, note.ErrUnsetSettings, settings.ErrUnset),
		c("RespellWritten inadmissible input", onInadmissible(), call(written, spelling.C()), want{}, note.ErrInadmissible, system.ErrInadmissible),
		c("RespellWritten inadmissible result", onPitch(spelling.C(), 4, ct(0), k25), call(written, spelling.CDoubleSharp()),
			want{}, note.ErrInadmissible, system.ErrInadmissible),
		c("RespellWritten Octave out of range", onPitch(spelling.C(), math.MaxInt, ct(0), s), call(written, spelling.BSharp()),
			want{}, note.ErrOutOfRange, system.ErrOutOfRange),

		// OP-13 with an explicit source (P-7, R-12).
		c("RespellWrittenIn unbound stays unbound", onPitch(spelling.C(), 4, ct(1), nil), callIn(s, spelling.BSharp()),
			wp(spelling.BSharp(), 5, ct(1)).free(), nil),
		c("RespellWrittenIn bound keeps its source", onPitch(spelling.C(), 4, ct(0), s), callIn(s24, spelling.D()),
			wp(spelling.D(), 4, ct(0)).boundTo(s), nil),
		c("RespellWrittenIn result inadmissible in the own set", onPitch(spelling.C(), 4, ct(0), s), callIn(s24, cHalf()),
			want{}, note.ErrInadmissible, system.ErrInadmissible),
		c("RespellWrittenIn unbound in set 24", onPitch(spelling.C(), 4, ct(0), nil), callIn(s24, cHalf()),
			wp(cHalf(), 4, ct(0)).free(), nil),
		c("RespellWrittenIn nil source", onPitch(spelling.C(), 4, ct(0), s), callIn(nil, spelling.D()), want{}, note.ErrUnset),
		c("RespellWrittenIn explicit unset settings", onPitch(spelling.C(), 4, ct(0), s), callIn(&unsetSource{}, spelling.D()),
			want{}, note.ErrUnsetSettings, settings.ErrUnset),
		c("RespellWrittenIn own unset settings", onStale(spelling.C(), 4), callIn(s, spelling.D()),
			want{}, note.ErrUnsetSettings, settings.ErrUnset),
		c("RespellWrittenIn unbound in the K-25 set", onPitch(spelling.C(), 4, ct(0), nil), callIn(k25, spelling.D()),
			wp(spelling.D(), 4, ct(0)).free(), nil),

		// OP-17: the record is replaced whole.
		c("ReplaceWithPitch adds the octave", onName(spelling.CSharp(), ct(30), s), callPitch(pt(spelling.D(), 4)),
			wp(spelling.D(), 4, ct(30)).boundTo(s), nil),
		c("ReplaceWithPitch replaces the pitch", onPitch(spelling.CSharp(), 2, ct(0), nil), callPitch(pt(spelling.EFlat(), 6)),
			wp(spelling.EFlat(), 6, ct(0)).free(), nil),
		c("ReplaceWithPitch unset pitch", onPitch(spelling.C(), 4, ct(0), s), callPitch(pitch.Pitch{}), want{}, note.ErrUnset),
		c("ReplaceWithPitch inadmissible", onPitch(spelling.C(), 4, ct(0), k25), callPitch(pt(spelling.CDoubleSharp(), 4)),
			want{}, note.ErrInadmissible, system.ErrInadmissible),
		c("ReplaceWithPitch unset settings", onStale(spelling.C(), 4), callPitch(pt(spelling.D(), 4)), want{}, note.ErrUnsetSettings, settings.ErrUnset),
		c("ReplaceWithSpelling removes the octave", onPitch(spelling.CSharp(), 4, ct(30), s), call(withName, spelling.D()),
			wn(spelling.D(), ct(30)).boundTo(s), nil),
		c("ReplaceWithSpelling without octave", onName(spelling.CSharp(), ct(0), nil), call(withName, spelling.FSharp()),
			wn(spelling.FSharp(), ct(0)).free(), nil),
		c("ReplaceWithSpelling unset name", onPitch(spelling.C(), 4, ct(0), s), call(withName, spelling.Spelling{}), want{}, note.ErrUnset),
		c("ReplaceWithSpelling inadmissible", onPitch(spelling.C(), 4, ct(0), k25), call(withName, spelling.CDoubleSharp()),
			want{}, note.ErrInadmissible, system.ErrInadmissible),
		c("ReplaceWithSpelling unset note", onUnset(), call(withName, spelling.D()), want{}, note.ErrUnset),
	})
}

// TestSetOctaveAndCents covers OP-14, OP-15, OP-16, OP-18 (K-20, G-4).
func TestSetOctaveAndCents(t *testing.T) {
	t.Parallel()

	s, fb, s19 := std(), stdFB(), set19()
	c := func(name string, start func(*testing.T) note.Note, f func(*note.Note) error, m func(*note.Note) *note.Note,
		w want, err error, extra ...error,
	) changeCase {
		return changeCase{name: name, start: start, op: f, must: m, want: w, err: err, extra: extra}
	}
	setOct := func(o octave.Octave) (func(*note.Note) error, func(*note.Note) *note.Note) {
		return func(n *note.Note) error { return n.SetOctave(o) }, func(n *note.Note) *note.Note { return n.MustSetOctave(o) }
	}
	setASPN := func(a int) (func(*note.Note) error, func(*note.Note) *note.Note) {
		return func(n *note.Note) error { return n.SetASPN(a) }, func(n *note.Note) *note.Note { return n.MustSetASPN(a) }
	}
	wOct := func(w octave.Octave) (func(*note.Note) error, func(*note.Note) *note.Note) {
		return func(n *note.Note) error { return n.SetWrittenOctave(w) },
			func(n *note.Note) *note.Note { return n.MustSetWrittenOctave(w) }
	}
	wASPN := func(a int) (func(*note.Note) error, func(*note.Note) *note.Note) {
		return func(n *note.Note) error { return n.SetWrittenASPN(a) },
			func(n *note.Note) *note.Note { return n.MustSetWrittenASPN(a) }
	}
	wNum := func(num int) (func(*note.Note) error, func(*note.Note) *note.Note) {
		return func(n *note.Note) error { return n.SetWrittenNumber(num) },
			func(n *note.Note) *note.Note { return n.MustSetWrittenNumber(num) }
	}
	wOctIn := func(src settings.Source, w octave.Octave) (func(*note.Note) error, func(*note.Note) *note.Note) {
		return func(n *note.Note) error { return n.SetWrittenOctaveIn(src, w) },
			func(n *note.Note) *note.Note { return n.MustSetWrittenOctaveIn(src, w) }
	}
	wNumIn := func(src settings.Source, num int) (func(*note.Note) error, func(*note.Note) *note.Note) {
		return func(n *note.Note) error { return n.SetWrittenNumberIn(src, num) },
			func(n *note.Note) *note.Note { return n.MustSetWrittenNumberIn(src, num) }
	}
	removeOct := func(n *note.Note) error { return n.RemoveOctave() }
	mustRemoveOct := func(n *note.Note) *note.Note { return n.MustRemoveOctave() }
	setCents := func(x cents.Cents) (func(*note.Note) error, func(*note.Note) *note.Note) {
		return func(n *note.Note) error { return n.SetCents(x) }, func(n *note.Note) *note.Note { return n.MustSetCents(x) }
	}
	var cases []changeCase
	add := func(name string, start func(*testing.T) note.Note, fm func() (func(*note.Note) error, func(*note.Note) *note.Note),
		w want, err error, extra ...error,
	) {
		f, m := fm()
		cases = append(cases, c(name, start, f, m, w, err, extra...))
	}
	fm := func(f func(*note.Note) error, m func(*note.Note) *note.Note) func() (func(*note.Note) error, func(*note.Note) *note.Note) {
		return func() (func(*note.Note) error, func(*note.Note) *note.Note) { return f, m }
	}
	// OP-14: SetOctave, SetASPN (group "N").
	f, m := setOct(oct(4))
	add("G-4 SetOctave adds the octave and keeps the cents", onName(spelling.CSharp(), ct(30), s), fm(f, m), wp(spelling.CSharp(), 4, ct(30)).boundTo(s), nil)
	f, m = setOct(oct(6))
	add("SetOctave replaces the Octave", onPitch(spelling.B(), 3, ct(0), nil), fm(f, m), wp(spelling.B(), 6, ct(0)).free(), nil)
	add("SetOctave on a note with an inadmissible name", onInadmissible(), fm(f, m), wp(spelling.CDoubleSharp(), 6, ct(0)), nil)
	add("SetOctave with unset settings does not read the source", onStale(spelling.C(), 4), fm(f, m), wp(spelling.C(), 6, ct(0)), nil)
	f, m = setOct(octave.Octave{})
	add("SetOctave unset octave", onPitch(spelling.B(), 3, ct(0), s), fm(f, m), want{}, note.ErrUnset)
	f, m = setOct(oct(4))
	add("SetOctave unset note", onUnset(), fm(f, m), want{}, note.ErrUnset)
	f, m = setASPN(math.MinInt)
	add("SetASPN any int", onName(spelling.E(), ct(0), s), fm(f, m), wp(spelling.E(), math.MinInt, ct(0)).boundTo(s), nil)
	add("SetASPN unset note", onUnset(), fm(f, m), want{}, note.ErrUnset)

	// OP-15: SetWrittenOctave and its forms (group "S").
	f, m = wOct(oct(3))
	add("SetWrittenOctave B♯ written 3 is Octave 4", onName(spelling.BSharp(), ct(2), s), fm(f, m), wp(spelling.BSharp(), 4, ct(2)).boundTo(s), nil)
	f, m = wOct(oct(5))
	add("SetWrittenOctave C♭ written 5 is Octave 4", onPitch(spelling.CFlat(), 3, ct(0), s), fm(f, m), wp(spelling.CFlat(), 4, ct(0)).boundTo(s), nil)
	add("SetWrittenOctave unbound with octave", onPitch(spelling.C(), 4, ct(0), nil), fm(f, m), want{}, note.ErrNoSource)
	add("SetWrittenOctave unbound without octave", onName(spelling.C(), ct(0), nil), fm(f, m), want{}, note.ErrNoSource)
	add("SetWrittenOctave unset settings", onStale(spelling.C(), 4), fm(f, m), want{}, note.ErrUnsetSettings, settings.ErrUnset)
	add("SetWrittenOctave inadmissible name", onInadmissible(), fm(f, m), want{}, note.ErrInadmissible, system.ErrInadmissible)
	f, m = wOct(octave.Octave{})
	add("SetWrittenOctave unset octave before no source", onPitch(spelling.C(), 4, ct(0), nil), fm(f, m), want{}, note.ErrUnset)
	f, m = wASPN(3)
	add("SetWrittenASPN", onPitch(spelling.BSharp(), 7, ct(0), s), fm(f, m), wp(spelling.BSharp(), 4, ct(0)).boundTo(s), nil)
	f, m = wASPN(math.MaxInt)
	add("SetWrittenASPN Octave out of range", onName(spelling.BSharp(), ct(0), s), fm(f, m), want{}, note.ErrOutOfRange, system.ErrOutOfRange)
	f, m = wNum(2)
	add("SetWrittenNumber Franco-Belgian", onName(spelling.BSharp(), ct(0), fb), fm(f, m), wp(spelling.BSharp(), 4, ct(0)).boundTo(fb), nil)
	f, m = wNum(0)
	add("SetWrittenNumber skipped number", onName(spelling.BSharp(), ct(0), fb), fm(f, m), want{}, nil, number.ErrSkipped)
	f, m = wNum(math.MaxInt)
	add("SetWrittenNumber number out of range", onName(spelling.C(), ct(0), fb), fm(f, m), want{}, note.ErrOutOfRange, octave.ErrOutOfRange, number.ErrOutOfRange)
	add("SetWrittenNumber unbound", onName(spelling.C(), ct(0), nil), fm(f, m), want{}, note.ErrNoSource)
	f, m = wOctIn(s, oct(3))
	add("SetWrittenOctaveIn unbound stays unbound", onName(spelling.BSharp(), ct(0), nil), fm(f, m), wp(spelling.BSharp(), 4, ct(0)).free(), nil)
	f, m = wOctIn(s19, oct(3))
	add("SetWrittenOctaveIn set 19 keeps the own source", onName(spelling.BSharp(), ct(0), s), fm(f, m), wp(spelling.BSharp(), 3, ct(0)).boundTo(s), nil)
	f, m = wOctIn(nil, oct(3))
	add("SetWrittenOctaveIn nil source", onName(spelling.BSharp(), ct(0), nil), fm(f, m), want{}, note.ErrUnset)
	f, m = wOctIn(&unsetSource{}, oct(3))
	add("SetWrittenOctaveIn unset settings", onName(spelling.BSharp(), ct(0), nil), fm(f, m), want{}, note.ErrUnsetSettings, settings.ErrUnset)
	f, m = wNumIn(fb, 2)
	add("SetWrittenNumberIn unbound", onName(spelling.BSharp(), ct(0), nil), fm(f, m), wp(spelling.BSharp(), 4, ct(0)).free(), nil)
	f, m = wNumIn(nil, 2)
	add("SetWrittenNumberIn nil source", onName(spelling.BSharp(), ct(0), s), fm(f, m), want{}, note.ErrUnset)
	f, m = wNumIn(fb, 0)
	add("SetWrittenNumberIn skipped number", onName(spelling.BSharp(), ct(0), nil), fm(f, m), want{}, nil, number.ErrSkipped)

	// OP-16 (K-20: add the octave, then remove it).
	add("K-20 RemoveOctave", onPitch(spelling.CSharp(), 4, ct(30), s), fm(removeOct, mustRemoveOct), wn(spelling.CSharp(), ct(30)).boundTo(s), nil)
	add("RemoveOctave without octave is unchanged", onName(spelling.CSharp(), ct(30), nil), fm(removeOct, mustRemoveOct), wn(spelling.CSharp(), ct(30)).free(), nil)
	add("RemoveOctave does not read the source", onStale(spelling.C(), 4), fm(removeOct, mustRemoveOct), wn(spelling.C(), ct(0)), nil)
	add("RemoveOctave unset note", onUnset(), fm(removeOct, mustRemoveOct), want{}, note.ErrUnset)

	// OP-18.
	f, m = setCents(ct(-10))
	add("SetCents", onPitch(spelling.C(), 4, ct(30), s), fm(f, m), wp(spelling.C(), 4, ct(-10)).boundTo(s), nil)
	add("SetCents without octave", onName(spelling.C(), ct(30), nil), fm(f, m), wn(spelling.C(), ct(-10)).free(), nil)
	add("SetCents does not read the source", onStale(spelling.C(), 4), fm(f, m), wp(spelling.C(), 4, ct(-10)), nil)
	f, m = setCents(cents.Cents{})
	add("SetCents zero value is 0", onPitch(spelling.C(), 4, ct(30), s), fm(f, m), wp(spelling.C(), 4, ct(0)).boundTo(s), nil)
	f, m = setCents(cents.MustFromInt(cents.MaxMagnitude))
	add("SetCents without a limit (P-9)", onPitch(spelling.C(), 4, ct(0), s), fm(f, m), wp(spelling.C(), 4, cents.MustFromInt(cents.MaxMagnitude)).boundTo(s), nil)
	add("SetCents unset note", onUnset(), fm(f, m), want{}, note.ErrUnset)

	runChangeCases(t, cases)
}

// TestChangeAlteration covers OP-19...OP-22 (K-11, K-12, K-25, K-27a).
func TestChangeAlteration(t *testing.T) {
	t.Parallel()

	s, k25, s24, s19 := std(), setK25(), set24(), set19()
	tr := system.Traditional()
	type fm = func() (func(*note.Note) error, func(*note.Note) *note.Note)
	mkop := func(f func(*note.Note) error, m func(*note.Note) *note.Note) fm {
		return func() (func(*note.Note) error, func(*note.Note) *note.Note) { return f, m }
	}
	raise := mkop(func(n *note.Note) error { return n.Raise() }, func(n *note.Note) *note.Note { return n.MustRaise() })
	lower := mkop(func(n *note.Note) error { return n.Lower() }, func(n *note.Note) *note.Note { return n.MustLower() })
	raiseBy := func(k int) fm {
		return mkop(func(n *note.Note) error { return n.RaiseBy(k) }, func(n *note.Note) *note.Note { return n.MustRaiseBy(k) })
	}
	lowerBy := func(k int) fm {
		return mkop(func(n *note.Note) error { return n.LowerBy(k) }, func(n *note.Note) *note.Note { return n.MustLowerBy(k) })
	}
	raiseByAlt := func(d alteration.Alteration) fm {
		return mkop(func(n *note.Note) error { return n.RaiseByAlteration(d) },
			func(n *note.Note) *note.Note { return n.MustRaiseByAlteration(d) })
	}
	lowerByAlt := func(d alteration.Alteration) fm {
		return mkop(func(n *note.Note) error { return n.LowerByAlteration(d) },
			func(n *note.Note) *note.Note { return n.MustLowerByAlteration(d) })
	}
	setAlt := func(a alteration.Alteration) fm {
		return mkop(func(n *note.Note) error { return n.SetAlteration(a) },
			func(n *note.Note) *note.Note { return n.MustSetAlteration(a) })
	}
	raiseW := mkop(func(n *note.Note) error { return n.RaiseWritten() }, func(n *note.Note) *note.Note { return n.MustRaiseWritten() })
	lowerW := mkop(func(n *note.Note) error { return n.LowerWritten() }, func(n *note.Note) *note.Note { return n.MustLowerWritten() })
	raiseWBy := func(k int) fm {
		return mkop(func(n *note.Note) error { return n.RaiseWrittenBy(k) },
			func(n *note.Note) *note.Note { return n.MustRaiseWrittenBy(k) })
	}
	lowerWBy := func(k int) fm {
		return mkop(func(n *note.Note) error { return n.LowerWrittenBy(k) },
			func(n *note.Note) *note.Note { return n.MustLowerWrittenBy(k) })
	}
	raiseWByAlt := func(d alteration.Alteration) fm {
		return mkop(func(n *note.Note) error { return n.RaiseWrittenByAlteration(d) },
			func(n *note.Note) *note.Note { return n.MustRaiseWrittenByAlteration(d) })
	}
	lowerWByAlt := func(d alteration.Alteration) fm {
		return mkop(func(n *note.Note) error { return n.LowerWrittenByAlteration(d) },
			func(n *note.Note) *note.Note { return n.MustLowerWrittenByAlteration(d) })
	}
	setAltW := func(a alteration.Alteration) fm {
		return mkop(func(n *note.Note) error { return n.SetAlterationWritten(a) },
			func(n *note.Note) *note.Note { return n.MustSetAlterationWritten(a) })
	}
	raiseWByAltIn := func(src settings.Source, d alteration.Alteration) fm {
		return mkop(func(n *note.Note) error { return n.RaiseWrittenByAlterationIn(src, d) },
			func(n *note.Note) *note.Note { return n.MustRaiseWrittenByAlterationIn(src, d) })
	}
	setAltWIn := func(src settings.Source, a alteration.Alteration) fm {
		return mkop(func(n *note.Note) error { return n.SetAlterationWrittenIn(src, a) },
			func(n *note.Note) *note.Note { return n.MustSetAlterationWrittenIn(src, a) })
	}
	natural := mkop(func(n *note.Note) error { return n.Naturalize() }, func(n *note.Note) *note.Note { return n.MustNaturalize() })
	naturalW := mkop(func(n *note.Note) error { return n.NaturalizeWritten() },
		func(n *note.Note) *note.Note { return n.MustNaturalizeWritten() })
	naturalWIn := func(src settings.Source) fm {
		return mkop(func(n *note.Note) error { return n.NaturalizeWrittenIn(src) },
			func(n *note.Note) *note.Note { return n.MustNaturalizeWrittenIn(src) })
	}
	c := func(name string, start func(*testing.T) note.Note, f fm, w want, err error, extra ...error) changeCase {
		op, must := f()
		return changeCase{name: name, start: start, op: op, must: must, want: w, err: err, extra: extra}
	}
	// fromPitch is the expectation of a pitch computed by a dependency.
	fromPitch := func(p pitch.Pitch, x cents.Cents) want { return wp(p.Spelling(), p.MustASPN(), x) }
	b3 := pt(spelling.B(), 3)
	c4 := pt(spelling.C(), 4)
	cFlat4 := pt(spelling.CFlat(), 3) // written "C♭4"
	big := 1 << 25                    // beyond the range of package alteration

	runChangeCases(t, []changeCase{
		// OP-19: the Octave is kept, as pitch does (K-11).
		c("K-11 Raise B3 keeps the Octave", onPitch(spelling.B(), 3, ct(4), s), raise, fromPitch(b3.MustRaise(), ct(4)).boundTo(s), nil),
		c("Lower", onPitch(spelling.C(), 4, ct(0), s), lower, fromPitch(c4.MustLower(), ct(0)).boundTo(s), nil),
		c("RaiseBy 2", onPitch(spelling.C(), 4, ct(0), nil), raiseBy(2), wp(spelling.CDoubleSharp(), 4, ct(0)).free(), nil),
		c("LowerBy 2", onName(spelling.D(), ct(9), s), lowerBy(2), wn(spelling.DDoubleFlat(), ct(9)).boundTo(s), nil),
		c("RaiseBy 0", onName(spelling.D(), ct(0), s), raiseBy(0), wn(spelling.D(), ct(0)).boundTo(s), nil),
		c("RaiseByAlteration 1/2 unbound", onPitch(spelling.C(), 4, ct(0), nil), raiseByAlt(alt(1, 2)), wp(cHalf(), 4, ct(0)).free(), nil),
		c("LowerByAlteration 1/2 unbound", onPitch(spelling.C(), 4, ct(0), nil), lowerByAlt(alt(1, 2)), wp(sp(letter.C, -1, 2), 4, ct(0)).free(), nil),
		c("RaiseByAlteration 1/2 bound to the standard set", onPitch(spelling.C(), 4, ct(0), s), raiseByAlt(alt(1, 2)), want{},
			note.ErrInadmissible, system.ErrInadmissible),
		c("SetAlteration", onPitch(spelling.E(), 4, ct(0), s), setAlt(alt(-2, 1)), wp(spelling.EDoubleFlat(), 4, ct(0)).boundTo(s), nil),
		c("SetAlteration natural", onName(spelling.ESharp(), ct(0), s), setAlt(alteration.Alteration{}), wn(spelling.E(), ct(0)).boundTo(s), nil),
		c("K-25 Raise bound C♯", onName(spelling.CSharp(), ct(0), k25), raise, want{}, note.ErrInadmissible, system.ErrInadmissible),
		c("K-25 Raise unbound C♯", onName(spelling.CSharp(), ct(0), nil), raise, wn(spelling.CDoubleSharp(), ct(0)).free(), nil),
		c("Raise checks only the result: Lower of an inadmissible name", onInadmissible(), lower, wp(spelling.CSharp(), 4, ct(0)), nil),
		c("Raise of an inadmissible name", onInadmissible(), raise, want{}, note.ErrInadmissible, system.ErrInadmissible),
		c("RaiseBy out of range", onPitch(spelling.C(), 4, ct(0), nil), raiseBy(big), want{}, note.ErrOutOfRange, alteration.ErrOutOfRange),
		c("LowerBy smallest int", onPitch(spelling.C(), 4, ct(0), nil), lowerBy(math.MinInt), want{}, note.ErrOutOfRange, alteration.ErrOutOfRange),
		c("out of range before inadmissible", onPitch(spelling.C(), 4, ct(0), k25), raiseBy(big), want{}, note.ErrOutOfRange, alteration.ErrOutOfRange),
		c("Raise unset settings", onStale(spelling.C(), 4), raise, want{}, note.ErrUnsetSettings, settings.ErrUnset),
		c("Raise unset note", onUnset(), raise, want{}, note.ErrUnset),
		c("SetAlteration unset note", onUnset(), setAlt(alt(1, 1)), want{}, note.ErrUnset),

		// OP-20: the written octave is kept (K-11).
		c("K-11 RaiseWritten B3", onPitch(spelling.B(), 3, ct(4), s), raiseW, fromPitch(tr.MustRaiseWritten(b3), ct(4)).boundTo(s), nil),
		c("LowerWritten C4 is C♭ in Octave 3", onPitch(spelling.C(), 4, ct(0), s), lowerW, fromPitch(tr.MustLowerWritten(c4), ct(0)).boundTo(s), nil),
		c("RaiseWrittenBy 2 B3", onPitch(spelling.B(), 3, ct(0), s), raiseWBy(2), fromPitch(tr.MustRaiseWrittenBy(b3, 2), ct(0)).boundTo(s), nil),
		c("LowerWrittenBy 1 C4", onPitch(spelling.C(), 4, ct(0), s), lowerWBy(1), fromPitch(tr.MustLowerWrittenBy(c4, 1), ct(0)).boundTo(s), nil),
		c("RaiseWrittenByAlteration 1", onPitch(spelling.B(), 3, ct(0), s), raiseWByAlt(alt(1, 1)), fromPitch(tr.MustRaiseWrittenByAlteration(b3, alt(1, 1)), ct(0)).boundTo(s), nil),
		c("LowerWrittenByAlteration 1", onPitch(spelling.C(), 4, ct(0), s), lowerWByAlt(alt(1, 1)), fromPitch(tr.MustLowerWrittenByAlteration(c4, alt(1, 1)), ct(0)).boundTo(s), nil),
		c("SetAlterationWritten +1", onPitch(spelling.B(), 3, ct(0), s), setAltW(alt(1, 1)), wp(spelling.BSharp(), 4, ct(0)).boundTo(s), nil),
		c("RaiseWrittenByAlteration 1/2 bound to the standard set", onPitch(spelling.C(), 4, ct(0), s), raiseWByAlt(alt(1, 2)), want{},
			note.ErrInadmissible, system.ErrInadmissible),
		c("RaiseWritten unbound with octave", onPitch(spelling.B(), 3, ct(0), nil), raiseW, want{}, note.ErrNoSource),
		c("RaiseWritten unbound without octave works like Raise", onName(spelling.CSharp(), ct(0), nil), raiseW, wn(spelling.CDoubleSharp(), ct(0)).free(), nil),
		c("RaiseWritten bound without octave checks the result", onName(spelling.CSharp(), ct(0), k25), raiseW, want{}, note.ErrInadmissible, system.ErrInadmissible),
		c("LowerWritten inadmissible input", onInadmissible(), lowerW, want{}, note.ErrInadmissible, system.ErrInadmissible),
		c("RaiseWrittenBy out of range before inadmissible result", onPitch(spelling.C(), 4, ct(0), k25), raiseWBy(big), want{}, note.ErrOutOfRange),
		c("RaiseWritten Octave out of range", onPitch(spelling.B(), math.MaxInt, ct(0), s), raiseW, want{}, note.ErrOutOfRange, system.ErrOutOfRange),
		c("RaiseWritten unset settings", onStale(spelling.B(), 3), raiseW, want{}, note.ErrUnsetSettings, settings.ErrUnset),

		// OP-20 with an explicit source (K-27a).
		c("K-27a bound to the standard set, set 24", onPitch(spelling.C(), 4, ct(0), s), raiseWByAltIn(s24, alt(1, 2)), want{},
			note.ErrInadmissible, system.ErrInadmissible),
		c("K-27a unbound, set 24", onPitch(spelling.C(), 4, ct(0), nil), raiseWByAltIn(s24, alt(1, 2)), wp(cHalf(), 4, ct(0)).free(), nil),
		c("RaiseWrittenByAlterationIn lowering by a negative change", onPitch(spelling.C(), 4, ct(0), s), raiseWByAltIn(s, alt(-1, 1)),
			fromPitch(tr.MustLowerWritten(c4), ct(0)).boundTo(s), nil),
		c("RaiseWrittenByAlterationIn set 19 keeps the own source", onPitch(spelling.B(), 3, ct(0), s), raiseWByAltIn(s19, alt(1, 1)),
			fromPitch(system.Positions19().MustRaiseWritten(b3), ct(0)).boundTo(s), nil),
		c("RaiseWrittenByAlterationIn nil source", onPitch(spelling.C(), 4, ct(0), s), raiseWByAltIn(nil, alt(1, 1)), want{}, note.ErrUnset),
		c("SetAlterationWrittenIn unbound", onPitch(spelling.C(), 4, ct(0), nil), setAltWIn(s24, alt(1, 2)), wp(cHalf(), 4, ct(0)).free(), nil),
		c("SetAlterationWrittenIn explicit unset settings", onPitch(spelling.C(), 4, ct(0), nil), setAltWIn(&unsetSource{}, alt(1, 2)), want{},
			note.ErrUnsetSettings, settings.ErrUnset),

		// OP-21, OP-22 (K-12).
		c("K-12 Naturalize C♭ keeps the Octave", onPitch(spelling.CFlat(), 3, ct(0), s), natural, fromPitch(cFlat4.Natural(), ct(0)).boundTo(s), nil),
		c("K-12 NaturalizeWritten C♭4 is C4", onPitch(spelling.CFlat(), 3, ct(0), s), naturalW, wp(spelling.C(), 4, ct(0)).boundTo(s), nil),
		c("Naturalize of an inadmissible name", onInadmissible(), natural, wp(spelling.C(), 4, ct(0)), nil),
		c("Naturalize unset settings", onStale(spelling.CSharp(), 4), natural, want{}, note.ErrUnsetSettings, settings.ErrUnset),
		c("Naturalize unset note", onUnset(), natural, want{}, note.ErrUnset),
		c("NaturalizeWritten unbound with octave", onPitch(spelling.CFlat(), 3, ct(0), nil), naturalW, want{}, note.ErrNoSource),
		c("NaturalizeWritten unbound without octave", onName(spelling.CFlat(), ct(3), nil), naturalW, wn(spelling.C(), ct(3)).free(), nil),
		c("NaturalizeWritten inadmissible input", onInadmissible(), naturalW, want{}, note.ErrInadmissible, system.ErrInadmissible),
		c("NaturalizeWritten Octave out of range", onPitch(spelling.BSharp(), math.MinInt, ct(0), s), naturalW, want{}, note.ErrOutOfRange, system.ErrOutOfRange),
		c("NaturalizeWrittenIn unbound", onPitch(spelling.CFlat(), 3, ct(0), nil), naturalWIn(s), wp(spelling.C(), 4, ct(0)).free(), nil),
		c("NaturalizeWrittenIn inadmissible input in the explicit set", onPitch(spelling.CDoubleSharp(), 4, ct(0), nil), naturalWIn(k25), want{},
			note.ErrInadmissible, system.ErrInadmissible),
		c("NaturalizeWrittenIn nil source", onPitch(spelling.CFlat(), 3, ct(0), nil), naturalWIn(nil), want{}, note.ErrUnset),
	})
}

// TestShiftAndMove covers OP-23 and OP-24 (K-6, K-20, K-33).
func TestShiftAndMove(t *testing.T) {
	t.Parallel()

	s, s24 := std(), set24()
	type fm = func() (func(*note.Note) error, func(*note.Note) *note.Note)
	mkop := func(f func(*note.Note) error, m func(*note.Note) *note.Note) fm {
		return func() (func(*note.Note) error, func(*note.Note) *note.Note) { return f, m }
	}
	shift := func(k int) fm {
		return mkop(func(n *note.Note) error { return n.Shift(k) }, func(n *note.Note) *note.Note { return n.MustShift(k) })
	}
	shiftDown := func(k int) fm {
		return mkop(func(n *note.Note) error { return n.ShiftDown(k) }, func(n *note.Note) *note.Note { return n.MustShiftDown(k) })
	}
	up := mkop(func(n *note.Note) error { return n.Up() }, func(n *note.Note) *note.Note { return n.MustUp() })
	down := mkop(func(n *note.Note) error { return n.Down() }, func(n *note.Note) *note.Note { return n.MustDown() })
	move := func(k system.Step, rule system.SpellingRule) fm {
		return mkop(func(n *note.Note) error { return n.Move(k, rule) }, func(n *note.Note) *note.Note { return n.MustMove(k, rule) })
	}
	moveIn := func(src settings.Source, k system.Step, rule system.SpellingRule) fm {
		return mkop(func(n *note.Note) error { return n.MoveIn(src, k, rule) },
			func(n *note.Note) *note.Note { return n.MustMoveIn(src, k, rule) })
	}
	c := func(name string, start func(*testing.T) note.Note, f fm, w want, err error, extra ...error) changeCase {
		op, must := f()
		return changeCase{name: name, start: start, op: op, must: must, want: w, err: err, extra: extra}
	}
	minDown := pt(spelling.C(), math.MinInt).MustShiftDown(math.MinInt)

	runChangeCases(t, []changeCase{
		c("K-6 Shift 2 keeps the cents", onPitch(spelling.C(), 4, ct(30), s), shift(2), wp(spelling.C(), 6, ct(30)).boundTo(s), nil),
		c("Shift negative", onPitch(spelling.C(), 4, ct(0), nil), shift(-3), wp(spelling.C(), 1, ct(0)).free(), nil),
		c("ShiftDown", onPitch(spelling.BSharp(), 4, ct(0), s), shiftDown(1), wp(spelling.BSharp(), 3, ct(0)).boundTo(s), nil),
		c("ShiftDown smallest int is exact", onPitch(spelling.C(), math.MinInt, ct(0), nil), shiftDown(math.MinInt),
			wp(spelling.C(), minDown.MustASPN(), ct(0)).free(), nil),
		c("Up", onPitch(spelling.C(), 4, ct(0), s), up, wp(spelling.C(), 5, ct(0)).boundTo(s), nil),
		c("Down", onPitch(spelling.C(), 4, ct(0), s), down, wp(spelling.C(), 3, ct(0)).boundTo(s), nil),
		c("Shift does not read the source", onStale(spelling.C(), 4), shift(1), wp(spelling.C(), 5, ct(0)), nil),
		c("K-20 Shift without octave", onName(spelling.CSharp(), ct(30), s), shift(1), want{}, note.ErrNoOctave),
		c("Up without octave unbound", onName(spelling.CSharp(), ct(30), nil), up, want{}, note.ErrNoOctave),
		c("Up out of range", onPitch(spelling.C(), math.MaxInt, ct(0), s), up, want{}, note.ErrOutOfRange, octave.ErrOutOfRange),
		c("Down out of range", onPitch(spelling.C(), math.MinInt, ct(0), s), down, want{}, note.ErrOutOfRange, octave.ErrOutOfRange),
		c("Shift unset note", onUnset(), shift(1), want{}, note.ErrUnset),

		c("K-33 Move +1 flats", onPitch(spelling.C(), 4, ct(30), s), move(1, system.Flats()), wp(spelling.DFlat(), 4, ct(30)).boundTo(s), nil),
		c("K-33 Move +1 flats without octave", onName(spelling.C(), ct(30), s), move(1, system.Flats()), wn(spelling.DFlat(), ct(30)).boundTo(s), nil),
		c("Move 0 chooses the name again", onPitch(spelling.DFlat(), 4, ct(0), s), move(0, system.Sharps()), wp(spelling.CSharp(), 4, ct(0)).boundTo(s), nil),
		c("Move down across the octave", onPitch(spelling.C(), 4, ct(0), s), move(-1, system.Sharps()), wp(spelling.B(), 3, ct(0)).boundTo(s), nil),
		c("Move unbound with octave", onPitch(spelling.C(), 4, ct(0), nil), move(1, system.Sharps()), want{}, note.ErrNoSource),
		c("Move unbound without octave", onName(spelling.C(), ct(0), nil), move(1, system.Sharps()), want{}, note.ErrNoSource),
		c("Move unset rule before no source", onName(spelling.C(), ct(0), nil), move(1, system.SpellingRule{}), want{}, note.ErrUnset),
		c("Move unset settings", onStale(spelling.C(), 4), move(1, system.Sharps()), want{}, note.ErrUnsetSettings, settings.ErrUnset),
		c("Move inadmissible input", onInadmissible(), move(1, system.Sharps()), want{}, note.ErrInadmissible, system.ErrInadmissible),
		c("Move bad rule: reason of system without note sentinel", onPitch(spelling.C(), 4, ct(0), s), move(1, badRule()), want{}, nil, system.ErrBadSpelling),
		c("Move Octave out of range", onPitch(spelling.C(), math.MaxInt, ct(0), s), move(12, system.Sharps()), want{}, note.ErrOutOfRange, system.ErrOutOfRange),
		c("MoveIn unbound set 24", onPitch(spelling.C(), 4, ct(0), nil), moveIn(s24, 1, system.Sharps()), wp(cHalf(), 4, ct(0)).free(), nil),
		c("MoveIn result inadmissible in the own set", onPitch(spelling.C(), 4, ct(0), s), moveIn(s24, 1, system.Sharps()), want{},
			note.ErrInadmissible, system.ErrInadmissible),
		c("MoveIn keeps the own source", onPitch(spelling.C(), 4, ct(0), s), moveIn(s24, 2, system.Sharps()), wp(spelling.CSharp(), 4, ct(0)).boundTo(s), nil),
		c("MoveIn without octave unbound", onName(spelling.C(), ct(0), nil), moveIn(s24, 1, system.Sharps()), wn(cHalf(), ct(0)).free(), nil),
		c("MoveIn nil source", onPitch(spelling.C(), 4, ct(0), s), moveIn(nil, 1, system.Sharps()), want{}, note.ErrUnset),
		c("MoveIn unset settings", onPitch(spelling.C(), 4, ct(0), s), moveIn(&unsetSource{}, 1, system.Sharps()), want{}, note.ErrUnsetSettings, settings.ErrUnset),
		c("MoveIn bad rule", onPitch(spelling.C(), 4, ct(0), nil), moveIn(s, 1, badRule()), want{}, nil, system.ErrBadSpelling),
	})
}

// TestChangeReadsTheSource covers the reading groups of architecture 6.6.1:
// "N" never reads, "N/S" reads once for a bound note and never for an
// unbound one, "S" reads once; an In form reads the explicit and the own
// source once each, or once when they are one source (P-5, P-5a, P-7).
func TestChangeReadsTheSource(t *testing.T) {
	t.Parallel()

	type op struct {
		name  string
		group string // "N", "N/S", "S"
		start func(*testing.T, settings.Source) note.Note
		call  func(*note.Note) error
	}
	withOctave := func(name spelling.Spelling, aspn int) func(*testing.T, settings.Source) note.Note {
		return func(t *testing.T, src settings.Source) note.Note { return mk(t, name, aspn, ct(0), src) }
	}
	b3 := withOctave(spelling.B(), 3)
	c4 := withOctave(spelling.C(), 4)
	ops := []op{
		{"SetOctave", "N", c4, func(n *note.Note) error { return n.SetOctave(oct(5)) }},
		{"SetASPN", "N", c4, func(n *note.Note) error { return n.SetASPN(5) }},
		{"RemoveOctave", "N", c4, func(n *note.Note) error { return n.RemoveOctave() }},
		{"SetCents", "N", c4, func(n *note.Note) error { return n.SetCents(ct(5)) }},
		{"Shift", "N", c4, func(n *note.Note) error { return n.Shift(1) }},
		{"ShiftDown", "N", c4, func(n *note.Note) error { return n.ShiftDown(1) }},
		{"Up", "N", c4, func(n *note.Note) error { return n.Up() }},
		{"Down", "N", c4, func(n *note.Note) error { return n.Down() }},
		{"Unbind", "N", c4, func(n *note.Note) error { return n.Unbind() }},
		{"Respell", "N/S", c4, func(n *note.Note) error { return n.Respell(spelling.BSharp()) }},
		{"ReplaceWithPitch", "N/S", c4, func(n *note.Note) error { return n.ReplaceWithPitch(pt(spelling.D(), 4)) }},
		{"ReplaceWithSpelling", "N/S", c4, func(n *note.Note) error { return n.ReplaceWithSpelling(spelling.D()) }},
		{"Raise", "N/S", c4, func(n *note.Note) error { return n.Raise() }},
		{"Lower", "N/S", c4, func(n *note.Note) error { return n.Lower() }},
		{"RaiseBy", "N/S", c4, func(n *note.Note) error { return n.RaiseBy(2) }},
		{"LowerBy", "N/S", c4, func(n *note.Note) error { return n.LowerBy(2) }},
		{"RaiseByAlteration", "N/S", c4, func(n *note.Note) error { return n.RaiseByAlteration(alt(1, 1)) }},
		{"LowerByAlteration", "N/S", c4, func(n *note.Note) error { return n.LowerByAlteration(alt(1, 1)) }},
		{"SetAlteration", "N/S", c4, func(n *note.Note) error { return n.SetAlteration(alt(1, 1)) }},
		{"Naturalize", "N/S", c4, func(n *note.Note) error { return n.Naturalize() }},
		{"RespellWritten", "S", c4, func(n *note.Note) error { return n.RespellWritten(spelling.BSharp()) }},
		{"SetWrittenOctave", "S", c4, func(n *note.Note) error { return n.SetWrittenOctave(oct(5)) }},
		{"SetWrittenASPN", "S", c4, func(n *note.Note) error { return n.SetWrittenASPN(5) }},
		{"SetWrittenNumber", "S", c4, func(n *note.Note) error { return n.SetWrittenNumber(5) }},
		{"RaiseWritten", "S", b3, func(n *note.Note) error { return n.RaiseWritten() }},
		{"LowerWritten", "S", c4, func(n *note.Note) error { return n.LowerWritten() }},
		{"RaiseWrittenBy", "S", b3, func(n *note.Note) error { return n.RaiseWrittenBy(1) }},
		{"LowerWrittenBy", "S", c4, func(n *note.Note) error { return n.LowerWrittenBy(1) }},
		{"RaiseWrittenByAlteration", "S", b3, func(n *note.Note) error { return n.RaiseWrittenByAlteration(alt(1, 1)) }},
		{"LowerWrittenByAlteration", "S", c4, func(n *note.Note) error { return n.LowerWrittenByAlteration(alt(1, 1)) }},
		{"SetAlterationWritten", "S", b3, func(n *note.Note) error { return n.SetAlterationWritten(alt(1, 1)) }},
		{"NaturalizeWritten", "S", c4, func(n *note.Note) error { return n.NaturalizeWritten() }},
		{"Move", "S", c4, func(n *note.Note) error { return n.Move(1, system.Sharps()) }},
	}
	for _, o := range ops {
		t.Run(o.name+" bound", func(t *testing.T) {
			t.Parallel()
			src := newCounting(std())
			n := o.start(t, src)
			before := src.reads()
			require.NoError(t, o.call(&n))
			want := int64(1)
			if o.group == "N" {
				want = 0
			}
			assert.Equal(t, want, src.reads()-before, "group %s", o.group)
		})
		if o.group == "S" {
			continue
		}
		t.Run(o.name+" unbound", func(t *testing.T) {
			t.Parallel()
			n := o.start(t, nil)
			require.NoError(t, o.call(&n), "groups N and N/S work without a set")
		})
	}

	inForms := []struct {
		name string
		call func(n *note.Note, src settings.Source) error
	}{
		{"RespellWrittenIn", func(n *note.Note, src settings.Source) error { return n.RespellWrittenIn(src, spelling.BSharp()) }},
		{"SetWrittenOctaveIn", func(n *note.Note, src settings.Source) error { return n.SetWrittenOctaveIn(src, oct(5)) }},
		{"SetWrittenNumberIn", func(n *note.Note, src settings.Source) error { return n.SetWrittenNumberIn(src, 5) }},
		{"RaiseWrittenByAlterationIn", func(n *note.Note, src settings.Source) error {
			return n.RaiseWrittenByAlterationIn(src, alt(1, 1))
		}},
		{"SetAlterationWrittenIn", func(n *note.Note, src settings.Source) error { return n.SetAlterationWrittenIn(src, alt(1, 1)) }},
		{"NaturalizeWrittenIn", func(n *note.Note, src settings.Source) error { return n.NaturalizeWrittenIn(src) }},
		{"MoveIn", func(n *note.Note, src settings.Source) error { return n.MoveIn(src, 1, system.Sharps()) }},
	}
	for _, f := range inForms {
		t.Run(f.name+" two sources", func(t *testing.T) {
			t.Parallel()
			own, explicit := newCounting(std()), newCounting(std())
			n := mk(t, spelling.C(), 4, ct(0), own)
			before := own.reads()
			require.NoError(t, f.call(&n, explicit))
			assert.EqualValues(t, 1, own.reads()-before, "the own source is read once")
			assert.EqualValues(t, 1, explicit.reads(), "the explicit source is read once")
			assert.True(t, n.Source() == settings.Source(own), "the note stays bound to its own source")
		})
		t.Run(f.name+" one source", func(t *testing.T) {
			t.Parallel()
			own := newCounting(std())
			n := mk(t, spelling.C(), 4, ct(0), own)
			before := own.reads()
			require.NoError(t, f.call(&n, own))
			assert.EqualValues(t, 1, own.reads()-before, "one source is read once")
		})
		t.Run(f.name+" unbound", func(t *testing.T) {
			t.Parallel()
			explicit := newCounting(std())
			n := mk(t, spelling.C(), 4, ct(0), nil)
			require.NoError(t, f.call(&n, explicit))
			assert.EqualValues(t, 1, explicit.reads())
			assert.False(t, n.IsBound())
		})
	}
}

// TestMustReturnsTheReceiver checks every Must variant of a changing method:
// it returns the same pointer for a variable and for a *Note (architecture
// 6.6.1, 6.11).
func TestMustReturnsTheReceiver(t *testing.T) {
	t.Parallel()

	s := std()
	calls := map[string]func(n *note.Note) *note.Note{
		"MustRespell":                    func(n *note.Note) *note.Note { return n.MustRespell(spelling.BSharp()) },
		"MustRespellWritten":             func(n *note.Note) *note.Note { return n.MustRespellWritten(spelling.BSharp()) },
		"MustRespellWrittenIn":           func(n *note.Note) *note.Note { return n.MustRespellWrittenIn(s, spelling.BSharp()) },
		"MustSetOctave":                  func(n *note.Note) *note.Note { return n.MustSetOctave(oct(5)) },
		"MustSetASPN":                    func(n *note.Note) *note.Note { return n.MustSetASPN(5) },
		"MustSetWrittenOctave":           func(n *note.Note) *note.Note { return n.MustSetWrittenOctave(oct(5)) },
		"MustSetWrittenASPN":             func(n *note.Note) *note.Note { return n.MustSetWrittenASPN(5) },
		"MustSetWrittenNumber":           func(n *note.Note) *note.Note { return n.MustSetWrittenNumber(5) },
		"MustSetWrittenOctaveIn":         func(n *note.Note) *note.Note { return n.MustSetWrittenOctaveIn(s, oct(5)) },
		"MustSetWrittenNumberIn":         func(n *note.Note) *note.Note { return n.MustSetWrittenNumberIn(s, 5) },
		"MustRemoveOctave":               func(n *note.Note) *note.Note { return n.MustRemoveOctave() },
		"MustReplaceWithPitch":           func(n *note.Note) *note.Note { return n.MustReplaceWithPitch(pt(spelling.D(), 2)) },
		"MustReplaceWithSpelling":        func(n *note.Note) *note.Note { return n.MustReplaceWithSpelling(spelling.D()) },
		"MustSetCents":                   func(n *note.Note) *note.Note { return n.MustSetCents(ct(1)) },
		"MustRaise":                      func(n *note.Note) *note.Note { return n.MustRaise() },
		"MustLower":                      func(n *note.Note) *note.Note { return n.MustLower() },
		"MustRaiseBy":                    func(n *note.Note) *note.Note { return n.MustRaiseBy(1) },
		"MustLowerBy":                    func(n *note.Note) *note.Note { return n.MustLowerBy(1) },
		"MustRaiseByAlteration":          func(n *note.Note) *note.Note { return n.MustRaiseByAlteration(alt(1, 1)) },
		"MustLowerByAlteration":          func(n *note.Note) *note.Note { return n.MustLowerByAlteration(alt(1, 1)) },
		"MustSetAlteration":              func(n *note.Note) *note.Note { return n.MustSetAlteration(alt(1, 1)) },
		"MustRaiseWritten":               func(n *note.Note) *note.Note { return n.MustRaiseWritten() },
		"MustLowerWritten":               func(n *note.Note) *note.Note { return n.MustLowerWritten() },
		"MustRaiseWrittenBy":             func(n *note.Note) *note.Note { return n.MustRaiseWrittenBy(1) },
		"MustLowerWrittenBy":             func(n *note.Note) *note.Note { return n.MustLowerWrittenBy(1) },
		"MustRaiseWrittenByAlteration":   func(n *note.Note) *note.Note { return n.MustRaiseWrittenByAlteration(alt(1, 1)) },
		"MustLowerWrittenByAlteration":   func(n *note.Note) *note.Note { return n.MustLowerWrittenByAlteration(alt(1, 1)) },
		"MustSetAlterationWritten":       func(n *note.Note) *note.Note { return n.MustSetAlterationWritten(alt(1, 1)) },
		"MustRaiseWrittenByAlterationIn": func(n *note.Note) *note.Note { return n.MustRaiseWrittenByAlterationIn(s, alt(1, 1)) },
		"MustSetAlterationWrittenIn":     func(n *note.Note) *note.Note { return n.MustSetAlterationWrittenIn(s, alt(1, 1)) },
		"MustNaturalize":                 func(n *note.Note) *note.Note { return n.MustNaturalize() },
		"MustNaturalizeWritten":          func(n *note.Note) *note.Note { return n.MustNaturalizeWritten() },
		"MustNaturalizeWrittenIn":        func(n *note.Note) *note.Note { return n.MustNaturalizeWrittenIn(s) },
		"MustShift":                      func(n *note.Note) *note.Note { return n.MustShift(1) },
		"MustShiftDown":                  func(n *note.Note) *note.Note { return n.MustShiftDown(1) },
		"MustUp":                         func(n *note.Note) *note.Note { return n.MustUp() },
		"MustDown":                       func(n *note.Note) *note.Note { return n.MustDown() },
		"MustMove":                       func(n *note.Note) *note.Note { return n.MustMove(1, system.Sharps()) },
		"MustMoveIn":                     func(n *note.Note) *note.Note { return n.MustMoveIn(s, 1, system.Sharps()) },
		"MustSetSource":                  func(n *note.Note) *note.Note { return n.MustSetSource(set19()) },
		"MustSetSourceWritten":           func(n *note.Note) *note.Note { return n.MustSetSourceWritten(set19()) },
		"MustUnbind":                     func(n *note.Note) *note.Note { return n.MustUnbind() },
	}
	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			n := mk(t, spelling.C(), 4, ct(0), s)
			assert.Same(t, &n, call(&n), "variable")

			p := new(mk(t, spelling.C(), 4, ct(0), s))
			assert.Same(t, p, call(p), "*Note")
		})
	}
}

// TestChangeNilPointer checks the nil *Note for every changing method of
// architecture 6.6.2 and 6.7: the plain form returns E-1 and the Must variant
// panics with it (architecture 6.1).
func TestChangeNilPointer(t *testing.T) {
	t.Parallel()

	for _, mu := range mutators() {
		t.Run(mu.name, func(t *testing.T) {
			t.Parallel()
			var p *note.Note
			var err error
			if !assertNoPanic(t, func() { err = mu.op(p) }) {
				return
			}
			assertErr(t, err, note.ErrUnset)
			assertPanicErr(t, func() { mu.must(p) }, note.ErrUnset)
		})
	}
}

// TestChains covers K-42, K-45, K-46 and SE-9: chains of Must variants,
// a failure in the middle keeps the steps already made.
func TestChains(t *testing.T) {
	t.Parallel()

	t.Run("K-42", func(t *testing.T) {
		t.Parallel()
		s := std()
		n := mk(t, spelling.C(), 4, ct(30), s)
		got := n.MustRaise().MustUp().MustSetCents(ct(-10))
		assert.Same(t, &n, got)
		assertNote(t, wp(spelling.CSharp(), 5, ct(-10)).boundTo(s), n)
		f := n.MustFrequency()
		assert.InDelta(t, 551.1724, f, 0.00005)
		assert.Equal(t, std().Tuning().MustShiftedFrequency(pt(spelling.CSharp(), 5), ct(-10)), f)
	})

	t.Run("K-42 on a pointer and reading at the end", func(t *testing.T) {
		t.Parallel()
		p := new(mk(t, spelling.C(), 4, ct(30), std()))
		f := p.MustRaise().MustUp().MustSetCents(ct(-10)).MustFrequency()
		assert.InDelta(t, 551.1724, f, 0.00005)
	})

	t.Run("K-45 and SE-9", func(t *testing.T) {
		t.Parallel()
		k25 := setK25()
		n := mk(t, spelling.CSharp(), 4, ct(30), k25)
		assertPanicErr(t, func() { n.MustUp().MustRaise().MustSetCents(ct(5)) }, note.ErrInadmissible, system.ErrInadmissible)
		assertNote(t, wp(spelling.CSharp(), 5, ct(30)).boundTo(k25), n, "the first step stays, the third did not run")
	})

	t.Run("K-44 failure leaves the note", func(t *testing.T) {
		t.Parallel()
		k25 := setK25()
		n := mk(t, spelling.CSharp(), 4, ct(30), k25)
		assertErr(t, n.Raise(), note.ErrInadmissible, system.ErrInadmissible)
		assertNote(t, wp(spelling.CSharp(), 4, ct(30)).boundTo(k25), n)
	})

	t.Run("K-46 unset note stays unset", func(t *testing.T) {
		t.Parallel()
		var n note.Note
		assertErr(t, n.SetOctave(oct(4)), note.ErrUnset)
		assertErr(t, n.SetCents(ct(1)), note.ErrUnset)
		assertErr(t, n.Unbind(), note.ErrUnset)
		assert.True(t, n == note.Note{})
		assert.False(t, n.IsValid())
	})
}

// TestCopies covers K-43, G-8 and P-10: a copy by assignment is independent
// in both directions; the source is shared.
func TestCopies(t *testing.T) {
	t.Parallel()

	t.Run("K-43", func(t *testing.T) {
		t.Parallel()
		s := std()
		n := mk(t, spelling.C(), 4, ct(30), s)
		m := n
		m.MustRaise()
		assertNote(t, wp(spelling.CSharp(), 4, ct(30)).boundTo(s), m)
		assertNote(t, wp(spelling.C(), 4, ct(30)).boundTo(s), n)
		n.MustUp()
		assertNote(t, wp(spelling.C(), 5, ct(30)).boundTo(s), n)
		assertNote(t, wp(spelling.CSharp(), 4, ct(30)).boundTo(s), m)
		assert.True(t, n.SameSource(m))
	})

	t.Run("copy through a pointer", func(t *testing.T) {
		t.Parallel()
		p := new(mk(t, spelling.C(), 4, ct(0), nil))
		m := *p
		p.MustSetCents(ct(7))
		assert.True(t, m.Cents().Equal(ct(0)))
		m.MustShift(2)
		o, err := p.Octave()
		require.NoError(t, err)
		assert.True(t, o.Equal(oct(4)))
	})

	t.Run("hi := n.MustRaise() is n", func(t *testing.T) {
		t.Parallel()
		n := mk(t, spelling.C(), 4, ct(0), nil)
		hi := n.MustRaise()
		hi.MustUp()
		assertNote(t, wp(spelling.CSharp(), 5, ct(0)).free(), n)
	})

	t.Run("elements of a slice", func(t *testing.T) {
		t.Parallel()
		notes := []note.Note{mk(t, spelling.C(), 4, ct(0), nil), mk(t, spelling.D(), 4, ct(0), nil)}
		for _, x := range notes {
			x.MustRaise() // changes the copy only
		}
		assertNote(t, wp(spelling.C(), 4, ct(0)), notes[0])
		for i := range notes {
			notes[i].MustRaise()
		}
		assertNote(t, wp(spelling.CSharp(), 4, ct(0)), notes[0])
		assertNote(t, wp(spelling.DSharp(), 4, ct(0)), notes[1])
	})

	t.Run("element of a map", func(t *testing.T) {
		t.Parallel()
		m := map[string]note.Note{"x": mk(t, spelling.C(), 4, ct(0), nil)}
		v := m["x"]
		v.MustLower()
		assertNote(t, wp(spelling.C(), 4, ct(0)), m["x"])
		m["x"] = v
		assertNote(t, wp(spelling.CFlat(), 4, ct(0)), m["x"])
	})
}
