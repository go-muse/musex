package score

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/track"
)

var (
	ErrUnset           = errors.New("score is unset")
	ErrInvalidArgument = errors.New("invalid score argument")
	ErrPart            = errors.New("unknown score part")
	ErrVoice           = errors.New("unknown score voice")
	ErrPosition        = errors.New("position outside score")
	ErrLocalClock      = errors.New("part has a local clock")
	ErrGridShift       = errors.New("replacement moves existing bar starts")
	ErrMapAddress      = errors.New("score map mark has no valid bar binding")
	ErrMapConflict     = errors.New("score map mark has conflicting positions")
	ErrPeriod          = errors.New("barline period is not defined for these grids")
	ErrOutOfRange      = errors.New("score result is out of range")
)

// Issue locates one failure; optional fields are unset when unavailable.
type Issue struct {
	Cause              error
	Part               PartID
	Place              track.Place
	Position, Proposed duration.Position
	Mark               *MapMark
	Bindings           []Binding
	Detail             *track.Issue
}

// Problems preserves independent failures in deterministic order.
type Problems struct{ Issues []Issue }

func (p *Problems) Error() string {
	if p == nil || len(p.Issues) == 0 {
		return ""
	}
	var out []string
	for _, x := range p.Issues {
		if x.Cause != nil {
			out = append(out, x.Cause.Error())
		}
	}
	return fmt.Sprintf("score: %d issues: %s", len(p.Issues), strings.Join(out, "; "))
}
func (p *Problems) Unwrap() []error {
	if p == nil {
		return nil
	}
	var out []error
	for _, x := range p.Issues {
		if x.Cause != nil {
			out = append(out, x.Cause)
		}
	}
	return out
}
func classify(err error) error {
	if errors.Is(err, fraction.ErrOutOfRange) {
		return errors.Join(ErrOutOfRange, err)
	}
	return err
}
func trackIssues(id PartID, err error) []Issue {
	if p, ok := errors.AsType[*track.Problems](err); ok {
		out := make([]Issue, len(p.Issues))
		for i, x := range p.Issues {
			c := x
			if c.AfterBar != nil {
				c.AfterBar = new(*c.AfterBar)
			}
			out[i] = Issue{Cause: classify(x.Cause), Part: id, Place: x.Place, Position: x.Position, Detail: &c}
		}
		return out
	}
	return []Issue{{Cause: classify(err), Part: id}}
}
func must[T any](x T, err error) T {
	if err != nil {
		panic(err)
	}
	return x
}
func must2[A, B any](a A, b B, err error) (A, B) {
	if err != nil {
		panic(err)
	}
	return a, b
}

// Binding records a mark's address and candidate position for one participation.
type Binding struct {
	Part     PartID
	Address  measure.Address
	AfterBar *int
	Position duration.Position
}
