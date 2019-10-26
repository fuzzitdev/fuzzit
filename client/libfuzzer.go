package client

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

func libFuzzerExitCodeToStatus(exitCode int) string {
	status := "pass"
	switch exitCode {
	case libFuzzerTimeoutExitCode:
		status = "timeout"
	case libFuzzerCrashExitCode:
		status = "crash"
	case libFuzzerLeakExitCode:
		status = "crash"
	case libFuzzerOOMExitCode:
		status = "oom"
	case libFuzzerSuccessExitCode:
		status = "pass"
	default:
		status = "failed"
	}

	return status
}

func (c *FuzzitClient) runlibFuzzerMerge() error {
	isEmpty, err := IsDirEmpty("corpus")
	if err != nil {
		return err
	}
	if isEmpty {
		log.Println("nothing to merge. skipping...")
		return nil
	}

	// this directory should not exist but do to some old bugs it might exist in the corpus, so we delete it
	if _, err := os.Stat("merge"); err == nil {
		if err = os.RemoveAll("merge"); err != nil {
			return err
		}
	}
	if err := os.Mkdir("merge", 0644); err != nil {
		return err
	}
	if _, err := os.Stat("/tmp/merge_control.txt"); err == nil {
		if err = os.Remove("/tmp/merge_control.txt"); err != nil {
			return err
		}
	}
	args := []string{
		"-print_final_stats=1",
		"-exact_artifact_path=./artifact",
		fmt.Sprintf("-error_exitcode=%d", libFuzzerTimeoutExitCode),
		"-merge_control_file=/tmp/merge_control.txt",
		"-merge=1",
		"merge",
		"corpus",
	}

	log.Println("Running merge with: ./fuzzer " + strings.Join(args, " "))
	cmd := exec.Command("./fuzzer",
		args...)
	if err := appendPrefixToCmd(cmd); err != nil {
		return err
	}

	if err := cmd.Run(); err != nil {
		return err
	}

	if err := c.refreshToken(); err != nil {
		return err
	}

	if err := os.RemoveAll("corpus"); err != nil {
		return err
	}

	if err := os.Rename("merge", "corpus"); err != nil {
		return err
	}

	if err := c.archiveAndUpload("corpus",
		fmt.Sprintf("orgs/%s/targets/%s/corpus.tar.gz", c.Org, c.currentJob.TargetId),
		"corpus.tar.gz"); err != nil {
		return err
	}

	return nil
}

func (c *FuzzitClient) uploadCrash(exitCode int) error {
	ctx := context.Background()

	if !c.updateDB {
		return nil
	}

	if _, err := os.Stat("artifact"); err == nil {
		colRef := c.firestoreClient.Collection(fmt.Sprintf("orgs/%s/targets/%s/jobs/%s/crashes", c.Org, c.currentJob.TargetId, c.jobId))
		crashRef := colRef.NewDoc()

		log.Printf("uploading crash...")
		if err = c.uploadFile("artifact",
			fmt.Sprintf("orgs/%s/targets/%s/jobs/%s/crashes/%s", c.Org, c.currentJob.TargetId, c.jobId, crashRef.ID),
			fmt.Sprintf("crash-%s", c.currentJob.TargetId)); err != nil {
			return err
		}

		_, err = crashRef.Set(ctx, crash{
			TargetName: c.currentJob.TargetId,
			JobId:      c.jobId,
			TargetId:   c.currentJob.TargetId,
			OrgId:      c.Org,
			ExitCode:   uint32(exitCode),
			Type:       "crash",
			LastLines:  strings.Join(lastLines, "\n"),
			V2:         true,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *FuzzitClient) runLibFuzzerFuzzing() error {
	ctx := context.Background()

	args := append(
		[]string{
			"-print_final_stats=1",
			"-exact_artifact_path=./artifact",
			"-error_exitcode=76",
			"-max_total_time=3600",
			"corpus",
			"additional-corpus",
			"seed",
		},
	)

	if c.currentJob.Args != "" {
		args = append(args, splitAndRemoveEmpty(c.currentJob.Args, " ")...)
	}

	var err error
	err = nil
	var exitCode int
	for err == nil {
		log.Println("Running fuzzing with: ./fuzzer " + strings.Join(args, " "))
		cmd := exec.Command("./fuzzer",
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

	err = c.transitionStatus(libFuzzerExitCodeToStatus(exitCode))
	if err != nil {
		return err
	}

	return nil
}

func (c *FuzzitClient) runLibFuzzerRegression() error {
	var corpusFiles []string
	var seedFiles []string

	corpusFiles, err := listFiles("corpus")
	if err != nil {
		return err
	}
	additionalFiles, err := listFiles("additional-corpus")
	if err != nil {
		return err
	}
	seedFiles, err = listFiles("seed")
	if err != nil {
		return err
	}

	regressionFiles := append(corpusFiles, seedFiles...)
	regressionFiles = append(regressionFiles, additionalFiles...)
	if len(regressionFiles) == 0 {
		log.Println("no files in corpus and seed. skipping run")
		c.transitionStatus("pass")
		return nil
	}

	args := append(
		[]string{
			"-print_final_stats=1",
			"-exact_artifact_path=./artifact",
			"-error_exitcode=76",
		},
		regressionFiles...,
	)
	if c.currentJob.Args != "" {
		args = append(args, splitAndRemoveEmpty(c.currentJob.Args, " ")...)
	}

	log.Println("Running regression...")
	cmd := exec.Command("./fuzzer",
		args...)
	if err := appendPrefixToCmd(cmd); err != nil {
		return err
	}

	exitCode := 0
	if err := cmd.Run(); err != nil {
		if !c.updateDB {
			// if this is local regression we want to exit with error code so the ci can fail
			return err
		}
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0

			// This works on both Unix and Windows. Although package
			// syscall is generally platform dependent, WaitStatus is
			// defined for both Unix and Windows and in both cases has
			// an ExitStatus() method with the same signature.
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
				log.Printf("Exit Status: %d", exitCode)
			}
		} else {
			return err
		}
	}

	if err := c.uploadCrash(exitCode); err != nil {
		return err
	}

	err = c.transitionStatus(libFuzzerExitCodeToStatus(exitCode))
	if err != nil {
		return err
	}

	return nil
}

func (c *FuzzitClient) runLibFuzzer() error {
	if _, err := os.Stat("fuzzer"); os.IsNotExist(err) {
		c.transitionStatus("failed")
		return fmt.Errorf("fuzzer executable doesnt exist")
	}
	if err := os.Chmod("./fuzzer", 0770); err != nil {
		return err
	}

	var err error
	if c.currentJob.Type == "regression" {
		err = c.runLibFuzzerRegression()
	} else {
		err = c.runLibFuzzerFuzzing()
	}

	return err
}
