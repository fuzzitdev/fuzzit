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
)

const (
	libFuzzerTimeoutExitCode = 77
	libFuzzerLeakExitCode    = 76
	libFuzzerCrashExitCode   = 1
	libFuzzerOOMExitCode     = -9
	libFuzzerSuccessExitCode = 0

	jqfCrashExitCode   = 3
	jqfSuccessExitCode = 0

	fuzzingInterval = 3600

	AgentGeneralError      = 1
	AgentNoPermissionError = 22
)

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

func (c *FuzzitClient) transitionToInProgress() error {
	ctx := context.Background()
	job := Job{}
	if c.updateDB {
		// transaction doesnt work for now at go client with oauth token
		jobRef := c.firestoreClient.Doc(fmt.Sprintf("orgs/%s/targets/%s/jobs/%s", c.Org, c.currentJob.TargetId, c.jobId))
		docsnap, err := jobRef.Get(ctx)
		if err != nil {
			return err
		}
		err = docsnap.DataTo(&job)
		if err != nil {
			return err
		}
		if job.Status == "queued" {
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
	jobRef := c.firestoreClient.Doc(fmt.Sprintf("orgs/%s/targets/%s/jobs/%s", c.Org, c.currentJob.TargetId, c.jobId))
	_, err := jobRef.Update(ctx, []firestore.Update{{Path: "status", Value: status}})
	if err != nil {
		return err
	}

	return nil
}

func (c *FuzzitClient) RunFuzzer(job Job, jobId string, updateDB bool) error {
	if err := c.refreshToken(); err != nil {
		return err
	}

	c.currentJob = job
	c.jobId = jobId
	c.updateDB = updateDB

	if err := c.transitionToInProgress(); err != nil {
		return err
	}

	if err := os.Mkdir("seed", 0644); err != nil {
		return err
	}

	log.Println("downloading seed")
	if err := c.DownloadAndExtractSeed("./seed", c.currentJob.TargetId); err != nil {
		if err.Error() == "404 Not Found" {
			log.Println("no seed corpus. continue...")
		} else {
			return err
		}
	}

	log.Println("downloading corpus")
	if err := c.DownloadAndExtractCorpus(".", c.currentJob.TargetId); err != nil {
		if err.Error() == "404 Not Found" {
			log.Println("no generating corpus yet. continue...")
		} else {
			return err
		}
	}

	if err := createDirIfNotExist("corpus"); err != nil {
		return err
	}

	if jobId != "" {
		log.Println("downloading fuzzer")
		if err := c.DownloadAndExtractFuzzer(".", c.currentJob.TargetId, jobId); err != nil {
			return err
		}

		log.Println("downloading additional corpus")
		if err := c.downloadAndExtract(
			"additional-corpus",
			fmt.Sprintf("orgs/%s/targets/%s/jobs/%s/additional-corpus", c.Org, c.currentJob.TargetId, c.jobId)); err != nil {
			if err.Error() == "404 Not Found" {
				log.Println("no additional-corpus. skipping...")
			} else {
				return err
			}
		}

		if err := createDirIfNotExist("additional-corpus"); err != nil {
			return err
		}

	}

	var err error
	if c.currentJob.Engine == "jqf" {
		err = c.RunJQF()
	} else if c.currentJob.Engine == "go-fuzz" {
		err = c.runGoFuzz()
	} else {
		err = c.runLibFuzzer()
	}

	return err
}
