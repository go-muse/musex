package chord_test

import (
	"testing"

	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/note"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T-7, allocations: architecture 5.14, the operations with "0" on the success
// path. testing.AllocsPerRun counts the allocations of all goroutines, so the
// test does not call t.Parallel.

// Package variables keep the results alive without allocations in the
// measured functions.
var (
	sinkChord chord.Chord
	sinkNote  note.Note
	sinkBool  bool
	sinkInt   int
	sinkErr   error
)

func TestAllocsZero(t *testing.T) {
	st := std()
	c := chord.MustNew(c4, pn(e4.Spelling(), 4, 0, st), g4, c4p30, cNo, c4)
	other := chord.MustNew(c4, pn(e4.Spelling(), 4, 0, st), g4, c4p30, cNo, c4)
	query := c4p30
	tests := []struct {
		name string
		f    func()
	}{
		{name: "copy by assignment", f: func() { sinkChord = c }},
		{name: "IsValid", f: func() { sinkBool = c.IsValid() }},
		{name: "IsEmpty", f: func() { sinkBool = c.IsEmpty() }},
		{name: "Len", f: func() { sinkInt = c.Len() }},
		{name: "At", f: func() { sinkNote, sinkErr = c.At(3) }},
		{name: "Contains", f: func() { sinkBool = c.Contains(query) }},
		{name: "Count", f: func() { sinkInt = c.Count(query) }},
		{name: "Find", f: func() { sinkInt, sinkNote, sinkBool = c.Find(query) }},
		{name: "ContainsRecord", f: func() { sinkBool = c.ContainsRecord(query) }},
		{name: "ContainsSpelling", f: func() { sinkBool = c.ContainsSpelling(query) }},
		{name: "HasNotesWithOctave", f: func() { sinkBool = c.HasNotesWithOctave() }},
		{name: "HasNotesWithoutOctave", f: func() { sinkBool = c.HasNotesWithoutOctave() }},
		{name: "EqualInOrder", f: func() { sinkBool = c.EqualInOrder(other) }},
		{name: "Clear", f: func() {
			sinkChord = c
			sinkErr = sinkChord.Clear()
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.f()
			require.NoError(t, sinkErr, "precondition: the success path")
			assert.Zero(t, testing.AllocsPerRun(100, tt.f), "allocations of %s", tt.name)
		})
	}
	require.True(t, c.EqualInOrder(other), "precondition: EqualInOrder compares all notes")
	require.True(t, c.Contains(query))
}
