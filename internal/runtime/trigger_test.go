package runtime

import (
	"context"
	"errors"
	"testing"

	"github.com/shipt/plinko"
	"github.com/stretchr/testify/assert"
)

const Created plinko.State = "Created"
const Opened plinko.State = "Opened"
const Claimed plinko.State = "Claimed"
const ArriveAtStore plinko.State = "ArrivedAtStore"
const MarkedAsPickedUp plinko.State = "MarkedAsPickedup"
const Delivered plinko.State = "Delivered"
const Canceled plinko.State = "Canceled"
const Returned plinko.State = "Returned"

const Submit plinko.Trigger = "Submit"
const Cancel plinko.Trigger = "Cancel"
const Open plinko.Trigger = "Open"
const Claim plinko.Trigger = "Claim"
const Deliver plinko.Trigger = "Deliver"
const Return plinko.Trigger = "Return"
const Reinstate plinko.Trigger = "Reinstate"

type testPayload struct {
	state     plinko.State
	condition bool
}

func (p *testPayload) GetState() plinko.State {
	return p.state
}

func createPlinkoDefinition() plinko.PlinkoDefinition {
	stateMap := make(map[plinko.State]*InternalStateDefinition)
	p := PlinkoDefinition{
		States: &stateMap,
	}

	p.Abs = AbstractSyntax{}

	return &p
}
func TestCanFireWithPermitIf(t *testing.T) {
	p := createPlinkoDefinition()

	p.Configure(Created).
		PermitIf(PermitIfPredicate, Open, Opened)

	p.Configure(Opened)

	co := p.Compile()

	psm := co.StateMachine

	payload := &testPayload{
		state:     Created,
		condition: true,
	}
	assert.Nil(t, psm.CanFire(context.TODO(), payload, Open))

	payload.condition = false
	assert.NotNil(t, psm.CanFire(context.TODO(), payload, Open))

}

func PermitIfPredicate(_ context.Context, p plinko.Payload, t plinko.TransitionInfo) error {
	tp := p.(*testPayload)

	if tp.condition {
		return nil
	}

	return errors.New("permit failed")
}
