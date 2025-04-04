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

type mockRoundTripper struct {
	Response []byte
}

func (m *mockRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	bodyReader := bytes.NewReader(m.Response)
	bodyReadCloser := io.NopCloser(bodyReader)

	return &http.Response{
		Request:    r,
		StatusCode: 200,
		Body:       bodyReadCloser,
	}, nil
}

func TestNiceClient(t *testing.T) {
	throttle := NewMemoryHTTPThrottle()
	cache := NewMemoryHTTPCache(context.Background(), 1000)

	roundTripper := &mockRoundTripper{
		Response: []byte("test"),
	}

	client := NewNiceClient(context.Background(), roundTripper, throttle, cache)
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	throttle.SetThrottle(req, time.Second)

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

	if time.Since(start) < 3*time.Second {
		t.FailNow()
	}
}

func TestNiceClientThrottle(t *testing.T) {
	cache := NewMemoryHTTPCache(context.Background(), 1000)
	throttle := NewMemoryHTTPThrottle()

	throttleDuration := 20 * time.Millisecond

	url, err := url.Parse("http://example.com")
	if err != nil {
		panic(err)
	}
	throttle.SetThrottle(&http.Request{
		Method: "GET",
		URL:    url,
	}, throttleDuration)
	throttle.SetThrottle(&http.Request{
		Method: "POST",
		URL:    url,
	}, throttleDuration*100)

	url, err = url.Parse("http://example.com/test")
	if err != nil {
		panic(err)
	}
	throttle.SetThrottle(&http.Request{
		Method: "GET",
		URL:    url,
	}, throttleDuration*100)

	roundTripper := &mockRoundTripper{
		Response: []byte("test"),
	}

	client := NewNiceClient(context.Background(), roundTripper, throttle, cache)

	for i := 0; i < 100; i++ {
		req, _ := http.NewRequest("GET", "http://example.com", nil)
		start := time.Now()
		client.RoundTrip(req)

		timeTaken := time.Since(start)

		if timeTaken < throttleDuration-5*time.Millisecond {
			t.Fatal("not enough time since last request")
		}

		if timeTaken > throttleDuration+5*time.Millisecond {
			t.Fatal("too much time since last request")
		}
	}

}
