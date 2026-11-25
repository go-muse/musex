package voiceleading

import (
	"math"
	"slices"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

// VoiceMove is the move of voice Voice from one sonority to the next: the
// written interval with its direction and the signed chromatic distance.
type VoiceMove struct {
	Voice     int
	From, To  note.Note
	Interval  interval.Interval
	Chromatic system.Step
}

// IsRepeat reports whether the voice repeats the same written note, that is,
// keeps a common tone.
func (m VoiceMove) IsRepeat() bool { return m.Interval.IsPerfectUnison() }

// IsStep reports whether the move is a second of any quality.
func (m VoiceMove) IsStep() bool { return m.Interval.Size() == 2 }

// IsLeap reports whether the move is a leap under def, as Move.IsLeap does.
func (m VoiceMove) IsLeap(def LeapDefinition) bool { return isLeap(m.Interval, def) }

// Connection describes the transition between two sonorities with an equal
// number of sounding voices. Moves lists the move of every voice; Common the
// classes by spelling present in both sonorities, in order of first
// appearance by voice of From; Total the sum of the absolute chromatic moves.
// The connection is harmonic when at least one voice keeps a common tone,
// otherwise melodic. From and To are the borrowed snapshots; Moves and Common
// belong to the caller.
type Connection struct {
	From, To *Sonority
	Moves    []VoiceMove
	Common   []spelling.Spelling
	Total    system.Step
	Kind     ConnectionKind
}

// errTotal reports a total movement that does not fit in an int.
var errTotal = cause("total movement", system.ErrOutOfRange)

// addAbs adds |x| to total and reports false when the result does not fit.
func addAbs(total, x system.Step) (system.Step, bool) {
	if x == math.MinInt {
		return 0, false
	}
	if x < 0 {
		x = -x
	}
	if total > math.MaxInt-x {
		return 0, false
	}
	return total + x, true
}

// Connect describes the connection of two sonorities, measured in their common
// system. It reads no sources. Errors, in this order: ErrUnset for from, then
// to; ErrNoOrder for from, then to; ErrSilentPart for from, then to;
// ErrVoiceCount for different numbers of voices or no voices;
// ErrDifferentSystems; measurement failures of a move (causes of package
// interval or system); a total that does not fit in an int (with
// system.ErrOutOfRange). The last two carry no sentinel of this package.
func Connect(from, to *Sonority) (Connection, error) {
	if err := checkConnectable(from, to); err != nil {
		return Connection{}, err
	}
	c := Connection{From: from, To: to, Moves: make([]VoiceMove, len(from.notes)), Common: []spelling.Spelling{}}
	for k := range from.notes {
		mv, err := measureVoiceMove(from, to, k)
		if err != nil {
			return Connection{}, err
		}
		c.Moves[k] = mv
	}
	if err := c.describe(); err != nil {
		return Connection{}, err
	}
	return c, nil
}

// MustConnect is like Connect but panics with its error.
func MustConnect(from, to *Sonority) Connection { return must(Connect(from, to)) }

func checkConnectable(from, to *Sonority) error {
	switch {
	case !from.IsSet() || !to.IsSet():
		return ErrUnset
	case from.order != LowToHigh || to.order != LowToHigh:
		return ErrNoOrder
	case len(from.silent) > 0 || len(to.silent) > 0:
		return ErrSilentPart
	case len(from.notes) != len(to.notes) || len(from.notes) == 0:
		return ErrVoiceCount
	case !from.sys.Equal(to.sys):
		return ErrDifferentSystems
	}
	return nil
}

func measureVoiceMove(from, to *Sonority, k int) (VoiceMove, error) {
	x, y := from.pitches[k], to.pitches[k]
	i, err := interval.BetweenPitches(from.sys, x, y)
	if err != nil {
		return VoiceMove{}, cause("voice move", err)
	}
	d, err := from.sys.Distance(x, y)
	if err != nil {
		return VoiceMove{}, cause("voice move", err)
	}
	return VoiceMove{Voice: k, From: from.notes[k], To: to.notes[k], Interval: i, Chromatic: d}, nil
}

// describe fills Common, Total and Kind from the measured moves.
func (c *Connection) describe() error {
	c.Common = commonClasses(c.From.notes, c.To.notes)
	c.Kind = MelodicConnection
	var total system.Step
	for _, mv := range c.Moves {
		var ok bool
		if total, ok = addAbs(total, mv.Chromatic); !ok {
			return errTotal
		}
		if mv.IsRepeat() {
			c.Kind = HarmonicConnection
		}
	}
	c.Total = total
	return nil
}

func commonClasses(from, to []note.Note) []spelling.Spelling {
	out := []spelling.Spelling{}
	for _, x := range from {
		name := x.Spelling()
		if slices.ContainsFunc(out, name.Equal) {
			continue
		}
		if slices.ContainsFunc(to, func(y note.Note) bool { return y.Spelling().Equal(name) }) {
			out = append(out, name)
		}
	}
	return out
}

// PartConnection is the connection of two adjacent part sonorities of an
// analysis. Connection.From and Connection.To are From.Voices and To.Voices.
type PartConnection struct {
	From, To   PartSonority
	Connection Connection
}

// Connections returns the connections of adjacent moments in which every part
// sounds at both ends, in time order. A rest at either end forms no
// connection, and moments are not joined across it. Pairwise v1 searches are
// independent of these connections. Errors, in this order: ErrUnset;
// ErrNoOrder without LowToHigh, also for empty material; a total that does not
// fit in an int (with system.ErrOutOfRange).
func (a *Analysis) Connections() ([]PartConnection, error) {
	if err := a.checkOrder(); err != nil {
		return nil, err
	}
	starts, end := a.moments()
	out := []PartConnection{}
	for i := 1; i < len(starts); i++ {
		pc, ok, err := a.partConnection(starts, i, end)
		if err != nil {
			return nil, err
		}
		if ok {
			out = append(out, pc)
		}
	}
	return out, nil
}

// MustConnections is like Connections but panics with its error.
func (a *Analysis) MustConnections() []PartConnection { return must(a.Connections()) }

func (a *Analysis) partConnection(starts []duration.Position, i int, end duration.Position) (PartConnection, bool, error) {
	from := a.partSonority(starts[i-1], starts[i])
	to := a.partSonority(starts[i], momentEnd(starts, i, end))
	if len(from.Voices.silent) > 0 || len(to.Voices.silent) > 0 {
		return PartConnection{}, false, nil
	}
	c, err := Connect(from.Voices, to.Voices)
	if err != nil {
		return PartConnection{}, false, err
	}
	return PartConnection{From: from, To: to, Connection: c}, true, nil
}

// ConnectionAt returns the connection of the moments starting at from and to.
// If a part is silent at either end it returns the zero value, false, nil.
// Errors, in this order: ErrUnset; ErrNoOrder; ErrUnset with duration.ErrUnset
// for from, then to; ErrTransition unless from and to are two adjacent
// increasing moment starts; a total that does not fit in an int.
func (a *Analysis) ConnectionAt(from, to duration.Position) (PartConnection, bool, error) {
	if err := a.checkOrder(); err != nil {
		return PartConnection{}, false, err
	}
	if !from.IsSet() {
		return PartConnection{}, false, unset("from", duration.ErrUnset)
	}
	if !to.IsSet() {
		return PartConnection{}, false, unset("to", duration.ErrUnset)
	}
	starts, end := a.moments()
	i, found := slices.BinarySearchFunc(starts, from, duration.Position.Compare)
	if !found || i+1 >= len(starts) || !starts[i+1].Equal(to) {
		return PartConnection{}, false, ErrTransition
	}
	return a.partConnection(starts, i+1, end)
}

// MustConnectionAt is like ConnectionAt but panics with its error.
func (a *Analysis) MustConnectionAt(from, to duration.Position) (PartConnection, bool) {
	return must2(a.ConnectionAt(from, to))
}
