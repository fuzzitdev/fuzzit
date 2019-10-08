FROM docker.pkg.github.com/fuzzitdev/fuzzit/debian-llvm8:latest

LABEL maintainer="Fuzzit.dev, inc."

RUN apt-get -qqy update && apt-get install -y git clang-8 libclang-common-8-dev libclang-8-dev libclang1-8 clang-format-8 cmake

WORKDIR /app
