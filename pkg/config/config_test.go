package config

import (
	"fmt"
	"testing"

	"github.com/shipt/plinko/internal/runtime"
	"github.com/shipt/plinko/pkg/plinko"
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
	p := CreateDefinition()
	ps := p.Configure("NewOrder")

	stateDef := ps.(runtime.StateDefinition)
	assert.Nil(t, stateDef.Callbacks.OnExitFn)
	assert.Nil(t, stateDef.Callbacks.OnEntryFn)

	ps = ps.OnEntry(entryFunctionForTest)

	ps = ps.OnExit(exitFunctionForTest)

	stateDef = ps.(runtime.StateDefinition)
	assert.NotNil(t, stateDef.Callbacks.OnExitFn)
	assert.NotNil(t, stateDef.Callbacks.OnEntryFn)

	assert.Equal(t, "github.com/shipt/plinko/pkg/config.entryFunctionForTest", stateDef.Callbacks.EntryFunctionChain[0])
	assert.Equal(t, "github.com/shipt/plinko/pkg/config.exitFunctionForTest", stateDef.Callbacks.ExitFunctionChain[0])
}

func TestPlinkoAsInterface(t *testing.T) {
	p := CreateDefinition()

	p.Configure("NewOrder").
		Permit("Submit", "PublishedOrder").
		Permit("Review", "ReviewedOrder")
}

func TestUndefinedStateCompile(t *testing.T) {
	p := CreateDefinition()

	p.Configure(NewOrder).
		Permit("Submit", "PublishedOrder")

	compilerOutput := p.Compile()
	assert.Equal(t, 1, len(compilerOutput.Messages))
	assert.Equal(t, plinko.CompileError, compilerOutput.Messages[0].CompileMessage)
	assert.Equal(t, "State 'PublishedOrder' undefined: Trigger 'Submit' declares a transition to this undefined state.", compilerOutput.Messages[0].Message)
}

func TestTriggerlessStateCompile(t *testing.T) {
	p := CreateDefinition()

	p.Configure(NewOrder).
		Permit("Submit", "PublishedOrder")
	p.Configure("PublishedOrder")

	compilerOutput := p.Compile()
	assert.Equal(t, 1, len(compilerOutput.Messages))
	assert.Equal(t, plinko.CompileWarning, compilerOutput.Messages[0].CompileMessage)
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

func TestPlinkoDefinition(t *testing.T) {
	stateMap := make(map[plinko.State]*runtime.StateDefinition)
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
