package mode_test

import (
	"math"
	"sync"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func readingPanic(t *testing.T, call func(), want error) {
	t.Helper()
	defer func() {
		value := recover()
		err, ok := value.(error)
		require.True(t, ok, "panic should contain error, got %T", value)
		assert.ErrorIs(t, err, want)
	}()
	call()
}

func assertRoleReading(t *testing.T, r mode.RoleReading, sys *system.System, p pitch.Pitch, text string, steps int, delta alteration.Alteration) {
	t.Helper()
	assert.True(t, p.Equal(r.Pitch), "pitch: got %v, want %v", r.Pitch, p)
	assert.True(t, interval.MustParse(text).Equal(r.Interval), "interval: got %v, want %v", r.Interval, text)
	assert.True(t, r.Shift.IsValid())
	require.NotNil(t, r.Shift.System())
	assert.True(t, sys.Equal(r.Shift.System()))
	assert.Equal(t, steps, r.Shift.Steps())
	assert.Equal(t, delta, r.Alteration)
}

func TestResolveRoleMusicalReadings(t *testing.T) {
	t.Parallel()
	traditional, quarter := system.Traditional(), system.QuarterTones24()
	boundaryA := system.MustNew(system.Definition{Name: "A boundary", Classes: 12, Boundary: letter.A, Offsets: [letter.Count]int{3, 5, 7, 8, 10, 0, 2}, Chromatic: 1, Alterations: system.Integers()})
	for _, tc := range []struct {
		name          string
		m             mode.Mode
		sys           *system.System
		tonic         pitch.Pitch
		from, to      mode.DegreePlace
		observed      int
		observedPitch pitch.Pitch
		expected      pitch.Pitch
		expectedText  string
		expectedSteps int
		candidate     pitch.Pitch
		candidateText string
		delta         alteration.Alteration
		match         bool
	}{
		{"natural minor confirms third", mode.MustFromCode("aeolian"), traditional, readingPitch(traditional, spelling.C(), 4), mode.DegreePlace{1, 0}, mode.DegreePlace{3, 0}, 3, readingPitch(traditional, spelling.DSharp(), 4), readingPitch(traditional, spelling.EFlat(), 4), "m3", 3, readingPitch(traditional, spelling.EFlat(), 4), "m3", altForReading(0), true},
		{"major second altered", mode.MustFromCode("ionian"), traditional, readingPitch(traditional, spelling.C(), 4), mode.DegreePlace{1, 0}, mode.DegreePlace{2, 0}, 3, readingPitch(traditional, spelling.EFlat(), 4), readingPitch(traditional, spelling.D(), 4), "M2", 2, readingPitch(traditional, spelling.DSharp(), 4), "A2", altForReading(1), false},
		{"major third altered", mode.MustFromCode("ionian"), traditional, readingPitch(traditional, spelling.C(), 4), mode.DegreePlace{1, 0}, mode.DegreePlace{3, 0}, 3, readingPitch(traditional, spelling.DSharp(), 4), readingPitch(traditional, spelling.E(), 4), "M3", 4, readingPitch(traditional, spelling.EFlat(), 4), "m3", altForReading(-1), false},
		{"descending", mode.MustFromCode("ionian"), traditional, readingPitch(traditional, spelling.C(), 4), mode.DegreePlace{3, 0}, mode.DegreePlace{1, 0}, -3, readingPitch(traditional, spelling.DFlat(), 4), readingPitch(traditional, spelling.C(), 4), "-M3", -4, readingPitch(traditional, spelling.CSharp(), 4), "-m3", altForReading(1), false},
		{"compound", mode.MustFromCode("ionian"), traditional, readingPitch(traditional, spelling.C(), 4), mode.DegreePlace{1, 0}, mode.DegreePlace{3, 1}, 15, readingPitch(traditional, spelling.DSharp(), 5), readingPitch(traditional, spelling.E(), 5), "M10", 16, readingPitch(traditional, spelling.EFlat(), 5), "m10", altForReading(-1), false},
		{"zero unison", mode.MustFromCode("ionian"), traditional, readingPitch(traditional, spelling.C(), 4), mode.DegreePlace{1, 0}, mode.DegreePlace{1, 0}, 0, readingPitch(traditional, spelling.BSharp(), 3), readingPitch(traditional, spelling.C(), 4), "P1", 0, readingPitch(traditional, spelling.C(), 4), "P1", altForReading(0), true},
		{"altered unison changes direction", mode.MustFromCode("ionian"), traditional, readingPitch(traditional, spelling.C(), 4), mode.DegreePlace{1, 0}, mode.DegreePlace{1, 0}, -1, readingPitch(traditional, spelling.B(), 3), readingPitch(traditional, spelling.C(), 4), "P1", 0, readingPitch(traditional, spelling.CFlat(), 4), "-A1", altForReading(-1), false},
		{"pentatonic III to IV", mode.MustClosedFromIntervals(intervals("M2 M2 m3 M2 m3")), traditional, readingPitch(traditional, spelling.C(), 4), mode.DegreePlace{3, 0}, mode.DegreePlace{4, 0}, 3, readingPitch(traditional, spelling.G(), 4), readingPitch(traditional, spelling.G(), 4), "m3", 3, readingPitch(traditional, spelling.G(), 4), "m3", altForReading(0), true},
		{"quarter tone correction", mode.MustFromCode("ionian"), quarter, readingPitch(quarter, spelling.C(), 4), mode.DegreePlace{1, 0}, mode.DegreePlace{3, 0}, 7, readingPitch(quarter, spelling.MustNew(letter.E, alteration.MustFromFraction(-1, 2)), 4), readingPitch(quarter, spelling.E(), 4), "M3", 8, readingPitch(quarter, spelling.MustNew(letter.E, alteration.MustFromFraction(-1, 2)), 4), "n3", alteration.MustFromFraction(-1, 2), false},
		{"system written boundary", mode.MustFromCode("ionian"), boundaryA, readingPitch(boundaryA, spelling.G(), 4), mode.DegreePlace{1, 0}, mode.DegreePlace{2, 0}, 3, readingPitch(boundaryA, spelling.BFlat(), 5), readingPitch(boundaryA, spelling.A(), 5), "M2", 2, readingPitch(boundaryA, spelling.ASharp(), 5), "A2", altForReading(1), false},
		{"negative cycles", mode.MustFromCode("ionian"), traditional, readingPitch(traditional, spelling.C(), 4), mode.DegreePlace{1, -1}, mode.DegreePlace{3, -1}, 4, readingPitch(traditional, spelling.E(), 3), readingPitch(traditional, spelling.E(), 3), "M3", 4, readingPitch(traditional, spelling.E(), 3), "M3", altForReading(0), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			before := tc.m.Template()
			for _, policy := range []mode.RolePolicy{mode.StrictRole, mode.AllowTargetAlteration} {
				for _, operation := range []string{"shift", "pitch"} {
					t.Run(operation+policyName(policy), func(t *testing.T) {
						t.Parallel()
						var got mode.RoleResult
						var err error
						observed := interval.MustNewChromatic(tc.sys, tc.observed)
						if operation == "shift" {
							got, err = tc.m.ResolveRole(tc.tonic, tc.from, tc.to, observed, policy)
						} else {
							got, err = tc.m.ResolveRolePitch(tc.sys, tc.tonic, tc.from, tc.to, tc.observedPitch, policy)
						}
						require.NoError(t, err)
						assert.Equal(t, tc.from, got.From)
						assert.Equal(t, tc.to, got.To)
						assertRoleReading(t, got.Expected, tc.sys, tc.expected, tc.expectedText, tc.expectedSteps, altForReading(0))
						if tc.match {
							assert.Equal(t, mode.RoleMatch, got.Status)
						} else {
							assert.Equal(t, mode.RoleMismatch, got.Status)
						}
						candidate := tc.match || policy == mode.AllowTargetAlteration
						assert.Equal(t, candidate, got.HasCandidate)
						if candidate {
							assertRoleReading(t, got.Candidate, tc.sys, tc.candidate, tc.candidateText, tc.observed, tc.delta)
						} else {
							assert.Equal(t, mode.RoleReading{}, got.Candidate)
						}
						var must mode.RoleResult
						if operation == "shift" {
							must = tc.m.MustResolveRole(tc.tonic, tc.from, tc.to, observed, policy)
						} else {
							must = tc.m.MustResolveRolePitch(tc.sys, tc.tonic, tc.from, tc.to, tc.observedPitch, policy)
						}
						assert.Equal(t, got, must)
						assert.True(t, before.Equal(tc.m.Template()), "the mode remains unchanged")
					})
				}
			}
		})
	}
}
func policyName(p mode.RolePolicy) string {
	if p == mode.StrictRole {
		return "/strict"
	}
	return "/alteration"
}

func TestResolveRoleNeedContextAndErrors(t *testing.T) {
	t.Parallel()
	m := mode.MustFromCode("ionian")
	sys := system.Traditional()
	tonic := readingPitch(sys, spelling.C(), 4)
	observed := interval.MustNewChromatic(sys, 3)
	from, to := mode.DegreePlace{1, 0}, mode.DegreePlace{3, 0}
	for _, tc := range []struct {
		name     string
		m        mode.Mode
		tonic    pitch.Pitch
		from, to mode.DegreePlace
		c        interval.Chromatic
		sys      *system.System
		p        pitch.Pitch
	}{
		{"mode", mode.Mode{}, tonic, from, to, observed, sys, tonic},
		{"tonic", m, pitch.Pitch{}, from, to, observed, sys, tonic},
		{"observation", m, tonic, from, to, interval.Chromatic{}, sys, pitch.Pitch{}},
		{"first degree", m, tonic, mode.DegreePlace{0, 1}, to, observed, sys, tonic},
		{"second degree", m, tonic, from, mode.DegreePlace{0, -1}, observed, sys, tonic},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			shift, err := tc.m.ResolveRole(tc.tonic, tc.from, tc.to, tc.c, mode.StrictRole)
			require.NoError(t, err)
			assert.Equal(t, mode.RoleResult{Status: mode.NeedContext, From: tc.from, To: tc.to}, shift)
			byPitch, err := tc.m.ResolveRolePitch(tc.sys, tc.tonic, tc.from, tc.to, tc.p, mode.AllowTargetAlteration)
			require.NoError(t, err)
			assert.Equal(t, mode.RoleResult{Status: mode.NeedContext, From: tc.from, To: tc.to}, byPitch)
		})
	}
	got, err := m.ResolveRolePitch(nil, tonic, from, to, tonic, mode.StrictRole)
	require.NoError(t, err)
	assert.Equal(t, mode.RoleResult{From: from, To: to}, got)
	open := mode.MustOpenFromIntervals(intervals("M2 m2"))
	for _, tc := range []struct {
		name     string
		m        mode.Mode
		tonic    pitch.Pitch
		from, to mode.DegreePlace
		policy   mode.RolePolicy
	}{
		{"unset policy", m, tonic, from, to, 0},
		{"unknown policy before context", mode.Mode{}, pitch.Pitch{}, mode.DegreePlace{}, mode.DegreePlace{}, 99},
		{"negative first before missing tonic", m, pitch.Pitch{}, mode.DegreePlace{-1, 0}, to, mode.StrictRole},
		{"negative second", m, tonic, from, mode.DegreePlace{-1, 0}, mode.StrictRole},
		{"first above length", m, tonic, mode.DegreePlace{8, 0}, to, mode.StrictRole},
		{"second above length before context", m, pitch.Pitch{}, from, mode.DegreePlace{8, 0}, mode.StrictRole},
		{"open first cycle", open, tonic, mode.DegreePlace{1, 1}, to, mode.StrictRole},
		{"open second cycle", open, tonic, from, mode.DegreePlace{3, -1}, mode.StrictRole},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result, err := tc.m.ResolveRole(tc.tonic, tc.from, tc.to, observed, tc.policy)
			assert.ErrorIs(t, err, mode.ErrInvalidArgument)
			assert.Equal(t, mode.RoleResult{}, result)
			result, err = tc.m.ResolveRolePitch(sys, tc.tonic, tc.from, tc.to, tonic, tc.policy)
			assert.ErrorIs(t, err, mode.ErrInvalidArgument)
			assert.Equal(t, mode.RoleResult{}, result)
		})
	}
	half := readingPitch(system.QuarterTones24(), spelling.MustNew(letter.E, alteration.MustFromFraction(-1, 2)), 4)
	for _, p := range []pitch.Pitch{tonic, half} {
		badTonic, observedPitch := half, tonic
		if p.Equal(half) {
			badTonic, observedPitch = tonic, half
		}
		result, err := m.ResolveRolePitch(sys, badTonic, from, to, observedPitch, mode.StrictRole)
		assert.ErrorIs(t, err, system.ErrInadmissible)
		assert.Equal(t, mode.RoleResult{}, result)
	}
	result, err := m.ResolveRole(half, from, to, observed, mode.StrictRole)
	assert.ErrorIs(t, err, system.ErrInadmissible)
	assert.Equal(t, mode.RoleResult{}, result)
	result, err = m.ResolveRole(tonic, from, mode.DegreePlace{3, math.MaxInt}, observed, mode.StrictRole)
	assert.ErrorIs(t, err, mode.ErrOutOfRange)
	assert.Equal(t, mode.RoleResult{}, result)
	result, err = m.ResolveRolePitch(sys, tonic, from, mode.DegreePlace{3, math.MaxInt}, tonic, mode.StrictRole)
	assert.ErrorIs(t, err, mode.ErrOutOfRange)
	assert.Equal(t, mode.RoleResult{}, result)
	readingPanic(t, func() { m.MustResolveRole(tonic, from, to, observed, 0) }, mode.ErrInvalidArgument)
	readingPanic(t, func() { m.MustResolveRolePitch(sys, tonic, from, to, tonic, 0) }, mode.ErrInvalidArgument)
	readingPanic(t, func() { m.MustMatchSpellings(spelling.C(), spelling.C(), spelling.C(), mode.CycleWindow{2, 1}) }, mode.ErrInvalidArgument)
	readingPanic(t, func() { m.MustMatchPitches(sys, tonic, tonic, tonic, 0, mode.CycleWindow{}) }, mode.ErrInvalidArgument)
}

func TestResolveRoleNoCandidate(t *testing.T) {
	t.Parallel()
	m := mode.MustFromCode("ionian")
	sys := system.MustNew(system.Definition{Name: "natural only", Classes: 12, Boundary: letter.C, Offsets: [letter.Count]int{0, 2, 4, 5, 7, 9, 11}, Chromatic: 1, Alterations: system.Single(alteration.Natural())})
	from, to := mode.DegreePlace{1, 0}, mode.DegreePlace{3, 0}
	for _, steps := range []int{3, math.MinInt, math.MaxInt} {
		result, err := m.ResolveRole(readingPitch(sys, spelling.C(), 4), from, to, interval.MustNewChromatic(sys, steps), mode.AllowTargetAlteration)
		require.NoError(t, err)
		assert.Equal(t, mode.RoleMismatch, result.Status)
		assert.False(t, result.HasCandidate)
		assert.Equal(t, mode.RoleReading{}, result.Candidate)
		assertRoleReading(t, result.Expected, sys, readingPitch(sys, spelling.E(), 4), "M3", 4, altForReading(0))
	}
	// Observed pitch is admissible, but target spelling E-flat is not.
	result, err := m.ResolveRolePitch(sys, readingPitch(sys, spelling.C(), 4), from, to, readingPitch(sys, spelling.D(), 4), mode.AllowTargetAlteration)
	require.NoError(t, err)
	assert.Equal(t, mode.RoleMismatch, result.Status)
	assert.False(t, result.HasCandidate)
}

func TestRoleAndMembershipConcurrentRead(t *testing.T) {
	t.Parallel()
	m := mode.MustFromCode("ionian")
	sys := system.Traditional()
	tonic := readingPitch(sys, spelling.C(), 4)
	shift := interval.MustNewChromatic(sys, 3)
	results := make([]mode.RoleResult, 16)
	memberships := make([]mode.PitchMembership, 16)
	errs := make([]error, 32)
	var wg sync.WaitGroup
	for worker := range results {
		wg.Go(func() {
			results[worker], errs[worker] = m.ResolveRole(tonic, mode.DegreePlace{1, 0}, mode.DegreePlace{3, 0}, shift, mode.AllowTargetAlteration)
			memberships[worker], errs[worker+16] = m.MatchPitches(sys, tonic, tonic, tonic, mode.WrittenPitch, mode.CycleWindow{-2, 2})
		})
	}
	wg.Wait()
	for _, err := range errs {
		require.NoError(t, err)
	}
	for index, result := range results {
		assert.Equal(t, mode.RoleMismatch, result.Status)
		assert.True(t, result.HasCandidate)
		assertRoleReading(t, result.Candidate, sys, readingPitch(sys, spelling.EFlat(), 4), "m3", 3, altForReading(-1))
		assert.Equal(t, []mode.DegreePlace{{1, 0}}, memberships[index].First)
		assert.Equal(t, []mode.DegreePlace{{1, 0}}, memberships[index].Second)
	}
}
