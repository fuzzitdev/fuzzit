[![Fuzzit Logo](https://app.fuzzit.dev/static/fuzzit-logo.svg)](https://fuzzit.dev)

[![fuzzit](https://app.fuzzit.dev/static/fuzzit-passing-green.svg)](https://app.fuzzit.dev)
[![license](https://app.fuzzit.dev/static/license-apache-blue.svg)](https://github.com/fuzzitdev/Fuzzit/blob/master/LICENSE)
[![Join the chat at https://slack.fuzzit.dev](https://app.fuzzit.dev/static/slack-join.svg)](https://slack.fuzzit.dev)

## Fuzzit
[Fuzzit](https://fuzzit.dev) helps you integrate Continuous Fuzzing to your [C/C++](https://github.com/fuzzitdev/example-c),
 [Go](https://github.com/fuzzitdev/example-go), [Rust](https://github.com/fuzzitdev/example-rust) and [Swift](https://github.com/fuzzitdev/example-swift)
 projects with your current CI/CD workflow

[![Fuzzit Introduction](https://img.youtube.com/vi/Va7rfTTPiNo/maxresdefault.jpg)](https://www.youtube.com/watch?v=Va7rfTTPiNo)

## Download

#### Precompiled Binaries

You can download the precompiled release binary from [releases](https://github.com/fuzzitdev/fuzzit/releases) via web
or via

```bash
wget https://github.com/fuzzitdev/fuzzit/releases/download/<version>/fuzzit_<version>_<os>_<arch>
```

#### Go Get

Also, you can use the following command to download and compile (This usually takes some time so, it's usually faster to either download a pre-compiled release or download the source and build locally):

```bash
go get -v -u github.com/fuzzitdev/fuzzit
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
* [Go example](https://github.com/fuzzitdev/example-go)
* [Rust example](https://github.com/fuzzitdev/example-rust)
* [Swift example](https://github.com/fuzzitdev/example-swift)

More information can be found in our [docs](https://docs.fuzzit.dev).

## Notable OSS Projects Using Fuzzit
* [coredns/coredns (Go)](http://github.com/coredns/coredns)
* [prometheus/prometheus (Go)](http://github.com/prometheus/prometheus)
* [google/syzkaller (Go)](http://github.com/google/syzkaller)
* [systemd/systemd (C/C++)](https://github.com/systemd/systemd)
* [radare/radare2 (C/C++)](https://github.com/radare/radare2) 

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

