# Stage 1: Build simplesshd
FROM golang:1.18 as build

ADD . /app/src/github.com/tkw1536/proxyssh
WORKDIR /app/src/github.com/tkw1536/proxyssh/
RUN CGO_ENABLED=0 GOOS=linux go build -o /simplesshd ./cmd/simplesshd

# Stage 2: Put it in the runtime image
FROM alpine
EXPOSE 2222
COPY --from=build /simplesshd /simplesshd
ENTRYPOINT [ "/simplesshd" "-shell" "/bin/sh" ]