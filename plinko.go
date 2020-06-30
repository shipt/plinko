package plinko

import (
	"fmt"
	"github.com/shipt/plinko/types"

	definitions "github.com/shipt/plinko/definitions"
	"github.com/shipt/plinko/interfaces"
)

type Uml string

type PlinkoCompilerOutput struct {
	PlinkoStateMachine PlinkoStateMachine
	Messages           []CompilerMessage
}

type PlinkoStateMachine interface {
	Fire(payload interfaces.PlinkoPayload, trigger types.Trigger) (interfaces.PlinkoPayload, error)
	//GetValidTriggers(payload interfaces.PlinkoPayload) ([]types.Trigger, error)
}

type plinkoStateMachine struct {
	pd plinkoDefinition
}

type TransitionInfo interface {
	GetSource() types.State
	GetDestination() types.State
	GetTrigger() types.Trigger
}

type transitionDef struct {
	source      types.State
	destination types.State
	trigger     types.Trigger
}

func (td transitionDef) GetSource() types.State {
	return td.source
}

func (td transitionDef) GetDestination() types.State {
	return td.destination
}

func (td transitionDef) GetTrigger() types.Trigger {
	return td.trigger
}

func (psm plinkoStateMachine) Fire(payload interfaces.PlinkoPayload, trigger types.Trigger) (interfaces.PlinkoPayload, error) {
	state := payload.GetState()
	sd2 := (*psm.pd.States)[state]

	if sd2 == nil {
		return payload, fmt.Errorf("State not found in definition of states: %s", state)
	}

	triggerData := sd2.Triggers[trigger]
	if sd2 == nil {
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
	stateMap := make(map[types.State]*stateDefinition)
	plinko := plinkoDefinition{
		States: &stateMap,
	}

	plinko.abs = abstractSyntax{}

	return &plinko
}

type abstractSyntax struct {
	States             []types.State
	TriggerDefinitions []TriggerDefinition
	StateDefinitions   []*stateDefinition
}

type PlinkoDefinition interface {
	CreateState(state types.State) StateDefinition
	Compile() PlinkoCompilerOutput
	RenderUml() (Uml, error)
}

type plinkoDefinition struct {
	States *map[types.State]*stateDefinition
	abs    abstractSyntax
}

func findDestinationState(states []types.State, searchState types.State) bool {
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

func (pd plinkoDefinition) Compile() PlinkoCompilerOutput {

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

	co := PlinkoCompilerOutput{
		Messages:           compilerMessages,
		PlinkoStateMachine: psm,
	}

	return co
}

func (pd *plinkoDefinition) CreateState(state types.State) StateDefinition {
	if _, ok := (*pd.States)[state]; ok {
		panic(fmt.Sprintf("State: %s - has already been defined, plinko configuration invalid.", state))
	}

	cbd := definitions.CallbackDefinitions{}
	sd := stateDefinition{
		State:     state,
		Triggers:  make(map[types.Trigger]*TriggerDefinition),
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

type StateDefinition interface {
	//State() string
	OnEntry(entryFn func(pp interfaces.PlinkoPayload, transitionInfo definitions.TransitionInfo) (interfaces.PlinkoPayload, error)) StateDefinition
	OnExit(exitFn func(pp interfaces.PlinkoPayload, transitionInfo definitions.TransitionInfo) (interfaces.PlinkoPayload, error)) StateDefinition
	Permit(triggerName types.Trigger, destinationState types.State) StateDefinition
	//TBD: AllowReentrance
}

type TriggerDefinition struct {
	Name             types.Trigger
	DestinationState types.State
}

type stateDefinition struct {
	State    types.State
	Triggers map[types.Trigger]*TriggerDefinition

	callbacks *definitions.CallbackDefinitions

	abs *abstractSyntax
}

type PlinkoDataStructure struct {
	States map[types.State]StateDefinition
}

func (sd stateDefinition) OnEntry(entryFn func(pp interfaces.PlinkoPayload, transitionInfo definitions.TransitionInfo) (interfaces.PlinkoPayload, error)) StateDefinition {
	sd.callbacks.OnEntryFn = entryFn

	return sd
}

func (sd stateDefinition) OnExit(exitFn func(pp interfaces.PlinkoPayload, transitionInfo definitions.TransitionInfo) (interfaces.PlinkoPayload, error)) StateDefinition {
	sd.callbacks.OnExitFn = exitFn

	return sd
}

func (sd stateDefinition) Permit(triggerName types.Trigger, destinationState types.State) StateDefinition {
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

type CompilerMessage struct {
	CompileMessage CompilerReportType
	Message        string
}

type CompilerReportType string

const (
	CompileError   CompilerReportType = "Compile Error"
	CompileWarning CompilerReportType = "Compile Warning"
	// CompileInfo CompilerReportType "Compile Info"
)
