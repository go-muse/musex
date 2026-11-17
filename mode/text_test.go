package mode_test

import (
	"encoding"
	"fmt"
	"testing"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextFormsAndParse(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ input, short, ru string }{
		{"б.2 м.2", "M2 m2", "б.2 м.2"},
		{"м.3 б.2 |", "m3 M2 |", "м.3 б.2 |"},
		{"maj2 m2", "M2 m2", "б.2 м.2"},
		{"P8", "P8", "ч.8"},
		{"m3 M2 A1 m2 m3 M2", "m3 M2 A1 m2 m3 M2", "м.3 б.2 ув.1 м.2 м.3 б.2"},
		{"M2 n2", "M2 n2", ""},
		{"{+1/3}2 M2 |", "{+1/3}2 M2 |", ""},
	} {
		t.Run(tt.input, func(t *testing.T) {
			m, err := mode.Parse(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.short, m.String())
			assert.Equal(t, tt.short, m.MustIntervalsString())
			ru, ok, err := m.RussianIntervals()
			require.NoError(t, err)
			assert.Equal(t, tt.ru, ru)
			assert.Equal(t, tt.ru != "", ok)
			if ok {
				assert.True(t, m.Equal(mode.MustParse(ru)))
			}
			assert.True(t, m.Equal(mode.MustParse(tt.short)))
			assert.False(t, m.HasOrigin())
		})
	}
	for _, input := range []string{"", "M2,M2", "M2  M2", " M2 m7", "M2 M2 ", "M2\tM2", "M2-M2", "2-2-1", "2 2 1", "|", "| M2", "M2 | M2", "M2|", "M2 | |", "Dorian: " + dorian, "Dorian (edited): " + aeolian} {
		t.Run("bad syntax_"+input, func(t *testing.T) {
			m, err := mode.Parse(input)
			checkError(t, err, mode.ErrParse, template.ErrParse)
			assert.Zero(t, m)
		})
	}
	m, err := mode.Parse("M2 what")
	checkError(t, err, mode.ErrParse, template.ErrParse, interval.ErrParse)
	assert.Zero(t, m)
	for _, input := range []string{"M2 -M2 P8", "M2 P1", "M2 P1 |"} {
		m, err := mode.Parse(input)
		checkError(t, err, mode.ErrInvalidStructure, template.ErrInvalidInterval)
		assert.Zero(t, m)
	}
}

func TestOriginAndDebugText(t *testing.T) {
	t.Parallel()
	var zero mode.Mode
	assert.Equal(t, "Mode(unset)", zero.String())
	assert.Equal(t, "mode.Mode{unset}", fmt.Sprintf("%#v", zero))
	m := mode.MustFromCode("dorian")
	assert.Equal(t, "дорийский: "+dorian, m.String())
	assert.Equal(t, `mode.Mode{intervals:"M2 m2 M2 M2 M2 m2 M2", origin:"dorian", selectedName:"", changed:false}`, m.GoString())
	m = mode.MustFromName("DORIAN")
	assert.Equal(t, "Dorian: "+dorian, m.String())
	require.NoError(t, m.SetDegreeInterval(6, interval.MustParse("m6")))
	assert.Equal(t, "Dorian (edited): "+aeolian, m.String())
	assert.Equal(t, `mode.Mode{intervals:"M2 m2 M2 M2 m2 M2 M2", origin:"dorian", selectedName:"Dorian", changed:true}`, fmt.Sprintf("%#v", m))
	parsed := mode.MustParse(m.MustIntervalsString())
	assert.True(t, m.Equal(parsed))
	assert.False(t, parsed.HasOrigin())
	assert.Equal(t, `mode.Mode{intervals:"M2 m2 M2 M2 m2 M2 M2", origin:"", selectedName:"", changed:false}`, parsed.GoString())
	_, isMarshaler := any(m).(encoding.TextMarshaler)
	assert.False(t, isMarshaler)
	_, isUnmarshaler := any(&m).(encoding.TextUnmarshaler)
	assert.False(t, isUnmarshaler)
}
