# proxyssh

[![Build Status](https://travis-ci.com/tkw1536/proxyssh.svg?branch=main)](https://travis-ci.com/tkw1536/proxyssh)

A golang package based on https://github.com/gliderlabs/ssh that enables easily running commands via ssh. 
Also includes Docker support. 

## Use Cases

The main use case of this package is to provide a simple interface to build an ssh server that runs commands on the host it is running on. 
See the `cmd/simplesshd` package. 

A secondary use case is to provide an ssh server that runs commands in matching docker containers. 
See the `cmd/dockersshd` package. 

For a more detailed overall documentation, see the [godoc](https://pkg.go.dev/github.com/tkw1536/proxyssh). 

## Dockerfiles

This repository contains Dockerfiles for both of the examples, called `Dockerfile.simple` and `Dockerfile` respectively. 
The Dockerfile for the docker scenario is available as the automated build [tkw01536/proxyssh](https://hub.docker.com/r/tkw01536/proxyssh). 
It is intended to be run as follows:

```bash
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -p 2222:2222 tkw01536/proxyssh
```
