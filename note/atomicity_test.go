package note_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/octave/number"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

// Atomicity (P-10, KP-3; plan, Sh1 "Atomicity"): for every changing method
// and every reachable reason of failure the note after the failure is == to
// its copy taken before the call: the record, the cents and the source. The
// same holds for the Must variant after recover.

// mutator is a changing method with its Must variant.
type mutator struct {
	name  string
	group string // "N", "N/S", "S", "In" (explicit source), "Src" (OP-28, OP-29)
	op    func(n *note.Note) error
	must  func(n *note.Note)
}

// mutators lists every changing method of architecture 6.6.2 and 6.7 with
// arguments that are valid for C4 bound to the standard set.
func mutators() []mutator {
	s := std()
	return []mutator{
		{"Respell", "N/S", func(n *note.Note) error { return n.Respell(spelling.D()) }, func(n *note.Note) { n.MustRespell(spelling.D()) }},
		{"RespellWritten", "S", func(n *note.Note) error { return n.RespellWritten(spelling.D()) }, func(n *note.Note) { n.MustRespellWritten(spelling.D()) }},
		{"RespellWrittenIn", "In", func(n *note.Note) error { return n.RespellWrittenIn(s, spelling.D()) }, func(n *note.Note) { n.MustRespellWrittenIn(s, spelling.D()) }},
		{"SetOctave", "N", func(n *note.Note) error { return n.SetOctave(oct(2)) }, func(n *note.Note) { n.MustSetOctave(oct(2)) }},
		{"SetASPN", "N", func(n *note.Note) error { return n.SetASPN(2) }, func(n *note.Note) { n.MustSetASPN(2) }},
		{"SetWrittenOctave", "S", func(n *note.Note) error { return n.SetWrittenOctave(oct(2)) }, func(n *note.Note) { n.MustSetWrittenOctave(oct(2)) }},
		{"SetWrittenASPN", "S", func(n *note.Note) error { return n.SetWrittenASPN(2) }, func(n *note.Note) { n.MustSetWrittenASPN(2) }},
		{"SetWrittenNumber", "S", func(n *note.Note) error { return n.SetWrittenNumber(2) }, func(n *note.Note) { n.MustSetWrittenNumber(2) }},
		{"SetWrittenOctaveIn", "In", func(n *note.Note) error { return n.SetWrittenOctaveIn(s, oct(2)) }, func(n *note.Note) { n.MustSetWrittenOctaveIn(s, oct(2)) }},
		{"SetWrittenNumberIn", "In", func(n *note.Note) error { return n.SetWrittenNumberIn(s, 2) }, func(n *note.Note) { n.MustSetWrittenNumberIn(s, 2) }},
		{"RemoveOctave", "N", func(n *note.Note) error { return n.RemoveOctave() }, func(n *note.Note) { n.MustRemoveOctave() }},
		{"ReplaceWithPitch", "N/S", func(n *note.Note) error { return n.ReplaceWithPitch(pt(spelling.D(), 2)) }, func(n *note.Note) { n.MustReplaceWithPitch(pt(spelling.D(), 2)) }},
		{"ReplaceWithSpelling", "N/S", func(n *note.Note) error { return n.ReplaceWithSpelling(spelling.D()) }, func(n *note.Note) { n.MustReplaceWithSpelling(spelling.D()) }},
		{"SetCents", "N", func(n *note.Note) error { return n.SetCents(ct(3)) }, func(n *note.Note) { n.MustSetCents(ct(3)) }},
		{"Raise", "N/S", func(n *note.Note) error { return n.Raise() }, func(n *note.Note) { n.MustRaise() }},
		{"Lower", "N/S", func(n *note.Note) error { return n.Lower() }, func(n *note.Note) { n.MustLower() }},
		{"RaiseBy", "N/S", func(n *note.Note) error { return n.RaiseBy(1) }, func(n *note.Note) { n.MustRaiseBy(1) }},
		{"LowerBy", "N/S", func(n *note.Note) error { return n.LowerBy(1) }, func(n *note.Note) { n.MustLowerBy(1) }},
		{"RaiseByAlteration", "N/S", func(n *note.Note) error { return n.RaiseByAlteration(alt(1, 1)) }, func(n *note.Note) { n.MustRaiseByAlteration(alt(1, 1)) }},
		{"LowerByAlteration", "N/S", func(n *note.Note) error { return n.LowerByAlteration(alt(1, 1)) }, func(n *note.Note) { n.MustLowerByAlteration(alt(1, 1)) }},
		{"SetAlteration", "N/S", func(n *note.Note) error { return n.SetAlteration(alt(1, 1)) }, func(n *note.Note) { n.MustSetAlteration(alt(1, 1)) }},
		{"RaiseWritten", "S", func(n *note.Note) error { return n.RaiseWritten() }, func(n *note.Note) { n.MustRaiseWritten() }},
		{"LowerWritten", "S", func(n *note.Note) error { return n.LowerWritten() }, func(n *note.Note) { n.MustLowerWritten() }},
		{"RaiseWrittenBy", "S", func(n *note.Note) error { return n.RaiseWrittenBy(1) }, func(n *note.Note) { n.MustRaiseWrittenBy(1) }},
		{"LowerWrittenBy", "S", func(n *note.Note) error { return n.LowerWrittenBy(1) }, func(n *note.Note) { n.MustLowerWrittenBy(1) }},
		{"RaiseWrittenByAlteration", "S", func(n *note.Note) error { return n.RaiseWrittenByAlteration(alt(1, 1)) }, func(n *note.Note) { n.MustRaiseWrittenByAlteration(alt(1, 1)) }},
		{"LowerWrittenByAlteration", "S", func(n *note.Note) error { return n.LowerWrittenByAlteration(alt(1, 1)) }, func(n *note.Note) { n.MustLowerWrittenByAlteration(alt(1, 1)) }},
		{"SetAlterationWritten", "S", func(n *note.Note) error { return n.SetAlterationWritten(alt(1, 1)) }, func(n *note.Note) { n.MustSetAlterationWritten(alt(1, 1)) }},
		{"RaiseWrittenByAlterationIn", "In", func(n *note.Note) error { return n.RaiseWrittenByAlterationIn(s, alt(1, 1)) }, func(n *note.Note) { n.MustRaiseWrittenByAlterationIn(s, alt(1, 1)) }},
		{"SetAlterationWrittenIn", "In", func(n *note.Note) error { return n.SetAlterationWrittenIn(s, alt(1, 1)) }, func(n *note.Note) { n.MustSetAlterationWrittenIn(s, alt(1, 1)) }},
		{"Naturalize", "N/S", func(n *note.Note) error { return n.Naturalize() }, func(n *note.Note) { n.MustNaturalize() }},
		{"NaturalizeWritten", "S", func(n *note.Note) error { return n.NaturalizeWritten() }, func(n *note.Note) { n.MustNaturalizeWritten() }},
		{"NaturalizeWrittenIn", "In", func(n *note.Note) error { return n.NaturalizeWrittenIn(s) }, func(n *note.Note) { n.MustNaturalizeWrittenIn(s) }},
		{"Shift", "N", func(n *note.Note) error { return n.Shift(1) }, func(n *note.Note) { n.MustShift(1) }},
		{"ShiftDown", "N", func(n *note.Note) error { return n.ShiftDown(1) }, func(n *note.Note) { n.MustShiftDown(1) }},
		{"Up", "N", func(n *note.Note) error { return n.Up() }, func(n *note.Note) { n.MustUp() }},
		{"Down", "N", func(n *note.Note) error { return n.Down() }, func(n *note.Note) { n.MustDown() }},
		{"Move", "S", func(n *note.Note) error { return n.Move(1, system.Sharps()) }, func(n *note.Note) { n.MustMove(1, system.Sharps()) }},
		{"MoveIn", "In", func(n *note.Note) error { return n.MoveIn(s, 1, system.Sharps()) }, func(n *note.Note) { n.MustMoveIn(s, 1, system.Sharps()) }},
		{"SetSource", "Src", func(n *note.Note) error { return n.SetSource(s) }, func(n *note.Note) { n.MustSetSource(s) }},
		{"SetSourceWritten", "Src", func(n *note.Note) error { return n.SetSourceWritten(s) }, func(n *note.Note) { n.MustSetSourceWritten(s) }},
		{"Unbind", "N", func(n *note.Note) error { return n.Unbind() }, func(n *note.Note) { n.MustUnbind() }},
	}
}

// checkAtomic calls op and must on separate copies of start and checks that
// both fail with the expected error and leave the note == to its snapshot.
func checkAtomic(t *testing.T, start note.Note, op func(*note.Note) error, must func(*note.Note), want error, extra ...error) {
	t.Helper()
	n, m := start, start
	assertErr(t, op(&n), want, extra...)
	assertSame(t, start, n, "plain form")
	assertPanicErr(t, func() { must(&m) }, want, extra...)
	assertSame(t, start, m, "Must form")
}

// TestAtomicityByGroup runs the reasons that apply to whole groups: E-1 for
// every method, E-3 for every method that reads the own source, E-2 for every
// method of group "S" on an unbound note with octave.
func TestAtomicityByGroup(t *testing.T) {
	t.Parallel()

	for _, mu := range mutators() {
		t.Run(mu.name+" E-1 unset note", func(t *testing.T) {
			t.Parallel()
			checkAtomic(t, note.Note{}, mu.op, mu.must, note.ErrUnset)
		})
		if mu.group != "N" {
			t.Run(mu.name+" E-3 own source returns unset settings", func(t *testing.T) {
				t.Parallel()
				checkAtomic(t, onStale(spelling.C(), 4)(t), mu.op, mu.must, note.ErrUnsetSettings, settings.ErrUnset)
			})
		}
		if mu.group == "S" {
			t.Run(mu.name+" E-2 unbound with octave", func(t *testing.T) {
				t.Parallel()
				checkAtomic(t, mk(t, spelling.C(), 4, ct(5), nil), mu.op, mu.must, note.ErrNoSource)
			})
		}
	}
}

// TestAtomicityByReason runs the reasons that depend on the method: E-1 of
// arguments, E-4, E-5, E-7, E-8 and the reasons of dependencies.
func TestAtomicityByReason(t *testing.T) {
	t.Parallel()

	s, k25, s24 := std(), setK25(), set24()
	unset := &unsetSource{}
	type row struct {
		name  string
		start func(*testing.T) note.Note
		op    func(*note.Note) error
		must  func(*note.Note)
		err   error
		extra []error
	}
	rows := []row{
		// E-1 of an argument.
		{"Respell unset name", onPitch(spelling.C(), 4, ct(5), s), func(n *note.Note) error { return n.Respell(spelling.Spelling{}) }, func(n *note.Note) { n.MustRespell(spelling.Spelling{}) }, note.ErrUnset, nil},
		{"RespellWrittenIn nil source", onPitch(spelling.C(), 4, ct(5), s), func(n *note.Note) error { return n.RespellWrittenIn(nil, spelling.D()) }, func(n *note.Note) { n.MustRespellWrittenIn(nil, spelling.D()) }, note.ErrUnset, nil},
		{"SetWrittenOctaveIn nil source", onPitch(spelling.C(), 4, ct(5), nil), func(n *note.Note) error { return n.SetWrittenOctaveIn(nil, oct(3)) }, func(n *note.Note) { n.MustSetWrittenOctaveIn(nil, oct(3)) }, note.ErrUnset, nil},
		{"SetWrittenNumberIn nil source", onPitch(spelling.C(), 4, ct(5), nil), func(n *note.Note) error { return n.SetWrittenNumberIn(nil, 3) }, func(n *note.Note) { n.MustSetWrittenNumberIn(nil, 3) }, note.ErrUnset, nil},
		{"RaiseWrittenByAlterationIn nil source", onPitch(spelling.C(), 4, ct(5), nil), func(n *note.Note) error { return n.RaiseWrittenByAlterationIn(nil, alt(1, 1)) }, func(n *note.Note) { n.MustRaiseWrittenByAlterationIn(nil, alt(1, 1)) }, note.ErrUnset, nil},
		{"SetAlterationWrittenIn nil source", onPitch(spelling.C(), 4, ct(5), nil), func(n *note.Note) error { return n.SetAlterationWrittenIn(nil, alt(1, 1)) }, func(n *note.Note) { n.MustSetAlterationWrittenIn(nil, alt(1, 1)) }, note.ErrUnset, nil},
		{"NaturalizeWrittenIn nil source", onPitch(spelling.C(), 4, ct(5), nil), func(n *note.Note) error { return n.NaturalizeWrittenIn(nil) }, func(n *note.Note) { n.MustNaturalizeWrittenIn(nil) }, note.ErrUnset, nil},
		{"MoveIn unset rule", onPitch(spelling.C(), 4, ct(5), nil), func(n *note.Note) error { return n.MoveIn(s, 1, system.SpellingRule{}) }, func(n *note.Note) { n.MustMoveIn(s, 1, system.SpellingRule{}) }, note.ErrUnset, nil},
		{"SetOctave unset octave", onPitch(spelling.C(), 4, ct(5), s), func(n *note.Note) error { return n.SetOctave(octave.Octave{}) }, func(n *note.Note) { n.MustSetOctave(octave.Octave{}) }, note.ErrUnset, nil},
		{"SetWrittenOctave unset octave", onPitch(spelling.C(), 4, ct(5), s), func(n *note.Note) error { return n.SetWrittenOctave(octave.Octave{}) }, func(n *note.Note) { n.MustSetWrittenOctave(octave.Octave{}) }, note.ErrUnset, nil},
		{"ReplaceWithPitch unset pitch", onPitch(spelling.C(), 4, ct(5), s), func(n *note.Note) error { return n.ReplaceWithPitch(pitch.Pitch{}) }, func(n *note.Note) { n.MustReplaceWithPitch(pitch.Pitch{}) }, note.ErrUnset, nil},
		{"Move unset rule", onPitch(spelling.C(), 4, ct(5), s), func(n *note.Note) error { return n.Move(1, system.SpellingRule{}) }, func(n *note.Note) { n.MustMove(1, system.SpellingRule{}) }, note.ErrUnset, nil},
		{"SetSource nil source", onPitch(spelling.C(), 4, ct(5), s), func(n *note.Note) error { return n.SetSource(nil) }, func(n *note.Note) { n.MustSetSource(nil) }, note.ErrUnset, nil},
		{"SetSourceWritten nil source", onPitch(spelling.C(), 4, ct(5), s), func(n *note.Note) error { return n.SetSourceWritten(nil) }, func(n *note.Note) { n.MustSetSourceWritten(nil) }, note.ErrUnset, nil},

		// E-3 of an explicit source.
		{"RespellWrittenIn explicit unset", onPitch(spelling.C(), 4, ct(5), nil), func(n *note.Note) error { return n.RespellWrittenIn(unset, spelling.D()) }, func(n *note.Note) { n.MustRespellWrittenIn(unset, spelling.D()) }, note.ErrUnsetSettings, []error{settings.ErrUnset}},
		{"SetSource new unset", onPitch(spelling.C(), 4, ct(5), s), func(n *note.Note) error { return n.SetSource(unset) }, func(n *note.Note) { n.MustSetSource(unset) }, note.ErrUnsetSettings, []error{settings.ErrUnset}},
		{"MoveIn explicit unset", onPitch(spelling.C(), 4, ct(5), s), func(n *note.Note) error { return n.MoveIn(unset, 1, system.Sharps()) }, func(n *note.Note) { n.MustMoveIn(unset, 1, system.Sharps()) }, note.ErrUnsetSettings, []error{settings.ErrUnset}},

		// E-4.
		{"Shift without octave", onName(spelling.C(), ct(5), s), func(n *note.Note) error { return n.Shift(1) }, func(n *note.Note) { n.MustShift(1) }, note.ErrNoOctave, nil},
		{"ShiftDown without octave", onName(spelling.C(), ct(5), s), func(n *note.Note) error { return n.ShiftDown(1) }, func(n *note.Note) { n.MustShiftDown(1) }, note.ErrNoOctave, nil},
		{"Up without octave", onName(spelling.C(), ct(5), nil), func(n *note.Note) error { return n.Up() }, func(n *note.Note) { n.MustUp() }, note.ErrNoOctave, nil},
		{"Down without octave", onName(spelling.C(), ct(5), nil), func(n *note.Note) error { return n.Down() }, func(n *note.Note) { n.MustDown() }, note.ErrNoOctave, nil},

		// E-5.
		{"Respell inadmissible", onPitch(spelling.C(), 4, ct(5), k25), func(n *note.Note) error { return n.Respell(spelling.CDoubleSharp()) }, func(n *note.Note) { n.MustRespell(spelling.CDoubleSharp()) }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"RespellWritten inadmissible input", onInadmissible(), func(n *note.Note) error { return n.RespellWritten(spelling.D()) }, func(n *note.Note) { n.MustRespellWritten(spelling.D()) }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"RespellWrittenIn inadmissible in own", onPitch(spelling.C(), 4, ct(5), s), func(n *note.Note) error { return n.RespellWrittenIn(s24, cHalf()) }, func(n *note.Note) { n.MustRespellWrittenIn(s24, cHalf()) }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"SetWrittenOctave inadmissible", onInadmissible(), func(n *note.Note) error { return n.SetWrittenOctave(oct(3)) }, func(n *note.Note) { n.MustSetWrittenOctave(oct(3)) }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"SetWrittenASPN inadmissible", onInadmissible(), func(n *note.Note) error { return n.SetWrittenASPN(3) }, func(n *note.Note) { n.MustSetWrittenASPN(3) }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"SetWrittenNumber inadmissible", onInadmissible(), func(n *note.Note) error { return n.SetWrittenNumber(3) }, func(n *note.Note) { n.MustSetWrittenNumber(3) }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"SetWrittenOctaveIn inadmissible", onPitch(spelling.CDoubleSharp(), 4, ct(5), nil), func(n *note.Note) error { return n.SetWrittenOctaveIn(k25, oct(3)) }, func(n *note.Note) { n.MustSetWrittenOctaveIn(k25, oct(3)) }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"SetWrittenNumberIn inadmissible", onPitch(spelling.CDoubleSharp(), 4, ct(5), nil), func(n *note.Note) error { return n.SetWrittenNumberIn(k25, 3) }, func(n *note.Note) { n.MustSetWrittenNumberIn(k25, 3) }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"ReplaceWithPitch inadmissible", onPitch(spelling.C(), 4, ct(5), k25), func(n *note.Note) error { return n.ReplaceWithPitch(pt(spelling.CDoubleSharp(), 1)) }, func(n *note.Note) { n.MustReplaceWithPitch(pt(spelling.CDoubleSharp(), 1)) }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"ReplaceWithSpelling inadmissible", onPitch(spelling.C(), 4, ct(5), k25), func(n *note.Note) error { return n.ReplaceWithSpelling(spelling.CDoubleSharp()) }, func(n *note.Note) { n.MustReplaceWithSpelling(spelling.CDoubleSharp()) }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"Raise inadmissible", onPitch(spelling.CSharp(), 4, ct(5), k25), func(n *note.Note) error { return n.Raise() }, func(n *note.Note) { n.MustRaise() }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"Lower inadmissible", onPitch(spelling.CFlat(), 4, ct(5), k25), func(n *note.Note) error { return n.Lower() }, func(n *note.Note) { n.MustLower() }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"RaiseBy inadmissible", onPitch(spelling.C(), 4, ct(5), k25), func(n *note.Note) error { return n.RaiseBy(2) }, func(n *note.Note) { n.MustRaiseBy(2) }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"LowerBy inadmissible", onPitch(spelling.C(), 4, ct(5), k25), func(n *note.Note) error { return n.LowerBy(2) }, func(n *note.Note) { n.MustLowerBy(2) }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"RaiseByAlteration inadmissible", onPitch(spelling.C(), 4, ct(5), s), func(n *note.Note) error { return n.RaiseByAlteration(alt(1, 2)) }, func(n *note.Note) { n.MustRaiseByAlteration(alt(1, 2)) }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"LowerByAlteration inadmissible", onPitch(spelling.C(), 4, ct(5), s), func(n *note.Note) error { return n.LowerByAlteration(alt(1, 2)) }, func(n *note.Note) { n.MustLowerByAlteration(alt(1, 2)) }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"SetAlteration inadmissible", onPitch(spelling.C(), 4, ct(5), k25), func(n *note.Note) error { return n.SetAlteration(alt(2, 1)) }, func(n *note.Note) { n.MustSetAlteration(alt(2, 1)) }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"RaiseWritten inadmissible result", onPitch(spelling.CSharp(), 4, ct(5), k25), func(n *note.Note) error { return n.RaiseWritten() }, func(n *note.Note) { n.MustRaiseWritten() }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"LowerWritten inadmissible result", onPitch(spelling.CFlat(), 4, ct(5), k25), func(n *note.Note) error { return n.LowerWritten() }, func(n *note.Note) { n.MustLowerWritten() }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"RaiseWrittenBy inadmissible", onPitch(spelling.C(), 4, ct(5), k25), func(n *note.Note) error { return n.RaiseWrittenBy(2) }, func(n *note.Note) { n.MustRaiseWrittenBy(2) }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"LowerWrittenBy inadmissible", onPitch(spelling.C(), 4, ct(5), k25), func(n *note.Note) error { return n.LowerWrittenBy(2) }, func(n *note.Note) { n.MustLowerWrittenBy(2) }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"RaiseWrittenByAlteration inadmissible", onPitch(spelling.C(), 4, ct(5), s), func(n *note.Note) error { return n.RaiseWrittenByAlteration(alt(1, 2)) }, func(n *note.Note) { n.MustRaiseWrittenByAlteration(alt(1, 2)) }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"LowerWrittenByAlteration inadmissible", onPitch(spelling.C(), 4, ct(5), s), func(n *note.Note) error { return n.LowerWrittenByAlteration(alt(1, 2)) }, func(n *note.Note) { n.MustLowerWrittenByAlteration(alt(1, 2)) }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"SetAlterationWritten inadmissible", onPitch(spelling.C(), 4, ct(5), k25), func(n *note.Note) error { return n.SetAlterationWritten(alt(2, 1)) }, func(n *note.Note) { n.MustSetAlterationWritten(alt(2, 1)) }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"RaiseWrittenByAlterationIn inadmissible in own", onPitch(spelling.C(), 4, ct(5), s), func(n *note.Note) error { return n.RaiseWrittenByAlterationIn(s24, alt(1, 2)) }, func(n *note.Note) { n.MustRaiseWrittenByAlterationIn(s24, alt(1, 2)) }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"SetAlterationWrittenIn inadmissible in own", onPitch(spelling.C(), 4, ct(5), s), func(n *note.Note) error { return n.SetAlterationWrittenIn(s24, alt(1, 2)) }, func(n *note.Note) { n.MustSetAlterationWrittenIn(s24, alt(1, 2)) }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"NaturalizeWritten inadmissible input", onInadmissible(), func(n *note.Note) error { return n.NaturalizeWritten() }, func(n *note.Note) { n.MustNaturalizeWritten() }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"NaturalizeWrittenIn inadmissible input", onPitch(spelling.CDoubleSharp(), 4, ct(5), nil), func(n *note.Note) error { return n.NaturalizeWrittenIn(k25) }, func(n *note.Note) { n.MustNaturalizeWrittenIn(k25) }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"Move inadmissible input", onInadmissible(), func(n *note.Note) error { return n.Move(1, system.Sharps()) }, func(n *note.Note) { n.MustMove(1, system.Sharps()) }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"MoveIn inadmissible in own", onPitch(spelling.C(), 4, ct(5), s), func(n *note.Note) error { return n.MoveIn(s24, 1, system.Sharps()) }, func(n *note.Note) { n.MustMoveIn(s24, 1, system.Sharps()) }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"RespellWrittenIn input inadmissible in the explicit set", onPitch(spelling.CDoubleSharp(), 4, ct(5), nil), func(n *note.Note) error { return n.RespellWrittenIn(k25, spelling.D()) }, func(n *note.Note) { n.MustRespellWrittenIn(k25, spelling.D()) }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"RaiseWrittenByAlterationIn input inadmissible in the explicit set", onPitch(spelling.CDoubleSharp(), 4, ct(5), nil), func(n *note.Note) error { return n.RaiseWrittenByAlterationIn(k25, alt(-1, 1)) }, func(n *note.Note) { n.MustRaiseWrittenByAlterationIn(k25, alt(-1, 1)) }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"SetAlterationWrittenIn input inadmissible in the explicit set", onPitch(spelling.CDoubleSharp(), 4, ct(5), nil), func(n *note.Note) error { return n.SetAlterationWrittenIn(k25, alt(0, 1)) }, func(n *note.Note) { n.MustSetAlterationWrittenIn(k25, alt(0, 1)) }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"MoveIn input inadmissible in the explicit set", onPitch(spelling.CDoubleSharp(), 4, ct(5), nil), func(n *note.Note) error { return n.MoveIn(k25, 1, system.Sharps()) }, func(n *note.Note) { n.MustMoveIn(k25, 1, system.Sharps()) }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"SetWrittenOctaveIn name inadmissible in the own set", onInadmissible(), func(n *note.Note) error { return n.SetWrittenOctaveIn(s, oct(3)) }, func(n *note.Note) { n.MustSetWrittenOctaveIn(s, oct(3)) }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"SetWrittenNumberIn name inadmissible in the own set", onInadmissible(), func(n *note.Note) error { return n.SetWrittenNumberIn(s, 3) }, func(n *note.Note) { n.MustSetWrittenNumberIn(s, 3) }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"SetSource inadmissible in the new set", onPitch(cHalf(), 4, ct(5), s24), func(n *note.Note) error { return n.SetSource(s) }, func(n *note.Note) { n.MustSetSource(s) }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"SetSourceWritten inadmissible in the new set", onPitch(cHalf(), 4, ct(5), s24), func(n *note.Note) error { return n.SetSourceWritten(s) }, func(n *note.Note) { n.MustSetSourceWritten(s) }, note.ErrInadmissible, []error{system.ErrInadmissible}},

		// E-7.
		{"SetSource loses expressibility", onPitch(spelling.ADoubleFlat(), 4, ct(5), s), func(n *note.Note) error { return n.SetSource(stdNotation("german-am")) }, func(n *note.Note) { n.MustSetSource(stdNotation("german-am")) }, note.ErrInexpressible, nil},
		{"SetSourceWritten loses expressibility", onPitch(spelling.CSharp(), 4, ct(5), s), func(n *note.Note) error { return n.SetSourceWritten(stdNotation("romance-italian")) }, func(n *note.Note) { n.MustSetSourceWritten(stdNotation("romance-italian")) }, note.ErrInexpressible, nil},

		// E-8 and the reasons of dependencies.
		{"RespellWritten Octave out of range", onPitch(spelling.C(), math.MaxInt, ct(5), s), func(n *note.Note) error { return n.RespellWritten(spelling.BSharp()) }, func(n *note.Note) { n.MustRespellWritten(spelling.BSharp()) }, note.ErrOutOfRange, []error{system.ErrOutOfRange}},
		{"SetWrittenASPN Octave out of range", onPitch(spelling.BSharp(), 4, ct(5), s), func(n *note.Note) error { return n.SetWrittenASPN(math.MaxInt) }, func(n *note.Note) { n.MustSetWrittenASPN(math.MaxInt) }, note.ErrOutOfRange, []error{system.ErrOutOfRange}},
		{"SetWrittenNumber skipped", onPitch(spelling.C(), 4, ct(5), stdFB()), func(n *note.Note) error { return n.SetWrittenNumber(0) }, func(n *note.Note) { n.MustSetWrittenNumber(0) }, nil, []error{number.ErrSkipped}},
		{"SetWrittenNumberIn out of range", onPitch(spelling.C(), 4, ct(5), nil), func(n *note.Note) error { return n.SetWrittenNumberIn(stdFB(), math.MaxInt) }, func(n *note.Note) { n.MustSetWrittenNumberIn(stdFB(), math.MaxInt) }, note.ErrOutOfRange, []error{number.ErrOutOfRange}},
		{"RaiseBy alteration out of range", onPitch(spelling.C(), 4, ct(5), s), func(n *note.Note) error { return n.RaiseBy(1 << 25) }, func(n *note.Note) { n.MustRaiseBy(1 << 25) }, note.ErrOutOfRange, nil},
		{"LowerBy alteration out of range", onPitch(spelling.C(), 4, ct(5), nil), func(n *note.Note) error { return n.LowerBy(math.MinInt) }, func(n *note.Note) { n.MustLowerBy(math.MinInt) }, note.ErrOutOfRange, nil},
		{"RaiseWritten Octave out of range", onPitch(spelling.B(), math.MaxInt, ct(5), s), func(n *note.Note) error { return n.RaiseWritten() }, func(n *note.Note) { n.MustRaiseWritten() }, note.ErrOutOfRange, []error{system.ErrOutOfRange}},
		{"LowerWritten Octave out of range", onPitch(spelling.C(), math.MinInt, ct(5), s), func(n *note.Note) error { return n.LowerWritten() }, func(n *note.Note) { n.MustLowerWritten() }, note.ErrOutOfRange, []error{system.ErrOutOfRange}},
		{"NaturalizeWritten Octave out of range", onPitch(spelling.BSharp(), math.MinInt, ct(5), s), func(n *note.Note) error { return n.NaturalizeWritten() }, func(n *note.Note) { n.MustNaturalizeWritten() }, note.ErrOutOfRange, []error{system.ErrOutOfRange}},
		{"Shift out of range", onPitch(spelling.C(), math.MaxInt, ct(5), nil), func(n *note.Note) error { return n.Shift(1) }, func(n *note.Note) { n.MustShift(1) }, note.ErrOutOfRange, nil},
		{"ShiftDown out of range", onPitch(spelling.C(), math.MinInt, ct(5), nil), func(n *note.Note) error { return n.ShiftDown(1) }, func(n *note.Note) { n.MustShiftDown(1) }, note.ErrOutOfRange, nil},
		{"Move out of range", onPitch(spelling.C(), math.MaxInt, ct(5), s), func(n *note.Note) error { return n.Move(12, system.Sharps()) }, func(n *note.Note) { n.MustMove(12, system.Sharps()) }, note.ErrOutOfRange, []error{system.ErrOutOfRange}},
		{"Move bad rule", onPitch(spelling.C(), 4, ct(5), s), func(n *note.Note) error { return n.Move(1, badRule()) }, func(n *note.Note) { n.MustMove(1, badRule()) }, nil, []error{system.ErrBadSpelling}},
		{"MoveIn bad rule", onPitch(spelling.C(), 4, ct(5), nil), func(n *note.Note) error { return n.MoveIn(s, 1, badRule()) }, func(n *note.Note) { n.MustMoveIn(s, 1, badRule()) }, nil, []error{system.ErrBadSpelling}},
		{"SetSourceWritten written octave out of range", onPitch(spelling.BSharp(), math.MinInt, ct(5), s), func(n *note.Note) error { return n.SetSourceWritten(set19()) }, func(n *note.Note) { n.MustSetSourceWritten(set19()) }, note.ErrOutOfRange, []error{system.ErrOutOfRange}},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			t.Parallel()
			checkAtomic(t, r.start(t), r.op, r.must, r.err, r.extra...)
		})
	}
}
