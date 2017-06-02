package optimization

import (
	"fmt"
)

const (
	Engine_LERCHSGROSSMANN = iota + 1
	Engine_DIMACSPROGRAM
	Engine_PSEUDOFLOW
)

const (
	PLUS    = true
	MINUS   = false
	STRONG  = true
	WEAK    = false
	ROOT    = -1
	NOTHING = -1
)

type (
	EngineParam struct {
		EngineType int     `json:"engine"`
		DimacsPath string  `json:"dimacs_path"`
		Precision  float64 `json:"precision"`
	}

	UltpitEngine interface {
		computeSolution(data []float64, pre *Precedence) ([]bool, int)
	}

	LG_Vertex struct {
		mass     float64
		rootEdge int
		myOffs   []int
		inEdges  []int
		outEdges []int
		strength bool
	}

	LG_Edge struct {
		mass      float64
		source    int
		target    int
		direction bool
	}

	IntStack struct {
		items []int
	}

	LG3D struct {
		V                []*LG_Vertex
		E                []*LG_Edge
		arcsAdded        int64
		countSinceChange int64
		count            int

		strongPlusses *IntStack
		strongMinuses *IntStack
	}
)

func getEngine(param *EngineParam) (UltpitEngine, error) {
	switch param.EngineType {
	case Engine_LERCHSGROSSMANN:
		return new(LG3D), nil
	case Engine_DIMACSPROGRAM:
		return newDimacsEngine(param)
	case Engine_PSEUDOFLOW:
		return newPseudoflowEngine(param)
	default:
		return nil, fmt.Errorf("Invalid engine type")
	}
}

func (this *LG3D) computeSolution(data []float64, pre *Precedence) (solution []bool, n int) {

	this.count = len(data)

	solution = make([]bool, this.count)

	this.initNormalizedTree(data, pre)

	this.solve()

	for i := 0; i < this.count; i++ {
		solution[i] = this.V[i].strength
	}

	return
}

func (this *LG3D) initNormalizedTree(data []float64, pre *Precedence) {

	this.V = make([]*LG_Vertex, this.count)
	this.E = make([]*LG_Edge, this.count)

	this.strongPlusses = new(IntStack)
	this.strongMinuses = new(IntStack)

	var vi *LG_Vertex

	for i := 0; i < this.count; i++ {

		if pre.keys[i] != NOTHING {
			vi = &LG_Vertex{myOffs: pre.defs[pre.keys[i]]}
		} else {
			vi = &LG_Vertex{}
		}

		vi.mass = data[i]
		vi.rootEdge = i
		vi.strength = (data[i] > 0)
		this.V[i] = vi

		ei := &LG_Edge{}
		ei.mass = data[i]
		ei.source = ROOT
		ei.target = i
		ei.direction = PLUS
		this.E[i] = ei
	}
}

func (this *LG3D) solve() {

	var xk int

	for this.countSinceChange++; this.countSinceChange <= int64(this.count); this.countSinceChange++ {

		if this.V[xk].strength {

			if xi := this.checkPrecedence(xk); xi != -1 {
				this.moveTowardFeasibility(xk, xi)
				this.arcsAdded++
			}

			for range this.strongPlusses.items {
				this.swapStrongPlus(this.strongPlusses.pop())
			}

			for range this.strongMinuses.items {
				this.swapStrongMinus(this.strongMinuses.pop())
			}
		}

		if xk++; xk >= this.count {
			xk = 0
		}
	}
}

func (this *LG3D) moveTowardFeasibility(xk, xi int) {

	xkStack := this.stackToRoot(xk)
	xiStack := this.stackToRoot(xi)

	lowestRootEdge := xkStack.pop()

	E := this.E
	V := this.V

	baseMass := E[lowestRootEdge].mass
	E[lowestRootEdge].source = xk
	E[lowestRootEdge].target = xi
	E[lowestRootEdge].direction = MINUS

	V[xk].rootEdge = lowestRootEdge
	V[xi].addInEdge(lowestRootEdge)

	// Fix edges along path back to xk
	itemcnt := len(xkStack.items)
	for idx := range xkStack.items {
		e := xkStack.items[itemcnt-1-idx]

		if E[e].direction {

			far := E[e].source
			near := E[e].target

			V[far].removeOutEdge(e)
			V[near].addInEdge(e)

			V[far].rootEdge = e
		} else {
			far := E[e].target
			near := E[e].source

			V[far].removeInEdge(e)
			V[near].addOutEdge(e)

			V[far].rootEdge = e
		}

		E[e].direction = !E[e].direction
		E[e].mass = baseMass - E[e].mass

		if this.isStrong(E[e]) {
			if E[e].direction {
				this.strongPlusses.push(e)
			} else {
				this.strongMinuses.push(e)
			}
		}
	}

	//----------------------------------------

	newRootEdge := xiStack.peek()
	newMass := E[newRootEdge].mass + baseMass

	// Now update the other chain
	itemcnt = len(xiStack.items)
	for idx := range xiStack.items {
		e := xiStack.items[itemcnt-1-idx]

		E[e].mass += baseMass

		if this.isStrong(E[e]) {
			if E[e].direction {
				this.strongPlusses.push(e)
			} else {
				this.strongMinuses.push(e)
			}
		}
	}

	if newMass > 0 {
		this.activateBranchToxk(newRootEdge, xk)
	} else {
		this.deactivateBranch(newRootEdge)
	}

	this.countSinceChange = 0
}

func (this *LG3D) activateBranchToxk(base, xk int) {

	var nextV int

	if this.E[base].direction {
		nextV = this.E[base].target
	} else {
		nextV = this.E[base].source
	}

	if nextV != xk {

		for _, edge := range this.V[nextV].outEdges {
			this.activateBranchToxk(edge, xk)
		}

		for _, edge := range this.V[nextV].inEdges {
			this.activateBranchToxk(edge, xk)
		}
	}

	this.V[nextV].strength = true
}

func (this *LG3D) deactivateBranch(base int) {

	var nextV int

	if this.E[base].direction {
		nextV = this.E[base].target
	} else {
		nextV = this.E[base].source
	}

	for _, edge := range this.V[nextV].outEdges {
		this.deactivateBranch(edge)
	}

	for _, edge := range this.V[nextV].inEdges {
		this.deactivateBranch(edge)
	}

	this.V[nextV].strength = false
}

func (this *LG3D) isStrong(e *LG_Edge) bool {
	if e.source != ROOT && e.target != ROOT {
		return ((e.mass > 0) == e.direction)
	} else {
		return false
	}
}

func (this *LG3D) stackToRoot(k int) *IntStack {

	var next int
	current := k
	stack := new(IntStack)

	for {
		edge := this.V[current].rootEdge

		if this.E[edge].direction {
			next = this.E[edge].source
		} else {
			next = this.E[edge].target
		}

		stack.push(edge)

		current = next

		if next == ROOT {
			break
		}
	}

	return stack
}

func (this *LG3D) checkPrecedence(k int) int {
	for _, off := range this.V[k].myOffs {
		if !this.V[k+off].strength {
			return k + off
		}
	}
	return -1
}

// Normalize
func (this *LG3D) swapStrongPlus(e int) {

	// Ensure that it is still a strong plus.
	if !this.isStrong(this.E[e]) {
		return
	}

	E := this.E
	V := this.V

	source := E[e].source
	target := E[e].target

	thisMass := E[e].mass

	var next, last int

	current := source

	for {
		last = current

		edge := V[current].rootEdge

		if E[edge].direction {
			next = E[edge].source
		} else {
			next = E[edge].target
		}

		E[edge].mass -= thisMass

		if current = next; current == ROOT {
			break
		}
	}

	V[source].removeOutEdge(e)

	E[e].source = ROOT

	baseEdge := V[last].rootEdge
	baseMass := E[baseEdge].mass

	if baseMass > 0 {
		if !V[source].strength {
			this.activateBranchToxk(e, -1)
		}
	} else if V[target].strength {
		this.deactivateBranch(baseEdge)
	}
}

func (this *LG3D) swapStrongMinus(e int) {

	// Ensure that it is still a strong minus.
	if !this.isStrong(this.E[e]) {
		return
	}

	E := this.E
	V := this.V

	source := E[e].source
	target := E[e].target

	thisMass := E[e].mass

	var next int

	current := target

	for {
		edge := V[current].rootEdge

		if E[edge].direction {
			next = E[edge].source
		} else {
			next = E[edge].target
		}

		E[edge].mass -= thisMass

		if current = next; current == ROOT {
			break
		}
	}

	E[e].direction = PLUS
	E[e].target = source
	E[e].source = ROOT
}

//---------------------------------------------------------------------------

func (this *LG_Vertex) addInEdge(e int) {
	this.inEdges = append(this.inEdges, e)
}

func (this *LG_Vertex) addOutEdge(e int) {
	this.outEdges = append(this.outEdges, e)
}

func (this *LG_Vertex) removeInEdge(e int) {
	for i, x := range this.inEdges {
		if x == e {
			cnt := len(this.inEdges)
			copy(this.inEdges[i:], this.inEdges[i+1:])
			this.inEdges = this.inEdges[:cnt-1]
		}
	}
}

func (this *LG_Vertex) removeOutEdge(e int) {
	for i, x := range this.outEdges {
		if x == e {
			cnt := len(this.outEdges)
			copy(this.outEdges[i:], this.outEdges[i+1:])
			this.outEdges = this.outEdges[:cnt-1]
		}
	}
}

//---------------------------------------------------------------------------

func (this *IntStack) push(t int) {
	this.items = append(this.items, t)
}

func (this *IntStack) pop() int {
	if l := len(this.items); l > 0 {
		t := this.items[l-1]
		this.items = this.items[:l-1]
		return t
	} else {
		panic("Empty Stack.")
		return -1
	}
}

func (this *IntStack) peek() int {
	if l := len(this.items); l > 0 {
		return this.items[l-1]
	} else {
		panic("Empty Stack.")
		return -1
	}
}

func (this *IntStack) empty() bool {
	return len(this.items) == 0
}

func (this *IntStack) notEmpty() bool {
	return len(this.items) != 0
}
