package config

import (
	"github.com/shipt/plinko"
	"github.com/shipt/plinko/internal/runtime"
)

// CreateDefinition creates a new structure used in defining the state machine.
func CreateDefinition() plinko.PlinkoDefinition {
	stateMap := make(map[plinko.State]*runtime.InternalStateDefinition)
	p := runtime.PlinkoDefinition{
		States: &stateMap,
	}

	p.Abs = runtime.AbstractSyntax{}

	return &p
}
