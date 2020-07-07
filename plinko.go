package plinko

import (
	"fmt"
)

type Uml string

type CompilerOutput struct {
	StateMachine StateMachine
	Messages     []CompilerMessage
}

type plinkoStateMachine struct {
	pd plinkoDefinition
}

type transitionDef struct {
	source      State
	destination State
	trigger     Trigger
}

func (td transitionDef) GetSource() State {
	return td.source
}

func (td transitionDef) GetDestination() State {
	return td.destination
}

func (td transitionDef) GetTrigger() Trigger {
	return td.trigger
}

func (psm plinkoStateMachine) EnumerateActiveTriggers(payload Payload) ([]Trigger, error) {
	state := payload.GetState()
	sd2 := (*psm.pd.States)[state]

	if sd2 == nil {
		return nil, fmt.Errorf("State %s not found in state machine definition", state)
	}

	keys := make([]Trigger, 0, len(sd2.Triggers))
	for k := range sd2.Triggers {
		keys = append(keys, k)
	}

	return keys, nil

}

func (psm plinkoStateMachine) CanFire(payload Payload, trigger Trigger) bool {
	state := payload.GetState()
	sd2 := (*psm.pd.States)[state]

	if sd2 == nil {
		return false
	}

	triggerData := sd2.Triggers[trigger]
	if triggerData == nil {
		return false
	}

	return true
}

func (psm plinkoStateMachine) Fire(payload Payload, trigger Trigger) (Payload, error) {
	state := payload.GetState()
	sd2 := (*psm.pd.States)[state]

	if sd2 == nil {
		return payload, fmt.Errorf("State not found in definition of states: %s", state)
	}

	triggerData := sd2.Triggers[trigger]
	if triggerData == nil {
		return payload, fmt.Errorf("Trigger '%s' not found in definition for state: %s", trigger, state)
	}

	destinationState := (*psm.pd.States)[triggerData.DestinationState]

	td := transitionDef{
		source:      state,
		destination: destinationState.State,
		trigger:     trigger,
	}

	if sd2.callbacks.OnExitFn != nil {
		sd2.callbacks.OnExitFn(payload, td)
	}

	if destinationState.callbacks.OnEntryFn != nil {
		destinationState.callbacks.OnEntryFn(payload, td)
	}

	return payload, nil
}

func CreateDefinition() PlinkoDefinition {
	stateMap := make(map[State]*stateDefinition)
	plinko := plinkoDefinition{
		States: &stateMap,
	}

	plinko.abs = abstractSyntax{}

	return &plinko
}

type abstractSyntax struct {
	States             []State
	TriggerDefinitions []TriggerDefinition
	StateDefinitions   []*stateDefinition
}

type plinkoDefinition struct {
	States *map[State]*stateDefinition
	abs    abstractSyntax
}

func findDestinationState(states []State, searchState State) bool {
	for _, searchVal := range states {
		if searchVal == searchState {
			return true
		}
	}

	return false
}

func (pd plinkoDefinition) RenderUml() (Uml, error) {
	cm := pd.Compile()

	for _, def := range cm.Messages {
		if def.CompileMessage == CompileError {
			return "", fmt.Errorf("critical errors exist in definition")
		}
	}

	var uml Uml
	uml = "@startuml\n"
	uml += Uml(fmt.Sprintf("[*] -> %s \n", pd.abs.StateDefinitions[0].State))

	for _, sd := range pd.abs.StateDefinitions {

		for _, td := range sd.Triggers {
			uml += Uml(fmt.Sprintf("%s --> %s : %s\n", sd.State, td.DestinationState, td.Name))
		}
	}

	uml += "@enduml"
	return uml, nil
}

func (pd plinkoDefinition) Compile() CompilerOutput {

	var compilerMessages []CompilerMessage

	for _, def := range pd.abs.TriggerDefinitions {
		if !findDestinationState(pd.abs.States, def.DestinationState) {
			compilerMessages = append(compilerMessages, CompilerMessage{
				CompileMessage: CompileError,
				Message:        fmt.Sprintf("State '%s' undefined: Trigger '%s' declares a transition to this undefined state.", def.DestinationState, def.Name),
			})
		}
	}

	for _, def := range pd.abs.StateDefinitions {
		if len(def.Triggers) == 0 {
			compilerMessages = append(compilerMessages, CompilerMessage{
				CompileMessage: CompileWarning,
				Message:        fmt.Sprintf("State '%s' is a state without any triggers (deadend state).", def.State),
			})
		}
	}

	psm := plinkoStateMachine{
		pd: pd,
	}

	co := CompilerOutput{
		Messages:     compilerMessages,
		StateMachine: psm,
	}

	return co
}

func (pd *plinkoDefinition) Configure(state State) StateDefinition {
	if _, ok := (*pd.States)[state]; ok {
		panic(fmt.Sprintf("State: %s - has already been defined, plinko configuration invalid.", state))
	}

	cbd := CallbackDefinitions{}
	sd := stateDefinition{
		State:     state,
		Triggers:  make(map[Trigger]*TriggerDefinition),
		abs:       &pd.abs,
		callbacks: &cbd,
	}

	(*pd.States)[state] = &sd

	pd.abs.States = append(pd.abs.States, state)
	pd.abs.StateDefinitions = append(pd.abs.StateDefinitions, &sd)

	return sd
}

type compileInfo struct {
}

type TriggerDefinition struct {
	Name             Trigger
	DestinationState State
}

type stateDefinition struct {
	State    State
	Triggers map[Trigger]*TriggerDefinition

	callbacks *CallbackDefinitions

	abs *abstractSyntax
}

type PlinkoDataStructure struct {
	States map[State]StateDefinition
}

func (sd stateDefinition) OnEntry(entryFn func(pp Payload, transitionInfo TransitionInfo) (Payload, error)) StateDefinition {
	sd.callbacks.OnEntryFn = entryFn

	return sd
}

func (sd stateDefinition) OnExit(exitFn func(pp Payload, transitionInfo TransitionInfo) (Payload, error)) StateDefinition {
	sd.callbacks.OnExitFn = exitFn

	return sd
}

func (sd stateDefinition) Permit(triggerName Trigger, destinationState State) StateDefinition {
	if _, ok := sd.Triggers[triggerName]; ok {
		panic(fmt.Sprintf("Trigger: %s - has already been defined, plinko configuration invalid.", triggerName))
	}
	td := TriggerDefinition{
		Name:             triggerName,
		DestinationState: destinationState,
	}
	sd.Triggers[triggerName] = &td

	sd.abs.TriggerDefinitions = append(sd.abs.TriggerDefinitions, td)

	return sd
}
