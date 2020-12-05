# proxyssh

![CI Status](https://github.com/tkw1536/proxyssh/workflows/CI/badge.svg)


A golang package based on https://github.com/gliderlabs/ssh that enables easily running commands via ssh. 
Also includes Docker support. 

## Use Cases

The main use case of this package is to provide a simple interface to build an ssh server that runs commands on the host it is running on. 
See the `cmd/simplesshd` package. 

A secondary use case is to provide an ssh server that runs commands in matching docker containers. 
See the `cmd/dockersshd` package. 

A third use case of this package is to provide an ssh server that provides no shell access, but permits port forwarding only. 
See the `cmd/exposshed` package.

For a more detailed overall documentation, see the [godoc](https://pkg.go.dev/github.com/tkw1536/proxyssh). 

## Dockerfiles

This repository contains Dockerfiles for all of the examples, called `docker/${example}/Dockerfile`. 

These are available as the GitHub Packages [dockersshd](https://github.com/tkw1536/proxyssh/packages/391072), [simplesshd](https://github.com/tkw1536/proxyssh/packages/391070) and [exposshed](https://github.com/tkw1536/proxyssh/packages/) respectively. 

The `dockersshd` image can be run as follows:

```bash
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -p 2222:2222 docker.pkg.github.com/tkw1536/proxyssh/dockersshd:latest
```

To e.g. allow clients to expose port 5000 the `exposshed` image can be run as follows

```bash
docker run -t -v /path/to/hostkeys:/data/ -i --rm -p 2222:localhost:2222 -p 5000:5000 docker.pkg.github.com/tkw1536/proxyssh/exposshed:latest -R 0.0.0.0:5000
```

For legacy reasons, the `dockersshd` Image is also available as the automated build [tkw01536/proxyssh](https://hub.docker.com/r/tkw01536/proxyssh). 
