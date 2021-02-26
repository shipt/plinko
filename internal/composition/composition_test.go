package composition

import (
	"errors"
	"testing"

	"github.com/shipt/plinko"
	"github.com/shipt/plinko/internal/sideeffects"
	"github.com/shipt/plinko/plinkoerror"
	"github.com/stretchr/testify/assert"
)

func TestAddEntry(t *testing.T) {
	cd := CallbackDefinitions{}

	cd.AddEntry(nil, func(pp plinko.Payload, transitionInfo plinko.TransitionInfo) (plinko.Payload, error) {
		return pp, nil
	})

	assert.Equal(t, 1, len(cd.OnEntryFn))
}

func TestAddExit(t *testing.T) {
	cd := CallbackDefinitions{}

	cd.AddExit(nil, func(pp plinko.Payload, transitionInfo plinko.TransitionInfo) (plinko.Payload, error) {
		return pp, nil
	})

	assert.Equal(t, 1, len(cd.OnExitFn))
}

func TestExecuteErrorChainSingleFunctionWithModifiedDestination(t *testing.T) {
	const Woo plinko.State = "woo"
	const ErrorState plinko.State = "bar2"
	const GoodState plinko.State = "bar"
	transitionDef := sideeffects.TransitionDef{
		Source:      "foo",
		Destination: GoodState,
		Trigger:     "baz",
	}
	list := []ChainedErrorCall{
		ChainedErrorCall{
			ErrorOperation: func(p plinko.Payload, m plinko.ModifiableTransitionInfo, e error) (plinko.Payload, error) {
				m.SetDestination(ErrorState)
				return p, nil
			},
		},
	}

	p, t1, e := executeErrorChain(list, nil, &transitionDef, errors.New("wizard"))

	assert.Equal(t, ErrorState, t1.GetDestination())
	assert.Equal(t, errors.New("wizard"), e)
	assert.Equal(t, p, nil)

}

func TestExecuteErrorChainMultiFunctionWithError(t *testing.T) {
	const Woo plinko.State = "woo"
	const ErrorState plinko.State = "bar2"
	const GoodState plinko.State = "bar"
	transitionDef := sideeffects.TransitionDef{
		Source:      "foo",
		Destination: GoodState,
		Trigger:     "baz",
	}
	counter := 0
	list := []ChainedErrorCall{
		ChainedErrorCall{
			ErrorOperation: func(p plinko.Payload, m plinko.ModifiableTransitionInfo, e error) (plinko.Payload, error) {
				counter++
				return p, errors.New("notwizard")
			},
		},
		ChainedErrorCall{
			ErrorOperation: func(p plinko.Payload, m plinko.ModifiableTransitionInfo, e error) (plinko.Payload, error) {
				m.SetDestination(ErrorState)
				counter++
				return p, nil
			},
		},
	}

	p, t1, e := executeErrorChain(list, nil, &transitionDef, errors.New("wizard"))

	assert.Equal(t, GoodState, t1.GetDestination())
	assert.Equal(t, 1, counter)
	assert.Equal(t, errors.New("notwizard"), e)
	assert.Equal(t, p, nil)
}

func TestChainedFunctionChainWithPanic(t *testing.T) {
	transitionDef := sideeffects.TransitionDef{
		Source:      "foo",
		Destination: "GoodState",
		Trigger:     "baz",
	}

	list := []ChainedFunctionCall{

		ChainedFunctionCall{
			Operation: func(p plinko.Payload, m plinko.TransitionInfo) (plinko.Payload, error) {
				return p, nil
			},
		},
		ChainedFunctionCall{
			Operation: func(p plinko.Payload, m plinko.TransitionInfo) (plinko.Payload, error) {
				panic(errors.New("OOGABOOGA"))
				//return p, errors.New("notwizard")
			},
		},
	}

	p, err := executeChain(list, nil, transitionDef)

	assert.Nil(t, p)
	assert.NotNil(t, err)

	e := err.(*plinkoerror.PlinkoPanicError)
	assert.NotNil(t, e)

	assert.Equal(t, "OOGABOOGA", e.InnerError.Error())
	assert.Nil(t, e.UnknownInnerError)
	assert.Equal(t, 1, e.StepNumber)

}

/*
func TestErrorFunctionChainWithPanic(t *testing.T) {
	transitionDef := sideeffects.TransitionDef{
		Source:      "foo",
		Destination: "GoodState",
		Trigger:     "baz",
	}

	list := []ChainedErrorCall{

		ChainedFunctionCall{
			Operation: func(p plinko.Payload, m plinko.TransitionInfo) (plinko.Payload, error) {
				return p, nil
			},
		},
		ChainedFunctionCall{
			Operation: func(p plinko.Payload, m plinko.TransitionInfo) (plinko.Payload, error) {
				panic(errors.New("OOGABOOGA"))
				//return p, errors.New("notwizard")
			},
		},
	}

	p, err := executeChain(list, nil, transitionDef)

	assert.Nil(t, p)
	assert.NotNil(t, err)

	e := err.(*plinkoerror.PlinkoPanicError)
	assert.NotNil(t, e)

	assert.Equal(t, "OOGABOOGA", e.InnerError.Error())
	assert.Nil(t, e.UnknownInnerError)
	assert.Equal(t, 1, e.StepNumber)

}
*/
