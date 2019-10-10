package client

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
)

func getCacheFile() (string, error) {
	// This is to solve problem with snap $HOME restrictions
	home := os.Getenv("HOME")
	if home == "" {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		home = usr.HomeDir
	}
	cacheFile := path.Join(home, ".fuzzit.cache")
	return cacheFile, nil
}

func catFile(path string) error {
	fh, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fh.Close()

	_, err = io.Copy(os.Stdout, fh)
	if err != nil {
		return err
	}

	return nil
}

func catLastBytes(path string, lastBytes int64) error {
	fh, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fh.Close()

	buf := make([]byte, lastBytes)
	stat, err := os.Stat(path)
	start := 0
	if stat.Size() > lastBytes {
		start = int(stat.Size() - lastBytes)
	}

	_, err = fh.ReadAt(buf, int64(start))
	if err != nil {
		return err
	}

	log.Printf("%s\n", buf)

	return nil
}

func splitAndRemoveEmpty(s string, delimiter string) []string {
	splitted := strings.Split(s, delimiter)
	var withoutEmptyStrings []string
	for _, str := range splitted {
		if str != "" {
			withoutEmptyStrings = append(withoutEmptyStrings, str)
		}
	}

	return withoutEmptyStrings
}

func GetValueFromEnv(variables ...string) string {
	for _, env := range variables {
		value := os.Getenv(env)
		if value != "" {
			return value
		}
	}
	return ""
}

func IsDirEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}

func copyFile(dst, src string) (int64, error) {
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
	nBytes, err := io.Copy(destination, source)
	errClose := destination.Close()
	if err != nil {
		return 0, err
	}
	if errClose != nil {
		return 0, errClose
	}

	return nBytes, nil
}

func listFiles(dst string) ([]string, error) {
	var fileList []string
	err := filepath.Walk(dst, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if !info.IsDir() {
			fileList = append(fileList, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return fileList, nil
}

func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func mergeDirectories(dst string, src string) error {
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if !info.IsDir() {
			fileName := info.Name()
			dstPath := filepath.Join(dst, fileName)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				err = os.Rename(filepath.Join(src, fileName), dstPath)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})

	return err
}

func createDirIfNotExist(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.Mkdir(path, 0644); err != nil {
			return err
		}
	}
	return nil
}

func Contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}
