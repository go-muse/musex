package instrument_test

import (
	"testing"

	"github.com/go-muse/musex/instrument"
	"github.com/go-muse/musex/settings"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ОП-18, ОП-19, П-7, П-1а, К-6, К-23, А 4.5.
func TestNewPreset(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		pname   string
		set     *settings.Settings
		r       instrument.Range
		wantErr error
	}{
		{name: "К-6 name, Н-0 and C4–G5", pname: "Lead", set: h0(), r: rng("C4", "G5")},
		{name: "К-6 Г-6 without a range", pname: "Lead", set: h0(), r: instrument.Range{}},
		{name: "КП-6 in Н-0", pname: "Lead", set: h0(), r: rangeKP6()},
		{name: "КП-6 in Н-19 is inverted", pname: "Lead", set: h19(), r: rangeKP6(), wantErr: instrument.ErrInverted},
		{name: "КП-6а in Н-19 is inverted", pname: "Lead", set: h19(), r: rangeKP6a(), wantErr: instrument.ErrInverted},
		{name: "КП-7 in Н-0 is inadmissible", pname: "Lead", set: h0(), r: rangeKP7(), wantErr: instrument.ErrInadmissible},
		{name: "КП-7 in Н-24", pname: "Lead", set: h24(), r: rangeKP7()},
		{name: "inverted in Н-0", pname: "Lead", set: h0(), r: rng("G5", "C4"), wantErr: instrument.ErrInverted},
		{name: "empty name", pname: "", set: h0(), r: rng("C4", "G5"), wantErr: instrument.ErrEmptyName},
		{name: "a space is a name", pname: " ", set: h0(), r: rng("C4", "G5")},
		{name: "nil set", pname: "Lead", set: nil, r: rng("C4", "G5"), wantErr: instrument.ErrUnset},
		{name: "zero set", pname: "Lead", set: &settings.Settings{}, r: rng("C4", "G5"), wantErr: instrument.ErrUnset},
		{name: "К-23 unset set and empty name", pname: "", set: nil, r: rng("C4", "G5"), wantErr: instrument.ErrUnset},
		{name: "К-23 empty name and inverted range", pname: "", set: h19(), r: rangeKP6(), wantErr: instrument.ErrEmptyName},
		{name: "inadmissible lower before inverted", pname: "Lead", set: h0(), r: instrument.MustNewRange(at(cHalf4().Spelling(), 6), p("C4")), wantErr: instrument.ErrInadmissible},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ps, err := instrument.NewPreset(tt.pname, tt.set, tt.r)
			if tt.wantErr != nil {
				assertCause(t, err, tt.wantErr)
				assert.Nil(t, ps)
				assertPanicsWith(t, tt.wantErr, err, func() { instrument.MustNewPreset(tt.pname, tt.set, tt.r) })
				return
			}
			require.NoError(t, err)
			require.NotNil(t, ps)
			assert.True(t, ps.IsValid())
			assert.Equal(t, tt.pname, ps.Name())
			assert.Same(t, tt.set, ps.SettingsSet(), "the preset keeps the pointer it was given (А 6)")
			assert.Equal(t, tt.r.IsValid(), ps.HasRange())
			assert.True(t, tt.r.Equal(ps.Range()))
			if !tt.r.IsValid() {
				assert.Equal(t, instrument.Range{}, ps.Range())
			}
			assert.True(t, ps.Equal(instrument.MustNewPreset(tt.pname, tt.set, tt.r)))
		})
	}
}

// П-18а, А 4.2: nil and &Preset{} are the unset preset.
func TestPreset_unset(t *testing.T) {
	t.Parallel()
	unset := map[string]*instrument.Preset{"nil": nil, "zero": {}}
	for name, ps := range unset {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assertNoPanic(t, func() {
				assert.False(t, ps.IsValid())
				assert.Empty(t, ps.Name())
				assert.Nil(t, ps.SettingsSet())
				assert.Equal(t, instrument.Range{}, ps.Range())
				assert.False(t, ps.HasRange())
				assert.Equal(t, "Preset(unset)", ps.String())
			})

			got, err := ps.WithName("Lead")
			assertCause(t, err, instrument.ErrUnset)
			assert.Nil(t, got)
			assertPanicsWith(t, instrument.ErrUnset, err, func() { ps.MustWithName("Lead") })

			got, err = ps.WithSettingsSet(h0())
			assertCause(t, err, instrument.ErrUnset)
			assert.Nil(t, got)
			assertPanicsWith(t, instrument.ErrUnset, err, func() { ps.MustWithSettingsSet(h0()) })

			got, err = ps.WithRange(rng("C4", "G5"))
			assertCause(t, err, instrument.ErrUnset)
			assert.Nil(t, got)
			assertPanicsWith(t, instrument.ErrUnset, err, func() { ps.MustWithRange(rng("C4", "G5")) })

			got, err = ps.WithoutRange()
			assertCause(t, err, instrument.ErrUnset)
			assert.Nil(t, got)
			assertPanicsWith(t, instrument.ErrUnset, err, func() { ps.MustWithoutRange() })

			// The order of ОП-20: the receiver first.
			_, err = ps.WithName("")
			assertCause(t, err, instrument.ErrUnset)
			_, err = ps.WithSettingsSet(nil)
			assertCause(t, err, instrument.ErrUnset)
			_, err = ps.WithRange(rangeKP7())
			assertCause(t, err, instrument.ErrUnset)
		})
	}
}

// ОП-20, К-7, П-7, А 4.5: a replacement gives a new preset; the source does
// not change.
func TestPreset_With(t *testing.T) {
	t.Parallel()
	base := func() *instrument.Preset { return preset("Lead", h0(), rangeKP6()) }
	tests := []struct {
		name    string
		with    func(ps *instrument.Preset) (*instrument.Preset, error)
		must    func(ps *instrument.Preset) *instrument.Preset
		want    *instrument.Preset
		wantErr error
	}{
		{
			name: "name",
			with: func(ps *instrument.Preset) (*instrument.Preset, error) { return ps.WithName("Second") },
			must: func(ps *instrument.Preset) *instrument.Preset { return ps.MustWithName("Second") },
			want: preset("Second", h0(), rangeKP6()),
		},
		{
			name:    "empty name",
			with:    func(ps *instrument.Preset) (*instrument.Preset, error) { return ps.WithName("") },
			must:    func(ps *instrument.Preset) *instrument.Preset { return ps.MustWithName("") },
			wantErr: instrument.ErrEmptyName,
		},
		{
			name: "set",
			with: func(ps *instrument.Preset) (*instrument.Preset, error) { return ps.WithSettingsSet(h415()) },
			must: func(ps *instrument.Preset) *instrument.Preset { return ps.MustWithSettingsSet(h415()) },
			want: preset("Lead", h415(), rangeKP6()),
		},
		{
			name:    "К-7 set Н-19 inverts КП-6",
			with:    func(ps *instrument.Preset) (*instrument.Preset, error) { return ps.WithSettingsSet(h19()) },
			must:    func(ps *instrument.Preset) *instrument.Preset { return ps.MustWithSettingsSet(h19()) },
			wantErr: instrument.ErrInverted,
		},
		{
			name:    "nil set",
			with:    func(ps *instrument.Preset) (*instrument.Preset, error) { return ps.WithSettingsSet(nil) },
			must:    func(ps *instrument.Preset) *instrument.Preset { return ps.MustWithSettingsSet(nil) },
			wantErr: instrument.ErrUnset,
		},
		{
			name: "range",
			with: func(ps *instrument.Preset) (*instrument.Preset, error) { return ps.WithRange(rng("C4", "G5")) },
			must: func(ps *instrument.Preset) *instrument.Preset { return ps.MustWithRange(rng("C4", "G5")) },
			want: preset("Lead", h0(), rng("C4", "G5")),
		},
		{
			name:    "range КП-7 in Н-0",
			with:    func(ps *instrument.Preset) (*instrument.Preset, error) { return ps.WithRange(rangeKP7()) },
			must:    func(ps *instrument.Preset) *instrument.Preset { return ps.MustWithRange(rangeKP7()) },
			wantErr: instrument.ErrInadmissible,
		},
		{
			name:    "inverted range",
			with:    func(ps *instrument.Preset) (*instrument.Preset, error) { return ps.WithRange(rng("G5", "C4")) },
			must:    func(ps *instrument.Preset) *instrument.Preset { return ps.MustWithRange(rng("G5", "C4")) },
			wantErr: instrument.ErrInverted,
		},
		{
			name: "К-7 the unset range is no range",
			with: func(ps *instrument.Preset) (*instrument.Preset, error) { return ps.WithRange(instrument.Range{}) },
			must: func(ps *instrument.Preset) *instrument.Preset { return ps.MustWithRange(instrument.Range{}) },
			want: preset("Lead", h0(), instrument.Range{}),
		},
		{
			name: "without range",
			with: func(ps *instrument.Preset) (*instrument.Preset, error) { return ps.WithoutRange() },
			must: func(ps *instrument.Preset) *instrument.Preset { return ps.MustWithoutRange() },
			want: preset("Lead", h0(), instrument.Range{}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			src := base()
			got, err := tt.with(src)
			assert.True(t, base().Equal(src), "the source does not change")
			if tt.wantErr != nil {
				assertCause(t, err, tt.wantErr)
				assert.Nil(t, got)
				assertPanicsWith(t, tt.wantErr, err, func() { tt.must(src) })
				return
			}
			require.NoError(t, err)
			assert.NotSame(t, src, got, "a new preset")
			assert.True(t, tt.want.Equal(got), "want %v, got %v", tt.want, got)
			assert.True(t, tt.want.Equal(tt.must(src)))
			assert.True(t, base().Equal(src), "the source does not change")
		})
	}

	t.Run("without range after without range", func(t *testing.T) {
		t.Parallel()
		got, err := preset("Lead", h0(), instrument.Range{}).WithoutRange()
		require.NoError(t, err)
		assert.False(t, got.HasRange())
	})

	t.Run("set of the new system checks the range", func(t *testing.T) {
		t.Parallel()
		got, err := preset("Lead", h24(), rangeKP7()).WithSettingsSet(h0())
		assertCause(t, err, instrument.ErrInadmissible)
		assert.Nil(t, got)
		got, err = preset("Lead", h0(), rng("C4", "G5")).WithSettingsSet(h19())
		require.NoError(t, err)
		assert.True(t, got.SettingsSet().Equal(h19()))
	})
}

// ОП-21, П-20, К-8.
func TestPreset_Equal(t *testing.T) {
	t.Parallel()
	base := preset("Lead", h0(), rng("C4", "G5"))
	tests := []struct {
		name string
		a, b *instrument.Preset
		want bool
	}{
		{name: "equal parts", a: base, b: preset("Lead", h0(), rng("C4", "G5")), want: true},
		{name: "itself", a: base, b: base, want: true},
		{name: "other name", a: base, b: preset("Lead 2", h0(), rng("C4", "G5")), want: false},
		{name: "name differs only in case", a: base, b: preset("lead", h0(), rng("C4", "G5")), want: false},
		{name: "other set", a: base, b: preset("Lead", h415(), rng("C4", "G5")), want: false},
		{name: "other range", a: base, b: preset("Lead", h0(), rng("C4", "A5")), want: false},
		{name: "range by other record", a: preset("Lead", h0(), rng("B♯3", "G5")), b: base, want: false},
		{name: "no range and a range", a: preset("Lead", h0(), instrument.Range{}), b: base, want: false},
		{name: "both without range", a: preset("Lead", h0(), instrument.Range{}), b: preset("Lead", h0(), instrument.Range{}), want: true},
		{name: "two unset", a: nil, b: &instrument.Preset{}, want: true},
		{name: "two nil", a: nil, b: nil, want: true},
		{name: "unset and set", a: nil, b: base, want: false},
		{name: "set and unset", a: base, b: &instrument.Preset{}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.a.Equal(tt.b))
			assert.Equal(t, tt.want, tt.b.Equal(tt.a), "symmetric")
		})
	}
}

// ОП-22, К-26, А 4.10.
func TestPreset_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		ps   *instrument.Preset
	}{
		{name: "with a range", ps: preset("Lead voice", h0(), rng("C4", "G5"))},
		{name: "without a range", ps: preset("Lead voice", h24(), instrument.Range{})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := tt.ps.String()
			assert.Contains(t, s, tt.ps.Name())
			assert.Contains(t, s, tt.ps.SettingsSet().String())
			if tt.ps.HasRange() {
				assert.Contains(t, s, tt.ps.Range().String())
			} else {
				assert.Contains(t, s, "no range")
			}
			assert.NotEqual(t, "Preset(unset)", s)
		})
	}

	t.Run("strings do not affect equality", func(t *testing.T) {
		t.Parallel()
		a := preset("Lead", h0(), rng("C4", "G5"))
		b := preset("Lead", h0(), rng("B♯3", "G5"))
		assert.False(t, a.Equal(b))
		assert.True(t, a.Equal(preset("Lead", settings.Standard(), rng("C4", "G5"))))
	})
}
