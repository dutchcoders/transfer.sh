# transfer.sh [![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/dutchcoders/transfer.sh?utm_source=badge&utm_medium=badge&utm_campaign=&utm_campaign=pr-badge&utm_content=badge) [![Go Report Card](https://goreportcard.com/badge/github.com/dutchcoders/transfer.sh)](https://goreportcard.com/report/github.com/dutchcoders/transfer.sh) [![Docker pulls](https://img.shields.io/docker/pulls/dutchcoders/transfer.sh.svg)](https://hub.docker.com/r/dutchcoders/transfer.sh/) [![Build Status](https://travis-ci.org/dutchcoders/transfer.sh.svg?branch=master)](https://travis-ci.org/dutchcoders/transfer.sh)

Easy and fast file sharing from the command-line. This code contains the server with everything you need to create your own instance.

Transfer.sh currently supports the s3 (Amazon S3), gdrive (Google Drive) providers, and local file system (local).

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

```
## Add alias to .bashrc or .zshrc
```
transfer() {
    curl --progress-bar --upload-file $1 https://transfer.sh/$(basename $1) | tee /dev/null;
}

alias transfer=transfer
```
Now run it like this
```
$ transfer test.txt
```

### On Windows

Put a file called transfer.cmd somewhere in your PATH with this inside it:
```
@echo off
setlocal
:: use env vars to pass names to PS, to avoid escaping issues
set FN=%~nx1
set FULL=%1
powershell -noprofile -command "$(Invoke-Webrequest -Method put -Infile $Env:FULL https://transfer.sh/$Env:FN).Content"
```

## Usage

Parameter | Description | Value | Env
--- | --- | --- | ---
listener | port to use for http (:80) | |
profile-listener | port to use for profiler (:6060)| |
force-https | redirect to https | false |
tls-listener | port to use for https (:443) | |
tls-listener-only | flag to enable tls listener only | |
tls-cert-file | path to tls certificate | |
tls-private-key | path to tls private key | |
http-auth-user | user for basic http auth on upload | |
http-auth-pass | pass for basic http auth on upload | |
temp-path | path to temp folder | system temp |
web-path | path to static web files (for development) | |
ga-key | google analytics key for the front end | |
uservoice-key | user voice key for the front end  | |
provider | which storage provider to use | (s3, grdrive or local) |
aws-access-key | aws access key | | AWS_ACCESS_KEY
aws-secret-key | aws access key | | AWS_SECRET_KEY
bucket | aws bucket | | BUCKET
basedir | path storage for local/gdrive provider| |
gdrive-client-json-filepath | path to client json config for gdrive provider| |
gdrive-local-config-path | path to local transfer.sh config cache for gdrive provider| |
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

For easy deployment, we've created a Docker container.

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
