package tempomap

import (
	"math/big"
	"sort"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tempo"
)

func ratOfFraction(f fraction.Fraction) *big.Rat {
	r, _ := f.BigRat()
	return r
}

func ratOfPosition(p duration.Position) *big.Rat {
	return ratOfFraction(p.MustFraction())
}

func ratOfSeconds(s tempo.Seconds) *big.Rat {
	return ratOfFraction(s.MustFraction())
}

func positionFromRat(op string, value *big.Rat) (duration.Position, error) {
	f, err := fraction.FromBigRat(value)
	if err != nil {
		return duration.Position{}, rangeError(op)
	}
	p, err := duration.PositionFromFraction(f)
	if err != nil {
		return duration.Position{}, dependencyArgumentError(op, err)
	}
	return p, nil
}

func secondsFromRat(op string, value *big.Rat) (tempo.Seconds, error) {
	f, err := fraction.FromBigRat(value)
	if err != nil {
		return tempo.Seconds{}, rangeError(op)
	}
	s, err := tempo.SecondsFromFraction(f)
	if err != nil {
		return tempo.Seconds{}, dependencyArgumentError(op, err)
	}
	return s, nil
}

func constantSegmentForm(start, end duration.Position, speed *big.Rat) logForm {
	length := new(big.Rat).Sub(ratOfPosition(end), ratOfPosition(start))
	seconds := new(big.Rat).Mul(big.NewRat(60, 1), length)
	seconds.Quo(seconds, speed)
	return rationalLogForm(seconds)
}

func gradualSegmentForm(start, end duration.Position, startSpeed, endSpeed *big.Rat) logForm {
	if startSpeed.Cmp(endSpeed) == 0 {
		return constantSegmentForm(start, end, startSpeed)
	}
	length := new(big.Rat).Sub(ratOfPosition(end), ratOfPosition(start))
	coefficient := new(big.Rat).Mul(big.NewRat(60, 1), length)
	delta := new(big.Rat).Sub(endSpeed, startSpeed)
	coefficient.Quo(coefficient, delta)
	base := new(big.Rat).Quo(endSpeed, startSpeed)
	return logForm{constant: new(big.Rat), terms: []logTerm{{coefficient: coefficient, base: base}}}
}

type timeEvent struct {
	position     duration.Position
	start        logForm
	end          logForm
	hasExtension bool
}

func prepareTimeline(idx *mapIndex) {
	prefix := zeroLogForm()
	approximate := false
	// The segment forms do not keep the speeds, so one pair of buffers serves
	// every segment.
	var startBuf, endBuf big.Rat
	for i := range idx.changes {
		idx.changes[i].startForm = cloneLogForm(prefix)
		idx.changes[i].approximateBefore = approximate
		if i+1 >= len(idx.changes) {
			continue
		}
		start, end := idx.changes[i].position, idx.changes[i+1].position
		if end.Compare(start) <= 0 {
			continue
		}
		startSpeed, _ := idx.changes[i].startSpeed.BigRatTo(&startBuf)
		endSpeed, _ := idx.changes[i+1].startSpeed.BigRatTo(&endBuf)
		var elapsed logForm
		if idx.changes[i].change.kind == ChangeGradual {
			elapsed = gradualSegmentForm(start, end, startSpeed, endSpeed)
			if startSpeed.Cmp(endSpeed) != 0 {
				approximate = true
			}
		} else {
			elapsed = constantSegmentForm(start, end, startSpeed)
		}
		prefix = logFormAdd(prefix, elapsed)
	}
	extensionSum := new(big.Rat)
	var secondsBuf big.Rat
	for i := range idx.extensions {
		idx.extensions[i].before = new(big.Rat).Set(extensionSum)
		seconds, _ := idx.extensions[i].extension.seconds.MustFraction().BigRatTo(&secondsBuf)
		extensionSum.Add(extensionSum, seconds)
	}
	buildTimelineEvents(idx)
}

func changeAtPosition(idx *mapIndex, position duration.Position) int {
	i := sort.Search(len(idx.changes), func(i int) bool {
		return idx.changes[i].position.Compare(position) > 0
	})
	return i - 1
}

func unfinishedAt(idx *mapIndex, position duration.Position) bool {
	last := idx.changes[len(idx.changes)-1]
	return last.change.kind == ChangeGradual && position.Compare(last.position) >= 0
}

func speedRatAtIndex(idx *mapIndex, position *big.Rat) *big.Rat {
	var buf big.Rat
	i := sort.Search(len(idx.changes), func(i int) bool {
		p, _ := idx.changes[i].position.MustFraction().BigRatTo(&buf)
		return p.Cmp(position) > 0
	}) - 1
	current := idx.changes[i]
	if current.change.kind != ChangeGradual {
		return ratOfFraction(current.startSpeed)
	}
	next := idx.changes[i+1]
	length := new(big.Rat).Sub(ratOfPosition(next.position), ratOfPosition(current.position))
	offset := new(big.Rat).Sub(position, ratOfPosition(current.position))
	delta := new(big.Rat).Sub(ratOfFraction(next.startSpeed), ratOfFraction(current.startSpeed))
	speed := new(big.Rat).Mul(delta, offset)
	speed.Quo(speed, length)
	return speed.Add(speed, ratOfFraction(current.startSpeed))
}

func speedAtIndex(idx *mapIndex, position duration.Position) (fraction.Fraction, error) {
	i := changeAtPosition(idx, position)
	if idx.changes[i].change.kind == ChangeGradual && i+1 >= len(idx.changes) {
		return fraction.Fraction{}, simpleError(ErrUnfinished, "tempomap.speedAtIndex", "unfinished gradual segment")
	}
	result, err := fraction.FromBigRat(speedRatAtIndex(idx, ratOfPosition(position)))
	if err != nil {
		return fraction.Fraction{}, rangeError("tempomap.speedAtIndex")
	}
	return result, nil
}

func segmentStartForm(idx *mapIndex, i int, position duration.Position) (logForm, bool) {
	current := idx.changes[i]
	form := cloneLogForm(current.startForm)
	approximate := current.approximateBefore
	if position.Compare(current.position) == 0 {
		return form, approximate
	}
	end := position
	endSpeed := ratOfFraction(current.startSpeed)
	if i+1 < len(idx.changes) {
		next := idx.changes[i+1]
		if position.Compare(next.position) == 0 {
			endSpeed = ratOfFraction(next.startSpeed)
		} else {
			endSpeed = speedRatAtIndex(idx, ratOfPosition(position))
		}
	}
	var elapsed logForm
	if current.change.kind == ChangeGradual {
		elapsed = gradualSegmentForm(current.position, end, ratOfFraction(current.startSpeed), endSpeed)
		if endSpeed.Cmp(ratOfFraction(current.startSpeed)) != 0 {
			approximate = true
		}
	} else {
		elapsed = constantSegmentForm(current.position, end, ratOfFraction(current.startSpeed))
	}
	return logFormAdd(form, elapsed), approximate
}

func extensionSumAt(idx *mapIndex, position duration.Position, include bool) *big.Rat {
	i := sort.Search(len(idx.extensions), func(i int) bool {
		cmp := idx.extensions[i].position.Compare(position)
		if include {
			return cmp > 0
		}
		return cmp >= 0
	})
	if i == 0 {
		return new(big.Rat)
	}
	previous := idx.extensions[i-1]
	return new(big.Rat).Add(previous.before, ratOfSeconds(previous.extension.seconds))
}

func startFormAt(idx *mapIndex, position duration.Position, end bool) (logForm, bool) {
	i := changeAtPosition(idx, position)
	form, approximate := segmentStartForm(idx, i, position)
	form = logFormAddRat(form, extensionSumAt(idx, position, end))
	return form, approximate
}

func buildTimelineEvents(idx *mapIndex) {
	count := len(idx.changes) + len(idx.extensions)
	idx.events = make([]timeEvent, 0, count)
	ci, ei := 0, 0
	for ci < len(idx.changes) || ei < len(idx.extensions) {
		var position duration.Position
		switch {
		case ei >= len(idx.extensions):
			position = idx.changes[ci].position
		case ci >= len(idx.changes):
			position = idx.extensions[ei].position
		case idx.changes[ci].position.Compare(idx.extensions[ei].position) <= 0:
			position = idx.changes[ci].position
		default:
			position = idx.extensions[ei].position
		}
		for ci < len(idx.changes) && idx.changes[ci].position.Compare(position) == 0 {
			ci++
		}
		hasExtension := false
		if ei < len(idx.extensions) && idx.extensions[ei].position.Compare(position) == 0 {
			hasExtension = true
			ei++
		}
		start, _ := startFormAt(idx, position, false)
		end := start
		if hasExtension {
			seconds, _ := extensionAt(idx, position)
			end = logFormAddRat(start, ratOfSeconds(seconds))
		}
		idx.events = append(idx.events, timeEvent{position: position, start: start, end: end, hasExtension: hasExtension})
	}
}

func extensionAt(idx *mapIndex, position duration.Position) (tempo.Seconds, bool) {
	i := sort.Search(len(idx.extensions), func(i int) bool {
		return idx.extensions[i].position.Compare(position) >= 0
	})
	if i == len(idx.extensions) || idx.extensions[i].position.Compare(position) != 0 {
		return tempo.Seconds{}, false
	}
	return idx.extensions[i].extension.seconds, true
}
