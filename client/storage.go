package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

func (c *FuzzitClient) getStorageLink(storagePath string, action string) (string, error) {
	uri := fmt.Sprintf("https://app.fuzzit.dev/getStorageLinkV3?path=%s&api_key=%s&action=%s",
		url.QueryEscape(storagePath),
		url.QueryEscape(c.ApiKey),
		action)
	r, err := httpClient.Get(uri)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()
	if r.StatusCode != 200 {
		return "", errors.New(r.Status)
	}

	var res storageLinkResponse
	err = json.NewDecoder(r.Body).Decode(&res)
	if err != nil {
		return "", err
	}

	return res.StorageLink, nil
}

func (c *FuzzitClient) uploadFile(filePath string, storagePath string, contentType string, filename string) error {
	data, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer data.Close()

	storageLink, err := c.getStorageLink(storagePath, "create")
	if err != nil {
		return err
	}

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

func (c *FuzzitClient) downloadFile(filePath string, storagePath string) error {
	storageLink, err := c.getStorageLink(storagePath, "read")
	if err != nil {
		return err
	}

	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(storageLink)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
