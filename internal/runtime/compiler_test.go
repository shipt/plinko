/**
 * Copyright (c) Shipt.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */
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
