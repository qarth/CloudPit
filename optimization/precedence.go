package optimization

import (
	"fmt"
	"math"

	log "github.com/cihub/seelog"
)

type (
	Precedence struct {
		Method     int     `json:"method"`
		Slope      float64 `json:"slope"`
		NumBenches int     `json:"num_benches"`
		//-------------------------------------
		keys []int
		defs [][]int
	}
)

const (
	BENCH       = 1
	MISSING     = -1
	MIN_BENCHES = 1
	MAX_BENCHES = 15
	MIN_SLOPE   = 10.0
	MAX_SLOPE   = 80.0
)

func (this *Precedence) init(ctx *Parameters, mask []bool) error {

	var e error

	if this.Method != BENCH {
		e = fmt.Errorf("Invalid Precedence method")
	} else if this.NumBenches < MIN_BENCHES || this.NumBenches > MAX_BENCHES {
		e = fmt.Errorf(
			"ERROR: benches must be between %v and %v. Supplied: %v",
			MIN_BENCHES, MAX_BENCHES, this.NumBenches,
		)
	} else if this.Slope < MIN_SLOPE || this.Slope > MAX_SLOPE {
		e = fmt.Errorf(
			"ERROR: slope must be between %v and %v. Supplied: %v",
			MIN_SLOPE, MAX_SLOPE, this.Slope,
		)
	} else if len(mask) != ctx.Input.Grid.gridCount() {
		e = fmt.Errorf("ERROR: mask size does not equal grid size")
	}

	if e != nil {
		log.Error(e)
		return e
	} else {
		this.genBench(ctx, mask)
		this.logExtraInfo()
		return nil
	}
}

func (this *Precedence) genBench(ctx *Parameters, mask []bool) {

	pg := &ctx.Input.Grid

	theta := this.Slope * math.Pi / 180.0
	maxVert := float64(this.NumBenches) * pg.SizZ
	maxRadius := maxVert / math.Tan(theta)

	xblock := int(maxRadius / pg.SizX)
	yblock := int(maxRadius / pg.SizY)

	xblocks := xblock*2 + 1
	yblocks := yblock*2 + 1
	zblocks := this.NumBenches

	xcenter := xblock
	ycenter := yblock

	log.Info("Template size in blocks")
	log.Infof("  x: %v", xblocks)
	log.Infof("  y: %v", yblocks)
	log.Infof("  z: %v", zblocks)

	//----------------------------------------

	offTemplate := make([][][]bool, zblocks)

	// Generate base template
	for z := 0; z < zblocks; z++ {

		zloc := float64(z+1) * pg.SizZ
		rad := zloc / math.Tan(theta)
		rad2 := rad * rad

		dimy := make([][]bool, yblocks)

		for y := 0; y < yblocks; y++ {

			yloc := float64(y-ycenter) * pg.SizY
			yloc2 := yloc * yloc

			dimx := make([]bool, xblocks)

			for x := 0; x < xblocks; x++ {

				xloc := float64(x-xcenter) * pg.SizX
				xloc2 := xloc * xloc

				dimx[x] = (xloc2+yloc2 <= rad2)
			}

			dimy[y] = dimx
		}

		offTemplate[z] = dimy
	}

	log.Infof("Number of naive arcs in template: %v", this.countTemplate(offTemplate))

	//---------------------------------------------------------------------------

	for z := zblocks - 1; z > 0; z-- {
		nz := z - 1
		for y := 0; y < yblocks; y++ {
			for x := 0; x < xblocks; x++ {
				if offTemplate[nz][y][x] {
					offTemplate[z][y][x] = false
				}
			}
		}
		offTemplate[z][yblock][xblock] = true
	}

	log.Infof("  after basic trimming: %v", this.countTemplate(offTemplate))

	//---------------------------------------------------------------------------

	var firstDef, ixs, iys, izs []int

	for z := 0; z < zblocks; z++ {
		zl := z + 1
		for y := 0; y < yblocks; y++ {
			yl := y - yblock
			for x := 0; x < xblocks; x++ {
				xl := x - xblock
				if offTemplate[z][y][x] {
					n := pg.gridIndex(xl, yl, zl)
					firstDef = append(firstDef, n)
					ixs = append(ixs, xl)
					iys = append(iys, yl)
					izs = append(izs, zl)
				}
			}
		}
	}

	this.addToDefs(firstDef)

	//---------------------------------------------------------------------------

	this.keys = make([]int, pg.gridCount())
	for i := range this.keys {
		this.keys[i] = MISSING
	}

	hit := make([]bool, pg.gridCount())
	loc := 0

	in := func(v, limit int) bool { return 0 <= v && v < limit }

	for z := 0; z < pg.NumZ-1; z++ {
		for y := 0; y < pg.NumY; y++ {
			for x := 0; x < pg.NumX; x++ {

				if !hit[loc] && !mask[loc] {
					loc++
					continue
				}

				var thisdef []int

				for i := range ixs {

					xl := x + ixs[i]
					yl := y + iys[i]
					zl := z + izs[i]

					if in(xl, pg.NumX) && in(yl, pg.NumY) && in(zl, pg.NumZ) {
						ind := pg.gridIndex(ixs[i], iys[i], izs[i])
						thisdef = append(thisdef, ind)
						hit[loc+ind] = true
					}
				}

				if len(thisdef) > 0 {
					this.keys[loc] = this.addToDefs(thisdef)
				}

				loc++
			}
		}
	}
}

// Try to add the given definition to the defs, return the key
func (this *Precedence) addToDefs(defs []int) int {

	// Check for duplicates
	for idx, array := range this.defs {

		if len(array) != len(defs) {
			continue
		}

		same := true

		for i := range array {
			if array[i] != defs[i] {
				same = false
				break
			}
		}

		if same {
			return idx
		}
	}

	this.defs = append(this.defs, defs)

	return len(this.defs) - 1
}

// count the trues in the template
func (this *Precedence) countTemplate(temp [][][]bool) (n int) {
	for _, bench := range temp {
		for _, row := range bench {
			for _, v := range row {
				if v {
					n++
				}
			}
		}
	}
	return
}

func (this *Precedence) logExtraInfo() {

	var count int
	var arcCount int64

	for _, v := range this.keys {
		if v != MISSING {
			count++
			arcCount += int64(len(this.defs[v]))
		}
	}

	log.Infof("Number of keys: %v", len(this.keys))
	log.Infof("  with arcs: %v", count)
	log.Infof("  without: %v", len(this.keys)-count)
	log.Infof("Number of different arc templates: %v", len(this.defs))

	log.Infof("Number of uncompressed arcs: %v", arcCount)
}
