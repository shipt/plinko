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
		state.Permit("Submit", "PublishedOrder", "OnPublish").
			Permit("Review", "ReviewOrder", "OnReview").
			Permit("Submit", "foo", "bar")
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
			Permit("Submit", "PublishedOrder", "OnPublish").
			Permit("Review", "ReviewOrder", "OnReview")

		plinko.CreateState("PublishedOrder")
		plinko.CreateState("ReviewOrder")
	})

	assert.Panics(t, func() {
		plinko.CreateState("NewOrder").
			Permit("Submit", "PublishedOrder", "OnPublish").
			Permit("Review", "ReviewOrder", "OnReview")

		plinko.CreateState("PublishedOrder")
		plinko.CreateState("ReviewOrder")
		plinko.CreateState("NewOrder")
	})
}

func TestPlinkoAsInterface(t *testing.T) {
	p := CreateDefinition()

	p.CreateState("NewOrder").
		Permit("Submit", "PublishedOrder", "OnPublish").
		Permit("Review", "ReviewedOrder", "OnReview")
}

func TestEntryAndExitFunctions(t *testing.T) {
	p := CreateDefinition()
	ps := p.CreateState(NewOrder)

	stateDef := ps.(stateDefinition)
	assert.Nil(t, stateDef.OnExitFn)
	assert.Nil(t, stateDef.OnEntryFn)

	ps = ps.OnEntry(func(pp *PlinkoPayload, transitionInfo TransitionInfo) (*PlinkoPayload, error) {
		return nil, fmt.Errorf("misc error")
	})

	ps = ps.OnExit(func(pp *PlinkoPayload, transitionInfo TransitionInfo) (*PlinkoPayload, error) {
		return nil, fmt.Errorf("misc error")
	})

	stateDef = ps.(stateDefinition)
	assert.NotNil(t, stateDef.OnExitFn)
	assert.NotNil(t, stateDef.OnEntryFn)
}

func TestUndefinedStateCompile(t *testing.T) {
	p := CreateDefinition()

	p.CreateState(NewOrder).
		Permit("Submit", "PublishedOrder", "OnPublish")

	messages := p.Compile()
	assert.Equal(t, 1, len(messages))
	assert.Equal(t, CompileError, messages[0].CompileMessage)
	assert.Equal(t, "State 'PublishedOrder' undefined: Trigger 'Submit' declares a transition to this undefined state.", messages[0].Message)
}

func TestTriggerlessStateCompile(t *testing.T) {
	p := CreateDefinition()

	p.CreateState(NewOrder).
		Permit("Submit", "PublishedOrder", "OnPublish")
	p.CreateState("PublishedOrder")

	messages := p.Compile()
	assert.Equal(t, 1, len(messages))
	assert.Equal(t, CompileWarning, messages[0].CompileMessage)
	assert.Equal(t, "State 'PublishedOrder' is a state without any triggers (deadend state).", messages[0].Message)

}

func TestUmlDiagramming(t *testing.T) {
	p := CreateDefinition()

	p.CreateState(NewOrder).
		Permit("Submit", "PublishedOrder", "OnPublish").
		Permit("Review", "UnderReview", "OnReview")

	p.CreateState("PublishedOrder")

	p.CreateState("UnderReview").
		Permit("CompleteReview", "PublishedOrder", "OnCompletedReview").
		Permit("RejectOrder", "RejectedOrder", "OnRejectOrder")

	p.CreateState("RejectedOrder")

	uml, err := p.RenderUml()

	fmt.Println(uml)

	assert.Nil(t, err)
	assert.Equal(t, "@startuml\n[*] -> NewOrder \nNewOrder", string(uml)[0:35])
	assert.Equal(t, "\n@enduml", string(uml)[len(uml)-8:])
}

const (
	NewOrder State = "NewOrder"
	Reviewed State = "Reviewed"
)

func IsPlatform(pp PlinkoPayload) bool {
	return true
}

func OnNewOrderEntry(pp *PlinkoPayload, transitionInfo TransitionInfo) (*PlinkoPayload, error) {
	return pp, nil
}

func TestPlinkoRunner(t *testing.T) {

	pd := CreateDefinition()

	pd.CreateState("NewOrder").
		OnEntry(OnNewOrderEntry).
		Permit("Submit", "PublishedOrder", "OnPublish").
		Permit("Review", "ReviewedOrder", "OnReview")

	/* plinkoDefinition := Plinko.CreateDefinition()

	plinkoDefinition.
		State("foo").
		PermitIf("trigger", "state", func() { return state.IsValidNumber() }).
		PermitIf("trigger", "state2", func() { return !state.IsValidNumber() }).
		PermitReentry("trigger")



	plinko, compilerOutput, err := Plinko.Compile(plinkoDefinition)


	plinko.GetTriggers(state)
	plinko.Fire(state, "Submit", item )


	*/

}
