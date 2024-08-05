package database

import "github.com/France-ioi/AlgoreaBackend/v2/golang"

// PropagationStep represents a step in the propagation process.
type PropagationStep string

const (
	// PropagationStepGroupAncestorsInit is the first step (initialization) of the group ancestors propagation.
	PropagationStepGroupAncestorsInit PropagationStep = "group ancestors: init"
	// PropagationStepGroupAncestorsMain is the main step of the group ancestors propagation.
	PropagationStepGroupAncestorsMain PropagationStep = "group ancestors: main step"

	// PropagationStepItemAncestorsInit is the first step (initialization) of the item ancestors propagation.
	PropagationStepItemAncestorsInit PropagationStep = "item ancestors: init"
	// PropagationStepItemAncestorsMain is the main step of the item ancestors propagation.
	PropagationStepItemAncestorsMain PropagationStep = "item ancestors: main step"

	// PropagationStepAccessMain is the main step of the access propagation.
	PropagationStepAccessMain PropagationStep = "access: main step"

	// PropagationStepResultsNamedLockAcquire is the step of acquiring the named lock for results propagation.
	PropagationStepResultsNamedLockAcquire PropagationStep = "results: acquire named lock"
	// PropagationStepResultsInsideNamedLockInsertIntoResultsPropagate is the step of inserting into results_propagate inside the named lock.
	PropagationStepResultsInsideNamedLockInsertIntoResultsPropagate PropagationStep = "results: inside named lock: " +
		"insert into results_propagate"
	// PropagationStepResultsInsideNamedLockMarkAndInsertResults is the step of marking and inserting results inside the named lock.
	PropagationStepResultsInsideNamedLockMarkAndInsertResults PropagationStep = "results: inside named lock: mark and insert results"
	// PropagationStepResultsInsideNamedLockMain is the main step of the results propagation inside the named lock.
	PropagationStepResultsInsideNamedLockMain PropagationStep = "results: inside named lock: main step"
	// PropagationStepResultsInsideNamedLockItemUnlocking is the step of unlocking the items inside the named lock.
	PropagationStepResultsInsideNamedLockItemUnlocking PropagationStep = "results: inside named lock: item unlocking"
	// PropagationStepResultsPropagationScheduling is the step of scheduling the propagation of permissions and results.
	PropagationStepResultsPropagationScheduling PropagationStep = "results: propagation scheduling"
)

// PropagationStepSetGroupAncestors returns a set of group ancestors propagation steps.
func PropagationStepSetGroupAncestors() *golang.Set[PropagationStep] {
	return golang.NewSet(PropagationStepGroupAncestorsInit, PropagationStepGroupAncestorsMain).MarkImmutable()
}

// PropagationStepSetItemAncestors returns a set of item ancestors propagation steps.
func PropagationStepSetItemAncestors() *golang.Set[PropagationStep] {
	return golang.NewSet(PropagationStepItemAncestorsInit, PropagationStepItemAncestorsMain).MarkImmutable()
}

// PropagationStepSetAccess returns a set of access (permissions) propagation steps.
func PropagationStepSetAccess() *golang.Set[PropagationStep] {
	return golang.NewSet(PropagationStepAccessMain).MarkImmutable()
}

// PropagationStepSetResults returns a set of results propagation steps.
func PropagationStepSetResults() *golang.Set[PropagationStep] {
	return golang.NewSet(
		PropagationStepResultsNamedLockAcquire,
		PropagationStepResultsInsideNamedLockInsertIntoResultsPropagate,
		PropagationStepResultsInsideNamedLockMarkAndInsertResults,
		PropagationStepResultsInsideNamedLockMain,
		PropagationStepResultsInsideNamedLockItemUnlocking,
		PropagationStepResultsPropagationScheduling,
	).MarkImmutable()
}

// BeforePropagationStep is a hook that is called before each propagation step.
var BeforePropagationStep = func(step PropagationStep) {}
