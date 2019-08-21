FROM debian:stretch

LABEL maintainer="Fuzzit.dev, inc."

RUN apt-get -qqy update && apt-get install -y wget gnupg2 unzip

RUN echo "deb http://apt.llvm.org/stretch/ llvm-toolchain-stretch-9 main" >> /etc/apt/sources.list
RUN echo "deb-src http://apt.llvm.org/stretch/ llvm-toolchain-stretch-9 main" >> /etc/apt/sources.list
RUN wget -O - https://apt.llvm.org/llvm-snapshot.gpg.key| apt-key add -
RUN apt-get update && apt-get install -y libllvm-9-ocaml-dev libllvm9 llvm-9 llvm-9-dev llvm-9-doc llvm-9-examples llvm-9-runtime
RUN ln -s /usr/lib/llvm-9/bin/llvm-symbolizer /bin/llvm-symbolizer

WORKDIR /app
