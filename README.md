# transfer.sh

Easy and fast file sharing from the command-line. This code contains the server with everything you need to create your own instance.

Transfer.sh support currently the s3 (Amazon S3) provider and local file system (local).

[![Build Status](https://travis-ci.org/dutchcoders/transfer.sh.svg?branch=master)](https://travis-ci.org/dutchcoders/transfer.sh)

## Usage

```
Upload:
$ curl --upload-file ./hello.txt https://transfer.sh/hello.txt

Encrypt & upload:
$ cat /tmp/hello.txt|gpg -ac -o-|curl -X PUT --upload-file "-" https://transfer.sh/test.txt

Download & decrypt:
$ curl https://transfer.sh/1lDau/test.txt|gpg -o- > /tmp/hello.txt

Upload to virustotal:
$ curl -X PUT --upload-file nhgbhhj https://transfer.sh/test.txt/virustotal

Add alias to .bashrc or .zshrc:
===
transfer() {
    # write to output to tmpfile because of progress bar
    tmpfile=$( mktemp -t transferXXX )
    curl --progress-bar --upload-file $1 https://transfer.sh/$(basename $1) >> $tmpfile;
    cat $tmpfile;
    rm -f $tmpfile;
}

alias transfer=transfer
===
$ transfer test.txt
```

## Development

```
npm install
bower install

go get github.com/PuerkitoBio/ghost/handlers
go get github.com/gorilla/mux
go get github.com/dutchcoders/go-clamd
go get github.com/goamz/goamz/s3
go get github.com/goamz/goamz/aws
go get github.com/golang/gddo/httputil/header
go get github.com/kennygrant/sanitize
go get github.com/dutchcoders/go-virustotal
go get github.com/russross/blackfriday

grunt serve
grunt build

go run transfersh-server/*.go -provider=local --port 8080 --temp=/tmp/ --basedir=/tmp/ 
```

## Build

```
go build -o transfersh-server *.go
```

## Docker

For easy deployment we've enabled Docker deployment.

```
cd ./transfer-server/
docker build -t transfersh .
docker run --publish 8080:8080 --rm transfersh --provider local --basedir /tmp/
```

## Contributions

Contributions are welcome.

## Creators 

**Remco Verhoef**
- <https://twitter.com/remco_verhoef>
- <https://twitter.com/dutchcoders>

**Uvis Grinfelds**

## Copyright and license

Code and documentation copyright 2011-2014 Remco Verhoef. 
Code released under [the MIT license](LICENSE). 
