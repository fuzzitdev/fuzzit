FROM ubuntu:18.04

LABEL maintainer="Fuzzit.dev, inc."

RUN apt-get -qqy update && apt-get install -y wget gnupg2 unzip git curl libxml2 libatomic1 libc6-dev lsb-release

# Install clang-9
RUN echo "deb http://apt.llvm.org/bionic/ llvm-toolchain-bionic-9 main" >> /etc/apt/sources.list
RUN echo "deb-src http://apt.llvm.org/bionic/ llvm-toolchain-bionic-9 main" >> /etc/apt/sources.list
RUN wget -O - https://apt.llvm.org/llvm-snapshot.gpg.key| apt-key add -
RUN apt-get update && apt-get install -y libllvm-9-ocaml-dev libllvm9 llvm-9 llvm-9-dev llvm-9-doc llvm-9-examples llvm-9-runtime clang-9 lldb-9 lld-9
RUN ln -s /usr/lib/llvm-9/bin/llvm-symbolizer /bin/llvm-symbolizer
RUN ln -s /usr/bin/clang-9 /bin/clang
RUN ln -s /usr/bin/clang++-9 /bin/clang++

# Install Swift development snapshot
RUN curl https://swift.org/keys/all-keys.asc | gpg2 --import -
ENV SWIFT_URL=https://swift.org/builds/development/ubuntu1804/swift-DEVELOPMENT-SNAPSHOT-2019-08-17-a/swift-DEVELOPMENT-SNAPSHOT-2019-08-17-a-ubuntu18.04.tar.gz
ENV ARCHIVE_NAME=swift-DEVELOPMENT-SNAPSHOT-2019-08-17-a-ubuntu18.04.tar.gz
# Install Swift toolchain for ubuntu
RUN wget $SWIFT_URL && \
    wget $SWIFT_URL.sig && \
    gpg2 --verify $ARCHIVE_NAME.sig && \
    tar -xvzf $ARCHIVE_NAME --directory / --strip-components=1 && \
    chmod -R o+r /usr/lib/swift

WORKDIR /app
