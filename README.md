# transfer.sh [![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/dutchcoders/transfer.sh?utm_source=badge&utm_medium=badge&utm_campaign=&utm_campaign=pr-badge&utm_content=badge) [![Go Report Card](https://goreportcard.com/badge/github.com/dutchcoders/transfer.sh)](https://goreportcard.com/report/github.com/dutchcoders/transfer.sh) [![Docker pulls](https://img.shields.io/docker/pulls/dutchcoders/transfer.sh.svg)](https://hub.docker.com/r/dutchcoders/transfer.sh/) [![Build Status](https://github.com/dutchcoders/transfer.sh/actions/workflows/test.yml/badge.svg?branch=master)](https://github.com/dutchcoders/transfer.sh/actions/workflows/test.yml?query=branch%3Amaster)

Easy and fast file sharing from the command-line. This code contains the server with everything you need to create your own instance.

Transfer.sh currently supports the s3 (Amazon S3), gdrive (Google Drive), storj (Storj) providers, and local file system (local).

## Disclaimer

The service at https://transfersh.com is of unknown origin and reported as cloud malware.

## Usage

### Upload:
<!-- MARKDOWN-AUTO-DOCS:START (CODE:src=./examples/readme/upload.sh) -->
<!-- MARKDOWN-AUTO-DOCS:END -->

### Encrypt & upload:
<!-- MARKDOWN-AUTO-DOCS:START (CODE:src=./examples/readme/encrypt-and-upload.sh) -->
<!-- MARKDOWN-AUTO-DOCS:END -->

### Download & decrypt:
<!-- MARKDOWN-AUTO-DOCS:START (CODE:src=./examples/readme/download-and-decrypt.sh) -->
<!-- MARKDOWN-AUTO-DOCS:END -->

### Upload to virustotal:
<!-- MARKDOWN-AUTO-DOCS:START (CODE:src=./examples/readme/upload-to-virustotal.sh) -->
<!-- MARKDOWN-AUTO-DOCS:END -->

### Deleting
<!-- MARKDOWN-AUTO-DOCS:START (CODE:src=./examples/readme/deleting.sh) -->
<!-- MARKDOWN-AUTO-DOCS:END -->

## Request Headers

### Max-Downloads
<!-- MARKDOWN-AUTO-DOCS:START (CODE:src=./examples/readme/max-downloads.sh) -->
<!-- MARKDOWN-AUTO-DOCS:END -->

### Max-Days
<!-- MARKDOWN-AUTO-DOCS:START (CODE:src=./examples/readme/max-days.sh) -->
<!-- MARKDOWN-AUTO-DOCS:END -->

## Response Headers

### X-Url-Delete

The URL used to request the deletion of a file. Returned as a response header.
<!-- MARKDOWN-AUTO-DOCS:START (CODE:src=./examples/readme/x-url-delete.sh) -->
<!-- MARKDOWN-AUTO-DOCS:END -->

## Examples

See good usage examples on [examples.md](examples.md)

## Link aliases

Create direct download link:

https://transfer.sh/1lDau/test.txt --> https://transfer.sh/get/1lDau/test.txt

Inline file:

https://transfer.sh/1lDau/test.txt --> https://transfer.sh/inline/1lDau/test.txt

## Usage

<!-- MARKDOWN-AUTO-DOCS:START (JSON_TO_HTML_TABLE:src=./examples/readme/usage.json) -->
<!-- MARKDOWN-AUTO-DOCS:END -->   

If you want to use TLS using lets encrypt certificates, set lets-encrypt-hosts to your domain, set tls-listener to :443 and enable force-https.

If you want to use TLS using your own certificates, set tls-listener to :443, force-https, tls-cert-file and tls-private-key.

## Development

Switched to GO111MODULE
<!-- MARKDOWN-AUTO-DOCS:START (CODE:src=./examples/readme/dev.sh) -->
<!-- MARKDOWN-AUTO-DOCS:END -->

## Build
<!-- MARKDOWN-AUTO-DOCS:START (CODE:src=./examples/readme/build.sh) -->
<!-- MARKDOWN-AUTO-DOCS:END -->

## Docker

For easy deployment, we've created a Docker container.
<!-- MARKDOWN-AUTO-DOCS:START (CODE:src=./examples/readme/docker-run.sh) -->
<!-- MARKDOWN-AUTO-DOCS:END -->

## S3 Usage

For the usage with a AWS S3 Bucket, you just need to specify the following options:
- provider
- aws-access-key
- aws-secret-key
- bucket
- s3-region

If you specify the s3-region, you don't need to set the endpoint URL since the correct endpoint will used automatically.

### Custom S3 providers

To use a custom non-AWS S3 provider, you need to specify the endpoint as defined from your cloud provider.

## Storj Network Provider

To use the Storj Network as storage provider you need to specify the following flags:
- provider `--provider storj`
- storj-access _(either via flag or environment variable STORJ_ACCESS)_
- storj-bucket _(either via flag or environment variable STORJ_BUCKET)_

### Creating Bucket and Scope

In preparation you need to create an access grant (or copy it from the uplink configuration) and a bucket.

To get started, login to your account and go to the Access Grant Menu and start the Wizard on the upper right.

Enter your access grant name of choice, hit *Next* and restrict it as necessary/preferred.
Aftwards continue either in CLI or within the Browser. You'll be asked for a Passphrase used as Encryption Key.
**Make sure to save it in a safe place, without it you will lose the ability to decrypt your files!**

Afterwards you can copy the access grant and then start the startup of the transfer.sh endpoint. 
For enhanced security its recommended to provide both the access grant and the bucket name as ENV Variables.

Example:
```
export STORJ_BUCKET=<BUCKET NAME>
export STORJ_ACCESS=<ACCESS GRANT>
transfer.sh --provider storj
```

## Google Drive Usage

For the usage with Google drive, you need to specify the following options:
- provider
- gdrive-client-json-filepath
- gdrive-local-config-path
- basedir

### Creating Gdrive Client Json

You need to create a Oauth Client id from console.cloud.google.com
download the file and place into a safe directory

### Usage example
<!-- MARKDOWN-AUTO-DOCS:START (CODE:src=./examples/readme/usage-example.sh) -->
<!-- MARKDOWN-AUTO-DOCS:END -->

## Contributions

Contributions are welcome.

## Creators

**Remco Verhoef**
- <https://twitter.com/remco_verhoef>
- <https://twitter.com/dutchcoders>

**Uvis Grinfelds**

## Maintainer

**Andrea Spacca**

**Stefan Benten**

## Copyright and license

Code and documentation copyright 2011-2018 Remco Verhoef.
Code released under [the MIT license](LICENSE).
