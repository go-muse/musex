package tendency_test

import (
	"fmt"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/problemsolving/tendency"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClassificationTables(t *testing.T) {
	t.Parallel()
	ids := []tendency.ID{tendency.AMConsonance, tendency.OMTHarmonic, tendency.OMTMelodic, tendency.OMTCounterpoint, tendency.MT21Consonance, tendency.SeredaConsonance, tendency.WPConsonance}
	// Source: consonance-dissonance, reference table, var-4 through var-9.
	rows := []struct{ input, kinds string }{
		{"P1", "CCCCC?C"}, {"m2", "DDCDDDD"}, {"M2", "DDCDDDD"},
		{"m3", "CCCCCCC"}, {"M3", "CCCCCCC"}, {"P4", "CDCDDDC"},
		{"A4", "DDDDD?D"}, {"d5", "DDDDD?D"}, {"P5", "CCCCC?C"},
		{"m6", "CCCCCCC"}, {"M6", "CCCCCCC"}, {"m7", "DDDDDDD"},
		{"M7", "DDDDDDD"}, {"P8", "CCCCC?C"}, {"A2", "?DDD???"}, {"dd5", "?DDD???"},
	}
	for _, row := range rows {
		for index, id := range ids {
			t.Run(fmt.Sprintf("%s/%s", id, row.input), func(t *testing.T) {
				t.Parallel()
				s := built(id)
				input := iv(row.input).MustDown()
				r := s.MustClassify(input)
				assert.Equal(t, input, r.Input)
				assert.Equal(t, iv(row.input), r.Classified)
				kind := row.kinds[index]
				if kind == '?' {
					assert.Equal(t, tendency.MissingRule, r.Knowledge)
					assert.Equal(t, tendency.Classification{}, r.Value)
					return
				}
				assertAnswer(t, s, r.Answer, tendency.Known)
				want := tendency.Dissonance
				if kind == 'C' {
					want = tendency.Consonance
				}
				assert.Equal(t, want, r.Value.Kind)
				subtype := tendency.ConsonanceUnset
				if kind == 'C' {
					switch id {
					case tendency.AMConsonance:
						switch row.input {
						case "P1", "P8":
							subtype = tendency.AbsoluteConsonance
						case "P4", "P5":
							subtype = tendency.PerfectConsonance
						default:
							subtype = tendency.ImperfectConsonance
						}
					case tendency.OMTCounterpoint, tendency.WPConsonance:
						switch row.input {
						case "P1", "P4", "P5", "P8":
							subtype = tendency.PerfectConsonance
						default:
							subtype = tendency.ImperfectConsonance
						}
					case tendency.SeredaConsonance:
						subtype = tendency.ImperfectConsonance
					}
				}
				assert.Equal(t, subtype, r.Value.Subtype)
				if id == tendency.AMConsonance && row.input == "P4" {
					assert.NotEmpty(t, r.Value.Caveat)
				}
			})
		}
	}
	for _, id := range ids {
		s := built(id)
		r := s.MustClassify(iv("M10"))
		reduced := id == tendency.OMTCounterpoint || id == tendency.MT21Consonance || id == tendency.SeredaConsonance
		if reduced {
			assert.Equal(t, tendency.Known, r.Knowledge)
			assert.Equal(t, iv("M3"), r.Classified)
		} else {
			assert.Equal(t, tendency.MissingRule, r.Knowledge)
			assert.Equal(t, iv("M10"), r.Classified)
		}
		assert.Equal(t, iv("P8"), s.MustClassify(iv("P8")).Classified)
		assert.Equal(t, tendency.MissingRule, s.MustClassify(interval.MustFromDeviation(3, alteration.MustFromFraction(-1, 2))).Knowledge)
	}
	for _, id := range []tendency.ID{tendency.OMTHarmonic, tendency.OMTMelodic} {
		assert.Equal(t, tendency.MissingRule, built(id).MustClassify(iv("A9")).Knowledge)
	}
}

func TestIntervalResolutions(t *testing.T) {
	t.Parallel()
	cases := []struct {
		source  string
		options [][3]string
	}{
		{"m2", [][3]string{{"-M2", "P1", "m3"}}},
		{"M2", [][3]string{{"-m2", "P1", "m3"}, {"-M2", "P1", "M3"}}},
		{"m7", [][3]string{{"P1", "-m2", "M6"}, {"P1", "-M2", "m6"}}},
		{"M7", [][3]string{{"P1", "-M2", "M6"}}},
		{"A4", [][3]string{{"-m2", "m2", "m6"}, {"-m2", "M2", "M6"}, {"-M2", "m2", "M6"}}},
		{"d5", [][3]string{{"m2", "-m2", "M3"}, {"m2", "-M2", "m3"}, {"M2", "-m2", "m3"}}},
	}
	for _, tc := range cases {
		t.Run(tc.source, func(t *testing.T) {
			t.Parallel()
			s := built(tendency.AMIntervalResolution)
			r := s.MustResolveInterval(iv(tc.source))
			assertAnswer(t, s, r.Answer, tendency.Known)
			require.Len(t, r.Allowed, len(tc.options))
			assert.Empty(t, r.Undetermined)
			actual := make([]tendency.Motions, 0, len(r.Allowed))
			for _, c := range r.Allowed {
				actual = append(actual, c.Motions)
				assert.False(t, c.Preferred)
				checkReasons(t, c.Reasons)
				check := s.MustCheckInterval(iv(tc.source), c.Motions)
				assert.Equal(t, tendency.Allowed, check.Decision)
			}
			want := make([]tendency.Motions, 0, len(tc.options))
			for _, o := range tc.options {
				motions := tendency.Motions{First: iv(o[0]), Second: iv(o[1])}
				want = append(want, motions)
				found := false
				for _, c := range r.Allowed {
					if c.Motions == motions {
						assert.Equal(t, iv(o[2]), c.Final)
						found = true
					}
				}
				assert.True(t, found)
			}
			assert.ElementsMatch(t, want, actual)
			for _, c := range r.Rejected {
				assert.Equal(t, tendency.Denied, s.MustCheckInterval(iv(tc.source), c.Motions).Decision)
			}
		})
	}
	s := built(tendency.AMIntervalResolution)
	for _, input := range []string{"P1", "m3", "M3", "P4", "P5", "m6", "M6", "P8"} {
		r := s.MustResolveInterval(iv(input))
		assert.Equal(t, tendency.Known, r.Knowledge)
		assert.Empty(t, r.Allowed)
		assert.Empty(t, r.Undetermined)
		assert.Empty(t, r.Rejected)
		assert.Contains(t, reasonRules(r.Reasons), "resolution.not-required")
	}
	for _, input := range []interval.Interval{iv("A2"), iv("M9"), iv("m2").MustDown(), interval.MustFromDeviation(2, alteration.MustFromFraction(-1, 2))} {
		r := s.MustResolveInterval(input)
		assert.Equal(t, tendency.OutsideScope, r.Knowledge)
		assert.Empty(t, r.Allowed)
		assert.Equal(t, tendency.Undetermined, s.MustCheckInterval(input, tendency.Motions{First: iv("P1"), Second: iv("P1")}).Decision)
	}
	assert.Equal(t, tendency.Denied, s.MustCheckInterval(iv("m2"), tendency.Motions{First: iv("m2").MustDown(), Second: iv("P1")}).Decision)
}

func TestModalResolutionsControlExamples(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name                       string
		m                          mode.Mode
		source                     tendency.Pair
		allowed, unknown, rejected []tendency.Pair
		preferred                  *tendency.Pair
	}{
		{"K11", ionian(), pair(2, 4), []tendency.Pair{pair(1, 3), pair(1, 5), pair(3, 5)}, []tendency.Pair{pair(3, 3)}, []tendency.Pair{}, new(pair(1, 3))},
		{"K12 diminished fifth", aeolian(), pair(2, 6), []tendency.Pair{pair(3, 5)}, []tendency.Pair{}, []tendency.Pair{pair(1, 5)}, new(pair(3, 5))},
		{"augmented fourth", ionian(), pair(4, 7), []tendency.Pair{{First: place(3, 0), Second: place(1, 1)}}, []tendency.Pair{}, []tendency.Pair{{First: place(5, 0), Second: place(1, 1)}}, new(tendency.Pair{First: place(3, 0), Second: place(1, 1)})},
		{"K21 compound final", ionian(), tendency.Pair{First: place(2, 0), Second: place(2, 1)}, []tendency.Pair{{First: place(3, 0), Second: place(1, 1)}}, []tendency.Pair{{First: place(1, 0), Second: place(3, 1)}}, []tendency.Pair{{First: place(1, 0), Second: place(1, 1)}, {First: place(3, 0), Second: place(3, 1)}}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s := built(tendency.AMModalResolution)
			r := s.MustResolveModal(tc.m, tc.source)
			assert.Equal(t, tc.source, r.Source)
			assert.Equal(t, tc.allowed, modalTargets(r.Allowed))
			assert.Equal(t, tc.unknown, modalTargets(r.Undetermined))
			assert.Equal(t, tc.rejected, modalTargets(r.Rejected))
			knowledge := tendency.Known
			if len(tc.unknown) > 0 {
				knowledge = tendency.Partial
			}
			assertAnswer(t, s, r.Answer, knowledge)
			for _, group := range []struct {
				values   []tendency.ModalCandidate
				decision tendency.Decision
			}{{r.Allowed, tendency.Allowed}, {r.Undetermined, tendency.Undetermined}, {r.Rejected, tendency.Denied}} {
				for _, c := range group.values {
					checkReasons(t, c.Reasons)
					assert.Equal(t, tc.source.First, c.First.From)
					assert.Equal(t, c.Target.First, c.First.To)
					assert.Equal(t, tc.source.Second, c.Second.From)
					assert.Equal(t, c.Target.Second, c.Second.To)
					assert.Equal(t, tendency.MustBetween(tc.m, c.Target.First, c.Target.Second).Interval, c.Final)
					assert.Equal(t, group.decision, s.MustCheckModal(tc.m, tc.source, c.Target).Decision)
					assert.Equal(t, tc.preferred != nil && *tc.preferred == c.Target, c.Preferred)
				}
			}
			if tc.name == "K21 compound final" {
				assert.Equal(t, iv("M10"), r.Undetermined[0].Final)
				assert.Contains(t, reasonRules(r.Undetermined[0].Reasons), "am.classification-missing")
				for _, c := range r.Rejected {
					assert.Contains(t, reasonRules(c.Reasons), "am.parallel-octaves")
				}
			}
		})
	}
	s := built(tendency.AMModalResolution)
	m := ionian()
	for _, source := range []tendency.Pair{pair(1, 3), pair(3, 5)} {
		r := s.MustResolveModal(m, source)
		assert.Equal(t, tendency.Known, r.Knowledge)
		assert.Empty(t, r.Allowed)
		assert.Empty(t, r.Undetermined)
		assert.Empty(t, r.Rejected)
		assert.Equal(t, tendency.Denied, s.MustCheckModal(m, source, source).Decision)
	}
	for _, source := range []tendency.Pair{{First: place(1, 0), Second: place(2, 1)}, pair(5, 2)} {
		assert.Equal(t, tendency.OutsideScope, s.MustResolveModal(m, source).Knowledge)
	}
	for _, tc := range []struct {
		name     string
		from, to tendency.Pair
		rule     string
	}{
		{"second unison", pair(2, 3), pair(3, 3), "am.second-to-unison"},
		{"seventh octave", pair(1, 7), tendency.Pair{First: place(1, 0), Second: place(1, 1)}, "am.seventh-to-octave"},
		{"parallel fifths", pair(2, 6), pair(1, 5), "am.parallel-fifths"},
		{"crossing", pair(2, 4), pair(5, 3), "am.crossing"},
		{"not support", pair(2, 4), pair(2, 3), "am.non-support"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := s.MustCheckModal(m, tc.from, tc.to)
			assert.Equal(t, tendency.Denied, r.Decision)
			assert.Contains(t, reasonRules(r.Reasons), tc.rule)
		})
	}
	// A known local violation dominates the independent converging-unison gap.
	r := s.MustCheckModal(m, pair(2, 6), pair(4, 4))
	assert.Equal(t, tendency.Denied, r.Decision)
}
