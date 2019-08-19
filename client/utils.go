package client

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path"
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

const runSh = `
#!/bin/sh

mkdir corpus_dir
mkdir seed_dir
touch corpus_dir/empty # This to avoid fuzzer stuck
touch seed_dir/empty # This is to avoid fuzzer stuck

echo "Downloading main corpus from Fuzzit servers..."
wget -O corpus.tar.gz $CORPUS_LINK || rm -f corpus.tar.gz # remove empty file if corpus doesn't exist
if test -f "corpus.tar.gz"; then
    tar -xzf corpus.tar.gz -C corpus_dir
else
    echo "corpus is still empty. continuing without..."
fi

echo "Downloading seed corpus from Fuzzit servers..."
wget -O seed $SEED_LINK || rm -f seed
if test -f "seed"; then
    case $(file --mime-type -b seed) in
        application/gzip|application/x-gzip)
           tar -xzf seed -C seed_dir
        ;;
        application/zip)
           unzip seed -d seed_dir
        ;;
        *)
           echo "seed in unknown format. Please upload seed in tar.gz or zip format. If you did and you believe it's
           a bug, please contact support@fuzzit.dev"
           exit 1
           ;;
    esac
else
    echo "seed corpus is empty. continuing without..."
fi

if test -f "fuzzer"; then
    chmod a+x fuzzer
	echo "running regression..."
    ./fuzzer -exact_artifact_path=./artifact -print_final_stats=1 $(find seed_dir -type f) ./corpus_dir/* $ARGS || exit 1
else
    echo "failed to locate fuzzer. does 'fuzzer' executable exist in the archive?"
	exit 1
fi
`
