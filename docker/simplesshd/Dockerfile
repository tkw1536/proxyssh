FROM golang:alpine as builder
RUN apk add --no-cache git

ADD . /app/src/github.com/tkw1536/proxyssh
WORKDIR /app/src/github.com/tkw1536/proxyssh/
RUN go build -o /simplesshd ./cmd/simplesshd

FROM alpine
EXPOSE 2222
COPY --from=builder /simplesshd /simplesshd
ENTRYPOINT [ "/simplesshd" "-shell" "/bin/sh" ]