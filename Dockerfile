FROM golang:1.7-alpine
LABEL maintainer="Thomas Schädler <thomas@lambda.li>"

# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/gufertum/transfer.sh

# build & install server
RUN go build -o /go/bin/transfersh github.com/gufertum/transfer.sh

ENTRYPOINT ["/go/bin/transfersh", "--listener", ":8080", "--provider", "s3"]  

EXPOSE 8080 8080
