package optimization

import (
	"github.com/clbanning/pseudo"
	"io"
	"log"
	"math"
)

type (
	PseudoSolver struct {
		N         struct{} // redundant?
		A         struct{} // likewise
		numNodes  uint
		numArcs   uint
		Precision float64
	}
)

func newPseudoflowEngine(param *EngineParam) (UltpitEngine, error) {

	engine := &PseudoSolver{
		Precision: param.Precision,
	}

	if math.Abs(engine.Precision) < 1e6 {
		engine.Precision = 100.0
	}

	//if e := engine.init(); e == nil {
	return engine, nil
	//} else {
	//	return nil, e
	//}
}

func (p *PseudoSolver) computeSolution(data []float64, pre *Precedence) (solution []bool, r int) {

	count := len(data)

	solution = make([]bool, count)

	p.sendInput(data, pre)
	//p.stdin.Close()
	//
	//scanner := bufio.NewScanner(p.stdout)
	//scanner.Split(bufio.ScanLines)
	//
	//for scanner.Scan() {
	//	items := strings.Fields(scanner.Text())
	//	if len(items) == 2 && items[0] == "n" {
	//		if n, e := strconv.Atoi(items[1]); e == nil && n != 1 {
	//			solution[n-2] = true
	//		}
	//	}
	//}
	//
	//p.cmd.Wait()

	return
}

func (p *PseudoSolver) sendInput(data []float64, pre *Precedence) {

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
	n := []pseudo.N{}
	a := []pseudo.A{}
	const SOURCE = 1
	n = append(n, pseudo.N{Val: SOURCE, Node: "s"})
	SINK := numNodes
	n = append(n, pseudo.N{Val: uint(SINK), Node: "t"})

	//fmt.Fprintf(st, "p max %v %v\n", numNodes, numArcs)
	//fmt.Fprintf(st, "n %v s\n", SOURCE)
	//fmt.Fprintf(st, "n %v t\n", SINK)

	var from_i, to_i int

	for i := 0; i < len(data); i++ {
		// incorrect hack using precision to avoid decimals and cast into uint
		// TODO: Actually we need to refactor pseudo.go etc such that
		// capacity is a float64 as it is generally a large +- number
		//capacity := math.Abs(data[i]) * p.Precision
		//capacity := data[i] * p.Precision
		capacity := uint(math.Abs(data[i] * p.Precision))
		if data[i] < 0 {
			from_i = i + 2
			to_i = SINK
		} else {
			from_i = SOURCE
			to_i = i + 2
		}
		a = append(a, pseudo.A{From: uint(from_i), To: uint(to_i), Capacity: int(capacity)})

		//fmt.Print(a)
		//fmt.Fprintf(st, "a %v %v %v\n", from_i, to_i, capacity)
		//fmt.Printf("a %v %v %v\n", from_i, to_i, capacity)
	}

	// Now the infinite ones
	for i := 0; i < len(data); i++ {
		from_i = i + 2 // + 1 for pseudo, +1 for source
		if ind := pre.keys[i]; ind != MISSING {
			for _, off := range pre.defs[ind] {
				a = append(a, pseudo.A{From: uint(from_i), To: uint(to_i + off), Capacity: math.MaxInt32})
				//fmt.Print(a)
				//fmt.Print("a %v %v %v\n", from_i, from_i+off, uint32(math.MaxUint32))
			}
		}
	}
	//var header string
	var w io.Writer
	header := H

	//s := pseudo.NewSession(pseudo.Context{LowestLabel: false, FifoBuckets: true, DisplayCut: true})
	s := pseudo.NewSession(pseudo.Context{})
	//(*pseudo.Session).RunNAWriter(s, uint(numNodes), uint(numArcs), n, a, w, header)
	var err error
	//var buf bytes.Buffer
	if err = s.RunNAWriter(uint(numNodes), uint(numArcs), n, a, w, header); err != nil {
		log.Fatal(err)
	}
	//result := string(buf.Bytes())
	//fmt.Print(result)
}
