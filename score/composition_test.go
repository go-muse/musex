package score_test

import (
	"sync"
	"testing"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/meter"
	"github.com/go-muse/musex/score"
	"github.com/go-muse/musex/tempomap"
	"github.com/go-muse/musex/track"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// K-1, K-14: identities belong to participations, not musical records.
func TestCompositionSnapshots(t *testing.T) {
	t.Parallel()
	empty, err := score.New(clock60())
	require.NoError(t, err)
	assert.True(t, empty.IsSet())
	assert.Empty(t, empty.MustParts())
	assertPos(t, duration.Origin(), empty.MustEnd())
	a := emptyPart(4, 4, 2)
	a = write(a, 1, tone(a, "C4", 1, 4, event.Marks{}))
	b := emptyPart(4, 4, 1)
	b = write(b, 1, tone(b, "D4", 1, 4, event.Marks{}))
	s, ids := assemble(clock60(), a, b, a)
	assert.NotEqual(t, ids[0], ids[2])
	assert.NotEqual(t, score.PartID{}, ids[0])
	assert.Equal(t, ids, s.MustParts())
	assertPos(t, pos(2, 1), s.MustEnd())
	assert.True(t, s.MustPart(ids[0]).Equal(a))
	assert.Empty(t, empty.MustParts())
	reordered := s.MustReorder(ids[2], ids[1], ids[0])
	assert.True(t, reordered.Equal(s)) // Equal records can exchange identities.
	reordered = s.MustReorder(ids[1], ids[0], ids[2])
	assert.False(t, reordered.Equal(s))
	assert.Equal(t, ids, s.MustParts())
	replaced := s.MustReplace(ids[1], a)
	assert.Equal(t, ids, replaced.MustParts())
	assert.True(t, replaced.MustPart(ids[1]).Equal(a))
	assert.True(t, s.MustPart(ids[1]).Equal(b))
	removed := replaced.MustRemove(ids[0])
	assert.Equal(t, ids[1:], removed.MustParts())
	_, err = removed.Part(ids[0])
	require.ErrorIs(t, err, score.ErrPart)
	assert.True(t, s.MustPart(ids[0]).Equal(a))
	independent, _ := assemble(clock60(), a, b, a)
	assert.True(t, independent.Equal(s))
	changedClock := clock60().MustWithChange(tempomap.MustNewMain(at(3, 1), bpm(90)))
	assert.False(t, s.MustWithMap(changedClock).Equal(s))
	assert.True(t, s.MustMap().Equal(clock60()))
	for _, id := range ids {
		s = s.MustRemove(id)
	}
	assert.Empty(t, s.MustParts())
	assert.True(t, s.MustMap().Equal(clock60()))
}

func TestReplaceGridStarts(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name                  string
		original, replacement *track.Track
		fail                  bool
	}{
		{"same lengths different notation", emptyPart(3, 4, 2), emptyPart(6, 8, 2), false},
		{"append tail", emptyPart(4, 4, 2), emptyPart(4, 4, 3), false},
		{"truncate tail", emptyPart(4, 4, 3), emptyPart(4, 4, 2), false},
		{"last length", emptyPart(4, 4, 1), emptyPart(3, 4, 1), false},
		{"moved second start", emptyPart(4, 4, 2), emptyPart(3, 4, 2), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, ids := assemble(clock60(), tc.original)
			got, err := s.Replace(ids[0], tc.replacement)
			if tc.fail {
				require.ErrorIs(t, err, score.ErrGridShift)
				assert.Nil(t, got)
				i := problems(t, err).Issues[0]
				assert.Equal(t, ids[0], i.Part)
				assert.Equal(t, 2, i.Place.Bar)
				assert.Equal(t, -1, i.Place.Index)
				assertPos(t, pos(1, 1), i.Position)
				assertPos(t, pos(3, 4), i.Proposed)
			} else {
				require.NoError(t, err)
				assert.True(t, got.MustPart(ids[0]).Equal(tc.replacement))
			}
			assert.True(t, s.MustPart(ids[0]).Equal(tc.original))
		})
	}
	p := emptyPart(4, 4, 2)
	g := p.MustGrid().MustWithMetersIn(1, 1, meter.MustNew(3, 4))
	replacement := p.MustEdit(track.Edit{Grid: &g})
	s, ids := assemble(clock60(), p)
	_, err := s.Replace(ids[0], replacement)
	require.ErrorIs(t, err, score.ErrGridShift)
	pickupGrid := p.MustGrid().MustWithPickup(dur(1, 4))
	withPickup := p.MustEdit(track.Edit{Grid: &pickupGrid})
	s, ids = assemble(clock60(), withPickup)
	_, err = s.Replace(ids[0], p)
	require.ErrorIs(t, err, score.ErrGridShift)
	issue := problems(t, err).Issues[0]
	assert.Equal(t, 1, issue.Place.Bar)
	assertPos(t, pos(1, 4), issue.Position)
	assertPos(t, pos(0, 1), issue.Proposed)
}

func TestConcurrentDerivativesHaveDistinctIDs(t *testing.T) {
	t.Parallel()
	s := score.MustNew(clock60())
	p := emptyPart(4, 4, 1)
	const count = 12
	type result struct {
		s   *score.Score
		id  score.PartID
		err error
	}
	results := make([]result, count)
	var wg sync.WaitGroup
	for i := range count {
		wg.Go(func() { results[i].s, results[i].id, results[i].err = s.Add(p) })
	}
	wg.Wait()
	seen := map[score.PartID]bool{}
	for _, r := range results {
		require.NoError(t, r.err)
		assert.False(t, seen[r.id])
		seen[r.id] = true
		assert.Len(t, r.s.MustParts(), 1)
		assert.True(t, r.s.MustPart(r.id).Equal(p))
	}
	assert.Empty(t, s.MustParts())
	_, err := results[0].s.Part(results[1].id)
	require.ErrorIs(t, err, score.ErrPart)
}

func TestConcurrentReadsAndMeterProposals(t *testing.T) {
	t.Parallel()
	p := emptyPart(4, 4, 2)
	p = write(p, 1, tone(p, "C4", 1, 4, event.Marks{}))
	s, ids := assemble(clock60(), p, p)
	type result struct {
		slice    []score.PartSlice
		sounds   []score.Sound
		part     *track.Track
		proposal score.MeterProposal
		errs     [4]error
	}
	results := make([]result, 8)
	var wg sync.WaitGroup
	for i := range results {
		wg.Go(func() {
			results[i].slice, results[i].errs[0] = s.At(pos(0, 1))
			results[i].sounds, results[i].errs[1] = s.Sounds()
			results[i].part, results[i].errs[2] = s.Part(ids[0])
			results[i].proposal, results[i].errs[3] = s.ChangeMeter(1, 1, meter.MustNew(3, 4))
		})
	}
	wg.Wait()
	for _, r := range results {
		for _, err := range r.errs {
			require.NoError(t, err)
		}
		assert.Len(t, r.slice, 2)
		assert.Len(t, r.sounds, 2)
		assert.True(t, r.part.Equal(p))
		assertPos(t, pos(3, 4), r.proposal.Score.MustPart(ids[0]).MustGrid().MustBarStart(2))
	}
	assertPos(t, pos(1, 1), s.MustPart(ids[0]).MustGrid().MustBarStart(2))
}
