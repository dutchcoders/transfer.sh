FROM golang
MAINTAINER Remco Verhoef <remco@dutchcoders.io>

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

# Copy the local package files to the container's workspace.
ADD ./transfersh-server /go/src/app

# install dependencies
RUN go get ./

# build & install server
RUN go install . 

ENTRYPOINT ["/go/bin/app", "--port", "8080"]  

EXPOSE 8080 6060
