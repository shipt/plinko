package dot_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/shipt/plinko"
	"github.com/shipt/plinko/internal/renderers/dot"
	"github.com/shipt/plinko/internal/runtime"
	"github.com/shipt/plinko/pkg/config"
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
	p := config.CreatePlinkoDefinition().(*runtime.PlinkoDefinition)

	p.Configure(NewOrder).
		Permit("Submit", "PublishedOrder").
		Permit("Review", "UnderReview")

	p.Configure("PublishedOrder")

	p.Configure("UnderReview").
		Permit("CompleteReview", "PublishedOrder").
		Permit("RejectOrder", "RejectedOrder")

	p.Configure("RejectedOrder")

	buf := bytes.NewBufferString("")

	dot.New(p).Write(buf)

	fmt.Println(buf.String())
}
