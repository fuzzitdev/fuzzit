package main

import (
	"encoding/json"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"time"
)
import "fmt"

var fuzzitCli *FuzzitCli


type storageLinkResponse struct {
	StorageLink string `json:"storage_link"`
}


func loadConfig(c * cli.Context) error {
	if len(c.Args()) == 0 {
		return nil
	}

	command := c.Args().First()
	if command == "auth" {
		return nil
	}

	usr, err := user.Current()
	if err != nil {
		return err
	}
	// reload configuration
	fuzzitCli, err = loadFromFile(usr.HomeDir + "/.fuzzit/conf")
	if err != nil {
		return err
	}

	return nil
}


func main() {

	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Name = "Fuzzit"
	app.Usage = "Continuous fuzzing made simple"
	app.Version = "1.2.2"
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "fuzzit.dev",
			Email: "info@fuzzit.dev",
		},
	}
	app.Copyright = "fuzzit.dev by (c) fuzzit.dev inc"

	app.Before = loadConfig

	app.Commands = []cli.Command{
		{
			Name:    "auth",
			Aliases: []string{"a"},
			Usage:   "Authenticate with Fuzzit servers",
			ArgsUsage: "[API_KEY]",
			Action:  func(c *cli.Context) error {
				if len(c.Args()) != 1 {
					return cli.NewExitError("You must specify an ApiKey (available in the dashboard under settings)", 1)
				}

				credsJson, err := json.MarshalIndent(map[string]string{"ApiKey": c.Args().First(), "Org": c.Args().Get(1)}, "", "")
				if err != nil {
					return err
				}

				usr, err := user.Current()
				if err != nil {
					log.Fatal(err)
				}
				confDir := usr.HomeDir + "/.fuzzit"
				confFile := confDir + "/conf"

				if _, err := os.Stat(confDir); os.IsNotExist(err) {
					err = os.Mkdir(confDir, 0700)
					if err != nil {
						return err
					}
				}

				err = ioutil.WriteFile(confFile, credsJson, 0644)
				if err != nil {
					return err
				}
				// Test Reloading the creds
				_, err = loadFromFile(confFile)
				if err != nil {
					return err
				}

				fmt.Println("Authenticated successfully")

				return nil
			},
		},
		{
			Name:    "get",
			Aliases: []string{"g"},
			Usage:   "Display specific resource",
			Action:  func(c *cli.Context) error {
				if len(c.Args()) == 0 {
					return cli.NewExitError("You must specify a resource", 1)
				}
				err := fuzzitCli.getResource(c.Args().First())
				if err != nil {
					return err
				}

				return nil
			},
		},
		{
			Name:    "create",
			Aliases: []string{"c"},
			Usage:   "create a resource",
			Subcommands: []cli.Command{
				{
					Name:  "target",
					Usage: "create a new fuzzing target",
					ArgsUsage: "[TARGET_NAME]",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name: "seed",
							Usage: "path to tar.gz seed corpus",
						},
					},
					Action: func(c *cli.Context) error {
						if len(c.Args()) == 0 {
							return cli.NewExitError("You must specify a target name", 1)
						}
						name := c.Args().First()
						docRef, err := fuzzitCli.createTarget(name, c.String("seed"))
						if err != nil {
							return err
						}
						fmt.Printf("Created new target: %s successfully with id: %s\n", name, docRef.ID)
						return nil
					},
				},
				{
					Name:  "job",
					Usage: "create a new fuzzing job",
					ArgsUsage: "[TARGET_ID] [FUZZER_PATH] [...additional files that should reside in the same path like shared libraries]",
					Flags: []cli.Flag{
						cli.DurationFlag{
							Name: "max_total_time",
							Usage: "maximum of seconds to run",
						},
						cli.UintFlag{
							Name: "cpus",
							Usage: "number of cpus to use (only relevant for fuzzing job)",
							Value: 1,
						},
						cli.StringFlag{
							Name: "type",
							Usage: "choose one from (fuzzing, sanity, merge)",
							Value: "fuzzing",
						},
						cli.StringFlag{
							Name: "host",
							Usage: "choose one from (stretch-llvm9, stretch-llvm8, bionic-llvm7)",
							Value: "stretch-llvm9",
						},
						cli.StringFlag{
							Name: "revision",
							Usage: "optional revision tag",
							Value: "",
						},
						cli.StringFlag{
							Name: "branch",
							Usage: "optional branch tag",
							Value: "",
						},
						cli.StringFlag{
							Name: "args",
							Usage: "additional args to libFuzzer",
							Value: "",
						},
						cli.StringFlag{
							Name: "asan_options",
							Usage: "additional args to ASAN_OPTIONS env variable",
							Value: "",
						},
						cli.StringFlag{
							Name: "ubsan_options",
							Usage: "additional args to UBSAN_OPTIONS env variable",
							Value: "",
						},
					},
					Action: func(c *cli.Context) error {
						if len(c.Args()) < 2 {
							return cli.NewExitError("You must specify a target ID and a fuzzer path", 1)
						}

						_, err := fuzzitCli.createJob(c.Args().First(),
													  c.String("type"),
													  c.String("host"),
													  c.String("args"),
													  c.String("asan_options"),
													  c.String("ubsan_options"),
													  c.String("revision"),
													  c.String("branch"),
													  uint16(c.Uint("cpus")),
													  c.Args()[1:])
						if err != nil {
							return err
						}
						return nil
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

