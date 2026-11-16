package degree_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/degree"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDegreeConcurrentPureOperationsAndReads(t *testing.T) {
	t.Parallel()
	n := degreeTestNote(t, spelling.B(), cents.Zero(), nil)
	d := degree.MustNewWithNote(4, n)
	const workers = 12
	start := make(chan struct{})
	errors := make(chan error, workers)
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for range 50 {
				if !d.IsValid() || d.Number() != 4 || !d.HasNote() || !n.Equal(d.Note()) {
					errors <- fmt.Errorf("concurrent read returned an unexpected degree")
					return
				}
				clone := degree.MustNewWithNote(4, n)
				if !d.Equal(clone) || d.Compare(clone) != 0 || !d.SameNumber(clone) || !d.SameNote(clone) {
					errors <- fmt.Errorf("concurrent comparison returned an unexpected result")
					return
				}
				decimal, err := d.Decimal()
				if err != nil || decimal != "4" {
					errors <- fmt.Errorf("concurrent Decimal = %q, %v", decimal, err)
					return
				}
				roman, err := d.Roman()
				if err != nil || roman != "IV" {
					errors <- fmt.Errorf("concurrent Roman = %q, %v", roman, err)
					return
				}
				parsed, err := degree.Parse("IV")
				if err != nil || !degree.MustNew(4).Equal(parsed) {
					errors <- fmt.Errorf("concurrent Parse returned an unexpected result: %v", err)
					return
				}
				next, wrapped, err := d.Next(7)
				if err != nil || next.Number() != 5 || wrapped {
					errors <- fmt.Errorf("concurrent Next returned an unexpected result: %v", err)
					return
				}
				previous, wrapped, err := d.Previous(7)
				if err != nil || previous.Number() != 3 || wrapped {
					errors <- fmt.Errorf("concurrent Previous returned an unexpected result: %v", err)
					return
				}
				shifted, cycles, err := d.Shift(7, 9)
				if err != nil || shifted.Number() != 6 || cycles != 1 || shifted.HasNote() {
					errors <- fmt.Errorf("concurrent Shift returned an unexpected result: %v", err)
					return
				}
				up, err := d.DistanceUp(degree.Fifth(), 7)
				if err != nil || up != 1 {
					errors <- fmt.Errorf("concurrent DistanceUp = %d, %v", up, err)
					return
				}
				down, err := d.DistanceDown(degree.Fifth(), 7)
				if err != nil || down != 6 {
					errors <- fmt.Errorf("concurrent DistanceDown = %d, %v", down, err)
					return
				}
				shortest, err := d.ShortestDistance(degree.Fifth(), 7)
				if err != nil || shortest != 1 {
					errors <- fmt.Errorf("concurrent ShortestDistance = %d, %v", shortest, err)
					return
				}
				normalized, cycles, err := d.Normalize(3)
				if err != nil || normalized.Number() != 1 || cycles != 1 {
					errors <- fmt.Errorf("concurrent Normalize returned an unexpected result: %v", err)
					return
				}
				denormalized, err := normalized.Denormalize(3, cycles)
				if err != nil || denormalized.Number() != 4 {
					errors <- fmt.Errorf("concurrent Denormalize returned an unexpected result: %v", err)
					return
				}
				inside, err := d.InCircle(7)
				if err != nil || !inside {
					errors <- fmt.Errorf("concurrent InCircle = %t, %v", inside, err)
					return
				}
			}
		}()
	}
	close(start)
	wg.Wait()
	close(errors)
	for err := range errors {
		require.NoError(t, err)
	}
}

func TestDegreeConcurrentMutationOfIndependentCopies(t *testing.T) {
	t.Parallel()
	n := degreeTestNote(t, spelling.B(), cents.Zero(), nil)
	base := degree.MustNewWithNote(4, n)
	const workers = 10
	copies := make([]degree.Degree, workers)
	for i := range copies {
		copies[i] = base
	}
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := range copies {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			copies[i].MustSetNumber(i + 1)
			copies[i].MustTransformNote(func(n note.Note) (note.Note, error) {
				if err := n.Raise(); err != nil {
					return note.Note{}, err
				}
				return n, nil
			})
		}()
	}
	close(start)
	wg.Wait()
	expected := degreeTestNote(t, spelling.B(), cents.Zero(), nil)
	expected.MustRaise()
	for i := range copies {
		assertDegreeState(t, copies[i], i+1, true, expected)
	}
	assertDegreeState(t, base, 4, true, n)
}

func TestDegreeAllSupportsConcurrentTraversals(t *testing.T) {
	t.Parallel()
	sequence := degree.MustAll(13)
	const workers = 8
	start := make(chan struct{})
	results := make(chan []int, workers)
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			numbers := make([]int, 0, 13)
			for d := range sequence {
				numbers = append(numbers, d.Number())
			}
			results <- numbers
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	for numbers := range results {
		require.Len(t, numbers, 13)
		for i, number := range numbers {
			assert.Equal(t, i+1, number)
		}
	}
}
