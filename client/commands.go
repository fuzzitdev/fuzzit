package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"github.com/mholt/archiver"
	"google.golang.org/api/iterator"
)

func (c *fuzzitClient) GetResource(resource string) error {
	err := c.ReAuthenticate(false)
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
	ctx := context.Background()
	docRef := c.firestoreClient.Doc("orgs/" + c.Org + "/targets/" + targetName)
	_, err := docRef.Get(ctx)
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

func (c *fuzzitClient) CreateJob(jobConfig Job, files []string) (*firestore.DocumentRef, error) {
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
	err := c.uploadFile(fuzzerPath, storagePath, "application/gzip", "fuzzer.tar.gz")
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
