package scalechoice

import (
	"errors"
	"slices"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

// Difference reports musical identity rather than order changes. Kept and
// Added belong to the new side, Removed to the original side.
type Difference[T any] struct{ Kept, Removed, Added []T }

// Motion reports independent written and positional measurements. A failed
// measurement has its own error and a zero value; cents are ignored.
type Motion struct {
	Named                      interval.Interval
	Chromatic                  interval.Chromatic
	NamedError, ChromaticError error
}

// EdgeChange compares one adjacent pair before and after a trial.
type EdgeChange struct {
	From, To      int
	Before, After Motion
}

// Change retains a complete independent After analysis, affected places,
// observed-tone differences, collection/readings differences and local edges.
// Matched and Changed are zero-based sorted indices of original records.
type Change struct {
	After            Analysis
	Matched, Changed []int
	Tones            Difference[Tone]
	Groups           Difference[int]
	Readings         Difference[Reading]
	Neighbors        []EdgeChange
}

// ReplaceAll changes every exact occurrence of oldName, in all registers.
// Known written octaves and cents are preserved. Without interpretable written
// register a spelling analysis retains the new name without an octave; range
// errors still fail the whole operation. Source snapshots are never reread.
func (a Analysis) ReplaceAll(oldName, newName spelling.Spelling) (Change, error) {
	if !a.IsValid() || !oldName.IsValid() || !newName.IsValid() {
		return Change{}, ErrUnset
	}
	if _, err := a.tones([]spelling.Spelling{newName}); err != nil {
		return Change{}, err
	}
	notes := slices.Clone(a.data.notes)
	var matched, changed []int
	for j, n := range notes {
		if !n.Spelling().Equal(oldName) {
			continue
		}
		matched = append(matched, j)
		if oldName.Equal(newName) {
			continue
		}
		replacement, err := a.replaceName(j, newName)
		if err != nil {
			return Change{}, err
		}
		notes[j] = replacement
		if !replacement.Equal(n) {
			changed = append(changed, j)
		}
	}
	return a.change(notes, matched, changed)
}

// MustReplaceAll panics with the error returned by ReplaceAll.
func (a Analysis) MustReplaceAll(oldName, newName spelling.Spelling) Change {
	return must(a.ReplaceAll(oldName, newName))
}
func (a Analysis) replaceName(index int, name spelling.Spelling) (note.Note, error) {
	n := a.data.notes[index]
	withoutOctave := func() (note.Note, error) { return note.FromSpelling(name, n.Cents(), nil) }
	if !n.HasOctave() {
		return withoutOctave()
	}
	sys := a.data.opts.System
	if sys == nil {
		sys = a.data.contexts[index].sys
	}
	if sys == nil {
		return withoutOctave()
	}
	written, err := sys.WrittenOctave(n.MustPitch())
	if err != nil {
		if a.data.opts.Comparison == BySpelling && errors.Is(err, system.ErrInadmissible) {
			return withoutOctave()
		}
		return note.Note{}, err
	}
	target, err := sys.FromWritten(name, written)
	if err != nil {
		if a.data.opts.Comparison == BySpelling && errors.Is(err, system.ErrInadmissible) {
			return withoutOctave()
		}
		return note.Note{}, err
	}
	return note.FromPitch(target, n.Cents(), nil)
}

// ReplaceAt replaces one registered or unregistered record of a phrase. It
// ignores the replacement's source, retaining this place's captured context.
// A changed correction counts as a record change, even at the same position.
func (a Analysis) ReplaceAt(index int, replacement note.Note) (Change, error) {
	if !a.IsValid() {
		return Change{}, ErrUnset
	}
	if !a.IsPhrase() || index < 0 || index >= len(a.data.notes) {
		return Change{}, ErrInvalidArgument
	}
	if !replacement.IsValid() {
		return Change{}, ErrUnset
	}
	if _, err := a.tones([]spelling.Spelling{replacement.Spelling()}); err != nil {
		return Change{}, err
	}
	if err := replacement.Unbind(); err != nil {
		return Change{}, err
	}
	notes := slices.Clone(a.data.notes)
	var changed []int
	if !notes[index].Equal(replacement) {
		changed = []int{index}
	}
	notes[index] = replacement
	return a.change(notes, []int{index}, changed)
}

// MustReplaceAt panics with the error returned by ReplaceAt.
func (a Analysis) MustReplaceAt(index int, replacement note.Note) Change {
	return must(a.ReplaceAt(index, replacement))
}
func (a Analysis) change(notes []note.Note, matched, changed []int) (Change, error) {
	after := a
	if len(changed) > 0 {
		var err error
		after, err = buildAnalysis(notes, a.data.contexts, a.data.scope, a.data.opts, a.data.phrase)
		if err != nil {
			return Change{}, err
		}
	}
	result := Change{After: after, Matched: matched, Changed: changed}
	result.Tones = toneDifference(a.data.observed, after.data.observed)
	result.Groups = groupDifference(a.data.groups, after.data.groups)
	result.Readings = itemDifference(allReadings(a.data.groups), allReadings(after.data.groups), func(r Reading) string { return r.identity })
	if a.IsPhrase() {
		edges := make(map[int]bool, 2*len(changed))
		for _, j := range changed {
			if j > 0 {
				edges[j-1] = true
			}
			if j+1 < len(notes) {
				edges[j] = true
			}
		}
		indices := make([]int, 0, len(edges))
		for j := range edges {
			indices = append(indices, j)
		}
		slices.Sort(indices)
		for _, j := range indices {
			result.Neighbors = append(result.Neighbors, EdgeChange{From: j, To: j + 1, Before: a.motion(j), After: after.motion(j)})
		}
	}
	return result, nil
}
func itemDifference[T any, K comparable](before, after []T, key func(T) K) Difference[T] {
	old := make(map[K]bool, len(before))
	next := make(map[K]bool, len(after))
	for _, v := range before {
		old[key(v)] = true
	}
	for _, v := range after {
		next[key(v)] = true
	}
	result := Difference[T]{}
	for _, v := range before {
		if !next[key(v)] {
			result.Removed = append(result.Removed, v)
		}
	}
	for _, v := range after {
		if old[key(v)] {
			result.Kept = append(result.Kept, v)
		} else {
			result.Added = append(result.Added, v)
		}
	}
	return result
}
func toneDifference(before, after []Tone) Difference[Tone] {
	result := itemDifference(before, after, keyOf)
	old := toneMap(before)
	for j := range result.Kept {
		t := &result.Kept[j]
		t.Names = unionNames(old[keyOf(*t)].Names, t.Names)
	}
	result.Removed = cloneTones(result.Removed)
	result.Added = cloneTones(result.Added)
	return result
}
func groupDifference(before, after []Group) Difference[int] {
	old := make(map[string]bool, len(before))
	next := make(map[string]bool, len(after))
	for _, g := range before {
		old[toneListKey(g.Tones)] = true
	}
	for _, g := range after {
		next[toneListKey(g.Tones)] = true
	}
	result := Difference[int]{}
	for j, g := range before {
		if !next[toneListKey(g.Tones)] {
			result.Removed = append(result.Removed, j)
		}
	}
	for j, g := range after {
		if old[toneListKey(g.Tones)] {
			result.Kept = append(result.Kept, j)
		} else {
			result.Added = append(result.Added, j)
		}
	}
	return result
}
func allReadings(groups []Group) []Reading {
	var result []Reading
	for _, g := range groups {
		result = append(result, g.Readings...)
	}
	return result
}
func (a Analysis) edgeSystem(index int) (*system.System, error) {
	if a.data.opts.System != nil {
		return a.data.opts.System, nil
	}
	left, right := a.data.contexts[index], a.data.contexts[index+1]
	if left.err != nil {
		return nil, left.err
	}
	if right.err != nil {
		return nil, right.err
	}
	if !left.sys.Equal(right.sys) {
		return nil, errors.Join(ErrDifferentSystems, note.ErrDifferentSystems)
	}
	return left.sys, nil
}
func (a Analysis) motion(index int) Motion {
	left, right := a.data.notes[index], a.data.notes[index+1]
	if !left.HasOctave() || !right.HasOctave() {
		return Motion{NamedError: interval.ErrNoOctave, ChromaticError: interval.ErrNoOctave}
	}
	sys, err := a.edgeSystem(index)
	if err != nil {
		return Motion{NamedError: err, ChromaticError: err}
	}
	// Chromatic exposes its system pointer. Give this result its own value so
	// even caller reassignment through that pointer cannot affect any analysis.
	sys = cloneSystem(sys)
	named, namedErr := interval.BetweenPitches(sys, left.MustPitch(), right.MustPitch())
	chromatic, chromaticErr := interval.ChromaticBetweenNotesIn(sys, left, right)
	return Motion{Named: named, Chromatic: chromatic, NamedError: namedErr, ChromaticError: chromaticErr}
}
