package proxy

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"
)

type mockRountTripper struct {
	Response []byte
}

func (m *mockRountTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	bodyReader := bytes.NewReader(m.Response)
	bodyReadCloser := io.NopCloser(bodyReader)

	return &http.Response{
		Request:    r,
		StatusCode: 200,
		Body:       bodyReadCloser,
	}, nil
}

func TestNiceClient(t *testing.T) {
	throttle := NewMemoryHTTPThrottle(time.Second)
	cache := NewMemoryHTTPCache(1000)

	roundTripper := &mockRountTripper{
		Response: []byte("test"),
	}

	client := NewNiceClient(context.Background(), roundTripper, throttle, cache)
	req, _ := http.NewRequest("GET", "http://example.com", nil)

	start := time.Now()
	client.RoundTrip(req)
	if time.Since(start) < time.Second {
		t.FailNow()
	}

	url, err := url.Parse("http://example.com")
	if err != nil {
		t.Fatal(err)
	}
	throttle.Block(&http.Request{
		URL:    url,
		Method: "GET",
	}, 3*time.Second)
	client.RoundTrip(req)
	if time.Since(start) < 4*time.Second {
		t.FailNow()
	}
}
