package score_test

import (
	"testing"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/meter"
	"github.com/go-muse/musex/score"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempomap"
	"github.com/go-muse/musex/track"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// K-9, K-21, K-24: strict replacement keeps events in their original bars.
func TestChangeMeterProposal(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name                  string
		an, ad, bn, bd, count int
		markN, markD          int
		targetN, targetD      int
		movedN, movedD        int
	}{
		{"common four quarters", 4, 4, 4, 4, 2, 1, 1, 3, 4, 3, 4},
		{"equal lengths different notation", 3, 4, 6, 8, 2, 3, 4, 2, 4, 1, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, b := emptyPart(tc.an, tc.ad, tc.count), emptyPart(tc.bn, tc.bd, tc.count)
			e := tone(a, "C4", 1, 4, event.Marks{})
			a = write(a, 1, e)
			a = write(a, 2, tone(a, "D4", 1, 4, event.Marks{}))
			clock := clock60().MustWithChange(tempomap.MustNewMain(at(tc.markN, tc.markD), bpm(120))).MustWithExtension(tempomap.MustNewExtension(at(tc.markN, tc.markD), tempo.MustNewSeconds(1, 1)))
			s, ids := assemble(clock, a, b)
			proposal, err := s.ChangeMeter(1, 1, meter.MustNew(tc.targetN, tc.targetD))
			require.NoError(t, err)
			require.NotNil(t, proposal.Score)
			require.Len(t, proposal.Parts, 2)
			require.Len(t, proposal.Marks, 3)
			assertKeys(t, ids, proposal.Score.MustParts())
			for i, change := range proposal.Parts {
				assertKey(t, ids[i], change.Part)
				assert.True(t, change.Before.Equal(s.MustPart(ids[i])))
				assert.True(t, change.After.Equal(proposal.Score.MustPart(ids[i])))
				assertPos(t, pos(tc.movedN, tc.movedD), change.After.MustGrid().MustBarStart(2))
				assert.True(t, change.After.MustGrid().MustBarMeter(2).Equal(s.MustPart(ids[i]).MustGrid().MustBarMeter(2)))
			}
			assertPos(t, pos(tc.markN, tc.markD), s.MustPart(ids[0]).MustGrid().MustBarStart(2))
			assert.True(t, s.MustMap().Equal(clock))
			expectedClock := clock60().MustWithChange(tempomap.MustNewMain(at(tc.movedN, tc.movedD), bpm(120))).MustWithExtension(tempomap.MustNewExtension(at(tc.movedN, tc.movedD), tempo.MustNewSeconds(1, 1)))
			assert.True(t, proposal.Score.MustMap().Equal(expectedClock))
			assertEvent(t, e, proposal.Score.MustPart(ids[0]).MustEntries(a.MustVoices()[0], 1)[0].Event)
			assert.Equal(t, 2, proposal.Score.MustPart(ids[0]).MustEntries(a.MustVoices()[0], 2)[0].Place.Bar)
			assertPos(t, pos(tc.movedN, tc.movedD), proposal.Score.MustPart(ids[0]).MustEntries(a.MustVoices()[0], 2)[0].Time.Start)
			assert.Equal(t, score.TempoMark, proposal.Marks[0].Mark.Kind)
			assertPos(t, pos(0, 1), proposal.Marks[0].From)
			assertPos(t, pos(0, 1), proposal.Marks[0].To)
			for _, move := range proposal.Marks[1:] {
				assertPos(t, pos(tc.markN, tc.markD), move.From)
				assertPos(t, pos(tc.movedN, tc.movedD), move.To)
				require.Len(t, move.Bindings, 2)
				for i, binding := range move.Bindings {
					assertKey(t, ids[i], binding.Part)
					assert.Nil(t, binding.AfterBar)
					assert.Equal(t, 2, binding.Address.MustBar())
					assertPos(t, pos(0, 1), binding.Address.MustOffset())
					assertPos(t, move.To, binding.Position)
				}
			}
			assert.Equal(t, score.ExtensionMark, proposal.Marks[2].Mark.Kind)
			assert.True(t, proposal.Marks[2].Mark.Extension.MustSeconds().Equal(tempo.MustNewSeconds(1, 1)))
		})
	}
}

func TestChangeMeterOverflowIsAtomicAndComplete(t *testing.T) {
	t.Parallel()
	for _, quarters := range []int{3, 4} {
		t.Run(string(rune('0'+quarters))+" quarters", func(t *testing.T) {
			p := emptyPart(4, 4, 2)
			for _, pitch := range []string{"C4", "D4", "E4", "F4"}[:quarters] {
				p = p.MustSplice(p.MustVoices()[0], 1, len(p.MustEntries(p.MustVoices()[0], 1)), 0, tone(p, pitch, 1, 4, event.Marks{}))
			}
			s, ids := assemble(clock60(), p)
			proposal, err := s.ChangeMeter(1, 1, meter.MustNew(3, 4))
			if quarters == 3 {
				require.NoError(t, err)
				assert.Len(t, proposal.Score.MustPart(ids[0]).MustEntries(p.MustVoices()[0], 1), 3)
			} else {
				require.ErrorIs(t, err, track.ErrOverflow)
				assert.Zero(t, proposal)
				i := problems(t, err).Issues[0]
				assertKey(t, ids[0], i.Part)
				assert.Equal(t, 1, i.Place.Bar)
				require.NotNil(t, i.Detail)
				assert.True(t, i.Detail.Amount.Equal(dur(1, 4)))
			}
			assert.True(t, s.MustPart(ids[0]).Equal(p))
			assert.Empty(t, s.MustPart(ids[0]).MustEntries(p.MustVoices()[0], 2))
		})
	}
	// Every invalid part is reported, rather than only the first overflow.
	a, b := emptyPart(4, 4, 2), emptyPart(4, 4, 2)
	a = write(a, 1, tone(a, "C4", 1, 2, event.Marks{}), tone(a, "D4", 1, 2, event.Marks{}))
	b = write(b, 1, tone(b, "E4", 1, 1, event.Marks{}))
	b = write(b, 2, tone(b, "F4", 1, 1, event.Marks{}))
	s, ids := assemble(clock60(), a, b)
	proposal, err := s.ChangeMeter(1, 2, meter.MustNew(3, 4))
	require.ErrorIs(t, err, track.ErrOverflow)
	assert.Zero(t, proposal)
	issues := problems(t, err).Issues
	require.Len(t, issues, 3)
	assertKey(t, ids[0], issues[0].Part)
	assertKey(t, ids[1], issues[1].Part)
	assertKey(t, ids[1], issues[2].Part)
	assert.Equal(t, 1, issues[1].Place.Bar)
	assert.Equal(t, 2, issues[2].Place.Bar)
}

// K-10, K-22, K-23: a mark needs a real address and one common candidate.
func TestChangeMeterMapFailures(t *testing.T) {
	t.Parallel()
	t.Run("new address outside shortened bar", func(t *testing.T) {
		clock := clock60().MustWithChange(tempomap.MustNewMain(at(7, 8), bpm(90))).MustWithExtension(tempomap.MustNewExtension(at(7, 8), tempo.ZeroSeconds()))
		s, ids := assemble(clock, emptyPart(4, 4, 2), emptyPart(4, 4, 2))
		proposal, err := s.ChangeMeter(1, 1, meter.MustNew(3, 4))
		require.ErrorIs(t, err, score.ErrMapAddress)
		assert.Zero(t, proposal)
		issues := problems(t, err).Issues
		require.Len(t, issues, 4)
		for i, issue := range issues {
			assertKey(t, ids[i%2], issue.Part)
			assertPos(t, pos(7, 8), issue.Position)
			require.NotNil(t, issue.Mark)
			assert.Equal(t, []score.MarkKind{score.TempoMark, score.TempoMark, score.ExtensionMark, score.ExtensionMark}[i], issue.Mark.Kind)
			require.Len(t, issue.Bindings, 1)
			assert.False(t, issue.Bindings[0].Position.IsSet())
		}
		assert.True(t, s.MustMap().Equal(clock))
	})
	t.Run("two candidates", func(t *testing.T) {
		clock := clock60().MustWithChange(tempomap.MustNewMain(at(1, 1), bpm(120)))
		s, ids := assemble(clock, emptyPart(4, 4, 3), emptyPart(3, 4, 3))
		proposal, err := s.ChangeMeter(1, 1, meter.MustNew(2, 4))
		require.ErrorIs(t, err, score.ErrMapConflict)
		assert.Zero(t, proposal)
		issues := problems(t, err).Issues
		require.Len(t, issues, 1)
		issue := issues[0]
		assert.Zero(t, issue.Part)
		assertPos(t, pos(1, 1), issue.Position)
		require.NotNil(t, issue.Mark)
		require.Len(t, issue.Bindings, 2)
		for i, want := range []duration.Position{pos(1, 2), pos(3, 4)} {
			assertKey(t, ids[i], issue.Bindings[i].Part)
			assert.Equal(t, 2, issue.Bindings[i].Address.MustBar())
			assertPos(t, want, issue.Bindings[i].Position)
		}
		assertPos(t, pos(0, 1), issue.Bindings[0].Address.MustOffset())
		assertPos(t, pos(1, 4), issue.Bindings[1].Address.MustOffset())
		assert.True(t, s.MustMap().Equal(clock))
		// A diagnostic is owned by its caller, including optional mark data.
		*issue.Mark = score.MapMark{}
		issue.Bindings[0].Position = duration.Position{}
		_, repeatedErr := s.ChangeMeter(1, 1, meter.MustNew(2, 4))
		repeated := problems(t, repeatedErr).Issues[0]
		require.NotNil(t, repeated.Mark)
		assert.Equal(t, score.TempoMark, repeated.Mark.Kind)
		assertPos(t, pos(1, 2), repeated.Bindings[0].Position)
	})
	for _, n := range []int{3, 5} {
		t.Run(pos(n, 2).String()+" old address", func(t *testing.T) {
			clock := clock60().MustWithChange(tempomap.MustNewMain(at(n, 2), bpm(90)))
			s, ids := assemble(clock, emptyPart(4, 4, 1), emptyPart(4, 4, 2))
			proposal, err := s.ChangeMeter(1, 1, meter.MustNew(3, 4))
			require.ErrorIs(t, err, score.ErrMapAddress)
			assert.Zero(t, proposal)
			issues := problems(t, err).Issues
			expected := 1
			if n == 5 {
				expected = 2
			}
			require.Len(t, issues, expected)
			for i, issue := range issues {
				assertKey(t, ids[i], issue.Part)
				assertPos(t, pos(n, 2), issue.Position)
			}
			assert.True(t, s.MustMap().Equal(clock))
		})
	}
}

func TestChangeMeterRequiresRangeInEveryPart(t *testing.T) {
	t.Parallel()
	a, b := emptyPart(4, 4, 2), emptyPart(4, 4, 1)
	s, ids := assemble(clock60(), a, b)
	proposal, err := s.ChangeMeter(1, 2, meter.MustNew(3, 4))
	require.Error(t, err)
	assert.Zero(t, proposal)
	issues := problems(t, err).Issues
	assertKey(t, ids[1], issues[0].Part)
	assert.True(t, s.MustPart(ids[0]).Equal(a))
	assert.True(t, s.MustPart(ids[1]).Equal(b))
}

func TestChangeMeterTerminalAndStructuralRecords(t *testing.T) {
	t.Parallel()
	t.Run("terminal hold binding", func(t *testing.T) {
		clock := clock60().MustWithExtension(tempomap.MustNewExtension(at(2, 1), tempo.MustNewSeconds(1, 1)))
		s, ids := assemble(clock, emptyPart(4, 4, 2), emptyPart(4, 4, 2))
		proposal := s.MustChangeMeter(1, 1, meter.MustNew(3, 4))
		require.Len(t, proposal.Marks, 2)
		move := proposal.Marks[1]
		assertPos(t, pos(7, 4), move.To)
		for i, binding := range move.Bindings {
			assertKey(t, ids[i], binding.Part)
			require.NotNil(t, binding.AfterBar)
			assert.Equal(t, 2, *binding.AfterBar)
			assert.False(t, binding.Address.IsSet())
		}
	})
	t.Run("same position with different binding forms", func(t *testing.T) {
		clock := clock60().MustWithExtension(tempomap.MustNewExtension(at(1, 1), tempo.MustNewSeconds(1, 1)))
		s, _ := assemble(clock, emptyPart(4, 4, 1), emptyPart(4, 4, 2))
		proposal := s.MustChangeMeter(1, 1, meter.MustNew(3, 4))
		bindings := proposal.Marks[1].Bindings
		require.Len(t, bindings, 2)
		require.NotNil(t, bindings[0].AfterBar)
		assert.Equal(t, 1, *bindings[0].AfterBar)
		assert.Nil(t, bindings[1].AfterBar)
		assert.Equal(t, 2, bindings[1].Address.MustBar())
		for _, b := range bindings {
			assertPos(t, pos(3, 4), b.Position)
		}
	})
	t.Run("multibar rest rejects mixed meters atomically", func(t *testing.T) {
		p := emptyPart(4, 4, 3)
		p = write(p, 1, rest(3, true))
		p = p.MustEdit(track.Edit{Barlines: []track.BarlineEdit{{AfterBar: 2, Fermata: true}}})
		newGrid := p.MustGrid().MustWithMetersIn(1, 2, meter.MustNew(3, 4))
		_, dependencyErr := p.Edit(track.Edit{Grid: &newGrid})
		require.ErrorIs(t, dependencyErr, track.ErrMeasureRest)
		var dependencyProblems *track.Problems
		require.ErrorAs(t, dependencyErr, &dependencyProblems)
		require.NotEmpty(t, dependencyProblems.Issues)
		s, ids := assemble(clock60(), p)
		proposal, err := s.ChangeMeter(1, 2, meter.MustNew(3, 4))
		require.ErrorIs(t, err, track.ErrMeasureRest)
		assert.Zero(t, proposal)
		issue := problems(t, err).Issues[0]
		assertKey(t, ids[0], issue.Part)
		require.NotNil(t, issue.Detail)
		assert.Equal(t, dependencyProblems.Issues[0], *issue.Detail)
		assertKey(t, dependencyProblems.Issues[0].Place, issue.Place)
		assertKey(t, p.MustVoices()[0], issue.Place.Voice)
		assert.GreaterOrEqual(t, issue.Place.Bar, 1)
		assert.LessOrEqual(t, issue.Place.Bar, 3)
		assert.True(t, s.MustPart(ids[0]).Equal(p))
		assert.True(t, s.MustMap().Equal(clock60()))
		assert.Equal(t, []int{2}, s.MustPart(ids[0]).MustFermataBars())
		assertPos(t, pos(3, 1), s.MustPart(ids[0]).MustEntries(p.MustVoices()[0], 1)[0].Time.End)
	})
	t.Run("multibar rest and fermata retain a uniform meter", func(t *testing.T) {
		p := emptyPart(4, 4, 3)
		p = write(p, 1, rest(3, true))
		p = p.MustEdit(track.Edit{Barlines: []track.BarlineEdit{{AfterBar: 2, Fermata: true}}})
		s, ids := assemble(clock60(), p)
		changed := s.MustChangeMeter(1, 3, meter.MustNew(3, 4)).Score.MustPart(ids[0])
		assert.Equal(t, []int{2}, changed.MustFermataBars())
		assert.Equal(t, 3, changed.MustEntries(p.MustVoices()[0], 1)[0].Event.MustDuration().MustMeasureCount())
		assertPos(t, pos(9, 4), changed.MustEntries(p.MustVoices()[0], 1)[0].Time.End)
		assert.True(t, changed.MustEntries(p.MustVoices()[0], 1)[0].Event.MustMarks().Fermata)
		assert.True(t, s.MustPart(ids[0]).Equal(p))
	})
	t.Run("tie cannot be broken by inserting empty space", func(t *testing.T) {
		p := emptyPart(3, 4, 2)
		p = p.MustEdit(track.Edit{Bars: []track.BarEdit{{Voice: p.MustVoices()[0], Bar: 1, Events: []*event.Event{tone(p, "C4", 3, 4, event.Marks{Ties: []int{0}})}}, {Voice: p.MustVoices()[0], Bar: 2, Events: []*event.Event{tone(p, "C4", 1, 4, event.Marks{})}}}})
		s, ids := assemble(clock60(), p)
		proposal, err := s.ChangeMeter(1, 1, meter.MustNew(4, 4))
		require.Error(t, err)
		assert.Zero(t, proposal)
		assertKey(t, ids[0], problems(t, err).Issues[0].Part)
		assert.True(t, s.MustPart(ids[0]).Equal(p))
	})
}
