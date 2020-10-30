package plinko

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStateDefinition(t *testing.T) {
	state := stateDefinition{
		State:    "NewOrder",
		Triggers: make(map[Trigger]*TriggerDefinition),
	}

	assert.Panics(t, func() {
		state.Permit("Submit", "PublishedOrder").
			Permit("Review", "ReviewOrder").
			Permit("Submit", "foo")
	})

	state = stateDefinition{
		State:    "NewOrder",
		Triggers: make(map[Trigger]*TriggerDefinition),
	}

}

func TestStateActionToFilterConversion(t *testing.T) {
	assert.Equal(t, AllowBeforeStateExit, getFilterDefinition(BeforeStateExit))
	assert.Equal(t, AllowAfterStateExit, getFilterDefinition(AfterStateExit))
	assert.Equal(t, AllowBeforeStateEntry, getFilterDefinition(BeforeStateEntry))
	assert.Equal(t, AllowAfterStateEntry, getFilterDefinition(AfterStateEntry))
}

func TestPlinkoDefinition(t *testing.T) {
	stateMap := make(map[State]*stateDefinition)
	plinko := plinkoDefinition{
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

func TestPlinkoAsInterface(t *testing.T) {
	p := CreateDefinition()

	p.Configure("NewOrder").
		Permit("Submit", "PublishedOrder").
		Permit("Review", "ReviewedOrder")
}

func entryFunctionForTest(pp Payload, transitionInfo TransitionInfo) (Payload, error) {
	return nil, fmt.Errorf("misc entry error")
}

func exitFunctionForTest(pp Payload, transitionInfo TransitionInfo) (Payload, error) {
	return nil, fmt.Errorf("misc exit error")
}

func TestEntryAndExitFunctions(t *testing.T) {
	p := CreateDefinition()
	ps := p.Configure(NewOrder)

	stateDef := ps.(stateDefinition)
	assert.Nil(t, stateDef.callbacks.OnExitFn)
	assert.Nil(t, stateDef.callbacks.OnEntryFn)

	ps = ps.OnEntry(entryFunctionForTest)

	ps = ps.OnExit(exitFunctionForTest)

	stateDef = ps.(stateDefinition)
	assert.NotNil(t, stateDef.callbacks.OnExitFn)
	assert.NotNil(t, stateDef.callbacks.OnEntryFn)

	assert.Equal(t, "github.com/shipt/plinko.entryFunctionForTest", stateDef.callbacks.EntryFunctionChain[0])
	assert.Equal(t, "github.com/shipt/plinko.exitFunctionForTest", stateDef.callbacks.ExitFunctionChain[0])
}

func TestUndefinedStateCompile(t *testing.T) {
	p := CreateDefinition()

	p.Configure(NewOrder).
		Permit("Submit", "PublishedOrder")

	compilerOutput := p.Compile()
	assert.Equal(t, 1, len(compilerOutput.Messages))
	assert.Equal(t, CompileError, compilerOutput.Messages[0].CompileMessage)
	assert.Equal(t, "State 'PublishedOrder' undefined: Trigger 'Submit' declares a transition to this undefined state.", compilerOutput.Messages[0].Message)
}

func TestTriggerlessStateCompile(t *testing.T) {
	p := CreateDefinition()

	p.Configure(NewOrder).
		Permit("Submit", "PublishedOrder")
	p.Configure("PublishedOrder")

	compilerOutput := p.Compile()
	assert.Equal(t, 1, len(compilerOutput.Messages))
	assert.Equal(t, CompileWarning, compilerOutput.Messages[0].CompileMessage)
	assert.Equal(t, "State 'PublishedOrder' is a state without any triggers (deadend state).", compilerOutput.Messages[0].Message)
}

func TestUmlDiagramming(t *testing.T) {
	p := CreateDefinition()

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

func TestCallSideEffectsWithNilSet(t *testing.T) {

	result := callSideEffects(BeforeStateExit, nil, nil, nil)

	assert.True(t, result == 0)
}

func TestCallEffects(t *testing.T) {
	var effects []sideEffectDefinition
	callCount := 0

	effects = append(effects, sideEffectDefinition{Filter: allowAllSideEffects, SideEffect: func(sa StateAction, p Payload, ti TransitionInfo) {
		callCount++
		assert.NotNil(t, p)
		assert.NotNil(t, ti)
	}})

	payload := testPayload{}
	trInfo := transitionDef{}

	result := callSideEffects(BeforeStateExit, effects, payload, trInfo)

	assert.Equal(t, result, 1)
}

func TestCallEffects_Multiple(t *testing.T) {
	var effects []sideEffectDefinition
	callCount := 0

	effects = append(effects, sideEffectDefinition{Filter: allowAllSideEffects, SideEffect: func(sa StateAction, p Payload, ti TransitionInfo) {
		callCount++
		assert.NotNil(t, p)
		assert.NotNil(t, ti)
	}})

	effects = append(effects, sideEffectDefinition{Filter: allowAllSideEffects, SideEffect: func(sa StateAction, p Payload, ti TransitionInfo) {
		callCount++
		assert.NotNil(t, p)
		assert.NotNil(t, ti)
	}})

	effects = append(effects, sideEffectDefinition{Filter: allowAllSideEffects, SideEffect: func(sa StateAction, p Payload, ti TransitionInfo) {
		callCount++
		assert.NotNil(t, p)
		assert.NotNil(t, ti)
	}})

	effects = append(effects, sideEffectDefinition{Filter: AllowAfterStateExit, SideEffect: func(sa StateAction, p Payload, ti TransitionInfo) {
		callCount++
		assert.NotNil(t, p)
		assert.NotNil(t, ti)
	}})

	payload := testPayload{}
	trInfo := transitionDef{}

	count := callSideEffects(BeforeStateExit, effects, payload, trInfo)

	assert.Equal(t, 3, callCount)
	assert.Equal(t, 3, count)

	callCount = 0
	count = callSideEffects(AfterStateExit, effects, payload, trInfo)

	assert.Equal(t, 4, callCount)
	assert.Equal(t, 4, count)
}

type testPayload struct {
	state State
}

func (p testPayload) GetState() State {
	return p.state
}

func (p testPayload) PutState(state State) {
	p.state = state
}

func TestStateMachine(t *testing.T) {
	p := CreateDefinition()

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
	p.SideEffect(func(sa StateAction, payload Payload, ti TransitionInfo) {
		visitCount += 1
	})

	compilerOutput := p.Compile()
	psm := compilerOutput.StateMachine

	payload := testPayload{state: NewOrder}

	psm.Fire(payload, "Submit")

	assert.Equal(t, 4, visitCount)

}

func TestCanFire(t *testing.T) {
	p := CreateDefinition()

	p.Configure(Created).
		Permit(Open, Opened)

	p.Configure(Opened)

	co := p.Compile()

	psm := co.StateMachine

	payload := testPayload{state: Created}

	assert.True(t, psm.CanFire(payload, Open))
	assert.False(t, psm.CanFire(payload, Deliver))
}

func findTrigger(triggers []Trigger, trigger Trigger) bool {
	for _, v := range triggers {
		if v == trigger {
			return true
		}
	}

	return false
}
func TestEnumerateTriggers(t *testing.T) {
	p := CreateDefinition()

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

func TestDiagramming(t *testing.T) {
	p := CreateDefinition()

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

const (
	NewOrder State = "NewOrder"
	Reviewed State = "Reviewed"
)

func IsPlatform(pp Payload) bool {
	return true
}

func OnNewOrderEntry(pp Payload, transitionInfo TransitionInfo) (Payload, error) {
	fmt.Printf("onentry: %+v", transitionInfo)
	return pp, nil
}

const Created State = "Created"
const Opened State = "Opened"
const Claimed State = "Claimed"
const ArriveAtStore State = "ArrivedAtStore"
const MarkedAsPickedUp State = "MarkedAsPickedup"
const Delivered State = "Delivered"
const Canceled State = "Canceled"
const Returned State = "Returned"

const Submit Trigger = "Submit"
const Cancel Trigger = "Cancel"
const Open Trigger = "Open"
const Claim Trigger = "Claim"
const Deliver Trigger = "Deliver"
const Return Trigger = "Return"
const Reinstate Trigger = "Reinstate"
