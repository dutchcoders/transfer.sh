// Package handlers define reusable handler components that focus on offering
// a single well-defined feature. Note that any http.Handler implementation
// can be used with Ghost's chainable or wrappable handlers design.
//
// Go's standard library provides a number of such useful handlers in net/http:
//
// - FileServer(http.FileSystem)
// - NotFoundHandler()
// - RedirectHandler(string, int)
// - StripPrefix(string, http.Handler)
// - TimeoutHandler(http.Handler, time.Duration, string)
//
// This package adds the following list of handlers:
//
// - BasicAuthHandler(http.Handler, func(string, string) (interface{}, bool), string)
// a Basic Authentication handler.
// - ContextHandler(http.Handler, int) : a volatile storage map valid only
// for the duration of the request, with no locking required.
// - FaviconHandler(http.Handler, string, time.Duration) : an efficient favicon
// handler.
// - GZIPHandler(http.Handler) : compress the content of the body if the client
// accepts gzip compression.
// - LogHandler(http.Handler, *LogOptions) : customizable request logger.
// - PanicHandler(http.Handler) : handle panics gracefully so that the client
// receives a response (status code 500).
// - SessionHandler(http.Handler, *SessionOptions) : a cookie-based, store-agnostic
// persistent session handler.
// - StaticFileHandler(string) : serve the contents of a specific file.
package handlers
