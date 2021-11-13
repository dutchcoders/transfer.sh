# Table of Contents

* [Aliases](#aliases)
* [Uploading and downloading](#uploading-and-downloading)
* [Archiving and backups](#archiving-and-backups)
* [Encrypting and decrypting](#encrypting-and-decrypting)
* [Scanning for viruses](#scanning-for-viruses)
* [Uploading and copy download command](#uploading-and-copy-download-command)

## Aliases
<a name="aliases"/>

## Add alias to .bashrc or .zshrc

### Using curl
```bash
transfer() {
    curl --progress-bar --upload-file "$1" https://transfer.sh/$(basename "$1") | tee /dev/null;
    echo
}

alias transfer=transfer
```

### Using wget
```bash
transfer() {
    wget -t 1 -qO - --method=PUT --body-file="$1" --header="Content-Type: $(file -b --mime-type "$1")" https://transfer.sh/$(basename "$1");
    echo
}

alias transfer=transfer
```

## Add alias for fish-shell

### Using curl
```fish
function transfer --description 'Upload a file to transfer.sh'
    if [ $argv[1] ]
        # write to output to tmpfile because of progress bar
        set -l tmpfile ( mktemp -t transferXXXXXX )
        curl --progress-bar --upload-file "$argv[1]" https://transfer.sh/(basename $argv[1]) >> $tmpfile
        cat $tmpfile
        command rm -f $tmpfile
    else
        echo 'usage: transfer FILE_TO_TRANSFER'
    end
end

funcsave transfer
```

### Using wget
```fish
function transfer --description 'Upload a file to transfer.sh'
    if [ $argv[1] ]
        wget -t 1 -qO - --method=PUT --body-file="$argv[1]" --header="Content-Type: (file -b --mime-type $argv[1])" https://transfer.sh/(basename $argv[1])
    else
        echo 'usage: transfer FILE_TO_TRANSFER'
    end
end

funcsave transfer
```

Now run it like this:
```bash
$ transfer test.txt
```

## Add alias on Windows

Put a file called `transfer.cmd` somewhere in your PATH with this inside it:
```cmd
@echo off
setlocal
:: use env vars to pass names to PS, to avoid escaping issues
set FN=%~nx1
set FULL=%1
powershell -noprofile -command "$(Invoke-Webrequest -Method put -Infile $Env:FULL https://transfer.sh/$Env:FN).Content"
```

## Uploading and Downloading
<a name="uploading-and-downloading"/>

### Uploading with wget
```bash
$ wget --method PUT --body-file=/tmp/file.tar https://transfer.sh/file.tar -O - -nv 
```

### Uploading with PowerShell
```posh
PS H:\> invoke-webrequest -method put -infile .\file.txt https://transfer.sh/file.txt 
```

### Upload using HTTPie
```bash
$ http https://transfer.sh/ -vv < /tmp/test.log 
```

### Uploading a filtered text file
```bash
$ grep 'pound' /var/log/syslog | curl --upload-file - https://transfer.sh/pound.log 
```

### Downloading with curl
```bash
$ curl https://transfer.sh/1lDau/test.txt -o test.txt
```

### Downloading with wget
```bash
$ wget https://transfer.sh/1lDau/test.txt
```

## Archiving and backups
<a name="archiving-and-backups"/>

### Backup, encrypt and transfer a MySQL dump
```bash
$ mysqldump --all-databases | gzip | gpg -ac -o- | curl -X PUT --upload-file "-" https://transfer.sh/test.txt
```

### Archive and upload directory
```bash
$ tar -czf - /var/log/journal | curl --upload-file - https://transfer.sh/journal.tar.gz
```

### Uploading multiple files at once
```bash
$ curl -i -F filedata=@/tmp/hello.txt -F filedata=@/tmp/hello2.txt https://transfer.sh/
```

### Combining downloads as zip or tar.gz archive
```bash
$ curl https://transfer.sh/(15HKz/hello.txt,15HKz/hello.txt).tar.gz
$ curl https://transfer.sh/(15HKz/hello.txt,15HKz/hello.txt).zip 
```

### Transfer and send email with link (using an alias)
```bash
$ transfer /tmp/hello.txt | mail -s "Hello World" user@yourmaildomain.com 
```
## Encrypting and decrypting
<a name="encrypting-and-decrypting"/>

### Encrypting files with password using gpg
```bash
$ cat /tmp/hello.txt | gpg -ac -o- | curl -X PUT --upload-file "-" https://transfer.sh/test.txt
```

### Downloading and decrypting
```bash
$ curl https://transfer.sh/1lDau/test.txt | gpg -o- > /tmp/hello.txt 
```

### Import keys from [keybase](https://keybase.io/)
```bash
$ keybase track [them] # Encrypt for recipient(s)
$ cat somebackupfile.tar.gz | keybase encrypt [them] | curl --upload-file '-' https://transfer.sh/test.txt # Decrypt
$ curl https://transfer.sh/sqUFi/test.md | keybase decrypt
```

## Scanning for viruses
<a name="scanning-for-viruses"/>

### Scan for malware or viruses using Clamav
```bash
$ wget http://www.eicar.org/download/eicar.com
$ curl -X PUT --upload-file ./eicar.com https://transfer.sh/eicar.com/scan
```

### Upload malware to VirusTotal, get a permalink in return
```bash
$ curl -X PUT --upload-file nhgbhhj https://transfer.sh/test.txt/virustotal 
```
## Uploading and copy download command

Download commands can be automatically copied to the clipboard after files are uploaded using transfer.sh.

It was designed for Linux or macOS.

### 1. Install xclip or xsel for Linux, macOS skips this step

- install xclip see https://command-not-found.com/xclip

- install xsel  see https://command-not-found.com/xsel

Install later, add pbcopy and pbpaste to .bashrc or .zshrc or its equivalent.

- If use xclip, paste the following lines:

```sh
alias pbcopy='xclip -selection clipboard'
alias pbpaste='xclip -selection clipboard -o'
```

- If use xsel, paste the following lines:

```sh
alias pbcopy='xsel --clipboard --input'
alias pbpaste='xsel --clipboard --output'
```

### 2. Add Uploading and copy download command shell function

1. Open .bashrc or .zshrc  or its equivalent.

2. Add the following shell script:

   ```sh
   transfer() {
     curl --progress-bar --upload-file "$1" https://transfer.sh/$(basename "$1") | pbcopy;
     echo "1) Download link:"
     echo "$(pbpaste)"
   
     echo "\n2) Linux or macOS download command:"
     linux_macos_download_command="wget $(pbpaste)"
     echo $linux_macos_download_command
   
     echo "\n3) Windows download command:"
     windows_download_command="Invoke-WebRequest -Uri "$(pbpaste)" -OutFile $(basename $1)"
     echo $windows_download_command
   
     case $2 in
       l|m)  echo $linux_macos_download_command | pbcopy
       ;;
       w)  echo $windows_download_command | pbcopy
       ;;
     esac
   }
   ```


### 3. Test

The transfer command has two parameters:

1. The first parameter is the path to upload the file.

2. The second parameter indicates which system's download command is copied. optional:

   - This parameter is empty to copy the download link.

   - `l` or `m` copy the Linux or macOS command that downloaded the file.

   -  `w` copy the Windows command that downloaded the file.

For example, The command to download the file on Windows will be copied:

```sh
$ transfer ~/temp/a.log w
######################################################################## 100.0%
1) Download link:
https://transfer.sh/y0qr2c/a.log

2) Linux or macOS download command:
wget https://transfer.sh/y0qr2c/a.log

3) Windows download command:
Invoke-WebRequest -Uri https://transfer.sh/y0qr2c/a.log -OutFile a.log 
```
