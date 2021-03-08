package dot

import (
	"fmt"
	"io"

	"github.com/shipt/plinko"
	"github.com/shipt/plinko/internal/runtime"
)

const rankdir = "rankdir=LR\n"
const graphDefaults = "graph [splines=\"spline\", ranksep=\"2\", nodesep=\"1\"];\n"
const nodeDefaults = "node [shape=plaintext];\n"
const edgeDefaults = "edge [constraint=true, fontname = \"sans-serif\"];\n"
const nodeTemplate = `"%s" [label=<<TABLE STYLE="ROUNDED" BGCOLOR="orange" BORDER="1" CELLSPACING="0" WIDTH="20"><TR><TD BORDER="1" sides="b">%s</TD></TR><TR><TD BORDER="0">%s</TD></TR></TABLE>>];
`
const edgeTemplate = "\"%s\" -> \"%s\"[label=\"%s\"];\n"

type Dot struct {
	pd  *runtime.PlinkoDefinition
	err error
}

func New(pd *runtime.PlinkoDefinition) *Dot {
	return &Dot{pd: pd}
}

//Write writes the contents of a Micromap into a file.
func (d *Dot) Write(w io.Writer) error {
	cm := d.pd.Compile()

	for _, def := range cm.Messages {
		if def.CompileMessage == plinko.CompileError {
			return fmt.Errorf("critical errors exist in definition")
		}
	}

	d.beginGraph(w)
	d.pd.IterateEdges(func(state, destinationState plinko.State, name plinko.Trigger) {
		if state == "" {
			state = "start"
		}
		if destinationState == "" {
			destinationState = "end"
		}
		d.node(w, string(state), "DESCRIPTION NOT SUPPORTED")
		d.node(w, string(destinationState), "DESCRIPTION NOT SUPPORTED")
		d.edge(w, string(state), string(destinationState), string(name))
	})
	d.endGraph(w)
	return d.err
}

func (d *Dot) beginGraph(w io.Writer) {
	d.write(w, []byte("digraph {\n"))
	d.write(w, []byte(rankdir))
	d.write(w, []byte("layout=fdp;\n"))
	d.write(w, []byte("size=\"3,3\";\n"))
	d.write(w, []byte("overlap=false;\n"))
	d.write(w, []byte("splines=false;\n"))
	// d.write(w, []byte("pack=true;\n"))
	d.write(w, []byte("start=\"random\";\n"))
	d.write(w, []byte("sep=0.8;\n"))
	d.write(w, []byte("inputscale=0.4;\n"))
	// K=0.50
	// maxiter=2000
	// start=1251

	d.write(w, []byte("K=0.50;\n"))
	d.write(w, []byte("maxiter=2000;\n"))
	d.write(w, []byte("start=1251;\n"))
	d.write(w, []byte(graphDefaults))
	d.write(w, []byte(nodeDefaults))
	d.write(w, []byte(edgeDefaults))
}

func (d *Dot) endGraph(w io.Writer) {
	d.write(w, []byte("}\n"))
}

func (d *Dot) edge(w io.Writer, a, b, label string) {
	d.write(w, []byte(fmt.Sprintf(edgeTemplate, a, b, label)))
}

func (d *Dot) node(w io.Writer, label, description string) {
	d.write(w, []byte(fmt.Sprintf(nodeTemplate, label, label, description)))
}

func (d *Dot) write(w io.Writer, p []byte) {
	if d.err != nil {
		return
	}

	_, err := w.Write(p)
	if err != nil {
		fmt.Println(err.Error())
	}
	if d.err != nil {
		d.err = err
	}
}
