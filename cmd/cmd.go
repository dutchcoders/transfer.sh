package cmd

import (
	"fmt"

	"os"

	"strings"

	"github.com/dutchcoders/transfer.sh/server"
	"github.com/fatih/color"
	"github.com/minio/cli"
	"log"
)

var Version = "0.1"
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
		Name:  "listener",
		Usage: "127.0.0.1:8080",
		Value: "127.0.0.1:8080",
	},
	// redirect to https?
	// hostnames
	cli.StringFlag{
		Name:  "profile-listener",
		Usage: "127.0.0.1:6060",
		Value: "",
	},
	cli.BoolFlag{
		Name:  "force-https",
		Usage: "",
	},
	cli.StringFlag{
		Name:  "tls-listener",
		Usage: "127.0.0.1:8443",
		Value: "",
	},
	cli.BoolFlag{
		Name:  "tls-listener-only",
		Usage: "",
	},
	cli.StringFlag{
		Name:  "tls-cert-file",
		Value: "",
	},
	cli.StringFlag{
		Name:  "tls-private-key",
		Value: "",
	},
	cli.StringFlag{
		Name:  "temp-path",
		Usage: "path to temp files",
		Value: os.TempDir(),
	},
	cli.StringFlag{
		Name:  "web-path",
		Usage: "path to static web files",
		Value: "",
	},
	cli.StringFlag{
		Name:  "ga-key",
		Usage: "key for google analytics (front end)",
		Value: "",
	},
	cli.StringFlag{
		Name:  "uservoice-key",
		Usage: "key for user voice (front end)",
		Value: "",
	},
	cli.StringFlag{
		Name:  "provider",
		Usage: "s3|gdrive|local",
		Value: "",
	},
	cli.StringFlag{
		Name:   "s3-endpoint",
		Usage:  "",
		Value:  "http://s3-eu-west-1.amazonaws.com",
		EnvVar: "S3_ENDPOINT",
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
	cli.StringFlag{
		Name:  "gdrive-client-json-filepath",
		Usage: "",
		Value: "",
	},
	cli.StringFlag{
		Name:  "gdrive-local-config-path",
		Usage: "",
		Value: "",
	},
	cli.IntFlag{
		Name:   "rate-limit",
		Usage:  "requests per minute",
		Value:  0,
		EnvVar: "",
	},
	cli.StringFlag{
		Name:   "lets-encrypt-hosts",
		Usage:  "host1, host2",
		Value:  "",
		EnvVar: "HOSTS",
	},
	cli.StringFlag{
		Name:  "log",
		Usage: "/var/log/transfersh.log",
		Value: "",
	},
	cli.StringFlag{
		Name:  "basedir",
		Usage: "path to storage",
		Value: "",
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
		Name:  "profiler",
		Usage: "enable profiling",
	},
	cli.StringFlag{
		Name:  "http-auth-user",
		Usage: "user for http basic auth",
		Value: "",
	},
	cli.StringFlag{
		Name:  "http-auth-pass",
		Usage: "pass for http basic auth",
		Value: "",
	},
}

type Cmd struct {
	*cli.App
}

func VersionAction(c *cli.Context) {
	fmt.Println(color.YellowString(fmt.Sprintf("transfer.sh: Easy file sharing from the command line")))
}

func New() *Cmd {
	logger := log.New(os.Stdout, "[transfer.sh]", log.LstdFlags)

	app := cli.NewApp()
	app.Name = "transfer.sh"
	app.Author = ""
	app.Usage = "transfer.sh"
	app.Description = `Easy file sharing from the command line`
	app.Flags = globalFlags
	app.CustomAppHelpTemplate = helpTemplate
	app.Commands = []cli.Command{
		{
			Name:   "version",
			Action: VersionAction,
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

		if v := c.Int("rate-limit"); v > 0 {
			options = append(options, server.RateLimit(v))
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
			options = append(options, server.ForceHTTPs())
		}

		if httpAuthUser := c.String("http-auth-user"); httpAuthUser == "" {
		} else if httpAuthPass := c.String("http-auth-pass"); httpAuthPass == "" {
		} else {
			options = append(options, server.HttpAuthCredentials(httpAuthUser, httpAuthPass))
		}

		switch provider := c.String("provider"); provider {
		case "s3":
			if accessKey := c.String("aws-access-key"); accessKey == "" {
				panic("access-key not set.")
			} else if secretKey := c.String("aws-secret-key"); secretKey == "" {
				panic("secret-key not set.")
			} else if bucket := c.String("bucket"); bucket == "" {
				panic("bucket not set.")
			} else if storage, err := server.NewS3Storage(accessKey, secretKey, bucket, c.String("s3-endpoint"), logger); err != nil {
				panic(err)
			} else {
				options = append(options, server.UseStorage(storage))
			}
		case "gdrive":
			if clientJsonFilepath := c.String("gdrive-client-json-filepath"); clientJsonFilepath == "" {
				panic("client-json-filepath not set.")
			} else if localConfigPath := c.String("gdrive-local-config-path"); localConfigPath == "" {
				panic("local-config-path not set.")
			} else if basedir := c.String("basedir"); basedir == "" {
				panic("basedir not set.")
			} else if storage, err := server.NewGDriveStorage(clientJsonFilepath, localConfigPath, basedir, logger); err != nil {
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
