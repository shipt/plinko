package runtime

import (
	"fmt"
	"reflect"
	"runtime"

	"github.com/shipt/plinko/internal/sideeffects"
	"github.com/shipt/plinko/pkg/plinko"
)

type CallbackDefinitions struct {
	OnEntryFn []chainedFunctionCall
	OnExitFn  func(pp plinko.Payload, transitionInfo plinko.TransitionInfo) (plinko.Payload, error)

	EntryFunctionChain []string
	ExitFunctionChain  []string
}
type plinkoStateMachine struct {
	pd PlinkoDefinition
}

type chainedFunctionCall struct {
	Predicate func(pp plinko.Payload, transitionInfo plinko.TransitionInfo) bool
	Operation func(pp plinko.Payload, transitionInfo plinko.TransitionInfo) (plinko.Payload, error)
}

type StateDefinition struct {
	State    plinko.State
	Triggers map[plinko.Trigger]*TriggerDefinition

	Callbacks *CallbackDefinitions

	Abs *AbstractSyntax
}

func (sd StateDefinition) OnEntry(entryFn func(pp plinko.Payload, transitionInfo plinko.TransitionInfo) (plinko.Payload, error)) plinko.StateDefinition {
	sd.Callbacks.OnEntryFn = append(sd.Callbacks.OnEntryFn, chainedFunctionCall{
		Predicate: nil,
		Operation: entryFn,
	})
	sd.Callbacks.EntryFunctionChain = append(sd.Callbacks.EntryFunctionChain, getFunctionName(entryFn))

	return sd
}

func (sd StateDefinition) OnTriggerEntry(trigger plinko.Trigger, entryFn func(pp plinko.Payload, transitionInfo plinko.TransitionInfo) (plinko.Payload, error)) plinko.StateDefinition {
	sd.Callbacks.OnEntryFn = append(sd.Callbacks.OnEntryFn, chainedFunctionCall{
		Predicate: func(_ plinko.Payload, t plinko.TransitionInfo) bool {
			return t.GetTrigger() == trigger
		},
		Operation: entryFn,
	})

	return sd
}

func (sd StateDefinition) OnExit(exitFn func(pp plinko.Payload, transitionInfo plinko.TransitionInfo) (plinko.Payload, error)) plinko.StateDefinition {
	sd.Callbacks.OnExitFn = exitFn
	sd.Callbacks.ExitFunctionChain = append(sd.Callbacks.ExitFunctionChain, getFunctionName(exitFn))

	return sd
}

func (sd StateDefinition) Permit(trigger plinko.Trigger, destinationState plinko.State) plinko.StateDefinition {
	addPermit(&sd, trigger, destinationState, nil)

	return sd
}

func (sd StateDefinition) PermitIf(predicate func(plinko.Payload, plinko.TransitionInfo) bool, trigger plinko.Trigger, destinationState plinko.State) plinko.StateDefinition {
	addPermit(&sd, trigger, destinationState, predicate)

	return sd
}

func (psm plinkoStateMachine) EnumerateActiveTriggers(payload plinko.Payload) ([]plinko.Trigger, error) {
	state := payload.GetState()
	sd2 := (*psm.pd.States)[state]

	if sd2 == nil {
		return nil, fmt.Errorf("State %s not found in state machine definition", state)
	}

	keys := make([]plinko.Trigger, 0, len(sd2.Triggers))
	for k := range sd2.Triggers {
		keys = append(keys, k)
	}

	return keys, nil

}

func (psm plinkoStateMachine) CanFire(payload plinko.Payload, trigger plinko.Trigger) bool {
	state := payload.GetState()
	sd2 := (*psm.pd.States)[state]

	if sd2 == nil {
		return false
	}

	triggerData := sd2.Triggers[trigger]
	if triggerData == nil {
		return false
	}

	if triggerData.Predicate != nil {
		return triggerData.Predicate(payload, sideeffects.TransitionDef{
			Destination: triggerData.DestinationState,
			Source:      state,
			Trigger:     triggerData.Name,
		})
	}

	return true
}

func (psm plinkoStateMachine) Fire(payload plinko.Payload, trigger plinko.Trigger) (plinko.Payload, error) {
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

	td := sideeffects.TransitionDef{
		Source:      state,
		Destination: destinationState.State,
		Trigger:     trigger,
	}

	sideeffects.Dispatch(plinko.BeforeTransition, psm.pd.SideEffects, payload, td)

	if sd2.Callbacks.OnExitFn != nil {
		sd2.Callbacks.OnExitFn(payload, td)
	}

	sideeffects.Dispatch(plinko.BetweenStates, psm.pd.SideEffects, payload, td)

	if destinationState.Callbacks.OnEntryFn != nil && len(destinationState.Callbacks.OnEntryFn) > 0 {
		for _, fn := range destinationState.Callbacks.OnEntryFn {
			if fn.Predicate != nil {
				if !fn.Predicate(payload, td) {
					continue
				}
			}

			payload, e := fn.Operation(payload, td)

			if e != nil {
				return payload, e
			}
		}
	}

	sideeffects.Dispatch(plinko.AfterTransition, psm.pd.SideEffects, payload, td)

	return payload, nil
}

type AbstractSyntax struct {
	States             []plinko.State
	TriggerDefinitions []TriggerDefinition
	StateDefinitions   []*StateDefinition
}

type PlinkoDefinition struct {
	States      *map[plinko.State]*StateDefinition
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

func (pd PlinkoDefinition) RenderUml() (plinko.Uml, error) {
	cm := pd.Compile()

	for _, def := range cm.Messages {
		if def.CompileMessage == plinko.CompileError {
			return "", fmt.Errorf("critical errors exist in definition")
		}
	}

	var uml plinko.Uml
	uml = "@startuml\n"
	uml += plinko.Uml(fmt.Sprintf("[*] -> %s \n", pd.Abs.StateDefinitions[0].State))

	for _, sd := range pd.Abs.StateDefinitions {

		for _, td := range sd.Triggers {
			uml += plinko.Uml(fmt.Sprintf("%s --> %s : %s\n", sd.State, td.DestinationState, td.Name))
		}
	}

	uml += "@enduml"
	return uml, nil
}

func (pd *PlinkoDefinition) SideEffect(sideEffect plinko.SideEffect) plinko.PlinkoDefinition {
	pd.SideEffects = append(pd.SideEffects, sideeffects.SideEffectDefinition{Filter: sideeffects.AllowAllSideEffects, SideEffect: sideEffect})

	return pd
}

func (pd *PlinkoDefinition) FilteredSideEffect(filter plinko.SideEffectFilter, sideEffect plinko.SideEffect) plinko.PlinkoDefinition {
	pd.SideEffects = append(pd.SideEffects, sideeffects.SideEffectDefinition{Filter: filter, SideEffect: sideEffect})

	return pd
}

func (pd PlinkoDefinition) Compile() plinko.CompilerOutput {

	var compilerMessages []plinko.CompilerMessage

	for _, def := range pd.Abs.TriggerDefinitions {
		if !findDestinationState(pd.Abs.States, def.DestinationState) {
			compilerMessages = append(compilerMessages, plinko.CompilerMessage{
				CompileMessage: plinko.CompileError,
				Message:        fmt.Sprintf("State '%s' undefined: Trigger '%s' declares a transition to this undefined state.", def.DestinationState, def.Name),
			})
		}
	}

	for _, def := range pd.Abs.StateDefinitions {
		if len(def.Triggers) == 0 {
			compilerMessages = append(compilerMessages, plinko.CompilerMessage{
				CompileMessage: plinko.CompileWarning,
				Message:        fmt.Sprintf("State '%s' is a state without any triggers (deadend state).", def.State),
			})
		}
	}

	psm := plinkoStateMachine{
		pd: pd,
	}

	co := plinko.CompilerOutput{
		Messages:     compilerMessages,
		StateMachine: psm,
	}

	return co
}

func (pd *PlinkoDefinition) Configure(state plinko.State) plinko.StateDefinition {
	if _, ok := (*pd.States)[state]; ok {
		panic(fmt.Sprintf("State: %s - has already been defined, plinko configuration invalid.", state))
	}

	cbd := CallbackDefinitions{}

	sd := StateDefinition{
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

func addPermit(sd *StateDefinition, trigger plinko.Trigger, destination plinko.State, predicate func(plinko.Payload, plinko.TransitionInfo) bool) {
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
