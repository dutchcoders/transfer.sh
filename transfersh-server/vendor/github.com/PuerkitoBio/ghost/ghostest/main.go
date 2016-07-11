// Ghostest is an interactive end-to-end Web site application to test
// the ghost packages. It serves the following URLs, with the specified
// features (handlers):
//
// / : panic;log;gzip;static; -> serve file index.html
// /public/styles.css : panic;log;gzip;StripPrefix;FileServer; -> serve directory public/
// /public/script.js : panic;log;gzip;StripPrefix;FileServer; -> serve directory public/
// /public/logo.pn : panic;log;gzip;StripPrefix;FileServer; -> serve directory public/
// /session : panic;log;gzip;session;context;Custom; -> serve dynamic Go template
// /session/auth : panic;log;gzip;session;context;basicAuth;Custom; -> serve dynamic template
// /panic : panic;log;gzip;Custom; -> panics
// /context : panic;log;gzip;context;Custom1;Custom2; -> serve dynamic Amber template
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/PuerkitoBio/ghost/handlers"
	"github.com/PuerkitoBio/ghost/templates"
	_ "github.com/PuerkitoBio/ghost/templates/amber"
	_ "github.com/PuerkitoBio/ghost/templates/gotpl"
	"github.com/bmizerany/pat"
)

const (
	sessionPageTitle     = "Session Page"
	sessionPageAuthTitle = "Authenticated Session Page"
	sessionPageKey       = "txt"
	contextPageKey       = "time"
	sessionExpiration    = 10 // Session expires after 10 seconds
)

var (
	// Create the common session store and secret
	memStore = handlers.NewMemoryStore(1)
	secret   = "testimony of the ancients"
)

// The struct used to pass data to the session template.
type sessionPageInfo struct {
	SessionID string
	Title     string
	Text      string
}

// Authenticate the Basic Auth credentials.
func authenticate(u, p string) (interface{}, bool) {
	if u == "user" && p == "pwd" {
		return u + p, true
	}
	return nil, false
}

// Handle the session page requests.
func sessionPageRenderer(w handlers.GhostWriter, r *http.Request) {
	var (
		txt   interface{}
		data  sessionPageInfo
		title string
	)

	ssn := w.Session()
	if r.Method == "GET" {
		txt = ssn.Data[sessionPageKey]
	} else {
		txt = r.FormValue(sessionPageKey)
		ssn.Data[sessionPageKey] = txt
	}
	if r.URL.Path == "/session/auth" {
		title = sessionPageAuthTitle
	} else {
		title = sessionPageTitle
	}
	if txt != nil {
		data = sessionPageInfo{ssn.ID(), title, txt.(string)}
	} else {
		data = sessionPageInfo{ssn.ID(), title, "[nil]"}
	}
	err := templates.Render("templates/session.tmpl", w, data)
	if err != nil {
		panic(err)
	}
}

// Prepare the context value for the chained handlers context page.
func setContext(w handlers.GhostWriter, r *http.Request) {
	w.Context()[contextPageKey] = time.Now().String()
}

// Retrieve the context value and render the chained handlers context page.
func renderContextPage(w handlers.GhostWriter, r *http.Request) {
	err := templates.Render("templates/amber/context.amber",
		w, &struct{ Val string }{w.Context()[contextPageKey].(string)})
	if err != nil {
		panic(err)
	}
}

// Prepare the web server and kick it off.
func main() {
	// Blank the default logger's prefixes
	log.SetFlags(0)

	// Compile the dynamic templates (native Go templates and Amber
	// templates are both registered via the for-side-effects-only imports)
	err := templates.CompileDir("./templates/")
	if err != nil {
		panic(err)
	}

	// Set the simple routes for static files
	mux := pat.New()
	mux.Get("/", handlers.StaticFileHandler("./index.html"))
	mux.Get("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("./public/"))))

	// Set the more complex routes for session handling and dynamic page (same
	// handler is used for both GET and POST).
	ssnOpts := handlers.NewSessionOptions(memStore, secret)
	ssnOpts.CookieTemplate.MaxAge = sessionExpiration
	hSsn := handlers.SessionHandler(
		handlers.ContextHandlerFunc(
			handlers.GhostHandlerFunc(sessionPageRenderer),
			1),
		ssnOpts)
	mux.Get("/session", hSsn)
	mux.Post("/session", hSsn)

	hAuthSsn := handlers.BasicAuthHandler(hSsn, authenticate, "")
	mux.Get("/session/auth", hAuthSsn)
	mux.Post("/session/auth", hAuthSsn)

	// Set the handler for the chained context route
	mux.Get("/context", handlers.ContextHandler(handlers.ChainHandlerFuncs(
		handlers.GhostHandlerFunc(setContext),
		handlers.GhostHandlerFunc(renderContextPage)),
		1))

	// Set the panic route, which simply panics
	mux.Get("/panic", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			panic("explicit panic")
		}))

	// Combine the top level handlers, that wrap around the muxer.
	// Panic is the outermost, so that any panic is caught and responded to with a code 500.
	// Log is next, so that every request is logged along with the URL, status code and response time.
	// GZIP is then applied, so that content is compressed.
	// Finally, the muxer finds the specific handler that applies to the route.
	h := handlers.FaviconHandler(
		handlers.PanicHandler(
			handlers.LogHandler(
				handlers.GZIPHandler(
					mux,
					nil),
				handlers.NewLogOptions(nil, handlers.Ltiny)),
			nil),
		"./public/favicon.ico",
		48*time.Hour)

	// Assign the combined handler to the server.
	http.Handle("/", h)

	// Start it up.
	if err := http.ListenAndServe(":9000", nil); err != nil {
		panic(err)
	}
}
