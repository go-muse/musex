package voice_test

import (
	"testing"

	"github.com/go-muse/musex/instrument"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/voice"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roleName struct {
	english, russian string
}

// Architecture 6.14, П-12: the closed lists, their order and names.
func TestPlacementNames(t *testing.T) {
	t.Parallel()
	want := map[voice.Placement]roleName{
		voice.Soprano: {"soprano", "сопрано"},
		voice.Alto:    {"alto", "альт"},
		voice.Tenor:   {"tenor", "тенор"},
		voice.Bass:    {"bass", "бас"},
		voice.Upper:   {"upper", "верхний"},
		voice.Middle:  {"middle", "средний"},
		voice.Lower:   {"lower", "нижний"},
	}
	all := voice.Placements()
	assert.Equal(t, []voice.Placement{voice.Soprano, voice.Alto, voice.Tenor, voice.Bass, voice.Upper, voice.Middle, voice.Lower}, all)
	for _, p := range all {
		t.Run(want[p].english, func(t *testing.T) {
			t.Parallel()
			assert.True(t, p.IsValid())
			assert.Equal(t, want[p].english, p.Name())
			assert.Equal(t, want[p].russian, p.RussianName())
			assert.Equal(t, want[p].english, p.String())
			for _, text := range []string{want[p].english, want[p].russian} {
				got, err := voice.ParsePlacement(text)
				require.NoError(t, err)
				assert.Equal(t, p, got)
				assert.Equal(t, p, voice.MustParsePlacement(text))
			}
		})
	}
}

func TestThematicRoleNames(t *testing.T) {
	t.Parallel()
	want := map[voice.ThematicRole]roleName{
		voice.Leading:      {"leading", "ведущий"},
		voice.Solo:         {"solo", "сольный"},
		voice.Accompanying: {"accompanying", "сопровождающий"},
		voice.Imitating:    {"imitating", "имитирующий"},
		voice.Counterpoint: {"counterpoint", "контрапункт"},
		voice.Doubling:     {"doubling", "удвоение"},
		voice.Inversion:    {"inversion", "противодвижение"},
		voice.Pedal:        {"pedal", "педаль"},
	}
	all := voice.ThematicRoles()
	assert.Equal(t, []voice.ThematicRole{voice.Leading, voice.Solo, voice.Accompanying, voice.Imitating,
		voice.Counterpoint, voice.Doubling, voice.Inversion, voice.Pedal}, all)
	for _, r := range all {
		t.Run(want[r].english, func(t *testing.T) {
			t.Parallel()
			assert.True(t, r.IsValid())
			assert.Equal(t, want[r].english, r.Name())
			assert.Equal(t, want[r].russian, r.RussianName())
			assert.Equal(t, want[r].english, r.String())
			for _, text := range []string{want[r].english, want[r].russian} {
				got, err := voice.ParseThematicRole(text)
				require.NoError(t, err)
				assert.Equal(t, r, got)
				assert.Equal(t, r, voice.MustParseThematicRole(text))
			}
		})
	}
}

func TestImportanceNames(t *testing.T) {
	t.Parallel()
	want := map[voice.Importance]roleName{
		voice.Principal:     {"principal", "главный"},
		voice.Subsidiary:    {"subsidiary", "побочный"},
		voice.Supplementary: {"supplementary", "дополняющий"},
		voice.Filler:        {"filler", "заполняющий"},
		voice.Harmonic:      {"harmonic", "гармонический"},
	}
	all := voice.Importances()
	assert.Equal(t, []voice.Importance{voice.Principal, voice.Subsidiary, voice.Supplementary, voice.Filler, voice.Harmonic}, all)
	for _, r := range all {
		t.Run(want[r].english, func(t *testing.T) {
			t.Parallel()
			assert.True(t, r.IsValid())
			assert.Equal(t, want[r].english, r.Name())
			assert.Equal(t, want[r].russian, r.RussianName())
			assert.Equal(t, want[r].english, r.String())
			for _, text := range []string{want[r].english, want[r].russian} {
				got, err := voice.ParseImportance(text)
				require.NoError(t, err)
				assert.Equal(t, r, got)
				assert.Equal(t, r, voice.MustParseImportance(text))
			}
		})
	}
}

// The lists are new slices on each call.
func TestRoleListsAreCopies(t *testing.T) {
	t.Parallel()
	p := voice.Placements()
	p[0] = voice.Bass
	assert.Equal(t, voice.Soprano, voice.Placements()[0])
	r := voice.ThematicRoles()
	r[0] = voice.Pedal
	assert.Equal(t, voice.Leading, voice.ThematicRoles()[0])
	i := voice.Importances()
	i[0] = voice.Harmonic
	assert.Equal(t, voice.Principal, voice.Importances()[0])
}

// Unset and out-of-list values (architecture 6.1, 6.14).
func TestRoleUnsetAndInvalidValues(t *testing.T) {
	t.Parallel()
	assert.False(t, voice.PlacementUnset.IsValid())
	assert.Empty(t, voice.PlacementUnset.Name())
	assert.Empty(t, voice.PlacementUnset.RussianName())
	assert.Equal(t, "Placement(unset)", voice.PlacementUnset.String())
	assert.False(t, voice.Placement(42).IsValid())
	assert.Empty(t, voice.Placement(42).Name())
	assert.Empty(t, voice.Placement(42).RussianName())
	assert.Equal(t, "Placement(invalid)", voice.Placement(42).String())

	assert.False(t, voice.ThematicUnset.IsValid())
	assert.Empty(t, voice.ThematicUnset.Name())
	assert.Empty(t, voice.ThematicUnset.RussianName())
	assert.Equal(t, "ThematicRole(unset)", voice.ThematicUnset.String())
	assert.False(t, voice.ThematicRole(42).IsValid())
	assert.Empty(t, voice.ThematicRole(42).Name())
	assert.Equal(t, "ThematicRole(invalid)", voice.ThematicRole(42).String())

	assert.False(t, voice.ImportanceUnset.IsValid())
	assert.Empty(t, voice.ImportanceUnset.Name())
	assert.Empty(t, voice.ImportanceUnset.RussianName())
	assert.Equal(t, "Importance(unset)", voice.ImportanceUnset.String())
	assert.False(t, voice.Importance(42).IsValid())
	assert.Empty(t, voice.Importance(42).Name())
	assert.Equal(t, "Importance(invalid)", voice.Importance(42).String())
}

// КП-38, КП-52, Е-12: parsing is exact and distinguishes the kind.
func TestParseRoleFailures(t *testing.T) {
	t.Parallel()
	w := fails(voice.ErrUnknownName).only()
	placements := []string{"Tenor ", "Tenor", "TENOR", " tenor", "", "principal", "leading", "Тенор", "soprano voice"}
	for _, text := range placements {
		t.Run("placement "+text, func(t *testing.T) {
			t.Parallel()
			assertFailure(t, w,
				func() (voice.Placement, error) { return voice.ParsePlacement(text) },
				func() voice.Placement { return voice.MustParsePlacement(text) })
		})
	}
	thematic := []string{"principal", "главный", "soprano", "Leading", "", "inversion "}
	for _, text := range thematic {
		t.Run("thematic "+text, func(t *testing.T) {
			t.Parallel()
			assertFailure(t, w,
				func() (voice.ThematicRole, error) { return voice.ParseThematicRole(text) },
				func() voice.ThematicRole { return voice.MustParseThematicRole(text) })
		})
	}
	importance := []string{"leading", "soprano", "Principal", "", "главный "}
	for _, text := range importance {
		t.Run("importance "+text, func(t *testing.T) {
			t.Parallel()
			assertFailure(t, w,
				func() (voice.Importance, error) { return voice.ParseImportance(text) },
				func() voice.Importance { return voice.MustParseImportance(text) })
		})
	}
}

// КП-38.
func TestParseTenor(t *testing.T) {
	t.Parallel()
	assert.Equal(t, voice.Tenor, voice.MustParsePlacement("tenor"))
	assert.Equal(t, voice.Tenor, voice.MustParsePlacement("тенор"))
	// КП-52.
	assert.Equal(t, voice.Principal, voice.MustParseImportance("principal"))
}

// КП-51, ОП-49: outer and inner placements.
func TestPlacementOuterInner(t *testing.T) {
	t.Parallel()
	tests := []struct {
		p            voice.Placement
		outer, inner bool
	}{
		{voice.Soprano, true, false},
		{voice.Alto, false, true},
		{voice.Tenor, false, true},
		{voice.Bass, true, false},
		{voice.Upper, true, false},
		{voice.Middle, false, true},
		{voice.Lower, true, false},
		{voice.PlacementUnset, false, false},
		{voice.Placement(42), false, false},
	}
	for _, tt := range tests {
		t.Run(tt.p.String(), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.outer, tt.p.IsOuter())
			assert.Equal(t, tt.inner, tt.p.IsInner())
		})
	}
}

// КП-21, ОП-26: the reference range of a placement is the range of the
// built-in voice preset of instrument.
func TestPlacementRange(t *testing.T) {
	t.Parallel()
	tests := []struct {
		p      voice.Placement
		preset *instrument.Preset
	}{
		{voice.Soprano, instrument.SopranoVoice()},
		{voice.Alto, instrument.AltoVoice()},
		{voice.Tenor, instrument.TenorVoice()},
		{voice.Bass, instrument.BassVoice()},
		{voice.Upper, nil},
		{voice.Middle, nil},
		{voice.Lower, nil},
		{voice.PlacementUnset, nil},
		{voice.Placement(42), nil},
	}
	for _, tt := range tests {
		t.Run(tt.p.String(), func(t *testing.T) {
			t.Parallel()
			r, ok := tt.p.Range()
			if tt.preset == nil {
				assert.False(t, ok)
				assert.False(t, r.IsValid())
				return
			}
			require.True(t, ok)
			assert.True(t, tt.preset.Range().Equal(r), "want %v, got %v", tt.preset.Range(), r)
		})
	}

	r, ok := voice.Soprano.Range()
	require.True(t, ok)
	assert.True(t, newRange(pt(spelling.C(), 4), pt(spelling.G(), 5)).Equal(r), "КП-21: C4–G5, got %v", r)
}
