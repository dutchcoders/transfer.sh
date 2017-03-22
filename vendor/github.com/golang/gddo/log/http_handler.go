package log

import (
	"encoding/hex"
	"math/rand"
	"net/http"
	"time"

	"github.com/inconshreveable/log15"
)

const (
	gaeRequestIDHeader = "X-AppEngine-Request-Log-Id"
)

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

type httpContextHandler struct {
	log         log15.Logger
	next        http.Handler
	onAppEngine bool
}

// NewHTTPContextHandler adds a context logger based on the given logger to
// each request. After a request passes through this handler,
// Error(req.Context(), "foo") will log to that logger and add useful context
// to each log entry.
func NewHTTPContextHandler(h http.Handler, l log15.Logger, onAppEngine bool) http.Handler {
	if l == nil {
		l = log15.Root()
	}

	return &httpContextHandler{
		log:         l,
		next:        h,
		onAppEngine: onAppEngine,
	}
}

func (h *httpContextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// We will accept an App Engine Request Header. If there isn't one, we will
	// fallback to 16 random bytes (hex encoded).
	reqID := r.Header.Get(gaeRequestIDHeader)
	if !h.onAppEngine || reqID == "" {
		buf := make([]byte, 16)
		random.Read(buf)
		reqID = hex.EncodeToString(buf)
	}

	requestLogger := h.log.New(log15.Ctx{
		"request_id": reqID,
	})

	r = r.WithContext(NewContext(ctx, requestLogger))

	h.next.ServeHTTP(w, r)
}
