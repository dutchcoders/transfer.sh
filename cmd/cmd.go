package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/dutchcoders/transfer.sh/server"
	"github.com/fatih/color"
	"github.com/urfave/cli"
	"google.golang.org/api/googleapi"
)

// Version is inject at build time
var Version = "0.0.0"
var helpTemplate = `NAME:
{{.Name}} - {{.Usage}}

DESCRIPTION:
{{.Description}}

USAGE:
{{.Name}} {{if .Flags}}[flags] {{end}}command{{if .Flags}}{{end}} [arguments...]

COMMANDS:
{{range .Commands}}{{join .Names ", "}}{{ "\t" }}{{.Usage}}
{{end}}{{if .Flags}}
FLAGS:
{{range .Flags}}{{.}}
{{end}}{{end}}
VERSION:
` + Version +
	`{{ "\n"}}`

var globalFlags = []cli.Flag{
	cli.StringFlag{
		Name:   "listener",
		Usage:  "127.0.0.1:8080",
		Value:  "127.0.0.1:8080",
		EnvVar: "LISTENER",
	},
	// redirect to https?
	// hostnames
	cli.StringFlag{
		Name:   "profile-listener",
		Usage:  "127.0.0.1:6060",
		Value:  "",
		EnvVar: "PROFILE_LISTENER",
	},
	cli.BoolFlag{
		Name:   "force-https",
		Usage:  "",
		EnvVar: "FORCE_HTTPS",
	},
	cli.StringFlag{
		Name:   "tls-listener",
		Usage:  "127.0.0.1:8443",
		Value:  "",
		EnvVar: "TLS_LISTENER",
	},
	cli.BoolFlag{
		Name:   "tls-listener-only",
		Usage:  "",
		EnvVar: "TLS_LISTENER_ONLY",
	},
	cli.StringFlag{
		Name:   "tls-cert-file",
		Value:  "",
		EnvVar: "TLS_CERT_FILE",
	},
	cli.StringFlag{
		Name:   "tls-private-key",
		Value:  "",
		EnvVar: "TLS_PRIVATE_KEY",
	},
	cli.StringFlag{
		Name:   "temp-path",
		Usage:  "path to temp files",
		Value:  os.TempDir(),
		EnvVar: "TEMP_PATH",
	},
	cli.StringFlag{
		Name:   "web-path",
		Usage:  "path to static web files",
		Value:  "",
		EnvVar: "WEB_PATH",
	},
	cli.StringFlag{
		Name:   "proxy-path",
		Usage:  "path prefix when service is run behind a proxy",
		Value:  "",
		EnvVar: "PROXY_PATH",
	},
	cli.StringFlag{
		Name:   "proxy-port",
		Usage:  "port of the proxy when the service is run behind a proxy",
		Value:  "",
		EnvVar: "PROXY_PORT",
	},
	cli.StringFlag{
		Name:   "ga-key",
		Usage:  "key for google analytics (front end)",
		Value:  "",
		EnvVar: "GA_KEY",
	},
	cli.StringFlag{
		Name:   "uservoice-key",
		Usage:  "key for user voice (front end)",
		Value:  "",
		EnvVar: "USERVOICE_KEY",
	},
	cli.StringFlag{
		Name:   "provider",
		Usage:  "s3|gdrive|local",
		Value:  "",
		EnvVar: "PROVIDER",
	},
	cli.StringFlag{
		Name:   "s3-endpoint",
		Usage:  "",
		Value:  "",
		EnvVar: "S3_ENDPOINT",
	},
	cli.StringFlag{
		Name:   "s3-region",
		Usage:  "",
		Value:  "eu-west-1",
		EnvVar: "S3_REGION",
	},
	cli.StringFlag{
		Name:   "aws-access-key",
		Usage:  "",
		Value:  "",
		EnvVar: "AWS_ACCESS_KEY",
	},
	cli.StringFlag{
		Name:   "aws-secret-key",
		Usage:  "",
		Value:  "",
		EnvVar: "AWS_SECRET_KEY",
	},
	cli.StringFlag{
		Name:   "bucket",
		Usage:  "",
		Value:  "",
		EnvVar: "BUCKET",
	},
	cli.BoolFlag{
		Name:   "s3-no-multipart",
		Usage:  "Disables S3 Multipart Puts",
		EnvVar: "S3_NO_MULTIPART",
	},
	cli.BoolFlag{
		Name:   "s3-path-style",
		Usage:  "Forces path style URLs, required for Minio.",
		EnvVar: "S3_PATH_STYLE",
	},
	cli.StringFlag{
		Name:   "gdrive-client-json-filepath",
		Usage:  "",
		Value:  "",
		EnvVar: "GDRIVE_CLIENT_JSON_FILEPATH",
	},
	cli.StringFlag{
		Name:   "gdrive-local-config-path",
		Usage:  "",
		Value:  "",
		EnvVar: "GDRIVE_LOCAL_CONFIG_PATH",
	},
	cli.IntFlag{
		Name:   "gdrive-chunk-size",
		Usage:  "",
		Value:  googleapi.DefaultUploadChunkSize / 1024 / 1024,
		EnvVar: "GDRIVE_CHUNK_SIZE",
	},
	cli.StringFlag{
		Name:   "storj-access",
		Usage:  "Access for the project",
		Value:  "",
		EnvVar: "STORJ_ACCESS",
	},
	cli.StringFlag{
		Name:   "storj-bucket",
		Usage:  "Bucket to use within the project",
		Value:  "",
		EnvVar: "STORJ_BUCKET",
	},
	cli.IntFlag{
		Name:   "rate-limit",
		Usage:  "requests per minute",
		Value:  0,
		EnvVar: "RATE_LIMIT",
	},
	cli.IntFlag{
		Name:   "purge-days",
		Usage:  "number of days after uploads are purged automatically",
		Value:  0,
		EnvVar: "PURGE_DAYS",
	},
	cli.IntFlag{
		Name:   "purge-interval",
		Usage:  "interval in hours to run the automatic purge for",
		Value:  0,
		EnvVar: "PURGE_INTERVAL",
	},
	cli.Int64Flag{
		Name:   "max-upload-size",
		Usage:  "max limit for upload, in kilobytes",
		Value:  0,
		EnvVar: "MAX_UPLOAD_SIZE",
	},
	cli.StringFlag{
		Name:   "lets-encrypt-hosts",
		Usage:  "host1, host2",
		Value:  "",
		EnvVar: "HOSTS",
	},
	cli.StringFlag{
		Name:   "log",
		Usage:  "/var/log/transfersh.log",
		Value:  "",
		EnvVar: "LOG",
	},
	cli.StringFlag{
		Name:   "basedir",
		Usage:  "path to storage",
		Value:  "",
		EnvVar: "BASEDIR",
	},
	cli.StringFlag{
		Name:   "clamav-host",
		Usage:  "clamav-host",
		Value:  "",
		EnvVar: "CLAMAV_HOST",
	},
	cli.StringFlag{
		Name:   "virustotal-key",
		Usage:  "virustotal-key",
		Value:  "",
		EnvVar: "VIRUSTOTAL_KEY",
	},
	cli.BoolFlag{
		Name:   "profiler",
		Usage:  "enable profiling",
		EnvVar: "PROFILER",
	},
	cli.StringFlag{
		Name:   "http-auth-user",
		Usage:  "user for http basic auth",
		Value:  "",
		EnvVar: "HTTP_AUTH_USER",
	},
	cli.StringFlag{
		Name:   "http-auth-pass",
		Usage:  "pass for http basic auth",
		Value:  "",
		EnvVar: "HTTP_AUTH_PASS",
	},
	cli.StringFlag{
		Name:   "ip-whitelist",
		Usage:  "comma separated list of ips allowed to connect to the service",
		Value:  "",
		EnvVar: "IP_WHITELIST",
	},
	cli.StringFlag{
		Name:   "ip-blacklist",
		Usage:  "comma separated list of ips not allowed to connect to the service",
		Value:  "",
		EnvVar: "IP_BLACKLIST",
	},
	cli.StringFlag{
		Name:   "cors-domains",
		Usage:  "comma separated list of domains allowed for CORS requests",
		Value:  "",
		EnvVar: "CORS_DOMAINS",
	},
	cli.IntFlag{
		Name:   "random-token-length",
		Usage:  "",
		Value:  6,
		EnvVar: "RANDOM_TOKEN_LENGTH",
	},
}

// Cmd wraps cli.app
type Cmd struct {
	*cli.App
}

func versionAction(c *cli.Context) {
	fmt.Println(color.YellowString(fmt.Sprintf("transfer.sh %s: Easy file sharing from the command line", Version)))
}

// New is the factory for transfer.sh
func New() *Cmd {
	logger := log.New(os.Stdout, "[transfer.sh]", log.LstdFlags)

	app := cli.NewApp()
	app.Name = "transfer.sh"
	app.Author = ""
	app.Usage = "transfer.sh"
	app.Description = `Easy file sharing from the command line`
	app.Version = Version
	app.Flags = globalFlags
	app.CustomAppHelpTemplate = helpTemplate
	app.Commands = []cli.Command{
		{
			Name:   "version",
			Action: versionAction,
		},
	}

	app.Before = func(c *cli.Context) error {
		return nil
	}

	app.Action = func(c *cli.Context) {
		options := []server.OptionFn{}
		if v := c.String("listener"); v != "" {
			options = append(options, server.Listener(v))
		}

		if v := c.String("cors-domains"); v != "" {
			options = append(options, server.CorsDomains(v))
		}

		if v := c.String("tls-listener"); v == "" {
		} else if c.Bool("tls-listener-only") {
			options = append(options, server.TLSListener(v, true))
		} else {
			options = append(options, server.TLSListener(v, false))
		}

		if v := c.String("profile-listener"); v != "" {
			options = append(options, server.ProfileListener(v))
		}

		if v := c.String("web-path"); v != "" {
			options = append(options, server.WebPath(v))
		}

		if v := c.String("proxy-path"); v != "" {
			options = append(options, server.ProxyPath(v))
		}

		if v := c.String("proxy-port"); v != "" {
			options = append(options, server.ProxyPort(v))
		}

		if v := c.String("ga-key"); v != "" {
			options = append(options, server.GoogleAnalytics(v))
		}

		if v := c.String("uservoice-key"); v != "" {
			options = append(options, server.UserVoice(v))
		}

		if v := c.String("temp-path"); v != "" {
			options = append(options, server.TempPath(v))
		}

		if v := c.String("log"); v != "" {
			options = append(options, server.LogFile(logger, v))
		} else {
			options = append(options, server.Logger(logger))
		}

		if v := c.String("lets-encrypt-hosts"); v != "" {
			options = append(options, server.UseLetsEncrypt(strings.Split(v, ",")))
		}

		if v := c.String("virustotal-key"); v != "" {
			options = append(options, server.VirustotalKey(v))
		}

		if v := c.String("clamav-host"); v != "" {
			options = append(options, server.ClamavHost(v))
		}

		if v := c.Int64("max-upload-size"); v > 0 {
			options = append(options, server.MaxUploadSize(v))
		}

		if v := c.Int("rate-limit"); v > 0 {
			options = append(options, server.RateLimit(v))
		}

		v := c.Int("random-token-length")
		options = append(options, server.RandomTokenLength(v))

		purgeDays := c.Int("purge-days")
		purgeInterval := c.Int("purge-interval")
		if purgeDays > 0 && purgeInterval > 0 {
			options = append(options, server.Purge(purgeDays, purgeInterval))
		}

		if cert := c.String("tls-cert-file"); cert == "" {
		} else if pk := c.String("tls-private-key"); pk == "" {
		} else {
			options = append(options, server.TLSConfig(cert, pk))
		}

		if c.Bool("profiler") {
			options = append(options, server.EnableProfiler())
		}

		if c.Bool("force-https") {
			options = append(options, server.ForceHTTPS())
		}

		if httpAuthUser := c.String("http-auth-user"); httpAuthUser == "" {
		} else if httpAuthPass := c.String("http-auth-pass"); httpAuthPass == "" {
		} else {
			options = append(options, server.HTTPAuthCredentials(httpAuthUser, httpAuthPass))
		}

		applyIPFilter := false
		ipFilterOptions := server.IPFilterOptions{}
		if ipWhitelist := c.String("ip-whitelist"); ipWhitelist != "" {
			applyIPFilter = true
			ipFilterOptions.AllowedIPs = strings.Split(ipWhitelist, ",")
			ipFilterOptions.BlockByDefault = true
		}

		if ipBlacklist := c.String("ip-blacklist"); ipBlacklist != "" {
			applyIPFilter = true
			ipFilterOptions.BlockedIPs = strings.Split(ipBlacklist, ",")
		}

		if applyIPFilter {
			options = append(options, server.FilterOptions(ipFilterOptions))
		}

		switch provider := c.String("provider"); provider {
		case "s3":
			if accessKey := c.String("aws-access-key"); accessKey == "" {
				panic("access-key not set.")
			} else if secretKey := c.String("aws-secret-key"); secretKey == "" {
				panic("secret-key not set.")
			} else if bucket := c.String("bucket"); bucket == "" {
				panic("bucket not set.")
			} else if storage, err := server.NewS3Storage(accessKey, secretKey, bucket, purgeDays, c.String("s3-region"), c.String("s3-endpoint"), c.Bool("s3-no-multipart"), c.Bool("s3-path-style"), logger); err != nil {
				panic(err)
			} else {
				options = append(options, server.UseStorage(storage))
			}
		case "gdrive":
			chunkSize := c.Int("gdrive-chunk-size")

			if clientJSONFilepath := c.String("gdrive-client-json-filepath"); clientJSONFilepath == "" {
				panic("client-json-filepath not set.")
			} else if localConfigPath := c.String("gdrive-local-config-path"); localConfigPath == "" {
				panic("local-config-path not set.")
			} else if basedir := c.String("basedir"); basedir == "" {
				panic("basedir not set.")
			} else if storage, err := server.NewGDriveStorage(clientJSONFilepath, localConfigPath, basedir, chunkSize, logger); err != nil {
				panic(err)
			} else {
				options = append(options, server.UseStorage(storage))
			}
		case "storj":
			if access := c.String("storj-access"); access == "" {
				panic("storj-access not set.")
			} else if bucket := c.String("storj-bucket"); bucket == "" {
				panic("storj-bucket not set.")
			} else if storage, err := server.NewStorjStorage(access, bucket, purgeDays, logger); err != nil {
				panic(err)
			} else {
				options = append(options, server.UseStorage(storage))
			}
		case "local":
			if v := c.String("basedir"); v == "" {
				panic("basedir not set.")
			} else if storage, err := server.NewLocalStorage(v, logger); err != nil {
				panic(err)
			} else {
				options = append(options, server.UseStorage(storage))
			}
		default:
			panic("Provider not set or invalid.")
		}

		srvr, err := server.New(
			options...,
		)

		if err != nil {
			logger.Println(color.RedString("Error starting server: %s", err.Error()))
			return
		}

		srvr.Run()
	}

	return &Cmd{
		App: app,
	}
}
