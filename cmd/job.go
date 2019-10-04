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
	"log"
	"strconv"
	"strings"

	"github.com/fuzzitdev/fuzzit/v2/client"
	"github.com/spf13/cobra"
	"gopkg.in/src-d/go-git.v4"
)

// jobCmd represents the job command

var newJob = client.Job{}

var allowedCPUs = map[string]bool{
	"0.1": true,
	"0.2": true,
	"0.3": true,
	"0.4": true,
	"0.5": true,
	"0.6": true,
	"0.7": true,
	"0.8": true,
	"0.9": true,
	"1":   true,
	"1.0": true,
}

var jobCmd = &cobra.Command{
	Use:   "job [target_id] [files...]",
	Short: "create new fuzzing job",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if newJob.Type != "fuzzing" && newJob.Type != "regression" && newJob.Type != "local-regression" {
			log.Fatalf("--type should be either fuzzing, regression or local-regression. Received: %s", newJob.Type)
		}

		if newJob.Engine != "libfuzzer" && newJob.Engine != "jqf" && newJob.Engine != "go-fuzz" {
			log.Fatalf("--engine should be one of libfuzzer/go-fuzz/jqf. Received: %s", newJob.Type)
		}

		if !allowedCPUs[newJob.CPUs] {
			log.Fatalf("got %s cpus. CPUs can only be one of 0.1,0.2,0.3,0.4,0.5,0.6,0.7,0.8,0.9,1,1.0\n", newJob.CPUs)
		}

		if !strings.HasSuffix(newJob.Memory, "Mi") {
			log.Fatalf("got %s memory. Memory should be suffixed by Mi\n", newJob.Memory)
		}
		megabytes := newJob.Memory[:len(newJob.Memory)-2]
		i, err := strconv.Atoi(megabytes)
		if err != nil {
			log.Fatalln(err)
		}
		if i > 2048 {
			log.Fatalf("got %d Mi memory. > 2048Mi memory is only supported for enterprise customers\n", i)
		}

		image := client.HostToDocker[newJob.Host]
		if image == "" {
			if newJob.Host == "" {
				if newJob.Engine == "jqf" {
					image = "openjdk:stretch"
				} else {
					image = "gcr.io/fuzzit-public/stretch-llvm8:64bdedf"
				}
			} else {
				image = newJob.Host
			}
		}
		newJob.Host = image

		skipIfNotExist, err := cmd.Flags().GetBool("skip-if-not-exists")
		if err != nil {
			log.Fatal(err)
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

		if newJob.Type == "local-regression" {
			err = gFuzzitClient.CreateLocalJob(newJob, args[1:])
			if err != nil && skipIfNotExist && (err.Error() == "401 Unauthorized" || err.Error() == "fuzzer exited with 22") {
				log.Println("Target doesn't exist yet. skipping...")
				return
			}
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
	} else {
		revision = client.GetValueFromEnv("TRAVIS_COMMIT", "CIRCLE_SHA1", "GITHUB_SHA")
	}

	branch := client.GetValueFromEnv("TRAVIS_BRANCH", "CIRCLE_BRANCH", "GITHUB_REF")

	jobCmd.Flags().StringVar(&newJob.Type, "type", "fuzzing", "fuzzing/regression/local-regression")
	jobCmd.Flags().StringVar(&newJob.Engine, "engine", "libfuzzer", "libfuzzer/jqf")
	jobCmd.Flags().StringVar(&newJob.CPUs, "cpus", "1", "number of cpus to use (only relevant for fuzzing job)")
	jobCmd.Flags().StringVar(&newJob.Memory, "memory", "2048Mi", "number of cpus to use (only relevant for fuzzing job)")
	jobCmd.Flags().MarkHidden("memory")
	jobCmd.Flags().MarkHidden("cpus")
	jobCmd.Flags().StringVar(&newJob.Revision, "revision", revision, "Revision tag of fuzzer (populates automatically from git,travis,circleci)")
	jobCmd.Flags().StringVar(&newJob.Branch, "branch", branch, "Branch of the fuzzer (populates automatically from git,travis,circleci)")
	jobCmd.Flags().StringVar(&newJob.Host, "host", "", "docker image to use when running the fuzzer. Options: stretch-llvm8/stretch-llvm9/bionic-swift51")
	jobCmd.Flags().StringArrayVarP(&newJob.EnvironmentVariables, "environment", "e", nil,
		"Additional environment variables for the fuzzer. For example ASAN_OPTINOS, UBSAN_OPTIONS or any other")
	jobCmd.Flags().StringVar(&newJob.Args, "args", "", "Additional runtime args for the fuzzer")
	jobCmd.Flags().Bool("skip-if-not-exists", false, "skip/don't fail if target doesnt exists yet. useful for automatic target creation")
}
