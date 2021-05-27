package renderers_test

import (
	"bytes"
	"testing"

	"github.com/shipt/plinko"
	"github.com/shipt/plinko/internal/renderers"
	"github.com/shipt/plinko/pkg/config"
	"github.com/shipt/plinko/pkg/config/state"
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
const NewOrder plinko.State = "NewOrder"

func Test_CreateDot(t *testing.T) {
	p := config.CreatePlinkoDefinition()

	p.Configure(NewOrder, state.WithName("Very much new order"), state.WithDescription("Where it all begins")).
		Permit("Submit", "PublishedOrder").
		Permit("Review", "UnderReview")

	p.Configure("PublishedOrder")

	p.Configure("UnderReview").
		Permit("CompleteReview", "PublishedOrder").
		Permit("RejectOrder", "RejectedOrder")

	p.Configure("RejectedOrder")

	buf := bytes.NewBufferString("")

	err := p.Render(renderers.NewDot(buf))
	assert.Nil(t, err)
	assert.Contains(t, buf.String(), `"UnderReview" -> "PublishedOrder"[label="CompleteReview"];`)
	assert.Contains(t, buf.String(), `Very much new order`)
	assert.Contains(t, buf.String(), `Where it all begins`)
}
