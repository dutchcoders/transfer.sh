# Default to Go 1.17
ARG GO_VERSION=1.17
FROM golang:${GO_VERSION}-alpine as build

# Necessary to run 'go get' and to compile the linked binary
RUN apk add git musl-dev

ADD . /go/src/github.com/dutchcoders/transfer.sh

WORKDIR /go/src/github.com/dutchcoders/transfer.sh

ENV GO111MODULE=on

# build & install server
RUN CGO_ENABLED=0 go build -tags netgo -ldflags "-X github.com/dutchcoders/transfer.sh/cmd.Version=$(git describe --tags) -a -s -w -extldflags '-static'" -o /go/bin/transfersh

ARG PUID=1000 \
    PGID=1000

RUN mkdir -p /tmp/useradd && \
    echo "transfersh:x:${PUID}:${PGID}::/nonexistent:/sbin/nologin" >> /tmp/useradd/passwd && \
    echo "transfersh:!:::::::" >> /tmp/useradd/shadow && \
    echo "transfersh:x:${PGID}:" >> /tmp/useradd/group && \
    echo "transfersh:!::" >> /tmp/useradd/groupshadow

FROM scratch AS final
LABEL maintainer="Andrea Spacca <andrea.spacca@gmail.com>"

COPY --from=build /tmp/useradd/* /etc/
COPY --from=build --chown=transfersh  /go/bin/transfersh /go/bin/transfersh
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

USER transfersh

ENTRYPOINT ["/go/bin/transfersh", "--listener", ":8080"]

EXPOSE 8080
