package config

import "github.com/shipt/plinko/pkg/plinko"
import "github.com/shipt/plinko/internal/runtime"

// CreateDefinition creates a new structure used in defining the state machine.
func CreateDefinition() plinko.PlinkoDefinition {
	stateMap := make(map[plinko.State]*runtime.StateDefinition)
	p := runtime.PlinkoDefinition{
		States: &stateMap,
	}

	p.Abs = runtime.AbstractSyntax{}

	return &p
}
