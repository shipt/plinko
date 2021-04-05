package sideeffects

import (
	"context"
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

	effects = append(effects, SideEffectDefinition{Filter: AllowAllSideEffects, SideEffect: func(_ context.Context, sa plinko.StateAction, p plinko.Payload, ti plinko.TransitionInfo, elapsed int64) {
		callCount++
		assert.NotNil(t, p)
		assert.NotNil(t, ti)
	}})

	effects = append(effects, SideEffectDefinition{Filter: AllowAllSideEffects, SideEffect: func(_ context.Context, sa plinko.StateAction, p plinko.Payload, ti plinko.TransitionInfo, elapsed int64) {
		callCount++
		assert.NotNil(t, p)
		assert.NotNil(t, ti)
	}})

	effects = append(effects, SideEffectDefinition{Filter: AllowAllSideEffects, SideEffect: func(_ context.Context, sa plinko.StateAction, p plinko.Payload, ti plinko.TransitionInfo, elapsed int64) {
		callCount++
		assert.NotNil(t, p)
		assert.NotNil(t, ti)
	}})

	effects = append(effects, SideEffectDefinition{Filter: plinko.AllowAfterTransition, SideEffect: func(_ context.Context, sa plinko.StateAction, p plinko.Payload, ti plinko.TransitionInfo, elapsed int64) {
		callCount++
		assert.NotNil(t, p)
		assert.NotNil(t, ti)
	}})

	payload := testPayload{}
	trInfo := TransitionDef{}

	count := Dispatch(context.TODO(), plinko.BeforeTransition, effects, payload, trInfo, 200)

	assert.Equal(t, 3, callCount)
	assert.Equal(t, 3, count)

	callCount = 0
	count = Dispatch(context.TODO(), plinko.AfterTransition, effects, payload, trInfo, 200)

	assert.Equal(t, 4, callCount)
	assert.Equal(t, 4, count)
}

func TestCallSideEffectsWithNilSet(t *testing.T) {

	result := Dispatch(context.TODO(), plinko.BeforeTransition, nil, nil, nil, 0)

	assert.True(t, result == 0)
}

func TestCallEffects(t *testing.T) {
	var effects []SideEffectDefinition
	callCount := 0

	effects = append(effects, SideEffectDefinition{Filter: AllowAllSideEffects, SideEffect: func(_ context.Context, sa plinko.StateAction, p plinko.Payload, ti plinko.TransitionInfo, em int64) {
		callCount++
		assert.NotNil(t, p)
		assert.NotNil(t, ti)
	}})

	payload := testPayload{}
	trInfo := TransitionDef{}

	result := Dispatch(context.TODO(), plinko.BeforeTransition, effects, payload, trInfo, 42)

	assert.Equal(t, result, 1)
}

func TestTransitionDefinition(t *testing.T) {
	td := TransitionDef{
		Source:      "foo1",
		Destination: "foo2",
		Trigger:     "foo3",
	}

	assert.Equal(t, plinko.State("foo1"), td.GetSource())
	assert.Equal(t, plinko.State("foo2"), td.GetDestination())
	assert.Equal(t, plinko.Trigger("foo3"), td.GetTrigger())

	td.SetDestination("foo4")
	assert.Equal(t, plinko.State("foo4"), td.GetDestination())
}

func TestStateActionToFilterMapping(t *testing.T) {
	assert.Equal(t, plinko.SideEffectFilter(1), getFilterDefinition(plinko.BeforeTransition))
	assert.Equal(t, plinko.SideEffectFilter(4), getFilterDefinition(plinko.AfterTransition))
	assert.Equal(t, plinko.SideEffectFilter(2), getFilterDefinition(plinko.BetweenStates))
	assert.Equal(t, plinko.SideEffectFilter(0), getFilterDefinition("unknown"))
}
