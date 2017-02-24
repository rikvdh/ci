CI
==

[![Build Status](https://travis-ci.org/rikvdh/ci.svg?branch=master)](https://travis-ci.org/rikvdh/ci)
[![Go Report Card](https://goreportcard.com/badge/github.com/rikvdh/ci)](https://goreportcard.com/report/github.com/rikvdh/ci)
[![GoDoc](https://godoc.org/github.com/rikvdh/ci?status.svg)](https://godoc.org/github.com/rikvdh/ci)
[![codebeat badge](https://codebeat.co/badges/e1d86b8b-eaa3-45f5-8ee9-02d6cb31352b)](https://codebeat.co/projects/github-com-rikvdh-ci)

Self-hosted Continuous Integration (CI) platform, easy deployment,
compatible with Travis-CI

**WARNING:** CI is **under development**, it works, but is definitly not stable yet. Also still in the need of a better name!

## Status

Currently it builds, monitors your GIT repository.

## Getting CI

You need the Golang toolchain 1.8 or above. Get it from [here](https://golang.org/dl/). No binary releases yet!
Also you must make sure you've installed the latest version of [Docker](https://www.docker.com/products/overview#install_the_platform)

```bash
$ go get github.com/rikvdh/ci
```

Then you can run `./ci` and a web-interface should be active on `localhost:8081`

## Travis compatibility

The aim is to be compatible with Travis-CI. Now the CI checks for `.ci.yml` in the root of your repository,
when it doesn't exist, it tries `.travis.yml`

The format is the same as for Travis-CI. When incompatibilities are found, please report them via the issue tracker!

## Contributing or problems

Of course we appreciate your contributions! Simply send a (pull-request)[https://github.com/rikvdh/ci/pulls]!
We shall try to reply within a day or so. If you're working on something bigger, let us know via an issue.

With the current status of CI, it is very likely things are broken. If you think you found a bug in the current implementation.
Let us know via the (GitHub issue tracker)[https://github.com/rikvdh/ci/issues]
