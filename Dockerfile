# Default to Go 1.17
ARG GO_VERSION=1.17
FROM golang:${GO_VERSION}-alpine as build

# empty dir and /etc/passwd file to run transfer with unprivileged user
RUN install -g nobody -o nobody -m 0770 -d /tmp/empty-dir-owned-by-nobody
RUN echo 'nobody:x:65534:65534:nobody:/:/sbin/nologin' > /tmp/passwd

# Necessary to run 'go get' and to compile the linked binary
RUN apk add git musl-dev

ADD . /go/src/github.com/dutchcoders/transfer.sh

WORKDIR /go/src/github.com/dutchcoders/transfer.sh

ENV GO111MODULE=on

# build & install server
RUN CGO_ENABLED=0 go build -tags netgo -ldflags "-X github.com/dutchcoders/transfer.sh/cmd.Version=$(git describe --tags) -a -s -w -extldflags '-static'" -o /go/bin/transfersh

FROM scratch AS final
LABEL maintainer="Andrea Spacca <andrea.spacca@gmail.com>"

COPY --chown=65534:65534 --from=build /tmp/empty-dir-owned-by-nobody /tmp
COPY --from=build  /tmp/passwd /etc/passwd
COPY --from=build  /go/bin/transfersh /go/bin/transfersh
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

USER nobody
ENTRYPOINT ["/go/bin/transfersh", "--listener", ":8080"]

EXPOSE 8080
