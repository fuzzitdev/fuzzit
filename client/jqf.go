package client

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"syscall"
	"time"
)

func jqfExitCodeToStatus(exitCode int) string {
	status := "pass"
	switch exitCode {
	case jqfCrashExitCode:
		status = "crash"
	case jqfSuccessExitCode:
		status = "pass"
	default:
		status = "failed"
	}

	return status
}

func (c *FuzzitClient) runJQFFuzzing() error {
	ctx := context.Background()

	log.Println("downloading zest cli...")
	err := DownloadFile("zest-cli.jar", "https://storage.googleapis.com/public-fuzzit/jqf-fuzz-1.3-SNAPSHOT-zest-cli.jar")

	args := []string{
		"-jar",
		"zest-cli.jar",
		"--exit-on-crash",
		"--exact-crash-path=artifact",
		"--libfuzzer-compat-output",
		"fuzzer",
	}
	if c.currentJob.Args != "" {
		args = append(args, splitAndRemoveEmpty(c.currentJob.Args, " ")...)
	}

	path, err := exec.LookPath("java")
	if err != nil {
		return fmt.Errorf("java must be installed in the docker to run JQF fuzzer")
	}

	var exitCode int
	for err == nil {
		log.Println(err)
		log.Printf("Running fuzzing with: %v", args)
		cmd := exec.Command(path,
			args...)
		if err := appendPrefixToCmd(cmd); err != nil {
			return err
		}
		err = cmd.Start()
		// Use a channel to signal completion so we can use a select statement
		done := make(chan error)
		go func() { done <- cmd.Wait() }()
		timeout := time.After(60 * time.Second)
		stopSession := false
		for !stopSession {
			select {
			case <-timeout:
				var fuzzingJob Job
				if err := c.refreshToken(); err != nil {
					return err
				}
				docRef := c.firestoreClient.Doc(fmt.Sprintf("orgs/%s/targets/%s/jobs/%s", c.Org, c.currentJob.TargetId, c.jobId))
				if docRef == nil {
					return fmt.Errorf("invalid resource")
				}
				docsnap, err := docRef.Get(ctx)
				if err != nil {
					return err
				}
				err = docsnap.DataTo(&fuzzingJob)
				if err != nil {
					return err
				}
				if fuzzingJob.Status == "in progress" {
					timeout = time.After(60 * time.Second)
				} else {
					log.Println("job was cancel by user. exiting...")
					cmd.Process.Kill()
					return nil
				}
			case err = <-done:
				stopSession = true
				if err != nil {
					log.Printf("process finished with error = %v\n", err)
					if exiterr, ok := err.(*exec.ExitError); ok {
						// The program has exited with an exit code != 0

						// This works on both Unix and Windows. Although package
						// syscall is generally platform dependent, WaitStatus is
						// defined for both Unix and Windows and in both cases has
						// an ExitStatus() method with the same signature.
						if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
							exitCode = status.ExitStatus()
							log.Printf("Exit Status: %d", status.ExitStatus())
						}
					} else {
						return err
					}
				} else {
					if err = c.runlibFuzzerMerge(); err != nil {
						return err
					}
					log.Print("process finished successfully")
				}
			}
		}
	}

	if err := c.refreshToken(); err != nil {
		return err
	}
	err = c.uploadCrash(exitCode)
	if err != nil {
		return err
	}

	err = c.transitionStatus(jqfExitCodeToStatus(exitCode))
	if err != nil {
		return err
	}

	return nil
}

func (c *FuzzitClient) RunJQF() error {
	var err error

	if c.currentJob.Type == "fuzzing" {
		err = c.runJQFFuzzing()
	} else {
		return fmt.Errorf("JQF currently only supports fuzzing")
	}

	return err
}
