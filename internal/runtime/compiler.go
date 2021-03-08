package runtime

import (
	"fmt"

	"github.com/shipt/plinko"
)

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

func (pd PlinkoDefinition) IterateEdges(edgeFunc func(state, destinationState plinko.State, name plinko.Trigger)) {
	for _, sd := range pd.Abs.StateDefinitions {
		for _, td := range sd.Triggers {
			edgeFunc(sd.State, td.DestinationState, td.Name)
		}
	}
}
