package frequency_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/go-muse/musex/frequency"

	"github.com/stretchr/testify/require"
)

func TestCriterion_ConcurrentUse(t *testing.T) {
	const workers = 24
	const iterations = 25
	for _, criterion := range definedCriteria() {
		t.Run(criterion.name, func(t *testing.T) {
			var wait sync.WaitGroup
			errors := make(chan error, workers)
			for worker := range workers {
				wait.Go(func() {
					for range iterations {
						text := criterion.c.String()
						near, err := criterion.c.Near(440, 440)
						if err != nil || !near {
							errors <- fmt.Errorf("worker %d: Near = %t, %v", worker, near, err)
							return
						}
						comparison, err := criterion.c.Compare(440, 440)
						if err != nil || comparison != 0 {
							errors <- fmt.Errorf("worker %d: Compare = %d, %v", worker, comparison, err)
							return
						}
						if !criterion.c.Equal(criterion.c) {
							errors <- fmt.Errorf("worker %d: Equal is false", worker)
							return
						}
						parsed, err := frequency.Parse(text)
						if err != nil || !criterion.c.Equal(parsed) {
							errors <- fmt.Errorf("worker %d: Parse round trip failed: %v", worker, err)
							return
						}
						marshaled, err := criterion.c.MarshalText()
						if err != nil || string(marshaled) != text {
							errors <- fmt.Errorf("worker %d: MarshalText = %q, %v", worker, marshaled, err)
							return
						}
						appended, err := criterion.c.AppendText(make([]byte, 0, len(text)))
						if err != nil || string(appended) != text {
							errors <- fmt.Errorf("worker %d: AppendText = %q, %v", worker, appended, err)
							return
						}
						var unmarshaled frequency.Criterion
						if err := unmarshaled.UnmarshalText([]byte(text)); err != nil || !criterion.c.Equal(unmarshaled) {
							errors <- fmt.Errorf("worker %d: UnmarshalText failed: %v", worker, err)
							return
						}
						for _, component := range criterion.c.Components() {
							if !component.Equal(component) {
								errors <- fmt.Errorf("worker %d: component is not reflexively equal", worker)
								return
							}
						}
					}
				})
			}
			wait.Wait()
			close(errors)
			for err := range errors {
				require.NoError(t, err)
			}
		})
	}
}
