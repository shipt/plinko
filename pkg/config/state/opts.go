// Copyright 2021, Shipt. All rights reserved.
// Licensed under the Apache License
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
