package client

import (
	"bytes"
	"cloud.google.com/go/firestore"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/mholt/archiver"
	"google.golang.org/api/iterator"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

func (c * fuzzitClient) GetResource(resource string) error {
	err := c.ReAuthenticate(false)
	if err != nil {
		return err
	}

	ctx := context.Background()
	rootColRef := "orgs/" + c.Org + "/"
	if (len(strings.Split(resource, "/")) % 2) == 0 {
		r := rootColRef + resource
		docRef := c.firestoreClient.Doc(rootColRef + resource)
		if docRef == nil {
			return fmt.Errorf("invalid resource %s", r)
		}
		docsnap, err := docRef.Get(ctx)
		if !docsnap.Exists() {
			return fmt.Errorf("resource %s doesn't exist", resource)
		}
		if err != nil {
			return err
		}

		jsonString, err := json.MarshalIndent(docsnap.Data(), "", " ")
		if err != nil {
			return err
		}
		fmt.Println(string(jsonString))
		return nil
	} else {
		iter := c.firestoreClient.Collection(rootColRef + resource).Documents(ctx)
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

func (c * fuzzitClient) CreateTarget(targetConfig Target, seedPath string) (*firestore.DocumentRef, error) {
	ctx := context.Background()
	collectionRef := c.firestoreClient.Collection("orgs/" + c.Org + "/targets")
	doc, _, err := collectionRef.Add(ctx,
		targetConfig)
	if err != nil {
		return nil, err
	}

	if seedPath != "" {
		storagePath := fmt.Sprintf("orgs/%s/targets/%s/seed", c.Org, doc.ID)
		err := c.uploadFile(seedPath, storagePath, "application/gzip", "seed.tar.gz")
		if err != nil {
			return nil, err
		}
	}
	return doc, nil
}

func (c * fuzzitClient) CreateJob(jobConfig Job, files [] string) (*firestore.DocumentRef, error) {
	ctx := context.Background()
	collectionRef := c.firestoreClient.Collection("orgs/" + c.Org + "/targets/" + jobConfig.TargetId + "/jobs")
	fullJob := job{}
	fullJob.Job = jobConfig
	fullJob.Completed = 0
	fullJob.OrgId = c.Org
	fullJob.Namespace = c.Namespace
	fullJob.Status = "in progress"
	doc, _, err := collectionRef.Add(ctx,
		fullJob)
	if err != nil {
		return nil, err
	}
	log.Println("Created new job ", doc.ID)

	fuzzerPath := files[0]
	splits := strings.Split(fuzzerPath, "/")
	filename := splits[len(splits) - 1]
	if !strings.HasSuffix(filename, ".tar.gz") {
		tmpDir, err := ioutil.TempDir("", "fuzzit")
		if err != nil {
			return nil, err
		}
		_, err = copyFile(fuzzerPath, tmpDir+"/fuzzer")
		if err != nil {
			return nil, err
		}

		prefix, err := uuid.NewRandom()
		if err != nil {
			return nil, err
		}
		filesToArchive := append([]string{tmpDir + "/fuzzer"}, files[1:]...)

		tmpfile := os.TempDir() + "/" + prefix.String() + ".tar.gz"
		z := archiver.NewTarGz()
		err = z.Archive(filesToArchive, tmpfile)
		if err != nil {
			return nil, err
		}
		fuzzerPath = tmpfile
	}

	storagePath := fmt.Sprintf("orgs/%s/targets/%s/jobs/%s/fuzzer", c.Org, jobConfig.TargetId, doc.ID)
	err = c.uploadFile(fuzzerPath, storagePath, "application/gzip", "fuzzer.tar.gz")
	if err != nil {
		return nil, err
	}

	jsonStr := []byte(fmt.Sprintf(`{"data": {"org_id": "%s", "target_id": "%s", "job_id": "%s"}}`, c.Org, jobConfig.TargetId, doc.ID))
	req, err := http.NewRequest("POST",
		"https://us-central1-fuzzit-b5fbf.cloudfunctions.net/startJob",
		bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer " + c.IdToken)
	req.Header.Set("Content-Type", "application/json")

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyBytes)
		defer res.Body.Close()
		return nil, fmt.Errorf(bodyString)
	}
	fmt.Printf("Job %s started succesfully\n", doc.ID)
	defer res.Body.Close()
	return doc, nil
}

