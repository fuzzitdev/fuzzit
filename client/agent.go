package client

import (
	"bufio"
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

const (
	libFuzzerTimeoutExitCode = 77
	libFuzzerLeakExitCode    = 76
	libFuzzerCrashExitCode   = 1
	libFuzzerOOMExitCode     = -9
	libFuzzerSuccessExitCode = 0

	fuzzingInterval = 3600
)

var libFuzzerArgs = []string{
	"-print_final_stats=1",
	"-exact_artifact_path=./artifact",
}

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

func appendPrefixToCmd(cmd *exec.Cmd) error {
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	go func() {
		merged := io.MultiReader(stderr, stdout)
		scanner := bufio.NewScanner(merged)
		for scanner.Scan() {
			msg := scanner.Text()
			fmt.Printf("FUZZER: %s\n", msg)
		}
	}()

	return nil
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

	if err := os.Mkdir("merge", 0644); err != nil {
		return err
	}
	if _, err := os.Stat("/tmp/merge_control.txt"); err == nil {
		if err = os.Remove("/tmp/merge_control.txt"); err != nil {
			return err
		}
	}
	args := append([]string{
		"-print_final_stats=1",
		"-exact_artifact_path=./artifact",
		fmt.Sprintf("-error_exitcode=%d", libFuzzerTimeoutExitCode),
		"-merge_control_file=/tmp/merge_control.txt",
		"-merge=1",
		"merge",
		"corpus",
	})

	log.Println("Running merge with: ./fuzzer " + strings.Join(args, " "))
	cmd := exec.Command("./fuzzer",
		args...)
	if err := appendPrefixToCmd(cmd); err != nil {
		return err
	}

	if err := cmd.Run(); err != nil {
		return err
	}

	c.refreshToken()
	if err := c.archiveAndUpload("merge",
		fmt.Sprintf("orgs/%s/targets/%s/corpus.tar.gz", c.Org, c.targetId),
		"corpus.tar.gz"); err != nil {
		return err
	}

	if err := os.RemoveAll("corpus"); err != nil {
		return err
	}

	if err := os.Rename("merge", "corpus"); err != nil {
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
		log.Printf("uploading crash...")
		if err = c.uploadFile("artifact",
			fmt.Sprintf("orgs/%s/targets/%s/crashes/%s-%s", c.Org, c.targetId, c.jobId, os.Getenv("POD_ID")),
			"crash"); err != nil {
			return err
		}
		colRef := c.firestoreClient.Collection(fmt.Sprintf("orgs/%s/targets/%s/crashes", c.Org, c.targetId))
		_, _, err = colRef.Add(ctx, crash{
			TargetName: c.targetId,
			PodId:      os.Getenv("POD_ID"),
			JobId:      c.jobId,
			TargetId:   c.targetId,
			OrgId:      c.Org,
			ExitCode:   uint32(exitCode),
			Type:       "CRASH",
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *FuzzitClient) runLibFuzzerFuzzing() error {
	ctx := context.Background()

	args := append([]string{
		"-print_final_stats=1",
		"-exact_artifact_path=./artifact",
		"-error_exitcode=76",
		"-max_total_time=3600",
		"corpus",
		"seed",
	})

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
		for stopSession == false {
			select {
			case <-timeout:
				var fuzzingJob job
				c.refreshToken()
				docRef := c.firestoreClient.Doc(fmt.Sprintf("orgs/%s/targets/%s/jobs/%s", c.Org, c.targetId, c.jobId))
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

	c.refreshToken()
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
	seedFiles, err = listFiles("seed")
	if err != nil {
		return err
	}

	regressionFiles := append(corpusFiles, seedFiles...)
	if len(regressionFiles) == 0 {
		log.Println("no files in corpus and seed. skipping run")
		c.transitionStatus("pass")
		return nil
	}

	args := append([]string{
		"-print_final_stats=1",
		"-exact_artifact_path=./artifact",
		"-error_exitcode=76",
	}, regressionFiles...)
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

func (c *FuzzitClient) transitionToInProgress() error {
	ctx := context.Background()

	if c.updateDB {
		// transaction doesnt work for now at go client with oauth token
		jobRef := c.firestoreClient.Doc(fmt.Sprintf("orgs/%s/targets/%s/jobs/%s", c.Org, c.targetId, c.jobId))
		docsnap, err := jobRef.Get(ctx)
		if err != nil {
			return err
		}
		err = docsnap.DataTo(&c.currentJob)
		if err != nil {
			return err
		}
		if c.currentJob.Status == "queued" {
			_, err := jobRef.Update(ctx, []firestore.Update{{Path: "status", Value: "in progress"}})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *FuzzitClient) transitionStatus(status string) error {
	ctx := context.Background()

	if !c.updateDB {
		return nil
	}

	// transaction doesnt work for now at go client with oauth token
	jobRef := c.firestoreClient.Doc(fmt.Sprintf("orgs/%s/targets/%s/jobs/%s", c.Org, c.targetId, c.jobId))
	_, err := jobRef.Update(ctx, []firestore.Update{{Path: "status", Value: status}})
	if err != nil {
		return err
	}

	return nil
}

func (c *FuzzitClient) RunLibFuzzer(targetId string, jobId string, updateDB bool, fuzzingType string) error {
	err := c.refreshToken()
	c.targetId = targetId
	c.jobId = jobId
	c.updateDB = updateDB

	if err = c.transitionToInProgress(); err != nil {
		return err
	}

	if err = os.Mkdir("corpus", 0644); err != nil {
		return err
	}
	if err = os.Mkdir("seed", 0644); err != nil {
		return err
	}

	if jobId != "" {
		log.Println("downloading fuzzer")
		if err := c.DownloadAndExtractFuzzer(".", targetId, jobId); err != nil {
			return err
		}
	}

	if _, err := os.Stat("fuzzer"); os.IsNotExist(err) {
		c.transitionStatus("failed")
		return fmt.Errorf("fuzzer executable doesnt exist")
	}

	if err := os.Chmod("./fuzzer", 0770); err != nil {
		return err
	}

	log.Println("downloading corpus")
	if err := c.DownloadAndExtractCorpus("./corpus", targetId); err != nil {
		if err.Error() == "404 Not Found" {
			log.Println("no generating corpus yet. continue...")
		} else {
			return err
		}
	}

	log.Println("downloading seed")
	if err := c.DownloadAndExtractSeed("./seed", targetId); err != nil {
		if err.Error() == "404 Not Found" {
			log.Println("no seed corpus. continue...")
		} else {
			return err
		}
	}

	if fuzzingType == "regression" {
		err = c.runLibFuzzerRegression()
	} else {
		err = c.runLibFuzzerFuzzing()
	}

	if err != nil {
		return err
	}

	return nil
}
