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

	"github.com/fuzzitdev/fuzzit/v2/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new Target or a Job",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		apiKey := viper.GetString("api-key")
		var err error
		if apiKey != "" {
			gFuzzitClient, err = client.NewFuzzitClient(apiKey)
			if err != nil {
				log.Fatalln(err)
			}
		} else {
			gFuzzitClient, err = client.LoadFuzzitFromCache()
			if err != nil {
				log.Fatalln(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}
