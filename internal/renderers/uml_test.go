// Copyright 2021, Shipt. All rights reserved.
// Licensed under the Apache License
package renderers_test

import (
	"bytes"
	"testing"

	"github.com/shipt/plinko/internal/renderers"
	"github.com/shipt/plinko/pkg/config"
	"github.com/stretchr/testify/assert"
)

func Test_CreateUML(t *testing.T) {
	p := config.CreatePlinkoDefinition()

	p.Configure(NewOrder).
		Permit("Submit", "PublishedOrder").
		Permit("Review", "UnderReview")

	p.Configure("PublishedOrder")

	p.Configure("UnderReview").
		Permit("CompleteReview", "PublishedOrder").
		Permit("RejectOrder", "RejectedOrder")

	p.Configure("RejectedOrder")

	buf := bytes.NewBufferString("")

	err := p.Render(renderers.NewUML(buf))
	assert.Nil(t, err)
	assert.Contains(t, buf.String(), "UnderReview --> PublishedOrder : CompleteReview")
}
