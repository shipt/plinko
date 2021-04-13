package runtime

import (
	"context"
	"fmt"
	"reflect"
	"runtime"

	"github.com/shipt/plinko"
	"github.com/shipt/plinko/internal/composition"
	"github.com/shipt/plinko/internal/sideeffects"
)

type plinkoStateMachine struct {
	pd PlinkoDefinition
}

type InternalStateDefinition struct {
	State    plinko.State
	Triggers map[plinko.Trigger]*TriggerDefinition
	info     plinko.StateConfig

	Callbacks *composition.CallbackDefinitions

	Abs *AbstractSyntax
}

func (sd InternalStateDefinition) OnEntry(entryFn plinko.Operation, opts ...plinko.OperationOption) plinko.StateDefinition {
	sd.Callbacks.AddEntry(nil, entryFn, newOperationConfig(entryFn, opts...))

	return sd
}

func (sd InternalStateDefinition) OnError(errorFn plinko.ErrorOperation, opts ...plinko.OperationOption) plinko.StateDefinition {
	sd.Callbacks.AddError(errorFn, newOperationConfig(errorFn, opts...))
	return sd
}

func (sd InternalStateDefinition) OnExit(exitFn plinko.Operation, opts ...plinko.OperationOption) plinko.StateDefinition {
	sd.Callbacks.AddExit(nil, exitFn, newOperationConfig(exitFn, opts...))

	return sd
}

func (sd InternalStateDefinition) OnTriggerEntry(trigger plinko.Trigger, entryFn plinko.Operation, opts ...plinko.OperationOption) plinko.StateDefinition {
	sd.Callbacks.AddEntry(func(_ context.Context, _ plinko.Payload, t plinko.TransitionInfo) error {
		if t.GetTrigger() == trigger {
			return nil
		}

		return fmt.Errorf("trigger '%s' not found for entry", trigger)
	}, entryFn, newOperationConfig(entryFn, opts...))

	return sd

}

func (sd InternalStateDefinition) OnTriggerExit(trigger plinko.Trigger, exitFn plinko.Operation, opts ...plinko.OperationOption) plinko.StateDefinition {
	sd.Callbacks.AddExit(func(_ context.Context, _ plinko.Payload, t plinko.TransitionInfo) error {
		if t.GetTrigger() == trigger {
			return nil
		}

		return fmt.Errorf("trigger '%s' not found for exit", trigger)
	}, exitFn, newOperationConfig(exitFn, opts...))

	return sd
}

func (sd InternalStateDefinition) Permit(trigger plinko.Trigger, destinationState plinko.State) plinko.StateDefinition {
	addPermit(&sd, trigger, destinationState, nil)

	return sd
}

func (sd InternalStateDefinition) PermitIf(predicate plinko.Predicate, trigger plinko.Trigger, destinationState plinko.State) plinko.StateDefinition {
	addPermit(&sd, trigger, destinationState, predicate)

	return sd
}

type AbstractSyntax struct {
	States             []plinko.State
	TriggerDefinitions []TriggerDefinition
	StateDefinitions   []*InternalStateDefinition
}

type PlinkoDefinition struct {
	States      *map[plinko.State]*InternalStateDefinition
	SideEffects []sideeffects.SideEffectDefinition
	Abs         AbstractSyntax
}

func findDestinationState(states []plinko.State, searchState plinko.State) bool {
	for _, searchVal := range states {
		if searchVal == searchState {
			return true
		}
	}

	return false
}

func (pd *PlinkoDefinition) SideEffect(sideEffect plinko.SideEffect) plinko.PlinkoDefinition {
	pd.SideEffects = append(pd.SideEffects, sideeffects.SideEffectDefinition{Filter: sideeffects.AllowAllSideEffects, SideEffect: sideEffect})

	return pd
}

func (pd *PlinkoDefinition) FilteredSideEffect(filter plinko.SideEffectFilter, sideEffect plinko.SideEffect) plinko.PlinkoDefinition {
	pd.SideEffects = append(pd.SideEffects, sideeffects.SideEffectDefinition{Filter: filter, SideEffect: sideEffect})

	return pd
}

func (pd *PlinkoDefinition) Configure(state plinko.State, opts ...plinko.StateOption) plinko.StateDefinition {
	if _, ok := (*pd.States)[state]; ok {
		panic(fmt.Sprintf("State: %s - has already been defined, plinko configuration invalid.", state))
	}

	cbd := composition.CallbackDefinitions{}

	sd := InternalStateDefinition{
		State:     state,
		Triggers:  make(map[plinko.Trigger]*TriggerDefinition),
		Abs:       &pd.Abs,
		Callbacks: &cbd,
		info:      newStateConfig(state, opts...),
	}

	(*pd.States)[state] = &sd

	pd.Abs.States = append(pd.Abs.States, state)
	pd.Abs.StateDefinitions = append(pd.Abs.StateDefinitions, &sd)

	return sd
}

type compileInfo struct {
}

type TriggerDefinition struct {
	Name             plinko.Trigger
	DestinationState plinko.State
	Predicate        func(context.Context, plinko.Payload, plinko.TransitionInfo) error
}

type PlinkoDataStructure struct {
	States map[plinko.State]plinko.StateDefinition
}

func addPermit(sd *InternalStateDefinition, trigger plinko.Trigger, destination plinko.State, predicate func(context.Context, plinko.Payload, plinko.TransitionInfo) error) {
	if _, ok := sd.Triggers[trigger]; ok {
		panic(fmt.Sprintf("Trigger: %s - has already been defined, plinko configuration invalid.", trigger))
	}

	td := TriggerDefinition{
		Name:             trigger,
		DestinationState: destination,
		Predicate:        predicate,
	}

	sd.Triggers[trigger] = &td
	sd.Abs.TriggerDefinitions = append(sd.Abs.TriggerDefinitions, td)
}

func newOperationConfig(op interface{}, opts ...plinko.OperationOption) plinko.OperationConfig {
	c := plinko.OperationConfig{
		Name: getFunctionName(op),
	}

	for _, opt := range opts {
		opt(&c)
	}

	return c
}

func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func newStateConfig(state plinko.State, opts ...plinko.StateOption) plinko.StateConfig {
	c := plinko.StateConfig{
		Name: string(state),
	}

	for _, opt := range opts {
		opt(&c)
	}

	return c
}
