package duration_test

import (
	"runtime"
	"testing"
	"time"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
)

// readingsAllocLimit bounds the bytes a single Readings call may allocate in
// TestReadingsAllocationBound, unless a case sets a tighter bound.
//
// The bound separates a search whose every step is bounded, which stays
// several times below it on every case, from the defects it guards against.
// A step whose cost grows with the nesting depth allocated about 1.5 GB on
// one quarter under 1:1 at depth 1000. A missing step bound allocated
// gigabytes on the dense 1/64 input and on the quarters-and-a-fifth input.
// 256 MB is a sixth of the smallest of these; 512 MB would leave too little
// room for the tighter case below.
const readingsAllocLimit = 256 << 20

// readingsTruncateAllocLimit bounds the call that truncates many readings to
// MaxReadings. Collecting and sorting every reading before truncating
// allocated about 66 MB plus 3.3 MB per quarter on 66·1/12 followed by
// quarters (28 657 readings): about 190 MB with 40 quarters. The race
// detector multiplies the resident memory of that defect about fourfold
// (272 MB allocated with 60 quarters reached 1 GB resident), so a larger
// input is not safe to run against it. 96 MB is half the defective
// allocation; a search that stops after the first readings stays several
// times below it.
const readingsTruncateAllocLimit = 96 << 20

// readingsAllocTimeLimit bounds the time of each call of the test.
const readingsAllocTimeLimit = 10 * time.Second

// The work and the memory of a Readings call are bounded independently of
// the input and of the number of readings. The test measures the bytes
// allocated by each call with runtime.MemStats.TotalAlloc, so it is not
// parallel: paused parallel tests do not run while it runs, and its subtests
// run one at a time. A call that exceeds its time limit fails the test at
// once and the measurement of later subtests is then not meaningful.
//
//nolint:paralleltest // TotalAlloc is process-wide; parallel tests would add their allocations.
func TestReadingsAllocationBound(t *testing.T) {
	sextuplets := []tuplet.Tuplet{tuplet.Triplet(), tuplet.Sextuplet()}

	tests := []struct {
		name        string
		ds          []duration.Duration
		opts        duration.ReadingsOptions
		hasReadings bool
		limit       uint64 // 0 means readingsAllocLimit
	}{
		{
			name: "dense 1/64 free nested",
			ds:   repeatDur(1000, 1, 64),
			opts: duration.ReadingsOptions{
				Allowed: fiveRatios(), MaxDepth: 1, MaxReadings: 4, FreeContents: true,
			},
			hasReadings: true,
		},
		{
			name: "quarters and a fifth at depth 7",
			ds:   []duration.Duration{dur(1, 4), dur(1, 4), dur(1, 4), dur(1, 4), dur(1, 5)},
			opts: duration.ReadingsOptions{Allowed: fiveRatios(), MaxDepth: 7},
		},
		{
			name: "quarters and a fifth at depth 20",
			ds:   []duration.Duration{dur(1, 4), dur(1, 4), dur(1, 4), dur(1, 4), dur(1, 5)},
			opts: duration.ReadingsOptions{Allowed: fiveRatios(), MaxDepth: 20},
		},
		{
			// 28 657 readings; the first four must not need all of them.
			name:        "many readings truncated to four",
			ds:          append(repeatDur(66, 1, 12), repeatDur(40, 1, 4)...),
			opts:        duration.ReadingsOptions{Allowed: sextuplets, MaxReadings: 4},
			hasReadings: true,
			limit:       readingsTruncateAllocLimit,
		},
		{
			name:        "one quarter under 1:1 at depth 1000",
			ds:          []duration.Duration{dur(1, 4)},
			opts:        duration.ReadingsOptions{Allowed: []tuplet.Tuplet{tp(1, 1)}, MaxDepth: 1000},
			hasReadings: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runtime.GC()

			var before, after runtime.MemStats

			runtime.ReadMemStats(&before)

			r := within(t, readingsAllocTimeLimit, func() result[[]duration.Reading] {
				readings, err := duration.Readings(tt.ds, tt.opts)

				return result[[]duration.Reading]{readings, err}
			})

			runtime.ReadMemStats(&after)

			limit := tt.limit
			if limit == 0 {
				limit = readingsAllocLimit
			}

			allocated := after.TotalAlloc - before.TotalAlloc
			assert.LessOrEqual(t, allocated, limit,
				"the call allocated %d MB, the bound is %d MB", allocated>>20, limit>>20)

			if tt.hasReadings {
				assertBoundedReadings(t, tt.ds, tt.opts, r)

				return
			}

			if r.err != nil {
				assertErr(t, r.err, duration.ErrSearchLimit)

				return
			}

			for _, reading := range r.v {
				assertExactReading(t, tt.ds, reading)
			}
		})
	}
}
