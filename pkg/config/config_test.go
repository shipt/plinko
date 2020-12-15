package config

import (
	"errors"
	"fmt"
	"testing"

	"github.com/shipt/plinko"
	"github.com/shipt/plinko/internal/runtime"
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
const NewOrder plinko.State = "NewOrder"

const Submit plinko.Trigger = "Submit"
const Cancel plinko.Trigger = "Cancel"
const Open plinko.Trigger = "Open"
const Claim plinko.Trigger = "Claim"
const Deliver plinko.Trigger = "Deliver"
const Return plinko.Trigger = "Return"
const Reinstate plinko.Trigger = "Reinstate"

func entryFunctionForTest(pp plinko.Payload, transitionInfo plinko.TransitionInfo) (plinko.Payload, error) {
	return nil, fmt.Errorf("misc entry error")
}

func exitFunctionForTest(pp plinko.Payload, transitionInfo plinko.TransitionInfo) (plinko.Payload, error) {
	return nil, fmt.Errorf("misc exit error")
}

func TestEntryAndExitFunctions(t *testing.T) {
	p := CreatePlinkoDefinition()
	ps := p.Configure("NewOrder")

	stateDef := ps.(runtime.InternalStateDefinition)
	assert.Nil(t, stateDef.Callbacks.OnExitFn)
	assert.Nil(t, stateDef.Callbacks.OnEntryFn)

	ps = ps.OnEntry(entryFunctionForTest)

	ps = ps.OnExit(exitFunctionForTest)

	stateDef = ps.(runtime.InternalStateDefinition)
	assert.NotNil(t, stateDef.Callbacks.OnExitFn)
	assert.NotNil(t, stateDef.Callbacks.OnEntryFn)

}

func TestPlinkoAsInterface(t *testing.T) {
	p := CreatePlinkoDefinition()

	p.Configure("NewOrder").
		Permit("Submit", "PublishedOrder").
		Permit("Review", "ReviewedOrder")
}

func TestUndefinedStateCompile(t *testing.T) {
	p := CreatePlinkoDefinition()

	p.Configure(NewOrder).
		Permit("Submit", "PublishedOrder")

	compilerOutput := p.Compile()
	assert.Equal(t, 1, len(compilerOutput.Messages))
	assert.Equal(t, plinko.CompileError, compilerOutput.Messages[0].CompileMessage)
	assert.Equal(t, "State 'PublishedOrder' undefined: Trigger 'Submit' declares a transition to this undefined state.", compilerOutput.Messages[0].Message)
}

func TestTriggerlessStateCompile(t *testing.T) {
	p := CreatePlinkoDefinition()

	p.Configure(NewOrder).
		Permit("Submit", "PublishedOrder")
	p.Configure("PublishedOrder")

	compilerOutput := p.Compile()
	assert.Equal(t, 1, len(compilerOutput.Messages))
	assert.Equal(t, plinko.CompileWarning, compilerOutput.Messages[0].CompileMessage)
	assert.Equal(t, "State 'PublishedOrder' is a state without any triggers (deadend state).", compilerOutput.Messages[0].Message)
}

func TestUmlDiagramming(t *testing.T) {
	p := CreatePlinkoDefinition()

	p.Configure(NewOrder).
		Permit("Submit", "PublishedOrder").
		Permit("Review", "UnderReview")

	p.Configure("PublishedOrder")

	p.Configure("UnderReview").
		Permit("CompleteReview", "PublishedOrder").
		Permit("RejectOrder", "RejectedOrder")

	p.Configure("RejectedOrder")

	uml, err := p.RenderUml()

	fmt.Println(uml)

	assert.Nil(t, err)
	assert.Equal(t, "@startuml\n[*] -> NewOrder \nNewOrder", string(uml)[0:35])
	assert.Equal(t, "\n@enduml", string(uml)[len(uml)-8:])
}

func TestPlinkoDefinition(t *testing.T) {
	stateMap := make(map[plinko.State]*runtime.InternalStateDefinition)
	plinko := runtime.PlinkoDefinition{
		States: &stateMap,
	}

	assert.NotPanics(t, func() {
		plinko.Configure("NewOrder").
			//			OnEntry()
			//			OnExit()
			Permit("Submit", "PublishedOrder").
			Permit("Review", "ReviewOrder")

		plinko.Configure("PublishedOrder")
		plinko.Configure("ReviewOrder")
	})

	assert.Panics(t, func() {
		plinko.Configure("NewOrder").
			Permit("Submit", "PublishedOrder").
			Permit("Review", "ReviewOrder")

		plinko.Configure("PublishedOrder")
		plinko.Configure("ReviewOrder")
		plinko.Configure("NewOrder")
	})
}

type testPayload struct {
	state     plinko.State
	condition bool
}

func (p testPayload) GetState() plinko.State {
	return p.state
}

func TestOnEntryTriggerOperation(t *testing.T) {

	p := CreatePlinkoDefinition()
	counter := 0

	p.Configure(NewOrder).
		Permit("Submit", "PublishedOrder").
		Permit("Review", "UnderReview")

	p.Configure("PublishedOrder").
		OnTriggerEntry("Resupply", func(pp plinko.Payload, transitionInfo plinko.TransitionInfo) (plinko.Payload, error) {
			counter++
			return pp, nil
		}).
		Permit("Resupply", "PublishedOrder").
		Permit("Resubmit", "PublishedOrder")

	compilerOutput := p.Compile()
	psm := compilerOutput.StateMachine

	payload := testPayload{state: "PublishedOrder"}

	psm.Fire(payload, "Resupply")
	assert.Equal(t, 1, counter)

	psm.Fire(payload, "Resubmit")
	assert.Equal(t, 1, counter)

	psm.Fire(payload, "Resupply")
	assert.Equal(t, 2, counter)

}

func OnNewOrderEntry(pp plinko.Payload, transitionInfo plinko.TransitionInfo) (plinko.Payload, error) {
	fmt.Printf("onentry: %+v", transitionInfo)
	return pp, nil
}
func TestStateMachine(t *testing.T) {
	p := CreatePlinkoDefinition()

	p.Configure(NewOrder).
		OnEntry(OnNewOrderEntry).
		Permit("Submit", "PublishedOrder").
		Permit("Review", "UnderReview")

	p.Configure("PublishedOrder").
		OnEntry(OnNewOrderEntry).
		Permit("Submit", NewOrder)

	p.Configure("UnderReview").
		Permit("CompleteReview", "PublishedOrder").
		Permit("RejectOrder", "RejectedOrder")

	p.Configure("RejectedOrder")

	visitCount := 0
	p.SideEffect(func(sa plinko.StateAction, payload plinko.Payload, ti plinko.TransitionInfo) {
		visitCount += 1
	})

	compilerOutput := p.Compile()
	psm := compilerOutput.StateMachine

	payload := testPayload{state: NewOrder}

	psm.Fire(payload, "Submit")

	assert.Equal(t, 3, visitCount)

}

func TestCanFire(t *testing.T) {
	p := CreatePlinkoDefinition()

	p.Configure(Created).
		Permit(Open, Opened)

	p.Configure(Opened)

	co := p.Compile()

	psm := co.StateMachine

	payload := testPayload{state: Created}

	assert.True(t, psm.CanFire(payload, Open))
	assert.False(t, psm.CanFire(payload, Deliver))
}

func PermitIfPredicate(p plinko.Payload, t plinko.TransitionInfo) bool {
	tp := p.(testPayload)

	return tp.condition
}

func TestCanFireWithPermitIf(t *testing.T) {
	p := CreatePlinkoDefinition()

	p.Configure(Created).
		PermitIf(PermitIfPredicate, Open, Opened)

	p.Configure(Opened)

	co := p.Compile()

	psm := co.StateMachine

	payload := testPayload{
		state:     Created,
		condition: true,
	}
	assert.True(t, psm.CanFire(payload, Open))

	payload.condition = false
	assert.False(t, psm.CanFire(payload, Open))

}

func TestDiagramming(t *testing.T) {
	p := CreatePlinkoDefinition()

	p.Configure(Created).
		OnEntry(OnNewOrderEntry).
		Permit(Open, Opened).
		Permit(Cancel, Canceled)

	p.Configure(Opened).
		Permit("AddItemToOrder", Opened).
		Permit(Claim, Claimed).
		Permit(Cancel, Canceled)

	p.Configure(Claimed).
		Permit("AddItemToOrder", Claimed).
		Permit(Submit, ArriveAtStore).
		Permit(Cancel, Canceled)

	p.Configure(ArriveAtStore).
		Permit(Submit, MarkedAsPickedUp).
		Permit(Cancel, Canceled)

	p.Configure(MarkedAsPickedUp).
		Permit(Deliver, Delivered).
		Permit(Cancel, Canceled)

	p.Configure(Delivered).
		Permit(Return, Returned)

	p.Configure(Canceled).
		Permit(Reinstate, Created)

	p.Configure(Returned)

	co := p.Compile()
	fmt.Printf("%+v\n", co.Messages)
	uml, err := p.RenderUml()

	fmt.Println(err)

	fmt.Println(uml)

}

func findTrigger(triggers []plinko.Trigger, trigger plinko.Trigger) bool {
	for _, v := range triggers {
		if v == trigger {
			return true
		}
	}

	return false
}

func TestEnumerateTriggers(t *testing.T) {
	p := CreatePlinkoDefinition()

	p.Configure(Created).
		Permit(Open, Opened).
		Permit(Cancel, Canceled)

	p.Configure(Opened)
	p.Configure(Canceled)

	co := p.Compile()

	psm := co.StateMachine
	payload := testPayload{state: Created}
	triggers, err := psm.EnumerateActiveTriggers(payload)

	assert.Nil(t, err)
	assert.True(t, findTrigger(triggers, Open))
	assert.True(t, findTrigger(triggers, Cancel))
	assert.False(t, findTrigger(triggers, Claim))

	payload = testPayload{state: Opened}
	triggers, err = psm.EnumerateActiveTriggers(payload)

	assert.Nil(t, err)
	assert.Equal(t, 0, len(triggers))

	// request a state that doesn't exist in the state machine definiton and get an error thrown
	payload = testPayload{state: Claimed}
	triggers, err = psm.EnumerateActiveTriggers(payload)

	assert.NotNil(t, err)
}

func ErroringStep(pp plinko.Payload, transitionInfo plinko.TransitionInfo) (plinko.Payload, error) {
	return pp, errors.New("not-wizard")
}

func ErrorHandler(p plinko.Payload, m plinko.ModifiableTransitionInfo, e error) (plinko.Payload, error) {
	m.SetDestination("RejectedOrder")

	return p, nil
}
func TestStateMachineErrorHandling(t *testing.T) {
	const RejectedOrder plinko.State = "RejectedOrder"
	p := CreatePlinkoDefinition()

	p.Configure(NewOrder).
		OnEntry(OnNewOrderEntry).
		Permit("Submit", "PublishedOrder").
		Permit("Review", "UnderReview")

	p.Configure("PublishedOrder").
		OnEntry(ErroringStep).
		OnError(ErrorHandler).
		Permit("Submit", NewOrder)

	p.Configure("UnderReview").
		Permit("CompleteReview", "PublishedOrder").
		Permit("RejectOrder", RejectedOrder)

	p.Configure(RejectedOrder)

	transitionVisitCount := 0
	var transitionInfo plinko.TransitionInfo
	transitionInfo = nil
	p.SideEffect(func(sa plinko.StateAction, payload plinko.Payload, ti plinko.TransitionInfo) {
		transitionInfo = ti
		transitionVisitCount++
	})

	compilerOutput := p.Compile()
	psm := compilerOutput.StateMachine

	payload := testPayload{state: NewOrder}

	p1, e := psm.Fire(payload, "Submit")

	assert.NotNil(t, transitionInfo)
	assert.NotNil(t, p1)
	assert.NotNil(t, e)
	assert.Equal(t, RejectedOrder, transitionInfo.GetDestination())
	assert.Equal(t, errors.New("not-wizard"), e)

	assert.Equal(t, 3, transitionVisitCount)

}
