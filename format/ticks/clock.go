package ticks

import (
	"math/big"
	"strconv"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
)

// TicksPerClock returns D/24, the exact ticks per MIDI clock; IsInteger of the
// result tells whether clocks fall on whole ticks.
func (r Resolution) TicksPerClock() (fraction.Fraction, error) {
	const op = "Resolution.TicksPerClock"
	if err := r.check(op); err != nil {
		return fraction.Fraction{}, err
	}
	return fractionOf(op, big.NewRat(int64(r.d), ClocksPerQuarter))
}

// MustTicksPerClock is like TicksPerClock but panics with its error.
func (r Resolution) MustTicksPerClock() fraction.Fraction { return must(r.TicksPerClock()) }

// TicksPerBeat returns D/4, the exact ticks per MIDI beat of six clocks.
func (r Resolution) TicksPerBeat() (fraction.Fraction, error) {
	const op = "Resolution.TicksPerBeat"
	if err := r.check(op); err != nil {
		return fraction.Fraction{}, err
	}
	return fractionOf(op, big.NewRat(int64(r.d), ClocksPerQuarter/ClocksPerBeat))
}

// MustTicksPerBeat is like TicksPerBeat but panics with its error.
func (r Resolution) MustTicksPerBeat() fraction.Fraction { return must(r.TicksPerBeat()) }

// countsOf returns the exact count p*perQuarter/w of a position.
func (r Resolution) countsOf(op string, p duration.Position, perQuarter int) (fraction.Fraction, error) {
	if err := r.check(op); err != nil {
		return fraction.Fraction{}, err
	}
	if err := checkPosition(op, p); err != nil {
		return fraction.Fraction{}, err
	}
	value := new(big.Rat).Mul(ratOfPosition(p), ratInt(perQuarter))
	return fractionOf(op, value.Quo(value, r.quarter()))
}

// positionOfCount returns the position count*w/perQuarter.
func (r Resolution) positionOfCount(op string, count, perQuarter int) (duration.Position, error) {
	if err := r.check(op); err != nil {
		return duration.Position{}, err
	}
	if count < 0 {
		return duration.Position{}, invalidError(op, "negative count "+strconv.Itoa(count))
	}
	value := new(big.Rat).Mul(ratInt(count), r.quarter())
	return positionOf(op, value.Quo(value, ratInt(perQuarter)))
}

// Clocks returns the exact MIDI clocks of the position p: 24*p/w.
func (r Resolution) Clocks(p duration.Position) (fraction.Fraction, error) {
	return r.countsOf("Resolution.Clocks", p, ClocksPerQuarter)
}

// MustClocks is like Clocks but panics with its error.
func (r Resolution) MustClocks(p duration.Position) fraction.Fraction { return must(r.Clocks(p)) }

// ClockPosition returns the position of clocks >= 0 MIDI clocks: clocks*w/24.
func (r Resolution) ClockPosition(clocks int) (duration.Position, error) {
	return r.positionOfCount("Resolution.ClockPosition", clocks, ClocksPerQuarter)
}

// MustClockPosition is like ClockPosition but panics with its error.
func (r Resolution) MustClockPosition(clocks int) duration.Position {
	return must(r.ClockPosition(clocks))
}

// Beats returns the exact MIDI beats (sixteenths of six clocks) of the
// position p: 4*p/w.
func (r Resolution) Beats(p duration.Position) (fraction.Fraction, error) {
	return r.countsOf("Resolution.Beats", p, ClocksPerQuarter/ClocksPerBeat)
}

// MustBeats is like Beats but panics with its error.
func (r Resolution) MustBeats(p duration.Position) fraction.Fraction { return must(r.Beats(p)) }

// BeatPosition returns the position of beats >= 0 MIDI beats: beats*w/4.
func (r Resolution) BeatPosition(beats int) (duration.Position, error) {
	return r.positionOfCount("Resolution.BeatPosition", beats, ClocksPerQuarter/ClocksPerBeat)
}

// MustBeatPosition is like BeatPosition but panics with its error.
func (r Resolution) MustBeatPosition(beats int) duration.Position {
	return must(r.BeatPosition(beats))
}
