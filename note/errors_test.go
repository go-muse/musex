package note_test

import (
	"errors"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
)

// Errors (architecture 6.2): eight distinct sentinels that do not wrap each
// other; the order of categories over all inputs (6.2.3). The reasons of each
// operation are checked next to it; assertErr checks everywhere that an error
// matches at most one sentinel of note.

func TestSentinels(t *testing.T) {
	t.Parallel()

	for i, a := range sentinels {
		assert.NotNil(t, a)
		assert.NotEmpty(t, a.Error())
		for j, b := range sentinels {
			if i == j {
				continue
			}
			assert.NotSame(t, a, b)
			assert.False(t, errors.Is(a, b), "%v must not match %v", a, b)
		}
	}
}

// TestCategoryOrder checks that each category is checked over all inputs
// before the next one (6.2.3): the receiver, the note arguments in order,
// then the other arguments.
func TestCategoryOrder(t *testing.T) {
	t.Parallel()

	s := std()
	c4 := mk(t, spelling.C(), 4, ct(0), s)
	free := mk(t, spelling.C(), 4, ct(0), nil)
	noOct := mkName(t, spelling.C(), ct(0), nil)
	stale := onStale(spelling.D(), 4)(t)
	inadm, _ := inadmissibleNote(t)
	s19note := mk(t, spelling.C(), 4, ct(0), set19())

	cases := []struct {
		name  string
		call  func() error
		err   error
		extra []error
	}{
		{"E-1 of an argument before E-4 of the receiver", func() error { _, err := noOct.Distance(note.Note{}); return err }, note.ErrUnset, nil},
		{"E-1 of the explicit source before E-4", func() error { _, err := noOct.FrequencyIn(nil); return err }, note.ErrUnset, nil},
		{"E-1 of the tuning before E-4", func() error { _, err := noOct.StepInTuning(nil); return err }, note.ErrUnset, nil},
		{"E-4 of an argument before E-2 of the receiver", func() error { _, err := free.CompareSound(noOct); return err }, note.ErrNoOctave, nil},
		{"E-2 of the receiver before E-3 of an argument", func() error { _, err := free.CompareSound(stale); return err }, note.ErrNoSource, nil},
		{"E-2 of an argument before E-3 of the receiver", func() error { _, err := stale.ComparePosition(free); return err }, note.ErrNoSource, nil},
		{"E-3 of an argument before E-5 of the receiver", func() error { _, err := inadm.ComparePosition(stale); return err }, note.ErrUnsetSettings, []error{settings.ErrUnset}},
		{"E-5 of an argument before E-6", func() error { _, err := s19note.Equivalent(inadm); return err }, note.ErrInadmissible, []error{system.ErrInadmissible}},
		{"E-4 before E-2 for one note", func() error { _, err := noOct.Frequency(); return err }, note.ErrNoOctave, nil},
		{"E-1 of the note before E-1 of an argument (both E-1)", func() error {
			_, err := note.Note{}.ComparePosition(note.Note{})
			return err
		}, note.ErrUnset, nil},
		{"E-6 for one bound and one bound in another system", func() error { _, err := c4.Distance(s19note); return err }, note.ErrDifferentSystems, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			assertErr(t, c.call(), c.err, c.extra...)
		})
	}
}

// TestOwnUnsetSettingsBeforeExplicitInadmissible checks the order of 6.2.3
// for the In forms that read the own source of a bound note (6.3, 6.9): E-3
// of the own source comes before E-5 of the input name in the explicit set.
// The note is C𝄪4 whose own source now gives the unset set; C𝄪 is not
// admissible in the explicit set of K-25.
func TestOwnUnsetSettingsBeforeExplicitInadmissible(t *testing.T) {
	t.Parallel()

	k25 := setK25()
	to := mk(t, spelling.E(), 4, ct(0), nil)
	cases := []struct {
		name string
		op   func(n *note.Note) error
		must func(n *note.Note)
	}{
		{"WalkIn",
			func(n *note.Note) error { _, err := n.WalkIn(k25, to, system.Sharps()); return err },
			func(n *note.Note) { n.MustWalkIn(k25, to, system.Sharps()) }},
		{"EnharmonicsIn",
			func(n *note.Note) error { _, err := n.EnharmonicsIn(k25, alteration.DoubleSharp()); return err },
			func(n *note.Note) { n.MustEnharmonicsIn(k25, alteration.DoubleSharp()) }},
		{"RespellWrittenIn",
			func(n *note.Note) error { return n.RespellWrittenIn(k25, spelling.D()) },
			func(n *note.Note) { n.MustRespellWrittenIn(k25, spelling.D()) }},
		{"RaiseWrittenByAlterationIn",
			func(n *note.Note) error { return n.RaiseWrittenByAlterationIn(k25, alt(-1, 1)) },
			func(n *note.Note) { n.MustRaiseWrittenByAlterationIn(k25, alt(-1, 1)) }},
		{"SetAlterationWrittenIn",
			func(n *note.Note) error { return n.SetAlterationWrittenIn(k25, alt(0, 1)) },
			func(n *note.Note) { n.MustSetAlterationWrittenIn(k25, alt(0, 1)) }},
		{"SetWrittenOctaveIn",
			func(n *note.Note) error { return n.SetWrittenOctaveIn(k25, oct(3)) },
			func(n *note.Note) { n.MustSetWrittenOctaveIn(k25, oct(3)) }},
		{"SetWrittenNumberIn",
			func(n *note.Note) error { return n.SetWrittenNumberIn(k25, 3) },
			func(n *note.Note) { n.MustSetWrittenNumberIn(k25, 3) }},
		{"NaturalizeWrittenIn",
			func(n *note.Note) error { return n.NaturalizeWrittenIn(k25) },
			func(n *note.Note) { n.MustNaturalizeWrittenIn(k25) }},
		{"MoveIn",
			func(n *note.Note) error { return n.MoveIn(k25, 1, system.Sharps()) },
			func(n *note.Note) { n.MustMoveIn(k25, 1, system.Sharps()) }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			start := onStale(spelling.CDoubleSharp(), 4)(t)
			n, m := start, start
			assertErr(t, c.op(&n), note.ErrUnsetSettings, settings.ErrUnset)
			assertSame(t, start, n)
			assertPanicErr(t, func() { c.must(&m) }, note.ErrUnsetSettings, settings.ErrUnset)
			assertSame(t, start, m)
		})
	}
}
