package plinko

import "fmt"

type State string
type Trigger string
type SideEffect string

func CreateDefinition() PlinkoDefinition {
	stateMap := make(map[State]*stateDefinition)
	plinko := plinkoDefinition{
		States: &stateMap,
	}

	return plinko
}

type PlinkoPayload interface {
}

type PlinkoDefinition interface {
	CreateState(state State) StateDefinition
	//Compile()
	//RenderPlantUml()
}

type plinkoDefinition struct {
	States *map[State]*stateDefinition
}

func (pd plinkoDefinition) CreateState(state State) StateDefinition {
	if _, ok := (*pd.States)[state]; ok {
		panic(fmt.Sprintf("State: %s - has already been defined, plinko configuration invalid.", state))
	}

	sd := stateDefinition{
		State:    state,
		Triggers: make(map[Trigger]*TriggerDefinition),
	}

	(*pd.States)[state] = &sd

	return sd
}

type StateDefinition interface {
	//State() string
	OnEntry(entryFn func(pp *PlinkoPayload) (*PlinkoPayload, error)) StateDefinition
	OnExit(exitFn func(pp *PlinkoPayload) (*PlinkoPayload, error)) StateDefinition
	Permit(triggerName Trigger, destinationState State, sideEffect SideEffect) StateDefinition
}

type TriggerDefinition struct {
	Name             Trigger
	DestinationState State
	SideEffect       SideEffect
}

type stateDefinition struct {
	State    State
	Triggers map[Trigger]*TriggerDefinition

	OnEntryFn func(pp *PlinkoPayload) (*PlinkoPayload, error)
	OnExitFn  func(pp *PlinkoPayload) (*PlinkoPayload, error)
}

type PlinkDataStructure struct {
	States map[State]StateDefinition
}

func (sd stateDefinition) OnEntry(entryFn func(pp *PlinkoPayload) (*PlinkoPayload, error)) StateDefinition {
	sd.OnEntryFn = entryFn

	return sd
}

func (sd stateDefinition) OnExit(exitFn func(pp *PlinkoPayload) (*PlinkoPayload, error)) StateDefinition {
	sd.OnExitFn = exitFn

	return sd
}

func (sd stateDefinition) Permit(triggerName Trigger, destinationState State, sideEffect SideEffect) StateDefinition {
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
