package scale

import (
	"iter"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/frequency"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pcset"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"
)

func must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}
func mustPair[A, B any](a A, b B, err error) (A, B) {
	if err != nil {
		panic(err)
	}
	return a, b
}

// MustNew calls New and panics with its error.
func MustNew(notes ...note.Note) *Scale {
	return must(New(notes...))
}

// MustAt calls At and panics with its error.
func (s *Scale) MustAt(index int) note.Note {
	return must(s.At(index))
}

// MustFirst calls First and panics with its error.
func (s *Scale) MustFirst() note.Note {
	return must(s.First())
}

// MustLast calls Last and panics with its error.
func (s *Scale) MustLast() note.Note {
	return must(s.Last())
}

// MustReplace calls Replace and panics with its error.
func (s *Scale) MustReplace(notes ...note.Note) *Scale {
	if err := s.Replace(notes...); err != nil {
		panic(err)
	}
	return s
}

// MustSet calls Set and panics with its error.
func (s *Scale) MustSet(index int, n note.Note) *Scale {
	if err := s.Set(index, n); err != nil {
		panic(err)
	}
	return s
}

// MustInsert calls Insert and panics with its error.
func (s *Scale) MustInsert(index int, notes ...note.Note) *Scale {
	if err := s.Insert(index, notes...); err != nil {
		panic(err)
	}
	return s
}

// MustAppend calls Append and panics with its error.
func (s *Scale) MustAppend(notes ...note.Note) *Scale {
	if err := s.Append(notes...); err != nil {
		panic(err)
	}
	return s
}

// MustRemove calls Remove and panics with its error.
func (s *Scale) MustRemove(index int) *Scale {
	if err := s.Remove(index); err != nil {
		panic(err)
	}
	return s
}

// MustDelete calls Delete and panics with its error.
func (s *Scale) MustDelete(begin, end int) *Scale {
	if err := s.Delete(begin, end); err != nil {
		panic(err)
	}
	return s
}

// MustClear calls Clear and panics with its error.
func (s *Scale) MustClear() *Scale {
	if err := s.Clear(); err != nil {
		panic(err)
	}
	return s
}

// MustMove calls Move and panics with its error.
func (s *Scale) MustMove(index, distance int) Element {
	return must(s.Move(index, distance))
}

// MustNext calls Next and panics with its error.
func (s *Scale) MustNext(index int) Element {
	return must(s.Next(index))
}

// MustPrevious calls Previous and panics with its error.
func (s *Scale) MustPrevious(index int) Element {
	return must(s.Previous(index))
}

// MustDistanceForward calls DistanceForward and panics with its error.
func (s *Scale) MustDistanceForward(from, to int) int {
	return must(s.DistanceForward(from, to))
}

// MustDistanceBackward calls DistanceBackward and panics with its error.
func (s *Scale) MustDistanceBackward(from, to int) int {
	return must(s.DistanceBackward(from, to))
}

// MustMirrorIndex calls MirrorIndex and panics with its error.
func (s *Scale) MustMirrorIndex(index, axis int) Element {
	return must(s.MirrorIndex(index, axis))
}

// MustIter calls Iter and panics with its error.
func (s *Scale) MustIter(start, count int, direction Direction) iter.Seq[note.Note] {
	return must(s.Iter(start, count, direction))
}

// MustRound calls Round and panics with its error.
func (s *Scale) MustRound(start int, direction Direction) iter.Seq[note.Note] {
	return must(s.Round(start, direction))
}

// MustCycle calls Cycle and panics with its error.
func (s *Scale) MustCycle(start int, direction Direction) iter.Seq[note.Note] {
	return must(s.Cycle(start, direction))
}

// MustTake calls Take and panics with its error.
func (s *Scale) MustTake(start, count int, direction Direction) *Scale {
	return must(s.Take(start, count, direction))
}

// MustOneRound calls OneRound and panics with its error.
func (s *Scale) MustOneRound(start int, direction Direction) *Scale {
	return must(s.OneRound(start, direction))
}

// MustMoveFrom calls MoveFrom and panics with its error.
func (s *Scale) MustMoveFrom(from Selection, distance int, by ...Comparison) Element {
	return must(s.MoveFrom(from, distance, by...))
}

// MustShiftMelody calls ShiftMelody and panics with its error.
func (s *Scale) MustShiftMelody(notes []note.Note, distance int, by ...Comparison) *Scale {
	return must(s.ShiftMelody(notes, distance, by...))
}

// MustMirrorMelody calls MirrorMelody and panics with its error.
func (s *Scale) MustMirrorMelody(notes []note.Note, axis int, by ...Comparison) *Scale {
	return must(s.MirrorMelody(notes, axis, by...))
}

// MustShiftSelected calls ShiftSelected and panics with its error.
func (s *Scale) MustShiftSelected(notes []Selection, distance int, by ...Comparison) *Scale {
	return must(s.ShiftSelected(notes, distance, by...))
}

// MustMirrorSelected calls MirrorSelected and panics with its error.
func (s *Scale) MustMirrorSelected(notes []Selection, axis int, by ...Comparison) *Scale {
	return must(s.MirrorSelected(notes, axis, by...))
}

// MustByPositionIn calls ByPositionIn and panics with its error.
func MustByPositionIn(sys *system.System) Comparison {
	return must(ByPositionIn(sys))
}

// MustBySoundIn calls BySoundIn and panics with its error.
func MustBySoundIn(t *tuning.Tuning) Comparison {
	return must(BySoundIn(t))
}

// MustFind calls Find and panics with its error.
func (s *Scale) MustFind(n note.Note, by ...Comparison) []int {
	return must(s.Find(n, by...))
}

// MustContains calls Contains and panics with its error.
func (s *Scale) MustContains(n note.Note, by ...Comparison) bool {
	return must(s.Contains(n, by...))
}

// MustCount calls Count and panics with its error.
func (s *Scale) MustCount(n note.Note, by ...Comparison) int {
	return must(s.Count(n, by...))
}

// MustFindFirst calls FindFirst and panics with its error.
func (s *Scale) MustFindFirst(n note.Note, by ...Comparison) (int, bool) {
	return mustPair(s.FindFirst(n, by...))
}

// MustFindLast calls FindLast and panics with its error.
func (s *Scale) MustFindLast(n note.Note, by ...Comparison) (int, bool) {
	return mustPair(s.FindLast(n, by...))
}

// MustEqualBy calls EqualBy and panics with its error.
func (s *Scale) MustEqualBy(other *Scale, by Comparison) bool {
	return must(s.EqualBy(other, by))
}

// MustFindSound calls FindSound and panics with its error.
func (s *Scale) MustFindSound(n note.Note, c frequency.Criterion) []int {
	return must(s.FindSound(n, c))
}

// MustFindSoundInTuning calls FindSoundInTuning and panics with its error.
func (s *Scale) MustFindSoundInTuning(t *tuning.Tuning, n note.Note, c frequency.Criterion) []int {
	return must(s.FindSoundInTuning(t, n, c))
}

// MustSameSet calls SameSet and panics with its error.
func (s *Scale) MustSameSet(other *Scale, by ...Comparison) bool {
	return must(s.SameSet(other, by...))
}

// MustSameMultiset calls SameMultiset and panics with its error.
func (s *Scale) MustSameMultiset(other *Scale, by ...Comparison) bool {
	return must(s.SameMultiset(other, by...))
}

// MustIsSubsetOf calls IsSubsetOf and panics with its error.
func (s *Scale) MustIsSubsetOf(other *Scale, by ...Comparison) bool {
	return must(s.IsSubsetOf(other, by...))
}

// MustIntersection calls Intersection and panics with its error.
func (s *Scale) MustIntersection(other *Scale, by ...Comparison) *Scale {
	return must(s.Intersection(other, by...))
}

// MustDifference calls Difference and panics with its error.
func (s *Scale) MustDifference(other *Scale, by ...Comparison) *Scale {
	return must(s.Difference(other, by...))
}

// MustUnion calls Union and panics with its error.
func (s *Scale) MustUnion(other *Scale, by ...Comparison) *Scale {
	return must(s.Union(other, by...))
}

// MustUnique calls Unique and panics with its error.
func (s *Scale) MustUnique(by ...Comparison) *Scale {
	return must(s.Unique(by...))
}

// MustRotationStarts calls RotationStarts and panics with its error.
func (s *Scale) MustRotationStarts(other *Scale, by ...Comparison) []int {
	return must(s.RotationStarts(other, by...))
}

// MustIsRotationOf calls IsRotationOf and panics with its error.
func (s *Scale) MustIsRotationOf(other *Scale, by ...Comparison) bool {
	return must(s.IsRotationOf(other, by...))
}

// MustRotate calls Rotate and panics with its error.
func (s *Scale) MustRotate(first ...int) *Scale {
	return must(s.Rotate(first...))
}

// MustReverse calls Reverse and panics with its error.
func (s *Scale) MustReverse() *Scale {
	return must(s.Reverse())
}

// MustSlice calls Slice and panics with its error.
func (s *Scale) MustSlice(begin, end int) *Scale {
	return must(s.Slice(begin, end))
}

// MustRepeat calls Repeat and panics with its error.
func (s *Scale) MustRepeat(times int) *Scale {
	return must(s.Repeat(times))
}

// MustWithoutOctaves calls WithoutOctaves and panics with its error.
func (s *Scale) MustWithoutOctaves() *Scale {
	return must(s.WithoutOctaves())
}

// MustShiftOctaves calls ShiftOctaves and panics with its error.
func (s *Scale) MustShiftOctaves(count int) *Scale {
	return must(s.ShiftOctaves(count))
}

// MustTranspose calls Transpose and panics with its error.
func (s *Scale) MustTranspose(i interval.Interval) *Scale {
	return must(s.Transpose(i))
}

// MustTransposeIn calls TransposeIn and panics with its error.
func (s *Scale) MustTransposeIn(sys *system.System, i interval.Interval) *Scale {
	return must(s.TransposeIn(sys, i))
}

// MustTransposeTo calls TransposeTo and panics with its error.
func (s *Scale) MustTransposeTo(first note.Note, direction interval.Direction) *Scale {
	return must(s.TransposeTo(first, direction))
}

// MustTransposeToIn calls TransposeToIn and panics with its error.
func (s *Scale) MustTransposeToIn(sys *system.System, first note.Note, direction interval.Direction) *Scale {
	return must(s.TransposeToIn(sys, first, direction))
}

// MustEnharmonics calls Enharmonics and panics with its error.
func (s *Scale) MustEnharmonics(index int, limit alteration.Alteration) iter.Seq[note.Note] {
	return must(s.Enharmonics(index, limit))
}

// MustEnharmonicsIn calls EnharmonicsIn and panics with its error.
func (s *Scale) MustEnharmonicsIn(src settings.Source, index int, limit alteration.Alteration) iter.Seq[note.Note] {
	return must(s.EnharmonicsIn(src, index, limit))
}

// MustWithEnharmonic calls WithEnharmonic and panics with its error.
func (s *Scale) MustWithEnharmonic(index int, name spelling.Spelling) *Scale {
	return must(s.WithEnharmonic(index, name))
}

// MustWithEnharmonicIn calls WithEnharmonicIn and panics with its error.
func (s *Scale) MustWithEnharmonicIn(src settings.Source, index int, name spelling.Spelling) *Scale {
	return must(s.WithEnharmonicIn(src, index, name))
}

// MustPitchClasses calls PitchClasses and panics with its error.
func (s *Scale) MustPitchClasses(sys *system.System) pcset.Set {
	return must(s.PitchClasses(sys))
}

// MustKeySignature calls KeySignature and panics with its error.
func (s *Scale) MustKeySignature() (Signature, bool) {
	return mustPair(s.KeySignature())
}

// MustFormat calls Format and panics with its error.
func (s *Scale) MustFormat() string {
	return must(s.Format())
}

// MustFormatIn calls FormatIn and panics with its error.
func (s *Scale) MustFormatIn(src settings.Source) string {
	return must(s.FormatIn(src))
}
