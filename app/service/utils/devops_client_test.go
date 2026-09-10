package utils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSubmitCodeReview(t *testing.T) {
	var requestMethod string
	var requestPath string
	var authorization string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestMethod = request.Method
		requestPath = request.URL.Path
		authorization = request.Header.Get("Authorization")
		writer.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	err := SubmitCodeReview(context.Background(), server.URL+"/", 123, "admin", "devops-token")
	if err != nil {
		t.Fatalf("SubmitCodeReview() error = %v", err)
	}

	if requestMethod != http.MethodGet {
		t.Fatalf("request method = %q, want %q", requestMethod, http.MethodGet)
	}
	if requestPath != "/ppm/submitCodeReview/123" {
		t.Fatalf("request path = %q, want %q", requestPath, "/ppm/submitCodeReview/123")
	}
	if authorization != "devops-token" {
		t.Fatalf("authorization = %q, want %q", authorization, "devops-token")
	}
}
