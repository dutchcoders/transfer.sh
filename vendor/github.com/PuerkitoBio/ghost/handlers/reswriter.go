package handlers

import (
	"net/http"
)

// This interface can be implemented by an augmented ResponseWriter, so that
// it doesn't hide other augmented writers in the chain.
type WrapWriter interface {
	http.ResponseWriter
	WrappedWriter() http.ResponseWriter
}

// Helper function to retrieve a specific ResponseWriter.
func GetResponseWriter(w http.ResponseWriter,
	predicate func(http.ResponseWriter) bool) (http.ResponseWriter, bool) {

	for {
		// Check if this writer is the one we're looking for
		if w != nil && predicate(w) {
			return w, true
		}
		// If it is a WrapWriter, move back the chain of wrapped writers
		ww, ok := w.(WrapWriter)
		if !ok {
			return nil, false
		}
		w = ww.WrappedWriter()
	}
}
