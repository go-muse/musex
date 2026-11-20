package voice_test

import (
	"testing"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/voice"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Architecture 6.3: the three forms of the duration of an element.
func TestDurationForms(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		d          voice.Duration
		kind       voice.DurationKind
		length     string
		written    duration.Written
		hasWritten bool
		tied       duration.Tied
		hasTied    bool
	}{
		{name: "written quarter", d: q(), kind: voice.DurationWritten, length: "1/4", written: wr(duration.Quarter, 0), hasWritten: true},
		{name: "written dotted quarter", d: dq(), kind: voice.DurationWritten, length: "3/8", written: wr(duration.Quarter, 1), hasWritten: true},
		{name: "value", d: val(3, 7), kind: voice.DurationValue, length: "3/7"},
		{name: "tied quarter and eighth", d: qe(), kind: voice.DurationTied, length: "3/8", tied: tiedQE(), hasTied: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.True(t, tt.d.IsSet())
			assert.Equal(t, tt.kind, tt.d.Kind())
			assert.NotEmpty(t, tt.d.String())
			assert.NotEqual(t, "Duration(unset)", tt.d.String())

			length, err := tt.d.Length()
			require.NoError(t, err)
			assert.Equal(t, tt.length, length.String())
			assert.Equal(t, tt.length, tt.d.MustLength().String())

			w, ok, err := tt.d.Written()
			require.NoError(t, err)
			assert.Equal(t, tt.hasWritten, ok)
			assert.True(t, tt.written.Equal(w) || (!tt.hasWritten && !w.IsSet()), "written: %v", w)
			mw, mok := tt.d.MustWritten()
			assert.Equal(t, ok, mok)
			assert.True(t, w.Equal(mw) || (!w.IsSet() && !mw.IsSet()))

			tied, ok, err := tt.d.Tied()
			require.NoError(t, err)
			assert.Equal(t, tt.hasTied, ok)
			assert.True(t, tt.tied.Equal(tied) || (!tt.hasTied && !tied.IsSet()), "tied: %v", tied)
			mt, mok := tt.d.MustTied()
			assert.Equal(t, ok, mok)
			assert.True(t, tied.Equal(mt) || (!tied.IsSet() && !mt.IsSet()))
		})
	}
}

// П-3, РМ-2: the record is kept as given; one notation does not replace
// another.
func TestDurationKeepsRecord(t *testing.T) {
	t.Parallel()
	d := voice.MustWrittenDuration(wr(duration.Quarter, 1))
	w, ok := d.MustWritten()
	require.True(t, ok)
	assert.Equal(t, 1, w.MustDots())
	assert.Equal(t, duration.Quarter, w.MustBase())

	tied := voice.MustTiedDuration(tiedQE())
	got, ok := tied.MustTied()
	require.True(t, ok)
	assert.Len(t, got.MustParts(), 2)
	assert.True(t, tiedQE().Equal(got))
}

func TestDurationConstructorsRejectUnset(t *testing.T) {
	t.Parallel()
	w := fails(voice.ErrUnset, duration.ErrUnset)
	assertFailure(t, w,
		func() (voice.Duration, error) { return voice.WrittenDuration(duration.Written{}) },
		func() voice.Duration { return voice.MustWrittenDuration(duration.Written{}) })
	assertFailure(t, w,
		func() (voice.Duration, error) { return voice.ValueDuration(duration.Duration{}) },
		func() voice.Duration { return voice.MustValueDuration(duration.Duration{}) })
	assertFailure(t, w,
		func() (voice.Duration, error) { return voice.TiedDuration(duration.Tied{}) },
		func() voice.Duration { return voice.MustTiedDuration(duration.Tied{}) })
}

// Architecture 6.1: the zero Duration is unset.
func TestUnsetDuration(t *testing.T) {
	t.Parallel()
	var d voice.Duration
	assert.False(t, d.IsSet())
	assert.Equal(t, voice.DurationUnset, d.Kind())
	assert.Equal(t, "Duration(unset)", d.String())
	assert.True(t, d.Equal(voice.Duration{}))
	assert.False(t, d.Equal(q()))
	assert.False(t, q().Equal(d))

	w := fails(voice.ErrUnset).only()
	assertFailure(t, w, func() (pair[duration.Written, bool], error) { return two(d.Written()) },
		func() pair[duration.Written, bool] { return mustTwo(d.MustWritten()) })
	assertFailure(t, w, func() (pair[duration.Tied, bool], error) { return two(d.Tied()) },
		func() pair[duration.Tied, bool] { return mustTwo(d.MustTied()) })
	assertFailure(t, w, d.Length, d.MustLength)
}

// Architecture 6.3: Equal compares the form and the record.
func TestDurationEqual(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		a, b voice.Duration
		want bool
	}{
		{"same written", q(), q(), true},
		{"other written", q(), e8(), false},
		{"written and value of the same length", q(), val(1, 4), false},
		{"dotted and tied of the same length", dq(), qe(), false},
		{"same value", val(1, 3), val(2, 6), true},
		{"other value", val(1, 3), val(1, 4), false},
		{"same tied", qe(), qe(), true},
		{"other tied", qe(), voice.MustTiedDuration(duration.MustNewTied(wr(duration.Eighth, 0), wr(duration.Quarter, 0))), false},
		{"set and unset", q(), voice.Duration{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.a.Equal(tt.b))
			assert.Equal(t, tt.want, tt.b.Equal(tt.a))
		})
	}
}

// Architecture 6.14: the strings of DurationKind.
func TestDurationKind(t *testing.T) {
	t.Parallel()
	tests := []struct {
		kind  voice.DurationKind
		valid bool
		text  string
	}{
		{voice.DurationUnset, false, "DurationKind(unset)"},
		{voice.DurationWritten, true, "written"},
		{voice.DurationValue, true, "value"},
		{voice.DurationTied, true, "tied"},
		{voice.DurationKind(42), false, "DurationKind(invalid)"},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.valid, tt.kind.IsValid())
			assert.Equal(t, tt.text, tt.kind.String())
		})
	}
}
