// Copyright 2021, Shipt. All rights reserved.
// Licensed under the Apache License
package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderUML(t *testing.T) {
	p := createPlinkoDefinition()

	p.Configure(Created).
		PermitIf(PermitIfPredicate, Open, Opened)

	p.Configure(Opened).
		PermitIf(PermitIfPredicate, Open, Opened)

	o, e := p.RenderUml()
	assert.Nil(t, e)
	assert.NotNil(t, o)
}

func TestRenderUmlWithBadDefinition(t *testing.T) {
	p := createPlinkoDefinition()

	p.Configure(Created).
		PermitIf(PermitIfPredicate, Open, Opened)

	o, e := p.RenderUml()
	assert.NotNil(t, e)
	assert.NotNil(t, o)
}
