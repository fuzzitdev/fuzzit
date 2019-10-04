package client

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func (c *FuzzitClient) uploadGoFuzzCrash(path string) error {
	ctx := context.Background()

	if !c.updateDB {
		return nil
	}

	colRef := c.firestoreClient.Collection(fmt.Sprintf("orgs/%s/targets/%s/jobs/%s/crashes", c.Org, c.currentJob.TargetId, c.jobId))
	crashRef := colRef.NewDoc()

	log.Printf("uploading crash...")
	if err := c.uploadFile(path,
		fmt.Sprintf("orgs/%s/targets/%s/jobs/%s/crashes/%s", c.Org, c.currentJob.TargetId, c.jobId, crashRef.ID),
		fmt.Sprintf("crash-%s", crashRef.ID)); err != nil {
		return err
	}

	_, err := crashRef.Set(ctx, crash{
		TargetName: c.currentJob.TargetId,
		JobId:      c.jobId,
		TargetId:   c.currentJob.TargetId,
		OrgId:      c.Org,
		Type:       "crash",
		V2:         true,
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *FuzzitClient) loadCurrentCrashes() (map[string]bool, error) {
	uniqueCrashes := make(map[string]bool)

	err := filepath.Walk("workdir/crashers", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		fileName := info.Name()
		if !info.IsDir() && !strings.Contains(fileName, ".") && !uniqueCrashes[fileName] && fileName != "crashers" {
			uniqueCrashes[fileName] = true
		}
		return nil
	})

	numCrashers := len(uniqueCrashes)
	if numCrashers > 0 {
		log.Printf("resuming run with %d crashers\n", numCrashers)
	}

	return uniqueCrashes, err
}

// this merges corpus into workdir
func (c *FuzzitClient) mergeCorpus() error {
	err := filepath.Walk("corpus", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if !info.IsDir() {
			fileName := info.Name()
			err = os.Rename("corpus/"+fileName, "workdir/corpus/"+fileName)
			if err != nil {
				return err
			}
		}
		return nil
	})

	return err
}

func (c *FuzzitClient) runGoFuzzFuzzing() error {
	ctx := context.Background()

	args := append(
		[]string{
			"-workdir=workdir",
			"-procs=1",
			"-bin=fuzzer.zip",
		},
	)

	var err error
	if runtime.GOOS == "linux" {
		err = DownloadFile("go-fuzz", "https://storage.googleapis.com/public-fuzzit/go-fuzz-linux")
	} else if runtime.GOOS == "darwin" {
		err = DownloadFile("go-fuzz", "https://storage.googleapis.com/public-fuzzit/go-fuzz-osx")
	} else {
		return fmt.Errorf("fuzzit with go-fuzz currently only supports linux or darwin")
	}
	if err != nil {
		return err
	}
	if err := os.Chmod("./go-fuzz", 0770); err != nil {
		return err
	}

	log.Println("downloading previous go-fuzz workdir")
	workdirPath := fmt.Sprintf("orgs/%s/targets/%s/jobs/%s/workdir.tar.gz", c.Org, c.currentJob.TargetId, c.jobId)
	err = c.downloadAndExtract(".", workdirPath)
	if err != nil {
		if err.Error() == "404 Not Found" {
			log.Println("no generating corpus yet. continue...")
		} else {
			return err
		}
	}
	if _, err := os.Stat("workdir"); os.IsNotExist(err) {
		if err := os.Mkdir("workdir", 0644); err != nil {
			return err
		}
	}

	if _, err := os.Stat("workdir/corpus"); os.IsNotExist(err) {
		if err := os.Mkdir("workdir/corpus", 0644); err != nil {
			return err
		}
	}

	if _, err := os.Stat("workdir/crashers"); os.IsNotExist(err) {
		if err := os.Mkdir("workdir/crashers", 0644); err != nil {
			return err
		}
	}

	err = c.mergeCorpus()
	if err != nil {
		log.Fatal(err)
	}

	uniqueCrashes, err := c.loadCurrentCrashes()
	if err != nil {
		return err
	}

	log.Println("Running: go-fuzz " + strings.Join(args, " "))
	cmd := exec.Command("./go-fuzz",
		args...)
	if err := appendPrefixToCmd(cmd); err != nil {
		return err
	}
	err = cmd.Start()

	lastUpload := time.Now()

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

			err = filepath.Walk("workdir/crashers", func(path string, info os.FileInfo, err error) error {
				if err != nil {
					log.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
					return err
				}
				fileName := info.Name()
				if !strings.Contains(fileName, ".") && !uniqueCrashes[fileName] && fileName != "crashers" {
					uniqueCrashes[fileName] = true
					err := c.uploadGoFuzzCrash("workdir/crashers/" + fileName)
					if err != nil {
						return err
					}
				}
				return nil
			})
			if err != nil {
				cmd.Process.Kill()
				return err
			}

			now := time.Now()
			if now.Sub(lastUpload).Seconds() > 3600 {
				log.Println("uploading workdir...")
				if err := c.archiveAndUpload("workdir",
					fmt.Sprintf("orgs/%s/targets/%s/jobs/%s/workdir.tar.gz", c.Org, c.currentJob.TargetId, c.jobId),
					"workdir.tar.gz"); err != nil {
					return err
				}

				log.Println("uploading corpus...")
				if err := c.archiveAndUpload("workdir/corpus",
					fmt.Sprintf("orgs/%s/targets/%s/corpus.tar.gz", c.Org, c.currentJob.TargetId),
					"corpus.tar.gz"); err != nil {
					return err
				}

				lastUpload = time.Now()
			}

		case err = <-done:
			stopSession = true
			if err != nil {
				log.Printf("process finished with error = %v\n", err)
				return err
			} else {
				log.Print("process finished successfully")
			}
		}
	}

	if err := c.refreshToken(); err != nil {
		return err
	}

	err = c.transitionStatus("pass")
	if err != nil {
		return err
	}

	return nil

}

func (c *FuzzitClient) runGoFuzz() error {
	var err error

	if c.currentJob.Type == "fuzzing" {
		err = c.runGoFuzzFuzzing()
	} else {
		return fmt.Errorf("JQF currently only supports fuzzing")
	}

	return err
}
