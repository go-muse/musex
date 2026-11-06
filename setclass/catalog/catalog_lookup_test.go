package catalog_test

import (
	"math"
	"slices"
	"testing"

	"github.com/go-muse/musex/setclass/catalog"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlgorithm_IsValid_And_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		alg       catalog.Algorithm
		wantValid bool
		wantStr   string
	}{
		{
			name:      "Rahn",
			alg:       catalog.Rahn,
			wantValid: true,
			wantStr:   "Rahn",
		},
		{
			name:      "Forte",
			alg:       catalog.Forte,
			wantValid: true,
			wantStr:   "Forte",
		},
		{
			name:      "invalid_2",
			alg:       catalog.Algorithm(2),
			wantValid: false,
			wantStr:   "Algorithm(invalid)",
		},
		{
			name:      "invalid_99",
			alg:       catalog.Algorithm(99),
			wantValid: false,
			wantStr:   "Algorithm(invalid)",
		},
		{
			name:      "invalid_255",
			alg:       catalog.Algorithm(255),
			wantValid: false,
			wantStr:   "Algorithm(invalid)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.wantValid, tt.alg.IsValid())
			assert.Equal(t, tt.wantStr, tt.alg.String())
		})
	}
}

func TestAlgorithm_Equal_And_Compare(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		a         catalog.Algorithm
		b         catalog.Algorithm
		wantEqual bool
		wantCmp   int
	}{
		{
			name:      "Rahn_equals_Rahn",
			a:         catalog.Rahn,
			b:         catalog.Rahn,
			wantEqual: true,
			wantCmp:   0,
		},
		{
			name:      "Forte_equals_Forte",
			a:         catalog.Forte,
			b:         catalog.Forte,
			wantEqual: true,
			wantCmp:   0,
		},
		{
			name:      "Rahn_less_than_Forte",
			a:         catalog.Rahn,
			b:         catalog.Forte,
			wantEqual: false,
			wantCmp:   -1,
		},
		{
			name:      "Forte_greater_than_Rahn",
			a:         catalog.Forte,
			b:         catalog.Rahn,
			wantEqual: false,
			wantCmp:   1,
		},
		{
			name:      "invalid_2_and_invalid_2",
			a:         catalog.Algorithm(2),
			b:         catalog.Algorithm(2),
			wantEqual: false,
			wantCmp:   0,
		},
		{
			name:      "invalid_2_and_invalid_255",
			a:         catalog.Algorithm(2),
			b:         catalog.Algorithm(255),
			wantEqual: false,
			wantCmp:   0,
		},
		{
			name:      "invalid_less_than_Rahn",
			a:         catalog.Algorithm(2),
			b:         catalog.Rahn,
			wantEqual: false,
			wantCmp:   -1,
		},
		{
			name:      "Rahn_greater_than_invalid",
			a:         catalog.Rahn,
			b:         catalog.Algorithm(99),
			wantEqual: false,
			wantCmp:   1,
		},
		{
			name:      "Forte_greater_than_invalid",
			a:         catalog.Forte,
			b:         catalog.Algorithm(255),
			wantEqual: false,
			wantCmp:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.wantEqual, tt.a.Equal(tt.b))
			assert.Equal(t, tt.wantCmp, tt.a.Compare(tt.b))
			if tt.a.IsValid() && tt.b.IsValid() {
				assert.Equal(t, tt.a == tt.b, tt.a.Equal(tt.b))
				assert.Equal(t, tt.a.Compare(tt.b) == 0, tt.a.Equal(tt.b))
			}
		})
	}

	m := map[catalog.Algorithm]string{
		catalog.Rahn:  "rahn",
		catalog.Forte: "forte",
	}
	assert.Len(t, m, 2)
	assert.Equal(t, "rahn", m[catalog.Rahn])
	assert.Equal(t, "forte", m[catalog.Forte])
}

func TestSuffix_IsValid_And_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		suffix    catalog.Suffix
		wantValid bool
		wantStr   string
	}{
		{
			name:      "SuffixNone",
			suffix:    catalog.SuffixNone,
			wantValid: true,
			wantStr:   "",
		},
		{
			name:      "SuffixA",
			suffix:    catalog.SuffixA,
			wantValid: true,
			wantStr:   "A",
		},
		{
			name:      "SuffixB",
			suffix:    catalog.SuffixB,
			wantValid: true,
			wantStr:   "B",
		},
		{
			name:      "invalid_3",
			suffix:    catalog.Suffix(3),
			wantValid: false,
			wantStr:   "Suffix(invalid)",
		},
		{
			name:      "invalid_99",
			suffix:    catalog.Suffix(99),
			wantValid: false,
			wantStr:   "Suffix(invalid)",
		},
		{
			name:      "invalid_255",
			suffix:    catalog.Suffix(255),
			wantValid: false,
			wantStr:   "Suffix(invalid)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.wantValid, tt.suffix.IsValid())
			assert.Equal(t, tt.wantStr, tt.suffix.String())
		})
	}
}

func TestSuffix_Equal_And_Compare(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		a         catalog.Suffix
		b         catalog.Suffix
		wantEqual bool
		wantCmp   int
	}{
		{
			name:      "SuffixNone_equals_SuffixNone",
			a:         catalog.SuffixNone,
			b:         catalog.SuffixNone,
			wantEqual: true,
			wantCmp:   0,
		},
		{
			name:      "SuffixA_equals_SuffixA",
			a:         catalog.SuffixA,
			b:         catalog.SuffixA,
			wantEqual: true,
			wantCmp:   0,
		},
		{
			name:      "SuffixB_equals_SuffixB",
			a:         catalog.SuffixB,
			b:         catalog.SuffixB,
			wantEqual: true,
			wantCmp:   0,
		},
		{
			name:      "SuffixNone_less_than_SuffixA",
			a:         catalog.SuffixNone,
			b:         catalog.SuffixA,
			wantEqual: false,
			wantCmp:   -1,
		},
		{
			name:      "SuffixA_less_than_SuffixB",
			a:         catalog.SuffixA,
			b:         catalog.SuffixB,
			wantEqual: false,
			wantCmp:   -1,
		},
		{
			name:      "SuffixB_greater_than_SuffixNone",
			a:         catalog.SuffixB,
			b:         catalog.SuffixNone,
			wantEqual: false,
			wantCmp:   1,
		},
		{
			name:      "SuffixB_greater_than_SuffixA",
			a:         catalog.SuffixB,
			b:         catalog.SuffixA,
			wantEqual: false,
			wantCmp:   1,
		},
		{
			name:      "invalid_3_and_invalid_3",
			a:         catalog.Suffix(3),
			b:         catalog.Suffix(3),
			wantEqual: false,
			wantCmp:   0,
		},
		{
			name:      "invalid_3_and_invalid_255",
			a:         catalog.Suffix(3),
			b:         catalog.Suffix(255),
			wantEqual: false,
			wantCmp:   0,
		},
		{
			name:      "invalid_less_than_SuffixNone",
			a:         catalog.Suffix(3),
			b:         catalog.SuffixNone,
			wantEqual: false,
			wantCmp:   -1,
		},
		{
			name:      "SuffixNone_greater_than_invalid",
			a:         catalog.SuffixNone,
			b:         catalog.Suffix(99),
			wantEqual: false,
			wantCmp:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.wantEqual, tt.a.Equal(tt.b))
			assert.Equal(t, tt.wantCmp, tt.a.Compare(tt.b))
			if tt.a.IsValid() && tt.b.IsValid() {
				assert.Equal(t, tt.a == tt.b, tt.a.Equal(tt.b))
				assert.Equal(t, tt.a.Compare(tt.b) == 0, tt.a.Equal(tt.b))
			}
		})
	}

	m := map[catalog.Suffix]int{
		catalog.SuffixNone: 0,
		catalog.SuffixA:    1,
		catalog.SuffixB:    2,
	}
	assert.Len(t, m, 3)
}

func TestSentinelErrors_Distinct(t *testing.T) {
	t.Parallel()

	sentinels := []struct {
		name string
		err  error
	}{
		{name: "ErrInvalidNotation", err: catalog.ErrInvalidNotation},
		{name: "ErrInvalidCardinality", err: catalog.ErrInvalidCardinality},
		{name: "ErrInvalidOrdinal", err: catalog.ErrInvalidOrdinal},
		{name: "ErrInvalidSuffix", err: catalog.ErrInvalidSuffix},
		{name: "ErrNoCarterCatalog", err: catalog.ErrNoCarterCatalog},
		{name: "ErrInvalidCarterNumber", err: catalog.ErrInvalidCarterNumber},
		{name: "ErrNotPrimeForm", err: catalog.ErrNotPrimeForm},
		{name: "ErrInvalidValue", err: catalog.ErrInvalidValue},
		{name: "ErrInvalidAlgorithm", err: catalog.ErrInvalidAlgorithm},
	}

	for i, s1 := range sentinels {
		require.NotNil(t, s1.err, "sentinel %s must not be nil", s1.name)
		assert.NotEmpty(t, s1.err.Error())
		for j, s2 := range sentinels {
			if i == j {
				assert.ErrorIs(t, s1.err, s2.err)
			} else {
				assert.NotErrorIs(t, s1.err, s2.err, "%s must not match %s", s1.name, s2.name)
			}
		}
	}
}

func TestParseForteClass_ValidRoundTrip(t *testing.T) {
	t.Parallel()

	all := catalog.ForteClasses()
	require.Len(t, all, 224)

	for _, fc := range all {
		label := fc.String()
		t.Run(label, func(t *testing.T) {
			t.Parallel()

			parsed, err := catalog.ParseForteClass(label)
			require.NoError(t, err)
			assert.True(t, parsed.IsValid())
			assert.Equal(t, fc, parsed)
			assert.True(t, parsed.Equal(fc))
			assert.Equal(t, label, parsed.String())

			mustParsed := catalog.MustParseForteClass(label)
			assert.Equal(t, fc, mustParsed)
		})
	}
}

func TestParseForteClass_Invalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		extraErrs []error
	}{
		{name: "empty", input: ""},
		{name: "whitespace_only", input: "   "},
		{name: "no_hyphen", input: "311"},
		{name: "two_hyphens", input: "3-11-1"},
		{name: "only_hyphen", input: "-"},
		{name: "missing_ordinal", input: "3-"},
		{name: "missing_cardinality", input: "-1"},
		{name: "invalid_diagnostic_forte", input: "ForteClass(invalid)"},
		{name: "invalid_diagnostic_tntype", input: "TnType(invalid)"},
		{name: "leading_space", input: " 3-11"},
		{name: "trailing_space", input: "3-11 "},
		{name: "spaces_around_hyphen", input: "3 - 11"},
		{name: "tab_prefix", input: "\t3-11"},
		{name: "newline_suffix", input: "3-11\n"},
		{name: "en_dash", input: "3–11"},
		{name: "em_dash", input: "3—11"},
		{name: "lowercase_z", input: "4-z15"},
		{name: "lowercase_z_hexachord", input: "6-z29"},
		{name: "plus_cardinality", input: "+3-1"},
		{name: "plus_ordinal", input: "3-+1"},
		{name: "negative_cardinality", input: "-1-1"},
		{name: "negative_ordinal", input: "3--1"},
		{name: "leading_zero_zero", input: "00-1"},
		{name: "leading_zero_cardinality", input: "03-11"},
		{name: "leading_zero_cardinality_4", input: "04-1"},
		{name: "leading_zero_ordinal", input: "3-01"},
		{name: "leading_zero_ordinal_4", input: "4-01"},
		{name: "leading_zero_after_Z", input: "4-Z015"},
		{name: "missing_Z_4_15", input: "4-15"},
		{name: "missing_Z_4_29", input: "4-29"},
		{name: "missing_Z_5_12", input: "5-12"},
		{name: "missing_Z_5_36", input: "5-36"},
		{name: "missing_Z_6_28", input: "6-28"},
		{name: "missing_Z_6_29", input: "6-29"},
		{name: "missing_Z_6_50", input: "6-50"},
		{name: "missing_Z_7_12", input: "7-12"},
		{name: "missing_Z_7_18", input: "7-18"},
		{name: "missing_Z_8_15", input: "8-15"},
		{name: "missing_Z_8_29", input: "8-29"},
		{name: "spurious_Z_0_1", input: "0-Z1"},
		{name: "spurious_Z_3_11", input: "3-Z11"},
		{name: "spurious_Z_4_14", input: "4-Z14"},
		{name: "spurious_Z_5_20", input: "5-Z20"},
		{name: "spurious_Z_6_1", input: "6-Z1"},
		{name: "spurious_Z_6_14", input: "6-Z14"},
		{name: "spurious_Z_6_35", input: "6-Z35"},
		{name: "spurious_Z_12_1", input: "12-Z1"},
		{name: "suffix_A_on_chiral", input: "3-11A"},
		{name: "suffix_B_on_chiral", input: "3-11B"},
		{name: "suffix_A_on_Z_chiral", input: "4-Z15A"},
		{name: "suffix_B_on_Z_chiral", input: "4-Z15B"},
		{name: "suffix_A_on_symmetric", input: "3-1A"},
		{name: "lowercase_suffix_a", input: "3-11a"},
		{name: "cardinality_13", input: "13-1", extraErrs: []error{catalog.ErrInvalidCardinality}},
		{name: "cardinality_99", input: "99-1", extraErrs: []error{catalog.ErrInvalidCardinality}},
		{name: "ordinal_0_card_0", input: "0-0", extraErrs: []error{catalog.ErrInvalidOrdinal}},
		{name: "ordinal_2_card_0", input: "0-2", extraErrs: []error{catalog.ErrInvalidOrdinal}},
		{name: "ordinal_2_card_1", input: "1-2", extraErrs: []error{catalog.ErrInvalidOrdinal}},
		{name: "ordinal_7_card_2", input: "2-7", extraErrs: []error{catalog.ErrInvalidOrdinal}},
		{name: "ordinal_0_card_3", input: "3-0", extraErrs: []error{catalog.ErrInvalidOrdinal}},
		{name: "ordinal_13_card_3", input: "3-13", extraErrs: []error{catalog.ErrInvalidOrdinal}},
		{name: "ordinal_0_card_4", input: "4-0", extraErrs: []error{catalog.ErrInvalidOrdinal}},
		{name: "ordinal_30_card_4", input: "4-30", extraErrs: []error{catalog.ErrInvalidOrdinal}},
		{name: "ordinal_39_card_5", input: "5-39", extraErrs: []error{catalog.ErrInvalidOrdinal}},
		{name: "ordinal_51_card_6", input: "6-51", extraErrs: []error{catalog.ErrInvalidOrdinal}},
		{name: "ordinal_39_card_7", input: "7-39", extraErrs: []error{catalog.ErrInvalidOrdinal}},
		{name: "ordinal_30_card_8", input: "8-30", extraErrs: []error{catalog.ErrInvalidOrdinal}},
		{name: "ordinal_13_card_9", input: "9-13", extraErrs: []error{catalog.ErrInvalidOrdinal}},
		{name: "ordinal_7_card_10", input: "10-7", extraErrs: []error{catalog.ErrInvalidOrdinal}},
		{name: "ordinal_2_card_11", input: "11-2", extraErrs: []error{catalog.ErrInvalidOrdinal}},
		{name: "ordinal_2_card_12", input: "12-2", extraErrs: []error{catalog.ErrInvalidOrdinal}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fc, err := catalog.ParseForteClass(tt.input)
			require.Error(t, err)
			assert.ErrorIs(t, err, catalog.ErrInvalidNotation)
			for _, extra := range tt.extraErrs {
				assert.ErrorIs(t, err, extra)
			}
			assert.False(t, fc.IsValid())
			assert.Equal(t, catalog.ForteClass{}, fc)

			assert.Panics(t, func() {
				_ = catalog.MustParseForteClass(tt.input)
			})
		})
	}
}

func TestParseTnType_ValidRoundTrip(t *testing.T) {
	t.Parallel()

	all := catalog.TnTypes()
	require.Len(t, all, 352)

	for _, tn := range all {
		label := tn.String()
		t.Run(label, func(t *testing.T) {
			t.Parallel()

			parsed, err := catalog.ParseTnType(label)
			require.NoError(t, err)
			assert.True(t, parsed.IsValid())
			assert.Equal(t, tn, parsed)
			assert.True(t, parsed.Equal(tn))
			assert.Equal(t, label, parsed.String())

			mustParsed := catalog.MustParseTnType(label)
			assert.Equal(t, tn, mustParsed)
		})
	}
}

func TestParseTnType_Invalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		extraErrs []error
	}{
		{name: "empty", input: ""},
		{name: "invalid_diagnostic", input: "TnType(invalid)"},
		{name: "leading_space", input: " 3-11A"},
		{name: "trailing_space", input: "3-11A "},
		{name: "space_before_suffix", input: "3-11 A"},
		{name: "lowercase_a", input: "3-11a"},
		{name: "lowercase_b", input: "3-11b"},
		{name: "lowercase_z", input: "4-z15A"},
		{name: "unknown_suffix_C", input: "3-11C"},
		{name: "leading_zero_card", input: "03-11A"},
		{name: "leading_zero_ord", input: "3-011A"},
		{name: "leading_zero_Z", input: "4-Z015A"},
		{name: "missing_Z_chiral_A", input: "4-15A"},
		{name: "missing_Z_chiral_B", input: "4-15B"},
		{name: "missing_Z_symmetric", input: "5-12"},
		{name: "missing_Z_hexachord", input: "6-29"},
		{name: "spurious_Z_chiral_A", input: "3-Z11A"},
		{name: "spurious_Z_symmetric", input: "3-Z1"},
		{name: "symmetric_with_A_0_1", input: "0-1A", extraErrs: []error{catalog.ErrInvalidSuffix}},
		{name: "symmetric_with_B_1_1", input: "1-1B", extraErrs: []error{catalog.ErrInvalidSuffix}},
		{name: "symmetric_with_A_2_3", input: "2-3A", extraErrs: []error{catalog.ErrInvalidSuffix}},
		{name: "symmetric_with_A_3_1", input: "3-1A", extraErrs: []error{catalog.ErrInvalidSuffix}},
		{name: "symmetric_with_B_3_6", input: "3-6B", extraErrs: []error{catalog.ErrInvalidSuffix}},
		{name: "symmetric_with_B_3_12", input: "3-12B", extraErrs: []error{catalog.ErrInvalidSuffix}},
		{name: "symmetric_with_A_4_1", input: "4-1A", extraErrs: []error{catalog.ErrInvalidSuffix}},
		{name: "symmetric_with_A_5_Z12", input: "5-Z12A", extraErrs: []error{catalog.ErrInvalidSuffix}},
		{name: "symmetric_with_B_5_Z17", input: "5-Z17B", extraErrs: []error{catalog.ErrInvalidSuffix}},
		{name: "symmetric_with_A_6_1", input: "6-1A", extraErrs: []error{catalog.ErrInvalidSuffix}},
		{name: "symmetric_with_A_6_Z4", input: "6-Z4A", extraErrs: []error{catalog.ErrInvalidSuffix}},
		{name: "symmetric_with_A_6_Z29", input: "6-Z29A", extraErrs: []error{catalog.ErrInvalidSuffix}},
		{name: "symmetric_with_B_6_Z29", input: "6-Z29B", extraErrs: []error{catalog.ErrInvalidSuffix}},
		{name: "symmetric_with_A_6_35", input: "6-35A", extraErrs: []error{catalog.ErrInvalidSuffix}},
		{name: "symmetric_with_A_7_Z12", input: "7-Z12A", extraErrs: []error{catalog.ErrInvalidSuffix}},
		{name: "symmetric_with_A_8_26", input: "8-26A", extraErrs: []error{catalog.ErrInvalidSuffix}},
		{name: "symmetric_with_B_12_1", input: "12-1B", extraErrs: []error{catalog.ErrInvalidSuffix}},
		{name: "chiral_without_suffix_3_2", input: "3-2", extraErrs: []error{catalog.ErrInvalidSuffix}},
		{name: "chiral_without_suffix_3_11", input: "3-11", extraErrs: []error{catalog.ErrInvalidSuffix}},
		{name: "chiral_without_suffix_4_2", input: "4-2", extraErrs: []error{catalog.ErrInvalidSuffix}},
		{name: "chiral_without_suffix_4_Z15", input: "4-Z15", extraErrs: []error{catalog.ErrInvalidSuffix}},
		{name: "chiral_without_suffix_4_Z29", input: "4-Z29", extraErrs: []error{catalog.ErrInvalidSuffix}},
		{name: "chiral_without_suffix_5_20", input: "5-20", extraErrs: []error{catalog.ErrInvalidSuffix}},
		{name: "chiral_without_suffix_6_14", input: "6-14", extraErrs: []error{catalog.ErrInvalidSuffix}},
		{name: "chiral_without_suffix_6_Z10", input: "6-Z10", extraErrs: []error{catalog.ErrInvalidSuffix}},
		{name: "chiral_without_suffix_7_Z18", input: "7-Z18", extraErrs: []error{catalog.ErrInvalidSuffix}},
		{name: "chiral_without_suffix_8_Z15", input: "8-Z15", extraErrs: []error{catalog.ErrInvalidSuffix}},
		{name: "cardinality_13", input: "13-1", extraErrs: []error{catalog.ErrInvalidCardinality}},
		{name: "cardinality_13_with_A", input: "13-1A", extraErrs: []error{catalog.ErrInvalidCardinality}},
		{name: "ordinal_0", input: "3-0A", extraErrs: []error{catalog.ErrInvalidOrdinal}},
		{name: "ordinal_30_card_4", input: "4-30A", extraErrs: []error{catalog.ErrInvalidOrdinal}},
		{name: "ordinal_51_card_6", input: "6-51", extraErrs: []error{catalog.ErrInvalidOrdinal}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tn, err := catalog.ParseTnType(tt.input)
			require.Error(t, err)
			assert.ErrorIs(t, err, catalog.ErrInvalidNotation)
			for _, extra := range tt.extraErrs {
				assert.ErrorIs(t, err, extra)
			}
			assert.False(t, tn.IsValid())
			assert.Equal(t, catalog.TnType{}, tn)

			assert.Panics(t, func() {
				_ = catalog.MustParseTnType(tt.input)
			})
		})
	}
}

func TestResolveForteClass_AllValidAndInvalid(t *testing.T) {
	t.Parallel()

	t.Run("all_224_ForteClass_labels", func(t *testing.T) {
		t.Parallel()

		for _, fc := range catalog.ForteClasses() {
			got, err := catalog.ResolveForteClass(fc.String())
			require.NoError(t, err)
			assert.Equal(t, fc, got)
			assert.Equal(t, fc, catalog.MustResolveForteClass(fc.String()))
		}
	})

	t.Run("all_352_TnType_labels", func(t *testing.T) {
		t.Parallel()

		for _, tn := range catalog.TnTypes() {
			got, err := catalog.ResolveForteClass(tn.String())
			require.NoError(t, err)
			assert.Equal(t, tn.ForteClass(), got)
			assert.Equal(t, tn.ForteClass(), catalog.MustResolveForteClass(tn.String()))
		}
	})

	t.Run("specific_resolutions", func(t *testing.T) {
		t.Parallel()

		c311 := catalog.MustParseForteClass("3-11")
		assert.Equal(t, c311, catalog.MustResolveForteClass("3-11"))
		assert.Equal(t, c311, catalog.MustResolveForteClass("3-11A"))
		assert.Equal(t, c311, catalog.MustResolveForteClass("3-11B"))

		c4Z15 := catalog.MustParseForteClass("4-Z15")
		assert.Equal(t, c4Z15, catalog.MustResolveForteClass("4-Z15"))
		assert.Equal(t, c4Z15, catalog.MustResolveForteClass("4-Z15A"))
		assert.Equal(t, c4Z15, catalog.MustResolveForteClass("4-Z15B"))

		c6Z29 := catalog.MustParseForteClass("6-Z29")
		assert.Equal(t, c6Z29, catalog.MustResolveForteClass("6-Z29"))
	})

	t.Run("invalid_labels", func(t *testing.T) {
		t.Parallel()

		invalidCases := []struct {
			input     string
			extraErrs []error
		}{
			{input: ""},
			{input: "4-15"},
			{input: "4-15A"},
			{input: "6-29"},
			{input: "3-Z11"},
			{input: "3-Z11A"},
			{input: "3-11a"},
			{input: "4-z15"},
			{input: "0-1A", extraErrs: []error{catalog.ErrInvalidSuffix}},
			{input: "3-1A", extraErrs: []error{catalog.ErrInvalidSuffix}},
			{input: "5-Z12A", extraErrs: []error{catalog.ErrInvalidSuffix}},
			{input: "6-Z29A", extraErrs: []error{catalog.ErrInvalidSuffix}},
			{input: "6-35B", extraErrs: []error{catalog.ErrInvalidSuffix}},
			{input: "13-1", extraErrs: []error{catalog.ErrInvalidCardinality}},
			{input: "4-30", extraErrs: []error{catalog.ErrInvalidOrdinal}},
		}

		for _, tc := range invalidCases {
			fc, err := catalog.ResolveForteClass(tc.input)
			require.Error(t, err, "input %q must fail", tc.input)
			assert.ErrorIs(t, err, catalog.ErrInvalidNotation)
			for _, extra := range tc.extraErrs {
				assert.ErrorIs(t, err, extra)
			}
			assert.False(t, fc.IsValid())
			assert.Panics(t, func() {
				_ = catalog.MustResolveForteClass(tc.input)
			})
		}
	})
}

func TestForteClassByNumber(t *testing.T) {
	t.Parallel()

	t.Run("all_224_valid_coordinates", func(t *testing.T) {
		t.Parallel()

		for _, fc := range catalog.ForteClasses() {
			got, err := catalog.ForteClassByNumber(fc.Cardinality(), fc.Ordinal())
			require.NoError(t, err)
			assert.Equal(t, fc, got)
			assert.Equal(t, fc, catalog.MustForteClassByNumber(fc.Cardinality(), fc.Ordinal()))
		}

		fc4Z15 := catalog.MustForteClassByNumber(4, 15)
		assert.Equal(t, "4-Z15", fc4Z15.String())
		assert.True(t, fc4Z15.IsZ())

		fc5Z12 := catalog.MustForteClassByNumber(5, 12)
		assert.Equal(t, "5-Z12", fc5Z12.String())
		assert.True(t, fc5Z12.IsZ())

		fc6Z29 := catalog.MustForteClassByNumber(6, 29)
		assert.Equal(t, "6-Z29", fc6Z29.String())
		assert.True(t, fc6Z29.IsZ())
	})

	t.Run("invalid_cardinality_and_ordinal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name    string
			c       int
			d       int
			wantErr error
		}{
			{name: "c_minus_1", c: -1, d: 1, wantErr: catalog.ErrInvalidCardinality},
			{name: "c_13", c: 13, d: 1, wantErr: catalog.ErrInvalidCardinality},
			{name: "c_min_int", c: math.MinInt, d: 1, wantErr: catalog.ErrInvalidCardinality},
			{name: "c_max_int", c: math.MaxInt, d: 1, wantErr: catalog.ErrInvalidCardinality},
			{name: "priority_c_over_d", c: -1, d: 0, wantErr: catalog.ErrInvalidCardinality},
			{name: "priority_c_13_over_d_99", c: 13, d: 99, wantErr: catalog.ErrInvalidCardinality},
			{name: "c0_d0", c: 0, d: 0, wantErr: catalog.ErrInvalidOrdinal},
			{name: "c0_d2", c: 0, d: 2, wantErr: catalog.ErrInvalidOrdinal},
			{name: "c1_d2", c: 1, d: 2, wantErr: catalog.ErrInvalidOrdinal},
			{name: "c2_d7", c: 2, d: 7, wantErr: catalog.ErrInvalidOrdinal},
			{name: "c3_d0", c: 3, d: 0, wantErr: catalog.ErrInvalidOrdinal},
			{name: "c3_d13", c: 3, d: 13, wantErr: catalog.ErrInvalidOrdinal},
			{name: "c4_d30", c: 4, d: 30, wantErr: catalog.ErrInvalidOrdinal},
			{name: "c5_d39", c: 5, d: 39, wantErr: catalog.ErrInvalidOrdinal},
			{name: "c6_d51", c: 6, d: 51, wantErr: catalog.ErrInvalidOrdinal},
			{name: "c7_d39", c: 7, d: 39, wantErr: catalog.ErrInvalidOrdinal},
			{name: "c8_d30", c: 8, d: 30, wantErr: catalog.ErrInvalidOrdinal},
			{name: "c9_d13", c: 9, d: 13, wantErr: catalog.ErrInvalidOrdinal},
			{name: "c10_d7", c: 10, d: 7, wantErr: catalog.ErrInvalidOrdinal},
			{name: "c11_d2", c: 11, d: 2, wantErr: catalog.ErrInvalidOrdinal},
			{name: "c12_d2", c: 12, d: 2, wantErr: catalog.ErrInvalidOrdinal},
			{name: "c6_d_neg", c: 6, d: -5, wantErr: catalog.ErrInvalidOrdinal},
			{name: "c6_d_max_int", c: 6, d: math.MaxInt, wantErr: catalog.ErrInvalidOrdinal},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				fc, err := catalog.ForteClassByNumber(tt.c, tt.d)
				require.ErrorIs(t, err, tt.wantErr)
				assert.False(t, fc.IsValid())
				assert.Panics(t, func() {
					_ = catalog.MustForteClassByNumber(tt.c, tt.d)
				})
			})
		}
	})
}

func TestTnTypeByNumber(t *testing.T) {
	t.Parallel()

	t.Run("all_352_valid_triples", func(t *testing.T) {
		t.Parallel()

		for _, tn := range catalog.TnTypes() {
			got, err := catalog.TnTypeByNumber(tn.Cardinality(), tn.Ordinal(), tn.Suffix())
			require.NoError(t, err)
			assert.Equal(t, tn, got)
			assert.Equal(t, tn, catalog.MustTnTypeByNumber(tn.Cardinality(), tn.Ordinal(), tn.Suffix()))
		}
	})

	t.Run("all_96_symmetric_reject_A_and_B", func(t *testing.T) {
		t.Parallel()

		symCount := 0
		for _, fc := range catalog.ForteClasses() {
			if fc.IsInversionallySymmetric() {
				symCount++
				for _, badSuffix := range []catalog.Suffix{catalog.SuffixA, catalog.SuffixB, catalog.Suffix(99)} {
					got, err := catalog.TnTypeByNumber(fc.Cardinality(), fc.Ordinal(), badSuffix)
					require.ErrorIs(t, err, catalog.ErrInvalidSuffix)
					assert.False(t, got.IsValid())
					assert.Panics(t, func() {
						_ = catalog.MustTnTypeByNumber(fc.Cardinality(), fc.Ordinal(), badSuffix)
					})
				}
			}
		}
		assert.Equal(t, 96, symCount)
	})

	t.Run("all_128_chiral_reject_SuffixNone", func(t *testing.T) {
		t.Parallel()

		chiralCount := 0
		for _, fc := range catalog.ForteClasses() {
			if fc.IsChiral() {
				chiralCount++
				for _, badSuffix := range []catalog.Suffix{catalog.SuffixNone, catalog.Suffix(3), catalog.Suffix(99)} {
					got, err := catalog.TnTypeByNumber(fc.Cardinality(), fc.Ordinal(), badSuffix)
					require.ErrorIs(t, err, catalog.ErrInvalidSuffix)
					assert.False(t, got.IsValid())
					assert.Panics(t, func() {
						_ = catalog.MustTnTypeByNumber(fc.Cardinality(), fc.Ordinal(), badSuffix)
					})
				}
			}
		}
		assert.Equal(t, 128, chiralCount)
	})

	t.Run("error_priority", func(t *testing.T) {
		t.Parallel()

		_, err := catalog.TnTypeByNumber(-1, 0, catalog.Suffix(99))
		require.ErrorIs(t, err, catalog.ErrInvalidCardinality)

		_, err = catalog.TnTypeByNumber(4, 0, catalog.Suffix(99))
		require.ErrorIs(t, err, catalog.ErrInvalidOrdinal)

		_, err = catalog.TnTypeByNumber(4, 30, catalog.Suffix(99))
		require.ErrorIs(t, err, catalog.ErrInvalidOrdinal)
	})
}

func TestForteClassByPrimeMask_And_PrimeForm(t *testing.T) {
	t.Parallel()

	rahnMasks := make(map[uint64]catalog.ForteClass, 224)
	forteMasks := make(map[uint64]catalog.ForteClass, 224)
	anyMasks := make(map[uint64]catalog.ForteClass, 230)

	for _, fc := range catalog.ForteClasses() {
		rMask := fc.PrimeMask()
		fMask := fc.FortePrimeMask()
		rForm := fc.PrimeForm()
		fForm := fc.FortePrimeForm()

		rahnMasks[rMask] = fc
		forteMasks[fMask] = fc
		anyMasks[rMask] = fc
		anyMasks[fMask] = fc

		gotRMask, err := catalog.ForteClassByPrimeMask(rMask)
		require.NoError(t, err)
		assert.Equal(t, fc, gotRMask)
		assert.Equal(t, fc, catalog.MustForteClassByPrimeMask(rMask))

		gotRMaskWith, err := catalog.ForteClassByPrimeMaskWith(rMask, catalog.Rahn)
		require.NoError(t, err)
		assert.Equal(t, fc, gotRMaskWith)
		assert.Equal(t, fc, catalog.MustForteClassByPrimeMaskWith(rMask, catalog.Rahn))

		gotFMaskWith, err := catalog.ForteClassByPrimeMaskWith(fMask, catalog.Forte)
		require.NoError(t, err)
		assert.Equal(t, fc, gotFMaskWith)
		assert.Equal(t, fc, catalog.MustForteClassByPrimeMaskWith(fMask, catalog.Forte))

		gotAnyR, err := catalog.ForteClassByAnyPrimeMask(rMask)
		require.NoError(t, err)
		assert.Equal(t, fc, gotAnyR)
		assert.Equal(t, fc, catalog.MustForteClassByAnyPrimeMask(rMask))

		gotAnyF, err := catalog.ForteClassByAnyPrimeMask(fMask)
		require.NoError(t, err)
		assert.Equal(t, fc, gotAnyF)
		assert.Equal(t, fc, catalog.MustForteClassByAnyPrimeMask(fMask))

		rFormCopy := slices.Clone(rForm)
		gotRForm, err := catalog.ForteClassByPrimeForm(rForm)
		require.NoError(t, err)
		assert.Equal(t, fc, gotRForm)
		assert.Equal(t, rFormCopy, rForm, "input slice must not be mutated")
		assert.Equal(t, fc, catalog.MustForteClassByPrimeForm(rForm))

		gotRFormWith, err := catalog.ForteClassByPrimeFormWith(rForm, catalog.Rahn)
		require.NoError(t, err)
		assert.Equal(t, fc, gotRFormWith)
		assert.Equal(t, fc, catalog.MustForteClassByPrimeFormWith(rForm, catalog.Rahn))

		gotFFormWith, err := catalog.ForteClassByPrimeFormWith(fForm, catalog.Forte)
		require.NoError(t, err)
		assert.Equal(t, fc, gotFFormWith)
		assert.Equal(t, fc, catalog.MustForteClassByPrimeFormWith(fForm, catalog.Forte))

		gotAnyFormR, err := catalog.ForteClassByAnyPrimeForm(rForm)
		require.NoError(t, err)
		assert.Equal(t, fc, gotAnyFormR)
		assert.Equal(t, fc, catalog.MustForteClassByAnyPrimeForm(rForm))

		gotAnyFormF, err := catalog.ForteClassByAnyPrimeForm(fForm)
		require.NoError(t, err)
		assert.Equal(t, fc, gotAnyFormF)
		assert.Equal(t, fc, catalog.MustForteClassByAnyPrimeForm(fForm))
	}

	assert.Len(t, rahnMasks, 224)
	assert.Len(t, forteMasks, 224)
	assert.Len(t, anyMasks, 230)

	t.Run("nil_and_empty_slice_resolve_to_0_1", func(t *testing.T) {
		t.Parallel()

		c01 := catalog.MustParseForteClass("0-1")
		for _, emptyInput := range [][]int{nil, {}} {
			got1, err := catalog.ForteClassByPrimeForm(emptyInput)
			require.NoError(t, err)
			assert.Equal(t, c01, got1)

			got2, err := catalog.ForteClassByPrimeFormWith(emptyInput, catalog.Rahn)
			require.NoError(t, err)
			assert.Equal(t, c01, got2)

			got3, err := catalog.ForteClassByPrimeFormWith(emptyInput, catalog.Forte)
			require.NoError(t, err)
			assert.Equal(t, c01, got3)

			got4, err := catalog.ForteClassByAnyPrimeForm(emptyInput)
			require.NoError(t, err)
			assert.Equal(t, c01, got4)
		}
	})

	t.Run("exhaustive_4096_masks", func(t *testing.T) {
		t.Parallel()

		rejectedAny := 0
		for mask := uint64(0); mask < 4096; mask++ {
			if _, ok := rahnMasks[mask]; !ok {
				_, err := catalog.ForteClassByPrimeMask(mask)
				require.ErrorIs(t, err, catalog.ErrNotPrimeForm)
				_, err = catalog.ForteClassByPrimeMaskWith(mask, catalog.Rahn)
				require.ErrorIs(t, err, catalog.ErrNotPrimeForm)
			}
			if _, ok := forteMasks[mask]; !ok {
				_, err := catalog.ForteClassByPrimeMaskWith(mask, catalog.Forte)
				require.ErrorIs(t, err, catalog.ErrNotPrimeForm)
			}
			if _, ok := anyMasks[mask]; !ok {
				rejectedAny++
				_, err := catalog.ForteClassByAnyPrimeMask(mask)
				require.ErrorIs(t, err, catalog.ErrNotPrimeForm)
			}
		}
		assert.Equal(t, 4096-230, rejectedAny)
	})

	t.Run("invalid_masks_slices_and_algorithms", func(t *testing.T) {
		t.Parallel()

		for _, badMask := range []uint64{145, 4096, 0xFFFF, math.MaxUint64} {
			_, err := catalog.ForteClassByPrimeMask(badMask)
			require.ErrorIs(t, err, catalog.ErrNotPrimeForm)
			assert.Panics(t, func() { _ = catalog.MustForteClassByPrimeMask(badMask) })

			_, err = catalog.ForteClassByPrimeMaskWith(badMask, catalog.Rahn)
			require.ErrorIs(t, err, catalog.ErrNotPrimeForm)
			assert.Panics(t, func() { _ = catalog.MustForteClassByPrimeMaskWith(badMask, catalog.Rahn) })

			_, err = catalog.ForteClassByPrimeMaskWith(badMask, catalog.Forte)
			require.ErrorIs(t, err, catalog.ErrNotPrimeForm)
			assert.Panics(t, func() { _ = catalog.MustForteClassByPrimeMaskWith(badMask, catalog.Forte) })

			_, err = catalog.ForteClassByAnyPrimeMask(badMask)
			require.ErrorIs(t, err, catalog.ErrNotPrimeForm)
			assert.Panics(t, func() { _ = catalog.MustForteClassByAnyPrimeMask(badMask) })
		}

		badSlices := [][]int{
			{-1},
			{12},
			{0, 12},
			{0, -1, 4},
			{0, 0, 4},
			{0, 7, 4},
			{2, 6, 9},
			{0, 4, 7},
			{0, 5, 9},
			{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		}
		for _, s := range badSlices {
			_, err := catalog.ForteClassByPrimeForm(s)
			require.ErrorIs(t, err, catalog.ErrNotPrimeForm)
			assert.Panics(t, func() { _ = catalog.MustForteClassByPrimeForm(s) })

			_, err = catalog.ForteClassByPrimeFormWith(s, catalog.Rahn)
			require.ErrorIs(t, err, catalog.ErrNotPrimeForm)
			assert.Panics(t, func() { _ = catalog.MustForteClassByPrimeFormWith(s, catalog.Rahn) })

			_, err = catalog.ForteClassByPrimeFormWith(s, catalog.Forte)
			require.ErrorIs(t, err, catalog.ErrNotPrimeForm)
			assert.Panics(t, func() { _ = catalog.MustForteClassByPrimeFormWith(s, catalog.Forte) })

			_, err = catalog.ForteClassByAnyPrimeForm(s)
			require.ErrorIs(t, err, catalog.ErrNotPrimeForm)
			assert.Panics(t, func() { _ = catalog.MustForteClassByAnyPrimeForm(s) })
		}

		// Cross-algorithm rejection for all 6 diverging ForteClasses
		differCount := 0
		for _, fc := range catalog.ForteClasses() {
			if !fc.RahnForteDiffer() {
				continue
			}
			differCount++
			rForm := fc.PrimeForm()
			fForm := fc.FortePrimeForm()
			rMask := fc.PrimeMask()
			fMask := fc.FortePrimeMask()

			_, err := catalog.ForteClassByPrimeForm(fForm)
			require.ErrorIs(t, err, catalog.ErrNotPrimeForm)
			assert.Panics(t, func() { _ = catalog.MustForteClassByPrimeForm(fForm) })

			_, err = catalog.ForteClassByPrimeFormWith(fForm, catalog.Rahn)
			require.ErrorIs(t, err, catalog.ErrNotPrimeForm)
			assert.Panics(t, func() { _ = catalog.MustForteClassByPrimeFormWith(fForm, catalog.Rahn) })

			_, err = catalog.ForteClassByPrimeFormWith(rForm, catalog.Forte)
			require.ErrorIs(t, err, catalog.ErrNotPrimeForm)
			assert.Panics(t, func() { _ = catalog.MustForteClassByPrimeFormWith(rForm, catalog.Forte) })

			_, err = catalog.ForteClassByPrimeMask(fMask)
			require.ErrorIs(t, err, catalog.ErrNotPrimeForm)
			assert.Panics(t, func() { _ = catalog.MustForteClassByPrimeMask(fMask) })

			_, err = catalog.ForteClassByPrimeMaskWith(fMask, catalog.Rahn)
			require.ErrorIs(t, err, catalog.ErrNotPrimeForm)
			assert.Panics(t, func() { _ = catalog.MustForteClassByPrimeMaskWith(fMask, catalog.Rahn) })

			_, err = catalog.ForteClassByPrimeMaskWith(rMask, catalog.Forte)
			require.ErrorIs(t, err, catalog.ErrNotPrimeForm)
			assert.Panics(t, func() { _ = catalog.MustForteClassByPrimeMaskWith(rMask, catalog.Forte) })
		}
		assert.Equal(t, 6, differCount)

		_, err := catalog.ForteClassByPrimeMaskWith(99999, catalog.Algorithm(99))
		require.ErrorIs(t, err, catalog.ErrInvalidAlgorithm)
		assert.Panics(t, func() { _ = catalog.MustForteClassByPrimeMaskWith(0, catalog.Algorithm(99)) })

		_, err = catalog.ForteClassByPrimeFormWith([]int{-1}, catalog.Algorithm(99))
		require.ErrorIs(t, err, catalog.ErrInvalidAlgorithm)
		assert.Panics(t, func() { _ = catalog.MustForteClassByPrimeFormWith([]int{0}, catalog.Algorithm(99)) })
	})
}

func TestTnTypeByPrimeMask_And_PrimeForm(t *testing.T) {
	t.Parallel()

	rahnMasks := make(map[uint64]catalog.TnType, 352)
	forteMasks := make(map[uint64]catalog.TnType, 352)
	anyMasks := make(map[uint64]catalog.TnType, 369)

	for _, tn := range catalog.TnTypes() {
		rMask := tn.PrimeMask()
		fMask := tn.FortePrimeMask()
		rForm := tn.PrimeForm()
		fForm := tn.FortePrimeForm()

		rahnMasks[rMask] = tn
		forteMasks[fMask] = tn
		anyMasks[rMask] = tn
		anyMasks[fMask] = tn

		gotRMask, err := catalog.TnTypeByPrimeMask(rMask)
		require.NoError(t, err)
		assert.Equal(t, tn, gotRMask)
		assert.Equal(t, tn, catalog.MustTnTypeByPrimeMask(rMask))

		gotRMaskWith, err := catalog.TnTypeByPrimeMaskWith(rMask, catalog.Rahn)
		require.NoError(t, err)
		assert.Equal(t, tn, gotRMaskWith)
		assert.Equal(t, tn, catalog.MustTnTypeByPrimeMaskWith(rMask, catalog.Rahn))

		gotFMaskWith, err := catalog.TnTypeByPrimeMaskWith(fMask, catalog.Forte)
		require.NoError(t, err)
		assert.Equal(t, tn, gotFMaskWith)
		assert.Equal(t, tn, catalog.MustTnTypeByPrimeMaskWith(fMask, catalog.Forte))

		gotAnyR, err := catalog.TnTypeByAnyPrimeMask(rMask)
		require.NoError(t, err)
		assert.Equal(t, tn, gotAnyR)
		assert.Equal(t, tn, catalog.MustTnTypeByAnyPrimeMask(rMask))

		gotAnyF, err := catalog.TnTypeByAnyPrimeMask(fMask)
		require.NoError(t, err)
		assert.Equal(t, tn, gotAnyF)
		assert.Equal(t, tn, catalog.MustTnTypeByAnyPrimeMask(fMask))

		rFormCopy := slices.Clone(rForm)
		gotRForm, err := catalog.TnTypeByPrimeForm(rForm)
		require.NoError(t, err)
		assert.Equal(t, tn, gotRForm)
		assert.Equal(t, rFormCopy, rForm, "input slice must not be mutated")
		assert.Equal(t, tn, catalog.MustTnTypeByPrimeForm(rForm))

		gotRFormWith, err := catalog.TnTypeByPrimeFormWith(rForm, catalog.Rahn)
		require.NoError(t, err)
		assert.Equal(t, tn, gotRFormWith)
		assert.Equal(t, tn, catalog.MustTnTypeByPrimeFormWith(rForm, catalog.Rahn))

		gotFFormWith, err := catalog.TnTypeByPrimeFormWith(fForm, catalog.Forte)
		require.NoError(t, err)
		assert.Equal(t, tn, gotFFormWith)
		assert.Equal(t, tn, catalog.MustTnTypeByPrimeFormWith(fForm, catalog.Forte))

		gotAnyFormR, err := catalog.TnTypeByAnyPrimeForm(rForm)
		require.NoError(t, err)
		assert.Equal(t, tn, gotAnyFormR)
		assert.Equal(t, tn, catalog.MustTnTypeByAnyPrimeForm(rForm))

		gotAnyFormF, err := catalog.TnTypeByAnyPrimeForm(fForm)
		require.NoError(t, err)
		assert.Equal(t, tn, gotAnyFormF)
		assert.Equal(t, tn, catalog.MustTnTypeByAnyPrimeForm(fForm))
	}

	assert.Len(t, rahnMasks, 352)
	assert.Len(t, forteMasks, 352)
	assert.Len(t, anyMasks, 369)

	assert.Equal(t, "3-11A", catalog.MustTnTypeByPrimeForm([]int{0, 3, 7}).String())
	assert.Equal(t, "3-11B", catalog.MustTnTypeByPrimeForm([]int{0, 4, 7}).String())

	t.Run("nil_and_empty_slice_resolve_to_0_1", func(t *testing.T) {
		t.Parallel()

		t01 := catalog.MustParseTnType("0-1")
		for _, emptyInput := range [][]int{nil, {}} {
			got1, err := catalog.TnTypeByPrimeForm(emptyInput)
			require.NoError(t, err)
			assert.Equal(t, t01, got1)

			got2, err := catalog.TnTypeByPrimeFormWith(emptyInput, catalog.Rahn)
			require.NoError(t, err)
			assert.Equal(t, t01, got2)

			got3, err := catalog.TnTypeByPrimeFormWith(emptyInput, catalog.Forte)
			require.NoError(t, err)
			assert.Equal(t, t01, got3)

			got4, err := catalog.TnTypeByAnyPrimeForm(emptyInput)
			require.NoError(t, err)
			assert.Equal(t, t01, got4)
		}
	})

	t.Run("exhaustive_4096_masks", func(t *testing.T) {
		t.Parallel()

		rejectedAny := 0
		for mask := uint64(0); mask < 4096; mask++ {
			if _, ok := rahnMasks[mask]; !ok {
				_, err := catalog.TnTypeByPrimeMask(mask)
				require.ErrorIs(t, err, catalog.ErrNotPrimeForm)
				_, err = catalog.TnTypeByPrimeMaskWith(mask, catalog.Rahn)
				require.ErrorIs(t, err, catalog.ErrNotPrimeForm)
			}
			if _, ok := forteMasks[mask]; !ok {
				_, err := catalog.TnTypeByPrimeMaskWith(mask, catalog.Forte)
				require.ErrorIs(t, err, catalog.ErrNotPrimeForm)
			}
			if _, ok := anyMasks[mask]; !ok {
				rejectedAny++
				_, err := catalog.TnTypeByAnyPrimeMask(mask)
				require.ErrorIs(t, err, catalog.ErrNotPrimeForm)
			}
		}
		assert.Equal(t, 4096-369, rejectedAny)
	})

	t.Run("invalid_masks_slices_and_algorithms", func(t *testing.T) {
		t.Parallel()

		for _, badMask := range []uint64{2, 4096, 0xFFFF, math.MaxUint64} {
			_, err := catalog.TnTypeByPrimeMask(badMask)
			require.ErrorIs(t, err, catalog.ErrNotPrimeForm)
			assert.Panics(t, func() { _ = catalog.MustTnTypeByPrimeMask(badMask) })

			_, err = catalog.TnTypeByPrimeMaskWith(badMask, catalog.Rahn)
			require.ErrorIs(t, err, catalog.ErrNotPrimeForm)
			assert.Panics(t, func() { _ = catalog.MustTnTypeByPrimeMaskWith(badMask, catalog.Rahn) })

			_, err = catalog.TnTypeByPrimeMaskWith(badMask, catalog.Forte)
			require.ErrorIs(t, err, catalog.ErrNotPrimeForm)
			assert.Panics(t, func() { _ = catalog.MustTnTypeByPrimeMaskWith(badMask, catalog.Forte) })

			_, err = catalog.TnTypeByAnyPrimeMask(badMask)
			require.ErrorIs(t, err, catalog.ErrNotPrimeForm)
			assert.Panics(t, func() { _ = catalog.MustTnTypeByAnyPrimeMask(badMask) })
		}

		badSlices := [][]int{
			{-1},
			{12},
			{0, 12},
			{0, -1, 4},
			{0, 0, 4},
			{0, 7, 4},
			{2, 6, 9},
			{0, 5, 9},
			{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		}
		for _, s := range badSlices {
			_, err := catalog.TnTypeByPrimeForm(s)
			require.ErrorIs(t, err, catalog.ErrNotPrimeForm)
			assert.Panics(t, func() { _ = catalog.MustTnTypeByPrimeForm(s) })

			_, err = catalog.TnTypeByPrimeFormWith(s, catalog.Rahn)
			require.ErrorIs(t, err, catalog.ErrNotPrimeForm)
			assert.Panics(t, func() { _ = catalog.MustTnTypeByPrimeFormWith(s, catalog.Rahn) })

			_, err = catalog.TnTypeByPrimeFormWith(s, catalog.Forte)
			require.ErrorIs(t, err, catalog.ErrNotPrimeForm)
			assert.Panics(t, func() { _ = catalog.MustTnTypeByPrimeFormWith(s, catalog.Forte) })

			_, err = catalog.TnTypeByAnyPrimeForm(s)
			require.ErrorIs(t, err, catalog.ErrNotPrimeForm)
			assert.Panics(t, func() { _ = catalog.MustTnTypeByAnyPrimeForm(s) })
		}

		// Cross-algorithm rejection for all 17 diverging TnTypes
		differCount := 0
		for _, tn := range catalog.TnTypes() {
			if !tn.RahnForteDiffer() {
				continue
			}
			differCount++
			rForm := tn.PrimeForm()
			fForm := tn.FortePrimeForm()
			rMask := tn.PrimeMask()
			fMask := tn.FortePrimeMask()

			_, err := catalog.TnTypeByPrimeForm(fForm)
			require.ErrorIs(t, err, catalog.ErrNotPrimeForm)
			assert.Panics(t, func() { _ = catalog.MustTnTypeByPrimeForm(fForm) })

			_, err = catalog.TnTypeByPrimeFormWith(fForm, catalog.Rahn)
			require.ErrorIs(t, err, catalog.ErrNotPrimeForm)
			assert.Panics(t, func() { _ = catalog.MustTnTypeByPrimeFormWith(fForm, catalog.Rahn) })

			_, err = catalog.TnTypeByPrimeFormWith(rForm, catalog.Forte)
			require.ErrorIs(t, err, catalog.ErrNotPrimeForm)
			assert.Panics(t, func() { _ = catalog.MustTnTypeByPrimeFormWith(rForm, catalog.Forte) })

			_, err = catalog.TnTypeByPrimeMask(fMask)
			require.ErrorIs(t, err, catalog.ErrNotPrimeForm)
			assert.Panics(t, func() { _ = catalog.MustTnTypeByPrimeMask(fMask) })

			_, err = catalog.TnTypeByPrimeMaskWith(fMask, catalog.Rahn)
			require.ErrorIs(t, err, catalog.ErrNotPrimeForm)
			assert.Panics(t, func() { _ = catalog.MustTnTypeByPrimeMaskWith(fMask, catalog.Rahn) })

			_, err = catalog.TnTypeByPrimeMaskWith(rMask, catalog.Forte)
			require.ErrorIs(t, err, catalog.ErrNotPrimeForm)
			assert.Panics(t, func() { _ = catalog.MustTnTypeByPrimeMaskWith(rMask, catalog.Forte) })
		}
		assert.Equal(t, 17, differCount)

		_, err := catalog.TnTypeByPrimeMaskWith(99999, catalog.Algorithm(99))
		require.ErrorIs(t, err, catalog.ErrInvalidAlgorithm)
		assert.Panics(t, func() { _ = catalog.MustTnTypeByPrimeMaskWith(0, catalog.Algorithm(99)) })

		_, err = catalog.TnTypeByPrimeFormWith([]int{-1}, catalog.Algorithm(99))
		require.ErrorIs(t, err, catalog.ErrInvalidAlgorithm)
		assert.Panics(t, func() { _ = catalog.MustTnTypeByPrimeFormWith([]int{0}, catalog.Algorithm(99)) })
	})
}

func FuzzParseForteClass(f *testing.F) {
	seeds := []string{
		"",
		"0-1",
		"1-1",
		"2-1",
		"3-11",
		"3-11A",
		"4-Z15",
		"4-15",
		"5-Z12",
		"5-20",
		"6-Z29",
		"6-35",
		"7-Z18",
		"8-26",
		"12-1",
		"04-1",
		"4-z15",
		"13-1",
		"6-51",
		"ForteClass(invalid)",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		fc, err := catalog.ParseForteClass(input)
		if err != nil {
			require.ErrorIs(t, err, catalog.ErrInvalidNotation)
			require.False(t, fc.IsValid())
			return
		}
		require.True(t, fc.IsValid())
		require.Equal(t, input, fc.String())
		require.Equal(t, fc, catalog.MustParseForteClass(input))
	})
}

func FuzzResolveForteClass(f *testing.F) {
	seeds := []string{
		"",
		"0-1",
		"3-11",
		"3-11A",
		"3-11B",
		"4-Z15",
		"4-Z15A",
		"4-Z15B",
		"5-Z12",
		"5-Z12A",
		"6-Z29",
		"3-1A",
		"4-15A",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		fc, err := catalog.ResolveForteClass(input)
		if err != nil {
			require.ErrorIs(t, err, catalog.ErrInvalidNotation)
			require.False(t, fc.IsValid())
			return
		}
		require.True(t, fc.IsValid())
		require.Equal(t, fc, catalog.MustResolveForteClass(input))
	})
}

func FuzzParseTnType(f *testing.F) {
	seeds := []string{
		"",
		"0-1",
		"3-1",
		"3-1A",
		"3-11",
		"3-11A",
		"3-11B",
		"4-Z15",
		"4-Z15A",
		"4-Z15B",
		"5-Z12",
		"5-Z12A",
		"6-Z29",
		"6-14A",
		"12-1",
		"TnType(invalid)",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		tn, err := catalog.ParseTnType(input)
		if err != nil {
			require.ErrorIs(t, err, catalog.ErrInvalidNotation)
			require.False(t, tn.IsValid())
			return
		}
		require.True(t, tn.IsValid())
		require.Equal(t, input, tn.String())
		require.Equal(t, tn, catalog.MustParseTnType(input))
	})
}

func TestCollections_And_Iterators(t *testing.T) {
	t.Parallel()

	t.Run("ForteClasses_and_AllForteClasses", func(t *testing.T) {
		t.Parallel()

		all := catalog.ForteClasses()
		require.Len(t, all, 224)

		var fromIter []catalog.ForteClass
		for fc := range catalog.AllForteClasses() {
			fromIter = append(fromIter, fc)
		}
		assert.Equal(t, all, fromIter)

		var early []catalog.ForteClass
		for fc := range catalog.AllForteClasses() {
			early = append(early, fc)
			if len(early) == 5 {
				break
			}
		}
		assert.Equal(t, all[:5], early)

		origFirst := all[0]
		all[0] = catalog.ForteClass{}
		assert.Equal(t, origFirst, catalog.ForteClasses()[0], "mutating ForteClasses() slice must not affect catalog")
	})

	t.Run("ForteClassesOfCardinality", func(t *testing.T) {
		t.Parallel()

		wantCounts := []int{1, 1, 6, 12, 29, 38, 50, 38, 29, 12, 6, 1, 1}
		total := 0
		for c, wantLen := range wantCounts {
			list, err := catalog.ForteClassesOfCardinality(c)
			require.NoError(t, err)
			assert.Len(t, list, wantLen)
			assert.Equal(t, list, catalog.MustForteClassesOfCardinality(c))
			for idx, fc := range list {
				assert.Equal(t, c, fc.Cardinality())
				assert.Equal(t, idx+1, fc.Ordinal())
			}
			orig := list[0]
			list[0] = catalog.ForteClass{}
			assert.Equal(t, orig, catalog.MustForteClassesOfCardinality(c)[0])
			total += len(list)
		}
		assert.Equal(t, 224, total)

		for _, badC := range []int{-1, -100, 13, 99, math.MinInt, math.MaxInt} {
			list, err := catalog.ForteClassesOfCardinality(badC)
			require.ErrorIs(t, err, catalog.ErrInvalidCardinality)
			assert.Nil(t, list)
			assert.Panics(t, func() { _ = catalog.MustForteClassesOfCardinality(badC) })
		}
	})

	t.Run("TnTypes_and_AllTnTypes", func(t *testing.T) {
		t.Parallel()

		all := catalog.TnTypes()
		require.Len(t, all, 352)

		var fromIter []catalog.TnType
		for tn := range catalog.AllTnTypes() {
			fromIter = append(fromIter, tn)
		}
		assert.Equal(t, all, fromIter)

		var early []catalog.TnType
		for tn := range catalog.AllTnTypes() {
			early = append(early, tn)
			if len(early) == 7 {
				break
			}
		}
		assert.Equal(t, all[:7], early)

		origFirst := all[0]
		all[0] = catalog.TnType{}
		assert.Equal(t, origFirst, catalog.TnTypes()[0], "mutating TnTypes() slice must not affect catalog")
	})

	t.Run("TnTypesOfCardinality", func(t *testing.T) {
		t.Parallel()

		wantCounts := []int{1, 1, 6, 19, 43, 66, 80, 66, 43, 19, 6, 1, 1}
		total := 0
		for c, wantLen := range wantCounts {
			list, err := catalog.TnTypesOfCardinality(c)
			require.NoError(t, err)
			assert.Len(t, list, wantLen)
			assert.Equal(t, list, catalog.MustTnTypesOfCardinality(c))
			for _, tn := range list {
				assert.Equal(t, c, tn.Cardinality())
			}
			orig := list[0]
			list[0] = catalog.TnType{}
			assert.Equal(t, orig, catalog.MustTnTypesOfCardinality(c)[0])
			total += len(list)
		}
		assert.Equal(t, 352, total)

		for _, badC := range []int{-1, -100, 13, 99, math.MinInt, math.MaxInt} {
			list, err := catalog.TnTypesOfCardinality(badC)
			require.ErrorIs(t, err, catalog.ErrInvalidCardinality)
			assert.Nil(t, list)
			assert.Panics(t, func() { _ = catalog.MustTnTypesOfCardinality(badC) })
		}
	})

	t.Run("CarterClasses_and_AllCarterClasses", func(t *testing.T) {
		t.Parallel()

		all := catalog.CarterClasses()
		require.Len(t, all, 135)

		var fromIter []catalog.CarterClass
		for cc := range catalog.AllCarterClasses() {
			fromIter = append(fromIter, cc)
		}
		assert.Equal(t, all, fromIter)

		var early []catalog.CarterClass
		for cc := range catalog.AllCarterClasses() {
			early = append(early, cc)
			if len(early) == 4 {
				break
			}
		}
		assert.Equal(t, all[:4], early)

		origFirst := all[0]
		all[0] = catalog.CarterClass{}
		assert.Equal(t, origFirst, catalog.CarterClasses()[0], "mutating CarterClasses() slice must not affect catalog")
	})

	t.Run("CarterClassesOfCardinality", func(t *testing.T) {
		t.Parallel()

		supported := map[int]int{2: 6, 3: 12, 4: 29, 6: 50, 7: 38}
		total := 0
		for c, wantLen := range supported {
			list, err := catalog.CarterClassesOfCardinality(c)
			require.NoError(t, err)
			assert.Len(t, list, wantLen)
			assert.Equal(t, list, catalog.MustCarterClassesOfCardinality(c))
			for idx, cc := range list {
				assert.Equal(t, c, cc.Cardinality())
				assert.Equal(t, idx+1, cc.Number())
			}
			orig := list[0]
			list[0] = catalog.CarterClass{}
			assert.Equal(t, orig, catalog.MustCarterClassesOfCardinality(c)[0])
			total += len(list)
		}
		assert.Equal(t, 135, total)

		for _, noCarterC := range []int{0, 1, 5, 8, 9, 10, 11, 12} {
			list, err := catalog.CarterClassesOfCardinality(noCarterC)
			require.ErrorIs(t, err, catalog.ErrNoCarterCatalog)
			assert.Nil(t, list)
			assert.Panics(t, func() { _ = catalog.MustCarterClassesOfCardinality(noCarterC) })
		}

		for _, badC := range []int{-1, 13, math.MinInt, math.MaxInt} {
			list, err := catalog.CarterClassesOfCardinality(badC)
			require.ErrorIs(t, err, catalog.ErrInvalidCardinality)
			assert.Nil(t, list)
			assert.Panics(t, func() { _ = catalog.MustCarterClassesOfCardinality(badC) })
		}
	})

	t.Run("ZPairs_and_AllZPairs", func(t *testing.T) {
		t.Parallel()

		all := catalog.ZPairs()
		require.Len(t, all, 23)

		var fromIter []catalog.ZPair
		for zp := range catalog.AllZPairs() {
			fromIter = append(fromIter, zp)
		}
		assert.Equal(t, all, fromIter)

		var early []catalog.ZPair
		for zp := range catalog.AllZPairs() {
			early = append(early, zp)
			if len(early) == 3 {
				break
			}
		}
		assert.Equal(t, all[:3], early)

		origFirst := all[0]
		all[0] = catalog.ZPair{}
		assert.Equal(t, origFirst, catalog.ZPairs()[0], "mutating ZPairs() slice must not affect catalog")
	})

	t.Run("ZPairsOfCardinality", func(t *testing.T) {
		t.Parallel()

		wantCounts := map[int]int{
			0: 0, 1: 0, 2: 0, 3: 0,
			4: 1, 5: 3, 6: 15, 7: 3, 8: 1,
			9: 0, 10: 0, 11: 0, 12: 0,
		}
		total := 0
		for c := 0; c <= 12; c++ {
			list, err := catalog.ZPairsOfCardinality(c)
			require.NoError(t, err)
			require.NotNil(t, list, "ZPairsOfCardinality(%d) must return non-nil slice", c)
			assert.Len(t, list, wantCounts[c])
			assert.Equal(t, list, catalog.MustZPairsOfCardinality(c))
			if len(list) > 0 {
				orig := list[0]
				list[0] = catalog.ZPair{}
				assert.Equal(t, orig, catalog.MustZPairsOfCardinality(c)[0])
			}
			total += len(list)
		}
		assert.Equal(t, 23, total)

		for _, badC := range []int{-1, 13, math.MinInt, math.MaxInt} {
			list, err := catalog.ZPairsOfCardinality(badC)
			require.ErrorIs(t, err, catalog.ErrInvalidCardinality)
			assert.Nil(t, list)
			assert.Panics(t, func() { _ = catalog.MustZPairsOfCardinality(badC) })
		}
	})
}

func TestCarterByNumber_And_Cardinalities(t *testing.T) {
	t.Parallel()

	cards := catalog.CarterCardinalities()
	assert.Equal(t, []int{2, 3, 4, 6, 7}, cards)
	cards[0] = 99
	assert.Equal(t, []int{2, 3, 4, 6, 7}, catalog.CarterCardinalities(), "CarterCardinalities must return a fresh slice")

	for c := -2; c <= 14; c++ {
		want := c == 2 || c == 3 || c == 4 || c == 6 || c == 7
		assert.Equal(t, want, catalog.HasCarterCardinality(c))
	}

	spotChecks := []struct {
		c         int
		num       int
		wantForte string
		wantStr   string
	}{
		{c: 2, num: 1, wantForte: "2-1", wantStr: "Carter(2, 1): 2-1"},
		{c: 2, num: 6, wantForte: "2-6", wantStr: "Carter(2, 6): 2-6"},
		{c: 3, num: 1, wantForte: "3-12", wantStr: "Carter(3, 1): 3-12"},
		{c: 3, num: 4, wantForte: "3-1", wantStr: "Carter(3, 4): 3-1"},
		{c: 3, num: 6, wantForte: "3-11", wantStr: "Carter(3, 6): 3-11"},
		{c: 3, num: 12, wantForte: "3-2", wantStr: "Carter(3, 12): 3-2"},
		{c: 4, num: 1, wantForte: "4-1", wantStr: "Carter(4, 1): 4-1"},
		{c: 4, num: 18, wantForte: "4-Z15", wantStr: "Carter(4, 18): 4-Z15"},
		{c: 4, num: 23, wantForte: "4-Z29", wantStr: "Carter(4, 23): 4-Z29"},
		{c: 4, num: 29, wantForte: "4-27", wantStr: "Carter(4, 29): 4-27"},
		{c: 6, num: 1, wantForte: "6-35", wantStr: "Carter(6, 1): 6-35"},
		{c: 6, num: 4, wantForte: "6-1", wantStr: "Carter(6, 4): 6-1"},
		{c: 6, num: 32, wantForte: "6-Z29", wantStr: "Carter(6, 32): 6-Z29"},
		{c: 6, num: 35, wantForte: "6-Z17", wantStr: "Carter(6, 35): 6-Z17"},
		{c: 6, num: 50, wantForte: "6-Z36", wantStr: "Carter(6, 50): 6-Z36"},
		{c: 7, num: 1, wantForte: "7-1", wantStr: "Carter(7, 1): 7-1"},
		{c: 7, num: 5, wantForte: "7-Z12", wantStr: "Carter(7, 5): 7-Z12"},
		{c: 7, num: 7, wantForte: "7-35", wantStr: "Carter(7, 7): 7-35"},
		{c: 7, num: 35, wantForte: "7-Z18", wantStr: "Carter(7, 35): 7-Z18"},
		{c: 7, num: 38, wantForte: "7-31", wantStr: "Carter(7, 38): 7-31"},
	}

	for _, sc := range spotChecks {
		cc, err := catalog.CarterByNumber(sc.c, sc.num)
		require.NoError(t, err)
		assert.True(t, cc.IsValid())
		assert.Equal(t, sc.c, cc.Cardinality())
		assert.Equal(t, sc.num, cc.Number())
		assert.Equal(t, sc.wantForte, cc.ForteClass().String())
		assert.Equal(t, sc.wantStr, cc.String())
		assert.Equal(t, cc, catalog.MustCarterByNumber(sc.c, sc.num))
	}

	errCases := []struct {
		name    string
		c       int
		num     int
		wantErr error
	}{
		{name: "c_neg", c: -1, num: 1, wantErr: catalog.ErrInvalidCardinality},
		{name: "c_13", c: 13, num: 1, wantErr: catalog.ErrInvalidCardinality},
		{name: "priority_c_neg_over_num_0", c: -1, num: 0, wantErr: catalog.ErrInvalidCardinality},
		{name: "c_0_no_carter", c: 0, num: 1, wantErr: catalog.ErrNoCarterCatalog},
		{name: "c_1_no_carter", c: 1, num: 1, wantErr: catalog.ErrNoCarterCatalog},
		{name: "c_5_no_carter", c: 5, num: 1, wantErr: catalog.ErrNoCarterCatalog},
		{name: "c_8_no_carter", c: 8, num: 1, wantErr: catalog.ErrNoCarterCatalog},
		{name: "c_9_no_carter", c: 9, num: 1, wantErr: catalog.ErrNoCarterCatalog},
		{name: "c_10_no_carter", c: 10, num: 1, wantErr: catalog.ErrNoCarterCatalog},
		{name: "c_11_no_carter", c: 11, num: 1, wantErr: catalog.ErrNoCarterCatalog},
		{name: "c_12_no_carter", c: 12, num: 1, wantErr: catalog.ErrNoCarterCatalog},
		{name: "priority_no_carter_over_bad_num", c: 5, num: 0, wantErr: catalog.ErrNoCarterCatalog},
		{name: "c2_num0", c: 2, num: 0, wantErr: catalog.ErrInvalidCarterNumber},
		{name: "c2_num7", c: 2, num: 7, wantErr: catalog.ErrInvalidCarterNumber},
		{name: "c3_num0", c: 3, num: 0, wantErr: catalog.ErrInvalidCarterNumber},
		{name: "c3_num13", c: 3, num: 13, wantErr: catalog.ErrInvalidCarterNumber},
		{name: "c4_num30", c: 4, num: 30, wantErr: catalog.ErrInvalidCarterNumber},
		{name: "c6_num51", c: 6, num: 51, wantErr: catalog.ErrInvalidCarterNumber},
		{name: "c7_num39", c: 7, num: 39, wantErr: catalog.ErrInvalidCarterNumber},
	}

	for _, tc := range errCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cc, err := catalog.CarterByNumber(tc.c, tc.num)
			require.ErrorIs(t, err, tc.wantErr)
			assert.False(t, cc.IsValid())
			assert.Panics(t, func() { _ = catalog.MustCarterByNumber(tc.c, tc.num) })
		})
	}
}

func TestZeroValues_AllTypes(t *testing.T) {
	t.Parallel()

	t.Run("ForteClass_zero", func(t *testing.T) {
		t.Parallel()

		var z catalog.ForteClass
		valid := catalog.MustParseForteClass("3-11")

		assert.False(t, z.IsValid())
		assert.Equal(t, 0, z.Cardinality())
		assert.Equal(t, 0, z.Ordinal())
		assert.False(t, z.IsZ())
		assert.Equal(t, "ForteClass(invalid)", z.String())
		assert.True(t, z == catalog.ForteClass{})
		assert.False(t, z.Equal(z))
		assert.False(t, z.Equal(valid))
		assert.False(t, valid.Equal(z))
		assert.Equal(t, 0, z.Compare(catalog.ForteClass{}))
		assert.Equal(t, -1, z.Compare(valid))
		assert.Equal(t, 1, valid.Compare(z))

		assert.Nil(t, z.PrimeForm())
		assert.Nil(t, z.FortePrimeForm())
		assert.Equal(t, uint64(0), z.PrimeMask())
		assert.Equal(t, uint64(0), z.FortePrimeMask())
		assert.False(t, z.RahnForteDiffer())

		pf, err := z.PrimeFormWith(catalog.Rahn)
		require.ErrorIs(t, err, catalog.ErrInvalidValue)
		assert.Nil(t, pf)
		_, err = z.PrimeFormWith(catalog.Algorithm(99))
		require.ErrorIs(t, err, catalog.ErrInvalidValue)
		assert.Panics(t, func() { _ = z.MustPrimeFormWith(catalog.Rahn) })

		pm, err := z.PrimeMaskWith(catalog.Rahn)
		require.ErrorIs(t, err, catalog.ErrInvalidValue)
		assert.Equal(t, uint64(0), pm)
		_, err = z.PrimeMaskWith(catalog.Algorithm(99))
		require.ErrorIs(t, err, catalog.ErrInvalidValue)
		assert.Panics(t, func() { _ = z.MustPrimeMaskWith(catalog.Rahn) })

		assert.Equal(t, [6]int{}, z.IntervalVector())
		assert.Nil(t, z.IntervalVectorSlice())
		assert.Equal(t, [8]int{}, z.InvarianceVector())
		assert.Nil(t, z.InvarianceVectorSlice())
		assert.Equal(t, 0, z.TnSymmetry())
		assert.Equal(t, 0, z.InSymmetry())
		assert.Equal(t, 0, z.DegreeOfSymmetry())
		assert.Equal(t, 0, z.OrbitSize())
		assert.False(t, z.IsInversionallySymmetric())
		assert.False(t, z.IsChiral())
		assert.Equal(t, 0, z.TnTypeCount())
		assert.Equal(t, 0, z.MinSpan())
		assert.Equal(t, catalog.ForteClass{}, z.Complement())
		assert.False(t, z.IsSelfComplementary())

		zp, ok := z.ZPartner()
		assert.False(t, ok)
		assert.Equal(t, catalog.ForteClass{}, zp)

		zpair, ok := z.ZPair()
		assert.False(t, ok)
		assert.Equal(t, catalog.ZPair{}, zpair)

		assert.Equal(t, catalog.ForteClass{}, z.M5Partner())
		assert.False(t, z.IsM5Invariant())
		assert.Nil(t, z.TnTypes())
		assert.Equal(t, catalog.TnType{}, z.TnTypeA())

		tnB, ok := z.TnTypeB()
		assert.False(t, ok)
		assert.Equal(t, catalog.TnType{}, tnB)

		assert.False(t, z.HasCarter())
		cNum, ok := z.CarterNumber()
		assert.False(t, ok)
		assert.Equal(t, 0, cNum)

		cc, ok := z.Carter()
		assert.False(t, ok)
		assert.Equal(t, catalog.CarterClass{}, cc)
	})

	t.Run("TnType_zero", func(t *testing.T) {
		t.Parallel()

		var z catalog.TnType
		valid := catalog.MustParseTnType("3-11A")

		assert.False(t, z.IsValid())
		assert.Equal(t, catalog.ForteClass{}, z.ForteClass())
		assert.Equal(t, 0, z.Cardinality())
		assert.Equal(t, 0, z.Ordinal())
		assert.False(t, z.IsZ())
		assert.Equal(t, catalog.SuffixNone, z.Suffix())
		assert.False(t, z.HasSuffix())
		assert.Equal(t, "TnType(invalid)", z.String())
		assert.True(t, z == catalog.TnType{})
		assert.False(t, z.Equal(z))
		assert.False(t, z.Equal(valid))
		assert.False(t, valid.Equal(z))
		assert.Equal(t, 0, z.Compare(catalog.TnType{}))
		assert.Equal(t, -1, z.Compare(valid))
		assert.Equal(t, 1, valid.Compare(z))

		assert.Nil(t, z.PrimeForm())
		assert.Nil(t, z.FortePrimeForm())
		assert.Equal(t, uint64(0), z.PrimeMask())
		assert.Equal(t, uint64(0), z.FortePrimeMask())
		assert.False(t, z.RahnForteDiffer())

		pf, err := z.PrimeFormWith(catalog.Rahn)
		require.ErrorIs(t, err, catalog.ErrInvalidValue)
		assert.Nil(t, pf)
		_, err = z.PrimeFormWith(catalog.Algorithm(99))
		require.ErrorIs(t, err, catalog.ErrInvalidValue)
		assert.Panics(t, func() { _ = z.MustPrimeFormWith(catalog.Rahn) })

		pm, err := z.PrimeMaskWith(catalog.Rahn)
		require.ErrorIs(t, err, catalog.ErrInvalidValue)
		assert.Equal(t, uint64(0), pm)
		_, err = z.PrimeMaskWith(catalog.Algorithm(99))
		require.ErrorIs(t, err, catalog.ErrInvalidValue)
		assert.Panics(t, func() { _ = z.MustPrimeMaskWith(catalog.Rahn) })

		assert.Equal(t, [6]int{}, z.IntervalVector())
		assert.Nil(t, z.IntervalVectorSlice())
		assert.Equal(t, [8]int{}, z.InvarianceVector())
		assert.Nil(t, z.InvarianceVectorSlice())
		assert.Equal(t, 0, z.TnSymmetry())
		assert.Equal(t, 0, z.InSymmetry())
		assert.Equal(t, 0, z.OrbitSize())
		assert.False(t, z.IsInversionallySymmetric())
		assert.False(t, z.IsChiral())
		assert.Equal(t, 0, z.MinSpan())
		assert.Equal(t, catalog.TnType{}, z.Inversion())
		assert.Equal(t, catalog.TnType{}, z.Complement())
		assert.False(t, z.IsSelfComplementary())

		zp, ok := z.ZPartner()
		assert.False(t, ok)
		assert.Equal(t, catalog.ForteClass{}, zp)

		assert.Equal(t, catalog.ForteClass{}, z.M5Partner())
		assert.False(t, z.IsM5Invariant())
		assert.False(t, z.HasCarter())

		cNum, ok := z.CarterNumber()
		assert.False(t, ok)
		assert.Equal(t, 0, cNum)

		cc, ok := z.Carter()
		assert.False(t, ok)
		assert.Equal(t, catalog.CarterClass{}, cc)
	})

	t.Run("CarterClass_zero", func(t *testing.T) {
		t.Parallel()

		var z catalog.CarterClass
		valid := catalog.MustCarterByNumber(3, 1)

		assert.False(t, z.IsValid())
		assert.Equal(t, 0, z.Cardinality())
		assert.Equal(t, 0, z.Number())
		assert.Equal(t, catalog.ForteClass{}, z.ForteClass())
		assert.Equal(t, "CarterClass(invalid)", z.String())
		assert.True(t, z == catalog.CarterClass{})
		assert.False(t, z.Equal(z))
		assert.False(t, z.Equal(valid))
		assert.False(t, valid.Equal(z))
		assert.Equal(t, 0, z.Compare(catalog.CarterClass{}))
		assert.Equal(t, -1, z.Compare(valid))
		assert.Equal(t, 1, valid.Compare(z))
	})

	t.Run("ZPair_zero", func(t *testing.T) {
		t.Parallel()

		var z catalog.ZPair
		valid := catalog.ZPairs()[0]

		assert.False(t, z.IsValid())
		assert.Equal(t, 0, z.Cardinality())
		assert.Equal(t, catalog.ForteClass{}, z.First())
		assert.Equal(t, catalog.ForteClass{}, z.Second())
		assert.Equal(t, [6]int{}, z.IntervalVector())
		assert.Nil(t, z.IntervalVectorSlice())
		assert.Equal(t, "ZPair(invalid)", z.String())
		assert.True(t, z == catalog.ZPair{})
		assert.False(t, z.Equal(z))
		assert.False(t, z.Equal(valid))
		assert.False(t, valid.Equal(z))
		assert.Equal(t, 0, z.Compare(catalog.ZPair{}))
		assert.Equal(t, -1, z.Compare(valid))
		assert.Equal(t, 1, valid.Compare(z))
	})
}
