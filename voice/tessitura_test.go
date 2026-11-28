package voice_test

import (
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/instrument"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/voice"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type location struct {
	index   int
	where   instrument.Location
	outside system.Step
}

func assertLocations(t *testing.T, want []location, got []voice.SoundLocation) {
	t.Helper()
	require.NotNil(t, got)
	require.Len(t, got, len(want))
	for i, w := range want {
		assert.Equal(t, w.index, got[i].Index, "sound %d", i)
		assert.Equal(t, w.where, got[i].Location, "sound %d", i)
		assert.Equal(t, w.outside, got[i].Outside, "sound %d", i)
	}
}

// assertTessitura compares two results by the value of the number and the
// kind.
func assertTessitura(t *testing.T, want, got voice.Tessitura) {
	t.Helper()
	assert.True(t, want.Number.Equal(got.Number), "number: want %v, got %v", want.Number, got.Number)
	assert.Equal(t, want.Kind, got.Kind)
}

// sopranoInstrument returns a new instrument from the built-in soprano
// voice; every test owns its instrument.
func sopranoInstrument() *instrument.Instrument {
	return instrument.MustNew(soprano())
}

type tessituraCase struct {
	name   string
	ref    func() voice.Reference
	l      *voice.Line
	number string
	kind   voice.TessituraKind
	count  string // the number weighted by count, when checked
}

// КП-16…КП-20, КП-50, П-11: the number of the tessitura and its kind.
func TestTessitura(t *testing.T) {
	t.Parallel()
	sopranoRef := func() voice.Reference { return voice.PresetReference(soprano()) }
	bassRef := func() voice.Reference { return voice.PresetReference(bass()) }
	tests := []tessituraCase{
		{"КП-16", sopranoRef, line(snd(c4(), q()), snd(e4(), q()), snd(g4(), h())), "9/38", voice.TessituraLow, "11/57"},
		{"КП-17", sopranoRef, line(snd(nt(spelling.A(), 4), w())), "9/19", voice.TessituraMiddle, "9/19"},
		{"КП-18", sopranoRef, line(snd(nt(spelling.E(), 5), h()), snd(nt(spelling.G(), 5), h())), "35/38", voice.TessituraHigh, "35/38"},
		{"КП-19 lower bound of the middle zone", bassRef, line(snd(nt(spelling.C(), 3), w())), "1/3", voice.TessituraMiddle, "1/3"},
		{"КП-50 upper bound of the middle zone", bassRef, line(snd(nt(spelling.G(), 3), w())), "2/3", voice.TessituraMiddle, "2/3"},
		{"КП-20 outside the range", bassRef, line(snd(nt(spelling.C(), 2), h()), snd(e4(), h())), "3/7", voice.TessituraMiddle, "3/7"},
		{"lower bound", sopranoRef, line(snd(c4(), q())), "0/1", voice.TessituraLow, "0/1"},
		{"upper bound", sopranoRef, line(snd(nt(spelling.G(), 5), q())), "1/1", voice.TessituraHigh, "1/1"},
		{"below the range", sopranoRef, line(snd(nt(spelling.A(), 3), q())), "-3/19", voice.TessituraLow, "-3/19"},
		{"above the range", sopranoRef, line(snd(nt(spelling.C(), 6), q())), "24/19", voice.TessituraHigh, "24/19"},
		{"rests are not counted, a tied sound has its full length", sopranoRef,
			line(snd(c4(), qe()), rst(h()), snd(g4(), q())), "14/95", voice.TessituraLow, "7/38"},
		{"role of a placement", func() voice.Reference { return voice.PlacementReference(voice.Soprano) },
			line(snd(c4(), q()), snd(e4(), q()), snd(g4(), h())), "9/38", voice.TessituraLow, "11/57"},
		{"range value", func() voice.Reference { return voice.RangeReference(soprano().Range()) },
			line(snd(c4(), q()), snd(e4(), q()), snd(g4(), h())), "9/38", voice.TessituraLow, "11/57"},
		{"instrument", func() voice.Reference { return voice.InstrumentReference(sopranoInstrument()) },
			line(snd(c4(), q()), snd(e4(), q()), snd(g4(), h())), "9/38", voice.TessituraLow, "11/57"},
		{"КП-43 unbound notes and an instrument", func() voice.Reference { return voice.InstrumentReference(sopranoInstrument()) },
			line(snd(free(spelling.C(), 4), q()), snd(free(spelling.E(), 4), q())), "2/19", voice.TessituraLow, "2/19"},
		{"КП-43 unbound notes and a preset", sopranoRef,
			line(snd(free(spelling.C(), 4), q()), snd(free(spelling.E(), 4), q())), "2/19", voice.TessituraLow, "2/19"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.l.Tessitura(tt.ref())
			require.NoError(t, err)
			assert.Equal(t, tt.number, got.Number.String(), "the number is a reduced fraction")
			assert.Equal(t, tt.kind, got.Kind)
			assertTessitura(t, got, tt.l.MustTessitura(tt.ref()))

			byLength, err := tt.l.TessituraBy(tt.ref(), voice.ByLength, voice.DefaultZones())
			require.NoError(t, err)
			assertTessitura(t, got, byLength)
			assertTessitura(t, got, tt.l.MustTessituraBy(tt.ref(), voice.ByLength, voice.DefaultZones()))

			byCount, err := tt.l.TessituraBy(tt.ref(), voice.ByCount, voice.DefaultZones())
			require.NoError(t, err)
			assert.Equal(t, tt.count, byCount.Number.String())

			in, err := tt.l.TessituraIn(traditional(), tt.ref())
			require.NoError(t, err)
			assertTessitura(t, got, in)
			assertTessitura(t, got, tt.l.MustTessituraIn(traditional(), tt.ref()))
			byIn, err := tt.l.TessituraByIn(traditional(), tt.ref(), voice.ByCount, voice.DefaultZones())
			require.NoError(t, err)
			assertTessitura(t, byCount, byIn)
			assertTessitura(t, byCount, tt.l.MustTessituraByIn(traditional(), tt.ref(), voice.ByCount, voice.DefaultZones()))
		})
	}
}

// П-11: the user may set the zones; a bound belongs to the middle zone.
func TestTessituraZones(t *testing.T) {
	t.Parallel()
	l := line(snd(c4(), q()), snd(e4(), q()), snd(g4(), h())) // r = 9/38
	ref := voice.PresetReference(soprano())
	tests := []struct {
		name      string
		low, high fraction.Fraction
		kind      voice.TessituraKind
	}{
		{"both bounds equal r", frac(9, 38), frac(9, 38), voice.TessituraMiddle},
		{"r on the low bound", frac(9, 38), frac(1, 2), voice.TessituraMiddle},
		{"r on the high bound", frac(0, 1), frac(9, 38), voice.TessituraMiddle},
		{"r below", frac(1, 2), frac(1, 1), voice.TessituraLow},
		{"r above", frac(0, 1), frac(1, 5), voice.TessituraHigh},
		{"whole range is middle", frac(0, 1), frac(1, 1), voice.TessituraMiddle},
		{"unreduced bounds", frac(18, 76), frac(2, 4), voice.TessituraMiddle},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			z := voice.MustNewZones(tt.low, tt.high)
			got, err := l.TessituraBy(ref, voice.ByLength, z)
			require.NoError(t, err)
			assert.Equal(t, "9/38", got.Number.String())
			assert.Equal(t, tt.kind, got.Kind)
		})
	}
}

// Architecture 6.10: Zones.
func TestZones(t *testing.T) {
	t.Parallel()
	d := voice.DefaultZones()
	assert.True(t, d.IsSet())
	assert.True(t, frac(1, 3).Equal(d.Low()))
	assert.True(t, frac(2, 3).Equal(d.High()))
	assert.NotEmpty(t, d.String())
	assert.NotEqual(t, "Zones(unset)", d.String())

	z, err := voice.NewZones(frac(2, 6), frac(4, 6))
	require.NoError(t, err)
	assert.True(t, z.IsSet())
	assert.True(t, z.Low().EqualExact(frac(2, 6)), "bounds are kept as given: %v", z.Low())
	assert.True(t, z.High().EqualExact(frac(4, 6)), "bounds are kept as given: %v", z.High())
	assert.True(t, z.Equal(d), "equality compares values")
	assert.True(t, d.Equal(z))
	assert.False(t, z.Equal(voice.MustNewZones(frac(1, 3), frac(1, 2))))

	for _, b := range [][2]fraction.Fraction{{frac(0, 1), frac(0, 1)}, {frac(1, 1), frac(1, 1)}, {frac(0, 1), frac(1, 1)}} {
		_, err := voice.NewZones(b[0], b[1])
		assert.NoError(t, err, "%v–%v", b[0], b[1])
	}

	var unset voice.Zones
	assert.False(t, unset.IsSet())
	assert.Equal(t, "Zones(unset)", unset.String())
	assert.False(t, unset.Low().IsValid())
	assert.False(t, unset.High().IsValid())
	assert.True(t, unset.Equal(voice.Zones{}))
	assert.False(t, unset.Equal(d))
	assert.False(t, d.Equal(unset))
}

// КП-24, Е-11 and the order of the checks of NewZones.
func TestNewZonesFailures(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		low, high fraction.Fraction
		want      want
	}{
		{"КП-24 inverted", frac(2, 3), frac(1, 3), fails(voice.ErrZones).only()},
		{"КП-24 negative", frac(-1, 3), frac(1, 2), fails(voice.ErrZones).only()},
		{"КП-24 above one", frac(1, 2), frac(4, 3), fails(voice.ErrZones).only()},
		{"unset low", fraction.Fraction{}, frac(1, 2), fails(voice.ErrUnset, fraction.ErrInvalid)},
		{"unset high", frac(1, 2), fraction.Fraction{}, fails(voice.ErrUnset, fraction.ErrInvalid)},
		{"unset low before invalid high", fraction.Fraction{}, frac(4, 3), fails(voice.ErrUnset, fraction.ErrInvalid)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertFailure(t, tt.want,
				func() (voice.Zones, error) { return voice.NewZones(tt.low, tt.high) },
				func() voice.Zones { return voice.MustNewZones(tt.low, tt.high) })
		})
	}
}

// КП-16, КП-20, КП-22, ОП-22, ОП-23: the position of every sound relative to
// the reference range and the sounds outside it.
func TestLocateAndWithin(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		ref     func() voice.Reference
		l       *voice.Line
		want    []location
		within  bool
		outside []int
	}{
		{"КП-16 all inside", func() voice.Reference { return voice.PresetReference(soprano()) },
			line(snd(c4(), q()), snd(e4(), q()), snd(g4(), h())),
			[]location{{0, instrument.Inside, 0}, {1, instrument.Inside, 0}, {2, instrument.Inside, 0}}, true, []int{}},
		{"КП-20 below and above", func() voice.Reference { return voice.PresetReference(bass()) },
			line(snd(nt(spelling.C(), 2), h()), snd(e4(), h())),
			[]location{{0, instrument.Below, 5}, {1, instrument.Above, 2}}, false, []int{0, 1}},
		{"КП-22 range of one pitch", func() voice.Reference {
			return voice.RangeReference(newRange(pt(spelling.C(), 4), pt(spelling.C(), 4)))
		},
			line(snd(c4(), q())), []location{{0, instrument.Inside, 0}}, true, []int{}},
		{"rests are skipped", func() voice.Reference { return voice.PlacementReference(voice.Bass) },
			line(rst(q()), snd(nt(spelling.F(), 2), q()), rst(q()), snd(nt(spelling.E(), 2), q()), snd(nt(spelling.E(), 4), q())),
			[]location{{1, instrument.Inside, 0}, {3, instrument.Below, 1}, {4, instrument.Above, 2}}, false, []int{3, 4}},
		{"instrument", func() voice.Reference { return voice.InstrumentReference(sopranoInstrument()) },
			line(snd(nt(spelling.B(), 3), q()), snd(nt(spelling.A(), 5), q())),
			[]location{{0, instrument.Below, 1}, {1, instrument.Above, 2}}, false, []int{0, 1}},
		{"empty line", func() voice.Reference { return voice.PresetReference(soprano()) }, line(), []location{}, true, []int{}},
		{"line of rests", func() voice.Reference { return voice.PresetReference(soprano()) }, line(rst(w())), []location{}, true, []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.l.Locate(tt.ref())
			require.NoError(t, err)
			assertLocations(t, tt.want, got)
			assertLocations(t, tt.want, tt.l.MustLocate(tt.ref()))
			in, err := tt.l.LocateIn(traditional(), tt.ref())
			require.NoError(t, err)
			assertLocations(t, tt.want, in)
			assertLocations(t, tt.want, tt.l.MustLocateIn(traditional(), tt.ref()))

			within, outside, err := tt.l.Within(tt.ref())
			require.NoError(t, err)
			assert.Equal(t, tt.within, within)
			require.NotNil(t, outside)
			assert.Equal(t, tt.outside, outside)
			mw, mo := tt.l.MustWithin(tt.ref())
			assert.Equal(t, tt.within, mw)
			assert.Equal(t, tt.outside, mo)
			iw, io, err := tt.l.WithinIn(traditional(), tt.ref())
			require.NoError(t, err)
			assert.Equal(t, tt.within, iw)
			assert.Equal(t, tt.outside, io)
			miw, mio := tt.l.MustWithinIn(traditional(), tt.ref())
			assert.Equal(t, tt.within, miw)
			assert.Equal(t, tt.outside, mio)
		})
	}
}

// The tessitura is a mean position: a line outside the range can have a
// middle tessitura (КП-20, НФ-4).
func TestTessituraIsNotWithin(t *testing.T) {
	t.Parallel()
	l := line(snd(nt(spelling.C(), 2), h()), snd(e4(), h()))
	ref := voice.PresetReference(bass())
	within, _ := l.MustWithin(ref)
	assert.False(t, within)
	assert.Equal(t, voice.TessituraMiddle, l.MustTessitura(ref).Kind)
}

// A preset and an instrument from it give the same results (П-9а).
func TestPresetAndInstrumentAgree(t *testing.T) {
	t.Parallel()
	l := line(snd(free(spelling.C(), 4), q()), snd(nt(spelling.A(), 5), h()), rst(q()), snd(free(spelling.B(), 3), e8()))
	preset := voice.PresetReference(soprano())
	inst := voice.InstrumentReference(sopranoInstrument())
	assertTessitura(t, l.MustTessitura(preset), l.MustTessitura(inst))
	assert.Equal(t, l.MustLocate(preset), l.MustLocate(inst))
	pw, po := l.MustWithin(preset)
	iw, io := l.MustWithin(inst)
	assert.Equal(t, pw, iw)
	assert.Equal(t, po, io)
}

// A Reference borrows the instrument and reads its state at every
// operation (architecture 4.2).
func TestInstrumentReferenceReadsCurrentState(t *testing.T) {
	t.Parallel()
	inst := sopranoInstrument()
	ref := voice.InstrumentReference(inst)
	l := line(snd(c4(), q()), snd(g4(), q()))
	within, _ := l.MustWithin(ref)
	assert.True(t, within)

	require.NoError(t, inst.SetRange(newRange(pt(spelling.C(), 4), pt(spelling.E(), 4))))
	within, outside := l.MustWithin(ref)
	assert.False(t, within)
	assert.Equal(t, []int{1}, outside)

	require.NoError(t, inst.RemoveRange())
	assertFailure(t, fails(voice.ErrNoRange).only(),
		func() (pair[bool, []int], error) { return two(l.Within(ref)) },
		func() pair[bool, []int] { return mustTwo(l.MustWithin(ref)) })
}

// КП-43, КП-44, П-9а: the system of the comparison without an explicit
// system.
func TestTessituraSystems(t *testing.T) {
	t.Parallel()
	freeLine := line(snd(free(spelling.C(), 4), q()), snd(free(spelling.E(), 4), q()))
	line19 := line(snd(n19(spelling.C(), 4), q()), snd(n19(spelling.E(), 4), q()))

	tests := []struct {
		name string
		l    *voice.Line
		ref  func() voice.Reference
		want want
	}{
		{"КП-43 unbound notes and a range value", freeLine,
			func() voice.Reference { return voice.RangeReference(soprano().Range()) }, failsAt(0, voice.ErrNoSystem)},
		{"КП-43 unbound notes and a placement", freeLine,
			func() voice.Reference { return voice.PlacementReference(voice.Soprano) }, failsAt(0, voice.ErrNoSystem)},
		{"КП-44 another system and an instrument", line19,
			func() voice.Reference { return voice.InstrumentReference(sopranoInstrument()) }, failsAt(0, voice.ErrDifferentSystems)},
		{"КП-44 another system and a preset", line19,
			func() voice.Reference { return voice.PresetReference(soprano()) }, failsAt(0, voice.ErrDifferentSystems)},
		{"bound note after unbound ones", line(snd(free(spelling.C(), 4), q()), rst(q()), snd(n19(spelling.E(), 4), q())),
			func() voice.Reference { return voice.PresetReference(soprano()) }, failsAt(2, voice.ErrDifferentSystems)},
		{"settings are read before systems are compared", line(snd(n19(spelling.C(), 4), q()), snd(brokenNote(spelling.E(), 4), q())),
			func() voice.Reference { return voice.PresetReference(soprano()) }, failsAt(1, note.ErrUnsetSettings)},
		{"name not admissible in the system of the preset", line(snd(c4(), q()), snd(halfSharpC(), q())),
			func() voice.Reference { return voice.PresetReference(soprano()) }, failsAt(1, system.ErrInadmissible).only()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l := tt.l
			assertFailure(t, tt.want, func() (voice.Tessitura, error) { return l.Tessitura(tt.ref()) },
				func() voice.Tessitura { return l.MustTessitura(tt.ref()) })
			assertFailure(t, tt.want, func() ([]voice.SoundLocation, error) { return l.Locate(tt.ref()) },
				func() []voice.SoundLocation { return l.MustLocate(tt.ref()) })
			assertFailure(t, tt.want, func() (pair[bool, []int], error) { return two(l.Within(tt.ref())) },
				func() pair[bool, []int] { return mustTwo(l.MustWithin(tt.ref())) })
		})
	}

	// КП-44: with the explicit traditional system the line is compared.
	got, err := line19.TessituraIn(traditional(), voice.InstrumentReference(sopranoInstrument()))
	require.NoError(t, err)
	assert.Equal(t, "2/19", got.Number.String())
	assert.Equal(t, voice.TessituraLow, got.Kind)

	// With an explicit system the sources of notes are not read.
	broken := line(snd(brokenNote(spelling.C(), 4), q()), snd(brokenNote(spelling.E(), 4), q()))
	got = broken.MustTessituraIn(traditional(), voice.PresetReference(soprano()))
	assert.Equal(t, "2/19", got.Number.String())

	// The explicit system has priority over the system of the preset: in the
	// 19-position system C4–G5 is 30 steps wide and E4 is 6 steps above C4.
	std := line(snd(c4(), q()), snd(e4(), q()))
	got = std.MustTessituraByIn(system.Positions19(), voice.PresetReference(soprano()), voice.ByCount, voice.DefaultZones())
	assert.Equal(t, "1/10", got.Number.String())
	assert.Equal(t, voice.TessituraLow, got.Kind)
}

// КП-22, КП-23 and the order of checks of architecture 6.10.
func TestReferenceFailures(t *testing.T) {
	t.Parallel()
	std := line(snd(c4(), q()), snd(e4(), q()))
	freeLine := line(snd(free(spelling.C(), 4), q()))
	badName := line(snd(free(spelling.C(), 4), q()), snd(halfSharpC(), q()))
	inverted := func() voice.Reference {
		return voice.RangeReference(newRange(pt(spelling.G(), 5), pt(spelling.C(), 4)))
	}
	point := func() voice.Reference {
		return voice.RangeReference(newRange(pt(spelling.C(), 4), pt(spelling.C(), 4)))
	}
	inadmissible := func() voice.Reference {
		return voice.RangeReference(newRange(halfSharpC().MustPitch(), pt(spelling.G(), 5)))
	}
	presetNoRange := func() voice.Reference { return voice.PresetReference(soprano().MustWithoutRange()) }
	instrumentNoRange := func() voice.Reference {
		i := sopranoInstrument()
		if err := i.RemoveRange(); err != nil {
			panic(err)
		}
		return voice.InstrumentReference(i)
	}
	var unsetLine *voice.Line

	tests := []struct {
		name string
		l    *voice.Line
		sys  *system.System // nil with in == false: the form without a system
		in   bool
		ref  func() voice.Reference
		// what each family of operations returns
		locate    want
		tessitura want
	}{
		{"КП-23 preset without a range", std, nil, false, presetNoRange,
			fails(voice.ErrNoRange).only(), fails(voice.ErrNoRange).only()},
		{"КП-23 instrument without a range", std, nil, false, instrumentNoRange,
			fails(voice.ErrNoRange).only(), fails(voice.ErrNoRange).only()},
		{"КП-23 unset range", std, nil, false, func() voice.Reference { return voice.RangeReference(instrument.Range{}) },
			fails(voice.ErrNoRange).only(), fails(voice.ErrNoRange).only()},
		{"КП-23 unset preset", std, nil, false, func() voice.Reference { return voice.PresetReference(&instrument.Preset{}) },
			fails(voice.ErrUnset, instrument.ErrUnset), fails(voice.ErrUnset, instrument.ErrUnset)},
		{"nil preset", std, nil, false, func() voice.Reference { return voice.PresetReference(nil) },
			fails(voice.ErrUnset, instrument.ErrUnset), fails(voice.ErrUnset, instrument.ErrUnset)},
		{"nil instrument", std, nil, false, func() voice.Reference { return voice.InstrumentReference(nil) },
			fails(voice.ErrUnset, instrument.ErrUnset), fails(voice.ErrUnset, instrument.ErrUnset)},
		{"unset instrument", std, nil, false, func() voice.Reference { return voice.InstrumentReference(&instrument.Instrument{}) },
			fails(voice.ErrUnset, instrument.ErrUnset), fails(voice.ErrUnset, instrument.ErrUnset)},
		{"unset reference", std, nil, false, func() voice.Reference { return voice.Reference{} },
			fails(voice.ErrUnset).only(), fails(voice.ErrUnset).only()},
		{"placement out of the list", std, nil, false, func() voice.Reference { return voice.PlacementReference(voice.Placement(42)) },
			fails(voice.ErrUnset).only(), fails(voice.ErrUnset).only()},
		{"unset placement", std, nil, false, func() voice.Reference { return voice.PlacementReference(voice.PlacementUnset) },
			fails(voice.ErrNoRange).only(), fails(voice.ErrNoRange).only()},
		{"placement without a range", std, nil, false, func() voice.Reference { return voice.PlacementReference(voice.Upper) },
			fails(voice.ErrNoRange).only(), fails(voice.ErrNoRange).only()},
		{"inverted range", std, nil, false, inverted,
			fails(instrument.ErrInverted), fails(instrument.ErrInverted)},
		{"inadmissible bound", std, nil, false, inadmissible,
			fails(instrument.ErrInadmissible, system.ErrInadmissible), fails(instrument.ErrInadmissible, system.ErrInadmissible)},
		{"system is chosen before the range is checked", freeLine, nil, false, inverted,
			failsAt(0, voice.ErrNoSystem), failsAt(0, voice.ErrNoSystem)},
		{"explicit system and an inverted range", freeLine, traditional(), true, inverted,
			fails(instrument.ErrInverted), fails(instrument.ErrInverted)},
		{"range is checked before the sounds", badName, traditional(), true, inverted,
			fails(instrument.ErrInverted), fails(instrument.ErrInverted)},
		{"КП-22 zero width", std, nil, false, point,
			want{}, fails(voice.ErrZeroWidth).only()},
		{"zero width before the sounds", badName, traditional(), true, point,
			failsAt(1, system.ErrInadmissible).only(), fails(voice.ErrZeroWidth).only()},
		{"system before zero width", freeLine, nil, false, point,
			failsAt(0, voice.ErrNoSystem), failsAt(0, voice.ErrNoSystem)},
		{"unset system", std, nil, true, func() voice.Reference { return voice.Reference{} },
			fails(voice.ErrUnset, system.ErrInvalid), fails(voice.ErrUnset, system.ErrInvalid)},
		{"unset line", unsetLine, nil, true, func() voice.Reference { return voice.Reference{} },
			fails(voice.ErrUnset).only(), fails(voice.ErrUnset).only()},
		{"unset line without a system", unsetLine, nil, false, func() voice.Reference { return voice.Reference{} },
			fails(voice.ErrUnset).only(), fails(voice.ErrUnset).only()},
		{"no range before no sounds", line(), nil, false, func() voice.Reference { return voice.RangeReference(instrument.Range{}) },
			fails(voice.ErrNoRange).only(), fails(voice.ErrNoRange).only()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l, sys := tt.l, tt.sys
			locate := func() ([]voice.SoundLocation, error) {
				if tt.in {
					return l.LocateIn(sys, tt.ref())
				}
				return l.Locate(tt.ref())
			}
			mustLocate := func() []voice.SoundLocation {
				if tt.in {
					return l.MustLocateIn(sys, tt.ref())
				}
				return l.MustLocate(tt.ref())
			}
			within := func() (pair[bool, []int], error) {
				if tt.in {
					return two(l.WithinIn(sys, tt.ref()))
				}
				return two(l.Within(tt.ref()))
			}
			mustWithin := func() pair[bool, []int] {
				if tt.in {
					return mustTwo(l.MustWithinIn(sys, tt.ref()))
				}
				return mustTwo(l.MustWithin(tt.ref()))
			}
			tessitura := func() (voice.Tessitura, error) {
				if tt.in {
					return l.TessituraByIn(sys, tt.ref(), voice.ByCount, voice.DefaultZones())
				}
				return l.TessituraBy(tt.ref(), voice.ByCount, voice.DefaultZones())
			}
			mustTessitura := func() voice.Tessitura {
				if tt.in {
					return l.MustTessituraByIn(sys, tt.ref(), voice.ByCount, voice.DefaultZones())
				}
				return l.MustTessituraBy(tt.ref(), voice.ByCount, voice.DefaultZones())
			}
			if tt.locate.is == nil {
				_, err := locate()
				require.NoError(t, err)
				_, err = within()
				require.NoError(t, err)
			} else {
				assertFailure(t, tt.locate, locate, mustLocate)
				assertFailure(t, tt.locate, within, mustWithin)
			}
			assertFailure(t, tt.tessitura, tessitura, mustTessitura)
		})
	}
}

// Architecture 6.10: a line without sounds gives empty results for Locate
// and Within without a choice of system and without a check of the range in
// a system; the tessitura needs sounds (Е-8).
func TestReferenceWithoutSounds(t *testing.T) {
	t.Parallel()
	inverted := voice.RangeReference(newRange(pt(spelling.G(), 5), pt(spelling.C(), 4)))
	for name, l := range map[string]*voice.Line{"empty": line(), "rests": line(rst(q()))} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, err := l.Locate(inverted)
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Empty(t, got)
			within, outside, err := l.Within(inverted)
			require.NoError(t, err)
			assert.True(t, within)
			require.NotNil(t, outside)
			assert.Empty(t, outside)

			assertFailure(t, fails(voice.ErrNoSounds).only(),
				func() (voice.Tessitura, error) { return l.Tessitura(voice.PresetReference(soprano())) },
				func() voice.Tessitura { return l.MustTessitura(voice.PresetReference(soprano())) })
			// Е-8 before Е-10.
			point := voice.RangeReference(newRange(pt(spelling.C(), 4), pt(spelling.C(), 4)))
			assertFailure(t, fails(voice.ErrNoSounds).only(),
				func() (voice.Tessitura, error) { return l.Tessitura(point) },
				func() voice.Tessitura { return l.MustTessitura(point) })
		})
	}
}

// Architecture 6.10: the parameters of the tessitura are checked after the
// reference and before Е-9.
func TestTessituraParameterFailures(t *testing.T) {
	t.Parallel()
	l := line(snd(c4(), q()))
	noRange := voice.RangeReference(instrument.Range{})
	tests := []struct {
		name string
		ref  voice.Reference
		w    voice.Weighting
		z    voice.Zones
		want want
	}{
		{"unset weighting", voice.PresetReference(soprano()), voice.WeightingUnset, voice.DefaultZones(), fails(voice.ErrUnset).only()},
		{"weighting out of the list", voice.PresetReference(soprano()), voice.Weighting(42), voice.DefaultZones(), fails(voice.ErrUnset).only()},
		{"unset zones", voice.PresetReference(soprano()), voice.ByLength, voice.Zones{}, fails(voice.ErrUnset).only()},
		{"weighting before Е-9", noRange, voice.WeightingUnset, voice.DefaultZones(), fails(voice.ErrUnset).only()},
		{"zones before Е-9", noRange, voice.ByLength, voice.Zones{}, fails(voice.ErrUnset).only()},
		{"reference before weighting", voice.PresetReference(nil), voice.WeightingUnset, voice.Zones{}, fails(voice.ErrUnset, instrument.ErrUnset)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertFailure(t, tt.want,
				func() (voice.Tessitura, error) { return l.TessituraBy(tt.ref, tt.w, tt.z) },
				func() voice.Tessitura { return l.MustTessituraBy(tt.ref, tt.w, tt.z) })
			assertFailure(t, tt.want,
				func() (voice.Tessitura, error) { return l.TessituraByIn(traditional(), tt.ref, tt.w, tt.z) },
				func() voice.Tessitura { return l.MustTessituraByIn(traditional(), tt.ref, tt.w, tt.z) })
		})
	}
}

// Architecture 6.10: Reference values.
func TestReference(t *testing.T) {
	t.Parallel()
	refs := map[string]voice.Reference{
		"range":          voice.RangeReference(soprano().Range()),
		"unset range":    voice.RangeReference(instrument.Range{}),
		"preset":         voice.PresetReference(soprano()),
		"nil preset":     voice.PresetReference(nil),
		"unset preset":   voice.PresetReference(&instrument.Preset{}),
		"instrument":     voice.InstrumentReference(sopranoInstrument()),
		"nil instrument": voice.InstrumentReference(nil),
		"placement":      voice.PlacementReference(voice.Soprano),
		"no placement":   voice.PlacementReference(voice.PlacementUnset),
		"bad placement":  voice.PlacementReference(voice.Placement(42)),
	}
	for name, ref := range refs {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.True(t, ref.IsSet())
			assert.NotEmpty(t, ref.String())
			assert.NotEqual(t, "Reference(unset)", ref.String())
		})
	}
	var unset voice.Reference
	assert.False(t, unset.IsSet())
	assert.Equal(t, "Reference(unset)", unset.String())
}

// Architecture 6.14: the strings of Weighting and TessituraKind.
func TestTessituraEnumerations(t *testing.T) {
	t.Parallel()
	weightings := []struct {
		w     voice.Weighting
		valid bool
		text  string
	}{
		{voice.WeightingUnset, false, "Weighting(unset)"},
		{voice.ByLength, true, "by-length"},
		{voice.ByCount, true, "by-count"},
		{voice.Weighting(42), false, "Weighting(invalid)"},
	}
	for _, tt := range weightings {
		assert.Equal(t, tt.valid, tt.w.IsValid(), tt.text)
		assert.Equal(t, tt.text, tt.w.String())
	}
	kinds := []struct {
		k     voice.TessituraKind
		valid bool
		text  string
	}{
		{voice.TessituraUnset, false, "TessituraKind(unset)"},
		{voice.TessituraLow, true, "low"},
		{voice.TessituraMiddle, true, "middle"},
		{voice.TessituraHigh, true, "high"},
		{voice.TessituraKind(42), false, "TessituraKind(invalid)"},
	}
	for _, tt := range kinds {
		assert.Equal(t, tt.valid, tt.k.IsValid(), tt.text)
		assert.Equal(t, tt.text, tt.k.String())
	}
}

// The distance outside the range is counted in steps of the system of the
// operation.
func TestLocateInOtherSystem(t *testing.T) {
	t.Parallel()
	preset := instrument.MustNewPreset("Voice 19", settings19(), newRange(pt(spelling.C(), 4), pt(spelling.E(), 4)))
	l := line(snd(n19(spelling.B(), 3), q()), snd(n19(spelling.D(), 4), q()), snd(n19(spelling.F(), 4), q()))
	got := l.MustLocate(voice.PresetReference(preset))
	assertLocations(t, []location{{0, instrument.Below, 2}, {1, instrument.Inside, 0}, {2, instrument.Above, 2}}, got)
	tess := l.MustTessituraBy(voice.PresetReference(preset), voice.ByCount, voice.DefaultZones())
	assert.Equal(t, "1/2", tess.Number.String(), "(−2 + 3 + 8) / (3 · 6)")
}

// Architecture 6.10, 6.16: the distance of a sound out of the range of steps
// is the failure of that sound for Locate and the tessitura; Within computes
// no distances and does not fail.
func TestDistanceOutOfRange(t *testing.T) {
	t.Parallel()
	far := nt(spelling.C(), 1<<60)
	l := line(snd(c4(), q()), rst(q()), snd(far, q()))
	ref := voice.PresetReference(soprano())
	w := failsAt(2, system.ErrOutOfRange)

	assertFailure(t, w, func() ([]voice.SoundLocation, error) { return l.Locate(ref) },
		func() []voice.SoundLocation { return l.MustLocate(ref) })
	assertFailure(t, w, func() ([]voice.SoundLocation, error) { return l.LocateIn(traditional(), ref) },
		func() []voice.SoundLocation { return l.MustLocateIn(traditional(), ref) })
	assertFailure(t, w, func() (voice.Tessitura, error) { return l.Tessitura(ref) },
		func() voice.Tessitura { return l.MustTessitura(ref) })
	assertFailure(t, w, func() (voice.Tessitura, error) {
		return l.TessituraByIn(traditional(), ref, voice.ByCount, voice.DefaultZones())
	},
		func() voice.Tessitura {
			return l.MustTessituraByIn(traditional(), ref, voice.ByCount, voice.DefaultZones())
		})

	within, outside, err := l.Within(ref)
	require.NoError(t, err)
	assert.False(t, within)
	assert.Equal(t, []int{2}, outside)
	within, outside, err = l.WithinIn(traditional(), ref)
	require.NoError(t, err)
	assert.False(t, within)
	assert.Equal(t, []int{2}, outside)
}

// Architecture 6.10: a sum of the tessitura out of the range of
// common/fraction fails without *voice.ItemError. With the lengths 1/p and
// (p−1)/p the number is D·(p−1)/(19·p) for the distance D of the high sound:
// its numerator has about 262 bits.
func TestTessituraSumOutOfRange(t *testing.T) {
	t.Parallel()
	invP, restP, _ := overflowDurations()
	l := line(snd(c4(), voice.MustValueDuration(invP)), snd(nt(spelling.C(), 1<<58), voice.MustValueDuration(restP)))
	ref := voice.PresetReference(soprano())
	assertFailure(t, fails(fraction.ErrOutOfRange), func() (voice.Tessitura, error) { return l.Tessitura(ref) },
		func() voice.Tessitura { return l.MustTessitura(ref) })
	assertFailure(t, fails(fraction.ErrOutOfRange),
		func() (voice.Tessitura, error) { return l.TessituraIn(traditional(), ref) },
		func() voice.Tessitura { return l.MustTessituraIn(traditional(), ref) })

	// Weighted by count the number is representable.
	_, err := l.TessituraBy(ref, voice.ByCount, voice.DefaultZones())
	require.NoError(t, err)
}

// Architecture 6.10 (as clarified for the move of the tessitura sums to
// fraction.Sum): the failure of the sums is decided only by the final
// values — the product wᵢ·D(xᵢ) of one sound, the final sums Σ wᵢ·D(xᵢ) and
// Σ wᵢ, the denominator W·Σ wᵢ and the final r — without *voice.ItemError.
// Each input below overflows exactly one of these values and keeps the
// others in range, so a check of r alone would not refuse the first three.
// M is fraction.MaxComponent; the soprano reference has W = 19, C4 is at
// D = 0, D4 at 2, E4 at 4 and C6 at 24.
func TestTessituraFinalValueFailures(t *testing.T) {
	t.Parallel()
	m := fraction.MaxComponentBig()
	one := big.NewInt(1)
	vd := func(num, den *big.Int) voice.Duration { return voice.MustValueDuration(bigDuration(num, den)) }
	ref := voice.PresetReference(soprano())

	// Σ wᵢ·D(xᵢ): a with M/24 < a ≤ M/19 and 5 ∤ a, split into x + y with
	// 5 ∤ x, y.
	five := big.NewInt(5)
	a := new(big.Int).Quo(m, big.NewInt(20))
	for new(big.Int).Mod(a, five).Sign() == 0 {
		a.Sub(a, one)
	}
	x := new(big.Int).Rsh(a, 1)
	for new(big.Int).Mod(x, five).Sign() == 0 || new(big.Int).Mod(new(big.Int).Sub(a, x), five).Sign() == 0 {
		x.Sub(x, one)
	}
	y := new(big.Int).Sub(a, x)

	// Σ wᵢ: s and t are coprime, 19 ∤ t and M/19 < s·t ≤ M.
	s := new(big.Int).Add(new(big.Int).Lsh(one, 127), one)
	tt := new(big.Int).Add(s, big.NewInt(2))
	s19 := new(big.Int).Mul(s, big.NewInt(19))

	// r: Q even with M/361 < Q and 19·Q + 1 ≤ M.
	bigQ := new(big.Int).Quo(m, big.NewInt(100))
	bigQ.Sub(bigQ, new(big.Int).And(bigQ, one))
	q19 := new(big.Int).Mul(bigQ, big.NewInt(19))

	halfMax := vd(m, big.NewInt(2))
	tests := []struct {
		name string
		l    *voice.Line
	}{
		// w₁·D = 4·(M−1)/M is out of range; Σ wᵢ·D(xᵢ) = 4, Σ wᵢ = 1,
		// W·Σ wᵢ = 19 and r = 4/19 are not.
		{"product of one sound", line(snd(e4(), vd(new(big.Int).Sub(m, one), m)), snd(e4(), vd(one, m)))},
		// Σ wᵢ·D(xᵢ) = 24·a/5 is out of range; the products 24·x/5 and
		// 24·y/5, Σ wᵢ = a/5, W·Σ wᵢ = 19·a/5 and r = 24/19 are not.
		{"final numerator sum", line(
			snd(nt(spelling.C(), 6), vd(x, five)),
			snd(nt(spelling.C(), 6), vd(y, five)),
		)},
		// Σ wᵢ = (19·s+t)/(19·s·t) is out of range; W·Σ wᵢ = (19·s+t)/(s·t),
		// Σ wᵢ·D(xᵢ) = 0 and r = 0 are not. The rest carries no weight and
		// keeps the positions 0, 1/(19·s), 1, 1+1/t representable.
		{"final weight sum", line(
			snd(c4(), vd(one, s19)),
			rst(vd(new(big.Int).Sub(s19, one), s19)),
			snd(c4(), vd(one, tt)),
		)},
		// W·Σ wᵢ = 19·M is out of range; Σ wᵢ = M, Σ wᵢ·D(xᵢ) = 0 and r = 0
		// are not.
		{"final denominator", line(snd(c4(), halfMax), snd(c4(), halfMax))},
		// r = 2/(19·(19·Q+1)) is out of range; Σ wᵢ = (19·Q+1)/(19·Q),
		// W·Σ wᵢ = (19·Q+1)/Q, Σ wᵢ·D(xᵢ) = 2/(19·Q) and the products are not.
		{"final r", line(snd(c4(), w()), snd(d4(), vd(one, q19)))},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assertFailure(t, fails(fraction.ErrOutOfRange),
				func() (voice.Tessitura, error) { return tc.l.Tessitura(ref) },
				func() voice.Tessitura { return tc.l.MustTessitura(ref) })
			assertFailure(t, fails(fraction.ErrOutOfRange),
				func() (voice.Tessitura, error) { return tc.l.TessituraIn(traditional(), ref) },
				func() voice.Tessitura { return tc.l.MustTessituraIn(traditional(), ref) })
		})
	}
}

// Architecture 6.10 (as clarified for the move of the tessitura sums to
// fraction.Sum): partial sums out of range do not fail the tessitura when
// every final value is in range, and the result does not depend on the
// order of the sounds. The three sounds stand at the same height (E4 is 4
// steps above C4) with the lengths 1/p, 1/q, (p−1)/p in the direct order or
// (p−1)/p, 1/q, 1/p in the reversed one: the partial sum of the first two
// has a denominator near 2^400, while Σ wᵢ = (q+1)/q, Σ wᵢ·D(xᵢ) = 4·(q+1)/q,
// W·Σ wᵢ and r = 4/19 are all in range. The rests between the sounds carry
// no weight and keep every position of the line representable. Before the
// sums move to fraction.Sum this test fails: the sequential accumulation
// refuses at the second sound with fraction.ErrOutOfRange.
func TestTessituraPartialSumsBeyondRange(t *testing.T) {
	t.Parallel()
	invP, restP, invQ := overflowDurations()
	_, q := bigDenominators()
	restQ := voice.MustValueDuration(bigDuration(new(big.Int).Sub(q, big.NewInt(1)), q))
	ref := voice.PresetReference(soprano())

	tests := []struct {
		name string
		l    *voice.Line
	}{
		// Positions 0, 1/p, 1, 1+1/q, 2, 2+(p−1)/p.
		{"direct", line(
			snd(e4(), voice.MustValueDuration(invP)), rst(voice.MustValueDuration(restP)),
			snd(e4(), voice.MustValueDuration(invQ)), rst(restQ),
			snd(e4(), voice.MustValueDuration(restP)),
		)},
		// Positions 0, (p−1)/p, 1, 1+1/q, 2, 2+1/p.
		{"reversed", line(
			snd(e4(), voice.MustValueDuration(restP)), rst(voice.MustValueDuration(invP)),
			snd(e4(), voice.MustValueDuration(invQ)), rst(restQ),
			snd(e4(), voice.MustValueDuration(invP)),
		)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l := tt.l

			got, err := l.Tessitura(ref)
			require.NoError(t, err)
			assert.Equal(t, "4/19", got.Number.String())
			assert.Equal(t, voice.TessituraLow, got.Kind)
			assertTessitura(t, got, l.MustTessitura(ref))

			in, err := l.TessituraIn(traditional(), ref)
			require.NoError(t, err)
			assertTessitura(t, got, in)
			assertTessitura(t, got, l.MustTessituraIn(traditional(), ref))

			// Weighted by count the number is the same and needs no big sums.
			byCount, err := l.TessituraBy(ref, voice.ByCount, voice.DefaultZones())
			require.NoError(t, err)
			assertTessitura(t, got, byCount)
		})
	}
}
