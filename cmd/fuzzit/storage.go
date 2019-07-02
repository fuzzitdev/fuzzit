package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

var httpClient = &http.Client{Timeout: 120 * time.Second}

func getStorageLink(storagePath string, apiKey string) (string, error) {
	r, err := httpClient.Get(fmt.Sprintf("https://app.fuzzit.dev/getStorageLink?path=%s&api_key=%s", storagePath, apiKey))
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

func uploadFile(filePath string, storagePath string, apiKey string, contentType string, filename string) error {
	data, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer data.Close()

	storageLink, err := getStorageLink(storagePath, apiKey)
	if err != nil {
		return err
	}

	fmt.Println("uploading fuzzer...")
	req, err := http.NewRequest("PUT", storageLink, data)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Disposition", "attachment; filename=" + filename)
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