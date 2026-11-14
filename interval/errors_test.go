package interval_test

import (
	"errors"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
)

// T-7, errors (F-3; P-20; business-spec 3.1; architecture 4.2): the
// sentinels, and the order of causes on inputs with two problems at once.
// The other tests check every failure of each operation where it occurs.

func TestErrorsSentinels(t *testing.T) {
	t.Parallel()

	assert.Len(t, sentinels, 12)

	for k, a := range sentinels {
		assert.NotNil(t, a.err, a.name)
		assert.NotEmpty(t, a.err.Error(), a.name)

		for l, b := range sentinels {
			assert.Equal(t, k == l, errors.Is(a.err, b.err), "%s matches %s", a.name, b.name)
		}
	}
}

// TestErrorsOrder covers the order of 4.2 (E-1 → E-2 → E-5 → E-3 → E-6 →
// causes of note when reading sources → E-7 → E-8 → causes of system and
// note for input names → E-11 → E-4 → E-9 → E-12 and causes for the result;
// for parsing E-10, then E-2 and E-12).
func TestErrorsOrder(t *testing.T) {
	t.Parallel()

	tr, irr := trad(), irregular()
	src := std()
	c4 := noteAt(t, pt(spelling.C(), 4), 0, src)
	e := noteName(t, spelling.E(), 0, src)
	freeE := noteName(t, eHalfFlat(), 0, nil)
	freeC := noteName(t, spelling.C(), 0, nil)
	free4 := noteAt(t, pt(spelling.C(), 4), 0, nil)
	eh4 := pt(eHalfFlat(), 4)
	cPlus4 := pt(spelling.MustNew(letter.C, ai(mag)), 4)

	off := newSwitch(std())
	eOff := noteAt(t, pt(spelling.E(), 4), 0, off)
	off.set(nil)

	tests := []struct {
		name   string
		op     func() error
		want   error
		causes []error
	}{
		{"E-1 before E-2: New", func() error { _, err := interval.New(interval.QualityUnset, 0); return err },
			interval.ErrUnset, nil},
		{"E-1 before E-2: AddOctaves", func() error { _, err := interval.Interval{}.AddOctaves(-1); return err },
			interval.ErrUnset, nil},
		{"E-1 before E-3: Between", func() error {
			_, err := interval.Between(spelling.Spelling{}, spelling.C(), interval.DirectionUnset)
			return err
		}, interval.ErrUnset, nil},
		{"E-1 before E-3: WithDirection", func() error {
			_, err := interval.Interval{}.WithDirection(interval.DirectionUnset)
			return err
		}, interval.ErrUnset, nil},
		{"E-1 before E-5: BetweenNotes", func() error {
			_, err := interval.BetweenNotes(note.Note{}, e, interval.DirectionUnset)
			return err
		}, interval.ErrUnset, nil},
		{"E-5 before E-3: BetweenNotes", func() error {
			_, err := interval.BetweenNotes(c4, e, interval.DirectionUnset)
			return err
		}, interval.ErrNoOctave, nil},
		{"E-5 before E-6: BetweenNotes", func() error {
			_, err := interval.BetweenNotes(free4, freeC, interval.Up)
			return err
		}, interval.ErrNoOctave, nil},
		{"E-3 before causes of system: BetweenNotesIn", func() error {
			_, err := interval.BetweenNotesIn(tr, freeC, freeE, interval.DirectionUnset)
			return err
		}, interval.ErrNoDirection, nil},
		{"E-6 before causes of note: BetweenNotes", func() error {
			_, err := interval.BetweenNotes(free4, eOff, interval.Up)
			return err
		}, interval.ErrNoSystem, nil},
		{"causes of note before E-7: BetweenNotes", func() error {
			_, err := interval.BetweenNotes(noteAt(t, pt(spelling.C(), 4), 0, set19()), eOff, interval.Up)
			return err
		}, nil, []error{note.ErrUnsetSettings}},
		{"E-1 before causes of system: BetweenPitches", func() error {
			_, err := interval.BetweenPitches(tr, eh4, pitch.Pitch{})
			return err
		}, interval.ErrUnset, nil},
		{"E-1 before causes of system: TransposePitch", func() error {
			_, err := interval.TransposePitch(tr, eh4, interval.Interval{})
			return err
		}, interval.ErrUnset, nil},
		{"E-1 before E-8: Steps", func() error { _, err := interval.Interval{}.Steps(irr, noStart); return err },
			interval.ErrUnset, nil},
		{"E-2 before E-8: FromSteps", func() error { _, err := interval.FromSteps(irr, 0, 3, noStart); return err },
			interval.ErrInvalidArgument, nil},
		{"E-2 before E-8: AllFromSteps", func() error { _, err := interval.AllFromSteps(irr, 3, -1, noStart); return err },
			interval.ErrInvalidArgument, nil},
		{"E-8 before causes of system: EnharmonicallyEqual", func() error {
			_, err := interval.EnharmonicallyEqual(irr, iv("A4"), eHalfFlat(), iv("d5"), noStart)
			return err
		}, interval.ErrStartRequired, nil},
		{"E-8 before E-9: Steps", func() error { _, err := iv("n3").Steps(irr, noStart); return err },
			interval.ErrStartRequired, nil},
		{"causes of system before E-9: EnharmonicallyEqual", func() error {
			_, err := interval.EnharmonicallyEqual(tr, iv("n3"), noStart, iv("M3"), eHalfFlat())
			return err
		}, nil, []error{system.ErrInadmissible}},
		{"E-1 before E-11: FifthsBetween", func() error {
			_, err := interval.FifthsBetween(spelling.Spelling{}, eHalfFlat())
			return err
		}, interval.ErrUnset, nil},
		{"E-4 before E-12: Invert", func() error { _, err := aug(mag, 8).Invert(); return err },
			interval.ErrNoInversion, nil},
		{"E-2 before E-12: NewMultiple", func() error {
			_, err := interval.NewMultiple(mag+1, interval.Diminished, 1)
			return err
		}, interval.ErrInvalidArgument, nil},
		{"E-2 before E-12: ExpandBy", func() error { _, err := iv("P1").ExpandBy(minInt); return err },
			interval.ErrInvalidArgument, nil},
		{"E-1 before E-2 and causes of system: Respellings", func() error {
			_, err := interval.Respellings(tr, eh4, pt(spelling.E(), 4), ai(-1), interval.ReplaceUnset)
			return err
		}, interval.ErrUnset, nil},
		{"E-2 before causes of system: Respellings", func() error {
			_, err := interval.Respellings(tr, eh4, pt(spelling.E(), 4), ai(-1), interval.ReplaceTo)
			return err
		}, interval.ErrInvalidArgument, nil},
		{"E-12 of the alteration before inadmissibility of the result: TransposePitch", func() error {
			_, err := interval.TransposePitch(tr, cPlus4, iv("{+1/2}1"))
			return err
		}, interval.ErrOutOfRange, []error{alteration.ErrOutOfRange}},
		{"E-10 before E-2: Parse", func() error { _, err := interval.Parse("{+2/4}1"); return err },
			interval.ErrParse, nil},
		{"E-2 before E-12: Parse", func() error { _, err := interval.Parse("M9223372036854775808"); return err },
			interval.ErrInvalidArgument, nil},
		{"E-10 before E-2: ParseName", func() error { _, err := interval.ParseName("major  fifth"); return err },
			interval.ErrParse, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assertErr(t, tt.op(), tt.want, tt.causes...)
		})
	}
}

// TestErrorsCausesKept covers P-20: a failure caused by a dependency keeps
// the cause next to the sentinel of interval.
func TestErrorsCausesKept(t *testing.T) {
	t.Parallel()

	_, err := interval.Transpose(spelling.MustNew(letter.C, ai(mag)), iv("A1"))
	assertErr(t, err, interval.ErrOutOfRange, alteration.ErrOutOfRange)

	_, err = interval.NewMultiple(mag+1, interval.Augmented, 4)
	assertErr(t, err, interval.ErrOutOfRange, alteration.ErrOutOfRange)

	_, err = interval.BetweenPitches(trad(), pt(spelling.C(), 4), pt(eHalfFlat(), 4))
	assertErr(t, err, nil, system.ErrInadmissible)

	_, err = interval.TransposeNote(noteAt(t, pt(spelling.C(), 4), 0, std()), iv("n3"))
	assertErr(t, err, nil, note.ErrInadmissible, system.ErrInadmissible)
}
