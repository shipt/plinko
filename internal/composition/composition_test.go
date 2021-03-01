package composition

import (
	"context"
	"errors"
	"testing"

	"github.com/shipt/plinko"
	"github.com/shipt/plinko/internal/sideeffects"
	"github.com/shipt/plinko/plinkoerror"
	"github.com/stretchr/testify/assert"
)

type testPayload struct {
	value string
}

func (t *testPayload) GetState() plinko.State {
	return "stub"
}

func TestAddEntry(t *testing.T) {
	cd := CallbackDefinitions{}

	cd.AddEntry(nil, func(_ context.Context, pp plinko.Payload, transitionInfo plinko.TransitionInfo) (plinko.Payload, error) {
		return pp, nil
	})

	assert.Equal(t, 1, len(cd.OnEntryFn))
}

func TestAddExit(t *testing.T) {
	cd := CallbackDefinitions{}

	cd.AddExit(nil, func(_ context.Context, pp plinko.Payload, transitionInfo plinko.TransitionInfo) (plinko.Payload, error) {
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
			ErrorOperation: func(_ context.Context, p plinko.Payload, m plinko.ModifiableTransitionInfo, e error) (plinko.Payload, error) {
				m.SetDestination(ErrorState)
				return p, nil
			},
		},
	}

	p, t1, e := executeErrorChain(context.TODO(), list, nil, &transitionDef, errors.New("wizard"))

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
			ErrorOperation: func(_ context.Context, p plinko.Payload, m plinko.ModifiableTransitionInfo, e error) (plinko.Payload, error) {
				counter++
				return p, errors.New("notwizard")
			},
		},
		ChainedErrorCall{
			ErrorOperation: func(_ context.Context, p plinko.Payload, m plinko.ModifiableTransitionInfo, e error) (plinko.Payload, error) {
				m.SetDestination(ErrorState)
				counter++
				return p, nil
			},
		},
	}

	p, t1, e := executeErrorChain(context.TODO(), list, nil, &transitionDef, errors.New("wizard"))

	assert.Equal(t, GoodState, t1.GetDestination())
	assert.Equal(t, 1, counter)
	assert.Equal(t, errors.New("notwizard"), e)
	assert.Equal(t, p, nil)
}

func TestChainedFunctionPassingProperly(t *testing.T) {
	payload := &testPayload{}
	transitionDef := sideeffects.TransitionDef{
		Source:      "foo",
		Destination: "GoodState",
		Trigger:     "baz",
	}

	list := []ChainedFunctionCall{

		ChainedFunctionCall{
			Operation: func(_ context.Context, p plinko.Payload, m plinko.TransitionInfo) (plinko.Payload, error) {
				t := p.(*testPayload)
				t.value = "foo"

				return p, nil
			},
		},
		ChainedFunctionCall{
			Operation: func(_ context.Context, p plinko.Payload, m plinko.TransitionInfo) (plinko.Payload, error) {
				te := p.(*testPayload)
				assert.Equal(t, "foo", te.value)

				return p, nil
			},
		},
	}

	p, err := executeChain(context.TODO(), list, payload, transitionDef)

	assert.NotNil(t, p)
	assert.Nil(t, err)
}

func TestChainedFunctionChainWithPanic(t *testing.T) {
	transitionDef := sideeffects.TransitionDef{
		Source:      "foo",
		Destination: "GoodState",
		Trigger:     "baz",
	}

	list := []ChainedFunctionCall{

		ChainedFunctionCall{
			Operation: func(_ context.Context, p plinko.Payload, m plinko.TransitionInfo) (plinko.Payload, error) {
				return p, nil
			},
		},
		ChainedFunctionCall{
			Operation: func(_ context.Context, p plinko.Payload, m plinko.TransitionInfo) (plinko.Payload, error) {
				panic(errors.New("panic-error"))
				//return p, errors.New("notwizard")
			},
		},
	}

	p, err := executeChain(context.TODO(), list, nil, transitionDef)

	assert.Nil(t, p)
	assert.NotNil(t, err)

	e := err.(*plinkoerror.PlinkoPanicError)
	assert.NotNil(t, e)

	assert.Equal(t, "panic-error", e.InnerError.Error())
	assert.Nil(t, e.UnknownInnerError)
	assert.Equal(t, 1, e.StepNumber)

}

func TestErrorFunctionChainWithPanic(t *testing.T) {
	transitionDef := sideeffects.TransitionDef{
		Source:      "foo",
		Destination: "GoodState",
		Trigger:     "baz",
	}

	list := []ChainedErrorCall{
		ChainedErrorCall{
			ErrorOperation: func(_ context.Context, p plinko.Payload, m plinko.ModifiableTransitionInfo, e error) (plinko.Payload, error) {

				panic(errors.New("panic-error"))
			},
		},
		ChainedErrorCall{
			ErrorOperation: func(_ context.Context, p plinko.Payload, m plinko.ModifiableTransitionInfo, e error) (plinko.Payload, error) {

				return p, nil
			},
		},
	}

	p, td2, err := executeErrorChain(context.TODO(), list, nil, &transitionDef, errors.New("encompassing-error"))

	assert.Nil(t, p)
	assert.NotNil(t, err)
	assert.Equal(t, plinko.State("GoodState"), td2.Destination)

	e := err.(*plinkoerror.PlinkoPanicError)
	assert.NotNil(t, e)

	assert.Equal(t, "panic-error", e.InnerError.Error())
	assert.Nil(t, e.UnknownInnerError)
	assert.Equal(t, 0, e.StepNumber)
}
