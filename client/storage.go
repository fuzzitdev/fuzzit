package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

var httpClient = &http.Client{Timeout: 120 * time.Second}

type storageLinkResponse struct {
	StorageLink string `json:"storage_link"`
}

func (c *fuzzitClient) getStorageLink(storagePath string) (string, error) {
	r, err := httpClient.Get(fmt.Sprintf("https://app.fuzzit.dev/getStorageLink?path=%s&api_key=%s", storagePath, c.ApiKey))
	if err != nil {
		return "", err
	}
	defer r.Body.Close()
	if r.StatusCode != 200 {
		return "", fmt.Errorf("API Key is not valid")
	}

	var res storageLinkResponse
	err = json.NewDecoder(r.Body).Decode(&res)
	if err != nil {
		return "", err
	}

	return res.StorageLink, nil
}

func (c *fuzzitClient) uploadFile(filePath string, storagePath string, contentType string, filename string) error {
	data, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer data.Close()

	storageLink, err := c.getStorageLink(storagePath)
	if err != nil {
		return err
	}

	fmt.Println("uploading fuzzer...")
	req, err := http.NewRequest("PUT", storageLink, data)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Disposition", "attachment; filename="+filename)
	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatal(err)
		}
		return errors.New(string(bodyBytes))
	}
	return nil
}
