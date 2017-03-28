package ratelimit

import (
	"net"
	"net/http"
	"strings"
)

// IP returns unique key per request IP.
func IP(r *http.Request) string {
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.IndexAny(xff, ",;"); i != -1 {
			xff = xff[:i]
		}
		ip += "," + xff
	}
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		ip += "," + xrip
	}
	return ip
}

// NOP returns empty key for each request.
func NOP(r *http.Request) string {
	return ""
}
