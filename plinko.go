package plinko

import "fmt"

type PlinkoData struct {
	States map[string]*stateDefinition
}

func (pd *PlinkoData) CreateState(state string) *stateDefinition {
	if _, ok := pd.States[state]; ok {
		panic(fmt.Sprintf("State: %s - has already been defined, plinko configuration invalid.", state))
	}

	sd := stateDefinition{
		State:    state,
		Triggers: make(map[string]*TriggerDefinition),
	}

	pd.States[state] = &sd

	return &sd
}

type StateDefinition interface {
	State() string
}

type TriggerDefinition struct {
	Name             string
	DestinationState string
	SideEffect       string
}

type stateDefinition struct {
	State    string
	Triggers map[string]*TriggerDefinition
}

type PlinkDataStructure struct {
	States map[string]StateDefinition
}

func (sd *stateDefinition) AddTrigger(triggerName string, destinationState string, sideEffect string) *stateDefinition {
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
