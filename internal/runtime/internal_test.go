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

type testPayload struct {
	state     plinko.State
	condition bool
}

func (p testPayload) GetState() plinko.State {
	return p.state
}

func TestCallEffects_Multiple(t *testing.T) {
	var effects []sideEffectDefinition
	callCount := 0

	effects = append(effects, sideEffectDefinition{Filter: allowAllSideEffects, SideEffect: func(sa plinko.StateAction, p plinko.Payload, ti plinko.TransitionInfo) {
		callCount++
		assert.NotNil(t, p)
		assert.NotNil(t, ti)
	}})

	effects = append(effects, sideEffectDefinition{Filter: allowAllSideEffects, SideEffect: func(sa plinko.StateAction, p plinko.Payload, ti plinko.TransitionInfo) {
		callCount++
		assert.NotNil(t, p)
		assert.NotNil(t, ti)
	}})

	effects = append(effects, sideEffectDefinition{Filter: allowAllSideEffects, SideEffect: func(sa plinko.StateAction, p plinko.Payload, ti plinko.TransitionInfo) {
		callCount++
		assert.NotNil(t, p)
		assert.NotNil(t, ti)
	}})

	effects = append(effects, sideEffectDefinition{Filter: plinko.AllowAfterTransition, SideEffect: func(sa plinko.StateAction, p plinko.Payload, ti plinko.TransitionInfo) {
		callCount++
		assert.NotNil(t, p)
		assert.NotNil(t, ti)
	}})

	payload := testPayload{}
	trInfo := transitionDef{}

	count := callSideEffects(plinko.BeforeTransition, effects, payload, trInfo)

	assert.Equal(t, 3, callCount)
	assert.Equal(t, 3, count)

	callCount = 0
	count = callSideEffects(plinko.AfterTransition, effects, payload, trInfo)

	assert.Equal(t, 4, callCount)
	assert.Equal(t, 4, count)
}

func TestCallSideEffectsWithNilSet(t *testing.T) {

	result := callSideEffects(plinko.BeforeTransition, nil, nil, nil)

	assert.True(t, result == 0)
}

func TestCallEffects(t *testing.T) {
	var effects []sideEffectDefinition
	callCount := 0

	effects = append(effects, sideEffectDefinition{Filter: allowAllSideEffects, SideEffect: func(sa plinko.StateAction, p plinko.Payload, ti plinko.TransitionInfo) {
		callCount++
		assert.NotNil(t, p)
		assert.NotNil(t, ti)
	}})

	payload := testPayload{}
	trInfo := transitionDef{}

	result := callSideEffects(plinko.BeforeTransition, effects, payload, trInfo)

	assert.Equal(t, result, 1)
}
