package optimization

import (
	log "github.com/cihub/seelog"
)

type (
	Parameters struct {
		Input       Data `json:"input"`
		Precedence  `json:"precedence"`
		EngineParam `json:"optimization"`
	}
)

func (ctx *Parameters) optimizing() ([][]bool, int) {

	nReal := len(ctx.Input.Ebv)
	nData := len(ctx.Input.Ebv[0])

	log.Infof("Number of realizations: %v", nReal)
	log.Infof("Number of rows: %v", nData)

	log.Info("Begin creating naive mask")
	mask := ctx.generateMask()

	log.Info("Begin creating precedence")
	if ctx.Precedence.init(ctx, mask) != nil {
		return nil, -1
	}

	//--------------------------------------------------

	log.Info("Updating mask")
	for i := 0; i < nData; i++ {
		if mask[i] {
			if key := ctx.Precedence.keys[i]; key != MISSING {
				for _, off := range ctx.Precedence.defs[key] {
					mask[i+off] = true
				}
			}
		}
	}

	//--------------------------------------------------

	log.Info("Begin compressing")

	var condensedEBV Data
	var condensedPre Precedence

	if !compressEverything(mask, &ctx.Input, &ctx.Precedence, &condensedEBV, &condensedPre) {
		log.Info("ERROR: Compressing everything failed")
		return nil, 1
	}

	// allocate the condensedSolutions
	rows := len(condensedEBV.Ebv)
	solutions := make([][]bool, rows)

	//--------------------------------------------------
	// Solve-em

	log.Info("Begin optimizing")

	for r := 0; r < nReal; r++ {

		engine, e := getEngine(&ctx.EngineParam)

		if engine == nil {
			log.Info("Error: failed initializing optimization engine: %v", e)
			return nil, 1
		}

		row, status := engine.computeSolution(condensedEBV.Ebv[r], &condensedPre)

		if status != 0 {
			return nil, status
		}

		solutions[r] = row

		// Output
		ebv := float64(0)
		count := int64(0)
		for i := range condensedEBV.Ebv[r] {
			if solutions[r][i] {
				ebv += condensedEBV.Ebv[r][i]
				count++
			}
		}
		log.Infof("Completed realization %3v. Blocks: %-6v, EBV: %f", r, count, ebv)
	}

	//--------------------------------------------------
	// Expand the solutions out

	log.Info("Decompressing solutions")

	selection := make([][]bool, nReal)
	for i := range selection {
		selection[i] = make([]bool, nData)
	}

	j := 0
	for i := 0; i < nData; i++ {
		if mask[i] {
			for r := 0; r < nReal; r++ {
				selection[r][i] = solutions[r][j]
			}
			j++
		}
	}

	// Fix air blocks
	log.Info("Fixing air blocks")
	for r := 0; r < nReal; r++ {
		for i := 0; i < nData; i++ {
			if selection[r][i] {
				if key := ctx.Precedence.keys[i]; key != MISSING {
					for _, off := range ctx.Precedence.defs[key] {
						selection[r][i+off] = true
					}
				}
			}
		}
	}

	return selection, 0
}

func (ctx *Parameters) generateMask() []bool {

	n := ctx.Input.Grid.gridCount()
	mask := make([]bool, n)

	for i := 0; i < n; i++ {
		// If one layer's value is greater than 0,then mask -> true
		for _, layer := range ctx.Input.Ebv {
			if len(layer) >= i && layer[i] >= 0 {
				mask[i] = true
				break
			}
		}
	}

	// Erase from end until first non zero value (removes air)
	for i := n - 1; i > 0; i-- {

		air := true

		// If one layer's value is not zero,than the postition is not empty
		for _, layer := range ctx.Input.Ebv {
			if len(layer) >= i && layer[i] != 0 {
				air = false
				break
			}
		}

		if air {
			mask[i] = false
		} else {
			break
		}
	}

	cnt := 0
	for _, v := range mask {
		if v {
			cnt++
		}
	}

	log.Infof("Count of values in mask: %v", cnt)

	return mask
}
