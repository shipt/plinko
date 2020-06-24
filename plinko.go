package plinko

import "fmt"

type State string
type Trigger string
type SideEffect string

type PlinkoData struct {
	States map[State]*stateDefinition
}

func (pd *PlinkoData) CreateState(state State) *stateDefinition {
	if _, ok := pd.States[state]; ok {
		panic(fmt.Sprintf("State: %s - has already been defined, plinko configuration invalid.", state))
	}

	sd := stateDefinition{
		State:    state,
		Triggers: make(map[Trigger]*TriggerDefinition),
	}

	pd.States[state] = &sd

	return &sd
}

type StateDefinition interface {
	State() string
}

type TriggerDefinition struct {
	Name             Trigger
	DestinationState State
	SideEffect       SideEffect
}

type stateDefinition struct {
	State    State
	Triggers map[Trigger]*TriggerDefinition
}

type PlinkDataStructure struct {
	States map[string]StateDefinition
}

func (sd *stateDefinition) Permit(triggerName Trigger, destinationState State, sideEffect SideEffect) *stateDefinition {
	if _, ok := sd.Triggers[triggerName]; ok {
		panic(fmt.Sprintf("Trigger: %s - has already been defined, plinko configuration invalid.", triggerName))
	}
	td := TriggerDefinition{
		Name:             triggerName,
		DestinationState: destinationState,
		SideEffect:       sideEffect,
	}
	sd.Triggers[triggerName] = &td

	return sd
}
