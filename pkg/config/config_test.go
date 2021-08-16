package config

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/shipt/plinko"
	"github.com/shipt/plinko/internal/runtime"
	"github.com/shipt/plinko/pkg/config/operation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func entryFunctionForTest(_ context.Context, pp plinko.Payload, transitionInfo plinko.TransitionInfo) (plinko.Payload, error) {
	return nil, fmt.Errorf("misc entry error")
}

func exitFunctionForTest(_ context.Context, pp plinko.Payload, transitionInfo plinko.TransitionInfo) (plinko.Payload, error) {
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

func (p *testPayload) GetState() plinko.State {
	return p.state
}

func TestOnEntryTriggerOperation(t *testing.T) {

	p := CreatePlinkoDefinition()
	counter := 0

	p.Configure(NewOrder).
		Permit("Submit", "PublishedOrder").
		Permit("Review", "UnderReview")

	p.Configure("PublishedOrder").
		OnTriggerEntry("Resupply", func(_ context.Context, pp plinko.Payload, transitionInfo plinko.TransitionInfo) (plinko.Payload, error) {
			p := pp.(*testPayload)
			p.condition = true

			counter++
			return pp, nil
		}).
		OnTriggerEntry("Resupply", func(_ context.Context, pp plinko.Payload, transitionInfo plinko.TransitionInfo) (plinko.Payload, error) {
			p := pp.(*testPayload)
			assert.True(t, p.condition)

			counter++
			return pp, nil
		}).
		Permit("Resupply", "PublishedOrder").
		Permit("Resubmit", "PublishedOrder")

	compilerOutput := p.Compile()
	psm := compilerOutput.StateMachine

	payload := &testPayload{
		state:     "PublishedOrder",
		condition: false,
	}

	p2, err := psm.Fire(context.TODO(), payload, "Resupply")

	testPayload := p2.(*testPayload)

	assert.Nil(t, err)
	assert.Equal(t, 2, counter)
	assert.True(t, testPayload.condition)

	pd, err := psm.Fire(context.TODO(), payload, "Resubmit")
	assert.Equal(t, 2, counter)
	assert.Nil(t, err)
	assert.Equal(t, plinko.State("PublishedOrder"), pd.GetState())

	_, err = psm.Fire(context.TODO(), payload, "Resupply")
	assert.Equal(t, 4, counter)
	assert.Nil(t, err)

}

func OnNewOrderEntry(_ context.Context, pp plinko.Payload, transitionInfo plinko.TransitionInfo) (plinko.Payload, error) {
	fmt.Printf("onentry: %+v", transitionInfo)
	return pp, nil
}

func OnPublishEntry_Step1(_ context.Context, pp plinko.Payload, transitionInfo plinko.TransitionInfo) (plinko.Payload, error) {
	fmt.Printf("onentry: %+v", transitionInfo)

	p := pp.(*testPayload)
	p.condition = true

	return pp, nil
}

func OnPublishEntry_Step2(_ context.Context, pp plinko.Payload, transitionInfo plinko.TransitionInfo) (plinko.Payload, error) {
	fmt.Printf("onentry: %+v", transitionInfo)

	p := pp.(*testPayload)

	if !p.condition {
		panic("error")
	}

	return pp, nil
}

func TestStateMachine(t *testing.T) {
	p := CreatePlinkoDefinition()

	p.Configure(NewOrder).
		OnEntry(OnNewOrderEntry).
		Permit("Submit", "PublishedOrder").
		Permit("Review", "UnderReview")

	p.Configure("PublishedOrder").
		OnEntry(OnPublishEntry_Step1).
		OnEntry(OnPublishEntry_Step2).
		Permit("Submit", NewOrder)

	p.Configure("UnderReview").
		Permit("CompleteReview", "PublishedOrder").
		Permit("RejectOrder", "RejectedOrder")

	p.Configure("RejectedOrder")

	visitCount := 0
	var lastStateAction plinko.StateAction
	p.SideEffect(func(_ context.Context, sa plinko.StateAction, payload plinko.Payload, ti plinko.TransitionInfo, elapsed int64) {
		visitCount += 1
		lastStateAction = sa
	})

	compilerOutput := p.Compile()
	psm := compilerOutput.StateMachine

	payload := &testPayload{state: NewOrder}

	_, err := psm.Fire(context.TODO(), payload, "Submit")

	assert.Equal(t, plinko.StateAction("AfterTransition"), lastStateAction)
	assert.Equal(t, 3, visitCount)
	assert.Nil(t, err)
}

func TestStateMachineSideEffectFiltering(t *testing.T) {
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
	var lastStateAction plinko.StateAction
	p.FilteredSideEffect(plinko.AllowAfterTransition, func(_ context.Context, sa plinko.StateAction, payload plinko.Payload, ti plinko.TransitionInfo, elapsed int64) {
		visitCount += 1
		lastStateAction = sa
	})

	compilerOutput := p.Compile()
	psm := compilerOutput.StateMachine

	payload := testPayload{state: NewOrder}

	_, err := psm.Fire(context.TODO(), &payload, "Submit")

	assert.Equal(t, plinko.StateAction("AfterTransition"), lastStateAction)
	assert.Equal(t, 1, visitCount)
	assert.Nil(t, err)
}

func TestCanFire(t *testing.T) {
	p := CreatePlinkoDefinition()

	p.Configure(Created).
		Permit(Open, Opened)

	p.Configure(Opened)

	co := p.Compile()

	psm := co.StateMachine

	payload := &testPayload{state: Created}

	assert.Nil(t, psm.CanFire(context.TODO(), payload, Open))
	assert.NotNil(t, psm.CanFire(context.TODO(), payload, Deliver))
}

func PermitIfPredicate(_ context.Context, p plinko.Payload, t plinko.TransitionInfo) error {
	tp := p.(*testPayload)

	if tp.condition {
		return nil
	}

	return errors.New("permit failed")
}

func TestCanFireWithPermitIf(t *testing.T) {
	p := CreatePlinkoDefinition()

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
	payload := &testPayload{state: Created}
	triggers, err := psm.EnumerateActiveTriggers(payload)

	assert.Nil(t, err)
	assert.True(t, findTrigger(triggers, Open))
	assert.True(t, findTrigger(triggers, Cancel))
	assert.False(t, findTrigger(triggers, Claim))

	payload = &testPayload{state: Opened}
	triggers, err = psm.EnumerateActiveTriggers(payload)

	assert.Nil(t, err)
	assert.Equal(t, 0, len(triggers))

	// request a state that doesn't exist in the state machine definiton and get an error thrown
	payload = &testPayload{state: Claimed}
	_, err = psm.EnumerateActiveTriggers(payload)

	assert.NotNil(t, err)
}

func ErroringStep(_ context.Context, pp plinko.Payload, transitionInfo plinko.TransitionInfo) (plinko.Payload, error) {
	return pp, errors.New("not-wizard")
}

func ErrorHandler(_ context.Context, p plinko.Payload, m plinko.ModifiableTransitionInfo, e error) (plinko.Payload, error) {
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
	p.SideEffect(func(_ context.Context, sa plinko.StateAction, payload plinko.Payload, ti plinko.TransitionInfo, elapsed int64) {
		transitionInfo = ti
		transitionVisitCount++
	})

	compilerOutput := p.Compile()
	psm := compilerOutput.StateMachine

	payload := &testPayload{state: NewOrder}

	p1, e := psm.Fire(context.TODO(), payload, "Submit")

	assert.NotNil(t, transitionInfo)
	assert.NotNil(t, p1)
	assert.NotNil(t, e)
	assert.Equal(t, RejectedOrder, transitionInfo.GetDestination())
	assert.Equal(t, errors.New("not-wizard"), e)

	assert.Equal(t, 2, transitionVisitCount)

}

func panickingTestOperation(c context.Context, p plinko.Payload, ti plinko.TransitionInfo) (plinko.Payload, error) {
	panic(errors.New("panics as intended"))
}

func TestStateMachinePanicSuppression(t *testing.T) {
	const StateA plinko.State = "TransA"
	const StateB plinko.State = "TransB"
	const StateC plinko.State = "TransC"
	const TransitionAB plinko.Trigger = "TransAB"
	const TransitionAC plinko.Trigger = "TransAC"
	p := CreatePlinkoDefinition()

	p.Configure(StateA).
		Permit(TransitionAB, StateB).
		Permit(TransitionAC, StateC)

	p.Configure(StateB).
		OnEntry(panickingTestOperation)

	p.Configure(StateC).
		OnEntry(func(c context.Context, p plinko.Payload, ti plinko.TransitionInfo) (plinko.Payload, error) {
			panic(errors.New("panics as intended"))
		}, operation.WithName("overridden function name"))

	psm := p.Compile().StateMachine

	_, e := psm.Fire(context.TODO(), &testPayload{state: StateA}, TransitionAB)

	require.Error(t, e)
	assert.Contains(t, e.Error(), "panickingTestOperation")

	_, e = psm.Fire(context.TODO(), &testPayload{state: StateA}, TransitionAC)

	require.Error(t, e)
	assert.Contains(t, e.Error(), "overridden function name")
}
