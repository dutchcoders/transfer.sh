/*
The MIT License (MIT)

Copyright (c) 2014-2017 DutchCoders [https://github.com/dutchcoders/]

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package server

import (
	crypto_rand "crypto/rand"
	"encoding/binary"
	"errors"
	gorillaHandlers "github.com/gorilla/handlers"
	"log"
	"math/rand"
	"mime"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	context "golang.org/x/net/context"

	"github.com/PuerkitoBio/ghost/handlers"
	"github.com/VojtechVitek/ratelimit"
	"github.com/VojtechVitek/ratelimit/memory"
	"github.com/gorilla/mux"

	// import pprof
	_ "net/http/pprof"

	"crypto/tls"

	web "github.com/dutchcoders/transfer.sh-web"
	assetfs "github.com/elazarl/go-bindata-assetfs"

	autocert "golang.org/x/crypto/acme/autocert"
	"path/filepath"
)

// parse request with maximum memory of _24Kilobits
const _24K = (1 << 3) * 24

// parse request with maximum memory of _5Megabytes
const _5M = (1 << 20) * 5

// OptionFn is the option function type
type OptionFn func(*Server)

// ClamavHost sets clamav host
func ClamavHost(s string) OptionFn {
	return func(srvr *Server) {
		srvr.ClamAVDaemonHost = s
	}
}

// VirustotalKey sets virus total key
func VirustotalKey(s string) OptionFn {
	return func(srvr *Server) {
		srvr.VirusTotalKey = s
	}
}

// Listener set listener
func Listener(s string) OptionFn {
	return func(srvr *Server) {
		srvr.ListenerString = s
	}

}

// CorsDomains sets CORS domains
func CorsDomains(s string) OptionFn {
	return func(srvr *Server) {
		srvr.CorsDomains = s
	}

}

// GoogleAnalytics sets GA key
func GoogleAnalytics(gaKey string) OptionFn {
	return func(srvr *Server) {
		srvr.gaKey = gaKey
	}
}

// UserVoice sets UV key
func UserVoice(userVoiceKey string) OptionFn {
	return func(srvr *Server) {
		srvr.userVoiceKey = userVoiceKey
	}
}

// TLSListener sets TLS listener and option
func TLSListener(s string, t bool) OptionFn {
	return func(srvr *Server) {
		srvr.TLSListenerString = s
		srvr.TLSListenerOnly = t
	}

}

// ProfileListener sets profile listener
func ProfileListener(s string) OptionFn {
	return func(srvr *Server) {
		srvr.ProfileListenerString = s
	}
}

// WebPath sets web path
func WebPath(s string) OptionFn {
	return func(srvr *Server) {
		if s[len(s)-1:] != "/" {
			s = s + string(filepath.Separator)
		}

		srvr.webPath = s
	}
}

// ProxyPath sets proxy path
func ProxyPath(s string) OptionFn {
	return func(srvr *Server) {
		if s[len(s)-1:] != "/" {
			s = s + string(filepath.Separator)
		}

		srvr.proxyPath = s
	}
}

// ProxyPort sets proxy port
func ProxyPort(s string) OptionFn {
	return func(srvr *Server) {
		srvr.proxyPort = s
	}
}

// TempPath sets temp path
func TempPath(s string) OptionFn {
	return func(srvr *Server) {
		if s[len(s)-1:] != "/" {
			s = s + string(filepath.Separator)
		}

		srvr.tempPath = s
	}
}

// LogFile sets log file
func LogFile(logger *log.Logger, s string) OptionFn {
	return func(srvr *Server) {
		f, err := os.OpenFile(s, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			logger.Fatalf("error opening file: %v", err)
		}

		logger.SetOutput(f)
		srvr.logger = logger
	}
}

// Logger sets logger
func Logger(logger *log.Logger) OptionFn {
	return func(srvr *Server) {
		srvr.logger = logger
	}
}

// MaxUploadSize sets max upload size
func MaxUploadSize(kbytes int64) OptionFn {
	return func(srvr *Server) {
		srvr.maxUploadSize = kbytes * 1024
	}

}

// RateLimit set rate limit
func RateLimit(requests int) OptionFn {
	return func(srvr *Server) {
		srvr.rateLimitRequests = requests
	}
}

// RandomTokenLength sets random token length
func RandomTokenLength(length int) OptionFn {
	return func(srvr *Server) {
		srvr.randomTokenLength = length
	}
}

// Purge sets purge days and option
func Purge(days, interval int) OptionFn {
	return func(srvr *Server) {
		srvr.purgeDays = time.Duration(days) * time.Hour * 24
		srvr.purgeInterval = time.Duration(interval) * time.Hour
	}
}

// ForceHTTPS sets forcing https
func ForceHTTPS() OptionFn {
	return func(srvr *Server) {
		srvr.forceHTTPS = true
	}
}

// EnableProfiler sets enable profiler
func EnableProfiler() OptionFn {
	return func(srvr *Server) {
		srvr.profilerEnabled = true
	}
}

// UseStorage set storage to use
func UseStorage(s Storage) OptionFn {
	return func(srvr *Server) {
		srvr.storage = s
	}
}

// UseLetsEncrypt set letsencrypt usage
func UseLetsEncrypt(hosts []string) OptionFn {
	return func(srvr *Server) {
		cacheDir := "./cache/"

		m := autocert.Manager{
			Prompt: autocert.AcceptTOS,
			Cache:  autocert.DirCache(cacheDir),
			HostPolicy: func(_ context.Context, host string) error {
				found := false

				for _, h := range hosts {
					found = found || strings.HasSuffix(host, h)
				}

				if !found {
					return errors.New("acme/autocert: host not configured")
				}

				return nil
			},
		}

		srvr.tlsConfig = &tls.Config{
			GetCertificate: m.GetCertificate,
		}
	}
}

// TLSConfig sets TLS config
func TLSConfig(cert, pk string) OptionFn {
	certificate, err := tls.LoadX509KeyPair(cert, pk)
	return func(srvr *Server) {
		srvr.tlsConfig = &tls.Config{
			GetCertificate: func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
				return &certificate, err
			},
		}
	}
}

// HTTPAuthCredentials sets basic http auth credentials
func HTTPAuthCredentials(user string, pass string) OptionFn {
	return func(srvr *Server) {
		srvr.AuthUser = user
		srvr.AuthPass = pass
	}
}

// FilterOptions sets ip filtering
func FilterOptions(options IPFilterOptions) OptionFn {
	for i, allowedIP := range options.AllowedIPs {
		options.AllowedIPs[i] = strings.TrimSpace(allowedIP)
	}

	for i, blockedIP := range options.BlockedIPs {
		options.BlockedIPs[i] = strings.TrimSpace(blockedIP)
	}

	return func(srvr *Server) {
		srvr.ipFilterOptions = &options
	}
}

// Server is the main application
type Server struct {
	AuthUser string
	AuthPass string

	logger *log.Logger

	tlsConfig *tls.Config

	profilerEnabled bool

	locks sync.Map

	maxUploadSize     int64
	rateLimitRequests int

	purgeDays     time.Duration
	purgeInterval time.Duration

	storage Storage

	forceHTTPS bool

	randomTokenLength int

	ipFilterOptions *IPFilterOptions

	VirusTotalKey    string
	ClamAVDaemonHost string

	tempPath string

	webPath      string
	proxyPath    string
	proxyPort    string
	gaKey        string
	userVoiceKey string

	TLSListenerOnly bool

	CorsDomains           string
	ListenerString        string
	TLSListenerString     string
	ProfileListenerString string

	Certificate string

	LetsEncryptCache string
}

// New is the factory fot Server
func New(options ...OptionFn) (*Server, error) {
	s := &Server{
		locks: sync.Map{},
	}

	for _, optionFn := range options {
		optionFn(s)
	}

	return s, nil
}

func init() {
	var seedBytes [8]byte
	if _, err := crypto_rand.Read(seedBytes[:]); err != nil {
		panic("cannot obtain cryptographically secure seed")
	}
	rand.Seed(int64(binary.LittleEndian.Uint64(seedBytes[:])))
}

// Run starts Server
func (s *Server) Run() {
	listening := false

	if s.profilerEnabled {
		listening = true

		go func() {
			s.logger.Println("Profiled listening at: :6060")

			http.ListenAndServe(":6060", nil)
		}()
	}

	r := mux.NewRouter()

	var fs http.FileSystem

	if s.webPath != "" {
		s.logger.Println("Using static file path: ", s.webPath)

		fs = http.Dir(s.webPath)

		htmlTemplates, _ = htmlTemplates.ParseGlob(s.webPath + "*.html")
		textTemplates, _ = textTemplates.ParseGlob(s.webPath + "*.txt")
	} else {
		fs = &assetfs.AssetFS{
			Asset:    web.Asset,
			AssetDir: web.AssetDir,
			AssetInfo: func(path string) (os.FileInfo, error) {
				return os.Stat(path)
			},
			Prefix: web.Prefix,
		}

		for _, path := range web.AssetNames() {
			bytes, err := web.Asset(path)
			if err != nil {
				s.logger.Panicf("Unable to parse: path=%s, err=%s", path, err)
			}

			htmlTemplates.New(stripPrefix(path)).Parse(string(bytes))
			textTemplates.New(stripPrefix(path)).Parse(string(bytes))
		}
	}

	staticHandler := http.FileServer(fs)

	r.PathPrefix("/images/").Handler(staticHandler).Methods("GET")
	r.PathPrefix("/styles/").Handler(staticHandler).Methods("GET")
	r.PathPrefix("/scripts/").Handler(staticHandler).Methods("GET")
	r.PathPrefix("/fonts/").Handler(staticHandler).Methods("GET")
	r.PathPrefix("/ico/").Handler(staticHandler).Methods("GET")
	r.HandleFunc("/favicon.ico", staticHandler.ServeHTTP).Methods("GET")
	r.HandleFunc("/robots.txt", staticHandler.ServeHTTP).Methods("GET")

	r.HandleFunc("/{filename:(?:favicon\\.ico|robots\\.txt|health\\.html)}", s.basicAuthHandler(http.HandlerFunc(s.putHandler))).Methods("PUT")

	r.HandleFunc("/health.html", healthHandler).Methods("GET")
	r.HandleFunc("/", s.viewHandler).Methods("GET")

	r.HandleFunc("/({files:.*}).zip", s.zipHandler).Methods("GET")
	r.HandleFunc("/({files:.*}).tar", s.tarHandler).Methods("GET")
	r.HandleFunc("/({files:.*}).tar.gz", s.tarGzHandler).Methods("GET")

	r.HandleFunc("/{token}/{filename}", s.headHandler).Methods("HEAD")
	r.HandleFunc("/{action:(?:download|get|inline)}/{token}/{filename}", s.headHandler).Methods("HEAD")

	r.HandleFunc("/{token}/{filename}", s.previewHandler).MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) (match bool) {
		match = false

		// The file will show a preview page when opening the link in browser directly or
		// from external link. If the referer url path and current path are the same it will be
		// downloaded.
		if !acceptsHTML(r.Header) {
			return false
		}

		match = (r.Referer() == "")

		u, err := url.Parse(r.Referer())
		if err != nil {
			s.logger.Fatal(err)
			return
		}

		match = match || (u.Path != r.URL.Path)
		return
	}).Methods("GET")

	getHandlerFn := s.getHandler
	if s.rateLimitRequests > 0 {
		getHandlerFn = ratelimit.Request(ratelimit.IP).Rate(s.rateLimitRequests, 60*time.Second).LimitBy(memory.New())(http.HandlerFunc(getHandlerFn)).ServeHTTP
	}

	r.HandleFunc("/{token}/{filename}", getHandlerFn).Methods("GET")
	r.HandleFunc("/{action:(?:download|get|inline)}/{token}/{filename}", getHandlerFn).Methods("GET")

	r.HandleFunc("/{filename}/virustotal", s.virusTotalHandler).Methods("PUT")
	r.HandleFunc("/{filename}/scan", s.scanHandler).Methods("PUT")
	r.HandleFunc("/put/{filename}", s.basicAuthHandler(http.HandlerFunc(s.putHandler))).Methods("PUT")
	r.HandleFunc("/upload/{filename}", s.basicAuthHandler(http.HandlerFunc(s.putHandler))).Methods("PUT")
	r.HandleFunc("/{filename}", s.basicAuthHandler(http.HandlerFunc(s.putHandler))).Methods("PUT")
	r.HandleFunc("/", s.basicAuthHandler(http.HandlerFunc(s.postHandler))).Methods("POST")
	// r.HandleFunc("/{page}", viewHandler).Methods("GET")

	r.HandleFunc("/{token}/{filename}/{deletionToken}", s.deleteHandler).Methods("DELETE")

	r.NotFoundHandler = http.HandlerFunc(s.notFoundHandler)

	mime.AddExtensionType(".md", "text/x-markdown")

	s.logger.Printf("Transfer.sh server started.\nusing temp folder: %s\nusing storage provider: %s", s.tempPath, s.storage.Type())

	var cors func(http.Handler) http.Handler
	if len(s.CorsDomains) > 0 {
		cors = gorillaHandlers.CORS(
			gorillaHandlers.AllowedHeaders([]string{"*"}),
			gorillaHandlers.AllowedOrigins(strings.Split(s.CorsDomains, ",")),
			gorillaHandlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"}),
		)
	} else {
		cors = func(h http.Handler) http.Handler {
			return h
		}
	}

	h := handlers.PanicHandler(
		ipFilterHandler(
			handlers.LogHandler(
				LoveHandler(
					s.RedirectHandler(cors(r))),
				handlers.NewLogOptions(s.logger.Printf, "_default_"),
			),
			s.ipFilterOptions,
		),
		nil,
	)

	if !s.TLSListenerOnly {
		srvr := &http.Server{
			Addr:    s.ListenerString,
			Handler: h,
		}

		listening = true
		s.logger.Printf("listening on port: %v\n", s.ListenerString)

		go func() {
			srvr.ListenAndServe()
		}()
	}

	if s.TLSListenerString != "" {
		listening = true
		s.logger.Printf("listening on port: %v\n", s.TLSListenerString)

		go func() {
			s := &http.Server{
				Addr:      s.TLSListenerString,
				Handler:   h,
				TLSConfig: s.tlsConfig,
			}

			if err := s.ListenAndServeTLS("", ""); err != nil {
				panic(err)
			}
		}()
	}

	s.logger.Printf("---------------------------")

	if s.purgeDays > 0 {
		go s.purgeHandler()
	}

	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt)
	signal.Notify(term, syscall.SIGTERM)

	if listening {
		<-term
	} else {
		s.logger.Printf("No listener active.")
	}

	s.logger.Printf("Server stopped.")
}
