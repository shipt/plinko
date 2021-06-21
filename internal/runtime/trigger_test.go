package runtime

import (
	"context"
	"errors"
	"testing"

	"github.com/shipt/plinko"
	"github.com/shipt/plinko/plinkoerror"
	"github.com/stretchr/testify/assert"
)

const Created plinko.State = "Created"
const Opened plinko.State = "Opened"
const Claimed plinko.State = "Claimed"
const ArriveAtStore plinko.State = "ArrivedAtStore"
const MarkedAsPickedUp plinko.State = "MarkedAsPickedup"
const Delivered plinko.State = "Delivered"
const Canceled plinko.State = "Canceled"
const Returned plinko.State = "Returned"

const Submit plinko.Trigger = "Submit"
const Cancel plinko.Trigger = "Cancel"
const Open plinko.Trigger = "Open"
const Claim plinko.Trigger = "Claim"
const Deliver plinko.Trigger = "Deliver"
const Return plinko.Trigger = "Return"
const Reinstate plinko.Trigger = "Reinstate"

type testPayload struct {
	state     plinko.State
	condition bool
}

func (p *testPayload) GetState() plinko.State {
	return p.state
}

func createPlinkoDefinition() plinko.PlinkoDefinition {
	stateMap := make(map[plinko.State]*InternalStateDefinition)
	p := PlinkoDefinition{
		States: &stateMap,
	}

	p.Abs = AbstractSyntax{}

	return &p
}
func TestCanFireWithPermitIf(t *testing.T) {
	p := createPlinkoDefinition()

	p.Configure(Created).
		PermitIf(PermitIfPredicate, Open, Opened)

	p.Configure(Opened)

	co := p.Compile()

	psm := co.StateMachine

	payload := &testPayload{
		state:     Created,
		condition: true,
	}
	assert.Nil(t, psm.CanFire(context.TODO(), payload, Open))

	payload.condition = false
	assert.NotNil(t, psm.CanFire(context.TODO(), payload, Open))

}

func TestCanFireWithNilPermitIf(t *testing.T) {
	p := createPlinkoDefinition()

	p.Configure(Created).
		PermitIf(nil, Open, Opened)

	p.Configure(Opened)

	co := p.Compile()

	psm := co.StateMachine

	payload := &testPayload{
		state:     Created,
		condition: true,
	}
	assert.Nil(t, psm.CanFire(context.TODO(), payload, Open))

}

func TransitionFn(errorOnCall bool) func(c context.Context, p plinko.Payload, t plinko.TransitionInfo) (plinko.Payload, error) {
	return func(c context.Context, p plinko.Payload, t plinko.TransitionInfo) (plinko.Payload, error) {
		if errorOnCall {
			return p, errors.New("error")
		}

		v := p.(*testPayload)
		v.state = t.GetDestination()

		return v, nil
	}
}
func TestFireWithEntryErrorHandling(t *testing.T) {
	p := createPlinkoDefinition()

	p.Configure(Created).
		Permit(Open, Opened).
		Permit(Cancel, Canceled).
		OnExit(TransitionFn(false))

	p.Configure(Opened).
		OnEntry(TransitionFn(true))

	co := p.Compile()

	psm := co.StateMachine

	payload := &testPayload{
		state:     Created,
		condition: true,
	}

	pr, err := psm.Fire(context.TODO(), payload, Open)
	assert.NotNil(t, err)
	assert.NotNil(t, pr)
}

func TestFireWithReentrancy(t *testing.T) {
	p := createPlinkoDefinition()

	p.Configure(Created).
		PermitReentry(Open).
		PermitReentry(Cancel).
		OnExit(TransitionFn(false))

	co := p.Compile()

	psm := co.StateMachine

	payload := &testPayload{
		state:     Created,
		condition: true,
	}

	pr, err := psm.Fire(context.TODO(), payload, Open)
	assert.Nil(t, err)
	assert.NotNil(t, pr)

	assert.Equal(t, Created, pr.GetState())
}

func TestFireWithReentrancyConditional(t *testing.T) {
	p := createPlinkoDefinition()

	p.Configure(Created).
		PermitReentryIf(func(_ context.Context, p plinko.Payload, t plinko.TransitionInfo) error {

			return nil
		}, Open).
		PermitReentry(Cancel).
		OnExit(TransitionFn(false))

	co := p.Compile()

	psm := co.StateMachine

	payload := &testPayload{
		state:     Created,
		condition: true,
	}

	pr, err := psm.Fire(context.TODO(), payload, Open)
	assert.Nil(t, err)
	assert.NotNil(t, pr)

	assert.Equal(t, Created, pr.GetState())
}

func TestFireWithReentrancyConditionalFalse(t *testing.T) {
	p := createPlinkoDefinition()

	p.Configure(Created).
		PermitReentryIf(func(_ context.Context, p plinko.Payload, t plinko.TransitionInfo) error {

			return errors.New("can't do that")
		}, Open).
		PermitReentry(Cancel).
		OnExit(TransitionFn(false))

	co := p.Compile()

	psm := co.StateMachine

	payload := &testPayload{
		state:     Created,
		condition: true,
	}

	pr, err := psm.Fire(context.TODO(), payload, Open)
	assert.NotNil(t, err)
	assert.Equal(t, "Conditional Trigger 'Open' conditions not met for state: Created", err.Error())
	assert.NotNil(t, pr)

	assert.Equal(t, Created, pr.GetState())
}

func TestFireWithExitErrorHandling(t *testing.T) {
	p := createPlinkoDefinition()

	p.Configure(Created).
		Permit(Open, Opened).
		Permit(Cancel, Canceled).
		OnExit(TransitionFn(true))

	p.Configure(Opened).
		OnEntry(TransitionFn(false)).
		OnTriggerEntry(Open, nil).
		OnError(nil).
		OnTriggerExit(Cancel, nil)

	co := p.Compile()

	psm := co.StateMachine

	payload := &testPayload{
		state:     Created,
		condition: true,
	}

	pr, err := psm.Fire(context.TODO(), payload, Open)
	assert.NotNil(t, err)
	assert.NotNil(t, pr)
}

func TestCanFireWithInvalidStateAndTrigger(t *testing.T) {
	p := createPlinkoDefinition()

	p.Configure(Created).
		PermitIf(PermitIfPredicate, Open, Opened)

	p.Configure(Opened)

	co := p.Compile()

	psm := co.StateMachine

	payload := &testPayload{
		state:     "not-a-real-state",
		condition: true,
	}
	err := psm.CanFire(context.TODO(), payload, Open)
	assert.NotNil(t, err)

	var pse *plinkoerror.PlinkoStateError
	assert.True(t, errors.As(err, &pse))

	assert.Equal(t, plinko.State("not-a-real-state"), pse.State)

	payload.state = Created
	err = psm.CanFire(context.TODO(), payload, "Run")

	assert.NotNil(t, err)

	var pte *plinkoerror.PlinkoTriggerError
	assert.True(t, errors.As(err, &pte))

	assert.Equal(t, plinko.Trigger("Run"), pte.Trigger)

	retval, err := psm.Fire(context.TODO(), payload, Open)

	assert.Nil(t, err)
	assert.NotNil(t, retval)

}

func TestFireWithInvalidStateAndTrigger(t *testing.T) {
	p := createPlinkoDefinition()

	p.Configure(Created).
		PermitIf(PermitIfPredicate, Open, Opened)

	p.Configure(Opened)

	co := p.Compile()

	psm := co.StateMachine

	payload := &testPayload{
		state:     "not-a-real-state",
		condition: true,
	}
	p1, err := psm.Fire(context.TODO(), payload, Open)
	assert.NotNil(t, err)
	assert.NotNil(t, p1)

	var pse *plinkoerror.PlinkoStateError
	assert.True(t, errors.As(err, &pse))

	assert.Equal(t, plinko.State("not-a-real-state"), pse.State)

	payload.state = Created
	p1, err = psm.Fire(context.TODO(), payload, "Run")

	assert.NotNil(t, err)
	assert.NotNil(t, p1)

	var pte *plinkoerror.PlinkoTriggerError
	assert.True(t, errors.As(err, &pte))

	assert.Equal(t, plinko.Trigger("Run"), pte.Trigger)

	retval, err := psm.Fire(context.TODO(), payload, Open)

	assert.Nil(t, err)
	assert.NotNil(t, retval)

}
func PermitIfPredicate(_ context.Context, p plinko.Payload, t plinko.TransitionInfo) error {
	tp := p.(*testPayload)

	if tp.condition {
		return nil
	}

	return errors.New("permit failed")
}

func TestEnumerateTriggers(t *testing.T) {
	p := createPlinkoDefinition()

	p.Configure(Created).
		Permit(Open, Opened).
		Permit(Cancel, Canceled)

	p.Configure(Opened)
	p.Configure(Canceled)

	co := p.Compile()

	psm := co.StateMachine
	payload := &testPayload{state: Created}
	triggers, err := psm.EnumerateActiveTriggers(payload)

	assert.Nil(t, err)

	payload = &testPayload{state: Opened}
	triggers, err = psm.EnumerateActiveTriggers(payload)

	assert.Nil(t, err)
	assert.Equal(t, 0, len(triggers))

	// request a state that doesn't exist in the state machine definiton and get an error thrown
	payload = &testPayload{state: Claimed}
	triggers, err = psm.EnumerateActiveTriggers(payload)

	assert.NotNil(t, err)
}
