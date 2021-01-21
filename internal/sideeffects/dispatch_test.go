package sideeffects

import (
	"testing"

	"github.com/shipt/plinko"
	"github.com/stretchr/testify/assert"
)

type testPayload struct {
	state     plinko.State
	condition bool
}

func (p testPayload) GetState() plinko.State {
	return p.state
}

func TestCallEffects_Multiple(t *testing.T) {
	var effects []SideEffectDefinition
	callCount := 0

	effects = append(effects, SideEffectDefinition{Filter: AllowAllSideEffects, SideEffect: func(sa plinko.StateAction, p plinko.Payload, ti plinko.TransitionInfo, elapsed int64) {
		callCount++
		assert.NotNil(t, p)
		assert.NotNil(t, ti)
	}})

	effects = append(effects, SideEffectDefinition{Filter: AllowAllSideEffects, SideEffect: func(sa plinko.StateAction, p plinko.Payload, ti plinko.TransitionInfo, elapsed int64) {
		callCount++
		assert.NotNil(t, p)
		assert.NotNil(t, ti)
	}})

	effects = append(effects, SideEffectDefinition{Filter: AllowAllSideEffects, SideEffect: func(sa plinko.StateAction, p plinko.Payload, ti plinko.TransitionInfo, elapsed int64) {
		callCount++
		assert.NotNil(t, p)
		assert.NotNil(t, ti)
	}})

	effects = append(effects, SideEffectDefinition{Filter: plinko.AllowAfterTransition, SideEffect: func(sa plinko.StateAction, p plinko.Payload, ti plinko.TransitionInfo, elapsed int64) {
		callCount++
		assert.NotNil(t, p)
		assert.NotNil(t, ti)
	}})

	payload := testPayload{}
	trInfo := TransitionDef{}

	count := Dispatch(plinko.BeforeTransition, effects, payload, trInfo, 200)

	assert.Equal(t, 3, callCount)
	assert.Equal(t, 3, count)

	callCount = 0
	count = Dispatch(plinko.AfterTransition, effects, payload, trInfo, 200)

	assert.Equal(t, 4, callCount)
	assert.Equal(t, 4, count)
}

func TestCallSideEffectsWithNilSet(t *testing.T) {

	result := Dispatch(plinko.BeforeTransition, nil, nil, nil, 0)

	assert.True(t, result == 0)
}

func TestCallEffects(t *testing.T) {
	var effects []SideEffectDefinition
	callCount := 0

	effects = append(effects, SideEffectDefinition{Filter: AllowAllSideEffects, SideEffect: func(sa plinko.StateAction, p plinko.Payload, ti plinko.TransitionInfo, em int64) {
		callCount++
		assert.NotNil(t, p)
		assert.NotNil(t, ti)
	}})

	payload := testPayload{}
	trInfo := TransitionDef{}

	result := Dispatch(plinko.BeforeTransition, effects, payload, trInfo, 42)

	assert.Equal(t, result, 1)
}
