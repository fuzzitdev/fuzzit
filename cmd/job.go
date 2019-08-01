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
	"github.com/fuzzitdev/fuzzit/client"
	"log"

	"github.com/spf13/cobra"
)

// jobCmd represents the job command

var newJob = client.Job{}

var jobCmd = &cobra.Command{
	Use:   "job [target_id] [files...]",
	Short: "create new fuzzing job",
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Creating job...")
		c, err := client.LoadFuzzitFromCache()
		if err != nil {
			log.Fatal(err)
		}
		newJob.TargetId = args[0]
		_, err = c.CreateJob(newJob, args[1:])
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Job created successfully")
	},
}

func init() {
	createCmd.AddCommand(jobCmd)

	jobCmd.Flags().StringVar(&newJob.Type, "type", "fuzzing", "fuzzing/sanity")
	jobCmd.Flags().Uint16Var(&newJob.Parallelism, "cpus", 1, "number of cpus to use (only relevant for fuzzing job)")
	jobCmd.Flags().StringVar(&newJob.Revision, "revision", "", "Revision tag of fuzzer")
	jobCmd.Flags().StringVar(&newJob.Branch, "branch", "", "Branch of the fuzzer")
	jobCmd.Flags().StringVar(&newJob.AsanOptions, "asan_options", "", "Additional args to ASAN_OPTIONS env VARIABLE")
	jobCmd.Flags().StringVar(&newJob.UbsanOptions, "ubsan_options", "", "Additional args to UBSAN_OPTIONS env VARIABLE")
	jobCmd.Flags().StringVar(&newJob.Args, "args", "", "Additional runtime args for the fuzzer")
}
