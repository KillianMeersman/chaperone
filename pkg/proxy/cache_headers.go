package proxy

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/KillianMeersman/chaperone/pkg/datastructures"
	"github.com/KillianMeersman/chaperone/pkg/log"
)

var UncachableHeaderValues = datastructures.NewSet("no-store", "no-cache")

type CachePolicy struct {
	FreshUntil           time.Duration
	CanUseStale          bool
	NoCache              bool
	NoStore              bool
	StaleWhileRevalidate bool
}

// type CachePolicy interface {
// 	MustRevalidate(url *url.URL) bool
// }

// Attempt to parse a Cache-Control header, returning the duration the associated
// response is allowed to be cached.
// Returns default duration if header could not be parsed.
func ParseCacheControl(header string, defaultDuration time.Duration) time.Duration {
	ttl := defaultDuration

	directives := strings.Split(header, ",")
	for _, directive := range directives {
		directive = strings.Trim(directive, " ")
		directive = strings.ToLower(directive)
		parts := strings.SplitN(directive, "=", 2)
		directive = parts[0]

		// check if no-cache or no-store
		if UncachableHeaderValues.Contains(directive) {
			return 0
		}

		switch directive {
		case "max-age", "s-max-age":
			if len(parts) < 2 {
				log.DefaultLogger.With("header", header).Error("Invalid cache-control header")
				continue
			}
			seconds, err := strconv.Atoi(parts[1])
			if err != nil {
				continue
			}
			ttl = time.Duration(seconds) * time.Second

		}
	}

	return ttl
}

// Attempt to parse the provided Expires header, returning the duration the associated
// response is allowed to be cached.
// Returns default duration if header could not be parsed.
func ParseExpiresHeader(header string, defaultDuration time.Duration) time.Duration {
	date, err := time.Parse(time.RFC1123, header)
	if err == nil {
		return time.Until(date)
	}

	return defaultDuration
}

// Calculate how long we can cache the response based on headers & other response parameters.
func GetResponseCacheDuration(res *http.Response, defaultDuration time.Duration) (time.Duration, error) {
	headers := res.Header
	if cacheControl := headers.Get("Cache-Control"); cacheControl != "" {
		return ParseCacheControl(cacheControl, defaultDuration), nil
	} else if expires := headers.Get("Expires"); expires != "" {
		return ParseExpiresHeader(expires, defaultDuration), nil
	}

	// HTTP 301: Moved permanently is cached for a very long time if no other caching headers are present.
	if res.StatusCode == 301 {
		return 365 * 24 * time.Hour, nil
	}

	return defaultDuration, nil
}

// Get the header names upon which the document at the provided url varies.
// This depends on the 'Vary' header.
func GetVaryHeaderNames(res *http.Response) []string {
	varyHeaders := make([]string, 0)
	for _, varyHeader := range strings.Split(res.Header.Get("Vary"), ",") {
		varyHeaders = append(varyHeaders, strings.TrimSpace(varyHeader))
	}

	return varyHeaders
}

// Get a cache key for the provided vary headers and request.
func GetCacheKey(varyHeaders []string, req *http.Request) string {
	requestVaryHeaderValues := datastructures.MapCopy(varyHeaders, func(el string) string {
		return fmt.Sprintf("%s=%s", el, req.Header.Get(el))
	})

	return fmt.Sprintf("%s:%s", req.URL.String(), strings.Join(requestVaryHeaderValues, ","))
}
