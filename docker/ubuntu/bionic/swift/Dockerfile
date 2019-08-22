FROM ubuntu:18.04

LABEL maintainer="Fuzzit.dev, inc."

RUN apt-get -qqy update && apt-get install -y wget gnupg2 unzip git curl

RUN echo "deb http://apt.llvm.org/bionic/ llvm-toolchain-bionic-9 main" >> /etc/apt/sources.list
RUN echo "deb-src http://apt.llvm.org/bionic/ llvm-toolchain-bionic-9 main" >> /etc/apt/sources.list
RUN wget -O - https://apt.llvm.org/llvm-snapshot.gpg.key| apt-key add -
RUN apt-get update && apt-get install -y clang-9 lldb-9 lld-9 libllvm-9-ocaml-dev libllvm9 llvm-9 llvm-9-dev llvm-9-doc llvm-9-examples llvm-9-runtime
RUN ln -s /usr/lib/llvm-9/bin/llvm-symbolizer /bin/llvm-symbolizer

# Install swiftenv
ENV SWIFTENV_ROOT="/root/.swiftenv"
ENV PATH="/root/.swiftenv/bin:$PATH"

RUN apt install -y  lsb-release curl git libxml2 libatomic1 libcurl4-openssl-dev libc6-dev python2.7 libpython2.7-dev libpython2.7
RUN git clone https://github.com/kylef/swiftenv.git ~/.swiftenv
RUN echo 'export SWIFTENV_ROOT="$HOME/.swiftenv"' >> ~/.bashrc
RUN echo 'export PATH="$SWIFTENV_ROOT/bin:$PATH"' >> ~/.bashrc
RUN echo 'eval "$(swiftenv init -)"' >> ~/.bashrc
RUN swiftenv install DEVELOPMENT-SNAPSHOT-2019-08-16-a

WORKDIR /app
