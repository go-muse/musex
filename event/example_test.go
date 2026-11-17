package event_test

import (
	"errors"
	"fmt"

	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/settings"
)

func ExampleEvent_WithMarks_grace() {
	normal := event.MustNew(
		event.MustNoteContent(note.MustParsePitch(settings.Standard(), "C4")),
		event.MustWrittenDuration(duration.Eighth.MustWritten()),
		event.Marks{},
	)
	marks := normal.MustMarks()
	marks.Grace = true
	grace := normal.MustWithMarks(marks)
	fmt.Println(normal.MustLength().Equal(grace.MustLength()))
	fmt.Println(grace.MustLength().IsZero())
	fmt.Println(grace.MustDuration().MustWritten().MustBase() == duration.Eighth)

	// Output:
	// false
	// true
	// true
}

func ExampleEvent_Length_measureRest() {
	rest := event.MustNew(event.RestContent(), event.MustMeasureDuration(2), event.Marks{})
	_, err := rest.Length()
	if errors.Is(err, event.ErrNeedsMeasureContext) {
		// The caller resolves these measures in its own grid.
		fmt.Println(rest.MustDuration().MustMeasureCount())
	}

	// Output:
	// 2
}

func ExampleEvent_WithContentAndMarks() {
	notation := settings.Standard()
	c := chord.MustNew(note.MustParsePitch(notation, "C4"), note.MustParsePitch(notation, "E4"))
	withChord := event.MustNew(
		event.MustChordContent(c),
		event.MustWrittenDuration(duration.Half.MustWritten()),
		event.Marks{Ties: []int{0}},
	)
	copyOfChord := withChord.MustContent().MustChord()
	copyOfChord.MustClear()
	fmt.Println(withChord.MustContent().MustLen())

	invalidMarks := withChord.MustMarks()
	invalidMarks.Ties[0] = -1
	_, err := withChord.WithMarks(invalidMarks)
	fmt.Println(errors.Is(err, event.ErrIndex))

	_, err = withChord.WithContent(event.RestContent())
	fmt.Println(errors.Is(err, event.ErrTiedNoteChanged))
	rest := withChord.MustWithContentAndMarks(event.RestContent(), event.Marks{})
	fmt.Println(rest.Kind() == event.KindRest, rest.MustDuration().Equal(withChord.MustDuration()))

	// Output:
	// 2
	// true
	// true
	// true true
}

func ExampleEvent_unset() {
	var unset *event.Event
	fmt.Println(unset.IsSet(), unset.Equal(unset))
	_, err := unset.Length()
	fmt.Println(errors.Is(err, event.ErrUnset))

	// Output:
	// false false
	// true
}
