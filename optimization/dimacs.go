package optimization

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os/exec"
	"strconv"
	"strings"
)

type (
	DimacsSolver struct {
		dimacs_program string
		precision      float64
		cmd            *exec.Cmd
		stdin          io.WriteCloser
		stdout         io.ReadCloser
	}
)

func newDimacsEngine(param *EngineParam) (UltpitEngine, error) {

	engine := &DimacsSolver{
		dimacs_program: param.DimacsPath,
		precision:      param.Precision,
	}

	if math.Abs(engine.precision) < 1e6 {
		engine.precision = 100.0
	}

	if e := engine.init(); e == nil {
		return engine, nil
	} else {
		return nil, e
	}
}

func (this *DimacsSolver) init() error {

	this.cmd = exec.Command(this.dimacs_program)

	var e error

	if this.stdin, e = this.cmd.StdinPipe(); e != nil {
		return e
	} else if this.stdout, e = this.cmd.StdoutPipe(); e != nil {
		return e
	} else if e = this.cmd.Start(); e != nil {
		return e
	} else {
		return nil
	}
}

func (this *DimacsSolver) computeSolution(data []float64, pre *Precedence) (solution []bool, r int) {

	count := len(data)

	solution = make([]bool, count)

	this.sendInput(data, pre, this.stdin)
	this.stdin.Close()

	scanner := bufio.NewScanner(this.stdout)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		items := strings.Fields(scanner.Text())
		if len(items) == 2 && items[0] == "n" {
			if n, e := strconv.Atoi(items[1]); e == nil && n != 1 {
				solution[n-2] = true
			}
		}
	}

	this.cmd.Wait()

	return
}

func (this *DimacsSolver) sendInput(data []float64, pre *Precedence, st io.Writer) {

	// source and sink
	numNodes := len(data) + 2
	// source to positive nodes, sink to negative nodes
	numArcs := len(data)

	for i := 0; i < len(data); i++ {
		// Each infinite arc
		if ind := pre.keys[i]; ind != MISSING {
			numArcs += len(pre.defs[ind])
		}
	}

	const SOURCE = 1
	SINK := numNodes

	fmt.Fprintf(st, "p max %v %v\n", numNodes, numArcs)
	fmt.Fprintf(st, "n %v s\n", SOURCE)
	fmt.Fprintf(st, "n %v t\n", SINK)

	var from_i, to_i int

	for i := 0; i < len(data); i++ {

		capacity := uint(math.Abs(data[i]) * this.precision)

		if data[i] < 0 {
			from_i = i + 2
			to_i = SINK
		} else {
			from_i = SOURCE
			to_i = i + 2
		}

		fmt.Fprintf(st, "a %v %v %v\n", from_i, to_i, capacity)
	}

	// Now the infinite ones
	for i := 0; i < len(data); i++ {
		from_i = i + 2 // + 1 for psuedo, +1 for source
		if ind := pre.keys[i]; ind != MISSING {
			for _, off := range pre.defs[ind] {
				fmt.Fprintf(st, "a %v %v %v\n", from_i, from_i+off, uint32(math.MaxUint32))
			}
		}
	}
}
