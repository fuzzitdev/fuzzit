package client

import (
	"cloud.google.com/go/firestore"
	"net/http"
	"time"
)

const FuzzitEndpoint = "https://app.fuzzit.dev"
const CacheFile = "/tmp/.fuzzit.cache"


type Target struct {
	Name string `firestore:"target_name"`
}

type Job struct {
	TargetId string  	`firestore:"target_id"`
	Args string  		`firestore:"args"`
	Type string 		`firestore:"type"`
	Host string 		`firestore:"host"`
	Revision string		`firestore:"revision"`
	Branch string		`firestore:"branch"`
	Parallelism uint16  `firestore:"parallelism"`
	AsanOptions string  `firestore:"asan_options"`
	UbsanOptions string `firestore:"ubsan_options"`
}


// Internal struct
type job struct {
	Completed uint16 	`firestore:"completed"`
	Status string 		`firestore:"status"`
	Namespace string 	`firestore:"namespace"`
	StartedAt time.Time `firestore:"started_at,serverTimestamp"`
	OrgId string 		`firestore:"org_id"`
	Job
}


type fuzzitClient struct {
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


func NewFuzzitClient(apiKey string) *fuzzitClient {
	c := &fuzzitClient{}
	c.httpClient = &http.Client{Timeout: 60 * time.Second}
	c.ApiKey = apiKey

	return c
}

func LoadFuzzitFromCache() (*fuzzitClient, error) {
	c := &fuzzitClient{}
	c.httpClient = &http.Client{Timeout: 60 * time.Second}
	err := c.ReAuthenticate(false)
	if err != nil {
		return nil, err
	}

	return c, nil
}

