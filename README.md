fuzzit.dev was [acquired](https://about.gitlab.com/press/releases/2020-06-11-gitlab-acquires-peach-tech-and-fuzzit-to-expand-devsecops-offering.html) by GitLab and the standalone service will soon be deperecated. The service will be available as part of GitLab Ultimate in the near future.

[![Fuzzit Logo](https://app.fuzzit.dev/static/fuzzit-logo.svg)](https://fuzzit.dev)

[![fuzzit](https://app.fuzzit.dev/static/fuzzit-passing-green.svg)](https://app.fuzzit.dev)
[![license](https://app.fuzzit.dev/static/license-apache-blue.svg)](https://github.com/fuzzitdev/Fuzzit/blob/master/LICENSE)

## Fuzzit
[Fuzzit](https://fuzzit.dev) helps you integrate Continuous Fuzzing to your [C/C++](https://github.com/fuzzitdev/example-c),
[Java](https://github.com/fuzzitdev/example-java), [Go](https://github.com/fuzzitdev/example-go), [Rust](https://github.com/fuzzitdev/example-rust) and [Swift](https://github.com/fuzzitdev/example-swift)
 projects with your current CI/CD workflow

[![Fuzzit Introduction](https://img.youtube.com/vi/Va7rfTTPiNo/maxresdefault.jpg)](https://www.youtube.com/watch?v=Va7rfTTPiNo)

## Download

#### Precompiled Binaries

You can download the precompiled release binary from [releases](https://github.com/fuzzitdev/fuzzit/releases) via web
or via

```bash
wget https://github.com/fuzzitdev/fuzzit/releases/download/<version>/fuzzit_<version>_<os>_<arch>
```

#### Go get

You can also use Go 1.12 or later to build the latest stable version from source:

```bash
GO111MODULE=on go get github.com/fuzzitdev/fuzzit/v2
```

#### Homebrew Tap

```bash
brew install fuzzitdev/tap/fuzzit
# After initial install you can upgrade the version via:
brew upgrade fuzzit
```

## Compilation

```bash
git clone git@github.com:fuzzitdev/fuzzit.git
cd fuzzit
go build .
```

## Usage

Fuzzit CLI can be used either locally or from your CI.

Run `fuzzit --help` to get a full list of commands, or check out our [docs](https://docs.fuzzit.dev).

## Examples

Fuzzit currently supports C/C++, Go and Rust

* [C/C++ example](https://github.com/fuzzitdev/example-c)
* [Java example](https://github.com/fuzzitdev/example-java)
* [Go example](https://github.com/fuzzitdev/example-go)
* [Rust example](https://github.com/fuzzitdev/example-rust)
* [Swift example](https://github.com/fuzzitdev/example-swift)

More information can be found in our [docs](https://docs.fuzzit.dev).

## OSS Projects Using Fuzzit
* GO
- [coredns/coredns](http://github.com/coredns/coredns)
- [prometheus/prometheus](http://github.com/prometheus/prometheus)
- [grpc-ecosystem/grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway)
- [google/syzkaller](http://github.com/google/syzkaller)
- [google/mtail](https://github.com/google/mtail)
- [open-policy-agent/opa](https://github.com/open-policy-agent/opa)
- [dutchcoders/transfer.sh](https://github.com/dutchcoders/transfer.sh)
- [tsenart/vegeta](https://github.com/tsenart/vegeta)
- [pelletier/go-toml](https://github.com/pelletier/go-toml)
- [mvdan/sh](https://github.com/mvdan/sh)
- [jdkato/prose](https://github.com/jdkato/prose)
- [klauspost/compress](https://github.com/klauspost/compress)
- [valyala/fasthttp](https://github.com/valyala/fasthttp)
- [lucas-clemente/quic-go](https://github.com/lucas-clemente/quic-go)
- [gomarkdown/markdown](https://github.com/gomarkdown/markdown)
- [pquerna/ffjson](https://github.com/pquerna/ffjson)
- [jaegertracing/jaeger](https://github.com/jaegertracing/jaeger)
- [caddyserver/caddy](https://github.com/caddyserver/caddy/tree/v2)
- [tealeg/xlsx](https://github.com/tealeg/xlsx)

* RUST
- [CraneStation/cranelift](https://github.com/CraneStation/cranelift)
- [image-rs/image-png](https://github.com/image-rs/image-png)
- [pest-parser/pest](https://github.com/pest-parser/pest)

* C/C++ 
- [systemd/systemd](https://github.com/systemd/systemd)
- [envoyproxy/envoy](https://github.com/envoyproxy/envoy)
- [apache/arrow](https://github.com/apache/arrow)
- [radare/radare2](https://github.com/radare/radare2)
- [AndreRenaud/PDFGen](https://github.com/AndreRenaud/PDFGen)

Use Fuzzit and you don't see your project here open a PR with your project!

## Contribution

Contributions are welcome. If you need an additional feature you can open a github issue, or send a PR
if you'd like to contribute it. Before contributing a big feature please open an issue so we can discuss and 
approve it before a lot of code is written. For bugfixes also open an issue or PR.


## Versioning

Fuzzit CLI Version contains three components x.y.z . an increase in `z` ensures backward compatability while increase
in `y` might introduce breaking changes.  

## Reporting Security Vulnerabilities

If you've found a vulnerability in Fuzzit please drop us a line at at [security@fuzzit.dev](security@fuzzit.dev)
. 

