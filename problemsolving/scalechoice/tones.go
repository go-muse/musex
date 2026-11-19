package scalechoice

import (
	"errors"
	"slices"
	"strconv"
	"strings"

	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/spelling"
)

// Tone represents one exact name or one valid positional class. Positional
// Names contains only spellings observed in the corresponding result.
type Tone struct {
	Names []spelling.Spelling
	Class pitchclass.Class
}

// Group unites all readings of one complete collection. Missing excludes the
// observed input. Every returned slice, including nested names, is caller-owned.
type Group struct {
	Tones    []Tone
	Missing  []Tone
	Readings []Reading
}

// Partition contains sorted group indices; together the lists cover the source.
type Partition struct{ Kept, Dropped []int }

// Continuation is an unplayed tone and its effect on already found groups.
type Continuation struct {
	Tone   Tone
	Groups Partition
}
type toneKey struct {
	name  spelling.Spelling
	class pitchclass.Class
}

func keyOf(t Tone) toneKey {
	if t.Class.IsValid() {
		return toneKey{class: t.Class}
	}
	return toneKey{name: t.Names[0]}
}
func toneMap(tones []Tone) map[toneKey]Tone {
	m := make(map[toneKey]Tone, len(tones))
	for _, t := range tones {
		m[keyOf(t)] = t
	}
	return m
}
func cloneTone(t Tone) Tone { t.Names = slices.Clone(t.Names); return t }
func cloneTones(tones []Tone) []Tone {
	result := make([]Tone, len(tones))
	for j, t := range tones {
		result[j] = cloneTone(t)
	}
	return result
}
func unionNames(a, b []spelling.Spelling) []spelling.Spelling {
	result := append(slices.Clone(a), b...)
	slices.SortFunc(result, spelling.Spelling.Compare)
	return slices.Compact(result)
}
func compareTones(a, b Tone) int {
	if a.Class.IsValid() {
		return a.Class.Compare(b.Class)
	}
	return a.Names[0].Compare(b.Names[0])
}
func mergeTones(a, b []Tone) []Tone {
	result := cloneTones(a)
	indices := make(map[toneKey]int, len(a)+len(b))
	for j, t := range result {
		indices[keyOf(t)] = j
	}
	for _, t := range b {
		if j, ok := indices[keyOf(t)]; ok {
			result[j].Names = unionNames(result[j].Names, t.Names)
		} else {
			indices[keyOf(t)] = len(result)
			result = append(result, cloneTone(t))
		}
	}
	slices.SortFunc(result, compareTones)
	return result
}
func toneListKey(tones []Tone) string {
	var b strings.Builder
	for _, t := range tones {
		if t.Class.IsValid() {
			b.WriteString(strconv.Itoa(t.Class.Number()))
		} else {
			b.WriteString(t.Names[0].String())
		}
		b.WriteByte('|')
	}
	return b.String()
}
func (a Analysis) tones(names []spelling.Spelling) ([]Tone, error) {
	result := make([]Tone, 0, len(names))
	indices := make(map[toneKey]int, len(names))
	for _, name := range names {
		if !name.IsValid() {
			return nil, ErrUnset
		}
		t := Tone{Names: []spelling.Spelling{name}}
		if a.data.opts.Comparison == ByPosition {
			class, err := a.data.opts.System.Class(name)
			if err != nil {
				return nil, errors.Join(ErrUninterpretableName, err)
			}
			t.Class = class
		}
		key := keyOf(t)
		if j, ok := indices[key]; ok {
			result[j].Names = unionNames(result[j].Names, t.Names)
		} else {
			indices[key] = len(result)
			result = append(result, t)
		}
	}
	slices.SortFunc(result, compareTones)
	return result, nil
}

// Groups returns complete collections in first-search-result order, with
// readings in search order. Mutating the output never alters the analysis.
func (a Analysis) Groups() []Group {
	if !a.IsValid() {
		return nil
	}
	result := make([]Group, len(a.data.groups))
	for j, g := range a.data.groups {
		result[j] = Group{Tones: cloneTones(g.Tones), Missing: cloneTones(g.Missing), Readings: slices.Clone(g.Readings)}
	}
	return result
}

// Common returns the intersection of complete found collections, including
// already played notes. Positional names are united across matching groups.
func (a Analysis) Common() []Tone {
	if !a.IsValid() || len(a.data.groups) == 0 {
		return nil
	}
	result := cloneTones(a.data.groups[0].Tones)
	for _, g := range a.data.groups[1:] {
		set := toneMap(g.Tones)
		next := make([]Tone, 0, len(result))
		for _, t := range result {
			if other, ok := set[keyOf(t)]; ok {
				t.Names = unionNames(t.Names, other.Names)
				next = append(next, t)
			}
		}
		result = next
	}
	return result
}
func (a Analysis) partition(tones []Tone) Partition {
	result := Partition{}
	for j, g := range a.data.groups {
		set := toneMap(g.Tones)
		keep := true
		for _, t := range tones {
			if _, ok := set[keyOf(t)]; !ok {
				keep = false
				break
			}
		}
		if keep {
			result.Kept = append(result.Kept, j)
		} else {
			result.Dropped = append(result.Dropped, j)
		}
	}
	return result
}

// AfterAdding filters existing groups, without searching again. Empty input
// preserves all groups. Names are validated even when no groups were found.
func (a Analysis) AfterAdding(names ...spelling.Spelling) (Partition, error) {
	if !a.IsValid() {
		return Partition{}, ErrUnset
	}
	tones, err := a.tones(names)
	if err != nil {
		return Partition{}, err
	}
	return a.partition(tones), nil
}

// MustAfterAdding panics with the error returned by AfterAdding.
func (a Analysis) MustAfterAdding(names ...spelling.Spelling) Partition {
	return must(a.AfterAdding(names...))
}

// Continuations enumerates all found but unplayed tones and the resulting
// partition, including continuations that preserve every group.
func (a Analysis) Continuations() []Continuation {
	if !a.IsValid() {
		return nil
	}
	var union []Tone
	for _, g := range a.data.groups {
		union = mergeTones(union, g.Tones)
	}
	observed := toneMap(a.data.observed)
	result := make([]Continuation, 0, len(union))
	for _, t := range union {
		if _, ok := observed[keyOf(t)]; !ok {
			result = append(result, Continuation{Tone: cloneTone(t), Groups: a.partition([]Tone{t})})
		}
	}
	return result
}
