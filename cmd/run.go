/*
Copyright © 2019 fuzzit.dev, inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"github.com/fuzzitdev/fuzzit/v2/client"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var runJob = client.Job{}

// authCmd represents the auth command
var runCmd = &cobra.Command{
	Use:   "run ORG_ID TARGET_ID [JOB_ID]",
	Short: "Run job locally (used by the agent)",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		updateDB, err := cmd.Flags().GetBool("update-db")
		if err != nil {
			log.Fatal(err)
		}

		gFuzzitClient.Org = args[0]
		runJob.OrgId = args[0]
		runJob.TargetId = args[1]
		jobId := ""
		if len(args) == 3 {
			jobId = args[2]
		}

		err = gFuzzitClient.RunFuzzer(runJob, jobId, updateDB)
		if err != nil {
			log.Println(err)
			if err.Error() == "401 Unauthorized" {
				os.Exit(client.AgentNoPermissionError)
			} else {
				os.Exit(client.AgentGeneralError)
			}
		}
	},
	Hidden: true,
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().Bool("update-db", false, "if this runs on fuzzit then update db")
	runCmd.Flags().StringVar(&runJob.Type, "type", "fuzzing", "fuzzing/regression")
	runCmd.Flags().StringVar(&runJob.Engine, "engine", "libfuzzer", "libfuzzer/jqf/go-fuzz")
	runCmd.Flags().StringVar(&runJob.Args, "args", "", "Additional runtime args for the fuzzer")
}
