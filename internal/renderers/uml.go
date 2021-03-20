package renderers

import (
	"fmt"
	"io"

	"github.com/shipt/plinko"
)

type UML struct {
	*writeWrapper
}

func NewUML(w io.Writer) *UML {
	return &UML{
		writeWrapper: &writeWrapper{writer: w},
	}
}

func (d *UML) Render(graph plinko.Graph) error {
	d.write([]byte("@startuml\n"))

	firstEdge := true
	graph.Edges(func(state, destinationState plinko.State, name plinko.Trigger) {
		if firstEdge {
			d.write([]byte(fmt.Sprintf("[*] -> %s \n", state)))
			firstEdge = false
		}
		d.write([]byte(fmt.Sprintf("%s --> %s : %s\n", state, destinationState, name)))
	})

	d.write([]byte("@enduml"))
	return nil
}
