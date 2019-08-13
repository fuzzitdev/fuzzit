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
	"github.com/fuzzitdev/fuzzit/client"
	"github.com/spf13/cobra"
	"gopkg.in/src-d/go-git.v4"
	"log"
	"strings"
)

// jobCmd represents the job command

var newJob = client.Job{}

var jobCmd = &cobra.Command{
	Use:   "job [target_id] [files...]",
	Short: "create new fuzzing job",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if newJob.Type != "fuzzing" && newJob.Type != "regression" {
			log.Fatalf("--type should be either fuzzing or regression. Recieved: %s", newJob.Type)
		}

		log.Println("Creating job...")

		target := args[0]
		targetSplice := strings.Split(args[0], "/")
		if len(targetSplice) > 2 {
			log.Fatalf("[TARGET] can only be of type 'target' or 'project/target-name'.")
		} else if len(targetSplice) == 2 {
			target = targetSplice[1]
			gFuzzitClient.Org = targetSplice[0]
		}

		newJob.TargetId = target

		var err error = nil
		if newJob.Local {
			err = gFuzzitClient.CreateLocalJob(newJob, args[1:])
		} else {
			_, err = gFuzzitClient.CreateJob(newJob, args[1:])
		}

		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Job created successfully")
	},
}

func init() {
	createCmd.AddCommand(jobCmd)

	revision := ""
	r, err := git.PlainOpen("")
	if err == nil {
		revisionHex, err := r.ResolveRevision("HEAD")
		if err == nil {
			revision = revisionHex.String()
		}
	}

	jobCmd.Flags().StringVar(&newJob.Type, "type", "fuzzing", "fuzzing/regression")
	jobCmd.Flags().BoolVar(&newJob.Local, "local", false, "run fuzzing/regression locally in a docker")
	jobCmd.Flags().Uint16Var(&newJob.Parallelism, "cpus", 1, "number of cpus to use (only relevant for fuzzing job)")
	jobCmd.Flags().StringVar(&newJob.Revision, "revision", revision, "Revision tag of fuzzer")
	jobCmd.Flags().StringVar(&newJob.Branch, "branch", "master", "Branch of the fuzzer")
	jobCmd.Flags().StringArrayVarP(&newJob.EnvironmentVariables, "environment", "e", nil,
		"Additional environment variables for the fuzzer. For example ASAN_OPTINOS, UBSAN_OPTIONS or any other")
	jobCmd.Flags().StringVar(&newJob.Args, "args", "", "Additional runtime args for the fuzzer")
}
