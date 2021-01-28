# Stage 1: Pull a docker binary
# TODO: We want to link directly into this, but I'm not sure how to do it
FROM alpine as dockerinstaller

# Download docker into /dockerclient
ARG DOCKER_VERSION="19.03.12"
RUN apk --update add curl \
    && mkdir -p /dockerclient/ \
    && curl -L "https://download.docker.com/linux/static/stable/x86_64/docker-$DOCKER_VERSION.tgz" | tar -xz -C /dockerclient

# Stage 2: Build the application
FROM golang as builder
ADD . /app/src/github.com/tkw1536/proxyssh
WORKDIR /app/src/github.com/tkw1536/proxyssh
RUN CGO_ENABLED=0 GOOS=linux go build -o /dockersshd ./cmd/dockersshd

# Stage 3: Put it all together
FROM alpine
EXPOSE 2222
COPY --from=dockerinstaller /dockerclient/docker/docker /usr/local/bin/docker
COPY --from=builder /dockersshd /dockersshd
ENTRYPOINT [ "/dockersshd" ]