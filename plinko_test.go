package plinko

import (
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

	assert.NotPanics(t, func() {
		state.Permit("Submit", "PublishedOrder", "OnPublish").
			Permit("Review", "ReviewOrder", "OnReview")
	})

}

func TestPlinkoDefinition(t *testing.T) {
	plinko := plinkoDefinition{
		States: make(map[State]*stateDefinition),
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

const (
	NewOrder State = "NewOrder"
	Reviewed State = "Reviewed"
)

func TestPlinkoRunner(t *testing.T) {

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
