package database

import (
	"sync"

	"github.com/France-ioi/AlgoreaBackend/v2/golang"
)

// PropagationStep represents a step in the propagation process.
type PropagationStep string

const (
	// PropagationStepAccessMain is the main step of the access propagation.
	PropagationStepAccessMain PropagationStep = "access: main step"

	// PropagationStepResultsNamedLockAcquire is the step of acquiring the named lock for results propagation.
	PropagationStepResultsNamedLockAcquire PropagationStep = "results: acquire named lock"
	// PropagationStepResultsInsideNamedLockMoveFromResultsPropagateToResultsPropagateInternal is the step of moving
	// from results_propagate into results_propagate_internal inside the named lock.
	PropagationStepResultsInsideNamedLockMoveFromResultsPropagateToResultsPropagateInternal PropagationStep = "results: inside named lock: " +
		"move from results_propagate"
	// PropagationStepResultsInsideNamedLockInsertIntoResultsPropagateInternal is the step of inserting into
	// results_propagate_internal inside the named lock.
	PropagationStepResultsInsideNamedLockInsertIntoResultsPropagateInternal PropagationStep = "results: inside named lock: " +
		"insert into results_propagate_internal"
	// PropagationStepResultsInsideNamedLockMarkAndInsertResults is the step of marking and inserting results inside the named lock.
	PropagationStepResultsInsideNamedLockMarkAndInsertResults PropagationStep = "results: inside named lock: mark and insert results"
	// PropagationStepResultsInsideNamedLockMain is the main step of the results propagation inside the named lock.
	PropagationStepResultsInsideNamedLockMain PropagationStep = "results: inside named lock: main step"
	// PropagationStepResultsInsideNamedLockItemUnlocking is the step of unlocking the items inside the named lock.
	PropagationStepResultsInsideNamedLockItemUnlocking PropagationStep = "results: inside named lock: item unlocking"
	// PropagationStepResultsPropagationScheduling is the step of scheduling the propagation of permissions and results.
	PropagationStepResultsPropagationScheduling PropagationStep = "results: propagation scheduling"
)

// PropagationStepSetAccess returns a set of access (permissions) propagation steps.
func PropagationStepSetAccess() *golang.Set[PropagationStep] {
	return golang.NewSet(PropagationStepAccessMain).MarkImmutable()
}

// PropagationStepSetResults returns a set of results propagation steps.
func PropagationStepSetResults() *golang.Set[PropagationStep] {
	return golang.NewSet(
		PropagationStepResultsNamedLockAcquire,
		PropagationStepResultsInsideNamedLockMoveFromResultsPropagateToResultsPropagateInternal,
		PropagationStepResultsInsideNamedLockInsertIntoResultsPropagateInternal,
		PropagationStepResultsInsideNamedLockMarkAndInsertResults,
		PropagationStepResultsInsideNamedLockMain,
		PropagationStepResultsInsideNamedLockItemUnlocking,
		PropagationStepResultsPropagationScheduling,
	).MarkImmutable()
}

// BeforePropagationStepHookFunc is a type of a function that is called before each propagation step.
type BeforePropagationStepHookFunc func(step PropagationStep)

//nolint:gochecknoglobals // this hook is for testing purposes only
var (
	// beforePropagationStepHook is a hook that is called before each propagation step.
	beforePropagationStepHook BeforePropagationStepHookFunc = func(_ PropagationStep) {}
	// beforePropagationStepMutex protects beforePropagationStepHook.
	beforePropagationStepMutex sync.RWMutex
)

// SetBeforePropagationStepHook sets a hook that is called before each propagation step.
func SetBeforePropagationStepHook(newHook BeforePropagationStepHookFunc) {
	beforePropagationStepMutex.Lock()
	defer beforePropagationStepMutex.Unlock()
	beforePropagationStepHook = newHook
}

// GetBeforePropagationStepHook returns a hook that is called before each propagation step.
func GetBeforePropagationStepHook() BeforePropagationStepHookFunc {
	beforePropagationStepMutex.RLock()
	defer beforePropagationStepMutex.RUnlock()
	return beforePropagationStepHook
}

// CallBeforePropagationStepHook calls the hook that is called before each propagation step.
func CallBeforePropagationStepHook(step PropagationStep) {
	GetBeforePropagationStepHook()(step)
}
