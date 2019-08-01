![Fuzzit Logo](https://app.fuzzit.dev/static/fuzzit.svg)

[![fuzzit](https://app.fuzzit.dev/static/fuzzit-passing-green.svg)](https://app.fuzzit.dev)
[![license](https://app.fuzzit.dev/static/license-apache-blue.svg)](https://github.com/fuzzitdev/Fuzzit/blob/master/LICENSE)
[![Join the chat at https://slack.fuzzit.dev](https://app.fuzzit.dev/static/slack-join.svg)](https://slack.fuzzit.dev)

## Fuzzit
[Fuzzit](https://fuzzit.dev) helps you integrate Continuous Fuzzing to your current CI/CD workflow

[![Fuzzit Introduction](https://img.youtube.com/vi/Va7rfTTPiNo/maxresdefault.jpg)](https://www.youtube.com/watch?v=Va7rfTTPiNo)

## Download

You can download the precompiled release binary from [releases](https://github.com/fuzzitdev/fuzzit/releases) via web
or via

`wget https://github.com/fuzzitdev/fuzzit/releases/download/<version>/fuzzit_<version>_<os>_<arch>`

Also, you can use the following command to download and compile (This usually takes some time so it's usually faster to either download a pre-compiled release or download the source and build locally):

`go get -v -u github.com/fuzzitdev/fuzzit/...`

## Compilation

```bash
git clone git@github.com:fuzzitdev/fuzzit.git
cd fuzzit
go build .
```

## Usage

Fuzzit CLI can be used either locally or from your CI.

Run `fuzzit --help` to get a full list of commands or checkout our [docs](https://docs.fuzzit.dev).

## Examples

Fuzzit currently supports C/C++ and Go

* [C/C++ examples](https://github.com/fuzzitdev/example-c)
* [Go examples](https://github.com/fuzzitdev/example-go)

More information can be at our [docs](https://docs.fuzzit.dev) 

## Contribution

Contributions are welcome. If you need additional feature either open a github issue or a PR
if you like to contribute it. Before contributing a big feature please open an issue so we can discuss and 
approve before a lot of code is written. For bugfixes also open an issue or PR.


## Versioning

Fuzzit CLI Version contains three components x.y.z . an increase in `z` ensures backward comparability while increase
in `y` might introduce breaking changes.  

## Reporting Security Vulnerabilities

If you've found a vulnerability in Fuzzit please drop us a line at at [security@fuzzit.dev](security@fuzzit.dev)
. 

