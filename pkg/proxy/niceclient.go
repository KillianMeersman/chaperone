package proxy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/KillianMeersman/chaperone/pkg/log"
)

const MaxWaitTimeMs int64 = 120e3 // two minutes

// Attempt to parse the provided Retry-After header.
// Returns the provided default duration if there is no such header present.
func ParseRetryAfterHeader(res *http.Response, defaultWait time.Duration) time.Duration {
	retryAfter := res.Header.Get("Retry-After")

	if len(retryAfter) > 0 {
		// Parse Retry-Duration as RFC1123 timestamp
		retryTime, err := time.Parse(time.RFC1123, retryAfter)
		if err == nil {
			return time.Until(retryTime)
		}

		// Parse Retry-Duration as seconds
		waitTime, err := time.ParseDuration(retryAfter + "s")
		if err == nil {
			return waitTime
		}
	}

	return defaultWait
}

// Request option struct for the NiceClient.RoundTripWithOptions method.
type RequestOptions struct {
	UserAgent       string
	MinCacheTTL     time.Duration
	MaxCacheTTL     time.Duration
	DefaultCacheTTL time.Duration
}

// A http client (RoundTripper) that performs retry logic, rate-limiting, robot-exclusion and caching.
type NiceClient struct {
	throttle     HTTPThrottle
	cache        *HTTPCache
	roundtripper http.RoundTripper
}

// Creates a new NiceClient with the provided options.
func NewNiceClient(ctx context.Context, roundTripper http.RoundTripper, throttle HTTPThrottle, cache *HTTPCache) *NiceClient {
	return &NiceClient{
		throttle,
		cache,
		roundTripper,
	}
}

// Implementation of the RoundTripper interface.
func (c *NiceClient) RoundTrip(req *http.Request) (*http.Response, error) {
	return c.RoundTripWithOptions(req, &RequestOptions{
		UserAgent:       "Chaperone",
		MinCacheTTL:     0,
		MaxCacheTTL:     24 * time.Hour,
		DefaultCacheTTL: 0,
	})
}

// Perform a round-trip with the provided options.
func (c *NiceClient) RoundTripWithOptions(req *http.Request, options *RequestOptions) (*http.Response, error) {
	attempt := 1

	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "ChaperoneBot/0.1")
	}

	ctx := req.Context()
	logger, _ := log.FromContext(req.Context())
	originalURL := req.URL.String()
	logger = logger.With("method", req.Method, "url", originalURL, "attempt", fmt.Sprintf("%d", attempt))

	// Special handling for GET requests
	if req.Method == http.MethodGet {
		// Check for cached responses and return if exists.
		cachedResponse, err := c.cache.Get(ctx, req)
		if err != nil {
			return nil, err
		}
		if cachedResponse != nil {
			return &http.Response{
				Status:     "",
				StatusCode: cachedResponse.StatusCode,
				Body:       cachedResponse.BodyReadCloser(),
			}, nil
		}
	}

	// ====== Request retry loop. ======
	if req.Body != nil {
		// Copy the request body into a buffer so we can retry the request multiple times.
		// This is necessary due to RoundTrip() always closing the request Body.
		bodyBuffer := bytes.NewBuffer(make([]byte, 0))
		written, err := io.Copy(bodyBuffer, req.Body)
		if err != nil {
			return nil, err
		}
		logger.With("buffer_size", fmt.Sprint(written)).Debug("buffered request body for retries")
		req.Body.Close()
		// The NopCloser won't do anything when RoundTrip() closes it.
		req.Body = io.NopCloser(bodyBuffer)
	}

	for {
		logger.Debug("waiting to make request")
		c.throttle.Wait(req)

		logger.Debug("making request")
		res, err := c.roundtripper.RoundTrip(req)
		if err != nil {
			return nil, err
		}

		logger = logger.With("status_code", fmt.Sprint(res.StatusCode))

		logger.Debug("got response", "status_code", fmt.Sprint(res.StatusCode))

		switch res.StatusCode {
		case 429, 503:
			// Backoff status codes, meaning we're rate-limited or the service is down.
			logger.Warning("got backoff status code")

			// Check if there was a Retry-After header and obey if present.
			waitTimeMs := ParseRetryAfterHeader(res, 3*time.Second).Milliseconds()
			if waitTimeMs > MaxWaitTimeMs {
				waitTimeMs = MaxWaitTimeMs
			}

			// Add random jitter to prevent thundering herd problem.
			jitter := rand.Int63n(2000)
			c.throttle.Block(res.Request, time.Duration(waitTimeMs+jitter)*time.Millisecond)
		case 301, 302, 307, 308:
			// Handle redirects.
			// We parse the url passed in the Location header and navigate there,
			// falling through to the default logic so that this redirect is transparent
			// to the caller.
			location, err := url.Parse(res.Header.Get("Location"))
			if err != nil {
				return nil, err
			}
			newReq := &http.Request{
				Method: req.Method,
				URL:    location,
				Header: req.Header,
				Body:   req.Body,
			}
			logger.With("to", location.String()).Info("Got redirect")
			res, err = c.RoundTripWithOptions(newReq, options)
			if err != nil {
				return nil, err
			}

			fallthrough
		default:
			// Default case, attempt caching and return the response.

			// Cache GET requests when possible.
			// Other HTTP methods should never be cached.
			if req.Method == http.MethodGet {
				res.Body, err = c.cache.Cache(ctx, originalURL, res, options.MinCacheTTL, options.MaxCacheTTL, options.DefaultCacheTTL)
			}

			// Success! Return response and any error.
			return res, err
		}

		// Break loop if the context is cancelled
		select {
		case <-ctx.Done():
			return nil, errors.New("context cancelled request retry loop")
		default:
		}

		attempt++
		logger = logger.With("attempt", fmt.Sprintf("%d", attempt))
	}
}

// Perform a GET request.
func (c *NiceClient) Get(ctx context.Context, url *url.URL, options *RequestOptions) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url.String(), nil)
	if err != nil {
		return nil, err
	}

	return c.RoundTripWithOptions(req, options)
}

// Perform a POST request.
func (c *NiceClient) Post(ctx context.Context, url *url.URL, body io.Reader, options *RequestOptions) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url.String(), body)
	if err != nil {
		return nil, err
	}

	return c.RoundTripWithOptions(req, options)
}

// Perform a PUT request.
func (c *NiceClient) Put(ctx context.Context, url *url.URL, body io.Reader, options *RequestOptions) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "PUT", url.String(), body)
	if err != nil {
		return nil, err
	}

	return c.RoundTripWithOptions(req, options)
}

// Perform a PATCH request.
func (c *NiceClient) Patch(ctx context.Context, url *url.URL, body io.Reader, options *RequestOptions) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "PATCH", url.String(), body)
	if err != nil {
		return nil, err
	}

	return c.RoundTripWithOptions(req, options)
}

// Perform a DELETE request.
func (c *NiceClient) Delete(ctx context.Context, url *url.URL, options *RequestOptions) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "DELETE", url.String(), nil)
	if err != nil {
		return nil, err
	}

	return c.RoundTripWithOptions(req, options)
}

// Perform an OPTIONS request.
func (c *NiceClient) Options(ctx context.Context, url *url.URL, options *RequestOptions) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "OPTIONS", url.String(), nil)
	if err != nil {
		return nil, err
	}

	return c.RoundTripWithOptions(req, options)
}
