package mode_test

import (
	"fmt"
	"math"
	"slices"
	"testing"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Profiles are the three explicit product definitions of task 18.2.
var familySteps = [3][7]int{{2, 2, 1, 2, 2, 2, 1}, {2, 1, 2, 2, 2, 2, 1}, {1, 2, 2, 2, 2, 2, 1}}

type witnessKey struct {
	family                    mode.IntervalFamily
	rotation, from, to, cycle int
}

func witnessKeys(ws []mode.IntervalWitness) []witnessKey {
	out := make([]witnessKey, 0, len(ws))
	for _, w := range ws {
		out = append(out, witnessKey{w.Family, w.Rotation, w.From.Number, w.To.Number, w.To.Cycle})
	}
	return out
}

// This bounded oracle uses the written seven-letter coordinate and manual
// semitone profiles, independently of FindInterval and of mode operations.
func familyOracle(want interval.Interval, families []mode.IntervalFamily, window *mode.CycleWindow) ([]witnessKey, map[witnessKey]template.Template) {
	result := make([]witnessKey, 0)
	templates := make(map[witnessKey]template.Template)
	for _, family := range families {
		steps := familySteps[int(family)-1]
		for rotation := range 7 {
			offsets := [7]int{}
			adjacent := make([]interval.Interval, 7)
			for index := range 7 {
				step := steps[(index+rotation)%7]
				if step == 1 {
					adjacent[index] = interval.MinorSecond()
				} else {
					adjacent[index] = interval.MajorSecond()
				}
				if index < 6 {
					offsets[index+1] = offsets[index] + step
				}
			}
			tmpl := template.MustClosedFromIntervals(adjacent)
			for from := range 7 {
				for to := range 7 {
					for cycle := -3; cycle <= 3; cycle++ {
						if window != nil && (cycle < window.First || cycle > window.Last) {
							continue
						}
						written := to + 7*cycle - from
						direction := interval.Up
						if written < 0 {
							direction = interval.Down
							written = -written
						}
						size := written + 1
						if size != want.Size() {
							continue
						}
						signed := offsets[to] + 12*cycle - offsets[from]
						if direction == interval.Down {
							signed = -signed
						}
						baseline := []int{0, 2, 4, 5, 7, 9, 11}[(size-1)%7] + 12*((size-1)/7)
						expected := interval.MustFromDeviation(size, altForReading(signed-baseline))
						if direction == interval.Down {
							expected = expected.MustDown()
						}
						if !expected.Equal(want) {
							continue
						}
						key := witnessKey{family, rotation, from + 1, to + 1, cycle}
						result = append(result, key)
						templates[key] = tmpl
					}
				}
			}
		}
	}
	return result, templates
}

func TestFindIntervalFamilies(t *testing.T) {
	t.Parallel()
	for _, text := range []string{"P1", "P8", "m3", "A5", "d3", "A2", "M10", "-m3", "-P8"} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			want := interval.MustParse(text)
			scope := mode.IntervalScope{Families: []mode.IntervalFamily{mode.FamilyIII, mode.FamilyI, mode.FamilyII, mode.FamilyIII}}
			got, err := mode.FindInterval(want, scope)
			require.NoError(t, err)
			keys, tmpls := familyOracle(want, []mode.IntervalFamily{mode.FamilyI, mode.FamilyII, mode.FamilyIII}, nil)
			assert.Equal(t, keys, witnessKeys(got))
			for _, w := range got {
				assert.Zero(t, w.From.Cycle)
				assert.True(t, want.Equal(w.Interval))
				key := witnessKey{w.Family, w.Rotation, w.From.Number, w.To.Number, w.To.Cycle}
				assert.True(t, tmpls[key].Equal(w.Template), "structural rotation %v", key)
			}
			assert.Equal(t, keys, witnessKeys(mode.MustFindInterval(want, scope)))
			// The API must return independent slices on each invocation.
			if len(got) > 0 {
				got[0].From.Number = 999
			}
			again, err := mode.FindInterval(want, scope)
			require.NoError(t, err)
			assert.Equal(t, keys, witnessKeys(again))
		})
	}
	// Musical presence claims are distinct from the classical quality flag.
	for _, tc := range []struct {
		text      string
		present   [3]bool
		classical bool
	}{{"A5", [3]bool{false, true, true}, false}, {"d3", [3]bool{false, false, true}, false}, {"A2", [3]bool{}, false}, {"A4", [3]bool{true, true, true}, true}} {
		want := interval.MustParse(tc.text)
		assert.Equal(t, tc.classical, want.IsDiatonic())
		for index, family := range []mode.IntervalFamily{mode.FamilyI, mode.FamilyII, mode.FamilyIII} {
			got, err := mode.FindInterval(want, mode.IntervalScope{Families: []mode.IntervalFamily{family}})
			require.NoError(t, err)
			assert.Equal(t, tc.present[index], len(got) > 0, "%v in family %v", want, family)
		}
	}
}

func TestFindIntervalExplicitAndWindows(t *testing.T) {
	t.Parallel()
	want := interval.PerfectOctave()
	window := &mode.CycleWindow{First: 1, Last: 1}
	scope := mode.IntervalScope{Families: []mode.IntervalFamily{mode.FamilyI}, Templates: []template.Template{template.Ionian(), template.Ionian()}, Cycles: window}
	got, err := mode.FindInterval(want, scope)
	require.NoError(t, err)
	keys, _ := familyOracle(want, []mode.IntervalFamily{mode.FamilyI}, window)
	for n := 1; n <= 7; n++ {
		keys = append(keys, witnessKey{0, 0, n, n, 1})
	}
	assert.Equal(t, keys, witnessKeys(got))
	for _, w := range got[len(got)-7:] {
		assert.True(t, w.Template.Equal(template.Ionian()))
	}
	window.First = -5
	scope.Families[0] = mode.FamilyIII
	scope.Templates[0] = template.MelodicMinor()
	assert.Equal(t, keys, witnessKeys(got), "returned results do not depend on caller slices or window")
	open := template.MustOpenFromIntervals(intervals("M2 m2"))
	found, err := mode.FindInterval(interval.MinorThird(), mode.IntervalScope{Templates: []template.Template{open}})
	require.NoError(t, err)
	assert.Equal(t, []witnessKey{{0, 0, 1, 3, 0}}, witnessKeys(found))
	found, err = mode.FindInterval(interval.MinorThird(), mode.IntervalScope{Templates: []template.Template{open}, Cycles: &mode.CycleWindow{First: 1, Last: 2}})
	require.NoError(t, err)
	assert.Empty(t, found)
	// A three-place P4 cycle: C D E | F G A | Bb C D ...
	fourth := template.MustClosedFromIntervals(intervals("M2 M2 m2"))
	found, err = mode.FindInterval(interval.PerfectFourth(), mode.IntervalScope{Templates: []template.Template{fourth}, Cycles: &mode.CycleWindow{First: 1, Last: 1}})
	require.NoError(t, err)
	assert.Equal(t, []witnessKey{{0, 0, 1, 1, 1}, {0, 0, 2, 2, 1}, {0, 0, 3, 3, 1}}, witnessKeys(found))
	// A2 in harmonic minor does not turn it into one of the three families.
	harmonic := template.MustClosedFromIntervals(intervals("M2 m2 M2 M2 m2 A2 m2"))
	found, err = mode.FindInterval(interval.AugmentedSecond(), mode.IntervalScope{Templates: []template.Template{harmonic}})
	require.NoError(t, err)
	assert.Equal(t, []witnessKey{{0, 0, 6, 7, 0}}, witnessKeys(found))
	// Very large sizes use exact written coordinates, not overflowing virtual indices.
	huge := interval.MustFromDeviation(math.MaxInt, altForReading(0))
	found, err = mode.FindInterval(huge, mode.IntervalScope{Families: []mode.IntervalFamily{mode.FamilyI}, Cycles: &mode.CycleWindow{First: 0, Last: 0}})
	require.NoError(t, err)
	assert.Empty(t, found)
	// Window filtering is inclusive in both directions.
	for _, window := range []mode.CycleWindow{{-2, -1}, {0, 0}, {-1, 1}} {
		got, err := mode.FindInterval(interval.MustParse("-M10"), mode.IntervalScope{Families: []mode.IntervalFamily{mode.FamilyII}, Cycles: &window})
		require.NoError(t, err)
		keys, _ := familyOracle(interval.MustParse("-M10"), []mode.IntervalFamily{mode.FamilyII}, &window)
		assert.Equal(t, keys, witnessKeys(got))
	}
}

func TestFindIntervalErrors(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name  string
		want  interval.Interval
		scope mode.IntervalScope
		err   error
	}{
		{"empty", interval.MajorThird(), mode.IntervalScope{}, mode.ErrEmptyInput},
		{"unset want", interval.Interval{}, mode.IntervalScope{Families: []mode.IntervalFamily{mode.FamilyI}}, mode.ErrUnset},
		{"unset explicit", interval.MajorThird(), mode.IntervalScope{Templates: []template.Template{{}}}, mode.ErrUnset},
		{"zero family", interval.MajorThird(), mode.IntervalScope{Families: []mode.IntervalFamily{0}}, mode.ErrInvalidArgument},
		{"unknown family", interval.MajorThird(), mode.IntervalScope{Families: []mode.IntervalFamily{4}}, mode.ErrInvalidArgument},
		{"reversed window", interval.MajorThird(), mode.IntervalScope{Families: []mode.IntervalFamily{mode.FamilyI}, Cycles: &mode.CycleWindow{First: 1, Last: -1}}, mode.ErrInvalidArgument},
		{"non octave needs window", interval.MajorThird(), mode.IntervalScope{Templates: []template.Template{template.MustClosedFromIntervals(intervals("M2 M2 m2"))}}, mode.ErrInvalidArgument},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := mode.FindInterval(tc.want, tc.scope)
			assert.ErrorIs(t, err, tc.err)
			assert.Nil(t, got)
		})
	}
	readingPanic(t, func() { mode.MustFindInterval(interval.MajorThird(), mode.IntervalScope{}) }, mode.ErrEmptyInput)
}

func TestFindIntervalDeterministicConcurrent(t *testing.T) {
	t.Parallel()
	scope := mode.IntervalScope{Families: []mode.IntervalFamily{mode.FamilyI, mode.FamilyII, mode.FamilyIII}}
	want := interval.AugmentedFifth()
	expected, _ := familyOracle(want, scope.Families, nil)
	results := make(chan string, 8)
	for range 8 {
		go func() {
			got, err := mode.FindInterval(want, scope)
			if err != nil {
				results <- fmt.Sprint(err)
				return
			}
			if !slices.Equal(expected, witnessKeys(got)) {
				results <- "different witnesses"
				return
			}
			results <- ""
		}()
	}
	for range 8 {
		assert.Empty(t, <-results)
	}
}

func TestFindIntervalMaxSizeWithoutVirtualOverflow(t *testing.T) {
	t.Parallel()
	// MaxInt has simple size seven. Lifting M7 by complete written octaves
	// gives the exact same witnesses, with their end cycle shifted analytically.
	want := interval.MustFromDeviation(math.MaxInt, altForReading(0))
	got, err := mode.FindInterval(want, mode.IntervalScope{Families: []mode.IntervalFamily{mode.FamilyI}})
	require.NoError(t, err)
	expected, _ := familyOracle(interval.MustParse("M7"), []mode.IntervalFamily{mode.FamilyI}, nil)
	for index := range expected {
		expected[index].cycle += (math.MaxInt - 7) / 7
	}
	assert.Equal(t, expected, witnessKeys(got))
	for _, w := range got {
		assert.True(t, want.Equal(w.Interval))
		assert.Zero(t, w.From.Cycle)
	}
}
