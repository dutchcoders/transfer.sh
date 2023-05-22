package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/dutchcoders/transfer.sh/server/storage"

	"github.com/dutchcoders/transfer.sh/server"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
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
	&cli.StringFlag{
		Name:    "listener",
		Usage:   "127.0.0.1:8080",
		Value:   "127.0.0.1:8080",
		EnvVars: []string{"LISTENER"},
	},
	// redirect to https?
	// hostnames
	&cli.StringFlag{
		Name:    "profile-listener",
		Usage:   "127.0.0.1:6060",
		Value:   "",
		EnvVars: []string{"PROFILE_LISTENER"},
	},
	&cli.BoolFlag{
		Name:    "force-https",
		Usage:   "",
		EnvVars: []string{"FORCE_HTTPS"},
	},
	&cli.StringFlag{
		Name:    "tls-listener",
		Usage:   "127.0.0.1:8443",
		Value:   "",
		EnvVars: []string{"TLS_LISTENER"},
	},
	&cli.BoolFlag{
		Name:    "tls-listener-only",
		Usage:   "",
		EnvVars: []string{"TLS_LISTENER_ONLY"},
	},
	&cli.StringFlag{
		Name:    "tls-cert-file",
		Value:   "",
		EnvVars: []string{"TLS_CERT_FILE"},
	},
	&cli.StringFlag{
		Name:    "tls-private-key",
		Value:   "",
		EnvVars: []string{"TLS_PRIVATE_KEY"},
	},
	&cli.StringFlag{
		Name:    "temp-path",
		Usage:   "path to temp files",
		Value:   os.TempDir(),
		EnvVars: []string{"TEMP_PATH"},
	},
	&cli.StringFlag{
		Name:    "web-path",
		Usage:   "path to static web files",
		Value:   "",
		EnvVars: []string{"WEB_PATH"},
	},
	&cli.StringFlag{
		Name:    "proxy-path",
		Usage:   "path prefix when service is run behind a proxy",
		Value:   "",
		EnvVars: []string{"PROXY_PATH"},
	},
	&cli.StringFlag{
		Name:    "proxy-port",
		Usage:   "port of the proxy when the service is run behind a proxy",
		Value:   "",
		EnvVars: []string{"PROXY_PORT"},
	},
	&cli.StringFlag{
		Name:    "email-contact",
		Usage:   "email address to link in Contact Us (front end)",
		Value:   "",
		EnvVars: []string{"EMAIL_CONTACT"},
	},
	&cli.StringFlag{
		Name:    "ga-key",
		Usage:   "key for google analytics (front end)",
		Value:   "",
		EnvVars: []string{"GA_KEY"},
	},
	&cli.StringFlag{
		Name:    "uservoice-key",
		Usage:   "key for user voice (front end)",
		Value:   "",
		EnvVars: []string{"USERVOICE_KEY"},
	},
	&cli.StringFlag{
		Name:    "provider",
		Usage:   "s3|gdrive|local",
		Value:   "",
		EnvVars: []string{"PROVIDER"},
	},
	&cli.StringFlag{
		Name:    "s3-endpoint",
		Usage:   "",
		Value:   "",
		EnvVars: []string{"S3_ENDPOINT"},
	},
	&cli.StringFlag{
		Name:    "s3-region",
		Usage:   "",
		Value:   "eu-west-1",
		EnvVars: []string{"S3_REGION"},
	},
	&cli.StringFlag{
		Name:    "aws-access-key",
		Usage:   "",
		Value:   "",
		EnvVars: []string{"AWS_ACCESS_KEY"},
	},
	&cli.StringFlag{
		Name:    "aws-secret-key",
		Usage:   "",
		Value:   "",
		EnvVars: []string{"AWS_SECRET_KEY"},
	},
	&cli.StringFlag{
		Name:    "bucket",
		Usage:   "",
		Value:   "",
		EnvVars: []string{"BUCKET"},
	},
	&cli.BoolFlag{
		Name:    "s3-no-multipart",
		Usage:   "Disables S3 Multipart Puts",
		EnvVars: []string{"S3_NO_MULTIPART"},
	},
	&cli.BoolFlag{
		Name:    "s3-path-style",
		Usage:   "Forces path style URLs, required for Minio.",
		EnvVars: []string{"S3_PATH_STYLE"},
	},
	&cli.StringFlag{
		Name:    "gdrive-client-json-filepath",
		Usage:   "",
		Value:   "",
		EnvVars: []string{"GDRIVE_CLIENT_JSON_FILEPATH"},
	},
	&cli.StringFlag{
		Name:    "gdrive-local-config-path",
		Usage:   "",
		Value:   "",
		EnvVars: []string{"GDRIVE_LOCAL_CONFIG_PATH"},
	},
	&cli.IntFlag{
		Name:    "gdrive-chunk-size",
		Usage:   "",
		Value:   googleapi.DefaultUploadChunkSize / 1024 / 1024,
		EnvVars: []string{"GDRIVE_CHUNK_SIZE"},
	},
	&cli.StringFlag{
		Name:    "storj-access",
		Usage:   "Access for the project",
		Value:   "",
		EnvVars: []string{"STORJ_ACCESS"},
	},
	&cli.StringFlag{
		Name:    "storj-bucket",
		Usage:   "Bucket to use within the project",
		Value:   "",
		EnvVars: []string{"STORJ_BUCKET"},
	},
	&cli.IntFlag{
		Name:    "rate-limit",
		Usage:   "requests per minute",
		Value:   0,
		EnvVars: []string{"RATE_LIMIT"},
	},
	&cli.IntFlag{
		Name:    "purge-days",
		Usage:   "number of days after uploads are purged automatically",
		Value:   0,
		EnvVars: []string{"PURGE_DAYS"},
	},
	&cli.IntFlag{
		Name:    "purge-interval",
		Usage:   "interval in hours to run the automatic purge for",
		Value:   0,
		EnvVars: []string{"PURGE_INTERVAL"},
	},
	&cli.Int64Flag{
		Name:    "max-upload-size",
		Usage:   "max limit for upload, in kilobytes",
		Value:   0,
		EnvVars: []string{"MAX_UPLOAD_SIZE"},
	},
	&cli.StringFlag{
		Name:    "lets-encrypt-hosts",
		Usage:   "host1, host2",
		Value:   "",
		EnvVars: []string{"HOSTS"},
	},
	&cli.StringFlag{
		Name:    "log",
		Usage:   "/var/log/transfersh.log",
		Value:   "",
		EnvVars: []string{"LOG"},
	},
	&cli.StringFlag{
		Name:    "basedir",
		Usage:   "path to storage",
		Value:   "",
		EnvVars: []string{"BASEDIR"},
	},
	&cli.StringFlag{
		Name:    "clamav-host",
		Usage:   "clamav-host",
		Value:   "",
		EnvVars: []string{"CLAMAV_HOST"},
	},
	&cli.BoolFlag{
		Name:    "perform-clamav-prescan",
		Usage:   "perform-clamav-prescan",
		EnvVars: []string{"PERFORM_CLAMAV_PRESCAN"},
	},
	&cli.StringFlag{
		Name:    "virustotal-key",
		Usage:   "virustotal-key",
		Value:   "",
		EnvVars: []string{"VIRUSTOTAL_KEY"},
	},
	&cli.BoolFlag{
		Name:    "profiler",
		Usage:   "enable profiling",
		EnvVars: []string{"PROFILER"},
	},
	&cli.StringFlag{
		Name:    "http-auth-user",
		Usage:   "user for http basic auth",
		Value:   "",
		EnvVars: []string{"HTTP_AUTH_USER"},
	},
	&cli.StringFlag{
		Name:    "http-auth-pass",
		Usage:   "pass for http basic auth",
		Value:   "",
		EnvVars: []string{"HTTP_AUTH_PASS"},
	},
	&cli.StringFlag{
		Name:    "http-auth-htpasswd",
		Usage:   "htpasswd file http basic auth",
		Value:   "",
		EnvVars: []string{"HTTP_AUTH_HTPASSWD"},
	},
	&cli.StringFlag{
		Name:    "http-auth-ip-whitelist",
		Usage:   "comma separated list of ips allowed to upload without being challenged an http auth",
		Value:   "",
		EnvVars: []string{"HTTP_AUTH_IP_WHITELIST"},
	},
	&cli.StringFlag{
		Name:    "ip-whitelist",
		Usage:   "comma separated list of ips allowed to connect to the service",
		Value:   "",
		EnvVars: []string{"IP_WHITELIST"},
	},
	&cli.StringFlag{
		Name:    "ip-blacklist",
		Usage:   "comma separated list of ips not allowed to connect to the service",
		Value:   "",
		EnvVars: []string{"IP_BLACKLIST"},
	},
	&cli.StringFlag{
		Name:    "cors-domains",
		Usage:   "comma separated list of domains allowed for CORS requests",
		Value:   "",
		EnvVars: []string{"CORS_DOMAINS"},
	},
	&cli.IntFlag{
		Name:    "random-token-length",
		Usage:   "",
		Value:   10,
		EnvVars: []string{"RANDOM_TOKEN_LENGTH"},
	},
}

// Cmd wraps cli.app
type Cmd struct {
	*cli.App
}

func versionCommand(_ *cli.Context) error {
	fmt.Println(color.YellowString("transfer.sh %s: Easy file sharing from the command line", Version))
	return nil
}

// New is the factory for transfer.sh
func New() *Cmd {
	logger := log.New(os.Stdout, "[transfer.sh]", log.LstdFlags)

	app := cli.NewApp()
	app.Name = "transfer.sh"
	app.Authors = []*cli.Author{}
	app.Usage = "transfer.sh"
	app.Description = `Easy file sharing from the command line`
	app.Version = Version
	app.Flags = globalFlags
	app.CustomAppHelpTemplate = helpTemplate
	app.Commands = []*cli.Command{
		{
			Name:   "version",
			Action: versionCommand,
		},
	}

	app.Before = func(c *cli.Context) error {
		return nil
	}

	app.Action = func(c *cli.Context) error {
		var options []server.OptionFn
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

		if v := c.String("email-contact"); v != "" {
			options = append(options, server.EmailContact(v))
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

		if v := c.Bool("perform-clamav-prescan"); v {
			if c.String("clamav-host") == "" {
				return errors.New("clamav-host not set")
			}

			options = append(options, server.PerformClamavPrescan(v))
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

		if httpAuthHtpasswd := c.String("http-auth-htpasswd"); httpAuthHtpasswd != "" {
			options = append(options, server.HTTPAuthHtpasswd(httpAuthHtpasswd))
		}

		if httpAuthIPWhitelist := c.String("http-auth-ip-whitelist"); httpAuthIPWhitelist != "" {
			ipFilterOptions := server.IPFilterOptions{}
			ipFilterOptions.AllowedIPs = strings.Split(httpAuthIPWhitelist, ",")
			ipFilterOptions.BlockByDefault = false
			options = append(options, server.HTTPAUTHFilterOptions(ipFilterOptions))
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
				return errors.New("access-key not set.")
			} else if secretKey := c.String("aws-secret-key"); secretKey == "" {
				return errors.New("secret-key not set.")
			} else if bucket := c.String("bucket"); bucket == "" {
				return errors.New("bucket not set.")
			} else if store, err := storage.NewS3Storage(c.Context, accessKey, secretKey, bucket, purgeDays, c.String("s3-region"), c.String("s3-endpoint"), c.Bool("s3-no-multipart"), c.Bool("s3-path-style"), logger); err != nil {
				return err
			} else {
				options = append(options, server.UseStorage(store))
			}
		case "gdrive":
			chunkSize := c.Int("gdrive-chunk-size") * 1024 * 1024

			if clientJSONFilepath := c.String("gdrive-client-json-filepath"); clientJSONFilepath == "" {
				return errors.New("gdrive-client-json-filepath not set.")
			} else if localConfigPath := c.String("gdrive-local-config-path"); localConfigPath == "" {
				return errors.New("gdrive-local-config-path not set.")
			} else if basedir := c.String("basedir"); basedir == "" {
				return errors.New("basedir not set.")
			} else if store, err := storage.NewGDriveStorage(c.Context, clientJSONFilepath, localConfigPath, basedir, chunkSize, logger); err != nil {
				return err
			} else {
				options = append(options, server.UseStorage(store))
			}
		case "storj":
			if access := c.String("storj-access"); access == "" {
				return errors.New("storj-access not set.")
			} else if bucket := c.String("storj-bucket"); bucket == "" {
				return errors.New("storj-bucket not set.")
			} else if store, err := storage.NewStorjStorage(c.Context, access, bucket, purgeDays, logger); err != nil {
				return err
			} else {
				options = append(options, server.UseStorage(store))
			}
		case "local":
			if v := c.String("basedir"); v == "" {
				return errors.New("basedir not set.")
			} else if store, err := storage.NewLocalStorage(v, logger); err != nil {
				return err
			} else {
				options = append(options, server.UseStorage(store))
			}
		default:
			return errors.New("Provider not set or invalid.")
		}

		srvr, err := server.New(
			options...,
		)

		if err != nil {
			logger.Println(color.RedString("Error starting server: %s", err.Error()))
			return err
		}

		srvr.Run()
		return nil
	}

	return &Cmd{
		App: app,
	}
}
