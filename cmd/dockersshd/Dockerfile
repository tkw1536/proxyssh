# Stage 1: Build the application
FROM golang:1.17 as builder
ADD . /app/src/github.com/tkw1536/proxyssh
WORKDIR /app/src/github.com/tkw1536/proxyssh
RUN CGO_ENABLED=0 GOOS=linux go build -o /dockersshd ./cmd/dockersshd

# Stage 2: Put it all together
FROM alpine
EXPOSE 2222
COPY --from=builder /dockersshd /dockersshd
ENTRYPOINT [ "/dockersshd" ]