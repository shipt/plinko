package interfaces

import "github.com/shipt/plinko/types"

type PlinkoPayload interface {
	GetState() types.State
}
