# transfer.sh [![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/dutchcoders/transfer.sh?utm_source=badge&utm_medium=badge&utm_campaign=&utm_campaign=pr-badge&utm_content=badge) [![Go Report Card](https://goreportcard.com/badge/github.com/dutchcoders/transfer.sh)](https://goreportcard.com/report/github.com/dutchcoders/transfer.sh) [![Docker pulls](https://img.shields.io/docker/pulls/dutchcoders/transfer.sh.svg)](https://hub.docker.com/r/dutchcoders/transfer.sh/) [![Build Status](https://travis-ci.org/dutchcoders/transfer.sh.svg?branch=master)](https://travis-ci.org/dutchcoders/transfer.sh)

Easy and fast file sharing from the command-line. This code contains the server with everything you need to create your own instance.

Transfer.sh support currently the s3 (Amazon S3) provider and local file system (local).

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

## Usage

Parameter | Description | Value | Env
--- | --- | --- | ---
listener | port to use for http (:80) | |
profile-listener | port to use for profiler (:6060)| |
force-https | redirect to https | false |
tls-listener | port to use for https (:443) | |
tls-cert-file | path to tls certificate | | 
tls-private-key | path to tls private key | |
temp-path | path to temp folder | system temp |
web-path | path to static web files (for development) | |
provider | which storage provider to use | (s3 or local) |
aws-access-key | aws access key | | AWS_ACCESS_KEY
aws-secret-key | aws access key | | AWS_SECRET_KEY
bucket | aws bucket | | BUCKET
basedir | path storage for local provider| | 
lets-encrypt-hosts | hosts to use for lets encrypt certificates (comma seperated) | | 
log | path to log file| | 

If you want to use TLS using lets encrypt certificates, set lets-encrypt-hosts to your domain, set tls-listener to :443 and enable force-https.

If you want to use TLS using your own certificates, set tls-listener to :443, force-https, tls-cert=file and tls-private-key.

## Development

Make sure your GOPATH is set correctly.

```
go run main.go -provider=local --listener :8080 --temp-path=/tmp/ --basedir=/tmp/ 
```

## Build

```
go build -o transfersh main.go
```

## Docker

For easy deployment we've created a Docker container.

```
docker run --publish 8080:8080 dutchcoders/transfer.sh:latest --provider local --basedir /tmp/
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
