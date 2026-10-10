package server

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOPMLCompareRejectsInvalidUploadsWithoutReplacingCache(t *testing.T) {
	for _, tc := range []struct {
		name    string
		content []byte
		status  int
	}{
		{"invalid XML", []byte("invalid"), http.StatusBadRequest},
		{"oversized file", bytes.Repeat([]byte("x"), maxOPMLSize+1), http.StatusRequestEntityTooLarge},
		{"oversized request", bytes.Repeat([]byte("x"), maxOPMLSize+(128<<10)), http.StatusRequestEntityTooLarge},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := &Server{opmlContent: []byte("previous"), opmlFilename: "previous.opml"}
			var body bytes.Buffer
			writer := multipart.NewWriter(&body)
			file, err := writer.CreateFormFile("opml", "new.opml")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := file.Write(tc.content); err != nil {
				t.Fatal(err)
			}
			if err := writer.Close(); err != nil {
				t.Fatal(err)
			}
			req := httptest.NewRequest(http.MethodPost, "/api/opml/compare", &body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			res := httptest.NewRecorder()
			s.handler().ServeHTTP(res, req)
			if res.Code != tc.status {
				t.Fatalf("status = %d, want %d", res.Code, tc.status)
			}
			if string(s.opmlContent) != "previous" || s.opmlFilename != "previous.opml" {
				t.Fatal("cache replaced on failed upload")
			}
		})
	}
}

func TestOPMLCompareEmptyAfterRestart(t *testing.T) {
	s := &Server{}
	res := httptest.NewRecorder()
	s.handler().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/opml/compare", nil))
	if res.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", res.Code)
	}
}
