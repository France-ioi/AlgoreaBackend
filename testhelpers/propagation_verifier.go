package testhelpers

import (
	"net/http/httptest"
	"testing"

	"github.com/France-ioi/AlgoreaBackend/v2/app"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/golang"
)

const (
	// PropagationStepSanityCheck is a fake propagation step which is called by PropagationVerifier
	// before triggering propagations in order to check that the hook is working without errors.
	PropagationStepSanityCheck database.PropagationStep = "sanity check"
	// PropagationStepAfterAllPropagations is a fake propagation step which is called by PropagationVerifier after all propagations.
	PropagationStepAfterAllPropagations database.PropagationStep = "after all propagations"

	maxAllowedNumberOfCallsForPropagationStep = 10 // maximum number of calls to each propagation step to detect infinite loops
)

type (
	// PropagationVerifierHookFunc is a function that is called by PropagationVerifier
	// before each propagation step and before and after all propagations.
	PropagationVerifierHookFunc func(step database.PropagationStep, allStepsCounter, currentStepCounter int,
		dataStore *database.DataStore, appServer *httptest.Server)

	// PropagationStepsSet is a set of propagation steps.
	PropagationStepsSet = golang.Set[database.PropagationStep]
)

// PropagationVerifier is a helper to test the propagation steps. It allows to run the code that triggers propagations
// and the code that should be executed before and after each propagation step.
// It also allows to check that the required propagation steps were called.
// Also, the propagation verifier detects infinite loops in the propagation steps by limiting the number of calls to each step.
type PropagationVerifier struct {
	fixture                  string
	hookFunc                 PropagationVerifierHookFunc
	requiredPropagationSteps *golang.Set[database.PropagationStep]
}

// NewPropagationVerifier creates a new PropagationVerifier with the required propagation steps.
func NewPropagationVerifier(requiredPropagationSteps *PropagationStepsSet) *PropagationVerifier {
	pv := &PropagationVerifier{
		requiredPropagationSteps: requiredPropagationSteps,
	}
	if pv.requiredPropagationSteps == nil {
		pv.requiredPropagationSteps = golang.NewSet[database.PropagationStep]()
	}
	return pv
}

// WithFixture sets the DB fixture to load before running the verification.
func (pv *PropagationVerifier) WithFixture(fixture string) *PropagationVerifier {
	pv.fixture = fixture
	return pv
}

// WithHook sets the hook function that will be called before each propagation step and before and after all propagations.
func (pv *PropagationVerifier) WithHook(hookFunc PropagationVerifierHookFunc) *PropagationVerifier {
	pv.hookFunc = hookFunc
	return pv
}

// Run runs the propagation verifier.
func (pv *PropagationVerifier) Run(
	t *testing.T,
	codeTriggeringPropagations func(dataStore *database.DataStore, appServer *httptest.Server),
) {
	t.Helper()

	db := SetupDBWithFixtureString(pv.fixture)

	dataStore := database.NewDataStore(db)
	err := dataStore.InTransaction(func(dataStore *database.DataStore) error {
		dataStore.SchedulePermissionsPropagation()
		dataStore.ScheduleResultsPropagation()
		return nil
	})
	_ = db.Close()
	if err != nil {
		t.Fatalf("initial propagation failed: %v", err)
	}

	// app server
	application, err := app.New()
	if err != nil {
		t.Fatalf("Unable to load a hooked app: %v", err)
	}
	defer func() { _ = application.Database.Close() }()
	appServer := httptest.NewServer(application.HTTPHandler)
	defer appServer.Close()

	db = application.Database
	dataStore = database.NewDataStore(db)

	// set up the hook
	oldBeforePropagationStepHook := database.GetBeforePropagationStepHook()
	defer func() { database.SetBeforePropagationStepHook(oldBeforePropagationStepHook) }()
	calledPropagationSteps := make(map[database.PropagationStep]int)
	stepCounter := 0
	database.SetBeforePropagationStepHook(func(step database.PropagationStep) {
		defer func() {
			if r := recover(); r != nil {
				if !t.Failed() {
					t.Errorf("beforePropagationStep(%q) panicked: %v", step, r)
				}
			}
		}()

		// Restore the original hook until the end of this function
		ourBeforePropagationStepHook := database.GetBeforePropagationStepHook()
		database.SetBeforePropagationStepHook(oldBeforePropagationStepHook)
		defer func() { database.SetBeforePropagationStepHook(ourBeforePropagationStepHook) }()

		stepCounter++

		if _, ok := calledPropagationSteps[step]; !ok {
			calledPropagationSteps[step] = 0
		}
		calledPropagationSteps[step]++

		if calledPropagationSteps[step] > maxAllowedNumberOfCallsForPropagationStep {
			_ = application.Database.Close() // stop all the app's API handlers
			t.Errorf("looks like an infinite loop in propagation step %q, called %d times", step, calledPropagationSteps[step])
			return
		}

		t.Logf("before propagation step %d: %q (%d)\n", stepCounter, step, calledPropagationSteps[step])

		if pv.hookFunc != nil {
			pv.hookFunc(step, stepCounter, calledPropagationSteps[step], dataStore, appServer)
		}
	})

	// Execute the code that should be executed before the propagation steps just to make sure it doesn't fail
	t.Log("sanity check before all propagations")
	database.CallBeforePropagationStepHook(PropagationStepSanityCheck)
	if t.Failed() {
		return
	}

	t.Log("triggering propagations")
	codeTriggeringPropagations(dataStore, appServer)

	t.Log("after all propagations")
	database.CallBeforePropagationStepHook(PropagationStepAfterAllPropagations)

	pv.assertAllRequiredStepsCalled(t, calledPropagationSteps)
}

func (pv *PropagationVerifier) assertAllRequiredStepsCalled(t *testing.T, calledPropagationSteps map[database.PropagationStep]int) {
	t.Helper()

	for _, requiredPropagationStep := range pv.requiredPropagationSteps.Values() {
		if _, ok := calledPropagationSteps[requiredPropagationStep]; !ok {
			t.Errorf("Required propagation step %q was not called", requiredPropagationStep)
		}
	}
}
