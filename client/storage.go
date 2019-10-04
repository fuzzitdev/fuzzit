package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/h2non/filetype"
	"github.com/mholt/archiver"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
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

func (c *FuzzitClient) uploadFile(filePath string, storagePath string, filename string) error {
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

func (c *FuzzitClient) archiveAndUpload(dirPath string, storagePath string, filename string) error {
	dir, err := ioutil.TempDir("", "archiveAndUpload")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir) // clean up

	tmpArchiveFileName := filepath.Join(dir, "archive.tar.gz")

	dirArchiver := archiver.NewTarGz()
	if err = dirArchiver.Archive([]string{dirPath}, tmpArchiveFileName); err != nil {
		return err
	}

	if err = c.uploadFile(tmpArchiveFileName, storagePath, filename); err != nil {
		return err
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

	split := strings.Split(resp.Header.Get("Content-Disposition"), "filename=")
	if len(split) == 2 {
		c.fuzzerFilename = split[1]
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func (c *FuzzitClient) downloadAndExtract(dirPath string, storagePath string) error {
	tmpArchiveFile, err := ioutil.TempFile("", "archive")
	if err != nil {
		return err
	}
	defer tmpArchiveFile.Close()

	if err := c.downloadFile(tmpArchiveFile.Name(), storagePath); err != nil {
		return err
	}

	if strings.HasSuffix(storagePath, "/fuzzer") && c.currentJob.Engine == "go-fuzz" {
		if _, err := copyFile(filepath.Join(dirPath, "fuzzer.zip"), tmpArchiveFile.Name()); err != nil {
			return err
		}

		return nil
	}

	buf, _ := ioutil.ReadFile(tmpArchiveFile.Name())

	var unarchiver archiver.Unarchiver

	// Try by filename fist
	if strings.HasSuffix(c.fuzzerFilename, ".jar") {
		if _, err := copyFile(filepath.Join(dirPath, "fuzzer.jar"), tmpArchiveFile.Name()); err != nil {
			return err
		}
		return nil
	} else if strings.HasSuffix(c.fuzzerFilename, ".tar.gz") {
		unarchiver = archiver.NewTarGz()
	} else if strings.HasSuffix(c.fuzzerFilename, ".zip") {
		unarchiver = archiver.NewZip()
	} else {
		kind, _ := filetype.Match(buf)
		switch kind.MIME.Value {
		case "application/gzip":
			unarchiver = archiver.NewTarGz()
		case "application/zip":
			unarchiver = archiver.NewZip()
		default:
			// assume executable
			if _, err := copyFile(filepath.Join(dirPath, "fuzzer"), tmpArchiveFile.Name()); err != nil {
				return err
			}
			return nil
		}
	}

	if err := unarchiver.Unarchive(tmpArchiveFile.Name(), dirPath); err != nil {
		return err
	}

	return nil
}
