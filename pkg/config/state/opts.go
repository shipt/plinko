package state

import "github.com/shipt/plinko"

func WithName(name string) func(*plinko.StateConfig) {
	return func(c *plinko.StateConfig) {
		c.Name = name
	}
}

func WithDescription(description string) func(*plinko.StateConfig) {
	return func(c *plinko.StateConfig) {
		c.Description = description
	}
}
