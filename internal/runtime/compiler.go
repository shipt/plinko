/**
 * Copyright (c) Shipt.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */
package runtime

import (
	"bytes"
	"fmt"

	"github.com/shipt/plinko"
	"github.com/shipt/plinko/internal/renderers"
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

	b := bytes.NewBuffer([]byte{})
	r := renderers.NewUML(b)
	err := pd.Render(r)

	return plinko.Uml(b.String()), err
}

func (pd PlinkoDefinition) Render(renderer plinko.Renderer) error {
	return renderer.Render(pd)
}

// Edges implements Edges method of the plinko.Graph interface
func (pd PlinkoDefinition) Edges(edgeFunc func(state, destinationState plinko.State, name plinko.Trigger)) {
	for _, sd := range pd.Abs.StateDefinitions {
		for _, td := range sd.Triggers {
			edgeFunc(sd.State, td.DestinationState, td.Name)
		}
	}
}

// Nodes implements Nodes method of the plinko.Graph interface
func (pd PlinkoDefinition) Nodes(nodeFunc func(state plinko.State, StateConfig plinko.StateConfig)) {
	for _, sd := range pd.Abs.StateDefinitions {
		nodeFunc(sd.State, sd.info)
	}
}
