package database

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/golang"
)

func TestPropagationStepSets(t *testing.T) {
	for _, test := range []struct {
		name            string
		set             *golang.Set[PropagationStep]
		expectedContent []PropagationStep
	}{
		{
			name:            "PropagationStepSetAccess",
			set:             PropagationStepSetAccess(),
			expectedContent: []PropagationStep{PropagationStepAccessMain},
		},
		{
			name: "PropagationStepSetResults",
			set:  PropagationStepSetResults(),
			expectedContent: []PropagationStep{
				PropagationStepResultsNamedLockAcquire,
				PropagationStepResultsInsideNamedLockInsertIntoResultsPropagate,
				PropagationStepResultsInsideNamedLockMarkAndInsertResults,
				PropagationStepResultsInsideNamedLockMain,
				PropagationStepResultsInsideNamedLockItemUnlocking,
				PropagationStepResultsPropagationScheduling,
			},
		},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.set.Size(), len(test.expectedContent))
			for _, step := range test.expectedContent {
				assert.True(t, test.set.Contains(step), "step %q not found in the set", step)
			}
			assert.True(t, test.set.IsImmutable())
		})
	}
}

func TestBeforePropagationStepHook(t *testing.T) {
	var steps []PropagationStep
	oldHook := GetBeforePropagationStepHook()
	defer SetBeforePropagationStepHook(oldHook)
	SetBeforePropagationStepHook(func(step PropagationStep) {
		steps = append(steps, step)
	})
	CallBeforePropagationStepHook(PropagationStepAccessMain)
	assert.Equal(t, []PropagationStep{PropagationStepAccessMain}, steps)
	SetBeforePropagationStepHook(func(step PropagationStep) {
		steps = append(steps, step, step)
	})
	CallBeforePropagationStepHook(PropagationStepResultsNamedLockAcquire)
	assert.Equal(t,
		[]PropagationStep{
			PropagationStepAccessMain,
			PropagationStepResultsNamedLockAcquire,
			PropagationStepResultsNamedLockAcquire,
		},
		steps)
}
