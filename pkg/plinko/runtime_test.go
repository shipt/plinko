package plinko

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEntryAndExitFunctions(t *testing.T) {
	p := CreateDefinition()
	ps := p.Configure(NewOrder)

	stateDef := ps.(stateDefinition)
	assert.Nil(t, stateDef.callbacks.OnExitFn)
	assert.Nil(t, stateDef.callbacks.OnEntryFn)

	ps = ps.OnEntry(entryFunctionForTest)

	ps = ps.OnExit(exitFunctionForTest)

	stateDef = ps.(stateDefinition)
	assert.NotNil(t, stateDef.callbacks.OnExitFn)
	assert.NotNil(t, stateDef.callbacks.OnEntryFn)

	assert.Equal(t, "github.com/shipt/plinko.entryFunctionForTest", stateDef.callbacks.EntryFunctionChain[0])
	assert.Equal(t, "github.com/shipt/plinko.exitFunctionForTest", stateDef.callbacks.ExitFunctionChain[0])
}
