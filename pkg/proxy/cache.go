package proxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/KillianMeersman/chaperone/pkg/datastructures/kvstore"
	"github.com/KillianMeersman/chaperone/pkg/log"
)

type CachedResponse struct {
	URL             string
	StatusCode      int
	Body            []byte
	FreshUntil      time.Time
	ResponseHeaders http.Header
	RequestHeaders  http.Header
}

// Return true if the cached response can be used for the provided request.
func (c *CachedResponse) IsValidForRequest(req *http.Request) bool {
	// Check the Vary header and if present, check that these headers match between
	// the original and the given request.
	if vary := c.ResponseHeaders.Get("Vary"); vary != "" {
		for _, varyHeader := range strings.Split(vary, ",") {
			varyHeader = strings.TrimSpace(varyHeader)

			if c.RequestHeaders.Get(varyHeader) != req.Header.Get(varyHeader) {
				return false
			}
		}
	}

	return true
}

// Returns a ReadCloser for the cached response's body.
// Closing this does nothing.
func (c *CachedResponse) BodyReadCloser() io.ReadCloser {
	return io.NopCloser(bytes.NewReader(c.Body))
}

// A HTTP cache caches responses according to their caching headers.
type HTTPCache interface {
	// Cache the response and return a ReadCloser so that the body can be re-read.
	// If re-using the response after caching, ensure the response body is replaced with the returned ReadCloser. e.g.
	// `res.Body, err = cache.Cache(ctx, url, res, ...)`
	Cache(ctx context.Context, url string, res *http.Response, minTTL, maxTTL, defaultTTL time.Duration) (io.ReadCloser, error)
	// Get the cached response for the given url. Returns nil if no response was cached.
	Get(ctx context.Context, req *http.Request) (*CachedResponse, error)
}

// An in-memory http cache.
type MemoryHTTPCache struct {
	// Stores the actual cached responses per cache-key (url + sorted vary headers).
	cachedResponses kvstore.KVStore[string, *CachedResponse]
	// Stores which headers to use in the cache keys per url.
	// This is based on the Vary header.
	urlVaryHeaders kvstore.KVStore[string, []string]
	currentSize    int
	maxSize        int
	IgnoreHeaders  bool
}

func NewMemoryHTTPCache(maxSize int) *MemoryHTTPCache {
	cache := &MemoryHTTPCache{
		cachedResponses: kvstore.NewMemoryKVStore[string, *CachedResponse](context.Background()),
		urlVaryHeaders:  kvstore.NewMemoryKVStore[string, []string](context.Background()),
		currentSize:     0,
		maxSize:         maxSize,
		IgnoreHeaders:   false,
	}

	return cache
}

func (c *MemoryHTTPCache) Cache(ctx context.Context, url string, res *http.Response, minTTL, maxTTL, defaultTTL time.Duration) (io.ReadCloser, error) {
	ttl := defaultTTL
	var err error

	logger, _ := log.FromContext(ctx)
	logger = logger.With("url", url)

	// If the cache is configured to ignore caching headers,
	// always cache with the default ttl.
	if !c.IgnoreHeaders {
		ttl, err = GetResponseCacheDuration(res, defaultTTL)
		if err != nil {
			return nil, err
		}
	}

	// Clamp the ttl according to the responses's cache headers
	// to the provided min. and maximum.
	if ttl < minTTL {
		logger.Debug("cache ttl too low, clamping to minimum")
		ttl = minTTL
	} else if ttl > maxTTL {
		logger.Debug("cache ttl to high, clamping to maximum")
		ttl = maxTTL
	}

	// If not allowed to cache, return early.
	if ttl <= 0 {
		logger.Debug("not allowed to cache")
		return res.Body, nil
	}

	// Store vary headers and compute cache key
	varyHeaders := GetVaryHeaderNames(res)

	logger = logger.With("vary_headers", strings.Join(varyHeaders, ","))

	// Store the headers upon which responses to the request's url vary.
	// We compute cache keys based on this value.
	c.urlVaryHeaders.Store(ctx, url, varyHeaders, ttl)
	cacheKey := GetCacheKey(varyHeaders, res.Request)

	// Check response size, assume the max allowed size unless specified by the Content-Length header.
	contentLength := int64(c.maxSize)
	if contentLengthHeader := res.Header.Get("Content-Length"); contentLengthHeader != "" {
		contentLength, err = strconv.ParseInt(contentLengthHeader, 10, 64)
		if err != nil {
			return nil, err
		}

		logger = logger.With("size", fmt.Sprintf("%d", contentLength))

		if c.currentSize+int(contentLength) > c.maxSize {
			logger.Warning("caching page would exceed max size, not caching")
			return res.Body, nil
		}
	}

	// Read response body to be cached until at most the content length (prevents certain DoS attacks).
	limitedBody := io.LimitReader(res.Body, contentLength)
	data, err := io.ReadAll(limitedBody)
	if err != nil {
		return nil, err
	}
	body := io.NopCloser(bytes.NewReader(data))

	logger = logger.With("size", fmt.Sprintf("%d", len(data)))

	if c.currentSize+len(data) > c.maxSize {
		logger.Warning("caching page would exceed max size, not caching")
		return body, nil
	}

	logger.With("ttl_seconds", fmt.Sprint(ttl.Seconds())).Debug("caching page")
	c.cachedResponses.Store(ctx, cacheKey, &CachedResponse{
		StatusCode:      res.StatusCode,
		Body:            data,
		ResponseHeaders: res.Header,
		RequestHeaders:  res.Request.Header,
	}, ttl)
	c.currentSize += len(data)

	return body, nil
}

func (c *MemoryHTTPCache) Get(ctx context.Context, req *http.Request) (*CachedResponse, error) {
	// Get the headers upon which responses at the request url vary.
	varyHeaders, _, err := c.urlVaryHeaders.Get(ctx, req.URL.String())
	if err != nil {
		return nil, err
	}

	logger, _ := log.FromContext(ctx)
	logger = logger.With("url", req.URL.String(), "vary_headers", strings.Join(varyHeaders, ","))

	cacheKey := GetCacheKey(varyHeaders, req)

	data, exists, err := c.cachedResponses.Get(ctx, cacheKey)
	if exists {
		logger.Debug("found cached page")
		return data, err
	}

	logger.Debug("did not find cached page")
	return nil, err
}
