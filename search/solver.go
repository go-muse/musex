package search

import (
	"math/big"
	"slices"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

// degreeData contains exact interval alterations and their positional offsets
// from one possible letter of the first degree.
type degreeData struct {
	deltas  []big.Rat
	offsets []int
}

// familySolver solves the finite alteration constraints of one tonic family.
// All big.Rat values reached through it are treated as immutable.
type familySolver struct {
	sys        *system.System
	class      int
	letters    []letter.Letter
	data       map[letter.Letter]degreeData
	inputNames map[spelling.Spelling]struct{}
}

func makeDegreeData(sys *system.System, intervals []interval.Interval, first letter.Letter) (degreeData, bool) {
	data := degreeData{
		deltas:  make([]big.Rat, len(intervals)),
		offsets: make([]int, len(intervals)),
	}
	chromatic := int64(sys.Chromatic())
	modulus := sys.Classes()

	for i, current := range intervals {
		r := (current.Size() - 1) % letter.Count
		delta := new(big.Rat).Sub(ratOfAlteration(current.Deviation()), big.NewRat(naturalDeviation(first, r), 1))
		scaled := new(big.Rat).Mul(delta, big.NewRat(chromatic, 1))
		if !scaled.IsInt() {
			return degreeData{}, false
		}

		steps := int64(0)
		for j := range r {
			step, err := sys.StepAfter(first.Shift(j))
			if err != nil {
				return degreeData{}, false
			}
			steps += int64(step)
		}

		data.deltas[i] = *delta
		data.offsets[i] = int(floorMod(steps+scaled.Num().Int64(), int64(modulus)))
	}

	return data, true
}

var (
	naturalOffsets = [letter.Count]int64{0, 2, 4, 5, 7, 9, 11}
	basicSemitones = [letter.Count]int64{0, 2, 4, 5, 7, 9, 11}
)

func naturalDeviation(first letter.Letter, r int) int64 {
	steps := (naturalOffsets[(first.Ordinal()+r)%letter.Count] - naturalOffsets[first.Ordinal()] + 12) % 12
	return steps - basicSemitones[r]
}

func ratOfAlteration(value alteration.Alteration) *big.Rat {
	num, den := value.Fraction()
	return new(big.Rat).SetFrac(big.NewInt(int64(num)), big.NewInt(int64(den)))
}

func alterationOfRat(value *big.Rat) (alteration.Alteration, bool) {
	num, den := value.Num(), value.Denom()
	if !num.IsInt64() || !den.IsInt64() {
		return alteration.Alteration{}, false
	}

	n, d := num.Int64(), den.Int64()
	if int64(int(n)) != n || int64(int(d)) != d {
		return alteration.Alteration{}, false
	}

	result, err := alteration.FromFraction(int(n), int(d))
	return result, err == nil
}

func streamValue(base, step *big.Rat, t int64) *big.Rat {
	return new(big.Rat).Add(base, new(big.Rat).Mul(step, big.NewRat(t, 1)))
}

// solutionStream is an ordered arithmetic stream of compatible alterations.
// Its values are base + step*(first + stride*n), within first..last.
type solutionStream struct {
	letter      letter.Letter
	base, step  *big.Rat
	first, last int64
	stride      int64
	single      bool
}

type valueConstraint struct {
	lo, hi      int64
	progression system.Progression
}

func (s familySolver) streams(limit *big.Rat) []solutionStream {
	progressions := s.sys.Alterations().Progressions()
	result := make([]solutionStream, 0)

	for _, currentLetter := range s.letters {
		data, exists := s.data[currentLetter]
		if !exists {
			continue
		}

		for _, progression := range progressions {
			result = append(result, s.streamsForProgression(currentLetter, data, progression, limit)...)
		}
	}

	return result
}

func (s familySolver) streamsForProgression(currentLetter letter.Letter, data degreeData,
	progression system.Progression, limit *big.Rat,
) []solutionStream {
	if progression.Den() == 0 {
		value := ratOfAlteration(progression.First())
		if !s.singleMatches(currentLetter, value) || !withinLimit(value, limit) || !s.compatible(value, data) {
			return nil
		}

		return []solutionStream{{
			letter: currentLetter,
			base:   value,
			step:   new(big.Rat),
			stride: 1,
			single: true,
		}}
	}

	d := int64(progression.Den())
	first := ratOfAlteration(progression.First())
	last := ratOfAlteration(progression.Last())
	chromatic := int64(s.sys.Chromatic())
	stepShift := chromatic / d
	firstShift := new(big.Rat).Mul(first, big.NewRat(chromatic, 1))
	if !firstShift.IsInt() {
		return nil
	}

	offset, err := s.sys.Offset(currentLetter)
	if err != nil {
		return nil
	}

	right := int64(s.class-offset) - firstShift.Num().Int64()
	n0, period, ok := solveCongruence(stepShift, right, int64(s.sys.Classes()))
	if !ok {
		return nil
	}

	base := new(big.Rat).Add(first, big.NewRat(n0, d))
	step := big.NewRat(period, d)
	nMax := new(big.Rat).Mul(new(big.Rat).Sub(last, first), big.NewRat(d, 1))
	if !nMax.IsInt() {
		return nil
	}

	lo := ceilDiv(-n0, period)
	hi := floorDiv(nMax.Num().Int64()-n0, period)
	if limit != nil {
		lower := new(big.Rat).Sub(new(big.Rat).Neg(limit), base)
		lower.Quo(lower, step)
		upper := new(big.Rat).Sub(limit, base)
		upper.Quo(upper, step)
		lo = max(lo, ceilRat(lower))
		hi = min(hi, floorRat(upper))
	}

	if lo > hi {
		return nil
	}

	// With only the first degree, the source progression already proves the
	// sole membership condition a ∈ A. Avoid preparing periodic residues.
	if len(data.deltas) == 1 {
		return []solutionStream{{
			letter: currentLetter,
			base:   base,
			step:   step,
			first:  lo,
			last:   hi,
			stride: 1,
		}}
	}

	return s.compatibleStreams(currentLetter, data, base, step, d, lo, hi)
}

func (s familySolver) singleMatches(currentLetter letter.Letter, value *big.Rat) bool {
	offset, err := s.sys.Offset(currentLetter)
	if err != nil {
		return false
	}

	shift := new(big.Rat).Mul(value, big.NewRat(int64(s.sys.Chromatic()), 1))
	if !shift.IsInt() {
		return false
	}

	return floorMod(int64(offset)+shift.Num().Int64(), int64(s.sys.Classes())) == int64(s.class)
}

func withinLimit(value, limit *big.Rat) bool {
	return limit == nil || new(big.Rat).Abs(value).Cmp(limit) <= 0
}

func (s familySolver) compatible(value *big.Rat, data degreeData) bool {
	rule := s.sys.Alterations()
	for _, delta := range data.deltas {
		altered, ok := alterationOfRat(new(big.Rat).Add(value, &delta))
		if !ok {
			return false
		}

		contains, err := rule.Contains(altered)
		if err != nil || !contains {
			return false
		}
	}

	return true
}

func (s familySolver) compatibleStreams(currentLetter letter.Letter, data degreeData,
	base, step *big.Rat, d, lo, hi int64,
) []solutionStream {
	constraints := make([][]valueConstraint, len(data.deltas))
	events := []int64{lo, hi + 1}
	progressions := s.sys.Alterations().Progressions()

	for j, delta := range data.deltas {
		for _, progression := range progressions {
			lower := shiftedBound(progression.First(), base, &delta, step)
			upper := shiftedBound(progression.Last(), base, &delta, step)
			constraint := valueConstraint{
				lo:          ceilRat(lower),
				hi:          floorRat(upper),
				progression: progression,
			}
			constraints[j] = append(constraints[j], constraint)
			events = append(events, constraint.lo, constraint.hi+1)
		}
	}

	slices.Sort(events)
	events = slices.Compact(events)
	result := make([]solutionStream, 0)

	for i := range len(events) - 1 {
		start, end := max(lo, events[i]), min(hi, events[i+1]-1)
		if start > end {
			continue
		}

		active := make([][]system.Progression, len(constraints))
		possible := true
		for j, degreeConstraints := range constraints {
			for _, constraint := range degreeConstraints {
				if constraint.lo <= start && end <= constraint.hi {
					active[j] = append(active[j], constraint.progression)
				}
			}
			if len(active[j]) == 0 {
				possible = false
				break
			}
		}
		if !possible {
			continue
		}

		lastResidue := min(end, start+d-1)
		for t := start; t <= lastResidue; t++ {
			if compatibleResidue(base, step, data.deltas, active, t) {
				result = append(result, solutionStream{
					letter: currentLetter,
					base:   base,
					step:   step,
					first:  t,
					last:   end,
					stride: d,
				})
			}
		}
	}

	return result
}

func shiftedBound(bound alteration.Alteration, base, delta, step *big.Rat) *big.Rat {
	value := new(big.Rat).Sub(ratOfAlteration(bound), base)
	value.Sub(value, delta)
	return value.Quo(value, step)
}

func compatibleResidue(base, step *big.Rat, deltas []big.Rat, active [][]system.Progression, t int64) bool {
	value := streamValue(base, step, t)
	for j, delta := range deltas {
		altered, ok := alterationOfRat(new(big.Rat).Add(value, &delta))
		if !ok {
			return false
		}

		accepted := false
		for _, progression := range active[j] {
			if progression.Contains(altered) {
				accepted = true
				break
			}
		}
		if !accepted {
			return false
		}
	}

	return true
}

// representative returns a winner after processing every cursor at the first
// reachable magnitude of alteration. It never enumerates the whole family.
func (s familySolver) representative() (spelling.Spelling, bool) {
	cursors := newSolutionCursors(s.streams(nil))
	var (
		winner    spelling.Spelling
		winnerKey int
		found     bool
	)

	for {
		var best *big.Rat
		for i := range cursors {
			if cursors[i].done {
				continue
			}

			magnitude := new(big.Rat).Abs(cursors[i].value())
			if best == nil || magnitude.Cmp(best) < 0 {
				best = magnitude
			}
		}
		if best == nil {
			return spelling.Spelling{}, false
		}

		for i := range cursors {
			cursor := &cursors[i]
			if cursor.done || new(big.Rat).Abs(cursor.value()).Cmp(best) != 0 {
				continue
			}

			name, ok := cursor.name()
			if ok {
				score := 0
				if _, exists := s.inputNames[name]; exists {
					score = 1
				}

				if !found || score > winnerKey || score == winnerKey && name.Compare(winner) < 0 {
					winner, winnerKey, found = name, score, true
				}
			}
			cursor.advance()
		}

		if found {
			return winner, true
		}
	}
}

type solutionCursor struct {
	stream solutionStream
	t      int64
	up     bool
	done   bool
}

func newSolutionCursors(streams []solutionStream) []solutionCursor {
	result := make([]solutionCursor, 0, 2*len(streams))
	for _, stream := range streams {
		if stream.single {
			cursor := solutionCursor{stream: stream, up: stream.base.Sign() >= 0}
			result = append(result, cursor)
			continue
		}

		zero := new(big.Rat).Quo(new(big.Rat).Neg(stream.base), stream.step)
		up := alignUp(ceilRat(zero), stream.first, stream.stride)
		if up <= stream.last {
			result = append(result, solutionCursor{stream: stream, t: up, up: true})
		}

		down := alignDown(floorRat(zero), stream.first, stream.stride)
		if down >= stream.first {
			result = append(result, solutionCursor{stream: stream, t: down})
		}
	}

	return result
}

func (c solutionCursor) value() *big.Rat {
	return streamValue(c.stream.base, c.stream.step, c.t)
}

func (c solutionCursor) name() (spelling.Spelling, bool) {
	value, ok := alterationOfRat(c.value())
	if !ok {
		return spelling.Spelling{}, false
	}

	name, err := spelling.New(c.stream.letter, value)
	return name, err == nil
}

func (c *solutionCursor) advance() {
	if c.stream.single {
		c.done = true
		return
	}

	if c.up {
		if c.t > c.stream.last-c.stream.stride {
			c.done = true
		} else {
			c.t += c.stream.stride
		}
		return
	}

	if c.t < c.stream.first+c.stream.stride {
		c.done = true
	} else {
		c.t -= c.stream.stride
	}
}

func (s familySolver) spellings(limit *big.Rat) func(func(spelling.Spelling) bool) {
	streams := s.streams(limit)
	return func(yield func(spelling.Spelling) bool) {
		for _, currentLetter := range s.letters {
			cursors := make([]solutionCursor, 0)
			for _, stream := range streams {
				if stream.letter == currentLetter {
					cursors = append(cursors, solutionCursor{stream: stream, t: stream.first, up: true})
				}
			}

			var (
				last    spelling.Spelling
				started bool
			)
			for {
				best := -1
				var value *big.Rat
				for i := range cursors {
					if cursors[i].done {
						continue
					}
					current := cursors[i].value()
					if best < 0 || current.Cmp(value) < 0 {
						best, value = i, current
					}
				}
				if best < 0 {
					break
				}

				name, ok := cursors[best].name()
				cursors[best].advance()
				if !ok || started && name == last {
					continue
				}
				if !yield(name) {
					return
				}
				last, started = name, true
			}
		}
	}
}

func alignUp(value, first, stride int64) int64 {
	return first + ceilDiv(value-first, stride)*stride
}

func alignDown(value, first, stride int64) int64 {
	return first + floorDiv(value-first, stride)*stride
}

func floorDiv(a, b int64) int64 {
	q := a / b
	if a%b != 0 && a < 0 {
		q--
	}
	return q
}

func ceilDiv(a, b int64) int64 {
	q := a / b
	if a%b != 0 && a > 0 {
		q++
	}
	return q
}

func floorRat(value *big.Rat) int64 {
	var quotient, remainder big.Int
	quotient.QuoRem(value.Num(), value.Denom(), &remainder)
	if value.Sign() < 0 && remainder.Sign() != 0 {
		quotient.Sub(&quotient, big.NewInt(1))
	}
	return quotient.Int64()
}

func ceilRat(value *big.Rat) int64 {
	var quotient, remainder big.Int
	quotient.QuoRem(value.Num(), value.Denom(), &remainder)
	if value.Sign() > 0 && remainder.Sign() != 0 {
		quotient.Add(&quotient, big.NewInt(1))
	}
	return quotient.Int64()
}

func floorMod(a, b int64) int64 {
	r := a % b
	if r < 0 {
		r += b
	}
	return r
}

func gcd(a, b int64) int64 {
	if a < 0 {
		a = -a
	}
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func solveCongruence(a, b, modulus int64) (start, period int64, ok bool) {
	divisor := gcd(a, modulus)
	b = floorMod(b, modulus)
	if b%divisor != 0 {
		return 0, 0, false
	}

	period = modulus / divisor
	if period == 1 {
		return 0, period, true
	}

	inverse := modInverse(floorMod(a/divisor, period), period)
	start = floorMod(floorMod(b/divisor, period)*inverse, period)
	return start, period, true
}

func modInverse(a, modulus int64) int64 {
	oldR, r := a, modulus
	oldS, s := int64(1), int64(0)
	for r != 0 {
		q := oldR / r
		oldR, r = r, oldR-q*r
		oldS, s = s, oldS-q*s
	}
	return floorMod(oldS, modulus)
}
