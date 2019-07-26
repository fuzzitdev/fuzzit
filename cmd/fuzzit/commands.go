package main

import (
	"bytes"
	"cloud.google.com/go/firestore"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/mholt/archiver"
	"golang.org/x/oauth2"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Target struct {
	TargetName string `firestore:"target_name"`
}

type Job struct {
	Completed uint16 	`firestore:"completed"`
	Status string 		`firestore:"status"`
	Args string  		`firestore:"args"`
	Type string 		`firestore:"type"`
	Host string 		`firestore:"host"`
	Revision string		`firestore:"revision"`
	Branch string		`firestore:"branch"`
	Parallelism uint16  `firestore:"parallelism"`
	Namespace string 	`firestore:"namespace"`
	AsanOptions string  `firestore:"asan_options"`
	UbsanOptions string  `firestore:"ubsan_options"`
	StartedAt time.Time `firestore:"started_at,serverTimestamp"`
	TargetId string `firestore:"target_id"`
	OrgId string `firestore:"org_id"`
}

type FuzzitCli struct {
	Org string
	Namespace string
	ApiKey string
	CustomToken string
	Kind string `json:"kind"`
	IdToken string `json:"idToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn string `json:"expiresIn"`
	LastRefresh int64
	firestoreClient *firestore.Client `json:"-"`
	httpClient *http.Client `json:"-"`
}

func copyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func loadFromFile(configFile string) (*FuzzitCli, error) {
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("please run fuzzit auth <api_key>")
	}

	file, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	c := &FuzzitCli{}
	err = json.NewDecoder(file).Decode(c)
	if err != nil {
		return nil, err
	}

	c.httpClient = &http.Client{Timeout: 60 * time.Second}

	if time.Now().Unix() - c.LastRefresh < 300 || c.LastRefresh == 0 {
		r, err := c.httpClient.Get("https://app.fuzzit.dev/createCustomToken?api_key=" + c.ApiKey)
		if err != nil {
			return nil, err
		}
		defer r.Body.Close()
		if r.StatusCode != 200 {
			return nil, fmt.Errorf("API Key is not valid")
		}
		err = json.NewDecoder(r.Body).Decode(c)
		if err != nil {
			return nil, err
		}

		r, err = c.httpClient.Post("https://www.googleapis.com/identitytoolkit/v3/relyingparty/verifyCustomToken?key=AIzaSyCs_Sm1VOKZwJZmTXdOCvs1wyn91vYMNSY",
					   	   "application/json",
					   			     bytes.NewBuffer([]byte(fmt.Sprintf(`{"token": "%s", "returnSecureToken": true}`, c.CustomToken))))
		if err != nil {
			return nil, err
		}
		defer r.Body.Close()
		err = json.NewDecoder(r.Body).Decode(c)
		if err != nil {
			return nil, err
		}
		c.LastRefresh = time.Now().Unix()

		cBytes, err := json.MarshalIndent(c, "", "")
		if err != nil {
			return nil, err
		}
		err = ioutil.WriteFile(configFile, cBytes, 0644)
		if err != nil {
			return nil, err
		}
	}

	token := oauth2.Token{
		AccessToken: c.IdToken,
		RefreshToken: c.RefreshToken,
		Expiry: time.Time{},
		TokenType: "Bearer",
	}

	tokenSource := oauth2.StaticTokenSource(&token)
	ctx := context.Background()
	c.firestoreClient, err = firestore.NewClient(ctx, "fuzzit-b5fbf", option.WithTokenSource(tokenSource))

	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c * FuzzitCli) getResource(resource string) error{
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

func (c * FuzzitCli) createTarget(name string, seedPath string) (*firestore.DocumentRef, error) {
	ctx := context.Background()
	collectionRef := c.firestoreClient.Collection("orgs/" + c.Org + "/targets")
	doc, _, err := collectionRef.Add(ctx,
		Target{
			TargetName: name,
		})
	if err != nil {
		return nil, err
	}

	if seedPath != "" {
		storagePath := fmt.Sprintf("orgs/%s/targets/%s/seed", c.Org, doc.ID)
		err := uploadFile(seedPath, storagePath, c.ApiKey, "application/gzip", "seed.tar.gz")
		if err != nil {
			return nil, err
		}
	}
	return doc, nil
}

func (c * FuzzitCli) createJob(targetId string, jobType string, host string,
	args string, asan_options string, ubsan_options string,
	revision string, branch string, cpus uint16, files [] string) (*firestore.DocumentRef, error) {
	ctx := context.Background()
	collectionRef := c.firestoreClient.Collection("orgs/" + c.Org + "/targets/" + targetId + "/jobs")
	doc, _, err := collectionRef.Add(ctx,
		Job{
			Completed: 0,
			Status: "in progress",
			Host: host,
			Args: args,
			AsanOptions: asan_options,
			UbsanOptions: ubsan_options,
			Type: jobType,
			Revision: revision,
			Branch: branch,
			Parallelism: cpus,
			OrgId: c.Org,
			TargetId: targetId,
			Namespace: c.Namespace,
		})
	if err != nil {
		return nil, err
	}
	fmt.Println("Created new job ", doc.ID)

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

	storagePath := fmt.Sprintf("orgs/%s/targets/%s/jobs/%s/fuzzer", c.Org, targetId, doc.ID)
	err = uploadFile(fuzzerPath, storagePath, c.ApiKey, "application/gzip", "fuzzer.tar.gz")
	if err != nil {
		return nil, err
	}

	jsonStr := []byte(fmt.Sprintf(`{"data": {"org_id": "%s", "target_id": "%s", "job_id": "%s"}}`, c.Org, targetId, doc.ID))
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
