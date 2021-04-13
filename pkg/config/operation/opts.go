package operation

import (
	"github.com/shipt/plinko"
)

func WithName(name string) func(*plinko.OperationConfig) {
	return func(c *plinko.OperationConfig) {
		c.Name = name
	}
}
