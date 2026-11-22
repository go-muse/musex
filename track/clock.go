package track

import (
	"errors"
	"slices"

	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempomap"
)

// Clock returns an independent container of the recorded schedule.
func (t *Track) Clock() (Clock, bool, error) {
	if !t.IsSet() {
		return Clock{}, false, ErrUnset
	}
	c := t.clock
	c.BarHolds = slices.Clone(c.BarHolds)
	return c, c.Map.IsSet(), nil
}
func (t *Track) validateClock() ([]Issue, error) {
	var out []Issue
	c := t.clock
	if !c.Map.IsSet() {
		if len(c.BarHolds) > 0 {
			return nil, ErrInvalidArgument
		}
		return nil, nil
	}
	if c.Map.MustAddressing() != tempomap.MeasureAddressing {
		return nil, ErrInvalidArgument
	}
	// Validate individual attachments even on an otherwise empty track.
	changes := c.Map.MustChanges()
	exts := c.Map.MustExtensions()
	for _, ch := range changes {
		a, _ := ch.MustAddress().MustBar()
		if ok, _ := t.grid.Contains(a); !ok {
			x := issue(ErrClockAddress, Place{Bar: a.MustBar(), Index: -1})
			x.Address = a
			x.Kind = TempoChange
			x.Slot = ch.MustSlot()
			out = append(out, x)
		}
	}
	for _, ex := range exts {
		a, _ := ex.MustAddress().MustBar()
		if ok, _ := t.grid.Contains(a); !ok {
			x := issue(ErrClockAddress, Place{Bar: a.MustBar(), Index: -1})
			x.Address = a
			x.Kind = Extension
			out = append(out, x)
		}
	}
	positions := map[duration.Position]bool{}
	for _, ex := range exts {
		a, _ := ex.MustAddress().MustBar()
		p, err := t.grid.PositionOf(a)
		if err == nil {
			positions[p] = true
		}
	}
	for _, h := range c.BarHolds {
		x := issue(ErrClockAddress, Place{Bar: h.AfterBar, Index: -1})
		x.AfterBar = new(h.AfterBar)
		x.Kind = Extension
		p, err := t.grid.BarEnd(h.AfterBar)
		if err != nil {
			out = append(out, x)
			continue
		}
		x.Position = p
		if !h.Seconds.IsSet() {
			x.Cause = errors.Join(ErrInvalidArgument, tempo.ErrUnset)
			out = append(out, x)
			continue
		}
		if positions[p] {
			x.Cause = ErrHoldConflict
			out = append(out, x)
		}
		positions[p] = true
	}
	if len(out) > 0 {
		return out, nil
	}
	if _, err := c.Map.Validate(t.grid); err != nil {
		x := issue(errors.Join(ErrClockAddress, err), Place{Index: -1})
		return []Issue{x}, nil
	}
	return nil, nil
}

// PerformanceMap resolves measured attachments and final-bar delays into a
// standalone positional map. Signs alone never add seconds.
func (t *Track) PerformanceMap() (tempomap.Map, error) {
	if !t.IsSet() {
		return tempomap.Map{}, ErrUnset
	}
	if !t.clock.Map.IsSet() {
		return tempomap.Map{}, ErrNoClock
	}
	m, err := t.clock.Map.ToPositions(t.grid)
	if err != nil {
		return tempomap.Map{}, err
	}
	for _, h := range t.clock.BarHolds {
		p, err := t.grid.BarEnd(h.AfterBar)
		if err != nil {
			return tempomap.Map{}, err
		}
		ex, err := tempomap.NewExtension(tempomap.MustAt(p), h.Seconds)
		if err != nil {
			return tempomap.Map{}, err
		}
		m, err = m.WithExtension(ex)
		if err != nil {
			return tempomap.Map{}, err
		}
	}
	return m, nil
}

// SoundSeconds uses map.End at both the attack and release.
func (t *Track) SoundSeconds(places int, mode rounding.Mode, voices ...VoiceID) ([]TimedSound, error) {
	if !t.IsSet() {
		return nil, ErrUnset
	}
	if places < 0 || places > tempomap.MaxPlaces || !mode.IsValid() {
		return nil, ErrInvalidArgument
	}
	m, err := t.PerformanceMap()
	if err != nil {
		return nil, err
	}
	ss, err := t.Sounds(voices...)
	if err != nil {
		return nil, err
	}
	out := make([]TimedSound, 0, len(ss))
	for _, s := range ss {
		a, err := m.End(tempomap.MustAt(s.Time.Start), places, mode)
		if err != nil {
			return nil, err
		}
		b, err := m.End(tempomap.MustAt(s.Time.End), places, mode)
		if err != nil {
			return nil, err
		}
		out = append(out, TimedSound{s, a, b})
	}
	return out, nil
}
