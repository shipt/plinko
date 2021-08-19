/**
 * Copyright (c) Shipt.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */
package operation

import (
	"github.com/shipt/plinko"
)

func WithName(name string) func(*plinko.OperationConfig) {
	return func(c *plinko.OperationConfig) {
		c.Name = name
	}
}
