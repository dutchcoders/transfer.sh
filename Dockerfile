FROM golang as builder
MAINTAINER Remco Verhoef <remco@dutchcoders.io>

# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/dutchcoders/transfer.sh
WORKDIR /go/src/github.com/dutchcoders/transfer.sh

# build binarie
RUN set -x && \
    go get -d -v . && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o transfersh .

# Take an empty container
FROM scratch

WORKDIR /root/

# Copy the binarie
RUN set -x && \
COPY --from=builder /go/src/github.com/dutchcoders/transfer.sh/transfersh .

EXPOSE 8080 8080

# Entrypoint to launch properly the container
ENTRYPOINT ["/root/transfersh"]
