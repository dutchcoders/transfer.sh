# Table of Contents

* [Aliases](#aliases)
* [Uploading and downloading](#uploading-and-downloading)
* [Archiving and backups](#archiving-and-backups)
* [Encrypting and decrypting](#encrypting-and-decrypting)
* [Scanning for viruses](#scanning-for-viruses)

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