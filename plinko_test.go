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
		state.AddTrigger("Submit", "PublishedOrder", "OnPublish").
			AddTrigger("Review", "ReviewOrder", "OnReview").
			AddTrigger("Submit", "foo", "bar")
	})

	state = stateDefinition{
		State:    "NewOrder",
		Triggers: make(map[string]*TriggerDefinition),
	}

	assert.NotPanics(t, func() {
		state.AddTrigger("Submit", "PublishedOrder", "OnPublish").
			AddTrigger("Review", "ReviewOrder", "OnReview")
	})

}

func TestPlinkoDefinition(t *testing.T) {
	plinko := PlinkoData{
		States: make(map[string]*stateDefinition),
	}

	assert.NotPanics(t, func() {
		plinko.CreateState("NewOrder").
			AddTrigger("Submit", "PublishedOrder", "OnPublish").
			AddTrigger("Review", "ReviewOrder", "OnReview")

		plinko.CreateState("PublishedOrder")
		plinko.CreateState("ReviewOrder")
	})

	assert.Panics(t, func() {
		plinko.CreateState("NewOrder").
			AddTrigger("Submit", "PublishedOrder", "OnPublish").
			AddTrigger("Review", "ReviewOrder", "OnReview")

		plinko.CreateState("PublishedOrder")
		plinko.CreateState("ReviewOrder")
		plinko.CreateState("NewOrder")
	})

}
