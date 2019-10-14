FROM debian:stretch

LABEL maintainer="Fuzzit.dev, inc."

RUN apt-get -qqy update && apt-get install -y wget gnupg2 unzip

RUN echo "deb http://apt.llvm.org/stretch/ llvm-toolchain-stretch-8 main" >> /etc/apt/sources.list
RUN echo "deb-src http://apt.llvm.org/stretch/ llvm-toolchain-stretch-8 main" >> /etc/apt/sources.list
RUN wget -O - https://apt.llvm.org/llvm-snapshot.gpg.key| apt-key add -
RUN apt-get update && apt-get install -y libllvm-8-ocaml-dev libllvm8 llvm-8 llvm-8-dev llvm-8-doc llvm-8-examples llvm-8-runtime
RUN ln -s /usr/lib/llvm-8/bin/llvm-symbolizer /bin/llvm-symbolizer

RUN apt-get -qqy update && apt-get install -y git clang-8 libclang-common-8-dev libclang-8-dev libclang1-8 clang-format-8 cmake

WORKDIR /app
