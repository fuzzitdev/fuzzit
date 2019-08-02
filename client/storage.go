package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

var httpClient = &http.Client{Timeout: 120 * time.Second}

type storageLinkResponse struct {
	StorageLink string `json:"storage_link"`
}

func (c *fuzzitClient) getStorageLink(storagePath string) (string, error) {
	uri := fmt.Sprintf("https://app.fuzzit.dev/getStorageLink?path=%s&api_key=%s", url.QueryEscape(storagePath), url.QueryEscape(c.ApiKey))
	r, err := httpClient.Get(uri)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()
	if r.StatusCode != 200 {
		return "", fmt.Errorf("API Key is not valid")
	}

	res := storageLinkResponse{}
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
	if res.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyBytes)
		return fmt.Errorf(bodyString)
	}
	defer res.Body.Close()
	return nil
}
