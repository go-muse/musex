package modal

import (
	"fmt"
	"iter"

	"github.com/go-muse/musex/mode"
)

// RelationTo returns the relation of the source degree from to the target
// degree to: the element of RelationsFrom(m, from) whose To is to. When
// to >= from the target is in the same cycle (Cycle 0); when to < from it is
// in the next cycle (Cycle 1), and the interval is the full upward interval
// m.BetweenDegrees(from, to+Len()). RelationTo(m, i, i) is the perfect unison,
// Neutral, Cycle 0. Only the requested interval is checked: an unsupported
// quality of another relation of the row does not prevent success.
//
// Errors, in this order: ErrUnset together with mode.ErrUnset for an unset
// mode; ErrInvalidDegree for from, then for to, outside 1..Len();
// ErrUnsupportedStructure for an open mode and to < from, since an open mode
// has no next cycle; an error matching mode.ErrOutOfRange and
// interval.ErrOutOfRange when the interval is not representable (other causes
// in the chain are not promised); ErrUnsupportedQuality for a neutral or other
// fractional quality. On error the result is the zero Relation. Time and
// memory are O(1).
func RelationTo(m mode.Mode, from, to int) (Relation, error) {
	if err := checkDegrees(m, from, to); err != nil {
		return Relation{}, err
	}
	if m.IsOpen() && to < from {
		return Relation{}, fmt.Errorf("relation %d -> %d: an open mode has no next cycle: %w", from, to, ErrUnsupportedStructure)
	}
	r, err := relationInterval(m, from, to)
	if err != nil {
		return Relation{}, err
	}
	if err := classifyRelation(&r); err != nil {
		return Relation{}, err
	}
	return r, nil
}

// RelationsFrom returns the row of relations of the source degree from, upward
// in order. For a closed mode with any cycle it has Len() relations: the
// targets from, from+1, ..., Len() with Cycle 0, then 1, ..., from-1 with
// Cycle 1. For an open mode it has Len()-from+1 relations, the targets from,
// ..., Len(), all with Cycle 0. The first relation is the perfect unison to
// from itself, Neutral; the interval to from in the next cycle is not
// included. Each Interval is the full specific interval upward from from,
// never reduced to an octave, and each Category classifies its quality as
// Relative does. Every element has From == from. The row from degree 1 matches
// RelativePositions by number, interval, and category.
//
// Errors, in this order: ErrUnset together with mode.ErrUnset; ErrInvalidDegree
// for from outside 1..Len(); the first interval of the row, in row order, that
// is not representable, matching mode.ErrOutOfRange and interval.ErrOutOfRange;
// then the first unsupported quality in row order, ErrUnsupportedQuality.
// Representability of the whole row is checked before any quality. On error
// the result is nil, without a partial row. Time and memory are O(Len()); the
// slice belongs to the caller.
func RelationsFrom(m mode.Mode, from int) ([]Relation, error) {
	if err := checkDegrees(m, from); err != nil {
		return nil, err
	}
	return relationRow(m, from)
}

// RelationTable returns the rows of relations from every degree, in the order
// of degree numbers 1..Len(). Row i equals RelationsFrom(m, i): a closed mode
// has Len() rows of Len() relations, and the row of degree i of an open mode
// has Len()-i+1 relations.
//
// The whole table is validated before the iterator is returned: ErrUnset
// together with mode.ErrUnset, then the first unrepresentable interval by
// source degree and row order (matching mode.ErrOutOfRange and
// interval.ErrOutOfRange), then the first unsupported quality in the same
// order (ErrUnsupportedQuality). On error the iterator is nil; there is no
// partial table. Validation costs O(Len()²) time.
//
// The iterator captures the structure of m when RelationTable is called;
// later changes of the caller's mode variable do not affect it. Rows are
// computed on demand: the walk does not store the whole table, and each row
// is a new caller-owned slice computed in O(Len()) time and memory. The
// iterator can be walked again with the same rows, stops at once when yield
// returns false, and can be walked from several goroutines at once. A walk
// neither returns errors nor panics on its own; a panic in yield is not
// recovered. Use slices.Collect for the complete table of a small mode and
// RelationsFrom for one row without validating the whole table.
func RelationTable(m mode.Mode) (iter.Seq[[]Relation], error) {
	if err := checkDegrees(m); err != nil {
		return nil, err
	}
	// Intervals checks representability of the whole table, in the order of
	// source degree and row, before returning its iterator. Qualities are
	// checked only after that, in the same order.
	intervals, err := m.Intervals()
	if err != nil {
		return nil, fmt.Errorf("relation table: %w", err)
	}
	for at := range intervals {
		if _, err := classify(at.Interval.Quality()); err != nil {
			return nil, fmt.Errorf("relation table: degree %d, distance %d (%s): %w", at.Degree.Number(), at.Distance, at.Interval, err)
		}
	}
	return func(yield func([]Relation) bool) {
		for from := 1; from <= m.Len(); from++ {
			row, err := relationRow(m, from)
			if err != nil {
				// The same immutable mode was validated in full before the
				// iterator was returned, so no row can fail here.
				panic(err)
			}
			if !yield(row) {
				return
			}
		}
	}, nil
}

// relationRow builds the row of a checked source degree. All intervals are
// built before any quality is checked.
func relationRow(m mode.Mode, from int) ([]Relation, error) {
	k := m.Len()
	count := k
	if m.IsOpen() {
		count = k - from + 1
	}
	row := make([]Relation, count)
	for n := range row {
		to := from + n
		if to > k {
			to -= k
		}
		r, err := relationInterval(m, from, to)
		if err != nil {
			return nil, err
		}
		row[n] = r
	}
	for n := range row {
		if err := classifyRelation(&row[n]); err != nil {
			return nil, err
		}
	}
	return row, nil
}

// relationInterval returns the relation of checked degrees without its
// category. A target below the source is taken in the next cycle through the
// virtual number to+Len(), which is at most twice mode.MaxDegrees.
func relationInterval(m mode.Mode, from, to int) (Relation, error) {
	r := Relation{From: from, To: to}
	target := to
	if to < from {
		r.Cycle = 1
		target = to + m.Len()
	}
	i, err := m.BetweenDegrees(from, target)
	if err != nil {
		return Relation{}, fmt.Errorf("relation %d -> %d: %w", from, to, err)
	}
	r.Interval = i
	return r, nil
}

func classifyRelation(r *Relation) error {
	c, err := classify(r.Interval.Quality())
	if err != nil {
		return fmt.Errorf("relation %d -> %d (%s): %w", r.From, r.To, r.Interval, err)
	}
	r.Category = c
	return nil
}
