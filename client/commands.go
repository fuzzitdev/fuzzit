package client

import (
	"archive/tar"
	"cloud.google.com/go/firestore"
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/google/uuid"
	"github.com/mholt/archiver"
	"google.golang.org/api/iterator"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	//"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	//"github.com/docker/docker/pkg/stdcopy"
)

func (c *fuzzitClient) archiveFiles(files []string) (string, error) {
	fuzzerPath := files[0]
	filename := filepath.Base(fuzzerPath)
	if !strings.HasSuffix(filename, ".tar.gz") {
		tmpDir, err := ioutil.TempDir("", "fuzzit")
		if err != nil {
			return "", err
		}
		dstPath := filepath.Join(tmpDir, "fuzzer")
		_, err = copyFile(dstPath, fuzzerPath)
		if err != nil {
			return "", err
		}

		prefix, err := uuid.NewRandom()
		if err != nil {
			return "", err
		}
		filesToArchive := append([]string{dstPath}, files[1:]...)

		tmpfile := filepath.Join(os.TempDir(), prefix.String()+".tar.gz")
		z := archiver.NewTarGz()
		err = z.Archive(filesToArchive, tmpfile)
		if err != nil {
			return "", err
		}
		fuzzerPath = tmpfile
	}

	return fuzzerPath, nil
}

func (c *fuzzitClient) DownloadSeed(dst string, target string) error {
	storagePath := fmt.Sprintf("orgs/%s/targets/%s/seed", c.Org, target)
	err := c.downloadFile(dst, storagePath)
	if err != nil {
		return err
	}
	return nil
}

func (c *fuzzitClient) DownloadCorpus(dst string, target string) error {
	storagePath := fmt.Sprintf("orgs/%s/targets/%s/corpus", c.Org, target)
	err := c.downloadFile(dst, storagePath)
	if err != nil {
		return err
	}
	return nil
}

func (c *fuzzitClient) DownloadFuzzer(dst string, target string, job string) error {
	storagePath := fmt.Sprintf("orgs/%s/targets/%s/jobs/%s/fuzzer", c.Org, target, job)
	err := c.downloadFile(dst, storagePath)
	if err != nil {
		return err
	}
	return nil
}

func (c *fuzzitClient) GetResource(resource string) error {
	err := c.refreshToken()
	if err != nil {
		return err
	}

	ctx := context.Background()
	rootColRef := "orgs/" + c.Org + "/"
	r := rootColRef + resource
	if (len(strings.Split(resource, "/")) % 2) == 0 {
		docRef := c.firestoreClient.Doc(r)
		if docRef == nil {
			return fmt.Errorf("invalid resource %s", r)
		}
		docsnap, err := docRef.Get(ctx)
		if err != nil {
			return err
		}
		if !docsnap.Exists() {
			return fmt.Errorf("resource %s doesn't exist", resource)
		}

		jsonString, err := json.MarshalIndent(docsnap.Data(), "", " ")
		if err != nil {
			return err
		}
		fmt.Println(string(jsonString))
		return nil
	} else {
		iter := c.firestoreClient.Collection(r).Documents(ctx)
		querySize := 0
		defer iter.Stop()

		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return err
			}
			data := doc.Data()
			data["id"] = doc.Ref.ID
			jsonString, err := json.MarshalIndent(data, "", " ")
			if err != nil {
				return err
			}
			fmt.Println(string(jsonString))
			querySize += 1
		}
		if querySize == 0 {
			return fmt.Errorf("no resources for %s", resource)
		}
		return nil
	}
}

func (c *fuzzitClient) CreateTarget(targetName string, seedPath string) (*firestore.DocumentRef, error) {
	err := c.refreshToken()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	docRef := c.firestoreClient.Doc("orgs/" + c.Org + "/targets/" + targetName)
	_, err = docRef.Get(ctx)
	if err == nil {
		return nil, fmt.Errorf("target %s already exist", targetName)
	}

	if seedPath != "" {
		storagePath := fmt.Sprintf("orgs/%s/targets/%s/seed", c.Org, targetName)
		err := c.uploadFile(seedPath, storagePath, "application/gzip", "seed.tar.gz")
		if err != nil {
			return nil, err
		}
	}

	_, err = docRef.Set(ctx, Target{Name: targetName})
	if err != nil {
		return nil, err
	}

	return docRef, nil
}

func (c *fuzzitClient) getRunShTar() (*os.File, error) {
	tmpfile, err := ioutil.TempFile("", "run.*.tar")
	if err != nil {
		log.Fatal(err)
	}
	tw := tar.NewWriter(tmpfile)
	hdr := &tar.Header{
		Name: "run.sh",
		Mode: 0777,
		Size: int64(len(runSh)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return nil, err
	}
	if _, err := tw.Write([]byte(runSh)); err != nil {
		return nil, err
	}
	if err := tw.Flush(); err != nil {
		return nil, err
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := tmpfile.Close(); err != nil {
		return nil, err
	}

	runShTar, err := os.Open(tmpfile.Name())
	if err != nil {
		return nil, err
	}

	return runShTar, nil
}

func (c *fuzzitClient) CreateLocalJob(jobConfig Job, files []string) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}
	cli.NegotiateAPIVersion(ctx)

	fuzzerPath, err := c.archiveFiles(files)
	if err != nil {
		return err
	}

	fuzzer, err := os.Open(fuzzerPath)
	if err != nil {
		return err
	}

	corpusPath := fmt.Sprintf("orgs/%s/targets/%s/corpus.tar.gz", c.Org, jobConfig.TargetId)
	log.Print(corpusPath)
	corpusLink, err := c.getStorageLink(corpusPath, "read")
	if err != nil {
		return err
	}

	seedPath := fmt.Sprintf("orgs/%s/targets/%s/seed", c.Org, jobConfig.TargetId)
	seedLink, err := c.getStorageLink(seedPath, "read")
	if err != nil {
		return err
	}

	log.Println("Pulling container")
	reader, err := cli.ImagePull(ctx, "docker.io/fuzzitdev/fuzzit:stretch-llvm8", types.ImagePullOptions{})
	if err != nil {
		return err
	}
	_, err = io.Copy(os.Stdout, reader)
	if err != nil {
		return err
	}

	log.Println("Creating container")
	createdContainer, err := cli.ContainerCreate(ctx, &container.Config{
		Env: []string{
			"CORPUS_LINK=" + corpusLink,
			"SEED_LINK=" + seedLink,
			"ASAN_OPTIONS=" + jobConfig.AsanOptions,
			"UBSAN_OPTIONS=" + jobConfig.UbsanOptions,
			"ARGS=" + jobConfig.Args,
			"LD_LIBRARY_PATH=/app"},
		Image: "docker.io/fuzzitdev/fuzzit:stretch-llvm8",
		//Cmd:         []string{"/bin/sh", "/app/run.sh"},
		Cmd:         []string{"/bin/sleep", "100000"},
		AttachStdin: true,
	}, nil, nil, "")
	if err != nil {
		return err
	}

	log.Println("Uploading fuzzer to container")
	err = cli.CopyToContainer(ctx, createdContainer.ID, "/app", fuzzer, types.CopyToContainerOptions{
		AllowOverwriteDirWithFile: true,
	})
	if err != nil {
		return err
	}

	runShTar, err := c.getRunShTar()
	if err != nil {
		return err
	}
	log.Println("Uploading run.sh to container")
	err = cli.CopyToContainer(ctx, createdContainer.ID, "/app/", runShTar, types.CopyToContainerOptions{
		AllowOverwriteDirWithFile: true,
	})
	if err != nil {
		return err
	}

	log.Println("Starting the container")
	err = cli.ContainerStart(ctx, createdContainer.ID, types.ContainerStartOptions{})
	if err != nil {
		return err
	}

	out, err := cli.ContainerLogs(ctx, createdContainer.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		return err
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)

	log.Println("Waiting for container")
	statusCh, errCh := cli.ContainerWait(ctx, createdContainer.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			cli.ContainerRemove(ctx, createdContainer.ID, types.ContainerRemoveOptions{})
			return err
		}
	case status := <-statusCh:
		if status.StatusCode != 0 {
			cli.ContainerRemove(ctx, createdContainer.ID, types.ContainerRemoveOptions{})
			return fmt.Errorf("fuzzer exited with %d", status.StatusCode)
		}
	}

	err = cli.ContainerRemove(ctx, createdContainer.ID, types.ContainerRemoveOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (c *fuzzitClient) CreateJob(jobConfig Job, files []string) (*firestore.DocumentRef, error) {
	err := c.refreshToken()
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	collectionRef := c.firestoreClient.Collection("orgs/" + c.Org + "/targets/" + jobConfig.TargetId + "/jobs")
	fullJob := job{}
	fullJob.Job = jobConfig
	fullJob.Completed = 0
	fullJob.OrgId = c.Org
	fullJob.Namespace = c.Namespace
	fullJob.Status = "in progress"
	fullJob.V2 = true

	jobRef := collectionRef.NewDoc()

	fuzzerPath := files[0]
	filename := filepath.Base(fuzzerPath)
	if !strings.HasSuffix(filename, ".tar.gz") {
		tmpDir, err := ioutil.TempDir("", "fuzzit")
		if err != nil {
			return nil, err
		}
		dstPath := filepath.Join(tmpDir, "fuzzer")
		_, err = copyFile(dstPath, fuzzerPath)
		if err != nil {
			return nil, err
		}

		prefix, err := uuid.NewRandom()
		if err != nil {
			return nil, err
		}
		filesToArchive := append([]string{dstPath}, files[1:]...)

		tmpfile := filepath.Join(os.TempDir(), prefix.String()+".tar.gz")
		z := archiver.NewTarGz()
		err = z.Archive(filesToArchive, tmpfile)
		if err != nil {
			return nil, err
		}
		fuzzerPath = tmpfile
	}

	storagePath := fmt.Sprintf("orgs/%s/targets/%s/jobs/%s/fuzzer", c.Org, jobConfig.TargetId, jobRef.ID)
	log.Println("Uploading fuzzer...")
	err = c.uploadFile(fuzzerPath, storagePath, "application/gzip", "fuzzer.tar.gz")
	if err != nil {
		return nil, err
	}

	log.Println("Starting job")
	_, err = jobRef.Set(ctx, fullJob)
	if err != nil {
		log.Printf("Please check that the target '%s' exists and you have sufficiant permissions",
			jobConfig.TargetId)
		return nil, err
	}

	log.Printf("Job %s started succesfully\n", jobRef.ID)
	return jobRef, nil
}
