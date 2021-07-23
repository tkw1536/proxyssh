# proxyssh

![CI Status](https://github.com/tkw1536/proxyssh/workflows/CI/badge.svg)


A golang package based on https://github.com/gliderlabs/ssh that enables easily creating different ssh servers. 
Also includes Docker support. 

## Use Cases

The main use case of this package is to provide a simple interface to build an ssh server that runs commands on the host it is running on. 
See the `cmd/simplesshd` package. 

A secondary use case is to provide an ssh server that runs commands in matching docker containers. 
See the `cmd/dockersshd` package. 

A third use case of this package is to provide an ssh server that provides no shell access, but permits port forwarding only. 
See the `cmd/exposshed` package.

For a more detailed overall documentation, see the [godoc](https://pkg.go.dev/github.com/tkw1536/proxyssh). 

## Tests

This package comes with a tests suite as well as a builtin memory-leak detector.
The tests can be run as any normal go test suite can:

    go test ./...

Some tests require `docker-compose` and `/bin/bash` to be installed on the local machine.
They furthermore require a network connection to download the `alpine` docker image during tests.
These special tests are not run by default, but only when the `dockertests` tag is provided.
To run these tests, use:

    go test -tag=dockertest ./...

The memory leak detector is not enabled by default and not used during the tests.
By default, all code calling the memory leak detector is removed during compilation. 

The detector can be enabled by adding the `leak` go build tag. 
When enabled, all executables output memory leak messages for every connection to the console.

## Dockerfiles

This repository contains Dockerfiles for all of the examples, called `cmd/${example}/Dockerfile`. 

These are available as the GitHub Packages [dockersshd](https://github.com/users/tkw1536/packages/container/package/dockersshd), [simplesshd](https://github.com/users/tkw1536/packages/container/package/simplesshd) and [exposshed](https://github.com/users/tkw1536/packages/container/package/exposshed) respectively. 

The `dockersshd` image can be run as follows:

```bash
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -p 2222:2222 ghcr.io/tkw1536/dockersshd:latest
```

To e.g. allow clients to expose port 8080 the `exposshed` image can be run as follows

```bash
docker run -t -v /path/to/hostkeys:/data/ -i --rm -p 2222:localhost:2222 -p 8080:8080 ghcr.io/tkw1536/exposshed:latest -R 0.0.0.0:8080
```
