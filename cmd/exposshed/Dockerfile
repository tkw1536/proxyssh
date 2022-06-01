# Stage 1: Create a www-data user
FROM alpine as permission

# Create a www-data user
RUN set -x ; \
  addgroup -g 82 -S www-data ; \
  adduser -u 82 -D -S -G www-data www-data && exit 0 ; exit 1

# Make a volume under data for the user
RUN mkdir -p /data/keys/ && chown -R www-data:www-data /data/ && chmod -R 777 /data/

# Stage 2: Build exposshed
FROM golang:1.18 as build
ADD . /app/src/github.com/tkw1536/proxyssh
WORKDIR /app/src/github.com/tkw1536/proxyssh/
RUN CGO_ENABLED=0 GOOS=linux go build -o /exposshed ./cmd/exposshed

# Stage 3: Put it all together
FROM scratch

# Copy user and password
COPY --from=permission /etc/passwd /etc/passwd
COPY --from=permission /etc/group /etc/group

# Copy over executable
COPY --from=build /exposshed /exposshed

# Copy over and make data volume
COPY --from=permission /data/ /data/
VOLUME /data/

# drop privilidges
USER www-data:www-data

EXPOSE 2222
EXPOSE 8080
ENTRYPOINT ["/exposshed", "-hostkey", "/data/keys/hostkey"]
CMD [""]