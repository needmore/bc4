package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestUploadAttachment(t *testing.T) {
	rt := roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/123/attachments.json" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		if r.URL.RawQuery != "name=test.txt" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}

		if got := r.Header.Get("Authorization"); got != "Bearer token" {
			t.Fatalf("unexpected authorization header: %s", got)
		}

		if got := r.Header.Get("Content-Type"); got == "" {
			t.Fatalf("content type should be set")
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}
		if string(body) != "hello world" {
			t.Fatalf("unexpected body: %s", string(body))
		}

		resp := AttachmentUploadResponse{
			AttachableSGID: "SGID123",
			Filename:       "test.txt",
			ContentType:    "text/plain",
			ByteSize:       int64(len(body)),
		}
		var buf bytes.Buffer
		_ = json.NewEncoder(&buf).Encode(resp)

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(&buf),
			Header:     make(http.Header),
		}, nil
	})

	client := NewClient("123", "token")
	client.baseURL = "http://example.com"
	client.httpClient = &http.Client{Transport: rt}

	upload, err := client.UploadAttachment("test.txt", []byte("hello world"), "")
	if err != nil {
		t.Fatalf("UploadAttachment returned error: %v", err)
	}

	if upload.AttachableSGID != "SGID123" {
		t.Fatalf("unexpected sgid: %s", upload.AttachableSGID)
	}

	if upload.Filename != "test.txt" {
		t.Fatalf("unexpected filename: %s", upload.Filename)
	}
}
