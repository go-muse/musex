package voiceleading

import (
	"slices"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/system"
)

// Equal compares part names and order, nominal order, the presence and value
// of the measured system, and every pairwise vertical and transition. Sources
// are not read or compared. Two unset analyses, including nil, are equal.
// Two absent optional intervals are equal within their containing result.
func (a *Analysis) Equal(other *Analysis) bool {
	if !a.IsSet() || !other.IsSet() {
		return a.IsSet() == other.IsSet()
	}
	if a.order != other.order || !slices.Equal(a.names, other.names) || (a.sys == nil) != (other.sys == nil) {
		return false
	}
	if a.sys != nil && !a.sys.Equal(other.sys) {
		return false
	}
	return slices.EqualFunc(a.data, other.data, func(x, y pairData) bool {
		return x.pair == y.pair && slices.EqualFunc(x.verticals, y.verticals, equalVertical) && slices.EqualFunc(x.transitions, y.transitions, equalTransition)
	})
}

func equalInterval(x, y interval.Interval) bool {
	if !x.IsValid() || !y.IsValid() {
		return x.IsValid() == y.IsValid()
	}
	return x.Equal(y)
}

func equalElement(x, y ElementAt) bool {
	return x.Index == y.Index && x.Element.Equal(y.Element) && x.Start.Equal(y.Start) && x.End.Equal(y.End)
}

func equalVertical(x, y Vertical) bool {
	return x.Pair == y.Pair && x.Start.Equal(y.Start) && x.End.Equal(y.End) && equalElement(x.First, y.First) && equalElement(x.Second, y.Second) && x.Sounding == y.Sounding && equalInterval(x.Specific, y.Specific) && x.Chromatic == y.Chromatic
}

func equalTransition(x, y Transition) bool {
	return equalVertical(x.From, y.From) && equalVertical(x.To, y.To) && equalInterval(x.First, y.First) && equalInterval(x.Second, y.Second) && x.FirstChromatic == y.FirstChromatic && x.SecondChromatic == y.SecondChromatic && x.Motion == y.Motion
}

// equalSystem compares the presence and value of two borrowed systems.
func equalSystem(x, y *system.System) bool {
	if x == nil || y == nil {
		return x == nil && y == nil
	}
	return x.Equal(y)
}

func equalMove(x, y Move) bool {
	return equalElement(x.From, y.From) && equalElement(x.To, y.To) && equalInterval(x.Interval, y.Interval) &&
		x.Chromatic == y.Chromatic && x.ThroughRest == y.ThroughRest
}
