package track

import (
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/instrument"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/tempomap"
	"github.com/go-muse/musex/voice"
)

// MustNew is like New but panics with the returned error.
func MustNew(inst *instrument.Instrument, grid measure.Grid) *Track { return must(New(inst, grid)) }

// MustInstrument is like Instrument but panics with the returned error.
func (t *Track) MustInstrument() *instrument.Instrument { return must(t.Instrument()) }

// MustGrid is like Grid but panics with the returned error.
func (t *Track) MustGrid() measure.Grid { return must(t.Grid()) }

// MustVoices is like Voices but panics with the returned error.
func (t *Track) MustVoices() []VoiceID { return must(t.Voices()) }

// MustAddVoice is like AddVoice but panics with the returned error.
func (t *Track) MustAddVoice() (*Track, VoiceID) { return must2(t.AddVoice()) }

// MustEdit is like Edit but panics with the returned error.
func (t *Track) MustEdit(e Edit) *Track { return must(t.Edit(e)) }

// MustSplice is like Splice but panics with the returned error.
func (t *Track) MustSplice(v VoiceID, bar, index, remove int, add ...*event.Event) *Track {
	return must(t.Splice(v, bar, index, remove, add...))
}

// MustEntries is like Entries but panics with the returned error.
func (t *Track) MustEntries(v VoiceID, bar int) []Entry { return must(t.Entries(v, bar)) }

// MustAt is like At but panics with the returned error.
func (t *Track) MustAt(v VoiceID, p duration.Position) (Entry, bool) { return must2(t.At(v, p)) }

// MustVoiceEnd is like VoiceEnd but panics with the returned error.
func (t *Track) MustVoiceEnd(v VoiceID) duration.Position { return must(t.VoiceEnd(v)) }

// MustFermataBars is like FermataBars but panics with the returned error.
func (t *Track) MustFermataBars() []int { return must(t.FermataBars()) }

// MustSounds is like Sounds but panics with the returned error.
func (t *Track) MustSounds(voices ...VoiceID) []Sound { return must(t.Sounds(voices...)) }

// MustSoundingAt is like SoundingAt but panics with the returned error.
func (t *Track) MustSoundingAt(p duration.Position, voices ...VoiceID) []Sound {
	return must(t.SoundingAt(p, voices...))
}

// MustClock is like Clock but panics with the returned error.
func (t *Track) MustClock() (Clock, bool) { return must2(t.Clock()) }

// MustPerformanceMap is like PerformanceMap but panics with the returned error.
func (t *Track) MustPerformanceMap() tempomap.Map { return must(t.PerformanceMap()) }

// MustSoundSeconds is like SoundSeconds but panics with the returned error.
func (t *Track) MustSoundSeconds(places int, mode rounding.Mode, voices ...VoiceID) []TimedSound {
	return must(t.SoundSeconds(places, mode, voices...))
}

// MustWriteEvents is like WriteEvents but panics with the returned error.
func (t *Track) MustWriteEvents(v VoiceID, at measure.Address, events []*event.Event, opts WriteOptions) Proposal {
	return must(t.WriteEvents(v, at, events, opts))
}

// MustRebar is like Rebar but panics with the returned error.
func (t *Track) MustRebar(grid measure.Grid, groups Groups) Proposal {
	return must(t.Rebar(grid, groups))
}

// MustQuantize is like Quantize but panics with the returned error.
func (t *Track) MustQuantize(v VoiceID, first, last int, step duration.Duration, mode rounding.Mode, groups Groups) Proposal {
	return must(t.Quantize(v, first, last, step, mode, groups))
}

// MustSplit is like Split but panics with the returned error.
func (t *Track) MustSplit(at Place, offset duration.Duration, groups Groups) Proposal {
	return must(t.Split(at, offset, groups))
}

// MustMerge is like Merge but panics with the returned error.
func (t *Track) MustMerge(v VoiceID, bar, index, count int, groups Groups) Proposal {
	return must(t.Merge(v, bar, index, count, groups))
}

// MustPackRests is like PackRests but panics with the returned error.
func (t *Track) MustPackRests(v VoiceID, first, last int) Proposal {
	return must(t.PackRests(v, first, last))
}

// MustUnpackRest is like UnpackRest but panics with the returned error.
func (t *Track) MustUnpackRest(at Place) Proposal { return must(t.UnpackRest(at)) }

// MustExtractLine is like ExtractLine but panics with the returned error.
func (t *Track) MustExtractLine(v VoiceID, mode LineMode) Extraction {
	return must(t.ExtractLine(v, mode))
}

// MustWriteLine is like WriteLine but panics with the returned error.
func (t *Track) MustWriteLine(v VoiceID, at measure.Address, line *voice.Line, opts WriteOptions) Proposal {
	return must(t.WriteLine(v, at, line, opts))
}
