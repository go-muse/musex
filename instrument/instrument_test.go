package instrument_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/go-muse/musex/instrument"
	"github.com/go-muse/musex/notation"
	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/octave/number"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ОП-30, ОП-32, ОП-33, ОП-34, К-12, А 4.8.
func TestNew(t *testing.T) {
	t.Parallel()
	presets := map[string]*instrument.Preset{
		"with a range":    preset("Lead", h0(), rng("C4", "G5")),
		"without a range": preset("Lead", h24(), instrument.Range{}),
		"КП-6 in Н-0":     preset("Lead", h0(), rangeKP6()),
	}
	for name, ps := range presets {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g, err := instrument.New(ps)
			require.NoError(t, err)
			require.NotNil(t, g)
			assert.True(t, g.IsValid())
			assertState(t, ps, g)
			assert.Same(t, ps.SettingsSet(), g.Settings(), "the instrument gives the set of the preset (А 6)")
			assert.True(t, ps.Equal(g.MustState()), "К-12 the state is the preset")
			assertState(t, ps, instrument.MustNew(ps))
		})
	}

	for name, ps := range map[string]*instrument.Preset{"nil": nil, "zero": {}} {
		t.Run("unset preset "+name, func(t *testing.T) {
			t.Parallel()
			g, err := instrument.New(ps)
			assertCause(t, err, instrument.ErrUnset)
			assert.Nil(t, g)
			assertPanicsWith(t, instrument.ErrUnset, err, func() { instrument.MustNew(ps) })
		})
	}

	t.Run("the instrument keeps its own copy of the preset", func(t *testing.T) {
		t.Parallel()
		ps := preset("Lead", h0(), rng("C4", "G5"))
		g := instrument.MustNew(ps)
		*ps = *preset("Changed", h415(), instrument.Range{})
		assertState(t, preset("Lead", h0(), rng("C4", "G5")), g)
	})
}

// ОП-31, К-12: creation from parts fails as the creation of a preset (К-6).
func TestFromParts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		pname   string
		set     *settings.Settings
		r       instrument.Range
		wantErr error
	}{
		{name: "parts", pname: "Lead", set: h0(), r: rng("C4", "G5")},
		{name: "without a range", pname: "Lead", set: h0(), r: instrument.Range{}},
		{name: "КП-7 in Н-24", pname: "Lead", set: h24(), r: rangeKP7()},
		{name: "empty name", pname: "", set: h0(), r: rng("C4", "G5"), wantErr: instrument.ErrEmptyName},
		{name: "nil set", pname: "Lead", set: nil, r: rng("C4", "G5"), wantErr: instrument.ErrUnset},
		{name: "КП-6 in Н-19", pname: "Lead", set: h19(), r: rangeKP6(), wantErr: instrument.ErrInverted},
		{name: "КП-6а in Н-19", pname: "Lead", set: h19(), r: rangeKP6a(), wantErr: instrument.ErrInverted},
		{name: "КП-7 in Н-0", pname: "Lead", set: h0(), r: rangeKP7(), wantErr: instrument.ErrInadmissible},
		{name: "К-23 unset set and empty name", pname: "", set: nil, r: rng("C4", "G5"), wantErr: instrument.ErrUnset},
		{name: "К-23 empty name and inverted range", pname: "", set: h19(), r: rangeKP6(), wantErr: instrument.ErrEmptyName},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g, err := instrument.FromParts(tt.pname, tt.set, tt.r)
			_, presetErr := instrument.NewPreset(tt.pname, tt.set, tt.r)
			assertSameAnswer(t, presetErr, err)
			if tt.wantErr != nil {
				assertCause(t, err, tt.wantErr)
				assert.Nil(t, g)
				assertPanicsWith(t, tt.wantErr, err, func() { instrument.MustFromParts(tt.pname, tt.set, tt.r) })
				return
			}
			require.NoError(t, err)
			want := presetOf(t, tt.pname, tt.set, tt.r)
			assertState(t, want, g)
			assert.Same(t, tt.set, g.Settings())
			assertState(t, want, instrument.MustFromParts(tt.pname, tt.set, tt.r))
		})
	}
}

// ОП-34, ОП-35, П-10, К-13, КП-23.
func TestInstrument_StateAndClone(t *testing.T) {
	t.Parallel()
	t.Run("the state is a new preset", func(t *testing.T) {
		t.Parallel()
		g := instrument.MustFromParts("Lead", h0(), rng("C4", "G5"))
		st, err := g.State()
		require.NoError(t, err)
		st2 := g.MustState()
		assert.NotSame(t, st, st2)
		*st = *preset("Changed", h415(), instrument.Range{})
		assertState(t, preset("Lead", h0(), rng("C4", "G5")), g)
	})

	t.Run("К-13 the clone has an equal state and is another instrument", func(t *testing.T) {
		t.Parallel()
		g := instrument.MustFromParts("Lead", h0(), rng("C4", "G5"))
		c, err := g.Clone()
		require.NoError(t, err)
		assert.True(t, g.Equal(c))
		assert.False(t, g.Same(c))
		assert.False(t, settings.SameSource(g, c))
		assert.True(t, g.Equal(g.MustClone()))
	})

	changes := map[string]func(g *instrument.Instrument){
		"name":  func(g *instrument.Instrument) { g.MustSetName("Changed") },
		"set":   func(g *instrument.Instrument) { g.MustRetune(h415()) },
		"range": func(g *instrument.Instrument) { g.MustSetRange(rng("C3", "C5")) },
		"no range": func(g *instrument.Instrument) {
			g.MustRemoveRange()
		},
	}
	for name, change := range changes {
		t.Run("К-13 changing the clone does not change the original: "+name, func(t *testing.T) {
			t.Parallel()
			g := instrument.MustFromParts("Lead", h0(), rng("C4", "G5"))
			c := g.MustClone()
			change(c)
			assertState(t, preset("Lead", h0(), rng("C4", "G5")), g)
			assert.False(t, g.Equal(c))
		})
		t.Run("К-13 changing the original does not change the clone: "+name, func(t *testing.T) {
			t.Parallel()
			g := instrument.MustFromParts("Lead", h0(), rng("C4", "G5"))
			c := g.MustClone()
			change(g)
			assertState(t, preset("Lead", h0(), rng("C4", "G5")), c)
		})
	}

	t.Run("КП-23 two instruments from one preset", func(t *testing.T) {
		t.Parallel()
		ps := preset("Lead", h0(), rng("C4", "G5"))
		first := instrument.MustNew(ps)
		second := instrument.MustNew(ps)
		require.NoError(t, first.Retune(h415()))
		assert.True(t, second.Settings().Equal(h0()))
		assert.True(t, first.Settings().Equal(h415()))
		assert.True(t, ps.Equal(preset("Lead", h0(), rng("C4", "G5"))), "the preset did not change")
		assert.True(t, ps.SettingsSet().Equal(h0()))
	})
}

// ОП-33, ОП-44, П-11, К-14, КП-24, А 4.8 (Same is settings.SameSource).
func TestInstrument_Source(t *testing.T) {
	t.Parallel()
	ps := preset("Lead", h0(), rng("C4", "G5"))
	g := instrument.MustNew(ps)
	other := instrument.MustNew(ps)
	clone := g.MustClone()

	var src settings.Source = g
	assert.True(t, src.Settings().Equal(h0()), "К-14 the source gives the current set")

	tests := []struct {
		name string
		a, b settings.Source
		want bool
	}{
		{name: "КП-24 the instrument and itself", a: g, b: g, want: true},
		{name: "КП-24 two instruments from one preset", a: g, b: other, want: false},
		{name: "КП-24 the instrument and the set it gives", a: g, b: g.Settings(), want: false},
		{name: "КП-24 the instrument and its clone", a: g, b: clone, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, settings.SameSource(tt.a, tt.b))
			assert.Equal(t, tt.want, settings.SameSource(tt.b, tt.a))
		})
	}

	t.Run("К-14 retuning keeps the identity", func(t *testing.T) {
		t.Parallel()
		h := instrument.MustFromParts("Lead", h0(), rng("C4", "G5"))
		var before settings.Source = h
		require.NoError(t, h.Retune(h415()))
		assert.True(t, settings.SameSource(before, h))
		assert.True(t, h.Same(h))
		assert.True(t, before.Settings().Equal(h415()))
	})
}

// ОП-44, ОП-45, П-20, П-20а, К-22, К-25, А 4.8, КР-16.
func TestInstrument_SameAndEqual(t *testing.T) {
	t.Parallel()
	ps := preset("Lead", h0(), rng("C4", "G5"))
	g := instrument.MustNew(ps)
	twin := instrument.MustNew(ps)
	renamed := instrument.MustNew(ps).MustSetName("Second")
	zero1 := &instrument.Instrument{}
	zero2 := &instrument.Instrument{}
	var nilInstrument *instrument.Instrument

	tests := []struct {
		name      string
		a, b      *instrument.Instrument
		wantSame  bool
		wantEqual bool
	}{
		{name: "К-22 itself", a: g, b: g, wantSame: true, wantEqual: true},
		{name: "К-22 two instruments from one preset", a: g, b: twin, wantSame: false, wantEqual: true},
		{name: "К-22 after renaming one", a: g, b: renamed, wantSame: false, wantEqual: false},
		{name: "two nil", a: nil, b: nilInstrument, wantSame: true, wantEqual: true},
		{name: "one zero value and itself", a: zero1, b: zero1, wantSame: true, wantEqual: true},
		{name: "two zero values", a: zero1, b: zero2, wantSame: false, wantEqual: true},
		{name: "nil and a zero value", a: nil, b: zero1, wantSame: false, wantEqual: true},
		{name: "unset and set", a: nil, b: g, wantSame: false, wantEqual: false},
		{name: "zero value and set", a: zero1, b: g, wantSame: false, wantEqual: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertNoPanic(t, func() {
				assert.Equal(t, tt.wantSame, tt.a.Same(tt.b), "Same")
				assert.Equal(t, tt.wantSame, tt.b.Same(tt.a), "Same is symmetric")
				assert.Equal(t, settings.SameSource(tt.a, tt.b), tt.a.Same(tt.b), "Same is settings.SameSource")
				assert.Equal(t, tt.wantEqual, tt.a.Equal(tt.b), "Equal")
				assert.Equal(t, tt.wantEqual, tt.b.Equal(tt.a), "Equal is symmetric")
				assert.True(t, tt.a.Equal(tt.a), "Equal is reflexive")
			})
		})
	}

	t.Run("equal states do not make one source", func(t *testing.T) {
		t.Parallel()
		a := instrument.MustFromParts("Lead", h0(), instrument.Range{})
		b := instrument.MustFromParts("Lead", h0(), instrument.Range{})
		assert.True(t, a.Equal(b))
		assert.False(t, settings.SameSource(a, b))
	})
}

// П-20а, К-25: the unset instrument does not panic; it gives the unset set and
// empty values, and other operations fail with Е-1.
func TestInstrument_unset(t *testing.T) {
	t.Parallel()
	var declared instrument.Instrument
	unset := map[string]*instrument.Instrument{"nil": nil, "zero": {}, "declared": &declared}
	for name, g := range unset {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assertNoPanic(t, func() {
				assert.False(t, g.IsValid())
				assert.Nil(t, g.Settings(), "the unset set")
				var src settings.Source = g
				assert.False(t, src.Settings().IsValid())
				assert.Empty(t, g.Name())
				assert.Equal(t, instrument.Range{}, g.Range())
				assert.False(t, g.HasRange())
				assert.Equal(t, "Instrument(unset)", g.String())
			})

			ops := map[string]func() error{
				"State": func() error {
					st, err := g.State()
					assert.Nil(t, st)
					return err
				},
				"Clone": func() error {
					c, err := g.Clone()
					assert.Nil(t, c)
					return err
				},
				"CheckRetune":          func() error { return g.CheckRetune(h0()) },
				"CheckRetune nil":      func() error { return g.CheckRetune(nil) },
				"CheckRetuneWithRange": func() error { return g.CheckRetuneWithRange(h0(), rng("C4", "G5")) },
				"Retune":               func() error { return g.Retune(h0()) },
				"Retune nil":           func() error { return g.Retune(nil) },
				"RetuneWithRange":      func() error { return g.RetuneWithRange(h0(), rng("C4", "G5")) },
				"RetuneFunc": func() error {
					return g.RetuneFunc(func(s *settings.Settings) (*settings.Settings, error) { return h0(), nil })
				},
				"RetuneFunc nil":       func() error { return g.RetuneFunc(nil) },
				"RetuneTuning":         func() error { return g.RetuneTuning(tuning.Standard()) },
				"RetuneTuning nil":     func() error { return g.RetuneTuning(nil) },
				"RetuneSystem":         func() error { return g.RetuneSystem(trad()) },
				"RetuneSystem nil":     func() error { return g.RetuneSystem(nil) },
				"RetuneNotation":       func() error { return g.RetuneNotation(notation.MustBuiltIn(notation.EnglishUnicodeID)) },
				"RetuneNotation unset": func() error { return g.RetuneNotation(notation.Notation{}) },
				"RetuneNumbering":      func() error { return g.RetuneNumbering(number.ASPN()) },
				"RetuneCatalog":        func() error { return g.RetuneCatalog(naming.English()) },
				"RetuneDesignations": func() error {
					return g.RetuneDesignations(notation.MustBuiltIn(notation.EnglishUnicodeID), number.ASPN(), naming.English())
				},
				"SetName":       func() error { return g.SetName("Lead") },
				"SetName empty": func() error { return g.SetName("") },
				"SetRange":      func() error { return g.SetRange(rng("C4", "G5")) },
				"RemoveRange":   func() error { return g.RemoveRange() },
				"SetRange КП-7": func() error { return g.SetRange(rangeKP7()) },
				"RetuneWithRange nil and КП-7": func() error {
					return g.RetuneWithRange(nil, rangeKP7())
				},
				"CheckRetuneWithRange nil and inverted": func() error {
					return g.CheckRetuneWithRange(nil, rng("G5", "C4"))
				},
				"Contains": func() error {
					_, err := g.Contains(p("C4"))
					return err
				},
				"Contains unset pitch": func() error {
					_, err := g.Contains(pitch.Pitch{})
					return err
				},
				"Locate": func() error {
					_, err := g.Locate(p("C4"))
					return err
				},
				"Width": func() error {
					_, err := g.Width()
					return err
				},
				"Count": func() error {
					_, err := g.Count()
					return err
				},
				"Folds": func() error {
					_, err := g.Folds(p("C4"))
					return err
				},
				"Fold": func() error {
					_, _, err := g.Fold(p("C4"))
					return err
				},
				"Pitches": func() error {
					_, err := g.Pitches(system.Sharps())
					return err
				},
			}
			for opName, op := range ops {
				var err error
				assertNoPanic(t, func() { err = op() })
				assert.ErrorIs(t, err, instrument.ErrUnset, opName)
				assert.Equal(t, instrument.ErrUnset, causeOf(err), opName)
			}
			assert.False(t, g.IsValid(), "an unset instrument never becomes set")

			musts := map[string]func(){
				"MustState":                func() { g.MustState() },
				"MustClone":                func() { g.MustClone() },
				"MustCheckRetune":          func() { g.MustCheckRetune(h0()) },
				"MustCheckRetuneWithRange": func() { g.MustCheckRetuneWithRange(h0(), instrument.Range{}) },
				"MustRetune":               func() { g.MustRetune(h0()) },
				"MustSetName":              func() { g.MustSetName("Lead") },
				"MustContains":             func() { g.MustContains(p("C4")) },
				"MustWidth":                func() { g.MustWidth() },
			}
			for mustName, f := range musts {
				v, panicked := recovered(f)
				require.True(t, panicked, mustName)
				err, ok := v.(error)
				require.True(t, ok, mustName)
				assert.ErrorIs(t, err, instrument.ErrUnset, mustName)
			}
		})
	}
}

// ОП-36, П-12, К-15, К-16, КП-18…КП-22, Г-7, Г-8, Г-9.
func TestInstrument_Retune(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		from    *instrument.Preset
		to      *settings.Settings
		wantErr error
	}{
		{name: "КП-18 Н-0 to Н-415", from: preset("Lead", h0(), instrument.Range{}), to: h415()},
		{name: "КП-19 Н-0 to Н-нем", from: preset("Lead", h0(), instrument.Range{}), to: hDe(), wantErr: instrument.ErrIncompatible},
		{name: "КП-20 Н-24 to Н-нем", from: preset("Lead", h24(), instrument.Range{}), to: hDe(), wantErr: instrument.ErrIncompatible},
		{name: "КП-21 Н-0 with C4–G5 to Н-19", from: preset("Lead", h0(), rng("C4", "G5")), to: h19()},
		{name: "КП-22 Г-9 КП-6 to Н-19", from: preset("Lead", h0(), rangeKP6()), to: h19(), wantErr: instrument.ErrInverted},
		{name: "КП-22 Г-9 КП-6а to Н-19", from: preset("Lead", h0(), rangeKP6a()), to: h19(), wantErr: instrument.ErrInverted},
		{name: "К-16 Г-7 an equal set", from: preset("Lead", h0(), rng("C4", "G5")), to: h0()},
		{name: "К-16 nil set", from: preset("Lead", h0(), rng("C4", "G5")), to: nil, wantErr: instrument.ErrUnset},
		{name: "К-16 zero set", from: preset("Lead", h0(), rng("C4", "G5")), to: &settings.Settings{}, wantErr: instrument.ErrUnset},
		{name: "Н-0 to Н-24 extends the structure", from: preset("Lead", h0(), rng("C4", "G5")), to: settings.Standard().MustWithTuning(h24().Tuning())},
		{name: "Н-24 to Н-0 narrows the structure", from: preset("Lead", h24(), rangeKP7()), to: h0(), wantErr: instrument.ErrIncompatible},
		{name: "Е-6 before Е-5", from: preset("Lead", h0(), rangeKP6()), to: hDe().MustWithTuning(tuning19()), wantErr: instrument.ErrIncompatible},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := instrument.MustNew(tt.from)
			checkErr := g.CheckRetune(tt.to)
			assertState(t, tt.from, g)
			err := g.Retune(tt.to)
			assertSameAnswer(t, checkErr, err) // К-17
			if tt.wantErr != nil {
				assertCause(t, err, tt.wantErr)
				assertState(t, tt.from, g) // К-25
				assertPanicsWith(t, tt.wantErr, err, func() { g.MustRetune(tt.to) })
				assertPanicsWith(t, tt.wantErr, checkErr, func() { g.MustCheckRetune(tt.to) })
				assertState(t, tt.from, g)
				return
			}
			require.NoError(t, err)
			assertState(t, tt.from.MustWithSettingsSet(tt.to), g)
			assert.Same(t, tt.to, g.Settings(), "after a success the source gives the passed set (А 4.8)")
			h := instrument.MustNew(tt.from)
			assert.Same(t, h, h.MustCheckRetune(tt.to), "MustCheckRetune returns the receiver")
			assertState(t, tt.from, h)
			assert.Same(t, h, h.MustRetune(tt.to), "MustRetune returns the receiver")
			assertState(t, tt.from.MustWithSettingsSet(tt.to), h)
		})
	}

	t.Run("КП-18 frequency before and after", func(t *testing.T) {
		t.Parallel()
		g := instrument.MustFromParts("Lead", h0(), instrument.Range{})
		var src settings.Source = g
		assert.InDelta(t, 440.0, src.Settings().Tuning().MustFrequency(p("A4")), 1e-9)
		require.NoError(t, g.Retune(h415()))
		assert.InDelta(t, 415.0, src.Settings().Tuning().MustFrequency(p("A4")), 1e-9)
	})

	t.Run("КП-19 report of designations with a counterexample", func(t *testing.T) {
		t.Parallel()
		g := instrument.MustFromParts("Lead", h0(), instrument.Range{})
		err := g.Retune(hDe())
		ie, ok := errors.AsType[*instrument.IncompatibleError](err)
		require.True(t, ok)
		require.NotNil(t, ie)
		assert.False(t, ie.Report.Compatible())
		v, violated := ie.Report.Violation(settings.RuleDesignations)
		require.True(t, violated)
		assert.True(t, v.Name.IsValid() || v.Pitch.IsValid(), "a counterexample")
		_, structure := ie.Report.Violation(settings.RuleStructure)
		assert.False(t, structure)
		assert.True(t, g.Settings().Equal(h0()))
	})

	t.Run("КП-20 both rules violated", func(t *testing.T) {
		t.Parallel()
		g := instrument.MustFromParts("Lead", h24(), instrument.Range{})
		err := g.Retune(hDe())
		ie, ok := errors.AsType[*instrument.IncompatibleError](err)
		require.True(t, ok)
		for _, rule := range []settings.Rule{settings.RuleStructure, settings.RuleDesignations} {
			v, violated := ie.Report.Violation(rule)
			assert.True(t, violated, "%v", rule)
			assert.True(t, v.Name.IsValid() || v.Pitch.IsValid(), "a counterexample of %v", rule)
		}
		assert.True(t, g.Settings().Equal(h24()))
	})

	t.Run("КП-21 Г-8 the answer of the range check changes", func(t *testing.T) {
		t.Parallel()
		g := instrument.MustFromParts("Lead", h0(), rng("C4", "G5"))
		aDoubleFlat5 := at(spelling.ADoubleFlat(), 5)
		assert.True(t, g.MustContains(aDoubleFlat5))
		require.NoError(t, g.Retune(h19()))
		assert.False(t, g.MustContains(aDoubleFlat5))
		assert.Equal(t, instrument.Above, g.MustLocate(aDoubleFlat5))
	})

	t.Run("КП-22 Е-5 is not Е-6, and the retuning succeeds without the range", func(t *testing.T) {
		t.Parallel()
		for _, r := range []instrument.Range{rangeKP6(), rangeKP6a()} {
			g := instrument.MustFromParts("Lead", h0(), r)
			err := g.Retune(h19())
			assertCause(t, err, instrument.ErrInverted)
			assert.NotErrorIs(t, err, instrument.ErrIncompatible)
			assert.NotErrorIs(t, err, settings.ErrIncompatible)
			assertState(t, preset("Lead", h0(), r), g)
			require.NoError(t, g.RemoveRange())
			require.NoError(t, g.Retune(h19()))
			assertState(t, preset("Lead", h19(), instrument.Range{}), g)
		}
	})

	t.Run("К-15 a new reference note with the same level", func(t *testing.T) {
		t.Parallel()
		g := instrument.MustFromParts("Lead", h0(), rng("C4", "G5"))
		prev := g.Settings()
		next := prev.MustWithTuning(prev.Tuning().MustWithReferenceNote(p("A5")))
		require.NoError(t, g.Retune(next))
		ref := g.Settings().Tuning().Reference()
		assert.True(t, p("A5").Equal(ref.Pitch()), "the reference note is A5")
		assert.InDelta(t, 880.0, ref.Frequency(), 1e-9)
		same, err := prev.Tuning().SoundsSame(g.Settings().Tuning())
		require.NoError(t, err)
		assert.True(t, same, "the old and the new tunings sound the same")
	})
}

// ОП-37, К-17, К-18б: the check gives the answer of the change and changes
// nothing.
func TestInstrument_CheckRetune(t *testing.T) {
	t.Parallel()
	g := instrument.MustFromParts("Lead", h0(), rangeKP6())
	before := g.MustState()
	cases := []struct {
		name    string
		set     *settings.Settings
		r       instrument.Range
		wantErr error
	}{
		{name: "Н-415", set: h415(), r: rng("C4", "G5")},
		{name: "Н-нем", set: hDe(), r: rng("C4", "G5"), wantErr: instrument.ErrIncompatible},
		{name: "Н-19", set: h19(), r: rng("C4", "G5")},
		{name: "nil", set: nil, r: rng("C4", "G5"), wantErr: instrument.ErrUnset},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			err := g.CheckRetuneWithRange(c.set, c.r)
			if c.wantErr != nil {
				assertCause(t, err, c.wantErr)
			} else {
				assert.NoError(t, err)
			}
			assertState(t, before, g)
		})
	}

	t.Run("К-17 Е-5 for the current range", func(t *testing.T) {
		t.Parallel()
		h := instrument.MustFromParts("Lead", h0(), rangeKP6())
		assertCause(t, h.CheckRetune(h19()), instrument.ErrInverted)
		assertState(t, preset("Lead", h0(), rangeKP6()), h)
	})
}

// ОП-48, П-12а, К-17, К-18б.
func TestInstrument_RetuneWithRange(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		from    *instrument.Preset
		set     *settings.Settings
		r       instrument.Range
		wantErr error
	}{
		{name: "К-18б Н-19 with C4–G5", from: preset("Lead", h0(), rangeKP6()), set: h19(), r: rng("C4", "G5")},
		{name: "К-18б Н-19 without a range", from: preset("Lead", h0(), rangeKP6()), set: h19(), r: instrument.Range{}},
		{name: "К-18б Н-19 with КП-6а", from: preset("Lead", h0(), rangeKP6()), set: h19(), r: rangeKP6a(), wantErr: instrument.ErrInverted},
		{name: "К-18б Н-нем with C4–G5", from: preset("Lead", h0(), rangeKP6()), set: hDe(), r: rng("C4", "G5"), wantErr: instrument.ErrIncompatible},
		{name: "К-18б Н-24 to Н-0 with КП-7: structure, not Е-4", from: preset("Lead", h24(), rangeKP7()), set: h0(), r: rangeKP7(), wantErr: instrument.ErrIncompatible},
		{name: "К-18б Н-0 to Н-415 with КП-7", from: preset("Lead", h0(), rangeKP6()), set: h415(), r: rangeKP7(), wantErr: instrument.ErrInadmissible},
		{name: "the previous range does not matter", from: preset("Lead", h0(), rangeKP6a()), set: h19(), r: rng("C4", "G5")},
		{name: "a range without a previous one", from: preset("Lead", h0(), instrument.Range{}), set: h415(), r: rng("C4", "G5")},
		{name: "inverted new range", from: preset("Lead", h0(), rng("C4", "G5")), set: h415(), r: rng("G5", "C4"), wantErr: instrument.ErrInverted},
		{name: "inadmissible lower before inverted", from: preset("Lead", h0(), rng("C4", "G5")), set: h415(), r: instrument.MustNewRange(at(cHalf4().Spelling(), 6), p("C4")), wantErr: instrument.ErrInadmissible},
		{name: "nil set", from: preset("Lead", h0(), rng("C4", "G5")), set: nil, r: rangeKP7(), wantErr: instrument.ErrUnset},
		{name: "equal set and range", from: preset("Lead", h0(), rng("C4", "G5")), set: h0(), r: rng("C4", "G5")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := instrument.MustNew(tt.from)
			checkErr := g.CheckRetuneWithRange(tt.set, tt.r)
			assertState(t, tt.from, g)
			err := g.RetuneWithRange(tt.set, tt.r)
			assertSameAnswer(t, checkErr, err)
			if tt.wantErr != nil {
				assertCause(t, err, tt.wantErr)
				assertState(t, tt.from, g)
				assertPanicsWith(t, tt.wantErr, err, func() { g.MustRetuneWithRange(tt.set, tt.r) })
				assertPanicsWith(t, tt.wantErr, checkErr, func() { g.MustCheckRetuneWithRange(tt.set, tt.r) })
				assertState(t, tt.from, g)
				return
			}
			require.NoError(t, err)
			want := presetOf(t, tt.from.Name(), tt.set, tt.r)
			assertState(t, want, g)
			assert.Same(t, tt.set, g.Settings())
			h := instrument.MustNew(tt.from)
			assert.Same(t, h, h.MustCheckRetuneWithRange(tt.set, tt.r))
			assert.Same(t, h, h.MustRetuneWithRange(tt.set, tt.r))
			assertState(t, want, h)
		})
	}

	t.Run("Е-4 carries the cause of system", func(t *testing.T) {
		t.Parallel()
		g := instrument.MustFromParts("Lead", h0(), rng("C4", "G5"))
		err := g.RetuneWithRange(h415(), rangeKP7())
		assert.ErrorIs(t, err, system.ErrInadmissible)
	})
}

// ОП-38, П-12, П-15, К-18, А 4.8.2.
func TestInstrument_RetuneFunc(t *testing.T) {
	t.Parallel()
	errBoom := errors.New("transformation failed")
	german := notation.MustBuiltIn(notation.GermanAMID)
	tests := []struct {
		name      string
		from      *instrument.Preset
		f         func(*settings.Settings) (*settings.Settings, error)
		want      *settings.Settings
		wantErr   error
		wantDep   error
		wantCalls int
	}{
		{
			name: "К-18 reference frequency 415",
			from: preset("Lead", h0(), rng("C4", "G5")),
			f: func(s *settings.Settings) (*settings.Settings, error) {
				next, err := s.Tuning().WithReferenceFrequency(415)
				if err != nil {
					return nil, err
				}
				return s.WithTuning(next)
			},
			want: h415(), wantCalls: 1,
		},
		{
			name:    "К-18 notation german-am",
			from:    preset("Lead", h0(), rng("C4", "G5")),
			f:       func(s *settings.Settings) (*settings.Settings, error) { return s.WithNotation(german) },
			wantErr: instrument.ErrIncompatible, wantCalls: 1,
		},
		{
			name:    "К-18 the error of the transformation",
			from:    preset("Lead", h0(), rng("C4", "G5")),
			f:       func(*settings.Settings) (*settings.Settings, error) { return nil, errBoom },
			wantDep: errBoom, wantCalls: 1,
		},
		{
			name:    "the result is not used on an error",
			from:    preset("Lead", h0(), rng("C4", "G5")),
			f:       func(*settings.Settings) (*settings.Settings, error) { return h415(), errBoom },
			wantDep: errBoom, wantCalls: 1,
		},
		{
			name:    "the transformation returns nil",
			from:    preset("Lead", h0(), rng("C4", "G5")),
			f:       func(*settings.Settings) (*settings.Settings, error) { return nil, nil },
			wantErr: instrument.ErrUnset, wantCalls: 1,
		},
		{
			name:    "the transformation returns the unset set",
			from:    preset("Lead", h0(), rng("C4", "G5")),
			f:       func(*settings.Settings) (*settings.Settings, error) { return &settings.Settings{}, nil },
			wantErr: instrument.ErrUnset, wantCalls: 1,
		},
		{
			name:    "the range becomes inverted",
			from:    preset("Lead", h0(), rangeKP6()),
			f:       func(s *settings.Settings) (*settings.Settings, error) { return s.WithTuning(tuning19()) },
			wantErr: instrument.ErrInverted, wantCalls: 1,
		},
		{
			name: "the same set",
			from: preset("Lead", h0(), rangeKP6()),
			f:    func(s *settings.Settings) (*settings.Settings, error) { return s, nil },
			want: h0(), wantCalls: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := instrument.MustNew(tt.from)
			current := g.Settings()
			calls := 0
			err := g.RetuneFunc(func(s *settings.Settings) (*settings.Settings, error) {
				calls++
				assert.Same(t, current, s, "the transformation gets the current set")
				assert.True(t, s.Equal(tt.from.SettingsSet()))
				return tt.f(s)
			})
			assert.Equal(t, tt.wantCalls, calls, "the transformation is called once")
			switch {
			case tt.wantDep != nil:
				assertDependencyCause(t, err, tt.wantDep)
				assertState(t, tt.from, g)
			case tt.wantErr != nil:
				assertCause(t, err, tt.wantErr)
				assertState(t, tt.from, g)
				assertPanicsWith(t, tt.wantErr, err, func() { g.MustRetuneFunc(tt.f) })
			default:
				require.NoError(t, err)
				assertState(t, tt.from.MustWithSettingsSet(tt.want), g)
				h := instrument.MustNew(tt.from)
				assert.Same(t, h, h.MustRetuneFunc(tt.f))
				assertState(t, tt.from.MustWithSettingsSet(tt.want), h)
			}
		})
	}

	t.Run("Е-11 keeps a cause of the package returned by the transformation", func(t *testing.T) {
		t.Parallel()
		g := instrument.MustFromParts("Lead", h0(), rangeKP6())
		err := g.RetuneFunc(func(*settings.Settings) (*settings.Settings, error) { return nil, instrument.ErrIncompatible })
		assert.ErrorIs(t, err, instrument.ErrIncompatible)
		assertState(t, preset("Lead", h0(), rangeKP6()), g)
	})

	t.Run("nil transformation", func(t *testing.T) {
		t.Parallel()
		g := instrument.MustFromParts("Lead", h0(), rng("C4", "G5"))
		err := g.RetuneFunc(nil)
		assertCause(t, err, instrument.ErrUnset)
		assertPanicsWith(t, instrument.ErrUnset, err, func() { g.MustRetuneFunc(nil) })
		assertState(t, preset("Lead", h0(), rng("C4", "G5")), g)
	})

	t.Run("an unset instrument does not call the transformation", func(t *testing.T) {
		t.Parallel()
		var g *instrument.Instrument
		calls := 0
		err := g.RetuneFunc(func(s *settings.Settings) (*settings.Settings, error) {
			calls++
			return h0(), nil
		})
		assertCause(t, err, instrument.ErrUnset)
		assert.Zero(t, calls)
	})

	t.Run("a panic of the transformation keeps the state and releases the instrument", func(t *testing.T) {
		t.Parallel()
		g := instrument.MustFromParts("Lead", h0(), rng("C4", "G5"))
		v, panicked := recovered(func() {
			_ = g.RetuneFunc(func(*settings.Settings) (*settings.Settings, error) { panic("boom") })
		})
		require.True(t, panicked, "the panic goes on")
		assert.Equal(t, "boom", v)
		assertState(t, preset("Lead", h0(), rng("C4", "G5")), g)
		require.NoError(t, g.Retune(h415()), "the instrument can be changed again")
		require.NoError(t, g.SetName("After"))
		assertState(t, preset("After", h415(), rng("C4", "G5")), g)
	})

	t.Run("the transformation may read the instrument", func(t *testing.T) {
		t.Parallel()
		g := instrument.MustFromParts("Lead", h0(), rng("C4", "G5"))
		err := g.RetuneFunc(func(s *settings.Settings) (*settings.Settings, error) {
			assert.Same(t, s, g.Settings())
			assert.Equal(t, "Lead", g.Name())
			assert.True(t, g.MustContains(p("C4")))
			_, err := g.State()
			assert.NoError(t, err)
			return h415(), nil
		})
		require.NoError(t, err)
		assert.True(t, g.Settings().Equal(h415()))
	})
}

// ОП-47, К-18а: a part of the set is replaced as by RetuneFunc with the
// operation of settings.
func TestInstrument_RetunePart(t *testing.T) {
	t.Parallel()
	german := notation.MustBuiltIn(notation.GermanAMID)
	tests := []struct {
		name    string
		from    *instrument.Preset
		change  func(g *instrument.Instrument) error
		must    func(g *instrument.Instrument) *instrument.Instrument
		want    *settings.Settings
		wantErr error
		wantDep error
	}{
		{
			name:   "tuning of Н-415",
			from:   preset("Lead", h0(), rng("C4", "G5")),
			change: func(g *instrument.Instrument) error { return g.RetuneTuning(h415().Tuning()) },
			must:   func(g *instrument.Instrument) *instrument.Instrument { return g.MustRetuneTuning(h415().Tuning()) },
			want:   h415(),
		},
		{
			name:    "notation german-am",
			from:    preset("Lead", h0(), rng("C4", "G5")),
			change:  func(g *instrument.Instrument) error { return g.RetuneNotation(german) },
			must:    func(g *instrument.Instrument) *instrument.Instrument { return g.MustRetuneNotation(german) },
			wantErr: instrument.ErrIncompatible,
		},
		{
			name:   "numbering Franco-Belgian",
			from:   preset("Lead", h0(), rng("C4", "G5")),
			change: func(g *instrument.Instrument) error { return g.RetuneNumbering(number.FrancoBelgian()) },
			must: func(g *instrument.Instrument) *instrument.Instrument {
				return g.MustRetuneNumbering(number.FrancoBelgian())
			},
			want: h0().MustWithNumbering(number.FrancoBelgian()),
		},
		{
			name:   "empty catalog",
			from:   preset("Lead", h0(), rng("C4", "G5")),
			change: func(g *instrument.Instrument) error { return g.RetuneCatalog(naming.MustNew(nil)) },
			must:   func(g *instrument.Instrument) *instrument.Instrument { return g.MustRetuneCatalog(naming.MustNew(nil)) },
			want:   h0().MustWithCatalog(naming.MustNew(nil)),
		},
		{
			name:   "nil catalog is the empty catalog",
			from:   preset("Lead", h0(), rng("C4", "G5")),
			change: func(g *instrument.Instrument) error { return g.RetuneCatalog(nil) },
			must:   func(g *instrument.Instrument) *instrument.Instrument { return g.MustRetuneCatalog(nil) },
			want:   h0().MustWithCatalog(naming.MustNew(nil)),
		},
		{
			name: "designations german-am, ASPN, English",
			from: preset("Lead", h0(), rng("C4", "G5")),
			change: func(g *instrument.Instrument) error {
				return g.RetuneDesignations(german, number.ASPN(), naming.English())
			},
			must: func(g *instrument.Instrument) *instrument.Instrument {
				return g.MustRetuneDesignations(german, number.ASPN(), naming.English())
			},
			wantErr: instrument.ErrIncompatible,
		},
		{
			name: "designations of Н-0",
			from: preset("Lead", h0(), rng("C4", "G5")),
			change: func(g *instrument.Instrument) error {
				return g.RetuneDesignations(h0().Notation(), number.FrancoBelgian(), naming.German())
			},
			must: func(g *instrument.Instrument) *instrument.Instrument {
				return g.MustRetuneDesignations(h0().Notation(), number.FrancoBelgian(), naming.German())
			},
			want: h0().MustWithDesignations(h0().Notation(), number.FrancoBelgian(), naming.German()),
		},
		{
			name:    "system 24-quarter-tone: the cause of tuning",
			from:    preset("Lead", h0(), rng("C4", "G5")),
			change:  func(g *instrument.Instrument) error { return g.RetuneSystem(qt24()) },
			must:    func(g *instrument.Instrument) *instrument.Instrument { return g.MustRetuneSystem(qt24()) },
			wantDep: tuning.ErrClassCount,
		},
		{
			name:   "system traditional",
			from:   preset("Lead", h0(), rng("C4", "G5")),
			change: func(g *instrument.Instrument) error { return g.RetuneSystem(trad()) },
			must:   func(g *instrument.Instrument) *instrument.Instrument { return g.MustRetuneSystem(trad()) },
			want:   h0(),
		},
		{
			name:    "tuning of Н-19 inverts КП-6",
			from:    preset("Lead", h0(), rangeKP6()),
			change:  func(g *instrument.Instrument) error { return g.RetuneTuning(tuning19()) },
			must:    func(g *instrument.Instrument) *instrument.Instrument { return g.MustRetuneTuning(tuning19()) },
			wantErr: instrument.ErrInverted,
		},
		{
			name:   "tuning of Н-19 with C4–G5",
			from:   preset("Lead", h0(), rng("C4", "G5")),
			change: func(g *instrument.Instrument) error { return g.RetuneTuning(tuning19()) },
			must:   func(g *instrument.Instrument) *instrument.Instrument { return g.MustRetuneTuning(tuning19()) },
			want:   h19(),
		},
		{
			name:    "unset tuning",
			from:    preset("Lead", h0(), rng("C4", "G5")),
			change:  func(g *instrument.Instrument) error { return g.RetuneTuning(nil) },
			must:    func(g *instrument.Instrument) *instrument.Instrument { return g.MustRetuneTuning(nil) },
			wantDep: settings.ErrUnset,
		},
		{
			name:    "unset system",
			from:    preset("Lead", h0(), rng("C4", "G5")),
			change:  func(g *instrument.Instrument) error { return g.RetuneSystem(nil) },
			must:    func(g *instrument.Instrument) *instrument.Instrument { return g.MustRetuneSystem(nil) },
			wantDep: settings.ErrUnset,
		},
		{
			name:   "unset notation",
			from:   preset("Lead", h0(), rng("C4", "G5")),
			change: func(g *instrument.Instrument) error { return g.RetuneNotation(notation.Notation{}) },
			must: func(g *instrument.Instrument) *instrument.Instrument {
				return g.MustRetuneNotation(notation.Notation{})
			},
			wantDep: settings.ErrUnset,
		},
		{
			name:   "unset numbering",
			from:   preset("Lead", h0(), rng("C4", "G5")),
			change: func(g *instrument.Instrument) error { return g.RetuneNumbering(number.Numbering{}) },
			must: func(g *instrument.Instrument) *instrument.Instrument {
				return g.MustRetuneNumbering(number.Numbering{})
			},
			wantDep: settings.ErrUnset,
		},
		{
			name: "unset notation of the designations",
			from: preset("Lead", h0(), rng("C4", "G5")),
			change: func(g *instrument.Instrument) error {
				return g.RetuneDesignations(notation.Notation{}, number.ASPN(), nil)
			},
			must: func(g *instrument.Instrument) *instrument.Instrument {
				return g.MustRetuneDesignations(notation.Notation{}, number.ASPN(), nil)
			},
			wantDep: settings.ErrUnset,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := instrument.MustNew(tt.from)
			err := tt.change(g)
			switch {
			case tt.wantDep != nil:
				assertDependencyCause(t, err, tt.wantDep)
				assert.NotErrorIs(t, err, settings.ErrIncompatible, "distinct from Е-6")
				assertState(t, tt.from, g)
				assertPanicsWith(t, tt.wantDep, err, func() { tt.must(g) })
			case tt.wantErr != nil:
				assertCause(t, err, tt.wantErr)
				assertState(t, tt.from, g)
				assertPanicsWith(t, tt.wantErr, err, func() { tt.must(g) })
			default:
				require.NoError(t, err)
				assertState(t, tt.from.MustWithSettingsSet(tt.want), g)
				h := instrument.MustNew(tt.from)
				assert.Same(t, h, tt.must(h))
				assertState(t, tt.from.MustWithSettingsSet(tt.want), h)
			}
			assertState(t, tt.from.MustWithSettingsSet(g.Settings()), g)
		})
	}

	t.Run("the other parts stay", func(t *testing.T) {
		t.Parallel()
		g := instrument.MustFromParts("Lead", h0(), rng("C4", "G5"))
		require.NoError(t, g.RetuneNumbering(number.FrancoBelgian()))
		s := g.Settings()
		assert.True(t, s.Tuning().Equal(h0().Tuning()))
		assert.True(t, s.Notation().Equal(h0().Notation()))
		assert.True(t, s.Catalog().Equal(h0().Catalog()))
		assert.True(t, s.Numbering().Equal(number.FrancoBelgian()))
	})
}

// ОП-39, ОП-40, ОП-41, К-19, П-1а.
func TestInstrument_NameAndRange(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		from    *instrument.Preset
		change  func(g *instrument.Instrument) error
		must    func(g *instrument.Instrument) *instrument.Instrument
		want    *instrument.Preset
		wantErr error
	}{
		{
			name:   "new name",
			from:   preset("Lead", h0(), rng("C4", "G5")),
			change: func(g *instrument.Instrument) error { return g.SetName("Second") },
			must:   func(g *instrument.Instrument) *instrument.Instrument { return g.MustSetName("Second") },
			want:   preset("Second", h0(), rng("C4", "G5")),
		},
		{
			name:    "empty name",
			from:    preset("Lead", h0(), rng("C4", "G5")),
			change:  func(g *instrument.Instrument) error { return g.SetName("") },
			must:    func(g *instrument.Instrument) *instrument.Instrument { return g.MustSetName("") },
			wantErr: instrument.ErrEmptyName,
		},
		{
			name:   "valid range",
			from:   preset("Lead", h0(), rng("C4", "G5")),
			change: func(g *instrument.Instrument) error { return g.SetRange(rng("C3", "C5")) },
			must:   func(g *instrument.Instrument) *instrument.Instrument { return g.MustSetRange(rng("C3", "C5")) },
			want:   preset("Lead", h0(), rng("C3", "C5")),
		},
		{
			name:   "range without a previous one",
			from:   preset("Lead", h0(), instrument.Range{}),
			change: func(g *instrument.Instrument) error { return g.SetRange(rangeKP6()) },
			must:   func(g *instrument.Instrument) *instrument.Instrument { return g.MustSetRange(rangeKP6()) },
			want:   preset("Lead", h0(), rangeKP6()),
		},
		{
			name:    "КП-7 in Н-0",
			from:    preset("Lead", h0(), rng("C4", "G5")),
			change:  func(g *instrument.Instrument) error { return g.SetRange(rangeKP7()) },
			must:    func(g *instrument.Instrument) *instrument.Instrument { return g.MustSetRange(rangeKP7()) },
			wantErr: instrument.ErrInadmissible,
		},
		{
			name:    "КП-6 in Н-19",
			from:    preset("Lead", h19(), rng("C4", "G5")),
			change:  func(g *instrument.Instrument) error { return g.SetRange(rangeKP6()) },
			must:    func(g *instrument.Instrument) *instrument.Instrument { return g.MustSetRange(rangeKP6()) },
			wantErr: instrument.ErrInverted,
		},
		{
			name:   "КП-7 in Н-24",
			from:   preset("Lead", h24(), instrument.Range{}),
			change: func(g *instrument.Instrument) error { return g.SetRange(rangeKP7()) },
			must:   func(g *instrument.Instrument) *instrument.Instrument { return g.MustSetRange(rangeKP7()) },
			want:   preset("Lead", h24(), rangeKP7()),
		},
		{
			name:   "the unset range is no range",
			from:   preset("Lead", h0(), rng("C4", "G5")),
			change: func(g *instrument.Instrument) error { return g.SetRange(instrument.Range{}) },
			must:   func(g *instrument.Instrument) *instrument.Instrument { return g.MustSetRange(instrument.Range{}) },
			want:   preset("Lead", h0(), instrument.Range{}),
		},
		{
			name:   "remove the range",
			from:   preset("Lead", h0(), rng("C4", "G5")),
			change: func(g *instrument.Instrument) error { return g.RemoveRange() },
			must:   func(g *instrument.Instrument) *instrument.Instrument { return g.MustRemoveRange() },
			want:   preset("Lead", h0(), instrument.Range{}),
		},
		{
			name:   "remove the range again",
			from:   preset("Lead", h0(), instrument.Range{}),
			change: func(g *instrument.Instrument) error { return g.RemoveRange() },
			must:   func(g *instrument.Instrument) *instrument.Instrument { return g.MustRemoveRange() },
			want:   preset("Lead", h0(), instrument.Range{}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := instrument.MustNew(tt.from)
			err := tt.change(g)
			if tt.wantErr != nil {
				assertCause(t, err, tt.wantErr)
				assertState(t, tt.from, g)
				assertPanicsWith(t, tt.wantErr, err, func() { tt.must(g) })
				assertState(t, tt.from, g)
				return
			}
			require.NoError(t, err)
			assertState(t, tt.want, g)
			h := instrument.MustNew(tt.from)
			assert.Same(t, h, tt.must(h))
			assertState(t, tt.want, h)
		})
	}

	t.Run("chain", func(t *testing.T) {
		t.Parallel()
		g := instrument.MustFromParts("Lead", h0(), instrument.Range{})
		g.MustSetName("Second").MustSetRange(rng("C4", "G5")).MustRetune(h415()).MustRemoveRange()
		assertState(t, preset("Second", h415(), instrument.Range{}), g)
	})
}

// ОП-42, ОП-43, К-20, К-21: the range operations of an instrument are those
// of its range in the system of its set.
func TestInstrument_rangeOperations(t *testing.T) {
	t.Parallel()
	instruments := map[string]*instrument.Preset{
		"soprano in Н-0":  preset("Soprano", h0(), rng("C4", "G5")),
		"soprano in Н-19": preset("Soprano", h19(), rng("C4", "G5")),
		"КП-6 in Н-0":     preset("Lead", h0(), rangeKP6()),
		"КП-7 in Н-24":    preset("Lead", h24(), rangeKP7()),
		"piano in Н-0":    preset("Keys", h0(), rng("A0", "C8")),
	}
	xs := []pitch.Pitch{p("B♯3"), p("B3"), p("C♭4"), p("G5"), p("A5"), p("E2"), p("B5"), at(spelling.ADoubleFlat(), 5), p("C4")}
	rules := []system.SpellingRule{system.Sharps(), system.Flats()}
	for name, ps := range instruments {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := instrument.MustNew(ps)
			sys := ps.SettingsSet().System()
			r := ps.Range()
			for _, x := range xs {
				in, err := g.Contains(x)
				wantIn, wantErr := r.Contains(sys, x)
				assertSameAnswer(t, wantErr, err)
				assert.Equal(t, wantIn, in, "Contains %v", x)

				loc, err := g.Locate(x)
				wantLoc, wantErr := r.Locate(sys, x)
				assertSameAnswer(t, wantErr, err)
				assert.Equal(t, wantLoc, loc, "Locate %v", x)

				folds, err := g.Folds(x)
				wantFolds, wantErr := r.Folds(sys, x)
				assertSameAnswer(t, wantErr, err)
				if err == nil {
					assertPitchesEqual(t, slices.Collect(wantFolds), slices.Collect(folds))
				}

				fold, ok, err := g.Fold(x)
				wantFold, wantOK, wantErr := r.Fold(sys, x)
				assertSameAnswer(t, wantErr, err)
				assert.Equal(t, wantOK, ok)
				assert.True(t, wantFold.SameSpelling(fold) && wantFold.SameOctave(fold) || !wantOK)
			}
			w, err := g.Width()
			require.NoError(t, err)
			assert.Equal(t, r.MustWidth(sys), w)
			assert.Equal(t, w, g.MustWidth())
			c, err := g.Count()
			require.NoError(t, err)
			assert.Equal(t, r.MustCount(sys), c)
			assert.Equal(t, c, g.MustCount())
			for _, rule := range rules {
				seq, err := g.Pitches(rule)
				require.NoError(t, err)
				assertPitchesEqual(t, slices.Collect(r.MustPitches(sys, rule)), slices.Collect(seq))
				assertPitchesEqual(t, slices.Collect(r.MustPitches(sys, rule)), slices.Collect(g.MustPitches(rule)))
			}
		})
	}

	t.Run("К-20 КП-1", func(t *testing.T) {
		t.Parallel()
		g := instrument.MustFromParts("Soprano", h0(), rng("C4", "G5"))
		for x, want := range map[string]bool{"B♯3": true, "B3": false, "C♭4": false, "G5": true, "A5": false} {
			got, err := g.Contains(p(x))
			require.NoError(t, err)
			assert.Equal(t, want, got, x)
			assert.Equal(t, want, g.MustContains(p(x)), x)
		}
		assert.Equal(t, instrument.Below, g.MustLocate(p("B3")))
		assert.Equal(t, instrument.Inside, g.MustLocate(p("C4")))
		assert.Equal(t, instrument.Above, g.MustLocate(p("A5")))
	})

	t.Run("К-21 КП-3, КП-4, КП-14", func(t *testing.T) {
		t.Parallel()
		bassPart := instrument.MustFromParts("Bass", h0(), rng("F2", "D4"))
		assertPitchesEqual(t, []pitch.Pitch{p("E3")}, slices.Collect(bassPart.MustFolds(p("E2"))))
		e, ok := bassPart.MustFold(p("E2"))
		require.True(t, ok)
		assertPitchesEqual(t, []pitch.Pitch{p("E3")}, []pitch.Pitch{e})

		keys := instrument.MustFromParts("Keys", h0(), rng("A0", "C8"))
		assert.Equal(t, system.Step(87), keys.MustWidth())
		assert.Equal(t, 88, keys.MustCount())

		small := instrument.MustFromParts("Small", h0(), rng("C4", "E4"))
		assertPitchesEqual(t, []pitch.Pitch{p("C4"), p("C♯4"), p("D4"), p("D♯4"), p("E4")}, slices.Collect(small.MustPitches(system.Sharps())))
		_, found, err := small.Fold(p("G4"))
		require.NoError(t, err)
		assert.False(t, found)
	})

	t.Run("К-20 errors", func(t *testing.T) {
		t.Parallel()
		noRange := instrument.MustFromParts("Lead", h0(), instrument.Range{})
		withRange := instrument.MustFromParts("Lead", h0(), rng("C4", "G5"))

		_, err := noRange.Contains(p("C4"))
		assertCause(t, err, instrument.ErrNoRange)
		assertPanicsWith(t, instrument.ErrNoRange, err, func() { noRange.MustContains(p("C4")) })
		_, err = noRange.Contains(pitch.Pitch{})
		assertCause(t, err, instrument.ErrNoRange)

		_, err = withRange.Contains(pitch.Pitch{})
		assertCause(t, err, instrument.ErrUnsetPitch)
		_, err = withRange.Contains(cHalf4())
		assertCause(t, err, instrument.ErrInadmissible)
		assertPanicsWith(t, instrument.ErrInadmissible, err, func() { withRange.MustContains(cHalf4()) })

		noRangeOps := map[string]func() error{
			"Locate": func() error {
				l, err := noRange.Locate(p("C4"))
				assert.Equal(t, instrument.LocationUnset, l)
				return err
			},
			"Width": func() error {
				w, err := noRange.Width()
				assert.Zero(t, w)
				return err
			},
			"Count": func() error {
				c, err := noRange.Count()
				assert.Zero(t, c)
				return err
			},
			"Folds": func() error {
				s, err := noRange.Folds(p("C4"))
				assert.Nil(t, s)
				return err
			},
			"Fold": func() error {
				_, ok, err := noRange.Fold(p("C4"))
				assert.False(t, ok)
				return err
			},
			"Pitches":            func() error { _, err := noRange.Pitches(system.Sharps()); return err },
			"Pitches unset rule": func() error { _, err := noRange.Pitches(system.SpellingRule{}); return err },
		}
		for _, op := range noRangeOps {
			assertCause(t, op(), instrument.ErrNoRange)
		}
		mustOps := map[string]func(){
			"MustLocate":  func() { noRange.MustLocate(p("C4")) },
			"MustWidth":   func() { noRange.MustWidth() },
			"MustCount":   func() { noRange.MustCount() },
			"MustFolds":   func() { noRange.MustFolds(p("C4")) },
			"MustFold":    func() { noRange.MustFold(p("C4")) },
			"MustPitches": func() { noRange.MustPitches(system.Sharps()) },
		}
		for name, f := range mustOps {
			v, panicked := recovered(f)
			require.True(t, panicked, name)
			err, ok := v.(error)
			require.True(t, ok, name)
			assert.ErrorIs(t, err, instrument.ErrNoRange, name)
		}

		_, err = withRange.Pitches(system.SpellingRule{})
		assertDependencyCause(t, err, system.ErrInvalid)
	})

	t.Run("a sequence keeps the state of its call", func(t *testing.T) {
		t.Parallel()
		g := instrument.MustFromParts("Lead", h0(), rng("C4", "E4"))
		pitches := g.MustPitches(system.Sharps())
		folds := g.MustFolds(p("D1"))
		require.NoError(t, g.RetuneWithRange(h19(), rng("C4", "G5")))
		assert.Len(t, slices.Collect(pitches), 5)
		assertPitchesEqual(t, []pitch.Pitch{p("D4")}, slices.Collect(folds))
	})
}

// ОП-46, К-26, А 4.10.
func TestInstrument_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		g    *instrument.Instrument
	}{
		{name: "with a range", g: instrument.MustFromParts("Lead voice", h0(), rng("C4", "G5"))},
		{name: "without a range", g: instrument.MustFromParts("Lead voice", h24(), instrument.Range{})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := tt.g.String()
			assert.Contains(t, s, tt.g.Name())
			assert.Contains(t, s, tt.g.Settings().String())
			if tt.g.HasRange() {
				assert.Contains(t, s, tt.g.Range().String())
			} else {
				assert.Contains(t, s, "no range")
			}
			assert.NotEqual(t, "Instrument(unset)", s)
		})
	}
}

// Е-6, КР-8, А 4.3: the error of an incompatible retuning.
func TestIncompatibleError(t *testing.T) {
	t.Parallel()
	t.Run("from a retuning", func(t *testing.T) {
		t.Parallel()
		g := instrument.MustFromParts("Lead", h0(), instrument.Range{})
		err := g.Retune(hDe())
		ie, ok := errors.AsType[*instrument.IncompatibleError](err)
		require.True(t, ok)
		assert.False(t, ie.Report.Compatible())
		assert.NotEmpty(t, ie.Error())
		assert.ErrorIs(t, ie, instrument.ErrIncompatible)
		assert.ErrorIs(t, ie, settings.ErrIncompatible)
		unwrapped := ie.Unwrap()
		require.Len(t, unwrapped, 2)
		assert.Contains(t, unwrapped, instrument.ErrIncompatible)
		reportErr := ie.Report.Err()
		require.Error(t, reportErr)
		assert.True(t, slices.ContainsFunc(unwrapped, func(e error) bool {
			return e != nil && e.Error() == reportErr.Error() && errors.Is(e, settings.ErrIncompatible)
		}), "Unwrap gives the error of the report")
		want := g.MustState().SettingsSet().MustCheckReplacement(hDe())
		assert.Equal(t, want, ie.Report, "the report of CheckReplacement")
	})

	t.Run("nil and the zero value do not panic", func(t *testing.T) {
		t.Parallel()
		for name, ie := range map[string]*instrument.IncompatibleError{"nil": nil, "zero": {}} {
			assertNoPanic(t, func() {
				_ = ie.Error()
				assert.Equal(t, []error{instrument.ErrIncompatible}, ie.Unwrap(), name)
			})
		}
		var err error = &instrument.IncompatibleError{}
		assert.ErrorIs(t, err, instrument.ErrIncompatible)
		assert.NotErrorIs(t, err, settings.ErrUnset, "the zero report is not included")
		assert.NotErrorIs(t, err, settings.ErrIncompatible)
	})
}

// К-23, А 4.3: the ten own causes are pairwise distinct and are not the
// causes of the dependencies.
func TestErrors(t *testing.T) {
	t.Parallel()
	for i, a := range sentinels {
		require.Error(t, a)
		for j, b := range sentinels {
			assert.Equal(t, i == j, errors.Is(a, b), "%v and %v", a, b)
		}
		for _, dep := range []error{settings.ErrUnset, settings.ErrIncompatible, system.ErrInvalid, system.ErrInadmissible, system.ErrOutOfRange} {
			assert.NotErrorIs(t, a, dep)
		}
	}
}
