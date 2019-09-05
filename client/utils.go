package client

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path"
	"path/filepath"
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

func GetValueFromEnv(variables ...string) string {
	for _, env := range variables {
		value := os.Getenv(env)
		if value != "" {
			return value
		}
	}
	return ""
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
