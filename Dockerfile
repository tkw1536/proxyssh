FROM golang:alpine as builder

ADD . /app/src/github.com/tkw1536/proxyssh
WORKDIR /app/src/github.com/tkw1536/proxyssh/_example/docker
RUN go build -o /main main.go

# Download docker into /dockerclient
ARG DOCKER_VERSION="19.03.12"
RUN apk --update add curl \
    && mkdir -p /dockerclient/ \
    && curl -L "https://download.docker.com/linux/static/stable/x86_64/docker-$DOCKER_VERSION.tgz" | tar -xz -C /dockerclient

FROM alpine
EXPOSE 2222
COPY --from=builder /dockerclient/docker/docker /usr/local/bin/docker
COPY --from=builder /main /main
ENTRYPOINT [ "/main" ]