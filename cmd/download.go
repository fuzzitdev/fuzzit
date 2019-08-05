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
	"github.com/spf13/cobra"
	"log"
	"strings"
)

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download [TARGET] [corpus|seed] [LOCAL_PATH]",
	Short: "download seed/corpus",
	Example: `./fuzzit download fuzzit/example-go corpus
./fuzzit download example-go corpus # This can work if you are already authenticated`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		target := args[0]
		resource := args[1]

		c, err := client.LoadFuzzitFromCache()
		if err != nil {
			log.Fatal(err)
		}

		targetSplice := strings.Split(target, "/")
		if len(targetSplice) > 2 {
			log.Fatalf("[TARGET] can only be of type 'target' or 'project/target-name'.")
		} else if len(targetSplice) == 2 {
			target = targetSplice[1]
		}

		if c.Org == "" {
			if len(targetSplice) != 2 {
				log.Fatalf("For unauthenticated requests [TARGET] should be 'project/target-name'")
			}
			c.Org = targetSplice[0]
		}

		if resource == "corpus" {
			localPath := "corpus"
			if len(args) > 2 {
				localPath = args[2]
			}
			log.Print("downloading corpus")
			err := c.DownloadCorpus(localPath, target)
			if err != nil {
				log.Fatal(err)
			}
		} else if resource == "seed" {
			localPath := "seed"
			if len(args) > 2 {
				localPath = args[2]
			}
			err := c.DownloadSeed(localPath, target)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatalf("resource should be either corpus or seed")
		}

	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)
}
