package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang/gddo/database"
	"github.com/golang/gddo/log"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	gaeProjectEnvVar = "GAE_LONG_APP_ID"
)

const (
	ConfigMaxAge            = "max_age"
	ConfigGetTimeout        = "get_timeout"
	ConfigRobotThreshold    = "robot"
	ConfigAssetsDir         = "assets"
	ConfigFirstGetTimeout   = "first_get_timeout"
	ConfigBindAddress       = "http"
	ConfigProject           = "project"
	ConfigTrustProxyHeaders = "trust_proxy_headers"
	ConfigSidebar           = "sidebar"
	ConfigDefaultGOOS       = "default_goos"
	ConfigSourcegraphURL    = "sourcegraph_url"
	ConfigGithubInterval    = "github_interval"
	ConfigCrawlInterval     = "crawl_interval"
	ConfigDialTimeout       = "dial_timeout"
	ConfigRequestTimeout    = "request_timeout"
)

// Initialize configuration
func init() {
	ctx := context.Background()

	// Automatically detect if we are on App Engine.
	if os.Getenv(gaeProjectEnvVar) != "" {
		viper.Set("on_appengine", true)
	} else {
		viper.Set("on_appengine", false)
	}

	// Setup command line flags
	flags := buildFlags()
	flags.Parse(os.Args)
	if err := viper.BindPFlags(flags); err != nil {
		panic(err)
	}

	// Also fetch from enviorment
	viper.SetEnvPrefix("gddo")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	// Automatically get project ID from env on Google App Engine
	viper.BindEnv(ConfigProject, gaeProjectEnvVar)

	// Read from config.
	readViperConfig(ctx)

	log.Info(ctx, "config values loaded", "values", viper.AllSettings())
}

func buildFlags() *pflag.FlagSet {
	flags := pflag.NewFlagSet("default", pflag.ExitOnError)

	flags.StringP("config", "c", "", "path to motd config file")
	flags.String("project", "", "Google Cloud Platform project used for Google services")
	// TODO(stephenmw): flags.Bool("enable-admin-pages", false, "When true, enables /admin pages")
	flags.Float64(ConfigRobotThreshold, 100, "Request counter threshold for robots.")
	flags.String(ConfigAssetsDir, filepath.Join(defaultBase("github.com/golang/gddo/gddo-server"), "assets"), "Base directory for templates and static files.")
	flags.Duration(ConfigGetTimeout, 8*time.Second, "Time to wait for package update from the VCS.")
	flags.Duration(ConfigFirstGetTimeout, 5*time.Second, "Time to wait for first fetch of package from the VCS.")
	flags.Duration(ConfigMaxAge, 24*time.Hour, "Update package documents older than this age.")
	flags.String(ConfigBindAddress, ":8080", "Listen for HTTP connections on this address.")
	flags.Bool(ConfigSidebar, false, "Enable package page sidebar.")
	flags.String(ConfigDefaultGOOS, "", "Default GOOS to use when building package documents.")
	flags.Bool(ConfigTrustProxyHeaders, false, "If enabled, identify the remote address of the request using X-Real-Ip in header.")
	flags.String(ConfigSourcegraphURL, "https://sourcegraph.com", "Link to global uses on Sourcegraph based at this URL (no need for trailing slash).")
	flags.Duration(ConfigGithubInterval, 0, "Github updates crawler sleeps for this duration between fetches. Zero disables the crawler.")
	flags.Duration(ConfigCrawlInterval, 0, "Package updater sleeps for this duration between package updates. Zero disables updates.")
	flags.Duration(ConfigDialTimeout, 5*time.Second, "Timeout for dialing an HTTP connection.")
	flags.Duration(ConfigRequestTimeout, 20*time.Second, "Time out for roundtripping an HTTP request.")

	// TODO(stephenmw): pass these variables at database creation time.
	flags.StringVar(&database.RedisServer, "db-server", database.RedisServer, "URI of Redis server.")
	flags.DurationVar(&database.RedisIdleTimeout, "db-idle-timeout", database.RedisIdleTimeout, "Close Redis connections after remaining idle for this duration.")
	flags.BoolVar(&database.RedisLog, "db-log", database.RedisLog, "Log database commands")

	return flags
}

// readViperConfig finds and then parses a config file. It will log.Fatal if the
// config file was specified or could not parse. Otherwise it will only warn
// that it failed to load the config.
func readViperConfig(ctx context.Context) {
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc")
	viper.SetConfigName("gddo")
	if viper.GetString("config") != "" {
		viper.SetConfigFile(viper.GetString("config"))
	}

	if err := viper.ReadInConfig(); err != nil {
		// If a config exists but could not be parsed, we should bail.
		if _, ok := err.(viper.ConfigParseError); ok {
			log.Fatal(ctx, "failed to parse config", "error", err)
		}

		// If the user specified a config file location in flags or env and
		// we failed to load it, we should bail. If not, it is just a warning.
		if viper.GetString("config") != "" {
			log.Fatal(ctx, "failed to load configuration file", "error", err)
		} else {
			log.Warn(ctx, "failed to load configuration file", "error", err)
		}
	} else {
		log.Info(ctx, "loaded configuration file successfully", "path", viper.ConfigFileUsed())
	}
}
