package runtime

import (
	"testing"

	"github.com/shipt/plinko"
	"github.com/stretchr/testify/assert"
)

const (
	NewOrder plinko.State = "NewOrder"
)

func TestStateDefinition(t *testing.T) {
	state := InternalStateDefinition{
		State:    "NewOrder",
		Triggers: make(map[plinko.Trigger]*TriggerDefinition),
	}

	assert.Panics(t, func() {
		state.Permit("Submit", "PublishedOrder").
			Permit("Review", "ReviewOrder").
			Permit("Submit", "foo")
	})

}

func TestStateRedeclarationPanic(t *testing.T) {
	p := createPlinkoDefinition()

	p.SideEffect(nil)
	p.FilteredSideEffect(plinko.AllowAfterTransition, nil)

	p.Configure("Open")
	assert.Panics(t, func() { p.Configure("Open") })

	assert.Panics(t, func() {
		p.Configure("Close").
			Permit("Go", "Open").
			Permit("Go", "Open")
	})
}
