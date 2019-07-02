![Fuzzit Logo](https://app.fuzzit.dev/static/fuzzit.svg)

[![fuzzit](https://app.fuzzit.dev/static/fuzzit-passing-green.svg)](https://app.fuzzit.dev)
[![license](https://app.fuzzit.dev/static/license-apache-blue.svg)](https://github.com/fuzzitdev/Fuzzit/blob/master/LICENSE)
[![Join the chat at https://slack.fuzzit.dev](https://app.fuzzit.dev/static/slack-join.svg)](https://slack.fuzzit.dev)

## Fuzzit
Fuzzit helps you integrate Continuous Fuzzing to your current CI/CD workflow


## Compilation

```bash
git clone git@github.com:fuzzitdev/fuzzit.git
cd fuzzit
go build ./...
```

## Usage

Fuzzit CLI can be used either locally or from your CI.

Run `snyk --help` to get a full list of commands or checkout our [docs](https://docs.fuzzit.dev).

## Versioning

Fuzzit CLI Version contains three components x.y.z . an increase in `z` ensures backward comparability while increase
in `y` might introduce breaking changes.  

## Reporting Security Vulnerabilities

If you've found a vulnerability in Fuzzit please drop us a line at at [security@fuzzit.dev](security@fuzzit.dev)
. 

