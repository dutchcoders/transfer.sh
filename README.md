# transfer.sh [![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/dutchcoders/transfer.sh?utm_source=badge&utm_medium=badge&utm_campaign=&utm_campaign=pr-badge&utm_content=badge) [![Go Report Card](https://goreportcard.com/badge/github.com/dutchcoders/transfer.sh)](https://goreportcard.com/report/github.com/dutchcoders/transfer.sh) [![Docker pulls](https://img.shields.io/docker/pulls/dutchcoders/transfer.sh.svg)](https://hub.docker.com/r/dutchcoders/transfer.sh/) [![Build Status](https://github.com/dutchcoders/transfer.sh/actions/workflows/test.yml/badge.svg?branch=master)](https://github.com/dutchcoders/transfer.sh/actions/workflows/test.yml?query=branch%3Amaster)

Easy and fast file sharing from the command-line. This code contains the server with everything you need to create your own instance.

Transfer.sh currently supports the s3 (Amazon S3), gdrive (Google Drive), storj (Storj) providers, and local file system (local).

## Disclaimer

The service at https://transfersh.com is of unknown origin and reported as cloud malware.

## Usage

### Upload:
<!-- MARKDOWN-AUTO-DOCS:START (CODE:src=./examples/readme/upload.sh) -->
<!-- The below code snippet is automatically added from ./examples/readme/upload.sh -->
```sh
curl --upload-file ./hello.txt https://transfer.sh/hello.txt
```
<!-- MARKDOWN-AUTO-DOCS:END -->

### Encrypt & upload:
<!-- MARKDOWN-AUTO-DOCS:START (CODE:src=./examples/readme/encrypt-and-upload.sh) -->
<!-- The below code snippet is automatically added from ./examples/readme/encrypt-and-upload.sh -->
```sh
cat /tmp/hello.txt|gpg -ac -o-|curl -X PUT --upload-file "-" https://transfer.sh/test.txt
```
<!-- MARKDOWN-AUTO-DOCS:END -->

### Download & decrypt:
<!-- MARKDOWN-AUTO-DOCS:START (CODE:src=./examples/readme/download-and-decrypt.sh) -->
<!-- The below code snippet is automatically added from ./examples/readme/download-and-decrypt.sh -->
```sh
curl https://transfer.sh/1lDau/test.txt|gpg -o- > /tmp/hello.txt
```
<!-- MARKDOWN-AUTO-DOCS:END -->

### Upload to virustotal:
<!-- MARKDOWN-AUTO-DOCS:START (CODE:src=./examples/readme/upload-to-virustotal.sh) -->
<!-- The below code snippet is automatically added from ./examples/readme/upload-to-virustotal.sh -->
```sh
curl -X PUT --upload-file nhgbhhj https://transfer.sh/test.txt/virustotal
```
<!-- MARKDOWN-AUTO-DOCS:END -->

### Deleting
<!-- MARKDOWN-AUTO-DOCS:START (CODE:src=./examples/readme/deleting.sh) -->
<!-- The below code snippet is automatically added from ./examples/readme/deleting.sh -->
```sh
curl -X DELETE <X-Url-Delete Response Header URL>
```
<!-- MARKDOWN-AUTO-DOCS:END -->

## Request Headers

### Max-Downloads
<!-- MARKDOWN-AUTO-DOCS:START (CODE:src=./examples/readme/max-downloads.sh) -->
<!-- The below code snippet is automatically added from ./examples/readme/max-downloads.sh -->
```sh
curl --upload-file ./hello.txt https://transfer.sh/hello.txt -H "Max-Downloads: 1" # Limit the number of downloads
```
<!-- MARKDOWN-AUTO-DOCS:END -->

### Max-Days
<!-- MARKDOWN-AUTO-DOCS:START (CODE:src=./examples/readme/max-days.sh) -->
<!-- The below code snippet is automatically added from ./examples/readme/max-days.sh -->
```sh
curl --upload-file ./hello.txt https://transfer.sh/hello.txt -H "Max-Days: 1" # Set the number of days before deletion
```
<!-- MARKDOWN-AUTO-DOCS:END -->

## Response Headers

### X-Url-Delete

The URL used to request the deletion of a file. Returned as a response header.
<!-- MARKDOWN-AUTO-DOCS:START (CODE:src=./examples/readme/x-url-delete.sh) -->
<!-- The below code snippet is automatically added from ./examples/readme/x-url-delete.sh -->
```sh
curl -sD - --upload-file ./hello https://transfer.sh/hello.txt | grep 'X-Url-Delete'
X-Url-Delete: https://transfer.sh/hello.txt/BAYh0/hello.txt/PDw0NHPcqU
```
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
<table class="JSON-TO-HTML-TABLE"><thead><tr><th class="parameter-th">Parameter</th><th class="description-th">Description</th><th class="value-th">Value</th><th class="env-th">Env</th></tr></thead><tbody ><tr ><td class="parameter-td td_text">listener</td><td class="description-td td_text">port to use for http (:80)</td><td class="value-td td_num"></td><td class="env-td td_text">LISTENER</td></tr>
<tr ><td class="parameter-td td_text">profile-listener</td><td class="description-td td_text">port to use for profiler (:6060)</td><td class="value-td td_num"></td><td class="env-td td_text">PROFILE_LISTENER</td></tr>
<tr ><td class="parameter-td td_text">force-https</td><td class="description-td td_text">redirect to https</td><td class="value-td td_text">false</td><td class="env-td td_text">FORCE_HTTPS</td></tr>
<tr ><td class="parameter-td td_text">tls-listener</td><td class="description-td td_text">port to use for https (:443)</td><td class="value-td td_num"></td><td class="env-td td_text">TLS_LISTENER</td></tr>
<tr ><td class="parameter-td td_text">tls-listener-only</td><td class="description-td td_text">flag to enable tls listener only</td><td class="value-td td_num"></td><td class="env-td td_text">TLS_LISTENER_ONLY</td></tr>
<tr ><td class="parameter-td td_text">tls-cert-file</td><td class="description-td td_text">path to tls certificate</td><td class="value-td td_num"></td><td class="env-td td_text">TLS_CERT_FILE</td></tr>
<tr ><td class="parameter-td td_text">tls-private-key</td><td class="description-td td_text">path to tls private key</td><td class="value-td td_num"></td><td class="env-td td_text">TLS_PRIVATE_KEY</td></tr>
<tr ><td class="parameter-td td_text">http-auth-user</td><td class="description-td td_text">user for basic http auth on upload</td><td class="value-td td_num"></td><td class="env-td td_text">HTTP_AUTH_USER</td></tr>
<tr ><td class="parameter-td td_text">http-auth-pass</td><td class="description-td td_text">pass for basic http auth on upload</td><td class="value-td td_num"></td><td class="env-td td_text">HTTP_AUTH_PASS</td></tr>
<tr ><td class="parameter-td td_text">ip-whitelist</td><td class="description-td td_text">comma separated list of ips allowed to connect to the service</td><td class="value-td td_num"></td><td class="env-td td_text">IP_WHITELIST</td></tr>
<tr ><td class="parameter-td td_text">ip-blacklist</td><td class="description-td td_text">comma separated list of ips not allowed to connect to the service</td><td class="value-td td_num"></td><td class="env-td td_text">IP_BLACKLIST</td></tr>
<tr ><td class="parameter-td td_text">temp-path</td><td class="description-td td_text">path to temp folder</td><td class="value-td td_text">system temp</td><td class="env-td td_text">TEMP_PATH</td></tr>
<tr ><td class="parameter-td td_text">web-path</td><td class="description-td td_text">path to static web files (for development or custom front end)</td><td class="value-td td_num"></td><td class="env-td td_text">WEB_PATH</td></tr>
<tr ><td class="parameter-td td_text">proxy-path</td><td class="description-td td_text">path prefix when service is run behind a proxy</td><td class="value-td td_num"></td><td class="env-td td_text">PROXY_PATH</td></tr>
<tr ><td class="parameter-td td_text">proxy-port</td><td class="description-td td_text">port of the proxy when the service is run behind a proxy</td><td class="value-td td_num"></td><td class="env-td td_text">PROXY_PORT</td></tr>
<tr ><td class="parameter-td td_text">ga-key</td><td class="description-td td_text">google analytics key for the front end</td><td class="value-td td_num"></td><td class="env-td td_text">GA_KEY</td></tr>
<tr ><td class="parameter-td td_text">provider</td><td class="description-td td_text">which storage provider to use</td><td class="value-td td_text">(s3, storj, gdrive or local)</td><td class="env-td td_num"></td></tr>
<tr ><td class="parameter-td td_text">uservoice-key</td><td class="description-td td_text">user voice key for the front end</td><td class="value-td td_num"></td><td class="env-td td_text">USERVOICE_KEY</td></tr>
<tr ><td class="parameter-td td_text">aws-access-key</td><td class="description-td td_text">aws access key</td><td class="value-td td_num"></td><td class="env-td td_text">AWS_ACCESS_KEY</td></tr>
<tr ><td class="parameter-td td_text">aws-secret-key</td><td class="description-td td_text">aws access key</td><td class="value-td td_num"></td><td class="env-td td_text">AWS_SECRET_KEY</td></tr>
<tr ><td class="parameter-td td_text">bucket</td><td class="description-td td_text">aws bucket</td><td class="value-td td_num"></td><td class="env-td td_text">BUCKET</td></tr>
<tr ><td class="parameter-td td_text">s3-endpoint</td><td class="description-td td_text">Custom S3 endpoint.</td><td class="value-td td_num"></td><td class="env-td td_text">S3_ENDPOINT</td></tr>
<tr ><td class="parameter-td td_text">s3-region</td><td class="description-td td_text">region of the s3 bucket</td><td class="value-td td_text">eu-west-1</td><td class="env-td td_text">S3_REGION</td></tr>
<tr ><td class="parameter-td td_text">s3-no-multipart</td><td class="description-td td_text">disables s3 multipart upload</td><td class="value-td td_text">false</td><td class="env-td td_text">S3_NO_MULTIPART</td></tr>
<tr ><td class="parameter-td td_text">s3-path-style</td><td class="description-td td_text">Forces path style URLs, required for Minio.</td><td class="value-td td_text">false</td><td class="env-td td_text">S3_PATH_STYLE</td></tr>
<tr ><td class="parameter-td td_text">storj-access</td><td class="description-td td_text">Access for the project</td><td class="value-td td_num"></td><td class="env-td td_text">STORJ_ACCESS</td></tr>
<tr ><td class="parameter-td td_text">storj-bucket</td><td class="description-td td_text">Bucket to use within the project</td><td class="value-td td_num"></td><td class="env-td td_text">STORJ_BUCKET</td></tr>
<tr ><td class="parameter-td td_text">basedir</td><td class="description-td td_text">path storage for local/gdrive provider</td><td class="value-td td_num"></td><td class="env-td td_text">BASEDIR</td></tr>
<tr ><td class="parameter-td td_text">gdrive-client-json-filepath</td><td class="description-td td_text">path to oauth client json config for gdrive provider</td><td class="value-td td_num"></td><td class="env-td td_text">GDRIVE_CLIENT_JSON_FILEPATH</td></tr>
<tr ><td class="parameter-td td_text">gdrive-local-config-path</td><td class="description-td td_text">path to store local transfer.sh config cache for gdrive provider</td><td class="value-td td_num"></td><td class="env-td td_text">GDRIVE_LOCAL_CONFIG_PATH</td></tr>
<tr ><td class="parameter-td td_text">gdrive-chunk-size</td><td class="description-td td_text">chunk size for gdrive upload in megabytes, must be lower than available memory (8 MB)</td><td class="value-td td_num"></td><td class="env-td td_text">GDRIVE_CHUNK_SIZE</td></tr>
<tr ><td class="parameter-td td_text">lets-encrypt-hosts</td><td class="description-td td_text">hosts to use for lets encrypt certificates (comma seperated)</td><td class="value-td td_num"></td><td class="env-td td_text">HOSTS</td></tr>
<tr ><td class="parameter-td td_text">log</td><td class="description-td td_text">path to log file</td><td class="value-td td_num"></td><td class="env-td td_text">LOG</td></tr>
<tr ><td class="parameter-td td_text">cors-domains</td><td class="description-td td_text">comma separated list of domains for CORS, setting it enable CORS</td><td class="value-td td_num"></td><td class="env-td td_text">CORS_DOMAINS</td></tr>
<tr ><td class="parameter-td td_text">clamav-host</td><td class="description-td td_text">host for clamav feature</td><td class="value-td td_num"></td><td class="env-td td_text">CLAMAV_HOST</td></tr>
<tr ><td class="parameter-td td_text">rate-limit</td><td class="description-td td_text">request per minute</td><td class="value-td td_num"></td><td class="env-td td_text">RATE_LIMIT</td></tr>
<tr ><td class="parameter-td td_text">max-upload-size</td><td class="description-td td_text">max upload size in kilobytes</td><td class="value-td td_num"></td><td class="env-td td_text">MAX_UPLOAD_SIZE</td></tr>
<tr ><td class="parameter-td td_text">purge-days</td><td class="description-td td_text">number of days after the uploads are purged automatically</td><td class="value-td td_num"></td><td class="env-td td_text">PURGE_DAYS</td></tr>
<tr ><td class="parameter-td td_text">purge-interval</td><td class="description-td td_text">interval in hours to run the automatic purge for (not applicable to S3 and Storj)</td><td class="value-td td_num"></td><td class="env-td td_text">PURGE_INTERVAL</td></tr>
<tr ><td class="parameter-td td_text">random-token-length</td><td class="description-td td_text">length of the random token for the upload path (double the size for delete path)</td><td class="value-td td_num">6</td><td class="env-td td_text">RANDOM_TOKEN_LENGTH</td></tr></tbody></table>
<!-- MARKDOWN-AUTO-DOCS:END -->   

If you want to use TLS using lets encrypt certificates, set lets-encrypt-hosts to your domain, set tls-listener to :443 and enable force-https.

If you want to use TLS using your own certificates, set tls-listener to :443, force-https, tls-cert-file and tls-private-key.

## Development

Switched to GO111MODULE
<!-- MARKDOWN-AUTO-DOCS:START (CODE:src=./examples/readme/dev.sh) -->
<!-- The below code snippet is automatically added from ./examples/readme/dev.sh -->
```sh
go run main.go --provider=local --listener :8080 --temp-path=/tmp/ --basedir=/tmp/
```
<!-- MARKDOWN-AUTO-DOCS:END -->

## Build
<!-- MARKDOWN-AUTO-DOCS:START (CODE:src=./examples/readme/build.sh) -->
<!-- The below code snippet is automatically added from ./examples/readme/build.sh -->
```sh
git clone git@github.com:dutchcoders/transfer.sh.git
cd transfer.sh
go build -o transfersh main.go
```
<!-- MARKDOWN-AUTO-DOCS:END -->

## Docker

For easy deployment, we've created a Docker container.
<!-- MARKDOWN-AUTO-DOCS:START (CODE:src=./examples/readme/docker-run.sh) -->
<!-- The below code snippet is automatically added from ./examples/readme/docker-run.sh -->
```sh
docker run --publish 8080:8080 dutchcoders/transfer.sh:latest --provider local --basedir /tmp/
```
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
<!-- The below code snippet is automatically added from ./examples/readme/usage-example.sh -->
```sh
go run main.go --provider gdrive --basedir /tmp/ --gdrive-client-json-filepath /[credential_dir] --gdrive-local-config-path [directory_to_save_config]
```
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
