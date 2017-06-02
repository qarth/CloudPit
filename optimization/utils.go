package optimization

import (
	"encoding/json"
	"io/ioutil"

	log "github.com/cihub/seelog"
)

func readJsonFile(file string, ptr interface{}) error {

	if c, e := ioutil.ReadFile(file); e != nil {
		log.Errorf("Error: failed initializing parameters: %v", e)
		return e
	} else if e = json.Unmarshal(c, ptr); e != nil {
		log.Errorf("Error: failed initializing parameters: %v", e)
		return e
	} else {
		return nil
	}
}

func sliceEqual(x, y []int) bool {
	if len(x) == 0 && len(y) == 0 {
		return true
	} else if len(x) != len(y) {
		return false
	} else {
		for i, a := range x {
			if y[i] != a {
				return false
			}
		}
		return true
	}
}
