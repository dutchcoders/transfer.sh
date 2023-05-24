# transfer.sh [![Go Report Card](https://goreportcard.com/badge/github.com/dutchcoders/transfer.sh)](https://goreportcard.com/report/github.com/dutchcoders/transfer.sh) [![Docker pulls](https://img.shields.io/docker/pulls/dutchcoders/transfer.sh.svg)](https://hub.docker.com/r/dutchcoders/transfer.sh/) [![Build Status](https://github.com/dutchcoders/transfer.sh/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/dutchcoders/transfer.sh/actions/workflows/test.yml?query=branch%3Amain)

Easy and fast file sharing from the command-line. This code contains the server with everything you need to create your own instance.

Transfer.sh currently supports the s3 (Amazon S3), gdrive (Google Drive), storj (Storj) providers, and local file system (local).

## Disclaimer

The service at transfersh.com is of unknown origin and reported as cloud malware.

## Usage

### Upload:
```bash
$ curl -v --upload-file ./hello.txt https://transfer.sh/hello.txt
```

### Encrypt & Upload:
```bash
$ cat /tmp/hello.txt|gpg -ac -o-|curl -X PUT --upload-file "-" https://transfer.sh/test.txt
````

### Download & Decrypt:
```bash
$ curl https://transfer.sh/1lDau/test.txt|gpg -o- > /tmp/hello.txt
```

### Upload to Virustotal:
```bash
$ curl -X PUT --upload-file nhgbhhj https://transfer.sh/test.txt/virustotal
```

### Deleting
```bash
$ curl -X DELETE <X-Url-Delete Response Header URL>
```

## Request Headers

### Max-Downloads
```bash
$ curl --upload-file ./hello.txt https://transfer.sh/hello.txt -H "Max-Downloads: 1" # Limit the number of downloads
```

### Max-Days
```bash
$ curl --upload-file ./hello.txt https://transfer.sh/hello.txt -H "Max-Days: 1" # Set the number of days before deletion
```

### X-Encrypt-Password
#### Beware, use this feature only on your self-hosted server: trusting a third-party service for server side encryption is at your own risk
```bash
$ curl --upload-file ./hello.txt https://your-transfersh-instance.tld/hello.txt -H "X-Encrypt-Password: test" # Encrypt the content sever side with AES265 using "test" as password
```

### X-Decrypt-Password
#### Beware, use this feature only on your self-hosted server: trusting a third-party service for server side encryption is at your own risk
```bash
$ curl https://your-transfersh-instance.tld/BAYh0/hello.txt -H "X-Decrypt-Password: test" # Decrypt the content sever side with AES265 using "test" as password
```

## Response Headers

### X-Url-Delete

The URL used to request the deletion of a file and returned as a response header.
```bash
curl -sD - --upload-file ./hello.txt https://transfer.sh/hello.txt | grep -i -E 'transfer\.sh|x-url-delete'
x-url-delete: https://transfer.sh/hello.txt/BAYh0/hello.txt/PDw0NHPcqU
https://transfer.sh/hello.txt/BAYh0/hello.txt
```

## Examples

See good usage examples on [examples.md](examples.md)

## Link aliases

Create direct download link:

https://transfer.sh/1lDau/test.txt --> https://transfer.sh/get/1lDau/test.txt

Inline file:

https://transfer.sh/1lDau/test.txt --> https://transfer.sh/inline/1lDau/test.txt

## Usage

Parameter | Description                                                                                 | Value                        | Env                         
--- |---------------------------------------------------------------------------------------------|------------------------------|-----------------------------
listener | port to use for http (:80)                                                                  |                              | LISTENER                    |
profile-listener | port to use for profiler (:6060)                                                            |                              | PROFILE_LISTENER            |
force-https | redirect to https                                                                           | false                        | FORCE_HTTPS                 
tls-listener | port to use for https (:443)                                                                |                              | TLS_LISTENER                |
tls-listener-only | flag to enable tls listener only                                                            |                              | TLS_LISTENER_ONLY           |
tls-cert-file | path to tls certificate                                                                     |                              | TLS_CERT_FILE               |
tls-private-key | path to tls private key                                                                     |                              | TLS_PRIVATE_KEY             |
http-auth-user | user for basic http auth on upload                                                          |                              | HTTP_AUTH_USER              |
http-auth-pass | pass for basic http auth on upload                                                          |                              | HTTP_AUTH_PASS              |
http-auth-htpasswd | htpasswd file path for basic http auth on upload                                            |                              | HTTP_AUTH_HTPASSWD          |
http-auth-ip-whitelist | comma separated list of ips allowed to upload without being challenged an http auth        |                              | HTTP_AUTH_IP_WHITELIST      |
ip-whitelist | comma separated list of ips allowed to connect to the service                               |                              | IP_WHITELIST                |
ip-blacklist | comma separated list of ips not allowed to connect to the service                           |                              | IP_BLACKLIST                |
temp-path | path to temp folder                                                                         | system temp                  | TEMP_PATH                   |
web-path | path to static web files (for development or custom front end)                              |                              | WEB_PATH                    |
proxy-path | path prefix when service is run behind a proxy                                              |                              | PROXY_PATH                  |
proxy-port | port of the proxy when the service is run behind a proxy                                    |                              | PROXY_PORT                  |
email-contact | email contact for the front end                                                             |                              | EMAIL_CONTACT               |
ga-key | google analytics key for the front end                                                      |                              | GA_KEY                      |
provider | which storage provider to use                                                               | (s3, storj, gdrive or local) |
uservoice-key | user voice key for the front end                                                            |                              | USERVOICE_KEY               |
aws-access-key | aws access key                                                                              |                              | AWS_ACCESS_KEY              |
aws-secret-key | aws access key                                                                              |                              | AWS_SECRET_KEY              |
bucket | aws bucket                                                                                  |                              | BUCKET                      |
s3-endpoint | Custom S3 endpoint.                                                                         |                              | S3_ENDPOINT                 |
s3-region | region of the s3 bucket                                                                     | eu-west-1                    | S3_REGION                   |
s3-no-multipart | disables s3 multipart upload                                                                | false                        | S3_NO_MULTIPART             |
s3-path-style | Forces path style URLs, required for Minio.                                                 | false                        | S3_PATH_STYLE               |
storj-access | Access for the project                                                                      |                              | STORJ_ACCESS                |
storj-bucket | Bucket to use within the project                                                            |                              | STORJ_BUCKET                |
basedir | path storage for local/gdrive provider                                                      |                              | BASEDIR                     |
gdrive-client-json-filepath | path to oauth client json config for gdrive provider                                        |                              | GDRIVE_CLIENT_JSON_FILEPATH |
gdrive-local-config-path | path to store local transfer.sh config cache for gdrive provider                            |                              | GDRIVE_LOCAL_CONFIG_PATH    |
gdrive-chunk-size | chunk size for gdrive upload in megabytes, must be lower than available memory (8 MB)       |                              | GDRIVE_CHUNK_SIZE           |
lets-encrypt-hosts | hosts to use for lets encrypt certificates (comma seperated)                                |                              | HOSTS                       |
log | path to log file                                                                            |                              | LOG                         |
cors-domains | comma separated list of domains for CORS, setting it enable CORS                            |                              | CORS_DOMAINS                |
clamav-host | host for clamav feature                                                                     |                              | CLAMAV_HOST                 |
perform-clamav-prescan | prescan every upload through clamav feature (clamav-host must be a local clamd unix socket) |                              | PERFORM_CLAMAV_PRESCAN      |
rate-limit | request per minute                                                                          |                              | RATE_LIMIT                  |
max-upload-size | max upload size in kilobytes                                                                |                              | MAX_UPLOAD_SIZE             |
purge-days | number of days after the uploads are purged automatically                                   |                              | PURGE_DAYS                  |   
purge-interval | interval in hours to run the automatic purge for (not applicable to S3 and Storj)           |                              | PURGE_INTERVAL              |   
random-token-length | length of the random token for the upload path (double the size for delete path)            | 6                            | RANDOM_TOKEN_LENGTH         |   

If you want to use TLS using lets encrypt certificates, set lets-encrypt-hosts to your domain, set tls-listener to :443 and enable force-https.

If you want to use TLS using your own certificates, set tls-listener to :443, force-https, tls-cert-file and tls-private-key.

## Development

Switched to GO111MODULE

```bash
go run main.go --provider=local --listener :8080 --temp-path=/tmp/ --basedir=/tmp/
```

## Build

```bash
$ git clone git@github.com:dutchcoders/transfer.sh.git
$ cd transfer.sh
$ go build -o transfersh main.go
```

## Docker

For easy deployment, we've created an official Docker container. There are two variants, differing only by which user runs the process.

The default one will run as `root`:

```bash
docker run --publish 8080:8080 dutchcoders/transfer.sh:latest --provider local --basedir /tmp/
```

The one tagged with the suffix `-noroot` will use `5000` as both UID and GID:
```bash
docker run --publish 8080:8080 dutchcoders/transfer.sh:latest-noroot --provider local --basedir /tmp/
```

### Building the Container
You can also build the container yourself. This allows you to choose which UID/GID will be used, e.g. when using NFS mounts:
```bash
# Build arguments:
# * RUNAS: If empty, the container will run as root.
#          Set this to anything to enable UID/GID selection.
# * PUID:  UID of the process. Needs RUNAS != "". Defaults to 5000.
# * PGID:  GID of the process. Needs RUNAS != "". Defaults to 5000.

docker build -t transfer.sh-noroot --build-arg RUNAS=doesntmatter --build-arg PUID=1337 --build-arg PGID=1338 .
```

## S3 Usage

For the usage with a AWS S3 Bucket, you just need to specify the following options:
- provider `--provider s3`
- aws-access-key _(either via flag or environment variable `AWS_ACCESS_KEY`)_
- aws-secret-key _(either via flag or environment variable `AWS_SECRET_KEY`)_
- bucket _(either via flag or environment variable `BUCKET`)_
- s3-region _(either via flag or environment variable `S3_REGION`)_

If you specify the s3-region, you don't need to set the endpoint URL since the correct endpoint will used automatically.

### Custom S3 providers

To use a custom non-AWS S3 provider, you need to specify the endpoint as defined from your cloud provider.

## Storj Network Provider

To use the Storj Network as a storage provider you need to specify the following flags:
- provider `--provider storj`
- storj-access _(either via flag or environment variable STORJ_ACCESS)_
- storj-bucket _(either via flag or environment variable STORJ_BUCKET)_

### Creating Bucket and Scope

You need to create an access grant (or copy it from the uplink configuration) and a bucket in preparation.

To get started, log in to your account and go to the Access Grant Menu and start the Wizard on the upper right.

Enter your access grant name of choice, hit *Next* and restrict it as necessary/preferred.
Afterwards continue either in CLI or within the Browser. Next, you'll be asked for a Passphrase used as Encryption Key.
**Make sure to save it in a safe place. Without it, you will lose the ability to decrypt your files!**

Afterwards, you can copy the access grant and then start the startup of the transfer.sh endpoint. 
It is recommended to provide both the access grant and the bucket name as ENV Variables for enhanced security.

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

You need to create an OAuth Client id from console.cloud.google.com, download the file, and place it into a safe directory.

### Usage example

```go run main.go --provider gdrive --basedir /tmp/ --gdrive-client-json-filepath /[credential_dir] --gdrive-local-config-path [directory_to_save_config] ```

## Shell functions

### Bash and zsh (multiple files uploaded as zip archive)
##### Add this to .bashrc or .zshrc or its equivalent
```bash
transfer(){ if [ $# -eq 0 ];then echo "No arguments specified.\nUsage:\n transfer <file|directory>\n ... | transfer <file_name>">&2;return 1;fi;if tty -s;then file="$1";file_name=$(basename "$file");if [ ! -e "$file" ];then echo "$file: No such file or directory">&2;return 1;fi;if [ -d "$file" ];then file_name="$file_name.zip" ,;(cd "$file"&&zip -r -q - .)|curl --progress-bar --upload-file "-" "https://transfer.sh/$file_name"|tee /dev/null,;else cat "$file"|curl --progress-bar --upload-file "-" "https://transfer.sh/$file_name"|tee /dev/null;fi;else file_name=$1;curl --progress-bar --upload-file "-" "https://transfer.sh/$file_name"|tee /dev/null;fi;}
```

#### Now you can use transfer function
```
$ transfer hello.txt
```


### Bash and zsh (with delete url, delete token output and prompt before uploading)
##### Add this to .bashrc or .zshrc or its equivalent

<details><summary>Expand</summary><p>

```bash
transfer()
{
    local file
    declare -a file_array
    file_array=("${@}")

    if [[ "${file_array[@]}" == "" || "${1}" == "--help" || "${1}" == "-h" ]]
    then
        echo "${0} - Upload arbitrary files to \"transfer.sh\"."
        echo ""
        echo "Usage: ${0} [options] [<file>]..."
        echo ""
        echo "OPTIONS:"
        echo "  -h, --help"
        echo "      show this message"
        echo ""
        echo "EXAMPLES:"
        echo "  Upload a single file from the current working directory:"
        echo "      ${0} \"image.img\""
        echo ""
        echo "  Upload multiple files from the current working directory:"
        echo "      ${0} \"image.img\" \"image2.img\""
        echo ""
        echo "  Upload a file from a different directory:"
        echo "      ${0} \"/tmp/some_file\""
        echo ""
        echo "  Upload all files from the current working directory. Be aware of the webserver's rate limiting!:"
        echo "      ${0} *"
        echo ""
        echo "  Upload a single file from the current working directory and filter out the delete token and download link:"
        echo "      ${0} \"image.img\" | awk --field-separator=\": \" '/Delete token:/ { print \$2 } /Download link:/ { print \$2 }'"
        echo ""
        echo "  Show help text from \"transfer.sh\":"
        echo "      curl --request GET \"https://transfer.sh\""
        return 0
    else
        for file in "${file_array[@]}"
        do
            if [[ ! -f "${file}" ]]
            then
                echo -e "\e[01;31m'${file}' could not be found or is not a file.\e[0m" >&2
                return 1
            fi
        done
        unset file
    fi

    local upload_files
    local curl_output
    local awk_output

    du -c -k -L "${file_array[@]}" >&2
    # be compatible with "bash"
    if [[ "${ZSH_NAME}" == "zsh" ]]
    then
        read $'upload_files?\e[01;31mDo you really want to upload the above files ('"${#file_array[@]}"$') to "transfer.sh"? (Y/n): \e[0m'
    elif [[ "${BASH}" == *"bash"* ]]
    then
        read -p $'\e[01;31mDo you really want to upload the above files ('"${#file_array[@]}"$') to "transfer.sh"? (Y/n): \e[0m' upload_files
    fi

    case "${upload_files:-y}" in
        "y"|"Y")
            # for the sake of the progress bar, execute "curl" for each file.
            # the parameters "--include" and "--form" will suppress the progress bar.
            for file in "${file_array[@]}"
            do
                # show delete link and filter out the delete token from the response header after upload.
                # it is important to save "curl's" "stdout" via a subshell to a variable or redirect it to another command,
                # which just redirects to "stdout" in order to have a sane output afterwards.
                # the progress bar is redirected to "stderr" and is only displayed,
                # if "stdout" is redirected to something; e.g. ">/dev/null", "tee /dev/null" or "| <some_command>".
                # the response header is redirected to "stdout", so redirecting "stdout" to "/dev/null" does not make any sense.
                # redirecting "curl's" "stderr" to "stdout" ("2>&1") will suppress the progress bar.
                curl_output=$(curl --request PUT --progress-bar --dump-header - --upload-file "${file}" "https://transfer.sh/")
                awk_output=$(awk \
                    'gsub("\r", "", $0) && tolower($1) ~ /x-url-delete/ \
                    {
                        delete_link=$2;
                        print "Delete command: curl --request DELETE " "\""delete_link"\"";

                        gsub(".*/", "", delete_link);
                        delete_token=delete_link;
                        print "Delete token: " delete_token;
                    }

                    END{
                        print "Download link: " $0;
                    }' <<< "${curl_output}")

                # return the results via "stdout", "awk" does not do this for some reason.
                echo -e "${awk_output}\n"

                # avoid rate limiting as much as possible; nginx: too many requests.
                if (( ${#file_array[@]} > 4 ))
                then
                    sleep 5
                fi
            done
            ;;

        "n"|"N")
            return 1
            ;;

        *)
            echo -e "\e[01;31mWrong input: '${upload_files}'.\e[0m" >&2
            return 1
    esac
}
```

</p></details>

#### Sample output
```bash
$ ls -lh
total 20M
-rw-r--r-- 1 <some_username> <some_username> 10M Apr  4 21:08 image.img
-rw-r--r-- 1 <some_username> <some_username> 10M Apr  4 21:08 image2.img
$ transfer image*
10240K  image2.img
10240K  image.img
20480K  total
Do you really want to upload the above files (2) to "transfer.sh"? (Y/n):
######################################################################################################################################################################################################################################## 100.0%
Delete command: curl --request DELETE "https://transfer.sh/wJw9pz/image2.img/mSctGx7pYCId"
Delete token: mSctGx7pYCId
Download link: https://transfer.sh/wJw9pz/image2.img

######################################################################################################################################################################################################################################## 100.0%
Delete command: curl --request DELETE "https://transfer.sh/ljJc5I/image.img/nw7qaoiKUwCU"
Delete token: nw7qaoiKUwCU
Download link: https://transfer.sh/ljJc5I/image.img

$ transfer "image.img" | awk --field-separator=": " '/Delete token:/ { print $2 } /Download link:/ { print $2 }'
10240K  image.img
10240K  total
Do you really want to upload the above files (1) to "transfer.sh"? (Y/n):
######################################################################################################################################################################################################################################## 100.0%
tauN5dE3fWJe
https://transfer.sh/MYkuqn/image.img
```

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

## Copyright and License

Code and documentation copyright 2011-2018 Remco Verhoef.
Code and documentation copyright 2018-2020 Andrea Spacca.
Code and documentation copyright 2020- Andrea Spacca and Stefan Benten.

Code released under [the MIT license](LICENSE).
