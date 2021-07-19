// Copyright 2021, Shipt. All rights reserved.
// Licensed under the Apache License

package operation

import (
	"github.com/shipt/plinko"
)

func WithName(name string) func(*plinko.OperationConfig) {
	return func(c *plinko.OperationConfig) {
		c.Name = name
	}
}
