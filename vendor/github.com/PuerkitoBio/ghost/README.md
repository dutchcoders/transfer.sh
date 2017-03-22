# Ghost

Ghost is a web development library loosely inspired by node's [Connect library][connect]. It provides a number of simple, single-responsibility HTTP handlers that can be combined to build a full-featured web server, and a generic template engine integration interface.

It stays close to the metal, not abstracting Go's standard library away. As a matter of fact, any stdlib handler can be used with Ghost's handlers, they simply are `net/http.Handler`'s.

## Installation and documentation

`go get github.com/PuerkitoBio/ghost`

[API reference][godoc]

*Status* : **Unmaintained**

## Example

See the /ghostest directory for a complete working example of a website built with Ghost. It shows all handlers and template support of Ghost.

## Handlers

Ghost offers the following handlers:

* BasicAuthHandler : basic authentication support.
* ContextHandler : key-value map provider for the duration of the request.
* FaviconHandler : simple and efficient favicon renderer.
* GZIPHandler : gzip-compresser for the body of the response.
* LogHandler : fully customizable request logger.
* PanicHandler : panic-catching handler to control the error response.
* SessionHandler : store-agnostic server-side session provider.
* StaticHandler : convenience handler that wraps a call to `net/http.ServeFile`.

Two stores are provided for the session persistence, `MemoryStore`, an in-memory map that is not suited for production environment, and `RedisStore`, a more robust and scalable [redigo][]-based Redis store. Because of the generic `SessionStore` interface, custom stores can easily be created as needed.

The `handlers` package also offers the `ChainableHandler` interface, which supports combining HTTP handlers in a sequential fashion, and the `ChainHandlers()` function that creates a new handler from the sequential combination of any number of handlers.

As a convenience, all functions that take a `http.Handler` as argument also have a corresponding function with the `Func` suffix that take a `http.HandlerFunc` instead as argument. This saves the type-cast when a simple handler function is passed (for example, `SessionHandler()` and `SessionHandlerFunc()`).

### Handlers Design

The HTTP handlers such as Basic Auth and Context need to store some state information to provide their functionality. Instead of using variables and a mutex to control shared access, Ghost augments the `http.ResponseWriter` interface that is part of the Handler's `ServeHTTP()` function signature. Because this instance is unique for each request and is not shared, there is no locking involved to access the state information.

However, when combining such handlers, Ghost needs a way to move through the chain of augmented ResponseWriters. This is why these *augmented writers* need to implement the `WrapWriter` interface. A single method is required, `WrappedWriter() http.ResponseWriter`, which returns the wrapped ResponseWriter.

And to get back a specific augmented writer, the `GetResponseWriter()` function is provided. It takes a ResponseWriter and a predicate function as argument, and returns the requested specific writer using the *comma-ok* pattern. Example, for the session writer:

```Go
func getSessionWriter(w http.ResponseWriter) (*sessResponseWriter, bool) {
	ss, ok := GetResponseWriter(w, func(tst http.ResponseWriter) bool {
		_, ok := tst.(*sessResponseWriter)
		return ok
	})
	if ok {
		return ss.(*sessResponseWriter), true
	}
	return nil, false
}
```

Ghost does not provide a muxer, there are already many great ones available, but I would recommend Go's native `http.ServeMux` or [pat][] because it has great features and plays well with Ghost's design. Gorilla's muxer is very popular, but since it depends on Gorilla's (mutex-based) context provider, this is redundant with Ghost's context.

## Templates

Ghost supports the following template engines:

* Go's native templates (needs work, at the moment does not work with nested templates)
* [Amber][]

TODO : Go's mustache implementation.

### Templates Design

The template engines can be registered much in the same way as database drivers, just by importing for side effects (using `_ "import/path"`). The `init()` function of the template engine's package registers the template compiler with the correct file extension, and the engine can be used.

## License

The [BSD 3-Clause license][lic].

[connect]: https://github.com/senchalabs/connect
[godoc]: http://godoc.org/github.com/PuerkitoBio/ghost
[lic]: http://opensource.org/licenses/BSD-3-Clause
[redigo]: https://github.com/garyburd/redigo
[pat]: https://github.com/bmizerany/pat
[amber]: https://github.com/eknkc/amber
