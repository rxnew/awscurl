# awscurl

![release](https://github.com/rxnew/awscurl/actions/workflows/release.yml/badge.svg?branch=release)

A curl-like CLI application for requesting endpoints protected by AWS Signature Version 4.

## Installation

### Linux and Mac

```shell
curl -L https://github.com/rxnew/awscurl/releases/latest/download/awscurl-$(uname -s)-$(uname -m).tar.gz | tar -zx
```

## Quick Start

```shell
env AWS_PROFILE=xxx awscurl -X GET https://example.com
```
