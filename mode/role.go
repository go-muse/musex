package mode

import (
	"errors"
	"math/big"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/system"
)

// RolePolicy selects strict comparison or an additional altered-target reading.
type RolePolicy uint8

const (
	// StrictRole confirms only the expected full shift.
	StrictRole RolePolicy = iota + 1
	// AllowTargetAlteration may additionally explain a mismatch by changing
	// only the expected target's alteration in its written register.
	AllowTargetAlteration
)

// RoleStatus distinguishes missing context, confirmation, and mismatch.
type RoleStatus uint8

const (
	NeedContext RoleStatus = iota
	RoleMatch
	RoleMismatch
)

// RoleReading records an absolute endpoint and its full relation to the source
// degree. Alteration is the change from the expected endpoint, not its sign.
// Values are safe to copy and read concurrently.
type RoleReading struct {
	Pitch      pitch.Pitch
	Interval   interval.Interval
	Shift      interval.Chromatic
	Alteration alteration.Alteration
}

// RoleResult retains the requested places. Candidate is meaningful only when
// HasCandidate is true; an altered candidate still leaves RoleMismatch.
type RoleResult struct {
	Status              RoleStatus
	From, To            DegreePlace
	Expected, Candidate RoleReading
	HasCandidate        bool
}

func (m Mode) checkRole(from, to DegreePlace, policy RolePolicy) error {
	if policy != StrictRole && policy != AllowTargetAlteration {
		return ErrInvalidArgument
	}
	for _, p := range []DegreePlace{from, to} {
		if p.Number < 0 {
			return ErrInvalidArgument
		}
		if m.IsValid() && (p.Number > m.Len() || (m.IsOpen() && p.Cycle != 0)) {
			return ErrInvalidArgument
		}
	}
	return nil
}
func roleError(err error) error {
	if errors.Is(err, system.ErrOutOfRange) {
		return errors.Join(ErrOutOfRange, interval.ErrOutOfRange, err)
	}
	return queryError(err)
}
func makeReading(sys *system.System, source, target pitch.Pitch) (RoleReading, error) {
	i, err := interval.BetweenPitches(sys, source, target)
	if err != nil {
		return RoleReading{}, roleError(err)
	}
	s, err := sys.Distance(source, target)
	if err != nil {
		return RoleReading{}, roleError(err)
	}
	return RoleReading{Pitch: target, Interval: i, Shift: interval.MustNewChromatic(sys, int(s))}, nil
}
func (m Mode) roleExpected(sys *system.System, tonic pitch.Pitch, from, to DegreePlace) (pitch.Pitch, RoleReading, error) {
	source, err := m.placePitch(sys, tonic, from)
	if err != nil {
		return pitch.Pitch{}, RoleReading{}, roleError(err)
	}
	target, err := m.placePitch(sys, tonic, to)
	if err != nil {
		return pitch.Pitch{}, RoleReading{}, roleError(err)
	}
	expected, err := makeReading(sys, source, target)
	return source, expected, err
}
func roleConclusion(sys *system.System, source pitch.Pitch, result RoleResult, observed interval.Chromatic, policy RolePolicy) RoleResult {
	if observed.Steps() == result.Expected.Shift.Steps() {
		result.Status = RoleMatch
		result.HasCandidate = true
		result.Candidate = result.Expected
		return result
	}
	result.Status = RoleMismatch
	if policy == StrictRole {
		return result
	}
	difference := new(big.Int).Sub(big.NewInt(int64(observed.Steps())), big.NewInt(int64(result.Expected.Shift.Steps())))
	change := new(big.Rat).SetFrac(difference, big.NewInt(int64(sys.Chromatic())))
	a, err := alterationOf(change)
	if err != nil {
		return result
	}
	target, err := sys.RaiseWrittenByAlteration(result.Expected.Pitch, a)
	if err != nil {
		return result
	}
	candidate, err := makeReading(sys, source, target)
	if err != nil {
		return result
	}
	candidate.Alteration = a
	candidate.Shift = observed
	result.Candidate = candidate
	result.HasCandidate = true
	return result
}

// ResolveRole compares an observed signed shift with two degree places.
// Missing context returns NeedContext, preserving places, without an error.
// Explicit invalid policy, degree, or open-mode cycle is checked first.
// Expected is always supplied for a complete query. Only exact agreement is
// RoleMatch; an optional target alteration explains but never confirms a role.
func (m Mode) ResolveRole(tonic pitch.Pitch, from, to DegreePlace, observed interval.Chromatic, policy RolePolicy) (RoleResult, error) {
	if err := m.checkRole(from, to, policy); err != nil {
		return RoleResult{}, err
	}
	result := RoleResult{From: from, To: to}
	if !m.IsValid() || !tonic.IsValid() || !observed.IsValid() || from.Number == 0 || to.Number == 0 {
		return result, nil
	}
	sys := observed.System()
	source, expected, err := m.roleExpected(sys, tonic, from, to)
	if err != nil {
		return RoleResult{}, err
	}
	result.Expected = expected
	return roleConclusion(sys, source, result, observed, policy), nil
}

// MustResolveRole panics with the error returned by ResolveRole.
func (m Mode) MustResolveRole(tonic pitch.Pitch, from, to DegreePlace, observed interval.Chromatic, policy RolePolicy) RoleResult {
	return must(m.ResolveRole(tonic, from, to, observed, policy))
}

// ResolveRolePitch measures the observed registered pitch in sys and applies
// ResolveRole's conclusion rules. Its written spelling supplies the measured
// position; the candidate keeps the expected target's letter and register.
func (m Mode) ResolveRolePitch(sys *system.System, tonic pitch.Pitch, from, to DegreePlace, observed pitch.Pitch, policy RolePolicy) (RoleResult, error) {
	if err := m.checkRole(from, to, policy); err != nil {
		return RoleResult{}, err
	}
	result := RoleResult{From: from, To: to}
	if !m.IsValid() || !sys.IsValid() || !tonic.IsValid() || !observed.IsValid() || from.Number == 0 || to.Number == 0 {
		return result, nil
	}
	source, expected, err := m.roleExpected(sys, tonic, from, to)
	if err != nil {
		return RoleResult{}, err
	}
	steps, err := sys.Distance(source, observed)
	if err != nil {
		return RoleResult{}, roleError(err)
	}
	result.Expected = expected
	return roleConclusion(sys, source, result, interval.MustNewChromatic(sys, int(steps)), policy), nil
}

// MustResolveRolePitch panics with the error returned by ResolveRolePitch.
func (m Mode) MustResolveRolePitch(sys *system.System, tonic pitch.Pitch, from, to DegreePlace, observed pitch.Pitch, policy RolePolicy) RoleResult {
	return must(m.ResolveRolePitch(sys, tonic, from, to, observed, policy))
}
