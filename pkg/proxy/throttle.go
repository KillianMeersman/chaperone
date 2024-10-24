package proxy

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// HTTPThrottle interface performs throttling (rate-limiting).
type HTTPThrottle interface {
	// Wait for the throttle to pass.
	Wait(req *http.Request)
	// Block any clients using the throttle for the provided url for a certain duration.
	Block(req *http.Request, duration time.Duration)
	// Set the waiting duration for the provided request url and method.
	SetThrottle(req *http.Request, duration time.Duration)
	// Stop and clean up the throttle.
	Stop()
}

type hostThrottle struct {
	ticker   *time.Ticker
	blockers *sync.WaitGroup
}

// The MemoryHTTPThrottle stores throttles per path segment, so that it's possible
// to have hierarchical throttles.
// e.g. example.com has a throttle of 1s, example.com/test has a throttle of 2:
// A request to example.com/test would have to wait on BOTH.
type MemoryHTTPThrottle struct {
	throttles       *sync.Map
	defaultDuration time.Duration
}

// Create an in-memory throttle that handles per path/http-method throttling.
func NewMemoryHTTPThrottle(defaultDuration time.Duration) *MemoryHTTPThrottle {
	return &MemoryHTTPThrottle{
		throttles:       &sync.Map{},
		defaultDuration: defaultDuration,
	}
}

// Get the cache key for a request and path.
// The path is NOT derived from the provided request so that partial paths can be passed in.
func getRequestKey(req *http.Request, path string) string {
	return fmt.Sprintf("%s %s://%s%s", req.Method, req.URL.Scheme, req.URL.Host, path)
}

func (t *MemoryHTTPThrottle) Wait(req *http.Request) {
	pathParts := strings.Split(req.URL.Path, "/")

	// Check if there's a throttle for every part of the path and wait on it if there is.
	isThrottled := false
	for i, _ := range pathParts {
		key := getRequestKey(req, strings.Join(pathParts[:i+1], "/"))
		throttle, ok := t.throttles.Load(key)
		if ok {
			isThrottled = true
			throttle := throttle.(hostThrottle)
			throttle.blockers.Wait()
			<-throttle.ticker.C
		}
	}

	// If the request had no explicit throttles, wait the default duration.
	if !isThrottled {
		time.Sleep(t.defaultDuration)
	}
}

func (t *MemoryHTTPThrottle) Block(req *http.Request, d time.Duration) {
	key := getRequestKey(req, req.URL.Path)
	throttleAny, _ := t.throttles.LoadOrStore(key, hostThrottle{
		ticker:   time.NewTicker(d),
		blockers: &sync.WaitGroup{},
	})

	throttle := throttleAny.(hostThrottle)
	throttle.blockers.Add(1)
	go func() {
		time.Sleep(d)
		throttle.blockers.Done()
	}()
}

func (t *MemoryHTTPThrottle) SetThrottle(req *http.Request, duration time.Duration) {
	key := getRequestKey(req, req.URL.Path)
	throttleAny, exists := t.throttles.LoadOrStore(key, hostThrottle{
		ticker:   time.NewTicker(duration),
		blockers: &sync.WaitGroup{},
	})
	throttle := throttleAny.(hostThrottle)

	if exists {
		throttle.ticker.Reset(duration)
	}
}

func (t *MemoryHTTPThrottle) Stop() {
	t.throttles.Range(func(key, value any) bool {
		throttle, ok := value.(hostThrottle)
		if !ok {
			panic("in-memory throttle contains invalid datatype")
		}

		throttle.ticker.Stop()
		return true
	})
}
