package score

import (
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/meter"
	"github.com/go-muse/musex/tempomap"
	"github.com/go-muse/musex/track"
)

// MustNew is like New but panics with its error.
func MustNew(clock tempomap.Map) *Score { return must(New(clock)) }

// MustParts is like Parts but panics with its error.
func (s *Score) MustParts() []PartID { return must(s.Parts()) }

// MustPart is like Part but panics with its error.
func (s *Score) MustPart(id PartID) *track.Track { return must(s.Part(id)) }

// MustEnd is like End but panics with its error.
func (s *Score) MustEnd() duration.Position { return must(s.End()) }

// MustMap is like Map but panics with its error.
func (s *Score) MustMap() tempomap.Map { return must(s.Map()) }

// MustWithMap is like WithMap but panics with its error.
func (s *Score) MustWithMap(clock tempomap.Map) *Score { return must(s.WithMap(clock)) }

// MustAdd is like Add but panics with its error.
func (s *Score) MustAdd(part *track.Track) (*Score, PartID) { return must2(s.Add(part)) }

// MustReplace is like Replace but panics with its error.
func (s *Score) MustReplace(id PartID, part *track.Track) *Score { return must(s.Replace(id, part)) }

// MustRemove is like Remove but panics with its error.
func (s *Score) MustRemove(id PartID) *Score { return must(s.Remove(id)) }

// MustReorder is like Reorder but panics with its error.
func (s *Score) MustReorder(ids ...PartID) *Score { return must(s.Reorder(ids...)) }

// MustAddWithoutClock is like AddWithoutClock but panics with its error.
func (s *Score) MustAddWithoutClock(part *track.Track) Inclusion {
	return must(s.AddWithoutClock(part))
}

// MustReplaceWithoutClock is like ReplaceWithoutClock but panics with its error.
func (s *Score) MustReplaceWithoutClock(id PartID, part *track.Track) Inclusion {
	return must(s.ReplaceWithoutClock(id, part))
}

// MustExtractWithMap is like ExtractWithMap but panics with its error.
func (s *Score) MustExtractWithMap(id PartID) *track.Track { return must(s.ExtractWithMap(id)) }

// MustAt is like At but panics with its error.
func (s *Score) MustAt(p duration.Position, selection ...Selection) []PartSlice {
	return must(s.At(p, selection...))
}

// MustSounds is like Sounds but panics with its error.
func (s *Score) MustSounds(selection ...Selection) []Sound { return must(s.Sounds(selection...)) }

// MustSoundingAt is like SoundingAt but panics with its error.
func (s *Score) MustSoundingAt(p duration.Position, selection ...Selection) []Sound {
	return must(s.SoundingAt(p, selection...))
}

// MustSoundSeconds is like SoundSeconds but panics with its error.
func (s *Score) MustSoundSeconds(places int, mode rounding.Mode, selection ...Selection) []TimedSound {
	return must(s.SoundSeconds(places, mode, selection...))
}

// MustBarView is like BarView but panics with its error.
func (s *Score) MustBarView(part PartID, voice track.VoiceID, bar int) []ViewEntry {
	return must(s.BarView(part, voice, bar))
}

// MustChangeMeter is like ChangeMeter but panics with its error.
func (s *Score) MustChangeMeter(first, last int, value meter.Meter) MeterProposal {
	return must(s.ChangeMeter(first, last, value))
}

// MustCommonBarlines is like CommonBarlines but panics with its error.
func (s *Score) MustCommonBarlines(from, to duration.Position, parts ...PartID) []Barline {
	return must(s.CommonBarlines(from, to, parts...))
}

// MustBarlinePeriod is like BarlinePeriod but panics with its error.
func (s *Score) MustBarlinePeriod(a, b PartID) duration.Duration { return must(s.BarlinePeriod(a, b)) }
