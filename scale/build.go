package scale

import (
	"errors"
	"fmt"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pcseq"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

// ErrToken reports a token that FromTokens could not read.
var ErrToken = errors.New("scale: invalid token")

// TokenError reports the failure of one token of [FromTokens]. Index is the
// zero-based index of the failing token; Err wraps [ErrToken] and the cause.
// The error is reachable with [errors.As], and both ErrToken and the cause
// match with [errors.Is].
type TokenError struct {
	// Index is the zero-based index of the failing token.
	Index int
	// Err is the wrapped cause.
	Err error
}

func (e *TokenError) Error() string { return fmt.Sprintf("scale: token %d: %v", e.Index, e.Err) }
func (e *TokenError) Unwrap() error { return e.Err }

// FromSpellings returns the scale of the name-only notes with the spellings
// names, in input order, preserving repetitions: no octaves, zero cents, and
// the source src, as by [note.FromSpellings]. An empty or nil names gives an
// empty scale. The errors are those of note.FromSpellings.
func FromSpellings(names []spelling.Spelling, src settings.Source) (*Scale, error) {
	ns, err := note.FromSpellings(names, src)
	if err != nil {
		return nil, err
	}
	return New(ns...)
}

// MustFromSpellings is like FromSpellings but panics with the returned error.
func MustFromSpellings(names []spelling.Spelling, src settings.Source) *Scale {
	return must(FromSpellings(names, src))
}

// FromTokens returns the scale of the notes read from tokens in the notation
// of the settings of src, in input order, preserving repetitions. Each token
// is read as a written pitch and, if that fails, as a name; names give notes
// without octaves, written pitches give notes with octaves, all with zero
// cents and bound to src.
//
// A nil src gives an error matching ErrUnset. A failing token gives a
// [*TokenError] carrying its index; the reported cause is that of the name
// reading. An empty token list gives an empty scale.
func FromTokens(src settings.Source, tokens ...string) (*Scale, error) {
	if src == nil {
		return nil, errors.Join(ErrUnset, note.ErrUnset)
	}
	notes := make([]note.Note, 0, len(tokens))
	for i, tok := range tokens {
		n, err := note.ParsePitch(src, tok)
		if err != nil {
			n, err = note.Parse(src, tok)
		}
		if err != nil {
			return nil, &TokenError{Index: i, Err: fmt.Errorf("%w: %w", ErrToken, err)}
		}
		notes = append(notes, n)
	}
	return New(notes...)
}

// MustFromTokens is like FromTokens but panics with the returned error.
func MustFromTokens(src settings.Source, tokens ...string) *Scale {
	return must(FromTokens(src, tokens...))
}

// Tokens returns the tokens of the notes of s, one per note, in order, each
// formatted in the notation of its own source. It returns an error wrapping
// ErrUnset if s is nil, and the errors of [note.Note.Format] otherwise. The
// caller owns the result.
func (s *Scale) Tokens() ([]string, error) {
	if s == nil {
		return nil, ErrUnset
	}
	out := make([]string, 0, len(s.notes))
	for i, n := range s.notes {
		text, err := n.Format()
		if err != nil {
			return nil, fmt.Errorf("note %d: %w", i, err)
		}
		out = append(out, text)
	}
	return out, nil
}

// MustTokens is like Tokens but panics with the returned error.
func (s *Scale) MustTokens() []string { return must(s.Tokens()) }

// TokensIn returns the tokens of the notes of s, one per note, in order, each
// formatted in the notation of the explicit source src. It returns an error
// wrapping ErrUnset if s or src is nil, and the errors of [note.Note.FormatIn]
// otherwise. The caller owns the result.
func (s *Scale) TokensIn(src settings.Source) ([]string, error) {
	if s == nil {
		return nil, ErrUnset
	}
	if src == nil {
		return nil, errors.Join(ErrUnset, settings.ErrUnset)
	}
	out := make([]string, 0, len(s.notes))
	for i, n := range s.notes {
		text, err := n.FormatIn(src)
		if err != nil {
			return nil, fmt.Errorf("note %d: %w", i, err)
		}
		out = append(out, text)
	}
	return out, nil
}

// MustTokensIn is like TokensIn but panics with the returned error.
func (s *Scale) MustTokensIn(src settings.Source) []string { return must(s.TokensIn(src)) }

// Spellings returns the spellings of the notes of s, in order, preserving
// repetitions. The caller owns the result; it is nil for a nil receiver.
func (s *Scale) Spellings() []spelling.Spelling {
	if s == nil {
		return nil
	}
	out := make([]spelling.Spelling, 0, len(s.notes))
	for _, n := range s.notes {
		out = append(out, n.Spelling())
	}
	return out
}

// ChromaticRun returns the playable chromatic run from the position of from
// to the position of to, both endpoints included, as a scale of notes bound
// to src with zero cents. The system is taken from one settings snapshot of
// src; the endpoints give positions, and every note, including the endpoints,
// is spelled by rule. C4 through C5 in the traditional system gives 13 notes.
//
// The caller passes the directional rule [system.ChromaticScale] for a key
// and direction (C major ascending gives B♭ at the tenth semitone) or an
// explicit [system.Sharps] or [system.Flats] policy. Direction follows the
// order of the endpoints; equal positions give one note.
//
// It returns an error wrapping ErrUnset if src is nil or its settings are
// unset, and the errors of [system.System.Walk] and [note.FromPitch]
// otherwise.
func ChromaticRun(from, to pitch.Pitch, rule system.SpellingRule, src settings.Source) (*Scale, error) {
	if src == nil {
		return nil, errors.Join(ErrUnset, settings.ErrUnset)
	}
	set := src.Settings()
	if !set.IsValid() {
		return nil, errors.Join(ErrUnset, settings.ErrUnset)
	}

	seq, err := set.System().Walk(from, to, rule)
	if err != nil {
		return nil, err
	}

	notes := make([]note.Note, 0)
	for p := range seq {
		n, err := note.FromPitch(p, cents.Cents{}, src)
		if err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return New(notes...)
}

// MustChromaticRun is like ChromaticRun but panics with the returned error.
func MustChromaticRun(from, to pitch.Pitch, rule system.SpellingRule, src settings.Source) *Scale {
	return must(ChromaticRun(from, to, rule, src))
}

// PitchClassSequence returns the ordered sequence of the name classes of the
// notes of s in the system sys: one output position per input note, preserving
// repetitions and enharmonic equivalents. C4, D♭4, C♯5, D4 in the traditional
// system give [0, 1, 1, 2]. An empty scale gives an empty sequence of the
// modulus of sys.
//
// It returns an error wrapping ErrUnset if s is nil, ErrUnset together with
// system.ErrInvalid if sys is unset, and the errors of [system.System.Class]
// otherwise.
func (s *Scale) PitchClassSequence(sys *system.System) (pcseq.Sequence, error) {
	if s == nil {
		return pcseq.Sequence{}, ErrUnset
	}
	if !sys.IsValid() {
		return pcseq.Sequence{}, errors.Join(ErrUnset, system.ErrInvalid)
	}
	if len(s.notes) == 0 {
		return pcseq.Empty(sys.Classes())
	}
	values := make([]int, len(s.notes))
	for i, n := range s.notes {
		c, err := sys.Class(n.Spelling())
		if err != nil {
			return pcseq.Sequence{}, err
		}
		values[i] = c.Number()
	}
	return pcseq.New(sys.Classes(), values...)
}

// MustPitchClassSequence is like PitchClassSequence but panics with the
// returned error.
func (s *Scale) MustPitchClassSequence(sys *system.System) pcseq.Sequence {
	return must(s.PitchClassSequence(sys))
}
