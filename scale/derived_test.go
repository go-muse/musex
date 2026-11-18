package scale_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pcset"
	"github.com/go-muse/musex/scale"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPitchClassesAndKeySignatures(t *testing.T) {
	t.Parallel()
	s := scale.MustNew(at("C+1", 4, cents.Zero(), nil), at("D-1", 5, cents.MustFromInt(30), nil), names("C+1")[0])
	got, err := s.PitchClasses(system.Traditional())
	require.NoError(t, err)
	assert.True(t, pcset.MustNew(12, 1).Equal(got))
	other, err := list("C D G").PitchClasses(system.Positions19())
	require.NoError(t, err)
	assert.Equal(t, 19, other.Modulus())
	assert.Equal(t, 3, other.Cardinality())
	empty, err := scale.MustNew().PitchClasses(system.Positions19())
	require.NoError(t, err)
	assert.True(t, empty.IsValid())
	assert.True(t, empty.IsEmpty())
	assert.Equal(t, 19, empty.Modulus())
	_, err = s.PitchClasses(nil)
	assert.ErrorIs(t, err, scale.ErrUnset)
	assert.ErrorIs(t, err, system.ErrInvalid)
	_, err = list("C+1/2").PitchClasses(system.Traditional())
	assert.ErrorIs(t, err, system.ErrInadmissible)
	for _, tt := range []struct {
		name, notes, signs string
		index              int
		ok                 bool
	}{
		{"K30", "D E F+1 G A B C+1", "F+1 C+1", 2, true},
		{"K31", "F+1 G+1 A+1 B C+1 D+1 E", "F+1 C+1 G+1 D+1 A+1", 5, true},
		{"K32", "C D E F G A B C", "", 0, true},
		{"K33", "E G A", "", 0, false},
		{"flat order", "F G A B-1 C D E", "B-1", -1, true},
		{"double sharps", "D+1 E+1 F+2 G+1 A+1 B+1 C+2", "F+2 C+2 G+1 D+1 A+1 E+1 B+1", 9, true},
		{"double flat", "F-1 G-1 A-1 B-2 C-1 D-1 E-1", "B-2 E-1 A-1 D-1 G-1 C-1 F-1", -8, true},
		{"fractional", "C D E F G A B+1/2", "", 0, false},
		{"not consecutive", "C D E F G A B+1", "", 0, false},
		{"empty", "", "", 0, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			sig, ok, err := list(tt.notes).KeySignature()
			require.NoError(t, err)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.index, sig.Index)
			if ok {
				require.NotNil(t, sig.Signs)
				actual := make([]string, 0, len(sig.Signs))
				for _, n := range sig.Signs {
					actual = append(actual, n.String())
				}
				assert.Equal(t, tt.signs, strings.Join(actual, " "))
				if len(sig.Signs) > 0 {
					sig.Signs[0] = names("C")[0].Spelling()
					fresh, _, err := list(tt.notes).KeySignature()
					require.NoError(t, err)
					assert.NotEqual(t, sig.Signs[0], fresh.Signs[0])
				}
			} else {
				assert.Nil(t, sig.Signs)
			}
		})
	}
}

func TestDebugAndNotationText(t *testing.T) {
	t.Parallel()
	var absent *scale.Scale
	assert.Equal(t, "Scale(unset)", absent.String())
	assert.Equal(t, "Scale(unset)", absent.GoString())
	assert.Equal(t, "Scale[]", scale.MustNew().String())
	src := settings.Standard()
	ns := []note.Note{at("G", 4, cents.MustFromInt(30), src), note.MustFromSpelling(names("A")[0].Spelling(), cents.Zero(), src)}
	s := scale.MustNew(ns...)
	wantDebug := "Scale[" + ns[0].String() + "; " + ns[1].String() + "]"
	assert.Equal(t, wantDebug, s.String())
	assert.Equal(t, wantDebug, s.GoString())
	assert.Equal(t, wantDebug, fmt.Sprintf("%#v", s))
	formatted := make([]string, len(ns))
	for i, n := range ns {
		formatted[i] = n.MustFormatIn(src) + " (" + n.Cents().Text(cents.WithUnit|cents.WithPlus) + ")"
	}
	want := "[" + strings.Join(formatted, "; ") + "]"
	got, err := s.Format()
	require.NoError(t, err)
	assert.Equal(t, want, got)
	got, err = s.FormatIn(src)
	require.NoError(t, err)
	assert.Equal(t, want, got)
	for _, explicit := range []bool{false, true} {
		var got string
		var err error
		if explicit {
			got, err = scale.MustNew(ns[0], names("G")[0]).FormatIn(src)
			require.NoError(t, err)
			assert.NotEmpty(t, got)
		} else {
			got, err = scale.MustNew(ns[0], names("G")[0]).Format()
			assert.Empty(t, got)
			assert.ErrorIs(t, err, note.ErrNoSource)
		}
	}
	got, err = scale.MustNew().Format()
	require.NoError(t, err)
	assert.Equal(t, "[]", got)
	got, err = scale.MustNew().FormatIn((*settings.Settings)(nil))
	require.NoError(t, err)
	assert.Equal(t, "[]", got)
	got, err = s.FormatIn((*settings.Settings)(nil))
	assert.Empty(t, got)
	assert.ErrorIs(t, err, note.ErrUnsetSettings)
	got, err = scale.MustNew().FormatIn(nil)
	assert.Empty(t, got)
	assert.ErrorIs(t, err, scale.ErrUnset)
	source := &source{set: src}
	n := at("G", 4, cents.Zero(), source)
	source.forbidden = true
	assert.NotEmpty(t, scale.MustNew(n).String())
	assert.NotEmpty(t, scale.MustNew(n).GoString())
	_, _, err = scale.MustNew(n).KeySignature()
	require.NoError(t, err)
	_, err = scale.MustNew(n).PitchClasses(system.Traditional())
	require.NoError(t, err)
}
