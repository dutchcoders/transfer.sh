package handlers

import (
	"encoding/json"
	"errors"
	"hash/crc32"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/ghost"
	"github.com/gorilla/securecookie"
	"github.com/nu7hatch/gouuid"
)

const defaultCookieName = "ghost.sid"

var (
	ErrSessionSecretMissing = errors.New("session secret is missing")
	ErrNoSessionID          = errors.New("session ID could not be generated")
)

// The Session holds the data map that persists for the duration of the session.
// The information stored in this map should be marshalable for the target Session store
// format (i.e. json, sql, gob, etc. depending on how the store persists the data).
type Session struct {
	isNew bool // keep private, not saved to JSON, will be false once read from the store
	internalSession
}

// Use a separate private struct to hold the private fields of the Session,
// although those fields are exposed (public). This is a trick to simplify
// JSON encoding.
type internalSession struct {
	Data    map[string]interface{} // JSON cannot marshal a map[interface{}]interface{}
	ID      string
	Created time.Time
	MaxAge  time.Duration
}

// Create a new Session instance. It panics in the unlikely event that a new random ID cannot be generated.
func newSession(maxAge int) *Session {
	uid, err := uuid.NewV4()
	if err != nil {
		panic(ErrNoSessionID)
	}
	return &Session{
		true, // is new
		internalSession{
			make(map[string]interface{}),
			uid.String(),
			time.Now(),
			time.Duration(maxAge) * time.Second,
		},
	}
}

// Gets the ID of the session.
func (ø *Session) ID() string {
	return ø.internalSession.ID
}

// Get the max age duration
func (ø *Session) MaxAge() time.Duration {
	return ø.internalSession.MaxAge
}

// Get the creation time of the session.
func (ø *Session) Created() time.Time {
	return ø.internalSession.Created
}

// Is this a new Session (created by the current request)
func (ø *Session) IsNew() bool {
	return ø.isNew
}

// TODO : Resets the max age property of the session to its original value (sliding expiration).
func (ø *Session) resetMaxAge() {
}

// Marshal the session to JSON.
func (ø *Session) MarshalJSON() ([]byte, error) {
	return json.Marshal(ø.internalSession)
}

// Unmarshal the JSON into the internal session struct.
func (ø *Session) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &ø.internalSession)
}

// Options object for the session handler. It specified the Session store to use for
// persistence, the template for the session cookie (name, path, maxage, etc.),
// whether or not the proxy should be trusted to determine if the connection is secure,
// and the required secret to sign the session cookie.
type SessionOptions struct {
	Store          SessionStore
	CookieTemplate http.Cookie
	TrustProxy     bool
	Secret         string
}

// Create a new SessionOptions struct, using default cookie and proxy values.
func NewSessionOptions(store SessionStore, secret string) *SessionOptions {
	return &SessionOptions{
		Store:  store,
		Secret: secret,
	}
}

// The augmented ResponseWriter struct for the session handler. It holds the current
// Session object and Session store, as well as flags and function to send the actual
// session cookie at the end of the request.
type sessResponseWriter struct {
	http.ResponseWriter
	sess         *Session
	sessStore    SessionStore
	sessSent     bool
	sendCookieFn func()
}

// Implement the WrapWriter interface.
func (ø *sessResponseWriter) WrappedWriter() http.ResponseWriter {
	return ø.ResponseWriter
}

// Intercept the Write() method to add the Set-Cookie header before it's too late.
func (ø *sessResponseWriter) Write(data []byte) (int, error) {
	if !ø.sessSent {
		ø.sendCookieFn()
		ø.sessSent = true
	}
	return ø.ResponseWriter.Write(data)
}

// Intercept the WriteHeader() method to add the Set-Cookie header before it's too late.
func (ø *sessResponseWriter) WriteHeader(code int) {
	if !ø.sessSent {
		ø.sendCookieFn()
		ø.sessSent = true
	}
	ø.ResponseWriter.WriteHeader(code)
}

// SessionHandlerFunc is the same as SessionHandler, it is just a convenience
// signature that accepts a func(http.ResponseWriter, *http.Request) instead of
// a http.Handler interface. It saves the boilerplate http.HandlerFunc() cast.
func SessionHandlerFunc(h http.HandlerFunc, opts *SessionOptions) http.HandlerFunc {
	return SessionHandler(h, opts)
}

// Create a Session handler to offer the Session behaviour to the specified handler.
func SessionHandler(h http.Handler, opts *SessionOptions) http.HandlerFunc {
	// Make sure the required cookie fields are set
	if opts.CookieTemplate.Name == "" {
		opts.CookieTemplate.Name = defaultCookieName
	}
	if opts.CookieTemplate.Path == "" {
		opts.CookieTemplate.Path = "/"
	}
	// Secret is required
	if opts.Secret == "" {
		panic(ErrSessionSecretMissing)
	}

	// Return the actual handler
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := getSessionWriter(w); ok {
			// Self-awareness
			h.ServeHTTP(w, r)
			return
		}

		if strings.Index(r.URL.Path, opts.CookieTemplate.Path) != 0 {
			// Session does not apply to this path
			h.ServeHTTP(w, r)
			return
		}

		// Create a new Session or retrieve the existing session based on the
		// session cookie received.
		var sess *Session
		var ckSessId string
		exCk, err := r.Cookie(opts.CookieTemplate.Name)
		if err != nil {
			sess = newSession(opts.CookieTemplate.MaxAge)
			ghost.LogFn("ghost.session : error getting session cookie : %s", err)
		} else {
			ckSessId, err = parseSignedCookie(exCk, opts.Secret)
			if err != nil {
				sess = newSession(opts.CookieTemplate.MaxAge)
				ghost.LogFn("ghost.session : error parsing signed cookie : %s", err)
			} else if ckSessId == "" {
				sess = newSession(opts.CookieTemplate.MaxAge)
				ghost.LogFn("ghost.session : no existing session ID")
			} else {
				// Get the session
				sess, err = opts.Store.Get(ckSessId)
				if err != nil {
					sess = newSession(opts.CookieTemplate.MaxAge)
					ghost.LogFn("ghost.session : error getting session from store : %s", err)
				} else if sess == nil {
					sess = newSession(opts.CookieTemplate.MaxAge)
					ghost.LogFn("ghost.session : nil session")
				}
			}
		}
		// Save the original hash of the session, used to compare if the contents
		// have changed during the handling of the request, so that it has to be
		// saved to the stored.
		oriHash := hash(sess)

		// Create the augmented ResponseWriter.
		srw := &sessResponseWriter{w, sess, opts.Store, false, func() {
			// This function is called when the header is about to be written, so that
			// the session cookie is correctly set.

			// Check if the connection is secure
			proto := strings.Trim(strings.ToLower(r.Header.Get("X-Forwarded-Proto")), " ")
			tls := r.TLS != nil || (strings.HasPrefix(proto, "https") && opts.TrustProxy)
			if opts.CookieTemplate.Secure && !tls {
				ghost.LogFn("ghost.session : secure cookie on a non-secure connection, cookie not sent")
				return
			}
			if !sess.IsNew() {
				// If this is not a new session, no need to send back the cookie
				// TODO : Handle expires?
				return
			}

			// Send the session cookie
			ck := opts.CookieTemplate
			ck.Value = sess.ID()
			err := signCookie(&ck, opts.Secret)
			if err != nil {
				ghost.LogFn("ghost.session : error signing cookie : %s", err)
				return
			}
			http.SetCookie(w, &ck)
		}}

		// Call wrapped handler
		h.ServeHTTP(srw, r)

		// TODO : Expiration management? srw.sess.resetMaxAge()
		// Do not save if content is the same, unless session is new (to avoid
		// creating a new session and sending a cookie on each successive request).
		if newHash := hash(sess); !sess.IsNew() && oriHash == newHash && newHash != 0 {
			// No changes to the session, no need to save
			ghost.LogFn("ghost.session : no changes to save to store")
			return
		}
		err = opts.Store.Set(sess)
		if err != nil {
			ghost.LogFn("ghost.session : error saving session to store : %s", err)
		}
	}
}

// Helper function to retrieve the session for the current request.
func GetSession(w http.ResponseWriter) (*Session, bool) {
	ss, ok := getSessionWriter(w)
	if ok {
		return ss.sess, true
	}
	return nil, false
}

// Helper function to retrieve the session store
func GetSessionStore(w http.ResponseWriter) (SessionStore, bool) {
	ss, ok := getSessionWriter(w)
	if ok {
		return ss.sessStore, true
	}
	return nil, false
}

// Internal helper function to retrieve the session writer object.
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

// Parse a signed cookie and return the cookie value
func parseSignedCookie(ck *http.Cookie, secret string) (string, error) {
	var val string

	sck := securecookie.New([]byte(secret), nil)
	err := sck.Decode(ck.Name, ck.Value, &val)
	if err != nil {
		return "", err
	}
	return val, nil
}

// Sign the specified cookie's value
func signCookie(ck *http.Cookie, secret string) error {
	sck := securecookie.New([]byte(secret), nil)
	enc, err := sck.Encode(ck.Name, ck.Value)
	if err != nil {
		return err
	}
	ck.Value = enc
	return nil
}

// Compute a CRC32 hash of the session's JSON-encoded contents.
func hash(s *Session) uint32 {
	data, err := json.Marshal(s)
	if err != nil {
		ghost.LogFn("ghost.session : error hash : %s", err)
		return 0 // 0 is always treated as "modified" session content
	}
	return crc32.ChecksumIEEE(data)
}
