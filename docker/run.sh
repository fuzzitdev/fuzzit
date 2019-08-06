#!/bin/sh
set -x

mkdir corpus_dir
mkdir seed_dir
touch corpus_dir/empty # This to avoid fuzzer stuck
touch seed_dir/empty # This is to avoid fuzzer stuck

wget -O corpus.tar.gz $CORPUS_LINK || rm -f corpus.tar.gz # remove empty file if corpus doesn't exist
if test -f "corpus.tar.gz"; then
    tar -xzvf corpus.tar.gz -C corpus_dir
else
    echo "corpus is still empty. continuing without..."
fi

wget -O seed $SEED_LINK || rm -f seed
if test -f "seed"; then
    case $(file --mime-type -b seed) in
        application/x-gzip)
           tar -xzvf seed -C seed_dir
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
    echo "running fuzzer"
    chmod a+x fuzzer
    ./fuzzer -exact_artifact_path=./artifact -print_final_stats=1 ./seed_dir/* ./corpus_dir/* $ARGS
else
    echo "failed to locate fuzzer. does 'fuzzer' executable exist in the archive?"
fi