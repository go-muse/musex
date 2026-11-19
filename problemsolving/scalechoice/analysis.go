// Package scalechoice supports a composer's choice of pitch collections and
// modal readings, and compares independent trial changes to a set or phrase.
// Analyses retain immutable snapshots, never live settings sources. Returned
// slices and systems belong to the caller. Copies and independent branches
// support concurrent reads; no background work or global cache is used.
package scalechoice

import (
	"errors"
	"fmt"
	"slices"
	"strconv"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/search"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/template"
)

// Error categories preserve the corresponding search semantics and any
// underlying cause. Use errors.Is rather than comparing diagnostic text.
var (
	ErrUnset               = search.ErrUnset
	ErrEmptyInput          = search.ErrEmptyInput
	ErrInvalidArgument     = search.ErrInvalidArgument
	ErrNoSystem            = search.ErrNoSystem
	ErrDifferentSystems    = search.ErrDifferentSystems
	ErrUninterpretableName = search.ErrUninterpretableName
	ErrContext             = search.ErrContext
)

// Comparison selects exact spelling or pitch-class equality.
type Comparison uint8

const (
	BySpelling Comparison = iota + 1
	ByPosition
)

// Options requires an explicit condition and comparison. An unset Center is
// unrestricted; nil System requests source snapshots, not a default system.
type Options struct {
	Condition  search.Condition
	Comparison Comparison
	Center     spelling.Spelling
	System     *system.System
}

// Analysis is an immutable analysis snapshot. Its zero value is unset.
// Assignment can share only private immutable data; use the public musical
// results rather than Go equality to compare analyses.
type Analysis struct{ data *analysisData }
type analysisData struct {
	notes    []note.Note
	contexts []placeContext
	scope    template.Catalog
	opts     Options
	phrase   bool
	groups   []Group
	observed []Tone
}
type placeContext struct {
	sys *system.System
	err error
}

func cloneSystem(sys *system.System) *system.System {
	if sys == nil {
		return nil
	}
	copy := *sys
	return &copy
}
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// Classical selects the seven classical diatonic rotations in degree order.
func Classical() template.Catalog {
	return template.MustNewCatalog(template.IonianEntry(), template.DorianEntry(), template.PhrygianEntry(), template.LydianEntry(), template.MixolydianEntry(), template.AeolianEntry(), template.LocrianEntry())
}

// Pentatonic selects the major and minor pentatonic entries in that order.
func Pentatonic() template.Catalog {
	return template.MustNewCatalog(template.PentatonicMajorEntry(), template.PentatonicMinorEntry())
}

// AnalyzeSet preserves every input record for trial replacement, without
// interpreting input order as adjacency. An empty catalog is a valid scope.
func AnalyzeSet(notes []note.Note, scope template.Catalog, opts Options) (Analysis, error) {
	return analyze(notes, scope, opts, false)
}

// MustAnalyzeSet panics with the error returned by AnalyzeSet.
func MustAnalyzeSet(notes []note.Note, scope template.Catalog, opts Options) Analysis {
	return must(AnalyzeSet(notes, scope, opts))
}

// AnalyzePhrase additionally retains adjacency between consecutive places.
func AnalyzePhrase(notes []note.Note, scope template.Catalog, opts Options) (Analysis, error) {
	return analyze(notes, scope, opts, true)
}

// MustAnalyzePhrase panics with the error returned by AnalyzePhrase.
func MustAnalyzePhrase(notes []note.Note, scope template.Catalog, opts Options) Analysis {
	return must(AnalyzePhrase(notes, scope, opts))
}
func analyze(notes []note.Note, scope template.Catalog, opts Options, phrase bool) (Analysis, error) {
	if len(notes) == 0 {
		return Analysis{}, ErrEmptyInput
	}
	if !opts.Condition.IsValid() || (opts.Comparison != BySpelling && opts.Comparison != ByPosition) {
		return Analysis{}, ErrInvalidArgument
	}
	if opts.System != nil && !opts.System.IsValid() {
		return Analysis{}, errors.Join(ErrUnset, system.ErrInvalid)
	}
	for j, n := range notes {
		if !n.IsValid() {
			return Analysis{}, fmt.Errorf("place %d: %w: %w", j, ErrUnset, note.ErrUnset)
		}
	}
	for entry := range scope.All() {
		m, err := mode.FromEntry(entry)
		if err != nil {
			return Analysis{}, fmt.Errorf("entry %s: %w: %w", entry.Code(), ErrInvalidArgument, err)
		}
		if !m.IsClosed() || !m.IsOctave() {
			return Analysis{}, fmt.Errorf("entry %s: %w", entry.Code(), ErrInvalidArgument)
		}
	}
	contexts, effective, contextErr := readContexts(notes, opts.System)
	if opts.Comparison == ByPosition && contextErr != nil {
		return Analysis{}, contextErr
	}
	opts.System = effective
	detached := slices.Clone(notes)
	for j := range detached {
		if err := detached[j].Unbind(); err != nil {
			return Analysis{}, err
		}
	}
	return buildAnalysis(detached, contexts, scope, opts, phrase)
}
func readContexts(notes []note.Note, explicit *system.System) ([]placeContext, *system.System, error) {
	contexts := make([]placeContext, len(notes))
	if explicit != nil {
		sys := cloneSystem(explicit)
		for j := range contexts {
			contexts[j].sys = sys
		}
		return contexts, sys, nil
	}
	var sources []note.Note
	var cached []placeContext
	var common *system.System
	var commonErr error
	for j, n := range notes {
		var current placeContext
		found := false
		for k, previous := range sources {
			if n.SameSource(previous) {
				current = cached[k]
				found = true
				break
			}
		}
		if !found {
			set, err := n.Settings()
			switch {
			case errors.Is(err, note.ErrNoSource):
				current.err = errors.Join(ErrNoSystem, err)
			case err != nil:
				current.err = errors.Join(ErrContext, err)
			default:
				current.sys = cloneSystem(set.System())
			}
			if n.IsBound() {
				sources = append(sources, n)
				cached = append(cached, current)
			}
		}
		if current.err != nil {
			current.err = fmt.Errorf("place %d: %w", j, current.err)
			if commonErr == nil {
				commonErr = current.err
			}
		} else if common == nil {
			common = current.sys
		} else if !common.Equal(current.sys) && commonErr == nil {
			commonErr = errors.Join(ErrDifferentSystems, note.ErrDifferentSystems)
		}
		contexts[j] = current
	}
	if commonErr != nil {
		return contexts, nil, commonErr
	}
	return contexts, common, nil
}
func buildAnalysis(notes []note.Note, contexts []placeContext, scope template.Catalog, opts Options, phrase bool) (Analysis, error) {
	d := &analysisData{notes: notes, contexts: contexts, scope: scope, opts: opts, phrase: phrase}
	a := Analysis{d}
	names := make([]spelling.Spelling, len(notes))
	for j, n := range notes {
		names[j] = n.Spelling()
	}
	var err error
	d.observed, err = a.tones(names)
	if err != nil {
		return Analysis{}, err
	}
	searcher := search.MustNew(&scope)
	readings := make([]Reading, 0)
	if opts.Comparison == BySpelling {
		matches, err := searcher.BySpelling(notes, opts.Condition)
		if err != nil {
			return Analysis{}, err
		}
		for _, match := range matches {
			if opts.Center.IsValid() && !opts.Center.Equal(match.Tonic()) {
				continue
			}
			r, err := makeReading(match.Mode(), match.Tonic(), opts, match.Tonic().String())
			if err != nil {
				return Analysis{}, err
			}
			readings = append(readings, r)
		}
	} else {
		var centerClass string
		if opts.Center.IsValid() {
			class, err := opts.System.Class(opts.Center)
			if err != nil {
				return Analysis{}, errors.Join(ErrUninterpretableName, err)
			}
			centerClass = strconv.Itoa(class.Number())
		}
		matches, err := searcher.ByPosition(opts.System, notes, opts.Condition)
		if err != nil {
			return Analysis{}, err
		}
		for _, match := range matches {
			f := match.Family()
			class := strconv.Itoa(f.Class().Number())
			if opts.Center.IsValid() && class != centerClass {
				continue
			}
			l, fixed := f.Letter()
			identity := class + ":" + strconv.FormatBool(fixed) + ":" + l.String()
			r, err := makeReading(match.Mode(), match.Tonic(), opts, identity)
			if err != nil {
				return Analysis{}, err
			}
			readings = append(readings, r)
		}
	}
	groupIndices := make(map[string]int)
	for _, r := range readings {
		tones, err := a.tones(r.degrees)
		if err != nil {
			return Analysis{}, err
		}
		key := toneListKey(tones)
		index, ok := groupIndices[key]
		if !ok {
			index = len(d.groups)
			groupIndices[key] = index
			d.groups = append(d.groups, Group{})
		}
		g := &d.groups[index]
		g.Tones = mergeTones(g.Tones, tones)
		g.Readings = append(g.Readings, r)
	}
	observed := toneMap(d.observed)
	for j := range d.groups {
		g := &d.groups[j]
		for k := range g.Tones {
			if t, ok := observed[keyOf(g.Tones[k])]; ok {
				g.Tones[k].Names = unionNames(g.Tones[k].Names, t.Names)
			} else {
				g.Missing = append(g.Missing, cloneTone(g.Tones[k]))
			}
		}
	}
	return a, nil
}

// IsValid reports whether a was produced by a successful analysis.
func (a Analysis) IsValid() bool { return a.data != nil }

// Notes returns independent records, retaining register and cents but no source.
func (a Analysis) Notes() []note.Note {
	if !a.IsValid() {
		return nil
	}
	return slices.Clone(a.data.notes)
}

// Scope returns the immutable catalog snapshot, including an empty scope.
func (a Analysis) Scope() template.Catalog {
	if !a.IsValid() {
		return template.Catalog{}
	}
	return a.data.scope
}

// IsPhrase distinguishes an ordered phrase from an unordered set.
func (a Analysis) IsPhrase() bool { return a.IsValid() && a.data.phrase }

// Options returns effective options and an independent system value, if known.
func (a Analysis) Options() Options {
	if !a.IsValid() {
		return Options{}
	}
	o := a.data.opts
	o.System = cloneSystem(o.System)
	return o
}

// Reading preserves the actual catalog entry, displayed tonic, degree names,
// and the comparison context. Copies share only private immutable state.
type Reading struct {
	entry      template.Entry
	tonic      spelling.Spelling
	degrees    []spelling.Spelling
	comparison Comparison
	sys        *system.System
	identity   string
}

func makeReading(m mode.Mode, tonic spelling.Spelling, opts Options, tonicIdentity string) (Reading, error) {
	entry, _ := m.Origin()
	r := Reading{entry: entry, tonic: tonic, comparison: opts.Comparison, sys: opts.System, identity: entry.Code() + "|" + tonicIdentity}
	for _, i := range m.DegreeIntervals() {
		name, err := interval.Transpose(tonic, i)
		if err != nil {
			return Reading{}, err
		}
		r.degrees = append(r.degrees, name)
	}
	return r, nil
}

// Entry returns the original immutable entry with all its names.
func (r Reading) Entry() template.Entry { return r.entry }

// Tonic returns the exact tonic or the search family's displayed representative.
func (r Reading) Tonic() spelling.Spelling { return r.tonic }

// Degrees returns an independent list in modal degree order, without closure.
func (r Reading) Degrees() []spelling.Spelling { return slices.Clone(r.degrees) }

// DegreesOf returns every matching degree, including for a newly supplied
// name. An unset reading/name is ErrUnset; positional interpretation errors
// preserve ErrUninterpretableName and the system cause.
func (r Reading) DegreesOf(name spelling.Spelling) ([]int, error) {
	if !r.entry.IsValid() || !name.IsValid() {
		return nil, ErrUnset
	}
	var classNumber int
	if r.comparison == ByPosition {
		class, err := r.sys.Class(name)
		if err != nil {
			return nil, errors.Join(ErrUninterpretableName, err)
		}
		classNumber = class.Number()
	}
	result := make([]int, 0)
	for j, degree := range r.degrees {
		same := degree.Equal(name)
		if r.comparison == ByPosition {
			class, err := r.sys.Class(degree)
			if err != nil {
				return nil, errors.Join(ErrUninterpretableName, err)
			}
			same = class.Number() == classNumber
		}
		if same {
			result = append(result, j+1)
		}
	}
	return result, nil
}

// MustDegreesOf panics with the error returned by DegreesOf.
func (r Reading) MustDegreesOf(name spelling.Spelling) []int { return must(r.DegreesOf(name)) }
