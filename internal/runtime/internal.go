package runtime

import (
	"fmt"
	"reflect"
	"runtime"

	"github.com/shipt/plinko/internal/composition"
	"github.com/shipt/plinko/internal/sideeffects"
	"github.com/shipt/plinko/pkg/plinko"
)

type plinkoStateMachine struct {
	pd PlinkoDefinition
}

type InternalStateDefinition struct {
	State    plinko.State
	Triggers map[plinko.Trigger]*TriggerDefinition

	Callbacks *composition.CallbackDefinitions

	Abs *AbstractSyntax
}

func (sd InternalStateDefinition) OnEntry(entryFn plinko.Operation) plinko.StateDefinition {
	sd.Callbacks.AddEntry(nil, entryFn)

	return sd

}

func (sd InternalStateDefinition) OnExit(exitFn plinko.Operation) plinko.StateDefinition {
	sd.Callbacks.AddExit(nil, exitFn)

	return sd
}

func (sd InternalStateDefinition) OnTriggerEntry(trigger plinko.Trigger, entryFn plinko.Operation) plinko.StateDefinition {
	sd.Callbacks.AddEntry(func(_ plinko.Payload, t plinko.TransitionInfo) bool {
		return t.GetTrigger() == trigger
	}, entryFn)

	return sd

}

func (sd InternalStateDefinition) OnTriggerExit(trigger plinko.Trigger, exitFn plinko.Operation) plinko.StateDefinition {
	sd.Callbacks.AddExit(func(_ plinko.Payload, t plinko.TransitionInfo) bool {
		return t.GetTrigger() == trigger
	}, exitFn)

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

func (pd *PlinkoDefinition) Configure(state plinko.State) plinko.StateDefinition {
	if _, ok := (*pd.States)[state]; ok {
		panic(fmt.Sprintf("State: %s - has already been defined, plinko configuration invalid.", state))
	}

	cbd := composition.CallbackDefinitions{}

	sd := InternalStateDefinition{
		State:     state,
		Triggers:  make(map[plinko.Trigger]*TriggerDefinition),
		Abs:       &pd.Abs,
		Callbacks: &cbd,
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
	Predicate        func(plinko.Payload, plinko.TransitionInfo) bool
}

type PlinkoDataStructure struct {
	States map[plinko.State]plinko.StateDefinition
}

func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func addPermit(sd *InternalStateDefinition, trigger plinko.Trigger, destination plinko.State, predicate func(plinko.Payload, plinko.TransitionInfo) bool) {
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
