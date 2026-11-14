package chord_test

import (
	"testing"

	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/note"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T-2. Membership, search and selection by a condition: ОП-8…ОП-14; П-3;
// К-7…К-11, К-32 (selection); architecture 5.7.

// query is one expectation of the search by one note.
type query struct {
	n       note.Note
	count   int
	indices []int
}

func TestMembershipByNote(t *testing.T) {
	t.Parallel()
	st := std()
	s19 := set19()
	c4st := pn(c4.Spelling(), 4, 0, st)
	c419 := pn(c4.Spelling(), 4, 0, s19)
	tests := []struct {
		name    string
		notes   []note.Note
		queries []query
	}{
		{
			name:  "K-7 C4 E4 C4",
			notes: list(c4, e4, c4),
			queries: []query{
				{n: c4, count: 2, indices: []int{0, 2}},
				{n: e4, count: 1, indices: []int{1}},
				{n: g4, count: 0, indices: []int{}},
			},
		},
		{
			name:  "K-8 C4 C4+5",
			notes: list(c4, c4p5),
			queries: []query{
				{n: c4, count: 1, indices: []int{0}},
				{n: c4p5, count: 1, indices: []int{1}},
				{n: c4p10, count: 0, indices: []int{}},
			},
		},
		{
			name:  "K-9 C C4",
			notes: list(cNo, c4),
			queries: []query{
				{n: cNo, count: 1, indices: []int{0}},
				{n: c4, count: 1, indices: []int{1}},
				{n: c5, count: 0, indices: []int{}},
			},
		},
		{
			name:  "P-3 a note bound to two sources is one note",
			notes: list(c4st, e4, c419, c4),
			queries: []query{
				{n: c4, count: 3, indices: []int{0, 2, 3}},
				{n: c419, count: 3, indices: []int{0, 2, 3}},
				{n: e4, count: 1, indices: []int{1}},
			},
		},
		{
			name:  "B♯4 and C4 differ",
			notes: list(bs4),
			queries: []query{
				{n: bs4, count: 1, indices: []int{0}},
				{n: c4, count: 0, indices: []int{}},
			},
		},
		{
			name:  "an unset query note",
			notes: list(c4, e4),
			queries: []query{
				{n: note.Note{}, count: 0, indices: []int{}},
			},
		},
		{
			name:  "empty chord",
			notes: list(),
			queries: []query{
				{n: c4, count: 0, indices: []int{}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := mk(t, tt.notes...)
			before := stateOf(c)
			for _, q := range tt.queries {
				assert.Equal(t, q.count > 0, c.Contains(q.n), "Contains(%v)", q.n)
				assert.Equal(t, q.count, c.Count(q.n), "Count(%v)", q.n)
				got := c.Indices(q.n)
				assert.NotNil(t, got, "Indices(%v) is not nil", q.n)
				assert.Equal(t, q.indices, got, "Indices(%v)", q.n)

				i, found, ok := c.Find(q.n)
				if q.count == 0 {
					assert.False(t, ok, "Find(%v)", q.n)
					assert.Equal(t, -1, i, "Find(%v)", q.n)
					assert.False(t, found.IsValid(), "Find(%v) gives the unset note", q.n)
					continue
				}
				assert.True(t, ok, "Find(%v)", q.n)
				assert.Equal(t, q.indices[0], i, "Find(%v): the first occurrence", q.n)
				assert.True(t, sameNote(tt.notes[q.indices[0]], found), "Find(%v) gives the note of the chord with its source: %v", q.n, found)
			}
			assertState(t, before, c, "reading does not change the chord")
		})
	}
}

func TestMembershipK10FindWithSource(t *testing.T) {
	t.Parallel()
	st := std()
	c := mk(t, pn(c4.Spelling(), 4, 0, st), e4)

	i, found, ok := c.Find(c4) // an unbound C4
	require.True(t, ok)
	assert.Equal(t, 0, i)
	assert.True(t, found.IsBound(), "the found note is the note of the chord")
	assert.True(t, found.SameSource(pn(c4.Spelling(), 4, 0, st)), "the found note is bound to the standard settings of the chord (note ОП-41)")
	assert.True(t, c.Contains(pn(c4.Spelling(), 4, 0, set19())), "contains C4 (19)")
}

func TestMembershipRecordAndSpelling(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name               string
		notes              []note.Note
		n                  note.Note
		record, spellingOK bool
	}{
		{name: "K-8 C4+10 by record", notes: list(c4, c4p5), n: c4p10, record: true, spellingOK: true},
		{name: "K-8 C without an octave", notes: list(c4, c4p5), n: cNo, record: false, spellingOK: true},
		{name: "K-8 C5", notes: list(c4, c4p5), n: c5, record: false, spellingOK: true},
		{name: "K-8 D4", notes: list(c4, c4p5), n: d4, record: false, spellingOK: false},
		{name: "K-9 C without an octave", notes: list(cNo, c4), n: cNo, record: true, spellingOK: true},
		{name: "K-9 C5", notes: list(cNo, c4), n: c5, record: false, spellingOK: true},
		{name: "only without an octave: C4", notes: list(cNo), n: c4, record: false, spellingOK: true},
		{name: "B♯4 and C4", notes: list(bs4), n: c4, record: false, spellingOK: false},
		{name: "an unset note", notes: list(c4), n: note.Note{}, record: false, spellingOK: false},
		{name: "empty chord", notes: list(), n: c4, record: false, spellingOK: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := mk(t, tt.notes...)
			assert.Equal(t, tt.record, c.ContainsRecord(tt.n), "ContainsRecord")
			assert.Equal(t, tt.spellingOK, c.ContainsSpelling(tt.n), "ContainsSpelling")
		})
	}
}

func TestMembershipAllAny(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		notes    []note.Note
		list     []note.Note
		all, any bool
	}{
		{name: "K-11 G4 C4", notes: list(c4, e4, g4), list: list(g4, c4), all: true, any: true},
		{name: "K-11 C4 C4", notes: list(c4, e4, g4), list: list(c4, c4), all: false, any: true},
		{name: "K-11 D4 E4", notes: list(c4, e4, g4), list: list(d4, e4), all: false, any: true},
		{name: "K-11 D4 F4", notes: list(c4, e4, g4), list: list(d4, f4), all: false, any: false},
		{name: "K-11 empty list", notes: list(c4, e4, g4), list: list(), all: true, any: false},
		{name: "K-11 C4 and an unset note", notes: list(c4, e4, g4), list: list(c4, note.Note{}), all: false, any: true},
		{name: "multiplicity: C4 C4 in C4 C4 E4", notes: list(c4, c4, e4), list: list(c4, c4), all: true, any: true},
		{name: "multiplicity: C4 C4 C4 in C4 C4 E4", notes: list(c4, c4, e4), list: list(c4, c4, c4), all: false, any: true},
		{name: "only unset notes", notes: list(c4), list: list(note.Note{}), all: false, any: false},
		{name: "corrections differ", notes: list(c4), list: list(c4p5), all: false, any: false},
		{name: "sources are not compared", notes: list(pn(c4.Spelling(), 4, 0, std())), list: list(pn(c4.Spelling(), 4, 0, set19())), all: true, any: true},
		{name: "empty chord, empty list", notes: list(), list: list(), all: true, any: false},
		{name: "empty chord", notes: list(), list: list(c4), all: false, any: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := mk(t, tt.notes...)
			in := stateNotes(tt.list)
			assert.Equal(t, tt.all, c.ContainsAll(tt.list...), "ContainsAll")
			assert.Equal(t, tt.any, c.ContainsAny(tt.list...), "ContainsAny")
			assertNotes(t, in, tt.list, "the list is not changed")
		})
	}
}

func TestMembershipUnsetChord(t *testing.T) {
	t.Parallel()
	var c chord.Chord // K-4
	assert.False(t, c.Contains(c4))
	assert.Zero(t, c.Count(c4))
	i, n, ok := c.Find(c4)
	assert.False(t, ok)
	assert.Equal(t, -1, i)
	assert.False(t, n.IsValid())
	got := c.Indices(c4)
	assert.NotNil(t, got)
	assert.Empty(t, got)
	assert.False(t, c.ContainsAll(), "the unset chord does not contain the empty list")
	assert.False(t, c.ContainsAll(c4))
	assert.False(t, c.ContainsAny(c4))
	assert.False(t, c.ContainsAny())
	assert.False(t, c.ContainsRecord(c4))
	assert.False(t, c.ContainsSpelling(c4))
}

func TestMembershipFunc(t *testing.T) {
	t.Parallel()
	// K-32: [C4, E, G4], the condition "the note has an octave".
	c := mk(t, c4, eNo, g4)
	before := stateOf(c)

	ok, err := c.ContainsFunc(hasOctave)
	require.NoError(t, err)
	assert.True(t, ok)

	n, err := c.CountFunc(hasOctave)
	require.NoError(t, err)
	assert.Equal(t, 2, n)

	idx, err := c.IndicesFunc(hasOctave)
	require.NoError(t, err)
	assert.Equal(t, []int{0, 2}, idx)

	notes, err := c.NotesFunc(hasOctave)
	require.NoError(t, err)
	assertNotes(t, list(c4, g4), notes)

	assert.True(t, c.MustContainsFunc(hasOctave))
	assert.Equal(t, 2, c.MustCountFunc(hasOctave))
	assert.Equal(t, []int{0, 2}, c.MustIndicesFunc(hasOctave))
	assertNotes(t, list(c4, g4), c.MustNotesFunc(hasOctave))

	assertState(t, before, c)
}

func TestMembershipFuncNoMatch(t *testing.T) {
	t.Parallel()
	cond := func(note.Note) (bool, error) { return false, nil }
	for name, c := range map[string]chord.Chord{
		"notes": mk(t, c4, e4),
		"empty": chord.MustNew(),
	} {
		ok, err := c.ContainsFunc(cond)
		require.NoError(t, err, name)
		assert.False(t, ok, name)

		n, err := c.CountFunc(cond)
		require.NoError(t, err, name)
		assert.Zero(t, n, name)

		idx, err := c.IndicesFunc(cond)
		require.NoError(t, err, name)
		assert.NotNil(t, idx, name)
		assert.Empty(t, idx, name)

		notes, err := c.NotesFunc(cond)
		require.NoError(t, err, name)
		assert.NotNil(t, notes, name)
		assert.Empty(t, notes, name)
	}
}

func TestMembershipFuncKeepsSources(t *testing.T) {
	t.Parallel()
	st := std()
	c4st := pn(c4.Spelling(), 4, 0, st)
	c := mk(t, c4st, e4, pn(c4.Spelling(), 4, 0, set19()))
	got, err := c.NotesFunc(func(n note.Note) (bool, error) { return n.Equal(c4), nil })
	require.NoError(t, err)
	assertNotes(t, list(c.Notes()[0], c.Notes()[2]), got, "the selected notes keep their sources")
}

func TestMembershipFuncUnsetChord(t *testing.T) {
	t.Parallel()
	var c chord.Chord // K-4: a set condition is not called; the results are zero
	calls := 0
	cond := func(note.Note) (bool, error) {
		calls++
		return true, nil
	}
	ok, err := c.ContainsFunc(cond)
	require.NoError(t, err)
	assert.False(t, ok)
	n, err := c.CountFunc(cond)
	require.NoError(t, err)
	assert.Zero(t, n)
	idx, err := c.IndicesFunc(cond)
	require.NoError(t, err)
	assert.NotNil(t, idx)
	assert.Empty(t, idx)
	notes, err := c.NotesFunc(cond)
	require.NoError(t, err)
	assert.NotNil(t, notes)
	assert.Empty(t, notes)
	assert.Zero(t, calls, "the condition is not called for the unset chord")

	assert.False(t, c.MustContainsFunc(cond))
	assert.Zero(t, c.MustCountFunc(cond))
	assert.Empty(t, c.MustIndicesFunc(cond))
	assert.Empty(t, c.MustNotesFunc(cond))
	assert.Zero(t, calls)
}

func TestMembershipFuncFailures(t *testing.T) {
	t.Parallel()
	failing := func(note.Note) (bool, error) { return false, errRule }
	tests := []struct {
		name   string
		c      chord.Chord
		cond   chord.Condition
		target error
	}{
		{name: "K-32 nil condition", c: mk(t, c4, eNo, g4), cond: nil, target: chord.ErrUnset},
		{name: "K-4 nil condition, unset chord", c: chord.Chord{}, cond: nil, target: chord.ErrUnset},
		{name: "nil condition, empty chord", c: chord.MustNew(), cond: nil, target: chord.ErrUnset},
		{name: "K-32 failing condition", c: mk(t, c4, eNo, g4), cond: failing, target: chord.ErrRuleFailed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			before := stateOf(tt.c)
			check := func(err error) {
				t.Helper()
				if tt.target == chord.ErrRuleFailed {
					assertRuleFailed(t, err, errRule)
					return
				}
				assertOnly(t, err, tt.target)
			}

			ok, err := tt.c.ContainsFunc(tt.cond)
			check(err)
			assert.False(t, ok, "ContainsFunc gives false on error")

			n, err := tt.c.CountFunc(tt.cond)
			check(err)
			assert.Zero(t, n, "CountFunc gives 0 on error")

			idx, err := tt.c.IndicesFunc(tt.cond)
			check(err)
			assert.Nil(t, idx, "IndicesFunc gives nil on error")

			notes, err := tt.c.NotesFunc(tt.cond)
			check(err)
			assert.Nil(t, notes, "NotesFunc gives nil on error")

			assertPanicsWith(t, tt.target, func() { tt.c.MustContainsFunc(tt.cond) })
			assertPanicsWith(t, tt.target, func() { tt.c.MustCountFunc(tt.cond) })
			assertPanicsWith(t, tt.target, func() { tt.c.MustIndicesFunc(tt.cond) })
			assertPanicsWith(t, tt.target, func() { tt.c.MustNotesFunc(tt.cond) })

			assertState(t, before, tt.c)
		})
	}
}

func TestMembershipFuncLiteral(t *testing.T) {
	t.Parallel()
	// Architecture 5.8: a function literal is passed without a conversion.
	c := mk(t, c4, eNo)
	n, err := c.CountFunc(func(n note.Note) (bool, error) { return !n.HasOctave(), nil })
	require.NoError(t, err)
	assert.Equal(t, 1, n)
}
