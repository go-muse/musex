package rounding_test

import (
	"fmt"
	"math/big"
	"sync"
	"testing"

	"github.com/go-muse/musex/common/rounding"

	"github.com/stretchr/testify/require"
)

func TestConcurrentCallsOnSharedInputs(t *testing.T) {
	t.Parallel()
	const workers = 8
	const iterations = 50
	n, d := big.NewInt(7), big.NewInt(2)
	x, step := big.NewRat(7, 3), big.NewRat(1, 4)
	zeroInt, zeroRat := new(big.Int), new(big.Rat)
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			for range iterations {
				q, r, err := rounding.HalfEven.QuoRemBig(n, d)
				if err != nil {
					errs <- err
					return
				}
				if q.Int64() != 4 || r.Int64() != -1 {
					errs <- fmt.Errorf("QuoRemBig = %v, %v", q, r)
					return
				}
				integer, err := rounding.HalfUp.RatInt(x)
				if err != nil {
					errs <- err
					return
				}
				if integer.Int64() != 3 {
					errs <- fmt.Errorf("RatInt = %v", integer)
					return
				}
				k, v, err := rounding.HalfEven.RatMultiple(x, step)
				if err != nil {
					errs <- err
					return
				}
				if k.Int64() != 9 || v.Cmp(big.NewRat(9, 4)) != 0 {
					errs <- fmt.Errorf("RatMultiple = %v, %v", k, v)
					return
				}
				decimal, err := rounding.HalfAwayFromZero.RatDecimal(x, 4)
				if err != nil {
					errs <- err
					return
				}
				if decimal != "0.6667" {
					errs <- fmt.Errorf("RatDecimal = %q", decimal)
					return
				}
				zeroQ, _, err := rounding.Down.QuoRemBig(zeroInt, d)
				if err != nil {
					errs <- err
					return
				}
				if zeroQ.Sign() != 0 {
					errs <- fmt.Errorf("zero QuoRemBig = %v", zeroQ)
					return
				}
				zeroInteger, err := rounding.Up.RatInt(zeroRat)
				if err != nil {
					errs <- err
					return
				}
				if zeroInteger.Sign() != 0 {
					errs <- fmt.Errorf("zero RatInt = %v", zeroInteger)
					return
				}
				zeroK, zeroV, err := rounding.HalfEven.RatMultiple(zeroRat, step)
				if err != nil {
					errs <- err
					return
				}
				if zeroK.Sign() != 0 || zeroV.Sign() != 0 {
					errs <- fmt.Errorf("zero RatMultiple = %v, %v", zeroK, zeroV)
					return
				}
				if _, err := rounding.HalfEven.FloatMultiple(0.15, 0.1); err != nil {
					errs <- err
					return
				}
				if _, err := rounding.HalfEven.ChooseIndex(3, 0); err != nil {
					errs <- err
					return
				}
			}
		})
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}
}
