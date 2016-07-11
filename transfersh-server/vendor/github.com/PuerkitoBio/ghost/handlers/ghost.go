package handlers

import (
	"net/http"
)

// Interface giving easy access to the most common augmented features.
type GhostWriter interface {
	http.ResponseWriter
	UserName() string
	User() interface{}
	Context() map[interface{}]interface{}
	Session() *Session
}

// Internal implementation of the GhostWriter interface.
type ghostWriter struct {
	http.ResponseWriter
	userName string
	user     interface{}
	ctx      map[interface{}]interface{}
	ssn      *Session
}

func (this *ghostWriter) UserName() string {
	return this.userName
}

func (this *ghostWriter) User() interface{} {
	return this.user
}

func (this *ghostWriter) Context() map[interface{}]interface{} {
	return this.ctx
}

func (this *ghostWriter) Session() *Session {
	return this.ssn
}

// Convenience handler that wraps a custom function with direct access to the
// authenticated user, context and session on the writer.
func GhostHandlerFunc(h func(w GhostWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if gw, ok := getGhostWriter(w); ok {
			// Self-awareness
			h(gw, r)
			return
		}
		uid, _ := GetUserName(w)
		usr, _ := GetUser(w)
		ctx, _ := GetContext(w)
		ssn, _ := GetSession(w)
		gw := &ghostWriter{
			w,
			uid,
			usr,
			ctx,
			ssn,
		}
		h(gw, r)
	}
}

// Check the writer chain to find a ghostWriter.
func getGhostWriter(w http.ResponseWriter) (*ghostWriter, bool) {
	gw, ok := GetResponseWriter(w, func(tst http.ResponseWriter) bool {
		_, ok := tst.(*ghostWriter)
		return ok
	})
	if ok {
		return gw.(*ghostWriter), true
	}
	return nil, false
}
