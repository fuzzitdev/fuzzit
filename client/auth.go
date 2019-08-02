package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/user"
	"path"
	"time"

	"cloud.google.com/go/firestore"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
)

func (c *fuzzitClient) ReAuthenticate(force bool) error {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	CacheFile := path.Join(usr.HomeDir, ".fuzzit.cache")

	if !force {
		file, err := os.Open(CacheFile)
		if err != nil {
			return err
		}
		defer file.Close()

		err = json.NewDecoder(file).Decode(c)
		if err != nil {
			return err
		}
	}

	if c.ApiKey == "" {
		return errors.New("API Key is no configured, please run fuzzit auth [api_key]")
	}

	if c.IdToken == "" || (time.Now().Unix()-c.LastRefresh) > 60*45 {
		createCustomTokenEndpoint := fmt.Sprintf("%s/createCustomToken?api_key=%s", FuzzitEndpoint, url.QueryEscape(c.ApiKey))
		r, err := c.httpClient.Get(createCustomTokenEndpoint)
		if err != nil {
			return err
		}
		defer r.Body.Close()
		if r.StatusCode != 200 {
			return errors.New("API Key is not valid")
		}

		err = json.NewDecoder(r.Body).Decode(c)
		if err != nil {
			return err
		}

		r, err = c.httpClient.Post(
			"https://www.googleapis.com/identitytoolkit/v3/relyingparty/verifyCustomToken?key=AIzaSyCs_Sm1VOKZwJZmTXdOCvs1wyn91vYMNSY",
			"application/json",
			bytes.NewBuffer([]byte(fmt.Sprintf(`{"token": "%s", "returnSecureToken": true}`, c.CustomToken))))
		if err != nil {
			return err
		}
		defer r.Body.Close()

		err = json.NewDecoder(r.Body).Decode(c)
		if err != nil {
			return nil
		}
		c.LastRefresh = time.Now().Unix()

		cBytes, err := json.MarshalIndent(c, "", "")
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(CacheFile, cBytes, 0644)
		if err != nil {
			return err
		}
	}

	token := oauth2.Token{
		AccessToken:  c.IdToken,
		RefreshToken: c.RefreshToken,
		Expiry:       time.Time{},
		TokenType:    "Bearer",
	}

	tokenSource := oauth2.StaticTokenSource(&token)
	ctx := context.Background()

	// some known issue with go afaik
	firestoreClient, err := firestore.NewClient(ctx, "fuzzit-b5fbf", option.WithTokenSource(tokenSource))
	c.firestoreClient = firestoreClient

	if err != nil {
		return err
	}

	return nil
}
