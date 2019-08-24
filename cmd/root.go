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
	"log"
	"os"
	"strings"

	"github.com/fuzzitdev/fuzzit/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var gFuzzitClient *client.FuzzitClient

var rootCmd = &cobra.Command{
	Use:     "fuzzit",
	Short:   "Continuous fuzzing made simple CLI",
	Version: "2.4.32",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().String("api-key", "", "Authentication token (can also be passed via env: FUZZIT_API_KEY)")
	if err := viper.BindPFlag("api-key", rootCmd.PersistentFlags().Lookup("api-key")); err != nil {
		log.Fatalln(err)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("FUZZIT")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
}
