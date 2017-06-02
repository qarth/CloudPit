package optimization

import (
	log "github.com/cihub/seelog"
)

func compressEverything(
	mask []bool,
	data *Data,
	precedence *Precedence,
	condensedEBV *Data,
	condensedPre *Precedence,
) bool {

	var count int

	for _, v := range mask {
		if v {
			count++
		}
	}

	log.Infof("Original: %v", len(mask))
	log.Infof("Compressed: %v", count)
	log.Infof("Percent Reduction: %f", float64(len(mask)-count)/float64(len(mask))*100.0)
	log.Infof("Precedence keys count: %v", len(precedence.keys))

	if !compressData(mask, data, count, condensedEBV) {
		return false
	}

	if !compressPrecedence(mask, count, precedence, condensedPre) {
		return false
	}

	condensedPre.logExtraInfo()

	return true
}

func compressData(mask []bool, data *Data, count int, condensedEBV *Data) bool {

	nReal := len(data.Ebv)

	condensedEBV.Ebv = make([][]float64, nReal)
	for i := range condensedEBV.Ebv {
		condensedEBV.Ebv[i] = make([]float64, count)
	}

	var j int

	for i, v := range mask {
		if v {
			for r := 0; r < nReal; r++ {
				condensedEBV.Ebv[r][j] = data.Ebv[r][i]
			}
			j++
		}
	}

	return true
}

func compressPrecedence(mask []bool, count int, pre *Precedence, condensedPre *Precedence) bool {

	zeroesBefore := make([]int, len(pre.keys))
	condensedPre.keys = make([]int, count)

	var currentZeroes, currentKey int

	j := count - 1

	for i := len(pre.keys) - 1; i >= 0; i-- {

		if !mask[i] {
			currentZeroes++
			zeroesBefore[i] = currentZeroes
		} else if pre.keys[i] == -1 {
			condensedPre.keys[j] = -1
			j--
			zeroesBefore[i] = currentZeroes
		} else {

			thisNewDef := []int{}

			for _, off := range pre.defs[pre.keys[i]] {
				if mask[i+off] {
					offZeroes := zeroesBefore[i+off]
					thisNewDef = append(thisNewDef, off-currentZeroes+offZeroes)
				}
			}

			if len(thisNewDef) > 0 {

				if len(condensedPre.defs) == 0 {
					condensedPre.defs = append(condensedPre.defs, thisNewDef)
				}

				if !sliceEqual(condensedPre.defs[currentKey], thisNewDef) {
					condensedPre.defs = append(condensedPre.defs, thisNewDef)
					currentKey++
				}
				condensedPre.keys[j] = currentKey
			} else {
				condensedPre.keys[j] = -1
			}

			j--
			zeroesBefore[i] = currentZeroes
		}
	}

	return true
}
