FROM ubuntu:18.04

LABEL maintainer="Fuzzit.dev, inc."

RUN apt-get -qqy update && apt-get install -y wget gnupg2 unzip git curl

# Install swiftenv
ENV SWIFTENV_ROOT="/root/.swiftenv"
ENV PATH="/root/.swiftenv/bin:$PATH"

RUN apt install -y  lsb-release curl git libxml2 libatomic1 libcurl4-openssl-dev libc6-dev
RUN git clone https://github.com/kylef/swiftenv.git ~/.swiftenv
RUN echo 'export SWIFTENV_ROOT="$HOME/.swiftenv"' >> ~/.bashrc
RUN echo 'export PATH="$SWIFTENV_ROOT/bin:$PATH"' >> ~/.bashrc
RUN echo 'eval "$(swiftenv init -)"' >> ~/.bashrc
RUN swiftenv install DEVELOPMENT-SNAPSHOT-2019-08-16-a

WORKDIR /app
