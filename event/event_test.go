package event_test

import (
	"math"
	"reflect"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/settings"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertMarks(t *testing.T, want, got event.Marks) {
	t.Helper()

	if want.Ties == nil {
		want.Ties = []int{}
	}
	assert.Equal(t, want, got)
}

func TestNewEventForms(t *testing.T) {
	t.Parallel()

	c := pitchNote("C4")
	e := pitchNote("E4")
	pitched := event.MustNoteContent(c)
	unpitched := event.MustNoteContent(note.MustParse(settings.Standard(), "C"))
	ch := chordContent(c, e)
	rest := event.RestContent()
	dotted := event.MustWrittenDuration(duration.MustNewWritten(duration.Quarter, 1))
	exact := valueDuration(3, 8)
	measures := event.MustMeasureDuration(3)

	cases := []struct {
		name    string
		content event.Content
		kind    event.Kind
		d       event.Duration
		dk      event.DurationKind
		length  fraction.Fraction
	}{
		{"dotted quarter note", pitched, event.KindNote, dotted, event.DurationWritten, fraction.MustNew(3, 8)},
		{"same note by value", pitched, event.KindNote, exact, event.DurationValue, fraction.MustNew(3, 8)},
		{"note without octave", unpitched, event.KindNote, dotted, event.DurationWritten, fraction.MustNew(3, 8)},
		{"written chord", ch, event.KindChord, writtenDuration(duration.Half), event.DurationWritten, fraction.MustNew(1, 2)},
		{"chord by value", ch, event.KindChord, exact, event.DurationValue, fraction.MustNew(3, 8)},
		{"written whole rest", rest, event.KindRest, writtenDuration(duration.Whole), event.DurationWritten, fraction.One()},
		{"rest by value", rest, event.KindRest, valueDuration(1, 3), event.DurationValue, fraction.MustNew(1, 3)},
		{"triplet half rest", rest, event.KindRest, event.MustWrittenDuration(duration.MustParseWritten("2[3:2]")), event.DurationWritten, fraction.MustNew(1, 3)},
		{"one measure rest", rest, event.KindRest, event.MustMeasureDuration(1), event.DurationMeasures, fraction.Fraction{}},
		{"three measure rest", rest, event.KindRest, measures, event.DurationMeasures, fraction.Fraction{}},
		{"maximum measure rest", rest, event.KindRest, event.MustMeasureDuration(math.MaxInt), event.DurationMeasures, fraction.Fraction{}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			marks := event.Marks{Fermata: true}
			e, err := event.New(tt.content, tt.d, marks)
			require.NoError(t, err)
			require.NotNil(t, e)
			assert.True(t, e.IsSet())
			assert.Equal(t, tt.kind, e.Kind())

			content, err := e.Content()
			require.NoError(t, err)
			assert.True(t, tt.content.Equal(content))
			assert.True(t, tt.content.Equal(e.MustContent()))

			d, err := e.Duration()
			require.NoError(t, err)
			assert.Equal(t, tt.dk, d.Kind())
			assert.True(t, tt.d.Equal(d))
			assert.True(t, tt.d.Equal(e.MustDuration()))

			gotMarks, err := e.Marks()
			require.NoError(t, err)
			assertMarks(t, marks, gotMarks)
			assertMarks(t, marks, e.MustMarks())
			other := event.MustNew(tt.content, tt.d, marks)
			require.NotNil(t, other)
			assert.NotSame(t, e, other)
			assert.True(t, e.Equal(other))

			if tt.dk == event.DurationMeasures {
				assert.Equal(t, tt.d.MustMeasureCount(), d.MustMeasureCount())
				assertFailure(t, event.ErrNeedsMeasureContext, e.Length, e.MustLength)
			} else {
				length, err := e.Length()
				require.NoError(t, err)
				assertFraction(t, tt.length, length)
				assertFraction(t, tt.length, e.MustLength())
			}
			tied, err := e.TiedNotes()
			require.NoError(t, err)
			assertNotes(t, nil, tied)
			assertNotes(t, nil, e.MustTiedNotes())
		})
	}
}

func TestGraceLengthKeepsPositiveWrittenValue(t *testing.T) {
	t.Parallel()

	c := pitchNote("C4")
	marks := event.Marks{Grace: true, Fermata: true, Ties: []int{0}}
	eighth := writtenDuration(duration.Eighth)
	grace := event.MustNew(event.MustNoteContent(c), eighth, marks)
	assertMarks(t, marks, grace.MustMarks())
	assertFraction(t, fraction.Zero(), grace.MustLength())
	assertFraction(t, fraction.MustNew(1, 8), grace.MustDuration().MustValue().MustFraction())
	assert.True(t, duration.Eighth.MustWritten().Equal(grace.MustDuration().MustWritten()))

	main := event.MustNew(event.MustNoteContent(c), eighth, event.Marks{})
	assertFraction(t, fraction.MustNew(1, 8), main.MustLength())

	chordGrace := event.MustNew(chordContent(pitchNote("G4"), c),
		writtenDuration(duration.Sixteenth),
		event.Marks{Grace: true, Fermata: true, Arpeggio: true, Ties: []int{1}})
	assertFraction(t, fraction.Zero(), chordGrace.MustLength())
	assertFraction(t, fraction.MustNew(1, 16), chordGrace.MustDuration().MustValue().MustFraction())
	assert.True(t, grace.MustEqualLength(chordGrace))
	assert.False(t, grace.Equal(chordGrace))
	assert.False(t, grace.MustEqualLength(main))
	assert.Equal(t, -1, grace.MustCompareLength(main))
}

func TestEventLengthComparison(t *testing.T) {
	t.Parallel()

	n := event.MustNoteContent(pitchNote("C4"))
	pitched := event.MustNew(n, event.MustWrittenDuration(duration.MustParseWritten("4.")), event.Marks{Fermata: true, Ties: []int{0}})
	exactNote := event.MustNew(n, valueDuration(6, 16), event.Marks{Fermata: true, Ties: []int{0}})
	exactRest := event.MustNew(event.RestContent(), valueDuration(3, 8), event.Marks{})
	lower := event.MustNew(event.RestContent(), valueDuration(1, 3), event.Marks{})
	exactThird := event.MustNew(event.RestContent(), event.MustWrittenDuration(duration.MustParseWritten("2[3:2]")), event.Marks{})
	nearlyWhole := event.MustNew(event.RestContent(), valueDuration(math.MaxInt-1, math.MaxInt), event.Marks{})
	slightlyShorter := event.MustNew(event.RestContent(), valueDuration(math.MaxInt-2, math.MaxInt-1), event.Marks{})
	assert.False(t, pitched.Equal(exactNote), "written and numeric note records differ")
	assert.False(t, lower.Equal(exactThird), "written and numeric rest records differ")

	for _, tt := range []struct {
		name string
		a, b *event.Event
		cmp  int
	}{
		{"same length with different duration forms", pitched, exactNote, 0},
		{"same length with different content and marks", pitched, exactRest, 0},
		{"same length with grouped notation", lower, exactThird, 0},
		{"greater", pitched, lower, 1},
		{"less", lower, pitched, -1},
		{"nearby fractions are compared exactly", nearlyWhole, slightlyShorter, 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.a.CompareLength(tt.b)
			require.NoError(t, err)
			assert.Equal(t, tt.cmp, got)
			assert.Equal(t, tt.cmp, tt.a.MustCompareLength(tt.b))
			equal, err := tt.a.EqualLength(tt.b)
			require.NoError(t, err)
			assert.Equal(t, tt.cmp == 0, equal)
			assert.Equal(t, tt.cmp == 0, tt.a.MustEqualLength(tt.b))
		})
	}

	measure := event.MustNew(event.RestContent(), event.MustMeasureDuration(1), event.Marks{})
	for _, tt := range []struct {
		name string
		a, b *event.Event
		err  error
	}{
		{"measure compared to itself", measure, measure, event.ErrNeedsMeasureContext},
		{"left needs measure context", measure, pitched, event.ErrNeedsMeasureContext},
		{"right needs measure context", pitched, measure, event.ErrNeedsMeasureContext},
		{"unset right takes precedence over left measure context", measure, nil, event.ErrUnset},
		{"zero right takes precedence over left measure context", measure, new(event.Event), event.ErrUnset},
		{"unset left takes precedence over right measure context", nil, measure, event.ErrUnset},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assertFailure(t, tt.err,
				func() (int, error) { return tt.a.CompareLength(tt.b) },
				func() int { return tt.a.MustCompareLength(tt.b) })
			assertFailure(t, tt.err,
				func() (bool, error) { return tt.a.EqualLength(tt.b) },
				func() bool { return tt.a.MustEqualLength(tt.b) })
		})
	}
}

func TestEventEqualityDistinguishesTheWrittenRecord(t *testing.T) {
	t.Parallel()

	c, e, g := pitchNote("C4"), pitchNote("E4"), pitchNote("G4")
	first := chordContent(c, e, c, g)
	base := writtenDuration(duration.Half)
	marks := event.Marks{Ties: []int{2, 0, 2}, Fermata: true, Arpeggio: true}
	a := event.MustNew(first, base, marks)
	b := event.MustNew(chordContent(pitchNote("C4"), e, c, g), base,
		event.Marks{Ties: []int{0, 2}, Fermata: true, Arpeggio: true})
	copyOfA := *a
	assert.False(t, reflect.TypeFor[event.Event]().Comparable())
	assert.False(t, reflect.TypeFor[event.Marks]().Comparable())
	assert.True(t, a.Equal(a))
	assert.True(t, a.Equal(b))
	assert.True(t, b.Equal(a))
	assert.True(t, b.Equal(&copyOfA))
	assert.True(t, a.Equal(&copyOfA))
	assert.Equal(t, 3, len(map[*event.Event]struct{}{a: {}, b: {}, &copyOfA: {}}))

	for _, tt := range []struct {
		name    string
		content event.Content
		d       event.Duration
		marks   event.Marks
	}{
		{"note order", chordContent(c, g, c, e), base, marks},
		{"note multiplicity", chordContent(c, e, c, g, g), base, marks},
		{"written base", first, writtenDuration(duration.Quarter), marks},
		{"written dots", first, event.MustWrittenDuration(duration.MustParseWritten("2.")), marks},
		{"duration form", first, valueDuration(1, 2), marks},
		{"tie indices", first, base, event.Marks{Ties: []int{0}, Fermata: true, Arpeggio: true}},
		{"fermata", first, base, event.Marks{Ties: []int{0, 2}, Arpeggio: true}},
		{"arpeggio", first, base, event.Marks{Ties: []int{0, 2}, Fermata: true}},
		{"grace", first, base, event.Marks{Ties: []int{0, 2}, Fermata: true, Arpeggio: true, Grace: true}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			changed := event.MustNew(tt.content, tt.d, tt.marks)
			assert.False(t, a.Equal(changed))
			assert.False(t, changed.Equal(a))
		})
	}

	assert.False(t, event.MustNew(event.MustNoteContent(c), base, event.Marks{}).
		Equal(event.MustNew(chordContent(c), base, event.Marks{})))
	assert.False(t, event.MustNew(event.RestContent(), event.MustMeasureDuration(1), event.Marks{}).
		Equal(event.MustNew(event.RestContent(), event.MustMeasureDuration(3), event.Marks{})))
	assert.False(t, event.MustNew(event.RestContent(), valueDuration(1, 2), event.Marks{}).
		Equal(event.MustNew(event.RestContent(), valueDuration(1, 3), event.Marks{})))

	adjusted := c
	adjusted.MustSetCents(cents.MustFromInt(25))
	for _, tt := range []struct {
		name string
		a, b note.Note
	}{
		{"enharmonic spellings", pitchNote("C♯4"), pitchNote("D♭4")},
		{"same pitch with different cents", c, adjusted},
	} {
		t.Run(tt.name, func(t *testing.T) {
			a := event.MustNew(event.MustNoteContent(tt.a), base, event.Marks{})
			b := event.MustNew(event.MustNoteContent(tt.b), base, event.Marks{})
			assert.False(t, a.Equal(b))
			assert.True(t, a.MustEqualLength(b))
		})
	}
}

func TestEventEqualityWithGroupedDurations(t *testing.T) {
	t.Parallel()

	content := event.MustNoteContent(pitchNote("C4"))
	marks := event.Marks{Ties: []int{0}, Fermata: true}
	for _, tt := range []struct {
		name       string
		first      string
		second     string
		length     fraction.Fraction
		sameRecord bool
	}{
		{"same group chain", "2[3:2,5:4]", "2[3:2,5:4]", fraction.MustNew(4, 15), true},
		{"different group order", "2[3:2,5:4]", "2[5:4,3:2]", fraction.MustNew(4, 15), false},
		{"equivalent group ratio", "2[3:2]", "2[6:4]", fraction.MustNew(1, 3), false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			a := event.MustNew(content, event.MustWrittenDuration(duration.MustParseWritten(tt.first)), marks)
			b := event.MustNew(content, event.MustWrittenDuration(duration.MustParseWritten(tt.second)), marks)
			assert.Equal(t, tt.sameRecord, a.Equal(b))
			assert.Equal(t, tt.sameRecord, b.Equal(a))
			assertFraction(t, tt.length, a.MustLength())
			assertFraction(t, tt.length, b.MustLength())

			sameLength, err := a.EqualLength(b)
			require.NoError(t, err)
			assert.True(t, sameLength)
			order, err := a.CompareLength(b)
			require.NoError(t, err)
			assert.Zero(t, order)
		})
	}
}
