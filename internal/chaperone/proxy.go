package chaperone

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/KillianMeersman/chaperone/pkg/log"
	"github.com/KillianMeersman/chaperone/pkg/proxy"
)

// The Chaperone proxy.
// Note that this is an http-only proxy, as such, it does not support https CONNECT requests.
// CONNECT requests would turn the proxy into a tcp relay, unaware of the http stream. This would prevent it from caching or rate-limiting based on url.
// ----
// To proxy https requests, change your request scheme to http and add the X-Upgrade-HTTPS=true header.
// Will fetch and respect the robots.txt file by default, use the X-Respect-Robots=false header to turn this off.
type ChaperoneProxy struct {
	client *proxy.NiceClient
	config *ConfigFile
}

func (p *ChaperoneProxy) Start(ctx context.Context) error {
	throttle := proxy.NewMemoryHTTPThrottle(time.Second)
	cache := proxy.NewMemoryHTTPCache(512e6)
	p.client = proxy.NewNiceClient(ctx, http.DefaultTransport, throttle, cache)

	configFile, err := ParseConfigFile(ConfigFileLocation)
	if err != nil {
		log.DefaultLogger.Fatal(err.Error())
	}

	p.config = configFile

	for _, rateLimit := range configFile.RateLimits {
		url, err := url.Parse(rateLimit.URL)
		if err != nil {
			panic(err)
		}
		log.DefaultLogger.Info("Setting throttle for url", "url", rateLimit.URL, "method", rateLimit.Method, "wait_time", rateLimit.WaitDuration.String())
		throttle.SetThrottle(&http.Request{
			Method: rateLimit.Method,
			URL:    url,
		}, rateLimit.WaitDuration)
	}

	listenAddr := fmt.Sprintf("0.0.0.0:%d", Port)
	return http.ListenAndServe(listenAddr, p)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te", // canonicalized version of "TE"
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

// Delete headers that are only relevant to the "hop" between the proxy and
// recipient server.
func delHopHeaders(header http.Header) {
	for _, h := range hopHeaders {
		header.Del(h)
	}
}

// Append the client's IP to the X-Forwarded-For header.
// This allows the downstream servers to record the original sender's IP.
func appendHostToXForwardHeader(header http.Header, host string) {
	// If we aren't the first proxy retain prior
	// X-Forwarded-For information as a comma+space
	// separated list and fold multiple headers into one.
	if prior, ok := header["X-Forwarded-For"]; ok {
		host = strings.Join(prior, ", ") + ", " + host
	}
	header.Set("X-Forwarded-For", host)
}

func (p *ChaperoneProxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	logger := log.DefaultLogger

	// Code from https://gist.github.com/yowu/f7dc34bd4736a65ff28d
	if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
		msg := "unsupported protocol scheme " + req.URL.Scheme
		http.Error(w, msg, http.StatusBadRequest)
		logger.Error(msg)
		return
	}

	//http: Request.RequestURI can't be set in client requests.
	//http://golang.org/src/pkg/net/http/client.go
	req.RequestURI = ""

	delHopHeaders(req.Header)

	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		appendHostToXForwardHeader(req.Header, clientIP)
	}

	// HTTPS upgrade directives
	upgradeConnection := true
	upgradeConnectionHeader := req.Header.Get("X-Upgrade-HTTPS")
	if strings.ToLower(upgradeConnectionHeader) == "false" || upgradeConnectionHeader == "0" {
		upgradeConnection = false
	}
	if upgradeConnection {
		req.URL.Scheme = "https"
	}

	minTTL := time.Duration(0)
	maxTTL := 24 * time.Hour
	defaultTTL := time.Duration(0)

	// Check if there is a cache override for the provided url.
	cacheOverride, ok := p.config.CacheOverrideForURL(req.URL.String())
	if ok {
		minTTL = cacheOverride.MinTTL
		maxTTL = cacheOverride.MaxTTL
		defaultTTL = cacheOverride.DefaultTTL
	}

	// Make proxied request.
	res, err := p.client.RoundTripWithOptions(req, &proxy.RequestOptions{
		UserAgent:       "Chaperone",
		MinCacheTTL:     minTTL,
		MaxCacheTTL:     maxTTL,
		DefaultCacheTTL: defaultTTL,
	})
	if err != nil {
		logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer res.Body.Close()

	delHopHeaders(res.Header)

	// Copy headers and body.
	copyHeader(w.Header(), res.Header)
	w.WriteHeader(res.StatusCode)
	_, err = io.Copy(w, res.Body)
	if err != nil {
		logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}
