package proxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func testCachingReturnsCorrectBody(t *testing.T, maxCacheSize int, headers http.Header, body []byte) {
	cache := NewMemoryHTTPCache(context.Background(), maxCacheSize)

	url, err := url.Parse("http://localhost:8080/test")
	if err != nil {
		t.Fatal(err)
	}
	request := &http.Request{
		Method: "GET",
		Host:   "localhost",
		URL:    url,
	}

	response := &http.Response{
		Request:    request,
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader(body)),
		Header:     headers,
	}

	returnedBody, err := cache.Cache(context.Background(), url.String(), response, 0, time.Hour, 0)
	if err != nil {
		t.Fatal(err)
	}

	data, err := io.ReadAll(returnedBody)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(data, body) {
		t.Fail()
	}
}

func TestContentLengthTooSmall(t *testing.T) {
	testCachingReturnsCorrectBody(t, 10e9, http.Header{
		"Content-Length": []string{"1"},
		"Cache-Control":  []string{"max-age=1000"},
	}, []byte("testtesttest"))

	testCachingReturnsCorrectBody(t, 1, http.Header{
		"Content-Length": []string{"1"},
		"Cache-Control":  []string{"max-age=1000"},
	}, []byte("testtesttest"))
}

func TestContentLengthTooLarge(t *testing.T) {
	testCachingReturnsCorrectBody(t, 10e9, http.Header{
		"Content-Length": []string{"1000000"},
		"Cache-Control":  []string{"max-age=1000"},
	}, []byte("testtesttest"))

	testCachingReturnsCorrectBody(t, 1, http.Header{
		"Content-Length": []string{"1"},
		"Cache-Control":  []string{"max-age=1000"},
	}, []byte("testtesttest"))
}

func TestContentLengthCorrect(t *testing.T) {
	body := []byte("testtesttest")
	testCachingReturnsCorrectBody(t, 10e9, http.Header{
		"Content-Length": []string{fmt.Sprint(len(body))},
		"Cache-Control":  []string{"max-age=1000"},
	}, body)

	testCachingReturnsCorrectBody(t, 1, http.Header{
		"Content-Length": []string{"1"},
		"Cache-Control":  []string{"max-age=1000"},
	}, []byte("testtesttest"))
}
