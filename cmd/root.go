// Copyright © 2017 Robert Wright a1210993@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/cihub/seelog"
	"github.com/qarth/CloudPit/optimization"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	log_cfg_tmpl = `<seelog minlevel="info">
				<outputs formatid="detail">
					{{OutputDest}}
				</outputs>
				<formats>
					<format id="detail" format="[%File:%Line][%Date(2006-01-02 15:04:05.000)] %Msg%n" />
				</formats>
			</seelog>`
	log_out_dest  = "{{OutputDest}}"
	log_file_tmpl = `<rollingfile filename="%s" type="size" maxsize="10247680" maxrolls="10"/>`
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "CloudPit",
	Short: fmt.Sprintf("%v %v %v", PROGRAM_NAME, PROGRAM_VERSION, COPYRIGHT),
	Long:  fmt.Sprintf("Usage: %s [options] parameter_file", PROGRAM_NAME),
	Run: func(cmd *cobra.Command, args []string) {
		doMiningOperation(cmd, args)
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {

	flagset := RootCmd.PersistentFlags()

	// These are global options
	flagset.StringP("input", "i", "", "The input file")
	flagset.StringP("output", "o", "", "The output file")
	flagset.StringP("log", "l", "", "Log information to a file")
}

func doMiningOperation(cmd *cobra.Command, args []string) {

	viper.BindPFlags(cmd.Flags())

	logfile := viper.GetString("log")
	infile := viper.GetString("input")
	outfile := viper.GetString("output")

	if len(infile) == 0 || len(outfile) == 0 || len(args) != 1 {
		cmd.Usage()
		return
	}

	//-------

	outputDest := "<console/>"

	if len(logfile) > 0 {
		outputDest = fmt.Sprintf(log_file_tmpl, logfile)
	}

	log_cfg := strings.Replace(log_cfg_tmpl, log_out_dest, outputDest, -1)

	logger, _ := log.LoggerFromConfigAsString(log_cfg)

	if logger != nil {
		log.ReplaceLogger(logger)
	}

	//-------

	param := optimization.MiningOptParams{
		InputFile:  infile,
		OutputFile: outfile,
		ParamFile:  args[0],
	}

	log.Info("ultpit begin")
	optimization.DoMiningOptimization(param)
	log.Info("ultpit finished")

	time.Sleep(time.Millisecond * 300)
}
