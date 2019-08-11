package client

import (
	"cloud.google.com/go/firestore"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/user"
	"path"
	"time"
)

const FuzzitEndpoint = "https://app.fuzzit.dev"

type Target struct {
	Name string `firestore:"target_name"`
}

type Job struct {
	TargetId             string `firestore:"target_id"`
	Args                 string `firestore:"args"`
	Local                bool
	Type                 string   `firestore:"type"`
	Host                 string   `firestore:"host"`
	Revision             string   `firestore:"revision"`
	Branch               string   `firestore:"branch"`
	Parallelism          uint16   `firestore:"parallelism"`
	EnvironmentVariables []string `firestore:"environment_variables"`
}

// Internal struct
type job struct {
	Completed uint16    `firestore:"completed"`
	Status    string    `firestore:"status"`
	Namespace string    `firestore:"namespace"`
	StartedAt time.Time `firestore:"started_at,serverTimestamp"`
	OrgId     string    `firestore:"org_id"`
	V2        bool      `firestore:"v2"`
	Job
}

type FuzzitClient struct {
	Org             string
	Namespace       string
	ApiKey          string
	CustomToken     string
	Kind            string `json:"kind"`
	IdToken         string `json:"idToken"`
	RefreshToken    string `json:"refreshToken"`
	ExpiresIn       string `json:"expiresIn"`
	LastRefresh     int64
	firestoreClient *firestore.Client
	httpClient      *http.Client
}

func NewFuzzitClient(apiKey string) (*FuzzitClient, error) {
	c := &FuzzitClient{}
	c.httpClient = &http.Client{Timeout: 60 * time.Second}
	c.ApiKey = apiKey
	err := c.refreshToken()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func LoadFuzzitFromCache() (*FuzzitClient, error) {
	c := &FuzzitClient{}
	c.httpClient = &http.Client{Timeout: 60 * time.Second}

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	cacheFile := path.Join(usr.HomeDir, ".fuzzit.cache")

	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		return c, nil
	}

	file, err := os.Open(cacheFile)
	if err != nil {
		return nil, err
	}

	err = json.NewDecoder(file).Decode(c)
	file.Close()
	if err != nil {
		// try to prevent being stuck forever if cache file gets corrupted
		os.Remove(cacheFile)    // if a file
		os.RemoveAll(cacheFile) // if a directory
		return nil, err
	}

	//if c.ApiKey == "" {
	//	return errors.New("API Key is not configured (will have access only to public repositories)")
	//}

	return c, nil
}
