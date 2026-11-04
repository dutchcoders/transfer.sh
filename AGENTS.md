# AGENTS.md

Canonical project instructions for coding agents. Maintain shared guidance here.

## Overview

transfer.sh is a Go file-sharing server (module `go.mod`, Go 1.25). This repo is a fork of `dutchcoders/transfer.sh` whose whole purpose is the WebDAV storage backend (`server/storage/webdav.go`): keep it working and documented when merging from `upstream`. Entry point is `main.go`, which calls `cmd.New()`. Helm chart sources live in `k8s/transfer.sh/`.

## Commands

```bash
# Build
go build ./...

# Lint (Makefile target, config in .golangci.yml, v2 format)
make lint

# CI runs exactly this (.github/workflows/test.yml)
go vet ./...
go test ./...

# Single package / single test
go test ./server -run TestName
go test ./server/storage -run TestName
go test ./cmd -run TestName

# Run locally with the local provider
go run main.go --provider=local --listener :8080 --temp-path=/tmp/ --basedir=/tmp/

# WebDAV provider (from README)
go run main.go --provider webdav --basedir /remote/path --webdav-url https://dav.example.com --webdav-username user --webdav-password pass
```

`flake.nix` only exposes the package and app (`packages.transfer-sh`, `apps.transfer-sh`), not a dev shell.

## Architecture

### CLI to server options (`cmd/cmd.go`)

- `globalFlags` is a list of `urfave/cli` flags. Each flag declares its env var through `EnvVars` (for example `listener` is `LISTENER`, `webdav-url` is `WEBDAV_URL`), so every flag can be set either way. The README has the full flag/env table.
- In `app.Action`, each set flag is turned into a `server.OptionFn` (functional options defined in `server/server.go`, such as `server.Listener`, `server.Purge`, `server.RateLimit`, `server.FilterOptions`). Validation errors (for example `clamav-host not set`) are returned from there.
- The end of the action picks the storage via `switch c.String("provider")` and passes it with `server.UseStorage(store)`. The options are then applied to `server.New(...)` and the server runs.

### Server package (`server/`)

- `server.go`: `Server` struct, option functions, and `Run()`, which builds a `gorilla/mux` router (view, zip/tar/tar.gz, get/download/inline, HEAD, PUT/upload with `basicAuthHandler`, DELETE with deletion token, virustotal and clamav scan routes, `/health.html`). A per-IP rate limit wraps the GET handler when `rate-limit` is set. The whole router is wrapped by `ipFilterHandler` plus CORS, force-HTTPS and proxy-path handling.
- `handlers.go`: all HTTP handlers. Uploading writes the file plus a `<filename>.metadata` JSON object through the same `Storage`. Metadata holds the deletion token, max downloads and max date (`Max-Days` / `Max-Downloads` headers). `checkMetadata` enforces it on download, `checkDeletionToken` on delete, and the `X-Url-Delete` response header returns the delete URL. `purgeHandler` is a ticker that calls `storage.Purge` with `purge-days` every `purge-interval` hours. `basicAuthHandler` checks `http-auth-user/pass` or an htpasswd file, with an optional IP allowlist for auth bypass.
- `ip_filter.go`: allow/block IP lists (`ip-whitelist`, `ip-blacklist`, `http-auth-ip-whitelist`).
- `clamav.go` and `virustotal.go`: optional virus scanning (`clamav-host`, `perform-clamav-prescan`, `virustotal-key`).
- `token.go`: random token generation, length set by `random-token-length`.

### Storage providers (`server/storage/`)

- The `Storage` interface lives in `server/storage/common.go`: `Get`, `Head`, `Put`, `Delete`, `IsNotExist`, `Purge`, `IsRangeSupported`, `Type`. `Range` and `CloseCheck` helpers are in the same file.
- Implementations: `local.go`, `s3.go`, `gdrive.go`, `storj.go`, `webdav.go`. Each has a `New<Name>Storage(...)` constructor.
- To add a provider: implement the interface in a new file, add the constructor call plus its required-flag checks as a new `case` in the provider `switch` in `cmd/cmd.go`, add the flags (with `EnvVars`) to `globalFlags`, and update the provider lists in `README.md` and the `provider` flag usage.
- `local` and `gdrive` need `basedir`; `webdav` needs `basedir`, `webdav-url`, `webdav-username`, `webdav-password`; `s3` needs `bucket`; `storj` needs an access grant and bucket.

### Web frontend

The UI comes from the external module `github.com/dutchcoders/transfer.sh-web`, embedded as bindata assets (`web.Asset`, `web.AssetDir`) and served through a `bindata` file system. If `--web-path` / `WEB_PATH` is set, `server.go` instead serves static files and parses `*.html` / `*.txt` templates from that directory (meant for development or a custom front end).

## README notes

- Providers listed in the README: s3, gdrive, storj, webdav, local. Per-provider usage sections (S3, Google Drive, Storj, WebDAV) explain the required options.
- Docker usage in the README: `docker run --publish 8080:8080 dutchcoders/transfer.sh:latest --provider local --basedir /tmp/` (a `latest-noroot` image also exists).
