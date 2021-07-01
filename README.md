# awscurl

![release](https://github.com/rxnew/awscurl/actions/workflows/release.yml/badge.svg?branch=release)

A curl-like CLI application for requesting endpoints protected by AWS Signature Version 4.

## Installation

### Linux and Mac

```
$ curl -L https://github.com/rxnew/awscurl/releases/latest/download/awscurl-$(uname -s)-$(uname -m).tar.gz | sudo tar -zxf -C /usr/local/bin
```

### Upgrade

Perform the installation procedure again.

### Uninstall

```
$ sudo rm /usr/local/bin/awscurl
```

## Quick Start

```
$ export AWS_PROFILE=xxx
$ awscurl -X GET https://example.com
```
