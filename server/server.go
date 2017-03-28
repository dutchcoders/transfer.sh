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
	"errors"
	"fmt"
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

	_ "net/http/pprof"

	"crypto/tls"

	web "github.com/dutchcoders/transfer.sh-web"
	assetfs "github.com/elazarl/go-bindata-assetfs"

	autocert "golang.org/x/crypto/acme/autocert"
)

const SERVER_INFO = "transfer.sh"

// parse request with maximum memory of _24Kilobits
const _24K = (1 << 20) * 24

type OptionFn func(*Server)

func ClamavHost(s string) OptionFn {
	return func(srvr *Server) {
		srvr.ClamAVDaemonHost = s
	}
}

func VirustotalKey(s string) OptionFn {
	return func(srvr *Server) {
		srvr.VirusTotalKey = s
	}
}

func Listener(s string) OptionFn {
	return func(srvr *Server) {
		srvr.ListenerString = s
	}

}

func TLSListener(s string) OptionFn {
	return func(srvr *Server) {
		srvr.TLSListenerString = s
	}

}

func ProfileListener(s string) OptionFn {
	return func(srvr *Server) {
		srvr.ProfileListenerString = s
	}
}

func WebPath(s string) OptionFn {
	return func(srvr *Server) {
		srvr.webPath = s
	}
}

func TempPath(s string) OptionFn {
	return func(srvr *Server) {
		srvr.tempPath = s
	}
}

func LogFile(s string) OptionFn {
	return func(srvr *Server) {
		f, err := os.OpenFile(s, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}

		log.SetOutput(f)
	}
}

func RateLimit(requests int) OptionFn {
	return func(srvr *Server) {
		srvr.rateLimitRequests = requests
	}
}

func ForceHTTPs() OptionFn {
	return func(srvr *Server) {
		srvr.forceHTTPs = true
	}
}

func EnableProfiler() OptionFn {
	return func(srvr *Server) {
		srvr.profilerEnabled = true
	}
}

func UseStorage(s Storage) OptionFn {
	return func(srvr *Server) {
		srvr.storage = s
	}
}

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

type Server struct {
	tlsConfig *tls.Config

	profilerEnabled bool

	locks map[string]*sync.Mutex

	rateLimitRequests int

	storage Storage

	forceHTTPs bool

	VirusTotalKey    string
	ClamAVDaemonHost string

	tempPath string

	webPath string

	ListenerString        string
	TLSListenerString     string
	ProfileListenerString string

	Certificate string

	LetsEncryptCache string
}

func New(options ...OptionFn) (*Server, error) {
	s := &Server{
		locks: map[string]*sync.Mutex{},
	}

	for _, optionFn := range options {
		optionFn(s)
	}

	return s, nil
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func (s *Server) Run() {
	if s.profilerEnabled {
		go func() {
			fmt.Println("Profiled listening at: :6060")

			http.ListenAndServe(":6060", nil)
		}()
	}

	r := mux.NewRouter()

	var fs http.FileSystem

	if s.webPath != "" {
		log.Println("Using static file path: ", s.webPath)

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
				log.Panicf("Unable to parse: path=%s, err=%s", path, err)
			}

			htmlTemplates.New(stripPrefix(path)).Parse(string(bytes))
			textTemplates.New(stripPrefix(path)).Parse(string(bytes))
		}
	}

	staticHandler := http.FileServer(fs)

	r.PathPrefix("/images/").Handler(staticHandler)
	r.PathPrefix("/styles/").Handler(staticHandler)
	r.PathPrefix("/scripts/").Handler(staticHandler)
	r.PathPrefix("/fonts/").Handler(staticHandler)
	r.PathPrefix("/ico/").Handler(staticHandler)
	r.PathPrefix("/favicon.ico").Handler(staticHandler)
	r.PathPrefix("/robots.txt").Handler(staticHandler)

	r.HandleFunc("/health.html", healthHandler).Methods("GET")
	r.HandleFunc("/", s.viewHandler).Methods("GET")

	r.HandleFunc("/({files:.*}).zip", s.zipHandler).Methods("GET")
	r.HandleFunc("/({files:.*}).tar", s.tarHandler).Methods("GET")
	r.HandleFunc("/({files:.*}).tar.gz", s.tarGzHandler).Methods("GET")

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
			log.Fatal(err)
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
	r.HandleFunc("/get/{token}/{filename}", getHandlerFn).Methods("GET")
	r.HandleFunc("/download/{token}/{filename}", getHandlerFn).Methods("GET")

	r.HandleFunc("/{filename}/virustotal", s.virusTotalHandler).Methods("PUT")
	r.HandleFunc("/{filename}/scan", s.scanHandler).Methods("PUT")
	r.HandleFunc("/put/{filename}", s.putHandler).Methods("PUT")
	r.HandleFunc("/upload/{filename}", s.putHandler).Methods("PUT")
	r.HandleFunc("/{filename}", s.putHandler).Methods("PUT")
	r.HandleFunc("/", s.postHandler).Methods("POST")
	// r.HandleFunc("/{page}", viewHandler).Methods("GET")

	r.NotFoundHandler = http.HandlerFunc(s.notFoundHandler)

	mime.AddExtensionType(".md", "text/x-markdown")

	log.Printf("Transfer.sh server started.\nlistening on port: %v\nusing temp folder: %s\nusing storage provider: %s", s.ListenerString, s.tempPath, s.storage.Type())
	log.Printf("---------------------------")

	h := handlers.PanicHandler(handlers.LogHandler(LoveHandler(s.RedirectHandler(r)), handlers.NewLogOptions(log.Printf, "_default_")), nil)

	srvr := &http.Server{
		Addr:    s.ListenerString,
		Handler: h,
	}

	go func() {
		srvr.ListenAndServe()
	}()

	if s.TLSListenerString != "" {
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

	/*
		cacheDir := "/var/cache/autocert"

		if s.LetsEncryptCache != "" {
			cacheDir = s.LetsEncryptCache
		}

		m := autocert.Manager{
			Prompt: autocert.AcceptTOS,
			Cache:  autocert.DirCache(cacheDir),
			HostPolicy: func(_ context.Context, host string) error {
				if !strings.HasSuffix(host, "transfer.sh") {
					return errors.New("acme/autocert: host not configured")
				}
				return nil
			},
		}

		if s.TLSListenerString != "" {
			go func() {
				s := &http.Server{
					Addr:      ":https",
					Handler:   lh,
					TLSConfig: &tls.Config{GetCertificate: m.GetCertificate},
				}

				if err := s.ListenAndServeTLS("", ""); err != nil {
					panic(err)
				}
			}()

			if err := http.ListenAndServe(c.ListenerString, RedirectHandler()); err != nil {
				panic(err)
			}
		}
	*/

	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt)
	signal.Notify(term, syscall.SIGTERM)

	<-term

	log.Printf("Server stopped.")
}
