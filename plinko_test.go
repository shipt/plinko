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

const (
	NewOrder State = "NewOrder"
	Reviewed State = "Reviewed"
)

func IsPlatform(pp PlinkoPayload) bool {
	return true
}

func OnNewOrderEntry(pp *PlinkoPayload) (*PlinkoPayload, error) {
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
