// Must methods have the same behavior as their error-returning counterparts.
package mode

import (
	"iter"

	"github.com/go-muse/musex/degree"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/template"
)

func must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

func mustPair[T, U any](first T, second U, err error) (T, U) {
	if err != nil {
		panic(err)
	}
	return first, second
}

// MustFromTemplate is like FromTemplate but panics with its error.
func MustFromTemplate(t template.Template) Mode {
	return must(FromTemplate(t))
}

// MustFromEntry is like FromEntry but panics with its error.
func MustFromEntry(e template.Entry) Mode {
	return must(FromEntry(e))
}

// MustFromCode is like FromCode but panics with its error.
func MustFromCode(code string) Mode {
	return must(FromCode(code))
}

// MustFromCodeIn is like FromCodeIn but panics with its error.
func MustFromCodeIn(catalog *template.Catalog, code string) Mode {
	return must(FromCodeIn(catalog, code))
}

// MustFromName is like FromName but panics with its error.
func MustFromName(name string) Mode {
	return must(FromName(name))
}

// MustFromNameIn is like FromNameIn but panics with its error.
func MustFromNameIn(catalog *template.Catalog, name string) Mode {
	return must(FromNameIn(catalog, name))
}

// MustClosedFromIntervals is like ClosedFromIntervals but panics with its error.
func MustClosedFromIntervals(adjacent []interval.Interval) Mode {
	return must(ClosedFromIntervals(adjacent))
}

// MustOpenFromIntervals is like OpenFromIntervals but panics with its error.
func MustOpenFromIntervals(adjacent []interval.Interval) Mode {
	return must(OpenFromIntervals(adjacent))
}

// MustClosedFromDegreeIntervals is like ClosedFromDegreeIntervals but panics with its error.
func MustClosedFromDegreeIntervals(degrees []interval.Interval, cycle interval.Interval) Mode {
	return must(ClosedFromDegreeIntervals(degrees, cycle))
}

// MustOpenFromDegreeIntervals is like OpenFromDegreeIntervals but panics with its error.
func MustOpenFromDegreeIntervals(degrees []interval.Interval) Mode {
	return must(OpenFromDegreeIntervals(degrees))
}

// MustClosedFromNotes is like ClosedFromNotes but panics with its error.
func MustClosedFromNotes(notes []note.Note, cycle interval.Interval) Mode {
	return must(ClosedFromNotes(notes, cycle))
}

// MustFromNotesInOctave is like FromNotesInOctave but panics with its error.
func MustFromNotesInOctave(notes []note.Note) Mode {
	return must(FromNotesInOctave(notes))
}

// MustOpenFromNotes is like OpenFromNotes but panics with its error.
func MustOpenFromNotes(notes []note.Note) Mode {
	return must(OpenFromNotes(notes))
}

// MustFromNumbers is like FromNumbers but panics with its error.
func MustFromNumbers(sys *system.System, adjacent []system.Step) Mode {
	return must(FromNumbers(sys, adjacent))
}

// MustFromNumbersFrom is like FromNumbersFrom but panics with its error.
func MustFromNumbersFrom(sys *system.System, adjacent []system.Step, start spelling.Spelling) Mode {
	return must(FromNumbersFrom(sys, adjacent, start))
}

// MustFromNumbersWithPlan is like FromNumbersWithPlan but panics with its error.
func MustFromNumbersWithPlan(sys *system.System, adjacent []system.Step, plan LetterPlan) Mode {
	return must(FromNumbersWithPlan(sys, adjacent, plan))
}

// MustFromNumbersFromWithPlan is like FromNumbersFromWithPlan but panics with its error.
func MustFromNumbersFromWithPlan(sys *system.System, adjacent []system.Step, start spelling.Spelling, plan LetterPlan) Mode {
	return must(FromNumbersFromWithPlan(sys, adjacent, start, plan))
}

// MustParse is like Parse but panics with its error.
func MustParse(text string) Mode {
	return must(Parse(text))
}

// MustFromDegree is like FromDegree but panics with its error.
func MustFromDegree(source Mode, first int) Mode {
	return must(FromDegree(source, first))
}

// MustRotations is like Rotations but panics with its error.
func MustRotations(source Mode) iter.Seq[Mode] {
	return must(Rotations(source))
}

// MustDegree is like Degree but panics with its error.
func (m Mode) MustDegree(number int) degree.Degree {
	return must(m.Degree(number))
}

// MustDegrees is like Degrees but panics with its error.
func (m Mode) MustDegrees() []degree.Degree {
	return must(m.Degrees())
}

// MustDegreeInterval is like DegreeInterval but panics with its error.
func (m Mode) MustDegreeInterval(number int) interval.Interval {
	return must(m.DegreeInterval(number))
}

// MustAdjacentIntervalAfter is like AdjacentIntervalAfter but panics with its error.
func (m Mode) MustAdjacentIntervalAfter(number int) interval.Interval {
	return must(m.AdjacentIntervalAfter(number))
}

// MustBetweenDegrees is like BetweenDegrees but panics with its error.
func (m Mode) MustBetweenDegrees(from, to int) interval.Interval {
	return must(m.BetweenDegrees(from, to))
}

// MustSegment is like Segment but panics with its error.
func (m Mode) MustSegment(from, to int) []interval.Interval {
	return must(m.Segment(from, to))
}

// MustNext is like Next but panics with its error.
func (m Mode) MustNext(number int) (degree.Degree, bool) {
	return mustPair(m.Next(number))
}

// MustPrevious is like Previous but panics with its error.
func (m Mode) MustPrevious(number int) (degree.Degree, bool) {
	return mustPair(m.Previous(number))
}

// MustShift is like Shift but panics with its error.
func (m Mode) MustShift(number, distance int) (degree.Degree, int) {
	return mustPair(m.Shift(number, distance))
}

// MustDistanceUp is like DistanceUp but panics with its error.
func (m Mode) MustDistanceUp(from, to int) int {
	return must(m.DistanceUp(from, to))
}

// MustDistanceDown is like DistanceDown but panics with its error.
func (m Mode) MustDistanceDown(from, to int) int {
	return must(m.DistanceDown(from, to))
}

// MustShortestDistance is like ShortestDistance but panics with its error.
func (m Mode) MustShortestDistance(from, to int) int {
	return must(m.ShortestDistance(from, to))
}

// MustNotesFrom is like NotesFrom but panics with its error.
func (m Mode) MustNotesFrom(first note.Note) []note.Note {
	return must(m.NotesFrom(first))
}

// MustNotesFromIn is like NotesFromIn but panics with its error.
func (m Mode) MustNotesFromIn(sys *system.System, first note.Note) []note.Note {
	return must(m.NotesFromIn(sys, first))
}

// MustDegreesFrom is like DegreesFrom but panics with its error.
func (m Mode) MustDegreesFrom(first note.Note) []degree.Degree {
	return must(m.DegreesFrom(first))
}

// MustDegreesFromIn is like DegreesFromIn but panics with its error.
func (m Mode) MustDegreesFromIn(sys *system.System, first note.Note) []degree.Degree {
	return must(m.DegreesFromIn(sys, first))
}

// MustNoteFrom is like NoteFrom but panics with its error.
func (m Mode) MustNoteFrom(first note.Note, number, cycle int) note.Note {
	return must(m.NoteFrom(first, number, cycle))
}

// MustNoteFromIn is like NoteFromIn but panics with its error.
func (m Mode) MustNoteFromIn(sys *system.System, first note.Note, number, cycle int) note.Note {
	return must(m.NoteFromIn(sys, first, number, cycle))
}

// MustKeySignatureFrom is like KeySignatureFrom but panics with its error.
func (m Mode) MustKeySignatureFrom(first spelling.Spelling) (Signature, bool) {
	return mustPair(m.KeySignatureFrom(first))
}

// MustFirstSpellingForSignature is like FirstSpellingForSignature but panics with its error.
func (m Mode) MustFirstSpellingForSignature(index int) spelling.Spelling {
	return must(m.FirstSpellingForSignature(index))
}

// MustDifference is like Difference but panics with its error.
func (m Mode) MustDifference(other Mode) Difference {
	return must(m.Difference(other))
}

// MustMismatchingDegrees is like MismatchingDegrees but panics with its error.
func (m Mode) MustMismatchingDegrees(other Mode, measure Measure) Mismatches {
	return must(m.MismatchingDegrees(other, measure))
}

// MustSameInSystem is like SameInSystem but panics with its error.
func (m Mode) MustSameInSystem(sys *system.System, other Mode) bool {
	return must(m.SameInSystem(sys, other))
}

// MustSameInSystemFrom is like SameInSystemFrom but panics with its error.
func (m Mode) MustSameInSystemFrom(sys *system.System, start spelling.Spelling, other Mode, otherStart spelling.Spelling) bool {
	return must(m.SameInSystemFrom(sys, start, other, otherStart))
}

// MustInclination is like Inclination but panics with its error.
func (m Mode) MustInclination() Inclination {
	return must(m.Inclination())
}

// MustIsDiatonic is like IsDiatonic but panics with its error.
func (m Mode) MustIsDiatonic() bool {
	return must(m.IsDiatonic())
}

// MustOnlyDiatonicIntervals is like OnlyDiatonicIntervals but panics with its error.
func (m Mode) MustOnlyDiatonicIntervals() bool {
	return must(m.OnlyDiatonicIntervals())
}

// MustBrightness is like Brightness but panics with its error.
func (m Mode) MustBrightness() (int, bool) {
	return mustPair(m.Brightness())
}

// MustCompareBrightness is like CompareBrightness but panics with its error.
func (m Mode) MustCompareBrightness(other Mode) int {
	return must(m.CompareBrightness(other))
}

// MustIntervals is like Intervals but panics with its error.
func (m Mode) MustIntervals() iter.Seq[IntervalAt] {
	return must(m.Intervals())
}

// MustDegreesForInterval is like DegreesForInterval but panics with its error.
func (m Mode) MustDegreesForInterval(want interval.Interval) []degree.Degree {
	return must(m.DegreesForInterval(want))
}

// MustChromaticIntervals is like ChromaticIntervals but panics with its error.
func (m Mode) MustChromaticIntervals() iter.Seq[IntervalAt] {
	return must(m.ChromaticIntervals())
}

// MustTetrachords is like Tetrachords but panics with its error.
func (m Mode) MustTetrachords() TetrachordPair {
	return must(m.Tetrachords())
}

// MustPeriod is like Period but panics with its error.
func (m Mode) MustPeriod() int {
	return must(m.Period())
}

// MustDistinctRotations is like DistinctRotations but panics with its error.
func (m Mode) MustDistinctRotations() int {
	return must(m.DistinctRotations())
}

// MustMinimalPart is like MinimalPart but panics with its error.
func (m Mode) MustMinimalPart() ([]interval.Interval, int) {
	return mustPair(m.MinimalPart())
}

// MustNumbers is like Numbers but panics with its error.
func (m Mode) MustNumbers(sys *system.System) Positions {
	return must(m.Numbers(sys))
}

// MustNumbersFrom is like NumbersFrom but panics with its error.
func (m Mode) MustNumbersFrom(sys *system.System, start spelling.Spelling) Positions {
	return must(m.NumbersFrom(sys, start))
}

// MustCycleIn is like CycleIn but panics with its error.
func (m Mode) MustCycleIn(sys *system.System) CycleInfo {
	return must(m.CycleIn(sys))
}

// MustCycleInFrom is like CycleInFrom but panics with its error.
func (m Mode) MustCycleInFrom(sys *system.System, start spelling.Spelling) CycleInfo {
	return must(m.CycleInFrom(sys, start))
}

// MustSymmetryIn is like SymmetryIn but panics with its error.
func (m Mode) MustSymmetryIn(sys *system.System) Symmetry {
	return must(m.SymmetryIn(sys))
}

// MustSymmetryInFrom is like SymmetryInFrom but panics with its error.
func (m Mode) MustSymmetryInFrom(sys *system.System, start spelling.Spelling) Symmetry {
	return must(m.SymmetryInFrom(sys, start))
}

// MustNewRepository is like NewRepository but panics with its error.
func MustNewRepository(catalog *template.Catalog) Repository {
	return must(NewRepository(catalog))
}

// MustIdentify is like Identify but panics with its error.
func (r Repository) MustIdentify(m Mode) []template.Entry {
	return must(r.Identify(m))
}

// MustIdentifyRotations is like IdentifyRotations but panics with its error.
func (r Repository) MustIdentifyRotations(m Mode) []Rotation {
	return must(r.IdentifyRotations(m))
}

// MustRelated is like Related but panics with its error.
func (r Repository) MustRelated(m Mode, maxDifferences int, measure Measure, rotations bool) []Relation {
	return must(r.Related(m, maxDifferences, measure, rotations))
}

// MustIdentifyInSystem is like IdentifyInSystem but panics with its error.
func (r Repository) MustIdentifyInSystem(m Mode, sys *system.System) []template.Entry {
	return must(r.IdentifyInSystem(m, sys))
}

// MustIdentifyInSystemFrom is like IdentifyInSystemFrom but panics with its error.
func (r Repository) MustIdentifyInSystemFrom(m Mode, sys *system.System, start spelling.Spelling) []template.Entry {
	return must(r.IdentifyInSystemFrom(m, sys, start))
}

// MustIdentifyRotationsInSystem is like IdentifyRotationsInSystem but panics with its error.
func (r Repository) MustIdentifyRotationsInSystem(m Mode, sys *system.System) []Rotation {
	return must(r.IdentifyRotationsInSystem(m, sys))
}

// MustIdentifyRotationsInSystemFrom is like IdentifyRotationsInSystemFrom but panics with its error.
func (r Repository) MustIdentifyRotationsInSystemFrom(m Mode, sys *system.System, start spelling.Spelling) []Rotation {
	return must(r.IdentifyRotationsInSystemFrom(m, sys, start))
}

// MustRebuild is like Rebuild but panics with its error.
func (m *Mode) MustRebuild(t template.Template) *Mode {
	if err := m.Rebuild(t); err != nil {
		panic(err)
	}
	return m
}

// MustRebuildFromEntry is like RebuildFromEntry but panics with its error.
func (m *Mode) MustRebuildFromEntry(e template.Entry) *Mode {
	if err := m.RebuildFromEntry(e); err != nil {
		panic(err)
	}
	return m
}

// MustRebuildFromCode is like RebuildFromCode but panics with its error.
func (m *Mode) MustRebuildFromCode(code string) *Mode {
	if err := m.RebuildFromCode(code); err != nil {
		panic(err)
	}
	return m
}

// MustRebuildFromCodeIn is like RebuildFromCodeIn but panics with its error.
func (m *Mode) MustRebuildFromCodeIn(catalog *template.Catalog, code string) *Mode {
	if err := m.RebuildFromCodeIn(catalog, code); err != nil {
		panic(err)
	}
	return m
}

// MustRebuildFromName is like RebuildFromName but panics with its error.
func (m *Mode) MustRebuildFromName(name string) *Mode {
	if err := m.RebuildFromName(name); err != nil {
		panic(err)
	}
	return m
}

// MustRebuildFromNameIn is like RebuildFromNameIn but panics with its error.
func (m *Mode) MustRebuildFromNameIn(catalog *template.Catalog, name string) *Mode {
	if err := m.RebuildFromNameIn(catalog, name); err != nil {
		panic(err)
	}
	return m
}

// MustRotate is like Rotate but panics with its error.
func (m *Mode) MustRotate(first int) *Mode {
	if err := m.Rotate(first); err != nil {
		panic(err)
	}
	return m
}

// MustSetDegreeInterval is like SetDegreeInterval but panics with its error.
func (m *Mode) MustSetDegreeInterval(number int, value interval.Interval) *Mode {
	if err := m.SetDegreeInterval(number, value); err != nil {
		panic(err)
	}
	return m
}

// MustAddDegreeInterval is like AddDegreeInterval but panics with its error.
func (m *Mode) MustAddDegreeInterval(value interval.Interval) *Mode {
	if err := m.AddDegreeInterval(value); err != nil {
		panic(err)
	}
	return m
}

// MustRemoveDegree is like RemoveDegree but panics with its error.
func (m *Mode) MustRemoveDegree(number int) *Mode {
	if err := m.RemoveDegree(number); err != nil {
		panic(err)
	}
	return m
}

// MustMirror is like Mirror but panics with its error.
func (m *Mode) MustMirror() *Mode {
	if err := m.Mirror(); err != nil {
		panic(err)
	}
	return m
}

// MustRepeat is like Repeat but panics with its error.
func (m *Mode) MustRepeat(times int) *Mode {
	if err := m.Repeat(times); err != nil {
		panic(err)
	}
	return m
}

// MustReduceToMinimalPart is like ReduceToMinimalPart but panics with its error.
func (m *Mode) MustReduceToMinimalPart() *Mode {
	if err := m.ReduceToMinimalPart(); err != nil {
		panic(err)
	}
	return m
}

// MustIntervalsString is like IntervalsString but panics with its error.
func (m Mode) MustIntervalsString() string {
	return must(m.IntervalsString())
}

// MustRussianIntervals is like RussianIntervals but panics with its error.
func (m Mode) MustRussianIntervals() (string, bool) {
	return mustPair(m.RussianIntervals())
}
