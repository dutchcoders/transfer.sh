package handlers

import (
	"net/http"
	"strings"
)

// Kind of match to apply to the header check.
type HeaderMatchType int

const (
	HmEquals HeaderMatchType = iota
	HmStartsWith
	HmEndsWith
	HmContains
)

// Check if the specified header matches the test string, applying the header match type
// specified.
func HeaderMatch(hdr http.Header, nm string, matchType HeaderMatchType, test string) bool {
	// First get the header value
	val := hdr[http.CanonicalHeaderKey(nm)]
	if len(val) == 0 {
		return false
	}
	// Prepare the match test
	test = strings.ToLower(test)
	for _, v := range val {
		v = strings.Trim(strings.ToLower(v), " \n\t")
		switch matchType {
		case HmEquals:
			if v == test {
				return true
			}
		case HmStartsWith:
			if strings.HasPrefix(v, test) {
				return true
			}
		case HmEndsWith:
			if strings.HasSuffix(v, test) {
				return true
			}
		case HmContains:
			if strings.Contains(v, test) {
				return true
			}
		}
	}
	return false
}
