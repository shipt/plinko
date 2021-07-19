// Copyright 2021, Shipt. All rights reserved.
// Licensed under the Apache License

package renderers

import (
	"fmt"
	"io"
	"os/exec"

	"github.com/shipt/plinko"
)

type Dot struct {
	*writeWrapper
	style dotStylesheet
}

func NewDot(w io.Writer) *Dot {
	return &Dot{
		writeWrapper: &writeWrapper{writer: w},
		style:        defaultDotStyle,
	}
}

func (d *Dot) Render(graph plinko.Graph) error {
	d.beginGraph()
	graph.Nodes(func(state plinko.State, info plinko.StateConfig) {
		d.node(string(state), info.Name, info.Description)
	})
	graph.Edges(func(state, destinationState plinko.State, name plinko.Trigger) {
		d.edge(string(state), string(destinationState), string(name))
	})
	d.endGraph()
	return d.err
}

func (d *Dot) beginGraph() {
	d.write([]byte("digraph {\n"))
	d.write([]byte(d.style.graphHeader))
	d.write([]byte(d.style.defaults.graph))
	d.write([]byte(d.style.defaults.node))
	d.write([]byte(d.style.defaults.edge))
}

func (d *Dot) endGraph() {
	d.write([]byte("}\n"))
}

func (d *Dot) edge(a, b, label string) {
	d.write([]byte(fmt.Sprintf(d.style.templates.edge, a, b, label)))
}

func (d *Dot) node(name, label, description string) {
	d.write([]byte(fmt.Sprintf(d.style.templates.node, name, label, description)))
}

//DotFileToImg runs the dot command to convert a dot file into an image file
func DotFileToImg(from, to, format string) error {
	_, err := exec.Command("sh", "-c", "dot -T"+format+" "+from+" -o "+to).Output()
	return err
}

type dotStylesheet struct {
	graphHeader string
	defaults    dotDefaultStyles
	templates   dotTemplates
}

type dotDefaultStyles struct {
	graph string
	node  string
	edge  string
}

type dotTemplates struct {
	node string
	edge string
}

var defaultDotStyle = dotStylesheet{
	graphHeader: `layout=fdp;
overlap=false;
sep=1.5;
maxiter=2000;
start=1251;
`,
	defaults: dotDefaultStyles{
		graph: "graph [splines=\"spline\", ranksep=\"2\", nodesep=\"1\"];\n",
		node:  "node [shape=plaintext];\n",
		edge:  "edge [constraint=true, fontname = \"sans-serif\"];\n",
	},
	templates: dotTemplates{
		node: `"%s" [label=<<TABLE STYLE="ROUNDED" BGCOLOR="orange" BORDER="1" CELLSPACING="0" WIDTH="20"><TR><TD BORDER="0">%s</TD></TR><TR><TD BORDER="1" SIDES="t">%s</TD></TR></TABLE>>];` + "\n",
		edge: "\"%s\" -> \"%s\"[label=\"%s\"];\n",
	},
}
