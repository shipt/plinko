package composition

import (
	"testing"

	"github.com/bmizerany/assert"
	"github.com/shipt/plinko/pkg/plinko"
)

func TestAddEntry(t *testing.T) {
	cd := CallbackDefinitions{}

	cd.AddEntry(nil, func(pp plinko.Payload, transitionInfo plinko.TransitionInfo) (plinko.Payload, error) {
		return pp, nil
	})

	assert.Equal(t, 1, len(cd.OnEntryFn))
}
