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

	plinko.abs = abstractSyntax{}

	return &plinko
}

type abstractSyntax struct {
	States             []State
	TriggerDefinitions []TriggerDefinition
	StateDefinitions   []*stateDefinition
}

type PlinkoPayload interface {
}

type PlinkoDefinition interface {
	CreateState(state State) StateDefinition
	Compile() []CompilerMessage
	//RenderPlantUml()
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

func (pd plinkoDefinition) Compile() []CompilerMessage {

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

	return compilerMessages
}

func (pd *plinkoDefinition) CreateState(state State) StateDefinition {
	if _, ok := (*pd.States)[state]; ok {
		panic(fmt.Sprintf("State: %s - has already been defined, plinko configuration invalid.", state))
	}

	sd := stateDefinition{
		State:    state,
		Triggers: make(map[Trigger]*TriggerDefinition),
		abs:      &pd.abs,
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
	OnEntry(entryFn func(pp *PlinkoPayload) (*PlinkoPayload, error)) StateDefinition
	OnExit(exitFn func(pp *PlinkoPayload) (*PlinkoPayload, error)) StateDefinition
	Permit(triggerName Trigger, destinationState State, sideEffect SideEffect) StateDefinition
	//TBD: AllowReentrance
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

	abs *abstractSyntax
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
