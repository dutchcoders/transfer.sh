FROM golang:latest

# Install redis, nginx, daemontools, etc.
RUN echo deb http://http.debian.net/debian wheezy-backports main > /etc/apt/sources.list.d/backports.list && \
	apt-get update && \
	apt-get install -y --no-install-recommends -t wheezy-backports redis-server && \
	apt-get install -y --no-install-recommends graphviz nginx-full daemontools unzip

# Configure redis.
ADD deploy/redis.conf /etc/redis/redis.conf

# Configure nginx.
RUN echo "daemon off;" >> /etc/nginx/nginx.conf && \
	rm /etc/nginx/sites-enabled/default
ADD deploy/gddo.conf /etc/nginx/sites-enabled/gddo.conf

# Configure daemontools services.
ADD deploy/services /services

# Manually fetch and install gddo-server dependencies (faster than "go get").
ADD https://github.com/garyburd/redigo/archive/779af66db5668074a96f522d9025cb0a5ef50d89.zip /x/redigo.zip
ADD https://github.com/golang/snappy/archive/master.zip /x/snappy-go.zip
RUN unzip /x/redigo.zip -d /x && unzip /x/snappy-go.zip -d /x && \
	mkdir -p /go/src/github.com/garyburd && \
	mkdir -p /go/src/github.com/golang && \
	mv /x/redigo-* /go/src/github.com/garyburd/redigo && \
	mv /x/snappy-master /go/src/github.com/golang/snappy && \
	rm -rf /x

# Build the local gddo files.
ADD . /go/src/github.com/golang/gddo
RUN go get github.com/golang/gddo/gddo-server

# Exposed ports and volumes.
# /ssl should contain SSL certs.
# /data should contain the Redis database, "dump.rdb".
EXPOSE 80 443
VOLUME ["/ssl", "/data"]

# How to start it all.
CMD svscan /services
