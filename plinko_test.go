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

func TestPlinkoDefinition(t *testing.T) {
	stateMap := make(map[State]*stateDefinition)
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

	ps = ps.OnEntry(func(pp PlinkoPayload, transitionInfo TransitionInfo) (PlinkoPayload, error) {
		return nil, fmt.Errorf("misc error")
	})

	ps = ps.OnExit(func(pp PlinkoPayload, transitionInfo TransitionInfo) (PlinkoPayload, error) {
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

const (
	NewOrder State = "NewOrder"
	Reviewed State = "Reviewed"
)

func IsPlatform(pp PlinkoPayload) bool {
	return true
}

func OnNewOrderEntry(pp PlinkoPayload, transitionInfo TransitionInfo) (PlinkoPayload, error) {
	fmt.Printf("onentry: %+v", transitionInfo)
	return pp, nil
}
