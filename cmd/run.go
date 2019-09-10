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
	"fmt"
	"github.com/fuzzitdev/fuzzit/v2/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
)

// authCmd represents the auth command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run job locally (used by the agent)",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := viper.GetString("api-key")
		gFuzzitClient, err := client.NewFuzzitClient(apiKey)
		if err != nil {
			log.Fatalln(err)
		}

		fuzzingType, err := cmd.Flags().GetString("type")
		if err != nil {
			log.Fatalln(err)
		}

		updateDB, err := cmd.Flags().GetBool("update-db")
		if err != nil {
			log.Fatal(err)
		}

		orgId := os.Getenv("ORG_ID")
		if orgId == "" {
			log.Fatalln(fmt.Errorf("ORG_ID environment variable should be provided"))
		}

		targetId := os.Getenv("TARGET_ID")
		if orgId == "" {
			log.Fatalln(fmt.Errorf("TARGET_ID environment variable should be provided"))
		}

		jobId := ""
		if updateDB {
			jobId = os.Getenv("JOB_ID")
			if orgId == "" {
				log.Fatalln(fmt.Errorf("JOB_ID environment variable should be provided"))
			}
		}

		gFuzzitClient.Org = orgId
		err = gFuzzitClient.RunLibFuzzer(targetId, jobId, updateDB, fuzzingType)

		if err != nil {
			log.Fatalln(err)
		}
	},
	Hidden: true,
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().Bool("update-db", false, "if this runs on fuzzit then update db")
	runCmd.Flags().String("type", "fuzzing", "fuzzing/regression")
}
