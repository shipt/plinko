package plinko

import (
	"fmt"
	"testing"

	"github.com/shipt/plinko/definitions"
	"github.com/shipt/plinko/interfaces"
	"github.com/shipt/plinko/types"
	"github.com/stretchr/testify/assert"
)

func TestStateDefinition(t *testing.T) {
	state := stateDefinition{
		State:    "NewOrder",
		Triggers: make(map[types.Trigger]*TriggerDefinition),
	}

	assert.Panics(t, func() {
		state.Permit("Submit", "PublishedOrder").
			Permit("Review", "ReviewOrder").
			Permit("Submit", "foo")
	})

	state = stateDefinition{
		State:    "NewOrder",
		Triggers: make(map[types.Trigger]*TriggerDefinition),
	}

}

func TestPlinkoDefinition(t *testing.T) {
	stateMap := make(map[types.State]*stateDefinition)
	plinko := plinkoDefinition{
		States: &stateMap,
	}

	assert.NotPanics(t, func() {
		plinko.CreateState("NewOrder").
			//			OnEntry()
			//			OnExit()
			Permit("Submit", "PublishedOrder").
			Permit("Review", "ReviewOrder")

		plinko.CreateState("PublishedOrder")
		plinko.CreateState("ReviewOrder")
	})

	assert.Panics(t, func() {
		plinko.CreateState("NewOrder").
			Permit("Submit", "PublishedOrder").
			Permit("Review", "ReviewOrder")

		plinko.CreateState("PublishedOrder")
		plinko.CreateState("ReviewOrder")
		plinko.CreateState("NewOrder")
	})
}

func TestPlinkoAsInterface(t *testing.T) {
	p := CreateDefinition()

	p.CreateState("NewOrder").
		Permit("Submit", "PublishedOrder").
		Permit("Review", "ReviewedOrder")
}

func TestEntryAndExitFunctions(t *testing.T) {
	p := CreateDefinition()
	ps := p.CreateState(NewOrder)

	stateDef := ps.(stateDefinition)
	assert.Nil(t, stateDef.callbacks.OnExitFn)
	assert.Nil(t, stateDef.callbacks.OnEntryFn)

	ps = ps.OnEntry(func(pp interfaces.PlinkoPayload, transitionInfo definitions.TransitionInfo) (interfaces.PlinkoPayload, error) {
		return nil, fmt.Errorf("misc error")
	})

	ps = ps.OnExit(func(pp interfaces.PlinkoPayload, transitionInfo definitions.TransitionInfo) (interfaces.PlinkoPayload, error) {
		return nil, fmt.Errorf("misc error")
	})

	stateDef = ps.(stateDefinition)
	assert.NotNil(t, stateDef.callbacks.OnExitFn)
	assert.NotNil(t, stateDef.callbacks.OnEntryFn)
}

func TestUndefinedStateCompile(t *testing.T) {
	p := CreateDefinition()

	p.CreateState(NewOrder).
		Permit("Submit", "PublishedOrder")

	compilerOutput := p.Compile()
	assert.Equal(t, 1, len(compilerOutput.Messages))
	assert.Equal(t, CompileError, compilerOutput.Messages[0].CompileMessage)
	assert.Equal(t, "State 'PublishedOrder' undefined: Trigger 'Submit' declares a transition to this undefined state.", compilerOutput.Messages[0].Message)
}

func TestTriggerlessStateCompile(t *testing.T) {
	p := CreateDefinition()

	p.CreateState(NewOrder).
		Permit("Submit", "PublishedOrder")
	p.CreateState("PublishedOrder")

	compilerOutput := p.Compile()
	assert.Equal(t, 1, len(compilerOutput.Messages))
	assert.Equal(t, CompileWarning, compilerOutput.Messages[0].CompileMessage)
	assert.Equal(t, "State 'PublishedOrder' is a state without any triggers (deadend state).", compilerOutput.Messages[0].Message)
}

func TestUmlDiagramming(t *testing.T) {
	p := CreateDefinition()

	p.CreateState(NewOrder).
		Permit("Submit", "PublishedOrder").
		Permit("Review", "UnderReview")

	p.CreateState("PublishedOrder")

	p.CreateState("UnderReview").
		Permit("CompleteReview", "PublishedOrder").
		Permit("RejectOrder", "RejectedOrder")

	p.CreateState("RejectedOrder")

	uml, err := p.RenderUml()

	fmt.Println(uml)

	assert.Nil(t, err)
	assert.Equal(t, "@startuml\n[*] -> NewOrder \nNewOrder", string(uml)[0:35])
	assert.Equal(t, "\n@enduml", string(uml)[len(uml)-8:])
}

type testPayload struct {
	state types.State
}

func (p testPayload) GetState() types.State {
	return p.state
}

func (p testPayload) PutState(state types.State) {
	p.state = state
}

func TestStateMachine(t *testing.T) {
	p := CreateDefinition()

	p.CreateState(NewOrder).
		OnEntry(OnNewOrderEntry).
		Permit("Submit", "PublishedOrder").
		Permit("Review", "UnderReview")

	p.CreateState("PublishedOrder").
		OnEntry(OnNewOrderEntry).
		Permit("Submit", NewOrder)

	p.CreateState("UnderReview").
		Permit("CompleteReview", "PublishedOrder").
		Permit("RejectOrder", "RejectedOrder")

	p.CreateState("RejectedOrder")

	compilerOutput := p.Compile()
	psm := compilerOutput.PlinkoStateMachine

	payload := testPayload{state: NewOrder}

	psm.Fire(payload, "Submit")

}

func TestJonathanDiagramming(t *testing.T) {
	p := CreateDefinition()

	p.CreateState(Created).
		Permit(Open, Opened).
		Permit(Cancel, Canceled)

	p.CreateState(Opened).
		Permit("AddItemToOrder", Opened).
		Permit(Claim, Claimed).
		Permit(Cancel, Canceled)

	p.CreateState(Claimed).
		Permit("AddItemToOrder", Claimed).
		Permit(Submit, ArriveAtStore).
		Permit(Cancel, Canceled)

	p.CreateState(ArriveAtStore).
		Permit(Submit, MarkedAsPickedUp).
		Permit(Cancel, Canceled)

	p.CreateState(MarkedAsPickedUp).
		Permit(Deliver, Delivered).
		Permit(Cancel, Canceled)

	p.CreateState(Delivered).
		Permit(Return, Returned)

	p.CreateState(Canceled).
		Permit(Reinstate, Created)
	p.CreateState(Returned)

	co := p.Compile()
	fmt.Printf("%+v\n", co.Messages)
	uml, err := p.RenderUml()

	fmt.Println(err)

	fmt.Println(uml)

}

const (
	NewOrder types.State = "NewOrder"
	Reviewed types.State = "Reviewed"
)

func IsPlatform(pp interfaces.PlinkoPayload) bool {
	return true
}

func OnNewOrderEntry(pp interfaces.PlinkoPayload, transitionInfo definitions.TransitionInfo) (interfaces.PlinkoPayload, error) {
	fmt.Printf("onentry: %+v", transitionInfo)
	return pp, nil
}

const Created types.State = "Created"
const Opened types.State = "Opened"
const Claimed types.State = "Claimed"
const ArriveAtStore types.State = "ArrivedAtStore"
const MarkedAsPickedUp types.State = "MarkedAsPickedup"
const Delivered types.State = "Delivered"
const Canceled types.State = "Canceled"
const Returned types.State = "Returned"

const Submit types.Trigger = "Submit"
const Cancel types.Trigger = "Cancel"
const Open types.Trigger = "Open"
const Claim types.Trigger = "Claim"
const Deliver types.Trigger = "Deliver"
const Return types.Trigger = "Return"
const Reinstate types.Trigger = "Reinstate"
