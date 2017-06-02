package optimization

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"

	log "github.com/cihub/seelog"
)

type (
	MiningOptParams struct {
		InputFile  string
		OutputFile string
		ParamFile  string
	}
)

var H string

func DoMiningOptimization(opt MiningOptParams) {

	log.Info("Being parsing parameters")

	var params Parameters

	if readJsonFile(opt.ParamFile, &params) != nil {
		return
	}

	log.Info("Begin reading input")
	H = opt.InputFile
	if params.Input.initializeFromGzip(opt.InputFile) != nil {
		return
	}

	selection, status := params.optimizing()

	if status != 0 {
		log.Info("ERROR: failed optimizing")
		return
	}

	var writer io.Writer
	var write_head bool
	var doclose func() error

	if len(opt.OutputFile) == 0 {
		writer = os.Stdout
		write_head = true
	} else {

		file, e := os.Create(opt.OutputFile)
		if e != nil {
			log.Infof("Failed to create output file %v: %v", opt.OutputFile, e)
			return
		}
		defer file.Close()
		writer = file

		if strings.HasSuffix(opt.OutputFile, ".gz") {
			zipwriter := gzip.NewWriter(writer)
			doclose = zipwriter.Close
			writer = zipwriter
		} else {
			write_head = false
		}
	}

	if write_head {
		fmt.Fprintln(writer, "ultpit output")
		fmt.Fprintln(writer, "1")
		fmt.Fprintln(writer, "Pit")
	}

	for _, row := range selection {
		for _, v := range row {
			if v {
				fmt.Fprintln(writer, "1")
			} else {
				fmt.Fprintln(writer, "0")
			}
		}
	}

	if doclose != nil {
		doclose()
	}
}
