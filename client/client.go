package client

import (
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
)

const FuzzitEndpoint = "https://app.fuzzit.dev"
const Version = "v2.4.51"

type Target struct {
	Name         string `firestore:"target_name"`
	PublicCorpus bool   `firestore:"public_corpus"`
}

type Job struct {
	TargetId             string    `firestore:"target_id"`
	Args                 string    `firestore:"args"`
	Type                 string    `firestore:"type"`
	Engine               string    `firestore:"engine"`
	Host                 string    `firestore:"host"`
	Revision             string    `firestore:"revision"`
	Branch               string    `firestore:"branch"`
	Parallelism          uint16    `firestore:"parallelism"`
	EnvironmentVariables []string  `firestore:"environment_variables"`
	Completed            uint16    `firestore:"completed"`
	Status               string    `firestore:"status"`
	Namespace            string    `firestore:"namespace"`
	StartedAt            time.Time `firestore:"started_at,serverTimestamp"`
	OrgId                string    `firestore:"org_id"`
}

type crash struct {
	TargetName string    `firestore:"target_name"`
	PodId      string    `firestore:"pod_id"`
	JobId      string    `firestore:"job_id"`
	TargetId   string    `firestore:"target_id"`
	OrgId      string    `firestore:"org_id"`
	ExitCode   uint32    `firestore:"exit_code"`
	Type       string    `firestore:"type"`
	Time       time.Time `firestore:"time,serverTimestamp"`
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
	currentJob      Job    // this is mainly used by the agent
	jobId           string // this is mainly used by the agent
	updateDB        bool   // this is mainly used by the agent
	fuzzerFilename  string // this is mainly used by the agent
}

func NewUnAuthenticatedClient() *FuzzitClient {
	c := &FuzzitClient{}
	c.httpClient = &http.Client{Timeout: 120 * time.Second}
	return c
}

func NewFuzzitClient(apiKey string) (*FuzzitClient, error) {
	c := &FuzzitClient{}
	c.httpClient = &http.Client{Timeout: 120 * time.Second}
	c.ApiKey = apiKey
	err := c.refreshToken()
	if err != nil {
		return nil, err
	}
	return c, nil
}
