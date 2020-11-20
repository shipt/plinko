package runtime

import (
	"testing"

	"github.com/shipt/plinko/pkg/plinko"
	"github.com/stretchr/testify/assert"
)

const (
	NewOrder plinko.State = "NewOrder"
)

func TestStateDefinition(t *testing.T) {
	state := StateDefinition{
		State:    "NewOrder",
		Triggers: make(map[plinko.Trigger]*TriggerDefinition),
	}

	assert.Panics(t, func() {
		state.Permit("Submit", "PublishedOrder").
			Permit("Review", "ReviewOrder").
			Permit("Submit", "foo")
	})

}
