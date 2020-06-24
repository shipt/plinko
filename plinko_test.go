package plinko

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStateDefinition(t *testing.T) {
	state := stateDefinition{
		State:    "NewOrder",
		Triggers: make(map[string]*TriggerDefinition),
	}

	assert.Panics(t, func() {
		state.Permit("Submit", "PublishedOrder", "OnPublish").
			Permit("Review", "ReviewOrder", "OnReview").
			Permit("Submit", "foo", "bar")
	})

	state = stateDefinition{
		State:    "NewOrder",
		Triggers: make(map[string]*TriggerDefinition),
	}

	assert.NotPanics(t, func() {
		state.Permit("Submit", "PublishedOrder", "OnPublish").
			Permit("Review", "ReviewOrder", "OnReview")
	})

}

func TestPlinkoDefinition(t *testing.T) {
	plinko := PlinkoData{
		States: make(map[string]*stateDefinition),
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

func TestPlinkoRunner(t *testing.T) {
	/* plinkoDefinition := Plinko.CreateDefinition()

	plinko, compilerOutput, err := Plinko.Compile(plinkoDefinition)


	plinko.GetTriggers(state)
	plinko.Fire(state, "Submit", item )


	*/

}
